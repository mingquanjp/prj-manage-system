# Cơ sở nghiệp vụ hiện hành — 2026-10-12

> **Mốc tài liệu trước rà soát 15/10/2026.** Nội dung không bị thay thế vẫn giữ; quyết định mới ưu tiên [phần hiện hành và Bàn giao](design/database-review-decisions-2026-10-14.md). Thiết kế hiện tại đọc [bản hợp nhất 32 bảng](design/database-proposal-2026-10-14.md) và dictionary 5A/5B. Các mô tả PM duyệt/ghi thay, assignment_events, JWT/refresh, version và số bảng cũ dưới đây không được dùng khôi phục quyết định đã bỏ. Giữ nội dung theo mốc lịch sử, không coi là kết quả kiểm thử schema mới.

**Bổ sung đã chốt:** [2026-10-13 — nguồn nhân lực chung](business-baseline-2026-10-13.md). Không lập đội trước; danh sách người từng được giao suy từ lịch sử phân công. Những phần khác giữ nguyên.

Quyết định đã chốt với người dùng về phân công, nghiệm thu, tra cứu và activity. Đây là **nghiệp vụ đích**, không xác nhận code, DB/API, Astah hoặc báo cáo gốc đã được sửa. Thay thế các quy định tương ứng trong [baseline 2026-10-07](business-baseline-2026-10-07.md); các quy định không bị thay thế tiếp tục áp dụng.

## 1. Trách nhiệm và quyền

| Chủ thể | Quyền trong phạm vi hiện hành |
|---|---|
| Giám đốc | Khởi tạo dự án, chỉ định/thay PM; xét duyệt kỹ năng từ CV và đề nghị kỹ năng; không duyệt tệp CV |
| PM | Chỉ định PTCM, theo dõi/điều phối cấp dự án, bắt đầu triển khai, hoàn thành/đóng theo điều kiện; không mặc nhiên phân công hoặc nghiệm thu |
| Phụ trách chuyên môn (PTCM) | Cấu trúc/đặc tả/yêu cầu kỹ năng; chọn task, tra cứu thủ công/chạy gợi ý, chọn người và xác nhận; thay/thu hồi người thực hiện; nghiệm thu |
| Nhân viên thực hiện | Xem công việc hiện được giao, thực hiện, ghi nhật ký, cập nhật trạng thái theo quyền, trao đổi, tự nộp kết quả |
| HR | Hồ sơ hành chính/CV nhân sự, xét duyệt yêu cầu thông tin cá nhân; không sửa kỹ năng chính thức để bỏ qua Giám đốc |
| Quản trị viên | Tạo/cập nhật/khóa/mở khóa/phân quyền tài khoản; không mặc nhiên có quyền chuyên môn hoặc HR |

PTCM là trách nhiệm trong dự án, không thêm role toàn hệ thống. PM có thể kiêm PTCM; quyền chuyên môn có được từ trách nhiệm đó. PM không kiêm PTCM không chạy gợi ý/xác nhận/thay/thu hồi phân công. Hệ thống chỉ đề xuất; PTCM quyết định cuối cùng. Điều kiện và quyền được kiểm tra ở server, không chỉ ẩn nút.

## 2. Không kiêm thực hiện và nghiệm thu trong cùng dự án

**PTCM không được phân công làm người thực hiện task trong chính dự án phụ trách**, kể cả PM kiêm PTCM. Ràng buộc chỉ theo dự án; vẫn có thể nhận task ở dự án khác nếu đủ điều kiện.

