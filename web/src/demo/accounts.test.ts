import assert from "node:assert/strict"
import test from "node:test"
import {
  createDemoWorkspaceState,
  demoWorkspaceReducer,
  type DemoWorkspaceAction,
} from "./workspace-state.ts"
import { createDemoWorkspaceActions } from "./workspace-actions.ts"
import {
  createAccountWithProfile,
  accountCreateErrors,
  accountOpenWork,
  nextAccountIdentifiers,
  changeAccount,
} from "./accounts.ts"
import { professionalPeople } from "./profiles.ts"
import { isAdmin, isHr, isDirector, canViewProject } from "./identity.ts"
import { employeeDeactivationBlockers } from "./employee-deactivation.ts"
import { selectWorkspaceNavigation, demoIdentityOptions } from "../app/navigation.ts"
import { personalTasks } from "./execution.ts"
import { createProject } from "./projects.ts"
import { saveLog } from "./effort.ts"
import type { AccountCreateInput, AccountCommand } from "../models/accounts.ts"

const at = "2026-10-28T08:00:00.000Z"
const input: AccountCreateInput = {
  fullname: "Nguyễn Demo",
  companyEmail: "demo@congty.example",
  positionId: "software-engineer",
  departmentId: null,
  jobLevelId: null,
  phone: null,
  systemRoles: [],
}
function fixture() {
  let state = createDemoWorkspaceState()
  const dispatch = (action: DemoWorkspaceAction) => {
    state = demoWorkspaceReducer(state, action)
  }
  return {
    get state() {
      return state
    },
    act: (actor: string) => createDemoWorkspaceActions(state, dispatch, actor),
    set: (action: DemoWorkspaceAction) => dispatch(action),
  }
}
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
function command(
  state: ReturnType<typeof createDemoWorkspaceState>,
  personId: string,
  change:
    { action: "lock" | "unlock" } | { action: "roles"; roles: AccountCreateInput["systemRoles"] },
): AccountCommand {
  const a = state.accounts.accounts.find((a) => a.personId === personId)!
  return {
    accountId: a.id,
    before: { status: a.status, systemRoles: [...a.systemRoles] },
    ...change,
  }
}

test("tạo tài khoản và hồ sơ một action, không sửa fixture, HR/projection/danh tính cùng thấy người mới", () => {
  const f = fixture(),
    before = structuredClone(f.state)
  f.act("admin").createAccount(input)
  const account = f.state.accounts.accounts.at(-1)!
  const profile = f.state.profileState.profiles.at(-1)!
  assert.equal(profile.personId, account.personId)
  assert.equal(profile.address, null)
  assert.equal(profile.positionId, input.positionId)
  assert.equal(profile.fullname, input.fullname)
  assert.equal(f.state.accounts.events.at(-1)?.actorId, "admin")
  assert.equal(f.state.accounts.events.at(-1)?.action, "CREATED")
  assert.equal(identity(f.state).people.at(-1)?.name, input.fullname)
  assert.ok(
    demoIdentityOptions(
      f.state.projects,
      undefined,
      identity(f.state).people,
      identity(f.state),
    ).some((i) => i.id === profile.personId),
  )
  assert.deepEqual(f.state.projects, before.projects)
  assert.deepEqual(f.state.employeeSkills, before.employeeSkills)
  assert.deepEqual(createDemoWorkspaceState(), before)
  assert.equal(isAdmin("admin", identity(f.state)), true)
  assert.equal(canViewProject(f.state.projects[0], "admin", identity(f.state)), false)
  assert.equal(
    selectWorkspaceNavigation(
      f.state.projects,
      f.state.tasks,
      "admin",
      "/accounts",
      identity(f.state).people,
      identity(f.state),
    ).canManageAccounts,
    true,
  )
})

