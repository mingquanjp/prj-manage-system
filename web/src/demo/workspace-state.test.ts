import assert from "node:assert/strict"
import { test } from "node:test"
import {
  createDemoWorkspaceState,
  demoWorkspaceReducer,
  type DemoWorkspaceAction,
} from "./workspace-state.ts"
import { createDemoWorkspaceActions } from "./workspace-actions.ts"
import { initialGroups, initialTasks, projects } from "./fixtures.ts"
import { appointPtcm, createProject } from "./projects.ts"
import { commitDraft } from "./assignment.ts"
import { saveLog, vietnamDay } from "./effort.ts"

const at = "2026-10-24T10:00:00+07:00"

test("khởi tạo độc lập với fixture; khôi phục cùng lúc toàn bộ dữ liệu phiên", () => {
  const first = createDemoWorkspaceState()
  const second = createDemoWorkspaceState()
  first.projects[0].name = "Chỉ đổi bản sao"
  assert.deepEqual(second.projects, projects)
  assert.notDeepEqual(first.projects, projects)
  const before = structuredClone(first)
  const reset = demoWorkspaceReducer(first, { type: "reset" })
  assert.deepEqual(first, before)
  assert.deepEqual(reset.projects, projects)
  assert.deepEqual(reset.tasks, initialTasks)
  assert.deepEqual(reset.groups, initialGroups)
  assert.deepEqual(reset.events, [])
  assert.deepEqual(reset.work, second.work)
  assert.equal(reset.notice, "Đã khôi phục dữ liệu mẫu.")
})

test("nhóm rồi task liên tiếp không mất nhóm; lượt cũ giữ tác giả/mốc kết thúc được truyền vào", () => {
  const initial = createDemoWorkspaceState()
  const renamedGroups = initial.groups.map((g) =>
    g.id === "kho-backend" ? { ...g, name: "API" } : g,
  )
  const changedTasks = initial.tasks.map((t) =>
    t.groupId === "kho-backend" ? { ...t, group: "API" } : t,
  )
  const grouped = demoWorkspaceReducer(initial, { type: "groupsChanged", groups: renamedGroups })
  const tasksAction: DemoWorkspaceAction = { type: "tasksChanged", tasks: changedTasks, at }
  const next = demoWorkspaceReducer(grouped, tasksAction)
  assert.deepEqual(next.groups, renamedGroups)
  assert.equal(next.tasks.find((t) => t.id === "t6")?.group, "API")
  assert.deepEqual(demoWorkspaceReducer(grouped, tasksAction), next)
  const reassigned = changedTasks.map((t) =>
    t.id === "t6"
      ? {
          ...t,
          assigneeId: "an",
          currentAssignment: { id: "new-assignment", employeeId: "an", assignedAt: at },
        }
      : t,
  )
  const final = demoWorkspaceReducer(next, { type: "tasksChanged", tasks: reassigned, at })
  const old = final.work.contributions.find((c) => c.assignment.id === "assignment-t6-initial")!
  assert.equal(old.assignment.employeeId, "dung")
  assert.equal(old.assignment.endedAt, new Date(at).toISOString())
  assert.equal(
    final.work.contributions.find((c) => c.assignment.id === "new-assignment")?.assignment
      .employeeId,
    "an",
  )
  assert.equal(initial.tasks.find((t) => t.id === "t6")?.assigneeId, "dung")
})

test("project và event cập nhật cùng bước, snapshot history không theo thay đổi sau đó", () => {
  let state = createDemoWorkspaceState()
  const dispatch = (action: DemoWorkspaceAction) => {
    state = demoWorkspaceReducer(state, action)
  }
  const changed = appointPtcm(state.projects[1], "thanh", "cuong", state.tasks, [])
  createDemoWorkspaceActions(state, dispatch, "thanh").changeProject(changed, "Chỉ định PTCM")
  assert.equal(state.projects[1].ptcmId, "cuong")
  assert.equal(state.events[0].before?.ptcmId, "khoa")
  assert.equal(state.events[0].after.ptcmId, "cuong")
  changed.name = "Không đổi snapshot"
  assert.notEqual(state.events[0].after.name, changed.name)
  const created = createProject(state.projects, "ha", {
    code: "NEW",
    name: "Dự án mới",
    description: "",
    pmId: "thanh",
    start: "2026-10-24",
    end: "2026-12-31",
    priority: "Cao",
  })
  const id = createDemoWorkspaceActions(state, dispatch, "ha").saveCreatedProject(created)
  assert.equal(state.projects.at(-1)?.id, id)
  assert.equal(state.events.at(-1)?.projectId, id)
  assert.equal(state.events.at(-1)?.actorId, "ha")
  assert.equal(state.projects.at(-1)?.pmId, "thanh")
})

