import { useState, type RefObject } from "react"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { AuthError } from "../session/auth-api"
import {
  catalogErrorMessage,
  trimCatalogName,
  validCatalogName,
  type CatalogItem,
} from "./catalog-api"

export type CatalogEditor =
  { action: "create" } | { action: "rename" | "delete"; entry: CatalogItem }
export function ServerCatalogDialog({
  editor,
  label,
  busy,
  onSave,
  onClose,
  onReload,
  returnFocus,
}: {
  editor: CatalogEditor
  label: string
  busy: boolean
  onSave: (name: string) => Promise<void>
  onClose: () => void
  onReload: () => void
  returnFocus: RefObject<HTMLElement | null>
}) {
  const [name, setName] = useState("entry" in editor ? editor.entry.name : "")
  const [error, setError] = useState("")
  const [blocked, setBlocked] = useState(false)
  const [saving, setSaving] = useState(false)
  const [discard, setDiscard] = useState(false)
  const deleting = editor.action === "delete"
  const dirty = !deleting && name !== ("entry" in editor ? editor.entry.name : "")
  const close = () => {
    if (!busy && !saving) {
      if (dirty && !blocked) setDiscard(true)
      else onClose()
    }
  }
  return (
    <>
      <Dialog
        open
        title={`${deleting ? "Xóa" : editor.action === "create" ? "Thêm" : "Sửa"} ${label.toLowerCase()}`}
        description={
          deleting
            ? `Bạn muốn xóa “${"entry" in editor ? editor.entry.name : ""}”? Mục còn được sử dụng sẽ không bị xóa.`
            : "Tên dùng chung cho thông tin công tác của nhân viên."
        }
        onOpenChange={(open) => {
          if (!open) close()
        }}
        onCloseAutoFocus={(event) => {
          event.preventDefault()
          if (returnFocus.current?.isConnected) return returnFocus.current.focus()
          document.getElementById("catalog-heading")?.focus()
        }}
      >
        <form
          className="space-y-5"
          onSubmit={async (event) => {
            event.preventDefault()
            if (blocked || busy || saving) return
            if (!deleting && !validCatalogName(name)) {
              setError("Nhập tên từ 1 đến 100 ký tự sau khi bỏ khoảng trắng đầu/cuối.")
              return
            }
            setSaving(true)
            setError("")
            try {
              await onSave(trimCatalogName(name))
              onClose()
            } catch (error) {
              setError(catalogErrorMessage(error, true))
              setBlocked(
                !(error instanceof AuthError) ||
                  error.status === 0 ||
                  error.status >= 500 ||
                  error.status === 404 ||
                  error.code === "STALE_DATA",
              )
            } finally {
              setSaving(false)
            }
          }}
        >
          {!deleting && (
            <div>
              <label className="field-label mb-2 block" htmlFor="server-catalog-name">
                Tên {label.toLowerCase()}
              </label>
              <input
                id="server-catalog-name"
                autoFocus
                className="field w-full"
                value={name}
                disabled={busy || saving || blocked}
                aria-required="true"
                aria-invalid={Boolean(error)}
                aria-describedby="catalog-name-hint catalog-form-error"
                onChange={(event) => {
                  setName(event.target.value)
                  setError("")
                }}
              />
              <p id="catalog-name-hint" className="mt-2 text-xs text-muted-foreground">
                Tối đa 100 ký tự, không trùng tên trong cùng danh mục.
              </p>
            </div>
          )}
          <p
            id="catalog-form-error"
            role={error ? "alert" : undefined}
            className="text-sm leading-6 text-warning"
          >
            {error}
          </p>
          <div className="flex flex-wrap justify-end gap-2">
            <Button variant="outline" disabled={busy || saving} onClick={close}>
              Hủy
            </Button>
            {blocked ? (
              <Button disabled={busy || saving} onClick={onReload}>
                Tải lại danh sách
              </Button>
            ) : (
              <Button
                type="submit"
                disabled={busy || saving}
                className={deleting ? "bg-warning text-white hover:bg-warning/90" : undefined}
              >
                {saving ? "Đang lưu…" : deleting ? "Xóa mục" : "Lưu"}
              </Button>
            )}
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
