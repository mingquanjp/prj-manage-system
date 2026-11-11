# Auth — chi tiết triển khai và bằng chứng theo ticket

Tài liệu tham chiếu của [proposal auth](backend-auth-session.md), tách ngày **12/11/2026** để phần trình bày chính tập trung vào thiết kế hiện hành và năm API.

**Cách đọc:** các mục dưới đây giữ nguyên nội dung bàn giao tại thời điểm từng ticket, gồm chữ ký thiết kế ban đầu, SQL/fixture plan, lệnh và kết quả kiểm tra. Những câu “chưa triển khai”, “task sau sẽ…” chỉ đúng tại mốc ghi nhận; không dùng để kết luận hiện trạng. Bàn giao DATN-18 tại mục 6.2.13 là bằng chứng tích hợp backend tiếp nối các đợt trước. Thiết kế hiện hành đọc ở proposal; chữ ký thực tế đọc ở code được dẫn, HTTP contract đọc ở OpenAPI. Không dùng chữ ký minh họa trước code thay cho interface Go hiện tại.

Giữ số mục cũ để đối chiếu ticket/WBS; đây không phải một proposal mới, nguồn nghiệp vụ mới hoặc xác nhận các ca kiểm thử đã được chạy lại ngày tách tài liệu. Không sửa ngược kết quả kiểm thử cũ thành kết quả hiện tại.

## Mục lục tham chiếu

