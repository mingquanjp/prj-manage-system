# Phạm vi sản phẩm — nghiệp vụ hiện hành

**Đồng bộ quyết định 15/10/2026:** [sổ quyết định, phần hiện hành và Bàn giao](../../docs/design/database-review-decisions-2026-10-14.md) ưu tiên cho nội dung đã thay thế; [dictionary/DBML hợp nhất](../../docs/design/database-proposal-2026-10-14.md) là thiết kế 32 bảng. Chưa triển khai code/DDL/migration.


**Bổ sung 2026-10-13:** [nguồn nhân lực chung](../../docs/business-baseline-2026-10-13.md). PTCM tra cứu/phân công toàn công ty; không lập đội trước, không UC thêm thành viên/duyệt điều động/tìm ngoài đội. Hiển thị tải xuyên dự án và số dự án đang có việc để tham khảo; không giới hạn hai dự án hoặc sửa công thức. Lịch sử người được giao suy từ assignments, không quản lý roster riêng.

> Baseline nghiệp vụ hiện hành: [2026-10-12](../../docs/business-baseline-2026-10-12.md), PTCM phân công và nghiệm thu, không nhận task trong dự án phụ trách; tra cứu giới hạn dữ liệu nghề nghiệp. Phần kỹ thuật V2 chưa đồng bộ không được ghi đè quyết định mới; công thức và quy tắc không bị thay thế giữ nguyên.

Rà soát cách diễn đạt và tổng hợp UC ngày 2026-10-16 theo các quyết định hiện hành dẫn ở đầu file. Mốc 2026-09-06 là lịch sử hình thành tài liệu, không phải mốc nghiệp vụ hiện hành. Đây là thiết kế đích, không phải danh sách chức năng đã implement.
Quyết định hiện hành: [baseline 2026-10-12](../../docs/business-baseline-2026-10-12.md). V1/V2 là lịch sử, chỉ giữ phần không bị quyết định mới thay thế.
Tài liệu gốc trong `docs/thesis-context/` chỉ đọc; các điểm khác phải được cập nhật vào bản báo cáo làm việc sau.

## 1. Mục tiêu

Một tổ chức nhỏ–vừa, giao diện tiếng Việt: hồ sơ năng lực đã duyệt → yêu cầu task → gợi ý/tra cứu → PTCM xác nhận → thực hiện và nghiệm thu. Đây là mô hình đề xuất cho đồ án, không khẳng định là cơ cấu chung của mọi doanh nghiệp.
Gợi ý dùng luật và công thức chấm điểm giải thích được; không bắt buộc mô hình học máy.
Không tuyên bố công thức tối ưu khi chưa có thực nghiệm. Thiết kế dùng `RECOMMENDED` cho nhánh gợi ý và `MANUAL` cho nhánh thủ công. Mã `AI` trong implementation cũ chưa chuyển đổi; tên file cũ giữ nguyên, không chứng minh sử dụng ML.

## 2. Tác nhân và chức năng

| Actor | Chức năng và ranh giới |
|---|---|
| Quản trị viên | Quản lý tài khoản; tạo tài khoản kèm hồ sơ nhân viên ban đầu trong một lần theo quyết định 20/10; HR quản lý hồ sơ về sau; không mặc định duyệt kỹ năng |
| HR | Xem/cập nhật/vô hiệu hóa hồ sơ nhân sự, quản lý CV nhân sự, xét duyệt cập nhật thông tin cá nhân; quản lý phòng ban và vị trí công việc; không duyệt kỹ năng hoặc cấp quyền hệ thống |
| Giám đốc | Xem hồ sơ/căn cứ và đánh giá/phê duyệt kỹ năng; quản lý danh mục kỹ năng (nhóm kỹ năng và kỹ năng); khởi tạo dự án, chỉ định/thay PM; không thay HR duyệt thông tin cá nhân |
| PM | Quản lý dự án, chỉ định PTCM, tra cứu nghề nghiệp, theo dõi và bắt đầu/hoàn thành/đóng dự án; không mặc định phân công hoặc nghiệm thu |
| EMPLOYEE | Hồ sơ của mình, gửi yêu cầu cập nhật kỹ năng/hồ sơ, cập nhật CV không cần duyệt tệp, thực hiện task hiện được giao; PTCM không nhận task trong dự án phụ trách |

