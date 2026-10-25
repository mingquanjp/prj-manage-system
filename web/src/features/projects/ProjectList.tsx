import { useState } from "react"
import { Link } from "react-router"
import { ArrowTopRightIcon, CubeIcon, PlusIcon } from "@radix-ui/react-icons"
import { Avatar, EmptyState, SearchInput, StatusBadge } from "@/components/shared"
import { getPerson, isDirector } from "@/demo/identity"
import { progress } from "@/demo/projects"
import type { Project, Task } from "@/models/workspace"
import { Button } from "@/components/ui/button"

export function ProjectList({
  tasks,
  projects,
  actorId,
  onCreate,
}: {
  tasks: Task[]
  projects: Project[]
  actorId: string
  onCreate: () => void
}) {
  const [search, setSearch] = useState("")
  const [filter, setFilter] = useState("all")
  const filtered = projects.filter(
    (p) =>
      `${p.name} ${p.code}`.toLocaleLowerCase("vi").includes(search.toLocaleLowerCase("vi")) &&
      (filter === "all" || p.status === filter),
  )
  return (
    <>
      <div className="mb-8 flex flex-wrap items-start justify-between gap-4">
        <div>
          <p className="eyebrow">KHÔNG GIAN LÀM VIỆC</p>
          <h1 className="mt-2 text-2xl font-semibold text-balance">
            {isDirector(actorId) ? "Dự án" : "Dự án của tôi"}
          </h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Theo dõi công việc và điều phối chuyên môn trong từng dự án.
          </p>
        </div>
        {isDirector(actorId) && (
          <Button onClick={onCreate}>
            <PlusIcon />
            Tạo dự án
          </Button>
        )}
      </div>
      <div className="mb-5 flex flex-wrap items-center justify-between gap-4">
        <div className="flex flex-wrap gap-1" aria-label="Lọc trạng thái dự án">
          {[
            ["all", "Tất cả"],
            ["IN_PROGRESS", "Đang thực hiện"],
            ["PLANNING", "Lập kế hoạch"],
            ["COMPLETED", "Đã hoàn thành"],
            ["CLOSED", "Đã đóng"],
            ["CANCELLED", "Đã hủy"],
          ].map(([value, label]) => (
            <button
              type="button"
              key={value}
              aria-pressed={filter === value}
              onClick={() => setFilter(value)}
              className={`filter-chip ${filter === value ? "filter-active" : ""}`}
            >
              {label}
            </button>
          ))}
        </div>
        <div className="w-full sm:w-72">
          <SearchInput value={search} onChange={setSearch} placeholder="Tìm tên hoặc mã dự án…" />
        </div>
      </div>
      <div className="overflow-x-auto rounded-lg border border-border bg-white">
        <table className="w-full min-w-[780px] text-left text-sm">
          <thead>
            <tr className="table-heading">
              <th className="py-3 pl-5">Tên dự án</th>
              <th>Trạng thái</th>
              <th>Tiến độ nghiệm thu</th>
              <th>PM</th>
              <th className="pr-5">Hạn dự kiến</th>
            </tr>
          </thead>
          <tbody>
            {filtered.map((project) => {
              const related = tasks.filter((t) => t.projectId === project.id)
              const percent = progress(related)
              const pending = related.filter((t) => t.status === "PENDING_ASSIGNMENT").length
              const pm = getPerson(project.pmId)!
              return (
                <tr key={project.id} className="border-t border-border hover:bg-background/70">
                  <td className="py-6 pl-5 pr-5">
                    <Link to={`/projects/${project.id}`} className="group flex items-center gap-3">
                      <span className="grid size-10 shrink-0 place-items-center rounded-lg border border-border bg-background">
                        <CubeIcon className="size-5" />
                      </span>
                      <span>
                        <span className="flex items-center gap-2 font-medium group-hover:underline">
                          {project.name}
                          <ArrowTopRightIcon className="text-muted-foreground" />
                        </span>
                        <span className="mt-1 block text-xs text-muted-foreground">
                          {project.code} · {related.length} công việc · {pending} chờ phân công
                        </span>
                      </span>
                    </Link>
                  </td>
                  <td className="pr-5">
                    <StatusBadge status={project.status} />
                  </td>
                  <td className="pr-8">
                    <div className="flex items-center gap-3">
                      <div className="h-1.5 w-24 overflow-hidden rounded-full bg-accent">
                        <div
                          style={{ width: `${percent ?? 0}%` }}
                          className="h-full rounded-full bg-success"
                        />
                      </div>
                      <span className="tabular-nums">{percent === null ? "—" : `${percent}%`}</span>
                    </div>
                  </td>
                  <td className="pr-5">
                    <span className="inline-flex items-center gap-2 whitespace-nowrap">
                      <Avatar initials={pm.initials} small />
                      {pm.name}
                    </span>
                  </td>
                  <td className="whitespace-nowrap pr-5 tabular-nums">
                    {new Date(`${project.end}T00:00:00`).toLocaleDateString("vi-VN")}
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
        {!filtered.length && (
          <EmptyState
            title="Không tìm thấy dự án"
            detail={
              projects.length
                ? "Thử tên khác hoặc thay đổi bộ lọc."
                : "Chưa có dự án thuộc phạm vi trách nhiệm của bạn trong phiên mô phỏng."
            }
          />
        )}
      </div>
      <p className="mt-4 text-xs text-muted-foreground">
        {filtered.length} dự án · Tiến độ dựa trên công việc đã nghiệm thu, không phải số công việc
        đã giao.
      </p>
    </>
  )
}
