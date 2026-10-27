# Bước 5B — Dictionary 22 bảng còn lại

**Chốt tiếp nối 27/10/2026 — đề nghị kỹ năng của nhân viên:** luồng hiện tại chỉ bổ sung hoặc điều chỉnh; suy ADD khi chưa có official, UPDATE khi đã có, không cho chọn loại hoặc gửi đề nghị gỡ. Không thêm quyền gỡ trực tiếp cho Giám đốc/HR. Giá trị REMOVE và mô tả xử lý bản ghi cũ được giữ để tương thích thiết kế/lịch sử, không là chức năng tạo đề nghị mới. Không đổi cấu trúc DB; việc thu hồi kỹ năng ngoài luồng này chưa triển khai. Chốt này ưu tiên các mô tả ADD/UPDATE/REMOVE cũ đúng phần thao tác của nhân viên.

**Chốt tiếp nối 26/10/2026:** cấp bậc theo công ty: thêm job_levels (mục 5.1), employees.job_level_id nullable thay seniority_level; HR quản trị và chọn từ danh mục. Tổng hiện hành 33 bảng, chưa DDL/migration. [Quyết định hiện hành](database-review-decisions-2026-10-14.md#quyết-định-tiếp-nối-26102026--cấp-bậc-theo-công-ty-hiện-hành).

**Bổ sung 20/10/2026:** [quyết định địa chỉ](../business-baseline-2026-10-20.md#quyết-định-tiếp-nối--địa-chỉ-khi-hoàn-thiện-hồ-sơ) thêm `employees.address TEXT NULL`; cần bổ sung khi hoàn thiện hồ sơ nhưng không bắt Admin nhập khi tạo. Allowlist yêu cầu hồ sơ có address; không cho xóa trắng. Tại mốc 20/10 tổng vẫn 32 bảng; chốt 26/10 ở trên thay phần cấp bậc/tổng số, chưa DDL/migration.

**Đồng bộ 15/10/2026:** ưu tiên [quyết định hiện hành và Bàn giao](database-review-decisions-2026-10-14.md), gồm xác nhận result_log_id. Chỉ thiết kế tài liệu, chưa DDL/migration. Các bước 1–4 là lịch sử.


Ngày 14/10/2026. **Bản thiết kế đề xuất để review, chưa DDL/migration.** Cùng [5A](database-step5a-core-dictionary-2026-10-14.md) bao phủ danh mục 33 bảng sau chốt 26/10; bao phủ danh mục không có nghĩa mọi chính sách đã chốt.

Nguồn: [3B](database-step3b-effort-profile-skills-cv-2026-10-13.md), [3C](database-step3c-roles-recommendation-technical-2026-10-13.md), [4C](database-step4c-shared-pool-2026-10-13.md), [quy trình kỹ năng](../../.claude/docs/skill_approval_workflow.md), [hồ sơ](../../.claude/docs/employee_profile_workflow.md), [bảo mật](../../.claude/docs/security.md). Quyết định nghiệp vụ mới ưu tiên hơn V2 và migration hiện có.

## Quy ước

`NN` bắt buộc; `?` nullable, default NULL. PK đơn là INT identity; FK INT, mặc định NO ACTION. Trường NN không ghi default phải được cung cấp. `timestamps` viết tắt hai cột created_at/updated_at TIMESTAMPTZ NN default now(), đường UPDATE phải cập nhật updated_at. `version` INT NN default 1, >0, service tăng. Trường trạng thái dùng tập hữu hạn, ENUM hay TEXT+CHECK chốt khi làm DDL. Text bắt buộc không trắng. Actor trỏ users; người có hồ sơ trỏ employees. Không suy quyền từ FK.

Các độ dài/kiểu kế thừa hiện trạng là đề xuất tương thích, không phải tiêu chuẩn nghiệp vụ. Không tự mở thêm trường hành chính. JSON phải có contract/validator và phiên bản định dạng: recommendation dùng run.schema_version, attachments dùng schema metadata cố định theo mục 9, các before/proposed/snapshot theo contract tương ứng; không dùng JSON chứa ID để thay quan hệ cần FK.

## 1. users

Một tài khoản đăng nhập; không lưu lại PM/PTCM như quyền toàn hệ thống.

| Cột | Kiểu / NULL / default | Nguồn và nghĩa |
|---|---|---|
| user_id | PK | Identity |
| username | VARCHAR(50) NN | Định danh đăng nhập UQ |
| password_hash | VARCHAR(255) NN | Server bcrypt; không DTO/audit |
| status | user_status NN / ACTIVE | Admin khóa/mở theo quyền |
| timestamps | Như quy ước | Server |

Role nằm ở user_roles. Reset/đổi role/khóa thu hồi session cùng transaction. Chính sách mật khẩu tạm hết hạn/kênh cấp còn mở D13; không còn cờ buộc đổi mật khẩu; chưa đủ hợp đồng để triển khai toàn bộ auth. Chuẩn hóa username phải thống nhất với seed/login trước migration, không tự đổi uniqueness hoa/thường của dữ liệu cũ.

## 2. user_roles

| Cột | Kiểu / NULL | Nghĩa |
|---|---|---|
| user_id | INT NN FK users | Tài khoản |
| role_code | TEXT NN | ADMIN / HR / DIRECTOR |

PK(user_id, role_code). Không thêm EMPLOYEE/PM/PTCM vào bảng này: tư cách nhân viên từ hồ sơ hợp lệ, trách nhiệm dự án từ projects. Lịch sử cấp/thu hồi ở audit_events, không thêm bảng roles/permissions. Bảo vệ Admin sử dụng được cuối cùng bằng giao dịch chung, không chỉ COUNT trước UPDATE.

## 3. employees

| Cột | Kiểu / NULL / default | Nghĩa |
|---|---|---|
| employee_id | PK | Hồ sơ chính thức |
| user_id | INT NN FK users, UQ | Một hồ sơ/tài khoản |
| employee_code | VARCHAR(30) NN, UQ | Mã nhân viên |
| fullname | VARCHAR(100) NN | Họ tên hiện hành |
| email | VARCHAR(100) NN, UQ | Email công ty; HR sửa trực tiếp, không đồng nghĩa username |
| phone | VARCHAR(20) ? | Liên hệ, không đưa vào candidate DTO mặc định |
| address | TEXT ? / NULL | Địa chỉ phải bổ sung khi hoàn thiện hồ sơ; chưa có khi tạo được giữ NULL; HR quản lý, không DTO tra cứu chuyên môn |
| department_id | INT ? FK departments | Phòng ban |
| position_id | INT NN FK positions | Chức danh, không cấp quyền |
| job_level_id | INT ? FK job_levels | Cấp bậc do HR quản lý trong danh mục công ty; NULL khi chưa cập nhật, không suy điểm kỹ năng |
| employment_status | employment_status NN / ACTIVE | HR quản lý, khác khóa tài khoản |
| timestamps | Như quy ước | Server |

Nhân viên đề nghị sửa fullname/phone/address qua HR. Email công ty/cơ cấu do HR sửa trực tiếp; position_id và job_level_id khác nghĩa; tên cấp bậc do công ty đặt. Địa chỉ được bổ sung theo quyết định 20/10, không mở thêm DOB/CCCD/lương. Tạo user+employee nguyên tử, không bắt địa chỉ tại bước tạo. Đối chiếu giá trị nền theo trường khi sửa, không còn profile version. Index department_id/position_id/job_level_id theo truy vấn danh sách; không lưu workload/project_count trên employee.

`address` cho phép NULL vì hồ sơ ban đầu có thể chưa được bổ sung; không dùng NOT NULL hoặc default chuỗi giả. Nếu có giá trị, phải có nội dung sau trim; ràng buộc tĩnh đích là NULL hoặc chuỗi có nội dung. Service không cho chuyển địa chỉ đã có về NULL/rỗng, cả cập nhật HR lẫn duyệt yêu cầu. Không cần trigger/bảng lịch sử địa chỉ riêng; dùng audit hiện có. Không chặn nghiệp vụ dự án vì địa chỉ NULL; UI nhắc bổ sung, không thêm trạng thái hồ sơ. Không đưa address vào DTO ứng viên/tra cứu/snapshot gợi ý.

## 4. departments

department_id PK; department_name VARCHAR(100) NN UQ; timestamps. Một phòng ban, không cây tổ chức mới. Không xóa cứng khi còn employee tham chiếu. HR quản trị departments/positions/job_levels; không mặc định Admin có quyền HR.

## 5. positions

position_id PK; position_name VARCHAR(100) NN UQ; timestamps. Một chức danh, không role hệ thống. Không tự tạo chức danh từ tên actor.

## 5.1. job_levels — bổ sung 26/10/2026

Một cấp bậc trong danh mục công ty, không phải cấp độ kỹ năng hoặc vai trò. Giữ số mục cũ để các tham chiếu 6–21 không bị dịch.

| Cột | Kiểu / NULL / default | Nghĩa |
|---|---|---|
| job_level_id | PK | Định danh cấp bậc |
| job_level_name | VARCHAR(100) NN UQ | Tên do HR cấu hình, có nội dung sau trim; cùng quy tắc kiểm tra tên trùng của danh mục tổ chức |
| timestamps | Như quy ước | Server |

HR được thêm/sửa tên; chỉ xóa khi không còn employee tham chiếu. FK employees.job_level_id nullable, NO ACTION; không cascade/set NULL khi xóa. Không thêm thứ tự, điểm, trạng thái ACTIVE/INACTIVE, công ty/multi-tenant hoặc liên kết vị trí/phòng ban. Không hardcode danh sách Intern/Junior/Senior cho mọi công ty. Chọn ID khi HR sửa hồ sơ (Admin chỉ chọn mục có sẵn khi tạo ban đầu theo quyền hiện có); nhân viên không đổi qua profile request. Snapshot/audit lưu tên lúc quyết định để đổi tên danh mục không viết lại lịch sử.

## 6. skill_categories

category_id PK; category_name VARCHAR(100) NN UQ; timestamps. Giám đốc quản trị skill_categories/skills. Nhóm tra cứu kỹ năng, không cấp độ và không trọng số đánh giá.

## 7. skills

skill_id PK; skill_name VARCHAR(100) NN UQ; category_id INT NN FK skill_categories; timestamps. Index category_id. Không có trạng thái danh mục; FK NO ACTION chặn xóa skill còn được tham chiếu. Sửa tên không được đổi nghĩa kỹ năng đang dùng trong lịch sử. Không thêm hạn hiệu lực kỹ năng.

## 8. employee_skills

| Cột | Kiểu / NULL / default | Nghĩa |
|---|---|---|
| employee_skill_id | PK | Dòng năng lực hiện đang được công nhận |
| employee_id, skill_id | INT NN FK employees/skills | UQ(employee_id, skill_id) |
| proficiency_level | proficiency_level NN | BASIC/INTERMEDIATE/ADVANCED/EXPERT đã công nhận |
| years_experience | NUMERIC(4,1) ? | ≥0 nếu có; thiếu không coi 0 |
| note | TEXT ? | Ghi nhận chính thức |
| approved_by | INT NN FK users | Actor quyết định gần nhất |
| approved_at | TIMESTAMPTZ NN | Thời điểm quyết định |
| timestamps | Như quy ước | Server |

Chỉ quyết định hợp lệ ghi bảng này; không tạo dòng pending giữ chỗ. REMOVE xóa dòng hiện tại sau quyết định, giữ request/căn cứ/snapshot. Công nhận lại tạo dòng mới qua yêu cầu mới; không cascade lịch sử. Nguồn MANUAL/CV chỉ giữ ở request. Actor khác tên cột legacy approved_by_employee_id: migration phải JOIN employees.user_id, không copy số ID trực tiếp. Giám đốc không tự duyệt. Index skill_id phục vụ tìm ứng viên. Lịch sử ở requests/audit, hai cột approval không thay toàn bộ lịch sử.

## 9. employee_skill_change_requests

| Cột | Kiểu / NULL / default | Nghĩa |
|---|---|---|
| request_id | PK | Một đề nghị cho một kỹ năng |
| employee_id, skill_id | INT NN FK employees/skills | Subject |
| operation | ADD/UPDATE/REMOVE NN | Nhân viên chỉ tạo ADD/UPDATE theo chốt 27/10; REMOVE giữ tương thích thiết kế/lịch sử, không tạo mới từ luồng này |
| source | MANUAL/CV NN | Đường khởi tạo |
| before_values | JSONB ? | Bản official đối chiếu; NULL nếu chưa có |
| proposed_values | JSONB ? | Level/years/note; REMOVE không cần đề xuất mức mới |
| explanation | TEXT NN | Giải trình bổ sung/điều chỉnh; lý do gỡ chỉ thuộc bản ghi cũ |
| external_evidence | TEXT ? | Căn cứ/đường dẫn ngoài, không quản lý file giả |
| expert_opinion | TEXT ? | Nguồn và nội dung ý kiến ngoài; không giả chữ ký |
| extracted_skill_id | INT ? FK cv_extracted_skills | Nguồn CV gốc, UQ khi khác NULL |
| evidence_cv_id | INT ? FK employee_cvs | Tối đa một CV dẫn trực tiếp, đúng chủ; không chép lặp CV nguồn |
| evidence_attachments | JSONB NN / [] | Metadata nhiều tệp minh chứng ngoài CV, server xác thực |
| requested_by | INT NN FK users | Người thực sự tạo, không giả nhân viên |
| requested_at | TIMESTAMPTZ NN / now() | Server |
| status | TEXT NN / PENDING | PENDING/APPROVED/REJECTED/WITHDRAWN/NO_CHANGE theo workflow |
| resolved_by | INT ? FK users | Actor kết thúc |
| resolved_at, decision_reason | TIMESTAMPTZ ?, TEXT ? | Quyết định/căn cứ kết thúc |
| version | INT NN / 1 | Phiên bản nội dung/căn cứ đề nghị; tăng khi bổ sung PENDING |

Partial UQ(employee_id, skill_id) WHERE status=PENDING, chung MANUAL/CV. Source CV bắt buộc extracted_skill_id; MANUAL không dùng pointer nguồn này. Kiểm tra chuỗi CV/employee và mapped skill khớp request. ADD/UPDATE cần proposed hợp lệ; REMOVE giữ before, không ép có level đề nghị. PENDING chưa có quyết định; terminal đủ actor/time/reason và bất biến. Khi kết thúc NO_CHANGE lưu before_values thực tế đã đối chiếu theo quy tắc vòng đời, không ghi official giả. Bổ sung căn cứ tăng request.version và audit; phê duyệt kiểm tra request.version người duyệt đã xem và đối chiếu before_values với official hiện tại (kể cả sự vắng mặt khi ADD) dưới khóa employee/request.

JSON before/proposed là hợp đồng giá trị official (schema_version, proficiency_level, years_experience, note), không nhận role/actor/version mới từ client. evidence_attachments là mảng metadata storage_key/original_file_name/mime_type/file_size_bytes; file thật trên cloud cấu hình chung, không binary/base64/URL công khai. Server kiểm tra quyền/schema/type/size/key, không tin key client tùy ý. Bổ sung khi PENDING tăng request.version; terminal giữ nguyên căn cứ. Loại/dung lượng/số tệp và cleanup orphan còn cần đặc tả, không thêm bảng tệp. CV nguồn suy từ extracted_skill_id, evidence_cv_id dành CV dẫn riêng khác mục đích.

## 10. employee_profile_change_requests

request_id PK; employee_id INT NN FK employees; before_values/proposed_values JSONB NN; explanation TEXT ?; requested_by INT NN FK users; requested_at TIMESTAMPTZ NN default now(); status PENDING/APPROVED/REJECTED NN default PENDING; resolved_by INT ? FK users; resolved_at TIMESTAMPTZ ?; decision_reason TEXT ?.

JSON chỉ chứa schema_version và các trường fullname/phone/address được đề nghị thay đổi. Với address: bỏ qua khóa nghĩa là không đổi; before có thể NULL, proposed phải là chuỗi có nội dung sau trim, không nhận NULL/rỗng để xóa. Không bắt mọi patch fullname/phone phải kèm address nếu chưa bổ sung; việc thiếu địa chỉ chỉ được nhắc trên UI. Email công ty/cơ cấu do HR sửa trực tiếp. Không nhận toàn bộ employee object. Không role, department/position, official skills hoặc trách nhiệm dự án. PENDING thiếu bộ quyết định; terminal có người/thời điểm, REJECTED có lý do. Approve đối chiếu before_values đúng các trường patch dưới khóa employee/request rồi ghi patch + quyết định nguyên tử; stale không ghi đè. APPROVED có thể không decision_reason. Không UQ pending/employee, không tự thêm WITHDRAWN hoặc luồng bổ sung như skill request. Index(employee_id, requested_at), hàng đợi(status, requested_at).

## 11. skill_request_task_evidence

evidence_id PK; request_id INT NN FK employee_skill_change_requests; assignment_id INT NN FK assignments; submission_id INT ? FK task_submissions; contribution_note TEXT ?.

Một dòng là một đóng góp dẫn chứng, không một dự án. Task/employee suy từ assignment. Nếu có submission, kiểm tra cùng assignment; dùng UQ(submission_id, assignment_id) ở task_submissions và FK ghép tương ứng. Subject request phải khớp employee của assignment bằng service dưới khóa request; FK đơn không bảo vệ điều này.

Chống lặp bằng hai unique có điều kiện: (request_id, assignment_id) khi submission_id NULL; (request_id, assignment_id, submission_id) khi submission_id khác NULL. Cho 0..n dòng, không bắt người mới có task. Sửa căn cứ chỉ khi request PENDING, tăng request.version/audit; terminal đóng băng. Giám đốc chỉ đọc DTO căn cứ, không được toàn project vì có liên kết.

## 12. employee_cvs

cv_id PK; employee_id INT NN FK employees; storage_key VARCHAR(255) NN; original_file_name VARCHAR(255) NN; mime_type VARCHAR(100) NN; file_size_bytes BIGINT NN >0 và ≤10485760; uploaded_by INT NN FK users; uploaded_at TIMESTAMPTZ NN default now(); is_current BOOLEAN NN default false.

UQ(storage_key), partial UQ(employee_id) WHERE is_current. Không file_url public, không trạng thái phê duyệt CV. Đổi bản hiện tại khóa employee, đổi hai cờ cùng transaction/audit; parser thành công không là điều kiện. HR upload không tự chọn; quyền chọn thay D06 còn mở. Lưu một cloud provider cấu hình chung, chưa chọn nhà cung cấp; không mặc định local disk. File phải kiểm tra nội dung thực tế, không chỉ extension/mime client. Không xóa tệp đang làm căn cứ; retention D14/D15 chưa chốt.

## 13. cv_extractions

extraction_id PK; cv_id INT NN FK employee_cvs UQ; processing_status TEXT NN default QUEUED (QUEUED/PROCESSING/SUCCEEDED/FAILED là đề xuất kỹ thuật); parser_version VARCHAR(50) NN; extracted_text TEXT ?; attempt_count INT NN default 0 CHECK ≥0; lease_token UUID ?; lease_expires_at TIMESTAMPTZ ?; started_at/finished_at TIMESTAMPTZ ?; error_code TEXT ?; timestamps.

API nhận diện tên kỹ năng/đoạn căn cứ; chưa chọn provider/model. Tạm giữ các cột text/retry/lease hiện có, chưa khẳng định phương án worker cuối. Một tác vụ/CV, retry kỹ thuật cùng ID, không thêm đợt xét duyệt. PROCESSING có lease/token; worker commit kiểm tra token còn sở hữu. Parse ngoài transaction; ghi items+SUCCEEDED nguyên tử. SUCCEEDED rỗng hợp lệ. Không lưu extracted_markdown trùng text nếu chưa có nhu cầu riêng. Trạng thái xét kỹ năng suy từ items/requests, không reviewed_by cấp extraction như phê duyệt tệp. Chạy lại sau thành công chưa thuộc retry đã đặc tả.

## 14. cv_extracted_skills

extracted_skill_id PK; extraction_id INT NN FK cv_extractions; source_ordinal INT NN >0; raw_skill_name VARCHAR(100) NN; skill_id INT ? FK skills; evidence_text TEXT ?; row_outcome TEXT ?; resolved_by INT ? FK users; resolved_at TIMESTAMPTZ ?; decision_reason TEXT ?.

UQ(extraction_id, source_ordinal); partial UQ(extraction_id, skill_id) khi đã map. Ordinal của mục nguồn sau hợp nhất phải ổn định khi retry. Không thêm confidence_score nếu parser chưa có cách đo được định nghĩa.

Nếu có request gốc: quyết định đọc từ request, bộ row_outcome/actor/time/reason đều NULL. Nếu không có request: có thể terminal NO_CHANGE/REJECTED với đủ bộ quyết định; NULL nghĩa chưa kết thúc. Bảo vệ loại trừ này qua giao dịch khóa item (chéo bảng, không chỉ CHECK). APPROVED phải đi qua request ghi official. Mapping muộn vào mục đã terminal vẫn D08: không merge mất căn cứ hoặc mở lại quyết định. Chưa có giải pháp cuối cho trường hợp đó, không tuyên bố bảng đã sẵn sàng DDL.

## 15. task_comments

comment_id PK; task_id INT NN FK tasks; employee_id INT NN FK employees; content TEXT NN; timestamps. Một trao đổi trên task, không quyết định nghiệm thu, không file/chat thread mới. Index(task_id, created_at, comment_id). Actor thực tế resolve employee.user_id; không cho viết thay bằng cách gửi employee_id. Quyền đọc/ghi theo task hiện hành; danh sách từng được giao không tự cấp quyền vĩnh viễn. Chưa thêm sửa/xóa mềm hoặc moderation khi nghiệp vụ chưa yêu cầu.

## 16. recommendation_runs

run_id PK; task_id INT NN FK tasks; generated_by INT NN FK users; snapshot_at TIMESTAMPTZ NN; generated_at TIMESTAMPTZ NN default now(); engine_version VARCHAR(50) NN; schema_version INT NN >0; params JSONB NN; input_snapshot JSONB NN; result_metadata JSONB NN.

UQ(run_id, task_id) phục vụ assignment FK; index(task_id, generated_at). Immutable. params là cấu hình chuẩn của run; input_snapshot chỉ dữ liệu đầu vào, không lặp params/engine_version thành nguồn thứ hai. schema_version trên run định nghĩa ba JSON của run (params/input_snapshot/result_metadata) và score_details của các item thuộc run. Metadata: số xét/loại, top_n, lý do rỗng nếu có; count không âm và phân loại có contract rõ. Không lưu output_snapshot lặp items, selected_employee_id hoặc history_revision/R05.

Input giữ đủ dữ liệu thực dùng cho replay: đặc tả, giá trị official skills, gồm sự vắng mặt của kỹ năng liên quan, tải và dữ liệu lịch sử liên quan theo công thức được chốt; không mặc định thêm Q/T hoặc số dự án vào điểm. Số dự án đang có việc chỉ thông tin tham khảo. D12/D16 còn mở cho contract công thức; không bịa danh sách thành phần đã được duyệt. Snapshot đọc nhất quán, không có nghĩa dữ liệu còn mới lúc confirm. Lỗi engine khác run rỗng hợp lệ.

## 17. recommendation_items

item_id PK; run_id INT NN FK recommendation_runs; employee_id INT NN FK employees; rank_no INT NN >0; matching_score NUMERIC(5,2) NN trong [0,100]; explanation TEXT NN; score_details JSONB NN.

UQ(run_id, employee_id), UQ(run_id, rank_no), UQ(item_id, run_id, employee_id). Cả tập rank liên tục, không quá top_n do service kiểm tra. Assignment ghép đúng item/run/người; không is_selected, không chép matching_score sang assignment.

**Đã chốt thay các cột điểm thành phần legacy bằng score_details theo run.schema_version** để không đóng băng skill_score/workload_score/perf_score trước chốt công thức. Tổng matching_score và rank vẫn cột quan hệ để kiểm tra/hiển thị. Không đồng thời giữ cột thành phần và JSON thành hai nguồn chuẩn. Trade-off: giảm thay DDL khi thay công thức nhưng validator/replay phải kiểm tra cấu trúc, miền giá trị và tính khớp tổng điểm; JSON không phải chỗ bỏ dữ liệu tùy ý. D12 phải chốt contract trước DDL phần này. Đã chốt cấu trúc lưu trữ, chưa chốt công thức. run.schema_version là định dạng chung của params/input_snapshot/result_metadata/score_details, không thêm schema_version cạnh tranh trong items.

## 18. auth_sessions

session_id PK; user_id INT NN FK users; token_hash BYTEA NN UQ (SHA-256, 32 byte); created_at TIMESTAMPTZ NN default now(); expires_at TIMESTAMPTZ NN; revoked_at TIMESTAMPTZ ?.

Cookie chứa secret ngẫu nhiên đủ entropy, DB chỉ lưu hash; session_id là ID nội bộ, không dùng làm cookie. Mỗi user nhiều phiên. CHECK expires_at > created_at, revoked_at >= created_at nếu có, octet_length(token_hash)=32; index(user_id). Kiểm tra phiên chưa thu hồi/còn hạn và quyền/status hiện hành mỗi request. Logout thu hồi phiên; khóa/reset/đổi role thu hồi các phiên trong cùng giao dịch.

Không JWT/refresh endpoint hoặc refresh_tokens. Timeout tuyệt đối cấu hình khi triển khai, chưa chốt số ngày; idle timeout chưa cam kết vì chưa có last_seen_at. Không âm thầm thêm cột/đẩy expires_at liên tục để giả có cả idle và hạn tuyệt đối. Cookie/CSRF theo security. Chưa chọn/cài thư viện session.

## 19. idempotency_records

idempotency_id PK; actor_id INT NN FK users; operation VARCHAR(100) NN; idempotency_key VARCHAR(128) NN; request_hash BYTEA NN (32 byte); result_reference JSONB NN; created_at TIMESTAMPTZ NN default now(); expires_at TIMESTAMPTZ ?.

Chỉ mutation dễ tạo trùng, đặc biệt tạo/bổ sung nhật ký; không mọi API/login. UQ(actor_id, operation, idempotency_key). JSON schema_version/type/IDs/result_code tối thiểu, service kiểm tra liên kết; không DTO nhạy cảm/token. Hash canonical do server. Placeholder và kết quả commit cùng transaction nghiệp vụ; không dòng IN_PROGRESS commit riêng. Retry kiểm tra quyền hiện hành rồi dựng DTO cũ, không làm lại mutation. expires_at NULL nghĩa chưa lập lịch xóa, không hứa bảo đảm vô hạn; cửa sổ retry/cleanup D14 còn chờ, chưa TTL tự chọn. Hết hạn không tự xóa và thực hiện lại mà chưa có chính sách.

## 20. audit_events

audit_id PK; actor_id INT NN FK users; entity_type TEXT NN; entity_id INT NN; operation TEXT NN; occurred_at TIMESTAMPTZ NN default now(); reason TEXT ?; before_values/after_values JSONB ?.

Chỉ thao tác con người; actor_id bắt buộc. Log worker/lỗi vận hành không thuộc bảng này. Append-only, entity_type/operation allowlist; entity_id đa loại không có FK tổng quát, service kiểm tra. JSON allowlist/schema_version, không credential/token/CV thô. Chỉ bổ sung lịch sử chưa đủ ở bảng miền: đổi role/trách nhiệm/profile trực tiếp/spec/chọn CV. Không sao chép mọi submission/history. Index(entity_type, entity_id, occurred_at).

## 21. notifications

notification_id PK; recipient_id INT NN FK users; notification_type TEXT NN; resource_type TEXT NN; resource_id INT NN; message TEXT NN; created_at TIMESTAMPTZ NN default now(); read_at TIMESTAMPTZ ?.

Yêu cầu hiệu chỉnh gửi PTCM hiện hành, không PM từ trách nhiệm PM. Ghi cùng giao dịch quyết định; chỉ in-app, không outbox/Slack. Tham chiếu đa loại kiểm tra service, đọc lại quyền khi mở. Message tối thiểu không chứa hồ sơ nhạy cảm. Index(recipient_id, created_at, notification_id); read_at ≥ created_at nếu có. Đọc không thay trạng thái yêu cầu.

## Kiểm kê và giới hạn

Nhóm nền tảng 9: users, user_roles, employees, departments, positions, job_levels, skill_categories, skills, employee_skills.
Nhóm yêu cầu/căn cứ 3: employee_skill_change_requests, employee_profile_change_requests, skill_request_task_evidence.
Nhóm CV 3: employee_cvs, cv_extractions, cv_extracted_skills.
Còn lại: task_comments (1), recommendation_runs/items (2), auth_sessions/idempotency_records/audit_events/notifications (4).

Tổng **22** bảng, cộng **11** bảng ở 5A thành **33**. Đã bỏ assignment_events và refresh_tokens; metadata upload không tăng bảng. Các mục quy ước/sơ đồ không phải bảng.

## Cổng hoàn thiện bản đề xuất cuối

1. Hợp nhất ERD 33 bảng, chỉ rõ FK khai báo và liên kết kiểm tra ở service; không vẽ quan hệ suy diễn như FK thật.
2. Đối chiếu hiện trạng→đích: đổi actor employee→user bằng JOIN, bỏ membership sau kiểm kê, giữ dữ liệu legacy và không sửa migration đã áp dụng.
3. D05 allowlist/C06 chủ danh mục đã chốt; còn DTO chi tiết, mapping muộn, contract công thức, auth/retention và chính sách tệp. Các điểm này không cần biến thành bảng mới nhưng không được coi đã giải quyết.
4. Review CHECK xử lý NULL, unique có điều kiện, FK ghép và giao dịch cạnh tranh trước SQL. Dictionary không thay kiểm thử thực tế.

Chưa đổi Q/T, công thức phân công, quyền đã chốt hoặc tạo ngoại lệ nghiệm thu. Không coi dữ liệu giả lập là bằng chứng vận hành. Đề xuất JSON điểm thành phần chỉ là lựa chọn lưu trữ có trade-off, phải đồng bộ API khi công thức được xác nhận.
