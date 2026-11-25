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
	pca "ai-staffing/internal/service/profilechange"
	"github.com/go-chi/chi/v5"
)

const pcaTestKey = "ABCDEF01-2345-4678-9ABC-0123456789AB"
const pcaSubmitJSON = `{"expected_before_values":{"fullname":" Old "},"proposed_values":{"fullname":"New"}}`

type pcaHTTPStub struct {
	calls  int
	err    error
	submit pca.SubmitInput
	query  pca.ListQuery
	detail pca.DetailQuery
	reject pca.RejectInput
}

func pcaHTTPDetail() pca.Detail {
	before, proposed := " Old ", "New"
	return pca.Detail{Summary: pca.Summary{RequestID: 7, EmployeeID: 142, EmployeeCode: "EMP000142", Fullname: " Old ", Status: "PENDING", ChangedFields: []string{"fullname"}, RequestedAt: time.Unix(1800000000, 0).UTC()}, Before: pca.Values{"fullname": &before}, Proposed: pca.Values{"fullname": &proposed}, Current: pca.Values{"fullname": &before}, RequestedBy: 42}
}
func (s *pcaHTTPStub) Submit(_ context.Context, in pca.SubmitInput) (pca.Detail, error) {
	s.calls++
	s.submit = in
	return pcaHTTPDetail(), s.err
}
func (s *pcaHTTPStub) List(_ context.Context, q pca.ListQuery) (pca.Page, error) {
	s.calls++
	s.query = q
	return pca.Page{Items: []pca.Summary{pcaHTTPDetail().Summary}, Total: 1, Page: q.Page, PageSize: q.PageSize}, s.err
}
func (s *pcaHTTPStub) Read(_ context.Context, q pca.DetailQuery) (pca.Detail, error) {
	s.calls++
	s.detail = q
	return pcaHTTPDetail(), s.err
}
func (s *pcaHTTPStub) Approve(_ context.Context, q pca.DetailQuery) (pca.Detail, error) {
	s.calls++
	s.detail = q
	return pcaHTTPDetail(), s.err
}
func (s *pcaHTTPStub) Reject(_ context.Context, q pca.RejectInput) (pca.Detail, error) {
	s.calls++
	s.reject = q
	return pcaHTTPDetail(), s.err
}

func pcaHTTPRouter(t *testing.T, actor *passwordChangeHTTPStub, s *pcaHTTPStub) http.Handler {
	t.Helper()
	clock := csrfTestClock{time.Unix(1800000000, 0)}
	h, err := NewAuthRouter(fakeHealth{}, "prod", loginTestOrigin, AuthServices{Sessions: actor, Login: auth.NewLoginService(&loginHTTPRepo{}, clock, 8*time.Hour), PasswordChange: actor, Clock: clock, ProfileChanges: s})
	if err != nil {
		t.Fatal(err)
	}
	return h
}
func pcaHTTPRequest(method, path, body string) *http.Request {
	r := passwordChangeRequest(method, path, body)
	r.Header.Set("Idempotency-Key", pcaTestKey)
	return r
}

