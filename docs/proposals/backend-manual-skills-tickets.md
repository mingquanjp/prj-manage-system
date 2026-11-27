# Ticket Jira — kỹ năng thủ công và Giám đốc xét duyệt

Ngày 27/11/2026. **1 Epic và 12 Task**, mã nguồn MSK. Bộ ticket triển khai luồng kỹ năng thủ công cho buổi báo cáo ngày 03/12/2026.

Nghiệp vụ và thiết kế áp dụng: [proposal](backend-manual-skills.md), [OpenAPI](../api/manual-skills.openapi.json), [SQL/test plan](backend-manual-skills-sql-plan.md), [FR](../product/functional-requirements.md) và [workflow](../ai/workflow.md).

## 1. Epic

**Summary:** Danh mục kỹ năng, đề nghị thủ công và Giám đốc xét duyệt trên UI thật

**Mục tiêu:** Nhân viên gửi đề nghị bổ sung hoặc điều chỉnh kỹ năng; Giám đốc đối chiếu, chấp thuận hoặc từ chối có lý do. Năng lực chính thức chỉ thay đổi sau chấp thuận. Hai bên nhận thông báo và tra cứu được kết quả trên UI với PostgreSQL thật.

**Phạm vi:** Giám đốc quản trị nhóm kỹ năng và kỹ năng; nhân viên xem năng lực, gửi đề nghị và xem lịch sử; Giám đốc xử lý hàng chờ. Đợt này nhận giải trình, căn cứ và ý kiến chuyên môn dạng văn bản. Giám đốc không xử lý đề nghị của chính mình; quyền HR hoặc Admin riêng không cấp quyền quản trị/xét duyệt kỹ năng.

**Tiêu chí hoàn thành:**

- [ ] Luồng danh mục → gửi đề nghị → xét duyệt → năng lực/lịch sử/thông báo chạy được trên UI, API và PostgreSQL; có kiểm quyền, dữ liệu cũ, gửi trùng và lỗi lưu.
- [ ] Có runbook chạy/dọn môi trường, kịch bản demo và bảng FR đã đạt/còn thiếu. FR-024/026 chỉ nghiệm thu phần văn bản trong đợt này; CV, tệp, căn cứ task, bổ sung/rút đề nghị thuộc đợt sau.

## 2. Rank, WBS và dependency

Rank theo thứ tự nhận việc từ trên xuống khi đủ đầu vào; không thay Blocks. Một cặp chỉ một link, không nhập đảo hai lần. Review/test/sửa lỗi nằm trong mỗi task.

| Rank | Mã nguồn | WBS | Is blocked by |
|---|---|---|---|
| 1 | MSK-T01 | MSK-01 | — |
| 2 | MSK-T02 | MSK-02 | MSK-T01 |
| 3 | MSK-T03 | MSK-03 | MSK-T02 |
| 4 | MSK-T04 | MSK-04 | MSK-T03 |
| 5 | MSK-T05 | MSK-05 | MSK-T02 |
| 6 | MSK-T06 | MSK-06 | MSK-T03, MSK-T05 |
| 7 | MSK-T07 | MSK-07 | MSK-T06 |
| 8 | MSK-T08 | MSK-08 | MSK-T06, MSK-T07 |
| 9 | MSK-T09 | MSK-09 | MSK-T08 |
| 10 | MSK-T10 | MSK-10 | MSK-T04, MSK-T08 |
| 11 | MSK-T11 | MSK-11 | MSK-T05, MSK-T09, MSK-T10 |
| 12 | MSK-T12 | MSK-12 | MSK-T04, MSK-T09, MSK-T10, MSK-T11 |

## 3. Quy ước kiểm và bàn giao

Description giữ mục tiêu, phạm vi và tiêu chí nghiệm thu. Tiến độ, lệnh đã chạy, kết quả kiểm và nhận xét review được cập nhật trong comment Jira hoặc runbook liên quan theo [workflow](../ai/workflow.md). Mỗi bàn giao ghi rõ file, môi trường, lệnh/kết quả, ca chưa kiểm và task tiếp nhận; chỉ dùng dữ liệu tổng hợp, không lưu credential hoặc thông tin riêng tư thật.

Kiểm tĩnh xác nhận cấu trúc contract/nguồn; unit và adapter tests kiểm logic riêng; integration kiểm HTTP/transaction trên PostgreSQL; browser kiểm luồng UI với backend thật. Các tầng này được ghi riêng trong bằng chứng. DB tests dùng runner sở hữu, chạy tuần tự; race dùng hai connection và barrier.

Khi triển khai áp dụng AGENTS và proposal mục 5: dùng thư viện/công cụ hiện có, hàm có trách nhiệm rõ, tách dữ liệu server/demo và giữ kiểm quyền/dữ liệu ở đúng ranh giới. Review và sửa lỗi nằm trong từng task. Quân kiểm chứng đầu ra; checkpoint Go tự viết được xác định trong brief nhận việc.

Thông tin import: 1 Epic/12 Task, 19 Blocks; nhãn `manual-skills`, `datn-msk-v1` và `source-msk-*`. Mã DATN/assignee được xác minh khi import. Estimate/remaining/actual chưa xác định; importer không cập nhật thời gian, trạng thái, priority, sprint hoặc deadline.

## 4. Description từng task

### [MSK-T01] Bàn giao contract, SQL plan và brief kỹ năng thủ công

**Mục tiêu**

Bàn giao bộ thiết kế thống nhất để các task migration, API và UI bắt đầu từ đúng contract đã duyệt.

**Đầu vào và điều kiện bắt đầu**

Proposal MSK-A01–A08, OpenAPI 14 thao tác và SQL/test plan hiện có; URL và phạm vi văn bản được xác nhận ngày 27/11/2026.

**Phạm vi thực hiện**

1. Đối chiếu proposal, OpenAPI, SQL/test plan và brief UI; ghi phiên bản/commit nguồn cùng mapping WBS → ticket.
2. Bàn giao URL `/skills`, `/skill-reviews`, `/skill-catalogs/categories`, `/skill-catalogs/skills` và deep link `request_id`; giữ đích login và cổng đổi mật khẩu hiện hành.
3. Ghi đầu vào T02 cần xác minh trên môi trường test: head/drift, tên/status legacy, liên kết người duyệt và actor thực hiện import. Dùng lại kết quả review/kiểm tĩnh còn khớp nguồn.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Các nguồn thiết kế thống nhất | Mỗi loại contract có một nguồn chuẩn; task sau tra được mục áp dụng và đầu ra cần dùng. |
| Dữ liệu legacy chưa kiểm thực tế | T02 có danh sách cần kiểm và điều kiện dừng migration; bàn giao thiết kế không xác nhận DB đã nâng cấp. |

**Tiêu chí hoàn thành**

- [ ] Proposal/OpenAPI/SQL/brief khớp phạm vi, 14 thao tác, URL và 12 WBS; kiểm tĩnh đạt.
- [ ] Danh sách đầu vào legacy và task tiếp nhận rõ ràng; có nguồn thiết kế, kết quả review và lệnh kiểm để dùng lại.

**Phụ thuộc và cách thực hiện**

- Is blocked by: —.
- Blocks: MSK-T02.
- WBS: MSK-01.

**Kiểm tra và bằng chứng**

Phạm vi kiểm là cấu trúc và tính nhất quán của bộ nguồn thiết kế; DB/HTTP/browser thuộc các task triển khai.

- Tại gốc: `node --test scripts/check-manual-skills-contract.test.mjs scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`.
- Kiểm nguồn và diff: `node scripts/check-harness.mjs`, `git diff --check`.

Đây là kiểm tĩnh; kết quả HTTP, PostgreSQL và browser được ghi riêng khi thực hiện các ca tương ứng.

**Đầu ra bàn giao**

Bộ contract, SQL/test plan, brief UI và nguồn phiên bản cho MSK-T02–T12.

