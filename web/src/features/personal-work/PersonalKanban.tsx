import { FileTextIcon } from "@radix-ui/react-icons"
import { statusLabels, type Project, type Task, type TaskStatus } from "@/demo/model"
import { cn } from "@/lib/utils"
import { StatusBadge } from "@/components/shared"

// Cột biểu thị giai đoạn xử lý, không thay đổi trạng thái nghiệp vụ của task.
const columns: { id: TaskStatus; label: string; statuses: TaskStatus[] }[] = [
  { id: "ASSIGNED", label: statusLabels.ASSIGNED, statuses: ["ASSIGNED"] },
  {
    id: "IN_PROGRESS",
    label: "Đang xử lý",
    statuses: ["IN_PROGRESS", "PAUSED"],
  },
  {
    id: "PENDING_REVIEW",
    label: statusLabels.PENDING_REVIEW,
    statuses: ["PENDING_REVIEW"],
  },
  { id: "DONE", label: statusLabels.DONE, statuses: ["DONE"] },
]

export function PersonalKanban({
  tasks,
  projects,
  selectedId,
  onSelect,
}: {
  tasks: Task[]
  projects: Project[]
  selectedId: string | null
  onSelect: (task: Task, trigger: HTMLButtonElement) => void
}) {
  return (
    <div
      role="region"
      aria-label="Bảng Kanban công việc cá nhân"
      tabIndex={0}
      className="overflow-x-auto rounded-lg border border-border bg-sidebar p-3"
    >
      <div className="flex min-h-80 items-stretch gap-3">
        {columns.map((column) => {
          const rows = tasks.filter((t) => column.statuses.includes(t.status))
          return (
            <section
              key={column.id}
              aria-label={column.label}
              className="w-64 shrink-0 rounded-lg bg-accent/50 p-2"
            >
              <div className="mb-3 flex items-center gap-2 px-1 py-2">
                <span
                  aria-hidden="true"
                  className={cn(
                    "size-2 rounded-full bg-muted-foreground",
                    column.id === "DONE" && "bg-success",
                    (column.id === "IN_PROGRESS" || column.id === "PENDING_REVIEW") && "bg-warning",
                  )}
                />
                <h2 className="field-label text-balance">{column.label}</h2>
                <span className="count ml-auto" aria-label={`${rows.length} công việc`}>
                  {rows.length}
                </span>
              </div>
              <div className="space-y-3">
                {rows.map((task) => (
                  <button
                    key={task.id}
                    type="button"
                    aria-label={`Xem ${task.name}`}
                    aria-pressed={selectedId === task.id}
                    onClick={(e) => onSelect(task, e.currentTarget)}
                    className={cn(
                      "block w-full rounded-lg border border-border bg-white p-3.5 text-left transition-colors hover:border-muted-foreground/50",
                      selectedId === task.id && "border-primary ring-1 ring-primary",
                    )}
                  >
                    <span className="flex items-center gap-2 text-xs text-muted-foreground">
                      <FileTextIcon />
                      {task.code}
                    </span>
                    <span className="mt-2 block break-words text-sm font-medium leading-6">
                      {task.name}
                    </span>
                    {task.status === "PAUSED" && (
                      <span className="mt-2 block">
                        <StatusBadge status="PAUSED" />
                        {task.executionHistory?.at(-1)?.reason && (
                          <span className="mt-2 block whitespace-pre-wrap break-words text-xs leading-5 text-muted-foreground">
                            Lý do: {task.executionHistory.at(-1)?.reason}
                          </span>
                        )}
                      </span>
                    )}
                    <span className="mt-3 block break-words text-xs leading-5 text-muted-foreground">
                      {projects.find((p) => p.id === task.projectId)?.name}
                    </span>
                    <span className="mt-3 flex items-center justify-between gap-2 border-t border-border pt-3 text-xs text-muted-foreground">
                      <span className="min-w-0 truncate">{task.group}</span>
                      <span className="shrink-0">{task.estimate} giờ</span>
                    </span>
                  </button>
                ))}
                {!rows.length && (
                  <p className="rounded-md border border-dashed border-border px-3 py-6 text-center text-xs text-muted-foreground">
                    Chưa có công việc
                  </p>
                )}
              </div>
            </section>
          )
        })}
      </div>
    </div>
  )
}
