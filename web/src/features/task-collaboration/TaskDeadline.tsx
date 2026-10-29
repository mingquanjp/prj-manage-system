import { useState } from "react"
import { useIdentity } from "@/components/identity"
import { usePeople } from "@/components/people-directory"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { canExtendTask } from "@/demo/task-collaboration"
import type { Project, Task } from "@/models/workspace"
import type { TaskDeadlineCommand } from "@/models/task-collaboration"

export function TaskDeadline({
  project,
  task,
  actorId,
  onExtend,
}: {
  project: Project
  task: Task
  actorId: string
  onExtend?: (command: TaskDeadlineCommand) => void
}) {
  const identity = useIdentity()
  const { getPerson } = usePeople()
  const [snapshot, setSnapshot] = useState<Task | null>(null)
  const [start, setStart] = useState("")
  const [end, setEnd] = useState("")
  const [reason, setReason] = useState("")
  const [error, setError] = useState("")
  const [discard, setDiscard] = useState(false)
  const allowed = onExtend && canExtendTask(project, task, actorId, identity)
  const date = (value?: string) => (value ? value.split("-").reverse().join("/") : "Chưa đặt")
  const close = () => {
    if (reason || start !== (snapshot?.start ?? "") || end !== (snapshot?.end ?? ""))
      setDiscard(true)
    else setSnapshot(null)
  }
  return (
    <section className="mt-4 border-t border-border pt-4 text-sm" aria-label="Kế hoạch công việc">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h3 className="section-title">Thời hạn công việc</h3>
        {allowed && (
          <Button
            size="sm"
            variant="outline"
            onClick={() => {
              setSnapshot(structuredClone(task))
              setStart(task.start ?? "")
              setEnd(task.end ?? "")
              setReason("")
              setError("")
            }}
          >
            {task.end ? "Gia hạn" : "Đặt thời hạn"}
          </Button>
        )}
      </div>
      <dl className="mt-3 space-y-2 text-xs">
        <div className="flex justify-between gap-3">
          <dt className="field-label">Hiện tại</dt>
          <dd>
            {date(task.start)} – {date(task.end)}
          </dd>
        </div>
        <div className="flex justify-between gap-3">
          <dt className="field-label">Kế hoạch ban đầu</dt>
          <dd>
            {date(task.originalStart ?? task.start)} – {date(task.originalEnd ?? task.end)}
          </dd>
        </div>
      </dl>
      {!!task.planHistory?.length && (
        <details className="mt-3 text-xs">
          <summary className="cursor-pointer font-medium">
            Lịch sử điều chỉnh ({task.planHistory.length})
          </summary>
          <ol className="mt-2 space-y-3">
            {task.planHistory.map((event, index) => (
              <li key={index} className="rounded-md bg-accent p-3 leading-5">
                <p>
                  {getPerson(event.actorId)?.name} · {new Date(event.at).toLocaleString("vi-VN")}
                </p>
                <p>
                  {date(event.before.end)} → {date(event.after.end)}
                </p>
                <p className="whitespace-pre-wrap break-words">{event.reason}</p>
              </li>
            ))}
          </ol>
        </details>
      )}
      {snapshot && (
        <Dialog
          open
          title={snapshot.end ? "Gia hạn công việc" : "Đặt thời hạn công việc"}
          description="Giữ nguyên ước tính và đặc tả. Ngày công việc phải nằm trong kế hoạch dự án."
          onOpenChange={(open) => !open && close()}
        >
          <form
            className="space-y-4"
            onSubmit={(event) => {
              event.preventDefault()
              try {
                onExtend?.({
                  taskId: snapshot.id,
                  beforeStart: snapshot.start,
                  beforeEnd: snapshot.end,
                  start,
                  end,
                  reason,
                })
                setSnapshot(null)
              } catch (err) {
                setError((err as Error).message)
              }
            }}
          >
            <p className="text-sm">
              {task.code} · Hạn dự án: {date(project.end)}
            </p>
            <label className="block">
              <span className="field-label">Ngày bắt đầu</span>
              <input
                type="date"
                className="field mt-2 w-full"
                required
                readOnly={Boolean(snapshot.start)}
                min={project.start}
                max={project.end}
                value={start}
                onChange={(event) => setStart(event.target.value)}
              />
            </label>
            <label className="block">
              <span className="field-label">Hạn mới</span>
              <input
                type="date"
                className="field mt-2 w-full"
                required
                min={snapshot.end ?? start ?? project.start}
                max={project.end}
                value={end}
                onChange={(event) => setEnd(event.target.value)}
              />
            </label>
            <label className="block">
              <span className="field-label">Lý do</span>
              <textarea
                className="field mt-2 min-h-20 w-full"
                required
                value={reason}
                onChange={(event) => setReason(event.target.value)}
              />
            </label>
            {error && (
              <p role="alert" className="text-sm text-destructive">
                {error}
              </p>
            )}
            <div className="flex justify-end gap-2">
              <Button type="button" variant="outline" onClick={close}>
                Hủy
              </Button>
              <Button type="submit">Lưu thời hạn</Button>
            </div>
          </form>
        </Dialog>
      )}
      <Dialog
        open={discard}
        onOpenChange={setDiscard}
        title="Bỏ thay đổi thời hạn?"
        description="Thời hạn đã lưu và lịch sử điều chỉnh được giữ nguyên."
      >
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={() => setDiscard(false)}>
            Tiếp tục chỉnh sửa
          </Button>
          <Button
            onClick={() => {
              setDiscard(false)
              setSnapshot(null)
            }}
          >
            Bỏ thay đổi
          </Button>
        </div>
      </Dialog>
    </section>
  )
}