**Ngoài phạm vi**

Code nghiệp vụ, scaffold, chạy DB/migration và tạo issue Jira.

**Tham chiếu**

[Proposal](backend-manual-skills.md) mục 4–10; [OpenAPI](../api/manual-skills.openapi.json) toàn bộ; [SQL/test plan](backend-manual-skills-sql-plan.md) mục 1–8; [workflow](../ai/workflow.md).

**Truy vết nhập Jira:** `MSK-T01`; nguồn `docs/proposals/backend-manual-skills-tickets.md`.

### [MSK-T02] Triển khai migration kỹ năng và guard legacy trên PostgreSQL test

**Mục tiêu**

Nâng schema kỹ năng trên PostgreSQL test, bảo toàn dữ liệu năng lực và người duyệt cũ, giữ auth/OAP/PCA hoạt động.

**Đầu vào và điều kiện bắt đầu**

Bộ thiết kế MSK-T01 và runner PG18 hiện có. Head dự kiến 000006 phải được kiểm lại; mapping dữ liệu legacy và actor import phải được xác minh trước đường nâng cấp có dữ liệu.

**Phạm vi thực hiện**

1. Viết migration tiến thêm, preflight trước khi đánh dấu dirty và guard lại dưới khóa trong DDL; kiểm owner/head/drift/dependency, giữ nguyên migration 000001–000006.
2. Chuẩn hóa ràng buộc tên hai danh mục; chuyển approver từ employee_id qua employees.user_id sang approved_by, giữ ID/giá trị/thời điểm cũ. Lưu nguồn và thông tin approval cũ bằng sự kiện SKILL_LEGACY_IMPORTED với actor import tường minh, rồi bỏ các cột legacy theo SQL plan.
3. Tạo request MANUAL với before/proposed, trạng thái, version, FK/index và unique một PENDING/người/kỹ năng. CHECK phân biệt SQL NULL/JSON null, giữ ngoại lệ note legacy không đổi và terminal có đủ người/thời điểm/lý do.
4. Cập nhật seed/localdev/consumer theo head mới, giữ fixture lịch sử. Kiểm nâng từ DB rỗng và legacy, down/up trên DB case hy sinh, lỗi giữa DDL và metadata dirty.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Legacy user_id khác employee_id | Map đúng người duyệt; giữ ID, giá trị năng lực, thời điểm, auth/session và dữ liệu OAP/PCA. |
| Tên trùng/trắng/padding, thiếu approval/linkage, skill INACTIVE hoặc drift | Preflight từ chối và giữ dữ liệu để review mapping; không tự gộp, sửa hoặc force version. |
| JSON/state/FK/index không hợp lệ | Constraint từ chối đúng trường hợp, kể cả NULL và terminal thiếu thông tin. |
| Down khi đã có skill/official/request/key/thông báo/audit MSK | Từ chối trước DDL; down hợp lệ chỉ trên DB case đáp ứng điều kiện rỗng của SQL plan. |

**Tiêu chí hoàn thành**

- [ ] MSK-DB01–03 có kết quả PostgreSQL thật cho hai đường nâng cấp, ràng buộc, guard, down/up và dirty.
- [ ] Các trường cần bảo toàn giữ nguyên; seed/localdev và regression auth/OAP/PCA chạy với schema mới. Có cách chạy/dọn và giới hạn rollback.
- [ ] Diff được review, góp ý chặn đã xử lý và kiểm lại; có hướng dẫn tái hiện, kết quả kiểm và giới hạn cho task tiếp nhận.

**Phụ thuộc và cách thực hiện**

- Is blocked by: MSK-T01.
- Blocks: MSK-T03, MSK-T05.
- WBS: MSK-02.

**Kiểm tra và bằng chứng**

MSK-DB01–03. Đối chiếu dữ liệu trước/sau, schema/index/FK và metadata migration; down không thay backup dữ liệu.

- Tại gốc: `go vet ./...`, `go test ./... -count=1`, `go build ./...`.
- Kiểm môi trường: `python3 scripts/authdb/manage.py check`. Chỉ chạy integration sau khi runner xác nhận PG18 test thuộc quyền sở hữu.
- Chạy tuần tự: `python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./migrations ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m`. Bổ sung package test mới vào lệnh bàn giao nếu có.

DB tests chỉ dùng runner test, không dùng DATABASE_URL/.env/Neon thay thế. Các ca cạnh tranh dùng hai connection và barrier; ghi rõ adapter thật hoặc writer kiểm soát được dùng trong phép kiểm.

- Tại gốc: `node --test scripts/check-manual-skills-contract.test.mjs scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`.
- Kiểm nguồn và diff: `node scripts/check-harness.mjs`, `git diff --check`.

Đây là kiểm tĩnh; kết quả HTTP, PostgreSQL và browser được ghi riêng khi thực hiện các ca tương ứng.

**Đầu ra bàn giao**

Migration, fixture rỗng/legacy, preflight/guard và bằng chứng bảo toàn cho MSK-T03/T05.

**Ngoài phạm vi**

Neon, chuyển dữ liệu chưa có mapping được review, bảng CV/task và nghiệp vụ gửi/duyệt kỹ năng.

**Tham chiếu**

[Proposal](backend-manual-skills.md) mục 4–5; [SQL/test plan](backend-manual-skills-sql-plan.md) mục 1–4, 6–8 và MSK-DB01–03; [workflow](../ai/workflow.md).

**Truy vết nhập Jira:** `MSK-T02`; nguồn `docs/proposals/backend-manual-skills-tickets.md`.

### [MSK-T03] Triển khai CRUD API nhóm kỹ năng và kỹ năng theo quyền Giám đốc

**Mục tiêu**

Cung cấp tám API nhóm kỹ năng/kỹ năng: mọi phiên thường được đọc để chọn, Giám đốc được thêm, sửa và xóa danh mục.

**Đầu vào và điều kiện bắt đầu**

Schema và fixture MSK-T02; cơ chế auth/CSRF, audit và catalog hiện có. Dùng OpenAPI MSK cho DTO và mã lỗi.

**Phạm vi thực hiện**

1. Viết service/ports/repository, DTO và router cho GET/POST `/api/skill-categories`, `/api/skills`; PATCH/DELETE item `/{id}` của từng loại. Service giữ độc lập HTTP/pgx.
2. Validate tên sau trim Unicode, unique không phân biệt hoa/thường; skill phải có nhóm. So expected_before chính xác, đổi tên/nhóm giữ ID, no-op không tạo audit; mutation và audit cùng transaction.
3. GET hỗ trợ tìm literal, lọc nhóm, phân trang với count/items cùng snapshot. Xóa bị chặn khi còn skill, năng lực chính thức hoặc request tham chiếu.
4. Áp quyền hiện hành, cổng limited, Origin/CSRF, JSON strict 8 KiB, no-store và error mapping. Kiểm race tạo trùng/xóa nhóm–tạo skill; race xóa skill–submit dùng adapter submit tại T06. Swagger dev chỉ công bố route đã nối.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Giám đốc CRUD; employee/HR/Admin riêng đọc | GET được phép; mutation thiếu DIRECTOR bị 403, limited bị chặn nghiệp vụ. |
| Tên 100/101 code point, padding hoặc trùng khác hoa/thường | Validate sau trim; unique chặn cả hai ghi đồng thời, không tự bỏ dấu/NFC. |
| Before cũ, mục mất hoặc xóa còn tham chiếu | 409 STALE_DATA/CATALOG_IN_USE hoặc 404 theo contract; không ghi/audit một phần. |
| Tìm %/_ và page rỗng | Ký tự tìm là literal; total/items cùng snapshot. |
| Body 8192/8193 byte, không Content-Length | Body trong giới hạn được kiểm JSON; vượt giới hạn trả 413, không ghi. |

**Tiêu chí hoàn thành**

