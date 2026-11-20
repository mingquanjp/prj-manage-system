//go:build integration

package main

import (
	"ai-staffing/internal/testutil/authdb"
	"encoding/json"
	"testing"
)

func TestAccountApplicationRealSessionsAndComposition(t *testing.T) {
	a := newAuthApp(t)
	f := a.fixture
	admin := f.AddAccount(t, authdb.AccountOptions{Email: "admin-account@example.invalid", Roles: []string{"ADMIN", "HR"}})
	f.AddAccount(t, authdb.AccountOptions{Email: "plain-account@example.invalid"})
	f.AddAccount(t, authdb.AccountOptions{Email: "limited-account@example.invalid", Roles: []string{"ADMIN"}, MustChangePassword: true})
	cookie := a.login(t, "admin-account@example.invalid", authdb.FixturePassword, nil)
	token := a.csrf(t, cookie)
	_, raw := catalogCall(t, a, "GET", "/api/employees/next-code", cookie, "", nil, 200)
	var preview struct {
		Code      string `json:"employee_code"`
		IsPreview bool   `json:"is_preview"`
	}
	authdb.Require(t, "decode preview composition", json.Unmarshal(raw, &preview))
	if preview.Code != "EMP000001" || !preview.IsPreview {
		t.Fatal("Preview sai")
	}
	var position int
	authdb.Require(t, "đọc vị trí tổng hợp", f.DB.Conn.QueryRow(t.Context(), `SELECT position_id FROM employees WHERE user_id=$1`, admin.UserID).Scan(&position))
	input := map[string]any{"fullname": "Nhân viên test API", "company_email": "created-api@example.invalid", "position_id": position, "system_roles": []string{}}
	_, raw = catalogCall(t, a, "POST", "/api/employees", cookie, token, input, 201)
	var created struct {
		Created bool `json:"account_created"`
		Account struct {
			Code string `json:"employee_code"`
		} `json:"account"`
		Mail struct {
			Status string `json:"status"`
		} `json:"mail"`
	}
	authdb.Require(t, "decode create composition", json.Unmarshal(raw, &created))
	if !created.Created || created.Account.Code != preview.Code || created.Mail.Status != "failed" {
		t.Fatal("Response create sai")
	}
	_, raw = catalogCall(t, a, "GET", "/api/users?q=created-api", cookie, "", nil, 200)
	var list struct {
		Items []map[string]any `json:"items"`
	}
	authdb.Require(t, "decode list composition", json.Unmarshal(raw, &list))
	if len(list.Items) != 1 {
		t.Fatal("Không đối soát được account")
	}
	response, err := a.server.Client().Get(a.server.URL + "/docs/catalog.openapi.json")
	authdb.Require(t, "đọc Swagger từ composition", err)
	defer response.Body.Close()
	var spec struct {
		Paths map[string]map[string]any `json:"paths"`
	}
	authdb.Require(t, "decode Swagger canonical", json.NewDecoder(response.Body).Decode(&spec))
	operations := 0
	for _, methods := range spec.Paths {
		operations += len(methods)
	}
	if response.StatusCode != 200 || operations != 16 || spec.Paths["/api/employees/next-code"]["get"] == nil || spec.Paths["/api/employees/{id}"]["get"] == nil {
		t.Fatal("Swagger thiếu operation hoặc mất profile khi nối account")
	}

	_, raw = catalogCall(t, a, "POST", "/api/employees", cookie, token, input, 409)
	assertErrorCode(t, raw, "EMAIL_CONFLICT")
	plain := a.login(t, "plain-account@example.invalid", authdb.FixturePassword, nil)
	limited := a.login(t, "limited-account@example.invalid", authdb.FixturePassword, nil)
	for _, path := range []string{"/api/users", "/api/employees/next-code"} {
		_, raw = catalogCall(t, a, "GET", path, plain, "", nil, 403)
		assertErrorCode(t, raw, "FORBIDDEN")
		_, raw = catalogCall(t, a, "GET", path, limited, "", nil, 403)
		assertErrorCode(t, raw, "PASSWORD_CHANGE_REQUIRED")
	}
}
