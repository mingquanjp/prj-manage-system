//go:build integration

package postgres_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"ai-staffing/internal/repository/postgres"
	org "ai-staffing/internal/service/organization"
	"ai-staffing/internal/testutil/authdb"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type accountBeginner interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}

func accountRepo(t *testing.T, db accountBeginner) *postgres.AccountRepository {
	t.Helper()
	r, err := postgres.NewAccountRepository(db, 10*time.Second)
	authdb.Require(t, "khởi tạo create adapter", err)
	return r
}
func accountCommand(t *testing.T, f *authdb.Fixture, a authdb.Account, email string) org.CreateAccountCommand {
	t.Helper()
	var position int
	authdb.Require(t, "vị trí fixture create", f.DB.Conn.QueryRow(t.Context(), `SELECT position_id FROM employees WHERE employee_id=$1`, a.EmployeeID).Scan(&position))
	return org.CreateAccountCommand{Input: org.CreateAccountInput{ActorID: a.UserID, Fullname: "Nhân viên tổng hợp", CompanyEmail: email, PositionID: position, SystemRoles: []string{}}, PasswordHash: f.PasswordHash}
}
func accountSnapshot(t *testing.T, f *authdb.Fixture) string {
	t.Helper()
	var s string
	authdb.Require(t, "snapshot create trong bộ nhớ", f.DB.Conn.QueryRow(t.Context(), `SELECT jsonb_build_object(
 'users',(SELECT jsonb_agg(to_jsonb(u) ORDER BY user_id) FROM users u),
 'employees',(SELECT jsonb_agg(to_jsonb(e) ORDER BY employee_id) FROM employees e),
 'roles',(SELECT jsonb_agg(to_jsonb(r) ORDER BY user_id,role_code) FROM user_roles r),
 'audit',(SELECT jsonb_agg(to_jsonb(a) ORDER BY audit_id) FROM audit_events a),
 'sessions',(SELECT jsonb_agg(to_jsonb(s) ORDER BY session_id) FROM auth_sessions s))::text`).Scan(&s))
	return s
}
func accountError(t *testing.T, got, want error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Fatalf("create sai loại lỗi muốn %v; không in lỗi thô", want)
	}
}
func accountNoResult(t *testing.T, got org.AccountProjection) {
	t.Helper()
	if !reflect.DeepEqual(got, org.AccountProjection{}) {
		t.Fatal("lỗi create không được trả projection thành công")
	}
}

