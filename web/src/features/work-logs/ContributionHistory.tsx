import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { type Project, type Task } from "@/demo/model"
import { type WorkState } from "@/demo/effort"
import { WorkLogPanel } from "@/features/personal-work/WorkLogPanel"
import { historicalContributions } from "@/demo/corrections"

// Tách quyền xem đóng góp của chính chủ khỏi quyền thực hiện công việc hiện tại.
export function ContributionHistory({ projects, tasks, work, actorId, onChange }: {
  projects: Project[]
  tasks: Task[]
  work: WorkState
  actorId: string
  onChange: (work: WorkState) => void
}) {
  const [selectedId, setSelectedId] = useState<string>()
  const old = historicalContributions(projects, tasks, work, actorId)
  const selected = old.find((entry) => entry.assignment.id === selectedId)
  const project = projects.find((p) => p.id === selected?.projectId)
  const task = tasks.find((t) => t.id === selected?.taskId)
  if (!old.length) return null
  return (
    <section className="mt-8 rounded-lg border border-border bg-white p-5">
      <h2 className="section-title">Lịch sử đóng góp của tôi</h2>
      <p className="mt-2 text-sm text-muted-foreground text-pretty">
        Chỉ xem nhật ký của bạn; không cấp quyền vào công việc hiện tại của người khác.
      </p>
      <ul className="mt-3 space-y-2">
        {old.map((entry) => (
          <li key={entry.assignment.id}>
            <Button
              variant="ghost"
              className="h-auto max-w-full whitespace-normal text-left"
              onClick={() => setSelectedId(entry.assignment.id)}
            >
              {entry.taskCode} · {entry.taskName}
            </Button>
          </li>
        ))}
      </ul>
      <Dialog
        side
        open={Boolean(selected)}
        onOpenChange={(open) => { if (!open) setSelectedId(undefined) }}
        title="Nhật ký đóng góp cũ"
        description={selected ? `${selected.taskCode} · ${selected.taskName}` : "Chỉ xem đóng góp của bạn."}
      >
        {selected && project && task && (
          <WorkLogPanel
            key={selected.assignment.id}
            project={project}
            task={task}
            contribution={selected}
            projects={projects}
            tasks={tasks}
            work={work}
            actorId={actorId}
            onChange={onChange}
          />
        )}
      </Dialog>
    </section>
  )
}
