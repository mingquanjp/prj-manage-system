import test from "node:test"
import assert from "node:assert/strict"
import {
  createHrReviewApi,
  parseReviewDetail,
  parseReviewPage,
  validReviewReason,
} from "./hr-review-api.ts"
import {
  canReviewServerProfiles,
  parseReviewLocation,
  serverHome,
} from "../session/server-navigation.ts"
import type { Principal } from "../session/auth-api.ts"

import { reviewPendingFixture as pending } from "./hr-review-fixture.test-helper.ts"
const response = (value: unknown) => Response.json(value)
test("review DTO allowlist, raw NULL/empty and terminal invariants", () => {
  assert.equal(parseReviewDetail(pending, 1).before_values.address, null)
  assert.equal(
    parseReviewDetail({ ...pending, before_values: { address: "" } }, 1).before_values.address,
    "",
  )
  for (const value of [
    { ...pending, password: "forbidden" },
    { ...pending, request_id: 2 },
    { ...pending, current_values: { phone: null } },
    { ...pending, resolved_by: 2 },
    { ...pending, status: "APPROVED" },
    { ...pending, proposed_values: { address: null } },
    {
      ...pending,
      status: "REJECTED",
      resolved_by: 1,
      resolved_at: "2026-11-24T01:00:00Z",
      decision_reason: null,
    },
  ])
    assert.throws(() => parseReviewDetail(value, 1))
})
test("review adapter pins scope and sends only decision DTO with CSRF, no replay", async () => {
  const calls: { url: string; init?: RequestInit }[] = []
  const api = createHrReviewApi(async (url, init) => {
    calls.push({ url: String(url), init })
    if (init?.method === "GET") return response({ items: [], page: 1, page_size: 20, total: 0 })
    const reject = String(url).endsWith("reject")
    return response({
      ...pending,
      status: reject ? "REJECTED" : "APPROVED",
      resolved_by: 3,
      resolved_at: "2026-11-24T01:00:00Z",
      decision_reason: reject ? "Lý do" : null,
    })
  })
  await api.list({ status: "PENDING", page: 1, pageSize: 20 })
  assert.match(calls[0].url, /scope=review/)
  await api.decide(1, "approve", "ignored", "csrf-test")
  assert.equal(calls[1].init?.body, "{}")
  assert.equal(new Headers(calls[1].init?.headers).get("X-CSRF-Token"), "csrf-test")
  assert.equal(calls[1].init?.credentials, "same-origin")
  assert.equal(calls[1].init?.redirect, "error")
  await api.decide(1, "reject", "  Lý do  ", "csrf-test")
  assert.deepEqual(JSON.parse(String(calls[2].init?.body)), { reason: "Lý do" })
  let count = 0
  const failed = createHrReviewApi(async () => {
    count++
    throw Error("lost")
  })
  await assert.rejects(failed.decide(1, "approve", "", "csrf-test"))
  assert.equal(count, 1)
})
test("invalid responses and error statuses are not successful decisions", async () => {
  for (const status of [401, 403, 404, 409, 503]) {
    const api = createHrReviewApi(async () => Response.json({ code: "TEST_ERROR" }, { status }))
    await assert.rejects(
      api.decide(1, "approve", "", "token"),
      (error: unknown) => (error as { status: number }).status === status,
    )
  }
  const wrong = createHrReviewApi(async () => response(pending))
  await assert.rejects(wrong.decide(1, "approve", "", "token"))
  const query = { page: 1, pageSize: 20 }
  assert.throws(() => parseReviewPage({ items: [], page: 2, page_size: 20, total: 0 }, query))
})
test("Unicode reason and strict HR routes preserve current login mapping", () => {
  assert.equal(validReviewReason("😀".repeat(2000)), true)
  assert.equal(validReviewReason("😀".repeat(2001)), false)
  assert.equal(validReviewReason("  \n "), false)
  assert.equal(validReviewReason("\uD800"), false)
  assert.equal(validReviewReason("\u0085"), false)
  assert.equal(validReviewReason("\u0085" + "a".repeat(2000) + "\u0085"), true)
  assert.equal(validReviewReason("\uFEFF"), true)
  const root = "/personnel/profile-requests"
  assert.deepEqual(parseReviewLocation(root, ""), {
    query: { status: "PENDING", page: 1, pageSize: 20 },
  })
  assert.equal(parseReviewLocation(`${root}/123`, "")?.id, 123)
  for (const query of [
    "?scope=self",
    "?status=bad",
    "?page=1&page=2",
    "?page=0",
    "?page=2147483648",
    "?page_size=101",
  ])
    assert.equal(parseReviewLocation(root, query), null)
  assert.equal(parseReviewLocation(`${root}/1`, "?page=1"), null)
  const actor = { must_change_password: false, system_roles: ["HR", "ADMIN"] } as Principal
  assert.equal(canReviewServerProfiles(actor), true)
  assert.equal(serverHome(actor), "/organization/departments")
  assert.equal(canReviewServerProfiles({ ...actor, system_roles: ["ADMIN"] }), false)
  assert.equal(canReviewServerProfiles({ ...actor, must_change_password: true }), false)
})
