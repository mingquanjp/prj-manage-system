# Review sau cập nhật baseline V1

Ngày 2026-09-06. Phạm vi: CLAUDE.md, 11 tài liệu miền, baseline D01–D38, đối chiếu schema 000001 và skeleton.
Đây là review thiết kế bằng walkthrough, không phải kết quả chạy nghiệp vụ trên sản phẩm.
Người dùng đã đồng ý V1 và xác nhận không bắt buộc ML; không cần xác nhận lại toàn bộ bộ quyết định.

## 1. Kết luận

Luồng lõi đã được thống nhất ở mức baseline:
hồ sơ có duyệt → requirements → gợi ý/tra cứu → PM xác nhận → thực hiện/review → dữ liệu đánh giá.
Không còn giữ hai quy định đối nghịch về mandatory, partial batch, trả việc tự chạy giờ, nơi lưu pending skill, hay cookie logout.
Chưa thể kết luận toàn bộ phân tích thiết kế đã khép kín: còn lựa chọn xử lý dữ liệu thời gian bất thường, dữ liệu thực nghiệm và chi tiết vận hành/auth.
Code/DB và báo cáo gốc vẫn lệch thiết kế đích; đó là việc triển khai/đồng bộ còn lại, không phải đã được sửa bằng cập nhật Markdown.

## 2. Các mâu thuẫn đã xử lý

| Trước đây | Thiết kế thống nhất |
|---|---|
| CV được gọi “học máy thật” dù dự kiến dictionary/fuzzy | Mô tả đúng rule-based, không bắt buộc ML; chất lượng phải đo |
| Pending nằm trong employee_skills unique | Tách request; chính thức cũ tiếp tục dùng trong lúc chờ |
| Mandatory dưới mức vẫn lọt Top-N với penalty | Thiếu/dưới mức đều loại; manual override có lý do |
| Batch vừa chỉ lưu dòng hợp lệ vừa rollback toàn bộ | All-or-nothing, báo lỗi theo dòng nhưng không partial write |
| PM trả review về IN_PROGRESS | Về ASSIGNED; nhân viên chủ động start |
| Không có điểm cắt giờ khi thay người | Đóng phiên cũ, giờ gắn người/assignment cũ, người mới chưa chạy |
| Có group riêng nhưng vẫn cây task nhiều cấp | Project → Group → Task, group cùng project bắt buộc |
| Estimate đổi làm sai lịch sử | Khóa estimate/requirements sau giao đầu; không sửa mô tả để đổi scope |
| Params được coi đủ replay | Snapshot input/output/version đủ dữ liệu scorer |
| is_selected và assignment cùng quyết định | Assignment commit là nguồn chuẩn |
| Tất cả skill reject vẫn có thể APPROVED | Tổng hợp review đúng, publication CV độc lập |
| JWT không query DB nhưng lại đòi khóa tức thì | Kiểm tra session/status/role DB cho API nghiệp vụ, chấp nhận chi phí |
| Refresh cookie không được gửi tới logout | Path=/api/auth; logout thu hồi server và xóa đúng path |
| SameSite được coi đủ CSRF | Origin + signed double-submit/session, có test refresh |
| Ví dụ 15:00 → 10:30 hôm sau tính 3,5h | 4,5h theo lịch thường |

## 3. Các điểm vẫn cần chốt hoặc kiểm chứng

### F01 — P1: Dữ liệu thời gian khi quên pause hoặc bị khóa

Hiện V1 không cho sửa giờ tay; phiên chỉ đóng khi pause/nộp/thay người/thu hồi/hủy.
Nếu người đang làm bị khóa rồi hôm sau PM mới bàn giao, giờ lịch trong khoảng đó vẫn có thể được cộng.
Một phiên mở/người ngăn cộng trùng, nhưng không chứng minh người đó thực sự làm việc.
PerfScore vì vậy vẫn có thể dùng mẫu nhiễu dù task chỉ có một người và actual >0.

Đề nghị chốt tiếp: giữ log gốc bất biến, cho PM đánh dấu mẫu thời gian bất thường có lý do để loại khỏi PerfScore; không sửa giờ gốc.
Tùy chọn khác là tự đóng phiên tại thời điểm khóa, nhưng làm auth tác động workflow/time và cần thêm quy tắc audit/thông báo.
**Chưa áp dụng hai mở rộng này** vì đây là lựa chọn nghiệp vụ mới. Trước khi nghiệm thu M6/M5 cần quyết định.

### F02 — P1: Công thức hợp lý về cấu trúc, chưa có bằng chứng chất lượng

0,7/0,2/0,1 là trọng số thử nghiệm; 80h, 3/10 mẫu và thang 1–4 cũng là giả định.
Tất cả requirements mandatory và đều đáp ứng → SkillScore=1 cho mọi người, thứ hạng chỉ khác tải/PerfScore.
Mọi người có tải ≥80h → WorkloadScore=0, không phân biệt 80h với 160h bằng thành phần này (tie-break chỉ tác động khi tổng hòa).
Đạt/vượt mức đều cap=1: không thưởng kỹ năng vượt yêu cầu. Đây là chủ ý “đáp ứng”, không thước đo “giỏi nhất”.
PerfScore bằng 0,8333 khi estimate=actual, còn người thiếu mẫu dùng 0,5; “trung tính” ở đây là giá trị mặc định, không phải trung bình đã đo.

