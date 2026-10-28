import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import type { AccountCreateInput, AccountsState } from "@/models/accounts"
import type { ProfileState } from "@/models/profile"
import type { OrganizationState } from "@/models/organization"
import { accountCreateErrors, nextAccountIdentifiers } from "@/demo/accounts"
import { RoleChoices } from "./RoleChoices"

const emptyInput: AccountCreateInput = {
  fullname: "",
  companyEmail: "",
  positionId: "",
  departmentId: null,
  jobLevelId: null,
  phone: null,
  systemRoles: [],
}
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
  const [values, setValues] = useState<AccountCreateInput>({ ...emptyInput })
  const [error, setError] = useState("")
  const [showErrors, setShowErrors] = useState(false)
  const [discard, setDiscard] = useState(false)
  const dirty = JSON.stringify(values) !== JSON.stringify(emptyInput)
  const requestClose = () => (dirty ? setDiscard(true) : onClose())
  const identifiers = nextAccountIdentifiers(accounts, profiles)
  const errors = showErrors ? accountCreateErrors(profiles, organization, values) : {}
  return (
    <>
      <Dialog
        open
        wide
        onOpenChange={(open) => {
          if (!open) requestClose()
        }}
        title="Tạo tài khoản kèm hồ sơ"
        description="Tạo cùng lúc tài khoản và hồ sơ nhân viên ban đầu trong bản mô phỏng; HR quản lý hồ sơ về sau."
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
            } catch (err) {
              setError(err instanceof Error ? err.message : "Chưa thể tạo tài khoản.")
            }
          }}
        >
          <p className="rounded-md bg-accent p-3 text-xs leading-5 text-muted-foreground">
            Chưa cấp thông tin đăng nhập thật. Mật khẩu, thư mời và kênh cấp tài khoản sẽ triển khai
            cùng phần xác thực.
          </p>
          <div className="grid gap-4 sm:grid-cols-2">
            {(
              [
                ["employeeCode", "Mã nhân viên"],
                ["username", "Tên đăng nhập"],
              ] as const
            ).map(([field, label]) => (
              <div key={field}>
                <label className="field-label mb-2 block" htmlFor={`account-${field}`}>
                  {label}
                </label>
                <input
                  id={`account-${field}`}
                  className="field w-full bg-accent"
                  value={identifiers[field]}
                  readOnly
                  aria-describedby="account-identifiers-note"
                />
              </div>
            ))}
            <p
              id="account-identifiers-note"
              className="text-xs leading-5 text-muted-foreground sm:col-span-2"
            >
              Mã dự kiến được cấp khi lưu. Tên đăng nhập chỉ tạm dùng trong bản demo; cách đăng nhập
              sẽ hoàn thiện cùng backend.
            </p>
            {(
              [
                ["fullname", "Họ và tên", 100],
                ["companyEmail", "Email công ty", 100],
                ["phone", "Số điện thoại", 20],
              ] as const
            ).map(([field, label, limit]) => (
              <div key={field}>
                <label className="field-label mb-2 block" htmlFor={`account-${field}`}>
                  {label}
                  {field !== "phone" && " *"}
                </label>
                <input
                  id={`account-${field}`}
                  className="field w-full"
                  value={values[field] ?? ""}
                  maxLength={limit}
                  type={field === "companyEmail" ? "email" : field === "phone" ? "tel" : "text"}
                  autoComplete="off"
                  required={field !== "phone"}
                  aria-invalid={Boolean(errors[field])}
                  aria-describedby={errors[field] ? `account-error-${field}` : undefined}
                  onChange={(e) => setValues({ ...values, [field]: e.target.value })}
                />
                {errors[field] && (
                  <p
                    id={`account-error-${field}`}
                    role="alert"
                    className="mt-1 text-xs text-warning"
                  >
                    {errors[field]}
                  </p>
                )}
              </div>
            ))}
            {(
              [
                ["positionId", "positions", "Vị trí công việc", true],
                ["departmentId", "departments", "Phòng ban", false],
                ["jobLevelId", "jobLevels", "Cấp bậc", false],
              ] as const
            ).map(([field, kind, label, required]) => (
              <div key={field}>
                <label className="field-label mb-2 block" htmlFor={`account-${field}`}>
                  {label}
                  {required && " *"}
                </label>
                <select
                  id={`account-${field}`}
                  className="field w-full"
                  value={values[field] ?? ""}
                  required={required}
                  aria-invalid={Boolean(errors[field])}
                  aria-describedby={errors[field] ? `account-error-${field}` : undefined}
                  onChange={(e) => setValues({ ...values, [field]: e.target.value || null })}
                >
                  <option value="">{required ? "Chọn vị trí" : "Chưa cập nhật"}</option>
                  {organization[kind].map((entry) => (
                    <option key={entry.id} value={entry.id}>
                      {entry.name}
                    </option>
                  ))}
                </select>
                {errors[field] && (
                  <p
                    id={`account-error-${field}`}
                    role="alert"
                    className="mt-1 text-xs text-warning"
                  >
                    {errors[field]}
                  </p>
                )}
              </div>
            ))}
          </div>
          <p className="text-xs leading-5 text-muted-foreground">
            Tên đăng nhập và email công ty là hai thông tin riêng. Địa chỉ cần bổ sung khi hoàn
            thiện hồ sơ, không nhập tại bước này.
          </p>
          <RoleChoices
            value={values.systemRoles}
            onChange={(roles) => setValues({ ...values, systemRoles: roles })}
          />
          {error && (
            <p role="alert" className="text-sm text-warning">
              {error}
            </p>
          )}
          <div className="flex flex-wrap justify-end gap-2 border-t border-border pt-4">
            <Button variant="outline" onClick={requestClose}>
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
