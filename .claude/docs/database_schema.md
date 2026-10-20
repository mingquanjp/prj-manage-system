# Thiết kế dữ liệu — hiện trạng và đích hiện hành

**Bổ sung 20/10/2026:** [quyết định địa chỉ](../../docs/business-baseline-2026-10-20.md): `employees.address TEXT NULL` hỗ trợ hồ sơ chưa bổ sung lúc tạo; cần có khi hoàn thiện hồ sơ. Allowlist yêu cầu fullname/phone/address; HR duyệt hoặc cập nhật trực tiếp, không xóa trắng địa chỉ đã có. Không DTO tra cứu/gợi ý, không thêm bảng/trạng thái, không chặn làm việc vì thiếu địa chỉ. Dictionary 5B/DBML là thiết kế đã đồng bộ; chưa migration.

Đồng bộ 15/10/2026 theo [quyết định hiện hành/Bàn giao](../../docs/design/database-review-decisions-2026-10-14.md). [Bản hợp nhất](../../docs/design/database-proposal-2026-10-14.md) là điểm đọc chính; [5A — 11 bảng](../../docs/design/database-step5a-core-dictionary-2026-10-14.md), [5B — 21 bảng](../../docs/design/database-step5b-supporting-dictionary-2026-10-14.md), [DBML 32 bảng](../../docs/design/database-proposal-2026-10-14.dbml), [mapping chuyển đổi](../../docs/design/database-step5c-migration-map-2026-10-14.md).

Đích không roster/assignment_events/refresh_tokens. Session phía server; nhật ký suy người/task từ assignment, xóa mềm deleted_at; history UPDATE/DELETE; yêu cầu ADD đã duyệt truy result_log_id. PTCM duyệt hiệu chỉnh. Official chỉ năng lực hiện tại, request giữ lịch sử/source; request.version kỹ năng vẫn còn. Upload minh chứng dùng metadata JSON trên request.

Hiện trạng migration 000001 vẫn 22 bảng; chưa thay code/migration/DB thật. Không sửa migration đã áp dụng. Baseline V2 và ghi chép dưới đây chỉ là lịch sử, không dùng khôi phục cột/quyền/luồng đã bỏ. Các tài liệu gốc docs/thesis-context chỉ đọc.

<details>
<summary>Thiết kế V2 và các mốc trước rà soát 15/10 — lịch sử, không phải hợp đồng hiện hành</summary>

# Thiết kế dữ liệu — hiện trạng và đích V2

**Điểm đọc chính từ bước 5C (14/10/2026):** [bản thiết kế hợp nhất](../../docs/design/database-proposal-2026-10-14.md) có danh mục 34 bảng, ERD quan hệ chính và nguồn dữ liệu; [đối chiếu hiện trạng → đích](../../docs/design/database-step5c-migration-map-2026-10-14.md) nêu các cổng bảo toàn dữ liệu. Đây là bản đề xuất trình bày, chưa DDL/migration; phần V2 bên dưới giữ để đối chiếu lịch sử.

**Tiếp nối thiết kế 14/10/2026:** [dictionary 5A](../../docs/design/database-step5a-core-dictionary-2026-10-14.md) có 12 bảng lõi; [dictionary 5B](../../docs/design/database-step5b-supporting-dictionary-2026-10-14.md) có 22 bảng còn lại; [pitching](../../docs/design/database-pitch-2026-10-14.md) giải thích nhóm bảng và ví dụ. Đã sửa lỗi đếm 13/21, tổng vẫn 34. Đây là đề xuất chưa migration, còn cần ERD hợp nhất và chốt các điểm mở; không dùng các trường proxy/quality/R05 trong V2 dưới đây để khôi phục nghiệp vụ đã bỏ.

**Quyết định mới về membership:** [4C](../../docs/design/database-step4c-shared-pool-2026-10-13.md) bỏ lưu project_members riêng ở thiết kế đích, suy từ assignments/tasks; không bảng số dự án/tải. Mô tả 000001 phía dưới vẫn là hiện trạng migration cũ; yêu cầu V2 ghi FROM_ASSIGNMENT đã bị thay thế. Chưa DROP/chuyển dữ liệu; cần kiểm kê dữ liệu/dependency thật trước migration mới. Phương án hiện tại 34 bảng, dictionary còn mở.

