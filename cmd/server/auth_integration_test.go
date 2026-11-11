//go:build integration

package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"ai-staffing/internal/config"
	"ai-staffing/internal/testutil/authdb"
)

const testOrigin = "http://localhost:5173"
const nextPassword = "Synthetic new password for integration"

type integrationClock struct{ nanos atomic.Int64 }

func (c *integrationClock) Now() time.Time { return time.Unix(0, c.nanos.Load()).UTC() }

type authApp struct {
	fixture *authdb.Fixture
	server  *httptest.Server
	clock   *integrationClock
}

func newAuthApp(t *testing.T) *authApp {
	t.Helper()
	f := authdb.NewFixture(t)
	clock := &integrationClock{}
	clock.nanos.Store(time.Date(2026, 11, 12, 9, 0, 0, 0, time.UTC).UnixNano())
	handler, err := buildHandler(&config.Config{AppEnv: "dev", AppOrigin: testOrigin, SessionTTL: 8 * time.Hour}, f.DB.Pool(t), clock)
	authdb.Require(t, "lắp startup HTTP test", err)
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return &authApp{fixture: f, server: server, clock: clock}
}

// Chỉ giữ response fixture trong bộ nhớ; assertion không in secret hoặc body thô.
func (a *authApp) request(ctx context.Context, method, path string, cookie *http.Cookie, token string, payload any) (*http.Response, []byte, error) {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, nil, err
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, a.server.URL+path, body)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Origin", testOrigin)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("X-CSRF-Token", token)
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	response, err := a.server.Client().Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	return response, data, err
}

func (a *authApp) call(t *testing.T, method, path string, cookie *http.Cookie, token string, payload any, status int) (http.Header, []byte) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	response, data, err := a.request(ctx, method, path, cookie, token, payload)
	authdb.Require(t, "request HTTP tích hợp", err)
	if response.StatusCode != status {
		t.Fatalf("%s %s: HTTP %d, cần %d", method, path, response.StatusCode, status)
	}
	if strings.HasPrefix(path, "/api/auth/") && response.Header.Get("Cache-Control") != "no-store" {
		t.Fatal("response auth thiếu no-store")
	}
	return response.Header, data
}

func (a *authApp) login(t *testing.T, email, password string, previous *http.Cookie) *http.Cookie {
	t.Helper()
	header, _ := a.call(t, "POST", "/api/auth/login", previous, "", map[string]string{"email": email, "password": password}, 200)
	cookies := (&http.Response{Header: header}).Cookies()
	if len(cookies) != 1 || cookies[0].Name != "staffing_session" || !cookies[0].HttpOnly || cookies[0].Path != "/api" || cookies[0].SameSite != http.SameSiteLaxMode || cookies[0].Secure || cookies[0].Domain != "" {
		t.Fatal("cookie dev không đúng contract")
	}
	if cookies[0].Expires.Sub(a.clock.Now()) != 8*time.Hour {
		t.Fatal("thời hạn cookie không khớp config")
	}
	raw, err := base64.RawURLEncoding.DecodeString(cookies[0].Value)
	authdb.Require(t, "đọc secret fixture", err)
	hash := sha256.Sum256(raw)
	var stored bool
	err = a.fixture.DB.Conn.QueryRow(t.Context(), `SELECT EXISTS(SELECT 1 FROM auth_sessions WHERE token_hash=$1 AND revoked_at IS NULL)`, hash[:]).Scan(&stored)
	authdb.Require(t, "kiểm cookie có phiên đã commit", err)
	if !stored {
		t.Fatal("cookie được cấp khi chưa có phiên đã commit")
	}
	return cookies[0]
}

func (a *authApp) csrf(t *testing.T, cookie *http.Cookie) string {
	t.Helper()
	header, data := a.call(t, "GET", "/api/auth/csrf", cookie, "", nil, 200)
	if header.Get("Set-Cookie") != "" {
		t.Fatal("csrf đã gia hạn cookie")
	}
	var body struct {
		Token string `json:"csrf_token"`
	}
	authdb.Require(t, "đọc DTO csrf", json.Unmarshal(data, &body))
	if body.Token == "" {
		t.Fatal("token csrf trống")
	}
	return body.Token
}

