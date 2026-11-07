# Bước 5C — Đối chiếu hiện trạng → thiết kế đích

Cập nhật 15/10/2026 theo [quyết định hiện hành](database-review-decisions-2026-10-14.md). Nguồn hiện trạng là [000001](../../migrations/000001_init.up.sql) trong repo, **không phải kiểm kê dữ liệu/schema đang chạy trên Neon**. Đích là [bản hợp nhất](database-proposal-2026-10-14.md). Chưa viết/chạy migration, không sửa migration đã áp dụng.

**Tiếp nối 26/10/2026:** đích có 33 bảng, thêm job_levels và employees.job_level_id nullable thay seniority_level. Kiểm kê giá trị cấp bậc cũ (kể cả text ở phiên FE), đối chiếu/chuẩn hóa tên với HR rồi tạo danh mục và map ID; trống giữ NULL, giá trị chưa xác minh phải phân loại, không đoán cấp bậc từ tên vị trí. Giữ snapshot/audit cũ. Đây là hướng chuyển đổi tài liệu, không đã backfill/chạy migration; giữ nguyên 000001.

## 0. Kiểm kê Neon thực tế 06/11/2026 — trước cập nhật

**Đã kết nối và chỉ đọc** project `ai-staffing` (`tiny-boat-30803147`), branch duy nhất/default `production` (`br-patient-meadow-azczu1hg`), database `neondb`. Tên branch không chứng minh đây là môi trường demo hay dữ liệu có thể xóa. Chưa tạo branch/snapshot, chạy DDL, backfill, seed/reset hoặc sửa `schema_migrations`.

Nguồn kiểm chứng: danh sách project/branch/database/bảng bằng Neon connector; SELECT metadata `information_schema`/`pg_catalog`, SELECT version/dirty, COUNT tổng và kiểm tra chất lượng dạng số tổng hợp. Không lấy giá trị password/hash, email, tên hoặc nội dung CV/hồ sơ. [SQL kiểm kê cấu trúc](../../scripts/neon-schema-inventory.sql) tái sử dụng với giao dịch READ ONLY; yêu cầu xác nhận bảng `schema_migrations` tồn tại trước chạy. Script chỉ ghi nhận có default, không xuất literal default. Chưa chạy chính file script qua psql; truy vấn SELECT kiểm kê bên trong đã chạy qua connector.

### Kết quả đã xác nhận

- PostgreSQL `18.6 (ac70818)`; migration version `1`, `dirty=false`.
- `public`: **22 bảng nghiệp vụ + schema_migrations**, 181 cột tính cả bảng migration, 65 index, 228 constraint gồm NOT NULL của PostgreSQL 18 và 13 trigger người dùng. Các số này là kiểm kê cấu trúc, không chứng minh logic nghiệp vụ đã đúng.
- Có dữ liệu: users 20, employees 20, departments 5, positions 10, skill_categories 8, skills 51, employee_skills 96. Các bảng nghiệp vụ còn lại trong public đều 0 dòng tại lúc SELECT count; không tự kết luận dữ liệu này là seed có thể xóa.
- Role legacy: ADMIN 1, PM 2, EMPLOYEE 17; chưa có `user_roles`, chưa có mapping xác nhận cho HR/DIRECTOR. Không tự cấp HR/DIRECTOR hoặc chuyển PM thành quyền toàn hệ thống.
- users không có employee: 0; nhóm email trùng sau lower/trim: 0. Đây là kiểm tra tại thời điểm đọc, không tự chốt chuẩn hóa email hoặc ràng buộc mới.
- 96 employee_skills không thiếu người/thời điểm duyệt; vẫn phải JOIN `approved_by_employee_id → employees.user_id` khi chuyển actor, không copy số ID.
- Cấp bậc có 3 giá trị không rỗng phân biệt và 1 hồ sơ chưa cập nhật. Chưa tự tạo danh mục/map khi chưa đối chiếu giá trị.
- Skill INACTIVE: 0; chưa dùng kết quả này làm quyền tự drop cột/status hoặc enum.

