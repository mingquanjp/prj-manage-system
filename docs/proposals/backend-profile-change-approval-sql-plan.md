# Kế hoạch SQL và kiểm chứng — đề nghị hồ sơ / HR duyệt

Ngày 21/11/2026. **Kế hoạch bàn giao sau chốt proposal; SQL cụ thể mục9 chưa thực thi DDL/DB**. Đọc [proposal](backend-profile-change-approval.md) mục4/6; nghiệp vụ chỉ dẫn nguồn tại đó. Không là lệnh áp Neon, không sửa 000001–000005. Tên migration tiếp theo xác định lại lúc code theo head thực tế.

## 1. Cổng môi trường và preflight

- Trước DDL: xác minh PostgreSQL 18 test sở hữu bằng `python3 scripts/authdb/manage.py check`; kiểm head/dirty/column/index/FK đối chiếu file, không force version. Không đọc/fallback DATABASE_URL/.env/Neon, không cấp DB mới ngoài runner được duyệt, không chạy nhiều suite DB song song.
- Head repo hiện 000005: audit đã có; bảng profile requests/idempotency chưa có; notification legacy cần chuyển. Empty database ở đây là đi từ migration 1 tới head rồi tiến thêm, không bỏ bước lịch sử.
- Legacy fixture bắt buộc user_id khác employee_id, nhiều role, phone/address NULL, hai pending khác trường/cùng trường sau nâng; giữ nguyên ID/link/code/email/hash/status/role/session/danh mục/audit. Không ghi PII thật trong fixture.
- Kiểm notification employee→users liên kết một-một, FK, type/title/content/ref_task_id/is_read/created_at, consumer/index/view/dependency. Dữ liệu task notification chưa map được type/resource/message an toàn thì **dừng**, không suy từ content hoặc bỏ dòng. Không dựa vào bảng local trống để tuyên bố chuyển được mọi dữ liệu legacy.

## 2. Bảng đề nghị và ràng buộc

`employee_profile_change_requests` theo dictionary 5B mục10: identity INT request_id; employee_id FK; requested_by/resolved_by FK users; before_values/proposed_values JSONB; explanation; requested_at; status; resolved_at; decision_reason. FK không CASCADE mất lịch sử.

- CHECK status allowlist; PENDING có resolved_by/resolved_at/decision_reason NULL; terminal đủ resolver/time không trước requested_at; REJECTED reason có nội dung sau trim. APPROVED reason có thể NULL.
- JSON là object schema_version 1, chỉ fullname/phone/address + schema_version; hai bộ khóa khớp và ít nhất một trường nghiệp vụ. Before đúng kiểu nullable theo cột nguồn; proposed theo A06/BR-001. Kiểm DB kiểu/khóa/trạng thái và service validation trước SQL; CHECK tránh UNKNOWN của NULL bằng điều kiện tường minh, không tưởng JSONB tự kiểm schema.
- Không UNIQUE pending/employee, không cột version hoặc updated_at làm token. Index `(employee_id, requested_at DESC, request_id DESC)` và `(status, requested_at DESC, request_id DESC)`; list count/items cùng snapshot.
- Không DELETE/UPDATE patch đã gửi trong application ports. Quyết định chỉ UPDATE WHERE status=PENDING dưới khóa; kiểm affected rows. Không trigger/framework audit toàn app. Nếu muốn DB-level terminal immutability, phải review trigger cùng SQL cụ thể; hiện thiết kế chỉ đảm bảo đường ghi của application và constraints.

## 3. Idempotency

Bảng theo dictionary 5B mục19: identity, actor FK users, operation VARCHAR100, key VARCHAR128, hash BYTEA32, result_reference JSONB schema_version/request_id/result_code, created_at, expires_at. UQ actor/operation/key; check expires_at>created_at, hash 32 byte. A03 chốt expires_at=created_at+24h; chưa có cleanup tự động.

Khóa user actor serialize submit cùng actor, tra record trước so stale. Nếu chưa có, ghi request/notification rồi INSERT record kết quả đầy đủ trong cùng transaction; không cần placeholder. UNIQUE là chốt bổ sung; writer rollback thì follower có thể thực hiện một lần sau recheck. Dùng clock_timestamp() sau chờ khóa cho expiry, không sleep/retry vô hạn. Record chỉ reference, không full response/PII; chi tiết mục9.3.

