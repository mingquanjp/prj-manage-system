import { useIdentity } from "@/components/identity"
import { usePeople } from "@/components/people-directory"
import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { SearchInput } from "@/components/shared"
import { isEligible } from "@/demo/identity"
import { loadFor, skillGaps } from "@/demo/assignment"
import { reassignmentBlockReason, type ReassignmentCommand } from "@/demo/reassignment"
import type { WorkState } from "@/demo/effort"
import type { Choice, Project, Task } from "@/models/workspace"
import { AssignmentCandidateTable } from "./AssignmentCandidateTable"

export function TaskReassignment({
  project,
  task,
  tasks,
  work,
  actorId,
  onReassign,
}: {
  project: Project
  task: Task
  tasks: Task[]
  work: WorkState
  actorId: string
  onReassign: (command: ReassignmentCommand) => void
}) {
  const identity = useIdentity()

  const { people, getPerson } = usePeople()
  const [expectedAssignmentId, setExpectedAssignmentId] = useState<string | null>(null)
  const [personId, setPersonId] = useState("")
  const [query, setQuery] = useState("")
  const [reason, setReason] = useState("")
  const [overrideReason, setOverrideReason] = useState("")
  const [error, setError] = useState("")
  const [confirmDiscard, setConfirmDiscard] = useState(false)
  const block = reassignmentBlockReason(project, task, work, actorId, identity)
  const person = getPerson(personId)
  const gaps = person ? skillGaps(task, person) : []
  const chosen: Choice | undefined = person
    ? { personId, mode: "manual", reason: "", overrideReason, hadRun: false }
    : undefined
  const candidates = people.filter(
    (p) =>
      isEligible(project, p) &&
      p.id !== task.assigneeId &&
      `${p.name} ${p.position}`
        .toLocaleLowerCase("vi")
        .includes(query.trim().toLocaleLowerCase("vi")),
  )
  const history = work.contributions.filter(
    (entry) => entry.taskId === task.id && entry.assignment.endReason?.startsWith("Đổi người:"),
  )
  const close = () => {
    setExpectedAssignmentId(null)
    setConfirmDiscard(false)
  }
  const requestClose = () => {
    if (personId || reason.trim() || overrideReason.trim()) setConfirmDiscard(true)
    else close()
  }

  return (
    <div className="mt-3">
      {!block && (
        <Button
          size="sm"
          variant="outline"
          onClick={() => {
            setPersonId("")
            setQuery("")
            setReason("")
            setOverrideReason("")
            setError("")
            setExpectedAssignmentId(task.currentAssignment!.id)
          }}
        >
          Đổi người thực hiện
        </Button>
      )}
      {history.length > 0 && (
        <details className="mt-3 text-xs">
          <summary className="field-label cursor-pointer">
            Lịch sử đổi người ({history.length})
          </summary>
          <ul className="mt-2 space-y-3">
            {history.map(({ assignment }) => (
              <li key={assignment.id} className="break-words leading-5 text-muted-foreground">
                <p className="whitespace-pre-wrap">{assignment.endReason}</p>
                <p>
                  {getPerson(assignment.endedBy)?.name} ·{" "}
                  {new Date(assignment.endedAt!).toLocaleString("vi-VN")}
                </p>
              </li>
            ))}
          </ul>
        </details>
      )}
      <Dialog
        wide
        open={expectedAssignmentId !== null}
        onOpenChange={(open) => {
          if (!open) requestClose()
        }}
        title="Đổi người thực hiện"
        description={`${task.code} · ${task.name}. Người hiện tại vẫn phụ trách cho tới khi bạn xác nhận thành công.`}
      >
        <p className="mb-4 text-sm">
          Hiện tại: <strong>{getPerson(task.assigneeId)?.name}</strong> · {task.estimate} giờ ước
          tính
        </p>
        <SearchInput value={query} onChange={setQuery} placeholder="Tìm người thay thế…" />
        <p className="my-3 text-xs text-muted-foreground">
          Tra cứu thủ công toàn công ty, không tính điểm hoặc xếp hạng. Mở Đối chiếu kỹ năng để so
          sánh yêu cầu.
        </p>
        <div className="max-h-64 overflow-y-auto rounded-md border border-border">
          <AssignmentCandidateTable
            task={task}
            tasks={tasks}
            candidates={candidates}
            ranked={[]}
            chosen={chosen}
            mode="manual"
            draft={chosen ? { [task.id]: chosen } : {}}
            query={query}
            onSelectPerson={(id) => {
              if (id !== personId) {
                setPersonId(id)
                setOverrideReason("")
                setError("")
              }
            }}
          />
        </div>
        <form
          className="mt-5 space-y-4"
          onSubmit={(event) => {
            event.preventDefault()
            try {
              if (!person) throw new Error("Vui lòng chọn người thay thế.")
              onReassign({
                projectId: project.id,
                taskId: task.id,
                expectedAssignmentId: expectedAssignmentId!,
                personId,
                reason,
                overrideReason,
              })
              close()
            } catch (err) {
              setError(err instanceof Error ? err.message : "Chưa thể đổi người.")
            }
          }}
        >
          {person && (
            <p className="rounded-md bg-accent p-3 text-sm leading-6" aria-live="polite">
              Đang chọn: <strong>{person.name}</strong>. Sau xác nhận, tải người mới là{" "}
              {loadFor(personId, tasks, identity) + task.estimate} giờ; tải{" "}
              {getPerson(task.assigneeId)?.name} còn{" "}
              {loadFor(task.assigneeId!, tasks, identity) - task.estimate} giờ. Người mới cần bấm
              Bắt đầu trước khi ghi nhật ký.
            </p>
          )}
          <label className="block">
            <span className="field-label">
              Lý do đổi người <span className="text-warning">*</span>
            </span>
            <textarea
              className="field workspace-field mt-2 min-h-20 w-full"
              value={reason}
              onChange={(event) => setReason(event.target.value)}
            />
          </label>
          {gaps.length > 0 && (
            <label className="block">
              <span className="field-label text-warning">
                Chưa đáp ứng kỹ năng bắt buộc — lý do ngoại lệ *
              </span>
              <textarea
                className="field workspace-field mt-2 min-h-20 w-full"
                value={overrideReason}
                onChange={(event) => setOverrideReason(event.target.value)}
              />
            </label>
          )}
          {(error || block) && (
            <p role="alert" className="text-sm text-warning">
              {error || block}
            </p>
          )}
          <div className="flex flex-wrap justify-end gap-2 border-t border-border pt-4">
            <Button variant="outline" onClick={requestClose}>
              Hủy
            </Button>
            <Button type="submit" disabled={!person || Boolean(block)}>
              Xác nhận đổi người
            </Button>
          </div>
        </form>
      </Dialog>
      <Dialog
        open={confirmDiscard}
        onOpenChange={setConfirmDiscard}
        title="Bỏ thay đổi chưa lưu?"
        description="Phân công hiện tại vẫn được giữ nguyên."
      >
        <div className="flex flex-wrap justify-end gap-2">
          <Button variant="outline" onClick={() => setConfirmDiscard(false)}>
            Tiếp tục chọn
          </Button>
          <Button onClick={close}>Bỏ thay đổi</Button>
        </div>
      </Dialog>
    </div>
  )
}
