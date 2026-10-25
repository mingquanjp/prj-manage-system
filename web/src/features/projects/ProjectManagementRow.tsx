import type { ReactNode } from "react"

export function ProjectManagementRow({
  title,
  description,
  children,
}: {
  title: string
  description: string
  children: ReactNode
}) {
  return (
    <div className="flex flex-col items-start gap-3 border-b border-border py-4 last:border-b-0 sm:flex-row sm:items-center sm:justify-between sm:gap-6">
      <div className="min-w-0 max-w-2xl">
        <h3 className="field-label text-balance">{title}</h3>
        <p className="mt-1 text-pretty text-xs leading-5 text-muted-foreground">{description}</p>
      </div>
      <div className="shrink-0">{children}</div>
    </div>
  )
}
