import { test } from "node:test"
import assert from "node:assert/strict"
import { appointPtcm } from "./projects.ts"
import { executionBlockReason, personalTasks, transitionPersonalTask } from "./execution.ts"
import { initialTasks, projects } from "./fixtures.ts"
import {
  dailyMinutes,
  emptyWorkState,
  recordAssignmentChanges,
  rememberContributions,
  reviewSubmission,
  saveLog,
  submitResult,
  type LogInput,
} from "./effort.ts"
import {
  canReadRequest,
  contributionFor,
  correctionBlockReason,
  defaultLogDate,
  historicalContributions,
  logDateBounds,
  requestConflict,
  resolveLogChange,
  sendLogChange,
} from "./corrections.ts"

const now = new Date("2026-10-24T08:00:00Z")
const input: LogInput = {
  workDate: "2026-10-21",
  minutes: 90,
  description: "Thực hiện API báo cáo",
  overLimitReason: "",
}
function fixture() {
  const tasks = transitionPersonalTask(projects, initialTasks, "dung", "t6", "ASSIGNED", "start")
  const work = saveLog(projects, tasks, emptyWorkState(), "dung", "t6", input, undefined, now)
  const pending = submitResult(projects, tasks, work, "dung", "t6", "Đầu ra để nghiệm thu", "", now)
  return {
    tasks: pending.tasks,
    work: pending.work,
    assignmentId: tasks.find((t) => t.id === "t6")!.currentAssignment!.id,
  }
}

test("người hoàn thành task rồi thành PTCM vẫn mở được nhật ký riêng qua lịch sử, không được thực hiện task", () => {
  const f = fixture()
  const accepted = reviewSubmission(
    projects,
    f.tasks,
    f.work,
    "khoa",
    f.work.submissions[0].id,
    "ACCEPTED",
    "",
    [],
    now,
  )
  const work = rememberContributions(accepted.work, accepted.tasks)
  const appointed = appointPtcm(projects[0], "thanh", "dung", accepted.tasks, [])
  const nextProjects = projects.map((project) =>
    project.id === appointed.id ? appointed : project,
  )
  const task = accepted.tasks.find((entry) => entry.id === "t6")!

  assert.deepEqual(personalTasks(nextProjects, accepted.tasks, "dung"), [])
  const history = historicalContributions(nextProjects, accepted.tasks, work, "dung")
  assert.deepEqual(
    history.map((entry) => entry.assignment.id),
    [f.assignmentId],
  )
  assert.equal(history[0].assignment.id, task.currentAssignment!.id)
  assert.ok(history[0].assignment.endedAt)
  assert.ok(executionBlockReason(appointed, task, "dung"))
  assert.equal(correctionBlockReason(appointed, task, history[0], "dung"), undefined)
  const sent = sendLogChange(
    nextProjects,
    accepted.tasks,
    work,
    "dung",
    f.assignmentId,
    "ADD",
    "Bổ sung thời gian cũ",
    input,
    undefined,
    "ptcm-history",
    now,
  )
  assert.equal(sent.requests[0].status, "PENDING")
  assert.deepEqual(sent.logs, work.logs)
  assert.deepEqual(sent.submissions, work.submissions)
  assert.deepEqual(historicalContributions(nextProjects, accepted.tasks, work, "unknown"), [])
  assert.ok(
    historicalContributions(nextProjects, accepted.tasks, work, "an").every(
      (entry) => entry.assignment.employeeId === "an",
    ),
  )
})

test("lịch sử không lặp task còn hiển thị; lượt giao cũ vẫn chỉ thuộc người cũ", () => {
  const f = fixture()
  assert.deepEqual(historicalContributions(projects, f.tasks, f.work, "dung"), [])
  const next = f.tasks.map((task) =>
    task.id === "t6"
      ? {
          ...task,
          status: "ASSIGNED" as const,
          assigneeId: "an",
          currentAssignment: {
            id: "assignment-new",
            employeeId: "an",
            assignedAt: now.toISOString(),
          },
        }
      : task,
  )
  const work = recordAssignmentChanges(f.work, f.tasks, next, now)
  assert.deepEqual(
    historicalContributions(projects, next, work, "dung").map((entry) => entry.assignment.id),
    [f.assignmentId],
  )
  assert.deepEqual(historicalContributions(projects, next, work, "an"), [])
  const closed = projects.map((project) => ({ ...project, status: "CLOSED" as const }))
  const old = historicalContributions(closed, next, work, "dung")[0]
  assert.ok(old)
  assert.ok(
    correctionBlockReason(
      closed[0],
      next.find((task) => task.id === "t6")!,
      old,
      "dung",
    ),
  )
})