func TestProfileChangeHTTPStrictSubmitJSON(t *testing.T) {
	padded := pcaSubmitJSON + strings.Repeat(" ", 65536-len(pcaSubmitJSON))
	for _, tc := range []struct {
		name, body string
		status     int
	}{
		{"valid", pcaSubmitJSON, 201},
		{"65536 actual bytes", padded, 201}, {"65537 actual bytes", padded + " ", 413},
		{"empty", "", 400}, {"null", "null", 400}, {"array", "[]", 400}, {"trailing", pcaSubmitJSON + "{}", 400},
		{"unknown top", strings.TrimSuffix(pcaSubmitJSON, "}") + `,"actor_id":1}`, 400},
		{"duplicate top", strings.TrimSuffix(pcaSubmitJSON, "}") + `,"proposed_values":{"fullname":"Other"}}`, 400},
		{"duplicate nested before", `{"expected_before_values":{"fullname":"Old","fullname":"Other"},"proposed_values":{"fullname":"New"}}`, 400},
		{"duplicate nested proposed", `{"expected_before_values":{"fullname":"Old"},"proposed_values":{"fullname":"New","fullname":"Other"}}`, 400},
		{"escaped duplicate", `{"expected_before_values":{"fullname":"Old"},"proposed_values":{"fullname":"New","full\u006eame":"Other"}}`, 400},
		{"unknown nested", `{"expected_before_values":{"fullname":"Old"},"proposed_values":{"fullname":"New","schema_version":1}}`, 400},
		{"nested numeric", strings.Replace(pcaSubmitJSON, `"New"`, `123`, 1), 400},
		{"nested array", strings.Replace(pcaSubmitJSON, `"New"`, `[]`, 1), 400},
		{"object null", `{"expected_before_values":null,"proposed_values":{"fullname":"New"}}`, 400},
		{"lone high surrogate", strings.Replace(pcaSubmitJSON, `"New"`, `"\ud800"`, 1), 400},
		{"lone low surrogate", strings.Replace(pcaSubmitJSON, `"New"`, `"\udc00"`, 1), 400},
		{"invalid UTF8", strings.Replace(pcaSubmitJSON, "New", string([]byte{255}), 1), 400},
		{"NUL", strings.Replace(pcaSubmitJSON, `"New"`, `"New\u0000"`, 1), 400},
		{"valid pair", strings.Replace(pcaSubmitJSON, `"New"`, `"\ud83d\ude00"`, 1), 201},
		{"replacement character", strings.Replace(pcaSubmitJSON, "New", "�", 1), 201},
		{"optional null explanation", strings.TrimSuffix(pcaSubmitJSON, "}") + `,"explanation":null}`, 201},
		{"nullable phone and legacy address", `{"expected_before_values":{"phone":"","address":null},"proposed_values":{"phone":null,"address":"New"}}`, 201},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &pcaHTTPStub{}
			w := httptest.NewRecorder()
			r := pcaHTTPRequest("POST", "/api/employees/142/profile-change-requests", tc.body)
			if r.ContentLength != -1 {
				t.Fatal("Ca biên phải không khai báo Content-Length")
			}
			pcaHTTPRouter(t, catalogHR(), s).ServeHTTP(w, r)
			code := ""
			if tc.status == 400 {
				code = "VALIDATION_ERROR"
			}
			if tc.status == 413 {
				code = "PAYLOAD_TOO_LARGE"
			}
			assertCatalogHTTP(t, w, tc.status, code)
			if tc.status == 201 {
				if s.calls != 1 || s.submit.ActorID != 42 || s.submit.EmployeeID != 142 || s.submit.Key != strings.ToLower(pcaTestKey) {
					t.Fatal("Sai nguồn actor/employee hoặc canonical key")
				}
			} else if s.calls != 0 {
				t.Fatal("Body lỗi đã gọi use case")
			}
		})
	}
}

