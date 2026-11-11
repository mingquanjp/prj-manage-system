package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ai-staffing/internal/service/auth"

	"github.com/go-chi/chi/v5"
)

const passwordChangeTestBody = `{"current_password":" cũ ","new_password":" Mật khẩu mới có dấu "}`

var passwordChangeTestSecret = auth.SessionSecret{1, 2, 3}

type passwordChangeHTTPStub struct {
	principal          auth.Principal
	authErr, changeErr error
	calls              int
	current, next      string
	secret             auth.SessionSecret
}

func (s *passwordChangeHTTPStub) Authenticate(context.Context, auth.SessionSecret) (auth.Principal, error) {
	return s.principal, s.authErr
}

func (s *passwordChangeHTTPStub) Logout(context.Context, auth.SessionSecret) error { return nil }

func (s *passwordChangeHTTPStub) ChangeInitialPassword(_ context.Context, secret auth.SessionSecret, current, next string) error {
	s.calls++
	s.secret, s.current, s.next = secret, current, next
	return s.changeErr
}

func passwordChangeRouter(t *testing.T, stub *passwordChangeHTTPStub) http.Handler {
	t.Helper()
	protection, err := newAuthProtection("prod", loginTestOrigin)
	if err != nil {
		t.Fatal(err)
	}
	clock := csrfTestClock{time.Unix(1800000000, 0)}
	return newRouter(fakeHealth{}, stub, protection, nil, &passwordChangeRoute{service: stub, limiter: newAuthRateLimiter(clock)})
}

func passwordChangeRequest(method, path, body string) *http.Request {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.ContentLength = -1
	r.Header.Set("Origin", loginTestOrigin)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set(csrfHeaderName, authEncoding.EncodeToString(csrfDigest(passwordChangeTestSecret)))
	r.AddCookie(&http.Cookie{Name: sessionCookieName, Value: authEncoding.EncodeToString(passwordChangeTestSecret[:])})
	return r
}

func TestPasswordChangeHTTP(t *testing.T) {
	padded := passwordChangeTestBody + strings.Repeat(" ", authBodyLimit-len(passwordChangeTestBody))
	for _, tc := range []struct {
		name, body, code   string
		edit               func(*http.Request)
		authErr, changeErr error
		status, calls      int
	}{
		{"thành công 8192", padded, "", nil, nil, nil, 204, 1},
		{"8193", padded + " ", "PAYLOAD_TOO_LARGE", nil, nil, nil, 413, 0},
		{"cookie thiếu", passwordChangeTestBody, "UNAUTHENTICATED", func(r *http.Request) { r.Header.Del("Cookie") }, nil, nil, 401, 0},
		{"phiên hết hạn", passwordChangeTestBody, "UNAUTHENTICATED", nil, auth.ErrUnauthenticated, nil, 401, 0},
		{"Origin sai", passwordChangeTestBody, "FORBIDDEN", func(r *http.Request) { r.Header.Set("Origin", "https://other.example.test") }, nil, nil, 403, 0},
		{"CSRF thiếu", passwordChangeTestBody, "FORBIDDEN", func(r *http.Request) { r.Header.Del(csrfHeaderName) }, nil, nil, 403, 0},
		{"CSRF sai", passwordChangeTestBody, "FORBIDDEN", func(r *http.Request) { r.Header.Set(csrfHeaderName, "invalid") }, nil, nil, 403, 0},
		{"media sai", passwordChangeTestBody, "UNSUPPORTED_MEDIA_TYPE", func(r *http.Request) { r.Header.Set("Content-Type", "text/plain") }, nil, nil, 415, 0},
		{"surrogate đơn ở mật khẩu mới", `{"current_password":"cũ","new_password":"mật khẩu mới hợp lệ\uDE00"}`, "VALIDATION_ERROR", nil, nil, nil, 400, 0},
		{"surrogate đơn ở mật khẩu hiện tại", `{"current_password":"cũ\uD83D","new_password":"mật khẩu mới hợp lệ"}`, "VALIDATION_ERROR", nil, nil, nil, 400, 0},
		{"field lạ", `{"current_password":"a","new_password":"b","user_id":4}`, "VALIDATION_ERROR", nil, nil, nil, 400, 0},
		{"field thiếu", `{"current_password":"a"}`, "VALIDATION_ERROR", nil, nil, nil, 400, 0},
		{"field trùng", `{"current_password":"a","current_password":"b","new_password":"b"}`, "VALIDATION_ERROR", nil, nil, nil, 400, 0},
		{"field null", `{"current_password":"a","new_password":null}`, "VALIDATION_ERROR", nil, nil, nil, 400, 0},
		{"JSON thừa", passwordChangeTestBody + "{}", "VALIDATION_ERROR", nil, nil, nil, 400, 0},
		{"policy sai", passwordChangeTestBody, "VALIDATION_ERROR", nil, nil, auth.ErrNewPasswordPolicy, 400, 1},
		{"sai current", passwordChangeTestBody, "INVALID_CURRENT_PASSWORD", nil, nil, auth.ErrInvalidCurrentPassword, 400, 1},
		{"cờ false", passwordChangeTestBody, "PASSWORD_CHANGE_NOT_REQUIRED", nil, nil, auth.ErrPasswordChangeNotRequired, 403, 1},
		{"nền thay đổi", passwordChangeTestBody, "UNAUTHENTICATED", nil, nil, auth.ErrUnauthenticated, 401, 1},
		{"DB unavailable", passwordChangeTestBody, "SERVICE_UNAVAILABLE", nil, nil, auth.ErrUnavailable, 503, 1},
		{"lỗi lưu trữ", passwordChangeTestBody, "INTERNAL_ERROR", nil, nil, auth.ErrPersistence, 500, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stub := &passwordChangeHTTPStub{
				principal: auth.Principal{UserID: 4, MustChangePassword: tc.changeErr != auth.ErrPasswordChangeNotRequired},
				authErr:   tc.authErr, changeErr: tc.changeErr,
			}
			r := passwordChangeRequest(http.MethodPost, "/api/auth/change-password", tc.body)
			if tc.edit != nil {
				tc.edit(r)
			}
			w := httptest.NewRecorder()
			passwordChangeRouter(t, stub).ServeHTTP(w, r)
			if w.Code != tc.status || w.Header().Get("Cache-Control") != "no-store" || stub.calls != tc.calls {
				t.Fatalf("HTTP = %d, cache = %q, calls = %d", w.Code, w.Header().Get("Cache-Control"), stub.calls)
			}
			if tc.status == 204 {
				cookies := w.Result().Cookies()
				if w.Body.Len() != 0 || len(cookies) != 1 {
					t.Fatal("204 phải có body rỗng và một cookie xóa")
				}
				cookie := cookies[0]
				if cookie.Name != sessionCookieName || cookie.Path != sessionCookiePath || cookie.Domain != "" ||
					cookie.Value != "" || cookie.MaxAge != -1 || !cookie.Expires.Before(time.Now()) ||
					!cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteLaxMode {
					t.Fatal("thuộc tính cookie xóa sai")
				}
				if stub.secret != passwordChangeTestSecret || stub.current != " cũ " || stub.next != " Mật khẩu mới có dấu " {
					t.Fatal("service không nhận đúng secret và chuỗi nguyên vẹn")
				}
				return
			}
			var response authErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil || response.Code != tc.code || response.RequestID == "" || w.Header().Get("Set-Cookie") != "" {
				t.Fatalf("phản hồi lỗi không khớp contract: code %q", response.Code)
			}
		})
	}
}