test("ngày form mới nằm trong khoảng thực tế; chưa có khoảng hợp lệ thì để trống", () => {
  const f = fixture()
  const contribution = contributionFor(f.work, f.tasks, f.assignmentId)!
  const ended = {
    ...contribution,
    assignment: { ...contribution.assignment, endedAt: "2026-10-22T17:30:00Z" },
  }
  // 17:30 UTC là ngày hôm sau tại Việt Nam.
  assert.equal(logDateBounds(projects[0], ended, now).max, "2026-10-23")
  assert.equal(defaultLogDate(projects[0], ended, now), "2026-10-23")
  assert.equal(defaultLogDate(projects[0], contribution, now), "2026-10-24")
  const future = {
    ...contribution,
    assignment: { ...contribution.assignment, assignedAt: "2026-10-25T00:00:00Z" },
  }
  assert.equal(defaultLogDate(projects[0], future, now), "")
  assert.equal(
    defaultLogDate({ ...projects[0], executionStartedAt: undefined }, contribution, now),
    "",
  )
  assert.equal(defaultLogDate(projects[0], undefined, now), "")
  assert.equal(f.work.logs[0].workDate, input.workDate)
})

test("pending UPDATE không đổi tổng/snapshot; chỉ PTCM duyệt, áp dụng một lần và có history/request", () => {
  const f = fixture(),
    before = structuredClone(f)
  const sent = sendLogChange(
    projects,
    f.tasks,
    f.work,
    "dung",
    f.assignmentId,
    "UPDATE",
    "Ghi thiếu thời gian",
    { ...input, minutes: 120 },
    f.work.logs[0],
    "update-1",
    now,
  )
  assert.equal(dailyMinutes(sent, "dung", input.workDate), 90)
  assert.deepEqual(sent.submissions, f.work.submissions)
  for (const actor of ["thanh", "ha", "dung"])
    assert.throws(() =>
      resolveLogChange(projects, f.tasks, sent, actor, "update-1", "APPROVED", "", now),
    )
  const approved = resolveLogChange(
    projects,
    f.tasks,
    sent,
    "khoa",
    "update-1",
    "APPROVED",
    "",
    now,
  )
  assert.equal(dailyMinutes(approved, "dung", input.workDate), 120)
  assert.equal(approved.history.length, 1)
  assert.equal(approved.history[0].requestId, "update-1")
  assert.equal(approved.history[0].actorId, "khoa")
  assert.deepEqual(approved.submissions, sent.submissions)
  assert.equal(approved.submissions[0].logSnapshot.logs[0].minutes, 90)
  assert.throws(
    () => resolveLogChange(projects, f.tasks, approved, "khoa", "update-1", "APPROVED", "", now),
    /kết thúc/,
  )
  assert.deepEqual(f, before)
})

