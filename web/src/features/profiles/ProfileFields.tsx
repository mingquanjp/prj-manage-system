export function ProfileFields({ fields }: { fields: [string, string | null][] }) {
  return (
    <dl className="space-y-3">
      {fields.map(([label, value]) => (
        <div
          key={label}
          className="grid min-w-0 gap-1 text-sm leading-6 sm:grid-cols-[7.5rem_minmax(0,1fr)] sm:gap-3"
        >
          <dt className="font-medium text-muted-foreground">{label}</dt>
          <dd className="min-w-0 whitespace-pre-wrap break-words text-foreground">
            {value ?? <span className="text-muted-foreground">Chưa cập nhật</span>}
          </dd>
        </div>
      ))}
    </dl>
  )
}
