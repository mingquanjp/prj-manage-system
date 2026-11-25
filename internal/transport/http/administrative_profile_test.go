package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"ai-staffing/internal/service/auth"
	org "ai-staffing/internal/service/organization"
)

type administrativeHTTPStub struct {
	calls   int
	err     error
	command org.EditAdministrativeProfileCommand
	query   org.ListAdministrativeEmployeesQuery
}

func administrativeHTTPProfile() org.AdministrativeProfile {
	return org.AdministrativeProfile{UserID: 72, EmployeeID: 172, EmployeeCode: "EMP000172", Fullname: "Hồ sơ tổng hợp", CompanyEmail: "synthetic@example.invalid", Position: org.CatalogItem{ID: 2, Name: "Vị trí"}, EmploymentStatus: "INACTIVE"}
}
func (s *administrativeHTTPStub) List(_ context.Context, q org.ListAdministrativeEmployeesQuery) (org.AdministrativeEmployeePage, error) {
	s.calls++
	s.query = q
	p := administrativeHTTPProfile()
	return org.AdministrativeEmployeePage{Items: []org.EmployeeSummary{{EmployeeID: p.EmployeeID, EmployeeCode: p.EmployeeCode, Fullname: p.Fullname, CompanyEmail: p.CompanyEmail, Position: p.Position, EmploymentStatus: p.EmploymentStatus}}, Page: q.Page, PageSize: q.PageSize, Total: 1}, s.err
}
func (s *administrativeHTTPStub) Read(_ context.Context, q org.AdministrativeProfileQuery) (org.AdministrativeProfile, error) {
	s.calls++
	return administrativeHTTPProfile(), s.err
}
func (s *administrativeHTTPStub) Edit(_ context.Context, c org.EditAdministrativeProfileCommand) (org.AdministrativeProfile, error) {
	s.calls++
	s.command = c
	return administrativeHTTPProfile(), s.err
}
func administrativeHTTPRouter(t *testing.T, actor *passwordChangeHTTPStub, s *administrativeHTTPStub) http.Handler {
	t.Helper()
	clock := csrfTestClock{time.Unix(1800000000, 0)}
	h, err := NewAuthRouter(fakeHealth{}, "dev", loginTestOrigin, AuthServices{Sessions: actor, Login: auth.NewLoginService(&loginHTTPRepo{}, clock, 8*time.Hour), PasswordChange: actor, Clock: clock, AdministrativeProfiles: s})
	if err != nil {
		t.Fatal(err)
	}
	return h
}

const administrativeHTTPBody = `{"expected_before_values":{"address":null},"changes":{"address":"Địa chỉ mới"}}`

