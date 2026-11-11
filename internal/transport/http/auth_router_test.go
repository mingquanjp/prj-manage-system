package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"

	"ai-staffing/internal/service/auth"
	"github.com/go-chi/chi/v5"
)

func TestAuthRouterCompositionContract(t *testing.T) {
	clock := csrfTestClock{time.Unix(1800000000, 0)}
	services := AuthServices{
		Sessions:       &passwordChangeHTTPStub{},
		Login:          auth.NewLoginService(&loginHTTPRepo{}, clock, 8*time.Hour),
		PasswordChange: &passwordChangeHTTPStub{},
		Clock:          clock,
	}
	handler, err := NewAuthRouter(fakeHealth{}, "prod", loginTestOrigin, services)
	if err != nil {
		t.Fatal(err)
	}
	routes := map[string]bool{}
	if err := chi.Walk(handler.(chi.Routes), func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		routes[method+" "+route] = true
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../../docs/api/auth.openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(raw, &spec); err != nil {
		t.Fatal(err)
	}
	expected := map[string]bool{"GET /api/healthz": true}
	for path, operations := range spec.Paths {
		for method := range operations {
			switch method {
			case "get":
				expected["GET "+path] = true
			case "post":
				expected["POST "+path] = true
			}
		}
	}
	if len(expected) != 6 || !reflect.DeepEqual(routes, expected) {
		t.Fatal("method/path runtime không khớp năm API OpenAPI và health")
	}
	for _, missing := range []string{"sessions", "login", "change", "clock", "origin"} {
		t.Run(missing, func(t *testing.T) {
			bad := services
			origin := loginTestOrigin
			switch missing {
			case "sessions":
				bad.Sessions = nil
			case "login":
				bad.Login = nil
			case "change":
				bad.PasswordChange = nil
			case "clock":
				bad.Clock = nil
			case "origin":
				origin = "invalid"
			}
			if h, err := NewAuthRouter(fakeHealth{}, "prod", origin, bad); err == nil || h != nil {
				t.Fatal("thiếu dependency/cấu hình vẫn mở auth")
			}
		})
	}
	for _, path := range []string{"/api/auth/me", "/api/auth/csrf"} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 401 {
			t.Fatal("route bảo vệ bị mở công khai")
		}
	}
}
