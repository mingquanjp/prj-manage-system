# Vai trò và phân quyền

Neo tài liệu: mục 2.2 (tác nhân), 2.4.1 (bảo mật).

## 1. Ba vai trò

`users.system_role` là enum `user_system_role` với đúng ba giá trị, khớp ba tác nhân Chương 2:

| Giá trị | Tác nhân trong tài liệu |
|---|---|
| `ADMIN` | Quản trị viên |
| `PM` | Quản lý dự án |
| `EMPLOYEE` | Nhân viên |

Tài liệu 4.2.3 hiện ghi `ADMIN/HR/EMPLOYEE`. `HR` **không xuất hiện ở bất kỳ use case nào** nên đã
bị bỏ; báo cáo cần sửa lại. Đừng "khôi phục" `HR` khi thấy chênh lệch.

Không có cột `positions.is_project_manager`. `system_role` đã đủ trả lời câu hỏi "ai được tạo dự án".

## 2. Phân quyền hai tầng

### Tầng 1 — RBAC theo `system_role`

| Nhóm chức năng | ADMIN | PM | EMPLOYEE |
|---|:---:|:---:|:---:|
| Quản lý tài khoản, phân quyền, khoá/mở | ✅ | — | — |
| Danh mục nền (phòng ban, chức danh, nhóm kỹ năng, kỹ năng) | ✅ | đọc | đọc |
| Hồ sơ nhân sự (thêm/sửa/vô hiệu hoá) | ✅ | đọc | đọc hồ sơ của chính mình |
| **Duyệt kỹ năng và kết quả trích xuất CV** | ✅ | — | — |
| Tra cứu hồ sơ nhân sự phục vụ phân công | ✅ | ✅ | — |
| Khởi tạo dự án | — | ✅ | — |
| Phân rã công việc, yêu cầu kỹ năng, phân công, đóng dự án | — | ✅ (xem tầng 2) | — |
| Gửi yêu cầu cập nhật hồ sơ / CV / kỹ năng | — | — | ✅ |
| Xem và cập nhật công việc được giao | — | — | ✅ |

Hai ranh giới tài liệu nêu tường minh (mục 2.4.1):

- Nhân viên gửi được yêu cầu cập nhật nhưng **không tự sửa hồ sơ năng lực chính thức**.
- PM tra cứu được hồ sơ nhân sự nhưng **không thay vai trò duyệt dữ liệu của ADMIN**.

### Tầng 2 — ABAC theo quyền sở hữu dự án

`system_role = PM` mới chỉ cho phép **tạo** dự án. Mọi thao tác trên một dự án cụ thể còn phải thoả:

```
project.pm_employee_id == current_employee_id
```

Neo: tiền điều kiện "dự án thuộc phạm vi quản lý của Quản lý dự án" ở UC001, UC002, UC003.
Không có kiểm tra này thì PM A sửa được dự án của PM B.

Áp dụng cho: phân rã công việc, sửa task, khai báo yêu cầu kỹ năng, chạy đề xuất, xác nhận phân
công, đóng dự án.

### Tầng 2 áp cho EMPLOYEE

