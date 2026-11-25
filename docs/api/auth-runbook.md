# Runbook auth và demo tổ chức, tài khoản, hồ sơ

Điểm vào để chạy, demo và đối chiếu bằng chứng auth/OAP/PCA. Thiết kế tại [proposal auth](../proposals/backend-auth-session.md), [proposal OAP](../proposals/backend-organization-accounts-profile.md), [proposal PCA](../proposals/backend-profile-change-approval.md); HTTP tại các OpenAPI canonical tương ứng. Runbook không thay contract. Phần4/6 giữ bằng chứng lịch sử DATN-21; phần7 tiếp nối DATN-35; phần8 bàn giao F1–F3 PCA/DATN-46. Chưa có reset, gửi lại hoặc đổi mật khẩu tự nguyện; phạm vi hồ sơ hiện hành xem phần8, không suy từ các ghi nhận lịch sử bên dưới.

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

## 8. DATN-46 / PCA-T10 — demo hồ sơ ngày26/11/2026

### Snapshot và điều kiện chạy

Bản kiểm bắt đầu từ main `2278de8` (đã merge PR56/57/58 của DATN-43/44/45), nhánh `feature/DATN-46-profile-e2e-handoff`. Nguồn: [proposal PCA mục4–9](../proposals/backend-profile-change-approval.md), [OpenAPI](profile-change-approval.openapi.json), [SQL/test plan](../proposals/backend-profile-change-approval-sql-plan.md). Không mở lại URL/quyền/nghiệp vụ; F4/FR-012 và PCA-DB11 không thuộc nghiệm thu này. Thay đổi ngoài phạm vi tại checkout gốc được giữ nguyên.

Môi trường thực chạy25/11: macOS arm64, Go1.27.4, Node24.15.0, pnpm11.19.0, PostgreSQL18 test owner/cluster đã qua guard; browser chính là in-app browser. Không dùng `.env`, Neon hoặc Gmail. Dependency nguyên lockfile, `pnpm install --frozen-lockfile` đạt. Không có biến runtime/dependency/migration mới.

Để tập/demo F1–F3, dùng harness opt-in với DB case hy sinh, frontend **5174 → backend18080**. Không chạy đồng thời integration hoặc `pnpm dev` chiếm5174. Harness tự kết thúc sau40phút; lên báo cáo cần khởi động lại trước buổi demo, không dùng tiến trình từ hôm trước. `pnpm dev` thường là5174→8080 và dùng fixture khác; không trộn tài khoản giữa hai môi trường.

```sh
# Terminal1 tại root checkout có state runner sở hữu
python3 scripts/authdb/manage.py check
python3 scripts/authdb/manage.py run -- go test -tags=integration,profileuidemo -p 1 ./cmd/server -run '^TestProfileUIDemoServer$' -count=1 -v -timeout=45m
```

Chờ `Profile UI test ready: 18080 PID=...`. Nếu code ở worktree chưa merge, chạy runner tại checkout giữ state, thêm `go -C /private/tmp/ai-staffing-DATN-46 test ...` thay `go test ...`; không copy state/URL/password.

```sh
# Terminal2 tại web/ của cùng bản code
VITE_AUTH_MODE=server TASKASSIGN_API_TARGET=http://127.0.0.1:18080 pnpm exec vite --host 127.0.0.1 --port 5174 --strictPort
```

Tài khoản harness: `employee@example.invalid`, `hr@example.invalid`, `limited@example.invalid`; mật khẩu tổng hợp ban đầu `12345678` (chỉ môi trường hy sinh, không credential vận hành). Login tại `http://127.0.0.1:5174/login`. Employee/HR có ID user khác employee; employee bắt đầu phone/address NULL. Limited có ADMIN+HR nhưng vẫn chỉ đổi mật khẩu/phiên. Không cần đổi credential hoặc gửi mail để demo PCA.

### Kịch bản báo cáo trên UI

