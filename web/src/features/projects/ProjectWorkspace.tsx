import { useRef, useState } from "react"
import { Link, useParams } from "react-router"
import {
  ArrowLeftIcon,
  CheckCircledIcon,
  ChevronDownIcon,
  Cross2Icon,
  CubeIcon,
  FileTextIcon,
  InfoCircledIcon,
  ListBulletIcon,
  MixerHorizontalIcon,
  RowsIcon,
} from "@radix-ui/react-icons"
import {
  Avatar,
  EmptyState,
  SearchInput,
  StatusBadge,
} from "@/components/shared"
import { Button } from "@/components/ui/button"
import {
  canAssign,
  getPerson,
  levels,
  progress,
  projects,
  type Task,
} from "@/demo/model"
import { AssignmentWorkspace } from "@/features/assignment/AssignmentWorkspace"
import { cn } from "@/lib/utils"

export function ProjectWorkspace({
  actorId,
  tasks,
  onTasksChange,
  onNotice,
}: {
  actorId: string
  tasks: Task[]
  onTasksChange: (tasks: Task[]) => void
  onNotice: (notice: string) => void
}) {
  const { projectId } = useParams()
  const project = projects.find((p) => p.id === projectId)
  const [search, setSearch] = useState("")
  const [filter, setFilter] = useState("all")
  const [group, setGroup] = useState("all")
  const [view, setView] = useState<"list" | "groups">("list")
  const [info, setInfo] = useState(true)
  const [detail, setDetail] = useState<string | null>(null)
  const [selected, setSelected] = useState<string[]>([])
  const [assignMode, setAssignMode] = useState<"manual" | "recommended" | null>(
    null,
  )
  const detailTrigger = useRef<HTMLButtonElement | null>(null)
  if (!project)
    return (
      <EmptyState
        title="Không tìm thấy dự án"
        detail="Chọn dự án khác trong thanh điều hướng."
      />
    )
  const related = tasks.filter((t) => t.projectId === project.id)
  const groups = [...new Set(related.map((t) => t.group))]
  const filtered = related.filter(
    (t) =>
      `${t.name} ${t.code}`
        .toLocaleLowerCase("vi")
        .includes(search.toLocaleLowerCase("vi")) &&
      (filter === "all" || t.status === filter) &&
      (group === "all" || t.group === group),
  )
  const eligible = filtered.filter((t) => t.status === "PENDING_ASSIGNMENT")
  const currentTask = related.find((t) => t.id === detail)
  const allowed = canAssign(project, actorId)
  const percent = progress(related)
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
          <tr className="table-heading">
            <th className="w-10 py-3 pl-4">
              {allowed && (
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
                  {allowed && (
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
        className="mb-5 inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeftIcon />
        Tất cả dự án
      </Link>
      <div className="mb-5 flex flex-wrap items-start justify-between gap-4">
        <div className="flex items-start gap-4">
          <span className="grid size-12 shrink-0 place-items-center rounded-xl border border-border bg-white">
            <CubeIcon className="size-7" />
          </span>
          <div>
            <p className="eyebrow">{project.code}</p>
            <h1 className="mt-1 text-2xl font-semibold">{project.name}</h1>
            <div className="mt-3 flex flex-wrap items-center gap-3 text-sm text-muted-foreground">
              <StatusBadge status={project.status} />
              <span>PM: {getPerson(project.pmId)?.name}</span>
              <span>PTCM: {getPerson(project.ptcmId)?.name}</span>
            </div>
          </div>
        </div>
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
      </div>
      <div className="mb-6 flex items-center justify-between gap-4 border-b border-border">
        <div className="flex">
          <button
            type="button"
            onClick={() => {
              setAssignMode(null)
              setView("list")
            }}
            className="tab-button tab-selected"
          >
            <ListBulletIcon />
            Công việc <span className="count">{related.length}</span>
          </button>
          <button
            type="button"
            disabled
            title="Nghiệm thu chưa dựng trong lát cắt thiết kế này"
            className="tab-button opacity-45"
          >
            <CheckCircledIcon />
            Nghiệm thu{" "}
            <span className="count">
              {related.filter((t) => t.status === "PENDING_REVIEW").length}
            </span>
          </button>
        </div>
        <span className="hidden text-xs text-muted-foreground md:block">
          Bạn đang có trách nhiệm{" "}
          {allowed ? "PTCM" : "PM · chỉ theo dõi chuyên môn"}
        </span>
      </div>
      {assignMode && allowed ? (
        <AssignmentWorkspace
          project={project}
          actorId={actorId}
          tasks={tasks}
          selectedIds={selected}
          initialMode={assignMode}
          onBack={() => setAssignMode(null)}
          onCommit={(next) => {
            onTasksChange(next)
            setAssignMode(null)
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
              (info || currentTask) && "xl:grid-cols-[minmax(0,1fr)_280px]",
            )}
          >
            <section className="min-w-0">
              <div className="mb-4 flex flex-wrap items-center gap-2">
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
                </div>
              </div>
              <div className="mb-4 grid gap-3 sm:grid-cols-[minmax(0,1fr)_auto]">
                <SearchInput
                  value={search}
                  onChange={setSearch}
                  placeholder="Tìm công việc…"
                />
                <label className="flex items-center gap-2 rounded-md border border-border bg-white pl-3 text-sm text-muted-foreground">
                  <MixerHorizontalIcon />
                  <select
                    aria-label="Nhóm công việc"
                    className="field border-0"
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
              <div className="overflow-hidden rounded-lg border border-border bg-white">
                {!filtered.length ? (
                  <EmptyState
                    title="Không có công việc phù hợp"
                    detail="Thử thay đổi từ khóa hoặc bộ lọc."
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
                        <summary className="flex cursor-pointer list-none items-center gap-2 bg-background/70 px-4 py-3 text-sm font-medium">
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
              <p className="mt-3 text-xs text-muted-foreground">
                {filtered.length} / {related.length} công việc · Nhấn tên để xem
                chi tiết mà không rời danh sách.
              </p>
            </section>
            {(info || currentTask) && (
              <aside
                className="rounded-lg border border-border bg-white p-5 xl:sticky xl:top-6"
                aria-label={
                  currentTask ? "Chi tiết công việc" : "Thông tin dự án"
                }
                onKeyDown={(e) => {
                  if (e.key === "Escape" && currentTask) closeDetail()
                }}
              >
                {currentTask ? (
                  <>
                    <div className="mb-4 flex items-center justify-between">
                      <p className="eyebrow">{currentTask.code}</p>
                      <Button
                        variant="ghost"
                        size="icon"
                        aria-label="Đóng chi tiết công việc"
                        onClick={closeDetail}
                      >
                        <Cross2Icon />
                      </Button>
                    </div>
                    <h2 className="text-lg font-semibold leading-7">
                      {currentTask.name}
                    </h2>
                    <div className="mt-3">
                      <StatusBadge status={currentTask.status} />
                    </div>
                    <p className="mt-5 text-sm leading-6 text-muted-foreground">
                      {currentTask.description}
                    </p>
                    <h3 className="mt-5 text-sm font-medium">
                      Tiêu chí nghiệm thu
                    </h3>
                    <p className="mt-2 text-sm leading-6 text-muted-foreground">
                      {currentTask.criteria}
                    </p>
                    <h3 className="mt-5 text-sm font-medium">
                      Yêu cầu kỹ năng
                    </h3>
                    <div className="mt-3 space-y-3">
                      {currentTask.requirements.map((r) => (
                        <div key={r.skill} className="text-sm">
                          <div className="flex justify-between gap-2">
                            <span>{r.skill}</span>
                            <span className="text-muted-foreground">
                              {levels[r.level]}
                            </span>
                          </div>
                          <p className="mt-1 text-xs text-muted-foreground">
                            {r.mandatory ? "Bắt buộc" : "Bổ trợ"}
                          </p>
                        </div>
                      ))}
                    </div>
                    <div className="mt-5 border-t border-border pt-4 text-sm">
                      Ước tính{" "}
                      <span className="float-right font-medium">
                        {currentTask.estimate} giờ
                      </span>
                    </div>
                  </>
                ) : (
                  <>
                    <h2 className="font-semibold">Về dự án</h2>
                    <p className="mt-3 text-sm leading-6 text-muted-foreground">
                      {project.description}
                    </p>
                    <dl className="mt-5 space-y-4 text-sm">
                      <div className="flex justify-between gap-3">
                        <dt className="text-muted-foreground">Ưu tiên</dt>
                        <dd>{project.priority}</dd>
                      </div>
                      <div className="flex justify-between gap-3">
                        <dt className="text-muted-foreground">Ngày bắt đầu</dt>
                        <dd>
                          {new Date(
                            `${project.start}T00:00:00`,
                          ).toLocaleDateString("vi-VN")}
                        </dd>
                      </div>
                      <div className="flex justify-between gap-3">
                        <dt className="text-muted-foreground">Hạn dự kiến</dt>
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
    </>
  )
}
