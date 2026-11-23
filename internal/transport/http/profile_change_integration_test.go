//go:build integration

package http

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"ai-staffing/internal/repository/postgres"
	"ai-staffing/internal/service/auth"
	pc "ai-staffing/internal/service/profilechange"
	"ai-staffing/internal/testutil/authdb"
	"github.com/jackc/pgx/v5"
)

// Các bí mật dưới đây chỉ dành cho session tổng hợp trong DB case hy sinh.
func pcaHTTPSecret(actor authdb.Account) auth.SessionSecret {
	return auth.SessionSecret(sha256.Sum256([]byte(fmt.Sprintf("PCA HTTP synthetic actor %d", actor.UserID))))
}

func pcaHTTPAccount(t *testing.T, f *authdb.Fixture, options authdb.AccountOptions) authdb.Account {
	t.Helper()
	a := f.AddAccount(t, options)
	secret := pcaHTTPSecret(a)
	hash := auth.SessionHash(sha256.Sum256(secret[:]))
	now := time.Now().UTC()
	f.AddSession(t, a.UserID, hash, now.Add(-time.Minute), now.Add(time.Hour), nil)
	return a
}

func pcaIntegrationRouter(t *testing.T, f *authdb.Fixture, db interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}, timeout time.Duration) http.Handler {
	t.Helper()
	clock := csrfTestClock{time.Now().UTC()}
	repo := postgres.NewAuthRepository(f.DB.Conn, clock)
	requests, err := postgres.NewProfileChangeRepository(db, timeout)
	authdb.Require(t, "dựng adapter PCA thật", err)
	h, err := NewAuthRouter(fakeHealth{}, "dev", loginTestOrigin, AuthServices{
		Sessions: auth.NewService(repo, clock), Login: auth.NewLoginService(repo, clock, 8*time.Hour),
		PasswordChange: auth.NewPasswordChangeService(repo, clock), Clock: clock,
		ProfileChanges: pc.NewService(requests),
	})
	authdb.Require(t, "dựng HTTP với phiên và PCA PostgreSQL thật", err)
	return h
}

func pcaIntegrationRequest(actor authdb.Account, method, path, body string, key int) *http.Request {
	r := passwordChangeRequest(method, path, body)
	secret := pcaHTTPSecret(actor)
	r.Header.Set("Cookie", (&http.Cookie{Name: sessionCookieName, Value: authEncoding.EncodeToString(secret[:])}).String())
	r.Header.Set(csrfHeaderName, authEncoding.EncodeToString(csrfDigest(secret)))
	if key > 0 {
		r.Header.Set("Idempotency-Key", fmt.Sprintf("00000000-0000-4000-8000-%012d", key))
	}
	return r
}

func pcaHTTPCall(t *testing.T, h http.Handler, actor authdb.Account, method, path, body string, key, status int, code string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	h.ServeHTTP(w, pcaIntegrationRequest(actor, method, path, body, key))
	assertCatalogHTTP(t, w, status, code)
	return w
}

func pcaHTTPSubmitPath(a authdb.Account) string {
	return fmt.Sprintf("/api/employees/%d/profile-change-requests", a.EmployeeID)
}

const pcaHTTPBody = `{"expected_before_values":{"address":null},"proposed_values":{"address":"Địa chỉ tổng hợp"},"explanation":" Giải trình tổng hợp "}`

func pcaHTTPID(t *testing.T, w *httptest.ResponseRecorder) int {
	t.Helper()
	var result struct {
		ID int `json:"request_id"`
	}
	authdb.Require(t, "đọc ID request", json.Unmarshal(w.Body.Bytes(), &result))
	if result.ID <= 0 {
		t.Fatal("response thiếu request_id")
	}
	return result.ID
}

func pcaHTTPCount(t *testing.T, f *authdb.Fixture, query string, args ...any) int {
	t.Helper()
	var count int
	authdb.Require(t, "kiểm tác động DB PCA", f.DB.Conn.QueryRow(t.Context(), query, args...).Scan(&count))
	return count
}

