import type { SkillEvidence, SkillSupplement } from "@/models/employee-skills"

export function SkillEvidenceEditor({
  value,
  available,
  onChange,
}: {
  value: SkillSupplement
  available: SkillEvidence[]
  onChange: (value: SkillSupplement) => void
}) {
  return (
    <div className="space-y-4">
      <label className="block">
        <span className="field-label">Giải trình / lý do *</span>
        <textarea
          className="field mt-2 min-h-28 w-full"
          value={value.explanation}
          aria-required="true"
          placeholder="Việc bạn đã làm, phần đóng góp, mức độ tự chủ và kết quả; hoặc thông tin cần điều chỉnh và lý do."
          onChange={(e) => onChange({ ...value, explanation: e.target.value })}
        />
      </label>
      <label className="block">
        <span className="field-label">Căn cứ ngoài hệ thống (tùy chọn)</span>
        <textarea
          className="field mt-2 min-h-20 w-full"
          value={value.externalEvidence}
          placeholder="Mô tả, liên kết hoặc nguồn nhận xét liên quan. Không nhập thông tin bí mật."
          onChange={(e) => onChange({ ...value, externalEvidence: e.target.value })}
        />
      </label>
      <fieldset className="min-w-0">
        <legend className="field-label">Dẫn công việc của bạn (tùy chọn)</legend>
        <p className="mt-2 text-pretty text-xs leading-5 text-muted-foreground">
          Có thể chọn nhiều lượt đóng góp hoặc không chọn. Không bắt buộc có công việc trong hệ
          thống.
        </p>
        <div className="mt-3 max-h-64 space-y-2 overflow-auto">
          {available.map((e) => {
            const selected = value.evidence.find((item) => item.assignmentId === e.assignmentId)
            return (
              <div key={e.assignmentId} className="rounded-lg border border-border p-3">
                <label className="flex items-start gap-2 text-sm">
                  <input
                    type="checkbox"
                    className="mt-1"
                    checked={Boolean(selected)}
                    onChange={(event) =>
                      onChange({
                        ...value,
                        evidence: event.target.checked
                          ? [
                              ...value.evidence,
                              { assignmentId: e.assignmentId, contributionNote: "" },
                            ]
                          : value.evidence.filter((item) => item.assignmentId !== e.assignmentId),
                      })
                    }
                  />
                  <span className="min-w-0 break-words">
                    {e.taskCode} · {e.taskName}
                    <span className="mt-1 block text-xs text-muted-foreground">
                      {e.submission
                        ? "Có kết quả nộp của lượt đóng góp này"
                        : "Chưa có kết quả nộp"}
                    </span>
                  </span>
                </label>
                {selected && (
                  <label className="mt-3 block">
                    <span className="field-label">
                      Đóng góp và liên hệ với kỹ năng tại {e.taskCode}
                    </span>
                    <textarea
                      className="field mt-2 min-h-16 w-full"
                      value={selected.contributionNote}
                      onChange={(event) =>
                        onChange({
                          ...value,
                          evidence: value.evidence.map((item) =>
                            item.assignmentId === e.assignmentId
                              ? { ...item, contributionNote: event.target.value }
                              : item,
                          ),
                        })
                      }
                    />
                  </label>
                )}
              </div>
            )
          })}
        </div>
        {!available.length && (
          <p className="mt-2 text-xs text-muted-foreground">
            Bạn chưa có lượt đóng góp nội bộ để dẫn chiếu.
          </p>
        )}
      </fieldset>
    </div>
  )
}
