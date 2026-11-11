//go:build integration

package http

import (
	stdhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ai-staffing/internal/repository/postgres"
	"ai-staffing/internal/service/auth"
	"ai-staffing/internal/testutil/authdb"
	"github.com/go-chi/chi/v5"
)

func TestAuthRouterPostgresRestrictedGuard(t *testing.T) {
	f := authdb.NewFixture(t)
	account := f.AddAccount(t, authdb.AccountOptions{Email: "guard@example.invalid", Roles: []string{"ADMIN"}, MustChangePassword: true})
	clock := csrfTestClock{time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)}
	repo := postgres.NewAuthRepository(f.DB.Pool(t), clock)
	sessions := auth.NewService(repo, clock)
	handler, err := NewAuthRouter(fakeHealth{}, "prod", loginTestOrigin, AuthServices{Sessions: sessions, Login: auth.NewLoginService(repo, clock, 8*time.Hour), PasswordChange: auth.NewPasswordChangeService(repo, clock), Clock: clock})
	authdb.Require(t, "lắp router guard với PostgreSQL", err)
	// Endpoint thăm dò chỉ tồn tại trong test, qua đúng middleware nghiệp vụ.
	called := false
	handler.(chi.Router).With(authenticateSession(sessions)).Get("/api/test-business", func(w stdhttp.ResponseWriter, _ *stdhttp.Request) { called = true; w.WriteHeader(204) })
	request := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"email":"guard@example.invalid","password":"`+authdb.FixturePassword+`"}`))
	request.Header.Set("Origin", loginTestOrigin)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != 200 || len(response.Result().Cookies()) != 1 {
		t.Fatal("login fixture hạn chế không thành công")
	}
	cookie := response.Result().Cookies()[0]
	if !cookie.Secure || !cookie.HttpOnly || cookie.Domain != "" {
		t.Fatal("cookie prod không đúng thuộc tính")
	}
	for _, limited := range []bool{true, false} {
		_, err = f.DB.Conn.Exec(t.Context(), `UPDATE users SET must_change_password=$1 WHERE user_id=$2`, limited, account.UserID)
		authdb.Require(t, "đổi cờ fixture", err)
		request = httptest.NewRequest("GET", "/api/test-business", nil)
		request.AddCookie(cookie)
		response = httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if limited {
			if response.Code != 403 || called || !strings.Contains(response.Body.String(), "PASSWORD_CHANGE_REQUIRED") {
				t.Fatal("ADMIN hạn chế vượt guard với DB thật")
			}
		} else if response.Code != 204 || !called {
			t.Fatal("guard không đọc cờ hiện hành từ DB")
		}
	}
}
