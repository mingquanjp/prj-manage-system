import { defaultIdentity, type IdentitySource } from "./identity.ts"
import type { ProfileState } from "../models/profile.ts"
import type { Project, Task } from "../models/workspace.ts"
import { isHr } from "./identity.ts"

export type DeactivationBlocker = { kind: "task" | "responsibility" | "admin"; label: string }

// Chỉ trả nhận diện tối thiểu để HR biết việc cần xử lý, không cấp đọc task/dự án.
export function employeeDeactivationBlockers(
  personId: string,
  projects: Project[],
  tasks: Task[],
  profiles: ProfileState["profiles"],
  identity: IdentitySource = defaultIdentity,
): DeactivationBlocker[] {
  const blockers: DeactivationBlocker[] = []
  for (const task of tasks) {
    if (
      task.assigneeId === personId &&
      ["ASSIGNED", "IN_PROGRESS", "PAUSED", "PENDING_REVIEW"].includes(task.status)
    )
      blockers.push({ kind: "task", label: `Công việc ${task.code} còn được giao.` })
  }
  for (const project of projects) {
    if (!["PLANNING", "IN_PROGRESS", "COMPLETED"].includes(project.status)) continue
    const roles = [
      ...(project.pmId === personId ? ["PM"] : []),
      ...(project.ptcmId === personId ? ["PTCM"] : []),
    ]
    if (roles.length)
      blockers.push({
        kind: "responsibility",
        label: `Trách nhiệm ${roles.join(" + ")} tại ${project.code} chưa kết thúc.`,
      })
  }
  const admins = identity.accounts.filter(
    (user) =>
      user.status === "ACTIVE" &&
      user.systemRoles.includes("ADMIN") &&
      profiles.some((p) => p.personId === user.personId && p.employmentStatus === "ACTIVE"),
  )
  if (admins.length === 1 && admins[0].personId === personId)
    blockers.push({ kind: "admin", label: "Đây là hồ sơ Quản trị viên hoạt động cuối cùng." })
  return blockers
}

export function deactivateEmployeeProfile(
  state: ProfileState,
  projects: Project[],
  tasks: Task[],
  actorId: string,
  personId: string,
  id: string,
  at: string,
  identity: IdentitySource = defaultIdentity,
): ProfileState {
  if (
    !isHr(actorId, identity) ||
    state.profiles.find((p) => p.personId === actorId)?.employmentStatus !== "ACTIVE"
  )
    throw new Error("Chỉ HR đang hoạt động được vô hiệu hóa hồ sơ nhân sự.")
  const profile = state.profiles.find((p) => p.personId === personId)
  if (!profile) throw new Error("Không tìm thấy hồ sơ nhân viên.")
  if (profile.employmentStatus !== "ACTIVE") throw new Error("Hồ sơ đã được vô hiệu hóa.")
  const blockers = employeeDeactivationBlockers(personId, projects, tasks, state.profiles, identity)
  if (blockers.length)
    throw new Error(`Chưa thể vô hiệu hóa. ${blockers.map((b) => b.label).join(" ")}`)
  return {
    ...state,
    profiles: state.profiles.map((p) =>
      p.personId === personId ? { ...p, employmentStatus: "INACTIVE" } : p,
    ),
    deactivations: [...state.deactivations, { id, employeeId: personId, actorId, at }],
  }
}
