import {
  ListBulletIcon,
  MixerHorizontalIcon,
  RowsIcon,
  ViewVerticalIcon,
} from "@radix-ui/react-icons"
import { SearchInput } from "@/components/shared"
import { statusLabels } from "@/models/workspace-labels"
import type { Task } from "@/models/workspace"
import { cn } from "@/lib/utils"

export function ProjectTaskToolbar({
  tab,
  related,
  waiting,
  view,
  search,
  filter,
  group,
  groups,
  isPtcm,
  onViewChange,
  onSearchChange,
  onFilterChange,
  onGroupChange,
}: {
  tab: "tasks" | "structure" | "reviews" | "corrections"
  related: Task[]
  waiting: number
  view: "list" | "groups" | "kanban"
  search: string
  filter: string
  group: string
  groups: string[]
  isPtcm: boolean
  onViewChange: (view: "list" | "groups" | "kanban") => void
  onSearchChange: (value: string) => void
  onFilterChange: (value: string) => void
  onGroupChange: (value: string) => void
}) {
  return (
    <>
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
                className={cn("filter-chip", filter === value && "filter-active")}
                onClick={() => onFilterChange(value)}
              >
                {label}
                <span className="ml-1 text-xs text-muted-foreground">
                  {related.filter((t) => value === "all" || t.status === value).length}
                </span>
              </button>
            ))}
          </div>
        )}
        {tab === "reviews" && (
          <p className="mr-auto text-sm text-muted-foreground">Công việc đang chờ · {waiting}</p>
        )}
        <div className="flex rounded-md border border-border bg-white p-0.5" aria-label="Cách xem">
          <button
            type="button"
            aria-label="Xem bảng"
            aria-pressed={view === "list"}
            className={cn("rounded p-2", view === "list" && "bg-accent")}
            onClick={() => onViewChange("list")}
          >
            <ListBulletIcon />
          </button>
          <button
            type="button"
            aria-label="Xem theo nhóm"
            aria-pressed={view === "groups"}
            className={cn("rounded p-2", view === "groups" && "bg-accent")}
            onClick={() => onViewChange("groups")}
          >
            <RowsIcon />
          </button>
          <button
            type="button"
            aria-label="Xem Kanban"
            aria-pressed={view === "kanban"}
            className={cn("rounded p-2", view === "kanban" && "bg-accent")}
            onClick={() => onViewChange("kanban")}
          >
            <ViewVerticalIcon />
          </button>
        </div>
      </div>
      <div className="mb-4 grid gap-3 sm:grid-cols-[minmax(0,1fr)_auto]">
        <SearchInput value={search} onChange={onSearchChange} placeholder="Tìm công việc…" />
        <label className="flex items-center gap-2 rounded-md border border-border bg-white pl-3 text-sm">
          <MixerHorizontalIcon />
          <span className="field-label">Nhóm</span>
          <select
            aria-label="Nhóm công việc"
            className="field workspace-field min-w-0 border-0"
            value={group}
            onChange={(e) => onGroupChange(e.target.value)}
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
            onChange={(e) => onFilterChange(e.target.value)}
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
          {isPtcm ? "Bạn xử lý theo trách nhiệm PTCM." : "Bạn chỉ theo dõi; quyết định thuộc PTCM."}
        </p>
      )}
    </>
  )
}
