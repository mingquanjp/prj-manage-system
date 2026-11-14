import { test } from "node:test"
import assert from "node:assert/strict"
import { createAuthApi, AuthError, parsePrincipal, authErrorMessage } from "./auth-api.ts"
import { authMode } from "./mode.ts"

export const person = {
  user_id: 12,
  employee_id: 25,
  employee_code: "E025",
  fullname: "Nhân viên thử",
  system_roles: ["HR"],
  must_change_password: false,
}
const ok = (value: unknown) => Response.json(value)
const fetchWith = (fn: (url: string, init: RequestInit) => Response | Promise<Response>) =>
  ((url: string | URL | Request, init?: RequestInit) => fn(String(url), init!)) as typeof fetch

test("bốn route dùng cookie same-origin, không Origin/Cookie/secret thủ công; logout body rỗng", async () => {
  const calls: { path: string; init: RequestInit }[] = []
  const token = "a".repeat(43)
  const api = createAuthApi(
    fetchWith((path, init) => {
      calls.push({ path, init })
      if (path.endsWith("logout")) return new Response(null, { status: 204 })
      return ok(path.endsWith("csrf") ? { csrf_token: token } : person)
    }),
  )
  assert.deepEqual(await api.login(" a@example.com ", "  pass  "), person)
  await api.me()
  await api.csrf()
  await api.logout(token)
  assert.deepEqual(
    calls.map((call) => call.path),
    ["/api/auth/login", "/api/auth/me", "/api/auth/csrf", "/api/auth/logout"],
  )
  for (const { init } of calls) {
    assert.equal(init.credentials, "same-origin")
    assert.equal(init.cache, "no-store")
    assert.equal(init.redirect, "error")
    const headers = new Headers(init.headers)
    assert.equal(headers.has("Cookie"), false)
    assert.equal(headers.has("Origin"), false)
  }
  assert.deepEqual(JSON.parse(calls[0].init.body as string), {
    email: " a@example.com ",
    password: "  pass  ",
  })
  assert.equal(calls[1].init.body, undefined)
  assert.equal(calls[3].init.body, undefined)
  assert.equal(new Headers(calls[3].init.headers).get("X-CSRF-Token"), token)
})

test("principal phân biệt user/employee, roles rỗng hợp lệ và phiên hạn chế không có role", () => {
  assert.equal(parsePrincipal(person).employee_id, 25)
  assert.deepEqual(parsePrincipal({ ...person, system_roles: [] }).system_roles, [])
  assert.equal(
    parsePrincipal({ ...person, system_roles: [], must_change_password: true })
      .must_change_password,
    true,
  )
  for (const value of [
    null,
    [],
    {},
    { ...person, user_id: 0 },
    { ...person, employee_id: 1.1 },
    { ...person, secret: "không được trả" },
    { ...person, fullname: "" },
    { ...person, system_roles: ["PM"] },
    { ...person, system_roles: ["HR", "HR"] },
    { ...person, must_change_password: true },
  ]) {
    assert.throws(() => parsePrincipal(value), AuthError)
  }
})

test("401/403/429/503 giữ status/code; chỉ nhận Retry-After giây hợp lệ, không hiển thị body thô", async () => {
  for (const status of [401, 403, 429, 503]) {
    const api = createAuthApi(
      fetchWith(() =>
        Response.json(
          { code: "TEST_CODE", message: "secret không hiển thị" },
          { status, headers: { "Retry-After": "17" } },
        ),
      ),
    )
    await assert.rejects(api.me(), (error: unknown) => {
      assert.ok(error instanceof AuthError)
      assert.equal(error.status, status)
      assert.equal(error.code, "TEST_CODE")
      assert.equal(error.retryAfterSeconds, 17)
      assert.doesNotMatch(authErrorMessage(error), /secret/)
      return true
    })
  }
  for (const value of ["bad", "-1", "1.5", "99999999999999999999999999"]) {
    await assert.rejects(
      createAuthApi(
        fetchWith(
          () => new Response("proxy failure", { status: 429, headers: { "Retry-After": value } }),
        ),
      ).me(),
      (error: unknown) => error instanceof AuthError && error.retryAfterSeconds === null,
    )
  }
})

test("mất mạng, proxy HTML, JSON lỗi, success status/DTO sai không dựng principal", async () => {
  const replies = [
    () => {
      throw new Error("credential trong lỗi mạng")
    },
    () => new Response("<html>proxy</html>"),
    () => new Response("{", { headers: { "Content-Type": "application/json" } }),
    () => ok({ ...person, password_hash: "không hợp lệ" }),
    () => new Response(null, { status: 204 }),
  ]
  for (const reply of replies) await assert.rejects(createAuthApi(fetchWith(reply)).me(), AuthError)
  await assert.rejects(createAuthApi(fetchWith(() => ok({ csrf_token: "" }))).csrf(), AuthError)
  await assert.rejects(createAuthApi(fetchWith(() => ok({}))).logout("token"), AuthError)
})

test("mode chỉ do cấu hình build, mặc định demo tương thích; giá trị lạ từ chối", () => {
  assert.equal(authMode(undefined), "demo")
  assert.equal(authMode("demo"), "demo")
  assert.equal(authMode("server"), "server")
  for (const value of ["", "SERVER", "?demo=true", "invalid"])
    assert.equal(authMode(value), "invalid")
})

test("đổi mật khẩu chỉ gửi current/new nguyên vẹn với CSRF; 204 không đọc JSON", async () => {
  const api = createAuthApi(
    fetchWith((url, init) => {
      assert.equal(url, "/api/auth/change-password")
      assert.equal(init.method, "POST")
      assert.equal(init.credentials, "same-origin")
      assert.equal(init.cache, "no-store")
      assert.equal(new Headers(init.headers).get("X-CSRF-Token"), "csrf-test")
      assert.equal(new Headers(init.headers).get("Content-Type"), "application/json")
      assert.deepEqual(JSON.parse(init.body as string), {
        current_password: " 12345678 ",
        new_password: " mật khẩu mới đủ dài 😀 ",
      })
      return new Response(null, { status: 204 })
    }),
  )
  await api.changePassword(" 12345678 ", " mật khẩu mới đủ dài 😀 ", "csrf-test")
  await assert.rejects(
    createAuthApi(fetchWith(() => Response.json({}))).changePassword("old", "new", "token"),
    AuthError,
  )
})

test("đổi mật khẩu giữ mã lỗi contract/Retry-After, lỗi mạng không tự gửi lại", async () => {
  for (const [status, code] of [
    [400, "INVALID_CURRENT_PASSWORD"],
    [400, "VALIDATION_ERROR"],
    [401, "UNAUTHENTICATED"],
    [403, "FORBIDDEN"],
    [403, "PASSWORD_CHANGE_NOT_REQUIRED"],
    [413, "PAYLOAD_TOO_LARGE"],
    [415, "UNSUPPORTED_MEDIA_TYPE"],
    [429, "RATE_LIMITED"],
    [500, "INTERNAL_ERROR"],
    [503, "SERVICE_UNAVAILABLE"],
  ] as const) {
    await assert.rejects(
      createAuthApi(
        fetchWith(() => Response.json({ code }, { status, headers: { "Retry-After": "8" } })),
      ).changePassword("old", "new", "token"),
      (error: unknown) =>
        error instanceof AuthError && error.status === status && error.code === code,
    )
  }
  let calls = 0
  await assert.rejects(
    createAuthApi(
      fetchWith(() => {
        calls++
        throw new Error("mất phản hồi")
      }),
    ).changePassword("old", "new", "token"),
    AuthError,
  )
  assert.equal(calls, 1)
})
