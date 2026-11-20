package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ai-staffing/internal/service/auth"
	org "ai-staffing/internal/service/organization"
	"github.com/go-chi/chi/v5"
)

type accountStub struct {
	calls int
	err   error
	input org.CreateAccountInput
}

func (s *accountStub) Create(_ context.Context, input org.CreateAccountInput) (org.AccountCreation, error) {
	s.calls++
	s.input = input
	return org.AccountCreation{Account: org.AccountProjection{UserID: 3, EmployeeID: 4, EmployeeCode: "EMP000001", Fullname: "Test", CompanyEmail: "a@example.invalid", Status: "ACTIVE", MustChangePassword: true}, Mail: org.MailAccepted}, s.err
}
func (s *accountStub) List(context.Context, org.ListAccountsQuery) (org.AccountPage, error) {
	s.calls++
	return org.AccountPage{Items: []org.AccountProjection{}, Page: 1, PageSize: 20}, s.err
}
func (s *accountStub) Preview(context.Context, int) (string, error) {
	s.calls++
	return "EMP000001", s.err
}

func accountTestRouter(t *testing.T, session *passwordChangeHTTPStub, service *accountStub) http.Handler {
	t.Helper()
	clock := csrfTestClock{time.Unix(1800000000, 0)}
	h, err := NewAuthRouter(fakeHealth{}, "dev", loginTestOrigin, AuthServices{
		Sessions: session, Login: auth.NewLoginService(&loginHTTPRepo{}, clock, 8*time.Hour), PasswordChange: session, Clock: clock, Accounts: service,
	})
	if err != nil {
		t.Fatal(err)
	}
	return h
}

const validAccountJSON = `{"fullname":"Test","company_email":"a@example.invalid","position_id":1,"system_roles":[]}`

func TestAccountHTTPStrictBodyAndGates(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		status     int
	}{
		{"valid", validAccountJSON, 201},
		{"optional", `{"fullname":"Test","company_email":"a@example.invalid","position_id":1,"phone":null,"address":null,"department_id":null,"job_level_id":null}`, 201},
		{"unknown", strings.TrimSuffix(validAccountJSON, "}") + `,"actor_id":123}`, 400},
		{"duplicate", strings.TrimSuffix(validAccountJSON, "}") + `,"position_id":2}`, 400},
		{"trailing", validAccountJSON + " {}", 400},
		{"null", "null", 400},
		{"array", "[]", 400},
		{"required_null", `{"fullname":null,"company_email":"a@example.invalid","position_id":1}`, 400},
		{"roles_null", `{"fullname":"Test","company_email":"a@example.invalid","position_id":1,"system_roles":null}`, 400},
		{"roles_duplicate", `{"fullname":"Test","company_email":"a@example.invalid","position_id":1,"system_roles":["HR","HR"]}`, 400},
		{"roles_bad", `{"fullname":"Test","company_email":"a@example.invalid","position_id":1,"system_roles":["PM"]}`, 400},
		{"lone_surrogate", `{"fullname":"\ud800","company_email":"a@example.invalid","position_id":1}`, 400},
		{"valid_surrogate_pair", `{"fullname":"\ud83d\ude00","company_email":"a@example.invalid","position_id":1}`, 201},
		{"integer", `{"fullname":"Test","company_email":"a@example.invalid","position_id":"1"}`, 400},
		{"invalid_utf8", strings.Replace(validAccountJSON, "Test", string([]byte{255}), 1), 400},
		{"8192", validAccountJSON + strings.Repeat(" ", 8192-len(validAccountJSON)), 201},
		{"8193", validAccountJSON + strings.Repeat(" ", 8193-len(validAccountJSON)), 413},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &accountStub{}
			actor := catalogHR()
			actor.principal.SystemRoles = []string{"ADMIN", "HR"}
			w := httptest.NewRecorder()
			accountTestRouter(t, actor, s).ServeHTTP(w, passwordChangeRequest("POST", "/api/employees", tc.body))
			assertCatalogHTTP(t, w, tc.status, "")
			if tc.status == 201 {
				if s.calls != 1 || s.input.ActorID != actor.principal.UserID || s.input.SystemRoles == nil {
					t.Fatal("Sai input allowlist")
				}
			} else if s.calls != 0 {
				t.Fatal("Request lỗi vẫn gọi use case")
			}
		})
	}
	for _, mode := range []string{"origin", "csrf", "content_type", "no_cookie", "limited", "not_admin", "query"} {
		t.Run(mode, func(t *testing.T) {
			actor := catalogHR()
			actor.principal.SystemRoles = []string{"ADMIN"}
			s := &accountStub{}
			r := passwordChangeRequest("POST", "/api/employees", validAccountJSON)
			status := 403
			switch mode {
			case "origin":
				r.Header.Set("Origin", "https://invalid.example")
			case "csrf":
				r.Header.Del("X-CSRF-Token")
			case "content_type":
				r.Header.Set("Content-Type", "text/plain")
				status = 415
			case "no_cookie":
				r.Header.Del("Cookie")
				status = 401
			case "limited":
				actor.principal.MustChangePassword = true
			case "not_admin":
				actor.principal.SystemRoles = []string{"HR"}
			case "query":
				r.URL.RawQuery = "unknown=1"
				status = 400
			}
			w := httptest.NewRecorder()
			accountTestRouter(t, actor, s).ServeHTTP(w, r)
			assertCatalogHTTP(t, w, status, "")
			if s.calls != 0 {
				t.Fatal("Cổng bảo vệ vẫn gọi use case")
			}
		})
	}
}

