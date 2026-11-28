/* Cấu hình local cố định, không nhận spec/config từ query string. */
window.addEventListener("DOMContentLoaded", async () => {
  try {
    const load = async (url, optional = false) => {
      const response = await fetch(url, { credentials: "same-origin", cache: "no-store" })
      if (optional && response.status === 404) return null
      if (!response.ok) throw new Error("Không tải được contract local.")
      return response.json()
    }
    const specs = { auth: await load("/docs/openapi.json") }
    const catalogs = await load("/docs/catalog.openapi.json", true)
    if (catalogs) specs.catalogs = catalogs
    const profileChanges = await load("/docs/profile-change.openapi.json", true)
    if (profileChanges) specs.profileChanges = profileChanges
    const notifications = await load("/docs/notifications.openapi.json", true)
    if (notifications) specs.notifications = notifications
    const administrativeProfiles = await load("/docs/administrative-profile.openapi.json", true)
    if (administrativeProfiles) specs.administrativeProfiles = administrativeProfiles
    const skillCatalogs = await load("/docs/skill-catalog.openapi.json", true)
    if (skillCatalogs) specs.skillCatalogs = skillCatalogs
    const operations = []
    for (const spec of Object.values(specs)) {
      for (const [path, methods] of Object.entries(spec.paths)) {
        for (const [method, operation] of Object.entries(methods)) {
          if (!["get", "post", "patch", "delete"].includes(method)) continue
          operations.push([method.toUpperCase(), path])
          for (const parameter of operation.parameters || []) {
            if (parameter.in === "header" && parameter.name.toLowerCase() === "origin") parameter.schema.default = window.location.origin
          }
        }
      }
      for (const parameter of Object.values(spec.components?.parameters || {})) {
        if (parameter.in === "header" && parameter.name.toLowerCase() === "origin") parameter.schema.default = window.location.origin
      }
    }
    const render = (spec) => SwaggerUIBundle({
      spec,
      dom_id: "#swagger-ui",
      validatorUrl: null,
      queryConfigEnabled: false,
      persistAuthorization: false,
      withCredentials: false,
      supportedSubmitMethods: ["get", "post", "patch", "delete"],
      deepLinking: true,
      docExpansion: "list",
      displayRequestDuration: true,
      requestInterceptor(request) {
        const target = new URL(request.url, window.location.href)
        const method = (request.method || "GET").toUpperCase()
        const specRequest = method === "GET" && ["/docs/openapi.json", "/docs/catalog.openapi.json", "/docs/profile-change.openapi.json", "/docs/notifications.openapi.json", "/docs/administrative-profile.openapi.json", "/docs/skill-catalog.openapi.json"].includes(target.pathname)
        const allowed = operations.some(([verb, path]) => {
          if (verb !== method) return false
          if (!path.includes("{id}")) return path === target.pathname
          const [base, suffix] = path.split("{id}")
          if (!target.pathname.startsWith(base) || !target.pathname.endsWith(suffix)) return false
          const id = target.pathname.slice(base.length, suffix ? -suffix.length : undefined)
          return /^[0-9]+$/.test(id) && Number(id) >= 1 && Number(id) <= 2147483647
        })
        if (target.origin !== window.location.origin || (!specRequest && !allowed)) throw new Error("Chỉ được gọi contract local và API đã triển khai cùng origin.")
        request.credentials = "same-origin"
        for (const name of Object.keys(request.headers || {})) {
          if (["cookie", "origin"].includes(name.toLowerCase())) delete request.headers[name]
        }
        return request
      },
    })
    const selector = document.getElementById("api-group")
    if (catalogs) {
      const option = document.createElement("option")
      option.value = "catalogs"
      option.textContent = "Tổ chức, tài khoản và hồ sơ"
      selector.append(option)
    }
    if (profileChanges) {
      const option = document.createElement("option")
      option.value = "profileChanges"
      option.textContent = "Đề nghị hồ sơ"
      selector.append(option)
    }
    if (notifications) {
      const option = document.createElement("option")
      option.value = "notifications"
      option.textContent = "Thông báo hồ sơ chính chủ"
      selector.append(option)
    }
    if (skillCatalogs) {
      const option = document.createElement("option")
      option.value = "skillCatalogs"
      option.textContent = "Danh mục kỹ năng"
      selector.append(option)
    }
    selector.addEventListener("change", () => render(specs[selector.value]))
    if (administrativeProfiles) {
      const option = document.createElement("option")
      option.value = "administrativeProfiles"
      option.textContent = "Hồ sơ hành chính HR"
      selector.append(option)
    }
    render(specs.auth)
  } catch {
    document.getElementById("swagger-ui").textContent = "Không thể mở tài liệu API. Hãy kiểm tra server và tải lại trang."
  }
})
