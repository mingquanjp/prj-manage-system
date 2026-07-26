# Quy tắc nghiệp vụ

Đọc file này **trước khi viết bất kỳ logic nghiệp vụ nào**. Mọi quy tắc ở đây đều có neo vào tài
liệu; phần nào không có neo được ghi rõ là suy ra hoặc TBD.

## 1. Quy tắc bất biến

Bốn quy tắc dưới đây là trọng tâm của đề tài. Không được nới lỏng, không được đi tắt.

**R1 — AI chỉ đề xuất, không tự phân công.**
Kết quả UC003.1 chỉ đi vào *bảng phân công dự kiến*. Bản ghi `assignments` chỉ được tạo ở bước xác
nhận của PM trong UC003. Neo: mục 2.4.5, UC003.1 bước 12.

**R2 — Task là đơn vị phân công duy nhất.**
`task_groups` chỉ để gom nhóm: không có người phụ trách, không có giờ ước tính, không có yêu cầu kỹ
năng, không xuất hiện trong bất kỳ truy vấn phân công nào. Neo: UC002 bước 7–8.

**R3 — Kỹ năng phải được duyệt mới thành hồ sơ chính thức.**
Kỹ năng nhân viên tự khai hoặc trích xuất từ CV chỉ là dữ liệu chờ duyệt. Engine đề xuất và chức
năng tra cứu **chỉ đọc bản ghi đã duyệt**. Neo: mục 2.4.2, UC003.1 dữ liệu đầu vào #3.

**R4 — Quyết định cuối cùng thuộc về Quản lý dự án.**
Kể cả khi Top-1 rõ ràng vượt trội, hệ thống vẫn phải để PM chọn và xác nhận. Neo: mục 2.1.5, 2.4.5.

## 2. Tài khoản và nhân sự

- Mỗi `employees` gắn với đúng một `users` (`user_id UNIQUE`). Không có nhân viên không tài khoản
  và ngược lại.
- `users.status`: `ACTIVE` | `LOCKED` | `DISABLED`. Chỉ `ACTIVE` mới đăng nhập được.
- `employees.employment_status`: `ACTIVE` | `INACTIVE` | `RESIGNED`.
- Nhân sự chỉ được **chọn để phân công** khi `employment_status = ACTIVE` **và** `users.status = ACTIVE`.
  Neo: UC003.2 alt 10a, UC003 bước 12.
- Nhân viên không được tự sửa `employee_skills` đã duyệt. Neo: mục 2.4.1.

## 3. Hồ sơ năng lực

- `employee_skills` UNIQUE `(employee_id, skill_id)` — mỗi người mỗi kỹ năng một bản ghi.
- `source ∈ {MANUAL, CV}` cho biết dữ liệu đến từ đâu.
- **Đã duyệt** ⇔ `approved_by_employee_id IS NOT NULL AND approved_at IS NOT NULL`. Đây là điều kiện
  lọc bắt buộc của engine đề xuất và của mọi màn hình tra cứu năng lực.
- Người duyệt phải có `system_role = ADMIN`. Neo: mục 2.1.2, 2.2.3.
- `skills.status = INACTIVE` thì không được chọn làm yêu cầu kỹ năng mới. Neo: UC002 alt 13d.

## 4. Dự án

- `projects.status`: `PLANNING` | `IN_PROGRESS` | `COMPLETED` | `CLOSED` | `CANCELLED`.
  Mặc định khi tạo: `PLANNING`. Neo: UC001 bước 6.
- `project_code` UNIQUE toàn hệ thống. Neo: UC001 alt 9b.
- `planned_end_date >= start_date`. Neo: UC001 dữ liệu đầu vào #5, ràng buộc `ck_projects_dates`.
- Người tạo dự án tự động thành `pm_employee_id`. Neo: UC001 bước 6.
- Dự án ở `COMPLETED` hoặc `CLOSED` **không cho phân rã công việc mới và không cho phân công mới**.
  Neo: UC002 alt 5a, UC003 alt 5a.
- `project_members` UNIQUE `(project_id, employee_id)`, `join_source ∈ {MANUAL, FROM_ASSIGNMENT}`.

## 5. Công việc

