import type {
  AccountsState,
  AccountCreateInput,
  AccountCommand,
  DemoAccount,
} from "../models/accounts.ts"
import { systemRoleLabels } from "../models/accounts.ts"
import type { ProfileState } from "../models/profile.ts"
import type { OrganizationState } from "../models/organization.ts"
import type { Project, Task, SystemRole } from "../models/workspace.ts"
import { demoUsers } from "./fixtures.ts"
import { isAdmin } from "./identity.ts"
import { professionalPeople } from "./profiles.ts"
import { employeeEditErrors, normalizeEmployeeEdit } from "./employee-administration.ts"

export function createAccountsState(): AccountsState {
  return {
    accounts: demoUsers.map((u) => ({
      ...u,
      id: `account-${u.personId}`,
      username: u.personId,
      status: "ACTIVE",
      systemRoles: [...u.systemRoles],
    })),
    events: [],
  }
}
// Mã dự kiến và mã lúc lưu dùng cùng nguồn. Không phải bộ cấp số đồng thời của server.
// Username theo mã chỉ tạm cho demo; chính sách đăng nhập backend chưa chốt.
export function nextAccountIdentifiers(accounts: AccountsState, profiles: ProfileState) {
  let latest = 0n
  for (const value of [
    ...profiles.profiles.map((p) => p.employeeCode),
    ...accounts.accounts.map((a) => a.username),
  ]) {
    const match = /^e(\d+)$/i.exec(value)
    if (match && BigInt(match[1]) > latest) latest = BigInt(match[1])
  }
  const suffix = String(latest + 1n).padStart(3, "0")
  if (suffix.length > 29) throw new Error("Chưa thể cấp mã mới. Cần kiểm tra lại dải mã nhân viên.")
  return { employeeCode: `E${suffix}`, username: `e${suffix}` }
}
export function normalizeRoles(roles: SystemRole[]) {
  if (!Array.isArray(roles) || roles.some((r) => !Object.hasOwn(systemRoleLabels, r)))
    throw new Error("Quyền hệ thống chỉ gồm Quản trị viên, Quản lý nhân sự và Giám đốc.")
  return (Object.keys(systemRoleLabels) as SystemRole[]).filter((r) => roles.includes(r))
}
export function activeAdminCount(state: AccountsState, profiles: ProfileState["profiles"]) {
  return state.accounts.filter(
    (a) =>
      a.status === "ACTIVE" &&
      a.systemRoles.includes("ADMIN") &&
      profiles.some((p) => p.personId === a.personId && p.employmentStatus === "ACTIVE"),
  ).length
}
function requireAdmin(
  state: AccountsState,
  profiles: ProfileState,
  organization: OrganizationState,
  actorId: string,
) {
  const identity = {
    accounts: state.accounts,
    people: professionalPeople(profiles.profiles, organization, undefined, state.accounts),
  }
  if (!isAdmin(actorId, identity))
    throw new Error("Chỉ Quản trị viên đang hoạt động được quản lý tài khoản.")
}
export function accountCreateErrors(
  profiles: ProfileState,
  organization: OrganizationState,
  input: AccountCreateInput,
) {
  const provisional = {
    personId: "",
    employeeCode: "",
    employmentStatus: "ACTIVE" as const,
    fullname: "",
    companyEmail: "",
    phone: null,
    address: null,
    departmentId: null,
    positionId: "",
    jobLevelId: null,
  }
  const values = normalizeEmployeeEdit({ ...input, address: null })
  const errors: Record<string, string> = {
    ...employeeEditErrors(profiles.profiles, organization, provisional, values),
  }
  if (!values.fullname || values.fullname.length > 100)
    errors.fullname = "Nhập họ tên từ 1 đến 100 ký tự."
  try {
    normalizeRoles(input.systemRoles)
  } catch (err) {
    errors.systemRoles = (err as Error).message
  }
  return errors
}
export function createAccountWithProfile(
  accounts: AccountsState,
  profileState: ProfileState,
  organization: OrganizationState,
  actorId: string,
  input: AccountCreateInput,
  personId: string,
  accountId: string,
  eventId: string,
  at: string,
) {
  requireAdmin(accounts, profileState, organization, actorId)
  const errors = accountCreateErrors(profileState, organization, input)
  if (Object.keys(errors).length) throw new Error(Object.values(errors).join(" "))
  if (
    profileState.profiles.some((p) => p.personId === personId) ||
    accounts.accounts.some((a) => a.id === accountId)
  )
    throw new Error("Định danh đã tồn tại. Hãy thử tạo lại.")
  const identifiers = nextAccountIdentifiers(accounts, profileState)
  const values = normalizeEmployeeEdit({ ...input, address: null })
  const account: DemoAccount = {
    id: accountId,
    personId,
    username: identifiers.username,
    status: "ACTIVE",
    systemRoles: normalizeRoles(input.systemRoles),
  }
  return {
    accounts: {
      accounts: [...accounts.accounts, account],
      events: [
        ...accounts.events,
        {
          id: eventId,
          accountId,
          actorId,
          at,
          action: "CREATED" as const,
          after: { status: account.status, systemRoles: [...account.systemRoles] },
        },
      ],
    },
    profileState: {
      ...profileState,
      profiles: [
        ...profileState.profiles,
        {
          ...values,
          personId,
          employeeCode: identifiers.employeeCode,
          employmentStatus: "ACTIVE" as const,
        },
      ],
    },
  }
}
export function accountOpenWork(personId: string, projects: Project[], tasks: Task[]) {
  return [
    ...tasks
      .filter(
        (t) =>
          t.assigneeId === personId &&
          ["ASSIGNED", "IN_PROGRESS", "PAUSED", "PENDING_REVIEW"].includes(t.status),
      )
      .map((t) => `Công việc ${t.code}: cần PTCM xem xét phân công.`),
    ...projects
      .filter(
        (p) =>
          ["PLANNING", "IN_PROGRESS", "COMPLETED"].includes(p.status) &&
          (p.pmId === personId || p.ptcmId === personId),
      )
      .map(
        (p) =>
          `Dự án ${p.code}: còn trách nhiệm ${[p.pmId === personId ? "PM" : "", p.ptcmId === personId ? "PTCM" : ""].filter(Boolean).join(" + ")}; cần người có thẩm quyền xử lý.`,
      ),
  ]
}
export function changeAccount(
  state: AccountsState,
  profiles: ProfileState,
  organization: OrganizationState,
  actorId: string,
  command: AccountCommand,
  eventId: string,
  at: string,
): AccountsState {
  requireAdmin(state, profiles, organization, actorId)
  const current = state.accounts.find((a) => a.id === command.accountId)
  if (!current) throw new Error("Không tìm thấy tài khoản.")
  if (
    current.status !== command.before.status ||
    JSON.stringify(normalizeRoles(current.systemRoles)) !==
      JSON.stringify(normalizeRoles(command.before.systemRoles))
  )
    throw new Error("Tài khoản đã thay đổi. Mở lại trước khi lưu.")
  const next = { ...current, systemRoles: [...current.systemRoles] }
  if (command.action === "roles") next.systemRoles = normalizeRoles(command.roles)
  else if (command.action === "lock") next.status = "LOCKED"
  else if (command.action === "unlock") next.status = "ACTIVE"
  else throw new Error("Thao tác tài khoản không hợp lệ.")
  const isLastAdmin =
    current.status === "ACTIVE" &&
    current.systemRoles.includes("ADMIN") &&
    profiles.profiles.some(
      (p) => p.personId === current.personId && p.employmentStatus === "ACTIVE",
    ) &&
    activeAdminCount(state, profiles.profiles) === 1
  if (isLastAdmin && (next.status !== "ACTIVE" || !next.systemRoles.includes("ADMIN")))
    throw new Error("Không được khóa hoặc thu hồi quyền Quản trị viên hoạt động cuối cùng.")
  if (
    next.status === current.status &&
    JSON.stringify(next.systemRoles) === JSON.stringify(normalizeRoles(current.systemRoles))
  )
    throw new Error("Chưa có thay đổi tài khoản.")
  return {
    accounts: state.accounts.map((a) => (a.id === next.id ? next : a)),
    events: [
      ...state.events,
      {
        id: eventId,
        accountId: current.id,
        actorId,
        at,
        action:
          command.action === "roles"
            ? "ROLES_CHANGED"
            : command.action === "lock"
              ? "LOCKED"
              : "UNLOCKED",
        before: { status: current.status, systemRoles: [...current.systemRoles] },
        after: { status: next.status, systemRoles: [...next.systemRoles] },
      },
    ],
  }
}