func TestAccountCreateAtomicProjectionAndAudit(t *testing.T) {
	for i, roles := range [][]string{{}, {"ADMIN"}, {"HR", "DIRECTOR", "ADMIN"}} {
		t.Run(fmt.Sprintf("roles-%d", i), func(t *testing.T) {
			f := authdb.NewFixture(t)
			admin := f.AddAccount(t, authdb.AccountOptions{Email: "actor@example.invalid", Roles: []string{"ADMIN"}})
			cmd := accountCommand(t, f, admin, "  CREATED@EXAMPLE.INVALID  ")
			cmd.Input.SystemRoles = roles
			if i == 2 {
				var dept, level int
				authdb.Require(t, "department", f.DB.Conn.QueryRow(t.Context(), `INSERT INTO departments(department_name) VALUES('Fixture department') RETURNING department_id`).Scan(&dept))
				authdb.Require(t, "job level", f.DB.Conn.QueryRow(t.Context(), `INSERT INTO job_levels(job_level_name) VALUES('Fixture level') RETURNING job_level_id`).Scan(&level))
				phone, address := "123456", "Synthetic address"
				cmd.Input.DepartmentID = &dept
				cmd.Input.JobLevelID = &level
				cmd.Input.Phone = &phone
				cmd.Input.Address = &address
			}
			got, err := accountRepo(t, f.DB.Conn).CreateAccount(t.Context(), cmd)
			authdb.Require(t, "create commit thật", err)
			if got.UserID < 1 || got.EmployeeID < 1 || got.UserID == admin.UserID || got.EmployeeID == admin.EmployeeID || got.EmployeeCode != "EMP000001" || got.CompanyEmail != "created@example.invalid" || got.Status != "ACTIVE" || !got.MustChangePassword {
				t.Fatal("projection create sai")
			}
			var user int
			var hash, legacy, status, employment, email, code string
			var limited bool
			var dept, level *int
			var phone, address *string
			authdb.Require(t, "dữ liệu đã commit", f.DB.Conn.QueryRow(t.Context(), `SELECT e.user_id,u.password_hash,u.system_role,u.status,u.must_change_password,e.employment_status,e.email,e.employee_code,e.department_id,e.job_level_id,e.phone,e.address FROM employees e JOIN users u USING(user_id) WHERE e.employee_id=$1`, got.EmployeeID).Scan(&user, &hash, &legacy, &status, &limited, &employment, &email, &code, &dept, &level, &phone, &address))
			if user != got.UserID || legacy != "EMPLOYEE" || status != "ACTIVE" || employment != "ACTIVE" || !limited || email != got.CompanyEmail || code != got.EmployeeCode || bcrypt.CompareHashAndPassword([]byte(hash), []byte(authdb.FixturePassword)) != nil || !reflect.DeepEqual(dept, cmd.Input.DepartmentID) || !reflect.DeepEqual(level, cmd.Input.JobLevelID) || !reflect.DeepEqual(phone, cmd.Input.Phone) || !reflect.DeepEqual(address, cmd.Input.Address) {
				t.Fatal("user/employee không nguyên tử hoặc liên kết sai")
			}
			var storedRoles []string
			authdb.Require(t, "roles đã commit", f.DB.Conn.QueryRow(t.Context(), `SELECT ARRAY(SELECT role_code FROM user_roles WHERE user_id=$1 ORDER BY CASE role_code WHEN 'ADMIN' THEN 1 WHEN 'HR' THEN 2 ELSE 3 END)`, got.UserID).Scan(&storedRoles))
			if len(storedRoles) != len(roles) || !reflect.DeepEqual(storedRoles, got.SystemRoles) {
				t.Fatal("role projection không đúng nguồn hoặc thứ tự")
			}
			var actor, entity int
			var kind, op string
			var before *string
			var raw []byte
			authdb.Require(t, "audit đã commit", f.DB.Conn.QueryRow(t.Context(), `SELECT actor_id,entity_type,entity_id,operation,before_values::text,after_values FROM audit_events`).Scan(&actor, &kind, &entity, &op, &before, &raw))
			var after map[string]any
			authdb.Require(t, "parse audit allowlist", json.Unmarshal(raw, &after))
			expected := []string{"schema_version", "user_id", "employee_id", "employee_code", "position_id", "department_id", "job_level_id", "system_roles"}
			if actor != admin.UserID || entity != got.EmployeeID || kind != "employee" || op != "CREATE_ACCOUNT" || before != nil || len(after) != len(expected) {
				t.Fatal("audit sai actor/entity hoặc field ngoài allowlist")
			}
			for _, key := range expected {
				if _, ok := after[key]; !ok {
					t.Fatal("audit thiếu field")
				}
			}
			wantAudit := map[string]any{
				"schema_version": float64(1), "user_id": float64(got.UserID), "employee_id": float64(got.EmployeeID),
				"employee_code": got.EmployeeCode, "position_id": float64(cmd.Input.PositionID), "department_id": nil, "job_level_id": nil,
				"system_roles": []any{},
			}
			if cmd.Input.DepartmentID != nil {
				wantAudit["department_id"] = float64(*cmd.Input.DepartmentID)
			}
			if cmd.Input.JobLevelID != nil {
				wantAudit["job_level_id"] = float64(*cmd.Input.JobLevelID)
			}
			ordered := []any{}
			for _, r := range got.SystemRoles {
				ordered = append(ordered, r)
			}
			wantAudit["system_roles"] = ordered
			if !reflect.DeepEqual(after, wantAudit) {
				t.Fatal("audit phải chứa đúng ID/code/catalog/role đã commit, không chỉ đúng keys")
			}

			if catalogCount(t, f, `SELECT count(*) FROM auth_sessions`) != 0 || catalogCount(t, f, `SELECT count(*) FROM audit_events`) != 1 {
				t.Fatal("không tạo session hoặc nhân audit")
			}
		})
	}
}

func TestAccountCreateRechecksActorAndCatalog(t *testing.T) {
	for _, tc := range []struct {
		name string
		opts authdb.AccountOptions
		want error
	}{
		{"no admin", authdb.AccountOptions{Roles: []string{"HR"}, LegacyRole: "ADMIN"}, org.ErrForbidden},
		{"locked", authdb.AccountOptions{Roles: []string{"ADMIN"}, UserStatus: "LOCKED"}, org.ErrUnauthenticated},
		{"inactive", authdb.AccountOptions{Roles: []string{"ADMIN"}, EmploymentStatus: "INACTIVE"}, org.ErrUnauthenticated},
		{"limited admin", authdb.AccountOptions{Roles: []string{"ADMIN"}, MustChangePassword: true}, org.ErrPasswordChangeRequired},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := authdb.NewFixture(t)
			tc.opts.Email = "actor@example.invalid"
			a := f.AddAccount(t, tc.opts)
			cmd := accountCommand(t, f, a, "new@example.invalid")
			before := accountSnapshot(t, f)
			got, err := accountRepo(t, f.DB.Conn).CreateAccount(t.Context(), cmd)
			accountError(t, err, tc.want)
			accountNoResult(t, got)
			if accountSnapshot(t, f) != before {
				t.Fatal("actor lỗi vẫn ghi")
			}
		})
	}
	f := authdb.NewFixture(t)
	a := f.AddAccount(t, authdb.AccountOptions{Email: "admin@example.invalid", Roles: []string{"ADMIN"}})
	for _, kind := range []string{"position", "department", "job-level"} {
		t.Run(kind, func(t *testing.T) {
			cmd := accountCommand(t, f, a, "missing@example.invalid")
			missing := 2147483647
			switch kind {
			case "position":
				cmd.Input.PositionID = missing
			case "department":
				cmd.Input.DepartmentID = &missing
			default:
				cmd.Input.JobLevelID = &missing
			}
			before := accountSnapshot(t, f)
			got, err := accountRepo(t, f.DB.Conn).CreateAccount(t.Context(), cmd)
			accountError(t, err, org.ErrValidation)
			accountNoResult(t, got)
			if accountSnapshot(t, f) != before {
				t.Fatal("catalog thiếu vẫn ghi")
			}
		})
	}
}

