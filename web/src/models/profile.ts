// Projection hồ sơ và yêu cầu trong phiên frontend; không phải schema/DTO server.
export type EmployeeProfile = {
  personId: string
  employeeCode: string
  fullname: string
  companyEmail: string
  phone: string | null
  address: string | null
  departmentId: string | null
  positionId: string
}

export const profileFields = ["fullname", "phone", "address"] as const
export type ProfileField = (typeof profileFields)[number]
export type ProfileValues = Pick<EmployeeProfile, ProfileField>
export type ProfilePatch = Partial<ProfileValues>
export const profileFieldLabels: Record<ProfileField, string> = {
  fullname: "Họ và tên",
  phone: "Số điện thoại",
  address: "Địa chỉ",
}

export type ProfileChangeRequest = {
  id: string
  employeeId: string
  before: ProfilePatch
  proposed: ProfilePatch
  explanation: string
  requestedBy: string
  requestedAt: string
  status: "PENDING" | "APPROVED" | "REJECTED"
  resolvedBy?: string
  resolvedAt?: string
  decisionReason?: string
}

export type ProfileState = {
  profiles: EmployeeProfile[]
  requests: ProfileChangeRequest[]
}
