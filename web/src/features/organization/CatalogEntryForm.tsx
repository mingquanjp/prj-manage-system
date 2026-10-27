import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import type {
  OrganizationCommand,
  OrganizationEntry,
  OrganizationKind,
} from "@/models/organization"

export function CatalogEntryForm({
  kind,
  entry,
  onSave,
  onClose,
}: {
  kind: OrganizationKind
  entry?: OrganizationEntry
  onSave: (command: OrganizationCommand) => void
  onClose: () => void
}) {
  const [name, setName] = useState(entry?.name ?? "")
  const [error, setError] = useState("")
  const [discard, setDiscard] = useState(false)
  const label =
    kind === "departments" ? "phòng ban" : kind === "positions" ? "vị trí công việc" : "cấp bậc"
  const dirty = name !== (entry?.name ?? "")
  const requestClose = () => (dirty ? setDiscard(true) : onClose())
  return (
    <>
      <Dialog
        open
        onOpenChange={(open) => {
          if (!open) requestClose()
        }}
        title={`${entry ? "Sửa" : "Thêm"} ${label}`}
        description={
          entry
            ? "Tên mới sẽ hiển thị tại các hồ sơ đang sử dụng mục này."
            : "Mục mới dùng chung cho thông tin công tác của nhân viên."
        }
      >
        <form
          className="space-y-5"
          onSubmit={(event) => {
            event.preventDefault()
            try {
              onSave(
                entry
                  ? { action: "update", kind, id: entry.id, beforeName: entry.name, name }
                  : { action: "create", kind, name },
              )
              onClose()
            } catch (err) {
              setError(err instanceof Error ? err.message : "Chưa thể lưu danh mục.")
            }
          }}
        >
          <div>
            <label className="field-label mb-2 block" htmlFor="organization-name">
              Tên {label} <span aria-hidden="true">*</span>
            </label>
            <input
              id="organization-name"
              className="field workspace-field w-full"
              value={name}
              onChange={(event) => {
                setName(event.target.value)
                setError("")
              }}
              maxLength={100}
              aria-required="true"
              aria-invalid={Boolean(error)}
              aria-describedby={error ? "organization-error" : "organization-name-hint"}
              autoFocus
            />
            <p id="organization-name-hint" className="mt-2 text-xs text-muted-foreground">
              Tối đa 100 ký tự, không trùng tên trong cùng danh mục.
            </p>
            {error && (
              <p id="organization-error" role="alert" className="mt-2 text-sm text-warning">
                {error}
              </p>
            )}
          </div>
          <div className="flex justify-end gap-2">
            <Button variant="outline" onClick={requestClose}>
              Hủy
            </Button>
            <Button type="submit" disabled={Boolean(entry) && name.trim() === entry?.name}>
              Lưu{" "}
              {kind === "departments" ? "phòng ban" : kind === "positions" ? "vị trí" : "cấp bậc"}
            </Button>
          </div>
        </form>
      </Dialog>
      <Dialog
        open={discard}
        onOpenChange={setDiscard}
        title="Bỏ thay đổi chưa lưu?"
        description="Danh mục vẫn giữ nguyên nếu bạn bỏ nội dung vừa nhập."
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
