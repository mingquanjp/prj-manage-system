# TaskAssign — Quản lý dự án và hỗ trợ phân công nhân sự

Đồ án tốt nghiệp xây dựng hệ thống quản lý dự án, công việc và hồ sơ năng lực, hỗ trợ phụ trách chuyên môn lựa chọn người thực hiện phù hợp.

Hướng nghiệp vụ dành cho doanh nghiệp nhỏ–vừa: **năng lực đã công nhận → yêu cầu công việc → tra cứu/gợi ý → xác nhận phân công → thực hiện → nghiệm thu**. Hệ thống hỗ trợ quyết định, không tự động giao việc.

> **Hiện trạng — 26/10/2026:** đang phát triển frontend bằng dữ liệu mô phỏng và duy trì backend skeleton. Chưa có sản phẩm nghiệp vụ chạy xuyên suốt với API/DB. Thiết kế DB đích gồm 33 bảng, chưa được triển khai bằng migration.

## Mục lục

- [Hiện trạng triển khai](#hiện-trạng-triển-khai)
- [Phạm vi nghiệp vụ](#phạm-vi-nghiệp-vụ)
- [Công nghệ](#công-nghệ)
- [Chạy frontend demo](#chạy-frontend-demo)
- [Backend và database](#backend-và-database)
- [Kiểm tra chất lượng](#kiểm-tra-chất-lượng)
- [Cấu trúc repository](#cấu-trúc-repository)
- [Tài liệu và nguồn chuẩn](#tài-liệu-và-nguồn-chuẩn)
- [Quy trình phát triển và nhánh Git](#quy-trình-phát-triển-và-nhánh-git)

## Hiện trạng triển khai

| Thành phần | Đã có | Chưa có / giới hạn |
|---|---|---|
| Frontend dự án | Shell, danh sách/chi tiết, tạo dự án kèm PM, chỉ định PTCM và bắt đầu dự án | Dữ liệu trong bộ nhớ; chưa API, đăng nhập hoặc phân quyền server |
| Cấu trúc công việc | Tạo/sửa nhóm và đặc tả task trước lần giao đầu; yêu cầu kỹ năng | Chưa dựng đầy đủ thao tác sau giao và vòng đời dự án |
| Phân công | Tra cứu thủ công, gợi ý mẫu, chọn vào nháp và xác nhận cập nhật state chung | Điểm/hạng minh họa, không phải engine thật; chưa transaction DB |
| Các phân hệ khác | Đã có đặc tả nghiệp vụ/UI | Chưa dựng đầy đủ hồ sơ, CV, duyệt kỹ năng, nhật ký, nghiệm thu và quản trị trong frontend local |
| Backend | Go HTTP skeleton, `GET /api/healthz`, cấu hình kết nối, công cụ migration/seed và test an toàn | Chưa triển khai các service nghiệp vụ hoặc API tích hợp frontend |
| Database | Migration/seed legacy; dictionary và DBML thiết kế đích 33 bảng | Schema legacy khác thiết kế đích; chưa chuyển đổi DB |
| Harness phát triển | Chỉ dẫn agent, mục lục nguồn, kiểm tra đường dẫn/nguồn, test và lệnh kiểm tra frontend | Không tự kiểm chứng ngữ nghĩa nghiệp vụ hoặc thay kiểm thử tích hợp/E2E |

Phạm vi chi tiết và bằng chứng kiểm tra từng đợt được ghi tại [web/README.md](web/README.md). Không dùng prototype Make cũ hoặc sự tồn tại của một bảng để kết luận chức năng đã được triển khai.

## Phạm vi nghiệp vụ

Đây là **thiết kế đích đã chốt**, không phải danh sách tính năng đều đã chạy.

### Vai trò và trách nhiệm

| Vai trò / trách nhiệm | Phạm vi chính |
|---|---|
| Quản trị viên — Admin | Quản lý tài khoản; tạo tài khoản kèm hồ sơ nhân viên ban đầu trong cùng thao tác |
| Nhân sự — HR | Quản lý hồ sơ hành chính/CV; duyệt cập nhật thông tin cá nhân; quản lý phòng ban, vị trí và cấp bậc |
| Giám đốc — Director | Quản lý nhóm/danh mục kỹ năng, xét duyệt kỹ năng có căn cứ; tạo dự án kèm chọn PM và thay PM |
| Quản lý dự án — PM | Chỉ định PTCM; theo dõi, bắt đầu, hoàn thành, đóng hoặc hủy dự án theo điều kiện |
| Phụ trách chuyên môn — PTCM | Thiết lập nhóm/task, yêu cầu kỹ năng; phân công, thay/thu hồi người thực hiện và nghiệm thu; duyệt hiệu chỉnh nhật ký sau khóa |
| Nhân viên | Quản lý thông tin/CV của mình, gửi đề nghị cập nhật; thực hiện task được giao, ghi nhật ký và nộp kết quả |

Admin, HR và Director là vai trò hệ thống. **PM/PTCM là trách nhiệm theo từng dự án**, không phải role toàn hệ thống; một người có thể kiêm nhiệm. Quyền cá nhân của nhân viên không đồng nghĩa được xem hồ sơ người khác.

### Nguyên tắc cốt lõi

- Dự án có cấu trúc **nhóm công việc → công việc**, không có cây task nhiều cấp. Task là đơn vị phân công và nghiệm thu.
- Giám đốc chọn PM ngay khi tạo dự án; người tạo không tự trở thành PM. PM không kiêm PTCM không có quyền phân công/nghiệm thu.
- PTCM không nhận/thực hiện task trong chính dự án phụ trách; chỉ định PTCM phải tránh công việc đang được giao và xung đột với đầu ra cũ của chính người đó. Không có reviewer thay thế.
- Nguồn ứng viên là nhân sự toàn công ty đủ điều kiện, không cần lập đội dự án trước. Nhánh thủ công và gợi ý dùng chung dữ liệu chính thức; quyết định cuối cùng thuộc PTCM.
- Kỹ năng chỉ trở thành năng lực chính thức sau xét duyệt của Giám đốc. CV/trích xuất không tự công nhận kỹ năng; không có bước phê duyệt tệp CV.
- Tra cứu PM/PTCM giới hạn ở thông tin nghề nghiệp, năng lực, kinh nghiệm/kết quả và tải tổng hợp được phép. Không mở dữ liệu liên hệ, địa chỉ, CV gốc hoặc toàn bộ task/nhật ký dự án khác.
- Nhân viên tự ghi nhật ký theo ngày, thời lượng và nội dung; không dùng timer, không tự sinh giờ từ trạng thái. Nghiệm thu đầu ra tách khỏi độ đầy đủ của nhật ký.
- Địa chỉ phải bổ sung khi hoàn thiện hồ sơ nhưng không bắt Admin nhập lúc tạo. Chủ hồ sơ đề nghị cập nhật, HR duyệt hoặc sửa trực tiếp; thiếu địa chỉ không chặn làm việc.
- Phương pháp gợi ý định hướng luật và chấm điểm giải thích được, không bắt buộc học máy. Công thức cuối, miền trọng số và tham số còn cần chốt; không coi số demo là ràng buộc nghiệp vụ.

Không mở rộng thành HRM đầy đủ, chấm công/lương, sprint, duyệt kế hoạch, tối ưu lịch, tích hợp GitHub/Slack, OCR hoặc tự động giao việc. Xem [phạm vi sản phẩm](.claude/docs/product_scope.md) và [phân quyền](.claude/docs/roles_permissions.md) để biết ranh giới đầy đủ.

## Công nghệ

| Thành phần | Công nghệ / trạng thái |
|---|---|
| Frontend | React 19, TypeScript, Vite, Tailwind CSS 4, React Router |
| UI | Primitives tùy biến theo mô hình shadcn/Radix; token màu ấm, bố cục workspace lấy cảm hứng từ Notion |
| Backend | Go; chi v5 cho HTTP, pgx v5 cho PostgreSQL |
| Database | PostgreSQL; cấu hình hiện có hỗ trợ Neon, không hardcode nhà cung cấp |
| Migration | golang-migrate, SQL nhúng bằng `embed.FS` |
| Kiểm tra | Node test runner cho harness/model demo; Go test/vet/build cho backend |

Backend hiện yêu cầu **Go 1.26.5** theo `go.mod`. Frontend dùng **Node 24.x**, **pnpm 11.x**, pin **pnpm 11.19.0** tại `web/package.json`; giữ lockfile khi cài đặt. Chưa tích hợp API/auth, TanStack Query hoặc nhúng frontend vào Go.

## Chạy frontend demo

Không cần tài khoản thật, backend hoặc database.

Từ thư mục gốc repository:

```bash
cd web
pnpm install --frozen-lockfile
pnpm dev
```

Mở [http://127.0.0.1:5173](http://127.0.0.1:5173). Dev server dùng `strictPort`: nếu cổng bận, kiểm tra tiến trình hiện có thay vì tự dừng tiến trình khác.

### Cách xem luồng chính

1. Dùng **Danh tính mô phỏng** để đổi người xem; đây chỉ là điều khiển prototype, không phải chức năng tự cấp quyền.
2. Với người có trách nhiệm PTCM, mở dự án để xem tab **Công việc / Cấu trúc**, tạo nhóm/task hoặc thử phân công.
3. Với Giám đốc, tạo dự án và chọn PM. Chuyển sang đúng PM để chỉ định PTCM/bắt đầu khi đủ điều kiện.

PM chỉ theo dõi và PTCM có hành động khác nhau; người không có trách nhiệm dự án có thể không thấy dự án. Màn hình công việc cá nhân chưa được dựng trong đợt local này.

**Giới hạn demo:**

- Dữ liệu dùng chung trong phiên nhưng chỉ lưu trong bộ nhớ. Reload hoặc khôi phục dữ liệu mẫu làm mất các thay đổi mô phỏng.
- Gợi ý dùng fixture; dữ liệu ứng viên/điểm không phải kết quả engine thật.
- Cảnh báo bỏ nháp hiện áp dụng cho tab trong workspace và nút quay lại danh sách công việc bên trong; chưa bảo vệ sidebar, browser back hoặc reload.
- Kiểm tra quyền ở frontend phục vụ trải nghiệm, không thay kiểm tra quyền server.

## Backend và database

### Phân biệt hiện trạng và thiết kế đích

[Migration 000001](migrations/000001_init.up.sql) và dữ liệu seed hiện tại thuộc schema legacy. Tài khoản/role, quy trình kỹ năng và cấu trúc dữ liệu trong đó **không đại diện đầy đủ cho nghiệp vụ mới**.

Thiết kế đích được mô tả bằng [bản hợp nhất 33 bảng](docs/design/database-proposal-2026-10-14.md), dictionary và DBML. Chưa có migration triển khai thiết kế này; không sửa migration đã áp dụng hoặc chạy reset để giả lập việc nâng cấp.

### Chạy skeleton khi cần làm backend

Chỉ thực hiện khi đã xác nhận môi trường và đúng DB phát triển:

1. Tạo cấu hình local theo `.env.example`, không ghi đè `.env` đang có hoặc commit thông tin kết nối.
2. Bảo đảm DB đã có schema legacy phù hợp. Nếu cần migration/seed, xử lý riêng theo bảng lệnh bên dưới.
3. Chạy từ thư mục gốc:

```bash
make dev
```

Mặc định API ở [http://localhost:8080](http://localhost:8080). Khi server đã chạy:

```bash
curl http://localhost:8080/api/healthz
```

Health kiểm tra kết nối DB, trả trạng thái tổng quát. Health thành công **không chứng minh** schema đã khớp thiết kế 33 bảng hoặc API nghiệp vụ đã có.

| Lệnh vận hành | Tác động / điều kiện |
|---|---|
| `make migrate-version` | Kết nối DB và đọc version migration |
| `make migrate-up` | Áp dụng migration SQL hiện có; không triển khai thiết kế DB mới |
| `APP_ENV=dev make seed` | Ghi dữ liệu mẫu legacy, không dùng cho DB thật |
| `APP_ENV=dev make migrate-down` | Rollback một migration; có thể mất dữ liệu |
| `APP_ENV=dev make migrate-reset` | Rollback toàn bộ rồi chạy lại migration; phá dữ liệu |

Seed và rollback yêu cầu `APP_ENV=dev` được đặt rõ trong môi trường tiến trình **trước khi đọc cấu hình DB**; không tự lấy giá trị mặc định hoặc nạp `.env` để cấp phép. Cờ dev không chứng minh DB đích an toàn. Reset dùng `down-all` để xử lý cả ENUM, không còn lệnh `drop`.

Code hiện hỗ trợ endpoint Neon `-pooler`: runtime chọn simple protocol, migration chuyển sang endpoint trực tiếp. Đây là cơ chế hiện có, không bảo đảm mọi cấu hình pooler. Frontend demo không cần bất kỳ lệnh DB nào ở trên.

## Kiểm tra chất lượng

### Frontend

Từ `web/`:

```bash
pnpm check
```

Lệnh chạy lần lượt: test harness → kiểm kê nguồn → TypeScript → test model demo → build Vite. Có thể chạy riêng `pnpm typecheck`, `pnpm test`, `pnpm build`. Chưa có cổng lint hoặc test React/E2E tự động.

### Backend

Từ thư mục gốc:

```bash
go vet ./...
go test ./... -count=1
go build ./...
```

Các test backend hiện có dùng fake/biến môi trường cho guard, health và forwarded IP; không kết nối DB thật. Chưa có kiểm thử migration/seed hoặc tích hợp nghiệp vụ trên DB. Khi bổ sung integration test phải dùng DB test riêng được cho phép.

### Tài liệu và harness

```bash
node --test scripts/check-harness.test.mjs
node scripts/check-harness.mjs
git diff --check
```

Harness kiểm tra điểm vào, nguồn cần thiết, đường dẫn và Git ignore; không chứng minh nội dung nghiệp vụ hoặc UI/API đúng. Chế độ `--staged` dành cho bước chuẩn bị bàn giao Git, không tự stage/commit.

`make check` có chạy fmt/tidy và có thể sửa file; không dùng như lệnh kiểm tra chỉ đọc hoặc thay cổng frontend. Chi tiết mức kiểm tra theo loại thay đổi tại [web/README.md](web/README.md).

## Cấu trúc repository

Các thư mục mã nguồn hiện có:

```text
.
├── web/
│   ├── src/
│   │   ├── components/           # UI primitives và thành phần dùng chung
│   │   ├── demo/                 # Model, fixture và test dữ liệu mô phỏng
│   │   └── features/             # Dự án và phân công
│   └── README.md                 # Thiết kế đã duyệt, quy trình và sổ công việc FE
├── cmd/
│   ├── server/                  # Khởi chạy HTTP server
│   ├── migrate/                 # Công cụ migration
│   └── seed/                    # Công cụ dữ liệu mẫu legacy
├── internal/
│   ├── config/                  # Cấu hình và guard vận hành
│   ├── repository/postgres/     # Kết nối PostgreSQL
│   └── transport/http/          # Router, health và test
├── migrations/                  # SQL legacy và embed
├── db/seed/                     # SQL dữ liệu mẫu legacy
├── scripts/                     # Kiểm tra harness
├── .claude/docs/                # Workflow và đặc tả triển khai
├── docs/                        # Mục lục nguồn, quyết định và thiết kế DB/UI
└── AGENTS.md                    # Chỉ dẫn agent; CLAUDE.md trỏ cùng nguồn
```

Các tầng domain/service/matching thuộc [kiến trúc đích](.claude/docs/architecture.md), chưa được triển khai thành package nghiệp vụ. Matching sẽ độc lập DB/HTTP; không coi cấu trúc dự kiến là code đã có.

## Tài liệu và nguồn chuẩn

Bắt đầu tại **[docs/README.md](docs/README.md)** để chọn đúng tài liệu theo miền.

| Cần tìm | Nguồn |
|---|---|
| Phạm vi, vai trò và workflow | [Phạm vi](.claude/docs/product_scope.md), [phân quyền](.claude/docs/roles_permissions.md), [dự án/phân công](.claude/docs/project_task_assignment_workflow.md) |
| Quyết định DB hiện hành | [Sổ quyết định — ưu tiên phần hiện hành/Bàn giao](docs/design/database-review-decisions-2026-10-14.md) |
| Chốt tạo tài khoản và địa chỉ | [Bổ sung 20/10/2026](docs/business-baseline-2026-10-20.md) |
| Bảng, quan hệ và trường dữ liệu | [Tổng quan 33 bảng](docs/design/database-proposal-2026-10-14.md), [dictionary 5A](docs/design/database-step5a-core-dictionary-2026-10-14.md), [dictionary 5B](docs/design/database-step5b-supporting-dictionary-2026-10-14.md), [DBML](docs/design/database-proposal-2026-10-14.dbml) |
| Chuẩn bị chuyển đổi schema | [Mapping hiện trạng → đích](docs/design/database-step5c-migration-map-2026-10-14.md) |
| Frontend và harness | [web/README.md](web/README.md), [AGENTS.md](AGENTS.md) |
| Kiến trúc, API, bảo mật | [Kiến trúc](.claude/docs/architecture.md), [API contract](.claude/docs/api_design.md), [bảo mật](.claude/docs/security.md) |

README là bản giới thiệu và hướng dẫn vào dự án, **không thay nguồn quyết định**. Quyết định đã xác nhận và phần hiện hành/Bàn giao ưu tiên nội dung cũ bị thay thế. Code legacy, Make/export và nghiên cứu không tự ghi đè nghiệp vụ mới; tài liệu lịch sử không phải checklist tính năng đã hoàn thành.

## Quy trình phát triển và nhánh Git

- `main`: bản tích hợp hiện hành; xem nhánh này khi đánh giá trạng thái mới nhất.
- Mỗi đợt feature/fix có phạm vi riêng, đối chiếu nguồn → triển khai → kiểm tra phù hợp → review → bàn giao; không tự mở rộng nghiệp vụ/schema/công thức.
- Chỉ dẫn cho agent tại [AGENTS.md](AGENTS.md); chuẩn UI và sổ từng đợt frontend tại [web/README.md](web/README.md).
- Commit dùng tiếng Anh; giao diện và bình luận code dùng tiếng Việt. Không đưa bí mật, `node_modules` hoặc `dist` vào Git.
- Các nhánh mốc dưới đây được giữ để tham khảo và trình bày; không tự xóa khi chưa được đồng ý.

| Nhánh mốc | Nội dung |
|---|---|
| `feature/ui-business-baseline-sync` | Mốc đồng bộ tài liệu nghiệp vụ và thiết kế DB |
| `feature/harness-handoff-checks` | Mốc xây dựng harness tối thiểu |
| `feature/project-setup-ui` | Mốc UI thiết lập dự án, nhóm/task và phân công |

Các mốc đã nằm trong lịch sử `main`, không phải feature còn chờ tích hợp.

<details>
<summary>Tag bảo toàn ba nhánh legacy — không dùng làm nguồn triển khai</summary>

| Tag | Commit gốc / ý nghĩa |
|---|---|
| `archive/documentation-inconsistencies-2026-10-21` | `cfa26cd` — tài liệu cũ đã được thay thế; mô tả pooling còn phù hợp được sửa riêng |
| `archive/two-level-task-hierarchy-2026-10-21` | `704b8e7` — giữ lịch sử quyết định hai tầng; không lấy lại sửa trực tiếp migration 000001 |
| `archive/migrate-reset-and-seed-hardening-2026-10-21` | `1ff7b0e` — sửa an toàn được áp dụng có chọn lọc; không lấy lại seed/role/quy trình duyệt cũ |

Các tag cho phép truy vết hoặc khôi phục commit cũ, không có nghĩa toàn bộ nội dung đó đã được merge hoặc còn phù hợp nghiệp vụ.

</details>
