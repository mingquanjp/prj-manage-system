# Bước 5C — Đối chiếu hiện trạng → thiết kế đích

Cập nhật 15/10/2026 theo [quyết định hiện hành](database-review-decisions-2026-10-14.md). Nguồn hiện trạng là [000001](../../migrations/000001_init.up.sql) trong repo, **không phải kiểm kê dữ liệu/schema đang chạy trên Neon**. Đích là [bản hợp nhất](database-proposal-2026-10-14.md). Chưa viết/chạy migration, không sửa migration đã áp dụng.

**Tiếp nối 26/10/2026:** đích có 33 bảng, thêm job_levels và employees.job_level_id nullable thay seniority_level. Kiểm kê giá trị cấp bậc cũ (kể cả text ở phiên FE), đối chiếu/chuẩn hóa tên với HR rồi tạo danh mục và map ID; trống giữ NULL, giá trị chưa xác minh phải phân loại, không đoán cấp bậc từ tên vị trí. Giữ snapshot/audit cũ. Đây là hướng chuyển đổi tài liệu, không đã backfill/chạy migration; giữ nguyên 000001.

## 1. Kiểm kê 22 bảng hiện có

| Bảng hiện trạng | Đích | Kiểm tra/chuyển đổi cần thiết trước ràng buộc mới |
|---|---|---|
| users | Giữ sáu cột, không version/force-change; role sang user_roles | ADMIN có thể ánh xạ theo dữ liệu; PM lấy từ project, HR/DIRECTOR cần xác nhận; không suy từ chức danh |
| departments | Giữ | Kiểm tra tên/unique và tham chiếu, không mở CRUD ngoài quyền đã chốt |
| positions | Giữ | Như danh mục; không biến chức danh thành quyền |
| employees | Giữ, không profile version; email công ty; bổ sung address TEXT NULL theo chốt 20/10; job_level_id theo chốt 26/10 | Kiểm tra user 1–1, email/position/status; địa chỉ chưa có giữ NULL, không backfill giả hoặc siết NOT NULL; không giả quyền sửa allowlist |
| skill_categories | Giữ | Giữ ID và liên kết skills |
| skills | Giữ | Bỏ trạng thái danh mục; kiểm kê ý nghĩa trạng thái legacy, không xóa dữ liệu tham chiếu |
| employee_skills | Chỉ official hiện tại; không source/hiệu lực/version, đổi actor sang user | Approval cũ NULL phải kiểm tra nguồn, không bịa người/ngày duyệt; actor JOIN employee.user_id |
| projects | Thêm PTCM/ngày gốc/note, PM chọn ngay khi tạo; không version | PTCM không suy từ assignee/PM; ngày hiện hành không chắc là ngày gốc; cần dữ liệu xác nhận |
| project_members | Bỏ ở đích | Kiểm kê MANUAL, không assignment, joined_at; không DROP hoặc tạo assignment giả để ép khớp |
| task_groups | Giữ, UQ(id, project) | Giữ ID, đổi cascade có kiểm soát |
| tasks | Bỏ cây/type/actual_hours nguồn độc lập; đầu ra gộp description, criteria/mốc; không version/spec_version | Task group NULL, parent/type, ngày một phía, DONE/CLOSED cũ cần phân loại; không tự bịa nhóm/criteria/ngày nghiệm thu |
| task_requirements | Giữ; weight theo task do PTCM đặt, miền/kiểu chờ công thức; mandatory explicit | Weight DECIMAL cũ cần đối chiếu contract mới, không round/clamp về 1..5. Task không requirement cần xử lý trước giao dịch đích |
| recommendation_runs | Thêm snapshot/schema/actor user, bỏ spec_version | Không tái dựng snapshot quá khứ từ dữ liệu hôm nay; run legacy không đủ replay phải được nhận diện trong kế hoạch chuyển đổi |
| recommendation_items | Bỏ is_selected; score_details có contract đề xuất | Giữ điểm/căn cứ cũ, không gọi là kết quả công thức mới; chốt D12 rồi ánh xạ không mất giá trị |
| assignments | AI→RECOMMENDED khi đủ căn cứ, item/run ghép; ended_at/ended_by/end_reason trên lượt | Kiểm tra item đúng người/task, run dùng nhiều lần, điểm lặp khác item; không suy unassign/reassign đầy đủ chỉ từ timestamp |
| work_logs | hours→minutes, gắn đúng assignment, history trước/sau UPDATE/DELETE, deleted_at; không version | Decimal 0.01 giờ =0.6 phút; phải chốt làm tròn/xử lý trước chuyển INT. Log 0/thiếu note/không xác định lượt không ép khớp; source SYSTEM cũ không chứng minh người thật đã nhập |
| task_comments | Giữ | Bảo toàn tác giả/nội dung; quyền đọc mới không suy từ roster cũ |
| task_status_history | Đổi actor employee→user, occurred_at/reason | SYSTEM/actor NULL cũ không thể ép NN actor user bằng người giả; cần chính sách legacy trước migration |
| notifications | Recipient employee→user, resource/read_at | is_read=true không biết thời điểm đọc: không gán now() thành sự kiện thật; ref_task NULL cũng cần xử lý trước resource NN |
| employee_cvs | is_current, cloud storage_key unique/metadata/uploader | ACTIVE cũ có thể nhiều bản; không tự chọn latest; kiểm tra file thật, không giả uploader. file_url không là quyền truy cập |
| cv_extractions | Processing riêng review, một extraction/CV | Review APPROVED cũ không đồng nghĩa parser SUCCEEDED; nhiều extraction/CV cần bảo toàn, không gộp mất nguồn |
| cv_extracted_skills | Nguồn/item và request gốc riêng | Kiểm tra mục trùng/mapped/terminal, actor quyết định thiếu; không tạo request giả như vừa được người dùng gửi |

