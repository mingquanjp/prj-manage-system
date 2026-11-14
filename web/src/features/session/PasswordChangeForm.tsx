import { useState } from "react"
import { Button } from "@/components/ui/button"
import {
  passwordSize,
  validatePasswordChange,
  type PasswordFields,
  type PasswordErrors,
} from "./password-validation"

const emptyFields: PasswordFields = { currentPassword: "", newPassword: "", confirmation: "" }
const fieldDefinitions = [
  { key: "currentPassword", label: "Mật khẩu hiện tại", autoComplete: "current-password" },
  { key: "newPassword", label: "Mật khẩu mới", autoComplete: "new-password" },
  { key: "confirmation", label: "Nhập lại mật khẩu mới", autoComplete: "new-password" },
] as const

export function PasswordChangeForm({
  busy,
  waitSeconds,
  issueCode,
  onSubmit,
}: {
  busy: boolean
  waitSeconds: number
  issueCode?: string
  onSubmit: (currentPassword: string, newPassword: string) => void
}) {
  const [fields, setFields] = useState<PasswordFields>(emptyFields)
  const [errors, setErrors] = useState<PasswordErrors>({})
  const [edited, setEdited] = useState(false)
  const size = passwordSize(fields.newPassword)
  return (
    <form
      className="mt-6 space-y-4"
      noValidate
      onSubmit={(event) => {
        event.preventDefault()
        if (busy || waitSeconds > 0) return
        const nextErrors = validatePasswordChange(fields)
        setErrors(nextErrors)
        if (Object.keys(nextErrors).length) return
        const { currentPassword, newPassword } = fields
        // Sau khi chuyển input cho request, form không tiếp tục giữ ba mật khẩu.
        setFields(emptyFields)
        setEdited(false)
        onSubmit(currentPassword, newPassword)
      }}
    >
      <p id="password-policy" className="text-sm leading-6 text-muted-foreground">
        Mật khẩu mới cần ít nhất 15 ký tự Unicode, tối đa 72 byte UTF-8 và khác mật khẩu hiện tại.
        Bạn có thể dùng khoảng trắng, chữ có dấu và dán mật khẩu.
      </p>
      {fieldDefinitions.map(({ key, label, autoComplete }) => {
        const error =
          errors[key] ||
          (!edited && key === "currentPassword" && issueCode === "INVALID_CURRENT_PASSWORD"
            ? "Mật khẩu hiện tại không đúng."
            : "")
        return (
          <div key={key}>
            <label htmlFor={`change-${key}`} className="field-label">
              {label}
            </label>
            <input
              id={`change-${key}`}
              name={key}
              className="field mt-2 w-full"
              type="password"
              autoComplete={autoComplete}
              required
              disabled={busy}
              value={fields[key]}
              aria-invalid={Boolean(error)}
              aria-describedby={
                [
                  key === "newPassword" ? "password-policy password-size" : "",
                  error ? `error-${key}` : "",
                ]
                  .filter(Boolean)
                  .join(" ") || undefined
              }
              onChange={(event) => {
                setFields((previous) => ({ ...previous, [key]: event.target.value }))
                setErrors({})
                setEdited(true)
              }}
            />
            {error && (
              <p id={`error-${key}`} role="alert" className="mt-2 text-sm text-destructive">
                {error}
              </p>
            )}
          </div>
        )
      })}
      <p id="password-size" className="text-xs text-muted-foreground">
        Mật khẩu mới: {size.codePoints} ký tự Unicode · {size.bytes}/72 byte UTF-8
      </p>
      <Button type="submit" className="w-full" disabled={busy || waitSeconds > 0}>
        {busy ? "Đang xử lý…" : waitSeconds > 0 ? `Thử lại sau ${waitSeconds}s` : "Đổi mật khẩu"}
      </Button>
      <p className="text-xs leading-5 text-muted-foreground">
        Sau khi đổi, tất cả phiên của tài khoản sẽ kết thúc. Bạn cần đăng nhập lại bằng mật khẩu
        mới.
      </p>
    </form>
  )
}
