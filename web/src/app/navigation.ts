import { defaultIdentity, type IdentitySource } from "../demo/identity.ts"
import type { Person, Project, Task } from "../models/workspace.ts"
import { canViewProject, isHr, isDirector, isAdmin } from "../demo/identity.ts"
import { personalTasks } from "../demo/execution.ts"
import { systemRoleLabels } from "../models/accounts.ts"
import { people } from "../demo/fixtures.ts"
import { canReviewProfiles } from "../demo/profiles.ts"
import { canLookupPersonnel } from "../demo/personnel.ts"

export function selectWorkspaceNavigation(
  projects: Project[],
  tasks: Task[],
  actorId: string,
  pathname: string,
  directory: Person[] = people,
  identity: IdentitySource = defaultIdentity,
) {
  const accountsView = pathname === "/accounts"
  const canManageAccounts = isAdmin(actorId, identity)
  const active = Boolean(directory.find((p) => p.id === actorId)?.active)
  const ownTasks = active ? personalTasks(projects, tasks, actorId, identity) : []
  const personalView = pathname === "/my-work"
  const personnelView = pathname === "/personnel"
  const profileView = pathname === "/my-profile"
  const profileReviewView = pathname === "/profile-requests"
  const employeeProfilesView = pathname === "/employee-profiles"
  const skillReviewView = pathname === "/skill-requests"
  const skillCatalogView = pathname === "/skill-catalog"
  const canManageSkills = active && isDirector(actorId, identity)
  const organizationView = pathname === "/organization"
  const canManageOrganization = active && isHr(actorId, identity)
  const canReviewProfile = active && canReviewProfiles(actorId, identity)
  const canLookup = active && canLookupPersonnel(projects, actorId, identity)
  const visibleProjects = projects.filter((p) => active && canViewProject(p, actorId, identity))
  const currentProject = projects.find(
    (p) => active && pathname === `/projects/${p.id}` && canViewProject(p, actorId, identity),
  )

  return {
    accountsView,
    canManageAccounts,
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
  identity: IdentitySource = defaultIdentity,
) {
  // Thứ tự trình diễn; không dùng để cấp quyền hoặc đổi trách nhiệm dự án.
  const demoOrder = ["thanh", "khoa", "dung", "ha", "hr", "cuong", "an", "binh"]
  const demoIdentities = [...identity.accounts]
    .sort(
      (a, b) =>
        (demoOrder.includes(a.personId) ? demoOrder.indexOf(a.personId) : demoOrder.length) -
        (demoOrder.includes(b.personId) ? demoOrder.indexOf(b.personId) : demoOrder.length),
    )
    .map((u) => {
      const responsibilities = (currentProject ? [currentProject] : projects).flatMap((project) => {
        const roles = [
          ...(project.pmId === u.personId ? ["PM"] : []),
          ...(project.ptcmId === u.personId ? ["PTCM"] : []),
        ]
        return roles.length ? [`${roles.join(" + ")} (${project.code})`] : []
      })
      const labels = [
        ...(u.status === "LOCKED" ? ["Đã khóa"] : []),
        ...u.systemRoles.map((role) => systemRoleLabels[role]),
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
