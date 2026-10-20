# Bản chỉ dẫn trước khi hợp nhất harness — 20/10/2026

> Lưu nguyên nội dung làm việc trước đợt dọn harness để không mất các ghi chú chưa commit. Không còn là chỉ dẫn triển khai. Điểm vào hiện hành: [AGENTS.md](../../AGENTS.md), [mục lục nguồn](../README.md). Các đường dẫn trong nội dung dưới đây được viết theo vị trí gốc CLAUDE.md.

# CLAUDE.md

**Ưu tiên quyết định 15/10/2026:** [sổ quyết định hiện hành/Bàn giao](docs/design/database-review-decisions-2026-10-14.md), [thiết kế DB 32 bảng](docs/design/database-proposal-2026-10-14.md). PTCM duyệt hiệu chỉnh, không PM ghi/gửi thay; HR quản trị departments/positions, Giám đốc quản trị skill_categories/skills. Session phía server thay JWT/refresh. Tài liệu đã đồng bộ, code/migration vẫn là hiện trạng cũ; không chạy DB thật từ việc cập nhật thiết kế.

File này là **index cấp cao**. Chi tiết nằm trong `.claude/docs/` — đọc file tương ứng
trước khi động vào miền đó.

**Bổ sung 20/10/2026:** [quyết định tạo ban đầu](docs/business-baseline-2026-10-20.md): Admin tạo tài khoản kèm hồ sơ nhân viên ban đầu trong một lần; HR quản lý/cập nhật hồ sơ về sau. Không mở quyền Admin sửa hành chính thường xuyên. [Bộ đầu vào Figma Make](docs/design/figma-make/README.md) đã thay brief/prompt lịch sử bằng ma trận màn hình/quyền và checklist hiện hành; không suy prototype đã cập nhật chỉ từ tài liệu.

**Chốt tiếp nối địa chỉ 20/10/2026:** địa chỉ phải bổ sung khi hoàn thiện hồ sơ, không bắt Admin nhập lúc tạo. Thiết kế `employees.address TEXT NULL`; chủ hồ sơ đề nghị fullname/phone/address để HR duyệt, HR có thể sửa trực tiếp. Không xóa trắng địa chỉ đã có; thiếu địa chỉ chỉ nhắc bổ sung, không chặn đăng nhập/công việc. Không DTO tra cứu/ứng viên/gợi ý, không thêm bảng hoặc trạng thái; chưa sửa/chạy migration. Quyết định này thay allowlist hai trường ở các mốc lịch sử.

Các workflow/quy tắc dùng bản chuẩn trong `.claude/docs/`; không duy trì bản sao ở thư mục gốc.
Các ghi chú về “bản đọc đồng bộ ở gốc” trong baseline lịch sử phản ánh cấu trúc trước khi dọn tài liệu.

**Nguồn nghiệp vụ ưu tiên từ 2026-10-12:** [cơ sở hiện hành](docs/business-baseline-2026-10-12.md). HR quản lý hồ sơ hành chính/CV và duyệt thông tin cá nhân; Giám đốc xét duyệt kỹ năng, tạo dự án và chỉ định/thay PM. Nhân viên đề nghị sửa fullname/phone/address theo chốt tiếp nối 20/10; email công ty/cơ cấu HR sửa trực tiếp theo quyết định 15/10.
PM và PTCM là trách nhiệm theo dự án có thể kiêm nhiệm. PTCM phân công và nghiệm thu, không nhận task trong dự án phụ trách; không có trách nhiệm nghiệm thu riêng. Kế hoạch chia sẻ ngoài hệ thống (ví dụ Slack), PM bắt đầu triển khai. Tra cứu PM/PTCM chỉ thông tin nghề nghiệp, không CV gốc người khác.
R01–R05 đã chốt giản lược: không duyệt riêng quên Start/Resume, không nộp kết quả thay, không mở lại mục CV, không có chức năng tái kích hoạt kỹ năng riêng, không xác nhận lịch sử gợi ý cũ. Chi tiết và giới hạn theo baseline mới; nhật ký đã khóa vẫn theo hiệu chỉnh chung.
Tài liệu role/DB/API đã đồng bộ quyết định 15/10; review V2 là lịch sử, code/migration chưa chuyển đổi. Các contract còn mở được liệt kê trong bản hợp nhất.
Phê duyệt kỹ năng theo [quy trình đã chốt](.claude/docs/skill_approval_workflow.md): Giám đốc duyệt có căn cứ/lý do, cần ý kiến chuyên môn nếu không đủ khả năng đánh giá và không tự duyệt. Giám đốc được đọc đúng căn cứ task dẫn chiếu, không toàn dự án. Cho bổ sung khi PENDING; terminal bất biến; không hạn ngạch tuần/tháng. Đã đồng bộ thiết kế DB/API, chưa triển khai code.

