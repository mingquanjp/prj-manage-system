import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { type Task, type Project } from "@/demo/model"
import { SubmissionHistory } from "@/components/submission-history"
import {
  logBlockReason,
  taskLogs,
  formatMinutes,
  WorkValidationError,
  validateResultInput,
  type WorkState,
} from "@/demo/effort"

export function ResultPanel({
  task,
  project,
  actorId,
  work,
  onSubmit,
}: {
  task: Task
  project: Project
  actorId: string
  work: WorkState
  onSubmit: (taskId: string, summary: string, links: string) => void
}) {
  const [open, setOpen] = useState(false),
    [confirming, setConfirming] = useState(false)
  const [summary, setSummary] = useState(""),
    [links, setLinks] = useState("")
  const [errors, setErrors] = useState<Record<string, string>>({})
  const submissions = work.submissions.filter(
    (s) => s.taskId === task.id && s.employeeId === actorId,
  )
  const logs = taskLogs(work, task, actorId)
  const block = logBlockReason(project, task, actorId)
  const allowed = !block && task.status === "IN_PROGRESS"
  return (
    <section aria-label="Kết quả công việc" className="mt-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h3 className="section-title">
          Các lần nộp · {submissions.length}
        </h3>
        {allowed && (
          <Button
            size="sm"
            onClick={() => {
              setOpen(true)
              setConfirming(false)
              setErrors({})
            }}
          >
            Nộp kết quả
          </Button>
        )}
      </div>
      <p className="mt-3 text-xs leading-5 text-muted-foreground">
        {task.status === "PENDING_REVIEW"
          ? "Kết quả đang chờ PTCM nghiệm thu. Không chỉnh sửa lần nộp hoặc nhật ký trực tiếp."
          : allowed
            ? "Nộp tóm tắt đầu ra và link tham khảo nếu có. Thiếu nhật ký vẫn được nộp, không khai giờ phỏng đoán."
            : (block ??
              "Bắt đầu hoặc tiếp tục công việc trước khi nộp kết quả.")}
      </p>
      {!submissions.length && (
        <p className="mt-4 rounded-md bg-accent p-3 text-sm leading-6">
          {["PENDING_REVIEW", "DONE", "CLOSED"].includes(task.status)
            ? "Dữ liệu mẫu cũ chưa có chi tiết lần nộp. Không tạo kết quả hoặc quyết định nghiệm thu giả."
            : "Chưa có lần nộp kết quả."}
        </p>
      )}
      <SubmissionHistory submissions={submissions} />
      <Dialog
        open={open}
        onOpenChange={setOpen}
        title={confirming ? "Xác nhận nộp kết quả?" : "Nộp kết quả công việc"}
        description={
          confirming
            ? "Sau khi nộp, công việc chuyển sang Chờ nghiệm thu và nhật ký khóa sửa trực tiếp. Nội dung lần nộp được giữ nguyên."
            : "Tóm tắt đầu ra theo tiêu chí đã chốt. Link chỉ là minh chứng tham khảo, hệ thống không kiểm tra PR hoặc tự nghiệm thu."
        }
      >
        {confirming ? (
          <div>
            <p className="whitespace-pre-wrap break-words text-sm leading-6">
              {summary}
            </p>
            {links.trim() && (
              <p className="mt-3 whitespace-pre-wrap break-all text-xs text-muted-foreground">
                Link minh chứng:
                <br />
                {links.trim()}
              </p>
            )}
            <p className="mt-4 text-sm">
              Nhật ký đính kèm:{" "}
              {logs.length
                ? `${logs.length} mục · ${formatMinutes(logs.reduce((sum, l) => sum + l.minutes, 0))}`
                : "Chưa có thời gian ghi nhận — vẫn được nộp."}
            </p>
            {Object.keys(errors).length > 0 && (
              <p role="alert" className="mt-3 text-sm text-destructive">
                {Object.values(errors).join(" ")}
              </p>
            )}
            <div className="mt-5 flex justify-end gap-2">
              <Button variant="outline" onClick={() => setConfirming(false)}>
                Quay lại
              </Button>
              <Button
                onClick={() => {
                  try {
                    onSubmit(task.id, summary, links)
                    setOpen(false)
                    setSummary("")
                    setLinks("")
                  } catch (err) {
                    setErrors(
                      err instanceof WorkValidationError
                        ? err.fields
                        : { form: (err as Error).message },
                    )
                    setConfirming(false)
                  }
                }}
              >
                Xác nhận nộp
              </Button>
            </div>
          </div>
        ) : (
          <form
            noValidate
            className="space-y-4"
            onSubmit={(e) => {
              e.preventDefault()
              try {
                validateResultInput(summary, links)
                setErrors({})
                setConfirming(true)
              } catch (err) {
                setErrors(
                  err instanceof WorkValidationError
                    ? err.fields
                    : { form: (err as Error).message },
                )
              }
            }}
          >
            <label className="block text-sm">
              Tóm tắt kết quả *
              <textarea
                className="field mt-2 min-h-32 w-full"
                value={summary}
                onChange={(e) => setSummary(e.target.value)}
                aria-invalid={!!errors.summary}
                aria-describedby={
                  errors.summary ? "result-summary-error" : undefined
                }
              />
              {errors.summary && (
                <p
                  id="result-summary-error"
                  className="mt-1 text-xs text-destructive"
                >
                  {errors.summary}
                </p>
              )}
            </label>
            <label className="block text-sm">
              Link minh chứng (tùy chọn, mỗi dòng một URL)
              <textarea
                className="field mt-2 min-h-24 w-full"
                value={links}
                onChange={(e) => setLinks(e.target.value)}
                aria-invalid={!!errors.links}
                aria-describedby="result-links-hint"
              />
              <span
                id="result-links-hint"
                className="mt-1 block text-xs text-muted-foreground"
              >
                Chỉ http/https. Không gửi token, mật khẩu hoặc link chứa bí mật.
              </span>
              {errors.links && (
                <p className="mt-1 text-xs text-destructive">{errors.links}</p>
              )}
            </label>
            {errors.form && (
              <p role="alert" className="text-sm text-destructive">
                {errors.form}
              </p>
            )}
            <div className="flex justify-end gap-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => setOpen(false)}
              >
                Hủy
              </Button>
              <Button type="submit">Kiểm tra trước khi nộp</Button>
            </div>
          </form>
        )}
      </Dialog>
    </section>
  )
}
