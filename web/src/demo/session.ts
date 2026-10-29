import type { DemoAccount } from "../models/accounts.ts"
import type { EmployeeProfile } from "../models/profile.ts"

// Phiên UI trong bộ nhớ, không phải thông tin xác thực hoặc cookie server.
export type DemoSession = { personId: string; rolesKey: string }
export type SessionIssue = "expired" | "locked" | "inactive" | "revoked"
const rolesKey = (account: DemoAccount) => [...account.systemRoles].sort().join(",")

export function sessionIssue(
  session: DemoSession,
  accounts: DemoAccount[],
  profiles: EmployeeProfile[],
): SessionIssue | null {
  const account = accounts.find((a) => a.personId === session.personId)
  if (!account || account.status !== "ACTIVE") return "locked"
  if (profiles.find((p) => p.personId === session.personId)?.employmentStatus !== "ACTIVE")
    return "inactive"
  if (rolesKey(account) !== session.rolesKey) return "revoked"
  return null
}

export function startDemoSession(
  personId: string,
  accounts: DemoAccount[],
  profiles: EmployeeProfile[],
): DemoSession {
  const account = accounts.find((a) => a.personId === personId)
  if (!account) throw new Error("Không thể mở phiên mô phỏng này.")
  const session = { personId, rolesKey: rolesKey(account) }
  const issue = sessionIssue(session, accounts, profiles)
  if (issue) throw new Error(sessionIssueMessages[issue])
  return session
}

export const sessionIssueMessages: Record<SessionIssue, string> = {
  expired: "Phiên mô phỏng đã hết hạn. Hãy đăng nhập lại để tiếp tục.",
  locked: "Tài khoản đã bị khóa. Vui lòng liên hệ quản trị viên.",
  inactive: "Hồ sơ nhân viên đã được vô hiệu hóa. Vui lòng liên hệ bộ phận nhân sự.",
  revoked: "Quyền hệ thống đã thay đổi. Hãy đăng nhập lại để cập nhật phiên.",
}

// Chỉ nhận đích nội bộ; không dùng URL do người gửi cung cấp để chuyển ra ngoài app.
export function safeReturnPath(value: string | null): string {
  if (!value || !value.startsWith("/") || value.startsWith("//") || /[\\\s]/.test(value)) return "/"
  const path = value.split(/[?#]/)[0]
  if (path === "/login") return "/"
  return value
}
