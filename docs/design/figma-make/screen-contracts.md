# Ma trận màn hình, dữ liệu và hành động

Ngày 20/10/2026. Đọc cùng [brief](ui-brief.md), [Guidelines](Guidelines.md) và [quy tắc](../../../.claude/docs/domain_rules.md). Các mã UI dưới đây phục vụ kiểm tra thiết kế, không tăng danh mục UC. Một hàng có thể được thể hiện bằng trang, tab hoặc drawer dùng chung.

## 1. Danh mục màn hình

| ID | Màn hình / tư cách | Nội dung và hành động chính | Kiểm tra và trạng thái cần có | Nguồn |
|---|---|---|---|---|
| UI-01 | Tài khoản — Admin | Danh sách tài khoản, username/trạng thái/vai trò; tạo tài khoản kèm hồ sơ; khóa/mở; cấp/thu hồi vai trò | Tạo đồng thời user/employee; không tài khoản mồ côi; bảo vệ Admin active cuối; khóa khẩn cấp không tự bàn giao. Chính sách cấp/reset mật khẩu còn mở | [Quyền](../../../.claude/docs/roles_permissions.md), [quyết định 20/10](../../business-baseline-2026-10-20.md) |
| UI-02 | Hồ sơ nhân sự — HR; góc nhìn Giám đốc theo mục đích | HR xem/cập nhật hành chính, quản lý CV, vô hiệu hóa; Giám đốc xem hồ sơ/căn cứ xét kỹ năng | HR không sửa official skill; Giám đốc không có nút HR từ vai trò Giám đốc. Vô hiệu hóa bị chặn khi còn task/trách nhiệm mở; khác khóa tài khoản | [Hồ sơ](../../../.claude/docs/employee_profile_workflow.md) |
| UI-03 | Hồ sơ cá nhân — chủ hồ sơ | Xem hồ sơ, nút mở đề nghị sửa fullname/phone/address, lịch sử kết quả; email công ty/cơ cấu chỉ đọc | Address thiếu hiện “Chưa cập nhật”/nhắc bổ sung; không xóa trắng. Pending không đổi hồ sơ; HR từ chối có lý do; không thêm rút/bổ sung yêu cầu như luồng kỹ năng | [Hồ sơ](../../../.claude/docs/employee_profile_workflow.md), [API](../../../.claude/docs/api_design.md) |
| UI-04 | Yêu cầu thông tin cá nhân — HR | Danh sách pending/lịch sử; so giá trị hiện tại, trước và đề nghị; chấp thuận/từ chối | Dữ liệu nền đổi thì báo xung đột, không ghi đè. Từ chối cần lý do; chấp thuận có thể không lý do | [Hồ sơ](../../../.claude/docs/employee_profile_workflow.md) |
| UI-05 | CV cá nhân / CV nhân sự — chủ hồ sơ, HR; căn cứ Giám đốc | Upload, danh sách phiên bản, tải theo quyền; chủ hồ sơ xác nhận bản hiện tại; trạng thái trích xuất và retry kỹ thuật | PDF text/DOCX ≤10 MiB; tệp hợp lệ dùng được khi parse lỗi; HR upload không chọn hiện tại thay; không nút duyệt/từ chối CV | [CV](../../../.claude/docs/cv_extraction_approval_workflow.md) |
| UI-06 | Yêu cầu kỹ năng — chủ hồ sơ / Giám đốc | ADD/UPDATE/REMOVE; kỹ năng, mức/năm theo dữ liệu hiện hành, giải trình; nhiều căn cứ task, tối đa một CV trực tiếp, tệp tùy chọn; bổ sung khi pending, rút, xem quyết định | Không cần task nội bộ; tối đa một pending/người/skill; terminal bất biến; không tự nâng kỹ năng từ CV. Loại/giới hạn tệp ngoài CV chưa chốt | [Kỹ năng](../../../.claude/docs/skill_approval_workflow.md) |
| UI-07 | Xét duyệt kỹ năng MANUAL/CV — Giám đốc | So official/đề nghị; rubric bốn mức; căn cứ đúng đóng góp; map dòng trích xuất; duyệt/từ chối có lý do | Không tự duyệt; thiếu chuyên môn cần ghi nhận ý kiến ngoài hệ thống; không thêm cấp duyệt. Pending có bổ sung thì màn hình duyệt cũ bị chặn. NO_CHANGE khác APPROVED | [Kỹ năng](../../../.claude/docs/skill_approval_workflow.md), [CV](../../../.claude/docs/cv_extraction_approval_workflow.md) |
| UI-08 | Danh mục tổ chức — HR | Phòng ban và vị trí công việc: danh sách, thêm, sửa, xóa khi không còn tham chiếu | Không cây phòng ban mới; không suy role hay mức kỹ năng từ vị trí/cấp bậc; lỗi tham chiếu hiển thị rõ | [Quyền](../../../.claude/docs/roles_permissions.md), [dictionary](../database-step5b-supporting-dictionary-2026-10-14.md) |
| UI-09 | Danh mục kỹ năng — Giám đốc | Nhóm kỹ năng và kỹ năng; kỹ năng thuộc nhóm; thêm/sửa/xóa khi hợp lệ | Không INACTIVE; không đổi nghĩa kỹ năng đã dùng hoặc xóa còn tham chiếu. Không đồng nhất CRUD danh mục với công nhận kỹ năng một người | [Kỹ năng](../../../.claude/docs/skill_approval_workflow.md), [dictionary](../database-step5b-supporting-dictionary-2026-10-14.md) |
| UI-10 | Dự án — Giám đốc/PM/PTCM theo quyền | Giám đốc tạo với mã/tên/ngày/ưu tiên và PM bắt buộc, thay PM; PM chỉ định PTCM/theo dõi/vòng đời; PTCM xem và điều hành chuyên môn | Người tạo không tự thành PM; PM không tự thành PTCM. Không roster/sprint/gửi duyệt plan. Thay PM không chờ bàn giao; PM/PTCM mới không đổi actor lịch sử | [Dự án](../../../.claude/docs/project_task_assignment_workflow.md) |
| UI-11 | Cấu trúc nhóm và task — PTCM | Nhóm trong dự án; task gồm mã/tên/mô tả có đầu ra/criteria/estimate/priority, ngày tùy chọn đủ cặp; ≥1 requirement, skill/level/weight/mandatory | Unique mã task/project, nhóm/project, skill/task; estimate >0; ngày trong project. Weight do PTCM đặt nhưng miền còn mở. Không deliverable riêng; không sửa frozen scope sau giao | [Dự án](../../../.claude/docs/project_task_assignment_workflow.md), [quy tắc](../../../.claude/docs/domain_rules.md) |
| UI-12 | Phân công — PTCM | Task đủ điều kiện → gợi ý theo task hoặc manual → draft → confirm; thay/thu hồi theo trạng thái | Toàn công ty; lọc PTCM cùng dự án; mandatory gate ở gợi ý; manual override kỹ năng có lý do; batch atomic; run có thời điểm; không R05 ack | [Gợi ý](../../../.claude/docs/ai_recommendation_workflow.md), [phân công](../../../.claude/docs/project_task_assignment_workflow.md) |
| UI-13 | Tra cứu nghề nghiệp — PM/PTCM | Danh sách/tìm kiếm riêng; nhận diện/công tác, official skills, kinh nghiệm/kết quả và tải tổng hợp được phép; ngữ cảnh task tái sử dụng trong manual | Không cấp quyền cho nhân viên thường; không CV gốc/contact/tên task/log dự án khác. Không tự biến assignment chưa thực hiện thành kinh nghiệm | [Quyền](../../../.claude/docs/roles_permissions.md) |
| UI-14 | Công việc cá nhân và nghiệm thu — người hiện được giao / PTCM | Người thực hiện Start/Pause/Resume, log, nộp tóm tắt và link tùy chọn; PTCM xem submission/criteria và chấp nhận hoặc yêu cầu chỉnh sửa | Thiếu log không chặn nộp/nghiệm thu; snapshot nộp chỉ đọc; trả kết quả về IN_PROGRESS để sửa rồi nộp lại; không nộp thay/timer/reviewer riêng/mở lại DONE | [Thực hiện](../../../.claude/docs/project_task_assignment_workflow.md), [effort](../../../.claude/docs/work_effort_workflow.md) |
| UI-15 | Đóng góp, nhật ký và yêu cầu hiệu chỉnh — chính chủ/PTCM | Log ngày/phút/nội dung; Edit/Delete trực tiếp khi được phép; ADD/UPDATE/DELETE qua yêu cầu khi khóa; PTCM duyệt, chính chủ rút pending; thông báo nội bộ | >480 phút cần lý do, >1440 chặn xuyên dự án; không ngày tương lai; old assignment giữ tác giả; pending không đổi tổng; ADD approved nối log kết quả; không restore. CLOSED/CANCELLED chỉ đọc | [Effort](../../../.claude/docs/work_effort_workflow.md), [API](../../../.claude/docs/api_design.md) |