- `task_code` UNIQUE theo `(project_id, task_code)` — trùng mã giữa hai dự án khác nhau là hợp lệ.
- `estimate_hours > 0`, bắt buộc. Neo: UC002 alt 13e, ràng buộc `ck_tasks_estimate`.
- Nếu nhập cả hai ngày thì `planned_end_date >= planned_start_date`. Neo: UC002 alt 13c.
- **Mỗi task phải có ít nhất một `task_requirements`** mới đủ điều kiện phân công. Neo: UC002 alt 13b,
  UC003 tiền điều kiện, UC003.1 alt 3a.
- `task_requirements`: UNIQUE `(task_id, skill_id)`, `weight > 0`, `is_mandatory` mặc định `false`.

### Vòng đời trạng thái công việc

```
PENDING_ASSIGNMENT ──phân công──▶ ASSIGNED ──nhân viên bắt đầu──▶ IN_PROGRESS
                                                                      │
                                                       nhân viên nộp  ▼
                                                                PENDING_REVIEW
                                                                 │         │
                                                        PM duyệt ▼         ▼ PM trả lại
                                                               DONE    IN_PROGRESS
                                                                 │
                                                        PM đóng  ▼
                                                              CLOSED
```

- Mọi chuyển trạng thái phải ghi một dòng `task_status_history`. Bản ghi đầu tiên có `old_status = NULL`.
- Chuyển trạng thái không đúng sơ đồ → báo lỗi, **không** cập nhật. Neo: UC004 alt 6a.
- Chỉ người đang phụ trách (`current_assignee_id`) mới được chuyển `IN_PROGRESS → PENDING_REVIEW`.
  Neo: UC004 dữ liệu đầu vào #2.
- `PENDING_REVIEW` là giá trị **bổ sung** so với tài liệu 4.2.3 — tài liệu cần cập nhật.

## 6. Phân công

Khi PM xác nhận, các thao tác sau phải nằm **trong một transaction duy nhất** (UC003 bước 13–15,
alt 16a — lỗi một dòng thì rollback trọn vẹn, giữ nguyên dữ liệu trước đó):

1. Kiểm tra lại từng dòng: trạng thái task, trạng thái nhân sự, quyền của PM trên dự án.
2. Tạo `assignments` (kèm `assignment_type`, `note`, `matching_score`, `recommendation_item_id` nếu có).
3. Cập nhật `tasks.status = ASSIGNED` và `tasks.current_assignee_id`.
4. Thêm `project_members` với `join_source = FROM_ASSIGNMENT` nếu người đó chưa thuộc dự án.
5. Ghi `task_status_history`.
6. Tạo `notifications` cho nhân sự được phân công.

Luồng thay thế:

- **15a** — một vài dòng lỗi: chỉ rõ dòng nào, lý do gì; PM sửa lại bảng dự kiến.
- **15b** — toàn bộ dòng lỗi: không lưu gì, yêu cầu phân công lại từ đầu.

`assignments` là bảng **append-only** về mặt nghiệp vụ: phân công lại thì thêm dòng mới, không sửa
dòng cũ — cần giữ vết để đánh giá engine.

## 7. Ghi nhận thời gian (UC004)

- Nhân viên **không nhập tay số giờ**. Hệ thống tự tính từ lịch sử trạng thái. Neo: mục 2.3.6 mở đầu.
- Khoảng tính: từ thời điểm task vào `IN_PROGRESS` đến thời điểm chuyển sang `PENDING_REVIEW`
  (hoặc trạng thái hoàn tất tương đương).
- Lịch làm việc chuẩn mặc định: 08:00–12:00 và 13:00–17:00, tức 8 giờ một ngày. Loại trừ nghỉ trưa,
  ngoài giờ và các khoảng cấu hình không tính. Neo: UC004 bước 8–9.
- Chưa có cấu hình lịch làm việc → **không tính, không lưu**, báo lỗi. Neo: UC004 alt 9a.
- Bản ghi ghi vào `work_logs` với `source = SYSTEM`, rồi cộng dồn vào `tasks.actual_hours`.
- Sau đó đối chiếu `actual_hours` với `estimate_hours` — đây là nguồn dữ liệu cho `PerfScore` của engine.

## 8. Thông báo

`notifications` được tạo khi phân công công việc (UC003 bước 15). Các sự kiện khác (đổi trạng thái,
bình luận mới) chưa được tài liệu yêu cầu — **TBD**, đừng tự thêm khi chưa hỏi.
