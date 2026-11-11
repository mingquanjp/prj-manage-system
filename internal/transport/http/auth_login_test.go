package http

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ai-staffing/internal/service/auth"

	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/crypto/bcrypt"
)

type loginHTTPRepo struct {
	hash                   string
	lookups, writes        int
	lookupErr, finalizeErr error
	snapshot               auth.SessionSnapshot
	command                auth.FinalizeLoginCommand
}

func (r *loginHTTPRepo) FindLoginIdentity(_ context.Context, email string) (auth.LoginIdentity, bool, error) {
	r.lookups++
	if email != "an@example.test" {
		return auth.LoginIdentity{}, false, r.lookupErr
	}
	return auth.LoginIdentity{UserID: 4, EmployeeID: 19, PasswordHash: r.hash, UserActive: true, EmployeeActive: true}, true, r.lookupErr
}

func (r *loginHTTPRepo) FinalizeLogin(_ context.Context, command auth.FinalizeLoginCommand) (auth.SessionSnapshot, error) {
	r.writes++
	r.command = command
	return r.snapshot, r.finalizeErr
}

const (
	loginTestOrigin = "https://app.example.test"
	loginTestBody   = `{"email":" AN@Example.Test ","password":" pass "}`
)

func makeLoginRouter(t *testing.T, repo *loginHTTPRepo) http.Handler {
	t.Helper()
	now := time.Unix(1800000000, 0)
	clock := csrfTestClock{now}
	repo.snapshot.UserActive = true
	repo.snapshot.EmployeeActive = true
	repo.snapshot.ExpiresAt = now.Add(8 * time.Hour)
	repo.snapshot.Principal.UserID = 4
	repo.snapshot.Principal.EmployeeID = 19
	repo.snapshot.Principal.SystemRoles = []string{"HR"}
	protection, err := newAuthProtection("prod", loginTestOrigin)
	if err != nil {
		t.Fatal(err)
	}
	return newRouter(fakeHealth{}, nil, protection, &loginRoute{
		service: auth.NewLoginService(repo, clock, 8*time.Hour), limiter: newAuthRateLimiter(clock), clock: clock,
	})
}

