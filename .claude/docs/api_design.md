# Thiết kế API — đồng bộ quyết định 15/10/2026

**Chốt tiếp nối 27/10/2026 — đề nghị kỹ năng của nhân viên:** luồng hiện tại chỉ bổ sung hoặc điều chỉnh; suy ADD khi chưa có official, UPDATE khi đã có, không cho chọn loại hoặc gửi đề nghị gỡ. Không thêm quyền gỡ trực tiếp cho Giám đốc/HR. Giá trị REMOVE và mô tả xử lý bản ghi cũ được giữ để tương thích thiết kế/lịch sử, không là chức năng tạo đề nghị mới. Không đổi cấu trúc DB; việc thu hồi kỹ năng ngoài luồng này chưa triển khai. Chốt này ưu tiên các mô tả ADD/UPDATE/REMOVE cũ đúng phần thao tác của nhân viên.

Nguồn: [sổ quyết định hiện hành](../../docs/design/database-review-decisions-2026-10-14.md), [dictionary/DBML 33 bảng (cấp bậc theo chốt 26/10/2026)](../../docs/design/database-proposal-2026-10-14.md), workflow trong thư mục này. Đây là hợp đồng thiết kế; chỉ GET /api/healthz có trong skeleton, không tuyên bố các route dưới đây đã chạy. Không triển khai auth/upload/migration trong đợt đồng bộ.

## 1. Quy ước và quyền

/api, JSON snake_case; DTO theo mục đích, không trả entity/raw snapshot. RFC3339 có offset, ngày YYYY-MM-DD theo Asia/Ho_Chi_Minh, phút nguyên; điểm 0–100 theo contract engine. Pagination mặc định 20/tối đa 100. PATCH chỉ cập nhật allowlist; PUT requirements chỉ trước giao lần đầu.

400 validation; 401 phiên; 403 quyền; 404 không tồn tại; 409 trạng thái/nền dữ liệu cũ/unique; 413 quá lớn; 415 kiểu file; 429 giới hạn; 500 lỗi chung. Batch cùng project nguyên tử, lỗi một dòng rollback toàn bộ, không lộ dữ liệu ngoài quyền.

Không expected_version chung cho mọi mutation. Yêu cầu kỹ năng dùng expected_request_version; hồ sơ/log dùng expected_before_values hoặc before_values request đã lưu. Đối chiếu đúng trường nghiệp vụ dưới khóa, không dùng updated_at như version. Confirm gợi ý đối chiếu đặc tả/requirements và kỹ năng liên quan với snapshot (cả sự có/không có), kiểm tra quyền/trạng thái hiện hành; không R05 acknowledgment. Tải hiện hành hiển thị riêng, không sửa rank cũ.

Idempotency-Key chỉ các mutation dễ tạo trùng, như tạo/bổ sung log, gửi yêu cầu hoặc batch confirm theo contract từng endpoint; không mọi API/login. Actor/operation/key/hash/result cùng transaction nghiệp vụ; cùng key khác nội dung 409. Retry kiểm tra quyền trước dựng lại kết quả; cửa sổ retry/retention còn cần đặc tả, không hứa vô hạn.

## 2. Tài khoản, danh mục và hồ sơ

```text
POST /api/auth/login
POST /api/auth/logout
GET  /api/auth/me
GET  /api/auth/csrf
POST /api/auth/change-password
GET  /api/users
PATCH /api/users/{id}                         # Admin, không quyền HR ngầm
POST /api/users/{id}/lock
POST /api/users/{id}/unlock
POST /api/users/{id}/reset-password
POST /api/employees                           # Admin tạo user + employee ban đầu nguyên tử, chốt 20/10
GET  /api/employees                           # DTO HR/chủ hồ sơ/tra cứu nghề nghiệp theo quyền
GET|PATCH /api/employees/{id}                 # PATCH HR allowlist, không official skill
GET|POST /api/departments                     # mutation HR
GET|POST /api/positions                       # mutation HR
GET|POST /api/job-levels                      # mutation HR, danh mục cấp bậc công ty
GET|POST /api/skill-categories                # mutation Giám đốc
GET|POST /api/skills                          # mutation Giám đốc, không status danh mục
PATCH /api/departments/{id}
PATCH /api/positions/{id}
PATCH /api/job-levels/{id}
DELETE /api/job-levels/{id}                   # HR, chỉ khi không còn hồ sơ tham chiếu
PATCH /api/skill-categories/{id}
PATCH /api/skills/{id}
POST /api/employees/{id}/profile-change-requests
GET  /api/profile-change-requests
POST /api/profile-change-requests/{id}/approve
POST /api/profile-change-requests/{id}/reject
POST /api/employees/{id}/skill-change-requests
GET  /api/skill-change-requests
POST /api/skill-change-requests/{id}/approve
POST /api/skill-change-requests/{id}/reject
POST /api/skill-change-requests/{id}/withdraw
```

