import test from "node:test"
import assert from "node:assert/strict"
import {
  changeOrganization,
  createOrganizationState,
  organizationReferences,
} from "./organization.ts"
import {
  createProfileState,
  professionalPeople,
  requestProfileChange,
  resolveProfileRequest,
} from "./profiles.ts"
import {
  createDemoWorkspaceState,
  demoWorkspaceReducer,
  type DemoWorkspaceAction,
} from "./workspace-state.ts"
import { createDemoWorkspaceActions } from "./workspace-actions.ts"
import { selectWorkspaceNavigation } from "../app/navigation.ts"
import { personnelProfiles } from "./personnel.ts"
import { recommendations } from "./assignment.ts"
import { isEligible } from "./identity.ts"

test("danh mục mẫu có định danh ổn định, mọi hồ sơ tham chiếu hợp lệ và cấp bậc không thành vị trí", () => {
  const state = createOrganizationState()
  const { profiles } = createProfileState()
  for (const profile of profiles) {
    assert.ok(state.positions.some((entry) => entry.id === profile.positionId))
    assert.ok(
      profile.departmentId === null ||
        state.departments.some((entry) => entry.id === profile.departmentId),
    )
    assert.equal("department" in profile, false)
  }
  assert.equal(
    profiles.find((entry) => entry.personId === "dung")?.positionId,
    profiles.find((entry) => entry.personId === "khoa")?.positionId,
  )
  assert.equal(
    state.positions.some((entry) => entry.name === "Senior"),
    false,
  )
})

test("chỉ HR quản lý; thêm/sửa thuần, tên hợp lệ duy nhất trong cùng danh mục, giữ ID", () => {
  const state = createOrganizationState()
  const { profiles } = createProfileState()
  const snapshot = structuredClone(state)
  const command = { action: "create", kind: "departments", name: "  Vận hành  " } as const
  for (const actor of ["khoa", "thanh", "ha", "dung", "unknown"]) {
    assert.throws(() => changeOrganization(state, profiles, actor, command, "operations"), /Chỉ HR/)
    assert.equal(
      selectWorkspaceNavigation([], [], actor, "/organization").canManageOrganization,
      false,
    )
  }
  assert.equal(selectWorkspaceNavigation([], [], "hr", "/organization").canManageOrganization, true)
  const created = changeOrganization(state, profiles, "hr", command, "operations")
  assert.deepEqual(state, snapshot)
  assert.deepEqual(created.departments.at(-1), { id: "operations", name: "Vận hành" })
  assert.throws(
    () =>
      changeOrganization(created, profiles, "hr", { ...command, name: " vận HÀNH " }, "duplicate"),
    /đã có/,
  )
  for (const name of [" ", "a".repeat(101)])
    assert.throws(
      () => changeOrganization(state, profiles, "hr", { ...command, name }, "invalid"),
      /100 ký tự/,
    )
  const renamed = changeOrganization(
    created,
    profiles,
    "hr",
    {
      action: "update",
      kind: "departments",
      id: "operations",
      beforeName: "Vận hành",
      name: "Điều hành",
    },
    "",
  )
  assert.equal(renamed.departments.at(-1)?.id, "operations")
  assert.equal(renamed.departments.at(-1)?.name, "Điều hành")
  assert.deepEqual(renamed.positions, state.positions)
  assert.throws(
    () =>
      changeOrganization(
        renamed,
        profiles,
        "hr",
        {
          action: "update",
          kind: "departments",
          id: "operations",
          beforeName: "Vận hành",
          name: "Tên cũ",
        },
        "",
      ),
    /Tên đã thay đổi/,
  )
  const separate = changeOrganization(
    state,
    profiles,
    "hr",
    { action: "create", kind: "positions", name: "Nhân sự" },
    "position-hr",
  )
  assert.equal(separate.positions.at(-1)?.name, "Nhân sự")
})

test("xóa chặn mọi hồ sơ tham chiếu; danh mục trống và định danh/tên cũ được xử lý rõ", () => {
  const state = createOrganizationState()
  const { profiles } = createProfileState()
  for (const kind of ["departments", "positions"] as const) {
    for (const entry of state[kind]) {
      assert.ok(organizationReferences(profiles, kind, entry.id).length)
      assert.throws(
        () =>
          changeOrganization(
            state,
            profiles,
            "hr",
            { action: "delete", kind, id: entry.id, beforeName: entry.name },
            "",
          ),
        /hồ sơ nhân viên đang sử dụng/,
      )
    }
    const empty = { ...state, [kind]: [] }
    const created = changeOrganization(
      empty,
      profiles,
      "hr",
      { action: "create", kind, name: "Mục chưa dùng" },
      "unused",
    )
    const removed = changeOrganization(
      created,
      profiles,
      "hr",
      { action: "delete", kind, id: "unused", beforeName: "Mục chưa dùng" },
      "",
    )
    assert.deepEqual(removed[kind], [])
    assert.throws(
      () =>
        changeOrganization(
          removed,
          profiles,
          "hr",
          { action: "delete", kind, id: "unused", beforeName: "Mục chưa dùng" },
          "",
        ),
      /Không tìm thấy/,
    )
    assert.throws(
      () =>
        changeOrganization(
          created,
          profiles,
          "hr",
          { action: "delete", kind, id: "unused", beforeName: "Sai tên" },
          "",
        ),
      /Tên đã thay đổi/,
    )
  }
})

