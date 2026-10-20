# Review đồng bộ nghiệp vụ — 2026-10-12

> LỊCH SỬ — báo cáo tại mốc 12/10, không phải xác nhận hiện trạng mới nhất. Quyết định sau đó ưu tiên theo [mục lục nguồn](../README.md), đặc biệt hiệu chỉnh nhật ký do PTCM duyệt, các trường thông tin cá nhân và thiết kế DB. Giữ nội dung dưới đây để truy vết, không áp dụng lại phần PM xử lý hiệu chỉnh hoặc điểm mở đã được chốt tiếp.

## Phạm vi và kết luận

Đồng bộ các quyết định trong trao đổi và UC/activity người dùng gửi vào [baseline hiện hành](../business-baseline-2026-10-12.md), README/CLAUDE và tài liệu miền `.claude/docs/`. Đây là review tài liệu, không xác nhận phần mềm hoặc sơ đồ Astah đã được triển khai/sửa.

Trong phạm vi quyền phân công, nghiệm thu và tra cứu vừa chốt, không còn phát hiện các quy định cũ đối nghịch trong bộ workflow hiện hành sau vòng kiểm tra cuối. Không tuyên bố toàn bộ thiết kế kỹ thuật hoặc nghiệp vụ chưa chốt đã hoàn thiện.

## Những mâu thuẫn đã sửa

| Nội dung cũ | Nội dung hiện hành | Nguồn chi tiết |
|---|---|---|
| PM chọn/gợi ý/xác nhận/thay người | PTCM hiện hành của đúng dự án; PM chỉ khi kiêm PTCM | roles_permissions, project_task_assignment_workflow, ai_recommendation_workflow |
| PTCM nhận task, có người khác nghiệm thu | PTCM không nhận task cùng dự án; không reviewer độc lập/thay thế | domain_rules, work_effort_workflow |
| Giao thêm trách nhiệm nghiệm thu | Bỏ cơ chế, UC và extension point đó; giữ UC Nghiệm thu của PTCM | product_scope, baseline |
| Chỉ kiểm tra ứng viên lúc tạo draft | Kiểm tra lại PTCM/candidate khi commit và đối ứng lúc thay PTCM | domain_rules, ai_recommendation_workflow |
| PM tải CV gốc từ tra cứu | PM/PTCM chỉ dữ liệu nghề nghiệp/tổng hợp theo quyền | roles_permissions, security, cv_extraction_approval_workflow |
| Tra cứu chỉ phục vụ PM phân công | Danh sách độc lập cho PM/PTCM; tìm người theo task là ngữ cảnh riêng | roles_permissions, employee_profile_workflow |
| Kế hoạch có thể bị hiểu là vòng duyệt mới | Chia sẻ bên ngoài, ví dụ Slack; không tích hợp/gửi–duyệt trong hệ thống | baseline, project_task_assignment_workflow |
| Luồng review không đạt đi thẳng nộp lại | Phải chỉnh sửa/thực hiện rồi nộp lại | project_task_assignment_workflow, testing_verification |

## Kiểm tra chiều ngược và giới hạn

- Chỉ cấm PTCM nhận task trong dự án phụ trách, không cấm nhận task tại dự án khác.
- Chỉ định PTCM cho người đang nhận task cũng phải bị chặn; không chỉ lọc người khi phân công.
- Không đổi tên tác giả/hủy đóng góp khi chuyển giao. Quy tắc không tự nghiệm thu vẫn áp dụng với đầu ra đã thực hiện trước khi trở thành PTCM; tránh xung đột khi chỉ định, không tạo reviewer thay thế. Đây là hệ quả bảo vệ quy tắc đã chốt, không mở quy trình mới.
- Giữ PM bắt đầu/hoàn thành/đóng dự án và xử lý hiệu chỉnh effort; không thay toàn bộ chữ PM bằng PTCM.
- Quyền hủy dự án/task hiện có được giữ, không tự suy từ thay đổi quyền phân công.
- Q/T, mã thử nghiệm, JSON kết quả và công thức không được sửa/tái chạy trong đợt này. Báo cáo thử nghiệm trước không chứng minh đã kiểm tra điều kiện PTCM mới.
- Các khoảng chưa chốt về trường cá nhân/HR chọn CV/DTO/chi tiết kỹ thuật trách nhiệm và tự duyệt correction của PM vẫn được ghi rõ, không mặc nhiên cấp quyền.

## Kiểm tra đã thực hiện

1. Đọc/đối chiếu các workflow hiện hành, ma trận quyền, state machine, gợi ý, CV, bảo mật, effort và kế hoạch kiểm thử.
2. Quét 7 cụm quy định lỗi thời trong 9 tài liệu nghiệp vụ liên quan: không còn khớp; kiểm tra 6 dấu hiệu quy định mới: đủ. Đây là kiểm tra văn bản, không chứng minh ngữ nghĩa tự động.
3. Kiểm tra 75 liên kết Markdown nội bộ trong 17 file (README, CLAUDE, baseline mới và 14 tài liệu miền): không có đích bị thiếu.
4. So sánh khối công thức `text` của workflow gợi ý trước/sau: 1 khối, giữ nguyên hoàn toàn.
5. `git diff --check`: không lỗi whitespace.
6. Bổ sung 22 scenario CM01–CM22 vào testing_verification.md, gồm ca hợp lệ, từ chối, cạnh tranh và riêng tư. **Đây là ca đặc tả, chưa chạy test phần mềm.**

## Phân biệt hiện hành và lịch sử

Baseline 2026-10-07 giữ nội dung để truy vết nhưng có thông báo phần bị thay thế. API/DB V2 có mục đối chiếu thay đổi 12/10 và cảnh báo chưa là hợp đồng kỹ thuật cuối. Prompt/brief Figma cũ được đánh dấu không dùng trực tiếp để triển khai mới; prototype chưa được sửa. Tài liệu nghiên cứu có ghi chú vai trò mới, không sửa kết quả công bố/thực nghiệm. Các review lịch sử và tài liệu gốc trong docs/thesis-context không dùng ghi đè quyết định mới.

Không thay code, migration, DB thật, công thức, kết quả thực nghiệm hoặc tệp Astah; không commit/stage, giữ các thay đổi có sẵn của người dùng. Thêm ngoại lệ .gitignore cho baseline và báo cáo này để không bị bỏ quên khi quản lý phiên bản.
