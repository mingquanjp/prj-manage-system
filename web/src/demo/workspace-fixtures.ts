import { initialTasks } from "./fixtures.ts"
import { emptyWorkState, rememberContributions, type Submission } from "./effort.ts"

// Lịch sử minh họa rõ nguồn, chỉ khởi tạo/reset prototype; không chạy nghiệp vụ hoặc gửi thông báo.
export function createDemoWorkState() {
  const submissions: Submission[] = [
    {
      id: "submission-t5-initial",
      taskId: "t5",
      assignmentId: "assignment-t5-initial",
      employeeId: "an",
      submittedAt: "2026-10-17T16:00:00+07:00",
      summary: "Dữ liệu mẫu: sơ đồ lưu trữ sản phẩm/kho và mô tả các ràng buộc đã hoàn thành.",
      links: [],
      logSnapshot: { schemaVersion: 2, logs: [] },
      review: {
        result: "ACCEPTED",
        reviewerId: "khoa",
        at: "2026-10-18T10:00:00+07:00",
        comment: "Dữ liệu mẫu: quan hệ, khóa duy nhất và ràng buộc bắt buộc đáp ứng tiêu chí.",
      },
    },
    {
      id: "submission-t7-initial",
      taskId: "t7",
      assignmentId: "assignment-t7-initial",
      employeeId: "cuong",
      submittedAt: "2026-10-18T16:00:00+07:00",
      summary:
        "Dữ liệu mẫu: đã bổ sung tìm kiếm/lọc danh sách sản phẩm, giữ bộ lọc khi mở chi tiết và hiển thị kết quả rỗng.",
      links: [],
      logSnapshot: { schemaVersion: 2, logs: [] },
    },
  ]
  return rememberContributions({ ...emptyWorkState(), submissions }, initialTasks)
}