Replay kiểm quyền hiện tại và linkage trước đọc kết quả; khác hash hoặc expired luôn conflict, không tự DELETE/INSERT lại. Kết quả cũ có thể terminal khi đọc lại, không dựng trạng thái PENDING giả. Missing reference là lỗi integrity → 500, không tạo mới. Không áp key chung cho create account/auth hoặc HR decision.

## 4. Chuyển notification legacy và rollback

Thiết kế đích dictionary 5B mục21: recipient_id FK users, notification_type/resource_type/resource_id/message, created_at/read_at. Không CASCADE lịch sử theo employee/task. Read_at>=created_at; index recipient/time/ID; thêm partial unread index nếu query plan chứng minh cần, chưa tự thêm dependency.

Đường SQL bàn giao mục9 chỉ chuyển khi notifications rỗng. Với bảng có dữ liệu, chưa có mapping type/content được xác minh nên dừng; khi có đầu vào mới, mapping recipient qua employee.user_id/type/resource/message phải review và validate trước bỏ cột cũ/index/FK. Notification profile mới resource_type=`PROFILE_CHANGE_REQUEST`, type=`PROFILE_CHANGE_SUBMITTED|PROFILE_CHANGE_APPROVED|PROFILE_CHANGE_REJECTED`. Message chung, không copy content legacy chứa PII sang message mới.

Nếu sau này mở đường nâng notification có dữ liệu: is_read=true không có thời điểm thực; phương án read_at=created_at chỉ bảo toàn trạng thái đọc, không phải thời điểm đọc thực. Chưa áp mapping đó trong đường migration bàn giao hiện tại. Type/ref không map được hoặc content cần bảo toàn chưa có nơi giữ thì dừng phần nâng dữ liệu đó, không tự dùng generic resource hoặc DROP. Fixture có bảng trống và fixture map được không thay mapping môi trường đích. Không đổi schema ngoài phạm vi này để lấp dữ liệu không hiểu.

Down không phải backup: request/notification/key đã phát sinh không thể trở về schema cũ mà giữ đầy đủ ngữ nghĩa. Down tự động **chỉ cho DB case có cả ba bảng rỗng**; từ chối atomically nếu còn request/idempotency/notification, kể cả notification legacy đã map. Không hứa mapping nghịch fixture là restore đầy đủ. Legacy read timestamps/cột cũ mất thông tin cần backup trước up nếu cần khôi phục nguyên trạng; không bịa ref_task_id/is_read từ profile notification. Runner dirty/transaction semantics phải được kiểm bằng failure injection thật, không force version để làm xanh.

## 5. Transaction, quyền và thứ tự khóa

Actor lấy user_id từ session server; middleware check chưa thay recheck trong transaction. Recheck actor user ACTIVE, employee ACTIVE, must_change_password=false và role HR khi cần; không dùng users.system_role legacy hoặc client principal làm quyền cuối. Session hết/revoke kiểm theo cơ chế auth hiện hành; không tự biến lỗi DB thành401.

Đã đối chiếu nguồn OAP: `organization_account.go` và `auth_sessions.go` khóa user trước employee; create giữ catalog KEY SHARE theo department → position → job_level; catalog mutation giữ FOR UPDATE rồi kiểm tham chiếu bằng SELECT thường. Thứ tự mở rộng dưới đây cần review SQL cụ thể trước code, không thêm đường khóa ngược:

