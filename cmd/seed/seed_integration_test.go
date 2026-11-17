//go:build integration

package main

import (
	"context"
	"testing"
	"time"

	"ai-staffing/internal/testutil/authdb"
	"ai-staffing/migrations"
)

func seedCase(t *testing.T) *authdb.Case {
	t.Helper()
	db := authdb.NewCase(t)
	m := db.OpenMigrator(t, migrations.FS)
	authdb.Require(t, "migrate schema seed đích", m.Up())
	m.Close(t)
	return db
}

func seedSynthetic(t *testing.T, db *authdb.Case, hash string) {
	t.Helper()
	tx, err := db.Conn.Begin(t.Context())
	authdb.Require(t, "begin seed", err)
	defer tx.Rollback(context.Background())
	authdb.Require(t, "seed danh mục", execSQLFile(t.Context(), tx, "01_reference.sql"))
	authdb.Require(t, "seed user và employee", seedPeople(t.Context(), tx, hash))
	authdb.Require(t, "seed kỹ năng theo email", execSQLFile(t.Context(), tx, "02_employee_skills.sql"))
	authdb.Require(t, "commit seed", tx.Commit(t.Context()))
}

func TestSeedHeadIsIdempotentAndPreservesExisting(t *testing.T) {
	db := seedCase(t)
	seedSynthetic(t, db, "synthetic hash first")
	// Mô phỏng người dùng đã đổi credential/quyền/trạng thái trước lượt seed sau.
	_, err := db.Conn.Exec(t.Context(), `UPDATE users SET password_hash='synthetic hash changed',status='LOCKED',must_change_password=true
 WHERE user_id=(SELECT user_id FROM employees WHERE email='admin@company.vn');
 DELETE FROM user_roles WHERE user_id=(SELECT user_id FROM employees WHERE email='admin@company.vn');
 UPDATE employees SET email='  AdMiN@Company.vn  ' WHERE email='admin@company.vn'`)
	authdb.Require(t, "đổi nền trước seed lại", err)
	snapshot := func() string {
		t.Helper()
		var result string
		authdb.Require(t, "snapshot seed trong bộ nhớ", db.Conn.QueryRow(t.Context(), `SELECT jsonb_build_object(
   'users',(SELECT jsonb_agg(to_jsonb(u) ORDER BY user_id) FROM users u),
   'employees',(SELECT jsonb_agg(to_jsonb(e) ORDER BY employee_id) FROM employees e),
   'roles',(SELECT jsonb_agg(to_jsonb(r) ORDER BY user_id,role_code) FROM user_roles r),
   'skills',(SELECT jsonb_agg(to_jsonb(s) ORDER BY employee_skill_id) FROM employee_skills s),
   'sequence',(SELECT jsonb_build_array(last_value,is_called) FROM employee_code_seq))::text`).Scan(&result))
		return result
	}
	before := snapshot()
	seedSynthetic(t, db, "synthetic hash replacement")
	if snapshot() != before {
		t.Fatal("seed lại đã đổi dữ liệu hoặc tiêu thụ thêm mã")
	}
	var ok bool
	authdb.Require(t, "kiểm schema seed M2", db.Conn.QueryRow(t.Context(), `SELECT
  (SELECT count(*)=20 AND bool_and(system_role='EMPLOYEE') FROM users)
  AND (SELECT count(*)=20 AND bool_and(employee_code ~ '^EMP[0-9]{6,}$') FROM employees)
  AND (SELECT count(*)>0 FROM employee_skills)
  AND NOT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name='users' AND column_name='username')`).Scan(&ok))
	if !ok {
		t.Fatal("seed chưa đúng schema M2")
	}
}