Auth session phía server/cookie secret: tra token_hash, kiểm tra status/quyền/expiry/revocation; không access JWT hoặc refresh route. Cookie/CSRF/Origin theo security.md. Reset không ngầm khôi phục must_change_password. Timeout/reset chi tiết chờ đặc tả.

Theo chốt 26/10, HR cập nhật job_level_id nullable bằng ID từ job_levels; NULL là chưa cập nhật/không gán cấp bậc. Không nhận seniority_level tự do, không tạo danh mục ngầm từ tên nhập trên form hồ sơ. DTO chủ hồ sơ/HR có ID và tên cấp bậc; projection nghề nghiệp chỉ cần tên hiện hành, không cấp thêm quyền. Các route trên là thiết kế, chưa triển khai backend.

Profile request chỉ fullname/phone/address, chính chủ gửi HR duyệt. Email công ty/cơ cấu HR sửa trực tiếp, username không tự đổi theo email. Không thêm withdraw profile hoặc bổ sung Pending như skill request. Giám đốc duyệt kỹ năng có căn cứ/lý do, không tự duyệt. REMOVE gỡ official hiện tại giữ request; công nhận lại tạo dòng mới. Source MANUAL/CV ở request.

**Địa chỉ — chốt 20/10:** POST tạo user+employee không yêu cầu address; bỏ qua hoặc NULL khi chưa có thì giữ NULL, nếu cung cấp phải là chuỗi có nội dung sau trim. PATCH HR và proposed_values của yêu cầu hồ sơ chỉ nhận address có nội dung khi khóa này xuất hiện; không nhận NULL/chuỗi rỗng/chỉ khoảng trắng để xóa. Không có khóa address nghĩa là không đổi. before_values.address có thể NULL khi bổ sung lần đầu, server đối chiếu đúng giá trị nền dưới khóa lúc duyệt; pending không sửa employee. Yêu cầu chỉ sửa fullname/phone không bị chặn vì address hiện còn NULL. DTO chủ hồ sơ/HR được xem address nullable; DTO tra cứu nghề nghiệp/ứng viên/run không chứa địa chỉ. Không thêm endpoint hoàn thiện hồ sơ hoặc điều kiện address để cấp phiên/phân công/thực hiện task.

Bổ sung căn cứ kỹ năng PENDING gồm giải trình, 0..n assignment/submission đúng người, tối đa một CV dẫn trực tiếp và evidence_attachments metadata JSON. Route upload/bổ sung cụ thể chưa chốt; phải có expected_request_version, server xác thực kết quả upload/storage_key, tăng version và giữ terminal bất biến. Không nhận danh sách key tùy ý hoặc binary/base64. Không thêm bảng tệp.

## 3. Dự án, công việc và phân công

```text
GET|POST /api/projects                        # POST Giám đốc, chọn PM ngay khi tạo
GET|PATCH /api/projects/{id}                  # PM metadata; đổi pm_employee_id chỉ Giám đốc
POST /api/projects/{id}/start                 # PM
POST /api/projects/{id}/complete              # PM
POST /api/projects/{id}/close                 # PM, không pending effort
POST /api/projects/{id}/cancel                # PM, lý do
GET|POST /api/projects/{id}/task-groups        # mutation PTCM
PATCH /api/task-groups/{id}
GET|POST /api/projects/{id}/tasks              # mutation PTCM
GET|PATCH /api/tasks/{id}
PUT /api/tasks/{id}/requirements
GET /api/projects/{id}/assignable-tasks
GET /api/projects/{id}/assignment-candidates  # PTCM, toàn công ty đủ điều kiện
POST /api/projects/{id}/recommendations       # task_ids, mỗi task một run
GET /api/recommendation-runs/{id}
POST /api/projects/{id}/assignments/confirm
POST /api/tasks/{id}/reassign
POST /api/tasks/{id}/unassign
GET /api/my/tasks
GET /api/my/contributions
POST /api/tasks/{id}/transitions
POST /api/tasks/{id}/submissions
POST /api/task-submissions/{id}/review        # PTCM hiện hành
GET|POST /api/tasks/{id}/comments
```

Chỉ định technical_lead_employee_id là thao tác PM, kiểm tra điều kiện không thực hiện task cùng dự án; route PATCH project không cho trộn trường khác quyền. Không transfer-pm workflow riêng. description gồm đầu ra, acceptance_criteria riêng; đủ rõ trước giao và đóng băng scope/estimate/requirements sau first_assigned_at. Weight do PTCM đặt theo task, miền/chuẩn hóa theo công thức chưa chốt.