test("ADD chống gửi lại; approved nối resultLogId, không history ADD; DELETE giữ phút/snapshot và xóa mềm", () => {
  const f = fixture()
  const sent = sendLogChange(
    projects,
    f.tasks,
    f.work,
    "dung",
    f.assignmentId,
    "ADD",
    "Quên ghi kiểm thử",
    input,
    undefined,
    "add-1",
    now,
  )
  assert.equal(
    sendLogChange(
      projects,
      f.tasks,
      sent,
      "dung",
      f.assignmentId,
      "ADD",
      "Quên ghi kiểm thử",
      input,
      undefined,
      "add-1",
      now,
    ),
    sent,
  )
  assert.throws(() =>
    sendLogChange(
      projects,
      f.tasks,
      sent,
      "dung",
      f.assignmentId,
      "ADD",
      "Nội dung khác",
      input,
      undefined,
      "add-1",
      now,
    ),
  )
  const approved = resolveLogChange(projects, f.tasks, sent, "khoa", "add-1", "APPROVED", "", now)
  assert.equal(approved.logs.length, 2)
  assert.equal(approved.history.length, 0)
  assert.equal(approved.requests[0].before, undefined)
  assert.equal(approved.requests[0].targetLogId, undefined)
  assert.equal(approved.requests[0].resultLogId, approved.logs[1].id)
  const deletedRequest = sendLogChange(
    projects,
    f.tasks,
    approved,
    "dung",
    f.assignmentId,
    "DELETE",
    "Ghi trùng",
    undefined,
    approved.logs[1],
    "delete-1",
    now,
  )
  const deleted = resolveLogChange(
    projects,
    f.tasks,
    deletedRequest,
    "khoa",
    "delete-1",
    "APPROVED",
    "",
    now,
  )
  assert.equal(deleted.logs[1].minutes, 90)
  assert.ok(deleted.logs[1].deletedAt)
  assert.equal(dailyMinutes(deleted, "dung", input.workDate), 90)
  assert.equal(deleted.history[0].requestId, "delete-1")
  assert.deepEqual(deleted.submissions, f.work.submissions)
  assert.throws(() =>
    sendLogChange(
      projects,
      f.tasks,
      deleted,
      "dung",
      f.assignmentId,
      "DELETE",
      "Lặp",
      undefined,
      deleted.logs[1],
      "delete-2",
      now,
    ),
  )
})

test("một pending/log; từ chối cần lý do; rút chính chủ; không thay nhật ký", () => {
  const f = fixture()
  const sent = sendLogChange(
    projects,
    f.tasks,
    f.work,
    "dung",
    f.assignmentId,
    "DELETE",
    "Ghi nhầm",
    undefined,
    f.work.logs[0],
    "delete-1",
    now,
  )
  assert.throws(
    () =>
      sendLogChange(
        projects,
        f.tasks,
        sent,
        "dung",
        f.assignmentId,
        "UPDATE",
        "Đổi ý",
        input,
        f.work.logs[0],
        "update-2",
        now,
      ),
    /chờ duyệt/,
  )
  assert.throws(
    () => resolveLogChange(projects, f.tasks, sent, "khoa", "delete-1", "REJECTED", " ", now),
    /lý do/,
  )
  assert.throws(() =>
    resolveLogChange(projects, f.tasks, sent, "khoa", "delete-1", "WITHDRAWN", "", now),
  )
  const rejected = resolveLogChange(
    projects,
    f.tasks,
    sent,
    "khoa",
    "delete-1",
    "REJECTED",
    "Cần làm rõ khoản ghi",
    now,
  )
  assert.deepEqual(rejected.logs, f.work.logs)
  assert.equal(rejected.history.length, 0)
  const withdrawn = resolveLogChange(
    projects,
    f.tasks,
    sent,
    "dung",
    "delete-1",
    "WITHDRAWN",
    "",
    now,
  )
  assert.deepEqual(withdrawn.logs, f.work.logs)
  assert.equal(withdrawn.requests[0].status, "WITHDRAWN")
  assert.throws(() =>
    resolveLogChange(projects, f.tasks, withdrawn, "dung", "delete-1", "WITHDRAWN", "", now),
  )
})

