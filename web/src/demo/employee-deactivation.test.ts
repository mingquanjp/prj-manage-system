import test from "node:test"
import assert from "node:assert/strict"
import { deactivateEmployeeProfile, employeeDeactivationBlockers } from "./employee-deactivation.ts"
import {
  createDemoWorkspaceState,
  demoWorkspaceReducer,
  type DemoWorkspaceAction,
} from "./workspace-state.ts"
import { createDemoWorkspaceActions } from "./workspace-actions.ts"
import { professionalPeople, requestProfileChange, resolveProfileRequest } from "./profiles.ts"
import { employeeEditValues, updateEmployeeProfile } from "./employee-administration.ts"
import { commitDraft, recommendations, validateDraft } from "./assignment.ts"
import { appointPtcm, validateProjectInput } from "./projects.ts"
import { personnelProfiles } from "./personnel.ts"
import { selectWorkspaceNavigation } from "../app/navigation.ts"
import { changeOrganization, organizationReferences } from "./organization.ts"
import type { ProjectInput, TaskStatus } from "../models/workspace.ts"

const at = "2026-10-27T03:00:00Z"

test("HR chặn mọi trạng thái việc mở và trách nhiệm dự án, chỉ trả nhận diện tối thiểu", () => {
  const state = createDemoWorkspaceState()
  for (const status of ["ASSIGNED", "IN_PROGRESS", "PAUSED", "PENDING_REVIEW"] as TaskStatus[]) {
    const tasks = [{ ...state.tasks[0], assigneeId: "cuong", status }]
    const blockers = employeeDeactivationBlockers("cuong", [], tasks, state.profileState.profiles)
    assert.equal(blockers.length, 1)
    assert.match(blockers[0].label, new RegExp(tasks[0].code))
    assert.deepEqual(Object.keys(blockers[0]).sort(), ["kind", "label"])
    assert.throws(
      () => deactivateEmployeeProfile(state.profileState, [], tasks, "hr", "cuong", "x", at),
      /Chưa thể/,
    )
  }
  for (const status of ["DONE", "CLOSED", "CANCELLED"] as const) {
    const tasks = [{ ...state.tasks[0], assigneeId: "cuong", status }]
    assert.equal(
      employeeDeactivationBlockers("cuong", [], tasks, state.profileState.profiles).length,
      0,
    )
  }
  for (const status of ["PLANNING", "IN_PROGRESS", "COMPLETED"] as const) {
    const project = { ...state.projects[0], pmId: "cuong", ptcmId: "cuong", status }
    assert.match(
      employeeDeactivationBlockers("cuong", [project], [], state.profileState.profiles)[0].label,
      /PM \+ PTCM/,
    )
    assert.throws(
      () => deactivateEmployeeProfile(state.profileState, [project], [], "hr", "cuong", "x", at),
      /Chưa thể/,
    )
  }
  for (const status of ["CLOSED", "CANCELLED"] as const)
    assert.equal(
      employeeDeactivationBlockers(
        "cuong",
        [{ ...state.projects[0], pmId: "cuong", ptcmId: "cuong", status }],
        [],
        state.profileState.profiles,
      ).length,
      0,
    )
})

test("Vô hiệu hóa giữ hồ sơ, yêu cầu và lịch sử; chỉ HR active, không áp dụng lặp", () => {
  const state = createDemoWorkspaceState()
  const before = state.profileState.profiles.find((p) => p.personId === "cuong")!
  const edited = updateEmployeeProfile(
    state.profileState,
    state.organization,
    "hr",
    before,
    { ...employeeEditValues(before), address: "Hà Nội" },
    "edit",
    at,
  )
  const current = edited.profiles.find((p) => p.personId === "cuong")!
  const pending = requestProfileChange(
    edited,
    "cuong",
    current,
    { ...current, phone: "0901234567" },
    "",
    "request",
    at,
  )
  const original = structuredClone(pending)
  for (const actor of ["khoa", "thanh", "ha", "cuong", "missing"])
    assert.throws(
      () => deactivateEmployeeProfile(pending, [], [], actor, "cuong", "off", at),
      /Chỉ HR/,
    )
  const next = deactivateEmployeeProfile(pending, [], [], "hr", "cuong", "off", at)
  assert.deepEqual(pending, original)
  assert.deepEqual(
    next.profiles.find((p) => p.personId === "cuong"),
    { ...current, employmentStatus: "INACTIVE" },
  )
  assert.equal(next.requests, pending.requests)
  assert.equal(next.edits, pending.edits)
  assert.deepEqual(next.deactivations, [{ id: "off", employeeId: "cuong", actorId: "hr", at }])
  assert.throws(
    () => deactivateEmployeeProfile(next, [], [], "hr", "cuong", "again", at),
    /đã được/,
  )
  assert.throws(
    () => requestProfileChange(next, "cuong", current, { ...current, phone: "09" }, "", "new", at),
    /không có quyền/,
  )
  assert.throws(
    () => resolveProfileRequest(next, "hr", "request", "APPROVED", "", at),
    /không còn hoạt động/,
  )
  assert.equal(
    organizationReferences(next.profiles, "positions", current.positionId).some(
      (p) => p.personId === "cuong",
    ),
    true,
  )
  assert.throws(
    () =>
      changeOrganization(
        state.organization,
        next.profiles,
        "hr",
        {
          action: "delete",
          kind: "positions",
          id: current.positionId,
          beforeName: "Kỹ sư Frontend",
        },
        "",
      ),
    /hồ sơ/,
  )
})

