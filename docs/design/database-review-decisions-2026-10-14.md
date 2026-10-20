# Ghi nhận quyết định rà soát database theo từng nhóm

Ngày bắt đầu: 14/10/2026.

**Quyết định tiếp nối 20/10/2026 — ưu tiên về địa chỉ:** theo [baseline 20/10](../business-baseline-2026-10-20.md), thêm `employees.address TEXT NULL`. Địa chỉ phải bổ sung khi hoàn thiện hồ sơ nhưng không bắt Admin nhập khi tạo; NULL biểu thị chưa bổ sung. Nhân viên đề nghị fullname/phone/address, HR duyệt hoặc sửa trực tiếp; giá trị địa chỉ được cung cấp phải có nội dung, không xóa trắng địa chỉ đã có. Không tra cứu chuyên môn, không chặn đăng nhập/công việc vì thiếu, không thêm trạng thái/bảng. Dictionary 5B/DBML đã đồng bộ thiết kế, tổng vẫn 32 bảng, chưa migration. Các ghi chép 15/10 chỉ fullname/phone được giữ để truy vết và bị quyết định này thay đúng phần địa chỉ.

**Bản bàn giao đã rà lại: 15/10/2026.** Người dùng đã đồng ý phương án upload minh chứng bằng một trường metadata JSON, không thêm bảng. Đọc các phần hiện hành và hướng dẫn bàn giao cuối file; phần trong `<details>` chỉ là lịch sử, không phải yêu cầu triển khai.

## Mục đích và cách sử dụng

Ghi nhận các thay đổi và nội dung giữ nguyên trong quá trình cùng người dùng phân tích database. Hoàn tất trao đổi các nhóm rồi mới cập nhật đồng bộ tài liệu nghiệp vụ, thiết kế và DBML.

File này là sổ quyết định, không phải dictionary hay migration. Đợt đồng bộ tài liệu ngày 15/10 đã cập nhật dictionary 11+21 bảng, DBML và workflow theo quyết định hiện hành, gồm result_log_id được người dùng đồng ý. Xem [review đồng bộ](../reviews/database-sync-review-2026-10-15.md) về kiểm tra và giới hạn; chưa thay DB/mã nguồn. Các câu “chưa cập nhật” trong ghi chép trước đợt đồng bộ phản ánh trạng thái lúc trao đổi, không phải trạng thái bàn giao mới.

Trạng thái ghi nhận:

- **Đã thống nhất:** người dùng đã đồng ý hoặc trực tiếp yêu cầu lựa chọn đó.
- **Đề xuất chờ chốt:** phương án được đưa ra để giải thích hoặc cân nhắc, chưa coi là quyết định của người dùng.
- **Chưa rà soát:** tiếp tục phân tích ở nhóm tương ứng; giữ nguyên thiết kế hiện tại trong lúc chờ.

Nguồn đối chiếu:

- [Bản đề xuất tổng hợp](database-proposal-2026-10-14.md).
- [Dictionary 5A — 11 bảng](database-step5a-core-dictionary-2026-10-14.md).
- [Dictionary 5B — 21 bảng](database-step5b-supporting-dictionary-2026-10-14.md).
- [Sơ đồ DBML](database-proposal-2026-10-14.dbml).

## Tiến độ

| Nhóm | Số bảng hiện tại | Trạng thái |
|---|---:|---|
| 1. Tài khoản, nhân viên và cơ cấu công tác | 5 | Đã chốt cấu trúc và chức danh/cấp bậc; HR quản trị cơ cấu |
| 2. Danh mục kỹ năng, năng lực chính thức và yêu cầu thay đổi | 6 | Đã chốt source/allowlist, CV căn cứ và upload metadata JSON |
| 3. CV và dữ liệu trích xuất | 3 | Đã chốt hướng giản lược; chi tiết vận hành chờ thiết kế triển khai |
| 4. Dự án, công việc, phân công và nghiệm thu | 9 | Đã chốt; PTCM đặt weight theo task, miền giá trị theo công thức |
| 5. Nhật ký công việc và hiệu chỉnh | 3 | Đã chốt hướng giản lược và xóa mềm |
| 6. Gợi ý phân công | 2 | Đã đồng ý cấu trúc và rà nhất quán; contract công thức chờ chốt |
| 7. Phiên đăng nhập, chống ghi lặp, lịch sử kiểm toán và thông báo | 4 | Đã chốt session phía server và phạm vi các bảng kỹ thuật |

Tổng đích dự kiến hiện tại: **32 bảng** = 5 + 6 + 3 + 9 + 3 + 2 + 4. Giảm từ 34 do đã chốt bỏ `assignment_events` và `refresh_tokens`. Các số 33/34 trong phần lịch sử phản ánh mốc trước các quyết định này.

## Tóm tắt quyết định hiện hành đến 15/10/2026

- Nhóm 1: bỏ `users.must_change_password`, `users.version`, `employees.version`; giữ `position_id` bắt buộc và `seniority_level` nullable với nghĩa khác nhau, theo mục 1 hiện hành bên dưới. HR quản trị departments/positions; quyền nhân viên từ hồ sơ, PM/PTCM theo dự án, user_roles giữ ADMIN/HR/DIRECTOR.
- Nhóm 2: giữ skill_categories, bỏ skills.status; employee_skills chỉ chứa năng lực hiện được công nhận, bỏ source/is_effective/version. REMOVE được duyệt gỡ dòng hiện tại, giữ lịch sử; công nhận lại qua yêu cầu mới. Yêu cầu kỹ năng bỏ base_version, giữ before_values và request.version; nguồn MANUAL/CV vẫn ở request. Profile request bỏ base_version, nhân viên đề nghị sửa fullname/phone, email công ty/cơ cấu do HR sửa trực tiếp. Giữ nhiều minh chứng task, tối đa một CV dẫn trực tiếp và upload tùy chọn qua evidence_attachments JSONB, không thêm bảng. Lý do quyết định hồ sơ có thể trống khi chờ/chấp thuận, từ chối cần lý do.
- Nhóm 3: chốt hướng lưu cloud một nhà cung cấp cấu hình chung, bỏ storage_provider; API nhận diện tên kỹ năng/đoạn căn cứ; bỏ suggested_level/suggested_years_experience. Chưa chọn Cloudinary/Gemini cụ thể và chưa chốt toàn bộ trường worker.
- Nhóm 4: chọn PM ngay khi tạo; bỏ projects.version, tasks.version/spec_version; bỏ deliverable riêng, gộp đầu ra vào description, giữ acceptance_criteria. Bỏ assignment_events, bổ sung ended_at/ended_by/end_reason vào assignments. Giữ project_status_history, task_status_history, task_comments và log_snapshot. PTCM đặt weight theo task trước giao; miền giá trị/chuẩn hóa theo công thức, không tự giữ thang 1..5 làm quyết định cuối.
- Nhóm 5: đã chốt ba bảng theo bản hiện hành ở đầu mục 5. Người thực hiện tự ghi, không nhập thay; PTCM duyệt hiệu chỉnh sau khóa. Bỏ version/base_version và dữ liệu lặp; history lưu trước/sau. Xóa mềm bằng deleted_at thay voided_at; history dùng UPDATE/DELETE, yêu cầu dùng ADD/UPDATE/DELETE. Không mở khôi phục hoặc xóa vật lý.
- Nhóm 6: giữ hai bảng run/items; run còn 10 trường sau bỏ spec_version, items 7 trường. Đã review quan hệ/phạm vi dữ liệu và các phụ thuộc version từ nhóm trước; cấu trúc JSON theo công thức còn cần đặc tả.
- Nhóm 7: chuyển sang session phía server/cookie secret; bỏ JWT access/refresh endpoint và refresh_tokens, thêm token_hash vào auth_sessions. Giữ idempotency_records phạm vi hẹp, audit_events chỉ lịch sử người dùng còn thiếu (bỏ actor_kind), giữ notifications. Đã đồng bộ dictionary, DBML và workflow ngày 15/10; chưa thay mã nguồn hoặc database thực tế.

Các ghi nhận ban đầu dưới đây được giữ để truy vết trao đổi; quyết định chốt tiếp nối thay thế đề xuất cũ tương ứng. Không dùng một câu "chờ chốt" cũ để phủ định quyết định đã chốt phía sau.

## Chốt bổ sung C01–C06 và upload minh chứng

**Đã thống nhất:** C01 giữ chức danh/cấp bậc riêng; C02 bỏ employee_skills.source; C03 bỏ profile request.base_version và chốt fullname/phone qua yêu cầu, email công ty/cơ cấu do HR sửa; C04 PTCM đặt weight theo task; C05 giữ nhiều task, tối đa một CV dẫn trực tiếp và cho upload minh chứng; C06 HR quản trị departments/positions, Giám đốc quản trị skill_categories/skills.

**Upload đã chốt:** thêm `employee_skill_change_requests.evidence_attachments JSONB NOT NULL DEFAULT '[]'::jsonb`, lưu danh sách metadata gồm `storage_key`, `original_file_name`, `mime_type`, `file_size_bytes`. Đây là một trường chứa được nhiều tệp. Không thêm `skill_request_attachments`, không tăng tổng 32 bảng.

- Tệp thật lưu trên cloud cấu hình chung; không lưu binary/base64 hoặc URL công khai làm định danh chuẩn trong DB. Nhà cung cấp cụ thể chưa chọn.
- Tệp thuộc riêng yêu cầu, không có quy trình duyệt/vòng đời dùng chung. Cho bổ sung khi PENDING, tăng request.version; sau quyết định giữ nguyên căn cứ. Upload không tự công nhận kỹ năng.
- Server xác thực upload và tạo metadata; kiểm tra schema JSON, loại/kích thước/quyền tệp. Không tin storage_key tùy ý từ client, không dùng JSON thay FK task/CV/user.
- Phân biệt external_evidence (giải trình/đường dẫn), skill_request_task_evidence (FK đóng góp), evidence_cv_id/extracted_skill_id (CV căn cứ/nguồn) và evidence_attachments (tệp ngoài CV). Không đưa chứng chỉ vào employee_cvs.
- Loại tệp, dung lượng/số tệp, cơ chế truy cập private và dọn tệp upload chưa gắn yêu cầu là chi tiết triển khai còn cần đặc tả; không tự mở upload ảnh CV/OCR hoặc phân hệ quản lý tài liệu.

<details>
<summary>Lịch sử chốt C01–C06/upload — các đề xuất chờ ở dưới đã được phần hiện hành thay thế</summary>

### Các điểm còn lại sau khi rà đủ 7 nhóm (mốc trước xác nhận)

Đã kiểm tra lại ngày 15/10/2026. Sổ quyết định đã ghi đủ những thay đổi được đồng ý; chưa phải dictionary 32 bảng hoàn chỉnh. Các phương án sau chỉ là đề xuất để chốt nốt, không tự nâng thành quyết định người dùng.

| Mã | Cần xác nhận | Phương án đề xuất |
|---|---|---|
| C01 | Chức danh/cấp bậc | Giữ position_id và seniority_level nullable với hai nghĩa riêng (Developer + Senior); không gộp cấp bậc vào position_name, không tự dùng làm điểm kỹ năng. |
| C02 | Nguồn năng lực chính thức | Bỏ employee_skills.source; nguồn MANUAL/CV vẫn lưu ở yêu cầu đã duyệt. Không bỏ source ở bảng yêu cầu. |
| C03 | Yêu cầu thông tin cá nhân | Bỏ employee_profile_change_requests.base_version để khớp employees.version đã bỏ; đối chiếu before_values của các trường đề nghị sửa. Đề xuất nhân viên gửi sửa fullname/phone; email công ty và thông tin tổ chức do HR cập nhật trực tiếp, cần xác nhận email là email công ty trước khi áp dụng. |
| C04 | Weight của kỹ năng task | PTCM xác định theo từng task trước khi giao, không trọng số cố định cho kỹ năng toàn công ty. Miền giá trị và chuẩn hóa chốt cùng công thức, khác trọng số nhóm tiêu chí công ty. |
| C05 | CV/tệp dẫn chứng yêu cầu kỹ năng | Giữ nhiều task nội bộ và căn cứ ngoài dạng giải trình/đường dẫn đã có; phiên bản này tối đa một CV đã lưu làm căn cứ qua evidence_cv_id, chưa thêm upload tệp minh chứng riêng. CV nguồn đã suy từ extracted_skill_id không chép lặp cùng mục đích. Cardinality một CV và giới hạn upload cần được đồng ý. |
| C06 | Người quản trị danh mục | Đề xuất HR quản trị departments/positions, Giám đốc quản trị skill_categories/skills; không mặc nhiên cấp cho Admin hoặc PTCM. Đây là phân quyền đề xuất, chưa có quyết định trước đó. |

Các điểm đã có hướng nghiệp vụ, không cần mở lại chỉ để tăng quyền: nhân viên chọn CV hiện tại; HR upload không tự chọn thay. Quyền đọc tra cứu vẫn theo bốn nhóm nghề nghiệp đã xác định và không CV gốc/contact/raw task khác dự án; chi tiết DTO sẽ do thiết kế thực hiện trong phạm vi đó.

Chi tiết triển khai còn cần đặc tả trước code, không nhất thiết là câu hỏi nghiệp vụ để người dùng duyệt từng cột:

