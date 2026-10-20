# Kiến trúc mã nguồn — V2

**Đồng bộ quyết định 15/10/2026:** [sổ quyết định, phần hiện hành và Bàn giao](../../docs/design/database-review-decisions-2026-10-14.md) ưu tiên cho nội dung đã thay thế; [dictionary/DBML hợp nhất](../../docs/design/database-proposal-2026-10-14.md) là thiết kế 32 bảng. Chưa triển khai code/DDL/migration.


Thiết kế đích đối chiếu [baseline hiện hành](../../docs/business-baseline-2026-10-12.md); skeleton chưa có service nghiệp vụ. Service kiểm tra trách nhiệm PTCM hiện hành cho phân công/nghiệm thu, loại PTCM khỏi người thực hiện cùng dự án và kiểm tra điều kiện đối ứng khi thay PTCM. Quyền PM điều hành dự án không thay quyền chuyên môn.

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
R06 read models EmployeeAdminView/AssignmentCandidateView/RecommendationView đọc cùng nguồn; repository projection + service ABAC + DTO allowlist, không chỉ giấu ở frontend.
Matching nhận snapshot đủ dữ liệu, trả explanation có cấu trúc để service chọn phần được phép xem; không phát raw input qua API.
Không dùng “service quản lý transaction” để buộc service import postgres adapter.

## 2. Hiện trạng và thành phần đích

Hiện có cmd/server, cmd/migrate, cmd/seed; internal/config, repository/postgres/db.go, transport/http/router.go;
migration 000001, seed. Web đã có pilot local được duyệt visual 20/10/2026; xem [hiện trạng và harness frontend](../../web/README.md). Không đồng nghĩa các service nghiệp vụ đã có.
Đích: domain, service, matching, adapters, DTO/auth middleware, storage.Provider, CV worker và dịch vụ work effort/audit.
Giữ Go + chi + pgx/PostgreSQL; SQL trực tiếp hiện tại, không tự thêm sqlc/query builder/parser dependency.

Frontend đã có Vite/React/TypeScript/Tailwind/React Router và primitives tùy biến shadcn/Radix. Pilot chỉ mock trong bộ nhớ; TanStack Query/API/auth chưa triển khai.
Đích build static nhúng embed.FS: một process, một port, cùng origin, không cần CORS; pilot local chưa tích hợp Go.
Auth dùng cookie secret và server-side session check DB; cùng origin không loại bỏ CSRF.

## 3. Transaction và snapshot

Unit of work cung cấp repository gắn cùng transaction.
Confirm batch, approve request, transition/log, xác nhận CV hiện tại, project cancel/close là các transaction nghiệp vụ.
Kiểm tra khóa/dữ liệu nền và request.version kỹ năng phải phối hợp với thay role/status/skill/ownership; thứ tự khóa nhất quán, retry deadlock có giới hạn.
Idempotency result ghi cùng transaction; không giữ transaction trong lúc người dùng xem draft.
Recommendation đọc snapshot nhất quán (ví dụ repeatable-read), scorer chạy thuần trên snapshot, persist input/output.
Không giữ transaction DB mở trong quá trình parse file hoặc chờ người duyệt.

## 4. CV và effort

CV cloud cấu hình chung qua storage.Provider; API nhận diện tên kỹ năng/đoạn căn cứ. Provider/model/cách chạy chưa chọn; tạm giữ DB jobs/lease để không bỏ retry trước khi có phương án thay.
Retry giới hạn, trạng thái rõ; file/DB không atomic nên có cleanup/recovery.
Work effort dùng ngày địa phương Asia/Ho_Chi_Minh và phút nguyên; không workcalendar/timer.
Service chung kiểm tra tổng theo employee/date xuyên dự án và quyền sửa; log, history UPDATE/DELETE hoặc result_log_id cho ADD, quyết định request ghi atomically; tổng suy ra không cache riêng.
Serialize mutation theo employee/date (cả ngày cũ/mới khi sửa), kết hợp khóa project/task/request và giá trị nền theo thứ tự thống nhất; close/cancel/approve không vượt nhau.
Submission snapshot bộ nhật ký, recommendation snapshot dữ liệu đầu vào là hai loại riêng; không ghi đè snapshot lịch sử.
R01 dùng service hiệu chỉnh chung, không backdate state history; R02 tách đầu ra khỏi độ đầy đủ nhật ký; không nộp kết quả thay.
R03 giữ quyết định mục CV đã kết thúc; bổ sung bằng yêu cầu mới, không attempts/close-review riêng. R04 bổ sung lại kỹ năng qua quy trình yêu cầu thông thường.
R05 không history revision/acknowledgement; service confirm kiểm tra điều kiện phân công hiện hành rồi ghi assignment, không có history acknowledgement, scorer không query DB; output run bất biến.
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