### Sai khác cần xử lý trước migration

**Schema Neon không trùng hoàn toàn migration 000001 trong repo:** `tasks.parent_task_id` và `tasks.task_type` có trong migration nhưng không còn ở Neon. Đối chiếu tự động tập bảng/cột tìm thấy hai sai khác này; chưa khẳng định default/index/function/check còn lại khớp từng biểu thức. Không có migration đánh dấu tiếp theo trong `schema_migrations`; chưa xác định ai/thời điểm đã sửa. Upgrade phải có preflight xử lý được cả DB tạo từ 000001 và Neon hiện tại, không sửa lịch sử 000001 hay giả version 2 đã chạy.

So với DBML đích: thiếu 12 bảng liệt kê ở mục 2; còn `project_members` ngoài thiết kế đích. `employees.address/job_level_id`, trách nhiệm PTCM, tiêu chí/ngày gốc task, duration_minutes/lượt giao nhật ký và các contract review/snapshot chưa được nâng cấp. Tập bảng cùng tên không có nghĩa cùng cột/ràng buộc. Bảng mục 1 tiếp tục là mapping thiết kế, không kết quả đã thực thi.

**Chốt cuối cùng 06/11:** thêm users.must_change_password BOOLEAN NOT NULL DEFAULT false; tài khoản cũ giữ false, cấp mật khẩu tạm ghi true, đổi thành công thay hash và đặt false cùng giao dịch. Mật khẩu tạm chưa có hạn, không thêm temporary_password_expires_at. Phương án hai cột trước đó bị thay thế; không thêm bảng. API/phiên hạn chế/gửi thư vẫn chưa triển khai. Migration 000002 chỉ thêm cờ, không dùng làm bằng chứng đã đồng bộ 33 bảng.

### Đường cập nhật cần hoàn thiện

1. Cấu trúc một cờ đã chốt, đồng bộ dictionary 5B/DBML và migration đơn lẻ 000002; review/diễn tập trước áp. Với phần còn lại, Xác nhận mapping role/cấp bậc và phạm vi dữ liệu giữ lại; không giả giá trị người duyệt/thời gian hoặc quyền mới.
2. Viết migration tiến thêm trong repo với preflight kiểm tra version, schema drift và dữ liệu. Tách thêm cấu trúc/backfill/siết ràng buộc/gỡ legacy; không export DBML nguyên xi. Các contract còn mở (ví dụ kiểu/miền weight) phải được xử lý rõ, không mặc định số minh họa thành quyết định.
3. Xác nhận phương án snapshot/phục hồi, diễn tập trên bản sao thử nghiệm của Neon và DB test tạo từ 000001; bảo toàn các ID và dữ liệu đang có. Không dùng DB chính làm TEST_DATABASE_URL.
4. Review diff/test và kết quả diễn tập trước áp vào branch đích; nếu dùng công cụ hoàn tất migration của Neon, phải hỏi người dùng trước theo yêu cầu công cụ. Chưa có migration được chuẩn bị/thực thi để duyệt ở thời điểm kiểm kê này.
5. Sau áp dụng: kiểm kê lại bảng/cột/constraint/version, đối chiếu số dòng và mapping, cập nhật trạng thái tài liệu. Chỉ lúc đó mới báo Neon đã đồng bộ; chỉ đủ số bảng không đủ nghiệm thu.

### Diễn tập migration 000002 — chưa áp branch đích

Migration repo: [up](../../migrations/000002_users_must_change_password.up.sql), [down](../../migrations/000002_users_must_change_password.down.sql). Up chỉ thêm cờ BOOLEAN NOT NULL DEFAULT false; không có IF NOT EXISTS để che schema sai. Down chỉ là phương án rollback sau ngừng consumer, mất trạng thái cờ nếu chạy, chưa diễn tập down hoặc cho phép chạy DB chính.

