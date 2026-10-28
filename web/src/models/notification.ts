// Thông báo nội bộ trong phiên demo; không phải DTO hoặc hàng đợi phía server.
export type NotificationTarget = {
  kind: "task" | "correction" | "profile-request" | "skill-request"
  id: string
}
export type WorkspaceNotification = {
  id: string
  recipientId: string
  message: string
  at: string
  readAt?: string
  target: NotificationTarget
}
