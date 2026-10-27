import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { usePeople } from "@/components/people-directory"
import { officialValues, sameSkillValues } from "@/demo/employee-skills"
import type { EmployeeSkillsState, SkillDecision, SkillRequest } from "@/models/employee-skills"
import { SkillRequestBadge, SkillRequestDetails, SkillValuesView } from "./SkillRequestDetails"

const rubric = [
  "Cơ bản: làm việc cơ bản khi có hướng dẫn.",
  "Trung cấp: tự làm việc thường gặp, biết khi nào cần hỗ trợ.",
  "Nâng cao: xử lý việc phức tạp, giải thích lựa chọn và giải quyết phát sinh.",
  "Chuyên gia: định hướng giải pháp, xử lý vấn đề khó và hỗ trợ người khác.",
]
export function SkillReviewDialog({
  request,
  state,
  actorId,
  employeeCode,
  employeeActive,
  onResolve,
  onClose,
}: {
  request: SkillRequest
  state: EmployeeSkillsState
  actorId: string
  employeeCode?: string
  employeeActive: boolean
  onResolve: (decision: SkillDecision) => void
  onClose: () => void
}) {
  const { getPerson } = usePeople()
  const [decision, setDecision] = useState<"APPROVED" | "REJECTED">()
  const [reason, setReason] = useState("")
  const [opinion, setOpinion] = useState("")
  const [error, setError] = useState("")
  const [discard, setDiscard] = useState(false)
  const live = state.requests.find((r) => r.id === request.id)
  const current = officialValues(state, request.employeeId, request.skillId)
  const stale = live?.version !== request.version || live?.status !== request.status
  const conflict = request.status === "PENDING" && !sameSkillValues(request.before, current)
  const prior = state.requests
    .filter(
      (r) =>
        r.employeeId === request.employeeId &&
        r.skillId === request.skillId &&
        r.id !== request.id &&
        r.status !== "PENDING" &&
        r.requestedAt <= request.requestedAt,
    )
    .reverse()
  const self = request.employeeId === actorId
  const disabled = stale || self || !employeeActive
  const close = () => (reason.trim() || opinion.trim() ? setDiscard(true) : onClose())
  return (
    <>
      <Dialog
        side
        alert={request.operation === "REMOVE" && decision === "APPROVED"}
        open
        onOpenChange={(open) => {
          if (!open) close()
        }}
        title={`Xét duyệt · ${getPerson(request.employeeId)?.name}`}
        description="Đối chiếu mức đề nghị với đóng góp, kết quả và căn cứ được cung cấp."
      >
        <div className="mb-5 space-y-1 text-sm text-muted-foreground">
          <p>
            {employeeCode ?? "Chưa có mã nhân viên"} · {getPerson(request.employeeId)?.position}
          </p>
          <p className="tabular-nums">
            Gửi lúc {new Date(request.requestedAt).toLocaleString("vi-VN")}
          </p>
        </div>
        <SkillRequestDetails
          request={request}
          targetGuidance={request.proposed ? rubric[request.proposed.level - 1] : undefined}
        />
        <details className="mt-5 rounded-lg border border-border p-3">
          <summary className="cursor-pointer text-sm font-medium">
            Lịch sử cùng kỹ năng ({prior.length})
          </summary>
          {!prior.length && (
            <p className="mt-3 text-sm text-muted-foreground">
              Chưa có quyết định trước đề nghị này trong phiên mô phỏng.
            </p>
          )}
          {prior.map((r) => (
            <section key={r.id} className="mt-3 space-y-2 border-t border-border pt-3 text-sm">
              <SkillRequestBadge status={r.status} />
              <SkillValuesView value={r.proposed} />
              <p className="whitespace-pre-wrap break-words text-pretty">{r.decisionReason}</p>
              <p className="text-xs tabular-nums text-muted-foreground">
                {getPerson(r.resolvedBy)?.name} ·{" "}
                {r.resolvedAt && new Date(r.resolvedAt).toLocaleString("vi-VN")}
              </p>
            </section>
          ))}
        </details>
        <details className="mt-5 rounded-lg border border-border p-3">
          <summary className="cursor-pointer text-sm font-medium">
            Hướng dẫn đối chiếu bốn mức
          </summary>
          <ul className="mt-3 space-y-2 text-pretty text-sm leading-6">
            {rubric.map((text) => (
              <li key={text}>{text}</li>
            ))}
          </ul>
          <p className="mt-3 text-pretty text-xs leading-5 text-muted-foreground">
            Số năm, số task hoặc điểm gợi ý không tự quyết định mức. Task được nghiệm thu không tự
            công nhận tất cả kỹ năng yêu cầu.
          </p>
        </details>
        {conflict && (
          <section className="mt-5 rounded-lg bg-warning-soft p-3">
            <h3 className="field-label mb-2">Năng lực hiện tại đã khác thời điểm gửi</h3>
            <SkillValuesView value={current} />
            <p className="mt-2 text-sm text-warning">
              Không thể chấp thuận và ghi đè bằng đề nghị cũ.
            </p>
          </section>
        )}
        {request.status === "PENDING" && (
          <>
            {(stale || self || !employeeActive) && (
              <p role="alert" className="mt-5 text-pretty text-sm text-warning">
                {stale
                  ? "Đề nghị đã thay đổi. Đóng và mở lại để đọc căn cứ mới nhất."
                  : self
                    ? "Bạn không được xét duyệt đề nghị của chính mình."
                    : "Hồ sơ nhân viên không còn hoạt động."}
              </p>
            )}
            <section className="mt-5 space-y-2">
              <label className="block">
                <span className="field-label">Ý kiến chuyên môn ngoài hệ thống (nếu có)</span>
                <textarea
                  className="field mt-2 min-h-20 w-full"
                  value={opinion}
                  onChange={(e) => setOpinion(e.target.value)}
                  placeholder="Nguồn, người nhận xét, nội dung quan sát và đánh giá."
                />
              </label>
              <p className="text-pretty text-xs leading-5 text-muted-foreground">
                Nếu chưa đủ chuyên môn để đánh giá, cần tham khảo và ghi nhận ý kiến phù hợp trước
                khi chấp thuận. Đây không phải một cấp duyệt khác.
              </p>
            </section>
            {!decision ? (
              <div className="mt-6 flex justify-end gap-2 border-t border-border pt-4">
                <Button
                  disabled={disabled}
                  variant="outline"
                  onClick={() => setDecision("REJECTED")}
                >
                  Từ chối
                </Button>
                <Button disabled={disabled || conflict} onClick={() => setDecision("APPROVED")}>
                  Chấp thuận
                </Button>
              </div>
            ) : (
              <form
                className="mt-5 space-y-4 border-t border-border pt-4"
                onSubmit={(event) => {
                  event.preventDefault()
                  try {
                    onResolve({
                      requestId: request.id,
                      version: request.version,
                      decision,
                      reason,
                      expertOpinion: opinion,
                    })
                    onClose()
                  } catch (err) {
                    setError(err instanceof Error ? err.message : "Chưa thể xét duyệt.")
                  }
                }}
              >
                <h3 className="section-title">
                  {decision === "APPROVED" ? "Xác nhận chấp thuận" : "Xác nhận từ chối"}
                </h3>
                <label className="block">
                  <span className="field-label">Lý do quyết định *</span>
                  <textarea
                    autoFocus
                    aria-required="true"
                    className="field mt-2 min-h-24 w-full"
                    value={reason}
                    onChange={(e) => setReason(e.target.value)}
                  />
                </label>
                {error && (
                  <p role="alert" className="text-sm text-warning">
                    {error}
                  </p>
                )}
                <div className="flex flex-wrap justify-end gap-2">
                  <Button variant="outline" onClick={() => setDecision(undefined)}>
                    Quay lại
                  </Button>
                  <Button
                    type="submit"
                    disabled={disabled || (decision === "APPROVED" && conflict)}
                  >
                    {decision === "APPROVED" ? "Xác nhận chấp thuận" : "Xác nhận từ chối"}
                  </Button>
                </div>
              </form>
            )}
          </>
        )}
      </Dialog>
      <Dialog
        open={discard}
        onOpenChange={setDiscard}
        title="Bỏ nội dung xét duyệt chưa lưu?"
        description="Đề nghị chưa được xử lý và năng lực chính thức giữ nguyên."
      >
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={() => setDiscard(false)}>
            Tiếp tục xử lý
          </Button>
          <Button onClick={onClose}>Bỏ nội dung</Button>
        </div>
      </Dialog>
    </>
  )
}
