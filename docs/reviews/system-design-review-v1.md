# Phân tích thiết kế hệ thống — baseline V1 đã xác nhận

> LỊCH SỬ — nội dung bên dưới phản ánh thời điểm V1, đã có các quyết định thay thế. Không dùng để triển khai các quy định cũ về timer, giới hạn một task đang làm, công thức lịch sử hoặc sửa effort. Nguồn hiện hành: [mục lục tài liệu](../README.md), ưu tiên quyết định/Bàn giao và các chốt tiếp nối được dẫn ở đó.

Ngày lập: 2026-09-05. Trạng thái: **CONFIRMED — người dùng đồng ý bộ đề xuất V1**. Rà soát tiếp: 2026-09-06.

Đây là baseline thiết kế đích từ tài liệu gốc, `.claude/docs`, migration và D01–D38. Người dùng đã đồng ý các đề xuất và xác nhận giảng viên không bắt buộc AI/ML, chấp nhận công thức phù hợp. Tài liệu miền đã được cập nhật theo baseline; code/DB và báo cáo gốc chưa triển khai các thay đổi. Từ “đề xuất” trong phần lập luận bên dưới mô tả nguồn gốc quyết định, không còn là yêu cầu xác nhận lại. Các chi tiết phát hiện thêm và phần chưa có bằng chứng được ghi trong [review sau cập nhật](system-design-review-v1-followup.md).

## 1. Kết luận và giới hạn bằng chứng

Lõi bài toán phù hợp: hồ sơ năng lực có kiểm duyệt → yêu cầu kỹ năng của task → đề xuất/tra cứu → PM xác nhận → theo dõi thực hiện. Khoảng trống lớn nằm ở các luồng cập nhật, trả việc, thay người, hủy và dữ liệu phục vụ đánh giá, không nằm ở lựa chọn framework.

Repository chỉ có skeleton M0 và migration 000001; chưa có service nghiệp vụ hay test nghiệp vụ. Lượt kiểm tra trước ghi nhận Neon version 1, dirty=false. Số dòng 20 nhân sự/51 kỹ năng/96 năng lực được lấy từ thống kê `pg_stat_user_tables`, là **ước tính thống kê**, không phải truy vấn COUNT chính xác. Không suy ra toàn bộ cấu trúc DB đang giống migration chỉ từ version. Lượt review này không kết nối lại hoặc thay đổi DB.

Nguồn bằng chứng chính:

- [Quy tắc](../../.claude/docs/domain_rules.md).
- [Phân quyền](../../.claude/docs/roles_permissions.md).
- [Recommendation](../../.claude/docs/ai_recommendation_workflow.md).
- [Luồng CV](../../.claude/docs/cv_extraction_approval_workflow.md).
- [Schema legacy](../../migrations/000001_init.up.sql).
- Use case gốc: `docs/thesis-context/tex/2_Khao_sat.tex` (tham khảo local, không nằm trong bộ nguồn Git).
- Thiết kế gốc: `docs/thesis-context/tex/4_Ket_qua_thuc_nghiem.tex` (tham khảo local, không nằm trong bộ nguồn Git).

## 2. Những kết luận trước cần sửa hoặc giới hạn lại

| Điểm | Kết quả review lại | Hệ quả |
|---|---|---|
| Tách bảng yêu cầu kỹ năng | Cần thiết để giữ đồng thời giá trị chính thức và đề nghị thay đổi; UNIQUE hiện tại không đáp ứng | Giữ đề xuất, bổ sung chống duyệt yêu cầu đã cũ |
| Bỏ cây task nhiều cấp | Là lựa chọn thu hẹp phạm vi, không phải bằng chứng rằng thiết kế gốc sai | Đã đồng ý thu hẹp phạm vi; báo cáo gốc có mô tả `parent_task_id` |
| ADMIN đọc toàn bộ dự án | Chưa có yêu cầu đủ rõ để cấp quyền rộng này | Mặc định giữ phạm vi quản trị tài khoản/nhân sự; không mở toàn bộ nội dung dự án |
| PM trả việc thẳng về IN_PROGRESS | Làm đồng hồ chạy khi nhân viên chưa nhận lại việc | Trả về ASSIGNED; nhân viên chủ động bắt đầu lại |
| Cho nhiều task IN_PROGRESS | Tính giờ theo khoảng trạng thái sẽ cộng trùng giờ của cùng người | Đề xuất một task đang chạy mỗi nhân viên và thêm PAUSED; đây là thay đổi đã đồng ý |
| Phân công lại khi đang chạy | Chỉ thêm assignment mới sẽ thiếu điểm cắt giờ và quyền người cũ | Đóng phiên cũ, giữ log cũ, tạo phân công mới, chuyển ASSIGNED |
| Sửa estimate sau phân công | Làm thay đổi mẫu số/cơ sở so sánh lịch sử | Đóng băng estimate và yêu cầu kỹ năng từ lần phân công đầu tiên trong V1 |
| Workload theo bốn tuần | Chưa có mô hình phân bổ giờ theo ngày; ngày task lại tùy chọn | V1 dùng tổng giờ ước tính của công việc đang mở, gọi đúng là chỉ báo tải tồn đọng |
| `estimate - actual` là remaining effort | Giờ đã trôi qua không chứng minh phần việc còn lại, task quá estimate vẫn có thể còn rất nhiều việc | Không dùng remaining effort nếu nhân viên/PM chưa cung cấp dữ liệu này |
| Params đủ để replay recommendation | Thiếu hồ sơ, yêu cầu và workload tại thời điểm chạy | Lưu snapshot đầu vào đầy đủ của tập ứng viên đã xét |
| Chỉ một run ACTIVE | Không cần thiết và dễ phải cập nhật hàng loạt khi dữ liệu đổi | Lưu kết quả bất biến; kiểm tra phiên bản/điều kiện khi xác nhận |
| Extraction APPROVED khi hết PENDING | Nếu tất cả dòng bị từ chối vẫn có thể thành APPROVED | Tách trạng thái xử lý và tổng hợp kết quả duyệt |
| CV là “học máy thật” | Dictionary/alias/fuzzy matching không tự trở thành machine learning | Mô tả đúng phương pháp; người dùng xác nhận không bắt buộc học máy |
| Transaction sau validate là đủ | Dữ liệu có thể đổi giữa validate và ghi | Kiểm tra cuối trong transaction dưới cơ chế khóa/kiểm soát cạnh tranh |
| Một work log một phiên nhưng vẫn giữ work_date | Phiên có thể qua nhiều ngày, work_date không rõ nghĩa | Phiên lưu start/end và lịch áp dụng; bỏ work_date hoặc thay bằng bảng phân bổ ngày nếu có nhu cầu |
| UC004 minh họa 15:00 → hôm sau 10:30 bằng 3,5h | Theo lịch đã nêu là 2h + 2,5h = **4,5h** | Sửa ví dụ và các tổng/chênh lệch phụ thuộc |