- [ ] Tám route có kiểm HTTP + PostgreSQL cho CRUD, quyền, DTO/lỗi/no-store, body và audit atomic/no-op.
- [ ] MSK-DB04, phần MSK-DB05 và MSK-DB11 thuộc danh mục đạt; có fixture/adapter cho T06 kiểm race với submit thật.
- [ ] Diff được review, góp ý chặn đã xử lý và kiểm lại; có hướng dẫn tái hiện, kết quả kiểm và giới hạn cho task tiếp nhận.

**Phụ thuộc và cách thực hiện**

- Is blocked by: MSK-T02.
- Blocks: MSK-T04, MSK-T06.
- WBS: MSK-03.

**Kiểm tra và bằng chứng**

MSK-DB04/05 phần danh mục và MSK-DB11 phần tám route. Kiểm regression auth/OAP/PCA sau khi nối router.

- Tại gốc: `go vet ./...`, `go test ./... -count=1`, `go build ./...`.
- Kiểm môi trường: `python3 scripts/authdb/manage.py check`. Chỉ chạy integration sau khi runner xác nhận PG18 test thuộc quyền sở hữu.
- Chạy tuần tự: `python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./migrations ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m`. Bổ sung package test mới vào lệnh bàn giao nếu có.

DB tests chỉ dùng runner test, không dùng DATABASE_URL/.env/Neon thay thế. Các ca cạnh tranh dùng hai connection và barrier; ghi rõ adapter thật hoặc writer kiểm soát được dùng trong phép kiểm.

- Tại gốc: `node --test scripts/check-manual-skills-contract.test.mjs scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`.
- Kiểm nguồn và diff: `node scripts/check-harness.mjs`, `git diff --check`.

Đây là kiểm tĩnh; kết quả HTTP, PostgreSQL và browser được ghi riêng khi thực hiện các ca tương ứng.

**Đầu ra bàn giao**

Tám API, catalog projections và fixture race cho MSK-T04/T06.

**Ngoài phạm vi**

Sửa trực tiếp năng lực nhân viên, đổi nghĩa kỹ năng đã dùng, trạng thái ACTIVE/INACTIVE của danh mục và UI.

**Tham chiếu**

[Proposal](backend-manual-skills.md) mục 5–6.2; [OpenAPI](../api/manual-skills.openapi.json) hai collection danh mục; [SQL/test plan](backend-manual-skills-sql-plan.md) mục 3/5/7; [FR](../product/functional-requirements.md) FR-016.

**Truy vết nhập Jira:** `MSK-T03`; nguồn `docs/proposals/backend-manual-skills-tickets.md`.

### [MSK-T04] Nối UI Giám đốc quản trị danh mục kỹ năng

**Mục tiêu**

Cho Giám đốc quản lý nhóm kỹ năng và kỹ năng trên UI bằng API/DB thật, theo thiết kế giao diện hiện có.

**Đầu vào và điều kiện bắt đầu**

Tám API MSK-T03, navigation server và các màn/primitives danh mục hiện có. Đọc web/AGENTS.md và brief UI ở web/README.md trước sửa.

**Phạm vi thực hiện**

1. Nối `/skill-catalogs/categories` và `/skill-catalogs/skills`; menu theo role server, dữ liệu và callbacks server tách demo.
2. Nối list/search/filter/page, form thêm/sửa/xóa và picker nhóm tải đủ trang. Giữ expected_before đã đọc; giải thích đổi tên/nhóm giữ định danh và không dùng để đổi nghĩa kỹ năng đã sử dụng.
3. Hiển thị validation, no-op, stale, duplicate, mục còn tham chiếu và lỗi dịch vụ. Khi mất phản hồi mutation, yêu cầu tải lại đối soát, không tự gửi lại.
4. Dọn dữ liệu đặc quyền/GET/draft khi mất quyền, limited hoặc đổi phiên; kiểm màn nhỏ, focus, Escape/nút đóng và xác nhận bỏ nháp.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Thêm/sửa/xóa rồi reload | Dữ liệu giữ trong DB; đổi tên/nhóm giữ ID và liên kết. |
| Hai tab sửa hoặc xóa mục còn tham chiếu | Hiện stale/reference rõ ràng, cho tải lại; không ghi đè dữ liệu mới. |
| Mất DIRECTOR/limited, sau đó /auth/me lỗi 503 | Xóa ngay màn và nháp đặc quyền, bỏ phản hồi GET cũ; không chờ me thành công để dọn. |
| Admin-only/HR/employee mở URL trực tiếp | Không mở giao diện quản trị hoặc dữ liệu demo thay thế. |
| Mất phản hồi lưu và thao tác bàn phím/mobile | Đối soát bằng GET; thông báo/focus dùng được và không tự gửi lại mutation. |

**Tiêu chí hoàn thành**

- [ ] Browser + PostgreSQL chứng minh CRUD/reload/stale/reference và guard theo URL/role.
- [ ] Adapter/resource tests và pnpm check đạt; kiểm focus/responsive 375 px và dọn dữ liệu khi đổi phiên.
- [ ] Diff được review, góp ý chặn đã xử lý và kiểm lại; có hướng dẫn tái hiện, kết quả kiểm và giới hạn cho task tiếp nhận.

**Phụ thuộc và cách thực hiện**

- Is blocked by: MSK-T03.
- Blocks: MSK-T10, MSK-T12.
- WBS: MSK-04.

**Kiểm tra và bằng chứng**

Browser theo các ca trong bảng; regression navigation tổ chức/hồ sơ và cổng đổi mật khẩu hiện hành.

- Tại `web/`: `pnpm check`.
- Kiểm browser với backend PostgreSQL test, dữ liệu tổng hợp và các ca trong bảng; ghi bước tái hiện, kết quả, môi trường và cách dọn. Adapter tests dùng HTTP giả chỉ chứng minh logic frontend.

- Tại gốc: `node --test scripts/check-manual-skills-contract.test.mjs scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`.
- Kiểm nguồn và diff: `node scripts/check-harness.mjs`, `git diff --check`.

Đây là kiểm tĩnh; kết quả HTTP, PostgreSQL và browser được ghi riêng khi thực hiện các ca tương ứng.

**Đầu ra bàn giao**

UI danh mục và navigation server cho MSK-T10/T12.

**Ngoài phạm vi**

Dashboard mới, thay toàn bộ thiết kế FE, sửa năng lực nhân viên hoặc action chưa có API.

**Tham chiếu**

[Proposal](backend-manual-skills.md) mục 6.2/6.4/7; [OpenAPI](../api/manual-skills.openapi.json) danh mục; [brief UI](../../web/README.md); [FR](../product/functional-requirements.md) FR-016.

**Truy vết nhập Jira:** `MSK-T04`; nguồn `docs/proposals/backend-manual-skills-tickets.md`.

### [MSK-T05] Mở rộng thông báo server cho sự kiện kỹ năng và giữ tương thích PCA

**Mục tiêu**

Đọc và đánh dấu thông báo kỹ năng qua API thông báo hiện có, đồng thời giữ thông báo hồ sơ hoạt động.

**Đầu vào và điều kiện bắt đầu**

Schema MSK-T02 và backend thông báo PCA đã bàn giao. Phần mở rộng type/resource trong canonical PCA đang là planned.

**Phạm vi thực hiện**

