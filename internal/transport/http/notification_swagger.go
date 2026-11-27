package http

import (
	apispec "ai-staffing/docs/api"
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

// Nhóm thông báo được bật theo dependency riêng để tài liệu khớp các route đang chạy.
func swaggerWithNotifications(s AuthServices, enabled bool) http.HandlerFunc {
	fallback := swaggerWithServices(s.Catalogs != nil, s.Profiles != nil, s.Accounts != nil, s.ProfileChanges != nil)
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/docs/administrative-profile.openapi.json" {
			administrativeSwaggerHandler(s.AdministrativeProfiles != nil)(w, r)
			return
		}
		if r.URL.Path != "/docs/notifications.openapi.json" {
			fallback(w, r)
			return
		}
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
		data, err := notificationSwaggerSpec()
		if err != nil {
			http.Error(w, "Không thể tải tài liệu API", 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(data))
	}
}
func notificationSwaggerSpec() ([]byte, error) {
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
	selected := map[string]map[string]json.RawMessage{
		"/api/notifications":           {"get": paths["/api/notifications"]["get"]},
		"/api/notifications/{id}/read": {"post": paths["/api/notifications/{id}/read"]["post"]},
	}
	doc["paths"], err = json.Marshal(selected)
	if err != nil {
		return nil, err
	}
	return selectedSwaggerDocument(doc, "TaskAssign — Thông báo hồ sơ", "Hai API thông báo chính chủ đã được lắp vào router này.")
}
