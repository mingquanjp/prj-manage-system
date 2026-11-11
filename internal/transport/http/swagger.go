package http

import (
	"bytes"
	"embed"
	"net/http"
	"time"

	apispec "ai-staffing/docs/api"
)

// Chỉ nhúng các file phục vụ UI; license và manifest vẫn giữ trong nguồn.
//
//go:embed swagger/index.html swagger/init.js swagger/docs.css swagger/swagger-ui-bundle.js swagger/swagger-ui.css
var swaggerFiles embed.FS

// Chỉ áp dụng cho trang tài liệu, không thay header hoặc bảo vệ của API.
const swaggerCSP = "default-src 'none'; script-src 'self'; style-src 'self'; style-src-attr 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'"

func swaggerHandler(w http.ResponseWriter, r *http.Request) {
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
