# Cơ sở nghiệp vụ hiện hành — 2026-10-05

> Bản quyết định theo thời điểm, được bổ sung/thay thế một phần bởi [baseline 2026-10-07](business-baseline-2026-10-07.md). HR tiếp nhận hồ sơ hành chính/CV và duyệt thông tin cá nhân; Giám đốc giữ xét duyệt kỹ năng và khởi tạo/chỉ định PM. Mô tả gộp quyền nhân sự và phần chưa chốt sơ đồ dưới đây là lịch sử; dùng bản 2026-10-07 khi có khác biệt.

Đã được người dùng xác nhận. Đây là nghiệp vụ đích, không phải xác nhận code, CSDL, API, sơ đồ hoặc báo cáo đã đồng bộ.
Tài liệu này ưu tiên về vai trò, tạo dự án, thay PM, cập nhật CV và nghiệm thu. [Cơ sở 2026-09-30](business-baseline-2026-09-30.md) giữ giá trị cho R01–R05 và phần không bị thay thế.

## 1. Vai trò và phạm vi

- Quản trị viên: tài khoản và chức năng quản trị hệ thống; không mặc nhiên quản lý hồ sơ nhân sự, duyệt kỹ năng, điều hành hoặc nghiệm thu.
- Giám đốc: quản lý hồ sơ nhân sự, đánh giá/phê duyệt kỹ năng; khởi tạo dự án và giao PM.
- PM: điều hành dự án được giao, phân công nhân sự, giao phụ trách chuyên môn, theo dõi và hoàn thành/đóng theo điều kiện hiện hành.
- Phụ trách chuyên môn: trách nhiệm trong dự án, do PM giao; thiết lập cấu trúc/đặc tả công việc và tổ chức nghiệm thu.
- Người nghiệm thu: thành viên được phụ trách chuyên môn giao trách nhiệm trong dự án, không phải role cố định toàn hệ thống.
- Người thực hiện: nhân viên được giao task; có thể đồng thời là phụ trách chuyên môn hoặc PM kiêm nhiệm.

PM có thể kiêm phụ trách chuyên môn. Một người có trách nhiệm khác nhau ở các dự án. Tag chỉ biểu thị trách nhiệm; quyền phải được kiểm tra theo phạm vi, không suy từ tên chức danh hoặc tag toàn cục.
Không định nghĩa enum, bảng hoặc trường mới trong đợt này.

## 2. Luồng chính

1. Giám đốc khởi tạo dự án và giao PM quản lý; Giám đốc không mặc nhiên là PM.
2. PM giao phụ trách chuyên môn, có thể kiêm nhiệm.
3. Phụ trách chuyên môn thiết lập công việc, đủ dữ liệu bắt buộc ngay khi tạo: kỹ năng, thời gian ước tính, đầu ra và tiêu chí nghiệm thu. Không thêm trạng thái nháp.
4. Phụ trách chuyên môn có thể giao thêm trách nhiệm nghiệm thu cho thành viên đủ chuyên môn trong dự án.
5. PM chọn nhiều/tất cả công việc đủ điều kiện, tra cứu thủ công hoặc chạy gợi ý, chọn và xác nhận. Hệ thống kiểm tra và ghi nhận; gợi ý không tự giao việc.
6. Người được giao thực hiện, ghi nhật ký và nộp kết quả.
7. Người nghiệm thu hợp lệ đánh giá theo tiêu chí: chưa đạt thì nêu chỉnh sửa và trả IN_PROGRESS; đạt thì ghi nhận DONE, người đánh giá và kết luận.
8. PM điều phối xuyên suốt, hoàn thành và đóng dự án khi đủ điều kiện hiện hành.

Thực hiện/nghiệm thu có thể song song giữa các task; sơ đồ tổng quan không phải lịch thực thi tuần tự cho toàn dự án. Không thêm Giám đốc phê duyệt từng lần giao việc hoặc nghiệm thu.

## 3. Tổ chức nghiệm thu

Làm rõ ngày 2026-10-06: chỉ chọn người đã là thành viên của dự án để giao trách nhiệm nghiệm thu. Thao tác này không tự thêm người ngoài dự án; thành viên không bắt buộc phải đang nhận task thực hiện. Cách đưa người vào danh sách thành viên được đặc tả riêng, không thêm quy trình ở đây.

