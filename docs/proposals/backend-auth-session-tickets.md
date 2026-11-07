# Ticket Jira — Đăng nhập và phiên truy cập

Cập nhật **07/11/2026** · Bản nội dung để duyệt, **chưa tạo issue Jira**.

## 1. Cách chia và nguồn chuẩn

Bản này thay bản 33 ticket một-một với WBS: **1 Epic và 20 Task theo đầu ra**. WBS giữ 33 đơn vị công việc để truy vết thiết kế/estimate; không tạo song song ticket Pnn.m và AUTH-Tnn. Các mã AUTH-T01–AUTH-T20 là **mã ticket nội bộ**, không phải Jira key hoặc mã quyết định A01–A09 trong proposal.

Nguồn nghiệp vụ/thiết kế: [proposal](backend-auth-session.md), [FR/AC](../product/functional-requirements.md), [OpenAPI](../api/auth.openapi.json). Description diễn đạt phần việc và tình huống cụ thể, không thay nguồn; khi nguồn đổi phải đối chiếu lại ticket bị ảnh hưởng trước tạo/cập nhật.

- Review code/test nằm trong tiêu chí hoàn thành của mỗi ticket. Chỉ giữ một việc chuẩn bị contract, một việc chuẩn bị dữ liệu/môi trường và một việc kiểm xuyên luồng/bàn giao có đầu ra độc lập.
- Phần tài liệu đã có được dùng làm đầu vào. Nếu review xác nhận đủ, ghi bằng chứng và hoàn tất phần còn lại; không soạn lại hoặc tạo việc giả chỉ để lấp estimate.
- Phần học Go tập trung vào API và luồng HTTP → service → repository → PostgreSQL. Quân cùng agent triển khai từng phần ở AUTH-T15 (me/logout), AUTH-T14 (login), AUTH-T16 (đổi lần đầu), theo phụ thuộc. Trước mỗi phần chốt đoạn Quân tự viết, ví dụ và test; agent giải thích/gỡ rối/review, không tự viết hết phần đã giao. AUTH-T02/AUTH-T10 vẫn là đầu ra kỹ thuật cần thiết, không còn mặc định là hai bài tự viết bắt buộc.
- Metadata đề xuất: Task dưới Epic nếu Jira hỗ trợ; assignee Quân chưa gán accountId, label `auth-session`, thêm `learn-go` cho AUTH-T14/AUTH-T15/AUTH-T16 theo định hướng cùng học qua API. Không tự đặt sprint, deadline, priority hoặc reviewer giả.
- Mỗi ticket có một description hoàn chỉnh bên dưới: mục tiêu, đầu vào, phạm vi, hành vi, tiêu chí hoàn thành, phụ thuộc, kiểm tra và nguồn. Khi chuyển Jira copy toàn bộ phần ticket, đổi đường dẫn repo thành link nguồn đã commit; không chỉ copy checklist.
- Estimate là **cộng cơ sở WBS đã có**, không phải estimate còn lại. Task gom có thể dài hơn 2–5h; các bước trong description là checkpoint làm dần, không bắt nhận thêm nhiều ticket. Cập nhật remaining khi nhận việc, không tự bỏ effort đã biết hoặc cộng trùng.

## 2. Epic

**Summary:** `Đăng nhập và quản lý phiên truy cập backend`

**Mục tiêu:** khép luồng đăng nhập email công ty, danh tính/quyền hiện hành, phiên hạn chế buộc đổi mật khẩu lần đầu và đăng xuất qua FE → HTTP → service → PostgreSQL test.

**Phạm vi:** năm route auth trong OpenAPI; FR-001/FR-002, phần nền hiện hành/thu hồi của FR-003 và phần lần đầu FR-004. API quản trị/dự án chưa có thì không chứng nhận ABAC hoặc toàn FR-003.

**Tiêu chí hoàn thành:**

- [ ] Luồng thường và hạn chế→đổi→login mới→reload/logout có kết quả kiểm xuyên FE/backend/DB test.
- [ ] Cookie/Origin/CSRF/limiter/guard, transaction/race và redaction có bằng chứng AUTH-S01–S10 theo phạm vi.
- [ ] Demo tách auth thật; có runbook, giới hạn và bug/điểm mở, không dùng mock làm bằng chứng quyền server.

**Ngoài phạm vi:** tạo/cấp tài khoản, username, gửi mail, reset/quên mật khẩu, đổi tự nguyện, API nghiệp vụ và nối toàn bộ FE. Không tự áp migration Neon, deploy hoặc commit/push.

**Estimate:** 105h cơ sở từ các ticket dưới, dự phòng 21h còn đề xuất; không nhập effort cộng trùng ở Epic và Task. Chưa ghi timesheet hoặc tự đánh dấu Done.

## 3. Danh mục và mapping WBS

Tên dưới là Summary đưa lên Jira. Mỗi mã WBS xuất hiện đúng một lần trong cột mapping; thứ tự mã không thay phụ thuộc.

| Mã nội bộ | Summary | WBS bao phủ | Giờ cơ sở | Is blocked by | Blocks |
|---|---|---|---:|---|---|
| AUTH-T01 | [AUTH-T01] Hoàn thiện contract auth để bắt đầu triển khai | P01.1, P01.2, P01.3, P01.4 | 12h | — | AUTH-T02, AUTH-T05, AUTH-T07, AUTH-T08, AUTH-T09, AUTH-T10, AUTH-T11, AUTH-T12, AUTH-T13, AUTH-T14, AUTH-T15, AUTH-T16, AUTH-T18 |
| AUTH-T02 | [AUTH-T02] Viết validation thuần cho cấu hình server | P02.1 | 3h | AUTH-T01 | AUTH-T03 |
| AUTH-T03 | [AUTH-T03] Tích hợp load cấu hình và kiểm tra an toàn khi khởi động | P02.2 | 2h | AUTH-T02 | AUTH-T04, AUTH-T12, AUTH-T13 |
| AUTH-T04 | [AUTH-T04] Sửa xử lý URL PostgreSQL và che lỗi kết nối khi khởi động | P03.1, P03.2 | 4h | AUTH-T03 | AUTH-T17 |
| AUTH-T05 | [AUTH-T05] Chuẩn bị kế hoạch dữ liệu và môi trường PostgreSQL test cho auth | P04.1, P04.2, P04.3 | 7h | AUTH-T01 | AUTH-T06, AUTH-T09 |
| AUTH-T06 | [AUTH-T06] Triển khai migration auth và kiểm thử bảo toàn dữ liệu | P05.1 | 4h | AUTH-T05 | AUTH-T07 |
| AUTH-T07 | [AUTH-T07] Triển khai truy vấn định danh và quyền hiện hành | P05.2 | 3h | AUTH-T01, AUTH-T06 | AUTH-T08 |
| AUTH-T08 | [AUTH-T08] Triển khai transaction tạo và thu hồi phiên | P05.3 | 5h | AUTH-T01, AUTH-T07 | AUTH-T09, AUTH-T17 |
| AUTH-T09 | [AUTH-T09] Triển khai transaction đổi mật khẩu lần đầu | P05.4 | 4h | AUTH-T01, AUTH-T05, AUTH-T08 | AUTH-T17 |
| AUTH-T10 | [AUTH-T10] Viết policy phiên và principal với clock truyền vào | P06.1 | 3h | AUTH-T01 | AUTH-T11, AUTH-T12, AUTH-T14, AUTH-T16 |
| AUTH-T11 | [AUTH-T11] Triển khai service xác thực và đăng xuất phiên | P06.3 | 3h | AUTH-T01, AUTH-T10 | AUTH-T12, AUTH-T15, AUTH-T17 |
| AUTH-T12 | [AUTH-T12] Triển khai cookie, kiểm Origin và API CSRF | P07.3 | 4h | AUTH-T01, AUTH-T03, AUTH-T10, AUTH-T11 | AUTH-T14, AUTH-T15, AUTH-T16 |
| AUTH-T13 | [AUTH-T13] Triển khai limiter cho đăng nhập và đổi mật khẩu | P07.5 | 4h | AUTH-T01, AUTH-T03 | AUTH-T14, AUTH-T16 |
| AUTH-T14 | [AUTH-T14] Triển khai luồng API đăng nhập bằng email công ty | P06.2, P07.1 | 7h | AUTH-T01, AUTH-T10, AUTH-T12, AUTH-T13 | AUTH-T17 |
| AUTH-T15 | [AUTH-T15] Triển khai API danh tính hiện hành và đăng xuất phiên | P07.2, P07.4 | 5h | AUTH-T01, AUTH-T11, AUTH-T12 | AUTH-T16, AUTH-T17, AUTH-T18 |
| AUTH-T16 | [AUTH-T16] Triển khai API đổi mật khẩu lần đầu và cổng phiên hạn chế | P06.4, P07.6 | 8h | AUTH-T01, AUTH-T10, AUTH-T12, AUTH-T13, AUTH-T15 | AUTH-T17 |
| AUTH-T17 | [AUTH-T17] Tích hợp năm API auth và kiểm thử transaction cạnh tranh | P09.1, P09.4 | 7h | AUTH-T04, AUTH-T08, AUTH-T09, AUTH-T11, AUTH-T14, AUTH-T15, AUTH-T16 | AUTH-T18, AUTH-T19 |
| AUTH-T18 | [AUTH-T18] Nối FE đăng nhập, khôi phục phiên và đăng xuất với backend | P08.1, P08.2, P08.3 | 10h | AUTH-T01, AUTH-T15, AUTH-T17 | AUTH-T19, AUTH-T20 |
| AUTH-T19 | [AUTH-T19] Nối FE bắt buộc đổi mật khẩu và đăng nhập lại | P08.4 | 4h | AUTH-T17, AUTH-T18 | AUTH-T20 |
| AUTH-T20 | [AUTH-T20] Kiểm tra luồng auth xuyên FE–backend và bàn giao runbook | P09.2, P09.3 | 6h | AUTH-T18, AUTH-T19 | — |

### Thứ tự nhận việc

