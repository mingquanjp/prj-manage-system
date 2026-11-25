import { AuthError } from "../session/auth-api.ts"
import {
  requestErrorMessage,
  type ProfileRequestsApi,
  type RequestDetail,
  type RequestPage,
} from "./profile-requests-api.ts"

type ReadState<T> = { phase: "loading" | "ready" | "error"; data?: T; message: string }
type State = {
  list: ReadState<RequestPage>
  detail: ReadState<RequestDetail>
  submission: "idle" | "pending" | "success" | "stale" | "uncertain" | "error"
  message: string
}
const empty = (): State => ({
  list: { phase: "loading", message: "" },
  detail: { phase: "loading", message: "" },
  submission: "idle",
  message: "",
})

// Thế hệ phiên và từng GET ngăn phản hồi cũ khôi phục PII hoặc đổi trang đã chọn.
export function createProfileRequestsResource(api: ProfileRequestsApi) {
  let state = empty()
  let generation = 0
  let listGeneration = 0
  let detailGeneration = 0
  let listController: AbortController | undefined
  let detailController: AbortController | undefined
  let submitController: AbortController | undefined
  const listeners = new Set<() => void>()
  const set = (next: State) => {
    state = next
    listeners.forEach((listener) => listener())
  }
  const clear = () => {
    ++generation
    ++listGeneration
    ++detailGeneration
    listController?.abort()
    detailController?.abort()
    submitController?.abort()
    set(empty())
  }
  function sessionError(error: unknown, recheck: () => void) {
    if (error instanceof AuthError && (error.status === 401 || error.status === 403)) {
      clear()
      recheck()
      return true
    }
    return false
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
    clearDetail() {
      ++detailGeneration
      detailController?.abort()
      set({ ...state, detail: { phase: "loading", message: "" } })
    },
    async loadList(employeeId: number, page: number, recheck: () => void) {
      const epoch = generation
      const current = ++listGeneration
      listController?.abort()
      listController = new AbortController()
      set({ ...state, list: { phase: "loading", message: "" } })
      try {
        const data = await api.list(employeeId, page, listController.signal)
        if (epoch === generation && current === listGeneration)
          set({ ...state, list: { phase: "ready", data, message: "" } })
      } catch (error) {
        if (epoch !== generation || current !== listGeneration) return
        if (!sessionError(error, recheck))
          set({ ...state, list: { phase: "error", message: requestErrorMessage(error) } })
      }
    },
    async loadDetail(employeeId: number, requestId: number, recheck: () => void) {
      const epoch = generation
      const current = ++detailGeneration
      detailController?.abort()
      detailController = new AbortController()
      set({ ...state, detail: { phase: "loading", message: "" } })
      try {
        const data = await api.read(employeeId, requestId, detailController.signal)
        if (epoch === generation && current === detailGeneration) {
          set({
            ...state,
            ...(data.status !== "PENDING" && state.submission === "success"
              ? { submission: "idle" as const, message: "" }
              : {}),
            detail: { phase: "ready", data, message: "" },
          })
          return data
        }
      } catch (error) {
        if (epoch !== generation || current !== detailGeneration) return
        if (!sessionError(error, recheck))
          set({ ...state, detail: { phase: "error", message: requestErrorMessage(error) } })
      }
    },
    async submit(
      operation: (post: (token: string) => Promise<RequestDetail>) => Promise<RequestDetail>,
      post: (token: string, signal: AbortSignal) => Promise<RequestDetail>,
      recheck: () => void,
    ): Promise<RequestDetail | undefined> {
      if (["pending", "uncertain", "stale"].includes(state.submission))
        throw new Error("Hãy tải lại hồ sơ/lịch sử để đối soát trước khi tạo đề nghị mới.")
      const epoch = generation
      let posted = false
      submitController = new AbortController()
      const signal = submitController.signal
      set({ ...state, submission: "pending", message: "" })
      try {
        const detail = await operation((token) => {
          posted = true
          return post(token, signal)
        })
        if (epoch !== generation) return
        set({
          ...state,
          submission: "success",
          message: "Đã gửi đề nghị. Hồ sơ chính thức giữ nguyên trong thời gian chờ HR xử lý.",
        })
        return detail
      } catch (error) {
        if (epoch !== generation) return
        if (sessionError(error, recheck)) return
        const uncertain =
          posted && (!(error instanceof AuthError) || error.status === 0 || error.status >= 500)
        const stale = error instanceof AuthError && error.code === "STALE_DATA"
        const message = uncertain
          ? "Chưa xác định kết quả gửi đề nghị. Hãy đóng form và tải lịch sử để đối soát; không gửi lại theo phỏng đoán."
          : requestErrorMessage(error)
        if (uncertain) {
          ++listGeneration
          listController?.abort()
        }
        set({
          ...state,
          ...(uncertain
            ? {
                list: {
                  phase: "error" as const,
                  message: "Hãy tải lại lịch sử để đối soát lần gửi chưa rõ kết quả.",
                },
              }
            : {}),
          submission: uncertain ? "uncertain" : stale ? "stale" : "error",
          message,
        })
        throw new Error(message, { cause: error })
      }
    },
    allowNewDraft() {
      if (
        state.submission === "pending" ||
        (state.submission === "uncertain" && state.list.phase !== "ready")
      )
        return false
      set({ ...state, submission: "idle", message: "" })
      return true
    },
  }
}
