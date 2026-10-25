import test from "node:test"
import assert from "node:assert/strict"
import { reassignUnstartedTask } from "./reassignment.ts"
import {
  createDemoWorkspaceState,
  demoWorkspaceReducer,
  type DemoWorkspaceAction,
} from "./workspace-state.ts"
import { createDemoWorkspaceActions } from "./workspace-actions.ts"
import { commitDraft, loadFor, validateDraft } from "./assignment.ts"
import { personalTasks, transitionPersonalTask } from "./execution.ts"
import { isTaskFrozen } from "./task-structure.ts"
import type { Draft } from "../models/workspace.ts"

const at = "2026-10-25T08:00:00.000Z"
const setup = () => {
  const state = createDemoWorkspaceState()
  const task = state.tasks.find((t) => t.id === "t6")!
  return {
    state,
    task,
    command: {
      projectId: task.projectId,
      taskId: task.id,
      expectedAssignmentId: task.currentAssignment!.id,
      personId: "an",
      reason: "  Điều chỉnh kế hoạch trước khi bắt đầu  ",
      overrideReason: "",
    },
  }
}

test("đổi người chưa bắt đầu: một dispatch, giữ đặc tả, kết thúc lượt cũ và đồng bộ quyền/tải", () => {
  const initial = setup()
  let state = initial.state
  const before = structuredClone(state)
  let count = 0
  const dispatch = (action: DemoWorkspaceAction) => {
    count++
    state = demoWorkspaceReducer(state, action)
  }
  createDemoWorkspaceActions(state, dispatch, "khoa").reassignTask(initial.command)
  const task = state.tasks.find((t) => t.id === "t6")!
  assert.equal(count, 1)
  assert.equal(task.status, "ASSIGNED")
  assert.equal(task.assigneeId, "an")
  assert.notEqual(task.currentAssignment!.id, initial.task.currentAssignment!.id)
  assert.equal(task.firstAssignedAt, initial.task.firstAssignedAt)
  assert.equal(isTaskFrozen(task), true)
  assert.equal(task.executionHistory?.at(-1)?.actorId, "khoa")
  const ended = state.work.contributions.find(
    (c) => c.assignment.id === initial.command.expectedAssignmentId,
  )!
  assert.equal(ended.assignment.endedBy, "khoa")
  assert.ok(ended.assignment.endedAt)
  assert.match(ended.assignment.endReason!, /Điều chỉnh kế hoạch/)
  assert.equal(loadFor("dung", state.tasks), loadFor("dung", before.tasks) - task.estimate)
  assert.equal(loadFor("an", state.tasks), loadFor("an", before.tasks) + task.estimate)
  assert.equal(
    personalTasks(state.projects, state.tasks, "dung").some((t) => t.id === "t6"),
    false,
  )
  assert.equal(
    personalTasks(state.projects, state.tasks, "an").some((t) => t.id === "t6"),
    true,
  )
  assert.throws(() =>
    transitionPersonalTask(state.projects, state.tasks, "dung", "t6", "ASSIGNED", "start"),
  )
  assert.equal(
    transitionPersonalTask(state.projects, state.tasks, "an", "t6", "ASSIGNED", "start").find(
      (t) => t.id === "t6",
    )?.status,
    "IN_PROGRESS",
  )
  assert.deepEqual(
    state.tasks.filter((t) => t.id !== "t6"),
    before.tasks.filter((t) => t.id !== "t6"),
  )
  assert.deepEqual(state.work.logs, before.work.logs)
  assert.deepEqual(state.work.submissions, before.work.submissions)
  assert.deepEqual(initial.state, before)
})

