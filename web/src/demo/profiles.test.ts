import test from "node:test"
import assert from "node:assert/strict"
import {
  createProfileState,
  requestProfileChange,
  resolveProfileRequest,
  professionalPeople,
  profilePatch,
} from "./profiles.ts"
import {
  createDemoWorkspaceState,
  demoWorkspaceReducer,
  type DemoWorkspaceAction,
} from "./workspace-state.ts"
import { createDemoWorkspaceActions } from "./workspace-actions.ts"
import { personnelProfiles } from "./personnel.ts"
import { recommendations } from "./assignment.ts"
import { demoIdentityOptions, selectWorkspaceNavigation } from "../app/navigation.ts"
import { createOrganizationState } from "./organization.ts"

const at = "2026-10-26T08:00:00.000Z"
const setup = () => {
  const state = createProfileState()
  const profile = state.profiles.find((entry) => entry.personId === "dung")!
  return { state, profile }
}

test("pending không sửa hồ sơ; HR duyệt một lần, giữ bản trước/đề nghị và người/thời điểm", () => {
  const { state, profile } = setup()
  const before = structuredClone(state)
  const pending = requestProfileChange(
    state,
    "dung",
    profile,
    { ...profile, fullname: "Phạm Quang Dũng Mới", address: "  Hà Nội  " },
    " Bổ sung hồ sơ ",
    "r1",
    at,
  )
  assert.deepEqual(state, before)
  assert.deepEqual(pending.profiles, state.profiles)
  assert.deepEqual(pending.requests[0].before, { fullname: profile.fullname, address: null })
  assert.deepEqual(pending.requests[0].proposed, {
    fullname: "Phạm Quang Dũng Mới",
    address: "Hà Nội",
  })
  const next = resolveProfileRequest(pending, "hr", "r1", "APPROVED", "", at)
  assert.equal(next.profiles.find((entry) => entry.personId === "dung")?.address, "Hà Nội")
  assert.equal(next.requests[0].resolvedBy, "hr")
  assert.equal(next.requests[0].resolvedAt, at)
  assert.deepEqual(next.requests[0].before, pending.requests[0].before)
  assert.throws(
    () => resolveProfileRequest(next, "hr", "r1", "REJECTED", "Đổi ý", at),
    /đã được xử lý/,
  )
})

test("chỉ HR xử lý; từ chối cần lý do và không đổi hồ sơ, người thường không đọc hàng đợi qua navigation", () => {
  const { state, profile } = setup()
  const pending = requestProfileChange(
    state,
    "dung",
    profile,
    { ...profile, phone: "0901234567" },
    "",
    "r1",
    at,
  )
  for (const actor of ["dung", "khoa", "thanh", "ha", "unknown"]) {
    assert.throws(() => resolveProfileRequest(pending, actor, "r1", "APPROVED", "", at), /Chỉ HR/)
  }
  assert.throws(() => resolveProfileRequest(pending, "hr", "r1", "REJECTED", " ", at), /lý do/)
  const rejected = resolveProfileRequest(
    pending,
    "hr",
    "r1",
    "REJECTED",
    " Cần kiểm tra lại số điện thoại ",
    at,
  )
  assert.deepEqual(rejected.profiles, state.profiles)
  assert.equal(rejected.requests[0].decisionReason, "Cần kiểm tra lại số điện thoại")
  assert.equal(
    selectWorkspaceNavigation([], [], "dung", "/profile-requests").canReviewProfile,
    false,
  )
  assert.equal(selectWorkspaceNavigation([], [], "hr", "/profile-requests").canReviewProfile, true)
})

test("nhiều pending hợp lệ; đối chiếu theo trường, không ghi đè nền đã đổi nhưng vẫn được từ chối", () => {
  const { state, profile } = setup()
  const first = requestProfileChange(
    state,
    "dung",
    profile,
    { ...profile, fullname: "Dũng Một" },
    "",
    "r1",
    at,
  )
  const second = requestProfileChange(
    first,
    "dung",
    profile,
    { ...profile, fullname: "Dũng Hai" },
    "",
    "r2",
    at,
  )
  const third = requestProfileChange(
    second,
    "dung",
    profile,
    { ...profile, phone: "0123456789" },
    "",
    "r3",
    at,
  )
  const approved = resolveProfileRequest(third, "hr", "r1", "APPROVED", "", at)
  assert.throws(
    () => resolveProfileRequest(approved, "hr", "r2", "APPROVED", "", at),
    /nền đã thay đổi/,
  )
  const rejected = resolveProfileRequest(
    approved,
    "hr",
    "r2",
    "REJECTED",
    "Tên đã đổi theo yêu cầu trước",
    at,
  )
  const phoneApproved = resolveProfileRequest(rejected, "hr", "r3", "APPROVED", "", at)
  assert.equal(
    phoneApproved.profiles.find((entry) => entry.personId === "dung")?.fullname,
    "Dũng Một",
  )
  assert.equal(
    phoneApproved.profiles.find((entry) => entry.personId === "dung")?.phone,
    "0123456789",
  )
  assert.throws(
    () =>
      requestProfileChange(
        phoneApproved,
        "dung",
        profile,
        { ...profile, fullname: "Dũng Ba" },
        "",
        "r4",
        at,
      ),
    /chính thức đã thay đổi/,
  )
})

