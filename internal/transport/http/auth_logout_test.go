package http

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ai-staffing/internal/service/auth"
	"github.com/go-chi/chi/v5"
)

type logoutTestRepo struct {
	sessions map[auth.SessionHash]auth.SessionSnapshot
	err      error
	revokes  int
}

func (r *logoutTestRepo) ReadSession(_ context.Context, hash auth.SessionHash) (auth.SessionSnapshot, bool, error) {
	session, found := r.sessions[hash]
	return session, found, nil
}

func (r *logoutTestRepo) RevokeCurrentSession(_ context.Context, hash auth.SessionHash) error {
	r.revokes++
	if r.err != nil {
		return r.err
	}
	delete(r.sessions, hash)
	return nil
}

func TestLogoutThroughRouter(t *testing.T) {
	const origin = "https://staffing.example"
	protection, err := newAuthProtection("prod", origin)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1000, 0)
	secret := sequenceSecret()
	otherSecret := secret
	otherSecret[0]++
	for _, tt := range []struct {
		name, origin, token, body, code string
		revokeError                     error
		status, revokes                 int
	}{
		{name: "thành công", origin: origin, token: csrfVector, status: 204, revokes: 1},
		{name: "sai origin", origin: "https://other.example", token: csrfVector, status: 403, code: "FORBIDDEN"},
		{name: "thiếu token", origin: origin, status: 403, code: "FORBIDDEN"},
		{name: "sai token", origin: origin, token: "wrong", status: 403, code: "FORBIDDEN"},
		{name: "token của phiên khác", origin: origin, token: authEncoding.EncodeToString(csrfDigest(otherSecret)), status: 403, code: "FORBIDDEN"},
		{name: "body không rỗng không rõ độ dài", origin: origin, token: csrfVector, body: " ", status: 400, code: "VALIDATION_ERROR"},
		{name: "phiên hết hạn trong khi revoke", origin: origin, token: csrfVector, revokeError: auth.ErrUnauthenticated, status: 401, code: "UNAUTHENTICATED", revokes: 1},
		{name: "DB tạm lỗi", origin: origin, token: csrfVector, revokeError: auth.ErrUnavailable, status: 503, code: "SERVICE_UNAVAILABLE", revokes: 1},
		{name: "lỗi lưu trữ", origin: origin, token: csrfVector, revokeError: auth.ErrPersistence, status: 500, code: "INTERNAL_ERROR", revokes: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			snapshot := auth.SessionSnapshot{
				Principal:  auth.Principal{UserID: 12, MustChangePassword: true},
				UserActive: true, EmployeeActive: true, ExpiresAt: now.Add(time.Hour),
			}
			repo := &logoutTestRepo{err: tt.revokeError, sessions: map[auth.SessionHash]auth.SessionSnapshot{
				sha256.Sum256(secret[:]): snapshot, sha256.Sum256(otherSecret[:]): snapshot,
			}}
			router := newRouter(fakeHealth{}, auth.NewService(repo, csrfTestClock{now}), protection, nil, nil)
			request := httptest.NewRequest(http.MethodPost, "/api/auth/logout", strings.NewReader(tt.body))
			request.ContentLength = -1
			request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: authEncoding.EncodeToString(secret[:])})
			request.Header.Set("Origin", tt.origin)
			request.Header.Set(csrfHeaderName, tt.token)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != tt.status || repo.revokes != tt.revokes || response.Header().Get("Cache-Control") != "no-store" {
				t.Fatalf("status = %d, revoke = %d, headers = %v", response.Code, repo.revokes, response.Header())
			}
			if tt.status != http.StatusNoContent {
				var body authErrorResponse
				if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || body.Code != tt.code || response.Header().Get("Set-Cookie") != "" || len(repo.sessions) != 2 {
					t.Fatalf("response lỗi hoặc trạng thái phiên không đúng: %s, lỗi = %v", response.Body, err)
				}
				return
			}
			cookies := response.Result().Cookies()
			if response.Body.Len() != 0 || response.Header().Get("Content-Type") != "" || len(cookies) != 1 {
				t.Fatalf("204 phải rỗng và xóa một cookie, headers = %v", response.Header())
			}
			cookie := cookies[0]
			if cookie.Name != sessionCookieName || cookie.Value != "" || cookie.Path != "/api" ||
				cookie.Domain != "" || !cookie.Secure || !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode ||
				cookie.MaxAge != -1 || cookie.Expires.IsZero() || !cookie.Expires.Before(now) {
				t.Fatalf("thuộc tính xóa cookie sai: %v", cookie)
			}
			replay := httptest.NewRecorder()
			router.ServeHTTP(replay, request)
			if replay.Code != http.StatusUnauthorized || repo.revokes != 1 || replay.Header().Get("Set-Cookie") != "" {
				t.Fatalf("gọi lại logout phải trả 401, status = %d, revoke = %d", replay.Code, repo.revokes)
			}
			for _, check := range []struct {
				secret auth.SessionSecret
				status int
			}{{secret, 401}, {otherSecret, 200}} {
				me := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
				me.AddCookie(&http.Cookie{Name: sessionCookieName, Value: authEncoding.EncodeToString(check.secret[:])})
				recorder := httptest.NewRecorder()
				router.ServeHTTP(recorder, me)
				if recorder.Code != check.status {
					t.Errorf("me status = %d, muốn %d", recorder.Code, check.status)
				}
			}
		})
	}
}

func TestSessionRouteWiring(t *testing.T) {
	router := newRouter(fakeHealth{}, auth.NewService(&logoutTestRepo{}, csrfTestClock{}), authProtection{}, nil, nil)
	want := map[string]int{
		"GET /api/healthz": 200, "GET /api/auth/me": 401, "GET /api/auth/csrf": 401, "POST /api/auth/logout": 401,
	}
	err := chi.Walk(router.(chi.Routes), func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		key := method + " " + route
		status, exists := want[key]
		if !exists {
			t.Errorf("route ngoài danh sách: %s", key)
			return nil
		}
		delete(want, key)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(method, route, nil))
		if response.Code != status {
			t.Errorf("%s trả %d khi thiếu cookie, muốn %d", key, response.Code, status)
		}
		return nil
	})
	if err != nil || len(want) != 0 {
		t.Fatalf("route còn thiếu = %v, lỗi = %v", want, err)
	}
}
