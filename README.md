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

### Proposal auth — cập nhật 06/11/2026

**Tiếp nối 07/11/2026:** người dùng đã chạy import và check trên Jira DATN/board 2; log được đối chiếu xác nhận Epic DATN-1, 20 Task DATN-2–DATN-21, 47 Blocks và Rank khớp. Các câu chưa giao Jira bên dưới ghi nhận giai đoạn trước bàn giao; chưa có auth runtime. Liên kết Development với GitHub còn cần xác minh sau push/PR có mã ticket.

- [Feature đầu: đăng nhập và phiên truy cập](docs/proposals/backend-auth-session.md), D01 phối hợp D02: năm route login/me/csrf/logout/đổi mật khẩu lần đầu; chưa gồm tạo tài khoản, gửi thư, reset hoặc API dự án.
- Người dùng đã duyệt A01–A09: email công ty được trim/so không phân biệt hoa thường; phiên 8 giờ tuyệt đối; cookie/CSRF; mật khẩu mới ≥15 ký tự Unicode, ≤72 byte UTF-8; phiên hạn chế và thu hồi mọi phiên khi đổi lần đầu. Xem [baseline](docs/business-baseline-2026-10-20.md), security và proposal để biết giới hạn/phần hoãn.
- Hướng test là PostgreSQL local riêng (Docker, major 18), dữ liệu giả; hướng DB là migration tiến thêm, giữ 000001. Chưa dựng môi trường hoặc duyệt mapping/SQL auth cụ thể. Riêng 000002 thêm cờ đã diễn tập nhánh Neon tạm, chưa áp production theo mapping 5C.
- Proposal theo 10 mục template; WBS 9 nhóm/33 đơn vị công việc (gom thành [20 ticket theo đầu ra](docs/proposals/backend-auth-session-tickets.md)), 105 giờ cơ sở và 21 giờ dự phòng đề xuất; không phải thời gian còn lại hoặc cam kết ngày lịch. Đã soạn [OpenAPI auth](docs/api/auth.openapi.json) và ports tại proposal mục 5.3; chưa giao task Go/Jira hoặc có code/test auth runtime. Các sơ đồ/kiểm tra tài liệu không chứng minh runtime đã hoạt động.

### Phân miền backend — 05/11/2026

- [Bản đồ miền](docs/product/business-domains.md): đề xuất 6 miền nghiệp vụ và 1 phần hỗ trợ thông báo trong cùng backend/DB; ánh xạ đủ 62 FR, trách nhiệm 33 bảng thiết kế, phụ thuộc và các luồng cần phối hợp. Nguồn là FR/workflow, kiến trúc và quyết định DB hiện hành; không sao chép lại AC hay đổi schema.
- Review các ranh giới: D06 đề xuất/D04 chốt giao việc; D04 nghiệm thu/D05 hiệu chỉnh; D02 hành chính/D03 công nhận kỹ năng; D01 quyền hệ thống/D04 trách nhiệm dự án. Phân miền không thành microservice hoặc cho phép commit rời khi nghiệp vụ cần nguyên tử. Các điểm mở O01–O09 giữ tại FR.
- Đã kiểm tra mỗi FR có một nơi chủ trì, mỗi bảng DBML có một đầu mối, số lượng từng nhóm và liên kết; 9 test harness, kiểm kê 202 nguồn và `git diff --check` đạt. Chỉ kiểm kê/review tài liệu, chưa kiểm chứng code/DB/concurrency. Chưa chốt package/API/feature đầu/WBS/Jira hoặc sửa Go/SQL/FE; giữ thay đổi có sẵn, chưa stage/commit/push. Sau duyệt bản đồ mới chọn feature và viết proposal.

### Chuẩn bị yêu cầu backend — 05/11/2026

