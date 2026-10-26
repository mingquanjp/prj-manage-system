import { useState } from "react"
import { Pencil1Icon, PlusIcon, TrashIcon } from "@radix-ui/react-icons"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { EmptyState, SearchInput } from "@/components/shared"
import { isHr } from "@/demo/identity"
import { organizationReferences } from "@/demo/organization"
import { searchKey } from "@/demo/personnel"
import type { EmployeeProfile } from "@/models/profile"
import {
  organizationLabels,
  type OrganizationCommand,
  type OrganizationEntry,
  type OrganizationKind,
  type OrganizationState,
} from "@/models/organization"
import { CatalogEntryForm } from "./CatalogEntryForm"

export function OrganizationCatalogs({
  actorId,
  organization,
  profiles,
  onChange,
}: {
  actorId: string
  organization: OrganizationState
  profiles: EmployeeProfile[]
  onChange: (command: OrganizationCommand) => void
}) {
  const [kind, setKind] = useState<OrganizationKind>("departments")
  const [query, setQuery] = useState("")
  const [editor, setEditor] = useState<{ kind: OrganizationKind; entry?: OrganizationEntry }>()
  const [removing, setRemoving] = useState<{ kind: OrganizationKind; entry: OrganizationEntry }>()
  const [error, setError] = useState("")
  if (!isHr(actorId))
    return (
      <EmptyState
        title="Bạn không có quyền quản lý danh mục"
        detail="Chức năng phòng ban và vị trí công việc dành cho HR."
      />
    )
  const entries = organization[kind].filter((entry) =>
    searchKey(entry.name).includes(searchKey(query)),
  )
  const references = removing
    ? organizationReferences(profiles, removing.kind, removing.entry.id)
    : []
  return (
    <div className="max-w-5xl">
      <div className="mb-6 flex flex-wrap items-start justify-between gap-4">
        <div>
          <p className="eyebrow">QUẢN TRỊ · HR</p>
          <h1 className="mt-2 text-2xl font-semibold text-balance">Phòng ban và vị trí</h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Danh mục dùng chung cho thông tin công tác của nhân viên.
          </p>
        </div>
        <Button onClick={() => setEditor({ kind })}>
          <PlusIcon />
          {kind === "departments" ? "Thêm phòng ban" : "Thêm vị trí"}
        </Button>
      </div>
      <div className="mb-5 flex flex-wrap items-end justify-between gap-4">
        <div className="flex gap-1" role="group" aria-label="Chọn danh mục tổ chức">
          {(["departments", "positions"] as const).map((value) => (
            <button
              key={value}
              type="button"
              className={`filter-chip ${kind === value ? "filter-active" : ""}`}
              aria-pressed={kind === value}
              onClick={() => {
                setKind(value)
                setQuery("")
              }}
            >
              {organizationLabels[value]}
            </button>
          ))}
        </div>
        <div className="w-full sm:w-80">
          <SearchInput
            value={query}
            onChange={setQuery}
            placeholder={kind === "departments" ? "Tìm phòng ban…" : "Tìm vị trí công việc…"}
          />
        </div>
      </div>
      <div className="overflow-x-auto rounded-xl border border-border bg-white">
        <table className="w-full min-w-[460px] text-left text-sm">
          <thead className="table-heading">
            <tr>
              <th scope="col" className="px-5 py-3">
                {kind === "departments" ? "Tên phòng ban" : "Tên vị trí công việc"}
              </th>
              <th scope="col" className="px-5 py-3 text-right">
                Hồ sơ sử dụng
              </th>
              <th scope="col" className="w-28 px-5 py-3 text-right">
                Thao tác
              </th>
            </tr>
          </thead>
          <tbody>
            {entries.map((entry) => (
              <tr key={entry.id} className="border-t border-border">
                <th
                  scope="row"
                  className="max-w-md whitespace-pre-wrap break-words px-5 py-3 font-medium"
                >
                  {entry.name}
                </th>
                <td className="px-5 py-3 text-right tabular-nums">
                  {organizationReferences(profiles, kind, entry.id).length}
                </td>
                <td className="px-4 py-2">
                  <div className="flex justify-end gap-1">
                    <Button
                      variant="ghost"
                      size="icon"
                      title={`Sửa ${entry.name}`}
                      aria-label={`Sửa ${entry.name}`}
                      onClick={() => setEditor({ kind, entry })}
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
                        setRemoving({ kind, entry })
                        setError("")
                      }}
                    >
                      <TrashIcon />
                    </Button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {!entries.length && (
          <EmptyState
            title={query ? "Không tìm thấy kết quả" : "Danh mục chưa có dữ liệu"}
            detail={
              query
                ? "Thử từ khóa khác hoặc xóa nội dung tìm kiếm."
                : "Chọn nút thêm để tạo mục đầu tiên."
            }
          />
        )}
      </div>
      <p role="status" className="mt-3 text-xs text-muted-foreground">
        {entries.length} / {organization[kind].length} mục hiển thị
      </p>
      {editor && (
        <CatalogEntryForm
          kind={editor.kind}
          entry={editor.entry}
          onSave={onChange}
          onClose={() => setEditor(undefined)}
        />
      )}
      <Dialog
        open={Boolean(removing)}
        onOpenChange={(open) => {
          if (!open) setRemoving(undefined)
        }}
        title={references.length ? "Chưa thể xóa mục này" : "Xóa khỏi danh mục?"}
        description={
          removing
            ? `${organizationLabels[removing.kind]}: ${removing.entry.name}`
            : "Kiểm tra mục trước khi xóa."
        }
      >
        {references.length ? (
          <>
            <p className="text-sm leading-6">
              Có {references.length} hồ sơ đang sử dụng mục này. HR cần cập nhật các hồ sơ liên quan
              trước khi xóa.
            </p>
            <ul className="mt-3 max-h-48 space-y-2 overflow-y-auto rounded-md bg-accent p-3 text-sm">
              {references.map((profile) => (
                <li key={profile.personId}>
                  {profile.fullname} · {profile.employeeCode}
                </li>
              ))}
            </ul>
            <div className="mt-5 flex justify-end">
              <Button variant="outline" onClick={() => setRemoving(undefined)}>
                Đã hiểu
              </Button>
            </div>
          </>
        ) : (
          <>
            <p className="text-sm leading-6">
              Mục chưa được sử dụng sẽ bị xóa. Thao tác này không thay đổi hồ sơ nhân viên.
            </p>
            {error && (
              <p role="alert" className="mt-3 text-sm text-warning">
                {error}
              </p>
            )}
            <div className="mt-5 flex justify-end gap-2">
              <Button variant="outline" onClick={() => setRemoving(undefined)}>
                Giữ lại
              </Button>
              <Button
                className="bg-warning text-white hover:bg-warning/90"
                onClick={() => {
                  if (!removing) return
                  try {
                    onChange({
                      action: "delete",
                      kind: removing.kind,
                      id: removing.entry.id,
                      beforeName: removing.entry.name,
                    })
                    setRemoving(undefined)
                  } catch (err) {
                    setError(err instanceof Error ? err.message : "Chưa thể xóa mục này.")
                  }
                }}
              >
                Xóa mục
              </Button>
            </div>
          </>
        )}
      </Dialog>
    </div>
  )
}
