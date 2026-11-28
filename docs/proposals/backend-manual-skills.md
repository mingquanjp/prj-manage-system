# Danh mục kỹ năng, đề nghị thủ công và Giám đốc xét duyệt

Ngày 27/11/2026. Đợt D03 cho báo cáo 03/12; mã thiết kế **MSK**, Epic **DATN-49**, task **DATN-50–61**. Đây là bộ thiết kế trước triển khai. Base đối chiếu `main 8fc3ebd` sau merge PR63–65.

## 1. Nguồn và hiện trạng

- [Delivery-plan](../ai/delivery-plan.md) mục3–6: D03 đi trước dự án; 15–20 giờ Quân/tuần có Codex hỗ trợ, dành thời gian review/kiểm chứng. Không cam kết số task tương đương số giờ.
- [FR](../product/functional-requirements.md) FR-016/017/023/024/026/027, phần sự kiện FR-059/060; [bản đồ miền](../product/business-domains.md) D03/H01.
- [Workflow kỹ năng](../../.claude/docs/skill_approval_workflow.md), [domain rules](../../.claude/docs/domain_rules.md) mục năng lực, [quyền](../../.claude/docs/roles_permissions.md), [chốt hiện hành](../design/database-review-decisions-2026-10-14.md), [dictionary](../design/database-step5b-supporting-dictionary-2026-10-14.md) mục6–9. Chốt 27/10 ưu tiên ADD/UPDATE, không gỡ kỹ năng.
- [Kiến trúc](../../.claude/docs/architecture.md), [API](../../.claude/docs/api_design.md), [bảo mật](../../.claude/docs/security.md), [screen contracts](../design/figma-make/screen-contracts.md) UI06/07/09, [web instructions](../../web/AGENTS.md), [web README](../../web/README.md), [workflow Git](../ai/workflow.md), [comment](../ai/code-comments.md).
- Tái sử dụng auth/session/cổng limited, PG runner, audit/idempotency/notification PCA, primitives và màn kỹ năng demo. Main có migration000001–000006; skills.status, employee_skills.source/approved_by_employee_id vẫn legacy. Chưa có service/API kỹ năng thật. Notification service/DTO hiện chỉ chấp nhận PROFILE_CHANGE_*: phải mở rộng cả validator, HTTP contract và UI, không chỉ INSERT loại mới.

## 2. Mục tiêu và phạm vi

F1 Giám đốc quản trị nhóm/kỹ năng. F2 chủ hồ sơ xem official, gửi ADD/UPDATE và đọc lịch sử/chi tiết. F3 Giám đốc đối chiếu, quyết định cùng official/thông báo; owner thấy kết quả sau reload. Request giữ lịch sử quyết định, không sao chép sang audit_events. Bốn mức BASIC/INTERMEDIATE/ADVANCED/EXPERT; số năm không suy mức. Năng lực được dùng tiếp cho D04/D06, chưa làm engine.

**Giới hạn được Quân xác nhận qua chat 27/11:** giải trình/căn cứ và ý kiến chuyên môn dạng văn bản; chưa CV/tệp/task evidence, bổ sung/rút pending. Đây là cắt đợt, không xóa quyền/chức năng đã chốt cho đợt sau. FR-024/026 Partial; FR-025/028 tiếp nối. Không tự nhận toàn notifications hoặc cả D03 Done.

Ngoài phạm vi: CV/upload/storage/model, task/project, công thức, tự ghi official ngoài quyết định, REMOVE/tự duyệt, đổi role/lock/reset, mail/push/polling, áp Neon/deploy. Không metadata file giả, fetch link ngoài, tự tạo issue/Jira/Git hoặc scaffold nghiệp vụ.

## 3. Luồng demo và kết quả

