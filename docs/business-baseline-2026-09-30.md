# Cơ sở nghiệp vụ lịch sử — 2026-09-30

> Bản ghi theo thời điểm. Nguồn hiện hành được dẫn tại [mục lục tài liệu](README.md); không áp dụng lại role, reviewer, catalog INACTIVE hoặc quyền hiệu chỉnh cũ từ nội dung bên dưới. Giữ nguyên nội dung lịch sử để truy vết, không phải đặc tả triển khai hiện tại.

> Lịch sử: các quyết định về vai trò, khởi tạo dự án và nghiệm thu đã được thay thế bởi [cơ sở 2026-10-05](business-baseline-2026-10-05.md). R01–R05 và các quy tắc không bị thay thế vẫn giữ nguyên. Không dùng mô tả ADMIN duyệt kỹ năng hoặc PM tạo dự án dưới đây làm quyền hiện hành.

Đây là thiết kế nghiệp vụ, không phải xác nhận code, CSDL hay báo cáo đã được cập nhật.
Tài liệu này ghi các quyết định mới đã chốt trong trao đổi và được ưu tiên tại những điểm khác baseline V2.
Các quy tắc V2 không bị thay thế vẫn giữ nguyên; không suy việc giản lược sơ đồ thành bỏ kiểm tra nghiệp vụ.

## 1. Trách nhiệm trong dự án

| Trách nhiệm | Phạm vi |
|---|---|
| Quản lý dự án (PM) | Khởi tạo, phân công phụ trách chuyên môn, phân công nhân sự, theo dõi/điều phối, hoàn thành và đóng dự án |
| Phụ trách chuyên môn | Thiết lập cấu trúc, đặc tả công việc và yêu cầu kỹ năng; nghiệm thu kết quả chuyên môn |
| Người thực hiện | Thực hiện công việc được giao, ghi nhật ký, nộp kết quả và sửa theo phản hồi |
| Người nghiệm thu | Đánh giá kết quả theo tiêu chí, chấp nhận hoặc yêu cầu chỉnh sửa |

PM có thể kiêm phụ trách chuyên môn nếu đủ chuyên môn. Phụ trách chuyên môn có thể trực tiếp làm task và sử dụng các chức năng của Nhân viên.
Đây là trách nhiệm theo dự án/công việc, không phải mỗi tên đều là một role hệ thống mới.
Không dùng quy tắc cũ “mỗi người chỉ có một vai trò” để cấm kiêm nhiệm.
Cách biểu diễn quyền trong DB/JWT sẽ thiết kế sau; không tự thêm enum hoặc bảng ở lượt cập nhật này.
ADMIN tiếp tục quản trị tài khoản, hồ sơ và duyệt kỹ năng; không mặc nhiên có quyền nghiệm thu chuyên môn.

## 2. Luồng chính

1. PM khởi tạo dự án và phân công phụ trách chuyên môn; có thể kiêm nhiệm.
2. Phụ trách chuyên môn thiết lập nhóm và công việc. Khi tạo task phải đủ thông tin bắt buộc, gồm yêu cầu kỹ năng, thời gian ước tính, đầu ra và tiêu chí nghiệm thu. Không thêm trạng thái nháp.
3. PM chọn nhiều hoặc tất cả công việc đủ điều kiện, chọn cách phân công thủ công hoặc theo gợi ý.
4. Hệ thống lấy yêu cầu kỹ năng đã lưu, lọc theo kỹ năng bắt buộc và mức độ từ hồ sơ chính thức đã duyệt, rồi xếp hạng cho từng công việc.
5. PM chọn nhân sự và xác nhận. Hệ thống kiểm tra điều kiện và ghi nhận phân công; gợi ý không tự giao việc.
6. Dự án bắt đầu triển khai; người được giao thực hiện, ghi nhật ký và nộp kết quả.
7. Người nghiệm thu đánh giá theo tiêu chí. Chưa đạt: nêu nội dung chỉnh sửa, công việc trở về IN_PROGRESS để sửa và nộp lại. Đạt: ghi nhận nghiệm thu, công việc DONE.
8. PM theo dõi xuyên suốt và điều phối các công việc còn lại; hoàn thành rồi đóng dự án khi đủ điều kiện hiện hành.

Các công việc có thể được thực hiện song song; sơ đồ tổng quan minh họa một luồng đại diện.
Bắt đầu dự án chỉ diễn ra lần đầu, không lặp ở các đợt phân công sau.
Trên activity tổng quan, giữ một bước “Kiểm tra và ghi nhận phân công”; không bắt buộc vẽ nhánh thất bại. Kiểm tra thất bại không được tạo phân công.

## 3. Nghiệm thu và tránh tự nghiệm thu

