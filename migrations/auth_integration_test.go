//go:build integration

// Các ca này chỉ được chạy qua scripts/authdb/manage.py run trên cluster test sở hữu.
package migrations_test

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"net/url"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"ai-staffing/migrations"
	"github.com/golang-migrate/migrate/v4"
	driver "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/bcrypt"
)

const baseDB = "ai_staffing_auth_test"
const dbUser = "ai_staffing_test"

// Không in err: pgx/migrate có thể kèm SQL, hash hoặc thông tin kết nối.
func must(t *testing.T, stage string, err error) {
	t.Helper()
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			t.Fatalf("%s: SQLSTATE %s (không xuất lỗi thô)", stage, pgErr.Code)
		}
		t.Fatalf("%s: không thành công (không xuất lỗi thô)", stage)
	}
}

func rootConfig(t *testing.T) *pgx.ConnConfig {
	t.Helper()
	raw := os.Getenv("TEST_DATABASE_URL")
	u, err := url.Parse(raw)
	must(t, "URL test", err)
	if raw == "" || u.Scheme != "postgresql" || u.Hostname() != "127.0.0.1" || u.Port() != "55438" ||
		u.User == nil || u.User.Username() != dbUser || u.Path != "/"+baseDB || u.RawQuery != "sslmode=disable" || u.Fragment != "" {
		t.Fatal("thiếu hoặc sai TEST_DATABASE_URL; không fallback DB khác")
	}
	cfg, err := pgx.ParseConfig(raw)
	must(t, "parse config test", err)
	cfg.ConnectTimeout = 5 * time.Second
	return cfg
}

func verifyCluster(t *testing.T, conn *pgx.Conn, name string) {
	t.Helper()
	expected := os.Getenv("AUTH_TEST_DB_SYSTEM_ID")
	if expected == "" {
		t.Fatal("thiếu cluster ID được runner xác minh")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var major int
	var database, user, system string
	var owner bool
	err := conn.QueryRow(ctx, `SELECT current_setting('server_version_num')::int/10000,
 current_database(),current_user,pg_get_userbyid(datdba)=current_user,
 (SELECT system_identifier::text FROM pg_control_system())
 FROM pg_database WHERE datname=current_database()`).Scan(&major, &database, &user, &owner, &system)
	must(t, "metadata cluster", err)
	if major != 18 || database != name || user != dbUser || !owner || system != expected {
		t.Fatal("đích/owner/cluster test không khớp; từ chối DDL/dọn")
	}
}

type caseDB struct {
	cfg  *pgx.ConnConfig
	conn *pgx.Conn
}

func newCaseDB(t *testing.T) *caseDB {
	t.Helper()
	cfg := rootConfig(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	root, err := pgx.ConnectConfig(ctx, cfg)
	must(t, "kết nối DB test gốc", err)
	defer root.Close(context.Background())
	verifyCluster(t, root, baseDB)
	var random [8]byte
	_, err = rand.Read(random[:])
	must(t, "sinh tên DB test", err)
	name := "auth_case_" + hex.EncodeToString(random[:])
	_, err = root.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{name}.Sanitize()+" TEMPLATE template0")
	must(t, "tạo DB case test", err)
	// Chỉ đăng ký tên vừa CREATE thành công; không nhận DB có sẵn để dọn.
	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		admin, err := pgx.ConnectConfig(cleanup, cfg)
		if err != nil {
			t.Error("không nối được để dọn DB case")
			return
		}
		defer admin.Close(cleanup)
		verifyCluster(t, admin, baseDB)
		var owned bool
		err = admin.QueryRow(cleanup, `SELECT pg_get_userbyid(datdba)=current_user FROM pg_database WHERE datname=$1`, name).Scan(&owned)
		if err != nil || !owned {
			t.Error("DB case không đúng owner; không xóa")
			return
		}
		_, err = admin.Exec(cleanup, "DROP DATABASE "+pgx.Identifier{name}.Sanitize())
		if err != nil {
			t.Error("không dọn được DB case; giữ để kiểm tra, không FORCE/prune")
		}
	})
	child := cfg.Copy()
	child.Database = name
	conn, err := pgx.ConnectConfig(ctx, child)
	must(t, "kết nối DB case", err)
	t.Cleanup(func() { conn.Close(context.Background()) })
	verifyCluster(t, conn, name)
	return &caseDB{child, conn}
}

