import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { SubmissionHistory } from "@/components/submission-history"
import { reviewBlockReason, type WorkState, type ReviewResult } from "@/demo/effort"
import type { Project, Task, OutputConflict } from "@/models/workspace"

export function TaskReviewPanel({
  project,
  task,
  work,
  actorId,
  conflicts,
  onReview,
}: {
  project: Project
  task: Task
  work: WorkState
  actorId: string
  conflicts: OutputConflict[]
  onReview: (submissionId: string, result: ReviewResult, comment: string) => void
}) {
  const [decision, setDecision] = useState<{
    result: ReviewResult
    submissionId: string
  } | null>(null)
  const [comment, setComment] = useState("")
  const [error, setError] = useState("")
  const submissions = work.submissions.filter((s) => s.taskId === task.id)
  const latest = submissions.at(-1)
  const isPtcm = project.ptcmId === actorId
  const block = reviewBlockReason(project, task, work, actorId, latest?.id ?? "", conflicts)
  return (
    <section className="mt-4" aria-label="Kết quả và nghiệm thu công việc">
      <h3 className="section-title">Kết quả và nghiệm thu</h3>
      <p className="mt-2 text-xs leading-5 text-muted-foreground">
        Đối chiếu đầu ra với tiêu chí đã chốt. Thiếu nhật ký không tự chặn nghiệm thu; không yêu cầu
        sửa ngoài phạm vi công việc.
      </p>
      {!isPtcm && (
        <p className="mt-3 rounded-md bg-accent p-3 text-xs leading-5">
          Bạn đang theo dõi kết quả. Chỉ PTCM hiện hành được quyết định nghiệm thu; bản chụp nhật ký
          chi tiết không mở từ tư cách PM.
        </p>
      )}
      {!latest && (
        <p className="mt-3 rounded-md bg-accent p-3 text-sm leading-6">
          {["PENDING_REVIEW", "DONE", "CLOSED"].includes(task.status)
            ? "Dữ liệu mẫu cũ chưa có chi tiết lần nộp. Không thể quyết định nghiệm thu khi thiếu bản nộp; hãy thử luồng nộp từ Công việc của tôi."
            : "Chưa có lần nộp kết quả."}
        </p>
      )}
      {latest && isPtcm && !block && (
        <div className="sticky top-0 z-10 mt-3 flex flex-wrap gap-2 border-y border-border bg-white py-3">
          {(
            [
              ["REJECTED", "Yêu cầu chỉnh sửa"],
              ["ACCEPTED", "Chấp nhận kết quả"],
            ] as const
          ).map(([result, label]) => (
            <Button
              key={result}
              size="sm"
              variant={result === "ACCEPTED" ? "default" : "outline"}
              onClick={() => {
                setDecision({ result, submissionId: latest.id })
                setComment("")
                setError("")
              }}
            >
              {label}
            </Button>
          ))}
        </div>
      )}
      {latest && isPtcm && block && task.status === "PENDING_REVIEW" && (
        <p className="mt-3 text-xs text-muted-foreground">{block}</p>
      )}
      <SubmissionHistory submissions={submissions} showLogs={isPtcm} />
      <Dialog
        open={!!decision}
        onOpenChange={(open) => {
          if (!open) setDecision(null)
        }}
        title={
          decision?.result === "REJECTED"
            ? "Yêu cầu chỉnh sửa đầu ra"
            : "Chấp nhận kết quả công việc?"
        }
        description={
          decision?.result === "REJECTED"
            ? "Nêu phần chưa đáp ứng tiêu chí. Công việc trở lại Đang thực hiện để người được giao sửa và nộp lại; lần nộp cũ được giữ nguyên."
            : "Công việc chuyển sang Đã nghiệm thu, kết thúc lượt giao và không mở lại. Thao tác này không tự hoàn thành dự án hoặc xác nhận nhật ký đã đầy đủ."
        }
      >
        <form
          noValidate
          onSubmit={(e) => {
            e.preventDefault()
            if (!decision) return
            if (decision.result === "REJECTED" && !comment.trim()) {
              setError("Nhập lý do yêu cầu chỉnh sửa đầu ra.")
              return
            }
            try {
              onReview(decision.submissionId, decision.result, comment)
              setDecision(null)
            } catch (err) {
              setError((err as Error).message)
            }
          }}
        >
          <p className="mb-4 text-sm font-medium">
            {task.code} · {task.name}
          </p>
          <label className="block text-sm">
            <span className="field-label">
              {decision?.result === "REJECTED"
                ? "Lý do yêu cầu chỉnh sửa *"
                : "Nhận xét (tùy chọn)"}
            </span>
            <textarea
              className="field mt-2 min-h-28 w-full"
              value={comment}
              onChange={(e) => setComment(e.target.value)}
              aria-invalid={!!error}
              aria-describedby={error ? "review-error" : undefined}
            />
          </label>
          {error && (
            <p id="review-error" role="alert" className="mt-2 text-sm text-destructive">
              {error}
            </p>
          )}
          <div className="mt-5 flex flex-wrap justify-end gap-2">
            <Button type="button" variant="outline" onClick={() => setDecision(null)}>
              Hủy
            </Button>
            <Button type="submit">
              {decision?.result === "REJECTED" ? "Xác nhận yêu cầu sửa" : "Xác nhận chấp nhận"}
            </Button>
          </div>
        </form>
      </Dialog>
    </section>
  )
}
