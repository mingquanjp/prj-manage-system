# Quy tắc nghiệp vụ — baseline hiện hành

**Chốt tiếp nối 27/10/2026 — đề nghị kỹ năng của nhân viên:** luồng hiện tại chỉ bổ sung hoặc điều chỉnh; suy ADD khi chưa có official, UPDATE khi đã có, không cho chọn loại hoặc gửi đề nghị gỡ. Không thêm quyền gỡ trực tiếp cho Giám đốc/HR. Giá trị REMOVE và mô tả xử lý bản ghi cũ được giữ để tương thích thiết kế/lịch sử, không là chức năng tạo đề nghị mới. Không đổi cấu trúc DB; việc thu hồi kỹ năng ngoài luồng này chưa triển khai. Chốt này ưu tiên các mô tả ADD/UPDATE/REMOVE cũ đúng phần thao tác của nhân viên.

**Đồng bộ quyết định 15/10/2026:** [sổ quyết định, phần hiện hành và Bàn giao](../../docs/design/database-review-decisions-2026-10-14.md) ưu tiên cho nội dung đã thay thế; [dictionary/DBML hợp nhất](../../docs/design/database-proposal-2026-10-14.md) là thiết kế 33 bảng (cấp bậc theo chốt 26/10/2026). Đây là thiết kế DB, chưa triển khai DDL/migration hoặc backend nghiệp vụ; tiến độ frontend mô phỏng xem [sổ công việc](../../web/README.md#sổ-công-việc).


> Baseline nghiệp vụ hiện hành: [2026-10-12](../../docs/business-baseline-2026-10-12.md), PTCM phân công và nghiệm thu, không nhận task trong dự án phụ trách; tra cứu giới hạn dữ liệu nghề nghiệp. Phần kỹ thuật V2 chưa đồng bộ không được ghi đè quyết định mới; công thức và quy tắc không bị thay thế giữ nguyên.

Rà soát cách diễn đạt ngày 2026-10-16 theo các quyết định hiện hành dẫn ở đầu file. Mốc 2026-09-06 và [baseline V2](../../docs/reviews/system-design-baseline-v2.md) là nguồn lịch sử, không thay thế quyết định hiện hành.
Schema 000001 chưa đáp ứng toàn bộ; xem `database_schema.md`.

## 1. Bất biến

- R1: Engine chỉ gợi ý; chỉ PTCM hiện hành của dự án được xác nhận phân công.
- R2: Task là đơn vị phân công duy nhất. Group chỉ gom nhóm.
- R3: Tra cứu và engine chỉ dùng kỹ năng chính thức đã duyệt, còn hiệu lực.
- R4: Một batch hoặc ghi toàn bộ hoặc không ghi gì; draft không đặt chỗ.
- R5: Mọi thay đổi trạng thái/người phụ trách/năng lực giữ audit; không API xóa vật lý lịch sử.
- R6: Cho phép nhiều task IN_PROGRESS; trạng thái không sinh giờ. Effort ghi theo ngày/phút/nội dung, không phân bổ trùng cùng thời gian.
- R7: Quyền, giá trị nền và điều kiện cuối phải kiểm tra trong transaction có kiểm soát cạnh tranh.
- R8: PTCM không nhận/thực hiện task trong dự án phụ trách. Manual, gợi ý, confirm và reassign đều áp dụng; không override điều kiện này. Khi chỉ định/thay PTCM, chặn người còn nhận task ASSIGNED/IN_PROGRESS/PAUSED/PENDING_REVIEW trong cùng dự án; chuyển giao trước, giữ lịch sử.
- R9: PTCM không nghiệm thu đầu ra mình đã thực hiện, kể cả trước khi chuyển giao. Chỉ định PTCM phải tránh xung đột đầu ra cũ còn cần nghiệm thu; không chỉ so current_assignee. Không có reviewer thay thế. Xem giới hạn xử lý trong baseline hiện hành.

## 2. Tài khoản và năng lực

Admin tạo user + employee ban đầu trong một transaction theo [quyết định 20/10/2026](../../docs/business-baseline-2026-10-20.md); HR quản lý hồ sơ về sau. Quyền tạo không cấp cho Admin quyền cập nhật hành chính thường xuyên. Cả user và employee phải ACTIVE để sử dụng API nghiệp vụ.
Người thực hiện phải có tài khoản và hồ sơ active, không là PTCM của dự án đang phân công. PM kiêm PTCM cũng không nhận task trong dự án đó. Trách nhiệm PTCM ở dự án khác không tự loại ứng viên. Không hạ quyền/vô hiệu hóa ADMIN active cuối cùng.
Thêm vai trò hệ thống không bị chặn chỉ vì còn công việc. Chỉ chặn thu hồi quyền nếu khiến người dùng không còn quyền xử lý công việc/trách nhiệm đang mở; đổi vai trò hệ thống không tự thay trách nhiệm dự án. HR chưa được vô hiệu hóa hồ sơ khi còn công việc/trách nhiệm dự án mở; hệ thống hiển thị các mục cần xử lý trước. Admin vẫn được khóa khẩn cấp và liệt kê việc cần xử lý, không tự bàn giao.

Giám đốc thay PM bằng thao tác chỉnh sửa người quản lý dự án: chọn PM mới đủ điều kiện và lưu một lần. Lưu thành công thì PM mới thay PM cũ; nếu lưu thất bại, PM cũ vẫn giữ trách nhiệm. Không yêu cầu xóa PM cũ trước, không có quy trình bàn giao riêng hoặc chờ PM cũ xác nhận. Giữ lịch sử thay đổi; không thay đổi trạng thái dự án, phân công công việc hoặc đóng góp đã có. PM mới chỉ có quyền hợp lệ theo trạng thái hiện hành; các trách nhiệm khác của PM cũ được xét độc lập. Với COMPLETED, PM mới đóng project khi đủ điều kiện; PTCM hiện hành xử lý correction; không nghiệm thu thêm, tạo task, phân công/thay người, mở lại execution hoặc đưa project về IN_PROGRESS.

`employee_skills` unique người/kỹ năng chỉ lưu năng lực hiện đang công nhận, đầy đủ người/ngày duyệt. Không source/is_effective/version; REMOVE gỡ dòng hiện tại, giữ request/căn cứ, ADD sau đó tạo dòng mới.
`employee_skill_change_requests`: ADD/UPDATE/REMOVE, MANUAL/CV; PENDING → APPROVED/REJECTED/WITHDRAWN/NO_CHANGE. NO_CHANGE chỉ áp dụng theo điều kiện không thay đổi giá trị nghiệp vụ chính thức nêu bên dưới; không đồng nghĩa được phê duyệt cập nhật.
Hai luồng độc lập: hệ thống trích xuất CV để Giám đốc đối chiếu kỹ năng/mức có căn cứ; hoặc nhân viên chủ động gửi đề nghị cập nhật, không cần CV mới. Dùng chung cơ chế xét duyệt không đồng nghĩa bắt nhân viên gửi thủ công từng mục CV. CV mới không tự ghi đè/xóa kỹ năng chính thức; mục không thay đổi không cần duyệt lại chỉ vì đổi CV.
Một pending/người/kỹ năng; duyệt kiểm tra request.version và giá trị official so before_values dưới khóa, ghi quyết định và cập nhật chính thức atomically.
Pending không thay giá trị đang dùng. Phê duyệt và từ chối kỹ năng đều bắt buộc lý do; đề nghị sửa sai/REMOVE phải nêu nội dung và lý do. Giám đốc nhập trực tiếp cũng qua service này với audit, không được duyệt kỹ năng của chính mình.
Quy trình chuẩn: [Phê duyệt cập nhật kỹ năng](skill_approval_workflow.md). Bổ sung/nâng mức cần giải trình; tệp hoặc liên kết không bắt buộc. Giám đốc đối chiếu đóng góp cá nhân và mô tả mức; nếu không đủ chuyên môn thì cần ý kiến chuyên môn phù hợp trước khi duyệt. Không tự xác nhận năng lực chỉ từ CV, chứng chỉ, số năm, task đã nghiệm thu hoặc điểm gợi ý.
Cho bổ sung căn cứ khi PENDING, không thêm trạng thái hoặc buộc gửi lại chỉ vì bổ sung minh chứng. Không đặt hạn ngạch tuần/tháng; sau từ chối gửi yêu cầu mới với căn cứ làm rõ/bổ sung. Giữ kiểm tra phiên bản và không mở lại quyết định đã kết thúc.
R04: kỹ năng ghi nhận sai được điều chỉnh hoặc loại khỏi hồ sơ qua xét duyệt, có lý do và giữ lịch sử. Có thể đề nghị bổ sung lại bằng yêu cầu mới khi có căn cứ; không có chức năng tái kích hoạt riêng. Cách lưu định danh/phiên bản thuộc thiết kế dữ liệu.
NO_CHANGE chỉ khi toàn bộ giá trị nghiệp vụ được đề nghị giống bản chính thức còn hiệu lực; không so mỗi level hoặc bỏ qua trạng thái thu hồi. Yêu cầu stale không ghi đè; reject/withdraw rồi gửi mới.
R03: yêu cầu đã kết thúc giữ nguyên quyết định; bổ sung bằng yêu cầu kỹ năng mới, có thể dùng CV cũ làm minh chứng. Không mở lại mục trích xuất đã xử lý; xem cv_extraction_approval_workflow.md.

Yêu cầu cập nhật thông tin cá nhân do nhân viên gửi, HR xét duyệt. Pending chưa đổi hồ sơ chính thức; chấp thuận cập nhật đúng nội dung được duyệt, từ chối giữ nguyên hồ sơ và có lý do. Giữ lịch sử, kiểm tra quyền/giá trị nền và ràng buộc dữ liệu áp dụng, không ghi đè từ yêu cầu lỗi thời.
Nhân viên đề nghị sửa fullname/phone/address; HR sửa trực tiếp email công ty/cơ cấu và quản trị departments/positions/job_levels; Giám đốc quản trị skill_categories/skills. Không nhận role, trạng thái, thông tin tổ chức, trách nhiệm dự án hoặc kỹ năng chính thức qua yêu cầu cá nhân. Không mặc định mọi trường cần duyệt hoặc tự chọn trường được sửa trực tiếp.

Địa chỉ cần bổ sung khi hoàn thiện hồ sơ, không bắt nhập lúc Admin tạo tài khoản; employees.address nullable cho giai đoạn chưa bổ sung. Khi cung cấp/sửa phải có nội dung sau trim; không xóa địa chỉ đã có về NULL/rỗng. HR được duyệt đề nghị hoặc cập nhật trực tiếp có lịch sử; patch bỏ qua address là không đổi. Không thêm trạng thái hồ sơ, hạn bổ sung hoặc chặn đăng nhập/nhận việc/thực hiện vì thiếu địa chỉ; UI nhắc bổ sung. Không trả địa chỉ trong tra cứu nghề nghiệp/ứng viên/snapshot gợi ý. Chi tiết theo [quy trình hồ sơ](employee_profile_workflow.md) và quyết định 20/10.
HR quản lý hồ sơ hành chính/CV, được cập nhật trực tiếp theo quyền nhưng không sửa kỹ năng chính thức để bỏ qua xét duyệt. Giám đốc xét duyệt kỹ năng, không thêm tầng HR tiền kiểm. Quản trị viên quản lý tài khoản. Vô hiệu hóa hồ sơ (HR) khác khóa tài khoản (Admin); giữ điều kiện active và lịch sử, không tự bàn giao công việc. Xem [quy trình hồ sơ](employee_profile_workflow.md).

Rubric BASIC: có hướng dẫn; INTERMEDIATE: độc lập việc thường gặp; ADVANCED: xử lý phức tạp; EXPERT: thiết kế/định hướng/hỗ trợ người khác.
Không tự suy mức từ số năm hoặc từ tên kỹ năng trong CV.
Đổi năng lực chính thức sau phân công không tự dừng task hoặc bắt PTCM duyệt lại; áp dụng cho tra cứu/gợi ý/phân công sau.

## 3. Dự án và task

Project code unique toàn hệ thống; ngày kết thúc ≥ ngày bắt đầu. Giám đốc khởi tạo kèm chọn PM trong cùng giao dịch; người tạo không mặc nhiên là PM.
Mỗi task bắt buộc thuộc một group cùng project, mã unique trong project, estimate > 0, ít nhất một requirement.
Tạo task và requirements cùng transaction. Requirement unique task/skill, weight do PTCM đặt theo task trước giao, miền giá trị/chuẩn hóa chờ công thức (không chốt 1–5); required level bốn mức.
Ngày task hoặc cùng trống hoặc cùng có, đúng thứ tự và trong ngày project; sửa ngày project không làm task ra ngoài.
Group là gói công việc/module, được phép trống khi lập kế hoạch; dự án nhỏ có thể dùng một nhóm chung. Không label/cây task nhiều cấp.
Task gộp đầu ra trong description, giữ acceptance_criteria; phải đủ rõ trước khi giao; một người chịu trách nhiệm. Group chỉ tổng hợp từ task.
Không chuyển task sang project khác. Chuyển group trong cùng project có audit, chỉ task chưa kết thúc.

| Project transition | Điều kiện |
|---|---|
| Tạo → PLANNING | Giám đốc khởi tạo và giao PM |
| PLANNING → IN_PROGRESS | PM, có ít nhất một task ASSIGNED |
| IN_PROGRESS → COMPLETED | PM xác nhận, có ≥1 DONE, mọi task còn lại DONE/CANCELLED |
| COMPLETED → CLOSED | PM; DONE → CLOSED cùng transaction |
| PLANNING/IN_PROGRESS → CANCELLED | PM có lý do; hủy task chưa DONE, giữ DONE và effort; kết thúc yêu cầu sửa đang chờ có lý do |

PLANNING cho phân rã/phân công nhưng không start task; IN_PROGRESS cho thực thi.
COMPLETED cho sửa sai effort qua yêu cầu có duyệt, không làm thêm việc. Đóng project chỉ khi không còn yêu cầu effort PENDING.
CLOSED/CANCELLED khóa dữ liệu nghiệp vụ, không mở lại; vẫn được đọc đúng quyền/đánh dấu notification.
Tiến độ = số DONE/CLOSED chia số task không CANCELLED; mẫu số 0 hiển thị chưa có task hiệu lực, không tự hoàn thành project.

| Task transition | Actor và tác động |
|---|---|
| Tạo → PENDING_ASSIGNMENT | Phụ trách chuyên môn, đủ dữ liệu bắt buộc, chưa có người thực hiện |
| PENDING_ASSIGNMENT → ASSIGNED | PTCM confirm; người nhận không là PTCM của dự án |
| ASSIGNED/PAUSED → IN_PROGRESS | Người đang được giao; project IN_PROGRESS; không sinh giờ |
| IN_PROGRESS → PAUSED | Người được giao, lý do; tạm dừng công việc, không phải chấm giờ nghỉ |
| IN_PROGRESS → PENDING_REVIEW | Người được giao nộp kết quả, không bắt buộc đủ nhật ký mới được nộp (R02), lưu bộ nhật ký tại lần nộp |
| PENDING_REVIEW → IN_PROGRESS | PTCM yêu cầu chỉnh sửa có lý do; người thực hiện sửa và nộp lại |
| PENDING_REVIEW → DONE | PTCM hợp lệ, khác người thực hiện, nghiệm thu đầu ra, lưu accepted_at; thiếu nhật ký không tự động chặn nghiệm thu, dữ liệu thời gian thiếu/không tin cậy không dùng cho thành phần hiệu suất phụ thuộc thời gian |
| DONE → CLOSED | Cùng lúc đóng project |
| PENDING_ASSIGNMENT/ASSIGNED/IN_PROGRESS/PAUSED/PENDING_REVIEW → CANCELLED | PM có lý do; giữ toàn bộ effort/lịch sử |

PTCM thực hiện thay người/thu hồi. Thu hồi: ASSIGNED/IN_PROGRESS/PAUSED → PENDING_ASSIGNMENT, xóa current assignee ngay khi thu hồi thành công. A không còn phụ trách; không tìm được người mới thì task vẫn PENDING_ASSIGNMENT, không tự khôi phục A.
Thay người trực tiếp chỉ khi project PLANNING/IN_PROGRESS: các trạng thái task trên → ASSIGNED với assignment mới trong một thao tác nguyên tử. A giữ phân công đến khi thay bằng B thành công; thất bại không làm đổi phân công của A hoặc trạng thái task. Thành công kết thúc trách nhiệm A, giao B và B phải Start trước khi tiếp tục. Giữ toàn bộ kết quả/log/đóng góp lịch sử của A, không chuyển sang B. Nếu đã bắt đầu, có ghi chú bàn giao kết quả/phần còn lại/vị trí tài liệu/vướng mắc; PTCM có thể lập khi người cũ vắng mặt.
Thay người trực tiếp tái sử dụng tra cứu/kiểm tra manual, không kế thừa tiền điều kiện PENDING_ASSIGNMENT của UC003/UC003.2. Giao lại bằng gợi ý phải thu hồi trước → PENDING_ASSIGNMENT → run mới → chọn người → confirm → ASSIGNED; không xem gợi ý người thay khi A còn giữ task. Cả hai giữ frozen specification và các validation/audit/notification/handover hiện hành.
Yêu cầu bàn giao cũng áp dụng khi đã có phần việc thực tế dù người thực hiện quên bấm Start; không phụ thuộc một yêu cầu duyệt R01.
PENDING_REVIEW phải trả lại trước khi thay người/thu hồi. Không reopen DONE.
Estimate/requirements/tiêu chí nghiệm thu và scope đóng băng sau lần giao đầu, kể cả đã thu hồi; không sửa mô tả để lách thay đổi scope.
Scope mới tạo task mới và hủy task cũ khi phù hợp. Ghi chú làm rõ không đổi scope được phép, có audit.

## 4. Phân công, quyền và lịch sử

Batch cùng project, không trùng task. Kiểm tra PTCM hiện hành, task, candidate, giá trị đặc tả/requirements/kỹ năng so snapshot và run; khóa/đối chiếu nền phối hợp mọi đường cập nhật liên quan.
Ghi assignments + current assignee/status + status history + notifications trong một transaction. Thiết kế đích không ghi membership riêng.
Lỗi một dòng rollback toàn bộ, trả danh sách lỗi; retry với idempotency key không tạo lại.
MANUAL cho override thiếu/dưới mức mandatory với cảnh báo và lý do; không override role/status hoặc ràng buộc PTCM không nhận task cùng dự án.
RECOMMENDED chọn hạng khác Top-1 bắt buộc lý do; chuyển manual sau run lưu run_id + lý do.

Assignments giữ bất biến nội dung giao; ended_at/ended_by/end_reason ghi một lần khi giao lại/thu hồi/DONE/hủy. Giao lại kết thúc lượt cũ và tạo lượt mới cùng giao dịch; tối đa một lượt mở/task. Không assignment_events. Task terminal giữ người cuối, không tính tải; lịch sử không mất quyền hiệu chỉnh hợp lệ.
Theo [bổ sung 2026-10-13](../../docs/business-baseline-2026-10-13.md), không lập đội trước; PTCM phân công từ nguồn toàn công ty. Người từng được giao suy từ assignments JOIN tasks, giữ cả lượt đã thay/thu hồi và task hủy; không đồng nghĩa đã thực hiện và không cấp quyền toàn dự án. Số dự án đang có việc là COUNT DISTINCT project PLANNING/IN_PROGRESS có task hiện giao cho người đó ở ASSIGNED/IN_PROGRESS/PAUSED/PENDING_REVIEW. Chỉ hiển thị tham khảo, không giới hạn cứng hai dự án hoặc đưa thêm vào công thức. Tải xét xuyên dự án; trách nhiệm PM/PTCM không tự tính là task đang nhận.
Theo lượt phân công đã kết thúc, nhân viên cũ chỉ đọc lịch sử đóng góp của mình. Quyền từ các trách nhiệm dự án khác còn hợp lệ được xét riêng, không bị thu hồi cùng lượt phân công.
Task terminal có thể giữ current_assignee_id là người phụ trách cuối; không tính vào tải và không cho thực thi.

## 5. Effort, kế hoạch và thông báo

PTCM chia sẻ kế hoạch chi tiết ngoài hệ thống (ví dụ Word qua Slack), để PM theo dõi, không phải PM duyệt phân công. Không có UC gửi/duyệt kế hoạch hoặc tích hợp Slack/sprint mới. PM giữ thao tác bắt đầu triển khai và theo dõi xuyên suốt.

Đặc tả chuẩn: [Ghi effort và sửa sai](work_effort_workflow.md).
Effort tự khai gồm việc tại công ty và ở nhà; tổng actual_hours = tổng duration_minutes của log hiệu lực / 60.
Không timer/lịch cắt giờ/chấm công/duyệt OT. 8h quy đổi ngày công và cảnh báo, >24h/người/ngày bị chặn toàn hệ thống.
R02: nghiệm thu dựa trên đầu ra và tiêu chí; thiếu nhật ký không tự động chặn nghiệm thu. Không có chức năng nộp kết quả thay. Dữ liệu thời gian chưa đầy đủ không dùng cho thành phần hiệu suất phụ thuộc thời gian; không ép khai giờ giả.
R01 (chốt tiếp nối 21/10): chỉ tạo/sửa/xóa nhật ký trực tiếp khi task/project IN_PROGRESS và đúng người hiện được giao. ASSIGNED cần Bắt đầu, PAUSED cần Tiếp tục; vẫn đọc nhật ký. Không timer, sinh giờ hoặc sửa ngược trạng thái; nhật ký đã khóa hoặc lượt giao đã kết thúc vẫn theo hiệu chỉnh chung. Xem work_effort_workflow mục 2–3.
Effort task CANCELLED vẫn thuộc tổng effort project; báo riêng phần hủy. Sửa sau DONE không mở lại công việc.
Giữ ngày kế hoạch ban đầu và ngày kế hoạch hiện hành. Gia hạn có lý do/audit, không sửa mốc ban đầu. Quyền/trạng thái và cặp kế hoạch đầu theo [chốt tiếp nối 28/10](../../docs/business-baseline-2026-10-20.md#quyết-định-tiếp-nối-28102026--thay-pm-hạn-task-và-trao-đổi); không đổi estimate/phạm vi.
Ngày kế hoạch không chặn ghi effort thực tế muộn hợp lệ; không buộc gia hạn để hợp thức hóa nhật ký.
Không suy estimate 8h = hoàn thành trong một ngày lịch.

Thông báo nội bộ cho giao/thay/thu hồi, nộp/trả review, kết quả duyệt yêu cầu; cùng transaction nghiệp vụ.
Không bình luận mới khi project COMPLETED/CLOSED/CANCELLED hoặc task terminal.
Đánh dấu notification đã đọc vẫn được; notification không vượt quyền tài nguyên.

## Trách nhiệm nghiệm thu và phạm vi giản lược

Phụ trách chuyên môn (PTCM) của dự án phân công và nghiệm thu công việc, không trực tiếp nhận/thực hiện task trong chính dự án đó. Không giao thêm trách nhiệm nghiệm thu, không có reviewer thay thế do PM chỉ định. PM chỉ có quyền chuyên môn khi kiêm PTCM và cũng phải tuân thủ giới hạn không nhận task. PTCM có thể thực hiện task ở dự án khác nếu đủ điều kiện. Lưu người đánh giá, kết luận và thời điểm; link PR/bình luận/merge không tự thay thế quyết định nghiệm thu trong hệ thống.
Kỹ năng không tự hết hạn; official chỉ chứa năng lực hiện được công nhận, danh mục không có trạng thái INACTIVE.
R01–R05 đã được giản lược theo cơ sở hiện hành; vẫn giữ lịch sử, giới hạn quyền và kiểm tra dữ liệu. Hiệu chỉnh do người thực hiện tự gửi, PTCM hiện hành duyệt, không có quyền PM tự duyệt hoặc ghi thay.
