import { useState } from "react"
import {
  ChevronDownIcon,
  FileTextIcon,
  LockClosedIcon,
  Pencil1Icon,
  PlusIcon,
  LayersIcon,
} from "@radix-ui/react-icons"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { StatusBadge } from "@/components/shared"
import {
  canAssign,
  isTaskFrozen,
  saveGroup,
  type Project,
  type Task,
  type TaskGroup,
} from "@/demo/model"

export function ProjectStructure({
  project,
  actorId,
  groups,
  tasks,
  onChange,
  onNotice,
  onCreateTask,
  onEditTask,
  onViewTask,
}: {
  project: Project
  actorId: string
  groups: TaskGroup[]
  tasks: Task[]
  onChange: (groups: TaskGroup[], tasks: Task[]) => void
  onNotice: (message: string) => void
  onCreateTask: (groupId: string) => void
  onEditTask: (task: Task) => void
  onViewTask: (task: Task) => void
}) {
  const [editing, setEditing] = useState<TaskGroup | "new" | null>(null)
  const [name, setName] = useState("")
  const [description, setDescription] = useState("")
  const [error, setError] = useState("")
  const allowed = canAssign(project, actorId)
  const relatedGroups = groups.filter((g) => g.projectId === project.id)
  const openGroup = (group: TaskGroup | "new") => {
    setEditing(group)
    setName(group === "new" ? "" : group.name)
    setDescription(group === "new" ? "" : group.description)
    setError("")
  }
  return (
    <section className="min-w-0" aria-label="Cấu trúc dự án">
      <div className="mb-5 flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold">Nhóm công việc</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            Tổ chức phạm vi theo module hoặc gói công việc. Nhóm có thể chưa có
            công việc.
          </p>
        </div>
        {allowed && (
          <Button variant="outline" onClick={() => openGroup("new")}>
            <PlusIcon />
            Tạo nhóm
          </Button>
        )}
      </div>
      <div className="space-y-4">
        {relatedGroups.map((group) => {
          const rows = tasks.filter(
            (t) => t.projectId === project.id && t.groupId === group.id,
          )
          return (
            <details
              key={group.id}
              open
              className="overflow-hidden rounded-lg border border-border bg-white"
            >
              <summary className="flex cursor-pointer list-none items-center gap-3 bg-background/50 px-4 py-4">
                <ChevronDownIcon />
                <LayersIcon className="text-muted-foreground" />
                <span className="min-w-0 break-words font-medium">
                  {group.name}
                </span>
                <span className="count">{rows.length}</span>
              </summary>
              <div className="flex flex-wrap items-center justify-between gap-2 border-t border-border px-4 py-3">
                <p className="min-w-0 text-sm text-muted-foreground">
                  {group.description || "Chưa có mô tả nhóm."}
                </p>
                {allowed && (
                  <div className="flex shrink-0 gap-1">
                    <Button
                      variant="ghost"
                      size="icon"
                      title={`Sửa nhóm ${group.name}`}
                      aria-label={`Sửa nhóm ${group.name}`}
                      onClick={() => openGroup(group)}
                    >
                      <Pencil1Icon />
                    </Button>
                    <Button
                      size="sm"
                      variant="ghost"
                      onClick={() => onCreateTask(group.id)}
                    >
                      <PlusIcon />
                      Thêm công việc
                    </Button>
                  </div>
                )}
              </div>
              {rows.length ? (
                <div className="overflow-x-auto">
                  <table className="w-full min-w-[560px] text-left text-sm">
                    <thead className="table-heading table-heading-strong">
                      <tr>
                        <th className="px-4 py-2.5">Công việc</th>
                        <th>Trạng thái</th>
                        <th className="pr-4 text-right">Ước tính</th>
                        <th className="relative w-14">
                          <span className="sr-only">Thao tác</span>
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      {rows.map((task) => (
                        <tr
                          key={task.id}
                          className="border-t border-border hover:bg-background"
                        >
                          <td className="px-4 py-3">
                            <button
                              type="button"
                              className="flex items-start gap-2 text-left hover:underline"
                              onClick={() => onViewTask(task)}
                            >
                              <FileTextIcon className="mt-1 text-muted-foreground" />
                              <span>
                                <span className="block font-medium">
                                  {task.name}
                                </span>
                                <span className="mt-1 block text-xs text-muted-foreground">
                                  {task.code}
                                </span>
                              </span>
                            </button>
                          </td>
                          <td>
                            <StatusBadge status={task.status} />
                          </td>
                          <td className="pr-4 text-right tabular-nums">
                            {task.estimate} giờ
                          </td>
                          <td>
                            {allowed &&
                              (isTaskFrozen(task) ? (
                                <span
                                  title="Đặc tả khóa sau lần phân công đầu"
                                  aria-label="Đặc tả đã khóa"
                                  className="grid size-10 place-items-center text-muted-foreground"
                                >
                                  <LockClosedIcon />
                                </span>
                              ) : (
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  aria-label={`Sửa ${task.name}`}
                                  title="Chỉnh sửa công việc"
                                  onClick={() => onEditTask(task)}
                                >
                                  <Pencil1Icon />
                                </Button>
                              ))}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : (
                <div className="border-t border-dashed border-border px-4 py-7 text-center text-sm text-muted-foreground">
                  Chưa có công việc.
                  {allowed && (
                    <button
                      type="button"
                      className="ml-2 font-medium text-foreground underline underline-offset-4"
                      onClick={() => onCreateTask(group.id)}
                    >
                      Tạo công việc đầu tiên
                    </button>
                  )}
                </div>
              )}
            </details>
          )
        })}
      </div>
      {!relatedGroups.length && (
        <div className="rounded-lg border border-dashed border-border py-12 text-center">
          <p>Chưa có nhóm công việc</p>
          <p className="mt-2 text-sm text-muted-foreground">
            {allowed
              ? "Tạo nhóm để bắt đầu thiết lập cấu trúc dự án."
              : "PTCM sẽ thiết lập cấu trúc dự án."}
          </p>
        </div>
      )}
      <p className="mt-4 text-xs leading-5 text-muted-foreground">
        Đặc tả công việc khóa sau lần giao đầu. Đổi tên nhóm không thay đổi yêu
        cầu của các công việc bên trong.
      </p>
      {editing && (
        <Dialog
          open
          onOpenChange={(open) => {
            if (!open) setEditing(null)
          }}
          title={editing === "new" ? "Tạo nhóm công việc" : "Chỉnh sửa nhóm"}
          description="Nhóm giúp tổ chức công việc trong dự án, không phải danh sách thành viên."
        >
          <form
            noValidate
            className="space-y-4"
            onSubmit={(e) => {
              e.preventDefault()
              try {
                const result = saveGroup(
                  project,
                  actorId,
                  groups,
                  tasks,
                  { name, description },
                  editing === "new" ? undefined : editing.id,
                )
                onChange(result.groups, result.tasks)
                setEditing(null)
                onNotice("Đã lưu nhóm công việc trong phiên mô phỏng.")
              } catch (err) {
                setError((err as Error).message)
              }
            }}
          >
            <div>
              <label
                htmlFor="group-name"
                className="mb-2 block text-sm font-medium"
              >
                Tên nhóm *
              </label>
              <input
                id="group-name"
                className="field w-full"
                value={name}
                maxLength={150}
                onChange={(e) => setName(e.target.value)}
                aria-invalid={Boolean(error)}
                aria-describedby={error ? "group-error" : undefined}
              />
            </div>
            <div>
              <label
                htmlFor="group-description"
                className="mb-2 block text-sm font-medium"
              >
                Mô tả{" "}
                <span className="font-normal text-muted-foreground">
                  · Không bắt buộc
                </span>
              </label>
              <textarea
                id="group-description"
                className="field min-h-24 w-full"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
              />
            </div>
            {error && (
              <p
                id="group-error"
                role="alert"
                className="text-sm text-destructive"
              >
                {error}
              </p>
            )}
            <div className="flex justify-end gap-2 border-t border-border pt-4">
              <Button variant="outline" onClick={() => setEditing(null)}>
                Hủy
              </Button>
              <Button type="submit">
                {editing === "new" ? "Tạo nhóm" : "Lưu thay đổi"}
              </Button>
            </div>
          </form>
        </Dialog>
      )}
    </section>
  )
}
