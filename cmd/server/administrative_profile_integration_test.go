//go:build integration

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"ai-staffing/internal/testutil/authdb"
)

// Kiểm xuyên composition root, cookie/CSRF và DB thật; không thay browser của T12.
func TestAdministrativeProfileApplicationComposition(t *testing.T) {
	a := newAuthApp(t)
	f := a.fixture
	hr := f.AddAccount(t, authdb.AccountOptions{Email: "hr-f4@example.invalid", Roles: []string{"HR"}})
	target := f.AddAccount(t, authdb.AccountOptions{Email: "target-f4@example.invalid"})
	admin := f.AddAccount(t, authdb.AccountOptions{Email: "admin-f4@example.invalid", Roles: []string{"ADMIN"}})
	cookie := a.login(t, "hr-f4@example.invalid", authdb.FixturePassword, nil)
	oldTarget := a.login(t, "target-f4@example.invalid", authdb.FixturePassword, nil)
	oldTarget2 := a.login(t, "target-f4@example.invalid", authdb.FixturePassword, nil)
	token := a.csrf(t, cookie)
	path := fmt.Sprintf("/api/employees/%d", target.EmployeeID)
	_, raw := catalogCall(t, a, "GET", "/api/employees?q=target-f4", cookie, "", nil, 200)
	var page struct {
		Items []map[string]json.RawMessage `json:"items"`
		Total int                          `json:"total"`
	}
	authdb.Require(t, "đọc danh sách F4", json.Unmarshal(raw, &page))
	if page.Total != 1 || len(page.Items) != 1 || len(page.Items[0]) != 8 {
		t.Fatal("projection danh sách HR không đúng allowlist")
	}
	for _, key := range []string{"phone", "address", "user_id", "system_roles", "password_hash"} {
		if _, exists := page.Items[0][key]; exists {
			t.Fatal("danh sách HR lộ trường ngoài mục đích")
		}
	}
	catalogCall(t, a, "GET", path+"/administrative-profile", cookie, "", nil, 200)
	adminCookie := a.login(t, "admin-f4@example.invalid", authdb.FixturePassword, nil)
	for _, p := range []string{"/api/employees", path + "/administrative-profile"} {
		_, data := catalogCall(t, a, "GET", p, adminCookie, "", nil, 403)
		assertErrorCode(t, data, "FORBIDDEN")
	}
	// Self endpoint không được mở rộng ngầm thành quyền đọc hành chính của HR.
	catalogCall(t, a, "GET", path, cookie, "", nil, 404)
	body := map[string]any{
		"expected_before_values": map[string]any{"company_email": "target-f4@example.invalid", "address": nil},
		"changes":                map[string]any{"company_email": "new-f4@example.invalid", "address": "Địa chỉ tổng hợp F4"},
	}
	catalogCall(t, a, "PATCH", path, cookie, token, body, 200)
	for _, c := range []*http.Cookie{oldTarget, oldTarget2} {
		a.call(t, "GET", "/api/auth/me", c, "", nil, 401)
	}
	_, data := a.call(t, "POST", "/api/auth/login", nil, "", map[string]string{"email": "target-f4@example.invalid", "password": authdb.FixturePassword}, 401)
	assertErrorCode(t, data, "INVALID_CREDENTIALS")
	newTarget := a.login(t, "new-f4@example.invalid", authdb.FixturePassword, nil)
	catalogCall(t, a, "GET", path, newTarget, "", nil, 200)
	var audit, pending int
	authdb.Require(t, "kiểm audit direct edit", f.DB.Conn.QueryRow(t.Context(), `SELECT count(*) FROM audit_events WHERE actor_id=$1 AND entity_id=$2`, hr.UserID, target.EmployeeID).Scan(&audit))
	authdb.Require(t, "không giả đề nghị nhân viên", f.DB.Conn.QueryRow(t.Context(), `SELECT count(*) FROM employee_profile_change_requests`).Scan(&pending))
	if audit != 1 || pending != 0 || admin.UserID == hr.UserID {
		t.Fatal("direct edit thiếu audit hoặc giả thành đề nghị")
	}
}

func TestAdministrativeProfileSelfEmailRequiresNewLogin(t *testing.T) {
	a := newAuthApp(t)
	hr := a.fixture.AddAccount(t, authdb.AccountOptions{Email: "self-hr-f4@example.invalid", Roles: []string{"HR"}})
	cookie := a.login(t, "self-hr-f4@example.invalid", authdb.FixturePassword, nil)
	token := a.csrf(t, cookie)
	path := fmt.Sprintf("/api/employees/%d", hr.EmployeeID)
	catalogCall(t, a, "PATCH", path, cookie, token, map[string]any{
		"expected_before_values": map[string]any{"company_email": "self-hr-f4@example.invalid"},
		"changes":                map[string]any{"company_email": "changed-self-hr-f4@example.invalid"},
	}, 200)
	a.call(t, "GET", "/api/auth/me", cookie, "", nil, 401)
	catalogCall(t, a, "GET", path+"/administrative-profile", cookie, "", nil, 401)
	newCookie := a.login(t, "changed-self-hr-f4@example.invalid", authdb.FixturePassword, nil)
	catalogCall(t, a, "GET", path+"/administrative-profile", newCookie, "", nil, 200)
}