func assertErrorCode(t *testing.T, data []byte, code string) {
	t.Helper()
	var body struct {
		Code      string `json:"code"`
		RequestID string `json:"request_id"`
	}
	authdb.Require(t, "đọc DTO lỗi", json.Unmarshal(data, &body))
	if body.Code != code || body.RequestID == "" {
		t.Fatalf("mã lỗi không đúng: cần %s", code)
	}
}

func TestAuthHTTPPostgresLifecycle(t *testing.T) {
	a := newAuthApp(t)
	account := a.fixture.AddAccount(t, authdb.AccountOptions{Email: "normal@example.invalid", Roles: []string{"HR"}})
	a.call(t, "GET", "/api/healthz", nil, "", nil, 200)
	first := a.login(t, " NORMAL@EXAMPLE.INVALID ", authdb.FixturePassword, nil)
	second := a.login(t, "normal@example.invalid", authdb.FixturePassword, nil)
	header, data := a.call(t, "GET", "/api/auth/me", first, "", nil, 200)
	var me struct {
		UserID     int      `json:"user_id"`
		EmployeeID int      `json:"employee_id"`
		Roles      []string `json:"system_roles"`
	}
	authdb.Require(t, "đọc DTO me", json.Unmarshal(data, &me))
	if me.UserID != account.UserID || me.EmployeeID != account.EmployeeID || !slices.Equal(me.Roles, []string{"HR"}) || header.Get("Set-Cookie") != "" {
		t.Fatal("me sai danh tính hoặc đã gia hạn cookie")
	}
	token := a.csrf(t, first)
	_, data = a.call(t, "POST", "/api/auth/logout", first, "wrong-token", nil, 403)
	assertErrorCode(t, data, "FORBIDDEN")
	a.call(t, "GET", "/api/auth/me", first, "", nil, 200)
	header, data = a.call(t, "POST", "/api/auth/logout", first, token, nil, 204)
	if len(data) != 0 || len((&http.Response{Header: header}).Cookies()) != 1 || (&http.Response{Header: header}).Cookies()[0].MaxAge != -1 {
		t.Fatal("logout thiếu clear-cookie hoặc có body")
	}
	a.call(t, "GET", "/api/auth/me", first, "", nil, 401)
	a.call(t, "GET", "/api/auth/me", second, "", nil, 200)
	rotated := a.login(t, "normal@example.invalid", authdb.FixturePassword, second)
	a.call(t, "GET", "/api/auth/me", second, "", nil, 401)
	a.call(t, "POST", "/api/auth/logout", rotated, a.csrf(t, rotated), nil, 204)
}

