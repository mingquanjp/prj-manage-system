package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"ai-staffing/internal/service/auth"
)

func docsTestRouter(t *testing.T, env string) http.Handler {
	t.Helper()
	clock := csrfTestClock{time.Unix(1800000000, 0)}
	h, err := NewAuthRouter(fakeHealth{}, env, loginTestOrigin, AuthServices{
		Sessions: &passwordChangeHTTPStub{}, Login: auth.NewLoginService(&loginHTTPRepo{}, clock, 8*time.Hour),
		PasswordChange: &passwordChangeHTTPStub{}, Clock: clock,
	})
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func TestSwaggerDevOnly(t *testing.T) {
	paths := []string{"/docs", "/docs/", "/docs/openapi.json", "/docs/assets/init.js", "/docs/assets/swagger-ui-bundle.js", "/docs/assets/swagger-ui.css", "/docs/assets/docs.css"}
	for _, env := range []string{"prod", "production", "", "DEV", "test"} {
		h := docsTestRouter(t, env)
		for _, path := range paths {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
			if w.Code != 404 || strings.Contains(w.Body.String(), "SwaggerUIBundle") {
				t.Fatalf("env %q mở tài liệu %s", env, path)
			}
		}
	}
	h := docsTestRouter(t, "dev")
	for _, path := range paths {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if path == "/docs" {
			if w.Code != 307 || w.Header().Get("Location") != "/docs/" {
				t.Fatal("redirect tài liệu không đúng")
			}
			continue
		}
		if w.Code != 200 || w.Body.Len() == 0 {
			t.Fatalf("không tải được %s: HTTP %d", path, w.Code)
		}
		if w.Header().Get("Cache-Control") != "no-store" || w.Header().Get("Content-Security-Policy") != swaggerCSP ||
			w.Header().Get("X-Content-Type-Options") != "nosniff" || w.Header().Get("X-Frame-Options") != "DENY" || w.Header().Get("Referrer-Policy") != "same-origin" {
			t.Fatalf("header tài liệu không đúng: %s", path)
		}
		if w.Header().Get("Set-Cookie") != "" || w.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Fatal("trang tài liệu cấp cookie hoặc mở CORS")
		}
		wantType := "text/html"
		switch {
		case strings.HasSuffix(path, ".json"):
			wantType = "application/json"
		case strings.HasSuffix(path, ".js"):
			wantType = "text/javascript"
		case strings.HasSuffix(path, ".css"):
			wantType = "text/css"
		}
		if !strings.HasPrefix(w.Header().Get("Content-Type"), wantType) {
			t.Fatalf("MIME sai: %s", path)
		}
		head := httptest.NewRecorder()
		h.ServeHTTP(head, httptest.NewRequest("HEAD", path, nil))
		if head.Code != 200 || head.Body.Len() != 0 {
			t.Fatal("HEAD trả body hoặc không tải được tài liệu")
		}
	}
}

func TestSwaggerServesCanonicalSpec(t *testing.T) {
	w := httptest.NewRecorder()
	docsTestRouter(t, "dev").ServeHTTP(w, httptest.NewRequest("GET", "/docs/openapi.json", nil))
	source, err := os.ReadFile("../../../docs/api/auth.openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(source, w.Body.Bytes()) {
		t.Fatal("contract phục vụ khác nguồn chuẩn")
	}
}

func TestSwaggerRejectsExtraFilesAndMutations(t *testing.T) {
	h := docsTestRouter(t, "dev")
	for _, path := range []string{"/docs/assets/", "/docs/assets/LICENSE", "/docs/assets/vendor.json", "/docs/unknown", "/docs/assets/../init.js", "/docs/assets/swagger-ui-bundle.js.map"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 404 {
			t.Fatalf("mở file ngoài allowlist %s", path)
		}
	}
	for _, method := range []string{"POST", "PUT", "DELETE", "OPTIONS"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(method, "/docs/openapi.json", nil))
		if w.Code != 405 || w.Header().Get("Allow") != "GET, HEAD" {
			t.Fatal("tài liệu nhận mutation hoặc preflight")
		}
	}
}

func TestSwaggerCatalogCanonicalSubset(t *testing.T) {
	h := catalogHTTPRouter(t, catalogHR(), &catalogStub{}, "dev")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/docs/catalog.openapi.json", nil))
	if w.Code != 200 || w.Header().Get("Cache-Control") != "no-store" || w.Header().Get("Content-Security-Policy") != swaggerCSP {
		t.Fatal("catalog docs/header sai")
	}
	var actual, source map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &actual); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../../docs/api/organization-accounts-profile.openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	paths := actual["paths"].(map[string]any)
	canonical := source["paths"].(map[string]any)
	count := 0
	for path, ops := range paths {
		if !strings.HasPrefix(path, "/api/departments") && !strings.HasPrefix(path, "/api/positions") && !strings.HasPrefix(path, "/api/job-levels") {
			t.Fatal("spec công bố route chưa có")
		}
		for method, operation := range ops.(map[string]any) {
			count++
			if !reflect.DeepEqual(operation, canonical[path].(map[string]any)[method]) {
				t.Fatal("operation runtime khác canonical")
			}
		}
	}
	if count != 12 {
		t.Fatal("phải có đúng 12 operation")
	}
	components := actual["components"].(map[string]any)
	if _, exists := components["schemas"].(map[string]any)["SelfProfile"]; exists {
		t.Fatal("không cần công bố schema hồ sơ chưa triển khai")
	}
	var walk func(any)
	walk = func(value any) {
		switch v := value.(type) {
		case map[string]any:
			if ref, ok := v["$ref"].(string); ok {
				parts := strings.Split(ref, "/")
				if len(parts) != 4 || parts[1] != "components" {
					t.Fatal("ref bất thường")
				}
				if components[parts[2]].(map[string]any)[parts[3]] == nil {
					t.Fatal("ref runtime không giải được")
				}
			}
			for _, child := range v {
				walk(child)
			}
		case []any:
			for _, child := range v {
				walk(child)
			}
		}
	}
	walk(actual)
	for _, env := range []string{"prod", "test"} {
		w := httptest.NewRecorder()
		catalogHTTPRouter(t, catalogHR(), &catalogStub{}, env).ServeHTTP(w, httptest.NewRequest("GET", "/docs/catalog.openapi.json", nil))
		if w.Code != 404 {
			t.Fatal("docs catalog bị mở ngoài dev")
		}
	}
	w = httptest.NewRecorder()
	docsTestRouter(t, "dev").ServeHTTP(w, httptest.NewRequest("GET", "/docs/catalog.openapi.json", nil))
	if w.Code != 404 {
		t.Fatal("router chưa nối catalog lại công bố docs")
	}
}