AUTH-T01 hoàn thiện phần thiếu trong bộ contract đã có; không cần bốn ticket review nối tiếp. Sau đó chuẩn bị nền AUTH-T02–AUTH-T04, dữ liệu AUTH-T05–AUTH-T09 và policy/helper AUTH-T10–AUTH-T13 theo phụ thuộc, không theo số mã máy móc. Phần học API có thể bắt đầu ở AUTH-T15 khi AUTH-T11/AUTH-T12 đã đủ, rồi AUTH-T14 và AUTH-T16 theo đầu vào; không cần chờ mọi adapter DB để học với fake, nhưng phải ghép thật ở AUTH-T17. Giữ một ticket code In Progress, `To Do → In Progress → In Review → Done`; vướng thì flag và ghi nguyên nhân.

Phụ thuộc đã gom từ WBS và loại cạnh nội bộ trong cùng ticket. `A blocks B` và `B is blocked by A` là hai chiều hiển thị của cùng một liên kết: tạo một lần, không tạo hai link đảo chiều hoặc dùng để biểu thị priority. Bảng ghi cả hai chiều; `Blocks` chỉ các ticket phụ thuộc trực tiếp, không liệt kê lại toàn bộ hậu duệ. Khi có Jira key thay mã nội bộ bằng key đã xác minh; điều kiện cần xác nhận vẫn nằm trong description. Đầu ra bộ phận có thể làm sớm với fake (ví dụ service/FE adapter) nhưng chưa ghi cả ticket Done khi integration bắt buộc chưa đạt.

## 4. Description từng ticket

Các gói kiểm tra (được ghi cụ thể trong từng ticket): DOC = kiểm tĩnh contract/harness + diff và review nội dung; GO = kiểm môi trường rồi `go vet ./...`, `go test ./... -count=1`, `go build ./...`; DB = GO và test migration/persistence trên DB test được xác minh; ENV = DOC + xác minh môi trường/quyền sở hữu/guard và cách chạy-dọn được phép, chưa kiểm auth runtime; FE = `pnpm check` trong web và các test HTTP giả/browser; E2E = GO + DB + FE + browser trên môi trường test. Không dùng `make check` như lệnh chỉ đọc. Chưa có runner DB thì phải xây/bàn giao từ AUTH-T05/AUTH-T06, không đoán lệnh hoặc fallback DATABASE_URL.

### [AUTH-T01] Hoàn thiện contract auth để bắt đầu triển khai

**Mục tiêu**

Có một bộ OpenAPI/ports nhất quán, đủ để nhận task code mà không làm lại thiết kế đã có.

**Đầu vào và điều kiện bắt đầu**

Đã có baseline, OpenAPI năm route và ports ở proposal 5.3. Chỉ xử lý phần thiếu/sai và ghi kết luận; nếu đối chiếu thấy đã đủ thì bàn giao bằng chứng, không soạn lại hoặc ghi trọn 12h như effort còn lại. Người dùng đã duyệt phạm vi và các quyết định A01–A09 và body 8 KiB; không yêu cầu duyệt lại.

**Phạm vi thực hiện**

1. Đối chiếu năm route login/me/csrf/logout/change-password với quyết định hiện hành; hoàn thiện request/response/lỗi/security còn thiếu.
2. Chốt chữ ký bốn use case, repository ports, clock/random, lỗi và giao thức khóa; không tạo package Go rỗng.
3. Ghi phần còn mở ngoài scope và điểm cần quyết định trong task sau; giữ một OpenAPI duy nhất.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Contract HTTP | Mỗi route có input, điều kiện bảo vệ, kết quả/lỗi; DTO không lộ hash/secret hoặc cho client truyền quyền. |
| Phiên hạn chế | Allowlist/cờ, đổi lần đầu, giới hạn mật khẩu/body/limiter và hành vi FE khớp quyết định đã chốt. |
| Contract Go | Service không phụ thuộc pgx/HTTP; trách nhiệm transaction và kiểm lại nền rõ cho adapter/tests. |

**Tiêu chí hoàn thành**

- [ ] Có OpenAPI/ports đã hoàn thiện hoặc kết luận không cần sửa, kèm danh sách đối chiếu.
- [ ] Kiểm tĩnh contract/harness đạt; ghi rõ chưa có handler/runtime, không coi việc soạn tài liệu là code xong.
- [ ] Diff/nội dung đã review; ghi lệnh, kết quả thực chạy và giới hạn. Có đầu ra bàn giao/commit/PR khi có; không nhận test fake/tĩnh thay bằng chứng DB/browser.

**Phụ thuộc / cách thực hiện / estimate**

- Bị chặn bởi: phạm vi đã duyệt và nguồn hiện hành; không có ticket tiền nhiệm.
- Blocks: AUTH-T02, AUTH-T05, AUTH-T07, AUTH-T08, AUTH-T09, AUTH-T10, AUTH-T11, AUTH-T12, AUTH-T13, AUTH-T14, AUTH-T15, AUTH-T16, AUTH-T18.
- WBS: P01.1, P01.2, P01.3, P01.4; cơ sở 12h, remaining/actual chưa xác định; không tính lại review như task riêng.
- Agent hỗ trợ triển khai/soạn; Quân review/kiểm tra và cùng chốt brief trước làm.

**Kiểm tra và bằng chứng**

Gói DOC cùng các ca trong bảng; phần bảo mật liên quan: AUTH-S10. Theo proposal 6.4.1/9.2.2, chỉ nhận phần bằng chứng đúng tầng của ticket. Bàn giao tên file/lệnh chạy/kết quả/chưa kiểm cho task sau; không ghi trước pass hoặc đưa secret vào Jira.

**Ngoài phạm vi**

Không chọn lại nghiệp vụ, viết backend, reset/gửi email hoặc làm contract profile ngoài feature.

**Tham chiếu**

[Proposal](backend-auth-session.md), mục 4–6.2, 9.2.1, WBS P01.1, P01.2, P01.3, P01.4; [FR/AC](../product/functional-requirements.md): FR-001/FR-002; phần nền FR-003; phần lần đầu FR-004; [OpenAPI](../api/auth.openapi.json). Đọc workflow/architecture/security do proposal dẫn khi triển khai, không copy toàn thiết kế vào ticket.

### [AUTH-T02] Viết validation thuần cho cấu hình server

**Mục tiêu**

Tạo validation dễ kiểm thử cho cấu hình server; phần học Go chính chuyển sang các API theo mục 1.

**Đầu vào và điều kiện bắt đầu**

Có loader/config cũ; chưa có validation auth của feature. Task brief phải nêu input/tên/default cấu hình cụ thể trước khi Quân viết. Chốt contract config trong task brief; không tự chọn default mới.

**Phạm vi thực hiện**

1. Tách hàm kiểm env/port/Origin/TTL trên input cấu hình, không đọc OS/file trong hàm thuần.
2. Trả lỗi có ngữ cảnh an toàn; không kèm secret/URL đầy đủ.
3. Viết table-driven tests trước khi nối loader.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Input hợp lệ | Được chấp nhận theo contract config đã chốt trong brief. |
| Thiếu/sai/biên | Lỗi đúng trường hợp, gồm dev/prod, port và Origin/TTL; không tự fallback ngoài contract. |
| Lỗi có dữ liệu nhạy cảm giả | Không lặp marker bí mật trong thông báo. |

**Tiêu chí hoàn thành**

- [ ] Có validation thuần và table-driven tests; phân công phần tự viết nếu Quân muốn nhận thêm trong brief.
- [ ] Test không cần DB hoặc .env thật; cổng Go đạt.
- [ ] Diff/nội dung đã review; ghi lệnh, kết quả thực chạy và giới hạn. Có đầu ra bàn giao/commit/PR khi có; không nhận test fake/tĩnh thay bằng chứng DB/browser.

**Phụ thuộc / cách thực hiện / estimate**

- Bị chặn bởi: AUTH-T01.
- Blocks: AUTH-T03.
- WBS: P02.1; cơ sở 3h, remaining/actual chưa xác định; không tính lại review như task riêng.
- Agent hỗ trợ triển khai; Quân cùng đọc hiểu/review, phần tự viết thêm chỉ giao khi chốt brief.

**Kiểm tra và bằng chứng**

Gói GO cùng các ca trong bảng; phần bảo mật liên quan: AUTH-S09. Theo proposal 6.4.1/9.2.2, chỉ nhận phần bằng chứng đúng tầng của ticket. Bàn giao tên file/lệnh chạy/kết quả/chưa kiểm cho task sau; không ghi trước pass hoặc đưa secret vào Jira.

**Ngoài phạm vi**

Không loader I/O, kết nối DB hoặc thay guard migration/seed.

**Tham chiếu**

[Proposal](backend-auth-session.md), mục 6.3, 9.3, WBS P02.1; [FR/AC](../product/functional-requirements.md): Hỗ trợ FR-001.AC2/AC5 và security; [OpenAPI](../api/auth.openapi.json). Đọc workflow/architecture/security do proposal dẫn khi triển khai, không copy toàn thiết kế vào ticket.

### [AUTH-T03] Tích hợp load cấu hình và kiểm tra an toàn khi khởi động

**Mục tiêu**

Đưa validation vào startup mà giữ hành vi cấu hình có thể kiểm chứng.

**Đầu vào và điều kiện bắt đầu**

Có validation AUTH-T02 và loader hiện hành; dùng môi trường/file tạm để test. Chỉ dùng input cấu hình đã chốt ở AUTH-T02.

**Phạm vi thực hiện**

1. Nối loader với validation trước khi cấu hình được dùng.
2. Chỉ nạp dotenv ở dev tường minh; xử lý precedence/missing theo contract.
3. Giữ RequireDev và guard migrate/seed.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Cấu hình môi trường và file tạm | Precedence đúng brief, không đọc .env thật. |
| Prod cấu hình thiếu/sai | Startup lỗi rõ, không tự rơi về dev. |
| Migrate/seed | Guard cũ vẫn có hiệu lực; không bỏ guard để test xanh. |

**Tiêu chí hoàn thành**

- [ ] Có test loader bằng input tạm và đường startup lỗi/thành công phù hợp.
- [ ] Không hồi quy guard, lỗi không lộ credential; cổng Go đạt.
- [ ] Diff/nội dung đã review; ghi lệnh, kết quả thực chạy và giới hạn. Có đầu ra bàn giao/commit/PR khi có; không nhận test fake/tĩnh thay bằng chứng DB/browser.