test("nền đổi chặn ghi đè nhưng vẫn từ chối/rút; đổi PTCM chuyển quyền hiện hành", () => {
  const f = fixture()
  const sent = sendLogChange(
    projects,
    f.tasks,
    f.work,
    "dung",
    f.assignmentId,
    "UPDATE",
    "Ghi thiếu",
    { ...input, minutes: 100 },
    f.work.logs[0],
    "update-1",
    now,
  )
  const changed = { ...sent, logs: sent.logs.map((l) => ({ ...l, description: "Nền đã đổi" })) }
  assert.ok(requestConflict(changed, changed.requests[0]))
  assert.throws(
    () => resolveLogChange(projects, f.tasks, changed, "khoa", "update-1", "APPROVED", "", now),
    /Không ghi đè/,
  )
  assert.equal(
    resolveLogChange(projects, f.tasks, changed, "khoa", "update-1", "REJECTED", "Nền đã đổi", now)
      .requests[0].status,
    "REJECTED",
  )
  assert.equal(
    resolveLogChange(projects, f.tasks, changed, "dung", "update-1", "WITHDRAWN", "", now)
      .requests[0].status,
    "WITHDRAWN",
  )
  const reassigned = projects.map((p) => ({ ...p, ptcmId: "cuong" }))
  assert.equal(canReadRequest(reassigned, f.tasks, sent, sent.requests[0], "khoa"), false)
  assert.throws(() =>
    resolveLogChange(reassigned, f.tasks, sent, "khoa", "update-1", "APPROVED", "", now),
  )
  assert.equal(
    resolveLogChange(reassigned, f.tasks, sent, "cuong", "update-1", "APPROVED", "", now)
      .requests[0].resolvedBy,
    "cuong",
  )
})

test("duyệt kiểm tra lại tổng xuyên dự án; >480 cần lý do, >1440 chặn; dữ liệu lỗi không áp dụng một phần", () => {
  const f = fixture()
  assert.throws(
    () =>
      sendLogChange(
        projects,
        f.tasks,
        f.work,
        "dung",
        f.assignmentId,
        "ADD",
        "Bổ sung",
        { ...input, minutes: 500 },
        undefined,
        "a",
        now,
      ),
    /lý do/,
  )
  const sent = sendLogChange(
    projects,
    f.tasks,
    f.work,
    "dung",
    f.assignmentId,
    "ADD",
    "Bổ sung",
    { ...input, minutes: 300 },
    undefined,
    "a",
    now,
  )
  const busy = {
    ...sent,
    logs: [
      ...sent.logs,
      {
        ...sent.logs[0],
        id: "outside",
        assignmentId: "other-project-assignment",
        taskId: "elsewhere",
        minutes: 1200,
      },
    ],
  }
  const snapshot = structuredClone(busy)
  assert.throws(
    () => resolveLogChange(projects, f.tasks, busy, "khoa", "a", "APPROVED", "", now),
    /24 giờ/,
  )
  assert.deepEqual(busy, snapshot)
  const overEight = {
    ...sent,
    logs: [...sent.logs, { ...sent.logs[0], id: "outside", assignmentId: "other", minutes: 120 }],
  }
  assert.throws(
    () => resolveLogChange(projects, f.tasks, overEight, "khoa", "a", "APPROVED", "", now),
    /lý do/,
  )
})

test("không gửi thay/không bypass Start Resume; CLOSED CANCELLED chỉ đọc; COMPLETED được hiệu chỉnh cũ", () => {
  const f = fixture(),
    task = f.tasks.find((t) => t.id === "t6")!
  assert.throws(() =>
    sendLogChange(
      projects,
      f.tasks,
      f.work,
      "thanh",
      f.assignmentId,
      "ADD",
      "Nhập thay",
      input,
      undefined,
      "a",
      now,
    ),
  )
  for (const status of ["ASSIGNED", "PAUSED", "IN_PROGRESS"] as const) {
    const tasks = f.tasks.map((t) => (t.id === task.id ? { ...t, status } : t))
    assert.ok(
      correctionBlockReason(
        projects[0],
        tasks.find((t) => t.id === task.id)!,
        contributionFor(f.work, tasks, f.assignmentId)!,
        "dung",
      ),
    )
  }
  const sent = sendLogChange(
    projects,
    f.tasks,
    f.work,
    "dung",
    f.assignmentId,
    "ADD",
    "Bổ sung",
    input,
    undefined,
    "a",
    now,
  )
  for (const status of ["CLOSED", "CANCELLED", "PLANNING"] as const) {
    const locked = projects.map((p) => ({ ...p, status }))
    assert.throws(() =>
      sendLogChange(
        locked,
        f.tasks,
        f.work,
        "dung",
        f.assignmentId,
        "ADD",
        "Bổ sung",
        input,
        undefined,
        "b",
        now,
      ),
    )
    assert.throws(() => resolveLogChange(locked, f.tasks, sent, "khoa", "a", "APPROVED", "", now))
    assert.throws(() => resolveLogChange(locked, f.tasks, sent, "dung", "a", "WITHDRAWN", "", now))
  }
  const completed = projects.map((p) => ({ ...p, status: "COMPLETED" as const }))
  const accepted = reviewSubmission(
    projects,
    f.tasks,
    sent,
    "khoa",
    sent.submissions[0].id,
    "ACCEPTED",
    "",
    [],
    now,
  )
  assert.equal(
    resolveLogChange(completed, accepted.tasks, accepted.work, "khoa", "a", "APPROVED", "", now)
      .requests[0].status,
    "APPROVED",
  )
})

