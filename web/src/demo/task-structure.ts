import type { Task, Project, TaskGroup, TaskInput, FormErrors } from "../models/workspace.ts"
import { createSkillCatalogState } from "./skill-catalog.ts"
import { canAssign } from "./identity.ts"

export function isTaskFrozen(task: Task) {
  return Boolean(task.firstAssignedAt) || task.status !== "PENDING_ASSIGNMENT"
}

export function validateTaskInput(
  project: Project,
  groups: TaskGroup[],
  tasks: Task[],
  input: TaskInput,
  editingId?: string,
  catalog = createSkillCatalogState(),
): FormErrors {
  const errors: FormErrors = {}
  if (!input.code.trim() || input.code.trim().length > 30)
    errors.code = "Nhập mã công việc, tối đa 30 ký tự."
  else if (
    tasks.some(
      (t) => t.projectId === project.id && t.id !== editingId && t.code === input.code.trim(),
    )
  )
    errors.code = "Mã công việc đã được dùng trong dự án."
  if (!input.name.trim() || input.name.trim().length > 150)
    errors.name = "Nhập tên công việc, tối đa 150 ký tự."
  if (!groups.some((g) => g.id === input.groupId && g.projectId === project.id))
    errors.groupId = "Chọn nhóm thuộc dự án này."
  if (!input.description.trim()) errors.description = "Mô tả phạm vi và đầu ra mong đợi."
  if (!input.criteria.trim()) errors.criteria = "Bổ sung tiêu chí nghiệm thu."
  if (!Number.isFinite(input.estimate) || input.estimate <= 0)
    errors.estimate = "Ước tính phải là số giờ lớn hơn 0."
  if (!["Thấp", "Trung bình", "Cao", "Khẩn cấp"].includes(input.priority))
    errors.priority = "Chọn mức ưu tiên hợp lệ."
  if (Boolean(input.start) !== Boolean(input.end))
    errors.dates = "Nhập đủ ngày bắt đầu và kết thúc, hoặc để trống cả hai."
  else if (input.start && input.end) {
    const validDate = (value: string) =>
      /^\d{4}-\d{2}-\d{2}$/.test(value) &&
      !Number.isNaN(Date.parse(value)) &&
      new Date(value).toISOString().slice(0, 10) === value
    if (
      !validDate(input.start) ||
      !validDate(input.end) ||
      input.start > input.end ||
      input.start < project.start ||
      input.end > project.end
    )
      errors.dates = "Ngày công việc phải đúng thứ tự và nằm trong kế hoạch dự án."
  }
  const previous = tasks.find((t) => t.id === editingId && t.projectId === project.id)
  // Gia hạn có lý do/lịch sử dùng thao tác riêng; không ghi đè cặp ngày đã lưu.
  if (
    previous?.start &&
    previous.end &&
    (input.start !== previous.start || input.end !== previous.end)
  ) {
    errors.dates =
      "Giữ nguyên kế hoạch đã lưu trong biểu mẫu đặc tả. Dùng thao tác Gia hạn tại chi tiết công việc."
  }
  if (!input.requirements.length) errors.requirements = "Thêm ít nhất một yêu cầu kỹ năng."
  input.requirements.forEach((r, i) => {
    if (!catalog.skills.some((skill) => skill.id === r.skill))
      errors[`skill-${i}`] = "Chọn kỹ năng trong danh mục."
    else if (input.requirements.some((other, j) => j !== i && other.skill === r.skill))
      errors[`skill-${i}`] = "Kỹ năng này đã có trong công việc."
    if (!Number.isInteger(r.level) || r.level < 1 || r.level > 4)
      errors[`level-${i}`] = "Chọn một trong bốn mức."
    // Chỉ kiểm tra dữ liệu số hữu hạn cho demo, chưa chốt miền trọng số nghiệp vụ.
    if (!Number.isFinite(r.weight)) errors[`weight-${i}`] = "Nhập giá trị số minh họa."
  })
  return errors
}

export function saveTask(
  project: Project,
  actorId: string,
  groups: TaskGroup[],
  tasks: Task[],
  input: TaskInput,
  editingId?: string,
  catalog = createSkillCatalogState(),
): Task[] {
  if (!canAssign(project, actorId))
    throw new Error("Chỉ PTCM của dự án đang lập kế hoạch/triển khai được sửa cấu trúc.")
  const existing = editingId
    ? tasks.find((t) => t.id === editingId && t.projectId === project.id)
    : undefined
  if (editingId && (!existing || isTaskFrozen(existing)))
    throw new Error("Công việc đã từng phân công không được sửa lại đặc tả bằng biểu mẫu này.")
  const errors = validateTaskInput(project, groups, tasks, input, editingId, catalog)
  if (Object.keys(errors).length) throw new Error(Object.values(errors).join(" "))
  const group = groups.find((g) => g.id === input.groupId)!
  const task: Task = {
    ...existing,
    ...structuredClone(input),
    id: existing?.id ?? crypto.randomUUID(),
    projectId: project.id,
    code: input.code.trim(),
    name: input.name.trim(),
    description: input.description.trim(),
    criteria: input.criteria.trim(),
    group: group.name,
    start: input.start || undefined,
    end: input.end || undefined,
    originalStart: existing?.originalStart ?? existing?.start ?? (input.start || undefined),
    originalEnd: existing?.originalEnd ?? existing?.end ?? (input.end || undefined),
    status: existing?.status ?? "PENDING_ASSIGNMENT",
  }
  return existing ? tasks.map((t) => (t.id === task.id ? task : t)) : [...tasks, task]
}

export function saveGroup(
  project: Project,
  actorId: string,
  groups: TaskGroup[],
  tasks: Task[],
  input: Pick<TaskGroup, "name" | "description">,
  editingId?: string,
) {
  if (!canAssign(project, actorId))
    throw new Error("Chỉ PTCM của dự án đang lập kế hoạch/triển khai được sửa nhóm.")
  const name = input.name.trim()
  if (!name || name.length > 150) throw new Error("Nhập tên nhóm, tối đa 150 ký tự.")
  if (groups.some((g) => g.projectId === project.id && g.id !== editingId && g.name === name))
    throw new Error("Tên nhóm đã có trong dự án.")
  if (editingId && !groups.some((g) => g.id === editingId && g.projectId === project.id))
    throw new Error("Không tìm thấy nhóm trong dự án.")
  const group = {
    id: editingId ?? crypto.randomUUID(),
    projectId: project.id,
    name,
    description: input.description.trim(),
  }
  return {
    groups: editingId ? groups.map((g) => (g.id === editingId ? group : g)) : [...groups, group],
    tasks: tasks.map((t) =>
      t.projectId === project.id && t.groupId === editingId ? { ...t, group: name } : t,
    ),
  }
}
