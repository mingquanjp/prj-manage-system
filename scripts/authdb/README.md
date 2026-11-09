# PostgreSQL test cho auth — DATN-6 / AUTH-T05

Nguồn SQL/mapping/fixture và phạm vi: [proposal mục 6.1.7](../../docs/proposals/backend-auth-session.md#617-bàn-giao-môi-trường-và-kế-hoạch-dữ-liệu--datn-6--auth-t05-08112026). Đây là công cụ local dành riêng cho test, không phải server/runner migration production.

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

Bộ repository kiểm hợp đồng đọc trên PostgreSQL thật. Bằng chứng và ranh giới với service/HTTP ở [proposal mục 6.1.9](../../docs/proposals/backend-auth-session.md#619-truy-vấn-định-danh-và-quyền-hiện-hành--datn-8--auth-t07-08112026).

## Transaction auth integration — DATN-9 / AUTH-T08

```sh
python3 scripts/authdb/manage.py up
python3 scripts/authdb/manage.py run -- go test -race -tags=integration -p 1 ./migrations ./internal/repository/postgres -count=1 -timeout=4m
python3 scripts/authdb/manage.py run -- go test -tags=integration ./internal/repository/postgres -run '^TestAuthRepositoryRootRemainsEmpty$' -count=1
```

Lệnh cuối chạy riêng sau toàn bộ suite để xác nhận DB gốc vẫn trống và không còn DB case, không phụ thuộc thứ tự đăng ký các test trong package. Giữ một lượt integration tại một thời điểm. `authdb.Case.Connect(t)` mở thêm connection từ bản sao config của DB case đã xác minh, kiểm lại cluster/name/owner và đăng ký cleanup. Không dùng `ConnString()` để chuyển DB đích.

Sáu nhóm transaction mới kiểm login/phiên cũ, nền thay đổi, logout, rollback, 13 ca cạnh tranh và commit mất phản hồi. Barrier xác nhận blocker/waiter thực bằng `pg_blocking_pids`; ticker chỉ thăm dò quan hệ khóa có timeout, không dùng sleep để giả định request đã chờ khóa. Clock điều khiển an toàn giữa goroutine, gồm biên hết hạn trong khi chờ user/session. Test gây lỗi bằng trigger chỉ tạo trong database case. Mô phỏng mất phản hồi gọi commit PostgreSQL thật rồi trả lỗi I/O từ wrapper để kiểm adapter không báo thành công giả; chưa thử cắt mạng thực.

Đã chạy các cổng Go cùng bộ integration ở trên trên PostgreSQL 18 test sở hữu, không đổi schema/migration. [Proposal mục 6.1.10](../../docs/proposals/backend-auth-session.md#6110-transaction-tạo-và-thu-hồi-phiên--datn-9--auth-t08-09112026) ghi bằng chứng và giới hạn cho DATN-10/DATN-18. Kiểm tĩnh harness/contract không thay bằng chứng transaction.

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

Chỉ chạy một lượt integration tại một thời điểm. Các ca cạnh tranh dùng hai connection thực thi, với connection quan sát riêng khi cả hai worker đều bận. Clock điều khiển thời điểm test. `pg_blocking_pids` xác nhận PostgreSQL đang chờ khóa trước khi nhả transaction đầu hoặc dịch clock. Trigger gây lỗi nằm trong database case. [Proposal mục 6.1.11](../../docs/proposals/backend-auth-session.md#6111-transaction-đổi-mật-khẩu-lần-đầu--datn-10--auth-t09-09112026) ghi phạm vi bàn giao và ranh giới với kiểm thử HTTP.

## Kiểm tra công cụ

```sh
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s scripts/authdb -p 'test_*.py' -v
node --test scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs
node scripts/check-harness.mjs
```

Unit tests dùng Docker giả, không là bằng chứng PostgreSQL thật. Bằng chứng local thực chạy được ghi riêng tại proposal mục 6.1.7. Không tự chạy `destroy` trong test suite. Không dùng seed ứng dụng hoặc Neon làm DB test.
