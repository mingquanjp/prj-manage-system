import { useState } from "react"
import { Pencil1Icon, TrashIcon, PlusIcon } from "@radix-ui/react-icons"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { type Project, type Task } from "@/demo/model"
import {
  dailyMinutes,
  deleteLog,
  formatMinutes,
  logBlockReason,
  saveLog,
  taskLogs,
  vietnamDay,
  WorkValidationError,
  type WorkLog,
  type WorkState,
} from "@/demo/effort"

export function WorkLogPanel({
  project,
  task,
  projects,
  tasks,
  actorId,
  work,
  onChange,
}: {
  project: Project
  task: Task
  projects: Project[]
  tasks: Task[]
  actorId: string
  work: WorkState
  onChange: (work: WorkState) => void
}) {
  const [editor, setEditor] = useState<{ before?: WorkLog } | null>(null)
  const [deleting, setDeleting] = useState<WorkLog | null>(null)
  const [date, setDate] = useState(vietnamDay())
  const [hours, setHours] = useState("0"),
    [minutes, setMinutes] = useState("0")
  const [description, setDescription] = useState(""),
    [overReason, setOverReason] = useState("")
  const [reason, setReason] = useState(""),
    [errors, setErrors] = useState<Record<string, string>>({})
  const [notice, setNotice] = useState("")
  const rows = taskLogs(work, task, actorId).sort(
    (a, b) =>
      b.workDate.localeCompare(a.workDate) ||
      b.createdAt.localeCompare(a.createdAt),
  )
  const block = logBlockReason(project, task, actorId)
  const earliestDay =
    project.executionStartedAt && task.currentAssignment
      ? [
          vietnamDay(new Date(project.executionStartedAt)),
          vietnamDay(new Date(task.currentAssignment.assignedAt)),
        ].sort()[1]
      : undefined
  const duration = Number(hours) * 60 + Number(minutes)
  const totalDay =
    dailyMinutes(work, actorId, date, editor?.before?.id) + duration
  const edits = work.history.filter(
    (h) =>
      h.before.taskId === task.id &&
      h.before.assignmentId === task.currentAssignment?.id &&
      h.before.employeeId === actorId,
  )
  const openEditor = (before?: WorkLog) => {
    setDate(before?.workDate ?? vietnamDay())
    setHours(String(Math.floor((before?.minutes ?? 0) / 60)))
    setMinutes(String((before?.minutes ?? 0) % 60))
    setDescription(before?.description ?? "")
    setOverReason(before?.overLimitReason ?? "")
    setReason("")
    setErrors({})
    setEditor({ before })
  }
  const error = (name: string) =>
    errors[name] ? (
      <p id={`log-${name}-error`} className="mt-1 text-xs text-destructive">
        {errors[name]}
      </p>
    ) : null
  const field = (name: string) => ({
    "aria-invalid": !!errors[name],
    "aria-describedby": errors[name] ? `log-${name}-error` : undefined,
  })
  const failed = (err: unknown) =>
    setErrors(
      err instanceof WorkValidationError
        ? err.fields
        : { form: (err as Error).message },
    )
  return (
    <section aria-label="Nhật ký công việc" className="mt-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h3 className="section-title">Nhật ký · {rows.length}</h3>
        {!block && (
          <Button size="sm" onClick={() => openEditor()}>
            <PlusIcon />
            Ghi nhật ký
          </Button>
        )}
      </div>
      <p className="mt-2 text-xs leading-5 text-muted-foreground">
        Ghi thời gian thực tế dành cho task, gồm ở công ty hoặc ở nhà; không gồm
        nghỉ/chờ thụ động. Không ghi trùng thời gian giữa các task.
      </p>
      {block && (
        <p className="mt-3 rounded-md bg-accent p-3 text-xs leading-5">
          {block}
        </p>
      )}
      {notice && (
        <p role="status" className="mt-3 text-xs text-success">
          {notice}
        </p>
      )}
      <p className="mt-3 text-sm font-medium">
        {rows.length
          ? `Đã ghi nhận: ${formatMinutes(rows.reduce((sum, l) => sum + l.minutes, 0))}`
          : "Chưa có thời gian ghi nhận"}
      </p>
      <p className="mt-1 text-xs text-muted-foreground">
        {rows.length
          ? "Tổng đã khai, không khẳng định đã ghi đủ."
          : "Không đồng nghĩa công việc thực tế chưa tốn thời gian."}
      </p>
      <ul className="mt-4 divide-y divide-border">
        {rows.map((log) => (
          <li key={log.id} className="py-3">
            <div className="flex items-center justify-between gap-2 text-sm">
              <span>{log.workDate.split("-").reverse().join("/")}</span>
              <strong className="font-medium">
                {formatMinutes(log.minutes)}
              </strong>
            </div>
            <p className="mt-2 whitespace-pre-wrap break-words text-sm leading-6">
              {log.description}
            </p>
            {log.overLimitReason && (
              <p className="mt-1 whitespace-pre-wrap break-words text-xs text-muted-foreground">
                Lý do vượt 8 giờ: {log.overLimitReason}
              </p>
            )}
            {!block && (
              <div className="mt-2 flex justify-end gap-1">
                <Button
                  size="icon"
                  variant="ghost"
                  aria-label={`Sửa nhật ký ${log.description}`}
                  title="Chỉnh sửa nhật ký"
                  onClick={() => openEditor(log)}
                >
                  <Pencil1Icon />
                </Button>
                <Button
                  size="icon"
                  variant="ghost"
                  className="text-destructive"
                  aria-label={`Xóa nhật ký ${log.description}`}
                  title="Xóa nhật ký"
                  onClick={() => {
                    setDeleting(log)
                    setReason("")
                    setErrors({})
                  }}
                >
                  <TrashIcon />
                </Button>
              </div>
            )}
          </li>
        ))}
      </ul>
      {edits.length > 0 && (
        <details className="mt-4 border-t border-border pt-3 text-xs">
          <summary className="cursor-pointer font-medium">
            Lịch sử sửa/xóa trong phiên ({edits.length})
          </summary>
          <ul className="mt-3 space-y-3">
            {edits.map((h, i) => (
              <li key={i} className="break-words leading-5">
                <p>
                  {h.after.deletedAt ? "Xóa" : "Sửa"} ·{" "}
                  {new Date(h.at).toLocaleString("vi-VN")} · {h.reason}
                </p>
                <p className="text-muted-foreground">
                  Trước: {h.before.workDate} · {formatMinutes(h.before.minutes)}{" "}
                  · {h.before.description}
                  {h.before.overLimitReason && ` · ${h.before.overLimitReason}`}
                </p>
                {!h.after.deletedAt && (
                  <p className="text-muted-foreground">
                    Sau: {h.after.workDate} · {formatMinutes(h.after.minutes)} ·{" "}
                    {h.after.description}
                    {h.after.overLimitReason && ` · ${h.after.overLimitReason}`}
                  </p>
                )}
              </li>
            ))}
          </ul>
        </details>
      )}
      <Dialog
        open={!!editor}
        onOpenChange={(open) => !open && setEditor(null)}
        title={editor?.before ? "Chỉnh sửa nhật ký" : "Ghi nhật ký công việc"}
        description="Chỉ ghi thời gian thực tế của bạn trong lượt được giao. Không tự ghi giờ từ trạng thái công việc."
      >
        <form
          className="space-y-4"
          noValidate
          onSubmit={(e) => {
            e.preventDefault()
            if (!editor) return
            if (
              !/^\d+$/.test(hours) ||
              !/^\d+$/.test(minutes) ||
              Number(minutes) > 59
            ) {
              setErrors({
                minutes: "Giờ phải là số nguyên không âm; phút từ 0 đến 59.",
              })
              return
            }
            try {
              onChange(
                saveLog(
                  projects,
                  tasks,
                  work,
                  actorId,
                  task.id,
                  {
                    workDate: date,
                    minutes: duration,
                    description,
                    overLimitReason: overReason,
                  },
                  editor.before ? { before: editor.before, reason } : undefined,
                ),
              )
              setEditor(null)
              setNotice(
                editor.before
                  ? "Đã cập nhật nhật ký và giữ lịch sử trong phiên."
                  : "Đã ghi nhật ký trong phiên mô phỏng.",
              )
            } catch (err) {
              failed(err)
            }
          }}
        >
          <label className="block text-sm">
            Ngày làm việc *
            <input
              type="date"
              className="field mt-2 w-full"
              value={date}
              min={earliestDay}
              max={vietnamDay()}
              onChange={(e) => setDate(e.target.value)}
              {...field("workDate")}
            />
            {error("workDate")}
          </label>
          {earliestDay && (
            <p className="text-xs leading-5 text-muted-foreground">
              Được ghi thời gian từ {earliestDay.split("-").reverse().join("/")}
              . Ngày quá hạn kế hoạch vẫn được ghi nếu là thời gian thực tế hợp
              lệ.
            </p>
          )}
          <div className="grid grid-cols-2 gap-3">
            <label className="text-sm">
              Giờ
              <input
                type="number"
                min="0"
                step="1"
                className="field mt-2 w-full"
                value={hours}
                onChange={(e) => setHours(e.target.value)}
                {...field("minutes")}
              />
            </label>
            <label className="text-sm">
              Phút
              <input
                type="number"
                min="0"
                max="59"
                step="1"
                className="field mt-2 w-full"
                value={minutes}
                onChange={(e) => setMinutes(e.target.value)}
                {...field("minutes")}
              />
            </label>
          </div>
          {error("minutes")}
          <p
            className={`text-xs leading-5 ${totalDay > 480 ? "text-warning" : "text-muted-foreground"}`}
          >
            Tổng cá nhân ngày này sau lưu:{" "}
            {Number.isFinite(totalDay) ? formatMinutes(totalDay) : "—"} (xuyên
            dự án). Trên 8 giờ cần lý do; trên 24 giờ không được lưu.
          </p>
          <label className="block text-sm">
            Nội dung đã làm *
            <textarea
              className="field mt-2 min-h-24 w-full"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              {...field("description")}
            />
            {error("description")}
          </label>
          {(totalDay > 480 || overReason || errors.overLimitReason) && (
            <label className="block text-sm">
              Lý do tổng ngày vượt 8 giờ *
              <textarea
                className="field mt-2 w-full"
                value={overReason}
                onChange={(e) => setOverReason(e.target.value)}
                {...field("overLimitReason")}
              />
              {error("overLimitReason")}
            </label>
          )}
          {editor?.before && (
            <label className="block text-sm">
              Lý do chỉnh sửa *
              <textarea
                className="field mt-2 w-full"
                value={reason}
                onChange={(e) => setReason(e.target.value)}
                {...field("reason")}
              />
              {error("reason")}
            </label>
          )}
          {errors.form && (
            <p role="alert" className="text-sm text-destructive">
              {errors.form}
            </p>
          )}
          {Object.keys(errors).length > 0 && (
            <p role="alert" className="sr-only">
              Chưa thể lưu nhật ký. Kiểm tra các trường được đánh dấu.
            </p>
          )}
          <div className="flex justify-end gap-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => setEditor(null)}
            >
              Hủy
            </Button>
            <Button type="submit">Lưu nhật ký</Button>
          </div>
        </form>
      </Dialog>
      <Dialog
        open={!!deleting}
        onOpenChange={(open) => !open && setDeleting(null)}
        title="Xóa nhật ký?"
        description="Nhật ký bị loại khỏi tổng thời gian nhưng vẫn giữ lịch sử. Không có thao tác khôi phục trong phiên bản hiện tại."
      >
        <form
          onSubmit={(e) => {
            e.preventDefault()
            if (!deleting) return
            try {
              onChange(
                deleteLog(
                  projects,
                  tasks,
                  work,
                  actorId,
                  task.id,
                  deleting,
                  reason,
                ),
              )
              setDeleting(null)
              setNotice("Đã xóa nhật ký khỏi tổng thời gian và giữ lịch sử.")
            } catch (err) {
              failed(err)
            }
          }}
        >
          <label className="block text-sm">
            Lý do xóa *
            <textarea
              className="field mt-2 w-full"
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              {...field("reason")}
            />
            {error("reason")}
          </label>
          {errors.form && (
            <p role="alert" className="mt-2 text-sm text-destructive">
              {errors.form}
            </p>
          )}
          <div className="mt-5 flex justify-end gap-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => setDeleting(null)}
            >
              Hủy
            </Button>
            <Button
              type="submit"
              className="bg-destructive text-white hover:bg-destructive/90"
            >
              Xóa nhật ký
            </Button>
          </div>
        </form>
      </Dialog>
    </section>
  )
}