1. Giám đốc tạo nhóm và kỹ năng; sửa tên giữ ID, xóa còn tham chiếu bị chặn.
2. Nhân viên mở /skills: official rỗng là dữ liệu GET đã xác minh, khác lỗi tải. Chọn kỹ năng và gửi mức/năm/note, giải trình; chỉ tạo PENDING.
3. Giám đốc mở thông báo/hàng chờ, xem before/proposed/current, hướng dẫn mức, giải trình và lịch sử cùng kỹ năng. Nếu cần ý kiến ngoài, nhập nguồn/nội dung vào quyết định; không giả chuyên gia đăng nhập.
4. Approve có lý do → official + terminal request + owner notification cùng commit. Reject có lý do → official không đổi. Request giữ before/proposed/người/thời điểm/lý do, theo dictionary mục20 không ghi thêm audit trùng lịch sử miền. Kết quả chỉ sau commit xác nhận.
5. Owner mở kết quả → GET kiểm quyền → read notification, tải lại official/history; reload giữ DB. Hai tab xử lý chỉ một kết luận; stale không ghi đè, không tự POST lại.

```mermaid
sequenceDiagram
  actor E as Nhân viên
  participant UI as Kỹ năng server
  participant API as Service và adapter
  participant DB as PostgreSQL
  actor D as Giám đốc
  E->>UI: Gửi mức và giải trình
  UI->>API: Submit before và key
  API->>DB: Khóa và kiểm self, official, pending, key
  DB->>DB: Request và thông báo Giám đốc cùng transaction
  DB-->>UI: Commit xác nhận rồi 201
  D->>API: Approve hoặc reject kèm version và lý do
  API->>DB: Recheck quyền, không self, version và PENDING
  DB->>DB: Official khi approve, terminal và notification
  DB-->>UI: Commit xác nhận rồi DTO
  Note over API,DB: Mất xác nhận commit là kết quả chưa rõ, không retry mutation
```

## 4. Lựa chọn và brief đã chốt

| Mã | Quyết định / nguồn |
|---|---|
| MSK-A01 | Quản trị và quyết định là DIRECTOR; HR/Admin không bypass. Theo workflow hiện hành, không phải quyền mới. |
| MSK-A02 | Chỉ MANUAL ADD/UPDATE; server suy operation từ before/official. Không endpoint gỡ, không direct edit official. Chốt27/10. |
| MSK-A03 | Một PENDING/người/kỹ năng, dùng chung nguồn khi nối CV sau; giữ request.version. Theo dictionary/workflow. |
| MSK-A04 | Mức/năm/note là giá trị official; before null nghĩa chưa có, years null khác0. Theo dictionary; contract validation mục6. |
| MSK-A05 | /skills; /skill-reviews; /skill-catalogs/categories và /skill-catalogs/skills; ?request_id mở chi tiết. Quân xác nhận27/11. Giữ home/login hiện hành. |
| MSK-A06 | Chỉ văn bản trong đợt; không CV/tệp/task hoặc supplement/withdraw. Quân xác nhận27/11, không giả các AC đó Done. |
| MSK-A07 | Tái sử dụng cookie/CSRF/Origin, JSON strict, 64KiB cho request; danh mục8KiB. Idempotency submit24h theo cơ chế PCA, không áp cho quyết định. Đây là thiết kế kỹ thuật đợt này. |
| MSK-A08 | Migration tiến thêm từ head000006; guard legacy/mapping, không sửa lịch sử hoặc áp Neon. Không ép dựng đủ33 bảng. |

Các giá trị kỹ thuật dưới đây là đề xuất triển khai của bộ thiết kế, không gán thành quyết định sản phẩm cũ. Đầu vào legacy thực tế chỉ xác minh tại ticket migration trên PG sở hữu; nguồn quyền và semantics đã chốt không chọn lại.

## 5. Kiến trúc và dữ liệu

Transport → service/ports → adapter PostgreSQL; service không import HTTP/pgx. Dùng package theo trách nhiệm skill catalog/skill request, không dựng framework tổng quát cho vài field. Composition root giữ năm auth +16OAP +10PCA. Canonical HTTP mới ở [OpenAPI](../api/manual-skills.openapi.json); SQL/transaction/DB tests ở [SQL plan](backend-manual-skills-sql-plan.md). Spec mới ghi planned; Swagger chỉ xuất operation khi dependency runtime thực sự có.

