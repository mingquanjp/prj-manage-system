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
	"ai-staffing/internal/service/skillcatalog"
)

func TestSkillCatalogSwaggerCanonicalSubset(t *testing.T) {
	data, err := skillCatalogSwaggerSpec()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../../docs/api/manual-skills.openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	var actual, canonical map[string]any
	if err = json.Unmarshal(data, &actual); err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, &canonical); err != nil {
		t.Fatal(err)
	}
	paths := actual["paths"].(map[string]any)
	if len(paths) != 4 {
		t.Fatal("spec danh mục phải có bốn path")
	}
	count := 0
	for path, methods := range paths {
		switch path {
		case "/api/skill-categories", "/api/skill-categories/{id}", "/api/skills", "/api/skills/{id}":
		default:
			t.Fatal("công bố API ngoài phạm vi danh mục")
		}
		for method, operation := range methods.(map[string]any) {
			count++
			if !reflect.DeepEqual(operation, canonical["paths"].(map[string]any)[path].(map[string]any)[method]) {
				t.Fatal("operation khác contract canonical")
			}
		}
	}
	if count != 8 {
		t.Fatal("spec phải có đúng tám operation")
	}
	schemas := actual["components"].(map[string]any)["schemas"].(map[string]any)
	for _, name := range []string{"Submit", "Decision", "OfficialPage", "RequestDetail", "RequestPage"} {
		if _, exists := schemas[name]; exists {
			t.Fatal("công bố schema đề nghị chưa triển khai")
		}
	}
}

func TestSkillCatalogSwaggerDependencyAndDevGate(t *testing.T) {
	for _, env := range []string{"dev", "production"} {
		for _, enabled := range []bool{false, true} {
			clock := csrfTestClock{time.Unix(1800000000, 0)}
			services := AuthServices{
				Sessions: &passwordChangeHTTPStub{}, Login: auth.NewLoginService(&loginHTTPRepo{}, clock, 8*time.Hour),
				PasswordChange: &passwordChangeHTTPStub{}, Clock: clock,
			}
			if enabled {
				services.SkillCatalogs = skillcatalog.NewService(nil)
			}
			handler, err := NewAuthRouter(fakeHealth{}, env, loginTestOrigin, services)
			if err != nil {
				t.Fatal(err)
			}
			for _, method := range []string{"GET", "HEAD", "POST"} {
				response := httptest.NewRecorder()
				handler.ServeHTTP(response, httptest.NewRequest(method, "/docs/skill-catalog.openapi.json", nil))
				want := http.StatusNotFound
				if env == "dev" && enabled {
					want = http.StatusOK
					if method == "POST" {
						want = http.StatusMethodNotAllowed
					}
				}
				if response.Code != want {
					t.Fatalf("gate tài liệu sai: env=%s enabled=%t method=%s nhận%d", env, enabled, method, response.Code)
				}
				if want == http.StatusOK && (response.Header().Get("Cache-Control") != "no-store" || response.Header().Get("Content-Security-Policy") != swaggerCSP || response.Header().Get("Set-Cookie") != "") {
					t.Fatal("header tài liệu sai")
				}
				if method == "HEAD" && want == http.StatusOK && response.Body.Len() != 0 {
					t.Fatal("HEAD không được trả body")
				}
			}
		}
	}
}
