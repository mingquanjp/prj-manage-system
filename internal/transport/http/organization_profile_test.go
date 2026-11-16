package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"ai-staffing/internal/service/auth"
	org "ai-staffing/internal/service/organization"
	"github.com/go-chi/chi/v5"
)

type profileStub struct {
	query org.SelfProfileQuery
	calls int
	err   error
}

func (s *profileStub) ReadSelf(_ context.Context, q org.SelfProfileQuery) (org.SelfProfile, error) {
	s.calls++
	s.query = q
	return org.SelfProfile{UserID: q.ActorID, EmployeeID: q.EmployeeID, EmployeeCode: "FIX-25", Fullname: "Nhân viên tổng hợp", CompanyEmail: "synthetic@example.invalid", Position: org.CatalogItem{ID: 3, Name: "Vị trí"}, EmploymentStatus: "ACTIVE"}, s.err
}
func profileHTTPRouter(t *testing.T, session *passwordChangeHTTPStub, profile profileUseCase, env string) http.Handler {
	t.Helper()
	clock := csrfTestClock{time.Unix(1800000000, 0)}
	h, err := NewApplicationRouter(fakeHealth{}, env, loginTestOrigin, AuthServices{Sessions: session, Login: auth.NewLoginService(&loginHTTPRepo{}, clock, 8*time.Hour), PasswordChange: session, Clock: clock, Catalogs: &catalogStub{}, Profiles: profile})
	if err != nil {
		t.Fatal(err)
	}
	return h
}
func TestSelfProfileHTTP(t *testing.T) {
	for _, tc := range []struct {
		name, path, method string
		cookie             bool
		limited            bool
		err                error
		status             int
		code               string
		calls              int
	}{
		{name: "self", path: "/api/employees/25", cookie: true, status: 200, calls: 1},
		{name: "missing cookie", path: "/api/employees/25", status: 401, code: "UNAUTHENTICATED"},
		{name: "limited", path: "/api/employees/25", cookie: true, limited: true, status: 403, code: "PASSWORD_CHANGE_REQUIRED"},
		{name: "bad id", path: "/api/employees/-1", cookie: true, status: 400, code: "VALIDATION_ERROR"},
		{name: "overflow", path: "/api/employees/2147483648", cookie: true, status: 400, code: "VALIDATION_ERROR"},
		{name: "query", path: "/api/employees/25?actor_id=7", cookie: true, status: 400, code: "VALIDATION_ERROR"},
		{name: "other", path: "/api/employees/1000", cookie: true, err: org.ErrNotFound, status: 404, code: "NOT_FOUND", calls: 1},
		{name: "actor changed", path: "/api/employees/25", cookie: true, err: org.ErrUnauthenticated, status: 401, code: "UNAUTHENTICATED", calls: 1},
		{name: "flag changed", path: "/api/employees/25", cookie: true, err: org.ErrPasswordChangeRequired, status: 403, code: "PASSWORD_CHANGE_REQUIRED", calls: 1},
		{name: "db down", path: "/api/employees/25", cookie: true, err: org.ErrUnavailable, status: 503, code: "SERVICE_UNAVAILABLE", calls: 1},
		{name: "db failure", path: "/api/employees/25", cookie: true, err: org.ErrPersistence, status: 500, code: "INTERNAL_ERROR", calls: 1},
		{name: "write forbidden", path: "/api/employees/25", method: "POST", cookie: true, status: 405, code: "METHOD_NOT_ALLOWED"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			session := catalogHR()
			session.principal.UserID = 12
			session.principal.EmployeeID = 25
			session.principal.SystemRoles = nil
			session.principal.MustChangePassword = tc.limited
			stub := &profileStub{err: tc.err}
			h := profileHTTPRouter(t, session, stub, "dev")
			method := tc.method
			if method == "" {
				method = "GET"
			}
			req := httptest.NewRequest(method, tc.path, nil)
			if tc.cookie {
				req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: authEncoding.EncodeToString(passwordChangeTestSecret[:])})
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)
			assertCatalogHTTP(t, w, tc.status, tc.code)
			if stub.calls != tc.calls {
				t.Fatal("số lần gọi service sai")
			}
			if stub.calls > 0 && stub.query.ActorID != 12 {
				t.Fatal("actor không lấy từ phiên")
			}
			if tc.status == 200 {
				var fields map[string]any
				if err := json.Unmarshal(w.Body.Bytes(), &fields); err != nil {
					t.Fatal(err)
				}
				if len(fields) != 11 || fields["user_id"] != float64(12) || fields["employee_id"] != float64(25) || fields["address"] != nil || fields["phone"] != nil || fields["department"] != nil || fields["job_level"] != nil {
					t.Fatal("DTO/null sai")
				}
				for _, key := range []string{"password_hash", "system_roles", "username", "audit", "session"} {
					if _, ok := fields[key]; ok {
						t.Fatal("DTO lộ trường ngoài phép")
					}
				}
			}
		})
	}
}
func TestSelfProfileRouteAndSwagger(t *testing.T) {
	h := profileHTTPRouter(t, catalogHR(), &profileStub{}, "dev")
	found := 0
	err := chi.Walk(h.(chi.Routes), func(method, path string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		if path == "/api/employees/{id}" {
			if method != "GET" {
				t.Fatal("mở mutation hồ sơ")
			}
			found++
		}
		return nil
	})
	if err != nil || found != 1 {
		t.Fatal("thiếu route self")
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/docs/catalog.openapi.json", nil))
	assertCatalogHTTP(t, w, 200, "")
	var actual, canonical map[string]any
	raw, e := os.ReadFile("../../../docs/api/organization-accounts-profile.openapi.json")
	if e != nil {
		t.Fatal(e)
	}
	if json.Unmarshal(raw, &canonical) != nil || json.Unmarshal(w.Body.Bytes(), &actual) != nil {
		t.Fatal("JSON spec sai")
	}
	paths := actual["paths"].(map[string]any)
	source := canonical["paths"].(map[string]any)
	if !reflect.DeepEqual(paths["/api/employees/{id}"], source["/api/employees/{id}"]) {
		t.Fatal("operation self khác canonical")
	}
	count := 0
	for _, ops := range paths {
		count += len(ops.(map[string]any))
	}
	if count != 13 || strings.Contains(w.Body.String(), "/api/accounts") || strings.Contains(w.Body.String(), "/api/employees/next-code") {
		t.Fatal("spec phải chỉ có 13 operation đã nối")
	}
	for _, env := range []string{"prod", "test"} {
		w := httptest.NewRecorder()
		profileHTTPRouter(t, catalogHR(), &profileStub{}, env).ServeHTTP(w, httptest.NewRequest("GET", "/docs/catalog.openapi.json", nil))
		if w.Code != 404 {
			t.Fatal("docs ngoài dev")
		}
	}
}

func TestApplicationRequiresProfileDependency(t *testing.T) {
	clock := csrfTestClock{time.Unix(1800000000, 0)}
	session := catalogHR()
	services := AuthServices{Sessions: session, Login: auth.NewLoginService(&loginHTTPRepo{}, clock, 8*time.Hour), PasswordChange: session, Clock: clock, Catalogs: &catalogStub{}}
	if h, err := NewApplicationRouter(fakeHealth{}, "dev", loginTestOrigin, services); err == nil || h != nil {
		t.Fatal("thiếu profile không được mở application")
	}
	if h, err := NewLocalDevelopmentApplicationRouter(fakeHealth{}, services); err == nil || h != nil {
		t.Fatal("thiếu profile không được mở local application")
	}
}
