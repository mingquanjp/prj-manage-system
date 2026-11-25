//go:build integration

package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ai-staffing/internal/repository/postgres"
	"ai-staffing/internal/service/auth"
	org "ai-staffing/internal/service/organization"
	"ai-staffing/internal/testutil/authdb"
	"github.com/jackc/pgx/v5"
)

func administrativeIntegrationRouter(t *testing.T, f *authdb.Fixture) http.Handler {
	return administrativeIntegrationRouterDB(t, f, f.DB.Conn, 5*time.Second)
}
func administrativeIntegrationRouterDB(t *testing.T, f *authdb.Fixture, db interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}, timeout time.Duration) http.Handler {
	t.Helper()
	clock := csrfTestClock{time.Now().UTC()}
	sessions := postgres.NewAuthRepository(f.DB.Conn, clock)
	repo, err := postgres.NewAdministrativeProfileRepository(db, timeout)
	authdb.Require(t, "dựng repository HR thật", err)
	h, err := NewAuthRouter(fakeHealth{}, "dev", loginTestOrigin, AuthServices{Sessions: auth.NewService(sessions, clock), Login: auth.NewLoginService(sessions, clock, 8*time.Hour), PasswordChange: auth.NewPasswordChangeService(sessions, clock), Clock: clock, AdministrativeProfiles: org.NewAdministrativeProfileService(repo)})
	authdb.Require(t, "dựng HTTP HR thật", err)
	return h
}

func TestAdministrativeHTTPPostgresTimeoutAndUnknownCommit(t *testing.T) {
	t.Run("read timeout", func(t *testing.T) {
		f := authdb.NewFixture(t)
		hr := pcaHTTPAccount(t, f, authdb.AccountOptions{Email: "f4-timeout@example.invalid", Roles: []string{"HR"}})
		pool := f.DB.Pool(t)
		defer pool.Close()
		locker := f.DB.Connect(t)
		tx, err := locker.Begin(t.Context())
		authdb.Require(t, "mở barrier read F4", err)
		defer tx.Rollback(t.Context())
		_, err = tx.Exec(t.Context(), `LOCK TABLE positions IN ACCESS EXCLUSIVE MODE`)
		authdb.Require(t, "khóa danh mục trước HTTP projection", err)
		h := administrativeIntegrationRouterDB(t, f, pool, 100*time.Millisecond)
		pcaHTTPCall(t, h, hr, "GET", fmt.Sprintf("/api/employees/%d/administrative-profile", hr.EmployeeID), "", 0, 503, "SERVICE_UNAVAILABLE")
	})
	t.Run("lost commit", func(t *testing.T) {
		f := authdb.NewFixture(t)
		hr := pcaHTTPAccount(t, f, authdb.AccountOptions{Email: "f4-commit-hr@example.invalid", Roles: []string{"HR"}})
		target := pcaHTTPAccount(t, f, authdb.AccountOptions{Email: "f4-commit-target@example.invalid"})
		h := administrativeIntegrationRouterDB(t, f, accountHTTPLostReplyDB{f.DB.Conn}, 5*time.Second)
		body := `{"expected_before_values":{"company_email":"f4-commit-target@example.invalid"},"changes":{"company_email":"f4-commit-new@example.invalid"}}`
		w := pcaHTTPCall(t, h, hr, "PATCH", fmt.Sprintf("/api/employees/%d", target.EmployeeID), body, 0, 503, "SERVICE_UNAVAILABLE")
		if strings.Contains(w.Body.String(), "employee_id") || strings.Contains(w.Body.String(), "f4-commit-new") {
			t.Fatal("unknown commit trả success/PII giả")
		}
		if pcaHTTPCount(t, f, `SELECT count(*) FROM employees WHERE employee_id=$1 AND email='f4-commit-new@example.invalid'`, target.EmployeeID) != 1 || pcaHTTPCount(t, f, `SELECT count(*) FROM audit_events`) != 1 || pcaHTTPCount(t, f, `SELECT count(*) FROM auth_sessions WHERE user_id=$1 AND revoked_at IS NULL`, target.UserID) != 0 {
			t.Fatal("mất xác nhận đã retry hoặc commit không nguyên tử")
		}
		normal := administrativeIntegrationRouter(t, f)
		w = pcaHTTPCall(t, normal, hr, "GET", fmt.Sprintf("/api/employees/%d/administrative-profile", target.EmployeeID), "", 0, 200, "")
		var result map[string]any
		authdb.Require(t, "GET đối soát commit thật", json.Unmarshal(w.Body.Bytes(), &result))
		if result["company_email"] != "f4-commit-new@example.invalid" {
			t.Fatal("GET không đối soát được email")
		}
	})
}

