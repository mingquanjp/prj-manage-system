import { useEffect, useRef, useState } from "react"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { AuthError } from "../session/auth-api"
import type { AuthSession } from "../session/auth-session"
import { createCatalogApi } from "../organization/catalog-api"
import { AccountFields, type AccountOptions } from "./AccountFields"
import { loadAccountOptions } from "./account-options"
import {
  accountDraftErrors,
  accountErrorMessage,
  createAccountApi,
  emptyAccountDraft,
  type AccountCreated,
  type AccountDraft,
} from "./account-api"

export function ServerAccountCreateForm({
  session,
  busy,
  onClose,
  onCreated,
  onReconcile,
}: {
  session: AuthSession
  busy: boolean
  onClose: () => void
  onCreated: (value: AccountCreated) => void
  onReconcile: (email: string) => void
}) {
  const [api] = useState(createAccountApi)
  const [catalogs] = useState(createCatalogApi)
  const [values, setValues] = useState<AccountDraft>(() => ({
    ...emptyAccountDraft,
    systemRoles: [],
  }))
  const [options, setOptions] = useState<AccountOptions>()
  const [preview, setPreview] = useState("")
  const [error, setError] = useState("")
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [blocked, setBlocked] = useState(false)
  const [discard, setDiscard] = useState(false)
  const [revision, setRevision] = useState(0)
  const active = useRef(true)
  const sending = useRef(false)
  useEffect(() => {
    active.current = true
    return () => {
      active.current = false
    }
  }, [])
  useEffect(() => {
    let current = true
    void Promise.all([api.preview(), loadAccountOptions(catalogs)])
      .then(([code, entries]) => {
        if (current) {
          setPreview(code)
          setOptions(entries)
          setLoading(false)
        }
      })
      .catch((issue: unknown) => {
        if (!current) return
        setError(accountErrorMessage(issue))
        setLoading(false)
        if (issue instanceof AuthError && [401, 403].includes(issue.status)) void session.refresh()
      })
    return () => {
      current = false
    }
  }, [api, catalogs, session, revision])
  const dirty = JSON.stringify(values) !== JSON.stringify(emptyAccountDraft)
  function close() {
    if (sending.current || busy) return
    if (dirty && !blocked) setDiscard(true)
    else onClose()
  }
  return (
    <>
      <Dialog
        open
        wide
        title="Tạo tài khoản kèm hồ sơ"
        description="Email công ty dùng để đăng nhập. Mật khẩu tạm được gửi qua email; nhân viên phải đổi ở lần đăng nhập đầu."
        onOpenChange={(open) => {
          if (!open) close()
        }}
        onCloseAutoFocus={(event) => {
          event.preventDefault()
          document.getElementById("account-create")?.focus()
        }}
      >
        <form
          className="space-y-5"
          noValidate
          onSubmit={async (event) => {
            event.preventDefault()
            if (sending.current || busy || blocked || !options || loading) return
            const validation = accountDraftErrors(values)
            setErrors(validation)
            if (Object.keys(validation).length) return
            sending.current = true
            setSaving(true)
            setError("")
            let posted = false
            try {
              const result = await session.catalogMutation((token) => {
                posted = true
                return api.create(values, token)
              })
              if (active.current) onCreated(result)
            } catch (issue) {
              if (!active.current) return
              setError(accountErrorMessage(issue, posted))
              if (issue instanceof AuthError && issue.code === "EMAIL_CONFLICT")
                setErrors({ companyEmail: "Email đã được sử dụng." })
              setBlocked(
                posted &&
                  (!(issue instanceof AuthError) || issue.status === 0 || issue.status >= 500),
              )
              if (issue instanceof AuthError && issue.status === 403) void session.refresh()
            } finally {
              sending.current = false
              if (active.current) setSaving(false)
            }
          }}
        >
          <label className="field-label block" htmlFor="account-preview">
            Mã nhân viên dự kiến
          </label>
          <input
            id="account-preview"
            className="field w-full bg-accent"
            readOnly
            value={preview}
            aria-describedby="account-preview-note"
          />
          <p id="account-preview-note" className="text-xs text-muted-foreground">
            Mã dự kiến không giữ chỗ. Mã sau khi lưu thành công mới là mã chính thức.
          </p>
          {loading && <p role="status">Đang tải danh mục và mã dự kiến…</p>}
          <fieldset disabled={saving || busy || blocked || loading} className="space-y-5">
            {options && (
              <AccountFields
                values={values}
                options={options}
                errors={errors}
                onChange={setValues}
              />
            )}
          </fieldset>
          {error && (
            <p role="alert" className="text-sm leading-6 text-warning">
              {error}
            </p>
          )}
          {blocked ? (
            <Button type="button" onClick={() => onReconcile(values.companyEmail)}>
              Đối soát trong danh sách
            </Button>
          ) : (
            <div className="flex flex-wrap justify-end gap-2 border-t border-border pt-4">
              <Button
                type="button"
                variant="outline"
                disabled={saving || busy || loading}
                onClick={() => {
                  setOptions(undefined)
                  setLoading(true)
                  setError("")
                  setRevision((n) => n + 1)
                }}
              >
                Tải lại lựa chọn
              </Button>
              <Button type="button" variant="outline" disabled={saving || busy} onClick={close}>
                Hủy
              </Button>
              <Button type="submit" disabled={saving || busy || loading || !options}>
                {saving ? "Đang tạo và gửi email…" : "Tạo tài khoản và hồ sơ"}
              </Button>
            </div>
          )}
        </form>
      </Dialog>
      <Dialog
        open={discard}
        onOpenChange={setDiscard}
        title="Bỏ thông tin chưa lưu?"
        description="Chưa gửi yêu cầu tạo tài khoản."
      >
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={() => setDiscard(false)}>
            Tiếp tục chỉnh sửa
          </Button>
          <Button onClick={onClose}>Bỏ thông tin</Button>
        </div>
      </Dialog>
    </>
  )
}
