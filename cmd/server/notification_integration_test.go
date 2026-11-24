//go:build integration

package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"ai-staffing/internal/repository/postgres"
	pc "ai-staffing/internal/service/profilechange"
	"ai-staffing/internal/testutil/authdb"
)

func TestNotificationApplicationPostgres(t *testing.T) {
	a := newAuthApp(t)
	f := a.fixture
	// Cả hai nhóm phải còn truy cập được sau khi gộp wiring DATN-41/42.
	for path, count := range map[string]int{"/docs/profile-change.openapi.json": 5, "/docs/notifications.openapi.json": 2} {
		_, body := a.call(t, "GET", path, nil, "", nil, 200)
		var spec struct {
			Paths map[string]json.RawMessage `json:"paths"`
		}
		authdb.Require(t, "Swagger hai nhóm tích hợp", json.Unmarshal(body, &spec))
		if len(spec.Paths) != count {
			t.Fatal("Swagger mất operation khi gộp DATN-41/42")
		}
	}
	hr := f.AddAccount(t, authdb.AccountOptions{Email: "notify-hr@example.invalid", Roles: []string{"HR", "ADMIN"}})
	actor := f.AddAccount(t, authdb.AccountOptions{Email: "notify-self@example.invalid"})
	f.AddAccount(t, authdb.AccountOptions{Email: "notify-other@example.invalid", Roles: []string{"ADMIN"}})
	f.AddAccount(t, authdb.AccountOptions{Email: "notify-limited@example.invalid", Roles: []string{"HR", "ADMIN"}, MustChangePassword: true})
	repo, e := postgres.NewProfileChangeRepository(f.DB.Conn, 10*time.Second)
	authdb.Require(t, "writer thật", e)
	s := pc.NewService(repo)
	value := "Địa chỉ tổng hợp"
	pending, e := s.Submit(t.Context(), pc.SubmitInput{ActorID: actor.UserID, EmployeeID: actor.EmployeeID, Key: "00000000-0000-4000-8000-000000000001", Before: pc.Values{"address": nil}, Proposed: pc.Values{"address": &value}})
	authdb.Require(t, "submit notify thật", e)
	cookie := a.login(t, "notify-hr@example.invalid", authdb.FixturePassword, nil)
	token := a.csrf(t, cookie)
	_, body := catalogCall(t, a, "GET", "/api/notifications", cookie, "", nil, 200)
	var page struct {
		Items              []map[string]any `json:"items"`
		Total, UnreadCount int64
	}
	var raw map[string]json.RawMessage
	authdb.Require(t, "DTO page", json.Unmarshal(body, &raw))
	if len(raw) != 5 {
		t.Fatal("page DTO ngoài allowlist")
	}
	authdb.Require(t, "items DTO", json.Unmarshal(raw["items"], &page.Items))
	authdb.Require(t, "total", json.Unmarshal(raw["total"], &page.Total))
	authdb.Require(t, "unread", json.Unmarshal(raw["unread_count"], &page.UnreadCount))
	if len(page.Items) != 1 || len(page.Items[0]) != 7 || page.Total != 1 || page.UnreadCount != 1 || page.Items[0]["resource_id"] != float64(pending.RequestID) || page.Items[0]["read_at"] != nil {
		t.Fatal("projection notification sai")
	}
	for _, key := range []string{"recipient_id", "fullname", "email", "phone", "address", "before_values", "proposed_values", "decision_reason"} {
		if _, ok := page.Items[0][key]; ok {
			t.Fatal("DTO lộ PII")
		}
	}
	path := "/api/notifications/" + strconv.Itoa(int(page.Items[0]["notification_id"].(float64))) + "/read"
	other := a.login(t, "notify-other@example.invalid", authdb.FixturePassword, nil)
	catalogCall(t, a, "POST", path, other, a.csrf(t, other), map[string]any{}, 404)
	catalogCall(t, a, "GET", "/api/notifications", nil, "", nil, 401)
	limited := a.login(t, "notify-limited@example.invalid", authdb.FixturePassword, nil)
	for _, method := range []string{"GET", "POST"} {
		target := "/api/notifications"
		if method == "POST" {
			target = path
		}
		_, b := catalogCall(t, a, method, target, limited, a.csrf(t, limited), map[string]any{}, 403)
		assertErrorCode(t, b, "PASSWORD_CHANGE_REQUIRED")
	}
	// Mất quyền resource sau khi nhận vẫn đọc own được; không đổi quyền resource.
	_, e = f.DB.Conn.Exec(t.Context(), `DELETE FROM user_roles WHERE user_id=$1`, hr.UserID)
	authdb.Require(t, "gỡ HR", e)
	for i := 0; i < 2; i++ {
		catalogCall(t, a, "POST", path, cookie, token, map[string]any{}, 204)
	}
	_, body = catalogCall(t, a, "GET", "/api/notifications?filter=unread&page=2147483647", cookie, "", nil, 200)
	authdb.Require(t, "page sau read", json.Unmarshal(body, &raw))
	if string(raw["total"]) != "0" || string(raw["unread_count"]) != "0" || string(raw["items"]) != "[]" {
		t.Fatal("read/filter count sai")
	}
	// Approve/reject writer thật -> owner notification, không chỉ fixture INSERT.
	_, e = f.DB.Conn.Exec(t.Context(), `INSERT INTO user_roles(user_id,role_code) VALUES($1,'HR')`, hr.UserID)
	authdb.Require(t, "khôi phục HR", e)
	_, e = s.Approve(t.Context(), pc.DetailQuery{ActorID: hr.UserID, RequestID: pending.RequestID})
	authdb.Require(t, "approve writer", e)
	phone := "12345"
	rejected, e := s.Submit(t.Context(), pc.SubmitInput{ActorID: actor.UserID, EmployeeID: actor.EmployeeID, Key: "00000000-0000-4000-8000-000000000002", Before: pc.Values{"phone": nil}, Proposed: pc.Values{"phone": &phone}})
	authdb.Require(t, "submit tiếp cho reject", e)
	_, e = s.Reject(t.Context(), pc.RejectInput{ActorID: hr.UserID, RequestID: rejected.RequestID, Reason: "Lý do tổng hợp không được đưa vào thông báo"})
	authdb.Require(t, "reject writer", e)
	owner := a.login(t, "notify-self@example.invalid", authdb.FixturePassword, nil)
	_, body = catalogCall(t, a, "GET", "/api/notifications", owner, "", nil, 200)
	if !strings.Contains(string(body), "PROFILE_CHANGE_APPROVED") || !strings.Contains(string(body), "PROFILE_CHANGE_REJECTED") || strings.Contains(string(body), value) || strings.Contains(string(body), phone) || strings.Contains(string(body), "Lý do tổng hợp") {
		t.Fatal("owner notification/message sai")
	}
	for _, query := range []string{"?filter=", "?filter=read", "?page=0", "?page_size=101", "?page=1&page=2", "?recipient_id=1", "?x=%zz"} {
		catalogCall(t, a, "GET", "/api/notifications"+query, cookie, "", nil, 400)
	}
	for _, tc := range []struct {
		body, ct, origin, csrf, encoding string
		status                           int
	}{
		{"{}", "application/json", testOrigin, token, "", 204},
		{"{}" + strings.Repeat(" ", 65534), "application/json", testOrigin, token, "", 204},
		{"{}" + strings.Repeat(" ", 65535), "application/json", testOrigin, token, "", 413},
		{"{}", "application/json", testOrigin, "", "", 403},
		{"{}", "application/json", "https://wrong.invalid", token, "", 403},
		{"{}", "text/plain", testOrigin, token, "", 415},
		{"{}", "application/json; charset=latin1", testOrigin, token, "", 415},
		{"{}", "application/json", testOrigin, token, "gzip", 415},
		{"null", "application/json", testOrigin, token, "", 400},
		{"[]", "application/json", testOrigin, token, "", 400},
		{"{}{}", "application/json", testOrigin, token, "", 400},
		{`{"read_at":null}`, "application/json", testOrigin, token, "", 400},
	} {
		req, e := http.NewRequestWithContext(t.Context(), "POST", a.server.URL+path, io.NopCloser(strings.NewReader(tc.body)))
		authdb.Require(t, "request thực không Content-Length", e)
		req.ContentLength = -1
		req.AddCookie(cookie)
		req.Header.Set("Content-Type", tc.ct)
		req.Header.Set("Origin", tc.origin)
		req.Header.Set("X-CSRF-Token", tc.csrf)
		if tc.encoding != "" {
			req.Header.Set("Content-Encoding", tc.encoding)
		}
		res, e := a.server.Client().Do(req)
		authdb.Require(t, "HTTP chunked", e)
		data, e := io.ReadAll(res.Body)
		res.Body.Close()
		authdb.Require(t, "response", e)
		if res.StatusCode != tc.status || res.Header.Get("Cache-Control") != "no-store" || res.Header.Get("Set-Cookie") != "" || (tc.status == 204 && len(data) != 0) {
			t.Fatalf("HTTP boundary sai: cần %d, nhận %d", tc.status, res.StatusCode)
		}
	}
	_, e = f.DB.Conn.Exec(t.Context(), `ALTER TABLE notifications RENAME TO notifications_fault`)
	authdb.Require(t, "fault DB test", e)
	_, body = catalogCall(t, a, "GET", "/api/notifications", cookie, "", nil, 500)
	assertErrorCode(t, body, "INTERNAL_ERROR")
	if strings.Contains(string(body), "notifications_fault") {
		t.Fatal("lộ lỗi provider")
	}
	_, e = f.DB.Conn.Exec(t.Context(), `ALTER TABLE notifications_fault RENAME TO notifications`)
	authdb.Require(t, "khôi phục fault", e)
	blocker, e := f.DB.Conn.Begin(t.Context())
	authdb.Require(t, "mở fault timeout", e)
	defer blocker.Rollback(t.Context())
	_, e = blocker.Exec(t.Context(), `LOCK TABLE notifications IN ACCESS EXCLUSIVE MODE`)
	authdb.Require(t, "khóa notification cho timeout thật", e)
	_, body = catalogCall(t, a, "GET", "/api/notifications", cookie, "", nil, 503)
	assertErrorCode(t, body, "SERVICE_UNAVAILABLE")
	authdb.Require(t, "nhả khóa fault", blocker.Rollback(t.Context()))
	a.call(t, "POST", "/api/auth/logout", cookie, token, nil, 204)
	catalogCall(t, a, "GET", "/api/notifications", cookie, "", nil, 401)
}