func pcaHTTPExec(t *testing.T, f *authdb.Fixture, query string, args ...any) {
	t.Helper()
	_, err := f.DB.Conn.Exec(t.Context(), query, args...)
	authdb.Require(t, "dựng điều kiện DB PCA tổng hợp", err)
}

func TestProfileChangeHTTPFiveRoutesPersistReplayAndProject(t *testing.T) {
	f := authdb.NewFixture(t)
	self := pcaHTTPAccount(t, f, authdb.AccountOptions{Email: "pca-self@example.invalid"})
	hr := pcaHTTPAccount(t, f, authdb.AccountOptions{Email: "pca-hr@example.invalid", Roles: []string{"HR", "ADMIN", "DIRECTOR"}})
	h := pcaIntegrationRouter(t, f, f.DB.Conn, 5*time.Second)
	w := pcaHTTPCall(t, h, self, "POST", pcaHTTPSubmitPath(self), pcaHTTPBody, 1, 201, "")
	id := pcaHTTPID(t, w)
	var detail map[string]any
	authdb.Require(t, "đọc detail allowlist", json.Unmarshal(w.Body.Bytes(), &detail))
	if len(detail) != 14 || detail["requested_by"] != float64(self.UserID) || detail["employee_id"] != float64(self.EmployeeID) || detail["explanation"] != "Giải trình tổng hợp" || detail["resolved_by"] != nil || detail["decision_reason"] != nil {
		t.Fatal("detail không đúng projection/NULL hoặc ghép nhầm user và employee")
	}
	if len(detail["current_values"].(map[string]any)) != 1 || detail["current_values"].(map[string]any)["address"] != nil {
		t.Fatal("current_values phải chỉ có key patch, giữ NULL")
	}
	for _, hidden := range []string{"schema_version", "password", "token_hash", "company_email", "system_roles", "changed_fields"} {
		if _, found := detail[hidden]; found {
			t.Fatal("detail vượt allowlist")
		}
	}
	if pcaHTTPCount(t, f, `SELECT count(*) FROM employees WHERE employee_id=$1 AND address IS NULL`, self.EmployeeID) != 1 || pcaHTTPCount(t, f, `SELECT count(*) FROM notifications WHERE recipient_id=$1`, hr.UserID) != 1 {
		t.Fatal("submit đã sửa official hoặc nhân thông báo HR nhiều role")
	}
	w = pcaHTTPCall(t, h, self, "POST", pcaHTTPSubmitPath(self), pcaHTTPBody, 1, 201, "")
	if pcaHTTPID(t, w) != id || pcaHTTPCount(t, f, `SELECT count(*) FROM notifications`) != 1 {
		t.Fatal("replay tạo request hoặc notification mới")
	}
	path := "/api/profile-change-requests/" + strconv.Itoa(id)
	pcaHTTPCall(t, h, self, "GET", path, "", 0, 200, "")
	for _, query := range []struct {
		actor        authdb.Account
		suffix       string
		total, items int
	}{
		{self, "", 1, 1}, {hr, "", 0, 0}, {hr, "?scope=review&status=PENDING&page_size=1", 1, 1}, {self, "?page=99", 1, 0},
	} {
		w = pcaHTTPCall(t, h, query.actor, "GET", "/api/profile-change-requests"+query.suffix, "", 0, 200, "")
		var page struct {
			Items    []map[string]any `json:"items"`
			Total    int              `json:"total"`
			Page     int              `json:"page"`
			PageSize int              `json:"page_size"`
		}
		authdb.Require(t, "đọc trang", json.Unmarshal(w.Body.Bytes(), &page))
		if page.Total != query.total || len(page.Items) != query.items || page.Page < 1 || page.PageSize < 1 {
			t.Fatal("scope/filter/page rỗng sai hoặc mất total")
		}
		for _, item := range page.Items {
			if len(item) != 8 || item["changed_fields"] == nil {
				t.Fatal("list phải có đúng projection tám trường")
			}
			for _, hidden := range []string{"address", "phone", "explanation", "proposed_values", "before_values", "current_values"} {
				if _, found := item[hidden]; found {
					t.Fatal("list lộ PII chi tiết")
				}
			}
		}
	}
	w = pcaHTTPCall(t, h, hr, "POST", path+"/approve", `{}`, 0, 200, "")
	if pcaHTTPID(t, w) != id || pcaHTTPCount(t, f, `SELECT count(*) FROM employees WHERE employee_id=$1 AND address='Địa chỉ tổng hợp'`, self.EmployeeID) != 1 {
		t.Fatal("approve không cập nhật official đúng request")
	}
	pcaHTTPCall(t, h, hr, "POST", path+"/reject", `{"reason":"Không áp dụng"}`, 0, 409, "REQUEST_ALREADY_RESOLVED")
	w = pcaHTTPCall(t, h, self, "POST", pcaHTTPSubmitPath(self), pcaHTTPBody, 1, 201, "")
	authdb.Require(t, "đọc replay terminal", json.Unmarshal(w.Body.Bytes(), &detail))
	if detail["status"] != "APPROVED" || detail["resolved_by"] != float64(hr.UserID) {
		t.Fatal("replay terminal không giữ kết quả hiện hành")
	}
	phoneBody := `{"expected_before_values":{"phone":null},"proposed_values":{"phone":"12345"}}`
	w = pcaHTTPCall(t, h, self, "POST", pcaHTTPSubmitPath(self), phoneBody, 2, 201, "")
	second := pcaHTTPID(t, w)
	pcaHTTPCall(t, h, hr, "POST", fmt.Sprintf("/api/profile-change-requests/%d/reject", second), `{"reason":" Lý do tổng hợp "}`, 0, 200, "")
	if pcaHTTPCount(t, f, `SELECT count(*) FROM employees WHERE employee_id=$1 AND phone IS NULL`, self.EmployeeID) != 1 || pcaHTTPCount(t, f, `SELECT count(*) FROM notifications WHERE recipient_id=$1`, self.UserID) != 2 || pcaHTTPCount(t, f, `SELECT count(*) FROM employee_profile_change_requests WHERE status='REJECTED' AND decision_reason='Lý do tổng hợp'`) != 1 {
		t.Fatal("reject không nguyên tử hoặc sửa official")
	}
}

