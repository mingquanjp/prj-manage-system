# Mẫu giao việc và bàn giao

Dùng phần cần thiết trong chat hoặc ticket; không bắt tạo file cho mỗi task. Quy ước chung ở [workflow](workflow.md), nguồn nghiệp vụ ở [mục lục](../README.md). Không sao chép toàn bộ nghiệp vụ vào brief.

## Trước triển khai

- Jira thật / mã feature-WBS / branch:
- Kết quả mong muốn và AC được bao phủ:
- Nguồn: đường dẫn + mục hiện hành; điểm chưa chốt:
- Phạm vi sửa / ngoài phạm vi:
- Đầu vào và task đang chặn; contract/bên sử dụng bị ảnh hưởng:
- Phần Quân tự viết Go / phần agent hỗ trợ hoặc triển khai:
- Cách kiểm phù hợp, điều kiện môi trường và bằng chứng cần có:

Ví dụ đặt nhánh/commit/PR dùng `DATN-2`; mã `AUTH-T01` giữ cho truy vết, không dùng thay mã Jira. Nếu nguồn mâu thuẫn hoặc cần đổi nghiệp vụ/schema, hỏi trước phần phụ thuộc. Không tự cấp DB test hoặc chạy migration chỉ vì ticket có chữ database.

## Bàn giao

- Đã thay đổi; là tài liệu/scaffold/mock/tích hợp thật:
- File/contract cập nhật:
- Lệnh và ca kiểm thực chạy, kết quả/môi trường:
- Chưa kiểm, blocker và quyết định cần chốt:
- Diff đã review; commit/PR thật nếu đã có:
- Bước tiếp và đầu ra cung cấp cho task phụ thuộc:

Bàn giao không thay PR description. Khi mở PR, dùng [PR template](../../.github/pull_request_template.md); không ghi đã pass hoặc đã merge cho việc chưa thực hiện.

Với việc dài nhiều phiên, chỉ tạo plan khi cần giữ ngữ cảnh: mục tiêu, link nguồn, các mốc, kết quả kiểm và bước tiếp. Jira giữ backlog/trạng thái; plan không nhân bản toàn bộ ticket và không trở thành nguồn quyết định nghiệp vụ mới.