Confirm nhận task_id/employee_id/run_id/item_id khi dùng/adjust_reason và dữ liệu nền cần đối chiếu; không nhận score/rank làm chuẩn. MANUAL không item/score, sau run có run và lý do. Một commit/run. Không roster, PM không duyệt từng phân công. Giao lại manual đóng lượt cũ/tạo lượt mới cùng transaction, gợi ý giao lại phải thu hồi trước. Bộ ended do server ghi, không API PATCH assignment tùy ý. Trạng thái DONE/hủy kết thúc lượt; rejected review giữ lượt mở.

Nộp kết quả bắt buộc có result_summary (tóm tắt kết quả); evidence_links (liên kết minh chứng) tùy chọn. Server chụp log chưa xóa của đúng assignment. Không nộp thay, không effort_complete/confirm-logs/quality workflow riêng. Review ACCEPTED/REJECTED, rejected cần lý do; thiếu log không chặn đầu ra. Hủy khi đang review giữ kết luận NULL, hàng chờ lọc trạng thái task/project, không giả rejected. Không tự đổi Q/T.

## 4. Nhật ký và hiệu chỉnh

```text
GET|POST /api/tasks/{id}/work-logs
PATCH /api/work-logs/{id}                     # chính chủ khi còn quyền Edit, before + reason
DELETE /api/work-logs/{id}                    # xóa mềm, giữ history; không xóa vật lý
POST /api/tasks/{id}/work-log-change-requests
GET /api/work-log-change-requests             # PTCM hiện hành hoặc đóng góp chính mình
POST /api/work-log-change-requests/{id}/approve
POST /api/work-log-change-requests/{id}/reject
POST /api/work-log-change-requests/{id}/withdraw
GET /api/notifications
POST /api/notifications/{id}/read
```

Log nhận assignment_id/work_date/duration_minutes/description/over_daily_limit_reason, không employee/actor tùy ý. Chính chủ tự ghi/gửi, không proxy. Sửa/xóa trực tiếp đối chiếu dữ liệu nền và ghi before/after history; log đã deleted_at không sửa/xóa lại/khôi phục. >480 phút/ngày cần lý do, >1440 chặn xuyên dự án. Không nhận tổng actual_hours client.

Request ADD/UPDATE/DELETE, target_log_id chỉ UPDATE/DELETE; before do server xác thực/chụp, proposed NULL cho DELETE, reason gồm căn cứ. Không base_version/evidence riêng. PTCM hiện hành duyệt trong project IN_PROGRESS/COMPLETED; CLOSED/CANCELLED chỉ đọc. ADD approved trả result_log_id do server tạo cùng log/quyết định; UPDATE/DELETE ghi history.request_id. Retry không áp dụng lại, không đòi mọi APPROVED có history.

Người cũ gửi qua contributions theo assignment lịch sử không mở quyền task detail hiện hành. Correction không đổi snapshot nộp/run hoặc accepted_at. Notification không cấp quyền tài nguyên.

## 5. CV, trích xuất và read models

```text
POST /api/employees/{id}/cvs                  # chủ hồ sơ hoặc HR
GET /api/employees/{id}/cvs
GET /api/cvs/{id}/download                    # quyền private theo mục đích
GET /api/cvs/{id}/processing
POST /api/cvs/{id}/retry                      # retry kỹ thuật, không mở lại mục đã xử lý
POST /api/cvs/{id}/set-current                # chủ hồ sơ, không phê duyệt tệp
GET /api/cv-extractions
POST /api/cv-extractions/{id}/skills/{extracted_skill_id}/map
POST /api/cv-extractions/{id}/skills/{extracted_skill_id}/reject
```

Một cloud cấu hình chung, PDF text/DOCX ≤10MiB, không ảnh/OCR; provider/model/cách chạy API chưa chọn. Chỉ nhận diện tên kỹ năng/đoạn căn cứ, không suggested_level/years/confidence. Giám đốc xác định mức qua request trước công nhận. Map hợp nhất trùng trước quyết định; mục có request lấy quyết định từ request, không có request dùng row_outcome; không hai nguồn. Mapping trùng muộn còn cần đặc tả, không mở terminal. Không publish/reject tệp, resubmit/close-review riêng. HR upload không tự set-current.

Tra cứu PM/PTCM chỉ bốn nhóm nghề nghiệp được phép: nhận diện/công tác, kỹ năng, kinh nghiệm/kết quả, tải tổng hợp. Không contact/CV gốc/raw task/log dự án khác; current CV không tự cấp quyền tải. Manual không rank/score; GET run có snapshot_at và items được phép, không raw snapshot. run.schema_version là contract chung JSON, công thức/DTO chi tiết còn cần đặc tả.

## 6. OpenAPI và kiểm chứng

OpenAPI 3.1 viết tay/Swagger nhúng chỉ APP_ENV=dev, không CDN; cookieAuth mô tả session cookie, CSRF header/Origin riêng. Bootstrap Swagger phải gửi header đúng, không coi HttpOnly đủ chống CSRF. Contract test schema/error/security và chi.Walk method/path khi code có route; không đưa endpoint kế hoạch thành tính năng đã hoạt động.
