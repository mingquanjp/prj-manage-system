# Quy tắc nghiệp vụ — baseline V1

Thiết kế đích đã đồng ý 2026-09-05. [Quyết định và truy vết](../../docs/reviews/system-design-review-v1.md).
Schema 000001 chưa đáp ứng toàn bộ; xem `database_schema.md`.

## 1. Bất biến

- R1: Engine chỉ gợi ý; chỉ PM phụ trách dự án được xác nhận phân công.
- R2: Task là đơn vị phân công duy nhất. Group chỉ gom nhóm.
- R3: Tra cứu và engine chỉ dùng kỹ năng chính thức đã duyệt, còn hiệu lực.
- R4: Một batch hoặc ghi toàn bộ hoặc không ghi gì; draft không đặt chỗ.
- R5: Mọi thay đổi trạng thái/người phụ trách/năng lực giữ audit; không API xóa vật lý lịch sử.
- R6: Một EMPLOYEE chỉ có một task IN_PROGRESS và một phiên mở trên toàn hệ thống.
- R7: Quyền, version và điều kiện cuối phải kiểm tra trong transaction có kiểm soát cạnh tranh.

## 2. Tài khoản và năng lực

Tạo user + employee trong một transaction. Cả user và employee phải ACTIVE để sử dụng API nghiệp vụ.
Chỉ role EMPLOYEE active được nhận việc. Không hạ quyền/vô hiệu hóa ADMIN active cuối cùng.
Chặn đổi role nếu còn task mở của EMPLOYEE hoặc dự án mở do PM phụ trách; khóa khẩn cấp vẫn được phép, liệt kê việc cần bàn giao.

`employee_skills` unique người/kỹ năng chỉ lưu chính thức; giữ phiên bản và trạng thái hiệu lực, đầy đủ người/ngày duyệt.
`employee_skill_change_requests`: ADD/UPDATE/REMOVE, MANUAL/CV; PENDING → APPROVED/REJECTED/WITHDRAWN.
Một pending/người/kỹ năng; duyệt kiểm tra version, ghi quyết định và cập nhật chính thức atomically.
Pending không thay giá trị đang dùng. Reject/remove bắt buộc lý do. ADMIN nhập trực tiếp cũng qua service này với audit.

`employee_profile_change_requests` chỉ cho tên/email/điện thoại; không nhận role, trạng thái, phòng ban, chức danh qua yêu cầu cá nhân.
Duyệt kiểm tra version và email unique. Các trường quản trị chỉ ADMIN thay đổi.

Rubric BASIC: có hướng dẫn; INTERMEDIATE: độc lập việc thường gặp; ADVANCED: xử lý phức tạp; EXPERT: thiết kế/định hướng/hỗ trợ người khác.
Không tự suy mức từ số năm hoặc từ tên kỹ năng trong CV.
Skill catalog INACTIVE chặn sử dụng mới, không tự thu hồi năng lực đã duyệt.
Task chưa từng giao phải thay skill inactive trước giao; task từng giao giữ requirements cũ và cảnh báo khi giao lại.

## 3. Dự án và task

Project code unique toàn hệ thống; ngày kết thúc ≥ ngày bắt đầu. PM tạo tự làm chủ.
Mỗi task bắt buộc thuộc một group cùng project, mã unique trong project, estimate > 0, ít nhất một requirement.
Tạo task và requirements cùng transaction. Requirement unique task/skill, weight nguyên 1–5, required level 1–4.
Ngày task hoặc cùng trống hoặc cùng có, đúng thứ tự và trong ngày project; sửa ngày project không làm task ra ngoài.
Không chuyển task sang project khác. Chuyển group trong cùng project có audit, chỉ task chưa kết thúc.

| Project transition | Điều kiện |
|---|---|
| Tạo → PLANNING | PM |
| PLANNING → IN_PROGRESS | PM, có ít nhất một task ASSIGNED |
| IN_PROGRESS → COMPLETED | PM xác nhận, có ≥1 DONE, mọi task còn lại DONE/CANCELLED |
| COMPLETED → CLOSED | PM; DONE → CLOSED cùng transaction |
| PLANNING/IN_PROGRESS → CANCELLED | PM có lý do; đóng phiên, hủy task chưa DONE, giữ DONE |

