import { AuthError } from "../session/auth-api.ts"
import type { AuthSession } from "../session/auth-session.ts"
import {
  type HrReviewApi,
  type ReviewDetail,
  type ReviewPage,
  type ReviewQuery,
  reviewErrorMessage,
} from "./hr-review-api.ts"

type ReviewState = {
  phase: "loading" | "ready" | "error"
  page?: ReviewPage
  detail?: ReviewDetail
  message: string
  deciding: boolean
  mustReload: boolean
}
// Chặn phản hồi cũ khôi phục PII khi rời màn hoặc đổi actor; không tự phát lại quyết định.
export function createHrReviewResource(api: HrReviewApi, session: AuthSession) {
  let state: ReviewState = { phase: "loading", message: "", deciding: false, mustReload: false }
  let generation = 0
  let controller: AbortController | undefined
  const listeners = new Set<() => void>()
  function set(next: ReviewState) {
    state = next
    listeners.forEach((listener) => listener())
  }
  function clear() {
    ++generation
    controller?.abort()
    set({ phase: "loading", message: "", deciding: false, mustReload: false })
  }
  function allowed() {
    const current = session.getSnapshot()
    return (
      current.phase === "authenticated" &&
      current.principal?.system_roles.includes("HR") &&
      !current.principal.must_change_password
    )
  }
  async function failure(error: unknown, mutation: boolean) {
    const denied = error instanceof AuthError && (error.status === 401 || error.status === 403)
    const uncertain =
      mutation && (!(error instanceof AuthError) || error.status === 0 || error.status >= 500)
    set({
      ...state,
      ...(denied ? { page: undefined, detail: undefined } : {}),
      phase: denied || (!state.detail && !state.page) ? "error" : "ready",
      deciding: false,
      mustReload:
        mutation &&
        (uncertain ||
          (error instanceof AuthError &&
            error.status === 409 &&
            error.code !== "EMPLOYEE_INACTIVE")),
      message: uncertain
        ? "Chưa xác định kết quả xử lý. Hãy tải lại đề nghị để đối soát; không tự gửi lại quyết định."
        : reviewErrorMessage(error),
    })
    if (denied) await session.refresh()
  }
  return {
    getSnapshot: () => state,
    subscribe(listener: () => void) {
      listeners.add(listener)
      return () => {
        listeners.delete(listener)
      }
    },
    clear,
    async load(input: number | ReviewQuery) {
      if (state.deciding) return
      clear()
      const id = generation
      const actor = session.getSnapshot().principal?.user_id
      controller = new AbortController()
      if (!allowed()) return
      try {
        const result =
          typeof input === "number"
            ? { detail: await api.detail(input, controller.signal) }
            : { page: await api.list(input, controller.signal) }
        if (id === generation && actor === session.getSnapshot().principal?.user_id && allowed())
          set({ phase: "ready", ...result, message: "", deciding: false, mustReload: false })
      } catch (error) {
        if (id === generation) await failure(error, false)
      }
    },
    async decide(decision: "approve" | "reject", reason: string) {
      if (
        !allowed() ||
        state.deciding ||
        state.mustReload ||
        state.detail?.status !== "PENDING" ||
        session.getSnapshot().busy
      )
        return false
      const current = generation
      const actor = session.getSnapshot().principal?.user_id
      const id = state.detail.request_id
      set({ ...state, deciding: true, message: "" })
      let posted = false
      try {
        const detail = await session.catalogMutation((token) => {
          posted = true
          return api.decide(id, decision, reason, token)
        })
        if (
          current !== generation ||
          actor !== session.getSnapshot().principal?.user_id ||
          !allowed()
        )
          return false
        set({
          phase: "ready",
          detail,
          deciding: false,
          mustReload: false,
          message:
            decision === "approve"
              ? "Đã chấp thuận đề nghị và cập nhật hồ sơ chính thức."
              : "Đã từ chối đề nghị. Hồ sơ chính thức được giữ nguyên.",
        })
        return true
      } catch (error) {
        if (current === generation) await failure(error, posted)
        return false
      }
    },
  }
}
