import { useRef, useState } from "react"
import { Link, useParams } from "react-router"
import {
  ArrowLeftIcon,
  ChevronDownIcon,
  CubeIcon,
  FileTextIcon,
  InfoCircledIcon,
  ListBulletIcon,
  MixerHorizontalIcon,
  RowsIcon,
  LayersIcon,
  PlusIcon,
  Pencil1Icon,
  LockClosedIcon,
  ViewVerticalIcon,
  CheckCircledIcon,
} from "@radix-ui/react-icons"
import {
  Avatar,
  EmptyState,
  SearchInput,
  StatusBadge,
} from "@/components/shared"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { TaskDetailHeader, TaskSpecification } from "@/components/task-detail"
import {
  canAssign,
  isTaskFrozen,
  getPerson,
  progress,
  canViewProject,
  isDirector,
  statusLabels,
  type Project,
  type ProjectEvent,
  type OutputConflict,
  type Task,
  type TaskGroup,
  type Draft,
} from "@/demo/model"
import { AssignmentWorkspace } from "@/features/assignment/AssignmentWorkspace"
import { ProjectStructure } from "./ProjectStructure"
import { TaskEditor } from "./TaskEditor"
import { ProjectResponsibilities } from "./ProjectResponsibilities"
import { ProjectKanban } from "./ProjectKanban"
import { TaskReviewPanel } from "./TaskReviewPanel"
import type { ReviewResult, WorkState } from "@/demo/effort"
import { cn } from "@/lib/utils"

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
  onReview: (
    submissionId: string,
    result: ReviewResult,
    comment: string,
  ) => void
}) {
  const { projectId } = useParams()
  const project = projects.find((p) => p.id === projectId)
  const [search, setSearch] = useState("")
  const [filter, setFilter] = useState("all")
  const [group, setGroup] = useState("all")
  const [view, setView] = useState<"list" | "groups" | "kanban">("list")
  const [tab, setTab] = useState<"tasks" | "structure" | "reviews">("tasks")
  const [detailTab, setDetailTab] = useState<"details" | "results">("details")
  const [editor, setEditor] = useState<{
    task?: Task
    groupId?: string
  } | null>(null)
  const [info, setInfo] = useState(false)
  const [detail, setDetail] = useState<string | null>(null)
  const [selected, setSelected] = useState<string[]>([])
  const [draft, setDraft] = useState<Draft>({})
  const [discardTab, setDiscardTab] = useState<
    "tasks" | "structure" | "reviews" | null
  >(null)
  const [assignMode, setAssignMode] = useState<"manual" | "recommended" | null>(
    null,
  )
  const detailTrigger = useRef<HTMLButtonElement | null>(null)
  if (!project || !canViewProject(project, actorId))
    return (
      <EmptyState
        title="Không tìm thấy dự án"
        detail="Chọn dự án khác trong thanh điều hướng."
      />
    )
  const related = tasks.filter((t) => t.projectId === project.id)
  const groups = allGroups
    .filter((g) => g.projectId === project.id)
    .map((g) => g.name)
  const filtered = related.filter(
    (t) =>
      `${t.name} ${t.code}`
        .toLocaleLowerCase("vi")
        .includes(search.toLocaleLowerCase("vi")) &&
      (tab === "reviews"
        ? project.status === "IN_PROGRESS" && t.status === "PENDING_REVIEW"
        : filter === "all" || t.status === filter) &&
      (group === "all" || t.group === group),
  )
  const eligible = filtered.filter((t) => t.status === "PENDING_ASSIGNMENT")
  const currentTask = related.find((t) => t.id === detail)
  const allowed = canAssign(project, actorId)
  const canSelectAssignments = allowed && tab === "tasks"
  const canReadResults = project.ptcmId === actorId || project.pmId === actorId
  const waiting =
    project.status === "IN_PROGRESS"
      ? related.filter((t) => t.status === "PENDING_REVIEW").length
      : 0
  const percent = progress(related)
  const openTab = (next: "tasks" | "structure" | "reviews") => {
    setDraft({})
    setAssignMode(null)
    setTab(next)
    setView("list")
    if (next !== "tasks") {
      setSelected([])
      setDetail(null)
    }
  }
  const requestTab = (next: "tasks" | "structure" | "reviews") => {
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
  const hasAll =
    eligible.length > 0 && eligible.every((t) => selected.includes(t.id))
  const renderTable = (rows: Task[]) => (
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
                  disabled={!eligible.length}
                  onChange={() =>
                    setSelected((prev) =>
                      hasAll
                        ? prev.filter(
                            (id) => !eligible.some((t) => t.id === id),
                          )
                        : [...new Set([...prev, ...eligible.map((t) => t.id)])],
                    )
                  }
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
                      onChange={() => toggle(task.id)}
                    />
                  )}
                </td>
                <td className="py-4 pr-3">
                  <button
                    type="button"
                    className="flex items-start gap-2.5 text-left hover:underline"
                    aria-label={`Xem ${task.name}`}
                    onClick={(e) => {
                      detailTrigger.current = e.currentTarget
                      setDetail(task.id)
                      setDetailTab(
                        task.status === "PENDING_REVIEW" && canReadResults
                          ? "results"
                          : "details",
                      )
                    }}
                  >
                    <FileTextIcon className="mt-0.5 shrink-0 text-muted-foreground" />
                    <span>
                      <span className="block font-medium leading-5">
                        {task.name}
                      </span>
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
                    <span className="text-muted-foreground">
                      Chưa phân công
                    </span>
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
              <h1 className="text-xl font-semibold text-balance sm:text-2xl">
                {project.name}
              </h1>
              <StatusBadge status={project.status} />
            </div>
          </div>
        </div>
        <div className="flex gap-2">
          {allowed && !assignMode && tab !== "reviews" && (
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
        project={project}
        actorId={actorId}
        tasks={tasks}
        conflicts={conflicts}
        events={events}
        onChange={onProjectChange}
      />
      <div className="mb-4 flex items-center justify-between gap-4 border-b border-border">
        <div className="flex min-w-0 overflow-x-auto">
          <button
            type="button"
            onClick={() => requestTab("tasks")}
            className={cn(
              "tab-button workspace-tab",
              tab === "tasks" && "tab-selected",
            )}
            aria-pressed={tab === "tasks"}
          >
            <ListBulletIcon />
            Công việc <span className="count">{related.length}</span>
          </button>
          <button
            type="button"
            className={cn(
              "tab-button workspace-tab",
              tab === "structure" && "tab-selected",
            )}
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
              className={cn(
                "tab-button workspace-tab",
                tab === "reviews" && "tab-selected",
              )}
              aria-pressed={tab === "reviews"}
            >
              <CheckCircledIcon />
              Nghiệm thu <span className="count">{waiting}</span>
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
                : isDirector(actorId)
                  ? "Giám đốc · chỉ theo dõi"
                  : "Chỉ xem"}
        </span>
      </div>
      {tab === "structure" ? (
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
              <div className="mb-4 flex flex-wrap items-center gap-2">
                {tab !== "reviews" && (
                  <div className="mr-auto flex flex-wrap gap-1">
                    {[
                      ["all", "Tất cả"],
                      ["PENDING_ASSIGNMENT", "Chờ phân công"],
                      ["IN_PROGRESS", "Đang thực hiện"],
                    ].map(([value, label]) => (
                      <button
                        type="button"
                        key={value}
                        aria-pressed={filter === value}
                        className={cn(
                          "filter-chip",
                          filter === value && "filter-active",
                        )}
                        onClick={() => setFilter(value)}
                      >
                        {label}
                        <span className="ml-1 text-xs text-muted-foreground">
                          {
                            related.filter(
                              (t) => value === "all" || t.status === value,
                            ).length
                          }
                        </span>
                      </button>
                    ))}
                  </div>
                )}
                {tab === "reviews" && (
                  <p className="mr-auto text-sm text-muted-foreground">
                    Công việc đang chờ · {waiting}
                  </p>
                )}
                <div
                  className="flex rounded-md border border-border bg-white p-0.5"
                  aria-label="Cách xem"
                >
                  <button
                    type="button"
                    aria-label="Xem bảng"
                    aria-pressed={view === "list"}
                    className={cn(
                      "rounded p-2",
                      view === "list" && "bg-accent",
                    )}
                    onClick={() => setView("list")}
                  >
                    <ListBulletIcon />
                  </button>
                  <button
                    type="button"
                    aria-label="Xem theo nhóm"
                    aria-pressed={view === "groups"}
                    className={cn(
                      "rounded p-2",
                      view === "groups" && "bg-accent",
                    )}
                    onClick={() => setView("groups")}
                  >
                    <RowsIcon />
                  </button>
                  <button
                    type="button"
                    aria-label="Xem Kanban"
                    aria-pressed={view === "kanban"}
                    className={cn(
                      "rounded p-2",
                      view === "kanban" && "bg-accent",
                    )}
                    onClick={() => setView("kanban")}
                  >
                    <ViewVerticalIcon />
                  </button>
                </div>
              </div>
              <div className="mb-4 grid gap-3 sm:grid-cols-[minmax(0,1fr)_auto]">
                <SearchInput
                  value={search}
                  onChange={setSearch}
                  placeholder="Tìm công việc…"
                />
                <label className="flex items-center gap-2 rounded-md border border-border bg-white pl-3 text-sm">
                  <MixerHorizontalIcon />
                  <span className="field-label">Nhóm</span>
                  <select
                    aria-label="Nhóm công việc"
                    className="field workspace-field min-w-0 border-0"
                    value={group}
                    onChange={(e) => setGroup(e.target.value)}
                  >
                    <option value="all">Tất cả nhóm</option>
                    {groups.map((g) => (
                      <option key={g}>{g}</option>
                    ))}
                  </select>
                </label>
              </div>
              {tab === "tasks" && (
                <label className="mb-4 flex w-fit max-w-full flex-wrap items-center gap-2">
                  <span className="field-label">Trạng thái</span>
                  <select
                    className="field workspace-field"
                    aria-label="Trạng thái công việc dự án"
                    value={filter}
                    onChange={(e) => setFilter(e.target.value)}
                  >
                    <option value="all">Tất cả trạng thái</option>
                    {Object.entries(statusLabels).map(([value, label]) => (
                      <option key={value} value={value}>
                        {label}
                      </option>
                    ))}
                  </select>
                </label>
              )}
              {tab === "reviews" && (
                <p className="mb-4 text-xs leading-5 text-muted-foreground">
                  Chọn công việc để xem lần nộp và tiêu chí.{" "}
                  {project.ptcmId === actorId
                    ? "Bạn xử lý theo trách nhiệm PTCM."
                    : "Bạn chỉ theo dõi; quyết định thuộc PTCM."}
                </p>
              )}
              {view === "kanban" && (
                <>
                  {filtered.some((t) =>
                    ["CLOSED", "CANCELLED"].includes(t.status),
                  ) && (
                    <p className="mb-3 text-xs text-muted-foreground">
                      {
                        filtered.filter((t) =>
                          ["CLOSED", "CANCELLED"].includes(t.status),
                        ).length
                      }{" "}
                      công việc đã đóng/hủy không nằm trên Kanban.{" "}
                      <button
                        type="button"
                        className="underline"
                        onClick={() => setView("list")}
                      >
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
                        task.status === "PENDING_REVIEW" && canReadResults
                          ? "results"
                          : "details",
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
                    renderTable(filtered)
                  ) : (
                    groups
                      .filter((g) => filtered.some((t) => t.group === g))
                      .map((g) => (
                        <details
                          key={g}
                          open
                          className="border-b border-border last:border-0"
                        >
                          <summary className="field-label flex cursor-pointer list-none items-center gap-2 bg-background/70 px-4 py-3">
                            <ChevronDownIcon />
                            {g}
                            <span className="count">
                              {filtered.filter((t) => t.group === g).length}
                            </span>
                          </summary>
                          {renderTable(filtered.filter((t) => t.group === g))}
                        </details>
                      ))
                  )}
                </div>
              )}
              <p className="mt-3 text-xs text-muted-foreground">
                {filtered.length} / {related.length} công việc · Nhấn tên để xem
                chi tiết mà không rời danh sách.
              </p>
            </section>
            {(info || currentTask) && (
              <aside
                className={cn(
                  "min-w-0 rounded-lg border border-border bg-white p-5 xl:sticky xl:top-6",
                  currentTask &&
                    "xl:max-h-[calc(100dvh-8rem)] xl:overflow-y-auto",
                )}
                aria-label={
                  currentTask ? "Chi tiết công việc" : "Thông tin dự án"
                }
                onKeyDown={(e) => {
                  if (
                    e.key === "Escape" &&
                    currentTask &&
                    !(e.target as HTMLElement).closest('[role="dialog"]')
                  )
                    closeDetail()
                }}
              >
                {currentTask ? (
                  <>
                    <TaskDetailHeader
                      task={currentTask}
                      onClose={closeDetail}
                    />
                    {canReadResults && (
                      <div
                        className="mt-4 flex gap-1"
                        role="group"
                        aria-label="Nội dung công việc dự án"
                      >
                        <Button
                          size="sm"
                          variant={
                            detailTab === "details" ? "default" : "ghost"
                          }
                          aria-pressed={detailTab === "details"}
                          className="workspace-tab"
                          onClick={() => setDetailTab("details")}
                        >
                          Chi tiết
                        </Button>
                        <Button
                          size="sm"
                          variant={
                            detailTab === "results" ? "default" : "ghost"
                          }
                          aria-pressed={detailTab === "results"}
                          className="workspace-tab"
                          onClick={() => setDetailTab("results")}
                        >
                          Kết quả
                        </Button>
                      </div>
                    )}
                    {detailTab === "results" && canReadResults ? (
                      <>
                        <details
                          className="mt-4 rounded-md border border-border p-3 text-sm"
                          open
                        >
                          <summary className="section-title cursor-pointer">
                            Tiêu chí nghiệm thu
                          </summary>
                          <p className="mt-2 whitespace-pre-wrap break-words leading-6 text-muted-foreground">
                            {currentTask.criteria}
                          </p>
                        </details>
                        <TaskReviewPanel
                          key={currentTask.id}
                          project={project}
                          task={currentTask}
                          work={work}
                          actorId={actorId}
                          conflicts={conflicts}
                          onReview={onReview}
                        />
                      </>
                    ) : (
                      <>
                        {allowed &&
                          (isTaskFrozen(currentTask) ? (
                            <p className="mt-4 flex items-start gap-2 rounded-md bg-accent p-3 text-xs leading-5 text-muted-foreground">
                              <LockClosedIcon className="mt-0.5" />
                              Đặc tả đã khóa sau lần phân công đầu. Biểu mẫu sửa
                              làm rõ/gia hạn chưa dựng trong đợt này.
                            </p>
                          ) : (
                            <Button
                              variant="outline"
                              size="sm"
                              className="mt-4"
                              onClick={() => setEditor({ task: currentTask })}
                            >
                              <Pencil1Icon />
                              Chỉnh sửa công việc
                            </Button>
                          ))}
                        <dl className="mt-4 space-y-2 border-t border-border pt-4 text-sm">
                          <div className="flex justify-between gap-2">
                            <dt className="field-label">
                              Người thực hiện
                            </dt>
                            <dd className="text-right">
                              {getPerson(currentTask.assigneeId)?.name ??
                                "Chưa phân công"}
                            </dd>
                          </div>
                          <div className="flex justify-between gap-2">
                            <dt className="field-label">Nhóm</dt>
                            <dd className="text-right">{currentTask.group}</dd>
                          </div>
                          <div className="flex justify-between">
                            <dt className="field-label">Ưu tiên</dt>
                            <dd>{currentTask.priority}</dd>
                          </div>
                          <div className="flex justify-between gap-2">
                            <dt className="field-label">Kế hoạch</dt>
                            <dd className="text-right">
                              {currentTask.start && currentTask.end
                                ? `${currentTask.start.split("-").reverse().join("/")} – ${currentTask.end.split("-").reverse().join("/")}`
                                : "Chưa đặt ngày"}
                            </dd>
                          </div>
                        </dl>
                        <div className="mt-5 border-t border-border pt-4 text-sm">
                          Ước tính{" "}
                          <span className="float-right font-medium">
                            {currentTask.estimate} giờ
                          </span>
                        </div>
                        <TaskSpecification
                          key={currentTask.id}
                          task={currentTask}
                          showWeights
                        />
                      </>
                    )}
                  </>
                ) : (
                  <>
                    <h2 className="section-title">Về dự án</h2>
                    <p className="mt-3 text-sm leading-6 text-muted-foreground">
                      {project.description}
                    </p>
                    <dl className="mt-5 space-y-4 text-sm">
                      <div className="flex justify-between gap-3">
                        <dt className="field-label">Ưu tiên</dt>
                        <dd>{project.priority}</dd>
                      </div>
                      <div className="flex justify-between gap-3">
                        <dt className="field-label">Ngày bắt đầu</dt>
                        <dd>
                          {new Date(
                            `${project.start}T00:00:00`,
                          ).toLocaleDateString("vi-VN")}
                        </dd>
                      </div>
                      <div className="flex justify-between gap-3">
                        <dt className="field-label">Hạn dự kiến</dt>
                        <dd>
                          {new Date(
                            `${project.end}T00:00:00`,
                          ).toLocaleDateString("vi-VN")}
                        </dd>
                      </div>
                    </dl>
                    <div className="mt-6 border-t border-border pt-5">
                      <div className="flex justify-between text-sm">
                        <span>Đã nghiệm thu</span>
                        <strong>
                          {percent === null ? "—" : `${percent}%`}
                        </strong>
                      </div>
                      <div className="mt-3 h-1.5 overflow-hidden rounded-full bg-accent">
                        <div
                          className="h-full rounded-full bg-success"
                          style={{ width: `${percent ?? 0}%` }}
                        />
                      </div>
                      <p className="mt-3 text-xs text-muted-foreground">
                        {related.filter((t) => t.status === "DONE").length} /{" "}
                        {related.filter((t) => t.status !== "CANCELLED").length}{" "}
                        công việc hiệu lực
                      </p>
                    </div>
                    <div className="mt-6 border-t border-border pt-5 text-xs leading-5 text-muted-foreground">
                      Không cần lập danh sách thành viên trước. PTCM lựa chọn
                      người phù hợp trong toàn công ty.
                    </div>
                  </>
                )}
              </aside>
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