test("dữ liệu tạo sai/trùng/danh mục cũ không tạo dở; chỉ Admin được tạo", () => {
  const f = fixture(),
    before = structuredClone(f.state)
  for (const patch of [
    { companyEmail: "hr@congty.example" },
    { fullname: " " },
    { companyEmail: "invalid" },
    { positionId: "missing" },
    { departmentId: "missing" },
    { jobLevelId: "missing" },
    { phone: "x".repeat(21) },
  ]) {
    assert.ok(
      Object.keys(
        accountCreateErrors(f.state.profileState, f.state.organization, {
          ...input,
          ...patch,
        }),
      ).length,
    )
    assert.throws(() => f.act("admin").createAccount({ ...input, ...patch }))
    assert.deepEqual(f.state, before)
  }
  for (const actor of ["hr", "ha", "khoa", "missing"])
    assert.throws(() => f.act(actor).createAccount(input), /Quản trị viên/)
  assert.throws(
    () => f.act("admin").createAccount({ ...input, systemRoles: ["PM" as never] }),
    /Quyền hệ thống/,
  )
  const result = createAccountWithProfile(
    f.state.accounts,
    f.state.profileState,
    f.state.organization,
    "admin",
    input,
    "person",
    "account",
    "event",
    at,
  )
  assert.equal(result.profileState.profiles.at(-1)?.personId, "person")
})

test("cấp/thu hồi role cập nhật quyền và helper thật của demo, không chỉ sidebar; trách nhiệm không đổi", () => {
  const f = fixture()
  f.act("admin").createAccount(input)
  const id = f.state.accounts.accounts.at(-1)!.personId
  const projects = structuredClone(f.state.projects)
  f.act("admin").changeAccount(command(f.state, id, { action: "roles", roles: ["HR", "DIRECTOR"] }))
  assert.equal(isHr(id, identity(f.state)), true)
  assert.equal(isDirector(id, identity(f.state)), true)
  f.act(id).changeOrganization({ action: "create", kind: "departments", name: "Demo department" })
  assert.ok(f.state.organization.departments.some((d) => d.name === "Demo department"))
  const p = f.state.projects[0]
  const created = createProject(
    f.state.projects,
    id,
    {
      code: "DEMO",
      name: "Dự án demo",
      description: "",
      pmId: "thanh",
      start: "2026-10-28",
      end: "2026-12-31",
      priority: "Thấp",
    },
    identity(f.state).people,
    identity(f.state),
  )
  assert.equal(created.at(-1)?.pmId, "thanh")
  assert.equal(created.at(-1)?.ptcmId, undefined)
  assert.equal(p.pmId, projects[0].pmId)
  f.act("admin").changeAccount(command(f.state, id, { action: "roles", roles: [] }))
  assert.equal(isHr(id, identity(f.state)), false)
  assert.equal(isDirector(id, identity(f.state)), false)
  assert.throws(
    () =>
      f.act(id).changeOrganization({ action: "create", kind: "departments", name: "Không được" }),
    /Chỉ HR/,
  )
  assert.throws(
    () =>
      createProject(
        f.state.projects,
        id,
        {
          code: "DEMO",
          name: "Dự án demo",
          description: "",
          pmId: "thanh",
          start: "2026-10-28",
          end: "2026-12-31",
          priority: "Thấp",
        },
        identity(f.state).people,
        identity(f.state),
      ),
    /Giám đốc/,
  )
  assert.deepEqual(f.state.projects, projects)
})

test("Admin cuối được bảo vệ khi khóa, thu hồi role hoặc HR vô hiệu hóa, kể cả sau thay người", () => {
  const f = fixture()
  assert.throws(
    () => f.act("admin").changeAccount(command(f.state, "admin", { action: "lock" })),
    /cuối cùng/,
  )
  assert.throws(
    () => f.act("admin").changeAccount(command(f.state, "admin", { action: "roles", roles: [] })),
    /cuối cùng/,
  )
  assert.throws(() => f.act("hr").deactivateEmployee("admin"), /cuối cùng/)
  f.act("admin").createAccount({ ...input, systemRoles: ["ADMIN"] })
  const id = f.state.accounts.accounts.at(-1)!.personId
  f.act("admin").changeAccount(command(f.state, "admin", { action: "roles", roles: [] }))
  assert.throws(() => f.act("admin").createAccount(input), /Quản trị viên/)
  assert.throws(
    () => f.act(id).changeAccount(command(f.state, id, { action: "lock" })),
    /cuối cùng/,
  )
  assert.ok(
    employeeDeactivationBlockers(
      id,
      f.state.projects,
      f.state.tasks,
      f.state.profileState.profiles,
      identity(f.state),
    ).some((b) => b.kind === "admin"),
  )
  assert.throws(() => f.act("hr").deactivateEmployee(id), /cuối cùng/)
})

