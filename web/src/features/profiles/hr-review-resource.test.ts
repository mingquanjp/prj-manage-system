import test from "node:test"
import assert from "node:assert/strict"
import { createAuthSession } from "../session/auth-session.ts"
import { AuthError, type AuthApi, type Principal } from "../session/auth-api.ts"
import { createHrReviewResource } from "./hr-review-resource.ts"
import type { HrReviewApi, ReviewDetail } from "./hr-review-api.ts"
import { reviewPendingFixture } from "./hr-review-fixture.test-helper.ts"
const pending = reviewPendingFixture as ReviewDetail
async function setup(override: Partial<HrReviewApi> = {}) {
  let actor = {
    user_id: 3,
    employee_id: 8,
    must_change_password: false,
    system_roles: ["HR"],
  } as Principal
  const session = createAuthSession({
    me: async () => actor,
    csrf: async () => "token",
  } as unknown as AuthApi)
  await session.check()
  const api: HrReviewApi = {
    list: async () => ({ items: [], total: 0, page: 1, page_size: 20 }),
    detail: async () => pending,
    decide: async () => ({
      ...pending,
      status: "APPROVED",
      resolved_by: 3,
      resolved_at: "2026-11-24T01:00:00Z",
    }),
    ...override,
  }
  const resource = createHrReviewResource(api, session)
  return {
    session,
    resource,
    changeActor: (next: Principal) => {
      actor = next
    },
  }
}
test("unknown decision blocks replay until GET; resolved result is never optimistic", async () => {
  let calls = 0
  const { resource } = await setup({
    decide: async () => {
      calls++
      throw new AuthError(503, "COMMIT_UNKNOWN")
    },
  })
  await resource.load(1)
  await resource.decide("approve", "")
  assert.equal(resource.getSnapshot().detail?.status, "PENDING")
  assert.equal(resource.getSnapshot().mustReload, true)
  assert.match(resource.getSnapshot().message, /Chưa xác định/)
  await resource.decide("approve", "")
  assert.equal(calls, 1)
  await resource.load(1)
  assert.equal(resource.getSnapshot().mustReload, false)
})
test("stale and resolved conflict require reload; inactive still allows reject", async () => {
  for (const code of ["STALE_DATA", "REQUEST_ALREADY_RESOLVED", "EMPLOYEE_INACTIVE"]) {
    const { resource } = await setup({
      decide: async () => {
        throw new AuthError(409, code)
      },
    })
    await resource.load(1)
    await resource.decide("approve", "")
    assert.equal(resource.getSnapshot().mustReload, code !== "EMPLOYEE_INACTIVE")
  }
})
test("late GET cannot restore PII after clear and forbidden response clears data", async () => {
  let resolve!: (value: ReviewDetail) => void
  const { resource } = await setup({
    detail: () =>
      new Promise((done) => {
        resolve = done
      }),
  })
  const loading = resource.load(1)
  resource.clear()
  resolve(pending)
  await loading
  assert.equal(resource.getSnapshot().detail, undefined)
  const denied = await setup({
    decide: async () => {
      throw new AuthError(403, "FORBIDDEN")
    },
  })
  await denied.resource.load(1)
  await denied.resource.decide("approve", "")
  assert.equal(denied.resource.getSnapshot().detail, undefined)
})
test("double click issues one mutation, terminal cannot be decided again", async () => {
  let finish!: (value: ReviewDetail) => void
  let calls = 0
  const { resource } = await setup({
    decide: () => {
      calls++
      return new Promise((done) => {
        finish = done
      })
    },
  })
  await resource.load(1)
  const operation = resource.decide("approve", "")
  await resource.decide("approve", "")
  await Promise.resolve()
  finish({ ...pending, status: "APPROVED", resolved_by: 3, resolved_at: "2026-11-24T01:00:00Z" })
  await operation
  await resource.decide("reject", "reason")
  assert.equal(calls, 1)
  assert.equal(resource.getSnapshot().detail?.status, "APPROVED")
})
