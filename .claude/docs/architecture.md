# Kiến trúc mã nguồn — V1

Thiết kế đích bám baseline; skeleton hiện tại chưa có service nghiệp vụ.

## 1. Chiều phụ thuộc

```text
HTTP transport → service → domain
                    ↓
                 matching → domain
PostgreSQL adapter → interface do service khai báo
Composition root → transport + service + adapters
```

Sơ đồ là chiều import, không phải tất cả luồng gọi runtime.
Domain và matching thuần dữ liệu, không DB/HTTP; matching chỉ trả đề xuất.
Service điều phối use case, kiểm tra quyền/nghiệp vụ và transaction runner qua interface.
Postgres adapter thực thi repository/transaction interface, không trả pgx type ra service.
Transport parse/validate định dạng, gọi service, map DTO/lỗi theo quyền.
Không dùng “service quản lý transaction” để buộc service import postgres adapter.

## 2. Hiện trạng và thành phần đích

Hiện có cmd/server, cmd/migrate, cmd/seed; internal/config, repository/postgres/db.go, transport/http/router.go;
migration 000001, seed; web chưa khởi tạo.
Đích: domain, service, matching, adapters, DTO/auth middleware, storage.Provider, CV worker và workcalendar.
Giữ Go + chi + pgx/PostgreSQL; SQL trực tiếp hiện tại, không tự thêm sqlc/query builder/parser dependency.

Frontend Vite/React/TypeScript/Tailwind/shadcn/TanStack Query/React Router đã chọn, chưa khởi tạo.
Build static nhúng embed.FS: một process, một port, cùng origin, không cần CORS.
Auth vẫn JWT/refresh cookie và session check DB; cùng origin không loại bỏ CSRF.

## 3. Transaction và snapshot

Unit of work cung cấp repository gắn cùng transaction.
Confirm batch, approve request, transition/log, publish CV, project cancel/close là các transaction nghiệp vụ.
Kiểm tra khóa/version phải phối hợp với thay role/status/skill/ownership; thứ tự khóa nhất quán, retry deadlock có giới hạn.
Idempotency result ghi cùng transaction; không giữ transaction trong lúc người dùng xem draft.
Recommendation đọc snapshot nhất quán (ví dụ repeatable-read), scorer chạy thuần trên snapshot, persist input/output.
Không giữ transaction DB mở trong quá trình parse file hoặc chờ người duyệt.

## 4. CV và lịch

CV local disk private qua storage.Provider; DB jobs do worker trong process nhận bằng lease.
Retry giới hạn, trạng thái rõ; file/DB không atomic nên có cleanup/recovery.
Work calendar nguồn cấu hình, snapshot immutable bền vững trong DB để phiên cũ còn tính được sau thay config.
Notification nội bộ ghi DB cùng transaction, không cần email/push broker.

## 5. DB và vận hành

DATABASE_URL là cấu hình kết nối, không hardcode Neon.
Migration dùng endpoint trực tiếp theo cmd/migrate; không chạy reset/drop DB thật để đồng bộ thiết kế.
Cấu hình pool/protocol hiện có xem mã NewPool; không coi nhận định về pooling/độ trễ cũ là bảo đảm cho mọi phiên bản dịch vụ.
Integration test dùng TEST_DATABASE_URL riêng, không tự trỏ DB phát triển thật.
Cold/warm phải đo riêng; không dùng con số autosuspend/cold start cố định làm cam kết.
Local Postgres là phương án dự phòng cần diễn tập cùng migration/config, không mặc định đã sẵn sàng.

Log slog có redaction; lỗi nghiệp vụ có mã, transport map HTTP, không panic.
Health công khai trả trạng thái chung; chi tiết DB không trả err.Error() cho client.
