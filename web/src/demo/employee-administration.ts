import { defaultIdentity, type IdentitySource } from "./identity.ts"
import type { EmployeeEditValues, EmployeeProfile, ProfileState } from "../models/profile.ts"
import { employeeEditFields } from "../models/profile.ts"
import type { OrganizationState } from "../models/organization.ts"
import { isHr } from "./identity.ts"
import { profilePatch, profilePatchErrors } from "./profiles.ts"

export function employeeEditValues(profile: EmployeeProfile): EmployeeEditValues {
  return Object.fromEntries(
    employeeEditFields.map((field) => [field, profile[field]]),
  ) as EmployeeEditValues
}

export function normalizeEmployeeEdit(values: EmployeeEditValues): EmployeeEditValues {
  return {
    fullname: values.fullname.trim(),
    companyEmail: values.companyEmail.trim(),
    phone: values.phone?.trim() || null,
    address: values.address?.trim() || null,
    departmentId: values.departmentId || null,
    positionId: values.positionId,
    jobLevelId: values.jobLevelId || null,
  }
}

export function employeeEditErrors(
  profiles: EmployeeProfile[],
  organization: OrganizationState,
  before: EmployeeProfile,
  proposed: EmployeeEditValues,
) {
  const values = normalizeEmployeeEdit(proposed)
  const errors: Partial<Record<keyof EmployeeEditValues, string>> = profilePatchErrors(
    profilePatch(before, values),
  )
  if (
    !values.companyEmail ||
    values.companyEmail.length > 100 ||
    !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(values.companyEmail)
  )
    errors.companyEmail = "Nhập email công ty hợp lệ, tối đa 100 ký tự."
  else if (
    profiles.some(
      (profile) =>
        profile.personId !== before.personId &&
        profile.companyEmail.toLocaleLowerCase() === values.companyEmail.toLocaleLowerCase(),
    )
  )
    errors.companyEmail = "Email công ty đã được sử dụng."
  if (!organization.positions.some((entry) => entry.id === values.positionId))
    errors.positionId = "Chọn vị trí công việc trong danh mục hiện hành."
  if (
    values.departmentId &&
    !organization.departments.some((entry) => entry.id === values.departmentId)
  )
    errors.departmentId = "Phòng ban không còn trong danh mục."
  if (values.jobLevelId && !organization.jobLevels.some((entry) => entry.id === values.jobLevelId))
    errors.jobLevelId = "Cấp bậc không còn trong danh mục."
  return errors
}

export function updateEmployeeProfile(
  state: ProfileState,
  organization: OrganizationState,
  actorId: string,
  before: EmployeeProfile,
  proposed: EmployeeEditValues,
  id: string,
  at: string,
  identity: IdentitySource = defaultIdentity,
): ProfileState {
  if (
    !isHr(actorId, identity) ||
    state.profiles.find((p) => p.personId === actorId)?.employmentStatus !== "ACTIVE"
  )
    throw new Error("Chỉ HR được cập nhật hồ sơ nhân sự.")
  const current = state.profiles.find((entry) => entry.personId === before.personId)
  if (!current) throw new Error("Không tìm thấy hồ sơ nhân viên.")
  const values = normalizeEmployeeEdit(proposed)
  const fields = employeeEditFields.filter((field) => values[field] !== before[field])
  if (!fields.length) throw new Error("Chưa có thông tin thay đổi.")
  if (fields.some((field) => current[field] !== before[field]))
    throw new Error("Hồ sơ đã thay đổi. Đóng biểu mẫu và kiểm tra lại trước khi lưu.")
  const errors = employeeEditErrors(state.profiles, organization, current, values)
  if (Object.keys(errors).length) throw new Error(Object.values(errors).join(" "))
  const capturedBefore = Object.fromEntries(fields.map((field) => [field, current[field]]))
  const after = Object.fromEntries(fields.map((field) => [field, values[field]]))
  // Giữ tên danh mục tại lúc sửa để lịch sử không đổi khi mục cũ được đổi tên/xóa.
  const displaySnapshot = (source: EmployeeEditValues) =>
    Object.fromEntries(
      fields.map((field) => {
        const value = source[field]
        const kind =
          field === "departmentId"
            ? "departments"
            : field === "positionId"
              ? "positions"
              : field === "jobLevelId"
                ? "jobLevels"
                : null
        return [
          field,
          kind ? (organization[kind].find((entry) => entry.id === value)?.name ?? null) : value,
        ]
      }),
    )
  return {
    ...state,
    profiles: state.profiles.map((entry) =>
      entry.personId === current.personId ? { ...entry, ...after } : entry,
    ),
    edits: [
      ...state.edits,
      {
        id,
        employeeId: current.personId,
        actorId,
        at,
        before: capturedBefore,
        after,
        beforeDisplay: displaySnapshot(current),
        afterDisplay: displaySnapshot(values),
      },
    ],
  }
}