Cần dữ liệu task/candidate, người đánh giá chuyên môn độc lập, nhãn phù hợp và CV được phép sử dụng.
So sánh ba baseline, độ nhạy trọng số và báo giới hạn; không tăng độ phức tạp ML trước khi có bằng chứng cần thiết.
**Chưa có dataset/người đánh giá được xác nhận.** Không thể đóng mục tiêu “công thức phù hợp” chỉ bằng unit test.

### F03 — P2: Cấp/reset mật khẩu chưa đủ đặc tả

Đã chốt đổi mật khẩu, ADMIN reset và buộc đổi; còn min/max mật khẩu, hạn mật khẩu tạm và kênh giao an toàn.
Đây là phần cần hoàn thiện trước M1, không dùng seed password làm quy trình vận hành.
CSRF bootstrap/refresh khi access hết hạn phải có hợp đồng/test; không bỏ kiểm tra chỉ để Swagger chạy.

### F04 — P2: Vận hành CV còn thiếu thông số

Đã chọn loại file, dung lượng, local private storage và worker DB.
Còn thư viện parser, giới hạn giải nén/thời gian/bộ nhớ, retry/lease cụ thể, retention và backup DB+file.
Các tham số kỹ thuật có thể chốt lúc M7; không ảnh hưởng nguyên tắc CV/skill duyệt độc lập.
Nếu dùng CV thật phải có quyền sử dụng và giới hạn quyền truy cập; snapshot scorer không chứa CV thô.

### F05 — P1 triển khai: DB hiện tại chưa đáp ứng V1

000001 còn task_type/parent_task_id, group nullable, thiếu trạng thái task, thiếu requests, work log chỉ ngày/giờ, thiếu session/snapshot.
Đã có kế hoạch dữ liệu ở database_schema.md; chưa viết/chạy migrations, chưa kiểm tra dữ liệu thật lại.
Không backfill timestamp/approval giả để vượt constraint. Các bảng audit phải xem lại cascade.
Route health hiện trả err.Error() khi DB lỗi; yêu cầu đích đã sửa trong tài liệu nhưng code chưa sửa ở lượt này.

### F06 — P1 báo cáo: Chưa đồng bộ báo cáo gốc

DoAn.pdf/TeX giữ nguyên theo quy định thư mục chỉ đọc.
Cần bản báo cáo làm việc để sửa Chương 2 (phạm vi/use case), Chương 3 (công thức/giả định),
Chương 4 (ERD/state/sequence/API/package), Chương 5 (số liệu thật).
Nguồn hiện có chưa đủ bộ build đầy đủ (main TeX/hình/bibliography); không khẳng định đã biên dịch báo cáo.

## 4. Các diễn giải liên miền được bổ sung trong tài liệu

- WITHDRAWN của request CV là dòng bỏ qua, không approved, để review không treo pending vô hạn.
- Duyệt/từ chối publication không tự duyệt/thu hồi năng lực; thu hồi năng lực cần REMOVE có audit.
- Task terminal giữ người phụ trách cuối phục vụ lịch sử, không tính tải và không cho thực thi.
- Reassign trực tiếp dùng manual có lý do; muốn ranking mới đi qua unassign → run → confirm; khoảng chưa giao là có chủ đích, không quảng cáo atomic cả ba bước.
- Ngoài estimate/requirements, không sửa mô tả để lách thay đổi scope; bình luận không mở lại terminal.
- GET lịch sử đóng góp trả DTO riêng, không mở quyền task mới cho người đã được thay.

Đây là diễn giải để các quy tắc V1 vận hành được; nếu muốn reassign bằng gợi ý trong một hành động atomic,
cần thêm hợp đồng run cho task đang ASSIGNED/PAUSED, thay vì tái dùng luồng chỉ nhận PENDING_ASSIGNMENT.

## 5. Kiểm tra đã làm và giới hạn

- Kiểm tra chéo trạng thái, quyền, dữ liệu nguồn, giao dịch, API và ca kiểm thử của 11 tài liệu miền.
- Đối chiếu các cột/enum hiện có trong migration và health handler; không gọi DB hoặc chạy seed/migrate/reset.
- Tính kiểm tra độc lập 4 ví dụ công thức: skill=1/load=0/perf=0,5 →95; load=40 →85; load=80 →75; skill=0,75/load=0 →77,5.
- Các ví dụ xác nhận số học, không phải test của engine (engine chưa implement), không đo chất lượng gợi ý.
- Kiểm tra git diff --check; không sửa mã Go, không thêm dependency.

## 6. Trình tự tiếp theo

1. Chốt F01 với người dùng; xác định dữ liệu/người đánh giá F02.
2. Hoàn thiện F03 trước M1; F04 trước M7.
3. Review DDL/migration/backfill riêng, sau đó mới triển khai trên DB thử.
4. Đồng bộ bản báo cáo làm việc, rồi kiểm thử chuỗi end-to-end theo testing_verification.md.
