import { AuthError, type SystemRole } from "../session/auth-api.ts"
import { trimCatalogName } from "../organization/catalog-api.ts"

export type Account = {
  user_id: number
  employee_id: number
  employee_code: string
  fullname: string
  company_email: string
  status: "ACTIVE" | "LOCKED" | "DISABLED"
  system_roles: SystemRole[]
  must_change_password: boolean
}
export type AccountPage = {
  items: Account[]
  pagination: { page: number; page_size: number; total: number }
}
export type AccountCreated = {
  account_created: true
  account: Account
  mail: { status: "accepted" | "failed" | "unknown" }
}
export type AccountDraft = {
  fullname: string
  companyEmail: string
  positionId: string
  departmentId: string | null
  jobLevelId: string | null
  phone: string | null
  address: string | null
  systemRoles: SystemRole[]
}
export const emptyAccountDraft: AccountDraft = {
  fullname: "",
  companyEmail: "",
  positionId: "",
  departmentId: null,
  jobLevelId: null,
  phone: null,
  address: null,
  systemRoles: [],
}
const object = (v: unknown): v is Record<string, unknown> =>
  typeof v === "object" && v !== null && !Array.isArray(v)
const exact = (v: Record<string, unknown>, keys: string[]) =>
  Object.keys(v).length === keys.length && keys.every((k) => k in v)
const id = (v: unknown): v is number =>
  Number.isInteger(v) && (v as number) > 0 && (v as number) <= 2147483647
const validText = (v: unknown, max = Infinity): v is string =>
  typeof v === "string" &&
  !!trimCatalogName(v) &&
  [...v].length <= max &&
  !/[\0\uD800-\uDFFF]/u.test(v)
const roles = (v: unknown): v is SystemRole[] =>
  Array.isArray(v) &&
  v.every((r) => ["ADMIN", "HR", "DIRECTOR"].includes(r)) &&
  new Set(v).size === v.length