func TestSeedRejectsLegacyKeyConflictWithoutHalfWrite(t *testing.T) {
	db := seedCase(t)
	tx, err := db.Conn.Begin(t.Context())
	authdb.Require(t, "begin fixture collision", err)
	authdb.Require(t, "reference fixture", execSQLFile(t.Context(), tx, "01_reference.sql"))
	var user int
	authdb.Require(t, "user collision", tx.QueryRow(t.Context(), `INSERT INTO users(password_hash,system_role) VALUES('synthetic','EMPLOYEE') RETURNING user_id`).Scan(&user))
	_, err = tx.Exec(t.Context(), `INSERT INTO employees(user_id,employee_code,fullname,email,position_id)
 SELECT $1,'EMP001','Synthetic conflict','different@example.invalid',position_id FROM positions ORDER BY position_id LIMIT 1`, user)
	authdb.Require(t, "employee collision", err)
	authdb.Require(t, "commit collision fixture", tx.Commit(t.Context()))
	tx, err = db.Conn.Begin(t.Context())
	authdb.Require(t, "begin conflicting seed", err)
	if seedPeople(t.Context(), tx, "synthetic") == nil {
		t.Fatal("seed phải dừng khi mã legacy thuộc email khác")
	}
	authdb.Require(t, "rollback seed", tx.Rollback(t.Context()))
	var count int
	authdb.Require(t, "check không user mồ côi", db.Conn.QueryRow(t.Context(), `SELECT count(*) FROM users`).Scan(&count))
	if count != 1 {
		t.Fatal("seed lỗi đã ghi thêm user")
	}
}

func TestSeedM1LegacyUpgradePreservesExistingAndSequence(t *testing.T) {
	db := authdb.NewCase(t)
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()
	m := db.OpenMigrator(t, migrations.FS)
	authdb.Require(t, "nền M1 legacy cho seed", m.Migrate(migrations.M1Version))
	_, err := db.Conn.Exec(ctx, `INSERT INTO positions(position_name) VALUES('Synthetic legacy position');
 INSERT INTO users(user_id,username,password_hash,system_role,status,must_change_password)
 OVERRIDING SYSTEM VALUE VALUES(101,'synthetic-legacy-admin','synthetic legacy credential','ADMIN','LOCKED',true);
 INSERT INTO employees(employee_id,user_id,employee_code,fullname,email,position_id)
 OVERRIDING SYSTEM VALUE SELECT 201,101,'EMP001','Synthetic legacy admin','  AdMiN@Company.vn  ',position_id FROM positions;
 INSERT INTO user_roles(user_id,role_code) VALUES(101,'ADMIN'),(101,'HR')`)
	authdb.Require(t, "dựng tài khoản legacy", err)
	authdb.Require(t, "nâng legacy M1 lên M2", m.Up())
	m.Close(t)
	original := func() string {
		t.Helper()
		var result string
		authdb.Require(t, "snapshot legacy seed trong bộ nhớ", db.Conn.QueryRow(ctx, `SELECT jsonb_build_object(
   'user',(SELECT to_jsonb(u) FROM users u WHERE user_id=101),
   'employee',(SELECT to_jsonb(e) FROM employees e WHERE employee_id=201),
   'roles',(SELECT jsonb_agg(to_jsonb(r) ORDER BY role_code) FROM user_roles r WHERE user_id=101))::text`).Scan(&result))
		return result
	}
	before := original()
	seedSynthetic(t, db, "synthetic replacement credential")
	seedSynthetic(t, db, "synthetic second replacement credential")
	if original() != before {
		t.Fatal("seed M2 đã đổi dữ liệu tài khoản legacy")
	}
	var ok bool
	authdb.Require(t, "dải cấp mã sau legacy và seed lại", db.Conn.QueryRow(ctx, `SELECT
  (SELECT employee_code='EMP001' AND email='  AdMiN@Company.vn  ' AND user_id=101 FROM employees WHERE employee_id=201)
  AND (SELECT employee_code='EMP000002' FROM employees WHERE email='pm.hoa@company.vn')
  AND (SELECT last_value=20 AND is_called FROM employee_code_seq)
  AND (SELECT count(*)=20 FROM users)
  AND (SELECT count(*)=19 AND bool_and(system_role='EMPLOYEE') FROM users WHERE user_id<>101)
  AND NOT EXISTS(SELECT 1 FROM employees WHERE employee_code='EMP000001')
  AND EXISTS(SELECT 1 FROM employee_skills WHERE approved_by_employee_id=201)`).Scan(&ok))
	if !ok {
		t.Fatal("seed sau nâng legacy cấp lại mã, reset sequence hoặc mất liên kết kỹ năng")
	}
}
