# Swagger UI cho API auth

Swagger đọc [contract hiện hành](auth.openapi.json), được nhúng bằng [embed.go](embed.go) và phục vụ bởi [HTTP transport](../../internal/transport/http/swagger.go). Không sinh bản spec thứ hai hoặc thêm annotation vào handler. Nguồn hành vi API và bằng chứng backend ở [proposal mục 6.2.13](../proposals/backend-auth-session.md#6213-tích-hợp-năm-api-với-postgresql--datn-18--auth-t17-12112026).

## Mở tài liệu

Khi server đã chạy trên DB phát triển được xác nhận và schema qua 000003, mở `/docs/` tại đúng `APP_ORIGIN`. Ví dụ `APP_ORIGIN=http://localhost:8080` thì mở `http://localhost:8080/docs/`. Chỉ `APP_ENV=dev` đăng ký `/docs`, `/docs/`, `/docs/openapi.json` và các assets allowlist; production trả 404. Không cần thêm biến môi trường. API auth vẫn qua bảo vệ hiện có.

Nếu đang dùng APP_ORIGIN của frontend ở cổng 5173 thì request từ docs trên cổng 8080 sẽ không khớp Origin. Chạy cấu hình dev phù hợp cho demo backend riêng; chưa thêm proxy docs/Vite trong đợt này. Không đổi APP_ORIGIN của tiến trình đang dùng chỉ để bỏ kiểm nguồn.

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

## Thử request

1. Mở POST login → Try it out → nhập email/password test → Execute. Trình duyệt tự nhận cookie HttpOnly từ response; không nhập cookie vào Authorize hoặc copy session secret.
2. GET me → Execute: 200 với principal hiện hành. GET csrf → Execute: 200 và `csrf_token`.
3. Authorize → mục `csrfHeader` → nhập token CSRF → Authorize → Close. Bỏ trống mục `cookieAuth`; cookie do trình duyệt tự gửi. Không thêm tiền tố Bearer.
4. POST logout → Execute: 204. GET me tiếp theo: 401.
5. Với email limited: login trả cờ true và role hiệu lực rỗng. Lấy CSRF mới rồi POST change-password bằng current/new theo policy. Thành công 204, cookie bị xóa. Mật khẩu cũ không login được; đăng nhập lại bằng mật khẩu mới để có phiên thường.

Origin/Cookie do trình duyệt quản lý. Để Swagger không chặn form vì ô Origin required bị trống, bootstrap điền origin hiện tại làm default trên object hiển thị; endpoint OpenAPI canonical vẫn nguyên nội dung nguồn. Đây không phải quyền sửa Origin gửi đi. Bootstrap bỏ hai header này khỏi request do Swagger xây dựng, không giả lập nguồn và không sửa middleware. CSRF phải khớp phiên mới nhất; sau mỗi login lấy token mới và thay giá trị Authorize cũ. Authorization chỉ trong bộ nhớ trang; `persistAuthorization=false`. Response/body có thể còn hiển thị khi trang mở: đóng/reload trang trước chia sẻ máy/ảnh, không chia sẻ credential/token thật. Thử request làm thay đổi dữ liệu thật trong DB đích.

UI tiếng Anh là giao diện gốc của Swagger; hướng dẫn và mô tả nghiệp vụ bằng tiếng Việt. UI không có tài khoản tự cấp, không tự mở quyền nghiệp vụ, không tự fetch spec ngoài. `requestInterceptor` chỉ cho contract local và năm API auth cùng origin; `validatorUrl=null`, `queryConfigEnabled=false`, không Petstore/validator/CDN khi chạy.

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
