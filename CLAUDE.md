# CLAUDE.md

File này là **index cấp cao**. Chi tiết nằm trong `.claude/docs/` — đọc file tương ứng
trước khi động vào miền đó.

## 1. Project Overview

Hệ thống quản lý dự án, quản lý nhân sự, phân rã công việc và phân công công việc, trong đó
AI đề xuất danh sách nhân sự phù hợp cho từng công việc dựa trên hồ sơ năng lực đã được duyệt.
AI chỉ **đề xuất và giải thích**; Quản lý dự án là người xác nhận phân công cuối cùng.

Đây là sản phẩm của một đồ án tốt nghiệp: mã nguồn phải bám sát use case Chương 2 và thiết kế
CSDL mục 4.2.3 trong `docs/thesis-context/DoAn.pdf`, vì báo cáo và sản phẩm phải khớp nhau.

## 2. Tech Stack

Đã chốt và **đã có trong repo**:

| Lớp | Công nghệ | Ghi chú |
|---|---|---|
| Ngôn ngữ | Go 1.26.5 | `go.mod` module `ai-staffing` |
| HTTP router | `github.com/go-chi/chi/v5` | `internal/transport/http/router.go` |
| Driver DB | `github.com/jackc/pgx/v5` + `pgxpool` | `internal/repository/postgres/db.go` |
| Migration | `golang-migrate/v4` dùng **dạng thư viện** + `embed.FS` | `cmd/migrate`, `migrations/` |
| CSDL | PostgreSQL trên **Neon** (serverless) | cấu hình duy nhất qua `DATABASE_URL` |
| Hash mật khẩu | `golang.org/x/crypto/bcrypt` | dùng ở `cmd/seed` |
| Log | `log/slog` (thư viện chuẩn) | |
| Xác thực | **JWT access 15 phút + refresh token xoay vòng**, cả hai trong cookie HttpOnly | đã chốt, chưa implement — xem `roles_permissions.md` mục 4 |

Đã chốt nhưng **chưa khởi tạo** (`web/` đang rỗng): Vite + React + TypeScript + Tailwind +
shadcn/ui + TanStack Query + React Router; build static rồi nhúng vào binary Go bằng `embed.FS`
→ một process, một port, không CORS.

**TBD — cần chốt trước khi code phần liên quan:**

1. **Sinh mã truy vấn**: `sqlc` + `squirrel` mới nằm trong kế hoạch, **chưa có trong `go.mod`**.
   Hiện tại viết SQL trực tiếp bằng pgx. Đừng thêm dependency khi chưa hỏi.
2. **Nơi lưu file CV** (chặn M7): local disk qua interface `storage.Provider` hay S3/MinIO.
3. **Đa ngôn ngữ**: đang giả định chỉ tiếng Việt.

## 3. Dev Commands

```bash
cp .env.example .env      # rồi điền DATABASE_URL của Neon; .env đã nằm trong .gitignore
make migrate-up           # áp dụng migration (tự dùng endpoint trực tiếp, không qua -pooler)
make seed                 # nạp dữ liệu mẫu, idempotent, chạy lại nhiều lần vẫn an toàn
make dev                  # chạy server; kiểm tra GET /api/healthz
make check                # fmt → tidy → vet → build → test, chạy trước khi commit
make migrate-reset        # down-all + up. Lệnh phá dữ liệu chỉ chạy khi APP_ENV=dev
```

Lint/typecheck riêng: **chưa có** (`golangci-lint` chưa cấu hình). `make check` là cổng kiểm tra
duy nhất hiện tại. Frontend: chưa có lệnh nào vì `web/` còn rỗng.

## 4. Core Domain Summary

- **Users** — tài khoản đăng nhập, `system_role ∈ {ADMIN, PM, EMPLOYEE}`, khoá/mở bằng `status`.
- **Employees** — hồ sơ nhân sự, 1-1 với `users`, gắn `departments` và `positions`.
- **Skills** — danh mục `skill_categories` → `skills`; `employee_skills` là **hồ sơ năng lực
  chính thức** (có `proficiency_level`, `source`, `approved_by_employee_id`, `approved_at`).
- **Projects** — dự án có đúng một `pm_employee_id` phụ trách; `project_members` là nhóm tham gia.
- **Tasks** — `task_groups` (chỉ để gom nhóm) → `tasks` → `task_requirements` (kỹ năng, mức độ,
  trọng số, bắt buộc). Task là đơn vị được phân công duy nhất.
- **Assignments** — kết quả phân công cuối cùng do PM xác nhận, kèm `assignment_type ∈ {AI, MANUAL}`.
- **AI recommendation** — `recommendation_runs` + `recommendation_items` lưu Top-N kèm điểm thành
  phần và câu giải thích, để tái lập được kết quả và đánh giá engine.
- **CV extraction** — `employee_cvs` → `cv_extractions` → `cv_extracted_skills`; là dữ liệu **chờ
  duyệt**, không phải hồ sơ chính thức.