**Bổ sung hiện hành 2026-10-13:** [nguồn nhân lực chung](docs/business-baseline-2026-10-13.md). PTCM phân công toàn công ty, không lập đội trước; tải xuyên dự án, số dự án đang có việc chỉ tham khảo, không giới hạn hai dự án. Thiết kế đích suy danh sách người từng được giao từ assignments, không lưu `project_members` riêng; migration cũ chưa thay đổi, dữ liệu cũ phải kiểm kê trước chuyển đổi.

## 1. Project Overview

Hệ thống quản lý dự án, quản lý nhân sự, phân rã công việc và phân công công việc, trong đó
Engine dựa trên luật/công thức đề xuất nhân sự từ hồ sơ năng lực đã duyệt.
Engine chỉ **đề xuất và giải thích**; phụ trách chuyên môn của dự án xác nhận cuối cùng. Không bắt buộc mô hình ML.

Đây là sản phẩm của một đồ án tốt nghiệp: mã nguồn phải bám sát use case Chương 2 và thiết kế
CSDL mục 4.2.3 trong `docs/thesis-context/DoAn.pdf`, nhưng quyết định nghiệp vụ mới được ưu tiên khi chưa đồng bộ báo cáo.
**Nghiệp vụ đích**: [cơ sở hiện hành](docs/business-baseline-2026-10-12.md) và `.claude/docs/`.
[Baseline V2](docs/reviews/system-design-baseline-v2.md) và [review R01–R06](docs/reviews/system-design-review-v2-r01-r06.md) là tài liệu lịch sử; không dùng cơ chế đã bỏ trong đó làm yêu cầu triển khai.
**Hiện trạng triển khai**: mã nguồn/migration hiện có; V2 chưa được chạy migration.
Báo cáo gốc chỉ đọc, chưa đồng bộ V2; không dùng nội dung cũ để ghi đè quyết định mới.

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
| Tài liệu API | **OpenAPI 3.1 viết tay** + Swagger UI nhúng tại `/api/docs`, chỉ bật khi `APP_ENV=dev` | đã chốt, chưa implement — xem `api_design.md` mục 3 |
| Xác thực | **Session phía server**, cookie secret HttpOnly, DB lưu token_hash | đã chốt, chưa implement — xem `roles_permissions.md` mục 4 |

Frontend đã chốt: Vite + React + TypeScript + Tailwind + shadcn/ui + TanStack Query + React Router.
**Lát cắt duyệt thiết kế 20/10/2026** đã khởi tạo trong `web/`: shell, danh sách/chi tiết dự án và phân công mô phỏng; Button/Dialog tùy biến theo mô hình shadcn/Radix, chưa phải toàn bộ sản phẩm. Đọc `web/AGENTS.md` và `web/DESIGN-PILOT.md` trước khi mở rộng. TanStack Query/API/auth chưa triển khai; dữ liệu chỉ trong bộ nhớ.
Đích triển khai vẫn là build static rồi nhúng vào binary Go bằng `embed.FS` → một process, một port, không CORS; lát cắt local chưa tích hợp Go.

**TBD — cần chốt trước khi code phần liên quan:**

1. **Sinh mã truy vấn**: `sqlc` + `squirrel` mới nằm trong kế hoạch, **chưa có trong `go.mod`**.
   Hiện tại viết SQL trực tiếp bằng pgx. Đừng thêm dependency khi chưa hỏi.
2. **Thư viện parser CV**: cần review trước khi thêm dependency.
   Hướng lưu cloud một nhà cung cấp cấu hình chung qua `storage.Provider`, API nhận diện tên kỹ năng/đoạn căn cứ; provider/model/cách chạy chưa chọn. PDF text/DOCX tối đa 10 MiB.
3. **Ngôn ngữ**: UI tiếng Việt; CV Việt/Anh trong phạm vi thử nghiệm, không OCR.

## 3. Dev Commands

