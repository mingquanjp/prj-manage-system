# Kế hoạch SQL và PostgreSQL test — Tổ chức, tài khoản, hồ sơ

Ngày 15/11/2026. **Thiết kế để review, chưa có migration mới hoặc SQL nào được chạy.** Phạm vi theo [proposal](backend-organization-accounts-profile.md), [OpenAPI](../api/organization-accounts-profile.openapi.json), [dictionary](../design/database-step5b-supporting-dictionary-2026-10-14.md) mục 1–5.1/20 và [mapping](../design/database-step5c-migration-map-2026-10-14.md). Không thay nghiệp vụ, gửi thư, sửa 000001–000003 hoặc áp Neon. Đây không phải script triển khai có thể copy chạy trực tiếp.

## 1. Nền thực tế và thứ tự chuyển đổi

Repo hiện có head 000003: departments/positions, employees.seniority_level, users.username NOT NULL UNIQUE, users.system_role legacy bắt buộc; 000002 thêm must_change_password, 000003 thêm user_roles/auth_sessions. Kiểm lại head và drift khi bắt đầu; không đặt số file trước. Không lấy kiểm kê Neon 06/11 làm trạng thái đích hôm nay.

| Đợt | Đầu ra DDL/mapping | Điều kiện trước thực thi | Consumer |
|---|---|---|---|
| M1 — F1/F3 | job_levels; employees.job_level_id/address; tên danh mục/index; audit_events; mapping cấp bậc có xác nhận | OAP-01/OAP-03, dữ liệu preflight không mâu thuẫn; mapping giá trị cụ thể được review | OAP-04–09; auth vẫn hoạt động |
| M2 — F2 | Sequence mã nhân viên, uniqueness email chuẩn hóa, DROP username | OAP-02/OAP-10, cập nhật code/seed/fixture/runner không còn cần cột ở schema đích; backup và kế hoạch dừng consumer | OAP-10–14 |

M1/M2 là nhóm thay đổi, không ép mỗi nhóm một file. DDL/mapping tách khi cần bảo toàn/review; mỗi file có timeout/transaction tương thích driver. Không IF NOT EXISTS để che drift, force version hoặc sửa schema_migrations. Dừng các consumer có thể ghi trong lúc chuyển đổi; không hứa zero downtime với DROP cột. Test bootstrap đường lịch sử phải chạy nguyên 000001 → 000002 → 000003 trước migration mới.

## 2. Preflight và đầu ra cần review

Chỉ chạy về sau trên DB test sở hữu qua guard hiện hành; đích thật là phạm vi khác. Kiểm host/port/database/owner/system identifier/PG major theo [runner](../../scripts/authdb/README.md), không fallback DATABASE_URL hoặc đọc .env thật. Các lệnh ALTER/cleanup chỉ nhận handle case được guard, không URL tùy ý.

- Đọc metadata: migration head/dirty; tên/kiểu/null/default/identity/enum/index/FK/trigger và phụ thuộc username, bảng đích tồn tại trái dự kiến; dừng nếu drift. Kiểm schema public và search_path an toàn.
- Danh mục: tên trắng, độ dài, duplicate sau chuẩn hóa A03; không tự xóa/gộp mục hoặc đổi ID để làm migration xanh. Rà cả dữ liệu dẫn chiếu.
- Cấp bậc: thống kê các giá trị seniority_level, rỗng/trắng và số hồ sơ; bảng mapping được review có cột giá trị legacy → job_level_id/tên mục được HR xác nhận hoặc chưa xác minh. Không suy từ position, không tạo Intern/Junior/Senior mặc định.
- Email: phát hiện NULL/trắng/sai kiểu và email đơn không hợp lệ (gồm display-name/list/CRLF), trùng theo lower(btrim(email)), kể cả user LOCKED/DISABLED và employee INACTIVE/RESIGNED; dừng, không LIMIT 1 hoặc tự viết lại email legacy. Kiểm input normalization tương đương auth (Go strings.ToLower/TrimSpace) trên fixture Unicode/case/space; nếu DB collation/trim cho kết quả lệch thì dừng để review quy tắc, không tự đổi auth.
- Mã: kiểm unique/độ dài, mã legacy E001/EMP001 và các mã EMP có phần số; không renumber. Dừng nếu giá trị vượt dải hoặc collision không có phương án. Không log email/name/hash/username; output review chỉ metadata/count và fixture tổng hợp. Mapping dữ liệu thật dùng kênh được phép, không commit PII.
- Username: kiểm views/functions/triggers/index, SQL ghi/đọc trong repo và các consumer được khai báo ở đích. Không DROP CASCADE; không kết luận hết phụ thuộc chỉ từ rg trong repo.
- Snapshot trước nâng cấp: số dòng, ID/liên kết/FK, giá trị các trường còn bảo toàn và bộ quyền/phiên. So sánh trong test, không in secret/hash/PII ra evidence. Username sẽ bị loại có chủ đích, không nằm trong cam kết bảo toàn sau DROP.