> **Chưa đồng bộ baseline [2026-10-12](../../docs/business-baseline-2026-10-12.md).** Enum vai trò, quyền nhân sự/duyệt, trạng thái công bố CV và R01–R05 cũ dưới đây không quyết định nghiệp vụ hiện hành. Bổ sung HR chưa có migration; không dùng allowlist hồ sơ cũ như danh sách trường cần duyệt đã chốt. Đợt này không sửa schema hoặc dữ liệu.

## Yêu cầu đồng bộ nghiệp vụ 2026-10-12 trước triển khai

- Quyền chạy gợi ý, xác nhận/thay/thu hồi phân công thuộc PTCM hiện hành, không phải PM owner. PM giữ bắt đầu/hoàn thành/đóng dự án và hiệu chỉnh effort theo quyền.
- PTCM không nhận task cùng dự án. Kiểm tra cả khi phân công và khi chỉ định/thay PTCM, phối hợp transaction; lưu lịch sử, không xóa đóng góp.
- Chỉ PTCM nghiệm thu; bỏ cấp quyền reviewer độc lập hoặc PM chỉ định người review. Không tự nghiệm thu đầu ra đã thực hiện trước khi chuyển trách nhiệm.
- Tra cứu PM/PTCM chỉ thông tin nghề nghiệp/tổng hợp được phép; không tải CV gốc người khác hoặc raw task/log dự án khác từ quyền tra cứu.
- Kế hoạch chia sẻ ngoài hệ thống; không cần bảng/endpoint gửi–duyệt kế hoạch hoặc tích hợp Slack.
- Bảng/endpoint/enum phía dưới là thiết kế V2 cũ, **không phải hợp đồng triển khai cuối**. Chưa đổi DDL, migration hoặc code trong đợt này; thiết kế vật lý trách nhiệm PTCM cần review riêng.

## 1. Phân biệt nguồn

- Hiện trạng mã migration: migrations/000001_init.up.sql, 22 bảng nghiệp vụ/15 enum.
- Lượt kiểm tra DB trước ghi nhận version 1, dirty=false; không chứng minh mọi cột/constraint đúng chỉ từ version.
- Thiết kế đích đã đồng ý: [baseline V2](../../docs/reviews/system-design-baseline-v2.md) và các tài liệu miền hiện hành.
- Thiết kế dữ liệu chi tiết đồng bộ đồ án ngày 2026-09-13: [database-design-v2.md](../../docs/design/database-design-v2.md). Đây là thiết kế đích, không chứng minh migration đã được triển khai.
- DoAn.pdf/TeX trong docs/thesis-context là tài liệu gốc chỉ đọc, chưa đồng bộ V2.

Không sửa migration 000001 đã áp dụng. Các thay đổi dưới đây CHƯA được tạo migration/chạy DB.
Không ghi số thứ tự migration cố định trước khi chia dependency và kiểm tra dữ liệu thật.

## 2. Hiện trạng 000001

Users/employees/departments/positions; skills/categories/employee_skills;
projects/project_members/task_groups/tasks/task_requirements;
recommendation_runs/items; assignments/work_logs/comments/task_status_history;
notifications; employee_cvs/cv_extractions/cv_extracted_skills.

Hiện vẫn có parent_task_id/task_type, task group nullable, is_selected; work_logs đã có work_date/hours_spent. V2 giữ work_date, không coi cột này là lỗi cần bỏ.
Chưa có PAUSED/CANCELLED cho task; chưa có bảng yêu cầu riêng, session/refresh hoặc snapshot đầy đủ.
Một số FK cascade có thể xóa chuỗi lịch sử. Không coi schema hiện tại đã bảo vệ đủ nghiệp vụ V2.

## 3. Bảng/cột thiết kế đích

