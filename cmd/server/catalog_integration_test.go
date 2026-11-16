//go:build integration

package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"ai-staffing/internal/config"
	"ai-staffing/internal/testutil/authdb"
)

type catalogDTO struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func catalogCall(t *testing.T, a *authApp, method, path string, cookie *http.Cookie, token string, body any, status int) (http.Header, []byte) {
	t.Helper()
	header, data := a.call(t, method, path, cookie, token, body, status)
	if header.Get("Cache-Control") != "no-store" || header.Get("Set-Cookie") != "" || header.Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("header catalog sai")
	}
	if status == 204 && (len(data) != 0 || header.Get("Content-Type") != "") {
		t.Fatal("204 phải không JSON")
	}
	return header, data
}
func TestCatalogApplicationPostgresCRUD(t *testing.T) {
	a := newAuthApp(t)
	ctx := t.Context()
	f := a.fixture
	hr := f.AddAccount(t, authdb.AccountOptions{Email: "catalog-hr@example.invalid", Roles: []string{"HR"}})
	f.AddAccount(t, authdb.AccountOptions{Email: "catalog-admin@example.invalid", Roles: []string{"ADMIN"}})
	f.AddAccount(t, authdb.AccountOptions{Email: "catalog-plain@example.invalid"})
	f.AddAccount(t, authdb.AccountOptions{Email: "catalog-reference@example.invalid"})
	f.AddAccount(t, authdb.AccountOptions{Email: "catalog-limited@example.invalid", Roles: []string{"HR", "ADMIN"}, MustChangePassword: true})
	hrCookie := a.login(t, "catalog-hr@example.invalid", authdb.FixturePassword, nil)
	hrToken := a.csrf(t, hrCookie)
	adminCookie := a.login(t, "catalog-admin@example.invalid", authdb.FixturePassword, nil)
	adminToken := a.csrf(t, adminCookie)
	plain := a.login(t, "catalog-plain@example.invalid", authdb.FixturePassword, nil)
	limited := a.login(t, "catalog-limited@example.invalid", authdb.FixturePassword, nil)
	for _, collection := range []string{"departments", "positions", "job-levels"} {
		t.Run(collection, func(t *testing.T) {
			path := "/api/" + collection
			header, body := catalogCall(t, a, "POST", path, hrCookie, hrToken, map[string]string{"name": "  Synthetic %_ name  "}, 201)
			var item catalogDTO
			authdb.Require(t, "decode catalog item", json.Unmarshal(body, &item))
			if item.ID < 1 || item.Name != "Synthetic %_ name" {
				t.Fatal("DTO tạo không đúng")
			}
			var fields map[string]any
			authdb.Require(t, "allowlist DTO", json.Unmarshal(body, &fields))
			if len(fields) != 2 {
				t.Fatal("DTO vượt id/name")
			}
			uri := path + "/" + strconv.Itoa(item.ID)
			if header.Get("Location") != uri {
				t.Fatal("Location sai")
			}
			_, body = catalogCall(t, a, "GET", path+"?q=%25_&page_size=1", adminCookie, "", nil, 200)
			var page struct {
				Items      []catalogDTO `json:"items"`
				Pagination struct {
					Page     int `json:"page"`
					PageSize int `json:"page_size"`
					Total    int `json:"total"`
				} `json:"pagination"`
			}
			authdb.Require(t, "decode page", json.Unmarshal(body, &page))
			if len(page.Items) != 1 || page.Items[0] != item || page.Pagination.Total != 1 || page.Pagination.Page != 1 || page.Pagination.PageSize != 1 {
				t.Fatal("list tìm literal không khớp")
			}
			for _, method := range []string{"POST", "PATCH", "DELETE"} {
				target := path
				payload := map[string]string{"name": "Admin must not write"}
				if method != "POST" {
					target = uri
					payload["expected_before_name"] = item.Name
				}
				if method == "DELETE" {
					delete(payload, "name")
				}
				_, data := catalogCall(t, a, method, target, adminCookie, adminToken, payload, 403)
				assertErrorCode(t, data, "FORBIDDEN")
			}
			for _, tc := range []struct {
				cookie *http.Cookie
				code   string
				status int
			}{{plain, "FORBIDDEN", 403}, {limited, "PASSWORD_CHANGE_REQUIRED", 403}, {nil, "UNAUTHENTICATED", 401}} {
				_, data := catalogCall(t, a, "GET", path, tc.cookie, "", nil, tc.status)
				assertErrorCode(t, data, tc.code)
			}
			_, body = catalogCall(t, a, "POST", path, hrCookie, hrToken, map[string]string{"name": "synthetic %_ NAME"}, 409)
			assertErrorCode(t, body, "DUPLICATE_NAME")
			_, body = catalogCall(t, a, "PATCH", uri, hrCookie, hrToken, map[string]string{"name": "Updated", "expected_before_name": "stale"}, 409)
			assertErrorCode(t, body, "STALE_DATA")
			_, body = catalogCall(t, a, "PATCH", uri, hrCookie, hrToken, map[string]string{"name": "Updated", "expected_before_name": item.Name}, 200)
			var renamed catalogDTO
			authdb.Require(t, "rename DTO", json.Unmarshal(body, &renamed))
			if renamed.ID != item.ID || renamed.Name != "Updated" {
				t.Fatal("rename sai DTO")
			}
			catalogCall(t, a, "DELETE", uri, hrCookie, hrToken, map[string]string{"expected_before_name": "Updated"}, 204)
			_, body = catalogCall(t, a, "DELETE", uri, hrCookie, hrToken, map[string]string{"expected_before_name": "Updated"}, 404)
			assertErrorCode(t, body, "NOT_FOUND")
			var n int
			entity := map[string]string{"departments": "department", "positions": "position", "job-levels": "job_level"}[collection]
			authdb.Require(t, "audit sau HTTP CRUD", f.DB.Conn.QueryRow(ctx, `SELECT count(*) FROM audit_events WHERE entity_type=$1 AND entity_id=$2 AND actor_id=$3`, entity, item.ID, hr.UserID).Scan(&n))
			if n != 3 {
				t.Fatal("audit CRUD không đủ hoặc lỗi đã ghi audit")
			}
			_, body = catalogCall(t, a, "POST", path, hrCookie, hrToken, map[string]string{"name": "Referenced"}, 201)
			authdb.Require(t, "decode reference", json.Unmarshal(body, &item))
			uri = path + "/" + strconv.Itoa(item.ID)
			column := map[string]string{"departments": "department_id", "positions": "position_id", "job-levels": "job_level_id"}[collection]
			_, err := f.DB.Conn.Exec(ctx, "UPDATE employees SET "+column+"=$1,employment_status='INACTIVE' WHERE email='catalog-reference@example.invalid'", item.ID)
			authdb.Require(t, "reference inactive", err)
			_, body = catalogCall(t, a, "DELETE", uri, hrCookie, hrToken, map[string]string{"expected_before_name": item.Name}, 409)
			assertErrorCode(t, body, "CATALOG_IN_USE")
		})
	}
	// Tổ hợp role cộng quyền từ DB, không cần đăng nhập lại để thay claim client.
	_, err := f.DB.Conn.Exec(ctx, `INSERT INTO user_roles(user_id,role_code) SELECT user_id,'HR' FROM employees WHERE email='catalog-admin@example.invalid'`)
	authdb.Require(t, "thêm HR fixture", err)
	catalogCall(t, a, "POST", "/api/departments", adminCookie, adminToken, map[string]string{"name": "Combined roles"}, 201)
	_, err = f.DB.Conn.Exec(ctx, `DELETE FROM user_roles WHERE user_id=$1`, hr.UserID)
	authdb.Require(t, "thu quyền HR fixture", err)
	_, body := catalogCall(t, a, "POST", "/api/departments", hrCookie, hrToken, map[string]string{"name": "Revoked role"}, 403)
	assertErrorCode(t, body, "FORBIDDEN")
}