Phụ trách chuyên môn mặc định nghiệm thu và lựa chọn thành viên đủ chuyên môn để giao trách nhiệm nghiệm thu ở cấp dự án; không cần PM duyệt. Có thể thêm 1–2 người theo nhu cầu, không có giới hạn cứng. Người được giao có thể nghiệm thu thay mà không cần xác nhận phụ trách chuyên môn đang bận; một người nghiệm thu là đủ. Được nghiệm thu kết quả của phụ trách chuyên môn hoặc PM khi họ trực tiếp thực hiện, nhưng không được tự nghiệm thu. Quyền chỉ áp dụng trong dự án được giao, không tự kèm quyền tạo task, phân công, đóng dự án hoặc giao tiếp quyền nghiệm thu. Ghi nhận người đánh giá và kết luận; không đặt ngưỡng cấp bậc tự động. PM có thể kiêm phụ trách chuyên môn. Cơ chế này thay thế chỉ định người nghiệm thu riêng từng task.

“Đủ chuyên môn” là trách nhiệm lựa chọn của phụ trách chuyên môn, đối chiếu kỹ năng/kinh nghiệm phù hợp; không đặt công thức hoặc yêu cầu cứng phải ngang chức danh.
Nghiệm thu kết quả do PM làm không phải đánh giá quyền quản lý của PM.
Chức danh cao hơn không tự cấp quyền và không bảo đảm đủ chuyên môn.

## 4. Phê duyệt kỹ năng

Chuyển trách nhiệm nhân sự/phê duyệt từ Admin sang Giám đốc, không thay bản chất quy trình.
Nhân viên giải trình kỹ năng, mức đề nghị, kinh nghiệm/đóng góp; tệp, chứng chỉ hoặc liên kết tùy chọn.
Hệ thống tập hợp mô tả mức, mức hiện tại/đề nghị, giải trình, căn cứ công việc, kết quả và nhận xét chuyên môn.
Giám đốc quyết định có lý do; khi chưa đủ chuyên môn cần ý kiến phù hợp, không thêm cấp duyệt bắt buộc.
Không tự duyệt; thiếu căn cứ không đồng nghĩa không có kỹ năng; điểm số/chức danh/số task không tự quyết định mức.
Giữ một PENDING/người/kỹ năng, không quota tuần/tháng, bổ sung khi PENDING, terminal bất biến và yêu cầu mới khi bổ sung sau kết thúc.
Kỹ năng không tự hết hạn; catalog INACTIVE giữ đúng semantic cũ.
Đặc tả chi tiết: [skill approval](../.claude/docs/skill_approval_workflow.md).

Giám đốc đọc căn cứ liên quan theo mục đích đánh giá, không mặc định đọc mọi nhật ký, bình luận hoặc dữ liệu nhạy cảm. Chỉ số tổng hợp cụ thể chưa chốt.
Không thêm đánh giá/phê duyệt CV. Trích xuất kỹ năng từ CV không tự công nhận năng lực.

### 4.1. Cập nhật CV — không phê duyệt tệp

Nhân viên tải lên và xác nhận sử dụng CV hiện tại sau kiểm tra tệp; không cần Giám đốc duyệt CV. Giữ bản cũ để tra cứu theo quyền hiện hành. CV là tài liệu tự khai, không chứng minh toàn bộ kỹ năng đã được công nhận. Trích xuất chỉ hỗ trợ khai báo; Giám đốc duyệt yêu cầu kỹ năng theo mục 4. Đổi CV không tự sửa kỹ năng chính thức, yêu cầu đang xử lý hoặc quyết định trước đó. Lỗi trích xuất không ngăn sử dụng tệp hợp lệ; nhân viên có thể khai báo kỹ năng thủ công.

### 4.2. Chỉnh sửa PM

Giám đốc thay PM bằng thao tác chỉnh sửa người quản lý dự án: chọn PM mới đủ điều kiện và lưu một lần. Lưu thành công thì PM mới thay PM cũ; nếu lưu thất bại, PM cũ vẫn giữ trách nhiệm. Không yêu cầu xóa PM cũ trước, không có quy trình bàn giao riêng hoặc chờ PM cũ xác nhận. Giữ lịch sử thay đổi; không thay đổi trạng thái dự án, phân công công việc hoặc đóng góp đã có. PM mới chỉ có quyền hợp lệ theo trạng thái hiện hành; các trách nhiệm khác của PM cũ được xét độc lập.
Quản trị viên quản lý tài khoản, không thay PM. Không mở thêm quyền ở COMPLETED hoặc mở lại CLOSED/CANCELLED; giữ giới hạn vòng đời hiện hành.

