import type { Project, Task } from "../models/workspace.ts"
import type {
  PmChangeCommand,
  TaskDeadlineCommand,
  TaskComment,
} from "../models/task-collaboration.ts"
import { getPerson, isDirector, type IdentitySource } from "./identity.ts"

const openProjects = ["PLANNING", "IN_PROGRESS"]
const terminalTasks = ["DONE", "CLOSED", "CANCELLED"]
export function replaceProjectPm(
  project: Project,
  actorId: string,
  command: PmChangeCommand,
  identity: IdentitySource,
): Project {
  if (!isDirector(actorId, identity)) throw new Error("Chỉ Giám đốc được thay PM.")
  if (![...openProjects, "COMPLETED"].includes(project.status))
    throw new Error("Dự án đã đóng/hủy, không thể thay PM.")
  if (command.projectId !== project.id || command.beforePmId !== project.pmId)
    throw new Error("PM đã thay đổi. Mở lại biểu mẫu để kiểm tra.")
  if (!getPerson(command.pmId, identity)?.active)
    throw new Error("Chọn nhân viên có tài khoản và hồ sơ đang hoạt động.")
  if (command.pmId === project.pmId) throw new Error("Hãy chọn PM khác người hiện tại.")
  return { ...project, pmId: command.pmId }
}
export function canExtendTask(
  project: Project,
  task: Task,
  actorId: string,
  identity: IdentitySource,
) {
  return (
    Boolean(getPerson(actorId, identity)?.active) &&
    task.projectId === project.id &&
    project.ptcmId === actorId &&
    openProjects.includes(project.status) &&
    !terminalTasks.includes(task.status)
  )
}
function validDate(value: string) {
  return (
    /^\d{4}-\d{2}-\d{2}$/.test(value) &&
    Number.isFinite(Date.parse(value)) &&
    new Date(value).toISOString().slice(0, 10) === value
  )
}
export function extendTaskDeadline(
  project: Project,
  task: Task,
  actorId: string,
  command: TaskDeadlineCommand,
  at: string,
  identity: IdentitySource,
): Task {
  if (!canExtendTask(project, task, actorId, identity))
    throw new Error("Chỉ PTCM được gia hạn công việc chưa kết thúc trong dự án đang mở.")
  if (
    command.taskId !== task.id ||
    command.beforeStart !== task.start ||
    command.beforeEnd !== task.end
  )
    throw new Error("Kế hoạch đã thay đổi. Mở lại biểu mẫu để kiểm tra.")
  if (!command.reason.trim()) throw new Error("Nhập lý do điều chỉnh hạn.")
  if (
    !validDate(command.start) ||
    !validDate(command.end) ||
    command.start > command.end ||
    command.start < project.start ||
    command.end > project.end
  )
    throw new Error("Nhập đủ cặp ngày hợp lệ trong kế hoạch dự án.")
  if (task.start && command.start !== task.start)
    throw new Error("Gia hạn giữ nguyên ngày bắt đầu đã lưu.")
  if (task.end && command.end <= task.end) throw new Error("Hạn mới phải sau hạn hiện tại.")
  return {
    ...task,
    originalStart: task.originalStart ?? task.start ?? command.start,
    originalEnd: task.originalEnd ?? task.end ?? command.end,
    start: command.start,
    end: command.end,
    planHistory: [
      ...(task.planHistory ?? []),
      {
        actorId,
        at,
        before: { start: task.start, end: task.end },
        after: { start: command.start, end: command.end },
        reason: command.reason.trim(),
      },
    ],
  }
}
export function canReadTaskDiscussion(
  project: Project,
  task: Task,
  actorId: string,
  identity: IdentitySource,
) {
  return (
    task.projectId === project.id &&
    Boolean(getPerson(actorId, identity)?.active) &&
    (project.ptcmId === actorId ||
      (task.assigneeId === actorId &&
        (!task.currentAssignment || task.currentAssignment.employeeId === actorId)))
  )
}
export function canPostTaskDiscussion(
  project: Project,
  task: Task,
  actorId: string,
  identity: IdentitySource,
) {
  return (
    canReadTaskDiscussion(project, task, actorId, identity) &&
    openProjects.includes(project.status) &&
    !terminalTasks.includes(task.status)
  )
}
export function addTaskComment(
  comments: TaskComment[],
  project: Project,
  task: Task,
  actorId: string,
  content: string,
  id: string,
  at: string,
  identity: IdentitySource,
): TaskComment[] {
  if (!canPostTaskDiscussion(project, task, actorId, identity))
    throw new Error("Chỉ người đang được giao và PTCM được trao đổi khi công việc/dự án còn mở.")
  if (!content.trim()) throw new Error("Nhập nội dung trao đổi.")
  if (comments.some((comment) => comment.id === id)) throw new Error("Trao đổi đã được ghi nhận.")
  return [...comments, { id, taskId: task.id, actorId, content: content.trim(), at }]
}
