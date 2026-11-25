package http

import (
	"bytes"
	"embed"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	apispec "ai-staffing/docs/api"
)

// Chỉ nhúng các file phục vụ UI; license và manifest vẫn giữ trong nguồn.
//
//go:embed swagger/index.html swagger/init.js swagger/docs.css swagger/swagger-ui-bundle.js swagger/swagger-ui.css
var swaggerFiles embed.FS

// Chỉ áp dụng cho trang tài liệu, không thay header hoặc bảo vệ của API.
const swaggerCSP = "default-src 'none'; script-src 'self'; style-src 'self'; style-src-attr 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'"

func swaggerHandler(w http.ResponseWriter, r *http.Request) { swaggerWithCatalogs(false)(w, r) }

func swaggerWithCatalogs(catalogs bool) http.HandlerFunc {
	return swaggerWithServices(catalogs, false)
}

func swaggerWithServices(catalogs, profiles bool, accounts ...bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Security-Policy", swaggerCSP)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "same-origin")
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "Phương thức không được hỗ trợ", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path == "/docs" {
			http.Redirect(w, r, "/docs/", http.StatusTemporaryRedirect)
			return
		}
		var data []byte
		var contentType string
		var err error
		switch r.URL.Path {
		case "/docs/":
			data, err = swaggerFiles.ReadFile("swagger/index.html")
			contentType = "text/html; charset=utf-8"
		case "/docs/openapi.json":
			data, err = apispec.Files.ReadFile("auth.openapi.json")
			contentType = "application/json"
		case "/docs/catalog.openapi.json":
			if !catalogs && !profiles && (len(accounts) == 0 || !accounts[0]) {
				http.NotFound(w, r)
				return
			}
			data, err = organizationSwaggerSpec(catalogs, profiles, accounts...)
			contentType = "application/json"
		case "/docs/profile-change.openapi.json":
			if len(accounts) < 2 || !accounts[1] {
				http.NotFound(w, r)
				return
			}
			data, err = profileChangeSwaggerSpec()
			contentType = "application/json"
		case "/docs/assets/init.js", "/docs/assets/swagger-ui-bundle.js":
			data, err = swaggerFiles.ReadFile("swagger/" + r.URL.Path[len("/docs/assets/"):])
			contentType = "text/javascript; charset=utf-8"
		case "/docs/assets/docs.css", "/docs/assets/swagger-ui.css":
			data, err = swaggerFiles.ReadFile("swagger/" + r.URL.Path[len("/docs/assets/"):])
			contentType = "text/css; charset=utf-8"
		default:
			http.NotFound(w, r)
			return
		}
		if err != nil {
			http.Error(w, "Không thể tải tài liệu API", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", contentType)
		http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(data))
	}

}

// Chọn paths từ canonical spec; không tạo bản sao contract viết tay.
func organizationSwaggerSpec(catalogs, profiles bool, accounts ...bool) ([]byte, error) {
	raw, err := apispec.Files.ReadFile("organization-accounts-profile.openapi.json")
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
	if catalogs {
		for _, kind := range catalogCollections {
			path := "/api/" + string(kind)
			selected[path] = map[string]json.RawMessage{"get": paths[path]["get"], "post": paths[path]["post"]}
			selected[path+"/{id}"] = map[string]json.RawMessage{"patch": paths[path+"/{id}"]["patch"], "delete": paths[path+"/{id}"]["delete"]}
		}
	}
	if profiles {
		selected["/api/employees/{id}"] = map[string]json.RawMessage{"get": paths["/api/employees/{id}"]["get"]}
	}
	if len(accounts) > 0 && accounts[0] {
		selected["/api/employees"] = map[string]json.RawMessage{"post": paths["/api/employees"]["post"]}
		selected["/api/employees/next-code"] = map[string]json.RawMessage{"get": paths["/api/employees/next-code"]["get"]}
		selected["/api/users"] = map[string]json.RawMessage{"get": paths["/api/users"]["get"]}
	}
	doc["paths"], err = json.Marshal(selected)
	if err != nil {
		return nil, err
	}
	return selectedSwaggerDocument(doc, "TaskAssign — Tổ chức và hồ sơ", "Các operation tổ chức, tài khoản và hồ sơ đã được lắp vào router này.")
}

func profileChangeSwaggerSpec() ([]byte, error) {
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
	for path, method := range map[string]string{
		"/api/employees/{id}/profile-change-requests": "post",
		"/api/profile-change-requests":                "get",
		"/api/profile-change-requests/{id}":           "get",
		"/api/profile-change-requests/{id}/approve":   "post",
		"/api/profile-change-requests/{id}/reject":    "post",
	} {
		selected[path] = map[string]json.RawMessage{method: paths[path][method]}
	}
	doc["paths"], err = json.Marshal(selected)
	if err != nil {
		return nil, err
	}
	return selectedSwaggerDocument(doc, "TaskAssign — Đề nghị hồ sơ", "Năm API đề nghị hồ sơ đã được lắp vào router này; thông báo và hồ sơ hành chính HR dùng nhóm tài liệu riêng khi có dependency.")
}

func selectedSwaggerDocument(doc map[string]json.RawMessage, title, description string) ([]byte, error) {
	var err error
	// Chỉ giữ components có thể tới từ paths/security được công bố.
	var all map[string]map[string]json.RawMessage
	if err = json.Unmarshal(doc["components"], &all); err != nil {
		return nil, err
	}
	keep := map[string]map[string]json.RawMessage{}
	var visit func(any)
	visit = func(value any) {
		switch v := value.(type) {
		case map[string]any:
			if ref, ok := v["$ref"].(string); ok {
				parts := strings.Split(ref, "/")
				if len(parts) == 4 && parts[0] == "#" && parts[1] == "components" {
					group, key := parts[2], parts[3]
					if keep[group] == nil {
						keep[group] = map[string]json.RawMessage{}
					}
					if _, seen := keep[group][key]; !seen {
						payload := all[group][key]
						keep[group][key] = payload
						var next any
						if json.Unmarshal(payload, &next) == nil {
							visit(next)
						}
					}
				}
			}
			if sec, ok := v["security"].([]any); ok {
				for _, req := range sec {
					if names, ok := req.(map[string]any); ok {
						for name := range names {
							if keep["securitySchemes"] == nil {
								keep["securitySchemes"] = map[string]json.RawMessage{}
							}
							keep["securitySchemes"][name] = all["securitySchemes"][name]
						}
					}
				}
			}
			for _, child := range v {
				visit(child)
			}
		case []any:
			for _, child := range v {
				visit(child)
			}
		}
	}
	var roots any
	rootRaw, _ := json.Marshal(map[string]json.RawMessage{"paths": doc["paths"], "security": doc["security"]})
	if err = json.Unmarshal(rootRaw, &roots); err != nil {
		return nil, err
	}
	visit(roots)
	doc["components"], err = json.Marshal(keep)
	if err != nil {
		return nil, err
	}
	var info map[string]any
	if err = json.Unmarshal(doc["info"], &info); err != nil {
		return nil, err
	}
	info["title"] = title
	info["description"] = description
	doc["info"], err = json.Marshal(info)
	if err != nil {
		return nil, err
	}
	return json.Marshal(doc)
}
