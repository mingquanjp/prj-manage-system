import type { SystemRole } from "./workspace.ts"

// Dữ liệu quản trị trong phiên demo; không chứa thông tin xác thực hay phiên đăng nhập.
export type DemoAccount = {
  id: string
  personId: string
  username: string
  status: "ACTIVE" | "LOCKED"
  systemRoles: SystemRole[]
}
export type AccountEvent = {
  id: string
  accountId: string
  actorId: string
  at: string
  action: "CREATED" | "ROLES_CHANGED" | "LOCKED" | "UNLOCKED"
  before?: Pick<DemoAccount, "status" | "systemRoles">
  after: Pick<DemoAccount, "status" | "systemRoles">
}
export type AccountsState = { accounts: DemoAccount[]; events: AccountEvent[] }
export type AccountCreateInput = {
  fullname: string
  companyEmail: string
  positionId: string
  departmentId: string | null
  jobLevelId: string | null
  phone: string | null
  systemRoles: SystemRole[]
}
export type AccountCommand = {
  accountId: string
  before: Pick<DemoAccount, "status" | "systemRoles">
} & ({ action: "roles"; roles: SystemRole[] } | { action: "lock" | "unlock" })
export const systemRoleLabels: Record<SystemRole, string> = {
  ADMIN: "Quản trị viên",
  HR: "Quản lý nhân sự",
  DIRECTOR: "Giám đốc",
}
