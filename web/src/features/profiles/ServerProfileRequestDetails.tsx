import { profileFields, profileFieldLabels } from "@/models/profile"
import { cn } from "@/lib/utils"
import type { RequestDetail } from "./profile-requests-api"

export const serverProfileRequestLabels = {
  PENDING: "Chờ xử lý",
  APPROVED: "Đã chấp thuận",
  REJECTED: "Đã từ chối",
} as const

export function ServerProfileRequestBadge({ status }: { status: RequestDetail["status"] }) {
  let statusClassName = "bg-accent text-muted-foreground"
  if (status === "APPROVED") {
    statusClassName = "bg-success-soft text-success"
  } else if (status === "PENDING") {
    statusClassName = "bg-warning-soft text-warning"
  }
  return (
    <span className={cn("inline-flex rounded-md px-2 py-1 text-xs font-medium", statusClassName)}>
      {serverProfileRequestLabels[status]}
    </span>
  )
}

export function ServerProfileRequestDetails({ request }: { request: RequestDetail }) {
  return (
    <div className="space-y-4 text-sm">
      <div className="flex flex-wrap items-center gap-3">
        <ServerProfileRequestBadge status={request.status} />
        <span className="text-muted-foreground">
          Gửi lúc {new Date(request.requested_at).toLocaleString("vi-VN")}
        </span>
      </div>
      <p className="leading-6 text-muted-foreground">
        Nội dung đề nghị được lưu riêng. Giá trị hiện tại được máy chủ đọc khi tải chi tiết này.
      </p>
      <div className="overflow-x-auto rounded-lg border border-border">
        <table className="w-full min-w-[480px] text-left text-sm">
          <thead className="table-heading">
            <tr>
              {["Thông tin", "Trước đề nghị", "Đề nghị", "Hiện tại"].map((label) => (
                <th key={label} scope="col" className="px-3 py-3">
                  {label}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {profileFields
              .filter((field) => field in request.proposed_values)
              .map((field) => (
                <tr key={field} className="border-t border-border align-top">
                  <th scope="row" className="px-3 py-3 font-medium">
                    {profileFieldLabels[field]}
                  </th>
                  <td className="max-w-48 break-words whitespace-pre-wrap px-3 py-3 text-muted-foreground">
                    {request.before_values[field] ?? "Chưa cập nhật"}
                  </td>
                  <td className="max-w-48 break-words whitespace-pre-wrap px-3 py-3">
                    {request.proposed_values[field] ?? "Chưa cập nhật"}
                  </td>
                  <td
                    className={cn(
                      "max-w-48 break-words whitespace-pre-wrap px-3 py-3",
                      request.status === "PENDING" &&
                        request.current_values[field] !== request.before_values[field] &&
                        "text-warning",
                    )}
                  >
                    {request.current_values[field] ?? "Chưa cập nhật"}
                  </td>
                </tr>
              ))}
          </tbody>
        </table>
      </div>
      {request.explanation && (
        <p className="whitespace-pre-wrap break-words leading-6">
          <span className="font-medium">Ghi chú của người gửi: </span>
          {request.explanation}
        </p>
      )}
      {request.resolved_at && (
        <div className="rounded-md bg-accent p-3 leading-6">
          <p>Đã xử lý lúc {new Date(request.resolved_at).toLocaleString("vi-VN")}</p>
          {request.decision_reason && (
            <p className="whitespace-pre-wrap break-words">
              <span className="font-medium">Lý do từ chối: </span>
              {request.decision_reason}
            </p>
          )}
        </div>
      )}
    </div>
  )
}
