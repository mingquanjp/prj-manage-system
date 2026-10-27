# Trình bày database với cô — bản chuẩn bị 14/10/2026

Bản đọc chính: [thiết kế hợp nhất 5C](database-proposal-2026-10-14.md). Phụ lục phản biện: [mapping cũ → mới](database-step5c-migration-map-2026-10-14.md). Đã có bản đồ quan hệ đủ 33 bảng; không nhầm với ERD vật lý đầy đủ mọi cột/FK hoặc schema đã chạy.

Dictionary đề xuất: [5A — 11 bảng lõi](database-step5a-core-dictionary-2026-10-14.md) và [5B — 22 bảng còn lại](database-step5b-supporting-dictionary-2026-10-14.md). Tổng 33 sau chốt cấp bậc 26/10; 11+22 theo vị trí dictionary. Chưa có DDL/migration, ERD hợp nhất và một số chính sách vẫn cần hoàn thiện.

## 1. Thông điệp mở đầu (khoảng 45 giây)

“Em thiết kế database từ luồng nghiệp vụ: Giám đốc tạo dự án và chỉ định PM; PM chỉ định phụ trách chuyên môn. Phụ trách chuyên môn đặc tả, phân công và nghiệm thu; nhân viên thực hiện, ghi thời gian và nộp kết quả. Database cần lưu được ai đã làm gì, ở lượt phân công nào, kết quả được quyết định ra sao, đồng thời cung cấp dữ liệu đã được duyệt cho gợi ý nhân sự. Hệ thống chỉ gợi ý, người phụ trách chuyên môn quyết định.”

Không mở đầu bằng “em có 33 bảng”. Mở đầu bằng ba yêu cầu: đúng trách nhiệm, không mất lịch sử, giải thích được quyết định.

## 2. Bố cục 8 slide, khoảng 10–12 phút

| Slide | Nội dung trên slide | Điều cần nói |
|---|---|---|
| 1. Bài toán và phạm vi | 5 chủ thể; PTCM phân công/nghiệm thu; PM theo dõi; nguồn nhân lực chung | Không roster trước, không tối ưu tự động toàn dự án, không quản lý kế hoạch Word/Slack |
| 2. Bản đồ dữ liệu | 7 nhóm bảng bên dưới | Nhóm theo trách nhiệm, không theo màn hình và không một bảng cho mỗi actor |
| 3. Xương sống dự án | projects → task_groups → tasks → assignments; requirements gắn tasks | Một task thuộc một nhóm, một người chịu trách nhiệm hiện hành, có nhiều lượt giao trong lịch sử |
| 4. Thực hiện và nghiệm thu | assignment → logs; assignment → submissions; quyết định trên submission | Ghi giờ không thay nghiệm thu; nộp lại không ghi đè kết quả cũ; không bảng gán reviewer |
| 5. Kỹ năng và gợi ý | request/CV → duyệt → employee_skills → run/items → assignment | Pending không phải kỹ năng chính thức; snapshot gợi ý không phải quyết định giao |
| 6. Tính nhất quán | 3 ví dụ FK/unique; 3 ví dụ transaction | DB bảo vệ cấu trúc, service bảo vệ quyền và chuỗi thao tác; không nói FK giải quyết mọi thứ |
| 7. Ví dụ xuyên suốt | Task T101: giao → ghi → nộp → trả → nộp → đạt → hiệu chỉnh | Chứng minh mỗi bảng có lý do tồn tại bằng cùng một tình huống |
| 8. Tinh gọn và trạng thái | Bỏ membership đích/reviewer/plan/score lặp; phần đã thiết kế và phần còn kiểm chứng | Đề xuất có cơ sở nhưng chưa gọi toàn bộ schema đã triển khai hoặc đạt 3NF tuyệt đối |

Cấp bậc là danh mục công ty do HR quản lý, chọn qua job_level_id nullable. Bảng job_levels giúp dùng tên thống nhất và chặn xóa khi còn hồ sơ sử dụng; không tự thêm thang kỹ năng, thứ hạng hay công thức điểm.

## 3. Bản đồ 33 bảng đề xuất hiện tại