type migrationEngine = migrate.Migrate

type runner struct {
	*migrationEngine
	closeOnce sync.Once
}

func (r *runner) close(t *testing.T) {
	t.Helper()
	r.closeOnce.Do(func() {
		a, b := r.migrationEngine.Close()
		if a != nil || b != nil {
			t.Error("đóng migration runner thất bại")
		}
	})
}

func migrator(t *testing.T, db *caseDB, source fs.FS) *runner {
	t.Helper()
	// Dùng đúng driver hiện hành; không loader/CLI dotenv, không SQL version thủ công.
	sqlDB := stdlibDB(t, db.cfg)
	d, err := driver.WithInstance(sqlDB, &driver.Config{MultiStatementEnabled: false, StatementTimeout: 35 * time.Second})
	if err != nil {
		sqlDB.Close()
	}
	must(t, "driver migration", err)
	src, err := iofs.New(source, ".")
	must(t, "nguồn migration", err)
	m, err := migrate.NewWithInstance("iofs", src, "pgx", d)
	must(t, "runner migration", err)
	r := &runner{migrationEngine: m}
	t.Cleanup(func() { r.close(t) })
	return r
}

func stdlibDB(t *testing.T, cfg *pgx.ConnConfig) *sql.DB {
	t.Helper()
	// ConnString() giữ URL gốc, không phản ánh cfg.Database đã đổi cho từng case.
	// OpenDB dùng trực tiếp config đã xác minh để không migrate nhầm DB gốc test.
	db := stdlib.OpenDB(*cfg)
	var name string
	err := db.QueryRow("SELECT current_database()").Scan(&name)
	if err != nil || name != cfg.Database {
		db.Close()
		t.Fatal("driver migration không đúng database case")
	}
	return db
}

func head(t *testing.T, m *runner, expected uint) {
	t.Helper()
	v, dirty, err := m.Version()
	must(t, "đọc version", err)
	if v != expected || dirty {
		t.Fatal("version/dirty không đúng kỳ vọng")
	}
}

func exec(t *testing.T, conn *pgx.Conn, query string, args ...any) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := conn.Exec(ctx, query, args...)
	must(t, "fixture/assert SQL", err)
}

func fixture(t *testing.T, db *caseDB) {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("Only synthetic local fixture 2026"), bcrypt.DefaultCost)
	must(t, "hash fixture", err)
	exec(t, db.conn, `INSERT INTO positions(position_name) VALUES ('Fixture position');
 INSERT INTO departments(department_name) VALUES ('Fixture department')`)
	for i, role := range []string{"ADMIN", "PM", "EMPLOYEE"} {
		// ID khác employee ID; toàn bộ dữ liệu giả, không dùng seed ứng dụng.
		userID := int64(101 + i)
		employeeID := int64(201 + i)
		exec(t, db.conn, `INSERT INTO users(user_id,username,password_hash,system_role,status,created_at,updated_at)
   OVERRIDING SYSTEM VALUE VALUES($1,$2,$3,$4,$5,'2026-01-01Z','2026-01-02Z')`, userID, "fixture-"+role, string(hash), role, []string{"ACTIVE", "LOCKED", "DISABLED"}[i])
		exec(t, db.conn, `INSERT INTO employees(employee_id,user_id,employee_code,fullname,email,position_id,department_id,employment_status,created_at,updated_at)
   OVERRIDING SYSTEM VALUE VALUES($1,$2,$3,$4,$5,1,$6,$7,'2026-01-01Z','2026-01-02Z')`, employeeID, userID, "FIX-"+role, "Fixture "+role, "fixture-"+role+"@example.invalid", []any{nil, 1, nil}[i], []string{"ACTIVE", "INACTIVE", "RESIGNED"}[i])
	}
}

