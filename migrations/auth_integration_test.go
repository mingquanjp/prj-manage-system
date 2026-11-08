//go:build integration

// Các ca này chỉ được chạy qua scripts/authdb/manage.py run trên cluster test sở hữu.
package migrations_test

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"ai-staffing/internal/testutil/authdb"
	"ai-staffing/migrations"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

func head(t *testing.T, m *authdb.Migrator, expected uint) {
	t.Helper()
	v, dirty, err := m.Version()
	authdb.Require(t, "đọc version", err)
	if v != expected || dirty {
		t.Fatal("version/dirty không đúng kỳ vọng")
	}
}

func exec(t *testing.T, conn *pgx.Conn, query string, args ...any) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := conn.Exec(ctx, query, args...)
	authdb.Require(t, "fixture/assert SQL", err)
}

func fixture(t *testing.T, db *authdb.Case) {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("Only synthetic local fixture 2026"), bcrypt.DefaultCost)
	authdb.Require(t, "hash fixture", err)
	exec(t, db.Conn, `INSERT INTO positions(position_name) VALUES ('Fixture position');
 INSERT INTO departments(department_name) VALUES ('Fixture department')`)
	for i, role := range []string{"ADMIN", "PM", "EMPLOYEE"} {
		// ID khác employee ID; toàn bộ dữ liệu giả, không dùng seed ứng dụng.
		userID := int64(101 + i)
		employeeID := int64(201 + i)
		exec(t, db.Conn, `INSERT INTO users(user_id,username,password_hash,system_role,status,created_at,updated_at)
   OVERRIDING SYSTEM VALUE VALUES($1,$2,$3,$4,$5,'2026-01-01Z','2026-01-02Z')`, userID, "fixture-"+role, string(hash), role, []string{"ACTIVE", "LOCKED", "DISABLED"}[i])
		exec(t, db.Conn, `INSERT INTO employees(employee_id,user_id,employee_code,fullname,email,position_id,department_id,employment_status,created_at,updated_at)
   OVERRIDING SYSTEM VALUE VALUES($1,$2,$3,$4,$5,1,$6,$7,'2026-01-01Z','2026-01-02Z')`, employeeID, userID, "FIX-"+role, "Fixture "+role, "fixture-"+role+"@example.invalid", []any{nil, 1, nil}[i], []string{"ACTIVE", "INACTIVE", "RESIGNED"}[i])
	}
}

func snapshot(t *testing.T, db *authdb.Case) string {
	t.Helper()
	var value string
	err := db.Conn.QueryRow(context.Background(), `SELECT json_build_object(
 'users',(SELECT json_agg(row_to_json(u) ORDER BY user_id) FROM users u),
 'employees',(SELECT json_agg(row_to_json(e) ORDER BY employee_id) FROM employees e),
 'positions',(SELECT json_agg(row_to_json(p) ORDER BY position_id) FROM positions p),
 'departments',(SELECT json_agg(row_to_json(d) ORDER BY department_id) FROM departments d))::text`).Scan(&value)
	authdb.Require(t, "snapshot nội bộ", err)
	return value
}

func roles(t *testing.T, db *authdb.Case) {
	t.Helper()
	var rows string
	err := db.Conn.QueryRow(context.Background(), `SELECT coalesce(string_agg(user_id::text||':'||role_code,',' ORDER BY user_id,role_code),'') FROM user_roles`).Scan(&rows)
	authdb.Require(t, "mapping quyền", err)
	if rows != "101:ADMIN" {
		t.Fatal("mapping legacy không đúng; không xuất dữ liệu")
	}
}

func TestAuthMigrationEmpty(t *testing.T) {
	db := authdb.NewCase(t)
	m := db.OpenMigrator(t, migrations.FS)
	authdb.Require(t, "up từ DB rỗng", m.Up())
	head(t, m, 3)
	var count int
	authdb.Require(t, "đếm nền auth", db.Conn.QueryRow(context.Background(), `SELECT count(*) FROM information_schema.tables WHERE table_schema='public' AND table_name IN ('user_roles','auth_sessions')`).Scan(&count))
	if count != 2 {
		t.Fatal("thiếu bảng auth")
	}
	if !errors.Is(m.Up(), migrate.ErrNoChange) {
		t.Fatal("up lại phải không thay đổi")
	}
}