func TestProfileChangeHTTPCurrentServerRightsAndSessions(t *testing.T) {
	f := authdb.NewFixture(t)
	self := pcaHTTPAccount(t, f, authdb.AccountOptions{Email: "pca-self@example.invalid"})
	hr := pcaHTTPAccount(t, f, authdb.AccountOptions{Email: "pca-hr@example.invalid", Roles: []string{"HR"}})
	admin := pcaHTTPAccount(t, f, authdb.AccountOptions{Email: "pca-admin@example.invalid", Roles: []string{"ADMIN"}})
	director := pcaHTTPAccount(t, f, authdb.AccountOptions{Email: "pca-director@example.invalid", Roles: []string{"DIRECTOR"}})
	h := pcaIntegrationRouter(t, f, f.DB.Conn, 5*time.Second)
	id := pcaHTTPID(t, pcaHTTPCall(t, h, self, "POST", pcaHTTPSubmitPath(self), pcaHTTPBody, 1, 201, ""))
	path := fmt.Sprintf("/api/profile-change-requests/%d", id)
	for _, actor := range []authdb.Account{self, admin, director} {
		pcaHTTPCall(t, h, actor, "GET", "/api/profile-change-requests?scope=review", "", 0, 403, "FORBIDDEN")
		for _, decision := range []string{"approve", "reject"} {
			body := `{}`
			if decision == "reject" {
				body = `{"reason":"Lý do"}`
			}
			pcaHTTPCall(t, h, actor, "POST", path+"/"+decision, body, 0, 403, "FORBIDDEN")
		}
	}
	for _, actor := range []authdb.Account{admin, director} {
		for _, target := range []string{path, "/api/profile-change-requests/2147483647"} {
			pcaHTTPCall(t, h, actor, "GET", target, "", 0, 404, "NOT_FOUND")
		}
		pcaHTTPCall(t, h, actor, "POST", pcaHTTPSubmitPath(self), pcaHTTPBody, 1, 404, "NOT_FOUND")
	}
	for _, gate := range []struct {
		sql    string
		status int
		code   string
	}{
		{`UPDATE users SET must_change_password=true WHERE user_id=$1`, 403, "PASSWORD_CHANGE_REQUIRED"},
		{`UPDATE users SET must_change_password=false,status='LOCKED' WHERE user_id=$1`, 401, "UNAUTHENTICATED"},
		{`UPDATE users SET status='ACTIVE' WHERE user_id=$1`, 200, ""},
		{`UPDATE employees SET employment_status='INACTIVE' WHERE user_id=$1`, 401, "UNAUTHENTICATED"},
		{`UPDATE employees SET employment_status='ACTIVE' WHERE user_id=$1`, 200, ""},
		{`UPDATE auth_sessions SET revoked_at=clock_timestamp() WHERE user_id=$1`, 401, "UNAUTHENTICATED"},
	} {
		pcaHTTPExec(t, f, gate.sql, hr.UserID)
		pcaHTTPCall(t, h, hr, "GET", path, "", 0, gate.status, gate.code)
	}
	if pcaHTTPCount(t, f, `SELECT count(*) FROM employee_profile_change_requests WHERE status='PENDING'`) != 1 || pcaHTTPCount(t, f, `SELECT count(*) FROM notifications`) != 1 {
		t.Fatal("gates đã phát sinh mutation")
	}
}