test("Projection và xác nhận mới loại người inactive, không đổi danh mục hoặc lịch sử", () => {
  const state = createDemoWorkspaceState()
  const profiles = deactivateEmployeeProfile(state.profileState, [], [], "hr", "cuong", "off", at)
  const directory = professionalPeople(profiles.profiles, state.organization)
  assert.equal(directory.find((p) => p.id === "cuong")!.active, false)
  assert.equal(
    personnelProfiles(state.projects, state.tasks, "khoa", "", directory).some(
      (p) => p.id === "cuong",
    ),
    false,
  )
  const task = state.tasks.find((t) => t.id === "t3")!
  assert.equal(
    recommendations(state.projects[0], task, directory).some((r) => r.person.id === "cuong"),
    false,
  )
  const draft = {
    [task.id]: {
      personId: "cuong",
      mode: "manual" as const,
      reason: "",
      overrideReason: "",
      hadRun: false,
    },
  }
  assert.match(
    validateDraft(state.projects[0], "khoa", state.tasks, draft, directory).join(),
    /không đủ điều kiện/,
  )
  assert.throws(
    () => commitDraft(state.projects[0], "khoa", state.tasks, draft, directory),
    /không đủ điều kiện/,
  )
  const input: ProjectInput = {
    code: "NEW",
    name: "Dự án",
    description: "",
    pmId: "cuong",
    start: "2026-10-27",
    end: "2026-12-31",
    priority: "Cao",
  }
  assert.match(validateProjectInput(state.projects, input, directory).pmId, /đang hoạt động/)
  assert.throws(
    () => appointPtcm(state.projects[1], "thanh", "cuong", [], [], directory),
    /không còn hoạt động/,
  )
  const nav = selectWorkspaceNavigation(
    state.projects,
    state.tasks,
    "cuong",
    "/employee-profiles",
    directory,
  )
  assert.equal(nav.canLookup, false)
  assert.equal(nav.visibleProjects.length, 0)
})

test("Adapter kiểm tra lại tại lúc lưu, một dispatch; actor inactive bị chặn, reset khôi phục", () => {
  const state = createDemoWorkspaceState()
  const dispatched: DemoWorkspaceAction[] = []
  const actions = createDemoWorkspaceActions(state, (a) => dispatched.push(a), "hr")
  assert.throws(() => actions.deactivateEmployee("dung"), /Chưa thể/)
  assert.equal(dispatched.length, 0)
  actions.deactivateEmployee("ha")
  assert.equal(dispatched.length, 1)
  const next = demoWorkspaceReducer(state, dispatched[0])
  assert.equal(next.tasks, state.tasks)
  assert.equal(next.projects, state.projects)
  assert.equal(next.work, state.work)
  const inactiveActions = createDemoWorkspaceActions(next, (a) => dispatched.push(a), "ha")
  assert.throws(() => inactiveActions.changeTasks(state.tasks), /đã được vô hiệu hóa/)
  assert.throws(
    () =>
      inactiveActions.updateEmployeeProfile(
        state.profileState.profiles[0],
        employeeEditValues(state.profileState.profiles[0]),
      ),
    /đã được vô hiệu hóa/,
  )
  assert.equal(dispatched.length, 1)
  const offCuong = {
    ...state,
    profileState: deactivateEmployeeProfile(state.profileState, [], [], "hr", "cuong", "off", at),
  }
  const ptcm = createDemoWorkspaceActions(
    offCuong,
    () => {
      throw Error("Không được dispatch")
    },
    "khoa",
  )
  const task = state.tasks.find((t) => t.id === "t6")!
  assert.throws(
    () =>
      ptcm.reassignTask({
        projectId: task.projectId,
        taskId: task.id,
        expectedAssignmentId: task.currentAssignment!.id,
        personId: "cuong",
        reason: "Đổi",
        overrideReason: "",
      }),
    /không còn hoạt động/,
  )
  assert.throws(
    () =>
      ptcm.changeTasks(
        state.tasks.map((t) => (t.id === task.id ? { ...t, assigneeId: "cuong" } : t)),
      ),
    /không còn hoạt động/,
  )
  assert.deepEqual(demoWorkspaceReducer(next, { type: "reset" }).profileState, state.profileState)
})