func TestAuthHTTPPostgresInitialPasswordChange(t *testing.T) {
	a := newAuthApp(t)
	user := a.fixture.AddAccount(t, authdb.AccountOptions{Email: "limited@example.invalid", Roles: []string{"ADMIN"}, MustChangePassword: true})
	a.fixture.AddAccount(t, authdb.AccountOptions{Email: "other@example.invalid"})
	first := a.login(t, "limited@example.invalid", authdb.FixturePassword, nil)
	second := a.login(t, "limited@example.invalid", authdb.FixturePassword, nil)
	other := a.login(t, "other@example.invalid", authdb.FixturePassword, nil)
	_, data := a.call(t, "GET", "/api/auth/me", first, "", nil, 200)
	var me struct {
		Limited bool     `json:"must_change_password"`
		Roles   []string `json:"system_roles"`
	}
	authdb.Require(t, "đọc principal hạn chế", json.Unmarshal(data, &me))
	if !me.Limited || me.Roles == nil || len(me.Roles) != 0 {
		t.Fatal("phiên hạn chế còn quyền hiệu lực")
	}
	token := a.csrf(t, first)
	header, data := a.call(t, "POST", "/api/auth/change-password", first, token, map[string]string{"current_password": authdb.FixturePassword, "new_password": nextPassword}, 204)
	if len(data) != 0 || header.Get("Set-Cookie") == "" {
		t.Fatal("đổi thiếu xóa cookie hoặc có body")
	}
	for _, cookie := range []*http.Cookie{first, second} {
		a.call(t, "GET", "/api/auth/me", cookie, "", nil, 401)
	}
	a.call(t, "GET", "/api/auth/me", other, "", nil, 200)
	_, data = a.call(t, "POST", "/api/auth/login", nil, "", map[string]string{"email": "limited@example.invalid", "password": authdb.FixturePassword}, 401)
	assertErrorCode(t, data, "INVALID_CREDENTIALS")
	fresh := a.login(t, "limited@example.invalid", nextPassword, nil)
	_, data = a.call(t, "GET", "/api/auth/me", fresh, "", nil, 200)
	authdb.Require(t, "đọc principal sau đổi", json.Unmarshal(data, &me))
	if me.Limited || !slices.Equal(me.Roles, []string{"ADMIN"}) {
		t.Fatal("login mới không khôi phục quyền hiện hành")
	}
	var limited bool
	var live int
	err := a.fixture.DB.Conn.QueryRow(t.Context(), `SELECT must_change_password,(SELECT count(*) FROM auth_sessions WHERE user_id=$1 AND revoked_at IS NULL) FROM users WHERE user_id=$1`, user.UserID).Scan(&limited, &live)
	authdb.Require(t, "đối chiếu cờ và phiên sau đổi", err)
	if limited || live != 1 {
		t.Fatal("cờ hoặc số phiên sau đổi sai")
	}
	a.call(t, "POST", "/api/auth/change-password", fresh, a.csrf(t, fresh), map[string]string{"current_password": nextPassword, "new_password": authdb.FixturePassword}, 403)
}

func TestAuthHTTPPostgresLivePrincipal(t *testing.T) {
	a := newAuthApp(t)
	account := a.fixture.AddAccount(t, authdb.AccountOptions{Email: "live@example.invalid"})
	cookie := a.login(t, "live@example.invalid", authdb.FixturePassword, nil)
	_, err := a.fixture.DB.Conn.Exec(t.Context(), `INSERT INTO user_roles(user_id,role_code) VALUES ($1,'DIRECTOR')`, account.UserID)
	authdb.Require(t, "thêm quyền fixture", err)
	_, data := a.call(t, "GET", "/api/auth/me", cookie, "", nil, 200)
	var me struct {
		Roles   []string `json:"system_roles"`
		Limited bool     `json:"must_change_password"`
	}
	authdb.Require(t, "đọc quyền hiện hành", json.Unmarshal(data, &me))
	if !slices.Equal(me.Roles, []string{"DIRECTOR"}) {
		t.Fatal("me dùng quyền cũ")
	}
	_, err = a.fixture.DB.Conn.Exec(t.Context(), `UPDATE users SET must_change_password=true WHERE user_id=$1`, account.UserID)
	authdb.Require(t, "bật cờ fixture", err)
	_, data = a.call(t, "GET", "/api/auth/me", cookie, "", nil, 200)
	authdb.Require(t, "đọc cờ hiện hành", json.Unmarshal(data, &me))
	if !me.Limited || len(me.Roles) != 0 {
		t.Fatal("me dùng cờ/quyền hiệu lực cũ")
	}
	_, err = a.fixture.DB.Conn.Exec(t.Context(), `UPDATE employees SET employment_status='INACTIVE' WHERE user_id=$1`, account.UserID)
	authdb.Require(t, "vô hiệu hồ sơ fixture", err)
	a.call(t, "GET", "/api/auth/me", cookie, "", nil, 401)
	_, err = a.fixture.DB.Conn.Exec(t.Context(), `UPDATE employees SET employment_status='ACTIVE' WHERE user_id=$1`, account.UserID)
	authdb.Require(t, "khôi phục fixture", err)
	a.clock.nanos.Add(int64(8 * time.Hour))
	a.call(t, "GET", "/api/auth/me", cookie, "", nil, 401)
	a.call(t, "GET", "/api/auth/csrf", cookie, "", nil, 401)
}

