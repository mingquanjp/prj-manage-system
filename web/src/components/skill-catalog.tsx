import { createContext, useContext, type ReactNode } from "react"
import type { SkillCatalogState } from "@/models/skill-catalog"
import { catalogSkillName } from "@/demo/skill-catalog"

const SkillCatalogContext = createContext<SkillCatalogState | null>(null)
export function SkillCatalogProvider({
  catalog,
  children,
}: {
  catalog: SkillCatalogState
  children: ReactNode
}) {
  return <SkillCatalogContext.Provider value={catalog}>{children}</SkillCatalogContext.Provider>
}
export function useSkillCatalog() {
  const catalog = useContext(SkillCatalogContext)
  if (!catalog) throw new Error("Thiếu nguồn danh mục kỹ năng của workspace.")
  return { catalog, skillName: (id: string) => catalogSkillName(catalog, id) }
}
