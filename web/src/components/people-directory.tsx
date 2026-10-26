import { createContext, useContext, type ReactNode } from "react"
import type { Person } from "@/models/workspace"

const PeopleContext = createContext<Person[] | null>(null)

// Chỉ chia sẻ projection nghề nghiệp để tên chính thức đồng bộ; không giữ state/quyền riêng.
export function PeopleDirectory({ people, children }: { people: Person[]; children: ReactNode }) {
  return <PeopleContext.Provider value={people}>{children}</PeopleContext.Provider>
}

export function usePeople() {
  const people = useContext(PeopleContext)
  if (!people) throw new Error("Thiếu nguồn nhân sự của workspace.")
  return { people, getPerson: (id?: string) => people.find((person) => person.id === id) }
}