| Nhóm | Các bảng | Số | Dictionary |
|---|---|---:|---|
| Tài khoản và cơ cấu | users, user_roles, employees, departments, positions, job_levels | 6 | 5B mục 1–5.1 |
| Kỹ năng và yêu cầu | skill_categories, skills, employee_skills, employee_skill_change_requests, employee_profile_change_requests, skill_request_task_evidence | 6 | 5B mục 6–11 |
| CV | employee_cvs, cv_extractions, cv_extracted_skills | 3 | 5B mục 12–14 |
| Dự án/công việc | projects, project_status_history, task_groups, tasks, task_requirements, assignments, task_submissions, task_status_history, task_comments | 9 | 5A; comments ở 5B mục 15 |
| Effort | work_logs, work_log_history, work_log_change_requests | 3 | 5A mục 11–13 |
| Gợi ý | recommendation_runs, recommendation_items | 2 | 5B mục 16–17 |
| Hạ tầng | auth_sessions, idempotency_records, audit_events, notifications | 4 | 5B mục 18–21 |
| Tổng | 11 bảng ở 5A + 22 bảng ở 5B | 33 | Không roster/event phân công/token refresh riêng |

29 bảng thuộc sáu nhóm đầu, bốn bảng hạ tầng. Chỉ giữ dữ liệu có trách nhiệm; metadata upload không thêm bảng.

## 4. Ví dụ xuyên suốt để nói trên slide 7

Ví dụ minh họa thiết kế, chưa phải dữ liệu hoặc kết quả chạy SQL:

1. Dự án P01 có PM Lan, PTCM Minh. Task T101 “Xây API tra cứu” thuộc nhóm Backend, estimate 8 giờ, yêu cầu kỹ năng và tiêu chí nghiệm thu được khai báo.
2. Minh xem gợi ý run R01. Run giữ đầu vào lúc chạy; item I01 ứng viên An. Minh xác nhận → assignment A01 có bộ ended còn NULL; task hiện giao An, first_assigned_at được đặt. Không ghi project_members.
3. An khai log L01 =180 phút. work_logs giữ dữ liệu hiện tại; tạo mới không có history Edit. Tổng ghi nhận 3 giờ, không khẳng định đã khai đủ hoặc tự giảm estimate.
4. An nộp S01. Snapshot giữ L01/180 phút. Task PENDING_REVIEW. Minh trả lại vì thiếu một tiêu chí → S01 REJECTED có lý do, task IN_PROGRESS.
5. An sửa, ghi L02 =60 phút rồi nộp S02. Snapshot S02 có L01 và L02, tổng 240 phút. S01 không thay đổi. Minh chấp nhận → S02 ACCEPTED, task DONE, accepted_at và kết thúc A01 với actor Minh.
6. An phát hiện L01 thực tế 210 phút, gửi yêu cầu UPDATE. Khi PTCM hiện hành duyệt: L01 thành 210 phút, history giữ trước 180/sau 210 cùng actor/lý do, request APPROVED. Tổng hiện hành 270 phút; S01 vẫn 180 phút, S02 vẫn 240 phút; accepted_at không đổi.

Ý nghĩa: snapshot trả lời “lúc nộp có gì”; history trả lời “số liệu thay đổi thế nào”; request trả lời “ai đề nghị/ai duyệt”; log trả lời “số liệu hiện hành là gì”. Không cộng cả bốn nguồn. Không dùng ví dụ này tự suy điểm KPI hay tiêu chí chốt kỳ Q/T.

## 5. Những câu cô có thể hỏi

### Sao không một bảng nhân viên cho tất cả?

users quản lý đăng nhập, employees quản lý hồ sơ nhân sự; khóa tài khoản khác nghỉ việc. PM/PTCM là trách nhiệm theo dự án, không bảng người riêng. Một người có thể phụ trách dự án A và thực hiện task ở B.

### Không có project_members thì biết thành viên dự án thế nào?

Phiên bản này không lập đội trước. Danh sách người từng được giao suy từ assignments/tasks; danh sách người hiện có việc suy thêm trạng thái và người hiện hành. Không gọi mốc giao đầu là ngày gia nhập đội. Quyền không tự phát sinh chỉ vì từng được giao.

### Gợi ý toàn công ty có vô lý khi một người thường chỉ làm 1–2 dự án?

Toàn công ty là phạm vi tìm, không cam kết mọi người đều có thể nhận việc. Kiểm tra điều kiện hiện hành và tải xuyên dự án; số dự án đang có việc là thông tin tham khảo. Chưa có căn cứ nghiệp vụ để cấm cứng dự án thứ ba. Nếu doanh nghiệp cần đặt đội/giữ chỗ trước, đó là mô hình khác phải bổ sung nghiệp vụ trước.

### Tại sao không còn bảng sự kiện phân công riêng?