**Chưa có dữ liệu đích để chốt mapping cấp bậc hoặc khởi tạo sequence thật.** Plan này định nghĩa thuật toán/điều kiện dừng, không bịa bảng mapping production. Fixture test có mapping tổng hợp tường minh được kiểm trước chạy. Việc duyệt map cụ thể thuộc OAP-03/OAP-10.

## 3. DDL/mapping M1

### Danh mục và cấp bậc

- job_levels: job_level_id INT GENERATED ALWAYS AS IDENTITY PK; job_level_name VARCHAR(100) NOT NULL; created_at/updated_at TIMESTAMPTZ NOT NULL DEFAULT now(). Không rank/status/điểm/liên kết department mới.
- departments/positions giữ nguyên ID/cột/timestamps. UNIQUE tên đã có không đủ chống case; thêm unique expression trên lower(btrim(tên)) cho từng bảng sau preflight và CHECK tên có nội dung. Với job_levels cũng tạo UNIQUE expression lower(btrim(job_level_name)) và CHECK tên có nội dung, cùng quy tắc departments/positions; không chỉ khai báo VARCHAR NOT NULL. Giữ unique hiện hữu nếu chưa có lý do review bỏ; không gộp uniqueness giữa ba loại.
- Chuẩn hóa input tên mới: trim Unicode ở service rồi kiểm 1–100 code point, không bỏ dấu/NFC ngầm; lưu tên đã trim. SQL không chỉ dựa SELECT trước INSERT. Test Unicode whitespace/collation để bảo đảm cách ghi app và biểu thức DB không mâu thuẫn; dữ liệu legacy chưa canonical phải phân loại, không lén UPDATE.
- Thêm employees.job_level_id INT NULL REFERENCES job_levels(job_level_id) ON DELETE NO ACTION, index theo FK; thêm employees.address TEXT NULL DEFAULT NULL, CHECK NULL hoặc có nội dung sau trim. Không bịa address để lấp NULL.
- Mapping cấp bậc chỉ theo bảng đã duyệt. Trống → NULL; giá trị chưa xác minh không tự map, không tuyên bố hoàn tất migrate. Giữ seniority_level cũ trong đợt này để đối chiếu/rollback, code đích không đọc làm cấp bậc chính thức. Không drop trước có quyết định riêng.
- Không thay FK position/department để ON DELETE SET NULL/CASCADE. Xóa danh mục còn bất kỳ employee tham chiếu, kể cả inactive, trả409. job_level nullable không có nghĩa xóa danh mục được tự gỡ liên kết.

### Audit theo dictionary, không thêm framework

Tạo audit_events đúng mục 20: audit_id INT identity PK; actor_id INT NOT NULL FK users; entity_type TEXT NOT NULL; entity_id INT NOT NULL; operation TEXT NOT NULL; occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(); reason TEXT NULL; before_values/after_values JSONB NULL. Index(entity_type,entity_id,occurred_at). JSON chứa schema_version=1, allowlist theo thao tác; không FK tổng quát entity_id.

| Thao tác | entity_type / operation | before / after cho phép |
|---|---|---|
| Tạo/đổi tên/xóa danh mục | department, position, job_level / CREATE, RENAME, DELETE | id/name; tạo before NULL, xóa after NULL, các object có schema_version |
| Tạo tài khoản kèm hồ sơ | employee / CREATE_ACCOUNT | before NULL; after schema_version,user_id,employee_id,employee_code,position_id,department_id,job_level_id,system_roles |

Không ghi password/hash/username/email/address/phone/mail body/token/provider raw error trong audit. Một event CREATE_ACCOUNT ghi cả role ban đầu, không nhân đôi event quyền nếu không có thao tác cấp riêng; actor là Admin tạo, không user mới. Audit thất bại phải rollback cả mutation; no-op rename không tạo lịch sử giả. Không cấp API audit/notification mới trong đợt này; append-only ở đường app, không bịa actor/time cho legacy.

