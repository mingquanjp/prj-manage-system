# SQL và kế hoạch kiểm PostgreSQL — kỹ năng thủ công

Ngày27/11/2026. Thiết kế MSK, **không thực thi SQL trong đợt chuẩn bị**. Nghiệp vụ/URL tại [proposal](backend-manual-skills.md); HTTP tại [canonical](../api/manual-skills.openapi.json). Head đọc từ main8fc3ebd là000006; số migration dự kiến000007 phải kiểm lại khi code. Không sửa000001–000006 hoặc áp Neon.

## 1. Preflight, ownership và drift

Dùng runner [authdb](../../scripts/authdb/README.md) kiểm PG18/owner/system-ID, schema_migrations head/dirty và metadata bảng/cột/index/FK; không fallback DATABASE_URL/.env. Không DB suite song song. Empty upgrade đi1→6→mới; không dựng thẳng schema đích rồi gọi là migration.

Legacy phải kiểm ít nhất: skills.status, source official, approved_by_employee_id/approved_at, approval linkage, trùng lower(trim Unicode) tên nhóm/kỹ năng, names padding/trắng, dependencies/view/trigger/FK không dự kiến. Bảng requests kỹ năng đã tồn tại hoặc column drift phải dừng. Guard preflight trước dirty, rồi guard lại dưới table lock trong DDL để chặn race sau preflight. Lỗi DDL rollback business state nhưng schema_migrations có thể dirty theo driver; không force version.

Chỉ nâng tự động skill status ACTIVE. INACTIVE có ý nghĩa cũ không tự biến thành đang dùng: dừng phần upgrade, xin mapping có review. Name padding/blank/duplicate không tự trim/gộp/xóa legacy. Official thiếu actor/time/user-link/enum hoặc source không MANUAL/CV phải dừng. Không buộc historical approver còn role DIRECTOR/ACTIVE ở thời điểm migrate; giữ sự kiện lịch sử, không cấp quyền hiện hành.

## 2. Biến đổi dữ liệu bảo toàn

- Khóa các bảng đang đổi theo thứ tự users, employees, skill_categories, skills, employee_skills, audit_events; transaction lock_timeout5s, statement_timeout30s. Không giữ transaction cho thao tác người dùng.
- `approved_by = employees.user_id WHERE employees.employee_id = old.approved_by_employee_id`, kiểm map1:1 và user tồn tại; không gán cùng số ID. Không tự lấy actor chạy migration, created_at hoặc now làm người/ngày duyệt lịch sử.
- Giữ official ID, employee_id/skill_id, level/years/note, approved_at, created_at/updated_at. FK employee chuyển CASCADE sang NO ACTION, không đổi ID. Approved_by INT NOT NULL FKusers, không dùng FKemployee mới.
- Trước bỏ source và approval legacy: append audit operation SKILL_LEGACY_IMPORTED, entity_type EMPLOYEE_SKILL, entity_id dòng official, actor_id là user người thực hiện chuyển đổi được xác minh trong mapping, occurred_at **thời điểm import**; before_values giữ source/approved_by_employee_id/approved_at nguyên gốc, after_values giữ mapped approved_by. Đây là sự kiện chuyển đổi, không giả approver lịch sử đã chạy migration hoặc thời điểm import là lần phê duyệt mới. Actor vận hành và actor approval là hai vai trò riêng; approved_by vẫn map đúng approver gốc. Nếu có dữ liệu cần archive nhưng chưa có actor chuyển đổi được xác minh, preflight phải dừng; fixture PG dùng actor tổng hợp tường minh, không default ngầm hoặc mạo danh người dùng. Không tạo request giả; không dùng audit này làm official source.
- Bỏ employee_skills.source, approved_by_employee_id và skills.status/index status sau guard. Không drop enum shared skill_source/skill_status đang có consumer; loại bỏ column không đồng nghĩa xóa type. Các source từ seed/fixture/localdev phải cập nhật head, fixture v1–v6 giữ nguyên lịch sử.
- users/employees/auth_sessions/user_roles/OAP catalogs/profile requests/idempotency/PCA audit/notifications giữ nguyên; không DROP/recreate notifications hoặc idempotency.

DDL minh họa phần approval, chưa là migration chạy được độc lập (guard/archive trước, tên FK đọc metadata):

```sql
ALTER TABLE public.employee_skills ADD COLUMN approved_by int;
UPDATE public.employee_skills es
SET approved_by = e.user_id
FROM public.employees e
WHERE e.employee_id = es.approved_by_employee_id;
ALTER TABLE public.employee_skills
  ALTER COLUMN approved_by SET NOT NULL,
  ALTER COLUMN approved_at SET NOT NULL,
  ADD CONSTRAINT fk_msk_approval_user FOREIGN KEY (approved_by) REFERENCES public.users(user_id);
```

