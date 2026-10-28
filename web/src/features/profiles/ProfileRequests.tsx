import { useSearchParams } from "react-router"
import { useIdentity } from "@/components/identity"
import { useState } from "react"
import { EmptyState, SearchInput } from "@/components/shared"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { usePeople } from "@/components/people-directory"
import { canReviewProfiles, describeProfileChanges, profileConflict } from "@/demo/profiles"
import { searchKey } from "@/demo/personnel"
import type { ProfileChangeRequest, ProfileState } from "@/models/profile"
import { ProfileRequestBadge, ProfileRequestDetails } from "./ProfileRequestDetails"

export function ProfileRequests({
  actorId,
  state,
  onResolve,
}: {
  actorId: string
  state: ProfileState
  onResolve: (requestId: string, decision: "APPROVED" | "REJECTED", reason: string) => void
}) {
  const [params] = useSearchParams()
  const initialRequest = state.requests.find((r) => r.id === params.get("request"))
  const identity = useIdentity()

  const { getPerson } = usePeople()
  const [filter, setFilter] = useState(initialRequest ? "all" : "PENDING")
  const [query, setQuery] = useState("")
  const [selectedId, setSelectedId] = useState<string | undefined>(initialRequest?.id)
  const [decision, setDecision] = useState<"APPROVED" | "REJECTED" | null>(null)
  const [reason, setReason] = useState("")
  const [error, setError] = useState("")
  const [discard, setDiscard] = useState(false)
  if (!canReviewProfiles(actorId, identity))
    return (
      <EmptyState
        title="Bạn không có quyền xử lý yêu cầu"
        detail="Chức năng này dành cho HR. Trách nhiệm PM/PTCM hoặc vai trò Giám đốc không cấp quyền duyệt thông tin cá nhân."
      />
    )
  const selected = state.requests.find((entry) => entry.id === selectedId)
  const current = state.profiles.find((profile) => profile.personId === selected?.employeeId)
  const conflict = Boolean(selected && current && profileConflict(current, selected))
  const filtered = state.requests
    .filter(
      (request) =>
        (filter === "all" || request.status === filter) &&
        searchKey(
          `${getPerson(request.employeeId)?.name} ${describeProfileChanges(request)}`,
        ).includes(searchKey(query)),
    )
    .reverse()
  const close = () => {
    setSelectedId(undefined)
    setDecision(null)
    setReason("")
    setError("")
    setDiscard(false)
  }
  const requestClose = () => (reason.trim() ? setDiscard(true) : close())
  const choose = (request: ProfileChangeRequest) => {
    setSelectedId(request.id)
    setDecision(null)
    setReason("")
    setError("")
  }
  return (
    <>
      <div className="mb-7">
        <p className="eyebrow">QUẢN TRỊ · HR</p>
        <h1 className="mt-2 text-2xl font-semibold text-balance">Yêu cầu thông tin cá nhân</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          Đối chiếu đề nghị và thông tin hiện tại trước khi cập nhật hồ sơ.
        </p>
      </div>
      <div className="mb-5 flex flex-wrap items-end justify-between gap-4">
        <div className="flex flex-wrap gap-1" aria-label="Lọc yêu cầu hồ sơ">
          {[
            ["PENDING", "Chờ xử lý"],
            ["APPROVED", "Đã chấp thuận"],
            ["REJECTED", "Đã từ chối"],
            ["all", "Tất cả"],
          ].map(([value, label]) => (
            <button
              type="button"
              key={value}
              className={`filter-chip ${filter === value ? "filter-active" : ""}`}
              aria-pressed={filter === value}
              onClick={() => setFilter(value)}
            >
              {label}
            </button>
          ))}
        </div>
        <div className="w-full sm:w-80">
          <SearchInput
            value={query}
            onChange={setQuery}
            placeholder="Tìm nhân viên hoặc thông tin thay đổi…"
          />
        </div>
      </div>
      <div className="overflow-x-auto rounded-xl border border-border bg-white">
        <table className="w-full min-w-[640px] text-left text-sm">
          <thead className="table-heading">
            <tr>
              <th scope="col" className="px-5 py-3">
                Nhân viên
              </th>
              <th scope="col" className="px-5 py-3">
                Thông tin đề nghị sửa
              </th>
              <th scope="col" className="px-5 py-3">
                Ngày gửi
              </th>
              <th scope="col" className="px-5 py-3">
                Trạng thái
              </th>
            </tr>
          </thead>
          <tbody>
            {filtered.map((request) => (
              <tr key={request.id} className="border-t border-border hover:bg-accent/30">
                <td className="px-5 py-4">
                  <button
                    className="font-medium text-left hover:underline"
                    type="button"
                    onClick={() => choose(request)}
                    aria-label={`Xem đề nghị của ${getPerson(request.employeeId)?.name}`}
                  >
                    {getPerson(request.employeeId)?.name}
                  </button>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {
                      state.profiles.find((profile) => profile.personId === request.employeeId)
                        ?.employeeCode
                    }
                  </p>
                </td>
                <td className="px-5 py-4">{describeProfileChanges(request)}</td>
                <td className="px-5 py-4 text-xs">
                  {new Date(request.requestedAt).toLocaleString("vi-VN")}
                </td>
                <td className="px-5 py-4">
                  <ProfileRequestBadge status={request.status} />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {!filtered.length && (
          <EmptyState
            title={query ? "Không tìm thấy yêu cầu" : "Chưa có yêu cầu trong mục này"}
            detail="Thử thay đổi bộ lọc hoặc từ khóa. Yêu cầu mới sẽ xuất hiện khi nhân viên gửi đề nghị."
          />
        )}
      </div>
      <p role="status" className="mt-3 text-xs text-muted-foreground">
        {filtered.length} yêu cầu hiển thị
      </p>
      <Dialog
        side
        open={Boolean(selected)}
        onOpenChange={(open) => {
          if (!open) requestClose()
        }}
        title={
          selected ? `Đề nghị của ${getPerson(selected.employeeId)?.name}` : "Chi tiết đề nghị"
        }
        description="HR xử lý thông tin cá nhân; không thay đổi vai trò, trách nhiệm dự án hoặc kỹ năng chính thức."
      >
        {selected && (
          <>
            <div className="mb-4 flex flex-wrap items-center gap-3">
              <ProfileRequestBadge status={selected.status} />
              <span className="text-xs text-muted-foreground">
                {new Date(selected.requestedAt).toLocaleString("vi-VN")}
              </span>
            </div>
            <ProfileRequestDetails request={selected} current={current} />
            {selected.status === "PENDING" && (
              <>
                {conflict && (
                  <p
                    role="alert"
                    className="mt-5 rounded-md bg-warning-soft p-3 text-sm leading-6 text-warning"
                  >
                    Thông tin nền đã thay đổi. Không thể chấp thuận yêu cầu cũ; bạn có thể từ chối
                    để nhân viên kiểm tra và gửi đề nghị mới.
                  </p>
                )}
                {!decision ? (
                  <div className="mt-6 flex flex-wrap justify-end gap-2 border-t border-border pt-5">
                    <Button variant="outline" onClick={() => setDecision("REJECTED")}>
                      Từ chối
                    </Button>
                    <Button disabled={conflict} onClick={() => setDecision("APPROVED")}>
                      Chấp thuận
                    </Button>
                  </div>
                ) : (
                  <form
                    className="mt-6 space-y-4 border-t border-border pt-5"
                    onSubmit={(event) => {
                      event.preventDefault()
                      try {
                        onResolve(selected.id, decision, reason)
                        setDecision(null)
                        setReason("")
                        setError("")
                      } catch (err) {
                        setError(err instanceof Error ? err.message : "Chưa thể xử lý yêu cầu.")
                      }
                    }}
                  >
                    <h3 className="section-title">
                      {decision === "APPROVED" ? "Xác nhận chấp thuận" : "Xác nhận từ chối"}
                    </h3>
                    <p className="text-sm leading-6 text-muted-foreground">
                      {decision === "APPROVED"
                        ? "Chỉ các trường đề nghị sẽ được cập nhật. Thông tin nền được kiểm tra lại khi xác nhận."
                        : "Hồ sơ chính thức sẽ giữ nguyên. Nhân viên xem được lý do từ chối trong lịch sử."}
                    </p>
                    <label className="block">
                      <span className="field-label">
                        {decision === "REJECTED" ? "Lý do từ chối *" : "Ghi chú xử lý (tùy chọn)"}
                      </span>
                      <textarea
                        className="field workspace-field mt-2 min-h-24 w-full"
                        value={reason}
                        onChange={(event) => setReason(event.target.value)}
                        autoFocus
                      />
                    </label>
                    {error && (
                      <p role="alert" className="text-sm text-warning">
                        {error}
                      </p>
                    )}
                    <div className="flex flex-wrap justify-end gap-2">
                      <Button
                        variant="outline"
                        onClick={() => {
                          setDecision(null)
                          setError("")
                        }}
                      >
                        Quay lại
                      </Button>
                      <Button type="submit" disabled={decision === "APPROVED" && conflict}>
                        {decision === "APPROVED" ? "Xác nhận chấp thuận" : "Xác nhận từ chối"}
                      </Button>
                    </div>
                  </form>
                )}
              </>
            )}
          </>
        )}
      </Dialog>
      <Dialog
        open={discard}
        onOpenChange={setDiscard}
        title="Bỏ ghi chú chưa lưu?"
        description="Yêu cầu chưa được xử lý và hồ sơ chính thức giữ nguyên."
      >
        <div className="flex flex-wrap justify-end gap-2">
          <Button variant="outline" onClick={() => setDiscard(false)}>
            Tiếp tục xử lý
          </Button>
          <Button onClick={close}>Bỏ ghi chú</Button>
        </div>
      </Dialog>
    </>
  )
}
