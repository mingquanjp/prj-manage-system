package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
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
