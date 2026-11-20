# Runbook auth và demo tổ chức, tài khoản, hồ sơ

Điểm vào để chạy, demo và đối chiếu bằng chứng auth/OAP. Thiết kế tại [proposal auth](../proposals/backend-auth-session.md) và [proposal OAP](../proposals/backend-organization-accounts-profile.md); HTTP tại [OpenAPI auth](auth.openapi.json) và [OpenAPI OAP](organization-accounts-profile.openapi.json). Runbook không thay contract. Phần 4/6 giữ bằng chứng lịch sử DATN-21; phần 7 tiếp nối DATN-35. Chưa có reset, gửi lại hoặc đổi mật khẩu tự nguyện; hồ sơ mới nối phần đọc chính thức.

## 1. Chạy hằng ngày

Cần Docker Desktop đang chạy, Python3, Go theo go.mod, Node 24 và pnpm 11; phiên bản FE tại [web README](../../web/README.md#chạy-local). Cài dependency bằng `pnpm install --frozen-lockfile` tại web nếu chưa có node_modules.

Từ root repo:

```sh
cd web
export GMAIL_OAUTH_CREDENTIALS_FILE="$HOME/.config/taskassign/gmail-oauth.json"
go run ../cmd/gmail-setup check
pnpm dev
```

| Địa chỉ | Mục đích |
|---|---|
| http://127.0.0.1:5174/ | FE mode server |
| http://127.0.0.1:8080/docs/ | Swagger trên cùng backend |
| http://127.0.0.1:8080/api/healthz | Health công khai |

Không cần mở thêm backend hoặc harness Go test. Không dùng localhost thay 127.0.0.1. FE proxy `/api` tới 8080, giữ Origin; command local chỉ cho đúng Origin 5174 và 8080, không CORS wildcard. Cookie theo host nên FE và Swagger chia sẻ phiên; logout hoặc login lại ở một bên ảnh hưởng bên còn lại, phải lấy CSRF mới sau rotation.

Runner xác minh Docker/cluster/owner PG 18, tạo DB `auth_dev_*` riêng theo lượt chạy, nâng schema head M2 (000001–000005) và nạp ba tài khoản tổng hợp. Không đọc `.env` thật, không lấy DATABASE_URL hoặc dùng Neon. File OAuth private ngoài repo được backend đọc qua biến trên; cách cấp quyền/check tại [runbook Gmail](../../cmd/gmail-setup/README.md). Không in connection string/state/cookie/token vào báo cáo. Cấu hình server thường và loader production xem [bàn giao config](../proposals/backend-auth-session-implementation.md#632-tích-hợp-loaderstartup--datn-4--auth-t03-08112026); local runner không đại diện cấu hình deploy.

| Định danh local giả | Trạng thái ban đầu |
|---|---|
| quan1652005@gmail.com | LOCAL-001, HR, phiên thường |
| quan1652005+limited@gmail.com | LOCAL-002, ADMIN buộc đổi, quyền hiệu lực rỗng |
| quan1652005+admin@gmail.com | LOCAL-003, ADMIN phiên thường |

Mật khẩu ban đầu local: `12345678`. Ba bản ghi trên là fixture, seed không gửi thư và mật khẩu này chỉ thuộc ứng dụng local. Tạo tài khoản qua UI dùng Gmail thật: chỉ nhập mailbox được phép. Mật khẩu mới cần ít nhất 15 code point, tối đa 72 byte UTF-8; không trim/cắt. Sau đổi, mật khẩu mới chỉ thuộc DB lượt chạy, không đưa vào tài liệu/chat. Khởi động runner mới tạo lại fixture ban đầu và mất tài khoản/danh mục vừa tạo; thư cũ còn trong Gmail không phục hồi tài khoản local.

Chỉ xem prototype: dừng server bằng Ctrl+C rồi chạy `pnpm dev:demo` tại web. Hai chế độ cùng 5174, chạy từng chế độ một. Demo không phải auth; lỗi backend không được fallback sang demo. Server: HR về `/organization/departments`, phiên thường khác về `/profile`; Admin có `/accounts`, HR+Admin có cả hai menu. Admin đọc danh mục qua deep link nhưng không mutation. Limited luôn về `/change-password`; URL nghiệp vụ khác báo chưa nối, không mở workspace mẫu.

## 2. Kịch bản demo và kiểm browser

Dùng dữ liệu local, không lưu HAR/payload/cookie/password vào evidence. Ghi ngày, browser, phiên bản code và kết quả quan sát; hướng dẫn dưới đây không tự là test pass.

| Ca | Thao tác | Kết quả cần quan sát |
|---|---|---|
| B01 thường | Login LOCAL-001, reload, mở /projects/kho | Danh tính HR từ server; URL nghiệp vụ thông báo chưa nối, không mở mock |
| B02 logout | Đăng xuất, Back, reload | Về login, không khôi phục phiên đã revoke; FE lấy CSRF trước POST logout |
| B03 hạn chế | Login LOCAL-002 trước đổi; reload và mở URL nghiệp vụ | Vẫn /change-password; ADMIN cũng không vào nghiệp vụ |
| B04 đổi | Nhập current/new/xác nhận, submit | 204 sau commit, xóa cookie/principal/CSRF/form, về login; không tự login |
| B05 sau đổi | Thử mật khẩu cũ, login mật khẩu mới, reload và mở /change-password | Cũ bị từ chối; mới vào đích theo role; reload còn phiên, không mở đổi tự nguyện |
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
node --test scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs
node --test scripts/check-swagger.test.mjs
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

## 6. Điều kiện trước vận hành và bàn giao DATN-21 (lịch sử)

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

## 7. DATN-35 / OAP-T13 — demo xuyên UI, DB và Gmail

Snapshot đầu lượt: `7377dc6` (DATN-34), nhánh `feature/DATN-35-oap-e2e-handoff`. Bằng chứng dưới không xác nhận PR đã merge hoặc tự chuyển Jira Done. Nguồn test DB chi tiết: [runbook PostgreSQL](../../scripts/authdb/README.md); các bước UI và giới hạn DATN-27/29/34 tại [web README](../../web/README.md).

### Chuẩn bị trước báo cáo

1. Mở Docker Desktop, kiểm hai cổng 5174/8080 chưa bị dùng; không dừng dự án khác. Từ root chạy `python3 scripts/authdb/manage.py check`. Nếu tài nguyên test đã có nhưng đang dừng, dùng `up` theo runbook DB rồi check lại.
2. Đặt `GMAIL_OAUTH_CREDENTIALS_FILE` ngoài repo rồi chạy `go run ./cmd/gmail-setup check`. Nếu cần cấp quyền lại, chủ hộp thư tự authorize theo runbook Gmail. Không gửi thư chẩn đoán thêm khi check thành công và không cần đối soát đường gửi.
3. Chạy `pnpm --dir web dev` trong cùng terminal có biến OAuth; giữ terminal này xuyên demo. Chuẩn bị mailbox nhận do mình kiểm soát, chưa tồn tại trong DB lượt chạy. Không dùng địa chỉ `.invalid` với Gmail thật.
4. Chỉ chia sẻ cửa sổ TaskAssign khi demo. Người nhận tự mở thư và nhập credential ngoài phần ghi hình/chụp ảnh. Evidence không chứa email cá nhân, nội dung thư, mật khẩu, token, message ID hoặc HAR.
5. Dữ liệu local theo lượt chạy: không restart giữa cấp tài khoản và login. Sau restart phải tạo lại tài khoản trên DB mới; không replay thư/POST cũ hoặc coi mail cũ là tài khoản còn tồn tại.

### Kịch bản báo cáo bằng UI (8–10 phút)

| Bước | Người thao tác | Thao tác và điều cần giải thích |
|---|---|---|
| 1 | HR local | Login → Danh mục; tạo phòng ban/vị trí/cấp bậc, reload để thấy DB giữ dữ liệu. Đổi tên giữ định danh; danh mục không sinh quyền. |
| 2 | Admin local | Login → Tài khoản; tạo nhân viên với các danh mục vừa tạo, roles trống; giải thích preview chỉ dự kiến, mã sau lưu mới chính thức. |
| 3 | Admin/nhận thư | UI báo tài khoản đã tạo và Google nhận gửi. Người nhận tự kiểm inbox/Spam; accepted không tự chứng minh thư tới inbox. Không chiếu mật khẩu. |
| 4 | Nhân viên mới | Login bằng mật khẩu tạm → cổng đổi. Reload/deep link vẫn không mở nghiệp vụ. Tự nhập current/new/xác nhận, đổi thành công về login. |
| 5 | Nhân viên mới | Tự kiểm mật khẩu cũ bị từ chối, login bằng mật khẩu mới → `/profile`; tên/mã/danh mục thật, phone/address chưa có; lịch sử đề nghị ghi chưa nối. |
| 6 | Nhân viên mới | Reload, logout, Back/reload: phiên cũ không mở hồ sơ. Không có dữ liệu mẫu hoặc principal người trước. |
| 7 (nếu còn giờ) | HR | Xóa danh mục đang được nhân viên tham chiếu bị chặn; hai tab sửa tên cũ nhận stale, tải lại trước khi sửa tiếp. |

Swagger chỉ dùng khi cần giải thích contract; không thay demo UI. Luồng tạo chỉ gửi một lần có chủ ý. Nếu mail failed/unknown hoặc mất response create: giữ tài khoản, dùng danh sách đối soát; không tạo lại/gửi lại để thử vận may. Nếu chưa nhận thư, giới hạn hiện tại chưa có recovery; dừng nhánh login mới và ghi chưa đạt.

### Kết quả thực chạy — 18–19/11/2026

| Cổng/ca | Kết quả | Ranh giới |
|---|---|---|
| Go vet/test/build | Đạt cả ba lệnh sau bổ sung test inventory | Unit không thay DB |
| FE | `pnpm check` đạt 176 tests FE/model sau sửa lỗi focus, 14 harness, lint/format/typecheck/build | Không React E2E tự động |
| Contract/harness/Swagger | 32 contract/harness và5 Swagger tests đạt; harness396 nguồn, diff check đạt | Kiểm tĩnh/VM, không render browser |
| PG18 | Owner/cluster đúng; DB gốc 0 bảng trước/sau suite | Không Neon |
| Integration bốn package | migrations20.494s, postgres49.963s, HTTP4.465s, server10.969s đều đạt | Fault injection/barrier; không cắt mạng vật lý |
| Router/Swagger | `TestApplicationOAPRouteInventoryMatchesCanonicalSwagger` đạt; Swagger local8080 khớp 16 operation OAP + 5 auth canonical | Đối chiếu method/path và nội dung operation; từng HTTP/DB ở suite integration |
| F1 browser | Tạo ba loại, rename phòng ban, reload còn dữ liệu; hai tab stale từ chối ghi đè; xóa phòng ban đang dùng bị chặn | Cùng FE5174 → Go8080 → PG18 sở hữu |
| F2 browser + Gmail | Một tài khoản mới có mã EMP000001, roles trống, cần đổi; UI báo Google nhận gửi; reload giữ row và không giữ mail outcome. Người nhận xác nhận đã dùng mail/mật khẩu được cấp để login | Xác nhận mailbox từ người dùng + browser thấy phiên hạn chế; không chụp/đọc body thư |
| Quyền browser | HR deep link accounts bị từ chối; Admin deep link danh mục chỉ đọc | Backend recheck/race ở suite DB |
| Login hạn chế | Browser thấy đúng EMP000001 ở cổng đổi; reload và deep link /profile, /accounts đều về /change-password | Không tự điền credential mới |
| Form390px | Viewport390, document scrollWidth375; nhãn/policy và form không tràn ngang; override đã reset | Không nhận mọi thiết bị/focus hoặc mọi nhánh validation là đã kiểm |
| Đổi → login mới → hồ sơ | Người dùng tự đổi/login mới; DB xác nhận cờ false. Browser thấy EMP000001 đúng tên/vị trí/phòng ban/cấp bậc vừa tạo; phone/address chưa có, lịch sử ghi chưa nối | Agent không biết password mới; không nhận toàn FR-009 |
| Reload/deep link sau đổi | Reload giữ hồ sơ; /change-password quay về /profile; employee vào accounts/catalog bị từ chối | Không form đổi tự nguyện hoặc mock fallback |
| Logout/Back/reload | Sau logout, Back và reload đều /login, không còn dữ liệu hồ sơ | Thử lại mật khẩu tạm cũ trên browser chưa được người dùng xác nhận riêng; HTTP/DB regression đã kiểm credential cũ mất hiệu lực |

Lệnh mới kiểm inventory: `go test ./internal/transport/http -run '^TestApplicationOAPRouteInventoryMatchesCanonicalSwagger$' -count=1`. Test so object JSON, kiểm đủ16 method/path và nội dung từng operation. Test mới và toàn bộ Go vet/test/build chạy lại đều đạt. Sau suite DB, `TestAuthRepositoryRootRemainsEmpty` qua runner đạt0.535s: DB gốc trống/không sót auth_case; auth_dev browser tách riêng, giữ cùng lượt xuyên cấp tài khoản/đổi/login.

Rà redaction: DTO account chỉ allowlist8 trường, profile11 trường; secret/command/result nội bộ có JSON/format/log guard. `TestSendAccountMIMEAndTokenReuse` kiểm sender không lộ marker khi format/log, `TestSendClassifiesFailureAndNeverReplays` kiểm token/provider lỗi an toàn và tối đa một POST gửi. HTTP+DB kiểm lỗi provider không lọt response, commit trước mail và không tạo session khi cấp tài khoản. Không thu thập runtime mail body/password hoặc nhận đây là audit log production.

**Lỗi phát hiện và xử lý19/11 (nguồn luồng DATN-20/19):** refresh do focus có thể dựng lại form limited; refresh signed-out chồng nhau có thể mất thông báo lỗi/cooldown. Người dùng gặp khó hiểu sau thử đổi/login; READ ONLY metadata cho thấy tại thời điểm lỗi cờ còn true, chưa commit đổi; browser có cooldown429. Không kết luận429 là password mới sai. Đã giữ phase/form khi kiểm nền, busy chặn mutation tới khi me xác nhận; giữ thông báo/Retry-After và bổ sung giải thích countdown. Hai test hồi quy mới đạt, `pnpm check`176 tests đạt. Người dùng thử lại đổi/login mới thành công; DB cờ false và browser hồ sơ/reload/logout/Back được quan sát sau sửa. Không reset limiter, credential, DB hoặc gửi lại thư; không tự tạo Jira Bug. Phần success message sau204 có controller test; agent không quan sát trực tiếp thời điểm thông báo sau đổi vì người dùng tự thao tác credential.

Ảnh local không credential/PII: `/tmp/datn35-catalog-in-use.jpg`, `/tmp/datn35-limited-mobile.jpg`, `/tmp/datn35-profile-header.jpg` (chỉ tiêu đề/tên tổng hợp, không chụp email công tác). Không đưa ảnh/mail body vào Git/Jira. Giữ localdev đang chạy phục vụ người dùng thử/demo, hiện đã logout. Dừng đúng terminal bằng Ctrl+C sẽ dọn DB lượt này; không gọi việc còn giữ auth_dev phục vụ demo là đã cleanup runtime. DB suite/case đã dọn riêng.

### Truy vết FR và ranh giới nghiệm thu

Đây là mức bao phủ trong snapshot kiểm, không thay trạng thái Jira hoặc tự nhận FR Done trước merge/bàn giao. Các dòng đủ bằng chứng là ứng viên Done trong phạm vi đã chốt, chờ bàn giao phù hợp workflow; các dòng Partial còn thiếu chức năng ghi rõ. FR không sửa lại AC để khớp code.

| FR | Mức bao phủ OAP | Bằng chứng/điều còn thiếu |
|---|---|---|
| FR-005.AC1–AC4 | Đủ bằng chứng phạm vi hiện hành: DB/HTTP/UI, gửi Gmail thật và người nhận login/đổi/login mới | Lỗi mail/commit ở test có kiểm soát; không có cấp lại/gửi lại, phải bàn giao giới hạn trước dùng thực tế. |
| FR-006 | Đủ phần cấp mã/preview trên PG test và UI | Race/rollback/sequence từ OAP-DB07/08; không áp dữ liệu Neon. |
| FR-007 | Đủ phần tra cứu tài khoản trên PG test và UI | Projection không phone/address/credential; không bao gồm sửa quyền/khóa. |
| FR-015 | Đủ CRUD danh mục trong phạm vi hiện hành | OAP-DB05/06, HTTP và browser F1; FK dữ liệu thật giữ liên kết. |
| FR-009 | Partial | Chỉ đọc dữ liệu chính thức và BR-001; chưa có lịch sử đề nghị, cập nhật/duyệt. |
| FR-064 | Partial | Chỉ roles ban đầu khi tạo; chưa cấp/thu hồi role tài khoản cũ, trách nhiệm/Admin cuối. |
| FR-003/004 | Partial | Có kiểm phiên/quyền hiện hành và đổi lần đầu; chưa API quản trị thu hồi/reset hoặc đổi tự nguyện. |

### Giới hạn và bước tiếp

- Không kiểm HTTPS/proxy production, cookie Secure thực, Neon, deploy, log hạ tầng hoặc pentest toàn app. Collation/mapping legacy chỉ được kiểm trên PG test; dữ liệu đích cần preflight riêng.
- Mail fault dùng fake provider/port ở test, commit ambiguity dùng fault sau commit thật; không coi đó là mất mạng vật lý hoặc thu hồi OAuth thật. Không cố làm hỏng kết nối Gmail để demo lỗi.
- Mở rộng hồ sơ đề nghị/HR duyệt hoặc miền kỹ năng theo FR/bản đồ miền là ứng viên tiếp theo; chốt proposal/phạm vi riêng. Storage/CV và công thức phân công vẫn cần quyết định tương ứng, không chốt trong DATN-35.
- Dừng bằng Ctrl+C dọn đúng DB lượt chạy. Không destroy container/volume, FORCE/drop wildcard hoặc dùng down như backup. M1/M2 rollback/dirty và khả năng mất dữ liệu theo runbook DB; không thực thi rollback trên DB demo đang có tài khoản cần dùng.