1. Gom toàn bộ user liên quan (actor, target khi phải khóa, recipient HR của send) → khóa theo user_id tăng dần; gom employee liên quan → employee_id tăng dần. Đọc trước chỉ lấy danh sách ID, recheck liên kết/quyền sau khóa. Không khóa actor rồi mới tùy ý thêm recipient nhỏ hơn. Loại trùng IDs, kể cả actor=target hoặc actor cũng là recipient. Query role dùng EXISTS/DISTINCT, không nhân notification theo số role.
2. Catalog liên quan giữ KEY SHARE theo department → position → job_level, ID tăng dần trong mỗi loại; rồi request theo request_id, session target theo session_id nếu cần revoke ở F4. Không khóa employee target trước user target. Tập recipient HR chụp ở đầu transaction và recheck dưới khóa, không mở vòng khóa thêm thiếu thứ tự. Khóa user hiện hữu không chặn phantom user/role mới; không hứa tập nhận là toàn bộ HR tại commit. Người mới có HR sau snapshot không backfill; hàng chờ vẫn theo role hiện hành.
3. Ghi dữ liệu request/official + notification + idempotency result; direct edit F4 thêm audit. Commit xong mới trả success. Chưa có chính sách tự retry deadlock mutation; lỗi trả để GET đối soát.

**Send:** parse key/typed payload → khóa actor/recipient theo bộ đã gom → kiểm self/current status → serialize key → so expected before đúng field → chụp before → insert PENDING → insert notifications cho recipient đã recheck → result reference → commit. Hai lần gửi cùng key chỉ một tác động. Hai key khác nhau được phép tạo hai pending, không tự deduplicate nghiệp vụ.

**Approve:** đọc employee_id request để lập bộ khóa → khóa actor/target → khóa request → đọc lại linkage/PENDING/quyền → so official IS NOT DISTINCT FROM before từng trường → UPDATE đúng patch → UPDATE kết luận → notification owner → commit. Nếu employee target không còn ACTIVE thì theo PCA-A09 trả409 `EMPLOYEE_INACTIVE`, không giả actor hết phiên. Stale → 409 không sửa request hoặc notification. Không so toàn employee object vì trường khác không liên quan không nên chặn.

**Reject:** cùng bộ khóa, recheck HR/PENDING/linkage; không yêu cầu before khớp, không UPDATE employee; lưu reason/resolver/time + notify owner atomically. Theo PCA-A09, employee target inactive vẫn cho HR kết thúc pending bằng reject; user target bị khóa riêng không tự làm HR mất quyền quản lý hành chính. Không áp check phiên actor cho target; theo lựa chọn đã chốt trong proposal.

**Read:** query projection tường minh, không SELECT *; current status/role tại snapshot; request self hoặc HR, notification own user. Count/items/unread_count dùng một SQL hoặc REPEATABLE READ READ ONLY nếu nhiều SELECT; READ COMMITTED với nhiều SELECT không đủ. Mutation dùng READ COMMITTED và đọc lại sau khóa. q literal escape `%/_`, sort có ID phân định khi thời gian trùng, page rỗng vẫn total đúng.

**F4 direct edit:** gom/khóa actor và target users theo ID → employees theo ID → recheck HR/status/liên kết → catalog KEY SHARE theo bước 2 → so expected before đúng tập patch → kiểm email unique/canonical và FK → update employee + audit allowlist → email đổi thì khóa/revoke target sessions theo ID → commit. Actor tự sửa email: response mutation có thể thành công, request sau phải login lại; không Set-Cookie tạo phiên mới. Chưa có sự kiện notification direct edit trong nguồn, không tự thêm.

Tái sử dụng context.WithTimeoutCause của repository OAP, nhận timeout từ composition root; Begin/query/Commit dùng context giới hạn, cleanup rollback có context riêng hữu hạn. Giá trị cụ thể chốt tại contract task theo cấu hình hiện hành. Deadlock/lock timeout/statement timeout rollback và map503 SERVICE_UNAVAILABLE, không409 stale và không tự retry mutation. Commit mất xác nhận giữ outcome chưa rõ, không bảo đảm chưa ghi. Không hứa race không chờ vô hạn chỉ từ sơ đồ.

## 6. Ma trận kiểm chứng dự kiến

Đây là **kế hoạch**, mọi ca dưới đây chưa chạy trong đợt proposal.

