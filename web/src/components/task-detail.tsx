import { Cross2Icon } from "@radix-ui/react-icons"
import { Button } from "@/components/ui/button"
import { StatusBadge } from "@/components/shared"
import { levels, type Task } from "@/demo/model"

export function TaskDetailHeader({
  task,
  onClose,
}: {
  task: Task
  onClose: () => void
}) {
  return (
    <header className="border-b border-border pb-4">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="eyebrow">{task.code}</p>
          <h2 className="mt-2 break-words text-lg font-semibold leading-7 text-balance">
            {task.name}
          </h2>
        </div>
        <Button
          size="icon"
          variant="ghost"
          aria-label="Đóng chi tiết công việc"
          onClick={onClose}
        >
          <Cross2Icon />
        </Button>
      </div>
      <div className="mt-3">
        <StatusBadge status={task.status} />
      </div>
    </header>
  )
}

// Chỉ trình bày đặc tả; nhật ký và kết quả sẽ được ghép vào khung khi có chức năng thật.
export function TaskSpecification({
  task,
  showWeights = false,
}: {
  task: Task
  showWeights?: boolean
}) {
  return (
    <section
      aria-label="Đặc tả công việc"
      className="mt-4 divide-y divide-border border-y border-border"
    >
      <details open className="py-3">
        <summary className="section-title cursor-pointer">
          Mô tả và đầu ra
        </summary>
        <p className="mt-3 whitespace-pre-wrap break-words text-sm leading-6 text-muted-foreground">
          {task.description}
        </p>
      </details>
      <details open className="py-3">
        <summary className="section-title cursor-pointer">
          Tiêu chí nghiệm thu
        </summary>
        <p className="mt-3 whitespace-pre-wrap break-words text-sm leading-6 text-muted-foreground">
          {task.criteria}
        </p>
      </details>
      <details className="py-3">
        <summary className="section-title cursor-pointer">
          Yêu cầu kỹ năng{" "}
          <span className="ml-1 text-muted-foreground">
            ({task.requirements.length})
          </span>
        </summary>
        <ul className="mt-3 space-y-3 text-sm">
          {task.requirements.map((r) => (
            <li key={r.skill}>
              <div className="flex justify-between gap-3">
                <span>{r.skill}</span>
                <span className="text-right text-muted-foreground">
                  {levels[r.level]}
                </span>
              </div>
              <p className="mt-1 text-xs text-muted-foreground">
                {r.mandatory ? "Bắt buộc" : "Bổ trợ"}
                {showWeights && ` · Trọng số ${r.weight} (minh họa)`}
              </p>
            </li>
          ))}
        </ul>
      </details>
    </section>
  )
}
