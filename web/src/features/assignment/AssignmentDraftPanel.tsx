import { useIdentity } from "@/components/identity"
import { usePeople } from "@/components/people-directory"
import { useState } from "react"
import { ChevronDownIcon, PersonIcon } from "@radix-ui/react-icons"
import { Avatar } from "@/components/shared"
import { addedLoad, loadFor } from "@/demo/assignment"
import type { Draft, Task } from "@/models/workspace"
import { cn } from "@/lib/utils"

export function AssignmentDraftPanel({
  selectedTasks,
  taskCount,
  tasks,
  draft,
  errors,
  complete,
  error,
  onSelectTask,
}: {
  selectedTasks: Task[]
  taskCount: number
  tasks: Task[]
  draft: Draft
  errors: string[]
  complete: boolean
  error: string
  onSelectTask: (id: string) => void
}) {
  const identity = useIdentity()

  const { getPerson } = usePeople()
  const [draftExpanded, setDraftExpanded] = useState(true)
  return (
    <aside
      aria-label="Bản nháp phân công"
      className="min-w-0 rounded-lg border border-border bg-white p-4 assignment-draft"
    >
      <div className="flex items-center justify-between">
        <h3>
          <button
            type="button"
            className="section-title inline-flex items-center gap-2 text-left"
            aria-expanded={draftExpanded}
            aria-controls="assignment-draft-items"
            onClick={() => setDraftExpanded((prev) => !prev)}
          >
            Bản nháp phân công{" "}
            <ChevronDownIcon
              className={cn("transition-transform", draftExpanded && "rotate-180")}
            />
          </button>
        </h3>
        <span className="count">
          {Object.keys(draft).length}/{taskCount}
        </span>
      </div>
      <p className="mt-2 text-xs leading-5 text-muted-foreground">
        Lựa chọn chưa làm thay đổi người phụ trách hoặc tải công việc.
      </p>
      <div
        id="assignment-draft-items"
        hidden={!draftExpanded}
        className="my-5 max-h-[40dvh] overflow-y-auto space-y-4"
      >
        {selectedTasks.map((t) => (
          <div key={t.id} className="border-b border-border pb-4 last:border-0 last:pb-0">
            <button
              type="button"
              className="text-left text-sm leading-5 hover:underline"
              onClick={() => onSelectTask(t.id)}
            >
              <span className="text-sm font-medium">{t.name}</span>
            </button>
            <div className="mt-2 flex items-center gap-2 text-xs text-muted-foreground">
              {draft[t.id] ? (
                <>
                  <Avatar initials={getPerson(draft[t.id].personId)!.initials} small />
                  <span>
                    {getPerson(draft[t.id].personId)?.name}
                    <span className="mt-0.5 block">
                      {draft[t.id].mode === "manual"
                        ? "Thủ công"
                        : `Gợi ý hạng ${draft[t.id].rank}`}
                    </span>
                  </span>
                </>
              ) : (
                <>
                  <PersonIcon />
                  Chưa chọn người
                </>
              )}
            </div>
          </div>
        ))}
      </div>
      <div className="border-t border-border pt-4">
        <h4 className="field-label text-balance">Tải sau bản nháp</h4>
        <div className="mt-3 space-y-3">
          {[...new Set(Object.values(draft).map((c) => c.personId))].map((id) => (
            <div key={id} className="text-xs">
              <p className="text-sm font-medium">{getPerson(id)?.name}</p>
              <p className="mt-1 text-muted-foreground tabular-nums">
                Hiện tại {loadFor(id, tasks, identity)} giờ · Nháp +{addedLoad(id, draft, tasks)}{" "}
                giờ
              </p>
              <p className="mt-1 text-muted-foreground tabular-nums">
                Dự kiến{" "}
                <strong className="text-foreground">
                  {loadFor(id, tasks, identity) + addedLoad(id, draft, tasks)} giờ
                </strong>
              </p>
            </div>
          ))}
          {!Object.keys(draft).length && (
            <p className="text-xs text-muted-foreground">Hiển thị sau khi chọn ứng viên.</p>
          )}
        </div>
      </div>
      {Object.keys(draft).length > 0 && errors.length > 0 && (
        <div className="mt-4 rounded-md bg-warning-soft p-3 text-xs leading-5 text-warning">
          {errors.join(" ")}
        </div>
      )}
      {!complete && (
        <p className="mt-2 text-xs leading-5 text-muted-foreground">
          Chọn người cho từng việc hoặc bỏ việc chưa cần giao.
        </p>
      )}
      {error && (
        <p role="alert" className="mt-3 text-sm text-destructive">
          {error}
        </p>
      )}
    </aside>
  )
}