## 2. Hành động theo trạng thái dự án

| Trạng thái / nhãn | Hành động được phép |
|---|---|
| PLANNING — Lập kế hoạch | PTCM cấu trúc/phân công; PM bắt đầu khi có ≥1 task ASSIGNED hoặc hủy có lý do; chưa thực hiện task/ghi effort |
| IN_PROGRESS — Đang thực hiện | PTCM cấu trúc task mới/phân công/nghiệm thu/hiệu chỉnh; nhân viên thực hiện; PM theo dõi, hủy hoặc hoàn thành khi đủ điều kiện |
| COMPLETED — Đã hoàn thành | Đọc, hiệu chỉnh effort có duyệt; PM đóng khi hết pending. Không tạo/thực hiện task hoặc bình luận mới; Giám đốc vẫn thay PM theo quyền |
| CLOSED — Đã đóng | Dữ liệu nghiệp vụ chỉ đọc theo quyền; vẫn đánh dấu thông báo đã đọc |
| CANCELLED — Đã hủy | Dữ liệu nghiệp vụ chỉ đọc; giữ task DONE, hủy task mở; không tạo kết luận nghiệm thu không đạt giả |

PM hoàn thành khi ≥1 task DONE và mọi task còn lại DONE/CANCELLED. PM đóng chuyển DONE → CLOSED cùng lúc. Không tự hoàn thành dự án từ nút nghiệm thu một task. Dự án không có task hiệu lực hiển thị trạng thái rỗng, không tự 100%.

