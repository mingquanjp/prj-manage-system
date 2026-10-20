import { useState } from "react"
import {
  ArrowLeftIcon,
  CheckIcon,
  Cross2Icon,
  ExclamationTriangleIcon,
  InfoCircledIcon,
  PersonIcon,
} from "@radix-ui/react-icons"
import { Avatar, EmptyState, SearchInput } from "@/components/shared"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import {
  addedLoad,
  commitDraft,
  getPerson,
  isEligible,
  levels,
  loadFor,
  people,
  recommendations,
  skillGaps,
  validateDraft,
  type Choice,
  type Draft,
  type Project,
  type Task,
} from "@/demo/model"
import { cn } from "@/lib/utils"

type Mode = "manual" | "recommended"
export function AssignmentWorkspace({
  project,
  actorId,
  tasks,
  selectedIds,
  initialMode,
  onBack,
  onCommit,
}: {
  project: Project
  actorId: string
  tasks: Task[]
  selectedIds: string[]
  initialMode: Mode
  onBack: () => void
  onCommit: (tasks: Task[]) => void
}) {
  const [ids, setIds] = useState(selectedIds)
  const [activeId, setActiveId] = useState(selectedIds[0])
  const [modes, setModes] = useState<Record<string, Mode>>({})
  const [runTimes, setRunTimes] = useState<Record<string, string>>(() =>
    initialMode === "recommended"
      ? Object.fromEntries(
          selectedIds.map((id) => [id, new Date().toLocaleTimeString("vi-VN")]),
        )
      : {},
  )
  const [draft, setDraft] = useState<Draft>({})
  const [query, setQuery] = useState("")
  const [confirm, setConfirm] = useState(false)
  const [discard, setDiscard] = useState(false)
  const [error, setError] = useState("")
  const task = tasks.find((t) => t.id === activeId)!
  const selectedTasks = ids
    .map((id) => tasks.find((t) => t.id === id)!)
    .filter(Boolean)
  const mode = modes[activeId] ?? initialMode
  const ranked = task ? recommendations(project, task) : []
  const candidates = task
    ? (mode === "recommended"
        ? ranked.map((r) => r.person)
        : people.filter((p) => isEligible(project, p))
      ).filter((p) =>
        `${p.name} ${p.position}`
          .toLocaleLowerCase("vi")
          .includes(query.toLocaleLowerCase("vi")),
      )
    : []
  const errors = validateDraft(project, actorId, tasks, draft)
  const complete =
    selectedTasks.length > 0 && selectedTasks.every((t) => draft[t.id])
  const total = selectedTasks.reduce((sum, t) => sum + t.estimate, 0)
  const chosen = task ? draft[task.id] : undefined
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
    if (
      draft[activeId]?.personId === personId &&
      draft[activeId]?.mode === mode
    )
      return
    const rank =
      mode === "recommended"
        ? ranked.find((r) => r.person.id === personId)?.rank
        : undefined
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
  const removeTask = (id: string) => {
    const next = ids.filter((value) => value !== id)
    setIds(next)
    if (id === activeId) setActiveId(next[0])
    setDraft((prev) =>
      Object.fromEntries(Object.entries(prev).filter(([key]) => key !== id)),
    )
  }

  return (
    <section aria-label="Không gian phân công">
      <div className="mb-5 flex flex-wrap items-center justify-between gap-4">
        <div>
          <button
            type="button"
            onClick={() =>
              Object.keys(draft).length ? setDiscard(true) : onBack()
            }
            className="mb-3 inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground"
          >
            <ArrowLeftIcon />
            Về danh sách công việc
          </button>
          <h2 className="text-xl font-semibold">Phân công công việc</h2>
          <p className="mt-1.5 text-sm text-muted-foreground">
            So sánh, lựa chọn và kiểm tra bản nháp trước khi giao việc.
          </p>
        </div>
        <span className="rounded-md border border-border bg-white px-3 py-2 text-sm">
          {selectedTasks.length} công việc{" "}
          <span className="mx-2 text-border">/</span> {total} giờ ước tính
        </span>
      </div>
      <div className="assignment-grid grid items-start gap-4">
        <aside className="min-w-0 rounded-lg border border-border bg-white">
          <h3 className="border-b border-border px-4 py-3 text-sm font-medium">
            Công việc đã chọn
          </h3>
          <div className="space-y-1 p-2">
            {selectedTasks.map((t) => (
              <div
                key={t.id}
                className={cn(
                  "flex rounded-md",
                  t.id === activeId ? "bg-accent" : "hover:bg-background",
                )}
              >
                <button
                  type="button"
                  aria-pressed={t.id === activeId}
                  className="min-w-0 flex-1 p-3 text-left"
                  onClick={() => {
                    setActiveId(t.id)
                    setQuery("")
                  }}
                >
                  <p className="text-xs text-muted-foreground">{t.code}</p>
                  <p className="mt-1 text-sm font-medium leading-5">{t.name}</p>
                  <div className="mt-2 flex items-center justify-between gap-2 text-xs text-muted-foreground">
                    <span>{t.estimate} giờ</span>
                    {draft[t.id] ? (
                      <span className="inline-flex items-center gap-1 text-success">
                        <CheckIcon />
                        Đã chọn
                      </span>
                    ) : (
                      <span>Chưa chọn</span>
                    )}
                  </div>
                </button>
                <button
                  type="button"
                  aria-label={`Bỏ ${t.name} khỏi lần phân công`}
                  className="mr-1 mt-2 self-start rounded p-1.5 text-muted-foreground hover:bg-white"
                  onClick={() => removeTask(t.id)}
                >
                  <Cross2Icon />
                </button>
              </div>
            ))}
          </div>
        </aside>
        <div className="min-w-0 rounded-lg border border-border bg-white">
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
                    <h3 className="mt-1 text-lg font-semibold">{task.name}</h3>
                  </div>
                  <span className="whitespace-nowrap text-sm text-muted-foreground">
                    {task.estimate} giờ
                  </span>
                </div>
                <div className="mt-4 flex flex-wrap gap-2">
                  {task.requirements.map((r) => (
                    <span
                      key={r.skill}
                      className="rounded border border-border bg-background px-2 py-1 text-xs"
                    >
                      {r.skill} · {levels[r.level]}
                      {r.mandatory ? " · bắt buộc" : ""}
                    </span>
                  ))}
                </div>
              </div>
              <div className="p-5">
                <div className="flex flex-wrap gap-1 rounded-md bg-background p-1">
                  {([
                    ["recommended", "Theo gợi ý"],
                    ["manual", "Tra cứu thủ công"],
                  ] as const).map(([value, label]) => (
                    <button
                      type="button"
                      key={value}
                      aria-pressed={mode === value}
                      onClick={() => changeMode(value)}
                      className={cn(
                        "h-9 flex-1 rounded px-3 text-sm",
                        mode === value
                          ? "bg-white font-medium shadow-sm"
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
                <SearchInput
                  value={query}
                  onChange={setQuery}
                  placeholder="Tìm ứng viên…"
                />
              </div>
              <div className="overflow-x-auto">
                <table className="w-full min-w-[470px] text-left text-sm">
                  <thead>
                    <tr className="table-heading">
                      <th className="py-3 pl-5">Ứng viên</th>
                      <th className="px-2">Tải hiện tại</th>
                      {mode === "recommended" && <th>Điểm mẫu</th>}
                      <th className="pr-5 text-right">Lựa chọn</th>
                    </tr>
                  </thead>
                  <tbody>
                    {candidates.map((person) => {
                      const recommendation = ranked.find(
                        (r) => r.person.id === person.id,
                      )
                      const checked =
                        chosen?.personId === person.id && chosen.mode === mode
                      const gaps = skillGaps(task, person)
                      return (
                        <tr
                          key={person.id}
                          className={cn(
                            "border-t border-border align-top",
                            checked && "bg-accent/50",
                          )}
                        >
                          <td className="py-4 pl-5 pr-2">
                            <div className="flex items-start gap-2.5">
                              <Avatar initials={person.initials} />
                              <div>
                                <p className="font-medium">{person.name}</p>
                                <p className="mt-0.5 text-xs text-muted-foreground">
                                  {person.position} · {person.seniority}
                                </p>
                                <details className="mt-2 text-xs">
                                  <summary className="cursor-pointer text-muted-foreground underline underline-offset-4">
                                    Đối chiếu kỹ năng
                                    {gaps.length > 0 ? " · còn thiếu" : ""}
                                  </summary>
                                  <div className="mt-3 space-y-2">
                                    {task.requirements.map((r) => (
                                      <p
                                        key={r.skill}
                                        className={cn(
                                          "leading-5",
                                          (person.skills[r.skill] ?? 0) <
                                            r.level && r.mandatory
                                            ? "text-warning"
                                            : "text-muted-foreground",
                                        )}
                                      >
                                        {r.skill}:{" "}
                                        {levels[person.skills[r.skill] ?? 0] ||
                                          "Chưa ghi nhận"}
                                        <br />
                                        <span className="text-muted-foreground">
                                          Cần{" "}
                                          {levels[r.level].toLocaleLowerCase(
                                            "vi",
                                          )}
                                          {r.mandatory ? " (bắt buộc)" : ""}
                                        </span>
                                      </p>
                                    ))}
                                  </div>
                                </details>
                              </div>
                            </div>
                          </td>
                          <td className="px-2 py-4 whitespace-nowrap tabular-nums">
                            <span className="font-medium">
                              {loadFor(person.id, tasks)} giờ
                            </span>
                            <p className="mt-1 text-xs text-muted-foreground">
                              Xuyên dự án
                            </p>
                          </td>
                          {mode === "recommended" && (
                            <td className="py-4 tabular-nums">
                              <strong className="font-medium">
                                {recommendation?.score.toLocaleString("vi-VN")}
                              </strong>
                              <p className="mt-1 text-xs text-muted-foreground">
                                Hạng {recommendation?.rank}
                              </p>
                            </td>
                          )}
                          <td className="py-4 pl-2 pr-5 text-right">
                            <Button
                              variant={checked ? "default" : "outline"}
                              size="sm"
                              aria-pressed={checked}
                              aria-label={`Chọn ${person.name}`}
                              onClick={() => selectPerson(person.id)}
                            >
                              {checked ? (
                                <>
                                  <CheckIcon />
                                  Đã chọn
                                </>
                              ) : (
                                "Chọn"
                              )}
                            </Button>
                          </td>
                        </tr>
                      )
                    })}
                  </tbody>
                </table>
                {!candidates.length && (
                  <EmptyState
                    title="Không có ứng viên phù hợp"
                    detail={
                      mode === "recommended"
                        ? "Thử tra cứu thủ công hoặc bỏ công việc này khỏi bản nháp."
                        : "Thử tìm bằng tên khác."
                    }
                  />
                )}
              </div>
              {chosen && (
                <div className="space-y-4 border-t border-border bg-background/40 p-5">
                  <p className="text-sm">
                    Đang chọn{" "}
                    <strong>{getPerson(chosen.personId)?.name}</strong>{" "}
                    <span className="text-muted-foreground">
                      ·{" "}
                      {chosen.mode === "manual"
                        ? "Thủ công"
                        : `Gợi ý hạng ${chosen.rank}`}
                    </span>
                  </p>
                  {((chosen.mode === "recommended" && chosen.rank !== 1) ||
                    (chosen.mode === "manual" && chosen.hadRun)) && (
                    <label className="block text-sm font-medium">
                      Lý do lựa chọn <span className="text-warning">*</span>
                      <textarea
                        className="field mt-2 min-h-20 w-full font-normal"
                        value={chosen.reason}
                        onChange={(e) =>
                          patchChoice({ reason: e.target.value })
                        }
                        placeholder="Vì sao lựa chọn khác gợi ý hạng đầu?"
                      />
                    </label>
                  )}
                  {skillGaps(task, getPerson(chosen.personId)!).length > 0 && (
                    <>
                      <p className="flex items-start gap-2 text-sm leading-6 text-warning">
                        <ExclamationTriangleIcon className="mt-1 shrink-0" />
                        Ứng viên chưa đáp ứng kỹ năng bắt buộc. Cần lý do để
                        phân công thủ công ngoại lệ.
                      </p>
                      <label className="block text-sm font-medium">
                        Lý do ngoại lệ kỹ năng{" "}
                        <span className="text-warning">*</span>
                        <textarea
                          className="field mt-2 min-h-20 w-full font-normal"
                          value={chosen.overrideReason}
                          onChange={(e) =>
                            patchChoice({ overrideReason: e.target.value })
                          }
                        />
                      </label>
                    </>
                  )}
                </div>
              )}
            </>
          )}
        </div>
        <aside className="rounded-lg border border-border bg-white p-4 assignment-draft">
          <div className="flex items-center justify-between">
            <h3 className="font-semibold">Bản nháp phân công</h3>
            <span className="count">
              {Object.keys(draft).length}/{ids.length}
            </span>
          </div>
          <p className="mt-2 text-xs leading-5 text-muted-foreground">
            Lựa chọn chưa làm thay đổi người phụ trách hoặc tải công việc.
          </p>
          <div className="my-5 space-y-4">
            {selectedTasks.map((t) => (
              <div
                key={t.id}
                className="border-b border-border pb-4 last:border-0 last:pb-0"
              >
                <button
                  type="button"
                  className="text-left text-sm leading-5 hover:underline"
                  onClick={() => {
                    setActiveId(t.id)
                    setQuery("")
                  }}
                >
                  {t.name}
                </button>
                <div className="mt-2 flex items-center gap-2 text-xs text-muted-foreground">
                  {draft[t.id] ? (
                    <>
                      <Avatar
                        initials={getPerson(draft[t.id].personId)!.initials}
                        small
                      />
                      <span>
                        {getPerson(draft[t.id].personId)?.name}
                        <span className="mt-0.5 block">
                          {draft[t.id].mode === "manual"
                            ? "Thủ công"
                            : `Gợi ý hạng ${draft[t.id].rank}`}
                        </span>
                      </span>
                    </>
                  ) : (
                    <>
                      <PersonIcon />
                      Chưa chọn người
                    </>
                  )}
                </div>
              </div>
            ))}
          </div>
          <div className="border-t border-border pt-4">
            <h4 className="text-sm font-medium">Tổng tải sau dự kiến</h4>
            <div className="mt-3 space-y-3">
              {[...new Set(Object.values(draft).map((c) => c.personId))].map(
                (id) => (
                  <div key={id} className="text-xs">
                    <p>{getPerson(id)?.name}</p>
                    <p className="mt-1 text-muted-foreground tabular-nums">
                      {loadFor(id, tasks)} + {addedLoad(id, draft, tasks)} ={" "}
                      <strong className="text-foreground">
                        {loadFor(id, tasks) + addedLoad(id, draft, tasks)} giờ
                      </strong>
                    </p>
                  </div>
                ),
              )}
              {!Object.keys(draft).length && (
                <p className="text-xs text-muted-foreground">
                  Hiển thị sau khi chọn ứng viên.
                </p>
              )}
            </div>
          </div>
          {Object.keys(draft).length > 0 && errors.length > 0 && (
            <div className="mt-4 rounded-md bg-warning-soft p-3 text-xs leading-5 text-warning">
              {errors.join(" ")}
            </div>
          )}
          <Button
            className="mt-5 w-full"
            disabled={!complete || errors.length > 0}
            onClick={() => setConfirm(true)}
          >
            Kiểm tra và xác nhận
          </Button>
          {!complete && (
            <p className="mt-2 text-xs leading-5 text-muted-foreground">
              Chọn người cho từng việc hoặc bỏ việc chưa cần giao.
            </p>
          )}
          {error && (
            <p role="alert" className="mt-3 text-sm text-destructive">
              {error}
            </p>
          )}
        </aside>
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
          Chỉ cập nhật khi tất cả lựa chọn hợp lệ. Tổng {total} giờ ước tính
          được bổ sung vào tải sau khi xác nhận.
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
                setError(
                  err instanceof Error ? err.message : "Chưa thể xác nhận.",
                )
                setConfirm(false)
              }
            }}
          >
            Xác nhận mô phỏng
          </Button>
        </div>
      </Dialog>
      <Dialog
        open={discard}
        onOpenChange={setDiscard}
        title="Bỏ bản nháp phân công?"
        description="Các lựa chọn chưa xác nhận sẽ không được lưu. Công việc vẫn giữ nguyên người phụ trách và trạng thái hiện tại."
      >
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={() => setDiscard(false)}>
            Tiếp tục chọn
          </Button>
          <Button onClick={onBack}>Bỏ bản nháp</Button>
        </div>
      </Dialog>
    </section>
  )
}
