package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ai-staffing/internal/service/auth"
)

func localTestRouter(t *testing.T) http.Handler {
	t.Helper()
	clock := csrfTestClock{time.Unix(1800000000, 0)}
	router, err := NewLocalDevelopmentAuthRouter(fakeHealth{}, AuthServices{
		Sessions: &passwordChangeHTTPStub{}, Login: auth.NewLoginService(&loginHTTPRepo{}, clock, 8*time.Hour),
		PasswordChange: &passwordChangeHTTPStub{}, Clock: clock,
	})
	if err != nil {
		t.Fatal(err)
	}
	return router
}

func TestLocalDocsOriginAndProductionIsolation(t *testing.T) {
	for _, origin := range []string{"http://127.0.0.1:5174", "http://127.0.0.1:8080", "http://localhost:8080", "http://127.0.0.1:8081", "https://evil.example", "null", ""} {
		r := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{}`))
		if origin != "" {
			r.Header.Set("Origin", origin)
		}
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		localTestRouter(t).ServeHTTP(w, r)
		allowed := origin == "http://127.0.0.1:5174" || origin == "http://127.0.0.1:8080"
		if (allowed && w.Code != 400) || (!allowed && w.Code != 403) {
			t.Fatalf("origin %q status %d", origin, w.Code)
		}
		if w.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Fatal("không được mở CORS")
		}
	}
	policy, err := newAuthProtection("prod", "https://staffing.example")
	if err != nil {
		t.Fatal(err)
	}
	for _, origin := range []string{"http://127.0.0.1:5174", "http://127.0.0.1:8080"} {
		r := httptest.NewRequest("POST", "/", nil)
		r.Header.Set("Origin", origin)
		if policy.allowsOrigin(r) {
			t.Fatal("production không được nhận origin local")
		}
	}
	r := httptest.NewRequest("POST", "/api/auth/login", nil)
	r.Header.Add("Origin", "http://127.0.0.1:8080")
	r.Header.Add("Origin", "http://127.0.0.1:5174")
	w := httptest.NewRecorder()
	localTestRouter(t).ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal("không nhận nhiều header Origin")
	}
}

func TestLocalDocsAndFEUseSameLimiter(t *testing.T) {
	router := localTestRouter(t)
	for i := 0; i < 31; i++ {
		origin := "http://127.0.0.1:5174"
		if i%2 == 0 {
			origin = "http://127.0.0.1:8080"
		}
		r := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{}`))
		r.Header.Set("Origin", origin)
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		if i < 30 && w.Code != 400 || i == 30 && w.Code != 429 {
			t.Fatalf("request %d status %d", i+1, w.Code)
		}
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest("GET", "/docs/", nil))
	if w.Code != 200 {
		t.Fatal("Swagger không nằm trong cùng router local")
	}
}