Một người có thể kiêm PM và phụ trách chuyên môn; trách nhiệm dự án không đồng nhất với một role hệ thống duy nhất. Project → TaskGroup → Task, không cây task nhiều cấp.
PTCM thiết lập công việc, phân công/thay/thu hồi và nghiệm thu trong dự án được giao. Tra cứu hồ sơ nhân sự là danh sách/tìm kiếm riêng cho PM và PTCM, chỉ thông tin nghề nghiệp/tổng hợp được phép; không mở CV gốc người khác. Tra cứu nhân sự phù hợp trong phân công dùng chung nguồn nhưng gắn task cụ thể. Xem phạm vi dữ liệu tại roles_permissions.md.
Gồm yêu cầu cập nhật hồ sơ và kỹ năng riêng, vòng đời dự án/task, pause/resume, yêu cầu chỉnh sửa kết quả khi nghiệm thu chưa đạt, thay người/thu hồi, thông báo nội bộ, ghi effort thủ công theo task và trích xuất kỹ năng từ CV và xét duyệt kỹ năng. Không có thao tác riêng để nhân viên tự trả lại phân công.
TaskGroup là gói công việc/module, task có đầu ra nghiệm thu và một người chịu trách nhiệm; không triển khai label trong V2.
Manual tra cứu/lọc không tự sinh matching_score; engine trả Top-N mặc định3 kèm điểm/giải thích, có thể ít hơn/rỗng. Cả hai đều cần PTCM confirm.
R01–R05 đã giản lược; R01 được chốt tiếp nối 21/10: tạo/sửa/xóa nhật ký trực tiếp chỉ khi task/project IN_PROGRESS, cần Bắt đầu/Tiếp tục trước, không thêm duyệt ngoại lệ hoặc timer. Nghiệm thu đầu ra tách khỏi độ đầy đủ nhật ký; bỏ nộp kết quả thay; bổ sung sau khi yêu cầu kỹ năng kết thúc bằng yêu cầu mới thay vì mở lại mục CV; không có chức năng tái kích hoạt riêng; gợi ý theo thời điểm tạo, không xác nhận lịch sử cũ. R06 giữ giới hạn dữ liệu theo mục đích, có bổ sung quyền đọc căn cứ xét duyệt kỹ năng dưới đây.
Công thức trọng số skill/workload/history cũ được giữ làm phương án thử nghiệm đối chứng; công thức triển khai cuối sẽ được chốt riêng. Các đề xuất nghiên cứu không tự thay thế nghiệp vụ hiện hành; không tự đổi phương pháp hoặc tham số trước khi được xác nhận.

Hai UC hỗ trợ ở tổng quan: **Quản lý phòng ban và vị trí công việc** (HR; departments/positions) và **Quản lý danh mục kỹ năng** (Giám đốc; skill_categories/skills). Không bắt buộc sơ đồ phân rã riêng; phạm vi và điều kiện thao tác được mô tả trong đặc tả. Danh mục dùng chung không đồng nhất với cập nhật hồ sơ cá nhân hoặc công nhận kỹ năng của một nhân viên; quyền quản trị và bảo vệ liên kết/lịch sử giữ nguyên.

Phê duyệt kỹ năng gồm giải trình của nhân viên, đối chiếu căn cứ và quyết định có lý do của Giám đốc; tệp/liên kết không bắt buộc. Hệ thống gom mô tả mức, kết quả công việc được dẫn chiếu, ý kiến chuyên môn và lịch sử liên quan để hỗ trợ xem xét. Giám đốc thiếu chuyên môn cần ý kiến phù hợp trước khi duyệt; không tự duyệt. R06 được bổ sung quyền đọc đúng căn cứ task được dẫn chiếu, không mở toàn dự án hoặc quyền điều hành. Cho bổ sung căn cứ khi PENDING; quyết định kết thúc bất biến, không có hạn ngạch tuần/tháng. Xem [phạm vi phê duyệt kỹ năng](skill_approval_workflow.md).

CV và yêu cầu cập nhật kỹ năng có hai điểm khởi đầu độc lập: kỹ năng được hệ thống trích xuất để Giám đốc đối chiếu, hoặc nhân viên chủ động gửi đề nghị và giải trình. Không bắt gửi đề nghị thủ công cho từng mục trích xuất; cả hai chỉ cập nhật năng lực chính thức sau xét duyệt. CV mới không tự thay thế kỹ năng đã công nhận. Xem [quy trình CV](cv_extraction_approval_workflow.md) và [quy trình kỹ năng](skill_approval_workflow.md).

HR xử lý yêu cầu cập nhật thông tin cá nhân; pending/từ chối không đổi hồ sơ chính thức, từ chối có lý do. HR cập nhật trực tiếp là thao tác riêng, không bao gồm kỹ năng chính thức. Nhân viên đề nghị sửa fullname/phone/address; HR sửa trực tiếp email công ty/cơ cấu. Địa chỉ cần bổ sung khi hoàn thiện hồ sơ nhưng không bắt Admin nhập khi tạo, không được xóa trắng khi đã có, không lộ trong tra cứu chuyên môn. Thiếu địa chỉ chỉ nhắc bổ sung, không chặn đăng nhập/nhận việc/thực hiện task hoặc thêm trạng thái hồ sơ mới. Không thêm tầng HR tiền kiểm kỹ năng, không phê duyệt CV hoặc mở rộng sang HRM đầy đủ. Xem [quy trình hồ sơ](employee_profile_workflow.md).