1. Mở validator/service/DTO cho SKILL_CHANGE_SUBMITTED, SKILL_CHANGE_APPROVED, SKILL_CHANGE_REJECTED với resource SKILL_CHANGE_REQUEST; giữ các cặp type/resource hồ sơ hiện có.
2. Giữ GET `/api/notifications` và POST `/api/notifications/{id}/read`: lọc/phân trang, total/unread_count cùng snapshot, own-read và projection thông báo tối thiểu.
3. Kiểm HTTP + PG với dữ liệu profile/skill trộn, read lặp lại, ID ngoài quyền và type/resource không hợp lệ. Chỉ chuyển trạng thái contract/Swagger sang đã triển khai cho phần thực nối; writer submit/decision được kiểm tại T06/T07.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Trang có cả thông báo hồ sơ và kỹ năng | Đọc đủ loại hợp lệ, filter/count đúng; không 500 vì validator chỉ biết loại cũ. |
| Đọc thông báo người khác hoặc type/resource sai | Không lộ nội dung; API từ chối theo contract. |
| Mất quyền mở request nhưng còn phiên thường | Vẫn đánh dấu thông báo của mình đã đọc; thao tác read không cấp quyền tài nguyên. |
| Read lặp lại hoặc lỗi DB | Read không nhân tác động; lỗi DB giữ đúng 503/500, không trả thành công giả. |

**Tiêu chí hoàn thành**

- [ ] Hai API thông báo có HTTP + PG cho mixed types, snapshot/count, quyền chính chủ và read; thông báo PCA regression đạt.
- [ ] DTO/canonical/Swagger mô tả đúng phần mở rộng đã triển khai, thông báo không chứa note/lý do hoặc dữ liệu liên hệ riêng tư.
- [ ] Diff được review, góp ý chặn đã xử lý và kiểm lại; có hướng dẫn tái hiện, kết quả kiểm và giới hạn cho task tiếp nhận.

**Phụ thuộc và cách thực hiện**

- Is blocked by: MSK-T02.
- Blocks: MSK-T06, MSK-T11.
- WBS: MSK-05.

**Kiểm tra và bằng chứng**

MSK-DB11 phần tương thích HTTP thông báo; fixture skill notification chỉ kiểm đọc/read, còn bằng chứng writer nghiệp vụ thuộc T06/T07/T12.

- Tại gốc: `go vet ./...`, `go test ./... -count=1`, `go build ./...`.
- Kiểm môi trường: `python3 scripts/authdb/manage.py check`. Chỉ chạy integration sau khi runner xác nhận PG18 test thuộc quyền sở hữu.
- Chạy tuần tự: `python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./migrations ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m`. Bổ sung package test mới vào lệnh bàn giao nếu có.

DB tests chỉ dùng runner test, không dùng DATABASE_URL/.env/Neon thay thế. Các ca cạnh tranh dùng hai connection và barrier; ghi rõ adapter thật hoặc writer kiểm soát được dùng trong phép kiểm.

- Tại gốc: `node --test scripts/check-manual-skills-contract.test.mjs scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`.
- Kiểm nguồn và diff: `node scripts/check-harness.mjs`, `git diff --check`.

Đây là kiểm tĩnh; kết quả HTTP, PostgreSQL và browser được ghi riêng khi thực hiện các ca tương ứng.

**Đầu ra bàn giao**

Backend thông báo tương thích và fixture mixed types cho MSK-T06/T11.

**Ngoài phạm vi**

Endpoint thông báo mới, polling, email/push và writer gửi/duyệt kỹ năng.

**Tham chiếu**

[Proposal](backend-manual-skills.md) mục 6.3; [OpenAPI PCA](../api/profile-change-approval.openapi.json) notifications; [SQL/test plan](backend-manual-skills-sql-plan.md) mục 5/7; [FR](../product/functional-requirements.md) phần sự kiện FR-059/060.

**Truy vết nhập Jira:** `MSK-T05`; nguồn `docs/proposals/backend-manual-skills-tickets.md`.

### [MSK-T06] Triển khai đọc năng lực, gửi và tra cứu đề nghị kỹ năng chính chủ

**Mục tiêu**

Nhân viên đọc năng lực chính thức, gửi đề nghị bổ sung/điều chỉnh và tra cứu lịch sử của mình; đề nghị chờ không sửa dữ liệu chính thức.

**Đầu vào và điều kiện bắt đầu**

Catalog adapter MSK-T03, backend thông báo MSK-T05 và schema/fixture đã bàn giao. Dùng ports auth hiện hành và giao thức khóa trong SQL plan.

**Phạm vi thực hiện**

1. Viết service/ports cho đọc official, submit, list/detail self hoặc review theo quyền. Validate ba giá trị mức/năm/note, giải trình và căn cứ văn bản; server suy ADD/UPDATE, không nhận actor/status/source từ client.
2. Giữ năm là chuỗi decimal hoặc null, phân biệt null với 0; before giữ dữ liệu đọc từ server. UPDATE giữ raw note khi proposed.note bằng before.note; ADD/note đổi thực sự mới trim, trắng về null và giới hạn 2000 code point.
3. Sinh hash từ payload có kiểu dữ liệu rõ ràng đã chuẩn hóa; key UUIDv4 lowercase chỉ cho submit trong 24 giờ. Khóa theo users → employees → catalog → request/official, tra key trước stale/pending; ghi request, kết quả key và thông báo Giám đốc cùng transaction.
4. Recheck self/status/cờ/role và recipient dưới khóa; nhiều role không nhân dòng, không có Giám đốc đủ điều kiện vẫn lưu request. List/detail dùng projection tường minh, phân trang cùng snapshot; official hỗ trợ lookup skill_id cho form.
5. Kiểm gửi đồng thời/key/pending, recipient/quyền đổi, xóa skill cạnh tranh submit bằng adapter thật T03, lỗi write/notification và commit chưa rõ.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| ADD/UPDATE hợp lệ, before null hoặc có giá trị | Một request PENDING với before đúng; official giữ nguyên, recipient theo user_id. |
| Cùng key/payload, khác payload hoặc key hết hạn | Replay cùng request_id; khác payload trả IDEMPOTENCY_CONFLICT, hết hạn trả IDEMPOTENCY_KEY_EXPIRED; không tự tạo request mới. |
| Key khác nhưng đã pending; before cũ hoặc no-op | PENDING_REQUEST_EXISTS, STALE_DATA hoặc NO_CHANGES đúng tình huống; không thêm notification. |
| Note legacy dài/padding không đổi, chỉ sửa mức/năm | Giữ chính xác before.note; note sửa thật chịu chuẩn mới. Before giả vẫn bị kiểm stale; replay sau terminal đọc được note cũ. |
| Không có Giám đốc hoặc actor cũng là Giám đốc | Request vẫn vào hàng chờ; loại actor khỏi recipients, không nhân thông báo vì tổ hợp role. |
| ID ngoài quyền, role/cờ đổi, lỗi transaction | Không lộ dữ liệu hoặc ghi một phần; commit chưa rõ trả kiểu kết quả riêng để HTTP xử lý. |

**Tiêu chí hoàn thành**

- [ ] MSK-DB06/07, phần MSK-DB05/09 và read projections có PostgreSQL thật; chứng minh pending không sửa official và request/key/notification nguyên tử.
- [ ] Service độc lập HTTP/pgx; kết quả/lỗi phân loại đủ cho T08, gồm commit chưa rõ. Note legacy và kiểu năm nhất quán giữa validate/hash/persist/read/replay.
- [ ] Diff được review, góp ý chặn đã xử lý và kiểm lại; có hướng dẫn tái hiện, kết quả kiểm và giới hạn cho task tiếp nhận.

**Phụ thuộc và cách thực hiện**

- Is blocked by: MSK-T03, MSK-T05.
- Blocks: MSK-T07, MSK-T08.
- WBS: MSK-06.

**Kiểm tra và bằng chứng**

MSK-DB05/06/07/09 theo phần submit/read; test unit Unicode/decimal/canonical hash và PG snapshot/khóa. T07 kiểm tiếp approve giữ note legacy.

- Tại gốc: `go vet ./...`, `go test ./... -count=1`, `go build ./...`.
- Kiểm môi trường: `python3 scripts/authdb/manage.py check`. Chỉ chạy integration sau khi runner xác nhận PG18 test thuộc quyền sở hữu.
- Chạy tuần tự: `python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./migrations ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m`. Bổ sung package test mới vào lệnh bàn giao nếu có.