Mặc định phụ trách chuyên môn nghiệm thu.
Nếu trực tiếp thực hiện task, người này yêu cầu người khác đủ chuyên môn nghiệm thu.
Không tự nghiệm thu; không bổ sung bước PM phê duyệt người nghiệm thu vào luồng đã chốt.
PM chỉ nghiệm thu khi thực sự đảm nhiệm trách nhiệm chuyên môn phù hợp, không phải vì mang vai trò PM.

“Người nghiệm thu” trên activity là trách nhiệm đối với công việc, không phải role toàn hệ thống.
Yêu cầu nghiệm thu không tự trao quyền quản lý toàn bộ dự án cho người được yêu cầu.
Cơ chế xác định người đủ chuyên môn và phạm vi quyền xem cần được đặc tả ở bước thiết kế quyền; chưa chốt tiêu chí tự động hoặc cấp phê duyệt mới.

## 4. Kỹ năng và CV

Kỹ năng không tự hết hạn theo thời gian. Không thêm ngày hết hạn hoặc quy trình gia hạn năng lực.
“Còn hiệu lực” trong tài liệu cũ chỉ có nghĩa bản ghi chính thức chưa bị thu hồi/điều chỉnh qua quy trình đã duyệt, không phải năng lực tự biến mất.
Catalog INACTIVE chỉ ngừng dùng kỹ năng để khai báo requirement mới; không làm mất năng lực đã duyệt.
Giữ quy tắc đã chốt: nhân viên được ADD/UPDATE skill inactive qua xét duyệt; task chưa từng giao phải sửa requirement inactive trước phân công; task đã giao giữ requirement đóng băng và vẫn đối chiếu năng lực chính thức tương ứng.
CV và trích xuất CV vẫn thuộc phạm vi; kết quả trích xuất không tự thành kỹ năng chính thức.
Không thiết kế thêm thuật toán trích xuất ở lần này.

### 4.1. Phê duyệt kỹ năng có căn cứ

Đặc tả chuẩn: [Phê duyệt cập nhật kỹ năng](../.claude/docs/skill_approval_workflow.md), áp dụng chung MANUAL/CV.

- ADMIN là người được tổ chức giao quyền duyệt hồ sơ, không mặc định là CEO hoặc chuyên gia mọi lĩnh vực. Thiếu chuyên môn đánh giá thì cần ý kiến chuyên môn phù hợp trước khi duyệt; không thêm cấp duyệt bắt buộc.
- Bổ sung/nâng mức cần kỹ năng, mức đề nghị và giải trình. Tệp/liên kết không bắt buộc; có thể dùng công việc, thực hành hoặc ý kiến chuyên môn, không bắt buộc có lịch sử dự án. Sửa sai/loại kỹ năng cần nội dung và lý do sửa.
- Đối chiếu đúng đóng góp cá nhân, mức tự chủ, độ phức tạp và chất lượng đầu ra theo bốn mức hiện hành. Không tự suy mức từ CV, chứng chỉ, số năm, số task hoặc điểm gợi ý.
- ADMIN được xem đúng căn cứ task được dẫn chiếu: mô tả, yêu cầu kỹ năng, người/đóng góp liên quan, kết quả, minh chứng và nhận xét nghiệm thu. Không mở toàn dự án, task khác, nhật ký hay toàn bộ trao đổi; không cấp quyền điều hành. Đây là bổ sung có chủ đích cho quyền đọc R06. Nhân viên không được dẫn chiếu để vượt quyền xem; task thay người phải đối chiếu đúng đóng góp lịch sử.
- Hệ thống trình bày tập trung mức hiện tại/đề nghị, mô tả mức, giải trình, căn cứ, ý kiến chuyên môn và lịch sử. Ý kiến ngoài hệ thống được ghi đúng nguồn, không giả định có chữ ký hoặc thao tác của chuyên gia trên hệ thống.
- Khi PENDING được bổ sung căn cứ, không thêm trạng thái hoặc buộc tạo lại chỉ vì bổ sung. Quyết định kết thúc bất biến; bổ sung sau bằng yêu cầu mới theo R03. Kiểm tra phiên bản vẫn giữ.
- Duyệt/từ chối đều có lý do; thiếu căn cứ không đồng nghĩa không có kỹ năng. Không tự duyệt. Pending hoặc từ chối nâng mức không làm mất kỹ năng chính thức hiện có.
- Một PENDING/người/kỹ năng; không hạn ngạch tuần/tháng. Yêu cầu mới sau từ chối cần căn cứ làm rõ/bổ sung, không triển khai phát hiện trùng tự động.

Không thêm thi trực tuyến, AI duyệt, kỹ năng hết hạn, tái đánh giá định kỳ, hội đồng hay vai trò chuyên gia mới. Chưa đổi CSDL/API/UI hoặc báo cáo; rubric là hướng dẫn trong phạm vi thử nghiệm, không tuyên bố đã chuẩn hóa cho mọi ngành.

