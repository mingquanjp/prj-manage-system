import test from "node:test"
import assert from "node:assert/strict"
import { createDemoWorkspaceState, demoWorkspaceReducer } from "./workspace-state.ts"
import { createDemoWorkspaceActions } from "./workspace-actions.ts"
import {
  collectNotifications,
  notificationsFor,
  markNotificationRead,
  notificationDestination,
} from "./notifications.ts"
import { professionalPeople } from "./profiles.ts"
import type { WorkspaceNotification } from "../models/notification.ts"

const at = "2026-10-29T03:00:00.000Z"
function identity(state: ReturnType<typeof createDemoWorkspaceState>) {
  return {
    people: professionalPeople(
      state.profileState.profiles,
      state.organization,
      state.employeeSkills,
      state.accounts.accounts,
    ),
    accounts: state.accounts.accounts,
  }
}
function notification(
  recipientId: string,
  target: WorkspaceNotification["target"],
): WorkspaceNotification {
  return { id: "notice", recipientId, target, message: "Có cập nhật cần xem.", at }
}

test("Giao/thay/thu hồi tạo đúng người nhận, không phát lại khi chỉ đọc/render", () => {
  const before = createDemoWorkspaceState()
  const task = before.tasks.find((t) => t.status === "PENDING_ASSIGNMENT")!
  const assigned = {
    ...task,
    status: "ASSIGNED" as const,
    assigneeId: "dung",
    currentAssignment: { id: "new-assignment", employeeId: "dung", assignedAt: at },
  }
  const after = demoWorkspaceReducer(before, {
    type: "tasksChanged",
    tasks: before.tasks.map((t) => (t.id === task.id ? assigned : t)),
    at,
  })
  assert.equal(after.notifications.length, 1)
  assert.equal(after.notifications[0]?.recipientId, "dung")
  assert.deepEqual(collectNotifications(after, after), [])
  const unchanged = demoWorkspaceReducer(after, { type: "noticeChanged", notice: "Đã xem" })
  assert.deepEqual(unchanged.notifications, after.notifications)
  const replaced = {
    ...assigned,
    assigneeId: "binh",
    currentAssignment: { id: "replacement", employeeId: "binh", assignedAt: at },
  }
  const next = demoWorkspaceReducer(after, {
    type: "tasksChanged",
    tasks: after.tasks.map((t) => (t.id === task.id ? replaced : t)),
    at,
  })
  assert.deepEqual(
    next.notifications.slice(1).map((n) => n.recipientId),
    ["binh", "dung"],
  )
  const removed = {
    ...replaced,
    status: "PENDING_ASSIGNMENT" as const,
    assigneeId: undefined,
    currentAssignment: undefined,
  }
  const last = demoWorkspaceReducer(next, {
    type: "tasksChanged",
    tasks: next.tasks.map((t) => (t.id === task.id ? removed : t)),
    at,
  })
  assert.equal(last.notifications.at(-1)?.recipientId, "binh")
  assert.match(last.notifications.at(-1)?.message ?? "", /thu hồi/)
})