## 4. DDL/mapping M2

### Mã nhân viên

Đề xuất sequence BIGINT MINVALUE 1 INCREMENT 1 NO CYCLE CACHE 1, sequence riêng không dùng employee_id. Giá trị bắt đầu được tính dưới cửa sổ dừng ghi có khóa thích hợp: max phần số của các mã legacy dạng EMP + chữ số (kể cả EMP001/EMP000001), cộng 1; nếu không có thì 1. E001 giữ nguyên, không đồng nhất dải E với EMP. Regex/CAST phải kiểm overflow trước, không cast text bẩn; MAX không dùng làm cơ chế cấp số mỗi request.

Khởi tạo sequence một lần trong migration, không setval/reset ở startup hoặc khi retry. Cấp bằng nextval dưới transaction tạo user–employee, UNIQUE(employee_code) là chốt bảo vệ. Format bằng đệm 0 **tối thiểu** 6 chữ số; không dùng lpad cố định làm cắt số dài, kiểm ≤ VARCHAR(30) và dải BIGINT trước ghi. Không rollback sequence để lấp khoảng trống; ID/code không tái dùng sau thất bại.

GET next-code chỉ đọc snapshot last_value/is_called của sequence CACHE1, format ứng viên tiếp theo; không nextval/khóa giữ số/ghi session. Dữ liệu đọc có thể cũ ngay lập tức; response is_preview=true. Nếu sequence hết dải trả lỗi nội bộ chung, không vòng về 1. Không hứa an toàn khi thao tác ngoài app tự chèn code không theo sequence; preflight/UNIQUE và quyền ghi phải bảo vệ. Không auto bỏ số vì username cũ trùng.

### Email và username

- Thêm UNIQUE expression lower(btrim(employees.email)) sau preflight. Không partial theo active/status; không cho tạo lại account bằng email đã tồn tại. Không sửa email legacy hoặc auth lookup. Account/SelfProfile trả company_email qua cùng helper chuẩn hóa trim/lower của auth, không trả raw email legacy có khoảng trắng; chuẩn hóa projection không UPDATE dữ liệu lưu. Preflight kiểm cả email sau chuẩn hóa đáp ứng định dạng/độ dài response; email không hợp lệ thì dừng để xử lý theo phương án được duyệt, không giả giá trị. Input mới chuẩn hóa dùng helper auth và lưu email canonical, ngăn CR/LF/display-name/multiple-address trước Gmail header; kiểm độ dài theo dictionary. Unicode/collation parity là ca bắt buộc trước DDL, không coi OpenAPI format=email là validator runtime.
- Bỏ cột users.username bằng migration mới, không CASCADE. Constraint/index thuộc cột được bỏ cùng cột sau kiểm metadata; views/functions/trigger phụ thuộc phải review và cập nhật tường minh. Không đổi password_hash/status/ID/user_roles/auth_sessions để làm DROP chạy được.
- Rà cmd/seed (hiện ON CONFLICT username), db/seed, cmd/localdev, internal/testutil/authdb, scripts/authdb/manage.py, migration/integration fixtures, FE model/form/demo. Seed schema đích dùng khóa nghiệp vụ email/code theo thiết kế đã review, không tạo alias thay thế. Fixture dựng v1/v2/v3 giữ username vì đó là schema lịch sử; fixture chạy head không ghi cột đã bỏ. Không sửa Swagger vendor vì từ khóa trùng.
- Giữ users.system_role legacy trong đợt này để tránh mở rộng DROP: user mới ghi EMPLOYEE làm giá trị tương thích; quyền thực chỉ lấy user_roles gồm ADMIN/HR/DIRECTOR do Admin chỉ định. Không suy PM từ chức danh hoặc legacy system_role. SQL plan đề xuất mapping này cần review trước migration/code, không backfill HR/DIRECTOR cho tài khoản cũ.
- Không thêm cột delivery_status/bảng outbox trong đợt này. Kết quả Gmail chỉ có ở response create; GET users không biết trạng thái thư cũ. Không audit token/mail failure như thao tác người dùng.

## 5. Truy vấn, khóa và transaction

Mọi SQL tham số hóa; collection/column allowlist nội bộ, không tên bảng từ input. Lỗi pgx không đi lên HTTP thô. Clock/transaction timeout truyền từ composition root, không query DB trực tiếp từ handler.

### Đọc

