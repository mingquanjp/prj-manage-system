// Kiểu dữ liệu frontend hiện hành; không phải DTO API hoặc schema database.
export type TaskStatus =
  | "PENDING_ASSIGNMENT"
  | "ASSIGNED"
  | "IN_PROGRESS"
  | "PAUSED"
  | "PENDING_REVIEW"
  | "DONE"
  | "CLOSED"
  | "CANCELLED"

export type Project = {
  id: string
  code: string
  name: string
  description: string
  status: "PLANNING" | "IN_PROGRESS" | "COMPLETED" | "CLOSED" | "CANCELLED"
  pmId: string
  ptcmId?: string
  start: string
  end: string
  priority: "Thấp" | "Trung bình" | "Cao" | "Khẩn cấp"
  // Projection từ lịch sử triển khai cho demo; không phải planned start hay cột DB mới.
  executionStartedAt?: string
}

export type Requirement = {
  skill: string
  level: number
  mandatory: boolean
  weight: number
}

export type TaskGroup = {
  id: string
  projectId: string
  name: string
  description: string
}

export type Task = {
  id: string
  projectId: string
  code: string
  name: string
  group: string
  groupId: string
  description: string
  criteria: string
  estimate: number
  status: TaskStatus
  requirements: Requirement[]
  assigneeId?: string
  firstAssignedAt?: string
  // Lượt giao hiện tại trong adapter demo, tách mốc giao đầu dùng khóa đặc tả.
  currentAssignment?: {
    id: string
    employeeId: string
    assignedAt: string
    endedAt?: string
    endedBy?: string
    endReason?: string
  }
  acceptedAt?: string
  priority: "Thấp" | "Trung bình" | "Cao" | "Khẩn cấp"
  start?: string
  end?: string
  // Lịch sử trạng thái trong phiên demo, không phải nhật ký thời gian hay audit server.
  executionHistory?: TaskExecutionEvent[]
}

export type ExecutionAction = "start" | "pause" | "resume"

export type TaskExecutionEvent = {
  actorId: string
  at: string
  from: TaskStatus
  to: TaskStatus
  reason?: string
}

export type Person = {
  id: string
  name: string
  initials: string
  position: string
  seniority: string
  active: boolean
  skills: Record<string, number>
  backgroundLoad: number
}

export type SystemRole = "ADMIN" | "HR" | "DIRECTOR"

// Projection minh họa từ lịch sử đóng góp chưa nghiệm thu; không phải cột/bảng DB mới.
export type OutputConflict = {
  projectId: string
  taskId: string
  personId: string
}

export type Choice = {
  personId: string
  mode: "manual" | "recommended"
  rank?: number
  reason: string
  overrideReason: string
  hadRun: boolean
}

export type Draft = Record<string, Choice>

export type TaskInput = Pick<
  Task,
  | "code"
  | "name"
  | "groupId"
  | "description"
  | "criteria"
  | "estimate"
  | "priority"
  | "start"
  | "end"
  | "requirements"
>

export type FormErrors = Record<string, string>

export type ProjectInput = Pick<
  Project,
  "code" | "name" | "pmId" | "start" | "end" | "priority" | "description"
>

export type ProjectEvent = {
  projectId: string
  actorId: string
  at: string
  action: string
  reason?: string
  before?: Project
  after: Project
}