test("điều phối phân công, log, nộp, trả sửa và nghiệm thu giữ các phần state đồng bộ", () => {
  let state = createDemoWorkspaceState()
  const dispatch = (action: DemoWorkspaceAction) => {
    state = demoWorkspaceReducer(state, action)
  }
  const actions = (actorId: string) => createDemoWorkspaceActions(state, dispatch, actorId)
  const assigned = commitDraft(state.projects[0], "khoa", state.tasks, {
    t1: { personId: "dung", mode: "manual", reason: "", overrideReason: "", hadRun: false },
  })
  actions("khoa").changeTasks(assigned)
  assert.equal(state.tasks[0].assigneeId, "dung")
  assert.ok(
    state.work.contributions.some((c) => c.assignment.id === state.tasks[0].currentAssignment?.id),
  )
  actions("dung").transitionTask("t6", "ASSIGNED", "start")
  actions("dung").changeWork(
    saveLog(state.projects, state.tasks, state.work, "dung", "t6", {
      workDate: vietnamDay(),
      minutes: 30,
      description: "Kiểm tra điều phối",
      overLimitReason: "",
    }),
  )
  assert.equal(state.work.logs.length, 1)
  actions("dung").submitResult("t6", "Đã có đầu ra", "")
  const submission = state.work.submissions.filter((s) => s.taskId === "t6").at(-1)!
  assert.equal(state.tasks.find((t) => t.id === "t6")?.status, "PENDING_REVIEW")
  assert.equal(submission.logSnapshot.logs[0].minutes, 30)
  actions("khoa").reviewResult(submission.id, "REJECTED", "Chỉnh lại đầu ra")
  assert.equal(state.tasks.find((t) => t.id === "t6")?.status, "IN_PROGRESS")
  actions("dung").submitResult("t6", "Đã chỉnh đầu ra", "")
  actions("khoa").reviewResult(
    state.work.submissions.filter((s) => s.taskId === "t6").at(-1)!.id,
    "ACCEPTED",
    "",
  )
  assert.equal(state.tasks.find((t) => t.id === "t6")?.status, "DONE")
  assert.ok(
    state.work.contributions.find((c) => c.assignment.id === submission.assignmentId)?.assignment
      .endedAt,
  )
  assert.equal(
    state.work.submissions.find((s) => s.id === submission.id)?.review?.result,
    "REJECTED",
  )
  assert.equal(
    state.work.submissions.filter((s) => s.taskId === "t6").at(-1)?.review?.result,
    "ACCEPTED",
  )
  assert.equal(state.projects[0].status, "IN_PROGRESS")
})

test("lỗi thao tác trả về caller trước dispatch; reducer lặp cùng payload không sinh dữ liệu mới", () => {
  let state = createDemoWorkspaceState()
  const dispatched: DemoWorkspaceAction[] = []
  const dispatch = (action: DemoWorkspaceAction) => {
    dispatched.push(action)
    state = demoWorkspaceReducer(state, action)
  }
  const before = structuredClone(state)
  assert.throws(() =>
    createDemoWorkspaceActions(state, dispatch, "thanh").transitionTask("t6", "ASSIGNED", "start"),
  )
  assert.throws(() =>
    createDemoWorkspaceActions(state, dispatch, "dung").submitResult("t6", "Chưa bắt đầu", ""),
  )
  assert.deepEqual(state, before)
  assert.equal(dispatched.length, 0)
  createDemoWorkspaceActions(state, dispatch, "dung").transitionTask("t6", "ASSIGNED", "start")
  const started = state
  createDemoWorkspaceActions(state, dispatch, "dung").submitResult("t6", "Kết quả không log", "")
  const action = dispatched.at(-1)!
  const once = demoWorkspaceReducer(started, action)
  const repeated = demoWorkspaceReducer(started, action)
  assert.deepEqual(once, repeated)
  assert.equal(once.work.submissions.length, started.work.submissions.length + 1)
  assert.equal(once.work.submissions.at(-1)!.id, state.work.submissions.at(-1)!.id)
  assert.equal(started.work.submissions.filter((s) => s.taskId === "t6").length, 0)
})