func TestCatalogApplicationChunkedBodies(t *testing.T) {
	a := newAuthApp(t)
	a.fixture.AddAccount(t, authdb.AccountOptions{Email: "chunked-hr@example.invalid", Roles: []string{"HR"}})
	cookie := a.login(t, "chunked-hr@example.invalid", authdb.FixturePassword, nil)
	token := a.csrf(t, cookie)
	for _, collection := range []string{"departments", "positions", "job-levels"} {
		path := "/api/" + collection
		_, data := catalogCall(t, a, "POST", path, cookie, token, map[string]string{"name": "Before"}, 201)
		var item catalogDTO
		authdb.Require(t, "item chunked", json.Unmarshal(data, &item))
		for _, method := range []string{"POST", "PATCH", "DELETE"} {
			target, body, want := path, `{"name":"Chunked"}`, 201
			if method == "PATCH" {
				target += "/" + strconv.Itoa(item.ID)
				body = `{"name":"Renamed","expected_before_name":"Before"}`
				want = 200
			}
			if method == "DELETE" {
				target += "/" + strconv.Itoa(item.ID)
				body = `{"expected_before_name":"Renamed"}`
				want = 204
			}
			padded := body + strings.Repeat(" ", 8192-len(body))
			for _, size := range []int{8193, 8192} {
				payload := padded
				if size == 8193 {
					payload += " "
				}
				req, err := http.NewRequestWithContext(t.Context(), method, a.server.URL+target, strings.NewReader(payload))
				authdb.Require(t, "request chunked", err)
				req.ContentLength = -1
				req.AddCookie(cookie)
				req.Header.Set("Origin", testOrigin)
				req.Header.Set("X-CSRF-Token", token)
				req.Header.Set("Content-Type", "application/json")
				response, err := a.server.Client().Do(req)
				authdb.Require(t, "HTTP chunked thật", err)
				raw, err := io.ReadAll(response.Body)
				response.Body.Close()
				authdb.Require(t, "body chunked response", err)
				expected := want
				if size == 8193 {
					expected = 413
				}
				if response.StatusCode != expected || response.Header.Get("Cache-Control") != "no-store" {
					t.Fatalf("%s body %d: HTTP %d thay vì %d", method, size, response.StatusCode, expected)
				}
				if size == 8193 {
					assertErrorCode(t, raw, "PAYLOAD_TOO_LARGE")
				}
			}
		}
	}
}

