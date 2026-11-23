//go:build integration

package migrations_test

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"ai-staffing/internal/service/auth"
	"ai-staffing/internal/testutil/authdb"
	"ai-staffing/migrations"
	"github.com/golang-migrate/migrate/v4"
)

func TestM2EmptyAndUnusedDownUp(t *testing.T) {
	db := authdb.NewCase(t)
	m := db.OpenMigrator(t, migrations.FS)
	authdb.Require(t, "nâng DB rỗng qua chuỗi lịch sử tới M2", m.Migrate(migrations.M2Version))
	head(t, m, migrations.M2Version)
	assertM2Present(t, db)
	assertM2Sequence(t, db, 1, false)
	if !errors.Is(m.Migrate(migrations.M2Version), migrate.ErrNoChange) {
		t.Fatal("startup/up lặp lại không được đổi schema hoặc sequence")
	}
	assertM2Sequence(t, db, 1, false)
	authdb.Require(t, "down M2 rỗng chưa từng cấp số", m.Steps(-1))
	head(t, m, migrations.M1Version)
	assertM2Absent(t, db)
	authdb.Require(t, "up M2 lại trên case chưa dùng", m.Migrate(migrations.M2Version))
	head(t, m, migrations.M2Version)
	assertM2Present(t, db)
	assertM2Sequence(t, db, 1, false)
}

func TestM2LegacyPreservesDataAndStartsAfterEMPNumbers(t *testing.T) {
	db := authdb.NewCase(t)
	m := db.OpenMigrator(t, migrations.FS)
	authdb.Require(t, "nền lịch sử v3", m.Migrate(3))
	fixture(t, db)
	exec(t, db.Conn, `UPDATE employees SET employee_code=CASE employee_id WHEN 201 THEN 'EMP001' WHEN 202 THEN 'EMP000001' ELSE 'E999999' END;
		UPDATE users SET must_change_password=true WHERE user_id=101;
		INSERT INTO user_roles VALUES(101,'HR'),(101,'DIRECTOR');
		INSERT INTO auth_sessions(user_id,token_hash,created_at,expires_at) VALUES
		(101,decode(repeat('ab',32),'hex'),'2026-01-01Z','2026-01-01 08:00Z');
		UPDATE employees SET email='  FIXTURE-ADMIN@EXAMPLE.INVALID  ' WHERE user_id=101`)
	authdb.Require(t, "nâng M1", m.Migrate(migrations.M1Version))
	before := m2Snapshot(t, db)
	authdb.Require(t, "nâng legacy M1 tới M2", m.Migrate(migrations.M2Version))
	head(t, m, migrations.M2Version)
	assertM2Present(t, db)
	assertM2Sequence(t, db, 2, false)
	sameJSON(t, before, m2Snapshot(t, db))

	// Ràng buộc canonical áp dụng cả user LOCKED và employee INACTIVE.
	expectState(t, db.Conn, "23505", `UPDATE employees SET email='fixture-admin@example.invalid' WHERE user_id=102`)
	sameJSON(t, before, m2Snapshot(t, db))
	if !errors.Is(m.Migrate(migrations.M2Version), migrate.ErrNoChange) {
		t.Fatal("up head lần hai phải giữ dữ liệu và sequence")
	}
	assertM2Sequence(t, db, 2, false)
}

func TestM2SequenceInitialRange(t *testing.T) {
	for _, tc := range []struct {
		name, code string
		want       int64
	}{
		{"only legacy E", "E001", 1},
		{"unpadded EMP", "EMP001", 2},
		{"padded EMP", "EMP000001", 2},
		{"seven digits", "EMP1000000", 1000001},
		{"long padded number", "EMP000000000000000000000000001", 2},
		{"near bigint maximum", "EMP9223372036854775806", 9223372036854775807},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := authdb.NewCase(t)
			m := db.OpenMigrator(t, migrations.FS)
			authdb.Require(t, "nền v3", m.Migrate(3))
			fixture(t, db)
			exec(t, db.Conn, `UPDATE employees SET employee_code=$1 WHERE user_id=101`, tc.code)
			authdb.Require(t, "nâng M2 với dải mã tổng hợp", m.Migrate(migrations.M2Version))
			assertM2Sequence(t, db, tc.want, false)
		})
	}
}

