import {
  executionBlockReason,
  getPerson,
  type OutputConflict,
  type Project,
  type Task,
} from "./model.ts"

// Read model trong bộ nhớ; không thay schema hoặc service/audit server.
export type WorkLog = {
  id: string
  assignmentId: string
  taskId: string
  employeeId: string
  workDate: string
  minutes: number
  description: string
  overLimitReason: string
  createdAt: string
  updatedAt: string
  deletedAt?: string
}
export type LogInput = Pick<
  WorkLog,
  "workDate" | "minutes" | "description" | "overLimitReason"
>
export type LogEdit = {
  before: WorkLog
  after: WorkLog
  actorId: string
  at: string
  reason: string
}
export type Submission = {
  id: string
  taskId: string
  assignmentId: string
  employeeId: string
  submittedAt: string
  summary: string
  links: string[]
  logSnapshot: {
    schemaVersion: 2
    logs: Pick<
      WorkLog,
      | "id"
      | "assignmentId"
      | "workDate"
      | "minutes"
      | "description"
      | "overLimitReason"
    >[]
  }
  review?: {
    result: "ACCEPTED" | "REJECTED"
    reviewerId: string
    at: string
    comment: string
  }
}
export type WorkState = {
  logs: WorkLog[]
  history: LogEdit[]
  submissions: Submission[]
}
export const emptyWorkState = (): WorkState => ({
  logs: [],
  history: [],
  submissions: [],
})
export const vietnamDay = (now = new Date()) =>
  new Intl.DateTimeFormat("en-CA", {
    timeZone: "Asia/Ho_Chi_Minh",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(now)
export const formatMinutes = (minutes: number) =>
  `${Math.floor(minutes / 60)} giờ ${minutes % 60} phút`

export function logBlockReason(project: Project, task: Task, actorId: string) {
  const block = executionBlockReason(project, task, actorId)
  if (block)
    return ["PENDING_REVIEW", "DONE", "CLOSED", "CANCELLED"].includes(
      task.status,
    ) && project.status === "IN_PROGRESS"
      ? "Nhật ký đã khóa sửa trực tiếp theo trạng thái công việc. Hiệu chỉnh sau khóa chưa dựng trong đợt này."
      : block
  if (task.status === "ASSIGNED")
    return "Bấm Bắt đầu ở phần Chi tiết trước khi ghi nhật ký công việc."
  if (task.status === "PAUSED")
    return "Công việc đang tạm dừng. Bấm Tiếp tục ở phần Chi tiết trước khi ghi hoặc chỉnh sửa nhật ký."
  if (
    !task.currentAssignment ||
    task.currentAssignment.employeeId !== actorId ||
    !project.executionStartedAt
  )
    return "Thiếu dữ liệu lượt phân công hoặc thời điểm triển khai thực tế trong bản mô phỏng."
  return undefined
}

export function taskLogs(state: WorkState, task: Task, actorId: string) {
  return state.logs.filter(
    (l) =>
      l.taskId === task.id &&
      l.employeeId === actorId &&
      l.assignmentId === task.currentAssignment?.id &&
      !l.deletedAt,
  )
}
export function dailyMinutes(
  state: WorkState,
  actorId: string,
  day: string,
  exceptId?: string,
) {
  return state.logs
    .filter(
      (l) =>
        l.employeeId === actorId &&
        l.workDate === day &&
        !l.deletedAt &&
        l.id !== exceptId,
    )
    .reduce((sum, l) => sum + l.minutes, 0)
}
function context(
  projects: Project[],
  tasks: Task[],
  actorId: string,
  taskId: string,
) {
  const task = tasks.find((t) => t.id === taskId)
  const project = projects.find((p) => p.id === task?.projectId)
  if (!task || !project) throw new Error("Không tìm thấy công việc.")
  const block = logBlockReason(project, task, actorId)
  if (block) throw new Error(block)
  return { task, project }
}
function validInput(
  state: WorkState,
  actorId: string,
  project: Project,
  task: Task,
  input: LogInput,
  now: Date,
  exceptId?: string,
) {
  const errors: Record<string, string> = {}
  const date = /^\d{4}-\d{2}-\d{2}$/.test(input.workDate)
    ? new Date(`${input.workDate}T00:00:00Z`)
    : new Date(NaN)
  if (
    !Number.isFinite(date.getTime()) ||
    date.toISOString().slice(0, 10) !== input.workDate
  )
    errors.workDate = "Chọn ngày làm việc hợp lệ."
  else if (input.workDate > vietnamDay(now))
    errors.workDate = "Không ghi nhật ký cho ngày tương lai."
  else if (
    input.workDate < vietnamDay(new Date(task.currentAssignment!.assignedAt)) ||
    input.workDate < vietnamDay(new Date(project.executionStartedAt!))
  )
    errors.workDate =
      "Ngày làm phải thuộc lượt được giao và giai đoạn dự án đã triển khai thực tế."
  if (!Number.isSafeInteger(input.minutes) || input.minutes <= 0)
    errors.minutes = "Thời lượng phải là số phút nguyên lớn hơn 0."
  if (!input.description.trim())
    errors.description = "Nhập nội dung công việc thực tế đã làm."
  const total =
    dailyMinutes(state, actorId, input.workDate, exceptId) + input.minutes
  if (total > 1440)
    errors.minutes =
      "Tổng nhật ký cá nhân trong ngày vượt 24 giờ (xuyên dự án)."
  if (total > 480 && !input.overLimitReason.trim())
    errors.overLimitReason =
      "Tổng ngày vượt 8 giờ: cần lý do cho phần thời gian thực tế này."
  return errors
}
export class WorkValidationError extends Error {
  fields: Record<string, string>
  constructor(fields: Record<string, string>) {
    super(Object.values(fields)[0])
    this.fields = fields
  }
}

export function saveLog(
  projects: Project[],
  tasks: Task[],
  state: WorkState,
  actorId: string,
  taskId: string,
  input: LogInput,
  edit?: { before: WorkLog; reason: string },
  now = new Date(),
): WorkState {
  const { task, project } = context(projects, tasks, actorId, taskId)
  const before = edit
    ? currentLog(state, task, actorId, edit.before)
    : undefined
  const errors = validInput(
    state,
    actorId,
    project,
    task,
    input,
    now,
    before?.id,
  )
  if (edit && !edit.reason.trim())
    errors.reason = "Cần lý do chỉnh sửa nhật ký."
  if (Object.keys(errors).length) throw new WorkValidationError(errors)
  const at = now.toISOString()
  const next: WorkLog = {
    id: before?.id ?? crypto.randomUUID(),
    assignmentId: task.currentAssignment!.id,
    taskId,
    employeeId: actorId,
    ...input,
    description: input.description.trim(),
    overLimitReason: input.overLimitReason.trim(),
    createdAt: before?.createdAt ?? at,
    updatedAt: at,
  }
  return {
    ...state,
    logs: before
      ? state.logs.map((l) => (l.id === before.id ? next : l))
      : [...state.logs, next],
    history: before
      ? [
          ...state.history,
          {
            before: structuredClone(before),
            after: structuredClone(next),
            actorId,
            at,
            reason: edit!.reason.trim(),
          },
        ]
      : state.history,
  }
}
function currentLog(
  state: WorkState,
  task: Task,
  actorId: string,
  expected: WorkLog,
) {
  const log = state.logs.find((l) => l.id === expected.id)
  if (
    !log ||
    log.deletedAt ||
    log.assignmentId !== task.currentAssignment?.id ||
    log.employeeId !== actorId ||
    log.taskId !== task.id
  )
    throw new Error("Không còn quyền sửa nhật ký này.")
  if (JSON.stringify(log) !== JSON.stringify(expected))
    throw new Error(
      "Nhật ký đã thay đổi. Đóng biểu mẫu và mở lại để xem dữ liệu mới.",
    )
  return log
}
export function deleteLog(
  projects: Project[],
  tasks: Task[],
  state: WorkState,
  actorId: string,
  taskId: string,
  expected: WorkLog,
  reason: string,
  now = new Date(),
): WorkState {
  const { task } = context(projects, tasks, actorId, taskId)
  const before = currentLog(state, task, actorId, expected)
  if (!reason.trim())
    throw new WorkValidationError({ reason: "Cần lý do xóa nhật ký." })
  const at = now.toISOString(),
    after = { ...before, deletedAt: at, updatedAt: at }
  return {
    ...state,
    logs: state.logs.map((l) => (l.id === before.id ? after : l)),
    history: [
      ...state.history,
      {
        before: structuredClone(before),
        after: structuredClone(after),
        actorId,
        at,
        reason: reason.trim(),
      },
    ],
  }
}

export function validateResultInput(summary: string, linkText: string) {
  const errors: Record<string, string> = {}
  if (!summary.trim()) errors.summary = "Nhập tóm tắt kết quả thực hiện."
  const links = linkText
    .split(/\r?\n/)
    .map((s) => s.trim())
    .filter(Boolean)
  for (const link of links) {
    try {
      const url = new URL(link)
      if (
        !["http:", "https:"].includes(url.protocol) ||
        url.username ||
        url.password
      )
        throw new Error()
    } catch {
      errors.links =
        "Mỗi dòng là một URL http/https hợp lệ, không chứa thông tin đăng nhập."
    }
  }
  if (Object.keys(errors).length) throw new WorkValidationError(errors)
  return links
}

export function submitResult(
  projects: Project[],
  tasks: Task[],
  state: WorkState,
  actorId: string,
  taskId: string,
  summary: string,
  linkText: string,
  now = new Date(),
) {
  const { task } = context(projects, tasks, actorId, taskId)
  if (state.submissions.some((s) => s.taskId === task.id && !s.review))
    throw new Error(
      "Công việc vẫn còn lần nộp chưa có kết luận. Vui lòng kiểm tra lại.",
    )
  if (task.status !== "IN_PROGRESS")
    throw new Error("Chỉ nộp kết quả khi công việc đang thực hiện.")
  const links = validateResultInput(summary, linkText)
  const at = now.toISOString()
  const submission: Submission = {
    id: crypto.randomUUID(),
    taskId,
    assignmentId: task.currentAssignment!.id,
    employeeId: actorId,
    submittedAt: at,
    summary: summary.trim(),
    links,
    logSnapshot: {
      schemaVersion: 2,
      logs: taskLogs(state, task, actorId).map(
        ({
          id,
          assignmentId,
          workDate,
          minutes,
          description,
          overLimitReason,
        }) => ({
          id,
          assignmentId,
          workDate,
          minutes,
          description,
          overLimitReason,
        }),
      ),
    },
  }
  return {
    work: { ...state, submissions: [...state.submissions, submission] },
    tasks: tasks.map(
      (t): Task =>
        t.id !== task.id
          ? t
          : {
              ...t,
              status: "PENDING_REVIEW",
              executionHistory: [
                ...(t.executionHistory ?? []),
                { actorId, at, from: "IN_PROGRESS", to: "PENDING_REVIEW" },
              ],
            },
    ),
  }
}

export type ReviewResult = "ACCEPTED" | "REJECTED"

// Guard của adapter demo: UI và thao tác xác nhận cùng đối chiếu dữ liệu hiện tại.
export function reviewBlockReason(
  project: Project,
  task: Task,
  state: WorkState,
  actorId: string,
  submissionId: string,
  conflicts: OutputConflict[],
) {
  if (
    project.id !== task.projectId ||
    project.ptcmId !== actorId ||
    !getPerson(actorId)?.active
  )
    return "Chỉ PTCM hiện hành của dự án được quyết định nghiệm thu."
  if (project.status !== "IN_PROGRESS" || task.status !== "PENDING_REVIEW")
    return "Chỉ nghiệm thu công việc đang chờ trong dự án đang triển khai."
  const submissions = state.submissions.filter((s) => s.taskId === task.id)
  const submission = submissions.at(-1)
  if (
    !submission ||
    submission.id !== submissionId ||
    submission.review ||
    submissions.some((s) => s.review?.result === "ACCEPTED")
  )
    return "Lần nộp đã thay đổi hoặc đã có kết luận. Mở lại để xem dữ liệu mới."
  if (
    !task.currentAssignment ||
    task.currentAssignment.endedAt ||
    submission.assignmentId !== task.currentAssignment.id ||
    submission.employeeId !== task.currentAssignment.employeeId ||
    submission.employeeId !== task.assigneeId
  )
    return "Lần nộp không còn khớp lượt phân công hiện hành."
  if (
    submission.employeeId === actorId ||
    submissions.some((s) => s.employeeId === actorId) ||
    conflicts.some(
      (c) =>
        c.projectId === project.id &&
        c.taskId === task.id &&
        c.personId === actorId,
    )
  )
    return "Không được tự nghiệm thu đầu ra có đóng góp của chính mình."
  return undefined
}

export function reviewSubmission(
  projects: Project[],
  tasks: Task[],
  state: WorkState,
  actorId: string,
  submissionId: string,
  result: ReviewResult,
  comment: string,
  conflicts: OutputConflict[],
  now = new Date(),
) {
  const submission = state.submissions.find((s) => s.id === submissionId)
  const task = tasks.find((t) => t.id === submission?.taskId)
  const project = projects.find((p) => p.id === task?.projectId)
  if (!submission || !task || !project)
    throw new Error("Không tìm thấy lần nộp.")
  const block = reviewBlockReason(
    project,
    task,
    state,
    actorId,
    submissionId,
    conflicts,
  )
  if (block) throw new Error(block)
  if (!["ACCEPTED", "REJECTED"].includes(result))
    throw new Error("Kết luận không hợp lệ.")
  if (result === "REJECTED" && !comment.trim())
    throw new WorkValidationError({
      comment: "Nhập lý do yêu cầu chỉnh sửa đầu ra.",
    })
  const at = now.toISOString()
  const to = result === "ACCEPTED" ? "DONE" : "IN_PROGRESS"
  return {
    work: {
      ...state,
      submissions: state.submissions.map((s) =>
        s.id === submissionId
          ? {
              ...s,
              review: {
                result,
                reviewerId: actorId,
                at,
                comment: comment.trim(),
              },
            }
          : s,
      ),
    },
    tasks: tasks.map(
      (t): Task =>
        t.id !== task.id
          ? t
          : {
              ...t,
              status: to,
              ...(result === "ACCEPTED"
                ? {
                    acceptedAt: at,
                    currentAssignment: {
                      ...t.currentAssignment!,
                      endedAt: at,
                      endedBy: actorId,
                      endReason: "DONE",
                    },
                  }
                : {}),
              executionHistory: [
                ...(t.executionHistory ?? []),
                {
                  actorId,
                  at,
                  from: "PENDING_REVIEW",
                  to,
                  ...(comment.trim() ? { reason: comment.trim() } : {}),
                },
              ],
            },
    ),
  }
}