- Chọn nhà cung cấp cloud/API trích xuất, dữ liệu model/prompt/worker, contract kết quả trích xuất và xử lý mapping trong giới hạn không mở lại quyết định.
- Công thức cuối và schema các JSON gợi ý; cách xác định dữ liệu effort được phép dùng trong đánh giá. Đồng ý cấu trúc bảng không chứng minh bộ tiêu chí/công thức đã chốt.
- Chính sách hết hạn phiên/không hoạt động, reset mật khẩu, cookie/CSRF; phạm vi endpoint chống ghi lặp, cửa sổ retry và dọn dữ liệu. Không khôi phục must_change_password/JWT/refresh đã bỏ bằng mô tả cũ.
- CHECK/FK/index, đối chiếu giá trị nền dưới khóa, snapshot contract sau bỏ version, lịch sử trạng thái và mốc kết thúc lượt giao; xử lý an toàn hủy đang review không giả thành nghiệm thu không đạt. Không thêm bảng/UC hoặc tự chọn ảnh hưởng Q/T cho ngoại lệ đã giới hạn.
- Mapping dữ liệu cũ và kiểm thử DDL/giao dịch trên DB thử nghiệm. Không suy dữ liệu lịch sử chưa có từ dữ liệu hiện tại.

Sau C01–C06, có thể hợp nhất tài liệu/schema theo các quyết định, đồng thời ghi rõ các hợp đồng kỹ thuật/công thức còn đợi. Không gọi sổ trao đổi này là DB đã triển khai hoặc mọi chính sách đã hoàn tất.

### Xác nhận C01–C06 và câu hỏi upload minh chứng

Người dùng đồng ý các đề xuất C01–C06, nhưng hỏi riêng việc không upload minh chứng. Ghi nhận:

- **Đã chốt C01:** giữ chức danh và cấp bậc riêng, seniority_level nullable, không tự suy mức kỹ năng.
- **Đã chốt C02:** bỏ employee_skills.source, giữ nguồn ở yêu cầu đã duyệt.
- **Đã chốt C03:** bỏ profile request.base_version, đối chiếu before_values; nhân viên đề nghị sửa fullname/phone. Email công ty/thông tin tổ chức do HR cập nhật trực tiếp; không tự mở quyền đổi email đăng nhập hoặc coi username là email. Nếu dữ liệu legacy email không phải email công ty thì phải phân loại khi chuyển đổi, không tự đổi nghĩa dữ liệu cũ.
- **Đã chốt C04:** PTCM đặt weight theo từng task trước giao; miền giá trị/cách chuẩn hóa theo công thức, không cố định trọng số từng kỹ năng toàn công ty.
- **Đã chốt C06:** HR quản trị departments/positions; Giám đốc quản trị skill_categories/skills. Quyền quản trị danh mục không tự mở quyền khác hoặc bỏ điều kiện bảo vệ liên kết/lịch sử.
- **C05:** giữ hướng nhiều task/giải trình/đường dẫn và tối đa một CV đã lưu được dẫn trực tiếp. Giới hạn "chưa upload minh chứng riêng" đang được hỏi lại, **chưa coi đã chốt loại trừ upload**.

Giải thích/đề xuất tiếp, chờ người dùng: thiết kế hiện chỉ có external_evidence TEXT và evidence_cv_id, chưa có cơ chế quản lý file chứng chỉ/ảnh/tài liệu ngoài CV. Đề xuất cho upload minh chứng tùy chọn trong yêu cầu kỹ năng, không bắt có file hoặc công việc nội bộ, không tự công nhận mức từ chứng chỉ. Giữ CV và tệp minh chứng khác mục đích; không đưa chứng chỉ vào employee_cvs.

Nếu chấp thuận upload, đề xuất xem một bảng liên kết trực tiếp request–tệp (`skill_request_attachments`) với metadata cloud tối thiểu thay vì hệ thống tài liệu/tệp chung nhiều tầng. Có thể nhiều tệp, sửa/bổ sung khi Pending và giữ căn cứ sau kết thúc; quyền xem theo yêu cầu. Chưa chốt bảng/cột, loại tệp/dung lượng hoặc tăng tổng 32 trước khi người dùng đồng ý phạm vi.

Các mục "còn mở" tương ứng ở lịch sử nhóm 1/2/4 được thay thế bởi xác nhận C01–C04/C06 ở đây. C05 upload chưa đóng; chưa sửa dictionary/DBML/workflow.

### Upload minh chứng — đã đồng ý nghiệp vụ, đề xuất một trường thay bảng riêng

Người dùng **đã chốt cho upload minh chứng** nhưng hỏi vì sao cần bảng riêng, ưu tiên một trường. Quyết định này thay giới hạn không upload trong C05. Các điểm đã thống nhất khác của C05 vẫn giữ.

Phương án lưu trữ được đề xuất tiếp: thêm `evidence_attachments JSONB` vào employee_skill_change_requests, default mảng rỗng, lưu danh sách metadata tệp (storage_key, original_file_name, mime_type, file_size_bytes). Tệp thực lưu cloud, không binary/base64 trong JSON; khóa lưu trữ không là URL công khai. Không chứa ID user/CV/task để thay FK; actor bổ sung tệp ghi qua cơ chế request.version/audit đã có. JSON phải kiểm tra cấu trúc và server tạo metadata từ kết quả upload hợp lệ, không tin key/tệp tùy ý từ client.

Lý do có thể không cần bảng: tệp chỉ là thành phần căn cứ thuộc một yêu cầu, không chia sẻ nhiều yêu cầu hoặc có vòng đời/duyệt riêng; bổ sung khi Pending, giữ nguyên sau kết thúc. Một trường có thể chứa nhiều mục metadata, không chỉ một tệp. Đề xuất skill_request_attachments trước đó không bắt buộc và chưa được chốt, nay không còn là phương án ưu tiên. Tổng đích vẫn **32 bảng**, chưa thêm bảng đính kèm.

Phân biệt: đường dẫn ngoài external_evidence TEXT, task evidence theo bảng liên kết có FK, CV theo evidence_cv_id/extracted_skill_id và tệp upload riêng là bốn loại căn cứ khác nhau. Không đưa chứng chỉ vào employee_cvs hoặc ID task vào JSON attachments.

Người dùng đã đồng ý upload; tên trường/JSON metadata là phương án thiết kế đang trình bày, chưa tự ghi rằng người dùng đã xác nhận chi tiết JSON. Loại tệp/dung lượng, quyền upload/xem và giữ/xóa file sẽ đặc tả cùng cơ chế cloud; không mở phân hệ quản lý file riêng.

</details>

## 1. Tài khoản, nhân viên và cơ cấu công tác — 5 bảng

### Quyết định hiện hành

- `users`: còn user_id, username, password_hash, status, created_at, updated_at; bỏ must_change_password/version. Username duy nhất; không tự đổi username thành email.
- `user_roles`: giữ user_id/role_code, PK ghép; ADMIN/HR/DIRECTOR. Nhân viên được xác định qua employees, PM/PTCM qua projects; không thêm bảng role/permission hoặc role EMPLOYEE chỉ để biểu thị hồ sơ.
- `employees`: giữ employee_id, user_id, employee_code, fullname, email, phone, department_id, position_id, seniority_level, employment_status, created_at, updated_at; bỏ version. Bổ sung address TEXT NULL theo quyết định 20/10 ở đầu file. User 1–0..1 employee, user_id và employee_code duy nhất. Department/position 1–n employees; position_id bắt buộc, department_id/seniority_level có thể trống theo thiết kế hiện tại.
- Giữ cả position_id (Developer/Tester…) và seniority_level (Junior/Middle/Senior…), không gộp cấp bậc vào tên chức danh, không suy mức từng kỹ năng hoặc tự cộng điểm phân công.
- `departments`: department_id, department_name, created_at, updated_at; `positions`: position_id, position_name, created_at, updated_at. HR quản trị hai danh mục; không thêm cây tổ chức hoặc lịch sử chuyển phòng riêng.
- Nhân viên đề nghị sửa fullname/phone/address qua profile request (address bổ sung 20/10); HR sửa email công ty/cơ cấu trực tiếp. Bỏ profile request.base_version, đối chiếu before_values của đúng trường đề nghị sửa trong giao dịch. Dữ liệu email legacy cần phân loại, không tự đổi nghĩa.
- Không lưu workload, project_count, KPI hoặc danh sách kỹ năng trong employees. Các chính sách reset mật khẩu/phiên cụ thể thuộc thiết kế triển khai, không khôi phục cột đã bỏ.

<details>
<summary>Lịch sử trao đổi nhóm 1 — các trạng thái chờ dưới đây đã được C01/C03/C06 thay thế</summary>

### 1.1. `users`

**Đã thống nhất:** bỏ `must_change_password` và `version` ở bảng `users` trong phiên bản hiện tại. Chưa cập nhật các tài liệu khác.

Danh sách trường sau thay đổi dự kiến:

| Trường | Mục đích |
|---|---|
| `user_id` | Định danh tài khoản |
| `username` | Tên đăng nhập duy nhất |
| `password_hash` | Giá trị băm mật khẩu |
| `status` | Trạng thái truy cập của tài khoản |
| `created_at` | Thời điểm tạo |
| `updated_at` | Thời điểm cập nhật gần nhất |

Việc bỏ `must_change_password` không tự quyết định toàn bộ chính sách cấp/reset mật khẩu. `updated_at` không thay lịch sử kiểm toán hoặc cơ chế bảo vệ cập nhật đồng thời.

### 1.2. `user_roles`

**Thiết kế hiện tại đã giải thích; chưa có yêu cầu thay đổi:** giữ hai trường `user_id`, `role_code`; khóa chính ghép `(user_id, role_code)`. Các mã vai trò hiện tại: `ADMIN`, `HR`, `DIRECTOR`.

- Nhân viên thông thường không cần dòng `EMPLOYEE` trong bảng này; tư cách nhân viên xác định từ hồ sơ `employees` hợp lệ liên kết với tài khoản.
- PM/PTCM là trách nhiệm theo từng dự án, xác định từ `projects`.
- Một tài khoản có thể có nhiều vai trò hệ thống nếu nghiệp vụ cho phép; không mặc nhiên cấp quyền kiêm nhiệm.
- Chưa đề xuất thêm các bảng `roles`, `permissions`, `role_permissions`.

Phần này cần được người dùng xác nhận khi chốt nhóm 1; việc đã giải thích không được ghi thành quyết định mới đã phê duyệt.

### 1.3. `employees`

**Đã thống nhất:** người dùng đồng ý bảng hồ sơ nhân viên về tổng thể và yêu cầu bỏ `employees.version`.

Các trường ngoài `version` giữ theo thiết kế hiện tại trong lúc xem xét riêng chức danh/cấp bậc: `employee_id`, `user_id`, `employee_code`, `fullname`, `email`, `phone`, `department_id`, `position_id`, `seniority_level`, `employment_status`, `created_at`, `updated_at`.

#### Ảnh hưởng tới yêu cầu cập nhật hồ sơ

**Đề xuất chờ chốt:** khi bỏ `employees.version`, bỏ tham chiếu `base_version` tương ứng trong `employee_profile_change_requests`; khi duyệt, đối chiếu giá trị hiện tại của các trường được đề nghị sửa với giá trị cũ trong `before_values`. Nếu đã khác, HR cần xem lại trước khi áp dụng. Kiểm tra và cập nhật phải thực hiện trong cùng giao dịch có bảo vệ cạnh tranh.

Ví dụ: yêu cầu sửa số điện thoại A → B đang chờ, HR đã sửa trực tiếp thành C. Không tự ghi đè C bằng yêu cầu cũ mà bỏ qua sự thay đổi.

Người dùng đã yêu cầu bỏ `employees.version`; cơ chế đối chiếu thay thế trên mới là đề xuất của trợ lý. Cần chốt khi rà nhóm 2, vì bảng yêu cầu cập nhật hồ sơ thuộc nhóm đó.

Không suy rộng việc bỏ version ở `users`/`employees` thành bỏ mọi trường version trong database.

#### Chức danh và cấp bậc

**Đề xuất chờ chốt:** giữ cả hai nếu sử dụng theo hai nghĩa riêng:

| Trường | Ý nghĩa | Ví dụ |
|---|---|---|
| `position_id` | Chức danh/chuyên môn công tác | Developer, Tester, Business Analyst |
| `seniority_level` | Cấp bậc nghề nghiệp; có thể trống | Junior, Middle, Senior |

Người dùng đã nêu trường hợp `Developer + Senior` để cân nhắc, chưa xác nhận cuối cùng giữ cả hai.

Nếu chọn hướng này, danh mục chức danh không gộp cấp bậc vào tên như `Senior Developer`. Cấp bậc nghề nghiệp không tự suy ra mức từng kỹ năng và chưa được quyết định đưa vào điểm phân công.

#### Những điểm còn cần làm rõ

- Ngữ nghĩa email và các trường nhân viên được đề nghị sửa.
- Giữ cả chức danh/cấp bậc hay chỉ chức danh.
- Cơ chế xử lý yêu cầu cập nhật hồ sơ sau khi bỏ version.

### 1.4. `departments`

**Đã thống nhất:** giữ bảng phòng ban như đã trình bày, gồm `department_id`, `department_name`, `created_at`, `updated_at`.

Danh mục dùng chung cho hồ sơ nhân viên. Không thêm cây tổ chức, trưởng phòng hoặc lịch sử chuyển phòng trong vòng này. Người quản trị danh mục vẫn cần xác định; đồng ý giữ bảng không tự cấp quyền quản trị cho HR/Admin.

### 1.5. `positions`

**Đề xuất chờ chốt:** giữ bảng với `position_id`, `position_name`, `created_at`, `updated_at`. Chức danh mô tả công tác, không tự cấp quyền hệ thống hoặc quyền PM/PTCM.

Chốt cùng quyết định `position_id`/`seniority_level` ở mục 1.3. Chưa thay đổi cấu trúc bảng.

</details>

## 2. Danh mục kỹ năng, năng lực chính thức và yêu cầu thay đổi — 6 bảng

### Quyết định hiện hành sau đối chiếu nhóm 2