func snapshot(t *testing.T, db *caseDB) string {
	t.Helper()
	var value string
	err := db.conn.QueryRow(context.Background(), `SELECT json_build_object(
 'users',(SELECT json_agg(row_to_json(u) ORDER BY user_id) FROM users u),
 'employees',(SELECT json_agg(row_to_json(e) ORDER BY employee_id) FROM employees e),
 'positions',(SELECT json_agg(row_to_json(p) ORDER BY position_id) FROM positions p),
 'departments',(SELECT json_agg(row_to_json(d) ORDER BY department_id) FROM departments d))::text`).Scan(&value)
	must(t, "snapshot nội bộ", err)
	return value
}

func roles(t *testing.T, db *caseDB) {
	t.Helper()
	var rows string
	err := db.conn.QueryRow(context.Background(), `SELECT coalesce(string_agg(user_id::text||':'||role_code,',' ORDER BY user_id,role_code),'') FROM user_roles`).Scan(&rows)
	must(t, "mapping quyền", err)
	if rows != "101:ADMIN" {
		t.Fatal("mapping legacy không đúng; không xuất dữ liệu")
	}
}

func TestAuthMigrationEmpty(t *testing.T) {
	db := newCaseDB(t)
	m := migrator(t, db, migrations.FS)
	must(t, "up từ DB rỗng", m.Up())
	head(t, m, 3)
	var count int
	must(t, "đếm nền auth", db.conn.QueryRow(context.Background(), `SELECT count(*) FROM information_schema.tables WHERE table_schema='public' AND table_name IN ('user_roles','auth_sessions')`).Scan(&count))
	if count != 2 {
		t.Fatal("thiếu bảng auth")
	}
	if !errors.Is(m.Up(), migrate.ErrNoChange) {
		t.Fatal("up lại phải không thay đổi")
	}
}

func TestAuthMigrationLegacy(t *testing.T) {
	db := newCaseDB(t)
	m := migrator(t, db, migrations.FS)
	must(t, "up version 1", m.Migrate(1))
	fixture(t, db)
	must(t, "preflight legacy v1", legacyPreflight(db, false))
	before := snapshot(t, db)
	must(t, "up version 2", m.Migrate(2))
	head(t, m, 2)
	var count int
	must(t, "cờ cũ false", db.conn.QueryRow(context.Background(), `SELECT count(*) FROM users WHERE must_change_password`).Scan(&count))
	if count != 0 {
		t.Fatal("000002 không giữ tài khoản cũ false")
	}
	// Bỏ đúng cột mới khi so snapshot toàn bộ hàng trước 000002.
	var usersAfter string
	must(t, "snapshot cũ sau nâng", db.conn.QueryRow(context.Background(), `SELECT json_build_object(
 'users',(SELECT json_agg(to_jsonb(u)-'must_change_password' ORDER BY user_id) FROM users u),
 'employees',(SELECT json_agg(row_to_json(e) ORDER BY employee_id) FROM employees e),
 'positions',(SELECT json_agg(row_to_json(p) ORDER BY position_id) FROM positions p),
 'departments',(SELECT json_agg(row_to_json(d) ORDER BY department_id) FROM departments d))::text`).Scan(&usersAfter))
	sameJSON(t, before, usersAfter)
	must(t, "preflight legacy v2", legacyPreflight(db, true))
	baseline := snapshot(t, db)
	must(t, "up auth", m.Up())
	head(t, m, 3)
	sameJSON(t, baseline, snapshot(t, db))
	roles(t, db)
	constraints(t, db)
}