| Bước | Thao tác | Điểm trình bày/kết quả cần thấy |
|---|---|---|
| 1 — chủ hồ sơ | Login employee, `/profile`, chỉ ra địa chỉ chưa cập nhật; mở Đề nghị cập nhật | Official/công tác lấy từ server; email/cơ cấu chỉ đọc, BR-001 nhắc bổ sung |
| 2 — gửi | Địa chỉ `Địa chỉ demo 26/11 — Hà Nội`, ghi chú tổng hợp; Gửi đề nghị | Chi tiết PENDING có trước/đề nghị/hiện tại; official chưa đổi. Đóng rồi reload vẫn có lịch sử |
| 3 — HR | Logout, login HR, mở chuông và filter chưa đọc, chọn thông báo đề nghị | Mở khung không tự đọc; click tải chi tiết đúng quyền rồi mới read. HR thấy ba cột để đối chiếu |
| 4 — chấp thuận | Chấp thuận → Xác nhận | Kết luận/người xử lý/thời gian được lưu; action xử lý lại bị gỡ, official thay đúng trường |
| 5 — owner | Logout, login employee; mở thông báo kết quả, đóng chi tiết rồi reload | Official địa chỉ mới; lịch sử APPROVED và kết luận đúng. Thông báo chung không chứa nội dung đề nghị |
| 6 — từ chối | Employee gửi thay đổi phone tổng hợp; HR từ chối kèm lý do; owner mở kết quả | Official phone giữ nguyên, lịch sử REJECTED, owner xem được lý do |
| 7 — kiểm quyền | Employee mở URL HR hoặc ID đề nghị HR; logout rồi Back | Không đọc người khác; không mở mock workspace; Back không phục hồi PII |

Phần chính khoảng8–10phút, phần lỗi chọn thêm nếu cô hỏi: mở cùng một PENDING ở hai tab HR, xử lý tab1 rồi tab2; tab2 phải yêu cầu đối soát, không ghi lần hai. Có thể chuẩn bị hai pending cùng trường: duyệt đầu, đề nghị sau stale; pending khác trường vẫn duyệt được. Demo thao tác tuần tự cho dễ theo dõi; race đồng thời hai connection được chứng minh bằng integration, không gọi hai click UI tuần tự là race DB.

### Bằng chứng browser thực chạy25/11

| ID | Kết quả quan sát trên UI/PG thật |
|---|---|
| B01 | Employee gửi request1 address NULL→có; detail PENDING giữ current NULL. HR mở notification, readcount1→0, approve; employee nhận kết quả và official mới/lịch sử APPROVED. Reload/profile GET lấy dữ liệu server |
| B02 | Request2 phone NULL→`demo-reject`; HR reject có lý do; employee mở notification thấy REJECTED/lý do, phone official vẫn NULL; filter chưa đọc đúng1 |
| B03 — lỗi đã sửa | Owner HR giữ `/profile`, tab khác duyệt request3: detail APPROVED nhưng sau đóng lịch sử vẫn PENDING. Sửa liên kết DATN-43/45: GET detail hợp lệ tải lại official/lịch sử; terminal xóa thông báo submit cũ. Retest request5 sau **reload bản code mới**: cùng phiên/hai tab, official `demo-sync-v2`, lịch sử APPROVED tự khớp, không cần reload thủ công. Request4 còn giữ closure cũ do HMR nên không dùng làm bằng chứng pass |
| B04 | Request6/7 cùng phone before: approve6, tab khác còn snapshot6 xử lý nhận already-resolved; approve7 stale, official không bị ghi đè. Request8 address khác trường approve thành công dù phone đã đổi; reload7 rồi reject vẫn được |
| B05 | Giữ form nháp fullname, đổi tab rồi quay lại vẫn còn bản nháp; hủy/bỏ nội dung không submit. GET lịch sử/official bổ sung không thay form bằng nền mới |
| B06 | Request9 dùng fault opt-in `PCA43_LOST`: handler commit201 thật rồi đóng socket. FE báo chưa xác định, khóa submit, không tự replay; đóng và GET lịch sử thấy đúng một request9 PENDING, official phone chưa đổi. Đây là fault có kiểm soát, không mất mạng vật lý |
| B07 | Employee đổi query sang request3 của HR: không tìm thấy self; vào reviewURL bị từ chối. Logout/Back về login. Đổi user bỏ lịch sử/draft cũ. Limited ADMIN+HR deep link/reload vẫn change-password, không chuông/nghiệp vụ; không thay mật khẩu |
| B08 | Harness SIGUSR2 bỏ HR của fixture đã login: click notification request9 từ danh sách cũ bị từ chối/không mở PII; own read thủ công vẫn được, unread11→10; Escape trả focus chuông; reload gỡ menu HR, catalogURL bị chặn |
| B09 | Kiểm màn nhỏ: viewport yêu cầu390×844 nhưng runtime đo thực319px; document scrollWidth319, panel287, không tràn. Escape/đóng trả focus chuông; override đã reset. Bằng chứng390px và phân trang hơn20 dòng của DATN-45 giữ riêng, không gọi là chạy lại ở lượt này |

