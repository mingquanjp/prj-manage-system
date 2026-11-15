# Ticket Jira — Tổ chức, cấp tài khoản và hồ sơ cá nhân

Ngày soạn: **15/11/2026**. **Bản nội dung chuẩn bị nhập Jira; chưa tạo issue, chưa có Jira key.** Đích duy nhất: [DATN / board 1](https://datn-mingquanjp.atlassian.net/jira/software/projects/DATN/boards/1), không dùng site cũ hoặc project CoLearnia.

## 1. Nguồn, cách chia và trạng thái

Nguồn thiết kế: [proposal/WBS](backend-organization-accounts-profile.md), [OpenAPI OAP](../api/organization-accounts-profile.openapi.json), [SQL/migration/test plan](backend-organization-accounts-profile-sql-plan.md), [FR/AC](../product/functional-requirements.md). Ticket diễn đạt đầu ra/nghiệm thu, không thay các nguồn này. Khi có thay đổi contract, đối chiếu ticket bị ảnh hưởng trước code.

**1 Epic + 13 Task**, ánh xạ đủ 15 dòng WBS: OAP-01/02/03 gom thành OAP-T01 để bàn giao thiết kế đã làm; OAP-04–15 thành 12 task triển khai. Không tạo task review riêng cho từng tầng. Task OAP-T01 đã có phần lớn đầu ra ở repo, không là ước lượng công việc mới; Jira chỉ Done theo workflow sau bàn giao/merge phù hợp, không tự ghi Done khi nhập. Nếu chọn không nhập task ghi nhận này, phải giữ link đầu vào thiết kế và cập nhật các phụ thuộc trước import, không âm thầm bỏ một node.

- `OAP-Tnn` là mã nội bộ của ticket, **không phải Jira key** hoặc số WBS `OAP-nn`. Không tự dự đoán DATN-22…; lấy key thật sau tạo, rồi dùng trong branch/commit/PR.
- Metadata: Task dưới Epic, assignee Quân sau xác minh accountId; label chung `organization-accounts-profile`, nhãn đợt `datn-oap-v1`, nhãn định danh `source-oap-t01`… và một nhãn nhóm theo bảng. Epic dùng `source-oap-epic`. Không tự gán priority/sprint/deadline/estimate hoặc reviewer giả.
- **Rank là thứ tự nhận việc mặc định**, không đồng nghĩa mọi task High hoặc một dây chặn nối tiếp. Nhận task khi đủ đầu vào; WIP một feature chính. T10 có thể chuẩn bị cấu hình sớm sau T01 trong khi F1/F3 đang chạy; chưa có quyền hộp thư thì giữ blocker rõ.
- Các giờ mới chưa được ước lượng theo brief; không chép 105h auth, không điền 0h hoặc coi số task là tiến độ. Review/test/sửa lỗi nằm trong chính task.
- Quân đã chọn agent triển khai và cùng đọc hiểu/review phần quan trọng. Không giao bài tự viết lớn mặc định; nếu Quân nhận đoạn tự viết thì chốt checkpoint/test trước, agent không viết thay. Checkpoint học ưu tiên request→service→SQL→response, transaction/sequence/FK, quyền server và boundary DB–mail.
- Nguồn hiện chưa bàn giao commit/remote: link repo là đầu vào đọc, **không bịa URL commit**. Khi import đổi thành link nguồn đã commit/push và được xác minh; description dùng đủ nội dung, không chỉ copy checklist.

## 2. Epic

**Summary:** `Tổ chức, cấp tài khoản và hồ sơ cá nhân trên UI thật`

**Mục tiêu:** HR cấu hình phòng ban/vị trí/cấp bậc; Admin tạo nguyên tử user–employee/quyền ban đầu và gửi mật khẩu tạm qua Gmail; nhân viên dùng auth hiện có, đổi mật khẩu rồi xem hồ sơ chính chủ từ DB.

**Phạm vi và đầu ra:** ba feature F1/F2/F3 theo proposal, 16 API mới và UI server tương ứng. Tận dụng năm API auth; ưu tiên demo danh mục/profile trước khi Gmail đủ điều kiện. Phủ FR-015/005/006/007 theo AC đã chốt; FR-009 chỉ phần đọc chính thức, FR-064 chỉ phần role ban đầu khi tạo. Không tuyên bố trọn hai FR này.

**Tiêu chí hoàn thành:**

- [ ] F1 có CRUD thật/đúng quyền, stale/duplicate/reference và audit atomic; F3 đọc self, không PII của người khác.
- [ ] F2 tạo user–employee–roles/audit nguyên tử, code unique/không reuse, bỏ username; mail sau commit, trạng thái tạo/gửi tách biệt và có kiểm mailbox thật.
- [ ] UI chạy xuyên auth thường/hạn chế/đổi/login mới/reload/logout; server/demo tách biệt; có bằng chứng/runbook/giới hạn và review phù hợp.

**Ngoài phạm vi:** sửa/duyệt hồ sơ HR, thay role/khóa/reset tài khoản cũ, cấp lại/gửi lại mail, CV/dự án/công thức matching, workspace mock giả dữ liệu thật, áp Neon/toàn 33 bảng/deploy. Đã hoãn recovery khi người nhận không nhận thư; không bù bằng plaintext/outbox/nút gửi lại.

**Điều kiện vận hành:** code/migration chỉ trên môi trường test đã xác minh và được giao; OAuth/mail thật cần brief và quyền cụ thể. Epic không cấp quyền chạy SQL hoặc gửi email ngay khi được tạo.

## 3. Rank, mapping và phụ thuộc

Các cạnh trong bảng là phụ thuộc trực tiếp. `A blocks B` và `B is blocked by A` là cùng một liên kết, tạo **một lần** từ provider A tới consumer B; không nhập hai cạnh đảo chiều. Các phụ thuộc auth đã hoàn thành là điều kiện nền dùng lại, không tạo lại AUTH-Txx hoặc tự thêm link Jira chưa xác minh.

| Rank | Mã nội bộ / Summary | WBS | Nhóm | Is blocked by | Blocks |
|---|---|---|---|---|---|
| 01 | [OAP-T01] Bàn giao contract API và kế hoạch SQL cho tổ chức, tài khoản, hồ sơ | OAP-01, OAP-02, OAP-03 | design | — | OAP-T02, OAP-T08, OAP-T10 |
| 02 | [OAP-T02] Triển khai migration danh mục, hồ sơ và audit trên PostgreSQL test | OAP-04 | organization | OAP-T01 | OAP-T03, OAP-T06, OAP-T08 |
| 03 | [OAP-T03] Triển khai truy vấn và transaction CRUD danh mục tổ chức | OAP-05 | organization | OAP-T02 | OAP-T04 |
| 04 | [OAP-T04] Tích hợp 12 API danh mục với auth, CSRF và PostgreSQL | OAP-06 | organization | OAP-T03 | OAP-T05 |
| 05 | [OAP-T05] Nối UI danh mục tổ chức và navigation theo quyền server | OAP-07 | organization | OAP-T04 | OAP-T07, OAP-T12 |
| 06 | [OAP-T06] Triển khai API đọc hồ sơ chính chủ với projection tối thiểu | OAP-08 | profile | OAP-T02 | OAP-T07 |
| 07 | [OAP-T07] Nối màn hồ sơ cá nhân với API thật và điều hướng sau đăng nhập | OAP-09 | profile | OAP-T05, OAP-T06 | OAP-T13 |
| 08 | [OAP-T08] Bỏ username và chuẩn bị sequence, email unique cho tạo tài khoản | OAP-10 | accounts | OAP-T01, OAP-T02 | OAP-T09 |
| 09 | [OAP-T09] Triển khai transaction tạo user, employee, quyền và audit nguyên tử | OAP-11 | accounts | OAP-T08 | OAP-T11 |
| 10 | [OAP-T10] Tích hợp Gmail OAuth và adapter gửi thông tin tài khoản | OAP-12 | mail | OAP-T01 | OAP-T11 |
| 11 | [OAP-T11] Tích hợp API tạo tài khoản, mã dự kiến và danh sách tài khoản | OAP-13 | accounts | OAP-T09, OAP-T10 | OAP-T12 |
| 12 | [OAP-T12] Nối UI Admin tạo và tra cứu tài khoản với kết quả gửi email | OAP-14 | accounts | OAP-T05, OAP-T11 | OAP-T13 |
| 13 | [OAP-T13] Kiểm xuyên luồng tổ chức, cấp tài khoản, hồ sơ và bàn giao demo | OAP-15 | handoff | OAP-T07, OAP-T12 | — |

**17 liên kết Blocks**, không có chu trình. T06 chỉ cần T02 nên có thể làm trước T05 nếu hợp lịch; T07 chờ T05 vì dùng navigation server. T08 không bị F3 chặn; T10 không bị migration/create chặn. T11 cần cả transaction và mail adapter thật trước nghiệm thu; chuẩn bị handler/fake sớm không làm cả task Done. T12 cần navigation server và API Admin; T13 chỉ nhận khi hai nhánh UI đã ghép.

**Ưu tiên báo cáo UI:** T01 → T02 → T03 → T04 → T05 khép F1; T06 → T07 khép F3 bằng tài khoản local tổng hợp. T08 → T09 cùng T10 chuẩn bị F2; T11 → T12 → T13 khép toàn luồng. Đây là thứ tự dự kiến theo đầu ra, không hứa hoàn tất tất cả trước 19/11 và không thay điều kiện phụ thuộc.

## 4. Quy ước kiểm tra và bàn giao cho mọi ticket

- **DOC:** `node --test scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`, `node scripts/check-harness.mjs`, `git diff --check`; review nội dung/nguồn. Kiểm tĩnh không thay OpenAPI validator đầy đủ hoặc HTTP/DB.
- **GO:** kiểm môi trường rồi `go vet ./...`, `go test ./... -count=1`, `go build ./...`; thêm DOC khi đổi contract/tài liệu. Race tests theo phạm vi đồng thời thực tế, không nhận unit fake là SQL/transaction thật.
- **DB:** GO + PG18 test sở hữu qua guard/runner hiện hành, đường rỗng/legacy khi đổi migration; dùng đúng package/test đã viết. `python3 scripts/authdb/manage.py check` chỉ kiểm môi trường; không bịa tên test/lệnh hoặc fallback DATABASE_URL/.env/Neon. Hai connection/barrier thay sleep ngẫu nhiên cho race.
- **FE:** đọc web/AGENTS và web/README; adapter tests HTTP giả + `pnpm check` tại web, browser trên backend test cho hành vi cần nghiệm thu. Dùng lại tokens/primitives, không thiết kế lại toàn FE.
- **E2E:** GO/DB/FE/DOC và browser/mail thật đúng phần đã cấp quyền, ma trận kết quả từng tầng. Không dùng Swagger render, mock hoặc review như bằng chứng runtime.

**Done chung (áp dụng cùng AC riêng bên dưới):** diff/nguồn đã review, góp ý chặn được xử lý; ghi lệnh thực/kết quả/môi trường/giới hạn; có đầu ra cho consumer và commit/PR khi được giao. Không ghi trước pass, không đưa PII/secret vào Jira. Done sau merge/bàn giao phù hợp [workflow](../ai/workflow.md); không tự commit/push/merge/chuyển trạng thái. Lỗi phát hiện phải sửa trong task đang làm hoặc Bug liên kết, không đợi E2E mới kiểm chức năng.

## 5. Description từng Task

Mỗi mục dưới là description đầy đủ để chuyển Jira cùng Summary tương ứng. Link nguồn chung ở mục 1 và các mục tham chiếu cụ thể phải được giữ khi chuyển ADF; bảng tình huống phải có hai cột thật, không xuất thành các bảng rỗng như lần nhập trước.

### [OAP-T01] Bàn giao contract API và kế hoạch SQL cho tổ chức, tài khoản, hồ sơ

**Mục tiêu**

Bàn giao một bộ thiết kế nhất quán để bắt đầu code mà không chọn lại nghiệp vụ đã duyệt hoặc soạn lại phần đã có.

**Đầu vào và điều kiện bắt đầu**

Proposal/OAP-A01–A08 đã duyệt; OpenAPI 16 thao tác và SQL/test plan đã được review. Đây là ghi nhận đầu ra thiết kế đã thực hiện, không phải việc review mới cần làm lại.

**Phạm vi thực hiện**

1. Đối chiếu snapshot bàn giao của proposal, OpenAPI và SQL plan; giữ một nguồn cho mỗi loại contract.
2. Ghi rõ các đầu vào thực thi còn thiếu: mapping dữ liệu legacy tại môi trường được phép; brief OAuth/dependency/secret store/timeout ở task mail. Không biến đầu vào đó thành quyết định sản phẩm chưa chốt.
3. Bàn giao WBS→ticket, phụ thuộc và lệnh kiểm tĩnh; ghi link commit/PR nguồn khi có.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Các nguồn đã đủ và nhất quán | Dùng lại bằng chứng review/kiểm tĩnh; không tạo code scaffold hoặc viết lại tài liệu để lấp task. |
| Chưa có mapping/OAuth thực tế | Ghi đúng task tiếp nhận, không bịa dữ liệu hoặc gọi thiết kế là runtime đã đạt. |

**Tiêu chí hoàn thành**

- [ ] Có bộ nguồn thiết kế cùng snapshot, review đã xử lý và danh sách đầu vào bàn giao rõ ràng.
- [ ] Kiểm tĩnh contract/harness và diff đạt; ghi lại bằng chứng đã chạy của đợt thiết kế, không tự chuyển Done trước bàn giao/merge theo workflow.
- [ ] Diff/nguồn đã review, góp ý chặn đã xử lý; bàn giao lệnh thực/kết quả/môi trường/giới hạn và đầu ra cho task phụ thuộc. Không đưa secret/PII vào evidence; Done sau merge/bàn giao phù hợp workflow, không nhận mock là kiểm chứng DB/mail/browser thật.

**Phụ thuộc và cách thực hiện**

- Is blocked by: —.
- Blocks: OAP-T02, OAP-T08, OAP-T10.
- WBS: OAP-01, OAP-02, OAP-03. Agent triển khai; Quân đọc hiểu/thử/review; checkpoint tự viết chỉ giao khi chốt brief.
- Estimate/remaining/actual: chưa xác định; không ghi lại công sức đã có như công việc còn lại.

**Kiểm tra và bằng chứng**

Dùng lại bằng chứng có sẵn nếu nguồn không đổi. Kiểm này không chứng minh OpenAPI được server thực thi hoặc SQL đã chạy.

- Kiểm contract tại thư mục gốc: `node --test scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`. Đây là kiểm tĩnh, không thay HTTP/DB/runtime.
- Kiểm nguồn/diff tại thư mục gốc: `node scripts/check-harness.mjs`, `git diff --check`. Bàn giao lệnh đầy đủ đã thực chạy, kết quả, môi trường và giới hạn; không ghi trước pass hoặc đưa secret/PII vào log/evidence.

**Đầu ra bàn giao**

Contract HTTP và SQL plan cho OAP-T02/OAP-T08; brief đầu vào cho OAP-T10.

**Ngoài phạm vi**

Không code nghiệp vụ, chạy DB/migration, gửi thư, cấu hình Google hoặc tạo Jira trong bước bàn giao thiết kế.

**Tham chiếu**

Proposal mục 4–7/9; SQL plan mục 1–8; OpenAPI toàn bộ; FR-005/006/007/015, phần đọc FR-009. Nguồn: [proposal](backend-organization-accounts-profile.md), [OpenAPI](../api/organization-accounts-profile.openapi.json), [SQL plan](backend-organization-accounts-profile-sql-plan.md), [FR/AC](../product/functional-requirements.md), [workflow](../ai/workflow.md). Đọc kiến trúc/bảo mật/frontend do proposal dẫn trước sửa; không copy toàn bộ thiết kế vào ticket.

**Truy vết nhập Jira:** `OAP-T01`; nguồn `docs/proposals/backend-organization-accounts-profile-tickets.md`.

### [OAP-T02] Triển khai migration danh mục, hồ sơ và audit trên PostgreSQL test

**Mục tiêu**

Cung cấp schema M1 và fixture hợp lệ để triển khai danh mục và hồ sơ thật, giữ auth hoạt động.

**Đầu vào và điều kiện bắt đầu**

Thiết kế OAP-T01; môi trường PG18/runner authdb hiện có phải được kiểm lại quyền sở hữu. Mapping cấp bậc tổng hợp của fixture phải tường minh; mapping dữ liệu đích chưa rõ thì dừng phần nâng cấp đó.

**Phạm vi thực hiện**

1. Kiểm migration head/drift và guard; viết migration tiến thêm M1 theo SQL plan: job_levels, employee job_level_id/address, ràng buộc tên của ba danh mục và audit_events.
2. Hiện thực preflight/mapping và fixture rỗng/legacy có user_id khác employee_id; giữ ID/link/hash/email/status/role/session/code và seniority_level legacy theo kế hoạch.
3. Chạy migration runner thật, constraints/index/FK, dirty và down/up trên DB case hy sinh; kiểm seed/localdev/auth còn tương thích M1.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Tên trùng sau normalize, mapping chưa xác minh hoặc drift | Dừng theo preflight, không tự gộp/xóa/đoán cấp bậc hoặc force version. |
| Nâng rỗng và legacy giả | Schema M1 đủ, ba loại tên có UNIQUE/CHECK, dữ liệu cần bảo toàn giữ nguyên. |
| Down/lỗi giữa DDL | Kết quả rollback/dirty đúng driver; mất dữ liệu M1 được ghi rõ, không xem down là backup. |

**Tiêu chí hoàn thành**

- [ ] Có migration/fixture và kết quả hai đường nâng cấp trên DB test sở hữu; kiểm cả audit schema/FK và uniqueness từng danh mục.
- [ ] Có lệnh tái hiện, cách dọn và regression auth phù hợp M1; không sửa migration lịch sử hoặc áp Neon.
- [ ] Diff/nguồn đã review, góp ý chặn đã xử lý; bàn giao lệnh thực/kết quả/môi trường/giới hạn và đầu ra cho task phụ thuộc. Không đưa secret/PII vào evidence; Done sau merge/bàn giao phù hợp workflow, không nhận mock là kiểm chứng DB/mail/browser thật.

**Phụ thuộc và cách thực hiện**

- Is blocked by: OAP-T01.
- Blocks: OAP-T03, OAP-T06, OAP-T08.
- WBS: OAP-04. Agent triển khai; Quân đọc hiểu/thử/review; checkpoint tự viết chỉ giao khi chốt brief.
- Estimate/remaining/actual: chưa xác định; không ghi lại công sức đã có như công việc còn lại.

**Kiểm tra và bằng chứng**

OAP-DB01–04, OAP-DB12 phần M1. Bàn giao schema/fixture cho task repository; chưa nhận race create/email của M2 là pass.

- Tại thư mục gốc: `go vet ./...`, `go test ./... -count=1`, `go build ./...`. Unit tests không thay integration có build tag.
- Trước integration: `python3 scripts/authdb/manage.py check`; sau khi xác minh PG18 test sở hữu, chạy `python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./migrations ./internal/repository/postgres -count=1 -timeout=5m`. Đây là nhóm package hiện có; nếu task thêm package test mới thì bổ sung lệnh thực khi bàn giao. Không fallback DATABASE_URL/.env/Neon; không chạy nhiều lượt DB song song. Race dùng hai connection/barrier, không sleep ngẫu nhiên.
- Kiểm nguồn/diff tại thư mục gốc: `node scripts/check-harness.mjs`, `git diff --check`. Bàn giao lệnh đầy đủ đã thực chạy, kết quả, môi trường và giới hạn; không ghi trước pass hoặc đưa secret/PII vào log/evidence.

**Đầu ra bàn giao**

Schema M1, fixture/guard và bằng chứng bảo toàn cho OAP-T03/OAP-T06/OAP-T08.

**Ngoài phạm vi**

Không DROP username/cấp sequence M2; không map dữ liệu Neon, gửi mail hoặc dựng đủ 33 bảng.

**Tham chiếu**

Proposal mục 6.1/7; SQL plan mục 1–3/6–7; FR-015, nền đọc FR-009. Nguồn: [proposal](backend-organization-accounts-profile.md), [OpenAPI](../api/organization-accounts-profile.openapi.json), [SQL plan](backend-organization-accounts-profile-sql-plan.md), [FR/AC](../product/functional-requirements.md), [workflow](../ai/workflow.md). Đọc kiến trúc/bảo mật/frontend do proposal dẫn trước sửa; không copy toàn bộ thiết kế vào ticket.

**Truy vết nhập Jira:** `OAP-T02`; nguồn `docs/proposals/backend-organization-accounts-profile-tickets.md`.

### [OAP-T03] Triển khai truy vấn và transaction CRUD danh mục tổ chức

**Mục tiêu**

Đọc danh mục và ghi thay đổi của HR có kiểm quyền hiện hành, chống stale/duplicate và xóa còn tham chiếu.

**Đầu vào và điều kiện bắt đầu**

Schema/fixture OAP-T02; policy/ports auth hiện hành, quy tắc khóa và normalize trong SQL plan. Chỉ HR mutation; Admin được đọc để tạo tài khoản.

**Phạm vi thực hiện**

1. Viết repository/service cho departments/positions/job_levels; pagination/search literal và count/items cùng snapshot, không SELECT *.
2. Mutation recheck actor ACTIVE/cờ/role dưới khóa; so expected_before_name chính xác; ghi audit cùng transaction, rename no-op không sinh audit giả.
3. Hiện thực unique/FK/error mapping và DB tests hai connection/barrier theo giao thức khóa đã review, gồm actor Admin tham chiếu catalog bị xóa.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Tên 100/101 code point, padding, khác hoa/thường | Validate sau trim; UNIQUE chặn ghi trùng đồng thời, không bỏ dấu/NFC ngầm. |
| Tên nền đã đổi hoặc mục mất | 409 STALE_DATA hoặc 404 NOT_FOUND; không ghi/audit nửa chừng. |
| Xóa mục còn tham chiếu, kể cả inactive | 409 CATALOG_IN_USE; không set NULL/CASCADE; race không orphan hoặc chờ vô hạn. |
| Count/items có mutation xen giữa | Một response vẫn cùng snapshot; page rỗng giữ total đúng snapshot. |

**Tiêu chí hoàn thành**

- [ ] DB tests chứng minh CRUD cả ba loại, audit atomic/no-op, quyền hiện hành và stale/reference/race.
- [ ] Service không import pgx/HTTP; kết quả/lỗi qua port đủ để HTTP map đúng contract; query q không coi %/_ là wildcard.
- [ ] Diff/nguồn đã review, góp ý chặn đã xử lý; bàn giao lệnh thực/kết quả/môi trường/giới hạn và đầu ra cho task phụ thuộc. Không đưa secret/PII vào evidence; Done sau merge/bàn giao phù hợp workflow, không nhận mock là kiểm chứng DB/mail/browser thật.

**Phụ thuộc và cách thực hiện**

- Is blocked by: OAP-T02.
- Blocks: OAP-T04.
- WBS: OAP-05. Agent triển khai; Quân đọc hiểu/thử/review; checkpoint tự viết chỉ giao khi chốt brief.
- Estimate/remaining/actual: chưa xác định; không ghi lại công sức đã có như công việc còn lại.

**Kiểm tra và bằng chứng**

OAP-DB05/06/10, snapshot pagination. Race create với adapter create thật được kiểm tiếp OAP-T09; ở đây mutation giả lập phải tuân cùng giao thức, ghi đúng giới hạn.

- Tại thư mục gốc: `go vet ./...`, `go test ./... -count=1`, `go build ./...`. Unit tests không thay integration có build tag.
- Trước integration: `python3 scripts/authdb/manage.py check`; sau khi xác minh PG18 test sở hữu, chạy `python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./internal/repository/postgres -count=1 -timeout=5m`. Đây là nhóm package hiện có; nếu task thêm package test mới thì bổ sung lệnh thực khi bàn giao. Không fallback DATABASE_URL/.env/Neon; không chạy nhiều lượt DB song song. Race dùng hai connection/barrier, không sleep ngẫu nhiên.
- Kiểm nguồn/diff tại thư mục gốc: `node scripts/check-harness.mjs`, `git diff --check`. Bàn giao lệnh đầy đủ đã thực chạy, kết quả, môi trường và giới hạn; không ghi trước pass hoặc đưa secret/PII vào log/evidence.

**Đầu ra bàn giao**

Ports/use cases/projections và fixture race cho OAP-T04/OAP-T09.

**Ngoài phạm vi**

Không HTTP/UI, sửa hồ sơ HR, suy quyền từ tên danh mục hoặc xây audit framework toàn hệ thống.

**Tham chiếu**

Proposal mục 5/6.2–6.4/7; SQL plan mục 3/5/7; FR-015.AC1–AC2. Nguồn: [proposal](backend-organization-accounts-profile.md), [OpenAPI](../api/organization-accounts-profile.openapi.json), [SQL plan](backend-organization-accounts-profile-sql-plan.md), [FR/AC](../product/functional-requirements.md), [workflow](../ai/workflow.md). Đọc kiến trúc/bảo mật/frontend do proposal dẫn trước sửa; không copy toàn bộ thiết kế vào ticket.

**Truy vết nhập Jira:** `OAP-T03`; nguồn `docs/proposals/backend-organization-accounts-profile-tickets.md`.

### [OAP-T04] Tích hợp 12 API danh mục với auth, CSRF và PostgreSQL

**Mục tiêu**

Cung cấp CRUD danh mục chạy thật qua HTTP và có quyền server để FE nối trực tiếp.

**Đầu vào và điều kiện bắt đầu**

Service/repository OAP-T03; OpenAPI OAP và middleware/cookie/CSRF auth đã có. Mỗi danh mục có GET/POST collection và PATCH/DELETE item.

**Phạm vi thực hiện**

1. Viết handler/DTO allowlist; wire 12 method/path trong router/composition root, giữ health và năm API auth.
2. Áp phiên thường/HR hoặc Admin-read; mutation Origin/CSRF, JSON strict/Content-Type/body 8192 byte thực; dùng error mapper/no-store.
3. Kiểm route thực và HTTP+DB, gồm request không Content-Length; đưa các operation đã triển khai vào Swagger dev từ canonical spec theo cơ chế local hiện hữu, không quảng cáo route chưa có như đã chạy.

**Các endpoint thuộc ticket**

| Danh mục | Collection | Một mục |
|---|---|---|
| Phòng ban | GET và POST `/api/departments` | PATCH và DELETE `/api/departments/{id}` |
| Vị trí | GET và POST `/api/positions` | PATCH và DELETE `/api/positions/{id}` |
| Cấp bậc | GET và POST `/api/job-levels` | PATCH và DELETE `/api/job-levels/{id}` |

Tổng 12 method/path. GET cho HR hoặc Admin; POST/PATCH/DELETE chỉ HR. Mọi route yêu cầu phiên thường, quyền của tổ hợp role được cộng theo nguồn server.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| HR CRUD; Admin-only GET | Đúng status/DTO; Admin-only mutation và employee-only truy cập bị 403. |
| Limited kể cả ADMIN/HR hoặc phiên lỗi | 403 PASSWORD_CHANGE_REQUIRED hoặc 401; lỗi DB vẫn 503/500, không giả hết phiên. |
| JSON/Origin/CSRF/body sai | 400/403/413/415 đúng contract, không ghi; biên 8192/8193 được kiểm. |
| Duplicate/stale/reference và 204 | 409 đúng mã theo thao tác; DELETE thành công không JSON; mọi response no-store. |

**Tiêu chí hoàn thành**

- [ ] 12 route hoạt động qua adapter PG thật, không endpoint nghiệp vụ giả và không nới CORS/auth.
- [ ] HTTP/DB tests và regression năm auth API đạt; spec/Swagger chỉ mô tả đúng trạng thái đã triển khai.
- [ ] Diff/nguồn đã review, góp ý chặn đã xử lý; bàn giao lệnh thực/kết quả/môi trường/giới hạn và đầu ra cho task phụ thuộc. Không đưa secret/PII vào evidence; Done sau merge/bàn giao phù hợp workflow, không nhận mock là kiểm chứng DB/mail/browser thật.

**Phụ thuộc và cách thực hiện**

- Is blocked by: OAP-T03.
- Blocks: OAP-T05.
- WBS: OAP-06. Agent triển khai; Quân đọc hiểu/thử/review; checkpoint tự viết chỉ giao khi chốt brief.
- Estimate/remaining/actual: chưa xác định; không ghi lại công sức đã có như công việc còn lại.

**Kiểm tra và bằng chứng**

Contract/httptest/route inventory; OAP-DB05/06 cùng ca bảo mật trong bảng. Swagger render không thay bằng chứng HTTP+DB.

- Tại thư mục gốc: `go vet ./...`, `go test ./... -count=1`, `go build ./...`. Unit tests không thay integration có build tag.
- Trước integration: `python3 scripts/authdb/manage.py check`; sau khi xác minh PG18 test sở hữu, chạy `python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m`. Đây là nhóm package hiện có; nếu task thêm package test mới thì bổ sung lệnh thực khi bàn giao. Không fallback DATABASE_URL/.env/Neon; không chạy nhiều lượt DB song song. Race dùng hai connection/barrier, không sleep ngẫu nhiên.
- Kiểm contract tại thư mục gốc: `node --test scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`. Đây là kiểm tĩnh, không thay HTTP/DB/runtime.
- Kiểm nguồn/diff tại thư mục gốc: `node scripts/check-harness.mjs`, `git diff --check`. Bàn giao lệnh đầy đủ đã thực chạy, kết quả, môi trường và giới hạn; không ghi trước pass hoặc đưa secret/PII vào log/evidence.

**Đầu ra bàn giao**

12 API, adapter input/output và lỗi có thể tái hiện cho OAP-T05; GET catalog dùng lại cho form Admin.

**Ngoài phạm vi**

Không nối toàn FE, endpoint account/profile hoặc mở production docs.

**Tham chiếu**

Proposal mục 6.2/6.4 F1/7; OpenAPI ba collection catalog; FR-015. Nguồn: [proposal](backend-organization-accounts-profile.md), [OpenAPI](../api/organization-accounts-profile.openapi.json), [SQL plan](backend-organization-accounts-profile-sql-plan.md), [FR/AC](../product/functional-requirements.md), [workflow](../ai/workflow.md). Đọc kiến trúc/bảo mật/frontend do proposal dẫn trước sửa; không copy toàn bộ thiết kế vào ticket.

**Truy vết nhập Jira:** `OAP-T04`; nguồn `docs/proposals/backend-organization-accounts-profile-tickets.md`.

### [OAP-T05] Nối UI danh mục tổ chức và navigation theo quyền server

**Mục tiêu**

Cho HR quản lý danh mục trên UI bằng DB thật, có navigation tối thiểu tách hẳn prototype.

**Đầu vào và điều kiện bắt đầu**

Checkpoint trước sửa UI: chốt trong brief URL danh mục, cấu trúc navigation server và hành vi khi vào `/`/URL chưa nối cho từng tổ hợp role. Phạm vi/mục menu đã duyệt, URL và đích mặc định chi tiết chưa được coi là đã chốt; không tự mở dashboard mock. Bàn giao quyết định URL cho OAP-T07/OAP-T12.

12 API OAP-T04; đọc web/AGENTS.md và web/README.md trước sửa. Tái sử dụng màn/primitives đã duyệt, không tạo dashboard hoặc workspace mới.

**Phạm vi thực hiện**

1. Tách adapter server/data callbacks khỏi demo cho màn danh mục; navigation từ principal server, tổ hợp role hiện đúng mục.
2. Nối list/search/pagination và form create/rename/delete, giữ expected_before_name từ dữ liệu đã đọc; không tự replay mutation.
3. Hiển thị loading/validation/stale/reference/phiên/DB/mạng; 403 cập nhật quyền theo me khi phù hợp, không fallback demo.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| HR CRUD rồi reload | Dữ liệu giữ ở DB; đổi tên giữ ID, xóa bị chặn hiển thị rõ tham chiếu. |
| Màn cũ gửi rename/delete | Hiện stale và cho tải lại, không tự ghi đè. |
| Admin-only/employee/limited deep link | Không dùng URL/client role để vượt quyền; màn nghiệp vụ chưa nối không mở dữ liệu mock. |
| 503 hoặc mất phản hồi POST | Giữ thông báo đúng, tải lại để đối soát; không giả success hoặc tự POST lại. |

**Tiêu chí hoàn thành**

- [ ] Browser với backend test chứng minh CRUD/reload/stale/xóa bị chặn và tách server/demo.
- [ ] Brief ghi mapping URL/navigation và đích tạm ở phạm vi đã nối; T07 tiếp nhận việc khép đích sau login với hồ sơ, T12 bổ sung tài khoản.
- [ ] Adapter tests, pnpm check và kiểm responsive/focus theo convention đạt; navigation dùng tiếp được cho hồ sơ/tài khoản.
- [ ] Diff/nguồn đã review, góp ý chặn đã xử lý; bàn giao lệnh thực/kết quả/môi trường/giới hạn và đầu ra cho task phụ thuộc. Không đưa secret/PII vào evidence; Done sau merge/bàn giao phù hợp workflow, không nhận mock là kiểm chứng DB/mail/browser thật.

**Phụ thuộc và cách thực hiện**

- Is blocked by: OAP-T04.
- Blocks: OAP-T07, OAP-T12.
- WBS: OAP-07. Agent triển khai; Quân đọc hiểu/thử/review; checkpoint tự viết chỉ giao khi chốt brief.
- Estimate/remaining/actual: chưa xác định; không ghi lại công sức đã có như công việc còn lại.

**Kiểm tra và bằng chứng**

Browser/DB thật từ OAP-T04; không dùng HTTP fake nhận phần persistence pass.

- Tại `web/`: `pnpm check` (gồm test adapter và các cổng FE hiện hành). Kiểm thêm browser với backend PostgreSQL test; ghi bước, dữ liệu tổng hợp, kết quả và ca lỗi chưa kiểm, không lấy HTTP giả thay browser/DB.
- Kiểm nguồn/diff tại thư mục gốc: `node scripts/check-harness.mjs`, `git diff --check`. Bàn giao lệnh đầy đủ đã thực chạy, kết quả, môi trường và giới hạn; không ghi trước pass hoặc đưa secret/PII vào log/evidence.

**Đầu ra bàn giao**

Navigation server và UI danh mục đã nối cho OAP-T07/OAP-T12; kịch bản demo F1.

**Ngoài phạm vi**

Không profile HR/edit/approve, mock workspace hoặc thay toàn bộ thiết kế FE.

**Tham chiếu**

Proposal mục 6.6/7; OpenAPI catalog; FR-015; web README. Nguồn: [proposal](backend-organization-accounts-profile.md), [OpenAPI](../api/organization-accounts-profile.openapi.json), [SQL plan](backend-organization-accounts-profile-sql-plan.md), [FR/AC](../product/functional-requirements.md), [workflow](../ai/workflow.md). Đọc kiến trúc/bảo mật/frontend do proposal dẫn trước sửa; không copy toàn bộ thiết kế vào ticket.

**Truy vết nhập Jira:** `OAP-T05`; nguồn `docs/proposals/backend-organization-accounts-profile-tickets.md`.

### [OAP-T06] Triển khai API đọc hồ sơ chính chủ với projection tối thiểu

**Mục tiêu**

Đọc hồ sơ hành chính/công tác của đúng user hiện hành, không cấp quyền đọc PII của người khác.

**Đầu vào và điều kiện bắt đầu**

Schema/fixture M1 OAP-T02 và auth principal; route GET /api/employees/{id} phải phân biệt user_id/employee_id.

**Phạm vi thực hiện**

1. Viết projection/service ReadSelf nối position bắt buộc, department/job_level nullable; company_email projection canonical nhưng không UPDATE legacy.
2. Nối handler/router với auth và cổng limited; recheck trạng thái/cờ hiện hành đúng phân loại 401/403/404; DTO allowlist không audit/CV/history/credential.
3. Kiểm HTTP+PG thật và bổ sung operation đã chạy vào Swagger dev theo cùng cơ chế canonical contract.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Self có user_id khác employee_id | 200 đúng liên kết và danh mục; address/phone/department/job_level NULL giữ đúng nghĩa. |
| ID người khác/missing kể cả Admin | Cùng 404 sau xác thực; không lộ dữ liệu người khác. |
| Actor inactive/cờ đổi/DB lỗi | 401/403 hoặc 503/500 đúng nguồn, không gộp hết thành 404. |
| Danh mục đổi tên/email legacy có padding | Tên hiện hành theo ID; email response đúng schema, DB email gốc giữ nguyên. |

**Tiêu chí hoàn thành**

- [ ] DB+HTTP tests chứng minh quyền self, không nhân row từ roles và không dùng client principal làm nguồn quyền cuối.
- [ ] Response/no-store khớp OpenAPI; regression auth đạt; lịch sử đề nghị được ghi ngoài phạm vi.
- [ ] Diff/nguồn đã review, góp ý chặn đã xử lý; bàn giao lệnh thực/kết quả/môi trường/giới hạn và đầu ra cho task phụ thuộc. Không đưa secret/PII vào evidence; Done sau merge/bàn giao phù hợp workflow, không nhận mock là kiểm chứng DB/mail/browser thật.

**Phụ thuộc và cách thực hiện**

- Is blocked by: OAP-T02.
- Blocks: OAP-T07.
- WBS: OAP-08. Agent triển khai; Quân đọc hiểu/thử/review; checkpoint tự viết chỉ giao khi chốt brief.
- Estimate/remaining/actual: chưa xác định; không ghi lại công sức đã có như công việc còn lại.

**Kiểm tra và bằng chứng**

Httptest/route inventory; OAP-DB11, normalization projection và PII allowlist.

- Tại thư mục gốc: `go vet ./...`, `go test ./... -count=1`, `go build ./...`. Unit tests không thay integration có build tag.
- Trước integration: `python3 scripts/authdb/manage.py check`; sau khi xác minh PG18 test sở hữu, chạy `python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m`. Đây là nhóm package hiện có; nếu task thêm package test mới thì bổ sung lệnh thực khi bàn giao. Không fallback DATABASE_URL/.env/Neon; không chạy nhiều lượt DB song song. Race dùng hai connection/barrier, không sleep ngẫu nhiên.
- Kiểm contract tại thư mục gốc: `node --test scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`. Đây là kiểm tĩnh, không thay HTTP/DB/runtime.
- Kiểm nguồn/diff tại thư mục gốc: `node scripts/check-harness.mjs`, `git diff --check`. Bàn giao lệnh đầy đủ đã thực chạy, kết quả, môi trường và giới hạn; không ghi trước pass hoặc đưa secret/PII vào log/evidence.

**Đầu ra bàn giao**

API self/profile adapter contract và fixture cho OAP-T07.

**Ngoài phạm vi**

Không profile edit/HR read-all/duyệt đề nghị, không mở rộng /auth/me.

**Tham chiếu**

Proposal mục 6.2–6.4 F3/6.6; SQL plan mục 5; OpenAPI self; FR-009.AC1 phần dữ liệu chính thức, AC2 và BR-001. Nguồn: [proposal](backend-organization-accounts-profile.md), [OpenAPI](../api/organization-accounts-profile.openapi.json), [SQL plan](backend-organization-accounts-profile-sql-plan.md), [FR/AC](../product/functional-requirements.md), [workflow](../ai/workflow.md). Đọc kiến trúc/bảo mật/frontend do proposal dẫn trước sửa; không copy toàn bộ thiết kế vào ticket.

**Truy vết nhập Jira:** `OAP-T06`; nguồn `docs/proposals/backend-organization-accounts-profile-tickets.md`.

### [OAP-T07] Nối màn hồ sơ cá nhân với API thật và điều hướng sau đăng nhập

**Mục tiêu**

Thay màn tạm “Đã đăng nhập” bằng vùng server đã nối; nhân viên xem hồ sơ thật sau login/đổi mật khẩu.

**Đầu vào và điều kiện bắt đầu**

API OAP-T06 và navigation OAP-T05; auth bootstrap/change-password hiện hành phải được giữ. Đọc hướng dẫn frontend trước sửa. Trước code, chốt brief tiếp nối URL hồ sơ và mapping đích sau login, reload/deep link, URL chưa nối cho nhân viên/HR/Admin/tổ hợp role; tái sử dụng mapping từ OAP-T05. Đây là checkpoint kỹ thuật chưa hoàn tất, không phải quyết định đã có từ ticket này.

**Phạm vi thực hiện**

1. Nối màn PersonalProfile qua adapter server, lấy employee_id từ principal rồi tải API self; không ghép người thật với employee fixture.
2. Định tuyến login/reload/deep link tới vùng server/hồ sơ theo mapping URL được chốt tại checkpoint đầu task; limited luôn về cổng đổi mật khẩu, kể cả role Admin.
3. Hiển thị address chưa có theo BR-001; lịch sử đề nghị ghi chưa kết nối, không giả [] là đã xác minh không có; ẩn action chưa hỗ trợ.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Login/reload/profile deep link | me/profile đọc từ server, hiển thị đúng công tác và dữ liệu NULL. |
| Limited/back/logout hoặc đổi ID | Không mở nghiệp vụ bằng cờ client/cookie đã revoke; không đọc người khác. |
| Profile 401/403/503/mạng | Phân biệt mất phiên/mất quyền/lỗi dịch vụ; không fallback mock hoặc giữ PII user trước. |

**Tiêu chí hoàn thành**

- [ ] Có mapping URL/đích sau login và hành vi reload/deep link được xác nhận trong brief; browser với backend test chứng minh mapping đó cùng self profile/logout/limited gate.
- [ ] Adapter tests/pnpm check đạt; không mở form cập nhật/lịch sử giả và không tuyên bố hoàn thành toàn FR-009.
- [ ] Diff/nguồn đã review, góp ý chặn đã xử lý; bàn giao lệnh thực/kết quả/môi trường/giới hạn và đầu ra cho task phụ thuộc. Không đưa secret/PII vào evidence; Done sau merge/bàn giao phù hợp workflow, không nhận mock là kiểm chứng DB/mail/browser thật.

**Phụ thuộc và cách thực hiện**

- Is blocked by: OAP-T05, OAP-T06.
- Blocks: OAP-T13.
- WBS: OAP-09. Agent triển khai; Quân đọc hiểu/thử/review; checkpoint tự viết chỉ giao khi chốt brief.
- Estimate/remaining/actual: chưa xác định; không ghi lại công sức đã có như công việc còn lại.

**Kiểm tra và bằng chứng**

Browser backend thật; regression auth FE thường/hạn chế, kiểm server/demo và clear dữ liệu khi đổi phiên.

- Tại `web/`: `pnpm check` (gồm test adapter và các cổng FE hiện hành). Kiểm thêm browser với backend PostgreSQL test; ghi bước, dữ liệu tổng hợp, kết quả và ca lỗi chưa kiểm, không lấy HTTP giả thay browser/DB.
- Kiểm nguồn/diff tại thư mục gốc: `node scripts/check-harness.mjs`, `git diff --check`. Bàn giao lệnh đầy đủ đã thực chạy, kết quả, môi trường và giới hạn; không ghi trước pass hoặc đưa secret/PII vào log/evidence.

**Đầu ra bàn giao**

Luồng F3 dùng tài khoản local tổng hợp trước F2; kịch bản kiểm lại bằng tài khoản tạo thật ở OAP-T13.

**Ngoài phạm vi**

Không dashboard nghiệp vụ mock, đề nghị/duyệt hồ sơ, CV hoặc API nghiệp vụ khác.

**Tham chiếu**

Proposal mục 6.4 F3/6.6/7; OpenAPI self; FR-009 phần đọc chính thức, BR-001; web README. Nguồn: [proposal](backend-organization-accounts-profile.md), [OpenAPI](../api/organization-accounts-profile.openapi.json), [SQL plan](backend-organization-accounts-profile-sql-plan.md), [FR/AC](../product/functional-requirements.md), [workflow](../ai/workflow.md). Đọc kiến trúc/bảo mật/frontend do proposal dẫn trước sửa; không copy toàn bộ thiết kế vào ticket.

**Truy vết nhập Jira:** `OAP-T07`; nguồn `docs/proposals/backend-organization-accounts-profile-tickets.md`.

### [OAP-T08] Bỏ username và chuẩn bị sequence, email unique cho tạo tài khoản

**Mục tiêu**

Nâng schema M2 và toàn bộ consumer chạy schema đích, không mất dữ liệu ngoài cột username đã quyết định bỏ.

**Đầu vào và điều kiện bắt đầu**

M1 OAP-T02, kế hoạch M2 OAP-T01; kiểm lại head/metadata. SQL cụ thể, compatibility legacy role và dữ liệu khởi tạo sequence phải review trước thực thi trên DB test.

**Phạm vi thực hiện**

1. Preflight email/code/dependency; migration tiến thêm DROP username không CASCADE, sequence BIGINT/unique email canonical theo SQL plan. Giữ users.system_role tương thích EMPLOYEE cho user mới, không là nguồn quyền.
2. Cập nhật seed/localdev/fixture/runner/consumer đích; giữ fixture dựng v1/v2/v3 đúng lịch sử, không sửa 000001–000003.
3. Kiểm nâng rỗng/legacy, sequence/code dài/collision/overflow/preview, normalization Go–PG; diễn tập down từ chối dữ liệu/sequence đã dùng và dirty thật.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Email trùng/malformed hoặc normalize lệch PG | Dừng để review; không tự đổi legacy hoặc tạo uniqueness không khớp auth. |
| EMP001/EMP000001/E001 và rollback sau nextval | Giữ mã cũ; khởi tạo đúng dải EMP, không reuse/reset/truncate số dài. |
| Schema head không username | Seed/localdev/auth chạy, tạo không alias; ID/link/hash/email/status/roles/sessions giữ nguyên. |
| Down có dữ liệu hoặc từng cấp số | Từ chối atomically theo plan; không bịa username hoặc reset high-water mark. |

**Tiêu chí hoàn thành**

- [ ] Migration M2/consumer đích được kiểm bằng hai đường nâng cấp và regression năm API auth trên PG test.
- [ ] Có bằng chứng constraint/email parity/sequence/rollback, cách chạy-dọn và giới hạn backup; không gọi down là khôi phục username.
- [ ] Diff/nguồn đã review, góp ý chặn đã xử lý; bàn giao lệnh thực/kết quả/môi trường/giới hạn và đầu ra cho task phụ thuộc. Không đưa secret/PII vào evidence; Done sau merge/bàn giao phù hợp workflow, không nhận mock là kiểm chứng DB/mail/browser thật.

**Phụ thuộc và cách thực hiện**

- Is blocked by: OAP-T01, OAP-T02.
- Blocks: OAP-T09.
- WBS: OAP-10. Agent triển khai; Quân đọc hiểu/thử/review; checkpoint tự viết chỉ giao khi chốt brief.
- Estimate/remaining/actual: chưa xác định; không ghi lại công sức đã có như công việc còn lại.

**Kiểm tra và bằng chứng**

OAP-DB01/02/03/07/08/12/13/14 phần M2. Chạy seed/test runner đúng schema đích; không dùng fixture giả làm bằng chứng cấp mã đồng thời từ create service.

- Tại thư mục gốc: `go vet ./...`, `go test ./... -count=1`, `go build ./...`. Unit tests không thay integration có build tag.
- Trước integration: `python3 scripts/authdb/manage.py check`; sau khi xác minh PG18 test sở hữu, chạy `python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./migrations ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m`. Đây là nhóm package hiện có; nếu task thêm package test mới thì bổ sung lệnh thực khi bàn giao. Không fallback DATABASE_URL/.env/Neon; không chạy nhiều lượt DB song song. Race dùng hai connection/barrier, không sleep ngẫu nhiên.
- Kiểm nguồn/diff tại thư mục gốc: `node scripts/check-harness.mjs`, `git diff --check`. Bàn giao lệnh đầy đủ đã thực chạy, kết quả, môi trường và giới hạn; không ghi trước pass hoặc đưa secret/PII vào log/evidence.

**Đầu ra bàn giao**

Schema head không username, allocator/fixture và quy tắc seed cho OAP-T09.

**Ngoài phạm vi**

Không áp Neon, chỉnh tài khoản legacy, thêm alias/delivery status/reset hoặc sửa migration lịch sử.

**Tham chiếu**

Proposal mục 4.1/6.1; SQL plan mục 2/4/6–7; FR-005.AC3, FR-006. Nguồn: [proposal](backend-organization-accounts-profile.md), [OpenAPI](../api/organization-accounts-profile.openapi.json), [SQL plan](backend-organization-accounts-profile-sql-plan.md), [FR/AC](../product/functional-requirements.md), [workflow](../ai/workflow.md). Đọc kiến trúc/bảo mật/frontend do proposal dẫn trước sửa; không copy toàn bộ thiết kế vào ticket.

**Truy vết nhập Jira:** `OAP-T08`; nguồn `docs/proposals/backend-organization-accounts-profile-tickets.md`.

### [OAP-T09] Triển khai transaction tạo user, employee, quyền và audit nguyên tử

**Mục tiêu**

Tạo tài khoản kèm hồ sơ và mã duy nhất sau recheck Admin; chỉ trả kết quả tạo khi commit được xác nhận.

**Đầu vào và điều kiện bắt đầu**

Schema/consumer M2 OAP-T08, fixture và ports auth/actor; thiết kế secret/roles/audit và thứ tự khóa đã chốt. Mail port chưa cần cấu hình thật.

**Phạm vi thực hiện**

1. Viết use case/port create: validate, sinh secret 32 byte và bcrypt ngoài transaction; secret chỉ trong luồng nội bộ cần gửi mail, không DTO/log.
2. Adapter recheck actor/ADMIN/cờ/status; khóa catalog đúng thứ tự, cấp nextval và ghi user/employee/user_roles/audit cùng transaction.
3. Kiểm hai connection/barrier cho create đồng thời/email collision/delete catalog/actor mất quyền; fault injection write/audit/commit ambiguity và bàn giao kết quả cho orchestration mail.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Dữ liệu hợp lệ, roles [] hoặc tổ hợp hợp lệ | User/employee/roles/audit commit cùng nhau; ACTIVE/true, không session mới. |
| Lỗi FK/email/write/audit hoặc random/bcrypt | Không half-write, không mail; sequence có thể mất số theo plan. |
| N create/race actor hoặc catalog | Mã duy nhất, quyền/FK hiện hành; không ghi dựa snapshot cũ. |
| Commit lỗi/mất xác nhận | Không coi chắc chắn chưa ghi hoặc thành công; không retry mutation/gửi mail theo phỏng đoán. |

**Tiêu chí hoàn thành**

- [ ] DB tests chứng minh atomicity/race/recheck/sequence và dữ liệu liên kết đúng; secret/hash không qua transport/log/evidence.
- [ ] Service không import pgx/HTTP; kết quả/lỗi phân loại đủ để OAP-T11 quyết định gửi thư sau commit và response đúng.
- [ ] Diff/nguồn đã review, góp ý chặn đã xử lý; bàn giao lệnh thực/kết quả/môi trường/giới hạn và đầu ra cho task phụ thuộc. Không đưa secret/PII vào evidence; Done sau merge/bàn giao phù hợp workflow, không nhận mock là kiểm chứng DB/mail/browser thật.

**Phụ thuộc và cách thực hiện**

- Is blocked by: OAP-T08.
- Blocks: OAP-T11.
- WBS: OAP-11. Agent triển khai; Quân đọc hiểu/thử/review; checkpoint tự viết chỉ giao khi chốt brief.
- Estimate/remaining/actual: chưa xác định; không ghi lại công sức đã có như công việc còn lại.

**Kiểm tra và bằng chứng**

Unit random/bcrypt thật hoặc lỗi có kiểm soát; OAP-DB06/07/09/10. Mail fake chỉ chứng minh không gọi trước commit khi ghép OAP-T11.

- Tại thư mục gốc: `go vet ./...`, `go test ./... -count=1`, `go build ./...`. Unit tests không thay integration có build tag.
- Trước integration: `python3 scripts/authdb/manage.py check`; sau khi xác minh PG18 test sở hữu, chạy `python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./internal/repository/postgres -count=1 -timeout=5m`. Đây là nhóm package hiện có; nếu task thêm package test mới thì bổ sung lệnh thực khi bàn giao. Không fallback DATABASE_URL/.env/Neon; không chạy nhiều lượt DB song song. Race dùng hai connection/barrier, không sleep ngẫu nhiên.
- Kiểm nguồn/diff tại thư mục gốc: `node scripts/check-harness.mjs`, `git diff --check`. Bàn giao lệnh đầy đủ đã thực chạy, kết quả, môi trường và giới hạn; không ghi trước pass hoặc đưa secret/PII vào log/evidence.

**Đầu ra bàn giao**

Create use case/repository, typed outcome và fixture/race cho OAP-T11.

**Ngoài phạm vi**

Không endpoint/FE/Gmail thực, session tự đăng nhập, reset hoặc thay quyền tài khoản có sẵn.

**Tham chiếu**

Proposal mục 5/6.3/6.4 F2/7; SQL plan mục 5/7; FR-005.AC1–AC3, FR-006; role ban đầu FR-064.AC1 phần tạo mới. Nguồn: [proposal](backend-organization-accounts-profile.md), [OpenAPI](../api/organization-accounts-profile.openapi.json), [SQL plan](backend-organization-accounts-profile-sql-plan.md), [FR/AC](../product/functional-requirements.md), [workflow](../ai/workflow.md). Đọc kiến trúc/bảo mật/frontend do proposal dẫn trước sửa; không copy toàn bộ thiết kế vào ticket.

**Truy vết nhập Jira:** `OAP-T09`; nguồn `docs/proposals/backend-organization-accounts-profile-tickets.md`.

### [OAP-T10] Tích hợp Gmail OAuth và adapter gửi thông tin tài khoản

**Mục tiêu**

Cung cấp MailSender có cấu hình an toàn và kết quả accepted/failed/unknown được kiểm chứng cho luồng cấp tài khoản.

**Đầu vào và điều kiện bắt đầu**

Gmail API/OAuth2/gmail.send đã chọn; chưa có cấu hình Google thực tế. Quân kiểm soát hộp thư gửi/nhận. Trước dependency/cấu hình, brief phải nêu OAuth client/redirect URI, thư viện, secret store và timeout cụ thể để xác nhận.

**Phạm vi thực hiện**

1. Chốt brief kỹ thuật rồi hướng dẫn chủ hộp thư cấu hình Cloud/OAuth/offline access; credential nhập qua kênh local an toàn, không chat/Git/FE. Không tự mua dịch vụ/bật billing hoặc xin scope rộng.
2. Hiện thực adapter OAuth/token refresh và messages.send qua MailSender; template tiếng Việt hướng dẫn email đăng nhập/mật khẩu tạm/buộc đổi, không username hoặc nội dung ngoài phạm vi.
3. Phân loại lỗi trước gửi/gửi không rõ/Google nhận gửi; không retry mù messages.send; test refresh/hết hạn/thu hồi/timeout/redaction và runbook kiểm kết nối trước demo.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Google nhận yêu cầu gửi | accepted, chưa gọi delivered; thư thật được kiểm ở mailbox được phép mà không chụp/log credential. |
| Cấu hình/token lỗi trước gửi | failed, không đưa provider raw error/token vào response/log. |
| Timeout/mất phản hồi sau gửi | unknown; không tự gửi lại, không sinh mật khẩu khác. |
| Refresh token Testing hết hạn/thu hồi | Hướng dẫn chủ hộp thư cấp quyền lại; không coi reconnect là gửi lại thư. |

**Tiêu chí hoàn thành**

- [ ] Brief/dependency được review, adapter tests/Go checks đạt; secret không nằm trong source/session table/FE.
- [ ] Có runbook và bằng chứng gửi thật tới mailbox được phép; nếu chưa cấp quyền/chưa có môi trường thì task còn blocked, không lấy fake thay mail thật.
- [ ] Diff/nguồn đã review, góp ý chặn đã xử lý; bàn giao lệnh thực/kết quả/môi trường/giới hạn và đầu ra cho task phụ thuộc. Không đưa secret/PII vào evidence; Done sau merge/bàn giao phù hợp workflow, không nhận mock là kiểm chứng DB/mail/browser thật.

**Phụ thuộc và cách thực hiện**

- Is blocked by: OAP-T01.
- Blocks: OAP-T11.
- WBS: OAP-12. Agent triển khai; Quân đọc hiểu/thử/review; checkpoint tự viết chỉ giao khi chốt brief.
- Estimate/remaining/actual: chưa xác định; không ghi lại công sức đã có như công việc còn lại.

**Kiểm tra và bằng chứng**

HTTP giả provider + kiểm Gmail thật sau khi được cho phép. Fake kiểm mapping/refresh/redaction; inbox test chỉ ghi metadata an toàn/kết quả, không mail body hoặc mật khẩu.

- Tại thư mục gốc: `go vet ./...`, `go test ./... -count=1`, `go build ./...`. Unit tests không thay integration có build tag.
- Kiểm nguồn/diff tại thư mục gốc: `node scripts/check-harness.mjs`, `git diff --check`. Bàn giao lệnh đầy đủ đã thực chạy, kết quả, môi trường và giới hạn; không ghi trước pass hoặc đưa secret/PII vào log/evidence.

**Đầu ra bàn giao**

MailSender/config/template/runbook và lỗi phân loại cho OAP-T11; hạn Testing 7 ngày được kiểm trước buổi demo.

**Ngoài phạm vi**

Không login Google cho nhân viên, SMTP, scope đọc mail, billing, hàng đợi/outbox/delivery storage hoặc endpoint cấp lại/gửi lại.

**Tham chiếu**

Proposal mục 6.5/7; OpenAPI MailResult; FR-005.AC4; security mục Gửi thư cấp tài khoản. Nguồn: [proposal](backend-organization-accounts-profile.md), [OpenAPI](../api/organization-accounts-profile.openapi.json), [SQL plan](backend-organization-accounts-profile-sql-plan.md), [FR/AC](../product/functional-requirements.md), [workflow](../ai/workflow.md). Đọc kiến trúc/bảo mật/frontend do proposal dẫn trước sửa; không copy toàn bộ thiết kế vào ticket.

**Truy vết nhập Jira:** `OAP-T10`; nguồn `docs/proposals/backend-organization-accounts-profile-tickets.md`.

### [OAP-T11] Tích hợp API tạo tài khoản, mã dự kiến và danh sách tài khoản

**Mục tiêu**

Cung cấp ba API Admin chạy thật, tách kết quả commit tài khoản với kết quả Gmail và đối soát được sau lỗi.

**Đầu vào và điều kiện bắt đầu**

Create transaction OAP-T09 và MailSender OAP-T10. Catalog GET từ OAP-T04 cần sẵn sàng trước nghiệm thu UI; không query DB trong handler.

**Phạm vi thực hiện**

1. Nối POST /api/employees, GET /api/employees/next-code và GET /api/users; viết preview/list projection, pagination/search/snapshot và DTO theo OpenAPI.
2. Orchestration gọi mail đúng một lần có chủ ý sau commit đã xác nhận; response 201 account_created/account/mail accepted|failed|unknown, không cookie/password; commit ambiguity không mail/auto retry.
3. Áp auth/business gate/ADMIN/Origin/CSRF/JSON8KiB, error mapping/no-store và wire composition root/config/Swagger dev cho các route đã thực thi.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Create commit rồi mail accepted/failed/unknown | Đều 201 account_created=true, phân biệt trạng thái thư; không rollback tài khoản vì mail lỗi. |
| Input/role/phiên/FK/email/body sai | 400/401/403/409/413/415 đúng nguồn, không half-write hoặc gửi mail. |
| Preview đồng thời create hoặc restart | is_preview=true không giữ chỗ/tiêu số; mã commit mới chính thức, không reset sequence. |
| Mất response/commit không rõ | Không success giả hoặc retry; list đối soát account, không suy kết quả gửi từ row tồn tại. |
| Admin list/search | Chỉ projection cho phép, roles hiện hành; không phone/address/secret/delivery status. |

**Tiêu chí hoàn thành**

- [ ] HTTP+PG thật chứng minh ba API, thứ tự commit→mail và lỗi từng tầng; có kiểm body thực/DTO/no-store/không Set-Cookie.
- [ ] Unit mail fault tests không thay bằng chứng PG; runbook cấu hình đủ chạy, regression auth và các route catalog/profile đã merge đạt. Swagger dev chỉ mở các operation đã triển khai từ nguồn canonical, không buộc hoàn thành toàn 16 API tại task này; kiểm đủ toàn phạm vi tại OAP-T13.
- [ ] Diff/nguồn đã review, góp ý chặn đã xử lý; bàn giao lệnh thực/kết quả/môi trường/giới hạn và đầu ra cho task phụ thuộc. Không đưa secret/PII vào evidence; Done sau merge/bàn giao phù hợp workflow, không nhận mock là kiểm chứng DB/mail/browser thật.

**Phụ thuộc và cách thực hiện**

- Is blocked by: OAP-T09, OAP-T10.
- Blocks: OAP-T12.
- WBS: OAP-13. Agent triển khai; Quân đọc hiểu/thử/review; checkpoint tự viết chỉ giao khi chốt brief.
- Estimate/remaining/actual: chưa xác định; không ghi lại công sức đã có như công việc còn lại.

**Kiểm tra và bằng chứng**

Contract/HTTP; OAP-DB09/13, snapshot list/preview và fake mail kiểm thứ tự/lỗi; cấu hình Gmail thật reuse bằng chứng OAP-T10, không spam lại mailbox.

- Tại thư mục gốc: `go vet ./...`, `go test ./... -count=1`, `go build ./...`. Unit tests không thay integration có build tag.
- Trước integration: `python3 scripts/authdb/manage.py check`; sau khi xác minh PG18 test sở hữu, chạy `python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m`. Đây là nhóm package hiện có; nếu task thêm package test mới thì bổ sung lệnh thực khi bàn giao. Không fallback DATABASE_URL/.env/Neon; không chạy nhiều lượt DB song song. Race dùng hai connection/barrier, không sleep ngẫu nhiên.
- Kiểm contract tại thư mục gốc: `node --test scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`. Đây là kiểm tĩnh, không thay HTTP/DB/runtime.
- Kiểm nguồn/diff tại thư mục gốc: `node scripts/check-harness.mjs`, `git diff --check`. Bàn giao lệnh đầy đủ đã thực chạy, kết quả, môi trường và giới hạn; không ghi trước pass hoặc đưa secret/PII vào log/evidence.

**Đầu ra bàn giao**

Ba API/config/outcome đối soát và lỗi có thể tái hiện cho OAP-T12.

**Ngoài phạm vi**

Không account reset/retry mail/role edit/lock, Idempotency-Key, delivery table hoặc Admin đọc full profile.

**Tham chiếu**

Proposal mục 6.2–6.5/7; SQL plan mục 4–5/7; OpenAPI employees/next-code/users; FR-005/006/007. Nguồn: [proposal](backend-organization-accounts-profile.md), [OpenAPI](../api/organization-accounts-profile.openapi.json), [SQL plan](backend-organization-accounts-profile-sql-plan.md), [FR/AC](../product/functional-requirements.md), [workflow](../ai/workflow.md). Đọc kiến trúc/bảo mật/frontend do proposal dẫn trước sửa; không copy toàn bộ thiết kế vào ticket.

**Truy vết nhập Jira:** `OAP-T11`; nguồn `docs/proposals/backend-organization-accounts-profile-tickets.md`.

### [OAP-T12] Nối UI Admin tạo và tra cứu tài khoản với kết quả gửi email

**Mục tiêu**

Admin tạo user–employee qua UI thật và hiểu rõ tài khoản đã lưu, mã thực và trạng thái gửi thư.

**Đầu vào và điều kiện bắt đầu**

API OAP-T11, navigation/catalog adapter OAP-T05/OAP-T04; dùng lại AccountCreateForm/AccountManagement và visual convention hiện có.

**Phạm vi thực hiện**

1. Tách demo/server, bỏ username khỏi form/product model đích; tải danh mục chọn position bắt buộc và optional department/job_level, roles ban đầu đúng nguồn.
2. Nối preview readonly/create/list/search/pagination; mã sau commit thay preview; không gửi actor/status/code/password từ client.
3. Hiển thị riêng accepted/failed/unknown và mất phản hồi; không báo delivered, không lưu credential/token, không tự retry/nút gửi lại; reload chỉ hiển thị account còn kết quả mail cũ không có thông tin.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Create thành công và mail accepted | Hiển thị tài khoản/mã thật và Google nhận gửi, không lộ mật khẩu hoặc tự login user mới. |
| Mail failed/unknown sau commit | Thông báo tài khoản đã tạo; không mời tạo lại/gửi lại hoặc tuyên bố chưa gửi chắc chắn. |
| Email conflict/validation/503/mạng | Lỗi đúng trường/hành vi; mất response cho đối soát list, không POST replay. |
| Admin mất quyền/limited hoặc server mode | Guard server quyết định; UI không rơi về prototype hoặc mở dữ liệu mock. |

**Tiêu chí hoàn thành**

- [ ] Adapter tests và browser backend test chứng minh create/list/preview cạnh tranh, lỗi mail và reload đúng contract; pnpm check đạt.
- [ ] Danh sách/kết quả tài khoản không trả phone/address/credential; form create vẫn nhận phone/address optional đúng contract. Trạng thái mail chỉ ở kết quả hiện tại, giới hạn không recovery được ghi rõ.
- [ ] Diff/nguồn đã review, góp ý chặn đã xử lý; bàn giao lệnh thực/kết quả/môi trường/giới hạn và đầu ra cho task phụ thuộc. Không đưa secret/PII vào evidence; Done sau merge/bàn giao phù hợp workflow, không nhận mock là kiểm chứng DB/mail/browser thật.

**Phụ thuộc và cách thực hiện**

- Is blocked by: OAP-T05, OAP-T11.
- Blocks: OAP-T13.
- WBS: OAP-14. Agent triển khai; Quân đọc hiểu/thử/review; checkpoint tự viết chỉ giao khi chốt brief.
- Estimate/remaining/actual: chưa xác định; không ghi lại công sức đã có như công việc còn lại.

**Kiểm tra và bằng chứng**

Browser/PG thật; mail lỗi có thể fake provider có kiểm soát, ghi rõ khác mail thật; tránh thu thập password trong screenshot/log.

- Tại `web/`: `pnpm check` (gồm test adapter và các cổng FE hiện hành). Kiểm thêm browser với backend PostgreSQL test; ghi bước, dữ liệu tổng hợp, kết quả và ca lỗi chưa kiểm, không lấy HTTP giả thay browser/DB.
- Kiểm nguồn/diff tại thư mục gốc: `node scripts/check-harness.mjs`, `git diff --check`. Bàn giao lệnh đầy đủ đã thực chạy, kết quả, môi trường và giới hạn; không ghi trước pass hoặc đưa secret/PII vào log/evidence.

**Đầu ra bàn giao**

Luồng F2 trên UI, dữ liệu demo và bước đối soát cho OAP-T13.

**Ngoài phạm vi**

Không khóa/reset/đổi role tài khoản cũ, đọc profile khác người, gửi lại mail hoặc thiết kế lại toàn FE.

**Tham chiếu**

Proposal mục 6.3/6.5–6.6/7; OpenAPI create/preview/users/catalog; FR-005/006/007; web README. Nguồn: [proposal](backend-organization-accounts-profile.md), [OpenAPI](../api/organization-accounts-profile.openapi.json), [SQL plan](backend-organization-accounts-profile-sql-plan.md), [FR/AC](../product/functional-requirements.md), [workflow](../ai/workflow.md). Đọc kiến trúc/bảo mật/frontend do proposal dẫn trước sửa; không copy toàn bộ thiết kế vào ticket.

**Truy vết nhập Jira:** `OAP-T12`; nguồn `docs/proposals/backend-organization-accounts-profile-tickets.md`.

### [OAP-T13] Kiểm xuyên luồng tổ chức, cấp tài khoản, hồ sơ và bàn giao demo

**Mục tiêu**

Bàn giao luồng UI chạy lại được, bằng chứng đúng tầng và giới hạn để báo cáo, không gắn pass cho phần chưa kiểm.

**Đầu vào và điều kiện bắt đầu**

F1/F3 từ OAP-T05/OAP-T07 và F2 từ OAP-T12; PG test/fixtures/mailbox được phép. Những ca unit/DB đã có dùng lại, không chờ ticket này mới test chức năng.

**Phạm vi thực hiện**

1. Chạy HR catalog→Admin create→nhận mail thật→login hạn chế→đổi mật khẩu→login mới→profile→reload/logout/back; người dùng tự thay credential qua browser khi công cụ yêu cầu.
2. Đối chiếu đủ 16 method/path với router/OpenAPI/Swagger dev; rà quyền HR/Admin/self/limited, race/error/commit ambiguity và redaction theo bằng chứng từng task; ghi lỗi mới liên kết task gốc/Bug, không mở rộng FR.
3. Mở rộng runbook hiện hữu với cấu hình OAuth/DB migration/rollback/cách chạy-dọn/demo; đối chiếu FR Done/Partial, đề xuất bước kế tiếp và review diff cuối.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Happy path qua UI/backend/DB/mail thật | Có bước tái hiện và kết quả từng tầng, không chỉ Swagger/mock. |
| Mất quyền/phiên, DB/mail lỗi, reload/back | Không lộ PII hoặc success giả; không auto retry create/send, không mở mock workspace. |
| Bàn giao môi trường/dữ liệu | Không credential/PII trong Git/Jira/evidence; HTTPS/Neon/deploy chưa kiểm phải ghi riêng. |

**Tiêu chí hoàn thành**

- [ ] Cổng GO/FE/contract/harness và browser bắt buộc có kết quả; lỗi chặn được xử lý/retest, ca chưa đạt ghi đúng.
- [ ] Runbook/bảng bằng chứng/giới hạn và kịch bản báo cáo UI đủ dùng lại; không nhận trọn FR-009/064 hoặc toàn app chỉ vì ba feature đã nối.
- [ ] Diff/nguồn đã review, góp ý chặn đã xử lý; bàn giao lệnh thực/kết quả/môi trường/giới hạn và đầu ra cho task phụ thuộc. Không đưa secret/PII vào evidence; Done sau merge/bàn giao phù hợp workflow, không nhận mock là kiểm chứng DB/mail/browser thật.

**Phụ thuộc và cách thực hiện**

- Is blocked by: OAP-T07, OAP-T12.
- Blocks: —.
- WBS: OAP-15. Agent triển khai; Quân đọc hiểu/thử/review; checkpoint tự viết chỉ giao khi chốt brief.
- Estimate/remaining/actual: chưa xác định; không ghi lại công sức đã có như công việc còn lại.

**Kiểm tra và bằng chứng**

Tổng hợp OAP-DB01–14 đúng phạm vi và ca bổ sung snapshot/normalization. Không cộng số unit tests làm số luồng E2E, không lấy review làm runtime proof.

- Tại thư mục gốc: `go vet ./...`, `go test ./... -count=1`, `go build ./...`. Unit tests không thay integration có build tag.
- Trước integration: `python3 scripts/authdb/manage.py check`; sau khi xác minh PG18 test sở hữu, chạy `python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./migrations ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m`. Đây là nhóm package hiện có; nếu task thêm package test mới thì bổ sung lệnh thực khi bàn giao. Không fallback DATABASE_URL/.env/Neon; không chạy nhiều lượt DB song song. Race dùng hai connection/barrier, không sleep ngẫu nhiên.
- Tại `web/`: `pnpm check` (gồm test adapter và các cổng FE hiện hành). Kiểm thêm browser với backend PostgreSQL test; ghi bước, dữ liệu tổng hợp, kết quả và ca lỗi chưa kiểm, không lấy HTTP giả thay browser/DB.
- Kiểm contract tại thư mục gốc: `node --test scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`. Đây là kiểm tĩnh, không thay HTTP/DB/runtime.
- Kiểm nguồn/diff tại thư mục gốc: `node scripts/check-harness.mjs`, `git diff --check`. Bàn giao lệnh đầy đủ đã thực chạy, kết quả, môi trường và giới hạn; không ghi trước pass hoặc đưa secret/PII vào log/evidence.

**Đầu ra bàn giao**

Kịch bản demo, ma trận FR/bằng chứng/giới hạn và đầu vào miền tiếp theo; không chốt công thức phân công.

**Ngoài phạm vi**

Không áp Neon/deploy/pentest toàn app, commit/push/merge hoặc chuyển Jira Done tự động.

**Tham chiếu**

Proposal mục 3/7–9; SQL plan mục 7; OpenAPI toàn phạm vi; FR-005/006/007/015, phần đọc FR-009; delivery plan và runbook auth. Nguồn: [proposal](backend-organization-accounts-profile.md), [OpenAPI](../api/organization-accounts-profile.openapi.json), [SQL plan](backend-organization-accounts-profile-sql-plan.md), [FR/AC](../product/functional-requirements.md), [workflow](../ai/workflow.md). Đọc kiến trúc/bảo mật/frontend do proposal dẫn trước sửa; không copy toàn bộ thiết kế vào ticket.

**Truy vết nhập Jira:** `OAP-T13`; nguồn `docs/proposals/backend-organization-accounts-profile-tickets.md`.

## 6. Điều kiện trước khi nhập Jira

1. Review bộ ticket và snapshot nguồn; chốt các chỉnh sửa cần thiết. Không cần chọn lại phạm vi/email provider/format mã đã duyệt.
2. Công cụ OAP đã chuẩn bị tại [hướng dẫn import](../../scripts/jira/README-oap.md), có plan/apply/check/diagnose và test offline; **script `scripts/jira/auth_import.py` hiện chỉ dành cho auth**, không chạy apply script đó để tạo OAP và không sửa journal auth thành OAP.
3. Plan chỉ đọc phải xác minh site mới/project DATN/board1, Kanban/features/filter/Rank, metadata type/parent/assignee/Blocks và tìm mã/label đã có để tránh trùng. Board/custom-field/project IDs lấy từ metadata, không lấy site cũ. Token chỉ nhập qua prompt ẩn, không chat/argument/file Git.
4. Hiển thị 1 Epic/13 Task, 15 WBS, tập Blocks và thứ tự Rank trước ghi; nội dung OAP-T01 ghi đúng đầu ra đã có, không tự đổi trạng thái. Khi apply được giao, chỉ tạo phần thiếu đã đối soát, parent/link/Rank đúng phạm vi; không tạo lại auth hoặc tự sửa các ticket đã làm.
5. Có journal riêng gắn site/đợt và đối soát sau ghi. Mất phản hồi create/link/Rank dừng đọc lại, không retry mù hoặc hứa exactly-once. Check đọc lại đủ nội dung/type/parent/labels/link/Rank; bàn giao key thật cho Git. Ưu tiên giữ nguồn đã commit/push để link mở được.

Đợt soạn này chỉ có nội dung và kế hoạch nhập; **chưa truy cập Jira, tạo issue/link/Rank, cấu hình mail, chạy DB, commit hoặc push**.

## 7. Kết quả kiểm tra bản soạn — 15/11/2026

Đã tự review và sub-agent review read-only; sửa phạm vi nghiệm thu T11 để không tạo phụ thuộc ẩn vào toàn bộ API hồ sơ/danh mục, và làm rõ phone/address bị loại khỏi account-list/result nhưng vẫn là input create optional. Chưa còn finding chặn bước chuẩn bị import.

Đã chạy 32/32 test contract OAP/auth/harness; harness kiểm kê 311 nguồn đạt; `git diff --check` đạt. Kiểm cấu trúc bản soạn xác nhận 13 Summary khớp 13 description đủ mục, 15 WBS không trùng/thiếu, 17 cạnh Blocks đối xứng hai cột và Rank hợp thứ tự phụ thuộc, không chu trình. Đây là kiểm tài liệu/đồ thị, không chứng minh đã tạo ticket hoặc chạy API/DB/mail/browser. Chưa thay công cụ import auth hoặc trạng thái Jira.

**Cập nhật độ rõ của description 15/11:** từng Task đã có lệnh kiểm áp dụng và yêu cầu evidence trong chính description; T04 liệt kê đủ 12 method/path; T05/T07 ghi checkpoint URL/đích sau login cần chốt trước code. Không đổi phạm vi, 13 Task/15 WBS/17 Blocks hoặc tự chọn URL UI mới. Đây là nội dung nguồn cập nhật trước import; nếu đã có journal import thì fingerprint đổi phải đối soát, không xóa journal để chạy lại.

Kiểm lại sau cập nhật: 52/52 test Python importer và 32/32 test contract/harness đạt; offline plan render đủ 14 description, kiểm giữ bảng endpoint ba cột/bảng tình huống hai cột và lệnh kiểm/checkpoint trong ADF. Harness315 nguồn và diff check đạt. Các lệnh Go/DB/FE trong ticket là hướng dẫn cho implementation, chưa được chạy trong đợt sửa nội dung này; không ghi Jira hoặc commit.