**Phụ thuộc / cách thực hiện / estimate**

- Bị chặn bởi: AUTH-T02.
- Blocks: AUTH-T04, AUTH-T12, AUTH-T13.
- WBS: P02.2; cơ sở 2h, remaining/actual chưa xác định; không tính lại review như task riêng.
- Agent hỗ trợ triển khai/soạn; Quân review/kiểm tra và cùng chốt brief trước làm.

**Kiểm tra và bằng chứng**

Gói GO cùng các ca trong bảng; phần bảo mật liên quan: Chuẩn bị đầu vào; chưa có ca AUTH-S runtime để nhận pass. Theo proposal 6.4.1/9.2.2, chỉ nhận phần bằng chứng đúng tầng của ticket. Bàn giao tên file/lệnh chạy/kết quả/chưa kiểm cho task sau; không ghi trước pass hoặc đưa secret vào Jira.

**Ngoài phạm vi**

Không chạy migration/seed/DB thật.

**Tham chiếu**

[Proposal](backend-auth-session.md), mục 6.3, WBS P02.2; [FR/AC](../product/functional-requirements.md): Hỗ trợ FR-001.AC2/AC5 và security; [OpenAPI](../api/auth.openapi.json). Đọc workflow/architecture/security do proposal dẫn khi triển khai, không copy toàn thiết kế vào ticket.

### [AUTH-T04] Sửa xử lý URL PostgreSQL và che lỗi kết nối khi khởi động

**Mục tiêu**

Loại bỏ thay chuỗi URL thiếu chính xác và rò credential trong lỗi startup.

**Đầu vào và điều kiện bắt đầu**

Có helper pooled/direct URL và đường startup cũ; chưa sửa trong feature. Định dạng URL được hỗ trợ cần chốt trong brief trước sửa.

**Phạm vi thực hiện**

1. Review định dạng hostname hỗ trợ rồi parse URL/hostname, không thay chuỗi trên toàn DSN.
2. Giữ credential/path/query không liên quan khi chuyển endpoint hợp lệ.
3. Che lỗi startup/driver bằng ngữ cảnh chẩn đoán an toàn.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Marker pooled trong password/path/query | Không bị nhận nhầm thành hostname hoặc bị sửa. |
| URL sai/không hỗ trợ | Hành vi từ chối/xử lý đúng contract đã review, không đoán cách chuyển. |
| Lỗi giả chứa DSN/password | Log/response không phát marker, vẫn có ngữ cảnh chẩn đoán. |

**Tiêu chí hoàn thành**

- [ ] Có unit tests URL và redaction bằng dữ liệu giả, không DB thật.
- [ ] Cổng Go đạt; giữ được ngữ cảnh startup và không lộ secret.
- [ ] Diff/nội dung đã review; ghi lệnh, kết quả thực chạy và giới hạn. Có đầu ra bàn giao/commit/PR khi có; không nhận test fake/tĩnh thay bằng chứng DB/browser.

**Phụ thuộc / cách thực hiện / estimate**

- Bị chặn bởi: AUTH-T03.
- Blocks: AUTH-T17.
- WBS: P03.1, P03.2; cơ sở 4h, remaining/actual chưa xác định; không tính lại review như task riêng.
- Agent hỗ trợ triển khai/soạn; Quân review/kiểm tra và cùng chốt brief trước làm.

**Kiểm tra và bằng chứng**

Gói GO cùng các ca trong bảng; phần bảo mật liên quan: AUTH-S09. Theo proposal 6.4.1/9.2.2, chỉ nhận phần bằng chứng đúng tầng của ticket. Bàn giao tên file/lệnh chạy/kết quả/chưa kiểm cho task sau; không ghi trước pass hoặc đưa secret vào Jira.

**Ngoài phạm vi**

Không đổi provider, dùng URL thật hoặc tự kết nối Neon.

**Tham chiếu**

[Proposal](backend-auth-session.md), mục 6.3, 6.4.1, WBS P03.1, P03.2; [FR/AC](../product/functional-requirements.md): Hỗ trợ nền auth; CR-001.AC2/security; [OpenAPI](../api/auth.openapi.json). Đọc workflow/architecture/security do proposal dẫn khi triển khai, không copy toàn thiết kế vào ticket.

### [AUTH-T05] Chuẩn bị kế hoạch dữ liệu và môi trường PostgreSQL test cho auth

**Mục tiêu**

Bàn giao môi trường test an toàn cùng SQL/mapping/fixture plan để bắt đầu migration.

**Đầu vào và điều kiện bắt đầu**

Đã có DDL/kế hoạch, kiểm kê lịch sử và diễn tập riêng 000002; chưa có PostgreSQL test/fixture auth được xác minh. Mapping/SQL cụ thể phải được review trước thực thi; quyền dựng môi trường không phải quyền áp Neon.

**Phạm vi thực hiện**

1. Hoàn thiện SQL/mapping quyền legacy, prerequisite/FK, đường nâng cấp và rollback cho phạm vi auth.
2. Kiểm Docker/PostgreSQL major 18 và dựng tài nguyên test được phép; xác nhận host/name/user/quyền sở hữu trước DDL/dọn.
3. Chuẩn bị TEST_DATABASE_URL/guard và cách chạy/dọn; không fallback DATABASE_URL hoặc nạp .env thật.
4. Đặc tả fixture cờ true/false, role/active/email mâu thuẫn, nhiều phiên và race; bàn giao AUTH-T06/AUTH-T07 thực thi sau có bảng.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| SQL/mapping | Được review cho đúng môi trường; không suy HR/DIRECTOR từ tên/vị trí. |
| DB đích không thuộc test sở hữu | Runner/guard từ chối thao tác thay vì dùng nhầm DB phát triển. |
| Fixture auth chưa có bảng | Chỉ thiết kế; không insert trước migration hoặc gọi plan là test đã chạy. |
| Đường migration | Có ca test rỗng và legacy giả, gồm 000002; tận dụng bằng chứng cũ, không diễn tập lại production. |

**Tiêu chí hoàn thành**

- [ ] Có kế hoạch SQL/mapping được xác nhận và tài nguyên test/cách chạy-dọn được kiểm chứng trong phạm vi cho phép.
- [ ] Có lệnh/guard cụ thể dùng cho AUTH-T06; chưa dựng được môi trường thì chưa Done, không lấy DB thật thay thế.
- [ ] Diff/nội dung đã review; ghi lệnh, kết quả thực chạy và giới hạn. Có đầu ra bàn giao/commit/PR khi có; không nhận test fake/tĩnh thay bằng chứng DB/browser.

**Phụ thuộc / cách thực hiện / estimate**

- Bị chặn bởi: AUTH-T01.
- Blocks: AUTH-T06, AUTH-T09.
- WBS: P04.1, P04.2, P04.3; cơ sở 7h, remaining/actual chưa xác định; không tính lại review như task riêng.
- Agent hỗ trợ triển khai/soạn; Quân review/kiểm tra và cùng chốt brief trước làm.

**Kiểm tra và bằng chứng**

Gói ENV cùng các ca trong bảng; phần bảo mật liên quan: Chuẩn bị đầu vào; chưa có ca AUTH-S runtime để nhận pass. Theo proposal 6.4.1/9.2.2, chỉ nhận phần bằng chứng đúng tầng của ticket. Bàn giao tên file/lệnh chạy/kết quả/chưa kiểm cho task sau; không ghi trước pass hoặc đưa secret vào Jira.

**Ngoài phạm vi**

Không chạy SQL production, sửa 000001, migrate cả 33 bảng hoặc cấp credential thật.

**Tham chiếu**

[Proposal](backend-auth-session.md), mục 6.1.3–6.1.6, WBS P04.1, P04.2, P04.3; [FR/AC](../product/functional-requirements.md): Hỗ trợ FR-001–FR-004 trong phạm vi; CR-002; [OpenAPI](../api/auth.openapi.json). Đọc workflow/architecture/security do proposal dẫn khi triển khai, không copy toàn thiết kế vào ticket.

### [AUTH-T06] Triển khai migration auth và kiểm thử bảo toàn dữ liệu

**Mục tiêu**

Tạo nền user_roles/auth_sessions trên DB test bằng migration tiến thêm.

**Đầu vào và điều kiện bắt đầu**

Có SQL/mapping và môi trường AUTH-T05; kiểm lại migration head trước đặt số file. Chỉ thực thi SQL/mapping đã duyệt trong môi trường test AUTH-T05.

**Phạm vi thực hiện**

1. Viết migration auth đúng schema đã duyệt và nối runner vào URL test đã xác minh.
2. Hiện thực fixture migration tối thiểu; nạp quyền/phiên sau khi bảng tồn tại.
3. Chạy đường test rỗng và legacy giả; kiểm constraints/index/mapping/bảo toàn.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Nâng DB rỗng | Qua đúng 000001/000002/prerequisite và migration auth, không force version. |
| Nâng legacy giả | ID/link user–employee, hash/username/email/status được giữ; quyền mới đúng mapping. |
| Ràng buộc/lỗi migration | PK/FK/unique/check/index được kiểm; trạng thái dirty/rollback đúng driver thực, không che lỗi. |

**Tiêu chí hoàn thành**

- [ ] Có file migration và bằng chứng hai đường nâng cấp trên DB test sở hữu.
- [ ] Có kết quả bảo toàn/ràng buộc và cách tái hiện, cổng Go/DB đạt.
- [ ] Diff/nội dung đã review; ghi lệnh, kết quả thực chạy và giới hạn. Có đầu ra bàn giao/commit/PR khi có; không nhận test fake/tĩnh thay bằng chứng DB/browser.

**Phụ thuộc / cách thực hiện / estimate**

- Bị chặn bởi: AUTH-T05.
- Blocks: AUTH-T07.
- WBS: P05.1; cơ sở 4h, remaining/actual chưa xác định; không tính lại review như task riêng.
- Agent hỗ trợ triển khai/soạn; Quân review/kiểm tra và cùng chốt brief trước làm.

**Kiểm tra và bằng chứng**

