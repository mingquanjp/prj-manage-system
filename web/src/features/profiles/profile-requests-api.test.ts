import { test } from "node:test"
import assert from "node:assert/strict"
import { AuthError } from "../session/auth-api.ts"
import {
  createProfileRequestsApi,
  parseRequestDetail,
  parseRequestPage,
  requestErrorMessage,
} from "./profile-requests-api.ts"
import { buildProfileRequest } from "./profile-request-input.ts"
const detail = {
  request_id: 42,
  employee_id: 25,
  employee_code: "EMP000025",
  fullname: "Nhân viên thử",
  status: "PENDING",
  before_values: { address: null },
  proposed_values: { address: "Địa chỉ mới" },
  current_values: { address: null },
  explanation: null,
  requested_by: 12,
  requested_at: "2026-11-24T09:00:00+07:00",
  resolved_by: null,
  resolved_at: null,
  decision_reason: null,
}
const summary = {
  request_id: 42,
  employee_id: 25,
  employee_code: "EMP000025",
  fullname: "Nhân viên thử",
  status: "PENDING",
  changed_fields: ["address"],
  requested_at: detail.requested_at,
  resolved_at: null,
}
const page = { items: [summary], page: 1, page_size: 20, total: 1 }
const json = (body: unknown, status = 200) =>
  new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } })
test("self list/detail dùng GET không quyền/PII client, submit allowlist/CSRF/key đúng một lần", async () => {
  const calls: { url: string; init?: RequestInit }[] = []
  const api = createProfileRequestsApi((async (url, init) => {
    calls.push({ url: String(url), init })
    return json(
      init?.method === "POST" ? detail : String(url).includes("?") ? page : detail,
      init?.method === "POST" ? 201 : 200,
    )
  }) as typeof fetch)
  assert.deepEqual(await api.list(25, 1), page)
  assert.deepEqual(await api.read(25, 42), detail)
  const input = buildProfileRequest(
    { fullname: "Cũ", phone: null, address: null },
    { fullname: "Cũ", phone: null, address: "Địa chỉ mới" },
    "",
  )
  assert.deepEqual(
    await api.submit(25, input, "A0000000-0000-4000-8000-000000000000", "csrf"),
    detail,
  )
  assert.equal(calls[0].url, "/api/profile-change-requests?scope=self&page=1&page_size=20")
  assert.equal(calls[1].url, "/api/profile-change-requests/42")
  assert.equal(calls[2].url, "/api/employees/25/profile-change-requests")
  for (const { init } of calls) {
    assert.equal(init?.credentials, "same-origin")
    assert.equal(init?.redirect, "error")
    assert.equal(init?.cache, "no-store")
    for (const key of ["Cookie", "Origin", "Authorization"])
      assert.equal(new Headers(init?.headers).has(key), false)
  }
  assert.equal(calls[0].init?.body, undefined)
  const headers = new Headers(calls[2].init?.headers)
  assert.equal(headers.get("Idempotency-Key"), "a0000000-0000-4000-8000-000000000000")
  assert.equal(headers.get("X-CSRF-Token"), "csrf")
  assert.deepEqual(JSON.parse(String(calls[2].init?.body)), input)
  assert.equal(calls.length, 3)
})
test("DTO exact/self/null/state/patch; chặn PII người khác và field ngoài allowlist", () => {
  assert.deepEqual(parseRequestDetail(detail, 25, 42), detail)
  for (const variant of [
    { ...detail, employee_id: 12 },
    { ...detail, request_id: 99 },
    { ...detail, credential: "secret" },
    { ...detail, resolved_by: 12 },
    { ...detail, resolved_at: detail.requested_at },
    { ...detail, current_values: { phone: null } },
    { ...detail, proposed_values: { address: null } },
    { ...detail, before_values: { address: "Địa chỉ mới" } },
    { ...detail, explanation: "\ud800" },
    {
      ...detail,
      status: "REJECTED",
      resolved_by: 12,
      resolved_at: detail.requested_at,
      decision_reason: null,
    },
  ])
    assert.throws(() => parseRequestDetail(variant, 25, 42), /INVALID_RESPONSE/)
  const approved = {
    ...detail,
    status: "APPROVED",
    resolved_by: 8,
    resolved_at: detail.requested_at,
  }
  assert.doesNotThrow(() => parseRequestDetail(approved, 25))
  assert.doesNotThrow(() =>
    parseRequestDetail(
      {
        ...detail,
        before_values: { address: "x".repeat(3000) },
        current_values: { address: "\n padded " },
      },
      25,
    ),
  )
  for (const variant of [
    { ...page, items: [{ ...summary, address: "PII" }] },
    { ...page, items: [{ ...summary, employee_id: 12 }] },
    { ...page, items: [summary, summary], total: 2 },
    { ...page, page: 2 },
    { ...page, total: -1 },
    { ...page, items: [{ ...summary, changed_fields: ["address", "address"] }] },
  ])
    assert.throws(() => parseRequestPage(variant, 25, 1), /INVALID_RESPONSE/)
  assert.deepEqual(parseRequestPage({ ...page, items: [], page: 3 }, 25, 3).total, 1)
})
test("HTTP/mạng/response lỗi được phân loại, redaction và không tự replay", async () => {
  for (const [status, code] of [
    [401, "UNAUTHENTICATED"],
    [403, "PASSWORD_CHANGE_REQUIRED"],
    [404, "NOT_FOUND"],
    [409, "STALE_DATA"],
    [503, "SERVICE_UNAVAILABLE"],
  ] as const) {
    let calls = 0
    const api = createProfileRequestsApi((async () => {
      calls++
      return json({ code, message: "PII-marker" }, status)
    }) as typeof fetch)
    await assert.rejects(
      api.read(25, 42),
      (e: unknown) =>
        e instanceof AuthError &&
        e.status === status &&
        e.code === code &&
        !e.message.includes("PII"),
    )
    assert.equal(calls, 1)
    assert.ok(requestErrorMessage(new AuthError(status, code)))
  }
  let calls = 0
  const api = createProfileRequestsApi((async () => {
    calls++
    throw new Error("private-marker")
  }) as typeof fetch)
  const input = buildProfileRequest(
    { fullname: "Cũ", phone: null, address: null },
    { fullname: "Mới", phone: null, address: null },
    "",
  )
  await assert.rejects(
    api.submit(25, input, "a0000000-0000-4000-8000-000000000000", "csrf"),
    /NETWORK_ERROR/,
  )
  assert.equal(calls, 1)
  await assert.rejects(api.submit(25, input, "not-v4", "csrf"), /VALIDATION_ERROR/)
  assert.equal(calls, 1)
  const malformed = createProfileRequestsApi((async () =>
    json({ ...detail, raw: true }, 201)) as typeof fetch)
  await assert.rejects(
    malformed.submit(25, input, "a0000000-0000-4000-8000-000000000000", "csrf"),
    /INVALID_RESPONSE/,
  )
})