func TestAuthMigrationPreservesFlagsAndRollback(t *testing.T) {
	db := newCaseDB(t)
	m := migrator(t, db, migrations.FS)
	must(t, "up version 2", m.Migrate(2))
	fixture(t, db)
	exec(t, db.conn, `UPDATE users SET must_change_password=true WHERE user_id=101`)
	must(t, "preflight trước auth", legacyPreflight(db, true))
	before := snapshot(t, db)
	must(t, "up auth", m.Up())
	head(t, m, 3)
	sameJSON(t, before, snapshot(t, db))
	roles(t, db)
	// Quyền mới và phiên chỉ nạp sau khi bảng đã tồn tại.
	exec(t, db.conn, `INSERT INTO user_roles VALUES(102,'HR'),(102,'DIRECTOR');
 INSERT INTO auth_sessions(user_id,token_hash,created_at,expires_at) VALUES(101,decode(repeat('ab',32),'hex'),'2026-01-01Z','2026-01-01 08:00Z')`)
	must(t, "down auth trên bản test", m.Steps(-1))
	head(t, m, 2)
	sameJSON(t, before, snapshot(t, db))
	var absent bool
	must(t, "bảng auth đã down", db.conn.QueryRow(context.Background(), `SELECT to_regclass('public.user_roles') IS NULL AND to_regclass('public.auth_sessions') IS NULL`).Scan(&absent))
	if !absent {
		t.Fatal("down chưa xóa đúng bảng auth")
	}
	must(t, "up auth lại", m.Up())
	head(t, m, 3)
	roles(t, db)
	var sessions int
	must(t, "phiên không phục hồi giả", db.conn.QueryRow(context.Background(), `SELECT count(*) FROM auth_sessions`).Scan(&sessions))
	if sessions != 0 {
		t.Fatal("down/up không được giả phục hồi phiên")
	}
	sameJSON(t, before, snapshot(t, db))
}

func sameJSON(t *testing.T, a, b string) {
	t.Helper()
	var left, right any
	must(t, "decode snapshot trước", json.Unmarshal([]byte(a), &left))
	must(t, "decode snapshot sau", json.Unmarshal([]byte(b), &right))
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

func constraints(t *testing.T, db *caseDB) {
	t.Helper()
	conn := db.conn
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
		must(t, "identity/default", conn.QueryRow(context.Background(), `SELECT bool_and(session_id>0 AND created_at IS NOT NULL) FROM auth_sessions`).Scan(&positive))
		if !positive {
			t.Fatal("identity/default không hợp lệ")
		}
		exec(t, conn, `INSERT INTO auth_sessions(user_id,token_hash,expires_at) VALUES(102,decode(repeat('ee',32),'hex'),now()+interval '1 hour')`)
	})
	t.Run("metadata", func(t *testing.T) {
		var columns int
		must(t, "số cột session", conn.QueryRow(context.Background(), `SELECT count(*) FROM information_schema.columns WHERE table_schema='public' AND table_name='auth_sessions'`).Scan(&columns))
		if columns != 6 {
			t.Fatal("schema session ngoài sáu cột đã chốt")
		}
		var hasIndex bool
		must(t, "index user", conn.QueryRow(context.Background(), `SELECT EXISTS(
   SELECT 1 FROM pg_index i JOIN pg_class c ON c.oid=i.indexrelid
   JOIN pg_attribute a ON a.attrelid=i.indrelid AND a.attnum=i.indkey[0]
   WHERE i.indrelid='public.auth_sessions'::regclass AND c.relname='idx_auth_sessions_user'
   AND i.indisvalid AND i.indisready AND i.indnkeyatts=1 AND a.attname='user_id')`).Scan(&hasIndex))
		if !hasIndex {
			t.Fatal("index user_id không đúng")
		}
		var noAction int
		must(t, "FK NO ACTION", conn.QueryRow(context.Background(), `SELECT count(*) FROM pg_constraint WHERE conrelid IN ('public.user_roles'::regclass,'public.auth_sessions'::regclass) AND contype='f' AND confrelid='public.users'::regclass AND confdeltype='a' AND confupdtype='a'`).Scan(&noAction))
		if noAction != 2 {
			t.Fatal("FK auth không đúng NO ACTION")
		}
		var flag bool
		must(t, "cờ bool NN/default", conn.QueryRow(context.Background(), `SELECT data_type='boolean' AND is_nullable='NO' AND column_default='false' FROM information_schema.columns WHERE table_schema='public' AND table_name='users' AND column_name='must_change_password'`).Scan(&flag))
		if !flag {
			t.Fatal("cấu trúc cờ không đúng")
		}
	})
}