Đã diễn tập bằng Neon prepare_database_migration trên bản sao tạm `br-autumn-rain-az7u9ss5`, parent `br-patient-meadow-azczu1hg`, database `neondb`, migration ID `e7e02698-3a85-4e87-8c08-cba0f2866d57`. Bản sao đạt version 2/dirty=false; cột BOOLEAN, NOT NULL, default false; 20 tài khoản cũ đều false, không NULL, không cột hạn mật khẩu tạm. Đối chiếu count và checksum JSON có thứ tự trên cả 22 bảng (users bỏ đúng cột mới khi so) với parent cùng lúc: khớp toàn bộ. Không xuất bản băm mật khẩu hoặc nội dung hàng; checksum chỉ dùng đối chiếu nội bộ, không lưu như dữ liệu demo.

Công cụ Neon thực thi SQL trực tiếp thay runner Go, nên payload có preflight version=1/dirty=false, khóa bảng schema_migrations, kiểm tra lại version dưới khóa và cập nhật version 2 cùng giao dịch DDL. DDL tương ứng file up; version 2 không có nghĩa đã nâng cấp đủ 33 bảng. Payload chính xác phải dùng lại khi hoàn tất:

```sql
SELECT 1 / CASE WHEN count(*) = 1 AND bool_and(version = 1 AND NOT dirty) THEN 1 ELSE 0 END AS preflight FROM public.schema_migrations;
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';
LOCK TABLE public.schema_migrations IN ACCESS EXCLUSIVE MODE;
SELECT 1 / CASE WHEN count(*) = 1 AND bool_and(version = 1 AND NOT dirty) THEN 1 ELSE 0 END AS locked_preflight FROM public.schema_migrations;
ALTER TABLE public.users ADD COLUMN must_change_password boolean NOT NULL DEFAULT false;
UPDATE public.schema_migrations SET version = 2, dirty = false WHERE version = 1 AND dirty = false;
```

Hai lần prepare trước không thành công do bộ tách SQL của connector không hỗ trợ block DO và dấu chấm phẩy trong comment; đã kiểm tra sau các lỗi chỉ còn parent, không branch tạm thất bại tồn đọng. Payload cuối không có các cấu trúc đó và đã chạy thành công. Branch tạm thành công đang giữ để người dùng duyệt; công cụ complete_database_migration yêu cầu hỏi trước apply hoặc discard. Không tự xóa branch tạm trước khi xử lý lựa chọn.

**Kiểm tra local sau sửa:** go vet, go test ./... -count=1, go build đều đạt với GOTOOLCHAIN=local và cache tạm (cache mặc định bị sandbox chặn). Harness 9/9, kiểm kê 206 nguồn, diff --check và parser Mermaid 4/4 đạt. DBML vẫn 33 bảng, có đúng một cờ Boolean NOT NULL default false, không cột hạn mật khẩu tạm. Đây không phải kiểm thử gửi thư/auth runtime hoặc xác nhận toàn bộ schema đích.

**Kiểm tra parent sau diễn tập:** vẫn version 1/dirty=false, chưa có must_change_password. Chưa áp migration lên `production`, chưa thực thi gửi thư/đổi mật khẩu/chặn API. Chưa kiểm thử từ DB rỗng + 000001 hoặc API auth; phạm vi diễn tập hiện chỉ là nâng cấu trúc trên bản sao Neon đã kiểm kê.

## 1. Kiểm kê 22 bảng hiện có

| Bảng hiện trạng | Đích | Kiểm tra/chuyển đổi cần thiết trước ràng buộc mới |
|---|---|---|
| users | Nền sáu cột và role sang user_roles; thêm must_change_password theo chốt cuối 06/11; không thêm hạn mật khẩu tạm hoặc version | ADMIN có thể ánh xạ theo dữ liệu; PM lấy từ project, HR/DIRECTOR cần xác nhận; không suy từ chức danh |
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
