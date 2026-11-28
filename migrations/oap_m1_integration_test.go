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

	"ai-staffing/db/seed"
	"ai-staffing/internal/testutil/authdb"
	"ai-staffing/migrations"
	"github.com/golang-migrate/migrate/v4"
)

func TestM1EmptyAndReferenceSeed(t *testing.T) {
	db := authdb.NewCase(t)
	m := db.OpenMigrator(t, migrations.FS)
	authdb.Require(t, "bootstrap đến auth", m.Migrate(3))
	authdb.Require(t, "preflight DB rỗng", db.CheckM1(t, nil))
	authdb.Require(t, "nâng M1", m.Migrate(migrations.M1Version))
	head(t, m, migrations.M1Version)
	if !errors.Is(m.Migrate(migrations.M1Version), migrate.ErrNoChange) {
		t.Fatal("up lại phải không thay đổi")
	}
	for _, table := range []struct{ name, column string }{
		{"departments", "department_name"}, {"positions", "position_name"}, {"job_levels", "job_level_name"},
	} {
		t.Run(table.name, func(t *testing.T) {
			insert := fmt.Sprintf("INSERT INTO public.%s (%s) VALUES ($1)", table.name, table.column)
			exec(t, db.Conn, insert, "Kỹ thuật")
			expectState(t, db.Conn, "23505", insert, "  KỸ THUẬT  ")
			expectState(t, db.Conn, "23514", insert, "\t  ")
			expectState(t, db.Conn, "23514", insert, "\u00a0")
		})
	}
	// Seed năng lực đi theo head; mốc M1 chỉ kiểm seed danh mục còn tương thích.
	content, err := fs.ReadFile(seed.FS, "01_reference.sql")
	authdb.Require(t, "đọc seed danh mục", err)
	for range 2 {
		exec(t, db.Conn, string(content))
	}
}

func TestM1LegacyAndDownUp(t *testing.T) {
	db := authdb.NewCase(t)
	m := db.OpenMigrator(t, migrations.FS)
	authdb.Require(t, "nền legacy v3", m.Migrate(3))
	fixture(t, db)
	exec(t, db.Conn, `UPDATE employees SET seniority_level = CASE employee_id WHEN 201 THEN 'Fixture legacy A' WHEN 202 THEN U&'\00A0' END;
		UPDATE users SET must_change_password=true WHERE user_id=101;
		INSERT INTO user_roles VALUES(101,'HR'),(101,'DIRECTOR');
		INSERT INTO auth_sessions(user_id,token_hash,created_at,expires_at) VALUES
		(101,decode(repeat('ab',32),'hex'),'2026-01-01Z','2026-01-01 08:00Z')`)
	before := m1LegacySnapshot(t, db)
	m.Close(t)

	approved := map[string]string{"Fixture legacy A": "Cấp bậc kiểm thử"}
	authdb.Require(t, "preflight mapping", db.CheckM1(t, approved))
	m = db.OpenMappedMigrator(t, migrations.FS, approved)
	authdb.Require(t, "nâng legacy lên M1", m.Migrate(migrations.M1Version))
	head(t, m, migrations.M1Version)
	sameJSON(t, before, m1LegacySnapshot(t, db))

	assertMapped := func() {
		t.Helper()
		var ok bool
		authdb.Require(t, "đọc mapping", db.Conn.QueryRow(context.Background(), `SELECT
			(SELECT count(*) FROM job_levels) = 1 AND
			(SELECT j.job_level_name FROM employees e JOIN job_levels j USING (job_level_id) WHERE e.employee_id=201) = 'Cấp bậc kiểm thử'
			AND (SELECT count(*) FROM employees WHERE job_level_id IS NULL) = 2
			AND NOT EXISTS (SELECT 1 FROM employees WHERE address IS NOT NULL)`).Scan(&ok))
		if !ok {
			t.Fatal("mapping cấp bậc hoặc địa chỉ NULL sai")
		}
	}
	assertMapped()
	expectState(t, db.Conn, "23503", `UPDATE employees SET job_level_id=-1 WHERE employee_id=201`)
	expectState(t, db.Conn, "23503", `DELETE FROM job_levels`)
	expectState(t, db.Conn, "23514", `UPDATE employees SET address=E' \t' WHERE employee_id=201`)
	expectState(t, db.Conn, "23503", `INSERT INTO audit_events(actor_id,entity_type,entity_id,operation) VALUES(-1,'department',1,'CREATE')`)
	exec(t, db.Conn, `INSERT INTO audit_events(actor_id,entity_type,entity_id,operation,after_values)
		VALUES(101,'department',1,'CREATE','{"schema_version":1}');
		ALTER TABLE employees DISABLE TRIGGER trg_employees_updated_at;
		UPDATE employees SET address='Synthetic test address' WHERE employee_id=201;
		ALTER TABLE employees ENABLE TRIGGER trg_employees_updated_at`)
	var schemaOK bool
	authdb.Require(t, "schema audit và index M1", db.Conn.QueryRow(context.Background(), `SELECT
		to_regclass('public.idx_employees_job_level') IS NOT NULL
		AND to_regclass('public.idx_audit_events_entity') IS NOT NULL
		AND (SELECT audit_id > 0 AND occurred_at IS NOT NULL AND reason IS NULL AND before_values IS NULL
			AND after_values->>'schema_version'='1' FROM audit_events)`).Scan(&schemaOK))
	if !schemaOK {
		t.Fatal("schema audit hoặc index M1 sai")
	}
	authdb.Require(t, "down M1 trên case hy sinh", m.Steps(-1))
	head(t, m, 3)
	assertM1Absent(t, db)
	sameJSON(t, before, m1LegacySnapshot(t, db))

	authdb.Require(t, "preflight sau down", db.CheckM1(t, approved))
	authdb.Require(t, "up M1 lại", m.Migrate(migrations.M1Version))
	head(t, m, migrations.M1Version)
	assertMapped()
	var count int
	authdb.Require(t, "kiểm dữ liệu audit không tự khôi phục", db.Conn.QueryRow(context.Background(), "SELECT count(*) FROM audit_events").Scan(&count))
	if count != 0 {
		t.Fatal("audit phải trống sau down/up")
	}
	sameJSON(t, before, m1LegacySnapshot(t, db))
}

