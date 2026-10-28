import { useState } from "react"
import { Dialog } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { usePeople } from "@/components/people-directory"
import type { Project } from "@/models/workspace"
import type { PmChangeCommand } from "@/models/task-collaboration"

export function ChangePmDialog({
  project,
  onSave,
  onClose,
}: {
  project: Project
  onSave: (command: PmChangeCommand) => void
  onClose: () => void
}) {
  const { people, getPerson } = usePeople()
  const [beforePmId] = useState(project.pmId)
  const [pmId, setPmId] = useState("")
  const [error, setError] = useState("")
  const close = () => {
    if (!pmId || window.confirm("Bỏ lựa chọn PM mới chưa lưu?")) onClose()
  }
  return (
    <Dialog
      open
      title="Thay người quản lý dự án"
      description="Lưu một lần để thay PM. PTCM, công việc và các đóng góp hiện có được giữ nguyên."
      onOpenChange={(open) => !open && close()}
    >
      <form
        className="space-y-4"
        onSubmit={(event) => {
          event.preventDefault()
          try {
            onSave({ projectId: project.id, beforePmId, pmId })
            onClose()
          } catch (err) {
            setError((err as Error).message)
          }
        }}
      >
        <p className="text-sm">
          PM hiện tại: <strong>{getPerson(project.pmId)?.name}</strong>
        </p>
        <label className="block">
          <span className="field-label">PM mới</span>
          <select
            className="field mt-2 w-full"
            required
            value={pmId}
            onChange={(event) => {
              setPmId(event.target.value)
              setError("")
            }}
          >
            <option value="">Chọn nhân viên</option>
            {people
              .filter((person) => person.active && person.id !== project.pmId)
              .map((person) => (
                <option key={person.id} value={person.id}>
                  {person.name}
                  {person.id === project.ptcmId ? " · PTCM hiện tại" : ""}
                </option>
              ))}
          </select>
        </label>
        {error && (
          <p role="alert" className="text-sm text-destructive">
            {error}
          </p>
        )}
        <div className="flex justify-end gap-2">
          <Button type="button" variant="outline" onClick={close}>
            Hủy
          </Button>
          <Button type="submit" disabled={!pmId}>
            Lưu PM mới
          </Button>
        </div>
      </form>
    </Dialog>
  )
}