- Giữ `skill_categories`; bỏ `skills.status` theo yêu cầu người dùng, không diễn giải kỹ năng có hạn dùng. Phải bỏ quy tắc catalog INACTIVE tương ứng ở tài liệu đích khi đồng bộ.
- `employee_skills` chỉ chứa năng lực hiện được công nhận; bỏ source/is_effective/version, REMOVE được duyệt gỡ dòng hiện tại nhưng giữ lịch sử quyết định/căn cứ. Công nhận lại tạo dòng mới sau xét duyệt. Nguồn MANUAL/CV giữ ở request, không xóa theo dòng năng lực.
- `employee_skill_change_requests`: bỏ base_version, đối chiếu before_values; giữ request.version để phát hiện nội dung/căn cứ thay đổi sau lúc mở màn hình. Không thêm bảng phiên bản.
- Giữ `employee_profile_change_requests`, bỏ base_version, đối chiếu before_values theo trường trong giao dịch; fullname/phone/address qua yêu cầu (address bổ sung 20/10), email công ty/cơ cấu do HR sửa trực tiếp. decision_reason nullable khi Pending/Approved, từ chối vẫn cần lý do.
- Giữ hướng minh chứng nhiều công việc, số task nội bộ có thể bằng 0. Liên kết theo assignment và submission tùy chọn, đúng người/đóng góp; không tự chứng nhận mức kỹ năng từ task đạt. Không thêm quy trình duyệt minh chứng riêng.
- Giám đốc quản trị skill_categories/skills và duyệt kỹ năng; HR không tự công nhận kỹ năng. Tối đa một CV dẫn trực tiếp qua evidence_cv_id; CV nguồn suy từ extracted_skill_id, không chép lặp cùng mục đích. Upload riêng dùng evidence_attachments JSONB theo phần chốt bổ sung, không bảng tệp mới. Request/CV/assignment/submission phải đúng nhân viên và cùng lượt giao tương ứng; không mở quyền đọc toàn dự án từ minh chứng.

<details>
<summary>Lịch sử trao đổi nhóm 2 — đề xuất cũ đã được thay thế theo quyết định hiện hành</summary>

Đã đối chiếu dictionary 5B và workflow kỹ năng/hồ sơ để trình bày. Các nội dung dưới đây là thiết kế hiện tại hoặc đề xuất của trợ lý, chưa có quyết định mới của người dùng cho nhóm 2.

### 2.1. `skill_categories`

Nhóm danh mục kỹ năng, gồm `category_id`, `category_name`, `created_at`, `updated_at`. Đề xuất giữ để phân nhóm tra cứu; không mang trọng số hay mức kỹ năng. Người quản trị danh mục chưa chốt.

### 2.2. `skills`

Danh mục kỹ năng dùng chung, gồm `skill_id`, `skill_name`, `category_id`, `status`, `created_at`, `updated_at`. Đề xuất giữ. INACTIVE ngừng dùng mới theo workflow, không xóa năng lực/lịch sử đang tham chiếu. Danh mục không lưu mức năng lực của từng người.

### 2.3. `employee_skills`

Năng lực chính thức của từng người/kỹ năng; duy nhất `(employee_id, skill_id)`. Giữ một identity khi gỡ/thêm lại; `is_effective` biểu thị còn được công nhận sử dụng, không xóa dòng lịch sử liên quan.

Đề xuất giữ cấu trúc hiện tại, bao gồm `version` phục vụ đối chiếu thay đổi năng lực. Không coi quyết định bỏ `employees.version` áp dụng cho bảng này. `source` chỉ nguồn cập nhật gần nhất, `approved_by`/`approved_at` chỉ quyết định gần nhất; lịch sử nằm ở yêu cầu/quyết định.

### 2.4. `employee_skill_change_requests`

Một yêu cầu cho một người/kỹ năng; dùng chung MANUAL/CV; ADD/UPDATE/REMOVE. Giám đốc xét duyệt, không tự duyệt; bổ sung căn cứ khi PENDING; quyết định kết thúc bất biến. Một PENDING/người/kỹ năng theo workflow đã có.

Đề xuất giữ bảng và giải thích riêng `base_version` (phiên bản năng lực được đối chiếu) với `version` (phiên bản nội dung yêu cầu). Chưa có quyết định bỏ hai trường này. JSON before/proposed phải có tập trường xác định, không nhận toàn bộ hồ sơ.

`extracted_skill_id` là mục trích xuất nguồn; `evidence_cv_id` là CV được dẫn làm căn cứ. Cần rà tiếp với nhóm 3 để tránh liên kết trùng khi cùng một CV; số CV/tệp minh chứng chưa chốt. `expert_opinion` ghi nguồn/nội dung nhận xét ngoài hệ thống, không tạo người duyệt mới.

### 2.5. `employee_profile_change_requests`

Yêu cầu thay đổi thông tin cá nhân, do HR xử lý; tách khỏi yêu cầu kỹ năng. Đề xuất giữ bảng với before/proposed, người gửi/thời điểm, trạng thái và quyết định. Danh sách trường được đề nghị sửa vẫn cần người dùng chốt.

Điểm chuyển tiếp từ nhóm 1: đề xuất bỏ `base_version` khi bỏ `employees.version`, đối chiếu các trường được đề nghị sửa với `before_values` dưới giao dịch bảo vệ cạnh tranh. Phương án thay thế chưa được người dùng chốt. Không coi timestamp là bảo vệ ghi đè.

### 2.6. `skill_request_task_evidence`

Liên kết một yêu cầu với đóng góp công việc nội bộ: `evidence_id`, `request_id`, `assignment_id`, `submission_id` tùy chọn, `contribution_note` tùy chọn. Cho 0..n minh chứng theo nghiệp vụ đã có; người mới có thể dùng giải trình/căn cứ ngoài.

Đề xuất giữ bảng liên kết thay vì mảng ID trong JSON. Dẫn assignment để xác định đúng người/lượt giao; submission nếu có phải thuộc assignment đó. Không suy quyền đọc toàn dự án từ liên kết. Không lặp employee/task vì suy được từ assignment.

### Những điểm chờ chốt của nhóm 2

- Giữ các bảng/trường sau khi người dùng trao đổi chi tiết; số lượng sáu bảng chưa mặc nhiên là quyết định cuối.
- Giữ hay thay cơ chế version ở năng lực/yêu cầu kỹ năng; hiện trợ lý đề xuất giữ.
- Cách đối chiếu khi duyệt yêu cầu hồ sơ sau quyết định bỏ employee version.
- Tập trường thông tin cá nhân được đề nghị sửa.
- Quan hệ CV nguồn/CV minh chứng và phạm vi tệp minh chứng, rà cùng nhóm 3.

### Phản hồi người dùng và rà soát tiếp nhóm 2

Mục này cập nhật các nhận xét ban đầu ở trên; tài liệu thiết kế khác chưa đổi.

- **Đã thống nhất:** giữ `skill_categories`.
- **Hướng người dùng yêu cầu:** bỏ `skills.status`, giữ các trường còn lại. Trường cũ là trạng thái danh mục, không phải hạn kỹ năng; trợ lý đề xuất không có chức năng vô hiệu hóa danh mục trong phiên bản hiện tại, bảo vệ liên kết bằng FK khi xóa. Quy tắc sửa/xóa danh mục cần đồng bộ sau.
- **Cần giải thích/chốt:** `employee_skills.source` là nguồn quyết định gần nhất MANUAL/CV; trợ lý đề xuất có thể bỏ vì nguồn đã lưu ở yêu cầu đã duyệt. Chưa coi đây là quyết định người dùng.
- **Hướng người dùng yêu cầu:** bỏ `employee_skills.is_effective`; người dùng cũng nghiêng về bỏ `version` nhưng cần trao đổi cơ chế thay thế. Trợ lý đề xuất bảng chỉ giữ năng lực hiện tại; REMOVE được duyệt xóa dòng hiện tại, giữ before/proposed/quyết định trong request và snapshot đã dùng. Đây là thay đổi so với quy tắc giữ identity cũ, cần được chốt riêng trước cập nhật; không xóa lịch sử và không chuyển mọi bảng sang xóa cứng.
- **Cần giải thích/chốt:** user chưa đồng ý bỏ hoặc giữ `employee_skill_change_requests.base_version`/`version`. Trợ lý đề xuất đối chiếu giá trị chính thức với `before_values` và bảo vệ giao dịch; nội dung yêu cầu được đọc lại khi duyệt, nếu cần phát hiện màn hình cũ thì đối chiếu bản nội dung người duyệt đã xem. Không tự bỏ bảo vệ ghi đè hoặc suy ra timestamp là đủ.
- **Đã thống nhất về bảng:** giữ `employee_profile_change_requests`. `decision_reason` có thể NULL khi PENDING và khi APPROVED theo đề xuất; REJECTED vẫn cần lý do theo nghiệp vụ hiện hành. Câu hỏi về nullable không tự bãi bỏ quy tắc từ chối có lý do.
- **Người dùng đánh giá phù hợp, yêu cầu rà thêm:** `skill_request_task_evidence`. Rà soát cho thấy bảng phù hợp nghiệp vụ nhiều đóng góp/0 task nội bộ. `assignment_id` xác định đúng người/lượt giao, `submission_id` tùy chọn xác định kết quả cụ thể; kiểm tra đúng subject, cùng assignment, quyền đọc có giới hạn và chống lặp. Dẫn task nghiệm thu không tự công nhận mức kỹ năng. Chưa thêm quy trình xác nhận minh chứng hoặc trạng thái mới.

### Quyết định chốt tiếp theo — năng lực hiện tại và phiên bản yêu cầu

**Đã thống nhất với người dùng:** phương án dưới đây thay các đề xuất còn chờ tương ứng ở trên.

1. `employee_skills` chỉ lưu kỹ năng hiện đang được công nhận. Bỏ `is_effective` và `version`. Khi yêu cầu REMOVE được duyệt, xóa dòng năng lực hiện tại; giữ nội dung trước thay đổi, căn cứ và quyết định trong yêu cầu đã xử lý, giữ nguyên các snapshot lịch sử đã dùng. Nếu công nhận lại, tạo dòng năng lực mới qua xét duyệt thông thường. Không còn quy tắc giữ identity năng lực sau REMOVE; phải rà liên kết trước khi triển khai, không xóa dây chuyền lịch sử.
2. `employee_skill_change_requests` bỏ `base_version`, giữ `before_values`. Khi duyệt, đối chiếu năng lực hiện tại với các giá trị nghiệp vụ trong bản trước thay đổi; trường hợp ADD chưa có năng lực thì kiểm tra vẫn chưa tồn tại. Nếu đã khác, báo để xem lại, không tự ghi đè. Kiểm tra và cập nhật trong cùng giao dịch có bảo vệ cạnh tranh.
3. Giữ `employee_skill_change_requests.version`, hệ thống tự quản lý. Khi nội dung/căn cứ đang PENDING thay đổi, tăng version. Khi quyết định, đối chiếu version người duyệt đã xem với version hiện tại; nếu khác thì yêu cầu tải lại/xem lại trước khi quyết định. Không thêm bảng lịch sử phiên bản hoặc chức năng quản lý phiên bản riêng.
4. Không suy rộng quyết định này sang trường version của task, nhật ký hoặc các bảng chưa rà soát.

Lý do giữ request.version: nghiệp vụ đã cho phép bổ sung căn cứ trong lúc chờ; một số nguyên giúp phát hiện màn hình xét duyệt cũ mà không phải so sánh toàn bộ nội dung và tập minh chứng.

Các điểm khác của nhóm 2 vẫn theo trạng thái riêng ở trên, đặc biệt `employee_skills.source`, các liên kết CV và allowlist thông tin cá nhân; chưa coi toàn bộ nhóm 2 đã chốt.

</details>

## 3. CV và dữ liệu trích xuất — 3 bảng

### Quyết định hiện hành sau đối chiếu nhóm 3

- Giữ ba bảng; lưu CV trên một dịch vụ cloud, bỏ storage_provider khi cấu hình chung, giữ định danh tệp và metadata/lựa chọn CV hiện tại.
- API nhận diện tên kỹ năng kèm đoạn căn cứ; bỏ suggested_level/suggested_years_experience. Giám đốc xét duyệt trước khi cập nhật năng lực chính thức.
- Không tự chọn Cloudinary/Gemini, không mở OCR/ảnh từ ví dụ lưu cloud. Cấu hình trích xuất, text/retry/lease và cách kết thúc mục chưa map vẫn cần thiết kế chi tiết, không phải các cột đã được chốt bỏ.

<details>
<summary>Lịch sử trao đổi nhóm 3 — đọc quyết định hiện hành trước các nhận xét ban đầu</summary>

Đã đối chiếu dictionary 5B và workflow CV. Chưa có quyết định mới của người dùng cho nhóm này.

### 3.1. `employee_cvs`

Metadata tệp và lựa chọn CV hiện tại: `cv_id`, `employee_id`, `storage_provider`, `storage_key`, `original_file_name`, `mime_type`, `file_size_bytes`, `uploaded_by`, `uploaded_at`, `is_current`.

Đề xuất giữ bảng. `is_current` biểu thị tài liệu nhân viên chọn sử dụng, không duyệt nội dung CV. Mỗi người tối đa một CV hiện tại; upload không tự chọn, parser lỗi không tự hủy lựa chọn. CV cũ còn làm căn cứ được giữ. HR upload thay ghi actor thực, quyền chọn thay vẫn chưa chốt.

`storage_provider`: đề xuất cân nhắc bỏ nếu chỉ dùng local private như phạm vi đã chọn và cấu hình lưu trữ áp dụng toàn hệ thống; cần rà storage_key/uniqueness khi chốt. Đây chưa là quyết định người dùng.

