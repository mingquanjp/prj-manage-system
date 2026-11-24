# PostgreSQL test cho auth — DATN-6 / AUTH-T05

Nguồn SQL/mapping/fixture và phạm vi: [proposal mục 6.1.7](../../docs/proposals/backend-auth-session-implementation.md#617-bàn-giao-môi-trường-và-kế-hoạch-dữ-liệu--datn-6--auth-t05-08112026). Đây là công cụ local dành riêng cho test, không phải server/runner migration production.

## Điều kiện và tài nguyên

Cần Python 3 stdlib, Go theo go.mod, Docker local qua Unix socket và image chính thức `postgres:18`. Lần `up` đầu tải image; container dùng image ID bất biến đã tải, state lưu ID để kiểm lại. Tag major nhận patch mới khi tạo mới; không khẳng định patch/digest đã pin cho mọi máy. PostgreSQL 18 dùng volume tại `/var/lib/postgresql` theo [tài liệu image chính thức](https://github.com/docker-library/docs/blob/master/postgres/README.md).

| Tài nguyên | Giá trị |
|---|---|
| Container | `ai-staffing-auth-test` |
| Volume | `ai-staffing-auth-test-data` |
| Endpoint host | `127.0.0.1:55438` → container `5432` |
| Database | `ai_staffing_auth_test` |
| User/DB owner | `ai_staffing_test` |
| State bí mật | `.local/authdb/state.json`, quyền 0600, Git ignore |

Password ngẫu nhiên chỉ dùng local test, sinh khi tạo mới; không in URL/password. User bootstrap của image là superuser trong cluster test riêng; không là mô hình quyền runtime production. Không dùng credential này cho app hoặc mailbox thật. Docker admin vẫn có thể inspect credential container: đây không phải kho bí mật production.

## Ứng dụng local và test dùng DB riêng

`pnpm dev` tại web gọi [runner local](../dev.py), tái dùng guard Docker/owner/cluster ở module này để dựng DB ngẫu nhiên `auth_dev_*` riêng. `cmd/localdev` chạy HTTP auth thực liên tục đến Ctrl+C và chỉ dọn DB vừa tạo; không dùng `auth_case_*` hoặc nạp fixture test làm runtime. DB gốc vẫn0bảng; kiểm thử `auth_case_*` và localdev không dùng chung dữ liệu. Không tự quét DB còn lại sau crash. Các lệnh `stop`/`destroy` bên dưới sẽ ngắt cả localdev đang dùng cluster; cần dừng `pnpm dev` trước, không chạy destroy để sửa UI.

## Lệnh từ root repo

```sh
python3 scripts/authdb/manage.py up
python3 scripts/authdb/manage.py check
python3 scripts/authdb/manage.py stop
# Khởi động lại cùng container/volume, không reset dữ liệu:
python3 scripts/authdb/manage.py up
# Xóa tài nguyên TEST sở hữu, mất dữ liệu test; không xóa image:
python3 scripts/authdb/manage.py destroy --confirm ai-staffing-auth-test
```

Không chạy `go run ./scripts/authdb/probe.go` trực tiếp: probe không có guard sở hữu Docker, chỉ được runner gọi sau khi kiểm đích.

`check` chỉ SELECT metadata: major/name/user/DB owner/số bảng public/cluster system ID. Không chạy migration, seed hoặc đọc dữ liệu nhân viên. Probe psql dùng TCP/password trong container; probe Go dùng TCP loopback từ host và đối chiếu cùng metadata/cluster. Không chỉ dùng pg_isready hoặc mở port làm bằng chứng credential/đích đúng.

`stop` giữ volume và state. `destroy` yêu cầu xác nhận tên, kiểm engine/container ID/image ID/nhãn owner trên container + volume/mount/port trước xóa. Chỉ xóa hai tài nguyên tên cố định, không prune hay dọn theo APP_ENV. Sai guard thì dừng.

Nếu tên container/volume đã tồn tại mà thiếu state: runner từ chối nhận làm tài nguyên của mình. Không xóa tài nguyên khác để vượt lỗi. Nếu tạo/xóa bị ngắt giữa chừng, có thể còn tài nguyên riêng chưa đủ state; giữ lại và đối chiếu thủ công ID/label/quyền sở hữu trước xử lý, không sửa state để nhận tài nguyên không rõ nguồn.

## Bàn giao runner cho DATN-7 và các DB tests sau

```sh
# Chỉ dùng khi task test/migration cụ thể đã được phép:
python3 scripts/authdb/manage.py run -- go test -tags=integration ./internal/repository/postgres -count=1
```

Runner kiểm đích trước, cấp `TEST_DATABASE_URL` từ state sở hữu và `AUTH_TEST_DB_SYSTEM_ID` từ metadata cluster đã đối chiếu vào subprocess, loại `DATABASE_URL`, `APP_ENV` và biến `PG*` từ shell. Nếu shell đã có `TEST_DATABASE_URL` khác state thì từ chối, không ghi đè âm thầm. Lệnh test phải yêu cầu biến test này, tự fail khi thiếu và không gọi loader dotenv. Không xuất URL/credential trong assertion, stdout hoặc lỗi driver.

`run` là wrapper chạy lệnh được người phát triển tin cậy, **không phải sandbox**: nó không ngăn một chương trình tùy ý tự đọc `.env` hoặc nối DB khác. Không đưa lệnh không review vào wrapper. Hai guard sở hữu/URL là điều kiện trước thao tác; không bảo vệ khỏi Docker admin thay container đồng thời.

Ghi nhận DATN-6 ở trên chỉ là wrapper và kiểm môi trường; DATN-7 đã thêm integration runner dùng `TEST_DATABASE_URL` trực tiếp với pgx/migrate driver, không gọi `cmd/migrate up` vì lệnh đó đọc `DATABASE_URL` và có thể nạp dotenv dev. Không đổi/bỏ RequireDev để chạy test.

## Migration auth integration — DATN-7 / AUTH-T06

```sh
python3 scripts/authdb/manage.py up
python3 scripts/authdb/manage.py run -- go test -tags=integration ./migrations -count=1 -v -timeout=4m
```

`go test ./...` thông thường không chạy bộ DB này (build tag `integration`). Khi chạy integration mà thiếu/sai URL hoặc thiếu cluster ID thì fail, không skip/fallback. Không gọi trực tiếp để bỏ guard. Go test dùng dependency hiện có, không dotenv/seed ứng dụng.

Mỗi ca xác minh cluster PG18/name/user/owner, CREATE một database `auth_case_<random>` bằng template0 trong cluster test đã chốt; chỉ đăng ký tên để cleanup sau CREATE thành công. Driver dùng `stdlib.OpenDB(*cfg)` với database case đã đổi và kiểm `current_database()` trước migration; không dùng `cfg.ConnString()` vì có thể trả URL gốc. Cleanup đóng các runner/connection trước DROP đúng database vừa tạo, kiểm lại cluster và DB owner, không FORCE/prune. Nếu bị kill/panic ngoài cleanup hoặc drop lỗi, giữ tài nguyên để đối chiếu, không nhận database có sẵn làm của mình.

DB gốc `ai_staffing_auth_test` giữ 0 bảng; migration/dữ liệu giả chỉ nằm trong database case. Ca kiểm cuối xác nhận DB gốc trống và không còn database case. Bộ test chạy tuần tự, không chạy hai lượt trên cùng cluster cùng lúc. Không dùng DB case làm DB runtime lâu dài cho task sau; mỗi bộ test tiếp theo cần fixture riêng qua cùng guard.

Phạm vi thực kiểm:

- MIG-01: DB rỗng qua 000001/000002/000003, up lại ErrNoChange.
- MIG-02: legacy giả version 1 qua cờ và auth; so snapshot nội bộ, không xuất hash/email.
- MIG-03: version 2 có true/false giữ nguyên khi nâng và down/up auth.
- MIG-04: SQL test riêng trong bộ nhớ lỗi sau DDL/backfill; đóng connection, xác minh rollback hai bảng, legacy giữ nguyên, version 3 dirty=true và lần up tiếp bị chặn. Không sửa file migration hoặc force version.
- MIG-05: down/up auth trên case hy sinh; role cấp bổ sung và phiên không phục hồi giả.
- Preflight: kiểm projection legacy/type/enum/PK/unique/FK, link thiếu/trùng, email lower/trim trùng, auth table có sẵn và cờ theo version. Các ca drift được từ chối trước nâng version. Đây là preflight rehearsal test, chưa phải công cụ kiểm kê/áp Neon.
- PK/FK/unique/NOT NULL/CHECK, identity/default, sáu cột phiên/index user_id và NO ACTION. Ca FK dùng user riêng không có employee để không nhận lỗi từ FK legacy thay auth.

Migration 000003 được nhúng vào binary migrate như các file SQL khác. **Không chạy `cmd/migrate up` lên Neon/DB phát triển chỉ vì file đã có**: mapping dữ liệu thật, preflight/drift/backup/consumer và quyền áp đích thật cần xác nhận riêng. Rollback auth làm mất dữ liệu hai bảng; không bảo đảm phục hồi quyền/phiên bằng up. Đợt này không diễn tập down 000002, không kiểm auth runtime/race; xem kết quả và giới hạn tại proposal mục 6.1.8.

## Truy vấn auth integration — DATN-8 / AUTH-T07

```sh
python3 scripts/authdb/manage.py up
python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./migrations ./internal/repository/postgres -count=1 -v -timeout=4m
```

`-p 1` chạy hai package tuần tự vì cả hai có ca xác minh cluster đã dọn hết database case. Dùng một lượt integration tại một thời điểm trên cluster.

Trong test mới, `authdb.NewFixture(t)` tạo database case đã migrate 000003. `AddAccount` ghi user, employee và role trong cùng transaction. Có thể chủ động bỏ employee để kiểm ca thiếu liên kết. `AddSession` nạp phiên với hash và thời điểm do test chọn. `NewCase(t)` dành cho ca cần database rỗng, ví dụ kiểm lỗi truy vấn khi thiếu bảng. Dữ liệu fixture được dọn cùng database case.

Bộ repository kiểm hợp đồng đọc trên PostgreSQL thật. Bằng chứng và ranh giới với service/HTTP ở [proposal mục 6.1.9](../../docs/proposals/backend-auth-session-implementation.md#619-truy-vấn-định-danh-và-quyền-hiện-hành--datn-8--auth-t07-08112026).

## Transaction auth integration — DATN-9 / AUTH-T08

```sh
python3 scripts/authdb/manage.py up
python3 scripts/authdb/manage.py run -- go test -race -tags=integration -p 1 ./migrations ./internal/repository/postgres -count=1 -timeout=4m
python3 scripts/authdb/manage.py run -- go test -tags=integration ./internal/repository/postgres -run '^TestAuthRepositoryRootRemainsEmpty$' -count=1
```

Lệnh cuối chạy riêng sau toàn bộ suite để xác nhận DB gốc vẫn trống và không còn DB case, không phụ thuộc thứ tự đăng ký các test trong package. Giữ một lượt integration tại một thời điểm. `authdb.Case.Connect(t)` mở thêm connection từ bản sao config của DB case đã xác minh, kiểm lại cluster/name/owner và đăng ký cleanup. Không dùng `ConnString()` để chuyển DB đích.

Sáu nhóm transaction mới kiểm login/phiên cũ, nền thay đổi, logout, rollback, 13 ca cạnh tranh và commit mất phản hồi. Barrier xác nhận blocker/waiter thực bằng `pg_blocking_pids`; ticker chỉ thăm dò quan hệ khóa có timeout, không dùng sleep để giả định request đã chờ khóa. Clock điều khiển an toàn giữa goroutine, gồm biên hết hạn trong khi chờ user/session. Test gây lỗi bằng trigger chỉ tạo trong database case. Mô phỏng mất phản hồi gọi commit PostgreSQL thật rồi trả lỗi I/O từ wrapper để kiểm adapter không báo thành công giả; chưa thử cắt mạng thực.

Đã chạy các cổng Go cùng bộ integration ở trên trên PostgreSQL 18 test sở hữu, không đổi schema/migration. [Proposal mục 6.1.10](../../docs/proposals/backend-auth-session-implementation.md#6110-transaction-tạo-và-thu-hồi-phiên--datn-9--auth-t08-09112026) ghi bằng chứng và giới hạn cho DATN-10/DATN-18. Kiểm tĩnh harness/contract không thay bằng chứng transaction.

```sh
go vet ./...
go test ./... -count=1
go build ./...
go vet -tags=integration ./migrations ./internal/repository/postgres ./internal/testutil/authdb
node --test scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs
node scripts/check-harness.mjs
git diff --check
```

## Đổi mật khẩu lần đầu integration — DATN-10 / AUTH-T09

Dùng bộ lệnh PostgreSQL 18, kiểm cleanup và các cổng Go/harness ở mục DATN-9 ngay trên. Để chạy riêng nhóm DATN-10 trên cùng môi trường đã xác minh:

```sh
python3 scripts/authdb/manage.py run -- go test -race -tags=integration ./internal/repository/postgres -run '^(TestReadPasswordChangeBasis|TestChangeInitialPassword.*)$' -count=1 -timeout=4m
```

Chỉ chạy một lượt integration tại một thời điểm. Các ca cạnh tranh dùng hai connection thực thi, với connection quan sát riêng khi cả hai worker đều bận. Clock điều khiển thời điểm test. `pg_blocking_pids` xác nhận PostgreSQL đang chờ khóa trước khi nhả transaction đầu hoặc dịch clock. Trigger gây lỗi nằm trong database case. [Proposal mục 6.1.11](../../docs/proposals/backend-auth-session-implementation.md#6111-transaction-đổi-mật-khẩu-lần-đầu--datn-10--auth-t09-09112026) ghi phạm vi bàn giao và ranh giới với kiểm thử HTTP.

## Tích hợp HTTP và PostgreSQL — DATN-18 / AUTH-T17

Năm API chạy qua cùng composition root `buildHandler` của startup. HTTP integration dùng pool thật trên từng DB case. Test guard dùng router/middleware và repository thật, endpoint thăm dò chỉ tồn tại trong test. Nguồn và ma trận bằng chứng ở [proposal mục 6.2.13](../../docs/proposals/backend-auth-session-implementation.md#6213-tích-hợp-năm-api-với-postgresql--datn-18--auth-t17-12112026).

```sh
python3 scripts/authdb/manage.py up
python3 scripts/authdb/manage.py run -- go test -race -tags=integration -p 1 ./migrations ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m
# Chạy sau suite để xác nhận DB test gốc/case đã dọn:
python3 scripts/authdb/manage.py run -- go test -tags=integration ./internal/repository/postgres -run '^TestAuthRepositoryRootRemainsEmpty$' -count=1
```

Giữ một lượt DB integration tại một thời điểm. Hai request cạnh tranh dùng pool và barrier PostgreSQL, fake clock atomic; không sleep để đoán thứ tự. Test lỗi ghi dùng trigger riêng trong DB case. Test mất phản hồi commit thực hiện commit DB thật rồi wrapper driver trả lỗi I/O để đối chiếu HTTP 503 với trạng thái đã commit. Chưa cắt mạng thật hoặc kiểm cookie qua browser. Password/email/tài khoản fixture đều giả; không in URL, cookie, hash hoặc mật khẩu khi test thất bại.

## Kiểm tra công cụ

```sh
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s scripts/authdb -p 'test_*.py' -v
node --test scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs
node scripts/check-harness.mjs
```

Unit tests dùng Docker giả, không là bằng chứng PostgreSQL thật. Bằng chứng local thực chạy được ghi riêng tại proposal mục 6.1.7. Không tự chạy `destroy` trong test suite. Không dùng seed ứng dụng hoặc Neon làm DB test.

## DATN-24 — Migration M1 cho danh mục, hồ sơ và audit

M1 là `000004_oap_m1`. Ca bootstrap và ca legacy đi qua driver pgx của
golang-migrate. Fixture auth dùng head hiện hành. Các ca lịch sử riêng của auth
vẫn kiểm mốc `000003`.

```sh
python3 scripts/authdb/manage.py check
python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./migrations ./internal/repository/postgres ./cmd/server ./cmd/localdev -count=1 -timeout=5m
python3 scripts/authdb/manage.py run -- go test -tags=integration ./internal/repository/postgres -run '^TestAuthRepositoryRootRemainsEmpty$' -count=1
```

Chạy từng lượt theo thứ tự. Runner xác minh PG18 và quyền sở hữu. Test đóng kết nối
rồi dọn đúng database case của từng lượt. Nếu cleanup thất bại, giữ case lại
để đối soát. Không dùng `destroy` như bước dọn thường ngày.

- OAP-DB01/02 kiểm schema M1, seed lặp lại và snapshot legacy gồm ID, liên kết,
  hash, email, trạng thái, quyền, phiên, mã và timestamp. Backfill chỉ tắt
  trigger timestamp của employee trong transaction giữ khóa rồi bật lại.
- OAP-DB03/04 kiểm runner dừng ở version 3 sạch khi thiếu mapping, tên trùng,
  tên trắng hoặc schema lệch. Mapping fixture nối giá trị seniority gốc với tên
  cấp bậc đã xác nhận, truyền qua tham số phiên của kết nối test. CLI mặc định
  nhận mapping rỗng, nên dừng nếu có giá trị legacy cần ánh xạ. Đường áp dụng
  mapping dữ liệu đích cần brief và review riêng trước khi triển khai.
- OAP-DB12 phần M1 kiểm down/up và lỗi giữa DDL. Lỗi DDL rollback schema nhưng
  driver giữ version 4 dirty. Down loại bỏ địa chỉ, cấp bậc và audit của M1.
  Dữ liệu legacy còn lại được giữ, còn dữ liệu mới cần backup nếu phải phục hồi.
- Hồi quy auth dùng repository và HTTP với PostgreSQL M1. Test localdev gọi
  bước chuẩn bị thật trên DB case riêng, kiểm schema/fixture mà không mở cổng.

Unique index dùng `lower(btrim(name))` theo SQL plan. Preflight chặn tên legacy
có padding Unicode còn lại sau `btrim`, thay vì tự sửa. CHECK nhận diện chuỗi
toàn khoảng trắng Unicode, kể cả NBSP. Kiểm case tiếng Việt chạy theo collation
của cluster test. Cần đối chiếu collation và mapping ở môi trường đích trước
khi lên kế hoạch áp dụng ngoài test.

## DATN-25 — Repository/service CRUD danh mục

Đầu ra cho OAP-T04 là `internal/service/organization.CatalogService`, port
`CatalogRepository` và adapter `postgres.NewCatalogRepository(db, timeout)`.
Composition root của ticket HTTP cấp pool và timeout dương; mỗi lời gọi dùng một
transaction với deadline hữu hạn. Service không import HTTP/pgx. `ActorID` phải
lấy từ principal đã được middleware xác thực, không nhận từ JSON/query của client.
Adapter chỉ dùng ID này để đọc lại trạng thái, cờ và quyền hiện hành trong DB;
không dùng role/status client gửi hoặc `users.system_role` legacy.

- List mặc định page 1, page_size 20, tối đa 100; page tối đa 2147483647,
  OFFSET int64. Transport phải từ chối giá trị query được truyền tường minh sai
  (kể cả `page=0`), chỉ dùng mặc định khi vắng input. `q` trim Unicode, tối đa
  100 code point; `%`, `_` và `\` tìm literal. Actor/count/items cùng transaction
  REPEATABLE READ READ ONLY, ID tăng dần, trang vượt tổng vẫn trả total.
- Mutation READ COMMITTED khóa user → employee của actor, kiểm ACTIVE/cờ/HR,
  rồi khóa catalog khi rename/delete. Missing trước stale; giá trị nền giữ nguyên
  để so chính xác. Tên mới trim Unicode và kiểm 1–100 code point; không NFC/bỏ dấu.
  Rename no-op không UPDATE timestamp hoặc ghi audit.
- Delete kiểm tham chiếu employee bằng SELECT thường sau khóa catalog, gồm cả
  inactive; FK NO ACTION bảo vệ cuối. Schema M1 hiện chỉ có employee tham chiếu
  ba danh mục này. Schema/consumer mới phải rà lại FK và thứ tự khóa trước tích hợp.
  Không thêm row lock employee sau catalog, tránh đảo khóa với Admin tạo tài khoản.
- Audit CREATE/RENAME/DELETE cùng transaction, JSON chỉ schema_version/id/name;
  create có before NULL, delete có after NULL. Không ghi credential, profile hay
  principal thô. Không tự retry mutation; sequence identity có thể mất số khi rollback.
- Port phân biệt validation, unauthenticated, password-change-required, forbidden,
  not-found, stale, duplicate-name, catalog-in-use, unavailable và persistence.
  Deadline nội bộ có `ErrUnavailable`; caller cancel/deadline được giữ để caller
  xử lý context. Lỗi COMMIT có thêm `ErrCommitUnknown`: chưa được khẳng định
  mutation chưa ghi hoặc tự gửi lại. Mapping HTTP/DTO thuộc OAP-T04.

Lệnh tái hiện trên PG18 test sở hữu, chạy từng lượt tuần tự:

```sh
python3 scripts/authdb/manage.py check
python3 scripts/authdb/manage.py run -- go test -race -tags=integration -p 1 ./internal/repository/postgres -count=1 -timeout=5m
python3 scripts/authdb/manage.py run -- go test -tags=integration ./internal/repository/postgres -run '^TestAuthRepositoryRootRemainsEmpty$' -count=1
```

Bằng chứng thực chạy ngày 15/11/2026: cổng `go vet ./...`,
`go test ./... -count=1`, `go build ./...` đạt; toàn bộ repository integration
với race detector đạt (57.815s), gồm hồi quy auth trên schema M1. Kiểm môi trường
xác nhận PostgreSQL 18, owner đúng, DB gốc 0 bảng. Test cleanup sau suite đạt
(0.581s), không còn DB case và DB gốc vẫn trống. Harness tests 14/14 đạt;
kiểm kê nguồn và `git diff --check` đạt. Không có migration/dependency mới.

| Ca | Test/bằng chứng | Giới hạn |
|---|---|---|
| OAP-DB05 | `TestCatalogCRUDAndAudit`, `TestCatalogCurrentActorAndReferences`: cả ba loại, 100/101 code point, Unicode trim, no-op/timestamp, duplicate, stale, missing, quyền hiện hành và inactive reference | Collation của PG test; chưa kiểm collation đích |
| Snapshot/literal search | `TestCatalogPaginationSnapshotAndLiteral`: connection khác DELETE/INSERT giữa count/items, trang thường/rỗng; `%`, `_`, `\` không wildcard | DB thật, barrier qua wrapper transaction chỉ trong test |
| OAP-DB06 | `TestCatalogDeleteWithAdminReferenceRace`, `TestCatalogDeleteWaitsForNewReference`: cả hai chiều chờ khóa, gồm Admin actor tham chiếu mục bị xóa; không orphan | SQL tạo fixture trên M1 và KEY SHARE theo giao thức, chưa phải adapter tạo tài khoản OAP-T09 |
| OAP-DB10 | `TestCatalogActorRecheckAfterLock`: user lock chặn request, đổi role/status/cờ rồi commit; request phải đọc lại | Mutation quản trị giả lập đúng thứ tự user → employee; chưa có API quản trị |
| Duplicate/stale race | `TestCatalogConcurrentDuplicateAndStale`: hai HR/connection tạo cùng tên khác case và rename cùng nền; UNIQUE/row lock quyết định kết quả | `pg_blocking_pids` xác nhận chờ thật, không sleep ngẫu nhiên |
| Atomic/lỗi | `TestCatalogAtomicAuditAndCommit`: trigger audit lỗi rollback CREATE/RENAME/DELETE; deferred trigger gây lỗi COMMIT thật; commit DB thật rồi wrapper trả lỗi I/O | Mất phản hồi được fault injection, chưa cắt mạng thật |
| Timeout | `TestCatalogLockTimeoutRollsBack` và unit tests adapter | Context deadline có giới hạn; không giả timeout thành stale hoặc mất quyền |

Chưa có endpoint/UI danh mục, wiring composition root hoặc quyền middleware của
OAP-T04. Bộ test này không nhận mail/browser/real create adapter là đã đạt. Không
chạm Neon, `.env` thật hoặc dữ liệu production. Các DB case do runner tạo/dọn;
không dùng `destroy`, FORCE hoặc prune để làm test xanh.

## DATN-26 — HTTP danh mục và Swagger dev

Composition root `internal/app.AuthHandler`/`LocalAuthHandler` lắp cùng pool cho
năm API auth và `organization.CatalogService`. Startup/localdev gọi router ứng
dụng có dependency catalog bắt buộc. Constructor auth-only vẫn phục vụ các test
riêng của auth, không công bố catalog khi chưa được lắp.

Đầu ra OAP-T05: GET/POST `/api/departments`, `/api/positions`, `/api/job-levels`;
PATCH/DELETE item `/{id}`. HTTP tests đối chiếu đúng 12 operation với canonical
OAP và vẫn đủ năm API auth/health. GET cần HR/Admin, mutation chỉ HR; role kết hợp
được cộng từ server. Mọi route cần phiên thường, kể cả ADMIN/HR bị chặn khi cờ
buộc đổi còn true. Handler không nhận actor/role từ JSON; repository kiểm lại
trạng thái/cờ/quyền hiện hành theo bàn giao DATN-25.

Input/output cho FE:

- GET: `page` mặc định1, `page_size` mặc định20/tối đa100, `q` trim Unicode
  tối đa100 code point. Query lạ/lặp/sai encoding bị400; chỉ mặc định khi vắng
  input, `page=0` là400. Response `{items:[{id,name}],pagination:{page,page_size,total}}`.
- POST: `{name}` →201 `{id,name}` và Location item tương ứng. Không có GET item
  riêng. PATCH: `{name,expected_before_name}` →200 `{id,name}`. DELETE:
  `{expected_before_name}` →204 body rỗng, không Content-Type JSON.
- Mutation dùng cookie và X-CSRF-Token của phiên hiện hành, Origin chính xác,
  application/json; body8192byte thực tối đa, kể cả chunked. Parser dùng chung
  auth: field lạ/trùng/thiếu/sai kiểu/null/trailing JSON/UTF-8 hoặc surrogate sai
  bị từ chối. Tên mới do service trim Unicode rồi kiểm1–100 code point; tên nền
  không trim. Không áp limiter auth sang catalog.
- Mọi response catalog no-store, không Set-Cookie hoặc CORS wildcard. Lỗi có
  `code,message,request_id` server cấp; không SQL/credential/driver error thô.
  400 VALIDATION_ERROR;401 UNAUTHENTICATED;403 FORBIDDEN/PASSWORD_CHANGE_REQUIRED;
  404 NOT_FOUND;409 DUPLICATE_NAME/STALE_DATA/CATALOG_IN_USE;413 PAYLOAD_TOO_LARGE;
  415 UNSUPPORTED_MEDIA_TYPE;503 SERVICE_UNAVAILABLE hoặc500 INTERNAL_ERROR.
- Mất phản hồi mutation/COMMIT chưa rõ: không tự retry POST/PATCH/DELETE; đọc lại
  danh sách để đối soát. FE vẫn phải phân biệt lỗi DB với hết phiên.

Swagger cùng backend, dev-only: `/docs/`, nhóm Auth/Danh mục. Contract auth giữ
nguyên byte nguồn ở `/docs/openapi.json`; `/docs/catalog.openapi.json` lọc đúng
12 operation và components cần thiết từ canonical OAP, không công bố bốn API
account/profile chưa triển khai. Interceptor chỉ cho operation đã công bố cùng
origin, thêm PATCH/DELETE và item ID hợp lệ. Authorize token lại khi chuyển nhóm;
không lưu authorization vào storage. Hướng dẫn thao tác ở [API README](../../docs/api/README.md#datn-26--12-api-danh-mục-qua-httppostgresql).

Lệnh kiểm:

```sh
go vet ./...
go test ./... -count=1
go build ./...
node --test scripts/check-swagger.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs
python3 scripts/authdb/manage.py check
python3 scripts/authdb/manage.py run -- go test -race -tags=integration -p 1 ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m
python3 scripts/authdb/manage.py run -- go test -tags=integration ./internal/repository/postgres -run '^TestAuthRepositoryRootRemainsEmpty$' -count=1
node scripts/check-harness.mjs
git diff --check
```

Không chạy hai lượt DB song song. Chỉ dùng cluster test PG18 đã guard; không đọc
`.env` thật, dùng DATABASE_URL fallback hoặc áp Neon. Không migration/dependency/
biến môi trường mới; catalog transaction có timeout10s trong timeout HTTP30s.

Bằng chứng thực chạy ngày 16/11/2026: `go vet ./...`, unit tests toàn bộ và
`go build ./...` đạt; 36 kiểm tra Node Swagger/contract/harness đạt. Guard xác
nhận PostgreSQL 18, đúng owner, DB gốc 0 bảng. Suite integration có race detector
đạt: repository 102.889s, transport 5.239s, cmd/server 94.668s, gồm hồi quy năm API
auth và 12 API danh mục qua composition root thật. Cleanup sau suite đạt (0.636s),
DB gốc vẫn trống và không còn DB case. Sub-agent đã review hai lượt,
không còn finding chặn. Một lỗi fixture ban đầu (dùng tài khoản kiểm quyền làm
inactive reference) đã tách riêng và suite cuối chạy lại đạt; không sửa quyền
để làm test xanh.

Các test mới:

| Ca | Bằng chứng | Giới hạn |
|---|---|---|
| 12 operation, quyền, lỗi/DTO/header | `TestCatalogRouteInventory`, `TestCatalogHTTPPermissionAndErrors`, `TestCatalogApplicationPostgresCRUD` | Route inventory/handler fake không thay HTTP+DB thật |
| Body/strict JSON/Origin/CSRF | `TestCatalogHTTPBodyAndProtection`, `TestCatalogApplicationChunkedBodies`, `TestCatalogApplicationSecurityAndAuditFailure` | Actual HTTP chunked8192/8193 cho cả3 loại × POST/PATCH/DELETE; không phải browser proof |
| Limited/missing session, localdev Origin | `TestCatalogAllRoutesRequireNormalSession`, `TestCatalogLocalDocsOriginIsolation`, HTTP PG role/limited cases | Origin từ request test; không giả browser đã kiểm |
| Audit/DB lỗi, rotation | HTTP PostgreSQL kiểm lỗi audit rollback catalog, lỗi query/pool không biến thành401, cookie cũ và CSRF cũ bị từ chối | DDL lỗi chỉ trên DB case; đóng pool/fault injection, không cắt mạng thật |
| Swagger | `TestSwaggerCatalogCanonicalSubset`, `scripts/check-swagger.test.mjs` | Canonical subset/ref/dev-only/selector/allowlist kiểm bằng httptest và VM; chưa kiểm rendering/cookie/CSP qua browser trong DATN-26 |

Chưa nối UI danh mục OAP-T05, chưa API tài khoản/hồ sơ hoặc cấp tài khoản/mail.
Các ca cạnh tranh repository dùng lại bộ DB thật DATN-25; create adapter thật vẫn
thuộc OAP-T09. Không tuyên bố toàn OAP/FR đã hoàn tất hoặc tự chuyển Jira Done.

## DATN-28 — API hồ sơ chính chủ

Mục tiêu: bàn giao GET `/api/employees/{id}` với projection tối thiểu cho OAP-T07.
Nguồn: proposal OAP mục6.2–6.4/F3, SQL plan mục5/OAP-DB11, SelfProfile trong
OpenAPI OAP và workflow hồ sơ/bảo mật. Không sửa schema, nối UI, cấp quyền HR/Admin
đọc người khác, trả lịch sử đề nghị/CV/audit hoặc mở rộng `/auth/me`.

### Triển khai và đầu vào cho OAP-T07

- Service/port: `internal/service/organization/profile.go`; HTTP/DTO:
  `internal/transport/http/organization_profile.go`; adapter PG:
  `internal/repository/postgres/organization_profile.go`.
- Composition root hiện có lắp profile cùng auth/catalog; application router thiếu
  dependency phải lỗi startup. Route yêu cầu phiên thường, không cần role đặc biệt.
- Actor lấy user_id từ middleware; requested ID là employee_id. Repository kiểm lại
  trạng thái user/employee và cờ trong transaction REPEATABLE READ READ ONLY, rồi
  query `WHERE user_id=$actor AND employee_id=$requested` trong cùng snapshot.
  Không đọc hồ sơ người khác hoặc JOIN roles làm nhân dòng. Inactive/mất actor401,
  cờ true403 PASSWORD_CHANGE_REQUIRED; other/missing cùng404 NOT_FOUND.
- Profile gồm đúng11 trường theo canonical schema. Position bắt buộc; phone/address/
  department/job_level NULL giữ null. Tên danh mục hiện hành theo ID. Company email
  dùng NormalizeLoginEmail khi đọc, không ghi lại email legacy. Projection dữ liệu
  hỏng trả500; lỗi DB/deadline phân loại500/503, không che thành404.
- OAP-T07 bootstrap me, lấy employee_id từ principal, gọi endpoint này rồi map DTO
  vào UI thật; không ghép principal với employee fixture. Cookie/no-store theo auth
  hiện có, không cấp/gia hạn cookie qua GET. Query lạ bị400 theo contract.
- Swagger dev cùng backend `/docs/`: nhóm Tổ chức và hồ sơ có13 operation đã nối,
  giữ URL spec `/docs/catalog.openapi.json`; canonical OAP không sao chép vào handler.
  Ba API tài khoản còn lại chưa công bố. Không mở docs ngoài dev.

### Ca kiểm chứng

| Ca | Bằng chứng | Giới hạn |
|---|---|---|
| Self/other/missing, user_id khác employee_id, nhiều role, nullable, email canonical không UPDATE, tên danh mục đổi giữ ID | `TestSelfProfileProjectionAndCurrentState`, `TestSelfProfileApplicationPostgres` | Dữ liệu tổng hợp trong DB case sở hữu; không dữ liệu Neon |
| Actor/cờ thay đổi sau principal cũ | `TestSelfProfileRejectsStaleMiddlewarePrincipal` dùng auth stub giữ principal cũ và adapter PG thật | Chứng minh recheck của profile; xác thực cookie thật kiểm riêng qua composition root |
| Snapshot khi connection khác commit giữa đọc actor/projection | `TestSelfProfileConsistentSnapshot` | Barrier sau Scan actor; không sleep. Request đọc snapshot hợp lệ đã bắt đầu, lần sau thấy trạng thái mới; không hứa hủy ngược request |
| Actor không có employee/missing, email malformed, thiếu position do drift | `TestSelfProfileMissingBasisAndMalformedData` | Fault FK chỉ trên case hy sinh, không sửa migration lịch sử |
| DB schema lỗi500, chờ khóa quá deadline503 | `TestSelfProfileDatabaseErrors`, `TestSelfProfileStorageTimeoutIs503`, HTTP composition test | Khóa DB thật trước query và deadline nội bộ100ms; không mô phỏng mất mạng thật |
| DTO11 trường, null/no-store/không Set-Cookie, query400, guard401/403, route/Swagger | `TestSelfProfileHTTP`, `TestSelfProfileRouteAndSwagger`, `TestApplicationRequiresProfileDependency`, Swagger VM | Unit/fake không thay HTTP/PG; không chứng minh browser rendering |

### Lệnh tái hiện

Tại thư mục gốc, mỗi lượt DB chạy tuần tự qua runner đã xác minh:

```sh
go vet ./...
go test ./... -count=1
go build ./...
python3 scripts/authdb/manage.py check
python3 scripts/authdb/manage.py run -- go test -race -tags=integration -p 1 ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m
python3 scripts/authdb/manage.py run -- go test -race -tags=integration -p 1 ./internal/transport/http -run '^TestSelfProfileStorageTimeoutIs503$' -count=1 -timeout=1m
python3 scripts/authdb/manage.py run -- go test -tags=integration ./internal/repository/postgres -run '^TestAuthRepositoryRootRemainsEmpty$' -count=1
node --test scripts/check-swagger.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs
node scripts/check-harness.mjs
git diff --check
```

Runner/NewCase tạo-dọn DB case riêng, DB gốc giữ trống; không fallback DATABASE_URL,
không nạp `.env` thật, không truy cập Neon hoặc chạy suite DB song song. Không có
migration mới cần rollback trong ticket này. Browser/HTTPS production và UI hồ sơ
chưa kiểm trong DATN-28; tests HTTP/DB không thay bằng chứng browser của OAP-T07.
Không nhận toàn FR-009, chức năng edit/duyệt hoặc HR read-all đã hoàn tất.

Bằng chứng thực chạy 16/11/2026: guard xác nhận PG18/owner đúng, DB gốc0 bảng;
`go vet ./...`, `go test ./... -count=1`, `go build ./...` đạt trên snapshot cuối.
37 tests Node Swagger/contract/harness đạt; source inventory345 đạt, diff check đạt.
Suite integration có race detector đạt: postgres77.035s, transport5.851s,
cmd/server65.179s (gồm regression auth/catalog và HTTP profile thật). Ca timeout503
được chạy riêng sau cùng đạt2.397s; root-cleanup check đạt0.501s. Không chạy các
lượt DB song song. Lượt trước đã sửa lỗi test đếm12 thay11 trường DTO và dùng nhầm
helper catalog cấm Set-Cookie cho logout; không thay contract để làm test xanh.
Sub-agent review chỉ đọc hai lượt, các finding đã xử lý và không còn điểm chặn.
Chưa commit/push hoặc tự chuyển Jira Done; Done sau review/bàn giao/merge theo workflow.

## DATN-30 — M2 bỏ username, sequence và email unique

Nguồn: [SQL plan mục 2/4/6–7](../../docs/proposals/backend-organization-accounts-profile-sql-plan.md),
[proposal OAP-A01/A02](../../docs/proposals/backend-organization-accounts-profile.md#4-các-lựa-chọn-đã-duyệt--15112026).
Đầu ra OAP-T08 cho OAP-T09; không có API create/next-code/Gmail mới trong đợt này.

### Schema và consumer

- Head `000005_oap_m2`: DROP `users.username` không CASCADE; giữ mọi cột/dữ liệu khác.
  UNIQUE `lower(btrim(employees.email))` áp dụng mọi trạng thái, không sửa raw email.
- `WithM2Preflight` được nối ở cmd/migrate, localdev và runner test; kiểm trước khi
  driver ghi dirty. Kiểm metadata M1, PK/FK/unique, default/trigger và dependency
  username, email đơn hợp lệ/unique/parity Go–PG, mã legacy/overflow. Không in email,
  username/hash hoặc tự sửa dữ liệu sai. SQL function body được rà bổ sung vì không
  phải mọi dependency động đều có pg_depend; kiểm kê consumer ngoài repo vẫn là
  đầu vào bắt buộc của một đợt chuyển đổi DB thật.
- Cửa sổ nâng cấp yêu cầu dừng consumer ghi suốt preflight và DDL; advisory lock
  chỉ tuần tự hóa migration runner. Không tuyên bố zero downtime. File SQL up được
  chạy qua runner có preflight, không copy chạy rời bỏ bước Go kiểm email.
- `employee_code_seq`: BIGINT/min1/increment1/no cycle/cache1. Khởi tạo một lần bằng
  max phần số `EMP[0-9]+` +1; giữ `EMP001`, `EMP000001`, `E001` và mã khác nguyên vẹn.
  Dải E không tham gia max. Max BIGINT legacy bị từ chối vì không còn số tiếp theo.
- `postgres.AllocateEmployeeCode(ctx, tx)` cấp bằng nextval trong transaction của
  caller. `FormatEmployeeCode` đệm tối thiểu6 chữ số, không cắt số dài.
  `PreviewEmployeeCode` chỉ đọc last_value/is_called, không tiêu thụ hoặc giữ chỗ.
  Collision ngoài allocator vẫn bị UNIQUE(employee_code) chặn; không skip/retry ngầm.
- Seed dùng email tổng hợp tường minh; user mới ghi legacy EMPLOYEE, quyền từ
  user_roles tường minh. Seed nhận diện email/code legacy, dừng nếu hai khóa xung
  đột, giữ credential/quyền/hồ sơ hiện có. User mới lấy mã qua allocator; chạy lại
  không nextval hoặc reset sequence cho người đã có. SQL kỹ năng nối bằng email.
  Đây là seed dev, không phải chức năng cấp mật khẩu tạm/Gmail.
- Localdev/fixture head không dùng username. Fixture lịch sử v1/v2/v3/M1 vẫn giữ
  schema lịch sử; các ca auth phòng vệ email trùng được pin M1 vì M2 cấm tạo trạng
  thái đó. Ca auth thường và các API danh mục/hồ sơ chạy head mới.
- Đã rà frontend: prototype tài khoản còn model/form username riêng, chưa nối DB
  và không được mở trong mode server. Việc nối UI tài khoản thuộc OAP-T12; không
  biến demo thành consumer M2. `autocomplete="username"` ở form email là hint
  trình duyệt, không phải cột/định danh đăng nhập thứ hai.

### Down, dirty và phục hồi

Down chỉ cho DB rỗng chưa từng cấp số (`is_called=false`). DB có dữ liệu, hoặc rỗng
nhưng từng nextval kể cả rollback, phải từ chối. Không tạo username giả/backup PII
trong schema hoặc reset high-water mark. DDL từ chối rollback nguyên khối nhưng
golang-migrate vẫn giữ target4 dirty khi down lỗi; up lỗi giữ5 dirty. Không force
version/retry mù. Các ca lỗi chỉ chạy trên case hy sinh do chính lượt test tạo.

Down không phục hồi username cũ. Với dữ liệu thật phải có backup được xác minh,
consumer tương thích và kế hoạch restore/sửa tiến thêm riêng; nếu restore cần giữ
high-water mark của số đã cấp. Không gọi down/up là backup hoặc quyền áp Neon.

### Lệnh và phạm vi kiểm chứng

Tại root, DB tests chạy tuần tự, guard xác minh PG18/owner/cluster trước mỗi lượt:

```sh
go vet ./...
go test ./... -count=1
go build ./...
python3 scripts/authdb/manage.py check
python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./migrations ./internal/repository/postgres ./internal/transport/http ./cmd/server ./cmd/seed ./cmd/localdev -count=1 -timeout=5m
node --test scripts/check-harness.test.mjs
node scripts/check-harness.mjs
git diff --check
```

| Kiểm chứng | Nguồn test |
|---|---|
| Rỗng qua lịch sử tới M2, unused down/up, legacy snapshot và ID/link/hash/email/status/role/session/code giữ nguyên | `TestM2EmptyAndUnusedDownUp`, `TestM2LegacyPreservesDataAndStartsAfterEMPNumbers` |
| Email lỗi/trùng, dependency/metadata drift, overflow, dải EMP và mã dài | `TestM2PreflightRejectsBeforeDirty`, `TestM2SequenceInitialRange` |
| Parity trên collation DB test thật, giữ raw email; mismatch phải dừng trước dirty | `TestM2EmailNormalizationParityUsesActualPostgres` |
| Hai connection/barrier cấp số riêng, transaction mở và rollback mất số | `TestM2NextvalRollbackAndConcurrentConnections` |
| Preview/format/collision/rollback, mã lớn và hết BIGINT | `TestEmployeeCodeAllocatorPreviewAndRollback`, `TestEmployeeCodeExhaustionAndLargeFormat`, `TestFormatEmployeeCode` |
| Down từ chối atomic, lỗi sau DROP phục hồi cả giá trị username, dirty driver thật | `TestM2DownRefusesDataAndUsedSequenceAtomically`, `TestM2DDLFailureRollsBackAndLeavesDirty` |
| Seed head chạy lại, dữ liệu legacy và lỗi khóa nhận diện, localdev guard+prepare | Tests `cmd/seed`, `TestPrepareM2Localdev` |
| Năm auth API, danh mục và self-profile trên schema mới | Suite `cmd/server`, `internal/repository/postgres`, `internal/transport/http` |

Không nhận test allocator là bằng chứng transaction create/account concurrent đã
hoàn thành (OAP-T09). Không kiểm UI/browser mới, gửi thư, Neon/restore backup thật
hoặc mọi Unicode/collation có thể có. DB gốc phải giữ0 bảng; cleanup chỉ các case
do lượt test tạo, không dọn auth_dev đang chạy. Không đọc `.env`/fallback DB khác.

### Bằng chứng thực chạy — 17/11/2026

- Guard xác nhận PostgreSQL18, DB `ai_staffing_auth_test`, owner `ai_staffing_test`,
  system identifier `7705732357505970219`, public0 bảng. Docker Desktop/container
  test được khởi động lại vì ban đầu chưa chạy; không chuyển sang DB khác.
- `go vet ./...`, `go test ./... -count=1`, `go build ./...`: đạt. Sau sửa lỗi trả về
  seed để không wrap driver thô, chạy lại vet/build, unit seed và ba seed integration:
  đạt. Không thay mật khẩu hoặc cấp quyền tài khoản thật.
- Lượt integration cuối đủ sáu package trong lệnh trên đạt: migrations15.162s,
  postgres23.685s, transport1.464s, server14.296s, seed1.057s, localdev0.862s.
  Gồm nâng rỗng/legacy, preflight/drift, down/dirty, rollback/sequence, năm API auth,
  danh mục và hồ sơ. Tests HTTP gọi adapter PostgreSQL thật; không dùng mock để
  nhận persistence pass. Lượt này không bật race detector; race cấp số dùng hai
  connection/barrier thật, các transaction auth dùng bộ test hiện hữu.
- Chạy riêng corpus normalization bằng `run -- go test -tags=integration -p 1
  ./migrations ./cmd/seed ./internal/repository/postgres -run
  '^(TestM2EmailNormalizationParityUsesActualPostgres|TestSeed.*|TestAuthRepositoryRootRemainsEmpty)$'
  -count=1 -v -timeout=2m`: đạt. DB test UTF8, libc provider `c`, ctype/collation
  `en_US.utf8`; bốn fixture ASCII/case/padding, tiếng Việt, Sigma và dotted-I khớp.
  Đây không phải bảo đảm mọi Unicode/collation. Các ca tab/NBSP padding lệch được
  preflight từ chối trước dirty, không viết lại raw email hoặc đổi auth.
- Root-cleanup test sau cùng đạt: DB gốc0 bảng và không còn auth_case. Localdev
  `prepare` được kiểm trên DB case riêng, không restart server đang dùng cổng8080.
- Harness14 tests, kiểm kê359 nguồn và `git diff --check`: đạt. Không có thay đổi
  migration000001–000004, dependency hoặc frontend production trong đợt này.
- Sub-agent review SQL trước DB và toàn diff sau bổ sung tests; các góp ý về
  PG18 NOT NULL dependency, metadata drift, seed email canonical và snapshot
  username khi DDL lỗi đã xử lý. Không còn góp ý chặn trên snapshot bàn giao.

Chưa commit/push hoặc tự chuyển Jira Done; Done sau review/bàn giao/merge theo workflow.

## DATN-31 — Transaction tạo tài khoản và hồ sơ

Nguồn: [proposal mục 6.3/6.4 F2](../../docs/proposals/backend-organization-accounts-profile.md#63-dto-và-quy-tắc-input),
[SQL plan mục 5/7](../../docs/proposals/backend-organization-accounts-profile-sql-plan.md#5-truy-vấn-khóa-và-transaction).
OAP-T09 dùng schema M2, không thêm migration, endpoint/Swagger, Gmail hoặc UI.

### Đầu ra cho OAP-T11

- `organization.NewAccountService(repo).Create(ctx, input)` validate/chuẩn hóa theo
  contract, sinh32 byte `crypto/rand`, base64url không padding và bcrypt DefaultCost
  ngoài transaction. Role chỉ ADMIN/HR/DIRECTOR, không trùng; trả thứ tự ổn định,
  role rỗng hợp lệ. Nullable giữ đúng nghĩa; pointer/slice được copy trước chờ DB.
- `CreateAccountInput.ActorID` phải do caller lấy từ principal đã xác thực; không
  là field client được tự chỉ định. Handler tương lai phải allowlist DTO và giữ
  auth/limited/ADMIN/Origin/CSRF/body gates. Validation service không thay các gate HTTP.
- `postgres.NewAccountRepository(db, timeout)` dùng READ COMMITTED, khóa user rồi
  employee actor; recheck ACTIVE/cờ/ADMIN từ user_roles. Thay quyền/status về sau
  phải theo cùng giao thức khóa user trước; test mutation giả không chứng nhận
  đã có API quản trị đổi quyền/khóa.
- Catalog được KEY SHARE theo department → position → job_level, giữ đến commit.
  Thiếu catalog là ErrValidation. Adapter kiểm email canonical khớp lower(btrim)
  của DB hiện tại; không đổi normalization auth để vượt collation khác.
- Cấp sequence rồi INSERT user(legacy EMPLOYEE, ACTIVE, true), employee(ACTIVE),
  user_roles và một audit employee/CREATE_ACCOUNT cùng transaction. Audit schema1
  chỉ IDs/code/catalogIDs/roles, actor là Admin; không credential/email/phone/address.
  ID user và employee độc lập. Không cấp session, không gửi email trong transaction.
- Repository chỉ trả `AccountProjection` sau commit xác nhận. Service trả
  `CreateAccountResult.Account` cùng credential nội bộ qua
  `PasswordForDelivery().RevealForDelivery()` cho bước mail được giao sau này.
  Result/command/TemporaryPassword chặn JSON và redact String/GoString/slog;
  không marshal toàn result hoặc log plaintext/hash. Go GC không bảo đảm xóa mọi
  bản sao string trong RAM; credential không được persist ngoài hash DB.
- Kết quả lỗi luôn zero account/credential. Error được sanitize, không wrap raw
  driver vào service. Có ErrValidation/Unauthenticated/PasswordChangeRequired/
  Forbidden/EmailConflict/Unavailable/Persistence và marker ErrCommitUnknown.
  Duplicate chỉ map EMAIL_CONFLICT với hai constraint email đã biết; collision
  mã ngoài allocator/lỗi constraint khác không bị báo nhầm email.

### Commit, timeout và ghép mail

Chỉ ErrTxCommitRollback hoặc PgError severity ERROR thuộc lỗi xác nhận hủy
transaction (class23, 40001/40P01/25P02) được xem là thất bại đã xác định. Lỗi driver,
class08/FATAL, 40003 hoặc phản hồi không rõ giữ ErrCommitUnknown kèm loại lỗi an
toàn. Không thấy phản hồi commit không có nghĩa chắc chắn chưa ghi.

OAP-T11 chỉ được gửi mail sau kết quả Create thành công; nếu lỗi/commit chưa xác
định thì không gửi, không tự retry transaction, phải đối soát bằng account list
khi endpoint đó được nối. Không có mail.status hoặc account_created=true được
bịa từ lỗi trong ticket này. Không có Gmail/mock mail call ở implementation này;
kiểm orchestration thứ tự DB → mail thuộc OAP-T11. Rollback có thể mất số sequence.

### Kiểm thử và lệnh tái hiện

Chỉ dùng PostgreSQL18 test sở hữu, case riêng và chạy DB suite tuần tự:

```sh
go vet ./...
go test ./... -count=1
go build ./...
go test -race ./internal/service/organization ./internal/repository/postgres -count=1
python3 scripts/authdb/manage.py check
python3 scripts/authdb/manage.py run -- go test -race -tags=integration -p 1 ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m
node --test scripts/check-harness.test.mjs
node scripts/check-harness.mjs
git diff --check
```

| Ca | Test chính |
|---|---|
| Input Unicode/length/nullable/email/roles, copy input, random/bcrypt/cancel lỗi sớm | `TestValidateAccountInput`, `TestAccountServiceRejectsBeforeRepository` |
| Bcrypt thật, secret32byte, redaction/JSON, lỗi port không trả credential | `TestAccountServiceBcryptAndSecretBoundary`, `TestAccountServiceFailureDoesNotReturnCredential` |
| User/employee/roles/audit atomic, nullable/populated, audit đúng8 field và giá trị, không session | `TestAccountCreateAtomicProjectionAndAudit` |
| Service + PG thật, generated user_id khác employee_id, hash khớp credential nội bộ | `TestAccountServiceUsesRealBcryptAndCommittedRepository` |
| Actor/status/cờ/legacyrole và catalog thiếu | `TestAccountCreateRechecksActorAndCatalog` |
| Fault từng write user/employee/role/audit, rollback tất cả row, sequence vẫn mất số | `TestAccountCreateWriteFailuresRollbackAllRows` |
| Hai Admin tranh email canonical, N create cấp mã khác nhau | `TestAccountConcurrentCreatesAndEmailCollision`, `TestAccountManyCreatesKeepCodesUnique` |
| Mutation giả theo giao thức khóa đổi actor trong lúc create chờ | `TestAccountActorChangesWhileWaiting` |
| Create/delete thật cả3 catalog, hai thứ tự, Admin cũng đang tham chiếu mục bị xóa | `TestAccountCreateAndCatalogDeleteRealAdapters` |
| Rollback xác nhận, deferred constraint tại COMMIT, commit thật mất phản hồi, lock timeout | `TestAccountCommitOutcomeAndNoFalseSuccess`, `TestAccountCreateLockTimeoutRollsBack` |

Barrier/pg_blocking_pids chứng minh chờ khóa, không sleep ngẫu nhiên. Fault SQL
chỉ trên DB case hy sinh; lỗi commit EOF/08006/40003 được mô phỏng tại biên driver
sau commit PostgreSQL thật, không phải cắt mạng vật lý. Không mock nhận thay bằng
chứng DB; unit fake chỉ kiểm từng tầng service. Không có bằng chứng Gmail/browser,
HTTPS production, Neon hoặc tạo tài khoản qua HTTP/FE trong DATN-31.

### Bằng chứng thực chạy — 17/11/2026

- Nhánh `feature/DATN-31-create-account-transaction` tách từ main sau merge DATN-30
  (`b17ea1e`); không đổi schema/migration, router, dependency hoặc UI.
- Guard xác nhận PostgreSQL18, owner đúng, DB gốc public0 bảng và cluster system
  identifier `7705732357505970219`. Không fallback DATABASE_URL/Neon hoặc đọc `.env`.
- `go vet ./...`, `go test ./... -count=1`, `go build ./...`: đạt. Unit race cho
  organization/postgres đạt (13.124s/1.870s), gồm tám tổ hợp role, biên validation,
  random/bcrypt lỗi sớm, hash thật và redaction/zero-result khi lỗi.
- Lượt đầu atomic/recheck/write-fault trên PG thật đạt8.109s. Lượt cuối theo lệnh
  integration `-race` ở trên đạt: postgres156.246s, transport6.886s, server96.667s.
  Gồm create concurrency, canonical email collision, actor/cờ/status thay đổi khi
  chờ, ba catalog delete races, commit ambiguity và regression auth/catalog/profile.
- Service+PG tạo employee_id khác user_id được kiểm liên kết và hash thật. Audit
  được so exact payload/schema_version/IDs/catalogIDs/roles, không chỉ đếm field.
  Không fake thay bằng chứng persistence. Có rollback thật bằng deferred constraint;
  lost-reply EOF/08006/40003 là fault wrapper sau commit thật, không mất mạng vật lý.
- Sau toàn bộ suite, chạy riêng qua runner:
  `go test -tags=integration -p 1 ./internal/repository/postgres -run '^TestAuthRepositoryRootRemainsEmpty$' -count=1 -timeout=1m`:
  đạt0.707s, DB gốc0 bảng và không còn auth_case. Không dọn DB localdev khác.
- Harness14 tests, kiểm kê368 nguồn và `git diff --check`: đạt. Không stage/commit
  trong bước triển khai; các file mới vẫn cần được đưa vào commit khi người dùng giao.
- Sub-agent review SQL/port/secret/commit trước DB và review code/tests/runbook cuối.
  Góp ý commit classifier quá rộng đã sửa theo allowlist xác nhận rollback; bổ sung
  unequal IDs, exact audit và fault commit thật. Không còn góp ý chặn được phát hiện.

Chưa commit/push hoặc đổi Jira Done. Bàn giao lõi transaction cho OAP-T11; phần mail
adapter thuộc OAP-T10, HTTP/orchestration và account list thuộc OAP-T11, UI thuộc
OAP-T12. Done sau review/bàn giao/merge phù hợp workflow, không nhận toàn luồng F2 đã xong.

## DATN-33 — API Admin cấp và tra cứu tài khoản

Nguồn [proposal mục 6.2–6.5](../../docs/proposals/backend-organization-accounts-profile.md), [OpenAPI](../../docs/api/organization-accounts-profile.openapi.json) và [SQL plan mục 4–5/7](../../docs/proposals/backend-organization-accounts-profile-sql-plan.md).

### Cấu hình chạy

Server và localdev yêu cầu `GMAIL_OAUTH_CREDENTIALS_FILE` trong môi trường tiến trình, trỏ tới file credential riêng đã cấp quyền theo [runbook Gmail](../../cmd/gmail-setup/README.md). Khởi tạo sender kiểm cấu hình file trước khi dựng router. Test DB inject MailSender riêng trong composition root để kiểm mail fault trên DB sở hữu. Lượt kiểm refresh/gửi chẩn đoán thật đã có tại DATN-32.

Ví dụ sau dùng file local ngoài repo đã được chủ hộp thư chuẩn bị. Chạy từ thư mục gốc, dùng môi trường PostgreSQL test riêng đã được guard.

```sh
export GMAIL_OAUTH_CREDENTIALS_FILE="$HOME/.config/taskassign/gmail-oauth.json"
go run ./cmd/gmail-setup check
pnpm --dir web dev
```

Swagger dev phục vụ tại `http://127.0.0.1:8080/docs/`. Admin phiên thường có thể xem mã dự kiến và danh sách. POST tạo tài khoản cần Origin/CSRF và JSON theo canonical spec. Việc thực hiện POST với Gmail thật cần hộp thư nhận đã được chủ hộp thư cho phép.

### Hành vi bàn giao OAP-T12

- `POST /api/employees` lấy actor từ phiên, kiểm quyền ADMIN và dữ liệu allowlist, dùng transaction DATN-31. Commit được xác nhận rồi mới gọi MailSender một lần. Ba trạng thái mail accepted/failed/unknown đều kèm HTTP201, account_created=true và account đã lưu.
- Lỗi trước xác nhận commit trả lỗi HTTP theo nguồn. Marker commit chưa xác định không cấp kết quả thành công và không gọi mail. Admin đối soát qua danh sách trước thao tác tiếp theo.
- `GET /api/employees/next-code` kiểm actor trong snapshot đọc, dùng lại PreviewEmployeeCode, trả is_preview=true. Preview không tiêu số hoặc giữ chỗ, mã đã commit là mã chính thức.
- `GET /api/users` đọc count/items và quyền trong một snapshot REPEATABLE READ, tìm literal trên employee_code/fullname/company_email và sắp user_id tăng dần. Role đọc user_roles. Email chuẩn hóa ở projection. Account list chỉ chứa allowlist account, kết quả thư chỉ có ở response create hiện tại.
- Account và employee có ID độc lập. List phát hiện user thiếu employee theo SQL plan và trả lỗi dữ liệu. Trang rỗng vẫn giữ total của snapshot.
- Mutation yêu cầu phiên thường, ADMIN, Origin/CSRF rồi kiểm Content-Type, body thực tối đa8192 byte, JSON strict và validation. Response account dùng no-store và không cấp cookie đăng nhập.

### Lệnh kiểm tra

```sh
go vet ./...
go test ./... -count=1
go build ./...
go test -race ./internal/service/organization ./internal/transport/http -count=1
python3 scripts/authdb/manage.py check
python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m
node --test scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs
node scripts/check-harness.mjs
git diff --check
```

HTTP giả chỉ kiểm handler/service. Kiểm HTTP+PostgreSQL dùng DB case riêng và fake MailSender để quan sát commit từ connection độc lập, fault response và số lần gửi. Mailbox thật tái sử dụng bằng chứng DATN-32. Kiểm UI cấp tài khoản thuộc OAP-T12, tổng hợp xuyên luồng thuộc OAP-T13.

### Bằng chứng thực chạy — 18/11/2026

- Môi trường Go1.26.5/macOS ARM64. Guard xác nhận PostgreSQL18, DB ai_staffing_auth_test, user/owner ai_staffing_test, public0 bảng và system identifier7705732357505970219 trước integration.
- Đạt `go vet ./...`, `go test ./... -count=1`, `go build ./...`. Unit race organization/postgres đạt12.509s/1.555s; organization/HTTP đạt16.159s/6.664s.
- Lượt tập trung HTTP+PG thật cho account đạt transport8.631s và server3.156s. Fake MailSender đọc user/hash/roles/audit qua connection độc lập trước gửi. Cả accepted/failed/unknown giữ201. FK/email/quyền/cờ/active lỗi không gửi. Fault wrapper sau commit PostgreSQL thật trả EOF, HTTP503 và list đối soát được account.
- Toàn bộ integration theo lệnh ba package ở trên đạt postgres40.970s, transport6.083s, server15.225s. Gồm regression auth/catalog/profile và các ca tạo tài khoản trước đó.
- List/preview có ca cùng snapshot giữa count/items, literal search, roles tổ hợp, thiếu employee, actor mất quyền, page rỗng và preview không tiêu số. Hai connection/barrier kiểm preview trong lúc create chờ commit, thêm boundary sequence hết BIGINT.
- Kiểm bổ sung composition root với phiên đăng nhập thật, CSRF, create/list/preview, email conflict và limited gate đạt1.486s. Swagger canonical từ router ứng dụng có16 operation OAP, gồm next-code tĩnh và profile theo ID.
- Parser từ chối JSON field lạ/trùng/null sai kiểu, trailing JSON, UTF-8 sai và surrogate đơn, chấp nhận surrogate đôi hợp lệ. Body thực8192/8193 byte không Content-Length được kiểm. Response account no-store, không Set-Cookie và không chứa credential/phone/address.
- Contract/harness32 tests và Swagger bootstrap5 tests đạt. Harness kiểm kê388 nguồn, `git diff --check` đạt.
- Sub-agent review service/read/composition và transport hoàn tất. Các góp ý phân loại lỗi preview, orphan user và giữ kết quả commit khi mail lỗi đã xử lý. Góp ý surrogate đơn được sửa bằng helper hiện có và test tương ứng đạt.
- Gmail thật tái sử dụng bằng chứng DATN-32. Fault mail mô phỏng tại port, lost commit reply mô phỏng sau commit thật. Kiểm mất mạng vật lý, UI cấp tài khoản và chuỗi nhận mật khẩu tạm trên UI tiếp tục ở OAP-T12/OAP-T13.

- Sau suite, chạy qua runner `go test -tags=integration -p 1 ./internal/repository/postgres -run '^TestAuthRepositoryRootRemainsEmpty$' -count=1 -timeout=1m` đạt0.352s. DB test gốc còn0 bảng và không sót DB auth_case.

## DATN-35 — ma trận bàn giao PostgreSQL OAP

Mã ca từ [SQL plan mục 7](../../docs/proposals/backend-organization-accounts-profile-sql-plan.md#7-ma-trận-postgresql-test-dự-kiến); không tạo thêm contract SQL. [Runbook demo](../../docs/api/auth-runbook.md#7-datn-35--oap-t13--demo-xuyên-ui-db-và-gmail) giữ luồng UI/Gmail và giới hạn FR. Kết quả dưới là trên PG18 local sở hữu, không phải Neon hoặc dữ liệu nhân viên thật.

Ngày 18/11/2026, từ snapshot DATN-34 `7377dc6`, đã chạy:

```sh
python3 scripts/authdb/manage.py check
python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./migrations ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m
python3 scripts/authdb/manage.py check
```

Guard trước/sau: PG18, đúng DB/user/owner, system ID7705732357505970219, public0 bảng. Bốn package đạt lần lượt20.494s/49.963s/4.465s/10.969s. Đây là chạy lại các ca hiện có, không cộng số unit tests vào số luồng E2E. Chạy suite khi không có suite DB/harness khác; browser localdev dùng DB auth_dev riêng, không mutation từ suite.

| Mã | File/test tái hiện chính (package migrations hoặc internal/repository/postgres) | Kết quả/phạm vi |
|---|---|---|
| OAP-DB01 | `oap_m1_integration_test.go`: `TestM1EmptyAndSeed`; `oap_m2_integration_test.go`: `TestM2EmptyAndUnusedDownUp` | Đạt đường rỗng đến head M2; seed đúng schema. |
| OAP-DB02 | `TestM1LegacyAndDownUp`, `TestM2LegacyPreservesDataAndStartsAfterEMPNumbers` | Đạt fixture legacy tổng hợp; ID/link/role/hash/email/session/code được bảo toàn theo snapshot, chỉ username bị bỏ ở M2. |
| OAP-DB03 | `TestM1PreflightRejectsBeforeMigration`, `TestM2PreflightRejectsBeforeDirty` | Đạt drift/duplicate/malformed/dependency guard; không cleanup tự động. |
| OAP-DB04 | `TestM1LegacyAndDownUp`, `TestM1PreflightRejectsBeforeMigration` | Đạt mapping tổng hợp và thiếu mapping bị chặn. Không xác nhận mapping dữ liệu Neon. |
| OAP-DB05 | `organization_catalog_integration_test.go`: `TestCatalogCRUDAndAudit`, `TestCatalogCurrentActorAndReferences`; concurrency: `TestCatalogConcurrentDuplicateAndStale` | Đạt ba danh mục, trim/code point/duplicate/no-op/stale/audit và reference. |
| OAP-DB06 | `organization_account_concurrency_integration_test.go`: `TestAccountCreateAndCatalogDeleteRealAdapters`; catalog concurrency: `TestCatalogDeleteWithAdminReferenceRace` | Đạt adapter create thật cạnh tranh delete và actor Admin có reference; khóa/FK không orphan. |
| OAP-DB07 | `TestAccountConcurrentCreatesAndEmailCollision`, `TestAccountManyCreatesKeepCodesUnique`, `TestEmployeeCodeAllocatorPreviewAndRollback`, `TestM2NextvalRollbackAndConcurrentConnections` | Đạt create nhiều connection, code unique, rollback có gap và preview không tiêu số. |
| OAP-DB08 | `TestM2SequenceInitialRange`, `TestEmployeeCodeExhaustionAndLargeFormat`, `TestAccountPreviewDoesNotConsumeAndSurvivesRepositoryRestart` | Đạt dải legacy/độ dài/overflow/restart repository. Không nhận đây là diễn tập deploy nhiều instance production. |
| OAP-DB09 | `TestAccountCreateWriteFailuresRollbackAllRows`, `TestAccountCommitOutcomeAndNoFalseSuccess`; HTTP: `TestAccountHTTPPersistsBeforeMailAndKeeps201ForMailOutcomes`, `TestAccountHTTPUnknownCommitSkipsMailAndCanReconcileByList` | Đạt write/audit/commit fault; connection độc lập thấy commit trước mail; ambiguity không gửi/retry. Mail port fake. |
| OAP-DB10 | `TestAccountActorChangesWhileWaiting`, `TestCatalogActorRecheckAfterLock`, `TestAccountReadsRecheckCurrentActor` | Đạt recheck role/status/cờ hiện hành dưới khóa/snapshot. Không có UI sửa quyền tài khoản cũ. |
| OAP-DB11 | `organization_profile_integration_test.go`: `TestSelfProfileProjectionAndCurrentState`, `TestSelfProfileConsistentSnapshot`, `TestSelfProfileMissingBasisAndMalformedData`; server: `TestSelfProfileApplicationPostgres` | Đạt self/other/missing, nullable/role không nhân row/email projection; Admin không bypass. |
| OAP-DB12 | `TestM1DDLFailureRollsBackAndLeavesDirty`, `TestM2DownRefusesDataAndUsedSequenceAtomically`, `TestM2DDLFailureRollsBackAndLeavesDirty` và down/up ở DB01/02 | Đạt trên DB case hy sinh. Down M1 mất dữ liệu bổ sung; M2 không phục hồi username, từ chối data/sequence đã dùng, không force dirty. |
| OAP-DB13 | `auth_integration_test.go`, `TestAccountApplicationRealSessionsAndComposition`, seed M1/M2; regression localdev DATN-34 | Đạt năm auth API và composition account không username. Fixture localdev đã kiểm DATN-34; không gọi suite bốn package là có chạy package localdev. |
| OAP-DB14 | `TestM2EmailNormalizationParityUsesActualPostgres`, `TestFindLoginIdentity`, `TestSelfProfileProjectionAndCurrentState` | Đạt Go–PG18 test/collation test và giữ email legacy gốc; collation/dữ liệu đích chưa kiểm. |
| Snapshot/literal | `TestCatalogPaginationSnapshotAndLiteral`, `TestAccountListSnapshotLiteralRolesAndProjection`, `TestSelfProfileConsistentSnapshot`, `TestAccountPreviewWhileCreateWaitsForCommit` | Đạt count/items cùng snapshot, q literal, projection và preview cạnh tranh. |

Fault injection lỗi DDL/write/commit không thay kiểm mất mạng vật lý. Race dùng connection/barrier và lock timeout có kiểm soát, không sleep ngẫu nhiên. Log/evidence chỉ tên ca/kết quả, không SQL parameter chứa dữ liệu cá nhân, password/hash/token hoặc raw error provider.

Migration/rollback vẫn dùng runner và bộ case hiện có, không chạy `cmd/migrate` loader production hoặc down trên DB demo đang dùng. Dừng app bằng Ctrl+C dọn đúng auth_dev vừa tạo, giữ container/volume. Giới hạn crash/kill9 và xử lý dirty xem các mục M1/M2 ở trên; không xem down là backup, không reset sequence để dùng lại mã.

## DATN-38 — PCA-T02 — schema đề nghị hồ sơ và thông báo

Ngày 22/11/2026. Schema head mới là `000006_pca_profile_change`; nguồn thiết kế tại [SQL plan mục9](../../docs/proposals/backend-profile-change-approval-sql-plan.md#9-sql-cụ-thể-cho-task-triển-khai). Đợt này chỉ bàn giao schema/fixture/guard, chưa có use case hoặc API PCA và chưa publish contract PCA vào Swagger.

- `migrations/000006_pca_profile_change.up.sql`: hai bảng request/idempotency, JSON/state/FK/index và notification nhận theo `users.user_id`. Không đổi migrations000001–000005, không reset identity notification hoặc employee code sequence.
- `migrations/pca.go`: preflight dưới advisory lock của runner, trước khi ghi dirty; chỉ nhận nguồn M2 sạch, notification legacy rỗng, metadata/dependency đã biết và namespace đích chưa bị chiếm. Nối ở `cmd/migrate`, `cmd/localdev`, `internal/testutil/authdb`; không chạy CLI production để kiểm test.
- DDL kiểm lại bảng rỗng dưới `ACCESS EXCLUSIVE`; notification được ghi sau preflight khiến up từ chối, bảo toàn dòng/schema legacy. Preflight không thay việc dừng consumer trong cửa sổ migration và không chứng minh đã kiểm mọi dạng SQL động của consumer ngoài repo.
- Down khóa cả ba bảng, từ chối khi có bất kỳ request/key/notification hoặc dependency local lạ (index/check/trigger/function consumer/cột thêm). Chỉ DB case rỗng mới về M2; khôi phục cột/FK/index notification legacy mà giữ sequence. Down không phải backup và không khôi phục lịch sử đã mất.
- `NewFixture` tự nâng tới head6; fixture lịch sử và test M1/M2 tiếp tục dựng đúng version. Localdev/test seed dùng schema đích không cần notification legacy. Không có writer notification trong app/seed hiện tại; consumer PCA sẽ được bổ sung ở DATN-39/42/47.

### Tái hiện và bằng chứng

Chạy từ gốc repo, Docker Desktop mở. `up` chỉ dùng runner đã chốt; nếu state/owner không khớp thì dừng, không xóa tài nguyên để vượt guard.

```sh
python3 scripts/authdb/manage.py up
python3 scripts/authdb/manage.py check

go vet ./...
go test ./... -count=1
go build ./...

python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./migrations ./internal/repository/postgres ./internal/transport/http ./cmd/server ./cmd/localdev -count=1 -timeout=5m

node --test scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs
node scripts/check-harness.mjs
git diff --check
```

Không chạy nhiều lượt DB song song. Mỗi ca tạo/dọn đúng DB `auth_case_*` riêng theo guard; DB gốc giữ0bảng. Không reset lượt `pnpm dev` hoặc DB lạ. Nếu DDL lỗi: dữ liệu/cấu trúc rollback, metadata runner vẫn có thể dirty. Không force version; giữ DB case lỗi để đối chiếu nếu cleanup không hoàn tất. `stop`/`destroy` chỉ theo điều kiện chung ở đầu runbook, không dùng để sửa metadata.

| Bằng chứng | Test thực trong `migrations/pca_integration_test.go` |
|---|---|
| PCA-DB01 | `TestPCAEmptyDownUp`: empty→head6, up lặp, down5/up6, identity high-water; `TestPCAJSONStateFKAndIndexes`: JSON/types/keyset/schema/NULL/giới hạn Unicode, trạng thái/resolver/time/reason, FK/unique/hash/expiry và định nghĩa index |
| PCA-DB02 | `TestPCALegacyPreservesAuthOAP`: legacy user101/employee201, roles/session/audit/email padding/hash/code giữ nguyên; recipient201 bị FK từ chối, recipient101 hợp lệ. `TestPCAPreflightRefusesBeforeDirty`: notification có dữ liệu hoặc schema/dependency/namespace lệch giữ version5 sạch và dữ liệu nguyên vẹn |
| PCA-DB03 | `TestPCADownRefusesEachHistoryTable`, `TestPCADownRefusesUnknownDependency`: không mất lịch sử/dependency; `TestPCADDLFailureAndTOCTOUGuard`: fault DDL và hai connection qua advisory barrier sau preflight, không sleep; `TestPCADownDDLFailureRollsBack`: rollback down lỗi giữa DDL |
| Dirty thực | Up DDL lỗi: version6 dirty; down DDL/guard lỗi: version5 dirty. Retry up bị `ErrDirty`; không force. Preflight từ chối trước DDL: version5 sạch |
| Regression | Bộ integration migration lịch sử, PostgreSQL repository, HTTP, server năm auth API/catalog/profile/account, localdev prepare cùng chạy trên head6 |

Kết quả thực chạy 22/11/2026: Go1.26.5 darwin/arm64, Node24.15.0, Python3.9.6, PostgreSQL18 test do runner xác minh owner; DB gốc0bảng sau suite. Các lệnh Go vet/unit/build ở trên đạt; integration đủ5package đạt. Sau chỉnh guard theo review, nhóm `./migrations -run '^TestPCA'` được chạy lại bằng cùng runner và đạt (6.162s); thay đổi cuối chỉ ở guard down/assertions, không đổi repository/HTTP. Test prepare localdev theo tên head mới cũng đạt. Bộ42 Node tests đạt, harness417file đạt, diff check đạt. Hai finding P2 của sub-agent (namespace collision và down tự loại dependency lạ) đã sửa, bổ sung ca tái hiện, retest và sub-agent xác nhận khép ở mức source review; không gọi review là runtime proof. Không lấy Node contract/unit làm bằng chứng PostgreSQL. Fixture toàn dữ liệu tổng hợp; snapshot/hash chỉ ở bộ nhớ, không in hoặc ghi vào evidence.

Giới hạn: không áp Neon, không map notification legacy có dữ liệu, không kiểm browser/mail/HTTPS/deploy trong task schema. UUID/TTL24h/result-reference đầy đủ, current actor/self/HR, tính bất biến terminal và atomic request→notification→key thuộc use case các ticket sau; các CHECK hiện tại không thay các quy tắc đó. Chưa nhận PCA-DB04–12 pass từ schema.

## DATN-39 — PCA-T03 — gửi và đọc đề nghị hồ sơ

Ngày22/11/2026. Dùng schema000006 từ DATN-38; không thay migration. Nhánh `feature/DATN-39-profile-request-use-cases` tiếp nối commit schema đã bàn giao. Tại thời điểm bắt đầu, PR DATN-38 còn mở; việc merge/bàn giao tuân workflow, không tự merge từ task này.

### Đầu ra và cách nối tiếp

- `internal/service/profilechange`: ports `Repository.Submit/List/Read`, `Service`, input/command/projection typed, validation và canonical hash. `Values` giữ sự hiện diện của trường và NULL; before không normalize/giới hạn lại legacy; proposed/explanation normalize theo proposal6.3. UUIDv4 nhận cả hoa/thường rồi canonical lowercase. Hash gồm version, employee ID, before/proposed cùng tập khóa và explanation canonical; không hash JSON theo thứ tự key client.
- `internal/repository/postgres/profile_change.go`: READ COMMITTED, chụp recipient trước khóa; users tăng ID rồi employees tăng ID, loại trùng theo linkage nguồn. Recheck actor và recipient dưới khóa; actor mới có HR sau snapshot không được tự thêm vào tập nhận. Tra key trước stale, TTL24h theo thời gian sau chờ khóa; reference phải đúng owner/employee/payload. Ghi PENDING + notification chung + record kết quả đầy đủ trong cùng transaction, không placeholder commit, không sửa official/audit.
- `profile_change_read.go`: self mặc định kể cả HR; scope review cần HR, Admin/Director không bypass. Quyền kiểm trước tải patch; detail chỉ current_values theo keyset patch, không JSONB thô. List chỉ projection nhận diện/trạng thái/thời gian/tên trường; REPEATABLE READ READ ONLY giữ actor/count/items cùng snapshot, page rỗng giữ total.
- Lỗi typed: validation, unauthenticated, password-change-required, forbidden, not-found, stale, key-reused, key-expired, persistence, unavailable và commit-unknown. Lỗi driver không mang SQL/patch ra ngoài. Commit chưa rõ không trả result thành công hoặc tự retry; rollback đã xác nhận không bị gọi unknown.
- Constructor adapter nhận timeout dương từ caller theo convention hiện hữu; khi T05 wire composition root dùng timeout nghiệp vụ10s, rollback cleanup hiện có tối đa5s. Service không import pgx/HTTP. Không dependency mới.

T04 tiếp nhận types/projection và giao thức khóa để triển khai approve/reject; T05 nối handler/DTO/error mapper/middleware. ActorID của port là nguồn nội bộ đã xác thực; **session/cookie/revoke/CSRF/Origin, JSON strict (key trùng/Unicode escape) và body64KiB thực đọc vẫn phải kiểm ở T05**. Unit/service input typed không chứng minh tầng HTTP. Không nhận status/role từ client để thay DB recheck.

### Kiểm và tái hiện

Từ gốc repo, mở Docker Desktop; xác minh PG18/owner/DB gốc0bảng trước integration, một suite DB mỗi lượt:

```sh
python3 scripts/authdb/manage.py check
go vet ./...
go test ./... -count=1
go build ./...
python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./migrations ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m
node --test scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs
node scripts/check-harness.mjs
git diff --check
```

Chạy riêng nhóm mới qua cùng runner khi cần điều tra:

```sh
python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./internal/repository/postgres -run '^TestProfileChange' -count=1 -timeout=5m
```

| Phạm vi bằng chứng | Test thực |
|---|---|
| PCA-DB04 | `TestProfileChangeInputAndSelfFailuresDoNotWrite`, `TestProfileChangeUnicodeBoundariesThroughUseCase`: before NULL/legacy, trim/code point, allowlist, stale/no-op/address sai không ghi; official giữ nguyên; nhiều pending/no HR được phép |
| PCA-DB05 | `TestProfileChangeSubmitReplayAndRead`, `TestProfileChangeSameKeyConcurrentAndRollbackFollower`: cùng key hai connection một tác động, writer rollback thì follower thực hiện một lần; key khác payload/expired conflict; replay sau đổi official/terminal không tạo history/notification mới |
| Key integrity/commit | `TestProfileChangeReadGatesAndReplayIntegrity`, `TestProfileChangeReplayReferenceCannotPointToOtherPayload`, `TestProfileChangeAtomicFaultsAndCommitOutcomes`: lỗi notification/key/deferred commit rollback thật; mô phỏng mất xác nhận sau commit PG thật thì không success/retry, key vẫn đối soát được; missing/wrong reference/hash/TTL fail closed |
| PCA-DB09 phần submit/read | `TestProfileChangeTwoHRActorsLockOrder`, `TestProfileChangeActorRecheckAndExpiryAfterLock`, `TestProfileChangeRecipientRecheck`, `TestProfileChangeNewHRDoesNotExtendSnapshot`, `TestProfileChangeActorGainingHRIsNotNewRecipient`: nhiều role/actor cũng HR, employee ID ngược user ID, actor/recipient đổi quyền/cờ/status trong lúc chờ; không mở rộng recipient snapshot |
| Read projection/snapshot | `TestProfileChangeReadSnapshotAndCorruptJSON`, `TestProfileChangeSubmitReplayAndRead`, `TestProfileChangeReadGatesAndReplayIntegrity`: self/HR, mất HR/limited/inactive, user_id khác employee_id, list/count cùng snapshot có mutation xen giữa, page rỗng, current keyset và JSONB schema hỏng |

Các race dùng connection thật, channel/barrier và kiểm `pg_blocking_pids`; không sleep ngẫu nhiên. Role/status mutation là SQL fixture có kiểm soát theo giao thức khóa, không bằng chứng API quản trị chưa có. Terminal được dựng bằng SQL fixture để kiểm replay/read; không nhận approve/reject của T04 đã chạy. Commit reply lost là fault wrapper quanh commit thật, không tuyên bố đã gây mất mạng vật lý.

Bằng chứng dùng dữ liệu tổng hợp, snapshot/hash/patch không xuất log/evidence. Cleanup vẫn theo runner: chỉ DB case vừa tạo được dọn, không force version/prune/đụng Neon hoặc DB demo. Read actor trạng thái/quyền theo snapshot; writer role/status tương lai phải giữ giao thức user→employee để recheck có hiệu lực.

Kết quả thực chạy22/11/2026 (Go1.26.5 darwin/arm64, PostgreSQL18 test owner đã xác minh): Go vet/unit/build đạt; suite integration4package ở lệnh trên đạt (migrations13.469s, repository27.920s, HTTP3.560s, server8.824s). Sau bổ sung ca biên Unicode, chạy lại toàn nhóm `^TestProfileChange` qua runner đạt10.555s. Bộ42 Node tests đạt, harness424file và diff check đạt. Sub-agent review độc lập không thấy finding P0/P1/P2 chặn; unit validation/hash/redaction/service/error-timeout họ chạy thêm cũng đạt, không chạy DB song song. Tự review đã siết recipient snapshot khi actor mới được cấp HR và reference trỏ nhầm payload, có test PG tái hiện.

Chưa HTTP/UI/browser/Gmail/Neon/deploy, chưa notify kết quả quyết định/mark-read hoặc F4; không nhận trọn PCA-DB09/10/12 hay toàn FR từ task này. Không lấy static/unit hoặc source review thay PostgreSQL. Chưa commit/push/PR trong lượt triển khai; Git bàn giao khi người dùng yêu cầu.

## DATN-40 — PCA-T04 — HR duyệt và từ chối đề nghị hồ sơ

Ngày23/11/2026. Nhánh `feature/DATN-40-profile-request-decisions` tiếp nối nguồn DATN-39 `4b921a8`; không giả định PR phụ thuộc đã vào main chỉ từ trạng thái MERGED. Không sửa schema000006 hoặc migration lịch sử.

### Đầu ra cho DATN-41

- `internal/service/profilechange/decision.go`: `Service.Approve(DetailQuery)` và `Service.Reject(RejectInput)`, thêm hai method vào repository port. ActorID do nguồn nội bộ server cung cấp; reject reason trim Go, 1–2000 code point, không NUL/UTF-8 lỗi hoặc NFC ngầm. Service kiểm result đúng request/status/resolver/reason; approve current phải bằng proposed. Không import pgx/HTTP, không dependency mới; input reject có redaction, không serialize thẳng thành DTO/log.
- `profile_change_decision.go`: READ COMMITTED, đọc ID trước chỉ để lập bộ khóa; users tăng ID → employees tăng ID → request. Loại trùng actor=target; sau khóa kiểm lại linkage, actor ACTIVE/employee ACTIVE/cờ/HR và PENDING. Không tự mở rộng bộ khóa khi linkage thay đổi; dữ liệu liên kết sai fail closed. Target khác actor không bị áp điều kiện phiên actor: user đích locked/limited riêng vẫn cho HR xử lý; employee target inactive chỉ chặn approve bằng `ErrEmployeeInactive`, reject vẫn được.
- Approve so before/current đúng từng trường, NULL khác chuỗi rỗng; ghi đúng giá trị patch và giữ trường vắng. Reject không so before và giữ official. UPDATE kết luận chỉ `WHERE status='PENDING'`, kiểm affected rows; notification owner dùng user_id, nội dung chung không patch/reason/PII. Official + terminal + notification cùng commit, không audit trùng lịch sử request. Terminal qua ports trả `ErrAlreadyResolved`, không đổi resolver/time/reason lần hai.
- Typed errors thêm `ErrAlreadyResolved` → contract409 REQUEST_ALREADY_RESOLVED, `ErrEmployeeInactive` →409 EMPLOYEE_INACTIVE; stale →409 STALE_DATA, còn lại dùng phân loại DATN-39. Commit mất xác nhận giữ `ErrCommitUnknown`, không trả result thành công hoặc tự retry; caller phải đọc chi tiết để đối soát. Timeout do caller cấu hình và rollback cleanup hữu hạn dùng lại convention hiện hành. Không thêm Idempotency-Key cho quyết định.

### Lệnh tái hiện

Từ gốc repo, Docker Desktop mở; runner phải xác minh PG18 test sở hữu và DB gốc0bảng. Chỉ một suite DB mỗi lượt:

```sh
python3 scripts/authdb/manage.py up
python3 scripts/authdb/manage.py check
go vet ./...
go test ./... -count=1
go build ./...
python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./migrations ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m
node --test scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs
node scripts/check-harness.mjs
git diff --check
```

Nhóm quyết định khi cần điều tra riêng, dùng cùng guard và không chạy song song suite trên:

```sh
python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./internal/repository/postgres -run '^TestProfileChangeDecision' -count=1 -timeout=5m
```

| Phạm vi bằng chứng | Test thực ở `profile_change_decision*_integration_test.go` |
|---|---|
| PCA-DB06 | `TestProfileChangeDecisionApproveRejectAndReplay`, `TestProfileChangeDecisionFaultsAndCommitOutcomes`: approve/reject thật, official/kết luận/thông báo atomic, terminal bất biến qua ports, replay submit sau approve thật; notification/deferred commit/known rollback không half-write |
| PCA-DB07 | `TestProfileChangeDecisionConcurrentSingleWinner`: hai HR approve–approve hoặc approve–reject; follower chờ writer rồi nhận already-resolved; chỉ một thông báo kết quả |
| PCA-DB08 | `TestProfileChangeDecisionStaleAndDisjointFields`, `TestProfileChangeDecisionConcurrentPendingFields`: cùng trường stale giữ PENDING, khác trường không conflict giả; NULL khác chuỗi rỗng, phone NULL là xóa; reject stale giữ official |
| PCA-DB09 phần decision | `TestProfileChangeDecisionTargetAndSelfRules`, `TestProfileChangeDecisionValidationAndGates`, `TestProfileChangeDecisionActorRecheckAfterLock`, `TestProfileChangeDecisionChangedOwnerLinkageFailsClosed`, `TestProfileChangeDecisionAndSubmitShareLockOrder`: HR tự xử lý/multiple roles, target employee inactive/user locked/limited riêng, actor mất HR/status/cờ lúc chờ khóa, linkage sai, submit và approve dùng cùng giao thức |
| PCA-DB12 phần commit | `TestProfileChangeDecisionFaultsAndCommitOutcomes`: commit PG thật rồi wrapper làm mất xác nhận; typed unknown không success/retry, đọc detail đối soát được một kết luận/thông báo. Không nhận fault wrapper là mất mạng vật lý |

Các race dùng connection thật, channel/barrier và `pg_blocking_pids`, không sleep ngẫu nhiên. Thay quyền/status/linkage là SQL fixture có kiểm soát, không nhận API quản trị đã chạy; writer tương lai phải tuân giao thức user→employee. Cleanup theo runner chỉ dọn DB case vừa tạo; không force version/prune/reset DB demo hoặc Neon. Không đưa patch/reason/credential/PII vào evidence.

Kết quả thực chạy23/11/2026: Go1.27.4 darwin/arm64, Node24.15.0, Python3.9.6; runner xác minh PostgreSQL18 test đúng owner và DB gốc0bảng trước/sau suite. `up` khởi động container test hiện có, `check` đạt trước suite. Go vet/unit/build đạt; sau thêm ca timeout quyết định, unit hai package service/profilechange và repository/postgres chạy lại đạt. Suite integration4package trong lệnh trên đạt: migrations29.223s, repository64.638s, HTTP6.669s, server17.991s. Bộ42 contract/harness tests đạt, harness429file và diff check đạt. Tự review và sub-agent review code/tests/tài liệu không còn finding P0/P1/P2 chặn; sub-agent unit/diff checks đạt, không chạy DB song song. Review không thay bằng chứng PG; compile-only integration của agent viết tests không được nhận làm runtime proof.

Giới hạn: chưa HTTP/UI/browser/CSRF/Origin/session revoke/strict JSON/body/no-store/Swagger PCA, notification list/mark-read hoặc F4; DATN-41 tiếp nhận integration HTTP. Không có trigger chống writer SQL tùy ý thay terminal; bất biến được bảo đảm qua application ports và constraints hiện hành. Không email, Neon/HTTPS/deploy, không nhận trọn PCA-DB09/10/11/12 hoặc toàn FR từ ticket này. Chưa commit/push/PR hoặc Jira Done trong lượt triển khai.
## DATN-41 — PCA-T05 — năm API đề nghị hồ sơ

Ngày kiểm: **23/11/2026**. Nhánh `feature/DATN-41-profile-request-api`, nguồn đầu vào `origin/main` tại `bacfbc1` đã có DATN-39/40. Đợt này nối năm method/path request ở canonical PCA, không migration/seed/reset, dependency mới, UI, Gmail hoặc Neon. Các ghi nhận DATN-38–40 phía trên giữ phạm vi tại thời điểm bàn giao; trạng thái HTTP hiện hành xem mục này.

**Nguồn thay đổi:** `internal/transport/http/profile_change.go` và `profile_change_json.go` nhận DTO strict/allowlist, lấy actor từ phiên server và map typed errors; `auth_router.go`, `organization_profile.go`, `internal/app/auth.go` lắp service/PostgreSQL timeout10s. Không query DB trong handler. `docs/api/embed.go`, `swagger.go` và `swagger/init.js` công bố đúng năm API PCA từ canonical; giữ năm auth/16 OAP, không mở API thông báo/F4. JSON PCA65536 byte không đổi auth/OAP8192.

**Môi trường thực:** macOS arm64, Go1.27.4, Node24.15.0, Python3.9.6. Runner `check` xác nhận PostgreSQL18, DB test `ai_staffing_auth_test`, owner `ai_staffing_test`,0 bảng public, system ID `7705732357505970219`. Chỉ dùng DB case runner cấp và dọn; không fallback DATABASE_URL/.env/Neon hoặc chạy nhiều lượt DB song song.

**Lệnh đã chạy tại gốc repo, đều đạt:**

```sh
go vet ./...
go test ./... -count=1
go build ./...
python3 scripts/authdb/manage.py check
python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./migrations ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m
node --test scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs scripts/check-swagger.test.mjs
node scripts/check-harness.mjs
git diff --check
```

Suite PG đầy đủ đạt: migrations14.542s, repository36.192s, HTTP5.332s, server9.425s. Unit Go không thay integration;48 tests Node chỉ là kiểm tĩnh/VM, không browser/SQL. Harness kiểm kê nguồn, không kiểm ngữ nghĩa nghiệp vụ. Kết quả kiểm cuối sau cập nhật tài liệu được bàn giao cùng diff.

| Bằng chứng mới | Tầng và kết quả |
|---|---|
| `internal/transport/http/profile_change_integration_test.go`: `TestProfileChangeHTTPFiveRoutesPersistReplayAndProject` | HTTP + auth session/PG thật: submit/list/detail/approve/reject; user_id khác employee_id, NULL, DTO allowlist, review scope, total trang rỗng, replay cả terminal; request/notification không nhân dòng |
| `TestProfileChangeHTTPCurrentServerRightsAndSessions` | Self/HR/Admin/Director; không HR không review/quyết định, ngoài quyền/missing404; limited403, locked/inactive/revoked401 từ server |
| `TestProfileChangeHTTPStrictInputAndActualBodyBoundary` | Không ghi khi JSON/header/query/media/Origin/CSRF/Unicode sai;65536 byte submit được,65537 trả413 không Content-Length; quyết định quá giới hạn không thay PENDING |
| `TestProfileChangeHTTPConflictCodesAndTargetGate` | Key reused/expired, stale submit/approve, target inactive, reject/terminal; giữ official và atomicity |
| `TestProfileChangeHTTPDatabaseFaultAndUnknownCommit` | Trigger lỗi notification rollback request/key/notification; wrapper thực sự commit PG rồi trả lỗi xác nhận cho submit/approve →503, GET đối soát; không tự retry hoặc success giả |
| `TestProfileChangeHTTPDatabaseTimeoutIs503` | Khóa bảng bằng connection thứ hai trước HTTP read, timeout DB100ms →503; không sleep ngẫu nhiên |
| `cmd/server/profile_change_integration_test.go`: `TestProfileChangeApplicationComposition` | HTTP server qua composition root thật: login →CSRF →submit →list/detail, cookie không bị cấp lại và Swagger PCA có đúng năm paths |
| `internal/transport/http/profile_change_test.go`, `scripts/check-swagger.test.mjs` | Parser/gates/error/DTO, inventory health+5auth+16OAP+5PCA; Swagger subset canonical, refs hợp lệ, không notification/F4; interceptor hỗ trợ ID giữa path và chặn route ngoài phạm vi |

**Review:** tự rà diff và sub-agent review read-only parser/quyền/DTO/wiring/Swagger/test; không còn góp ý chặn P0/P1/P2. Đã sửa assertion test đếm14 trường detail thành đúng canonical; review không thay runtime proof.

**Tái hiện và dọn:** chạy đúng lệnh runner bên trên; các fixture dùng dữ liệu tổng hợp và DB case hy sinh, tự dọn theo runner. Không reset DB demo/volume hoặc force migration version. Chạy local và `/docs/` theo [hướng dẫn API](../../docs/api/README.md#datn-41--năm-api-đề-nghị-hồ-sơ); không cần thêm biến môi trường PCA.

**Giới hạn/consumer:** bằng chứng HTTP là httptest/server + PG local, không browser PCA, HTTPS/Neon/deploy hoặc physical network loss. Wrapper commit unknown chỉ chứng minh xử lý mất xác nhận sau commit PG thật; race/atomicity sâu PCA-DB04–09 tái sử dụng suite DATN-39/40 đã chạy lại, không nhận toàn PCA-DB01–12 hoặc toàn FR đã đạt. DATN-43/44 tiếp nhận input/DTO/lỗi canonical; DATN-42 tiếp tục API thông báo, F4 còn ngoài phạm vi. Chưa commit/push/PR/merge hoặc đổi Jira Done trong lượt này.

## DATN-42 — PCA-T06 — thông báo chính chủ

Nguồn thực thi: `internal/service/notification/notification.go`, `internal/repository/postgres/notification.go`, `internal/transport/http/notification.go` và `notification_swagger.go`; wiring tại `internal/app/auth.go`, `internal/transport/http/auth_router.go`. Canonical vẫn là `docs/api/profile-change-approval.openapi.json`; không sửa migration000001–000006 hoặc thêm dependency. Branch `feature/DATN-42-own-notifications` nằm ở worktree riêng `/private/tmp/ai-staffing-DATN-42` để không đè DATN-41 trong checkout chính.

Hai API cho phiên thường, mọi tổ hợp role, recipient từ **user_id** trong phiên:

- `GET /api/notifications?filter=all|unread&page=1&page_size=20`: mặc định all/1/20, tối đa100mục, ID/page INT32 và offset INT64. Unknown/repeated query trả400. `total` sau filter, `unread_count` toàn own; actor/count/items trong REPEATABLE READ READ ONLY. Không đánh dấu đọc khi list; sort created_at DESC rồi notification_id DESC.
- `POST /api/notifications/{id}/read`: Origin/CSRF, JSON UTF-8 body chính xác `{}`, giới hạn byte thực65536 kể cả chunked; không Content-Encoding. User→employee→notification cùng chiều writer; recheck trạng thái/cờ actor, UPDATE own với COALESCE(read_at,clock_timestamp()). Read lại204 giữ mốc đầu; ngoài own/missing cùng404. Không kiểm quyền resource để mark own, cũng không cấp quyền resource. Commit chưa rõ không trả success hoặc tự retry.

DTO không trả recipient/contact/patch/reason hoặc credential; message tổng quát do writer hồ sơ hiện có tạo. Mọi response API no-store, không Set-Cookie;204 không JSON. Nhóm Swagger thông báo ở dev chỉ có hai operation đã lắp; nhóm này không công bố năm API request hoặc ba API HR khác từ canonical10operation. Production không mở docs.

Lệnh tái hiện sau checkout nhánh (tại gốc repo):

```sh
go vet ./...
go test ./... -count=1
go build ./...
python3 scripts/authdb/manage.py check
python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./migrations ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m
node --test scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs scripts/check-swagger.test.mjs
node scripts/check-harness.mjs
git diff --check
```

Đợt worktree thực chạy runner gốc đã có state0600, không sao chép/in credential hoặc tạo tài nguyên mới:

```sh
python3 /Users/quanjp/Projects/ai-staffing/scripts/authdb/manage.py run -- sh -c 'cd /private/tmp/ai-staffing-DATN-42 && go test -tags=integration -p 1 ./migrations ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m'
```

Không chạy hai suite DB đồng thời giữa các chat. Runner xác minh cluster18/owner và cấp TEST_DATABASE_URL chỉ cho subprocess; không dùng DATABASE_URL/.env/Neon. Fixture tự tạo/dọn `auth_case_*` thuộc chính lượt test; không dọn DB demo/checkout khác hoặc chạy destroy/reset để sửa test lỗi.

| Phần bằng chứng | Test thực |
|---|---|
| PCA-DB10 own/filter/page rỗng/read | `TestNotificationOwnSnapshotFilterAndRead`: user_id≠employee_id; Admin không bypass; mất HR làm resource404 nhưng own read vẫn204; read lặp giữ mốc; total/unread và page INT32max |
| PCA-DB10 snapshot | `TestNotificationListConcurrentSnapshot`: hai connection, mutation insert hoặc read giữa count/items qua hook; snapshot hiện tại không trộn dữ liệu, lượt sau thấy commit |
| PCA-DB10 cạnh tranh/recheck | `TestNotificationConcurrentReadAndActorRecheck`: `pg_blocking_pids` xác nhận chờ thật; so timestamp sau hai commit với mốc UPDATE đầu; actor limited/locked/employee inactive thay đổi trong lúc chờ bị từ chối |
| Commit outcomes | `TestNotificationCommitFaults`: rollback xác nhận và commit PG thật nhưng wrapper làm mất xác nhận; không success giả/retry, list đối soát DB. Wrapper không phải mô phỏng mất mạng vật lý |
| HTTP+PG | `TestNotificationApplicationPostgres`: submit/approve/reject writer thật→list own; role/limited/revoke/ID/DTO; chunked65536/65537, JSON/charset/encoding/Origin/CSRF/query;204 không body/no-store/không cookie; fault schema500 và khóa bảng gây timeout DB503 thật |
| Writer rollback tái sử dụng | `TestProfileChangeAtomicFaultsAndCommitOutcomes`, `TestProfileChangeDecisionFaultsAndCommitOutcomes` trong suite repo: notification không thoát transaction lỗi; xem giới hạn/bằng chứng DATN-39/40. Không gọi unit/fake là transaction thật |
| Kiểm tĩnh/unit | Notification service validation/offset/projection/error; HTTP mapper/inventory/dev-only/dependency;6 Node Swagger tests gồm suffix `/{id}/read`, không mở operation chưa có |

Kết quả đợt23/11/2026: Go1.27.4 darwin/arm64, Node24.15.0, Python3.9.6. `check` trước/sau integration xác minh PG18/owner/DBgốc0bảng/systemID7705732357505970219. Go vet/unit/build đạt; notification HTTP/service unit bổ sung đạt;48 contract/harness/Swagger tests đạt. Lượt integration đầu đạt migrations/repository/HTTP nhưng server test dùng nhầm assertion không Set-Cookie của nghiệp vụ cho logout; đã sửa dùng auth helper (logout xóa cookie đúng contract). Retest `./cmd/server -run Notification -count=1 -timeout=5m` đạt11.093s, gồm fault503 thật. Suite tổng hợp cuối đạt cả4package: migrations14.298s, repository35.163s, HTTP4.388s, server19.259s; gồm assertion timestamp concurrent đã siết. Harness437file và diff check đạt.

Tự review và sub-agent source/test review đã xử lý regression mock Swagger; đã bổ sung so timestamp concurrent trực tiếp theo góp ý. Review không thay bằng chứng DB. Chưa UI/browser/khung chuông, DATN-41 gộp chung, F4, polling/realtime/email/push, Neon/HTTPS/deploy; không nhận toàn FR-059 hoặc quyền resource từ link notification. Chưa commit/push/PR/Jira Done trong lượt triển khai này.

**Tiếp nối bàn giao24/11/2026:** đã rebase nhánh DATN-42 lên `origin/main` tại54b75ab, chứa PR54/DATN-41 đã merge. Hợp nhất composition/router/Swagger và tài liệu hai task; fallback giữ cờ ProfileChanges, hai nhóm Swagger vẫn có5operation request và2operation notification. Hai notification operation chuyển sang implemented trong canonical/checker, ba HR/F4 vẫn planned. `TestNotificationApplicationPostgres` kiểm thêm cả hai URL Swagger qua composition root thật; không chỉ render/VM.

Kiểm lại bản tích hợp: Go vet/unit/build đạt,49 contract/harness/Swagger tests đạt, integration4package tuần tự đạt: migrations15.609s, repository49.513s, HTTP5.548s, server23.947s. Runner `up` chỉ khởi động lại container test sở hữu đang dừng; `check` xác minh PG18/owner/DBgốc0bảng trước suite. Harness442file và diff/staged checks được kiểm trước commit bàn giao. Sub-agent rà lại tương thích không còn góp ý chặn; giới hạn chưa gộp DATN-41 của snapshot23/11 đã được khép, các giới hạn UI/browser/F4/Neon/HTTPS/deploy vẫn giữ. Commit/push/PR do người dùng giao trong lượt bàn giao này; không tự merge hoặc chuyển Jira Done.