| Miền | Thay đổi bắt buộc |
|---|---|
| User/employee | Tạo atomically; version cho dữ liệu cần chống ghi đè; cơ chế buộc đổi mật khẩu |
| Official skills | Giữ unique employee/skill; approval actor/time bắt buộc; hiệu lực, version; REMOVE không xóa; ADD tái kích hoạt cùng identity nếu đã thu hồi |
| employee_skill_change_requests | request_id, employee_id, skill_id, ADD/UPDATE/REMOVE, MANUAL/CV, proposed/before values, base_version, evidence, extracted_skill_id (không unique toàn vòng đời), attempt_no/supersedes_request_id cho CV, submitter/time, status, reviewer/time/reason |
| employee_profile_change_requests | Trường đề nghị được allowlist, before/proposed, base_version, actor/time, quyết định/reason; không nhận role/status |
| Projects | version, ngày kế hoạch ban đầu/hiện hành, project_status_history; audit chuyển PM ghi owner cũ/mới/lý do |
| Tasks | Group NOT NULL cùng project; bỏ parent_task_id/task_type; PAUSED/CANCELLED; version, spec_version, first_assigned_at, accepted_at; acceptance_criteria, ngày kế hoạch ban đầu/hiện hành; effort_quality và lý do |
| Assignments | Append-only; run_id, item_id, adjust_reason, score server; R05 history_change_ack/reason/acknowledged_revision/time; truy vết task/employee/người xác nhận |
| assignment_events | ASSIGN/REASSIGN/UNASSIGN, task, assignment cũ/mới, actor/time/reason, handover_note khi việc đã bắt đầu |
| Recommendation runs | engine_version, params, snapshot_at, spec_version, input_snapshot/output_snapshot JSONB có schema version, history_revision của từng ứng viên đã xét; không chứa CV thô; raw snapshot chỉ nội bộ |
| Recommendation items | Top-N/rank, skill/workload/perf/tổng, matched/missing, explanation; bỏ is_selected, lựa chọn suy từ assignment |
| work_logs | assignment_id, task_id, employee_id, work_date DATE, duration_minutes INTEGER >0, description; created_by/updated_by, version, voided_at; phân biệt người ghi và người thực hiện |
| work_log_history | ADD/UPDATE/VOID, before/after, actor/time/reason, request_id nếu có; append-only, không xóa bản gốc |
| work_log_change_requests | task/employee/assignment, operation ADD/UPDATE/VOID, target_log_id (NULL cho ADD), before/proposed, base_version, requester/time/reason, exception_type (MISSED_START_RESUME nếu R01), evidence/context, status, reviewer/time/decision_reason |
| task_submissions | task/assignment, kết quả, link minh chứng tùy chọn, submitter/time, log revision snapshot, effort_complete, incomplete_reason, proxy_reason tách riêng; lần nộp cũ không ghi đè. Quality/reason và snapshot xác nhận lúc nghiệm thu/sau correction lưu trong audit cùng tham chiếu submission; tasks giữ quality hiện hành |
| CV | cv_status thêm PENDING/REJECTED; tách processing_status và review_status; job lease/attempt/error code; suggested_level nullable; mục canonical/current_request_id hoặc attempt_no hiện hành, nguồn dòng hợp nhất, quyết định NO_CHANGE có audit; review_closed_at/actor/revision |
| auth_sessions | session_id, user_id, created_at, expires_at, revoked_at; hạn tuyệt đối 7 ngày |
| refresh_tokens | hash unique, session_id, expiry/revoked/replacement; giữ token đã dùng để phát hiện reuse |
| idempotency records | actor + operation + key unique, request hash, response/result; cùng transaction nghiệp vụ |
| Audit | before/after có giới hạn dữ liệu, actor/time/reason cho spec, quyền, ownership; append-only |

Tên vật lý/DDL cuối cùng cần được review khi viết migration; bảng trên là hợp đồng dữ liệu tối thiểu, không DDL đã chạy.
Request skill terminal: APPROVED/REJECTED/WITHDRAWN; official update và quyết định cùng transaction.
CV publication PENDING/ACTIVE/REPLACED/REJECTED/ARCHIVED;
processing QUEUED/PROCESSING/SUCCEEDED/FAILED;
review PENDING_REVIEW/APPROVED/PARTIALLY_APPROVED/REJECTED/NO_CHANGES; close_reason phân biệt NO_SKILLS_FOUND/NO_CHANGE/ALL_WITHDRAWN.
Effort request PENDING → APPROVED/REJECTED/WITHDRAWN; project cancel kết thúc pending bằng REJECTED với reason PROJECT_CANCELLED.
Effort quality COMPLETE/INCOMPLETE/UNRELIABLE; lưu actor/time/reason, không đồng nhất với trạng thái task.
Tên enum/cột mới là thiết kế đích tối thiểu, phải chuyển sang DDL được review ở bước thiết kế chi tiết.

## 4. Ràng buộc và nguồn sự thật

