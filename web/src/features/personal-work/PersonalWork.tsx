import { TaskDiscussion } from "@/features/task-collaboration/TaskDiscussion"
import { TaskDeadline } from "@/features/task-collaboration/TaskDeadline"
import type { TaskComment } from "@/models/task-collaboration"
import { useIdentity } from "@/components/identity"
import { usePeople } from "@/components/people-directory"
import { useRef, useState } from "react"
import {
  FileTextIcon,
  PlayIcon,
  PauseIcon,
  ListBulletIcon,
  ViewVerticalIcon,
} from "@radix-ui/react-icons"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { EmptyState, SearchInput, StatusBadge } from "@/components/shared"
import { TaskDetailHeader, TaskSpecification } from "@/components/task-detail"
import { executionBlockReason, personalTasks } from "@/demo/execution"
import { statusLabels } from "@/models/workspace-labels"
import type { ExecutionAction, Project, Task, TaskStatus } from "@/models/workspace"
import { cn } from "@/lib/utils"
import { PersonalKanban } from "./PersonalKanban"
import { WorkLogPanel } from "@/features/work-logs/WorkLogPanel"
import { ResultPanel } from "./ResultPanel"
import type { WorkState } from "@/demo/effort"
import { ContributionHistory } from "@/features/work-logs/ContributionHistory"

