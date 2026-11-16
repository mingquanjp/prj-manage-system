import { test } from "node:test"
import assert from "node:assert/strict"
import { AuthError } from "../session/auth-api.ts"
import { createProfileApi, parseSelfProfile, profileErrorMessage } from "./profile-api.ts"
const actor = { user_id: 12, employee_id: 25 }
const profile = {
  ...actor,
  employee_code: "E025",
  fullname: "Nhân viên tổng hợp",
  company_email: "test@example.invalid",
  phone: null,
  address: null,
  department: null,
  job_level: null,
  position: { id: 7, name: "Kỹ sư" },
  employment_status: "ACTIVE",
}
const json = (body: unknown, status = 200) =>
  new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } })
test("profile GET dùng employee_id từ principal và cookie browser; không query/body/CSRF", async () => {
  let count = 0
  const api = createProfileApi((async (url, init) => {
    count++
    assert.equal(url, "/api/employees/25")
    assert.equal(init?.method, "GET")
    assert.equal(init?.credentials, "same-origin")
    assert.equal(init?.cache, "no-store")
    assert.equal(init?.redirect, "error")
    assert.equal(init?.body, undefined)
    for (const name of ["Cookie", "Origin", "X-CSRF-Token", "Authorization"])
      assert.equal(new Headers(init?.headers).has(name), false)
    return json(profile)
  }) as typeof fetch)
  assert.deepEqual(await api.readSelf(actor), profile)
  for (const id of [0, -1, 2147483648, NaN])
    await assert.rejects(api.readSelf({ ...actor, employee_id: id }), /INVALID_RESPONSE/)
  assert.equal(count, 1)
})
test("DTO exact/null và user khác employee; không nhận người khác hoặc secret", () => {
  assert.deepEqual(parseSelfProfile(profile, actor), profile)
  const populated = {
    ...profile,
    address: "Địa chỉ tổng hợp",
    phone: "0123456789",
    department: { id: 3, name: "Phòng A" },
    job_level: { id: 5, name: "Cấp A" },
  }
  assert.deepEqual(parseSelfProfile(populated, actor), populated)
  const missing = { ...profile } as Record<string, unknown>
  delete missing.address
  for (const value of [
    null,
    [],
    missing,
    { ...profile, user_id: 25 },
    { ...profile, employee_id: 12 },
    { ...profile, password_hash: "marker" },
    { ...profile, address: 9 },
    { ...profile, company_email: "@@@" },
    { ...profile, company_email: "Tên <a@example.invalid>" },
    { ...profile, position: null },
    { ...profile, position: { id: 0, name: "A" } },
    { ...profile, department: { id: 3, name: "A", raw: true } },
    { ...profile, employment_status: "INACTIVE" },
    { ...profile, fullname: "x".repeat(101) },
    { ...profile, phone: "1".repeat(21) },
    { ...profile, address: "\ud800" },
    { ...profile, fullname: "a\0b" },
    { ...profile, company_email: "bad\r\n@example.invalid" },
  ])
    assert.throws(() => parseSelfProfile(value, actor), /INVALID_RESPONSE/)
})
test("HTTP/DB/mạng giữ phân loại, không message thô hoặc auto retry", async () => {
  for (const [status, code] of [
    [401, "UNAUTHENTICATED"],
    [403, "PASSWORD_CHANGE_REQUIRED"],
    [404, "NOT_FOUND"],
    [503, "SERVICE_UNAVAILABLE"],
    [500, "INTERNAL_ERROR"],
  ] as const) {
    let calls = 0
    const api = createProfileApi((async () => {
      calls++
      return json({ code, message: "private-marker" }, status)
    }) as typeof fetch)
    await assert.rejects(
      api.readSelf(actor),
      (error: unknown) =>
        error instanceof AuthError &&
        error.status === status &&
        error.code === code &&
        !error.message.includes("private"),
    )
    assert.equal(calls, 1)
    assert.ok(profileErrorMessage(new AuthError(status, code)).length)
  }
  const api = createProfileApi((async () => {
    throw new Error("private-url")
  }) as typeof fetch)
  await assert.rejects(api.readSelf(actor), /NETWORK_ERROR/)
  for (const response of [
    new Response("proxy"),
    json(profile, 201),
    json({ ...profile, user_id: 99 }),
  ])
    await assert.rejects(
      createProfileApi((async () => response) as typeof fetch).readSelf(actor),
      /INVALID_RESPONSE/,
    )
})
