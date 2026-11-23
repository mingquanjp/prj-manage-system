//go:build integration

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"ai-staffing/internal/testutil/authdb"
)

// Đi qua composition root thực, login và bootstrap CSRF trước mutation PCA.
func TestProfileChangeApplicationComposition(t *testing.T) {
	a := newAuthApp(t)
	actor := a.fixture.AddAccount(t, authdb.AccountOptions{Email: "pca-composition@example.invalid"})
	cookie := a.login(t, "pca-composition@example.invalid", authdb.FixturePassword, nil)
	token := a.csrf(t, cookie)
	path := fmt.Sprintf("/api/employees/%d/profile-change-requests", actor.EmployeeID)
	payload := `{"expected_before_values":{"address":null},"proposed_values":{"address":"Địa chỉ tổng hợp"}}`
	req, err := http.NewRequestWithContext(t.Context(), "POST", a.server.URL+path, strings.NewReader(payload))
	authdb.Require(t, "tạo request PCA composition", err)
	req.AddCookie(cookie)
	req.Header.Set("Origin", testOrigin)
	req.Header.Set("X-CSRF-Token", token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "00000000-0000-4000-8000-000000000041")
	res, err := a.server.Client().Do(req)
	authdb.Require(t, "gọi PCA qua composition", err)
	defer res.Body.Close()
	var detail struct {
		ID int `json:"request_id"`
	}
	if res.StatusCode != 201 || json.NewDecoder(res.Body).Decode(&detail) != nil || detail.ID < 1 || res.Header.Get("Cache-Control") != "no-store" || res.Header.Get("Set-Cookie") != "" {
		t.Fatal("Composition PCA không trả kết quả commit đúng contract")
	}
	a.call(t, "GET", fmt.Sprintf("/api/profile-change-requests/%d", detail.ID), cookie, "", nil, 200)
	a.call(t, "GET", "/api/profile-change-requests", cookie, "", nil, 200)
	_, raw := a.call(t, "GET", "/docs/profile-change.openapi.json", nil, "", nil, 200)
	var spec struct {
		Paths map[string]json.RawMessage `json:"paths"`
	}
	if json.Unmarshal(raw, &spec) != nil || len(spec.Paths) != 5 {
		t.Fatal("Composition không mở đúng năm path PCA")
	}
}