```bash
cp .env.example .env      # rồi điền DATABASE_URL của Neon; .env đã nằm trong .gitignore
make migrate-up           # áp dụng migration (tự dùng endpoint trực tiếp, không qua -pooler)
make seed                 # nạp dữ liệu mẫu, idempotent, chạy lại nhiều lần vẫn an toàn
make dev                  # chạy server; kiểm tra GET /api/healthz
make check                # fmt → tidy → vet → build → test, chạy trước khi commit
make migrate-reset        # drop + up, CHỈ dùng khi phát triển
```

Lint/typecheck riêng: **chưa có** (`golangci-lint` chưa cấu hình). `make check` là cổng kiểm tra
backend hiện tại. Frontend tại `web/`: `pnpm dev` (5173), `pnpm typecheck`, `pnpm test`, `pnpm build`; `pnpm check` chạy ba kiểm tra frontend, không truy cập DB. Chưa có lint riêng.

## 4. Core Domain Summary

- **Users** — enum `system_role ∈ {ADMIN, PM, EMPLOYEE}` là thiết kế kỹ thuật cũ, chưa phản ánh HR, Giám đốc và trách nhiệm theo dự án. Không dùng enum này để phủ định baseline 2026-10-12.
- **Employees** — hồ sơ nhân sự, 1-1 với `users`, gắn `departments` và `positions`.
- **Skills** — danh mục `skill_categories` → `skills`; `employee_skills` là **hồ sơ năng lực
  chính thức** (có `proficiency_level`, `approved_by` trỏ users, `approved_at`; nguồn ở request, không source/is_effective/version trên official).
- **Projects** — dự án có đúng một `pm_employee_id` phụ trách; danh sách người từng được giao suy từ lịch sử assignments. `project_members` thuộc schema cũ, không phải roster bắt buộc của thiết kế đích.
- **Tasks** — `task_groups` (chỉ để gom nhóm) → `tasks` → `task_requirements` (kỹ năng, mức độ,
  trọng số, bắt buộc). Task là đơn vị được phân công duy nhất.
- **Assignments** — kết quả phân công cuối cùng do PTCM xác nhận, nhánh `RECOMMENDED` hoặc `MANUAL`; mã `AI` còn trong thiết kế/implementation cũ chưa được chuyển đổi.
- **Recommendation** — Top-N mặc định3 có điểm/giải thích; manual tra cứu không tự sinh matching_score. Runs/items snapshot bất biến, hiển thị thời điểm tạo và cho chạy lại; R05 không còn cơ chế xác nhận lịch sử cũ bằng lý do/version. Vẫn kiểm tra điều kiện phân công hiện hành.
- **Yêu cầu cập nhật** — tách profile/skill change requests khỏi hồ sơ chính thức trong thiết kế đích.
- **CV extraction** — `employee_cvs` → `cv_extractions` → `cv_extracted_skills`; là dữ liệu **chờ
  duyệt**, không phải hồ sơ chính thức.
- **Theo dõi thực hiện** — `work_logs` (effort tự khai theo ngày/phút/nội dung, gồm cả làm ở nhà; không chấm công), `task_status_history`,
  `task_comments`, `notifications`. R01 quên Start/Resume không cần duyệt ngoại lệ riêng; nhật ký đã khóa theo hiệu chỉnh chung. R02 nghiệm thu theo đầu ra, không ép khai giờ giả và không nộp kết quả thay.

## 5. Key Constraints

Những điều **không được tự ý thay đổi hoặc giả định**:

1. **AI chỉ đề xuất, không tự phân công.** Không có đường code nào ghi `assignments` mà không đi
   qua thao tác xác nhận của PTCM hiện hành.
2. **Task là đơn vị phân công duy nhất.** `task_groups` chỉ gom nhóm — không gán người, không có
   giờ ước tính, không có yêu cầu kỹ năng.
3. **Kỹ năng trích xuất từ CV phải được Giám đốc duyệt** mới được ghi vào `employee_skills`. Engine
   đề xuất chỉ đọc bản ghi đã duyệt.
