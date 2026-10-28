import { useIdentity } from "@/components/identity"
import { usePeople } from "@/components/people-directory"
import { useState, type ReactNode } from "react"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { createProject, validateProjectInput } from "@/demo/projects"
import { priorities } from "@/models/workspace-labels"
import type { FormErrors, Project, ProjectInput } from "@/models/workspace"

export function CreateProjectDialog({
  projects,
  actorId,
  onClose,
  onSave,
}: {
  projects: Project[]
  actorId: string
  onClose: () => void
  onSave: (projects: Project[]) => void
}) {
  const identity = useIdentity()

  const { people } = usePeople()
  const [input, setInput] = useState<ProjectInput>({
    code: "",
    name: "",
    pmId: "",
    start: "",
    end: "",
    priority: "Trung bình",
    description: "",
  })
  const [errors, setErrors] = useState<FormErrors>({})
  const [error, setError] = useState("")
  const field = (key: keyof ProjectInput, label: string, control: ReactNode) => (
    <div className="min-w-0 space-y-2">
      <label htmlFor={`project-${key}`} className="field-label block">
        {label}
      </label>
      {control}
      {errors[key] && (
        <p id={`project-${key}-error`} className="text-sm text-destructive">
          {errors[key]}
        </p>
      )}
    </div>
  )
  const props = (key: keyof ProjectInput) => ({
    id: `project-${key}`,
    className: "field w-full",
    value: input[key],
    "aria-invalid": !!errors[key],
    "aria-describedby": errors[key] ? `project-${key}-error` : undefined,
    onChange: (e: { target: { value: string } }) =>
      setInput((prev) => ({ ...prev, [key]: e.target.value })),
  })
  return (
    <Dialog
      open
      onOpenChange={(open) => !open && onClose()}
      title="Tạo dự án"
      description="Chọn PM ngay khi tạo. PM sẽ chỉ định phụ trách chuyên môn và tổ chức triển khai sau."
      wide
    >
      <form
        noValidate
        onSubmit={(e) => {
          e.preventDefault()
          const nextErrors = validateProjectInput(projects, input, people)
          setErrors(nextErrors)
          setError("")
          if (Object.keys(nextErrors).length) {
            document.getElementById(`project-${Object.keys(nextErrors)[0]}`)?.focus()
            return
          }
          try {
            onSave(createProject(projects, actorId, input, people, identity))
          } catch (err) {
            setError((err as Error).message)
          }
        }}
        className="space-y-5"
      >
        <div className="grid gap-5 sm:grid-cols-2">
          {field(
            "code",
            "Mã dự án *",
            <input {...props("code")} maxLength={30} placeholder="VD: CRM-2026" required />,
          )}
          {field("name", "Tên dự án *", <input {...props("name")} maxLength={150} required />)}
          {field(
            "pmId",
            "Quản lý dự án (PM) *",
            <select {...props("pmId")} required>
              <option value="">Chọn người quản lý</option>
              {people
                .filter((p) => p.active)
                .map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.name}
                  </option>
                ))}
            </select>,
          )}
          {field(
            "priority",
            "Ưu tiên *",
            <select {...props("priority")}>
              {priorities.map((value) => (
                <option key={value}>{value}</option>
              ))}
            </select>,
          )}
          {field("start", "Ngày bắt đầu *", <input {...props("start")} type="date" required />)}
          {field(
            "end",
            "Hạn dự kiến *",
            <input {...props("end")} type="date" min={input.start || undefined} required />,
          )}
        </div>
        {field(
          "description",
          "Mô tả",
          <textarea {...props("description")} rows={3} placeholder="Mục tiêu và phạm vi dự án" />,
        )}
        <p className="text-xs leading-5 text-muted-foreground">
          * Thông tin bắt buộc. Dự án mới ở trạng thái Lập kế hoạch; chưa có PTCM, nhóm hoặc công
          việc.
        </p>
        {(error || Object.keys(errors).length > 0) && (
          <p role="alert" className="text-sm text-destructive">
            {error || "Kiểm tra lại các trường được đánh dấu."}
          </p>
        )}
        <div className="flex justify-end gap-2 border-t border-border pt-5">
          <Button variant="outline" type="button" onClick={onClose}>
            Hủy
          </Button>
          <Button type="submit">Tạo dự án</Button>
        </div>
      </form>
    </Dialog>
  )
}