- Catalog list: HR hoặc Admin, phiên thường; id tăng dần, page/page_size và literal substring q theo name. %/_ phải escape nếu dùng ILIKE; count/items cùng snapshot để pagination nhất quán trong một response: dùng một câu SQL trả cả hai, hoặc transaction REPEATABLE READ READ ONLY cho hai câu SELECT. READ COMMITTED READ ONLY với hai câu SELECT không đủ; không dùng cơ chế này cho transaction mutation vốn dùng READ COMMITTED. page vượt tổng trả items=[], total thực; kiểm overflow OFFSET bằng int64.
- Account list: Admin, projection JOIN employees/users và role aggregate user_roles ổn định ADMIN,HR,DIRECTOR; q theo code/fullname/email, user_id tăng dần. Count/items dùng cùng cơ chế snapshot như catalog list. Không SELECT *, address/phone hoặc delivery status. Một user không có employee là lỗi nền cần phân loại, không ghép fixture; preflight phải phát hiện trước triển khai.
- Self profile: query một projection nối user→employee→position, LEFT JOIN department/job_levels. WHERE user_id=$actor AND employee_id=$requested; không thấy trả404, không query người khác rồi vô tình log PII. Recheck user ACTIVE/employee ACTIVE và cờ hiện hành trước trả; trạng thái actor không hợp lệ trả401, cờ true trả403, khác với404 của ID không phải chính chủ/không tồn tại. Projection không lọc mất thông tin này rồi map tất cả thành404; principal từ request không thay quyền DB khi nguồn thay đổi. Không đọc CV/history hoặc cấp quyền HR/Admin đọc người khác.
- Không lấy quyền từ users.system_role. Cờ/quyền đọc lại theo nguồn hiện hành; HTTP middleware vẫn bắt buộc, không nhận actor/status/roles từ client để thay quyền người gọi.

List/projection tests phải kiểm company_email legacy có khoảng trắng được trả canonical đúng schema nhưng DB giữ nguyên; email malformed bị preflight từ chối. Với cả ba danh mục, test duplicate khác hoa/thường dưới cạnh tranh. Test pagination có insert/delete giữa count và items bằng barrier để chứng minh một snapshot, kể cả page rỗng.

### Khóa chung với auth

Thứ tự cơ sở: user → employee → session (khi cần), theo giao thức auth hiện hành. Với nhiều user/employee cần khóa IDs tăng dần trong từng loại. Quan trọng: mọi đường sửa status/role về sau phải khóa user trước mới sửa user_roles, để lock actor thực sự chặn race. Ticket này chỉ mô phỏng mutation đó trong DB tests, không chứng nhận API quản trị tương lai.

- Catalog mutation ở READ COMMITTED: BEGIN → khóa actor user/employee, recheck ACTIVE/cờ/role HR → khóa FOR UPDATE danh mục mục tiêu (UPDATE/DELETE) → so expected_before_name chính xác → write → audit → COMMIT. So nền trước kiểm duplicate/reference; missing404, stale409. Name mới bằng hiện hành sau normalize → 200 không update/audit giả. Sau khi giữ khóa catalog, DELETE kiểm EXISTS tham chiếu bằng SELECT thường không khóa row employee; còn tham chiếu trả409 trước khi chạy DELETE. Bước này tránh FK check đi ngược khóa employee đang giữ bởi Admin create; FK NO ACTION vẫn là chốt chống race, không dựa EXISTS đơn lẻ. Ca test phải có actor Admin cũng là employee đang tham chiếu đúng mục HR muốn xóa, và create chờ KEY SHARE catalog. Truy vấn EXISTS thực hiện sau khi lấy lock ở READ COMMITTED để thấy row vừa commit. Kiểm đầy đủ FK tham chiếu từ preflight; không chỉ một nguồn nếu schema có thêm consumer.
- Create: chuẩn bị secret/bcrypt ngoài transaction → BEGIN → khóa actor user/employee, kiểm ADMIN/ACTIVE/cờ → kiểm/giữ khóa KEY SHARE các danh mục theo thứ tự department→position→job_level, ID tăng dần trong mỗi bảng → cấp sequence → INSERT user(legacy EMPLOYEE,ACTIVE,true) → employee(ACTIVE) → user_roles → audit → COMMIT → gọi Gmail ngoài transaction. Email/name/FK lỗi rollback mọi row nghiệp vụ, sequence có thể mất số. Không cấp auth session/Set-Cookie.
- SQL insert có thể thực hiện recheck FK bằng SELECT FOR KEY SHARE cùng FK thật; rename catalog đồng thời không đổi ID và không cần viết lại lịch sử. Người gọi mất quyền trong khi chờ lock phải bị từ chối sau recheck; session bị revoke sau middleware không hứa hủy ngược request đã bắt đầu, phù hợp auth hiện hành. Điểm kiểm session bổ sung nếu có phải giữ cùng thứ tự khóa.
- Deadlock 40P01/lock timeout 55P03/statement timeout có thể xảy ra: rollback, map lỗi tạm thời 503 SERVICE_UNAVAILABLE theo helper/contract; không retry mutation tự động hoặc ghi audit thành công. Với connection mất sau COMMIT phải phân loại chưa xác định riêng. Không coi timeout hoặc deadlock là 409 dữ liệu nếu chưa xác định nguyên nhân nghiệp vụ.
- Commit trả lỗi/chưa rõ: không gửi mail hoặc tự gọi lại transaction; đối soát theo account list. Test phân biệt lỗi trước commit với lỗi phản hồi sau commit; không hứa không ghi khi commit thực tế đã thành công.
- Mail trả lỗi sau commit: service trả account_created=true và mail.status phù hợp; không đổi hash/role/cờ, rollback hoặc sinh lại secret. Thiếu cấu hình/token lỗi trước gửi → failed; request gửi đã ra nhưng timeout → unknown. accepted chỉ khi Gmail API xác nhận. Không phân loại mọi HTTP5xx thành chắc chắn chưa gửi; mapping cụ thể trong adapter test. Response mất ở FE → hướng dẫn đối soát, không retry POST.

