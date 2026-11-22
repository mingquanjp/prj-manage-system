# Ticket Jira — Đề nghị hồ sơ, HR duyệt và thông báo

Ngày21/11/2026. **Nội dung chuẩn bị nhập Jira, chưa có issue/key thật**. Đích [DATN / board1](https://datn-mingquanjp.atlassian.net/jira/software/projects/DATN/boards/1). Không tự gửi API Jira trong đợt này.

## 1. Nguồn và cách dùng

[proposal](backend-profile-change-approval.md), [OpenAPI canonical](../api/profile-change-approval.openapi.json), [SQL/test plan](backend-profile-change-approval-sql-plan.md), [FR/AC](../product/functional-requirements.md), [workflow](../ai/workflow.md). Đọc kiến trúc/bảo mật/frontend do proposal dẫn trước sửa; ticket không thay contract.

**1 Epic +12 Task**, ánh xạ1:1 PCA-01–12. T01 ghi nhận đầu ra thiết kế hiện có, không tính lại như công việc thiết kế còn lại. T02–T10 ưu tiên F1–F3 báo cáo26/11; T11/T12 là F4 tiếp nối, không chặnT10. Không hứa mọi task Done theo ngày hoặc số lượng. Review/test/sửa lỗi nằm trong task, không tạo task review riêng.

PCA-Tnn là mã nguồn nội bộ, không thay mã DATN; sau import lấy key thật cho branch/commit/PR. Metadata dự kiến: Epic source-pca-epic; Task source-pca-t01…source-pca-t12, nhãn chung profile-change-approval và datn-pca-v1; assignee chỉ sau xác minh accountId. Không tự đặt status/priority/sprint/deadline/estimate/remaining/actual. Links nguồn remote thêm sau commit/push được xác minh, chưa bịa SHA hoặc URL.

## 2. Epic

**Summary:** Đề nghị cập nhật hồ sơ, HR xét duyệt và thông báo trên UI thật

**Mục tiêu:** nhân viên gửi patch riêng, HR so before/proposed/current và quyết định một lần; owner thấy kết quả/official/lịch sử và thông báo đúng quyền. HR direct edit tách F4.

**Tiêu chí hoàn thành:**

- [ ] F1–F3 chạy xuyên UI/HTTP/PG, quyền self/HR/limited, stale/terminal/idempotency/notification atomic đúng contract.
- [ ] F4 chỉ nhận khi ba API/UI/PG/browser đã đạt; FR-009–012 báo theo AC thực, FR-059/060 chỉ phần hồ sơ, không tự nhận toàn app.
- [ ] Runbook/ma trận bằng chứng/demo/regression và giới hạn có thể chạy lại, không credential/PII; Done sau merge/bàn giao theo workflow.

**Ngoài phạm vi:** lock/reset/role/deactivate/CV/kỹ năng/dự án/công thức, email/push/realtime, áp Neon/deploy hoặc migration33bảng.

## 3. Rank, WBS và phụ thuộc

Rank là thứ tự nhận việc khi đủ đầu vào, không thay Blocks; một feature chính WIP. Mỗi cặp chỉ tạo một link A blocks B, không nhập đảo hai lần. 17 cạnh trực tiếp.

| Rank | Mã nguồn | WBS | Is blocked by | Nhóm |
|---|---|---|---|---|
| 1 | PCA-T01 | PCA-01 | — | design |
| 2 | PCA-T02 | PCA-02 | PCA-T01 | db |
| 3 | PCA-T03 | PCA-03 | PCA-T02 | db |
| 4 | PCA-T04 | PCA-04 | PCA-T03 | db |
| 5 | PCA-T05 | PCA-05 | PCA-T03, PCA-T04 | http |
| 6 | PCA-T06 | PCA-06 | PCA-T02 | http |
| 7 | PCA-T07 | PCA-07 | PCA-T05 | web |
| 8 | PCA-T08 | PCA-08 | PCA-T05 | web |
| 9 | PCA-T09 | PCA-09 | PCA-T06, PCA-T07, PCA-T08 | web |
| 10 | PCA-T10 | PCA-10 | PCA-T07, PCA-T08, PCA-T09 | e2e |
| 11 | PCA-T11 | PCA-11 | PCA-T02 | http |
| 12 | PCA-T12 | PCA-12 | PCA-T11, PCA-T08 | web |

## 4. Quy ước nghiệm thu

Agent triển khai, Quân đọc hiểu/thử/review phần quan trọng; checkpoint tự viết chỉ giao khi đã chốt brief. Không ghi trước pass hoặc coi mock là proof DB/browser. Môi trường DB phải qua guard, không fallback DATABASE_URL/.env/Neon, không chạy suite DB song song; race dùng hai connection/barrier, không sleep ngẫu nhiên. Không tự reset môi trường demo đang chạy. Secret/PII không vào Jira/log/evidence. Task còn thiếu runtime/mapping đầu vào phải ghi rõ, không auto Done.

## 5. Description từng Task

Bảng tình huống khi chuyển ADF phải có hai cột thật; giữ đầy đủ nội dung/ref, không chỉ upload checklist. Importer và hướng dẫn tại [README PCA](../../scripts/jira/README-pca.md); chuẩn bị script không đồng nghĩa đã tạo issue trên Jira.

### [PCA-T01] Bàn giao contract API, SQL plan và brief UI hồ sơ đã chốt

**Mục tiêu**

Bàn giao snapshot thiết kế hiện có để các task bắt đầu code đúng nguồn. Đây là ghi nhận thiết kế đã làm, không tạo scaffold hoặc review giả để lấp task.

**Đầu vào và điều kiện bắt đầu**

Proposal đã được chốt hướng triển khai; đọc đúng contract/SQL/brief hiện hành. Phần đầu vào chưa đạt không được coi là runtime đã sẵn sàng.

**Phạm vi thực hiện**

1. Đối chiếu proposal PCA-A01–A09, OpenAPI10 thao tác, SQL plan mục9 và mapping URL với các nguồn hiện hành.
2. Ghi đầu vào legacy notification còn thiếu đúng tại T02; không nhận mapping dữ liệu đích/SQL runtime là đã kiểm từ tài liệu.
3. Bàn giao WBS→ticket, lệnh kiểm tĩnh thực chạy, nguồn commit/PR khi có; không đoán mã DATN hoặc auto Done.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Thiết kế hiện có, nguồn không đổi | Tái sử dụng bằng chứng; không soạn lại nghiệp vụ/scaffold. |
| DB notification có dữ liệu chưa biết | T02 nhận guard/blocker mapping, không bịa dữ liệu hoặc gọi runtime pass. |

**Tiêu chí hoàn thành**

- [ ] Snapshot proposal/OpenAPI/SQL/ticket nhất quán; brief URL và lựa chọn đã xác nhận; kiểm tĩnh đạt.
- [ ] Diff/nguồn đã review, góp ý chặn xử lý/retest; bàn giao file/lệnh thực/kết quả/môi trường/giới hạn và đầu ra consumer. Có commit/PR khi được giao, Done sau merge/bàn giao theo workflow; không secret/PII hoặc nhận fake là DB/browser thật.

**Phụ thuộc và cách thực hiện**

- Is blocked by: —.
- Blocks: PCA-T02.
- WBS: PCA-01. Agent triển khai; Quân review/kiểm chứng. Estimate/remaining/actual: chưa xác định, không điền0h hoặc tính công sức đã làm thành remaining.

**Kiểm tra và bằng chứng**

Dùng lại bằng chứng thiết kế nếu nguồn không đổi.

Tại thư mục gốc: `node --test scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`, `node scripts/check-harness.mjs`, `git diff --check`. Đây là kiểm tĩnh, không chứng minh runtime.

**Đầu ra bàn giao**

Contract/brief/SQL cho T02–T12.

**Ngoài phạm vi**

Không code nghiệp vụ, chạy DB, tạo Jira hoặc gửi mail.

**Tham chiếu**

Proposal mục4–9; SQL plan mục1–9 và ma trận PCA-DB tương ứng. [proposal](backend-profile-change-approval.md), [OpenAPI canonical](../api/profile-change-approval.openapi.json), [SQL/test plan](backend-profile-change-approval-sql-plan.md), [FR/AC](../product/functional-requirements.md), [workflow](../ai/workflow.md). Đọc kiến trúc/bảo mật/frontend do proposal dẫn trước sửa; ticket không thay contract.

**Truy vết nhập Jira:** `PCA-T01`; nguồn `docs/proposals/backend-profile-change-approval-tickets.md`.

### [PCA-T02] Triển khai migration đề nghị hồ sơ, idempotency và thông báo có guard legacy

**Mục tiêu**

Cấp schema và fixture thật cho các use case hồ sơ, giữ auth/OAP hoạt động.

**Đầu vào và điều kiện bắt đầu**

Đầu ra đã bàn giao của PCA-T01; đọc đúng contract/SQL/brief hiện hành. Phần đầu vào chưa đạt không được coi là runtime đã sẵn sàng.

**Phạm vi thực hiện**

1. Kiểm PG18 owner/head/drift/dirty/dependency bằng runner hiện có; migration tiến thêm theo SQL plan9.2, không sửa000001–000005.
2. Profile JSON/state/FK/index, idempotency và notification recipient user_id; preflight từ chối notifications có dữ liệu/dependency lạ trước dirty, rồi guard lại dưới khóa trong DDL. Lỗi trong DDL rollback cấu trúc/dữ liệu; kiểm metadata dirty riêng theo driver, không force version.
3. Kiểm empty và legacy user_id≠employee_id với notification rỗng; notification có dữ liệu phải giữ nguyên khi từ chối. Down/up DB hy sinh, dirty/fault giữa DDL; cập nhật fixture/localdev/seed/consumer head liên quan.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Notification legacy không rỗng hoặc drift | Dừng atomically; không đoán type/content, DROP để làm xanh hoặc force version. |
| Nâng empty/legacy có auth và dữ liệu OAP | Giữ ID/link/hash/email/code/status/roles/session/audit; schema mới đủ constraints. |
| Down có request/key/notification | Từ chối atomically; không nhận down là backup. |

**Tiêu chí hoàn thành**

- [ ] PCA-DB01–03 có bằng chứng PG thật; JSON/state/FK/index/guard/down-up và auth/OAP regression đạt.
- [ ] Diff/nguồn đã review, góp ý chặn xử lý/retest; bàn giao file/lệnh thực/kết quả/môi trường/giới hạn và đầu ra consumer. Có commit/PR khi được giao, Done sau merge/bàn giao theo workflow; không secret/PII hoặc nhận fake là DB/browser thật.

**Phụ thuộc và cách thực hiện**

- Is blocked by: PCA-T01.
- Blocks: PCA-T03, PCA-T06, PCA-T11.
- WBS: PCA-02. Agent triển khai; Quân review/kiểm chứng. Estimate/remaining/actual: chưa xác định, không điền0h hoặc tính công sức đã làm thành remaining.

**Kiểm tra và bằng chứng**

Tại gốc: `go vet ./...`, `go test ./... -count=1`, `go build ./...`. Trước integration: `python3 scripts/authdb/manage.py check`; chỉ sau xác minh PG18 test sở hữu, chạy tuần tự `python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./migrations ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m`. Bổ sung package test mới nếu có; unit không thay integration.

Tại thư mục gốc: `node --test scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`, `node scripts/check-harness.mjs`, `git diff --check`. Đây là kiểm tĩnh, không chứng minh runtime.

**Đầu ra bàn giao**

Schema/fixture/guard cho T03/T06/T11.

**Ngoài phạm vi**

Không áp Neon hoặc chuyển notification legacy có dữ liệu khi chưa có mapping được review.

**Tham chiếu**

Proposal mục4–9; SQL plan mục1–9 và ma trận PCA-DB tương ứng. [proposal](backend-profile-change-approval.md), [OpenAPI canonical](../api/profile-change-approval.openapi.json), [SQL/test plan](backend-profile-change-approval-sql-plan.md), [FR/AC](../product/functional-requirements.md), [workflow](../ai/workflow.md). Đọc kiến trúc/bảo mật/frontend do proposal dẫn trước sửa; ticket không thay contract.

**Truy vết nhập Jira:** `PCA-T02`; nguồn `docs/proposals/backend-profile-change-approval-tickets.md`.

### [PCA-T03] Triển khai use case gửi và đọc đề nghị chính chủ với chống gửi trùng

**Mục tiêu**

Nhân viên gửi patch đúng quyền, lưu PENDING riêng và có thể đối soát sau mất phản hồi.

**Đầu vào và điều kiện bắt đầu**

Đầu ra đã bàn giao của PCA-T02; đọc đúng contract/SQL/brief hiện hành. Phần đầu vào chưa đạt không được coi là runtime đã sẵn sàng.

**Phạm vi thực hiện**

1. Service/port validate allowlist/Unicode/BR-001; hash canonical typed payload, UUIDv4 lowercase, key chỉ submit24h.
2. Repository khóa actor/recipient theo bộ ID tăng dần, recheck self/status, tra key trước stale; ghi request+notifications+record kết quả đầy đủ cùng transaction, không placeholder commit.
3. List/detail projection self hoặc HR theo scope hiện hành, current_values chỉ trường patch; count/items cùng snapshot, không trả raw JSONB.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Cùng key hoặc key khác payload/expired | Một tác động; replay hợp lệ cùng request_id, conflict theo contract; không âm thầm tạo lại. |
| Before stale/no-op/địa chỉ null hoặc trắng | Lỗi đúng contract, official không đổi, không notification. |
| Recipient nhiều role/actor cũng HR | Không nhân dòng/khóa đôi; recipient snapshot có recheck, hàng chờ vẫn dùng được nếu không có HR đủ điều kiện. |

**Tiêu chí hoàn thành**

- [ ] PCA-DB04/05/09 và snapshot projection có PG thật; service không import pgx/HTTP; lỗi/commit chưa rõ typed outcome.
- [ ] Diff/nguồn đã review, góp ý chặn xử lý/retest; bàn giao file/lệnh thực/kết quả/môi trường/giới hạn và đầu ra consumer. Có commit/PR khi được giao, Done sau merge/bàn giao theo workflow; không secret/PII hoặc nhận fake là DB/browser thật.

**Phụ thuộc và cách thực hiện**

- Is blocked by: PCA-T02.
- Blocks: PCA-T04, PCA-T05.
- WBS: PCA-03. Agent triển khai; Quân review/kiểm chứng. Estimate/remaining/actual: chưa xác định, không điền0h hoặc tính công sức đã làm thành remaining.

**Kiểm tra và bằng chứng**

Tại gốc: `go vet ./...`, `go test ./... -count=1`, `go build ./...`. Trước integration: `python3 scripts/authdb/manage.py check`; chỉ sau xác minh PG18 test sở hữu, chạy tuần tự `python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./migrations ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m`. Bổ sung package test mới nếu có; unit không thay integration.

Tại thư mục gốc: `node --test scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`, `node scripts/check-harness.mjs`, `git diff --check`. Đây là kiểm tĩnh, không chứng minh runtime.

**Đầu ra bàn giao**

Ports/use cases/projections cho T04/T05.

**Ngoài phạm vi**

Không HTTP/UI, không mở rút/bổ sung pending hoặc unique pending/employee.

**Tham chiếu**

Proposal mục4–9; SQL plan mục1–9 và ma trận PCA-DB tương ứng. [proposal](backend-profile-change-approval.md), [OpenAPI canonical](../api/profile-change-approval.openapi.json), [SQL/test plan](backend-profile-change-approval-sql-plan.md), [FR/AC](../product/functional-requirements.md), [workflow](../ai/workflow.md). Đọc kiến trúc/bảo mật/frontend do proposal dẫn trước sửa; ticket không thay contract.

**Truy vết nhập Jira:** `PCA-T03`; nguồn `docs/proposals/backend-profile-change-approval-tickets.md`.

### [PCA-T04] Triển khai transaction HR duyệt hoặc từ chối đề nghị hồ sơ

**Mục tiêu**

HR quyết định đúng một lần, không ghi đè dữ liệu nền đã đổi và không lưu nửa chừng.

**Đầu vào và điều kiện bắt đầu**

Đầu ra đã bàn giao của PCA-T03; đọc đúng contract/SQL/brief hiện hành. Phần đầu vào chưa đạt không được coi là runtime đã sẵn sàng.

**Phạm vi thực hiện**

1. Khóa user/employee/request theo giao thức; recheck HR/actor/target/linkage/PENDING, không dùng principal client làm quyền cuối.
2. Approve so before đúng từng trường, update official+terminal+notification atomically; reject có lý do, không cần before khớp và giữ official.
3. Kiểm hai connection/barrier approve/approve, approve/reject, pending khác/cùng trường, actor mất quyền/flag, target inactive theo PCA-A09; fault notification/commit.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Hai người cùng xử lý hoặc lần gọi thứ hai | Chỉ một quyết định/notification; còn lại409 REQUEST_ALREADY_RESOLVED. |
| Before cùng trường đổi/khác trường đổi | Cùng trường stale không ghi đè; khác trường không conflict giả. |
| Employee target inactive/user locked riêng | Approve employee inactive409; reject vẫn được; không áp điều kiện phiên actor cho target. |

**Tiêu chí hoàn thành**

- [ ] PCA-DB06–09 đạt PG thật; terminal bất biến qua ports, atomicity/races/actor=target được kiểm; commit ambiguity không auto retry.
- [ ] Diff/nguồn đã review, góp ý chặn xử lý/retest; bàn giao file/lệnh thực/kết quả/môi trường/giới hạn và đầu ra consumer. Có commit/PR khi được giao, Done sau merge/bàn giao theo workflow; không secret/PII hoặc nhận fake là DB/browser thật.

**Phụ thuộc và cách thực hiện**

- Is blocked by: PCA-T03.
- Blocks: PCA-T05.
- WBS: PCA-04. Agent triển khai; Quân review/kiểm chứng. Estimate/remaining/actual: chưa xác định, không điền0h hoặc tính công sức đã làm thành remaining.

**Kiểm tra và bằng chứng**

Tại gốc: `go vet ./...`, `go test ./... -count=1`, `go build ./...`. Trước integration: `python3 scripts/authdb/manage.py check`; chỉ sau xác minh PG18 test sở hữu, chạy tuần tự `python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./migrations ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m`. Bổ sung package test mới nếu có; unit không thay integration.

Tại thư mục gốc: `node --test scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`, `node scripts/check-harness.mjs`, `git diff --check`. Đây là kiểm tĩnh, không chứng minh runtime.

**Đầu ra bàn giao**

Decision ports/outcomes/fixtures cho T05.

**Ngoài phạm vi**

Không email, partial approval, đổi proposed, tự thêm cấm HR xử lý đề nghị của mình.

**Tham chiếu**

Proposal mục4–9; SQL plan mục1–9 và ma trận PCA-DB tương ứng. [proposal](backend-profile-change-approval.md), [OpenAPI canonical](../api/profile-change-approval.openapi.json), [SQL/test plan](backend-profile-change-approval-sql-plan.md), [FR/AC](../product/functional-requirements.md), [workflow](../ai/workflow.md). Đọc kiến trúc/bảo mật/frontend do proposal dẫn trước sửa; ticket không thay contract.

**Truy vết nhập Jira:** `PCA-T04`; nguồn `docs/proposals/backend-profile-change-approval-tickets.md`.

### [PCA-T05] Tích hợp năm API đề nghị hồ sơ với auth, CSRF và PostgreSQL

**Mục tiêu**

Cung cấp các thao tác request chạy thật cho UI self/HR.

**Đầu vào và điều kiện bắt đầu**

Đầu ra đã bàn giao của PCA-T03, PCA-T04; đọc đúng contract/SQL/brief hiện hành. Phần đầu vào chưa đạt không được coi là runtime đã sẵn sàng.

**Phạm vi thực hiện**

1. Wire POST self submit, GET request list/detail, POST approve/reject theo canonical; giữ16OAP và năm auth hiện có.
2. DTO allowlist/typed errors/no-store, business gate/quyền server; JSON strict64KiB thực đọc, duplicate keys/Unicode/query/header UUID/Origin/CSRF theo contract.
3. HTTP+PG/route inventory, timeout/commit chưa rõ; chỉ publish operation đã triển khai trên Swagger dev, không đổi giới hạn auth/OAP8KiB.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Self/HR/Admin-only/Director/limited/ID ngoài quyền | Đúng401/403/404, không PII và không Admin bypass. |
| Body65536/65537 không Content-Length, media/query/key/CSRF sai | Đúng400/403/413/415, không ghi. |
| Key replay/stale/terminal/DB lỗi | Đúng DTO/code/status, không Set-Cookie hoặc success giả. |

**Tiêu chí hoàn thành**

- [ ] Năm method/path có HTTP+PG thật, contract/route inventory khớp; auth/OAP regression đạt; không query DB trong handler.
- [ ] Diff/nguồn đã review, góp ý chặn xử lý/retest; bàn giao file/lệnh thực/kết quả/môi trường/giới hạn và đầu ra consumer. Có commit/PR khi được giao, Done sau merge/bàn giao theo workflow; không secret/PII hoặc nhận fake là DB/browser thật.

**Phụ thuộc và cách thực hiện**

- Is blocked by: PCA-T03, PCA-T04.
- Blocks: PCA-T07, PCA-T08.
- WBS: PCA-05. Agent triển khai; Quân review/kiểm chứng. Estimate/remaining/actual: chưa xác định, không điền0h hoặc tính công sức đã làm thành remaining.

**Kiểm tra và bằng chứng**

Tại gốc: `go vet ./...`, `go test ./... -count=1`, `go build ./...`. Trước integration: `python3 scripts/authdb/manage.py check`; chỉ sau xác minh PG18 test sở hữu, chạy tuần tự `python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./migrations ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m`. Bổ sung package test mới nếu có; unit không thay integration.

Tại thư mục gốc: `node --test scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`, `node scripts/check-harness.mjs`, `git diff --check`. Đây là kiểm tĩnh, không chứng minh runtime.

**Đầu ra bàn giao**

Năm API request/adapter input/error cho T07/T08.

**Ngoài phạm vi**

Không nối UI hoặc publish ba API F4 chưa triển khai.

**Tham chiếu**

Proposal mục4–9; SQL plan mục1–9 và ma trận PCA-DB tương ứng. [proposal](backend-profile-change-approval.md), [OpenAPI canonical](../api/profile-change-approval.openapi.json), [SQL/test plan](backend-profile-change-approval-sql-plan.md), [FR/AC](../product/functional-requirements.md), [workflow](../ai/workflow.md). Đọc kiến trúc/bảo mật/frontend do proposal dẫn trước sửa; ticket không thay contract.

**Truy vết nhập Jira:** `PCA-T05`; nguồn `docs/proposals/backend-profile-change-approval-tickets.md`.

### [PCA-T06] Triển khai API thông báo chính chủ, phân trang và đánh dấu đọc

**Mục tiêu**

Tải đúng own notifications, giữ unread_count/snapshot và quyền mở tài nguyên tách biệt.

**Đầu vào và điều kiện bắt đầu**

Đầu ra đã bàn giao của PCA-T02; đọc đúng contract/SQL/brief hiện hành. Phần đầu vào chưa đạt không được coi là runtime đã sẵn sàng.

**Phạm vi thực hiện**

1. GET notifications filter/page: own recipient user_id, total sau filter/unread_count toàn own chưa đọc cùng snapshot.
2. POST notification/{id}/read body{}, Origin/CSRF; read_at lần đầu, gọi lại204, own404 khi ngoài quyền, không cấp quyền resource.
3. Wire/gate/no-store/DTO/HTTP+PG và Swagger operation đã chạy; message tối thiểu không PII; chưa phát sự kiện ngoài hồ sơ.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Mở khung hoặc chuyển bộ lọc | Không tự read; list/total/unread_count đúng kể cả page rỗng. |
| Read hai lần/đổi ID/mất quyền tài nguyên | Own read204 giữ thời điểm đầu; ID khác404; không cần quyền tài nguyên để đánh dấu own. |
| Mutation request rollback | Không có notification từ thao tác lỗi; bằng chứng writer dùng từ T03/T04 khi sẵn. |

**Tiêu chí hoàn thành**

- [ ] PCA-DB10 + HTTP/PG snapshot/read/own gate đạt; nhận đúng phạm vi các sự kiện hồ sơ, không nhận toàn FR-059.
- [ ] Diff/nguồn đã review, góp ý chặn xử lý/retest; bàn giao file/lệnh thực/kết quả/môi trường/giới hạn và đầu ra consumer. Có commit/PR khi được giao, Done sau merge/bàn giao theo workflow; không secret/PII hoặc nhận fake là DB/browser thật.

**Phụ thuộc và cách thực hiện**

- Is blocked by: PCA-T02.
- Blocks: PCA-T09.
- WBS: PCA-06. Agent triển khai; Quân review/kiểm chứng. Estimate/remaining/actual: chưa xác định, không điền0h hoặc tính công sức đã làm thành remaining.

**Kiểm tra và bằng chứng**

Tại gốc: `go vet ./...`, `go test ./... -count=1`, `go build ./...`. Trước integration: `python3 scripts/authdb/manage.py check`; chỉ sau xác minh PG18 test sở hữu, chạy tuần tự `python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./migrations ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m`. Bổ sung package test mới nếu có; unit không thay integration.

Tại thư mục gốc: `node --test scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`, `node scripts/check-harness.mjs`, `git diff --check`. Đây là kiểm tĩnh, không chứng minh runtime.

**Đầu ra bàn giao**

Hai API own notifications cho T09.

**Ngoài phạm vi**

Không polling/realtime/email/push, trang thông báo mới hoặc quyền đọc resource từ link cũ.

**Tham chiếu**

Proposal mục4–9; SQL plan mục1–9 và ma trận PCA-DB tương ứng. [proposal](backend-profile-change-approval.md), [OpenAPI canonical](../api/profile-change-approval.openapi.json), [SQL/test plan](backend-profile-change-approval-sql-plan.md), [FR/AC](../product/functional-requirements.md), [workflow](../ai/workflow.md). Đọc kiến trúc/bảo mật/frontend do proposal dẫn trước sửa; ticket không thay contract.

**Truy vết nhập Jira:** `PCA-T06`; nguồn `docs/proposals/backend-profile-change-approval-tickets.md`.

### [PCA-T07] Nối UI hồ sơ cá nhân gửi đề nghị và lịch sử server

**Mục tiêu**

Nhân viên gửi phần thay đổi và thấy rõ pending/kết luận tách khỏi official.

**Đầu vào và điều kiện bắt đầu**

Đầu ra đã bàn giao của PCA-T05; đọc đúng contract/SQL/brief hiện hành. Phần đầu vào chưa đạt không được coi là runtime đã sẵn sàng.

**Phạm vi thực hiện**

1. Đọc web instructions; reuse profile/form/detail/history qua adapter async, giữ /profile và request_id deep link đã chốt; server/demo tách biệt.
2. Chỉ gửi before/proposed fullname/phone/address/explanation và key; không actor/schema_version/status/cơ cấu. Hủy form không POST, dữ liệu không đổi không có trong patch.
3. Xử lý validation/stale/401/403/503/mất response; không tự replay submit, GET lịch sử đối soát; clear PII/form theo phiên, không remount do background refresh.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Gửi rồi reload | Official giữ nguyên khi pending; lịch sử/detail server khôi phục đúng người. |
| Address chưa có/xóa trắng/phone null | BR-001 và contract; email/cơ cấu vẫn chỉ đọc ở self. |
| Mất response/back/logout/đổi user | Không success giả, tự POST lại, fallback demo hoặc giữ PII user cũ. |

**Tiêu chí hoàn thành**

- [ ] Adapter tests/pnpm check và browser backend PG thật đạt send/history/detail/reload/stale/phiên; Unicode/body boundaries theo contract.
- [ ] Diff/nguồn đã review, góp ý chặn xử lý/retest; bàn giao file/lệnh thực/kết quả/môi trường/giới hạn và đầu ra consumer. Có commit/PR khi được giao, Done sau merge/bàn giao theo workflow; không secret/PII hoặc nhận fake là DB/browser thật.

**Phụ thuộc và cách thực hiện**

- Is blocked by: PCA-T05.
- Blocks: PCA-T09, PCA-T10.
- WBS: PCA-07. Agent triển khai; Quân review/kiểm chứng. Estimate/remaining/actual: chưa xác định, không điền0h hoặc tính công sức đã làm thành remaining.

**Kiểm tra và bằng chứng**

Tại `web/`: `pnpm check`; thêm browser với backend PostgreSQL test thật, dữ liệu tổng hợp. Ghi bước/kết quả/môi trường/ca chưa kiểm, không dùng adapter HTTP giả thay browser/DB.

Tại thư mục gốc: `node --test scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`, `node scripts/check-harness.mjs`, `git diff --check`. Đây là kiểm tĩnh, không chứng minh runtime.

**Đầu ra bàn giao**

F1 self UI cho T09/T10.

**Ngoài phạm vi**

Không profile edit trực tiếp hoặc fake [] thay lịch sử chưa tải.

**Tham chiếu**

Proposal mục4–9; SQL plan mục1–9 và ma trận PCA-DB tương ứng. [proposal](backend-profile-change-approval.md), [OpenAPI canonical](../api/profile-change-approval.openapi.json), [SQL/test plan](backend-profile-change-approval-sql-plan.md), [FR/AC](../product/functional-requirements.md), [workflow](../ai/workflow.md). Đọc kiến trúc/bảo mật/frontend do proposal dẫn trước sửa; ticket không thay contract.

**Truy vết nhập Jira:** `PCA-T07`; nguồn `docs/proposals/backend-profile-change-approval-tickets.md`.

### [PCA-T08] Nối UI HR hàng chờ, đối chiếu và quyết định hồ sơ

**Mục tiêu**

HR duyệt/từ chối trên dữ liệu thật, hiểu rõ before/proposed/current và xung đột.

**Đầu vào và điều kiện bắt đầu**

Đầu ra đã bàn giao của PCA-T05; đọc đúng contract/SQL/brief hiện hành. Phần đầu vào chưa đạt không được coi là runtime đã sẵn sàng.

**Phạm vi thực hiện**

1. Menu/URL /personnel/profile-requests và /{id}, list scope review, status/page và chi tiết server; giữ mapping login hiện có.
2. Approve body{}, reject reason; không sửa proposed, version hoặc replay mutation. Terminal không có action xử lý lại.
3. Hiện loading/lỗi/stale/target inactive/commit chưa rõ; recheck me khi403 phù hợp, clear dữ liệu mất quyền; focus/mobile đúng primitives.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| HR approve/reject rồi reload | Official/kết luận/lý do đúng DB; không UI success trước commit. |
| Hai tab xử lý hoặc before đã đổi | 409 rõ, tải lại; không tự override hoặc gửi lại. |
| Admin-only/limited/mất HR/ID lạ | Không mở từ role client hoặc mock; server giữ quyền cuối. |

**Tiêu chí hoàn thành**

- [ ] pnpm check và browser+PG thật approve/reject/stale/terminal/guard đạt; brief/navigation/PII được kiểm.
- [ ] Diff/nguồn đã review, góp ý chặn xử lý/retest; bàn giao file/lệnh thực/kết quả/môi trường/giới hạn và đầu ra consumer. Có commit/PR khi được giao, Done sau merge/bàn giao theo workflow; không secret/PII hoặc nhận fake là DB/browser thật.

**Phụ thuộc và cách thực hiện**

- Is blocked by: PCA-T05.
- Blocks: PCA-T09, PCA-T10, PCA-T12.
- WBS: PCA-08. Agent triển khai; Quân review/kiểm chứng. Estimate/remaining/actual: chưa xác định, không điền0h hoặc tính công sức đã làm thành remaining.

**Kiểm tra và bằng chứng**

Tại `web/`: `pnpm check`; thêm browser với backend PostgreSQL test thật, dữ liệu tổng hợp. Ghi bước/kết quả/môi trường/ca chưa kiểm, không dùng adapter HTTP giả thay browser/DB.

Tại thư mục gốc: `node --test scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`, `node scripts/check-harness.mjs`, `git diff --check`. Đây là kiểm tĩnh, không chứng minh runtime.

**Đầu ra bàn giao**

F2 HR UI cho T09/T10/T12.

**Ngoài phạm vi**

Không CV/kỹ năng/direct edit hoặc dashboard mới.

**Tham chiếu**

Proposal mục4–9; SQL plan mục1–9 và ma trận PCA-DB tương ứng. [proposal](backend-profile-change-approval.md), [OpenAPI canonical](../api/profile-change-approval.openapi.json), [SQL/test plan](backend-profile-change-approval-sql-plan.md), [FR/AC](../product/functional-requirements.md), [workflow](../ai/workflow.md). Đọc kiến trúc/bảo mật/frontend do proposal dẫn trước sửa; ticket không thay contract.

**Truy vết nhập Jira:** `PCA-T08`; nguồn `docs/proposals/backend-profile-change-approval-tickets.md`.

### [PCA-T09] Nối khung thông báo server và mở đúng tài nguyên hồ sơ

**Mục tiêu**

Người dùng nhận biết sự kiện và mở đúng ngữ cảnh sau kiểm quyền hiện hành.

**Đầu vào và điều kiện bắt đầu**

Đầu ra đã bàn giao của PCA-T06, PCA-T07, PCA-T08; đọc đúng contract/SQL/brief hiện hành. Phần đầu vào chưa đạt không được coi là runtime đã sẵn sàng.

**Phạm vi thực hiện**

1. Reuse chuông/panel hiện có theo screen contract, filter/page/unread_count; không trang mới hoặc polling.
2. Map resource_type/ID allowlist tới self query hoặc URL HR phù hợp, GET kiểm quyền thành công rồi read/đóng; mất quyền không mở nhưng vẫn cho own read.
3. Clear cache/PII/read requests khi đổi phiên, không đánh dấu đọc chỉ vì mở khung; Escape/ngoài/nút đóng/focus/mobile theo convention.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Mở chuông/filter hoặc chọn nội dung | Không tự read khi chỉ mở; đúng tài nguyên mới read sau thành công. |
| Link cũ sau mất HR/terminal/đổi phiên | Không quyền từ link, không PII cũ; read không phải approve. |
| Network/503/read response mất | Không lặp create/send; own read retry thủ công không thay quyết định miền. |

**Tiêu chí hoàn thành**

- [ ] Adapter tests/pnpm check và browser PG thật chứng minh submit→HR notification→decision→owner notification, filter/count/read/permission/focus.
- [ ] Diff/nguồn đã review, góp ý chặn xử lý/retest; bàn giao file/lệnh thực/kết quả/môi trường/giới hạn và đầu ra consumer. Có commit/PR khi được giao, Done sau merge/bàn giao theo workflow; không secret/PII hoặc nhận fake là DB/browser thật.

**Phụ thuộc và cách thực hiện**

- Is blocked by: PCA-T06, PCA-T07, PCA-T08.
- Blocks: PCA-T10.
- WBS: PCA-09. Agent triển khai; Quân review/kiểm chứng. Estimate/remaining/actual: chưa xác định, không điền0h hoặc tính công sức đã làm thành remaining.

**Kiểm tra và bằng chứng**

Tại `web/`: `pnpm check`; thêm browser với backend PostgreSQL test thật, dữ liệu tổng hợp. Ghi bước/kết quả/môi trường/ca chưa kiểm, không dùng adapter HTTP giả thay browser/DB.

Tại thư mục gốc: `node --test scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`, `node scripts/check-harness.mjs`, `git diff --check`. Đây là kiểm tĩnh, không chứng minh runtime.

**Đầu ra bàn giao**

F3 notification UI cho T10.

**Ngoài phạm vi**

Không email/push/realtime hoặc thêm sự kiện chưa chốt.

**Tham chiếu**

Proposal mục4–9; SQL plan mục1–9 và ma trận PCA-DB tương ứng. [proposal](backend-profile-change-approval.md), [OpenAPI canonical](../api/profile-change-approval.openapi.json), [SQL/test plan](backend-profile-change-approval-sql-plan.md), [FR/AC](../product/functional-requirements.md), [workflow](../ai/workflow.md). Đọc kiến trúc/bảo mật/frontend do proposal dẫn trước sửa; ticket không thay contract.

**Truy vết nhập Jira:** `PCA-T09`; nguồn `docs/proposals/backend-profile-change-approval-tickets.md`.

### [PCA-T10] Kiểm xuyên luồng đề nghị hồ sơ và bàn giao demo thứ Năm

**Mục tiêu**

Bàn giao F1–F3 tái hiện được cùng bằng chứng đúng tầng, lỗi/giới hạn và FR coverage.

**Đầu vào và điều kiện bắt đầu**

Đầu ra đã bàn giao của PCA-T07, PCA-T08, PCA-T09; đọc đúng contract/SQL/brief hiện hành. Phần đầu vào chưa đạt không được coi là runtime đã sẵn sàng.

**Phạm vi thực hiện**

1. Browser self gửi→HR nhận→approve/reject→owner nhận→official/lịch sử→reload/logout/Back; DB thật/dữ liệu tổng hợp.
2. Tổng hợp PCA-DB01–10/12, auth/OAP regression, bảy operation PCA đã chạy với canonical/Swagger; lỗi mới sửa/retest đúng task gốc hoặc Bug.
3. Mở rộng runbook hiện hữu, FR Done/Partial và kịch bản26/11; ghi lệnh/môi trường/giới hạn/hướng sau, không nhận F4 hoặc toàn notifications nếu chưa đạt.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Happy path và reject/stale/quyền/phiên/DB | Có bằng chứng browser/HTTP/DB đúng phần, không lấy fake thay persistence. |
| Báo cáo và dọn môi trường | Lặp lại được, không secret/PII; HTTPS/Neon/deploy chưa kiểm ghi riêng. |

**Tiêu chí hoàn thành**

- [ ] Go/PG/FE/contract/harness và browser bắt buộc có kết quả; lỗi chặn xử lý/retest; runbook/ma trận đủ, không gắn pass phần chưa chạy.
- [ ] Diff/nguồn đã review, góp ý chặn xử lý/retest; bàn giao file/lệnh thực/kết quả/môi trường/giới hạn và đầu ra consumer. Có commit/PR khi được giao, Done sau merge/bàn giao theo workflow; không secret/PII hoặc nhận fake là DB/browser thật.

**Phụ thuộc và cách thực hiện**

- Is blocked by: PCA-T07, PCA-T08, PCA-T09.
- Blocks: —.
- WBS: PCA-10. Agent triển khai; Quân review/kiểm chứng. Estimate/remaining/actual: chưa xác định, không điền0h hoặc tính công sức đã làm thành remaining.

**Kiểm tra và bằng chứng**

Tại gốc: `go vet ./...`, `go test ./... -count=1`, `go build ./...`. Trước integration: `python3 scripts/authdb/manage.py check`; chỉ sau xác minh PG18 test sở hữu, chạy tuần tự `python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./migrations ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m`. Bổ sung package test mới nếu có; unit không thay integration.

Tại `web/`: `pnpm check`; thêm browser với backend PostgreSQL test thật, dữ liệu tổng hợp. Ghi bước/kết quả/môi trường/ca chưa kiểm, không dùng adapter HTTP giả thay browser/DB.

Tại thư mục gốc: `node --test scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`, `node scripts/check-harness.mjs`, `git diff --check`. Đây là kiểm tĩnh, không chứng minh runtime.

**Đầu ra bàn giao**

Demo/runbook/bằng chứng F1–F3 và đầu vào bước kế tiếp.

**Ngoài phạm vi**

Không commit/push/merge/Done tự động, không chốt công thức hoặc mở workflow mới.

**Tham chiếu**

Proposal mục4–9; SQL plan mục1–9 và ma trận PCA-DB tương ứng. [proposal](backend-profile-change-approval.md), [OpenAPI canonical](../api/profile-change-approval.openapi.json), [SQL/test plan](backend-profile-change-approval-sql-plan.md), [FR/AC](../product/functional-requirements.md), [workflow](../ai/workflow.md). Đọc kiến trúc/bảo mật/frontend do proposal dẫn trước sửa; ticket không thay contract.

**Truy vết nhập Jira:** `PCA-T10`; nguồn `docs/proposals/backend-profile-change-approval-tickets.md`.

### [PCA-T11] Triển khai ba API HR đọc và sửa trực tiếp hồ sơ có audit

**Mục tiêu**

Hoàn thành backend FR-012 riêng, không giả thao tác quản lý là đề nghị nhân viên.

**Đầu vào và điều kiện bắt đầu**

Đầu ra đã bàn giao của PCA-T02; đọc đúng contract/SQL/brief hiện hành. Phần đầu vào chưa đạt không được coi là runtime đã sẵn sàng.

**Phạm vi thực hiện**

1. Service/ports/repository HR list/admin detail/PATCH; projection purpose-specific, expected_before_values/changes cùng keys, current permission và target dưới khóa.
2. Catalog KEY SHARE theo OAP, email canonical M2/unique; raw padding không gây stale giả; thay email thực update+audit+revoke target sessions atomically, no-op không audit/revoke.
3. Wire ba route đúng spec/DTO/gate/CSRF/body64KiB/no-store/Swagger; DB races catalog delete/email collision và rollback audit. Race với approve dùng adapter PCA-T04 nếu đã bàn giao; nếu chưa có, dùng writer PG có kiểm soát tuân cùng giao thức và ghi rõ giới hạn, kiểm lại adapter thật tại PCA-T12.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| HR đổi address/email/cơ cấu rồi reload | Đúng allowlist/FK; email đổi thật làm login lại, không cookie mới hoặc đổi role/status. |
| Email canonical không đổi/địa chỉ legacy | Không sửa padding/no-op giả; before so theo projection đã đọc, không cắt dữ liệu legacy. |
| Admin-only/stale/catalog mất/DB/audit lỗi | Đúng quyền/lỗi, không half-write; target inactive không bị tự activate. |

**Tiêu chí hoàn thành**

- [ ] PCA-DB11 phần F4 + HTTP/PG ba route, auth/catalog/self regression đạt; service không import HTTP/pgx, audit/quyền/PII đúng mục đích. Ghi rõ đối tác race approve thật hay writer có kiểm soát; không nhận writer đó là bằng chứng adapter approve.
- [ ] Diff/nguồn đã review, góp ý chặn xử lý/retest; bàn giao file/lệnh thực/kết quả/môi trường/giới hạn và đầu ra consumer. Có commit/PR khi được giao, Done sau merge/bàn giao theo workflow; không secret/PII hoặc nhận fake là DB/browser thật.

**Phụ thuộc và cách thực hiện**

- Is blocked by: PCA-T02.
- Blocks: PCA-T12.
- WBS: PCA-11. Agent triển khai; Quân review/kiểm chứng. Estimate/remaining/actual: chưa xác định, không điền0h hoặc tính công sức đã làm thành remaining.

**Kiểm tra và bằng chứng**

Tại gốc: `go vet ./...`, `go test ./... -count=1`, `go build ./...`. Trước integration: `python3 scripts/authdb/manage.py check`; chỉ sau xác minh PG18 test sở hữu, chạy tuần tự `python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./migrations ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m`. Bổ sung package test mới nếu có; unit không thay integration.

Tại thư mục gốc: `node --test scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`, `node scripts/check-harness.mjs`, `git diff --check`. Đây là kiểm tĩnh, không chứng minh runtime.

**Đầu ra bàn giao**

Ba API và fixtures cho T12.

**Ngoài phạm vi**

Không vô hiệu hóa/role/CV/kỹ năng, không profile người khác từ quyền Admin.

**Tham chiếu**

Proposal mục4–9; SQL plan mục1–9 và ma trận PCA-DB tương ứng. [proposal](backend-profile-change-approval.md), [OpenAPI canonical](../api/profile-change-approval.openapi.json), [SQL/test plan](backend-profile-change-approval-sql-plan.md), [FR/AC](../product/functional-requirements.md), [workflow](../ai/workflow.md). Đọc kiến trúc/bảo mật/frontend do proposal dẫn trước sửa; ticket không thay contract.

**Truy vết nhập Jira:** `PCA-T11`; nguồn `docs/proposals/backend-profile-change-approval-tickets.md`.

### [PCA-T12] Nối UI HR đọc và sửa hồ sơ trực tiếp, bàn giao FR-012

**Mục tiêu**

HR quản lý hành chính trên UI thật với lịch sử actor và dữ liệu hiện hành.

**Đầu vào và điều kiện bắt đầu**

Đầu ra đã bàn giao của PCA-T11, PCA-T08; đọc đúng contract/SQL/brief hiện hành. Phần đầu vào chưa đạt không được coi là runtime đã sẵn sàng.

**Phạm vi thực hiện**

1. Reuse personnel/profile/edit form, URL /personnel/employees và /{id}; tải catalog/chi tiết thật, không demo fallback hoặc action deactivate/CV chưa nối.
2. Gửi patch/before đúng trường thay đổi, giữ field readonly/phân biệt no-op/stale/email unique; giải thích target phải login lại sau email thực đổi.
3. Kiểm browser PG thật edit/reload/role/limited/NULL/address/email session và regression đề nghị; chạy lại race direct edit với adapter approve PCA-T04 thật, khép PCA-DB11; cập nhật runbook/FR-012/giới hạn.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Sửa trực tiếp và duyệt request cũ | Official giữ thay đổi mới; request cùng trường stale, actor history đúng. |
| Email đổi/503/mất response hoặc mất HR | Không success giả/replay; đối soát GET; không lộ hoặc giữ PII ngoài quyền. |

**Tiêu chí hoàn thành**

- [ ] pnpm check và browser+PG thật đạt; đủ AC FR-012 đã kiểm, không mở dữ liệu giả hoặc nhận FR-013; bàn giao lệnh/kết quả/giới hạn.
- [ ] Diff/nguồn đã review, góp ý chặn xử lý/retest; bàn giao file/lệnh thực/kết quả/môi trường/giới hạn và đầu ra consumer. Có commit/PR khi được giao, Done sau merge/bàn giao theo workflow; không secret/PII hoặc nhận fake là DB/browser thật.

**Phụ thuộc và cách thực hiện**

- Is blocked by: PCA-T11, PCA-T08.
- Blocks: —.
- WBS: PCA-12. Agent triển khai; Quân review/kiểm chứng. Estimate/remaining/actual: chưa xác định, không điền0h hoặc tính công sức đã làm thành remaining.

**Kiểm tra và bằng chứng**

Kiểm lại PCA-DB11 bằng adapter direct edit và approve thật: tại gốc chạy `go vet ./...`, `go test ./... -count=1`, `go build ./...`. Trước integration: `python3 scripts/authdb/manage.py check`; chỉ sau xác minh PG18 test sở hữu, chạy tuần tự `python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./migrations ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m`. Khi triển khai, ghi thêm tên test/package thực chứng minh race direct edit–approve; dùng hai connection/barrier, không sleep ngẫu nhiên hoặc chạy suite DB song song. Không fallback DATABASE_URL/.env/Neon; unit hoặc writer có kiểm soát không thay bằng chứng hai adapter thật.

Tại `web/`: `pnpm check`; thêm browser với backend PostgreSQL test thật, dữ liệu tổng hợp. Ghi bước/kết quả/môi trường/ca chưa kiểm, không dùng adapter HTTP giả thay browser/DB.

Tại thư mục gốc: `node --test scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`, `node scripts/check-harness.mjs`, `git diff --check`. Đây là kiểm tĩnh, không chứng minh runtime.

**Đầu ra bàn giao**

F4 UI/runbook và E2E tiếp nối.

**Ngoài phạm vi**

Không thiết kế lại FE, vô hiệu hóa/lock/reset/quyền/CV.

**Tham chiếu**

Proposal mục4–9; SQL plan mục1–9 và ma trận PCA-DB tương ứng. [proposal](backend-profile-change-approval.md), [OpenAPI canonical](../api/profile-change-approval.openapi.json), [SQL/test plan](backend-profile-change-approval-sql-plan.md), [FR/AC](../product/functional-requirements.md), [workflow](../ai/workflow.md). Đọc kiến trúc/bảo mật/frontend do proposal dẫn trước sửa; ticket không thay contract.

**Truy vết nhập Jira:** `PCA-T12`; nguồn `docs/proposals/backend-profile-change-approval-tickets.md`.
