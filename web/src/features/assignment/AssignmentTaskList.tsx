import { CheckIcon, Cross2Icon } from "@radix-ui/react-icons"
import type { Draft, Task } from "@/models/workspace"
import { cn } from "@/lib/utils"

export function AssignmentTaskList({
  selectedTasks,
  activeId,
  draft,
  onSelectTask,
  onRemoveTask,
}: {
  selectedTasks: Task[]
  activeId: string
  draft: Draft
  onSelectTask: (id: string) => void
  onRemoveTask: (id: string) => void
}) {
  return (
    <aside
      aria-label="Công việc đã chọn"
      className="assignment-tasks min-w-0 rounded-lg border border-border bg-white"
    >
      <h3 className="section-title border-b border-border px-4 py-3">Công việc đã chọn</h3>
      <div className="assignment-task-list p-2">
        {selectedTasks.map((t) => (
          <div
            key={t.id}
            className={cn(
              "assignment-task flex rounded-md",
              t.id === activeId ? "bg-accent" : "hover:bg-background",
            )}
          >
            <button
              type="button"
              aria-pressed={t.id === activeId}
              className="min-w-0 flex-1 p-3 text-left"
              onClick={() => onSelectTask(t.id)}
            >
              <p className="text-xs text-muted-foreground">{t.code}</p>
              <p className="mt-1 text-sm font-medium leading-5 line-clamp-2">{t.name}</p>
              <div className="mt-2 flex items-center justify-between gap-2 text-xs text-muted-foreground">
                <span>{t.estimate} giờ</span>
                {draft[t.id] ? (
                  <span className="inline-flex items-center gap-1 text-success">
                    <CheckIcon />
                    Đã chọn
                  </span>
                ) : (
                  <span>Chưa chọn</span>
                )}
              </div>
            </button>
            <button
              type="button"
              aria-label={`Bỏ ${t.name} khỏi lần phân công`}
              className="mr-1 mt-2 self-start rounded p-1.5 text-muted-foreground hover:bg-white"
              onClick={() => onRemoveTask(t.id)}
            >
              <Cross2Icon />
            </button>
          </div>
        ))}
      </div>
    </aside>
  )
}
