import type { Principal } from "../session/auth-api.ts"
import { AuthError } from "../session/auth-api.ts"
import { type ProfileApi, type SelfProfile, profileErrorMessage } from "./profile-api.ts"

type ProfileState = { phase: "loading" | "ready" | "error"; profile?: SelfProfile; message: string }
// Mỗi lượt tải có thế hệ riêng: phản hồi cũ không khôi phục PII sau khi đổi phiên/rời màn.
export function createProfileResource(api: ProfileApi) {
  let state: ProfileState = { phase: "loading", message: "" }
  let generation = 0
  let controller: AbortController | undefined
  const listeners = new Set<() => void>()
  function set(next: ProfileState) {
    state = next
    listeners.forEach((listener) => listener())
  }
  function clear() {
    ++generation
    controller?.abort()
    set({ phase: "loading", message: "" })
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
    async load(actor: Pick<Principal, "user_id" | "employee_id">, recheckSession: () => void) {
      clear()
      const current = generation
      controller = new AbortController()
      try {
        const profile = await api.readSelf(actor, controller.signal)
        if (current === generation) set({ phase: "ready", profile, message: "" })
      } catch (error) {
        if (current !== generation) return
        set({ phase: "error", message: profileErrorMessage(error) })
        if (error instanceof AuthError && (error.status === 401 || error.status === 403))
          recheckSession()
      }
    },
  }
}
