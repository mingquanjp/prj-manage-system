# Thiết kế cơ sở dữ liệu

Nguồn chuẩn và trạng thái hiện tại của schema.

## 1. Nguồn nào đáng tin

| Nguồn | Mức tin cậy |
|---|---|
| `migrations/000001_init.up.sql` | **Sự thật của sản phẩm.** Đã chạy và kiểm chứng trên Neon |
| `Hình 4.5` (ERD) và `Hình 4.6` (mô hình quan hệ) trong `DoAn.pdf` | **Nguồn chuẩn của tài liệu.** Là ảnh nhúng, chi tiết hơn hẳn file `.md` |
| `docs/thesis-context/tex/4_Ket_qua_thuc_nghiem.tex` mục 4.2.3 | Bảng mô tả từng cột, khớp với hai hình trên |
| `docs/thesis-context/Database Schema Specification ... .md` | ⚠️ **Bản tóm tắt thiếu cột.** Đừng dùng để kết luận "tài liệu chưa có cột X" |

Hai hình trong PDF **mâu thuẫn nhau ở hai chỗ**: `Hình 4.5` có `TaskGroups` và
`ProjectMembers.project_role`, `Hình 4.6` không có. Cách xử lý đã chốt ở mục 3 dưới đây.

Máy này không có `pdftotext`/`poppler`. Để đọc lại ảnh trong PDF, dùng ghostscript (`gs`) —
xem memory `thesis-schema-source-of-truth`.

## 2. Hiện trạng: 22 bảng, 15 enum

**ENUM** — `user_system_role`, `user_status`, `employment_status`, `priority_level`,
`project_status`, `task_status`, `skill_status`, `proficiency_level`, `join_source`,
`assignment_type`, `cv_status`, `extraction_status`, `decision_status`, `change_source`,
`skill_source`.

**Bảng theo nhóm:**

| Nhóm | Bảng |
|---|---|
| Tài khoản và nhân sự | `users`, `departments`, `positions`, `employees` |
| Hồ sơ năng lực | `skill_categories`, `skills`, `employee_skills` |
| Dự án và công việc | `projects`, `project_members`, `task_groups`, `tasks`, `task_requirements` |
| Đề xuất nhân sự | `recommendation_runs`, `recommendation_items` |
| Phân công và theo dõi | `assignments`, `work_logs`, `task_comments`, `task_status_history` |
| Thông báo | `notifications` |
| CV | `employee_cvs`, `cv_extractions`, `cv_extracted_skills` |

Cộng bảng `schema_migrations` do golang-migrate tự tạo là 23 bảng trong CSDL.

## 3. Chênh lệch có chủ đích so với tài liệu

Mọi mục dưới đây đều là **quyết định đã duyệt**, kèm việc cần sửa lại trong báo cáo để tài liệu
khớp sản phẩm.

| Thay đổi | Lý do | Việc cần sửa trong báo cáo |
|---|---|---|
| Thêm bảng `task_groups` | `Hình 4.5` đã có, `Hình 4.6` thiếu. Phương án thay thế là gộp nhóm vào chính bảng `tasks` và phân biệt bằng một cột loại, nhưng như vậy `estimate_hours NOT NULL`, `current_assignee_id`, `task_requirements`, `work_logs` đều vô nghĩa với dòng nhóm và phải thêm ~6 luật chặn | Bổ sung vào `Hình 4.6` và bảng liệt kê mục 4.2.3.3 |
| Thêm `PENDING_REVIEW` vào `task_status` | UC004 bước 5 yêu cầu trạng thái "Chờ review" | Bổ sung vào enum ở mục 4.2.3 |
| Thêm enum `skill_source` và các cột `source`, `approved_by_employee_id`, `approved_at` vào `employee_skills` | Chứng minh luận điểm "chỉ dùng hồ sơ đã duyệt" — trọng tâm đề tài (mục 2.4.2) | Bổ sung enum và ba cột |
| Thêm `recommendation_runs`, `recommendation_items` | Lưu tham số và kết quả để tái lập, tránh tính lại (mục 2.4.3); cũng là dữ liệu đánh giá engine ở Chương 5 | Bổ sung vào mục 4.2.3.3 |
| Thêm `assignments.recommendation_item_id` | Truy vết phân công về đúng đề xuất sinh ra nó | Bổ sung cột |
| Thêm bảng `notifications` | Hậu điều kiện tường minh của UC003 bước 15 | Bổ sung vào mục 4.2.3.3 |
| Thêm `work_logs.source` (`change_source`) | UC004 quy định hệ thống tự tính, không cho nhân viên nhập tuỳ ý | Bổ sung cột |
| `user_system_role` đổi `HR` → `PM` | `HR` không xuất hiện ở bất kỳ use case nào; ba tác nhân Chương 2 là Quản trị viên / Quản lý dự án / Nhân viên | Sửa enum ở mục 4.2.3; sửa mục 2.15 "HR duyệt" → "Quản trị viên duyệt" |
| `tasks.task_code` UNIQUE `(project_id, task_code)` thay vì UNIQUE toàn cục | UC002 chỉ nói "không trùng theo quy tắc của hệ thống" | Ghi rõ trong mô tả bảng `Tasks` |
| Bỏ `ProjectMembers.project_role` | Có ở `Hình 4.5`, không có ở `Hình 4.6`, không use case nào dùng | Xoá khỏi `Hình 4.5` |
| **Bỏ `Tasks.parent_task_id` và `Tasks.task_type`** | Tài liệu mục 4.2.3.4 có cả hai để biểu diễn cây công việc nhiều tầng. Nhưng UC002 chỉ mô tả nhóm công việc chứa công việc, không use case nào nói tới công việc con. Giữ chúng cùng với `task_groups` là có hai cơ chế nhóm chồng nhau, làm quy tắc R2 mơ hồ: task có con thì có phân công được không? | Xoá hai cột khỏi `Hình 4.5`, `Hình 4.6` và bảng mô tả `Tasks`; bỏ câu "Mỗi công việc có thể có công việc cha thông qua khóa ngoại tự tham chiếu" ở mục 4.2.3.2 và 4.2.3.4 |
| Không tạo bảng `Certificates` | Ngoài phạm vi UC001–UC004 | Bỏ "chứng chỉ" khỏi biểu đồ use case 2.2.3 / 2.2.4 / 2.2.6 |
| Thêm bảng `refresh_tokens` (migration `000002`, M1) | Xác thực dùng JWT access + refresh xoay vòng; cần lưu bản băm để thu hồi được. Xem `roles_permissions.md` mục 4 | Bổ sung vào mục 4.2.3.3 |
| Lịch làm việc để ở file config, không vào DB | UC004 alt 9a chỉ đòi "có cấu hình" | Nêu rõ trong Chương 4 |