DB tests chỉ dùng runner test, không dùng DATABASE_URL/.env/Neon thay thế. Các ca cạnh tranh dùng hai connection và barrier; ghi rõ adapter thật hoặc writer kiểm soát được dùng trong phép kiểm.

- Tại gốc: `node --test scripts/check-manual-skills-contract.test.mjs scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`.
- Kiểm nguồn và diff: `node scripts/check-harness.mjs`, `git diff --check`.

Đây là kiểm tĩnh; kết quả HTTP, PostgreSQL và browser được ghi riêng khi thực hiện các ca tương ứng.

**Đầu ra bàn giao**

Ports/use cases, projections, kiểu kết quả/lỗi và fixture/race cho MSK-T07/T08.

**Ngoài phạm vi**

HTTP/UI, CV/task/tệp, bổ sung/rút pending, REMOVE và sửa official trực tiếp.

**Tham chiếu**

[Proposal](backend-manual-skills.md) mục 6.1–6.3; [OpenAPI](../api/manual-skills.openapi.json) official/submit/list/detail; [SQL/test plan](backend-manual-skills-sql-plan.md) mục 3–5/7; [FR](../product/functional-requirements.md) FR-017/023 và phần văn bản FR-024.

**Truy vết nhập Jira:** `MSK-T06`; nguồn `docs/proposals/backend-manual-skills-tickets.md`.

### [MSK-T07] Triển khai transaction Giám đốc chấp thuận hoặc từ chối kỹ năng

**Mục tiêu**

Giám đốc quyết định một lần trên đề nghị của người khác; chấp thuận cập nhật năng lực và thông báo cùng transaction, từ chối giữ năng lực cũ.

**Đầu vào và điều kiện bắt đầu**

Use cases/ports/fixture MSK-T06 và giao thức khóa đã duyệt. Actor lấy từ phiên server; request đã có version và snapshot before.

**Phạm vi thực hiện**

1. Snapshot actor/target/recipient rồi khóa user/employee theo ID tăng dần, catalog và request đúng thứ tự. Recheck DIRECTOR, liên kết, không self và PENDING trước khi kiểm expected_request_version.
2. Approve so toàn bộ before với official hiện hành, kể cả absence; kiểm employee target ACTIVE rồi insert/update official cùng terminal và owner notification. Reject vẫn kiểm version/lý do nhưng không yêu cầu before khớp hoặc target ACTIVE.
3. Cả hai quyết định có reason; expert_opinion tùy chọn ghi nguồn/nội dung do Giám đốc nhập. Request giữ người/thời điểm/before/proposed/lý do/ý kiến, không thêm audit trùng lịch sử quyết định.
4. Kiểm approve/approve, approve/reject, stale, actor mất quyền/cờ đổi, actor=target, target inactive/user locked; lỗi từng write/notification và commit chưa rõ. Note legacy không đổi được giữ nguyên sau approve/read/replay.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Hai người/tab xử lý hoặc gọi lại request kết thúc | Chỉ một quyết định và owner notification; lần sau trả REQUEST_ALREADY_RESOLVED. |
| Version cũ hoặc approve có before không còn khớp | STALE_REQUEST hoặc STALE_DATA; request vẫn PENDING, không ghi đè official. |
| Giám đốc xử lý đề nghị của chính mình | Từ chối quyền cả approve/reject, kể cả role tổ hợp. |
| Employee target inactive hoặc user target locked/limited | Inactive chặn approve bằng TARGET_INACTIVE nhưng vẫn reject được; trạng thái user target không áp như điều kiện phiên của người duyệt. |
| Reject khi official đã thay đổi | Được từ chối nếu các điều kiện quyết định còn hợp lệ; official hiện hành giữ nguyên. |
| Lỗi lưu/notification hoặc mất xác nhận commit | Lỗi đã xác định rollback toàn bộ; commit chưa rõ không trả thành công hoặc tự retry. |

**Tiêu chí hoàn thành**

- [ ] MSK-DB08, phần MSK-DB09/10 có PostgreSQL thật, hai connection/barrier; terminal bất biến qua ports và chỉ một quyết định thắng.
- [ ] Official/terminal/notification nguyên tử; lý do bắt buộc, note legacy giữ nguyên, không sinh audit trùng submission/decision. Có kiểu kết quả riêng cho commit chưa rõ.
- [ ] Diff được review, góp ý chặn đã xử lý và kiểm lại; có hướng dẫn tái hiện, kết quả kiểm và giới hạn cho task tiếp nhận.

**Phụ thuộc và cách thực hiện**

- Is blocked by: MSK-T06.
- Blocks: MSK-T08.
- WBS: MSK-07.

**Kiểm tra và bằng chứng**

MSK-DB08/09/10 phần decision; race dùng adapter submit/decision thật. Kiểm quyền hiện hành, snapshot participants và lỗi ở từng bước transaction.

- Tại gốc: `go vet ./...`, `go test ./... -count=1`, `go build ./...`.
- Kiểm môi trường: `python3 scripts/authdb/manage.py check`. Chỉ chạy integration sau khi runner xác nhận PG18 test thuộc quyền sở hữu.
- Chạy tuần tự: `python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./migrations ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m`. Bổ sung package test mới vào lệnh bàn giao nếu có.

DB tests chỉ dùng runner test, không dùng DATABASE_URL/.env/Neon thay thế. Các ca cạnh tranh dùng hai connection và barrier; ghi rõ adapter thật hoặc writer kiểm soát được dùng trong phép kiểm.

- Tại gốc: `node --test scripts/check-manual-skills-contract.test.mjs scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`.
- Kiểm nguồn và diff: `node scripts/check-harness.mjs`, `git diff --check`.

Đây là kiểm tĩnh; kết quả HTTP, PostgreSQL và browser được ghi riêng khi thực hiện các ca tương ứng.

**Đầu ra bàn giao**

Decision ports, kiểu kết quả/lỗi và fixture/race cho MSK-T08.

**Ngoài phạm vi**

Sửa proposed/căn cứ qua decision, tự duyệt, sửa official trực tiếp, REMOVE, email và cấp phê duyệt mới.

**Tham chiếu**

[Proposal](backend-manual-skills.md) mục 6.1–6.3; [OpenAPI](../api/manual-skills.openapi.json) approve/reject; [SQL/test plan](backend-manual-skills-sql-plan.md) mục 5/7; [FR](../product/functional-requirements.md) FR-027 và phần văn bản FR-026.

**Truy vết nhập Jira:** `MSK-T07`; nguồn `docs/proposals/backend-manual-skills-tickets.md`.

### [MSK-T08] Tích hợp sáu API năng lực và đề nghị kỹ năng với PostgreSQL

**Mục tiêu**

Cung cấp sáu API năng lực và đề nghị kỹ năng để UI chính chủ/Giám đốc dùng trực tiếp với PostgreSQL.

**Đầu vào và điều kiện bắt đầu**

Use cases và kiểu kết quả/lỗi MSK-T06/T07; canonical MSK, middleware auth/CSRF và composition root hiện có.

**Phạm vi thực hiện**

