import type { Project, Task } from "../models/workspace.ts"
import { canAssign, getPerson, isEligible } from "./identity.ts"
import { skillGaps } from "./assignment.ts"
import { rememberContributions, type WorkState } from "./effort.ts"
import type { DemoWorkspaceState } from "./workspace-state.ts"

export type ReassignmentCommand = {
  projectId: string
  taskId: string
  expectedAssignmentId: string
  personId: string
  reason: string
  overrideReason: string
}

export function reassignmentBlockReason(
  project: Project,
  task: Task,
  work: WorkState,
  actorId: string,
) {
  if (!canAssign(project, actorId) || !getPerson(actorId)?.active)
    return "Chỉ PTCM hiện hành được đổi người khi dự án còn lập kế hoạch hoặc đang triển khai."
  if (
    task.projectId !== project.id ||
    task.status !== "ASSIGNED" ||
    !task.currentAssignment ||
    task.currentAssignment.endedAt ||
    task.currentAssignment.employeeId !== task.assigneeId
  )
    return "Chỉ đổi người cho công việc đã phân công và chưa bắt đầu."
  // Không suy ra chưa thực hiện chỉ từ trạng thái nếu dữ liệu còn dấu vết thực tế.
  if (
    task.executionHistory?.some((event) =>
      [event.from, event.to].some((status) =>
        ["IN_PROGRESS", "PAUSED", "PENDING_REVIEW", "DONE", "CLOSED"].includes(status),
      ),
    ) ||
    work.logs.some((log) => log.taskId === task.id) ||
    work.submissions.some((submission) => submission.taskId === task.id) ||
    work.history.some((edit) => edit.before.taskId === task.id) ||
    work.requests.some((request) =>
      work.contributions.some(
        (entry) => entry.taskId === task.id && entry.assignment.id === request.assignmentId,
      ),
    )
  )
    return "Công việc đã có thực hiện; thay người giữa chừng chưa triển khai trong đợt này."
  return undefined
}

export function reassignUnstartedTask(
  state: DemoWorkspaceState,
  actorId: string,
  command: ReassignmentCommand,
  at: string,
  assignmentId: string,
) {
  const project = state.projects.find((p) => p.id === command.projectId)
  const task = state.tasks.find((t) => t.id === command.taskId)
  if (!project || !task) throw new Error("Không tìm thấy công việc hoặc dự án.")
  const block = reassignmentBlockReason(project, task, state.work, actorId)
  if (block) throw new Error(block)
  if (task.currentAssignment!.id !== command.expectedAssignmentId)
    throw new Error("Phân công đã thay đổi. Đóng hộp thoại và kiểm tra lại người thực hiện.")
  const person = getPerson(command.personId)
  if (!person || !isEligible(project, person))
    throw new Error("Người được chọn không đủ điều kiện nhận công việc.")
  if (person.id === task.assigneeId) throw new Error("Vui lòng chọn người thực hiện khác.")
  const reason = command.reason.trim()
  if (!reason) throw new Error("Vui lòng nhập lý do đổi người.")
  const overrideReason = command.overrideReason.trim()
  const gaps = skillGaps(task, person)
  if (gaps.length && !overrideReason) throw new Error("Vui lòng nhập lý do ngoại lệ kỹ năng.")
  const historyReason = `Đổi người: ${getPerson(task.assigneeId)?.name ?? task.assigneeId} → ${person.name}. ${reason}${gaps.length ? ` Ngoại lệ kỹ năng: ${overrideReason}` : ""}`
  const endedTask: Task = {
    ...task,
    currentAssignment: {
      ...task.currentAssignment!,
      endedAt: at,
      endedBy: actorId,
      endReason: historyReason,
    },
  }
  const nextTask: Task = {
    ...task,
    firstAssignedAt: task.firstAssignedAt ?? task.currentAssignment!.assignedAt,
    assigneeId: person.id,
    currentAssignment: { id: assignmentId, employeeId: person.id, assignedAt: at },
    executionHistory: [
      ...(task.executionHistory ?? []),
      { actorId, at, from: "ASSIGNED", to: "ASSIGNED", reason: historyReason },
    ],
  }
  const tasks = state.tasks.map((t) => (t.id === task.id ? nextTask : t))
  const work = rememberContributions(rememberContributions(state.work, [endedTask]), tasks)
  return {
    tasks,
    work,
    notice: `Đã chuyển ${task.code} cho ${person.name}. Người mới cần bấm Bắt đầu.`,
  }
}
