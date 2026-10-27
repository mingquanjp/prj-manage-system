import { useState } from "react"
import { PlusIcon } from "@radix-ui/react-icons"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { useSkillCatalog } from "@/components/skill-catalog"
import { usePeople } from "@/components/people-directory"
import { levels } from "@/models/workspace-labels"
import {
  skillOperationLabels,
  type EmployeeSkillsState,
  type SkillRequest,
  type SkillRequestInput,
  type SkillSupplement,
} from "@/models/employee-skills"
import type { WorkState } from "@/demo/effort"
import type { Task } from "@/models/workspace"
import { availableSkillEvidence } from "@/demo/employee-skills"
import { SkillRequestForm } from "./SkillRequestForm"
import { SkillRequestBadge, SkillRequestDetails } from "./SkillRequestDetails"

export type PersonalSkillActions = {
  requestSkillChange: (input: SkillRequestInput) => void
  supplementSkillChange: (id: string, version: number, input: SkillSupplement) => void
  withdrawSkillChange: (id: string, version: number, reason: string) => void
}
export function PersonalSkills({
  actorId,
  state,
  work,
  tasks,
  actions,
}: {
  actorId: string
  state: EmployeeSkillsState
  work: WorkState
  tasks: Task[]
  actions: PersonalSkillActions
}) {
  const { skillName } = useSkillCatalog()
  const { getPerson } = usePeople()
  const [editor, setEditor] = useState<{ request?: SkillRequest }>()
  const [withdrawing, setWithdrawing] = useState<SkillRequest>()
  const [reason, setReason] = useState("")
  const [error, setError] = useState("")
  const official = state.official.filter((s) => s.employeeId === actorId)
  const requests = state.requests.filter((r) => r.employeeId === actorId).reverse()
  const available = availableSkillEvidence(work, tasks, actorId)
  return (
    <div className="space-y-7">
      <section>
        <div className="mb-4 flex flex-wrap items-start justify-between gap-3">
          <div>
            <h2 className="section-title">Kỹ năng chính thức</h2>
            <p className="mt-1 text-pretty text-sm text-muted-foreground">
              Chỉ thông tin đã được công nhận mới dùng để đối chiếu phân công.
            </p>
          </div>
          <Button onClick={() => setEditor({})}>
            <PlusIcon /> Đề nghị kỹ năng
          </Button>
        </div>
        <div className="overflow-x-auto rounded-xl border border-border bg-white">
          <table className="w-full min-w-[540px] text-left text-sm">
            <thead className="table-heading">
              <tr>
                <th scope="col" className="px-4 py-3">
                  Kỹ năng
                </th>
                <th scope="col" className="px-4 py-3">
                  Mức được công nhận
                </th>
                <th scope="col" className="px-4 py-3">
                  Số năm kinh nghiệm
                </th>
                <th scope="col" className="px-4 py-3">
                  Công nhận gần nhất
                </th>
              </tr>
            </thead>
            <tbody>
              {official.map((s) => (
                <tr key={s.id} className="border-t border-border">
                  <th scope="row" className="px-4 py-3 font-medium">
                    {skillName(s.skillId)}
                    {s.note && (
                      <p className="mt-1 max-w-xs whitespace-pre-wrap break-words text-xs font-normal text-muted-foreground">
                        {s.note}
                      </p>
                    )}
                  </th>
                  <td className="px-4 py-3">{levels[s.level]}</td>
                  <td className="px-4 py-3 tabular-nums">{s.years ?? "Chưa cập nhật"}</td>
                  <td className="px-4 py-3 text-xs text-muted-foreground">
                    {s.approvedAt ? (
                      <>
                        {getPerson(s.approvedBy)?.name}
                        <br />
                        {new Date(s.approvedAt).toLocaleString("vi-VN")}
                      </>
                    ) : (
                      "Dữ liệu mẫu ban đầu"
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {!official.length && (
            <p className="p-5 text-sm text-muted-foreground">
              Chưa có kỹ năng chính thức. Chọn Đề nghị kỹ năng để gửi thông tin và căn cứ.
            </p>
          )}
        </div>
      </section>
      <section aria-label="Lịch sử đề nghị kỹ năng">
        <div className="mb-4 flex items-center justify-between gap-3">
          <h2 className="section-title">Đề nghị và lịch sử</h2>
          <span className="text-xs tabular-nums text-muted-foreground">
            {requests.length} đề nghị
          </span>
        </div>
        <div className="space-y-3">
          {requests.map((request) => (
            <details key={request.id} className="rounded-lg border border-border bg-white p-4">
              <summary className="cursor-pointer text-sm leading-6">
                <span className="mr-3 font-medium">
                  {skillName(request.skillId)} · {skillOperationLabels[request.operation]}
                </span>
                <SkillRequestBadge status={request.status} />
                <span className="ml-3 text-xs text-muted-foreground">
                  {new Date(request.requestedAt).toLocaleString("vi-VN")}
                </span>
              </summary>
              <div className="mt-5">
                <SkillRequestDetails request={request} />
                {request.status === "PENDING" && (
                  <div className="mt-5 flex flex-wrap justify-end gap-2 border-t border-border pt-4">
                    <Button
                      variant="outline"
                      onClick={() => {
                        setWithdrawing(request)
                        setReason("")
                        setError("")
                      }}
                    >
                      Rút đề nghị
                    </Button>
                    <Button variant="outline" onClick={() => setEditor({ request })}>
                      Bổ sung căn cứ
                    </Button>
                  </div>
                )}
              </div>
            </details>
          ))}
        </div>
        {!requests.length && (
          <p className="border-t border-border pt-4 text-sm text-muted-foreground">
            Chưa có đề nghị kỹ năng. Năng lực chính thức chỉ thay đổi sau khi được Giám đốc chấp
            thuận.
          </p>
        )}
      </section>
      {editor && (
        <SkillRequestForm
          employeeId={actorId}
          state={state}
          available={available}
          request={editor.request}
          onSubmit={actions.requestSkillChange}
          onSupplement={actions.supplementSkillChange}
          onClose={() => setEditor(undefined)}
        />
      )}
      <Dialog
        alert
        open={Boolean(withdrawing)}
        onOpenChange={(open) => {
          if (!open) setWithdrawing(undefined)
        }}
        title="Rút đề nghị kỹ năng?"
        description="Đề nghị sẽ kết thúc và không thể bổ sung nữa. Bạn có thể gửi đề nghị mới khi cần."
      >
        <form
          className="space-y-4"
          onSubmit={(event) => {
            event.preventDefault()
            if (!withdrawing) return
            try {
              actions.withdrawSkillChange(withdrawing.id, withdrawing.version, reason)
              setWithdrawing(undefined)
            } catch (err) {
              setError(err instanceof Error ? err.message : "Chưa thể rút đề nghị.")
            }
          }}
        >
          <label className="block">
            <span className="field-label">Lý do rút *</span>
            <textarea
              autoFocus
              className="field mt-2 min-h-20 w-full"
              value={reason}
              aria-required="true"
              onChange={(event) => setReason(event.target.value)}
            />
          </label>
          {error && (
            <p role="alert" className="text-sm text-warning">
              {error}
            </p>
          )}
          <div className="flex justify-end gap-2">
            <Button variant="outline" onClick={() => setWithdrawing(undefined)}>
              Giữ đề nghị
            </Button>
            <Button type="submit">Xác nhận rút</Button>
          </div>
        </form>
      </Dialog>
    </div>
  )
}