func TestM2PreflightRejectsBeforeDirty(t *testing.T) {
	for _, tc := range []struct{ name, sql string }{
		{"normalized duplicate", `UPDATE employees SET email=' FIXTURE-ADMIN@EXAMPLE.INVALID ' WHERE user_id=102`},
		{"blank email", `UPDATE employees SET email='   ' WHERE user_id=101`},
		{"malformed email", `UPDATE employees SET email='broken@@example.invalid' WHERE user_id=101`},
		{"display name", `UPDATE employees SET email='Fixture <fixture@example.invalid>' WHERE user_id=101`},
		{"address list", `UPDATE employees SET email='a@example.invalid,b@example.invalid' WHERE user_id=101`},
		{"header newline", `UPDATE employees SET email=E'fixture@example.invalid\r\nX-Test: fixture' WHERE user_id=101`},
		{"unicode trim mismatch", `UPDATE employees SET email=U&'\00A0fixture@example.invalid\00A0' WHERE user_id=101`},
		{"tab trim mismatch", `UPDATE employees SET email=E'\tfixture@example.invalid\t' WHERE user_id=101`},
		{"sequence exhausted by legacy", `UPDATE employees SET employee_code='EMP9223372036854775807' WHERE user_id=101`},
		{"numeric overflow", `UPDATE employees SET employee_code='EMP9223372036854775808' WHERE user_id=101`},
		{"preexisting sequence", `CREATE SEQUENCE public.employee_code_seq`},
		{"preexisting canonical index", `CREATE INDEX uq_employees_normalized_email ON employees(email)`},
		{"username view dependency", `CREATE VIEW public.m2_username_dependency AS SELECT username FROM public.users`},
		{"username type drift", `ALTER TABLE users ALTER COLUMN username TYPE text`},
		{"email type drift", `ALTER TABLE employees ALTER COLUMN email TYPE text`},
		{"username default drift", `ALTER TABLE users ALTER COLUMN username SET DEFAULT 'unreviewed-default'`},
		{"username check dependency", `ALTER TABLE users ADD CONSTRAINT ck_m2_username CHECK (length(username)>0)`},
		{"username function dependency", `CREATE FUNCTION public.m2_read_username() RETURNS text LANGUAGE sql AS $$ SELECT username FROM public.users LIMIT 1 $$`},
		{"custom users trigger", `CREATE FUNCTION public.m2_trigger_fn() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RETURN NEW; END $$;
			CREATE TRIGGER m2_custom_trigger BEFORE UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION public.m2_trigger_fn()`},
		{"employee user FK cascade drift", `ALTER TABLE employees DROP CONSTRAINT employees_user_id_fkey;
			ALTER TABLE employees ADD FOREIGN KEY(user_id) REFERENCES users(user_id) ON DELETE CASCADE`},
		{"employee code unique missing", `ALTER TABLE employees DROP CONSTRAINT employees_employee_code_key`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := authdb.NewCase(t)
			m := db.OpenMigrator(t, migrations.FS)
			authdb.Require(t, "nền v3", m.Migrate(3))
			fixture(t, db)
			authdb.Require(t, "nền M1", m.Migrate(migrations.M1Version))
			exec(t, db.Conn, tc.sql)
			before := m2Snapshot(t, db)
			if err := m.Migrate(migrations.M2Version); err == nil || !strings.HasPrefix(err.Error(), "M2:") {
				t.Fatal("phải dừng tại preflight M2 trước DDL/dirty; không xuất dữ liệu lỗi")
			}
			head(t, m, migrations.M1Version)
			sameJSON(t, before, m2Snapshot(t, db))
			var usernameExists bool
			authdb.Require(t, "username chưa bị drop", db.Conn.QueryRow(context.Background(), `SELECT EXISTS(
				SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name='users' AND column_name='username')`).Scan(&usernameExists))
			if !usernameExists {
				t.Fatal("preflight từ chối nhưng username đã mất")
			}
		})
	}
}