func TestAuthMigrationLegacy(t *testing.T) {
	db := authdb.NewCase(t)
	m := db.OpenMigrator(t, migrations.FS)
	authdb.Require(t, "up version 1", m.Migrate(1))
	fixture(t, db)
	authdb.Require(t, "preflight legacy v1", legacyPreflight(db, false))
	before := snapshot(t, db)
	authdb.Require(t, "up version 2", m.Migrate(2))
	head(t, m, 2)
	var count int
	authdb.Require(t, "cờ cũ false", db.Conn.QueryRow(context.Background(), `SELECT count(*) FROM users WHERE must_change_password`).Scan(&count))
	if count != 0 {
		t.Fatal("000002 không giữ tài khoản cũ false")
	}
	// Bỏ đúng cột mới khi so snapshot toàn bộ hàng trước 000002.
	var usersAfter string
	authdb.Require(t, "snapshot cũ sau nâng", db.Conn.QueryRow(context.Background(), `SELECT json_build_object(
 'users',(SELECT json_agg(to_jsonb(u)-'must_change_password' ORDER BY user_id) FROM users u),
 'employees',(SELECT json_agg(row_to_json(e) ORDER BY employee_id) FROM employees e),
 'positions',(SELECT json_agg(row_to_json(p) ORDER BY position_id) FROM positions p),
 'departments',(SELECT json_agg(row_to_json(d) ORDER BY department_id) FROM departments d))::text`).Scan(&usersAfter))
	sameJSON(t, before, usersAfter)
	authdb.Require(t, "preflight legacy v2", legacyPreflight(db, true))
	baseline := snapshot(t, db)
	authdb.Require(t, "up auth", m.Up())
	head(t, m, 3)
	sameJSON(t, baseline, snapshot(t, db))
	roles(t, db)
	constraints(t, db)
}

func TestAuthMigrationPreservesFlagsAndRollback(t *testing.T) {
	db := authdb.NewCase(t)
	m := db.OpenMigrator(t, migrations.FS)
	authdb.Require(t, "up version 2", m.Migrate(2))
	fixture(t, db)
	exec(t, db.Conn, `UPDATE users SET must_change_password=true WHERE user_id=101`)
	authdb.Require(t, "preflight trước auth", legacyPreflight(db, true))
	before := snapshot(t, db)
	authdb.Require(t, "up auth", m.Up())
	head(t, m, 3)
	sameJSON(t, before, snapshot(t, db))
	roles(t, db)
	// Quyền mới và phiên chỉ nạp sau khi bảng đã tồn tại.
	exec(t, db.Conn, `INSERT INTO user_roles VALUES(102,'HR'),(102,'DIRECTOR');
 INSERT INTO auth_sessions(user_id,token_hash,created_at,expires_at) VALUES(101,decode(repeat('ab',32),'hex'),'2026-01-01Z','2026-01-01 08:00Z')`)
	authdb.Require(t, "down auth trên bản test", m.Steps(-1))
	head(t, m, 2)
	sameJSON(t, before, snapshot(t, db))
	var absent bool
	authdb.Require(t, "bảng auth đã down", db.Conn.QueryRow(context.Background(), `SELECT to_regclass('public.user_roles') IS NULL AND to_regclass('public.auth_sessions') IS NULL`).Scan(&absent))
	if !absent {
		t.Fatal("down chưa xóa đúng bảng auth")
	}
	authdb.Require(t, "up auth lại", m.Up())
	head(t, m, 3)
	roles(t, db)
	var sessions int
	authdb.Require(t, "phiên không phục hồi giả", db.Conn.QueryRow(context.Background(), `SELECT count(*) FROM auth_sessions`).Scan(&sessions))
	if sessions != 0 {
		t.Fatal("down/up không được giả phục hồi phiên")
	}
	sameJSON(t, before, snapshot(t, db))
}

func sameJSON(t *testing.T, a, b string) {
	t.Helper()
	var left, right any
	authdb.Require(t, "decode snapshot trước", json.Unmarshal([]byte(a), &left))
	authdb.Require(t, "decode snapshot sau", json.Unmarshal([]byte(b), &right))
	if !reflect.DeepEqual(left, right) {
		t.Fatal("dữ liệu legacy thay đổi; không xuất snapshot/hash/PII")
	}
}

func expectState(t *testing.T, conn *pgx.Conn, state, query string, args ...any) {
	t.Helper()
	_, err := conn.Exec(context.Background(), query, args...)
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != state {
		if errors.As(err, &pgErr) {
			t.Fatalf("SQLSTATE nhận %s, cần %s; không xuất SQL/dữ liệu", pgErr.Code, state)
		}
		t.Fatal("SQLSTATE ràng buộc không đúng kỳ vọng; không xuất SQL/dữ liệu")
	}
}