Không viết SQL cụ thể DROPconstraint bằng tên đoán; T02 phải đối chiếu metadata từ migration lịch sử và test drift. Không nhận đoạn trên là complete upgrade hoặc runtime proof.

## 3. Danh mục và official

Tên1..100 code point đã trim Unicode White_Space theo Go (U+0009..000D,0020,0085,00A0,1680,2000..200A,2028/2029/202F,205F,3000). CHECK không whitespace đầu/cuối/trắng; unique expression lower(name), global từng loại; PG collation/Go lower parity cần fixture tiếng Việt/Unicode. Không ASCII btrim thay TrimSpace; không NFC/bỏ dấu. Legacy khác canonical dừng mapping, không UPDATEâm thầm.

Giữ UNIQUE cũ names khi thêm expression index nếu tên hiện tại chưa chuẩn. Index skills(category_id), employee_skills(skill_id), UNIQUE(employee_id,skill_id) giữ; FK category/skill NO ACTION. Với tên used, chỉ cho sửa cách viết/tổ chức theo contract; thay đổi nghĩa không được mô hình tự suy hoặc tự gộp. UI/đặc tả phải chỉ rõ tác động đổi nhóm; không sửa ID/official qua rename.

Query SELECT projection cột tường minh; q literal bằng escaped LIKE hoặc strpos trên lower, bound parameters; count/items/actor/roles trong một REPEATABLE READ read-only snapshot; sort name,id hoặc requested_at DESC,idDESC, page rỗng vẫn total. Không nhiều rows do user_roles join: EXISTS/aggregation đúng projection. official GET self nhận skill_id optional; lookup form page1 trả total0/1 cùng snapshot, không coi thiếu dòng ở một trang là chưa có skill. official GET self trả employee_id và danh sách có ID/skill/category/value/approval actor ID/time, không email/private profile. Note legacy không trim/cắt khi GET; không áp max mới lên before.

## 4. Request MANUAL — schema tiến thêm

Không dựng CV/task tables rỗng cho đủ danh sách. Đợt này tạo employee_skill_change_requests với:

| Cột | Ràng buộc đợt này |
|---|---|
| request_id | INT identity PK |
| employee_id, skill_id | INT NOT NULL FK NO ACTION employees/skills |
| operation | TEXT ADD/UPDATE; server suy theo before |
| source | TEXT NOT NULL MANUAL; CHECK MANUAL, future migration mở CV cùng FK thật |
| before_values | JSONB nullable; NULL chỉ ADD |
| proposed_values | JSONB NOT NULL, đầy đủ typed values |
| explanation | TEXT NOT NULL,1..2000 code point sau trim |
| external_evidence | TEXT nullable1..2000 sau normalize, văn bản không fetch URL |
| expert_opinion | TEXT nullable1..2000; chỉ terminal decision từ Director |
| requested_by/requested_at | INT FKusers NOT NULL; TIMESTAMPTZ NOT NULL server |
| status | TEXT PENDING/APPROVED/REJECTED trong đường mới |
| resolved_by/resolved_at/decision_reason | FKusers/time/reason terminal; pending đều NULL |
| version | INT NOT NULL DEFAULT1 CHECK>0; không client chọn mới |

CHECK source/operation/status/JSON shape/decision/times, schema_version=1 trong JSONB; input HTTP không nhận schema_version. HTTP Values có đúng ba khóa nghiệp vụ: proficiency_level, years_experience và note. JSONB persisted có đúng bốn khóa: ba khóa nghiệp vụ đó cộng schema_version=1. Riêng before SQL NULL của ADD vẫn là SQL NULL, không bọc metadata. proficiency_level là enum, years_experience là canonical decimal-string hoặc JSON null, note là string hoặc null. JSONnull khác SQLNULL object: before SQLNULL biểu thị absence, không lưu object toàn NULL. Validators trả boolean rõ, bọc `(predicate) IS TRUE` không để SQLNULL vượt CHECK. Chặn JSON object rỗng/keys lạ/type sai/NUL (PGtext đã cấm)/số lẻ vượtscale. Before note legacy không giới hạn mới. Proposed note phải normalize/max2000, **hoặc** là UPDATE có note JSON bằng đúng note trong before: giữ nguyên raw kể cả padding/dài hơn2000. CHECK xét hai nhánh này cùng before/proposed, không chỉ validate proposed độc lập; adapter xác minh before với official dưới khóa trước ghi mới. Decoder persisted/DTO/replay cũng chấp nhận nhánh note giữ nguyên, kể cả sau request terminal. years canonical string regex `^(0|[1-9][0-9]{0,2})\.[0-9]$`, range0.0..999.9; HTTP cho integer-string rồi normalize thành một chữ số thập phân. Không round để hợp lệ.

