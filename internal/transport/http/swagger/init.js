/* Cấu hình local cố định, không nhận spec/config từ query string. */
window.addEventListener("DOMContentLoaded", async () => {
  const operations = new Set([
    "POST /api/auth/login", "GET /api/auth/me", "GET /api/auth/csrf",
    "POST /api/auth/logout", "POST /api/auth/change-password",
  ])
  try {
    const response = await fetch("/docs/openapi.json", { credentials: "same-origin", cache: "no-store" })
    if (!response.ok) throw new Error("Không tải được contract local.")
    const spec = await response.json()
    // Form Swagger kiểm required trước interceptor. Default chỉ giúp nhập liệu;
    // endpoint canonical vẫn giữ nguyên và Origin thật do browser quản lý.
    for (const methods of Object.values(spec.paths)) {
      for (const operation of Object.values(methods)) {
        for (const parameter of operation.parameters || []) {
          if (parameter.in === "header" && parameter.name.toLowerCase() === "origin") {
            parameter.schema.default = window.location.origin
          }
        }
      }
    }
    SwaggerUIBundle({
      spec,
      dom_id: "#swagger-ui",
      validatorUrl: null,
      queryConfigEnabled: false,
      persistAuthorization: false,
      withCredentials: false,
      supportedSubmitMethods: ["get", "post"],
      deepLinking: true,
      docExpansion: "list",
      displayRequestDuration: true,
      requestInterceptor(request) {
        const target = new URL(request.url, window.location.href)
        const method = (request.method || "GET").toUpperCase()
        const spec = method === "GET" && target.pathname === "/docs/openapi.json"
        if (target.origin !== window.location.origin || (!spec && !operations.has(`${method} ${target.pathname}`))) {
          throw new Error("Chỉ được gọi contract local và năm API auth cùng origin.")
        }
        request.credentials = "same-origin"
        // Hai header này do trình duyệt quản lý; ô Authorize không cấp session.
        for (const name of Object.keys(request.headers || {})) {
          if (["cookie", "origin"].includes(name.toLowerCase())) delete request.headers[name]
        }
        return request
      },
    })
  } catch {
    document.getElementById("swagger-ui").textContent = "Không thể mở tài liệu API. Hãy kiểm tra server và tải lại trang."
  }
})
