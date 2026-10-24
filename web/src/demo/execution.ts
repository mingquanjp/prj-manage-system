import type { Project, Task, TaskStatus, ExecutionAction } from "../models/workspace.ts"
import { getPerson } from "./identity.ts"

export function personalTasks(projects: Project[], tasks: Task[], actorId: string) {
  const person = getPerson(actorId)
  if (!person?.active) return []
  return tasks.filter((task) => {
    const project = projects.find((p) => p.id === task.projectId)
    return (
      project &&
      project.ptcmId !== actorId &&
      task.assigneeId === actorId &&
      task.status !== "PENDING_ASSIGNMENT"
    )
  })
}

export function executionBlockReason(project: Project, task: Task, actorId: string) {
  if (
    task.projectId !== project.id ||
    task.assigneeId !== actorId ||
    !getPerson(actorId)?.active ||
    project.ptcmId === actorId
  )
    return "Bạn không có quyền thực hiện công việc này."
  if (project.status === "PLANNING")
    return "Dự án chưa bắt đầu triển khai. Bạn có thể xem công việc nhưng chưa thể thực hiện."
  if (project.status !== "IN_PROGRESS")
    return "Dự án đã kết thúc giai đoạn thực hiện. Công việc chỉ được xem."
  if (!["ASSIGNED", "IN_PROGRESS", "PAUSED"].includes(task.status))
    return "Trạng thái hiện tại không cho phép bắt đầu, tạm dừng hoặc tiếp tục."
  return undefined
}

export function transitionPersonalTask(
  projects: Project[],
  tasks: Task[],
  actorId: string,
  taskId: string,
  expectedStatus: TaskStatus,
  action: ExecutionAction,
  reason = "",
): Task[] {
  const task = tasks.find((t) => t.id === taskId)
  const project = projects.find((p) => p.id === task?.projectId)
  if (!task || !project) throw new Error("Không tìm thấy công việc.")
  const block = executionBlockReason(project, task, actorId)
  if (block) throw new Error(block)
  if (task.status !== expectedStatus)
    throw new Error("Trạng thái công việc đã thay đổi. Vui lòng kiểm tra lại.")
  const from: Record<ExecutionAction, TaskStatus> = {
    start: "ASSIGNED",
    pause: "IN_PROGRESS",
    resume: "PAUSED",
  }
  if (task.status !== from[action])
    throw new Error("Thao tác không phù hợp với trạng thái công việc.")
  if (action === "pause" && !reason.trim()) throw new Error("Vui lòng nhập lý do tạm dừng.")
  const to = action === "pause" ? "PAUSED" : "IN_PROGRESS"
  return tasks.map((t) =>
    t.id === taskId
      ? {
          ...t,
          status: to,
          executionHistory: [
            ...(t.executionHistory ?? []),
            {
              actorId,
              at: new Date().toISOString(),
              from: task.status,
              to,
              ...(action === "pause" ? { reason: reason.trim() } : {}),
            },
          ],
        }
      : t,
  )
}