function account(v: unknown): Account {
  if (
    !object(v) ||
    !exact(v, [
      "user_id",
      "employee_id",
      "employee_code",
      "fullname",
      "company_email",
      "status",
      "system_roles",
      "must_change_password",
    ]) ||
    !id(v.user_id) ||
    !id(v.employee_id) ||
    !validText(v.employee_code, 30) ||
    !validText(v.fullname, 100) ||
    !validText(v.company_email, 100) ||
    !["ACTIVE", "LOCKED", "DISABLED"].includes(String(v.status)) ||
    !roles(v.system_roles) ||
    typeof v.must_change_password !== "boolean"
  )
    throw new AuthError(0, "INVALID_RESPONSE")
  return v as Account
}
export function accountDraftErrors(v: AccountDraft) {
  const errors: Record<string, string> = {}
  if (!validText(trimCatalogName(v.fullname), 100))
    errors.fullname = "Nhập họ tên từ 1 đến 100 ký tự."
  const email = trimCatalogName(v.companyEmail)
  // Chỉ kiểm tra sơ bộ ở FE; cú pháp email (kể cả local-part có quote)
  // và chuẩn hóa cuối cùng do net/mail cùng helper auth phía server quyết định.
  if (!validText(email, 100) || !email.includes("@") || /[\r\n]/.test(v.companyEmail))
    errors.companyEmail = "Nhập email đơn hợp lệ, tối đa 100 ký tự."
  for (const key of ["positionId", "departmentId", "jobLevelId"] as const) {
    const value = v[key]
    if (
      (key === "positionId" || value !== null) &&
      (!value || !/^[1-9]\d*$/.test(value) || !id(Number(value)))
    )
      errors[key] = key === "positionId" ? "Chọn vị trí công việc." : "Lựa chọn chưa hợp lệ."
  }
  if (v.phone !== null && !validText(trimCatalogName(v.phone), 20))
    errors.phone = "Số điện thoại tối đa 20 ký tự, không chỉ khoảng trắng."
  if (v.address !== null && !validText(trimCatalogName(v.address)))
    errors.address = "Địa chỉ không được chỉ có khoảng trắng."
  if (!roles(v.systemRoles)) errors.systemRoles = "Quyền hệ thống chưa hợp lệ."
  return errors
}
export function accountPayload(v: AccountDraft) {
  return {
    fullname: v.fullname,
    company_email: v.companyEmail,
    position_id: Number(v.positionId),
    department_id: v.departmentId === null ? null : Number(v.departmentId),
    job_level_id: v.jobLevelId === null ? null : Number(v.jobLevelId),
    phone: v.phone,
    address: v.address,
    system_roles: [...v.systemRoles],
  }
}
export function createAccountApi(fetcher: typeof fetch = fetch) {
  async function request(path: string, status: number, init: RequestInit = {}) {
    let response: Response
    try {
      response = await fetcher(`/api/${path}`, {
        ...init,
        credentials: "same-origin",
        cache: "no-store",
        redirect: "error",
        // Create chờ DB rồi mail; timeout FE không rút ngắn cửa sổ 30s của server.
        signal: AbortSignal.timeout(init.method === "POST" ? 35000 : 15000),
        headers: { Accept: "application/json", ...init.headers },
      })
    } catch {
      throw new AuthError(0, "NETWORK_ERROR")
    }
    if (!response.ok) {
      let code = "HTTP_ERROR"
      try {
        const body: unknown = await response.json()
        if (object(body) && typeof body.code === "string") code = body.code
      } catch {
        /* Proxy có thể trả HTML. */
      }
      throw new AuthError(response.status, code)
    }
    if (response.status !== status) throw new AuthError(0, "INVALID_RESPONSE")
    try {
      if (!response.headers.get("Content-Type")?.toLowerCase().startsWith("application/json"))
        throw new Error()
      return (await response.json()) as unknown
    } catch {
      throw new AuthError(0, "INVALID_RESPONSE")
    }
  }
  return {
    async preview(): Promise<string> {
      const v = await request("employees/next-code", 200)
      if (
        !object(v) ||
        !exact(v, ["employee_code", "is_preview"]) ||
        v.is_preview !== true ||
        typeof v.employee_code !== "string" ||
        !/^EMP\d{6,}$/.test(v.employee_code) ||
        v.employee_code.length > 30
      )
        throw new AuthError(0, "INVALID_RESPONSE")
      return v.employee_code
    },
    async list(q: string, page: number, pageSize: number): Promise<AccountPage> {
      const params = new URLSearchParams({ q, page: String(page), page_size: String(pageSize) })
      const v = await request(`users?${params}`, 200)
      if (
        !object(v) ||
        !exact(v, ["items", "pagination"]) ||
        !Array.isArray(v.items) ||
        !object(v.pagination) ||
        !exact(v.pagination, ["page", "page_size", "total"])
      )
        throw new AuthError(0, "INVALID_RESPONSE")
      const p = v.pagination
      if (
        p.page !== page ||
        p.page_size !== pageSize ||
        !Number.isSafeInteger(p.total) ||
        (p.total as number) < 0 ||
        v.items.length > pageSize ||
        v.items.length > (p.total as number)
      )
        throw new AuthError(0, "INVALID_RESPONSE")
      const items = v.items.map(account)
      if (new Set(items.map((a) => a.user_id)).size !== items.length)
        throw new AuthError(0, "INVALID_RESPONSE")
      return { items, pagination: p as AccountPage["pagination"] }
    },
    async create(draft: AccountDraft, token: string): Promise<AccountCreated> {
      const v = await request("employees", 201, {
        method: "POST",
        headers: { "Content-Type": "application/json", "X-CSRF-Token": token },
        body: JSON.stringify(accountPayload(draft)),
      })
      if (
        !object(v) ||
        !exact(v, ["account_created", "account", "mail"]) ||
        v.account_created !== true ||
        !object(v.mail) ||
        !exact(v.mail, ["status"]) ||
        !["accepted", "failed", "unknown"].includes(String(v.mail.status))
      )
        throw new AuthError(0, "INVALID_RESPONSE")
      const created = account(v.account)
      if (created.status !== "ACTIVE" || !created.must_change_password)
        throw new AuthError(0, "INVALID_RESPONSE")
      return { account_created: true, account: created, mail: v.mail as AccountCreated["mail"] }
    },
  }
}
export const mailMessages: Record<AccountCreated["mail"]["status"], string> = {
  accepted: "Google đã nhận yêu cầu gửi email. Chưa xác nhận thư tới hộp thư đến.",
  failed: "Gửi email thất bại. Tài khoản vẫn đã được tạo; hiện chưa hỗ trợ gửi lại.",
  unknown: "Chưa xác định được kết quả gửi email. Thư có thể đã được gửi; không tự gửi lại.",
}
export function accountErrorMessage(error: unknown, posted = false) {
  if (posted && (!(error instanceof AuthError) || error.status === 0 || error.status >= 500))
    return "Chưa xác nhận được kết quả tạo tài khoản. Hãy tìm email trong danh sách để đối soát; không tạo lại hoặc gửi lại yêu cầu này. Danh sách không cho biết kết quả gửi email."
  if (error instanceof AuthError) {
    if (error.code === "EMAIL_CONFLICT")
      return "Email đã được sử dụng. Hãy kiểm tra tài khoản trong danh sách."
    if (error.status === 401) return "Phiên không còn hiệu lực. Vui lòng đăng nhập lại."
    if (error.status === 403) return "Quyền hoặc phiên đã thay đổi. Đang kiểm tra lại với máy chủ."
    if (error.status === 400)
      return "Dữ liệu chưa hợp lệ hoặc danh mục đã thay đổi. Kiểm tra form và tải lại các lựa chọn."
    if (error.status === 413)
      return "Nội dung vượt giới hạn 8192 byte. Hãy rút ngắn địa chỉ hoặc các trường nhập."
    if (error.code === "REQUEST_BUSY")
      return "Phiên đang được kiểm tra. Vui lòng chờ rồi thao tác lại."
  }
  return "Chưa thể tải dữ liệu từ hệ thống. Hãy kiểm tra kết nối rồi tải lại."
}