func TestProfileChangeHTTPMutationProtectionAndHeaders(t *testing.T) {
	for _, operation := range []struct {
		path, body string
		status     int
	}{
		{"/api/employees/142/profile-change-requests", pcaSubmitJSON, 201},
		{"/api/profile-change-requests/7/approve", `{}`, 200},
		{"/api/profile-change-requests/7/reject", `{"reason":"Không phù hợp"}`, 200},
	} {
		for _, tc := range []struct {
			name   string
			edit   func(*http.Request)
			status int
			code   string
		}{
			{"valid", func(*http.Request) {}, operation.status, ""},
			{"origin missing", func(r *http.Request) { r.Header.Del("Origin") }, 403, "FORBIDDEN"},
			{"origin invalid", func(r *http.Request) { r.Header.Set("Origin", "https://other.invalid") }, 403, "FORBIDDEN"},
			{"origin duplicate", func(r *http.Request) { r.Header.Add("Origin", loginTestOrigin) }, 403, "FORBIDDEN"},
			{"csrf missing", func(r *http.Request) { r.Header.Del(csrfHeaderName) }, 403, "FORBIDDEN"},
			{"csrf wrong", func(r *http.Request) { r.Header.Set(csrfHeaderName, "wrong") }, 403, "FORBIDDEN"},
			{"cookie missing", func(r *http.Request) { r.Header.Del("Cookie") }, 401, "UNAUTHENTICATED"},
			{"media missing", func(r *http.Request) { r.Header.Del("Content-Type") }, 415, "UNSUPPORTED_MEDIA_TYPE"},
			{"media wrong", func(r *http.Request) { r.Header.Set("Content-Type", "text/plain") }, 415, "UNSUPPORTED_MEDIA_TYPE"},
			{"media charset", func(r *http.Request) { r.Header.Set("Content-Type", "application/json; charset=iso-8859-1") }, 415, "UNSUPPORTED_MEDIA_TYPE"},
			{"UTF8 charset", func(r *http.Request) { r.Header.Set("Content-Type", "application/json; charset=UTF-8") }, operation.status, ""},
			{"compressed", func(r *http.Request) { r.Header.Set("Content-Encoding", "gzip") }, 415, "UNSUPPORTED_MEDIA_TYPE"},
			{"query", func(r *http.Request) { r.URL.RawQuery = "actor_id=1" }, 400, "VALIDATION_ERROR"},
		} {
			t.Run(operation.path+"/"+tc.name, func(t *testing.T) {
				s := &pcaHTTPStub{}
				r := pcaHTTPRequest("POST", operation.path, operation.body)
				tc.edit(r)
				w := httptest.NewRecorder()
				pcaHTTPRouter(t, catalogHR(), s).ServeHTTP(w, r)
				assertCatalogHTTP(t, w, tc.status, tc.code)
				if tc.status >= 400 && s.calls != 0 {
					t.Fatal("Request lỗi đã gọi mutation")
				}
			})
		}
	}
	for _, tc := range []struct {
		name   string
		values []string
	}{
		{"missing", nil}, {"duplicate", []string{pcaTestKey, pcaTestKey}}, {"comma", []string{pcaTestKey + "," + pcaTestKey}},
		{"version", []string{"abcdef01-2345-1678-9abc-0123456789ab"}}, {"variant", []string{"abcdef01-2345-4678-7abc-0123456789ab"}}, {"padding", []string{" " + pcaTestKey}},
	} {
		t.Run("key/"+tc.name, func(t *testing.T) {
			s := &pcaHTTPStub{}
			r := pcaHTTPRequest("POST", "/api/employees/142/profile-change-requests", pcaSubmitJSON)
			r.Header.Del("Idempotency-Key")
			for _, v := range tc.values {
				r.Header.Add("Idempotency-Key", v)
			}
			w := httptest.NewRecorder()
			pcaHTTPRouter(t, catalogHR(), s).ServeHTTP(w, r)
			assertCatalogHTTP(t, w, 400, "VALIDATION_ERROR")
			if s.calls != 0 {
				t.Fatal("Key sai đã gọi submit")
			}
		})
	}
}