// So sánh với collation thật của case test, không giả định Unicode lower luôn trùng Go.
func TestM2EmailNormalizationParityUsesActualPostgres(t *testing.T) {
	metadataRecorded := false
	for i, raw := range []string{"  MIXED.CASE@EXAMPLE.INVALID  ", "ĐẶNG@example.invalid", "Σ@example.invalid", "İ@example.invalid"} {
		t.Run(fmt.Sprintf("normalization-case-%d", i+1), func(t *testing.T) {
			db := authdb.NewCase(t)
			m := db.OpenMigrator(t, migrations.FS)
			authdb.Require(t, "nền v3", m.Migrate(3))
			fixture(t, db)
			authdb.Require(t, "nền M1", m.Migrate(migrations.M1Version))
			if !metadataRecorded {
				var encoding, ctype, provider, collation string
				authdb.Require(t, "metadata normalization PostgreSQL test", db.Conn.QueryRow(context.Background(),
					`SELECT current_setting('server_encoding'), datctype, datlocprovider::text, datcollate
					 FROM pg_database WHERE datname=current_database()`).Scan(&encoding, &ctype, &provider, &collation))
				t.Logf("normalization environment: server_encoding=%s lc_ctype=%s provider=%s datcollate=%s", encoding, ctype, provider, collation)
				metadataRecorded = true
			}
			exec(t, db.Conn, `UPDATE employees SET email=$1 WHERE user_id=101`, raw)
			var pgNormalized string
			authdb.Require(t, "chuẩn hóa theo PostgreSQL đích", db.Conn.QueryRow(context.Background(), `SELECT lower(btrim($1::text))`, raw).Scan(&pgNormalized))
			before := m2Snapshot(t, db)
			matches := auth.NormalizeLoginEmail(raw) == pgNormalized
			t.Logf("Go/PostgreSQL normalization match=%t", matches)
			err := m.Migrate(migrations.M2Version)
			if !matches {
				if err == nil || !strings.HasPrefix(err.Error(), "M2:") {
					t.Fatal("normalization lệch phải từ chối trước DDL")
				}
				head(t, m, migrations.M1Version)
			} else {
				authdb.Require(t, "normalization khớp cho phép nâng M2", err)
				head(t, m, migrations.M2Version)
			}
			sameJSON(t, before, m2Snapshot(t, db))
		})
	}
}

func TestM2NextvalRollbackAndConcurrentConnections(t *testing.T) {
	db := authdb.NewCase(t)
	m := db.OpenMigrator(t, migrations.FS)
	authdb.Require(t, "nâng M2", m.Migrate(migrations.M2Version))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	second := db.Connect(t)
	firstTx, err := db.Conn.Begin(ctx)
	authdb.Require(t, "transaction cấp mã đầu", err)
	defer firstTx.Rollback(context.Background())
	var first int64
	authdb.Require(t, "cấp số trước rollback", firstTx.QueryRow(ctx, `SELECT nextval('public.employee_code_seq')`).Scan(&first))
	if first != 1 {
		t.Fatal("lần cấp đầu phải là 1")
	}
	// Barrier kênh chỉ mở connection thứ hai sau khi transaction đầu đã cấp số.
	// Connection hai phải cấp được số khác ngay cả khi transaction đầu chưa kết thúc.
	type outcome struct {
		number int64
		err    error
	}
	ready := make(chan struct{})
	result := make(chan outcome, 1)
	go func() {
		<-ready
		var number int64
		err := second.QueryRow(ctx, `SELECT nextval('public.employee_code_seq')`).Scan(&number)
		result <- outcome{number, err}
	}()
	close(ready)
	select {
	case got := <-result:
		authdb.Require(t, "cấp số connection hai trong lúc transaction một mở", got.err)
		if got.number != 2 {
			t.Fatal("connection hai không được cấp trùng số")
		}
	case <-ctx.Done():
		t.Fatal("connection hai không cấp được số tại barrier")
	}
	authdb.Require(t, "rollback row nghiệp vụ không rollback sequence", firstTx.Rollback(ctx))
	var next int64
	authdb.Require(t, "số sau rollback", db.Conn.QueryRow(ctx, `SELECT nextval('public.employee_code_seq')`).Scan(&next))
	if next != 3 {
		t.Fatal("sequence phải giữ khoảng trống, không tái dùng số rollback")
	}
}

