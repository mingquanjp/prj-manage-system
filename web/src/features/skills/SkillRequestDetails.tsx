import { usePeople } from "@/components/people-directory"
import { useSkillCatalog } from "@/components/skill-catalog"
import { levels } from "@/models/workspace-labels"
import {
  skillOperationLabels,
  skillRequestLabels,
  type SkillRequest,
  type SkillValues,
} from "@/models/employee-skills"
import { cn } from "@/lib/utils"

export function SkillRequestBadge({ status }: { status: SkillRequest["status"] }) {
  return (
    <span
      className={cn(
        "inline-flex rounded px-2 py-1 text-xs font-medium",
        status === "PENDING"
          ? "bg-warning-soft text-warning"
          : status === "APPROVED"
            ? "bg-success-soft text-success"
            : "bg-accent text-muted-foreground",
      )}
    >
      {skillRequestLabels[status]}
    </span>
  )
}
export function SkillValuesView({ value }: { value: SkillValues | null }) {
  return value ? (
    <div className="space-y-1 text-sm">
      <p className="font-medium">{levels[value.level]}</p>
      <p className="tabular-nums text-muted-foreground">
        {value.years === null ? "Chưa cập nhật số năm" : `${value.years} năm kinh nghiệm`}
      </p>
      {value.note && <p className="whitespace-pre-wrap break-words text-pretty">{value.note}</p>}
    </div>
  ) : (
    <p className="text-sm text-muted-foreground">Chưa có kỹ năng chính thức</p>
  )
}
export function SkillRequestDetails({
  request,
  targetGuidance,
}: {
  request: SkillRequest
  targetGuidance?: string
}) {
  const { skillName } = useSkillCatalog()
  const { getPerson } = usePeople()
  return (
    <div className="space-y-5 text-sm">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h3 className="section-title">
          {skillName(request.skillId)} · {skillOperationLabels[request.operation]}
        </h3>
        <SkillRequestBadge status={request.status} />
      </div>
      <div className="grid gap-4 rounded-lg border border-border p-4 sm:grid-cols-2">
        <section>
          <h4 className="field-label mb-2">Trước đề nghị</h4>
          <SkillValuesView value={request.before} />
        </section>
        <section>
          <h4 className="field-label mb-2">Đề nghị</h4>
          {request.operation === "REMOVE" ? (
            <p>Gỡ khỏi năng lực chính thức; giữ lịch sử quyết định.</p>
          ) : (
            <SkillValuesView value={request.proposed} />
          )}
        </section>
      </div>
      {targetGuidance && (
        <section className="rounded-lg bg-accent p-3">
          <h4 className="field-label mb-2">Cần đối chiếu ở mức đề nghị</h4>
          <p className="text-pretty leading-6">{targetGuidance}</p>
          <p className="mt-2 text-pretty text-xs leading-5 text-muted-foreground">
            Căn cứ cần thể hiện đúng kỹ năng và phần đóng góp cá nhân. Số năm hoặc một công việc đã
            nghiệm thu không tự chứng nhận mức này.
          </p>
        </section>
      )}
      <section>
        <h4 className="field-label mb-2">Giải trình / lý do</h4>
        <p className="whitespace-pre-wrap break-words text-pretty leading-6">
          {request.explanation}
        </p>
      </section>
      {request.externalEvidence && (
        <section>
          <h4 className="field-label mb-2">Căn cứ ngoài hệ thống</h4>
          <p className="whitespace-pre-wrap break-words text-pretty leading-6">
            {request.externalEvidence}
          </p>
        </section>
      )}
      <section>
        <h4 className="field-label mb-2">Căn cứ công việc ({request.evidence.length})</h4>
        {!request.evidence.length && (
          <p className="text-pretty text-muted-foreground">
            Không dẫn công việc nội bộ. Đối chiếu giải trình và căn cứ khác khi xét duyệt.
          </p>
        )}
        <div className="space-y-3">
          {request.evidence.map((e) => (
            <section key={e.assignmentId} className="rounded-lg border border-border p-3">
              <h5 className="font-medium text-pretty">
                {e.taskCode} · {e.taskName}
              </h5>
              <p className="mt-1 text-xs font-medium">
                {e.submission?.review
                  ? e.submission.review.result === "ACCEPTED"
                    ? "Nghiệm thu đạt"
                    : "Yêu cầu chỉnh sửa"
                  : e.submission
                    ? "Đã nộp · Chưa nghiệm thu"
                    : "Chưa nộp kết quả"}
              </p>
              <p className="mt-2 whitespace-pre-wrap break-words text-pretty leading-6">
                <strong>Đóng góp và liên hệ với kỹ năng: </strong>
                {e.contributionNote || "Chưa mô tả. Cần đối chiếu giải trình và căn cứ khác."}
              </p>
              {!e.requirements.some((r) => r.skill === request.skillId) && (
                <p className="mt-2 text-pretty text-xs leading-5 text-muted-foreground">
                  Yêu cầu công việc không ghi kỹ năng đang đề nghị. Cần làm rõ sự liên quan trong
                  phần đóng góp.
                </p>
              )}
              {!e.submission && (
                <p className="mt-2 text-pretty text-xs text-muted-foreground">
                  Việc được giao chưa tự chứng minh năng lực.
                </p>
              )}
              <details className="mt-3">
                <summary className="cursor-pointer font-medium">
                  Xem đặc tả, kết quả và nhận xét
                </summary>
                <div className="mt-3 space-y-3 leading-6">
                  <p className="text-pretty text-muted-foreground">
                    Đóng góp của {getPerson(request.employeeId)?.name}. Nội dung công việc và lần
                    nộp được ghi nhận khi gửi/bổ sung căn cứ.
                  </p>
                  <p className="whitespace-pre-wrap break-words">
                    <strong>Mô tả: </strong>
                    {e.description || "Chưa có mô tả"}
                  </p>
                  <p className="whitespace-pre-wrap break-words">
                    <strong>Tiêu chí nghiệm thu: </strong>
                    {e.criteria || "Chưa có tiêu chí"}
                  </p>
                  <ul className="space-y-1">
                    {e.requirements.map((r) => (
                      <li key={r.skill}>
                        {skillName(r.skill)} · {levels[r.level]}
                        {r.mandatory ? " · Bắt buộc" : ""}
                      </li>
                    ))}
                  </ul>
                  {e.submission ? (
                    <>
                      <p className="whitespace-pre-wrap break-words">
                        <strong>Kết quả nộp: </strong>
                        {e.submission.summary}
                      </p>
                      {e.submission.links.map((link) => (
                        <p key={link} className="break-words text-muted-foreground">
                          {link}
                        </p>
                      ))}
                      {e.submission.review ? (
                        <p className="whitespace-pre-wrap break-words">
                          <strong>
                            {e.submission.review.result === "ACCEPTED"
                              ? "Nghiệm thu đạt"
                              : "Yêu cầu chỉnh sửa"}
                            :{" "}
                          </strong>
                          {e.submission.review.comment} ·{" "}
                          {getPerson(e.submission.review.reviewerId)?.name}
                        </p>
                      ) : (
                        <p className="text-muted-foreground">Lần nộp chưa có kết quả nghiệm thu.</p>
                      )}
                    </>
                  ) : (
                    <p className="text-muted-foreground">
                      Lượt đóng góp này chưa có lần nộp kết quả. Việc được giao chưa tự chứng minh
                      năng lực.
                    </p>
                  )}
                </div>
              </details>
            </section>
          ))}
        </div>
      </section>
      {request.supplements.length > 0 && (
        <p className="text-xs text-muted-foreground">
          Đã bổ sung {request.supplements.length} lần · Gần nhất{" "}
          {new Date(request.supplements.at(-1)!.at).toLocaleString("vi-VN")}
        </p>
      )}
      {request.expertOpinion && (
        <section>
          <h4 className="field-label mb-2">Ý kiến chuyên môn được ghi nhận</h4>
          <p className="whitespace-pre-wrap break-words">{request.expertOpinion}</p>
        </section>
      )}
      {request.status !== "PENDING" && (
        <section className="rounded-lg bg-accent p-4">
          <h4 className="field-label mb-2">
            {request.status === "WITHDRAWN" ? "Lý do rút" : "Lý do quyết định"}
          </h4>
          <p className="whitespace-pre-wrap break-words">{request.decisionReason}</p>
          <p className="mt-2 text-xs text-muted-foreground">
            {getPerson(request.resolvedBy)?.name} ·{" "}
            {request.resolvedAt && new Date(request.resolvedAt).toLocaleString("vi-VN")}
          </p>
        </section>
      )}
    </div>
  )
}
