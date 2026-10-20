# Luồng dự án → task → phân công → thực hiện — nghiệp vụ hiện hành

**Đồng bộ quyết định 15/10/2026:** [sổ quyết định, phần hiện hành và Bàn giao](../../docs/design/database-review-decisions-2026-10-14.md) ưu tiên cho nội dung đã thay thế; [dictionary/DBML hợp nhất](../../docs/design/database-proposal-2026-10-14.md) là thiết kế 32 bảng. Chưa triển khai code/DDL/migration.


> Baseline nghiệp vụ hiện hành: [2026-10-12](../../docs/business-baseline-2026-10-12.md), PTCM phân công và nghiệm thu, không nhận task trong dự án phụ trách; tra cứu giới hạn dữ liệu nghề nghiệp. Phần kỹ thuật V2 chưa đồng bộ không được ghi đè quyết định mới; công thức và quy tắc không bị thay thế giữ nguyên.

Đây là đặc tả đích cập nhật từ UC001–UC004, không giữ số bước cũ ở các luồng đã thay đổi.
Bảng trạng thái chuẩn ở domain_rules.md; quyết định nguồn ở [baseline hiện hành](../../docs/business-baseline-2026-10-12.md). Baseline V2 chỉ là bối cảnh kỹ thuật cho phần chưa bị thay thế.

## Đối chiếu activity tổng thể

Giám đốc khởi tạo/chỉ định PM → PM giao phụ trách chuyên môn → thiết lập công việc → PTCM phân công → PM bắt đầu triển khai → nhân viên thực hiện/nộp → nghiệm thu → hoàn thành/đóng dự án. HR không có thao tác trong luồng này; không thêm làn HR chỉ vì đã có actor HR.

Nhánh gợi ý: lấy yêu cầu kỹ năng → lọc kỹ năng bắt buộc từ hồ sơ chính thức → chấm điểm/xếp hạng → PTCM chọn. Sơ đồ tổng thể có thể gộp “Kiểm tra và ghi nhận phân công”, kèm note chỉ ghi khi hợp lệ; không bỏ kiểm tra/rollback trong UC003. Bắt đầu triển khai chỉ ở lần đầu. Theo dõi diễn ra xuyên suốt, các task có thể song song; một task đạt không tự hoàn thành dự án.

Nhánh nghiệm thu chưa đạt phải đi qua nhân viên chỉnh sửa rồi nộp lại. Làn Phụ trách chuyên môn biểu thị trách nhiệm theo dự án, không thêm role hệ thống. Bỏ làn Người nghiệm thu; Đánh giá kết quả nằm ở PTCM. Nhánh không đạt hợp nhất trước bước thực hiện/chỉnh sửa, không ngay trước bước nộp.

## UC001 — Khởi tạo và vòng đời dự án

Giám đốc nhập mã/tên, chọn PM bắt buộc, start_date/planned_end_date, priority; mô tả/ghi chú tùy chọn.
Kiểm tra mã unique, ngày đúng thứ tự; tạo PLANNING và giao PM quản lý, lưu history. Người tạo dự án không mặc nhiên là PM.
Hủy form không lưu. Sai định dạng 400, trùng mã/conflict 409.

