# Cơ sở nghiệp vụ lịch sử — 2026-10-07

> Điểm vào cho triển khai mới: [mục lục nguồn hiện hành](README.md). Các quyết định DB, quyền hiệu chỉnh, danh mục và thông tin cá nhân được chốt sau mốc này ưu tiên phần lịch sử bên dưới.

> Các quy định PM phân công, giao thêm người nghiệm thu, PTCM nhận task và phạm vi tra cứu trong bản này đã được thay thế bởi [baseline hiện hành 2026-10-12](business-baseline-2026-10-12.md). Giữ nội dung bên dưới để truy vết quyết định, không dùng phần đã thay thế làm yêu cầu triển khai.

Đã được người dùng xác nhận trước khi đồng bộ tài liệu. Đây là **nghiệp vụ đích**, không phải xác nhận code, CSDL, API, tệp Astah hoặc báo cáo đã triển khai đúng.

Tài liệu này ưu tiên khi có khác biệt về vai trò HR/Giám đốc, hồ sơ cá nhân, CV và biểu diễn sơ đồ. [Baseline 2026-10-05](business-baseline-2026-10-05.md) và [baseline 2026-09-30](business-baseline-2026-09-30.md) tiếp tục áp dụng cho phần không bị thay thế. Chi tiết miền nằm trong `.claude/docs/`; không duy trì bản sao ở thư mục gốc.

## 1. Phạm vi và vai trò

Mô hình đề xuất cho doanh nghiệp nhỏ–vừa, không khẳng định mọi doanh nghiệp đều tổ chức như vậy. Không mở rộng thành hệ thống HRM đầy đủ. Một người có thể kiêm nhiệm; quyền được xét theo vai trò, trách nhiệm và phạm vi hiện hành, không suy từ cấp bậc hoặc tag.

| Vai trò/trách nhiệm | Phạm vi đã chốt |
|---|---|
| Quản trị viên | Tạo/cập nhật/khóa/mở khóa tài khoản, phân quyền hệ thống; không mặc nhiên có quyền nhân sự, duyệt kỹ năng hoặc điều hành dự án |
| HR — Nhân sự | Xem/cập nhật/vô hiệu hóa hồ sơ nhân sự, quản lý CV nhân sự, xét duyệt yêu cầu cập nhật thông tin cá nhân |
| Giám đốc | Xem hồ sơ và căn cứ theo mục đích đánh giá; xét duyệt kỹ năng từ CV và yêu cầu cập nhật kỹ năng; khởi tạo dự án, chỉ định/thay PM |
| PM | Điều phối dự án được giao, chỉ định phụ trách chuyên môn, phân công, theo dõi và hoàn thành/đóng dự án theo điều kiện hiện hành |
| Nhân viên | Hồ sơ và CV cá nhân, gửi yêu cầu cập nhật thông tin/kỹ năng, thực hiện công việc được giao |
| Phụ trách chuyên môn | Trách nhiệm do PM giao trong từng dự án; thiết lập công việc và yêu cầu kỹ năng, nghiệm thu, giao thêm trách nhiệm nghiệm thu |
| Người nghiệm thu | Phụ trách chuyên môn hoặc thành viên được người này giao trách nhiệm nghiệm thu trong cùng dự án |

HR không duyệt kỹ năng, không tạo dự án/chỉ định PM hoặc cấp quyền hệ thống từ vai trò HR. Giám đốc không còn là người duyệt thông tin cá nhân hoặc quản lý tệp CV từ riêng vai trò Giám đốc; vẫn được đọc căn cứ CV phù hợp để xét duyệt kỹ năng. Admin không dùng phân quyền tài khoản để giao trách nhiệm chuyên môn/nghiệm thu theo dự án.

## 2. Hồ sơ và yêu cầu cập nhật thông tin cá nhân

Nhân viên gửi yêu cầu → HR đối chiếu → chấp thuận hoặc từ chối → hệ thống lưu kết quả. Hồ sơ chính thức giữ nguyên khi đang chờ hoặc bị từ chối; khi chấp thuận chỉ cập nhật nội dung được duyệt. Nhân viên xem được kết quả và lý do từ chối. Giữ lịch sử, kiểm tra quyền/phiên bản và các ràng buộc dữ liệu áp dụng; không ghi đè từ yêu cầu lỗi thời.