ADD beforeNULL, UPDATE before object; before khác proposed sau canonicalization. Không thêm official placeholder pending. Pending expert_opinion/resolver/time/reasonNULL. APPROVED/REJECTED đủ reason/actor/time; decision actor khác requested_by; resolved_at>=requested_at; version>0. Kiểm requested_by liên kết employee trong adapter vì FK đơn không đủ. Mọi writer tuân terminal bất biến; test port+HTTP không cho sửa terminal, không tuyên bố DBA SQL bị cấm bởi API.

```sql
CREATE UNIQUE INDEX uq_msk_pending
ON public.employee_skill_change_requests(employee_id, skill_id)
WHERE status = 'PENDING';
CREATE INDEX idx_msk_owner
ON public.employee_skill_change_requests(employee_id, requested_at DESC, request_id DESC);
CREATE INDEX idx_msk_review
ON public.employee_skill_change_requests(status, requested_at DESC, request_id DESC);
CREATE INDEX idx_msk_employee_skill
ON public.employee_skill_change_requests(employee_id,skill_id,requested_at DESC,request_id DESC);
```

Rút/bổ sung/NO_CHANGE/REMOVE/CV chưa writer/schema đợt này; migration tiếp theo mở ràng buộc khi có use case, không sửa migration lịch sử. Giữ request.version hiện tại và protocol để không thiết kế lại bảo vệ màn cũ.

## 5. Khóa và transaction nghiệp vụ

Dùng context10s cho mutation như composition hiện có; rollback cleanup5s. Không auto retry mutation/commit. Reuse đúng giao thức PCA: snapshot tập participants trước khi khóa, không khóa actor riêng rồi mới bổ sung ID thấp hơn.

**Thứ tự chung**: users tăng user_id → employees tăng employee_id → skill_categories tăng ID → skills tăngID → request → official pair. Các writer auth/status/role/HR/PCA hiện hành và writer mới phải cùng users→employees, không lấy userlock sau employee. Re-read linkage sau lock; đổi linkage thì từ chối, không mở vòng thu thập recipient mới sau snapshot.

- Catalog mutation: participants actor; old/new category IDs snapshot, khóa categories theo ID, rồi skill nếu có; recheck skill category/before dưới khóa, stale trả409 thay vì đổi thứ tự khóa. Create skill FK KEY SHARE category. Delete category không lock child skill sau khi phát hiện FK; DELETE FK chặn và map CATALOG_IN_USE. Catalogdelete actor có employment position/category tổ chức là miền khác, không đọc official để cấp quyền.
- Submit: snapshot actor + DIRECTOR đủ điều kiện, lock toàn bộ users/employees tăngID, recheck actor/status/flag/self và recipient role; bỏ actor khỏi recipient. Tra idempotency trước official stale/pending, resource replay vẫn kiểm quyền. Khóa category KEY SHARE rồi skill KEY SHARE; actor employee FOR UPDATE serialize absence của official pair. Đọc official và pending cùng transaction; compare before, insert request version1, notifications perrecipient, kết quả key đầy đủ. Partial UQ là backstop, lỗi conflict rollback. Không khóa official row không tồn tại làm bảo vệ duy nhất.
- Decision: đọc request để biết subject; snapshot actor+target user/employee (owner nhậnnotification), lock users/employees sorted, recheck link và DIRECTOR/nonself. Category/skill KEY SHARE, request FOR UPDATE; PENDING rồi version, approve targetemployeeACTIVE và currentvalues=before, update/insert official cùng terminal/ownernotification. Reject bỏ stale official/targetactive check, vẫn version/nonself; terminal remains immutable. Request đã lưu đủ before/proposed/actor/time/reason/opinion, không sao chép submission/decision sang audit_events theo dictionary mục20. Audit chỉ dành catalog và import legacy có mục đích riêng. Catalog không thể mất tham chiếu khi request tồn tại. READ COMMITTED dưới lock đủ cho mutation, reads dùng RR.
- Role snapshot recipients chỉ đảm bảo người đã trong tập còn eligible khi ghi; role grant sau snapshot không bắt buộc nhận. Owner bị locked/limited không giả session, notification vẫn gắn đúng user_id. Không cùng user nhiều role tạo nhiều notification.

Official before so đầy đủ level/years/note kể cả absence; years canonical PG numeric không float. ABA cùng businessvalues được xem như giống theo quyết định không official.version; không hứa phát hiện mọi lịch sử thay đổi. Requestversion mismatch giữPENDING. Lỗi write/notification rollback request/official/decision/key; lỗi audit rollback catalog/import trong transaction tương ứng. Sequence có gaps không là halfwrite. Commit chưa rõ khôngsuccess, không autoPOST/retry hoặc rollback giả; GET có quyền để đối soát.

