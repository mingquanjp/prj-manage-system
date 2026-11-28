package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	apispec "ai-staffing/docs/api"
)

func swaggerWithSkillCatalogs(services AuthServices) http.HandlerFunc {
	fallback := swaggerWithNotifications(services, services.Notifications != nil)
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/docs/skill-catalog.openapi.json" {
			fallback(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Security-Policy", swaggerCSP)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "same-origin")
		if services.SkillCatalogs == nil {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "Phương thức không được hỗ trợ", http.StatusMethodNotAllowed)
			return
		}
		data, err := skillCatalogSwaggerSpec()
		if err != nil {
			http.Error(w, "Không thể tải tài liệu API", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(data))
	}
}

func skillCatalogSwaggerSpec() ([]byte, error) {
	raw, err := apispec.Files.ReadFile("manual-skills.openapi.json")
	if err != nil {
		return nil, err
	}
	var doc map[string]json.RawMessage
	if err = json.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	var paths map[string]map[string]json.RawMessage
	if err = json.Unmarshal(doc["paths"], &paths); err != nil {
		return nil, err
	}
	selected := map[string]map[string]json.RawMessage{}
	for _, path := range []string{"/api/skill-categories", "/api/skills"} {
		selected[path] = map[string]json.RawMessage{"get": paths[path]["get"], "post": paths[path]["post"]}
		selected[path+"/{id}"] = map[string]json.RawMessage{"patch": paths[path+"/{id}"]["patch"], "delete": paths[path+"/{id}"]["delete"]}
	}
	doc["paths"], err = json.Marshal(selected)
	if err != nil {
		return nil, err
	}
	return selectedSwaggerDocument(doc, "TaskAssign — Danh mục kỹ năng", "Tám API nhóm kỹ năng và kỹ năng đã được lắp vào router này.")
}
