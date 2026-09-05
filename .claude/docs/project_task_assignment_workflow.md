# Luồng dự án → task → phân công → thực hiện — V1

Đây là đặc tả đích cập nhật từ UC001–UC004, không giữ số bước cũ ở các luồng đã thay đổi.
Bảng trạng thái chuẩn ở domain_rules.md; quyết định nguồn ở [baseline](../../docs/reviews/system-design-review-v1.md).

## UC001 — Khởi tạo và vòng đời dự án

PM nhập mã/tên, start_date/planned_end_date, priority; mô tả/ghi chú tùy chọn.
Kiểm tra mã unique, ngày đúng thứ tự; tạo PLANNING, PM hiện tại là owner, lưu history.
Hủy form không lưu. Sai định dạng 400, trùng mã/conflict 409.

Bắt đầu khi có task ASSIGNED. Nghiệm thu project khi có ≥1 DONE và phần còn lại DONE/CANCELLED.
PM đóng project COMPLETED: chuyển project CLOSED và mọi DONE thành CLOSED cùng transaction.
Hủy project đang PLANNING/IN_PROGRESS: bắt buộc lý do, đóng phiên và hủy task mở, giữ DONE/lịch sử.
ADMIN có luồng bàn giao PM riêng, không được phân công thay PM.

## UC002 — Nhóm và task

Chỉ project PLANNING/IN_PROGRESS. Tạo group tên unique trong project.
Task có group cùng project, mã unique/project, tên, estimate >0, ≥1 requirement hợp lệ.
Requirements unique skill, weight nguyên 1–5, level 1–4; skill active khi sử dụng mới.
Ngày task tùy chọn nhưng nhập đủ cặp, trong ngày project; create task+requirements atomically.
Không parent task/task_type, group không được giao người.

Estimate/requirements khóa sau lần phân công đầu; không đổi scope bằng chỉnh mô tả.
Thu hồi không mở khóa spec. Chỉ sửa làm rõ có audit; scope mới tạo task mới.
Không tạo/sửa khi project COMPLETED/CLOSED/CANCELLED.

## UC003 — Xác nhận phân công

1. PM chọn danh sách PENDING_ASSIGNMENT cùng project, bỏ task không đủ điều kiện.
2. Chọn gợi ý UC003.1 hoặc tra cứu UC003.2; kết quả vào draft trên UI.
3. Hiển thị người chọn, cảnh báo kỹ năng/tải và tổng tải sau cả draft.
4. Confirm gửi expected_version, idempotency key; không gửi điểm như dữ liệu đáng tin.
5. Transaction khóa/kiểm tra lại owner, project/task, người được chọn, spec/skill và run.
6. Tất cả hợp lệ mới ghi assignment, task ASSIGNED/current assignee, membership, history/event, notification và kết quả idempotency.
7. Commit trả kết quả toàn batch.

Một dòng sai hoặc lỗi lưu → rollback toàn bộ; trả task/index/code lỗi để sửa draft.
Không partial success, không lưu “các dòng hợp lệ” trước. Cùng key/payload retry trả kết quả đã commit, khác payload 409.
Chạy gợi ý không phải phân công; bỏ draft không tạo assignment.

## UC003.2 — Manual và ngoại lệ

Tra cứu theo kỹ năng/phòng ban/vị trí/kinh nghiệm/trạng thái/tham gia dự án.
Chỉ EMPLOYEE active được chọn; hồ sơ chính thức, workload hiện hành, thông tin tham gia giới hạn theo quyền.
Thiếu/dưới mức mandatory được cảnh báo và bắt buộc lý do override.
Không override user locked/employee inactive/sai role. Tải cao cảnh báo, không kết luận vi phạm lịch.
Chuyển manual sau run phải lưu run_id và lý do; không gán điểm gợi ý cho người ngoài run.

## Thay người/thu hồi

Chỉ ASSIGNED/IN_PROGRESS/PAUSED; PENDING_REVIEW phải trả lại trước.
Reassign có lý do: đóng phiên cũ, giữ giờ người cũ, append assignment mới, chuyển ASSIGNED; người mới tự start.
Unassign có lý do: đóng phiên, thêm event, chuyển PENDING_ASSIGNMENT, xóa current assignee, giữ spec đóng băng.
Mọi thay đổi + notification cùng transaction; người cũ mất quyền task hiện tại, chỉ còn lịch sử đóng góp của mình.
Giao lại bằng gợi ý tạo run mới và đi qua confirm cho task chờ giao; reassign trực tiếp dùng manual với cảnh báo/lý do.

## UC004 — Thực hiện và thời gian

ASSIGNED/PAUSED → IN_PROGRESS: current employee, project IN_PROGRESS, chưa có task chạy khác, có lịch snapshot.
IN_PROGRESS → PAUSED hoặc PENDING_REVIEW: đóng phiên và cộng giây hợp lệ.
PM trả review về ASSIGNED có lý do; không tự mở đồng hồ. PM duyệt DONE giữ accepted_at.
DONE chỉ thành CLOSED khi project đóng; phát sinh sau nghiệm thu tạo task mới.

Phiên gắn assignment/employee, thời gian server, không nhập giờ tay.
Lịch Asia/Ho_Chi_Minh thứ Hai–Sáu, 08–12/13–17, trừ ngày nghỉ; pin version lịch khi start.
Đổi trạng thái, log, history, tổng giờ cùng transaction; retry không cộng trùng.
20/05/2026 15:00 → 21/05/2026 10:30 = 4,5 giờ trong lịch thường.
Đầu ra: start/end, giờ phiên, tổng giờ, chênh lệch estimate; đây là giờ theo workflow, không chấm công xác minh.