func constraints(t *testing.T, db *authdb.Case) {
	t.Helper()
	conn := db.Conn
	t.Run("roles", func(t *testing.T) {
		expectState(t, conn, "23505", `INSERT INTO user_roles VALUES(101,'ADMIN')`)
		expectState(t, conn, "23503", `INSERT INTO user_roles VALUES(999999,'HR')`)
		for _, role := range []string{"PM", "PTCM", "EMPLOYEE", "admin", ""} {
			expectState(t, conn, "23514", `INSERT INTO user_roles VALUES(102,$1)`, role)
		}
		expectState(t, conn, "23502", `INSERT INTO user_roles VALUES(102,NULL)`)
		expectState(t, conn, "23502", `INSERT INTO user_roles VALUES(NULL,'HR')`)
		exec(t, conn, `INSERT INTO user_roles VALUES(102,'HR'),(102,'DIRECTOR')`)
		exec(t, conn, `INSERT INTO users(user_id,username,password_hash,system_role) OVERRIDING SYSTEM VALUE
 VALUES(104,'role-fk-only','synthetic-not-login','EMPLOYEE')`)
		exec(t, conn, `INSERT INTO user_roles VALUES(104,'HR')`)
		expectState(t, conn, "23503", `DELETE FROM users WHERE user_id=104`)
	})
	t.Run("sessions", func(t *testing.T) {
		query := `INSERT INTO auth_sessions(user_id,token_hash,created_at,expires_at,revoked_at) VALUES($1,$2,$3,$4,$5)`
		created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		expiry := created.Add(8 * time.Hour)
		hash := make([]byte, 32)
		exec(t, conn, query, 101, hash, created, expiry, nil)
		expectState(t, conn, "23505", query, 101, hash, created, expiry, nil)
		other := make([]byte, 32)
		other[0] = 1
		expectState(t, conn, "23503", query, 999999, other, created, expiry, nil)
		for _, length := range []int{0, 31, 33} {
			expectState(t, conn, "23514", query, 101, make([]byte, length), created, expiry, nil)
		}
		for _, end := range []time.Time{created, created.Add(-time.Second)} {
			expectState(t, conn, "23514", query, 101, other, created, end, nil)
		}
		expectState(t, conn, "23514", query, 101, other, created, expiry, created.Add(-time.Second))
		expectState(t, conn, "23502", query, nil, other, created, expiry, nil)
		expectState(t, conn, "23502", query, 101, nil, created, expiry, nil)
		expectState(t, conn, "23502", query, 101, other, nil, expiry, nil)
		expectState(t, conn, "23502", query, 101, other, created, nil, nil)
		// Cấu trúc cho phép revoke đúng lúc tạo, nhiều phiên/user và hạn khác 8h.
		exec(t, conn, query, 101, other, created, created.Add(time.Hour), created)
		exec(t, conn, `INSERT INTO users(user_id,username,password_hash,system_role) OVERRIDING SYSTEM VALUE
 VALUES(105,'session-fk-only','synthetic-not-login','EMPLOYEE')`)
		exec(t, conn, query, 105, []byte(strings.Repeat("x", 32)), created, expiry, nil)
		expectState(t, conn, "23503", `DELETE FROM users WHERE user_id=105`)
		expectState(t, conn, "23505", `INSERT INTO auth_sessions(session_id,user_id,token_hash,created_at,expires_at)
 OVERRIDING SYSTEM VALUE SELECT session_id,101,decode(repeat('cc',32),'hex'),created_at,expires_at FROM auth_sessions LIMIT 1`)
		expectState(t, conn, "428C9", `INSERT INTO auth_sessions(session_id,user_id,token_hash,expires_at) VALUES(500,101,decode(repeat('ff',32),'hex'),now()+interval '1 hour')`)
		var positive bool
		authdb.Require(t, "identity/default", conn.QueryRow(context.Background(), `SELECT bool_and(session_id>0 AND created_at IS NOT NULL) FROM auth_sessions`).Scan(&positive))
		if !positive {
			t.Fatal("identity/default không hợp lệ")
		}
		exec(t, conn, `INSERT INTO auth_sessions(user_id,token_hash,expires_at) VALUES(102,decode(repeat('ee',32),'hex'),now()+interval '1 hour')`)
	})
	t.Run("metadata", func(t *testing.T) {
		var columns int
		authdb.Require(t, "số cột session", conn.QueryRow(context.Background(), `SELECT count(*) FROM information_schema.columns WHERE table_schema='public' AND table_name='auth_sessions'`).Scan(&columns))
		if columns != 6 {
			t.Fatal("schema session ngoài sáu cột đã chốt")
		}
		var hasIndex bool
		authdb.Require(t, "index user", conn.QueryRow(context.Background(), `SELECT EXISTS(
   SELECT 1 FROM pg_index i JOIN pg_class c ON c.oid=i.indexrelid
   JOIN pg_attribute a ON a.attrelid=i.indrelid AND a.attnum=i.indkey[0]
   WHERE i.indrelid='public.auth_sessions'::regclass AND c.relname='idx_auth_sessions_user'
   AND i.indisvalid AND i.indisready AND i.indnkeyatts=1 AND a.attname='user_id')`).Scan(&hasIndex))
		if !hasIndex {
			t.Fatal("index user_id không đúng")
		}
		var noAction int
		authdb.Require(t, "FK NO ACTION", conn.QueryRow(context.Background(), `SELECT count(*) FROM pg_constraint WHERE conrelid IN ('public.user_roles'::regclass,'public.auth_sessions'::regclass) AND contype='f' AND confrelid='public.users'::regclass AND confdeltype='a' AND confupdtype='a'`).Scan(&noAction))
		if noAction != 2 {
			t.Fatal("FK auth không đúng NO ACTION")
		}
		var flag bool
		authdb.Require(t, "cờ bool NN/default", conn.QueryRow(context.Background(), `SELECT data_type='boolean' AND is_nullable='NO' AND column_default='false' FROM information_schema.columns WHERE table_schema='public' AND table_name='users' AND column_name='must_change_password'`).Scan(&flag))
		if !flag {
			t.Fatal("cấu trúc cờ không đúng")
		}
	})
}

