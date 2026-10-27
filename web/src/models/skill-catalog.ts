// Danh mục trong phiên frontend; ID tham chiếu độc lập với tên hiển thị.
export type SkillCategory = { id: string; name: string }
export type CatalogSkill = { id: string; name: string; categoryId: string }
export type SkillCatalogState = { categories: SkillCategory[]; skills: CatalogSkill[] }
export type SkillCatalogCommand =
  | { action: "create"; kind: "categories"; name: string }
  | { action: "create"; kind: "skills"; name: string; categoryId: string }
  | { action: "update"; kind: "categories"; id: string; beforeName: string; name: string }
  | {
      action: "update"
      kind: "skills"
      id: string
      beforeName: string
      beforeCategoryId: string
      name: string
      categoryId: string
    }
  | { action: "delete"; kind: "categories"; id: string; beforeName: string }
  | { action: "delete"; kind: "skills"; id: string; beforeName: string; beforeCategoryId: string }
