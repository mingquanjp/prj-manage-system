# Bổ sung nghiệp vụ — Nguồn nhân lực chung và lịch sử tham gia dự án

> **Mốc tài liệu trước rà soát 15/10/2026.** Nội dung không bị thay thế vẫn giữ; quyết định mới ưu tiên [phần hiện hành và Bàn giao](design/database-review-decisions-2026-10-14.md). Thiết kế hiện tại đọc [bản hợp nhất 32 bảng](design/database-proposal-2026-10-14.md) và dictionary 5A/5B. Các mô tả PM duyệt/ghi thay, assignment_events, JWT/refresh, version và số bảng cũ dưới đây không được dùng khôi phục quyết định đã bỏ. Giữ nội dung theo mốc lịch sử, không coi là kết quả kiểm thử schema mới.

Ngày 2026-10-13. Đã được người dùng đồng ý triển khai đồng bộ tài liệu. Bổ sung cho [baseline 2026-10-12](business-baseline-2026-10-12.md); chỉ thay phần phạm vi ứng viên và cách ghi nhận tham gia tương ứng. Không xác nhận code, migration, DB thật hoặc sơ đồ Astah đã được cập nhật.

## 1. Mô hình tổ chức được chọn

Công ty sử dụng nguồn nhân lực chung cho nhiều dự án. Không bắt buộc lập đội hoặc thêm nhân viên vào dự án trước khi giao task. PTCM hiện hành được phân công trực tiếp người đủ điều kiện trong công ty cho task của dự án mình phụ trách. Không thêm bước PM duyệt phân công, xin điều động, tìm ngoài đội hoặc booking nguồn lực.

Đây là mô hình nghiệp vụ của sản phẩm, không khẳng định mọi doanh nghiệp đều vận hành như vậy. Ưu điểm: chọn người theo yêu cầu từng công việc, giữ luồng UC/activity hiện tại. Giới hạn: chưa phục vụ doanh nghiệp bắt buộc cấp đội/duyệt điều động/giữ chỗ nguồn lực trước. Xem xét lại nếu xuất hiện yêu cầu đó.

PM vẫn chỉ định PTCM, bắt đầu triển khai, theo dõi và quản lý vòng đời dự án theo quyền đã chốt. PTCM đặc tả, phân công và nghiệm thu; không nhận task trong chính dự án phụ trách. Trách nhiệm dự án không đồng nghĩa vai trò toàn công ty hoặc quyền đọc mọi dữ liệu.

## 2. Tra cứu, gợi ý và xác nhận

- Cả manual và gợi ý tìm từ nguồn nhân lực toàn công ty, không lọc theo membership có sẵn. User/employee phải hợp lệ và loại PTCM của chính dự án.
- Gợi ý lọc kỹ năng bắt buộc rồi xếp hạng. Manual giữ ngoại lệ kỹ năng có cảnh báo/lý do đã chốt; không override quyền, trạng thái hoặc điều kiện PTCM. Không thay công thức ở vòng này.
- Tải hiện hành tổng hợp xuyên dự án từ công việc hiện được giao; không lấy số dự án thay tải, không coi tổng tải là lịch trống hoặc giờ còn lại chính xác.
- Hiển thị thêm số dự án đang có công việc chưa kết thúc để PTCM tham khảo. Không có giới hạn cứng hai dự án, không tự thêm điểm phạt/trọng số/ranking theo số dự án. Một hoặc hai dự án là bối cảnh sử dụng, không phải constraint.
- Draft không phải phân công. Hiển thị tác động tổng của các lựa chọn; confirm kiểm tra lại dữ liệu hiện hành và ghi nguyên tử. Sau commit, lần đọc/chạy mới phản ánh phân công mới; run cũ giữ snapshot và thời điểm, không sửa ngược.
- Không có ứng viên thì trả rỗng, không tự hạ điều kiện hoặc tự giao. Hệ thống đề xuất, PTCM quyết định.
- PM/PTCM chỉ xem dữ liệu nghề nghiệp được phép và tổng hợp tải/số dự án. Chỉ số xuyên dự án không cấp quyền xem tên, task, nhật ký, CV gốc hoặc hồ sơ nhạy cảm ngoài phạm vi được phép.

## 3. Ba khái niệm không được nhập nhằng

| Khái niệm | Nguồn và cách hiểu |
|---|---|
| Người đã từng được phân công trong dự án | DISTINCT người nhận từ assignments đã commit JOIN tasks theo project. Giữ cả lượt đã thay/thu hồi và task đã hủy; không suy rằng đã làm hoặc đã hoàn thành |
| Số dự án đang có công việc chưa kết thúc của một người | COUNT DISTINCT project của task hiện được giao cho người đó, trạng thái ASSIGNED/IN_PROGRESS/PAUSED/PENDING_REVIEW, project PLANNING/IN_PROGRESS. Nhiều task cùng dự án chỉ đếm một; PENDING_REVIEW vẫn tính vì chưa kết thúc |
| Kinh nghiệm/đóng góp thực tế | Dữ liệu thực hiện, kết quả và căn cứ hợp lệ; không suy từ việc có assignment hoặc có tên trong danh sách |

Người có task cuối cùng kết thúc hoặc bị thu hồi giảm số dự án đang có việc tương ứng, nhưng không mất lịch sử phân công. Task PENDING_ASSIGNMENT không đếm người cũ. PM/PTCM lấy từ trách nhiệm trên project, không tự cộng vào số dự án đang có việc thực hiện. Chỉ số này không đo tổng gánh nặng quản lý/chuyên môn của họ.

Mốc lần đầu được giao = MIN(thời điểm giao đã commit), không gọi là ngày chính thức gia nhập đội. Không dựa riêng vào work_logs vì thiếu log không chứng minh chưa đóng góp. Quyền từ lượt giao hiện hành và trách nhiệm dự án được kiểm tra riêng; danh sách lịch sử không cấp quyền.

## 4. Hệ quả thiết kế dữ liệu và sơ đồ

Thiết kế đích bỏ lưu riêng `project_members` nếu trách nhiệm duy nhất là lịch sử người được giao việc: dùng truy vấn/view thường từ assignments và tasks. Không thêm bảng thành viên, workload, số dự án hoặc giữ chỗ. Giữ assignments/events và các lịch sử gốc.

Migration 000001 vẫn có `project_members`, `joined_at`, `join_source` và chưa được sửa. Trước migration chuyển đổi phải kiểm kê dữ liệu thật, membership không có assignment, nguồn MANUAL và chênh mốc thời gian; bảo toàn dữ liệu độc lập hợp lệ, không tạo assignment giả hoặc xóa để ép khớp thiết kế. Xem [quyết định dữ liệu](design/database-step4c-shared-pool-2026-10-13.md).

UC phân công thủ công/theo gợi ý, tra cứu độc lập và activity giữ actor/luồng hiện hành. Đặc tả ghi rõ nguồn toàn công ty, không thêm UC quản lý đội. Không cần vẽ mỗi chỉ số hiển thị thành UC riêng. Bản Astah/ảnh/báo cáo gốc chưa được chỉnh trong vòng này.
