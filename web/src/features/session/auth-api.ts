export type SystemRole = "ADMIN" | "HR" | "DIRECTOR"
export type Principal = {
  user_id: number
  employee_id: number
  employee_code: string
  fullname: string
  system_roles: SystemRole[]
  must_change_password: boolean
}

// Không giữ message/body thô từ mạng: lỗi có thể chứa dữ liệu ngoài contract.
export class AuthError extends Error {
  readonly status: number
  readonly code: string
  readonly retryAfterSeconds: number | null
  constructor(status: number, code: string, retryAfterSeconds: number | null = null) {
    super(code)
    this.name = "AuthError"
    this.status = status
    this.code = code
    this.retryAfterSeconds = retryAfterSeconds
  }
}

function object(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value)
}

export function parsePrincipal(value: unknown): Principal {
  const keys = [
    "user_id",
    "employee_id",
    "employee_code",
    "fullname",
    "system_roles",
    "must_change_password",
  ]
  if (
    !object(value) ||
    Object.keys(value).length !== keys.length ||
    keys.some((key) => !(key in value)) ||
    !Number.isSafeInteger(value.user_id) ||
    (value.user_id as number) < 1 ||
    !Number.isSafeInteger(value.employee_id) ||
    (value.employee_id as number) < 1 ||
    typeof value.employee_code !== "string" ||
    !value.employee_code.trim() ||
    typeof value.fullname !== "string" ||
    !value.fullname.trim() ||
    typeof value.must_change_password !== "boolean" ||
    !Array.isArray(value.system_roles) ||
    value.system_roles.some((role) => !["ADMIN", "HR", "DIRECTOR"].includes(role)) ||
    new Set(value.system_roles).size !== value.system_roles.length ||
    (value.must_change_password && value.system_roles.length !== 0)
  ) {
    throw new AuthError(0, "INVALID_RESPONSE")
  }
  return value as Principal
}

export type AuthApi = {
  me: () => Promise<Principal>
  login: (email: string, password: string) => Promise<Principal>
  csrf: () => Promise<string>
  logout: (csrf: string) => Promise<void>
}

export function createAuthApi(fetcher: typeof fetch = fetch): AuthApi {
  async function request(path: string, init: RequestInit = {}): Promise<Response> {
    let response: Response
    try {
      response = await fetcher(`/api/auth/${path}`, {
        ...init,
        credentials: "same-origin",
        cache: "no-store",
        redirect: "error",
        signal: AbortSignal.timeout(15000),
        headers: { Accept: "application/json", ...init.headers },
      })
    } catch {
      throw new AuthError(0, "NETWORK_ERROR")
    }
    if (!response.ok) {
      let code = "HTTP_ERROR"
      try {
        const value: unknown = await response.json()
        if (object(value) && typeof value.code === "string") code = value.code
      } catch {
        /* Lỗi proxy có thể không phải JSON. */
      }
      const raw = response.headers.get("Retry-After")
      const seconds = raw && /^\d+$/.test(raw) ? Number(raw) : NaN
      throw new AuthError(
        response.status,
        code,
        Number.isSafeInteger(seconds) && seconds >= 0 ? seconds : null,
      )
    }
    return response
  }
  async function json(response: Response): Promise<unknown> {
    if (!response.headers.get("Content-Type")?.toLowerCase().startsWith("application/json")) {
      throw new AuthError(0, "INVALID_RESPONSE")
    }
    try {
      return await response.json()
    } catch {
      throw new AuthError(0, "INVALID_RESPONSE")
    }
  }
  async function principal(response: Response): Promise<Principal> {
    if (response.status !== 200) throw new AuthError(0, "INVALID_RESPONSE")
    return parsePrincipal(await json(response))
  }
  return {
    me: async () => principal(await request("me")),
    login: async (email, password) =>
      principal(
        await request("login", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ email, password }),
        }),
      ),
    csrf: async () => {
      const response = await request("csrf")
      const value = await json(response)
      if (
        response.status !== 200 ||
        !object(value) ||
        Object.keys(value).length !== 1 ||
        typeof value.csrf_token !== "string" ||
        !/^[A-Za-z0-9_-]{43}$/.test(value.csrf_token)
      ) {
        throw new AuthError(0, "INVALID_RESPONSE")
      }
      return value.csrf_token
    },
    logout: async (csrf) => {
      const response = await request("logout", {
        method: "POST",
        headers: { "X-CSRF-Token": csrf },
      })
      if (response.status !== 204 || (await response.text()) !== "")
        throw new AuthError(0, "INVALID_RESPONSE")
    },
  }
}

export function authErrorMessage(error: AuthError): string {
  if (error.code === "INVALID_CREDENTIALS") return "Email hoặc mật khẩu không hợp lệ."
  if (error.code === "PASSWORD_CHANGE_REQUIRED") return "Bạn cần đổi mật khẩu trước khi tiếp tục."
  if (error.status === 429) return "Bạn đã thử quá nhiều lần. Vui lòng chờ trước khi thử lại."
  if (error.status === 403)
    return "Yêu cầu chưa được xác nhận hợp lệ. Hãy kiểm tra lại phiên rồi thử lại."
  if (error.status === 400) return "Thông tin gửi chưa hợp lệ. Hãy kiểm tra email và mật khẩu."
  if (error.status === 413) return "Thông tin gửi vượt kích thước cho phép."
  if (error.status === 415 || error.code === "INVALID_RESPONSE")
    return "Không thể đọc phản hồi hợp lệ từ hệ thống."
  if (error.code === "NETWORK_ERROR")
    return "Không thể kết nối. Chưa xác nhận được kết quả; hãy kiểm tra lại phiên."
  return "Hệ thống tạm thời không thể xử lý. Vui lòng thử lại sau."
}