### 4.3. Hai luồng ghi nhận kỹ năng — làm rõ 2026-10-06

- **Từ CV:** nhân viên tải CV; hệ thống trích xuất kỹ năng và hiển thị nội dung nguồn; Giám đốc đối chiếu, xác nhận kỹ năng và mức có căn cứ trước khi cập nhật năng lực chính thức. Không buộc nhân viên gửi yêu cầu thủ công cho từng mục trích xuất. Chưa đủ căn cứ thì làm rõ hoặc nêu lý do chưa công nhận, không từ chối tệp CV.
- **Chủ động cập nhật:** nhân viên gửi đề nghị kỹ năng, mức đề nghị và giải trình phù hợp với loại thay đổi; tệp/liên kết tùy chọn. Không cần CV mới hay kết quả trích xuất. Giám đốc xét duyệt theo cùng tiêu chí.

Hai luồng dùng chung nguyên tắc xét duyệt, không tạo hai cấp phê duyệt. Chỉ kỹ năng chính thức đã duyệt được dùng để gợi ý. CV mới không tự ghi đè hoặc xóa kỹ năng cũ; kỹ năng không thay đổi không cần duyệt lại chỉ vì xuất hiện trong CV mới. Giữ lịch sử và giới hạn một yêu cầu PENDING/người/kỹ năng; quyết định đã kết thúc không mở lại.

Thay hoặc thu hồi người thực hiện chỉ kết thúc quyền theo lượt phân công đó; không làm mất quyền từ trách nhiệm dự án khác còn hợp lệ.

## 5. Sơ đồ và giới hạn

Actor tổng quan: Quản trị viên, Giám đốc, PM, Nhân viên; có thể giữ Người dùng hệ thống.
Không cần actor riêng cho phụ trách chuyên môn/người nghiệm thu. Hành vi có điều kiện được làm rõ ở phần Nhân viên.
Extension point là điểm mở rộng hành vi use case, không phải role hoặc cơ chế cấp quyền. Chưa chốt UC nền/tên điểm/vị trí nối; sửa sơ đồ ở đợt sau.
Không thêm đa cấp duyệt, biểu quyết, giới hạn cứng số reviewer, state, công thức hay CSDL mới.

## 6. Quyết định thay thế và phần giữ nguyên

| Quyết định cũ | Quyết định mới |
|---|---|
| ADMIN duyệt kỹ năng/quản lý hồ sơ nhân sự | Giám đốc đảm nhiệm |
| PM tạo và tự làm chủ dự án | Giám đốc tạo và giao PM |
| ADMIN bàn giao PM bằng luồng riêng | Giám đốc chỉnh sửa PM, chọn người thay và lưu một lần |
| Phê duyệt công bố CV | Nhân viên xác nhận CV hiện tại; chỉ xét duyệt kỹ năng |
| Actor chuyên môn riêng ở tổng quan | Trách nhiệm của nhân viên theo dự án |
| Chỉ định reviewer riêng từng task | Giao trách nhiệm nghiệm thu theo dự án |
| Cân nhắc PM giao reviewer | Phụ trách chuyên môn giao, không thêm PM duyệt |

Giữ R01–R05, phân công nguyên tử, lịch sử đóng góp, frozen spec, skill INACTIVE, vòng đời task/project, giới hạn time correction và thuật toán hiện hành. Nghiên cứu công thức gợi ý không tự trở thành specification triển khai.

## 7. Các điểm chưa chốt — không tự mở rộng

- Chưa thiết kế chi tiết dữ liệu/quyền thay thế, thu hồi trách nhiệm chuyên môn/nghiệm thu hoặc DTO cho reviewer. Quyền đã chốt chỉ trong dự án, không tự tạo quy trình mới.
- Tự duyệt time correction của PM khi là người thực hiện vẫn chưa chốt.
- Chỉ số cụ thể trên màn hình Giám đốc, công thức và điều kiện thử nghiệm chưa được quyết định trong đợt này.

## 8. Nguồn chuẩn

Tài liệu miền ở `.claude/docs/`. Các bản cùng tên tại thư mục gốc là bản đọc đồng bộ.
Các review lịch sử, DB/API, tài liệu UI và PDF chưa đồng bộ không được ghi đè baseline này.
