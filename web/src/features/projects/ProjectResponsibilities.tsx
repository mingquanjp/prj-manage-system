import { usePeople } from "@/components/people-directory"
import { useState } from "react"
import { ChevronDownIcon, GearIcon } from "@radix-ui/react-icons"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { appointPtcm, ptcmBlockReason, startProject } from "@/demo/projects"
import type { OutputConflict, Project, ProjectEvent, Task } from "@/models/workspace"
import type { WorkState } from "@/demo/effort"
import type { LifecycleCommand } from "@/demo/project-lifecycle"
import { ProjectLifecycleActions } from "./ProjectLifecycleActions"
import { ProjectManagementRow } from "./ProjectManagementRow"

export function ProjectResponsibilities({
  project,
  actorId,
  tasks,
  conflicts,
  events,
  onChange,
  work,
  onLifecycle,
  hasDraft,
}: {
  project: Project
  actorId: string
  tasks: Task[]
  conflicts: OutputConflict[]
  events: ProjectEvent[]
  onChange: (project: Project, action: string) => void
  work: WorkState
  onLifecycle: (command: LifecycleCommand) => void
  hasDraft: boolean
}) {
  const { people, getPerson } = usePeople()
  const [editing, setEditing] = useState(false)
  const [candidate, setCandidate] = useState(project.ptcmId ?? "")
  const [error, setError] = useState("")
  const [starting, setStarting] = useState(false)
  const [managing, setManaging] = useState(false)
  const pm = project.pmId === actorId
  const editable = ["PLANNING", "IN_PROGRESS"].includes(project.status)
  const manageable = editable || project.status === "COMPLETED"
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
        {((pm && manageable) || history.length > 0) && (
          <Button
            size="sm"
            variant="ghost"
            aria-expanded={managing}
            aria-controls="project-management"
            onClick={() => setManaging(!managing)}
          >
            <GearIcon />
            {pm && manageable ? "Quản lý dự án" : "Lịch sử thay đổi"}
            <ChevronDownIcon className={managing ? "rotate-180" : ""} />
          </Button>
        )}
      </div>
      {["COMPLETED", "CLOSED", "CANCELLED"].includes(project.status) && (
        <p className="mt-3 rounded-md bg-accent p-3 text-xs leading-5">
          {project.status === "COMPLETED"
            ? "Dự án đã hoàn thành: không thực hiện thêm công việc. PTCM vẫn xử lý hiệu chỉnh nhật ký; PM đóng khi hết yêu cầu chờ."
            : "Dự án đã đóng/hủy: dữ liệu nghiệp vụ chỉ đọc, lịch sử và minh chứng được giữ nguyên. Không mở lại dự án."}
          {project.status === "CANCELLED" &&
            work.requests.some(
              (request) =>
                request.decisionReason?.startsWith("PROJECT_CANCELLED") &&
                work.contributions.some(
                  (entry) =>
                    entry.projectId === project.id && entry.assignment.id === request.assignmentId,
                ),
            ) &&
            " Có yêu cầu hiệu chỉnh chưa được giải quyết trước khi hủy; nhật ký có thể chưa đầy đủ. Xem lịch sử hiệu chỉnh để đối chiếu."}
        </p>
      )}
      {!project.ptcmId && !managing && (
        <p className="mt-1 text-xs text-muted-foreground">
          Chưa có PTCM để thiết lập và phân công công việc.
        </p>
      )}
      {managing && (
        <div id="project-management" className="mt-3 rounded-lg border border-border bg-white px-4">
          {pm && editable && (
            <ProjectManagementRow
              title="Phụ trách chuyên môn"
              description={
                project.ptcmId
                  ? (getPerson(project.ptcmId)?.name ?? "Chưa có thông tin nhân viên")
                  : "Chưa chỉ định người thiết lập và phân công công việc."
              }
            >
              <Button
                variant="outline"
                size="sm"
                aria-label={
                  project.ptcmId ? "Thay đổi phụ trách chuyên môn" : "Chỉ định phụ trách chuyên môn"
                }
                onClick={() => {
                  setCandidate(project.ptcmId ?? "")
                  setError("")
                  setEditing(true)
                }}
              >
                {project.ptcmId ? "Thay đổi" : "Chỉ định PTCM"}
              </Button>
            </ProjectManagementRow>
          )}
          {pm && project.status === "PLANNING" && (
            <ProjectManagementRow
              title="Bắt đầu dự án"
              description={
                hasAssigned
                  ? "Cho phép người được giao bắt đầu thực hiện công việc."
                  : "Cần ít nhất một công việc đã phân công để bắt đầu triển khai."
              }
            >
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
            </ProjectManagementRow>
          )}
          <ProjectLifecycleActions
            project={project}
            tasks={tasks}
            work={work}
            actorId={actorId}
            onTransition={onLifecycle}
            hasDraft={hasDraft}
          />
          {history.length > 0 && (
            <details className="border-t border-border py-3 text-xs">
              <summary className="field-label cursor-pointer">
                Thay đổi trong phiên mô phỏng ({history.length})
              </summary>
              <ul className="mt-2 space-y-2">
                {history.map((e, i) => (
                  <li key={i}>
                    {e.action} · {getPerson(e.actorId)?.name} ·{" "}
                    {new Date(e.at).toLocaleTimeString("vi-VN")}
                    {e.reason && (
                      <p className="mt-1 whitespace-pre-wrap break-words text-muted-foreground">
                        Lý do: {e.reason}
                      </p>
                    )}
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