### 3.2. `cv_extractions`

Một tác vụ trích xuất cho mỗi CV, retry cùng tác vụ: `extraction_id`, `cv_id`, `processing_status`, `parser_version`, `extracted_text`, `attempt_count`, `lease_token`, `lease_expires_at`, `started_at`, `finished_at`, `error_code`, `created_at`, `updated_at`.

Đề xuất giữ bảng; tách tiến độ xử lý máy khỏi xét duyệt kỹ năng. `parser_version` là phiên bản bộ trích xuất, không phiên bản cập nhật hồ sơ. `lease_token`/`lease_expires_at` phục vụ worker DB lease/retry theo workflow hiện tại, không đưa thành thao tác nghiệp vụ. Chưa tự bỏ worker fields khi chưa đổi cách thực thi. `extracted_text` phục vụ đối chiếu nguồn; không lưu thêm markdown trùng lặp.

### 3.3. `cv_extracted_skills`

Các kỹ năng hệ thống nhận diện trước công nhận: `extracted_skill_id`, `extraction_id`, `source_ordinal`, `raw_skill_name`, `skill_id`, `suggested_level`, `suggested_years_experience`, `evidence_text`, `row_outcome`, `resolved_by`, `resolved_at`, `decision_reason`.

Đề xuất giữ bảng. suggested level/years được để trống nếu parser chưa có căn cứ, không tự suy mức từ tên kỹ năng. `skill_id` có thể chưa map. Mục có request lấy quyết định từ request; bộ quyết định dòng để trống. Mục không có request có thể ghi kết quả kết thúc trực tiếp theo workflow; không lưu hai quyết định cho cùng mục. Hợp nhất các dòng map cùng kỹ năng trước xét duyệt, không ghi đè tuần tự.

### Liên hệ nhóm 2 và điểm cần chốt

- `request.extracted_skill_id` dẫn qua extraction tới đúng CV nguồn; không cần chép thêm CV nguồn vào `evidence_cv_id` cho cùng mục đích.
- `evidence_cv_id` vẫn có mục đích riêng khi yêu cầu mới/thủ công dùng CV cũ làm căn cứ mà không mở lại mục trích xuất. Giữ hay giản lược và số CV/tệp được dẫn còn chờ chốt.
- Xem xét nhu cầu `storage_provider`, parser suggested fields và cơ chế worker ở mức phù hợp triển khai hiện tại.
- Quyết định gỡ skills.status ở nhóm 2 khi cập nhật phải gỡ cả quy tắc catalog INACTIVE tương ứng trong workflow CV; chưa chỉnh workflow trong lượt này.
- Không thêm phê duyệt CV, điểm confidence hoặc tự động cập nhật năng lực từ kết quả trích xuất.

### Quyết định ngày 15/10/2026 — phiên bản CV giản lược

**Đã thống nhất với người dùng:** giữ ba bảng và hướng triển khai sau. Các nhận xét sơ bộ ở trên phải đọc theo quyết định này.

1. Lưu CV trên một dịch vụ cloud. Bỏ `employee_cvs.storage_provider` khi nhà cung cấp được cấu hình chung toàn hệ thống; giữ định danh tệp (`storage_key`, tên cuối rà khi cập nhật). Định danh có thể do dịch vụ trả về. Giữ tên gốc, chủ CV, người/thời điểm upload, metadata tệp và lựa chọn CV hiện tại. Cloudinary mới là ví dụ, chưa chốt nhà cung cấp. Đây là thay đổi hướng lưu local trong tài liệu cũ, cần đồng bộ sau.
2. Dùng API trích xuất để nhận diện **tên kỹ năng kèm đoạn căn cứ**. Gemini là phương án đang cân nhắc, chưa chốt dịch vụ/model. Kết quả được đối chiếu với danh mục và qua Giám đốc xét duyệt trước khi ghi năng lực chính thức; không tự công nhận mức kỹ năng.
3. Bỏ `cv_extracted_skills.suggested_level` và `suggested_years_experience` trong phiên bản đầu. Mức công nhận được xác định trong luồng xét duyệt. Không thêm confidence hoặc điểm AI chưa có nhu cầu/cách đo.
4. Giữ `cv_extractions` để theo dõi trạng thái, kết quả và lỗi của tác vụ. Chi tiết cấu hình model/prompt, lưu văn bản, retry, lease và các trường kỹ thuật chỉ chốt sau khi xác định cách chạy API thực tế; chưa mặc nhiên duyệt hoặc bỏ toàn bộ danh sách trường worker hiện tại.
5. Giữ phần cốt lõi `cv_extracted_skills`: định danh mục/tác vụ, tên nhận diện, liên kết danh mục và đoạn căn cứ. Các trường số thứ tự/xử lý dòng đã giải thích, chưa có quyết định bỏ; rà để chỉ lưu quyết định ở một nơi theo từng trường hợp. Không xóa cơ chế kết thúc mục chưa map mà chưa có phương án thay thế.
6. Không mở rộng loại tệp sang ảnh/OCR chỉ từ ví dụ lưu ảnh/PDF trên cloud. Phạm vi tệp hiện tại vẫn PDF có text/DOCX cho đến khi người dùng quyết định khác. Cách truy cập tệp phải giữ quyền xem CV đã có.

Chỉ cập nhật sổ quyết định trong lượt này; chưa sửa dictionary, DBML, workflow, cấu hình hoặc mã nguồn.

</details>

## 4. Dự án, công việc, phân công và nghiệm thu — 9 bảng sau giản lược (ban đầu 10)

### Quyết định hiện hành sau đối chiếu nhóm 4

- Giám đốc chọn PM ngay khi tạo dự án; PM bắt buộc. Bỏ projects.version.
- Giữ project_status_history; task_groups và quan hệ đã trình bày được đồng ý.
- Bỏ tasks.version/spec_version; gộp deliverable vào description, giữ acceptance_criteria rõ trước khi giao. Quy tắc khóa đặc tả sau giao đầu vẫn giữ. Nhóm 6 phải thay phụ thuộc spec_version bằng đối chiếu đầu vào snapshot.
- Giữ assignments; bỏ assignment_events; thêm ended_at/ended_by/end_reason. Kết thúc lượt cũ và tạo lượt mới khi giao lại, không đổi người trên lượt cũ.
- Giữ task_submissions và log_snapshot, task_status_history, task_comments. Snapshot phải được cập nhật contract theo quyết định nhóm 5 bỏ log.version, không ghi đè dữ liệu đã chụp.
- Giữ task_requirements về cấu trúc; đã chốt PTCM đặt weight theo task trước giao, không trọng số cố định cho kỹ năng toàn công ty. Miền giá trị/cách chuẩn hóa và quyền cấu hình trọng số nhóm tiêu chí công thức chốt khi đối chiếu công thức. Requirements đóng băng sau giao đầu.
- Các câu hỏi task_id/employee_id/submitted_by của submission mới được đánh dấu để rà; chưa có quyết định bỏ giống work_logs.

<details>
<summary>Lịch sử trao đổi và nghiên cứu nhóm 4 — bảng assignment_events trong phần này là thiết kế cũ</summary>

Đã đối chiếu dictionary 5A và task_comments trong 5B để trình bày đủ trường và quan hệ. Chưa có quyết định sửa mới của người dùng cho nhóm 4.

| Bảng | Trách nhiệm và quan hệ chính |
|---|---|
| `projects` | Dự án; mỗi dự án có PM và PTCM hiện hành trỏ employees; có nhiều nhóm/task/lịch sử trạng thái. |
| `project_status_history` | Mỗi dòng là một chuyển trạng thái dự án, trỏ projects và actor users. |
| `task_groups` | Nhóm để gom task trong một projects; có nhiều tasks, không có người thực hiện/estimate riêng. |
| `tasks` | Đơn vị giao việc; thuộc một project và một nhóm cùng project; nhiều requirements/assignments/submissions/history/comments. |
| `task_requirements` | Liên kết task–skill, kèm mức yêu cầu, trọng số, tính bắt buộc; duy nhất task/skill. |
| `assignments` | Một lượt giao đã xác nhận, trỏ task/người nhận/người giao; run/item tùy phương thức; có nhiều lần nộp/log/minh chứng. |
| `assignment_events` | Sự kiện giao/thay/thu hồi; trỏ task, lượt cũ/mới cùng task và actor; dùng xác định chuỗi lượt giao. |
| `task_submissions` | Một lần nộp gắn đúng lượt/người/task; giữ kết quả, snapshot log và tối đa một quyết định nghiệm thu. |
| `task_status_history` | Một chuyển trạng thái task, trỏ tasks và actor users; không thay quyết định nghiệm thu/lịch sử phân công. |
| `task_comments` | Trao đổi gắn task và tác giả employees; không là quyết định nghiệm thu. |

Các điểm cần cùng người dùng rà tiếp:

- `projects.pm_employee_id` hiện bắt buộc, trong khi luồng vẽ tạo dự án rồi chỉ định PM: chốt chỉ định trong cùng lần tạo hay cho dự án chưa có PM; không tự sửa nullable.
- Mục đích `original_*`, `version`, `spec_version`, `first_assigned_at`, `accepted_at` và dữ liệu projection hiện hành; không suy rộng quyết định bỏ version nhóm 1–2.
- `task_requirements.weight` 1..5 là thiết kế hiện tại, cần đối chiếu công thức phân công trước chốt; không coi đó là trọng số KPI.
- `assignments` khác `assignment_events`: bản lượt giao và sự kiện thay đổi chuỗi. Chưa chốt giản lược hay gộp.
- task/employee lặp ở submission nhằm FK ghép, có thể suy từ assignment về mặt logic; cần giải thích trước quyết định giữ/bỏ.
- `submitted_by` bị buộc là chủ assignment theo nghiệp vụ không nộp thay; xem có cần giữ cột riêng.
- Snapshot nhật ký phục vụ giữ dữ liệu lần nộp nhưng nội dung chi tiết chốt cùng nhóm 5; không tự bỏ theo nhận xét về phiên bản hồ sơ.
- Giữ ranh giới trách nhiệm đã có: PTCM phân công/nghiệm thu, không nhận task trong dự án phụ trách; PM bắt đầu/theo dõi/hoàn thành/đóng; không thêm reviewer riêng, roster hoặc kế hoạch/sprint trong hệ thống.

### Phản hồi và nghiên cứu tiếp nhóm 4 — 15/10/2026

**Người dùng đồng ý:** quan hệ projects, task_groups, assignments, task_submissions (cần giải thích snapshot), task_status_history và task_comments. Task requirements được đồng ý về bảng, còn hỏi ai xác lập weight. Không coi các câu hỏi về bỏ trường/bảng là đã chốt phương án thay thế.

Các đề xuất sau nghiên cứu, chờ người dùng chốt:

- Giám đốc chọn PM ngay trong thao tác tạo dự án; tạo dự án và gán PM cùng giao dịch, `pm_employee_id` tiếp tục bắt buộc. Không thêm bước duyệt PM. Thay PM sau này vẫn theo quyền Giám đốc.
- Có thể bỏ `projects.version` và `tasks.version` nếu các thao tác ghi đọc lại điều kiện hiện hành dưới giao dịch, cập nhật đúng trường, đối chiếu giá trị nền khi cần chống ghi đè màn hình cũ. Khóa giao dịch không tự phát hiện màn hình cũ.
- Đề xuất bỏ `tasks.spec_version`, nhưng phải đồng bộ `recommendation_runs.spec_version` ở nhóm 6. Lúc confirm, đối chiếu đầu vào đặc tả/requirements với snapshot của run; nếu khác thì chạy lại. Không giữ con trỏ tới phiên bản đã bỏ hoặc bỏ kiểm tra gợi ý cũ. Quy tắc khóa đặc tả sau lần giao đầu vẫn giữ.
- Đề xuất giữ `project_status_history`: nguồn mốc triển khai/hoàn thành/đóng thực tế và người/lý do chuyển trạng thái. Nếu chọn bỏ, phải xác định nơi giữ các mốc/lịch sử thay thế; không mất dữ liệu chỉ vì projects.status còn tồn tại. Chưa chốt bỏ.
- Nghiên cứu thực hành xác nhận cần đầu ra và điều kiện nghiệm thu rõ; không chứng minh phải có hai cột riêng. Đề xuất giữ `acceptance_criteria` làm văn bản ngắn, gộp đầu ra vào `description`, bỏ cột `deliverable`. PTCM đảm bảo nội dung đủ trước khi giao, có thể kế thừa/viết cùng người liên quan, không bắt văn bản dài hoặc quy trình duyệt mới. Cần đồng bộ mọi quy tắc đặc tả/snapshot/API nếu chọn hướng này.
- `task_requirements.weight` là mức quan trọng của kỹ năng trong task, đề xuất PTCM xác định theo task; khác trọng số các nhóm tiêu chí trong công thức do công ty cấu hình theo chính sách được chốt. Chưa cấp quyền cấu hình công thức cho Giám đốc chỉ từ câu hỏi này; thang 1..5 chờ đối chiếu công thức.
- Đề xuất bỏ `assignment_events`, bổ sung `ended_at`, `ended_by` (users), `end_reason` vào assignments. Chỉ cập nhật bộ kết thúc một lần; không đổi task/người nhận/người giao/thời điểm giao/run/item cũ. Giao lại: kết thúc lượt cũ và tạo lượt mới cùng giao dịch; thu hồi chỉ kết thúc lượt cũ; bảo đảm tối đa một lượt chưa kết thúc cho mỗi task. Nhãn lý do phải phân biệt thay người/thu hồi/kết thúc công việc theo quy tắc được chốt. Không dùng cập nhật employee_id trên dòng cũ, không giữ cả event và bộ kết thúc như hai nguồn chuẩn. Mốc kết thúc lượt không tự làm mất quyền hiệu chỉnh nhật ký lịch sử.
- `log_snapshot` là bản chụp nhật ký của lượt giao tại lúc nộp; giữ lịch sử lần nộp khi nhật ký sau đó được sửa hợp lệ. Snapshot không phải số effort hiện hành hoặc phép chốt công thức Q/T; không lấy giờ snapshot cộng lại với work_logs. Đề xuất giữ, chi tiết rà nhóm 5.

