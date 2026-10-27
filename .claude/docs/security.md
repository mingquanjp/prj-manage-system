# Bảo mật — thiết kế đích V2

**Đồng bộ quyết định 15/10/2026:** [sổ quyết định, phần hiện hành và Bàn giao](../../docs/design/database-review-decisions-2026-10-14.md) ưu tiên cho nội dung đã thay thế; [dictionary/DBML hợp nhất](../../docs/design/database-proposal-2026-10-14.md) là thiết kế 33 bảng (cấp bậc theo chốt 26/10/2026). Đây là thiết kế DB, chưa triển khai DDL/migration hoặc backend nghiệp vụ; tiến độ frontend mô phỏng xem [sổ công việc](../../web/README.md#sổ-công-việc).


> Quyền nghiệp vụ theo [baseline 2026-10-12](../../docs/business-baseline-2026-10-12.md) và [phân quyền](roles_permissions.md). Admin quản lý tài khoản; HR quản lý hồ sơ hành chính/CV, duyệt thông tin cá nhân; Giám đốc duyệt kỹ năng. Không vai trò nào mặc nhiên là superset; HR không được sửa official skill hoặc đọc toàn dự án. Thiết kế role/claims cũ chưa đồng bộ, không thay kiểm tra quyền hiện hành.

Đọc trước auth, upload và mọi endpoint có ID. Quyền/session chi tiết ở roles_permissions.md.
Đây là yêu cầu triển khai, không khẳng định code skeleton đã đáp ứng.

## 1. Bí mật và dữ liệu trả về

Không commit/log mật khẩu, hash, token, DATABASE_URL, CV thô hay liên hệ cá nhân.
.env chỉ local; .env.example không bí mật. Không suy “chưa lộ bí mật” chỉ từ kiểm tra worktree.
DTO theo vai trò và mục đích R06; không trả entity DB/raw snapshot. ADMIN không là superset mọi quyền PM.
Địa chỉ cá nhân (address) theo chốt 20/10 chỉ thuộc hồ sơ hành chính/chính chủ và HR theo mục đích; không trong tra cứu PM/PTCM, candidate DTO hoặc snapshot gợi ý. Quyền Giám đốc đọc căn cứ kỹ năng không tự cấp quyền đọc địa chỉ. Không ghi địa chỉ vào log kỹ thuật; audit hành chính theo quyền, không trả toàn bộ audit cho tra cứu chuyên môn.
Cấm hash/token/secret trong DTO mọi vai trò; candidate view không chứa pending skill/profile, lịch sử khóa/reset hay nội dung task/log ngoài project.
Gợi ý chỉ hiển thị thời điểm tạo và thông tin tổng hợp được phép, không lý do hiệu chỉnh riêng tư của project khác. Lỗi API/health tổng quát, lỗi nội bộ đã che bí mật chỉ vào log.
SQL tham số hóa; cột/direction sort allowlist; không nối giá trị client vào SQL.
React escape mặc định không thay thế toàn bộ phòng XSS; tránh HTML tùy ý, URL nguy hiểm và script bên ngoài.

## 2. Auth và CSRF

Session phía server: cookie secret ngẫu nhiên, lưu hash SHA-256 duy nhất trong auth_sessions, không lưu secret rõ/ID tăng dần trong cookie. Mọi API kiểm tra phiên còn hạn/chưa revoked, status và quyền hiện hành. Không JWT/refresh endpoint.
Khóa/đổi role/reset thu hồi tất cả phiên, logout phiên hiện tại. Cookie host-only, HttpOnly, Secure, SameSite=Lax, Path=/api; chỉ dev localhost mới được tắt Secure, production HTTPS.

Mutation cần Origin chính xác và token CSRF gắn session trong header. GET /api/auth/csrf bootstrap cho phiên hợp lệ; login kiểm tra nguồn, logout cần CSRF. GET không ghi nghiệp vụ ngoài bootstrap cookie. Hết hạn phiên phải đăng nhập lại, không đường refresh. Giao thức CSRF/rotation secret khi login cần contract test trước M1.

Chưa chốt thời lượng tuyệt đối/idle/cleanup. Sáu cột auth_sessions hỗ trợ expires_at tuyệt đối; không tuyên bố có idle timeout khi chưa có phương án lưu hoạt động. Không mặc định 7 ngày. Bỏ cờ buộc đổi mật khẩu, không khôi phục bằng mô tả API.


Nguồn nguyên tắc: [OWASP CSRF](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html),
[MDN Set-Cookie](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Set-Cookie).

## 3. Quyền và cạnh tranh

RBAC + ABAC trên project/task/CV/run/notification và dữ liệu lịch sử; membership không cấp thêm quyền.
Actor lấy từ phiên; ID người được chọn trong request là resource subject, validate riêng.
403/404 khác nhau có thể tiết lộ sự tồn tại; không trả nội dung trái quyền.
Duyệt/phân công/chuyển trạng thái kiểm tra lại quyền và giá trị nền trong transaction.
Effort không cho client giả actor/employee/assignment; chỉ người thực hiện tự ghi/gửi, PTCM áp dụng yêu cầu đã duyệt.
R01 (21/10): ghi/sửa/xóa nhật ký trực tiếp kiểm tra lại đúng người hiện được giao, task/project IN_PROGRESS tại thời điểm lưu; ASSIGNED/PAUSED phải Bắt đầu/Tiếp tục trước. Assignment/ngày/tổng/khóa vẫn kiểm tra. Nhật ký đã khóa theo hiệu chỉnh chung, không backdate history hoặc duyệt ngoại lệ quên Start/Resume.
R02 cho nộp khi thiếu nhật ký, không có nộp kết quả thay; không ép khai giờ giả.
R03 không mở lại mục CV/attempt; R04 bổ sung lại bằng yêu cầu kỹ năng mới. Kiểm tra quyền/dữ liệu nền và request.version kỹ năng, không sửa/xóa quyết định đã kết thúc.
R05 không có history acknowledgement; snapshot bất biến, kiểm tra điều kiện phân công hiện hành trong transaction, không tin điểm client.
Lịch sử đóng góp/effort request trả DTO giới hạn, không mở lại quyền task hiện hành cho người cũ.
Link PR/tài liệu chỉ là minh chứng tham khảo: validate scheme http/https, không tự fetch URL/đăng nhập/kiểm tra merge; không đưa token hoặc URL chứa bí mật vào nhật ký.
Phối hợp khóa theo thứ tự thống nhất trên mọi đường đổi trạng thái/vai trò/năng lực; không chỉ khóa task.

## 4. Mật khẩu và giới hạn

Bcrypt DefaultCost; không lưu rõ. Có đổi mật khẩu (xác minh mật khẩu cũ), ADMIN reset theo chính sách sẽ đặc tả; chưa có luồng buộc đổi trước API nghiệp vụ khác.
Chính sách độ dài/min-max, mật khẩu tạm hết hạn và kênh giao mật khẩu tạm còn phải đặc tả trước M1.
Không giả định seed password là quy trình cấp tài khoản hoàn chỉnh.

Rate limit in-memory cho một process: cấu hình thử nghiệm 30 lần/5 phút/IP, 5 lần thất bại/5 phút/username chuẩn hóa.
429 + Retry-After, không tự khóa tài khoản vĩnh viễn. Chỉ tin forwarded IP từ proxy đã cấu hình.
Login báo lỗi chung, không tiết lộ username tồn tại.

## 5. Upload và vận hành

PDF text/DOCX ≤10 MiB, kiểm tra nội dung thật và container DOCX; giới hạn giải nén, timeout/memory parser.
Tệp trên một cloud cấu hình chung; server xác thực upload và giữ storage_key private, download phân quyền + attachment + nosniff. Không public URL chuẩn.
Quyền tra cứu của PM/PTCM không cho tải CV gốc của người khác, kể cả CV hiện tại. Download kiểm tra quyền chủ hồ sơ/HR/căn cứ xét duyệt của Giám đốc; không suy quyền từ biết ID hoặc bản hiện tại.
Chỉ cho API trích xuất kết nối provider đã cấu hình theo allowlist/timeout; không fetch URL từ nội dung CV, không thực thi macro hoặc chỉ dẫn trong CV. Provider/model chưa chọn.
Xử lý file/DB thất bại có cleanup và retry idempotent; không log text CV.
Cloud cần chính sách backup/quyền private/cleanup phối hợp DB; provider chưa chọn. Minh chứng ngoài CV lưu metadata JSON trên request: server kiểm tra storage_key thuộc upload được phép, tên/type/size hợp lệ, chỉ bổ sung PENDING; không binary/base64 hoặc ID task/CV thay FK. Loại/số/dung lượng minh chứng và cleanup upload chưa gắn còn cần đặc tả.

Giới hạn body/list; pagination default 20/max 100. Header CSP, nosniff, frame DENY, Referrer-Policy same-origin.
CSP cụ thể chốt cùng frontend/Swagger, không mở CDN mặc định.
Audit nghiệp vụ append-only; migration xem lại cascade để không mất lịch sử.
govulncheck trước đóng gói; dependency mới cần duyệt. Không khai triển yêu cầu này thành kết quả scan chưa chạy.
