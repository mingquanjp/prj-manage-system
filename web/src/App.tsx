import { useState } from "react"
import { Link, NavLink, Route, Routes, useLocation } from "react-router"
import {
  ArchiveIcon,
  ChevronDownIcon,
  ChevronRightIcon,
  CubeIcon,
  DashboardIcon,
  MixerHorizontalIcon,
  ResetIcon,
} from "@radix-ui/react-icons"
import { Avatar } from "@/components/shared"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { getPerson, initialTasks, projects, type Task } from "@/demo/model"
import { ProjectList } from "@/features/projects/ProjectList"
import { ProjectWorkspace } from "@/features/projects/ProjectWorkspace"
import { cn } from "@/lib/utils"

export default function App() {
  const [actorId, setActorId] = useState("khoa")
  const [tasks, setTasks] = useState<Task[]>(() =>
    structuredClone(initialTasks),
  )
  const [resetOpen, setResetOpen] = useState(false)
  const [notice, setNotice] = useState("")
  const location = useLocation()
  const currentProject = projects.find((p) =>
    location.pathname.startsWith(`/projects/${p.id}`),
  )
  const person = getPerson(actorId)!

  return (
    <div className="min-h-dvh bg-background text-foreground lg:grid lg:grid-cols-[224px_minmax(0,1fr)]">
      <aside className="hidden flex-col border-r border-border bg-sidebar lg:sticky lg:top-0 lg:flex lg:h-dvh">
        <Link to="/" className="flex h-[76px] items-center gap-3 px-5">
          <span className="grid size-9 place-items-center rounded-lg bg-primary text-white">
            <ArchiveIcon className="size-5" />
          </span>
          <div>
            <strong className="block text-base">TaskAssign</strong>
            <span className="text-xs text-muted-foreground">
              Không gian làm việc
            </span>
          </div>
        </Link>
        <div className="px-3 pt-5">
          <div className="px-3 pb-3 text-xs font-medium tracking-wider text-muted-foreground uppercase">
            Không gian dự án
          </div>
          <NavLink
            to="/"
            end
            className={({ isActive }) =>
              cn(
                "flex h-10 items-center gap-3 rounded-md px-3 text-sm font-medium",
                isActive || currentProject ? "bg-accent" : "hover:bg-accent",
              )
            }
          >
            <DashboardIcon />
            Dự án của tôi
            <span className="ml-auto text-muted-foreground">
              {projects.length}
            </span>
          </NavLink>
          <nav
            aria-label="Dự án"
            className="mt-3 space-y-1 border-l border-border ml-5 pl-3"
          >
            {projects.map((project) => (
              <NavLink
                key={project.id}
                to={`/projects/${project.id}`}
                className={({ isActive }) =>
                  cn(
                    "flex items-center gap-2 rounded-md px-2 py-2.5 text-sm",
                    isActive
                      ? "bg-accent text-foreground"
                      : "text-muted-foreground hover:bg-accent",
                  )
                }
              >
                <CubeIcon className="shrink-0" />
                <span className="truncate">{project.name}</span>
              </NavLink>
            ))}
          </nav>
        </div>
        <div className="mx-5 mt-8 border-t border-border pt-5 text-sm leading-6 text-muted-foreground">
          <p className="font-medium text-foreground">Bản duyệt thiết kế</p>
          <p className="mt-1">
            Dự án, công việc và phân công. Các khu vực khác sẽ bổ sung sau.
          </p>
        </div>
        <div className="mt-auto px-4 pb-5 pt-8">
          <details className="mb-4 rounded-md border border-border bg-white p-3">
            <summary className="flex cursor-pointer list-none items-center justify-between gap-2 text-xs font-medium">
              <span className="flex items-center gap-2">
                <MixerHorizontalIcon />
                Điều khiển prototype
              </span>
              <ChevronDownIcon />
            </summary>
            <label
              className="mt-3 block text-xs text-muted-foreground"
              htmlFor="identity"
            >
              Danh tính mô phỏng
            </label>
            <select
              id="identity"
              className="field mt-1 w-full text-sm"
              value={actorId}
              onChange={(e) => {
                setActorId(e.target.value)
                setNotice("")
              }}
            >
              <option value="khoa">Ngô Minh Khoa · PTCM</option>
              <option value="thanh">Phan Thị Thanh · PM</option>
            </select>
            <p className="mt-2 text-xs leading-5 text-muted-foreground">
              Chỉ kiểm thử giao diện, không phải tính năng tự cấp quyền.
            </p>
            <Button
              size="sm"
              variant="ghost"
              className="mt-2 w-full"
              onClick={() => setResetOpen(true)}
            >
              <ResetIcon />
              Khôi phục dữ liệu mẫu
            </Button>
          </details>
          <div className="flex items-center gap-3">
            <Avatar initials={person.initials} />
            <div className="min-w-0">
              <p className="truncate text-sm font-medium">{person.name}</p>
              <p className="text-xs text-muted-foreground">
                {actorId === "khoa" ? "Phụ trách chuyên môn" : "Quản lý dự án"}
              </p>
            </div>
          </div>
        </div>
      </aside>
      <div className="min-w-0">
        <header className="flex h-14 items-center justify-between gap-3 border-b border-border bg-sidebar px-6 text-sm lg:px-8">
          <nav
            aria-label="Đường dẫn"
            className="flex min-w-0 items-center gap-2 text-muted-foreground"
          >
            <Link to="/" className="hover:text-foreground">
              Dự án
            </Link>
            {currentProject && (
              <>
                <ChevronRightIcon className="shrink-0" />
                <span className="truncate text-foreground">
                  {currentProject.name}
                </span>
              </>
            )}
          </nav>
          <span className="shrink-0 rounded border border-border px-2 py-1 text-xs text-muted-foreground">
            Dữ liệu mô phỏng
          </span>
        </header>
        <div className="flex items-center justify-between gap-3 border-b border-border bg-sidebar px-4 py-3 lg:hidden">
          <Link to="/" className="font-semibold">
            TaskAssign
          </Link>
          <label className="flex items-center gap-2 text-xs text-muted-foreground">
            Mô phỏng
            <select
              aria-label="Danh tính mô phỏng trên màn hình nhỏ"
              className="field max-w-48"
              value={actorId}
              onChange={(e) => setActorId(e.target.value)}
            >
              <option value="khoa">Ngô Minh Khoa · PTCM</option>
              <option value="thanh">Phan Thị Thanh · PM</option>
            </select>
          </label>
        </div>
        <main className="mx-auto max-w-[1600px] px-4 py-8 md:px-8">
          {notice && (
            <div
              role="status"
              className="mb-5 flex items-center justify-between gap-4 rounded-md border border-success/20 bg-success-soft px-4 py-3 text-sm text-success"
            >
              {notice}
              <Button size="sm" variant="ghost" onClick={() => setNotice("")}>
                Ẩn
              </Button>
            </div>
          )}
          <Routes>
            <Route path="/" element={<ProjectList tasks={tasks} />} />
            <Route
              path="/projects/:projectId"
              element={
                <ProjectWorkspace
                  key={`${currentProject?.id}-${actorId}`}
                  actorId={actorId}
                  tasks={tasks}
                  onTasksChange={setTasks}
                  onNotice={setNotice}
                />
              }
            />
            <Route
              path="*"
              element={
                <div>
                  <h1 className="text-2xl font-semibold">
                    Không tìm thấy trang
                  </h1>
                  <Button asChild variant="outline" className="mt-4">
                    <Link to="/">Về danh sách dự án</Link>
                  </Button>
                </div>
              }
            />
          </Routes>
        </main>
      </div>
      <Dialog
        open={resetOpen}
        onOpenChange={setResetOpen}
        title="Khôi phục dữ liệu mẫu?"
        description="Các phân công mô phỏng trong phiên này sẽ được xóa. Không có dữ liệu máy chủ bị thay đổi."
      >
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={() => setResetOpen(false)}>
            Giữ nguyên
          </Button>
          <Button
            onClick={() => {
              setTasks(structuredClone(initialTasks))
              setResetOpen(false)
              setNotice("Đã khôi phục dữ liệu mẫu.")
            }}
          >
            Khôi phục
          </Button>
        </div>
      </Dialog>
    </div>
  )
}
