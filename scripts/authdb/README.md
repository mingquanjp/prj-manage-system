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
