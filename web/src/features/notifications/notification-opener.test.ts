import test from "node:test"
import assert from "node:assert/strict"
import { AuthError, type Principal } from "../session/auth-api.ts"
import type { SessionState } from "../session/auth-session.ts"
import type { RequestDetail } from "../profiles/profile-requests-api.ts"
import type { ServerNotification } from "./server-notifications-api.ts"
import { createNotificationOpener } from "./notification-opener.ts"

const notification: ServerNotification = {
  notification_id: 7,
  notification_type: "PROFILE_CHANGE_APPROVED",
  resource_type: "PROFILE_CHANGE_REQUEST",
  resource_id: 12,
  message: "Đề nghị đã được chấp thuận.",
  created_at: "2026-11-25T01:00:00Z",
  read_at: null,
}
function deferred() {
  let resolve!: (value: RequestDetail) => void
  let reject!: (error: unknown) => void
  const promise = new Promise<RequestDetail>((done, fail) => {
    resolve = done
    reject = fail
  })
  return { promise, resolve: () => resolve({} as RequestDetail), reject }
}
function setup() {
  let principal: Principal | null = {
    user_id: 1,
    employee_id: 101,
    employee_code: "EMP000001",
    fullname: "Nhân viên thử",
    system_roles: ["HR"],
    must_change_password: false,
  }
  const calls: string[] = []
  const pending: ReturnType<typeof deferred>[] = []
  const signals: AbortSignal[] = []
  let lastMessage = ""
  const read = (label: string, id: number, signal?: AbortSignal) => {
    calls.push(`${label}:${id}`)
    if (signal) signals.push(signal)
    const next = deferred()
    pending.push(next)
    return next.promise
  }
  const opener = createNotificationOpener({
    session: {
      getSnapshot: () =>
        ({
          phase: principal ? "authenticated" : "signed-out",
          principal,
          busy: false,
        }) as SessionState,
      refresh: async () => {
        calls.push("refresh")
      },
    },
    selfApi: { read: (employeeId, id, signal) => read(`self:${employeeId}`, id, signal) },
    reviewApi: { detail: (id, signal) => read("review", id, signal) },
    onState: (state) => {
      lastMessage = state.message
    },
    onOpen: (destination) => {
      calls.push(`open:${destination}`)
    },
    markRead: async (id) => {
      calls.push(`read:${id}`)
      return true
    },
  })
  return {
    opener,
    calls,
    pending,
    signals,
    message: () => lastMessage,
    setPrincipal: (value: Principal | null) => {
      principal = value
    },
    principal: () => principal!,
  }
}

test("đợi GET self thành công rồi mở đúng query và mới read; thông báo terminal vẫn mở được", async () => {
  const f = setup()
  const task = f.opener.open(notification)
  assert.deepEqual(f.calls, ["self:101:12"])
  f.pending[0].resolve()
  await task
  assert.deepEqual(f.calls, ["self:101:12", "open:/profile?request_id=12", "read:7"])
})

test("SUBMITTED kiểm HR detail; đã đọc không POST read lại", async () => {
  const f = setup()
  const task = f.opener.open({
    ...notification,
    notification_type: "PROFILE_CHANGE_SUBMITTED",
    read_at: notification.created_at,
  })
  assert.deepEqual(f.calls, ["review:12"])
  f.pending[0].resolve()
  await task
  assert.deepEqual(f.calls, ["review:12", "open:/personnel/profile-requests/12"])
})

test("GET401/403/404/503/mạng không mở hoặc đánh dấu đọc, chỉ401/403 kiểm lại phiên", async () => {
  for (const status of [401, 403, 404, 503, 0]) {
    const f = setup()
    const task = f.opener.open(notification)
    f.pending[0].reject(new AuthError(status, "TEST_ERROR"))
    await task
    assert.deepEqual(
      f.calls,
      status === 401 || status === 403 ? ["self:101:12", "refresh"] : ["self:101:12"],
    )
    assert.notEqual(f.message(), "")
  }
})

test("đóng khung hoặc đổi phiên khi GET chờ không navigate/read dù provider bỏ qua abort", async () => {
  for (const action of ["close", "logout", "other-user", "lost-role", "limited"] as const) {
    const f = setup()
    const task = f.opener.open(notification)
    if (action === "close") f.opener.cancel()
    if (action === "logout") f.setPrincipal(null)
    if (action === "other-user") f.setPrincipal({ ...f.principal(), user_id: 2, employee_id: 202 })
    if (action === "lost-role") f.setPrincipal({ ...f.principal(), system_roles: [] })
    if (action === "limited") f.setPrincipal({ ...f.principal(), must_change_password: true })
    f.pending[0].resolve()
    await task
    assert.deepEqual(f.calls, ["self:101:12"], action)
    if (action === "close") assert.equal(f.signals[0].aborted, true)
  }
})

test("GET mới vô hiệu kết quả thành công/lỗi cũ và chỉ tài nguyên cuối được mở/read", async () => {
  for (const firstFails of [false, true]) {
    const f = setup()
    const first = f.opener.open(notification)
    const second = f.opener.open({ ...notification, notification_id: 8, resource_id: 13 })
    assert.equal(f.signals[0].aborted, true)
    if (firstFails) f.pending[0].reject(new AuthError(403, "FORBIDDEN"))
    else f.pending[0].resolve()
    await first
    assert.equal(f.message(), "")
    f.pending[1].resolve()
    await second
    assert.deepEqual(f.calls, [
      "self:101:12",
      "self:101:13",
      "open:/profile?request_id=13",
      "read:8",
    ])
  }
})