| ID | Ca và bằng chứng cần có |
|---|---|
| PCA-DB01 | Nâng từ DB trống qua migration lịch sử tới head mới; constraints/index/FK/metadata đủ |
| PCA-DB02 | Nâng legacy giả user_id≠employee_id; bảo toàn hash/email/code/link/role/session/audit; notification mapping/unknown-type dừng |
| PCA-DB03 | Down/up DB case trống; từng bảng có dữ liệu thì down bị từ chối atomically, kể cả notification legacy; lỗi giữa DDL và dirty đúng runner/driver |
| PCA-DB04 | Gửi patch allowlist, address NULL→có; noop/unknown key/null/address trắng/Unicode/giới hạn; official không đổi |
| PCA-DB05 | Cùng key/hai connection chỉ một request và một bộ notification theo recipient; tìm key cũ trước khi so stale; key khác payload/expired conflict; lỗi ghi result phải rollback request/notification/key cùng nhau; replay sau quyết định |
| PCA-DB06 | Approve/reject; lỗi notification làm rollback cả quyết định/official; terminal không xử lý lại |
| PCA-DB07 | Hai HR approve hoặc approve/reject bằng barrier: một kết luận; không notification đôi |
| PCA-DB08 | Hai pending cùng trường stale; khác trường không stale giả; NULL so đúng; reject stale vẫn giữ official |
| PCA-DB09 | Actor role/status/flag đổi trước ghi; owner/HR/limited/ID khác; recipient snapshot/recheck, nhiều role không nhân dòng, actor=target/recipient không khóa lặp; target inactive và HR tự xử lý theo lựa chọn đã chốt |
| PCA-DB10 | Snapshot page/count/unread, read idempotent/own only; link cũ không cấp quyền; notification không PII |
| PCA-DB11 | F4: direct edit và approve tranh nhau, email unique/session revoke, catalog race/FK, no-op/audit failure rollback |
| PCA-DB12 | Commit không rõ + mất response: không auto replay/success giả; HTTP DTO/no-store/security/body PCA65536/65537 byte không Content-Length, Unicode/before legacy dài; auth/OAP vẫn8192; browser dữ liệu sau reload |

Race dùng hai connection/barrier, không sleep ngẫu nhiên. Test lỗi commit controlled chỉ chứng minh mapping và transaction theo fault đã dựng; không gọi là mất mạng vật lý. Notification recipient race cần giao thức tất cả writer role/status hiện có; nếu writer tương lai chưa tuân khóa thì ghi giới hạn, không tuyên bố đã test API chưa có.

## 7. Lệnh kiểm và bàn giao

Khi triển khai Go: `go vet ./...`, `go test ./... -count=1`, `go build ./...`. Sau guard PG18 sở hữu, tuần tự:

```sh
python3 scripts/authdb/manage.py check
python3 scripts/authdb/manage.py run -- go test -tags=integration -p 1 ./migrations ./internal/repository/postgres ./internal/transport/http ./cmd/server -count=1 -timeout=5m
```

Bổ sung package thực nếu code tạo package test mới. FE tại `web/`: `pnpm check`, rồi browser với backend test thật. Đợt tài liệu này chỉ chạy `node --test scripts/check-harness.test.mjs`, `node scripts/check-harness.mjs`, `git diff --check`; không dùng Go/PG/browser giả để nhận migration hoặc feature pass. Khi có contract mới thêm static contract test; khi có runtime thêm route inventory khớp canonical.

Bàn giao từng task: file/lệnh/kết quả thực/môi trường/mapping đã xác minh/cách cleanup và giới hạn. DB case runner dọn đúng DB sở hữu; không reset lượt `pnpm dev` đang dùng để demo hoặc dọn các DB lạ. HTTPS/Neon/deploy/mail ngoài phạm vi giữ chưa kiểm. Không đưa credential hoặc patch PII thật vào Git/Jira/evidence.

## 8. Review thiết kế 21/11

Đã sửa: body PCA riêng đủ cho before/proposed Unicode; recipient snapshot không hứa chặn phantom; loại trùng actor/target/recipient và role; down chỉ DB trống; query/count/projection tường minh; timeout map503; target inactive đã tách và chốt tại PCA-A09. Đây là tự review tài liệu, chưa là kiểm chứng khóa/migration/runtime. OpenAPI và SQL cụ thể tại mục9 đã soạn; mapping notification có dữ liệu chưa được xác minh, nên đường up hiện tại chỉ nhận bảng rỗng.

## 9. SQL cụ thể cho task triển khai

