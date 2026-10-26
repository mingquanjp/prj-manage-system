# CV → trích xuất → xét duyệt kỹ năng

**Đồng bộ quyết định 15/10/2026:** [sổ quyết định, phần hiện hành và Bàn giao](../../docs/design/database-review-decisions-2026-10-14.md) ưu tiên cho nội dung đã thay thế; [dictionary/DBML hợp nhất](../../docs/design/database-proposal-2026-10-14.md) là thiết kế 33 bảng (cấp bậc theo chốt 26/10/2026). Đây là thiết kế DB, chưa triển khai DDL/migration hoặc backend nghiệp vụ; tiến độ frontend mô phỏng xem [sổ công việc](../../web/README.md#sổ-công-việc).


> Baseline nghiệp vụ hiện hành: [2026-10-12](../../docs/business-baseline-2026-10-12.md), PTCM phân công và nghiệm thu, không nhận task trong dự án phụ trách; tra cứu giới hạn dữ liệu nghề nghiệp. Phần kỹ thuật V2 chưa đồng bộ không được ghi đè quyết định mới; công thức và quy tắc không bị thay thế giữ nguyên.

Thiết kế đích, không yêu cầu mô hình học máy. Rà soát cách diễn đạt ngày 2026-10-16 theo các quyết định hiện hành dẫn ở đầu file.
Mốc 2026-09-06 và [baseline V2](../../docs/reviews/system-design-baseline-v2.md) là nguồn lịch sử, không thay thế quyết định hiện hành. Xem thêm `database_schema.md`.

## 1. Tiếp nhận và xử lý

Nhân viên upload CV của mình; HR quản lý CV nhân sự và có thể upload thay mặt, ghi người thao tác. Giám đốc đọc CV làm căn cứ xét duyệt kỹ năng, không có quyền quản lý/upload thay chỉ từ vai trò Giám đốc. Không có tầng HR duyệt trước khi Giám đốc xét kỹ năng.
PDF có text/DOCX, tối đa 10 MiB; không OCR/PDF mã hóa. CV Việt/Anh là phạm vi thử nghiệm phải đo riêng.
Một dịch vụ cloud cấu hình chung qua storage.Provider; storage_key không là URL công khai, download phân quyền. Chưa chọn nhà cung cấp, không mở ảnh/OCR.
Kiểm tra kiểu nội dung thật, giới hạn giải nén DOCX, thời gian/bộ nhớ parser; không chỉ kiểm tra đuôi file.

Lưu CV mới sau kiểm tra tệp; chỉ thay CV hiện tại khi nhân viên xác nhận sử dụng. Không tạo yêu cầu phê duyệt CV.
processing_status: QUEUED → PROCESSING → SUCCEEDED/FAILED.
API trích xuất tên kỹ năng/đoạn căn cứ; chưa chọn provider/model/cách chạy cuối. Tạm giữ thiết kế job/lease/retry và các cột hiện có; phải đặc tả adapter/timeout/retry trước triển khai, không tự cài thư viện.
Lưu file bền vững trước enqueue; DB/file không cùng transaction, dọn file tạm/orphan khi thất bại.
Retry không tạo trùng dòng kỹ năng. Lỗi trích xuất không làm mất tệp hợp lệ hoặc tự thay lựa chọn CV hiện tại.

## 2. Nhận diện kỹ năng

Luồng CV: tải tệp → hệ thống trích xuất kỹ năng → Giám đốc đối chiếu, xác nhận kỹ năng và mức có căn cứ → cập nhật năng lực chính thức. Nhân viên không phải gửi một yêu cầu thủ công cho từng kỹ năng trích xuất. Luồng này hỗ trợ hình thành và bổ sung hồ sơ năng lực, không phê duyệt tệp CV.

Hệ thống trình bày kỹ năng nhận diện cùng đoạn nội dung CV liên quan để người duyệt đối chiếu. Nếu CV chưa đủ căn cứ xác định mức, Giám đốc làm rõ với nhân viên hoặc tham khảo ý kiến chuyên môn theo quy trình chung; không suy mức chỉ từ tên kỹ năng. Chưa đủ căn cứ thì nêu rõ điểm thiếu, không công nhận kỹ năng chỉ để hoàn tất xử lý CV và không ảnh hưởng việc sử dụng tệp.

Áp dụng [quy trình phê duyệt kỹ năng chung](skill_approval_workflow.md) cho cả MANUAL/CV: giải trình và căn cứ theo mức, ý kiến chuyên môn khi Giám đốc không đủ khả năng đánh giá, lý do cho cả duyệt/từ chối và không tự duyệt. Giám đốc chỉ xem căn cứ task được dẫn chiếu trong phạm vi đã chốt, không được đọc toàn dự án. Có thể bổ sung căn cứ khi request PENDING; không thêm trạng thái hoặc buộc tạo lại chỉ vì bổ sung minh chứng. Không có hạn ngạch tuần/tháng.

Parse text → normalize → exact/alias danh mục → fuzzy candidate → Giám đốc map hoặc reject.
Không tự suy proficiency/năm kinh nghiệm từ việc có tên skill.
Không lưu suggested_level/suggested_years_experience/confidence_score. Muốn approve phải có skill_id, mức và căn cứ do Giám đốc xác nhận qua request.

Dòng map sang skill dùng cùng employee_skill_change_requests, source=CV; extracted_skill_id liên kết nguồn trích xuất; mỗi mục chỉ được xử lý một lần trong đợt trích xuất đó.
Đây là cách tái sử dụng cơ chế lưu/xét duyệt hiện có, không có nghĩa nhân viên phải chủ động gửi đề nghị thủ công trước khi Giám đốc xem kết quả trích xuất. Hai luồng khác điểm khởi đầu nhưng dùng chung tiêu chí, quyền duyệt và lịch sử quyết định; không thêm bảng hoặc cấp duyệt.
Nhiều dòng map cùng skill phải hợp nhất minh chứng, không ghi đè tuần tự.
Đã có pending manual cùng người/skill → báo conflict; giải quyết/rút trước.
Approval kiểm tra request.version và before_values của official dưới khóa, ghi quyết định + official skill cùng transaction.
WITHDRAWN là dòng bỏ qua, không phải approved/rejected. NO_CHANGE chỉ khi giá trị đề nghị giống bản chính thức còn hiệu lực; kỹ năng từng bị loại có thể được đề nghị bổ sung lại qua xét duyệt thông thường (R04), không tự khôi phục chỉ vì CV còn ghi kỹ năng đó.
Dòng unmapped phải map/reject; nhiều dòng cùng skill hợp nhất thành một mục đánh giá để không nhân số quyết định.
Request stale không được ghi đè; cho reject có lý do hoặc withdraw rồi tạo yêu cầu mới theo R03 bên dưới. Nhân viên inactive: Giám đốc kết thúc pending có lý do, không tự cập nhật chính thức.
Danh mục không có trạng thái INACTIVE; bảo vệ các liên kết khi sửa/xóa, không đổi nghĩa danh mục đang dùng. REMOVE official gỡ dòng hiện tại, không gỡ request/căn cứ hoặc kết quả trích xuất cũ.

### R03 — Một lần xử lý mục CV; bổ sung bằng yêu cầu mới

Sau hợp nhất dòng trùng, mỗi mục kỹ năng có một kết quả xử lý trong đợt trích xuất. Giữ nguồn và minh chứng, không đếm trùng kỹ năng.
Yêu cầu đi qua quy trình xét duyệt kỹ năng chung; tối đa một yêu cầu PENDING/người/kỹ năng. Quyết định đã kết thúc không bị sửa hoặc mở lại.
Nếu cần bổ sung sau khi yêu cầu kết thúc, tạo yêu cầu kỹ năng mới, có thể dùng CV cũ làm minh chứng. Yêu cầu mới không mở lại mục CV hoặc thay đổi kết quả tổng hợp của đợt trích xuất cũ.
Không quản lý nhiều attempts/current decision cho cùng mục và không có thao tác đóng đợt review riêng. Đợt trích xuất được xử lý xong khi không còn mục chờ xử lý.
Chống trùng do retry và kiểm tra phiên bản khi duyệt vẫn giữ; không ghi đè quyết định hoặc kỹ năng chính thức từ dữ liệu cũ.

## 3. CV hiện tại và xét duyệt kỹ năng độc lập

Nhân viên chọn và xác nhận CV hiện tại; hệ thống ghi nhận bản mới và giữ bản cũ. Tối đa một CV hiện tại/người. Không có bước HR, Giám đốc hoặc Quản trị viên phê duyệt, từ chối công bố CV; không cần quy trình xin cập nhật CV. Quyền HR quản lý/upload thay không tự bao gồm xác nhận CV hiện tại thay nhân viên; chi tiết quyền đó chưa chốt, giữ cơ chế nhân viên xác nhận.
CV hiện tại là tài liệu do nhân viên lựa chọn, không phải chứng nhận nội dung hoặc kỹ năng đã được duyệt. PM/PTCM không có quyền xem/tải CV gốc của người khác chỉ từ quyền tra cứu; chỉ xem thông tin nghề nghiệp được phép. Chủ hồ sơ, HR và Giám đốc dùng quyền CV/căn cứ tương ứng, không công khai tệp.

Xác nhận sử dụng CV không phụ thuộc trích xuất thành công: tệp hợp lệ vẫn được dùng khi parser lỗi; nhân viên có thể gửi yêu cầu kỹ năng thủ công. Chỉ kết quả trích xuất thành công mới cung cấp các mục để xét duyệt theo mục 2.
Thay CV không tự tạo quyết định kỹ năng, không rút/từ chối yêu cầu đang xử lý và không sửa kỹ năng đã duyệt. Giữ nguồn CV, minh chứng và lịch sử của từng yêu cầu; yêu cầu kết thúc vẫn bất biến theo R03.
Kỹ năng không xuất hiện trong CV mới không bị xóa khỏi hồ sơ chính thức. Kỹ năng không thay đổi không phải duyệt lại chỉ vì xuất hiện trong CV mới; xử lý NO_CHANGE theo mục 2. Nhân viên có thể chủ động gửi yêu cầu cập nhật kỹ năng độc lập, không cần tải CV mới hoặc chờ kết quả trích xuất.

Các trạng thái publication PENDING/REJECTED theo cơ chế duyệt CV cũ không còn là nghiệp vụ hiện hành. Cách ánh xạ trạng thái lưu trữ sẽ đồng bộ khi sửa CSDL/API; không thêm enum ở đợt này. Trạng thái xử lý trích xuất và kết quả xét duyệt kỹ năng vẫn riêng biệt.

### Kết quả tổng hợp xét duyệt kỹ năng

Review còn mục chưa xử lý/request PENDING → PENDING_REVIEW. Sau khi hết pending, đếm kết quả xử lý của từng mục đã hợp nhất (không tính yêu cầu bổ sung mới ngoài đợt trích xuất):

- A = số approved, R = số rejected, S = số NO_CHANGE/WITHDRAWN.
- A>0, R=0 → APPROVED (báo riêng S, không tính S là được duyệt).
- A>0, R>0 → PARTIALLY_APPROVED.
- A=0, R>0 → REJECTED (báo riêng S).
- A=0, R=0 → NO_CHANGES; reason NO_SKILLS_FOUND nếu không nhận diện được mục nào, NO_CHANGE nếu không cần cập nhật, ALL_WITHDRAWN nếu tất cả đã rút.

Extraction rỗng vẫn SUCCEEDED; ghi nhận NO_CHANGES/NO_SKILLS_FOUND, UI 'Không có kỹ năng để cập nhật'.
NO_CHANGES là tên kỹ thuật cho kết quả không cập nhật đã chốt, không phải chức năng hay cấp duyệt mới.
Quyết định của yêu cầu gắn đợt trích xuất là nguồn kết quả cho mục có request; mục không có request dùng quyết định dòng có audit. Trạng thái tổng hợp được suy ra từ các mục, không cập nhật độc lập trái nhau.
Kết quả tổng hợp ở trên chỉ mô tả xét duyệt kỹ năng, không phê duyệt hoặc từ chối tệp CV; engine chỉ dùng official skills.

## 4. Nghiệm thu

Upload chưa xác nhận không thay CV hiện tại. Xác nhận CV mới không đổi kỹ năng; parse lỗi không ngăn sử dụng tệp hợp lệ. Không tự kết thúc yêu cầu kỹ năng khi đổi CV.
Không approve dòng chưa map/thiếu mức; concurrent approve chỉ một quyết định.
Đo precision/recall/F1 nhận diện skill trên CV có nhãn và được phép sử dụng, không lấy proficiency chưa có căn cứ làm nhãn.
Thư viện parser chưa chọn; kiểm tra bảo trì/bảo mật trước khi thêm dependency.
