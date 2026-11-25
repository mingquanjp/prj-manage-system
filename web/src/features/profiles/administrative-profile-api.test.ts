import test from "node:test"
import assert from "node:assert/strict"
import {
  administrativeValues,
  administrativePatch,
  administrativeErrors,
  createAdministrativeProfileApi,
  parseAdministrativeProfile,
  parseAdministrativePage,
  type AdministrativeProfile,
} from "./administrative-profile-api.ts"
import { parseAdministrativeLocation } from "../session/server-navigation.ts"
const profile: AdministrativeProfile = {
  user_id: 101,
  employee_id: 1001,
  employee_code: "EMP000001",
  fullname: "Nhân viên",
  company_email: "employee@example.invalid",
  phone: null,
  address: null,
  position: { id: 1, name: "Vị trí" },
  department: null,
  job_level: null,
  employment_status: "ACTIVE",
}
test("HR DTO accepts inactive/raw legacy and excludes extra contact from list", () => {
  const legacy = {
    ...profile,
    employment_status: "INACTIVE",
    phone: "x".repeat(40),
    address: " ".repeat(3000),
  }
  assert.equal(parseAdministrativeProfile(legacy, 1001).phone?.length, 40)
  assert.throws(() => parseAdministrativeProfile({ ...legacy, password: "secret" }, 1001))
  assert.throws(() => parseAdministrativeProfile(legacy, 101))
  assert.throws(() =>
    parseAdministrativePage(
      { items: [profile], page: 1, page_size: 20, total: 1 },
      { q: "", page: 1, pageSize: 20 },
    ),
  )
})
test("patch raw before/null/equal keys/unchanged legacy and canonical ASCII email", () => {
  const before = administrativeValues({ ...profile, phone: " 0900 ", address: "x".repeat(3000) })
  assert.deepEqual(administrativePatch(before, before).changes, {})
  const patch = administrativePatch(before, {
    ...before,
    phone: null,
    fullname: "\u0085Tên mới\u0085",
  })
  assert.deepEqual(patch, {
    expected_before_values: { fullname: "Nhân viên", phone: " 0900 " },
    changes: { fullname: "Tên mới", phone: null },
  })
  assert.deepEqual(administrativeErrors(patch), {})
  assert.deepEqual(
    administrativePatch(before, { ...before, company_email: " EMPLOYEE@EXAMPLE.INVALID " }).changes,
    {},
  )
  assert.equal(
    administrativePatch(before, { ...before, company_email: "İ@EXAMPLE.INVALID" }).changes
      .company_email,
    "İ@EXAMPLE.INVALID",
  )
  assert.ok(administrativeErrors(administrativePatch(before, { ...before, address: "  " })).address)
})
test("Unicode/codepoint/UTF8 boundaries and FEFF Go trim parity", () => {
  const before = administrativeValues(profile)
  for (const [value, valid] of [
    ["😀".repeat(100), true],
    ["😀".repeat(101), false],
    ["\uFEFF", true],
    ["\ud800", false],
  ] as const)
    assert.equal(
      !administrativeErrors(administrativePatch(before, { ...before, fullname: value })).fullname,
      valid,
    )
  const patch = { expected_before_values: { address: "" }, changes: { address: "a" } }
  const size = new TextEncoder().encode(JSON.stringify(patch)).length
  patch.expected_before_values.address = "b".repeat(65536 - size)
  assert.deepEqual(administrativeErrors(patch), {})
  patch.expected_before_values.address += "b"
  assert.ok(administrativeErrors(patch).address)
})
test("PATCH exact URL/body/CSRF no retries after lost response", async () => {
  let calls = 0
  const patch = {
    expected_before_values: { fullname: "Nhân viên" },
    changes: { fullname: "Tên mới" },
  }
  const api = createAdministrativeProfileApi(async (url, init) => {
    calls++
    assert.equal(url, "/api/employees/1001")
    assert.equal(init?.method, "PATCH")
    assert.equal((init?.headers as Record<string, string>)["X-CSRF-Token"], "csrf")
    assert.deepEqual(JSON.parse(String(init?.body)), patch)
    assert.equal(init?.credentials, "same-origin")
    assert.equal(init?.cache, "no-store")
    throw new Error("lost")
  })
  await assert.rejects(api.edit(1001, patch, "csrf"))
  assert.equal(calls, 1)
})
test("HR location validates query and IDs", () => {
  assert.equal(parseAdministrativeLocation("/personnel/employees/1001", "")?.id, 1001)
  for (const [path, search] of [
    ["/personnel/employees/1001", "?role=HR"],
    ["/personnel/employees/0", ""],
    ["/personnel/employees", "?department=1"],
    ["/personnel/employees", "?page=1&page=2"],
  ])
    assert.equal(parseAdministrativeLocation(path, search), null)
})
test("shared catalog GET carries caller abort without changing existing contract", async () => {
  const { createCatalogApi } = await import("../organization/catalog-api.ts")
  let signal: AbortSignal | null | undefined
  const caller = new AbortController()
  const api = createCatalogApi(async (_url, init) => {
    signal = init?.signal
    return new Response(
      JSON.stringify({ items: [], pagination: { page: 1, page_size: 100, total: 0 } }),
      { status: 200, headers: { "Content-Type": "application/json" } },
    )
  })
  await api.list("positions", "", 1, 100, caller.signal)
  assert.equal(signal?.aborted, false)
  caller.abort()
  assert.equal(signal?.aborted, true)
})