Harness hoàn tất PASS474.04s (package474.519s):3 requests của employee, tổng request ID1–9 đã thao tác trên UI gồm6 của HR;17 notifications/10 unread trước dọn. Guard sau dọn: PG18 đúng owner/system-ID7705732357505970219, DB gốc0bảng,0 DB case. Dừng Vite và đóng các tab thử nghiệm. Không lưu screenshot credential hoặc body hồ sơ thật. Browser Chrome riêng thử đăng nhập/đọc được nhưng click không tác động, Vite ghi lỗi extension `M_ID`; bỏ lượt đó, không nhận kiểm tương tác Chrome pass. In-app browser hoàn tất các ca trên.

**Review và smoke bản cuối:** sub-agent phát hiện thêm nhánh nút “Tải lại chi tiết” chưa dùng đường đồng bộ; đã gom cùng callback với effect, không đổi contract. Tự review và sub-agent đọc lại xác nhận không còn blocker P1/P2 trong diff; sub-agent chạy riêng22 focused tests/typecheck trước lần gom cuối, không thay browser. Agent chính chạy lại `pnpm check` sau gom:214FE+15harness/lint/format/typecheck/build PASS. Smoke PG/browser cuối: mở self detail1 chưa tồn tại nhận404; tab khác cùng HR gửi request1; nút Tải lại chi tiết lấy được PENDING và lịch sử; tab khác approve, owner giữ profile mở notification kết quả rồi đóng: official và lịch sử APPROVED tự khớp. Logout/Back về login. Harness PASS66.48s (package67.036s),0request employee/1request HR,2notifications/1unread; guard sau dọn0case/DB gốc0bảng,5174/18080 không còn listener. Không gọi retry404→có tài nguyên là ca phục hồi DB outage.

### Cổng và ma trận bằng chứng

| Lệnh thực | Kết quả/giới hạn |
|---|---|
| `go vet ./...`; `go test ./... -count=1`; `go build ./...` | PASS trước kiểm và chạy lại cả ba sau sửa inventory. Unit không thay PG |
| `go vet -tags=integration,profileuidemo ./cmd/server` | PASS harness opt-in |
| `python3 scripts/authdb/manage.py check` | PASS PG18/owner/system-ID/DB gốc0 bảng |
| `python3 scripts/authdb/manage.py run -- go -C /private/tmp/ai-staffing-DATN-46 test -tags=integration -p 1 ./migrations ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m` | Lượt1 FAIL cleanup vì DB case bỏ lại không có connection; không lỗi nghiệp vụ được ghi nhận từ lượt đó. Sau người dùng cho phép dọn đúng case, lượt2 PASS: migrations14.335s, repository36.359s, HTTP5.377s, server19.682s |
| Tại `web/`: `pnpm check` | PASS trước sửa213FE+15harness; sau sửa214FE+15harness, lint/format/typecheck/build đạt. Test resource mới kiểm terminal và bỏ kết quả sau clear; browser mới là bằng chứng khép component cùng phiên |
| `node --test scripts/check-profile-contract.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs scripts/check-swagger.test.mjs` | PASS49 tests; bốn file ticket yêu cầu là42 tests trong bộ này |
| `node scripts/check-harness.mjs`; `git diff --check` | PASS kiểm kê464 nguồn và diff sau bổ sung runbook. Chưa kiểm staged/commit vì chưa được giao |

