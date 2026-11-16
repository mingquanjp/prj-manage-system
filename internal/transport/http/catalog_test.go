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
	"github.com/go-chi/chi/v5"
)

type catalogStub struct {
	calls   int
	command org.CatalogCommand
	query   org.ListCatalogQuery
	err     error
}

func (s *catalogStub) List(_ context.Context, q org.ListCatalogQuery) (org.CatalogPage, error) {
	s.calls++
	s.query = q
	return org.CatalogPage{Items: []org.CatalogItem{{ID: 7, Name: "Synthetic"}}, Page: q.Page, PageSize: q.PageSize, Total: 1}, s.err
}
func (s *catalogStub) Create(_ context.Context, c org.CatalogCommand) (org.CatalogItem, error) {
	s.calls++
	s.command = c
	return org.CatalogItem{ID: 7, Name: c.Name}, s.err
}
func (s *catalogStub) Rename(ctx context.Context, c org.CatalogCommand) (org.CatalogItem, error) {
	return s.Create(ctx, c)
}
func (s *catalogStub) Delete(ctx context.Context, c org.CatalogCommand) error {
	_, err := s.Create(ctx, c)
	return err
}
func catalogHTTPRouter(t *testing.T, session *passwordChangeHTTPStub, catalogs *catalogStub, env string) http.Handler {
	t.Helper()
	clock := csrfTestClock{time.Unix(1800000000, 0)}
	h, err := NewApplicationRouter(fakeHealth{}, env, loginTestOrigin, AuthServices{Sessions: session, Login: auth.NewLoginService(&loginHTTPRepo{}, clock, 8*time.Hour), PasswordChange: session, Clock: clock, Catalogs: catalogs})
	if err != nil {
		t.Fatal(err)
	}
	return h
}
func catalogHR() *passwordChangeHTTPStub {
	return &passwordChangeHTTPStub{principal: auth.Principal{UserID: 42, EmployeeID: 142, SystemRoles: []string{"HR"}}}
}
func assertCatalogHTTP(t *testing.T, w *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if w.Code != status {
		t.Fatalf("HTTP %d thay vì %d", w.Code, status)
	}
	if w.Header().Get("Cache-Control") != "no-store" || w.Header().Get("Set-Cookie") != "" || w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("header cache/cookie/CORS sai")
	}
	if code != "" {
		var body authErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || body.Code != code || body.RequestID == "" || body.Message == "" {
			t.Fatal("envelope lỗi không đúng contract")
		}
	}
}
func TestCatalogHTTPBodyAndProtection(t *testing.T) {
	for _, method := range []string{"POST", "PATCH", "DELETE"} {
		t.Run(method, func(t *testing.T) {
			body := `{"name":"Valid"}`
			path := "/api/departments"
			success := 201
			if method == "PATCH" {
				body = `{"name":"Valid","expected_before_name":" Old "}`
				path += "/7"
				success = 200
			}
			if method == "DELETE" {
				body = `{"expected_before_name":" Old "}`
				path += "/7"
				success = 204
			}
			padded := body + strings.Repeat(" ", 8192-len(body))
			for _, tc := range []struct {
				name, body string
				edit       func(*http.Request)
				status     int
				code       string
				calls      int
			}{
				{"8192 byte thực", padded, nil, success, "", 1},
				{"8193 byte thực", padded + " ", nil, 413, "PAYLOAD_TOO_LARGE", 0},
				{"JSON rỗng", "", nil, 400, "VALIDATION_ERROR", 0},
				{"field thiếu", `{}`, nil, 400, "VALIDATION_ERROR", 0},
				{"field lạ", strings.TrimSuffix(body, "}") + `,"actor_id":1}`, nil, 400, "VALIDATION_ERROR", 0},
				{"field trùng", strings.TrimSuffix(body, "}") + `,"name":"again"}`, nil, 400, "VALIDATION_ERROR", 0},
				{"trailing", body + `{}`, nil, 400, "VALIDATION_ERROR", 0},
				{"array", `[]`, nil, 400, "VALIDATION_ERROR", 0},
				{"null field", strings.Replace(body, `"Valid"`, `null`, 1), nil, 400, "VALIDATION_ERROR", 0},
				{"Origin thiếu", body, func(r *http.Request) { r.Header.Del("Origin") }, 403, "FORBIDDEN", 0},
				{"Origin sai", body, func(r *http.Request) { r.Header.Set("Origin", "https://other.invalid") }, 403, "FORBIDDEN", 0},
				{"Origin nhiều", body, func(r *http.Request) { r.Header.Add("Origin", loginTestOrigin) }, 403, "FORBIDDEN", 0},
				{"CSRF thiếu", body, func(r *http.Request) { r.Header.Del(csrfHeaderName) }, 403, "FORBIDDEN", 0},
				{"CSRF sai", body, func(r *http.Request) { r.Header.Set(csrfHeaderName, "invalid") }, 403, "FORBIDDEN", 0},
				{"Content-Type thiếu", body, func(r *http.Request) { r.Header.Del("Content-Type") }, 415, "UNSUPPORTED_MEDIA_TYPE", 0},
				{"Content-Type sai", body, func(r *http.Request) { r.Header.Set("Content-Type", "text/plain") }, 415, "UNSUPPORTED_MEDIA_TYPE", 0},
				{"query lạ", body, func(r *http.Request) { r.URL.RawQuery = "actor_id=1" }, 400, "VALIDATION_ERROR", 0},
				{"cookie thiếu", body, func(r *http.Request) { r.Header.Del("Cookie") }, 401, "UNAUTHENTICATED", 0},
			} {
				t.Run(tc.name, func(t *testing.T) {
					// DELETE không có name: chọn trường nền để kiểm null thay vì body không đổi.
					raw := tc.body
					if method == "DELETE" && tc.name == "null field" {
						raw = `{"expected_before_name":null}`
					}
					stub := &catalogStub{}
					h := catalogHTTPRouter(t, catalogHR(), stub, "prod")
					r := passwordChangeRequest(method, path, raw)
					if tc.edit != nil {
						tc.edit(r)
					}
					w := httptest.NewRecorder()
					h.ServeHTTP(w, r)
					assertCatalogHTTP(t, w, tc.status, tc.code)
					if stub.calls != tc.calls {
						t.Fatal("request không hợp lệ đã gọi mutation")
					}
					if stub.calls > 0 && stub.command.ActorID != 42 {
						t.Fatal("actor phải lấy từ principal")
					}
					if tc.status == 204 && (w.Body.Len() != 0 || w.Header().Get("Content-Type") != "") {
						t.Fatal("204 phải rỗng")
					}
					if tc.status == 201 && w.Header().Get("Location") != "/api/departments/7" {
						t.Fatal("Location sai")
					}
					if stub.calls > 0 && method != "POST" && stub.command.ExpectedBeforeName != " Old " {
						t.Fatal("không được trim giá trị nền")
					}
				})
			}
		})
	}
}
func TestCatalogHTTPPermissionAndErrors(t *testing.T) {
	for _, roles := range [][]string{{"HR"}, {"ADMIN"}, {"HR", "ADMIN", "DIRECTOR"}, {"DIRECTOR"}, {}} {
		for _, method := range []string{"GET", "POST", "PATCH", "DELETE"} {
			session := catalogHR()
			session.principal.SystemRoles = roles
			stub := &catalogStub{}
			h := catalogHTTPRouter(t, session, stub, "prod")
			path := "/api/positions"
			body := `{"name":"A"}`
			if method == "PATCH" {
				path += "/7"
				body = `{"name":"A","expected_before_name":"B"}`
			}
			if method == "DELETE" {
				path += "/7"
				body = `{"expected_before_name":"B"}`
			}
			want := 403
			allowed := false
			for _, role := range roles {
				if role == "HR" || method == "GET" && role == "ADMIN" {
					allowed = true
				}
			}
			if allowed {
				want = 200
				if method == "POST" {
					want = 201
				}
				if method == "DELETE" {
					want = 204
				}
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, passwordChangeRequest(method, path, body))
			code := ""
			if want == 403 {
				code = "FORBIDDEN"
			}
			assertCatalogHTTP(t, w, want, code)
			if !allowed && stub.calls != 0 {
				t.Fatal("quyền client vượt handler")
			}
		}
	}
	for _, tc := range []struct {
		err    error
		status int
		code   string
	}{
		{org.ErrValidation, 400, "VALIDATION_ERROR"}, {org.ErrUnauthenticated, 401, "UNAUTHENTICATED"}, {org.ErrPasswordChangeRequired, 403, "PASSWORD_CHANGE_REQUIRED"}, {org.ErrForbidden, 403, "FORBIDDEN"}, {org.ErrNotFound, 404, "NOT_FOUND"}, {org.ErrStale, 409, "STALE_DATA"}, {org.ErrDuplicateName, 409, "DUPLICATE_NAME"}, {org.ErrCatalogInUse, 409, "CATALOG_IN_USE"}, {org.ErrUnavailable, 503, "SERVICE_UNAVAILABLE"}, {org.ErrPersistence, 500, "INTERNAL_ERROR"}, {errors.Join(org.ErrCommitUnknown, org.ErrUnavailable), 503, "SERVICE_UNAVAILABLE"},
	} {
		stub := &catalogStub{err: tc.err}
		w := httptest.NewRecorder()
		catalogHTTPRouter(t, catalogHR(), stub, "prod").ServeHTTP(w, passwordChangeRequest("POST", "/api/job-levels", `{"name":"A"}`))
		assertCatalogHTTP(t, w, tc.status, tc.code)
	}
	for _, tc := range []struct {
		err     error
		limited bool
		status  int
		code    string
	}{{nil, true, 403, "PASSWORD_CHANGE_REQUIRED"}, {auth.ErrUnauthenticated, false, 401, "UNAUTHENTICATED"}, {auth.ErrUnavailable, false, 503, "SERVICE_UNAVAILABLE"}, {auth.ErrPersistence, false, 500, "INTERNAL_ERROR"}} {
		session := catalogHR()
		session.principal.MustChangePassword = tc.limited
		session.authErr = tc.err
		stub := &catalogStub{}
		w := httptest.NewRecorder()
		catalogHTTPRouter(t, session, stub, "prod").ServeHTTP(w, passwordChangeRequest("GET", "/api/job-levels", ""))
		assertCatalogHTTP(t, w, tc.status, tc.code)
		if stub.calls != 0 {
			t.Fatal("phiên sai vẫn gọi catalog")
		}
	}
}
func TestCatalogHTTPQueryAndID(t *testing.T) {
	for _, path := range []string{"/api/departments?page=0", "/api/departments?page=", "/api/departments?page=1&page=2", "/api/departments?page_size=101", "/api/departments?role=HR", "/api/departments?q=a&q=b", "/api/departments?page=2147483648", "/api/departments?q=%zz"} {
		stub := &catalogStub{}
		w := httptest.NewRecorder()
		catalogHTTPRouter(t, catalogHR(), stub, "prod").ServeHTTP(w, passwordChangeRequest("GET", path, ""))
		assertCatalogHTTP(t, w, 400, "VALIDATION_ERROR")
		if stub.calls != 0 {
			t.Fatal("query sai gọi service")
		}
	}
	for _, id := range []string{"0", "-1", "abc", "2147483648", "1.0", "+1"} {
		stub := &catalogStub{}
		w := httptest.NewRecorder()
		catalogHTTPRouter(t, catalogHR(), stub, "prod").ServeHTTP(w, passwordChangeRequest("DELETE", "/api/departments/"+id, `{"expected_before_name":"A"}`))
		assertCatalogHTTP(t, w, 400, "VALIDATION_ERROR")
		if stub.calls != 0 {
			t.Fatal("ID sai gọi service")
		}
	}
	stub := &catalogStub{}
	w := httptest.NewRecorder()
	catalogHTTPRouter(t, catalogHR(), stub, "prod").ServeHTTP(w, passwordChangeRequest("GET", "/api/departments?q=%25_&page=2&page_size=3", ""))
	assertCatalogHTTP(t, w, 200, "")
	if stub.query.Query != "%_" || stub.query.Page != 2 || stub.query.PageSize != 3 {
		t.Fatal("query bị sửa")
	}
}
func TestCatalogRouteInventory(t *testing.T) {
	h := catalogHTTPRouter(t, catalogHR(), &catalogStub{}, "prod")
	got := map[string]bool{}
	err := chi.Walk(h.(chi.Routes), func(method, path string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		got[method+" "+strings.TrimSuffix(path, "/")] = true
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	expected := map[string]bool{"GET /api/healthz": true}
	for _, file := range []string{"auth.openapi.json", "organization-accounts-profile.openapi.json"} {
		raw, err := os.ReadFile("../../../docs/api/" + file)
		if err != nil {
			t.Fatal(err)
		}
		var doc struct{ Paths map[string]map[string]any }
		if err = json.Unmarshal(raw, &doc); err != nil {
			t.Fatal(err)
		}
		for path, ops := range doc.Paths {
			if file != "auth.openapi.json" && !strings.HasPrefix(path, "/api/departments") && !strings.HasPrefix(path, "/api/positions") && !strings.HasPrefix(path, "/api/job-levels") {
				continue
			}
			for method := range ops {
				expected[strings.ToUpper(method)+" "+path] = true
			}
		}
	}
	if len(expected) != 18 || !reflect.DeepEqual(got, expected) {
		t.Fatal("runtime phải có health + 5 auth + 12 catalog đúng spec")
	}
	if h, err := NewApplicationRouter(fakeHealth{}, "prod", loginTestOrigin, AuthServices{}); h != nil || err == nil {
		t.Fatal("dependency thiếu vẫn mở server")
	}
}

func TestCatalogAllRoutesRequireNormalSession(t *testing.T) {
	for _, kind := range catalogCollections {
		for _, method := range []string{"GET", "POST", "PATCH", "DELETE"} {
			for _, limited := range []bool{false, true} {
				stub := &catalogStub{}
				session := catalogHR()
				session.principal.SystemRoles = []string{"HR", "ADMIN"}
				session.principal.MustChangePassword = limited
				path := "/api/" + string(kind)
				if method == "PATCH" || method == "DELETE" {
					path += "/1"
				}
				request := passwordChangeRequest(method, path, `{}`)
				status, code := 403, "PASSWORD_CHANGE_REQUIRED"
				if !limited {
					request.Header.Del("Cookie")
					status, code = 401, "UNAUTHENTICATED"
				}
				w := httptest.NewRecorder()
				catalogHTTPRouter(t, session, stub, "prod").ServeHTTP(w, request)
				assertCatalogHTTP(t, w, status, code)
				if stub.calls != 0 {
					t.Fatal("route cho phép phiên thiếu/hạn chế")
				}
			}
		}
	}
}

func TestCatalogLocalDocsOriginIsolation(t *testing.T) {
	for _, origin := range []string{"http://127.0.0.1:5174", "http://127.0.0.1:8080", "http://localhost:8080", "http://127.0.0.1:8081", "https://evil.invalid", "null", ""} {
		clock := csrfTestClock{time.Unix(1800000000, 0)}
		session := catalogHR()
		stub := &catalogStub{}
		h, err := NewLocalDevelopmentApplicationRouter(fakeHealth{}, AuthServices{Sessions: session, Login: auth.NewLoginService(&loginHTTPRepo{}, clock, 8*time.Hour), PasswordChange: session, Clock: clock, Catalogs: stub})
		if err != nil {
			t.Fatal(err)
		}
		r := passwordChangeRequest("POST", "/api/departments", `{"name":"A"}`)
		r.Header.Del("Origin")
		if origin != "" {
			r.Header.Set("Origin", origin)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		status, code := 403, "FORBIDDEN"
		if origin == "http://127.0.0.1:5174" || origin == "http://127.0.0.1:8080" {
			status, code = 201, ""
		}
		assertCatalogHTTP(t, w, status, code)
		w = httptest.NewRecorder()
		catalogHTTPRouter(t, catalogHR(), &catalogStub{}, "prod").ServeHTTP(w, r)
		assertCatalogHTTP(t, w, 403, "FORBIDDEN")
	}
}
