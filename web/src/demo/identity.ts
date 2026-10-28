import { demoUsers, people } from "./fixtures.ts"
import type { Project, Person, SystemRole } from "../models/workspace.ts"

export type IdentitySource = {
  people: Person[]
  accounts: { personId: string; status: "ACTIVE" | "LOCKED"; systemRoles: SystemRole[] }[]
}
// Chỉ dành kiểm tra cũ/fixture; app truyền nguồn hiện hành, không sửa singleton này.
export const defaultIdentity: IdentitySource = {
  people,
  accounts: demoUsers.map((user) => ({ ...user, status: "ACTIVE" })),
}
export function getPerson(id?: string, identity: IdentitySource = defaultIdentity) {
  const person = identity.people.find((p) => p.id === id)
  if (!person) return undefined
  const account = identity.accounts.find((a) => a.personId === id)
  return { ...person, active: person.active && account?.status === "ACTIVE" }
}
export const isDirector = (actorId: string, identity: IdentitySource = defaultIdentity) =>
  Boolean(getPerson(actorId, identity)?.active) &&
  identity.accounts.some((u) => u.personId === actorId && u.systemRoles.includes("DIRECTOR"))
export const isHr = (actorId: string, identity: IdentitySource = defaultIdentity) =>
  Boolean(getPerson(actorId, identity)?.active) &&
  identity.accounts.some((u) => u.personId === actorId && u.systemRoles.includes("HR"))
export const isAdmin = (actorId: string, identity: IdentitySource = defaultIdentity) =>
  Boolean(getPerson(actorId, identity)?.active) &&
  identity.accounts.some((u) => u.personId === actorId && u.systemRoles.includes("ADMIN"))
export const canViewProject = (
  project: Project,
  actorId: string,
  identity: IdentitySource = defaultIdentity,
) =>
  Boolean(getPerson(actorId, identity)?.active) &&
  (isDirector(actorId, identity) || project.pmId === actorId || project.ptcmId === actorId)
export function canAssign(project: Project, actorId: string) {
  return actorId === project.ptcmId && ["PLANNING", "IN_PROGRESS"].includes(project.status)
}
export function isEligible(project: Project, person: Person) {
  return person.active && project.ptcmId !== person.id
}
