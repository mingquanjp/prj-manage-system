import { FileTextIcon } from "@radix-ui/react-icons"
import { Avatar, StatusBadge } from "@/components/shared"
import { getPerson } from "@/demo/identity"
import type { Task } from "@/models/workspace"
import { cn } from "@/lib/utils"

export function ProjectTaskTable({
  rows,
  canSelectAssignments,
  hasAll,
  hasEligible,
  selected,
  detail,
  onToggleAll,
  onToggle,
  onOpenTask,
}: {
  rows: Task[]
  canSelectAssignments: boolean
  hasAll: boolean
  hasEligible: boolean
  selected: string[]
  detail: string | null
  onToggleAll: () => void
  onToggle: (id: string) => void
  onOpenTask: (task: Task, trigger: HTMLButtonElement) => void
}) {
  return (
    <div className="overflow-x-auto">
      <table className="w-full min-w-[650px] text-left text-sm">
        <thead>
          <tr className="table-heading table-heading-strong">
            <th className="w-10 py-3 pl-4">
              {canSelectAssignments && (
                <input
                  type="checkbox"
                  aria-label="Chọn tất cả công việc đủ điều kiện theo bộ lọc"
                  checked={hasAll}
                  disabled={!hasEligible}
                  onChange={onToggleAll}
                />
              )}
            </th>
            <th className="min-w-[220px]">Tên công việc</th>
            <th>Trạng thái</th>
            <th>Người thực hiện</th>
            <th className="whitespace-nowrap pr-4 text-right">Ước tính</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((task) => {
            const person = getPerson(task.assigneeId)
            return (
              <tr
                key={task.id}
                className={cn(
                  "border-t border-border hover:bg-background",
                  selected.includes(task.id) && "bg-accent/60",
                  detail === task.id && "bg-accent",
                )}
              >
                <td className="pl-4">
                  {canSelectAssignments && (
                    <input
                      type="checkbox"
                      aria-label={`Chọn ${task.name}`}
                      disabled={task.status !== "PENDING_ASSIGNMENT"}
                      title={
                        task.status !== "PENDING_ASSIGNMENT"
                          ? "Chỉ chọn công việc đang chờ phân công"
                          : undefined
                      }
                      checked={selected.includes(task.id)}
                      onChange={() => onToggle(task.id)}
                    />
                  )}
                </td>
                <td className="py-4 pr-3">
                  <button
                    type="button"
                    className="flex items-start gap-2.5 text-left hover:underline"
                    aria-label={`Xem ${task.name}`}
                    onClick={(e) => onOpenTask(task, e.currentTarget)}
                  >
                    <FileTextIcon className="mt-0.5 shrink-0 text-muted-foreground" />
                    <span>
                      <span className="block font-medium leading-5">{task.name}</span>
                      <span className="mt-1 block text-xs text-muted-foreground">
                        {task.code} <span className="mx-1">/</span> {task.group}
                      </span>
                    </span>
                  </button>
                </td>
                <td className="pr-3">
                  <StatusBadge status={task.status} />
                </td>
                <td className="pr-3">
                  {person ? (
                    <span className="inline-flex items-center gap-2 whitespace-nowrap">
                      <Avatar initials={person.initials} small />
                      <span>{person.name}</span>
                    </span>
                  ) : (
                    <span className="text-muted-foreground">Chưa phân công</span>
                  )}
                </td>
                <td className="pr-4 text-right whitespace-nowrap tabular-nums">
                  {task.estimate} giờ
                </td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}