PM chỉ định PTCM; có thể kiêm nhiệm nếu đủ chuyên môn. PTCM không nhận task trong dự án phụ trách. Khi chỉ định/thay PTCM, phải chuyển giao task đang nhận trước, không tự thu hồi hoặc xóa đóng góp; không cho tự nghiệm thu đầu ra đã thực hiện. Xem quy tắc và giới hạn ở baseline hiện hành.
PTCM chia sẻ kế hoạch chi tiết ngoài hệ thống, ví dụ qua Slack; không có UC gửi/duyệt kế hoạch, không tích hợp Slack hoặc bổ sung quản lý sprint. PM theo dõi xuyên suốt; chia sẻ không phải PM phê duyệt từng phân công.
PM bắt đầu triển khai khi có task ASSIGNED. Hoàn thành project khi có ≥1 DONE và phần còn lại DONE/CANCELLED.
Project COMPLETED vẫn tiếp nhận sửa sai effort, không tiếp tục thực hiện task.
PM đóng project COMPLETED khi không còn yêu cầu sửa effort PENDING: chuyển project CLOSED và mọi DONE thành CLOSED cùng transaction.
Hủy project đang PLANNING/IN_PROGRESS: bắt buộc lý do, hủy task mở, giữ DONE/effort/lịch sử; kết thúc yêu cầu effort PENDING với lý do hủy dự án.
Giám đốc thay PM bằng thao tác chỉnh sửa người quản lý dự án: chọn PM mới đủ điều kiện và lưu một lần. Lưu thành công thì PM mới thay PM cũ; nếu lưu thất bại, PM cũ vẫn giữ trách nhiệm. Không yêu cầu xóa PM cũ trước, không có quy trình bàn giao riêng hoặc chờ PM cũ xác nhận. Giữ lịch sử thay đổi; không thay đổi trạng thái dự án, phân công công việc hoặc đóng góp đã có. PM mới chỉ có quyền hợp lệ theo trạng thái hiện hành; các trách nhiệm khác của PM cũ được xét độc lập. Với COMPLETED, PM mới đóng project khi đủ điều kiện; PTCM hiện hành xử lý correction, không nghiệm thu thêm/tạo task/phân công/thay người/mở lại execution hoặc đưa project về IN_PROGRESS.

## UC002 — Nhóm và task

Phụ trách chuyên môn thực hiện; PM chỉ thực hiện khi kiêm nhiệm trách nhiệm này.
Chỉ project PLANNING/IN_PROGRESS. Tạo group tên unique trong project.
Task có group cùng project, mã unique/project, tên, estimate >0, ≥1 requirement hợp lệ.
Requirements unique skill; PTCM đặt weight theo task trước giao, miền giá trị/chuẩn hóa chờ công thức; level bốn mức chính thức.
Ngày task tùy chọn nhưng nhập đủ cặp, trong ngày project; create task+requirements atomically.
Không parent task/task_type/label, group không được giao người; nhóm là gói công việc/module, được phép trống khi lập kế hoạch.
Đầu ra nằm trong description, acceptance_criteria giữ riêng; đủ rõ trước giao. Ngày kế hoạch ban đầu được giữ; sửa ngày hiện hành có lý do/audit, không chặn effort muộn hợp lệ.

Estimate/requirements/tiêu chí nghiệm thu và scope khóa sau lần phân công đầu; không đổi scope bằng chỉnh mô tả.
Thu hồi không mở khóa spec. Chỉ sửa làm rõ có audit; scope mới tạo task mới.
Không tạo/sửa khi project COMPLETED/CLOSED/CANCELLED.

## UC003 — Xác nhận phân công

Phạm vi đã chốt: [nguồn nhân lực chung 2026-10-13](../../docs/business-baseline-2026-10-13.md). Không lập đội trước; PTCM chọn trong toàn công ty theo điều kiện hiện hành, không cần PM duyệt từng phân công. Hiển thị tải xuyên dự án và số dự án đang có việc theo định nghĩa baseline; số dự án chỉ tham khảo, không giới hạn cứng hai dự án hoặc tự đưa vào điểm.

1. PTCM chọn danh sách PENDING_ASSIGNMENT cùng project, bỏ task không đủ điều kiện.
2. Chọn gợi ý UC003.1 hoặc tra cứu UC003.2; kết quả vào draft trên UI.
3. Hiển thị người chọn, cảnh báo kỹ năng/tải và tổng tải sau cả draft; gợi ý có rank/điểm/giải thích, manual không tự gán matching_score.
4. Confirm gửi dữ liệu nền cần đối chiếu và idempotency key; không gửi điểm như dữ liệu đáng tin.
5. Transaction khóa/kiểm tra lại trách nhiệm PTCM hiện hành, project/task, người được chọn, spec/skill và run. Gợi ý phản ánh thời điểm tạo; thay đổi lịch sử tính điểm không yêu cầu xác nhận riêng R05. PTCM có thể chạy lại, không trộn điểm mới vào snapshot cũ.
6. Tất cả hợp lệ mới ghi assignment, task ASSIGNED/current assignee, status history, notification và kết quả idempotency. Thiết kế đích không ghi membership riêng; danh sách người từng được giao suy từ lịch sử assignment.
7. Commit trả kết quả toàn batch.