- [Bản FR/AC dự thảo](docs/product/functional-requirements.md): sau review có 62 FR đang dùng, 1 quy tắc BR và 2 yêu cầu dùng chung CR. Ba ID cũ chuyển loại, ba ID mới tách hành vi khác mục đích; có bảng ánh xạ, không tăng phạm vi nghiệp vụ. Đối chiếu UC dạng văn bản, phân biệt đích với FE mô phỏng/backend skeleton, ghi riêng điểm mở. Các nhóm đọc chưa phải miền/package Go đã chốt.
- Phạm vi đợt tài liệu: lập yêu cầu trước; không sửa Go/SQL/DBML, chọn dependency, chạy DB hoặc tạo Jira. Tiếp theo phân miền → chọn feature → proposal/API/WBS/estimate → chốt phần tự viết Go → triển khai từng task. Chưa bắt đầu bài thực hành config.
- Review nội dung: phân biệt đổi mật khẩu/reset, tra cứu tài khoản/cấp quyền và effort/tải; chuyển địa chỉ thành BR và lịch sử/chống lặp thành CR có phạm vi áp dụng. Bớt chi tiết lưu trữ/API/bố cục trong FR, giữ dẫn chiếu contract đã chốt; cập nhật bảng bao phủ UC và điểm mở theo ID mới. Chênh lệch phạm vi chọn tất cả giữa workflow/contract FE vẫn ghi mở, không tự sửa nguồn.
- Kiểm tra sau chuẩn hóa: 9 test harness và kiểm kê 201 nguồn đạt; kiểm tra cấu trúc/nguồn/AC của 62 FR + 1 BR + 2 CR, đủ ánh xạ UC/ID cũ–mới và đường dẫn nội bộ; `git diff --check` đạt. Các kiểm tra cấu trúc không thay review ngữ nghĩa. Không chạy lại UI/Go/DB vì không sửa code; giữ thay đổi Git/harness có sẵn, chưa stage/commit/push.

### Chuẩn bị Git/harness backend — 05/11/2026

- Mục tiêu đợt đầu: sửa phạm vi ignore binary và kiểm kê nguồn backend để entrypoint server không bị bỏ sót khi bàn giao Git.
- Nguồn: [chỉ dẫn agent](AGENTS.md), [kiến trúc](.claude/docs/architecture.md), [kiểm thử](.claude/docs/testing_verification.md); giữ nguyên nghiệp vụ hiện hành.
- Ngoài phạm vi: sửa config/URL DB/log khởi động, API nghiệp vụ, schema, dependency, chạy server/migration/seed hoặc kết nối DB. Config là phần người thực hiện tự viết với agent hướng dẫn và review, chưa hoàn thành trong đợt Git/harness.
- Tiêu chí đạt: binary ở gốc vẫn bị ignore; `cmd/server/main.go` không bị ignore; harness kiểm kê nguồn Go/SQL cùng module/Makefile và từ chối nguồn chưa có trong index khi dùng `--staged`; test hồi quy harness đạt.
- Bài thực hành **dự kiến, chưa bắt đầu**: validation config và table-driven test không đọc `.env`. Phạm vi kỹ thuật/tiêu chí sẽ chốt trong proposal/WBS và phân công task trước khi người dùng tự viết; không coi nội dung gợi ý trước đây là lệnh triển khai ngay. Cách nạp `.env`, xử lý hostname và che lỗi khởi động là các mục nền còn cần thiết kế riêng.
- Kết quả Git/harness: 9 test harness đạt; kiểm kê 200 file nguồn đạt; `git diff --check` đạt. Entry point server hiện được nhận diện là untracked, cần đưa vào commit khi người dùng yêu cầu. Đây chưa là xác nhận nguồn đã được commit/clone đầy đủ; không thay kiểm tra Go hoặc DB. Không sửa mã Go trong lượt này.

### Phân biệt hiện trạng và thiết kế đích

[Migration 000001](migrations/000001_init.up.sql) và dữ liệu seed hiện tại thuộc schema legacy. Tài khoản/role, quy trình kỹ năng và cấu trúc dữ liệu trong đó **không đại diện đầy đủ cho nghiệp vụ mới**.

Thiết kế đích được mô tả bằng [bản hợp nhất 33 bảng](docs/design/database-proposal-2026-10-14.md), dictionary và DBML. Chưa có migration triển khai thiết kế này; không sửa migration đã áp dụng hoặc chạy reset để giả lập việc nâng cấp.

### Chạy skeleton khi cần làm backend