Gói DB cùng các ca trong bảng; phần bảo mật liên quan: Chuẩn bị đầu vào; chưa có ca AUTH-S runtime để nhận pass. Theo proposal 6.4.1/9.2.2, chỉ nhận phần bằng chứng đúng tầng của ticket. Bàn giao tên file/lệnh chạy/kết quả/chưa kiểm cho task sau; không ghi trước pass hoặc đưa secret vào Jira.

**Ngoài phạm vi**

Không reset DB thật, sửa 000001, tự chạy down production hoặc migrate toàn thiết kế.

**Tham chiếu**

[Proposal](backend-auth-session.md), mục 6.1.1, 6.1.3–6.1.6, WBS P05.1; [FR/AC](../product/functional-requirements.md): Hỗ trợ FR-001/FR-002; CR-002; [OpenAPI](../api/auth.openapi.json). Đọc workflow/architecture/security do proposal dẫn khi triển khai, không copy toàn thiết kế vào ticket.

### [AUTH-T07] Triển khai truy vấn định danh và quyền hiện hành

**Mục tiêu**

Cung cấp projection auth tối thiểu và dữ liệu kiểm thử dùng lại cho transaction.

**Đầu vào và điều kiện bắt đầu**

Schema test đã migrate; fixture ca auth đã đặc tả ở AUTH-T05. Fixture chỉ nạp vào schema đã migrate ở AUTH-T06.

**Phạm vi thực hiện**

1. Hiện thực lookup email chuẩn hóa và đọc session/principal/role theo ports.
2. Hiện thực fixture auth trên schema test cho ca thường/hạn chế/khóa/trùng email.
3. Phân loại không tìm thấy và lỗi DB/dữ liệu; không đẩy entity thô lên HTTP.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Không tìm thấy email sau chuẩn hóa | Repository trả kết quả không tìm thấy, phân biệt với lỗi DB; luồng login trả 401 INVALID_CREDENTIALS chung, không tiết lộ email tồn tại. |
| Nhiều tài khoản trùng email sau chuẩn hóa | Xử lý là lỗi dữ liệu; không LIMIT 1 chọn tùy ý, không cấp phiên; luồng HTTP trả 500 INTERNAL_ERROR chung theo contract. |
| Quyền/active/cờ đổi | Projection đọc nguồn hiện hành; không suy quyền từ legacy role/position. |
| user_id khác employee_id | Liên kết vẫn đúng; DTO/projection không làm lộ trường ngoài phép. |

**Tiêu chí hoàn thành**

- [ ] Persistence tests chứng minh các ca trên và tập role kết hợp đúng nguồn.
- [ ] Fixture dùng lại cho AUTH-T08/AUTH-T09, không SELECT * ra response hoặc in hash/PII.
- [ ] Diff/nội dung đã review; ghi lệnh, kết quả thực chạy và giới hạn. Có đầu ra bàn giao/commit/PR khi có; không nhận test fake/tĩnh thay bằng chứng DB/browser.

**Phụ thuộc / cách thực hiện / estimate**

- Bị chặn bởi: AUTH-T01, AUTH-T06.
- Blocks: AUTH-T08.
- WBS: P05.2; cơ sở 3h, remaining/actual chưa xác định; không tính lại review như task riêng.
- Agent hỗ trợ triển khai/soạn; Quân review/kiểm tra và cùng chốt brief trước làm.

**Kiểm tra và bằng chứng**

Gói DB cùng các ca trong bảng; phần bảo mật liên quan: AUTH-S01, AUTH-S05. Theo proposal 6.4.1/9.2.2, chỉ nhận phần bằng chứng đúng tầng của ticket. Bàn giao tên file/lệnh chạy/kết quả/chưa kiểm cho task sau; không ghi trước pass hoặc đưa secret vào Jira.

**Ngoài phạm vi**

Không API hồ sơ, quyền dự án hoặc backfill ngoài mapping đã duyệt.

**Tham chiếu**

[Proposal](backend-auth-session.md), mục 5.3, 6.1.2/6.1.6, WBS P05.2; [FR/AC](../product/functional-requirements.md): FR-001.AC1/AC3/AC5; phần nền FR-003; [OpenAPI](../api/auth.openapi.json). Đọc workflow/architecture/security do proposal dẫn khi triển khai, không copy toàn thiết kế vào ticket.

### [AUTH-T08] Triển khai transaction tạo và thu hồi phiên

**Mục tiêu**

Bảo đảm session chỉ được cấp sau kiểm lại nền và logout đúng một phiên.

**Đầu vào và điều kiện bắt đầu**

Có projection/fixture AUTH-T07 và giao thức khóa trong contract AUTH-T01. Mutation khóa/reset trong test chỉ mô phỏng giao thức DB, không chứng nhận API quản trị đã có.

**Phạm vi thực hiện**

1. Hiện thực FinalizeLogin/revoke/logout qua adapter theo ports.
2. Kiểm lại hash/status dưới khóa; xử lý cookie cũ cùng user và thời gian sau chờ khóa theo contract.
3. Kiểm rollback, lỗi commit và race với mutation giả lập có barrier.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Login thành công | Session lưu hash; chỉ commit thành công mới có kết quả cấp phiên. |
| Nền đã đổi hoặc commit lỗi | Không cấp phiên thành công giả hoặc để write nửa chừng. |
| Logout/cookie cũ | Thu hồi đúng phiên theo contract; không thu hồi user khác hoặc mọi thiết bị. |

**Tiêu chí hoàn thành**

- [ ] Có DB tests với hai connection/barrier cho ca cạnh tranh bắt buộc, không sleep ngẫu nhiên.
- [ ] Kết quả rollback/revoke/time boundary được lưu và bàn giao service/integration.
- [ ] Diff/nội dung đã review; ghi lệnh, kết quả thực chạy và giới hạn. Có đầu ra bàn giao/commit/PR khi có; không nhận test fake/tĩnh thay bằng chứng DB/browser.

**Phụ thuộc / cách thực hiện / estimate**

- Bị chặn bởi: AUTH-T01, AUTH-T07.
- Blocks: AUTH-T09, AUTH-T17.
- WBS: P05.3; cơ sở 5h, remaining/actual chưa xác định; không tính lại review như task riêng.
- Agent hỗ trợ triển khai/soạn; Quân review/kiểm tra và cùng chốt brief trước làm.

**Kiểm tra và bằng chứng**

Gói DB cùng các ca trong bảng; phần bảo mật liên quan: AUTH-S02, AUTH-S07. Theo proposal 6.4.1/9.2.2, chỉ nhận phần bằng chứng đúng tầng của ticket. Bàn giao tên file/lệnh chạy/kết quả/chưa kiểm cho task sau; không ghi trước pass hoặc đưa secret vào Jira.

**Ngoài phạm vi**

Không tạo API reset/khóa hoặc hứa hủy ngược request đã commit.

**Tham chiếu**

[Proposal](backend-auth-session.md), mục 5.3, 6.1.2, WBS P05.3; [FR/AC](../product/functional-requirements.md): FR-001.AC1/AC2; FR-002.AC1–AC2; phần nền FR-003; CR-002; [OpenAPI](../api/auth.openapi.json). Đọc workflow/architecture/security do proposal dẫn khi triển khai, không copy toàn thiết kế vào ticket.

### [AUTH-T09] Triển khai transaction đổi mật khẩu lần đầu

**Mục tiêu**

Đổi hash/cờ và thu hồi mọi phiên đúng user trong cùng transaction.

**Đầu vào và điều kiện bắt đầu**

Có session adapter/fixture và thiết kế cờ đã duyệt. Giữ cột/schema đã duyệt, không tự thêm scope/version.

**Phạm vi thực hiện**

1. Hiện thực kiểm lại user/employee/session/hash/cờ dưới thứ tự khóa thống nhất.
2. Thay hash, đặt cờ false và revoke mọi phiên cùng user atomically.
3. Kiểm race hai lần đổi, login hash cũ chờ khóa và lỗi giữa các write.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Đổi đúng | Ba thay đổi commit cùng nhau, user khác giữ phiên. |
| Hai lần đổi hoặc nền đã đổi | Chỉ một lần hợp lệ commit; lần còn lại không ghi đè bằng snapshot cũ. |
| Lỗi write/chờ khóa vượt hạn | Rollback hoặc từ chối đúng contract; không có trạng thái nửa chừng. |

**Tiêu chí hoàn thành**

- [ ] DB tests hai connection/fake clock/barrier chứng minh race/rollback.
- [ ] Không sinh session mới sau đổi; có bằng chứng để AUTH-T17 kiểm xuyên HTTP.
- [ ] Diff/nội dung đã review; ghi lệnh, kết quả thực chạy và giới hạn. Có đầu ra bàn giao/commit/PR khi có; không nhận test fake/tĩnh thay bằng chứng DB/browser.

**Phụ thuộc / cách thực hiện / estimate**

- Bị chặn bởi: AUTH-T01, AUTH-T05, AUTH-T08.
- Blocks: AUTH-T17.
- WBS: P05.4; cơ sở 4h, remaining/actual chưa xác định; không tính lại review như task riêng.
- Agent hỗ trợ triển khai/soạn; Quân review/kiểm tra và cùng chốt brief trước làm.

**Kiểm tra và bằng chứng**

Gói DB cùng các ca trong bảng; phần bảo mật liên quan: AUTH-S06, AUTH-S07. Theo proposal 6.4.1/9.2.2, chỉ nhận phần bằng chứng đúng tầng của ticket. Bàn giao tên file/lệnh chạy/kết quả/chưa kiểm cho task sau; không ghi trước pass hoặc đưa secret vào Jira.

**Ngoài phạm vi**

Không đổi tự nguyện, reset hoặc tự áp SQL production.

**Tham chiếu**

[Proposal](backend-auth-session.md), mục 5.3, 6.1.2, 6.2.5, WBS P05.4; [FR/AC](../product/functional-requirements.md): FR-001.AC4; FR-004.AC1–AC4 phần lần đầu; CR-002; [OpenAPI](../api/auth.openapi.json). Đọc workflow/architecture/security do proposal dẫn khi triển khai, không copy toàn thiết kế vào ticket.

### [AUTH-T10] Viết policy phiên và principal với clock truyền vào

**Mục tiêu**