- **Theo dõi thực hiện** — `work_logs` (giờ do hệ thống tự tính), `task_status_history`,
  `task_comments`, `notifications`.

## 5. Key Constraints

Những điều **không được tự ý thay đổi hoặc giả định**:

1. **AI chỉ đề xuất, không tự phân công.** Không có đường code nào ghi `assignments` mà không đi
   qua thao tác xác nhận của PM.
2. **Task là đơn vị phân công duy nhất.** `task_groups` chỉ gom nhóm — không gán người, không có
   giờ ước tính, không có yêu cầu kỹ năng.
3. **Kỹ năng trích xuất từ CV phải được ADMIN duyệt** mới được ghi vào `employee_skills`. Engine
   đề xuất chỉ đọc bản ghi đã duyệt.
4. **Phân quyền hai tầng**: RBAC theo `system_role`, cộng ABAC — mọi thao tác trên một dự án cụ
   thể phải thoả `project.pm_employee_id == current_employee_id`.
5. **Không đổi schema** khi chưa đối chiếu `Hình 4.5`/`Hình 4.6` trong `DoAn.pdf`. File
   `docs/thesis-context/Database Schema Specification ... .md` là **bản tóm tắt thiếu cột**, không
   phải nguồn chuẩn.
6. **Không tự chọn công nghệ** cho các mục TBD ở phần 2.
7. **Không sửa file trong `docs/thesis-context/`** — đó là tài liệu gốc, chỉ đọc.
8. Engine chấm điểm phải nằm trong `internal/matching`, **không biết tới HTTP hay DB** (yêu cầu
   mục 2.4.6 về khả năng bảo trì).
9. Toàn bộ chuỗi ghi khi xác nhận phân công phải nằm **trong một transaction** (UC003 alt 16a).
10. Trao đổi và bình luận code bằng **tiếng Việt**; định danh kỹ thuật (bảng, hàm, biến,
    endpoint) bằng **tiếng Anh**.
11. **Quản lý nhánh** Trước khi thêm feature hay fix bug, luôn làm việc trên 1 branch mới. Không được commit thẳng lên main, bug thì nên được đặt là bug/[des], feature thì nên được đặt là feature/[des]
12. **Bảo mật là ưu tiên xuyên suốt**, không phải việc làm thêm ở cuối. Từ chối mặc định; không
    tin dữ liệu client gửi; không hardcode bí mật; không log token hay chuỗi kết nối. Đọc
    `.claude/docs/security.md` **trước khi** làm auth, upload tệp, hoặc bất kỳ endpoint nào có `{id}`.

## 6. Additional Documentation

| File | Đọc khi nào |
|---|---|
| [.claude/docs/product_scope.md](.claude/docs/product_scope.md) | Cần biết chức năng nào trong phạm vi, chức năng nào cố ý bỏ, và lộ trình M0–M8 |
| [.claude/docs/architecture.md](.claude/docs/architecture.md) | Trước khi tạo package mới hoặc quyết định đặt logic ở tầng nào |
| [.claude/docs/domain_rules.md](.claude/docs/domain_rules.md) | Trước khi viết bất kỳ logic nghiệp vụ nào — chứa đầy đủ ràng buộc và vòng đời trạng thái |
| [.claude/docs/database_schema.md](.claude/docs/database_schema.md) | Khi viết truy vấn, migration, hoặc cần biết chỗ nào code cố ý lệch khỏi tài liệu |
| [.claude/docs/roles_permissions.md](.claude/docs/roles_permissions.md) | Khi viết middleware, kiểm tra quyền, hoặc thêm endpoint mới |
| [.claude/docs/security.md](.claude/docs/security.md) | **Trước khi** làm auth, upload tệp, truy vấn động, hoặc endpoint có `{id}` — checklist chống IDOR, SQL injection, rò rỉ bí mật |
| [.claude/docs/project_task_assignment_workflow.md](.claude/docs/project_task_assignment_workflow.md) | Khi làm UC001 / UC002 / UC003 / UC003.2 / UC004 — có đủ luồng chính và luồng thay thế |
| [.claude/docs/ai_recommendation_workflow.md](.claude/docs/ai_recommendation_workflow.md) | Khi làm UC003.1 và gói `internal/matching` — có công thức chấm điểm đầy đủ |
| [.claude/docs/cv_extraction_approval_workflow.md](.claude/docs/cv_extraction_approval_workflow.md) | Khi làm M7 — luồng CV và những điểm tài liệu chưa đặc tả |
| [.claude/docs/api_design.md](.claude/docs/api_design.md) | Khi thêm endpoint — quy ước đường dẫn, mã lỗi, định dạng response |
| [.claude/docs/testing_verification.md](.claude/docs/testing_verification.md) | Khi viết test hoặc cần nghiệm thu một mốc |
