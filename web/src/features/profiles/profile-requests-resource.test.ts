import { test } from "node:test"
import assert from "node:assert/strict"
import { AuthError } from "../session/auth-api.ts"
import { createProfileRequestsResource } from "./profile-requests-resource.ts"
import type { ProfileRequestsApi, RequestDetail, RequestPage } from "./profile-requests-api.ts"

const detail: RequestDetail = {
  request_id: 9,
  employee_id: 2,
  employee_code: "EMP000002",
  fullname: "Nhân viên",
  status: "PENDING",
  requested_at: "2026-11-24T02:00:00Z",
  resolved_at: null,
  requested_by: 7,
  resolved_by: null,
  decision_reason: null,
  explanation: null,
  before_values: { phone: null },
  proposed_values: { phone: "123" },
  current_values: { phone: null },
}
const page: RequestPage = { items: [], total: 0, page: 1, page_size: 20 }
function deferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((done) => {
    resolve = done
  })
  return { promise, resolve }
}
function api(overrides: Partial<ProfileRequestsApi> = {}): ProfileRequestsApi {
  return {
    list: async () => page,
    read: async () => detail,
    submit: async () => detail,
    ...overrides,
  }
}
const noop = () => {}

test("clear discards late list/detail responses after leaving session", async () => {
  const list = deferred<RequestPage>(),
    read = deferred<RequestDetail>()
  const resource = createProfileRequestsResource(
    api({ list: () => list.promise, read: () => read.promise }),
  )
  const loading = resource.loadList(2, 1, noop),
    reading = resource.loadDetail(2, 9, noop)
  resource.clear()
  list.resolve(page)
  read.resolve(detail)
  await Promise.all([loading, reading])
  assert.equal(resource.getSnapshot().list.data, undefined)
  assert.equal(resource.getSnapshot().detail.data, undefined)
})
test("newer page wins even if earlier GET ignores abort", async () => {
  const first = deferred<RequestPage>()
  const resource = createProfileRequestsResource(
    api({ list: async (_, p) => (p === 1 ? first.promise : { ...page, page: 2 }) }),
  )
  const loading = resource.loadList(2, 1, noop)
  await resource.loadList(2, 2, noop)
  first.resolve(page)
  await loading
  assert.equal(resource.getSnapshot().list.data?.page, 2)
})
test("uncertain POST invalidates old history and requires fresh GET, never replays", async () => {
  let posts = 0
  const resource = createProfileRequestsResource(api())
  await resource.loadList(2, 1, noop)
  const attempt = () =>
    resource.submit(
      (post) => post("csrf"),
      async () => {
        posts++
        throw new AuthError(503, "SERVICE_UNAVAILABLE")
      },
      noop,
    )
  await assert.rejects(attempt(), /Chưa xác định/)
  assert.equal(resource.getSnapshot().submission, "uncertain")
  assert.equal(resource.getSnapshot().list.data, undefined)
  assert.equal(resource.allowNewDraft(), false)
  await assert.rejects(attempt())
  assert.equal(posts, 1)
  await resource.loadList(2, 1, noop)
  assert.equal(resource.getSnapshot().submission, "uncertain")
  assert.equal(resource.allowNewDraft(), true)
})
test("CSRF failure before callback is ordinary error, no POST", async () => {
  let posts = 0
  const resource = createProfileRequestsResource(api())
  await assert.rejects(
    resource.submit(
      async () => {
        throw new AuthError(503, "SERVICE_UNAVAILABLE")
      },
      async () => {
        posts++
        return detail
      },
      noop,
    ),
  )
  assert.equal(posts, 0)
  assert.equal(resource.getSnapshot().submission, "error")
})
test("401/403 clear PII and request recheck; 503 preserves explicit read error", async () => {
  for (const status of [401, 403, 503]) {
    let checks = 0
    const resource = createProfileRequestsResource(
      api({
        list: async () => {
          throw new AuthError(status, "HTTP_ERROR")
        },
      }),
    )
    await resource.loadDetail(2, 9, noop)
    await resource.loadList(2, 1, () => {
      checks++
    })
    assert.equal(checks, status === 503 ? 0 : 1)
    assert.equal(resource.getSnapshot().detail.data, status === 503 ? detail : undefined)
    assert.equal(resource.getSnapshot().list.phase, status === 503 ? "error" : "loading")
  }
})
test("late submit after clear cannot restore success or PII; double click blocked", async () => {
  const pending = deferred<RequestDetail>()
  const resource = createProfileRequestsResource(api())
  const submitting = resource.submit(
    (post) => post("csrf"),
    () => pending.promise,
    noop,
  )
  await assert.rejects(
    resource.submit(
      (post) => post("csrf"),
      async () => detail,
      noop,
    ),
  )
  resource.clear()
  pending.resolve(detail)
  assert.equal(await submitting, undefined)
  assert.equal(resource.getSnapshot().submission, "idle")
})
test("stale blocks further POST, successful submit stays successful if history reload fails", async () => {
  const resource = createProfileRequestsResource(
    api({
      list: async () => {
        throw new AuthError(503, "SERVICE_UNAVAILABLE")
      },
    }),
  )
  await assert.rejects(
    resource.submit(
      (post) => post("csrf"),
      async () => {
        throw new AuthError(409, "STALE_DATA")
      },
      noop,
    ),
  )
  assert.equal(resource.getSnapshot().submission, "stale")
  resource.allowNewDraft()
  await resource.submit(
    (post) => post("csrf"),
    async () => detail,
    noop,
  )
  await resource.loadList(2, 1, noop)
  assert.equal(resource.getSnapshot().submission, "success")
  assert.equal(resource.getSnapshot().list.phase, "error")
})