Một dòng assignments lưu lượt giao, nội dung giao bất biến. Bộ ended_at/ended_by/end_reason ghi một lần lúc kết thúc. Giao lại đóng lượt cũ/tạo lượt mới nguyên tử; thu hồi chỉ đóng lượt. Tối đa một lượt mở/task. Tải vẫn dựa trạng thái task/project, không chỉ nhìn ended_at. Không cần giữ bảng event có trách nhiệm lặp.

### Sao phải nhiều bảng cho nhật ký?

Hiện hành, lịch sử và yêu cầu chờ duyệt khác nhau. Edit khi chưa khóa không có yêu cầu nhưng vẫn có lịch sử. Request bị từ chối không được làm thay đổi giờ hiện hành. Gộp chúng dễ cộng nhầm đề nghị vào effort hoặc mất lịch sử sửa trực tiếp.

### Vì sao dùng JSON mà nói database quan hệ?

Quan hệ chính vẫn có FK. JSON chỉ dùng bản chụp có cấu trúc/phiên bản, không biến project/task/người thành JSON tùy ý. Đánh đổi: từng ID trong snapshot không có FK; server phải dựng, kiểm tra và đối soát. Không nói JSON tự bảo đảm toàn vẹn.

### Có chuẩn hóa hoàn toàn không?

Không tuyên bố như vậy khi chưa rà toàn bộ phụ thuộc hàm. Có dữ liệu lặp có chủ đích: project_id trên task để kiểm tra group cùng dự án; task/employee trên submission ràng buộc bằng FK ghép; log suy người/task từ assignment; current_assignee/accepted_at là projection ghi cùng giao dịch. Mỗi phần phải có nguồn chuẩn và cách đối soát.

### Cơ sở nào cho khóa và ràng buộc?

Nghiệp vụ lấy từ UC/workflow đã thống nhất. Cơ chế FK/UNIQUE/CHECK tham khảo [PostgreSQL Constraints](https://www.postgresql.org/docs/18/ddl-constraints.html); unique theo tập con tham khảo [Partial indexes](https://www.postgresql.org/docs/18/indexes-partial.html). Tài liệu kỹ thuật chứng minh cơ chế có hỗ trợ, không chứng minh schema ứng dụng đã chạy đúng.

### Đã hoàn thiện/chạy chưa?

“Em đã hợp nhất danh mục và thiết kế chi tiết phần lõi. Dictionary và DBML 33 bảng đã đồng bộ quyết định; contract công thức/API trích xuất/timeout và migration/test còn cần hoàn thiện. Em phân biệt thiết kế đề xuất với schema đang chạy; chưa thay DB thật và chưa coi những tình huống rà soát tài liệu là test pass.”

## 6. Đường hoàn thiện có điểm kết thúc

1. **5A đã viết:** dictionary 11 bảng lõi và quan hệ rút gọn; nhánh hủy submission và chất lượng effort còn giới hạn, không thêm ngoại lệ vào phần thuyết trình chính.
2. **5B đã viết:** dictionary 22 bảng còn lại; các điểm mở được giữ rõ, chưa coi toàn bộ trường/chính sách đã được duyệt.
3. **5C đã viết:** bản hợp nhất trỏ dictionary chuẩn + ERD quan hệ chính đủ 33 bảng (lược actor FK/thuộc tính) + ma trận FK ghép + bảng cũ→mới. ERD vật lý đầy đủ và DDL vẫn cần hoàn thiện sau các điểm mở.
4. **6:** migration mới trong môi trường thử, dữ liệu ví dụ và kiểm thử constraint/transaction/rollback/cạnh tranh. Kiểm kê dữ liệu cũ trước kế hoạch chuyển đổi; không sửa migration 000001 hoặc reset DB thật.

Đích để trình cô: ERD đọc được, mỗi bảng/cột quan trọng giải thích được, một ví dụ nhất quán và công bố đúng tình trạng kiểm chứng. Đích triển khai cao hơn: SQL/migration cùng test và kế hoạch bảo toàn dữ liệu.

Bổ sung khi cô hỏi ADD nhật ký đã khóa: PTCM duyệt, log mới và result_log_id trên yêu cầu cùng commit, không tạo history ADD giả. Request truy người gửi/người duyệt; history dành sửa/xóa mềm. Auth dùng server-side session, không refresh_tokens; tệp chứng cứ riêng lưu metadata JSON trên request, không công nhận kỹ năng tự động.
