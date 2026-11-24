import { AuthError } from "../session/auth-api.ts"

export const reviewStatuses = ["PENDING", "APPROVED", "REJECTED"] as const
export type ReviewStatus = (typeof reviewStatuses)[number]
export const reviewFields = ["fullname", "phone", "address"] as const
export type ReviewField = (typeof reviewFields)[number]
export type ReviewValues = Partial<Record<ReviewField, string | null>>
export type ReviewSummary = {
  request_id: number
  employee_id: number
  employee_code: string
  fullname: string
  status: ReviewStatus
  changed_fields: ReviewField[]
  requested_at: string
  resolved_at: string | null
}
export type ReviewDetail = Omit<ReviewSummary, "changed_fields"> & {
  before_values: ReviewValues
  proposed_values: ReviewValues
  current_values: ReviewValues
  explanation: string | null
  requested_by: number
  resolved_by: number | null
  decision_reason: string | null
}
export type ReviewPage = { items: ReviewSummary[]; page: number; page_size: number; total: number }
export type ReviewQuery = { status?: ReviewStatus; page: number; pageSize: number }
export function reviewId(value: unknown): value is number {
  return Number.isSafeInteger(value) && (value as number) > 0 && (value as number) <= 2147483647
}
function object(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === "object" && !Array.isArray(value)
}
function exact(value: Record<string, unknown>, keys: string[]) {
  return Object.keys(value).length === keys.length && keys.every((key) => key in value)
}
function text(value: unknown, min = 0, max = Infinity): value is string {
  return (
    typeof value === "string" &&
    [...value].length >= min &&
    [...value].length <= max &&
    !value.includes("\0") &&
    !/[\uD800-\uDFFF]/u.test(value)
  )
}
function date(value: unknown): value is string {
  return (
    typeof value === "string" &&
    /^\d{4}-\d\d-\d\dT/.test(value) &&
    Number.isFinite(Date.parse(value))
  )
}
function base(value: Record<string, unknown>) {
  return (
    reviewId(value.request_id) &&
    reviewId(value.employee_id) &&
    text(value.employee_code, 1, 30) &&
    text(value.fullname, 1, 100) &&
    reviewStatuses.includes(value.status as ReviewStatus) &&
    date(value.requested_at) &&
    (value.status === "PENDING" ? value.resolved_at === null : date(value.resolved_at))
  )
}
function values(value: unknown, proposed = false): value is ReviewValues {
  if (!object(value) || !Object.keys(value).length) return false
  return Object.entries(value).every(
    ([key, entry]) =>
      reviewFields.includes(key as ReviewField) &&
      (entry === null
        ? key === "phone" || (!proposed && key === "address")
        : text(
            entry,
            key === "fullname" || proposed ? 1 : 0,
            proposed ? { fullname: 100, phone: 20, address: 2000 }[key as ReviewField] : Infinity,
          )),
  )
}
function invalid(): never {
  throw new AuthError(0, "INVALID_RESPONSE")
}
export function parseReviewDetail(value: unknown, id: number): ReviewDetail {
  if (
    !object(value) ||
    !exact(value, [
      "request_id",
      "employee_id",
      "employee_code",
      "fullname",
      "status",
      "before_values",
      "proposed_values",
      "current_values",
      "explanation",
      "requested_by",
      "requested_at",
      "resolved_by",
      "resolved_at",
      "decision_reason",
    ]) ||
    !base(value) ||
    value.request_id !== id ||
    !reviewId(value.requested_by) ||
    !values(value.before_values) ||
    !values(value.current_values) ||
    !values(value.proposed_values, true) ||
    !(value.explanation === null || text(value.explanation, 0, 2000))
  )
    invalid()
  const keys = Object.keys(value.before_values as object)
    .sort()
    .join(",")
  if (
    [value.proposed_values, value.current_values].some(
      (entry) =>
        Object.keys(entry as object)
          .sort()
          .join(",") !== keys,
    )
  )
    invalid()
  if (
    value.status === "PENDING"
      ? value.resolved_by !== null || value.decision_reason !== null
      : !reviewId(value.resolved_by) ||
        (value.status === "APPROVED"
          ? value.decision_reason !== null
          : !text(value.decision_reason, 1, 2000))
  )
    invalid()
  return value as ReviewDetail
}
export function parseReviewPage(value: unknown, query: ReviewQuery): ReviewPage {
  if (
    !object(value) ||
    !exact(value, ["items", "page", "page_size", "total"]) ||
    value.page !== query.page ||
    value.page_size !== query.pageSize ||
    !Number.isSafeInteger(value.total) ||
    (value.total as number) < 0 ||
    !Array.isArray(value.items) ||
    value.items.length > query.pageSize
  )
    invalid()
  const items = value.items as unknown[]
  const seen = new Set<number>()
  for (const entry of items) {
    if (
      !object(entry) ||
      !exact(entry, [
        "request_id",
        "employee_id",
        "employee_code",
        "fullname",
        "status",
        "changed_fields",
        "requested_at",
        "resolved_at",
      ]) ||
      !base(entry) ||
      (query.status && entry.status !== query.status) ||
      !Array.isArray(entry.changed_fields) ||
      entry.changed_fields.length < 1 ||
      entry.changed_fields.length > 3 ||
      new Set(entry.changed_fields).size !== entry.changed_fields.length ||
      entry.changed_fields.some((field) => !reviewFields.includes(field as ReviewField)) ||
      seen.has(entry.request_id as number)
    )
      invalid()
    seen.add(entry.request_id as number)
  }
  return value as ReviewPage
}
export function trimReviewReason(value: string) {
  // Unicode White_Space trùng Go strings.TrimSpace; FEFF không là khoảng trắng của Go.
  return value.replace(/^\p{White_Space}+|\p{White_Space}+$/gu, "")
}
export function validReviewReason(value: string) {
  return text(trimReviewReason(value), 1, 2000)
}
export function createHrReviewApi(fetcher: typeof fetch = fetch) {
  async function request(path: string, signal?: AbortSignal, body?: object, token?: string) {
    let response: Response
    try {
      response = await fetcher(`/api/profile-change-requests${path}`, {
        method: body ? "POST" : "GET",
        credentials: "same-origin",
        cache: "no-store",
        redirect: "error",
        signal: signal
          ? AbortSignal.any([signal, AbortSignal.timeout(15000)])
          : AbortSignal.timeout(15000),
        headers: {
          Accept: "application/json",
          ...(body ? { "Content-Type": "application/json", "X-CSRF-Token": token ?? "" } : {}),
        },
        ...(body ? { body: JSON.stringify(body) } : {}),
      })
    } catch {
      throw new AuthError(0, "NETWORK_ERROR")
    }
    if (!response.ok) {
      let code = "HTTP_ERROR"
      try {
        const error: unknown = await response.json()
        if (object(error) && typeof error.code === "string") code = error.code
      } catch {
        /* Lỗi proxy không nhất thiết là JSON. */
      }
      throw new AuthError(response.status, code)
    }
    try {
      if (
        response.status !== 200 ||
        !response.headers.get("Content-Type")?.toLowerCase().startsWith("application/json")
      )
        invalid()
      return (await response.json()) as unknown
    } catch {
      return invalid()
    }
  }
  return {
    async list(query: ReviewQuery, signal?: AbortSignal) {
      if (
        !reviewId(query.page) ||
        !reviewId(query.pageSize) ||
        query.pageSize > 100 ||
        (query.status && !reviewStatuses.includes(query.status))
      )
        throw new AuthError(400, "INVALID_QUERY")
      const params = new URLSearchParams({
        scope: "review",
        page: String(query.page),
        page_size: String(query.pageSize),
      })
      if (query.status) params.set("status", query.status)
      return parseReviewPage(await request(`?${params}`, signal), query)
    },
    async detail(id: number, signal?: AbortSignal) {
      if (!reviewId(id)) throw new AuthError(400, "INVALID_ID")
      return parseReviewDetail(await request(`/${id}`, signal), id)
    },
    async decide(id: number, decision: "approve" | "reject", reason: string, token: string) {
      if (!reviewId(id) || (decision === "reject" && !validReviewReason(reason)))
        throw new AuthError(400, "INVALID_REASON")
      const detail = parseReviewDetail(
        await request(
          `/${id}/${decision}`,
          undefined,
          decision === "approve" ? {} : { reason: trimReviewReason(reason) },
          token,
        ),
        id,
      )
      if (detail.status !== (decision === "approve" ? "APPROVED" : "REJECTED")) invalid()
      return detail
    },
  }
}
export type HrReviewApi = ReturnType<typeof createHrReviewApi>
export function reviewErrorMessage(error: unknown) {
  if (!(error instanceof AuthError)) return "Chưa tải được đề nghị. Hãy thử tải lại."
  if (error.code === "STALE_DATA")
    return "Dữ liệu nền đã thay đổi. Hãy tải lại để đối chiếu; hệ thống không ghi đè dữ liệu hiện tại."
  if (error.code === "REQUEST_ALREADY_RESOLVED")
    return "Đề nghị đã được xử lý. Hãy tải lại để xem kết luận."
  if (error.code === "EMPLOYEE_INACTIVE")
    return "Hồ sơ nhân viên đã ngừng hoạt động nên không thể chấp thuận. Bạn vẫn có thể từ chối đề nghị đang chờ."
  if (error.status === 401) return "Phiên không còn hiệu lực. Đang kiểm tra lại với máy chủ."
  if (error.status === 403) return "Bạn không còn được phép xử lý đề nghị. Đang kiểm tra lại quyền."
  if (error.status === 404) return "Không tìm thấy đề nghị hoặc bạn không được phép xem."
  if (error.status === 400) return "Nội dung hoặc đường dẫn chưa hợp lệ. Hãy kiểm tra lại."
  if (error.status === 503 || error.status >= 500)
    return "Dịch vụ chưa sẵn sàng. Hãy tải lại để kiểm tra."
  return "Chưa kết nối được hoặc phản hồi chưa hợp lệ. Hãy tải lại để kiểm tra."
}
