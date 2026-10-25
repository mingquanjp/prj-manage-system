import type {
  Project,
  ProjectInput,
  FormErrors,
  Task,
  OutputConflict,
} from "../models/workspace.ts"
import { getPerson, isDirector } from "./identity.ts"
import { priorities } from "../models/workspace-labels.ts"

export function validateProjectInput(all: Project[], input: ProjectInput): FormErrors {
  const errors: FormErrors = {}
  if (!input.code.trim() || input.code.trim().length > 30)
    errors.code = "Nhập mã dự án, tối đa 30 ký tự."
  else if (all.some((p) => p.code === input.code.trim())) errors.code = "Mã dự án đã tồn tại."
  if (!input.name.trim() || input.name.trim().length > 150)
    errors.name = "Nhập tên dự án, tối đa 150 ký tự."
  if (!getPerson(input.pmId)?.active) errors.pmId = "Chọn nhân viên đang hoạt động làm PM."
  const validDate = (value: string) =>
    /^\d{4}-\d{2}-\d{2}$/.test(value) &&
    Number.isFinite(Date.parse(value)) &&
    new Date(value).toISOString().slice(0, 10) === value
  if (!validDate(input.start)) errors.start = "Nhập ngày bắt đầu hợp lệ."
  if (!validDate(input.end)) errors.end = "Nhập hạn dự kiến hợp lệ."
  else if (!errors.start && input.end < input.start)
    errors.end = "Hạn dự kiến không trước ngày bắt đầu."
  if (!priorities.includes(input.priority)) errors.priority = "Chọn mức ưu tiên."
  return errors
}

export function createProject(all: Project[], actorId: string, input: ProjectInput): Project[] {
  if (!isDirector(actorId) || !getPerson(actorId)?.active)
    throw new Error("Chỉ Giám đốc được tạo dự án.")
  const errors = validateProjectInput(all, input)
  if (Object.keys(errors).length) throw new Error(Object.values(errors)[0])
  return [
    ...all,
    {
      ...input,
      id: crypto.randomUUID(),
      code: input.code.trim(),
      name: input.name.trim(),
      description: input.description.trim(),
      status: "PLANNING",
    },
  ]
}

export function ptcmBlockReason(
  project: Project,
  personId: string,
  tasks: Task[],
  conflicts: OutputConflict[],
): string | undefined {
  if (!getPerson(personId)?.active) return "Nhân viên không còn hoạt động."
  if (
    tasks.some(
      (t) =>
        t.projectId === project.id &&
        t.assigneeId === personId &&
        ["ASSIGNED", "IN_PROGRESS", "PAUSED", "PENDING_REVIEW"].includes(t.status),
    )
  )
    return "Còn công việc đang được giao trong dự án. Cần xử lý phân công trước khi chỉ định PTCM."
  if (
    conflicts.some(
      (c) =>
        c.projectId === project.id &&
        c.personId === personId &&
        tasks.some(
          (t) =>
            t.id === c.taskId &&
            t.projectId === project.id &&
            !["DONE", "CLOSED", "CANCELLED"].includes(t.status),
        ),
    )
  )
    return "Còn đầu ra đã đóng góp cần nghiệm thu; không thể tự đánh giá kết quả của mình."
}

export function appointPtcm(
  project: Project,
  actorId: string,
  personId: string,
  tasks: Task[],
  conflicts: OutputConflict[],
): Project {
  if (
    project.pmId !== actorId ||
    !getPerson(actorId)?.active ||
    !["PLANNING", "IN_PROGRESS"].includes(project.status)
  )
    throw new Error("Chỉ PM của dự án đang lập kế hoạch hoặc triển khai được chỉ định PTCM.")
  const reason = ptcmBlockReason(project, personId, tasks, conflicts)
  if (reason) throw new Error(reason)
  return { ...project, ptcmId: personId }
}

export function startProject(project: Project, actorId: string, tasks: Task[]): Project {
  if (project.pmId !== actorId || !getPerson(actorId)?.active)
    throw new Error("Chỉ PM của dự án được bắt đầu triển khai.")
  if (project.status !== "PLANNING") throw new Error("Dự án không còn ở trạng thái lập kế hoạch.")
  if (!tasks.some((t) => t.projectId === project.id && t.status === "ASSIGNED"))
    throw new Error("Cần ít nhất một công việc đã phân công để bắt đầu dự án.")
  return {
    ...project,
    status: "IN_PROGRESS",
    executionStartedAt: new Date().toISOString(),
  }
}

export function progress(tasks: Task[]) {
  const effective = tasks.filter((t) => t.status !== "CANCELLED")
  return effective.length
    ? Math.round(
        (effective.filter((t) => ["DONE", "CLOSED"].includes(t.status)).length / effective.length) *
          100,
      )
    : null
}