test("địa chỉ thiếu không ép patch họ tên/SĐT; địa chỉ đã có không xóa trắng, không nhận đổi email/cơ cấu", () => {
  const { state, profile } = setup()
  assert.deepEqual(profilePatch(profile, { ...profile, phone: "0901234567", address: "" }), {
    phone: "0901234567",
  })
  assert.throws(
    () => requestProfileChange(state, "dung", profile, { ...profile, fullname: " " }, "", "r1", at),
    /họ và tên/,
  )
  assert.throws(
    () =>
      requestProfileChange(
        state,
        "dung",
        profile,
        { ...profile, phone: "1".repeat(21) },
        "",
        "r1",
        at,
      ),
    /20 ký tự/,
  )
  const addressed = { ...profile, address: "Địa chỉ chính thức" }
  const withAddress = {
    ...state,
    profiles: state.profiles.map((entry) => (entry.personId === "dung" ? addressed : entry)),
  }
  assert.throws(
    () =>
      requestProfileChange(
        withAddress,
        "dung",
        addressed,
        { ...addressed, address: "  " },
        "",
        "r1",
        at,
      ),
    /Địa chỉ/,
  )
  const expandedProposal = {
    ...profile,
    phone: "0901234567",
    companyEmail: "khac@example.com",
    department: "Khác",
    departmentId: "Khác",
    positionId: "Khác",
  }
  const pending = requestProfileChange(state, "dung", profile, expandedProposal, "", "r1", at)
  assert.deepEqual(Object.keys(pending.requests[0].proposed), ["phone"])
  assert.throws(
    () =>
      requestProfileChange(
        state,
        "unknown",
        profile,
        { ...profile, phone: "0901234567" },
        "",
        "r1",
        at,
      ),
    /không có quyền/,
  )
})

test("tên chính thức đồng bộ projection/tra cứu/gợi ý/demo identity, không lộ dữ liệu hành chính", () => {
  const { state, profile } = setup()
  const pending = requestProfileChange(
    state,
    "dung",
    profile,
    { ...profile, fullname: "Phạm Dũng Mới", phone: "0901234567", address: "Riêng tư" },
    "",
    "r1",
    at,
  )
  const approved = resolveProfileRequest(pending, "hr", "r1", "APPROVED", "", at)
  const workspace = createDemoWorkspaceState()
  const directory = professionalPeople(approved.profiles, createOrganizationState())
  const person = directory.find((entry) => entry.id === "dung")!
  assert.equal(person.name, "Phạm Dũng Mới")
  for (const field of [
    "phone",
    "address",
    "companyEmail",
    "department",
    "departmentId",
    "positionId",
  ])
    assert.equal(field in person, false)
  const lookup = personnelProfiles(
    workspace.projects,
    workspace.tasks,
    "khoa",
    "Dũng Mới",
    directory,
  )
  assert.equal(lookup[0].id, "dung")
  assert.equal(JSON.stringify(lookup).includes("Riêng tư"), false)
  const ranked = recommendations(workspace.projects[0], workspace.tasks[0], directory)
  assert.equal(ranked.find((entry) => entry.person.id === "dung")?.person.name, "Phạm Dũng Mới")
  assert.equal(
    demoIdentityOptions(workspace.projects, undefined, directory).find(
      (entry) => entry.id === "dung",
    )?.name,
    "Phạm Dũng Mới",
  )
})

test("action cập nhật hồ sơ/quyết định cùng bước, lỗi không dispatch và reset khôi phục mọi nguồn", () => {
  let state = createDemoWorkspaceState()
  let count = 0
  const dispatch = (action: DemoWorkspaceAction) => {
    count++
    state = demoWorkspaceReducer(state, action)
  }
  const before = state.profileState.profiles.find((entry) => entry.personId === "dung")!
  createDemoWorkspaceActions(state, dispatch, "dung").requestProfileChange(
    before,
    { ...before, phone: "0901234567" },
    "",
  )
  const id = state.profileState.requests[0].id
  assert.equal(state.profileState.profiles.find((entry) => entry.personId === "dung")?.phone, null)
  assert.throws(() =>
    createDemoWorkspaceActions(state, dispatch, "ha").resolveProfileRequest(id, "APPROVED", ""),
  )
  assert.equal(count, 1)
  createDemoWorkspaceActions(state, dispatch, "hr").resolveProfileRequest(id, "APPROVED", "")
  assert.equal(count, 2)
  assert.equal(state.profileState.requests[0].status, "APPROVED")
  assert.equal(
    state.profileState.profiles.find((entry) => entry.personId === "dung")?.phone,
    "0901234567",
  )
  createDemoWorkspaceActions(state, dispatch, "hr").reset()
  assert.equal(state.profileState.requests.length, 0)
  assert.equal(state.profileState.profiles.find((entry) => entry.personId === "dung")?.phone, null)
})
