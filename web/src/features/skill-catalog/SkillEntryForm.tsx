import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import type {
  CatalogSkill,
  SkillCategory,
  SkillCatalogCommand,
  SkillCatalogState,
} from "@/models/skill-catalog"

export type SkillCatalogEditor =
  { kind: "categories"; entry?: SkillCategory } | { kind: "skills"; entry?: CatalogSkill }
export function SkillEntryForm({
  editor,
  catalog,
  onSave,
  onClose,
}: {
  editor: SkillCatalogEditor
  catalog: SkillCatalogState
  onSave: (command: SkillCatalogCommand) => void
  onClose: () => void
}) {
  const [name, setName] = useState(editor.entry?.name ?? "")
  const [categoryId, setCategoryId] = useState(
    editor.kind === "skills" ? (editor.entry?.categoryId ?? "") : "",
  )
  const [error, setError] = useState("")
  const [discard, setDiscard] = useState(false)
  const initialCategory = editor.kind === "skills" ? (editor.entry?.categoryId ?? "") : ""
  const dirty = name !== (editor.entry?.name ?? "") || categoryId !== initialCategory
  const requestClose = () => (dirty ? setDiscard(true) : onClose())
  const label = editor.kind === "categories" ? "nhóm kỹ năng" : "kỹ năng"
  return (
    <>
      <Dialog
        open
        onOpenChange={(open) => {
          if (!open) requestClose()
        }}
        title={`${editor.entry ? "Sửa" : "Thêm"} ${label}`}
        description="Danh mục dùng chung cho yêu cầu công việc và năng lực chính thức."
      >
        <form
          className="space-y-5"
          noValidate
          onSubmit={(event) => {
            event.preventDefault()
            try {
              if (editor.kind === "categories")
                onSave(
                  editor.entry
                    ? {
                        action: "update",
                        kind: "categories",
                        id: editor.entry.id,
                        beforeName: editor.entry.name,
                        name,
                      }
                    : { action: "create", kind: "categories", name },
                )
              else
                onSave(
                  editor.entry
                    ? {
                        action: "update",
                        kind: "skills",
                        id: editor.entry.id,
                        beforeName: editor.entry.name,
                        beforeCategoryId: editor.entry.categoryId,
                        name,
                        categoryId,
                      }
                    : { action: "create", kind: "skills", name, categoryId },
                )
              onClose()
            } catch (err) {
              setError(err instanceof Error ? err.message : "Chưa thể lưu danh mục.")
            }
          }}
        >
          <div>
            <label className="field-label mb-2 block" htmlFor="skill-entry-name">
              Tên {label} *
            </label>
            <input
              id="skill-entry-name"
              className="field w-full"
              value={name}
              maxLength={100}
              aria-required="true"
              aria-describedby="skill-entry-hint"
              autoFocus
              onChange={(event) => {
                setName(event.target.value)
                setError("")
              }}
            />
            <p id="skill-entry-hint" className="mt-2 text-xs text-muted-foreground">
              Tối đa 100 ký tự, không trùng tên trong cùng danh mục.
            </p>
          </div>
          {editor.kind === "skills" && (
            <div>
              <label className="field-label mb-2 block" htmlFor="skill-entry-category">
                Nhóm kỹ năng *
              </label>
              <select
                id="skill-entry-category"
                className="field w-full"
                value={categoryId}
                aria-required="true"
                onChange={(event) => {
                  setCategoryId(event.target.value)
                  setError("")
                }}
              >
                <option value="">Chọn nhóm kỹ năng</option>
                {catalog.categories.map((entry) => (
                  <option key={entry.id} value={entry.id}>
                    {entry.name}
                  </option>
                ))}
              </select>
              {!catalog.categories.length && (
                <p className="mt-2 text-sm text-warning">
                  Thêm nhóm kỹ năng trước khi tạo kỹ năng.
                </p>
              )}
            </div>
          )}
          {editor.entry && (
            <p className="rounded-md bg-accent p-3 text-pretty text-sm leading-6">
              Chỉ chỉnh tên hoặc cách phân nhóm, không đổi nghĩa kỹ năng đã sử dụng. Nếu cần kỹ năng
              khác, hãy tạo mục mới. Thay đổi danh mục không công nhận thêm năng lực cho nhân viên.
            </p>
          )}
          {error && (
            <p role="alert" className="text-sm text-warning">
              {error}
            </p>
          )}
          <div className="flex justify-end gap-2">
            <Button variant="outline" onClick={requestClose}>
              Hủy
            </Button>
            <Button type="submit" disabled={Boolean(editor.entry) && !dirty}>
              Lưu {label}
            </Button>
          </div>
        </form>
      </Dialog>
      <Dialog
        open={discard}
        onOpenChange={setDiscard}
        title="Bỏ thay đổi chưa lưu?"
        description="Danh mục chưa thay đổi nếu bạn bỏ nội dung vừa nhập."
      >
        <div className="flex flex-wrap justify-end gap-2">
          <Button variant="outline" onClick={() => setDiscard(false)}>
            Tiếp tục chỉnh sửa
          </Button>
          <Button onClick={onClose}>Bỏ thay đổi</Button>
        </div>
      </Dialog>
    </>
  )
}
