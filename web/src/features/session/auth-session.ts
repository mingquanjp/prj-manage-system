import { AuthError, authErrorMessage, type AuthApi, type Principal } from "./auth-api.ts"

export type SessionState = {
  phase: "checking" | "signed-out" | "authenticated" | "unavailable" | "password-uncertain"
  principal: Principal | null
  busy: boolean
  message: string
  retryAt: number
  issueCode?: string
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
  let recoveryLogin = false
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
  function failed(
    error: unknown,
    previous: SessionState,
    operation: "check" | "login" | "logout" | "change-password",
  ) {
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
      issueCode: issue.code,
      retryAt:
        issue.status === 429 && issue.retryAfterSeconds !== null
          ? now() + issue.retryAfterSeconds * 1000
          : previous.retryAt,
    })
  }
  async function check() {
    if (
      state.busy ||
      (recoveryLogin && state.phase === "signed-out") ||
      state.phase === "password-uncertain"
    )
      return
    const id = ++generation
    const previous = state
    csrf = null
    set({ phase: "checking", principal: null, busy: false, message: "", retryAt: state.retryAt })
    try {
      const principal = await api.me()
      if (id === generation)
        set({
          phase: "authenticated",
          principal,
          busy: false,
          message: "",
          retryAt: previous.retryAt,
        })
    } catch (error) {
      if (id === generation) failed(error, previous, "check")
    }
  }
  return {
    getSnapshot: () => state,
    subscribe: (listener: () => void) => {
      listeners.add(listener)
      return () => {
        listeners.delete(listener)
      }
    },
    check,
    async login(email: string, password: string) {
      if (state.phase !== "signed-out" || state.busy || now() < state.retryAt) return
      const previous = state
      const id = ++generation
      csrf = null
      set({ ...state, busy: true, message: "" })
      try {
        const principal = await api.login(email, password)
        recoveryLogin = false
        if (id === generation)
          set({ phase: "authenticated", principal, busy: false, message: "", retryAt: 0 })
      } catch (error) {
        if (id === generation) failed(error, previous, "login")
      }
    },
    async changePassword(currentPassword: string, newPassword: string) {
      if (
        state.phase !== "authenticated" ||
        !state.principal?.must_change_password ||
        state.busy ||
        now() < state.retryAt
      )
        return
      const previous = state
      const id = ++generation
      let posted = false
      set({ ...state, busy: true, message: "", issueCode: undefined })
      try {
        csrf ??= await api.csrf()
        posted = true
        await api.changePassword(currentPassword, newPassword, csrf)
        if (id === generation) {
          recoveryLogin = true
          signedOut("Đổi mật khẩu thành công. Vui lòng đăng nhập bằng mật khẩu mới.", 0)
        }
      } catch (error) {
        if (id !== generation) return
        csrf = null
        const issue = error instanceof AuthError ? error : new AuthError(0, "NETWORK_ERROR")
        if (posted && (issue.status === 0 || issue.status >= 500)) {
          set({
            phase: "password-uncertain",
            principal: null,
            busy: false,
            retryAt: 0,
            message:
              "Chưa thể xác nhận kết quả đổi mật khẩu. Hãy đăng nhập bằng mật khẩu mới để kiểm tra. Nếu không thành công, thử mật khẩu cũ.",
          })
        } else if (issue.status === 403 && issue.code === "PASSWORD_CHANGE_NOT_REQUIRED") {
          set({ ...previous, busy: false })
          await check()
        } else {
          failed(issue, previous, "change-password")
          if (issue.status === 400)
            set({
              ...state,
              message:
                issue.code === "INVALID_CURRENT_PASSWORD"
                  ? "Mật khẩu hiện tại không đúng. Vui lòng nhập lại."
                  : "Mật khẩu mới chưa đáp ứng yêu cầu. Hãy kiểm tra lại nội dung nhập.",
            })
        }
      }
    },
    loginAfterUncertainChange() {
      if (state.phase !== "password-uncertain") return
      recoveryLogin = true
      ++generation
      signedOut(state.message, 0)
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
