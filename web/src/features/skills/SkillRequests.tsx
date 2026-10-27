import { useState } from "react"
import { EmptyState, SearchInput } from "@/components/shared"
import { usePeople } from "@/components/people-directory"
import { useSkillCatalog } from "@/components/skill-catalog"
import { isDirector } from "@/demo/identity"
import { searchKey } from "@/demo/personnel"
import { cn } from "@/lib/utils"
import {
  skillOperationLabels,
  skillRequestLabels,
  type EmployeeSkillsState,
  type SkillRequest,
  type SkillDecision,
} from "@/models/employee-skills"
import { SkillRequestBadge } from "./SkillRequestDetails"
import { SkillReviewDialog } from "./SkillReviewDialog"

export function SkillRequests({
  actorId,
  employees,
  state,
  onResolve,
}: {
  actorId: string
  employees: { personId: string; employeeCode: string }[]
  state: EmployeeSkillsState
  onResolve: (decision: SkillDecision) => void
}) {
  const { getPerson } = usePeople()
  const { skillName } = useSkillCatalog()
  const [filter, setFilter] = useState("PENDING")
  const [query, setQuery] = useState("")
  // Giữ bản đã đọc để không tự duyệt nội dung vừa được người khác bổ sung.
  const [selected, setSelected] = useState<SkillRequest>()
  if (!isDirector(actorId))
    return (
      <EmptyState
        title="Bạn không có quyền xét duyệt kỹ năng"
        detail="Chức năng này dành cho Giám đốc; HR và trách nhiệm PM/PTCM không cấp quyền xét duyệt."
      />
    )
  const requests = state.requests
    .filter(
      (r) =>
        (filter === "all" || r.status === filter) &&
        searchKey(`${getPerson(r.employeeId)?.name} ${skillName(r.skillId)}`).includes(
          searchKey(query),
        ),
    )
    .reverse()
  return (
    <div className="w-full">
      <header className="mb-6">
        <p className="eyebrow">QUẢN TRỊ · GIÁM ĐỐC</p>
        <h1 className="mt-2 text-2xl font-semibold text-balance">Xét duyệt kỹ năng</h1>
        <p className="mt-2 text-pretty text-sm text-muted-foreground">
          Đối chiếu đề nghị thủ công và căn cứ trước khi công nhận năng lực.
        </p>
      </header>
      <div className="mb-5 flex flex-wrap items-end justify-between gap-4">
        <div className="flex flex-wrap gap-1" role="group" aria-label="Lọc đề nghị kỹ năng">
          {[
            ["PENDING", skillRequestLabels.PENDING],
            ["APPROVED", skillRequestLabels.APPROVED],
            ["REJECTED", skillRequestLabels.REJECTED],
            ["WITHDRAWN", skillRequestLabels.WITHDRAWN],
            ["all", "Tất cả"],
          ].map(([value, label]) => (
            <button
              key={value}
              type="button"
              className={cn("filter-chip", filter === value && "filter-active")}
              aria-pressed={filter === value}
              onClick={() => setFilter(value)}
            >
              {label}
            </button>
          ))}
        </div>
        <div className="w-full sm:w-72">
          <SearchInput
            value={query}
            onChange={setQuery}
            placeholder="Tìm nhân viên hoặc kỹ năng…"
          />
        </div>
      </div>
      <div className="overflow-x-auto rounded-xl border border-border bg-white">
        <table className="w-full min-w-[620px] text-left text-sm">
          <thead className="table-heading">
            <tr>
              <th scope="col" className="px-4 py-3">
                Nhân viên
              </th>
              <th scope="col" className="px-4 py-3">
                Kỹ năng / đề nghị
              </th>
              <th scope="col" className="px-4 py-3">
                Ngày gửi
              </th>
              <th scope="col" className="px-4 py-3">
                Trạng thái
              </th>
            </tr>
          </thead>
          <tbody>
            {requests.map((r) => (
              <tr key={r.id} className="border-t border-border hover:bg-accent/30">
                <th scope="row" className="px-4 py-3 text-left font-medium">
                  <button
                    type="button"
                    className="text-left hover:underline"
                    onClick={() => setSelected(structuredClone(r))}
                    aria-label={`Xem đề nghị ${skillName(r.skillId)} của ${getPerson(r.employeeId)?.name}`}
                  >
                    {getPerson(r.employeeId)?.name}
                  </button>
                </th>
                <td className="px-4 py-3">
                  {skillName(r.skillId)}
                  <p className="mt-1 text-xs text-muted-foreground">
                    {skillOperationLabels[r.operation]}
                  </p>
                </td>
                <td className="px-4 py-3 text-xs tabular-nums">
                  {new Date(r.requestedAt).toLocaleString("vi-VN")}
                </td>
                <td className="px-4 py-3">
                  <SkillRequestBadge status={r.status} />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {!requests.length && (
          <EmptyState
            title="Chưa có đề nghị trong mục này"
            detail="Thử thay đổi bộ lọc hoặc từ khóa. Đề nghị mới sẽ xuất hiện khi nhân viên gửi."
          />
        )}
      </div>
      <p role="status" className="mt-3 text-xs tabular-nums text-muted-foreground">
        {requests.length} đề nghị hiển thị
      </p>
      {selected && (
        <SkillReviewDialog
          request={selected}
          state={state}
          actorId={actorId}
          employeeCode={employees.find((e) => e.personId === selected.employeeId)?.employeeCode}
          employeeActive={Boolean(getPerson(selected.employeeId)?.active)}
          onResolve={onResolve}
          onClose={() => setSelected(undefined)}
        />
      )}
    </div>
  )
}
