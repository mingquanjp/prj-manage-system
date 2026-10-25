import test from "node:test"
import assert from "node:assert/strict"
import { createDemoWorkspaceState } from "./workspace-state.ts"
import { transitionProjectLifecycle } from "./project-lifecycle.ts"
import { createDemoWorkspaceActions } from "./workspace-actions.ts"
import { demoWorkspaceReducer, type DemoWorkspaceAction } from "./workspace-state.ts"
import { saveLog, vietnamDay } from "./effort.ts"
import { progress } from "./projects.ts"

const now = "2026-10-25T08:00:00.000Z"
const setup = () => {
  const state = createDemoWorkspaceState()
  const project = state.projects.find((p) => p.id === "kho")!
  return { state, project }
}

test("vòng đời chỉ dành cho PM, không tự hoàn thành khi còn task mở hoặc không có DONE", () => {
  const { state, project } = setup()
  const run = (actorId = project.pmId) =>
    transitionProjectLifecycle(
      state,
      actorId,
      { projectId: project.id, expectedStatus: project.status, action: "complete" },
      now,
    )
  assert.throws(() => run(project.ptcmId), /PM/)
  assert.throws(() => run(), /công việc/)
  state.tasks = state.tasks.map((t) =>
    t.projectId === project.id ? { ...t, status: "CANCELLED" } : t,
  )
  assert.throws(() => run(), /nghiệm thu/)
})

test("hoàn thành giữ DONE; đóng đổi DONE sang CLOSED, giữ acceptedAt và tiến độ", () => {
  const { state, project } = setup()
  state.tasks = state.tasks.map((t) =>
    t.projectId === project.id && t.status !== "DONE" ? { ...t, status: "CANCELLED" } : t,
  )
  const done = state.tasks.find((t) => t.projectId === project.id && t.status === "DONE")!
  const completed = transitionProjectLifecycle(
    state,
    project.pmId,
    { projectId: project.id, expectedStatus: "IN_PROGRESS", action: "complete" },
    now,
  )
  assert.equal(completed.project.status, "COMPLETED")
  assert.equal(completed.tasks.find((t) => t.id === done.id)?.status, "DONE")
  state.projects = state.projects.map((p) => (p.id === project.id ? completed.project : p))
  state.tasks = completed.tasks
  const closed = transitionProjectLifecycle(
    state,
    project.pmId,
    { projectId: project.id, expectedStatus: "COMPLETED", action: "close" },
    now,
  )
  assert.equal(closed.project.status, "CLOSED")
  assert.equal(closed.tasks.find((t) => t.id === done.id)?.status, "CLOSED")
  assert.equal(closed.tasks.find((t) => t.id === done.id)?.acceptedAt, done.acceptedAt)
  assert.equal(progress(closed.tasks.filter((t) => t.projectId === project.id)), 100)
  assert.throws(
    () =>
      transitionProjectLifecycle(
        { ...state, projects: [closed.project] },
        project.pmId,
        { projectId: project.id, expectedStatus: "COMPLETED", action: "close" },
        now,
      ),
    /thay đổi/,
  )
})

test("nộp có log rồi hủy dự án: một dispatch giữ snapshot/log, kết thúc lượt giao và khóa hành động", () => {
  const initial = setup()
  let state = initial.state
  const project = initial.project
  let dispatches = 0
  const dispatch = (action: DemoWorkspaceAction) => {
    dispatches++
    state = demoWorkspaceReducer(state, action)
  }
  const actions = (actorId: string) => createDemoWorkspaceActions(state, dispatch, actorId)
  actions("dung").transitionTask("t6", "ASSIGNED", "start")
  actions("dung").changeWork(
    saveLog(state.projects, state.tasks, state.work, "dung", "t6", {
      workDate: vietnamDay(),
      minutes: 30,
      description: "Đóng góp trước hủy",
      overLimitReason: "",
    }),
  )
  actions("dung").submitResult("t6", "Kết quả đang chờ", "https://example.com/pr/1")
  const before = structuredClone(state)
  const count = dispatches
  actions(project.pmId).transitionLifecycle({
    projectId: project.id,
    expectedStatus: project.status,
    action: "cancelProject",
    reason: "Dừng dự án",
  })
  assert.equal(dispatches, count + 1)
  assert.equal(state.projects.find((p) => p.id === project.id)?.status, "CANCELLED")
  assert.deepEqual(state.work.logs, before.work.logs)
  assert.deepEqual(state.work.submissions, before.work.submissions)
  assert.equal(state.work.submissions[0].review, undefined)
  assert.ok(state.work.contributions.find((entry) => entry.taskId === "t6")?.assignment.endedAt)
  assert.deepEqual(
    state.tasks.filter((t) => t.projectId !== project.id),
    before.tasks.filter((t) => t.projectId !== project.id),
  )
  assert.equal(state.events.at(-1)?.reason, "Dừng dự án")
  assert.throws(() => actions("khoa").reviewResult(state.work.submissions[0].id, "ACCEPTED", ""))
  assert.throws(() => actions("dung").submitResult("t6", "Nộp lại", ""))
  assert.throws(() =>
    actions(project.pmId).transitionLifecycle({
      projectId: project.id,
      expectedStatus: "CANCELLED",
      action: "complete",
    }),
  )
  assert.equal(dispatches, count + 1)
})

