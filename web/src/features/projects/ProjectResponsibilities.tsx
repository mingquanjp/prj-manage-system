import { useState } from "react"
import { ChevronDownIcon, GearIcon } from "@radix-ui/react-icons"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { appointPtcm, ptcmBlockReason, startProject } from "@/demo/projects"
import { getPerson } from "@/demo/identity"
import { people } from "@/demo/fixtures"
import type { OutputConflict, Project, ProjectEvent, Task } from "@/models/workspace"

export function ProjectResponsibilities({
  project,
  actorId,
  tasks,
  conflicts,
  events,
  onChange,
}: {
  project: Project
  actorId: string
  tasks: Task[]
  conflicts: OutputConflict[]
  events: ProjectEvent[]
  onChange: (project: Project, action: string) => void
}) {
  const [editing, setEditing] = useState(false)
  const [candidate, setCandidate] = useState(project.ptcmId ?? "")
  const [error, setError] = useState("")
  const [starting, setStarting] = useState(false)
  const [managing, setManaging] = useState(false)
  const pm = project.pmId === actorId
  const editable = ["PLANNING", "IN_PROGRESS"].includes(project.status)
  const block = candidate ? ptcmBlockReason(project, candidate, tasks, conflicts) : undefined
  const hasAssigned = tasks.some((t) => t.projectId === project.id && t.status === "ASSIGNED")
  const history = events.filter((e) => e.projectId === project.id)
  return (
    <section aria-label="Trách nhiệm và triển khai" className="mb-3 border-b border-border pb-3">
      <div className="flex flex-wrap items-center justify-between gap-x-5 gap-y-2">
        <div>
          <h2 className="sr-only">Trách nhiệm dự án</h2>
          <p className="flex flex-wrap gap-x-5 gap-y-1 text-xs leading-6">
            <span>
              <span className="mr-1.5 text-muted-foreground">PM</span>
              <strong className="font-medium">{getPerson(project.pmId)?.name}</strong>
            </span>
            <span>
              <span className="mr-1.5 text-muted-foreground">PTCM</span>
              <strong className="font-medium">
                {getPerson(project.ptcmId)?.name ?? "Chưa chỉ định"}
              </strong>
            </span>
          </p>
        </div>
        {((pm && editable) || history.length > 0) && (
          <Button
            size="sm"
            variant="ghost"
            aria-expanded={managing}
            aria-controls="project-management"
            onClick={() => setManaging(!managing)}
          >
            <GearIcon />
            {pm && editable ? "Quản lý dự án" : "Lịch sử thay đổi"}
            <ChevronDownIcon className={managing ? "rotate-180" : ""} />
          </Button>
        )}
      </div>
      {!project.ptcmId && !managing && (
        <p className="mt-1 text-xs text-muted-foreground">
          Chưa có PTCM để thiết lập và phân công công việc.
        </p>
      )}
      {managing && (
        <div id="project-management" className="mt-3 rounded-lg border border-border bg-white p-4">
          <div className="flex flex-wrap gap-2">
            {pm && editable && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  setCandidate(project.ptcmId ?? "")
                  setError("")
                  setEditing(true)
                }}
              >
                {project.ptcmId ? "Thay PTCM" : "Chỉ định PTCM"}
              </Button>
            )}
            {pm && project.status === "PLANNING" && (
              <Button
                size="sm"
                disabled={!hasAssigned}
                title={!hasAssigned ? "Cần ít nhất một công việc đã phân công" : undefined}
                onClick={() => {
                  setError("")
                  setStarting(true)
                }}
              >
                Bắt đầu dự án
              </Button>
            )}
          </div>
          <p className="mt-3 text-xs leading-5 text-muted-foreground">
            {!project.ptcmId
              ? "Chưa có PTCM nên chưa thể tạo nhóm, công việc hoặc phân công. PM cần chỉ định người phụ trách chuyên môn."
              : "PTCM tổ chức, phân công và nghiệm thu; không nhận công việc trong chính dự án này."}
            {pm &&
              project.status === "PLANNING" &&
              !hasAssigned &&
              " Cần ít nhất một công việc đã phân công để bắt đầu triển khai."}
          </p>
          {history.length > 0 && (
            <details className="mt-3 border-t border-border pt-3 text-xs">
              <summary className="field-label cursor-pointer">
                Thay đổi trong phiên mô phỏng ({history.length})
              </summary>
              <ul className="mt-2 space-y-2">
                {history.map((e, i) => (
                  <li key={i}>
                    {e.action} · {getPerson(e.actorId)?.name} ·{" "}
                    {new Date(e.at).toLocaleTimeString("vi-VN")}
                    {e.action === "Chỉ định PTCM" &&
                      ` · ${getPerson(e.before?.ptcmId)?.name ?? "Chưa chỉ định"} → ${getPerson(e.after.ptcmId)?.name}`}
                  </li>
                ))}
              </ul>
            </details>
          )}
        </div>
      )}
      {editing && (
        <Dialog
          open
          onOpenChange={(open) => !open && setEditing(false)}
          title="Chỉ định phụ trách chuyên môn"
          description="PM có thể kiêm nhiệm. Không tự thu hồi công việc hoặc thay đổi đóng góp đã có."
        >
          <form
            className="space-y-4"
            onSubmit={(e) => {
              e.preventDefault()
              try {
                onChange(
                  appointPtcm(project, actorId, candidate, tasks, conflicts),
                  "Chỉ định PTCM",
                )
                setEditing(false)
              } catch (err) {
                setError((err as Error).message)
              }
            }}
          >
            <label htmlFor="ptcm-candidate" className="field-label block">
              Nhân viên phụ trách
            </label>
            <select
              id="ptcm-candidate"
              className="field w-full"
              value={candidate}
              aria-describedby="ptcm-hint"
              onChange={(e) => {
                setCandidate(e.target.value)
                setError("")
              }}
            >
              <option value="">Chọn nhân viên</option>
              {people
                .filter((p) => p.active)
                .map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.name}
                    {p.id === project.pmId ? " · PM kiêm nhiệm" : ""}
                  </option>
                ))}
            </select>
            <p
              id="ptcm-hint"
              className={`text-sm leading-6 ${block ? "text-destructive" : "text-muted-foreground"}`}
            >
              {block ??
                "Người được chọn không còn công việc mở hoặc đầu ra của mình cần nghiệm thu trong dự án."}
            </p>
            {error && (
              <p role="alert" className="text-sm text-destructive">
                {error}
              </p>
            )}
            <div className="flex justify-end gap-2 pt-3">
              <Button type="button" variant="outline" onClick={() => setEditing(false)}>
                Hủy
              </Button>
              <Button
                type="submit"
                disabled={!candidate || !!block || candidate === project.ptcmId}
              >
                Lưu PTCM
              </Button>
            </div>
          </form>
        </Dialog>
      )}
      {starting && (
        <Dialog
          open
          onOpenChange={setStarting}
          title="Bắt đầu triển khai dự án?"
          description="Dự án chuyển từ Lập kế hoạch sang Đang thực hiện. Các phân công hiện có được giữ nguyên."
        >
          {error && (
            <p role="alert" className="mb-4 text-sm text-destructive">
              {error}
            </p>
          )}
          <div className="flex justify-end gap-2">
            <Button variant="outline" onClick={() => setStarting(false)}>
              Chưa bắt đầu
            </Button>
            <Button
              onClick={() => {
                try {
                  onChange(startProject(project, actorId, tasks), "Bắt đầu dự án")
                  setStarting(false)
                } catch (err) {
                  setError((err as Error).message)
                }
              }}
            >
              Bắt đầu triển khai
            </Button>
          </div>
        </Dialog>
      )}
    </section>
  )
}
