# Bảo mật — thiết kế đích V1

Đọc trước auth, upload và mọi endpoint có ID. Quyền/session chi tiết ở roles_permissions.md.
Đây là yêu cầu triển khai, không khẳng định code skeleton đã đáp ứng.

## 1. Bí mật và dữ liệu trả về

Không commit/log mật khẩu, hash, token, JWT_SECRET, DATABASE_URL, CV thô hay liên hệ cá nhân.
.env chỉ local; .env.example không bí mật. Không suy “chưa lộ bí mật” chỉ từ kiểm tra worktree.
DTO theo vai trò; không trả trực tiếp entity DB. Lỗi API/health tổng quát, lỗi nội bộ đã che bí mật chỉ vào log.
SQL tham số hóa; cột/direction sort allowlist; không nối giá trị client vào SQL.
React escape mặc định không thay thế toàn bộ phòng XSS; tránh HTML tùy ý, URL nguy hiểm và script bên ngoài.

## 2. Auth và CSRF

JWT 15 phút + refresh rotation + session DB 7 ngày; mọi API nghiệp vụ kiểm tra session/status/role hiện hành.
Khóa/đổi role/reset mật khẩu thu hồi tất cả session, logout phiên hiện tại.
Cookie host-only, HttpOnly, Secure, SameSite=Lax; access /api, refresh /api/auth.
Chỉ cấu hình dev localhost mới cho phép tắt Secure, production HTTPS.

Mutation cần Origin chính xác và signed double-submit token gắn session trong header.
Login kiểm tra nguồn; GET không ghi nghiệp vụ (ngoại trừ phát cookie CSRF phục vụ bootstrap).
GET /api/auth/csrf cấp token cho phiên hợp lệ, dùng session/refresh khi access hết hạn; không đổi trạng thái nghiệp vụ.
Refresh/logout vẫn cần CSRF; xóa cookie đúng Path. Swagger phải gửi header, không tuyên bố cookie tự đủ an toàn.
Chi tiết giao thức/bootstrap phải được contract test trước M1, không tự bỏ CSRF khi token access hết hạn.

Nguồn nguyên tắc: [OWASP CSRF](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html),
[MDN Set-Cookie](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Set-Cookie).

## 3. Quyền và cạnh tranh

RBAC + ABAC trên project/task/CV/run/notification và dữ liệu lịch sử; membership không cấp thêm quyền.
Actor lấy từ phiên; ID người được chọn trong request là resource subject, validate riêng.
403/404 khác nhau có thể tiết lộ sự tồn tại; không trả nội dung trái quyền.
Duyệt/phân công/chuyển trạng thái kiểm tra lại quyền và version trong transaction.
Phối hợp khóa theo thứ tự thống nhất trên mọi đường đổi trạng thái/vai trò/năng lực; không chỉ khóa task.

## 4. Mật khẩu và giới hạn

Bcrypt DefaultCost; không lưu rõ. Có đổi mật khẩu (xác minh mật khẩu cũ), ADMIN reset và buộc đổi trước API nghiệp vụ khác.
Chính sách độ dài/min-max, mật khẩu tạm hết hạn và kênh giao mật khẩu tạm còn phải đặc tả trước M1.
Không giả định seed password là quy trình cấp tài khoản hoàn chỉnh.

Rate limit in-memory cho một process: cấu hình thử nghiệm 30 lần/5 phút/IP, 5 lần thất bại/5 phút/username chuẩn hóa.
429 + Retry-After, không tự khóa tài khoản vĩnh viễn. Chỉ tin forwarded IP từ proxy đã cấu hình.
Login báo lỗi chung, không tiết lộ username tồn tại.

## 5. Upload và vận hành

PDF text/DOCX ≤10 MiB, kiểm tra nội dung thật và container DOCX; giới hạn giải nén, timeout/memory parser.
Tên server sinh, file private ngoài static web root, download phân quyền + attachment + nosniff.
Không cho parser truy cập mạng tùy ý; không thực thi macro/nội dung CV.
Xử lý file/DB thất bại có cleanup và retry idempotent; không log text CV.
Local disk cần volume bền vững và backup đồng bộ DB/file; mã hóa ổ đĩa/quyền đọc cấu hình lúc triển khai.

Giới hạn body/list; pagination default 20/max 100. Header CSP, nosniff, frame DENY, Referrer-Policy same-origin.
CSP cụ thể chốt cùng frontend/Swagger, không mở CDN mặc định.
Audit nghiệp vụ append-only; migration xem lại cascade để không mất lịch sử.
govulncheck trước đóng gói; dependency mới cần duyệt. Không khai triển yêu cầu này thành kết quả scan chưa chạy.