PLANNING cho phân rã/phân công nhưng không start task; IN_PROGRESS cho thực thi.
COMPLETED chỉ xem/đóng; CLOSED/CANCELLED chỉ đọc, không mở lại.
Tiến độ = số DONE/CLOSED chia số task không CANCELLED; mẫu số 0 hiển thị chưa có task hiệu lực, không tự hoàn thành project.

| Task transition | Actor và tác động |
|---|---|
| Tạo → PENDING_ASSIGNMENT | PM, chưa tính giờ |
| PENDING_ASSIGNMENT → ASSIGNED | PM confirm |
| ASSIGNED/PAUSED → IN_PROGRESS | Người đang được giao; project IN_PROGRESS; mở phiên |
| IN_PROGRESS → PAUSED | Người được giao, lý do; đóng phiên |
| IN_PROGRESS → PENDING_REVIEW | Người được giao nộp; đóng phiên |
| PENDING_REVIEW → ASSIGNED | PM trả có lý do; không mở phiên |
| PENDING_REVIEW → DONE | PM nghiệm thu, lưu accepted_at |
| DONE → CLOSED | Cùng lúc đóng project |
| PENDING_ASSIGNMENT/ASSIGNED/IN_PROGRESS/PAUSED/PENDING_REVIEW → CANCELLED | PM có lý do; đóng phiên nếu có |

Thu hồi: ASSIGNED/IN_PROGRESS/PAUSED → PENDING_ASSIGNMENT, xóa current assignee.
Thay người: các trạng thái trên → ASSIGNED với assignment mới; đóng phiên người cũ.
PENDING_REVIEW phải trả lại trước khi thay người/thu hồi. Không reopen DONE.
Estimate/requirements đóng băng sau lần giao đầu, kể cả đã thu hồi; không sửa mô tả để lách thay đổi scope.
Scope mới tạo task mới và hủy task cũ khi phù hợp. Ghi chú làm rõ không đổi scope được phép, có audit.

## 4. Phân công, quyền và lịch sử

Batch cùng project, không trùng task. Kiểm tra PM, task, candidate, spec/skill version và run; khóa/version phối hợp mọi đường cập nhật liên quan.
Ghi assignments + current assignee/status + membership + history/events + notifications trong một transaction.
Lỗi một dòng rollback toàn bộ, trả danh sách lỗi; retry với idempotency key không tạo lại.
MANUAL cho override thiếu/dưới mức mandatory với cảnh báo và lý do; không override role/status.
AI chọn hạng khác Top-1 bắt buộc lý do; chuyển manual sau run lưu run_id + lý do.

Assignments append-only. Assignment events ASSIGN/REASSIGN/UNASSIGN; history lưu cả chuyển người không đổi trạng thái (qua event).
`project_members` là “đã tham gia”, không cấp quyền đọc toàn dự án, không xóa khi thay người.
Nhân viên cũ chỉ đọc lịch sử đóng góp của mình qua DTO riêng, không đọc nội dung mới của task.
Task terminal có thể giữ current_assignee_id là người phụ trách cuối; không tính vào tải và không cho thực thi.

## 5. Thời gian và thông báo

Work log là phiên start/end gắn employee/task/assignment và snapshot lịch.
Asia/Ho_Chi_Minh, thứ Hai–Sáu, 08–12 và 13–17, trừ ngày nghỉ; thiếu lịch thì chặn start.
Tính giây giao lịch; tổng actual_hours từ tổng giây phiên đóng, làm tròn khi hiển thị.
Không tính pause/review; chuyển người giữ giờ của người cũ. Không nhập tay giờ.
Đây là thời gian theo trạng thái, không phải giờ lao động được xác minh.
Ví dụ 20/05/2026 15:00 → 21/05/2026 10:30 = 4,5h nếu không có ngày nghỉ.

Thông báo nội bộ cho giao/thay/thu hồi, nộp/trả review, kết quả duyệt yêu cầu; cùng transaction nghiệp vụ.
Không bình luận mới khi project COMPLETED/CLOSED/CANCELLED hoặc task terminal.
Đánh dấu notification đã đọc vẫn được; notification không vượt quyền tài nguyên.