func TestAuthMigrationAtomicFailure(t *testing.T) {
	db := authdb.NewCase(t)
	m := db.OpenMigrator(t, migrations.FS)
	authdb.Require(t, "up version 2", m.Migrate(2))
	fixture(t, db)
	authdb.Require(t, "preflight failure fixture", legacyPreflight(db, true))
	before := snapshot(t, db)
	m.Close(t)
	// Nguồn thử nghiệm chỉ trong bộ nhớ; không sửa file migration production.
	source := fstest.MapFS{}
	entries, err := fs.ReadDir(migrations.FS, ".")
	authdb.Require(t, "đọc nguồn test", err)
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		content, err := fs.ReadFile(migrations.FS, entry.Name())
		authdb.Require(t, "đọc SQL", err)
		if entry.Name() == "000003_auth_roles_sessions.up.sql" {
			content = []byte(strings.Replace(string(content), "COMMIT;", "SELECT 1/0;\nCOMMIT;", 1))
		}
		source[entry.Name()] = &fstest.MapFile{Data: content, Mode: 0600}
	}
	broken := db.OpenMigrator(t, source)
	if broken.Up() == nil {
		t.Fatal("migration lỗi cố ý phải thất bại")
	}
	// Đóng connection kết thúc transaction lỗi trước khi xác minh rollback.
	broken.Close(t)
	var version uint
	var dirty bool
	authdb.Require(t, "dirty sau lỗi thực", db.Conn.QueryRow(context.Background(), `SELECT version,dirty FROM schema_migrations`).Scan(&version, &dirty))
	if version != 3 || !dirty {
		t.Fatal("driver phải giữ version lỗi dirty")
	}
	var absent bool
	authdb.Require(t, "rollback DDL", db.Conn.QueryRow(context.Background(), `SELECT to_regclass('public.user_roles') IS NULL AND to_regclass('public.auth_sessions') IS NULL`).Scan(&absent))
	if !absent {
		t.Fatal("DDL/backfill không rollback toàn bộ")
	}
	sameJSON(t, before, snapshot(t, db))
	retry := db.OpenMigrator(t, migrations.FS)
	var dirtyErr migrate.ErrDirty
	if !errors.As(retry.Up(), &dirtyErr) {
		t.Fatal("runner phải từ chối DB dirty; không force version")
	}
}

