# Kiến trúc mã nguồn

Neo tài liệu: mục 2.4.6 (Khả năng bảo trì) — "phần AI đề xuất nên được tách khỏi thao tác xác nhận
phân công để có thể thay đổi thuật toán mà không làm ảnh hưởng lớn đến luồng nghiệp vụ chính".

## 1. Phân tầng

```
transport (HTTP)  →  service  →  repository  →  domain
                          ↓
                      matching        (thuần hàm, không I/O)
```

Quy tắc phụ thuộc, **một chiều**:

- `domain` không import gói nào khác trong dự án. Chỉ chứa entity, enum, và quy tắc nghiệp vụ thuần.
- `repository` biết `domain` và pgx. Trả về struct của `domain`, không trả kiểu của pgx ra ngoài.
- `service` điều phối use case và **quản lý transaction**. Không biết HTTP.
- `transport` chỉ làm: parse request → validate định dạng → gọi service → map lỗi sang HTTP status.
- `matching` nhận struct thuần từ `domain`, trả kết quả chấm điểm. **Không** biết DB, không biết HTTP.

Sơ đồ này cũng chính là biểu đồ gói UML cần vẽ cho mục 4.1 của báo cáo.

## 2. Cấu trúc thư mục

Hiện có trong repo:

```
cmd/
  server/main.go              wiring, đọc config, chạy HTTP server, graceful shutdown
  migrate/main.go             up | down | drop | version
  seed/main.go                nạp dữ liệu mẫu, idempotent
internal/
  config/config.go            đọc env + .env; IsPooledEndpoint(), DirectURL()
  repository/postgres/db.go   NewPool(ctx, url, simpleProtocol)
  transport/http/router.go    chi router, middleware, GET /api/healthz
migrations/                   000001_init.up.sql / .down.sql + embed.go
db/seed/                      01_reference.sql, 02_employee_skills.sql + embed.go
web/                          rỗng — frontend chưa khởi tạo
```

Dự kiến bổ sung theo lộ trình:

```
internal/
  domain/{user,employee,skill,project,task,assignment}/
  service/
  repository/postgres/        thêm các repo cụ thể
  transport/http/             thêm handler, middleware auth/rbac, dto, validate
  matching/                   scorer.go rules.go explain.go weights.go
  storage/                    interface lưu file CV + impl (M7)
  notify/                     tạo bản ghi notifications
  config/workcalendar.yaml    lịch làm việc chuẩn cho UC004, nằm cạnh config.go
web/src/{pages,components,api,hooks,types}
```

Khi tạo package mới, đặt đúng tầng theo quy tắc trên. Nếu một logic không rõ thuộc tầng nào,
mặc định đưa vào `service` — không bao giờ nhét vào handler.

## 3. Kết nối cơ sở dữ liệu

`postgres.NewPool(ctx, databaseURL, simpleProtocol bool)`.

Tham số `simpleProtocol` tắt prepared statement. Bật trong hai trường hợp:

1. `DATABASE_URL` trỏ tới endpoint `-pooler` của Neon (PgBouncer transaction pooling tái dùng
   connection giữa các transaction → prepared statement sinh lỗi "prepared statement already exists").
2. Cần `Exec` một file SQL chứa **nhiều câu lệnh** — extended protocol chỉ chạy được một câu lệnh
   mỗi lần gọi. Đây là lý do `cmd/seed` luôn truyền `true`.

Pool: `MaxConns=10`, `MinConns=1`, `MaxConnIdleTime=2m`, `MaxConnLifetime=30m`. Idle ngắn vì Neon
tự ngủ khi không dùng; giữ connection quá lâu chỉ tổ nhận lỗi khi compute tỉnh lại.

## 4. Quy ước làm việc với Neon

- Cấu hình **duy nhất qua `DATABASE_URL`** → code chạy được với cả Neon lẫn Postgres local, không
  hardcode Neon ở đâu cả.
- **Migration luôn đi endpoint trực tiếp.** `cmd/migrate` gọi `cfg.DirectURL()`, tự bỏ `-pooler`
  khỏi hostname. Lý do: golang-migrate cần advisory lock, PgBouncer ở chế độ transaction pooling
  không giữ được lock giữa các câu lệnh.
- **`DATABASE_URL` trỏ vào endpoint nào cũng chạy được.** Code đã xử lý cả hai: migration tự chuyển
  sang endpoint trực tiếp, còn runtime thì `IsPooledEndpoint()` bật `simpleProtocol` để tránh lỗi
  prepared statement. Endpoint trực tiếp vẫn là **khuyến nghị** vì ít lớp trung gian hơn, nhưng
  không phải điều kiện bắt buộc.
- `sslmode=require` bắt buộc. `.env` nằm trong `.gitignore`; repo chỉ commit `.env.example`.
- Branch `main` để phát triển, branch `test` cho integration test (`TEST_DATABASE_URL`).
- **Hiệu năng (mục 2.4.3)**: free tier autosuspend sau ~5 phút, query đầu tiên cold start
  khoảng 0.5–1s. Phải nêu rõ điều này khi đo và bình luận số liệu trong báo cáo.
- Dự phòng hôm bảo vệ: chuẩn bị Postgres chạy local, đổi `DATABASE_URL` là chạy offline được ngay.

## 5. Đóng gói

Frontend build static → nhúng vào binary Go bằng `embed.FS` → **một process, một port, không CORS**.
Hệ quả cần nhớ khi làm auth: frontend và backend cùng origin, nên session cookie đủ dùng, không cần
token đi kèm mỗi request.

## 6. Xử lý lỗi và log

- Log bằng `log/slog`. Không log chuỗi kết nối, mật khẩu, hay nội dung CV.
- Lỗi nghiệp vụ định nghĩa ở `domain` (ví dụ `ErrProjectClosed`), service trả lên, transport map
  sang HTTP status. Không dùng `panic` cho lỗi nghiệp vụ — `chi/middleware.Recoverer` chỉ là lưới an toàn.
- Xem `api_design.md` để biết bảng ánh xạ lỗi → status code.
