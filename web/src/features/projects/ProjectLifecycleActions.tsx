import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import type { Project, Task } from "@/models/workspace"
import type { WorkState } from "@/demo/effort"
import { getPerson } from "@/demo/identity"
import { cn } from "@/lib/utils"
import { ProjectManagementRow } from "./ProjectManagementRow"
import {
  lifecycleBlockReason,
  lifecycleLabels,
  isOpenTask,
  type LifecycleAction,
  type LifecycleCommand,
} from "@/demo/project-lifecycle"

const descriptions: Record<LifecycleAction, string> = {
  complete:
    "Kết thúc giai đoạn thực hiện. Nhật ký cũ vẫn được hiệu chỉnh qua yêu cầu; không tạo hoặc thực hiện thêm công việc.",
  close:
    "Khóa dữ liệu nghiệp vụ dự án. Các công việc đã nghiệm thu chuyển sang Đã đóng; lịch sử và minh chứng được giữ nguyên. Không mở lại dự án.",
  cancelProject:
    "Hủy các công việc còn mở, giữ công việc đã nghiệm thu và toàn bộ nhật ký/minh chứng. Yêu cầu hiệu chỉnh đang chờ sẽ kết thúc do hủy dự án, không áp dụng đề nghị. Lần nộp chưa nghiệm thu không bị coi là không đạt. Không mở lại dự án.",
  cancelTask:
    "Kết thúc lượt giao và dừng thực hiện công việc. Nhật ký, lịch sử và các lần nộp được giữ nguyên; lần nộp chưa nghiệm thu không bị coi là không đạt. Không mở lại công việc.",
}

const rowDescriptions: Record<LifecycleAction, string> = {
  complete: "Kết thúc thực hiện khi các công việc đã nghiệm thu hoặc đã hủy.",
  close: "Khóa dữ liệu nghiệp vụ sau khi xử lý hết yêu cầu hiệu chỉnh nhật ký.",
  cancelProject: "Dừng dự án; giữ lịch sử, nhật ký và minh chứng đã có.",
  cancelTask: "Dừng công việc; giữ lịch sử và minh chứng đã có.",
}

export function ProjectLifecycleActions({
  project,
  tasks,
  work,
  actorId,
  task,
  onTransition,
  hasDraft = false,
}: {
  project: Project
  tasks: Task[]
  work: WorkState
  actorId: string
  task?: Task
  hasDraft?: boolean
  onTransition: (command: LifecycleCommand) => void
}) {
  const [command, setCommand] = useState<LifecycleCommand | null>(null)
  const [reason, setReason] = useState("")
  const [error, setError] = useState("")
  if (project.pmId !== actorId || !getPerson(actorId)?.active) return null
  const active = ["PLANNING", "IN_PROGRESS"].includes(project.status)
  const actions: LifecycleAction[] = task
    ? active && isOpenTask(task)
      ? ["cancelTask"]
      : []
    : project.status === "COMPLETED"
      ? ["close"]
      : project.status === "IN_PROGRESS"
        ? ["complete", "cancelProject"]
        : project.status === "PLANNING"
          ? ["cancelProject"]
          : []
  return (
    <div className={cn(task && "flex flex-wrap gap-2")}>
      {actions.map((action) => {
        const block = lifecycleBlockReason(project, tasks, work, actorId, action, task)
        const button = (
          <Button
            size="sm"
            variant="outline"
            disabled={!!block}
            title={block}
            className={cn(action.startsWith("cancel") && "text-destructive")}
            onClick={() => {
              setCommand({
                projectId: project.id,
                expectedStatus: project.status,
                action,
                ...(task ? { taskId: task.id, expectedTaskStatus: task.status } : {}),
              })
              setReason("")
              setError("")
            }}
          >
            {lifecycleLabels[action]}
          </Button>
        )
        return task ? (
          <div key={action} className="max-w-sm">
            {button}
            {block && <p className="mt-1 text-xs leading-5 text-muted-foreground">{block}</p>}
          </div>
        ) : (
          <ProjectManagementRow
            key={action}
            title={lifecycleLabels[action]}
            description={block ?? rowDescriptions[action]}
          >
            {button}
          </ProjectManagementRow>
        )
      })}
      {command && (
        <Dialog
          open
          onOpenChange={(open) => !open && setCommand(null)}
          title={`${lifecycleLabels[command.action]}?`}
          description={descriptions[command.action]}
        >
          <form
            onSubmit={(event) => {
              event.preventDefault()
              try {
                onTransition({ ...command, reason })
                setCommand(null)
              } catch (err) {
                setError((err as Error).message)
              }
            }}
          >
            <p className="mb-4 text-sm font-medium">
              {task ? `${task.code} · ${task.name}` : `${project.code} · ${project.name}`}
            </p>
            {hasDraft && (
              <p className="mb-4 rounded-md bg-accent p-3 text-sm">
                Bản nháp phân công chưa xác nhận sẽ bị bỏ khi tiếp tục.
              </p>
            )}
            {command.action.startsWith("cancel") && (
              <label className="block">
                <span className="field-label">Lý do hủy *</span>
                <textarea
                  autoFocus
                  className="field mt-2 min-h-24 w-full"
                  value={reason}
                  onChange={(event) => setReason(event.target.value)}
                  aria-invalid={!!error}
                  aria-describedby={error ? "lifecycle-error" : undefined}
                />
              </label>
            )}
            {error && (
              <p id="lifecycle-error" role="alert" className="mt-3 text-sm text-destructive">
                {error}
              </p>
            )}
            <div className="mt-5 flex justify-end gap-2">
              <Button type="button" variant="outline" onClick={() => setCommand(null)}>
                Quay lại
              </Button>
              <Button type="submit">
                Xác nhận {lifecycleLabels[command.action].toLocaleLowerCase("vi-VN")}
              </Button>
            </div>
          </form>
        </Dialog>
      )}
    </div>
  )
}
