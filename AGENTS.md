# Chỉ dẫn agent

## Nguồn và cách đọc
- Bắt đầu tại [mục lục nguồn](docs/README.md); chỉ đọc nguồn đúng miền, không nạp toàn bộ lịch sử.
- Quyết định hiện hành/Bàn giao và chốt tiếp nối có xác nhận ưu tiên phần cũ bị thay thế; workflow `.claude/docs/` là đặc tả triển khai.
- Code/migration cũ, Make/export, nghiên cứu và báo cáo gốc không ghi đè nghiệp vụ đã chốt. Xung đột chưa rõ: hỏi trước khi code.
- Frontend: đọc [web/AGENTS.md](web/AGENTS.md), [quy trình và thiết kế đã duyệt](web/README.md), rồi contract/workflow liên quan.
- Không tạo bản sao nghiệp vụ trong kế hoạch; dẫn đường dẫn và mục áp dụng.

## Phạm vi và an toàn
- Chỉ làm đợt người dùng đã giao; trước sửa ghi ngắn mục tiêu, nguồn, ngoài phạm vi, tiêu chí đạt.
- Không tự đổi nghiệp vụ/schema/công thức hoặc chọn dependency cho mục chưa chốt.
- Không chạy migration/seed/reset hoặc truy cập DB thật từ tác vụ tài liệu/UI. Không đọc/in bí mật `.env`.
- Không sửa tài liệu gốc `docs/thesis-context/` hoặc export `/Users/quanjp/Documents/DATN` từ công việc trong repo này.
- Giữ thay đổi có sẵn; không reset, commit, push, deploy hoặc cấp quyền rộng khi chưa được yêu cầu.
- Feature/fix làm nhánh riêng; feature dùng `feature/<tên>`, bug dùng `bug/<tên>` theo quy ước repo. Không tự chuyển/mất thay đổi người dùng.
- UI và bình luận code tiếng Việt; định danh kỹ thuật tiếng Anh.

## Kiến trúc và kiểm tra
- Đọc [kiến trúc](.claude/docs/architecture.md) trước khi thêm tầng/package; matching không phụ thuộc DB/HTTP.
- Auth/upload/endpoint có ID: đọc [bảo mật](.claude/docs/security.md); quyền client không thay quyền server.
- Frontend dùng pnpm/lockfile; lệnh thực và mức kiểm tra tại [web/README.md](web/README.md).
- `node --test scripts/check-harness.test.mjs` và `node scripts/check-harness.mjs`: test harness/kiểm kê nguồn; thêm `--staged` cho script khi chuẩn bị bàn giao Git. Không kiểm chứng ngữ nghĩa nghiệp vụ.
- Backend: `go vet ./...`, `go test ./... -count=1`, `go build ./...` khi có thay đổi Go; kiểm tra môi trường trước, không tự cấp DB test thật.
- `make check` có fmt/tidy sửa file; không dùng như lệnh chỉ đọc hoặc cổng frontend.
- Bàn giao: đã sửa, check thực chạy/kết quả, chưa kiểm tra, điểm mở và bước kế tiếp. Không gọi mock là API/transaction thật.

## Commit attribution
- Chỉ commit khi người dùng yêu cầu; ghi Co-Authored-By với danh tính agent được môi trường cung cấp, không giả danh người dùng hoặc bịa danh tính.
