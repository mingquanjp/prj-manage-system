import { defaultIdentity, type IdentitySource } from "./identity.ts"
import { getPerson } from "./identity.ts"
import { personalTasks } from "./execution.ts"
import type { Project, Task } from "../models/workspace.ts"
import {
  rememberContributions,
  validInput,
  vietnamDay,
  WorkValidationError,
  type Contribution,
  type LogChangeRequest,
  type LogInput,
  type WorkLog,
  type WorkState,
} from "./effort.ts"

export const changeLabels = { ADD: "Bổ sung", UPDATE: "Chỉnh sửa", DELETE: "Xóa" }
export const requestLabels = {
  PENDING: "Chờ duyệt",
  APPROVED: "Đã chấp thuận",
  REJECTED: "Đã từ chối",
  WITHDRAWN: "Đã rút",
}

export function contributionFor(state: WorkState, tasks: Task[], assignmentId: string) {
  const task = tasks.find((entry) => entry.currentAssignment?.id === assignmentId)
  if (task?.currentAssignment)
    return {
      taskId: task.id,
      projectId: task.projectId,
      taskCode: task.code,
      taskName: task.name,
      assignment: task.currentAssignment,
    }
  return state.contributions.find((entry) => entry.assignment.id === assignmentId)
}

// Lối vào riêng cho đóng góp không còn xuất hiện trong danh sách thực hiện.
// Bao gồm lượt đã kết thúc vẫn giữ cùng ID khi chính chủ được bổ nhiệm PTCM.
export function historicalContributions(
  projects: Project[],
  tasks: Task[],
  state: WorkState,
  actorId: string,
  identity: IdentitySource = defaultIdentity,
) {
  if (!getPerson(actorId, identity)?.active) return []
  const visibleAssignments = new Set(
    personalTasks(projects, tasks, actorId, identity).map((task) => task.currentAssignment?.id),
  )
  return state.contributions.filter(
    (entry) =>
      entry.assignment.employeeId === actorId &&
      !visibleAssignments.has(entry.assignment.id) &&
      projects.some((project) => project.id === entry.projectId) &&
      tasks.some((task) => task.id === entry.taskId && task.projectId === entry.projectId),
  )
}

export function logDateBounds(project: Project, contribution?: Contribution, now = new Date()) {
  const min =
    project.executionStartedAt && contribution
      ? [
          vietnamDay(new Date(project.executionStartedAt)),
          vietnamDay(new Date(contribution.assignment.assignedAt)),
        ].sort()[1]
      : undefined
  const today = vietnamDay(now)
  const end = contribution?.assignment.endedAt
  const max = end ? [today, vietnamDay(new Date(end))].sort()[0] : today
  return { min, max }
}

export function defaultLogDate(project: Project, contribution?: Contribution, now = new Date()) {
  const { min, max } = logDateBounds(project, contribution, now)
  return min && min <= max ? max : ""
}

export function correctionBlockReason(
  project: Project,
  task: Task,
  contribution: Contribution,
  actorId: string,
  identity: IdentitySource = defaultIdentity,
) {
  if (
    !getPerson(actorId, identity)?.active ||
    contribution.assignment.employeeId !== actorId ||
    contribution.taskId !== task.id ||
    contribution.projectId !== project.id
  )
    return "Chỉ chính chủ lượt phân công được gửi yêu cầu hiệu chỉnh."
  if (!["IN_PROGRESS", "COMPLETED"].includes(project.status))
    return "Dự án chưa triển khai hoặc đã đóng/hủy; không được hiệu chỉnh nhật ký."
  if (!project.executionStartedAt)
    return "Thiếu thời điểm triển khai thực tế trong dữ liệu mô phỏng."
  const current =
    task.currentAssignment?.id === contribution.assignment.id && !contribution.assignment.endedAt
  if (current && ["ASSIGNED", "PAUSED"].includes(task.status))
    return "Bắt đầu/Tiếp tục công việc trước khi ghi hoặc sửa trực tiếp, không gửi yêu cầu để bỏ qua bước này."
  if (current && task.status === "IN_PROGRESS" && project.status === "IN_PROGRESS")
    return "Nhật ký còn được sửa trực tiếp; không cần gửi yêu cầu hiệu chỉnh."
  if (!current && !contribution.assignment.endedAt)
    return "Thiếu thời điểm kết thúc lượt giao cũ; chưa thể xác minh khoảng thực hiện."
  if (["DONE", "CLOSED", "CANCELLED"].includes(task.status) && !contribution.assignment.endedAt)
    return "Thiếu thời điểm kết thúc thực tế; không tự suy từ hạn kế hoạch."
  return undefined
}

export function requestContext(
  projects: Project[],
  tasks: Task[],
  state: WorkState,
  assignmentId: string,
) {
  const contribution = contributionFor(state, tasks, assignmentId)
  const task = tasks.find((t) => t.id === contribution?.taskId)
  const project = projects.find((p) => p.id === contribution?.projectId)
  if (!contribution || !task || !project) throw new Error("Không tìm thấy lượt đóng góp.")
  return { contribution, task, project }
}