test("nghiệm thu không tự xử lý yêu cầu pending; ADD sau DONE bị giới hạn ngày kết thúc, không đổi acceptedAt/snapshot", () => {
  const f = fixture()
  const sent = sendLogChange(
    projects,
    f.tasks,
    f.work,
    "dung",
    f.assignmentId,
    "ADD",
    "Bổ sung",
    input,
    undefined,
    "a",
    now,
  )
  const acceptedAt = new Date("2026-10-24T08:05:00Z")
  const later = new Date("2026-10-26T08:00:00Z")
  const accepted = reviewSubmission(
    projects,
    f.tasks,
    sent,
    "khoa",
    sent.submissions[0].id,
    "ACCEPTED",
    "",
    [],
    acceptedAt,
  )
  const work = rememberContributions(accepted.work, accepted.tasks)
  assert.equal(work.requests[0].status, "PENDING")
  const approved = resolveLogChange(
    projects,
    accepted.tasks,
    work,
    "khoa",
    "a",
    "APPROVED",
    "",
    later,
  )
  assert.equal(accepted.tasks.find((t) => t.id === "t6")!.acceptedAt, acceptedAt.toISOString())
  assert.deepEqual(approved.submissions, accepted.work.submissions)
  assert.throws(
    () =>
      sendLogChange(
        projects,
        accepted.tasks,
        work,
        "dung",
        f.assignmentId,
        "ADD",
        "Ghi sau",
        { ...input, workDate: "2026-10-25" },
        undefined,
        "b",
        later,
      ),
    /kết thúc/,
  )
  assert.throws(
    () =>
      sendLogChange(
        projects,
        f.tasks,
        f.work,
        "dung",
        f.assignmentId,
        "ADD",
        "Ngày sai",
        { ...input, workDate: "2026-11-01" },
        undefined,
        "c",
        now,
      ),
    /tương lai/,
  )
})

test("lượt cũ giữ tác giả/ngày kết thúc, không chuyển giờ cho người nhận mới", () => {
  const f = fixture(),
    task = f.tasks.find((t) => t.id === "t6")!
  const endedAt = new Date("2026-10-24T08:05:00Z")
  const later = new Date("2026-10-26T08:00:00Z")
  const next = f.tasks.map((t) =>
    t.id === "t6"
      ? {
          ...t,
          status: "ASSIGNED" as const,
          assigneeId: "an",
          currentAssignment: {
            id: "new-assignment",
            employeeId: "an",
            assignedAt: endedAt.toISOString(),
          },
        }
      : t,
  )
  const work = recordAssignmentChanges(f.work, f.tasks, next, endedAt)
  const old = contributionFor(work, next, task.currentAssignment!.id)!
  assert.equal(old.assignment.endedAt, endedAt.toISOString())
  const sent = sendLogChange(
    projects,
    next,
    work,
    "dung",
    f.assignmentId,
    "ADD",
    "Bổ sung phần cũ",
    input,
    undefined,
    "a",
    later,
  )
  const approved = resolveLogChange(projects, next, sent, "khoa", "a", "APPROVED", "", later)
  assert.equal(approved.logs.at(-1)!.employeeId, "dung")
  assert.equal(approved.logs.at(-1)!.assignmentId, f.assignmentId)
  assert.throws(
    () =>
      sendLogChange(
        projects,
        next,
        work,
        "dung",
        f.assignmentId,
        "ADD",
        "Sau bàn giao",
        { ...input, workDate: "2026-10-25" },
        undefined,
        "b",
        later,
      ),
    /kết thúc/,
  )
})