test("Hành trình thật Start → nộp → trả sửa → nộp lại → nghiệm thu phát thông báo", () => {
  let state = createDemoWorkspaceState()
  const dispatch = (action: Parameters<typeof demoWorkspaceReducer>[1]) => {
    state = demoWorkspaceReducer(state, action)
  }
  const task = state.tasks.find((t) => t.code === "KHO-T006")!
  const project = state.projects.find((p) => p.id === task.projectId)!
  assert.throws(() =>
    createDemoWorkspaceActions(state, dispatch, "dung").submitResult(task.id, "Chưa bắt đầu", ""),
  )
  assert.equal(state.notifications.length, 0)
  createDemoWorkspaceActions(state, dispatch, "dung").transitionTask(task.id, "ASSIGNED", "start")
  assert.equal(state.notifications.length, 0)
  createDemoWorkspaceActions(state, dispatch, "dung").submitResult(
    task.id,
    "Đã hoàn thiện API báo cáo tồn kho.",
    "",
  )
  assert.equal(state.notifications.at(-1)?.recipientId, project.ptcmId)
  const firstSubmission = state.work.submissions.at(-1)!
  createDemoWorkspaceActions(state, dispatch, project.ptcmId!).reviewResult(
    firstSubmission.id,
    "REJECTED",
    "Bổ sung xử lý ngày không có phát sinh.",
  )
  assert.equal(state.notifications.at(-1)?.recipientId, "dung")
  assert.match(state.notifications.at(-1)?.message ?? "", /chỉnh sửa/)
  createDemoWorkspaceActions(state, dispatch, "dung").submitResult(
    task.id,
    "Đã bổ sung trường hợp không có phát sinh.",
    "",
  )
  createDemoWorkspaceActions(state, dispatch, project.ptcmId!).reviewResult(
    state.work.submissions.at(-1)!.id,
    "ACCEPTED",
    "Đạt tiêu chí.",
  )
  assert.equal(state.notifications.length, 4)
  assert.match(state.notifications.at(-1)?.message ?? "", /nghiệm thu/)
  assert.match(
    notificationDestination(state, "dung", state.notifications.at(-1)!.id, identity(state)),
    /my-work\?task=.*panel=results/,
  )
})

test("Yêu cầu và quyết định hồ sơ/kỹ năng gửi đúng HR/Giám đốc/chính chủ, nội dung tối thiểu", () => {
  let state = createDemoWorkspaceState()
  const dispatch = (action: Parameters<typeof demoWorkspaceReducer>[1]) => {
    state = demoWorkspaceReducer(state, action)
  }
  const profile = state.profileState.profiles.find((p) => p.personId === "dung")!
  const values = { fullname: profile.fullname, phone: profile.phone, address: profile.address }
  createDemoWorkspaceActions(state, dispatch, "dung").requestProfileChange(
    values,
    { ...values, phone: "0901234567" },
    "Cập nhật số điện thoại",
  )
  assert.deepEqual(
    state.notifications.map((n) => n.recipientId),
    ["hr"],
  )
  const request = state.profileState.requests.at(-1)!
  assert.match(
    notificationDestination(state, "hr", state.notifications[0]!.id, identity(state)),
    /profile-requests\?request=/,
  )
  createDemoWorkspaceActions(state, dispatch, "hr").resolveProfileRequest(
    request.id,
    "APPROVED",
    "Đã đối chiếu",
  )
  assert.equal(state.notifications.at(-1)?.recipientId, "dung")
  assert.match(
    notificationDestination(state, "dung", state.notifications.at(-1)!.id, identity(state)),
    /my-profile\?request=/,
  )
  const skillRequest = {
    id: "skill-request",
    employeeId: "dung",
    source: "MANUAL" as const,
    operation: "ADD" as const,
    skillId: "React",
    before: null,
    proposed: { level: 2, years: null, note: "" },
    explanation: "Căn cứ riêng",
    externalEvidence: "",
    evidence: [],
    version: 1,
    requestedAt: at,
    status: "PENDING" as const,
    supplements: [],
  }
  state = demoWorkspaceReducer(state, {
    type: "employeeSkillsChanged",
    employeeSkills: { ...state.employeeSkills, requests: [skillRequest] },
    notice: "Gửi",
  })
  assert.equal(state.notifications.at(-1)?.recipientId, "ha")
  state = demoWorkspaceReducer(state, {
    type: "employeeSkillsChanged",
    employeeSkills: {
      ...state.employeeSkills,
      requests: [
        {
          ...skillRequest,
          status: "REJECTED",
          resolvedAt: at,
          resolvedBy: "ha",
          decisionReason: "Chưa đủ căn cứ",
        },
      ],
    },
    notice: "Xử lý",
  })
  assert.equal(state.notifications.at(-1)?.recipientId, "dung")
  assert.match(
    notificationDestination(state, "dung", state.notifications.at(-1)!.id, identity(state)),
    /tab=skills/,
  )
  assert.equal(
    state.notifications.some((n) => /0901234567|Căn cứ riêng|Chưa đủ căn cứ/.test(n.message)),
    false,
  )
})