Các khối dưới là thiết kế SQL để chuyển thành migration/repository khi nhận ticket, **không chạy trong đợt contract**. Head phải kiểm lại trước đặt số migration. Transaction nghiệp vụ timeout context10s (đối chiếu internal/app/auth.go), rollback cleanup tối đa5s; migration lock_timeout5s/statement_timeout30s. Chốt giá trị cấu hình ở composition root, không mở transaction chờ người dùng.

### 9.1. Giới hạn chuyển notification có dữ liệu

Kiểm nguồn21/11: không có INSERT notifications trong cmd/internal/db/seed; chỉ DDL000001. Vì chưa có mapping loại/content được xác minh, đường migration bàn giao **chỉ nhận notifications rỗng**. Fixture legacy có notification phải chứng minh dừng và giữ nguyên dữ liệu/cấu trúc nghiệp vụ, không được nhận là đã migrate notification có dữ liệu. Đây là guard cho đầu vào chưa rõ, không đổi quyết định bảo toàn lịch sử.

Preflight của runner phải từ chối dữ liệu/dependency đã phát hiện trước khi đánh dấu version dirty, theo cơ chế M1/M2 hiện có. Guard trong DDL vẫn bắt buộc dưới ACCESS EXCLUSIVE để chặn thay đổi sau preflight. Nếu guard DDL hoặc DDL tiếp theo lỗi, transaction rollback dữ liệu/cấu trúc nhưng metadata schema_migrations có thể đã ở version dirty theo driver; kiểm riêng kết quả đó, không hứa metadata runner bất biến hoặc force version để chạy tiếp.

Khi gặp môi trường có notification: chủ dữ liệu xác nhận mapping type/resource/message, cách giữ title/content và hạn chế read_at lịch sử; review SQL chuyển riêng trước khi bỏ guard. Không chạy mapping giả trên dữ liệu đích hoặc dùng read_at=created_at như thời điểm đọc có thật. M1/M2 user/employee/role/session legacy vẫn phải nâng thành công nếu notification rỗng. Không chặn demo local mới vì dữ liệu của miền chưa có writer.

### 9.2. DDL tiến thêm dự kiến