func TestAuthMigrationAtomicFailure(t *testing.T) {
	db := newCaseDB(t)
	m := migrator(t, db, migrations.FS)
	must(t, "up version 2", m.Migrate(2))
	fixture(t, db)
	must(t, "preflight failure fixture", legacyPreflight(db, true))
	before := snapshot(t, db)
	m.close(t)
	// Nguồn thử nghiệm chỉ trong bộ nhớ; không sửa file migration production.
	source := fstest.MapFS{}
	entries, err := fs.ReadDir(migrations.FS, ".")
	must(t, "đọc nguồn test", err)
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		content, err := fs.ReadFile(migrations.FS, entry.Name())
		must(t, "đọc SQL", err)
		if entry.Name() == "000003_auth_roles_sessions.up.sql" {
			content = []byte(strings.Replace(string(content), "COMMIT;", "SELECT 1/0;\nCOMMIT;", 1))
		}
		source[entry.Name()] = &fstest.MapFile{Data: content, Mode: 0600}
	}
	broken := migrator(t, db, source)
	if broken.Up() == nil {
		t.Fatal("migration lỗi cố ý phải thất bại")
	}
	// Đóng connection kết thúc transaction lỗi trước khi xác minh rollback.
	broken.close(t)
	var version uint
	var dirty bool
	must(t, "dirty sau lỗi thực", db.conn.QueryRow(context.Background(), `SELECT version,dirty FROM schema_migrations`).Scan(&version, &dirty))
	if version != 3 || !dirty {
		t.Fatal("driver phải giữ version lỗi dirty")
	}
	var absent bool
	must(t, "rollback DDL", db.conn.QueryRow(context.Background(), `SELECT to_regclass('public.user_roles') IS NULL AND to_regclass('public.auth_sessions') IS NULL`).Scan(&absent))
	if !absent {
		t.Fatal("DDL/backfill không rollback toàn bộ")
	}
	sameJSON(t, before, snapshot(t, db))
	retry := migrator(t, db, migrations.FS)
	var dirtyErr migrate.ErrDirty
	if !errors.As(retry.Up(), &dirtyErr) {
		t.Fatal("runner phải từ chối DB dirty; không force version")
	}
}

// Preflight này dành cho rehearsal local; không chứng nhận dữ liệu Neon hiện hành.
func legacyPreflight(db *caseDB, flag bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var clean bool
	err := db.conn.QueryRow(ctx, `SELECT
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
			db := newCaseDB(t)
			m := migrator(t, db, migrations.FS)
			must(t, "up version 2", m.Migrate(2))
			fixture(t, db)
			exec(t, db.conn, tc.sql)
			if legacyPreflight(db, true) == nil {
				t.Fatal("preflight phải từ chối drift/mâu thuẫn")
			}
			head(t, m, 2)
		})
	}
}

func TestAuthMigrationRootRemainsEmpty(t *testing.T) {
	cfg := rootConfig(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := pgx.ConnectConfig(ctx, cfg)
	must(t, "probe DB gốc", err)
	defer conn.Close(ctx)
	verifyCluster(t, conn, baseDB)
	var tables, cases int
	must(t, "DB gốc không bị migrate", conn.QueryRow(ctx, `SELECT count(*) FROM pg_tables WHERE schemaname='public'`).Scan(&tables))
	must(t, "DB case đã dọn", conn.QueryRow(ctx, `SELECT count(*) FROM pg_database WHERE datname LIKE 'auth_case_%'`).Scan(&cases))
	if tables != 0 || cases != 0 {
		t.Fatal("DB gốc phải trống và không còn DB case; không tự dọn tài nguyên không thuộc lần chạy")
	}
}