func TestProfileChangeHTTPStrictInputAndActualBodyBoundary(t *testing.T) {
	f := authdb.NewFixture(t)
	a := pcaHTTPAccount(t, f, authdb.AccountOptions{Email: "pca-input@example.invalid", Roles: []string{"HR"}})
	h := pcaIntegrationRouter(t, f, f.DB.Conn, 5*time.Second)
	for _, tc := range []struct {
		name, body string
		edit       func(*http.Request)
		status     int
		code       string
	}{
		{"duplicate root", strings.TrimSuffix(pcaHTTPBody, "}") + `,"explanation":"x"}`, nil, 400, "VALIDATION_ERROR"},
		{"duplicate nested", `{"expected_before_values":{"address":null,"address":null},"proposed_values":{"address":"x"}}`, nil, 400, "VALIDATION_ERROR"},
		{"unknown root", strings.TrimSuffix(pcaHTTPBody, "}") + `,"actor_id":1}`, nil, 400, "VALIDATION_ERROR"},
		{"unknown nested", `{"expected_before_values":{"company_email":"a"},"proposed_values":{"company_email":"b"}}`, nil, 400, "VALIDATION_ERROR"},
		{"schema version", `{"expected_before_values":{"address":null},"proposed_values":{"address":"x","schema_version":1}}`, nil, 400, "VALIDATION_ERROR"},
		{"surrogate", strings.Replace(pcaHTTPBody, "Địa chỉ tổng hợp", `\ud800`, 1), nil, 400, "VALIDATION_ERROR"},
		{"NUL", strings.Replace(pcaHTTPBody, "Địa chỉ tổng hợp", `\u0000`, 1), nil, 400, "VALIDATION_ERROR"},
		{"UTF8", strings.Replace(pcaHTTPBody, "Địa chỉ tổng hợp", string([]byte{0xff}), 1), nil, 400, "VALIDATION_ERROR"},
		{"null object", `{"expected_before_values":null,"proposed_values":{"address":"x"}}`, nil, 400, "VALIDATION_ERROR"},
		{"address null", `{"expected_before_values":{"address":null},"proposed_values":{"address":null}}`, nil, 400, "VALIDATION_ERROR"},
		{"trailing", pcaHTTPBody + `{}`, nil, 400, "VALIDATION_ERROR"},
		{"array", `[]`, nil, 400, "VALIDATION_ERROR"},
		{"missing key", pcaHTTPBody, func(r *http.Request) { r.Header.Del("Idempotency-Key") }, 400, "VALIDATION_ERROR"},
		{"wrong version", pcaHTTPBody, func(r *http.Request) { r.Header.Set("Idempotency-Key", "00000000-0000-1000-8000-000000000001") }, 400, "VALIDATION_ERROR"},
		{"duplicate key header", pcaHTTPBody, func(r *http.Request) { r.Header.Add("Idempotency-Key", r.Header.Get("Idempotency-Key")) }, 400, "VALIDATION_ERROR"},
		{"comma key", pcaHTTPBody, func(r *http.Request) {
			r.Header.Set("Idempotency-Key", r.Header.Get("Idempotency-Key")+","+r.Header.Get("Idempotency-Key"))
		}, 400, "VALIDATION_ERROR"},
		{"Origin", pcaHTTPBody, func(r *http.Request) { r.Header.Set("Origin", "https://other.invalid") }, 403, "FORBIDDEN"},
		{"CSRF", pcaHTTPBody, func(r *http.Request) { r.Header.Del(csrfHeaderName) }, 403, "FORBIDDEN"},
		{"no cookie", pcaHTTPBody, func(r *http.Request) { r.Header.Del("Cookie") }, 401, "UNAUTHENTICATED"},
		{"media", pcaHTTPBody, func(r *http.Request) { r.Header.Set("Content-Type", "text/plain") }, 415, "UNSUPPORTED_MEDIA_TYPE"},
		{"charset", pcaHTTPBody, func(r *http.Request) { r.Header.Set("Content-Type", "application/json; charset=iso-8859-1") }, 415, "UNSUPPORTED_MEDIA_TYPE"},
		{"encoding", pcaHTTPBody, func(r *http.Request) { r.Header.Set("Content-Encoding", "gzip") }, 415, "UNSUPPORTED_MEDIA_TYPE"},
		{"query", pcaHTTPBody, func(r *http.Request) { r.URL.RawQuery = "scope=self" }, 400, "VALIDATION_ERROR"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := pcaIntegrationRequest(a, "POST", pcaHTTPSubmitPath(a), tc.body, 1)
			if tc.edit != nil {
				tc.edit(r)
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			assertCatalogHTTP(t, w, tc.status, tc.code)
		})
	}
	for _, suffix := range []string{"?page=0", "?page=2147483648", "?page=1&page=2", "?page_size=101", "?scope=all", "?scope=review&scope=self", "?status=pending", "?actor_id=1", "?page=%zz"} {
		pcaHTTPCall(t, h, a, "GET", "/api/profile-change-requests"+suffix, "", 0, 400, "VALIDATION_ERROR")
	}
	if pcaHTTPCount(t, f, `SELECT count(*) FROM employee_profile_change_requests`) != 0 || pcaHTTPCount(t, f, `SELECT count(*) FROM notifications`) != 0 || pcaHTTPCount(t, f, `SELECT count(*) FROM idempotency_records`) != 0 {
		t.Fatal("input sai đã ghi DB")
	}
	body := pcaHTTPBody + strings.Repeat(" ", 65536-len(pcaHTTPBody))
	pcaHTTPCall(t, h, a, "POST", pcaHTTPSubmitPath(a), body+" ", 1, 413, "PAYLOAD_TOO_LARGE")
	w := pcaHTTPCall(t, h, a, "POST", pcaHTTPSubmitPath(a), body, 1, 201, "")
	id := pcaHTTPID(t, w)
	path := fmt.Sprintf("/api/profile-change-requests/%d", id)
	for _, action := range []struct{ suffix, body string }{{"/approve", `{}`}, {"/reject", `{"reason":"Lý do"}`}} {
		oversized := action.body + strings.Repeat(" ", 65537-len(action.body))
		pcaHTTPCall(t, h, a, "POST", path+action.suffix, oversized, 0, 413, "PAYLOAD_TOO_LARGE")
	}
	if pcaHTTPCount(t, f, `SELECT count(*) FROM employee_profile_change_requests WHERE status='PENDING'`) != 1 {
		t.Fatal("body lớn đã quyết định request")
	}
	if pcaHTTPCount(t, f, `SELECT count(*) FROM idempotency_records`) != 1 {
		t.Fatal("body boundary tạo key sai")
	}
}