func TestCatalogApplicationDatabaseFailure(t *testing.T) {
	f := authdb.NewFixture(t)
	f.AddAccount(t, authdb.AccountOptions{Email: "db-error-hr@example.invalid", Roles: []string{"HR"}})
	clock := &integrationClock{}
	clock.nanos.Store(time.Date(2026, 11, 16, 10, 0, 0, 0, time.UTC).UnixNano())
	pool := f.DB.Pool(t)
	h, err := buildHandler(&config.Config{AppEnv: "dev", AppOrigin: testOrigin, SessionTTL: 8 * time.Hour}, pool, clock)
	authdb.Require(t, "composition DB lỗi", err)
	server := httptest.NewServer(h)
	t.Cleanup(server.Close)
	a := &authApp{fixture: f, server: server, clock: clock}
	cookie := a.login(t, "db-error-hr@example.invalid", authdb.FixturePassword, nil)
	token := a.csrf(t, cookie)
	// Fault DDL chỉ trên case sở hữu: auth còn hoạt động, query danh mục lỗi thật.
	_, err = f.DB.Conn.Exec(t.Context(), `ALTER TABLE departments RENAME TO departments_unavailable`)
	authdb.Require(t, "lỗi catalog tổng hợp", err)
	_, data := catalogCall(t, a, "GET", "/api/departments", cookie, "", nil, 500)
	assertErrorCode(t, data, "INTERNAL_ERROR")
	pool.Close()
	_, data = catalogCall(t, a, "GET", "/api/positions", cookie, "", nil, 500)
	assertErrorCode(t, data, "INTERNAL_ERROR")
	_, data = catalogCall(t, a, "POST", "/api/positions", cookie, token, map[string]string{"name": "Cannot write"}, 500)
	assertErrorCode(t, data, "INTERNAL_ERROR")
}

