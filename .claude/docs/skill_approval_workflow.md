# Phê duyệt cập nhật kỹ năng

**Đồng bộ quyết định 15/10/2026:** [sổ quyết định, phần hiện hành và Bàn giao](../../docs/design/database-review-decisions-2026-10-14.md) ưu tiên cho nội dung đã thay thế; [dictionary/DBML hợp nhất](../../docs/design/database-proposal-2026-10-14.md) là thiết kế 33 bảng (cấp bậc theo chốt 26/10/2026). Đây là thiết kế DB, chưa triển khai DDL/migration hoặc backend nghiệp vụ; tiến độ frontend mô phỏng xem [sổ công việc](../../web/README.md#sổ-công-việc).


Quy trình chốt ngày 2026-09-30, chuyển người duyệt sang Giám đốc ngày 2026-10-05 và phân định HR theo [baseline 2026-10-12](../../docs/business-baseline-2026-10-12.md). Áp dụng cho yêu cầu MANUAL và kỹ năng được nhận diện từ CV; đây là đặc tả nghiệp vụ, chưa phải thiết kế bảng, API hoặc chức năng đã triển khai.

## 1. Mục đích và trách nhiệm

Xác nhận thông tin năng lực có căn cứ để sử dụng trong hồ sơ và phân công. Hệ thống hỗ trợ đối chiếu, không tự chứng nhận năng lực hoặc thay thế đánh giá chuyên môn.

- Nhân viên đề nghị cập nhật và giải trình căn cứ của mình.
- Giám đốc chịu trách nhiệm đánh giá và phê duyệt kỹ năng; chức danh quản lý không mặc nhiên chứng minh chuyên môn trong mọi lĩnh vực.
- HR quản lý hồ sơ hành chính/CV và duyệt cập nhật thông tin cá nhân, không duyệt kỹ năng hoặc sửa kỹ năng chính thức qua quyền cập nhật hồ sơ. Không thêm tầng HR duyệt trước. Phân định theo [baseline 2026-10-12](../../docs/business-baseline-2026-10-12.md).
- Giám đốc tự đánh giá khi đủ chuyên môn. Nếu không đủ khả năng đánh giá lĩnh vực đó, phải có ý kiến chuyên môn phù hợp trước khi phê duyệt.
- Người góp ý chuyên môn có thể là người phụ trách chuyên môn, người hướng dẫn hoặc người đã quan sát công việc. Đây không phải vai trò hệ thống mới hay một cấp phê duyệt bắt buộc cho mọi yêu cầu.
- Không được duyệt yêu cầu kỹ năng của chính mình, kể cả Giám đốc có hồ sơ cá nhân. Không bổ sung ngoại lệ tự duyệt khi tổ chức chỉ có một Giám đốc.

## 2. Nội dung đề nghị và căn cứ

Đề nghị bổ sung hoặc nâng mức cần có kỹ năng, mức đề nghị và giải trình kinh nghiệm/căn cứ liên quan. Nhân viên mô tả việc đã làm, phần mình thực hiện, mức độ tự chủ và kết quả đạt được. Có thể dùng công việc thực tế, việc có hướng dẫn, bài thực hành hoặc tình huống đã được đánh giá; không bắt buộc có lịch sử dự án trong hệ thống.

Tệp, liên kết, chứng chỉ, tài liệu đào tạo, công việc nội bộ và nhận xét chuyên môn là căn cứ bổ sung, không bắt buộc mọi yêu cầu phải có tệp hoặc liên kết. Không có tệp không đồng nghĩa không có kỹ năng; giải trình chưa đủ căn cứ cũng không tự trở thành lý do phê duyệt.

**Chốt bổ sung ngày 2026-10-13:** một yêu cầu được dẫn nhiều công việc trong hệ thống và/hoặc căn cứ ngoài hệ thống. Số công việc nội bộ được dẫn có thể là 0, 1 hoặc nhiều; mỗi căn cứ phải gắn đúng đóng góp của người đề nghị theo mục 5. Không có task nội bộ không miễn yêu cầu giải trình phù hợp và không tự đủ điều kiện phê duyệt. Đã chốt tiếp 15/10: tối đa một CV dẫn trực tiếp qua evidence_cv_id; CV nguồn suy qua extracted_skill_id, không lặp. Cho upload nhiều tệp riêng tùy chọn qua evidence_attachments JSONB metadata, file trên cloud cấu hình chung, không thêm bảng. Loại/số/dung lượng tệp và cleanup còn cần đặc tả.

Đề nghị sửa sai hoặc loại thông tin kỹ năng phải nêu thông tin cần sửa và lý do, không buộc chứng minh năng lực như một đề nghị nâng mức.

Kỹ năng trích xuất từ CV chỉ là dữ liệu đề xuất. Trước khi duyệt vẫn phải xác định kỹ năng, mức và căn cứ theo quy trình này; không tự suy mức từ tên kỹ năng xuất hiện trong CV.

## 3. Tiêu chí đối chiếu

Người duyệt xem xét ba điểm:

1. Căn cứ liên quan đúng người, đúng kỹ năng và xác định được đóng góp cá nhân.
2. Mức đề nghị phù hợp với mức độ tự chủ, độ phức tạp và chất lượng kết quả thể hiện trong căn cứ.
3. Giải trình và nhận xét đủ rõ để kiểm tra, giải thích quyết định; khi chưa đủ thì cần làm rõ, không suy đoán năng lực.

Giữ bốn mức hiện hành, với hướng dẫn đối chiếu trong phạm vi thử nghiệm của đồ án:

| Mức | Biểu hiện để đối chiếu |
|---|---|
| BASIC | Thực hiện công việc cơ bản khi có hướng dẫn; giải thích được phần mình thực hiện |
| INTERMEDIATE | Tự thực hiện công việc thường gặp, đạt yêu cầu đầu ra và biết khi nào cần hỗ trợ |
| ADVANCED | Xử lý công việc phức tạp, giải thích lựa chọn và giải quyết vấn đề phát sinh trong chuyên môn |
| EXPERT | Định hướng hoặc thiết kế giải pháp trong chuyên môn, giải quyết vấn đề khó và hỗ trợ người khác |

Đây là hướng dẫn mô tả, không phải thang đo đã kiểm chứng cho mọi ngành. Ví dụ chuyên môn cần phù hợp với bối cảnh thử nghiệm, không thêm mức kỹ năng hoặc ngưỡng số lượng minh chứng.

Số năm, chứng chỉ, số task hoàn thành, điểm gợi ý hoặc việc CV có tên kỹ năng không tự quyết định mức. Task đã nghiệm thu là căn cứ hỗ trợ, không tự chứng nhận người thực hiện đạt mọi kỹ năng và mức trong yêu cầu task.

## 4. Ý kiến chuyên môn

Ý kiến cần cho biết người nhận xét, công việc hoặc tình huống đã quan sát/kiểm tra, năng lực thể hiện và sự phù hợp với mức đề nghị. Ý kiến phải liên quan đến đóng góp của người đề nghị, không chỉ là lời xác nhận chung.

Trao đổi hoặc đánh giá thực hành có thể diễn ra ngoài hệ thống. Giám đốc ghi nhận nguồn và nội dung ý kiến làm căn cứ; không thể hiện như người chuyên môn đã đăng nhập, ký hoặc duyệt trên hệ thống nếu thực tế không có thao tác đó. Không thêm luồng chuyển duyệt, hội đồng, chữ ký hoặc bài kiểm tra trực tuyến.

## 5. Quyền xem căn cứ công việc

Trong quá trình xét duyệt, Giám đốc được xem mô tả, yêu cầu kỹ năng, người thực hiện/đóng góp liên quan, kết quả nộp, minh chứng và nhận xét nghiệm thu của công việc được dẫn chiếu làm căn cứ.

Quyền này không mở rộng sang toàn dự án hoặc các công việc khác; không mặc định bao gồm nhật ký thời gian hay toàn bộ trao đổi. Không cấp quyền sửa, phân công hoặc nghiệm thu công việc cho Giám đốc.

Nếu công việc đã thay người, căn cứ phải gắn với đóng góp lịch sử của người đề nghị, không dùng kết quả của người nhận sau như kết quả của người trước. Nhân viên chỉ được chọn/dẫn chiếu căn cứ trong phạm vi dữ liệu mình được phép xem; chức năng tìm căn cứ không mở quyền tìm kiếm task khác. Không thêm quy trình xin chia sẻ riêng.

## 6. Quy trình và kết quả

Hai luồng độc lập cùng dùng tiêu chí xét duyệt ở trên:

- **Từ CV:** hệ thống trích xuất; Giám đốc xem nội dung nguồn, xác định kỹ năng và mức có căn cứ trước khi công nhận. Không buộc nhân viên gửi đề nghị thủ công cho từng mục. Chi tiết tại [quy trình CV](cv_extraction_approval_workflow.md).
- **Chủ động cập nhật:** nhân viên đề nghị bổ sung, điều chỉnh hoặc sửa thông tin kỹ năng và giải trình, không cần CV mới. Các bước dưới đây mô tả luồng này; bước đối chiếu và quyết định cũng áp dụng cho kỹ năng từ CV.

1. Nhân viên gửi đề nghị có thông tin và giải trình phù hợp với loại thay đổi.
2. Hệ thống hiển thị tập trung kỹ năng, mức hiện tại/mức đề nghị, mô tả mức, giải trình, căn cứ, ý kiến chuyên môn và lịch sử yêu cầu liên quan.
3. Giám đốc đối chiếu tiêu chí; khi cần thì trao đổi, bổ sung căn cứ hoặc tham khảo chuyên môn. Yêu cầu vẫn PENDING, không thêm trạng thái NEEDS_INFO.
4. Giám đốc phê duyệt hoặc từ chối, nêu lý do ngắn gọn và cụ thể. Kiểm tra quyền, không tự duyệt và phiên bản; ghi quyết định/cập nhật kỹ năng chính thức theo giao dịch hiện hành.
5. Giữ quyết định, người duyệt, thời điểm và căn cứ để tra cứu. Yêu cầu đã kết thúc không được sửa hoặc mở lại; thông tin bổ sung sau đó đi qua yêu cầu mới.

Có thể bổ sung căn cứ khi PENDING mà không phải rút rồi gửi lại chỉ vì thiếu minh chứng. Đối chiếu before_values với giá trị official hiện tại dưới khóa, gồm việc có/không có kỹ năng; đồng thời kiểm tra request.version người duyệt đã xem; bổ sung căn cứ không cho phép ghi đè dữ liệu mới.

Cả phê duyệt và từ chối đều có lý do. Khi căn cứ chưa đủ, nêu điểm chưa rõ để xác nhận mức đề nghị, không kết luận nhân viên không có kỹ năng. Yêu cầu đang chờ hoặc nâng mức bị từ chối không làm mất kỹ năng chính thức hiện có.

## 7. Giới hạn gửi yêu cầu

Tối đa một yêu cầu PENDING cho mỗi người/kỹ năng, dùng chung MANUAL và CV. Có thể gửi yêu cầu cho các kỹ năng khác nhau; không đặt hạn ngạch tuần/tháng hoặc thời gian chờ cố định.

Sau từ chối có thể gửi yêu cầu mới khi căn cứ được làm rõ hoặc bổ sung. Gửi lại cùng nội dung không tạo thêm căn cứ; người duyệt đối chiếu lịch sử. Không bổ sung cơ chế tự động phát hiện nội dung trùng hoặc chấm điểm minh chứng.

## 8. Giới hạn phạm vi

Không thêm kỹ năng tự hết hạn, tái đánh giá định kỳ, nền tảng thi/chứng chỉ, AI phê duyệt, vai trò chuyên gia riêng hoặc phê duyệt nhiều cấp bắt buộc. Chứng chỉ có thể là tệp căn cứ, không phải phân hệ quản lý chứng chỉ.

Không còn skills.status hoặc official is_effective/source/version. Official chỉ lưu năng lực hiện được công nhận; REMOVE gỡ dòng hiện tại, giữ quyết định/căn cứ, công nhận lại qua yêu cầu mới và dòng mới. Nguồn MANUAL/CV giữ ở request. Giám đốc quản trị danh mục kỹ năng; không xóa danh mục còn tham chiếu. Giữ request.version cho nội dung/căn cứ PENDING và các vòng đời công việc/dự án. R03 vẫn áp dụng cho yêu cầu đã kết thúc; yêu cầu mới dùng CV cũ không mở lại mục trích xuất cũ. Các thông tin cần hiển thị ở đây không mặc nhiên làm phát sinh bảng hoặc trường dữ liệu mới.

Upload phải do server xác thực và tạo metadata storage_key/original_file_name/mime_type/file_size_bytes; không tin khóa tệp tùy ý từ client. Tệp riêng của request, bổ sung khi PENDING tăng request.version, terminal giữ nguyên. Không đưa chứng chỉ vào employee_cvs; không JSON hóa ID task/CV thay FK.