Tạo logic phiên thuần dữ liệu; giải thích type/interface và clock để Quân dùng khi triển khai API.

**Đầu vào và điều kiện bắt đầu**

Ports/policy contract AUTH-T01 đã rõ; fake clock, không cần DB adapter. Không cần chờ AUTH-T06–AUTH-T09 để làm với fake.

**Phạm vi thực hiện**

1. Định nghĩa principal/policy cần thiết theo contract, tránh interface cho mọi hàm thuần.
2. Kiểm expiry/revoked/user và employee active từ input hiện hành.
3. Viết table-driven tests với clock điều khiển được.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| now bằng expires_at | Phiên bị từ chối, không tự gia hạn do hoạt động. |
| Revoked hoặc inactive | Từ chối đúng policy. |
| Không có role đặc biệt | Vẫn là principal nền hợp lệ khi điều kiện khác đạt; không suy PM/PTCM từ vị trí. |

**Tiêu chí hoàn thành**

- [ ] Có policy/tests và giải thích bất biến phiên; phân công phần tự viết nếu Quân nhận thêm trong brief.
- [ ] Logic không import DB/HTTP, test không DB và cổng Go đạt.
- [ ] Diff/nội dung đã review; ghi lệnh, kết quả thực chạy và giới hạn. Có đầu ra bàn giao/commit/PR khi có; không nhận test fake/tĩnh thay bằng chứng DB/browser.

**Phụ thuộc / cách thực hiện / estimate**

- Bị chặn bởi: AUTH-T01.
- Blocks: AUTH-T11, AUTH-T12, AUTH-T14, AUTH-T16.
- WBS: P06.1; cơ sở 3h, remaining/actual chưa xác định; không tính lại review như task riêng.
- Agent hỗ trợ triển khai; Quân cùng đọc hiểu/review, phần tự viết thêm chỉ giao khi chốt brief.

**Kiểm tra và bằng chứng**

Gói GO cùng các ca trong bảng; phần bảo mật liên quan: AUTH-S02. Theo proposal 6.4.1/9.2.2, chỉ nhận phần bằng chứng đúng tầng của ticket. Bàn giao tên file/lệnh chạy/kết quả/chưa kiểm cho task sau; không ghi trước pass hoặc đưa secret vào Jira.

**Ngoài phạm vi**

Không dựng RBAC engine hoặc role dự án toàn hệ thống.

**Tham chiếu**

[Proposal](backend-auth-session.md), mục 5.3, 6.2.4, 9.3, WBS P06.1; [FR/AC](../product/functional-requirements.md): FR-001.AC2/AC3/AC5; phần nền FR-003; [OpenAPI](../api/auth.openapi.json). Đọc workflow/architecture/security do proposal dẫn khi triển khai, không copy toàn thiết kế vào ticket.

### [AUTH-T11] Triển khai service xác thực và đăng xuất phiên

**Mục tiêu**

Đọc quyền hiện hành và revoke phiên qua port mà không phụ thuộc transport.

**Đầu vào và điều kiện bắt đầu**

Có policy AUTH-T10 và fake repository theo contract AUTH-T01. Runtime DB sẽ kiểm tại AUTH-T17, fake không thay bằng chứng đó.

**Phạm vi thực hiện**

1. Viết Authenticate dùng snapshot hiện hành và policy.
2. Viết Logout thu hồi phiên qua port; phân loại lỗi không có phiên với lỗi DB.
3. Kiểm bằng fake/clock; không query trực tiếp từ service.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Snapshot hợp lệ | Trả principal hiện hành, không quyền lấy từ client. |
| Phiên thiếu/hết hạn/revoked/inactive | Bị từ chối; DB lỗi không đổi thành 401 giả. |
| Logout | Chỉ phiên hiện tại bị thu hồi, phiên khác giữ nguyên. |

**Tiêu chí hoàn thành**

- [ ] Unit tests fake repo/clock bao phủ ca thành công/lỗi và dữ liệu thay đổi.
- [ ] Service không import pgx/HTTP; cổng Go đạt.
- [ ] Diff/nội dung đã review; ghi lệnh, kết quả thực chạy và giới hạn. Có đầu ra bàn giao/commit/PR khi có; không nhận test fake/tĩnh thay bằng chứng DB/browser.

**Phụ thuộc / cách thực hiện / estimate**

- Bị chặn bởi: AUTH-T01, AUTH-T10.
- Blocks: AUTH-T12, AUTH-T15, AUTH-T17.
- WBS: P06.3; cơ sở 3h, remaining/actual chưa xác định; không tính lại review như task riêng.
- Agent hỗ trợ triển khai/soạn; Quân review/kiểm tra và cùng chốt brief trước làm.

**Kiểm tra và bằng chứng**

Gói GO cùng các ca trong bảng; phần bảo mật liên quan: AUTH-S05. Theo proposal 6.4.1/9.2.2, chỉ nhận phần bằng chứng đúng tầng của ticket. Bàn giao tên file/lệnh chạy/kết quả/chưa kiểm cho task sau; không ghi trước pass hoặc đưa secret vào Jira.

**Ngoài phạm vi**

Không bảo đảm đảo ngược request đã commit hoặc làm API quản trị.

**Tham chiếu**

[Proposal](backend-auth-session.md), mục 5.3, 6.1.2, 6.2.4, WBS P06.3; [FR/AC](../product/functional-requirements.md): FR-001.AC2/AC3/AC5; FR-002.AC1–AC2; phần nền FR-003; [OpenAPI](../api/auth.openapi.json). Đọc workflow/architecture/security do proposal dẫn khi triển khai, không copy toàn thiết kế vào ticket.

### [AUTH-T12] Triển khai cookie, kiểm Origin và API CSRF

**Mục tiêu**

Cung cấp helper bảo vệ mutation và bootstrap CSRF cho phiên hợp lệ.

**Đầu vào và điều kiện bắt đầu**

Có config và Authenticate service; chi/stdlib/dependency hiện hữu. Proxy/Origin dev phải khớp contract cấu hình, không tin forwarded header tùy ý.

**Phạm vi thực hiện**

1. Hiện thực tạo/xóa cookie đúng thuộc tính và expiry, không cất secret trong JSON/localStorage.
2. Hiện thực kiểm Origin và HMAC CSRF gắn phiên, so constant-time.
3. Nối GET /api/auth/csrf với Authenticate, áp no-store và map lỗi.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Phiên thường/hạn chế hợp lệ | GET csrf trả 200 với token dẫn xuất; không tạo/gia hạn session. |
| Phiên lỗi/DB lỗi | 401 hoặc 503/500 theo contract, không token thành công giả. |
| Origin/CSRF mutation sai | Bị từ chối; token cũ không hợp lệ sau rotation. |
| Cookie local/prod | Đúng HttpOnly/SameSite/path/host-only/Secure theo nguồn và ngoại lệ dev tường minh. |

**Tiêu chí hoàn thành**

- [ ] httptest kiểm cookie/clear-cookie, token/rotation, Origin scheme/host/port và thiếu/sai token.
- [ ] GET csrf không mở CORS rộng hoặc sinh phiên; cổng Go đạt.
- [ ] Diff/nội dung đã review; ghi lệnh, kết quả thực chạy và giới hạn. Có đầu ra bàn giao/commit/PR khi có; không nhận test fake/tĩnh thay bằng chứng DB/browser.

**Phụ thuộc / cách thực hiện / estimate**

- Bị chặn bởi: AUTH-T01, AUTH-T03, AUTH-T10, AUTH-T11.
- Blocks: AUTH-T14, AUTH-T15, AUTH-T16.
- WBS: P07.3; cơ sở 4h, remaining/actual chưa xác định; không tính lại review như task riêng.
- Agent hỗ trợ triển khai/soạn; Quân review/kiểm tra và cùng chốt brief trước làm.

**Kiểm tra và bằng chứng**

Gói GO cùng các ca trong bảng; phần bảo mật liên quan: AUTH-S02, AUTH-S04, AUTH-S09, AUTH-S10. Theo proposal 6.4.1/9.2.2, chỉ nhận phần bằng chứng đúng tầng của ticket. Bàn giao tên file/lệnh chạy/kết quả/chưa kiểm cho task sau; không ghi trước pass hoặc đưa secret vào Jira.

**Ngoài phạm vi**

Không bảng CSRF, CORS rộng hoặc session library mới.

**Tham chiếu**

[Proposal](backend-auth-session.md), mục 6.2.4–6.2.5; OpenAPI csrf, WBS P07.3; [FR/AC](../product/functional-requirements.md): Hỗ trợ FR-001.AC2/AC4; FR-002; FR-004 theo security; [OpenAPI](../api/auth.openapi.json). Đọc workflow/architecture/security do proposal dẫn khi triển khai, không copy toàn thiết kế vào ticket.

### [AUTH-T13] Triển khai limiter cho đăng nhập và đổi mật khẩu

**Mục tiêu**

Giới hạn abuse bằng hai policy độc lập với clock và bộ nhớ có giới hạn.

**Đầu vào và điều kiện bắt đầu**

Có ngưỡng/policy trong contract, không chọn lại tham số. Ngưỡng login thử nghiệm và ngưỡng đổi riêng theo nguồn hiện hành.

**Phạm vi thực hiện**

1. Hiện thực bucket login IP/email và đổi lần đầu IP/user đã xác thực.
2. Tách kiểm trước bcrypt với ghi nhận xác minh sai; tính Retry-After theo policy.
3. Giới hạn kích thước/cleanup bucket và bảo vệ thao tác đồng thời.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Giới hạn login theo IP | Kiểm biên 30/31 request trong 5 phút/IP theo tham số thử nghiệm ở proposal 6.2.4; vượt giới hạn trả 429 kèm Retry-After. |
| Giới hạn login theo email | Kiểm biên 5/6 lần xác thực thất bại trong 5 phút/email chuẩn hóa; không dùng cửa sổ 15 phút của luồng đổi mật khẩu. |
| Giới hạn đổi mật khẩu theo IP | Kiểm biên 30/31 request trong 5 phút/IP theo proposal 6.2.5. |
| Giới hạn đổi mật khẩu theo user | Kiểm biên 5/6 lần xác minh mật khẩu hiện tại sai trong 15 phút/user đã xác thực; không dùng email làm khóa hoặc cửa sổ 5 phút của login. |
| Validation/CSRF/DB lỗi | Không tính nhầm thành sai mật khẩu hiện tại. |
| Cửa sổ hết/đồng thời/spoof header | Cleanup và quyết định giới hạn đúng, không tin X-Forwarded-For chưa cấu hình. |