func loginRequest(router http.Handler, origin, mediaType, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	request.ContentLength = -1
	request.Header.Set("Origin", origin)
	request.Header.Set("Content-Type", mediaType)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func TestLoginThroughRouter(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte(" pass "), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	padded := loginTestBody + strings.Repeat(" ", 8192-len(loginTestBody))
	cases := []struct {
		name, origin, media, body string
		status                    int
		code                      string
		lookups, writes           int
		restricted                bool
		lookupErr, finalizeErr    error
	}{
		{"thường ở biên 8192", loginTestOrigin, "application/json", padded, 200, "", 1, 1, false, nil, nil},
		{"hạn chế", loginTestOrigin, "application/json; charset=utf-8", loginTestBody, 200, "", 1, 1, true, nil, nil},
		{"Origin sai", "https://other.example.test", "application/json", loginTestBody, 403, "FORBIDDEN", 0, 0, false, nil, nil},
		{"Content-Type sai", loginTestOrigin, "text/plain", loginTestBody, 415, "UNSUPPORTED_MEDIA_TYPE", 0, 0, false, nil, nil},
		{"8193 không Content-Length", loginTestOrigin, "application/json", padded + " ", 413, "PAYLOAD_TOO_LARGE", 0, 0, false, nil, nil},
		{"JSON thừa", loginTestOrigin, "application/json", loginTestBody + "{}", 400, "VALIDATION_ERROR", 0, 0, false, nil, nil},
		{"surrogate đơn", loginTestOrigin, "application/json", `{"email":"an@example.test","password":"\uD83D"}`, 400, "VALIDATION_ERROR", 0, 0, false, nil, nil},
		{"trường lạ", loginTestOrigin, "application/json", `{"email":"an@example.test","password":" pass ","role":"ADMIN"}`, 400, "VALIDATION_ERROR", 0, 0, false, nil, nil},
		{"trường trùng", loginTestOrigin, "application/json", `{"email":"an@example.test","password":" pass ","email":"an@example.test"}`, 400, "VALIDATION_ERROR", 0, 0, false, nil, nil},
		{"sai mật khẩu", loginTestOrigin, "application/json", strings.Replace(loginTestBody, " pass ", "wrong", 1), 401, "INVALID_CREDENTIALS", 1, 0, false, nil, nil},
		{"nền thay đổi", loginTestOrigin, "application/json", loginTestBody, 401, "INVALID_CREDENTIALS", 1, 1, false, nil, auth.ErrInvalidCredentials},
		{"email trùng", loginTestOrigin, "application/json", loginTestBody, 500, "INTERNAL_ERROR", 1, 0, false, auth.ErrIdentityConflict, nil},
		{"lỗi commit", loginTestOrigin, "application/json", loginTestBody, 503, "SERVICE_UNAVAILABLE", 1, 1, false, nil, auth.ErrUnavailable},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &loginHTTPRepo{hash: string(hash), lookupErr: tc.lookupErr, finalizeErr: tc.finalizeErr}
			repo.snapshot.Principal.MustChangePassword = tc.restricted
			router := makeLoginRouter(t, repo)
			w := loginRequest(router, tc.origin, tc.media, tc.body)
			if w.Code != tc.status || repo.lookups != tc.lookups || repo.writes != tc.writes || w.Header().Get("Cache-Control") != "no-store" {
				t.Fatalf("status=%d lookup=%d finalize=%d headers=%v", w.Code, repo.lookups, repo.writes, w.Header().Get("Cache-Control"))
			}
			cookies := w.Result().Cookies()
			if tc.status != 200 {
				var body authErrorResponse
				if len(cookies) != 0 || json.Unmarshal(w.Body.Bytes(), &body) != nil || body.Code != tc.code || body.RequestID == "" {
					t.Fatal("lỗi không khớp contract hoặc có cookie")
				}
				return
			}
			if len(cookies) != 1 {
				t.Fatalf("cần đúng một cookie, nhận %d", len(cookies))
			}
			cookie := cookies[0]
			if cookie.Name != sessionCookieName || !cookie.HttpOnly || !cookie.Secure || cookie.Path != "/api" ||
				cookie.SameSite != http.SameSiteLaxMode || cookie.Domain != "" || !cookie.Expires.Equal(repo.snapshot.ExpiresAt) {
				t.Fatal("thuộc tính cookie không khớp contract")
			}
			var body authIdentityResponse
			var fields map[string]json.RawMessage
			if json.Unmarshal(w.Body.Bytes(), &body) != nil || json.Unmarshal(w.Body.Bytes(), &fields) != nil || len(fields) != 6 ||
				body.UserID != 4 || body.EmployeeID != 19 || body.MustChangePassword != tc.restricted ||
				body.SystemRoles == nil || (tc.restricted && len(body.SystemRoles) != 0) ||
				(!tc.restricted && (len(body.SystemRoles) != 1 || body.SystemRoles[0] != "HR")) ||
				strings.Contains(w.Body.String(), cookie.Value) {
				t.Fatal("DTO danh tính không khớp contract")
			}
		})
	}
}

