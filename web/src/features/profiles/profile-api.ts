import { AuthError, type Principal } from "../session/auth-api.ts"

export type SelfProfile = {
  user_id: number
  employee_id: number
  employee_code: string
  fullname: string
  company_email: string
  phone: string | null
  address: string | null
  position: { id: number; name: string }
  department: { id: number; name: string } | null
  job_level: { id: number; name: string } | null
  employment_status: "ACTIVE"
}
function object(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value)
}
function exact(value: Record<string, unknown>, keys: string[]) {
  return Object.keys(value).length === keys.length && keys.every((key) => key in value)
}
function text(value: unknown, min: number, max = Infinity): value is string {
  return (
    typeof value === "string" &&
    [...value].length >= min &&
    [...value].length <= max &&
    !value.includes("\0") &&
    !/[\uD800-\uDFFF]/u.test(value)
  )
}
function id(value: unknown): value is number {
  return Number.isSafeInteger(value) && (value as number) > 0 && (value as number) <= 2147483647
}
function catalog(value: unknown) {
  return object(value) && exact(value, ["id", "name"]) && id(value.id) && text(value.name, 1, 100)
}
export function parseSelfProfile(
  value: unknown,
  actor: Pick<Principal, "user_id" | "employee_id">,
): SelfProfile {
  const keys = [
    "user_id",
    "employee_id",
    "employee_code",
    "fullname",
    "company_email",
    "phone",
    "address",
    "position",
    "department",
    "job_level",
    "employment_status",
  ]
  if (
    !object(value) ||
    !exact(value, keys) ||
    !id(value.user_id) ||
    value.user_id !== actor.user_id ||
    !id(value.employee_id) ||
    value.employee_id !== actor.employee_id ||
    !text(value.employee_code, 1, 30) ||
    !text(value.fullname, 1, 100) ||
    !text(value.company_email, 3, 100) ||
    /[\r\n]/.test(value.company_email) ||
    !/^(?:"(?:[^"\\\r\n]|\\.)+"|[^@\s<>",;]+)@[^@\s<>",;]+$/u.test(value.company_email) ||
    !(value.phone === null || text(value.phone, 0, 20)) ||
    !(value.address === null || text(value.address, 0)) ||
    !catalog(value.position) ||
    !(value.department === null || catalog(value.department)) ||
    !(value.job_level === null || catalog(value.job_level)) ||
    value.employment_status !== "ACTIVE"
  )
    throw new AuthError(0, "INVALID_RESPONSE")
  return value as SelfProfile
}
export function createProfileApi(fetcher: typeof fetch = fetch) {
  return {
    async readSelf(actor: Pick<Principal, "user_id" | "employee_id">, signal?: AbortSignal) {
      if (!id(actor.user_id) || !id(actor.employee_id)) throw new AuthError(0, "INVALID_RESPONSE")
      let response: Response
      try {
        const timeout = AbortSignal.timeout(15000)
        response = await fetcher(`/api/employees/${actor.employee_id}`, {
          method: "GET",
          credentials: "same-origin",
          cache: "no-store",
          redirect: "error",
          signal: signal ? AbortSignal.any([signal, timeout]) : timeout,
          headers: { Accept: "application/json" },
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
          /* Proxy có thể trả lỗi không phải JSON. */
        }
        throw new AuthError(response.status, code)
      }
      try {
        if (
          response.status !== 200 ||
          !response.headers.get("Content-Type")?.toLowerCase().startsWith("application/json")
        )
          throw new Error()
        return parseSelfProfile(await response.json(), actor)
      } catch {
        throw new AuthError(0, "INVALID_RESPONSE")
      }
    },
  }
}
export type ProfileApi = ReturnType<typeof createProfileApi>
export function profileErrorMessage(error: unknown) {
  if (!(error instanceof AuthError)) return "Chưa tải được hồ sơ. Hãy thử lại."
  if (error.status === 401) return "Phiên không còn hiệu lực. Đang kiểm tra lại với máy chủ."
  if (error.code === "PASSWORD_CHANGE_REQUIRED")
    return "Bạn cần đổi mật khẩu. Đang kiểm tra lại phiên."
  if (error.status === 403) return "Bạn chưa được phép mở hồ sơ này. Đang kiểm tra lại phiên."
  if (error.status === 404)
    return "Không tìm thấy hồ sơ chính chủ. Hãy kiểm tra lại với người quản lý."
  if (error.status >= 500) return "Hệ thống tạm thời chưa tải được hồ sơ. Hãy thử lại sau."
  if (error.code === "NETWORK_ERROR")
    return "Không thể kết nối để tải hồ sơ. Hãy kiểm tra mạng rồi thử lại."
  return "Chưa đọc được dữ liệu hồ sơ hợp lệ. Hãy thử lại."
}
