// Dữ liệu trong phiên demo; không phải DTO hoặc lịch sử bền vững của server.
export type PmChangeCommand = { projectId: string; beforePmId: string; pmId: string }
export type TaskDeadlineCommand = {
  taskId: string
  beforeStart?: string
  beforeEnd?: string
  start: string
  end: string
  reason: string
}
export type TaskPlanEvent = {
  actorId: string
  at: string
  before: { start?: string; end?: string }
  after: { start: string; end: string }
  reason: string
}
export type TaskComment = {
  id: string
  taskId: string
  actorId: string
  at: string
  content: string
}
