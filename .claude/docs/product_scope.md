# Phạm vi sản phẩm — baseline V1

Đã được người dùng đồng ý ngày 2026-09-05. Đây là thiết kế đích, không phải danh sách chức năng đã implement.
Chi tiết và truy vết D01–D38: [baseline V1](../../docs/reviews/system-design-review-v1.md).
Tài liệu gốc trong `docs/thesis-context/` chỉ đọc; các điểm khác phải được cập nhật vào bản báo cáo làm việc sau.

## 1. Mục tiêu

Một tổ chức, giao diện tiếng Việt: hồ sơ năng lực đã duyệt → yêu cầu task → gợi ý/tra cứu → PM xác nhận → thực hiện và nghiệm thu.
Gợi ý dùng luật và công thức chấm điểm giải thích được; không bắt buộc mô hình học máy.
Không tuyên bố công thức tối ưu khi chưa có thực nghiệm. Tên kỹ thuật `AI` trong enum/file cũ là nhãn tương thích cho nhánh gợi ý, không chứng minh sử dụng ML.

## 2. Tác nhân và chức năng

| Actor | Chức năng và ranh giới |
|---|---|
| ADMIN | Tạo tài khoản kèm hồ sơ, danh mục, duyệt hồ sơ/kỹ năng/CV, khóa tài khoản, chuyển PM có audit; không mặc định đọc toàn bộ nội dung dự án |
| PM | Quản lý dự án của mình, tra cứu năng lực chính thức toàn tổ chức, phân công, nghiệm thu; không duyệt năng lực |
| EMPLOYEE | Hồ sơ của mình, gửi yêu cầu sửa/CV, thực hiện task hiện được giao; là vai trò duy nhất nhận task |

Mỗi tài khoản một vai trò. Project → TaskGroup → Task, không cây task nhiều cấp.
Gồm yêu cầu cập nhật hồ sơ và kỹ năng riêng, vòng đời dự án/task, pause/resume, trả việc, thay người/thu hồi, thông báo nội bộ, tính giờ theo lịch và trích xuất CV có duyệt.

## 3. Ngoài phạm vi

Chứng chỉ riêng, đa tổ chức, PM kiêm người thực thi, task dependency, tối ưu lịch/nguồn lực theo ngày, nghỉ phép, chấm công/lương, overtime, OCR, tự động giao việc, email/push và mở lại task đã nghiệm thu hoặc project terminal.
Tải công việc là tải tồn đọng, không phải khả năng đáp ứng deadline.

## 4. Lộ trình và hiện trạng

| Mốc | Nội dung thiết kế đích |
|---|---|
| M0 | Skeleton và migration 000001 hiện có; chưa có service nghiệp vụ |
| M1 | JWT/refresh/session revocation, CSRF, RBAC/ABAC, tài khoản kèm nhân sự, đổi/reset mật khẩu |
| M2 | Danh mục; yêu cầu hồ sơ/kỹ năng; phê duyệt có version/audit |
| M3 | UC001/UC002, nhóm–task, requirements; start/complete/close/cancel project |
| M4 | UC003/UC003.2, batch atomic, reassign/unassign và retry an toàn |
| M5 | UC003.1, công thức Top-3, snapshot và truy vết quyết định |
| M6 | UC004, pause/resume/review, phiên làm việc không chồng lấn |
| M7 | PDF text/DOCX → trích xuất dựa trên luật → duyệt CV và kỹ năng độc lập |
| M8 | Kiểm thử liên miền, đánh giá công thức, đồng bộ báo cáo và đóng gói |

M1–M6 là lõi; bản đồ án đầy đủ gồm M7–M8. Không coi M7 là “phần học máy bắt buộc”.
Local disk private qua storage interface đã chọn cho V1; thư viện parser cần kiểm tra và duyệt trước khi thêm.
