import type { Task, Person, Draft, Project } from "../models/workspace.ts"
import { getPerson, isEligible, canAssign } from "./identity.ts"

export function skillGaps(task: Task, person: Person) {
  return task.requirements.filter((r) => r.mandatory && (person.skills[r.skill] ?? 0) < r.level)
}

export function loadFor(personId: string, tasks: Task[]) {
  return (
    (getPerson(personId)?.backgroundLoad ?? 0) +
    tasks
      .filter(
        (t) =>
          t.assigneeId === personId &&
          ["ASSIGNED", "IN_PROGRESS", "PAUSED", "PENDING_REVIEW"].includes(t.status),
      )
      .reduce((sum, t) => sum + t.estimate, 0)
  )
}

export function addedLoad(personId: string, draft: Draft, tasks: Task[]) {
  return Object.entries(draft)
    .filter(([, choice]) => choice.personId === personId)
    .reduce((sum, [id]) => sum + (tasks.find((t) => t.id === id)?.estimate ?? 0), 0)
}

// Dữ liệu minh họa bố cục, không phải công thức cuối hoặc xếp hạng thời gian thực.
const demoRanks = {
  Backend: ["an", "dung", "binh"],
  Frontend: ["cuong", "binh"],
}

export function recommendations(project: Project, task: Task) {
  const fixtureGroup = task.requirements.some((r) => r.skill === "React")
    ? "Frontend"
    : task.requirements.some((r) => r.skill === "Node.js")
      ? "Backend"
      : undefined
  const ids = fixtureGroup ? demoRanks[fixtureGroup] : []
  return ids
    .map((id) => getPerson(id)!)
    .filter((p) => isEligible(project, p) && skillGaps(task, p).length === 0)
    .slice(0, 3)
    .map((person, index) => ({
      person,
      rank: index + 1,
      score: [94.6, 87.2, 81.8][index],
    }))
}

export function validateDraft(
  project: Project,
  actorId: string,
  tasks: Task[],
  draft: Draft,
): string[] {
  if (!canAssign(project, actorId)) return ["Bạn không có trách nhiệm phân công tại dự án này."]
  if (!Object.keys(draft).length) return ["Chưa có lựa chọn trong bản nháp."]
  const errors: string[] = []
  for (const [id, choice] of Object.entries(draft)) {
    const task = tasks.find((t) => t.id === id && t.projectId === project.id)
    const person = getPerson(choice.personId)
    if (!task || task.status !== "PENDING_ASSIGNMENT") {
      errors.push("Có công việc không còn chờ phân công.")
      continue
    }
    if (!person || !isEligible(project, person)) {
      errors.push(`${task.code}: người được chọn không đủ điều kiện.`)
      continue
    }
    if (
      choice.mode === "recommended" &&
      !recommendations(project, task).some(
        (r) => r.person.id === person.id && r.rank === choice.rank,
      )
    )
      errors.push(`${task.code}: kết quả gợi ý không hợp lệ.`)
    if (
      ((choice.mode === "recommended" && choice.rank !== 1) ||
        (choice.mode === "manual" && choice.hadRun)) &&
      !choice.reason.trim()
    )
      errors.push(`${task.code}: bổ sung lý do lựa chọn.`)
    if (
      skillGaps(task, person).length &&
      (!choice.overrideReason.trim() || choice.mode !== "manual")
    )
      errors.push(`${task.code}: bổ sung lý do ngoại lệ kỹ năng.`)
  }
  return errors
}

export function commitDraft(
  project: Project,
  actorId: string,
  tasks: Task[],
  draft: Draft,
): Task[] {
  const errors = validateDraft(project, actorId, tasks, draft)
  if (errors.length) throw new Error(errors.join(" "))
  return tasks.map((task) =>
    draft[task.id]
      ? {
          ...task,
          status: "ASSIGNED",
          assigneeId: draft[task.id].personId,
          firstAssignedAt: task.firstAssignedAt ?? new Date().toISOString(),
          currentAssignment: {
            id: crypto.randomUUID(),
            employeeId: draft[task.id].personId,
            assignedAt: new Date().toISOString(),
          },
        }
      : task,
  )
}
