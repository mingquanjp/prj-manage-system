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
	sc "ai-staffing/internal/service/skillcatalog"
	"github.com/go-chi/chi/v5"
)

type skillCatalogStub struct {
	calls   int
	command sc.Command
	query   sc.ListQuery
	err     error
}

func (s *skillCatalogStub) List(_ context.Context, q sc.ListQuery) (sc.Page, error) {
	s.calls++
	s.query = q
	return sc.Page{Items: []sc.Item{{ID: 7, Name: "Skill", Category: sc.Category{ID: 3, Name: "Group"}}}, Page: q.Page, PageSize: q.PageSize, Total: 1}, s.err
}
func (s *skillCatalogStub) Create(_ context.Context, c sc.Command) (sc.Item, error) {
	s.calls++
	s.command = c
	return sc.Item{ID: 7, Name: c.Name, Category: sc.Category{ID: c.CategoryID, Name: "Group"}}, s.err
}
func (s *skillCatalogStub) Update(ctx context.Context, c sc.Command) (sc.Item, error) {
	return s.Create(ctx, c)
}
func (s *skillCatalogStub) Delete(ctx context.Context, c sc.Command) error {
	_, e := s.Create(ctx, c)
	return e
}
func skillCatalogHTTPRouter(t *testing.T, actor *passwordChangeHTTPStub, s *skillCatalogStub) http.Handler {
	t.Helper()
	clock := csrfTestClock{time.Unix(1800000000, 0)}
	h, e := NewAuthRouter(fakeHealth{}, "prod", loginTestOrigin, AuthServices{Sessions: actor, Login: auth.NewLoginService(&loginHTTPRepo{}, clock, 8*time.Hour), PasswordChange: actor, Clock: clock, SkillCatalogs: s})
	if e != nil {
		t.Fatal(e)
	}
	return h
}
func skillDirector() *passwordChangeHTTPStub {
	return &passwordChangeHTTPStub{principal: auth.Principal{UserID: 42, EmployeeID: 142, SystemRoles: []string{"DIRECTOR"}}}
}
func TestSkillCatalogHTTPStrictBody(t *testing.T) {
	for _, collection := range []string{"skill-categories", "skills"} {
		for _, method := range []string{"POST", "PATCH", "DELETE"} {
			t.Run(collection+method, func(t *testing.T) {
				name := `"name":"New"`
				before := `"name":" Raw before "`
				if collection == "skills" {
					name += `,"category_id":3`
					before += `,"category_id":2`
				}
				body := `{` + name + `}`
				path := "/api/" + collection
				want := 201
				if method == "PATCH" {
					body = `{` + name + `,"expected_before":{` + before + `}}`
					path += "/7"
					want = 200
				}
				if method == "DELETE" {
					body = `{"expected_before":{` + before + `}}`
					path += "/7"
					want = 204
				}
				for _, tc := range []struct {
					name, raw string
					status    int
					edit      func(*http.Request)
				}{
					{"valid", body, want, nil}, {"8192", body + strings.Repeat(" ", 8192-len(body)), want, nil}, {"8193", body + strings.Repeat(" ", 8193-len(body)), 413, nil},
					{"missing", `{}`, 400, nil}, {"unknown", strings.TrimSuffix(body, "}") + `,"actor_id":9}`, 400, nil}, {"trailing", body + `{}`, 400, nil}, {"array", `[]`, 400, nil},
					{"duplicate", strings.Replace(body, `"name":`, `"name":"Duplicate","name":`, 1), 400, nil}, {"escaped duplicate", strings.Replace(body, `"name":`, `"na\u006de":"Duplicate","name":`, 1), 400, nil},
					{"null", strings.Replace(body, `"New"`, `null`, 1), 400, nil},
					{"invalid UTF8", strings.Replace(body, "New", string([]byte{255}), 1), 400, nil}, {"NUL", strings.Replace(body, "New", `New\u0000`, 1), 400, nil}, {"surrogate", strings.Replace(body, "New", `\ud800`, 1), 400, nil},
					{"Origin", body, 403, func(r *http.Request) { r.Header.Del("Origin") }}, {"CSRF", body, 403, func(r *http.Request) { r.Header.Del(csrfHeaderName) }},
					{"media", body, 415, func(r *http.Request) { r.Header.Set("Content-Type", "text/plain") }}, {"encoding", body, 415, func(r *http.Request) { r.Header.Set("Content-Encoding", "identity") }},
				} {
					raw := tc.raw
					// DELETE chỉ chứa before; các ca Unicode thay đúng chuỗi nền để thực sự kiểm parser.
					if method == "DELETE" {
						switch tc.name {
						case "null":
							raw = strings.Replace(body, `" Raw before "`, `null`, 1)
						case "invalid UTF8":
							raw = strings.Replace(body, "Raw", string([]byte{255}), 1)
						case "NUL":
							raw = strings.Replace(body, "Raw", `\u0000`, 1)
						case "surrogate":
							raw = strings.Replace(body, "Raw", `\ud800`, 1)
						}
					}
					s := &skillCatalogStub{}
					w := httptest.NewRecorder()
					r := passwordChangeRequest(method, path, raw)
					if tc.edit != nil {
						tc.edit(r)
					}
					skillCatalogHTTPRouter(t, skillDirector(), s).ServeHTTP(w, r)
					assertCatalogHTTP(t, w, tc.status, "")
					if tc.status >= 400 && s.calls != 0 {
						t.Fatalf("%s gọi adapter khi đầu vào bị từ chối", tc.name)
					}
					if tc.status == want && method != "POST" && s.command.ExpectedBefore.Name != " Raw before " {
						t.Fatal("before bị trim")
					}
				}
			})
		}
	}
}
func TestSkillCatalogHTTPRolesQueriesAndErrors(t *testing.T) {
	for _, roles := range [][]string{nil, {"HR"}, {"ADMIN"}, {"DIRECTOR"}, {"DIRECTOR", "ADMIN", "HR"}} {
		a := skillDirector()
		a.principal.SystemRoles = roles
		s := &skillCatalogStub{}
		for _, path := range []string{"/api/skill-categories", "/api/skills?category_id=3&q=%25_&page=2&page_size=1"} {
			w := httptest.NewRecorder()
			skillCatalogHTTPRouter(t, a, s).ServeHTTP(w, passwordChangeRequest("GET", path, ""))
			assertCatalogHTTP(t, w, 200, "")
		}
		w := httptest.NewRecorder()
		skillCatalogHTTPRouter(t, a, s).ServeHTTP(w, passwordChangeRequest("POST", "/api/skill-categories", `{"name":"X"}`))
		want := 403
		for _, r := range roles {
			if r == "DIRECTOR" {
				want = 201
			}
		}
		assertCatalogHTTP(t, w, want, "")
	}
	for _, path := range []string{"/api/skills?category_id=0", "/api/skills?category_id=2147483648", "/api/skills?q=a&q=b", "/api/skills?page_size=101", "/api/skill-categories?category_id=3", "/api/skills?actor_id=1", "/api/skills?page=-1"} {
		s := &skillCatalogStub{}
		w := httptest.NewRecorder()
		skillCatalogHTTPRouter(t, skillDirector(), s).ServeHTTP(w, passwordChangeRequest("GET", path, ""))
		assertCatalogHTTP(t, w, 400, "VALIDATION_ERROR")
		if s.calls != 0 {
			t.Fatal("query sai gọi adapter")
		}
	}
	for _, tc := range []struct {
		err    error
		status int
		code   string
	}{{sc.ErrValidation, 400, "VALIDATION_ERROR"}, {sc.ErrUnauthenticated, 401, "UNAUTHENTICATED"}, {sc.ErrPasswordChangeRequired, 403, "PASSWORD_CHANGE_REQUIRED"}, {sc.ErrForbidden, 403, "FORBIDDEN"}, {sc.ErrNotFound, 404, "NOT_FOUND"}, {sc.ErrStale, 409, "STALE_DATA"}, {sc.ErrDuplicateName, 409, "DUPLICATE_NAME"}, {sc.ErrCatalogInUse, 409, "CATALOG_IN_USE"}, {sc.ErrPersistence, 500, "INTERNAL_ERROR"}, {sc.ErrUnavailable, 503, "SERVICE_UNAVAILABLE"}, {errors.Join(sc.ErrCommitUnknown, sc.ErrUnavailable), 503, "COMMIT_OUTCOME_UNKNOWN"}} {
		s := &skillCatalogStub{err: tc.err}
		w := httptest.NewRecorder()
		skillCatalogHTTPRouter(t, skillDirector(), s).ServeHTTP(w, passwordChangeRequest("POST", "/api/skill-categories", `{"name":"X"}`))
		assertCatalogHTTP(t, w, tc.status, tc.code)
	}
	a := skillDirector()
	a.principal.MustChangePassword = true
	s := &skillCatalogStub{}
	w := httptest.NewRecorder()
	skillCatalogHTTPRouter(t, a, s).ServeHTTP(w, passwordChangeRequest("GET", "/api/skills", ""))
	assertCatalogHTTP(t, w, 403, "PASSWORD_CHANGE_REQUIRED")
	if s.calls != 0 {
		t.Fatal("limited gọi nghiệp vụ")
	}
}
func TestSkillCatalogCompleteRouteInventory(t *testing.T) {
	actor := skillDirector()
	clock := csrfTestClock{time.Unix(1800000000, 0)}
	h, e := NewAuthRouter(fakeHealth{}, "prod", loginTestOrigin, AuthServices{Sessions: actor, Login: auth.NewLoginService(&loginHTTPRepo{}, clock, 8*time.Hour), PasswordChange: actor, Clock: clock, Catalogs: &catalogStub{}, Profiles: &profileStub{}, Accounts: &accountStub{}, ProfileChanges: &pcaHTTPStub{}, Notifications: &notificationStub{}, AdministrativeProfiles: &administrativeHTTPStub{}, SkillCatalogs: &skillCatalogStub{}})
	if e != nil {
		t.Fatal(e)
	}
	got := map[string]bool{}
	if e = chi.Walk(h.(chi.Routes), func(method, path string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		got[method+" "+strings.TrimSuffix(path, "/")] = true
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	expected := map[string]bool{"GET /api/healthz": true}
	for _, filename := range []string{"auth.openapi.json", "organization-accounts-profile.openapi.json", "profile-change-approval.openapi.json", "manual-skills.openapi.json"} {
		raw, err := os.ReadFile("../../../docs/api/" + filename)
		if err != nil {
			t.Fatal(err)
		}
		var doc struct {
			Paths map[string]map[string]struct {
				Status string `json:"x-implementation-status"`
			}
		}
		if json.Unmarshal(raw, &doc) != nil {
			t.Fatal("spec sai")
		}
		for path, methods := range doc.Paths {
			for method, operation := range methods {
				if (filename != "manual-skills.openapi.json" || operation.Status == "implemented") && (method == "get" || method == "post" || method == "patch" || method == "delete") {
					expected[strings.ToUpper(method)+" "+path] = true
				}
			}
		}
	}
	if len(expected) != 40 || !reflect.DeepEqual(got, expected) {
		t.Fatalf("inventory sai: got %v expected %v", got, expected)
	}
}
