# DATN-32 — Gmail OAuth và kiểm gửi thử

Áp dụng [proposal mục 6.5](../../docs/proposals/backend-organization-accounts-profile.md#65-gửi-email--gmail-api--oauth-20) và [MailResult](../../docs/api/organization-accounts-profile.openapi.json).
[DATN-32 trên Jira](https://datn-mingquanjp.atlassian.net/browse/DATN-32).

## Chuẩn bị trên Google Cloud

1. Trong project Google Cloud do chủ hộp thư quản lý, bật Gmail API. Cấu hình Google Auth Platform với đối tượng **External**, trạng thái **Testing**.
2. Thêm đúng hộp thư gửi đã thống nhất làm **test user**. Scope duy nhất là `https://www.googleapis.com/auth/gmail.send`. Hộp thư nhận không cần cấp quyền OAuth.
3. Tạo OAuth client loại **Desktop app**, tải file JSON vào thư mục riêng ngoài repo và đặt quyền `0600`. CLI dùng loopback IPv4 với cổng được hệ điều hành cấp. Callback có `state` và PKCE S256. Trình duyệt và CLI chạy trên cùng máy.
4. Đặt file client tại `~/.config/taskassign/google-desktop-client.json`. Thư mục local dùng quyền `0700`. Dùng lệnh sau từ thư mục gốc của repo, thay địa chỉ ví dụ bằng hộp thư gửi đã thống nhất.

```sh
mkdir -p -m 700 "$HOME/.config/taskassign"
chmod 700 "$HOME/.config/taskassign"
chmod 600 "$HOME/.config/taskassign/google-desktop-client.json"
export GMAIL_OAUTH_CREDENTIALS_FILE="$HOME/.config/taskassign/gmail-oauth.json"
go run ./cmd/gmail-setup authorize -client "$HOME/.config/taskassign/google-desktop-client.json" -from 'sender@example.invalid'
```

Mở URL trong terminal trên trình duyệt của máy đang chạy CLI. Kiểm tra tài khoản đang chọn trên màn hình Google trước khi đồng ý quyền gửi. Đóng tab callback sau khi CLI báo kết quả. Giữ URL cấp quyền, URL callback và file JSON trên máy local. Mỗi lượt authorize có tối đa 5 phút chờ thao tác và 10 giây đổi code. CLI kiểm scope Google trả về và yêu cầu có refresh token trước khi thay file runtime.

Nguồn Google: [Desktop OAuth](https://developers.google.com/identity/protocols/oauth2/native-app), [Gmail scopes](https://developers.google.com/workspace/gmail/api/auth/scopes), [hạn refresh token](https://developers.google.com/identity/protocols/oauth2#expiration).

## Kiểm trước demo

```sh
go run ./cmd/gmail-setup check
```

Lệnh này tạo sender mới và thử refresh qua Google. Với External/Testing và scope Gmail, refresh token có hạn 7 ngày từ Google cấp. Nếu CLI báo cần cấp lại quyền, chạy lại authorize bằng hộp thư gửi rồi chạy check. Việc này chỉ phục hồi quyền gửi.

Khi chủ hộp thư đã cho phép lượt gửi thật và xác nhận địa chỉ nhận, thay địa chỉ ví dụ sau bằng mailbox nhận đã thống nhất.

```sh
go run ./cmd/gmail-setup send-check -to 'recipient@example.invalid'
```

Thư chẩn đoán có nội dung cố định và không chứa mật khẩu. Khi nhận `accepted`, người nhận kiểm inbox hoặc Spam và xác nhận có thư từ đúng hộp gửi. Ghi thời điểm, trạng thái API và kết quả kiểm hộp thư. Giữ email cá nhân, token, URL callback, message ID và nội dung thư ngoài bằng chứng. Khi nhận `unknown`, đối chiếu hộp thư trước mọi quyết định tiếp theo.

## Bàn giao cho OAP-T11

- `gmail.NewFromEnv()` đọc file từ `GMAIL_OAUTH_CREDENTIALS_FILE`. Khi deploy, mount file riêng chỉ đọc mà user tiến trình đọc được, ví dụ mode `0400`. Loader yêu cầu file thường với quyền group/other bằng 0. Access token lưu trong bộ nhớ của sender để phục vụ các lần gọi kế tiếp.
- Composition root của luồng tạo tài khoản sẽ tạo sender một lần. Sau commit được xác nhận, OAP-T11 gọi `MailSender.SendAccount(ctx, email, result.PasswordForDelivery())`. Mật khẩu tạm được mở đúng tại adapter để tạo mail tiếng Việt.
- Tổng thời gian tối đa 10 giây trên mỗi lần gửi. Trong đó lấy token tối đa 3 giây, gọi Gmail tối đa 7 giây và tôn trọng deadline ngắn hơn của caller.
- `accepted` cần phản hồi thành công kèm ID Gmail. `failed` áp dụng cho lỗi chuẩn bị/refresh hoặc Gmail từ chối rõ ràng. `unknown` áp dụng khi đã gọi gửi mà phản hồi không đủ kết luận, gồm lỗi mạng, timeout, lỗi máy chủ hoặc phản hồi thành công thiếu ID. Mỗi lần gọi gửi có tối đa một POST. OAP-T11 giữ riêng trạng thái mail và trạng thái tạo tài khoản.
- CLI kiểm Gmail ở mức adapter. Nghiệm thu gửi thư mật khẩu tạm xuyên luồng tạo tài khoản thuộc bước tích hợp tiếp theo.

## Kiểm thử trong repo

```sh
go vet ./...
go test ./... -count=1
go build ./...
go test -race ./internal/notify/gmail -count=1
node scripts/check-harness.mjs
git diff --check
```

HTTP giả kiểm refresh và hết hạn access token, `invalid_grant`, timeout, mapping gửi và không phát lại request. Test MIME gọi luồng mật khẩu tạm bằng dữ liệu tổng hợp.

### Kết quả chạy — 17/11/2026

- Go 1.26.5 trên macOS ARM64. `go vet ./...`, `go test ./... -count=1`, `go build ./...` và `go test -race ./internal/notify/gmail -count=1` đều đạt.
- Test callback xác minh `state` và không echo authorization code. Test file xác minh mode `0600`, từ chối symlink và quyền đọc quá rộng. HTTP giả xác minh MIME, token reuse/refresh, từ chối quyền, phản hồi không rõ, timeout ở từng bước và số lần gọi gửi.
- Harness kiểm kê nguồn và `git diff --check` đạt.

- Kiểm Gmail thật với OAuth Desktop ở chế độ External / Testing. Chủ hộp thư hoàn tất `authorize`, chạy `check` và xác nhận refresh thành công. Một lượt `send-check` trả `accepted`. Chủ hộp thư nhận đối chiếu đúng người gửi và thư chẩn đoán trong Hộp thư đến. Bằng chứng này xác nhận đường gửi thư chẩn đoán qua adapter. Kiểm luồng thư cấp tài khoản sau commit thuộc OAP-T11.

### DATN-35 — kiểm cấp tài khoản từ UI, 18–19/11/2026

- CLI `check` refresh thành công trước mở localdev. Một lần tạo account qua UI Admin dùng composition root/PostgreSQL local thật và Gmail adapter thật; UI nhận account đã tạo/mã chính thức và mail accepted. Không chạy thêm send-check hoặc gửi lại.
- Chủ hộp thư xác nhận dùng email/mật khẩu được cấp để login; browser thấy đúng EMP000001 ở cổng đổi sau reload/deep link. Người dùng tự đổi/login mới sau sửa lỗi FE focus; DB cờ buộc đổi false, browser hồ sơ/reload/logout/Back đạt. Không đọc/chụp thư chứa mật khẩu hoặc đưa email cá nhân/message ID vào evidence. [Runbook OAP](../../docs/api/auth-runbook.md#7-datn-35--oap-t13--demo-xuyên-ui-db-và-gmail) giữ kết quả/giới hạn từng bước; không nhận accepted tự chứng minh delivered.
- Giữ đúng lượt localdev xuyên cấp tài khoản/login: DB local bị dọn khi dừng nên thư cũ không chứng minh account còn tồn tại sau restart. Kết nối provider được cấp lại không đồng nghĩa được gửi lại thư cũ.