HR cập nhật hồ sơ trực tiếp và HR xét duyệt yêu cầu là hai thao tác riêng. Quyền cập nhật hồ sơ hành chính không bao gồm sửa kỹ năng chính thức hoặc quyền tài khoản/dự án. Vô hiệu hóa hồ sơ khác với khóa tài khoản; giữ điều kiện hồ sơ/tài khoản active và giới hạn nghiệp vụ hiện hành, không tự phát sinh việc xóa lịch sử hoặc bàn giao tự động.

HR chỉ vô hiệu hóa hồ sơ sau khi công việc/trách nhiệm dự án còn mở đã được xử lý; hệ thống hiển thị các mục còn vướng, không thêm vòng duyệt. Admin vẫn có thể khóa tài khoản khẩn cấp. Thêm vai trò hệ thống không bị chặn chỉ vì còn công việc; chỉ chặn thu hồi quyền cần thiết cho trách nhiệm đang mở. Đổi vai trò hệ thống không tự thay trách nhiệm trong dự án; thay PM vẫn do Giám đốc.

**Danh sách trường cần duyệt chưa chốt.** Danh sách tên/email/điện thoại trong thiết kế cũ không còn được coi là allowlist đã chốt. Không mặc định mọi trường đều phải duyệt; cũng không mặc định trường nào được tự sửa. Role, trách nhiệm dự án và kỹ năng chính thức không đi qua yêu cầu sửa thông tin cá nhân. Chi tiết: [quy trình hồ sơ](../.claude/docs/employee_profile_workflow.md).

## 3. Kỹ năng và CV

Giữ hai luồng độc lập: (1) hệ thống trích xuất kỹ năng từ CV để Giám đốc đối chiếu; (2) nhân viên chủ động gửi đề nghị kỹ năng và giải trình, không cần CV mới. Không bắt nhân viên gửi thủ công từng mục trích xuất; không tự công nhận năng lực từ CV.

Giám đốc quyết định dựa trên mô tả mức, giải trình, đóng góp/kết quả công việc và nhận xét liên quan; cần ý kiến chuyên môn phù hợp nếu không đủ khả năng đánh giá. Không có tầng HR duyệt trước. Tệp/liên kết tùy chọn; thiếu căn cứ không đồng nghĩa không có kỹ năng. Không tự duyệt, cả chấp thuận/từ chối có lý do; một PENDING/người/kỹ năng, bổ sung khi PENDING, terminal bất biến, bổ sung sau bằng yêu cầu mới, không quota tuần/tháng. Giữ R01–R05 và quy tắc catalog INACTIVE.

Nhân viên quản lý CV cá nhân; HR quản lý CV nhân sự, upload thay có ghi người thao tác. Giữ cơ chế nhân viên xác nhận CV hiện tại; không tự thêm quyền HR chọn thay CV hiện tại khi chưa chốt. Không phê duyệt/công bố CV bằng một vòng duyệt. Tệp hợp lệ vẫn sử dụng được nếu trích xuất lỗi. Tối đa một CV hiện tại/người, giữ nguồn/lịch sử; CV mới không tự sửa/xóa kỹ năng chính thức hoặc quyết định/yêu cầu đang tồn tại. Giám đốc duyệt kỹ năng, không duyệt tệp.

Quyền xem hồ sơ theo mục đích: HR phục vụ quản lý hành chính; Giám đốc phục vụ đánh giá kỹ năng; PM tra cứu nghề nghiệp phục vụ phân công; nhân viên xem của mình. Quyền HR không mặc nhiên mở căn cứ task hoặc mọi dữ liệu dự án. Quyền Giám đốc đọc đúng task được dẫn chiếu không mở toàn bộ dự án, nhật ký hay bình luận. Danh sách chỉ số tổng hợp và DTO cụ thể chưa chốt.

## 4. Dự án, phân công và nghiệm thu