func TestAccountCreateWriteFailuresRollbackAllRows(t *testing.T) {
	for _, table := range []string{"users", "employees", "user_roles", "audit_events"} {
		t.Run(table, func(t *testing.T) {
			f := authdb.NewFixture(t)
			a := f.AddAccount(t, authdb.AccountOptions{Email: "admin@example.invalid", Roles: []string{"ADMIN"}})
			cmd := accountCommand(t, f, a, "fault@example.invalid")
			cmd.Input.SystemRoles = []string{"HR"}
			_, err := f.DB.Conn.Exec(t.Context(), fmt.Sprintf(`CREATE FUNCTION public.account_test_failure() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'synthetic failure'; END $$;CREATE TRIGGER account_test_failure BEFORE INSERT ON public.%s FOR EACH ROW EXECUTE FUNCTION public.account_test_failure()`, table))
			authdb.Require(t, "cài fault case hy sinh", err)
			before := accountSnapshot(t, f)
			got, err := accountRepo(t, f.DB.Conn).CreateAccount(t.Context(), cmd)
			accountError(t, err, org.ErrPersistence)
			accountNoResult(t, got)
			if accountSnapshot(t, f) != before {
				t.Fatal("write lỗi đã half-write")
			}
			var last int64
			var called bool
			authdb.Require(t, "sequence vẫn mất số", f.DB.Conn.QueryRow(t.Context(), `SELECT last_value,is_called FROM employee_code_seq`).Scan(&last, &called))
			if !called || last != 1 {
				t.Fatal("rollback không được lấp số")
			}
		})
	}
}

func TestAccountServiceUsesRealBcryptAndCommittedRepository(t *testing.T) {
	f := authdb.NewFixture(t)
	a := f.AddAccount(t, authdb.AccountOptions{Email: "admin@example.invalid", Roles: []string{"ADMIN"}})
	// Chỉ điều khiển identity trong case hy sinh để buộc hai ID khác nhau.
	_, err := f.DB.Conn.Exec(t.Context(), `ALTER TABLE public.employees ALTER COLUMN employee_id RESTART WITH 500`)
	authdb.Require(t, "tách dải identity user/employee trong fixture", err)
	input := accountCommand(t, f, a, "service@example.invalid").Input
	result, err := org.NewAccountService(accountRepo(t, f.DB.Conn)).Create(t.Context(), input)
	if err == nil && (result.Account.UserID == result.Account.EmployeeID || result.Account.EmployeeID != 500) {
		t.Fatal("fixture phải chứng minh user_id khác employee_id")
	}
	authdb.Require(t, "service tới repository commit thật", err)
	var hash string
	var linkedUser int
	authdb.Require(t, "liên kết user/employee thực", f.DB.Conn.QueryRow(t.Context(), `SELECT user_id FROM employees WHERE employee_id=$1`, result.Account.EmployeeID).Scan(&linkedUser))
	if linkedUser != result.Account.UserID {
		t.Fatal("employee phải liên kết user chính xác dù ID khác")
	}
	authdb.Require(t, "hash thực đã lưu", f.DB.Conn.QueryRow(t.Context(), `SELECT password_hash FROM users WHERE user_id=$1`, result.Account.UserID).Scan(&hash))
	password := result.PasswordForDelivery().RevealForDelivery()
	if len(password) != 43 || bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		t.Fatal("secret/bcrypt/commit không liên kết đúng")
	}
	if _, err := json.Marshal(result); err == nil {
		t.Fatal("kết quả nội bộ không được thành JSON transport")
	}
	if catalogCount(t, f, `SELECT count(*) FROM audit_events WHERE operation='CREATE_ACCOUNT'`) != 1 || catalogCount(t, f, `SELECT count(*) FROM auth_sessions`) != 0 {
		t.Fatal("service không được tạo session hoặc nhân audit")
	}
}
