import test from "node:test"
import assert from "node:assert/strict"
import { employeeEditValues, updateEmployeeProfile } from "./employee-administration.ts"
import {
  createProfileState,
  professionalPeople,
  requestProfileChange,
  resolveProfileRequest,
} from "./profiles.ts"
import { createOrganizationState, changeOrganization } from "./organization.ts"
import {
  createDemoWorkspaceState,
  demoWorkspaceReducer,
  type DemoWorkspaceAction,
} from "./workspace-state.ts"
import { createDemoWorkspaceActions } from "./workspace-actions.ts"
const at = "2026-10-26T09:00:00Z"
function fixture() {
  const state = createProfileState(),
    organization = createOrganizationState()
  const before = state.profiles.find((profile) => profile.personId === "dung")!
  return { state, organization, before, values: employeeEditValues(before) }
}
test("HR sửa trực tiếp, giữ lịch sử và projection không lộ liên hệ", () => {
  const { state, organization, before, values } = fixture()
  const original = structuredClone(state)
  for (const actor of ["dung", "khoa", "thanh", "ha", "missing"])
    assert.throws(
      () =>
        updateEmployeeProfile(
          state,
          organization,
          actor,
          before,
          { ...values, fullname: "Tên mới" },
          "edit",
          at,
        ),
      /Chỉ HR/,
    )
  const next = updateEmployeeProfile(
    state,
    organization,
    "hr",
    before,
    {
      ...values,
      fullname: "  Phạm Dũng mới  ",
      companyEmail: "dung.new@congty.example",
      phone: " 0901234567 ",
      address: "  Hà Nội ",
      departmentId: null,
      positionId: "backend-engineer",
      jobLevelId: "senior",
    },
    "edit",
    at,
  )
  assert.deepEqual(state, original)
  const profile = next.profiles.find((entry) => entry.personId === before.personId)!
  assert.equal(profile.employeeCode, before.employeeCode)
  assert.equal(profile.fullname, "Phạm Dũng mới")
  assert.equal(profile.address, "Hà Nội")
  assert.equal(next.requests, state.requests)
  assert.equal(next.edits[0].actorId, "hr")
  assert.equal(next.edits[0].at, at)
  assert.equal(next.edits[0].before.address, null)
  assert.equal(next.edits[0].after.address, "Hà Nội")
  assert.equal(next.edits[0].beforeDisplay.positionId, "Kỹ sư phần mềm")
  assert.equal(next.edits[0].afterDisplay.positionId, "Kỹ sư Backend")
  const renamed = changeOrganization(
    organization,
    next.profiles,
    "hr",
    {
      action: "update",
      kind: "positions",
      id: "backend-engineer",
      beforeName: "Kỹ sư Backend",
      name: "Backend mới",
    },
    "",
  )
  assert.equal(
    professionalPeople(next.profiles, renamed).find((entry) => entry.id === "dung")!.position,
    "Backend mới",
  )
  assert.equal(next.edits[0].afterDisplay.positionId, "Kỹ sư Backend")
  const projected = professionalPeople(next.profiles, organization).find(
    (entry) => entry.id === "dung",
  )!
  assert.equal(projected.name, profile.fullname)
  assert.equal(projected.position, "Kỹ sư Backend")
  assert.equal(projected.seniority, "Senior")
  for (const key of ["companyEmail", "phone", "address", "departmentId", "positionId"])
    assert.equal(key in projected, false)
  assert.deepEqual(
    projected.skills,
    professionalPeople(state.profiles, organization).find((entry) => entry.id === "dung")!.skills,
  )
})
test("Validation email, danh mục, giới hạn và bảo toàn địa chỉ", () => {
  const { state, organization, before, values } = fixture()
  for (const patch of [
    { fullname: " " },
    { fullname: "a".repeat(101) },
    { companyEmail: " " },
    { companyEmail: "bad@" },
    { companyEmail: "HR@CONGTY.EXAMPLE" },
    { phone: "1".repeat(21) },
    { jobLevelId: "missing" },
    { positionId: "" },
    { positionId: "missing" },
    { departmentId: "missing" },
  ])
    assert.throws(() =>
      updateEmployeeProfile(state, organization, "hr", before, { ...values, ...patch }, "edit", at),
    )
  assert.throws(
    () => updateEmployeeProfile(state, organization, "hr", before, values, "edit", at),
    /Chưa có/,
  )
  const added = updateEmployeeProfile(
    state,
    organization,
    "hr",
    before,
    { ...values, address: "Hà Nội" },
    "first",
    at,
  )
  const current = added.profiles.find((profile) => profile.personId === before.personId)!
  for (const address of [null, "", "   "])
    assert.throws(
      () =>
        updateEmployeeProfile(
          added,
          organization,
          "hr",
          current,
          { ...employeeEditValues(current), address },
          "second",
          at,
        ),
      /Địa chỉ/,
    )
  const changedPhone = updateEmployeeProfile(
    state,
    organization,
    "hr",
    before,
    { ...values, phone: "0901234567" },
    "phone",
    at,
  )
  assert.equal(changedPhone.profiles.find((profile) => profile.personId === "dung")!.address, null)
  const malformed = {
    ...values,
    fullname: "Dũng mới",
    employeeCode: "HACK",
    personId: "hr",
    systemRoles: ["HR"],
  }
  const safe = updateEmployeeProfile(state, organization, "hr", before, malformed, "safe", at)
  const saved = safe.profiles.find((profile) => profile.personId === before.personId)!
  assert.equal(saved.employeeCode, before.employeeCode)
  assert.equal("systemRoles" in saved, false)
})
test("Giữ pending, chặn duyệt nền cũ; sửa trường khác không ghi đè", () => {
  const { state, organization, before, values } = fixture()
  const pending = requestProfileChange(
    state,
    "dung",
    before,
    { ...before, phone: "0901234567" },
    "",
    "request",
    at,
  )
  const changed = updateEmployeeProfile(
    pending,
    organization,
    "hr",
    before,
    { ...values, phone: "0912345678" },
    "edit",
    at,
  )
  assert.equal(changed.requests[0].status, "PENDING")
  assert.throws(
    () => resolveProfileRequest(changed, "hr", "request", "APPROVED", "", at),
    /Thông tin nền/,
  )
  const rejected = resolveProfileRequest(
    changed,
    "hr",
    "request",
    "REJECTED",
    "Hồ sơ đã cập nhật",
    at,
  )
  assert.deepEqual(rejected.edits, changed.edits)
  const unrelated = updateEmployeeProfile(
    pending,
    organization,
    "hr",
    before,
    { ...values, jobLevelId: "senior" },
    "other",
    at,
  )
  const approved = resolveProfileRequest(unrelated, "hr", "request", "APPROVED", "", at)
  assert.equal(
    approved.profiles.find((profile) => profile.personId === "dung")!.jobLevelId,
    "senior",
  )
  assert.equal(
    approved.profiles.find((profile) => profile.personId === "dung")!.phone,
    "0901234567",
  )
  assert.deepEqual(approved.edits, unrelated.edits)
  assert.throws(
    () =>
      updateEmployeeProfile(
        changed,
        organization,
        "hr",
        before,
        { ...values, phone: "0922222222" },
        "stale",
        at,
      ),
    /Hồ sơ đã thay đổi/,
  )
})
test("Adapter một dispatch, reset lịch sử và cập nhật tham chiếu danh mục", () => {
  const state = createDemoWorkspaceState()
  const before = state.profileState.profiles.find((entry) => entry.personId === "hr")!
  const dispatched: DemoWorkspaceAction[] = []
  const actions = createDemoWorkspaceActions(state, (action) => dispatched.push(action), "hr")
  assert.throws(() =>
    actions.updateEmployeeProfile(before, { ...employeeEditValues(before), positionId: "" }),
  )
  assert.equal(dispatched.length, 0)
  actions.updateEmployeeProfile(before, {
    ...employeeEditValues(before),
    departmentId: "engineering",
    jobLevelId: null,
  })
  assert.equal(dispatched.length, 1)
  const next = demoWorkspaceReducer(state, dispatched[0])
  assert.equal(next.projects, state.projects)
  assert.equal(next.tasks, state.tasks)
  assert.equal(next.profileState.edits.length, 1)
  const removed = changeOrganization(
    next.organization,
    next.profileState.profiles,
    "hr",
    { action: "delete", kind: "departments", id: "human-resources", beforeName: "Nhân sự" },
    "",
  )
  assert.equal(removed.departments.length, 1)
  assert.deepEqual(demoWorkspaceReducer(next, { type: "reset" }).profileState, state.profileState)
})

