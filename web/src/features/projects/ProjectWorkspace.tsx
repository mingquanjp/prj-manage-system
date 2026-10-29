import { toggleTaskSelection } from "./task-selection"
import { useAssignmentDraftGuard } from "@/app/NavigationGuard"
import type { PmChangeCommand, TaskDeadlineCommand, TaskComment } from "@/models/task-collaboration"
import { useIdentity } from "@/components/identity"
import { ProjectTaskTable } from "./ProjectTaskTable"
import { ProjectTaskToolbar } from "./ProjectTaskToolbar"
import { ProjectTaskSidebar } from "./ProjectTaskSidebar"
import { useRef, useState } from "react"
import { Link, useParams, useSearchParams } from "react-router"
import {
  ArrowLeftIcon,
  ChevronDownIcon,
  CubeIcon,
  InfoCircledIcon,
  ListBulletIcon,
  LayersIcon,
  PlusIcon,
  CheckCircledIcon,
} from "@radix-ui/react-icons"
import { EmptyState, StatusBadge } from "@/components/shared"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { canAssign, canViewProject, isDirector } from "@/demo/identity"
import { progress } from "@/demo/projects"
import type {
  Project,
  ProjectEvent,
  OutputConflict,
  Task,
  TaskGroup,
  Draft,
} from "@/models/workspace"
import { AssignmentWorkspace } from "@/features/assignment/AssignmentWorkspace"
import { ProjectStructure } from "./ProjectStructure"
import { TaskEditor } from "./TaskEditor"
import { ProjectResponsibilities } from "./ProjectResponsibilities"
import { ProjectKanban } from "./ProjectKanban"
import type { ReviewResult, WorkState } from "@/demo/effort"
import { cn } from "@/lib/utils"
import { CorrectionRequests } from "@/features/work-logs/CorrectionRequests"
import { contributionFor } from "@/demo/corrections"
import type { LifecycleCommand } from "@/demo/project-lifecycle"
import type { ReassignmentCommand } from "@/demo/reassignment"