## 3. Phạm vi và phân quyền thống nhất — D01, D02, D03, D09, D27

Đề xuất V1 phục vụ một tổ chức, UI tiếng Việt; quản lý năng lực, dự án, task, phân công, review, thời gian và CV. Bỏ quản lý chứng chỉ riêng, nhiều tổ chức, lịch nguồn lực, dependency task, nghỉ phép và tối ưu phân công đồng thời. Đây là giới hạn sản phẩm, phải viết đồng nhất trong báo cáo.

| Actor | Được thực hiện | Phạm vi |
|---|---|---|
| ADMIN | Tài khoản, danh mục, hồ sơ nhân sự, duyệt năng lực/CV, thay PM phụ trách khi cần | Dữ liệu nhân sự toàn tổ chức; không mặc định đọc bình luận/nội dung mọi dự án |
| PM | Tạo và quản lý dự án, tra cứu năng lực đã duyệt, đề xuất, phân công, review | Dự án có `pm_employee_id` là mình; tra cứu nhân sự toàn tổ chức với trường dữ liệu được phép |
| EMPLOYEE | Gửi yêu cầu hồ sơ/kỹ năng/CV, xem và thực hiện task, bình luận | Hồ sơ của mình và task hiện đang phụ trách |

V1 mỗi tài khoản có một vai trò; chỉ EMPLOYEE active được nhận task, cả ở nhánh AI và thủ công. PM có hồ sơ năng lực để tra cứu nhưng không nhận task trong V1. Nếu cho PM kiêm thực thi, phải sửa đồng thời quyền `/my/tasks`, cập nhật trạng thái, candidate pool và test; không dùng một cờ chỉ nằm trong scorer.

Tạo user và employee trong một transaction qua một luồng tạo nhân sự duy nhất; không có API tạo user rời rồi quên tạo employee. `employees.user_id UNIQUE NOT NULL` chỉ bảo đảm chiều employee → user, không tự bảo đảm mọi user đều có employee; service phải bảo đảm chiều còn lại.

Tài khoản được thao tác khi cả user và employee active. Khóa/cho nghỉ việc thu hồi phiên. Khi còn task đang mở hoặc dự án đang quản lý, màn hình phải liệt kê ảnh hưởng; khóa khẩn cấp vẫn được thực hiện. PM xử lý task của người bị khóa bằng chuyển người/hủy. ADMIN có hành động chuyển PM của dự án sang một PM active, bắt buộc lý do và audit; hành động này không cho ADMIN quyền tự phân công task.

Chuyển role EMPLOYEE → vai trò khác bị chặn khi còn task chưa kết thúc; chuyển PM → vai trò khác bị chặn khi còn dự án đang mở do họ phụ trách. Không cho vô hiệu hóa/hạ quyền ADMIN active cuối cùng.

`project_members` trong V1 biểu diễn **đã tham gia**, được thêm từ phân công và giữ lại khi thay người. Nó không cấp quyền xem mọi task. PM hiện tại được thể hiện qua `pm_employee_id`; giao diện đội dự án hiển thị PM riêng. Không có thao tác thêm/rời member độc lập trong V1; `MANUAL` có trong enum cũ không bắt buộc phải có chức năng sử dụng. Lịch sử tham gia khi tra cứu chỉ trả mã/tên dự án và thông tin tham gia phù hợp, không trả nội dung task của dự án khác.

## 4. Hồ sơ và duyệt năng lực — D04, D22, D25

`employee_skills` chỉ chứa kỹ năng chính thức. Cặp `(employee_id, skill_id)` vẫn unique; approval actor/time đều bắt buộc với dữ liệu chính thức mới. Kỹ năng có thể được thu hồi hiệu lực có lý do, giữ vết cũ; không xóa lịch sử.

Thêm `employee_skill_change_requests`: nhân sự, kỹ năng, thao tác ADD/UPDATE/REMOVE, nguồn MANUAL/CV, giá trị đề nghị, giá trị trước duyệt, phiên bản hồ sơ cơ sở, người gửi, thời điểm, quyết định, người duyệt, lý do. Trạng thái PENDING → APPROVED/REJECTED hoặc người gửi WITHDRAWN. Tối đa một yêu cầu PENDING trên mỗi người/kỹ năng đã được map.

Quy trình: gửi yêu cầu → ADMIN đối chiếu minh chứng và giá trị chính thức hiện tại → duyệt/từ chối. Trong lúc chờ, engine tiếp tục dùng giá trị chính thức cũ. Khi duyệt, khóa/kiểm tra phiên bản hiện hành và cập nhật hồ sơ cùng quyết định trong một transaction. Nếu hồ sơ đã đổi, trả conflict để ADMIN xem lại; không âm thầm ghi đè. Từ chối và thu hồi kỹ năng bắt buộc lý do.

ADMIN nhập trực tiếp cũng dùng cùng service phê duyệt: tạo yêu cầu và duyệt ngay với dấu vết nguồn/người thực hiện. Dữ liệu kỹ năng của ADMIN không tham gia candidate pool V1 nên không đặt thêm quy trình hai ADMIN duyệt chéo trong phạm vi này.

Yêu cầu cập nhật thông tin cá nhân là chức năng đã có ở Chương 2 nhưng danh sách D01–D38 trước bỏ sót thiết kế lưu trữ. Đề xuất thêm `employee_profile_change_requests` với các trường được phép đề nghị như tên, email, điện thoại; ADMIN duyệt. Role, mã nhân viên, phòng ban, chức danh và trạng thái làm việc chỉ ADMIN quản lý, không nhận qua body cập nhật hồ sơ tự do. Email trùng được kiểm tra lại khi duyệt.

Danh mục proficiency cần rubric: BASIC làm việc đơn giản có hướng dẫn; INTERMEDIATE độc lập với việc thường gặp; ADVANCED xử lý tình huống phức tạp; EXPERT thiết kế/định hướng và hỗ trợ người khác. ADMIN chịu trách nhiệm xác nhận dựa trên minh chứng chuyên môn/đánh giá PM, không coi việc duyệt là chứng nhận khách quan tuyệt đối. Số năm kinh nghiệm là dữ liệu bổ trợ, không quy đổi tự động thành proficiency.

