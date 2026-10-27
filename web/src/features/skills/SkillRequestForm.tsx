import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { useSkillCatalog } from "@/components/skill-catalog"
import { levels } from "@/models/workspace-labels"
import { officialValues } from "@/demo/employee-skills"
import type {
  EmployeeSkillsState,
  SkillEvidence,
  SkillRequest,
  SkillRequestInput,
  SkillSupplement,
} from "@/models/employee-skills"
import { SkillEvidenceEditor } from "./SkillEvidenceEditor"
import { SkillValuesView } from "./SkillRequestDetails"

export function SkillRequestForm({
  employeeId,
  state,
  available,
  request,
  onSubmit,
  onSupplement,
  onClose,
}: {
  employeeId: string
  state: EmployeeSkillsState
  available: SkillEvidence[]
  request?: SkillRequest
  onSubmit: (input: SkillRequestInput) => void
  onSupplement: (id: string, version: number, input: SkillSupplement) => void
  onClose: () => void
}) {
  const { catalog, skillName } = useSkillCatalog()
  const [skillId, setSkillId] = useState(request?.skillId ?? "")
  const [level, setLevel] = useState(String(request?.proposed?.level ?? 1))
  const [years, setYears] = useState(request?.proposed?.years?.toString() ?? "")
  const [note, setNote] = useState(request?.proposed?.note ?? "")
  const [evidence, setEvidence] = useState<SkillSupplement>(() => ({
    explanation: request?.explanation ?? "",
    externalEvidence: request?.externalEvidence ?? "",
    evidence:
      request?.evidence.map((e) => ({
        assignmentId: e.assignmentId,
        contributionNote: e.contributionNote,
      })) ?? [],
  }))
  const [error, setError] = useState("")
  const [discard, setDiscard] = useState(false)
  // Giá trị nền được chụp khi mở form, không âm thầm thay trước khi gửi.
  const [beforeBySkill] = useState(() =>
    Object.fromEntries(catalog.skills.map((s) => [s.id, officialValues(state, employeeId, s.id)])),
  )
  const initialEvidence = {
    explanation: request?.explanation ?? "",
    externalEvidence: request?.externalEvidence ?? "",
    evidence:
      request?.evidence.map((e) => ({
        assignmentId: e.assignmentId,
        contributionNote: e.contributionNote,
      })) ?? [],
  }
  const dirty = Boolean(
    request
      ? JSON.stringify(evidence) !== JSON.stringify(initialEvidence)
      : skillId ||
          level !== "1" ||
          years ||
          note ||
          evidence.explanation ||
          evidence.externalEvidence ||
          evidence.evidence.length,
  )
  const close = () => (dirty ? setDiscard(true) : onClose())
  const before = beforeBySkill[skillId] ?? null
  const operation = before ? "UPDATE" : "ADD"
  return (
    <>
      <Dialog
        wide
        open
        onOpenChange={(open) => {
          if (!open) close()
        }}
        title={request ? `Bổ sung căn cứ · ${skillName(skillId)}` : "Đề nghị cập nhật kỹ năng"}
        description={
          request
            ? "Chỉ bổ sung căn cứ khi đang chờ; kỹ năng và nội dung đề nghị giữ nguyên."
            : "Đề nghị được Giám đốc xét duyệt trước khi thay đổi năng lực chính thức."
        }
      >
        <form
          noValidate
          className="space-y-5"
          onSubmit={(event) => {
            event.preventDefault()
            try {
              if (request) onSupplement(request.id, request.version, evidence)
              else
                onSubmit({
                  skillId,
                  operation,
                  before,
                  proposed: {
                    level: Number(level),
                    years: years.trim() === "" ? null : Number(years),
                    note,
                  },
                  ...evidence,
                })
              onClose()
            } catch (err) {
              setError(err instanceof Error ? err.message : "Chưa thể gửi đề nghị.")
            }
          }}
        >
          {!request && (
            <>
              <div className="grid gap-4">
                <label className="block">
                  <span className="field-label">Kỹ năng *</span>
                  <select
                    autoFocus
                    className="field mt-2 w-full"
                    value={skillId}
                    aria-required="true"
                    onChange={(event) => {
                      const id = event.target.value
                      const current = beforeBySkill[id]
                      setSkillId(id)
                      setLevel(String(current?.level ?? 1))
                      setYears(current?.years?.toString() ?? "")
                      setNote(current?.note ?? "")
                      setError("")
                    }}
                  >
                    <option value="">Chọn kỹ năng</option>
                    {catalog.skills.map((s) => (
                      <option
                        key={s.id}
                        value={s.id}
                        disabled={state.requests.some(
                          (r) =>
                            r.employeeId === employeeId &&
                            r.skillId === s.id &&
                            r.status === "PENDING",
                        )}
                      >
                        {s.name}
                        {state.requests.some(
                          (r) =>
                            r.employeeId === employeeId &&
                            r.skillId === s.id &&
                            r.status === "PENDING",
                        )
                          ? " · Đang chờ"
                          : ""}
                      </option>
                    ))}
                  </select>
                </label>
              </div>
              {skillId && (
                <section className="rounded-lg bg-accent p-3">
                  <h3 className="field-label mb-2">
                    {before ? "Điều chỉnh kỹ năng đã có" : "Bổ sung kỹ năng mới"}
                  </h3>
                  <SkillValuesView value={before} />
                </section>
              )}
              <>
                <div className="grid gap-4 sm:grid-cols-2">
                  <label className="block">
                    <span className="field-label">Mức đề nghị *</span>
                    <select
                      className="field mt-2 w-full"
                      aria-required="true"
                      value={level}
                      onChange={(e) => setLevel(e.target.value)}
                    >
                      {levels.slice(1).map((name, i) => (
                        <option key={name} value={i + 1}>
                          {name}
                        </option>
                      ))}
                    </select>
                  </label>
                  <label className="block">
                    <span className="field-label">Số năm kinh nghiệm (tùy chọn)</span>
                    <input
                      type="number"
                      min="0"
                      max="999.9"
                      step="0.1"
                      className="field mt-2 w-full"
                      value={years}
                      onChange={(e) => setYears(e.target.value)}
                    />
                  </label>
                </div>
                <label className="block">
                  <span className="field-label">Ghi chú năng lực (tùy chọn)</span>
                  <textarea
                    className="field mt-2 min-h-16 w-full"
                    value={note}
                    onChange={(e) => setNote(e.target.value)}
                  />
                </label>
              </>
            </>
          )}
          <SkillEvidenceEditor value={evidence} available={available} onChange={setEvidence} />
          {error && (
            <p role="alert" className="text-sm text-warning">
              {error}
            </p>
          )}
          <div className="flex flex-wrap justify-end gap-2 border-t border-border pt-4">
            <Button variant="outline" onClick={close}>
              Hủy
            </Button>
            <Button type="submit">{request ? "Lưu căn cứ bổ sung" : "Gửi đề nghị kỹ năng"}</Button>
          </div>
        </form>
      </Dialog>
      <Dialog
        open={discard}
        onOpenChange={setDiscard}
        title="Bỏ nội dung chưa gửi?"
        description="Đề nghị chưa được lưu; năng lực chính thức giữ nguyên."
      >
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={() => setDiscard(false)}>
            Tiếp tục nhập
          </Button>
          <Button onClick={onClose}>Bỏ nội dung</Button>
        </div>
      </Dialog>
    </>
  )
}
