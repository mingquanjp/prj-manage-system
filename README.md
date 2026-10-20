# Hệ thống quản lý dự án hỗ trợ phân công nhân sự dựa trên hồ sơ năng lực

Hệ thống quản lý dự án, phân rã công việc và phân công nhân sự có hỗ trợ đề xuất
nhân sự phù hợp — đồ án tốt nghiệp.

Điểm cốt lõi: engine chấm điểm chỉ **đề xuất** danh sách nhân sự phù hợp, quyết định
phân công cuối cùng luôn thuộc về phụ trách chuyên môn của dự án. Dữ liệu năng lực dùng cho việc đề xuất
bắt buộc phải đã được Giám đốc phê duyệt theo quy trình kỹ năng.

Nghiệp vụ hiện hành: [cơ sở 2026-10-12](docs/business-baseline-2026-10-12.md), đề xuất cho doanh nghiệp nhỏ–vừa. HR quản lý hồ sơ hành chính/CV và duyệt thông tin cá nhân; Giám đốc duyệt kỹ năng, tạo dự án và giao PM. PM chỉ định PTCM và bắt đầu triển khai/theo dõi dự án; PTCM phân công và nghiệm thu, không nhận task trong dự án phụ trách. Không có phân công reviewer thay thế. Kế hoạch chia sẻ bên ngoài, ví dụ Slack. PM/PTCM tra cứu thông tin nghề nghiệp giới hạn, không tải CV gốc người khác. Không tự nghiệm thu, không duyệt CV. Đây là thiết kế nghiệp vụ, chưa phải xác nhận code/CSDL đã đồng bộ.

Thiết kế DB hiện hành: [32 bảng và các quan hệ](docs/design/database-proposal-2026-10-14.md), [DBML trực quan](docs/design/database-proposal-2026-10-14.dbml), [quyết định rà soát](docs/design/database-review-decisions-2026-10-14.md). Cập nhật 15/10/2026; chưa migration/code. Lệnh migrate-up bên dưới chạy schema 000001 cũ, không triển khai bản thiết kế 32 bảng.

## Công nghệ

| Thành phần | Lựa chọn |
|---|---|
| Backend | Go 1.26, chi, pgx v5 |
| Cơ sở dữ liệu | PostgreSQL (Neon serverless) |
| Migration | golang-migrate dạng thư viện, SQL nhúng bằng `embed.FS` |
| Frontend | Vite + React + TypeScript + Tailwind + React Router; primitives shadcn/Radix; pilot local đã duyệt, chưa API/auth |

## Chạy frontend mock — không cần DB

Node 24.x, pnpm; phiên bản đã kiểm tra và giới hạn tại [báo cáo pilot](web/DESIGN-PILOT.md).

```bash
cd web
pnpm install --frozen-lockfile
pnpm dev                 # http://localhost:5173
pnpm check               # typecheck → test demo → build
```

Pilot chưa phải toàn bộ sản phẩm; reload khôi phục dữ liệu mẫu. Không cần chạy migration/seed để xem UI.

## Backend skeleton — lệnh vận hành, không tự chạy để sửa UI

Các lệnh dưới đây kết nối/thay đổi DB theo cấu hình. Chỉ dùng khi được yêu cầu vận hành backend và đã xác nhận đúng môi trường; migration hiện có không phải bản thiết kế đích 32 bảng.

```bash
cp .env.example .env       # rồi điền DATABASE_URL của Neon
make migrate-up            # tạo 22 bảng và 15 kiểu ENUM
make seed                  # nạp dữ liệu mẫu
make dev                   # chạy server tại http://localhost:8080
```

Kiểm tra nhanh:

```bash
curl -s localhost:8080/api/healthz
```

Tài khoản mẫu dùng chung mật khẩu `Password@123`:
`admin` (Quản trị viên), `pm.hoa` (Quản lý dự án), `dev.an` (Nhân viên).

> **Lưu ý về Neon**: `DATABASE_URL` phải trỏ tới endpoint **trực tiếp**, không phải
> endpoint có `-pooler`. pgx dùng prepared statement nên không tương thích hoàn toàn
> với transaction pooling, và migration cần advisory lock mà pooler không giữ được.

Chạy `make help` để xem toàn bộ lệnh.

## Cấu trúc

```
cmd/            server, migrate, seed
internal/
  config/       cấu hình từ biến môi trường
  domain/       thực thể và quy tắc nghiệp vụ thuần
  service/      điều phối use case, quản lý transaction
  repository/   truy cập cơ sở dữ liệu
  transport/    HTTP: định tuyến, middleware, DTO
  matching/     engine chấm điểm đề xuất nhân sự
migrations/     SQL migration
db/seed/        SQL dữ liệu mẫu
docs/           tài liệu phân tích thiết kế
web/            frontend local và pilot thiết kế
```

Quy tắc phụ thuộc chi tiết theo [architecture](.claude/docs/architecture.md): transport gọi service; service/domain không import Postgres adapter, adapter triển khai interface do service khai báo. Gói `matching` không biết tới cơ sở dữ liệu hay HTTP.

## Tiến độ

- [x] **M0** — Khung dự án, schema, dữ liệu mẫu
- [ ] **M1** — Xác thực, phân quyền, quản lý tài khoản
- [ ] **M2** — Danh mục; HR quản lý hồ sơ/duyệt thông tin cá nhân; Giám đốc duyệt kỹ năng
- [ ] **M3** — Khởi tạo dự án và phân rã công việc (UC001, UC002)
- [ ] **M4** — Phân công thủ công (UC003, UC003.2)
- [ ] **M5** — Engine đề xuất nhân sự (UC003.1)
- [ ] **M6** — Vòng đời công việc và ghi nhận thời gian (UC004)
- [ ] **M7** — Trích xuất kỹ năng từ CV

Quyết định nghiệp vụ 2026-10-05: Giám đốc chỉnh sửa PM bằng chọn người thay và lưu một lần, không có quy trình bàn giao riêng. Nhân viên tự xác nhận CV hiện tại sau kiểm tra tệp; không phê duyệt CV. Giám đốc chỉ xét duyệt yêu cầu kỹ năng; đổi CV không tự sửa yêu cầu hoặc kỹ năng chính thức. Xem [baseline](docs/business-baseline-2026-10-05.md).
