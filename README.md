# ai-staffing

Hệ thống quản lý dự án, phân rã công việc và phân công nhân sự có hỗ trợ đề xuất
nhân sự phù hợp — đồ án tốt nghiệp.

Điểm cốt lõi: engine chấm điểm chỉ **đề xuất** danh sách nhân sự phù hợp, quyết định
phân công cuối cùng luôn thuộc về Quản lý dự án. Dữ liệu năng lực dùng cho việc đề xuất
bắt buộc phải đã được Quản trị viên xác nhận.

## Công nghệ

| Thành phần | Lựa chọn |
|---|---|
| Backend | Go 1.26, chi, pgx v5 |
| Cơ sở dữ liệu | PostgreSQL (Neon serverless) |
| Migration | golang-migrate dạng thư viện, SQL nhúng bằng `embed.FS` |
| Frontend | Vite + React + TypeScript *(chưa bắt đầu)* |

## Bắt đầu

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
```

Quy tắc phụ thuộc: `transport → service → repository → domain`. Gói `matching` không
biết tới cơ sở dữ liệu hay HTTP, nhờ vậy thay thuật toán chấm điểm không ảnh hưởng
tới luồng nghiệp vụ.

## Tiến độ

- [x] **M0** — Khung dự án, schema, dữ liệu mẫu
- [ ] **M1** — Xác thực, phân quyền, quản lý tài khoản
- [ ] **M2** — Danh mục, hồ sơ nhân sự, duyệt hồ sơ năng lực
- [ ] **M3** — Khởi tạo dự án và phân rã công việc (UC001, UC002)
- [ ] **M4** — Phân công thủ công (UC003, UC003.2)
- [ ] **M5** — Engine đề xuất nhân sự (UC003.1)
- [ ] **M6** — Vòng đời công việc và ghi nhận thời gian (UC004)
- [ ] **M7** — Trích xuất kỹ năng từ CV
