import { AuthError } from "../session/auth-api.ts"
import {
  requestFields,
  trimProfileText,
  validProfileText,
  validateProfileRequestInput,
  type ProfileRequestInput,
  type ProfileValues,
  type RequestField,
} from "./profile-request-input.ts"

export type RequestStatus = "PENDING" | "APPROVED" | "REJECTED"
export type RequestSummary = {
  request_id: number
  employee_id: number
  employee_code: string
  fullname: string
  status: RequestStatus
  changed_fields: RequestField[]
  requested_at: string
  resolved_at: string | null
}
export type RequestDetail = Omit<RequestSummary, "changed_fields"> & {
  before_values: Partial<ProfileValues>
  proposed_values: Partial<ProfileValues>
  current_values: Partial<ProfileValues>
  explanation: string | null
  requested_by: number
  resolved_by: number | null
  decision_reason: string | null
}
export type RequestPage = {
  items: RequestSummary[]
  page: number
  page_size: number
  total: number
}
const summaryKeys = [
  "request_id",
  "employee_id",
  "employee_code",
  "fullname",
  "status",
  "changed_fields",
  "requested_at",
  "resolved_at",
]
const detailKeys = [
  ...summaryKeys.filter((key) => key !== "changed_fields"),
  "before_values",
  "proposed_values",
  "current_values",
  "explanation",
  "requested_by",
  "resolved_by",
  "decision_reason",
]
function object(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value)
}
function exact(value: Record<string, unknown>, keys: string[]) {
  return Object.keys(value).length === keys.length && keys.every((key) => Object.hasOwn(value, key))
}
function id(value: unknown): value is number {
  return Number.isInteger(value) && (value as number) > 0 && (value as number) <= 2147483647
}
function text(value: unknown, min: number, max = Infinity) {
  return validProfileText(value) && [...value].length >= min && [...value].length <= max
}
function date(value: unknown): value is string {
  return (
    typeof value === "string" &&
    /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})$/.test(value) &&
    Number.isFinite(Date.parse(value))
  )
}
function base(value: Record<string, unknown>, employeeId: number) {
  return (
    id(value.request_id) &&
    value.employee_id === employeeId &&
    text(value.employee_code, 1, 30) &&
    text(value.fullname, 1, 100) &&
    ["PENDING", "APPROVED", "REJECTED"].includes(value.status as string) &&
    date(value.requested_at) &&
    (value.status === "PENDING"
      ? value.resolved_at === null
      : date(value.resolved_at) && Date.parse(value.resolved_at) >= Date.parse(value.requested_at))
  )
}
function values(value: unknown, proposed: boolean): value is Partial<ProfileValues> {
  if (
    !object(value) ||
    !Object.keys(value).length ||
    Object.keys(value).some((key) => !requestFields.includes(key as RequestField))
  )
    return false
  return Object.entries(value).every(([key, item]) => {
    if (item === null) return key !== "fullname" && (!proposed || key === "phone")
    if (!text(item, key === "fullname" || proposed ? 1 : 0)) return false
    return (
      !proposed ||
      (typeof item === "string" &&
        trimProfileText(item) === item &&
        [...item].length <= { fullname: 100, phone: 20, address: 2000 }[key as RequestField])
    )
  })
}
function invalid(): never {
  throw new AuthError(0, "INVALID_RESPONSE")
}
export function parseRequestDetail(
  value: unknown,
  employeeId: number,
  requestId?: number,
): RequestDetail {
  if (
    !object(value) ||
    !exact(value, detailKeys) ||
    !base(value, employeeId) ||
    (requestId !== undefined && value.request_id !== requestId) ||
    !id(value.requested_by) ||
    !values(value.before_values, false) ||
    !values(value.proposed_values, true) ||
    !values(value.current_values, false) ||
    !(value.explanation === null || text(value.explanation, 0, 2000))
  )
    invalid()
  const before = value.before_values as Partial<ProfileValues>
  const proposed = value.proposed_values as Partial<ProfileValues>
  if (
    Object.keys(before).sort().join() !== Object.keys(proposed).sort().join() ||
    Object.keys(before).sort().join() !==
      Object.keys(value.current_values as object)
        .sort()
        .join() ||
    Object.keys(before).every(
      (key) => before[key as RequestField] === proposed[key as RequestField],
    )
  )
    invalid()
  if (value.status === "PENDING") {
    if (value.resolved_by !== null || value.decision_reason !== null) invalid()
  } else {
    if (!id(value.resolved_by)) invalid()
    if (value.status === "APPROVED") {
      if (value.decision_reason !== null) invalid()
    } else if (
      !text(value.decision_reason, 1, 2000) ||
      trimProfileText(value.decision_reason as string) === ""
    ) {
      invalid()
    }
  }
  return value as unknown as RequestDetail
}
export function parseRequestPage(value: unknown, employeeId: number, page: number): RequestPage {
  if (
    !object(value) ||
    !exact(value, ["items", "page", "page_size", "total"]) ||
    value.page !== page ||
    value.page_size !== 20 ||
    !Number.isSafeInteger(value.total) ||
    (value.total as number) < 0 ||
    !Array.isArray(value.items) ||
    value.items.length > 20 ||
    value.items.length > (value.total as number)
  )
    invalid()
  const seen = new Set<number>()
  for (const item of value.items) {
    if (
      !object(item) ||
      !exact(item, summaryKeys) ||
      !base(item, employeeId) ||
      !Array.isArray(item.changed_fields) ||
      item.changed_fields.length < 1 ||
      item.changed_fields.length > 3 ||
      item.changed_fields.some((field) => !requestFields.includes(field as RequestField)) ||
      new Set(item.changed_fields).size !== item.changed_fields.length ||
      seen.has(item.request_id as number)
    )
      invalid()
    seen.add(item.request_id as number)
  }
  return value as unknown as RequestPage
}
const errorCodes = new Set([
  "VALIDATION_ERROR",
  "UNAUTHENTICATED",
  "FORBIDDEN",
  "PASSWORD_CHANGE_REQUIRED",
  "NOT_FOUND",
  "STALE_DATA",
  "IDEMPOTENCY_KEY_REUSED",
  "IDEMPOTENCY_KEY_EXPIRED",
  "PAYLOAD_TOO_LARGE",
  "UNSUPPORTED_MEDIA_TYPE",
  "INTERNAL_ERROR",
  "SERVICE_UNAVAILABLE",
])
export function createProfileRequestsApi(fetcher: typeof fetch = fetch) {
  async function request(
    url: string,
    status: number,
    signal?: AbortSignal,
    mutation?: { input: ProfileRequestInput; key: string; csrf: string },
  ) {
    let response: Response
    try {
      const timeout = AbortSignal.timeout(15000)
      response = await fetcher(url, {
        method: mutation ? "POST" : "GET",
        credentials: "same-origin",
        cache: "no-store",
        redirect: "error",
        signal: signal ? AbortSignal.any([signal, timeout]) : timeout,
        headers: mutation
          ? {
              Accept: "application/json",
              "Content-Type": "application/json",
              "X-CSRF-Token": mutation.csrf,
              "Idempotency-Key": mutation.key,
            }
          : { Accept: "application/json" },
        ...(mutation ? { body: JSON.stringify(mutation.input) } : {}),
      })
    } catch {
      throw new AuthError(0, "NETWORK_ERROR")
    }
    if (!response.ok) {
      let code = "HTTP_ERROR"
      try {
        const body: unknown = await response.json()
        if (object(body) && typeof body.code === "string" && errorCodes.has(body.code))
          code = body.code
      } catch {
        /* Proxy không nhất thiết trả JSON. */
      }
      throw new AuthError(response.status, code)
    }
    try {
      if (
        response.status !== status ||
        !response.headers.get("Content-Type")?.toLowerCase().startsWith("application/json")
      )
        invalid()
      return (await response.json()) as unknown
    } catch {
      return invalid()
    }
  }
  return {
    async list(employeeId: number, page = 1, signal?: AbortSignal) {
      if (!id(employeeId) || !id(page)) invalid()
      return parseRequestPage(
        await request(
          `/api/profile-change-requests?scope=self&page=${page}&page_size=20`,
          200,
          signal,
        ),
        employeeId,
        page,
      )
    },
    async read(employeeId: number, requestId: number, signal?: AbortSignal) {
      if (!id(employeeId) || !id(requestId)) invalid()
      return parseRequestDetail(
        await request(`/api/profile-change-requests/${requestId}`, 200, signal),
        employeeId,
        requestId,
      )
    },
    async submit(
      employeeId: number,
      input: ProfileRequestInput,
      key: string,
      csrf: string,
      signal?: AbortSignal,
    ) {
      if (
        !id(employeeId) ||
        !/^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(key) ||
        !csrf
      )
        throw new AuthError(400, "VALIDATION_ERROR")
      validateProfileRequestInput(input)
      return parseRequestDetail(
        await request(`/api/employees/${employeeId}/profile-change-requests`, 201, signal, {
          input,
          key: key.toLowerCase(),
          csrf,
        }),
        employeeId,
      )
    },
  }
}
export type ProfileRequestsApi = ReturnType<typeof createProfileRequestsApi>
export function requestErrorMessage(error: unknown) {
  if (!(error instanceof AuthError)) return "Chưa thể xử lý đề nghị. Hãy kiểm tra nội dung."
  if (error.status === 401) return "Phiên không còn hiệu lực. Vui lòng đăng nhập lại."
  if (error.code === "PASSWORD_CHANGE_REQUIRED") return "Bạn cần đổi mật khẩu trước khi tiếp tục."
  if (error.status === 403)
    return "Bạn không còn quyền thực hiện thao tác. Đang kiểm tra lại phiên."
  if (error.status === 404) return "Không tìm thấy đề nghị thuộc hồ sơ của bạn."
  if (error.code === "STALE_DATA")
    return "Thông tin nền đã thay đổi. Hãy tải lại hồ sơ và xem lại đề nghị."
  if (error.code === "IDEMPOTENCY_KEY_REUSED" || error.code === "IDEMPOTENCY_KEY_EXPIRED")
    return "Khóa gửi không còn hợp lệ. Hãy xem lịch sử để đối soát đề nghị."
  if (error.status === 413) return "Nội dung gửi vượt giới hạn 64 KiB."
  if (error.status >= 500) return "Dịch vụ tạm thời không sẵn sàng. Hãy thử tải lại sau."
  if (error.status === 0) return "Chưa kết nối hoặc chưa đọc được phản hồi hợp lệ từ máy chủ."
  return "Nội dung đề nghị chưa hợp lệ. Hãy kiểm tra lại các trường nhập."
}
