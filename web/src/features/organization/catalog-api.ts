import { AuthError } from "../session/auth-api.ts"

export const catalogKinds = ["departments", "positions", "job-levels"] as const
export type CatalogKind = (typeof catalogKinds)[number]
export type CatalogItem = { id: number; name: string }
export type CatalogPage = {
  items: CatalogItem[]
  pagination: { page: number; page_size: number; total: number }
}
export const catalogLabels: Record<CatalogKind, string> = {
  departments: "Phòng ban",
  positions: "Vị trí công việc",
  "job-levels": "Cấp bậc",
}

// Trim theo unicode.IsSpace của Go, không ngầm chuẩn hóa NFC hoặc bỏ dấu.
export function trimCatalogName(value: string) {
  return value.replace(/^[\p{White_Space}]+|[\p{White_Space}]+$/gu, "")
}
export function validCatalogName(value: string) {
  const name = trimCatalogName(value)
  return (
    name.length > 0 &&
    [...name].length <= 100 &&
    !name.includes("\0") &&
    !/[\uD800-\uDFFF]/u.test(name)
  )
}
function object(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value)
}
function exact(value: Record<string, unknown>, keys: string[]) {
  return Object.keys(value).length === keys.length && keys.every((key) => key in value)
}
function positive(value: unknown): value is number {
  return Number.isSafeInteger(value) && (value as number) > 0
}
function item(value: unknown): CatalogItem {
  if (
    !object(value) ||
    !exact(value, ["id", "name"]) ||
    !positive(value.id) ||
    value.id > 2147483647 ||
    typeof value.name !== "string" ||
    !validCatalogName(value.name)
  )
    throw new AuthError(0, "INVALID_RESPONSE")
  return value as CatalogItem
}
export function createCatalogApi(fetcher: typeof fetch = fetch) {
  async function request(kind: CatalogKind, suffix: string, init: RequestInit, status: number) {
    let response: Response
    try {
      response = await fetcher(`/api/${kind}${suffix}`, {
        ...init,
        credentials: "same-origin",
        cache: "no-store",
        redirect: "error",
        signal: init.signal
          ? AbortSignal.any([init.signal, AbortSignal.timeout(15000)])
          : AbortSignal.timeout(15000),
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
        /* Proxy có thể trả lỗi không phải JSON. */
      }
      throw new AuthError(response.status, code)
    }
    if (response.status !== status) throw new AuthError(0, "INVALID_RESPONSE")
    return response
  }
  async function json(response: Response): Promise<unknown> {
    try {
      if (!response.headers.get("Content-Type")?.toLowerCase().startsWith("application/json"))
        throw new Error()
      return await response.json()
    } catch {
      throw new AuthError(0, "INVALID_RESPONSE")
    }
  }
  async function write(
    kind: CatalogKind,
    method: string,
    entry: CatalogItem | undefined,
    name: string | undefined,
    token: string,
  ) {
    const response = await request(
      kind,
      entry ? `/${entry.id}` : "",
      {
        method,
        headers: { "Content-Type": "application/json", "X-CSRF-Token": token },
        body: JSON.stringify({
          ...(name !== undefined ? { name } : {}),
          ...(entry ? { expected_before_name: entry.name } : {}),
        }),
      },
      method === "POST" ? 201 : method === "DELETE" ? 204 : 200,
    )
    if (method === "DELETE") {
      if ((await response.text()) !== "") throw new AuthError(0, "INVALID_RESPONSE")
      return
    }
    return item(await json(response))
  }
  return {
    async list(
      kind: CatalogKind,
      query: string,
      page: number,
      pageSize: number,
      signal?: AbortSignal,
    ): Promise<CatalogPage> {
      const params = new URLSearchParams({
        q: query,
        page: String(page),
        page_size: String(pageSize),
      })
      const value = await json(await request(kind, `?${params}`, { signal }, 200))
      if (
        !object(value) ||
        !exact(value, ["items", "pagination"]) ||
        !Array.isArray(value.items) ||
        !object(value.pagination) ||
        !exact(value.pagination, ["page", "page_size", "total"])
      )
        throw new AuthError(0, "INVALID_RESPONSE")
      const p = value.pagination
      if (
        p.page !== page ||
        p.page_size !== pageSize ||
        !positive(p.page) ||
        !positive(p.page_size) ||
        p.page_size > 100 ||
        !Number.isSafeInteger(p.total) ||
        (p.total as number) < 0 ||
        value.items.length > pageSize ||
        value.items.length > (p.total as number)
      )
        throw new AuthError(0, "INVALID_RESPONSE")
      const items = value.items.map(item)
      if (new Set(items.map((entry) => entry.id)).size !== items.length)
        throw new AuthError(0, "INVALID_RESPONSE")
      return { items, pagination: p as CatalogPage["pagination"] }
    },
    create: (kind: CatalogKind, name: string, token: string) =>
      write(kind, "POST", undefined, name, token),
    rename: (kind: CatalogKind, entry: CatalogItem, name: string, token: string) =>
      write(kind, "PATCH", entry, name, token),
    remove: (kind: CatalogKind, entry: CatalogItem, token: string) =>
      write(kind, "DELETE", entry, undefined, token),
  }
}
export type CatalogApi = ReturnType<typeof createCatalogApi>
export function catalogErrorMessage(error: unknown, mutation = false) {
  if (!(error instanceof AuthError)) return "Chưa thể kết nối với hệ thống. Hãy tải lại dữ liệu."
  if (mutation && (error.status === 0 || error.status >= 500))
    return "Chưa thể xác nhận kết quả lưu. Hãy tải lại danh sách để đối soát; yêu cầu không được tự gửi lại."
  if (error.code === "DUPLICATE_NAME")
    return "Tên đã tồn tại trong danh mục này. Vui lòng chọn tên khác."
  if (error.code === "STALE_DATA")
    return "Mục này đã thay đổi. Hãy tải lại dữ liệu trước khi chỉnh sửa tiếp."
  if (error.code === "CATALOG_IN_USE")
    return "Mục đang được hồ sơ sử dụng, kể cả hồ sơ không hoạt động. Cần cập nhật các tham chiếu trước khi xóa."
  if (error.status === 404)
    return mutation
      ? "Mục này không còn tồn tại. Hãy tải lại danh sách."
      : "Chưa thể truy cập danh mục này trên máy chủ. Hãy kiểm tra backend rồi tải lại."
  if (error.status === 401) return "Phiên đã hết hiệu lực. Vui lòng đăng nhập lại."
  if (error.status === 403) return "Quyền hoặc phiên đã thay đổi. Đang kiểm tra lại với máy chủ."
  if (error.status === 400) return "Dữ liệu chưa hợp lệ. Tên và từ khóa tìm kiếm tối đa 100 ký tự."
  if (error.status === 503) return "Hệ thống tạm thời chưa sẵn sàng. Hãy thử tải lại sau."
  return "Chưa tải được dữ liệu hợp lệ. Hãy thử tải lại."
}