Một dòng sai hoặc lỗi lưu → rollback toàn bộ; trả task/index/code lỗi để sửa draft.
Không partial success, không lưu “các dòng hợp lệ” trước. Cùng key/payload retry trả kết quả đã commit, khác payload 409.
Chạy gợi ý không phải phân công; bỏ draft không tạo assignment.

## UC003.2 — Manual và ngoại lệ

Luồng đưa lựa chọn vào draft UC003 áp dụng cho task PENDING_ASSIGNMENT. Thay người trực tiếp chỉ tái sử dụng tra cứu, lựa chọn và kiểm tra manual dưới đây, không kế thừa tiền điều kiện PENDING_ASSIGNMENT hoặc bước confirm của UC003; xem luồng Thay người/thu hồi.
Tra cứu theo kỹ năng/phòng ban/vị trí/kinh nghiệm/trạng thái/tham gia dự án.
Người thực hiện active được chọn nhưng loại PTCM của dự án, kể cả PM kiêm PTCM; hồ sơ chính thức, workload hiện hành, thông tin tham gia giới hạn theo quyền.
Thiếu/dưới mức mandatory được cảnh báo và bắt buộc lý do override.
Không override user locked/employee inactive/sai quyền hoặc điều kiện người được chọn là PTCM của dự án. Tải cao cảnh báo, không kết luận vi phạm lịch.
Chuyển manual sau run phải lưu run_id và lý do; assignment MANUAL luôn item_id=NULL, không lưu score trên assignment, kể cả người đó từng nằm Top-N.
MANUAL sau run dùng dữ liệu tra cứu hiện hành và vẫn kiểm tra điều kiện phân công; không dùng điểm/rank của run cũ.
Tra cứu manual hiển thị hồ sơ chính thức, tải và tổng hợp lịch sử được phép xem, không xếp hạng theo matching_score khi chưa chạy engine.
Gợi ý UC003.1 trả Top-N (mặc định3, có thể0–2), điểm tổng/thành phần và giải thích cho từng ứng viên; bấm chi tiết vẫn cùng giới hạn R06.
Hai nhánh dùng cùng dữ liệu chính thức và cùng kiểm tra phân công, chỉ khác cách lựa chọn; không đồng nhất một thứ tự danh sách tìm kiếm với rank gợi ý.

## Thay người/thu hồi

Chỉ PTCM hiện hành được thay người/thu hồi, khi project PLANNING/IN_PROGRESS và task ASSIGNED/IN_PROGRESS/PAUSED; PENDING_REVIEW phải trả lại trước.
Reassign trực tiếp có lý do, là thao tác nguyên tử: A vẫn phụ trách tới khi thay bằng B thành công. Nếu thất bại, phân công của A và trạng thái task không bị thay đổi bởi thao tác này. Nếu thành công, kết thúc trách nhiệm A, append assignment mới cho B, chuyển ASSIGNED; B phải tự Start trước khi tiếp tục thực hiện. Giữ toàn bộ kết quả/log/đóng góp gắn người và assignment cũ, không chuyển sang B; không mở khóa frozen spec.
Task đã bắt đầu cần ghi chú bàn giao: đã làm/còn lại/tài liệu/vướng mắc; PTCM lập thay được, không bắt người cũ xác nhận.
Đã có thực hiện bao gồm phần việc thực tế được ghi nhận, không chỉ dựa vào timestamp Start; không có yêu cầu duyệt riêng R01.
Unassign có lý do là thao tác riêng: giữ kết quả/log/đóng góp, ghi bộ kết thúc lượt giao và lịch sử trạng thái tương ứng, chuyển PENDING_ASSIGNMENT, xóa current assignee ngay khi thu hồi thành công, giữ spec đóng băng. Không có bảng assignment_events riêng.
Mọi thay đổi + notification cùng transaction; người cũ mất quyền theo lượt phân công đã kết thúc và giữ lịch sử đóng góp. Thay hoặc thu hồi người thực hiện không làm thay đổi các trách nhiệm dự án khác còn hợp lệ của người đó.
Giao lại bằng gợi ý: A đang phụ trách → thu hồi → PENDING_ASSIGNMENT → run mới → chọn người → confirm UC003 → ASSIGNED. Sau thu hồi, A không còn phụ trách; nếu không có người phù hợp hoặc chưa confirm thành công thì task vẫn PENDING_ASSIGNMENT, không tự khôi phục A. Không bổ sung xem gợi ý người thay khi A còn giữ task. Reassign trực tiếp dùng cách tra cứu/kiểm tra manual với cảnh báo/lý do hiện hành, không kế thừa tiền điều kiện PENDING_ASSIGNMENT của UC003/UC003.2.

