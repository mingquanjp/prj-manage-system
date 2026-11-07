# Bảo mật — thiết kế đích V2

**Đồng bộ quyết định 15/10/2026:** [sổ quyết định, phần hiện hành và Bàn giao](../../docs/design/database-review-decisions-2026-10-14.md) ưu tiên cho nội dung đã thay thế; [dictionary/DBML hợp nhất](../../docs/design/database-proposal-2026-10-14.md) là thiết kế 33 bảng (cấp bậc theo chốt 26/10/2026). Đây là thiết kế DB, chưa triển khai DDL/migration hoặc backend nghiệp vụ; tiến độ frontend mô phỏng xem [sổ công việc](../../web/README.md#sổ-công-việc).


> Quyền nghiệp vụ theo [baseline 2026-10-12](../../docs/business-baseline-2026-10-12.md) và [phân quyền](roles_permissions.md). Admin quản lý tài khoản; HR quản lý hồ sơ hành chính/CV, duyệt thông tin cá nhân; Giám đốc duyệt kỹ năng. Không vai trò nào mặc nhiên là superset; HR không được sửa official skill hoặc đọc toàn dự án. Thiết kế role/claims cũ chưa đồng bộ, không thay kiểm tra quyền hiện hành.

Đọc trước auth, upload và mọi endpoint có ID. Quyền/session chi tiết ở roles_permissions.md.
Đây là yêu cầu triển khai, không khẳng định code skeleton đã đáp ứng.

Theo [chốt tài khoản 28/10](../../docs/business-baseline-2026-10-20.md#quyết-định-tiếp-nối-28102026--tự-cấp-mã-và-nhãn-quyền-tài-khoản), e010 chỉ là tên đăng nhập demo tạm. [Chốt 05/11](../../docs/business-baseline-2026-10-20.md#quyết-định-tiếp-nối-05112026--định-danh-đăng-nhập-backend) chọn email công ty để đăng nhập backend phiên bản đầu; username vẫn là dữ liệu riêng. Chuẩn hóa email đã chốt trim đầu/cuối và so không phân biệt hoa/thường (trùng thì dừng, không sửa dữ liệu tự động); quy tắc cấp username còn cần đặc tả; kênh cấp tài khoản mới bằng mật khẩu tạm qua email theo [chốt 06/11](../../docs/business-baseline-2026-10-20.md#quyết-định-tiếp-nối-06112026--cấp-mật-khẩu-tạm-qua-email). Không suy chính sách từ fixture.

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

Mutation cần Origin chính xác và token CSRF gắn session trong header. GET /api/auth/csrf bootstrap cho phiên hợp lệ; login kiểm tra nguồn, logout cần CSRF. GET không ghi nghiệp vụ ngoài bootstrap cookie. Hết hạn phiên phải đăng nhập lại, không đường refresh. A07 đã duyệt HMAC-SHA256 gắn secret phiên, cookie staffing_session, FE dev proxy /api; login sinh secret mới và thu hồi phiên cookie cũ hợp lệ của cùng user trong transaction, không thu hồi xuyên user. Giao thức vẫn phải qua contract test trước triển khai.

Theo chốt auth A02–A09 ngày 06/11, phiên có hạn tuyệt đối 8 giờ từ login, gồm phiên hạn chế; không tự gia hạn/idle timeout/ghi nhớ. Sáu cột session đủ cho hạn tuyệt đối, không thêm last_seen_at. Chưa xóa vật lý session hết hạn/revoked trong đợt này; chúng vẫn bị từ chối. Retention/lịch dọn chốt trước vận hành lâu dài. Theo [chốt 06/11](../../docs/business-baseline-2026-10-20.md#quyết-định-tiếp-nối-06112026--cấp-mật-khẩu-tạm-qua-email), tài khoản dùng mật khẩu tạm phải đổi trước khi truy cập nghiệp vụ. Chốt cuối cùng: users.must_change_password BOOLEAN NOT NULL DEFAULT false; không có hạn mật khẩu tạm. Server phải kiểm tra cờ hiện hành, không chỉ FE hoặc role trong phiên. A08/A09 đã duyệt: phiên hạn chế chỉ cho me/csrf/logout/đổi lần đầu, role không vượt cổng này; [proposal auth mục 6.2.5](../../docs/proposals/backend-auth-session.md#625-đổi-mật-khẩu-lần-đầu-và-bootstrap-phiên-hạn-chế--đã-duyệt-a08a09) chi tiết contract; sáu cột session không tự thực thi việc chặn nghiệp vụ.


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

Bcrypt DefaultCost; DB chỉ lưu bản băm. Có đổi mật khẩu chính chủ (xác minh mật khẩu cũ); ADMIN reset theo chính sách sẽ đặc tả.
Theo [chốt 06/11](../../docs/business-baseline-2026-10-20.md#quyết-định-tiếp-nối-06112026--cấp-mật-khẩu-tạm-qua-email), hệ thống sinh mật khẩu tạm ngẫu nhiên khi cấp tài khoản mới, gửi qua email đã đăng ký; không log hoặc hiển thị mật khẩu trong giao diện quản trị. Đăng nhập bằng mật khẩu tạm chỉ được đổi mật khẩu và thao tác phiên cần thiết, không truy cập API nghiệp vụ. Mật khẩu tạm chưa có thời hạn trong phiên bản hiện tại; thời hạn phiên vẫn riêng. Đổi thành công vô hiệu mật khẩu tạm, sau đó đăng nhập bằng mật khẩu mới.
Cờ buộc đổi đã chốt; tạo tài khoản mới với mật khẩu tạm phải ghi true (không dựa default false), đổi thành công thay hash và đặt false atomically. A08 đã duyệt thay hash + false + revoke mọi phiên của user cùng transaction, sau đó xóa cookie và yêu cầu login mới; không tự nâng phiên cũ. A09 đã duyệt mật khẩu mới tối thiểu 15 code point Unicode, tối đa 72 byte UTF-8, khác mật khẩu hiện tại; không cắt/trim/tự chuẩn hóa hoặc ép nhóm ký tự, cho phép khoảng trắng/chữ có dấu/paste. Áp cho đặt mật khẩu mới lần đầu, không áp ngược lên login credential cũ. Gửi lỗi/gửi lại vẫn chưa chốt. Không tự mở rộng sang quên mật khẩu hay mặc định Admin reset cùng quy trình.
Không giả định seed password là quy trình cấp tài khoản hoàn chỉnh.

Body JSON login/đổi mật khẩu lần đầu tối đa 8 KiB (8192 byte), đã duyệt riêng 06/11; lớn hơn trả 413/PAYLOAD_TOO_LARGE trước bcrypt/ghi dữ liệu. Kiểm byte thực đọc, không chỉ Content-Length; không áp sang upload hoặc thay min 15 code point/max 72 byte của mật khẩu. HTTP tests cần kiểm 8192/8193 byte và request không khai báo độ dài.

Rate limit in-memory cho một process: cấu hình thử nghiệm 30 lần/5 phút/IP, 5 lần thất bại/5 phút/định danh đăng nhập chuẩn hóa (email công ty theo chốt 05/11, không username demo).
429 + Retry-After, không tự khóa tài khoản vĩnh viễn. Chỉ tin forwarded IP từ proxy đã cấu hình.
Đổi lần đầu có limiter riêng đã duyệt: 5 lần xác minh mật khẩu hiện tại sai/15 phút/user và 30 request/5 phút/IP trước bcrypt; 429/Retry-After, không khóa vĩnh viễn, bucket có hạn/dọn bộ nhớ. Validation/CSRF/DB lỗi không tính là sai mật khẩu. Các ngưỡng là cấu hình thử nghiệm, chưa là kết quả đo.
Login báo lỗi chung, không tiết lộ tài khoản tồn tại.

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
