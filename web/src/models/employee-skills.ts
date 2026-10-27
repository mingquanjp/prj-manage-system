// Read model của phiên mô phỏng; không phải schema hoặc DTO server.
export type SkillValues = { level: number; years: number | null; note: string }
export type OfficialSkill = SkillValues & {
  id: string
  employeeId: string
  skillId: string
  approvedBy?: string
  approvedAt?: string
}
export type SkillEvidence = {
  assignmentId: string
  taskId: string
  taskCode: string
  taskName: string
  description: string
  criteria: string
  requirements: { skill: string; level: number; mandatory: boolean }[]
  contributionNote: string
  submission?: {
    id: string
    summary: string
    links: string[]
    submittedAt: string
    review?: { result: "ACCEPTED" | "REJECTED"; reviewerId: string; at: string; comment: string }
  }
}
export type SkillRequestInput = {
  operation: "ADD" | "UPDATE" | "REMOVE"
  skillId: string
  before: SkillValues | null
  proposed: SkillValues | null
  explanation: string
  externalEvidence: string
  evidence: { assignmentId: string; contributionNote: string }[]
}
export type SkillRequest = Omit<SkillRequestInput, "evidence"> & {
  id: string
  employeeId: string
  source: "MANUAL"
  evidence: SkillEvidence[]
  version: number
  requestedAt: string
  status: "PENDING" | "APPROVED" | "REJECTED" | "WITHDRAWN"
  resolvedBy?: string
  resolvedAt?: string
  decisionReason?: string
  expertOpinion?: string
  supplements: { at: string; version: number }[]
}
export type EmployeeSkillsState = { official: OfficialSkill[]; requests: SkillRequest[] }
export type SkillSupplement = Pick<
  SkillRequestInput,
  "explanation" | "externalEvidence" | "evidence"
>
export type SkillDecision = {
  requestId: string
  version: number
  decision: "APPROVED" | "REJECTED"
  reason: string
  expertOpinion: string
}
export const skillOperationLabels = {
  ADD: "Thêm kỹ năng",
  UPDATE: "Điều chỉnh",
  REMOVE: "Gỡ kỹ năng",
}
export const skillRequestLabels = {
  PENDING: "Chờ xét duyệt",
  APPROVED: "Đã chấp thuận",
  REJECTED: "Đã từ chối",
  WITHDRAWN: "Đã rút",
}
