import { getPerson } from "@/demo/model"
import { formatMinutes, type Submission } from "@/demo/effort"

// Cùng một cách đọc kết quả ở cá nhân và dự án; caller giới hạn dữ liệu theo tư cách.
export function SubmissionHistory({
  submissions,
  showLogs = true,
}: {
  submissions: Submission[]
  showLogs?: boolean
}) {
  return (
    <ol className="mt-4 space-y-4" aria-label="Lịch sử nộp và nghiệm thu">
      {[...submissions].reverse().map((s, index) => (
        <li key={s.id} className="rounded-lg border border-border p-3 text-sm">
          <div className="flex flex-wrap justify-between gap-2">
            <strong className="field-label">Lần nộp {submissions.length - index}</strong>
            <span className="text-xs text-muted-foreground">
              {new Date(s.submittedAt).toLocaleString("vi-VN", {
                timeZone: "Asia/Ho_Chi_Minh",
              })}
            </span>
          </div>
          <p className="mt-1 text-xs text-muted-foreground">
            Người nộp: {getPerson(s.employeeId)?.name ?? "Người thực hiện"}
          </p>
          <p className="mt-2 whitespace-pre-wrap break-words leading-6">{s.summary}</p>
          {!!s.links.length && (
            <ul className="mt-2 space-y-1">
              {s.links.map((url, i) => (
                <li key={i}>
                  <a
                    href={url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="break-all underline"
                  >
                    {url}
                  </a>
                </li>
              ))}
            </ul>
          )}
          <p className="mt-3 rounded-md bg-accent px-2 py-1 text-xs">
            {s.review
              ? s.review.result === "ACCEPTED"
                ? "Đã nghiệm thu"
                : "Yêu cầu chỉnh sửa"
              : "Chưa có kết luận nghiệm thu"}
          </p>
          {s.review && (
            <p className="mt-2 whitespace-pre-wrap break-words text-xs leading-5">
              {getPerson(s.review.reviewerId)?.name ?? "PTCM"} ·{" "}
              {new Date(s.review.at).toLocaleString("vi-VN", {
                timeZone: "Asia/Ho_Chi_Minh",
              })}
              <br />
              {s.review.comment || "Đã chấp nhận đầu ra theo tiêu chí nghiệm thu."}
            </p>
          )}
          {showLogs && (
            <details className="mt-3 text-xs">
              <summary className="field-label cursor-pointer">
                Nhật ký tại lần nộp ·{" "}
                {s.logSnapshot.logs.length
                  ? formatMinutes(s.logSnapshot.logs.reduce((sum, l) => sum + l.minutes, 0))
                  : "Chưa có thời gian ghi nhận"}
              </summary>
              <p className="mt-2 text-muted-foreground">
                Bản chụp chỉ đọc của lượt giao này; không khẳng định đã khai đủ thời gian thực tế.
              </p>
              <ul className="mt-2 space-y-2">
                {s.logSnapshot.logs.map((l) => (
                  <li key={l.id} className="whitespace-pre-wrap break-words leading-5">
                    {l.workDate} · {formatMinutes(l.minutes)}
                    <br />
                    {l.description}
                    {l.overLimitReason && <p>Lý do vượt 8 giờ: {l.overLimitReason}</p>}
                  </li>
                ))}
              </ul>
            </details>
          )}
        </li>
      ))}
    </ol>
  )
}