// Preflight này dành cho rehearsal local; không chứng nhận dữ liệu Neon hiện hành.
func legacyPreflight(db *authdb.Case, flag bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var clean bool
	err := db.Conn.QueryRow(ctx, `SELECT
 to_regclass('public.user_roles') IS NULL AND to_regclass('public.auth_sessions') IS NULL
 AND NOT EXISTS(SELECT 1 FROM users u LEFT JOIN employees e ON e.user_id=u.user_id WHERE e.user_id IS NULL)
 AND NOT EXISTS(SELECT 1 FROM employees e LEFT JOIN users u ON u.user_id=e.user_id WHERE u.user_id IS NULL)
 AND NOT EXISTS(SELECT user_id FROM employees GROUP BY user_id HAVING count(*)>1)
 AND NOT EXISTS(SELECT lower(btrim(email)) FROM employees GROUP BY lower(btrim(email)) HAVING count(*)>1)
 AND NOT EXISTS(SELECT 1 FROM users WHERE system_role::text NOT IN ('ADMIN','PM','EMPLOYEE'))
 AND (SELECT array_agg(enumlabel::text ORDER BY enumsortorder) FROM pg_enum WHERE enumtypid='user_system_role'::regtype)=ARRAY['ADMIN','PM','EMPLOYEE']
 AND (SELECT array_agg(enumlabel::text ORDER BY enumsortorder) FROM pg_enum WHERE enumtypid='user_status'::regtype)=ARRAY['ACTIVE','LOCKED','DISABLED']
 AND (SELECT array_agg(enumlabel::text ORDER BY enumsortorder) FROM pg_enum WHERE enumtypid='employment_status'::regtype)=ARRAY['ACTIVE','INACTIVE','RESIGNED']
 AND (SELECT count(*) FROM (VALUES
 ('users','user_id','int4','NO'),('users','username','varchar','NO'),('users','password_hash','varchar','NO'),
 ('users','system_role','user_system_role','NO'),('users','status','user_status','NO'),
 ('employees','employee_id','int4','NO'),('employees','user_id','int4','NO'),('employees','email','varchar','NO'),
 ('employees','employment_status','employment_status','NO'),('employees','position_id','int4','NO'),('employees','department_id','int4','YES'),
 ('positions','position_id','int4','NO'),('departments','department_id','int4','NO')) AS expected(t,c,typ,nullable)
 JOIN information_schema.columns a ON a.table_schema='public' AND a.table_name=expected.t AND a.column_name=expected.c
 AND a.udt_name=expected.typ AND a.is_nullable=expected.nullable)=13
 AND (SELECT count(*) FROM pg_constraint WHERE conrelid IN ('users'::regclass,'employees'::regclass,'positions'::regclass,'departments'::regclass) AND contype='p')=4
 AND EXISTS(SELECT 1 FROM pg_constraint WHERE conrelid='employees'::regclass AND contype='u' AND conkey=ARRAY[(SELECT attnum FROM pg_attribute WHERE attrelid='employees'::regclass AND attname='user_id')]::smallint[])
 AND EXISTS(SELECT 1 FROM pg_constraint WHERE conrelid='employees'::regclass AND contype='u' AND conkey=ARRAY[(SELECT attnum FROM pg_attribute WHERE attrelid='employees'::regclass AND attname='email')]::smallint[])
 AND (SELECT count(*) FROM pg_constraint WHERE conrelid='employees'::regclass AND contype='f' AND confrelid IN ('users'::regclass,'positions'::regclass,'departments'::regclass))=3
 AND (SELECT count(*) FROM information_schema.columns WHERE table_schema='public' AND table_name='users' AND column_name='must_change_password')=$1
 AND (SELECT count(*) FROM information_schema.columns WHERE table_schema='public' AND table_name='users' AND column_name='must_change_password' AND data_type='boolean' AND is_nullable='NO' AND column_default='false')=$1
 AND NOT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name='users' AND column_name='temporary_password_expires_at')`, map[bool]int{false: 0, true: 1}[flag]).Scan(&clean)
	if err != nil || !clean {
		return errors.New("preflight legacy không đạt; không xuất SQL/dữ liệu")
	}
	return nil
}

func TestAuthMigrationPreflightRejectsDrift(t *testing.T) {
	cases := []struct{ name, sql string }{
		{"missing employee", `DELETE FROM employees WHERE user_id=103`},
		{"normalized email collision", `UPDATE employees SET email=' FIXTURE-ADMIN@EXAMPLE.INVALID ' WHERE user_id=102`},
		{"missing unique link", `ALTER TABLE employees DROP CONSTRAINT employees_user_id_key`},
		{"preexisting auth table", `CREATE TABLE user_roles(dummy int)`},
		{"wrong column type", `ALTER TABLE users ALTER COLUMN username TYPE text`},
		{"wrong legacy enum", `ALTER TYPE user_system_role ADD VALUE 'HR'`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := authdb.NewCase(t)
			m := db.OpenMigrator(t, migrations.FS)
			authdb.Require(t, "up version 2", m.Migrate(2))
			fixture(t, db)
			exec(t, db.Conn, tc.sql)
			if legacyPreflight(db, true) == nil {
				t.Fatal("preflight phải từ chối drift/mâu thuẫn")
			}
			head(t, m, 2)
		})
	}
}

func TestAuthMigrationRootRemainsEmpty(t *testing.T) {
	authdb.AssertRootEmpty(t)
}