Giữ PK khi có thể. Mọi đổi employee actor → user dùng JOIN theo quan hệ, không cast/copy ID vì có thể tình cờ trùng số. Không tự xem dữ liệu cũ là vô hiệu chỉ vì không đủ cột mới.

## 2. 12 bảng mới trong thiết kế

| Bảng | Lý do có thêm | Nguồn khởi tạo |
|---|---|---|
| job_levels | Danh mục cấp bậc do công ty cấu hình, thay chuỗi tự do | Tên đối chiếu với HR; giá trị cũ chưa rõ không tự map |
| user_roles | Một người nhiều quyền hệ thống, khác trách nhiệm dự án | Role xác nhận; audit khi đổi |
| employee_skill_change_requests | Đề nghị và quyết định khác official | Yêu cầu mới; lịch sử cũ chỉ chuyển nếu đủ căn cứ |
| employee_profile_change_requests | Hồ sơ chờ HR duyệt | Yêu cầu mới |
| skill_request_task_evidence | 0..n đóng góp nội bộ cho yêu cầu | Dẫn chứng thực tế của người đề nghị |
| project_status_history | Lịch sử vòng đời dự án | Sự kiện mới; không bịa toàn lịch sử trước cutover |
| task_submissions | Từng lần nộp và review | Lần nộp mới/nguồn cũ đủ tin cậy, không tạo từ DONE đơn thuần |
| work_log_history | History UPDATE/DELETE trước/sau; không history tạo mới | Sự kiện mới; checkpoint nhập dữ liệu nếu có phải ghi là chuyển đổi, không giả thao tác gốc |
| work_log_change_requests | PTCM duyệt log đã khóa; ADD truy result_log_id | Đề nghị mới, không tạo cho mọi edit |
| auth_sessions | Session phía server, cookie secret chỉ lưu hash | Đăng nhập sau chuyển đổi |
| idempotency_records | Retry mutation không lặp | Mutation theo hợp đồng mới |
| audit_events | Thay đổi chưa có lịch sử miền đầy đủ | Sự kiện thật, không dựng lại quá khứ thiếu căn cứ |

**22 hiện có − 1 membership + 12 mới = 33 bảng đích.** Con số không thay việc kiểm kê dữ liệu. Nếu dữ liệu legacy không thỏa đích, phải chọn phương án bảo toàn/ngoại lệ chuyển đổi trước SQL; chưa thêm một kho/bảng legacy vào danh mục chỉ vì có rủi ro.

## 3. Trình tự chuyển đổi đề xuất — chưa thực thi

1. Kiểm kê read-only DB thật, version, enum, constraint, view/trigger và consumer; chỉ làm khi bước triển khai được yêu cầu. Xuất báo cáo sai khác không in dữ liệu cá nhân/secret.
2. Chốt các điểm mở ảnh hưởng phạm vi, phân loại hàng không chuyển được; tuyệt đối không dùng default giả để vượt NOT NULL.
3. Backup/snapshot và diễn tập restore trên bản sao thử nghiệm. Chọn cutover phù hợp hệ thống một process; không mặc định cần dual-write phức tạp.
4. Migration mới theo dependency: danh mục/user → hồ sơ/CV → project/task → run/items/assignment → submission/log/request/history/evidence. Tạo bảng trước, FK vòng/tham chiếu chéo thêm sau khi đủ đích UQ. Không sửa 000001.
5. Backfill có đối chiếu số dòng/ID/tổng, xác minh các FK ghép, trạng thái, actor, độ chính xác giờ→phút và giới hạn NULL. Chỉ sau kiểm tra mới siết NN/unique/check.
6. Cập nhật consumer và test quyền/giao dịch/retry. Bỏ cột/bảng cũ sau khi bảo toàn dữ liệu và xác nhận không còn phụ thuộc; không cùng lúc DROP để “dọn đẹp”.
7. Rollback phải bảo toàn dữ liệu đã phát sinh sau cutover; down migration xóa bảng mới không phải kế hoạch phục hồi đủ. Thử restore/forward-fix theo kịch bản đã chọn.

## 4. Tiêu chí nghiệm thu triển khai sau này

- Không mất lượt giao, căn cứ CV, nhật ký hoặc quyết định; số dòng/tổng trước–sau có giải thích.
- Không quyền phân công từ PM/roster; không reviewer độc lập; PTCM không được nhận task cùng dự án.
- Hai confirm cùng task/run, hai review, hai cập nhật effort xuyên dự án, hai pending cùng skill: không ghi trùng hoặc vượt ràng buộc.
- Chuyển role/status trong khi mutation chạy không vượt quyền; stale before_values/request.version kỹ năng không ghi đè.
- Retry request đã commit không tạo thêm log/assignment; cookie của session đã thu hồi không dùng được.
- Sửa log hiện tại không thay snapshot cũ; dữ liệu thiếu không thành 0/đạt/không đạt giả.

Đây là checklist chấp nhận, **không phải test đã pass**. Mapping đã tìm thấy các khoảng trống legacy; chưa có căn cứ để hứa migration tự động 100%.

Không suy ended_at/ended_by từ updated_at hoặc current_assignee đơn lẻ. Không tạo history ADD để ép dữ liệu cũ vào history UPDATE/DELETE. Nếu không xác định được log kết quả của yêu cầu ADD cũ, phải phân loại trước siết result_log_id. Upload metadata chỉ ánh xạ tệp thực có quyền; chuyển local/cloud cần kế hoạch riêng, không tự tải tệp thật. JWT/refresh cũ nếu có phải chấm dứt theo cutover auth đã kiểm thử, không triển khai đường refresh ở đích.