func TestCatalogApplicationSecurityAndAuditFailure(t *testing.T) {
	a := newAuthApp(t)
	a.fixture.AddAccount(t, authdb.AccountOptions{Email: "security-hr@example.invalid", Roles: []string{"HR"}})
	cookie := a.login(t, "security-hr@example.invalid", authdb.FixturePassword, nil)
	oldToken := a.csrf(t, cookie)
	fresh := a.login(t, "security-hr@example.invalid", authdb.FixturePassword, cookie)
	token := a.csrf(t, fresh)
	for _, collection := range []string{"departments", "positions", "job-levels"} {
		path := "/api/" + collection
		_, body := catalogCall(t, a, "POST", path, fresh, oldToken, map[string]string{"name": "Wrong CSRF"}, 403)
		assertErrorCode(t, body, "FORBIDDEN")
		_, body = catalogCall(t, a, "POST", path, cookie, token, map[string]string{"name": "Revoked cookie"}, 401)
		assertErrorCode(t, body, "UNAUTHENTICATED")
		for _, tc := range []struct {
			payload string
			edit    func(*http.Request)
			status  int
			code    string
		}{
			{`{"name":"Wrong origin"}`, func(r *http.Request) { r.Header.Set("Origin", "https://evil.invalid") }, 403, "FORBIDDEN"},
			{`{"name":"Missing CSRF"}`, func(r *http.Request) { r.Header.Del("X-CSRF-Token") }, 403, "FORBIDDEN"},
			{`{"name":"Text body"}`, func(r *http.Request) { r.Header.Set("Content-Type", "text/plain") }, 415, "UNSUPPORTED_MEDIA_TYPE"},
			{`{"name":"Unknown","role":"HR"}`, nil, 400, "VALIDATION_ERROR"},
			{`{"name":"First","name":"Second"}`, nil, 400, "VALIDATION_ERROR"},
			{`{"name":"\ud800"}`, nil, 400, "VALIDATION_ERROR"},
			{`{"name":"` + strings.Repeat("ế", 101) + `"}`, nil, 400, "VALIDATION_ERROR"},
		} {
			req, err := http.NewRequestWithContext(t.Context(), "POST", a.server.URL+path, strings.NewReader(tc.payload))
			authdb.Require(t, "security request", err)
			req.ContentLength = -1
			req.AddCookie(fresh)
			req.Header.Set("Origin", testOrigin)
			req.Header.Set("X-CSRF-Token", token)
			req.Header.Set("Content-Type", "application/json")
			if tc.edit != nil {
				tc.edit(req)
			}
			response, err := a.server.Client().Do(req)
			authdb.Require(t, "security HTTP", err)
			data, err := io.ReadAll(response.Body)
			response.Body.Close()
			authdb.Require(t, "security response", err)
			if response.StatusCode != tc.status || response.Header.Get("Cache-Control") != "no-store" {
				t.Fatal("security status/header không đúng")
			}
			assertErrorCode(t, data, tc.code)
		}
	}
	var audits int
	authdb.Require(t, "không ghi khi request sai", a.fixture.DB.Conn.QueryRow(t.Context(), `SELECT count(*) FROM audit_events`).Scan(&audits))
	if audits != 0 {
		t.Fatal("request sai đã ghi audit")
	}
	_, err := a.fixture.DB.Conn.Exec(t.Context(), `CREATE FUNCTION fail_http_audit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'synthetic audit failure'; END $$; CREATE TRIGGER fail_http_audit BEFORE INSERT ON audit_events FOR EACH ROW EXECUTE FUNCTION fail_http_audit()`)
	authdb.Require(t, "audit failure trên case test", err)
	_, body := catalogCall(t, a, "POST", "/api/departments", fresh, token, map[string]string{"name": "Must rollback"}, 500)
	assertErrorCode(t, body, "INTERNAL_ERROR")
	if strings.Contains(string(body), "synthetic") {
		t.Fatal("lộ lỗi SQL")
	}
	var rows int
	authdb.Require(t, "rollback qua HTTP", a.fixture.DB.Conn.QueryRow(t.Context(), `SELECT count(*) FROM departments`).Scan(&rows))
	if rows != 0 {
		t.Fatal("audit lỗi để lại catalog")
	}
}
