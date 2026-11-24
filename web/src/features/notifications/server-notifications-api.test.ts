import test from "node:test"
import assert from "node:assert/strict"
import { AuthError } from "../session/auth-api.ts"
import {
  createServerNotificationsApi,
  notificationDestination,
  parseNotificationPage,
  type ServerNotification,
} from "./server-notifications-api.ts"

const item: ServerNotification = {
  notification_id: 5,
  notification_type: "PROFILE_CHANGE_SUBMITTED",
  resource_type: "PROFILE_CHANGE_REQUEST",
  resource_id: 12,
  message: "Có đề nghị hồ sơ mới.",
  created_at: "2026-11-25T01:00:00Z",
  read_at: null,
}
const page = { items: [item], page: 1, page_size: 20, total: 1, unread_count: 1 }
const json = (body: unknown, status = 200) =>
  new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } })
test("list/read gửi đúng allowlist, CSRF, cache và không tự phát lại", async () => {
  const calls: { url: string; init?: RequestInit }[] = []
  const api = createServerNotificationsApi((async (url, init) => {
    calls.push({ url: String(url), init })
    return init?.method === "POST" ? new Response(null, { status: 204 }) : json(page)
  }) as typeof fetch)
  assert.deepEqual(await api.list("unread", 1), page)
  await api.read(5, "csrf")
  assert.equal(calls[0].url, "/api/notifications?filter=unread&page=1&page_size=20")
  assert.equal(calls[1].url, "/api/notifications/5/read")
  assert.equal(calls[1].init?.body, "{}")
  assert.equal(new Headers(calls[1].init?.headers).get("X-CSRF-Token"), "csrf")
  for (const { init } of calls) {
    assert.equal(init?.cache, "no-store")
    assert.equal(init?.credentials, "same-origin")
    assert.equal(init?.redirect, "error")
    assert.equal(new Headers(init?.headers).has("Cookie"), false)
    assert.equal(new Headers(init?.headers).has("Origin"), false)
  }
  assert.equal(calls.length, 2)
})
test("DTO strict, own filter/count và ID/resource allowlist", () => {
  assert.deepEqual(parseNotificationPage(page, "all", 1), page)
  assert.equal(parseNotificationPage({ ...page, items: [], page: 3 }, "all", 3).total, 1)
  for (const variant of [
    { ...page, extra: "PII" },
    { ...page, page: 2 },
    { ...page, page_size: 100 },
    { ...page, total: -1 },
    { ...page, unread_count: 0.5 },
    { ...page, unread_count: 2 },
    { ...page, items: [item, item], total: 2 },
    ...[
      { notification_id: 0 },
      { resource_id: 2147483648 },
      { resource_type: "URL" },
      { notification_type: "NEW_EVENT" },
      { message: "\ud800" },
      { message: "x".repeat(201) },
      { created_at: "invalid" },
      { created_at: "2026-11-25T01:00:00" },
      { read_at: "invalid" },
      { fullname: "PII" },
    ].map((change) => ({ ...page, items: [{ ...item, ...change }] })),
  ])
    assert.throws(() => parseNotificationPage(variant, "all", 1), /INVALID_RESPONSE/)
  assert.throws(
    () => parseNotificationPage({ ...page, unread_count: 2 }, "unread", 1),
    /INVALID_RESPONSE/,
  )
  assert.throws(
    () =>
      parseNotificationPage(
        { ...page, items: [{ ...item, read_at: item.created_at }] },
        "unread",
        1,
      ),
    /INVALID_RESPONSE/,
  )
  assert.equal(notificationDestination(item), "/personnel/profile-requests/12")
  for (const notification_type of ["PROFILE_CHANGE_APPROVED", "PROFILE_CHANGE_REJECTED"] as const)
    assert.equal(notificationDestination({ ...item, notification_type }), "/profile?request_id=12")
  assert.throws(() => notificationDestination({ ...item, resource_id: NaN }), /INVALID_RESPONSE/)
})
test("HTTP/status/media/mạng lỗi redacted; POST unknown không retry", async () => {
  for (const [status, code] of [
    [401, "UNAUTHENTICATED"],
    [403, "PASSWORD_CHANGE_REQUIRED"],
    [404, "NOT_FOUND"],
    [503, "SERVICE_UNAVAILABLE"],
  ] as const) {
    let calls = 0
    const api = createServerNotificationsApi((async () => {
      calls++
      return json({ code, message: "private-marker" }, status)
    }) as typeof fetch)
    await assert.rejects(
      api.read(5, "token"),
      (error: unknown) =>
        error instanceof AuthError &&
        error.status === status &&
        error.code === code &&
        !error.message.includes("private-marker"),
    )
    assert.equal(calls, 1)
  }
  for (const response of [
    json(page, 201),
    new Response("{}", { headers: { "Content-Type": "text/html" } }),
    new Response("{bad", { headers: { "Content-Type": "application/json" } }),
  ]) {
    const api = createServerNotificationsApi((async () => response) as typeof fetch)
    await assert.rejects(api.list(), /INVALID_RESPONSE/)
  }
  const raw = createServerNotificationsApi((async () =>
    json({ code: "private-marker" }, 500)) as typeof fetch)
  await assert.rejects(raw.read(5, "token"), /HTTP_ERROR/)
  let calls = 0
  const lost = createServerNotificationsApi((async () => {
    calls++
    throw new Error("private-marker")
  }) as typeof fetch)
  await assert.rejects(lost.read(5, "token"), /NETWORK_ERROR/)
  assert.equal(calls, 1)
  for (const invalid of [0, -1, NaN, 2147483648]) {
    await assert.rejects(lost.read(invalid, "token"), /VALIDATION_ERROR/)
    await assert.rejects(lost.list("all", invalid), /VALIDATION_ERROR/)
  }
  assert.equal(calls, 1)
  const wrongRead = createServerNotificationsApi((async () => json({})) as typeof fetch)
  await assert.rejects(wrongRead.read(5, "token"), /INVALID_RESPONSE/)
})