func TestLoginRateLimitThroughRouter(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte(" pass "), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	repo := &loginHTTPRepo{hash: string(hash), finalizeErr: auth.ErrInvalidCredentials}
	router := makeLoginRouter(t, repo)
	for range 6 {
		if got := loginRequest(router, loginTestOrigin, "application/json", loginTestBody); got.Code != 401 {
			t.Fatalf("nền thay đổi không tính lỗi xác minh, status=%d", got.Code)
		}
	}
	wrong := strings.Replace(loginTestBody, " pass ", "wrong", 1)
	for range 5 {
		if got := loginRequest(router, loginTestOrigin, "application/json", wrong); got.Code != 401 {
			t.Fatalf("status=%d, muốn 401", got.Code)
		}
	}
	assertLimited := func(w *httptest.ResponseRecorder) {
		t.Helper()
		var body authErrorResponse
		if w.Code != 429 || w.Header().Get("Retry-After") != "300" || w.Header().Get("Set-Cookie") != "" ||
			json.Unmarshal(w.Body.Bytes(), &body) != nil || body.Code != "RATE_LIMITED" {
			t.Fatalf("phản hồi giới hạn không hợp lệ, status=%d", w.Code)
		}
	}
	assertLimited(loginRequest(router, loginTestOrigin, "application/json", wrong))
	if repo.lookups != 11 || repo.writes != 6 {
		t.Fatalf("limiter email phải chặn trước service, lookup=%d finalize=%d", repo.lookups, repo.writes)
	}
	for range 18 {
		if got := loginRequest(router, loginTestOrigin, "application/json", "{}"); got.Code != 400 {
			t.Fatalf("status=%d, muốn 400", got.Code)
		}
	}
	assertLimited(loginRequest(router, loginTestOrigin, "application/json", "{}"))
	if repo.lookups != 11 {
		t.Fatal("limiter IP phải chặn trước service")
	}
}

func TestLoginPreviousCookieThroughRouter(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte(" pass "), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	old := sequenceSecret()
	for _, tc := range []struct {
		name, cookie string
		valid        bool
	}{
		{"cookie đúng", authEncoding.EncodeToString(old[:]), true},
		{"cookie sai", "invalid", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &loginHTTPRepo{hash: string(hash)}
			router := makeLoginRouter(t, repo)
			r := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(loginTestBody))
			r.Header.Set("Origin", loginTestOrigin)
			r.Header.Set("Content-Type", "application/json")
			r.AddCookie(&http.Cookie{Name: sessionCookieName, Value: tc.cookie})
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)
			if w.Code != 200 || len(w.Result().Cookies()) != 1 {
				t.Fatalf("đăng nhập không cấp cookie mới, status=%d", w.Code)
			}
			previous := repo.command.PreviousTokenHash
			if (previous != nil) != tc.valid || (previous != nil && *previous != sha256.Sum256(old[:])) {
				t.Fatal("hash cookie cũ chuyển xuống transaction không hợp lệ")
			}
			var newSecret auth.SessionSecret
			if !decodeAuthValue(newSecret[:], w.Result().Cookies()[0].Value) || sha256.Sum256(newSecret[:]) != repo.command.NewTokenHash {
				t.Fatal("cookie mới không khớp hash đã chuyển xuống transaction")
			}
		})
	}
}

func TestLoginErrorsRedactSensitiveInput(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	repo := &loginHTTPRepo{lookupErr: fmt.Errorf("DRIVER_MARKER: %w", auth.ErrUnavailable)}
	r := httptest.NewRequest(http.MethodPost, "/api/auth/login",
		strings.NewReader(`{"email":"EMAIL_MARKER@example.test","password":"PASSWORD_MARKER"}`))
	r.Header.Set("Origin", loginTestOrigin)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set(middleware.RequestIDHeader, "REQUEST_MARKER")
	r.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "COOKIE_MARKER"})
	w := httptest.NewRecorder()
	makeLoginRouter(t, repo).ServeHTTP(w, r)

	var body authErrorResponse
	if w.Code != 503 || w.Header().Get("Set-Cookie") != "" || json.Unmarshal(w.Body.Bytes(), &body) != nil ||
		body.RequestID == "" || !strings.Contains(logs.String(), body.RequestID) {
		t.Fatal("phản hồi hoặc log thiếu mã đối chiếu nội bộ")
	}
	for _, marker := range []string{"DRIVER_MARKER", "EMAIL_MARKER", "email_marker", "PASSWORD_MARKER", "COOKIE_MARKER", "REQUEST_MARKER"} {
		if strings.Contains(w.Body.String(), marker) || strings.Contains(logs.String(), marker) {
			t.Fatal("phản hồi hoặc log chứa dữ liệu nhạy cảm")
		}
	}
}