export function canReadRequest(
  projects: Project[],
  tasks: Task[],
  state: WorkState,
  request: LogChangeRequest,
  actorId: string,
  identity: IdentitySource = defaultIdentity,
) {
  if (!getPerson(actorId, identity)?.active) return false
  const contribution = contributionFor(state, tasks, request.assignmentId)
  const project = projects.find((p) => p.id === contribution?.projectId)
  return Boolean(project && (request.requestedBy === actorId || project.ptcmId === actorId))
}

// Chỉ đối chiếu các giá trị nền nghiệp vụ, không dùng version hay thứ tự khóa JSON.
export function sameLogBase(a: WorkLog, b: WorkLog) {
  return (
    a.id === b.id &&
    a.assignmentId === b.assignmentId &&
    a.employeeId === b.employeeId &&
    a.taskId === b.taskId &&
    a.workDate === b.workDate &&
    a.minutes === b.minutes &&
    a.description === b.description &&
    a.overLimitReason === b.overLimitReason &&
    a.deletedAt === b.deletedAt
  )
}
export function requestConflict(state: WorkState, request: LogChangeRequest) {
  if (request.operation === "ADD") return undefined
  const log = state.logs.find((l) => l.id === request.targetLogId)
  return !log || !request.before || log.deletedAt || !sameLogBase(log, request.before)
    ? "Nhật ký hiện tại khác dữ liệu khi gửi. Không ghi đè; hãy từ chối/rút và gửi yêu cầu mới."
    : undefined
}

export function sendLogChange(
  projects: Project[],
  tasks: Task[],
  state: WorkState,
  actorId: string,
  assignmentId: string,
  operation: LogChangeRequest["operation"],
  reason: string,
  proposed?: LogInput,
  expected?: WorkLog,
  requestId: string = crypto.randomUUID(),
  now = new Date(),
  identity: IdentitySource = defaultIdentity,
): WorkState {
  const { contribution, task, project } = requestContext(projects, tasks, state, assignmentId)
  const block = correctionBlockReason(project, task, contribution, actorId, identity)
  if (block) throw new Error(block)
  if (!["ADD", "UPDATE", "DELETE"].includes(operation))
    throw new Error("Loại yêu cầu không hợp lệ.")
  if (!reason.trim()) throw new WorkValidationError({ reason: "Nhập lý do và căn cứ hiệu chỉnh." })
  if ((operation === "ADD" && expected) || (operation === "DELETE" && proposed))
    throw new Error("Dữ liệu không phù hợp loại yêu cầu.")
  const normalized = proposed
    ? {
        ...proposed,
        description: proposed.description.trim(),
        overLimitReason: proposed.overLimitReason.trim(),
      }
    : undefined
  const retried = state.requests.find((r) => r.id === requestId)
  if (retried) {
    if (
      retried.requestedBy !== actorId ||
      retried.assignmentId !== assignmentId ||
      retried.operation !== operation ||
      retried.reason !== reason.trim() ||
      JSON.stringify(retried.proposed) !== JSON.stringify(normalized) ||
      retried.targetLogId !== expected?.id
    )
      throw new Error("Mã gửi lại không khớp nội dung yêu cầu trước.")
    return state
  }
  let before: WorkLog | undefined
  if (operation !== "ADD") {
    before = state.logs.find((l) => l.id === expected?.id)
    if (
      !before ||
      !expected ||
      before.deletedAt ||
      before.assignmentId !== assignmentId ||
      before.employeeId !== actorId ||
      before.taskId !== task.id ||
      !sameLogBase(before, expected)
    )
      throw new Error(
        "Nhật ký đã đổi hoặc không thuộc lượt đóng góp của bạn. Mở lại để xem dữ liệu mới.",
      )
    if (state.requests.some((r) => r.status === "PENDING" && r.targetLogId === before!.id))
      throw new Error("Nhật ký này đã có yêu cầu chờ duyệt. Rút hoặc chờ xử lý trước khi gửi mới.")
  }
  if (operation !== "DELETE") {
    if (!normalized) throw new Error("Thiếu nội dung đề nghị.")
    const errors = validInput(
      state,
      actorId,
      project,
      { ...task, currentAssignment: contribution.assignment },
      normalized,
      now,
      before?.id,
    )
    if (Object.keys(errors).length) throw new WorkValidationError(errors)
  }
  const at = now.toISOString()
  const request: LogChangeRequest = {
    id: requestId,
    assignmentId,
    operation,
    reason: reason.trim(),
    requestedBy: actorId,
    requestedAt: at,
    status: "PENDING",
    ...(before ? { targetLogId: before.id, before: structuredClone(before) } : {}),
    ...(normalized ? { proposed: normalized } : {}),
  }
  return {
    ...rememberContributions(state, tasks),
    requests: [...state.requests, request],
    notifications: project.ptcmId
      ? [
          ...state.notifications,
          {
            id: crypto.randomUUID(),
            recipientId: project.ptcmId,
            requestId,
            at,
            message: "Có yêu cầu hiệu chỉnh nhật ký mới cần xử lý.",
          },
        ]
      : state.notifications,
  }
}

