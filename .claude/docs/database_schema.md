# Thiết kế dữ liệu — hiện trạng và đích V1

## 1. Phân biệt nguồn

- Hiện trạng mã migration: migrations/000001_init.up.sql, 22 bảng nghiệp vụ/15 enum.
- Lượt kiểm tra DB trước ghi nhận version 1, dirty=false; không chứng minh mọi cột/constraint đúng chỉ từ version.
- Thiết kế đích đã đồng ý: [baseline V1](../../docs/reviews/system-design-review-v1.md) và các tài liệu miền hiện hành.
- DoAn.pdf/TeX trong docs/thesis-context là tài liệu gốc chỉ đọc, chưa đồng bộ V1.

Không sửa migration 000001 đã áp dụng. Các thay đổi dưới đây CHƯA được tạo migration/chạy DB.
Không ghi số thứ tự migration cố định trước khi chia dependency và kiểm tra dữ liệu thật.

## 2. Hiện trạng 000001

Users/employees/departments/positions; skills/categories/employee_skills;
projects/project_members/task_groups/tasks/task_requirements;
recommendation_runs/items; assignments/work_logs/comments/task_status_history;
notifications; employee_cvs/cv_extractions/cv_extracted_skills.

Hiện vẫn có parent_task_id/task_type, task group nullable, is_selected và work_date.
Chưa có PAUSED/CANCELLED cho task; chưa có bảng yêu cầu riêng, session/refresh hoặc snapshot đầy đủ.
Một số FK cascade có thể xóa chuỗi lịch sử. Không coi schema hiện tại đã bảo vệ đủ nghiệp vụ V1.

## 3. Bảng/cột thiết kế đích

| Miền | Thay đổi bắt buộc |
|---|---|
| User/employee | Tạo atomically; version cho dữ liệu cần chống ghi đè; cơ chế buộc đổi mật khẩu |
| Official skills | Giữ unique employee/skill; approval actor/time bắt buộc; hiệu lực, version; không xóa khi REMOVE |
| employee_skill_change_requests | request_id, employee_id, skill_id, ADD/UPDATE/REMOVE, MANUAL/CV, proposed/before values, base_version, evidence, extracted_skill_id, submitter/time, status, reviewer/time/reason |
| employee_profile_change_requests | Trường đề nghị được allowlist, before/proposed, base_version, actor/time, quyết định/reason; không nhận role/status |
| Projects | version, project_status_history; audit chuyển PM ghi owner cũ/mới/lý do |
| Tasks | Group NOT NULL cùng project; bỏ parent_task_id/task_type; PAUSED/CANCELLED; version, spec_version, first_assigned_at, accepted_at |
| Assignments | Append-only; run_id, item_id, adjust_reason, score server; truy vết task/employee/người xác nhận |
| assignment_events | ASSIGN/REASSIGN/UNASSIGN, task, assignment cũ/mới, actor/time/reason |
| Recommendation runs | engine_version, params, snapshot_at, spec_version, input_snapshot/output_snapshot JSONB có schema version; không chứa CV thô |
| Recommendation items | Top-N/rank, skill/workload/perf/tổng, matched/missing, explanation; bỏ is_selected, lựa chọn suy từ assignment |
| work_logs | assignment_id, task/employee, started_at, ended_at, valid_work_seconds, calendar_version, close_reason, history references; bỏ work_date |
| Work calendar | Cấu hình nguồn có thể là file; snapshot bất biến lưu bền vững và truy xuất bằng version trong DB |
| CV | cv_status thêm PENDING/REJECTED; tách processing_status và review_status; job lease/attempt/error code; suggested_level nullable |
| auth_sessions | session_id, user_id, created_at, expires_at, revoked_at; hạn tuyệt đối 7 ngày |
| refresh_tokens | hash unique, session_id, expiry/revoked/replacement; giữ token đã dùng để phát hiện reuse |
| idempotency records | actor + operation + key unique, request hash, response/result; cùng transaction nghiệp vụ |
| Audit | before/after có giới hạn dữ liệu, actor/time/reason cho spec, quyền, ownership; append-only |

Tên vật lý/DDL cuối cùng cần được review khi viết migration; bảng trên là hợp đồng dữ liệu tối thiểu, không DDL đã chạy.
Request skill terminal: APPROVED/REJECTED/WITHDRAWN; official update và quyết định cùng transaction.
CV publication PENDING/ACTIVE/REPLACED/REJECTED/ARCHIVED;
processing QUEUED/PROCESSING/SUCCEEDED/FAILED;
review PENDING_REVIEW/APPROVED/PARTIALLY_APPROVED/REJECTED.

## 4. Ràng buộc và nguồn sự thật

- Unique project_code; unique task_code trong project; unique group_name trong project.
- Composite FK tasks(project_id, task_group_id) → task_groups(project_id, task_group_id) với unique tương ứng.
- Estimate >0; weight nguyên 1–5; ngày task cùng NULL hoặc đủ cặp, đúng thứ tự.
- Task có ≥1 requirement và ngày trong project: service transaction, không CHECK truy vấn bảng khác.
- Unique employee/skill chính thức; partial unique pending request/người/skill; extracted_skill_id liên kết request unique khi khác NULL.
- Một CV ACTIVE/người bằng partial unique; xuất bản khóa hồ sơ người để tránh hai lần duyệt đồng thời.
- Unique run/employee, run/rank; rank >0; mọi điểm persist 0–100.
- Một assignment có run_id khác NULL cho mỗi run (partial unique); item phải khớp run/task/employee.
- AI cần run/item/score; MANUAL item/score NULL, có thể có run và reason. Kiểm tra quan hệ và lý do trong service.
- Một phiên mở/employee và /task bằng partial unique; end ≥ start; phiên đóng có seconds ≥0, phiên mở chưa có tổng.
- current_assignee là nguồn người hiện tại; assignment/events là lịch sử. PENDING_ASSIGNMENT không người; trạng thái thực thi phải có người.
- actual_hours là cache từ tổng seconds phiên đóng, không cộng giờ làm tròn từng đoạn; đối soát được.
- Snapshot đầu vào scorer nhất quán trong một giao dịch đọc; confirm khóa/version kiểm tra điều kiện hiện hành.
- Không API xóa vật lý entity có lịch sử; thay cascade nguy hiểm bằng RESTRICT/NO ACTION phù hợp, kiểm tra cả đường gián tiếp.

## 5. Migration và backfill

1. Kiểm tra COUNT/null/duplicate/orphan/enum thật trên DB đích; không dùng ước lượng thống kê thay COUNT.
2. Tách migrations theo auth → request/version → task/assignment/time → recommendation → CV, điều chỉnh theo FK.
3. Add nullable/backfill/validate rồi mới NOT NULL/drop. Task group thiếu cần gán nhóm hợp lệ có kế hoạch.
4. Official skill cũ thiếu approval cần phân loại, không tự gán ADMIN giả để vượt constraint.
5. Work log cũ chỉ có ngày/giờ không thể tự suy start/end thật: giữ legacy/archive, không tạo lịch sử giả.
6. Test up/down trên DB thử; rollback mất dữ liệu phải có backup/khôi phục, không quảng cáo down luôn an toàn.
7. Sửa seed/test/OpenAPI/ERD cùng phiên bản sau khi DDL được duyệt.

Identity PK, snake_case, timestamptz; giờ/điểm decimal, thời lượng chuẩn lưu giây.
Giữ trigger updated_at hiện có; version cập nhật atomically khác với timestamp.
Seed Password@123 chỉ dev; không chạy seed/reset DB thật trong lượt cập nhật tài liệu này.
