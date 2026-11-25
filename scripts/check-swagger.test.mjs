// Kiểm tài nguyên và cấu hình bootstrap; không thay bằng chứng trình duyệt/API.
import assert from "node:assert/strict"
import { test } from "node:test"
import { readFile } from "node:fs/promises"
import { createHash } from "node:crypto"
import vm from "node:vm"

const base = new URL("../internal/transport/http/swagger/", import.meta.url)

test("tài nguyên Swagger được ghim và khớp checksum vendor", async () => {
  const manifest = JSON.parse(await readFile(new URL("vendor.json", base)))
  assert.equal(manifest.package, "swagger-ui-dist")
  assert.equal(manifest.version, "5.34.1")
  assert.equal(manifest.license, "Apache-2.0")
  for (const [file, sha256] of Object.entries(manifest.files_sha256)) {
    assert.equal(createHash("sha256").update(await readFile(new URL(file, base))).digest("hex"), sha256)
  }
})

test("bootstrap Swagger chặn cấu hình ngoài và chỉ gửi API đã triển khai cùng origin", async () => {
  let config
  let ready
  const source = JSON.parse(await readFile(new URL("../docs/api/auth.openapi.json", import.meta.url)))
  const spec = structuredClone(source)
  vm.runInNewContext(await readFile(new URL("init.js", base), "utf8"), {
    URL,
    window: { location: { href: "http://localhost:8080/docs/?url=https://evil.invalid/spec", origin: "http://localhost:8080" }, addEventListener: (_, fn) => { ready = fn() } },
    fetch: async (url, options) => {
      if (["/docs/catalog.openapi.json", "/docs/profile-change.openapi.json", "/docs/notifications.openapi.json", "/docs/administrative-profile.openapi.json"].includes(url)) return { ok: false, status: 404 }
      assert.equal(url, "/docs/openapi.json")
      assert.equal(options.credentials, "same-origin")
      assert.equal(options.cache, "no-store")
      return { ok: true, json: async () => spec }
    },
    document: { getElementById: () => ({ addEventListener() {} }) },
    SwaggerUIBundle: (value) => { config = value },
  })
  await ready
  assert.equal(config.url, undefined)
  assert.equal(config.spec, spec)
  for (const [path, methods] of Object.entries(source.paths)) {
    for (const [method, operation] of Object.entries(methods)) {
      for (const parameter of operation.parameters || []) {
        if (parameter.name === "Origin") {
          assert.equal(parameter.schema.default, undefined)
          assert.equal(config.spec.paths[path][method].parameters.find((p) => p.name === "Origin").schema.default, "http://localhost:8080")
        }
      }
    }
  }
  assert.equal(config.validatorUrl, null)
  assert.equal(config.queryConfigEnabled, false)
  assert.equal(config.persistAuthorization, false)
  assert.equal(config.withCredentials, false)
  const request = config.requestInterceptor({ url: "/api/auth/logout", method: "POST", headers: { Cookie: "ignored", Origin: "https://evil.invalid", "X-CSRF-Token": "synthetic" } })
  assert.equal(request.credentials, "same-origin")
  assert.deepEqual(Object.keys(request.headers), ["X-CSRF-Token"])
  for (const url of ["https://evil.invalid/api/auth/login", "/api/projects", "/docs/unknown"]) {
    assert.throws(() => config.requestInterceptor({ url, method: "GET", headers: {} }))
  }
  assert.throws(() => config.requestInterceptor({ url: "/api/auth/me", method: "DELETE" }))
  assert.doesNotThrow(() => config.requestInterceptor({ url: "/docs/openapi.json", method: "GET" }))
})

test("bootstrap lỗi tải spec hiển thị thông báo, không dựng UI thành công giả", async () => {
  let ready
  const element = { textContent: "" }
  vm.runInNewContext(await readFile(new URL("init.js", base), "utf8"), {
    window: { addEventListener: (_, fn) => { ready = fn() } },
    fetch: async () => ({ ok: false }),
    document: { getElementById: () => element },
    SwaggerUIBundle: () => assert.fail("không được dựng UI khi spec lỗi"),
  })
  await ready
  assert.match(element.textContent, /Không thể mở tài liệu API/)
})

