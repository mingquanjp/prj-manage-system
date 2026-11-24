import { AuthError } from "../session/auth-api.ts"
import type { AuthSession } from "../session/auth-session.ts"
import type { HrReviewApi } from "../profiles/hr-review-api.ts"
import type { ProfileRequestsApi } from "../profiles/profile-requests-api.ts"
import { notificationDestination, type ServerNotification } from "./server-notifications-api.ts"

type OpenState = { openingId: number | null; message: string }

// Một GET mới hoặc đóng/rời khung làm vô hiệu kết quả cũ, kể cả fetch không tôn trọng abort.
export function createNotificationOpener({
  session,
  selfApi,
  reviewApi,
  onState,
  onOpen,
  markRead,
}: {
  session: Pick<AuthSession, "getSnapshot" | "refresh">
  selfApi: Pick<ProfileRequestsApi, "read">
  reviewApi: Pick<HrReviewApi, "detail">
  onState: (state: OpenState) => void
  onOpen: (destination: string) => void
  markRead: (id: number) => Promise<boolean>
}) {
  let generation = 0
  let controller: AbortController | undefined
  function identity() {
    const current = session.getSnapshot()
    const principal = current.principal
    return current.phase === "authenticated" && principal && !principal.must_change_password
      ? `${principal.user_id}:${principal.employee_id}:${principal.system_roles.join(",")}`
      : undefined
  }
  function cancel() {
    ++generation
    controller?.abort()
    onState({ openingId: null, message: "" })
  }
  return {
    cancel,
    async open(item: ServerNotification) {
      const owner = identity()
      const employeeId = session.getSnapshot().principal?.employee_id
      if (!owner || !employeeId || session.getSnapshot().busy) return
      controller?.abort()
      controller = new AbortController()
      const signal = controller.signal
      const current = ++generation
      const active = () => current === generation && owner === identity() && !signal.aborted
      onState({ openingId: item.notification_id, message: "" })
      let destination: string
      try {
        destination = notificationDestination(item)
        if (item.notification_type === "PROFILE_CHANGE_SUBMITTED")
          await reviewApi.detail(item.resource_id, signal)
        else await selfApi.read(employeeId, item.resource_id, signal)
      } catch (error) {
        if (!active()) return
        onState({
          openingId: null,
          message:
            error instanceof AuthError && (error.status === 403 || error.status === 404)
              ? "Không thể mở đề nghị này theo quyền hiện hành. Bạn vẫn có thể đánh dấu thông báo của mình đã đọc."
              : error instanceof AuthError && error.status === 401
                ? "Phiên không còn hiệu lực. Đang kiểm tra lại với máy chủ."
                : "Chưa mở được đề nghị. Hãy thử lại khi kết nối hoặc dịch vụ sẵn sàng.",
        })
        if (error instanceof AuthError && (error.status === 401 || error.status === 403))
          await session.refresh()
        return
      }
      if (!active()) return
      onState({ openingId: null, message: "" })
      // Không giữ DTO chi tiết/PII trong khung; màn đích tải theo quyền hiện hành lần nữa.
      onOpen(destination)
      if (item.read_at === null) await markRead(item.notification_id)
    },
  }
}
