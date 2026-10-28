import { useState } from "react"
import { Button } from "@/components/ui/button"
import { useIdentity } from "@/components/identity"
import { usePeople } from "@/components/people-directory"
import { canReadTaskDiscussion, canPostTaskDiscussion } from "@/demo/task-collaboration"
import type { Project, Task } from "@/models/workspace"
import type { TaskComment } from "@/models/task-collaboration"

export function TaskDiscussion({
  project,
  task,
  actorId,
  comments,
  onPost,
}: {
  project: Project
  task: Task
  actorId: string
  comments: TaskComment[]
  onPost: (taskId: string, content: string) => void
}) {
  const identity = useIdentity()
  const { getPerson } = usePeople()
  const [content, setContent] = useState("")
  const [error, setError] = useState("")
  if (!canReadTaskDiscussion(project, task, actorId, identity)) return null
  const entries = comments.filter((entry) => entry.taskId === task.id)
  const writable = canPostTaskDiscussion(project, task, actorId, identity)
  return (
    <section className="mt-4 space-y-4" aria-label="Trao đổi công việc">
      <div>
        <h3 className="section-title">Trao đổi với phụ trách chuyên môn</h3>
        <p className="mt-2 text-xs leading-5 text-muted-foreground">
          Hỏi, báo vướng mắc và hướng dẫn xử lý. Trao đổi không thay đổi đặc tả hoặc quyết định
          nghiệm thu.
        </p>
      </div>
      {!entries.length && (
        <p className="rounded-md bg-accent p-3 text-sm text-muted-foreground">
          Chưa có trao đổi. Bạn có thể nêu khó khăn hoặc câu hỏi tại đây.
        </p>
      )}
      <ol className="space-y-3">
        {entries.map((entry) => (
          <li key={entry.id} className="rounded-lg border border-border p-3">
            <div className="flex flex-wrap items-center justify-between gap-2 text-xs">
              <strong className="font-medium">
                {getPerson(entry.actorId)?.name ?? "Nhân viên"}
              </strong>
              <time dateTime={entry.at} className="text-muted-foreground">
                {new Date(entry.at).toLocaleString("vi-VN")}
              </time>
            </div>
            <p className="mt-2 whitespace-pre-wrap break-words text-sm leading-6">
              {entry.content}
            </p>
          </li>
        ))}
      </ol>
      {writable ? (
        <form
          className="space-y-3"
          onSubmit={(event) => {
            event.preventDefault()
            try {
              onPost(task.id, content)
              setContent("")
              setError("")
            } catch (err) {
              setError((err as Error).message)
            }
          }}
        >
          <label className="block">
            <span className="field-label">Nội dung trao đổi</span>
            <textarea
              className="field mt-2 min-h-24 w-full"
              required
              value={content}
              onChange={(event) => {
                setContent(event.target.value)
                setError("")
              }}
              placeholder="Bạn đang gặp vướng mắc gì?"
            />
          </label>
          {error && (
            <p role="alert" className="text-sm text-destructive">
              {error}
            </p>
          )}
          <div className="flex justify-end">
            <Button type="submit" size="sm" disabled={!content.trim()}>
              Gửi trao đổi
            </Button>
          </div>
        </form>
      ) : (
        <p className="text-xs text-muted-foreground">
          Công việc hoặc dự án đã kết thúc. Trao đổi chỉ đọc.
        </p>
      )}
    </section>
  )
}
