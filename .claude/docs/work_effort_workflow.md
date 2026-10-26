# Nhật ký effort và hiệu chỉnh — nghiệp vụ hiện hành

**Đồng bộ quyết định 15/10/2026:** [sổ quyết định, phần hiện hành và Bàn giao](../../docs/design/database-review-decisions-2026-10-14.md) ưu tiên cho nội dung đã thay thế; [dictionary/DBML hợp nhất](../../docs/design/database-proposal-2026-10-14.md) là thiết kế 33 bảng (cấp bậc theo chốt 26/10/2026). Đây là thiết kế DB, chưa triển khai DDL/migration hoặc backend nghiệp vụ; tiến độ frontend mô phỏng xem [sổ công việc](../../web/README.md#sổ-công-việc).


> Baseline nghiệp vụ hiện hành: [2026-10-12](../../docs/business-baseline-2026-10-12.md), PTCM phân công và nghiệm thu, không nhận task trong dự án phụ trách; tra cứu giới hạn dữ liệu nghề nghiệp. Phần kỹ thuật V2 chưa đồng bộ không được ghi đè quyết định mới; công thức và quy tắc không bị thay thế giữ nguyên.

Rà soát cách diễn đạt ngày 2026-10-16 theo các quyết định hiện hành dẫn ở đầu file. Mốc 2026-09-06 và [baseline V2](../../docs/reviews/system-design-baseline-v2.md) là nguồn lịch sử, không thay thế quyết định hiện hành.
Đây là thiết kế đích, chưa triển khai. Thay thế toàn bộ mô hình phiên/timer/lịch làm việc của V1.

## 1. Ý nghĩa và phạm vi

Effort là thời lượng người thực hiện tự khai thực sự dành cho task, gồm tại công ty, ở nhà, ngoài giờ và cuối tuần.
Ví dụ 6h ở công ty + 4h ở nhà cho cùng task = 10h effort, không phải 6h.
Không lưu địa điểm, không suy ra nơi làm từ tổng giờ. 10h cũng có thể trải trên nhiều ngày; mỗi ngày ghi đúng phần của ngày đó.
Định nghĩa phải hiện trên form/hướng dẫn, không chỉ ngầm hiểu giữa người phát triển.

Bao gồm phân tích, triển khai, kiểm thử, nghiên cứu và trao đổi trực tiếp phục vụ task, sửa lỗi theo tiêu chí nghiệm thu.
Không gồm nghỉ, việc cá nhân, chờ thụ động hoặc hoạt động chung không thuộc task; không ép nhân viên phân bổ đủ 8h/ngày.
Cùng một khoảng làm việc không ghi vào hai task; cuộc trao đổi phục vụ nhiều task phải chia thời lượng, không nhân đôi.
Task một người chịu trách nhiệm không đồng nghĩa không ai hỗ trợ. V2 không thêm quản lý nhóm cộng tác/chấm giờ người hỗ trợ:
nếu hỗ trợ đáng kể mà dữ liệu không phản ánh đầy đủ, đánh dấu không đủ tin cậy cho so sánh cá nhân; không gán giờ của người khác cho người chịu trách nhiệm.

Không chấm công, tính lương, xác nhận OT hoặc xác minh lao động thực tế.
8h = 1 ngày công quy đổi effort, không phải một ngày lịch, ngày có mặt hay quyền lợi tiền lương.

## 2. Dữ liệu và kiểm tra nhập

Mỗi log lưu assignment_id, work_date, duration_minutes nguyên dương, description và metadata xóa mềm/thời điểm. Task/người thực hiện suy từ assignment. Chỉ người thực hiện tự ghi, không ghi/gửi thay. History chỉ UPDATE/DELETE, giữ before/after/actor/lý do; tạo mới không sinh history Edit.
Cho phép nhiều log cùng ngày/task; UI nhập giờ/phút nhưng chuẩn hóa thành phút nguyên, không làm tròn tăng giờ.

- Ngày theo Asia/Ho_Chi_Minh, không tương lai; làm qua nửa đêm tách hai ngày.
- Ngày ghi phải thuộc khoảng người đó thực sự được phân công/thực hiện task; không ghi thời gian trước khi dự án bắt đầu hoặc sau khi công việc đã kết thúc.
- Đối chiếu ngày làm thực tế với lượt phân công và thời gian dự án được triển khai; không dùng planned_end_date hoặc thời điểm bấm Start/Resume làm hạn nhập. Quyền sửa trực tiếp và khóa nhật ký vẫn áp dụng.
- Ngày quá hạn kế hoạch vẫn hợp lệ nếu công việc thực tế diễn ra trong giai đoạn thực hiện; không ép PM gia hạn để nhận nhật ký.
- Duration-only không xác minh giờ bắt đầu/kết thúc hoặc điểm bàn giao trong ngày. Những ngày trùng mốc chuyển trạng thái cần mô tả rõ phần đã làm; không hứa phát hiện mọi chồng lấn.
- Tổng log hiệu lực của một employee trong một ngày trên mọi task/project >480 phút: cảnh báo và bắt buộc lý do vượt 8h.
- Tổng >1440 phút: chặn. Đây là giới hạn logic, không khuyến nghị làm 24h hoặc xác nhận toàn bộ thời lượng đã khai.
- Không thêm trần bằng số phút đã trôi qua hôm nay trong V2: chưa có dữ liệu khoảng giờ để kiểm chứng, vẫn áp dụng không ghi trước công việc thực tế.
- Kiểm tra tổng tại lúc ghi/duyệt, không chỉ lúc mở form. Sửa ngày/thời lượng tính lại cả ngày cũ và mới; tăng/đổi khiến ngày vượt 8h phải có lý do.
- Đối chiếu giá trị nền dưới khóa, chặn ghi đè từ màn hình cũ; retry cùng idempotency key không tạo thêm log. Không tự loại log giống nội dung/ngày vì có thể là hai lần làm thật.

Không đòi người dùng xem nội dung các project khác để hiểu cảnh báo: trả tổng cá nhân/ngày và phần mình có quyền xem.

### R01 — Bắt đầu/Tiếp tục trước khi ghi nhật ký

Theo [chốt tiếp nối 21/10](../../docs/business-baseline-2026-10-20.md#quyết-định-tiếp-nối--điều-kiện-ghi-nhật-ký-21102026), chỉ tạo/sửa/xóa mềm trực tiếp khi người hiện được giao hợp lệ đang thực hiện task IN_PROGRESS trong project IN_PROGRESS. Quyết định này thay R01 cũ cho ghi khi ASSIGNED/PAUSED vì quên Start/Resume.
ASSIGNED cần Bắt đầu, PAUSED cần Tiếp tục; vẫn đọc nhật ký đã có. Không thêm bước xác nhận nhận việc, yêu cầu duyệt ngoại lệ hoặc tự chuyển trạng thái khi lưu log.
Start/Resume không phải timer/chấm công, không sinh giờ hoặc sửa ngược lịch sử. Ngày ghi vẫn theo quy tắc mục 2, không suy thời lượng hay thêm ràng buộc timestamp từ thao tác trạng thái.
Nhật ký đã khóa và đóng góp thuộc lượt phân công đã kết thúc vẫn theo quy trình hiệu chỉnh. Giữ kiểm tra ngày, thời lượng, lượt phân công và khóa dự án; không ghi công việc trước phân công, trong khoảng đã thu hồi hoặc phát sinh sau khi task kết thúc.

## 3. Quyền theo trạng thái

| Bối cảnh | Thao tác |
|---|---|
| Project PLANNING, task chưa bắt đầu | Không log effort thực hiện |
| IN_PROGRESS, người hiện được giao, task IN_PROGRESS | Tạo, sửa, xóa mềm log của mình; sửa/xóa có lý do và history |
| Task PAUSED, project IN_PROGRESS, người hiện được giao | Chỉ đọc nhật ký; cần Tiếp tục về IN_PROGRESS trước khi tạo/sửa/xóa trực tiếp |
| Task ASSIGNED, project IN_PROGRESS, người hiện được giao | Chỉ đọc nhật ký; cần Bắt đầu về IN_PROGRESS trước khi tạo/sửa/xóa trực tiếp |
| PENDING_REVIEW | Nhật ký nộp đã khóa; sửa qua yêu cầu, không sửa trực tiếp |
| DONE hoặc task CANCELLED, project IN_PROGRESS/COMPLETED | Chỉ hiệu chỉnh dữ liệu cũ qua yêu cầu; không ghi công việc phát sinh sau kết thúc |
| Người đã được thay/thu hồi | Theo lượt phân công cũ, chỉ xem đóng góp của mình và gửi yêu cầu hiệu chỉnh; quyền từ trách nhiệm dự án khác còn hợp lệ được xét riêng |
| Người khóa/nghỉ/đã đổi role không còn quyền tự phục vụ | Không mở ghi/gửi thay; yêu cầu đã gửi do PTCM hiện hành xử lý theo quyền |
| Project CLOSED/CANCELLED | Không tạo/sửa/duyệt effort; chỉ đọc theo quyền |

Nhật ký đã khóa hoặc thuộc lượt giao cũ dùng luồng yêu cầu; ASSIGNED/PAUSED cần Bắt đầu/Tiếp tục, không mở yêu cầu ngoại lệ chỉ để bỏ qua điều kiện này.
**Chốt 15/10:** khi còn quyền Edit, nhân viên sửa/xóa mềm có lý do và history, không cần duyệt. Khi khóa, gửi yêu cầu do PTCM hiện hành xử lý; không PM duyệt từ trách nhiệm PM, không nhập/gửi thay. Không thêm ngoại lệ PM tự duyệt. Quản trị viên không sửa effort; PTCM không nhận task cùng dự án. Đổi PTCM không làm mất lịch sử hoặc đổi actor cũ.

Log đã deleted_at không được sửa/xóa lại hoặc khôi phục trong phạm vi hiện tại. Xóa mềm không dùng duration=0, không xóa vật lý.

## 4. Nộp và nghiệm thu

Người thực hiện nộp tóm tắt kết quả và link PR/commit/tài liệu/test tùy chọn; thông tin nhật ký phải phản ánh đúng thực tế. Lưu bộ nhật ký tại lần nộp để truy vết, kể cả khi chưa có log.
R02: thiếu nhật ký không tự động chặn nộp hoặc nghiệm thu đầu ra. Không yêu cầu tạo log giả hay đoán giờ để vượt kiểm tra.
Link là minh chứng tham khảo do PTCM kiểm tra, không tích hợp GitHub hay tự suy merge là nghiệm thu.

PTCM hiện hành, không phải người đã thực hiện đầu ra đang xét, xem đầu ra theo acceptance criteria đóng băng và xem nhật ký cùng lần nghiệm thu:

- Thiếu/sai đầu ra trong scope: trả về IN_PROGRESS có lý do; người thực hiện chỉnh sửa, ghi effort thực tế và nộp lại.
- Scope mới: task mới, không sửa estimate/criteria cũ hoặc ép làm miễn phí dưới nhãn rework.
- Chỉ sai/thiếu nhật ký: có thể yêu cầu bổ sung hoặc hiệu chỉnh theo quyền; không bắt làm lại đầu ra và không coi việc hoàn tất hiệu chỉnh là điều kiện cứng để nghiệm thu.
- Nghiệm thu DONE lưu thời điểm chấp nhận; không tự tính giờ từ trạng thái và không đồng nghĩa nhật ký đã đầy đủ.

Không có chức năng nộp kết quả thay hoặc ghi/gửi nhật ký thay trong phạm vi hiện tại.
Bộ log rỗng hiển thị 'Chưa có thời gian ghi nhận', không khẳng định thực tế không tốn giờ. Có tổng thời gian lớn hơn 0 cũng không chứng minh đã ghi đủ.
Dữ liệu thời gian chưa đầy đủ hoặc không đáng tin cậy không dùng cho thành phần đánh giá hiệu suất phụ thuộc thời gian. Cách đánh dấu tối thiểu được chốt khi thiết kế PerfScore; không giữ quy trình xác nhận chất lượng nhiều cấp của R02 cũ.
Yêu cầu hiệu chỉnh đang chờ tiếp tục theo luồng riêng sau nghiệm thu; không tự duyệt hoặc xóa khi task DONE. Điều kiện đóng dự án vẫn giữ theo mục 6.

## 5. Yêu cầu hiệu chỉnh

ADD/UPDATE/DELETE theo assignment; không task_id/employee_id/base_version/evidence lặp. Căn cứ tùy chọn gộp reason. UPDATE/DELETE có target_log_id và before_values; ADD giữ target/before NULL. proposed_values có với ADD/UPDATE, NULL với DELETE. result_log_id chỉ có khi ADD đã APPROVED (người dùng đồng ý 15/10).

1. Người có đóng góp tự gửi; không PM/PTCM gửi thay. PENDING không thay log/tổng chính thức.
2. PTCM hiện hành kiểm tra project IN_PROGRESS/COMPLETED, ngày thuộc lượt thực hiện, tổng phút xuyên dự án, quyền và dữ liệu nền trong giao dịch có khóa. Đối chiếu before_values gồm ngày/phút/nội dung/lý do vượt ngưỡng/deleted_at; nếu khác thì không ghi đè.
3. ADD được duyệt: tạo log và lưu result_log_id/APPROVED/người/thời điểm/thông báo nguyên tử. Không tạo history ADD. Khóa request còn PENDING ngăn áp dụng hai lần.
4. UPDATE/DELETE được duyệt: cập nhật log, history before/after/request_id, quyết định và thông báo cùng commit. History.request_id unique, log đúng assignment. DELETE đặt deleted_at, giữ phút cũ.
5. REJECTED có lý do; APPROVED có thể không decision_reason. Người gửi WITHDRAWN khi PENDING; terminal không xử lý lại. Dữ liệu nền cũ có đường từ chối/rút và gửi mới, không tự sửa đề nghị đã gửi.
6. Đổi PTCM: người mới có trách nhiệm tiếp nhận yêu cầu; không phụ thuộc PTCM đã giao task. Người gửi mất quyền không tự chuyển yêu cầu thành duyệt, PTCM xử lý theo điều kiện hiện hành.

Tối đa một PENDING/log cho UPDATE/DELETE; ADD có thể nhiều khoản thực tế, chống retry bằng idempotency. Không giữ request.version vì chưa có sửa/bổ sung nội dung Pending. Sửa effort không thay accepted_at, không mở lại task hoặc sửa snapshot nộp. Không suy rằng mọi APPROVED phải có history; ADD truy bằng result_log_id.

## 6. Tổng hợp, khóa và cạnh tranh

actual_hours task/người = SUM(duration_minutes của log có deleted_at NULL theo đối tượng)/60; làm tròn khi hiển thị, không cộng các số đã làm tròn.
Tổng project gồm cả effort task CANCELLED; hiển thị phần thành công/đang làm/phần hủy riêng.
Công việc chờ review vẫn tính toàn bộ estimate vào open_load; ghi thêm effort không giảm tải tồn đọng.

COMPLETED chỉ cho hiệu chỉnh hành chính, không thêm task/thực hiện công việc/bình luận mới.
Close project phải không còn yêu cầu effort PENDING; DONE → CLOSED và khóa dữ liệu cùng transaction.
Cancel project là đường khẩn cấp: kết thúc pending bằng REJECTED, reason PROJECT_CANCELLED; giữ log chính thức,
đánh dấu dữ liệu chưa đầy đủ khi có vấn đề chưa giải quyết, hủy task mở, giữ task DONE. Không tự duyệt các proposed values.
CLOSED/CANCELLED không mở lại/sửa sổ effort trong V2; đây là giới hạn phạm vi cần công bố trước thao tác khóa.

Tất cả đường ghi nhật ký trực tiếp/duyệt/xóa mềm cùng service và chiến lược khóa employee/date xuyên project.
Close/cancel/submit/reassign cạnh tranh với sửa log phải có thứ tự khóa và đối chiếu dữ liệu nền xác định.
Log phải khớp assignment cũ, không chuyển effort sang người mới khi reassign.
Recommendation đã lưu bất biến; chỉ lần chạy sau dùng số liệu/eligibility mới.

## 7. Giới hạn phải nói rõ khi bảo vệ

Tự khai có thể thiếu, sai hoặc cố tình thấp; có lý do và audit không chứng minh con số đúng.
Không có timestamp bắt đầu/kết thúc nên không tự phát hiện chính xác trùng giờ trong ngày.
Một người chịu trách nhiệm không đo đủ tổng chi phí lao động của mọi người hỗ trợ.
Tỷ lệ estimate/effort chịu sai lệch estimate, độ khó, học việc và chất lượng dữ liệu; chỉ dùng thành phần lịch sử thử nghiệm.
Phương án này ưu tiên effort có thể ghi nhận đủ hơn lịch 8h, không tuyên bố loại bỏ mọi sai lệch hay đo năng suất khách quan.

## Trách nhiệm nghiệm thu và phạm vi giản lược

Phụ trách chuyên môn (PTCM) của dự án phân công và nghiệm thu công việc, không trực tiếp nhận/thực hiện task trong chính dự án đó. Không giao thêm trách nhiệm nghiệm thu, không có reviewer thay thế do PM chỉ định. PM chỉ có quyền chuyên môn khi kiêm PTCM và cũng phải tuân thủ giới hạn không nhận task. PTCM có thể thực hiện task ở dự án khác nếu đủ điều kiện. Lưu người đánh giá, kết luận và thời điểm; link PR/bình luận/merge không tự thay thế quyết định nghiệm thu trong hệ thống.
Kỹ năng không tự hết hạn; danh mục không có trạng thái INACTIVE. PTCM duyệt hiệu chỉnh, không người review riêng. PM không kiêm PTCM nếu có đóng góp thì gửi yêu cầu như người thực hiện, không được tự duyệt từ trách nhiệm PM. Không thêm ngoại lệ hoặc người duyệt mới.
