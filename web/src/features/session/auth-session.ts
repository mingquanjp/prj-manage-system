import { AuthError, authErrorMessage, type AuthApi, type Principal } from "./auth-api.ts"

export type SessionState = {
  phase: "checking" | "signed-out" | "authenticated" | "unavailable"
  principal: Principal | null
  busy: boolean
  message: string
  retryAt: number
}

// State của phiên server không nhận fixture hoặc danh tính do client chọn.
// Số thứ tự ngăn phản hồi cũ khôi phục principal sau một lần kiểm tra mới hơn.
export function createAuthSession(api: AuthApi, now: () => number = Date.now) {
  let state: SessionState = {
    phase: "checking",
    principal: null,
    busy: false,
    message: "",
    retryAt: 0,
  }
  let generation = 0
  let csrf: string | null = null
  const listeners = new Set<() => void>()
  function set(next: SessionState) {
    state = next
    listeners.forEach((listener) => listener())
  }
  function signedOut(message: string, retryAt = state.retryAt) {
    csrf = null
    set({ phase: "signed-out", principal: null, busy: false, message, retryAt })
  }
  function failed(error: unknown, previous: SessionState, operation: "check" | "login" | "logout") {
    csrf = null
    const issue = error instanceof AuthError ? error : new AuthError(0, "NETWORK_ERROR")
    if (issue.status === 401 && operation !== "login") {
      signedOut(
        operation === "logout" || previous.principal
          ? "Phiên không còn hiệu lực. Vui lòng đăng nhập lại."
          : previous.phase === "signed-out"
            ? previous.message
            : "",
      )
      return
    }
    const uncertain = issue.status === 0 || issue.status >= 500
    set({
      phase: uncertain ? "unavailable" : operation === "check" ? "unavailable" : previous.phase,
      principal: uncertain || operation === "check" ? null : previous.principal,
      busy: false,
      message: authErrorMessage(issue),
      retryAt:
        issue.status === 429 && issue.retryAfterSeconds !== null
          ? now() + issue.retryAfterSeconds * 1000
          : previous.retryAt,
    })
  }
  return {
    getSnapshot: () => state,
    subscribe: (listener: () => void) => {
      listeners.add(listener)
      return () => {
        listeners.delete(listener)
      }
    },
    async check() {
      if (state.busy) return
      const id = ++generation
      const previous = state
      csrf = null
      set({ phase: "checking", principal: null, busy: false, message: "", retryAt: state.retryAt })
      try {
        const principal = await api.me()
        if (id === generation)
          set({ phase: "authenticated", principal, busy: false, message: "", retryAt: 0 })
      } catch (error) {
        if (id === generation) failed(error, previous, "check")
      }
    },
    async login(email: string, password: string) {
      if (state.phase !== "signed-out" || state.busy || now() < state.retryAt) return
      const previous = state
      const id = ++generation
      csrf = null
      set({ ...state, busy: true, message: "" })
      try {
        const principal = await api.login(email, password)
        if (id === generation)
          set({ phase: "authenticated", principal, busy: false, message: "", retryAt: 0 })
      } catch (error) {
        if (id === generation) failed(error, previous, "login")
      }
    },
    async logout() {
      if (state.phase !== "authenticated" || state.busy || now() < state.retryAt) return
      const previous = state
      const id = ++generation
      set({ ...state, busy: true, message: "" })
      try {
        csrf ??= await api.csrf()
        await api.logout(csrf)
        if (id === generation) signedOut("Bạn đã đăng xuất.", 0)
      } catch (error) {
        if (id === generation) failed(error, previous, "logout")
      }
    },
  }
}
