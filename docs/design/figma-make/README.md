# Bộ đầu vào sửa giao diện DATN — 20/10/2026

> Phát triển hiện chuyển sang `web/` trong repo; xem [quy trình local và visual đã duyệt](../../../web/README.md). Contract nghiệp vụ ở bộ này còn dùng; prompt Make và thứ tự build Make bên dưới chỉ để tham khảo lịch sử, không tự chạy lại. Bố cục local đã duyệt ưu tiên visual Make cũ, không thay quyền/nghiệp vụ.

Bộ này chuyển nghiệp vụ hiện hành thành hướng dẫn sửa prototype. Nguồn chuẩn vẫn là [phạm vi](../../../.claude/docs/product_scope.md), [phân quyền](../../../.claude/docs/roles_permissions.md), [quy tắc trạng thái](../../../.claude/docs/domain_rules.md) và các workflow được dẫn trong ma trận. Không dùng bản import cũ trong Make hoặc kế hoạch cũ để ghi đè chúng.

File đích: [DATN trên Figma Make](https://www.figma.com/make/fOXzi51sKaRQbuSKmHZQmD/DATN). Đây là Make app, khác file Design có các frame chỉnh qua Plugin API.

## Thứ tự sử dụng

1. Đọc [brief hiện hành](ui-brief.md) và nạp [Guidelines](Guidelines.md) thay hướng dẫn cũ.
2. Dùng [ma trận màn hình và quyền](screen-contracts.md) để xác định hành vi; ID UI-01…UI-15 dùng truy vết, không phải UC mới.
3. Dùng [prompt theo đợt](launch-prompts.md), sửa trên file Make hiện có. Bắt đầu bằng đợt 1 về quyền và vòng đời dự án; kế tiếp phân công, thực hiện/nghiệm thu, rồi hồ sơ/CV/danh mục.
4. Kiểm tra bằng [checklist tình huống](review-checklist.md). Mỗi đợt phải ghi rõ đã sửa, đã kiểm tra trên preview, hoặc chưa dựng.
5. Các bản import trùng tên trong src/imports và plan cũ phải được thay/đánh dấu hết hiệu lực trong Make. Chỉ gắn tài liệu liên quan đợt đang sửa.

## Phần đã chốt và phần cần giữ chỗ

- Chốt tiếp nối địa chỉ 20/10: cần bổ sung khi hoàn thiện hồ sơ, không bắt Admin nhập lúc tạo; NULL khi chưa bổ sung, không xóa trắng khi đã có. Nhân viên đề nghị để HR duyệt, HR được sửa trực tiếp. Chỉ nhắc khi thiếu, không chặn công việc hoặc thêm trạng thái; không tra cứu chuyên môn. Xem [quyết định](../../business-baseline-2026-10-20.md). Quy tắc này thay prompt polish từng cho xóa trường tùy chọn; bản Make đã chạy không tự được cập nhật từ việc sửa các file cục bộ.

- Đã chốt 20/10: Admin tạo user và hồ sơ nhân viên ban đầu cùng một lần; HR quản lý hồ sơ về sau. Chi tiết ở [quyết định](../../business-baseline-2026-10-20.md).
- Công thức cuối, miền trọng số, thang điểm thành phần chưa chốt: dữ liệu điểm hiện có chỉ dùng minh họa, ghi rõ. Không cung cấp cấu hình công thức như chức năng đã hoàn thiện.
- Trường tra cứu chỉ chọn trong bốn nhóm nghề nghiệp đã cho phép; chỉ số lịch sử cụ thể còn phải chốt. Chưa dựng KPI tổng hợp tự đặt.
- Giới hạn tệp minh chứng ngoài CV, cách cấp/reset mật khẩu, thời hạn phiên và mapping CV trùng muộn còn cần đặc tả. Không điền số/quy tắc tùy ý.
- PDF text/DOCX ≤10 MiB là giới hạn riêng của CV, không tự áp cho mọi minh chứng.

## Trạng thái thực hiện

Bộ tài liệu đã được chuẩn bị để sửa Make. Việc có prompt/checklist không chứng minh prototype đã được sửa hoặc kiểm thử. Báo cáo đọc Make trước đây tại `docs/reviews/ui-readiness-review-2026-10-20.md` là tham khảo local, chưa nằm trong bộ tài liệu được Git theo dõi; không dùng làm bằng chứng kiểm thử hiện tại. Trạng thái frontend local và kết quả từng đợt xem [web/README.md](../../../web/README.md).