## 6. Down, dirty và bảo toàn

Down chỉ DBcase hy sinh, không backup. Chỉ cho down khi requests rỗng, không keyoperation SKILL_CHANGE_SUBMIT, notification resource SKILL_CHANGE_REQUEST, audit nghiệp vụ MSK hoặc official/import archive; bảng employee_skills phải rỗng. Điều này cố ý từ chối legacyimport không phục hồi được source gốc. Kiểm dưới locks trong BEGIN trước bất kỳDDL; không xóa rows để down thành công.

Down hợp lệ: bỏ request/functions/index mới, restore source MANUAL/approved_by_employee_idnullable/skills.statusACTIVE/index trên DB trống tương ứng; giữ original catalogIDs/timestamps. Vì mất status legacy đã bỏ, down cũng từ chối nếu skills còn rows; không tự coi ACTIVE reconstructed là dữ liệu cũ. Kiểm columns/index/FK trở vềhead6, down/up một lượt. Lỗi DDL giữa chừng rollback schema/data; metadata dirty đo theo driver riêng. Không force version hoặc gọi down là backup. Backup/restore dữ liệu thật cần kế hoạch riêng trước vận hành.

## 7. Ma trận DB và giới hạn

| Ca | Bằng chứng phải thu ở implementation | Task |
|---|---|---|
| MSK-DB01 | Nâng DB rỗng qua lịch sử đến head mới; constraints/index/FK và auth/OAP/PCA giữ nguyên | T02 |
| MSK-DB02 | Legacy user_id khác employee_id; map approver và lưu nguồn cũ; thiếu actor/time, INACTIVE, duplicate hoặc drift phải dừng | T02 |
| MSK-DB03 | Down rỗng thành công, down có dữ liệu từ chối; lỗi DDL và metadata dirty đúng driver | T02 |
| MSK-DB04 | CRUD hai danh mục; biên tên 100/101, unique race, tìm literal và count/items cùng snapshot | T03 |
| MSK-DB05 | Xóa nhóm cạnh tranh tạo skill; xóa skill cạnh tranh submit: dùng hai adapter thật và barrier, không orphan | T03/T06 |
| MSK-DB06 | Submit ADD/UPDATE, before/absence, năm NULL khác0, no-op; legacy note>2000/padding không đổi giữ raw khi chỉ đổi level/years; note thực đổi vẫn normalize/max, before giả bị stale; pending không đổi official | T06 |
| MSK-DB07 | Replay key, payload khác, key hết hạn, replay sau terminal/mất quyền; key/request/notification cùng commit | T06 |
| MSK-DB08 | Approve/approve và approve/reject; version/before stale, không tự duyệt, target inactive/locked | T07 |
| MSK-DB09 | Race quyền/cờ/linkage và recipient; nhiều role không nhân notification, không Director vẫn có queue | T06/T07 |
| MSK-DB10 | Official/terminal/notification nguyên tử; lỗi từng bước và commit chưa rõ; không sinh audit lặp submission/decision, note legacy giữ nguyên qua approve/GET/replay terminal | T07/T08 |
| MSK-DB11 | HTTP+PG14 operation; quyền, biên body8/64KiB, Unicode/duplicate key/DTO/no-store/không cookie mới | T03/T05/T08 |
| MSK-DB12 | Browser catalog→self→Director→official/history/notification, reload/logout/Back; stale, mạng và quyền | T12 |

Không coi barrier writer SQL là adapter use case khác đã đạt; nếu chưa có đối tác adapter thật, ghi phần race còn mở và khép tạiT12. Seed/app fixture theo head mới; historicalmigration tests giữ schema version đúng mốc. Fault là lỗi dựng có kiểm soát, không suy mọi outage/mất mạng vật lý. Không chứng minh toàn CV/task từ chỉ nguồn MANUAL.

## 8. Lệnh và bàn giao

Khi nhận ticket code Go: `go vet ./...`, `go test ./... -count=1`, `go build ./...`; sau `python3 scripts/authdb/manage.py check`, chạy tuần tự `python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./migrations ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m`. Bổ sung package thực khi có. FE `pnpm check` + browser PG thật, tách server/demo và clear PII. Ghi lệnh/file/kết quả/môi trường/cleanup/chưa kiểm; không secret/PIIthật.

Đợt thiết kế chỉ chạy contract/harness/feature-plan/diff tĩnh; mọi MSK-DB ở trên là yêu cầu kiểm **chưa chạy**, không gắnPASS từ parser. Cổng tĩnh cụ thể xem proposal phần10 và tickets.
