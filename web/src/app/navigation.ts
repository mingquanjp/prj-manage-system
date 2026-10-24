import type { Project, Task } from "../models/workspace.ts"
import { canViewProject, getPerson } from "../demo/identity.ts"
import { personalTasks } from "../demo/execution.ts"
import { demoUsers } from "../demo/fixtures.ts"
import { canLookupPersonnel } from "../demo/personnel.ts"

export function selectWorkspaceNavigation(
  projects: Project[],
  tasks: Task[],
  actorId: string,
  pathname: string,
) {
  const ownTasks = personalTasks(projects, tasks, actorId)
  const personalView = pathname === "/my-work"
  const personnelView = pathname === "/personnel"
  const canLookup = canLookupPersonnel(projects, actorId)
  const visibleProjects = projects.filter((p) => canViewProject(p, actorId))
  const currentProject = projects.find(
    (p) => pathname === `/projects/${p.id}` && canViewProject(p, actorId),
  )

  return { ownTasks, personalView, personnelView, canLookup, visibleProjects, currentProject }
}
export type WorkspaceNavigation = ReturnType<typeof selectWorkspaceNavigation>

export function demoIdentityOptions(projects: Project[], currentProject?: Project) {
  // Thứ tự trình diễn; không dùng để cấp quyền hoặc đổi trách nhiệm dự án.
  const demoOrder = ["thanh", "khoa", "dung", "ha", "cuong", "an", "binh"]
  const demoIdentities = [...demoUsers]
    .sort((a, b) => demoOrder.indexOf(a.personId) - demoOrder.indexOf(b.personId))
    .map((u) => {
      const responsibilities = (currentProject ? [currentProject] : projects).flatMap((project) => {
        const roles = [
          ...(project.pmId === u.personId ? ["PM"] : []),
          ...(project.ptcmId === u.personId ? ["PTCM"] : []),
        ]
        return roles.length ? [`${roles.join(" + ")} (${project.code})`] : []
      })
      const labels = [
        ...u.systemRoles.map(
          (role) => ({ ADMIN: "Quản trị viên", HR: "HR", DIRECTOR: "Giám đốc" })[role],
        ),
        ...responsibilities,
      ]
      return {
        id: u.personId,
        name: getPerson(u.personId)?.name,
        responsibility: labels.length ? labels.join("; ") : "Nhân viên",
      }
    })

  return demoIdentities
}
