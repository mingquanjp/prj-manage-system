import type { EmployeeSkillsState } from "../models/employee-skills.ts"
import { createEmployeeSkillsState } from "./employee-skills.ts"
import type { SkillCatalogCommand, SkillCatalogState } from "../models/skill-catalog.ts"
import type { EmployeeProfile } from "../models/profile.ts"
import type { Task } from "../models/workspace.ts"
import { isDirector } from "./identity.ts"

export function createSkillCatalogState(): SkillCatalogState {
  return {
    categories: [
      { id: "programming", name: "Lập trình" },
      { id: "databases", name: "Cơ sở dữ liệu" },
    ],
    // Giữ khóa mẫu cũ để không đổi tham chiếu; sau đổi tên ID vẫn giữ nguyên.
    skills: [
      { id: "Node.js", name: "Node.js", categoryId: "programming" },
      { id: "PostgreSQL", name: "PostgreSQL", categoryId: "databases" },
      { id: "React", name: "React", categoryId: "programming" },
      { id: "TypeScript", name: "TypeScript", categoryId: "programming" },
    ],
  }
}

export function catalogSkillName(catalog: SkillCatalogState, id: string) {
  return catalog.skills.find((skill) => skill.id === id)?.name ?? "Kỹ năng không còn trong danh mục"
}

// Chỉ thống kê tham chiếu của các miền demo đã dựng, không trả dữ liệu hồ sơ/task cho Giám đốc.
export function skillReferences(
  catalog: SkillCatalogState,
  tasks: Task[],
  kind: "categories" | "skills",
  id: string,
  employeeSkills: EmployeeSkillsState = createEmployeeSkillsState(),
) {
  if (kind === "categories") {
    const count = catalog.skills.filter((skill) => skill.categoryId === id).length
    return count ? [`${count} kỹ năng thuộc nhóm này`] : []
  }
  const profiles = employeeSkills.official.filter((s) => s.skillId === id).length
  const requests = employeeSkills.requests.filter(
    (r) =>
      r.skillId === id ||
      r.evidence.some((e) => e.requirements.some((requirement) => requirement.skill === id)),
  ).length
  const requirements = tasks.filter((task) => task.requirements.some((r) => r.skill === id)).length
  return [
    ...(requests ? [`${requests} đề nghị hoặc căn cứ kỹ năng tham chiếu`] : []),
    ...(profiles ? [`${profiles} hồ sơ có kỹ năng chính thức`] : []),
    ...(requirements ? [`${requirements} công việc tham chiếu kỹ năng`] : []),
  ]
}

export function changeSkillCatalog(
  catalog: SkillCatalogState,
  profiles: EmployeeProfile[],
  tasks: Task[],
  actorId: string,
  command: SkillCatalogCommand,
  newId: string,
  employeeSkills: EmployeeSkillsState = createEmployeeSkillsState(),
): SkillCatalogState {
  if (
    !isDirector(actorId) ||
    profiles.find((p) => p.personId === actorId)?.employmentStatus !== "ACTIVE"
  )
    throw new Error("Chỉ Giám đốc đang hoạt động được quản lý danh mục kỹ năng.")
  const entries = catalog[command.kind]
  if (command.action !== "create") {
    const current = entries.find((entry) => entry.id === command.id)
    if (!current) throw new Error("Không tìm thấy mục này. Vui lòng kiểm tra lại danh sách.")
    if (
      current.name !== command.beforeName ||
      (command.kind === "skills" &&
        catalog.skills.find((s) => s.id === command.id)?.categoryId !== command.beforeCategoryId)
    )
      throw new Error("Danh mục đã thay đổi. Đóng hộp thoại và kiểm tra lại trước khi lưu.")
    if (command.action === "delete") {
      if (skillReferences(catalog, tasks, command.kind, command.id, employeeSkills).length)
        throw new Error(
          "Chưa thể xóa vì mục này còn được tham chiếu. Dữ liệu đang dùng và lịch sử phải được giữ nguyên.",
        )
      return command.kind === "categories"
        ? { ...catalog, categories: catalog.categories.filter((entry) => entry.id !== command.id) }
        : { ...catalog, skills: catalog.skills.filter((entry) => entry.id !== command.id) }
    }
  }
  const name = command.name.trim()
  if (!name || name.length > 100) throw new Error("Nhập tên từ 1 đến 100 ký tự.")
  if (
    entries.some(
      (entry) =>
        entry.id !== (command.action === "update" ? command.id : undefined) &&
        entry.name.toLocaleLowerCase("vi") === name.toLocaleLowerCase("vi"),
    )
  )
    throw new Error("Tên này đã có trong danh mục. Vui lòng dùng tên khác.")
  if (
    command.kind === "skills" &&
    !catalog.categories.some((category) => category.id === command.categoryId)
  )
    throw new Error("Chọn nhóm kỹ năng trong danh mục.")
  if (command.action === "create" && (!newId || entries.some((entry) => entry.id === newId)))
    throw new Error("Không thể tạo định danh mới. Vui lòng thử lại.")
  const id = command.action === "create" ? newId : command.id
  if (command.kind === "categories") {
    const entry = { id, name }
    return {
      ...catalog,
      categories:
        command.action === "create"
          ? [...catalog.categories, entry]
          : catalog.categories.map((old) => (old.id === id ? entry : old)),
    }
  }
  const entry = { id, name, categoryId: command.categoryId }
  return {
    ...catalog,
    skills:
      command.action === "create"
        ? [...catalog.skills, entry]
        : catalog.skills.map((old) => (old.id === id ? entry : old)),
  }
}
