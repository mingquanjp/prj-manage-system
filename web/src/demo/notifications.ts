import type { WorkspaceNotification } from "../models/notification.ts"
import type { DemoWorkspaceState } from "./workspace-state.ts"
import { getPerson, isDirector, isHr, type IdentitySource } from "./identity.ts"
import { canReadRequest } from "./corrections.ts"
import { personalTasks } from "./execution.ts"

// Chỉ suy sự kiện từ kết quả thao tác thành công. ID/thời điểm lấy từ dữ liệu miền,
// không phát sinh thời gian hoặc ID trong reducer, không dựng thông báo từ fixture ban đầu.
export function collectNotifications(
  before: DemoWorkspaceState,
  after: DemoWorkspaceState,
): WorkspaceNotification[] {
  const added: WorkspaceNotification[] = []
  const emit = (
    id: string,
    recipientId: string | undefined,
    message: string,
    at: string | undefined,
    target: WorkspaceNotification["target"],
  ) => {
    if (recipientId && at)
      added.push({ id: `${id}:${recipientId}`, recipientId, message, at, target })
  }
  for (const task of after.tasks) {
    const previous = before.tasks.find((t) => t.id === task.id)
    const current = task.currentAssignment
    if (previous && current && current.id !== previous.currentAssignment?.id && task.assigneeId) {
      emit(
        `assigned:${current.id}`,
        task.assigneeId,
        "Bạn được giao một công việc. Mở chi tiết để xem nội dung.",
        current.assignedAt,
        { kind: "task", id: task.id },
      )
      if (previous.assigneeId && previous.assigneeId !== task.assigneeId)
        emit(
          `replaced:${current.id}`,
          previous.assigneeId,
          "Một công việc đã được chuyển sang người thực hiện khác.",
          current.assignedAt,
          { kind: "task", id: task.id },
        )
    }
    if (previous?.assigneeId && !task.assigneeId && previous.currentAssignment) {
      const contribution = after.work.contributions.find(
        (c) => c.assignment.id === previous.currentAssignment?.id,
      )
      emit(
        `unassigned:${previous.currentAssignment.id}`,
        previous.assigneeId,
        "Một phân công công việc của bạn đã được thu hồi.",
        contribution?.assignment.endedAt,
        { kind: "task", id: task.id },
      )
    }
  }
  for (const submission of after.work.submissions) {
    const old = before.work.submissions.find((s) => s.id === submission.id)
    const task = after.tasks.find((t) => t.id === submission.taskId)
    const project = after.projects.find((p) => p.id === task?.projectId)
    if (!old)
      emit(
        `submitted:${submission.id}`,
        project?.ptcmId,
        "Có kết quả công việc mới chờ nghiệm thu.",
        submission.submittedAt,
        { kind: "task", id: submission.taskId },
      )
    if (!old?.review && submission.review)
      emit(
        `reviewed:${submission.id}`,
        submission.employeeId,
        submission.review.result === "ACCEPTED"
          ? "Kết quả công việc của bạn đã được nghiệm thu."
          : "Kết quả công việc cần chỉnh sửa. Mở để xem phản hồi.",
        submission.review.at,
        { kind: "task", id: submission.taskId },
      )
  }
  const recipients = (role: "HR" | "DIRECTOR", employeeId: string) =>
    after.accounts.accounts
      .filter(
        (account) =>
          account.status === "ACTIVE" &&
          account.systemRoles.includes(role) &&
          account.personId !== employeeId &&
          after.profileState.profiles.some(
            (p) => p.personId === account.personId && p.employmentStatus === "ACTIVE",
          ),
      )
      .map((a) => a.personId)
  for (const request of after.profileState.requests) {
    const old = before.profileState.requests.find((r) => r.id === request.id)
    const target = { kind: "profile-request" as const, id: request.id }
    if (!old)
      for (const recipient of recipients("HR", request.employeeId))
        emit(
          `profile-sent:${request.id}`,
          recipient,
          "Có đề nghị cập nhật thông tin cá nhân cần xử lý.",
          request.requestedAt,
          target,
        )
    if (old?.status === "PENDING" && request.status !== "PENDING")
      emit(
        `profile-result:${request.id}`,
        request.requestedBy,
        request.status === "APPROVED"
          ? "Đề nghị cập nhật thông tin của bạn đã được chấp thuận."
          : "Đề nghị cập nhật thông tin của bạn đã bị từ chối.",
        request.resolvedAt,
        target,
      )
  }
  for (const request of after.employeeSkills.requests) {
    const old = before.employeeSkills.requests.find((r) => r.id === request.id)
    const target = { kind: "skill-request" as const, id: request.id }
    if (!old)
      for (const recipient of recipients("DIRECTOR", request.employeeId))
        emit(
          `skill-sent:${request.id}`,
          recipient,
          "Có đề nghị kỹ năng mới cần xét duyệt.",
          request.requestedAt,
          target,
        )
    if (old?.status === "PENDING" && ["APPROVED", "REJECTED"].includes(request.status))
      emit(
        `skill-result:${request.id}`,
        request.employeeId,
        request.status === "APPROVED"
          ? "Đề nghị kỹ năng của bạn đã được chấp thuận."
          : "Đề nghị kỹ năng của bạn đã bị từ chối.",
        request.resolvedAt,
        target,
      )
  }
  const known = new Set(before.notifications.map((n) => n.id))
  return added.filter((n) => {
    if (known.has(n.id)) return false
    known.add(n.id)
    return true
  })
}