func TestProfileChangeHTTPQueryIDAndDecisionJSON(t *testing.T) {
	for _, query := range []string{"page=0", "page=", "page=1&page=2", "page=2147483648", "page_size=101", "page_size=0", "scope=", "scope=all", "scope=self&scope=review", "status=", "status=pending", "unknown=1", "page=%zz"} {
		s := &pcaHTTPStub{}
		w := httptest.NewRecorder()
		pcaHTTPRouter(t, catalogHR(), s).ServeHTTP(w, pcaHTTPRequest("GET", "/api/profile-change-requests?"+query, ""))
		assertCatalogHTTP(t, w, 400, "VALIDATION_ERROR")
		if s.calls != 0 {
			t.Fatal("Query sai đã gọi service", query)
		}
	}
	for _, id := range []string{"0", "-1", "abc", "2147483648", "1.0", "+1"} {
		for _, path := range []string{"/api/profile-change-requests/" + id, "/api/employees/" + id + "/profile-change-requests"} {
			s := &pcaHTTPStub{}
			w := httptest.NewRecorder()
			method, body := "GET", ""
			if strings.Contains(path, "employees") {
				method, body = "POST", pcaSubmitJSON
			}
			pcaHTTPRouter(t, catalogHR(), s).ServeHTTP(w, pcaHTTPRequest(method, path, body))
			assertCatalogHTTP(t, w, 400, "VALIDATION_ERROR")
			if s.calls != 0 {
				t.Fatal("ID sai đã gọi service")
			}
		}
	}
	for _, tc := range []struct{ action, body string }{
		{"approve", ""}, {"approve", "null"}, {"approve", `{"reason":"No"}`}, {"approve", `{} {}`},
		{"reject", `{}`}, {"reject", `{"reason":null}`}, {"reject", `{"reason":1}`}, {"reject", `{"reason":"A","reason":"B"}`}, {"reject", `{"reason":"A","actor_id":1}`}, {"reject", `{"reason":"\ud800"}`},
	} {
		s := &pcaHTTPStub{}
		w := httptest.NewRecorder()
		pcaHTTPRouter(t, catalogHR(), s).ServeHTTP(w, pcaHTTPRequest("POST", "/api/profile-change-requests/7/"+tc.action, tc.body))
		assertCatalogHTTP(t, w, 400, "VALIDATION_ERROR")
		if s.calls != 0 {
			t.Fatal("Body quyết định sai đã gọi service")
		}
	}
	s := &pcaHTTPStub{}
	w := httptest.NewRecorder()
	pcaHTTPRouter(t, catalogHR(), s).ServeHTTP(w, pcaHTTPRequest("GET", "/api/profile-change-requests?scope=review&status=PENDING&page=2147483647&page_size=100", ""))
	assertCatalogHTTP(t, w, 200, "")
	if s.query != (pca.ListQuery{ActorID: 42, Scope: "review", Status: "PENDING", Page: 2147483647, PageSize: 100}) {
		t.Fatal("Sai query lớn hợp lệ")
	}
	// Giá trị nền phải giữ khoảng trắng; normalization giá trị đề xuất thuộc service.
	s = &pcaHTTPStub{}
	w = httptest.NewRecorder()
	pcaHTTPRouter(t, catalogHR(), s).ServeHTTP(w, pcaHTTPRequest("POST", "/api/employees/142/profile-change-requests", pcaSubmitJSON))
	if s.submit.Before["fullname"] == nil || *s.submit.Before["fullname"] != " Old " {
		t.Fatal("Transport đã đổi giá trị nền")
	}
}

