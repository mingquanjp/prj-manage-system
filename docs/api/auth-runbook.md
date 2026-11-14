# Runbook auth FE–backend — DATN-21 / AUTH-T20

Điểm vào để chạy, demo và đối chiếu bằng chứng auth. Thiết kế tại [proposal](../proposals/backend-auth-session.md), HTTP tại [OpenAPI](auth.openapi.json); không dùng runbook thay contract. Chỉ bao gồm FR-001/FR-002, nền xác định quyền của FR-003 và đổi mật khẩu lần đầu của FR-004. Chưa có API nghiệp vụ, cấp tài khoản/email, reset hoặc đổi mật khẩu tự nguyện.

## 1. Chạy hằng ngày

Cần Docker Desktop đang chạy, Python3, Go theo go.mod, Node 24 và pnpm 11; phiên bản FE tại [web README](../../web/README.md#chạy-local). Cài dependency bằng `pnpm install --frozen-lockfile` tại web nếu chưa có node_modules.

Từ root repo:

```sh
cd web
pnpm dev
```

| Địa chỉ | Mục đích |
|---|---|
| http://127.0.0.1:5174/ | FE mode server |
| http://127.0.0.1:8080/docs/ | Swagger trên cùng backend |
| http://127.0.0.1:8080/api/healthz | Health công khai |

Không cần mở thêm backend hoặc harness Go test. Không dùng localhost thay 127.0.0.1. FE proxy `/api` tới 8080, giữ Origin; command local chỉ cho đúng Origin 5174 và 8080, không CORS wildcard. Cookie theo host nên FE và Swagger chia sẻ phiên; logout hoặc login lại ở một bên ảnh hưởng bên còn lại, phải lấy CSRF mới sau rotation.

Runner xác minh Docker/cluster/owner PG 18, tạo DB `auth_dev_*` riêng theo lượt chạy, nâng 000001–000003 và nạp hai tài khoản tổng hợp. Không đọc `.env` thật, không lấy DATABASE_URL hoặc dùng Neon. Không in connection string/state/cookie/token vào báo cáo. Cấu hình server thường và loader production xem [bàn giao config](../proposals/backend-auth-session-implementation.md#632-tích-hợp-loaderstartup--datn-4--auth-t03-08112026); local runner không đại diện cấu hình deploy.

| Định danh local giả | Trạng thái ban đầu |
|---|---|
| quan1652005@gmail.com | LOCAL-001, HR, phiên thường |
| quan1652005+limited@gmail.com | LOCAL-002, ADMIN buộc đổi, quyền hiệu lực rỗng |

Mật khẩu ban đầu local: `12345678`. Đây là fixture riêng cho demo, không credential/mailbox thật và không gửi thư. Mật khẩu mới cần ít nhất 15 code point, tối đa 72 byte UTF-8; không trim/cắt. Sau đổi, mật khẩu mới chỉ thuộc DB lượt chạy, không đưa vào tài liệu/chat. Khởi động runner mới tạo lại fixture ban đầu.

Chỉ xem prototype: dừng server bằng Ctrl+C rồi chạy `pnpm dev:demo` tại web. Hai chế độ cùng 5174, chạy từng chế độ một. Demo không phải auth; lỗi backend không được fallback sang demo. Phiên server dừng ở màn “Đã đăng nhập” hoặc thông báo nghiệp vụ chưa nối, không ánh xạ principal vào nhân viên fixture.

## 2. Kịch bản demo và kiểm browser

Dùng dữ liệu local, không lưu HAR/payload/cookie/password vào evidence. Ghi ngày, browser, phiên bản code và kết quả quan sát; hướng dẫn dưới đây không tự là test pass.

| Ca | Thao tác | Kết quả cần quan sát |
|---|---|---|
| B01 thường | Login LOCAL-001, reload, mở /projects/kho | Danh tính HR từ server; URL nghiệp vụ thông báo chưa nối, không mở mock |
| B02 logout | Đăng xuất, Back, reload | Về login, không khôi phục phiên đã revoke; FE lấy CSRF trước POST logout |
| B03 hạn chế | Login LOCAL-002 trước đổi; reload và mở URL nghiệp vụ | Vẫn /change-password; ADMIN cũng không vào nghiệp vụ |
| B04 đổi | Nhập current/new/xác nhận, submit | 204 sau commit, xóa cookie/principal/CSRF/form, về login; không tự login |
| B05 sau đổi | Thử mật khẩu cũ, login mật khẩu mới, reload và mở /change-password | Cũ bị từ chối; mới vào màn tiếp nối; reload còn phiên, không mở đổi tự nguyện |
| B06 nhiều phiên | Giữ phiên cũ ở browser khác trước đổi; kiểm lại sau đổi | Phiên cũ mất hiệu lực; user khác giữ phiên (DB suite đối chiếu riêng) |
| B07 responsive | Form đổi ở 390×844, lỗi bắt buộc, focus/bàn phím | Không tràn trang, nhãn/lỗi rõ; không ghi mật khẩu vào screenshot |
| B08 lỗi | Đối chiếu các ca bên dưới | Không giả success, không dùng quyền client/fallback demo |

Khi công cụ browser yêu cầu người dùng tự đổi credential: người dùng tự nhập và submit mật khẩu mới, rồi tự login mới; agent chỉ kiểm màn kết quả. Không thu thập mật khẩu để hoàn tất test.

Lỗi kiểm ở đúng tầng: Origin/CSRF sai →403 không mutation; phiên thiếu/revoked/inactive/hết hạn →401; DB lỗi →503/500, không giả401; limiter →429/Retry-After. Dùng integration và fake clock/barrier cho lỗi/race, không thêm endpoint gây lỗi production hoặc dừng DB đang dùng để tạo lỗi tùy tiện. FE HTTP giả kiểm cách hiển thị và không retry; không chứng minh lỗi mạng thật. Mất phản hồi sau POST có thể đã commit: không tự gửi lại, chỉ đối soát bằng GET hoặc login thủ công theo màn hướng dẫn.

## 3. Chạy kiểm tra tái hiện

Từ root, chạy các lệnh này khi không có lượt DB integration/harness khác. `pnpm dev` dùng DB auth_dev riêng; integration chỉ dùng auth_case, không chung dữ liệu.

```sh
go vet ./...
go test ./... -count=1
go build ./...
python3 scripts/authdb/manage.py check
python3 scripts/authdb/manage.py run -- go test -race -tags=integration -p 1 ./migrations ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m
python3 scripts/authdb/manage.py run -- go test -tags=integration ./internal/repository/postgres -run '^TestAuthRepositoryRootRemainsEmpty$' -count=1
node --test scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs
node scripts/check-harness.mjs
git diff --check
```

Nếu cluster đang dừng, chạy `python3 scripts/authdb/manage.py up` trước check. Tại web chạy `pnpm check`. Khi chuẩn bị commit theo yêu cầu riêng mới kiểm staged diff và harness `--staged`. Không dùng make check như cổng chỉ đọc.

Guard dùng TEST_DATABASE_URL từ state sở hữu, loại DATABASE_URL/PG*; không tự copy URL từ Neon hoặc bỏ guard khi lỗi. Mỗi case tạo/migrate/dọn DB riêng; race dùng connection thực và barrier pg_blocking_pids/fake clock, không sleep đoán thứ tự. [Runbook DB](../../scripts/authdb/README.md) giữ chi tiết fixture, mapping, đường rỗng/legacy, rollback/dirty và cleanup.

## 4. Bằng chứng DATN-21 — 14/11/2026

Mốc đầu lượt: kế thừa commit `a194f74` (DATN-20), nhánh `feature/DATN-21-auth-e2e-runbook`. Dữ liệu browser đang dùng DB local của runner có sẵn, không reset mật khẩu người dùng đã đổi. Trạng thái dưới chỉ ghi kết quả thực chạy; đã bổ sung kiểm Chrome LOCAL-002 sau khi người dùng tự login bằng mật khẩu mới. Review read-only từ sub-agent đã đối chiếu lệnh/nguồn/giới hạn; phát hiện runbook bị ignore đã được sửa bằng allowlist riêng, harness 305 nguồn và diff check đạt.

| Cổng/ca | Kết quả lượt này | Giới hạn |
|---|---|---|
| Go vet/test/build | Đạt cả ba lệnh ở mục 3 | Unit thường không chạy DB tag integration |
| FE pnpm check | Đạt lint/format/typecheck/build, 153 test FE và 14 harness | HTTP giả không chứng minh transaction/browser |
| Contract/harness | 25 test đạt (11 contract + 14 harness); 305 nguồn sau thêm runbook | Kiểm tĩnh không chứng minh ngữ nghĩa runtime |
| PostgreSQL guard | PG 18, owner đúng, DB gốc 0 bảng | Không kiểm Neon |
| DB integration/race | Đạt bốn package migrations, repository/postgres, transport/http, cmd/server với -race/-p1 | Trigger/lỗi commit có kiểm soát; không cắt mạng thật |
| Cleanup sau suite | Đạt TestAuthRepositoryRootRemainsEmpty chạy riêng | DB gốc trống và không còn DB auth_case; auth_dev runtime tách riêng |
| B01/B02 IAB | Login LOCAL-001→reload→URL nghiệp vụ chưa nối→logout→Back→reload đều đúng | Qua FE 5174→Go 8080→PG 18; chưa là kiểm HTTPS |
| B05 mật khẩu cũ | LOCAL-002 dùng mật khẩu cũ bị từ chối | Không chứng minh login mới; người dùng giữ mật khẩu mới riêng |
| B03/B04 lịch sử DATN-20 | Browser hạn chế/reload/deep link; người dùng đổi rồi login mới, agent quan sát màn kết quả | Không tính là chạy lại DATN-21 |
| B06 lịch sử DATN-20 và DB suite | Browser giữ phiên cũ đã về login sau đổi; DB test kiểm mọi phiên cùng user mất hiệu lực và user khác giữ phiên | Không nhận toàn bộ B06 là đã kiểm trên browser |
| B05 Chrome LOCAL-002 sau đổi | Người dùng login mới; agent thấy LOCAL-002/Quản trị viên, reload giữ đúng danh tính; /change-password chuyển về /; logout về login, Back vẫn login | Reload cuối sau logout và console bị extension Chrome chặn; không ghi pass hai bước này |
| B07 form đổi trên màn nhỏ | Chờ kiểm | Không gắn pass |

### Ma trận AUTH-S01–S10

Đây là truy vết bằng chứng, không chứng nhận an toàn toàn hệ thống. Nguồn ca tại [proposal 6.4.1](../proposals/backend-auth-session.md#641-bằng-chứng-bảo-mật-cần-có-khi-triển-khai); chi tiết backend tại [DATN-18](../proposals/backend-auth-session-implementation.md#6213-tích-hợp-năm-api-với-postgresql--datn-18--auth-t17-12112026).

| Mã | Bằng chứng có thể tái hiện | Phần chưa kiểm/ranh giới |
|---|---|---|
| S01 | service/handler login; PostgreSQL lifecycle/errors và projection; B05 cũ bị từ chối | Không hứa thời gian bcrypt giống hệt mọi hash legacy |
| S02 | cookie header tests; HTTP PostgreSQL lifecycle/expiry/revoke; B01/B02 | Secure qua HTTPS/proxy production chưa kiểm |
| S03 | TestLimitedSessionRoutes và PostgreSQL restricted guard; B03 lịch sử | Không chứng nhận API nghiệp vụ chưa có |
| S04 | Origin/HMAC/rotation tests, lifecycle CSRF sai không revoke; logout browser qua proxy | Không gọi đó là pentest mọi nguồn cross-site |
| S05 | TestAuthHTTPPostgresLivePrincipal và projection; ID/role/cờ/status từ DB | Không toàn RBAC nghiệp vụ |
| S06 | initial password change + contention/expiry, rollback; DB hai connection/barrier | B06 browser lịch sử không thay race DB |
| S07 | failure atomicity/unknown commit outcome/unavailable; unit random; FE uncertain outcome | Mất phản hồi ở driver được mô phỏng sau commit thật; chưa cắt mạng thật |
| S08 | policy Unicode/byte, limiter boundaries/concurrency/Retry-After và shared limiter; FE controller | 429 UI chủ yếu HTTP giả, không tuyên bố browser mọi ca |
| S09 | redaction tests marker giả; rà nguồn auth FE không storage/log credential | Không thu thập HAR/secret; chưa audit log hạ tầng production |
| S10 | route Chi/OpenAPI, DTO/header tests, năm route integration và browser nền | Không validator response schema đầy đủ hoặc toàn FE đã nối |

## 5. Dừng, cleanup, rollback và khắc phục

- **Dừng bình thường:** Ctrl+C terminal pnpm dev. Runner dừng đúng child group, đóng HTTP/pool rồi dọn đúng auth_dev vừa tạo; container/volume giữ lại. Không kill theo tên hoặc dọn theo wildcard.
- **Port bận:** dùng `lsof -nP -iTCP:5174 -iTCP:8080 -sTCP:LISTEN` chỉ để xác định; đối chiếu cwd/tiến trình trước xử lý. Dừng terminal DATN cũ bằng Ctrl+C; không kill project khác hoặc dùng cổng tự nhảy.
- **ECONNREFUSED8080:** kiểm runner/health và terminal lỗi. pnpm dev khởi động cả FE/BE; Vite đơn lẻ không tự có backend. Không trỏ credential sang backend dự án khác.
- **Origin403:** dùng đúng 127.0.0.1/scheme/port; không bỏ guard, không giả forwarded header. Server thường cần APP_ORIGIN đã validate, local constructor chỉ hai origin cố định.
- **CSRF403:** sau login/rotation phải lấy token mới; không tái dùng token từ phiên trước hoặc lưu localStorage. FE tự bootstrap qua API.
- **Docker/guard lỗi:** mở Docker Desktop, chạy check. Không sửa state để nhận container/DB không rõ sở hữu, không fallback DATABASE_URL. Không chia sẻ `.local/authdb/state.json`.
- **Crash/kill9/cleanup lỗi:** giữ nguyên tên/metadata đã biết để đối soát owner/cluster; không FORCE DROP/prune hoặc xóa auth_dev/auth_case theo prefix. Các DB còn lại cần xử lý cụ thể sau xác nhận, không tự động nhận làm tài nguyên mới.
- **Rollback ứng dụng:** dừng runner, chọn lại commit/PR đã review theo quy trình Git được người dùng giao, rồi chạy lại; không reset/stash thay đổi người dùng. DB local theo lượt chạy không là backup.
- **Rollback migration:** chỉ diễn tập trên DB case hy sinh bằng bộ migration tests. Down 000003 làm mất role/phiên; up không khôi phục dữ liệu đó. Dirty migration phải điều tra, không force version cho xanh. Không chạy down/reset Neon từ runbook này.
- **Dọn toàn môi trường test:** không cần cho chạy hằng ngày. Dừng app/harness trước; chỉ khi được giao xóa dữ liệu mới dùng lệnh destroy có confirm tên ở runbook DB. Không dùng destroy để chữa lỗi UI.

## 6. Điều kiện trước vận hành và bàn giao

Checklist đang mở của ticket:

- [x] Cổng Go, FE, contract/harness; DB integration/race và cleanup.
- [x] Browser luồng thường qua proxy, logout/Back/reload.
- [x] Runbook, truy vết bằng chứng và review từ sub-agent.
- [x] Browser LOCAL-002 sau login bằng mật khẩu mới: reload, /change-password, logout/Back.
- [ ] Reload cuối sau logout LOCAL-002: bước tự động bị UI extension Chrome chặn, chưa có kết quả quan sát.
- [ ] Form đổi mật khẩu trên viewport nhỏ; nếu cần tạo lại fixture phải chốt dừng runner trước, không reset DB đang dùng.
- [ ] Rà console browser trong lượt sau đổi: UI extension Chrome chặn trước khi đọc log; không coi rà mã nguồn/redaction tests là đã quan sát mọi log runtime.

Chưa kiểm HTTPS thật, reverse proxy deploy, mất mạng HTTP/DB thật hoặc log hạ tầng. Cần cấu hình cùng-origin production, Secure cookie và CSP FE rồi kiểm riêng trước vận hành. Swagger chỉ dev, production 404 đã có test; không bật docs công khai để thay kiểm thử bảo mật.

Không tự commit/push/deploy/chuyển Jira Done. DATN-21 chỉ sẵn sàng nghiệm thu sau khi kết quả browser bắt buộc và suite/cleanup được ghi đầy đủ, lỗi chặn đã xử lý. SMTP/quên mật khẩu/API hồ sơ/dự án, áp migration Neon và toàn FR-003/004 nằm ngoài bàn giao này.
