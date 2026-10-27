import { useSkillCatalog } from "@/components/skill-catalog"
import { CheckIcon } from "@radix-ui/react-icons"
import { Avatar, EmptyState } from "@/components/shared"
import { Button } from "@/components/ui/button"
import { addedLoad, loadFor, recommendations, skillGaps } from "@/demo/assignment"
import type { Choice, Draft, Person, Task } from "@/models/workspace"
import { levels } from "@/models/workspace-labels"
import { cn } from "@/lib/utils"

export function AssignmentCandidateTable({
  task,
  tasks,
  candidates,
  ranked,
  chosen,
  mode,
  draft,
  query,
  onSelectPerson,
}: {
  task: Task
  tasks: Task[]
  candidates: Person[]
  ranked: ReturnType<typeof recommendations>
  chosen: Choice | undefined
  mode: "manual" | "recommended"
  draft: Draft
  query: string
  onSelectPerson: (id: string) => void
}) {
  const { skillName } = useSkillCatalog()
  return (
    <div className="overflow-x-auto">
      <table className="w-full min-w-[470px] text-left text-sm">
        <thead>
          <tr className="table-heading table-heading-strong">
            <th className="py-3 pl-5">Ứng viên</th>
            <th className="px-2">Tải công việc</th>
            {mode === "recommended" && <th>Điểm mẫu</th>}
            <th className="pr-5 text-right">Lựa chọn</th>
          </tr>
        </thead>
        <tbody>
          {candidates.map((person) => {
            const recommendation = ranked.find((r) => r.person.id === person.id)
            const checked = chosen?.personId === person.id && chosen.mode === mode
            const gaps = skillGaps(task, person)
            return (
              <tr
                key={person.id}
                className={cn("border-t border-border align-top", checked && "bg-accent/50")}
              >
                <td className="py-4 pl-5 pr-2">
                  <div className="flex items-start gap-2.5">
                    <Avatar initials={person.initials} />
                    <div>
                      <p className="font-medium">{person.name}</p>
                      <p className="mt-0.5 text-xs text-muted-foreground">
                        {person.position} · {person.seniority}
                      </p>
                      <details key={`${task.id}-${person.id}`} className="mt-2 text-xs">
                        <summary className="field-label cursor-pointer underline underline-offset-4">
                          Đối chiếu kỹ năng
                          {gaps.length > 0 ? " · còn thiếu" : ""}
                        </summary>
                        <div className="mt-3 space-y-2">
                          {task.requirements.map((r) => (
                            <p
                              key={r.skill}
                              className={cn(
                                "leading-5",
                                (person.skills[r.skill] ?? 0) < r.level && r.mandatory
                                  ? "text-warning"
                                  : "text-muted-foreground",
                              )}
                            >
                              <strong className="text-sm font-medium text-foreground">
                                {skillName(r.skill)}
                              </strong>
                              {r.mandatory ? " · Bắt buộc" : " · Bổ trợ"}
                              <span className="block text-muted-foreground">
                                <span className="font-medium text-foreground">Yêu cầu:</span>{" "}
                                {levels[r.level]}
                              </span>
                              <span className="block">
                                <span className="font-medium text-foreground">Hiện có:</span>{" "}
                                {levels[person.skills[r.skill] ?? 0] || "Chưa ghi nhận"}
                              </span>
                            </p>
                          ))}
                        </div>
                      </details>
                    </div>
                  </div>
                </td>
                <td className="px-2 py-4 whitespace-nowrap tabular-nums">
                  <span className="font-medium">Hiện tại {loadFor(person.id, tasks)} giờ</span>
                  <p className="mt-1 text-xs text-muted-foreground">Xuyên dự án</p>
                  <p className="mt-2 text-xs text-muted-foreground">
                    Nháp +{addedLoad(person.id, draft, tasks)} giờ
                  </p>
                  <p className="mt-1 text-xs">
                    Sau nháp {loadFor(person.id, tasks) + addedLoad(person.id, draft, tasks)} giờ
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
                    onClick={() => onSelectPerson(person.id)}
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
          title={query.trim() ? "Không tìm thấy ứng viên" : "Không có ứng viên phù hợp"}
          detail={
            query.trim()
              ? "Thử tên khác hoặc xóa từ khóa tìm kiếm."
              : mode === "recommended"
                ? "Thử tra cứu thủ công hoặc bỏ công việc này khỏi bản nháp."
                : "Thử tìm bằng tên khác."
          }
        />
      )}
    </div>
  )
}