func TestAccountReadHTTPQueriesAndErrors(t *testing.T) {
	for _, path := range []string{"/api/employees/next-code", "/api/users", "/api/users?q=%25_&page=2&page_size=10"} {
		actor := catalogHR()
		actor.principal.SystemRoles = []string{"ADMIN"}
		s := &accountStub{}
		w := httptest.NewRecorder()
		accountTestRouter(t, actor, s).ServeHTTP(w, passwordChangeRequest("GET", path, ""))
		assertCatalogHTTP(t, w, 200, "")
		if s.calls != 1 {
			t.Fatal("Route đọc không tới service")
		}
	}
	for _, path := range []string{"/api/employees/next-code?x=1", "/api/users?page=0", "/api/users?page_size=101", "/api/users?q=a&q=b", "/api/users?role=ADMIN"} {
		actor := catalogHR()
		actor.principal.SystemRoles = []string{"ADMIN"}
		s := &accountStub{}
		w := httptest.NewRecorder()
		accountTestRouter(t, actor, s).ServeHTTP(w, passwordChangeRequest("GET", path, ""))
		assertCatalogHTTP(t, w, 400, "VALIDATION_ERROR")
		if s.calls != 0 {
			t.Fatal("Query sai vẫn gọi service")
		}
	}
	for _, tc := range []struct {
		err    error
		status int
		code   string
	}{
		{org.ErrEmailConflict, 409, "EMAIL_CONFLICT"}, {org.ErrUnavailable, 503, "SERVICE_UNAVAILABLE"},
		{errors.Join(org.ErrCommitUnknown, org.ErrPersistence), 500, "INTERNAL_ERROR"},
		{errors.Join(org.ErrCommitUnknown, org.ErrUnavailable), 503, "SERVICE_UNAVAILABLE"},
	} {
		actor := catalogHR()
		actor.principal.SystemRoles = []string{"ADMIN"}
		s := &accountStub{err: tc.err}
		w := httptest.NewRecorder()
		accountTestRouter(t, actor, s).ServeHTTP(w, passwordChangeRequest("POST", "/api/employees", validAccountJSON))
		assertCatalogHTTP(t, w, tc.status, tc.code)
	}
}

func TestAccountRouteInventoryAndCanonicalSwagger(t *testing.T) {
	actor := catalogHR()
	actor.principal.SystemRoles = []string{"ADMIN"}
	h := accountTestRouter(t, actor, &accountStub{})
	routes := map[string]bool{}
	if err := chi.Walk(h.(chi.Routes), func(method, path string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		routes[method+" "+strings.TrimSuffix(path, "/")] = true
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	for _, route := range []string{"POST /api/employees", "GET /api/employees/next-code", "GET /api/users"} {
		if !routes[route] {
			t.Fatal("Thiếu route", route)
		}
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/docs/catalog.openapi.json", nil))
	var spec struct {
		Paths map[string]map[string]any `json:"paths"`
	}
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &spec) != nil || len(spec.Paths) != 3 {
		t.Fatal("Swagger không khớp route account đã triển khai")
	}
}