export function PersonalWork({
  projects,
  tasks,
  actorId,
  onTransition,
  work,
  onWorkChange,
  onSubmitResult,
  comments,
  onPostComment,
}: {
  comments: TaskComment[]
  onPostComment: (taskId: string, content: string) => void
  projects: Project[]
  tasks: Task[]
  actorId: string
  work: WorkState
  onWorkChange: (work: WorkState) => void
  onSubmitResult: (taskId: string, summary: string, links: string) => void
  onTransition: (
    taskId: string,
    expected: TaskStatus,
    action: ExecutionAction,
    reason?: string,
  ) => void
}) {
  const identity = useIdentity()

  const { getPerson } = usePeople()
  const [search, setSearch] = useState("")
  const [projectId, setProjectId] = useState("all")
  const [status, setStatus] = useState("all")
  const [view, setView] = useState<"list" | "kanban">("list")
  const [detailId, setDetailId] = useState<string | null>(null)
  const [detailTab, setDetailTab] = useState<"details" | "logs" | "results" | "discussion">(
    "details",
  )
  const [pausing, setPausing] = useState(false)
  const [reason, setReason] = useState("")
  const [error, setError] = useState("")
  const trigger = useRef<HTMLButtonElement | null>(null)
  const viewControls = useRef<HTMLDivElement | null>(null)
  const own = personalTasks(projects, tasks, actorId, identity)
  const availableProjects = projects.filter((p) => own.some((t) => t.projectId === p.id))
  const filtered = own.filter(
    (t) =>
      `${t.code} ${t.name}`.toLocaleLowerCase("vi").includes(search.toLocaleLowerCase("vi")) &&
      (projectId === "all" || t.projectId === projectId) &&
      (status === "all" || t.status === status),
  )
  const detail = filtered.find((t) => t.id === detailId)
  const archivedCount = filtered.filter(
    (t) => t.status === "CLOSED" || t.status === "CANCELLED",
  ).length
  const project = projects.find((p) => p.id === detail?.projectId)
  const block =
    detail && project ? executionBlockReason(project, detail, actorId, identity) : undefined
  const perform = (action: ExecutionAction) => {
    if (!detail) return
    try {
      onTransition(detail.id, detail.status, action, reason)
      setError("")
      setPausing(false)
      setReason("")
    } catch (err) {
      setError((err as Error).message)
    }
  }
  const closeDetail = () => {
    setDetailId(null)
    setError("")
    if (trigger.current?.isConnected) trigger.current.focus()
    else
      viewControls.current?.querySelector<HTMLButtonElement>('button[aria-pressed="true"]')?.focus()
  }
  const date = (value?: string) =>
    value ? new Date(`${value}T00:00:00`).toLocaleDateString("vi-VN") : "Chưa đặt"
  return (
    <>
      <p className="eyebrow">KHÔNG GIAN CÁ NHÂN</p>
      <h1 className="mt-2 text-3xl font-semibold text-balance">Công việc của tôi</h1>
      <p className="mt-3 text-sm leading-6 text-muted-foreground">
        Theo dõi các công việc được giao và cập nhật tiến độ thực hiện.
      </p>
      <div className="mt-7 flex flex-wrap items-end gap-3">
        <div className="min-w-0 flex-1 basis-60">
          <SearchInput
            value={search}
            onChange={setSearch}
            placeholder="Tìm tên hoặc mã công việc…"
          />
        </div>
        <label className="grid gap-1.5">
          <span className="field-label">Dự án</span>
          <select
            className="field workspace-field max-w-full sm:max-w-64"
            value={projectId}
            onChange={(e) => setProjectId(e.target.value)}
          >
            <option value="all">Tất cả dự án</option>
            {availableProjects.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name}
              </option>
            ))}
          </select>
        </label>
        <label className="grid gap-1.5">
          <span className="field-label">Trạng thái công việc</span>
          <select
            className="field workspace-field"
            value={status}
            onChange={(e) => setStatus(e.target.value)}
          >
            <option value="all">Tất cả trạng thái</option>
            {Object.entries(statusLabels)
              .filter(([key]) => key !== "PENDING_ASSIGNMENT")
              .map(([key, label]) => (
                <option key={key} value={key}>
                  {label}
                </option>
              ))}
          </select>
        </label>
      </div>
      <div className="mt-5 flex flex-wrap items-center justify-between gap-3">
        <div
          role="group"
          aria-label="Cách xem công việc"
          ref={viewControls}
          className="inline-flex rounded-md border border-border bg-white p-1"
        >
          <Button
            size="sm"
            variant="ghost"
            aria-pressed={view === "list"}
            className={cn("workspace-tab", view === "list" && "bg-accent")}
            onClick={() => setView("list")}
          >
            <ListBulletIcon />
            Danh sách
          </Button>
          <Button
            size="sm"
            variant="ghost"
            aria-pressed={view === "kanban"}
            className={cn("workspace-tab", view === "kanban" && "bg-accent")}
            onClick={() => setView("kanban")}
          >
            <ViewVerticalIcon />
            Kanban
          </Button>
        </div>
        {view === "kanban" && (
          <p className="text-xs text-muted-foreground">
            Nhấn thẻ để xem chi tiết. Thay đổi trạng thái bằng nút thao tác, không kéo thả.
          </p>
        )}
      </div>
      <div
        className={cn(
          "mt-5 grid items-start gap-5",
          detail && "xl:grid-cols-[minmax(0,1fr)_360px]",
        )}
      >
        <section className="min-w-0" aria-label="Danh sách công việc cá nhân">
          <div className="overflow-hidden rounded-lg border border-border bg-white">
            {!filtered.length ? (
              <EmptyState
                title={
                  own.length ? "Không có công việc phù hợp" : "Bạn chưa có công việc được giao"
                }
                detail={
                  own.length
                    ? "Thử đổi từ khóa hoặc bộ lọc để xem các công việc khác."
                    : "Công việc sẽ xuất hiện tại đây sau khi PTCM xác nhận phân công. PTCM không nhận việc trong chính dự án mình phụ trách."
                }
              />
            ) : view === "kanban" ? (
              <PersonalKanban
                tasks={filtered}
                projects={projects}
                selectedId={detailId}
                onSelect={(task, element) => {
                  trigger.current = element
                  setDetailId(task.id)
                  setDetailTab("details")
                  setError("")
                }}
              />
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full min-w-[660px] text-left text-sm">
                  <thead>
                    <tr className="table-heading table-heading-strong">
                      <th className="px-4 py-3">Công việc</th>
                      <th className="pr-4">Dự án</th>
                      <th className="pr-4">Trạng thái</th>
                      <th className="pr-4 text-right">Ước tính</th>
                      <th className="pr-4">Hạn kế hoạch</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filtered.map((task) => (
                      <tr
                        key={task.id}
                        className={cn(
                          "border-t border-border hover:bg-background",
                          task.id === detailId && "bg-accent/60",
                        )}
                      >
                        <td className="px-4 py-4">
                          <button
                            className="flex items-start gap-2 text-left hover:underline"
                            aria-label={`Xem ${task.name}`}
                            aria-pressed={detailId === task.id}
                            onClick={(e) => {
                              trigger.current = e.currentTarget
                              setDetailId(task.id)
                              setDetailTab("details")
                              setError("")
                            }}
                          >
                            <FileTextIcon className="mt-1 shrink-0 text-muted-foreground" />
                            <span>
                              <span className="block font-medium">{task.name}</span>
                              <span className="mt-1 block text-xs text-muted-foreground">
                                {task.code} · {task.group}
                              </span>
                            </span>
                          </button>
                        </td>
                        <td className="max-w-48 pr-4">
                          {projects.find((p) => p.id === task.projectId)?.name}
                        </td>
                        <td className="pr-4">
                          <StatusBadge status={task.status} />
                        </td>
                        <td className="pr-4 text-right whitespace-nowrap">{task.estimate} giờ</td>
                        <td className="pr-4 whitespace-nowrap">{date(task.end)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
          {view === "kanban" && archivedCount > 0 && (
            <div className="mt-3 flex flex-wrap items-center gap-2 rounded-md border border-border bg-white px-3 py-2 text-xs text-muted-foreground">
              <span>{archivedCount} công việc Đã đóng/Đã hủy không hiển thị trên Kanban.</span>
              <Button size="sm" variant="ghost" onClick={() => setView("list")}>
                Xem trong danh sách
              </Button>
            </div>
          )}
          <p className="mt-3 text-xs text-muted-foreground">
            {filtered.length} / {own.length} công việc · Chỉ hiển thị công việc bạn được giao, không
            mở quyền xem toàn bộ dự án.
          </p>
        </section>
        {detail && project && (
          <aside
            aria-label="Chi tiết công việc cá nhân"
            className="min-w-0 rounded-lg border border-border bg-white p-5 xl:sticky xl:top-6 xl:max-h-[calc(100dvh-8rem)] xl:overflow-y-auto"
            onKeyDown={(e) => {
              if (
                e.key === "Escape" &&
                !pausing &&
                !(e.target instanceof Element && e.target.closest('[role="dialog"]'))
              )
                closeDetail()
            }}
          >
            <TaskDetailHeader task={detail} onClose={closeDetail} />
            <div
              role="group"
              aria-label="Nội dung công việc"
              className="mt-3 flex border-b border-border"
            >
              {(
                [
                  ["details", "Chi tiết"],
                  ["logs", "Nhật ký"],
                  ["results", "Kết quả"],
                  ["discussion", "Trao đổi"],
                ] as const
              ).map(([id, label]) => (
                <button
                  key={id}
                  type="button"
                  aria-pressed={detailTab === id}
                  onClick={() => setDetailTab(id)}
                  className={cn(
                    "tab-button workspace-tab flex-1 justify-center px-2",
                    detailTab === id && "tab-selected",
                  )}
                >
                  {label}
                </button>
              ))}
            </div>
            {detailTab === "discussion" && (
              <TaskDiscussion
                key={detail.id}
                task={detail}
                project={project}
                actorId={actorId}
                comments={comments}
                onPost={onPostComment}
              />
            )}
            {detailTab === "logs" && (
              <WorkLogPanel
                key={detail.id}
                project={project}
                task={detail}
                projects={projects}
                tasks={tasks}
                actorId={actorId}
                work={work}
                onChange={onWorkChange}
              />
            )}
            {detailTab === "details" && (
              <TaskDeadline
                key={`plan-${detail.id}`}
                task={detail}
                project={project}
                actorId={actorId}
              />
            )}
            {detailTab === "results" && (
              <ResultPanel
                key={detail.id}
                task={detail}
                project={project}
                actorId={actorId}
                work={work}
                onSubmit={onSubmitResult}
              />
            )}
            {detailTab === "details" && (
              <>
                <dl className="mt-5 space-y-3 text-sm">
                  <div>
                    <dt className="field-label">Dự án</dt>
                    <dd className="mt-1">
                      {project.name} · <StatusBadge status={project.status} />
                    </dd>
                  </div>
                  <div>
                    <dt className="field-label">Nhóm / PTCM</dt>
                    <dd className="mt-1">
                      {detail.group} / {getPerson(project.ptcmId)?.name ?? "Chưa chỉ định"}
                    </dd>
                  </div>
                  <div>
                    <dt className="field-label">Ước tính / Ưu tiên</dt>
                    <dd className="mt-1">
                      {detail.estimate} giờ / {detail.priority}
                    </dd>
                  </div>
                  <div>
                    <dt className="field-label">Ngày kế hoạch</dt>
                    <dd className="mt-1">
                      {date(detail.start)} — {date(detail.end)}
                    </dd>
                  </div>
                </dl>
                <div className="mt-5 border-t border-border pt-4">
                  {block ? (
                    <p className="text-sm leading-6 text-muted-foreground">{block}</p>
                  ) : (
                    <>
                      {detail.status === "ASSIGNED" && (
                        <Button onClick={() => perform("start")}>
                          <PlayIcon />
                          Bắt đầu công việc
                        </Button>
                      )}
                      {detail.status === "IN_PROGRESS" && (
                        <Button
                          variant="outline"
                          onClick={() => {
                            setReason("")
                            setError("")
                            setPausing(true)
                          }}
                        >
                          <PauseIcon />
                          Tạm dừng
                        </Button>
                      )}
                      {detail.status === "PAUSED" && (
                        <Button onClick={() => perform("resume")}>
                          <PlayIcon />
                          Tiếp tục công việc
                        </Button>
                      )}
                    </>
                  )}
                  {!pausing && error && (
                    <p role="alert" className="mt-3 text-sm text-destructive">
                      {error}
                    </p>
                  )}
                  <p className="mt-3 text-xs leading-5 text-muted-foreground">
                    Cập nhật trạng thái không tự ghi thời gian. Mở Nhật ký để ghi phần đã làm; mở
                    Kết quả để nộp khi công việc đang thực hiện.
                  </p>
                </div>
                <TaskSpecification key={detail.id} task={detail} />
                {!!detail.executionHistory?.length && (
                  <details className="mt-5 border-t border-border pt-4 text-sm">
                    <summary className="field-label cursor-pointer">
                      Cập nhật trong phiên mô phỏng
                    </summary>
                    <ul className="mt-3 space-y-3">
                      {detail.executionHistory.map((event, i) => (
                        <li key={i} className="text-xs leading-5 text-muted-foreground">
                          {statusLabels[event.from]} → {statusLabels[event.to]}
                          <br />
                          {getPerson(event.actorId)?.name} ·{" "}
                          {new Date(event.at).toLocaleString("vi-VN")}
                          {event.reason && (
                            <p className="mt-1 whitespace-pre-wrap break-words text-foreground">
                              Lý do: {event.reason}
                            </p>
                          )}
                        </li>
                      ))}
                    </ul>
                  </details>
                )}
              </>
            )}
          </aside>
        )}
      </div>
      <ContributionHistory
        projects={projects}
        tasks={tasks}
        work={work}
        actorId={actorId}
        onChange={onWorkChange}
      />
      <Dialog
        open={pausing && !!detail}
        onOpenChange={(open) => {
          setPausing(open)
          if (!open) setError("")
        }}
        title="Tạm dừng công việc"
        description="Ghi rõ lý do để PTCM theo dõi. Thao tác này không tự ghi thời gian làm việc."
      >
        <form
          onSubmit={(e) => {
            e.preventDefault()
            perform("pause")
          }}
        >
          <label htmlFor="pause-reason" className="field-label">
            Lý do tạm dừng *
          </label>
          <textarea
            id="pause-reason"
            className="field mt-2 min-h-28 w-full"
            value={reason}
            onChange={(e) => {
              setReason(e.target.value)
              setError("")
            }}
            aria-invalid={!!error}
            aria-describedby={error ? "pause-error" : undefined}
          />
          {error && (
            <p id="pause-error" role="alert" className="mt-2 text-sm text-destructive">
              {error}
            </p>
          )}
          <div className="mt-5 flex justify-end gap-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => {
                setPausing(false)
                setError("")
              }}
            >
              Hủy
            </Button>
            <Button type="submit">Xác nhận tạm dừng</Button>
          </div>
        </form>
      </Dialog>
    </>
  )
}
