import { test } from "node:test"
import assert from "node:assert/strict"
import { AuthError } from "../session/auth-api.ts"
import { type SelfProfile } from "./profile-api.ts"
import { createProfileResource } from "./profile-resource.ts"
const actor = { user_id: 12, employee_id: 25 }
const profile: SelfProfile = {
  ...actor,
  employee_code: "E025",
  fullname: "Tổng hợp",
  company_email: "test@example.invalid",
  phone: null,
  address: null,
  department: null,
  job_level: null,
  position: { id: 3, name: "Vị trí" },
  employment_status: "ACTIVE",
}
function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason: unknown) => void
  const promise = new Promise<T>((yes, no) => {
    resolve = yes
    reject = no
  })
  return { promise, resolve, reject }
}
test("response cũ không ghi đè tải mới; clear xóa PII ngay", async () => {
  const old = deferred<SelfProfile>(),
    next = deferred<SelfProfile>()
  const signals: AbortSignal[] = []
  let calls = 0
  const resource = createProfileResource({
    readSelf: async (_, signal) => {
      signals.push(signal!)
      return ++calls === 1 ? old.promise : next.promise
    },
  })
  const first = resource.load(actor, () => assert.fail())
  const second = resource.load(actor, () => assert.fail())
  assert.equal(signals[0].aborted, true)
  next.resolve(profile)
  await second
  assert.equal(resource.getSnapshot().profile, profile)
  old.resolve({ ...profile, fullname: "Response cũ" })
  await first
  assert.equal(resource.getSnapshot().profile, profile)
  resource.clear()
  assert.equal(resource.getSnapshot().profile, undefined)
})
test("rời màn/logout/đổi user chặn success/lỗi muộn và refresh phiên mới", async () => {
  for (const error of [false, true]) {
    const pending = deferred<SelfProfile>()
    const resource = createProfileResource({ readSelf: () => pending.promise })
    const loading = resource.load(actor, () => assert.fail("không kiểm phiên mới"))
    resource.clear()
    if (error) pending.reject(new AuthError(401, "UNAUTHENTICATED"))
    else pending.resolve(profile)
    await loading
    assert.equal(resource.getSnapshot().profile, undefined)
    assert.equal(resource.getSnapshot().phase, "loading")
  }
})
test("lỗi xóa PII; 401/403 recheck một lần, DB/mạng không giả hết phiên", async () => {
  for (const status of [401, 403, 404, 500, 503, 0]) {
    let failed = false,
      rechecks = 0,
      calls = 0
    const resource = createProfileResource({
      readSelf: async () => {
        calls++
        if (failed) throw new AuthError(status, "ERROR")
        return profile
      },
    })
    await resource.load(actor, () => rechecks++)
    assert.equal(resource.getSnapshot().profile, profile)
    failed = true
    await resource.load(actor, () => rechecks++)
    assert.equal(resource.getSnapshot().phase, "error")
    assert.equal(resource.getSnapshot().profile, undefined)
    assert.equal(rechecks, status === 401 || status === 403 ? 1 : 0)
    assert.equal(calls, 2)
  }
})