Skill INACTIVE vẫn hiển thị trong lịch sử, không được thêm vào hồ sơ/yêu cầu task mới. Task chưa từng phân công chứa skill vừa inactive phải chỉnh trước khi phân công lần đầu. Task từng phân công giữ requirement đóng băng, kể cả sau thu hồi; việc ngừng dùng skill trong danh mục không tự xóa hiệu lực kỹ năng nhân sự đã duyệt. Khi giao lại task này vẫn đối chiếu requirement cũ với năng lực chính thức, đồng thời hiển thị cảnh báo danh mục đã ngừng sử dụng. Thu hồi năng lực của một nhân viên là quyết định riêng, không đồng nhất với ngừng dùng skill trong danh mục.

## 5. Cấu trúc và vòng đời dự án/công việc — D05–D08

Đề xuất cấu trúc `Project → TaskGroup → Task`; mỗi task bắt buộc thuộc một group cùng project. Bỏ `parent_task_id` và `task_type` trong thiết kế đích đã đồng ý. Đây là thay đổi so với cây phân cấp trong báo cáo gốc. Group được đổi tên/mô tả; task chưa kết thúc có thể chuyển group trong cùng project, ghi audit. Không chuyển task sang project khác.

Task được tạo cùng ít nhất một requirement trong một transaction; giao diện có thể giữ nháp chưa hợp lệ nhưng không lưu thành task hợp lệ trong DB. `estimate_hours > 0`; mã task unique trong project. Ngày task có thể bỏ trống, nhưng nếu nhập thì nhập cả hai, đúng thứ tự và nằm trong khoảng dự kiến của project. Thu hẹp ngày project không được làm ngày task nằm ngoài khoảng.

| Project: từ → sang | Actor và điều kiện |
|---|---|
| Tạo → PLANNING | PM tạo, tự gán PM phụ trách |
| PLANNING → IN_PROGRESS | PM bắt đầu, có ít nhất một task ASSIGNED |
| IN_PROGRESS → COMPLETED | PM xác nhận; có ít nhất một task DONE và tất cả task còn lại DONE hoặc CANCELLED |
| COMPLETED → CLOSED | PM khóa sổ; chuyển mọi task DONE thành CLOSED trong cùng transaction |
| PLANNING/IN_PROGRESS → CANCELLED | PM nhập lý do; kết thúc phiên đang chạy, hủy task chưa DONE; giữ nguyên task DONE và lịch sử |

PLANNING cho tạo/sửa/phân công, nhưng chưa cho nhân viên bắt đầu. IN_PROGRESS cho tạo thêm task và phân công. COMPLETED chỉ cho xem và đóng dự án; CLOSED/CANCELLED chỉ đọc. Không mở lại dự án trong V1. Dự án toàn task hủy thì dùng CANCELLED, không đánh dấu COMPLETED. Đổi trạng thái project luôn ghi lịch sử.

| Task: từ → sang | Actor | Tác động thời gian |
|---|---|---|
| Tạo → PENDING_ASSIGNMENT | PM | Không tính giờ |
| PENDING_ASSIGNMENT → ASSIGNED | PM xác nhận | Không tính giờ |
| ASSIGNED → IN_PROGRESS | Người được giao; project IN_PROGRESS | Mở phiên |
| IN_PROGRESS → PAUSED | Người được giao, có lý do | Đóng phiên, lưu giờ |
| PAUSED → IN_PROGRESS | Người được giao | Mở phiên mới |
| IN_PROGRESS → PENDING_REVIEW | Người được giao nộp | Đóng phiên, lưu giờ |
| PENDING_REVIEW → ASSIGNED | PM trả lại, có lý do | Không tự mở phiên |
| PENDING_REVIEW → DONE | PM nghiệm thu | Không cộng thời gian chờ review |
| DONE → CLOSED | Khi PM đóng project | Không tính giờ |
| PENDING_ASSIGNMENT/ASSIGNED/IN_PROGRESS/PAUSED/PENDING_REVIEW → CANCELLED | PM hủy có lý do, project còn mở | Đóng phiên nếu đang chạy; không mở phiên mới |

PAUSED và CANCELLED là bổ sung đã đồng ý. Một nhân viên nhận nhiều task nhưng chỉ có một task IN_PROGRESS trên toàn hệ thống tại một thời điểm. Muốn chuyển việc, nhân viên pause việc hiện tại rồi bắt đầu việc mới. Khi người đang chạy bị khóa tài khoản, PM có thể chuyển người hoặc hủy để giải phóng phiên; không tự xóa thời gian đã ghi.

Chỉ sửa estimate/requirements trước lần phân công đầu tiên. Sau đó khóa hai nhóm dữ liệu này để giữ mốc so sánh; scope thay đổi đáng kể được tạo task mới và hủy task cũ có lý do. Ghi chú làm rõ không thay scope vẫn được phép, có audit. Không reopen DONE trong V1; phát sinh sau nghiệm thu là task mới.

Tiến độ dự án: `số task DONE hoặc CLOSED / số task không CANCELLED`, thể hiện theo số công việc, không gọi là phần trăm effort. Mẫu số bằng 0 hiển thị chưa có công việc hiệu lực. COMPLETED là xác nhận của PM, không tự chuyển chỉ vì dashboard hiện 100%.

## 6. Phân công, thu hồi và thay người — D09–D12, D24, D26

Một batch chỉ chứa task cùng project, không trùng task ID. Draft nằm trên giao diện, không đặt chỗ nhân sự và không làm tăng workload chính thức. PM có thể chọn nhiều task cho một người; trước xác nhận phải thấy tổng tải sau các lựa chọn dự kiến.

Nhánh thủ công cho phép thiếu/không đủ kỹ năng bắt buộc khi PM xác nhận cảnh báo và nhập lý do. Cả hai nhánh luôn chặn người inactive, tài khoản bị khóa, người không thuộc role EMPLOYEE. Workload cao là cảnh báo; không suy ra người đó trống lịch hay có thể đáp ứng deadline.

All-or-nothing: validation sơ bộ để hiển thị lỗi; khi xác nhận phải kiểm tra lại trong transaction với khóa project/task và các bản ghi nhân sự liên quan theo thứ tự nhất quán. Mọi đường đổi quyền/trạng thái phải phối hợp cùng quy ước khóa/version. Chỉ bọc các câu INSERT bằng transaction không loại trừ race condition. Lỗi bất kỳ dòng nào → không ghi assignment, task, membership hoặc notification của cả batch.