**Tiêu chí hoàn thành**

- [ ] Unit tests clock/biên/cửa sổ/đồng thời/bộ nhớ và Retry-After đạt.
- [ ] Có API limiter dùng lại cho AUTH-T14/AUTH-T16, không khóa account vĩnh viễn hoặc log raw email/password.
- [ ] Diff/nội dung đã review; ghi lệnh, kết quả thực chạy và giới hạn. Có đầu ra bàn giao/commit/PR khi có; không nhận test fake/tĩnh thay bằng chứng DB/browser.

**Phụ thuộc / cách thực hiện / estimate**

- Bị chặn bởi: AUTH-T01, AUTH-T03.
- Blocks: AUTH-T14, AUTH-T16.
- WBS: P07.5; cơ sở 4h, remaining/actual chưa xác định; không tính lại review như task riêng.
- Agent hỗ trợ triển khai/soạn; Quân review/kiểm tra và cùng chốt brief trước làm.

**Kiểm tra và bằng chứng**

Gói GO cùng các ca trong bảng; phần bảo mật liên quan: AUTH-S08. Theo proposal 6.4.1/9.2.2, chỉ nhận phần bằng chứng đúng tầng của ticket. Bàn giao tên file/lệnh chạy/kết quả/chưa kiểm cho task sau; không ghi trước pass hoặc đưa secret vào Jira.

**Ngoài phạm vi**

Không Redis, dịch vụ limiter mới hoặc hợp nhất ngưỡng hai luồng.

**Tham chiếu**

[Proposal](backend-auth-session.md), mục 6.2.4–6.2.5, WBS P07.5; [FR/AC](../product/functional-requirements.md): Hỗ trợ FR-001.AC1; FR-004.AC4/security; [OpenAPI](../api/auth.openapi.json). Đọc workflow/architecture/security do proposal dẫn khi triển khai, không copy toàn thiết kế vào ticket.

### [AUTH-T14] Triển khai luồng API đăng nhập bằng email công ty

**Mục tiêu**

Hoàn thiện service + handler login, từ request hợp lệ đến cookie sau commit.

**Đầu vào và điều kiện bắt đầu**

Ports/limiter/cookie đã rõ; service dùng fake repository khi phát triển, adapter thật ghép AUTH-T17. Các helper AUTH-T12/AUTH-T13 phải có trước khi nghiệm thu handler; có thể viết service/fake sớm sau AUTH-T10.

**Phạm vi thực hiện**

1. Viết Login dùng bcrypt thật, hash giả khi không có user, nguồn random và FinalizeLogin port.
2. Nối POST /api/auth/login nhận đúng email/password; áp Origin, Content-Type/body thực/limiter trước xác thực phù hợp.
3. Map lỗi/DTO và đặt cookie chỉ khi service thành công; không query DB từ handler.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Credential hợp lệ, cờ false/true | 200 + cookie + principal; cờ true trả quyền hiệu lực rỗng. |
| Sai credential/locked/inactive | 401 INVALID_CREDENTIALS chung; không tiết lộ account tồn tại. |
| JSON/field lạ, Origin, Content-Type | 400/403/415 tương ứng contract. |
| Body >8192 byte | 413 PAYLOAD_TOO_LARGE; không bcrypt/ghi session cho body vượt giới hạn. |
| Chạm limiter | 429 RATE_LIMITED và Retry-After, không tiếp tục xác thực trái giới hạn. |
| Random/nền thay đổi/DB/commit lỗi | Không cookie thành công giả; lỗi đúng contract, không secret trong JSON/log. |

**Tiêu chí hoàn thành**

- [ ] Có unit tests service với fake repo/bcrypt và httptest cho bảng trên, gồm body 8192/8193 byte không Content-Length.
- [ ] Response DTO/no-store/Set-Cookie khớp OpenAPI; cổng Go đạt, chưa dùng fake để nhận DB integration pass.
- [ ] Diff/nội dung đã review; ghi lệnh, kết quả thực chạy và giới hạn. Có đầu ra bàn giao/commit/PR khi có; không nhận test fake/tĩnh thay bằng chứng DB/browser.

**Phụ thuộc / cách thực hiện / estimate**

- Bị chặn bởi: AUTH-T01, AUTH-T10, AUTH-T12, AUTH-T13.
- Blocks: AUTH-T17.
- WBS: P06.2, P07.1; cơ sở 7h, remaining/actual chưa xác định; không tính lại review như task riêng.
- Quân cùng agent triển khai API: lần lượt theo request → handler/middleware → service → repository port → SQL/transaction → response. Trước mỗi checkpoint chốt phần Quân tự viết và test tương ứng; agent hướng dẫn/gỡ rối/review, không viết thay phần đã nhận. Dùng fake để học/test từng tầng, sau đó đối chiếu adapter thật ở AUTH-T17.

**Kiểm tra và bằng chứng**

Gói GO cùng các ca trong bảng; phần bảo mật liên quan: AUTH-S01, AUTH-S04, AUTH-S05, AUTH-S07, AUTH-S08, AUTH-S09, AUTH-S10. Theo proposal 6.4.1/9.2.2, chỉ nhận phần bằng chứng đúng tầng của ticket. Bàn giao tên file/lệnh chạy/kết quả/chưa kiểm cho task sau; không ghi trước pass hoặc đưa secret vào Jira.

**Ngoài phạm vi**

Không cấp tài khoản, SMTP, reset hoặc màn FE; transaction persistence thuộc AUTH-T08.

**Tham chiếu**

[Proposal](backend-auth-session.md), mục 5.3, 6.2.1–6.2.4; OpenAPI login, WBS P06.2, P07.1; [FR/AC](../product/functional-requirements.md): FR-001.AC1/AC4/AC5; CR-002; [OpenAPI](../api/auth.openapi.json). Đọc workflow/architecture/security do proposal dẫn khi triển khai, không copy toàn thiết kế vào ticket.

### [AUTH-T15] Triển khai API danh tính hiện hành và đăng xuất phiên

**Mục tiêu**

Cho FE khôi phục principal và kết thúc đúng phiên qua middleware dùng chung.

**Đầu vào và điều kiện bắt đầu**

Authenticate/Logout service AUTH-T11 và cookie/CSRF AUTH-T12. AUTH-T16 bổ sung cổng phiên hạn chế; AUTH-T15 riêng chưa là toàn feature đủ phát hành.

**Phạm vi thực hiện**

1. Viết middleware phiên, deny-by-default cho route bảo vệ và GET /api/auth/me.
2. Nối POST /api/auth/logout với kiểm Origin/CSRF trước revoke.
3. Map lỗi và no-store, không nhận actor/role/cờ từ client để cấp quyền.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| GET me hợp lệ | 200 DTO hiện hành, không Set-Cookie gia hạn phiên. |
| Thiếu/hết hạn/revoked/inactive | 401; DB lỗi 503/500 thay vì giả hết phiên. |
| Logout đúng nguồn/token | 204 không JSON; revoke đúng phiên rồi xóa cookie đúng thuộc tính. |
| Logout sai nguồn/token hoặc DB lỗi | 403 hoặc 503/500, không revoke/success giả. |
| Route công khai/bảo vệ | Health/login đúng contract; cổng cờ hoàn tất AUTH-T16 trước bàn giao auth. |

**Tiêu chí hoàn thành**

- [ ] httptest đầy đủ ca me/logout, bao gồm gọi lại cookie đã revoke và phiên khác giữ nguyên.
- [ ] Route thử/chi.Walk chứng minh wiring bảo vệ phù hợp, không thêm endpoint nghiệp vụ giả production.
- [ ] Diff/nội dung đã review; ghi lệnh, kết quả thực chạy và giới hạn. Có đầu ra bàn giao/commit/PR khi có; không nhận test fake/tĩnh thay bằng chứng DB/browser.

**Phụ thuộc / cách thực hiện / estimate**

- Bị chặn bởi: AUTH-T01, AUTH-T11, AUTH-T12.
- Blocks: AUTH-T16, AUTH-T17, AUTH-T18.
- WBS: P07.2, P07.4; cơ sở 5h, remaining/actual chưa xác định; không tính lại review như task riêng.
- Quân cùng agent triển khai API: lần lượt theo request → handler/middleware → service → repository port → SQL/transaction → response. Trước mỗi checkpoint chốt phần Quân tự viết và test tương ứng; agent hướng dẫn/gỡ rối/review, không viết thay phần đã nhận. Dùng fake để học/test từng tầng, sau đó đối chiếu adapter thật ở AUTH-T17.

**Kiểm tra và bằng chứng**

Gói GO cùng các ca trong bảng; phần bảo mật liên quan: AUTH-S02, AUTH-S04, AUTH-S05, AUTH-S09, AUTH-S10. Theo proposal 6.4.1/9.2.2, chỉ nhận phần bằng chứng đúng tầng của ticket. Bàn giao tên file/lệnh chạy/kết quả/chưa kiểm cho task sau; không ghi trước pass hoặc đưa secret vào Jira.

**Ngoài phạm vi**

Không API profile hoặc logout tất cả thiết bị.

**Tham chiếu**

[Proposal](backend-auth-session.md), mục 6.2.1–6.2.4; OpenAPI me/logout, WBS P07.2, P07.4; [FR/AC](../product/functional-requirements.md): FR-001.AC2/AC3/AC5; FR-002.AC1–AC2; phần nền FR-003; [OpenAPI](../api/auth.openapi.json). Đọc workflow/architecture/security do proposal dẫn khi triển khai, không copy toàn thiết kế vào ticket.

### [AUTH-T16] Triển khai API đổi mật khẩu lần đầu và cổng phiên hạn chế

**Mục tiêu**

Khép service + handler đổi lần đầu và ngăn phiên hạn chế truy cập nghiệp vụ.

**Đầu vào và điều kiện bắt đầu**

