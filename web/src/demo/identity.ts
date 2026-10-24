import { demoUsers, people } from "./fixtures.ts"
import type { Project, Person } from "../models/workspace.ts"

export const isDirector = (actorId: string) =>
  demoUsers.some((u) => u.personId === actorId && u.systemRoles.includes("DIRECTOR"))

export const canViewProject = (project: Project, actorId: string) =>
  isDirector(actorId) || project.pmId === actorId || project.ptcmId === actorId

export function getPerson(id?: string) {
  return people.find((p) => p.id === id)
}

export function canAssign(project: Project, actorId: string) {
  return actorId === project.ptcmId && ["PLANNING", "IN_PROGRESS"].includes(project.status)
}

export function isEligible(project: Project, person: Person) {
  return person.active && project.ptcmId !== person.id
}