```sql
BEGIN;
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';
LOCK TABLE public.notifications IN ACCESS EXCLUSIVE MODE;
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM public.notifications) THEN
    RAISE EXCEPTION 'PCA: notification legacy cần mapping đã review';
  END IF;
END $$;

CREATE FUNCTION public.pca_profile_values_valid(v jsonb, proposed boolean)
RETURNS boolean LANGUAGE plpgsql IMMUTABLE
SET search_path = pg_catalog AS $$
DECLARE k text; item jsonb; t text; max_len integer;
BEGIN
  IF v IS NULL OR jsonb_typeof(v) <> 'object' THEN RETURN false; END IF;
  IF v->'schema_version' IS DISTINCT FROM '1'::jsonb THEN RETURN false; END IF;
  IF (v - ARRAY['schema_version','fullname','phone','address']) <> '{}'::jsonb
     OR (v - 'schema_version') = '{}'::jsonb THEN RETURN false; END IF;
  FOR k,item IN SELECT key,value FROM jsonb_each(v - 'schema_version') LOOP
    IF item = 'null'::jsonb THEN
      IF k='fullname' OR (proposed AND k='address') THEN RETURN false; END IF;
      CONTINUE;
    END IF;
    IF jsonb_typeof(item) <> 'string' THEN RETURN false; END IF;
    t := item #>> '{}';
    IF k='fullname' AND length(t)=0 THEN RETURN false; END IF;
    IF proposed THEN
      max_len := CASE k WHEN 'fullname' THEN 100 WHEN 'phone' THEN 20 ELSE 2000 END;
      IF length(t) NOT BETWEEN 1 AND max_len
         OR t !~ U&'[^\0009\000A\000B\000C\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]'
         OR t ~ U&'^[\0009\000A\000B\000C\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]|[\0009\000A\000B\000C\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]$'
      THEN RETURN false; END IF;
    END IF;
  END LOOP;
  RETURN true;
END $$;

CREATE FUNCTION public.pca_same_profile_keys(a jsonb,b jsonb)
RETURNS boolean LANGUAGE plpgsql IMMUTABLE
SET search_path = pg_catalog AS $$
BEGIN
  IF a IS NULL OR b IS NULL OR jsonb_typeof(a)<>'object'
     OR jsonb_typeof(b)<>'object' THEN RETURN false; END IF;
  RETURN NOT EXISTS (
    SELECT 1 FROM (
      (SELECT jsonb_object_keys(a) EXCEPT SELECT jsonb_object_keys(b))
      UNION ALL
      (SELECT jsonb_object_keys(b) EXCEPT SELECT jsonb_object_keys(a))
    ) AS differing_keys
  );
END $$;

CREATE TABLE public.employee_profile_change_requests (
  request_id int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  employee_id int NOT NULL REFERENCES public.employees(employee_id),
  before_values jsonb NOT NULL,
  proposed_values jsonb NOT NULL,
  explanation text,
  requested_by int NOT NULL REFERENCES public.users(user_id),
  requested_at timestamptz NOT NULL DEFAULT now(),
  status text NOT NULL DEFAULT 'PENDING',
  resolved_by int REFERENCES public.users(user_id),
  resolved_at timestamptz,
  decision_reason text,
  CONSTRAINT ck_pca_values CHECK (
    public.pca_profile_values_valid(before_values,false)
    AND public.pca_profile_values_valid(proposed_values,true)
    AND public.pca_same_profile_keys(before_values,proposed_values)
    AND before_values <> proposed_values
  ),
  CONSTRAINT ck_pca_status CHECK (status IN ('PENDING','APPROVED','REJECTED')),
  CONSTRAINT ck_pca_decision CHECK ((
    (status='PENDING' AND resolved_by IS NULL AND resolved_at IS NULL AND decision_reason IS NULL)
    OR (status='APPROVED' AND resolved_by IS NOT NULL AND resolved_at IS NOT NULL AND decision_reason IS NULL)
    OR (status='REJECTED' AND resolved_by IS NOT NULL AND resolved_at IS NOT NULL
      AND decision_reason IS NOT NULL AND length(decision_reason) BETWEEN 1 AND 2000
      AND decision_reason ~ U&'[^\0009\000A\000B\000C\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]')
  ) IS TRUE),
  CONSTRAINT ck_pca_times CHECK (resolved_at IS NULL OR resolved_at>=requested_at),
  CONSTRAINT ck_pca_explanation CHECK (explanation IS NULL OR length(explanation)<=2000)
);
CREATE INDEX idx_pca_owner ON public.employee_profile_change_requests
  (employee_id,requested_at DESC,request_id DESC);
CREATE INDEX idx_pca_queue ON public.employee_profile_change_requests
  (status,requested_at DESC,request_id DESC);

CREATE TABLE public.idempotency_records (
  idempotency_id int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  actor_id int NOT NULL REFERENCES public.users(user_id),
  operation varchar(100) NOT NULL,
  idempotency_key varchar(128) NOT NULL,
  request_hash bytea NOT NULL CHECK (octet_length(request_hash)=32),
  result_reference jsonb NOT NULL CHECK (jsonb_typeof(result_reference)='object'),
  created_at timestamptz NOT NULL DEFAULT now(),
  expires_at timestamptz,
  CONSTRAINT uq_idempotency_actor_operation_key UNIQUE(actor_id,operation,idempotency_key),
  CONSTRAINT ck_idempotency_expiry CHECK (expires_at IS NULL OR expires_at>created_at)
);

-- Bảng đang rỗng và đã được khóa/guard; không reset sequence notification_id.
DROP INDEX public.idx_notifications_employee;
ALTER TABLE public.notifications
  DROP COLUMN employee_id,
  DROP COLUMN type,
  DROP COLUMN title,
  DROP COLUMN content,
  DROP COLUMN ref_task_id,
  DROP COLUMN is_read,
  ADD COLUMN recipient_id int NOT NULL REFERENCES public.users(user_id),
  ADD COLUMN notification_type text NOT NULL,
  ADD COLUMN resource_type text NOT NULL,
  ADD COLUMN resource_id int NOT NULL CHECK (resource_id>0),
  ADD COLUMN message text NOT NULL CHECK (length(message) BETWEEN 1 AND 200),
  ADD COLUMN read_at timestamptz,
  ADD CONSTRAINT ck_notifications_read_at CHECK (read_at IS NULL OR read_at>=created_at);
CREATE INDEX idx_notifications_recipient ON public.notifications
  (recipient_id,created_at DESC,notification_id DESC);
COMMIT;
```

