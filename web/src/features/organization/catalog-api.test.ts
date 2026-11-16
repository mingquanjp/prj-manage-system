import { test } from "node:test"
import assert from "node:assert/strict"
import {
  createCatalogApi,
  catalogKinds,
  validCatalogName,
  trimCatalogName,
  catalogErrorMessage,
} from "./catalog-api.ts"
import { AuthError } from "../session/auth-api.ts"
const json = (body: unknown, status = 200) =>
  new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } })
test("ba danh mục gửi list literal và mutation allowlist/cookie/CSRF, giữ tên nền chính xác", async () => {
  for (const kind of catalogKinds) {
    const seen: { url: string; init: RequestInit }[] = []
    const api = createCatalogApi((async (url, init) => {
      seen.push({ url: String(url), init: init! })
      if (!init?.method)
        return json({
          items: [{ id: 7, name: "Tên cũ" }],
          pagination: { page: 2, page_size: 10, total: 11 },
        })
      return init.method === "DELETE"
        ? new Response(null, { status: 204 })
        : json({ id: 7, name: "Tên mới" }, init.method === "POST" ? 201 : 200)
    }) as typeof fetch)
    await api.list(kind, "a%_\\ + có dấu", 2, 10)
    const old = { id: 7, name: "Tên cũ " }
    await api.create(kind, "Tên mới", "csrf")
    await api.rename(kind, old, "Tên mới", "csrf")
    await api.remove(kind, old, "csrf")
    assert.equal(new URL(seen[0].url, "http://local").searchParams.get("q"), "a%_\\ + có dấu")
    for (const call of seen) {
      assert.equal(call.init.credentials, "same-origin")
      assert.equal(call.init.cache, "no-store")
      assert.equal(call.init.redirect, "error")
      assert.equal(new Headers(call.init.headers).has("Cookie"), false)
      assert.equal(new Headers(call.init.headers).has("Origin"), false)
    }
    assert.deepEqual(
      seen.slice(1).map((call) => [call.url, call.init.method, JSON.parse(String(call.init.body))]),
      [
        [`/api/${kind}`, "POST", { name: "Tên mới" }],
        [`/api/${kind}/7`, "PATCH", { name: "Tên mới", expected_before_name: old.name }],
        [`/api/${kind}/7`, "DELETE", { expected_before_name: old.name }],
      ],
    )
    assert.equal(new Headers(seen[2].init.headers).get("X-CSRF-Token"), "csrf")
  }
})
test("lỗi HTTP/DB/network giữ status/code, không message thô hoặc retry mutation", async () => {
  for (const status of [400, 401, 403, 404, 409, 500, 503]) {
    let calls = 0
    const api = createCatalogApi((async () => {
      calls++
      return json({ code: "STALE_DATA", message: "secret" }, status)
    }) as typeof fetch)
    await assert.rejects(
      api.create("departments", "Tên", "token"),
      (e: unknown) =>
        e instanceof AuthError &&
        e.status === status &&
        e.code === "STALE_DATA" &&
        !e.message.includes("secret"),
    )
    assert.equal(calls, 1)
  }
  let calls = 0
  const api = createCatalogApi((async () => {
    calls++
    throw new Error("raw secret")
  }) as typeof fetch)
  await assert.rejects(api.remove("positions", { id: 1, name: "old" }, "token"), /NETWORK_ERROR/)
  assert.equal(calls, 1)
  assert.match(catalogErrorMessage(new AuthError(503, "SERVICE_UNAVAILABLE"), true), /đối soát/)
})
test("DTO/status/schema lỗi không thành success", async () => {
  const base = { items: [{ id: 1, name: "Tên" }], pagination: { page: 1, page_size: 20, total: 1 } }
  for (const body of [
    null,
    { ...base, raw: "secret" },
    { ...base, items: null },
    { ...base, items: [{ id: 1, name: "Tên", users: [] }] },
    { ...base, pagination: { page: 2, page_size: 20, total: 1 } },
    { ...base, pagination: { page: 1, page_size: 20, total: -1 } },
  ]) {
    const api = createCatalogApi((async () => json(body)) as typeof fetch)
    await assert.rejects(api.list("departments", "", 1, 20), /INVALID_RESPONSE/)
  }
  for (const response of [new Response("<html>proxy</html>"), json({ id: 1, name: "Tên" }, 200)]) {
    const api = createCatalogApi((async () => response) as typeof fetch)
    await assert.rejects(api.create("departments", "Tên", "csrf"), /INVALID_RESPONSE/)
  }
})
test("validation Go: 100/101 code point, Unicode whitespace, không bỏ dấu/NFC", () => {
  assert.equal(validCatalogName("😀".repeat(100)), true)
  assert.equal(validCatalogName("😀".repeat(101)), false)
  assert.equal(validCatalogName("\u0085\u2003\u3000"), false)
  assert.equal(trimCatalogName("\u0085 Phòng Á \u3000"), "Phòng Á")
  assert.equal(trimCatalogName("e\u0301"), "e\u0301")
  assert.equal(trimCatalogName("\uFEFFtên\uFEFF"), "\uFEFFtên\uFEFF")
  assert.equal(validCatalogName("a\0b"), false)
  assert.equal(validCatalogName("\ud800"), false)
})