## 3. Hành động theo trạng thái task

| Trạng thái / nhãn | Người thực hiện | PTCM / PM |
|---|---|---|
| PENDING_ASSIGNMENT — Chờ phân công | Không thực hiện | PTCM phân công nếu project PLANNING/IN_PROGRESS |
| ASSIGNED — Đã phân công | Start khi project IN_PROGRESS; ghi bổ sung phần thực tế theo R01 khi còn quyền | PTCM thay/thu hồi; PM hủy có lý do |
| IN_PROGRESS — Đang thực hiện | Pause có lý do; ghi/sửa/xóa log theo quyền; nộp | PTCM thay/thu hồi; PM hủy |
| PAUSED — Tạm dừng | Resume; bổ sung log thực tế theo R01 khi còn quyền | PTCM thay/thu hồi; PM hủy |
| PENDING_REVIEW — Chờ nghiệm thu | Xem kết quả đã nộp; log khóa, dùng hiệu chỉnh | PTCM chấp nhận → DONE hoặc yêu cầu chỉnh sửa → IN_PROGRESS; phải trả lại trước thay/thu hồi; PM hủy |
| DONE — Đã nghiệm thu | Xem đóng góp, hiệu chỉnh dữ liệu cũ qua yêu cầu khi project cho phép | Không mở lại; đóng dự án đổi sang CLOSED |
| CLOSED — Đã đóng | Chỉ đọc | Chỉ đọc |
| CANCELLED — Đã hủy | Hiệu chỉnh dữ liệu cũ qua yêu cầu nếu project IN_PROGRESS/COMPLETED | Không thực hiện tiếp; giữ lịch sử |

