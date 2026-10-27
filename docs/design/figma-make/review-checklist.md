# Checklist review prototype

Ngày 20/10/2026. Đây là tình huống phải kiểm tra trên preview, chưa phải kết quả pass. Đối chiếu [ma trận](screen-contracts.md); mỗi lần sửa ghi đạt/lỗi/chưa dựng và màn hình liên quan.

| ID | Tình huống | Kết quả bắt buộc |
|---|---|---|
| T01 | PM không kiêm PTCM vào dự án | Có theo dõi/chỉ định PTCM/vòng đời theo state; không tạo task/phân công/nghiệm thu/duyệt hiệu chỉnh |
| T02 | PTCM không kiêm PM | Có công cụ chuyên môn, không bắt đầu/hoàn thành/đóng dự án |
| T03 | PM kiêm PTCM | Có cả hai trách nhiệm, vẫn bị loại khỏi ứng viên task cùng dự án; không bị loại khỏi dự án khác chỉ vì trách nhiệm này |
| T04 | Giám đốc tạo dự án thiếu PM / hủy form | Không tạo dự án dở; người tạo không tự là PM |
| T05 | Thay PM / thay PTCM | Đúng actor; mất quyền cũ sau đổi, giữ tác giả lịch sử; không tự thu hồi task để cho phép chỉ định PTCM |
| T06 | Thêm task | Mô tả gồm đầu ra, criteria riêng, estimate dương, ≥1 requirement, không skill trùng/ngày thiếu cặp |
| T07 | Task đã từng được giao rồi thu hồi | Spec/estimate/requirements vẫn khóa; không nút sửa scope để lách |
| T08 | Chọn tất cả theo bộ lọc qua nhiều trang | Số lượng/phạm vi rõ, không chọn task không đủ điều kiện, có bỏ chọn; tập rỗng không chạy |
| T09 | Chạy gợi ý nhiều task | 0–3 ứng viên/task; không tự chọn Top-1; không tạo assignment hay tăng tải |
| T10 | Người BASIC so requirement BASIC | Đáp ứng mức; thiếu/dưới mandatory loại ở gợi ý; manual cảnh báo/lý do |
| T11 | Chuyển manual sau run | Có lý do, không rank/score trên lựa chọn manual; giữ truy vết run |
| T12 | Giao nhiều task cho cùng người trong draft, bỏ bớt một task | Tải dự kiến dùng tổng estimate còn chọn; tải chính thức chưa đổi |
| T13 | Confirm batch lỗi một dòng / thành công | Lỗi: mọi task giữ nguyên. Thành công: mọi task ASSIGNED đúng người, pending count/tải cập nhật khi quay lại |
| T14 | Run cũ có lịch sử/effort hoặc tải thay đổi | Không checkbox/lý do R05; giữ thời điểm ranking, cho chạy lại. Skill/spec/quyền/trạng thái đổi vẫn kiểm tra khi confirm |
| T15 | Ứng viên chưa từng nhận việc tại dự án / đang có việc ở hai dự án | Không cần add vào đội; không chặn dự án thứ ba, chỉ hiện tải/số dự án tham khảo |
| T16 | Nhân viên Start/Pause/Resume | Không tự sinh giờ; Start cần project IN_PROGRESS; không đổi state chỉ bằng kéo thả |
| T17 | Quên Start nhưng có effort thực tế hợp lệ | Khi còn quyền nhập trực tiếp được bổ sung, không luồng duyệt ngoại lệ riêng |
| T18 | Tổng ngày 480/481/1440/1441 phút xuyên dự án | 480 không bắt lý do vượt; 481–1440 cần lý do; 1441 chặn; không nhân đôi phút |
| T19 | Nộp tóm tắt có nội dung, không log hoặc không link | Cho nộp; “Chưa có thời gian ghi nhận”; không cờ effort_complete hay log 0 phút |
| T20 | Nghiệm thu chưa đạt | PTCM nhập lý do, task IN_PROGRESS; nhân viên chỉnh sửa rồi tự nộp lại |
| T21 | Chỉ sai giờ nhưng đầu ra đạt | Không bắt làm lại đầu ra hoặc chặn nghiệm thu vì thiếu giờ |
| T22 | Log PENDING_REVIEW / assignment cũ | Khóa Edit trực tiếp, chính chủ gửi hiệu chỉnh; PTCM hiện hành duyệt; PM không duyệt từ tư cách PM |
| T23 | Hiệu chỉnh ADD / UPDATE / DELETE / rút | Pending không đổi tổng; duyệt đổi đúng log; ADD có log kết quả; DELETE xóa mềm, không restore; terminal không xử lý lại |
| T24 | Một task DONE nhưng còn task mở | Dự án chưa tự COMPLETED; close bị chặn nếu correction pending |
| T25 | Project COMPLETED/CLOSED/CANCELLED | COMPLETED chỉ đọc/hiệu chỉnh hợp lệ/đóng; terminal chỉ đọc. Hủy chờ review không tạo REJECTED giả |
| T26 | Admin/HR/Giám đốc thao tác chéo quyền | Admin tài khoản; HR hồ sơ/cơ cấu; Giám đốc kỹ năng/nhóm kỹ năng; không quyền superset |
| T27 | HR sửa trực tiếp / nhân viên gửi profile | Nhân viên fullname/phone/address; HR email/cơ cấu; before_values cũ không ghi đè; pending không đổi hồ sơ. Address chưa bổ sung được NULL lúc tạo, UI nhắc nhưng không chặn làm việc; bổ sung/sửa cần nội dung, không xóa trắng/NULL, bỏ khóa là không đổi; không lộ ở tra cứu/gợi ý |
| T28 | HR upload CV / nhân viên chọn CV hiện tại / parse FAILED | HR không chọn thay; tệp hợp lệ vẫn dùng được; không nút duyệt CV |
| T29 | Yêu cầu kỹ năng không có task nội bộ / bổ sung pending / quyết định terminal | Vẫn gửi có giải trình; bổ sung cập nhật nội dung duyệt; terminal bất biến, bổ sung sau bằng yêu cầu mới |
| T30 | Giám đốc duyệt kỹ năng của mình / HR duyệt kỹ năng | Bị chặn; hai luồng MANUAL/CV cùng tiêu chí; không tự suy level từ tên CV |
| T31 | Tra cứu PM/PTCM / nhân viên thường / Giám đốc xem căn cứ | PM/PTCM nghề nghiệp giới hạn; nhân viên thường không xem người khác; Giám đốc chỉ căn cứ liên quan |
| T32 | Xóa danh mục còn tham chiếu | Bị chặn và có giải thích; không tạo INACTIVE để thay thế |
| T33 | Người đã bị thay mở đóng góp cũ | Chỉ đóng góp của mình; không thấy toàn task hiện tại/log người mới hoặc dữ liệu project khác |
| T34 | Đổi demo identity rồi quay lại danh sách/detail | Cùng state mô phỏng, quyền theo người/trách nhiệm; điều khiển demo không tự cấp quyền sản phẩm |
| T35 | Soát giao diện và keyboard | Đúng token chung, tiếng Việt rõ, focus/dialog hợp lý; không nút giả đã hoạt động hoặc KPI tự đặt |
| T36 | HR quản lý phòng ban/vị trí/cấp bậc; xóa mục còn tham chiếu | HR thêm/sửa tên; tên bắt buộc/không trùng, chỉ xóa khi không còn tham chiếu. Hồ sơ chọn cấp bậc từ danh mục công ty hoặc Chưa cập nhật; không nhập tự do/enum IT cố định. Đổi tên cập nhật hiển thị hiện hành, không viết lại snapshot lịch sử; nhân viên không đổi cấp bậc qua yêu cầu cá nhân. Không tạo INACTIVE để thay thế |

Review tài liệu tĩnh chỉ kiểm tra hướng dẫn nhất quán. Chỉ đánh dấu tình huống pass sau khi thao tác preview thật; build/typecheck xanh không thay kiểm tra quyền và luồng.