1. Nối GET `/api/employees/{id}/skills`, POST `/api/employees/{id}/skill-change-requests`, GET `/api/skill-change-requests`, GET `/api/skill-change-requests/{id}`, POST `/api/skill-change-requests/{id}/approve` và POST `/api/skill-change-requests/{id}/reject`. Handler chỉ gọi service/port.
2. Áp phiên thường, quyền self/review/DIRECTOR, projection/DTO allowlist và error mapper. Mutation kiểm Origin/CSRF, Content-Type, JSON strict, Unicode, duplicate keys và body 64 KiB thực đọc; submit kiểm UUIDv4 key, GET kiểm query allowlist.
3. Kiểm HTTP + PG cho replay/pending/stale/terminal, lỗi DB/timeout/commit chưa rõ; response no-store, không Set-Cookie nghiệp vụ. Giữ giới hạn body auth/OAP/catalog và regression router.
4. Bổ sung sáu operation đã nối vào Swagger dev từ canonical, giữ tám catalog và toàn bộ API auth/OAP/PCA hiện có.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Self, Giám đốc review, HR/Admin-only, ID ngoài quyền hoặc limited | 401/403/404 đúng nguồn; không Admin bypass hoặc lộ hồ sơ riêng tư. |
| Submit replay, pending/stale và decision đã kết thúc | Đúng status/DTO/mã lỗi; replay hợp lệ chỉ ở submit, không nhân tác động. |
| Body 65536/65537 byte không Content-Length | Trong giới hạn vẫn qua kiểm JSON; vượt giới hạn trả 413, không ghi. |
| Media/query/key/Unicode/duplicate JSON key/Origin/CSRF sai | 400/403/415 theo contract trước khi mutation; không đổi giới hạn route cũ. |
| DB lỗi hoặc commit chưa rõ | 503/500 đúng nguồn; COMMIT_OUTCOME_UNKNOWN khi chưa xác nhận, không success giả/cookie mới. |

**Tiêu chí hoàn thành**

- [ ] Sáu method/path có HTTP + PostgreSQL thật; route inventory/DTO/security/lỗi khớp canonical và no-store/không Set-Cookie.
- [ ] MSK-DB10/11 phần sáu route đạt; regression năm auth, 16 OAP, 10 PCA và tám catalog đạt. Swagger chỉ xuất operation đã triển khai.
- [ ] Diff được review, góp ý chặn đã xử lý và kiểm lại; có hướng dẫn tái hiện, kết quả kiểm và giới hạn cho task tiếp nhận.

**Phụ thuộc và cách thực hiện**

- Is blocked by: MSK-T06, MSK-T07.
- Blocks: MSK-T09, MSK-T10.
- WBS: MSK-08.

**Kiểm tra và bằng chứng**

MSK-DB10/11 phần HTTP; kiểm body byte thực, response allowlist, key replay và lỗi từng tầng. HTTP giả không thay transaction PG.

- Tại gốc: `go vet ./...`, `go test ./... -count=1`, `go build ./...`.
- Kiểm môi trường: `python3 scripts/authdb/manage.py check`. Chỉ chạy integration sau khi runner xác nhận PG18 test thuộc quyền sở hữu.
- Chạy tuần tự: `python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./migrations ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m`. Bổ sung package test mới vào lệnh bàn giao nếu có.

DB tests chỉ dùng runner test, không dùng DATABASE_URL/.env/Neon thay thế. Các ca cạnh tranh dùng hai connection và barrier; ghi rõ adapter thật hoặc writer kiểm soát được dùng trong phép kiểm.

- Tại gốc: `node --test scripts/check-manual-skills-contract.test.mjs scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`.
- Kiểm nguồn và diff: `node scripts/check-harness.mjs`, `git diff --check`.

Đây là kiểm tĩnh; kết quả HTTP, PostgreSQL và browser được ghi riêng khi thực hiện các ca tương ứng.

**Đầu ra bàn giao**

Sáu API, DTO input/output và lỗi tái hiện cho MSK-T09/T10.

**Ngoài phạm vi**

UI, upload, endpoint bổ sung/rút và sửa năng lực trực tiếp.

**Tham chiếu**

[Proposal](backend-manual-skills.md) mục 5–6; [OpenAPI](../api/manual-skills.openapi.json) sáu operation official/request; [SQL/test plan](backend-manual-skills-sql-plan.md) mục 7; [FR](../product/functional-requirements.md) FR-017/023/027.

**Truy vết nhập Jira:** `MSK-T08`; nguồn `docs/proposals/backend-manual-skills-tickets.md`.

### [MSK-T09] Nối UI năng lực chính thức, gửi đề nghị và lịch sử cá nhân

**Mục tiêu**

Nhân viên xem năng lực được công nhận, gửi đề nghị thủ công và tra cứu lịch sử trên `/skills`, tách rõ dữ liệu chờ với dữ liệu chính thức.

**Đầu vào và điều kiện bắt đầu**

Sáu API MSK-T08, catalog GET và navigation server hiện có. Dùng PersonalSkills/form/detail/history theo web/AGENTS.md và brief web/README.md.

**Phạm vi thực hiện**

1. Nối danh sách năng lực, lịch sử/detail self và deep link `/skills?request_id=<id>` sau GET có quyền. Phân trang năng lực riêng; query page/page_size/status trên URL dành cho lịch sử.
2. Picker tải đủ trang; trước mở form GET official theo skill_id ở page 1, total 0/1 để dựng before. Gửi đúng payload canonical và key UUIDv4 trong bộ nhớ; server tự suy ADD/UPDATE.
3. Giữ before/nháp qua refresh cùng phiên; hủy không POST. Năm giữ chuỗi, kiểm mẫu/scale và thêm .0 khi cần, không Number/toFixed. Note không sửa gửi nguyên snapshot; note sửa áp chuẩn mới.
4. Hiển thị validation/no-op/pending/stale/503 và mất phản hồi. Sau kết quả chưa rõ, khóa submit lại và hướng dẫn GET lịch sử/detail đối soát; clear dữ liệu/abort khi 401/403/đổi phiên, không demo fallback.
5. Hiển thị hướng dẫn bốn mức, null là chưa cung cấp; kiểm Unicode/body, focus/mobile/dirty. Chỉ mở action văn bản thuộc đợt này.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Gửi ADD/UPDATE rồi reload | Pending/lịch sử khôi phục từ server; official giữ nguyên đến khi được chấp thuận. |
| Năm null, 0 hoặc nhập quá một chữ số lẻ | Phân biệt chưa cung cấp với 0; không làm tròn để hợp lệ, không tự suy mức. |
| Note legacy dài/padding, chỉ sửa mức/năm | Giữ raw note; không cắt/trim trường không sửa hoặc buộc người dùng sửa note cũ. |
| Hơn 20 năng lực, chọn kỹ năng ngoài trang đang xem | Lookup skill_id trước mở form; dựng đúng before, không nhầm thành ADD. |
| Before cũ, pending trùng hoặc mất phản hồi submit | Thông báo đúng và GET đối soát; không tự gửi lại hoặc tự đổi key để né conflict. |
| Deep link/Back/logout/đổi người hoặc refresh nền | Không lộ dữ liệu người trước; nháp cùng phiên không bị dựng lại bởi refresh. |

**Tiêu chí hoàn thành**

- [ ] Adapter/resource tests và pnpm check đạt; browser + PG chứng minh gửi/lịch sử/detail/reload và năng lực chính thức tách pending.
- [ ] Kiểm lookup ngoài trang, note legacy, năm/null, stale/unknown, phiên/deep link và focus/responsive 375 px; không lưu payload/key/dữ liệu riêng tư vào localStorage/URL.
- [ ] Diff được review, góp ý chặn đã xử lý và kiểm lại; có hướng dẫn tái hiện, kết quả kiểm và giới hạn cho task tiếp nhận.

**Phụ thuộc và cách thực hiện**

- Is blocked by: MSK-T08.
- Blocks: MSK-T11, MSK-T12.
- WBS: MSK-09.

**Kiểm tra và bằng chứng**

Browser theo các ca trong bảng; adapter tests kiểm Unicode/body/payload và phản hồi cũ. Căn cứ đợt này chỉ có giải trình/văn bản ngoài hệ thống.

- Tại `web/`: `pnpm check`.
- Kiểm browser với backend PostgreSQL test, dữ liệu tổng hợp và các ca trong bảng; ghi bước tái hiện, kết quả, môi trường và cách dọn. Adapter tests dùng HTTP giả chỉ chứng minh logic frontend.

