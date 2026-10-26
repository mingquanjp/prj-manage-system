import type {
  EmployeeProfile,
  ProfileChangeRequest,
  ProfilePatch,
  ProfileState,
  ProfileValues,
} from "../models/profile.ts"
import { profileFields, profileFieldLabels } from "../models/profile.ts"
import { people, demoUsers } from "./fixtures.ts"
import { getPerson } from "./identity.ts"

export function createProfileState(): ProfileState {
  return {
    profiles: people.map((person, index) => ({
      personId: person.id,
      employeeCode: `E${String(index + 1).padStart(3, "0")}`,
      fullname: person.name,
      companyEmail: `${person.id}@congty.example`,
      phone: null,
      address: null,
      department:
        person.id === "hr" ? "Nhân sự" : person.id === "ha" ? null : "Công nghệ thông tin",
    })),
    requests: [],
  }
}

export const canReviewProfiles = (actorId: string) =>
  Boolean(getPerson(actorId)?.active) &&
  demoUsers.some((user) => user.personId === actorId && user.systemRoles.includes("HR"))

// Chỉ tên/viết tắt được ghép vào Person; không đưa liên hệ/địa chỉ sang miền chuyên môn.
export function professionalPeople(profiles: EmployeeProfile[]) {
  return people.map((person) => {
    const profile = profiles.find((entry) => entry.personId === person.id)
    if (!profile) return person
    return {
      ...person,
      name: profile.fullname,
      initials: profile.fullname
        .split(/\s+/)
        .slice(-2)
        .map((word) => word[0])
        .join("")
        .toLocaleUpperCase("vi"),
    }
  })
}

export function profilePatch(before: ProfileValues, proposed: ProfileValues): ProfilePatch {
  const patch: ProfilePatch = {}
  for (const field of profileFields) {
    const value = proposed[field]
    const next = typeof value === "string" ? value.trim() : value
    const normalized =
      next === "" && (field === "phone" || (field === "address" && before.address === null))
        ? null
        : next
    if (before[field] !== normalized) Object.assign(patch, { [field]: normalized })
  }
  return patch
}

export function profilePatchErrors(patch: ProfilePatch) {
  const errors: Partial<Record<keyof ProfileValues, string>> = {}
  if (
    "fullname" in patch &&
    (typeof patch.fullname !== "string" || !patch.fullname.trim() || patch.fullname.length > 100)
  )
    errors.fullname = "Nhập họ và tên từ 1 đến 100 ký tự."
  if (
    "phone" in patch &&
    patch.phone !== null &&
    (typeof patch.phone !== "string" || patch.phone.length > 20)
  )
    errors.phone = "Số điện thoại tối đa 20 ký tự."
  if ("address" in patch && (!patch.address || !patch.address.trim()))
    errors.address = "Địa chỉ bổ sung hoặc chỉnh sửa không được để trống."
  return errors
}

export function profileConflict(profile: EmployeeProfile, request: ProfileChangeRequest) {
  return profileFields.some(
    (field) => field in request.proposed && profile[field] !== request.before[field],
  )
}

export function requestProfileChange(
  state: ProfileState,
  actorId: string,
  before: ProfileValues,
  proposed: ProfileValues,
  explanation: string,
  id: string,
  at: string,
): ProfileState {
  const profile = state.profiles.find((entry) => entry.personId === actorId)
  if (!profile || !getPerson(actorId)?.active)
    throw new Error("Bạn không có quyền gửi yêu cầu cập nhật hồ sơ này.")
  const patch = profilePatch(before, proposed)
  if (!Object.keys(patch).length) throw new Error("Chưa có thông tin thay đổi.")
  const errors = profilePatchErrors(patch)
  if (Object.keys(errors).length) throw new Error(Object.values(errors).join(" "))
  const captured: ProfilePatch = {}
  for (const field of profileFields) {
    if (!(field in patch)) continue
    if (before[field] !== profile[field])
      throw new Error(
        "Thông tin chính thức đã thay đổi. Đóng biểu mẫu và kiểm tra lại trước khi gửi.",
      )
    Object.assign(captured, { [field]: profile[field] })
  }
  return {
    ...state,
    requests: [
      ...state.requests,
      {
        id,
        employeeId: actorId,
        before: captured,
        proposed: patch,
        explanation: explanation.trim(),
        requestedBy: actorId,
        requestedAt: at,
        status: "PENDING",
      },
    ],
  }
}

export function resolveProfileRequest(
  state: ProfileState,
  actorId: string,
  requestId: string,
  decision: "APPROVED" | "REJECTED",
  reason: string,
  at: string,
): ProfileState {
  if (!canReviewProfiles(actorId)) throw new Error("Chỉ HR được xử lý yêu cầu thông tin cá nhân.")
  const request = state.requests.find((entry) => entry.id === requestId)
  const profile = state.profiles.find((entry) => entry.personId === request?.employeeId)
  if (!request || !profile) throw new Error("Không tìm thấy yêu cầu hoặc hồ sơ.")
  if (request.status !== "PENDING") throw new Error("Yêu cầu đã được xử lý. Vui lòng kiểm tra lại.")
  if (!getPerson(request.employeeId)?.active)
    throw new Error("Hồ sơ nhân viên không còn hoạt động.")
  if (decision === "REJECTED" && !reason.trim()) throw new Error("Vui lòng nhập lý do từ chối.")
  if (decision === "APPROVED") {
    if (profileConflict(profile, request))
      throw new Error(
        "Thông tin nền đã thay đổi. Không thể chấp thuận và ghi đè dữ liệu mới; bạn có thể từ chối để nhân viên gửi yêu cầu mới.",
      )
    const errors = profilePatchErrors(request.proposed)
    if (Object.keys(errors).length) throw new Error(Object.values(errors).join(" "))
  }
  const updated = { ...profile }
  if (decision === "APPROVED") {
    for (const field of profileFields) {
      if (field in request.proposed) updated[field] = request.proposed[field]!
    }
  }
  return {
    profiles: state.profiles.map((entry) =>
      entry.personId === profile.personId ? updated : entry,
    ),
    requests: state.requests.map((entry) =>
      entry.id === request.id
        ? {
            ...entry,
            status: decision,
            resolvedBy: actorId,
            resolvedAt: at,
            decisionReason: reason.trim(),
          }
        : entry,
    ),
  }
}

export const profileRequestLabels = {
  PENDING: "Chờ xử lý",
  APPROVED: "Đã chấp thuận",
  REJECTED: "Đã từ chối",
}

export function describeProfileChanges(request: ProfileChangeRequest) {
  return profileFields
    .filter((field) => field in request.proposed)
    .map((field) => profileFieldLabels[field])
    .join(", ")
}
