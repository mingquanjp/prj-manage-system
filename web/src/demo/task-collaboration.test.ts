import test from "node:test"
import assert from "node:assert/strict"
import { createDemoWorkspaceState, demoWorkspaceReducer } from "./workspace-state.ts"
import { createDemoWorkspaceActions } from "./workspace-actions.ts"
import { professionalPeople } from "./profiles.ts"
import { canViewProject } from "./identity.ts"
import {
  replaceProjectPm,
  extendTaskDeadline,
  addTaskComment,
  canReadTaskDiscussion,
  canPostTaskDiscussion,
} from "./task-collaboration.ts"

function setup() {
  const state = createDemoWorkspaceState()
  const project = state.projects[0]!
  const task = {
    ...state.tasks.find((task) => task.projectId === project.id)!,
    start: project.start,
    end: "2026-11-01",
  }
  const identity = {
    people: professionalPeople(
      state.profileState.profiles,
      state.organization,
      state.employeeSkills,
      state.accounts.accounts,
    ),
    accounts: state.accounts.accounts,
  }
  const director = identity.accounts.find((a) => a.systemRoles.includes("DIRECTOR"))!.personId
  const candidate = identity.people.find(
    (p) => p.id !== project.pmId && p.id !== project.ptcmId,
  )!.id
  return { state, project, task, identity, director, candidate }
}
const at = "2026-10-28T08:00:00.000Z"

test("Thay PM chỉ đổi trách nhiệm PM; quyền cũ/mới và PTCM độc lập", () => {
  const { project, identity, director, candidate } = setup()
  const before = structuredClone(project)
  const next = replaceProjectPm(
    project,
    director,
    { projectId: project.id, beforePmId: project.pmId, pmId: candidate },
    identity,
  )
  assert.deepEqual(project, before)
  assert.deepEqual(next, { ...project, pmId: candidate })
  assert.equal(canViewProject(next, candidate, identity), true)
  assert.equal(canViewProject(next, project.pmId, identity), project.pmId === project.ptcmId)
  const combined = { ...project, ptcmId: project.pmId }
  assert.equal(
    canViewProject(
      replaceProjectPm(
        combined,
        director,
        { projectId: project.id, beforePmId: project.pmId, pmId: candidate },
        identity,
      ),
      project.pmId,
      identity,
    ),
    true,
  )
})

test("Thay PM chặn sai quyền, người khóa, dữ liệu cũ và dự án đóng; COMPLETED được thay", () => {
  const { project, identity, director, candidate } = setup()
  const command = { projectId: project.id, beforePmId: project.pmId, pmId: candidate }
  assert.throws(() => replaceProjectPm(project, project.pmId, command, identity), /Giám đốc/)
  assert.throws(
    () => replaceProjectPm(project, director, { ...command, beforePmId: "old" }, identity),
    /đã thay đổi/,
  )
  const locked = {
    ...identity,
    accounts: identity.accounts.map((a) =>
      a.personId === candidate ? { ...a, status: "LOCKED" as const } : a,
    ),
  }
  assert.throws(() => replaceProjectPm(project, director, command, locked), /đang hoạt động/)
  for (const status of ["CLOSED", "CANCELLED"] as const)
    assert.throws(
      () => replaceProjectPm({ ...project, status }, director, command, identity),
      /đóng\/hủy/,
    )
  assert.equal(
    replaceProjectPm({ ...project, status: "COMPLETED" }, director, command, identity).status,
    "COMPLETED",
  )
})

test("Gia hạn nhiều lần giữ kế hoạch gốc, estimate, đặc tả, phân công và tác giả lịch sử", () => {
  const { project, task, identity } = setup()
  const before = structuredClone(task)
  const command = {
    taskId: task.id,
    beforeStart: task.start,
    beforeEnd: task.end,
    start: task.start,
    end: "2026-11-05",
    reason: "Chờ API",
  }
  const next = extendTaskDeadline(project, task, project.ptcmId!, command, at, identity)
  const again = extendTaskDeadline(
    project,
    next,
    project.ptcmId!,
    { ...command, beforeEnd: next.end, end: "2026-11-10" },
    at,
    identity,
  )
  assert.deepEqual(task, before)
  assert.equal(again.originalEnd, task.end)
  assert.equal(again.originalStart, task.start)
  assert.equal(again.planHistory?.length, 2)
  const { originalStart: _s, originalEnd: _e, planHistory: _h, ...rest } = again
  assert.deepEqual(rest, { ...task, end: "2026-11-10" })
  assert.equal(_s, task.start)
  assert.equal(_e, task.end)
  assert.equal(_h?.length, 2)
  assert.equal(again.planHistory?.[0]?.actorId, project.ptcmId)
})