1. Giám đốc khởi tạo dự án, chỉ định PM. Thay PM là chọn người mới đủ điều kiện và lưu nguyên tử; thất bại giữ PM cũ, thành công giữ lịch sử. Không có quy trình bàn giao PM riêng hoặc mở lại dự án terminal.
2. PM chỉ định phụ trách chuyên môn; có thể kiêm nhiệm nếu đủ chuyên môn.
3. Phụ trách chuyên môn thiết lập cấu trúc và đặc tả công việc, đủ yêu cầu kỹ năng, thời gian ước tính, đầu ra và tiêu chí nghiệm thu ngay khi tạo.
4. Khi cần, phụ trách chuyên môn giao trách nhiệm nghiệm thu cho thành viên đã thuộc dự án và đủ chuyên môn; không cần PM duyệt, không tự thêm người ngoài, không cần task giả, không giới hạn cứng số reviewer.
5. PM chọn công việc, tra cứu thủ công hoặc chạy gợi ý, chọn người và xác nhận. Gợi ý lấy yêu cầu → lọc kỹ năng bắt buộc từ năng lực chính thức → chấm điểm/xếp hạng. Không tự giao việc; kiểm tra điều kiện hiện hành trước khi ghi nhận, giữ nguyên tử toàn batch.
6. Nhân viên thực hiện, ghi nhật ký và tự nộp kết quả. Hệ thống lưu kết quả/chuyển chờ nghiệm thu; người có trách nhiệm nghiệm thu đánh giá. Chưa đạt: nêu nội dung cần sửa, trả đang thực hiện, **chỉnh sửa rồi nộp lại**. Đạt: ghi nhận người đánh giá/kết luận và hoàn thành công việc.
7. PM theo dõi xuyên suốt, hoàn thành/đóng dự án theo điều kiện hiện hành. Một task hoàn thành không đồng nghĩa cả dự án hoàn thành. Khởi động dự án chỉ ở lần đầu, không lặp khi giao thêm việc.

Không tự nghiệm thu, kể cả PM/phụ trách chuyên môn trực tiếp làm task. Người được giao nghiệm thu có thể chia sẻ việc, không cần chứng minh phụ trách chuyên môn đang bận; không tự có quyền tạo task, phân công, đóng dự án hoặc giao tiếp quyền nghiệm thu. Thay người thực hiện chỉ kết thúc quyền theo lượt phân công đó; quyền từ trách nhiệm khác còn hợp lệ được xét riêng.

Không thay công thức gợi ý, vòng đời task/project, frozen specification, điều kiện manual override, hiệu chỉnh effort hoặc lịch sử đóng góp trong đợt này.

## 5. Quy ước sơ đồ và đặc tả

- Tổng quan có Admin, HR, Giám đốc, PM, Nhân viên và Người dùng hệ thống cho hành vi chung. Hai actor cùng nối một UC tổng quát không có nghĩa được mọi chức năng chi tiết như nhau.
- Sơ đồ phân rã nối actor trực tiếp tới chức năng thuộc quyền. Không bắt buộc lặp oval UC tổng quát chỉ để nối các UC bằng đường thường; tiêu đề xác định phạm vi hình.
- UC Quản lý công việc được giao giữ hai extension point `Trách nhiệm chuyên môn`, `Trách nhiệm nghiệm thu`, với hành vi mở rộng tương ứng và điều kiện trách nhiệm trong dự án. Đây là cách trình bày đã thống nhất, không thay kiểm tra phân quyền.
- Phụ trách chuyên môn/người nghiệm thu không là actor toàn hệ thống cố định; có thể dùng làn theo trách nhiệm trên activity, ghi rõ phạm vi dự án.
- UC chuyên môn vẫn có Nghiệm thu công việc; xuất hiện ở hai hình là cùng chức năng, không tạo hai luồng nghiệm thu.
- Activity tổng thể chỉ vẽ luồng chính. Có thể gộp “Kiểm tra và ghi nhận phân công” với note chỉ ghi khi hợp lệ; luồng thất bại/rollback vẫn bắt buộc trong đặc tả chi tiết. Không thêm HR vào quy trình dự án khi không có thao tác của HR.
- Tên hiển thị thống nhất: “Mở khóa tài khoản”, “Xem CV hiện tại”. Ghi chú ngắn, rõ điều kiện và trách nhiệm.

## 6. Điểm chưa chốt và giới hạn đồng bộ

- Danh sách trường thông tin cá nhân cần HR duyệt; không tự giữ allowlist kỹ thuật cũ làm quyết định mới.
- Chi tiết quyền HR thao tác CV ngoài phạm vi quản lý/upload thay đã có; không tự thêm quyền xác nhận CV hiện tại thay nhân viên.
- Chi tiết thu hồi/thay trách nhiệm chuyên môn/nghiệm thu, cách thêm thành viên không nhận task, DTO và chỉ số trên màn hình Giám đốc giữ trạng thái cần đặc tả; không tạo quy trình mới ở đợt này.
- Tự duyệt time correction của PM khi trực tiếp thực hiện vẫn chưa chốt; không suy từ quyền nghiệm thu hoặc quyền HR.
- Không sửa code, migration, công thức, file Astah hoặc tài liệu gốc `docs/thesis-context/`. Thiết kế kỹ thuật cũ phải đối chiếu baseline này trước khi triển khai.
