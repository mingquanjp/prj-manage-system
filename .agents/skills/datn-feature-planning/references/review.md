# Hai lượt review

Giao sub-agent chỉ đọc, không sửa file/DB/network/Jira/Git mutation. Cung cấp yêu cầu, phạm vi đã duyệt, file và nguồn cần thiết; để agent tự tìm vấn đề, không dẫn kết luận hoặc coi tests xanh là thiết kế đúng.

Mẫu: “Review độc lập [proposal hoặc bộ contract/SQL/brief/tickets]. Đọc AGENTS/docs README và nguồn miền. Chỉ đọc. Tìm mâu thuẫn/thiếu sót gây lỗi thực khi triển khai; trả P1/P2/P3 với file:line, tình huống và hướng xử lý. Phân biệt chưa có runtime proof với lỗi thiết kế. Không mở rộng phạm vi. Nếu không có finding, ghi phạm vi/giới hạn đã kiểm.”

## Proposal

- Đúng FR/AC/nguồn hiện hành, quyền, scope/non-goals/luồng demo? Có nhầm mock/runtime hoặc nhận Done quá phần đã nối?
- Actor/target, role tổ hợp, self/ID khác, limited, PII, notification link có đúng? Điểm nghiệp vụ chưa chốt được hỏi thay vì tự chọn?
- Luồng thường/lỗi/mất response/reload/Back, transaction và phụ thuộc có khả thi? WBS theo đầu ra, phần tiếp nối không chặn demo ngoài ý muốn?
- Sơ đồ parse được và nhánh lỗi không bị hiểu thành commit/success? Xem preview khi thay hình, không chỉ đọc mã Mermaid.

## Contract/SQL/brief/tickets

- HTTP canonical, inventory/DTO/nullable/status/lỗi/security/body/query khớp proposal; route planned không nhận runtime.
- SQL head/preflight/mapping/dependency/up/down/dirty/fixture bảo toàn; thứ tự khóa/snapshot/stale/idempotency/commit ambiguity có thể thực thi; test race ghi đúng adapter thật hoặc giới hạn fake.
- Brief UI có URL/đích login/deep link/quyền/lỗi/clear dữ liệu; dùng lại visual, tách demo/server?
- WBS coverage/dependency/Rank/consumer nhất quán; AC/lệnh kiểm ở đúng tầng? Task thiết kế đã làm không thành remaining giả.
- Link/ref/sơ đồ/examples đúng; secret/PII/raw errors không vào DTO/log/evidence. Không nhận static/parser/review thay HTTP/DB/mail/browser.

Sửa lỗi rõ trong scope, quyết định sản phẩm mới hỏi Quân. Gửi kiểm lại findings và phần ảnh hưởng. P1/P2 chưa xử lý không nhận sẵn sàng; P3 xử lý hoặc ghi lý do hoãn. Báo kết quả thực, không khẳng định review khiến không thể còn lỗi.