## 4. Quy ước đặt tên và kiểu dữ liệu

- Tài liệu dùng PascalCase (`EmployeeSkills`), CSDL dùng snake_case (`employee_skills`). Ánh xạ 1-1,
  không đổi ngữ nghĩa.
- Khoá chính: `int GENERATED ALWAYS AS IDENTITY PRIMARY KEY`. Tên cột theo tài liệu
  (`employee_id`, `task_id`, …), không dùng `id` trống.
- `datetime` trong tài liệu → `timestamptz`. Bắt buộc, vì UC004 tính giờ theo lịch làm việc thật.
- Enum dùng `CREATE TYPE ... AS ENUM` native của PostgreSQL, không dùng bảng tra cứu.
- Tiền tệ/giờ: `decimal`, không dùng `float`. `estimate_hours`/`actual_hours` là `decimal(10,2)`;
  các loại điểm là `decimal(5,2)` — nên `matching_score` lưu **theo thang 0–100**, không phải 0–1.
- `updated_at` do trigger `set_updated_at()` tự cập nhật, áp cho 13 bảng có cột này. **Không** set
  `updated_at` thủ công trong câu `UPDATE`.

## 5. Ràng buộc quan trọng

| Ràng buộc | Bảng | Ý nghĩa |
|---|---|---|
| `ck_projects_dates` | `projects` | `planned_end_date >= start_date` |
| `uq_tasks_code` | `tasks` | UNIQUE `(project_id, task_code)` |
| `ck_tasks_estimate` | `tasks` | `estimate_hours > 0` |
| `ck_tasks_dates` | `tasks` | ngày kết thúc không nhỏ hơn ngày bắt đầu nếu nhập cả hai |
| `uq_task_groups_name` | `task_groups` | UNIQUE `(project_id, group_name)` |
| `uq_employee_skills` | `employee_skills` | UNIQUE `(employee_id, skill_id)` |
| `uq_task_requirements` | `task_requirements` | UNIQUE `(task_id, skill_id)` |
| `ck_task_req_weight` | `task_requirements` | `weight > 0` |
| `uq_project_members` | `project_members` | UNIQUE `(project_id, employee_id)` |
| `uq_recommendation_items` | `recommendation_items` | UNIQUE `(run_id, employee_id)` |

`ON DELETE CASCADE` chỉ dùng cho quan hệ sở hữu thật (task → requirement, project → task,
extraction → extracted skill). Các FK trỏ tới `employees` cố ý **không** cascade để không mất vết
lịch sử khi xoá nhân sự.

## 6. Migration và seed

- Migration chạy bằng golang-migrate **dạng thư viện** (`cmd/migrate`), file nhúng qua `embed.FS`.
  Không cần cài CLI. Luôn đi endpoint trực tiếp, xem `architecture.md` mục 4.
- Thêm thay đổi schema: tạo cặp file mới `000002_<tên>.up.sql` / `.down.sql`. **Không sửa
  `000001`** khi đã áp dụng lên DB thật.
- Seed (`make seed`) idempotent, chạy trong một transaction, nạp: 5 phòng ban, 10 chức danh,
  8 nhóm kỹ năng, 51 kỹ năng, 20 tài khoản + hồ sơ nhân sự (1 ADMIN, 2 PM, 17 EMPLOYEE),
  96 bản ghi năng lực đã đánh dấu duyệt. Mật khẩu chung `Password@123` — chỉ dùng cho môi trường
  phát triển.