func TestAdministrativeHTTPStrictJSONAndBody(t *testing.T) {
	padded := administrativeHTTPBody + strings.Repeat(" ", 65536-len(administrativeHTTPBody))
	for _, tc := range []struct {
		body   string
		status int
	}{
		{administrativeHTTPBody, 200}, {padded, 200}, {padded + " ", 413},
		{"null", 400}, {"[]", 400}, {"{}", 400}, {administrativeHTTPBody + "{}", 400},
		{`{"expected_before_values":{"address":null},"changes":{"address":"a","address":"b"}}`, 400},
		{`{"expected_before_values":{"address":null,"address":null},"changes":{"address":"b"}}`, 400},
		{`{"expected_before_values":{"address":null},"changes":{"address":"b"},"changes":{"address":"c"}}`, 400},
		{`{"expected_before_values":{"address":null},"changes":{"address":"b","actor_id":99}}`, 400},
		{`{"expected_before_values":{"address":null},"changes":{"phone":"b"}}`, 400},
		{`{"expected_before_values":{"address":null},"changes":{"address":"\ud800"}}`, 400},
		{`{"expected_before_values":{"address":null},"changes":{"address":"\u0000"}}`, 400},
		{`{"expected_before_values":{"address":null},"changes":{"address":null}}`, 400},
		{`{"expected_before_values":{"position_id":1},"changes":{"position_id":1.5}}`, 400},
		{`{"expected_before_values":{"position_id":1},"changes":{"position_id":"2"}}`, 400},
	} {
		t.Run(tc.body[:min(len(tc.body), 45)], func(t *testing.T) {
			s := &administrativeHTTPStub{}
			h := administrativeHTTPRouter(t, catalogHR(), s)
			r := passwordChangeRequest("PATCH", "/api/employees/172", tc.body)
			r.ContentLength = -1
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			code := ""
			if tc.status == 400 {
				code = "VALIDATION_ERROR"
			}
			if tc.status == 413 {
				code = "PAYLOAD_TOO_LARGE"
			}
			assertCatalogHTTP(t, w, tc.status, code)
			if (s.calls == 1) != (tc.status == 200) {
				t.Fatal("request lỗi đã gọi use case")
			}
			if tc.status == 200 && (s.command.ActorID != 42 || s.command.EmployeeID != 172 || !s.command.ExpectedBeforeValues.Address.Present || s.command.ExpectedBeforeValues.Address.Value != nil) {
				t.Fatal("actor/target hoặc NULL bị sai")
			}
		})
	}
}
func TestAdministrativeHTTPGatesErrorsAndProjection(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
		code   string
	}{{org.ErrUnauthenticated, 401, "UNAUTHENTICATED"}, {org.ErrPasswordChangeRequired, 403, "PASSWORD_CHANGE_REQUIRED"}, {org.ErrForbidden, 403, "FORBIDDEN"}, {org.ErrNotFound, 404, "NOT_FOUND"}, {org.ErrStale, 409, "STALE_DATA"}, {org.ErrEmailConflict, 409, "EMAIL_CONFLICT"}, {org.ErrCatalogNotFound, 409, "CATALOG_NOT_FOUND"}, {org.ErrUnavailable, 503, "SERVICE_UNAVAILABLE"}, {org.ErrCommitUnknown, 503, "SERVICE_UNAVAILABLE"}, {errors.New("raw-secret-provider-error"), 500, "INTERNAL_ERROR"}} {
		s := &administrativeHTTPStub{err: tc.err}
		h := administrativeHTTPRouter(t, catalogHR(), s)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, passwordChangeRequest("PATCH", "/api/employees/172", administrativeHTTPBody))
		assertCatalogHTTP(t, w, tc.status, tc.code)
		if strings.Contains(w.Body.String(), "raw-secret") {
			t.Fatal("lộ lỗi nguồn")
		}
	}
	for _, roles := range [][]string{{"ADMIN"}, {"DIRECTOR"}, {"EMPLOYEE"}, nil} {
		actor := catalogHR()
		actor.principal.SystemRoles = roles
		s := &administrativeHTTPStub{}
		h := administrativeHTTPRouter(t, actor, s)
		for _, path := range []string{"/api/employees", "/api/employees/172/administrative-profile"} {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, passwordChangeRequest("GET", path, ""))
			assertCatalogHTTP(t, w, 403, "FORBIDDEN")
		}
		if s.calls != 0 {
			t.Fatal("role ngoài HR vượt gate")
		}
	}
	s := &administrativeHTTPStub{}
	h := administrativeHTTPRouter(t, catalogHR(), s)
	for _, path := range []string{"/api/employees?page=1&page=2", "/api/employees?status=ACTIVE", "/api/employees?page_size=101", "/api/employees/172/administrative-profile?x=1", "/api/employees/0/administrative-profile"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, passwordChangeRequest("GET", path, ""))
		assertCatalogHTTP(t, w, 400, "VALIDATION_ERROR")
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, passwordChangeRequest("GET", "/api/employees?q=%25_&page=2&page_size=1", ""))
	assertCatalogHTTP(t, w, 200, "")
	if s.query.Query != "%_" || s.query.Page != 2 {
		t.Fatal("query bị thay wildcard hoặc phân trang sai")
	}
	var page struct{ Items []map[string]any }
	if json.Unmarshal(w.Body.Bytes(), &page) != nil || len(page.Items) != 1 || len(page.Items[0]) != 8 {
		t.Fatal("list sai allowlist")
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, passwordChangeRequest("GET", "/api/employees/172/administrative-profile", ""))
	assertCatalogHTTP(t, w, 200, "")
	var detail map[string]any
	if json.Unmarshal(w.Body.Bytes(), &detail) != nil || len(detail) != 11 || detail["phone"] != nil || detail["address"] != nil || detail["employment_status"] != "INACTIVE" {
		t.Fatal("detail sai projection/NULL/status")
	}
}
func TestAdministrativeHTTPMutationProtection(t *testing.T) {
	for _, kind := range []string{"origin", "csrf", "content-type", "encoding"} {
		s := &administrativeHTTPStub{}
		h := administrativeHTTPRouter(t, catalogHR(), s)
		r := passwordChangeRequest("PATCH", "/api/employees/172", administrativeHTTPBody)
		status, code := 403, "FORBIDDEN"
		switch kind {
		case "origin":
			r.Header.Set("Origin", "https://evil.invalid")
		case "csrf":
			r.Header.Del(csrfHeaderName)
		case "content-type":
			r.Header.Set("Content-Type", "text/plain")
			status, code = 415, "UNSUPPORTED_MEDIA_TYPE"
		case "encoding":
			r.Header.Set("Content-Encoding", "gzip")
			status, code = 415, "UNSUPPORTED_MEDIA_TYPE"
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		assertCatalogHTTP(t, w, status, code)
		if s.calls != 0 {
			t.Fatal("request lỗi đã ghi")
		}
	}
}

func TestAdministrativeSwaggerCanonicalAndDependency(t *testing.T) {
	h := administrativeHTTPRouter(t, catalogHR(), &administrativeHTTPStub{})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/docs/administrative-profile.openapi.json", nil))
	if w.Code != 200 || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("thiếu Swagger F4")
	}
	var published, canonical struct {
		Paths map[string]map[string]json.RawMessage
	}
	raw, err := os.ReadFile("../../../docs/api/profile-change-approval.openapi.json")
	if err != nil || json.Unmarshal(raw, &canonical) != nil || json.Unmarshal(w.Body.Bytes(), &published) != nil {
		t.Fatal("spec không hợp lệ")
	}
	if len(published.Paths) != 3 {
		t.Fatal("Swagger không đúng3routes")
	}
	for path, ops := range published.Paths {
		for method, operation := range ops {
			var got, want any
			_ = json.Unmarshal(operation, &got)
			_ = json.Unmarshal(canonical.Paths[path][method], &want)
			if !reflect.DeepEqual(got, want) {
				t.Fatal("operation khác canonical")
			}
		}
	}
	w = httptest.NewRecorder()
	swaggerWithNotifications(AuthServices{}, false)(w, httptest.NewRequest("GET", "/docs/administrative-profile.openapi.json", nil))
	if w.Code != 404 {
		t.Fatal("publish khi chưa có service")
	}
}
