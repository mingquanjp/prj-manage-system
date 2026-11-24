import { useRef, useState } from "react"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { profileFields, profileFieldLabels } from "@/models/profile"
import { buildProfileRequest, type ProfileValues } from "./profile-request-input"

export function ServerProfileChangeForm({
  before: initialBefore,
  busy,
  blocked = false,
  onSubmit,
  onClose,
}: {
  before: ProfileValues
  busy: boolean
  blocked?: boolean
  onSubmit: (values: ProfileValues, explanation: string) => Promise<void>
  onClose: () => void
}) {
  // Giữ nền lúc mở form để refresh nền không thay bản nháp hoặc che stale.
  const [before] = useState(() => ({ ...initialBefore }))
  const [values, setValues] = useState<ProfileValues>(() => ({ ...initialBefore }))
  const [explanation, setExplanation] = useState("")
  const [error, setError] = useState("")
  const [discard, setDiscard] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const sending = useRef(false)
  const errorRef = useRef<HTMLParagraphElement>(null)
  const disabled = busy || submitting || blocked
  const changed = profileFields.some((field) => (values[field] ?? "") !== (before[field] ?? ""))
  const dirty = changed || Boolean(explanation)
  const requestClose = () => {
    if (sending.current || busy) return
    if (dirty) setDiscard(true)
    else onClose()
  }
  async function submit() {
    if (sending.current || busy || blocked) return
    setError("")
    try {
      buildProfileRequest(before, values, explanation)
    } catch (issue) {
      setError(issue instanceof Error ? issue.message : "Hãy kiểm tra nội dung đề nghị.")
      requestAnimationFrame(() => errorRef.current?.focus())
      return
    }
    sending.current = true
    setSubmitting(true)
    try {
      await onSubmit(values, explanation)
    } catch (issue) {
      setError(issue instanceof Error ? issue.message : "Chưa thể gửi đề nghị.")
      requestAnimationFrame(() => errorRef.current?.focus())
    } finally {
      sending.current = false
      setSubmitting(false)
    }
  }
  return (
    <>
      <Dialog
        open
        onOpenChange={(open) => {
          if (!open) requestClose()
        }}
        title="Đề nghị cập nhật thông tin"
        description="Chỉ thông tin thay đổi được gửi tới HR. Hồ sơ chính thức giữ nguyên trong thời gian chờ xử lý."
        onCloseAutoFocus={(event) => {
          event.preventDefault()
          const trigger = document.getElementById(
            "profile-request-trigger",
          ) as HTMLButtonElement | null
          if (trigger && !trigger.disabled) trigger.focus()
          else document.getElementById("profile-heading")?.focus()
        }}
      >
        <form
          className="space-y-5"
          aria-busy={submitting}
          onSubmit={(event) => {
            event.preventDefault()
            void submit()
          }}
        >
          <p className="text-sm leading-6 text-muted-foreground">
            Bạn có thể đề nghị sửa họ tên, số điện thoại và địa chỉ. Email công ty và thông tin công
            tác do HR quản lý.
          </p>
          {profileFields.map((field) => (
            <div key={field}>
              <label className="field-label mb-2 block" htmlFor={`server-profile-${field}`}>
                {profileFieldLabels[field]}
              </label>
              {field === "address" ? (
                <textarea
                  id={`server-profile-${field}`}
                  className="field workspace-field min-h-20 w-full"
                  value={values.address ?? ""}
                  disabled={disabled}
                  onChange={(event) => setValues({ ...values, address: event.target.value })}
                  aria-describedby="server-profile-address-hint"
                />
              ) : (
                <input
                  id={`server-profile-${field}`}
                  className="field workspace-field w-full"
                  type={field === "phone" ? "tel" : "text"}
                  value={values[field] ?? ""}
                  disabled={disabled}
                  onChange={(event) =>
                    setValues({
                      ...values,
                      [field]: field === "phone" && !event.target.value ? null : event.target.value,
                    })
                  }
                  aria-describedby={`server-profile-${field}-hint`}
                />
              )}
              <p
                id={`server-profile-${field}-hint`}
                className="mt-2 text-xs leading-5 text-muted-foreground"
              >
                {field === "address"
                  ? "Tối đa 2000 ký tự Unicode. Có thể giữ nguyên địa chỉ chưa cập nhật khi sửa trường khác; không gửi địa chỉ trắng hoặc xóa địa chỉ đã có."
                  : field === "phone"
                    ? "Tối đa 20 ký tự Unicode; để trống để xóa số điện thoại."
                    : "Từ 1 đến 100 ký tự Unicode sau khi bỏ khoảng trắng đầu/cuối."}
              </p>
            </div>
          ))}
          <div>
            <label className="field-label mb-2 block" htmlFor="server-profile-explanation">
              Ghi chú cho HR <span className="font-normal text-muted-foreground">(tùy chọn)</span>
            </label>
            <textarea
              id="server-profile-explanation"
              className="field workspace-field min-h-20 w-full"
              value={explanation}
              disabled={disabled}
              onChange={(event) => setExplanation(event.target.value)}
              aria-describedby="server-profile-explanation-hint"
            />
            <p
              id="server-profile-explanation-hint"
              className="mt-2 text-xs leading-5 text-muted-foreground"
            >
              Tối đa 2000 ký tự Unicode sau khi bỏ khoảng trắng đầu/cuối.
            </p>
          </div>
          {error && (
            <p ref={errorRef} tabIndex={-1} role="alert" className="text-sm text-warning">
              {error}
            </p>
          )}
          <div className="flex flex-wrap justify-end gap-2 border-t border-border pt-4">
            <Button variant="outline" disabled={busy || submitting} onClick={requestClose}>
              Hủy
            </Button>
            <Button type="submit" disabled={disabled || !changed}>
              {submitting ? "Đang gửi…" : "Gửi đề nghị"}
            </Button>
          </div>
        </form>
      </Dialog>
      <Dialog
        open={discard}
        onOpenChange={setDiscard}
        onCloseAutoFocus={(event) => {
          event.preventDefault()
          const field = document.getElementById(
            "server-profile-fullname",
          ) as HTMLInputElement | null
          const trigger = document.getElementById(
            "profile-request-trigger",
          ) as HTMLButtonElement | null
          if (field && !field.disabled) field.focus()
          else if (trigger && !trigger.disabled) trigger.focus()
          else document.getElementById("profile-heading")?.focus()
        }}
        title={blocked ? "Đóng nội dung đề nghị?" : "Bỏ nội dung chưa gửi?"}
        description={
          blocked
            ? "Đóng không gửi thêm đề nghị. Hãy tải lại hồ sơ và lịch sử để đối soát trước khi mở bản nháp mới."
            : "Đóng bản nháp không gửi thêm đề nghị và không thay đổi hồ sơ chính thức."
        }
      >
        <div className="flex flex-wrap justify-end gap-2">
          <Button variant="outline" onClick={() => setDiscard(false)}>
            Tiếp tục chỉnh sửa
          </Button>
          <Button onClick={onClose}>{blocked ? "Đóng và đối soát" : "Bỏ nội dung"}</Button>
        </div>
      </Dialog>
    </>
  )
}