Thay người cho ASSIGNED/IN_PROGRESS/PAUSED: bắt buộc lý do, đóng phiên đang chạy của người cũ, thêm assignment mới, chuyển ASSIGNED với người mới. Thời gian đã ghi luôn thuộc người cũ. Không thay người khi PENDING_REVIEW; PM phải trả lại trước. Task đã có người thực hiện khác không dùng làm mẫu PerfScore cá nhân trong V1.

Thu hồi để chưa chọn người thay: cùng các trạng thái trên, kết thúc phiên nếu cần, chuyển PENDING_ASSIGNMENT, xóa current assignee; estimate/requirements vẫn đóng băng nếu task từng được giao. Thêm `assignment_events` cho ASSIGN/REASSIGN/UNASSIGN và lý do; assignments tiếp tục append-only. Event chỉ diễn tả thay đổi người phụ trách, status history diễn tả thay đổi trạng thái.

Không cung cấp API xóa vật lý tài khoản, nhân sự, dự án, task hoặc dữ liệu audit. Không kết luận “đã bảo vệ lịch sử” chỉ từ append-only service khi FK vẫn cascade; migration đích phải kiểm tra đường xóa và dùng RESTRICT cho quan hệ chứa lịch sử cần giữ.

Notifications trong DB là thông báo nội bộ: phân công mới/thay người/thu hồi, nộp review/trả lại, kết quả duyệt yêu cầu. Chúng được ghi cùng transaction; không cam kết email hay push. Project đóng/cancel thì không cho tạo bình luận mới, nhưng chủ notification vẫn được đánh dấu đã đọc. Sau chuyển người, người cũ xem được lịch sử đóng góp/thông báo của mình; không được mở nội dung mới của task thông qua ID hoặc notification.

## 7. Recommendation và dữ liệu nghiên cứu — D13–D21

Candidate pool: EMPLOYEE active của toàn tổ chức; chỉ đọc kỹ năng chính thức hiệu lực. Không giới hạn thành viên project. V1 bỏ bonus thành viên để giảm một tham số chưa có cơ sở; vẫn hiển thị trạng thái tham gia để PM tham khảo.

Đề xuất làm rõ `is_mandatory` thành **mức tối thiểu bắt buộc của nhánh AI**: thiếu kỹ năng hoặc dưới required_level đều không vào Top-N. Nhánh thủ công cho override có lý do. Điều này thay đề xuất cũ “thiếu hoàn toàn thì loại, dưới mức thì nhân 0,7”, giúp từ “bắt buộc” thống nhất với hành vi. Đây là lựa chọn nghiệp vụ đã đồng ý.

SkillScore giữ công thức trung bình có trọng số `Σ(w_i × min(level_employee/level_required, 1)) / Σw_i`, missing optional skill có giá trị 0. Mức 1–4 là phép mã hóa heuristic của thang thứ bậc, không khẳng định BASIC bằng một phần tư EXPERT. Weight nguyên 1–5 với nhãn từ thấp đến cao; không phạt thêm 0,7 khi đã hard filter kỹ năng bắt buộc.

WorkloadScore V1: `clamp(1 - open_estimate_hours/80, 0, 1)`, trong đó open tasks là ASSIGNED/IN_PROGRESS/PAUSED/PENDING_REVIEW trên tất cả project và theo current assignee. Task chưa phân công không tính vào tải cá nhân. Task chờ review vẫn có thể phải làm lại, nên giữ trong tổng và hiển thị riêng số lượng chờ review. **80h là ngưỡng chuẩn hóa thử nghiệm cho tải tồn đọng, không phải lịch còn trống trong hai/bốn tuần**. Deadline/priority phục vụ hiển thị và thứ tự PM xem task; không tuyên bố scorer đã tối ưu lịch.

Giữ phiên bản thực nghiệm `0,7 × SkillScore + 0,2 × WorkloadScore + 0,1 × PerfScore` để so với thiết kế trước. PerfScore dùng tối đa 10 task DONE/CLOSED gần nhất theo thời điểm nghiệm thu, có actual_hours > 0, chỉ một người thực hiện và estimate không thay đổi. Dưới 3 mẫu dùng 0,5. Đủ mẫu dùng công thức cũ `clamp(Σestimate/Σactual, 0, 1.2)/1.2`, lưu sample size và chỉ rõ đây là chỉ báo thời gian theo workflow. Các số 10, 3, 80 và bộ trọng số là cấu hình thử nghiệm, phải khảo sát độ nhạy, không gọi đã tối ưu.

Engine trả tối đa Top-3, có thể rỗng. Matching score lưu/hiển thị thang 0–100, làm tròn hai chữ số; các điểm thành phần cũng cùng thang khi persist/API. Phá hòa theo matching score đã làm tròn giảm dần, tải tăng dần, employee ID tăng dần. Bộ tham số và quy tắc làm tròn/phá hòa nằm trong phiên bản engine.

Mỗi task mỗi lần chạy có một run riêng; nhiều task được xếp hạng độc lập trên cùng snapshot, không phải tối ưu phân công theo lô. Do đó một người có thể Top-1 cho nhiều task. Bảng dự kiến cộng tải sau lựa chọn để PM quyết định, không ngầm thay đổi ranking trong batch.

Run lưu engine version, params, thời điểm snapshot, task/requirements version, tập ứng viên đầu vào gồm cả người bị loại, kỹ năng chính thức, trạng thái, workload và số liệu hiệu suất đủ để chạy lại. Output snapshot lưu Top-N, điểm, giải thích và lý do loại. Hạn chế snapshot ở dữ liệu phục vụ scorer, không sao chép CV thô/thông tin liên hệ. `params` riêng lẻ không đủ cho replay.

Kết quả run bất biến; không thêm ba cờ ACTIVE/STALE/CONSUMED làm nguồn sự thật thứ hai. Task specification thay đổi hoặc task không còn chờ phân công thì run không dùng để tạo assignment mới. Thay đổi hồ sơ của người được chọn phải được kiểm tra khi xác nhận: nếu phiên bản năng lực đổi, yêu cầu chạy lại hoặc chuyển manual có lý do. Tải các ứng viên thay đổi theo thời gian không vô hiệu mọi run; giao diện ghi rõ ranking ở thời điểm nào và hiển thị tải mới nhất khi xác nhận. Kết quả không được quảng bá là thứ hạng thời gian thực.

| Quyết định PM | assignment_type | Dữ liệu truy vết |
|---|---|---|
| Chọn Top-1 của run | AI | run_id + item_id + score snapshot |
| Chọn hạng 2/3 | AI | như trên + adjust_reason bắt buộc |
| Chuyển manual sau xem run | MANUAL | run_id, item_id NULL, adjust_reason; không gán điểm Top-1 cho người ngoài danh sách |
| Manual ngay từ đầu | MANUAL | run_id/item_id/score NULL; lý do nếu override cảnh báo |

