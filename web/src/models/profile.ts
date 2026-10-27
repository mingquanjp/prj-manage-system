// Projection hồ sơ và yêu cầu trong phiên frontend; không phải schema/DTO server.
export type EmployeeProfile = {
  personId: string
  employmentStatus: "ACTIVE" | "INACTIVE"
  employeeCode: string
  fullname: string
  companyEmail: string
  phone: string | null
  address: string | null
  departmentId: string | null
  positionId: string
  jobLevelId: string | null
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

export const employeeEditFields = [
  "fullname",
  "companyEmail",
  "phone",
  "address",
  "departmentId",
  "positionId",
  "jobLevelId",
] as const
export type EmployeeEditField = (typeof employeeEditFields)[number]
export type EmployeeEditValues = Pick<EmployeeProfile, EmployeeEditField>
export const employeeEditLabels: Record<EmployeeEditField, string> = {
  ...profileFieldLabels,
  companyEmail: "Email công ty",
  departmentId: "Phòng ban",
  positionId: "Vị trí công việc",
  jobLevelId: "Cấp bậc",
}
export type EmployeeProfileEdit = {
  id: string
  employeeId: string
  actorId: string
  at: string
  before: Partial<EmployeeEditValues>
  after: Partial<EmployeeEditValues>
  beforeDisplay: Partial<Record<EmployeeEditField, string | null>>
  afterDisplay: Partial<Record<EmployeeEditField, string | null>>
}

export type EmployeeDeactivation = {
  id: string
  employeeId: string
  actorId: string
  at: string
}

export type ProfileState = {
  profiles: EmployeeProfile[]
  requests: ProfileChangeRequest[]
  edits: EmployeeProfileEdit[]
  deactivations: EmployeeDeactivation[]
}