4. **Phân quyền hai tầng**: RBAC theo `user_roles` ADMIN/HR/DIRECTOR và ABAC: PM quản lý project của mình,
   EMPLOYEE thực hiện task hiện được giao; Giám đốc chỉnh sửa PM, ADMIN không có quyền thay PM từ quyền quản trị tài khoản.
   R06 PM/Giám đốc đọc cùng official source qua góc nhìn theo mục đích; ADMIN không mặc nhiên duyệt kỹ năng. Quyền phân công/nghiệm thu thuộc PTCM theo dự án; không trả raw engine snapshot hoặc dữ liệu mật. Phần claims/quyền kỹ thuật cũ cần thiết kế lại theo baseline hiện hành.
5. **Không đổi schema** khi chưa đối chiếu baseline hiện hành, thiết kế V2 chưa đồng bộ và `Hình 4.5`/`Hình 4.6` trong `DoAn.pdf`. Quyết định nghiệp vụ mới được ưu tiên nếu báo cáo cũ khác biệt. File
   `docs/thesis-context/Database Schema Specification ... .md` là **bản tóm tắt thiếu cột**, không
   phải nguồn chuẩn.
6. **Không tự chọn công nghệ** cho các mục TBD ở phần 2.
7. **Không sửa file trong `docs/thesis-context/`** — đó là tài liệu gốc, chỉ đọc.
8. Engine chấm điểm phải nằm trong `internal/matching`, **không biết tới HTTP hay DB** (yêu cầu
   mục 2.4.6 về khả năng bảo trì).
9. Toàn bộ chuỗi ghi khi xác nhận phân công phải nằm **trong một transaction** (UC003 alt 16a).
10. Trao đổi và bình luận code bằng **tiếng Việt**; định danh kỹ thuật (bảng, hàm, biến,
    endpoint) bằng **tiếng Anh**.
11. **Quản lý nhánh** Trước khi thêm feature hay fix bug, luôn làm việc trên 1 branch mới. Không được commit thẳng lên main, bug thì nên được đặt là bug/[des], feature thì nên được đặt là feature/[des], nội dung commit thì ngắn gọn, phù hợp, tránh dài dòng
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
| [.claude/docs/ai_recommendation_workflow.md](.claude/docs/ai_recommendation_workflow.md) | Khi làm UC003.1 và gói `internal/matching` — luồng hiện hành và công thức đối chứng cũ, không phải công thức cuối |
| [.claude/docs/cv_extraction_approval_workflow.md](.claude/docs/cv_extraction_approval_workflow.md) | Khi làm M7 — luồng CV và những điểm tài liệu chưa đặc tả |
| [.claude/docs/skill_approval_workflow.md](.claude/docs/skill_approval_workflow.md) | Khi làm phê duyệt kỹ năng MANUAL/CV — căn cứ, tiêu chí, quyền đọc và quyết định |
| [.claude/docs/api_design.md](.claude/docs/api_design.md) | Khi thêm endpoint — quy ước đường dẫn, mã lỗi, định dạng response |
| [.claude/docs/work_effort_workflow.md](.claude/docs/work_effort_workflow.md) | Ghi effort, sửa sai, nghiệm thu, khóa và giới hạn dữ liệu |
| [Baseline hiện hành](docs/business-baseline-2026-10-12.md) | PTCM phân công/nghiệm thu, không nhận task; tra cứu giới hạn, kế hoạch ngoài hệ thống; giữ quyền HR/Giám đốc |
| [.claude/docs/employee_profile_workflow.md](.claude/docs/employee_profile_workflow.md) | Hồ sơ hành chính, HR xét duyệt thông tin cá nhân và ranh giới với kỹ năng |
| [Baseline V2 lịch sử](docs/reviews/system-design-baseline-v2.md) | Bối cảnh thiết kế cũ; chỉ dùng phần không bị quyết định mới thay thế |
| [Review lịch sử V2](docs/reviews/system-design-review-v2-r01-r06.md) | Bối cảnh thiết kế cũ; không dùng quy tắc đã bị baseline mới thay thế |
| [.claude/docs/testing_verification.md](.claude/docs/testing_verification.md) | Khi viết test hoặc cần nghiệm thu một mốc |

Quyết định nghiệp vụ 2026-10-05: Giám đốc chỉnh sửa PM bằng chọn người thay và lưu một lần, không có quy trình bàn giao riêng. Nhân viên tự xác nhận CV hiện tại sau kiểm tra tệp; không phê duyệt CV. Giám đốc chỉ xét duyệt yêu cầu kỹ năng; đổi CV không tự sửa yêu cầu hoặc kỹ năng chính thức. Xem [baseline](docs/business-baseline-2026-10-05.md).
