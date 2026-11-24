import { AuthError } from "../session/auth-api.ts"
import type { AuthSession } from "../session/auth-session.ts"
import {
  notificationErrorMessage,
  type NotificationFilter,
  type NotificationPage,
  type ServerNotificationsApi,
} from "./server-notifications-api.ts"

export type NotificationsState = {
  phase: "idle" | "loading" | "ready" | "error"
  data?: NotificationPage
  filter: NotificationFilter
  page: number
  message: string
  readingId?: number
}
export function createServerNotificationsResource(
  api: ServerNotificationsApi,
  session: AuthSession,
) {
  let state: NotificationsState = { phase: "idle", filter: "all", page: 1, message: "" }
  let generation = 0
  let loadController: AbortController | undefined
  let readController: AbortController | undefined
  const listeners = new Set<() => void>()
  function set(next: NotificationsState) {
    state = next
    listeners.forEach((listener) => listener())
  }
  function identity() {
    const current = session.getSnapshot()
    return current.phase === "authenticated" &&
      current.principal &&
      !current.principal.must_change_password
      ? `${current.principal.user_id}:${current.principal.employee_id}`
      : undefined
  }
  let actor = identity()
  function clear() {
    ++generation
    loadController?.abort()
    readController?.abort()
    set({ phase: "idle", filter: "all", page: 1, message: "" })
  }
  let unsubscribe: (() => void) | undefined
  function sessionChanged() {
    const next = identity()
    if (next !== actor) {
      actor = next
      clear()
    }
  }
  async function failure(error: unknown, posted: boolean) {
    const denied = error instanceof AuthError && (error.status === 401 || error.status === 403)
    set({
      ...state,
      phase: "error",
      data: undefined,
      readingId: undefined,
      message:
        posted && (!(error instanceof AuthError) || error.status === 0 || error.status >= 500)
          ? "Chưa xác định kết quả đánh dấu đọc. Hãy tải lại để đối soát; chỉ thử đánh dấu lại khi bạn chủ động chọn."
          : notificationErrorMessage(error),
    })
    if (denied) await session.refresh()
  }
  async function load(filter: NotificationFilter = state.filter, page = state.page) {
    // Chờ read kết thúc để GET sau204 không bị thay bằng snapshot trước commit.
    if (state.readingId !== undefined) return
    loadController?.abort()
    const current = ++generation
    const owner = identity()
    if (!owner) {
      clear()
      return
    }
    loadController = new AbortController()
    set({ ...state, phase: "loading", data: undefined, filter, page, message: "" })
    try {
      const data = await api.list(filter, page, loadController.signal)
      if (current === generation && owner === identity())
        set({ ...state, phase: "ready", data, message: "" })
    } catch (error) {
      if (current === generation && owner === identity()) await failure(error, false)
    }
  }
  return {
    getSnapshot: () => state,
    subscribe(listener: () => void) {
      listeners.add(listener)
      if (!unsubscribe) {
        sessionChanged()
        unsubscribe = session.subscribe(sessionChanged)
      }
      return () => {
        listeners.delete(listener)
        if (!listeners.size) {
          unsubscribe?.()
          unsubscribe = undefined
          clear()
        }
      }
    },
    clear,
    dispose() {
      clear()
      unsubscribe?.()
      unsubscribe = undefined
      listeners.clear()
    },
    load,
    async markRead(notificationId: number) {
      const owner = identity()
      if (!owner || state.readingId !== undefined || session.getSnapshot().busy) return false
      const current = generation
      readController = new AbortController()
      const signal = readController.signal
      set({ ...state, readingId: notificationId, message: "" })
      let posted = false
      try {
        await session.catalogMutation((csrf) => {
          posted = true
          return api.read(notificationId, csrf, signal)
        })
        if (current !== generation || owner !== identity()) return false
        set({ ...state, readingId: undefined })
        // Count/list chỉ lấy lại từ snapshot server; không trừ lạc quan theo đồng hồ client.
        await load()
        return true
      } catch (error) {
        if (current === generation && owner === identity()) await failure(error, posted)
        else if (owner !== identity()) clear()
        return false
      } finally {
        if (owner === identity() && state.readingId === notificationId)
          set({ ...state, readingId: undefined })
      }
    },
  }
}