- Loại PTCM trước xếp hạng gợi ý; nhánh thủ công không cho chọn. Không dùng điểm KPI/kỹ năng cao hoặc lý do override để bỏ qua.
- Xác nhận, giao lại và thay người kiểm tra lại PTCM hiện hành; draft/run cũ không bảo đảm còn hợp lệ.
- Khi chỉ định/thay PTCM, người được chọn không còn nhận task mở (ASSIGNED/IN_PROGRESS/PAUSED/PENDING_REVIEW) trong dự án. Chuyển giao trước; PENDING_REVIEW phải trả lại theo luồng hiện hành rồi mới thay/thu hồi người thực hiện. Không âm thầm xóa assignment, log hoặc kết quả.
- Hai chiều chỉ định PTCM và phân công phải kiểm tra nhất quán trong transaction để không cùng commit thành trạng thái vi phạm. Đổi PTCM phải thu hồi quyền chuyên môn của người cũ; không đổi lịch sử quyết định.
- Không tự nghiệm thu đầu ra mình đã thực hiện, kể cả đóng góp trước khi chuyển giao task. Chỉ kiểm tra current_assignee là chưa đủ. Nếu thay PTCM tạo xung đột với đầu ra cũ còn phải nghiệm thu, phải chọn người khác không xung đột; không tự cấp quyền ngoại lệ. Trường hợp không có người phù hợp cần xử lý nhân sự ngoài hệ thống, không tự nghiệm thu hoặc khôi phục cơ chế reviewer thay thế.

Lịch sử task đã kết thúc không cấm một người được chỉ định PTCM về sau; giữ nguyên người thực hiện, người nghiệm thu và đóng góp quá khứ. Nhật ký hiệu chỉnh cho đóng góp cũ không biến thành phân công thực hiện mới.

## 3. Nghiệm thu đơn giản

Nhân viên nộp → hệ thống lưu/chờ nghiệm thu → PTCM hiện hành đánh giá → hệ thống cập nhật kết luận. Không đạt: trả IN_PROGRESS có lý do, nhân viên chỉnh sửa/ghi effort thực tế rồi nộp lại. Đạt: ghi kết luận, người đánh giá, thời điểm theo quy tắc hiện hành.

Bỏ trách nhiệm nghiệm thu độc lập cấp dự án, UC giao người nghiệm thu và phương án PM chỉ định reviewer thay thế. Link PR là minh chứng tùy chọn; đọc/bình luận/merge trên GitHub không tự cấp quyền hoặc tạo kết luận nghiệm thu trong hệ thống. Không tích hợp GitHub trong phạm vi này.

Thiếu nhật ký không tự chặn nghiệm thu đầu ra; không ép giờ giả. Giữ cơ chế hiệu chỉnh effort của PM, không chuyển quyền này sang PTCM chỉ vì chuyển quyền phân công. Tự duyệt hiệu chỉnh cho đóng góp của chính PM vẫn là điểm chưa chốt nếu phát sinh; không suy rằng cấm PTCM nhận task giải quyết mọi lịch sử của PM.

## 4. Kế hoạch và vòng đời dự án

PTCM chia sẻ kế hoạch chi tiết ngoài hệ thống, ví dụ tệp Word qua Slack. Không thêm UC gửi/duyệt kế hoạch, tích hợp Slack hay phân hệ sprint. Chia sẻ để PM theo dõi, không phải vòng phê duyệt phân công.

PM giữ thao tác bắt đầu triển khai: PLANNING → IN_PROGRESS khi có ít nhất một task ASSIGNED. Không lặp bước này khi giao thêm task sau đó. Theo dõi diễn ra xuyên suốt; một task đạt không tự hoàn thành dự án. Giữ điều kiện hoàn thành/đóng/hủy dự án trong [quy tắc miền](../.claude/docs/domain_rules.md), giữ quyền hủy dự án/task đã có, không tự mở rộng đợt này.

## 5. Tra cứu hồ sơ nhân sự

Giữ một chức năng danh sách/tìm kiếm độc lập tên **Tra cứu hồ sơ nhân sự**, dùng bởi PM và nhân viên đang được giao PTCM. Nhân viên thông thường chỉ có quyền hồ sơ của mình, không mặc nhiên xem người khác.

| Nhóm thông tin | Phạm vi |
|---|---|
| Thông tin nhân sự | Nhận diện/công tác cần thiết; không hồ sơ hành chính nhạy cảm |
| Kỹ năng và năng lực | Kỹ năng chính thức đã được xác nhận; không coi CV hay đề nghị đang chờ là năng lực đã duyệt |
| Kinh nghiệm và kết quả thực hiện | Tổng hợp nghề nghiệp được phép, nêu kỳ/số mẫu khi có; không raw task/log/minh chứng dự án khác |
| Tải công việc | Tổng hợp phục vụ lựa chọn; không đồng nhất với lịch rảnh hoặc giờ còn lại |

