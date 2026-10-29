import type { Task } from "../../models/workspace.ts"

// Phạm vi là các dòng của bảng gọi: toàn bộ tập lọc hoặc một nhóm cụ thể.
export function taskSelection(rows: Task[], selected: string[]) {
  const eligibleIds = rows
    .filter((task) => task.status === "PENDING_ASSIGNMENT")
    .map((task) => task.id)
  const count = eligibleIds.filter((id) => selected.includes(id)).length
  return {
    eligibleIds,
    hasAll: eligibleIds.length > 0 && count === eligibleIds.length,
    isPartial: count > 0 && count < eligibleIds.length,
  }
}
export function toggleTaskSelection(rows: Task[], selected: string[]) {
  const { eligibleIds, hasAll } = taskSelection(rows, selected)
  return hasAll
    ? selected.filter((id) => !eligibleIds.includes(id))
    : [...new Set([...selected, ...eligibleIds])]
}