Chi tiết PCA-DB01–10/12 và auth/OAP regression nối tới test thật tại [ma trận DATN-46 PG](../../scripts/authdb/README.md#datn-46--pca-t10--ma-trận-pg-và-regression). Không cộng số unit tests thành số luồng E2E.

Runtime Swagger đã GET qua harness và đối chiếu nguyên operation canonical/no-store: auth5, OAP16, PCA request5, notifications2. Inventory router ghép kiểm29method/path gồm28API+health trong `TestProfileChangeHTTPCompleteInventoryAndSwagger`; đây là adapter/route test có dependency fake, phần thực thi PG ở suite/composition/browser. Ba F4 không publish; render Swagger không thay chứng minh UI/DB.

### FR coverage và giới hạn bàn giao

| FR/AC | Mức bao phủ F1–F3 | Nguồn bằng chứng/giới hạn |
|---|---|---|
| FR-009.AC1–2 | Đủ bằng chứng phạm vi đọc official/công tác/lịch sử và quyền self, ứng viên Done sau bàn giao/merge | B01–03/B07, HTTP self/request/PG projections; BR-001 NULL và trường chỉ đọc. Không sửa official trực tiếp |
| FR-010.AC1–2 | Đủ bằng chứng phạm vi gửi patch/hủy/xem kết quả, ứng viên Done sau bàn giao/merge | B01/B02/B05/B06, PCA-DB04/05; không thêm rút/bổ sung pending |
| FR-011.AC1–2 | Đủ bằng chứng phạm vi HR duyệt/từ chối/chống stale/terminal, ứng viên Done sau bàn giao/merge | B01/B02/B04, PCA-DB06–09, HTTP quyền Admin-only/Director và lỗi. Không nói tất cả tổ hợp role đã chạy browser |
| FR-059 | Partial toàn FR | Sự kiện hồ sơ/own/filter/count đã kiểm B01/B02/B08 +PCA-DB10; sự kiện phân công/kỹ năng/effort chưa thuộc PCA |
| FR-060 | Đủ hành vi trong tài nguyên hồ sơ; Partial toàn phạm vi tài nguyên ứng dụng | GET quyền trước mở/read, own read sau mất quyền, focus B08/B09; chưa khép các miền khác |
| FR-012/013/064 | Không nhận hoàn tất từ task này | F4/skills/quản trị quyền cũ là đầu ra khác; không dùng số màn để tính toàn app80–85% |

Không diễn tập browser DB vật lý outage hoặc read204 đã commit rồi mất phản hồi; DB timeout503/commit ambiguity được kiểm ở HTTP+PG với fault có kiểm soát, FE lỗi read được kiểm adapter/deferred và bằng chứng DATN-45. Không gắn browser pass cho chúng. Không có HTTPS/Secure-cookie production, Neon/deploy/pentest, Gmail hoặc mọi sự kiện thông báo; không thay credential người dùng. Log quan sát chỉ mã lỗi/request_id chung, không raw SQL/payload; có log hủy GET do điều hướng/HMR, không nhận chúng là bằng chứng outage. Cảnh báo React do đổi dependency array khi HMR đã được loại khỏi lượt retest bằng reload; không gọi là lỗi runtime bản build.

Lỗi B03 được ghi tại đây và bàn giao về consumer DATN-43/45; đã sửa/retest trong diff DATN-46, chưa tạo Bug/Jira tự động. Bước tiếp: Quân đọc diff/thử kịch bản, bàn giao Git khi được yêu cầu; F4 DATN-47/48 là đầu ra riêng rồi mới xem miền tiếp theo theo delivery-plan. Không chốt công thức phân công, tự merge hoặc chuyển Done.

### Dừng, khởi động lại và rollback

`kill -USR1 <PID>` đúng PID harness đã in để đối soát/dọn case; Ctrl+C Vite. Chạy guard và kiểm không còn case trước integration tiếp. Nếu gặp case bỏ lại, đối chiếu owner/connection/tiến trình, xin phép dọn đúng case nếu nguồn không thuộc lượt hiện hành; không DROP wildcard/FORCE hoặc destroy cluster để làm xanh test. Dữ liệu harness mất khi dọn, lần khởi động mới trở về fixture; không hứa lưu demo qua đêm.

Migration000006/preflight/guard/down được kiểm ở PCA-DB01–03; dùng runner trên case hy sinh theo SQL plan, không gọi `cmd/migrate` hoặc down trên DB demo thật. Down từ chối khi có request/key/notification; không phải backup/khôi phục hồ sơ. Notification legacy có dữ liệu chưa có mapping vẫn phải dừng; Neon và kế hoạch backup đích cần quyền/review riêng.

## 9. DATN-48 / PCA-T12 — UI HR hồ sơ hành chính

Nguồn thực thi: main750cae4 đã merge API DATN-47; nhánh `feature/DATN-48-hr-personnel-ui`. Thiết kế vẫn theo proposal PCA/SQL9.4/OpenAPI canonical, không thay FR hoặc mở FR-013. `/personnel/employees` tìm theo mã/họ tên/email và phân trang; `/{employee_id}` tải chi tiết HR riêng. Menu HR cộng quyền theo server; giữ đích login/cổng limited hiện hành. Không có action CV, vô hiệu hóa, quyền/reset hoặc dữ liệu demo trên màn server.

### Chạy và trình bày F4

Dùng DB test sở hữu sau guard; không chạy harness và suite DB cùng lúc. Khi code ở worktree, giữ runner ở checkout có state và dùng `go -C /private/tmp/ai-staffing-DATN-48`; sau merge bỏ `-C`. Không sao chép state hoặc đọc `.env`/Neon.

```sh
python3 scripts/authdb/manage.py check
python3 scripts/authdb/manage.py run -- go -C /private/tmp/ai-staffing-DATN-48 test -tags=integration,profileuidemo -p 1 ./cmd/server -run '^TestProfileUIDemoServer$' -count=1 -v -timeout=45m
```

Tại `web/` worktree:

```sh
VITE_AUTH_MODE=server TASKASSIGN_API_TARGET=http://127.0.0.1:18080 pnpm exec vite --host 127.0.0.1 --port 5174 --strictPort
```

Đợi `Profile UI test ready: 18080 PID=...`. Fixture tổng hợp `employee@example.invalid`, `hr@example.invalid`, `limited@example.invalid`, `admin@example.invalid`; mật khẩu ban đầu theo phần8. Đây là DB case dùng một lần, không Gmail thật. `SIGUSR2` đổi HR của đúng fixture, `SIGUSR1` dừng/đối soát/dọn; Ctrl+C dừng Vite. Không dùng cổng5175 nếu không đổi Origin harness tương ứng.

1. HR → Hồ sơ nhân sự → tìm nhân viên → xem chi tiết. List không chứa phone/address; chi tiết mới có liên hệ theo quyền HR.
2. Chỉnh sửa fullname/phone/address/email/cơ cấu. Code/status chỉ đọc. Bản nháp giữ snapshot lúc mở; không gửi trường không đổi, không gửi actor/role/status/code. Phone trống là NULL; address chưa có giữ nguyên được nhưng không xóa trắng khi sửa. Chọn danh mục thật, gồm các trang sau100 mục.
3. Lưu → thông báo sau200. Reload đối soát official; email thay đổi thực sự thu hồi mọi phiên của target và dùng email mới khi login. HR sửa chính mình cũng phải login lại. Canonical email không đổi không rewrite padding/audit/revoke từ backend.
4. Hai tab cùng hồ sơ: sửa tab1, gửi snapshot cũ tab2 → stale, đóng/tải lại; không ghi đè. Đề nghị cùng trường tạo trước direct edit cũng stale khi approve; khác trường vẫn duyệt được.
5. Mất phản hồi/503 sau PATCH: không nhận success hoặc gửi lại; đóng/tải lại GET đối soát. Kết quả200 đã xác nhận không bị đổi thành thất bại chỉ vì lần GET sau lỗi.401/403 bỏ dữ liệu phiên/quyền cũ, không fallback demo.

Fault chỉ trong harness build tag: fullname `PCA48_DB` gây lỗi PG57014 trong transaction (không là DB vật lý outage); `PCA48_LOST` chạy PATCH qua adapter PG thật,200 rồi đóng socket trước phản hồi. Sau lỗi chỉ dùng GET đối soát. Không dùng marker trên dữ liệu thật. Khi dừng, harness đếm audit hành chính và xác minh actor là HR fixture; không in payload/contact/credential.

Audit được ghi atomically phía server. Không có API đọc audit trong bộ contract này nên không dựng bảng lịch sử giả trên UI. FR-012.AC2 được đối chiếu bằng DB audit, không được mô tả thành màn tra cứu lịch sử đã nối. Kết quả thực chạy và giới hạn được bổ sung sau kiểm; trạng thái Jira/merge không tự đổi.

### Kết quả và giới hạn DATN-48 (25–26/11/2026)

Môi trường: macOS arm64, Go1.27.4, Node24.15.0/pnpm11.19.0, PostgreSQL18 test sở hữu (owner đúng, system-ID7705732357505970219, public_tables0 qua guard). Không dùng `.env`, Neon, Gmail thật hoặc credential người dùng. Lượt offline install thiếu tarball, sandbox registry/Go cache bị chặn đã chạy lại bằng quyền phù hợp; không nhận các lượt lỗi môi trường/typecheck đang sửa là PASS. Giữ dependency/lockfile hiện hành.

| Bằng chứng browser + PG thật ngày25/11 | Kết quả |
|---|---|
| HR login/menu → list/detail | Quan sát4 hồ sơ từ DB; list không phone/address; detail user101/employee1001 hiển thị NULL đúng nghĩa. Đích login HR vẫn danh mục. |
| Sửa fullname/address → reload | Đạt: official mới giữ sau reload; không tạo request/notification. Có thông báo200 sau lưu; phone/department/job_level NULL giữ nguyên. |
| Hai tab cùng trường | Tab1 lưu, tab2 gửi before cũ → lỗi stale, khóa form và yêu cầu GET đối soát; không ghi đè. |
| Actor audit và cleanup | Harness PASS2402.079s;2 audit `UPDATE_ADMINISTRATIVE_PROFILE`, actor HR fixture khớp;0 request/notification. DB case dọn; guard sau dọn đạt26/11. Đây không là bằng chứng browser đọc lịch sử audit. |
| Phần chưa khép ở lượt agent ban đầu | Các ca email/session/quyền/fault/responsive/đề nghị và kiểm form bổ sung chưa hoàn tất khi công cụ browser mất khả dụng. Lượt kiểm thủ công do Quân xác nhận ngày26/11 được ghi riêng bên dưới; không suy PASS từ unit/review. |

Công cụ in-app browser mất khả dụng khi tiếp nối (`Local browser access is unavailable for this chat`), reset vẫn lỗi; native fallback hết phiên/timeout. Chrome do người dùng sử dụng nên tránh thao tác tranh cửa sổ. Lượt harness tự hết40phút và dọn đúng fixture. Không báo chặn do DB; regression PG vẫn đạt và guard sau dọn hợp lệ.

| Cổng thực chạy | Kết quả đúng tầng |
|---|---|
| `go vet ./...`, `go test ./... -count=1`, `go build ./...` tại worktree | PASS; unit/build không thay integration. |
| `python3 scripts/authdb/manage.py check` tại checkout giữ state | PASS trước integration và sau browser cleanup, đúng PG18/owner/system-ID/public_tables0. |
| `python3 scripts/authdb/manage.py run -- go -C /private/tmp/ai-staffing-DATN-48 test -tags=integration -p 1 ./migrations ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m` | PASS tuần tự:124.346s/170.034s/25.425s/63.223s. Regression auth/OAP/PCA, không chạy song song với harness. |
| `go vet -tags=integration,profileuidemo ./cmd/server` | PASS kiểm nguồn harness build tag; không nhận compile/vet là fault browser đã chạy. |
| `pnpm check` tại `web/` | PASS229 FE/model tests (15 focused mới),15 harness tests, lint/format/typecheck/build; không skip. Bao gồm hủy catalog GET theo phiên, không replay và no-op; HTTP fake không thay browser/PG. |
| Bốn file Node contract/harness theo ticket; `node scripts/check-harness.mjs`; `git diff --check` | PASS42 tests;481 nguồn và diff cuối đạt. Kiểm tĩnh/kiểm kê không thay runtime. |

**PCA-DB11:** đã chạy lại `TestAdministrativeProfileRaceActualApprove` và `TestAdministrativeProfileRaceApproveBeforeDirect` trong `./internal/repository/postgres` của suite trên. Hai connection/barrier quan sát lock, **cả direct edit và Submit/Approve PCA adapter thật**, hai thứ tự cùng/khác trường; không writer giả/sleep ngẫu nhiên. Bằng chứng HTTP/PG email/revoke/audit/rollback từ các test DATN-47 đã được chạy lại trong suite, xem [ma trận PG](../../scripts/authdb/README.md#datn-47--pca-t11--api-hr-sửa-trực-tiếp-hồ-sơ); không đổi chúng thành ca browser đã đạt. Race login đồng thời đổi email chưa có ca trực tiếp; HTTPS/Neon/deploy/DB vật lý outage chưa kiểm.

| FR | Phạm vi bàn giao hiện tại |
|---|---|
| FR-012.AC1 | Đủ bằng chứng trong phạm vi bàn giao F4: UI/adapter list/search/pagination/detail và7 trường PATCH; PG/HTTP cùng browser agent và lượt kiểm thủ công do Quân xác nhận26/11. Ứng viên Done sau bàn giao/merge theo workflow, không tự đổi Jira. |
| FR-012.AC2, BR-001 | Audit đúng actor/atomic/no-op và không tạo employee request được kiểm PG; NULL/address có browser agent, phần kiểm thủ công tiếp nối được Quân xác nhận. Đủ phạm vi AC lưu lịch sử phía server; không mở quyền/status/skills hoặc quảng cáo giao diện audit-read chưa có. Ứng viên Done sau merge/bàn giao. |
| FR-013/064, toàn app/notifications | Ngoài phạm vi; không suy hoàn thành từ F4. |

Bước tiếp hiện hành: Quân đã xác nhận các ca kiểm thủ công tiếp nối PASS và giao commit/push/tạo PR ngày26/11. Bàn giao Git theo yêu cầu này; merge và Jira Done vẫn do Quân quyết định theo workflow.

**Review cuối:** agent chính đọc diff và sub-agent review độc lập hai lượt; đã sửa lower Unicode email khác Go (chỉ tối ưu no-op ASCII, Unicode để server canonical), class input theo convention, xác nhận bỏ nháp/focus, thêm AbortSignal cho catalog GET và hiển thị no-op. Focused adapter/resource tests15/15 nằm trong229tests. Không có finding P1/P2 còn mở từ review nguồn; review không tự khép browser. Các ca browser còn thiếu ở lượt agent được tiếp nhận từ xác nhận kiểm thủ công của Quân bên dưới. Không tự merge hoặc chuyển Jira Done.

### Tiếp nhận kiểm thủ công và bàn giao Git — 26/11/2026

Sau khi nhận hướng dẫn chạy đúng worktree/harness PG test và checklist trong chat, Quân báo lượt kiểm cơ bản ổn. Agent yêu cầu xác nhận thêm năm nhóm ca quan trọng; Quân xác nhận: “Mình đã thử các ca trên và đều PASS, harness dọn cũng PASS”, rồi giao cập nhật evidence, commit/push/tạo PR. Bảng dưới là **kết quả browser kiểm thủ công do Quân báo**, không phải agent trực tiếp quan sát lại hoặc test tự động. Không thu thập mật khẩu, cookie, CSRF, HAR, ảnh hay payload; không tự đặt PID, thời lượng, số audit/request cho lượt này.

| Nhóm kiểm thủ công theo hướng dẫn | Kết quả do Quân xác nhận |
|---|---|
| Email đổi thật và phiên cũ | PASS: phiên cũ hết hiệu lực; login bằng email mới được. Hướng dẫn có cả target và HR tự đổi email; không tự login hoặc đổi mật khẩu. |
| Direct edit và đề nghị cũ | PASS: cùng trường stale, không ghi đè official mới; khác trường vẫn approve được. Đây là luồng UI tiếp nối; race đồng thời được chứng minh riêng bằng PG hai adapter thật. |
| `PCA48_DB` | PASS: lỗi PG trong transaction không lưu, không báo success; GET đối soát giữ dữ liệu trước đó. Không gọi là DB vật lý outage. |
| `PCA48_LOST` | PASS: báo chưa rõ kết quả, không replay; GET thấy dữ liệu đã commit. Đây là đóng socket có kiểm soát sau200 ở harness, không suy mọi dạng lỗi mạng đã kiểm. |
| Mất HR, Admin-only, limited | PASS: không mở hồ sơ ngoài quyền; limited vẫn ở cổng đổi mật khẩu. |
| Responsive và focus | PASS theo checklist mobile/focus đã hướng dẫn; không tự gán viewport/thiết bị hoặc ảnh đo từ agent cho lượt thủ công. |
| Dọn harness | Quân xác nhận PASS; không gán thời lượng hoặc số bản ghi cho lượt này. |

Kết hợp kiểm tự động/PG, browser agent ngày25/11 và xác nhận kiểm thủ công26/11: không còn ca browser chặn đã nêu trong câu hỏi xác nhận; phạm vi F4/FR-012 đủ bàn giao review/merge. Không bổ sung bằng chứng chi tiết chưa được cung cấp cho các kiểm tra mở rộng như trên100 lựa chọn qua browser (đã có adapter test), mọi tổ hợp role, race login–email đồng thời, read204 mất phản hồi, DB vật lý outage, HTTPS/Neon/deploy/pentest. Không nhận FR-013/064 hoặc toàn app.
