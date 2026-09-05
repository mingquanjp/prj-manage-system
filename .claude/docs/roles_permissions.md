# Vai trò và phân quyền — V1

Thiết kế đích đã đồng ý 2026-09-05; chưa implement auth.
Nguồn quyết định: [baseline V1](../../docs/reviews/system-design-review-v1.md).

## 1. RBAC + ABAC

Một user một role ADMIN/PM/EMPLOYEE; không HR, không suy quyền từ chức danh.

| Chức năng | ADMIN | PM | EMPLOYEE |
|---|---|---|---|
| Tài khoản, danh mục, hồ sơ chính thức, duyệt năng lực/CV | Quản trị | Đọc dữ liệu được phép | Hồ sơ của mình |
| Tra cứu phục vụ phân công | Có | Năng lực chính thức toàn tổ chức | Không |
| Tạo/quản lý dự án, gợi ý, phân công, nghiệm thu | Không | Chỉ project.pm_employee_id = mình | Không |
| Chuyển PM khi bàn giao | Có, lý do/audit | Không | Không |
| Gửi yêu cầu hồ sơ/kỹ năng/CV | Nhập thay có audit | Không trong V1 | Của mình |
| Nhận và thực hiện task | Không | Không | task.current_assignee_id = mình |

ADMIN không mặc định đọc task/bình luận tất cả dự án; màn hình chuyển PM chỉ trả metadata cần thiết.
PM chỉ xem CV ACTIVE và thông tin nghề nghiệp được phép, không xem CV pending hoặc dữ liệu xác thực.
Project membership là lịch sử tham gia, không cấp quyền đọc mọi task.
Người cũ sau reassign chỉ xem đóng góp lịch sử của mình qua DTO riêng, không task detail hiện tại.
Notification/download/run/comment đều phải kiểm tra quyền, không dùng ID như bằng chứng được phép.

## 2. Kiểm tra quyền hiện hành

Verify JWT → xác minh session còn hạn/chưa revoked + user/employee ACTIVE + role hiện hành → RBAC → service ABAC.
Mọi API nghiệp vụ phải kiểm tra hiện trạng DB, có thể gộp truy vấn; chấp nhận roundtrip để thu hồi quyền ngay.
Không dùng role/employee_id claims làm nguồn hiện hành duy nhất.
Actor lấy từ auth; employee_id được chọn trong body phân công là đối tượng, phải validate riêng.

Mutation kiểm tra lại quyền/version trong transaction với quy ước khóa chung.
Yêu cầu đang chạy cạnh tranh với khóa/chuyển PM phải có thứ tự commit xác định; không hứa hủy ngược giao dịch đã commit.
Sai phiên: 401; sai quyền: 403; không tồn tại: 404. Chính sách này không che tuyệt đối sự tồn tại tài nguyên.
Deny by default, ngoại lệ public khai báo tường minh.

## 3. Tài khoản và bàn giao

Tạo user + employee atomically, không API tạo tài khoản mồ côi.
Chỉ EMPLOYEE active nhận task; đổi role bị chặn nếu còn việc mở/dự án mở theo vai trò cũ.
Không vô hiệu hóa/hạ quyền ADMIN active cuối cùng.
Khóa khẩn cấp được phép dù còn việc; thu hồi mọi session và liệt kê việc cần PM bàn giao/hủy.
ADMIN chuyển ownership sang PM active, bắt buộc lý do, history/audit; không trao quyền phân công cho ADMIN.
Đổi role/khóa/reset mật khẩu thu hồi tất cả session; logout thu hồi session hiện tại.

## 4. Xác thực

| Thành phần | Thiết kế |
|---|---|
| Access | JWT HS256, 15 phút; sub, employee_id, role, sid, iat, exp, jti |
| Session | auth_sessions: user_id, created_at, expires_at, revoked_at; hạn tuyệt đối 7 ngày từ login |
| Refresh | Ngẫu nhiên 32 byte, chỉ lưu SHA-256 hash; liên kết session, expiry, revoked_at, replacement |
| Rotation | Thu hồi cũ/phát mới atomically; reuse thu hồi cả session |
| Cookie | HttpOnly, Secure, SameSite=Lax, host-only; access Path=/api, refresh Path=/api/auth |
| CSRF | Signed double-submit gắn session + header, kiểm tra Origin chính xác; login cũng kiểm tra nguồn |
| Logout | Thu hồi phía server; xóa cookie đúng tên/path; JWT cũ bị từ chối do session revoked |

Refresh không kéo dài session quá 7 ngày. Frontend phối hợp refresh giữa các request/tab, không vô tình tái dùng token.
Không localStorage/sessionStorage cho auth tokens. HttpOnly hạn chế đọc token, không ngăn XSS thực hiện hành động dưới phiên nạn nhân.
JWT_SECRET không có mặc định, thiếu thì từ chối khởi động.
Chọn JWT được giữ theo quyết định trước; không còn lập luận JWT loại bỏ mọi truy vấn phiên.

Bcrypt đã dùng ở seed. Đổi/reset mật khẩu và chính sách cụ thể xem security.md.
Tài khoản seed chỉ dev, không dùng mật khẩu mẫu cho triển khai thực.
