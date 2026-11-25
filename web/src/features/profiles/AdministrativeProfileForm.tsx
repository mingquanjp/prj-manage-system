import { useState } from "react"
import { Button } from "@/components/ui/button"
import type { CatalogItem, CatalogKind } from "../organization/catalog-api"
import {
  administrativeErrors,
  administrativePatch,
  administrativeValues,
  type AdministrativeField,
  type AdministrativePatch,
  type AdministrativeProfile,
} from "./administrative-profile-api"
export const administrativeLabels: Record<AdministrativeField, string> = {
  fullname: "Họ và tên",
  company_email: "Email công ty",
  phone: "Số điện thoại",
  address: "Địa chỉ",
  position_id: "Vị trí công việc",
  department_id: "Phòng ban",
  job_level_id: "Cấp bậc",
}
export function AdministrativeProfileForm({
  profile,
  catalogs,
  locked,
  fieldError,
  onSave,
  onCancel,
}: {
  profile: AdministrativeProfile
  catalogs: Record<CatalogKind, CatalogItem[]>
  locked: boolean
  fieldError?: string
  onSave: (patch: AdministrativePatch) => Promise<boolean>
  onCancel: () => void
}) {
  const [before] = useState(() => administrativeValues(profile))
  const [draft, setDraft] = useState(before)
  const [errors, setErrors] = useState<Partial<Record<AdministrativeField, string>>>({})
  const patch = administrativePatch(before, draft)
  const changed = Object.keys(patch.changes).length > 0
  return (
    <form
      className="space-y-4"
      onSubmit={(event) => {
        event.preventDefault()
        if (locked || !changed) return
        const next = administrativeErrors(patch)
        setErrors(next)
        if (!Object.keys(next).length) void onSave(patch)
      }}
    >
      <p className="text-sm text-muted-foreground">
        Mã nhân viên và trạng thái chỉ đọc. Sửa email thực sự sẽ thu hồi các phiên của nhân viên, kể
        cả phiên HR nếu đang sửa chính mình.
      </p>
      {(["fullname", "company_email", "phone", "address"] as const).map((key) => (
        <label className="block space-y-1 text-sm" key={key}>
          <span className="font-medium">{administrativeLabels[key]}</span>
          {key === "address" ? (
            <textarea
              rows={3}
              className="field workspace-field w-full"
              value={String(draft[key] ?? "")}
              disabled={locked}
              onChange={(e) => setDraft({ ...draft, [key]: e.target.value })}
              aria-invalid={!!errors[key]}
              aria-describedby={`admin-${key}-error`}
            />
          ) : (
            <input
              className="field workspace-field w-full"
              type="text"
              autoComplete="off"
              value={String(draft[key] ?? "")}
              disabled={locked}
              onChange={(e) =>
                setDraft({
                  ...draft,
                  [key]: key === "phone" && e.target.value === "" ? null : e.target.value,
                })
              }
              aria-invalid={!!errors[key] || fieldError === key}
              aria-describedby={`admin-${key}-error`}
            />
          )}
          <span id={`admin-${key}-error`} className="block text-xs text-destructive">
            {errors[key] || (fieldError === key ? "Email đã được sử dụng." : "")}
          </span>
          {key === "address" && (
            <span className="block text-xs text-muted-foreground">
              Nếu chưa có địa chỉ, bạn có thể bổ sung sau. Khi sửa cần 1–2000 ký tự, không xóa
              trắng.
            </span>
          )}
          {key === "phone" && (
            <span className="block text-xs text-muted-foreground">
              Để trống để xóa số điện thoại.
            </span>
          )}
        </label>
      ))}
      {(
        [
          ["position_id", "positions"],
          ["department_id", "departments"],
          ["job_level_id", "job-levels"],
        ] as const
      ).map(([key, kind]) => (
        <label className="block space-y-1 text-sm" key={key}>
          <span className="font-medium">{administrativeLabels[key]}</span>
          <select
            className="field workspace-field w-full"
            disabled={locked}
            value={draft[key] ?? ""}
            onChange={(e) =>
              setDraft({ ...draft, [key]: e.target.value ? Number(e.target.value) : null })
            }
            aria-invalid={!!errors[key]}
            aria-describedby={`admin-${key}-error`}
          >
            <option value="">
              {key === "position_id" ? "Chọn vị trí (bắt buộc)" : "Chưa cập nhật"}
            </option>
            {catalogs[kind].map((item) => (
              <option key={item.id} value={item.id}>
                {item.name}
              </option>
            ))}
          </select>
          <span id={`admin-${key}-error`} className="text-xs text-destructive">
            {errors[key]}
          </span>
        </label>
      ))}
      <p className="text-xs text-muted-foreground">
        Chỉ lưu những trường thay đổi. Giữ nguyên mã nhân viên, trạng thái và quyền của tài khoản.
      </p>
      {!changed && (
        <p role="status" className="text-sm text-muted-foreground">
          Chưa có thay đổi để lưu.
        </p>
      )}
      <div className="flex flex-wrap gap-3">
        <Button disabled={locked || !changed} type="submit">
          Lưu thay đổi
        </Button>
        <Button type="button" variant="outline" disabled={locked} onClick={onCancel}>
          Hủy chỉnh sửa
        </Button>
      </div>
    </form>
  )
}
