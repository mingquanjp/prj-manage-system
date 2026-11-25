import { AuthError } from "../session/auth-api.ts"
import type { AuthSession } from "../session/auth-session.ts"
import {
  catalogKinds,
  createCatalogApi,
  type CatalogApi,
  type CatalogItem,
  type CatalogKind,
} from "../organization/catalog-api.ts"
import {
  administrativeErrorMessage,
  administrativeValues,
  type AdministrativePage,
  type AdministrativePatch,
  type AdministrativeProfile,
  type AdministrativeProfileApi,
  type AdministrativeQuery,
} from "./administrative-profile-api.ts"
type State = {
  phase: "loading" | "ready" | "error"
  page?: AdministrativePage
  detail?: AdministrativeProfile
  catalogs?: Record<CatalogKind, CatalogItem[]>
  saving: boolean
  mustReload: boolean
  message: string
  fieldError?: "company_email"
}
export function createAdministrativeProfileResource(
  api: AdministrativeProfileApi,
  session: AuthSession,
  catalogs: Pick<CatalogApi, "list"> = createCatalogApi(),
) {
  let state: State = { phase: "loading", saving: false, mustReload: false, message: "" }
  let generation = 0
  let controller: AbortController | undefined
  const listeners = new Set<() => void>()
  const set = (next: State) => {
    state = next
    listeners.forEach((fn) => fn())
  }
  const allowed = () => {
    const s = session.getSnapshot()
    return (
      s.phase === "authenticated" &&
      !!s.principal?.system_roles.includes("HR") &&
      !s.principal.must_change_password
    )
  }
  function clear() {
    ++generation
    controller?.abort()
    set({ phase: "loading", saving: false, mustReload: false, message: "" })
  }
  const current = (version: number, actor: number | undefined) =>
    version === generation && actor === session.getSnapshot().principal?.user_id && allowed()
  async function failure(error: unknown, posted = false) {
    const denied = error instanceof AuthError && [401, 403].includes(error.status)
    const uncertain =
      posted && (!(error instanceof AuthError) || error.status === 0 || error.status >= 500)
    set({
      ...state,
      ...(denied ? { page: undefined, detail: undefined, catalogs: undefined } : {}),
      phase: denied || (!state.page && !state.detail) ? "error" : "ready",
      saving: false,
      mustReload:
        uncertain ||
        (posted &&
          error instanceof AuthError &&
          ["STALE_DATA", "CATALOG_NOT_FOUND"].includes(error.code)),
      fieldError:
        error instanceof AuthError && error.code === "EMAIL_CONFLICT" ? "company_email" : undefined,
      message: uncertain
        ? "Chưa xác định kết quả lưu. Hãy tải lại hồ sơ để đối soát; không gửi lại thay đổi này."
        : administrativeErrorMessage(error),
    })
    if (denied) await session.refresh()
  }
  async function allCatalogs(signal: AbortSignal) {
    const result = {} as Record<CatalogKind, CatalogItem[]>
    for (const kind of catalogKinds) {
      const items: CatalogItem[] = []
      for (let page = 1; ; page++) {
        if (signal.aborted) throw new AuthError(0, "NETWORK_ERROR")
        const batch = await catalogs.list(kind, "", page, 100, signal)
        if (signal.aborted) throw new AuthError(0, "NETWORK_ERROR")
        for (const entry of batch.items) {
          if (items.some((old) => old.id === entry.id)) throw new AuthError(0, "INVALID_RESPONSE")
          items.push(entry)
        }
        if (page * 100 >= batch.pagination.total) break
        if (!batch.items.length) throw new AuthError(0, "INVALID_RESPONSE")
      }
      result[kind] = items
    }
    return result
  }
  return {
    getSnapshot: () => state,
    subscribe(fn: () => void) {
      listeners.add(fn)
      return () => {
        listeners.delete(fn)
      }
    },
    clear,
    async load(input: number | AdministrativeQuery) {
      if (state.saving) return
      clear()
      const version = generation,
        actor = session.getSnapshot().principal?.user_id
      controller = new AbortController()
      const signal = controller.signal
      if (!allowed()) return
      try {
        const result =
          typeof input === "number"
            ? { detail: await api.detail(input, signal), catalogs: await allCatalogs(signal) }
            : { page: await api.list(input, signal) }
        if (current(version, actor))
          set({ phase: "ready", ...result, saving: false, mustReload: false, message: "" })
      } catch (error) {
        if (current(version, actor)) await failure(error)
      }
    },
    async save(patch: AdministrativePatch) {
      if (
        !allowed() ||
        state.saving ||
        state.mustReload ||
        !state.detail ||
        session.getSnapshot().busy
      )
        return false
      const version = generation,
        actor = session.getSnapshot().principal?.user_id
      const before = state.detail
      let posted = false
      set({ ...state, saving: true, message: "", fieldError: undefined })
      try {
        const detail = await session.catalogMutation((token) => {
          posted = true
          return api.edit(before.employee_id, patch, token, controller?.signal)
        })
        if (!current(version, actor)) return false
        const emailChanged = before.company_email !== detail.company_email
        const changed =
          JSON.stringify(administrativeValues(before)) !==
          JSON.stringify(administrativeValues(detail))
        set({
          ...state,
          detail,
          saving: false,
          mustReload: false,
          message: emailChanged
            ? "Đã lưu hồ sơ. Email đăng nhập đã đổi; nhân viên phải đăng nhập lại bằng email mới."
            : changed
              ? "Đã lưu hồ sơ hành chính."
              : "Không có thay đổi dữ liệu; hồ sơ được giữ nguyên.",
        })
        if (emailChanged && detail.user_id === actor) await session.refresh()
        return true
      } catch (error) {
        if (current(version, actor)) await failure(error, posted)
        return false
      }
    },
  }
}