Có contract, policy, middleware, CSRF và limiter; transaction adapter AUTH-T09 ghép ở AUTH-T17. Không nhận fake tests làm bằng chứng transaction; DB/race kiểm tại AUTH-T09/AUTH-T17.

**Phạm vi thực hiện**

1. Viết service kiểm current/new password với bcrypt ngoài transaction và gọi port đổi có kiểm lại nền.
2. Nối POST /api/auth/change-password, áp body/field/Content-Type/Origin/CSRF/limiter theo contract.
3. Tích hợp guard cờ vào login/me/middleware; allowlist phiên hạn chế độc lập role kể cả ADMIN.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Cờ true, input/phiên hợp lệ | 204 sau commit, xóa cookie, không cấp phiên mới. |
| Cờ false hoặc phiên/nền lỗi | 403 PASSWORD_CHANGE_NOT_REQUIRED hoặc 401 đúng nguồn, không ghi. |
| Sai current hoặc sai policy new | 400 riêng theo contract; biên 14/15 code point, 72/73 byte, không trim/cắt. |
| Thiếu/sai CSRF/Origin, oversized/Content-Type, limiter | 403/413/415/429 theo contract, không ghi ngoài điều kiện hợp lệ. |
| Phiên cờ true gọi nghiệp vụ | 403 PASSWORD_CHANGE_REQUIRED kể cả ADMIN; me/csrf/logout/đổi vẫn theo allowlist. |

**Tiêu chí hoàn thành**

- [ ] Unit tests policy/service và httptest chứng minh bảng trên, Unicode/spaces/paste, field lạ/body thực và no-store.
- [ ] Không tự login sau đổi hoặc retry mù; route thử guard chỉ trong test, cổng Go đạt.
- [ ] Diff/nội dung đã review; ghi lệnh, kết quả thực chạy và giới hạn. Có đầu ra bàn giao/commit/PR khi có; không nhận test fake/tĩnh thay bằng chứng DB/browser.

**Phụ thuộc / cách thực hiện / estimate**

- Bị chặn bởi: AUTH-T01, AUTH-T10, AUTH-T12, AUTH-T13, AUTH-T15.
- Blocks: AUTH-T17.
- WBS: P06.4, P07.6; cơ sở 8h, remaining/actual chưa xác định; không tính lại review như task riêng.
- Quân cùng agent triển khai API: lần lượt theo request → handler/middleware → service → repository port → SQL/transaction → response. Trước mỗi checkpoint chốt phần Quân tự viết và test tương ứng; agent hướng dẫn/gỡ rối/review, không viết thay phần đã nhận. Dùng fake để học/test từng tầng, sau đó đối chiếu adapter thật ở AUTH-T17.

**Kiểm tra và bằng chứng**

Gói GO cùng các ca trong bảng; phần bảo mật liên quan: AUTH-S03, AUTH-S04, AUTH-S05, AUTH-S07, AUTH-S08, AUTH-S09, AUTH-S10. Theo proposal 6.4.1/9.2.2, chỉ nhận phần bằng chứng đúng tầng của ticket. Bàn giao tên file/lệnh chạy/kết quả/chưa kiểm cho task sau; không ghi trước pass hoặc đưa secret vào Jira.

**Ngoài phạm vi**

Không reset, đổi tự nguyện, limiter thứ hai hoặc schema mới.

**Tham chiếu**

[Proposal](backend-auth-session.md), mục 5.3, 6.2.2/6.2.5; OpenAPI change-password, WBS P06.4, P07.6; [FR/AC](../product/functional-requirements.md): FR-001.AC4; FR-004.AC1–AC4 phần lần đầu; CR-002; [OpenAPI](../api/auth.openapi.json). Đọc workflow/architecture/security do proposal dẫn khi triển khai, không copy toàn thiết kế vào ticket.

### [AUTH-T17] Tích hợp năm API auth và kiểm thử transaction cạnh tranh

**Mục tiêu**

Ghép config/repo/service/router và chứng minh auth trên HTTP + PostgreSQL test.

**Đầu vào và điều kiện bắt đầu**

Các API/service và persistence đã kiểm riêng; DB test/fixture/barrier sẵn. Có thể ghép/test bốn route nền trước rồi thêm route thứ năm trong cùng ticket; chưa đủ năm route/guard thì AUTH-T17 chưa Done.

**Phạm vi thực hiện**

1. Wire composition root với adapter thật, giữ health không hồi quy.
2. Kiểm năm method/path/DTO/error/header/cookie/guard với OpenAPI và route thực.
3. Chạy luồng thường/hạn chế/đổi/login mới cùng nhiều phiên; kiểm race/rollback/lỗi commit trên hai connection.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Login → me/csrf → logout | Cookie cấp sau commit, quyền/cờ đọc hiện hành, cookie đã revoke không dùng lại. |
| Login hạn chế → đổi → login mới | Mọi phiên cũ đúng user mất hiệu lực, mật khẩu cũ không login được, user khác giữ phiên. |
| Đổi đồng thời/login hash cũ chờ khóa | Không cấp phiên hoặc commit bằng snapshot cũ trái invariant. |
| Lỗi DB/random/commit/mất phản hồi | Không success giả hoặc write nửa chừng; không retry mutation mù. |

**Tiêu chí hoàn thành**

- [ ] Có bằng chứng HTTP + DB thật trong môi trường test sở hữu, barrier/fake clock kiểm race không sleep ngẫu nhiên.
- [ ] Có đối chiếu AUTH-S01–S10 phần backend, giới hạn và đầu ra để FE nối; không cộng lại toàn bộ unit tests.
- [ ] Diff/nội dung đã review; ghi lệnh, kết quả thực chạy và giới hạn. Có đầu ra bàn giao/commit/PR khi có; không nhận test fake/tĩnh thay bằng chứng DB/browser.

**Phụ thuộc / cách thực hiện / estimate**

- Bị chặn bởi: AUTH-T04, AUTH-T08, AUTH-T09, AUTH-T11, AUTH-T14, AUTH-T15, AUTH-T16.
- Blocks: AUTH-T18, AUTH-T19.
- WBS: P09.1, P09.4; cơ sở 7h, remaining/actual chưa xác định; không tính lại review như task riêng.
- Agent hỗ trợ triển khai/soạn; Quân review/kiểm tra và cùng chốt brief trước làm.

**Kiểm tra và bằng chứng**

Gói DB cùng các ca trong bảng; phần bảo mật liên quan: AUTH-S01, AUTH-S02, AUTH-S03, AUTH-S04, AUTH-S05, AUTH-S06, AUTH-S07, AUTH-S08, AUTH-S09, AUTH-S10. Theo proposal 6.4.1/9.2.2, chỉ nhận phần bằng chứng đúng tầng của ticket. Bàn giao tên file/lệnh chạy/kết quả/chưa kiểm cho task sau; không ghi trước pass hoặc đưa secret vào Jira.

**Ngoài phạm vi**

Không API quản trị/dự án hoặc áp migration Neon.

**Tham chiếu**

[Proposal](backend-auth-session.md), mục 5.1, 6.1.6, 6.4.1, WBS P09.1, P09.4; [FR/AC](../product/functional-requirements.md): FR-001/FR-002; phần nền FR-003; phần lần đầu FR-004; CR-002; [OpenAPI](../api/auth.openapi.json). Đọc workflow/architecture/security do proposal dẫn khi triển khai, không copy toàn thiết kế vào ticket.

### [AUTH-T18] Nối FE đăng nhập, khôi phục phiên và đăng xuất với backend

**Mục tiêu**

Dùng auth server làm nguồn danh tính, giữ prototype demo tách biệt.

**Đầu vào và điều kiện bắt đầu**

Có FE phiên mô phỏng; adapter mock và auth thật chưa nối. Đọc web/AGENTS/web/README trước sửa. Đích sau login còn cần xác nhận trước code; chỉ nối live khi API nền đã kiểm. Có thể soạn adapter bằng HTTP giả trước AUTH-T17.

**Phạm vi thực hiện**

1. Viết adapter bốn route nền/proxy dev, DTO/lỗi; không lưu credential/token vào localStorage.
2. Nối login/bootstrap me khi reload và csrf/logout; hiển thị loading/validation/hết phiên/lỗi DB phù hợp.
3. Tách cấu hình demo/auth và điều khiển prototype; không ghép principal thật với employee fixture.
4. Chốt đích sau login và cách báo URL nghiệp vụ chưa nối API trong brief trước code; không tự tạo workspace mới.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Login/reload thường | Principal đến từ server, không selector hoặc role client. |
| Logout/back/reload | Không dùng lại phiên đã revoke; DB lỗi không giả logout thành công. |
| 401/403/429/503 và lỗi mạng | Hành vi/Retry-After đúng contract, không fallback sang demo. |
| Đích nghiệp vụ chưa có API | Ghi rõ giới hạn, không mở dữ liệu mock như quyền/dữ liệu thật. |
| Cờ true | Bàn giao principal/điều hướng hạn chế cho AUTH-T19; không mở workspace trước hoàn thiện AUTH-T19. |

**Tiêu chí hoàn thành**

- [ ] Adapter tests HTTP giả và browser với backend test chứng minh login/reload/logout/error.
- [ ] pnpm check đạt, giữ token/visual convention; cả đợt FE chỉ bàn giao sau AUTH-T19/AUTH-T20.
- [ ] Diff/nội dung đã review; ghi lệnh, kết quả thực chạy và giới hạn. Có đầu ra bàn giao/commit/PR khi có; không nhận test fake/tĩnh thay bằng chứng DB/browser.

**Phụ thuộc / cách thực hiện / estimate**

- Bị chặn bởi: AUTH-T01, AUTH-T15, AUTH-T17.
- Blocks: AUTH-T19, AUTH-T20.
- WBS: P08.1, P08.2, P08.3; cơ sở 10h, remaining/actual chưa xác định; không tính lại review như task riêng.
- Agent hỗ trợ triển khai/soạn; Quân review/kiểm tra và cùng chốt brief trước làm.

**Kiểm tra và bằng chứng**

