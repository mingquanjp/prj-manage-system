import { AuthError } from "../session/auth-api.ts"
import { trimCatalogName } from "../organization/catalog-api.ts"
import type { CatalogItem } from "../organization/catalog-api.ts"
import { reviewId } from "./hr-review-api.ts"

export const administrativeFields = [
  "fullname",
  "company_email",
  "phone",
  "address",
  "position_id",
  "department_id",
  "job_level_id",
] as const
export type AdministrativeField = (typeof administrativeFields)[number]
export type AdministrativeValues = Record<AdministrativeField, string | number | null>
export type AdministrativePatch = {
  expected_before_values: Partial<AdministrativeValues>
  changes: Partial<AdministrativeValues>
}
export type AdministrativeSummary = {
  employee_id: number
  employee_code: string
  fullname: string
  company_email: string
  position: CatalogItem
  department: CatalogItem | null
  job_level: CatalogItem | null
  employment_status: "ACTIVE" | "INACTIVE" | "RESIGNED"
}
export type AdministrativeProfile = AdministrativeSummary & {
  user_id: number
  phone: string | null
  address: string | null
}
export type AdministrativeQuery = { q: string; page: number; pageSize: number }
export type AdministrativePage = {
  items: AdministrativeSummary[]
  page: number
  page_size: number
  total: number
}
const object = (v: unknown): v is Record<string, unknown> =>
  typeof v === "object" && v !== null && !Array.isArray(v)
const exact = (v: Record<string, unknown>, keys: string[]) =>
  Object.keys(v).length === keys.length && keys.every((k) => k in v)
const text = (v: unknown, min = 0, max = Infinity): v is string =>
  typeof v === "string" &&
  [...v].length >= min &&
  [...v].length <= max &&
  !/[\0\uD800-\uDFFF]/u.test(v)
const catalog = (v: unknown) =>
  object(v) && exact(v, ["id", "name"]) && reviewId(v.id) && text(v.name, 1, 100)
