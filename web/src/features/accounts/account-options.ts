import { AuthError } from "../session/auth-api.ts"
import { catalogKinds, type CatalogApi, type CatalogKind } from "../organization/catalog-api.ts"

// Đọc đủ các trang; ID trùng do danh mục đổi giữa trang thì yêu cầu tải lại.
export async function loadAccountOptions(api: CatalogApi) {
  const entries = await Promise.all(
    catalogKinds.map(async (kind: CatalogKind) => {
      const items = []
      let page = 1
      let total = -1
      while (true) {
        const value = await api.list(kind, "", page, 100)
        if (total !== -1 && total !== value.pagination.total) throw new AuthError(409, "STALE_DATA")
        total = value.pagination.total
        items.push(...value.items)
        if (items.length >= total) break
        if (!value.items.length) throw new AuthError(409, "STALE_DATA")
        page++
      }
      if (new Set(items.map((item) => item.id)).size !== items.length)
        throw new AuthError(409, "STALE_DATA")
      return items.map((item) => ({ ...item, id: String(item.id) }))
    }),
  )
  return { departments: entries[0], positions: entries[1], jobLevels: entries[2] }
}
