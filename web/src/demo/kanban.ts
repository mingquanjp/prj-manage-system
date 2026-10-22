import { statusLabels, type Task, type TaskStatus } from "./model.ts"

// Chỉ nhóm trạng thái để trình bày; không phải luồng kéo thả/chuyển trạng thái.
export const projectColumns: {
  id: TaskStatus
  label: string
  statuses: TaskStatus[]
}[] = [
  {
    id: "PENDING_ASSIGNMENT",
    label: statusLabels.PENDING_ASSIGNMENT,
    statuses: ["PENDING_ASSIGNMENT"],
  },
  { id: "ASSIGNED", label: statusLabels.ASSIGNED, statuses: ["ASSIGNED"] },
  {
    id: "IN_PROGRESS",
    label: "Đang xử lý",
    statuses: ["IN_PROGRESS", "PAUSED"],
  },
  {
    id: "PENDING_REVIEW",
    label: statusLabels.PENDING_REVIEW,
    statuses: ["PENDING_REVIEW"],
  },
  { id: "DONE", label: statusLabels.DONE, statuses: ["DONE"] },
]
export function projectBoard(tasks: Task[]) {
  return projectColumns.map((column) => ({
    ...column,
    tasks: tasks.filter((t) => column.statuses.includes(t.status)),
  }))
}