func TestProfileChangeHTTPGatesErrorsAndAllowlist(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
		code   string
	}{
		{pca.ErrValidation, 400, "VALIDATION_ERROR"}, {pca.ErrUnauthenticated, 401, "UNAUTHENTICATED"}, {pca.ErrPasswordChangeRequired, 403, "PASSWORD_CHANGE_REQUIRED"}, {pca.ErrForbidden, 403, "FORBIDDEN"}, {pca.ErrNotFound, 404, "NOT_FOUND"}, {pca.ErrStale, 409, "STALE_DATA"}, {pca.ErrAlreadyResolved, 409, "REQUEST_ALREADY_RESOLVED"}, {pca.ErrEmployeeInactive, 409, "EMPLOYEE_INACTIVE"}, {pca.ErrKeyReused, 409, "IDEMPOTENCY_KEY_REUSED"}, {pca.ErrKeyExpired, 409, "IDEMPOTENCY_KEY_EXPIRED"}, {pca.ErrPersistence, 500, "INTERNAL_ERROR"}, {pca.ErrUnavailable, 503, "SERVICE_UNAVAILABLE"}, {pca.ErrCommitUnknown, 503, "SERVICE_UNAVAILABLE"}, {context.DeadlineExceeded, 503, "SERVICE_UNAVAILABLE"}, {errors.Join(pca.ErrCommitUnknown, pca.ErrPersistence), 503, "SERVICE_UNAVAILABLE"},
	} {
		s := &pcaHTTPStub{err: tc.err}
		w := httptest.NewRecorder()
		pcaHTTPRouter(t, catalogHR(), s).ServeHTTP(w, pcaHTTPRequest("POST", "/api/employees/142/profile-change-requests", pcaSubmitJSON))
		assertCatalogHTTP(t, w, tc.status, tc.code)
		if s.calls != 1 {
			t.Fatal("HTTP không được tự retry mutation")
		}
	}
	for _, path := range []string{"/api/profile-change-requests", "/api/profile-change-requests/7", "/api/employees/142/profile-change-requests", "/api/profile-change-requests/7/approve", "/api/profile-change-requests/7/reject"} {
		for _, limited := range []bool{false, true} {
			actor := catalogHR()
			actor.principal.MustChangePassword = limited
			s := &pcaHTTPStub{}
			method, body := "GET", ""
			if strings.Contains(path, "employees") {
				method, body = "POST", pcaSubmitJSON
			} else if strings.HasSuffix(path, "approve") {
				method, body = "POST", `{}`
			} else if strings.HasSuffix(path, "reject") {
				method, body = "POST", `{"reason":"No"}`
			}
			r := pcaHTTPRequest(method, path, body)
			status, code := 403, "PASSWORD_CHANGE_REQUIRED"
			if !limited {
				r.Header.Del("Cookie")
				status, code = 401, "UNAUTHENTICATED"
			}
			w := httptest.NewRecorder()
			pcaHTTPRouter(t, actor, s).ServeHTTP(w, r)
			assertCatalogHTTP(t, w, status, code)
			if s.calls != 0 {
				t.Fatal("Phiên thiếu/hạn chế đã tới use case")
			}
		}
	}
	for _, roles := range [][]string{{"ADMIN"}, {"DIRECTOR"}, {}} {
		for _, path := range []string{"/api/profile-change-requests/7/approve", "/api/profile-change-requests/7/reject", "/api/profile-change-requests?scope=review"} {
			actor := catalogHR()
			actor.principal.SystemRoles = roles
			s := &pcaHTTPStub{}
			method, body := "POST", `{}`
			if strings.Contains(path, "reject") {
				body = `{"reason":"No"}`
			} else if strings.Contains(path, "?") {
				method, body = "GET", ""
			}
			w := httptest.NewRecorder()
			pcaHTTPRouter(t, actor, s).ServeHTTP(w, pcaHTTPRequest(method, path, body))
			assertCatalogHTTP(t, w, 403, "FORBIDDEN")
			if s.calls != 0 {
				t.Fatal("Role không có HR đã vượt cổng")
			}
		}
	}
	s := &pcaHTTPStub{}
	w := httptest.NewRecorder()
	pcaHTTPRouter(t, catalogHR(), s).ServeHTTP(w, pcaHTTPRequest("GET", "/api/profile-change-requests/7", ""))
	assertCatalogHTTP(t, w, 200, "")
	var detail map[string]json.RawMessage
	if json.Unmarshal(w.Body.Bytes(), &detail) != nil {
		t.Fatal("DTO không phải JSON")
	}
	want := []string{"request_id", "employee_id", "employee_code", "fullname", "status", "before_values", "proposed_values", "current_values", "explanation", "requested_by", "requested_at", "resolved_by", "resolved_at", "decision_reason"}
	if len(detail) != len(want) {
		t.Fatalf("DTO detail có trường ngoài allowlist: %v", reflect.ValueOf(detail).MapKeys())
	}
	for _, key := range want {
		if _, ok := detail[key]; !ok {
			t.Fatal("Thiếu trường DTO", key)
		}
	}
	for _, key := range []string{"explanation", "resolved_by", "resolved_at", "decision_reason"} {
		if string(detail[key]) != "null" {
			t.Fatal("NULL bị đổi nghĩa", key)
		}
	}
	w = httptest.NewRecorder()
	pcaHTTPRouter(t, catalogHR(), &pcaHTTPStub{}).ServeHTTP(w, pcaHTTPRequest("GET", "/api/profile-change-requests", ""))
	assertCatalogHTTP(t, w, 200, "")
	var page struct {
		Items []map[string]json.RawMessage `json:"items"`
	}
	if json.Unmarshal(w.Body.Bytes(), &page) != nil || len(page.Items) != 1 || len(page.Items[0]) != 8 {
		t.Fatal("List summary vượt allowlist")
	}
}

