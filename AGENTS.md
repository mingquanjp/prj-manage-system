# Chỉ dẫn agent

## Nguồn và cách đọc
- Bắt đầu tại [mục lục nguồn](docs/README.md); chỉ đọc nguồn đúng miền, không nạp toàn bộ lịch sử.
- Quyết định hiện hành/Bàn giao và chốt tiếp nối có xác nhận ưu tiên phần cũ bị thay thế; workflow `.claude/docs/` là đặc tả triển khai.
- Code/migration cũ, Make/export, nghiên cứu và báo cáo gốc không ghi đè nghiệp vụ đã chốt. Xung đột chưa rõ: hỏi trước khi code.
- Frontend: đọc [web/AGENTS.md](web/AGENTS.md), [quy trình và thiết kế đã duyệt](web/README.md), rồi contract/workflow liên quan.
- Không tạo bản sao nghiệp vụ trong kế hoạch; dẫn đường dẫn và mục áp dụng.
- Khi nhận task/chuẩn bị Git: theo [workflow Jira–GitHub](docs/ai/workflow.md); dùng mã Jira thật trong branch/commit/PR, giữ checkpoint Go người dùng tự viết.
- Description PR viết cho người review chưa đọc chat: theo [chuẩn viết PR](docs/ai/workflow.md#viết-description-pr-cho-người-review). Nêu hành vi và lý do trước; tóm tắt kiểm thử, dẫn evidence chi tiết về runbook. Trước tạo/sửa PR, tự đọc lại để bỏ thuật ngữ thừa và nhật ký làm việc.
- Khi chuẩn bị/tiếp nối một đợt feature từ delivery-plan/FR: dùng [skill lập kế hoạch](.agents/skills/datn-feature-planning/SKILL.md); không áp toàn quy trình cho một ticket code đã có brief.

## Phạm vi và an toàn
- Chỉ làm đợt người dùng đã giao; trước sửa ghi ngắn mục tiêu, nguồn, ngoài phạm vi, tiêu chí đạt. Refactor giữ nguyên hành vi; phát hiện ngoài phạm vi thì báo riêng, không tự sửa hàng loạt code cũ.
- Không tự đổi nghiệp vụ/schema/công thức hoặc chọn dependency cho mục chưa chốt.
- Không chạy migration/seed/reset hoặc truy cập DB thật từ tác vụ tài liệu/UI. Không đọc/in bí mật `.env`.
- Không sửa tài liệu gốc `docs/thesis-context/` hoặc export `/Users/quanjp/Documents/DATN` từ công việc trong repo này.
- Giữ thay đổi có sẵn; không reset, commit, push, deploy hoặc cấp quyền rộng khi chưa được yêu cầu.
- Feature/fix làm nhánh riêng; feature dùng `feature/<tên>`, bug dùng `bug/<tên>` theo quy ước repo. Không tự chuyển/mất thay đổi người dùng.
- UI và bình luận code tiếng Việt; định danh kỹ thuật tiếng Anh.
- Comment theo [quy ước comment](docs/ai/code-comments.md): giải thích lý do/ràng buộc, doc comment nêu contract cho caller; không ghi nhật ký triển khai. Đọc khi viết code và rà lại cùng diff trước bàn giao.

## Viết code đơn giản

- Ưu tiên code dễ đọc, dễ giải thích khi bảo vệ đồ án: tên rõ nghĩa, luồng trực tiếp; chấp nhận dài hơn nếu dễ hiểu hơn.
- Dùng `if/else` thay toán tử ba ngôi cho code mới hoặc phần đang sửa.
- Ưu tiên thư viện chuẩn và công cụ đã có; không tự viết lại chức năng được hỗ trợ phù hợp hoặc thêm dependency chỉ để rút ngắn code.
- Chỉ tách hàm/abstraction khi có trách nhiệm rõ hoặc nhu cầu dùng lại thực tế. Không dựng helper/framework tổng quát cho vài trường cố định; có thể lặp đoạn ngắn nếu dễ hiểu hơn.
- Tránh xử lý và lưu dữ liệu trùng lặp; dựng payload một lần rồi truyền tiếp. Giữ kiểm tra cần thiết tại từng ranh giới frontend/API/backend.
- Chỉ xử lý ngoại lệ có căn cứ từ yêu cầu hoặc tác động thực tế; không thêm nhánh dự phòng giả định. Giữ bảo vệ quyền, dữ liệu, transaction, chống gửi trùng và kết quả thao tác chưa xác định dù ít xảy ra.

## Kiến trúc và kiểm tra
- Đọc [kiến trúc](.claude/docs/architecture.md) trước khi thêm tầng/package; matching không phụ thuộc DB/HTTP.
- Auth/upload/endpoint có ID: đọc [bảo mật](.claude/docs/security.md); quyền client không thay quyền server.
- Frontend dùng pnpm/lockfile; lệnh thực và mức kiểm tra tại [web/README.md](web/README.md).
- `node --test scripts/check-harness.test.mjs` và `node scripts/check-harness.mjs`: test harness/kiểm kê nguồn; thêm `--staged` cho script khi chuẩn bị bàn giao Git. Không kiểm chứng ngữ nghĩa nghiệp vụ.
- Backend: `go vet ./...`, `go test ./... -count=1`, `go build ./...` khi có thay đổi Go; kiểm tra môi trường trước, không tự cấp DB test thật.
- `make check` có fmt/tidy sửa file; không dùng như lệnh chỉ đọc hoặc cổng frontend.
- Kiểm tra theo rủi ro và chạy cổng bắt buộc của phạm vi; tận dụng test hiện có, bổ sung cho hành vi có thể bị ảnh hưởng, không viết test chỉ lặp cấu trúc implementation.
- Bàn giao: đã sửa, check thực chạy/kết quả, chưa kiểm tra, điểm mở và bước kế tiếp. Không gọi mock là API/transaction thật.

## Commit attribution
- Chỉ commit khi người dùng yêu cầu; ghi Co-Authored-By với danh tính agent được môi trường cung cấp, không giả danh người dùng hoặc bịa danh tính.