export function ProjectWorkspace({
  actorId,
  projects,
  conflicts,
  events,
  onProjectChange,
  tasks,
  groups: allGroups,
  onGroupsChange,
  onTasksChange,
  onNotice,
  work,
  onReview,
  onWorkChange,
  onLifecycle,
  onReassign,
  onReplacePm,
  onExtendDeadline,
  comments,
  onPostComment,
}: {
  actorId: string
  projects: Project[]
  conflicts: OutputConflict[]
  events: ProjectEvent[]
  onProjectChange: (project: Project, action: string) => void
  tasks: Task[]
  groups: TaskGroup[]
  onGroupsChange: (groups: TaskGroup[]) => void
  onTasksChange: (tasks: Task[]) => void
  onNotice: (notice: string) => void
  work: WorkState
  onWorkChange: (work: WorkState) => void
  onLifecycle: (command: LifecycleCommand) => void
  onReplacePm: (command: PmChangeCommand) => void
  onExtendDeadline: (command: TaskDeadlineCommand) => void
  comments: TaskComment[]
  onPostComment: (taskId: string, content: string) => void
  onReassign: (command: ReassignmentCommand) => void
  onReview: (submissionId: string, result: ReviewResult, comment: string) => void
}) {
  const identity = useIdentity()

  const { projectId } = useParams()
  const [params] = useSearchParams()
  const project = projects.find((p) => p.id === projectId)
  const [search, setSearch] = useState("")
  const [filter, setFilter] = useState("all")
  const [group, setGroup] = useState("all")
  const [view, setView] = useState<"list" | "groups" | "kanban">("list")
  const [tab, setTab] = useState<"tasks" | "structure" | "reviews" | "corrections">("tasks")
  const [detailTab, setDetailTab] = useState<"details" | "results" | "discussion">(
    params.get("panel") === "results" ? "results" : "details",
  )
  const [editor, setEditor] = useState<{
    task?: Task
    groupId?: string
  } | null>(null)
  const [info, setInfo] = useState(false)
  const [detail, setDetail] = useState<string | null>(params.get("task"))
  const [selected, setSelected] = useState<string[]>([])
  const [draft, setDraft] = useState<Draft>({})
  const [discardTab, setDiscardTab] = useState<
    "tasks" | "structure" | "reviews" | "corrections" | null
  >(null)
  const [assignMode, setAssignMode] = useState<"manual" | "recommended" | null>(null)
  useAssignmentDraftGuard(Boolean(assignMode && Object.keys(draft).length))
  const detailTrigger = useRef<HTMLButtonElement | null>(null)
  if (!project || !canViewProject(project, actorId, identity))
    return (
      <EmptyState title="Không tìm thấy dự án" detail="Chọn dự án khác trong thanh điều hướng." />
    )
  const related = tasks.filter((t) => t.projectId === project.id)
  const groups = allGroups.filter((g) => g.projectId === project.id).map((g) => g.name)
  const filtered = related.filter(
    (t) =>
      `${t.name} ${t.code}`.toLocaleLowerCase("vi").includes(search.toLocaleLowerCase("vi")) &&
      (tab === "reviews"
        ? project.status === "IN_PROGRESS" && t.status === "PENDING_REVIEW"
        : filter === "all" || t.status === filter) &&
      (group === "all" || t.group === group),
  )
  const currentTask = related.find((t) => t.id === detail)
  const allowed = canAssign(project, actorId)
  const canSelectAssignments = allowed && tab === "tasks"
  const canReadResults = project.ptcmId === actorId || project.pmId === actorId
  const waiting =
    project.status === "IN_PROGRESS"
      ? related.filter((t) => t.status === "PENDING_REVIEW").length
      : 0
  const percent = progress(related)
  const openTab = (next: "tasks" | "structure" | "reviews" | "corrections") => {
    setDraft({})
    setAssignMode(null)
    setTab(next)
    setView("list")
    if (next !== "tasks") {
      setSelected([])
      setDetail(null)
    }
  }
  const requestTab = (next: "tasks" | "structure" | "reviews" | "corrections") => {
    if (assignMode && Object.keys(draft).length) setDiscardTab(next)
    else openTab(next)
  }
  const closeDetail = () => {
    setDetail(null)
    detailTrigger.current?.focus()
  }
  const toggle = (id: string) =>
    setSelected((prev) =>
      prev.includes(id) ? prev.filter((value) => value !== id) : [...prev, id],
    )
  const toggleAll = (rows: Task[]) => setSelected((prev) => toggleTaskSelection(rows, prev))
  const openTableTask = (task: Task, trigger: HTMLButtonElement) => {
    detailTrigger.current = trigger
    setDetail(task.id)
    setDetailTab(task.status === "PENDING_REVIEW" && canReadResults ? "results" : "details")
  }
  const transitionLifecycle = (command: LifecycleCommand) => {
    onLifecycle(command)
    setDraft({})
    setSelected([])
    setAssignMode(null)
  }

  return (
    <>
      <Link
        to="/"
        className="mb-3 inline-flex items-center gap-2 text-xs text-muted-foreground hover:text-foreground"
      >
        <ArrowLeftIcon />
        Tất cả dự án
      </Link>
      <div className="mb-3 flex flex-wrap items-center justify-between gap-3">
        <div className="flex min-w-0 items-center gap-3">
          <span className="grid size-10 shrink-0 place-items-center rounded-lg border border-border bg-white">
            <CubeIcon className="size-5" />
          </span>
          <div>
            <p className="eyebrow">{project.code}</p>
            <div className="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1">
              <h1 className="text-xl font-semibold text-balance sm:text-2xl">{project.name}</h1>
              <StatusBadge status={project.status} />
            </div>
          </div>
        </div>
        <div className="flex gap-2">
          {allowed && !assignMode && tab !== "reviews" && tab !== "corrections" && (
            <Button
              onClick={() => {
                if (!groups.length) {
                  setTab("structure")
                  onNotice(
                    "Bạn cần tạo nhóm công việc trước. Nhấn Tạo nhóm để bắt đầu, sau đó thêm công việc vào nhóm.",
                  )
                } else setEditor({})
              }}
            >
              <PlusIcon />
              Tạo công việc
            </Button>
          )}
          {tab === "tasks" && !assignMode && (
            <Button
              variant="outline"
              onClick={() => {
                setInfo(!info)
                setDetail(null)
              }}
              aria-pressed={info}
            >
              <InfoCircledIcon />
              Thông tin
            </Button>
          )}
        </div>
      </div>
      <ProjectResponsibilities
        onReplacePm={onReplacePm}
        project={project}
        actorId={actorId}
        tasks={tasks}
        conflicts={conflicts}
        events={events}
        onChange={onProjectChange}
        work={work}
        onLifecycle={transitionLifecycle}
        hasDraft={Object.keys(draft).length > 0}
      />
      <div className="mb-4 flex items-center justify-between gap-4 border-b border-border">
        <div className="flex min-w-0 overflow-x-auto">
          <button
            type="button"
            onClick={() => requestTab("tasks")}
            className={cn("tab-button workspace-tab", tab === "tasks" && "tab-selected")}
            aria-pressed={tab === "tasks"}
          >
            <ListBulletIcon />
            Công việc <span className="count">{related.length}</span>
          </button>
          <button
            type="button"
            className={cn("tab-button workspace-tab", tab === "structure" && "tab-selected")}
            aria-pressed={tab === "structure"}
            onClick={() => requestTab("structure")}
          >
            <LayersIcon />
            Cấu trúc <span className="count">{groups.length}</span>
          </button>
          {canReadResults && (
            <button
              type="button"
              onClick={() => requestTab("reviews")}
              className={cn("tab-button workspace-tab", tab === "reviews" && "tab-selected")}
              aria-pressed={tab === "reviews"}
            >
              <CheckCircledIcon />
              Nghiệm thu <span className="count">{waiting}</span>
            </button>
          )}
          {project.ptcmId === actorId && (
            <button
              type="button"
              onClick={() => requestTab("corrections")}
              className={cn("tab-button workspace-tab", tab === "corrections" && "tab-selected")}
              aria-pressed={tab === "corrections"}
            >
              Hiệu chỉnh nhật ký{" "}
              <span className="count">
                {
                  work.requests.filter(
                    (r) =>
                      r.status === "PENDING" &&
                      contributionFor(work, tasks, r.assignmentId)?.projectId === project.id,
                  ).length
                }
              </span>
            </button>
          )}
        </div>
        <span className="hidden text-xs text-muted-foreground md:block">
          Bạn đang có trách nhiệm{" "}
          {project.pmId === actorId && project.ptcmId === actorId
            ? "PM + PTCM"
            : project.ptcmId === actorId
              ? "PTCM"
              : project.pmId === actorId
                ? "PM · chỉ theo dõi chuyên môn"
                : isDirector(actorId, identity)
                  ? "Giám đốc · chỉ theo dõi"
                  : "Chỉ xem"}
        </span>
      </div>
      {tab === "corrections" && project.ptcmId === actorId ? (
        <CorrectionRequests
          key={project.id}
          projects={projects}
          tasks={tasks}
          work={work}
          actorId={actorId}
          projectId={project.id}
          onChange={onWorkChange}
        />
      ) : tab === "structure" ? (
        <ProjectStructure
          project={project}
          actorId={actorId}
          groups={allGroups}
          tasks={tasks}
          onChange={(nextGroups, nextTasks) => {
            onGroupsChange(nextGroups)
            onTasksChange(nextTasks)
          }}
          onNotice={onNotice}
          onCreateTask={(groupId) => setEditor({ groupId })}
          onEditTask={(task) => setEditor({ task })}
          onViewTask={(task) => {
            setTab("tasks")
            setDetail(task.id)
            setDetailTab("details")
            setInfo(false)
            setSearch("")
            setFilter("all")
            setGroup("all")
          }}
        />
      ) : assignMode && allowed ? (
        <AssignmentWorkspace
          project={project}
          actorId={actorId}
          tasks={tasks}
          selectedIds={selected}
          initialMode={assignMode}
          draft={draft}
          onDraftChange={setDraft}
          onBack={() => requestTab("tasks")}
          onCommit={(next) => {
            onTasksChange(next)
            setAssignMode(null)
            setDraft({})
            setSelected([])
            onNotice(
              "Đã xác nhận phân công mô phỏng. Tải và danh sách công việc đã được cập nhật trong phiên.",
            )
          }}
        />
      ) : (
        <>
          <div
            className={cn(
              "grid items-start gap-6",
              (info || currentTask) && "xl:grid-cols-[minmax(0,1fr)_360px]",
            )}
          >
            <section className="min-w-0">
              <ProjectTaskToolbar
                tab={tab}
                related={related}
                waiting={waiting}
                view={view}
                search={search}
                filter={filter}
                group={group}
                groups={groups}
                isPtcm={project.ptcmId === actorId}
                onViewChange={setView}
                onSearchChange={setSearch}
                onFilterChange={setFilter}
                onGroupChange={setGroup}
              />
              {view === "kanban" && (
                <>
                  {filtered.some((t) => ["CLOSED", "CANCELLED"].includes(t.status)) && (
                    <p className="mb-3 text-xs text-muted-foreground">
                      {filtered.filter((t) => ["CLOSED", "CANCELLED"].includes(t.status)).length}{" "}
                      công việc đã đóng/hủy không nằm trên Kanban.{" "}
                      <button type="button" className="underline" onClick={() => setView("list")}>
                        Xem danh sách
                      </button>
                    </p>
                  )}
                  <ProjectKanban
                    tasks={filtered}
                    selectedId={detail}
                    selectedAssignments={selected}
                    canSelect={canSelectAssignments}
                    onToggle={toggle}
                    onSelect={(task, trigger) => {
                      detailTrigger.current = trigger
                      setDetail(task.id)
                      setInfo(false)
                      setDetailTab(
                        task.status === "PENDING_REVIEW" && canReadResults ? "results" : "details",
                      )
                    }}
                  />
                </>
              )}
              {view !== "kanban" && (
                <div className="overflow-hidden rounded-lg border border-border bg-white">
                  {!filtered.length ? (
                    <EmptyState
                      title={
                        tab === "reviews"
                          ? "Không có công việc chờ nghiệm thu"
                          : related.length
                            ? "Không có công việc phù hợp"
                            : "Dự án chưa có công việc"
                      }
                      detail={
                        tab === "reviews"
                          ? "Thử thay đổi từ khóa/nhóm hoặc chờ người thực hiện nộp kết quả."
                          : related.length
                            ? "Thử thay đổi từ khóa hoặc bộ lọc."
                            : allowed
                              ? "Mở tab Cấu trúc để tạo nhóm, sau đó thêm công việc đầu tiên."
                              : "PTCM sẽ thiết lập nhóm và công việc cho dự án."
                      }
                    />
                  ) : view === "list" ? (
                    <ProjectTaskTable
                      rows={filtered}
                      canSelectAssignments={canSelectAssignments}
                      selected={selected}
                      detail={detail}
                      onToggleAll={toggleAll}
                      onToggle={toggle}
                      onOpenTask={openTableTask}
                    />
                  ) : (
                    groups
                      .filter((g) => filtered.some((t) => t.group === g))
                      .map((g) => (
                        <details key={g} open className="border-b border-border last:border-0">
                          <summary className="field-label flex cursor-pointer list-none items-center gap-2 bg-background/70 px-4 py-3">
                            <ChevronDownIcon />
                            {g}
                            <span className="count">
                              {filtered.filter((t) => t.group === g).length}
                            </span>
                          </summary>
                          <ProjectTaskTable
                            rows={filtered.filter((t) => t.group === g)}
                            canSelectAssignments={canSelectAssignments}
                            selectionLabel={`Chọn tất cả công việc đủ điều kiện trong nhóm ${g}`}
                            selected={selected}
                            detail={detail}
                            onToggleAll={toggleAll}
                            onToggle={toggle}
                            onOpenTask={openTableTask}
                          />
                        </details>
                      ))
                  )}
                </div>
              )}
              <p className="mt-3 text-xs text-muted-foreground">
                {filtered.length} / {related.length} công việc · Nhấn tên để xem chi tiết mà không
                rời danh sách.
              </p>
            </section>
            {(info || currentTask) && (
              <ProjectTaskSidebar
                onExtendDeadline={onExtendDeadline}
                comments={comments}
                onPostComment={onPostComment}
                currentTask={currentTask}
                project={project}
                actorId={actorId}
                allowed={allowed}
                canReadResults={canReadResults}
                detailTab={detailTab}
                work={work}
                conflicts={conflicts}
                percent={percent}
                related={related}
                onCloseDetail={closeDetail}
                onDetailTabChange={setDetailTab}
                onEditTask={(task) => setEditor({ task })}
                onReview={onReview}
                tasks={tasks}
                onLifecycle={transitionLifecycle}
                onReassign={onReassign}
              />
            )}
          </div>
          {selected.length > 0 && allowed && (
            <div className="sticky bottom-4 z-10 mt-6 flex flex-wrap items-center justify-between gap-3 rounded-lg border border-border bg-primary px-4 py-3 text-white shadow-lg">
              <span className="text-sm">
                <strong>{selected.length}</strong> công việc đã chọn{" "}
                <button
                  type="button"
                  className="ml-3 underline underline-offset-4 text-white/75"
                  onClick={() => setSelected([])}
                >
                  Bỏ chọn
                </button>
              </span>
              <div className="flex flex-wrap gap-2">
                <Button
                  variant="outline"
                  className="text-foreground"
                  onClick={() => setAssignMode("manual")}
                >
                  Phân công thủ công
                </Button>
                <Button
                  className="border border-white/30 bg-white/10 hover:bg-white/20"
                  onClick={() => setAssignMode("recommended")}
                >
                  Xem gợi ý<ChevronDownIcon />
                </Button>
              </div>
            </div>
          )}
        </>
      )}
      <Dialog
        open={discardTab !== null}
        onOpenChange={(open) => {
          if (!open) setDiscardTab(null)
        }}
        title="Bỏ bản nháp phân công?"
        description="Các lựa chọn chưa xác nhận sẽ không được lưu. Công việc vẫn giữ nguyên người phụ trách và trạng thái hiện tại."
      >
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={() => setDiscardTab(null)}>
            Tiếp tục chọn
          </Button>
          <Button
            onClick={() => {
              if (discardTab) openTab(discardTab)
              setDiscardTab(null)
            }}
          >
            Bỏ bản nháp
          </Button>
        </div>
      </Dialog>
      {editor && allowed && (
        <TaskEditor
          project={project}
          actorId={actorId}
          groups={allGroups}
          tasks={tasks}
          task={editor.task}
          defaultGroupId={editor.groupId}
          onClose={() => setEditor(null)}
          onSave={(nextTasks, taskId) => {
            onTasksChange(nextTasks)
            setEditor(null)
            setTab("tasks")
            setAssignMode(null)
            setSearch("")
            setFilter("all")
            setGroup("all")
            setSelected([])
            setDetail(taskId)
            onNotice(
              editor.task
                ? "Đã lưu thay đổi công việc trong phiên mô phỏng."
                : "Đã tạo công việc chờ phân công trong phiên mô phỏng.",
            )
          }}
        />
      )}
    </>
  )
}