## 6. Rollback và phiên bản consumer

- Trước M1/M2, dừng đường ghi, kiểm head/backup theo môi trường. Không chạy down ở Neon/DB phát triển từ task này. Test chỉ dùng DB case hy sinh.
- M1 down: mất cấp bậc/address/audit thêm mới; chỉ diễn tập sau xác nhận chấp nhận mất dữ liệu và consumer không còn đọc cột. Giữ seniority_level không có nghĩa các thay đổi job_level sau migrate tự đồng bộ ngược.
- M2 down: tái tạo username không lấy lại được giá trị đã DROP. Down an toàn phải dừng nếu còn row mà không có nguồn phục hồi đã review; không dùng email/code/random làm username giả để vượt NOT NULL. Down tự động chỉ cho DB rỗng chưa từng cấp sequence (is_called=false), có thể tái tạo cấu trúc rồi bỏ sequence chưa dùng. DB rỗng nhưng từng cấp số cũng phải từ chối. Legacy có dữ liệu phải chứng minh down từ chối không để partial, hoặc restore backup thực sự theo kế hoạch riêng. Không tự tạo bảng backup/PII trong schema app.
- Sequence đã cấp không được reset bởi down/up để tái dùng mã; thiết kế rollback không được DROP rồi CREATE sequence từ max committed như cách khôi phục tương đương. Thiết kế hiện chọn từ chối down sau khi sequence đã cấp bất kỳ số nào; nếu có nhu cầu restore/rollback khác, phải đối soát high-water mark được lưu an toàn theo kế hoạch riêng trước khi chạy lại. DROP username/M2 không có rollback dữ liệu đối xứng; ưu tiên sửa tiến thêm/restore backup theo quy trình được phép, không tuyên bố reversible chỉ vì có file down.
- Ghi kết quả thực driver: lỗi DDL giữa file có rollback không, version dirty ở đâu; không nuốt lỗi/force version. Mỗi đường nâng cấp chạy bằng migration runner thật khi implement, không lấy SQL chạy tay thay chứng minh version chain.

## 7. Ma trận PostgreSQL test dự kiến

**Tất cả ca dưới chưa chạy.** Test head trong case riêng; hai connection/barrier/clock điều khiển được cho race, không sleep ngẫu nhiên. Mail fake chỉ chứng minh thứ tự gọi; inbox thật thuộc OAP-12/15 được cấp quyền riêng.