test("HR chọn ID cấp bậc hoặc chưa cập nhật; lịch sử giữ tên lúc lưu và request cá nhân không đổi cấp bậc", () => {
  const { state, organization, before, values } = fixture()
  const catalog = changeOrganization(
    organization,
    state.profiles,
    "hr",
    { action: "create", kind: "jobLevels", name: "Chuyên viên bậc 3" },
    "level-3",
  )
  const next = updateEmployeeProfile(
    state,
    catalog,
    "hr",
    before,
    { ...values, jobLevelId: "level-3" },
    "edit",
    at,
  )
  assert.equal(next.profiles.find((entry) => entry.personId === "dung")!.jobLevelId, "level-3")
  assert.equal(next.edits[0].beforeDisplay.jobLevelId, "Middle")
  assert.equal(next.edits[0].afterDisplay.jobLevelId, "Chuyên viên bậc 3")
  const renamed = changeOrganization(
    catalog,
    next.profiles,
    "hr",
    {
      action: "update",
      kind: "jobLevels",
      id: "level-3",
      beforeName: "Chuyên viên bậc 3",
      name: "Chuyên viên chính",
    },
    "",
  )
  assert.equal(
    professionalPeople(next.profiles, renamed).find((entry) => entry.id === "dung")!.seniority,
    "Chuyên viên chính",
  )
  assert.equal(next.edits[0].afterDisplay.jobLevelId, "Chuyên viên bậc 3")
  const current = next.profiles.find((entry) => entry.personId === "dung")!
  assert.throws(
    () =>
      updateEmployeeProfile(
        next,
        renamed,
        "hr",
        current,
        { ...employeeEditValues(current), jobLevelId: "not-in-catalog" },
        "bad",
        at,
      ),
    /Cấp bậc/,
  )
  const cleared = updateEmployeeProfile(
    next,
    renamed,
    "hr",
    current,
    { ...employeeEditValues(current), jobLevelId: null },
    "clear",
    at,
  )
  assert.equal(
    professionalPeople(cleared.profiles, renamed).find((entry) => entry.id === "dung")!.seniority,
    "Chưa cập nhật",
  )
  const removed = changeOrganization(
    renamed,
    cleared.profiles,
    "hr",
    { action: "delete", kind: "jobLevels", id: "level-3", beforeName: "Chuyên viên chính" },
    "",
  )
  assert.equal(
    removed.jobLevels.some((entry) => entry.id === "level-3"),
    false,
  )
  assert.equal(cleared.edits[0].afterDisplay.jobLevelId, "Chuyên viên bậc 3")
  const proposedWithLevel = { ...before, phone: "0901234567", jobLevelId: "senior" }
  const pending = requestProfileChange(state, "dung", before, proposedWithLevel, "", "request", at)
  assert.equal("jobLevelId" in pending.requests[0].proposed, false)
  assert.equal(
    resolveProfileRequest(pending, "hr", "request", "APPROVED", "", at).profiles.find(
      (entry) => entry.personId === "dung",
    )!.jobLevelId,
    before.jobLevelId,
  )
})
