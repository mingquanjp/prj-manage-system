# Tài khoản demo TaskAssign

Dùng để chuẩn bị báo cáo UI cho DATN-35. Danh sách fixture được đối chiếu với `cmd/localdev/main.go`; cách chạy và bằng chứng xem [runbook](auth-runbook.md#7-datn-35--oap-t13--demo-xuyên-ui-db-và-gmail). Đây là tài khoản ứng dụng local, không phải mật khẩu Gmail hoặc tài khoản production/Neon.

## 1. Mở ứng dụng

- UI: <http://127.0.0.1:5174/login>.
- Swagger dự phòng: <http://127.0.0.1:8080/docs>.
- Cần chạy môi trường server theo [runbook](auth-runbook.md); prototype có mục chọn danh tính là chế độ mô phỏng, không dùng để chứng minh luồng DB thật.

## 2. Ba tài khoản local dựng sẵn

| Dùng khi demo | Email đăng nhập | Mật khẩu ban đầu | Mã nhân viên | Quyền và đích sau login |
| --- | --- | --- | --- | --- |
| HR quản lý danh mục | `quan1652005@gmail.com` | `12345678` | `LOCAL-001` | HR; vào danh mục tổ chức, tạo/đổi tên/xóa danh mục theo điều kiện tham chiếu |
| Admin cấp tài khoản | `quan1652005+admin@gmail.com` | `12345678` | `LOCAL-003` | ADMIN; vào hồ sơ, mở mục Tài khoản để tạo/tra cứu; danh mục chỉ đọc |
| Kiểm cổng buộc đổi mật khẩu | `quan1652005+limited@gmail.com` | `12345678` | `LOCAL-002` | ADMIN nhưng phải đổi mật khẩu trước; chưa được mở nghiệp vụ |

Mật khẩu trong bảng là dữ liệu kiểm thử công khai được seed cho local. Nếu đã đổi mật khẩu của một fixture trong lượt chạy hiện tại, dùng mật khẩu mới bạn đã đặt; bảng không ghi đè hoặc khôi phục mật khẩu đó. Chính sách đặt mật khẩu mới là tối thiểu 15 ký tự Unicode, tối đa 72 byte UTF-8, không bắt buộc trộn chữ hoa/số/ký tự đặc biệt.

Ba fixture không tự gửi email. Địa chỉ có `+admin`/`+limited` là định danh đăng nhập fixture; không cần mở hộp thư để dùng chúng.

## 3. Tài khoản nhân viên đã tạo qua UI và email

Lượt kiểm DATN-35 ngày 18–19/11/2026 đã tạo và kiểm tài khoản sau:

| Thông tin | Giá trị |
| --- | --- |
| Họ tên tổng hợp | Nhân viên Demo OAP |
| Mã đã lưu | `EMP000001` |
| Email đăng nhập | Email nhận thư cấp tài khoản mà bạn đã nhập khi tạo; xem lại thư hoặc danh sách Tài khoản bằng Admin |
| Mật khẩu hiện tại | Mật khẩu mới bạn tự đặt sau khi nhận thư; không lưu trong tài liệu |
| Quyền hệ thống | Không có role ADMIN/HR; xem hồ sơ chính chủ |
| Phòng ban | Phòng Công nghệ Demo |
| Vị trí | Kỹ sư phần mềm Demo |
| Cấp bậc | Junior Demo |
| Trạng thái đã kiểm | Đã đổi mật khẩu, đăng nhập mới và xem hồ sơ; kết thúc kiểm đã đăng xuất |

Đây là dữ liệu lịch sử của lượt kiểm. Sau đó môi trường đã được khởi động lại theo yêu cầu để người dùng tự diễn tập; không dùng bảng này để khẳng định tài khoản còn tồn tại trong DB hiện tại. Đối soát bằng danh sách Tài khoản trước khi demo. Mật khẩu tạm trong thư không còn dùng để đăng nhập sau khi đổi thành công. Không chép mật khẩu tạm/mới, token OAuth hoặc nội dung thư cấp tài khoản vào Git, Jira hay ảnh báo cáo.

## 4. Chuyển tài khoản theo luồng báo cáo

1. **HR**: tạo/đổi tên danh mục → reload → đăng xuất.
2. **Admin**: mở Tài khoản → tạo nhân viên bằng mailbox được phép → xem mã chính thức và kết quả gửi thư → đăng xuất. Kiểm email chưa có tài khoản trước khi tạo, không dùng lại email đã tồn tại.
3. **Nhân viên mới**: tự lấy mật khẩu tạm trong Gmail → đăng nhập → đổi mật khẩu → đăng nhập lại → xem Hồ sơ của tôi → reload → đăng xuất/Back.
4. **HR**: đăng nhập lại → thử xóa phòng ban đã gán cho nhân viên để thấy thông báo bị chặn.
5. Nếu email đến chậm, chỉ dùng tài khoản nhân viên đã chuẩn bị nếu đã xác nhận nó còn tồn tại trong lượt chạy hiện tại; nêu rõ đây là tài khoản đã chuẩn bị trước.

Chỉ dùng tài khoản limited dựng sẵn nếu muốn minh họa riêng màn buộc đổi; nó không thay bằng chứng tài khoản vừa được cấp qua Gmail. Không chiếu mật khẩu trong thư khi trình bày.

## 5. Lưu ý trước buổi báo cáo

- Giữ runner đang chạy nếu cần dùng lại tài khoản `EMP000001` và danh mục đã chuẩn bị. DB local tạm không phải dữ liệu bền vững qua các lượt runner.
- Dừng runner bằng Ctrl+C sẽ dọn DB tạm theo cơ chế hiện có. Lượt chạy mới dựng lại ba fixture; tài khoản tạo qua UI và danh mục vừa tạo không được phục hồi từ thư Gmail cũ.
- Không mặc định tài khoản mới vẫn là `EMP000001` ở lượt demo khác: mã hiển thị sau lưu mới chính thức, preview không giữ chỗ và sequence có thể có khoảng trống.
- Trước demo gửi thư, chạy kiểm kết nối Gmail theo [runbook Gmail](../../cmd/gmail-setup/README.md). Hộp thư gửi cần OAuth đã cấp quyền; không dùng mật khẩu ứng dụng local để đăng nhập Gmail.
- Sau reload, danh sách cho biết tài khoản đã lưu; không khôi phục kết quả gửi thư trước đó. Không tạo lại hoặc gửi lại tự động khi mất phản hồi.