Khi code, dùng thư viện chuẩn Go/Browser và công cụ sẵn có: cookie/CSRF/gate, kiểm JSON strict và byte thực trong transport hiện hữu, transaction/rollback của PG adapter, resource abort/generation và primitives dialog/bảng/phân trang. Tách phần dùng chung nhỏ chỉ khi MSK thật sự dùng cùng contract; không dựng bộ CRUD/parser/validator/state framework hoặc ép catalog OAP đổi DTO để dùng chung. Typed payload chuẩn hóa một lần ở service rồi dùng cho hash/persist; adapter vẫn kiểm dữ liệu và quyền hiện hành dưới khóa. Comment giải thích lý do/ràng buộc theo quy ước chung. Reuse màn demo ở phần visual/primitives, không giữ demo store, Number(years) hoặc guard chặn cả reject khi target inactive.

Schema tiến thêm: danh mục giữ ID và normalize tên; official bỏ source và map approved_by_employee_id qua employees.user_id sang approved_by; không fake request để hợp thức hóa legacy. Request MANUAL lưu before/proposed typed JSON, explanation/external_evidence, expert_opinion từ quyết định, actor/time/status/version; chưa thêm FK đến CV/task chưa tồn tại. Migration tương lai mở source CV và FK thật; không lưu ID vào JSON hoặc cột vô ràng buộc cho feature chưa nối. Official không source/version/is_effective. Bảo toàn provenance legacy theo SQL plan; thiếu approval/time/mapping thì dừng phần upgrade đó.

## 6. Contract hành vi

### 6.1. Quyền và projection

Phiên thường: cả user/employee ACTIVE và không must_change_password. Roles là tập server hiện hành; title không cấp quyền. Tất cả actor_id từ session, user_id khác employee_id. Mọi mutation recheck dưới khóa.

- Catalog GET cho mọi phiên thường để chọn kỹ năng; mutation chỉ DIRECTOR. Không tự mở danh mục tổ chức HR cho Director.
- Official GET self (kể cả Admin/Director self); ID khác cùng404. Director xem current người khác chỉ từ chi tiết request có quyền, không API read-all official/profile.
- Requests scope=self cho mọi phiên thường; scope=review chỉ DIRECTOR. Không tự duyệt khi actor.user_id=request.requested_by, kể cả role tổ hợp. List chỉ nhận diện/code/vị trí/kỹ năng/status/ngày; detail gồm giải trình/before/proposed/current/decision và lịch sử cùng skill đọc qua list filter. Không email/phone/address/hash/token, raw audit hay raw JSONB. Tên/mã là hiện hành, không gọi snapshot lịch sử.
- Actor inactive/phiên lỗi401; limited403 PASSWORD_CHANGE_REQUIRED; thiếu role403 FORBIDDEN; resource ngoài quyền/missing404; lỗi DB503/500 đúng nguồn. Target employee inactive: approve409 TARGET_INACTIVE, reject vẫn được; target user locked/cờ limited không là điều kiện phiên của người duyệt và không tự activate target. Cùng nguyên tắc PCA, workflow CV cho kết thúc inactive không ghi official.

### 6.2. Input và cạnh tranh

Tên danh mục sau Go TrimSpace1..100code point Unicode; unique lower(trim) toàn cục mỗi loại, không bỏ dấu/NFC. Skill thuộc nhóm bắt buộc. Rename không đổi nghĩa kỹ năng đã dùng; UI nói rõ chỉ sửa cách viết/tổ chức, không có thao tác thay bản chất. expected_before={name,category_id} cho skill; category chỉ name. So raw before đọc từ server, không trim before; no-op200 không audit. Delete nhận cùng before,204 không JSON; FK còn official/request/skill chặn409 CATALOG_IN_USE. Không cascade/set null.