func TestProfileChangeHTTPConflictCodesAndTargetGate(t *testing.T) {
	f := authdb.NewFixture(t)
	a := pcaHTTPAccount(t, f, authdb.AccountOptions{Email: "pca-conflict@example.invalid"})
	hr := pcaHTTPAccount(t, f, authdb.AccountOptions{Email: "pca-hr@example.invalid", Roles: []string{"HR"}})
	h := pcaIntegrationRouter(t, f, f.DB.Conn, 5*time.Second)
	id := pcaHTTPID(t, pcaHTTPCall(t, h, a, "POST", pcaHTTPSubmitPath(a), pcaHTTPBody, 1, 201, ""))
	path := fmt.Sprintf("/api/profile-change-requests/%d", id)
	pcaHTTPCall(t, h, a, "POST", pcaHTTPSubmitPath(a), strings.Replace(pcaHTTPBody, "Địa chỉ tổng hợp", "Địa chỉ khác", 1), 1, 409, "IDEMPOTENCY_KEY_REUSED")
	pcaHTTPExec(t, f, `UPDATE idempotency_records SET created_at=clock_timestamp()-interval '2 days'`)
	pcaHTTPExec(t, f, `UPDATE idempotency_records SET expires_at=created_at+interval '24 hours'`)
	pcaHTTPCall(t, h, a, "POST", pcaHTTPSubmitPath(a), pcaHTTPBody, 1, 409, "IDEMPOTENCY_KEY_EXPIRED")
	pcaHTTPExec(t, f, `UPDATE employees SET address='Official đã đổi' WHERE employee_id=$1`, a.EmployeeID)
	pcaHTTPCall(t, h, hr, "POST", path+"/approve", `{}`, 0, 409, "STALE_DATA")
	pcaHTTPCall(t, h, a, "POST", pcaHTTPSubmitPath(a), pcaHTTPBody, 2, 409, "STALE_DATA")
	pcaHTTPExec(t, f, `UPDATE employees SET employment_status='INACTIVE' WHERE employee_id=$1`, a.EmployeeID)
	pcaHTTPCall(t, h, hr, "POST", path+"/approve", `{}`, 0, 409, "EMPLOYEE_INACTIVE")
	pcaHTTPCall(t, h, hr, "POST", path+"/reject", `{"reason":"Không còn áp dụng"}`, 0, 200, "")
	pcaHTTPCall(t, h, hr, "POST", path+"/approve", `{}`, 0, 409, "REQUEST_ALREADY_RESOLVED")
	if pcaHTTPCount(t, f, `SELECT count(*) FROM employee_profile_change_requests`) != 1 || pcaHTTPCount(t, f, `SELECT count(*) FROM notifications`) != 2 || pcaHTTPCount(t, f, `SELECT count(*) FROM employees WHERE employee_id=$1 AND address='Official đã đổi'`, a.EmployeeID) != 1 {
		t.Fatal("conflict để lại ghi dở hoặc ghi đè official")
	}
}