func TestM1PreflightRejectsBeforeMigration(t *testing.T) {
	for _, test := range []struct{ name, sql string }{
		{"catalog duplicate", `INSERT INTO departments(department_name) VALUES ('Fixture'), (' fixture ')`},
		{"blank catalog", `INSERT INTO positions(position_name) VALUES (E'\t ')`},
		{"unicode padding", `INSERT INTO positions(position_name) VALUES (U&'\00A0Fixture')`},
		{"mapping missing", `UPDATE employees SET seniority_level='Unreviewed test level' WHERE employee_id=201`},
		{"unexpected catalog", `CREATE TABLE job_levels (dummy int)`},
		{"foreign key drift", `ALTER TABLE employees DROP CONSTRAINT employees_position_id_fkey;
			ALTER TABLE employees ADD FOREIGN KEY (position_id) REFERENCES positions(position_id) ON DELETE CASCADE`},
		{"column drift", `ALTER TABLE positions ALTER COLUMN position_name TYPE text`},
	} {
		t.Run(test.name, func(t *testing.T) {
			db := authdb.NewCase(t)
			m := db.OpenMigrator(t, migrations.FS)
			authdb.Require(t, "nền v3", m.Migrate(3))
			fixture(t, db)
			exec(t, db.Conn, test.sql)
			if err := m.Migrate(migrations.M1Version); err == nil || !strings.HasPrefix(err.Error(), "M1:") {
				t.Fatal("runner phải từ chối tại preflight M1")
			}
			head(t, m, 3)
			var absent bool
			authdb.Require(t, "kiểm preflight không ghi audit", db.Conn.QueryRow(context.Background(),
				"SELECT to_regclass('public.audit_events') IS NULL").Scan(&absent))
			if !absent {
				t.Fatal("preflight không được ghi DDL")
			}
		})
	}
}

func TestM1DDLFailureRollsBackAndLeavesDirty(t *testing.T) {
	db := authdb.NewCase(t)
	m := db.OpenMigrator(t, migrations.FS)
	authdb.Require(t, "nền v3", m.Migrate(3))
	fixture(t, db)
	before := m1LegacySnapshot(t, db)
	m.Close(t)
	authdb.Require(t, "preflight trước lỗi DDL", db.CheckM1(t, nil))

	source := fstest.MapFS{}
	entries, err := fs.ReadDir(migrations.FS, ".")
	authdb.Require(t, "đọc nguồn migration", err)
	for _, entry := range entries {
		content, err := fs.ReadFile(migrations.FS, entry.Name())
		authdb.Require(t, "đọc SQL migration", err)
		if entry.Name() == "000004_oap_m1.up.sql" {
			content = []byte(strings.Replace(string(content), "COMMIT;", "CREATE TABLE public.job_levels (dummy int);\nCOMMIT;", 1))
		}
		source[entry.Name()] = &fstest.MapFile{Data: content}
	}
	m = db.OpenMigrator(t, source)
	if m.Migrate(migrations.M1Version) == nil {
		t.Fatal("DDL lỗi cố ý phải thất bại")
	}
	m.Close(t)
	assertM1Absent(t, db)
	sameJSON(t, before, m1LegacySnapshot(t, db))
	m = db.OpenMigrator(t, migrations.FS)
	var dirty migrate.ErrDirty
	if !errors.As(m.Migrate(migrations.M1Version), &dirty) || dirty.Version != migrations.M1Version {
		t.Fatal("driver phải giữ version M1 dirty và từ chối retry")
	}
}

func m1LegacySnapshot(t *testing.T, db *authdb.Case) string {
	t.Helper()
	var snapshot string
	authdb.Require(t, "snapshot bảo toàn nội bộ", db.Conn.QueryRow(context.Background(), `SELECT json_build_object(
		'users',(SELECT jsonb_agg(to_jsonb(u) ORDER BY user_id) FROM users u),
		'employees',(SELECT jsonb_agg(to_jsonb(e)-'job_level_id'-'address' ORDER BY employee_id) FROM employees e),
		'roles',(SELECT jsonb_agg(to_jsonb(r) ORDER BY user_id,role_code) FROM user_roles r),
		'sessions',(SELECT jsonb_agg(to_jsonb(s) ORDER BY session_id) FROM auth_sessions s),
		'departments',(SELECT jsonb_agg(to_jsonb(d) ORDER BY department_id) FROM departments d),
		'positions',(SELECT jsonb_agg(to_jsonb(p) ORDER BY position_id) FROM positions p))::text`).Scan(&snapshot))
	return snapshot
}

func assertM1Absent(t *testing.T, db *authdb.Case) {
	t.Helper()
	var absent bool
	authdb.Require(t, "kiểm cấu trúc sau rollback", db.Conn.QueryRow(context.Background(), `SELECT
		to_regclass('public.job_levels') IS NULL AND to_regclass('public.audit_events') IS NULL
		AND to_regclass('public.uq_departments_normalized_name') IS NULL
		AND to_regclass('public.uq_positions_normalized_name') IS NULL
		AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='public'
			AND table_name='employees' AND column_name IN ('address','job_level_id'))`).Scan(&absent))
	if !absent {
		t.Fatal("còn cấu trúc M1 sau rollback")
	}
}