Một run chỉ gắn tối đa một quyết định phân công đã commit; request thất bại không đánh dấu đã chọn. Bỏ cột `is_selected` trong thiết kế đích hoặc chỉ coi nó là dữ liệu suy ra từ assignment; không có hai nguồn sự thật độc lập. Nếu PM bỏ màn hình hoặc chạy lại mà không quyết định, không tính là từ chối Top-1.

Đánh giá: tỷ lệ chọn Top-1/Top-3 lấy mẫu số là số quyết định đã commit có run không rỗng, bao gồm chuyển manual; báo riêng số run rỗng, hủy, chạy lại. Tách lần phân công đầu và phân công lại. Bổ sung bộ task–candidate do người đánh giá độc lập gán mức phù hợp, so sánh baseline skill-only, skill+workload và ba thành phần; đo chất lượng Top-K, tính ổn định, thời gian xử lý và độ nhạy trọng số. Bộ kiểm thử sinh từ chính công thức chỉ kiểm tra implementation, không chứng minh chất lượng recommendation.

## 8. CV và yêu cầu duyệt — D22, D23

Đề xuất V1 nhận PDF có text và DOCX, tối đa 10 MiB; chưa hỗ trợ OCR/PDF mã hóa. Tiếng Việt và tiếng Anh trong CV được chấp nhận ở mức thử nghiệm, cần đo riêng. Chọn local disk qua storage interface cho triển khai một máy; tên tệp do server sinh, file private, download qua endpoint phân quyền. Đây là phương án đã đồng ý; chưa chọn thư viện hoặc cài dependency.

Tách `processing_status = QUEUED/PROCESSING/SUCCEEDED/FAILED` khỏi `review_status = PENDING_REVIEW/APPROVED/PARTIALLY_APPROVED/REJECTED`. Chỉ xử lý thành công mới được review. Worker trong cùng ứng dụng lấy job lưu ở DB, có giới hạn thời gian, số lần retry và nhận lại job bị gián đoạn; không cần thêm hàng đợi ngoài. File storage và DB không chung transaction: upload thất bại phải dọn file tạm; chỉ enqueue sau khi file bền vững, thao tác retry phải idempotent.

Nhân viên gửi CV tạo bản PENDING; ADMIN upload thay mặt cũng giữ audit. **Upload mới không lập tức thay CV chính thức**. Sau xử lý thành công ADMIN duyệt việc xuất bản CV; lúc đó CV mới ACTIVE, CV cũ REPLACED trong một transaction, tối đa một ACTIVE mỗi người. Từ chối publication chuyển REJECTED; parse lỗi giữ PENDING với processing FAILED để retry hoặc rút/ARCHIVED, giữ CV cũ. Duyệt tệp CV và duyệt năng lực là hai quyết định: CV có thể được chấp nhận nhưng từng kỹ năng vẫn chờ duyệt. Khi đóng review extraction rỗng, đặt REJECTED với mã lý do NO_SKILLS_FOUND; nhãn UI là “Không có kỹ năng để cập nhật”, không hàm ý từ chối toàn bộ tệp CV.

Nhận diện kỹ năng: parse → normalize → alias/exact match theo danh mục → fuzzy candidate → ADMIN map. Không suy proficiency/năm kinh nghiệm chỉ vì tên kỹ năng xuất hiện. Khi không đủ minh chứng, giữ suggested_level NULL, ADMIN cung cấp mức và căn cứ trước khi duyệt. `confidence_score` nếu chưa được hiệu chỉnh chỉ gọi là điểm tương đồng; có thể NULL, không hiển thị như xác suất đúng.

Sau khi map một extracted skill sang skill_id, đưa vào cùng luồng `employee_skill_change_requests` với tham chiếu `extracted_skill_id` unique. Nhiều dòng CV map cùng một skill phải hợp nhất minh chứng/chọn một đề nghị, không last-write-wins. Nếu đã có yêu cầu pending nhập tay, ADMIN/người gửi giải quyết hoặc rút yêu cầu trước; không âm thầm ghi đè. Request quyết định là nguồn chuẩn; trạng thái extracted skill suy ra từ request hoặc quyết định từ chối dòng không map được.

Khi còn bất kỳ dòng chưa xử lý thì review_status là PENDING_REVIEW. Khi hết pending: tất cả approved → APPROVED; có cả approved và rejected → PARTIALLY_APPROVED; tất cả rejected → REJECTED. Không tìm thấy kỹ năng là kết quả xử lý thành công rỗng, cần ADMIN đóng review với lý do, không coi là parser failure. Mỗi quyết định dòng là một transaction; không giữ transaction mở trong suốt buổi duyệt của ADMIN.

Thuật toán từ điển/fuzzy là xử lý ngôn ngữ dựa trên luật. Người dùng xác nhận giảng viên không bắt buộc mô hình học máy; V1 không cần bổ sung ML chỉ để có nhãn AI. Đánh giá trích xuất bằng precision/recall/F1 trên CV đã gán nhãn; không suy proficiency thiếu căn cứ. Chấp nhận phương pháp không đồng nghĩa đã kiểm chứng chất lượng trên dữ liệu thực.

## 9. Thời gian và ý nghĩa của actual_hours — D28–D30

Theo UC004, dữ liệu đo là **thời gian trong trạng thái làm việc, cắt theo lịch chuẩn**. Nó không chứng minh số giờ lao động thực tế: người dùng có thể quên pause hoặc rời máy. Báo cáo/UI phải dùng cách gọi tương ứng; không dùng làm bảng chấm công/lương hoặc khẳng định hiệu suất lao động khách quan.

Phương án đề xuất chống cộng trùng: một phiên mở/nhân viên, task PAUSED khi chuyển việc. `work_logs` lưu một phiên gồm task, employee, assignment, started_at, ended_at, valid_work_seconds, hours_spent, calendar_version, close_reason và mốc history liên quan. Phiên đang mở có ended_at NULL và hours_spent NULL; phiên đóng mới được cộng vào actual_hours. work_date cũ bỏ; báo cáo ngày suy từ khoảng giao lịch nếu cần, không nhét ngày đầu tiên làm ngày đại diện cho toàn phiên.

Lịch chuẩn: Asia/Ho_Chi_Minh, thứ Hai–Sáu, 08:00–12:00 và 13:00–17:00, danh sách ngày nghỉ, không overtime. Lưu snapshot/version cấu hình có thể truy xuất; pin lịch khi bắt đầu phiên. Đổi cấu hình chỉ áp dụng phiên mới; phiên cũ không tính lại ngầm. Thiếu lịch khi bắt đầu thì từ chối mở phiên, tránh để người dùng làm xong mới phát hiện không tính được giờ.

