import { defaultIdentity, type IdentitySource } from "./identity.ts"
import type { Project, Task, TaskStatus } from "../models/workspace.ts"
import type { DemoWorkspaceState } from "./workspace-state.ts"
import { contributionFor } from "./corrections.ts"
import { rememberContributions, type WorkState } from "./effort.ts"
import { getPerson } from "./identity.ts"

export type LifecycleAction = "complete" | "close" | "cancelProject" | "cancelTask"
export type LifecycleCommand = {
  projectId: string
  expectedStatus: Project["status"]
  action: LifecycleAction
  taskId?: string
  expectedTaskStatus?: TaskStatus
  reason?: string
}
export const lifecycleLabels: Record<LifecycleAction, string> = {
  complete: "Hoàn thành dự án",
  close: "Đóng dự án",
  cancelProject: "Hủy dự án",
  cancelTask: "Hủy công việc",
}
export const isOpenTask = (task: Task) => !["DONE", "CLOSED", "CANCELLED"].includes(task.status)

export function pendingProjectCorrections(project: Project, tasks: Task[], work: WorkState) {
  return work.requests.filter(
    (request) =>
      request.status === "PENDING" &&
      contributionFor(work, tasks, request.assignmentId)?.projectId === project.id,
  )
}

export function lifecycleBlockReason(
  project: Project,
  tasks: Task[],
  work: WorkState,
  actorId: string,
  action: LifecycleAction,
  task?: Task,
  identity: IdentitySource = defaultIdentity,
): string | undefined {
  if (!getPerson(actorId, identity)?.active || project.pmId !== actorId)
    return "Chỉ PM hiện hành của dự án được thực hiện thao tác này."
  const related = tasks.filter((entry) => entry.projectId === project.id)
  switch (action) {
    case "complete":
      if (project.status !== "IN_PROGRESS") return "Chỉ hoàn thành dự án đang thực hiện."
      if (!related.some((entry) => entry.status === "DONE"))
        return "Cần ít nhất một công việc đã nghiệm thu."
      if (related.some((entry) => !["DONE", "CANCELLED"].includes(entry.status)))
        return "Còn công việc chưa nghiệm thu hoặc chưa hủy."
      return undefined
    case "close":
      if (project.status !== "COMPLETED") return "Chỉ đóng dự án đã hoàn thành."
      if (pendingProjectCorrections(project, tasks, work).length)
        return "Còn yêu cầu hiệu chỉnh nhật ký chờ PTCM xử lý; chưa thể đóng dự án."
      return undefined
    case "cancelProject":
    case "cancelTask":
      if (!["PLANNING", "IN_PROGRESS"].includes(project.status))
        return "Chỉ hủy khi dự án đang lập kế hoạch hoặc thực hiện."
      if (action === "cancelTask" && (!task || task.projectId !== project.id || !isOpenTask(task)))
        return "Chỉ hủy công việc mở thuộc dự án này; không hủy công việc đã nghiệm thu/đóng/hủy."
      return undefined
    default:
      return "Thao tác không hợp lệ."
  }
}

function cancelTask(task: Task, actorId: string, at: string, reason: string): Task {
  return {
    ...task,
    status: "CANCELLED",
    currentAssignment:
      task.currentAssignment && !task.currentAssignment.endedAt
        ? { ...task.currentAssignment, endedAt: at, endedBy: actorId, endReason: reason }
        : task.currentAssignment,
    executionHistory: [
      ...(task.executionHistory ?? []),
      { actorId, at, from: task.status, to: "CANCELLED", reason },
    ],
  }
}

// Chuẩn bị cả dự án/task/effort trước một dispatch; không phải transaction server.
export function transitionProjectLifecycle(
  state: DemoWorkspaceState,
  actorId: string,
  command: LifecycleCommand,
  at: string,
  identity: IdentitySource = defaultIdentity,
) {
  const project = state.projects.find((entry) => entry.id === command.projectId)
  if (!project) throw new Error("Không tìm thấy dự án.")
  const task = state.tasks.find((entry) => entry.id === command.taskId)
  const block = lifecycleBlockReason(
    project,
    state.tasks,
    state.work,
    actorId,
    command.action,
    task,
    identity,
  )
  if (
    project.status !== command.expectedStatus ||
    (command.action === "cancelTask" && task?.status !== command.expectedTaskStatus)
  )
    throw new Error("Trạng thái đã thay đổi. Kiểm tra lại trước khi xác nhận.")
  if (block) throw new Error(block)
  const cancelling = command.action === "cancelProject" || command.action === "cancelTask"
  const reason = command.reason?.trim() ?? ""
  if (cancelling && !reason) throw new Error("Nhập lý do hủy.")
  const status =
    command.action === "complete"
      ? "COMPLETED"
      : command.action === "close"
        ? "CLOSED"
        : command.action === "cancelProject"
          ? "CANCELLED"
          : project.status
  const nextProject: Project = { ...project, status }
  const tasks = state.tasks.map((entry): Task => {
    if (entry.projectId !== project.id) return entry
    if (command.action === "close" && entry.status === "DONE")
      return {
        ...entry,
        status: "CLOSED",
        executionHistory: [
          ...(entry.executionHistory ?? []),
          { actorId, at, from: "DONE", to: "CLOSED" },
        ],
      }
    if (
      isOpenTask(entry) &&
      (command.action === "cancelProject" ||
        (command.action === "cancelTask" && entry.id === task?.id))
    )
      return cancelTask(entry, actorId, at, reason)
    return entry
  })
  let work = rememberContributions(state.work, tasks)
  if (command.action === "cancelProject") {
    const pending = pendingProjectCorrections(project, state.tasks, state.work)
    const ids = new Set(pending.map((entry) => entry.id))
    work = {
      ...work,
      requests: work.requests.map((request) =>
        ids.has(request.id)
          ? {
              ...request,
              status: "REJECTED",
              resolvedBy: actorId,
              resolvedAt: at,
              decisionReason: `PROJECT_CANCELLED · ${reason}`,
            }
          : request,
      ),
      notifications: [
        ...work.notifications,
        ...pending.map((request) => ({
          id: `project-cancel:${request.id}`,
          recipientId: request.requestedBy,
          requestId: request.id,
          at,
          message:
            "Yêu cầu hiệu chỉnh đã kết thúc do dự án bị hủy; nhật ký chính thức được giữ nguyên.",
        })),
      ],
    }
  }
  const label = lifecycleLabels[command.action]
  return {
    project: nextProject,
    tasks,
    work,
    event: {
      projectId: project.id,
      actorId,
      at,
      action: task && command.action === "cancelTask" ? `${label} · ${task.code}` : label,
      before: structuredClone(project),
      after: structuredClone(nextProject),
      ...(reason ? { reason } : {}),
    },
    notice: `Đã ${label.toLocaleLowerCase("vi-VN")} trong phiên mô phỏng.`,
  }
}
