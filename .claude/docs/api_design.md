# Thiết kế API

## 0. Tài liệu API — OpenAPI 3.1 viết tay + Swagger UI nhúng

**Đã chốt.** Chưa implement, thuộc M1.

| Hạng mục | Quyết định |
|---|---|
| Nguồn chuẩn | `api/openapi.yaml`, **viết tay**, đặc tả OpenAPI **3.1** |
| Giao diện thử | Swagger UI phục vụ tại `GET /api/docs`, spec tại `GET /api/openapi.yaml` |
| Bật khi nào | **Chỉ khi `APP_ENV=dev`.** Ở môi trường khác thì route không được đăng ký, trả 404 |
| Assets | Nhúng vào binary bằng `embed.FS`. **Không kéo từ CDN** — CSP ở `security.md` mục 10 chặn nguồn ngoài |
| Chống lệch spec | Test dùng `chi.Walk` duyệt mọi route đã đăng ký, đối chiếu với đường dẫn trong spec. Thiếu hoặc thừa đều làm test đỏ |

### Vì sao viết tay chứ không dùng annotation

`swaggo/swag` v1 (v1.16.6) chỉ sinh được **OpenAPI 2.0**, mà 2.0 không có kiểu security scheme
`in: cookie`. Hệ thống này xác thực bằng cookie HttpOnly, nên spec sẽ mô tả sai cơ chế xác thực —
điều hội đồng có thể hỏi tới. Bản `swag/v2` hỗ trợ 3.1 nhưng vẫn là `v2.0.0-rc5`, không nên đưa
bản RC vào đồ án.

Viết tay đổi lại lấy: khai báo đúng `cookieAuth`, không thêm bước sinh mã, không thêm dependency
sinh mã (giảm bề mặt chuỗi cung ứng — `security.md` mục 12). Nhược điểm duy nhất là spec có thể
lệch khỏi code, và đó chính là thứ test `chi.Walk` canh.

### Khai báo xác thực trong spec

```yaml
components:
  securitySchemes:
    cookieAuth:
      type: apiKey
      in: cookie
      name: access_token
security:
  - cookieAuth: []      # mặc định cho mọi endpoint
```

Các endpoint công khai (`/api/auth/login`, `/api/healthz`) ghi đè bằng `security: []`.

Swagger UI **không cần cấu hình gì thêm để thử**: nó phục vụ cùng origin với API, nên trình duyệt
tự gắn cookie sau khi gọi `POST /api/auth/login` ngay trong giao diện. Đây cũng là lý do khác để
nhúng UI vào chính binary thay vì mở file HTML rời.

### Bố cục dự kiến

```
api/openapi.yaml                        nguồn chuẩn
internal/transport/http/docs.go         đăng ký route /api/docs, nhúng spec + UI
internal/transport/http/swaggerui/      dist của Swagger UI, nhúng qua embed.FS
internal/transport/http/docs_test.go    chi.Walk đối chiếu route với spec
```

Còn một chi tiết nhỏ để lại tới lúc làm: lấy bộ Swagger UI dist bằng cách nào. Nghiêng về **tự đưa
vào repo** bản dist chính thức thay vì phụ thuộc `swaggo/http-swagger` (bản mới nhất từ 2023), để
tự kiểm soát phiên bản UI khi có bản vá.

## 1. Hiện trạng

Endpoint duy nhất đã có: `GET /api/healthz` → `{"status":"ok","database":"ok","database_latency_ms":78}`.

Router: `chi/v5`, middleware đang bật theo thứ tự `RequestID → RealIP → Recoverer → Timeout`
(`internal/transport/http/router.go`).

## 2. Quy ước

- Tiền tố `/api`. Đường dẫn dùng **danh từ số nhiều, tiếng Anh, kebab-case**:
  `/api/task-groups`, `/api/skill-approvals`.
- Tài nguyên con lồng dưới cha khi quan hệ là sở hữu: `/api/projects/{id}/tasks`.
  Tài nguyên có định danh toàn cục thì phẳng: `/api/tasks/{id}`.
- Động từ chỉ dùng cho hành động không map được sang CRUD, đặt ở cuối:
  `/api/assignments/confirm`, `/api/skill-approvals/{id}/approve`.
- JSON, key **snake_case**, khớp tên cột CSDL để giảm tầng ánh xạ.
- Thời gian trả về theo **RFC3339 có offset** (`2026-07-27T14:30:00+07:00`), đúng với `timestamptz`.
- Điểm số trả về theo thang **0–100** (`matching_score: 86.00`), khớp `decimal(5,2)` trong DB.
  Nếu UI cần "0.86/1" như ví dụ trong tài liệu thì chia ở frontend.
- Frontend nhúng cùng origin → **không cấu hình CORS**.

## 3. Bề mặt API dự kiến

Nhóm theo use case. Đây là kế hoạch, chưa implement — điều chỉnh khi làm từng mốc.