Tính bằng tổng số giây giao với lịch chuẩn, làm tròn giờ để hiển thị. Tổng task tính từ tổng giây của mọi phiên đã đóng, không cộng những giờ đã làm tròn từng đoạn gây sai số. Snapshot lịch phải còn tồn tại để pause/nộp/hủy được dù cấu hình hiện tại thay đổi. Task actual_hours giữ toàn bộ đóng góp của các nhân viên từng làm task; báo cáo cá nhân dựa trên log của từng người.

Ví dụ gốc 20/05/2026 15:00 → 21/05/2026 10:30 = 4,5h nếu hai ngày đều làm việc bình thường. Phiên chỉ qua cuối tuần có thể bằng 0h; vẫn lưu mốc để audit, không đưa task actual=0 vào PerfScore. Thời gian chờ review không tính; PM trả lại ASSIGNED cũng không tính đến khi nhân viên bấm bắt đầu.

Đổi trạng thái, mở/đóng phiên, ghi history và cập nhật tổng giờ cùng một transaction. Hai request bắt đầu đồng thời cho cùng người chỉ một request thành công, dùng partial unique index phiên mở theo employee và task cùng khóa thích hợp. Retry nộp/pause không tạo log trùng.

## 10. Kiến trúc, API và xác thực — D31–D35

Giữ Go + chi + pgx + PostgreSQL; frontend cùng origin, build static nhúng Go. Domain và matching thuần dữ liệu; service điều phối use case và transaction runner, khai báo interface repository cần dùng; postgres adapter thực thi interface, composition root nối dependency. Biểu đồ dependency phải phân biệt luồng gọi runtime với chiều import; service không import adapter PostgreSQL.

API thay đổi một phần dùng PATCH; PUT dùng khi thay toàn bộ tập requirements. Không trả thẳng entity DB vì tên snake_case giống nhau: DTO theo vai trò phải loại password hash, token hash, nội dung CV pending và trường nhân sự ngoài phạm vi. Actor ID từ phiên xác thực; employee_id trong body phân công là **đối tượng được chọn**, vẫn cần nhận và kiểm tra, không thay nó bằng employee_id của PM.

API thao tác nhạy cảm dùng expected_version và idempotency key. Khóa idempotency gắn actor + operation + key; lưu request hash và kết quả trong cùng transaction ghi nghiệp vụ. Cùng key/cùng payload trả kết quả cũ; cùng key/khác payload trả 409. Không cần áp dụng hạ tầng này cho mọi GET hoặc thay đổi ít rủi ro.

200 cho batch thành công; 400 lỗi định dạng; 401 phiên không hợp lệ; 403 sai quyền; 404 không tồn tại; 409 trạng thái/version không hợp lệ; lỗi batch trả chỉ số dòng/task ID và mã lỗi. Chính sách 403/404 hiện tại được giữ nhưng phải bỏ lời hứa không tiết lộ sự tồn tại tài nguyên, vì hai mã khác nhau có thể tiết lộ. Không trả nội dung tài nguyên khi thiếu quyền. Danh sách page/limit/total, mặc định limit 20, tối đa 100.

Giữ JWT access 15 phút + refresh rotation 7 ngày như quyết định trước. Đề xuất kiểm tra user/employee active, role hiện hành và session chưa revoked ở mọi API nghiệp vụ, có thể gộp vào truy vấn quyền. Điều này chấp nhận roundtrip DB để khóa/đổi quyền/logout có hiệu lực ngay; phải bỏ lập luận “JWT loại bỏ mọi truy vấn phiên”. Không âm thầm đổi sang session opaque; đó là phương án thay thế có thể bàn sau nếu muốn giảm phức tạp.

Refresh cookie Path=/api/auth; access Path=/api; host-only, HttpOnly, Secure, SameSite=Lax. JWT có session ID; bảng auth_sessions lưu user, trạng thái thu hồi và expiry tuyệt đối của phiên; refresh_tokens lưu hash, session/family ID, expires_at, revoked_at, replacement. Rotation atomically thu hồi token cũ và phát token mới; reuse thu hồi phiên. Frontend serialize refresh để tránh nhiều tab tự gây reuse; backend vẫn an toàn khi có cạnh tranh. Logout thu hồi phiên hiện tại, khóa user/đổi role thu hồi tất cả phiên. Giới hạn 7 ngày tính từ login, refresh không kéo dài vô hạn.

Với SPA cùng origin: request ghi cần header chống CSRF theo chiến lược thống nhất (signed double-submit gắn phiên), kiểm tra Origin chính xác; login cũng kiểm tra nguồn, không chỉ dựa vào SameSite. Cookie path quyết định request nào mang cookie; SameSite không thay thế toàn bộ bảo vệ CSRF. Cơ sở: [MDN Set-Cookie](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Set-Cookie), [OWASP CSRF](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html).

Rate limit in-memory cho triển khai một process, theo IP thật và username chuẩn hóa; thông số ban đầu phục vụ test: 30 lần/5 phút/IP, 5 lần thất bại/5 phút/username, trả 429 có Retry-After, không tự khóa vĩnh viễn tài khoản. Chỉ tin forwarded IP từ proxy được cấu hình. Chính sách mật khẩu và cơ chế đổi/reset mật khẩu cần đặc tả ở M1; đồ án cần có đường đổi mật khẩu và ADMIN reset buộc đổi lần đầu, không chỉ login bằng seed password.

OpenAPI viết tay + UI nhúng giữ nguyên. Test chi.Walk chỉ kiểm tra method/path, không chứng minh request/response/auth đúng schema. Bổ sung contract test cho schema response, lỗi và quyền của các endpoint trọng tâm. Spec/UI chỉ bật dev; cần route lấy CSRF token và hướng dẫn thử auth trong Swagger. Health trả trạng thái tổng quát, chi tiết lỗi DB chỉ log nội bộ đã che thông tin nhạy cảm.

