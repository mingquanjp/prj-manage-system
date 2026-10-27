import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import type { EmployeeProfile, EmployeeEditValues } from "@/models/profile"
import { employeeEditFields, employeeEditLabels } from "@/models/profile"
import type { OrganizationState } from "@/models/organization"
import {
  employeeEditErrors,
  employeeEditValues,
  normalizeEmployeeEdit,
} from "@/demo/employee-administration"

export function EmployeeEditForm({
  profile,
  profiles,
  organization,
  onSave,
  onClose,
}: {
  profile: EmployeeProfile
  profiles: EmployeeProfile[]
  organization: OrganizationState
  onSave: (before: EmployeeProfile, values: EmployeeEditValues) => void
  onClose: () => void
}) {
  const [before] = useState(() => ({ ...profile }))
  const [values, setValues] = useState(() => employeeEditValues(before))
  const [error, setError] = useState("")
  const [showErrors, setShowErrors] = useState(false)
  const [discard, setDiscard] = useState(false)
  const normalized = normalizeEmployeeEdit(values)
  const dirty = employeeEditFields.some((field) => (values[field] ?? "") !== (before[field] ?? ""))
  const changed = employeeEditFields.some((field) => normalized[field] !== before[field])
  const errors = showErrors ? employeeEditErrors(profiles, organization, before, values) : {}
  const close = () => (dirty ? setDiscard(true) : onClose())
  return (
    <>
      <Dialog
        open
        wide
        onOpenChange={(open) => {
          if (!open) close()
        }}
        title="Chỉnh sửa hồ sơ nhân sự"
        description={`${before.employeeCode} · ${before.fullname}. HR lưu trực tiếp vào hồ sơ chính thức.`}
      >
        <form
          noValidate
          onSubmit={(event) => {
            event.preventDefault()
            setShowErrors(true)
            setError("")
            if (Object.keys(employeeEditErrors(profiles, organization, before, values)).length)
              return
            try {
              onSave(before, values)
              onClose()
            } catch (err) {
              setError(err instanceof Error ? err.message : "Chưa thể lưu hồ sơ.")
            }
          }}
        >
          <div className="grid gap-4 sm:grid-cols-2">
            {employeeEditFields.map((field) => {
              const required =
                field === "fullname" || field === "companyEmail" || field === "positionId"
              const id = `employee-${field}`
              const common = {
                id,
                className: "field workspace-field w-full",
                value: values[field] ?? "",
                "aria-required": required,
                "aria-invalid": Boolean(errors[field]),
                "aria-describedby": errors[field]
                  ? `${id}-error`
                  : field === "address"
                    ? "employee-address-hint"
                    : undefined,
                onChange: (
                  event: React.ChangeEvent<
                    HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement
                  >,
                ) => {
                  setValues({ ...values, [field]: event.target.value })
                  setError("")
                },
              }
              return (
                <div key={field} className={field === "address" ? "sm:col-span-2" : ""}>
                  <label className="field-label mb-2 block" htmlFor={id}>
                    {employeeEditLabels[field]}
                    {required && <span aria-hidden="true"> *</span>}
                  </label>
                  {field === "departmentId" || field === "positionId" || field === "jobLevelId" ? (
                    <select {...common}>
                      <option value="">
                        {field === "departmentId"
                          ? "Chưa thuộc phòng ban"
                          : field === "jobLevelId"
                            ? "Chưa cập nhật"
                            : "Chọn vị trí công việc"}
                      </option>
                      {organization[
                        field === "departmentId"
                          ? "departments"
                          : field === "positionId"
                            ? "positions"
                            : "jobLevels"
                      ].map((entry) => (
                        <option key={entry.id} value={entry.id}>
                          {entry.name}
                        </option>
                      ))}
                    </select>
                  ) : field === "address" ? (
                    <textarea {...common} rows={2} />
                  ) : (
                    <input
                      {...common}
                      type={field === "companyEmail" ? "email" : field === "phone" ? "tel" : "text"}
                      maxLength={field === "phone" ? 20 : 100}
                    />
                  )}
                  {errors[field] && (
                    <p role="alert" id={`${id}-error`} className="mt-1 text-sm text-warning">
                      {errors[field]}
                    </p>
                  )}
                  {field === "address" && (
                    <p id="employee-address-hint" className="mt-1 text-xs text-muted-foreground">
                      Có thể bổ sung sau nếu chưa có. Địa chỉ đã có không được xóa trắng.
                    </p>
                  )}
                </div>
              )
            })}
          </div>
          {error && (
            <p role="alert" className="mt-4 text-sm text-warning">
              {error}
            </p>
          )}
          <div className="mt-5 flex justify-end gap-2 border-t border-border pt-4">
            <Button variant="outline" onClick={close}>
              Hủy
            </Button>
            <Button type="submit" disabled={!changed}>
              Lưu thay đổi
            </Button>
          </div>
        </form>
      </Dialog>
      <Dialog
        open={discard}
        onOpenChange={setDiscard}
        title="Bỏ thay đổi chưa lưu?"
        description="Hồ sơ chính thức sẽ giữ nguyên."
      >
        <div className="flex flex-wrap justify-end gap-2">
          <Button variant="outline" onClick={() => setDiscard(false)}>
            Tiếp tục chỉnh sửa
          </Button>
          <Button onClick={onClose}>Bỏ thay đổi</Button>
        </div>
      </Dialog>
    </>
  )
}