func TestAuthHTTPPostgresFailureAtomicity(t *testing.T) {
	for _, operation := range []string{"login", "change-password", "logout"} {
		for _, stage := range []string{"write", "commit"} {
			t.Run(operation+"/"+stage, func(t *testing.T) {
				a := newAuthApp(t)
				account := a.fixture.AddAccount(t, authdb.AccountOptions{Email: "rollback@example.invalid", MustChangePassword: true})
				cookie := a.login(t, "rollback@example.invalid", authdb.FixturePassword, nil)
				token := a.csrf(t, cookie)
				_, err := a.fixture.DB.Conn.Exec(t.Context(), `CREATE FUNCTION reject_auth_fixture() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'synthetic rollback marker'; END $$`)
				authdb.Require(t, "tạo hàm gây lỗi test", err)
				event := "UPDATE"
				if operation == "login" {
					event = "INSERT"
				}
				trigger := "CREATE TRIGGER fail_auth BEFORE " + event + " ON auth_sessions FOR EACH ROW EXECUTE FUNCTION reject_auth_fixture()"
				if stage == "commit" {
					trigger = "CREATE CONSTRAINT TRIGGER fail_auth AFTER " + event + " ON auth_sessions DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION reject_auth_fixture()"
				}
				_, err = a.fixture.DB.Conn.Exec(t.Context(), trigger)
				authdb.Require(t, "tạo trigger test", err)
				var payload any
				if operation == "login" {
					payload = map[string]string{"email": "rollback@example.invalid", "password": authdb.FixturePassword}
				}
				if operation == "change-password" {
					payload = map[string]string{"current_password": authdb.FixturePassword, "new_password": nextPassword}
				}
				header, data := a.call(t, "POST", "/api/auth/"+operation, cookie, token, payload, 500)
				assertErrorCode(t, data, "INTERNAL_ERROR")
				if header.Get("Set-Cookie") != "" || bytes.Contains(data, []byte("synthetic rollback marker")) {
					t.Fatal("lỗi ghi vẫn phát cookie hoặc lộ lỗi SQL")
				}
				var hash string
				var limited bool
				var live, total int
				err = a.fixture.DB.Conn.QueryRow(t.Context(), `SELECT password_hash,must_change_password,(SELECT count(*) FROM auth_sessions WHERE user_id=$1 AND revoked_at IS NULL),(SELECT count(*) FROM auth_sessions WHERE user_id=$1) FROM users WHERE user_id=$1`, account.UserID).Scan(&hash, &limited, &live, &total)
				authdb.Require(t, "kiểm rollback sau HTTP lỗi", err)
				if hash != a.fixture.PasswordHash || !limited || live != 1 || total != 1 {
					t.Fatal("transaction để lại write nửa chừng")
				}
			})
		}
	}
}