## 5. Ngoại lệ: giữ điều cần thiết, không mở rộng

| Nội dung | Xử lý hiện tại |
|---|---|
| Tự nghiệm thu | Cấm; dùng người khác đủ chuyên môn như mục 3 |
| Thay người trực tiếp | Nguyên tử; thất bại giữ A và trạng thái; thành công giao B, về ASSIGNED, B phải Start |
| Lịch sử khi thay/thu hồi | Giữ kết quả, nhật ký và đóng góp gắn người/lượt giao cũ; không chuyển sang B |
| Gợi ý khi giao lại | Thu hồi trước; task chờ phân công nếu chưa tìm được người, không tự khôi phục A |
| Kiểm tra phân công thất bại | Giữ kiểm tra, không cần nhánh riêng trên activity tổng quan |
| Kỹ năng tự hết hạn | Không thuộc nghiệp vụ |
| R01 — Nhật ký bổ sung | Khi còn quyền sửa trực tiếp, cho ghi thời gian thực tế thuộc lượt phân công, không phụ thuộc mốc Start/Resume; không duyệt ngoại lệ riêng hoặc sửa ngược trạng thái. Nhật ký đã khóa/lượt giao đã kết thúc vẫn theo hiệu chỉnh chung |
| R02 — Nộp và nghiệm thu | Bỏ nộp kết quả thay; thiếu nhật ký không tự chặn nộp/nghiệm thu đầu ra, không ép khai giờ giả. Dữ liệu thời gian chưa đầy đủ không dùng cho thành phần hiệu suất phụ thuộc thời gian |
| R03 — CV | Không mở lại nhiều vòng mục trích xuất hoặc đóng đợt review riêng; bổ sung bằng yêu cầu kỹ năng mới, có thể dùng CV cũ làm minh chứng. Giữ quyết định cũ, không mở lại kết quả tổng hợp đợt CV |
| R04 — Sửa sai kỹ năng | Bổ sung/điều chỉnh/loại thông tin sai qua xét duyệt chung, giữ lịch sử; có thể đề nghị bổ sung lại. Không có chức năng tái kích hoạt riêng hoặc kỹ năng tự hết hạn |
| R05 — Gợi ý theo thời điểm | Hiển thị thời điểm tạo, PM có thể chạy lại; bỏ xác nhận/lý do riêng do lịch sử tính điểm thay đổi. Giữ snapshot, truy vết và kiểm tra điều kiện phân công hiện tại |
| PerfScore | Giữ định hướng dùng chỉ số lịch sử; công thức, trọng số và điều kiện mẫu cũ là phương án thử nghiệm, chưa chốt cuối |

Giữ nguyên các quy tắc khác: task là đơn vị phân công; tối đa một người phụ trách hiện tại; batch ghi nguyên tử; đặc tả khóa sau lần giao đầu; không tự mở lại công việc đã nghiệm thu; giữ ranh giới hoàn thành/đóng dự án.
ADMIN bàn giao PM ở COMPLETED không mở lại thực thi hay mở thêm quyền theo trạng thái.

## 6. Những điểm chưa được tự quyết định

- Cách biểu diễn kiêm nhiệm bằng role/permission và dữ liệu dự án.
- Cách chọn/ghi nhận người nghiệm thu thay thế, phạm vi truy cập minh chứng.
- Cách xử lý tự duyệt time correction khi PM đồng thời là người thực hiện: quy tắc cũ chưa bao phủ tình huống này. Không mặc nhiên áp dụng quyền PM tự duyệt cho nhật ký của chính mình.
- Cách đánh dấu tối thiểu độ đầy đủ của nhật ký cho PerfScore; không mặc nhiên giữ cơ chế xác nhận chất lượng nhiều cấp cũ.
- Công thức PerfScore và phương pháp thực nghiệm.

Các điểm trên không làm phát sinh chức năng mới ở lần sửa này; cần chốt trước khi triển khai phần liên quan.

## 7. Nguồn tài liệu

Quyết định R01–R05 ở mục 5 thay thế cơ chế tương ứng của V2. Bỏ nộp kết quả thay không bỏ ghi/hiệu chỉnh nhật ký thay; bỏ lý do R05 không bỏ lý do điều chỉnh phân công khác. Không thay các quy tắc ngoài phạm vi đã chốt.

Tài liệu miền chuẩn ở `.claude/docs/`: domain_rules, roles_permissions, project_task_assignment_workflow, ai_recommendation_workflow, work_effort_workflow, product_scope, cv_extraction_approval_workflow và skill_approval_workflow.
Các file workflow cùng tên ở thư mục gốc là bản đồng bộ để đọc thuận tiện, không phải baseline độc lập.
Tài liệu thiết kế DB/API, báo cáo và review cũ chưa phản ánh đầy đủ thay đổi trách nhiệm; không dùng chúng để ghi đè cơ sở nghiệp vụ này.