test("Hiệu chỉnh dùng bản ghi gốc, đánh dấu đọc cùng nguồn, không đổi nghiệp vụ", () => {
  const base = createDemoWorkspaceState()
  const state = {
    ...base,
    work: {
      ...base.work,
      notifications: [
        { id: "legacy", recipientId: "dung", requestId: "r", message: "Có kết quả hiệu chỉnh", at },
      ],
    },
    notifications: [notification("dung", { kind: "task", id: "task" })],
  }
  const snapshot = structuredClone(state)
  const rows = notificationsFor(state, "dung", identity(state))
  assert.equal(rows.length, 2)
  const next = markNotificationRead(state, "dung", "effort:legacy", at, identity(state))
  assert.equal(next.work.notifications[0]?.readAt, at)
  assert.equal(next.notifications[0]?.readAt, undefined)
  assert.equal(next.work.requests, state.work.requests)
  const updated = { ...state, ...next }
  assert.equal(
    notificationsFor(updated, "dung", identity(updated)).filter((n) => !n.readAt).length,
    1,
  )
  const again = markNotificationRead(
    updated,
    "dung",
    "effort:legacy",
    "2026-10-30T00:00:00.000Z",
    identity(updated),
  )
  assert.equal(again.work.notifications[0]?.readAt, at)
  assert.deepEqual(state, snapshot)
  assert.throws(() => markNotificationRead(state, "khoa", "effort:legacy", at, identity(state)))
})

test("Thông báo không khôi phục quyền: mất phân công/role, tài khoản khóa, id người khác", () => {
  const state = createDemoWorkspaceState()
  const task = state.tasks.find((t) => t.code === "KHO-T006")!
  state.notifications = [notification("dung", { kind: "task", id: task.id })]
  assert.match(notificationDestination(state, "dung", "notice", identity(state)), /my-work/)
  assert.throws(() => notificationDestination(state, "khoa", "notice", identity(state)))
  const reassigned = {
    ...state,
    tasks: state.tasks.map((t) => (t.id === task.id ? { ...t, assigneeId: "binh" } : t)),
  }
  assert.throws(
    () => notificationDestination(reassigned, "dung", "notice", identity(reassigned)),
    /không còn quyền/,
  )
  const locked = {
    ...state,
    accounts: {
      ...state.accounts,
      accounts: state.accounts.accounts.map((a) =>
        a.personId === "dung" ? { ...a, status: "LOCKED" as const } : a,
      ),
    },
  }
  assert.deepEqual(notificationsFor(locked, "dung", identity(locked)), [])
  const missing = { ...state, tasks: [] }
  assert.throws(() => notificationDestination(missing, "dung", "notice", identity(missing)))
  const profileRequest = {
    id: "r",
    employeeId: "dung",
    requestedBy: "dung",
    requestedAt: at,
    status: "PENDING" as const,
    before: {},
    proposed: {},
    explanation: "",
  }
  state.profileState.requests = [profileRequest]
  state.notifications = [notification("hr", { kind: "profile-request", id: "r" })]
  const revoked = {
    ...state,
    accounts: {
      ...state.accounts,
      accounts: state.accounts.accounts.map((a) =>
        a.personId === "hr" ? { ...a, systemRoles: [] } : a,
      ),
    },
  }
  assert.throws(
    () => notificationDestination(revoked, "hr", "notice", identity(revoked)),
    /không còn quyền/,
  )
  assert.equal(notificationsFor(revoked, "hr", identity(revoked)).length, 1)
  assert.equal(
    markNotificationRead(revoked, "hr", "notice", at, identity(revoked)).notifications[0]?.readAt,
    at,
  )
})

test("Reset xóa thông báo phiên, không dựng thông báo giả từ fixture", () => {
  const state = createDemoWorkspaceState()
  assert.deepEqual(state.notifications, [])
  state.notifications = [notification("dung", { kind: "task", id: "missing" })]
  const next = demoWorkspaceReducer(state, { type: "reset" })
  assert.deepEqual(next.notifications, [])
  assert.deepEqual(next.work.notifications, [])
})