Submit: skill_id, expected_before_values (null hoặc đầy đủ proficiency_level/years_experience/note), proposed_values đầy đủ ba field, explanation bắt buộc, external_evidence optional text. Không nhận source/operation/status/employee_id/actor/version/schema_version/tệp. Với ADD hoặc note thực sự đổi, normalize note bằng Go TrimSpace, rỗng về null, tối đa2000 code point. Với UPDATE có proposed.note bằng chính xác before.note, giữ nguyên note đó, kể cả padding hoặc dài hơn2000; không biến việc đổi mức/năm thành sửa note. Before vẫn phải khớp official dưới khóa, nên client không thể dùng ngoại lệ này để chèn note dài khác. external_evidence/expert_opinion optional trim, trắng về null, tối đa2000; explanation và decision reason trim1..2000, reason bắt buộc cả approve/reject. UTF-8/NUL/duplicate JSON key/khóa lạ sai400. Before giữ nguyên legacy, chỉ validate shape/type; quá body thì dừng để review dữ liệu, không cắt before/note. Quy tắc note không đổi áp dụng cả validate, hash, JSONB và DTO/replay sau terminal; không validate lại theo official mới trước tra key.

years_experience nhận chuỗi decimal canonical **0..999.9 tối đa một chữ số lẻ hoặc null**, không float nhị phân; response cũng chuỗi một chữ số thập phân (0.0), before dùng đúng response. Đây là giới hạn NUMERIC(4,1), không tiêu chí xét mức. FE giữ text từ input, kiểm đúng mẫu/scale rồi thêm .0 khi cần; không qua Number/toFixed hoặc làm tròn để dữ liệu hợp lệ. Mức enum bốn tên; không enum số demo. proposed=before sau normalize →400 NO_CHANGES, không tạo NO_CHANGE giả; NO_CHANGE của luồng CV/terminal cũ không là approve thủ công. Client giữ before lúc mở form.

Submit key UUIDv4 lowercase, actor+SKILL_CHANGE_SUBMIT+key unique24h; SHA256 canonical typed payload sau validate, phân biệt null/0, không lệ thuộc thứ tự JSON. Khóa/tra key trước stale/pending: replay hợp lệ trả201 cùngrequest_id cùng projection hiện hành, không nhân notification; khác payload409 IDEMPOTENCY_CONFLICT; hết hạn409 IDEMPOTENCY_KEY_EXPIRED, không âm thầm tạo lại. Không cleanup records trong đợt này. Different key nhưng đang pending409 PENDING_REQUEST_EXISTS; cùng actor/skill/official serialize dưới employee lock, unique partial là hàng rào cuối.

Approve/reject nhận expected_request_version integer1..MaxInt32, reason và expert_opinion optional (nguồn/nội dung ngoài). Không sửa proposed/căn cứ bằng decision; lưu opinion một lần cùng terminal. Request luôn version1 trong đường mới của đợt này, vẫn kiểm version để tương thích bổ sung về sau và dữ liệu race có kiểm soát; không giả đã có API supplement. Recheck PENDING trước version; resolved409 REQUEST_ALREADY_RESOLVED, màn version cũ409 STALE_REQUEST, approve official khác before409 STALE_DATA, vẫn PENDING. Reject không cần before khớp, nhưng phải kiểm version/không self. No-op submit400; không approve để tự sửa trước/proposed. Commit unknown503 COMMIT_OUTCOME_UNKNOWN, không trả success/auto retry.

Mutation Origin/CSRF/strict JSON/Content-Type/application/json/body byte thực, không Content-Length vẫn kiểm. GET query allowlist, không duplicate param; page default1/size20 max100, IDs1..MaxInt32; q trim tối đa100 Unicode tìm literal %/_; count/items cùng REPEATABLE READ snapshot (kể cả page rỗng); sort requested_at/request_id DESC cho request, name/id ASC cho danh mục/official. Mọi response no-store, không Set-Cookie nghiệp vụ; giữ giới hạn auth/OAP hiện có và hạn đọc15s của server.

### 6.3. Thông báo

