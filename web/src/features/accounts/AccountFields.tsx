import type { AccountDraft } from "./account-api"
import { RoleChoices } from "./RoleChoices"

export type AccountOptions = Record<
  "positions" | "departments" | "jobLevels",
  { id: string; name: string }[]
>
// Trường trình bày dùng chung; dữ liệu và validation do adapter demo/server cung cấp.
export function AccountFields({
  values,
  options,
  errors,
  onChange,
}: {
  values: AccountDraft
  options: AccountOptions
  errors: Record<string, string>
  onChange: (value: AccountDraft) => void
}) {
  return (
    <>
      <div className="grid gap-4 sm:grid-cols-2">
        {(
          [
            ["fullname", "Họ và tên", "text"],
            ["companyEmail", "Email công ty", "email"],
            ["phone", "Số điện thoại", "tel"],
            ["address", "Địa chỉ", "text"],
          ] as const
        ).map(([field, label, type]) => (
          <div key={field}>
            <label className="field-label mb-2 block" htmlFor={`account-${field}`}>
              {label}
              {(field === "fullname" || field === "companyEmail") && " *"}
            </label>
            <input
              id={`account-${field}`}
              className="field w-full"
              type={type}
              autoComplete="off"
              value={values[field] ?? ""}
              aria-invalid={Boolean(errors[field])}
              aria-describedby={errors[field] ? `account-error-${field}` : undefined}
              onChange={(e) =>
                onChange({
                  ...values,
                  [field]:
                    (field === "phone" || field === "address") && e.target.value === ""
                      ? null
                      : e.target.value,
                })
              }
            />
            {errors[field] && (
              <p id={`account-error-${field}`} role="alert" className="mt-1 text-xs text-warning">
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
              aria-invalid={Boolean(errors[field])}
              aria-describedby={errors[field] ? `account-error-${field}` : undefined}
              onChange={(e) =>
                onChange({ ...values, [field]: e.target.value || (required ? "" : null) })
              }
            >
              <option value="">{required ? "Chọn vị trí" : "Chưa cập nhật"}</option>
              {options[kind].map((entry) => (
                <option key={entry.id} value={entry.id}>
                  {entry.name}
                </option>
              ))}
            </select>
            {errors[field] && (
              <p id={`account-error-${field}`} role="alert" className="mt-1 text-xs text-warning">
                {errors[field]}
              </p>
            )}
          </div>
        ))}
      </div>
      <RoleChoices
        value={values.systemRoles}
        onChange={(systemRoles) => onChange({ ...values, systemRoles })}
      />
      {errors.systemRoles && (
        <p role="alert" className="text-sm text-warning">
          {errors.systemRoles}
        </p>
      )}
    </>
  )
}
