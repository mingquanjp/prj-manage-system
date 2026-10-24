import { getPerson } from "./identity.ts"
import { loadFor } from "./assignment.ts"
import { people } from "./fixtures.ts"
import type { Project, Task } from "../models/workspace.ts"

// Projection demo cho UI-13; không trả entity, dữ liệu liên hệ hoặc task gốc.
export type ProfessionalProfile = {
  id: string
  name: string
  initials: string
  position: string
  seniority: string
  skills: { name: string; level: number }[]
  currentLoad: number
}

export function canLookupPersonnel(projects: Project[], actorId: string) {
  return (
    Boolean(getPerson(actorId)?.active) &&
    projects.some((project) => project.pmId === actorId || project.ptcmId === actorId)
  )
}

export function searchKey(value: string) {
  return value
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .replace(/đ/g, "d")
    .replace(/Đ/g, "D")
    .toLowerCase()
    .trim()
}

export function personnelProfiles(
  projects: Project[],
  tasks: Task[],
  actorId: string,
  query = "",
): ProfessionalProfile[] {
  if (!canLookupPersonnel(projects, actorId)) return []
  const needle = searchKey(query)
  return people
    .filter((person) => person.active)
    .map((person) => ({
      id: person.id,
      name: person.name,
      initials: person.initials,
      position: person.position,
      seniority: person.seniority,
      skills: Object.entries(person.skills).map(([name, level]) => ({ name, level })),
      currentLoad: loadFor(person.id, tasks),
    }))
    .filter((person) =>
      searchKey(
        [
          person.name,
          person.position,
          person.seniority,
          ...person.skills.map((skill) => skill.name),
        ].join(" "),
      ).includes(needle),
    )
}