Gói FE cùng các ca trong bảng. Bổ sung bằng chứng browser với backend test cho phần AUTH-S02 (cookie/phiên khi login, reload và logout), AUTH-S04 (bootstrap CSRF và mutation qua proxy) và AUTH-S09 (không lộ bí mật qua FE) theo proposal 6.4.1/9.2.2. Test adapter bằng HTTP giả chỉ chứng minh xử lý DTO/lỗi phía FE; không thay bằng chứng browser, transaction hoặc kiểm soát quyền ở server. AUTH-T17 cung cấp bằng chứng backend, AUTH-T20 tổng hợp xuyên luồng; ticket này không tự nhận toàn bộ AUTH-S đã pass. HTTPS trước vận hành nếu chưa kiểm phải ghi riêng. Bàn giao tên file/lệnh chạy/kết quả/chưa kiểm; không ghi trước pass hoặc đưa secret vào Jira.

**Ngoài phạm vi**

Không API nghiệp vụ, mapping người thật→fixture hoặc thiết kế lại toàn FE.

**Tham chiếu**

[Proposal](backend-auth-session.md), mục 6.2–6.4, 9.2.2; web/README, WBS P08.1, P08.2, P08.3; [FR/AC](../product/functional-requirements.md): FR-001.AC1/AC2/AC3/AC5; FR-002.AC1–AC2; [OpenAPI](../api/auth.openapi.json). Đọc workflow/architecture/security do proposal dẫn khi triển khai, không copy toàn thiết kế vào ticket.

### [AUTH-T19] Nối FE bắt buộc đổi mật khẩu và đăng nhập lại

**Mục tiêu**

Cho người dùng hoàn tất cổng buộc đổi bằng luồng UI rõ ràng.

**Đầu vào và điều kiện bắt đầu**

Có FE auth AUTH-T18 và backend năm route AUTH-T17. Giữ màn/visual hiện có nếu phù hợp, phần UI mới cần đúng brief đã chốt.

**Phạm vi thực hiện**

1. Dùng cờ từ me/login để điều hướng màn đổi kể cả reload/deep link.
2. Nối form current/new/nhập lại; đối chiếu cục bộ, chỉ gửi current/new và CSRF.
3. Hiển thị giới hạn Unicode/byte/lỗi/Retry-After; xóa dữ liệu nhạy cảm khi kết thúc phiên hoặc thành công.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Input hợp lệ | 204 → xóa principal/CSRF/form, quay login với hướng dẫn; không tự login. |
| Input sai/current sai/429 | Hiện lỗi phù hợp, không khẳng định credential đã đổi. |
| Phiên hết hoặc DB lỗi | Phân biệt 401 và 503, không báo success giả. |
| Mất phản hồi sau POST | Không tự gửi lại; hướng dẫn kiểm bằng login mới vì commit có thể đã xảy ra. |
| Back/reload/URL workspace | Không mở nghiệp vụ bằng phiên cũ hoặc bỏ cổng chỉ từ cờ client. |

**Tiêu chí hoàn thành**

- [ ] Test FE đếm code point/byte đúng server và không gửi ô xác nhận password.
- [ ] Browser luồng hạn chế→đổi→login mới đạt, pnpm check đạt; không lưu/log password.
- [ ] Diff/nội dung đã review; ghi lệnh, kết quả thực chạy và giới hạn. Có đầu ra bàn giao/commit/PR khi có; không nhận test fake/tĩnh thay bằng chứng DB/browser.

**Phụ thuộc / cách thực hiện / estimate**

- Bị chặn bởi: AUTH-T17, AUTH-T18.
- Blocks: AUTH-T20.
- WBS: P08.4; cơ sở 4h, remaining/actual chưa xác định; không tính lại review như task riêng.
- Agent hỗ trợ triển khai/soạn; Quân review/kiểm tra và cùng chốt brief trước làm.

**Kiểm tra và bằng chứng**

Gói FE cùng các ca trong bảng; phần bảo mật liên quan: AUTH-S03, AUTH-S08. Theo proposal 6.4.1/9.2.2, chỉ nhận phần bằng chứng đúng tầng của ticket. Bàn giao tên file/lệnh chạy/kết quả/chưa kiểm cho task sau; không ghi trước pass hoặc đưa secret vào Jira.

**Ngoài phạm vi**

Không đổi tự nguyện, quên/reset mật khẩu hoặc lưu credential để tự đăng nhập.

**Tham chiếu**

[Proposal](backend-auth-session.md), mục 6.2.5, 9.2.2; web/README, WBS P08.4; [FR/AC](../product/functional-requirements.md): FR-001.AC4; FR-004.AC1–AC4 phần lần đầu; [OpenAPI](../api/auth.openapi.json). Đọc workflow/architecture/security do proposal dẫn khi triển khai, không copy toàn thiết kế vào ticket.

### [AUTH-T20] Kiểm tra luồng auth xuyên FE–backend và bàn giao runbook

**Mục tiêu**

Bàn giao luồng chạy lại được cùng bằng chứng đúng phạm vi và điểm còn mở.

**Đầu vào và điều kiện bắt đầu**

Backend AUTH-T17 và FE AUTH-T18/AUTH-T19 đã ghép; runbook có thể soạn từ AUTH-T05. Không chờ AUTH-T20 mới test chức năng; lỗi đã phát hiện phải sửa ở ticket tương ứng hoặc Bug liên kết.

**Phạm vi thực hiện**

1. Kiểm browser luồng thường và hạn chế→đổi→login mới→reload/csrf/logout/back.
2. Rà cookie/proxy/Origin/CSRF/log, tách demo/auth và lỗi phiên/DB trên môi trường test.
3. Hoàn thiện runbook config, DB test/migration/rollback, cách chạy/dọn/demo và giới hạn.
4. Tổng hợp AUTH-S01–S10, ca chưa kiểm và bug mới có liên kết; review diff cuối.

**Hành vi và tình huống cần đạt**

| Tình huống | Kết quả mong đợi |
|---|---|
| Happy path thường/hạn chế | Có bước tái hiện và kết quả xuyên FE–HTTP–DB, không chỉ mock. |
| Lỗi nguồn/phiên/DB và back/reload | Hành vi đúng contract, không cấp quyền hoặc success giả. |
| Bàn giao | Cấu hình/test/demo tái hiện được, không chứa secret; HTTPS trước vận hành nếu chưa kiểm phải ghi riêng. |

**Tiêu chí hoàn thành**

- [ ] Cổng Go/FE/harness và luồng browser bắt buộc có kết quả; ca chưa đạt không được gắn pass.
- [ ] Có runbook/bảng bằng chứng/giới hạn/bước sau, không tuyên bố toàn FR-003/004 hoặc toàn FE đã nối API.
- [ ] Diff/nội dung đã review; ghi lệnh, kết quả thực chạy và giới hạn. Có đầu ra bàn giao/commit/PR khi có; không nhận test fake/tĩnh thay bằng chứng DB/browser.

**Phụ thuộc / cách thực hiện / estimate**

- Bị chặn bởi: AUTH-T18, AUTH-T19.
- Blocks: —; không có ticket phụ thuộc trực tiếp.
- WBS: P09.2, P09.3; cơ sở 6h, remaining/actual chưa xác định; không tính lại review như task riêng.
- Agent hỗ trợ triển khai/soạn; Quân review/kiểm tra và cùng chốt brief trước làm.

**Kiểm tra và bằng chứng**

Gói E2E cùng các ca trong bảng; phần bảo mật liên quan: AUTH-S02, AUTH-S03, AUTH-S04, AUTH-S08, AUTH-S09. Theo proposal 6.4.1/9.2.2, chỉ nhận phần bằng chứng đúng tầng của ticket. Bàn giao tên file/lệnh chạy/kết quả/chưa kiểm cho task sau; không ghi trước pass hoặc đưa secret vào Jira.

**Ngoài phạm vi**

Không tự commit/push/deploy, pentest toàn hệ thống hoặc mở thêm workflow.

**Tham chiếu**

[Proposal](backend-auth-session.md), mục 6.3–6.4, 9.3–9.4, WBS P09.2, P09.3; [FR/AC](../product/functional-requirements.md): FR-001/FR-002; phần nền FR-003; phần lần đầu FR-004; [OpenAPI](../api/auth.openapi.json). Đọc workflow/architecture/security do proposal dẫn khi triển khai, không copy toàn thiết kế vào ticket.

## 5. Cổng trước khi tạo Jira

- Xác minh project key/URL và type Epic/Task/parent/link fields/assignee; mã AUTH-Tnn là nội bộ, không suy thành Jira key hoặc dùng project CoLearnia.
- Chỉ tạo ticket đợt đã duyệt. Không tạo thêm 33 ticket WBS cũ; dùng mapping mục 3 để tránh trùng việc/giờ. Ghi mapping AUTH-Tnn → Jira key sau thành công, không coi file này là đã upload.
- Đổi đường dẫn repo sang URL nguồn đúng branch/commit đã có trên remote; nguồn chưa push thì ghi chờ bàn giao, không bịa link.
- Kiểm lại đầu vào/effort còn lại khi nhận việc. Contract/SQL đã đủ thì ghi bằng chứng thay vì làm lại; không tự ghi Done, hours thực hoặc giảm estimate mà chưa đối soát.
- Đích FE sau login và SQL/mapping/môi trường cụ thể vẫn cần chốt tại ticket tương ứng; không tự chọn từ code/mock.

## 6. Nguồn tham khảo cách tổ chức

- [Atlassian — User stories](https://www.atlassian.com/agile/project-management/user-stories): mô tả mục đích/giá trị và tiêu chí nghiệm thu. DATN dùng Epic cho mục tiêu, Task kỹ thuật cho đầu ra; đây là lựa chọn phù hợp solo, không phải mẫu bắt buộc Atlassian.
- [Atlassian — Definition of Done](https://www.atlassian.com/agile/project-management/definition-of-done): tiêu chí riêng gắn đầu ra; kiểm thử/review là điều kiện hoàn thành, không mặc định thành ticket độc lập.
- [Atlassian Support — Dependencies](https://support.atlassian.com/jira-software-cloud/docs/create-or-remove-dependencies-on-your-timeline/): dùng blocks/is blocked by cho phụ thuộc thực thi; Jira quản lý tiến độ, không thay nguồn nghiệp vụ trong repo.
