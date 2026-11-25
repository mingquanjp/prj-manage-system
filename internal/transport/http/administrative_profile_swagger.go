package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	apispec "ai-staffing/docs/api"
)

// Chỉ công bố ba operation F4 khi dependency đã được lắp.
func administrativeSwaggerHandler(enabled bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Security-Policy", swaggerCSP)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if !enabled {
			http.NotFound(w, r)
			return
		}
		if r.Method != "GET" && r.Method != "HEAD" {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "Phương thức không được hỗ trợ", 405)
			return
		}
		data, err := administrativeSwaggerSpec()
		if err != nil {
			http.Error(w, "Không thể tải tài liệu API", 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(data))
	}
}
func administrativeSwaggerSpec() ([]byte, error) {
	raw, err := apispec.Files.ReadFile("profile-change-approval.openapi.json")
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
	for path, method := range map[string]string{"/api/employees": "get", "/api/employees/{id}/administrative-profile": "get", "/api/employees/{id}": "patch"} {
		selected[path] = map[string]json.RawMessage{method: paths[path][method]}
	}
	doc["paths"], err = json.Marshal(selected)
	if err != nil {
		return nil, err
	}
	return selectedSwaggerDocument(doc, "TaskAssign — Hồ sơ hành chính HR", "Ba API HR đọc và cập nhật trực tiếp hồ sơ đã được lắp vào router này.")
}
