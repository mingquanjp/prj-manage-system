import { useEffect, useState, useSyncExternalStore } from "react"
import { Link, useLocation, useNavigate } from "react-router"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import type { AuthSession } from "../session/auth-session"
import {
  createHrReviewApi,
  reviewFields,
  validReviewReason,
  trimReviewReason,
  type ReviewDetail,
  type ReviewQuery,
} from "./hr-review-api"
import { createHrReviewResource } from "./hr-review-resource"
import { parseReviewLocation } from "../session/server-navigation"

const labels = { fullname: "Họ và tên", phone: "Số điện thoại", address: "Địa chỉ" }
const statuses = { PENDING: "Chờ xử lý", APPROVED: "Đã chấp thuận", REJECTED: "Đã từ chối" }
const root = "/personnel/profile-requests"
function time(value: string) {
  return new Date(value).toLocaleString("vi-VN")
}
function show(value: string | null | undefined) {
  return value === null ? "Chưa có thông tin" : value === "" ? "Chuỗi rỗng" : value
}
function queryUrl(query: ReviewQuery) {
  const params = new URLSearchParams({
    page: String(query.page),
    page_size: String(query.pageSize),
  })
  if (query.status) params.set("status", query.status)
  else params.set("status", "all")
  return `${root}?${params}`
}
function Comparison({ detail }: { detail: ReviewDetail }) {
  return (
    <>
      <div className="overflow-x-auto rounded-lg border border-border bg-white">
        <table className="w-full min-w-[520px] text-left text-sm">
          <caption className="sr-only">Đối chiếu giá trị khi gửi, đề nghị và hiện tại</caption>
          <thead className="table-heading">
            <tr>
              {["Thông tin", "Khi gửi", "Đề nghị", "Hiện tại"].map((label) => (
                <th key={label} scope="col" className="px-4 py-3">
                  {label}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {reviewFields
              .filter((field) => field in detail.proposed_values)
              .map((field) => (
                <tr key={field} className="border-t border-border align-top">
                  <th scope="row" className="px-4 py-3">
                    {labels[field]}
                  </th>
                  {[detail.before_values, detail.proposed_values, detail.current_values].map(
                    (values, index) => (
                      <td
                        key={index}
                        className="max-w-64 whitespace-pre-wrap break-words px-4 py-3"
                      >
                        {show(values[field])}
                      </td>
                    ),
                  )}
                </tr>
              ))}
          </tbody>
        </table>
      </div>
      <p className="text-xs leading-5 text-muted-foreground">
        “Hiện tại” là giá trị lúc tải chi tiết. Máy chủ kiểm tra lại khi chấp thuận.
      </p>
      {detail.explanation !== null && (
        <p className="whitespace-pre-wrap break-words text-sm">
          <strong>Giải thích của người gửi: </strong>
          {show(detail.explanation)}
        </p>
      )}
      {detail.status !== "PENDING" && (
        <div className="rounded-lg border border-border bg-accent p-4 text-sm leading-6">
          <p>
            {statuses[detail.status]} · {time(detail.resolved_at!)} · Mã người xử lý:{" "}
            {detail.resolved_by}
          </p>
          {detail.decision_reason !== null && (
            <p className="whitespace-pre-wrap break-words">
              <strong>Lý do từ chối: </strong>
              {detail.decision_reason}
            </p>
          )}
          <p>Đề nghị đã kết thúc, không thể xử lý lại.</p>
        </div>
      )}
    </>
  )
}
export function ServerProfileReviews({ session, busy }: { session: AuthSession; busy: boolean }) {
  const location = useLocation()
  const navigate = useNavigate()
  const parsed = parseReviewLocation(location.pathname, location.search)
  const [resource] = useState(() => createHrReviewResource(createHrReviewApi(), session))
  const state = useSyncExternalStore(resource.subscribe, resource.getSnapshot)
  const [decision, setDecision] = useState<"approve" | "reject" | null>(null)
  const [reason, setReason] = useState("")
  const [validation, setValidation] = useState("")
  const routeKey = `${location.pathname}${location.search}`
  useEffect(() => {
    const input = parseReviewLocation(location.pathname, location.search)
    if (input) void resource.load(input.id ?? input.query)
    document.getElementById("review-heading")?.focus()
    return resource.clear
  }, [resource, location.pathname, location.search])
  const reload = () => {
    document.getElementById("review-heading")?.focus()
    if (parsed) void resource.load(parsed.id ?? parsed.query)
  }
  const detail = state.detail
  const stale =
    detail &&
    reviewFields.some(
      (field) =>
        field in detail.before_values &&
        detail.before_values[field] !== detail.current_values[field],
    )
  const locked = busy || state.deciding
  const close = () => {
    if (!state.deciding) {
      setDecision(null)
      setReason("")
      setValidation("")
    }
  }
  const choose = (next: "approve" | "reject") => {
    setDecision(next)
    setReason("")
    setValidation("")
  }
  return (
    <section className="space-y-5" aria-busy={locked || state.phase === "loading"}>
      <header>
        <p className="eyebrow">QUẢN TRỊ · HR</p>
        <h1 id="review-heading" tabIndex={-1} className="mt-2 text-2xl font-semibold">
          {parsed?.id ? "Chi tiết đề nghị hồ sơ" : "Đề nghị hồ sơ"}
        </h1>
        <p className="mt-2 text-sm text-muted-foreground">
          Đối chiếu thông tin trước khi quyết định. Quyền và kết quả được xác nhận từ máy chủ.
        </p>
      </header>
      {!parsed ? (
        <p role="alert">
          Đường dẫn hoặc bộ lọc chưa hợp lệ.{" "}
          <Link to={root} className="underline">
            Về hàng chờ
          </Link>
        </p>
      ) : (
        <>
          {parsed.id ? (
            <Button asChild variant="outline">
              <Link to={root}>Về hàng chờ</Link>
            </Button>
          ) : (
            <div className="flex flex-wrap gap-2" aria-label="Lọc trạng thái đề nghị">
              {(["PENDING", "APPROVED", "REJECTED", "all"] as const).map((status) => (
                <Button
                  key={status}
                  variant={
                    parsed.query.status === status || (!parsed.query.status && status === "all")
                      ? "default"
                      : "outline"
                  }
                  disabled={locked}
                  aria-pressed={
                    parsed.query.status === status || (!parsed.query.status && status === "all")
                  }
                  onClick={() =>
                    navigate(
                      queryUrl({
                        ...parsed.query,
                        status: status === "all" ? undefined : status,
                        page: 1,
                      }),
                    )
                  }
                >
                  {status === "all" ? "Tất cả" : statuses[status]}
                </Button>
              ))}
            </div>
          )}
          {state.message && (
            <p role="alert" className="rounded-lg bg-accent p-4 text-sm leading-6">
              {state.message}
            </p>
          )}
          <Button variant="outline" disabled={locked || state.phase === "loading"} onClick={reload}>
            Tải lại để đối soát
          </Button>
          {state.phase === "loading" ? (
            <p role="status">Đang tải đề nghị từ máy chủ…</p>
          ) : state.page ? (
            <>
              <div className="overflow-x-auto rounded-xl border border-border bg-white">
                <table className="w-full min-w-[620px] text-left text-sm">
                  <thead className="table-heading">
                    <tr>
                      {[
                        "Nhân viên",
                        "Thông tin đề nghị sửa",
                        "Ngày gửi",
                        "Trạng thái",
                        "Chi tiết",
                      ].map((label) => (
                        <th key={label} scope="col" className="px-4 py-3">
                          {label}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {state.page.items.map((item) => (
                      <tr key={item.request_id} className="border-t border-border">
                        <th scope="row" className="px-4 py-3 font-medium">
                          {item.fullname}
                          <span className="block text-xs text-muted-foreground">
                            {item.employee_code}
                          </span>
                        </th>
                        <td className="px-4 py-3">
                          {item.changed_fields.map((field) => labels[field]).join(", ")}
                        </td>
                        <td className="px-4 py-3">{time(item.requested_at)}</td>
                        <td className="px-4 py-3">{statuses[item.status]}</td>
                        <td className="px-4 py-3">
                          <Link
                            className="underline"
                            aria-label={`Xem đề nghị ${item.request_id} của ${item.fullname}`}
                            to={`${root}/${item.request_id}`}
                          >
                            Xem đề nghị
                          </Link>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              {!state.page.items.length && <p>Không có đề nghị ở trang và bộ lọc này.</p>}
              <div className="flex flex-wrap items-center gap-3 text-sm">
                <span>
                  Trang {state.page.page} · Tổng {state.page.total} đề nghị
                </span>
                <Button
                  variant="outline"
                  disabled={locked || parsed.query.page <= 1}
                  onClick={() =>
                    navigate(queryUrl({ ...parsed.query, page: parsed.query.page - 1 }))
                  }
                >
                  Trang trước
                </Button>
                <Button
                  variant="outline"
                  disabled={
                    locked ||
                    parsed.query.page * parsed.query.pageSize >= state.page.total ||
                    parsed.query.page >= 2147483647
                  }
                  onClick={() =>
                    navigate(queryUrl({ ...parsed.query, page: parsed.query.page + 1 }))
                  }
                >
                  Trang sau
                </Button>
              </div>
            </>
          ) : detail ? (
            <>
              <div className="space-y-2 text-sm">
                <h2 className="text-lg font-semibold">
                  {detail.fullname} · {detail.employee_code}
                </h2>
                <p>
                  Đề nghị #{detail.request_id} · {statuses[detail.status]} · Gửi lúc{" "}
                  {time(detail.requested_at)}
                </p>
                <p>Mã người gửi: {detail.requested_by}</p>
              </div>
              <Comparison detail={detail} />
              {detail.status === "PENDING" && (
                <>
                  {stale && (
                    <p role="alert" className="rounded-lg bg-warning-soft p-3 text-sm">
                      Giá trị hiện tại khác lúc gửi. Chấp thuận có thể bị từ chối vì dữ liệu nền đã
                      đổi; bạn có thể từ chối kèm lý do.
                    </p>
                  )}
                  <div className="flex flex-wrap gap-3">
                    <Button disabled={locked || state.mustReload} onClick={() => choose("approve")}>
                      Chấp thuận
                    </Button>
                    <Button
                      variant="outline"
                      disabled={locked || state.mustReload}
                      onClick={() => choose("reject")}
                    >
                      Từ chối
                    </Button>
                  </div>
                </>
              )}
            </>
          ) : null}
          <Dialog
            key={routeKey}
            open={decision !== null}
            onOpenChange={(open) => {
              if (!open) close()
            }}
            onCloseAutoFocus={(event) => {
              event.preventDefault()
              document.getElementById("review-heading")?.focus()
            }}
            title={decision === "approve" ? "Chấp thuận đề nghị?" : "Từ chối đề nghị?"}
            description={
              decision === "approve"
                ? "Toàn bộ thông tin đề nghị sẽ được cập nhật nếu máy chủ xác nhận dữ liệu nền còn khớp."
                : "Hồ sơ chính thức được giữ nguyên. Nhân viên sẽ xem được lý do từ chối."
            }
          >
            <form
              onSubmit={(event) => {
                event.preventDefault()
                if (!decision || locked) return
                if (decision === "reject" && !validReviewReason(reason)) {
                  setValidation(
                    "Lý do phải có từ 1 đến 2.000 ký tự Unicode sau khi bỏ khoảng trắng ở biên.",
                  )
                  return
                }
                void resource.decide(decision, reason).then(() => {
                  setDecision(null)
                  setReason("")
                  setValidation("")
                  document.getElementById("review-heading")?.focus()
                })
              }}
              className="space-y-4"
            >
              {decision === "reject" && (
                <div>
                  <label htmlFor="review-reason" className="text-sm font-medium">
                    Lý do từ chối
                  </label>
                  <textarea
                    id="review-reason"
                    value={reason}
                    disabled={locked}
                    aria-describedby="review-reason-help"
                    aria-invalid={!!validation}
                    onChange={(event) => {
                      setReason(event.target.value)
                      setValidation("")
                    }}
                    className="mt-2 min-h-32 w-full rounded-lg border border-border p-3 text-sm"
                  />
                  <p id="review-reason-help" className="mt-1 text-xs text-muted-foreground">
                    {[...trimReviewReason(reason)].length}/2.000 ký tự Unicode. Không thay đổi nội
                    dung đề nghị.
                  </p>
                </div>
              )}
              {validation && (
                <p role="alert" className="text-sm">
                  {validation}
                </p>
              )}
              <div className="flex flex-wrap justify-end gap-3">
                <Button type="button" variant="outline" disabled={locked} onClick={close}>
                  Hủy
                </Button>
                <Button type="submit" disabled={locked}>
                  {state.deciding ? "Đang xác nhận…" : "Xác nhận"}
                </Button>
              </div>
            </form>
          </Dialog>
        </>
      )}
    </section>
  )
}
