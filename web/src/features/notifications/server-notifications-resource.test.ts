import test from "node:test"
import assert from "node:assert/strict"
import { AuthError, type AuthApi, type Principal } from "../session/auth-api.ts"
import { createAuthSession } from "../session/auth-session.ts"
import { createServerNotificationsResource } from "./server-notifications-resource.ts"
import type { NotificationPage, ServerNotificationsApi } from "./server-notifications-api.ts"

const page: NotificationPage = {
  items: [
    {
      notification_id: 5,
      notification_type: "PROFILE_CHANGE_APPROVED",
      resource_type: "PROFILE_CHANGE_REQUEST",
      resource_id: 12,
      message: "Đề nghị đã được chấp thuận.",
      created_at: "2026-11-25T01:00:00Z",
      read_at: null,
    },
  ],
  total: 1,
  page: 1,
  page_size: 20,
  unread_count: 1,
}
async function setup(
  override: Partial<ServerNotificationsApi> = {},
  csrf: () => Promise<string> = async () => "token",
) {
  let actor = {
    user_id: 1,
    employee_id: 8,
    fullname: "Nhân viên thử",
    employee_code: "EMP000008",
    system_roles: [],
    must_change_password: false,
  } as Principal
  const session = createAuthSession({ me: async () => actor, csrf } as unknown as AuthApi)
  await session.check()
  const api: ServerNotificationsApi = { list: async () => page, read: async () => {}, ...override }
  const resource = createServerNotificationsResource(api, session)
  return {
    session,
    resource,
    setActor: (next: Principal) => {
      actor = next
    },
  }
}
test("list mở/filter không đọc;204 tải lại count server và không optimistic", async () => {
  let reads = 0
  let lists = 0
  const { resource } = await setup({
    list: async () => {
      lists++
      return reads
        ? {
            ...page,
            unread_count: 0,
            items: [{ ...page.items[0], read_at: page.items[0].created_at }],
          }
        : page
    },
    read: async (_id, token) => {
      assert.equal(token, "token")
      reads++
    },
  })
  await resource.load()
  assert.equal(reads, 0)
  assert.equal(await resource.markRead(5), true)
  assert.equal(reads, 1)
  assert.equal(lists, 2)
  assert.equal(resource.getSnapshot().data?.unread_count, 0)
})
test("unknown read không tự retry, GET đối soát và retry chỉ do người dùng", async () => {
  let calls = 0
  const { resource } = await setup({
    read: async () => {
      calls++
      throw new AuthError(503, "SERVICE_UNAVAILABLE")
    },
  })
  await resource.load()
  assert.equal(await resource.markRead(5), false)
  assert.equal(calls, 1)
  assert.equal(resource.getSnapshot().data, undefined)
  assert.match(resource.getSnapshot().message, /Chưa xác định/)
  await resource.load()
  assert.equal(calls, 1)
  await resource.markRead(5)
  assert.equal(calls, 2)
})
test("lỗi trước CSRF không coi là POST unknown; double click chỉ một read", async () => {
  let calls = 0
  const preflight = await setup(
    {
      read: async () => {
        calls++
      },
    },
    async () => {
      throw new AuthError(503, "SERVICE_UNAVAILABLE")
    },
  )
  await preflight.resource.load()
  await preflight.resource.markRead(5)
  assert.equal(calls, 0)
  assert.doesNotMatch(preflight.resource.getSnapshot().message, /Chưa xác định/)
  let finish!: () => void
  const active = await setup({
    read: async () => {
      calls++
      await new Promise<void>((resolve) => {
        finish = resolve
      })
    },
  })
  await active.resource.load()
  const reading = active.resource.markRead(5)
  await active.resource.markRead(5)
  await Promise.resolve()
  assert.equal(calls, 1)
  finish()
  await reading
})
test("clear abort và bỏ GET/read cũ; đổi actor xóa cache qua subscription", async () => {
  let resolve!: (value: NotificationPage) => void
  let signal: AbortSignal | undefined
  const { resource, session, setActor } = await setup({
    list: async (_filter, _page, currentSignal) => {
      signal = currentSignal
      return new Promise((done) => {
        resolve = done
      })
    },
  })
  const unsubscribe = resource.subscribe(() => {})
  const loading = resource.load()
  resource.clear()
  assert.equal(signal?.aborted, true)
  resolve(page)
  await loading
  assert.equal(resource.getSnapshot().data, undefined)
  setActor({ ...session.getSnapshot().principal!, user_id: 2, employee_id: 9 })
  await session.refresh()
  assert.equal(resource.getSnapshot().phase, "idle")
  unsubscribe()
  // StrictEffects cleanup rồi subscribe lại phải hoạt động với cùng resource.
  const again = resource.subscribe(() => {})
  const next = resource.load()
  resolve(page)
  await next
  assert.equal(resource.getSnapshot().phase, "ready")
  again()
})
test("401/403 xóa data và kiểm phiên; read hoàn tất muộn không ghi lại state sau clear", async () => {
  for (const status of [401, 403]) {
    const { resource } = await setup({
      read: async () => {
        throw new AuthError(status, "FORBIDDEN")
      },
    })
    await resource.load()
    await resource.markRead(5)
    assert.equal(resource.getSnapshot().data, undefined)
  }
  let finish!: () => void
  let signal: AbortSignal | undefined
  const { resource } = await setup({
    read: async (_id, _token, currentSignal) => {
      signal = currentSignal
      await new Promise<void>((resolve) => {
        finish = resolve
      })
    },
  })
  await resource.load()
  const reading = resource.markRead(5)
  await Promise.resolve()
  resource.clear()
  assert.equal(signal?.aborted, true)
  finish()
  assert.equal(await reading, false)
  assert.equal(resource.getSnapshot().data, undefined)
  assert.equal(resource.getSnapshot().phase, "idle")
})
test("load chen trong read không thay GET sau commit bằng snapshot cũ", async () => {
  let finish!: () => void
  let reads = 0
  let lists = 0
  const { resource } = await setup({
    list: async () => {
      lists++
      return { ...page, unread_count: reads ? 0 : 1 }
    },
    read: async () => {
      await new Promise<void>((resolve) => {
        finish = resolve
      })
      reads++
    },
  })
  await resource.load()
  const reading = resource.markRead(5)
  await Promise.resolve()
  await resource.load()
  assert.equal(lists, 1)
  finish()
  await reading
  assert.equal(lists, 2)
  assert.equal(resource.getSnapshot().data?.unread_count, 0)
})