- Tại gốc: `node --test scripts/check-manual-skills-contract.test.mjs scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`.
- Kiểm nguồn và diff: `node scripts/check-harness.mjs`, `git diff --check`.

Đây là kiểm tĩnh; kết quả HTTP, PostgreSQL và browser được ghi riêng khi thực hiện các ca tương ứng.

**Đầu ra bàn giao**

Luồng self F2, adapter/resource và kịch bản kiểm cho MSK-T11/T12.

**Ngoài phạm vi**

CV/tệp/căn cứ task, bổ sung/rút pending, sửa official trực tiếp; FR-024 chưa đủ toàn bộ AC và FR-025/028 còn phần tiếp nối.

**Tham chiếu**

[Proposal](backend-manual-skills.md) mục 6.2/6.4/7; [OpenAPI](../api/manual-skills.openapi.json) official/submit/list/detail; [brief UI](../../web/README.md); [FR](../product/functional-requirements.md) FR-017/023 và phần văn bản FR-024.

**Truy vết nhập Jira:** `MSK-T09`; nguồn `docs/proposals/backend-manual-skills-tickets.md`.

### [MSK-T10] Nối UI Giám đốc đối chiếu và quyết định kỹ năng

**Mục tiêu**

Giám đốc đối chiếu và chấp thuận/từ chối đề nghị của người khác với lý do, dữ liệu nền và phiên bản đã đọc.

**Đầu vào và điều kiện bắt đầu**

UI danh mục/navigation MSK-T04, sáu API MSK-T08 và primitives hàng chờ/dialog hiện có. Dùng brief UI server, không dùng guard/store demo làm nguồn quyền.

**Phạm vi thực hiện**

1. Nối `/skill-reviews` với mặc định PENDING; filter status/employee_id/skill_id, phân trang, deep link request_id và lịch sử cùng người/kỹ năng. Đóng detail giữ filter/page.
2. Hiển thị mã/vị trí/ngày gửi, before/proposed/current, hướng dẫn bốn mức, giải trình và căn cứ văn bản. Form quyết định có lý do cho cả hai thao tác; ý kiến chuyên môn tùy chọn ghi rõ nguồn/nội dung.
3. Giữ snapshot expected_request_version; không cho tự xử lý request của mình. Target inactive chặn approve nhưng vẫn cho reject; UI theo đúng điều kiện server.
4. Hiển thị stale/version/terminal và unknown outcome; tải lại GET để đối soát, không tự quyết lại. Khi 403 mất role/limited, xóa ngay list/detail/draft và bỏ phản hồi cũ kể cả /auth/me tiếp theo 503; kiểm focus/mobile.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Approve/reject có lý do rồi reload | Terminal, lý do/người/thời điểm đúng; approve đổi official, reject giữ official. |
| Giám đốc mở request của mình, kể cả role tổ hợp | Không có đường tự approve hoặc reject. |
| Target employee inactive hoặc official đã đổi | Vẫn cho reject hợp lệ; approve bị TARGET_INACTIVE hoặc STALE_DATA, không ghi đè. |
| Hai tab/version cũ hoặc request đã xử lý | Hiện conflict tương ứng và tải lại; không gửi quyết định dựa màn cũ. |
| Mất DIRECTOR/limited, me 503 hoặc mất phản hồi quyết định | Dọn ngay dữ liệu ngoài quyền; kết quả chưa rõ được đối soát bằng GET, không tự replay. |

**Tiêu chí hoàn thành**

- [ ] Browser + PG chứng minh hàng chờ/detail/history, approve/reject, không self, target inactive và stale/version/terminal.
- [ ] Adapter/resource tests và pnpm check đạt; kiểm clear quyền/phiên, unknown outcome, focus và responsive 375 px. Phần căn cứ task/CV chưa nối được thể hiện đúng giới hạn.
- [ ] Diff được review, góp ý chặn đã xử lý và kiểm lại; có hướng dẫn tái hiện, kết quả kiểm và giới hạn cho task tiếp nhận.

**Phụ thuộc và cách thực hiện**

- Is blocked by: MSK-T04, MSK-T08.
- Blocks: MSK-T11, MSK-T12.
- WBS: MSK-10.

**Kiểm tra và bằng chứng**

Browser theo các ca trong bảng, ít nhất hai tab quyết định; kiểm regression navigation HR/Admin và cổng limited.

- Tại `web/`: `pnpm check`.
- Kiểm browser với backend PostgreSQL test, dữ liệu tổng hợp và các ca trong bảng; ghi bước tái hiện, kết quả, môi trường và cách dọn. Adapter tests dùng HTTP giả chỉ chứng minh logic frontend.

- Tại gốc: `node --test scripts/check-manual-skills-contract.test.mjs scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`.
- Kiểm nguồn và diff: `node scripts/check-harness.mjs`, `git diff --check`.

Đây là kiểm tĩnh; kết quả HTTP, PostgreSQL và browser được ghi riêng khi thực hiện các ca tương ứng.

**Đầu ra bàn giao**

Luồng Giám đốc F3, adapter/resource và kịch bản kiểm cho MSK-T11/T12.

**Ngoài phạm vi**

HR duyệt trước, sửa proposed, tự thêm cấp chuyên gia/chữ ký và mở quyền task/CV chưa nối.

**Tham chiếu**

[Proposal](backend-manual-skills.md) mục 6.1–6.4/7; [OpenAPI](../api/manual-skills.openapi.json) list/detail/approve/reject; [brief UI](../../web/README.md); [FR](../product/functional-requirements.md) FR-027 và phần văn bản FR-026.

**Truy vết nhập Jira:** `MSK-T10`; nguồn `docs/proposals/backend-manual-skills-tickets.md`.

### [MSK-T11] Nối thông báo kỹ năng tới đúng màn cá nhân hoặc Giám đốc

**Mục tiêu**

Người dùng nhận biết sự kiện kỹ năng và mở đúng đề nghị chính chủ hoặc hàng chờ Giám đốc từ khung thông báo hiện có.

**Đầu vào và điều kiện bắt đầu**

Backend thông báo MSK-T05, self UI MSK-T09, review UI MSK-T10 và khung thông báo PCA hiện có.

**Phạm vi thực hiện**

1. Mở adapter/UI allowlist cho ba loại SKILL_CHANGE_* và resource SKILL_CHANGE_REQUEST; giữ chuông/panel/filter/page/unread_count, cùng thông báo hồ sơ.
2. Điều hướng owner tới `/skills?request_id=<id>`, người review có quyền tới `/skill-reviews?request_id=<id>`. GET tài nguyên được phép thành công rồi mới đánh dấu read/đóng; chỉ mở panel không read.
3. Mất quyền tài nguyên vẫn cho own-read riêng; lỗi tải/read hiển thị đúng và chỉ retry read thủ công. Dọn cache/GET/read đang chờ khi đổi phiên, bỏ phản hồi cũ.
4. Kiểm submit → Giám đốc nhận → quyết định → owner nhận, mixed profile/skill, link cũ/mất quyền và Escape/outside/đóng/focus/mobile.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Mở chuông, đổi filter hoặc phân trang | Count/danh sách đúng; thông báo chưa tự bị đánh dấu đã đọc. |
| Chọn thông báo gửi hoặc kết quả | Đúng màn và request_id sau kiểm quyền; read chỉ sau GET thành công, không phải quyết định nghiệp vụ. |
| Link cũ sau mất DIRECTOR hoặc đổi phiên | Không mở nội dung ngoài quyền hoặc dữ liệu cũ; thông báo chính chủ vẫn có read riêng. |
| GET/read 503 hoặc mất phản hồi read | Hiển thị lỗi, cho đối soát/retry read thủ công; không lặp submit/decision. |
| Thông báo hồ sơ và kỹ năng cùng tồn tại | Cả hai mở đúng tài nguyên; message không chứa note/lý do hoặc dữ liệu liên hệ riêng tư. |

