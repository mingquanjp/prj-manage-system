import { AuthError } from "../session/auth-api.ts"

export type NotificationFilter = "all" | "unread"
const notificationTypes = [
  "PROFILE_CHANGE_SUBMITTED",
  "PROFILE_CHANGE_APPROVED",
  "PROFILE_CHANGE_REJECTED",
] as const
export type ServerNotification = {
  notification_id: number
  notification_type: (typeof notificationTypes)[number]
  resource_type: "PROFILE_CHANGE_REQUEST"
  resource_id: number
  message: string
  created_at: string
  read_at: string | null
}
export type NotificationPage = {
  items: ServerNotification[]
  page: number
  page_size: number
  total: number
  unread_count: number
}
function id(value: unknown): value is number {
  return Number.isSafeInteger(value) && (value as number) > 0 && (value as number) <= 2147483647
}
function object(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === "object" && !Array.isArray(value)
}
function exact(value: Record<string, unknown>, keys: string[]) {
  return Object.keys(value).length === keys.length && keys.every((key) => key in value)
}
function date(value: unknown): value is string {
  return (
    typeof value === "string" &&
    /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})$/.test(value) &&
    Number.isFinite(Date.parse(value))
  )
}
function invalid(): never {
  throw new AuthError(0, "INVALID_RESPONSE")
}
function validNotification(value: unknown): value is ServerNotification {
  return (
    object(value) &&
    exact(value, [
      "notification_id",
      "notification_type",
      "resource_type",
      "resource_id",
      "message",
      "created_at",
      "read_at",
    ]) &&
    id(value.notification_id) &&
    id(value.resource_id) &&
    value.resource_type === "PROFILE_CHANGE_REQUEST" &&
    notificationTypes.includes(
      value.notification_type as ServerNotification["notification_type"],
    ) &&
    typeof value.message === "string" &&
    [...value.message].length >= 1 &&
    [...value.message].length <= 200 &&
    !value.message.includes("\0") &&
    !/[\uD800-\uDFFF]/u.test(value.message) &&
    date(value.created_at) &&
    (value.read_at === null || date(value.read_at))
  )
}
export function parseNotificationPage(
  value: unknown,
  filter: NotificationFilter,
  page: number,
): NotificationPage {
  if (
    !object(value) ||
    !exact(value, ["items", "page", "page_size", "total", "unread_count"]) ||
    value.page !== page ||
    value.page_size !== 20 ||
    !Array.isArray(value.items) ||
    value.items.length > 20 ||
    !Number.isSafeInteger(value.total) ||
    (value.total as number) < 0 ||
    !Number.isSafeInteger(value.unread_count) ||
    (value.unread_count as number) < 0
  )
    invalid()
  const items = value.items as unknown[]
  const seen = new Set<number>()
  for (const entry of items) {
    if (
      !validNotification(entry) ||
      seen.has(entry.notification_id) ||
      (filter === "unread" && entry.read_at !== null)
    )
      invalid()
    seen.add(entry.notification_id)
  }
  if (
    (value.total as number) < items.length ||
    (filter === "all" && (value.unread_count as number) > (value.total as number)) ||
    (filter === "unread" && value.total !== value.unread_count)
  )
    invalid()
  return value as NotificationPage
}
export function notificationDestination(item: ServerNotification) {
  // Không dùng URL từ nội dung thông báo; GET có quyền vẫn là điều kiện trước điều hướng.
  if (!validNotification(item)) invalid()
  return item.notification_type === "PROFILE_CHANGE_SUBMITTED"
    ? `/personnel/profile-requests/${item.resource_id}`
    : `/profile?request_id=${item.resource_id}`
}
const errorCodes = [
  "VALIDATION_ERROR",
  "UNAUTHENTICATED",
  "FORBIDDEN",
  "PASSWORD_CHANGE_REQUIRED",
  "NOT_FOUND",
  "PAYLOAD_TOO_LARGE",
  "UNSUPPORTED_MEDIA_TYPE",
  "INTERNAL_ERROR",
  "SERVICE_UNAVAILABLE",
  "CSRF_INVALID",
  "ORIGIN_FORBIDDEN",
]
export function createServerNotificationsApi(fetcher: typeof fetch = fetch) {
  async function request(path: string, signal?: AbortSignal, csrf?: string) {
    let response: Response
    try {
      response = await fetcher(`/api/notifications${path}`, {
        method: csrf === undefined ? "GET" : "POST",
        credentials: "same-origin",
        cache: "no-store",
        redirect: "error",
        signal: signal
          ? AbortSignal.any([signal, AbortSignal.timeout(15000)])
          : AbortSignal.timeout(15000),
        headers: {
          Accept: "application/json",
          ...(csrf === undefined
            ? {}
            : { "Content-Type": "application/json", "X-CSRF-Token": csrf }),
        },
        ...(csrf === undefined ? {} : { body: "{}" }),
      })
    } catch {
      throw new AuthError(0, "NETWORK_ERROR")
    }
    if (!response.ok) {
      let code = "HTTP_ERROR"
      try {
        const error: unknown = await response.json()
        if (object(error) && typeof error.code === "string" && errorCodes.includes(error.code))
          code = error.code
      } catch {
        /* Lỗi proxy có thể không trả JSON. */
      }
      throw new AuthError(response.status, code)
    }
    return response
  }
  return {
    async list(filter: NotificationFilter = "all", page = 1, signal?: AbortSignal) {
      if (!["all", "unread"].includes(filter) || !id(page))
        throw new AuthError(400, "VALIDATION_ERROR")
      const params = new URLSearchParams({ filter, page: String(page), page_size: "20" })
      const response = await request(`?${params}`, signal)
      if (
        response.status !== 200 ||
        !response.headers.get("Content-Type")?.toLowerCase().startsWith("application/json")
      )
        invalid()
      let body: unknown
      try {
        body = await response.json()
      } catch {
        invalid()
      }
      return parseNotificationPage(body, filter, page)
    },
    async read(notificationId: number, csrf: string, signal?: AbortSignal) {
      if (!id(notificationId) || !csrf) throw new AuthError(400, "VALIDATION_ERROR")
      const response = await request(`/${notificationId}/read`, signal, csrf)
      if (response.status !== 204) invalid()
    },
  }
}
export type ServerNotificationsApi = ReturnType<typeof createServerNotificationsApi>
export function notificationErrorMessage(error: unknown) {
  if (error instanceof AuthError) {
    if (error.status === 401) return "Phiên không còn hiệu lực. Đang kiểm tra lại với máy chủ."
    if (error.status === 403)
      return "Chưa được phép thực hiện. Đang kiểm tra lại quyền với máy chủ."
    if (error.status === 404)
      return "Không tìm thấy thông báo hoặc tài nguyên trong quyền hiện tại."
    if (error.status >= 500) return "Dịch vụ thông báo chưa sẵn sàng. Hãy tải lại để kiểm tra."
  }
  return "Chưa tải được thông báo. Hãy thử tải lại."
}