PM dùng để theo dõi/lựa chọn PTCM; PTCM dùng để lựa chọn người thực hiện. Không cấp quyền sửa hồ sơ, phân công xuyên dự án hoặc tải CV gốc người khác từ quyền tra cứu. Quyền chủ hồ sơ, HR quản lý CV và Giám đốc xem căn cứ kỹ năng được xét riêng. DTO/allowlist chi tiết còn phải thiết kế; không trả entity rồi giấu ở UI.

**Tra cứu nhân sự phù hợp** trong UC phân công là tìm/chọn theo task cụ thể, cùng nguồn dữ liệu chính thức nhưng khác ngữ cảnh; không nhân đôi hồ sơ hoặc tự sinh rank/score cho danh sách thủ công.

## 6. Đồng bộ UC và activity

- Tổng quan giữ PM và Nhân viên nối Tra cứu hồ sơ nhân sự. Note ở phân rã: “Nhân viên chỉ được tra cứu khi được giao phụ trách chuyên môn trong dự án.” Không cần lặp note ở tổng quan; đặc tả vẫn phải nêu điều kiện.
- Chuyển nhóm Phân công công việc sang Quản lý chuyên môn dự án. Giữ Nghiệm thu công việc trong nhóm này.
- Chỉ giữ extension point Trách nhiệm chuyên môn; bỏ Trách nhiệm nghiệm thu. Ký hiệu extension không thay kiểm tra quyền.
- Quan hệ bắt buộc Phân công công việc → Xác nhận phân công dùng include, không dùng đường thường để diễn tả bước. Note điều kiện/extension point gắn quan hệ extend tương ứng.
- Activity gồm Giám đốc, PM, PTCM, Hệ thống, Nhân viên. Bỏ lane Người nghiệm thu; chuyển lựa chọn/xác nhận phân công và đánh giá kết quả sang PTCM. Các thao tác lọc/chấm điểm/kiểm tra/lưu vẫn ở Hệ thống.
- Nhánh không đạt quay về trước bước thực hiện/chỉnh sửa rồi nộp lại, không chỉ trước bước nộp. Note: “PTCM không trực tiếp thực hiện task trong dự án phụ trách.”
- Activity tổng quan biểu diễn luồng chính; lỗi validation/rollback/không có ứng viên giữ trong đặc tả chi tiết. Không khẳng định ảnh/Astah đã sửa bởi việc cập nhật Markdown.

## 7. Phần giữ nguyên và giới hạn

Không thay công thức Q/T hoặc chốt công thức mức phù hợp, trọng số, thang kỹ năng, ngưỡng tải. Các báo cáo thực nghiệm là lịch sử, không tái diễn giải thành phép thử đã có ràng buộc PTCM mới. Các chỉ số chỉ được dùng theo trạng thái đề xuất/đã chốt riêng của chúng.

Giữ frozen specification, batch nguyên tử, giới hạn manual override kỹ năng, kiểm tra active/version, lịch sử đóng góp, quyền HR/Giám đốc và CV không cần phê duyệt. Không sửa code, migration, DB thật, file Astah hoặc tài liệu gốc trong docs/thesis-context.

Điểm còn cần thiết kế trước triển khai: trường thông tin cá nhân cần HR duyệt; quyền HR xác nhận CV hiện tại thay nhân viên; allowlist DTO tra cứu; chi tiết lưu trách nhiệm PTCM và xử lý thay/thu hồi ở trạng thái terminal. Không tự tạo quy trình reviewer/plan approval để lấp những khoảng này.

## 8. Nguồn vận hành

- [Vai trò và quyền](../.claude/docs/roles_permissions.md)
- [Quy tắc miền và trạng thái](../.claude/docs/domain_rules.md)
- [Dự án, phân công, thực hiện và nghiệm thu](../.claude/docs/project_task_assignment_workflow.md)
- [Gợi ý và điều kiện ứng viên](../.claude/docs/ai_recommendation_workflow.md)
- [Effort và hiệu chỉnh](../.claude/docs/work_effort_workflow.md)
- [Ma trận kiểm thử đích](../.claude/docs/testing_verification.md)

API/DB V2 và prompt prototype cũ là tài liệu chưa đồng bộ kỹ thuật, không dùng làm nguồn cấp quyền trái baseline này.
