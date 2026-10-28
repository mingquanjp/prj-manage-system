import type { SystemRole } from "@/models/workspace"
import { systemRoleLabels } from "@/models/accounts"

export function RoleChoices({
  value,
  onChange,
}: {
  value: SystemRole[]
  onChange: (roles: SystemRole[]) => void
}) {
  return (
    <fieldset>
      <legend className="field-label mb-2">Quyền hệ thống</legend>
      <div className="flex flex-wrap gap-4">
        {(Object.keys(systemRoleLabels) as SystemRole[]).map((role) => (
          <label key={role} className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={value.includes(role)}
              onChange={(e) =>
                onChange(e.target.checked ? [...value, role] : value.filter((r) => r !== role))
              }
            />
            {systemRoleLabels[role]}
          </label>
        ))}
      </div>
      <p className="mt-2 text-xs leading-5 text-muted-foreground">
        Không chọn quyền bổ sung nào vẫn là nhân viên. PM/PTCM được chỉ định riêng trong từng dự án.
      </p>
    </fieldset>
  )
}
