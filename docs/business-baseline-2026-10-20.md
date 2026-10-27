# Chốt bổ sung phục vụ giao diện — 20/10/2026

Tiếp nối [baseline 12/10](business-baseline-2026-10-12.md), [nguồn nhân lực chung 13/10](business-baseline-2026-10-13.md) và [quyết định DB/Bàn giao](design/database-review-decisions-2026-10-14.md). Bổ sung quyền tạo ban đầu và quyết định địa chỉ đã được người dùng xác nhận; các ranh giới khác giữ theo nguồn hiện hành.

> Tiếp nối 26/10: phần cấp bậc và tổng số bảng ưu tiên [quyết định cấp bậc theo công ty](design/database-review-decisions-2026-10-14.md#quyết-định-tiếp-nối-26102026--cấp-bậc-theo-công-ty-hiện-hành). Mốc 20/10 dưới đây có 32 bảng; hiện tại là 33, cấp bậc dùng danh mục HR và khóa ngoại nullable. Các quy tắc tạo ban đầu/địa chỉ giữ nguyên.

## Quyết định đã xác nhận

**Admin tạo tài khoản kèm hồ sơ nhân viên ban đầu trong một lần; HR quản lý/cập nhật hồ sơ về sau.**

- Tạo user và employee trong cùng giao dịch, không tạo tài khoản mồ côi hoặc hồ sơ tạm chờ bước duyệt mới.
- Thông tin bắt buộc ban đầu theo dictionary: username/thông tin xác thực theo chính sách, mã nhân viên, họ tên, email công ty, vị trí công việc; phone/phòng ban/cấp bậc nullable theo thiết kế. Trạng thái và định danh do server quản lý theo hợp đồng hiện hành.
- Admin được nhập thông tin hành chính cần cho tạo ban đầu. Quyền tạo không cấp quyền sửa hồ sơ hành chính thường xuyên, duyệt thông tin cá nhân/kỹ năng hoặc quản lý danh mục thuộc HR/Giám đốc.
- HR tiếp tục cập nhật hồ sơ trực tiếp, duyệt đề nghị fullname/phone/address và quản lý email công ty/cơ cấu theo quyền hiện hành.
- User_roles chỉ ADMIN/HR/DIRECTOR; PM/PTCM vẫn theo dự án. Không tự gán trách nhiệm dự án khi tạo tài khoản.
- Riêng quyết định tạo ban đầu không thêm bảng, cột, trạng thái hoặc phê duyệt. Trường địa chỉ bổ sung theo quyết định tiếp nối bên dưới. Chính sách cấp/reset mật khẩu vẫn cần đặc tả; không khôi phục must_change_password.

## Quyết định tiếp nối — địa chỉ khi hoàn thiện hồ sơ

**Địa chỉ là thông tin phải bổ sung khi hoàn thiện hồ sơ, không bắt buộc tại thời điểm tạo tài khoản.** Quyết định này thay cách gọi “địa chỉ tùy chọn” và phương án bắt Admin nhập/đặt NOT NULL ngay đã được trao đổi trước đó.

- Admin không bắt buộc nhập địa chỉ khi tạo tài khoản kèm hồ sơ ban đầu. Không thêm bước HR duyệt tạo tài khoản.
- Thiết kế `employees.address` là `TEXT NULL`, mặc định NULL. NULL biểu thị chưa bổ sung, không phải địa chỉ giả hoặc dấu hiệu khóa quyền sử dụng.
- Chủ hồ sơ bổ sung/sửa qua yêu cầu cập nhật thông tin cá nhân hiện có; HR chấp thuận/từ chối. HR cũng được cập nhật trực tiếp theo quyền hành chính, có lịch sử.
- Khi cung cấp hoặc đề nghị thay đổi địa chỉ phải có nội dung sau khi bỏ khoảng trắng đầu/cuối. Không nhận NULL, chuỗi rỗng hoặc chỉ khoảng trắng như thao tác xóa; địa chỉ đã có chỉ được thay bằng địa chỉ có nội dung.
- Bỏ qua trường trong patch nghĩa là không thay đổi. `before_values.address` có thể NULL khi bổ sung lần đầu; giá trị đề nghị phải là chuỗi có nội dung. Pending/từ chối không đổi địa chỉ chính thức; chấp thuận đối chiếu dữ liệu nền đúng trường, không ghi đè giá trị đã đổi.
- Hồ sơ chưa có địa chỉ hiển thị “Chưa cập nhật” và nhắc bổ sung; không chặn đăng nhập, nhận việc hoặc thực hiện task vì thiếu địa chỉ. Không thêm trạng thái hoàn thiện hồ sơ, hạn bổ sung hoặc quy trình duyệt mới.
- Không trả địa chỉ trong tra cứu nghề nghiệp, ứng viên phân công hoặc snapshot gợi ý. Quyền xem hồ sơ/căn cứ theo mục đích không tự mở thêm quyền xem địa chỉ.
- Chỉ thêm một cột trong thiết kế, giữ 32 bảng; dùng JSON patch của `employee_profile_change_requests` hiện có. Không thêm bảng địa chỉ, danh mục địa giới, chỉ mục hoặc tích hợp bản đồ.
- Đợt này cập nhật tài liệu/dictionary/DBML, chưa sửa hoặc chạy migration. Khi chuyển đổi dữ liệu cũ, địa chỉ chưa biết giữ NULL, không bịa dữ liệu để lấp chỗ trống.

## Quyết định tiếp nối — điều kiện ghi nhật ký (21/10/2026)

Người dùng đã xác nhận luồng: Được phân công → Bắt đầu → thực hiện/ghi nhật ký → nộp → nghiệm thu. Quyết định này thay quyền ghi trực tiếp khi ASSIGNED/PAUSED trong R01 cũ, kể cả phần baseline 30/09 còn được dẫn chiếu.

- Chỉ người hiện được giao hợp lệ được tạo/sửa/xóa mềm nhật ký trực tiếp khi task IN_PROGRESS và project IN_PROGRESS. ASSIGNED phải Bắt đầu; PAUSED phải Tiếp tục trước, vẫn được đọc nhật ký đã có.
- Bắt đầu là xác nhận bắt tay thực hiện, không thêm bước xác nhận nhận việc; Start/Resume không sinh giờ hoặc chạy timer. Lưu nhật ký không tự đổi trạng thái.
- Giữ kiểm tra ngày thực tế thuộc lượt giao/giai đoạn triển khai, không tương lai, phút nguyên và tổng xuyên dự án. Không thêm cột thời gian hay bộ đo để suy giờ thực tế từ Start/Resume.
- Giữ khóa khi PENDING_REVIEW, hiệu chỉnh khi khóa theo quyền hiện hành và quy tắc thiếu log không chặn nộp/nghiệm thu. ASSIGNED/PAUSED không tạo quy trình duyệt ngoại lệ mới; dùng Bắt đầu/Tiếp tục, không sửa ngược lịch sử trạng thái.
- Không đổi cấu trúc DB, công thức hoặc mở rộng nghiệm thu.

## Bộ đầu vào giao diện

[Bộ Figma Make](design/figma-make/README.md) chuyển nghiệp vụ thành màn hình và tình huống kiểm tra; không phải nguồn để tự mở rộng nghiệp vụ. “Yêu cầu chỉnh sửa kết quả” là kết luận review chưa đạt; thu hồi phân công là thao tác PTCM. Không suy thêm nút nhân viên tự trả việc từ tên gọi cũ.

File đích do người dùng cung cấp: [DATN](https://www.figma.com/make/fOXzi51sKaRQbuSKmHZQmD/DATN). Cập nhật tài liệu không chứng minh đã sửa/kiểm thử prototype.
