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

## 4. Xác thực — TBD

**Chưa chốt.** Kế hoạch giả định session cookie HttpOnly + SameSite=Lax (frontend nhúng cùng origin
nên không cần token). Người dùng chưa xác nhận. **Không implement khi chưa hỏi.**

Đã chốt phần còn lại: mật khẩu băng bcrypt (`golang.org/x/crypto/bcrypt`, DefaultCost) — đã dùng
trong `cmd/seed`.

## 5. Tài khoản mẫu

Sau `make seed`, mật khẩu chung `Password@123`:

| Username | Vai trò | Mã nhân viên |
|---|---|---|
| `admin` | ADMIN | EMP001 |
| `pm.hoa`, `pm.nam` | PM | EMP002, EMP003 |
| `dev.an`, `qa.phuc`, `ops.son`, … | EMPLOYEE | EMP004–EMP020 |

Dùng đúng ba tài khoản `admin` / `pm.hoa` / `dev.an` khi đi thử luồng và chụp màn hình cho báo cáo.