Cơ sở concurrency: [PostgreSQL Transaction Isolation](https://www.postgresql.org/docs/current/transaction-iso.html) nêu các SELECT liên tiếp ở Read Committed có thể thấy dữ liệu khác nhau. Snapshot đầu vào recommendation lấy nhất quán; xác nhận dùng kiểm tra/khóa phù hợp, không giả định cùng transaction đồng nghĩa dữ liệu không đổi.

## 11. Thay đổi dữ liệu và hợp đồng cần đồng bộ

| Miền | Thay đổi thiết kế đích | Invariant/test bắt buộc |
|---|---|---|
| Users/employees | Tạo atomically, kiểm tra role/status hiện hành | Không có account mồ côi qua luồng ứng dụng |
| Skills | Bảng change_requests; hồ sơ chính thức có version/hiệu lực và approval đầy đủ | Request pending không thay hồ sơ; stale approval bị chặn |
| Profile | Bảng profile_change_requests | Không lẫn quyền sửa role/status vào sửa hồ sơ cá nhân |
| Projects | State history, hành động chuyển PM, version | Không có task bắt đầu trước project; hủy đóng phiên atomically |
| Tasks | group bắt buộc cùng project, bỏ parent/type, thêm PAUSED/CANCELLED, khóa spec sau giao | Composite FK hoặc constraint tương đương; không nhóm chéo project |
| Assignment | run_id, item_id, adjust_reason, event thu hồi/thay người | Task/employee/run khớp; không ghi trùng; score không nhận từ client |
| Recommendation | Snapshot input/output, task version, rank unique, score range | Unique(run,employee), unique(run,rank), tối đa một quyết định commit/run |
| Work logs | Phiên start/end, assignment, calendar snapshot/version, giây hợp lệ; bỏ work_date | Một phiên mở/người và task; không cộng trùng |
| CV | Tách file publication/processing/review; map bằng extracted_skill_id | Một CV ACTIVE/người; không approve skill chưa map/có level NULL |
| Auth | auth_sessions + refresh_tokens | Rotation/reuse/logout/lock nhất quán |
| Idempotency/audit | request hash/result; audit thay cấu hình/phân quyền/spec | Retry không ghi lại; không log bí mật |

Các quan hệ nhiều bảng như item phải thuộc đúng task và employee cần composite FK khi phù hợp hoặc kiểm tra transaction có khóa; không giả định CHECK PostgreSQL tự tra bảng khác. Source-of-truth: current assignee từ task; lịch sử từ assignment_events/assignments; selected recommendation suy từ assignment; actual_hours là cache được đối soát với log. Mọi transaction ghi phải giữ các bản sao này khớp.

API thiếu cần bổ sung: tạo nhân sự kèm user; gửi/duyệt profile request; gửi/duyệt/thu hồi skill request; project start/complete/cancel/transfer-PM; task start/pause/resume/submit-review/approve/return/cancel; unassign/reassign; đọc run đã lưu; upload/process status/publish CV; đổi/reset mật khẩu. Các action route phải gọi cùng các service invariants, không làm hai implementation riêng cho từng màn hình. Endpoint duyệt CV dùng extracted_skill_id, không dùng skill_id vì skill_id có thể NULL hoặc trùng.

Không chốt mọi thay đổi vào migration 000002 chỉ vì đây là số kế tiếp. Sau xác nhận chia migrations theo miền và dependency, mỗi migration có kế hoạch backfill/rollback. Không sửa 000001 đã áp dụng. Trước thay cột/constraint phải COUNT và kiểm tra dữ liệu thật; version và thống kê không thay thế bước này.

## 12. Kiểm tra các kịch bản liên miền bằng lập luận

Đây là walkthrough thiết kế, **chưa phải kết quả chạy test**.

| Ca | Kết quả mong đợi trong V1 |
|---|---|
| Java chính thức mức 2, đề nghị mức 3 | Tra cứu/AI vẫn mức 2; duyệt mới đổi mức 3 |
| CV và nhập tay cùng đề nghị Java | Yêu cầu pending trùng được báo rõ, không ghi đè |
| ADMIN A/B duyệt cùng yêu cầu | Một quyết định commit; request sau conflict/idempotent |
| Upload CV mới nhưng parse lỗi | CV cũ vẫn ACTIVE; hồ sơ kỹ năng không đổi |
| Một extracted skill chưa map | Không approve; chỉ map hoặc reject |
| Tất cả skill CV bị reject | Review REJECTED, không APPROVED |
| Task ở project A dùng group B | Bị chặn trước lưu và bởi ràng buộc DB |
| Task ASSIGNED trong project PLANNING | Xem được, chưa start được |
| PM trả review lúc 08:00, nhân viên bắt đầu 10:00 | 08:00–10:00 không tính |
| Nhân viên bấm start hai task cùng lúc | Chỉ một phiên được mở |
| Pause thứ Sáu, resume thứ Hai | Khoảng pause không tính; lịch cũng loại cuối tuần |
| Thay người trong phiên đang chạy | Giờ trước mốc thuộc người cũ; người mới ASSIGNED chưa chạy |
| Người đang làm bị khóa khẩn cấp | Không gọi API nữa; PM vẫn được chuyển người/hủy |
| PM nghỉ việc | ADMIN chuyển ownership có audit; PM mới thao tác được |
| Batch có một nhân viên bị khóa | Toàn bộ batch không ghi dữ liệu |
| Retry confirm sau timeout | Cùng idempotency key nhận lại kết quả, không assignment trùng |
| Task spec đổi sau khi chạy AI | Không xác nhận AI từ run cũ |
| Tải đổi nhưng skill/spec không đổi | Hiển thị tải mới, giữ ranking có timestamp; PM xác nhận cảnh báo nếu có |
| Một người Top-1 của 5 task | Hiển thị tổng tải dự kiến, không tuyên bố tối ưu cả lô |
| AI không có ứng viên đủ mandatory | Trả Top-N rỗng; PM có thể manual override có lý do |
| Project bị cancel khi có người đang chạy | Đóng phiên và hủy task mở cùng transaction; lịch sử được giữ |
| Task DONE rồi project CLOSED | Task vẫn tham gia mẫu hiệu suất nếu đủ điều kiện |
| actual_hours = 0 | Hiển thị 0, không chia cho 0, loại khỏi mẫu PerfScore |
| PM xem hồ sơ nhân viên từng làm dự án PM khác | Chỉ thấy thông tin tham gia được phép, không truy cập task chi tiết |

## 13. Liên kết D01–D38 với nội dung V1

| Nhóm mã | Mục xử lý | Trạng thái review |
|---|---|---|
| D01–D03 | 3 | Phạm vi/actor/user–employee đề xuất rõ; không tự cấp quyền đọc mọi project cho ADMIN |
| D04 | 4 | Tách chính thức/pending, thêm version và xử lý yêu cầu hồ sơ còn thiếu |
| D05–D08 | 5 | Đầy đủ project/task transitions, pause/cancel, khóa spec |
| D09–D12 | 3, 6, 7 | Availability, batch, thay người/thu hồi và override gắn run |
| D13–D21 | 7 | Bộ công thức thử nghiệm, replay, freshness, batch semantics, mẫu số đánh giá |
| D22–D23 | 8 | Publication tệp tách duyệt kỹ năng, luồng lỗi, không gọi dictionary là ML |
| D24–D27 | 3, 4, 6, 11 | Ràng buộc, audit, chống cascade mất lịch sử, nghĩa project member |
| D28–D30 | 9 | Phiên, lịch có version, tránh overlap, ý nghĩa dữ liệu đo |
| D31–D35 | 10 | Dependency, API còn thiếu, auth/CSRF/logout/concurrency |
| D36–D38 | 14 | Kế hoạch đồng bộ tài liệu và nghiệm thu; chưa có số liệu thực nghiệm |

## 14. Đồng bộ báo cáo và tiêu chí nghiệm thu — D36–D38

Sau xác nhận, cập nhật tài liệu thiết kế hiện hành theo V1; `CLAUDE.md` làm index, không chứa bản rule khác. Bộ nguồn gốc `docs/thesis-context` đang quy định chỉ đọc: lập danh sách sửa và bản báo cáo làm việc riêng/nguồn LaTeX đầy đủ để người viết đồng bộ; không coi PDF cũ là bản thiết kế đã cập nhật.

Repository hiện không chứa `DoAn.tex` gốc, thư mục hình `Hinhve` hoặc bộ bibliography đầy đủ dùng để build báo cáo. Có thể đọc ảnh trong PDF nhưng chưa thể tái biên dịch/chỉnh đồng bộ toàn bộ báo cáo từ các `.tex` rời hiện có. Cần bộ nguồn báo cáo khi bước chỉnh sửa/bản build được yêu cầu.

Chương 2 cần bổ sung use case duyệt năng lực/CV, trả/nhận lại việc, pause/resume, thay người, hủy/đóng dự án và chuyển PM; sửa số bước luồng thay thế UC003 đang trỏ tới sau bước lưu thay vì bước kiểm tra. Sửa ví dụ 3,5h thành 4,5h; bỏ chứng chỉ ở cả văn bản và biểu đồ theo phạm vi đã duyệt.

Chương 3 trình bày công thức, giả định và lựa chọn công nghệ có nguồn. Chương 4 đồng bộ ERD, quan hệ, enum, DTO, state/sequence/package diagrams. Chương 5 dùng dữ liệu đo thật, phân biệt khảo sát người dùng thật với tình huống giả định; không viết đã phỏng vấn/chạy thử nếu chưa có bằng chứng. Không suy việc đã trích dẫn TaskAllocator là đã chứng minh công thức 0,7/0,2/0,1 của đồ án.

| Quy tắc | Use case | Service/API | Dữ liệu | Ca kiểm thử |
|---|---|---|---|---|
| Chỉ kỹ năng chính thức | Duyệt năng lực, UC003.1 | approve request, matching input | skill requests + official skills | Pending/stale/concurrent approval |
| PM quyết định | UC003/003.1/003.2 | confirm batch | assignments/run/events | Top1/Top3/manual/rollback/retry |
| Quyền theo project | UC001–004 | mọi action project/task | owner + current actor | PM chéo project, chuyển PM |
| Đúng phiên làm việc | UC004 + pause/thay người | state transition | logs/history/calendar | Return/start/overlap/reassign/holiday |
| CV chưa duyệt không đổi năng lực | Duyệt CV | parse/map/approve | CV + request | Parse failure/duplicate mapping/reject |

M1–M6 đạt khi chạy được chuỗi nghiệp vụ và các ca cạnh tranh/rollback cốt lõi; bản đồ án đầy đủ đề xuất gồm M7 CV và M8 đánh giá. `go test` không có test không được tính là nghiệm thu. Unit tests kiểm tra scorer/lịch, integration kiểm tra ràng buộc/transaction, API tests kiểm tra hợp đồng/quyền, browser E2E kiểm tra thao tác người dùng; httptest riêng lẻ không thay browser E2E.

Mục tiêu đo ban đầu để trao đổi: API thông thường p95 ≤ 1 giây khi DB đã warm; recommendation batch 20 task × 100 nhân sự ≤ 3 giây; CV text trong giới hạn tệp ≤ 30 giây với trạng thái xử lý rõ. Đây là mục tiêu đề xuất chưa đo, phải ghi máy chạy, concurrency, dataset và cold/warm riêng. Không lấy con số cold start từ tài liệu cũ làm cam kết hiện tại. Dataset đánh giá và người đánh giá thực tế còn cần xác nhận; chưa thể bảo đảm chất lượng thuật toán khi chưa có chúng.

## 15. Bộ quyết định đã được người dùng xác nhận

1. **Phạm vi và tác nhân:** một tổ chức; EMPLOYEE nhận task; PM quản lý; ADMIN không mặc định xem mọi nội dung dự án, có quyền chuyển PM khi cần; bỏ Certificates và cây task nhiều cấp.
2. **Duyệt dữ liệu:** tách hồ sơ chính thức và request; CV mới chờ duyệt trước khi thay CV chính thức; mọi thay đổi kỹ năng có vết kiểm tra và không ghi đè pending.
3. **Workflow/thời gian:** thêm PAUSED/CANCELLED; một task IN_PROGRESS mỗi nhân viên; PM trả về ASSIGNED; thời gian là số giờ theo trạng thái và lịch, không phải chấm công xác minh.
4. **Phân công:** all-or-nothing; thay người cắt phiên, thu hồi có event; khóa estimate/requirements sau phân công đầu; không reopen DONE/project terminal trong V1.
5. **Recommendation:** mandatory phải đủ mức ở nhánh AI; workload là tải tồn đọng, bỏ bonus member; ba thành phần là cấu hình thử nghiệm cần đánh giá; snapshot đủ replay và quyết định PM truy vết về run.
6. **CV/học thuật:** PDF text/DOCX, local storage, pipeline dựa trên luật; giảng viên không bắt buộc ML. Công thức phải được đánh giá, không gọi trọng số thử nghiệm là tối ưu.
7. **Auth:** giữ JWT/refresh nhưng kiểm tra phiên/trạng thái hiện hành cho API nghiệp vụ để thu hồi ngay, chấp nhận truy vấn DB; bổ sung CSRF, logout đúng path và đổi/reset mật khẩu.

Các mục trên đã được đồng ý như **một bộ thiết kế liên kết**. Nếu đổi một mục, cập nhật các hệ quả phụ thuộc. Dataset/người đánh giá thực tế vẫn chưa có bằng chứng; không thể suy ra từ source code. Danh sách việc còn lại ở review sau cập nhật, không yêu cầu xác nhận lại toàn bộ V1.
