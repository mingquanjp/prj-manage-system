# Brief sửa giao diện — nghiệp vụ hiện hành 20/10/2026

**Tiếp nối 27/10/2026:** đề nghị kỹ năng nhân viên chỉ bổ sung/điều chỉnh, không chọn loại/gỡ; màn Giám đốc theo UI-06/07 tại [screen-contracts](screen-contracts.md), làm rõ nhận diện, mức đề nghị, đóng góp, tình trạng căn cứ và lịch sử trước quyết định. Không đổi schema hoặc thêm quyền gỡ.

Sửa prototype DATN hiện có theo [ma trận](screen-contracts.md) và [Guidelines](Guidelines.md). Giữ phong cách, component và các tương tác còn đúng. Nguồn nghiệp vụ: [phạm vi](../../../.claude/docs/product_scope.md), [phân quyền](../../../.claude/docs/roles_permissions.md), [workflow dự án](../../../.claude/docs/project_task_assignment_workflow.md), [nhật ký](../../../.claude/docs/work_effort_workflow.md), [quyết định mới](../../business-baseline-2026-10-20.md). Nội dung này thay brief/prompt và bản import trước 20/10 trong Make.

## Hành trình chính

Giám đốc tạo dự án và chọn PM → PM chỉ định PTCM → PTCM tạo nhóm/task/yêu cầu kỹ năng → PTCM phân công thủ công hoặc theo gợi ý → PM bắt đầu triển khai → nhân viên thực hiện, ghi nhật ký, nộp → PTCM nghiệm thu → PM hoàn thành/đóng khi đủ điều kiện.

Kế hoạch chi tiết chia sẻ ngoài hệ thống. Theo dõi tiến độ diễn ra xuyên suốt; một task đạt không tự hoàn thành dự án. Khi trả kết quả chưa đạt, nhân viên chỉnh sửa rồi nộp lại.

## Vai trò và điều hướng

| Tư cách hiện hành | Chức năng được hiển thị |
|---|---|
| Admin | Tài khoản: tạo kèm hồ sơ ban đầu, khóa/mở, cấp/thu hồi vai trò, reset theo chính sách được đặc tả |
| HR | Hồ sơ nhân sự/CV; yêu cầu thông tin cá nhân; phòng ban, vị trí và cấp bậc |
| Giám đốc | Tạo dự án kèm PM, thay PM; xét duyệt kỹ năng; danh mục kỹ năng gồm nhóm và kỹ năng |
| PM theo dự án | Theo dõi/cập nhật dự án theo quyền, chỉ định PTCM; bắt đầu/hoàn thành/đóng/hủy; tra cứu nghề nghiệp |
| PTCM theo dự án | Nhóm/task/requirements; phân công/thay/thu hồi; nghiệm thu; duyệt hiệu chỉnh; tra cứu nghề nghiệp |
| Người có hồ sơ | Hồ sơ/CV/yêu cầu của mình; công việc đang giao và đóng góp lịch sử của mình |

Theo chốt tiếp nối 28/10 tại [baseline](../../business-baseline-2026-10-20.md), Admin thấy mã nhân viên và tên đăng nhập dự kiến chỉ đọc; mã nhân viên được tự cấp, tên đăng nhập theo mã chỉ tạm cho demo, chưa chốt chính sách auth backend. Nhãn UI dùng “Quyền hệ thống”, HR hiển thị “Quản lý nhân sự”; vị trí/cấp bậc không tự cấp quyền.

Một tài khoản có thể có nhiều vai trò hệ thống; PM/PTCM là trách nhiệm theo từng dự án. Nhân viên không phải vai trò cần Admin gán trong user_roles. Cho thấy trách nhiệm trong header dự án; PM không kiêm PTCM không có hành động chuyên môn. Người kiêm PM/PTCM có cả hai nhóm hành động và vẫn không nhận task tại dự án đó.

Nếu dùng bộ chọn người/role để demo, ghi rõ đây là điều khiển mô phỏng; đổi demo identity không phải chức năng tự cấp quyền trong sản phẩm. Menu hợp nhất theo quyền hiện hành, không nhân đôi hồ sơ kỹ năng hoặc bắt người kiêm nhiệm đăng nhập hai tài khoản.

## Quy tắc thể hiện xuyên suốt