func TestAuthHTTPPostgresErrors(t *testing.T) {
	a := newAuthApp(t)
	a.fixture.AddAccount(t, authdb.AccountOptions{Email: "duplicate@example.invalid"})
	a.fixture.AddAccount(t, authdb.AccountOptions{Email: " DUPLICATE@example.invalid "})
	header, data := a.call(t, "POST", "/api/auth/login", nil, "", map[string]string{"email": "duplicate@example.invalid", "password": authdb.FixturePassword}, 500)
	assertErrorCode(t, data, "INTERNAL_ERROR")
	if header.Get("Set-Cookie") != "" {
		t.Fatal("email trùng vẫn cấp cookie")
	}
	var count int
	err := a.fixture.DB.Conn.QueryRow(t.Context(), `SELECT count(*) FROM auth_sessions`).Scan(&count)
	authdb.Require(t, "kiểm không cấp phiên trùng email", err)
	if count != 0 {
		t.Fatal("đã cấp phiên cho email trùng")
	}
	a.fixture.AddAccount(t, authdb.AccountOptions{Email: "db-error@example.invalid"})
	cookie := a.login(t, "db-error@example.invalid", authdb.FixturePassword, nil)
	_, err = a.fixture.DB.Conn.Exec(t.Context(), `ALTER TABLE auth_sessions RENAME TO hidden_sessions_fixture`)
	authdb.Require(t, "mô phỏng thiếu bảng trong DB case", err)
	_, data = a.call(t, "GET", "/api/auth/me", cookie, "", nil, 500)
	assertErrorCode(t, data, "INTERNAL_ERROR")
}

func TestAuthHTTPPostgresSharedLimiter(t *testing.T) {
	a := newAuthApp(t)
	a.fixture.AddAccount(t, authdb.AccountOptions{Email: "limiter@example.invalid", MustChangePassword: true})
	for i := 0; i < 6; i++ {
		status := 401
		if i == 5 {
			status = 429
		}
		header, _ := a.call(t, "POST", "/api/auth/login", nil, "", map[string]string{"email": " LIMITER@EXAMPLE.INVALID ", "password": "wrong"}, status)
		if status == 429 && header.Get("Retry-After") == "" {
			t.Fatal("limiter runtime thiếu Retry-After")
		}
	}
	a.clock.nanos.Add(int64(5 * time.Minute))
	cookie := a.login(t, "limiter@example.invalid", authdb.FixturePassword, nil)
	token := a.csrf(t, cookie)
	for i := 0; i < 6; i++ {
		status := 400
		if i == 5 {
			status = 429
		}
		header, data := a.call(t, "POST", "/api/auth/change-password", cookie, token, map[string]string{"current_password": "wrong", "new_password": nextPassword}, status)
		if status == 400 {
			assertErrorCode(t, data, "INVALID_CURRENT_PASSWORD")
		}
		if status == 429 && header.Get("Retry-After") == "" {
			t.Fatal("limiter đổi thiếu Retry-After")
		}
	}
	// Login thành công vẫn dùng bucket độc lập với đổi mật khẩu.
	a.login(t, "limiter@example.invalid", authdb.FixturePassword, nil)
}

func TestAuthHTTPPostgresUnavailable(t *testing.T) {
	a := newAuthApp(t)
	a.fixture.AddAccount(t, authdb.AccountOptions{Email: "unavailable@example.invalid"})
	// SQLSTATE mô phỏng lỗi DB tạm thời trong case, không tắt cluster test.
	_, err := a.fixture.DB.Conn.Exec(t.Context(), `CREATE FUNCTION unavailable_fixture() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'synthetic unavailable marker' USING ERRCODE='57P03'; END $$;
 CREATE TRIGGER fail_insert BEFORE INSERT ON auth_sessions FOR EACH ROW EXECUTE FUNCTION unavailable_fixture()`)
	authdb.Require(t, "tạo lỗi DB tạm thời test", err)
	header, data := a.call(t, "POST", "/api/auth/login", nil, "", map[string]string{"email": "unavailable@example.invalid", "password": authdb.FixturePassword}, 503)
	assertErrorCode(t, data, "SERVICE_UNAVAILABLE")
	if header.Get("Set-Cookie") != "" {
		t.Fatal("DB lỗi vẫn cấp cookie")
	}
	var count int
	err = a.fixture.DB.Conn.QueryRow(t.Context(), `SELECT count(*) FROM auth_sessions`).Scan(&count)
	authdb.Require(t, "kiểm DB sau lỗi tạm thời", err)
	if count != 0 {
		t.Fatal("DB lỗi vẫn có phiên")
	}
}