## UC004 — Thực hiện, nhật ký và nghiệm thu

1. Người hiện được giao (không phải PTCM của dự án) bắt đầu ASSIGNED → IN_PROGRESS khi project IN_PROGRESS; có thể có nhiều task đang làm.
2. Ghi ngày + thời lượng phút + nội dung cho effort thực tế, gồm cả làm ở nhà. Không tự sinh giờ từ trạng thái.
3. PAUSED dùng khi tạm dừng, có lý do; Resume để cập nhật tiến độ khi tiếp tục. Khi còn quyền sửa trực tiếp, được ghi bổ sung thời gian thực tế theo R01 mà không sửa ngược timestamp trạng thái.
4. Nộp từ IN_PROGRESS: tóm tắt kết quả, minh chứng tùy chọn và bộ nhật ký tại lần nộp. Thiếu nhật ký không tự động chặn nộp; không tạo log giả.
5. PTCM đối chiếu tiêu chí đã chốt. Thiếu đầu ra: trả về IN_PROGRESS có lý do để chỉnh sửa và nộp lại; không bổ sung scope mới dưới dạng yêu cầu sửa.
6. Chỉ sai/thiếu nhật ký: có thể yêu cầu bổ sung hoặc hiệu chỉnh; không bắt làm lại đầu ra và không coi đủ nhật ký là điều kiện cứng để nghiệm thu.
7. PTCM hợp lệ, khác người thực hiện, duyệt DONE theo đầu ra, lưu accepted_at. Nhật ký chưa đầy đủ không dùng cho thành phần hiệu suất phụ thuộc thời gian. Không có chức năng nộp kết quả thay.
8. DONE không reopen. Sửa sai effort theo quy trình riêng khi project IN_PROGRESS/COMPLETED; CLOSED/CANCELLED khóa.

[Đặc tả effort](work_effort_workflow.md) là nguồn chuẩn về quyền nhập/sửa, kiểm tra ngày/tổng, yêu cầu hiệu chỉnh và khóa.
PTCM xem nhật ký cùng nghiệm thu, không triển khai phê duyệt timesheet/chấm công/OT độc lập.

## Trách nhiệm nghiệm thu và phạm vi giản lược

Phụ trách chuyên môn (PTCM) của dự án phân công và nghiệm thu công việc, không trực tiếp nhận/thực hiện task trong chính dự án đó. Không giao thêm trách nhiệm nghiệm thu, không có reviewer thay thế do PM chỉ định. PM chỉ có quyền chuyên môn khi kiêm PTCM và cũng phải tuân thủ giới hạn không nhận task. PTCM có thể thực hiện task ở dự án khác nếu đủ điều kiện. Lưu người đánh giá, kết luận và thời điểm; link PR/bình luận/merge không tự thay thế quyết định nghiệm thu trong hệ thống.
Kỹ năng không tự hết hạn; official chỉ chứa năng lực hiện được công nhận, danh mục không có trạng thái INACTIVE.
R01–R05 đã được giản lược theo cơ sở hiện hành; vẫn giữ lịch sử, giới hạn quyền và kiểm tra dữ liệu. Hiệu chỉnh do người thực hiện tự gửi, PTCM hiện hành duyệt, không có quyền PM tự duyệt hoặc ghi thay.

Kết thúc lượt giao theo dictionary 5A: ghi ended_at/ended_by/end_reason cùng transaction khi giao lại, thu hồi, DONE hoặc hủy; không sửa nội dung lượt cũ. Hủy khi đang review giữ submission chưa kết luận, loại khỏi hàng chờ qua trạng thái task/project, không giả REJECTED và không tự đổi Q/T.