func TestM2DownRefusesDataAndUsedSequenceAtomically(t *testing.T) {
	for _, tc := range []string{"data", "rolled back nextval", "used then emptied"} {
		t.Run(tc, func(t *testing.T) {
			db := authdb.NewCase(t)
			m := db.OpenMigrator(t, migrations.FS)
			if tc == "data" {
				authdb.Require(t, "nền v3", m.Migrate(3))
				fixture(t, db)
			}
			authdb.Require(t, "nâng M2", m.Migrate(migrations.M2Version))
			if tc != "data" {
				tx, err := db.Conn.Begin(context.Background())
				authdb.Require(t, "transaction thử cấp số", err)
				var number int64
				authdb.Require(t, "cấp số", tx.QueryRow(context.Background(), `SELECT nextval('public.employee_code_seq')`).Scan(&number))
				if tc == "rolled back nextval" {
					authdb.Require(t, "rollback sau nextval", tx.Rollback(context.Background()))
				} else {
					authdb.Require(t, "commit cấp số khi không có row", tx.Commit(context.Background()))
				}
			}
			before := m2Snapshot(t, db)
			var last int64
			var called bool
			authdb.Require(t, "snapshot sequence trước down", db.Conn.QueryRow(context.Background(), `SELECT last_value,is_called FROM public.employee_code_seq`).Scan(&last, &called))
			if m.Steps(-1) == nil {
				t.Fatal("down phải từ chối khi còn dữ liệu hoặc từng cấp số")
			}
			m.Close(t)
			assertM2Present(t, db)
			assertM2Sequence(t, db, last, called)
			sameJSON(t, before, m2Snapshot(t, db))
			var version uint
			var dirty bool
			authdb.Require(t, "dirty của down thực", db.Conn.QueryRow(context.Background(), `SELECT version,dirty FROM public.schema_migrations`).Scan(&version, &dirty))
			if version != migrations.M1Version || !dirty {
				t.Fatal("down lỗi phải giữ target version M1 dirty theo driver, không force")
			}
		})
	}
}

func TestM2DDLFailureRollsBackAndLeavesDirty(t *testing.T) {
	db := authdb.NewCase(t)
	m := db.OpenMigrator(t, migrations.FS)
	authdb.Require(t, "nền v3", m.Migrate(3))
	fixture(t, db)
	authdb.Require(t, "nền M1", m.Migrate(migrations.M1Version))
	before := m2Snapshot(t, db)
	var usernamesBefore string
	authdb.Require(t, "snapshot username trước lỗi DDL chỉ trong bộ nhớ", db.Conn.QueryRow(context.Background(),
		`SELECT jsonb_agg(jsonb_build_array(user_id,username) ORDER BY user_id)::text FROM public.users`).Scan(&usernamesBefore))
	m.Close(t)
	source := fstest.MapFS{}
	entries, err := fs.ReadDir(migrations.FS, ".")
	authdb.Require(t, "đọc migration source", err)
	for _, entry := range entries {
		content, err := fs.ReadFile(migrations.FS, entry.Name())
		authdb.Require(t, "đọc migration", err)
		if entry.Name() == "000005_oap_m2.up.sql" {
			content = []byte(strings.Replace(string(content), "COMMIT;", "SELECT 1/0;\nCOMMIT;", 1))
		}
		source[entry.Name()] = &fstest.MapFile{Data: content}
	}
	broken := db.OpenMigrator(t, source)
	if broken.Up() == nil {
		t.Fatal("lỗi DDL cố ý phải làm migration thất bại")
	}
	broken.Close(t)
	assertM2Absent(t, db)
	sameJSON(t, before, m2Snapshot(t, db))
	var usernamesAfter string
	authdb.Require(t, "username được rollback nguyên vẹn", db.Conn.QueryRow(context.Background(),
		`SELECT jsonb_agg(jsonb_build_array(user_id,username) ORDER BY user_id)::text FROM public.users`).Scan(&usernamesAfter))
	sameJSON(t, usernamesBefore, usernamesAfter)
	retry := db.OpenMigrator(t, migrations.FS)
	var dirty migrate.ErrDirty
	if !errors.As(retry.Up(), &dirty) || dirty.Version != migrations.M2Version {
		t.Fatal("driver phải giữ M2 dirty và từ chối retry")
	}
}