Không dùng DROP CASCADE; preflight rà phụ thuộc lạ trước khi chạy. DROP COLUMN loại constraint/index nội bộ phụ thuộc cột theo PostgreSQL nhưng không được bỏ qua view/FK bên ngoài. Function giới hạn proposed canonical, before giữ nguyên nullable/độ dài legacy. NUL/Unicode hỏng, key trùng và JSON limit phải bị chặn trước DB; dictionary/state/linkage không chỉ dựa CHECK. Actor self=requested_by và trường terminal bất biến qua ports/recheck, không claim trigger toàn DB. Notification polymorphic kiểm resource_type/ID và liên kết ở service; không FK giả tới nhiều loại tài nguyên.

### 9.3. Serialize key, so nền và cập nhật

Không cần commit hoặc giữ placeholder có result rỗng. Mọi submit cùng actor trước hết khóa user actor theo giao thức mục5, vì vậy serialize được cùng actor/operation/key, kể cả không có row idempotency. UNIQUE là chốt bổ sung; đường ghi mới cùng operation phải giữ giao thức đó.

```sql
-- Sau khóa bộ user/employee tăng dần và recheck actor/self.
SELECT request_hash,result_reference,expires_at
FROM public.idempotency_records
WHERE actor_id=$1 AND operation='PROFILE_CHANGE_SUBMIT' AND idempotency_key=$2
FOR UPDATE;
```

Nếu có row: so hash, rồi expiry theo `clock_timestamp()` sau chờ lock; expired409, không chạy lại. Chỉ replay khi reference đúng schema_version1/request_id/result_code và request còn đúng owner; thiếu/linkage hỏng500. Nếu không có: so before, INSERT request, INSERT notifications cho recipient đã chọn/recheck, rồi INSERT record kết quả đầy đủ trước cùng COMMIT. Record dùng created_at=`clock_timestamp()`, expires_at=created_at+interval'24 hours'. UUIDv4 parse đúng một header, canonical lowercase; operation cố định PROFILE_CHANGE_SUBMIT; SHA256 canonical typed tuple(schema_version1,employee_id,expected_before_values,normalized_proposed_values,normalized_explanation). Giữ null khác string; key JSON sắp thứ tự, không tin hash client. Cùng key khác raw whitespace nhưng cùng nội dung sau normalize là cùng payload, trước dữ liệu nền không normalize.

Đối chiếu patch ví dụ address dưới khóa employee/request; fullname/phone làm cùng mẫu SQL tham số hóa. Allowlist cột từ code, không nối cột client:

```sql
SELECT address IS NOT DISTINCT FROM $2::text AS basis_matches
FROM public.employees WHERE employee_id=$1;
-- Chỉ sau xác nhận tất cả trường patch còn khớp và quyền/status/PENDING.
UPDATE public.employees SET address=$2,updated_at=now() WHERE employee_id=$1;
UPDATE public.employee_profile_change_requests
SET status='APPROVED',resolved_by=$2,resolved_at=clock_timestamp(),decision_reason=NULL
WHERE request_id=$1 AND status='PENDING'
RETURNING request_id;
-- Phải đúng1 dòng; INSERT notification owner trước COMMIT.
```

Reject chỉ UPDATE request WHERE PENDING, giữ official; reason canonical và notification cùng transaction. Sau khóa không tìm thấy request404; terminal409 trước so before; employee inactive409 chỉ approve. Mọi lỗi write/notification/result/commit đi qua typed outcome; rollback không tự replay.

```sql
-- Own read: actor đã qua business gate/recheck, không cần có quyền request/resource.
UPDATE public.notifications
SET read_at=COALESCE(read_at,clock_timestamp())
WHERE notification_id=$1 AND recipient_id=$2
RETURNING notification_id;
```

Không có row404, có row204 kể cả đã đọc; không nhận read_at client. Với list/count/unread nhiều SELECT dùng REPEATABLE READ READ ONLY, page rỗng vẫn total; role filter EXISTS tránh nhân row.