- Unique project_code; unique task_code trong project; unique group_name trong project.
- Composite FK tasks(project_id, task_group_id) → task_groups(project_id, task_group_id) với unique tương ứng.
- Estimate >0; weight nguyên 1–5; ngày task cùng NULL hoặc đủ cặp, đúng thứ tự.
- Task có ≥1 requirement và ngày trong project: service transaction, không CHECK truy vấn bảng khác.
- Unique employee/skill chính thức; ADD tái kích hoạt không INSERT trùng; kiểm tra version/hiệu lực lúc duyệt. Catalog INACTIVE vẫn cho ADD/UPDATE năng lực qua xét duyệt. INACTIVE chỉ chặn requirement mới; task chưa từng giao phải sửa requirement inactive, task từng giao giữ frozen requirement và dùng official skill còn hiệu lực khi lọc/chấm điểm/gợi ý/manual/reassignment.
- Partial unique request PENDING/người/skill và PENDING/extracted_skill_id khi khác NULL; KHÔNG unique extracted_skill_id trên mọi request. Unique extracted_skill_id/attempt_no cho CV.
- Mục canonical có tối đa một attempt APPROVED; chỉ resubmit sau REJECTED/WITHDRAWN trước review_closed_at và khi publication cho phép. Khóa mục/extraction khi đổi current attempt/duyệt/đóng; review đếm mục không đếm attempts.
- Current decision chỉ có một nguồn REQUEST hoặc ROW_DECISION; NO_CHANGE sau attempt terminal dùng quyết định dòng mới và current_request_id=NULL, giữ mọi liên kết request cũ.
- Một CV ACTIVE/người bằng partial unique; xuất bản khóa hồ sơ người để tránh hai lần duyệt đồng thời.
- Unique run/employee, run/rank; rank >0; mọi điểm persist 0–100.
- Một assignment có run_id khác NULL cho mỗi run (partial unique); item phải khớp run/task/employee.
- RECOMMENDED cần run/item/score; MANUAL item/score NULL, có thể có run và reason. Kiểm tra quan hệ và lý do trong service.
- Work log ràng buộc assignment/task/employee khớp lịch sử; phút nguyên >0, ngày không tương lai. R01 chỉ vượt mốc thao tác start/resume qua request được duyệt, không vượt assignment/project start/terminal/tổng ngày hoặc backdate history.
- Bộ log rỗng có tổng suy ra0 nhưng submission R02 phải lưu thiếu dữ liệu; UI không khẳng định thực tế0h. Không tạo row duration_minutes=0.
- Tổng log hiệu lực employee/date ≤1440 xuyên project: transaction có khóa chung, không CHECK riêng một row. >480 cần reason; không unique employee/date/task vì cho nhiều lần ghi.
- Một yêu cầu PENDING trên mỗi log đích UPDATE/VOID; ADD chống retry bằng idempotency, không cấm nhiều phần việc hợp lệ.
- Review/approve kiểm tra phiên bản log và request; không cho hai yêu cầu cũ ghi đè lẫn nhau. Đổi ngày khóa cả ngày cũ/mới.
- project close chặn pending effort; cancel kết thúc chúng cùng transaction. Sửa nhật ký không thay accepted_at.
- current_assignee là nguồn người hiện tại; assignment/events là lịch sử. PENDING_ASSIGNMENT không người; trạng thái thực thi phải có người.
- Thay người trực tiếp: khóa task, append assignment/event và đổi current_assignee/status ASSIGNED atomically; lỗi giữ A/trạng thái cũ, thành công B phải Start. Không cần PENDING_ASSIGNMENT trước thao tác, không chuyển log/kết quả của A sang B. Thu hồi là giao dịch riêng về PENDING_ASSIGNMENT ngay; recommendation giao lại chỉ sau thu hồi, không tự khôi phục A.
- task_submissions gắn task/assignment, tách submitted_by khỏi người thực hiện. Mỗi lần nộp giữ log ID/version snapshot; history cho phép tái dựng. Nghiệm thu/xác nhận lại quality sau correction ghi actor/time/reason và snapshot hiện hành trong audit, không ghi đè snapshot nộp cũ.
- Bàn giao PM ở project COMPLETED chỉ đổi owner hiện tại và ghi audit old/new/reason cùng transaction; không đổi project state hoặc cấp thêm quyền thực thi. PM mới xử lý correction/đóng theo state machine; PM cũ mất quyền.
- actual_hours là giá trị suy ra/cache từ SUM(duration_minutes của log không void)/60, không nhận trực tiếp từ client; đối soát theo task/người/project.
- Tổng project gồm cả task CANCELLED, tách phần hủy khi hiển thị; pending request không tham gia tổng.
- Snapshot đầu vào scorer nhất quán trong một giao dịch đọc; confirm khóa/version kiểm tra điều kiện hiện hành.
- History revision per employee (hoặc tương đương) tăng cùng transaction khi dữ liệu lịch sử/eligibility đổi: log, quality, task nghiệm thu và tạo/rút/từ chối/duyệt correction. Phát hiện cả mẫu mới ngoài tập10 cũ; thay đổi đóng góp kiểm tra mọi employee liên quan.
- R05 acknowledgment ghi ở assignment, không UPDATE output run; revision ack stale thì rollback batch. Raw snapshot không trả theo vai trò ADMIN/PM.
- Read model PM và ADMIN chung official source, không nhân đôi bảng năng lực.
- Không API xóa vật lý entity có lịch sử; thay cascade nguy hiểm bằng RESTRICT/NO ACTION phù hợp, kiểm tra cả đường gián tiếp.