func TestAdministrativeHTTPPostgresThreeRoutesAndEmailRevocation(t *testing.T) {
	f := authdb.NewFixture(t)
	hr := pcaHTTPAccount(t, f, authdb.AccountOptions{Email: "f4-hr@example.invalid", Roles: []string{"HR", "ADMIN"}})
	target := pcaHTTPAccount(t, f, authdb.AccountOptions{Email: " F4-Target@Example.Invalid "})
	admin := pcaHTTPAccount(t, f, authdb.AccountOptions{Email: "f4-admin@example.invalid", Roles: []string{"ADMIN"}})
	limited := pcaHTTPAccount(t, f, authdb.AccountOptions{Email: "f4-limited@example.invalid", Roles: []string{"HR", "ADMIN"}, MustChangePassword: true})
	h := administrativeIntegrationRouter(t, f)
	detailPath := fmt.Sprintf("/api/employees/%d/administrative-profile", target.EmployeeID)
	patchPath := fmt.Sprintf("/api/employees/%d", target.EmployeeID)
	w := pcaHTTPCall(t, h, hr, "GET", detailPath, "", 0, 200, "")
	var detail map[string]any
	authdb.Require(t, "đọc projection F4", json.Unmarshal(w.Body.Bytes(), &detail))
	if len(detail) != 11 || detail["user_id"] != float64(target.UserID) || detail["employee_id"] != float64(target.EmployeeID) || detail["company_email"] != "f4-target@example.invalid" || detail["address"] != nil {
		t.Fatal("projection canonical/ID/NULL sai")
	}
	w = pcaHTTPCall(t, h, hr, "GET", "/api/employees?q=f4-target%40example.invalid&page_size=1", "", 0, 200, "")
	var list struct {
		Items []map[string]any
		Total int
	}
	authdb.Require(t, "đọc list HR", json.Unmarshal(w.Body.Bytes(), &list))
	if list.Total != 1 || len(list.Items) != 1 || len(list.Items[0]) != 8 {
		t.Fatal("list sai projection hoặc canonical search")
	}
	pcaHTTPCall(t, h, admin, "GET", detailPath, "", 0, 403, "FORBIDDEN")
	pcaHTTPCall(t, h, limited, "GET", detailPath, "", 0, 403, "PASSWORD_CHANGE_REQUIRED")
	noop := `{"expected_before_values":{"company_email":"f4-target@example.invalid"},"changes":{"company_email":" F4-TARGET@example.invalid "}}`
	pcaHTTPCall(t, h, hr, "PATCH", patchPath, noop, 0, 200, "")
	if pcaHTTPCount(t, f, `SELECT count(*) FROM audit_events`) != 0 || pcaHTTPCount(t, f, `SELECT count(*) FROM auth_sessions WHERE user_id=$1 AND revoked_at IS NULL`, target.UserID) != 1 {
		t.Fatal("canonical no-op audit/revoke giả")
	}
	if pcaHTTPCount(t, f, `SELECT count(*) FROM employees WHERE employee_id=$1 AND email=$2`, target.EmployeeID, " F4-Target@Example.Invalid ") != 1 {
		t.Fatal("no-op đã sửa email legacy")
	}
	pcaHTTPCall(t, h, hr, "PATCH", patchPath, administrativeHTTPBody, 0, 200, "")
	pcaHTTPCall(t, h, hr, "PATCH", patchPath, administrativeHTTPBody, 0, 409, "STALE_DATA")
	if pcaHTTPCount(t, f, `SELECT count(*) FROM employees WHERE employee_id=$1 AND address='Địa chỉ mới'`, target.EmployeeID) != 1 {
		t.Fatal("address chưa persist")
	}
	changeEmail := `{"expected_before_values":{"company_email":"f4-target@example.invalid"},"changes":{"company_email":"f4-new@example.invalid"}}`
	pcaHTTPCall(t, h, hr, "PATCH", patchPath, changeEmail, 0, 200, "")
	if pcaHTTPCount(t, f, `SELECT count(*) FROM auth_sessions WHERE user_id=$1 AND revoked_at IS NULL`, target.UserID) != 0 || pcaHTTPCount(t, f, `SELECT count(*) FROM audit_events`) != 2 {
		t.Fatal("email/audit/revoke không cùng tác động")
	}
	pcaHTTPCall(t, h, target, "GET", "/api/auth/me", "", 0, 401, "UNAUTHENTICATED")
	pcaHTTPExec(t, f, `UPDATE employees SET employment_status='INACTIVE' WHERE employee_id=$1`, target.EmployeeID)
	pcaHTTPCall(t, h, hr, "PATCH", patchPath, `{"expected_before_values":{"fullname":"`+target.Fullname+`"},"changes":{"fullname":"Tên mới"}}`, 0, 200, "")
	if pcaHTTPCount(t, f, `SELECT count(*) FROM employees WHERE employee_id=$1 AND employment_status='INACTIVE'`, target.EmployeeID) != 1 {
		t.Fatal("sửa hành chính tự activate")
	}
	pcaHTTPCall(t, h, hr, "PATCH", patchPath, `{"expected_before_values":{"position_id":1},"changes":{"position_id":999999}}`, 0, 409, "CATALOG_NOT_FOUND")
	pcaHTTPCall(t, h, hr, "PATCH", patchPath, `{"expected_before_values":{"company_email":"f4-new@example.invalid"},"changes":{"company_email":"f4-admin@example.invalid"}}`, 0, 409, "EMAIL_CONFLICT")
	// Audit lỗi phải rollback email và giữ phiên của target còn nguyên.
	pcaHTTPExec(t, f, `CREATE FUNCTION f4_http_audit_fault() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'synthetic audit failure'; END $$`)
	pcaHTTPExec(t, f, `CREATE TRIGGER f4_http_audit_fault BEFORE INSERT ON audit_events FOR EACH ROW EXECUTE FUNCTION f4_http_audit_fault()`)
	hrPath := fmt.Sprintf("/api/employees/%d", hr.EmployeeID)
	pcaHTTPCall(t, h, hr, "PATCH", hrPath, `{"expected_before_values":{"company_email":"f4-hr@example.invalid"},"changes":{"company_email":"f4-hr-new@example.invalid"}}`, 0, 500, "INTERNAL_ERROR")
	if pcaHTTPCount(t, f, `SELECT count(*) FROM employees WHERE employee_id=$1 AND email='f4-hr@example.invalid'`, hr.EmployeeID) != 1 || pcaHTTPCount(t, f, `SELECT count(*) FROM auth_sessions WHERE user_id=$1 AND revoked_at IS NULL`, hr.UserID) != 1 {
		t.Fatal("audit lỗi không rollback email/revoke")
	}
	pcaHTTPExec(t, f, `DROP TRIGGER f4_http_audit_fault ON audit_events`)
	pcaHTTPExec(t, f, `DROP FUNCTION f4_http_audit_fault()`)
	// Header/body được kiểm qua router dùng session và adapter PG thật.
	for _, n := range []int{65536, 65537} {
		body := administrativeHTTPBody + strings.Repeat(" ", n-len(administrativeHTTPBody))
		r := pcaIntegrationRequest(hr, "PATCH", patchPath, body, 0)
		r.ContentLength = -1
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if n == 65537 {
			assertCatalogHTTP(t, w, 413, "PAYLOAD_TOO_LARGE")
		} else {
			assertCatalogHTTP(t, w, 409, "STALE_DATA")
		}
	}
	pcaHTTPExec(t, f, `DELETE FROM user_roles WHERE user_id=$1 AND role_code='HR'`, hr.UserID)
	pcaHTTPCall(t, h, hr, "GET", detailPath, "", 0, 403, "FORBIDDEN")
}
