import { usePeople } from "@/components/people-directory"
import type { EmployeeProfile, ProfileChangeRequest } from "@/models/profile"
import { profileFields, profileFieldLabels } from "@/models/profile"
import { profileRequestLabels } from "@/demo/profiles"
import { cn } from "@/lib/utils"

export function ProfileRequestBadge({ status }: { status: ProfileChangeRequest["status"] }) {
  return (
    <span
      className={cn(
        "inline-flex rounded-md px-2 py-1 text-xs font-medium",
        status === "APPROVED"
          ? "bg-success-soft text-success"
          : status === "PENDING"
            ? "bg-warning-soft text-warning"
            : "bg-accent text-muted-foreground",
      )}
    >
      {profileRequestLabels[status]}
    </span>
  )
}

export function ProfileRequestDetails({
  request,
  current,
}: {
  request: ProfileChangeRequest
  current?: EmployeeProfile
}) {
  const { getPerson } = usePeople()
  return (
    <div className="space-y-4 text-sm">
      <div className="overflow-x-auto rounded-lg border border-border">
        <table className="w-full min-w-[360px] text-left text-sm">
          <thead className="table-heading">
            <tr>
              <th scope="col" className="px-3 py-3">
                Thông tin
              </th>
              <th scope="col" className="px-3 py-3">
                Trước đề nghị
              </th>
              <th scope="col" className="px-3 py-3">
                Đề nghị
              </th>
              {current && (
                <th scope="col" className="px-3 py-3">
                  Hiện tại
                </th>
              )}
            </tr>
          </thead>
          <tbody>
            {profileFields
              .filter((field) => field in request.proposed)
              .map((field) => (
                <tr key={field} className="border-t border-border align-top">
                  <th scope="row" className="px-3 py-3 font-medium">
                    {profileFieldLabels[field]}
                  </th>
                  <td className="max-w-48 break-words whitespace-pre-wrap px-3 py-3 text-muted-foreground">
                    {request.before[field] ?? "Chưa cập nhật"}
                  </td>
                  <td className="max-w-48 break-words whitespace-pre-wrap px-3 py-3">
                    {request.proposed[field] ?? "Chưa cập nhật"}
                  </td>
                  {current && (
                    <td
                      className={cn(
                        "max-w-48 break-words whitespace-pre-wrap px-3 py-3",
                        request.status === "PENDING" &&
                          current[field] !== request.before[field] &&
                          "text-warning",
                      )}
                    >
                      {current[field] ?? "Chưa cập nhật"}
                    </td>
                  )}
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
      {request.resolvedAt && (
        <div className="rounded-md bg-accent p-3 leading-6">
          <p>
            {getPerson(request.resolvedBy)?.name} ·{" "}
            {new Date(request.resolvedAt).toLocaleString("vi-VN")}
          </p>
          {request.decisionReason && (
            <p className="whitespace-pre-wrap break-words">{request.decisionReason}</p>
          )}
        </div>
      )}
    </div>
  )
}