test("khóa còn task/trách nhiệm không thu hồi ngầm; actor/candidate bị chặn, mở khóa không hồi sinh hồ sơ", () => {
  const f = fixture(),
    before = structuredClone(f.state)
  assert.ok(accountOpenWork("khoa", f.state.projects, f.state.tasks).length)
  f.act("admin").changeAccount(command(f.state, "khoa", { action: "lock" }))
  assert.deepEqual(f.state.tasks, before.tasks)
  assert.deepEqual(f.state.projects, before.projects)
  assert.equal(identity(f.state).people.find((p) => p.id === "khoa")?.active, false)
  assert.throws(
    () => f.act("khoa").addCv("khoa", { name: "cv.pdf", size: 10, type: "application/pdf" }),
    /hoạt động/,
  )
  assert.throws(() => f.act("khoa").changeTasks(f.state.tasks), /vô hiệu hóa/)
  f.act("admin").changeAccount(command(f.state, "khoa", { action: "unlock" }))
  assert.equal(identity(f.state).people.find((p) => p.id === "khoa")?.active, true)
  f.act("admin").createAccount(input)
  const id = f.state.accounts.accounts.at(-1)!.personId
  f.act("admin").changeAccount(command(f.state, id, { action: "lock" }))
  f.act("hr").deactivateEmployee(id)
  f.act("admin").changeAccount(command(f.state, id, { action: "unlock" }))
  assert.equal(f.state.accounts.accounts.at(-1)?.status, "ACTIVE")
  assert.equal(f.state.profileState.profiles.at(-1)?.employmentStatus, "INACTIVE")
  assert.equal(identity(f.state).people.find((p) => p.id === id)?.active, false)
})

test("người mới dùng CV/hồ sơ và thực hiện task qua nguồn hiện hành, reset không giữ tài khoản lẻ", () => {
  const f = fixture(),
    original = structuredClone(f.state)
  f.act("admin").createAccount(input)
  const id = f.state.accounts.accounts.at(-1)!.personId
  f.act(id).addCv(id, { name: "cv.pdf", size: 10, type: "application/pdf" })
  f.act(id).selectCurrentCv(f.state.employeeCvs.versions.at(-1)!.id)
  f.act(id).requestProfileChange(
    { fullname: input.fullname, phone: null, address: null },
    { fullname: input.fullname, phone: null, address: "Địa chỉ demo" },
    "",
  )
  f.act("hr").resolveProfileRequest(f.state.profileState.requests.at(-1)!.id, "APPROVED", "")
  assert.equal(f.state.profileState.profiles.at(-1)?.address, "Địa chỉ demo")
  const task = f.state.tasks.find(
    (t) =>
      t.status === "ASSIGNED" &&
      f.state.projects.find((p) => p.id === t.projectId)?.status === "IN_PROGRESS",
  )!
  const assigned = {
    ...task,
    assigneeId: id,
    currentAssignment: {
      id: "new-assignment",
      employeeId: id,
      assignedAt: "2026-10-27T01:00:00.000Z",
    },
  }
  f.act("khoa").changeTasks(f.state.tasks.map((t) => (t.id === task.id ? assigned : t)))
  assert.equal(personalTasks(f.state.projects, f.state.tasks, id, identity(f.state)).length, 1)
  f.act(id).transitionTask(task.id, "ASSIGNED", "start")
  const log = saveLog(
    f.state.projects,
    f.state.tasks,
    f.state.work,
    id,
    task.id,
    { workDate: "2026-10-28", minutes: 30, description: "Thực hiện demo", overLimitReason: "" },
    undefined,
    new Date(at),
    identity(f.state),
  )
  assert.ok(log.logs.some((l) => l.employeeId === id))
  f.act(id).changeWork(log)
  f.act(id).submitResult(task.id, "Kết quả demo", "")
  assert.equal(f.state.tasks.find((t) => t.id === task.id)?.status, "PENDING_REVIEW")
  f.act("admin").reset()
  assert.deepEqual(f.state.accounts, original.accounts)
  assert.deepEqual(f.state.profileState, original.profileState)
})