- Gợi ý toàn công ty đủ điều kiện; không roster, thêm thành viên trước giao hoặc luồng tìm ngoài đội. Số dự án đang có việc chỉ tham khảo; không chặn dự án thứ ba.
- Chọn tất cả task đủ điều kiện theo bộ lọc gồm các trang chưa hiện; không chọn sẵn task/ứng viên. Một lần chạy cho tập chọn, mỗi task một run và 0–3 ứng viên.
- Kết quả gợi ý có thời điểm, hạng/điểm/giải thích minh họa. Công thức cuối chưa chốt. Tra cứu thủ công không có matching score/rank; dùng cùng nguồn chính thức.
- Draft không tăng tải chính thức. Tổng tải dự kiến một người = tải hiện hành + tổng estimate các task đang chọn cho người đó. Commit thành công cập nhật state toàn batch và tải cho lượt sau.
- Chọn hạng khác Top-1 hoặc chuyển manual sau run cần lý do; manual thiếu/dưới mức bắt buộc cần lý do override. Không override người inactive/locked hoặc PTCM cùng dự án.
- Bỏ checkbox/lý do chấp nhận lịch sử R05 cũ. Run cũ bất biến, có thể chạy lại; khi confirm vẫn kiểm tra skill/spec/quyền/trạng thái hiện hành. Tải mới và thứ hạng cũ có nhãn thời điểm riêng.
- Task thuộc một nhóm, một người chịu trách nhiệm. Mô tả gồm đầu ra; tiêu chí nghiệm thu riêng. Không tạo trường deliverable độc lập; spec/estimate/requirements khóa sau lần giao đầu, kể cả đã thu hồi.
- Nhân viên tự nộp tóm tắt bắt buộc, liên kết minh chứng tùy chọn. Không ép có nhật ký hoặc khai thiếu giờ để được nộp; nhật ký rỗng ghi “Chưa có thời gian ghi nhận”.
- Ghi effort ngày/phút/nội dung; theo chốt 21/10 chỉ tạo/sửa/xóa trực tiếp khi task/project IN_PROGRESS, ASSIGNED cần Bắt đầu, PAUSED cần Tiếp tục; vẫn đọc log đã có. Không timer hoặc bước xác nhận nhận việc riêng. Trên 480 phút/ngày cần lý do, trên 1440 chặn xuyên dự án.
- Nhật ký đã khóa hoặc lượt giao cũ dùng yêu cầu ADD/UPDATE/DELETE; PTCM hiện hành duyệt. Không PM ghi/gửi/duyệt thay từ tư cách PM.
- Giám đốc duyệt kỹ năng; HR duyệt fullname/phone/address; CV không được phê duyệt. Kỹ năng chờ duyệt và trích xuất không trộn vào năng lực chính thức.
- Địa chỉ phải bổ sung khi hoàn thiện hồ sơ, không bắt lúc Admin tạo tài khoản. Hồ sơ cá nhân/HR hiện “Chưa cập nhật” cùng nhắc bổ sung khi chưa có; không chặn đăng nhập/công việc. Form đề nghị mở bằng nút khi cần; bổ sung/sửa phải có nội dung, không cho xóa trắng địa chỉ đã có. Không thêm trạng thái hoàn thiện hồ sơ.
- Danh mục kỹ năng không có trạng thái INACTIVE. Xóa danh mục còn liên kết bị chặn, không thêm archive/soft-delete chỉ để đủ CRUD.
- PM/PTCM tra cứu nghề nghiệp giới hạn; không email/phone/address/CV gốc/raw task hoặc log dự án khác. Giám đốc đọc đúng căn cứ xét duyệt, không tự mở quyền xem địa chỉ; Admin không là người xem tất cả.

## Cách tổ chức màn hình

Dùng các trang danh sách/chi tiết cùng dialog hoặc drawer cho hành động liên quan. Danh mục nhóm kỹ năng và kỹ năng có thể chung một trang; phòng ban, vị trí và cấp bậc có thể chung trang với ba nhóm danh mục. Đây là tổ chức UI, không thêm tầng UC.

Dự án có chi tiết/nhóm công việc/tiến độ và các hành động theo trách nhiệm. Công việc cá nhân có thể giữ Kanban; kéo thả không thay thế thao tác Start/Pause/Resume/Nộp/nghiệm thu. Không bổ sung dashboard KPI, sprint, timer, roster, reviewer riêng, duyệt CV hoặc tích hợp Slack/GitHub.

## Dữ liệu và prototype

UI tiếng Việt, trạng thái có nhãn dễ hiểu. Dùng mock data nhất quán giữa danh sách, chi tiết, lượt phân công, log và hàng chờ. Hiển thị ghi chú “Dữ liệu mô phỏng” ở vị trí chung; không lặp cảnh báo kỹ thuật trên từng thẻ.

Đặc tả và prototype không chứng minh backend đã hoạt động. Các mục còn mở theo [README](README.md) phải được ghi vào kết quả đợt sửa; không tự điền công thức/ngưỡng/quyền còn thiếu.
