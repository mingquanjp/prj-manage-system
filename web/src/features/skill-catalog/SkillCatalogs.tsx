import { useIdentity } from "@/components/identity"
import type { EmployeeSkillsState } from "@/models/employee-skills"
import { useState } from "react"
import { Pencil1Icon, PlusIcon, TrashIcon } from "@radix-ui/react-icons"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { EmptyState, SearchInput } from "@/components/shared"
import { cn } from "@/lib/utils"
import { isDirector } from "@/demo/identity"
import { searchKey } from "@/demo/personnel"
import { skillReferences } from "@/demo/skill-catalog"
import type { SkillCatalogCommand, SkillCatalogState } from "@/models/skill-catalog"
import type { Task } from "@/models/workspace"
import { SkillEntryForm, type SkillCatalogEditor } from "./SkillEntryForm"

type Removing = { kind: "categories" | "skills"; id: string; name: string; categoryId?: string }
export function SkillCatalogs({
  actorId,
  catalog,
  tasks,
  employeeSkills,
  onChange,
}: {
  actorId: string
  catalog: SkillCatalogState
  tasks: Task[]
  employeeSkills: EmployeeSkillsState
  onChange: (command: SkillCatalogCommand) => void
}) {
  const identity = useIdentity()

  const [kind, setKind] = useState<"categories" | "skills">("skills")
  const [query, setQuery] = useState("")
  const [category, setCategory] = useState("")
  const [editor, setEditor] = useState<SkillCatalogEditor>()
  const [removing, setRemoving] = useState<Removing>()
  const [error, setError] = useState("")
  if (!isDirector(actorId, identity))
    return (
      <EmptyState
        title="Bạn không có quyền quản lý danh mục"
        detail="Chức năng danh mục kỹ năng dành cho Giám đốc."
      />
    )
  const entries = catalog[kind].filter(
    (entry) =>
      searchKey(entry.name).includes(searchKey(query)) &&
      (kind !== "skills" ||
        !category ||
        catalog.skills.find((skill) => skill.id === entry.id)?.categoryId === category),
  )
  const references = removing
    ? skillReferences(catalog, tasks, removing.kind, removing.id, employeeSkills)
    : []
  return (
    <div className="w-full">
      <div className="mb-6 flex flex-wrap items-start justify-between gap-4">
        <div>
          <p className="eyebrow">QUẢN TRỊ · GIÁM ĐỐC</p>
          <h1 className="mt-2 text-2xl font-semibold text-balance">Danh mục kỹ năng</h1>
          <p className="mt-2 max-w-2xl text-pretty text-sm text-muted-foreground">
            Thống nhất tên và nhóm kỹ năng dùng trong công việc và hồ sơ năng lực. Quản lý danh mục
            không thay thế việc xét duyệt kỹ năng của nhân viên.
          </p>
        </div>
        <Button onClick={() => setEditor({ kind })}>
          <PlusIcon />
          {kind === "skills" ? "Thêm kỹ năng" : "Thêm nhóm kỹ năng"}
        </Button>
      </div>
      <div className="mb-5 flex flex-wrap items-end gap-4">
        <div className="mr-auto flex gap-1" role="group" aria-label="Chọn danh mục kỹ năng">
          {(["skills", "categories"] as const).map((value) => (
            <button
              key={value}
              type="button"
              className={cn("filter-chip", kind === value && "filter-active")}
              aria-pressed={kind === value}
              onClick={() => {
                setKind(value)
                setQuery("")
                setCategory("")
              }}
            >
              {value === "skills" ? "Kỹ năng" : "Nhóm kỹ năng"}
            </button>
          ))}
        </div>
        {kind === "skills" && (
          <div className="w-full sm:w-52">
            <label className="field-label mb-2 block" htmlFor="skill-category-filter">
              Nhóm kỹ năng
            </label>
            <select
              id="skill-category-filter"
              className="field w-full"
              value={category}
              onChange={(event) => setCategory(event.target.value)}
            >
              <option value="">Tất cả nhóm</option>
              {catalog.categories.map((entry) => (
                <option key={entry.id} value={entry.id}>
                  {entry.name}
                </option>
              ))}
            </select>
          </div>
        )}
        <div className="w-full sm:w-64">
          <SearchInput
            value={query}
            onChange={setQuery}
            placeholder={kind === "skills" ? "Tìm kỹ năng…" : "Tìm nhóm kỹ năng…"}
          />
        </div>
      </div>
      <div className="overflow-x-auto rounded-xl border border-border bg-white">
        <table className="w-full min-w-[520px] text-left text-sm">
          <thead className="table-heading">
            <tr>
              <th scope="col" className="px-5 py-3">
                {kind === "skills" ? "Tên kỹ năng" : "Tên nhóm kỹ năng"}
              </th>
              {kind === "skills" && (
                <th scope="col" className="px-5 py-3">
                  Nhóm kỹ năng
                </th>
              )}
              {kind === "categories" && (
                <th scope="col" className="px-5 py-3 text-right">
                  Số kỹ năng
                </th>
              )}
              <th scope="col" className="w-28 px-5 py-3 text-right">
                Thao tác
              </th>
            </tr>
          </thead>
          <tbody>
            {entries.map((entry) => {
              const skill =
                kind === "skills"
                  ? catalog.skills.find((skill) => skill.id === entry.id)
                  : undefined
              return (
                <tr key={entry.id} className="border-t border-border">
                  <th scope="row" className="max-w-sm break-words px-5 py-3 font-medium">
                    {entry.name}
                  </th>
                  {kind === "skills" ? (
                    <td className="max-w-xs break-words px-5 py-3 text-muted-foreground">
                      {
                        catalog.categories.find((category) => category.id === skill?.categoryId)
                          ?.name
                      }
                    </td>
                  ) : (
                    <td className="px-5 py-3 text-right tabular-nums">
                      {catalog.skills.filter((skill) => skill.categoryId === entry.id).length}
                    </td>
                  )}
                  <td className="px-4 py-2">
                    <div className="flex justify-end gap-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        title={`Sửa ${entry.name}`}
                        aria-label={`Sửa ${entry.name}`}
                        onClick={() =>
                          setEditor(
                            skill
                              ? { kind: "skills", entry: skill }
                              : { kind: "categories", entry },
                          )
                        }
                      >
                        <Pencil1Icon />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="text-warning"
                        title={`Xóa ${entry.name}`}
                        aria-label={`Xóa ${entry.name}`}
                        onClick={() => {
                          setRemoving({
                            kind,
                            id: entry.id,
                            name: entry.name,
                            categoryId: skill?.categoryId,
                          })
                          setError("")
                        }}
                      >
                        <TrashIcon />
                      </Button>
                    </div>
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
        {!entries.length && (
          <EmptyState
            title={query || category ? "Không tìm thấy kết quả" : "Danh mục chưa có dữ liệu"}
            detail={
              query || category
                ? "Thử đổi bộ lọc hoặc từ khóa tìm kiếm."
                : "Chọn nút thêm để tạo mục đầu tiên."
            }
          />
        )}
      </div>
      <p role="status" className="mt-3 text-xs text-muted-foreground tabular-nums">
        {entries.length} / {catalog[kind].length} mục hiển thị
      </p>
      {editor && (
        <SkillEntryForm
          editor={editor}
          catalog={catalog}
          onSave={onChange}
          onClose={() => setEditor(undefined)}
        />
      )}
      <Dialog
        open={Boolean(removing)}
        alert
        onOpenChange={(open) => {
          if (!open) setRemoving(undefined)
        }}
        title={references.length ? "Chưa thể xóa mục này" : "Xóa khỏi danh mục?"}
        description={removing?.name ?? "Kiểm tra mục trước khi xóa."}
      >
        {references.length ? (
          <>
            <p className="text-pretty text-sm leading-6">
              Mục này còn được sử dụng, kể cả trong dữ liệu đã kết thúc. Không xóa hoặc tự thay thế
              tham chiếu.
            </p>
            <ul className="mt-3 space-y-2 rounded-md bg-accent p-3 text-sm">
              {references.map((label) => (
                <li key={label}>{label}</li>
              ))}
            </ul>
            <p className="mt-3 text-xs text-muted-foreground">
              Chỉ hiển thị thống kê tham chiếu; quyền danh mục không cấp quyền đọc nội dung công
              việc hoặc hồ sơ nhân sự.
            </p>
          </>
        ) : (
          <p className="text-pretty text-sm leading-6">
            Mục chưa được sử dụng sẽ bị xóa khỏi danh mục trong phiên mô phỏng.
          </p>
        )}
        {error && (
          <p role="alert" className="mt-3 text-sm text-warning">
            {error}
          </p>
        )}
        <div className="mt-5 flex justify-end gap-2">
          <Button variant="outline" onClick={() => setRemoving(undefined)}>
            {references.length ? "Đã hiểu" : "Giữ lại"}
          </Button>
          {!references.length && (
            <Button
              className="bg-warning text-white hover:bg-warning/90"
              onClick={() => {
                if (!removing) return
                try {
                  onChange(
                    removing.kind === "skills"
                      ? {
                          action: "delete",
                          kind: "skills",
                          id: removing.id,
                          beforeName: removing.name,
                          beforeCategoryId: removing.categoryId!,
                        }
                      : {
                          action: "delete",
                          kind: "categories",
                          id: removing.id,
                          beforeName: removing.name,
                        },
                  )
                  setRemoving(undefined)
                } catch (err) {
                  setError(err instanceof Error ? err.message : "Chưa thể xóa mục này.")
                }
              }}
            >
              Xóa khỏi danh mục
            </Button>
          )}
        </div>
      </Dialog>
    </div>
  )
}
