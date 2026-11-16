//go:build integration

package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"ai-staffing/internal/testutil/authdb"
)

func TestSelfProfileApplicationPostgres(t *testing.T) {
	a := newAuthApp(t)
	f := a.fixture
	actor := f.AddAccount(t, authdb.AccountOptions{Email: "  Profile@Example.invalid  ", Roles: []string{"ADMIN", "HR", "DIRECTOR"}})
	other := f.AddAccount(t, authdb.AccountOptions{Email: "other-profile@example.invalid"})
	limited := f.AddAccount(t, authdb.AccountOptions{Email: "limited-profile@example.invalid", Roles: []string{"ADMIN"}, MustChangePassword: true})
	cookie := a.login(t, "profile@example.invalid", authdb.FixturePassword, nil)
	path := "/api/employees/" + strconv.Itoa(actor.EmployeeID)
	_, data := catalogCall(t, a, "GET", path, cookie, "", nil, 200)
	var fields map[string]any
	authdb.Require(t, "DTO self", json.Unmarshal(data, &fields))
	if len(fields) != 11 || fields["user_id"] != float64(actor.UserID) || fields["employee_id"] != float64(actor.EmployeeID) || fields["company_email"] != "profile@example.invalid" || fields["address"] != nil || fields["phone"] != nil || fields["department"] != nil || fields["job_level"] != nil {
		t.Fatal("DTO self/PII/null sai")
	}
	for _, key := range []string{"password_hash", "system_roles", "must_change_password", "username", "audit_events", "cv", "history"} {
		if _, ok := fields[key]; ok {
			t.Fatal("DTO ngoài phép")
		}
	}
	var raw string
	authdb.Require(t, "email legacy nguyên vẹn", f.DB.Conn.QueryRow(t.Context(), `SELECT email FROM employees WHERE user_id=$1`, actor.UserID).Scan(&raw))
	if raw != "  Profile@Example.invalid  " {
		t.Fatal("đọc ghi lại email")
	}
	var otherError string
	for _, id := range []int{other.EmployeeID, 2147483647} {
		_, body := catalogCall(t, a, "GET", "/api/employees/"+strconv.Itoa(id), cookie, "", nil, 404)
		assertErrorCode(t, body, "NOT_FOUND")
		var e map[string]any
		authdb.Require(t, "decode lỗi", json.Unmarshal(body, &e))
		message := e["message"].(string)
		if otherError != "" && message != otherError {
			t.Fatal("404 phân biệt ID người khác/missing")
		}
		otherError = message
	}
	_, data = catalogCall(t, a, "GET", path+"?actor_id=1", cookie, "", nil, 400)
	assertErrorCode(t, data, "VALIDATION_ERROR")
	catalogCall(t, a, "GET", path, nil, "", nil, 401)
	lc := a.login(t, "limited-profile@example.invalid", authdb.FixturePassword, nil)
	_, data = catalogCall(t, a, "GET", "/api/employees/"+strconv.Itoa(limited.EmployeeID), lc, "", nil, 403)
	assertErrorCode(t, data, "PASSWORD_CHANGE_REQUIRED")
	// Không role hệ thống vẫn đọc chính chủ; không nhận role từ client để mở quyền.
	_, err := f.DB.Conn.Exec(t.Context(), `DELETE FROM user_roles WHERE user_id=$1`, actor.UserID)
	authdb.Require(t, "xóa role fixture", err)
	catalogCall(t, a, "GET", path, cookie, "", nil, 200)
	_, err = f.DB.Conn.Exec(t.Context(), `UPDATE users SET must_change_password=true WHERE user_id=$1`, actor.UserID)
	authdb.Require(t, "cờ hiện hành", err)
	_, data = catalogCall(t, a, "GET", path, cookie, "", nil, 403)
	assertErrorCode(t, data, "PASSWORD_CHANGE_REQUIRED")
	_, err = f.DB.Conn.Exec(t.Context(), `UPDATE users SET must_change_password=false,status='LOCKED' WHERE user_id=$1`, actor.UserID)
	authdb.Require(t, "khóa actor", err)
	_, data = catalogCall(t, a, "GET", path, cookie, "", nil, 401)
	assertErrorCode(t, data, "UNAUTHENTICATED")
	_, err = f.DB.Conn.Exec(t.Context(), `UPDATE users SET status='ACTIVE' WHERE user_id=$1`, actor.UserID)
	authdb.Require(t, "mở khóa fixture", err)
	_, err = f.DB.Conn.Exec(t.Context(), `ALTER TABLE departments RENAME TO departments_unavailable`)
	authdb.Require(t, "fault projection DB test", err)
	_, data = catalogCall(t, a, "GET", path, cookie, "", nil, 500)
	assertErrorCode(t, data, "INTERNAL_ERROR")
	if strings.Contains(string(data), "departments") {
		t.Fatal("lộ lỗi DB")
	}
	_, err = f.DB.Conn.Exec(t.Context(), `ALTER TABLE departments_unavailable RENAME TO departments`)
	authdb.Require(t, "khôi phục fixture", err)
	a.call(t, "POST", "/api/auth/logout", cookie, a.csrf(t, cookie), nil, http.StatusNoContent)
	catalogCall(t, a, "GET", path, cookie, "", nil, 401)
}