## 5. Migration và backfill

1. Kiểm tra COUNT/null/duplicate/orphan/enum thật trên DB đích; không dùng ước lượng thống kê thay COUNT.
2. Tách migrations theo auth → request/version → task/assignment/time → recommendation → CV, điều chỉnh theo FK.
3. Add nullable/backfill/validate rồi mới NOT NULL/drop. Task group thiếu cần gán nhóm hợp lệ có kế hoạch.
4. Official skill cũ thiếu approval cần phân loại, không tự gán ADMIN giả để vượt constraint.
5. Giữ work_date cũ; kiểm tra hours_spent×60 trước chuyển phút nguyên. Giá trị không chia hết phút, thiếu assignment/description/nguồn xác nhận phải phân loại legacy và xử lý riêng, không làm tròn âm thầm hoặc bịa lịch sử. Legacy chưa xác minh không làm mẫu cá nhân.
6. Test up/down trên DB thử; rollback mất dữ liệu phải có backup/khôi phục, không quảng cáo down luôn an toàn.
7. Sửa seed/test/OpenAPI/ERD cùng phiên bản sau khi DDL được duyệt.

Identity PK, snake_case, timestamptz; giờ/điểm hiển thị decimal, thời lượng chuẩn lưu phút nguyên; work_date là DATE địa phương, audit là timestamptz.
Giữ trigger updated_at hiện có; version cập nhật atomically khác với timestamp.
Seed Password@123 chỉ dev; không chạy seed/reset DB thật trong lượt cập nhật tài liệu này.

## 6. Rà soát tính cần thiết — 2026-09-18

Giữ 33 bảng thiết kế đích, không thêm bảng hoặc chức năng. deliverable/acceptance_criteria bắt buộc ngay khi tạo task: NOT NULL và không chỉ có khoảng trắng. Sự kiện phân công dùng FK ghép lượt/task; ASSIGN chỉ có lượt mới, REASSIGN có hai lượt khác nhau, UNASSIGN chỉ có lượt cũ. Khóa task để kết thúc đúng lượt đang hiệu lực, không xác định lượt chỉ theo người hoặc timestamp; giữ log/submission của lượt cũ khi cùng người nhận lại task.

Phân định nguồn chuẩn: lịch sử trạng thái ở task/project_status_history; lịch sử phân công ở assignment_events; phiên bản nhật ký ở work_log_history; các biến động khác và nội dung xác nhận nghiệm thu/chất lượng ở audit_events. Các bản ghi bổ sung nhau trong cùng giao dịch, không nhân đôi toàn bộ cùng một lịch sử.

output_snapshot là kết quả đầy đủ bất biến; items là Top-N được tách từ cùng kết quả. matching_score ở assignment là bản sao điểm quyết định do server lấy từ item, không sửa độc lập. Giữ các trường R05 và ba loại version. project_members chỉ ghi nhận đã tham gia/mốc đầu, không lưu nhiều đợt; FROM_ASSIGNMENT là nguồn tạo trong luồng hiện hành, MANUAL không tự tạo thêm chức năng.

Bỏ markdown_content trong thiết kế đích CV vì chưa có nhu cầu lưu thêm bản Markdown. suggested_level vẫn nullable và không tự suy từ tên skill; mức chính thức lấy từ yêu cầu được ADMIN duyệt. confidence_score là điểm đối sánh tên, không phải xác suất năng lực. Chi tiết tại mục hoàn thiện 2026-09-18 của database-design-v2.md. Không áp dụng thay đổi vào migration hoặc DB trong lượt chỉnh tài liệu.

</details>