test("hủy task không áp dụng cho DONE/sai dự án/nền cũ; PLANNING hủy không tự triển khai", () => {
  const { state, project } = setup()
  const done = state.tasks.find((t) => t.projectId === project.id && t.status === "DONE")!
  const command = {
    projectId: project.id,
    expectedStatus: project.status,
    action: "cancelTask" as const,
    taskId: done.id,
    expectedTaskStatus: done.status,
    reason: "Hủy",
  }
  assert.throws(() => transitionProjectLifecycle(state, project.pmId, command, now))
  assert.throws(
    () =>
      transitionProjectLifecycle(
        state,
        project.pmId,
        { ...command, taskId: "t6", expectedTaskStatus: "PAUSED" },
        now,
      ),
    /thay đổi/,
  )
  const portal = state.projects.find((p) => p.id !== project.id)!
  assert.throws(() =>
    transitionProjectLifecycle(
      state,
      portal.pmId,
      { ...command, projectId: portal.id, expectedStatus: portal.status },
      now,
    ),
  )
  const next = transitionProjectLifecycle(
    state,
    portal.pmId,
    {
      projectId: portal.id,
      expectedStatus: "PLANNING",
      action: "cancelProject",
      reason: "Dừng kế hoạch",
    },
    now,
  )
  assert.equal(next.project.status, "CANCELLED")
  assert.equal(next.project.executionStartedAt, undefined)
})

test("hủy task cần lý do, kết thúc lượt giao và giữ log/lần nộp; không ảnh hưởng task dự án khác", () => {
  const { state, project } = setup()
  const task = state.tasks.find((t) => t.projectId === project.id && t.status === "PENDING_REVIEW")!
  const command = {
    projectId: project.id,
    expectedStatus: project.status,
    action: "cancelTask" as const,
    taskId: task.id,
    expectedTaskStatus: task.status,
    reason: "  Dừng phạm vi  ",
  }
  assert.throws(
    () => transitionProjectLifecycle(state, project.pmId, { ...command, reason: " " }, now),
    /lý do/,
  )
  const before = structuredClone(state)
  const next = transitionProjectLifecycle(state, project.pmId, command, now)
  const cancelled = next.tasks.find((t) => t.id === task.id)!
  assert.equal(cancelled.status, "CANCELLED")
  assert.equal(cancelled.currentAssignment?.endedAt, now)
  assert.equal(cancelled.currentAssignment?.endedBy, project.pmId)
  assert.equal(cancelled.executionHistory?.at(-1)?.reason, "Dừng phạm vi")
  assert.deepEqual(next.work.submissions, state.work.submissions)
  assert.deepEqual(next.work.logs, state.work.logs)
  assert.deepEqual(
    next.tasks.filter((t) => t.id !== task.id),
    state.tasks.filter((t) => t.id !== task.id),
  )
  assert.deepEqual(state, before)
})

test("pending hiệu chỉnh chặn đóng; hủy dự án kết thúc pending trong đúng dự án, không tự duyệt", () => {
  const { state, project } = setup()
  const task = state.tasks.find((t) => t.projectId === project.id && t.currentAssignment)!
  const other = state.tasks.find((t) => t.projectId !== project.id && t.currentAssignment)!
  state.work.requests = [task, other].map((t, index) => ({
    id: `r${index}`,
    assignmentId: t.currentAssignment!.id,
    operation: "ADD",
    requestedBy: t.assigneeId!,
    requestedAt: now,
    reason: "Bổ sung",
    status: "PENDING",
  }))
  state.projects = state.projects.map((p) =>
    p.id === project.id ? { ...p, status: "COMPLETED" } : p,
  )
  assert.throws(
    () =>
      transitionProjectLifecycle(
        state,
        project.pmId,
        { projectId: project.id, expectedStatus: "COMPLETED", action: "close" },
        now,
      ),
    /hiệu chỉnh/,
  )
  state.projects = state.projects.map((p) => (p.id === project.id ? project : p))
  const next = transitionProjectLifecycle(
    state,
    project.pmId,
    {
      projectId: project.id,
      expectedStatus: project.status,
      action: "cancelProject",
      reason: "Dừng dự án",
    },
    now,
  )
  assert.equal(next.project.status, "CANCELLED")
  assert.equal(next.work.requests[0].status, "REJECTED")
  assert.match(next.work.requests[0].decisionReason!, /PROJECT_CANCELLED/)
  assert.equal(next.work.requests[1].status, "PENDING")
  assert.equal(next.work.notifications.length, 1)
  assert.deepEqual(next.work.logs, state.work.logs)
  assert.equal(
    next.tasks.find((t) => t.projectId === project.id && t.status === "DONE")?.status,
    "DONE",
  )
})
