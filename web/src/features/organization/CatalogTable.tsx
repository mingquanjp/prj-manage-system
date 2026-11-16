import { Pencil1Icon, TrashIcon } from "@radix-ui/react-icons"
import { Button } from "@/components/ui/button"

// Bảng trình bày dùng chung; số tham chiếu chỉ có khi adapter thực sự cung cấp.
export function CatalogTable<T extends { id: string | number; name: string }>({
  entries,
  label,
  onEdit,
  onRemove,
  referenceCount,
  disabled = false,
}: {
  entries: T[]
  label: string
  onEdit?: (entry: T) => void
  onRemove?: (entry: T) => void
  referenceCount?: (entry: T) => number
  disabled?: boolean
}) {
  return (
    <table className="w-full min-w-[460px] text-left text-sm">
      <thead className="table-heading">
        <tr>
          <th scope="col" className="px-5 py-3">
            {label}
          </th>
          {referenceCount && (
            <th scope="col" className="px-5 py-3 text-right">
              Hồ sơ sử dụng
            </th>
          )}
          {(onEdit || onRemove) && (
            <th scope="col" className="w-28 px-5 py-3 text-right">
              Thao tác
            </th>
          )}
        </tr>
      </thead>
      <tbody>
        {entries.map((entry) => (
          <tr key={entry.id} className="border-t border-border">
            <th
              scope="row"
              className="max-w-md whitespace-pre-wrap break-words px-5 py-3 font-medium"
            >
              {entry.name}
            </th>
            {referenceCount && (
              <td className="px-5 py-3 text-right tabular-nums">{referenceCount(entry)}</td>
            )}
            {(onEdit || onRemove) && (
              <td className="px-4 py-2">
                <div className="flex justify-end gap-1">
                  {onEdit && (
                    <Button
                      variant="ghost"
                      size="icon"
                      disabled={disabled}
                      title={`Sửa ${entry.name}`}
                      aria-label={`Sửa ${entry.name}`}
                      onClick={() => onEdit(entry)}
                    >
                      <Pencil1Icon />
                    </Button>
                  )}
                  {onRemove && (
                    <Button
                      variant="ghost"
                      size="icon"
                      disabled={disabled}
                      className="text-warning"
                      title={`Xóa ${entry.name}`}
                      aria-label={`Xóa ${entry.name}`}
                      onClick={() => onRemove(entry)}
                    >
                      <TrashIcon />
                    </Button>
                  )}
                </div>
              </td>
            )}
          </tr>
        ))}
      </tbody>
    </table>
  )
}