test("nền vai trò/trạng thái cũ và actor bị khóa không ghi đè hoặc ghi một phần", () => {
  const f = fixture()
  const old = command(f.state, "dung", { action: "roles", roles: ["HR"] })
  f.act("admin").changeAccount(command(f.state, "dung", { action: "lock" }))
  const before = structuredClone(f.state)
  assert.throws(() => f.act("admin").changeAccount(old), /đã thay đổi/)
  assert.deepEqual(f.state, before)
  assert.throws(
    () =>
      changeAccount(
        f.state.accounts,
        f.state.profileState,
        f.state.organization,
        "dung",
        command(f.state, "dung", { action: "unlock" }),
        "event",
        at,
      ),
    /Quản trị viên/,
  )
})

test("khóa tài khoản không vô hiệu hóa hồ sơ hoặc chặn HR quản lý thay người đó", () => {
  const f = fixture()
  f.act("dung").requestProfileChange(
    { fullname: "Phạm Quang Dũng", phone: null, address: null },
    { fullname: "Phạm Quang Dũng", phone: null, address: "Địa chỉ demo" },
    "",
  )
  f.act("dung").requestSkillChange({
    skillId: "React",
    operation: "UPDATE",
    before: { level: 2, years: null, note: "" },
    proposed: { level: 3, years: null, note: "" },
    explanation: "Đối chiếu thực hành",
    externalEvidence: "Nhận xét chuyên môn",
    evidence: [],
  })
  f.act("admin").changeAccount(command(f.state, "dung", { action: "lock" }))
  f.act("hr").addCv("dung", { name: "cv.pdf", size: 10, type: "application/pdf" })
  f.act("hr").resolveProfileRequest(f.state.profileState.requests.at(-1)!.id, "APPROVED", "")
  f.act("ha").resolveSkillChange({
    requestId: f.state.employeeSkills.requests.at(-1)!.id,
    version: 1,
    decision: "APPROVED",
    reason: "Đã đối chiếu",
    expertOpinion: "Đánh giá thực hành",
  })
  assert.equal(
    f.state.profileState.profiles.find((p) => p.personId === "dung")?.employmentStatus,
    "ACTIVE",
  )
  assert.equal(
    f.state.profileState.profiles.find((p) => p.personId === "dung")?.address,
    "Địa chỉ demo",
  )
  assert.equal(
    f.state.employeeSkills.official.find((s) => s.employeeId === "dung" && s.skillId === "React")
      ?.level,
    3,
  )
  assert.equal(identity(f.state).people.find((p) => p.id === "dung")?.active, false)
  assert.throws(
    () => f.act("dung").selectCurrentCv(f.state.employeeCvs.versions.at(-1)!.id),
    /hoạt động/,
  )
})

test("tự cấp mã chỉ đọc: không lấp mã trống, không dùng lại mã khóa/inactive, tính lại khi lưu", () => {
  const f = fixture()
  const preview = nextAccountIdentifiers(f.state.accounts, f.state.profileState)
  assert.deepEqual(preview, { employeeCode: "E010", username: "e010" })
  f.act("admin").createAccount(input)
  assert.equal(f.state.profileState.profiles.at(-1)?.employeeCode, preview.employeeCode)
  assert.equal(f.state.accounts.accounts.at(-1)?.username, preview.username)
  const id = f.state.accounts.accounts.at(-1)!.personId
  f.act("admin").changeAccount(command(f.state, id, { action: "lock" }))
  f.act("hr").deactivateEmployee(id)
  f.act("admin").createAccount({ ...input, companyEmail: "second@congty.example" })
  assert.equal(f.state.profileState.profiles.at(-1)?.employeeCode, "E011")
  assert.equal(f.state.accounts.accounts.at(-1)?.username, "e011")
  const accounts = structuredClone(f.state.accounts)
  accounts.accounts[0].username = "E1000"
  assert.deepEqual(nextAccountIdentifiers(accounts, f.state.profileState), {
    employeeCode: "E1001",
    username: "e1001",
  })
})