Nguồn đã đọc và phạm vi áp dụng:

- [Atlassian — Acceptance criteria](https://www.atlassian.com/work-management/project-management/acceptance-criteria): điều kiện rõ, ngắn, kiểm tra được và có thể xây dựng cùng người liên quan; không quy định schema hai cột deliverable/criteria.
- [Towards Just-Enough Documentation for Agile Effort Estimation](https://arxiv.org/abs/2107.02420): khảo sát về tài liệu phục vụ ước tính, gồm acceptance criteria; không coi là bằng chứng mô hình DB của ta đã được kiểm chứng.
- [PostgreSQL — Explicit locking](https://www.postgresql.org/docs/current/explicit-locking.html): khóa dòng trong giao dịch bảo vệ thao tác đồng thời; không bảo vệ bản màn hình cũ tự động.
- [Microsoft — Temporal tables](https://learn.microsoft.com/en-us/sql/relational-databases/temporal/overview?view=sql-server-ver17): minh họa sự khác nhau giữa dữ liệu hiện tại và dữ liệu lịch sử theo thời gian. Không áp dụng cơ chế temporal SQL Server vào PostgreSQL; đề xuất ended_at là lựa chọn thiết kế của ta, không phải schema phân công bắt buộc từ tài liệu này.

Chưa cập nhật dictionary/DBML/workflow theo các đề xuất này. Số bảng vẫn ghi 34 cho đến khi người dùng chốt việc bỏ bảng.

### Quyết định tiếp nối nhóm 4 — 15/10/2026

Người dùng đồng ý các đề xuất ở lượt trước, ngoại trừ còn hỏi rõ weight và project_status_history. Ghi nhận **đã thống nhất**:

- Giám đốc chọn PM ngay trong thao tác tạo dự án; `pm_employee_id` bắt buộc.
- Bỏ `projects.version`, `tasks.version`, `tasks.spec_version`; giữ kiểm tra quyền/trạng thái và chống ghi đè phù hợp trong giao dịch. Đồng bộ cơ chế đối chiếu đặc tả snapshot ở nhóm 6 thay phụ thuộc `recommendation_runs.spec_version` cũ.
- Bỏ `tasks.deliverable`, gộp đầu ra vào `description`; giữ `acceptance_criteria` ngắn, đủ để đối chiếu nghiệm thu trước khi giao.
- Bỏ `assignment_events`; thay bằng `assignments.ended_at`, `ended_by`, `end_reason` theo phương án đã trình bày. Giữ identity/người nhận/nội dung lượt cũ; giao lại tạo lượt mới, không ghi đè employee_id. Chi tiết ràng buộc kết thúc và liên kết rà lúc hợp nhất. Đây là bỏ một bảng đã được người dùng đồng ý, số đích tạm thời giảm từ 34 xuống **33 bảng**, nhóm 4 từ 10 xuống **9 bảng**; danh mục 34 phía trên là mốc bắt đầu, chưa sửa dictionary/DBML.
- Giữ `task_submissions.log_snapshot`, nội dung rà cùng nhóm 5; không tự đổi công thức Q/T.

**Còn chờ chốt:** project_status_history và cách cấu hình weight.

Giải thích tiếp: weight của từng dòng requirement do PTCM chọn theo yêu cầu task, không trọng số cố định toàn công ty cho từng kỹ năng; miền giá trị/cách chuẩn hóa chốt theo công thức. Quy tắc đóng băng requirements sau giao đầu vẫn còn, nên "tùy lúc" không mặc nhiên cho đổi tùy ý sau giao. Trọng số nhóm tiêu chí của công thức là cấu hình khác, quyền cấu hình chưa chốt.

projects hiện chỉ có ngày kế hoạch hiện tại/ban đầu, trạng thái hiện tại và timestamps tạo/cập nhật. Chưa có ngày thực tế started/completed/closed/cancelled hoặc actor/lý do từng chuyển trạng thái. Đề xuất giữ project_status_history để lưu các mốc và người thao tác; ví dụ PM hiện tại có thể khác người đã đánh dấu hoàn thành. Chưa ghi việc giữ bảng này là đã được người dùng đồng ý.

### Chốt giữ lịch sử trạng thái dự án — 15/10/2026

**Đã thống nhất với người dùng:** giữ `project_status_history` và quan hệ với `projects`/actor `users`. Bảng tự ghi các lần chuyển trạng thái, thời điểm thực tế, người thao tác và lý do khi cần; không thêm bước nhập lịch sử riêng. Quyết định này kết thúc câu hỏi giữ/bỏ bảng ở các mục trước.

Số bảng đích tạm thời vẫn **33** do đã bỏ `assignment_events`. Quyền đặt weight và miền giá trị vẫn theo mục trao đổi riêng; lời đồng ý giữ bảng lịch sử không tự chốt cấu hình trọng số.

</details>

## 5. Nhật ký công việc và hiệu chỉnh — 3 bảng

### Bản chốt hiện hành — 15/10/2026

Người dùng đã đồng ý phương án xóa mềm ở lượt trao đổi mới nhất. Nội dung này thay thế các phương án cũ/chờ chốt về nhóm 5 bên dưới.

| Bảng | Danh sách trường sau chốt |
|---|---|
| `work_logs` — 9 trường | `timelog_id`, `assignment_id`, `work_date`, `duration_minutes`, `description`, `over_daily_limit_reason`, `deleted_at`, `created_at`, `updated_at` |
| `work_log_history` — 9 trường | `history_id`, `timelog_id`, `operation`, `before_values`, `after_values`, `actor_id`, `occurred_at`, `reason`, `request_id` |
| `work_log_change_requests` — 14 trường | `request_id`, `assignment_id`, `operation`, `target_log_id`, `result_log_id`, `before_values`, `proposed_values`, `requested_by`, `requested_at`, `reason`, `status`, `resolved_by`, `resolved_at`, `decision_reason` |

Quy tắc đã thống nhất:

- **Bổ sung được người dùng đồng ý ngày 15/10: `result_log_id` nullable.** Chỉ ADD đã APPROVED có giá trị, liên kết log được tạo thuộc cùng assignment; các trường hợp khác NULL. ADD vẫn có target_log_id/before_values NULL. Tạo log, ghi result_log_id và quyết định cùng giao dịch; khóa yêu cầu PENDING để không áp dụng lặp. Không tạo history ADD giả, không thêm bảng. Log kết quả chỉ thuộc một yêu cầu ADD; xóa mềm log không xóa liên kết truy vết này.
- Người thực hiện tự ghi/gửi yêu cầu, không nhập hoặc gửi thay. Người làm/task suy từ assignment; bỏ created_by/updated_by/task_id/employee_id lặp ở log và task_id/employee_id ở request. Actor sửa vẫn lưu tại history vì PTCM có thể áp dụng yêu cầu đã duyệt.
- work_date là ngày cụ thể, không số ngày hoặc khoảng ngày. Thời lượng lưu phút; nhiều log/ngày hợp lệ. Giữ các kiểm tra thời lượng/ngày đã có.
- Còn quyền Edit: người thực hiện sửa hoặc xóa có lý do, hệ thống ghi history. Đã khóa: gửi yêu cầu, PTCM hiện tại của dự án duyệt. CLOSED/CANCELLED chỉ đọc. Pending không thay số liệu chính thức.
- Xóa trên giao diện là xóa mềm: đặt deleted_at và loại khỏi tổng/danh sách thông thường. Không thêm is_deleted hoặc trạng thái trùng nghĩa; không xóa vật lý, không dùng 0 phút. Nhật ký đã xóa chưa có chức năng khôi phục/sửa/xóa lại.
- History ghi UPDATE/DELETE, giữ before/after/người/thời điểm/lý do, request_id tùy chọn cho thao tác áp dụng từ yêu cầu. Không version hoặc bảng xóa riêng. Tạo mới không tạo bản history Edit; lượt giao và created_at xác định nguồn ghi của người thực hiện.
- Yêu cầu dùng ADD/UPDATE/DELETE; bỏ base_version và evidence riêng, căn cứ tùy chọn gộp vào reason. ADD không có log đích/bản trước; UPDATE có target/before/proposed; DELETE có target/before và proposed NULL. Log đích phải thuộc assignment. Từ chối cần decision_reason, chấp thuận có thể để trống.
- Đối chiếu dữ liệu nền trước sửa/duyệt và kiểm tra trạng thái dưới giao dịch bảo vệ cạnh tranh; quyết định và thay đổi dữ liệu ghi nguyên tử, không áp dụng cùng yêu cầu hai lần. Không giữ request.version vì luồng này chưa mở bổ sung/sửa nội dung Pending như yêu cầu kỹ năng.
- Giữ snapshot lần nộp bất biến, bỏ phần log.version khi cập nhật contract. Tổng hiện tại chỉ cộng work_logs có deleted_at NULL; không cộng history/snapshot hoặc tự đổi Q/T. Tổng đích hiện hành là 32 bảng sau quyết định nhóm 7.

<details>
<summary>Lịch sử trao đổi nhóm 5 — các nhắc tới PM duyệt/ghi thay, VOID/voided_at và version là thiết kế cũ</summary>

Phần tiếp theo ghi lại lịch sử phân tích, không phải danh sách trường hiện hành. Các nhắc tới PM duyệt/ghi thay, VOID/voided_at và version dưới đây đã bị các quyết định tiếp nối thay thế.

Đã đối chiếu dictionary 5A và workflow effort. Chưa có quyết định mới của người dùng cho nhóm 5.

### 5.1. `work_logs`

Nhật ký thời lượng thực tế tự khai, không chấm công. Một dòng thuộc một lượt giao và một ngày; nhiều dòng/ngày/task hợp lệ. Trường hiện tại: `timelog_id`, `assignment_id`, `task_id`, `employee_id`, `work_date`, `duration_minutes`, `description`, `over_daily_limit_reason`, `created_by`, `updated_by`, `version`, `voided_at`, `created_at`, `updated_at`.

Quan hệ: assignments 1–n logs; employees/users được tham chiếu với nghĩa người làm/người thao tác; mỗi log 1–n history, 0–n correction requests theo thời gian. Task/employee suy được từ assignment, lưu lặp hiện nhằm FK ghép, cần rà khi hợp nhất thay vì tự bỏ.

Đề xuất giữ log hiệu lực là nguồn tính tổng thời gian, VOID chỉ loại khỏi tổng và giữ lịch sử. `version` trong nhóm này đang là số thứ tự bản nhật ký và phục vụ hiệu chỉnh, không tự áp dụng quyết định bỏ version hồ sơ/task. Chưa chốt bỏ/giữ với người dùng.

### 5.2. `work_log_history`

Lịch sử ADD/UPDATE/VOID tự ghi cùng thao tác thay đổi log: `history_id`, `timelog_id`, `version`, `operation`, `after_values`, `actor_id`, `occurred_at`, `reason`, `request_id` tùy chọn.

Quan hệ: mỗi dòng thuộc một work_log; actor trỏ users; request_id khi có trỏ đúng yêu cầu đã áp dụng. Mỗi yêu cầu được duyệt tạo một dòng history, yêu cầu chưa duyệt/từ chối/rút không tạo history thay đổi log. ADD đầu tiên tạo version 1; giá trị trước của bản n lấy bản n−1. Đề xuất giữ để biết ai đổi thời lượng/nội dung, không cộng history vào effort.

### 5.3. `work_log_change_requests`

Yêu cầu hiệu chỉnh nhật ký đã khóa, do PM hiện tại xử lý: `request_id`, `assignment_id`, `task_id`, `employee_id`, `operation`, `target_log_id`, `base_version`, `before_values`, `proposed_values`, `requested_by`, `requested_at`, `reason`, `evidence`, `status`, `resolved_by`, `resolved_at`, `decision_reason`.

Quan hệ: luôn thuộc một assignment; UPDATE/VOID trỏ log đích cùng assignment/người/task, ADD chưa có log đích; actor gửi/xử lý trỏ users. Duyệt ADD tạo log, duyệt UPDATE/VOID tác động log hiện hữu; history.request_id liên kết kết quả áp dụng. Tối đa một PENDING cho cùng log đích theo workflow hiện tại.

Phân biệt reason (lý do gửi), evidence (căn cứ tùy chọn), decision_reason (lý do quyết định, cần khi từ chối). Không thêm bảng minh chứng hay cơ chế upload cho correction.

### Luồng, snapshot và điểm cần rà

- Còn quyền Edit: nhân viên sửa trực tiếp, hệ thống ghi history, không cần tạo request/PM duyệt. Nhật ký đã khóa: gửi request, Pending chưa đổi log; PM duyệt mới cập nhật log/history/quyết định cùng giao dịch.
- `log_snapshot` ở submission là bản chụp dữ liệu khi nộp, không tự đổi sau hiệu chỉnh. work_logs là hiện tại, history là các thay đổi, snapshot là bối cảnh lần nộp; không cộng ba nơi thành tổng.
- Đề xuất giữ cả ba bảng vì ba trách nhiệm khác nhau. Version/base_version đang là phương án hiện tại, cần trao đổi riêng theo nhu cầu truy vết nhật ký; không trình bày như người dùng đã chốt.
- Các kiểm tra ngày/thời lượng, lý do vượt 8h và giới hạn tổng 24h là nghiệp vụ hiện có, không phải chức năng mới phát sinh khi trình bày nhóm này. Rà lại nếu người dùng muốn giản lược.
- Không mở lại execution sau nghiệm thu để sửa giờ; project CLOSED/CANCELLED chỉ đọc; không thêm luồng PM tự duyệt đóng góp của mình đã đưa ngoài phạm vi.

### Rà soát giản lược nhóm 5 theo phản hồi người dùng — 15/10/2026

**Quyết định nghiệp vụ trực tiếp từ người dùng:** PTCM hiện tại của dự án xử lý yêu cầu hiệu chỉnh nhật ký, thay PM trong quy trình cũ. Phải đồng bộ quyền/workflow/API/điều kiện đóng dự án và mô tả quyết định/thông báo liên quan; không coi PM còn quyền duyệt từ vai trò PM. Quyền PM ghi thay trước đây là nội dung riêng, chưa tự thu hồi hoặc chuyển toàn bộ cho PTCM. Không tự thêm cơ chế tự duyệt của người kiêm nhiệm.

Giải thích work_date: một ngày cụ thể (DATE), không số ngày hoặc khoảng từ–đến. Làm 09–11/10 ghi phần thời gian thực tế riêng cho từng ngày. Không chia đều tự động một tổng giờ qua khoảng ngày.

Các phương án dưới đây là **đề xuất chờ người dùng chốt**, chưa sửa thiết kế khác:

1. `work_logs`: bỏ version và updated_by (người sửa đọc từ history), bỏ task_id/employee_id lưu lặp nếu dùng assignment_id làm nguồn quan hệ chuẩn. Giữ created_by vì có nghiệp vụ PM ghi thay; khác người thực hiện suy từ assignment. Giữ work_date/duration/description, over_daily_limit_reason, created_at/updated_at.
2. `voided_at` trước đây để loại khoản ghi nhầm khỏi tổng mà giữ bản ghi. Không tự xóa cứng log vì history/correction đang tham chiếu. Đề xuất phiên bản hiện tại chỉ hỗ trợ tạo và sửa, bỏ thao tác VOID cùng voided_at; đây là giới hạn nghiệp vụ mới cần người dùng chốt. Khoản ghi trùng hoàn toàn sẽ chưa có chức năng loại/xóa; không dùng phút 0 hoặc sửa gian nội dung để lách. Nếu người dùng cần loại khoản trùng thì phải giữ cách biểu diễn phù hợp, chưa quyết định ngầm.
3. `work_log_history` giản lược thành lịch sử Edit: history_id, timelog_id, before_values, after_values, actor_id, occurred_at, reason, request_id tùy chọn. Bỏ version/operation trong phương án chỉ ghi Edit; tạo mới truy vết từ created_by/created_at của work_logs. Chuyển từ after-only sang before+after để mỗi lần sửa tự giải thích được. Ghi history cùng sửa log dưới khóa log, không sửa/xóa history.
4. `work_log_change_requests`: giữ request_id, assignment_id, operation ADD/UPDATE, target_log_id, before_values, proposed_values, requested_by, requested_at, reason, status, resolved_by, resolved_at, decision_reason (13 cột). Bỏ task_id/employee_id lưu lặp, base_version và evidence riêng (căn cứ/đường dẫn nếu cần gộp vào reason). ADD không target/before, UPDATE có target/before; target phải cùng assignment. Không dùng NULL ẩn loại thao tác thay cột operation rõ nghĩa.
5. Bỏ log/base version thì đối chiếu giá trị hiện tại với before_values trước duyệt UPDATE và khi Edit từ màn hình cũ; kiểm tra/ghi trong giao dịch bảo vệ cạnh tranh. Không có bổ sung nội dung yêu cầu khi PENDING trong luồng hiện tại nên không tự thêm request.version. APPROVED bất biến, một request chỉ áp dụng một lần; FK/history unique không tự thay kiểm tra giao dịch.
6. Snapshot lần nộp bỏ tham chiếu log.version nếu bỏ cột, vẫn giữ dữ liệu tại thời điểm nộp. Tổng hiện tại lấy work_logs; history/snapshot không cộng vào tổng và không tự đổi công thức Q/T.

Chưa coi người dùng đã đồng ý bộ 9 trường work_logs/8 trường history/13 trường correction hoặc phạm vi không xóa. Chỉ chuyển người duyệt sang PTCM là quyết định đã nêu rõ.

### Quyết định tiếp nối nhóm 5 — bỏ ghi thay, chốt giản lược, giữ mở cách xử lý ghi nhầm

**Đã thống nhất với người dùng:**

1. Bỏ nghiệp vụ nhập nhật ký thay. Người thực hiện tự ghi nhật ký của mình; không cho PM/PTCM tạo hoặc gửi yêu cầu hiệu chỉnh thay người thực hiện. PTCM vẫn duyệt và áp dụng thay đổi được yêu cầu qua quy trình, không được hiểu đây là thao tác nhập thay. Người dùng bỏ case ghi thay và đồng ý các giản lược còn lại; không mở luồng xử lý thay cho người mất quyền.
2. Bỏ `work_logs.created_by` vì người tạo luôn là người thực hiện suy từ assignment. Bỏ `updated_by` (người sửa lưu ở history), `version`, `task_id`, `employee_id` lưu lặp; assignment_id là liên kết chuẩn. Ngày làm là DATE cụ thể, không khoảng ngày. Các trường còn lại giữ theo mục đã giải thích; voided_at chưa quyết định.
3. Giữ lịch sử Edit với `history_id`, `timelog_id`, `before_values`, `after_values`, `actor_id`, `occurred_at`, `reason`, `request_id` tùy chọn; bỏ số phiên bản lịch sử. Bộ 8 trường mô tả luồng Edit đã được đồng ý. Nếu sau đó cần ghi cả xóa/loại thì phải rà bổ sung biểu diễn thao tác, chưa chốt bỏ operation cho mọi tình huống.
4. Đồng ý hướng 13 trường yêu cầu hiệu chỉnh được trình bày: bỏ task_id/employee_id lặp, base_version, evidence riêng; gộp căn cứ tùy chọn vào reason, giữ actor/thời điểm gửi và xử lý để truy vết. requested_by lấy từ phiên và phải đúng chủ assignment, không cho gửi thay. Các trường operation/target/before phải rà theo quyết định xử lý khoản ghi nhầm sau này.
5. Đối chiếu before_values với dữ liệu hiện tại trước sửa/duyệt, dưới giao dịch bảo vệ cạnh tranh; history và quyết định ghi cùng thay đổi. Snapshot lần nộp giữ nguyên nội dung đã chụp, bỏ phần log.version khi cập nhật contract. Không tự đổi Q/T.

**Chưa chốt:** cách xử lý khoản nhật ký ghi nhầm/trùng hoàn toàn. Người dùng đồng ý cần xem kỹ, không đồng ý giới hạn chắc chắn chỉ tạo/Edit. Do đó chưa bỏ VOID/voided_at trong quyết định cuối và chưa chuyển sang xóa cứng hoặc dùng 0 phút. Cần chọn cách xử lý bảo toàn truy vết trước khi cập nhật tài liệu đích.

Ảnh hưởng phải cập nhật sau: workflow effort/roles/API và bảng quyền ghi thay cũ, các ví dụ PM nhập thay, ràng buộc actor/assignment, dữ liệu history/correction, contract snapshot và migration. Chưa cập nhật các file đó trong vòng trao đổi này.

### Phương án xóa khoản ghi nhầm/trùng — đã được chấp thuận

**Đã thống nhất:** dùng thao tác giao diện "Xóa nhật ký", triển khai xóa mềm bằng `work_logs.deleted_at` thay tên `voided_at`. NULL là đang được tính; có thời điểm là đã xóa khỏi danh sách thông thường/tổng hiện tại. Một cột duy nhất, không thêm is_deleted hoặc trạng thái lưu cùng nghĩa.

- Còn quyền sửa trực tiếp: chủ nhật ký xóa có lý do; nhật ký đã khóa: gửi yêu cầu DELETE, PTCM xử lý theo luồng hiệu chỉnh đã có. Project CLOSED/CANCELLED vẫn chỉ đọc. Không nhập thay/duyệt tự động.
- Không xóa vật lý để giữ FK từ history/requests và bảo toàn bối cảnh snapshot. Không giữ 0 phút giả. Snapshot lần nộp không bị sửa theo xóa sau đó; tổng hiện tại bỏ dòng có deleted_at. Dòng đã xóa không được sửa/xóa lại trong phiên bản này, chưa thêm khôi phục.
- `work_log_history` cần `operation` với UPDATE/DELETE để phân biệt hai thao tác. Bản Edit đã đồng ý giữ before/after; DELETE cũng lưu bản trước/sau với deleted_at thay đổi. Như vậy history dự kiến 9 trường, không version, không thêm bảng lịch sử xóa.
- `work_log_change_requests.operation`: ADD/UPDATE/DELETE thay VOID. DELETE có target_log_id/before_values/lý do, proposed_values để trống. Giữ bộ 13 trường đã đề xuất; phê duyệt kiểm tra log chưa xóa, cùng assignment và giá trị nền còn khớp rồi đánh dấu/ghi history/quyết định cùng giao dịch. Tối đa một Pending/log đích dùng chung UPDATE/DELETE.
- Nguồn log không xóa nghĩa vụ truy vết; lịch sử và kết quả gợi ý đã lưu không bị viết lại. Không tự thay quy tắc Q/T vì xóa nhật ký.

Người dùng đã đồng ý phương án này; bản chốt hiện hành ở đầu mục 5 tổng hợp đầy đủ các thay đổi liên quan.

</details>

## Kiểm tra đối chiếu quyết định nhóm 1–5 — 15/10/2026

Đã đối chiếu lại với các lời đồng ý/yêu cầu thay đổi trong cuộc trao đổi. Các quyết định đã được ghi; không tuyên bố mọi chi tiết của năm nhóm đều đã đóng. Đã bổ sung skills.status bị thiếu trong phần tóm tắt, đưa quyết định hiện hành lên đầu từng nhóm 2–4 và gom diễn tiến cũ vào phần thu gọn để tránh đọc nhầm đề xuất cũ thành bản cuối.

Kiểm tra nhất quán xuyên nhóm:

- Không còn đưa version hồ sơ/task/log đã bỏ thành quyết định hiện hành; request.version kỹ năng vẫn được giữ có chủ đích. Contract liên quan phải đồng bộ khi chỉnh tài liệu đích.
- Cột trạng thái skill/is_effective đã bỏ khác deleted_at nhật ký đã được chốt giữ; không suy rộng xóa năng lực hiện tại thành xóa lịch sử nhật ký.
- Quyền HR duyệt thông tin cá nhân/Giám đốc duyệt kỹ năng không đổi; PTCM duyệt hiệu chỉnh nhật ký thay PM. Bỏ ghi nhật ký thay không tự bỏ quyền HR upload CV thay.
- Giữ đúng ranh giới dữ liệu hiện tại/lịch sử/snapshot; sửa hoặc xóa mềm log không sửa lại snapshot nộp/gợi ý đã có.
- Tại mốc kiểm tra nhóm 1–5, tổng còn 33; sau chốt nhóm 7, tổng hiện hành là **5 + 6 + 3 + 9 + 3 + 2 + 4 = 32 bảng**. Upload metadata JSON không tăng bảng.

Các điểm còn mở đều được ghi rõ tại từng nhóm; chúng là nội dung chưa quyết định, không phải quyết định đã chốt bị bỏ sót. Dictionary/DBML/workflow/migration chưa cập nhật trong vòng rà soát này.

## 6. Gợi ý phân công — 2 bảng

### Bản chốt và kết quả review — 15/10/2026

Người dùng đồng ý hai bảng sau phần trình bày, yêu cầu review trước nhóm cuối. Đã đối chiếu dictionary 5B, DBML hiện tại, workflow gợi ý và thiết kế 3C/4C với các quyết định nhóm 1–5. Đây là review tài liệu/quan hệ, chưa chạy DDL hoặc kiểm thử engine.

| Bảng | Trường giữ sau quyết định |
|---|---|
| recommendation_runs — 10 trường | run_id, task_id, generated_by, snapshot_at, generated_at, engine_version, schema_version, params, input_snapshot, result_metadata |
| recommendation_items — 7 trường | item_id, run_id, employee_id, rank_no, matching_score, explanation, score_details |

Giữ engine_version/schema_version vì định danh logic tính/định dạng lưu, khác version chống cập nhật đã bỏ. Giữ hai timestamp với nghĩa lấy dữ liệu/tạo kết quả. Không bổ sung bảng hay cột trong lượt review này.

Kết quả và điều kiện đồng bộ:

1. Quan hệ phù hợp: task 1–n runs, run 1–0..n items, employee 1–n items; assignment dùng đúng run của task và item của người nhận. RECOMMENDED cần run/item; MANUAL từ đầu không có cả hai, MANUAL sau run có run và item NULL. FK ghép cần CHECK NULL theo phương thức, không chỉ FK.
2. Run chỉ được dùng tối đa một lần xác nhận, kể cả manual sau run. Duy nhất người/run và hạng/run, score 0..100, rank dương; liên tục/giới hạn Top-N kiểm tra cả tập. Tạo run và tập items nguyên tử, không công bố kết quả một phần. Run rỗng hợp lệ khác lỗi tính.
3. Không trùng nguồn: params chứa tham số, input_snapshot chứa đầu vào, result_metadata chứa thông tin chung; items chứa kết quả từng người. Không output_snapshot/is_selected/matching_score trên assignment. rank_no giữ vì thứ tự có thể dùng quy tắc phá hòa ngoài điểm hiển thị đã làm tròn.
4. Khi cập nhật tài liệu đích, bỏ run.spec_version theo nhóm 4 và bỏ tham chiếu employee_skills.version/is_effective theo nhóm 2; so giá trị đặc tả/requirements và kỹ năng liên quan với snapshot. Đối chiếu cả thiếu/có kỹ năng, không chỉ các dòng đã tồn tại lúc chạy. Kiểm tra quyền/trạng thái hiện hành trong giao dịch xác nhận.
5. Xóa mềm/sửa log và hiệu chỉnh được duyệt không viết lại run/items cũ. Lần chạy mới dùng số liệu hiện tại nếu công thức cần; không tự mặc định Q/T hoặc effort vào công thức. Draft không tăng tải chính thức; giao việc đã commit phải được tính vào lượt gợi ý sau.
6. Phạm vi snapshot là dữ liệu cần tính/giải thích, không CV gốc/contact. DTO trả người dùng không lộ raw task/log dự án khác. Engine xác định điểm; không nhận điểm/rank client làm dữ liệu chuẩn.
7. Schema JSON phải thống nhất: dùng run.schema_version làm định dạng chung cho params/input_snapshot/result_metadata và score_details của items thuộc run; không duy trì hai version định dạng cạnh tranh. Cấu trúc trường/miền giá trị/sự khớp điểm và chính sách tham số đặc tả khi chốt công thức. Lưu engine_version phải gắn với logic thực được quản lý, không hứa replay chỉ nhờ nhãn.

Kết luận review: không thấy mâu thuẫn cấu trúc buộc thêm bảng hoặc đổi mô hình hai bảng. Các phụ thuộc thiết kế cũ được ghi rõ để sửa khi đồng bộ. Việc công thức/cấu trúc JSON chưa chốt không được báo thành đã kiểm chứng điểm gợi ý hoặc sẵn sàng migration.

<details>
<summary>Phần trình bày ban đầu nhóm 6 — đọc bản chốt phía trên</summary>

Đã đối chiếu dictionary 5B và workflow gợi ý. Giải thích hai trách nhiệm: run lưu bối cảnh một lần gợi ý cho một task, items lưu từng ứng viên trong kết quả. Chưa có quyết định mới của người dùng giữ/bỏ trường nhóm này.

### 6.1. `recommendation_runs`

Trường hiện tại: run_id, task_id, generated_by, snapshot_at, generated_at, engine_version, spec_version, schema_version, params, input_snapshot, result_metadata. Theo quyết định nhóm 4, spec_version phải bỏ/đổi cơ chế liên quan; phương án trình bày bỏ cột, còn 10 trường.

- task_id trỏ tasks: một task có nhiều lần chạy; batch chọn nhiều task tạo run theo từng task, không thêm bảng batch.
- generated_by trỏ users: người chạy phải có trách nhiệm PTCM hiện hành; một tài khoản có thể tạo nhiều run.
- Một run có 0..n items; không có ứng viên là kết quả rỗng hợp lệ, khác lỗi engine.
- Một run được liên kết tối đa một lượt phân công đã xác nhận theo thiết kế hiện tại. Chạy gợi ý/chọn draft chưa tạo assignment.
- params lưu cấu hình đã dùng; input_snapshot lưu đầu vào thực dùng; result_metadata lưu thông tin chung như top_n/số xét/lý do rỗng. Không lưu thêm output_snapshot lặp items.
- engine_version định danh phương pháp/logic tính; schema_version định dạng JSON. Đây không phải phiên bản sửa hồ sơ và không đồng nghĩa chốt công thức cũ.

### 6.2. `recommendation_items`

item_id, run_id, employee_id, rank_no, matching_score, explanation, score_details (7 trường). Mỗi dòng là ứng viên trong một run. run_id trỏ runs, employee_id trỏ employees; một nhân viên có thể xuất hiện trong nhiều run. Duy nhất người/run và hạng/run. Item được chọn liên kết từ assignments đúng run/người; không thêm is_selected, không chép score vào assignment.

score_details là JSON có cấu trúc theo công thức sau chốt, không mặc định skill/workload/performance hoặc Q/T thành các thành phần đã được duyệt. explanation phục vụ giải thích bằng ngôn ngữ, không cần gọi LLM để tạo. Ví dụ điểm dùng khi trình bày là giả lập, không kết quả tính công thức đã kiểm chứng.

### Đồng bộ quyết định trước và các điểm rà tiếp

- Bỏ run.spec_version cùng tasks.spec_version; so sánh đặc tả/requirements hiện tại với snapshot khi confirm. Skill official đã bỏ version: đối chiếu dữ liệu kỹ năng thực dùng, không giữ kiểm tra skill.version không tồn tại.
- Khi confirm kiểm tra lại trách nhiệm PTCM, trạng thái task/dự án/người nhận, kỹ năng và điều kiện hiện hành. Không gọi run cũ là kết quả thời gian thực; tải mới cần hiển thị theo quy tắc đã có, không sửa điểm/rank run cũ.
- Snapshot chỉ dữ liệu tính toán cần thiết, không CV gốc/contact hoặc raw task/log dự án khác trả ra màn hình. Phần dữ liệu replay và DTO hiển thị có phạm vi khác nhau.
- Chạy lại tạo run mới, giữ run/items cũ. Manual từ đầu không tạo run/score giả; manual sau run liên kết run nhưng item NULL theo nhóm 4.
- Chờ trao đổi các trường engine/schema version, params/snapshot/metadata/score_details và thời điểm lấy dữ liệu/tạo kết quả để giữ đúng nhu cầu, chưa tự thêm cấu hình công ty hoặc công thức cũ vào thiết kế cuối.

</details>

## 7. Phiên đăng nhập, chống ghi lặp, lịch sử kiểm toán và thông báo — 4 bảng sau giản lược (ban đầu 5)

### Bản chốt hiện hành — 15/10/2026

Người dùng đã đồng ý phương án nghiên cứu nhóm 7. Quyết định này thay phương án JWT + refresh rotation cũ và các ghi chú chưa chốt phía dưới.

| Bảng | Trường sau chốt |
|---|---|
| auth_sessions — 6 trường | session_id, user_id, token_hash, created_at, expires_at, revoked_at |
| idempotency_records — 8 trường | idempotency_id, actor_id, operation, idempotency_key, request_hash, result_reference, created_at, expires_at |
| audit_events — 9 trường | audit_id, actor_id, entity_type, entity_id, operation, occurred_at, reason, before_values, after_values |
| notifications — 8 trường | notification_id, recipient_id, notification_type, resource_type, resource_id, message, created_at, read_at |

- auth_sessions dùng session phía server và cookie secret ngẫu nhiên; session_id là ID nội bộ, token_hash duy nhất để tra cứu/đối chiếu. Không cookie ID số tăng dần hoặc secret rõ trong DB. Mỗi user có nhiều phiên; kiểm tra quyền/status hiện hành và thu hồi khi logout/khóa/reset/đổi role theo chính sách hiện có.
- Bỏ refresh_tokens và luồng access JWT/refresh của phiên bản hiện tại; không chỉ xóa bảng mà giữ mô tả rotation/reuse. Chưa cài thư viện session. Giữ bảo vệ cookie/CSRF/Origin; chốt hạn tuyệt đối/inactivity và cleanup khi thiết kế triển khai, không coi 7 ngày là chuẩn nguồn ngoài.
- Giữ idempotency_records dùng chọn lọc cho mutation dễ tạo dòng trùng, đặc biệt tạo/bổ sung nhật ký; không mọi API. Key/hash/kết quả và nghiệp vụ cùng transaction. Cùng key khác nội dung báo conflict; kết quả dựng lại theo quyền hiện hành. Retention/cửa sổ retry chưa đặt tùy ý; hết hạn không tự hứa chống lặp vô hạn.
- Giữ audit_events cho khóa/role, thay PM/PTCM, HR cập nhật trực tiếp và đổi CV hiện tại khi chưa có history riêng đủ. Bỏ actor_kind, actor_id bắt buộc FK users trong phạm vi chỉ thao tác con người. Không sao chép history/decision miền, không ghi credential/token/CV thô. Log worker/lỗi vận hành nằm ngoài audit nghiệp vụ này.
- Giữ notifications trong app, recipient_id trỏ users. resource_type/resource_id đa loại do service kiểm tra; mở tài nguyên phải kiểm tra quyền. PTCM nhận yêu cầu hiệu chỉnh theo nhóm 5, không giữ PM như người duyệt. Không thêm email/Slack integration.
- Tổng đích **32 bảng**, nhóm 7 có 4; scope auth đã chốt nhưng chi tiết thời hạn, contract idempotency/audit và các điểm mở của nhóm khác chưa tự trở thành đã hoàn thiện.

Ảnh hưởng cần đồng bộ trong đợt sửa tài liệu sau: dictionary/DBML, security, roles_permissions, architecture, api_design, CLAUDE.md, testing_verification và các mô tả JWT_SECRET/access/refresh cookie/refresh endpoint cũ. Việc xác thực/thu hồi trong giao dịch vẫn cần kiểm thử khi triển khai, không chỉ đổi tên bảng.

<details>
<summary>Lịch sử trình bày/nghiên cứu nhóm 7 — đọc bản chốt hiện hành phía trên</summary>

Đã đọc dictionary 5B/security và đối chiếu OWASP Session Management, RFC 9700 về nguyên tắc phiên/rotation. Các nguồn không bắt buộc schema 5 bảng; đây là lựa chọn thiết kế của hệ thống. Chưa có quyết định mới của người dùng giữ/bỏ nhóm 7.

### 7.1. `auth_sessions`

session_id, user_id, created_at, expires_at, revoked_at. Một phiên đăng nhập thuộc một user, user có nhiều phiên; phiên có nhiều refresh_tokens theo vòng xoay. Thu hồi phiên khi logout/khóa/reset/đổi role theo chính sách hiện tại, kiểm tra trạng thái/quyền hiện hành. Đề xuất giữ nếu tiếp tục phương án JWT + refresh rotation + session DB đã có; thời hạn 7 ngày là cấu hình thiết kế hiện tại, không chuẩn bắt buộc.

### 7.2. `refresh_tokens`

token_id, session_id, token_hash, created_at, expires_at, consumed_at, revoked_at, replacement_id. Hash token để đối chiếu, không lưu token rõ. replacement_id tự tham chiếu token kế tiếp cùng session. Phân biệt token đã đổi sang token mới và bị thu hồi; chuỗi dùng phát hiện reuse. Không ghi access JWT từng cái vào DB. Đề xuất giữ theo cơ chế auth hiện có, không mặc định mọi ứng dụng phải có bảng này.

### 7.3. `idempotency_records`

idempotency_id, actor_id, operation, idempotency_key, request_hash, result_reference, created_at, expires_at. User 1–n records; result_reference JSON tham chiếu đa loại, không FK tổng quát. Cùng actor/operation/key chỉ một bản ghi; cùng key khác payload báo xung đột, không dùng lại kết quả sai. Ghi record/kết quả cùng transaction nghiệp vụ và kiểm tra quyền khi retry. Không áp dụng cho mọi GET/login/refresh; dùng thao tác ghi cần chống thực thi lặp. TTL/lưu trữ còn chờ, không tự đặt.

Ví dụ chính: thêm nhật ký đã commit nhưng client chưa nhận phản hồi, retry cùng khóa lấy kết quả cũ thay vì tạo log mới. Khác ràng buộc một quyết định/run hoặc một Pending/log, vốn không phân biệt hai khoản giờ giống nhau có thật.

### 7.4. `audit_events`

audit_id, actor_kind USER/SYSTEM, actor_id nullable, entity_type, entity_id, operation, occurred_at, reason, before_values, after_values. Actor USER có FK users, SYSTEM không giả user. Entity đa loại không FK trực tiếp; service kiểm tra. Dùng cho đổi quyền/khóa tài khoản, thay PM/PTCM, sửa profile trực tiếp/chọn CV khi bảng miền chưa có lịch sử đủ; không sao chép mọi history/decision sang audit. Không password/hash/token/CV thô. actor_kind SYSTEM cần rà nếu chỉ chọn ghi thao tác người dùng.

### 7.5. `notifications`

notification_id, recipient_id, notification_type, resource_type, resource_id, message, created_at, read_at. User 1–n notifications; resource đa loại kiểm tra bằng service, không FK tổng quát. Thông báo trong app, không Slack/email queue. Nhắc đúng PTCM khi có yêu cầu hiệu chỉnh theo nhóm 5, nhân viên nhận kết quả; HR/Giám đốc theo luồng tương ứng. Đã đọc không là đã duyệt, mở link phải kiểm tra quyền hiện hành.

### Kết luận trình bày và điểm rà tiếp

Đề xuất giữ 5 bảng theo phương án hiện tại vì trách nhiệm khác nhau; không đánh đồng bảng auth/idempotency với chức năng nghiệp vụ mới. Auth_sessions/refresh_tokens phụ thuộc phương án auth, không được bỏ rời rạc mà giữ nguyên yêu cầu rotation/thu hồi. Idempotency_record là bảo vệ retry, không phải lịch sử audit. Audit bổ sung phần thiếu ở bảng miền, tránh ghi trùng. Notifications phục vụ UI thông báo hiện có, không thêm tích hợp ngoài.

Chờ người dùng đánh giá các trường, thời hạn/retention, phạm vi audit SYSTEM và mức cần thiết của từng cơ chế. Không tự tăng bảng hoặc coi đã sẵn sàng migration.

### Kết quả nghiên cứu tiếp nhóm 7 — chờ chốt phương án

Người dùng **đã đồng ý giữ auth_sessions và notifications**, yêu cầu nghiên cứu sâu refresh_tokens/idempotency_records/audit_events. Đã nghiên cứu và lưu [bản phân tích nguồn, so sánh và phương án](database-group7-research-2026-10-15.md).

- Refresh API không đồng nghĩa bắt buộc bảng token; nếu giữ JWT + refresh rotation thì vẫn cần trạng thái phục vụ rotation/reuse. Phương án mới đề xuất cho web cùng backend: server-side session/cookie secret, bỏ access JWT/refresh endpoint và refresh_tokens, thêm token_hash vào auth_sessions (6 trường dự kiến). Đây là thay đổi chiến lược auth cũ, **chưa được người dùng chốt**.
- Idempotency: có thể dùng client request ID trên bảng đích nhưng phải kiểm tra payload/phạm vi và dữ liệu entity đã sửa; không chỉ disable nút hoặc unique nội dung log. Đề xuất giữ bảng chung cho mutation dễ tạo trùng, không áp dụng mọi API; TTL/retention chưa tự chọn.
- Audit: đề xuất giữ bảng chung chỉ cho lịch sử chưa có ở bảng miền (role/khóa, thay PM/PTCM, HR sửa trực tiếp, đổi CV hiện tại). Bỏ actor_kind trong phạm vi chỉ thao tác con người, actor_id bắt buộc; 9 trường dự kiến. Không sao chép history/decision đã có, không biến log lỗi worker thành audit nghiệp vụ.
- Nếu chốt session mới và giữ hai bảng kỹ thuật kia: tổng 32 bảng, nhóm 7 còn 4. **Hiện tổng đã chốt vẫn 33**; không tự bỏ bảng/token hoặc cập nhật tài liệu auth trước đồng ý.
- Đã dùng skill auth-implementation-patterns để rà lựa chọn xác thực và vòng đời thông tin; nguồn chính: Django sessions, Go SCS, OWASP, RFC 9700/9110, AWS Builders' Library, Stripe và Django Admin LogEntry. Không cài thư viện hoặc thay code.

</details>

## Kế hoạch cập nhật sau khi chốt các nhóm

1. Dùng quyết định hiện hành trong file này; các lựa chọn C01–C06/upload đã chốt, không hỏi lại từ phần lịch sử. Liệt kê riêng chi tiết kỹ thuật còn cần đặc tả.
2. Cập nhật dictionary 5A/5B, bản đề xuất tổng hợp và DBML.
3. Đồng bộ workflow nghiệp vụ, thiết kế API, bản đồ migration và tài liệu trình bày có liên quan.
4. Rà toàn bộ tham chiếu trường bị bỏ/đổi, khóa ngoại, quy tắc duyệt và số lượng bảng/trường; phân biệt bản hiện hành với tài liệu lịch sử.
5. Ghi nhận kết quả kiểm tra và giới hạn còn lại. Việc cập nhật tài liệu không đồng nghĩa đã triển khai SQL/migration.

Trong vòng rà soát theo nhóm, chỉ file ghi nhận này được tạo và cập nhật liên tục. Những tài liệu khác chưa được cập nhật theo các quyết định của vòng rà soát này.

## Bàn giao cho chat cập nhật tài liệu và thiết kế DB

### 1. Nguồn chuẩn và phạm vi

Đối với địa chỉ, đọc quyết định tiếp nối 20/10 ở đầu file và baseline được dẫn: address nullable lúc tạo, phải bổ sung khi hoàn thiện, không xóa trắng; allowlist fullname/phone/address. Không dùng bản chốt C03 hai trường của ngày 15/10 để bỏ địa chỉ hoặc ép NOT NULL ngay khi tạo.

File này là **đầu vào quyết định thay đổi**, không phải từ điển đầy đủ của mọi cột. Đọc cùng dictionary 5A/5B và DBML để giữ các phần không thay đổi. Ưu tiên quyết định hiện hành ở đầu mỗi nhóm và phần chốt C01–C06/upload; không dùng nội dung trong `<details>` hoặc schema cũ để khôi phục trường/bảng/quyền đã bỏ.

Đọc `CLAUDE.md`, baseline nghiệp vụ 12–13/10 và workflow liên quan trước sửa. Khi mâu thuẫn, quyết định mới cụ thể trong file này thay đúng nội dung cũ tương ứng; các quy tắc nghiệp vụ khác tiếp tục giữ. Không chỉnh tài liệu gốc trong `docs/thesis-context/`.

Đợt tiếp theo là đồng bộ **tài liệu và schema thiết kế**. Không coi đó là quyền chạy migration lên Neon, reset DB, chuyển đổi dữ liệu thật, cài thư viện hoặc triển khai auth/upload. Nếu cần tạo SQL migration, phải kiểm kê hiện trạng và kiểm thử trên DB thử nghiệm; không sửa/xóa migration đã áp dụng để giả rằng schema mới đã có từ đầu.

### 2. Danh mục đích — đúng 32 bảng

| Nhóm | Các bảng giữ |
|---|---|
| 1 — 5 | users, user_roles, employees, departments, positions |
| 2 — 6 | skill_categories, skills, employee_skills, employee_skill_change_requests, employee_profile_change_requests, skill_request_task_evidence |
| 3 — 3 | employee_cvs, cv_extractions, cv_extracted_skills |
| 4 — 9 | projects, project_status_history, task_groups, tasks, task_requirements, assignments, task_submissions, task_status_history, task_comments |
| 5 — 3 | work_logs, work_log_history, work_log_change_requests |
| 6 — 2 | recommendation_runs, recommendation_items |
| 7 — 4 | auth_sessions, idempotency_records, audit_events, notifications |

Bỏ assignment_events và refresh_tokens so với bản 34 bảng. Không khôi phục project_members/đội dự án từ schema legacy. Không thêm skill_request_attachments: metadata nằm trên request.

### 3. Ranh giới nghiệp vụ không được làm lệch

- PTCM phân công thủ công/theo gợi ý và nghiệm thu; không thực hiện task trong dự án phụ trách. Không có người nghiệm thu riêng, không giao trách nhiệm review riêng, không tự nghiệm thu. PM chỉ định PTCM, bắt đầu/theo dõi/hoàn thành/đóng; Giám đốc tạo dự án kèm chọn PM và có quyền thay PM.
- PM và PTCM có thể cùng người. Gợi ý trong nguồn nhân lực toàn công ty đủ điều kiện, không lập đội trước, không giới hạn cứng hai dự án. Kế hoạch Word/Slack chia sẻ ngoài hệ thống, không thêm bảng plan/sprint hoặc quy trình PM duyệt từng lần phân công.
- PM/PTCM tra cứu giới hạn thông tin nghề nghiệp, kinh nghiệm/kết quả, kỹ năng và tải tổng hợp; nhân viên chỉ tra cứu trong trách nhiệm PTCM. Không cấp đọc CV gốc/contact/nhật ký hoặc nội dung task dự án khác từ chức năng tra cứu.
- Nhân viên tự ghi và gửi hiệu chỉnh nhật ký; PTCM hiện hành duyệt sau khóa, không PM duyệt/ghi thay. HR duyệt thông tin cá nhân; Giám đốc duyệt kỹ năng có căn cứ và không tự duyệt. ADMIN không mặc nhiên có quyền nghiệp vụ của các vai trò khác.
- Gợi ý không tự ghi phân công; draft không tăng tải, commit phải cập nhật nguồn tải cho lượt gợi ý sau. Giao lại giữ lượt cũ, tạo lượt mới; dữ liệu hiện tại, lịch sử và snapshot có trách nhiệm khác nhau.
- Không tự thay công thức Q/T, đưa Q/T vào điểm phân công, khôi phục công thức cũ hoặc tuyên bố công thức đã được kiểm chứng chỉ từ việc chốt cấu trúc DB.

### 4. Điểm kỹ thuật cần giải quyết khi hợp nhất — không được che bằng tuyên bố “đã hoàn tất”

| Điểm | Yêu cầu khi hoàn thiện thiết kế |
|---|---|
| ADD nhật ký qua yêu cầu | Đã được người dùng đồng ý thêm result_log_id nullable ngày 15/10, chỉ có cho ADD APPROVED. FK ghép với assignment và unique log kết quả; target/before vẫn NULL. Tạo log, liên kết và quyết định nguyên tử dưới khóa PENDING; không history ADD, không áp dụng lặp. Request hiện có 14 trường. |
| Kết thúc lượt giao | Xác định đầy đủ khi nào ghi ended_at/ended_by/end_reason (giao lại, thu hồi, hoàn tất, hủy), quan hệ với current_assignee_id và truy vấn tải. Không dùng “chưa ended” làm nguồn tải duy nhất, không mất quyền xem/hiệu chỉnh đóng góp lịch sử. |
| FK sau bỏ trường lặp | Work logs/requests suy task/người từ assignment; sửa FK ghép/index cũ để không trỏ cột đã bỏ. UPDATE/DELETE phải trỏ log cùng assignment; evidence submission cùng assignment và đúng nhân viên yêu cầu. |
| Snapshot/before_values | Bỏ tham chiếu version/spec_version đã xóa; định nghĩa tập trường so sánh dưới khóa, kiểm tra cả sự vắng mặt của kỹ năng. Snapshot bất biến; không lấy updated_at làm bảo vệ màn hình cũ. |
| CV/API và upload | Chọn provider/model/cách chạy worker, mapping trùng/mục chưa map trong giới hạn không mở lại quyết định CV. Xác định metadata JSON, quyền xem private, loại/dung lượng/số tệp và cleanup tệp chưa gắn. Không tự bỏ các cột retry/lease khi chưa có phương án chạy API. |
| Hủy khi đang review | Không gán giả “không đạt”, không sinh thêm phân hệ ngoại lệ. Ghi rõ cách kết thúc an toàn trong trạng thái/giao dịch hiện có; không tự đổi tập tính Q/T. |
| Công thức và JSON engine | Miền weight/chuẩn hóa, params/input/score_details và phiên bản logic cần đặc tả; cấu trúc hai bảng đã chốt không thay việc chốt công thức. |
| Session/idempotency | Thiết kế timeout tuyệt đối/idle, reset/revoke, cookie/CSRF, endpoint dùng key và retention. Không khôi phục refresh/JWT. Nếu idle timeout cần dữ liệu chưa có ở bộ sáu cột thì trình bày giải pháp, không âm thầm thêm cột. |
| Chuyển đổi legacy | Phân biệt users ID với employees ID, giờ thập phân với phút, email cá nhân/công ty, trạng thái/lượt giao cũ. Không dựng lịch sử/actor giả từ updated_at hoặc trạng thái hiện tại. |

Các điểm trên không mở lại quyết định nghiệp vụ hoặc buộc thêm bảng. Chat tiếp theo được hoàn thiện chi tiết trong phạm vi đã chốt; nếu thực sự cần thay cấu trúc đã chốt hoặc mở rộng quyền/luồng thì trình bày tác động và hỏi người dùng trước.

### 5. Tài liệu cần đồng bộ và kiểm tra bàn giao

1. Cập nhật `database-step5a-core-dictionary-2026-10-14.md`, `database-step5b-supporting-dictionary-2026-10-14.md`, `database-proposal-2026-10-14.md` và `.dbml`: 32 bảng, đầy đủ cột/kiểu/NULL/default/PK/FK/UNIQUE/CHECK/index và quan hệ. Phân bổ lại 5A/5B theo bảng thực giữ, không duy trì nhãn 12+22 cũ. Không đoán tổng số cột/FK trước kiểm đếm.
2. Đồng bộ `.claude/docs/`: database_schema, domain_rules, roles_permissions, employee_profile_workflow, skill_approval_workflow, cv_extraction_approval_workflow, project_task_assignment_workflow, work_effort_workflow, ai_recommendation_workflow, security, architecture, api_design, product_scope, testing_verification; cập nhật chỉ mục CLAUDE/README nếu mô tả hiện hành bị ảnh hưởng.
3. Cập nhật bản đồ chuyển đổi `database-step5c-migration-map-2026-10-14.md` và bản trình bày `database-pitch-2026-10-14.md`. Tài liệu review/baseline lịch sử giữ nguyên sự kiện cũ hoặc thêm nhãn/link thay thế, không viết lại thành kết quả kiểm thử schema mới.
4. Rà tham chiếu trường/bảng/quyền đã bỏ **trong tài liệu hiện hành**. Không thay thế chuỗi hàng loạt trong lịch sử, không xóa request.version/engine_version/schema_version/parser_version chỉ vì cùng có từ version. Không xóa source ở request hoặc status của các bảng khác vì bỏ skills.status.
5. Kiểm đếm bảng, kiểm tra link Markdown/FK đích và sơ đồ; review xuyên luồng tạo dự án → giao việc → nộp/nghiệm thu → hiệu chỉnh → hoàn thành/đóng, và yêu cầu kỹ năng/CV/upload → duyệt → official. Ghi rõ công cụ nào đã chạy, kiểm tra nào chỉ tĩnh, phần nào chưa có DB thử nghiệm/parser DBML; không tuyên bố render/DDL/migration đã thành công nếu chưa chạy.

### 6. Kết quả rà soát bản ghi nhận này

Đã thống nhất trạng thái C01–C06/upload ở phần hiện hành, đưa trao đổi cũ vào phần lịch sử, sửa số tổng còn sót và bổ sung danh mục 32 bảng cùng hướng dẫn bàn giao. File đủ làm đầu vào cập nhật tài liệu; **chưa phải schema SQL cuối đã kiểm thử**. Những điểm kỹ thuật ở mục 4 phải được báo lại trong đầu ra đợt tiếp theo, không bỏ qua hoặc che dưới chữ “giữ nguyên”.
