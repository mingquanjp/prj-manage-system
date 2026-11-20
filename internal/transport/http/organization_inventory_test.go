package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	apispec "ai-staffing/docs/api"
	"ai-staffing/internal/service/auth"
	"github.com/go-chi/chi/v5"
)

// Đối chiếu toàn phạm vi tại một router, tránh từng nhóm test xanh nhưng khi ghép lại thiếu route.
func TestApplicationOAPRouteInventoryMatchesCanonicalSwagger(t *testing.T) {
	clock := csrfTestClock{time.Unix(1800000000, 0)}
	handler, err := NewApplicationRouter(fakeHealth{}, "dev", loginTestOrigin, AuthServices{
		Sessions: catalogHR(), Login: auth.NewLoginService(&loginHTTPRepo{}, clock, 8*time.Hour),
		PasswordChange: catalogHR(), Clock: clock, Catalogs: &catalogStub{},
		Profiles: &profileStub{}, Accounts: &accountStub{},
	})
	if err != nil {
		t.Fatal(err)
	}
	var canonical struct {
		Paths map[string]map[string]any `json:"paths"`
	}
	raw, err := apispec.Files.ReadFile("organization-accounts-profile.openapi.json")
	if err != nil || json.Unmarshal(raw, &canonical) != nil {
		t.Fatal("Không đọc được contract canonical")
	}
	wanted := map[string]bool{}
	for path, item := range canonical.Paths {
		for method := range item {
			if method == "get" || method == "post" || method == "patch" || method == "delete" {
				wanted[strings.ToUpper(method)+" "+path] = true
			}
		}
	}
	actual := map[string]bool{}
	err = chi.Walk(handler.(chi.Routes), func(method, path string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		path = strings.TrimSuffix(path, "/")
		if strings.HasPrefix(path, "/api/") && path != "/api/healthz" && !strings.HasPrefix(path, "/api/auth/") {
			actual[method+" "+path] = true
		}
		return nil
	})
	if err != nil || len(wanted) != 16 || !reflect.DeepEqual(actual, wanted) {
		t.Fatal("Router ghép không khớp 16 method/path OAP canonical")
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/docs/catalog.openapi.json", nil))
	var published struct {
		Paths map[string]map[string]any `json:"paths"`
	}
	if response.Code != http.StatusOK || json.Unmarshal(response.Body.Bytes(), &published) != nil ||
		!reflect.DeepEqual(published.Paths, canonical.Paths) {
		t.Fatal("Swagger dev không công bố đúng mọi operation canonical đã lắp")
	}
}
