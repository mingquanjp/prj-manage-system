import { LockClosedIcon, Pencil1Icon } from "@radix-ui/react-icons"
import { Button } from "@/components/ui/button"
import { TaskDetailHeader, TaskSpecification } from "@/components/task-detail"
import { getPerson } from "@/demo/identity"
import { isTaskFrozen } from "@/demo/task-structure"
import type { Project, Task, OutputConflict } from "@/models/workspace"
import type { ReviewResult, WorkState } from "@/demo/effort"
import { cn } from "@/lib/utils"
import { TaskReviewPanel } from "./TaskReviewPanel"
import { ProjectLifecycleActions } from "./ProjectLifecycleActions"
import type { LifecycleCommand } from "@/demo/project-lifecycle"
import type { ReassignmentCommand } from "@/demo/reassignment"
import { TaskReassignment } from "@/features/assignment/TaskReassignment"

export function ProjectTaskSidebar({
  currentTask,
  project,
  actorId,
  allowed,
  canReadResults,
  detailTab,
  work,
  conflicts,
  percent,
  related,
  onCloseDetail,
  onDetailTabChange,
  onEditTask,
  onReview,
  tasks,
  onLifecycle,
  onReassign,
}: {
  currentTask: Task | undefined
  project: Project
  actorId: string
  allowed: boolean
  canReadResults: boolean
  detailTab: "details" | "results"
  work: WorkState
  conflicts: OutputConflict[]
  percent: number | null
  related: Task[]
  onCloseDetail: () => void
  onDetailTabChange: (tab: "details" | "results") => void
  onEditTask: (task: Task) => void
  onReview: (submissionId: string, result: ReviewResult, comment: string) => void
  tasks: Task[]
  onLifecycle: (command: LifecycleCommand) => void
  onReassign: (command: ReassignmentCommand) => void
}) {
  return (
    <aside
      className={cn(
        "min-w-0 rounded-lg border border-border bg-white p-5 xl:sticky xl:top-6",
        currentTask && "xl:max-h-[calc(100dvh-8rem)] xl:overflow-y-auto",
      )}
      aria-label={currentTask ? "Chi tiết công việc" : "Thông tin dự án"}
      onKeyDown={(e) => {
        if (
          e.key === "Escape" &&
          currentTask &&
          !(e.target as HTMLElement).closest('[role="dialog"]')
        )
          onCloseDetail()
      }}
    >
      {currentTask ? (
        <>
          <TaskDetailHeader task={currentTask} onClose={onCloseDetail} />
          <div className="mt-3">
            <ProjectLifecycleActions
              project={project}
              tasks={tasks}
              work={work}
              actorId={actorId}
              task={currentTask}
              onTransition={onLifecycle}
            />
          </div>
          {currentTask.status === "CANCELLED" && (
            <p className="mt-3 whitespace-pre-wrap break-words rounded-md bg-accent p-3 text-xs leading-5">
              Công việc đã hủy. Nhật ký và các lần nộp được giữ nguyên, không kết luận nghiệm thu
              thay.
              {currentTask.executionHistory?.at(-1)?.reason &&
                ` Lý do: ${currentTask.executionHistory.at(-1)?.reason}`}
            </p>
          )}
          {canReadResults && (
            <div className="mt-4 flex gap-1" role="group" aria-label="Nội dung công việc dự án">
              <Button
                size="sm"
                variant={detailTab === "details" ? "default" : "ghost"}
                aria-pressed={detailTab === "details"}
                className="workspace-tab"
                onClick={() => onDetailTabChange("details")}
              >
                Chi tiết
              </Button>
              <Button
                size="sm"
                variant={detailTab === "results" ? "default" : "ghost"}
                aria-pressed={detailTab === "results"}
                className="workspace-tab"
                onClick={() => onDetailTabChange("results")}
              >
                Kết quả
              </Button>
            </div>
          )}
          {detailTab === "results" && canReadResults ? (
            <>
              <details className="mt-4 rounded-md border border-border p-3 text-sm" open>
                <summary className="section-title cursor-pointer">Tiêu chí nghiệm thu</summary>
                <p className="mt-2 whitespace-pre-wrap break-words leading-6 text-muted-foreground">
                  {currentTask.criteria}
                </p>
              </details>
              <TaskReviewPanel
                key={currentTask.id}
                project={project}
                task={currentTask}
                work={work}
                actorId={actorId}
                conflicts={conflicts}
                onReview={onReview}
              />
            </>
          ) : (
            <>
              {allowed &&
                (isTaskFrozen(currentTask) ? (
                  <p className="mt-4 flex items-start gap-2 rounded-md bg-accent p-3 text-xs leading-5 text-muted-foreground">
                    <LockClosedIcon className="mt-0.5" />
                    Đặc tả đã khóa sau lần phân công đầu. Biểu mẫu sửa làm rõ/gia hạn chưa dựng
                    trong đợt này.
                  </p>
                ) : (
                  <Button
                    variant="outline"
                    size="sm"
                    className="mt-4"
                    onClick={() => onEditTask(currentTask)}
                  >
                    <Pencil1Icon />
                    Chỉnh sửa công việc
                  </Button>
                ))}
              <dl className="mt-4 space-y-2 border-t border-border pt-4 text-sm">
                <div className="flex justify-between gap-2">
                  <dt className="field-label">Người thực hiện</dt>
                  <dd className="text-right">
                    {getPerson(currentTask.assigneeId)?.name ?? "Chưa phân công"}
                  </dd>
                </div>
                <div className="flex justify-between gap-2">
                  <dt className="field-label">Nhóm</dt>
                  <dd className="text-right">{currentTask.group}</dd>
                </div>
                <div className="flex justify-between">
                  <dt className="field-label">Ưu tiên</dt>
                  <dd>{currentTask.priority}</dd>
                </div>
                <div className="flex justify-between gap-2">
                  <dt className="field-label">Kế hoạch</dt>
                  <dd className="text-right">
                    {currentTask.start && currentTask.end
                      ? `${currentTask.start.split("-").reverse().join("/")} – ${currentTask.end.split("-").reverse().join("/")}`
                      : "Chưa đặt ngày"}
                  </dd>
                </div>
              </dl>
              {allowed && (
                <TaskReassignment
                  key={`reassignment-${currentTask.id}`}
                  project={project}
                  task={currentTask}
                  tasks={tasks}
                  work={work}
                  actorId={actorId}
                  onReassign={onReassign}
                />
              )}
              <div className="mt-5 border-t border-border pt-4 text-sm">
                Ước tính <span className="float-right font-medium">{currentTask.estimate} giờ</span>
              </div>
              <TaskSpecification key={currentTask.id} task={currentTask} showWeights />
            </>
          )}
        </>
      ) : (
        <>
          <h2 className="section-title">Về dự án</h2>
          <p className="mt-3 text-sm leading-6 text-muted-foreground">{project.description}</p>
          <dl className="mt-5 space-y-4 text-sm">
            <div className="flex justify-between gap-3">
              <dt className="field-label">Ưu tiên</dt>
              <dd>{project.priority}</dd>
            </div>
            <div className="flex justify-between gap-3">
              <dt className="field-label">Ngày bắt đầu</dt>
              <dd>{new Date(`${project.start}T00:00:00`).toLocaleDateString("vi-VN")}</dd>
            </div>
            <div className="flex justify-between gap-3">
              <dt className="field-label">Hạn dự kiến</dt>
              <dd>{new Date(`${project.end}T00:00:00`).toLocaleDateString("vi-VN")}</dd>
            </div>
          </dl>
          <div className="mt-6 border-t border-border pt-5">
            <div className="flex justify-between text-sm">
              <span>Đã nghiệm thu</span>
              <strong>{percent === null ? "—" : `${percent}%`}</strong>
            </div>
            <div className="mt-3 h-1.5 overflow-hidden rounded-full bg-accent">
              <div
                className="h-full rounded-full bg-success"
                style={{ width: `${percent ?? 0}%` }}
              />
            </div>
            <p className="mt-3 text-xs text-muted-foreground">
              {related.filter((t) => ["DONE", "CLOSED"].includes(t.status)).length} /{" "}
              {related.filter((t) => t.status !== "CANCELLED").length} công việc hiệu lực
            </p>
          </div>
          <div className="mt-6 border-t border-border pt-5 text-xs leading-5 text-muted-foreground">
            Không cần lập danh sách thành viên trước. PTCM lựa chọn người phù hợp trong toàn công
            ty.
          </div>
        </>
      )}
    </aside>
  )
}
