# Hồ sơ nhân sự và cập nhật thông tin cá nhân

**Đồng bộ quyết định 15/10/2026:** [sổ quyết định, phần hiện hành và Bàn giao](../../docs/design/database-review-decisions-2026-10-14.md) ưu tiên cho nội dung đã thay thế; [dictionary/DBML hợp nhất](../../docs/design/database-proposal-2026-10-14.md) là thiết kế 33 bảng (cấp bậc theo chốt 26/10/2026). Đây là thiết kế DB, chưa triển khai DDL/migration hoặc backend nghiệp vụ; tiến độ frontend mô phỏng xem [sổ công việc](../../web/README.md#sổ-công-việc).


Nghiệp vụ đích theo [baseline 2026-10-12](../../docs/business-baseline-2026-10-12.md); chưa phải thiết kế DB/API hoặc chức năng đã triển khai.

## 1. Phạm vi quyền

| Chủ thể | Hành vi |
|---|---|
| Nhân viên | Xem hồ sơ của mình, gửi yêu cầu cập nhật thông tin cá nhân, xem kết quả và lý do từ chối |
| HR | Xem/cập nhật/vô hiệu hóa hồ sơ nhân sự; xét duyệt yêu cầu cập nhật thông tin cá nhân; quản lý CV nhân sự |
| Giám đốc | Xem hồ sơ/căn cứ theo mục đích đánh giá và xét duyệt kỹ năng; không thay HR duyệt thông tin cá nhân chỉ vì là Giám đốc |
| Quản trị viên | Quản lý tài khoản, tạo tài khoản kèm hồ sơ ban đầu trong một lần; không mặc nhiên sửa hồ sơ hành chính về sau, duyệt thông tin cá nhân hoặc kỹ năng |
| PM | Tra cứu thông tin nghề nghiệp giới hạn để theo dõi/lựa chọn PTCM; không duyệt hồ sơ từ trách nhiệm PM |
| PTCM | Tra cứu thông tin nghề nghiệp giới hạn để chọn người thực hiện; không sửa hồ sơ, không tải CV gốc người khác từ quyền tra cứu |

Người kiêm nhiệm chỉ dùng quyền tương ứng còn hợp lệ. Không có vai trò là superset mọi quyền. Cập nhật hồ sơ không bao gồm sửa kỹ năng chính thức hoặc cấp quyền hệ thống/dự án.

## 2. Yêu cầu cập nhật thông tin cá nhân

1. Nhân viên đề nghị thay đổi thông tin thuộc phạm vi cá nhân được phép.
2. Hệ thống lưu yêu cầu riêng; hồ sơ chính thức chưa thay đổi.
3. HR xem nội dung hiện tại và đề nghị, đối chiếu căn cứ phù hợp với thông tin cần sửa.
4. HR chấp thuận hoặc từ chối. Chấp thuận cập nhật đúng nội dung được duyệt; từ chối giữ nguyên hồ sơ và nêu lý do để nhân viên hiểu.
5. Hệ thống giữ người xử lý, thời điểm, nội dung thay đổi và kết quả; nhân viên xem kết quả. Giữ kiểm tra quyền, đối chiếu before_values theo trường dưới khóa và ràng buộc dữ liệu, không ghi đè từ yêu cầu cũ khi hồ sơ đã đổi.

Chờ xử lý không thay dữ liệu đang sử dụng. Không tự thêm tầng Giám đốc duyệt tiếp, trạng thái mới, hạn mức gửi theo tuần/tháng hoặc yêu cầu mọi thay đổi phải có giấy tờ. Không sao chép máy móc quy tắc một PENDING/người/kỹ năng sang yêu cầu hồ sơ.

**Đã chốt, bổ sung 20/10/2026:** nhân viên đề nghị sửa fullname/phone/address; email là email công ty, thông tin cơ cấu do HR cập nhật trực tiếp. Không đồng nhất username với email. Không nhận thay đổi role, trách nhiệm dự án hoặc kỹ năng chính thức qua luồng này; thông tin tổ chức thuộc quản lý HR, không tự cho nhân viên đổi.

### Địa chỉ khi hoàn thiện hồ sơ

Theo [quyết định 20/10](../../docs/business-baseline-2026-10-20.md), địa chỉ là thông tin phải bổ sung khi hoàn thiện hồ sơ, không bắt buộc Admin nhập tại bước tạo tài khoản. `employees.address` dùng TEXT nullable, mặc định NULL để biểu thị chưa bổ sung. Không có địa chỉ giả/default và không thêm trạng thái hoàn thiện hồ sơ.

- Hồ sơ cá nhân và hồ sơ HR hiển thị “Chưa cập nhật” cùng lời nhắc bổ sung nếu address NULL. Thiếu địa chỉ không chặn đăng nhập, nhận việc hoặc thực hiện task.
- Nhân viên bổ sung/sửa qua yêu cầu hiện có; HR duyệt hoặc cập nhật trực tiếp theo quyền. Pending/từ chối không đổi địa chỉ chính thức.
- Khi cung cấp hoặc đề nghị đổi địa chỉ phải có nội dung sau trim; không cho xóa địa chỉ đã có về NULL/rỗng/chỉ khoảng trắng, kể cả qua HR cập nhật trực tiếp.
- Patch không có khóa address nghĩa là không đổi, không bắt yêu cầu chỉ sửa fullname/phone phải bổ sung địa chỉ ngay. Khi thêm lần đầu, before_values.address có thể NULL; proposed_values.address phải có nội dung. HR đối chiếu trước/đề nghị/hiện tại theo đúng trường, giữ kiểm tra xung đột và lịch sử.
- Form mặc định chỉ đọc; “Đề nghị cập nhật” mở modal/drawer có hủy. Giá trị địa chỉ chưa thay đổi được bỏ khỏi patch; xóa trắng trường đã có báo lỗi, không gửi như một yêu cầu xóa.
- Địa chỉ là thông tin hành chính riêng tư; không có trong tra cứu nghề nghiệp, ứng viên hoặc snapshot gợi ý. Không mở thêm quyền xem địa chỉ cho PM/PTCM/Giám đốc từ tư cách tra cứu/xét kỹ năng; Admin chỉ nhập nếu có ở bước tạo, không được sửa hành chính thường xuyên.

## 3. Cập nhật trực tiếp và vô hiệu hóa

Theo [quyết định 20/10/2026](../../docs/business-baseline-2026-10-20.md), Admin tạo tài khoản kèm hồ sơ nhân viên ban đầu trong một lần. HR quản lý và cập nhật hồ sơ về sau. Thông tin ban đầu theo dictionary hiện hành; không thêm bước HR duyệt tạo hoặc hồ sơ/tài khoản tạm tách rời.

HR cập nhật hồ sơ trực tiếp là thao tác quản lý riêng, không giả lập nhân viên đã gửi yêu cầu. Giữ lịch sử thay đổi và các kiểm tra dữ liệu chung. Quyền hành chính không cho phép sửa kỹ năng chính thức để bỏ qua Giám đốc xét duyệt.

Vô hiệu hóa hồ sơ (HR) khác với khóa tài khoản (Admin). Giữ nguyên điều kiện tài khoản và hồ sơ phải active để sử dụng nghiệp vụ; không đồng nhất hai trạng thái hoặc tự mở lại trạng thái còn lại. Không xóa lịch sử công việc/năng lực khi vô hiệu hóa. Không phát sinh bàn giao tự động hoặc quyền thay PM cho HR.

Trước khi vô hiệu hóa, kiểm tra công việc và trách nhiệm dự án còn mở. Nếu còn, chưa cho vô hiệu hóa và hiển thị các mục cần người có thẩm quyền xử lý trước. Đây là điều kiện của thao tác, không thêm quy trình phê duyệt/bàn giao. Khi cần ngăn truy cập khẩn cấp, Admin vẫn được khóa tài khoản theo quy tắc hiện hành.

## 4. Liên hệ CV và kỹ năng

HR quản lý CV nhân sự theo [quy trình CV](cv_extraction_approval_workflow.md); nhân viên quản lý CV cá nhân. Không có phê duyệt CV. Quyền quản lý tệp không cấp quyền công nhận kỹ năng trích xuất.

Hai luồng kỹ năng do Giám đốc xử lý theo [quy trình kỹ năng](skill_approval_workflow.md), không qua tầng HR tiền kiểm bắt buộc. HR không được đọc toàn bộ task/nhật ký/căn cứ dự án chỉ vì quản lý hồ sơ; giới hạn theo [phân quyền](roles_permissions.md).

## 5. UC cần đồng bộ

- Nhân viên: Xem hồ sơ cá nhân; Yêu cầu cập nhật thông tin cá nhân; Quản lý CV cá nhân; Yêu cầu cập nhật kỹ năng.
- HR: Xem hồ sơ nhân sự; Cập nhật hồ sơ nhân sự; Vô hiệu hóa hồ sơ nhân sự; Quản lý CV nhân sự; Xét duyệt cập nhật thông tin cá nhân.
- Giám đốc: Xem hồ sơ nhân sự; Xét duyệt yêu cầu cập nhật kỹ năng; Xét duyệt kỹ năng từ CV.

Xem hồ sơ bao gồm phần năng lực được phép xem; không buộc tạo thêm UC chỉ để biểu diễn một phần màn hình. Tên UC tổng quan không cấp chung mọi quyền của HR cho Giám đốc hoặc ngược lại.

HR quản trị departments/positions/job_levels; cấp bậc là danh mục do công ty đặt tên, tách biệt vị trí và mức kỹ năng. Theo chốt 26/10/2026, `employees.job_level_id` nullable thay `seniority_level` nhập tự do. HR chọn mục có sẵn hoặc để “Chưa cập nhật”; không enum cố định Intern/Junior/Senior cho mọi công ty, không suy từ tên vị trí, không thứ tự/hệ số hoặc quan hệ theo từng vị trí/phòng ban. HR thêm/sửa tên/xóa cấp bậc như danh mục tổ chức: tên bắt buộc, không trùng, chặn xóa khi còn hồ sơ tham chiếu; giữ ID và lịch sử đã chụp khi đổi tên. Không tự cấp quyền hoặc suy mức kỹ năng. Nhân viên không đề nghị đổi cấp bậc qua yêu cầu thông tin cá nhân.

**Đối chiếu UC danh mục dùng chung (16/10, cập nhật 26/10/2026):** HR thực hiện **Quản lý phòng ban, vị trí và cấp bậc**; Giám đốc thực hiện **Quản lý danh mục kỹ năng**, bao gồm nhóm kỹ năng và kỹ năng. Đây là hai UC hỗ trợ riêng ở tổng quan, không phải các bước của cập nhật hồ sơ cá nhân hoặc xét duyệt kỹ năng. Không bắt buộc sơ đồ phân rã riêng. Xem tổng hợp tại [phạm vi sản phẩm](product_scope.md#2-tác-nhân-và-chức-năng); không thay đổi quyền quản trị hoặc quy tắc bảo vệ dữ liệu đã chốt.