Nhân viên chỉ thao tác được trên task mà `tasks.current_assignee_id == current_employee_id`
(UC004 dữ liệu đầu vào #2). Xem hồ sơ cá nhân cũng chỉ giới hạn ở `employee_id` của chính mình.

## 3. Middleware cần có (M1)

Thứ tự trong chuỗi xử lý:

1. `RequireAuth` — chưa đăng nhập hoặc phiên hết hạn → **401**.
2. `RequireRole(roles...)` — sai vai trò → **403**.
3. Kiểm tra ABAC — nằm ở **tầng service**, không nằm ở middleware, vì cần đọc dự án/task từ DB.
   Sai quyền sở hữu → **403**; không tồn tại → **404**.

Nguyên tắc: **deny by default**. Endpoint mới mà quên khai báo vai trò thì phải bị chặn, không
phải được mở.

Kiểm tra trạng thái tài khoản mỗi request: `users.status != ACTIVE` → 401 và huỷ phiên. Đây là lý
do nghiêng về session cookie hơn JWT — khoá tài khoản có hiệu lực tức thì, JWT phải thêm blacklist
mới làm được.

## 4. Xác thực — đã chốt: JWT access + refresh

Mật khẩu băm bằng bcrypt (`golang.org/x/crypto/bcrypt`, DefaultCost) — đã dùng trong `cmd/seed`.

### Quyết định

| Thành phần | Quyết định |
|---|---|
| Access token | JWT **HS256**, TTL **15 phút** |
| Claims | `sub` (user_id), `employee_id`, `role`, `iat`, `exp`, `jti` |
| Refresh token | Chuỗi ngẫu nhiên 32 byte, **lưu bản băm SHA-256** trong bảng `refresh_tokens`, TTL **7 ngày** |
| Xoay vòng | Mỗi lần refresh sinh token mới và thu hồi token cũ |
| Phát hiện tái sử dụng | Dùng lại token đã thu hồi → **thu hồi toàn bộ họ token** của phiên đó |
| Nơi lưu | Cả hai là cookie **HttpOnly + Secure + SameSite=Lax**; refresh cookie giới hạn `Path=/api/auth/refresh` |
| CSRF | SameSite=Lax cộng cùng origin là đủ; mọi thao tác ghi đều là POST/PUT |
| Khoá tài khoản | Xoá toàn bộ refresh token của user → chậm nhất 15 phút là bị đẩy ra |

### Vì sao JWT chứ không phải session opaque

Độ trễ tới Neon đo được **~78ms**. Session opaque buộc mỗi request phải query DB để tra phiên,
tức cộng thẳng một roundtrip vào mọi API call. JWT chỉ verify chữ ký trong bộ nhớ.

Lợi thế thường thấy của JWT (nhiều service, nhiều domain, mobile) **không áp dụng** ở đây vì
frontend nhúng cùng origin — lý do chọn là độ trễ, không phải tính stateless.

### Đánh đổi phải nêu trong báo cáo

Khoá tài khoản có **cửa sổ trễ tối đa 15 phút**, trong khi session opaque thì tức thì. Đây là đánh
đổi có ý thức giữa độ trễ và tính tức thời của thu hồi quyền.

Giảm nhẹ: các thao tác nhạy cảm (duyệt kỹ năng, xác nhận phân công, đóng dự án) đằng nào cũng đọc
`employees` từ DB, nên **chèn kiểm tra `users.status` ngay tại đó** để chặn tức thì. Không cần làm
với các endpoint chỉ đọc.

### Lưu ý implement

- **Không** lưu token trong `localStorage`. HttpOnly cookie thì JavaScript không đọc được → XSS
  không ăn cắp được token. Vì cùng origin nên không mất gì.
- `employee_id` và `role` nằm sẵn trong claims → tầng service không phải query lại `users`/`employees`
  cho mỗi lần kiểm tra ABAC.
- Khoá ký lấy từ biến môi trường `JWT_SECRET`, thêm vào `.env.example`. Không hardcode, không commit.
- Bảng `refresh_tokens` cần **migration `000002`**: `token_id`, `user_id`, `token_hash`, `expires_at`,
  `revoked_at`, `replaced_by_token_id`, `user_agent`, `created_at`.
- Thêm dependency `github.com/golang-jwt/jwt/v5`.
- Endpoint: `POST /api/auth/login`, `POST /api/auth/refresh`, `POST /api/auth/logout`, `GET /api/auth/me`.

## 5. Tài khoản mẫu

Sau `make seed`, mật khẩu chung `Password@123`:

| Username | Vai trò | Mã nhân viên |
|---|---|---|
| `admin` | ADMIN | EMP001 |
| `pm.hoa`, `pm.nam` | PM | EMP002, EMP003 |
| `dev.an`, `qa.phuc`, `ops.son`, … | EMPLOYEE | EMP004–EMP020 |

Dùng đúng ba tài khoản `admin` / `pm.hoa` / `dev.an` khi đi thử luồng và chụp màn hình cho báo cáo.