func TestProfileChangeHTTPDatabaseFaultAndUnknownCommit(t *testing.T) {
	for _, mode := range []string{"notification rollback", "lost submit commit", "lost decision commit"} {
		t.Run(mode, func(t *testing.T) {
			f := authdb.NewFixture(t)
			a := pcaHTTPAccount(t, f, authdb.AccountOptions{Email: "pca-fault@example.invalid", Roles: []string{"HR"}})
			h := pcaIntegrationRouter(t, f, f.DB.Conn, 5*time.Second)
			id := 0
			if mode == "lost decision commit" {
				id = pcaHTTPID(t, pcaHTTPCall(t, h, a, "POST", pcaHTTPSubmitPath(a), pcaHTTPBody, 1, 201, ""))
			}
			if mode == "notification rollback" {
				pcaHTTPExec(t, f, `CREATE FUNCTION pca_http_fault() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'synthetic confidential provider detail' USING ERRCODE='23514'; END $$`)
				pcaHTTPExec(t, f, `CREATE TRIGGER pca_http_fault BEFORE INSERT ON notifications FOR EACH ROW EXECUTE FUNCTION pca_http_fault()`)
			} else {
				h = pcaIntegrationRouter(t, f, accountHTTPLostReplyDB{f.DB.Conn}, 5*time.Second)
			}
			method, path, body, key := "POST", pcaHTTPSubmitPath(a), pcaHTTPBody, 1
			if mode == "lost decision commit" {
				path = fmt.Sprintf("/api/profile-change-requests/%d/approve", id)
				body = `{}`
				key = 0
			}
			status, code := 503, "SERVICE_UNAVAILABLE"
			if mode == "notification rollback" {
				status, code = 500, "INTERNAL_ERROR"
			}
			w := pcaHTTPCall(t, h, a, method, path, body, key, status, code)
			if strings.Contains(w.Body.String(), "synthetic") || strings.Contains(w.Body.String(), "request_id\":") && strings.Contains(w.Body.String(), "employee_id") {
				t.Fatal("fault lộ dữ liệu nội bộ hoặc success giả")
			}
			if mode == "notification rollback" {
				for _, table := range []string{"employee_profile_change_requests", "notifications", "idempotency_records"} {
					if pcaHTTPCount(t, f, "SELECT count(*) FROM "+table) != 0 {
						t.Fatal("lỗi notification không rollback đầy đủ")
					}
				}
				return
			}
			// Wrapper chỉ làm mất xác nhận sau commit PG thật; không mô phỏng mất mạng thật.
			if pcaHTTPCount(t, f, `SELECT count(*) FROM employee_profile_change_requests`) != 1 {
				t.Fatal("unknown đã tự retry hoặc không commit thật")
			}
			normal := pcaIntegrationRouter(t, f, f.DB.Conn, 5*time.Second)
			w = pcaHTTPCall(t, normal, a, "GET", "/api/profile-change-requests", "", 0, 200, "")
			var page struct {
				Items []struct {
					ID     int    `json:"request_id"`
					Status string `json:"status"`
				} `json:"items"`
			}
			authdb.Require(t, "đọc list đối soát", json.Unmarshal(w.Body.Bytes(), &page))
			if len(page.Items) != 1 || (mode == "lost decision commit" && page.Items[0].Status != "APPROVED") {
				t.Fatal("list không đối soát được commit")
			}
			if mode == "lost submit commit" {
				w = pcaHTTPCall(t, normal, a, "POST", pcaHTTPSubmitPath(a), pcaHTTPBody, 1, 201, "")
				if pcaHTTPID(t, w) != page.Items[0].ID || pcaHTTPCount(t, f, `SELECT count(*) FROM notifications`) != 1 {
					t.Fatal("đối soát key sinh tác động mới")
				}
			} else if pcaHTTPCount(t, f, `SELECT count(*) FROM notifications`) != 2 || pcaHTTPCount(t, f, `SELECT count(*) FROM employees WHERE employee_id=$1 AND address='Địa chỉ tổng hợp'`, a.EmployeeID) != 1 {
				t.Fatal("decision commit không atomic")
			}
		})
	}
}

func TestProfileChangeHTTPDatabaseTimeoutIs503(t *testing.T) {
	f := authdb.NewFixture(t)
	a := pcaHTTPAccount(t, f, authdb.AccountOptions{Email: "pca-timeout@example.invalid"})
	pool := f.DB.Pool(t)
	defer pool.Close()
	locker := f.DB.Connect(t)
	tx, err := locker.Begin(t.Context())
	authdb.Require(t, "mở barrier DB", err)
	defer tx.Rollback(t.Context())
	_, err = tx.Exec(t.Context(), `LOCK TABLE employee_profile_change_requests IN ACCESS EXCLUSIVE MODE`)
	authdb.Require(t, "khóa trước HTTP read", err)
	h := pcaIntegrationRouter(t, f, pool, 100*time.Millisecond)
	pcaHTTPCall(t, h, a, "GET", "/api/profile-change-requests", "", 0, 503, "SERVICE_UNAVAILABLE")
}