**Tiêu chí hoàn thành**

- [ ] Adapter tests/pnpm check và browser + PG chứng minh đủ chuỗi gửi → nhận → quyết định → nhận kết quả, filter/count/read và quyền.
- [ ] Regression thông báo hồ sơ đạt; kiểm phiên cũ, deep link, focus/Escape/outside và màn 375 px.
- [ ] Diff được review, góp ý chặn đã xử lý và kiểm lại; có hướng dẫn tái hiện, kết quả kiểm và giới hạn cho task tiếp nhận.

**Phụ thuộc và cách thực hiện**

- Is blocked by: MSK-T05, MSK-T09, MSK-T10.
- Blocks: MSK-T12.
- WBS: MSK-11.

**Kiểm tra và bằng chứng**

Browser với writer submit/decision thật và backend PG; fixture đọc thông báo riêng không thay bằng chứng chuỗi nghiệp vụ.

- Tại `web/`: `pnpm check`.
- Kiểm browser với backend PostgreSQL test, dữ liệu tổng hợp và các ca trong bảng; ghi bước tái hiện, kết quả, môi trường và cách dọn. Adapter tests dùng HTTP giả chỉ chứng minh logic frontend.

- Tại gốc: `node --test scripts/check-manual-skills-contract.test.mjs scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`.
- Kiểm nguồn và diff: `node scripts/check-harness.mjs`, `git diff --check`.

Đây là kiểm tĩnh; kết quả HTTP, PostgreSQL và browser được ghi riêng khi thực hiện các ca tương ứng.

**Đầu ra bàn giao**

Thông báo kỹ năng trên UI và kịch bản F2/F3 cho MSK-T12.

**Ngoài phạm vi**

Trang thông báo mới, email, push, polling/realtime và sự kiện ngoài contract.

**Tham chiếu**

[Proposal](backend-manual-skills.md) mục 6.3–6.4/7; [OpenAPI PCA](../api/profile-change-approval.openapi.json) notifications; [brief UI](../../web/README.md); [FR](../product/functional-requirements.md) phần sự kiện FR-059/060.

**Truy vết nhập Jira:** `MSK-T11`; nguồn `docs/proposals/backend-manual-skills-tickets.md`.

### [MSK-T12] Kiểm xuyên luồng kỹ năng và bàn giao demo ngày 03/12

**Mục tiêu**

Bàn giao demo kỹ năng F1–F3 chạy lại được cho ngày 03/12/2026, cùng bằng chứng kiểm thử, FR coverage và giới hạn còn lại.

**Đầu vào và điều kiện bắt đầu**

UI danh mục MSK-T04, self MSK-T09, review MSK-T10 và thông báo MSK-T11; môi trường PG test và tài khoản tổng hợp đủ vai trò.

**Phạm vi thực hiện**

1. Kiểm browser: Giám đốc tạo danh mục → nhân viên gửi ADD/UPDATE → Giám đốc nhận thông báo → approve/reject → owner nhận → đối chiếu official/lịch sử → reload/logout/Back.
2. Kiểm stale, một pending, hai người quyết định, mất quyền/limited, lỗi DB/mạng và kết quả chưa rõ; đối chiếu UI, HTTP và dữ liệu PG theo từng ca. Khép race catalog–submit bằng hai adapter thật.
3. Tổng hợp MSK-DB01–12, 14 method/path MSK và hai API notification mở rộng với router/OpenAPI/Swagger; chạy regression auth/OAP/PCA.
4. Cập nhật runbook chạy/dọn, fixture/tài khoản demo, flow báo cáo và bảng FR đã đạt/Partial/chưa kiểm; liên kết lỗi còn mở với task gốc. Ghi đầu vào kỹ năng cho miền dự án/phân công tiếp theo.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| ADD được chấp thuận, UPDATE được chấp thuận hoặc từ chối | Năng lực, lịch sử và thông báo đúng ở UI/HTTP/PG; pending/reject không sửa official. |
| Stale/quyền/phiên/DB/mất phản hồi | Không ghi đè, lộ dữ liệu, tự gửi lại hoặc hiển thị success khi chưa xác nhận. |
| Chạy lại demo và dọn môi trường | Lệnh/tài khoản tổng hợp/bước đổi vai rõ; cleanup có kết quả, không cần dữ liệu hay credential ngoài tài liệu local an toàn. |
| Đối chiếu FR và phần chưa nối | FR-024/026 ghi Partial theo phạm vi văn bản; CV/task/tệp/bổ sung/rút và các ca chưa chạy được nêu rõ. |

**Tiêu chí hoàn thành**

- [ ] Go/PG/FE/contract/harness và browser F1–F3 có kết quả; MSK-DB01–12 tổng hợp đúng tầng, race dùng adapter thật và lỗi chặn đã sửa/kiểm lại.
- [ ] Runbook đủ để Quân chạy lại demo từ môi trường sạch; có FR coverage, lỗi/giới hạn và bước tiếp, không nhận toàn D03/notifications chỉ từ luồng MANUAL.
- [ ] Diff được review, góp ý chặn đã xử lý và kiểm lại; có hướng dẫn tái hiện, kết quả kiểm và giới hạn cho task tiếp nhận.

**Phụ thuộc và cách thực hiện**

- Is blocked by: MSK-T04, MSK-T09, MSK-T10, MSK-T11.
- Blocks: —.
- WBS: MSK-12.

**Kiểm tra và bằng chứng**

MSK-DB01–12 và browser xuyên luồng; dùng lại bằng chứng còn khớp nguồn, chạy bổ sung phần thiếu. Ghi riêng kiểm bằng fault kiểm soát với lỗi môi trường thực.

- Tại gốc: `go vet ./...`, `go test ./... -count=1`, `go build ./...`.
- Kiểm môi trường: `python3 scripts/authdb/manage.py check`. Chỉ chạy integration sau khi runner xác nhận PG18 test thuộc quyền sở hữu.
- Chạy tuần tự: `python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./migrations ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m`. Bổ sung package test mới vào lệnh bàn giao nếu có.

DB tests chỉ dùng runner test, không dùng DATABASE_URL/.env/Neon thay thế. Các ca cạnh tranh dùng hai connection và barrier; ghi rõ adapter thật hoặc writer kiểm soát được dùng trong phép kiểm.

- Tại `web/`: `pnpm check`.
- Kiểm browser với backend PostgreSQL test, dữ liệu tổng hợp và các ca trong bảng; ghi bước tái hiện, kết quả, môi trường và cách dọn. Adapter tests dùng HTTP giả chỉ chứng minh logic frontend.

- Tại gốc: `node --test scripts/check-manual-skills-contract.test.mjs scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`.
- Kiểm nguồn và diff: `node scripts/check-harness.mjs`, `git diff --check`.

Đây là kiểm tĩnh; kết quả HTTP, PostgreSQL và browser được ghi riêng khi thực hiện các ca tương ứng.

**Đầu ra bàn giao**

Runbook, ma trận bằng chứng/FR, tài khoản tổng hợp và kịch bản demo ngày 03/12; đầu vào năng lực cho D04.

**Ngoài phạm vi**

CV/storage/task, công thức phân công, Neon/HTTPS/deploy/pentest toàn app và thao tác Git/Jira tự động.

**Tham chiếu**

[Proposal](backend-manual-skills.md) mục 3/7–10; [OpenAPI](../api/manual-skills.openapi.json) toàn bộ và [notification PCA](../api/profile-change-approval.openapi.json); [SQL/test plan](backend-manual-skills-sql-plan.md) mục 7–8; [FR](../product/functional-requirements.md) FR-016/017/023/024/026/027, phần FR-059/060; [delivery-plan](../ai/delivery-plan.md).

**Truy vết nhập Jira:** `MSK-T12`; nguồn `docs/proposals/backend-manual-skills-tickets.md`.

