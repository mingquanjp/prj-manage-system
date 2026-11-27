import { useRef, useState } from "react"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { profileFields, profileFieldLabels } from "@/models/profile"
import {
  buildProfileRequest,
  type ProfileRequestInput,
  type ProfileValues,
} from "./profile-request-input"

const fieldHints = {
  fullname: "Từ 1 đến 100 ký tự Unicode sau khi bỏ khoảng trắng đầu/cuối.",
  phone: "Tối đa 20 ký tự Unicode; để trống để xóa số điện thoại.",
  address:
    "Tối đa 2000 ký tự Unicode. Có thể giữ nguyên địa chỉ chưa cập nhật khi sửa trường khác; không gửi địa chỉ trắng hoặc xóa địa chỉ đã có.",
}

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
  onSubmit: (input: ProfileRequestInput) => Promise<void>
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
  let submitLabel = "Gửi đề nghị"
  if (submitting) submitLabel = "Đang gửi…"
  let discardTitle = "Bỏ nội dung chưa gửi?"
  let discardDescription =
    "Đóng bản nháp không gửi thêm đề nghị và không thay đổi hồ sơ chính thức."
  let discardLabel = "Bỏ nội dung"
  if (blocked) {
    discardTitle = "Đóng nội dung đề nghị?"
    discardDescription =
      "Đóng không gửi thêm đề nghị. Hãy tải lại hồ sơ và lịch sử để đối soát trước khi mở bản nháp mới."
    discardLabel = "Đóng và đối soát"
  }
  const requestClose = () => {
    if (sending.current || busy) return
    if (dirty) setDiscard(true)
    else onClose()
  }
  async function submit() {
    if (sending.current || busy || blocked) return
    setError("")
    let input: ProfileRequestInput
    try {
      input = buildProfileRequest(before, values, explanation)
    } catch (issue) {
      let message = "Hãy kiểm tra nội dung đề nghị."
      if (issue instanceof Error) message = issue.message
      setError(message)
      requestAnimationFrame(() => errorRef.current?.focus())
      return
    }
    sending.current = true
    setSubmitting(true)
    try {
      await onSubmit(input)
    } catch (issue) {
      let message = "Chưa thể gửi đề nghị."
      if (issue instanceof Error) message = issue.message
      setError(message)
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
          {profileFields.map((field) => {
            let control
            if (field === "address") {
              control = (
                <textarea
                  id={`server-profile-${field}`}
                  className="field workspace-field min-h-20 w-full"
                  value={values.address ?? ""}
                  disabled={disabled}
                  onChange={(event) => setValues({ ...values, address: event.target.value })}
                  aria-describedby="server-profile-address-hint"
                />
              )
            } else {
              let inputType = "text"
              if (field === "phone") inputType = "tel"
              control = (
                <input
                  id={`server-profile-${field}`}
                  className="field workspace-field w-full"
                  type={inputType}
                  value={values[field] ?? ""}
                  disabled={disabled}
                  onChange={(event) => {
                    let value: string | null = event.target.value
                    if (field === "phone" && !value) value = null
                    setValues({ ...values, [field]: value })
                  }}
                  aria-describedby={`server-profile-${field}-hint`}
                />
              )
            }
            return (
              <div key={field}>
                <label className="field-label mb-2 block" htmlFor={`server-profile-${field}`}>
                  {profileFieldLabels[field]}
                </label>
                {control}
                <p
                  id={`server-profile-${field}-hint`}
                  className="mt-2 text-xs leading-5 text-muted-foreground"
                >
                  {fieldHints[field]}
                </p>
              </div>
            )
          })}
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
              {submitLabel}
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
        title={discardTitle}
        description={discardDescription}
      >
        <div className="flex flex-wrap justify-end gap-2">
          <Button variant="outline" onClick={() => setDiscard(false)}>
            Tiếp tục chỉnh sửa
          </Button>
          <Button onClick={onClose}>{discardLabel}</Button>
        </div>
      </Dialog>
    </>
  )
}