const summaryKeys = [
  "employee_id",
  "employee_code",
  "fullname",
  "company_email",
  "position",
  "department",
  "job_level",
  "employment_status",
]
function invalid(): never {
  throw new AuthError(0, "INVALID_RESPONSE")
}
function summary(v: Record<string, unknown>) {
  return (
    reviewId(v.employee_id) &&
    text(v.employee_code, 1, 30) &&
    text(v.fullname, 1, 100) &&
    text(v.company_email, 3, 100) &&
    !/[\r\n]/.test(v.company_email) &&
    catalog(v.position) &&
    (v.department === null || catalog(v.department)) &&
    (v.job_level === null || catalog(v.job_level)) &&
    ["ACTIVE", "INACTIVE", "RESIGNED"].includes(String(v.employment_status))
  )
}
export function parseAdministrativeProfile(v: unknown, id: number): AdministrativeProfile {
  if (
    !object(v) ||
    !exact(v, [...summaryKeys, "user_id", "phone", "address"]) ||
    !summary(v) ||
    v.employee_id !== id ||
    !reviewId(v.user_id) ||
    !(v.phone === null || text(v.phone)) ||
    !(v.address === null || text(v.address))
  )
    invalid()
  return v as AdministrativeProfile
}
export function parseAdministrativePage(
  v: unknown,
  query: AdministrativeQuery,
): AdministrativePage {
  if (
    !object(v) ||
    !exact(v, ["items", "page", "page_size", "total"]) ||
    v.page !== query.page ||
    v.page_size !== query.pageSize ||
    !Number.isSafeInteger(v.total) ||
    (v.total as number) < 0 ||
    !Array.isArray(v.items) ||
    v.items.length > query.pageSize ||
    v.items.length > (v.total as number)
  )
    invalid()
  let last = 0
  for (const item of v.items as unknown[]) {
    if (
      !object(item) ||
      !exact(item, summaryKeys) ||
      !summary(item) ||
      (item.employee_id as number) <= last
    )
      invalid()
    last = item.employee_id as number
  }
  return v as AdministrativePage
}
export function administrativeValues(profile: AdministrativeProfile): AdministrativeValues {
  return {
    fullname: profile.fullname,
    company_email: profile.company_email,
    phone: profile.phone,
    address: profile.address,
    position_id: profile.position.id,
    department_id: profile.department?.id ?? null,
    job_level_id: profile.job_level?.id ?? null,
  }
}
// Giữ raw before, chỉ chuẩn hóa trường thực sự được sửa; không cắt legacy dài.
export function administrativePatch(
  before: AdministrativeValues,
  draft: AdministrativeValues,
): AdministrativePatch {
  const patch: AdministrativePatch = { expected_before_values: {}, changes: {} }
  for (const key of administrativeFields) {
    if (draft[key] === before[key]) continue
    let value = draft[key]
    if (typeof value === "string") value = trimCatalogName(value)
    // Chỉ ASCII có lower tương đương Go; Unicode do server chuẩn hóa cuối cùng.
    if (
      key === "company_email" &&
      typeof value === "string" &&
      [...value].every((character) => character.codePointAt(0)! < 128)
    )
      value = value.toLowerCase()
    if (value === before[key]) continue
    patch.expected_before_values[key] = before[key]
    patch.changes[key] = value
  }
  return patch
}
export function administrativeErrors(patch: AdministrativePatch) {
  const errors: Partial<Record<AdministrativeField, string>> = {}
  for (const key of administrativeFields) {
    if (!(key in patch.changes)) continue
    const v = patch.changes[key]
    if (["position_id", "department_id", "job_level_id"].includes(key)) {
      if (!(v === null && key !== "position_id") && !reviewId(v))
        errors[key] = "Chọn danh mục hợp lệ."
    } else if (!(v === null && key === "phone")) {
      const limit = { fullname: 100, company_email: 100, phone: 20, address: 2000 }[
        key as "fullname"
      ]
      if (
        !text(v, 1, limit) ||
        (key === "company_email" && (/[\r\n]/.test(String(v)) || !String(v).includes("@")))
      )
        errors[key] = `Nhập thông tin hợp lệ từ 1 đến ${limit} ký tự.`
    }
  }
  if (new TextEncoder().encode(JSON.stringify(patch)).length > 65536)
    errors.address =
      "Nội dung vượt 65536 byte UTF-8, gồm cả giá trị nền. Hãy rút ngắn nội dung hoặc liên hệ quản lý dữ liệu legacy."
  return errors
}
export function createAdministrativeProfileApi(fetcher: typeof fetch = fetch) {
  async function request(
    path: string,
    signal?: AbortSignal,
    patch?: AdministrativePatch,
    token?: string,
  ) {
    let response: Response
    try {
      response = await fetcher(`/api/employees${path}`, {
        method: patch ? "PATCH" : "GET",
        credentials: "same-origin",
        cache: "no-store",
        redirect: "error",
        signal: signal
          ? AbortSignal.any([signal, AbortSignal.timeout(15000)])
          : AbortSignal.timeout(15000),
        headers: {
          Accept: "application/json",
          ...(patch ? { "Content-Type": "application/json", "X-CSRF-Token": token ?? "" } : {}),
        },
        ...(patch ? { body: JSON.stringify(patch) } : {}),
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
        /* Proxy có thể không trả JSON. */
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
    async list(query: AdministrativeQuery, signal?: AbortSignal) {
      if (
        !reviewId(query.page) ||
        !reviewId(query.pageSize) ||
        query.pageSize > 100 ||
        !text(query.q, 0, 100)
      )
        throw new AuthError(400, "INVALID_QUERY")
      return parseAdministrativePage(
        await request(
          `?${new URLSearchParams({ q: query.q, page: String(query.page), page_size: String(query.pageSize) })}`,
          signal,
        ),
        query,
      )
    },
    async detail(id: number, signal?: AbortSignal) {
      if (!reviewId(id)) throw new AuthError(400, "INVALID_ID")
      return parseAdministrativeProfile(await request(`/${id}/administrative-profile`, signal), id)
    },
    async edit(id: number, patch: AdministrativePatch, token: string, signal?: AbortSignal) {
      if (
        !reviewId(id) ||
        !Object.keys(patch.changes).length ||
        Object.keys(administrativeErrors(patch)).length
      )
        throw new AuthError(400, "INVALID_INPUT")
      return parseAdministrativeProfile(await request(`/${id}`, signal, patch, token), id)
    },
  }
}
export type AdministrativeProfileApi = ReturnType<typeof createAdministrativeProfileApi>
export function administrativeErrorMessage(error: unknown) {
  if (error instanceof AuthError) {
    if (error.code === "STALE_DATA")
      return "Hồ sơ đã thay đổi. Hãy tải lại để đối chiếu; không ghi đè dữ liệu mới."
    if (error.code === "EMAIL_CONFLICT") return "Email đã được sử dụng. Hãy chọn email khác."
    if (error.code === "CATALOG_NOT_FOUND")
      return "Danh mục đã thay đổi hoặc bị xóa. Hãy tải lại hồ sơ và danh mục."
    if (error.status === 401) return "Phiên không còn hiệu lực. Đang kiểm tra lại phiên."
    if (error.status === 403)
      return "Bạn không còn quyền HR hoặc phiên đã thay đổi. Đang kiểm tra lại quyền."
    if (error.status === 404) return "Không tìm thấy hồ sơ."
    if (error.status === 400 || error.status === 413)
      return "Dữ liệu chưa hợp lệ hoặc vượt giới hạn. Hãy kiểm tra các trường đã sửa."
    if (error.status >= 500) return "Dịch vụ chưa sẵn sàng. Hãy thử tải lại."
  }
  return "Chưa tải được dữ liệu hợp lệ. Hãy kiểm tra kết nối rồi tải lại."
}