Mọi hàng chịu quyền/trạng thái dự án ở bảng trên. Nhánh “Yêu cầu chỉnh sửa” khác “Thu hồi phân công”; không nút nhân viên tự “Trả việc”. Kéo thả Kanban không tự chuyển trạng thái.

## 4. Form và cách hiện dữ liệu

- Tạo tài khoản: username/mật khẩu ban đầu theo chính sách sẽ đặc tả; hồ sơ bắt buộc employee_code/fullname/email công ty/position_id; phone/department/seniority tùy chọn theo dictionary. Vai trò hệ thống là quyền Admin, không gán PM/PTCM ở form này. Không thêm must_change_password. HR quản lý hành chính sau tạo.
- Địa chỉ phải bổ sung khi hoàn thiện hồ sơ, không bắt Admin nhập lúc tạo. UI cá nhân/HR hiện address hoặc “Chưa cập nhật” và nhắc bổ sung, không chặn đăng nhập/nhận việc/thực hiện task. Form mặc định chỉ đọc, “Đề nghị cập nhật” mở modal/drawer có hủy; trường address không đổi thì bỏ khỏi patch, bổ sung/sửa phải có nội dung sau trim, không nút xóa hoặc gửi NULL/rỗng. Before có thể NULL khi bổ sung lần đầu; HR duyệt/ghi trực tiếp giữ kiểm tra xung đột. Không ép patch chỉ sửa fullname/phone phải kèm address; không thêm trạng thái hoàn thiện hồ sơ hoặc lộ address ở UI-12/UI-13.
- Tạo dự án: PM bắt buộc ngay trong form; thất bại không tạo dự án dở. PTCM có thể chỉ định sau, chưa có PTCM thì hiện lý do chưa thể tạo task/phân công.
- Yêu cầu kỹ năng: loại thay đổi quyết định trường đề nghị; REMOVE không ép nhập mức/năm như ADD. Phần kỹ năng chính thức và đề nghị phải tách rõ. Năm chưa biết không hiển thị thành 0.
- CV: phân biệt bản hiện tại, trạng thái xử lý trích xuất và kết quả xét kỹ năng. APPROVED/PARTIALLY_APPROVED/REJECTED/NO_CHANGES là kết quả kỹ năng, không nhãn phê duyệt tệp.
- Hiệu chỉnh: so trước/đề nghị và lý do; người gửi tự gửi. Khi dữ liệu nền đã đổi, giữ nội dung để đọc, báo xung đột và cho từ chối/rút rồi gửi mới theo quyền.
- Tổng giờ là tổng phút hiệu lực/60; làm tròn khi hiển thị. Tổng effort dự án gồm phần task hủy, báo riêng; không lấy estimate trừ actual để suy lịch rảnh.
- Thông báo chỉ dẫn tới tài nguyên còn được đọc; người đã mất quyền nhận thông báo lỗi quyền, không mở lại dữ liệu vì có link cũ.

## 5. Thành phần dùng chung và trường hợp lỗi

Dùng chung shell/menu theo quyền, bộ lọc, bảng, badge, form, drawer căn cứ, dialog xác nhận và thông báo. Mỗi màn hình cần loading/rỗng/lỗi/không có quyền, lỗi từng trường, lỗi dữ liệu đã đổi, hủy form và lưu thành công. Nút không hợp lệ hiển thị lý do khi cần giải thích; ẩn hành động hoàn toàn ngoài quyền.

Chỉ đưa trường/chỉ số đã có nguồn vào mock. Công thức/miền weight và chi tiết DTO còn mở được ghi trong kết quả sửa, không biến thành yêu cầu mới. Menu/màu/layout là quyết định giao diện có thể tinh chỉnh mà không đổi nghiệp vụ.
