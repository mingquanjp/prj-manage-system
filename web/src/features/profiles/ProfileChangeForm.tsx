import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import type { ProfileValues } from "@/models/profile"
import { profileFields, profileFieldLabels } from "@/models/profile"
import { profilePatch, profilePatchErrors } from "@/demo/profiles"

export function ProfileChangeForm({
  before: initialBefore,
  onSubmit,
  onClose,
}: {
  before: ProfileValues
  onSubmit: (before: ProfileValues, proposed: ProfileValues, explanation: string) => void
  onClose: () => void
}) {
  const [before] = useState<ProfileValues>(() => ({ ...initialBefore }))
  const [values, setValues] = useState<ProfileValues>({ ...before })
  const [explanation, setExplanation] = useState("")
  const [error, setError] = useState("")
  const [showErrors, setShowErrors] = useState(false)
  const [discard, setDiscard] = useState(false)
  const patch = profilePatch(before, values)
  const errors = showErrors ? profilePatchErrors(patch) : {}
  const dirty =
    profileFields.some((field) => (values[field] ?? "") !== (before[field] ?? "")) ||
    Boolean(explanation)
  const requestClose = () => (dirty ? setDiscard(true) : onClose())
  return (
    <>
      <Dialog
        open
        onOpenChange={(open) => {
          if (!open) requestClose()
        }}
        title="Đề nghị cập nhật thông tin"
        description="Chỉ thông tin thay đổi được gửi tới HR. Hồ sơ chính thức giữ nguyên trong thời gian chờ xử lý."
      >
        <form
          className="space-y-5"
          onSubmit={(event) => {
            event.preventDefault()
            setShowErrors(true)
            setError("")
            if (Object.keys(profilePatchErrors(patch)).length) return
            try {
              onSubmit(before, values, explanation)
              onClose()
            } catch (err) {
              setError(err instanceof Error ? err.message : "Chưa thể gửi yêu cầu.")
            }
          }}
        >
          <p className="text-sm leading-6 text-muted-foreground">
            Bạn có thể đề nghị sửa họ tên, số điện thoại và địa chỉ. Email công ty và thông tin công
            tác do HR quản lý.
          </p>
          {profileFields.map((field) => (
            <div key={field}>
              <label className="field-label mb-2 block" htmlFor={`profile-${field}`}>
                {profileFieldLabels[field]}
              </label>
              {field === "address" ? (
                <textarea
                  id={`profile-${field}`}
                  className="field workspace-field min-h-20 w-full"
                  value={values[field] ?? ""}
                  onChange={(event) => setValues({ ...values, [field]: event.target.value })}
                  aria-invalid={Boolean(errors[field])}
                  aria-describedby={
                    errors[field] ? `profile-${field}-error` : "profile-address-hint"
                  }
                />
              ) : (
                <input
                  id={`profile-${field}`}
                  className="field workspace-field w-full"
                  type={field === "phone" ? "tel" : "text"}
                  maxLength={field === "fullname" ? 100 : 20}
                  value={values[field] ?? ""}
                  onChange={(event) => setValues({ ...values, [field]: event.target.value })}
                  aria-invalid={Boolean(errors[field])}
                  aria-describedby={errors[field] ? `profile-${field}-error` : undefined}
                />
              )}
              {errors[field] && (
                <p id={`profile-${field}-error`} role="alert" className="mt-2 text-sm text-warning">
                  {errors[field]}
                </p>
              )}
              {field === "address" && (
                <p
                  id="profile-address-hint"
                  className="mt-2 text-xs leading-5 text-muted-foreground"
                >
                  Cần bổ sung khi hoàn thiện hồ sơ. Nếu chỉ sửa họ tên/SĐT, bạn có thể giữ nguyên
                  địa chỉ chưa cập nhật. Địa chỉ đã có không được xóa trắng.
                </p>
              )}
            </div>
          ))}
          <div>
            <label className="field-label mb-2 block" htmlFor="profile-explanation">
              Ghi chú cho HR <span className="font-normal text-muted-foreground">(tùy chọn)</span>
            </label>
            <textarea
              id="profile-explanation"
              className="field workspace-field min-h-20 w-full"
              value={explanation}
              onChange={(event) => setExplanation(event.target.value)}
            />
          </div>
          {error && (
            <p role="alert" className="text-sm text-warning">
              {error}
            </p>
          )}
          <div className="flex flex-wrap justify-end gap-2 border-t border-border pt-4">
            <Button variant="outline" onClick={requestClose}>
              Hủy
            </Button>
            <Button type="submit" disabled={!Object.keys(patch).length}>
              Gửi đề nghị
            </Button>
          </div>
        </form>
      </Dialog>
      <Dialog
        open={discard}
        onOpenChange={setDiscard}
        title="Bỏ nội dung chưa gửi?"
        description="Hồ sơ chính thức không bị thay đổi."
      >
        <div className="flex flex-wrap justify-end gap-2">
          <Button variant="outline" onClick={() => setDiscard(false)}>
            Tiếp tục chỉnh sửa
          </Button>
          <Button onClick={onClose}>Bỏ nội dung</Button>
        </div>
      </Dialog>
    </>
  )
}
