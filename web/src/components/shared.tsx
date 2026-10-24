import { MagnifyingGlassIcon } from "@radix-ui/react-icons"
import { cn } from "@/lib/utils"
import { statusLabels } from "@/models/workspace-labels"
import type { TaskStatus } from "@/models/workspace"

export function Avatar({ initials, small = false }: { initials: string; small?: boolean }) {
  return (
    <span
      aria-hidden="true"
      className={cn(
        "inline-flex shrink-0 items-center justify-center rounded-full bg-accent font-medium text-muted-foreground",
        small ? "size-7 text-xs" : "size-9 text-sm",
      )}
    >
      {initials}
    </span>
  )
}
export function StatusBadge({ status }: { status: TaskStatus | "PLANNING" | "COMPLETED" }) {
  const label =
    status === "PLANNING"
      ? "Lập kế hoạch"
      : status === "COMPLETED"
        ? "Đã hoàn thành"
        : statusLabels[status]
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-md px-2 py-1 text-sm whitespace-nowrap",
        status === "DONE"
          ? "bg-success-soft text-success"
          : status === "IN_PROGRESS" || status === "PENDING_REVIEW"
            ? "bg-warning-soft text-warning"
            : "bg-accent text-muted-foreground",
      )}
    >
      <span className="size-1.5 rounded-full bg-current" />
      {label}
    </span>
  )
}
export function SearchInput({
  id,
  value,
  onChange,
  placeholder,
}: {
  id?: string
  value: string
  onChange: (value: string) => void
  placeholder: string
}) {
  return (
    <div className="relative min-w-0">
      <MagnifyingGlassIcon className="pointer-events-none absolute left-3 top-3 size-4 text-muted-foreground" />
      <input
        id={id}
        type="search"
        aria-label={placeholder}
        className="field w-full pl-9"
        placeholder={placeholder}
        value={value}
        onChange={(event) => onChange(event.target.value)}
      />
    </div>
  )
}
export function EmptyState({ title, detail }: { title: string; detail: string }) {
  return (
    <div className="py-16 text-center">
      <p className="section-title">{title}</p>
      <p className="mt-2 text-sm text-muted-foreground">{detail}</p>
    </div>
  )
}
