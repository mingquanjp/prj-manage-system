import { useIdentity } from "@/components/identity"
import { useState } from "react"
import { Dialog } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import type { EmployeeProfile, ProfileState } from "@/models/profile"
import type { Project, Task } from "@/models/workspace"
import { employeeDeactivationBlockers } from "@/demo/employee-deactivation"

export function EmployeeDeactivationDialog({
  profile,
  profiles,
  projects,
  tasks,
  onConfirm,
  onClose,
}: {
  profile: EmployeeProfile
  profiles: ProfileState["profiles"]
  projects: Project[]
  tasks: Task[]
  onConfirm: (personId: string) => void
  onClose: () => void
}) {
  const identity = useIdentity()
  const [error, setError] = useState("")
  const blockers = employeeDeactivationBlockers(
    profile.personId,
    projects,
    tasks,
    profiles,
    identity,
  )
  return (
    <Dialog
      open
      alert
      onOpenChange={(open) => {
        if (!open) onClose()
      }}
      title={blockers.length ? "Chưa thể vô hiệu hóa hồ sơ" : "Vô hiệu hóa hồ sơ nhân viên?"}
      description={`${profile.employeeCode} · ${profile.fullname}`}
    >
      {blockers.length ? (
        <>
          <p className="text-pretty text-sm leading-6">
            Cần người có thẩm quyền xử lý các mục sau trước. Thao tác này không tự bàn giao công
            việc hoặc thay trách nhiệm dự án.
          </p>
          <ul className="mt-3 max-h-56 overflow-y-auto space-y-2 rounded-lg bg-accent p-3 text-sm">
            {blockers.map((blocker, index) => (
              <li key={`${blocker.kind}-${index}`}>{blocker.label}</li>
            ))}
          </ul>
          <p className="mt-3 text-xs leading-5 text-muted-foreground">
            HR chỉ xem các mục cần xử lý tại đây; quyền quản lý hồ sơ không cấp quyền mở nội dung
            công việc hoặc dự án.
          </p>
        </>
      ) : (
        <p className="text-pretty text-sm leading-6">
          Nhân viên sẽ không thể thực hiện nghiệp vụ hoặc được chọn cho phân công và trách nhiệm
          mới. Hồ sơ cùng lịch sử được giữ nguyên. Đây không phải thao tác khóa tài khoản.
        </p>
      )}
      {error && (
        <p role="alert" className="mt-3 text-sm text-warning">
          {error}
        </p>
      )}
      <div className="mt-5 flex flex-wrap justify-end gap-2">
        <Button variant="outline" onClick={onClose}>
          {blockers.length ? "Đã hiểu" : "Giữ hồ sơ hoạt động"}
        </Button>
        {!blockers.length && (
          <Button
            className="bg-warning text-white hover:bg-warning/90"
            onClick={() => {
              try {
                onConfirm(profile.personId)
                onClose()
              } catch (err) {
                setError(err instanceof Error ? err.message : "Chưa thể vô hiệu hóa hồ sơ.")
              }
            }}
          >
            Vô hiệu hóa hồ sơ
          </Button>
        )}
      </div>
    </Dialog>
  )
}
