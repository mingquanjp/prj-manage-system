import type { Principal } from "./auth-api.ts"
import { catalogKinds, type CatalogKind } from "../organization/catalog-api.ts"

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