func TestProfileChangeHTTPCompleteInventoryAndSwagger(t *testing.T) {
	actor := catalogHR()
	clock := csrfTestClock{time.Unix(1800000000, 0)}
	services := AuthServices{Sessions: actor, Login: auth.NewLoginService(&loginHTTPRepo{}, clock, 8*time.Hour), PasswordChange: actor, Clock: clock, Catalogs: &catalogStub{}, Profiles: &profileStub{}, Accounts: &accountStub{}, ProfileChanges: &pcaHTTPStub{}, Notifications: &notificationStub{}, AdministrativeProfiles: &administrativeHTTPStub{}}
	h, err := NewAuthRouter(fakeHealth{}, "prod", loginTestOrigin, services)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	if err = chi.Walk(h.(chi.Routes), func(method, path string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		got[method+" "+strings.TrimSuffix(path, "/")] = true
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	expected := map[string]bool{"GET /api/healthz": true}
	for _, filename := range []string{"auth.openapi.json", "organization-accounts-profile.openapi.json", "profile-change-approval.openapi.json"} {
		raw, e := os.ReadFile("../../../docs/api/" + filename)
		if e != nil {
			t.Fatal(e)
		}
		var doc struct{ Paths map[string]map[string]any }
		if json.Unmarshal(raw, &doc) != nil {
			t.Fatal("Spec JSON sai")
		}
		for path, operations := range doc.Paths {
			for method := range operations {
				if method == "get" || method == "post" || method == "patch" || method == "delete" {
					expected[strings.ToUpper(method)+" "+path] = true
				}
			}
		}
	}
	if len(expected) != 32 || !reflect.DeepEqual(got, expected) {
		t.Fatalf("Inventory runtime lệch 5 auth +16 OAP +10 PCA + health: got %v expected %v", got, expected)
	}
	h, err = NewAuthRouter(fakeHealth{}, "dev", loginTestOrigin, services)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/docs/profile-change.openapi.json", nil))
	if w.Code != 200 || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("Swagger PCA không khả dụng/no-store")
	}
	var published, canonical struct {
		Paths map[string]map[string]json.RawMessage
	}
	if json.Unmarshal(w.Body.Bytes(), &published) != nil {
		t.Fatal("Swagger JSON sai")
	}
	raw, err := os.ReadFile("../../../docs/api/profile-change-approval.openapi.json")
	if err != nil || json.Unmarshal(raw, &canonical) != nil {
		t.Fatal("Không đọc canonical")
	}
	count := 0
	for path, operations := range published.Paths {
		for method, operation := range operations {
			count++
			var gotOperation, wantOperation any
			if json.Unmarshal(operation, &gotOperation) != nil || json.Unmarshal(canonical.Paths[path][method], &wantOperation) != nil {
				t.Fatal("Operation không phải JSON")
			}
			if !strings.Contains(path, "profile-change-requests") || !reflect.DeepEqual(gotOperation, wantOperation) {
				t.Fatal("Publish operation chưa triển khai hoặc khác canonical", path, method)
			}
		}
	}
	if count != 5 {
		t.Fatal("Swagger phải chỉ có năm operation PCA", count)
	}
	var filtered map[string]any
	if json.Unmarshal(w.Body.Bytes(), &filtered) != nil {
		t.Fatal("Spec đã lọc không phải object")
	}
	components := filtered["components"].(map[string]any)
	var checkRefs func(any)
	checkRefs = func(value any) {
		switch v := value.(type) {
		case map[string]any:
			if ref, ok := v["$ref"].(string); ok {
				parts := strings.Split(ref, "/")
				if len(parts) != 4 || parts[0] != "#" || parts[1] != "components" {
					t.Fatal("Ref ngoài canonical local", ref)
				}
				group, ok := components[parts[2]].(map[string]any)
				if !ok || group[parts[3]] == nil {
					t.Fatal("Ref không được resolve sau khi lọc", ref)
				}
			}
			for _, child := range v {
				checkRefs(child)
			}
		case []any:
			for _, child := range v {
				checkRefs(child)
			}
		}
	}
	checkRefs(filtered)
	for group, raw := range components {
		for name := range raw.(map[string]any) {
			if strings.Contains(name, "Notification") || strings.Contains(name, "Administrative") || strings.Contains(name, "Employee") || name == "Filter" || name == "Query" || name == "CatalogItem" {
				t.Fatal("Component notification/F4 chưa triển khai đã được publish", group, name)
			}
		}
	}
	services.ProfileChanges = nil
	h, err = NewAuthRouter(fakeHealth{}, "dev", loginTestOrigin, services)
	if err != nil {
		t.Fatal(err)
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/docs/profile-change.openapi.json", nil))
	if w.Code != 404 {
		t.Fatal("Swagger quảng cáo PCA khi thiếu service")
	}
}
