# Swagger UI dev cho auth và danh mục tổ chức

Cách chạy hằng ngày, kịch bản FE và ma trận kiểm chứng: [runbook auth FE–backend](auth-runbook.md). Các harness dưới đây chỉ phục vụ kiểm riêng, không thay `pnpm dev`.

Swagger đọc [contract auth](auth.openapi.json) và phần danh mục/hồ sơ đã triển khai từ [contract OAP](organization-accounts-profile.openapi.json), được nhúng bằng [embed.go](embed.go) và phục vụ bởi [HTTP transport](../../internal/transport/http/swagger.go). Chọn nhóm Xác thực/Tổ chức và hồ sơ trên trang; bản hiển thị OAP lọc 13 operation từ canonical spec, không tạo bản contract viết tay hoặc annotation trong handler. Nguồn hành vi API và bằng chứng backend ở [proposal mục 6.2.13](../proposals/backend-auth-session-implementation.md#6213-tích-hợp-năm-api-với-postgresql--datn-18--auth-t17-12112026).

## Mở tài liệu cùng ứng dụng local

Từ `web/`, chạy `pnpm dev`, rồi mở:

- FE: `http://127.0.0.1:5174/`.
- Swagger trên **chính backend Go**: `http://127.0.0.1:8080/docs/` (`/docs` redirect về `/docs/`).

Không cần chạy Swagger harness hoặc thêm terminal backend. Localdev lắp một router/DB chung cho năm API auth, 12 API danh mục và tài liệu; limiter auth giữ policy riêng. Để FE và Swagger đều thử mutation được, chỉ command localdev nhận đúng hai Origin `http://127.0.0.1:5174` và `http://127.0.0.1:8080`; vẫn kiểm cookie/CSRF, không CORS/Origin wildcard, không nhận localhost thay127.0.0.1. Cookie theo host, không theo port: FE và Swagger local chia sẻ cùng phiên, đăng xuất ở một nơi ảnh hưởng nơi còn lại; login lại phải lấy CSRF mới.

`cmd/server` bình thường vẫn chỉ chấp nhận `APP_ORIGIN` cấu hình. Chỉ APP_ENV=dev mở docs, production404. Chạy backend production không tự bật ngoại lệ local hoặc Swagger. Cổng8080 bị chiếm thì runner dừng, không kill dự án khác. Cách chạy/dừng/dữ liệu giả tại [web README](../../web/README.md#chạy-fe-auth-thật--datn-19--auth-t18).

## Demo trên DB test riêng

Harness thủ công có build tag riêng `swaggerdemo`, dùng production composition root và fixture PostgreSQL thật. Không đọc dotenv, không fallback DATABASE_URL, không gọi migration/seed ứng dụng hoặc dùng DB case làm môi trường lâu dài. Runner xác minh Docker/cluster test trước khi harness tạo database case và nâng schema trong case.

```sh
python3 scripts/authdb/manage.py check
python3 scripts/authdb/manage.py run -- go test -tags=integration,swaggerdemo ./cmd/server -run '^TestSwaggerDemoServer$' -count=1 -v -timeout=25m
```

Mở `http://127.0.0.1:18080/docs/`. Cổng chỉ bind loopback; nếu đã bị dùng thì dừng thay vì chiếm cổng. Hai tài khoản **giả**:

| Email | Mục đích |
|---|---|
| `quan1652005@gmail.com` | Phiên thường với quyền HR |
| `quan1652005+limited@gmail.com` | ADMIN buộc đổi mật khẩu |

Mật khẩu ban đầu cho cả hai: `12345678`, chốt riêng cho demo local để dễ nhập; chỉ lưu bcrypt hash trong database case. Hai email chỉ là định danh trong dữ liệu demo, không gửi thư hoặc kết nối Gmail. Mỗi lượt dựng lại fixture ban đầu; thay đổi không được giữ lâu dài. Khi thử đổi mật khẩu lần đầu vẫn phải nhập mật khẩu mới đáp ứng policy ít nhất 15 ký tự Unicode/tối đa 72 byte UTF-8, ví dụ `QuanDemo2026Moi!`; không dùng lại `12345678` làm mật khẩu mới.

Harness in PID, tự dừng sau 20 phút. Để dừng sớm và cho cleanup đóng server/pool rồi xóa đúng DB case, chạy `kill -USR1 <PID-đã-in>` từ terminal khác. Không kill -9, không chạy song song bộ integration khác trên cluster này. Dòng PASS của harness chỉ có nghĩa đã đóng/dọn harness, **không chứng minh người dùng đã thử đủ các luồng trên browser**.

## Harness FE auth riêng — DATN-19

**Ứng dụng local chạy liên tục:** từ `web/`, dùng `pnpm dev` để mở cả FE+Go, `pnpm dev:demo` để xem prototype. Runner này backend8080/docs chạy cùng ứng dụng, không dùng Go test và không có timer20phút; Ctrl+C dọn DB giả theo lượt chạy.

**Kiểm harness tạm:** FE dùng proxy `/api` tại `http://127.0.0.1:5174`, backend vẫn bind18080. Chạy cùng lệnh runner phía trên nhưng chọn `-run '^TestFrontendDemoServer$'`; harness đặt APP_ORIGIN chính xác5174, không nới Origin/cookie/CSRF. Không dùng harness Swagger18080 hoặc chạy cả hai cùng lúc. Nếu dùng harness này với FE, dừng localdev rồi chạy Vite riêng với `TASKASSIGN_API_TARGET=http://127.0.0.1:18080 VITE_AUTH_MODE=server`; mặc định proxy ứng dụng hiện là8080. Hướng dẫn FE, mode và cách tự kiểm tại [web README](../../web/README.md#chạy-fe-auth-thật--datn-19--auth-t18). Form đổi mật khẩu FE chưa thuộc DATN-19.

## Thử request

1. Mở POST login → Try it out → nhập email/password test → Execute. Trình duyệt tự nhận cookie HttpOnly từ response; không nhập cookie vào Authorize hoặc copy session secret.
2. GET me → Execute: 200 với principal hiện hành. GET csrf → Execute: 200 và `csrf_token`.
3. Authorize → mục `csrfHeader` → nhập token CSRF → Authorize → Close. Bỏ trống mục `cookieAuth`; cookie do trình duyệt tự gửi. Không thêm tiền tố Bearer.
4. POST logout → Execute: 204. GET me tiếp theo: 401.
5. Với email limited: login trả cờ true và role hiệu lực rỗng. Lấy CSRF mới rồi POST change-password bằng current/new theo policy. Thành công 204, cookie bị xóa. Mật khẩu cũ không login được; đăng nhập lại bằng mật khẩu mới để có phiên thường.

Origin/Cookie do trình duyệt quản lý. Để Swagger không chặn form vì ô Origin required bị trống, bootstrap điền origin hiện tại làm default trên object hiển thị; endpoint OpenAPI canonical vẫn nguyên nội dung nguồn. Đây không phải quyền sửa Origin gửi đi. Bootstrap bỏ hai header này khỏi request do Swagger xây dựng, không giả lập nguồn và không sửa middleware. CSRF phải khớp phiên mới nhất; sau mỗi login lấy token mới và thay giá trị Authorize cũ. Authorization chỉ trong bộ nhớ trang; `persistAuthorization=false`. Response/body có thể còn hiển thị khi trang mở: đóng/reload trang trước chia sẻ máy/ảnh, không chia sẻ credential/token thật. Thử request làm thay đổi dữ liệu thật trong DB đích.

UI tiếng Anh là giao diện gốc của Swagger; hướng dẫn và mô tả nghiệp vụ bằng tiếng Việt. UI không có tài khoản tự cấp, không tự mở quyền nghiệp vụ, không tự fetch spec ngoài. `requestInterceptor` chỉ cho contract local và các operation auth/danh mục đã công bố cùng origin, gồm PATCH/DELETE item hợp lệ; `validatorUrl=null`, `queryConfigEnabled=false`, không Petstore/validator/CDN khi chạy.

## Tài nguyên và CSP

Vendored `swagger-ui-dist` **5.34.1**, Apache-2.0, chỉ bundle JS/CSS cần thiết. [Manifest](../../internal/transport/http/swagger/vendor.json) ghi URL tarball chính thức, integrity SHA-512 của archive và SHA-256 từng file; LICENSE/NOTICE được giữ cạnh assets, không mở directory listing. Archive đã kiểm khớp metadata npm trước khi trích file. Checksum không thay kiểm tra lỗ hổng hoặc chữ ký publisher; chưa nhận một lượt scan vulnerability dependency này là đã chạy.

Không sửa bundle vendor; cập nhật bằng gói version cụ thể từ npm chính thức, xác minh integrity, thay đúng assets + license/notice + manifest, rồi kiểm lại CSP và browser. Không dùng `latest` ở runtime/build hoặc thêm Node/Go dependency cho server. Nhúng spec/assets vào binary tăng kích thước binary, nhưng production không có route docs.

CSP chỉ áp cho docs: `default-src 'none'`, script/style từ self, connect từ self, ảnh self/data, chặn base/form/frame. `style-src-attr 'unsafe-inline'` dành cho style attribute do Swagger/React render; không mở inline script hoặc eval. Header nosniff, frame DENY, Referrer-Policy same-origin và Cache-Control no-store được đặt cho docs. Không thay CSP/headers của frontend/API ngoài đợt này.

## Kiểm tra

```sh
go vet ./...
go test ./... -count=1
go build ./...
node --test scripts/check-swagger.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs
node scripts/check-harness.mjs
```

Go httptest kiểm dev/prod, asset/method allowlist, MIME, security headers, HEAD và spec phục vụ khớp byte nguồn. Node kiểm checksum vendor và bootstrap request policy bằng VM; không phải rendering/browser test. Integration HTTP/DB hồi quy dùng lệnh [DATN-18](../../scripts/authdb/README.md#tích-hợp-http-và-postgresql--datn-18--auth-t17). Bằng chứng browser thực và giới hạn mỗi lượt phải ghi riêng, không suy từ static tests.

Lượt triển khai 12/11/2026 đã đạt Go vet/test/build; vet build tag integration/swaggerdemo; 28 kiểm thử Node Swagger/contract/harness; kiểm kê harness và git diff check. Integration HTTP transport + cmd/server với `-race -tags=integration -p 1` đạt trên PostgreSQL test sở hữu. Harness demo đã mở loopback và HEAD `/docs/` trả 200 cùng CSP/security headers; chưa tính là browser proof. Công cụ browser hiện báo local browser use bị tắt trong Work settings nên rendering/CSP, cookie tự gửi và CSRF qua Swagger trên trình duyệt vẫn chờ kiểm; mất mạng thật, proxy FE và scan dependency cũng chưa thực hiện. Review read-only của agent hỗ trợ không phát hiện lỗi actionable sau khi sửa allowlist embed và default Origin.

Tham khảo chính thức: [phân phối Swagger UI](https://swagger.io/docs/open-source-tools/swagger-ui/usage/installation/), [cấu hình](https://swagger.io/docs/open-source-tools/swagger-ui/usage/configuration/), [giới hạn header cookie của browser](https://swagger.io/docs/open-source-tools/swagger-ui/usage/limitations/).


## DATN-26 — 12 API danh mục qua HTTP/PostgreSQL

[OpenAPI OAP](organization-accounts-profile.openapi.json) giữ thiết kế 16 operation.
Runtime hiện nối 12 operation của departments/positions/job-levels: GET/POST
collection, PATCH/DELETE item. Bốn operation tài khoản/hồ sơ chưa triển khai,
không công bố trong Swagger dev. `/docs/openapi.json` giữ contract auth canonical;
`/docs/catalog.openapi.json` là phần paths/components tới được của danh mục,
được lọc từ canonical OAP. Cả hai chỉ phục vụ trong dev.

Cách thử qua backend local `http://127.0.0.1:8080/docs/`:

1. Chạy `pnpm dev` trong `web/` theo runbook hiện hành, dùng DB local test sở hữu.
2. Ở nhóm Xác thực, login bằng tài khoản HR test local và lấy CSRF token.
3. Chọn Danh mục tổ chức, Authorize lại bằng token hiện hành. Cookie vẫn được
   trình duyệt giữ; không nhập Cookie/Origin thủ công. Chuyển nhóm không lưu
   authorization vào storage; token không được giữ thay khi UI được dựng lại.
4. GET collection → POST `{"name":"Phòng ban thử nghiệm"}` → ghi lại id/name.
   POST trả201 và Location; không có GET item riêng trong đợt này.
5. PATCH item với `name` mới và `expected_before_name` đúng tên vừa đọc.
   Dùng tên nền cũ sẽ nhận409 STALE_DATA; trùng tên trong cùng danh mục nhận409
   DUPLICATE_NAME. Rename no-op trả200 nhưng không sinh audit/timestamp giả.
6. DELETE item với JSON `expected_before_name` hiện hành →204 body rỗng.
   Mục còn employee tham chiếu, kể cả inactive, nhận409 CATALOG_IN_USE.
7. Admin-only được GET, không được mutation; HR+Admin cộng quyền từ server.
   Phiên limited bị403 PASSWORD_CHANGE_REQUIRED trên mọi API danh mục.

GET nhận page/page_size/q theo contract, không query lạ/trùng; q tìm literal
không phân biệt hoa/thường, `%`/`_` không là wildcard. Mutation yêu cầu JSON
strict, Origin/CSRF và body thực tối đa8192byte, kể cả request chunked.
Mọi response catalog no-store, không cấp/gia hạn cookie, không CORS wildcard.
Actor lấy từ principal middleware; repository kiểm lại ACTIVE/cờ/role dưới khóa.

Kiểm tự động:

```sh
go vet ./...
go test ./... -count=1
go build ./...
node --test scripts/check-swagger.test.mjs scripts/check-organization-contract.test.mjs scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs
python3 scripts/authdb/manage.py check
python3 scripts/authdb/manage.py run -- go test -race -tags=integration -p 1 ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m
python3 scripts/authdb/manage.py run -- go test -tags=integration ./internal/repository/postgres -run '^TestAuthRepositoryRootRemainsEmpty$' -count=1
```

Chạy mỗi lượt DB tuần tự qua runner đã xác minh PG18/owner. Bằng chứng, phạm vi
fault injection và kết quả thực chạy ở [runbook PostgreSQL](../../scripts/authdb/README.md#datn-26--http-danh-mục-và-swagger-dev).
Swagger httptest/VM không thay browser rendering, cookie hoặc CSP qua browser.
UI danh mục vẫn prototype, chưa nối API; không dùng ticket này nhận OAP-T05 đạt.
Không thêm migration, biến môi trường, dependency hoặc truy cập Neon.


## DATN-28 — API hồ sơ chính chủ

GET `/api/employees/{id}` đã nối qua composition root thật và được thêm vào nhóm
**Tổ chức và hồ sơ** tại backend `/docs/`. URL spec vẫn là
`/docs/catalog.openapi.json` để giữ cơ chế local hiện hữu; khi application có đủ
dependency, spec gồm 12 operation danh mục và 1 operation hồ sơ, lọc từ canonical
OAP cùng components được tham chiếu. Không công bố ba operation tài khoản chưa nối.
Ghi nhận DATN-26 ở trên mô tả phạm vi tại thời điểm bàn giao cũ.

Cách thử: login phiên thường trong nhóm Xác thực, lấy `employee_id` từ response
login hoặc GET `/api/auth/me`, rồi gọi GET hồ sơ với ID đó. Không dùng `user_id`
thay `employee_id`; GET này không cần CSRF và không tạo/gia hạn cookie. Người không
có role đặc biệt vẫn đọc được chính chủ. HR/Admin không có ngoại lệ để đọc ID khác:
ID người khác và ID không tồn tại cùng404. Phiên hạn chế trả403, phiên không hợp lệ
trả401; lỗi DB giữ500/503 theo nguồn.

Response chỉ có 11 trường theo SelfProfile. Phone/address/department/job_level
chưa có giữ JSON null; position bắt buộc. Company email chuẩn hóa khi đọc bằng
helper auth, không UPDATE email legacy. Không có credential, role, CV, lịch sử
đề nghị hoặc audit. Tất cả response no-store; không log PII/body lỗi DB.

Kiểm route/DTO/canonical subset bằng Go httptest; kiểm interceptor Swagger bằng
`scripts/check-swagger.test.mjs`. Các kiểm này không thay browser rendering hoặc
HTTPS production. HTTP + PostgreSQL test và bàn giao OAP-T07 xem
[runbook PostgreSQL](../../scripts/authdb/README.md#datn-28--api-hồ-sơ-chính-chủ).
Chưa nối UI hồ sơ; không tự mở thêm quyền đọc toàn bộ hồ sơ hoặc sửa hồ sơ.
