import { AssignmentTaskList } from "./AssignmentTaskList"
import { AssignmentCandidateTable } from "./AssignmentCandidateTable"
import { AssignmentDraftPanel } from "./AssignmentDraftPanel"
import { useState, type Dispatch, type SetStateAction } from "react"
import { ArrowLeftIcon, ExclamationTriangleIcon, InfoCircledIcon } from "@radix-ui/react-icons"
import { EmptyState, SearchInput } from "@/components/shared"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { commitDraft, recommendations, skillGaps, validateDraft } from "@/demo/assignment"
import { getPerson, isEligible } from "@/demo/identity"
import { levels } from "@/models/workspace-labels"
import { people } from "@/demo/fixtures"
import type { Choice, Draft, Project, Task } from "@/models/workspace"
import { cn } from "@/lib/utils"

type Mode = "manual" | "recommended"
export function AssignmentWorkspace({
  project,
  actorId,
  tasks,
  selectedIds,
  initialMode,
  draft,
  onDraftChange: setDraft,
  onBack,
  onCommit,
}: {
  project: Project
  actorId: string
  tasks: Task[]
  selectedIds: string[]
  initialMode: Mode
  draft: Draft
  onDraftChange: Dispatch<SetStateAction<Draft>>
  onBack: () => void
  onCommit: (tasks: Task[]) => void
}) {
  const [ids, setIds] = useState(selectedIds)
  const [activeId, setActiveId] = useState(selectedIds[0])
  const [modes, setModes] = useState<Record<string, Mode>>({})
  const [runTimes, setRunTimes] = useState<Record<string, string>>(() =>
    initialMode === "recommended"
      ? Object.fromEntries(selectedIds.map((id) => [id, new Date().toLocaleTimeString("vi-VN")]))
      : {},
  )
  const [query, setQuery] = useState("")
  const [confirm, setConfirm] = useState(false)
  const [error, setError] = useState("")
  const task = tasks.find((t) => t.id === activeId)!
  const selectedTasks = ids.map((id) => tasks.find((t) => t.id === id)!).filter(Boolean)
  const mode = modes[activeId] ?? initialMode
  const ranked = task ? recommendations(project, task) : []
  const candidates = task
    ? (mode === "recommended"
        ? ranked.map((r) => r.person)
        : people.filter((p) => isEligible(project, p))
      ).filter((p) =>
        `${p.name} ${p.position}`.toLocaleLowerCase("vi").includes(query.toLocaleLowerCase("vi")),
      )
    : []
  const errors = validateDraft(project, actorId, tasks, draft)
  const complete = selectedTasks.length > 0 && selectedTasks.every((t) => draft[t.id])
  const total = selectedTasks.reduce((sum, t) => sum + t.estimate, 0)
  const chosen = task ? draft[task.id] : undefined
  const chosenIsVisible = candidates.some((p) => p.id === chosen?.personId)
  const openConfirmation = () => {
    setError("")
    setConfirm(true)
  }
  const changeMode = (value: Mode) => {
    setModes((prev) => ({ ...prev, [activeId]: value }))
    setQuery("")
    if (value === "recommended" && !runTimes[activeId])
      setRunTimes((prev) => ({
        ...prev,
        [activeId]: new Date().toLocaleTimeString("vi-VN"),
      }))
  }
  const selectPerson = (personId: string) => {
    if (draft[activeId]?.personId === personId && draft[activeId]?.mode === mode) return
    const rank =
      mode === "recommended" ? ranked.find((r) => r.person.id === personId)?.rank : undefined
    setDraft((prev) => ({
      ...prev,
      [activeId]: {
        personId,
        mode,
        rank,
        reason: "",
        overrideReason: "",
        hadRun: Boolean(runTimes[activeId]),
      },
    }))
    setError("")
  }
  const patchChoice = (patch: Partial<Choice>) =>
    setDraft((prev) => ({
      ...prev,
      [activeId]: { ...prev[activeId], ...patch },
    }))
  const selectTask = (id: string) => {
    setActiveId(id)
    setQuery("")
  }
  const removeTask = (id: string) => {
    const next = ids.filter((value) => value !== id)
    setIds(next)
    if (id === activeId) {
      setActiveId(next[0])
      setQuery("")
    }
    setDraft((prev) => Object.fromEntries(Object.entries(prev).filter(([key]) => key !== id)))
  }

  return (
    <section aria-label="Không gian phân công" className="assignment-shell min-w-0">
      <div className="mb-5 flex flex-wrap items-center justify-between gap-4">
        <div>
          <button
            type="button"
            onClick={onBack}
            className="mb-3 inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground"
          >
            <ArrowLeftIcon />
            Về danh sách công việc
          </button>
          <h2 className="text-xl font-semibold text-balance">Phân công công việc</h2>
          <p className="mt-1.5 text-sm text-muted-foreground">
            So sánh, lựa chọn và kiểm tra bản nháp trước khi giao việc.
          </p>
        </div>
        <span className="rounded-md border border-border bg-white px-3 py-2 text-sm">
          {selectedTasks.length} công việc <span className="mx-2 text-border">/</span> {total} giờ
          ước tính
        </span>
      </div>
      <div className="assignment-toolbar sticky top-4 z-10 mb-4 flex flex-wrap items-center justify-between gap-3 rounded-lg border border-border bg-background p-3 shadow-sm">
        <div className="min-w-0 text-sm">
          <p className="field-label">
            {task?.code ?? "Chưa chọn việc"} · {mode === "manual" ? "Thủ công" : "Theo gợi ý"}
          </p>
          <p className="mt-1 text-xs text-muted-foreground" aria-live="polite">
            Đã chọn người cho {Object.keys(draft).length}/{ids.length} việc · Chưa lưu phân công
          </p>
          {Object.keys(draft).length > 0 && errors.length > 0 && (
            <p className="mt-1 text-xs text-warning" aria-live="polite">
              {errors[0]}
            </p>
          )}
        </div>
        <Button disabled={!complete || errors.length > 0} onClick={openConfirmation}>
          Kiểm tra và xác nhận
        </Button>
      </div>
      <div className="assignment-grid grid items-start gap-4">
        <AssignmentTaskList
          selectedTasks={selectedTasks}
          activeId={activeId}
          draft={draft}
          onSelectTask={selectTask}
          onRemoveTask={removeTask}
        />
        <div className="assignment-candidates min-w-0 rounded-lg border border-border bg-white">
          {!task ? (
            <EmptyState
              title="Chưa có công việc được chọn"
              detail="Quay lại danh sách để chọn công việc cần phân công."
            />
          ) : (
            <>
              <div className="border-b border-border p-5">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <p className="eyebrow">
                      {task.code} · {task.group}
                    </p>
                    <h3 className="mt-1 text-lg font-semibold text-balance">{task.name}</h3>
                  </div>
                  <span className="whitespace-nowrap text-sm text-muted-foreground">
                    {task.estimate} giờ
                  </span>
                </div>
                <p className="field-label mt-4">Yêu cầu kỹ năng</p>
                <div className="mt-2 flex flex-wrap gap-2">
                  {task.requirements.map((r) => (
                    <span
                      key={r.skill}
                      className="rounded border border-border bg-background px-2 py-1 text-xs"
                    >
                      <span className="font-medium">{r.skill}</span> · {levels[r.level]}
                      {r.mandatory ? " · bắt buộc" : ""}
                    </span>
                  ))}
                </div>
              </div>
              <div className="p-5">
                <div className="flex flex-wrap gap-1 rounded-md bg-background p-1">
                  {(
                    [
                      ["recommended", "Theo gợi ý"],
                      ["manual", "Tra cứu thủ công"],
                    ] as const
                  ).map(([value, label]) => (
                    <button
                      type="button"
                      key={value}
                      aria-pressed={mode === value}
                      onClick={() => changeMode(value)}
                      className={cn(
                        "workspace-tab h-9 flex-1 rounded px-3",
                        mode === value
                          ? "bg-white shadow-sm"
                          : "text-muted-foreground hover:text-foreground",
                      )}
                    >
                      {label}
                    </button>
                  ))}
                </div>
                <div className="my-4 flex items-start gap-2 text-xs leading-5 text-muted-foreground">
                  <InfoCircledIcon className="mt-0.5 shrink-0" />
                  <p>
                    {mode === "recommended"
                      ? `Gợi ý mẫu lúc ${runTimes[activeId]}. Điểm và thứ hạng chỉ minh họa, chưa phải công thức cuối.`
                      : "Tra cứu toàn công ty từ năng lực chính thức. Không xếp hạng hoặc tính điểm ở nhánh thủ công."}
                  </p>
                </div>
                <SearchInput value={query} onChange={setQuery} placeholder="Tìm ứng viên…" />
                <p className="mt-2 text-xs text-muted-foreground" aria-live="polite">
                  {candidates.length} ứng viên hiển thị. Tìm kiếm không thay đổi bản nháp.
                </p>
                {chosen && !chosenIsVisible && (
                  <p className="mt-2 rounded-md bg-background p-2 text-xs leading-5 text-muted-foreground">
                    {getPerson(chosen.personId)?.name} vẫn ở trong bản nháp cho {task.code}, nhưng
                    không nằm trong danh sách đang xem.
                  </p>
                )}
                {chosen && chosen.mode !== mode && (
                  <p className="mt-2 text-xs leading-5 text-muted-foreground">
                    Đổi nhánh chỉ thay danh sách đang xem. Lựa chọn trong bản nháp giữ nguyên cho
                    tới khi bạn chọn lại hoặc bỏ lựa chọn.
                  </p>
                )}
              </div>
              <AssignmentCandidateTable
                task={task}
                tasks={tasks}
                candidates={candidates}
                ranked={ranked}
                chosen={chosen}
                mode={mode}
                draft={draft}
                query={query}
                onSelectPerson={selectPerson}
              />
              {chosen && (
                <div className="space-y-4 border-t border-border bg-background/40 p-5">
                  <div className="flex flex-wrap items-start justify-between gap-2">
                    <p className="text-sm">
                      Bản nháp {task.code}: <strong>{getPerson(chosen.personId)?.name}</strong>{" "}
                      <span className="text-muted-foreground">
                        · {chosen.mode === "manual" ? "Thủ công" : `Gợi ý hạng ${chosen.rank}`}
                      </span>
                    </p>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() =>
                        setDraft((prev) =>
                          Object.fromEntries(Object.entries(prev).filter(([id]) => id !== task.id)),
                        )
                      }
                    >
                      Bỏ lựa chọn
                    </Button>
                  </div>
                  {((chosen.mode === "recommended" && chosen.rank !== 1) ||
                    (chosen.mode === "manual" && chosen.hadRun)) && (
                    <label className="block">
                      <span className="field-label">
                        Lý do lựa chọn <span className="text-warning">*</span>
                      </span>
                      <textarea
                        className="field workspace-field mt-2 min-h-20 w-full"
                        value={chosen.reason}
                        onChange={(e) => patchChoice({ reason: e.target.value })}
                        placeholder={
                          chosen.mode === "manual"
                            ? "Vì sao chuyển sang phân công thủ công?"
                            : "Vì sao chọn ứng viên khác hạng đầu?"
                        }
                      />
                    </label>
                  )}
                  {skillGaps(task, getPerson(chosen.personId)!).length > 0 && (
                    <>
                      <p className="flex items-start gap-2 text-sm leading-6 text-warning">
                        <ExclamationTriangleIcon className="mt-1 shrink-0" />
                        Ứng viên chưa đáp ứng kỹ năng bắt buộc. Cần lý do để phân công thủ công
                        ngoại lệ.
                      </p>
                      <label className="block">
                        <span className="field-label">
                          Lý do ngoại lệ kỹ năng <span className="text-warning">*</span>
                        </span>
                        <textarea
                          className="field workspace-field mt-2 min-h-20 w-full"
                          value={chosen.overrideReason}
                          onChange={(e) => patchChoice({ overrideReason: e.target.value })}
                        />
                      </label>
                    </>
                  )}
                </div>
              )}
            </>
          )}
        </div>
        <AssignmentDraftPanel
          selectedTasks={selectedTasks}
          taskCount={ids.length}
          tasks={tasks}
          draft={draft}
          errors={errors}
          complete={complete}
          error={error}
          onSelectTask={selectTask}
        />
      </div>
      <Dialog
        open={confirm}
        onOpenChange={setConfirm}
        title="Xác nhận phân công"
        description="Kiểm tra toàn bộ lựa chọn. Đây là thao tác mô phỏng; không lưu vào máy chủ."
      >
        <div className="mb-6 divide-y divide-border">
          {selectedTasks.map((t) => (
            <div key={t.id} className="flex justify-between gap-4 py-3 text-sm">
              <span>{t.name}</span>
              <strong className="text-right font-medium">
                {getPerson(draft[t.id]?.personId)?.name}
              </strong>
            </div>
          ))}
        </div>
        <p className="mb-5 text-sm text-muted-foreground">
          Chỉ cập nhật khi tất cả lựa chọn hợp lệ. Tổng {total} giờ ước tính được bổ sung vào tải
          sau khi xác nhận.
        </p>
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={() => setConfirm(false)}>
            Quay lại chỉnh sửa
          </Button>
          <Button
            onClick={() => {
              try {
                const next = commitDraft(project, actorId, tasks, draft)
                setConfirm(false)
                onCommit(next)
              } catch (err) {
                setError(err instanceof Error ? err.message : "Chưa thể xác nhận.")
                setConfirm(false)
              }
            }}
          >
            Xác nhận mô phỏng
          </Button>
        </div>
      </Dialog>
    </section>
  )
}
