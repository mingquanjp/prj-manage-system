import type { EmployeeProfile } from "../models/profile.ts"
import type {
  OrganizationCommand,
  OrganizationKind,
  OrganizationState,
} from "../models/organization.ts"
import { isHr } from "./identity.ts"

export function createOrganizationState(): OrganizationState {
  return {
    departments: [
      { id: "engineering", name: "Công nghệ thông tin" },
      { id: "human-resources", name: "Nhân sự" },
    ],
    positions: [
      { id: "software-engineer", name: "Kỹ sư phần mềm" },
      { id: "backend-engineer", name: "Kỹ sư Backend" },
      { id: "frontend-engineer", name: "Kỹ sư Frontend" },
      { id: "project-manager", name: "Quản lý dự án" },
      { id: "hr-specialist", name: "Chuyên viên nhân sự" },
      { id: "director", name: "Giám đốc" },
    ],
  }
}

// Tham chiếu mẫu được khai báo riêng, không suy chức danh/quyền từ tên actor.
export const initialEmployment: Record<
  string,
  { departmentId: string | null; positionId: string }
> = {
  hr: { departmentId: "human-resources", positionId: "hr-specialist" },
  ha: { departmentId: null, positionId: "director" },
  khoa: { departmentId: "engineering", positionId: "software-engineer" },
  thanh: { departmentId: "engineering", positionId: "project-manager" },
  an: { departmentId: "engineering", positionId: "backend-engineer" },
  binh: { departmentId: "engineering", positionId: "frontend-engineer" },
  dung: { departmentId: "engineering", positionId: "software-engineer" },
  cuong: { departmentId: "engineering", positionId: "frontend-engineer" },
}

export function organizationReferences(
  profiles: EmployeeProfile[],
  kind: OrganizationKind,
  id: string,
) {
  return profiles.filter(
    (profile) => (kind === "departments" ? profile.departmentId : profile.positionId) === id,
  )
}

export function changeOrganization(
  state: OrganizationState,
  profiles: EmployeeProfile[],
  actorId: string,
  command: OrganizationCommand,
  newId: string,
): OrganizationState {
  if (!isHr(actorId)) throw new Error("Chỉ HR được quản lý phòng ban và vị trí công việc.")
  const entries = state[command.kind]
  if (command.action !== "create") {
    const current = entries.find((entry) => entry.id === command.id)
    if (!current) throw new Error("Không tìm thấy mục này. Vui lòng kiểm tra lại danh sách.")
    if (current.name !== command.beforeName)
      throw new Error("Tên đã thay đổi. Đóng hộp thoại và kiểm tra lại trước khi lưu.")
    if (command.action === "delete") {
      const references = organizationReferences(profiles, command.kind, command.id)
      if (references.length)
        throw new Error(
          `Chưa thể xóa vì có ${references.length} hồ sơ nhân viên đang sử dụng. HR cần cập nhật các hồ sơ liên quan trước.`,
        )
      return { ...state, [command.kind]: entries.filter((entry) => entry.id !== command.id) }
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
  if (command.action === "create") {
    if (!newId || entries.some((entry) => entry.id === newId))
      throw new Error("Không thể tạo định danh mới. Vui lòng thử lại.")
    return { ...state, [command.kind]: [...entries, { id: newId, name }] }
  }
  return {
    ...state,
    [command.kind]: entries.map((entry) => (entry.id === command.id ? { ...entry, name } : entry)),
  }
}