## 3. Ngoài phạm vi

Chứng chỉ riêng, đa tổ chức, task dependency, tối ưu lịch/nguồn lực theo ngày, nghỉ phép, chấm công/lương, phê duyệt OT, label, tích hợp GitHub, OCR, tự động giao việc, email/push và mở lại task đã nghiệm thu hoặc project terminal.
Không xây nền tảng thi/đánh giá năng lực tự động, kỹ năng tự hết hạn, tái đánh giá định kỳ hoặc duyệt nhiều cấp bắt buộc. Chứng chỉ được dùng làm căn cứ tùy chọn, không phải phân hệ quản lý chứng chỉ.
Kế hoạch chi tiết chia sẻ ngoài hệ thống (ví dụ Word qua Slack), không có gửi/duyệt kế hoạch, tích hợp Slack hay phân hệ sprint mới. Không có cấp trách nhiệm reviewer riêng hoặc PM chỉ định người nghiệm thu thay thế.
Tải công việc là tải tồn đọng, không phải khả năng đáp ứng deadline.
Effort gồm cả làm ở nhà/ngoài giờ; 8 giờ là đơn vị ngày công quy đổi và ngưỡng cảnh báo, không phải trần.
Không timer, xác minh vị trí hoặc suy ra lương/OT. Chỉ số lịch sử thực hiện phục vụ thử nghiệm, không đánh giá năng suất khách quan.

## 4. Lộ trình và hiện trạng

| Mốc | Nội dung thiết kế đích |
|---|---|
| M0 | Skeleton và migration 000001 hiện có; chưa có service nghiệp vụ |
| M1 | Session phía server/cookie secret, thu hồi phiên, CSRF, RBAC/ABAC, tài khoản kèm nhân sự, đổi/reset mật khẩu |
| M2 | Danh mục; HR quản lý hồ sơ và duyệt thông tin cá nhân; Giám đốc duyệt kỹ năng; request.version kỹ năng/đối chiếu before_values/audit, allowlist fullname/phone/address |
| M3 | UC001/UC002, nhóm–task, requirements; start/complete/close/cancel project |
| M4 | UC003/UC003.2, batch atomic, reassign/unassign và retry an toàn |
| M5 | UC003.1, công thức Top-3, snapshot và truy vết quyết định |
| M6 | UC004, nhật ký ngày/thời lượng/nội dung, sửa sai có lịch sử, nghiệm thu và khóa |
| M7 | PDF text/DOCX → API nhận diện tên kỹ năng/đoạn căn cứ → trích xuất và xét duyệt kỹ năng; nhân viên xác nhận CV hiện tại, không phê duyệt CV |
| M8 | Kiểm thử liên miền, đánh giá công thức, đồng bộ báo cáo và đóng gói |

M1–M6 là lõi; bản đồ án đầy đủ gồm M7–M8. Không coi M7 là “phần học máy bắt buộc”.
Cloud một provider cấu hình chung và API trích xuất đã chọn về hướng; provider/model chưa chốt, không cài thư viện. Upload minh chứng riêng trên request bằng metadata JSON, không thêm phân hệ tài liệu.

## Trách nhiệm theo dự án — chốt 2026-10-12

Phụ trách chuyên môn (PTCM) của dự án phân công và nghiệm thu công việc, không trực tiếp nhận/thực hiện task trong chính dự án đó. Không giao thêm trách nhiệm nghiệm thu, không có reviewer thay thế do PM chỉ định. PM chỉ có quyền chuyên môn khi kiêm PTCM và cũng phải tuân thủ giới hạn không nhận task. PTCM có thể thực hiện task ở dự án khác nếu đủ điều kiện. Lưu người đánh giá, kết luận và thời điểm; link PR/bình luận/merge không tự thay thế quyết định nghiệm thu trong hệ thống.
Nhân viên là nền tảng chung; PTCM là trách nhiệm theo dự án, không phải role toàn hệ thống. UC Quản lý công việc được giao chỉ giữ extension point Trách nhiệm chuyên môn; bỏ Trách nhiệm nghiệm thu và UC giao người nghiệm thu. Extension point mô tả điểm mở rộng hành vi, không thay mô hình phân quyền. Actor HR được bổ sung; sơ đồ phân rã xác định quyền khác nhau khi HR/Giám đốc cùng nối UC Quản lý hồ sơ nhân sự ở tổng quan.

HR quản trị departments/positions; Giám đốc quản trị skill_categories/skills. Hiệu chỉnh nhật ký sau khóa do PTCM hiện hành duyệt, không PM ghi/gửi thay; xóa mềm không khôi phục trong phiên bản này.