export function resolveLogChange(
  projects: Project[],
  tasks: Task[],
  state: WorkState,
  actorId: string,
  requestId: string,
  action: "APPROVED" | "REJECTED" | "WITHDRAWN",
  reason: string,
  now = new Date(),
  identity: IdentitySource = defaultIdentity,
): WorkState {
  const request = state.requests.find((r) => r.id === requestId)
  if (!request || !canReadRequest(projects, tasks, state, request, actorId, identity))
    throw new Error("Không có quyền xử lý yêu cầu.")
  if (request.status !== "PENDING") throw new Error("Yêu cầu đã kết thúc; không xử lý lại.")
  const { contribution, task, project } = requestContext(
    projects,
    tasks,
    state,
    request.assignmentId,
  )
  if (!["IN_PROGRESS", "COMPLETED"].includes(project.status))
    throw new Error("Dự án đã khóa; yêu cầu chỉ được đọc.")
  if (!["APPROVED", "REJECTED", "WITHDRAWN"].includes(action))
    throw new Error("Thao tác không hợp lệ.")
  if (action === "WITHDRAWN" ? request.requestedBy !== actorId : project.ptcmId !== actorId)
    throw new Error("Chỉ chính chủ được rút; chỉ PTCM hiện hành được xét duyệt.")
  if (action === "REJECTED" && !reason.trim())
    throw new WorkValidationError({ reason: "Nhập lý do từ chối." })
  const at = now.toISOString()
  let logs = state.logs,
    history = state.history,
    resultLogId: string | undefined
  if (action === "APPROVED") {
    const conflict = requestConflict(state, request)
    if (conflict) throw new Error(conflict)
    if (!project.executionStartedAt) throw new Error("Thiếu thời điểm triển khai thực tế.")
    if (
      (["DONE", "CLOSED", "CANCELLED"].includes(task.status) ||
        task.currentAssignment?.id !== contribution.assignment.id) &&
      !contribution.assignment.endedAt
    )
      throw new Error(
        "Thiếu thời điểm kết thúc thực tế của lượt đóng góp; không thể duyệt ghi đè khoảng thời gian.",
      )
    const before = state.logs.find((log) => log.id === request.targetLogId)
    if (request.operation !== "DELETE") {
      if (!request.proposed) throw new Error("Thiếu nội dung đề nghị.")
      const errors = validInput(
        state,
        contribution.assignment.employeeId,
        project,
        { ...task, currentAssignment: contribution.assignment },
        request.proposed,
        now,
        before?.id,
      )
      if (Object.keys(errors).length) throw new WorkValidationError(errors)
    }
    if (request.operation === "ADD") {
      resultLogId = crypto.randomUUID()
      logs = [
        ...logs,
        {
          ...request.proposed!,
          id: resultLogId,
          assignmentId: request.assignmentId,
          taskId: contribution.taskId,
          employeeId: contribution.assignment.employeeId,
          createdAt: at,
          updatedAt: at,
        },
      ]
    } else {
      if (
        !before ||
        before.assignmentId !== request.assignmentId ||
        before.employeeId !== request.requestedBy
      )
        throw new Error("Nhật ký không thuộc đúng lượt đóng góp.")
      const after =
        request.operation === "DELETE"
          ? { ...before, deletedAt: at, updatedAt: at }
          : { ...before, ...request.proposed!, updatedAt: at }
      logs = logs.map((log) => (log.id === before.id ? after : log))
      history = [
        ...history,
        {
          before: structuredClone(before),
          after: structuredClone(after),
          actorId,
          at,
          reason: request.reason,
          requestId,
        },
      ]
    }
  }
  return {
    ...state,
    logs,
    history,
    requests: state.requests.map((r) =>
      r.id === requestId
        ? {
            ...r,
            status: action,
            resolvedBy: actorId,
            resolvedAt: at,
            decisionReason: reason.trim(),
            ...(resultLogId ? { resultLogId } : {}),
          }
        : r,
    ),
    notifications: [
      ...state.notifications,
      {
        id: crypto.randomUUID(),
        recipientId: action === "WITHDRAWN" ? (project.ptcmId ?? actorId) : request.requestedBy,
        requestId,
        at,
        message:
          action === "WITHDRAWN"
            ? "Một yêu cầu hiệu chỉnh đã được người gửi rút."
            : `Yêu cầu hiệu chỉnh của bạn ${action === "APPROVED" ? "đã được chấp thuận" : "đã bị từ chối"}.`,
      },
    ],
  }
}