func m2Snapshot(t *testing.T, db *authdb.Case) string {
	t.Helper()
	var snapshot string
	authdb.Require(t, "snapshot bảo toàn M2 trong bộ nhớ", db.Conn.QueryRow(context.Background(), `SELECT json_build_object(
		'users',(SELECT jsonb_agg(to_jsonb(u)-'username' ORDER BY user_id) FROM public.users u),
		'employees',(SELECT jsonb_agg(to_jsonb(e) ORDER BY employee_id) FROM public.employees e),
		'roles',(SELECT jsonb_agg(to_jsonb(r) ORDER BY user_id,role_code) FROM public.user_roles r),
		'sessions',(SELECT jsonb_agg(to_jsonb(s) ORDER BY session_id) FROM public.auth_sessions s),
		'departments',(SELECT jsonb_agg(to_jsonb(d) ORDER BY department_id) FROM public.departments d),
		'positions',(SELECT jsonb_agg(to_jsonb(p) ORDER BY position_id) FROM public.positions p),
		'levels',(SELECT jsonb_agg(to_jsonb(j) ORDER BY job_level_id) FROM public.job_levels j),
		'audit',(SELECT jsonb_agg(to_jsonb(a) ORDER BY audit_id) FROM public.audit_events a))::text`).Scan(&snapshot))
	return snapshot
}

func assertM2Present(t *testing.T, db *authdb.Case) {
	t.Helper()
	var ok bool
	authdb.Require(t, "schema M2", db.Conn.QueryRow(context.Background(), `SELECT
		NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name='users' AND column_name='username')
		AND EXISTS (SELECT 1 FROM pg_sequences WHERE schemaname='public' AND sequencename='employee_code_seq'
		AND data_type='bigint'::regtype AND min_value=1 AND increment_by=1 AND cache_size=1 AND NOT cycle)
		AND EXISTS (SELECT 1 FROM pg_index i JOIN pg_class c ON c.oid=i.indexrelid
		WHERE c.relnamespace='public'::regnamespace AND c.relname='uq_employees_normalized_email'
		AND i.indisunique AND i.indisvalid AND i.indisready AND i.indpred IS NULL)
		AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name='users'
		AND column_name='system_role' AND is_nullable='NO')`).Scan(&ok))
	if !ok {
		t.Fatal("schema M2/sequence/index không đúng hoặc username chưa bỏ")
	}
}

func assertM2Absent(t *testing.T, db *authdb.Case) {
	t.Helper()
	var ok bool
	authdb.Require(t, "schema M2 chưa có", db.Conn.QueryRow(context.Background(), `SELECT
		EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name='users'
		AND column_name='username' AND data_type='character varying' AND character_maximum_length=50 AND is_nullable='NO')
		AND to_regclass('public.employee_code_seq') IS NULL
		AND to_regclass('public.uq_employees_normalized_email') IS NULL`).Scan(&ok))
	if !ok {
		t.Fatal("M2 phải không để DDL nửa chừng")
	}
}

func assertM2Sequence(t *testing.T, db *authdb.Case, last int64, called bool) {
	t.Helper()
	var actual int64
	var actualCalled bool
	authdb.Require(t, "đọc trạng thái sequence", db.Conn.QueryRow(context.Background(), `SELECT last_value,is_called FROM public.employee_code_seq`).Scan(&actual, &actualCalled))
	if actual != last || actualCalled != called {
		t.Fatal("sequence last_value/is_called sai; không xuất dữ liệu")
	}
}