for (const profiles of [false, true]) test(`Swagger chọn đúng ${profiles ? 13 : 12} operation đã triển khai`, async () => {
  let ready, config, change
  const selector = { value: "auth", append() {}, addEventListener(_, fn) { change = fn } }
  const auth = JSON.parse(await readFile(new URL("../docs/api/auth.openapi.json", import.meta.url)))
  const catalogs = JSON.parse(await readFile(new URL("../docs/api/organization-accounts-profile.openapi.json", import.meta.url)))
  catalogs.paths = Object.fromEntries(Object.entries(catalogs.paths).filter(([p]) => /^\/api\/(departments|positions|job-levels)(\/\{id\})?$/.test(p)))
  if (profiles) {
    const canonical = JSON.parse(await readFile(new URL("../docs/api/organization-accounts-profile.openapi.json", import.meta.url)))
    catalogs.paths["/api/employees/{id}"] = canonical.paths["/api/employees/{id}"]
  }
  vm.runInNewContext(await readFile(new URL("init.js", base), "utf8"), {
    URL,
    window: { location: { href: "http://127.0.0.1:8080/docs/", origin: "http://127.0.0.1:8080" }, addEventListener: (_, fn) => { ready = fn() } },
    fetch: async (url) => ["/docs/profile-change.openapi.json", "/docs/notifications.openapi.json", "/docs/administrative-profile.openapi.json"].includes(url) ? { ok: false, status: 404 } : ({ ok: true, json: async () => url === "/docs/openapi.json" ? auth : catalogs }),
    document: { getElementById: () => selector, createElement: () => ({}) },
    SwaggerUIBundle: value => { config = value },
  })
  await ready
  selector.value = "catalogs"
  change()
  assert.equal(config.spec, catalogs)
  assert.equal(config.spec.components.parameters.Origin.schema.default, "http://127.0.0.1:8080")
  assert.deepEqual(Array.from(config.supportedSubmitMethods), ["get", "post", "patch", "delete"])
  for (const collection of ["departments", "positions", "job-levels"]) {
    for (const method of ["GET", "POST"]) assert.doesNotThrow(() => config.requestInterceptor({ url: `/api/${collection}?page=1`, method }))
    for (const method of ["PATCH", "DELETE"]) assert.doesNotThrow(() => config.requestInterceptor({ url: `/api/${collection}/12`, method }))
  }
  if (profiles) assert.doesNotThrow(() => config.requestInterceptor({ url: "/api/employees/25", method: "GET" }))
  else assert.throws(() => config.requestInterceptor({ url: "/api/employees/25", method: "GET" }))
  assert.throws(() => config.requestInterceptor({ url: "/api/employees/25", method: "PATCH" }))
  for (const [url, method] of [["/api/users", "GET"], ["/api/employees", "POST"], ["/api/departments/0", "PATCH"], ["/api/departments/2147483648", "DELETE"], ["/api/departments/1/extra", "PATCH"], ["/api/departments/1", "GET"], ["https://other.invalid/api/departments", "GET"]]) {
    assert.throws(() => config.requestInterceptor({ url, method }))
  }
})

test("Swagger cho phép ID ở giữa đường dẫn PCA và chặn operation chưa triển khai", async () => {
  let ready, config, change
  const selector = { value: "auth", append() {}, addEventListener(_, fn) { change = fn } }
  const auth = JSON.parse(await readFile(new URL("../docs/api/auth.openapi.json", import.meta.url)))
  const pca = JSON.parse(await readFile(new URL("../docs/api/profile-change-approval.openapi.json", import.meta.url)))
  pca.paths = Object.fromEntries(Object.entries(pca.paths).filter(([p]) => p.includes("profile-change-requests")))
  vm.runInNewContext(await readFile(new URL("init.js", base), "utf8"), {
    URL,
    window: { location: { href: "http://127.0.0.1:8080/docs/", origin: "http://127.0.0.1:8080" }, addEventListener: (_, fn) => { ready = fn() } },
    fetch: async url => ["/docs/catalog.openapi.json", "/docs/notifications.openapi.json", "/docs/administrative-profile.openapi.json"].includes(url) ? { ok: false, status: 404 } : { ok: true, json: async () => url === "/docs/openapi.json" ? auth : pca },
    document: { getElementById: () => selector, createElement: () => ({}) },
    SwaggerUIBundle: value => { config = value },
  })
  await ready
  selector.value = "profileChanges"
  change()
  assert.equal(config.spec, pca)
  for (const [url, method] of [["/api/employees/25/profile-change-requests", "POST"], ["/api/profile-change-requests", "GET"], ["/api/profile-change-requests/42", "GET"], ["/api/profile-change-requests/42/approve", "POST"], ["/api/profile-change-requests/42/reject", "POST"]]) {
    assert.doesNotThrow(() => config.requestInterceptor({ url, method }))
  }
  for (const [url, method] of [["/api/employees/0/profile-change-requests", "POST"], ["/api/profile-change-requests/2147483648/approve", "POST"], ["/api/profile-change-requests/42/approve/extra", "POST"], ["/api/notifications", "GET"], ["/api/hr/employees", "GET"]]) {
    assert.throws(() => config.requestInterceptor({ url, method }))
  }
})