Reuse GET /api/notifications và POST /api/notifications/{id}/read hiện có, không endpoint mới: mở thêm SKILL_CHANGE_SUBMITTED/APPROVED/REJECTED, resource SKILL_CHANGE_REQUEST trong service validator, DTO enum canonical PCA và UI allowlist. Submit snapshot DIRECTOR đủ điều kiện, bỏ chính actor/loại trùng multiple roles; không có recipient vẫn lưu request để queue dùng được. Approve/reject một owner notification, recipient user_id; bị locked/limited không xóa lịch sử hoặc giả có phiên, quyền đọc sau vẫn do auth.

Snapshot recipient có recheck dưới cùng user/employee locks; role mới sau snapshot không hứa nhận thông báo cũ. Message chung không PII/level/note/reason; link allowlist/self hoặc review theo quyền hiện hành. Chỉ mở panel không read; GET tài nguyên thành công rồi mark own read; mất quyền vẫn cho read riêng, không đọc nội dung request. Không polling hoặc email.

### 6.4. Lỗi và luồng FE

401 xóa form/cache/PII. Khi403 mất DIRECTOR hoặc chuyển limited, xóa ngay list/detail/draft đặc quyền, abort GET và bỏ phản hồi cũ, kể cả me tiếp theo503; refresh me chỉ cập nhật navigation, không trì hoãn xóa. Chỉ tải lại tài nguyên sau xác thực/quyền hiện hành và GET mới thành công; own notification-read vẫn độc lập. Không replay POST;503/network không giả hết phiên. Sau submit/decision mất response: khóa gửi lại, GET list/detail đối soát; không suy failed chắc chắn. Refresh cùng phiên không remount draft. Logout/Back/đổi user bỏ phản hồi cũ, không mock fallback. Không lưu payload/key/PII vào localStorage hoặc URL.

## 7. Brief UI và URL

Reuse PersonalSkills/SkillRequestForm/SkillRequestDetails, SkillRequests/SkillReviewDialog và SkillCatalogs/SkillEntryForm; adapter server async riêng, không tra người thật vào demo fixture. Không thiết kế dashboard mới.

- /skills: official + lịch sử self, query request_id/page/page_size/status; default status=all. request_id ID chỉ sau GETself. Chọn skill bằng catalog picker tải đủ các trang, không chỉ20 mục. Trước mở form, GET official self với skill_id, page1: total0/1 là snapshot để dựng before, không suy absence từ trang official đang hiển thị. Pagination official riêng trong màn, page/page_size URL của /skills dùng cho lịch sử; null years hiển thị Chưa cung cấp.
- /skill-reviews: DIRECTOR, default PENDING; status=all/APPROVED/REJECTED; query employee_id/skill_id cho lịch sử cùng người/kỹ năng. request_id cùng GET review. Giữ filter/page khi đóng detail, lịch sử không mở dữ liệu ngoài quyền.
- /skill-catalogs/categories và /skill-catalogs/skills: DIRECTOR menu; q/page/page_size, skill thêm category_id filter. GET dùng cho self picker qua adapter, không vì có GET mà mở menu mutation.
- Menu: mọi phiên thường Kỹ năng của tôi; DIRECTOR thêm Xét duyệt kỹ năng/Danh mục kỹ năng; tổ hợp cộng mục theo server. Home HR vẫn /organization/departments, các user khác /profile; limited luôn /change-password. URL chưa nối báo giới hạn, không prototype.
- Biểu mẫu có labels/hints/hướng dẫn bốn mức, error focus; dialog Escape/outside/đóng/dirty confirm theo convention. Không hiển thị action supplement/withdraw/upload chưa có. Responsive375px, bảng cuộn vùng riêng, focus trả đúng trigger/heading. Browser phải kiểm thật, không lấy parser thay visual.

## 8. Kiểm chứng và rủi ro

[SQL plan](backend-manual-skills-sql-plan.md) MSK-DB01–12 nối tới task; [tickets](backend-manual-skills-tickets.md) ghi check đúng tầng. Go/PG/FE/contract/browser khi triển khai, không chạy từ đợt thiết kế. Rủi ro: legacy thiếu actor/time hoặc inactive skill; ABA official chỉ so business values theo quyết định bỏ version; recipient snapshot không hứa phantom; nguồn CV/task chưa nối; deadline sát03/12. Không bỏ khóa/quyền để kịp lịch.