test("Gia hạn chặn PM-only, terminal, ngoài dự án, ngày sai, lý do trống và nền cũ", () => {
  const { project, task, identity } = setup()
  const cmd = {
    taskId: task.id,
    beforeStart: task.start,
    beforeEnd: task.end,
    start: task.start,
    end: "2026-11-05",
    reason: "Chờ API",
  }
  assert.throws(() => extendTaskDeadline(project, task, project.pmId, cmd, at, identity), /PTCM/)
  for (const status of ["DONE", "CLOSED", "CANCELLED"] as const)
    assert.throws(
      () => extendTaskDeadline(project, { ...task, status }, project.ptcmId!, cmd, at, identity),
      /PTCM/,
    )
  for (const patch of [
    { end: "2027-01-01" },
    { end: task.end },
    { start: "2026-10-02" },
    { end: "2026-02-30" },
    { reason: " " },
    { beforeEnd: "2026-10-31" },
  ])
    assert.throws(() =>
      extendTaskDeadline(project, task, project.ptcmId!, { ...cmd, ...patch }, at, identity),
    )
  const noDates = { ...task, start: undefined, end: undefined }
  const first = extendTaskDeadline(
    project,
    noDates,
    project.ptcmId!,
    { ...cmd, beforeStart: undefined, beforeEnd: undefined },
    at,
    identity,
  )
  assert.equal(first.originalEnd, cmd.end)
})

test("Trao đổi đúng task, người đang được giao và PTCM; giữ tác giả khi đổi PTCM", () => {
  const { project, task, identity, candidate } = setup()
  const assigned = {
    ...task,
    assigneeId: candidate,
    currentAssignment: { id: "assignment", employeeId: candidate, assignedAt: at },
  }
  const comments = addTaskComment(
    [],
    project,
    assigned,
    candidate,
    "  Em đang vướng API  ",
    "c1",
    at,
    identity,
  )
  const next = addTaskComment(
    comments,
    project,
    assigned,
    project.ptcmId!,
    "Dùng mock để triển khai trước.",
    "c2",
    at,
    identity,
  )
  assert.equal(next[0]?.content, "Em đang vướng API")
  assert.equal(next[0]?.actorId, candidate)
  assert.equal(comments.length, 1)
  assert.equal(
    canReadTaskDiscussion(project, { ...assigned, assigneeId: project.pmId }, candidate, identity),
    false,
  )
  assert.equal(
    canReadTaskDiscussion(
      { ...project, ptcmId: project.pmId },
      assigned,
      project.ptcmId!,
      identity,
    ),
    false,
  )
})

test("Trao đổi không cấp quyền cho PM-only/director/Admin, không viết khi khóa/terminal", () => {
  const { project, task, identity, director, candidate } = setup()
  const assigned = { ...task, assigneeId: candidate }
  assert.equal(canReadTaskDiscussion(project, assigned, project.pmId, identity), false)
  assert.equal(canReadTaskDiscussion(project, assigned, director, identity), false)
  assert.equal(canReadTaskDiscussion(project, assigned, "admin", identity), false)
  assert.throws(
    () => addTaskComment([], project, assigned, candidate, " ", "c1", at, identity),
    /Nhập/,
  )
  for (const status of ["COMPLETED", "CLOSED", "CANCELLED"] as const) {
    assert.equal(canReadTaskDiscussion({ ...project, status }, assigned, candidate, identity), true)
    assert.equal(
      canPostTaskDiscussion({ ...project, status }, assigned, candidate, identity),
      false,
    )
  }
  const locked = {
    ...identity,
    accounts: identity.accounts.map((a) =>
      a.personId === candidate ? { ...a, status: "LOCKED" as const } : a,
    ),
  }
  assert.equal(canReadTaskDiscussion(project, assigned, candidate, locked), false)
  for (const status of ["DONE", "CLOSED", "CANCELLED"] as const)
    assert.equal(
      canPostTaskDiscussion(project, { ...assigned, status }, candidate, identity),
      false,
    )
})

test("Biên action/reducer giữ dữ liệu miền khác, chia sẻ trao đổi và reset", () => {
  const { state: initial, project, task, director, candidate } = setup()
  let state = {
    ...initial,
    tasks: initial.tasks.map((t) => (t.id === task.id ? { ...task, assigneeId: candidate } : t)),
  }
  const dispatch = (action: Parameters<typeof demoWorkspaceReducer>[1]) => {
    state = demoWorkspaceReducer(state, action)
  }
  createDemoWorkspaceActions(state, dispatch, candidate).postTaskComment(task.id, "Em cần hỗ trợ")
  assert.equal(state.comments[0]?.actorId, candidate)
  const work = state.work
  const tasks = state.tasks
  createDemoWorkspaceActions(state, dispatch, director).replacePm({
    projectId: project.id,
    beforePmId: project.pmId,
    pmId: candidate,
  })
  assert.equal(state.tasks, tasks)
  assert.equal(state.work, work)
  assert.equal(state.events.at(-1)?.before?.pmId, project.pmId)
  assert.equal(state.events.at(-1)?.actorId, director)
  createDemoWorkspaceActions(state, dispatch, project.ptcmId!).extendDeadline({
    taskId: task.id,
    beforeStart: task.start,
    beforeEnd: task.end,
    start: task.start,
    end: "2026-11-05",
    reason: "Chờ API",
  })
  assert.equal(state.work, work)
  assert.equal(state.tasks.find((t) => t.id === task.id)?.end, "2026-11-05")
  state = demoWorkspaceReducer(state, { type: "reset" })
  assert.deepEqual(state.comments, [])
  assert.equal(
    state.tasks.some((t) => t.planHistory?.length),
    false,
  )
})
