# Kiểm tài liệu feature

Từ gốc repo, Node hiện có, không package mới. Truyền đúng file thực, bỏ loại không áp dụng thay vì tạo file rỗng:

```sh
node scripts/check-feature-plan.mjs \
  --proposal docs/proposals/backend-profile-change-approval.md \
  --tickets docs/proposals/backend-profile-change-approval-tickets.md \
  --openapi docs/api/profile-change-approval.openapi.json \
  --sql-plan docs/proposals/backend-profile-change-approval-sql-plan.md \
  --source .claude/docs/employee_profile_workflow.md
node --test scripts/check-feature-plan.test.mjs
node scripts/check-harness.mjs
git diff --check
```

Tool xuất hash file đầu vào, inventory API, số link/Mermaid/WBS/task/cạnh đã kiểm; không xuất SQL/PII. --source lặp lại cho nguồn thực sự review. Không tự xác nhận approval/review freshness/FR Done. Link chỉ kiểm file, không anchor. OpenAPI chỉ kiểm inventory/ref local, không là OAS/JSON Schema validator tổng quát. Dùng thêm test contract theo miền.

Format hỗ trợ: bảng proposal có cột WBS và mã như PCA-01; heading task `### [PCA-T01] ...`, dòng WBS/Is blocked by/Blocks và mục theo bộ PCA/OAP. Cho phép nhiều WBS/task nếu mỗi WBS giao một lần. Format khác/thiếu mapping báo lỗi, không đoán. Kiểm graph/mapping không chứng minh đủ nghiệp vụ hoặc estimate đúng.

## Mermaid

Tự tìm parser trong VS Code macOS tại extension mermaid-markdown-features/markdown-preview-out. Máy khác dùng --mermaid-dir hoặc MERMAID_PREVIEW_DIR trỏ thư mục đó. Chỉ dùng thư mục tin cậy vì Node import module local; không tải dependency mạng hoặc chạy mã trong diagram.

Hỗ trợ sequenceDiagram và flowchart/graph. Sequence dùng parser/DB Mermaid; flowchart dùng grammar với semantic callbacks không ghi model để chạy ngoài DOM. Chỉ kiểm cú pháp, không node/render/HTML/link thực thi; xem Markdown Preview khi đổi sơ đồ. Loại chưa hỗ trợ hoặc bundle VS Code đổi phải báo lỗi, không bỏ qua.

Có Mermaid mà không có parser thì fail. --skip-mermaid chỉ cho kiểm một phần, ghi INCOMPLETE/exit2, không nhận full PASS. Exit0 nghĩa cấu trúc được yêu cầu đạt; exit1 là lỗi. Proposal-only chỉ được nhận đã kiểm proposal, không contract/tickets chưa truyền vào.

Go/PG/FE/browser chạy khi triển khai đúng workflow/runbook, không từ tool này. Staged checks chỉ khi chuẩn bị Git được yêu cầu. Nguồn đổi sau review thì so hash và review lại phần ảnh hưởng.