func TestPasswordChangeLimiterWiring(t *testing.T) {
	stub := &passwordChangeHTTPStub{principal: auth.Principal{UserID: 4, MustChangePassword: true}}
	router := passwordChangeRouter(t, stub)
	request := func(want int) *httptest.ResponseRecorder {
		t.Helper()
		w := httptest.NewRecorder()
		router.ServeHTTP(w, passwordChangeRequest(http.MethodPost, "/api/auth/change-password", passwordChangeTestBody))
		if w.Code != want {
			t.Fatalf("HTTP = %d, muốn %d", w.Code, want)
		}
		return w
	}
	for range 5 {
		stub.changeErr = auth.ErrNewPasswordPolicy
		request(400)
		stub.changeErr = auth.ErrUnavailable
		request(503)
		stub.changeErr = auth.ErrInvalidCurrentPassword
		request(400)
	}
	if w := request(429); w.Header().Get("Retry-After") != "900" || stub.calls != 15 {
		t.Fatal("limiter user phải đếm riêng đúng 5 lần xác minh sai")
	}

	stub.principal.UserID = 5
	stub.changeErr = auth.ErrNewPasswordPolicy
	for range 14 {
		request(400)
	}
	if w := request(429); w.Header().Get("Retry-After") != "300" || stub.calls != 29 {
		t.Fatal("limiter IP phải chặn request thứ 31 trước service")
	}
}

func TestLimitedSessionRoutes(t *testing.T) {
	stub := &passwordChangeHTTPStub{principal: auth.Principal{UserID: 4, SystemRoles: []string{"ADMIN"}, MustChangePassword: true}}
	api := passwordChangeRouter(t, stub)
	for _, tc := range []struct {
		method, path, body string
		status             int
	}{
		{http.MethodGet, "/api/auth/me", "", 200},
		{http.MethodGet, "/api/auth/csrf", "", 200},
		{http.MethodPost, "/api/auth/logout", "", 204},
		{http.MethodPost, "/api/auth/change-password", passwordChangeTestBody, 204},
	} {
		w := httptest.NewRecorder()
		api.ServeHTTP(w, passwordChangeRequest(tc.method, tc.path, tc.body))
		if w.Code != tc.status {
			t.Errorf("%s trả %d, muốn %d", tc.path, w.Code, tc.status)
		}
	}
	router := chi.NewRouter()
	router.With(authenticateSession(stub)).Get("/test-business", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, passwordChangeRequest(http.MethodGet, "/test-business", ""))
	var response authErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil || w.Code != 403 || response.Code != "PASSWORD_CHANGE_REQUIRED" {
		t.Fatal("phiên hạn chế ADMIN phải bị chặn trước handler")
	}
	stub.principal.MustChangePassword = false
	w = httptest.NewRecorder()
	router.ServeHTTP(w, passwordChangeRequest(http.MethodGet, "/test-business", ""))
	if w.Code != 204 {
		t.Fatalf("phiên thường trả %d", w.Code)
	}
}
