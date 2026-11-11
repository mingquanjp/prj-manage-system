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

test("bootstrap Swagger chặn cấu hình ngoài và chỉ gửi auth cùng origin", async () => {
  let config
  let ready
  const source = JSON.parse(await readFile(new URL("../docs/api/auth.openapi.json", import.meta.url)))
  const spec = structuredClone(source)
  vm.runInNewContext(await readFile(new URL("init.js", base), "utf8"), {
    URL,
    window: { location: { href: "http://localhost:8080/docs/?url=https://evil.invalid/spec", origin: "http://localhost:8080" }, addEventListener: (_, fn) => { ready = fn() } },
    fetch: async (url, options) => {
      assert.equal(url, "/docs/openapi.json")
      assert.equal(options.credentials, "same-origin")
      assert.equal(options.cache, "no-store")
      return { ok: true, json: async () => spec }
    },
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
