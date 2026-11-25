import { reviewId, reviewStatuses, type ReviewQuery } from "../profiles/hr-review-api.ts"
import type { Principal } from "./auth-api.ts"
import { catalogKinds, type CatalogKind } from "../organization/catalog-api.ts"

export function canManageAccounts(principal: Principal) {
  return !principal.must_change_password && principal.system_roles.includes("ADMIN")
}
export function canReadCatalogs(principal: Principal) {
  return (
    !principal.must_change_password &&
    principal.system_roles.some((role) => role === "HR" || role === "ADMIN")
  )
}
export function canManageCatalogs(principal: Principal) {
  return !principal.must_change_password && principal.system_roles.includes("HR")
}
export function serverHome(principal: Principal) {
  return principal.must_change_password
    ? "/change-password"
    : canManageCatalogs(principal)
      ? "/organization/departments"
      : "/profile"
}
export function catalogKindAt(path: string): CatalogKind | undefined {
  return catalogKinds.find((kind) => path === `/organization/${kind}`)
}
export function parseCatalogQuery(search: string) {
  const params = new URLSearchParams(search)
  const allowed = ["q", "page", "page_size"]
  if ([...params.keys()].some((key) => !allowed.includes(key) || params.getAll(key).length !== 1))
    return null
  const q = params.get("q") ?? ""
  const page = params.get("page") ?? "1"
  const pageSize = params.get("page_size") ?? "20"
  if (
    [...q].length > 100 ||
    q.includes("\0") ||
    !/^[1-9]\d*$/.test(page) ||
    !/^[1-9]\d*$/.test(pageSize)
  )
    return null
  const p = Number(page),
    size = Number(pageSize)
  if (!Number.isSafeInteger(p) || p > 2147483647 || !Number.isSafeInteger(size) || size > 100)
    return null
  return { q, page: p, pageSize: size }
}

export function canReviewServerProfiles(principal: Principal) {
  return !principal.must_change_password && principal.system_roles.includes("HR")
}
export function parseReviewLocation(
  path: string,
  search: string,
): { id?: number; query: ReviewQuery } | null {
  const root = "/personnel/profile-requests"
  const query: ReviewQuery = { status: "PENDING", page: 1, pageSize: 20 }
  if (path !== root) {
    const match = /^\/personnel\/profile-requests\/([1-9]\d*)$/.exec(path)
    const id = match ? Number(match[1]) : 0
    return reviewId(id) && !search ? { id, query } : null
  }
  const params = new URLSearchParams(search)
  if (
    [...params.keys()].some(
      (key) => !["status", "page", "page_size"].includes(key) || params.getAll(key).length !== 1,
    )
  )
    return null
  const status = params.get("status")
  if (status !== null) {
    if (status === "all") query.status = undefined
    else if (reviewStatuses.includes(status as (typeof reviewStatuses)[number]))
      query.status = status as (typeof reviewStatuses)[number]
    else return null
  }
  const page = params.get("page") ?? "1",
    size = params.get("page_size") ?? "20"
  if (
    !/^[1-9]\d*$/.test(page) ||
    !/^[1-9]\d*$/.test(size) ||
    !reviewId(Number(page)) ||
    !reviewId(Number(size)) ||
    Number(size) > 100
  )
    return null
  query.page = Number(page)
  query.pageSize = Number(size)
  return { query }
}

export function parseAdministrativeLocation(path: string, search: string) {
  const root = "/personnel/employees"
  if (path === root) {
    const query = parseCatalogQuery(search)
    return query ? { query } : null
  }
  const match = /^\/personnel\/employees\/([1-9]\d*)$/.exec(path)
  const id = match ? Number(match[1]) : 0
  return reviewId(id) && !search ? { id, query: { q: "", page: 1, pageSize: 20 } } : null
}