```
# Xác thực (M1) — JWT access 15 phút + refresh xoay vòng, xem roles_permissions.md mục 4
POST   /api/auth/login       # đặt cả access và refresh cookie
POST   /api/auth/refresh     # xoay vòng refresh token, cấp access token mới
POST   /api/auth/logout      # thu hồi refresh token, xoá cookie
GET    /api/auth/me

# Tài khoản (M1, ADMIN)
GET    /api/users
POST   /api/users
PUT    /api/users/{id}
POST   /api/users/{id}/lock
POST   /api/users/{id}/unlock

# Danh mục nền (M2, ADMIN ghi / mọi vai trò đọc)
GET|POST /api/departments      /api/positions
GET|POST /api/skill-categories /api/skills

# Nhân sự và năng lực (M2)
GET    /api/employees?skill=&position=&department=&seniority=&status=   # bộ lọc UC003.2 bước 4
GET    /api/employees/{id}                       # hồ sơ + kỹ năng + lịch sử dự án (2.2.6)
POST   /api/employees/{id}/skills                # tạo bản ghi chờ duyệt
GET    /api/skill-approvals                      # ADMIN: danh sách chờ duyệt
POST   /api/skill-approvals/{id}/approve
POST   /api/skill-approvals/{id}/reject

# Dự án và công việc (M3)
GET|POST /api/projects                           # UC001
GET    /api/projects/{id}
POST   /api/projects/{id}/close
POST   /api/projects/{id}/task-groups            # UC002
GET|POST /api/projects/{id}/tasks
PUT    /api/tasks/{id}
PUT    /api/tasks/{id}/requirements

# Phân công (M4, M5)
GET    /api/projects/{id}/assignable-tasks       # UC003 bước 6-7
POST   /api/tasks/recommendations                # body: [task_id] → chạy engine (UC003.1)
POST   /api/assignments/confirm                  # body: [{task_id, employee_id, note,
                                                 #         recommendation_item_id, adjust_reason}]

# Công việc được giao (M6)
GET    /api/my/tasks
PUT    /api/tasks/{id}/status                    # UC004 — sinh work_log tự động
GET|POST /api/tasks/{id}/comments
GET    /api/notifications
POST   /api/notifications/{id}/read

# CV (M7)
POST   /api/employees/{id}/cvs
GET    /api/cv-extractions?status=PENDING_REVIEW
POST   /api/cv-extractions/{id}/skills/{skill_id}/approve|reject
```

## 4. Ánh xạ lỗi sang HTTP status

| Tình huống | Status | Ghi chú |
|---|:---:|---|
| Dữ liệu không hợp lệ, thiếu trường bắt buộc | 400 | Trả **danh sách lỗi theo từng trường** — mục 2.4.4 yêu cầu "thông báo lỗi cụ thể" |
| Chưa đăng nhập hoặc phiên hết hạn | 401 | |
| Sai vai trò, hoặc PM đụng dự án của PM khác | 403 | Tiêu chí nghiệm thu M1 |
| Không tồn tại | 404 | |
| Vi phạm trạng thái nghiệp vụ (dự án đã đóng, chuyển trạng thái sai) | 409 | UC002 alt 5a, UC003 alt 5a, UC004 alt 6a |
| Trùng mã dự án / mã công việc | 409 | UC001 alt 9b |
| Lỗi hệ thống, transaction rollback | 500 | UC003 alt 16a, UC004 alt 11a |

Định dạng lỗi thống nhất:

```json
{
  "error": {
    "code": "PROJECT_CLOSED",
    "message": "Dự án đã đóng, không thể phân rã công việc mới.",
    "fields": { "planned_end_date": "Ngày kết thúc không được nhỏ hơn ngày bắt đầu." }
  }
}
```

`message` bằng **tiếng Việt** (hiển thị thẳng cho người dùng); `code` bằng **tiếng Anh** để frontend
xử lý theo nhánh.

## 5. Phân công theo lô

`POST /api/assignments/confirm` nhận **nhiều dòng** vì UC003 cho phép phân công nhiều công việc một
lượt. Response phải phân biệt được ba tình huống:

- Tất cả hợp lệ → 200, trả danh sách kết quả.
- Một vài dòng lỗi (alt 15a) → trả rõ **chỉ số dòng và lý do**, để PM sửa bảng dự kiến.
- Toàn bộ dòng lỗi (alt 15b) → không lưu gì.

Toàn bộ chuỗi ghi nằm trong một transaction; lỗi giữa chừng thì rollback trọn vẹn (alt 16a).

## 6. Tác vụ chạy lâu

Mục 2.4.3 yêu cầu "cơ chế thông báo trạng thái xử lý rõ ràng" cho phân tích CV và sinh danh sách đề
xuất. Cách hiện thực (đồng bộ có loading, hay bất đồng bộ có polling) **chưa chốt** — quyết định khi
làm M5 và M7, dựa trên thời gian đo được thực tế.
