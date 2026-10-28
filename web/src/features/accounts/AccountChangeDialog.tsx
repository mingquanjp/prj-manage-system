import { useState } from "react"
import { Dialog } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import type { DemoAccount, AccountCommand } from "@/models/accounts"
import type { Project, Task } from "@/models/workspace"
import { accountOpenWork } from "@/demo/accounts"
import { RoleChoices } from "./RoleChoices"

export function AccountChangeDialog({
  account,
  action,
  fullname,
  isEmployeeActive,
  projects,
  tasks,
  onSave,
  onClose,
}: {
  account: DemoAccount
  action: "roles" | "lock" | "unlock"
  fullname: string
  isEmployeeActive: boolean
  projects: Project[]
  tasks: Task[]
  onSave: (command: AccountCommand) => void
  onClose: () => void
}) {
  const [roles, setRoles] = useState([...account.systemRoles])
  const [error, setError] = useState("")
  const [discard, setDiscard] = useState(false)
  const dirty =
    action === "roles" &&
    JSON.stringify([...roles].sort()) !== JSON.stringify([...account.systemRoles].sort())
  const requestClose = () => (dirty ? setDiscard(true) : onClose())
  const openWork = action === "lock" ? accountOpenWork(account.personId, projects, tasks) : []
  const title =
    action === "roles"
      ? "Quản lý quyền hệ thống"
      : action === "lock"
        ? "Khóa tài khoản?"
        : "Mở khóa tài khoản?"
  return (
    <>
      <Dialog
        open
        alert={action !== "roles"}
        onOpenChange={(open) => {
          if (!open) requestClose()
        }}
        title={title}
        description={`${account.username} · ${fullname}`}
      >
        <form
          className="space-y-4"
          onSubmit={(event) => {
            event.preventDefault()
            setError("")
            try {
              const base = {
                accountId: account.id,
                before: { status: account.status, systemRoles: [...account.systemRoles] },
              }
              onSave(action === "roles" ? { ...base, action, roles } : { ...base, action })
              onClose()
            } catch (err) {
              setError(err instanceof Error ? err.message : "Chưa thể cập nhật tài khoản.")
            }
          }}
        >
          {action === "roles" ? (
            <>
              <RoleChoices value={roles} onChange={setRoles} />
              <p className="text-xs leading-5 text-muted-foreground">
                Thay quyền có hiệu lực trong phiên demo, không tự đổi trách nhiệm PM/PTCM. Admin
                không mặc nhiên có quyền Quản lý nhân sự hoặc Giám đốc.
              </p>
            </>
          ) : (
            <>
              <p className="text-sm leading-6">
                {action === "lock"
                  ? "Tài khoản không thể sử dụng nghiệp vụ sau khi khóa. Hồ sơ, dữ liệu và lịch sử được giữ; thao tác này không tự chuyển công việc hoặc trách nhiệm dự án."
                  : "Mở khóa tài khoản không thay đổi trạng thái hồ sơ nhân viên hoặc tự gán lại trách nhiệm."}
              </p>
              {!isEmployeeActive && (
                <p className="rounded-md bg-accent p-3 text-sm">
                  Hồ sơ nhân viên đã vô hiệu hóa. Mở khóa tài khoản vẫn không cho phép sử dụng
                  nghiệp vụ.
                </p>
              )}
              {openWork.length > 0 && (
                <section className="rounded-lg border border-warning/30 p-3">
                  <h3 className="field-label text-warning">Còn công việc/trách nhiệm cần xử lý</h3>
                  <ul className="mt-2 list-disc space-y-2 pl-4 text-xs leading-5">
                    {openWork.map((text) => (
                      <li key={text}>{text}</li>
                    ))}
                  </ul>
                  <p className="mt-2 text-xs leading-5">
                    Vẫn có thể khóa khẩn cấp. Admin không được giao lại hoặc thay PM từ màn hình
                    này.
                  </p>
                </section>
              )}
            </>
          )}
          <p className="text-xs leading-5 text-muted-foreground">
            Chưa có phiên đăng nhập thật. Thu hồi phiên khi đổi quyền/khóa sẽ được xử lý ở backend.
          </p>
          {error && (
            <p role="alert" className="text-sm text-warning">
              {error}
            </p>
          )}
          <div className="flex justify-end gap-2 border-t border-border pt-4">
            <Button variant="outline" onClick={requestClose}>
              Hủy
            </Button>
            <Button
              type="submit"
              disabled={action === "roles" && !dirty}
              className={action === "lock" ? "bg-warning text-white" : undefined}
            >
              {action === "roles" ? "Lưu quyền" : action === "lock" ? "Xác nhận khóa" : "Mở khóa"}
            </Button>
          </div>
        </form>
      </Dialog>
      <Dialog
        open={discard}
        onOpenChange={setDiscard}
        title="Bỏ thay đổi quyền?"
        description="Các quyền chưa lưu sẽ không được áp dụng."
      >
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={() => setDiscard(false)}>
            Tiếp tục
          </Button>
          <Button onClick={onClose}>Bỏ thay đổi</Button>
        </div>
      </Dialog>
    </>
  )
}