// Hiệu chỉnh giữ bản ghi gốc trong work; adapter này chỉ trình bày cùng một danh sách.
export function notificationsFor(
  state: DemoWorkspaceState,
  actorId: string,
  identity: IdentitySource,
): WorkspaceNotification[] {
  if (!getPerson(actorId, identity)?.active) return []
  return [
    ...state.notifications,
    ...state.work.notifications.map((n): WorkspaceNotification => ({
      ...n,
      id: `effort:${n.id}`,
      target: { kind: "correction", id: n.requestId },
    })),
  ]
    .filter((n) => n.recipientId === actorId)
    .sort((a, b) => b.at.localeCompare(a.at) || b.id.localeCompare(a.id))
}
export function markNotificationRead(
  state: DemoWorkspaceState,
  actorId: string,
  id: string,
  at: string,
  identity: IdentitySource,
): Pick<DemoWorkspaceState, "notifications" | "work"> {
  const entry = notificationsFor(state, actorId, identity).find((n) => n.id === id)
  if (!entry) throw new Error("Không thể cập nhật thông báo này.")
  if (!Number.isFinite(Date.parse(at)) || Date.parse(at) < Date.parse(entry.at))
    throw new Error("Thời điểm đọc thông báo không hợp lệ.")
  return {
    notifications: state.notifications.map((n) =>
      n.id === id && n.recipientId === actorId ? { ...n, readAt: n.readAt ?? at } : n,
    ),
    work: {
      ...state.work,
      notifications: state.work.notifications.map((n) =>
        `effort:${n.id}` === id && n.recipientId === actorId ? { ...n, readAt: n.readAt ?? at } : n,
      ),
    },
  }
}
export function notificationDestination(
  state: DemoWorkspaceState,
  actorId: string,
  id: string,
  identity: IdentitySource,
): string {
  const notification = notificationsFor(state, actorId, identity).find((n) => n.id === id)
  if (!notification) throw new Error("Không tìm thấy thông báo thuộc tài khoản của bạn.")
  const target = notification.target
  const query = (key: string, value: string) => new URLSearchParams({ [key]: value }).toString()
  if (target.kind === "task") {
    const task = state.tasks.find((t) => t.id === target.id)
    const project = state.projects.find((p) => p.id === task?.projectId)
    if (task && project) {
      const panel = /^(submitted|reviewed):/.test(notification.id) ? "results" : "details"
      if (project.pmId === actorId || project.ptcmId === actorId)
        return `/projects/${encodeURIComponent(project.id)}?${query("task", task.id)}&panel=${panel}`
      if (
        personalTasks(state.projects, state.tasks, actorId, identity).some((t) => t.id === task.id)
      )
        return `/my-work?${query("task", task.id)}&panel=${panel}`
    }
  }
  if (target.kind === "correction") {
    const request = state.work.requests.find((r) => r.id === target.id)
    if (
      request &&
      canReadRequest(state.projects, state.tasks, state.work, request, actorId, identity)
    )
      return `/notifications/corrections?${query("request", request.id)}`
  }
  if (target.kind === "profile-request") {
    const request = state.profileState.requests.find((r) => r.id === target.id)
    if (request?.employeeId === actorId) return `/my-profile?${query("request", request.id)}`
    if (request && isHr(actorId, identity))
      return `/profile-requests?${query("request", request.id)}`
  }
  if (target.kind === "skill-request") {
    const request = state.employeeSkills.requests.find((r) => r.id === target.id)
    if (request?.employeeId === actorId)
      return `/my-profile?tab=skills&${query("request", request.id)}`
    if (request && isDirector(actorId, identity))
      return `/skill-requests?${query("request", request.id)}`
  }
  throw new Error(
    "Bạn không còn quyền mở nội dung này hoặc tài nguyên không còn tồn tại. Thông báo không cấp thêm quyền.",
  )
}

export function notificationContext(
  state: DemoWorkspaceState,
  actorId: string,
  id: string,
  identity: IdentitySource,
): string {
  const entry = notificationsFor(state, actorId, identity).find((n) => n.id === id)
  if (!entry) return ""
  try {
    notificationDestination(state, actorId, id, identity)
  } catch {
    return "Nội dung không còn truy cập được"
  }
  if (entry.target.kind === "task")
    return state.tasks.find((t) => t.id === entry.target.id)?.code ?? "Công việc"
  if (entry.target.kind === "skill-request") {
    const request = state.employeeSkills.requests.find((r) => r.id === entry.target.id)
    return (
      state.skillCatalog.skills.find((s) => s.id === request?.skillId)?.name ?? "Đề nghị kỹ năng"
    )
  }
  if (entry.target.kind === "correction") return "Hiệu chỉnh nhật ký"
  return "Đề nghị thông tin cá nhân"
}