test("đổi tên vị trí phản ánh tra cứu/gợi ý; không đổi trách nhiệm, kỹ năng, eligibility hay cấp bậc", () => {
  const state = createDemoWorkspaceState()
  const before = professionalPeople(state.profileState.profiles, state.organization)
  const organization = changeOrganization(
    state.organization,
    state.profileState.profiles,
    "hr",
    {
      action: "update",
      kind: "positions",
      id: "software-engineer",
      beforeName: "Kỹ sư phần mềm",
      name: "Kỹ sư phát triển phần mềm",
    },
    "",
  )
  const directory = professionalPeople(state.profileState.profiles, organization)
  const dung = directory.find((person) => person.id === "dung")!
  assert.equal(dung.position, "Kỹ sư phát triển phần mềm")
  assert.equal(dung.seniority, before.find((person) => person.id === "dung")?.seniority)
  assert.deepEqual(dung.skills, before.find((person) => person.id === "dung")?.skills)
  assert.equal(
    isEligible(state.projects[0], dung),
    isEligible(
      state.projects[0],
      before.find((person) => person.id === "dung")!,
    ),
  )
  assert.equal(
    personnelProfiles(state.projects, state.tasks, "khoa", "phát triển", directory).some(
      (person) => person.id === "dung",
    ),
    true,
  )
  const task = state.tasks.find((entry) =>
    entry.requirements.some((requirement) => requirement.skill === "Node.js"),
  )!
  const oldRank = recommendations(state.projects[0], task, before)
  const newRank = recommendations(state.projects[0], task, directory)
  assert.deepEqual(
    newRank.map((entry) => entry.person.id),
    oldRank.map((entry) => entry.person.id),
  )
  assert.equal(newRank.find((entry) => entry.person.id === "dung")?.person.position, dung.position)
  for (const key of ["departmentId", "positionId", "phone", "address", "companyEmail"])
    assert.equal(key in dung, false)
  const profile = state.profileState.profiles.find((entry) => entry.personId === "dung")!
  const pending = requestProfileChange(
    state.profileState,
    "dung",
    profile,
    { ...profile, fullname: "Dũng Mới" },
    "",
    "r",
    "2026-10-26T08:00:00Z",
  )
  const approved = resolveProfileRequest(pending, "hr", "r", "APPROVED", "", "2026-10-26T09:00:00Z")
  assert.equal(
    professionalPeople(approved.profiles, organization).find((person) => person.id === "dung")
      ?.position,
    dung.position,
  )
})

test("action lỗi không dispatch; cập nhật không đổi profile/request và reset khôi phục cả tham chiếu", () => {
  const state = createDemoWorkspaceState()
  const actions: DemoWorkspaceAction[] = []
  const hr = createDemoWorkspaceActions(state, (action) => actions.push(action), "hr")
  assert.throws(
    () =>
      hr.changeOrganization({
        action: "delete",
        kind: "positions",
        id: "director",
        beforeName: "Giám đốc",
      }),
    /đang sử dụng/,
  )
  assert.equal(actions.length, 0)
  hr.changeOrganization({ action: "create", kind: "departments", name: "Vận hành" })
  assert.equal(actions.length, 1)
  const next = demoWorkspaceReducer(state, actions[0])
  assert.equal(next.organization.departments.length, state.organization.departments.length + 1)
  assert.equal(next.profileState, state.profileState)
  assert.equal(next.projects, state.projects)
  const reset = demoWorkspaceReducer(next, { type: "reset" })
  assert.deepEqual(reset.organization, state.organization)
  assert.deepEqual(reset.profileState, state.profileState)
})

test("cấp bậc do HR đặt tên theo công ty, không suy từ vị trí/quyền và không xóa khi đang dùng", () => {
  const { profiles } = createProfileState()
  const initial = createOrganizationState()
  for (const profile of profiles)
    assert.ok(
      profile.jobLevelId === null ||
        initial.jobLevels.some((entry) => entry.id === profile.jobLevelId),
    )
  for (const actor of ["dung", "khoa", "thanh", "ha"])
    assert.throws(
      () =>
        changeOrganization(
          initial,
          profiles,
          actor,
          { action: "create", kind: "jobLevels", name: "Chuyên viên bậc 3" },
          "level-3",
        ),
      /Chỉ HR/,
    )
  const created = changeOrganization(
    initial,
    profiles,
    "hr",
    { action: "create", kind: "jobLevels", name: "  Chuyên viên bậc 3  " },
    "level-3",
  )
  assert.equal(created.jobLevels.find((entry) => entry.id === "level-3")!.name, "Chuyên viên bậc 3")
  for (const name of [" ", "x".repeat(101), "CHUYÊN VIÊN BẬC 3"])
    assert.throws(() =>
      changeOrganization(
        created,
        profiles,
        "hr",
        { action: "create", kind: "jobLevels", name },
        "another",
      ),
    )
  assert.throws(
    () =>
      changeOrganization(
        created,
        profiles,
        "hr",
        { action: "delete", kind: "jobLevels", id: "middle", beforeName: "Middle" },
        "",
      ),
    /hồ sơ/,
  )
  const removed = changeOrganization(
    created,
    profiles,
    "hr",
    { action: "delete", kind: "jobLevels", id: "level-3", beforeName: "Chuyên viên bậc 3" },
    "",
  )
  assert.deepEqual(removed, initial)
  const renamed = changeOrganization(
    initial,
    profiles,
    "hr",
    {
      action: "update",
      kind: "jobLevels",
      id: "middle",
      beforeName: "Middle",
      name: "Bậc chuyên viên",
    },
    "",
  )
  const before = professionalPeople(profiles, initial).find((entry) => entry.id === "dung")!
  const after = professionalPeople(profiles, renamed).find((entry) => entry.id === "dung")!
  assert.equal(after.seniority, "Bậc chuyên viên")
  assert.equal(after.position, before.position)
  assert.deepEqual(after.skills, before.skills)
  assert.equal(after.active, before.active)
  assert.equal(after.backgroundLoad, before.backgroundLoad)
})