test("lỗi không dispatch: sai quyền, người cũ/PTCM/không tồn tại, lý do trống hoặc nền phân công cũ", () => {
  const { state, command } = setup()
  let count = 0
  const actions = (actor = "khoa") => createDemoWorkspaceActions(state, () => count++, actor)
  for (const actor of ["thanh", "dung", "ha", "unknown"])
    assert.throws(() => actions(actor).reassignTask(command))
  for (const personId of ["dung", "khoa", "unknown"])
    assert.throws(() => actions().reassignTask({ ...command, personId }))
  assert.throws(() => actions().reassignTask({ ...command, reason: " " }), /lý do/)
  assert.throws(
    () => actions().reassignTask({ ...command, expectedAssignmentId: "old" }),
    /thay đổi/,
  )
  assert.throws(() => actions().reassignTask({ ...command, projectId: "portal" }))
  assert.equal(count, 0)
})

test("không mở rộng sang giữa chừng hoặc task/project terminal, kể cả có log mà trạng thái vẫn ASSIGNED", () => {
  const { state, task, command } = setup()
  for (const status of [
    "IN_PROGRESS",
    "PAUSED",
    "PENDING_REVIEW",
    "DONE",
    "CLOSED",
    "CANCELLED",
    "PENDING_ASSIGNMENT",
  ] as const) {
    assert.throws(() =>
      reassignUnstartedTask({ ...state, tasks: [{ ...task, status }] }, "khoa", command, at, "new"),
    )
  }
  for (const status of ["COMPLETED", "CLOSED", "CANCELLED"] as const) {
    assert.throws(() =>
      reassignUnstartedTask(
        { ...state, projects: state.projects.map((p) => ({ ...p, status })) },
        "khoa",
        command,
        at,
        "new",
      ),
    )
  }
  const logged = {
    ...state,
    work: {
      ...state.work,
      logs: [
        {
          id: "l",
          assignmentId: command.expectedAssignmentId,
          taskId: task.id,
          employeeId: "dung",
          workDate: "2026-10-25",
          minutes: 30,
          description: "Đã làm",
          overLimitReason: "",
          createdAt: at,
          updatedAt: at,
        },
      ],
    },
  }
  assert.throws(() => reassignUnstartedTask(logged, "khoa", command, at, "new"), /đã có thực hiện/)
})

test("manual thiếu kỹ năng bắt buộc cần lý do ngoại lệ; PLANNING vẫn đổi được, không tự start dự án", () => {
  const { state, command } = setup()
  const planning = {
    ...state,
    projects: state.projects.map((p) => ({ ...p, status: "PLANNING" as const })),
  }
  const exception = { ...command, personId: "cuong" }
  assert.throws(() => reassignUnstartedTask(planning, "khoa", exception, at, "new"), /ngoại lệ/)
  const next = reassignUnstartedTask(
    planning,
    "khoa",
    { ...exception, overrideReason: "Có người hướng dẫn" },
    at,
    "new",
  )
  assert.equal(next.tasks.find((t) => t.id === "t6")?.status, "ASSIGNED")
  assert.equal(planning.projects[0].status, "PLANNING")
  assert.match(
    next.work.contributions.find((c) => c.assignment.id === command.expectedAssignmentId)!
      .assignment.endReason!,
    /Có người hướng dẫn/,
  )
})

test("đổi lựa chọn nháp chưa ghi lượt giao/tải; confirm chỉ giao cho lựa chọn cuối", () => {
  const { state } = setup()
  const project = state.projects.find((p) => p.id === "kho")!
  const before = structuredClone(state)
  const draft: Draft = {
    t1: { personId: "an", mode: "manual", reason: "", overrideReason: "", hadRun: false },
  }
  draft.t1 = { ...draft.t1, personId: "dung" }
  assert.deepEqual(validateDraft(project, "khoa", state.tasks, draft), [])
  assert.deepEqual(state, before)
  const committed = commitDraft(project, "khoa", state.tasks, draft)
  assert.equal(committed.find((t) => t.id === "t1")?.assigneeId, "dung")
  assert.equal(loadFor("an", committed), loadFor("an", state.tasks))
})
