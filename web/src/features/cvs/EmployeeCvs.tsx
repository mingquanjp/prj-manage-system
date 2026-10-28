import { useIdentity } from "@/components/identity"
import { useState } from "react"
import { FileTextIcon, PlusIcon, CheckIcon } from "@radix-ui/react-icons"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { EmptyState } from "@/components/shared"
import { usePeople } from "@/components/people-directory"
import { canManageCvs } from "@/demo/employee-cvs"
import { cvProcessingLabels } from "@/models/employee-cvs"
import type { CvFileInput, CvProcessingStatus, EmployeeCvsState } from "@/models/employee-cvs"
import { CvUploadForm } from "./CvUploadForm"

export type CvActions = {
  addCv: (employeeId: string, file: CvFileInput) => void
  selectCurrentCv: (cvId: string) => void
  simulateCvProcessing: (cvId: string, status: CvProcessingStatus) => void
}

export function EmployeeCvs({
  actorId,
  employeeId,
  isActive,
  state,
  actions,
}: {
  actorId: string
  employeeId: string
  isActive: boolean
  state: EmployeeCvsState
  actions: CvActions
}) {
  const identity = useIdentity()

  const { getPerson } = usePeople()
  const [uploading, setUploading] = useState(false)
  const [confirmId, setConfirmId] = useState<string | null>(null)
  const [error, setError] = useState("")
  if (!canManageCvs(actorId, employeeId, identity))
    return (
      <EmptyState
        title="Không có quyền quản lý CV"
        detail="CV gốc không thuộc thông tin tra cứu nghề nghiệp của người khác."
      />
    )
  const isOwner = actorId === employeeId
  const versions = state.versions
    .filter((cv) => cv.employeeId === employeeId)
    .slice()
    .reverse()
  const current = versions.find((cv) => cv.isCurrent)
  const selected = versions.find((cv) => cv.id === confirmId)
  return (
    <section aria-label="Các phiên bản CV" className="min-w-0">
      <div className="mb-4 flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="section-title">{isOwner ? "CV của tôi" : "CV nhân sự"}</h2>
          <p className="mt-1 text-xs leading-5 text-muted-foreground">
            {isOwner
              ? "Bạn tự bổ sung và chọn CV hiện tại; các phiên bản trước được giữ lại."
              : "Nhân viên chủ động cập nhật CV. HR hỗ trợ bổ sung thay khi cần; nhân viên tự chọn bản hiện tại."}
          </p>
        </div>
        {isActive && (
          <Button size="sm" variant="outline" onClick={() => setUploading(true)}>
            <PlusIcon /> {isOwner ? "Bổ sung CV" : "Bổ sung CV thay nhân viên"}
          </Button>
        )}
      </div>
      <div className="mb-4 rounded-lg border border-border bg-accent/40 px-4 py-3 text-sm">
        <p className="font-medium">{current ? "CV hiện tại" : "Chưa chọn CV hiện tại"}</p>
        <p className="mt-1 break-words text-muted-foreground">
          {current?.filename ??
            (!isActive
              ? "Nhân viên chưa xác nhận bản sử dụng. Hồ sơ đã vô hiệu hóa."
              : isOwner
                ? versions.length
                  ? "Chọn một phiên bản bên dưới làm CV hiện tại."
                  : "Bổ sung CV rồi chọn bản bạn muốn sử dụng."
                : "Nhân viên chưa xác nhận bản sử dụng.")}
        </p>
        <p className="mt-2 text-xs leading-5 text-muted-foreground">
          Lựa chọn CV không xác nhận nội dung hoặc thay đổi kỹ năng chính thức.
        </p>
      </div>
      {versions.length ? (
        <div className="space-y-3">
          {versions.map((cv) => (
            <article key={cv.id} className="rounded-lg border border-border bg-white p-4">
              <div className="flex items-start gap-3">
                <FileTextIcon className="mt-1 size-4 shrink-0 text-muted-foreground" />
                <div className="min-w-0 flex-1">
                  <div className="flex flex-wrap items-center gap-2">
                    <h3 className="min-w-0 break-all text-sm font-medium">{cv.filename}</h3>
                    {cv.isCurrent && (
                      <span className="inline-flex items-center gap-1 rounded bg-success/10 px-2 py-1 text-xs font-medium text-success">
                        <CheckIcon /> Hiện tại
                      </span>
                    )}
                  </div>
                  <p className="mt-2 text-xs leading-5 text-muted-foreground">
                    {(cv.sizeBytes / 1024).toLocaleString("vi-VN", { maximumFractionDigits: 1 })}{" "}
                    KiB · Bổ sung bởi {getPerson(cv.uploadedBy)?.name ?? "Người dùng"}
                    {cv.uploadedBy !== cv.employeeId && " (HR bổ sung thay)"}
                    <br />
                    {new Date(cv.uploadedAt).toLocaleString("vi-VN")}
                  </p>
                </div>
              </div>
              <div className="mt-3 flex flex-wrap items-center justify-between gap-3 border-t border-border pt-3">
                <span className="text-xs font-medium">
                  {cvProcessingLabels[cv.processingStatus]}{" "}
                  <span className="font-normal text-muted-foreground">· Mô phỏng</span>
                </span>
                {isOwner && isActive && !cv.isCurrent && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setError("")
                      setConfirmId(cv.id)
                    }}
                  >
                    Dùng làm CV hiện tại
                  </Button>
                )}
              </div>
              {cv.processingStatus === "FAILED" && (
                <p className="mt-2 text-xs leading-5 text-warning">
                  Không trích xuất được kỹ năng. Nhân viên vẫn có thể chọn bản CV này làm hiện tại.
                </p>
              )}
              {isActive && (
                <details className="mt-3 text-xs text-muted-foreground">
                  <summary className="cursor-pointer">Điều khiển prototype: trích xuất</summary>
                  <label className="mt-3 block" htmlFor={`processing-${cv.id}`}>
                    Trạng thái trích xuất mô phỏng
                  </label>
                  <select
                    id={`processing-${cv.id}`}
                    className="field mt-2 w-full"
                    value={cv.processingStatus}
                    onChange={(event) => {
                      try {
                        actions.simulateCvProcessing(
                          cv.id,
                          event.target.value as CvProcessingStatus,
                        )
                        setError("")
                      } catch (err) {
                        setError(err instanceof Error ? err.message : "Chưa thể đổi trạng thái.")
                      }
                    }}
                  >
                    {Object.entries(cvProcessingLabels).map(([id, label]) => (
                      <option key={id} value={id}>
                        {label}
                      </option>
                    ))}
                  </select>
                  <p className="mt-2 leading-5">
                    Chỉ minh họa giao diện, không chạy parser/AI hoặc tạo kỹ năng. Xét duyệt kỹ năng
                    từ CV sẽ được triển khai riêng.
                  </p>
                </details>
              )}
            </article>
          ))}
        </div>
      ) : (
        <div className="rounded-lg border border-dashed border-border p-5 text-sm text-muted-foreground">
          Chưa có CV.{" "}
          {isActive
            ? isOwner
              ? "Chọn Bổ sung CV để thêm phiên bản đầu tiên."
              : "Nhân viên có thể tự bổ sung; HR hỗ trợ bổ sung thay khi cần."
            : "Hồ sơ đã vô hiệu hóa; không thể bổ sung CV."}
        </div>
      )}
      <p className="mt-4 text-xs leading-5 text-muted-foreground">
        {versions.length} phiên bản · Chỉ thông tin mô phỏng trong phiên; chưa lưu tệp hoặc hỗ trợ
        tải xuống. Tải lại trang khôi phục dữ liệu mẫu.
      </p>
      {error && !selected && (
        <p role="alert" className="mt-3 text-sm text-warning">
          {error}
        </p>
      )}
      {uploading && (
        <CvUploadForm
          employeeName={getPerson(employeeId)?.name ?? employeeId}
          isOnBehalf={!isOwner}
          onSave={(file) => actions.addCv(employeeId, file)}
          onClose={() => setUploading(false)}
        />
      )}
      {selected && (
        <Dialog
          open
          onOpenChange={(open) => {
            if (!open) setConfirmId(null)
          }}
          title="Dùng bản này làm CV hiện tại?"
          description="Bản trước vẫn được giữ lại. Kỹ năng chính thức và các đề nghị đang chờ không thay đổi."
        >
          <p className="break-all text-sm font-medium">{selected.filename}</p>
          {error && (
            <p role="alert" className="mt-3 text-sm text-warning">
              {error}
            </p>
          )}
          <div className="mt-5 flex flex-wrap justify-end gap-2">
            <Button variant="outline" onClick={() => setConfirmId(null)}>
              Hủy
            </Button>
            <Button
              onClick={() => {
                try {
                  actions.selectCurrentCv(selected.id)
                  setConfirmId(null)
                  setError("")
                } catch (err) {
                  setError(err instanceof Error ? err.message : "Chưa thể đổi CV hiện tại.")
                }
              }}
            >
              Xác nhận sử dụng
            </Button>
          </div>
        </Dialog>
      )}
    </section>
  )
}
