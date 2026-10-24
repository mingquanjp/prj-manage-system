import type { TaskStatus, Project } from "./workspace.ts"

export const statusLabels: Record<TaskStatus, string> = {
  PENDING_ASSIGNMENT: "Chờ phân công",
  ASSIGNED: "Đã phân công",
  IN_PROGRESS: "Đang thực hiện",
  PAUSED: "Tạm dừng",
  PENDING_REVIEW: "Chờ nghiệm thu",
  DONE: "Đã nghiệm thu",
  CLOSED: "Đã đóng",
  CANCELLED: "Đã hủy",
}

export const levels = ["", "Cơ bản", "Trung cấp", "Nâng cao", "Chuyên gia"]

export const priorities: Project["priority"][] = ["Thấp", "Trung bình", "Cao", "Khẩn cấp"]