| Mã | Dữ liệu/thao tác | Kết quả cần chứng minh | WBS |
|---|---|---|---|
| OAP-DB01 | DB rỗng → 000001/2/3 → M1/M2 | Head sạch; schema đích đủ; không username; không force | 04/10 |
| OAP-DB02 | Legacy giả có user_id khác employee_id, role kết hợp/phiên/Unicode/code cũ | Bảo toàn ID/link/hash/email/status/roles/sessions/code; chỉ username bị loại | 04/10 |
| OAP-DB03 | Drift, tên trắng/case trùng, email lower/trim trùng, FK/mapping chưa rõ | Dừng trước DDL không tự cleanup hoặc quyền mới | 03/04/10 |
| OAP-DB04 | Mapping cấp bậc đã duyệt giả, NULL và giá trị chưa xác minh | Đúng ID; chưa rõ bị chặn, không suy từ position; giữ seniority legacy | 04 |
| OAP-DB05 | Tên catalog trim/case/Unicode100/101, raw có padding rồi trim về100, duplicate/no-op/stale/delete absent | UNIQUE/CHECK + audit đúng; status/mã lỗi đúng contract | 05/06 |
| OAP-DB06 | Xóa catalog có reference với create employee đang chờ | FK/barrier, gồm actor Admin tham chiếu mục bị xóa: 409 khi còn ref, không orphan/set NULL/chờ khóa vô hạn | 05/11 |
| OAP-DB07 | N create cùng lúc, rollback sau nextval, preview nhiều lần, mã >6 chữ số | Code unique/gap cho phép, preview không tiêu số, không truncation/reuse | 10/11 |
| OAP-DB08 | EMP001/EMP000001/E001, sequence exhausted, cold start nhiều instance | Khởi tạo đúng dải, lỗi có kiểm soát, startup không reset | 10 |
| OAP-DB09 | Lỗi write user/employee/role/audit, lỗi commit | Không half-write; không mail trước commit; ambiguity không auto retry | 11/13 |
| OAP-DB10 | Mất quyền/khóa/cờ actor đổi trong lúc chờ user lock | Recheck từ chối; fake mutation tuân giao thức, không cấp quyền theo snapshot cũ | 05/11 |
| OAP-DB11 | Self/other/missing ID, position bắt buộc, dept/level NULL, status/cờ đổi | Chỉ self, không PII/row multiplication, không grant qua Admin | 08 |
| OAP-DB12 | M1 down/up, M2 down rỗng chưa cấp số/rỗng sau rollback nextval/có dữ liệu, lỗi giữa DDL | Mất dữ liệu/giới hạn rõ; down từ chối khi không phục hồi được username, dirty đúng driver | 04/10 |
| OAP-DB13 | Auth regression, seed/localdev/fixtures head, create không username | Năm auth API vẫn đúng; đường chạy hiện hành không đòi cột đã bỏ | 10/15 |
| OAP-DB14 | Email normalization parity Go/Postgres và collation đích | Case/space/Unicode không cấp trùng hoặc lookup không xác định; lệch dừng review | 03/10 |

Lệnh hiện hành để kiểm môi trường (chỉ chạy khi task implementation được giao): `python3 scripts/authdb/manage.py check`. Sau khi có tests mới, dùng runner `run -- go test -tags=integration -p 1 ...` đúng package/test đã viết, không chép tên test chưa tồn tại để báo pass. Cổng Go và regression auth theo workflow; root DB test giữ 0 bảng, case cleanup không đụng auth_dev_* đang chạy. DB tests chạy tuần tự trên cluster riêng.

## 8. Đầu ra để bàn giao và review trước triển khai

- OpenAPI/SQL plan được review cùng nhau; mapping fixture/legacy cụ thể, diff migration, test file và lệnh tái hiện là đầu ra tiếp theo OAP-01–04/10, chưa được tạo giả trong plan này.
- SQL M2/user legacy EMPLOYEE, biểu thức normalize/collation, lock scope và down cần review khi có SQL cụ thể; đây là thiết kế đề xuất kỹ thuật, không quyền chạy DB thật.
- Gmail OAuth client/secret store/thư viện/timeout sẽ có brief OAP-12; không cần chọn lại provider. Không lưu delivery status trong đợt này, không endpoint cấp lại/gửi lại.
- Evidence chỉ nhận metadata/kết quả đã redaction, không credential/PII; không gọi mock là DB/runtime. Chưa có Jira mới/commit/push trong đợt soạn này.

Căn cứ PostgreSQL cho review SQL: [row lock/deadlock](https://www.postgresql.org/docs/18/explicit-locking.html), [Read Committed và snapshot](https://www.postgresql.org/docs/18/transaction-iso.html), [lpad có thể cắt chuỗi](https://www.postgresql.org/docs/18/functions-string.html), [sequence](https://www.postgresql.org/docs/18/functions-sequence.html). Các căn cứ này không thay barrier tests của adapter thật.