FR-016/017/023/027 là mục tiêu nghiệm thu theo AC, không ghi Done trước runtime/merge; FR-024/026 Partial do chưa tệp/task; FR-059/060 chỉ sự kiện kỹ năng và phần đã nối. FR-025/028/CV tiếp nối. Nếu chậm, bàn giao catalog + official read đã kiểm và ghi phần submit/decision chưa đạt, không demo success giả. Không thu nhỏ mẫu số62FR.

## 9. WBS và kế hoạch nhận việc

| WBS | Ticket | Đầu ra | Consumer |
|---|---|---|---|
| MSK-01 | DATN-50 / MSK-T01 | Bàn giao contract, SQL plan và brief kỹ năng thủ công | MSK-02 |
| MSK-02 | DATN-51 / MSK-T02 | Triển khai migration kỹ năng và guard legacy trên PostgreSQL test | MSK-03, MSK-05 |
| MSK-03 | DATN-52 / MSK-T03 | Triển khai CRUD API nhóm kỹ năng và kỹ năng theo quyền Giám đốc | MSK-04, MSK-06 |
| MSK-04 | DATN-53 / MSK-T04 | Nối UI Giám đốc quản trị danh mục kỹ năng | MSK-10, MSK-12 |
| MSK-05 | DATN-54 / MSK-T05 | Mở rộng thông báo server cho sự kiện kỹ năng và giữ tương thích PCA | MSK-06, MSK-11 |
| MSK-06 | DATN-55 / MSK-T06 | Triển khai đọc năng lực, gửi và tra cứu đề nghị kỹ năng chính chủ | MSK-07, MSK-08 |
| MSK-07 | DATN-56 / MSK-T07 | Triển khai transaction Giám đốc chấp thuận hoặc từ chối kỹ năng | MSK-08 |
| MSK-08 | DATN-57 / MSK-T08 | Tích hợp sáu API năng lực và đề nghị kỹ năng với PostgreSQL | MSK-09, MSK-10 |
| MSK-09 | DATN-58 / MSK-T09 | Nối UI năng lực chính thức, gửi đề nghị và lịch sử cá nhân | MSK-11, MSK-12 |
| MSK-10 | DATN-59 / MSK-T10 | Nối UI Giám đốc đối chiếu và quyết định kỹ năng | MSK-11, MSK-12 |
| MSK-11 | DATN-60 / MSK-T11 | Nối thông báo kỹ năng tới đúng màn cá nhân hoặc Giám đốc | MSK-12 |
| MSK-12 | DATN-61 / MSK-T12 | Kiểm xuyên luồng kỹ năng và bàn giao demo ngày 03/12 | Bàn giao D04 |

