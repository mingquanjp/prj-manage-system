import type { Person, Project, Task } from "../models/workspace.ts"
import { canViewProject, isHr, isDirector } from "../demo/identity.ts"
import { personalTasks } from "../demo/execution.ts"
import { demoUsers, people } from "../demo/fixtures.ts"
import { canReviewProfiles } from "../demo/profiles.ts"
import { canLookupPersonnel } from "../demo/personnel.ts"

export function selectWorkspaceNavigation(
  projects: Project[],
  tasks: Task[],
  actorId: string,
  pathname: string,
  directory: Person[] = people,
) {
  const active = Boolean(directory.find((p) => p.id === actorId)?.active)
  const ownTasks = active ? personalTasks(projects, tasks, actorId) : []
  const personalView = pathname === "/my-work"
  const personnelView = pathname === "/personnel"
  const profileView = pathname === "/my-profile"
  const profileReviewView = pathname === "/profile-requests"
  const employeeProfilesView = pathname === "/employee-profiles"
  const skillReviewView = pathname === "/skill-requests"
  const skillCatalogView = pathname === "/skill-catalog"
  const canManageSkills = active && isDirector(actorId)
  const organizationView = pathname === "/organization"
  const canManageOrganization = active && isHr(actorId)
  const canReviewProfile = active && canReviewProfiles(actorId)
  const canLookup = active && canLookupPersonnel(projects, actorId)
  const visibleProjects = projects.filter((p) => active && canViewProject(p, actorId))
  const currentProject = projects.find(
    (p) => active && pathname === `/projects/${p.id}` && canViewProject(p, actorId),
  )

  return {
    skillReviewView,
    skillCatalogView,
    canManageSkills,
    ownTasks,
    personalView,
    personnelView,
    profileView,
    profileReviewView,
    organizationView,
    employeeProfilesView,
    canManageOrganization,
    canReviewProfile,
    canLookup,
    visibleProjects,
    currentProject,
  }
}
export type WorkspaceNavigation = ReturnType<typeof selectWorkspaceNavigation>

export function demoIdentityOptions(
  projects: Project[],
  currentProject?: Project,
  directory: Person[] = people,
) {
  // Thứ tự trình diễn; không dùng để cấp quyền hoặc đổi trách nhiệm dự án.
  const demoOrder = ["thanh", "khoa", "dung", "ha", "hr", "cuong", "an", "binh"]
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
        name: directory.find((person) => person.id === u.personId)?.name,
        responsibility: labels.length ? labels.join("; ") : "Nhân viên",
      }
    })

  return demoIdentities
}
