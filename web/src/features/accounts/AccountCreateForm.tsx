import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import type { AccountCreateInput, AccountsState } from "@/models/accounts"
import type { ProfileState } from "@/models/profile"
import type { OrganizationState } from "@/models/organization"
import { accountCreateErrors, nextAccountIdentifiers } from "@/demo/accounts"
import { AccountFields } from "./AccountFields"
import { emptyAccountDraft, type AccountDraft } from "./account-api"

export function AccountCreateForm({
  accounts,
  profiles,
  organization,
  onSave,
  onClose,
}: {
  accounts: AccountsState
  profiles: ProfileState
  organization: OrganizationState
  onSave: (input: AccountCreateInput) => void
  onClose: () => void
}) {
  const [values, setValues] = useState<AccountDraft>(() => ({
    ...emptyAccountDraft,
    systemRoles: [],
  }))
  const [error, setError] = useState("")
  const [showErrors, setShowErrors] = useState(false)
  const [discard, setDiscard] = useState(false)
  const dirty = JSON.stringify(values) !== JSON.stringify(emptyAccountDraft)
  const close = () => (dirty ? setDiscard(true) : onClose())
  const preview = nextAccountIdentifiers(accounts, profiles).employeeCode
  const errors = showErrors ? accountCreateErrors(profiles, organization, values) : {}
  return (
    <>
      <Dialog
        open
        wide
        onOpenChange={(open) => {
          if (!open) close()
        }}
        title="Tạo tài khoản kèm hồ sơ"
        description="Bản mô phỏng trong bộ nhớ, chưa cấp thông tin đăng nhập hoặc gửi email thật."
      >
        <form
          className="space-y-5"
          noValidate
          onSubmit={(event) => {
            event.preventDefault()
            setShowErrors(true)
            setError("")
            if (Object.keys(accountCreateErrors(profiles, organization, values)).length) return
            try {
              onSave(values)
              onClose()
            } catch (issue) {
              setError(issue instanceof Error ? issue.message : "Chưa thể tạo tài khoản.")
            }
          }}
        >
          <label className="field-label block" htmlFor="demo-account-code">
            Mã nhân viên dự kiến
          </label>
          <input
            id="demo-account-code"
            className="field w-full bg-accent"
            readOnly
            value={preview}
          />
          <p className="text-xs text-muted-foreground">
            Mã demo tính trong bộ nhớ; mã chính thức do backend cấp. Email công ty là thông tin đăng
            nhập.
          </p>
          <AccountFields
            values={values}
            options={organization}
            errors={errors}
            onChange={setValues}
          />
          {error && (
            <p role="alert" className="text-sm text-warning">
              {error}
            </p>
          )}
          <div className="flex justify-end gap-2 border-t border-border pt-4">
            <Button variant="outline" onClick={close}>
              Hủy
            </Button>
            <Button type="submit">Tạo tài khoản và hồ sơ</Button>
          </div>
        </form>
      </Dialog>
      <Dialog
        open={discard}
        onOpenChange={setDiscard}
        title="Bỏ thông tin chưa lưu?"
        description="Chưa có tài khoản hoặc hồ sơ mới được tạo."
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
