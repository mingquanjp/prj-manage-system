import { useState, type ReactNode } from "react"
import { PlusIcon, TrashIcon } from "@radix-ui/react-icons"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { levels } from "@/models/workspace-labels"
import { saveTask, validateTaskInput } from "@/demo/task-structure"
import { skillCatalog } from "@/demo/fixtures"
import type { FormErrors, Project, Task, TaskGroup, TaskInput } from "@/models/workspace"

function Field({
  id,
  label,
  error,
  children,
  hint,
}: {
  id: string
  label: string
  error?: string
  children: ReactNode
  hint?: string
}) {
  return (
    <div className="min-w-0 space-y-2">
      <label htmlFor={id} className="field-label block">
        {label}
      </label>
      {children}
      {hint && (
        <p id={`${id}-hint`} className="text-xs leading-5 text-muted-foreground">
          {hint}
        </p>
      )}
      {error && (
        <p id={`${id}-error`} className="text-sm text-destructive">
          {error}
        </p>
      )}
    </div>
  )
}

export function TaskEditor({
  project,
  actorId,
  groups,
  tasks,
  task,
  defaultGroupId,
  onClose,
  onSave,
}: {
  project: Project
  actorId: string
  groups: TaskGroup[]
  tasks: Task[]
  task?: Task
  defaultGroupId?: string
  onClose: () => void
  onSave: (tasks: Task[], taskId: string) => void
}) {
  const [input, setInput] = useState<TaskInput>(() =>
    task
      ? structuredClone(task)
      : {
          code: "",
          name: "",
          groupId: defaultGroupId ?? "",
          description: "",
          criteria: "",
          estimate: Number.NaN,
          priority: "Trung bình",
          start: "",
          end: "",
          requirements: [{ skill: "", level: 1, weight: 1, mandatory: true }],
        },
  )
  const [errors, setErrors] = useState<FormErrors>({})
  const [failure, setFailure] = useState("")
  const update = <K extends keyof TaskInput>(key: K, value: TaskInput[K]) =>
    setInput((prev) => ({ ...prev, [key]: value }))
  const a11y = (key: string) => ({
    "aria-invalid": Boolean(errors[key]),
    "aria-describedby": errors[key] ? `${key}-error` : undefined,
  })
  const requirement = (index: number, value: Partial<TaskInput["requirements"][number]>) =>
    update(
      "requirements",
      input.requirements.map((r, i) => (i === index ? { ...r, ...value } : r)),
    )

  return (
    <Dialog
      open
      onOpenChange={(open) => {
        if (!open) onClose()
      }}
      title={task ? "Chỉnh sửa công việc" : "Tạo công việc"}
      description="Làm rõ phạm vi, đầu ra và yêu cầu kỹ năng trước khi phân công."
      wide
    >
      <form
        noValidate
        onSubmit={(event) => {
          event.preventDefault()
          const nextErrors = validateTaskInput(project, groups, tasks, input, task?.id)
          setErrors(nextErrors)
          setFailure("")
          if (Object.keys(nextErrors).length) {
            requestAnimationFrame(() =>
              document.querySelector<HTMLElement>('[aria-invalid="true"]')?.focus(),
            )
            return
          }
          try {
            const next = saveTask(project, actorId, groups, tasks, input, task?.id)
            onSave(next, task?.id ?? next[next.length - 1].id)
          } catch (error) {
            setFailure((error as Error).message)
          }
        }}
        className="space-y-6"
      >
        {Object.keys(errors).length > 0 && (
          <p
            role="alert"
            className="rounded-md border border-destructive/20 p-3 text-sm text-destructive"
          >
            Chưa lưu. Vui lòng kiểm tra các trường được đánh dấu bên dưới.
          </p>
        )}
        <section className="space-y-4" aria-label="Đặc tả công việc">
          <div className="grid gap-4 sm:grid-cols-[180px_minmax(0,1fr)]">
            <Field id="code" label="Mã công việc *" error={errors.code}>
              <input
                id="code"
                className="field w-full"
                placeholder={`${project.code.split("-")[0]}-T…`}
                value={input.code}
                maxLength={30}
                onChange={(e) => update("code", e.target.value)}
                {...a11y("code")}
              />
            </Field>
            <Field id="name" label="Tên công việc *" error={errors.name}>
              <input
                id="name"
                className="field w-full"
                value={input.name}
                maxLength={150}
                onChange={(e) => update("name", e.target.value)}
                {...a11y("name")}
              />
            </Field>
          </div>
          <div className="grid gap-4 sm:grid-cols-3">
            <Field id="groupId" label="Nhóm công việc *" error={errors.groupId}>
              <select
                id="groupId"
                className="field w-full"
                value={input.groupId}
                onChange={(e) => update("groupId", e.target.value)}
                {...a11y("groupId")}
              >
                <option value="">Chọn nhóm</option>
                {groups
                  .filter((g) => g.projectId === project.id)
                  .map((g) => (
                    <option key={g.id} value={g.id}>
                      {g.name}
                    </option>
                  ))}
              </select>
            </Field>
            <Field id="estimate" label="Ước tính (giờ) *" error={errors.estimate}>
              <input
                id="estimate"
                type="number"
                step="0.01"
                className="field w-full"
                value={Number.isNaN(input.estimate) ? "" : input.estimate}
                onChange={(e) => update("estimate", e.target.valueAsNumber)}
                {...a11y("estimate")}
              />
            </Field>
            <Field id="priority" label="Ưu tiên" error={errors.priority}>
              <select
                id="priority"
                className="field w-full"
                value={input.priority}
                onChange={(e) => update("priority", e.target.value as TaskInput["priority"])}
              >
                {["Thấp", "Trung bình", "Cao", "Khẩn cấp"].map((p) => (
                  <option key={p}>{p}</option>
                ))}
              </select>
            </Field>
          </div>
          <Field id="description" label="Mô tả và đầu ra mong đợi *" error={errors.description}>
            <textarea
              id="description"
              className="field min-h-28 w-full resize-y"
              placeholder="Cần thực hiện gì? Kết quả bàn giao gồm những gì?"
              value={input.description}
              onChange={(e) => update("description", e.target.value)}
              {...a11y("description")}
            />
          </Field>
          <Field id="criteria" label="Tiêu chí nghiệm thu *" error={errors.criteria}>
            <textarea
              id="criteria"
              className="field min-h-24 w-full resize-y"
              placeholder="Các điều kiện cụ thể để xác nhận công việc đạt yêu cầu."
              value={input.criteria}
              onChange={(e) => update("criteria", e.target.value)}
              {...a11y("criteria")}
            />
          </Field>
          <details
            className="rounded-md border border-border px-4 py-3"
            open={Boolean(input.start || input.end || errors.dates)}
          >
            <summary className="field-label">
              Ngày kế hoạch{" "}
              <span className="ml-1 font-normal text-muted-foreground">· Không bắt buộc</span>
            </summary>
            <p className="mt-2 text-xs text-muted-foreground">
              Trong khoảng {project.start.split("-").reverse().join("/")} –{" "}
              {project.end.split("-").reverse().join("/")}. Nhập đủ hai ngày hoặc để trống.
            </p>
            {task?.start && task.end && (
              <p className="mt-2 text-xs text-muted-foreground">
                Giữ cặp ngày đã lưu. Điều chỉnh kế hoạch có lý do và lịch sử sẽ được dựng ở đợt
                riêng.
              </p>
            )}
            <div className="mt-4 grid gap-4 sm:grid-cols-2">
              <Field id="start" label="Ngày bắt đầu">
                <input
                  id="start"
                  type="date"
                  disabled={Boolean(task?.start && task.end)}
                  className="field w-full"
                  min={project.start}
                  max={project.end}
                  value={input.start ?? ""}
                  onChange={(e) => update("start", e.target.value)}
                  aria-invalid={Boolean(errors.dates)}
                  aria-describedby={errors.dates ? "dates-error" : undefined}
                />
              </Field>
              <Field id="end" label="Ngày kết thúc">
                <input
                  id="end"
                  type="date"
                  disabled={Boolean(task?.start && task.end)}
                  className="field w-full"
                  min={project.start}
                  max={project.end}
                  value={input.end ?? ""}
                  onChange={(e) => update("end", e.target.value)}
                  aria-invalid={Boolean(errors.dates)}
                  aria-describedby={errors.dates ? "dates-error" : undefined}
                />
              </Field>
            </div>
            {errors.dates && (
              <p id="dates-error" className="mt-2 text-sm text-destructive">
                {errors.dates}
              </p>
            )}
          </details>
        </section>
        <section className="border-t border-border pt-5" aria-labelledby="requirements-heading">
          <div className="mb-2 flex flex-wrap items-center justify-between gap-3">
            <h3 id="requirements-heading" className="section-title">
              Yêu cầu kỹ năng
            </h3>
            <Button
              size="sm"
              variant="outline"
              onClick={() =>
                update("requirements", [
                  ...input.requirements,
                  { skill: "", level: 1, weight: 1, mandatory: false },
                ])
              }
            >
              <PlusIcon />
              Thêm kỹ năng
            </Button>
          </div>
          <p className="mb-4 text-xs leading-5 text-muted-foreground" id="weight-note">
            Trọng số do PTCM đặt. Giá trị hiện tại chỉ minh họa, chưa chốt miền số và chưa dùng để
            tính điểm gợi ý.
          </p>
          {errors.requirements && (
            <p role="alert" className="mb-3 text-sm text-destructive">
              {errors.requirements}
            </p>
          )}
          <div className="space-y-3">
            {input.requirements.map((r, i) => (
              <fieldset key={i} className="rounded-md border border-border p-3">
                <legend className="field-label px-1">Kỹ năng {i + 1}</legend>
                <div className="grid items-start gap-3 sm:grid-cols-[minmax(0,1.2fr)_minmax(0,1fr)_100px_40px]">
                  <Field id={`skill-${i}`} label="Kỹ năng *" error={errors[`skill-${i}`]}>
                    <select
                      id={`skill-${i}`}
                      className="field w-full"
                      value={r.skill}
                      onChange={(e) => requirement(i, { skill: e.target.value })}
                      {...a11y(`skill-${i}`)}
                    >
                      <option value="">Chọn kỹ năng</option>
                      {skillCatalog.map((s) => (
                        <option
                          key={s}
                          disabled={input.requirements.some(
                            (other, j) => j !== i && other.skill === s,
                          )}
                        >
                          {s}
                        </option>
                      ))}
                    </select>
                  </Field>
                  <Field id={`level-${i}`} label="Mức yêu cầu" error={errors[`level-${i}`]}>
                    <select
                      id={`level-${i}`}
                      className="field w-full"
                      value={r.level}
                      onChange={(e) => requirement(i, { level: Number(e.target.value) })}
                      {...a11y(`level-${i}`)}
                    >
                      {levels.slice(1).map((level, index) => (
                        <option key={level} value={index + 1}>
                          {level}
                        </option>
                      ))}
                    </select>
                  </Field>
                  <Field id={`weight-${i}`} label="Trọng số" error={errors[`weight-${i}`]}>
                    <input
                      id={`weight-${i}`}
                      type="number"
                      step="any"
                      className="field w-full"
                      value={Number.isNaN(r.weight) ? "" : r.weight}
                      onChange={(e) => requirement(i, { weight: e.target.valueAsNumber })}
                      {...a11y(`weight-${i}`)}
                      aria-describedby={`weight-note${errors[`weight-${i}`] ? ` weight-${i}-error` : ""}`}
                    />
                  </Field>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="sm:mt-7"
                    aria-label={`Xóa yêu cầu kỹ năng ${i + 1}`}
                    title="Xóa yêu cầu kỹ năng"
                    onClick={() =>
                      update(
                        "requirements",
                        input.requirements.filter((_, index) => index !== i),
                      )
                    }
                  >
                    <TrashIcon />
                  </Button>
                </div>
                <label className="mt-3 flex items-center gap-2 text-sm">
                  <input
                    type="checkbox"
                    checked={r.mandatory}
                    onChange={(e) => requirement(i, { mandatory: e.target.checked })}
                  />
                  Kỹ năng bắt buộc
                </label>
              </fieldset>
            ))}
          </div>
        </section>
        {failure && (
          <p role="alert" className="text-sm text-destructive">
            {failure}
          </p>
        )}
        <div className="flex flex-wrap items-center justify-between gap-3 border-t border-border pt-4">
          <p className="text-xs text-muted-foreground">Đặc tả khóa sau lần phân công đầu tiên.</p>
          <div className="flex gap-2">
            <Button variant="outline" onClick={onClose}>
              Hủy
            </Button>
            <Button type="submit">{task ? "Lưu thay đổi" : "Tạo công việc"}</Button>
          </div>
        </div>
      </form>
    </Dialog>
  )
}