### 9.4. F4 và email legacy

SQL locks user/employee IDs tăng dần, rồi catalog department→position→job_level KEY SHARE (kể cả ID mới/null phải validate trước query); lock session cuối theo session_id khi revoke. Giao thức catalog/delete giữ theo OAP, không SELECT FOR UPDATE employee sau giữ catalog trong đường khác.

Company_email response đã canonical nên expected_before_values.company_email phải so **canonical email hiện tại theo M2**, không so raw padding trong DB. Before client không normalize thêm; nếu khác canonical hiện tại thì stale. proposed email normalize M2 trước so no-op. Canonical không đổi thì không UPDATE email legacy chỉ để bỏ padding/case, không audit/revoke giả. Các trường fullname/phone/address và ID so đúng raw/NULL đã đọc. Khi email thực đổi: UPDATE employee.email, UNIQUE M2 kiểm race, audit before/after allowlist, revoke sessions cùng transaction. Chỉ phần nào thực đổi mới audit; no-op toàn patch200 không ghi. Department/job_level null là tháo liên kết được phép; position không null. Không thay role/status/must_change_password/secret.

### 9.5. Down dự kiến và kiểm SQL

Down khóa ba bảng ACCESS EXCLUSIVE, từ chối nếu bất kỳ bảng còn row. Chỉ khi trống: bỏ index/cột notification đích, khôi phục chính xác cột/index/FK notifications000001 (employee_id/type/title/content/ref_task_id/is_read) và giữ sequence notification_id; DROP request/idempotency rồi DROP hai function PCA, tất cả trong một BEGIN/COMMIT với timeout như up. Không down M2/M1 kèm theo, không bịa dữ liệu lịch sử. Mã DDL down thực phải đối chiếu000001 và driver dirty bằng test trước bàn giao.

Các khối SQL trên mới được review tĩnh, chưa parse/execute qua PostgreSQL. Task migration phải kiểm cả lỗi giữa DDL, before JSON hỏng/NULL/unknown schema, code point DB/Go, pending/terminal constraints và guard legacy; không ghi pass SQL từ Node contract test.

DDL down tương ứng để review (chưa chạy):

```sql
BEGIN;
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';
LOCK TABLE public.employee_profile_change_requests,
           public.idempotency_records,public.notifications IN ACCESS EXCLUSIVE MODE;
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM public.employee_profile_change_requests)
     OR EXISTS (SELECT 1 FROM public.idempotency_records)
     OR EXISTS (SELECT 1 FROM public.notifications) THEN
    RAISE EXCEPTION 'PCA: từ chối down khi còn lịch sử/key/notification';
  END IF;
END $$;
DROP INDEX public.idx_notifications_recipient;
ALTER TABLE public.notifications
  DROP COLUMN recipient_id,
  DROP COLUMN notification_type,
  DROP COLUMN resource_type,
  DROP COLUMN resource_id,
  DROP COLUMN message,
  DROP COLUMN read_at,
  ADD COLUMN employee_id int NOT NULL REFERENCES public.employees(employee_id) ON DELETE CASCADE,
  ADD COLUMN type varchar(50) NOT NULL,
  ADD COLUMN title varchar(200) NOT NULL,
  ADD COLUMN content text,
  ADD COLUMN ref_task_id int REFERENCES public.tasks(task_id) ON DELETE CASCADE,
  ADD COLUMN is_read boolean NOT NULL DEFAULT false;
CREATE INDEX idx_notifications_employee ON public.notifications
  (employee_id,is_read,created_at DESC);
DROP TABLE public.employee_profile_change_requests;
DROP TABLE public.idempotency_records;
DROP FUNCTION public.pca_same_profile_keys(jsonb,jsonb);
DROP FUNCTION public.pca_profile_values_valid(jsonb,boolean);
COMMIT;
```

CASCADE ở FK khôi phục là đúng DDL legacy000001 trên bảng trống, không phải lệnh DROP CASCADE để vượt guard. Cột created_at và identity notification_id/sequence được giữ xuyên up/down; không reset high-water mark để lấp số. Nếu có dependency lạ, down phải dừng và rollback toàn bộ.