Chỉ thực hiện khi đã xác nhận môi trường và đúng DB phát triển:

1. Tạo cấu hình local theo `.env.example`, không ghi đè `.env` đang có hoặc commit thông tin kết nối.
2. Bảo đảm DB đã có schema legacy phù hợp. Nếu cần migration/seed, xử lý riêng theo bảng lệnh bên dưới.
3. Chạy từ thư mục gốc:

```bash
APP_ENV=dev make dev
```

DATN-4: `APP_ENV` phải được đặt rõ trong môi trường chạy (`dev`/`prod`); prod không đọc `.env`. Biến môi trường ưu tiên file, kể cả giá trị rỗng. Server yêu cầu `DATABASE_URL` và `APP_ORIGIN`; port mặc định `8080`, thời hạn phiên mặc định `8h` chỉ khi chưa khai báo. Giá trị đã khai báo sai/rỗng không được tự thay. Origin dev HTTP chỉ localhost/IP loopback, prod yêu cầu HTTPS. File dev không có được bỏ qua nếu cấu hình môi trường đã đủ; file lỗi đọc/định dạng bị từ chối. Xem định dạng dotenv tại `.env.example`.

`LoadServer()` kiểm tra trước kết nối DB; migrate/seed dùng `Load()` không yêu cầu Origin/TTL/port, vẫn bắt buộc `APP_ENV` và `DATABASE_URL`. Ví dụ lệnh đọc schema: `APP_ENV=dev make migrate-version` (chỉ chạy khi đã xác nhận DB đích). Các lệnh `make migrate-*` bên dưới cũng cần chọn môi trường tường minh.

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

DATN-5: loader parse `DATABASE_URL` thành `Config.Database` (`RuntimeURL`, `DirectURL`, `Pooled`) trước khi cấu hình được dùng. Chỉ hỗ trợ URL postgres/postgresql một host/IP và port hợp lệ nếu có; không keyword DSN, Unix socket, nhiều host, fragment hoặc query override endpoint/định danh/service. Không log cấu trúc này vì có credential.

Chỉ hostname Neon AWS đúng mẫu `ep-…-pooler.region.aws.neon.tech` được nhận diện pooled: runtime giữ nguyên URL/simple protocol, migration đổi đúng phần hostname sang direct, giữ nguyên credential/port/path/query. Host khác giữ nguyên; nếu pooler ngoài mẫu thì phải cung cấp direct URL cho migration, không suy đã được tự chuyển.

Lỗi parse/tạo pool/ping startup trả nhóm an toàn `DB_URL_INVALID` / `DB_POOL_INIT_FAILED` / `DB_CONNECT_FAILED`, không giữ raw driver error trong error chain hoặc log. Chưa có bằng chứng kết nối DB thật từ unit tests; xử lý lỗi SQL khi chạy migration/seed vẫn là legacy ngoài phạm vi đợt startup. Frontend demo không cần bất kỳ lệnh DB nào ở trên.

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

- Theo [workflow solo Jira–GitHub](docs/ai/workflow.md) và [mẫu giao việc](docs/ai/task-template.md); PR dùng [template chung](.github/pull_request_template.md). Ví dụ nhánh `feature/DATN-2-auth-contract`, commit/PR `docs(auth): finalize authentication contract [DATN-2]`; mã `AUTH-T01` là mã nội bộ, không thay mã Jira.
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

**Tiếp nối auth 06/11/2026:** đã chốt cấp tài khoản mới bằng mật khẩu tạm qua email, buộc đổi trước nghiệp vụ. Xem quyết định tiếp nối trong `docs/business-baseline-2026-10-20.md` và mục 4.1 proposal auth; DBML/dictionary đã chốt thêm users.must_change_password; mật khẩu tạm chưa có hạn. Proposal đã bổ sung năm route và WBS 33 task/105 giờ cơ sở; dự phòng 21 giờ còn đề xuất. A02–A09 đã duyệt; OpenAPI/ports đã có bản soạn để review; SQL/mapping legacy và fixture cụ thể còn phải hoàn thiện. Migration riêng chỉ thêm cờ, chưa cập nhật đủ 33 bảng hoặc triển khai gửi thư.
