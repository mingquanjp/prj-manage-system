import { test } from "node:test"
import assert from "node:assert/strict"
import {
  createAccountApi,
  accountDraftErrors,
  accountPayload,
  emptyAccountDraft,
  accountErrorMessage,
  mailMessages,
} from "./account-api.ts"
import { loadAccountOptions } from "./account-options.ts"
import { createCatalogApi } from "../organization/catalog-api.ts"
import { AuthError } from "../session/auth-api.ts"
const account = {
  user_id: 1,
  employee_id: 22,
  employee_code: "EMP000008",
  fullname: "Người thử",
  company_email: "test@example.invalid",
  status: "ACTIVE",
  system_roles: ["ADMIN", "HR"],
  must_change_password: true,
}
const draft = {
  ...emptyAccountDraft,
  fullname: "  Người thử  ",
  companyEmail: "test@example.invalid",
  positionId: "4",
  address: "Địa chỉ thử",
}
const json = (v: unknown, status = 200) =>
  new Response(JSON.stringify(v), { status, headers: { "Content-Type": "application/json" } })
test("create allowlist, cookie/CSRF/no-store; ba kết quả mail giữ account đã commit và mã thật", async () => {
  for (const status of ["accepted", "failed", "unknown"] as const) {
    let calls = 0
    const api = createAccountApi((async (url, init) => {
      calls++
      assert.equal(url, "/api/employees")
      assert.equal(init?.method, "POST")
      assert.equal(init?.credentials, "same-origin")
      assert.equal(init?.cache, "no-store")
      assert.equal(init?.redirect, "error")
      const headers = new Headers(init?.headers)
      assert.equal(headers.get("X-CSRF-Token"), "csrf")
      assert.equal(headers.has("Cookie"), false)
      assert.equal(headers.has("Origin"), false)
      assert.deepEqual(JSON.parse(String(init?.body)), {
        fullname: "  Người thử  ",
        company_email: "test@example.invalid",
        position_id: 4,
        department_id: null,
        job_level_id: null,
        phone: null,
        address: "Địa chỉ thử",
        system_roles: [],
      })
      return json({ account_created: true, account, mail: { status } }, 201)
    }) as typeof fetch)
    const result = await api.create(
      { ...draft, actor_id: 999, password: "must-not-send" } as typeof draft,
      "csrf",
    )
    assert.equal(result.account.employee_code, "EMP000008")
    assert.equal(result.mail.status, status)
    assert.equal(calls, 1)
    assert.ok(mailMessages[status])
  }
})
test("preview/list dùng server, query literal, page rỗng giữ total; preview không POST", async () => {
  const seen: string[] = []
  const api = createAccountApi((async (url, init) => {
    seen.push(String(url))
    assert.equal(init?.method, undefined)
    return String(url).includes("next-code")
      ? json({ employee_code: "EMP000007", is_preview: true })
      : json({ items: [], pagination: { page: 2, page_size: 20, total: 3 } })
  }) as typeof fetch)
  assert.equal(await api.preview(), "EMP000007")
  assert.equal((await api.list("%_\\ có dấu", 2, 20)).pagination.total, 3)
  assert.equal(new URL(seen[1], "http://local").searchParams.get("q"), "%_\\ có dấu")
})
test("DTO strict không nhận credential/PII/delivery status; lỗi 201 không được coi là success", async () => {
  for (const body of [
    {
      account_created: true,
      account: { ...account, password: "secret" },
      mail: { status: "accepted" },
    },
    {
      account_created: true,
      account: { ...account, address: "private" },
      mail: { status: "accepted" },
    },
    { account_created: true, account, mail: { status: "delivered" } },
    { account_created: false, account, mail: { status: "failed" } },
    {
      account_created: true,
      account: { ...account, system_roles: ["ADMIN", "ADMIN"] },
      mail: { status: "unknown" },
    },
  ])
    await assert.rejects(
      createAccountApi((async () => json(body, 201)) as typeof fetch).create(draft, "token"),
      (e: unknown) => e instanceof AuthError && e.code === "INVALID_RESPONSE",
    )
  await assert.rejects(
    createAccountApi((async () =>
      json({ employee_code: "EMP000001", is_preview: false })) as typeof fetch).preview(),
  )
  await assert.rejects(
    createAccountApi((async () =>
      json({
        items: [account, account],
        pagination: { page: 1, page_size: 20, total: 2 },
      })) as typeof fetch).list("", 1, 20),
  )
})
test("HTTP/network không retry, không đưa message provider vào lỗi; mutation không rõ phải đối soát", async () => {
  for (const status of [400, 401, 403, 409, 413, 415, 500, 503, 0]) {
    let calls = 0
    const api = createAccountApi((async () => {
      calls++
      if (!status) throw new Error("private")
      return json(
        { code: status === 409 ? "EMAIL_CONFLICT" : "TEST_ERROR", message: "private" },
        status,
      )
    }) as typeof fetch)
    await assert.rejects(api.create(draft, "csrf"), (e: unknown) => {
      assert.ok(e instanceof AuthError)
      assert.equal(e.status, status)
      assert.ok(!e.message.includes("private"))
      if (status === 0 || status >= 500) assert.match(accountErrorMessage(e, true), /đối soát/)
      return true
    })
    assert.equal(calls, 1)
  }
})
test("validation code point, Unicode/space/null, role và ID; payload không lấy field ngoài", () => {
  assert.deepEqual(accountDraftErrors(draft), {})
  assert.equal(accountDraftErrors({ ...draft, fullname: "😀".repeat(100) }).fullname, undefined)
  assert.ok(accountDraftErrors({ ...draft, fullname: "😀".repeat(101) }).fullname)
  for (const value of ["  ", "\uD800", "a\0"])
    assert.ok(accountDraftErrors({ ...draft, fullname: value }).fullname)
  assert.ok(accountDraftErrors({ ...draft, address: " " }).address)
  assert.ok(accountDraftErrors({ ...draft, phone: "1".repeat(21) }).phone)
  assert.ok(accountDraftErrors({ ...draft, positionId: "2147483648" }).positionId)
  assert.ok(accountDraftErrors({ ...draft, companyEmail: "invalid-email" }).companyEmail)
  assert.equal(
    accountDraftErrors({ ...draft, companyEmail: '"local name"@example.invalid' }).companyEmail,
    undefined,
  )
  assert.ok(accountDraftErrors({ ...draft, systemRoles: ["HR", "HR"] }).systemRoles)
  assert.equal(Object.hasOwn(accountPayload(draft), "actor_id"), false)
})
test("catalog choices tải đủ >100 mục, từ chối tổng thay đổi hoặc ID trùng giữa trang", async () => {
  const api = createCatalogApi((async (url) => {
    const p = new URL(String(url), "http://local").searchParams
    const page = Number(p.get("page"))
    return json({
      items:
        page === 1
          ? Array.from({ length: 100 }, (_, i) => ({ id: i + 1, name: `Mục ${i + 1}` }))
          : [{ id: 101, name: "Mục 101" }],
      pagination: { page, page_size: 100, total: 101 },
    })
  }) as typeof fetch)
  const result = await loadAccountOptions(api)
  assert.equal(result.positions.length, 101)
  assert.equal(result.positions[100].id, "101")
  const duplicate = createCatalogApi((async (url) => {
    const page = Number(new URL(String(url), "http://local").searchParams.get("page"))
    return json({
      items: [{ id: 1, name: "Trùng" }],
      pagination: { page, page_size: 100, total: 2 },
    })
  }) as typeof fetch)
  await assert.rejects(loadAccountOptions(duplicate))
})
