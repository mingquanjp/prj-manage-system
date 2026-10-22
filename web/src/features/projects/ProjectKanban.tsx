import { FileTextIcon } from "@radix-ui/react-icons"
import { Avatar, StatusBadge } from "@/components/shared"
import { getPerson, type Task } from "@/demo/model"
import { projectBoard } from "@/demo/kanban"
import { cn } from "@/lib/utils"

export function ProjectKanban({
  tasks,
  selectedId,
  selectedAssignments,
  canSelect,
  onToggle,
  onSelect,
}: {
  tasks: Task[]
  selectedId: string | null
  selectedAssignments: string[]
  canSelect: boolean
  onToggle: (id: string) => void
  onSelect: (task: Task, trigger: HTMLButtonElement) => void
}) {
  return (
    <div
      role="region"
      aria-label="Bảng Kanban dự án"
      tabIndex={0}
      className="overflow-x-auto rounded-lg border border-border bg-sidebar p-3"
    >
      <div className="flex min-h-80 items-stretch gap-3">
        {projectBoard(tasks).map((column) => (
          <section
            key={column.id}
            aria-label={column.label}
            className="w-64 shrink-0 rounded-lg bg-accent/50 p-2"
          >
            <div className="mb-3 flex items-center gap-2 px-1 py-2">
              <h2 className="text-sm font-medium">{column.label}</h2>
              <span
                className="count ml-auto"
                aria-label={`${column.tasks.length} công việc`}
              >
                {column.tasks.length}
              </span>
            </div>
            <div className="space-y-3">
              {column.tasks.map((task) => {
                const person = getPerson(task.assigneeId)
                return (
                  <article
                    key={task.id}
                    className={cn(
                      "rounded-lg border border-border bg-white p-3.5",
                      (selectedId === task.id ||
                        selectedAssignments.includes(task.id)) &&
                        "border-primary ring-1 ring-primary",
                    )}
                  >
                    <div className="flex items-center justify-between gap-2 text-xs text-muted-foreground">
                      <span className="flex items-center gap-2">
                        <FileTextIcon />
                        {task.code}
                      </span>
                      {canSelect && task.status === "PENDING_ASSIGNMENT" && (
                        <input
                          type="checkbox"
                          aria-label={`Chọn ${task.name}`}
                          checked={selectedAssignments.includes(task.id)}
                          onChange={() => onToggle(task.id)}
                        />
                      )}
                    </div>
                    <button
                      type="button"
                      aria-label={`Xem ${task.name}`}
                      aria-pressed={selectedId === task.id}
                      className="mt-2 block w-full break-words text-left text-sm font-medium leading-6 hover:underline"
                      onClick={(e) => onSelect(task, e.currentTarget)}
                    >
                      {task.name}
                    </button>
                    {task.status === "PAUSED" && (
                      <div className="mt-2">
                        <StatusBadge status="PAUSED" />
                        {task.executionHistory?.at(-1)?.reason && (
                          <p className="mt-2 whitespace-pre-wrap break-words text-xs leading-5 text-muted-foreground">
                            Lý do: {task.executionHistory.at(-1)?.reason}
                          </p>
                        )}
                      </div>
                    )}
                    <div className="mt-3 flex items-center gap-2 text-xs text-muted-foreground">
                      {person && <Avatar initials={person.initials} small />}
                      <span>{person?.name ?? "Chưa phân công"}</span>
                    </div>
                    <div className="mt-3 flex items-center justify-between gap-2 border-t border-border pt-3 text-xs text-muted-foreground">
                      <span className="min-w-0 truncate">{task.group}</span>
                      <span className="shrink-0">{task.estimate} giờ</span>
                    </div>
                  </article>
                )
              })}
              {!column.tasks.length && (
                <p className="rounded-md border border-dashed border-border px-3 py-6 text-center text-xs text-muted-foreground">
                  Chưa có công việc
                </p>
              )}
            </div>
          </section>
        ))}
      </div>
    </div>
  )
}
