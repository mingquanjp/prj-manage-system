import { usePeople } from "@/components/people-directory"
import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import type { Project, Task } from "@/models/workspace"
import {
  formatMinutes,
  WorkValidationError,
  type LogChangeRequest,
  type LogInput,
  type WorkState,
} from "@/demo/effort"
import {
  canReadRequest,
  changeLabels,
  contributionFor,
  requestConflict,
  requestLabels,
  resolveLogChange,
} from "@/demo/corrections"

function LogValues({ value }: { value?: LogInput }) {
  if (!value) return <p className="text-sm text-muted-foreground">Không có</p>
  return (
    <div className="space-y-1 text-sm leading-6">
      <p className="tabular-nums">
        {value.workDate.split("-").reverse().join("/")} · {formatMinutes(value.minutes)}
      </p>
      <p className="whitespace-pre-wrap break-words">{value.description}</p>
      {value.overLimitReason && (
        <p className="break-words text-xs text-muted-foreground">
          Lý do vượt 8 giờ: {value.overLimitReason}
        </p>
      )}
    </div>
  )
}

export function CorrectionRequests({
  projects,
  tasks,
  work,
  actorId,
  assignmentId,
  projectId,
  onChange,
}: {
  projects: Project[]
  tasks: Task[]
  work: WorkState
  actorId: string
  assignmentId?: string
  projectId?: string
  onChange: (work: WorkState) => void
}) {
  const { getPerson } = usePeople()
  const [filter, setFilter] = useState("all")
  const [decision, setDecision] = useState<{
    request: LogChangeRequest
    action: "APPROVED" | "REJECTED" | "WITHDRAWN"
  }>()
  const [reason, setReason] = useState("")
  const [error, setError] = useState("")
  const [notice, setNotice] = useState("")
  const rows = work.requests
    .filter((r) => {
      const contribution = contributionFor(work, tasks, r.assignmentId)
      return (
        canReadRequest(projects, tasks, work, r, actorId) &&
        (projectId
          ? contribution?.projectId === projectId &&
            projects.some((p) => p.id === projectId && p.ptcmId === actorId)
          : r.requestedBy === actorId) &&
        (!assignmentId || r.assignmentId === assignmentId)
      )
    })
    .slice()
    .reverse()
  const visibleRows = rows.filter((request) => filter === "all" || request.status === filter)
  const readableRequestIds = new Set(rows.map((request) => request.id))
  const notifications = work.notifications.filter(
    (notification) =>
      notification.recipientId === actorId &&
      !notification.readAt &&
      readableRequestIds.has(notification.requestId),
  )
  const markNotificationsRead = () => {
    const ids = new Set(notifications.map((notification) => notification.id))
    const readAt = new Date().toISOString()
    onChange({
      ...work,
      notifications: work.notifications.map((notification) =>
        ids.has(notification.id) ? { ...notification, readAt } : notification,
      ),
    })
  }
  const open = (request: LogChangeRequest, action: "APPROVED" | "REJECTED" | "WITHDRAWN") => {
    setDecision({ request: structuredClone(request), action })
    setReason("")
    setError("")
  }
  const conflict = decision ? requestConflict(work, decision.request) : undefined
  return (
    <section className="mt-5 border-t border-border pt-4" aria-label="Yêu cầu hiệu chỉnh nhật ký">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h3 className="section-title">Yêu cầu hiệu chỉnh · {rows.length}</h3>
        {rows.length > 0 && (
          <label className="flex items-center gap-2 text-sm">
            <span className="field-label">Trạng thái yêu cầu</span>
            <select className="field" value={filter} onChange={(e) => setFilter(e.target.value)}>
              <option value="all">Tất cả</option>
              {Object.entries(requestLabels).map(([key, value]) => (
                <option key={key} value={key}>
                  {value}
                </option>
              ))}
            </select>
          </label>
        )}
      </div>
      {notifications.length > 0 && (
        <div className="mt-3 rounded-md bg-accent p-3 text-sm" role="status">
          <p>{notifications.length} thông báo hiệu chỉnh chưa đọc trong phiên.</p>
          <ul className="mt-2 space-y-1 text-xs">
            {notifications.slice(0, 3).map((n) => (
              <li key={n.id}>{n.message}</li>
            ))}
          </ul>
          <Button size="sm" variant="ghost" className="mt-2" onClick={markNotificationsRead}>
            Đánh dấu đã đọc
          </Button>
        </div>
      )}
      <p className="mt-2 text-xs leading-5 text-muted-foreground">
        Yêu cầu chờ duyệt không thay đổi tổng giờ. Hiệu chỉnh không mở lại công việc hoặc sửa
        snapshot lần nộp.
      </p>
      {notice && (
        <p role="status" className="mt-3 text-sm text-success">
          {notice}
        </p>
      )}
      {!visibleRows.length && (
        <p className="py-6 text-sm text-muted-foreground">Chưa có yêu cầu phù hợp.</p>
      )}
      <ul className="mt-3 space-y-3">
        {visibleRows.map((r) => {
          const contribution = contributionFor(work, tasks, r.assignmentId)!
          const project = projects.find((p) => p.id === contribution.projectId)!
          const writable = ["IN_PROGRESS", "COMPLETED"].includes(project.status)
          const pending = r.status === "PENDING"
          const endedByCancellation =
            r.status === "REJECTED" && r.decisionReason?.startsWith("PROJECT_CANCELLED")
          const stale = requestConflict(work, r)
          return (
            <li key={r.id} className="rounded-lg border border-border bg-white p-4">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <p className="font-medium text-sm">
                  {changeLabels[r.operation]} nhật ký · {contribution.taskCode}
                </p>
                <span className="rounded bg-accent px-2 py-1 text-xs">
                  {endedByCancellation ? "Kết thúc do hủy dự án" : requestLabels[r.status]}
                </span>
              </div>
              <p className="mt-1 text-sm text-muted-foreground">{contribution.taskName}</p>
              <p className="mt-2 text-xs text-muted-foreground">
                {getPerson(r.requestedBy)?.name} · {new Date(r.requestedAt).toLocaleString("vi-VN")}
              </p>
              <p className="mt-2 whitespace-pre-wrap break-words text-sm">{r.reason}</p>
              <details className="mt-3 text-sm">
                <summary className="field-label cursor-pointer">So sánh trước và đề nghị</summary>
                <div className="mt-3 grid gap-4 sm:grid-cols-2">
                  <div>
                    <p className="field-label mb-2">Trước khi gửi</p>
                    <LogValues value={r.before} />
                  </div>
                  <div>
                    <p className="field-label mb-2">
                      Đề nghị {r.operation === "DELETE" ? "xóa khỏi tổng" : "cập nhật"}
                    </p>
                    <LogValues value={r.proposed} />
                  </div>
                </div>
              </details>
              {pending && stale && <p className="mt-3 text-xs text-warning">{stale}</p>}
              {r.resolvedAt && (
                <p className="mt-3 break-words text-xs leading-5 text-muted-foreground">
                  {getPerson(r.resolvedBy)?.name} · {new Date(r.resolvedAt).toLocaleString("vi-VN")}
                  {r.decisionReason &&
                    ` · ${endedByCancellation ? r.decisionReason.replace(/^PROJECT_CANCELLED\s*·?\s*/, "Hủy dự án: ") : r.decisionReason}`}
                  {r.resultLogId && " · Đã liên kết nhật ký được tạo từ yêu cầu này."}
                </p>
              )}
              {pending && !writable && (
                <p className="mt-3 text-xs text-muted-foreground">Dự án đã khóa, chỉ được đọc.</p>
              )}
              {pending && writable && (
                <div className="mt-3 flex flex-wrap justify-end gap-2">
                  {r.requestedBy === actorId && (
                    <Button size="sm" variant="outline" onClick={() => open(r, "WITHDRAWN")}>
                      Rút yêu cầu
                    </Button>
                  )}
                  {projectId && project.ptcmId === actorId && (
                    <>
                      <Button size="sm" variant="outline" onClick={() => open(r, "REJECTED")}>
                        Từ chối
                      </Button>
                      <Button
                        size="sm"
                        disabled={Boolean(stale)}
                        onClick={() => open(r, "APPROVED")}
                      >
                        Xét chấp thuận
                      </Button>
                    </>
                  )}
                </div>
              )}
            </li>
          )
        })}
      </ul>
      <Dialog
        open={Boolean(decision)}
        onOpenChange={(value) => {
          if (!value) setDecision(undefined)
        }}
        title={
          decision?.action === "APPROVED"
            ? "Chấp thuận hiệu chỉnh?"
            : decision?.action === "REJECTED"
              ? "Từ chối hiệu chỉnh"
              : "Rút yêu cầu hiệu chỉnh?"
        }
        description="Yêu cầu đã kết thúc không được xử lý lại. Nhật ký/snapshot và trạng thái công việc là các dữ liệu riêng biệt."
      >
        {decision && (
          <form
            noValidate
            className="space-y-4"
            onSubmit={(e) => {
              e.preventDefault()
              try {
                onChange(
                  resolveLogChange(
                    projects,
                    tasks,
                    work,
                    actorId,
                    decision.request.id,
                    decision.action,
                    reason,
                  ),
                )
                setNotice(
                  decision.action === "APPROVED"
                    ? "Đã chấp thuận và cập nhật nhật ký trong phiên. Snapshot cũ giữ nguyên."
                    : decision.action === "REJECTED"
                      ? "Đã từ chối. Nhật ký không thay đổi."
                      : "Đã rút yêu cầu. Nhật ký không thay đổi.",
                )
                setDecision(undefined)
              } catch (err) {
                setError(
                  err instanceof WorkValidationError
                    ? Object.values(err.fields).join(" ")
                    : (err as Error).message,
                )
              }
            }}
          >
            <div className="grid gap-4 sm:grid-cols-2">
              <div>
                <p className="field-label mb-2">Trước khi gửi</p>
                <LogValues value={decision.request.before} />
              </div>
              <div>
                <p className="field-label mb-2">
                  Đề nghị ({changeLabels[decision.request.operation]})
                </p>
                <LogValues value={decision.request.proposed} />
              </div>
            </div>
            {conflict && <p className="text-sm text-warning">{conflict}</p>}
            <label className="block">
              <span className="field-label">
                {decision.action === "REJECTED" ? "Lý do từ chối *" : "Ghi chú (không bắt buộc)"}
              </span>
              <textarea
                className="field mt-2 w-full"
                value={reason}
                onChange={(e) => setReason(e.target.value)}
                aria-invalid={Boolean(error)}
              />
            </label>
            {error && (
              <p role="alert" className="text-sm text-destructive">
                {error}
              </p>
            )}
            <div className="flex flex-wrap justify-end gap-2">
              <Button type="button" variant="outline" onClick={() => setDecision(undefined)}>
                Quay lại
              </Button>
              <Button type="submit" disabled={decision.action === "APPROVED" && Boolean(conflict)}>
                {decision.action === "APPROVED"
                  ? "Chấp thuận"
                  : decision.action === "REJECTED"
                    ? "Xác nhận từ chối"
                    : "Xác nhận rút"}
              </Button>
            </div>
          </form>
        )}
      </Dialog>
    </section>
  )
}