test("Swagger notification dùng đúng suffix read và không mở API PCA chưa triển khai", async () => {
  let ready, config, change
  const selector = { value: "auth", append() {}, addEventListener(_, fn) { change = fn } }
  const auth = JSON.parse(await readFile(new URL("../docs/api/auth.openapi.json", import.meta.url)))
  const notifications = JSON.parse(await readFile(new URL("../docs/api/profile-change-approval.openapi.json", import.meta.url)))
  notifications.paths = Object.fromEntries(Object.entries(notifications.paths).filter(([p]) => p.startsWith("/api/notifications")))
  vm.runInNewContext(await readFile(new URL("init.js", base), "utf8"), {
    URL,
    window: { location: { href: "http://127.0.0.1:8080/docs/", origin: "http://127.0.0.1:8080" }, addEventListener: (_, fn) => { ready = fn() } },
    fetch: async url => ["/docs/catalog.openapi.json", "/docs/profile-change.openapi.json", "/docs/administrative-profile.openapi.json"].includes(url) ? { ok: false, status: 404 } : { ok: true, json: async () => url === "/docs/openapi.json" ? auth : notifications },
    document: { getElementById: () => selector, createElement: () => ({}) },
    SwaggerUIBundle: value => { config = value },
  })
  await ready
  selector.value = "notifications"; change()
  assert.equal(config.spec, notifications)
  for (const [url, method] of [["/api/notifications?filter=unread", "GET"], ["/api/notifications/42/read", "POST"]]) assert.doesNotThrow(() => config.requestInterceptor({ url, method }))
  for (const [url, method] of [["/api/notifications/42", "POST"], ["/api/notifications/42/read/extra", "POST"], ["/api/notifications/0/read", "POST"], ["/api/notifications/2147483648/read", "POST"], ["/api/notifications/42/read", "GET"], ["/api/employees/42/profile-change-requests", "POST"]]) assert.throws(() => config.requestInterceptor({ url, method }))
})

test("Swagger HR công bố ba operation F4 riêng và giữ allowlist cùng origin", async () => {
  let ready, config, change
  const selector = { value: "auth", append() {}, addEventListener(_, fn) { change = fn } }
  const auth = JSON.parse(await readFile(new URL("../docs/api/auth.openapi.json", import.meta.url)))
  const administrative = JSON.parse(await readFile(new URL("../docs/api/profile-change-approval.openapi.json", import.meta.url)))
  administrative.paths = Object.fromEntries(Object.entries(administrative.paths).filter(([p]) => ["/api/employees", "/api/employees/{id}/administrative-profile", "/api/employees/{id}"].includes(p)))
  vm.runInNewContext(await readFile(new URL("init.js", base), "utf8"), {
    URL,
    window: { location: { href: "http://127.0.0.1:8080/docs/", origin: "http://127.0.0.1:8080" }, addEventListener: (_, fn) => { ready = fn() } },
    fetch: async url => ["/docs/catalog.openapi.json", "/docs/profile-change.openapi.json", "/docs/notifications.openapi.json"].includes(url) ? { ok: false, status: 404 } : { ok: true, json: async () => url === "/docs/openapi.json" ? auth : administrative },
    document: { getElementById: () => selector, createElement: () => ({}) },
    SwaggerUIBundle: value => { config = value },
  })
  await ready
  selector.value = "administrativeProfiles"; change()
  assert.equal(config.spec, administrative)
  for (const [url, method] of [["/api/employees?q=demo", "GET"], ["/api/employees/42/administrative-profile", "GET"], ["/api/employees/42", "PATCH"], ["/docs/administrative-profile.openapi.json", "GET"]]) assert.doesNotThrow(() => config.requestInterceptor({ url, method }))
  for (const [url, method] of [["/api/employees/42", "GET"], ["/api/employees", "POST"], ["/api/employees/0", "PATCH"], ["/api/employees/2147483648/administrative-profile", "GET"], ["/api/employees/42/administrative-profile/extra", "GET"], ["https://other.invalid/api/employees/42", "PATCH"]]) assert.throws(() => config.requestInterceptor({ url, method }))
})