- [Ports và service/policy](#53-contract-servicerepository--đầu-ra-thiết-kế-p013).
- [SQL, môi trường, migration và repository](#615-kế-hoạch-sql-và-db-test--chuẩn-bị-p04-chưa-thực-thi).
- [Contract, HTTP và tích hợp DATN-18](#626-review-contract-trước-sqltest--06112026).
- [Config/startup](#631-bàn-giao-validation-cấu-hình--datn-3--auth-t02-07112026).
- [Kiểm kê đầu vào WBS tại thời điểm 07/11](#921-trạng-thái-đầu-vào-khi-chuyển-sang-ticket--07112026).

### 5.3. Contract service/repository — đầu ra thiết kế P01.3

Đây là chữ ký thiết kế để review trước task Go, không phải package/interface đã triển khai. Giữ các tầng hiện hữu; chưa tạo package rỗng hoặc thay NewRouter. Transport dùng DTO OpenAPI, không marshal kiểu nội bộ chứa credential/hash. Clock và nguồn random là hai điểm truyền phụ thuộc để kiểm biên/thất bại; không dựng framework DI hoặc generic repository.

```go
// Các kiểu nội bộ chỉ minh họa contract, không phải DTO JSON.
type SessionSecret [32]byte
type SessionHash [32]byte

type Clock interface {
    Now() time.Time
}

type SecretSource interface {
    NewSessionSecret() (SessionSecret, error)
}

type Principal struct {
    UserID int
    EmployeeID int
    EmployeeCode string
    Fullname string
    SystemRoles []string // DTO luôn là array, kể cả [] khi không có quyền hiệu lực.
    MustChangePassword bool
}

type LoginIdentity struct {
    UserID int
    EmployeeID int
    Email string
    PasswordHash string
    UserActive bool
    EmployeeActive bool
}

type SessionSnapshot struct {
    Principal Principal
    CreatedAt time.Time
    ExpiresAt time.Time
    RevokedAt *time.Time
    UserActive bool
    EmployeeActive bool
}

type PasswordChangeBasis struct {
    Session SessionSnapshot
    PasswordHash string
}

type FinalizeLoginCommand struct {
    VerifiedIdentity LoginIdentity
    NewTokenHash SessionHash
    PreviousTokenHash *SessionHash
    TTL time.Duration
}

type ChangeInitialPasswordCommand struct {
    TokenHash SessionHash
    UserID int
    VerifiedPasswordHash string
    NewPasswordHash string
}

type LoginResult struct {
    Principal Principal
    Secret SessionSecret
    ExpiresAt time.Time
}

type AuthRepository interface {
    FindLoginIdentity(context.Context, string) (LoginIdentity, bool, error)
    ReadSession(context.Context, SessionHash) (SessionSnapshot, bool, error)
    ReadPasswordChangeBasis(context.Context, SessionHash) (PasswordChangeBasis, bool, error)
    FinalizeLogin(context.Context, FinalizeLoginCommand) (SessionSnapshot, error)
    ChangeInitialPassword(context.Context, ChangeInitialPasswordCommand) error
    RevokeCurrentSession(context.Context, SessionHash) error
}

type AuthUseCases interface {
    Login(context.Context, string, string, *SessionSecret) (LoginResult, error)
    Authenticate(context.Context, SessionSecret) (Principal, error)
    ChangeInitialPassword(context.Context, SessionSecret, string, string) error
    Logout(context.Context, SessionSecret) error
}
```

Chữ ký trên dùng `context`/`time` của thư viện chuẩn; tên cuối và phạm vi export được xác nhận trong task triển khai. Không mang HTTP status, cookie, pgx/SQL transaction hoặc row driver vào các interface này. Không cần interface cho mọi hàm thuần; bcrypt có thể gọi trực tiếp trong service theo dependency đã có. Không khẳng định Go có thể xóa tuyệt đối mọi bản sao mật khẩu trong bộ nhớ.

| Phép gọi / thành phần | Bất biến bắt buộc và lỗi phân loại |
|---|---|
| FindLoginIdentity | Email lookup đã trim/so không phân biệt hoa thường; found=false khác lỗi DB. Nhiều kết quả là lỗi dữ liệu, không LIMIT 1. Projection nội bộ không trả lên HTTP/log. |
| ReadSession | Cùng statement/snapshot đọc session, user/employee active, cờ và tập role thực trong DB. Service kiểm hiệu lực; DTO cờ true trả roles hiệu lực rỗng. Không gia hạn hoặc tự ghi. |
| ReadPasswordChangeBasis | Chỉ biên service/repo được nhận hash hiện tại; gắn đúng user/session từ cookie, không nhận actor từ body. Service kiểm active/expiry/cờ trước bcrypt. |
| FinalizeLogin | Adapter mở/commit transaction ngắn, khóa user → employee → session cũ cùng user; kiểm lại hash/email/active, đọc cờ/role hiện hành; revoke cookie cũ hợp lệ cùng user rồi insert hash mới. NewTokenHash do server sinh, không client. Clock chung được truyền cho adapter, đọc lại sau chờ khóa để đánh giá phiên cũ/thời điểm cấp mới; TTL=8h. Chỉ trả snapshot sau commit. |
| ChangeInitialPassword | Đối chiếu UserID nội bộ với chủ token, không tin actor từ request. Adapter khóa user → employee → session theo ID, đọc lại active/hash/cờ/expiry/revoke bằng clock sau chờ khóa. So hash đã verify với hash hiện hành. Thay hash + false + revoke mọi phiên cùng transaction; lỗi rollback. Hash/nền/phiên đổi trả lỗi mất xác thực, không ghi đè. |
| RevokeCurrentSession | Tìm đúng user/session từ hash, khóa user → employee → session, đọc lại active/hiệu lực dưới khóa và revoke có điều kiện; không revoke phiên khác, không trả thành công trước commit. |
| Service | Normalize email/kiểm policy, bcrypt verify/hash và random ở ngoài transaction; Authenticate không tự cấp quyền nghiệp vụ khi cờ true. Mutation auth đã recheck trong adapter; mutation nghiệp vụ sau phải tái kiểm tra quyền/trạng thái trong transaction riêng. |
| Transport | Parse cookie/JSON/header, Origin/CSRF/limiter trước bcrypt, DTO allowlist và no-store; cookie chỉ sau commit. Middleware xác thực cho me/csrf/logout/change-password khác guard nghiệp vụ chặn cờ true. Route tương lai không được bỏ guard mặc định. |

`FinalizeLogin`/`ChangeInitialPassword`/`RevokeCurrentSession` là các atomic use-case ports, không CRUD rời để service vô tình ghi nửa chừng; transaction do adapter hiện thực theo contract, service không import pgx. Context cancellation/lỗi commit không được đổi thành “thành công”. Không tự retry thao tác đổi mật khẩu khi kết quả commit không rõ.

Lỗi nội bộ dự kiến phân biệt: InvalidCredentials, Unauthenticated, PasswordChangeRequired, PasswordChangeNotRequired, InvalidCurrentPassword, Validation, RateLimited, Unavailable và Internal. Transport map sang mã tại mục 6.2/OpenAPI; driver error được phân loại/redaction trong adapter, không map mọi lỗi DB thành 401. Tên/type error cụ thể và cách ghi wrap an toàn chốt trong task Go; không log toàn command, LoginIdentity, PasswordChangeBasis hoặc LoginResult.

Clock của service và adapter dùng cùng nguồn được inject, không dùng thời gian client. Kiểm expiry là `now >= expires_at` thì không còn hiệu lực; request chờ khóa vượt hạn phải bị từ chối. Với mutation, adapter gọi `Clock.Now()` sau khi lấy đủ khóa cần thiết, rồi dùng thời điểm này để kiểm expiry và ghi timestamp liên quan: login truyền tường minh `created_at` cùng `expires_at = created_at + 8h`; logout/đổi lần đầu truyền `revoked_at` khi thu hồi phiên. Không dùng thời điểm chụp trước khi chờ khóa hoặc SQL `now()` để thay kiểm thời gian hiện hành sau khóa. Nguồn random lỗi → không ghi/không cookie. Secret chỉ tồn tại ở biên service → transport để đặt cookie, repo chỉ thấy hash SHA-256. Authenticate chỉ trả principal, không hash hoặc secret.

#### 5.3.1. Bàn giao policy phiên — DATN-11 / AUTH-T10, 09/11/2026

[`evaluateSession`](../../internal/service/auth/session_policy.go) nhận `SessionSnapshot` và thời điểm kiểm tra. Service AUTH-T11 sẽ đọc snapshot hiện hành, xử lý riêng lỗi đọc và trường hợp không tìm thấy phiên, rồi truyền `Clock.Now()` vào policy. Policy trả `ErrUnauthenticated` cùng principal rỗng khi phiên hết hạn (kể cả đúng `expires_at`), đã thu hồi, hoặc tài khoản/hồ sơ ngừng hoạt động. Hạn phiên lấy từ snapshot và giữ nguyên qua mỗi lần kiểm.

Principal hợp lệ giữ danh tính và cờ buộc đổi mật khẩu. Quyền hiệu lực là bản sao quyền hệ thống hiện hành. Phiên hạn chế và nhân viên không có quyền đặc biệt nhận slice rỗng. AUTH-T16 sẽ dùng cờ buộc đổi để chặn API nghiệp vụ, kể cả nghiệp vụ chỉ cần tư cách nhân viên. Việc kiểm lại sau khi lấy khóa vẫn thuộc adapter cho từng transaction.

Kiểm chứng trên Go 1.26.5: `go test ./internal/service/auth -run '^TestEvaluateSession' -count=1 -v` đạt 9 ca con, gồm biên hết hạn, thu hồi, từng trạng thái ngừng hoạt động, quyền rỗng/kiêm nhiệm và phiên hạn chế. `go vet ./...`, `go test ./... -count=1`, `go build ./...`, 24 ca kiểm contract/harness và kiểm kê nguồn đều đạt. Đây là bằng chứng tầng Go của AUTH-S02. Kiểm tra wiring service, PostgreSQL, HTTP, cookie và trình duyệt nằm ở các đợt tích hợp tương ứng.

#### 5.3.2. Bàn giao service xác thực và đăng xuất — DATN-12 / AUTH-T11, 10/11/2026

[`auth.NewService`](../../internal/service/auth/session_service.go) nhận repository có `ReadSession` và `RevokeCurrentSession` cùng `Clock`. `Authenticate` và `Logout` nhận `SessionSecret` gồm 32 byte đã giải mã. Cả hai chuyển SHA-256 của các byte này sang repository và truyền tiếp context. Phần giải mã cookie và ánh xạ lỗi HTTP thuộc transport.

`Authenticate` đọc lại snapshot mỗi lần gọi, giữ lỗi từ repository và trả `ErrUnauthenticated` nếu không tìm thấy phiên. Sau khi đọc thành công, service lấy giờ và gọi policy DATN-11 để xác định danh tính và quyền hiệu lực. `Logout` gọi một lần `RevokeCurrentSession` theo hash và chuyển tiếp kết quả. Việc kiểm hiệu lực dưới khóa và commit thuộc adapter. AUTH-T12/AUTH-T15 có thể dùng hai method này cho luồng CSRF, me và logout. Khi ghép thật ở AUTH-T17, service và adapter dùng cùng nguồn giờ.

Kiểm chứng trên Go 1.26.5: `go test ./internal/service/auth -run '^(TestAuthenticate|TestLogout)' -count=1 -v` đạt 4 nhóm test với fake repo/clock. Các ca kiểm hash và context truyền qua port, snapshot thay đổi giữa hai lần đọc, hết hạn trong lúc đọc, lỗi lưu trữ và thu hồi một phiên trong fake. `go vet ./...`, `go test ./... -count=1`, `go build ./...`, 24 ca contract/harness và kiểm kê nguồn đều đạt. Đây là bằng chứng service cho phần đọc quyền/trạng thái hiện hành của AUTH-S05. Việc từ chối field lạ từ client, tích hợp PostgreSQL và hành vi HTTP/cookie cần bằng chứng ở các ticket tương ứng.


#### 6.1.5. Kế hoạch SQL và DB test — chuẩn bị P04, chưa thực thi

**Phạm vi:** hiện thực `user_roles` và `auth_sessions` đã có trong thiết kế, sử dụng cờ 000002; chỉ projection auth trên users/employees. Giữ 000001 và dữ liệu cũ. Không nâng toàn bộ 33 bảng, drop role legacy, đổi username/email/hash hoặc chọn HR/DIRECTOR từ chức danh. Không đọc `.env`, chạy migration/seed/reset, tạo container hoặc truy cập Neon trong đợt lập kế hoạch này.

**Đầu vào thực tế:** kiểm kê Neon 06/11 tại [mapping 5C mục 0](../design/database-step5c-migration-map-2026-10-14.md) là bằng chứng tại thời điểm kiểm kê, không phải xác nhận lại DB hôm nay. Parent khi đó version 1/dirty=false, 000002 chỉ diễn tập nhánh tạm. Có drift tasks so với 000001, nên đường nâng cấp từ migration repo và từ bản kiểm kê thật là hai việc khác nhau. Không dùng `force version`, IF NOT EXISTS hoặc sửa checksum để bỏ qua sai khác.

| Bước | Đầu ra / điều kiện chuyển bước |
|---|---|
| DBP-01 — Preflight trên test | Xác nhận DB riêng đúng host/name/user, major 18, quyền chạy DDL thuộc DB test, version/dirty, users–employees 1–1, email chuẩn hóa trùng, tập legacy role. Kiểm metadata/đếm tổng, không in hash/PII. Chưa có DB test thì chưa chạy. |
| DBP-02 — Đường migration | DB test rỗng → 000001 nguyên trạng → 000002 → migration auth kế tiếp. Đường nâng cấp thứ hai bắt đầu từ fixture legacy version 1 có dữ liệu giả để kiểm bảo toàn. Mã 000003 bên dưới là số dự kiến, kiểm lại migration head trước tạo file. |
| DBP-03 — SQL auth | Một migration tạo user_roles/auth_sessions, constraints/index và mapping được duyệt; runner hiện có quản lý version. Bảo đảm DDL/backfill cùng transaction theo driver thực, diễn tập lỗi và kiểm dirty trước khi dùng. Không chép payload sửa schema_migrations của connector sang file migration Go. |
| DBP-04 — Role legacy | Chuẩn bị bảng mapping tường minh user_id → role_code có nguồn xác nhận. Có thể đề xuất ADMIN cũ → ADMIN, PM/EMPLOYEE không sinh quyền đặc biệt; chưa coi là phê duyệt mapping production. HR/DIRECTOR chỉ cấp khi có xác nhận, không suy từ tên/phòng ban/position. |
| DBP-05 — Kiểm bảo toàn | Số dòng/ID/link user–employee, hash/username/email/status giữ nguyên (so trong test, không xuất giá trị). Role mới đúng mapping, không nhân đôi và không đọc legacy role làm nguồn quyền thứ hai. |
| DBP-06 — Ghép auth | Sau P02/P03 xử lý config/log và các adapter, chạy persistence/concurrency với hai kết nối độc lập, rồi ghép HTTP. Không đánh dấu API quản trị đã có từ transaction giả lập khóa/đổi role. |
| DBP-07 — Áp môi trường thật | Kế hoạch SQL/mapping đã review, kiểm lại target/drift/version, backup/khôi phục và tương thích consumer trước khi xin áp DB đích. Hướng migration đã duyệt không thay duyệt áp SQL cụ thể; không dùng DB test làm bằng chứng production đã nâng. |

DDL dưới đây là **bản soạn để review**, không migration đã tạo hoặc được phép chạy. Bám dictionary 5B mục 2/18; chưa thêm partial index hoặc bảng mới ngoài thiết kế. Constraint CHECK không tự cấp quyền, xác thực hoặc thu hồi phiên.

```sql
CREATE TABLE public.user_roles (
    user_id int NOT NULL REFERENCES public.users (user_id),
    role_code text NOT NULL,
    CONSTRAINT pk_user_roles PRIMARY KEY (user_id, role_code),
    CONSTRAINT ck_user_roles_code CHECK (role_code IN ('ADMIN', 'HR', 'DIRECTOR'))
);

CREATE TABLE public.auth_sessions (
    session_id int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id int NOT NULL REFERENCES public.users (user_id),
    token_hash bytea NOT NULL UNIQUE,
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    CONSTRAINT ck_auth_sessions_hash CHECK (octet_length(token_hash) = 32),
    CONSTRAINT ck_auth_sessions_expiry CHECK (expires_at > created_at),
    CONSTRAINT ck_auth_sessions_revoke CHECK (
        revoked_at IS NULL OR revoked_at >= created_at
    )
);
CREATE INDEX idx_auth_sessions_user ON public.auth_sessions (user_id);
```

Giữ `DEFAULT now()` trong DDL theo thiết kế cấu trúc, nhưng adapter auth phải truyền tường minh các timestamp theo clock sau khóa ở mục 5.3, không dựa default để tính thời điểm cấp/hạn phiên. PostgreSQL `now()` là thời điểm bắt đầu transaction, không tiến theo thời gian chờ khóa; xem [tài liệu thời gian PostgreSQL](https://www.postgresql.org/docs/18/functions-datetime.html#FUNCTIONS-DATETIME-CURRENT). Kiểm DB có barrier phải chứng minh phiên hết hạn trong lúc chờ bị từ chối và session mới có cặp created_at/expires_at từ cùng thời điểm cấp; giữ nhiệm vụ kiểm thực tế ở AUTH-T08/T09/T17. Không sửa default/schema hoặc chạy SQL từ task này.

Cả hai FK mặc định NO ACTION. Không thêm CHECK cứng `expires_at=created_at+8h`: service áp 8h theo A03, bảng giữ ràng buộc cấu trúc đã chốt. Không thêm scope/last_seen/version hoặc quyền PM/PTCM. Email index/ràng buộc chuẩn hóa chưa được duyệt riêng; lookup phải phát hiện nhiều kết quả, không LIMIT 1.

**Tương thích:** tạm giữ users.system_role legacy NOT NULL, nhưng auth mới chỉ dùng user_roles. Seed/consumer cũ có thể chỉ ghi legacy role; không gọi seed chung để cấp quyền hoặc account auth mới. Task tạo tài khoản sau này phải xử lý cột legacy còn giữ hoặc có migration bỏ đã duyệt, không tự sinh hai nguồn quyền. Thêm bảng không tự yêu cầu chuyển enum/status/cấp bậc/email của employees sang toàn bộ thiết kế đích.

**Rollback:** không tự chạy down. Với DB test mới do test sở hữu có thể diễn tập up/down/up theo task được phép; môi trường có phiên/quyền thật ưu tiên dừng consumer và sửa tiến thêm. Drop user_roles làm mất quyền đặc biệt, drop auth_sessions mất toàn bộ phiên; không hứa rollback ứng dụng cũ giữ được quyền mới. 000002 down mất cờ buộc đổi. Chưa có kế hoạch tương thích/khôi phục được duyệt thì chưa áp DB thật.

#### 6.1.6. Thiết kế môi trường và fixture test

- PostgreSQL local riêng major 18 theo A05, dự kiến container do đợt test sở hữu; kiểm Docker/port/tên DB trước dựng, không giả đã có Docker hoặc pin digest ảnh đã kiểm chứng. Chưa viết Compose/cài dependency trong bước tài liệu này.
- Test harness backend phải nhận `TEST_DATABASE_URL` riêng và **không fallback sang DATABASE_URL**. Đặt DB/user test riêng, kiểm host loopback, tên DB allowlist và quyền sở hữu trước mọi DDL/dọn dữ liệu. Host local đơn lẻ chưa đủ chứng minh an toàn (có thể là tunnel); xác nhận chính container/DB do test tạo. Không nạp `.env` thật hoặc in URL khi lỗi.
- `cmd/migrate` hiện đọc DATABASE_URL, không nhận TEST_DATABASE_URL tự động. Runner test phải nối rõ cấu hình test vào driver migration hoặc truyền biến vào subprocess chỉ từ URL test đã xác nhận; không gọi `go run ./cmd/migrate up` theo môi trường shell hiện có. Không vô hiệu RequireDev/guard seed để chạy test xanh.
- Không chạy seed ứng dụng hiện tại. Fixture tối thiểu tự tạo position bắt buộc, user và employee cùng transaction; department có thể NULL. Legacy `system_role` chỉ cung cấp giá trị tương thích cho cột cũ trong fixture, còn quyền auth ghi tường minh vào user_roles. Không suy role mới từ legacy fixture.
- Credential fixture sinh/chuẩn bị ở thời điểm test, hash bcrypt đúng dependency hiện hành; không dùng mật khẩu seed làm chính sách cấp tài khoản, không có mailbox/SMTP thật. Credential dùng demo thủ công chỉ bàn giao qua kênh local được kiểm soát, không in trong CI/Jira hoặc commit vào repo.
- Test có DB riêng/schema được phép theo phiên chạy; chỉ dọn tài nguyên test sở hữu, không truncate/drop diện rộng dựa vào APP_ENV=dev. Sau lỗi không tự force migration version hoặc lấy DB phát triển thay thế.

| Fixture / ca | Mục đích / kết quả cần chứng minh |
|---|---|
| Nhân viên active, cờ false, không role đặc biệt | Login thường, roles=[], không nhầm với phiên hạn chế; user_id khác employee_id |
| Active, cờ true; thêm ca ADMIN + cờ true | Bắt buộc đổi, role không vượt cổng; me/csrf/logout vẫn dùng được |
| HR/DIRECTOR và tài khoản kết hợp role giả | Đọc nhiều role đúng tập, không có PM/PTCM global hoặc quyền suy từ position |
| User LOCKED/DISABLED, employee INACTIVE/RESIGNED, user thiếu employee | Login lỗi chung; phiên đang có bị từ chối khi đọc hiện hành |
| Email khác hoa/thường/khoảng trắng khớp; hai email chuẩn hóa trùng | A02 lookup đúng hoặc fail closed, không tự sửa dữ liệu |
| Session mới, revoked, sát biên 8 giờ, cookie/hash sai | Không gia hạn; tại now=expires_at từ chối; raw secret không lưu DB |
| Hai phiên cùng user và một phiên user khác | Logout đúng một; đổi lần đầu revoke toàn bộ đúng user, giữ user khác |
| Hai connection có barrier và clock điều khiển | Hai lần đổi chỉ một commit; login với hash cũ không cấp phiên sau đổi; chờ khóa vượt expiry bị chặn |
| Lỗi giữa các write/commit và mất phản hồi HTTP | Không để hash/cờ/phiên nửa chừng; không cookie/success giả; kết quả commit không rõ không retry mù |

Không dùng `sleep` ngẫu nhiên để kết luận race đã được kiểm. Fake clock dùng chung service/adapter, barrier điều khiển hai kết nối; rollback fixture/transaction phải có cleanup dù test thất bại. Không cần tạo HTTP CRUD quản trị chỉ để mô phỏng mutation DB trong kiểm thử phối hợp.

**Bằng chứng để P04 hoàn tất:** metadata/fixture plan, mapping được xác nhận cho môi trường cụ thể, quyền sở hữu DB test và cách chạy/dọn, danh sách migration/transaction/rollback, cùng các điểm còn mở. Hiện mới có kế hoạch và DDL soạn; chưa có môi trường/fixture hoặc kết quả migration test. P05/P09 không được đánh dấu Done từ kế hoạch này.

#### 6.1.7. Bàn giao môi trường và kế hoạch dữ liệu — DATN-6 / AUTH-T05, 08/11/2026

Tiếp nối mục 6.1.3–6.1.6: người dùng đã duyệt dựng PostgreSQL 18 local riêng và mapping **dữ liệu giả** dưới đây. Không áp quyền/migration lên Neon. DATN-5 đã merge trên `origin/main` tại `62e0128`; branch DATN-6 tách từ mốc đó. Không sửa 000001/000002 hoặc nâng cả thiết kế 33 bảng.

**Đầu ra vận hành:** [runbook và lệnh chạy/dọn](../../scripts/authdb/README.md), [runner guard](../../scripts/authdb/manage.py), [probe TCP host](../../scripts/authdb/probe.go), [unit tests guard](../../scripts/authdb/test_manage.py). Các nguồn này được harness kiểm kê; credential/owner state chỉ ở `.local/authdb/`, ignore và 0600. Guard từ chối thiếu state, URL khác, engine/container/image/volume/label/mount/port không đúng. Check SQL chỉ đọc metadata, không tạo bảng marker hoặc insert fixture. DB owner là bootstrap superuser của cluster test riêng, không mô hình tài khoản production.

**SQL/mapping được xác nhận cho môi trường test này:** dùng DDL hai bảng đúng mục 6.1.5, cùng FK NO ACTION/check/index. Migration tiếp theo dự kiến `000003` (head hiện là 000002; DATN-7 kiểm lại trước tạo). Không dùng `IF NOT EXISTS`, force version, sửa migration cũ hoặc copy SQL sửa schema_migrations từ diễn tập connector. Auth mới chỉ đọc user_roles; legacy system_role còn giữ để consumer cũ hoạt động, không cấp quyền từ cả hai nguồn.

| Nguồn legacy giả | Đích user_roles | Kiểm phải đạt |
|---|---|---|
| ADMIN | ADMIN đúng user_id | Một dòng mỗi user, không suy quyền HR/DIRECTOR |
| PM, EMPLOYEE | Không thêm dòng quyền đặc biệt | Không biến PM/PTCM thành role toàn hệ thống |
| HR/DIRECTOR/multiple-role fixture mới | Insert tường minh sau khi bảng auth tồn tại | Không suy từ tên/position/department; không nhét giá trị mới vào enum legacy |
| Giá trị role ngoài enum/tập đã kiểm kê, liên kết mâu thuẫn | Preflight dừng | Không tự gán EMPLOYEE hoặc sửa dữ liệu để vượt lỗi |

Backfill SQL cụ thể để DATN-7 đưa vào transaction migration sau review:

```sql
INSERT INTO public.user_roles (user_id, role_code)
SELECT user_id, 'ADMIN'
FROM public.users
WHERE system_role = 'ADMIN';
```

Không có file mapping ID sản xuất trong repo. Mapping ADMIN trên dữ liệu thật vẫn cần kiểm kê/xác nhận đích riêng; các tài khoản HR/DIRECTOR thật chưa có mapping được duyệt. Không gán quyền mới để khớp UI demo. Auth_sessions là bảng mới, không tái tạo phiên lịch sử hoặc token từ seed.

**Prerequisite/preflight cho DATN-7:**

1. Guard sở hữu và probe PG18 phải đạt trước DDL. DB rỗng chưa có schema_migrations thì báo trạng thái chưa khởi tạo, không SELECT một bảng chưa tồn tại. DB có lịch sử phải có đúng một version, dirty=false; version ngoài đường dự kiến dừng.
2. Đường legacy kiểm enum/kiểu/PK/unique/FK cần cho users/employees/positions; positions là FK NOT NULL của employee, departments nullable. Fixture tạo position trước employee, giữ user_id khác employee_id để bắt lỗi copy ID. Không nâng profile/address/job_levels hoặc các bảng nghiệp vụ khác trong auth.
3. Kiểm users thiếu employee, employee thiếu user, user link trùng (kể cả schema sai đã gỡ unique), role ngoài tập, email chuẩn hóa trùng. Tất cả output là count/metadata, không xuất tên/email/hash. Bộ legacy sạch kỳ vọng counts mâu thuẫn bằng 0; bộ fixture bất thường chủ động dành cho lookup fail closed, tách DB/case, không dùng để chứng minh upgrade sạch.
4. Kiểm cột must_change_password đúng bool/NN/default false sau 000002, dữ liệu cũ false; từ version 2 có cờ true thì phải bảo toàn qua migration auth. Không có temporary_password_expires_at. Kiểm pre-existing user_roles/auth_sessions hoặc drift liên quan thì dừng thay vì bỏ qua bằng IF NOT EXISTS.
5. Kiểm số dòng/ID/user–employee link và so trong test username/hash/email/status/timestamp trước–sau; không in giá trị/hash-checksum từ dữ liệu nhạy cảm. Schema drift tasks của Neon đã ghi tại mapping 5C không nằm trong auth projection; không giả local rehearsal bằng chứng đã xử lý toàn drift Neon.

**Các đường chạy được bàn giao; chưa thực thi ở DATN-6:**

| Mã | Thiết lập / đường nâng | Kết quả để DATN-7 nghiệm thu |
|---|---|---|
| MIG-01 | DB test rỗng → runner thực áp 000001 nguyên trạng → 000002 → migration auth | Head/dirty đúng, hai bảng/PK/FK/unique/check/index đủ; không có fixture trước prerequisite |
| MIG-02 | Áp 000001 → nạp legacy giả có ADMIN/PM/EMPLOYEE → 000002 → auth | Cột false, dữ liệu giữ nguyên; chỉ ADMIN sinh quyền ADMIN; user_id không bị nhầm employee_id |
| MIG-03 | DB version 2 giả, tài khoản true/false → auth | Giữ đúng cờ/hash/ID và liên kết; không đặt lại tất cả false |
| MIG-04 | Lỗi backfill/DDL cố ý ở bản sao test | Kiểm atomic DDL/backfill qua driver thực và trạng thái dirty; không nhận partial schema là thành công, không force version |
| MIG-05 | Bản test hy sinh đã nâng auth → down auth → up auth | Kiểm đường rollback và giới hạn dữ liệu; bảng quyền/phiên mới bị mất khi down, không hứa giữ các role cấp sau backfill |
| MIG-06 | Cần diễn tập down 000002: bản test hy sinh, consumer dừng | Mất cờ là có chủ đích; up lại chỉ default false, không phục hồi trạng thái true; không chạy trên DB có dữ liệu cần giữ |

Đã đọc adapter pgx/v5 của golang-migrate v4.19.1 cài local: Run với MultiStatementEnabled=false đọc toàn file và gọi ExecContext một lần; SetVersion quản lý version/dirty bằng transaction riêng. Kế hoạch DATN-7 giữ MultiStatementEnabled=false, SQL auth gồm BEGIN; SET LOCAL lock_timeout/statement_timeout; DDL hai bảng/index; backfill ADMIN; COMMIT trong cùng file. Không đưa cập nhật schema_migrations vào SQL nghiệp vụ. DATN-7 phải kiểm atomic DDL/backfill và trạng thái dirty qua driver thật, gồm lỗi giữa file; kiểm source adapter không thay bằng chứng DB. Test rollback không dùng để biện minh chạy down production. Bản app cũ không đọc role mới hoặc buộc đổi mật khẩu; rollback app không bảo toàn chính sách auth mới. Đích thật ưu tiên sửa tiến thêm sau backup/đánh giá consumer, xin áp riêng.

**Fixture plan sau prerequisite (không nạp trong DATN-6):**

| Nhóm | Dữ liệu giả / nơi dùng |
|---|---|
| F01 — legacy sạch | Position bắt buộc, department NULL và non-NULL, ADMIN/PM/EMPLOYEE, ID user khác employee; DATN-7 kiểm bảo toàn |
| F02 — cờ | Active false, active true, ADMIN+true; DATN-7 cấu trúc/bảo toàn, DATN-8 lookup, DATN-10 transaction đổi lần đầu |
| F03 — quyền hiện hành | roles rỗng, ADMIN/HR/DIRECTOR, kết hợp; quyền chỉ ghi user_roles sau migration; không lấy position làm role |
| F04 — trạng thái/liên kết | User LOCKED/DISABLED, employee INACTIVE/RESIGNED, user thiếu employee; không phá FK thật để tạo employee orphan |
| F05 — email | Stored case/space và query chuẩn hóa; hai chuỗi khác raw nhưng cùng lower(btrim(email)); giữ unique email raw, không tự thêm expression index |
| F06 — phiên | Hash 32 byte giả, created/expires/revoked hợp lệ; biên expires_at=now; hai phiên user A và một user B. Ca hash sai độ dài/expiry sai phải bị CHECK từ chối |
| F07 — khóa/race | Hai connection riêng + barrier/clock chung; đổi cạnh tranh, login hash cũ chờ khóa, chờ vượt expiry; thực thi ở DATN-9/10/18, không sleep đoán race |
| F08 — failure | Chèn lỗi từng write/commit theo hook của test; kiểm rollback hash/cờ/revoke và không báo success giả; không tự coi commit-unknown là rollback |

Credential/hash fixture tạo lúc test bằng bcrypt hiện hành, không mailbox/SMTP thật. Timestamp rõ bằng clock test, không lấy now() transaction-start thay clock sau khóa (mục 5.3). Từng case DB fixture độc lập hoặc cleanup có guard; không seed chung, không dọn theo APP_ENV. DATN-7 triển khai fixture cấu trúc trước, DATN-8/9/10 mở rộng đúng ca dùng; không phải tất cả concurrency tests đều thuộc DATN-7.

**Bằng chứng thực chạy DATN-6, 08/11/2026:**

- Docker Desktop local qua Unix socket đã kiểm hoạt động; tạo mới bằng image chính thức postgres:18 (image ID ghi trong state local, không pin tag patch cho mọi máy). Container/volume/port/name/user theo runbook. Image/mount PG18 đối chiếu [nguồn chính thức](https://github.com/docker-library/docs/blob/master/postgres/README.md).
- `manage.py up` và `check`: psql TCP/password trong container + pgx TCP từ host trả cùng cluster system ID; major=18, DB/user đúng, owner=true, public_tables=0. Không xuất credential/hash/email; chưa có schema_migrations/bảng auth.
- `run -- python3 -c ...`: subprocess nhận URL test sở hữu và không có DATABASE_URL/APP_ENV. Đây là kiểm plumbing môi trường, không là migration test.
- `stop → up`: giữ nguyên cluster và không reset. `destroy --confirm ai-staffing-auth-test → up`: chỉ xóa/tạo lại container/volume test, cluster ID mới, public_tables=0. Container mới được để chạy cho DATN-7; credential/state cũ không tái sử dụng. Chưa thử phục hồi sau crash ở giữa thao tác lifecycle.
- Unit guard offline: 8/8 đạt (Docker giả), gồm thiếu state, URL khác, engine/container/image/volume/mount/port sai, secret redaction. Live runner cũng đã từ chối URL khác và destroy thiếu xác nhận; check sau đó DB vẫn trống. Chưa thay engine hoặc phá container thật để diễn tập các ca ownership sai.
- `go vet ./...`, `go test ./... -count=1`, `go build ./...`: đạt; probe mới compile trong gói scripts/authdb, không có test Go riêng. Auth runtime chưa có.
- Contract/harness cuối 24/24 đạt; kiểm kê 226 nguồn đạt; diff --check đạt. Test tĩnh không thay bằng chứng DB.
- Sub-agent GPT-6.1 Sol medium review chỉ đọc theo context bàn giao, không truy cập Docker/DB/secret: không phát hiện lỗi chặn; đã sửa ghi nhận lịch sử trong mục lục và thêm cảnh báo không gọi probe ngoài guard. Tác giả đối chiếu diff và kết luận; không phải phê duyệt độc lập của người dùng.

Không chạy migrations/seed/fixture/auth API/race hoặc truy cập Neon. Không đổi Jira Done, commit hoặc push. Đầu vào DATN-7: môi trường trống có guard, DDL/mapping test/transaction plan và ma trận MIG/F ở trên; trước viết migration vẫn kiểm head/source/đích một lần nữa.

#### 6.1.8. Migration auth và bảo toàn dữ liệu — DATN-7 / AUTH-T06, 08/11/2026

Người dùng đã duyệt triển khai/thực thi trên cluster PostgreSQL test DATN-6, không Neon. Branch `feature/DATN-7-auth-migration` từ origin/main `9720b87` (DATN-6 đã merge), head kiểm trước sửa là 000002. Thêm [000003 up](../../migrations/000003_auth_roles_sessions.up.sql)/[down](../../migrations/000003_auth_roles_sessions.down.sql) đúng DDL/backfill tại 6.1.5/6.1.7; giữ nguyên 000001/000002, không thay schema ngoài hai bảng auth, không tạo HR/DIRECTOR từ dữ liệu legacy.

**Cầu nối và cô lập:** [integration tests](../../migrations/auth_integration_test.go) dùng build tag integration, pgx + golang-migrate hiện hành; lệnh và guard ở [runbook](../../scripts/authdb/README.md#migration-auth-integration--datn-7--auth-t06). Runner truyền TEST_DATABASE_URL + AUTH_TEST_DB_SYSTEM_ID từ cluster đã kiểm. Go xác minh lại đích, tạo database case tên ngẫu nhiên trong đúng cluster, không nhận DB có sẵn; driver kiểm current_database trước DDL. Cleanup chỉ DROP DB vừa CREATE thành công sau khi đóng connection và kiểm cluster/owner; không force/prune hoặc dọn rộng. DB gốc giữ trống. Không chạy đồng thời hai lượt suite trên cùng cluster.

**Đường dữ liệu đã hiện thực:** DB rỗng → 1 → 2 → 3; version 1 + legacy giả → 2 → 3; version 2 + cờ true/false → 3; auth down/up trên bản test; lỗi SQL test trong bộ nhớ sau DDL/backfill, kiểm rollback dữ liệu/bảng + version 3 dirty=true và từ chối retry. Không force version hoặc sửa file migration để làm test xanh. Down auth mất role/phiên bổ sung, up chỉ backfill lại ADMIN; không gọi đây là khôi phục đầy đủ. MIG-06 (down 000002) không chạy, không cần cho AC nền auth này.

**Preflight và fixture:** kiểm type/enum/PK/unique/FK của projection legacy, số liên kết thiếu/trùng, email chuẩn hóa trùng, tập quyền cũ, auth table chưa có và cờ đúng version. Bộ âm gồm user thiếu employee, email case/space trùng, mất unique link, auth table có sẵn, type username sai và enum legacy khác; tất cả phải dừng trước nâng version. Fixture dùng position trước employee, department nullable/non-null, ID user khác employee, bcrypt sinh khi test, role/status khác nhau. Snapshot toàn bộ hàng users/employees/positions/departments so trong bộ nhớ; sau 000002 bỏ đúng cột mới khi so bản version 1, sau auth so toàn bộ gồm cờ. Không xuất dữ liệu/hash/email hoặc token.

**Giới hạn áp dụng:** preflight/fixture là rehearsal local, không chứng nhận Neon hay thay công cụ preflight dữ liệu thật. File 000003 nay được embed vào cmd/migrate, nhưng chưa được phép áp DB đích thật; mapping/backup/drift/consumer và xác nhận áp đích vẫn là cổng riêng. SQL PK/FK/CHECK không thực thi auth runtime/thu hồi phiên/8h hoặc sửa chính sách phiên.

**Bằng chứng thực chạy DATN-7, 08/11/2026:**

- `python3 scripts/authdb/manage.py run -- go test -tags=integration ./migrations -count=1 -v -timeout=4m`: đạt trên PostgreSQL 18 local. Sáu nhóm test gồm đường rỗng, legacy (role/session/metadata), bảo toàn cờ + down/up, lỗi atomic/dirty, sáu ca drift preflight và kiểm DB gốc/cleanup. Không fake driver/DB ở bộ này.
- MIG-01/02/03/04/05 có bằng chứng DB thật như mô tả trên; MIG-06 (down 000002), race/API/login/CSRF/đổi mật khẩu chưa chạy. Bộ constraint xác minh SQLSTATE PK/FK/unique/NN/check và metadata index/default/identity; FK runtime được cô lập khỏi FK employee cũ.
- Regression từng gặp khi viết test: cfg.ConnString() giữ URL gốc sau khi đổi cfg.Database. Đã chuyển sang stdlib.OpenDB(*cfg), kiểm current_database trước DDL và thêm ca DB gốc trống/không còn auth_case; đã tạo lại đúng cluster test sở hữu để loại dữ liệu từ lượt thử sai đích. Không ảnh hưởng Neon/DB ứng dụng. Kết quả cuối: DB gốc 0 bảng và không còn DB case.
- `go vet ./...`, `go test ./... -count=1`, `go build ./...`, `go vet -tags=integration ./migrations`: đạt. Unit tests mặc định không chạy DB integration; command riêng ở trên mới là bằng chứng DB.
- Unit guard Python 8/8; contract/harness 24/24; kiểm kê 229 nguồn và diff --check đạt. Không gọi kiểm tĩnh là runtime auth pass.
- Sol medium review chỉ đọc: góp ý thiếu preflight và FK assertion bị FK legacy che đã xử lý; lượt cuối không còn finding chặn. Metadata PK/FK legacy kiểm cấu trúc cần cho rehearsal từ repo, chưa đối chiếu mọi biểu thức/cột/action để làm validator drift production tổng quát. Tác giả đối chiếu kết luận và test DB, không giả review agent là người phê duyệt độc lập.

**Bàn giao DATN-8 / AUTH-T07:** hai bảng auth/mapping có migration và bằng chứng local; dùng cùng runner/cluster và dựng fixture riêng cho repository tests, không tái dùng DB case đã dọn. Cột must_change_password có sẵn từ 000002; user_roles là nguồn quyền auth, không đọc system_role làm nguồn quyền thứ hai. Lookup email chuẩn hóa trùng vẫn phải fail closed ở DATN-8, không thêm unique expression index ngoài quyết định. Trước áp DB thật cần kiểm kê/mapping/backup/consumer và xác nhận riêng. Chưa commit/push hoặc đổi Jira tự động trong lượt triển khai này.

#### 6.1.9. Truy vấn định danh và quyền hiện hành — DATN-8 / AUTH-T07, 08/11/2026

[Adapter PostgreSQL](../../internal/repository/postgres/auth_identity.go) hiện thực hai phương thức của [IdentityReader](../../internal/service/auth/identity.go). `FindLoginIdentity` giữ email đã lưu để DATN-9 có thể đối chiếu sau khi khóa. Truy vấn kiểm tối đa hai dòng khớp `lower(btrim(email))` trước khi xét trạng thái. Hai dòng khớp trả `ErrIdentityConflict`, kể cả khi một tài khoản đang khóa. Cờ buộc đổi thuộc snapshot phiên và sẽ được đọc lại khi hoàn tất login dưới khóa.

`ReadSession` đọc theo token hash bằng một statement. Kết quả gồm thông tin phiên, danh tính tối thiểu, trạng thái user/employee, cờ và role từ `user_roles`. Hết hạn, thu hồi, inactive và giới hạn quyền của phiên là quyết định ở service từ dữ kiện đã đọc. Khi không có hash khớp hoặc thiếu employee thì `found=false, err=nil`. Lỗi driver được chuyển sang loại lỗi nội bộ an toàn, tách khỏi không tìm thấy và dữ liệu mâu thuẫn.

[Fixture dùng chung](../../internal/testutil/authdb/fixture.go) dựng DB case mới, nâng tới 000003, rồi thêm tài khoản, hồ sơ, role và phiên giả theo từng ca. Bộ migration DATN-7 và bộ truy vấn dùng cùng [guard DB case](../../internal/testutil/authdb/database.go). Xem [lệnh integration và cách dùng fixture](../../scripts/authdb/README.md#truy-vấn-auth-integration--datn-8--auth-t07).

**Bằng chứng đã chạy:** PostgreSQL 18 test local đạt 4 nhóm integration cho repository và 6 nhóm migration hồi quy. Đã kiểm email hoa/thường và khoảng trắng, trùng với tài khoản khóa, trạng thái hai nguồn, hai ID khác nhau, role kết hợp/rỗng, đọc lại sau đổi role/cờ/trạng thái, phiên thu hồi, phiên thiếu employee, lỗi SQL thật và dọn DB case. `go vet ./...`, `go test ./... -count=1`, `go build ./...` và vet với tag integration đều đạt. Đây là phần persistence của AUTH-S01/AUTH-S05. Ánh xạ HTTP, hash giả khi thiếu tài khoản, hiệu lực phiên và kiểm cạnh tranh chờ các task tương ứng.

**Bàn giao DATN-9 / AUTH-T08:** dùng `LoginIdentity` làm đầu vào đã xác minh và so lại email/hash/trạng thái sau khóa theo mục 5.3. Dùng `SessionSnapshot` để đọc trạng thái, fixture để dựng nhiều phiên cho cùng user và cho user khác. DATN-10 bổ sung `ReadPasswordChangeBasis` cùng giao dịch đổi mật khẩu lần đầu.

#### 6.1.10. Transaction tạo và thu hồi phiên — DATN-9 / AUTH-T08, 09/11/2026

[SessionWriter](../../internal/service/auth/identity.go) cung cấp `FinalizeLogin` và `RevokeCurrentSession`. [Adapter PostgreSQL](../../internal/repository/postgres/auth_sessions.go) mở và kết thúc transaction cho từng port theo mục 5.3. Constructor `NewAuthRepository(db, clock)` nhận nguồn giờ tường minh, dùng được với pool hoặc connection; mọi caller hiện có đã truyền clock. Composition root về sau phải truyền cùng nguồn giờ cho service và adapter.

`FinalizeLogin` khóa user, employee rồi phiên cũ khớp hash và cùng user. Adapter so lại ID liên kết, email đã lưu, password hash và trạng thái hoạt động, lấy cờ/role hiện hành dưới giao thức khóa đã chốt. Sau đủ khóa, một lần đọc clock quyết định hiệu lực phiên cũ và cặp thời điểm cấp/hạn phiên mới. TTL khác 8 giờ là đầu vào nội bộ sai, trả lỗi trước khi ghi. Phiên cũ cùng user còn hiệu lực được thu hồi cùng transaction tạo phiên mới. Phiên không tồn tại, thuộc user khác, hết hạn hoặc đã thu hồi được bỏ qua. Snapshot trả về dùng timestamp thực được PostgreSQL lưu qua `RETURNING` và chỉ được trả sau commit thành công.

`RevokeCurrentSession` tìm chủ phiên từ hash rồi khóa user, employee và đúng session. Adapter đọc lại active/hạn/revoke dưới khóa, lấy giờ sau khi chờ đủ khóa và chỉ thu hồi phiên hợp lệ đang gọi. Phiên mất hiệu lực trả `ErrUnauthenticated`, nền login thay đổi trả `ErrInvalidCredentials`. Lỗi DB đi qua bộ phân loại an toàn đã dùng cho truy vấn DATN-8. Cleanup rollback có context hữu hạn riêng để vẫn thực hiện khi request bị hủy.

Bộ kiểm tra bổ sung sáu nhóm transaction integration gồm cấp phiên, nền thay đổi, logout, rollback, cạnh tranh và kết quả commit không rõ. Trong đó 13 ca cạnh tranh dùng hai connection riêng, barrier `pg_blocking_pids` và clock điều khiển; chỉ dịch clock/nhả khóa sau khi PostgreSQL xác nhận đang bị chặn. Ca reset hash giả lập giữ khóa user rồi đổi hash/thu hồi phiên. Các ca giữ riêng khóa employee/session kiểm việc chờ tại từng tầng khóa, không chứng nhận giao thức của API quản trị chưa có. Ca thiếu employee kiểm trực tiếp bảng session để tránh JOIN che dữ liệu ngoài ý muốn.

Ca unique hash gây lỗi insert sau khi thu hồi phiên cũ kiểm rollback cả hai write. Trigger gây lỗi update và constraint trigger trì hoãn tới commit chỉ tồn tại trong DB case; kiểm dữ liệu sau lỗi để xác nhận PostgreSQL rollback. Ca mất phản hồi commit dùng wrapper quanh transaction thật, gọi commit DB trước rồi trả lỗi I/O giả; adapter trả lỗi/snapshot rỗng trong khi dữ liệu đã commit vẫn quan sát được. Đây là mô phỏng biên driver, chưa phải thử cắt mạng thực. Vì vậy không khẳng định mọi lỗi commit đều rollback hoặc retry là an toàn.

DATN-10 / AUTH-T09 và DATN-18 / AUTH-T17 tái sử dụng thứ tự khóa, clock, phân loại lỗi và [fixture/connection test đã xác minh](../../internal/testutil/authdb/database.go). DATN-10 bổ sung transaction đổi lần đầu theo mục 5.3, giữ bcrypt ngoài khóa và khóa các session theo ID. DATN-18 kiểm tích hợp các port với service/HTTP. Bằng chứng DATN-9 chỉ là phần adapter/PostgreSQL của AUTH-S02/AUTH-S07. Cookie, CSRF, limiter, phản hồi HTTP và API khóa/reset chưa được kiểm chứng ở đợt này. Schema và migration giữ nguyên.

Lệnh tái hiện, kiểm cuối DB gốc/case và các cổng Go/harness nằm trong [runbook DATN-9](../../scripts/authdb/README.md#transaction-auth-integration--datn-9--auth-t08). Chỉ dùng PostgreSQL 18 test local sở hữu theo guard DATN-6, không Neon hoặc DB ứng dụng.

#### 6.1.11. Transaction đổi mật khẩu lần đầu — DATN-10 / AUTH-T09, 09/11/2026

[`ReadPasswordChangeBasis`](../../internal/repository/postgres/auth_identity.go) đọc theo hash token bằng một statement. Kết quả gồm snapshot phiên và hash mật khẩu hiện tại dành cho service. Truy vấn `ReadSession` thông thường chỉ đọc snapshot. Service sẽ kiểm hiệu lực, cờ buộc đổi và xác minh mật khẩu hiện tại, sau đó tạo hash mới trước khi gọi repository.

[`ChangeInitialPassword`](../../internal/repository/postgres/auth_password_change.go) nhận ID user nội bộ, hash token, hash đã xác minh và hash mới. Transaction khóa user → employee → các phiên của user theo `session_id` tăng dần. Adapter kiểm trạng thái và nền dưới khóa, đối chiếu token với đúng user, rồi dùng clock sau khi lấy đủ khóa để kiểm phiên còn hiệu lực. Nền đã đổi, token sai chủ hoặc phiên hết hạn/đã thu hồi trả `ErrUnauthenticated`. Trường hợp cờ vốn đã false ở lần đọc đầu thuộc kiểm tra của service theo mục 5.3.

Một commit lưu hash mới, đặt cờ false và thu hồi mọi phiên chưa thu hồi của user, kể cả phiên đã hết hạn. Timestamp của phiên đã thu hồi và dữ liệu của user khác được giữ. Đổi mật khẩu không cấp phiên mới. Khi lỗi SQL, adapter trả lỗi lưu trữ đã phân loại và dọn transaction. Nếu PostgreSQL đã commit nhưng driver mất phản hồi, adapter trả lỗi và không retry.

Các ca PostgreSQL bao phủ snapshot theo token, dữ liệu giữ nguyên khi từ chối, rollback khi lỗi ghi user/phiên hoặc lỗi ở commit, và ba ca chờ khóa. Hai request đổi dùng hai phiên cùng user, chỉ request đầu commit. Login đã đọc hash cũ bị từ chối sau khi transaction đổi commit; login từ nền mới có thể cấp phiên với cờ false. Ca hết hạn giữ khóa ở phiên có ID lớn hơn phiên gọi và tiến clock đúng tới hạn sau khi PostgreSQL xác nhận request đang chờ. Ca mất phản hồi dùng wrapper trả lỗi I/O sau commit thật.

Đã chạy suite migration và repository với `-race -tags=integration -p 1` qua runner PostgreSQL 18 test local. Cả hai package đạt. Ca kiểm chạy riêng sau suite xác nhận DB gốc vẫn trống và không còn DB case. `go vet ./...`, `go test ./... -count=1`, `go build ./...` và vet với tag integration đều đạt.

Bàn giao DATN-18 / AUTH-T17 sử dụng hai port mới theo mục 5.3. Bằng chứng thuộc tầng PostgreSQL cho AUTH-S06 và phần lỗi transaction của AUTH-S07. Phần xác minh bcrypt của service, HTTP, CSRF/cookie, chính sách đăng nhập lại và thử mất mạng thực cần kiểm riêng. Lệnh tái hiện và kiểm cleanup nằm trong [runbook DATN-10](../../scripts/authdb/README.md#đổi-mật-khẩu-lần-đầu-integration--datn-10--auth-t09).


#### 6.2.6. Review contract trước SQL/test — 06/11/2026

Đối chiếu OpenAPI, ports mục 5.3 và A01–A09: năm route có cùng DTO/cờ/lỗi, mutation yêu cầu cookie **và** CSRF trong cùng Security Requirement, Origin riêng; me/csrf dùng được khi cờ true. Roles=[] không đủ xác định phiên hạn chế, FE/server phải dùng must_change_password. Đổi lần đầu 204 không tự cấp session mới; sai current password là 400, DB lỗi không giả 401/204. Repo atomic ports kiểm lại nền/expiry sau chờ khóa, không chạy bcrypt trong transaction.

Kiểm tĩnh hiện có kiểm ref nội bộ, allowlist, cấu trúc security, DTO và no-store. Đây không phải validator đầy đủ theo metaschema OpenAPI 3.1 hoặc test handler. Constraint 72 byte là extension cần server kiểm; minLength của JSON Schema chỉ kiểm code point. Header cookie/Origin/CSRF và transaction/concurrency chỉ được chứng minh khi có runtime.

**Chốt bổ sung trước nhận task HTTP/config:** body JSON login/đổi lần đầu ≤8192 byte, vượt trả 413, đã duyệt riêng cùng ngày; không thay giới hạn mật khẩu. Kiểm thử HTTP phải gồm 8192/8193 byte, body không khai báo độ dài/Content-Length không đáng tin và không có tác dụng phụ khi chặn. P04 còn review mapping/phạm vi SQL cụ thể và dựng DB test; estimate 105h/21h dự phòng không tự thành thời gian đã làm hoặc ngày hoàn thành.

#### 6.2.7. Bàn giao contract DATN-2 / AUTH-T01 — 07/11/2026

Đối chiếu theo FR-001/002, phần nền FR-003, phần lần đầu FR-004 và A01–A09/body 8 KiB; không bổ sung quyết định nghiệp vụ hoặc schema. OpenAPI là nguồn HTTP duy nhất, mục 5.3 là port nội bộ. Bản này sẵn sàng để người dùng review; không tự chuyển Jira Done hoặc tuyên bố đã có auth runtime.

| Điểm đối chiếu | Kết luận / đầu vào cho task sau |
|---|---|
| Năm route | Giữ đúng method/path ở mục 6.2; không reset/refresh/cấp tài khoản. Bổ sung 413/415 vào bảng tóm tắt để khớp OpenAPI. |
| Input và lỗi | Login/đổi lần đầu nhận JSON theo allowlist và body 8192 byte. Logout body rỗng, không bắt Content-Type; body không rỗng là 400, không tự áp lỗi 415 của hai route JSON. |
| Danh tính và phiên hạn chế | Login/me dùng cùng Principal; system_roles luôn array (Go nil slice không được marshal thành null). Cờ true → [] kể cả ADMIN; cờ false + [] vẫn là nhân viên hợp lệ. Thêm ví dụ hai nhánh và lỗi chung trong OpenAPI cho FE/httptest đối chiếu. |
| Cookie/CSRF | Cookie chỉ cấp sau commit login; me/csrf không gia hạn. Mutation cần cookie AND CSRF cùng Origin; login chỉ Origin. 204 logout/đổi lần đầu chỉ xóa cookie sau commit; lỗi không cấp/xóa cookie. |
| Port và khóa | Giữ bốn use case và sáu repo ports mục 5.3. Làm rõ logout khóa user → employee → session để kiểm lại cả employee active, thống nhất với login/đổi lần đầu; không chạy bcrypt dưới khóa. |
| Thời gian sau khóa | Adapter đọc clock sau khi lấy đủ khóa; kiểm hạn và truyền timestamp tường minh, không dựa DEFAULT now() của PostgreSQL. Giữ nguyên schema; ca DB cạnh tranh thuộc AUTH-T08/T09/T17. |
| Config | Giữ tên APP_ORIGIN và SESSION_TTL; TTL phải phản ánh 8h đã duyệt. AUTH-T02/T03 hoàn thiện load/validation, không lấy default dev cũ làm chính sách production. |

**Phân việc Go để triển khai đúng tầng:**

| Tầng | Đầu vào / đầu ra và trách nhiệm |
|---|---|
| HTTP handler/middleware | Request → kiểm byte/JSON/header/cookie, Origin/CSRF/limiter → gọi use case → DTO hoặc mã lỗi. Không truy vấn pgx, không tự ghi/revoke session; không serialize LoginResult trực tiếp vì có secret. |
| Auth service | Credential/secret nội bộ → kiểm policy, bcrypt ngoài transaction, tạo secret/hash, gọi atomic port. Authenticate trả Principal hiện hành, không cache role; DTO hạn chế không sửa role DB. |
| Repository adapter | Projection và atomic ports → SQL tham số hóa, thứ tự khóa/đọc lại nền, commit/rollback. Không bcrypt/HTTP/cookie; dữ liệu lỗi trùng email khác found=false. Không trả thành công khi kết quả commit không rõ. |
| Composition root | Ghép config, repo, service, clock/random và router. Cùng clock cho service/adapter; không tạo framework DI hoặc package rỗng trong AUTH-T01. |

Lỗi ở service/adapter không mang HTTP status: InvalidCredentials → login 401; Unauthenticated → route bảo vệ 401; InvalidCurrentPassword/Validation → 400; PasswordChangeNotRequired → 403; Unavailable/Internal → 503/500. Transport sở hữu FORBIDDEN do Origin/CSRF, PAYLOAD_TOO_LARGE (413), UNSUPPORTED_MEDIA_TYPE (415) và RATE_LIMITED (429); PASSWORD_CHANGE_REQUIRED là cổng nghiệp vụ, không lỗi me/csrf của phiên hợp lệ. Mỗi lỗi có no-store và request_id; không trả lỗi driver/secret.

**Ca nhận việc HTTP/FE bổ sung cho ma trận mục 6.4:**

| Ca | Kỳ vọng / nơi kiểm khi có runtime |
|---|---|
| Không role đặc biệt, cờ false; ADMIN có cờ true | Login/me 200 với [] ở cả hai nhưng cờ khác; FE chỉ vào đổi lần đầu khi true. HTTP/service + FE, không suy từ role rỗng. |
| Body đúng 8192 byte và 8193 byte, không khai độ dài | 8192 vẫn phải qua các validation khác; 8193 là 413 trước bcrypt/write. HTTP, không chỉ kiểm schema. |
| Logout body rỗng, thiếu Content-Type nhưng có phiên/Origin/CSRF | Revoke commit rồi 204/xóa cookie; không bị 415. Body không rỗng là 400 và không revoke. |
| Employee bị vô hiệu trong lúc logout chờ khóa | Adapter đọc lại dưới khóa, không trả 204 từ snapshot active cũ; DB concurrency test ở AUTH-T08/T17. |
| Login/đổi lỗi hoặc commit không rõ | Không Set-Cookie thành công/xóa cookie; phân biệt 401/400/403/429 với DB 503/500. Không tự retry POST. |
| Role/cờ hoặc actor giả trong request | 400 vì field ngoài allowlist; không cấp quyền, không thay cờ DB. |

**Còn lại đúng task phụ thuộc:** AUTH-T02/T03 chốt hành vi loader/config và kiểm môi trường; AUTH-T05/T06 review mapping legacy/SQL/fixture và quyền DB test; AUTH-T12 chốt biểu diễn token/cookie trong helper và kiểm browser thực tế theo A07; AUTH-T13 hiện thực limiter có biên/đồng thời/dọn bộ nhớ; AUTH-T18 chốt màn đích auth thật tách mock. Các chi tiết biểu diễn/helper không được đổi entropy/TTL/thuật toán/allowlist đã chốt. Reset/gửi thư/retention vẫn ngoài feature. Không chạy lại import Jira chỉ vì nguồn proposal/OpenAPI được chỉnh: journal đã gắn fingerprint bản import cũ; tiến độ/bằng chứng đưa lên ticket thủ công khi được giao, không xóa journal hoặc tạo lại ticket.

Bằng chứng kiểm ngày 07/11: `node --test scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs` đạt **23/23** (11 contract, 12 harness); `node scripts/check-harness.mjs` đạt **215 file**; `git diff --check` đạt. Kiểm tương thích công cụ Jira bằng test offline đạt **29/29**, offline plan vẫn 1 Epic/20 Task/33 WBS/105h/47 links; không gọi Jira thật. Không đổi Go/FE nên không chạy lại cổng Go/FE trong task tài liệu này.

Kiểm tĩnh nằm ở scripts/check-auth-contract.test.mjs; mở rộng ma trận status/code và ví dụ DTO, không gọi đó là validator OpenAPI đầy đủ. AUTH-S01–S10 vẫn cần bằng chứng runtime ở đúng task; chưa chạy HTTP/SQL/browser. Lệnh/kết quả kiểm tra thực tế bàn giao cùng diff, không đánh dấu trước các checklist runtime hoặc Jira Done.

#### 6.2.8. Bàn giao cookie, Origin và CSRF — DATN-13 / AUTH-T12, 10/11/2026

[Route CSRF](../../internal/transport/http/auth_csrf.go) giải mã một cookie phiên, gọi `Authenticate` và trả `csrf_token` cho cả phiên thường lẫn phiên hạn chế hợp lệ. Token là HMAC theo A07, biểu diễn bằng base64url không padding như cookie. GET chỉ đọc phiên và không gửi `Set-Cookie`. JSON lỗi có `request_id` do server cấp để đối chiếu log. Route che lỗi lưu trữ và nội dung panic, giữ `no-store` và không mở CORS.

- [Bộ helper HTTP](../../internal/transport/http/auth_protection.go): DATN-16 dùng `allowsOrigin` trước xác thực login, đặt cookie sau commit với mốc `expires_at`. `setSessionCookie` trả false nếu mốc cookie đã qua. DATN-15/DATN-17 dùng `allowsMutation` với secret của phiên đã xác thực, rồi xóa cookie sau khi thao tác lưu trữ đã commit.
- [Parser Origin dùng chung](../../internal/config/validation.go): startup và HTTP so scheme, host và port hiệu lực từ cùng cách đọc. Cấu hình dùng hostname ASCII theo dạng browser gửi, ví dụ punycode cho tên miền quốc tế, hoặc địa chỉ IP đầy đủ. Host Unicode và IPv4 rút gọn/hex bị chặn tại startup để tránh so sai origin hoặc đợi đến mutation mới phát hiện cấu hình lỗi. Parser kiểm cấu trúc và tập ký tự, không xác minh DNS hoặc tính hợp lệ của mọi nhãn punycode.
- [Router Chi](../../internal/transport/http/router.go): đăng ký `GET /api/auth/csrf` khi có dependency xác thực. DATN-18 chịu trách nhiệm ghép config, adapter PostgreSQL, service và router ở `cmd/server` theo P09.1/P09.4. Hành vi cookie trên trình duyệt HTTPS và Vite proxy cần kiểm theo P09.2.

Bằng chứng trên Go 1.26.5: `go test ./internal/transport/http -count=1 -v` kiểm route Chi với service thật và repository giả, gồm mã 200/401/503/500, phiên hạn chế, HMAC vector độc lập, biên hết hạn, cookie trùng, Origin sai, token cũ và che marker nhạy cảm. `go vet ./...`, `go test ./... -count=1`, `go build ./...`, `go test -race ./internal/config ./internal/transport/http -count=1` và 24 ca contract/harness đều đạt. Bằng chứng này thuộc tầng Go/HTTP của AUTH-S02, S04, S09 và S10. PostgreSQL tích hợp, trình duyệt và các mutation dùng helper được nghiệm thu ở các ticket tương ứng.

#### 6.2.9. Bàn giao limiter — DATN-14 / AUTH-T13, 10/11/2026

[Limiter HTTP](../../internal/transport/http/auth_rate_limit.go) giữ bốn bộ đếm trong một tiến trình, theo các ngưỡng ở mục 6.2.4–6.2.5. Mỗi request đã nhận hoặc lần xác minh thất bại hết hiệu lực tại đúng mốc cửa sổ tính từ sự kiện đó. Mỗi bảng giữ tối đa 4.096 khóa, dọn khóa hết hạn khi có lưu lượng và không đẩy khóa còn hiệu lực ra để nhận khóa mới. Lượt chờ kết quả vẫn được giữ cho đến khi caller kết thúc. Khi bão hòa bộ nhớ, khóa mới nhận 429 với thời gian chờ thử lại theo nhịp dọn. Bộ đếm mất khi tiến trình khởi động lại.

- Tạo `newAuthRateLimiter(clock)` một lần khi lắp handler. Dùng `authPeerIP` lấy IP từ `RemoteAddr`; nếu parse thất bại, trả lỗi nội bộ và dừng thay vì bỏ qua giới hạn. `allowLoginIP` và `allowPasswordChangeIP` trả thời gian chờ dương nếu bị chặn. Gọi cổng IP trước phần xác thực tốn chi phí. Backend sau proxy chỉ thấy IP của proxy khi chưa có cấu hình proxy tin cậy.
- DATN-16 gọi `beginLoginAttempt` với email qua validation và dùng cùng quy tắc trim/chữ thường cho định danh login. Khóa trong bộ nhớ là SHA-256 của email đã chuẩn hóa. Ghi nhận thất bại nhất quán cho email không tồn tại, tài khoản không dùng được hoặc mật khẩu sai ở bước xác minh để 429 không tiết lộ sự tồn tại của tài khoản. Lỗi do nền thay đổi sau khi đã xác minh đúng cần tín hiệu nội bộ riêng, dù có thể cùng trả 401. Kiểm tính nhất quán giữa chuẩn hóa ở Go và lookup PostgreSQL khi nối luồng thật.
- DATN-17 gọi `beginPasswordChangeAttempt` với principal lấy từ phiên hợp lệ đã qua kiểm tra Origin/CSRF và quyền đổi lần đầu. Khóa là `UserID`. Chỉ lần xác minh mật khẩu hiện tại sai mới được ghi vào bộ đếm định danh.
- Nếu `begin…` trả callback, đặt `defer finish(false)` ngay để trả lượt trên mọi đường thoát, rồi gọi `finish(true)` khi xác minh thực sự thất bại. Chỉ lần gọi đầu có hiệu lực. Nếu bị chặn, gọi `writeAuthRateLimited` trong chuỗi auth cấp request ID nội bộ. Helper trả 429, `RATE_LIMITED`, `Cache-Control: no-store` và `Retry-After` làm tròn lên theo giây. Khi các lượt đang xác minh giữ hết chỗ, thời gian thử lại là một giây.

Bằng chứng trên Go 1.26.5: `go vet ./...`, `go test ./... -count=1`, `go build ./...`, `go test -race ./internal/transport/http -count=1` và 30 lượt chạy [bộ test limiter](../../internal/transport/http/auth_rate_limit_test.go) đều đạt. Contract/harness đạt 24 ca. DATN-14 bàn giao tầng Go của AUTH-S08. Kiểm response qua hai handler thật, dữ liệu PostgreSQL và việc đi qua proxy thuộc các ticket nối route và tích hợp.

#### 6.2.10. Bàn giao danh tính và đăng xuất — DATN-15 / AUTH-T15, 11/11/2026

[Middleware phiên](../../internal/transport/http/auth_session.go) xác thực cookie một lần và chuyển principal cùng secret qua context của request. [Router Chi](../../internal/transport/http/router.go) dùng middleware chung cho `me`, `csrf` và `logout`. [Handler me](../../internal/transport/http/auth_me.go) ánh xạ sáu trường từ principal hiệu lực sang DTO. [Handler logout](../../internal/transport/http/auth_logout.go) kiểm Origin, CSRF và body rỗng trước khi gọi service. Service thu hồi đúng phiên qua repository và handler chỉ xóa cookie sau khi lời gọi thành công.

DATN-17 dùng tiếp middleware này cho bước xác thực trước khi đổi mật khẩu lần đầu và bổ sung cổng phiên hạn chế cho nhóm route nghiệp vụ. DATN-18 truyền service và cấu hình Origin/cookie thật vào router, rồi kiểm với PostgreSQL test. Bộ `httptest` ở DATN-15 chạy qua Chi và auth service với repository giả, gồm danh tính phiên thường/hạn chế, lỗi phiên/lưu trữ, route yêu cầu cookie, logout 204, cookie cũ bị từ chối và phiên thứ hai vẫn dùng được. Kiểm HTTPS và Vite proxy nằm ở đợt tích hợp.

Các lệnh đã đạt trên Go 1.26.5 gồm `go vet ./...`, `go test ./... -count=1`, `go build ./...` và `go test -race ./internal/transport/http -count=1`. Contract và harness đạt 24 ca với `node --test scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`. Kiểm kê harness đạt với `node scripts/check-harness.mjs`.

#### 6.2.11. Bàn giao đăng nhập bằng email — DATN-16 / AUTH-T14, 11/11/2026

[Handler login](../../internal/transport/http/auth_login.go) kiểm Origin, giới hạn IP, Content-Type, giới hạn 8192 byte đọc thực tế và hai trường JSON trước khi giữ lượt xác minh theo email. Email dùng chung quy tắc chuẩn hóa với limiter và service. [Service login](../../internal/service/auth/login.go) xác minh bcrypt bên ngoài transaction. Thiếu định danh hoặc trạng thái không hoạt động dùng hash giả với cost mặc định của dự án. Mật khẩu giữ nguyên khoảng trắng và không áp min 15 của mật khẩu mới. Đầu vào vượt 72 byte bị từ chối như credential sai để tránh bcrypt chấp nhận theo tiền tố 72 byte. Phép so giả giảm chênh lệch thời gian, không bảo đảm thời gian giống nhau nếu hash cũ có cost khác.

Service phân biệt xác minh thất bại với nền định danh đổi tại `FinalizeLogin`. Hai lỗi cùng trả 401/INVALID_CREDENTIALS, chỉ lỗi xác minh cộng vào bucket email. Service sinh secret 32 byte và gửi hash SHA-256 xuống adapter. Nếu request mang cookie cũ đọc được, chỉ chuyển hash cũ. Adapter hiện có quyết định thu hồi theo chủ phiên và hiệu lực dưới khóa. Handler dùng chung DTO với me và đặt cookie theo expires_at sau khi nhận kết quả thành công.

Đạt `go vet ./...`, `go test ./... -count=1`, `go build ./...` và `go test -race ./internal/service/auth ./internal/transport/http -count=1`. Test service và Chi dùng bcrypt thật cùng repository giả. Bằng chứng gồm body 8192/8193 byte không Content-Length, cookie/DTO cho phiên thường và hạn chế, lỗi random, lỗi lưu trữ và limiter trước xác minh. Các test fake chỉ xác nhận lời gọi qua port và kết quả HTTP.

DATN-18 ghép `NewLoginService`, service phiên và adapter PostgreSQL dùng cùng clock. Khởi tạo một limiter cho tiến trình và truyền TTL đã validate. Kiểm chuẩn hóa email qua SQL, giao dịch cùng HTTP và cookie/Origin qua browser hoặc proxy ở đợt tích hợp. `NewRouter` của server hiện giữ cách khởi tạo cũ.

#### 6.2.12. Bàn giao đổi mật khẩu lần đầu — DATN-17 / AUTH-T16

[Service đổi lần đầu](../../internal/service/auth/password_change.go) đọc nền bằng hash phiên, kiểm hiệu lực/cờ, áp policy 15 code point và 72 byte UTF-8 cho mật khẩu mới, rồi so và băm bằng bcrypt ngoài transaction. Mật khẩu hiện tại vượt 72 byte được xử lý như xác minh sai khi các kiểm tra trước đó đạt, tránh chấp nhận theo tiền tố. Port nhận hash đã xác minh và hash mới để adapter ở mục 6.1.11 kiểm lại dưới khóa. Service không thử lại mutation.

[Handler đổi lần đầu](../../internal/transport/http/auth_password_change.go) dùng Origin/CSRF, giới hạn theo IP, JSON 8 KiB và limiter theo UserID hiện hành. Parser JSON dùng chung với login trả 400 cho Unicode escape chứa surrogate đơn trước khi gọi service, đồng thời giữ cặp hợp lệ và ký tự U+FFFD được nhập hợp lệ. Chỉ lỗi xác minh sai mới tính vào bộ đếm sai của user. Sau khi service trả thành công, handler xóa cookie và trả 204 rỗng. [Router](../../internal/transport/http/router.go) khai báo tường minh bốn route được phép với phiên hạn chế. Route nghiệp vụ mới dùng `authenticateSession` để từ chối cờ true trước khi xét role. Route thử của guard chỉ có trong test.

[Unit tests](../../internal/service/auth/password_change_test.go) chứng minh policy và command qua fake repository. [HTTP tests](../../internal/transport/http/auth_password_change_test.go) kiểm route, cookie, lỗi, limiter và guard bằng stub. DATN-18 chịu trách nhiệm nối router với PostgreSQL trong startup và kiểm lại luồng trên adapter thật.

Kiểm tra đã chạy và đạt: `go vet ./...`, `go test ./... -count=1`, `go build ./...`, `go test -race ./internal/service/auth ./internal/transport/http -count=1`, `node --test scripts/check-auth-contract.test.mjs scripts/check-harness.test.mjs`, `node scripts/check-harness.mjs` và `git diff --check`. Bằng chứng đợt này ở tầng Go/HTTP và contract tĩnh.

#### 6.2.13. Tích hợp năm API với PostgreSQL — DATN-18 / AUTH-T17, 12/11/2026

`cmd/server.startServer` gọi `buildHandler` sau khi config đã validate và pool đã mở. Composition root khởi tạo một `AuthRepository`, ba service dùng chung clock và TTL từ config. `NewAuthRouter` nhận đủ use case, dựng protection/limiter một lần cho router và đăng ký đủ năm route. Dependency thiếu trả lỗi trước khi server lắng nghe. Health giữ đường kiểm tra hiện có. Các ghi nhận chưa nối startup ở mục 6.2.8–6.2.12 được tiếp nối bởi đợt này.

[HTTP integration startup](../../cmd/server/auth_integration_test.go) gọi cùng `buildHandler` như server, chạy request qua `httptest.Server`, pool pgx và PostgreSQL 18 test thật. [Pool test](../../internal/testutil/authdb/database.go) sao chép cấu hình DB case đã xác minh, kiểm cluster/name/owner và đóng trước cleanup case. Mỗi fixture nâng schema tới 000003 trong case riêng. Runner cấp biến test trực tiếp, không gọi loader dotenv hoặc dùng DATABASE_URL.

Đã kiểm login/email chuẩn hóa, cookie khớp hash phiên đã commit, me/csrf không gia hạn, logout đúng phiên, rotation cùng user, nhiều phiên, đổi lần đầu và login mới. Role/cờ/status đọc lại từ DB, ID user khác employee vẫn trả đúng DTO. [Guard integration](../../internal/transport/http/auth_router_integration_test.go) chặn ADMIN hạn chế bằng repository thật; endpoint thăm dò chỉ được đăng ký trong test. [Route test](../../internal/transport/http/auth_router_test.go) đối chiếu method/path Chi với năm API trong OpenAPI và health, kiểm constructor thiếu dependency.

[HTTP concurrency](../../cmd/server/auth_concurrency_integration_test.go) dùng hai request qua pool, connection điều khiển và barrier PostgreSQL. Advisory lock trong trigger fixture giữ transaction đầu ở sau khi lấy row lock, `pg_blocking_pids` xác nhận request sau thực sự chờ. Kiểm hai lần đổi và login đã xác minh hash cũ bị từ chối sau khi nền thay đổi. Ca logout hết hạn sau chờ khóa dùng blocker riêng và clock atomic, kiểm không có write. Timeout chặn treo test; không dùng sleep để suy request đã tới barrier.

Trigger fixture kiểm rollback ở write và deferred constraint tại commit cho login/logout/đổi lần đầu; HTTP trả 500, không Set-Cookie, hash/cờ/phiên không có write nửa chừng. SQLSTATE 57P03 mô phỏng DB tạm lỗi kiểm 503. [Commit mất phản hồi](../../cmd/server/auth_commit_integration_test.go) thực hiện commit PostgreSQL rồi wrapper driver trả lỗi I/O: HTTP 503, không cookie và chỉ một lần commit; DB vẫn phản ánh mutation đã commit. Đây là mô phỏng mất phản hồi ở biên driver, chưa cắt mạng DB hoặc HTTP thật. Lỗi random tiếp tục được chứng minh bằng unit test service/HTTP đã có.

| Bằng chứng | Truy vết backend và giới hạn |
|---|---|
| AUTH-S01 | Test login service/HTTP và projection PostgreSQL hiện có; integration kiểm chuẩn hóa, credential sai và email mâu thuẫn. Hash giả được kiểm ở service. |
| AUTH-S02 | HTTP lifecycle kiểm cookie sau commit, thu hồi đúng phiên, hạn 8h và biên hết hạn; browser/HTTPS thật còn ở đợt FE/vận hành. |
| AUTH-S03 | Guard PostgreSQL kiểm ADMIN còn cờ bị 403 trước handler, đổi cờ được đọc hiện hành. |
| AUTH-S04 | HTTP lifecycle kiểm CSRF sai không revoke; unit Origin/HMAC/rotation hiện có. Browser proxy chưa kiểm trong đợt này. |
| AUTH-S05 | Live principal kiểm role/cờ/status từ DB; projection/HTTP tests giữ ID và DTO allowlist, field lạ bị chặn. |
| AUTH-S06 | HTTP hai lần đổi/login hash cũ cạnh tranh, mọi phiên cùng user bị thu hồi, user khác còn phiên; tái sử dụng suite transaction. |
| AUTH-S07 | HTTP lỗi write/commit, DB tạm lỗi và commit mất phản hồi; random lỗi ở unit, không nhận là thử cắt mạng thật. |
| AUTH-S08 | HTTP dùng chung limiter giữa nhiều request, bucket login/đổi độc lập, Retry-After; policy/Unicode/body và các biên limiter ở unit hiện có. |
| AUTH-S09 | Test redaction input giả hiện có, HTTP integration kiểm lỗi SQL không lộ marker; không dùng secret hoặc dữ liệu nhân viên thật. |
| AUTH-S10 | Chi Walk đối chiếu method/path OpenAPI, năm route đi qua real services/repository, no-store và cookie/CSRF theo các ca HTTP. Chưa phải validator response schema OpenAPI đầy đủ. |

Kiểm tra đạt trên Go 1.26.5 và PostgreSQL 18 local test sở hữu: `go vet ./...`, `go test ./... -count=1`, `go build ./...`; vet với tag integration cho cmd/server, HTTP và testutil; 24 test contract/harness và kiểm kê nguồn. Bộ integration `-race -tags=integration -p 1` đạt cho migrations, repository/postgres, transport/http và cmd/server. Lượt gộp ban đầu phát hiện import thừa ở test commit; đã bỏ import và chạy lại cmd/server đạt. Lệnh tái hiện/kiểm cleanup ở [runbook DATN-18](../../scripts/authdb/README.md#tích-hợp-http-và-postgresql--datn-18--auth-t17).

Bàn giao AUTH-T18/AUTH-T19: FE dùng email công ty/password cho login, khôi phục danh tính bằng me, lấy csrf_token để gửi X-CSRF-Token cùng cookie cho mutation. Cờ true có role hiệu lực rỗng, chỉ dùng me/csrf/logout/change-password. Đổi thành công trả 204 và xóa cookie, phải login mới. 401 yêu cầu xác thực lại; 403 PASSWORD_CHANGE_REQUIRED đưa tới đổi mật khẩu; 429 đọc Retry-After; lỗi 5xx/mất phản hồi không tự retry mutation. FE thật, cookie qua browser/proxy, SMTP và Swagger UI còn ở đợt tương ứng. Đợt này không áp migration Neon hoặc thay schema/contract.


#### 6.3.1. Bàn giao validation cấu hình — DATN-3 / AUTH-T02, 07/11/2026

Người dùng đã duyệt brief trước code: `ValidateServer(Config) error` trong `internal/config`, nhận input đã có kiểu dữ liệu và không đọc OS/file, gọi DNS/DB hoặc thay input. Hàm trả lỗi đầu tiên theo thứ tự env → port → Origin → TTL; lỗi chỉ nêu trường/quy tắc, không chứa giá trị cấu hình hoặc lỗi parse URL thô. Không thêm dependency.

| Input đã parse | Contract của validator |
|---|---|
| AppEnv | Đúng `dev` hoặc `prod`, bắt buộc tường minh; không trim/fallback. |
| HTTPPort | 1..65535; không tự thay số không hợp lệ. |
| AppOrigin | HTTP/HTTPS + host và port hợp lệ nếu có; không userinfo/path (kể cả slash cuối)/query/fragment. HTTPS cho prod; ngoại lệ HTTP chỉ dev với localhost hoặc địa chỉ IP loopback, không phân giải DNS. |
| SessionTTL | Đúng 8h theo A03; zero/âm/thời hạn khác bị từ chối, không tự áp default. |

`Config` thêm AppOrigin và SessionTTL (`time.Duration`); Load chưa đọc hai trường này hoặc gọi validation, nên zero value từ loader cũ chưa được dùng để khởi tạo auth. Việc parse/default/load và tích hợp fail-fast trước kết nối DB thuộc AUTH-T03, không bật kiểm tra auth cho migrate/seed. Chưa thay hành vi server/health, RequireDev, URL PostgreSQL hoặc .env. DATN-3 không kiểm tra khả năng kết nối hay tính đúng của DatabaseURL; phần loader/kết nối có trách nhiệm riêng.

Code và table-driven tests tại `internal/config/validation.go` / `validation_test.go`. Lượt kiểm ngày 07/11 đạt 46 ca validation, `go vet ./...`, `go test ./... -count=1`, `go build ./...`; contract/harness 23/23 và kiểm kê 217 file đạt, `git diff --check` đạt. Cổng Go toàn repo chạy lại ngoài sandbox hạn chế build cache; không chạy server hoặc DB. Kiểm thử phủ cấu hình hợp lệ, biên port/TTL, Origin sai cấu trúc/nguồn/IPv6, giữ input và redaction bằng marker giả; không gọi Load hoặc đọc bí mật. Bằng chứng cổng Go/harness bàn giao theo lượt chạy thực tế, không đồng nghĩa đã có auth runtime. Giữ checkpoint Quân tự viết API theo bộ ticket, không nhận test cấu hình thay kiểm thử HTTP/SQL sau này.

#### 6.3.2. Tích hợp loader/startup — DATN-4 / AUTH-T03, 08/11/2026

Chốt tiếp nối được người dùng duyệt trước triển khai: `LoadServer()` đọc/parse cấu hình và gọi `ValidateServer()` trước kết nối DB; `Load()` dành migrate/seed không yêu cầu HTTP/auth. Hai đường dùng chung nguồn đọc, không thay OS env. Phần “chưa nối loader” ở bàn giao DATN-3 mô tả thời điểm trước đợt này, không là hiện trạng mới.

| Nguồn/giá trị | Hành vi đã chốt |
|---|---|
| APP_ENV | Chỉ lấy môi trường tiến trình, đúng dev/prod, không default/trim; .env không chọn hoặc đổi môi trường. Áp dụng cả CLI. |
| Dotenv | Chỉ đọc khi APP_ENV=dev tường minh; prod không đọc. File không tồn tại được bỏ qua; lỗi đọc/định dạng trả lỗi an toàn. |
| Precedence | OS env > file > default; phân biệt không đặt với đặt rỗng. Không os.Setenv, không lấy file để cấp quyền guard. |
| HTTP_PORT | Default 8080 khi không đặt; đã đặt rỗng/sai bị từ chối. Chỉ server dùng. |
| APP_ORIGIN | Server bắt buộc, theo validator DATN-3; không tự đoán. |
| SESSION_TTL | Default 8h khi không đặt; parse time.Duration rồi kiểm đúng 8h. Không mở khả năng đổi hạn nghiệp vụ. |
| DATABASE_URL | Server/CLI đều bắt buộc, không rỗng/toàn khoảng trắng; không parse kết nối hoặc sửa pooler ở đợt này. |

Dotenv hỗ trợ KEY=VALUE, khóa ASCII chữ/underscore đầu và chữ/số/underscore tiếp theo, comment cả dòng, dòng trống, CRLF, giá trị bọc nháy đơn/đôi. Không nội suy/escape/export/comment cuối dòng/nhiều dòng; khóa trùng và định dạng sai bị từ chối. `$VAR` trong giá trị không được nội suy; cấu hình có chứa dấu `#` trong giá trị giữ nguyên. Lỗi chỉ nêu dòng/quy tắc, không in giá trị hoặc lỗi OS/parse thô. Không thêm thư viện parser.

`cmd/server` dùng cổng `runWithStartup`: tải hợp lệ mới gọi `startServer`, nơi mở pool/HTTP. Test thay riêng bước chạy bằng callback để kiểm cấu hình hợp lệ được bàn giao và lỗi dừng trước kết nối; không giả làm test DB/HTTP thật. RequireDev, validateCommand và thứ tự guard seed/rollback được giữ nguyên. Xử lý URL pooler/lỗi kết nối thuộc AUTH-T04; chưa coi toàn bộ lỗi DB legacy đã được redaction.

Nguồn code: `internal/config/config.go`, `validation.go`, `loader_test.go`; cổng startup và test tại `cmd/server/main.go`, `main_test.go`. `.env.example`/README cập nhật cách khởi động rõ APP_ENV; không đọc/sửa .env thật. Kiểm tra ngày 08/11: `go vet ./...`, `go test ./... -count=1`, `go build ./...` và race tests config/server đạt; contract/harness 23/23, kiểm kê 219 nguồn và diff/gofmt đạt. Cổng Go chạy ngoài sandbox để dùng build cache; không mở DB/server thật hoặc đọc .env thật. Không nhận kiểm thử cấu hình thay test auth API/SQL.

Review bàn giao DATN-4 ngày 08/11: đối chiếu AC và bổ sung kiểm thử startup với dotenv sai định dạng/lỗi đọc thật bằng đường dẫn thư mục tạm; dev phải dừng trước callback, prod bỏ qua cùng nguồn lỗi. Bổ sung kiểm giá trị parser giữ nguyên dấu `=`, `#`, khoảng trắng trong nháy, chuỗi `$VAR` và giá trị rỗng. Chạy lại toàn bộ cổng Go/race và 23 test contract/harness đều đạt; không phát hiện lỗi chặn bàn giao trong phạm vi này. Coverage nhánh chạy DB/HTTP (`startServer`) chưa có, không dùng coverage cổng cấu hình để nhận là server/DB đã chạy.

Đầu vào cho task phụ thuộc:

| Task nhận | Đầu vào/giới hạn bàn giao |
|---|---|
| AUTH-T04 / DATN-5 | LoadServer/Load đã chọn nguồn an toàn; DatabaseURL chỉ được kiểm bắt buộc, chưa parse/che lỗi kết nối. Tiếp tục sửa IsPooledEndpoint/DirectURL và lỗi adapter theo contract riêng; không suy URL hiện tại đã an toàn. |
| AUTH-T12 / DATN-13 | Config.AppEnv, AppOrigin và SessionTTL đã parse/validate khi lấy từ LoadServer; không dùng Load của CLI để dựng auth. Helper cookie/Origin/CSRF còn phải thực thi và kiểm thử policy, cùng các phụ thuộc AUTH-T10/T11. |
| AUTH-T13 / DATN-14 | Có cổng startup/default đã chốt; dùng ngưỡng limiter hiện hành trong contract, không tự thêm biến môi trường cho ngưỡng chưa được duyệt. |

Đầu ra sẵn sàng review/commit/PR; chưa tự chuyển Jira Done hoặc nhận đã merge. Không chạy lại import Jira để cập nhật fingerprint tài liệu.

#### 6.3.3. URL PostgreSQL và lỗi startup — DATN-5 / AUTH-T04, 08/11/2026

Brief đã được người dùng duyệt trước code: loại bỏ tìm/thay marker trên toàn DSN, parse URL/hostname và không chuyển tiếp driver error có thể chứa credential. Không đổi provider/TLS/pool size/protocol hay mở API/migration/schema. Mục này tiếp nối giới hạn URL/chẩn đoán còn lại của DATN-4.

`ParseDatabaseURL` trong `internal/config/database_url.go` là hàm thuần: không file/OS env/DNS/DB. Loader server/CLI parse một lần sau kiểm bắt buộc/kiểu dữ liệu; Config.Database chứa DatabaseEndpoint (`RuntimeURL`, `DirectURL`, `Pooled`). URL sai không trả endpoint một phần. Các trường chứa bí mật, không log Config/DatabaseEndpoint; chỉ truyền tới adapter tương ứng.

| Định dạng/ca | Hành vi đã chốt/triển khai |
|---|---|
| postgres:// hoặc postgresql://, một hostname/IPv4/IPv6, port 1..65535 nếu có | Hỗ trợ cú pháp; IPv6 trong [], không zone. Host DNS ASCII/punycode, không username/password/database bắt buộc mới. Không chứng minh host tồn tại/kết nối thành công. |
| Keyword DSN, nhiều host, Unix socket, host thiếu/sai, fragment, port sai | Từ chối an toàn bằng DB_URL_INVALID, không đoán endpoint. |
| Query host/hostaddr/port/user/password/dbname/database/service/servicefile | Từ chối để không ghi đè authority hoặc chuyển sang service/socket/multi-host; xét khóa sau URL-decode, không thay TLS/application_name/query không liên quan. |
| Neon AWS ep-…-pooler.region.aws.neon.tech | Pooled=true; runtime giữ URL gốc. DirectURL chỉ thay span hostname đã parse, bỏ -pooler ở nhãn endpoint; giữ từng byte credential/port/path/query. |
| Marker trong password/path/query, host ngoài mẫu, miền giả/suffix thêm | Không nhận pooled nhầm hoặc thay marker. Host ngoài mẫu giữ nguyên; cung cấp direct rõ ràng cho migration nếu cần. Không cam kết tự chuyển GCP/custom domain. |

Mẫu endpoint tham khảo [Neon connection details](https://github.com/neondatabase/website/blob/main/content/faqs/find-connection-details-neon-console.md); đây là phạm vi hỗ trợ của ứng dụng, không phải cam kết mọi dạng hostname Neon đều được chuyển.

`internal/repository/postgres.NewPool` nhận RuntimeURL đã qua loader, vẫn để pgx parse cấu hình driver. Ba bước parse/create/ping trả sentinel an toàn DB_URL_INVALID/DB_POOL_INIT_FAILED/DB_CONNECT_FAILED, không wrap/giữ nguyên nhân thô. Giữ max/min conn, timeout, lifetime và simple protocol cũ; pool đóng nếu ping thất bại, thành công để caller sở hữu. Test truyền poolDependencies để kiểm thứ tự/cleanup bằng lỗi giả, không khởi tạo socket/đọc file pgx thật.

`cmd/server.openStartupPool` chỉ trả sentinel đã biết (không trả wrapper dù errors.Is nhận nhóm đó), lỗi lạ quy về DB_CONNECT_FAILED; logStartupError ghi lỗi an toàn cùng ngữ cảnh. Unit tests kiểm log thực của logger startup với marker DSN/password/cookie giả. Migrate dùng DirectURL và che lỗi sql.Open/driver init; seed dùng RuntimeURL. Giữ nguyên RequireDev/thứ tự guard. Không refactor toàn bộ lỗi SQL của migration/seed hoặc nhận đã giải quyết mọi lỗi runtime từ task này.

Đầu vào AUTH-T17: gọi LoadServer rồi dùng Config.Database.RuntimeURL/Pooled cho NewPool; không dựng cấu hình auth bằng Load CLI hoặc log endpoint. Chưa nối API hoặc có bằng chứng DB/browser/transaction thật. Checkpoint Quân tự viết API giữ nguyên.

Bằng chứng ngày 08/11: parent đã chạy go vet ./..., go test ./... -count=1, go build ./...; race tests config/postgres/server và 23 test contract/harness đạt, kiểm kê 222 nguồn và diff/gofmt đạt. Test lỗi parse/create/ping, timeout/cleanup, protocol true/false, marker trong error chain/định dạng log; không mở socket DB hoặc chạy parser pgx thật từ pool tests. Loader/startup fixtures đã chuyển từ chuỗi giả không đúng URL sang URL giả đúng cú pháp, không đổi guard để làm test xanh.

Sub-agent GPT-6.1 Sol medium được bàn giao bối cảnh/quyết định/nguồn/phạm vi và chỉ đọc review diff (gồm untracked tests), đối chiếu pgx parser cục bộ; không tìm thấy finding cần sửa. Agent chạy offline go test config/postgres/server/migrate/seed và diff check đạt, không sửa file hoặc dùng DB/.env/network. Agent nêu thiếu ca protocol=false (đã bổ sung); hai cổng lỗi migration sql.Open/driver init chưa có test riêng, không coi là bằng chứng migration thật. Parent xác minh kết luận, rà diff và giữ giới hạn này để kiểm ở đợt DB test được cấp. Không tự chuyển Jira Done, commit/push hoặc chạy lại công cụ import. Một lượt agent review không thay kiểm thử runtime/tác giả hiểu code.


#### 9.2.1. Trạng thái đầu vào khi chuyển sang ticket — 07/11/2026

| Phần | Đã có | Còn phải nghiệm thu |
|---|---|---|
| P01.1/P01.4 | Quyết định A01–A09, body 8 KiB và thiết kế đổi lần đầu đã ghi nhận | Review tính khớp nguồn/contract, ghi kết quả và điểm còn mở; không hỏi duyệt lại lựa chọn |
| P01.2 | [OpenAPI năm route](../api/auth.openapi.json) và script kiểm tĩnh | Review/hoàn thiện contract cùng P01.4; kiểm tĩnh không thay handler/runtime |
| P01.3 | Chữ ký ports và bất biến ở mục 5.3 | Review đủ cả bốn use case và ca fake/test trước viết Go |
| P04.1–P04.3 | Kế hoạch SQL/môi trường/fixture ở mục 6.1.5–6.1.6; bằng chứng diễn tập 000002 được dẫn tại đó | Review SQL/mapping cụ thể; chuẩn bị môi trường test/guard và đặc tả fixture; chưa có DB test hoặc fixture auth chạy |
| P02/P03/P05–P09 | Thiết kế và WBS | Chưa triển khai code của feature; không đánh dấu Done từ proposal hoặc kiểm tĩnh |

Đây là kiểm kê đầu vào, không tự thay trạng thái Jira. Khi tạo ticket ghi rõ phần đã có, liên kết bằng chứng và phần còn lại; estimate gốc giữ để truy vết, phần còn lại ước lượng riêng sau khi nhận việc. Không tạo thêm ticket để soạn lại nội dung đã có.