test("không nhận status/media/DTO sai; abort và raw provider code không lộ dữ liệu", async () => {
  for (const response of [
    json(detail, 201),
    new Response(JSON.stringify(detail), { headers: { "Content-Type": "text/html" } }),
    new Response("{bad", { headers: { "Content-Type": "application/json" } }),
    json({ ...detail, employee_id: 26 }),
  ]) {
    const api = createProfileRequestsApi((async () => response) as typeof fetch)
    await assert.rejects(api.read(25, 42), /INVALID_RESPONSE/)
  }
  const denied = createProfileRequestsApi((async () =>
    json({ code: "PII-marker", message: "private-marker" }, 503)) as typeof fetch)
  await assert.rejects(
    denied.read(25, 42),
    (error: unknown) => error instanceof AuthError && error.code === "HTTP_ERROR",
  )
  let calls = 0
  const api = createProfileRequestsApi((async (_url, init) => {
    calls++
    assert.equal(init?.signal?.aborted, true)
    throw new DOMException("private-marker", "AbortError")
  }) as typeof fetch)
  const controller = new AbortController()
  controller.abort()
  await assert.rejects(api.read(25, 42, controller.signal), /NETWORK_ERROR/)
  assert.equal(calls, 1)
  for (const id of [0, -1, 2147483648, NaN]) {
    await assert.rejects(api.read(id, 42), /INVALID_RESPONSE/)
    await assert.rejects(api.list(25, id), /INVALID_RESPONSE/)
  }
  assert.equal(calls, 1)
})