Rank là thứ tự ưu tiên nhận việc; Blocks xác định đầu vào bắt buộc. Từ đợt MSK, triển khai tối đa hai task độc lập đồng thời theo [delivery-plan](../ai/delivery-plan.md#triển-khai-song-song-từ-đợt-msk): mỗi task có branch/worktree riêng, DB/browser dùng môi trường chung được kiểm luân phiên.

Mốc dự kiến: 27–28/11 bàn giao thiết kế, 29/11–01/12 triển khai, 02/12 regression và tập demo, 03/12 báo cáo. Quân review/kiểm chứng phần quan trọng; checkpoint Go tự viết được xác định trong brief nhận việc.

## 10. Bàn giao triển khai

### Nguồn và Jira

Nguồn bàn giao DATN-50 là [commit merge `1d84eefde6519b27559ed00974383eeb9f4d4aed`](https://github.com/mingquanjp/prj-manage-system/commit/1d84eefde6519b27559ed00974383eeb9f4d4aed) của [PR #66](https://github.com/mingquanjp/prj-manage-system/pull/66). Proposal, OpenAPI MSK/PCA, SQL plan, tickets và brief UI tại commit này khớp bản đã review `7c4b2dc`. Các bổ sung bàn giao bên dưới không thay contract hoặc quyết định MSK-A01–A08.

| Nguồn chuẩn | Phần áp dụng | Task tiếp nhận |
|---|---|---|
| Proposal này | Mục 4–8: lựa chọn đã chốt, hành vi và URL; mục 9: WBS → ticket | T02–T12 |
| [OpenAPI MSK](../api/manual-skills.openapi.json) | 14 thao tác planned: 8 danh mục và 6 năng lực/đề nghị | T03/T06–T10 |
| [OpenAPI PCA](../api/profile-change-approval.openapi.json) | Hai API thông báo hiện có và các giá trị mở rộng kỹ năng planned | T05/T11 |
| [SQL/test plan](backend-manual-skills-sql-plan.md) | Mục 1–6: nâng cấp/transaction; mục 7–8: ma trận và lệnh kiểm | T02/T03/T05–T08/T12 |
| [Brief MSK tại web README](../../web/README.md#brief-msk--kỹ-năng-thủ-công-cho-0312) | Tái sử dụng UI, URL, phân tách server/demo và kiểm theo phiên; chi tiết tại mục 7 proposal | T04/T09–T12 |
| [Tickets](backend-manual-skills-tickets.md) | 12 task, phạm vi/AC và 19 phụ thuộc | T01–T12 |

Snapshot dùng khi nhập Jira: commit `496054cef48c6f446a86b1789680c6a0a73e36ca`. Ticket giữ nguyên nội dung nhập; proposal hiện hành rút gọn phần bàn giao. [Hướng dẫn importer](../../scripts/jira/README-msk.md) ghi cách đối soát snapshot mà không đổi journal hoặc ghi đè Jira.

Quân xác nhận bộ ticket đã lên Jira ngày 27/11/2026. Mapping từ journal local: Epic [DATN-49](https://datn-mingquanjp.atlassian.net/browse/DATN-49); MSK-T01–T12 lần lượt là DATN-50–61. Bắt đầu bằng [DATN-50](https://datn-mingquanjp.atlassian.net/browse/DATN-50) bàn giao thiết kế, sau đó [DATN-51](https://datn-mingquanjp.atlassian.net/browse/DATN-51) migration.

### Review và kiểm chứng thiết kế

Proposal, OpenAPI MSK/PCA, SQL/test plan, brief UI và tickets đã được đối chiếu với workflow kỹ năng, FR-016–028, quyền và dictionary áp dụng. Các lượt review độc lập về nghiệp vụ, HTTP–SQL và UI/phụ thuộc đã xử lý các finding; nội dung cuối giữ đúng note legacy, lookup official ngoài trang, lịch sử quyết định, input năm và quyền hiện hành.

Lệnh kiểm từ gốc repo:

```sh
node scripts/check-feature-plan.mjs --proposal docs/proposals/backend-manual-skills.md --tickets docs/proposals/backend-manual-skills-tickets.md --openapi docs/api/manual-skills.openapi.json --sql-plan docs/proposals/backend-manual-skills-sql-plan.md --source .claude/docs/skill_approval_workflow.md --source docs/product/functional-requirements.md
node --test scripts/check-manual-skills-contract.test.mjs scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs scripts/check-feature-plan.test.mjs
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s scripts/jira -p 'test_*.py' -v
node scripts/check-harness.mjs
git diff --check
```

Kết quả chuẩn bị ngày 27/11: 61 tests contract/harness/feature-plan và 86 tests Jira offline đạt; kiểm cấu trúc đủ 14 operation planned, 12 WBS/task, 19 cạnh và một sequence Mermaid. Sơ đồ được preview bằng bundle Mermaid local. Kiểm tĩnh/HTTP giả không chứng minh API, migration hay UI MSK đã chạy; MSK-DB01–12 là các ca phải kiểm khi triển khai. SQL plan là thiết kế, không phải migration thực thi độc lập.

Giới hạn nguồn: dictionary 5B còn link lịch sử `database-step3b-effort-profile-skills-cv-2026-10-13.md` không có trong repo. Nội dung dictionary áp dụng đã được đọc; kiểm toàn bộ link của tài liệu này còn thiếu, không thuộc kết quả kiểm nguồn MSK ở trên.

### Kiểm bàn giao DATN-50

Ngày 27/11/2026, tại gốc worktree từ `main 1d84eef`, macOS arm64 / Node v24.15.0:

```sh
node --test scripts/check-manual-skills-contract.test.mjs scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs
node scripts/check-harness.mjs
git diff --check
```

Kết quả: 48/48 tests đạt, không skip; harness kiểm kê 492 nguồn và diff check đạt. Lệnh `check-feature-plan.mjs` đầy đủ ở trên cũng đạt: 14 operation planned, 12 WBS/task, 19 phụ thuộc và một Mermaid parse được. Review độc lập từ PR #66 được dùng lại cho phần thiết kế không đổi; phần bổ sung mục 9–10 đã đối chiếu với tickets và SQL plan. Không có kiểm PostgreSQL, HTTP hoặc browser MSK trong DATN-50; các ca MSK-DB01–12 vẫn thuộc task triển khai tương ứng.

### Đầu vào cho các task sau

DATN-51 / T02 tiếp nhận các mục sau để xác minh trên PostgreSQL test sở hữu. DATN-50 chỉ đối chiếu thiết kế; chưa xác nhận dữ liệu hoặc trạng thái DB thực tế.

| Đầu vào cần xác minh | Nguồn áp dụng | Điều kiện dừng |
|---|---|---|
| PG18, owner/system-ID; head dự kiến 000006, dirty/drift và dependency | SQL plan mục 1 | Chưa xác minh môi trường hoặc metadata lệch thiết kế; không force version |
| Tên nhóm/kỹ năng và `skills.status` legacy | SQL plan mục 1/3 | Tên trắng, padding, trùng sau normalize hoặc skill INACTIVE chưa có mapping được review; không tự sửa/gộp/xóa |
| `employee_skills.source`, người/thời điểm duyệt và liên kết `employees.user_id` | SQL plan mục 1–2 | Source/enum không hợp lệ, thiếu actor/time hoặc linkage không đủ; không suy user_id từ employee_id, không yêu cầu approver lịch sử còn quyền hiện hành |
| Actor thực hiện import khi có official cần lưu provenance | SQL plan mục 2 | Chưa xác minh user thực hiện chuyển đổi; fixture phải dùng actor tổng hợp tường minh, không mặc định actor import là approver cũ |

T02 phải kiểm lại guard dưới khóa, bảo toàn dữ liệu và down/dirty theo SQL plan mục 2/6; thu bằng chứng MSK-DB01–03 tại mục 7. Không dùng down thay backup hoặc gọi dữ liệu fixture là mapping môi trường đích.

- T09/T10 dùng URL/semantics input/version tại mục 6–7, giữ dữ liệu theo phiên và dọn ngay khi mất quyền.
- T12 thu bằng chứng DB/browser, race hai adapter thật, cập nhật runbook/FR coverage và kịch bản báo cáo.

DATN-51 / MSK-T02 triển khai migration `000007_manual_skills`, guard dùng chung trước dirty/trong DDL và fixture legacy với actor import tường minh. [Runbook PostgreSQL MSK](../../scripts/authdb/README.md#datn-51--msk-t02--migration-kỹ-năng) giữ lệnh, ma trận MSK-DB01–03, kết quả thực chạy và giới hạn cho DATN-52/DATN-54. Không thay snapshot thiết kế hoặc coi fixture là mapping dữ liệu đích.

CV/tệp/task, bổ sung/rút pending và công thức phân công tiếp tục ở đợt riêng.

DATN-52 / MSK-T03 cung cấp tám API danh mục cho MSK-T04/T06. [Runbook danh mục](../../scripts/authdb/README.md#datn-52--msk-t03--api-danh-mục-kỹ-năng) giữ DTO, giao thức khóa, fixture và kết quả HTTP/PostgreSQL; race xóa kỹ năng–submit được kiểm tiếp bằng adapter submit thật tại MSK-T06.
