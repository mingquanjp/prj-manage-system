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
python3 scripts/authdb/manage.py run -- go test ./internal/repository/postgres -count=1
```

Runner kiểm đích trước, cấp `TEST_DATABASE_URL` từ state sở hữu vào subprocess, loại `DATABASE_URL`, `APP_ENV` và biến `PG*` từ shell. Nếu shell đã có `TEST_DATABASE_URL` khác state thì từ chối, không ghi đè âm thầm. Lệnh test phải yêu cầu biến test này, tự fail khi thiếu và không gọi loader dotenv. Không xuất URL/credential trong assertion, stdout hoặc lỗi driver.

`run` là wrapper chạy lệnh được người phát triển tin cậy, **không phải sandbox**: nó không ngăn một chương trình tùy ý tự đọc `.env` hoặc nối DB khác. Không đưa lệnh không review vào wrapper. Hai guard sở hữu/URL là điều kiện trước thao tác; không bảo vệ khỏi Docker admin thay container đồng thời.

DATN-7 phải thêm integration runner dùng `TEST_DATABASE_URL` đã xác minh trực tiếp với pgx/migrate driver, không gọi `cmd/migrate up` vì lệnh đó đọc `DATABASE_URL` và có thể nạp dotenv dev. Không đổi/bỏ RequireDev để chạy test. Không có lệnh migration trong DATN-6; lệnh `go test` trên hiện chỉ chạy test đã tồn tại, không chứng minh migration auth.

## Kiểm tra công cụ

```sh
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s scripts/authdb -p 'test_*.py' -v
node --test scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs
node scripts/check-harness.mjs
```

Unit tests dùng Docker giả, không là bằng chứng PostgreSQL thật. Bằng chứng local thực chạy được ghi riêng tại proposal mục 6.1.7. Không tự chạy `destroy` trong test suite. Không dùng seed ứng dụng hoặc Neon làm DB test.
