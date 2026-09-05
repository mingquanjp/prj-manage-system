# Thiết kế API — V1

Thiết kế đích, chưa implement ngoại trừ GET /api/healthz trong skeleton.
Quyền ở roles_permissions.md; nghiệp vụ ở domain_rules.md, không lặp logic riêng từng endpoint.

## 1. Hợp đồng

/api, JSON snake_case, DTO theo vai trò; không trả entity DB trực tiếp.
Thời gian RFC3339 có offset; điểm 0–100; pagination page/limit/total, mặc định 20 tối đa 100.
PATCH cập nhật một phần; PUT thay toàn bộ requirements trước lần giao đầu.
Mutation nhạy cảm gửi expected_version và Idempotency-Key; cùng key khác payload →409.
Batch confirm 200 toàn bộ thành công; không partial success.
400 định dạng/validation; 401 phiên; 403 quyền; 404 không tồn tại; 409 trạng thái/version/unique; 413 quá lớn; 415 kiểu file; 429 rate limit; 500 lỗi chung.
Batch lỗi có task_id/index/code, không rò dữ liệu trái quyền. Điểm/rank/actor phải do server xác định.

## 2. Bề mặt dự kiến

```text
POST /api/auth/login
POST /api/auth/refresh
POST /api/auth/logout
GET  /api/auth/me
GET  /api/auth/csrf
POST /api/auth/change-password

GET  /api/users
PATCH /api/users/{id}
POST /api/users/{id}/lock
POST /api/users/{id}/unlock
POST /api/users/{id}/reset-password
POST /api/employees                       # tạo user + employee atomically
GET  /api/employees
GET  /api/employees/{id}
PATCH /api/employees/{id}                  # ADMIN, allowlist quản trị

GET|POST /api/departments
GET|POST /api/positions
GET|POST /api/skill-categories
GET|POST /api/skills
PATCH /api/departments/{id}
PATCH /api/positions/{id}
PATCH /api/skill-categories/{id}
PATCH /api/skills/{id}

POST /api/employees/{id}/profile-change-requests
GET  /api/profile-change-requests
POST /api/profile-change-requests/{id}/approve
POST /api/profile-change-requests/{id}/reject
POST /api/profile-change-requests/{id}/withdraw
POST /api/employees/{id}/skill-change-requests
GET  /api/skill-change-requests
POST /api/skill-change-requests/{id}/approve
POST /api/skill-change-requests/{id}/reject
POST /api/skill-change-requests/{id}/withdraw

GET|POST /api/projects
GET|PATCH /api/projects/{id}
POST /api/projects/{id}/start
POST /api/projects/{id}/complete
POST /api/projects/{id}/close
POST /api/projects/{id}/cancel
POST /api/projects/{id}/transfer-pm         # ADMIN, metadata tối thiểu
GET|POST /api/projects/{id}/task-groups
PATCH /api/task-groups/{id}
GET|POST /api/projects/{id}/tasks
GET|PATCH /api/tasks/{id}
PUT /api/tasks/{id}/requirements

GET  /api/projects/{id}/assignable-tasks
POST /api/projects/{id}/recommendations    # task_ids; mỗi task một run
GET  /api/recommendation-runs/{id}
POST /api/projects/{id}/assignments/confirm
POST /api/tasks/{id}/reassign
POST /api/tasks/{id}/unassign

GET /api/my/tasks
GET /api/my/contributions                 # lịch sử của mình, không task detail mới
POST /api/tasks/{id}/transitions          # to_status, reason, expected_version
GET|POST /api/tasks/{id}/comments
GET /api/tasks/{id}/work-logs
GET /api/notifications
POST /api/notifications/{id}/read

POST /api/employees/{id}/cvs
GET  /api/employees/{id}/cvs
GET  /api/cvs/{id}/download
GET  /api/cvs/{id}/processing
POST /api/cvs/{id}/retry
POST /api/cvs/{id}/withdraw
POST /api/cvs/{id}/publish
POST /api/cvs/{id}/reject
GET  /api/cv-extractions
POST /api/cv-extractions/{id}/close-review
POST /api/cv-extractions/{id}/skills/{extracted_skill_id}/map
POST /api/cv-extractions/{id}/skills/{extracted_skill_id}/reject
```

Skill CV map tạo/liên kết request; approve qua skill-change-requests, không có hai nguồn duyệt độc lập.
Withdraw CV chỉ bản PENDING; publish/reject ADMIN; retry chủ upload hoặc ADMIN theo ownership.
Transitions dùng bảng trạng thái chuẩn, không cho client PATCH status tùy ý.
Reassign trực tiếp manual; muốn dùng ranking mới thì unassign → run → confirm (hai thao tác có chủ đích, không giả atomic).
Confirm body chứa task_id, employee_id, expected_version, run_id/item_id nếu có, adjust_reason; idempotency gắn actor/operation.
Không endpoint tạo user rời, không DELETE lịch sử.

## 3. OpenAPI và giao diện thử

OpenAPI 3.1 viết tay tại api/openapi.yaml, Swagger UI/spec nhúng chỉ APP_ENV=dev, không CDN.
cookieAuth mô tả access_token; refresh/CSRF dùng hợp đồng riêng, login/health public tường minh.
Swagger cần bootstrap và gửi header CSRF, không chỉ tự gắn cookie.
chi.Walk đối chiếu method/path; bổ sung contract test request/response/error/security.
Các route trên là kế hoạch, không đưa tất cả vào spec như đã hoạt động khi backend chưa implement.
