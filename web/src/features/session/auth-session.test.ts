import { test } from "node:test"
import assert from "node:assert/strict"
import { AuthError, type AuthApi, type Principal } from "./auth-api.ts"
import { createAuthSession } from "./auth-session.ts"

const person: Principal = {
  user_id: 12,
  employee_id: 25,
  employee_code: "E025",
  fullname: "Nhân viên thử",
  system_roles: ["HR"],
  must_change_password: false,
}
function fake(overrides: Partial<AuthApi> = {}): AuthApi {
  return {
    me: async () => person,
    login: async () => person,
    csrf: async () => "token",
    logout: async () => {},
    changePassword: async () => {},
    ...overrides,
  }
}
function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (error: unknown) => void
  const promise = new Promise<T>((yes, no) => {
    resolve = yes
    reject = no
  })
  return { promise, resolve, reject }
}

test("bootstrap 401 đến login; reload đọc quyền/cờ mới từ server, không giữ principal cũ", async () => {
  let response: Principal | null = null
  const session = createAuthSession(
    fake({
      me: async () => {
        if (!response) throw new AuthError(401, "UNAUTHORIZED")
        return response
      },
    }),
  )
  await session.check()
  assert.equal(session.getSnapshot().phase, "signed-out")
  response = person
  await session.check()
  assert.deepEqual(session.getSnapshot().principal, person)
  response = { ...person, system_roles: [], must_change_password: true }
  const checking = session.check()
  assert.equal(session.getSnapshot().principal, null)
  await checking
  assert.equal(session.getSnapshot().principal?.must_change_password, true)
})

test("login đúng/sai và hạn chế; không lặp POST hoặc chấp nhận role client", async () => {
  const pending = deferred<Principal>()
  let calls = 0
  const session = createAuthSession(
    fake({
      me: async () => {
        throw new AuthError(401, "UNAUTHORIZED")
      },
      login: async () => {
        calls++
        return pending.promise
      },
    }),
  )
  await session.check()
  const login = session.login("test@example.com", "pass")
  await session.login("test@example.com", "pass")
  await session.check()
  assert.equal(calls, 1)
  pending.resolve({ ...person, system_roles: [], must_change_password: true })
  await login
  assert.equal(session.getSnapshot().principal?.must_change_password, true)
  assert.deepEqual(session.getSnapshot().principal?.system_roles, [])
  const bad = createAuthSession(
    fake({
      me: async () => {
        throw new AuthError(401, "UNAUTHORIZED")
      },
      login: async () => {
        throw new AuthError(401, "INVALID_CREDENTIALS")
      },
    }),
  )
  await bad.check()
  await bad.login("bad@example.com", "wrong")
  assert.equal(bad.getSnapshot().phase, "signed-out")
  assert.match(bad.getSnapshot().message, /Email hoặc mật khẩu/)
})

test("logout lấy token trước mutation, thành công xóa phiên; phiên revoke không khôi phục", async () => {
  let revoked = false
  const calls: string[] = []
  const session = createAuthSession(
    fake({
      me: async () => {
        if (revoked) throw new AuthError(401, "UNAUTHORIZED")
        return person
      },
      csrf: async () => {
        calls.push("csrf")
        return "token-phiên"
      },
      logout: async (token) => {
        calls.push(token)
        revoked = true
      },
    }),
  )
  await session.check()
  await session.logout()
  assert.deepEqual(calls, ["csrf", "token-phiên"])
  assert.equal(session.getSnapshot().principal, null)
  assert.match(session.getSnapshot().message, /đã đăng xuất/)
  await session.check()
  assert.equal(session.getSnapshot().phase, "signed-out")
})

test("503/mất mạng không giả hết phiên/logout thành công; khôi phục bằng GET, không retry POST", async () => {
  for (const error of [
    new AuthError(503, "SERVICE_UNAVAILABLE"),
    new AuthError(0, "NETWORK_ERROR"),
  ]) {
    let logoutCount = 0
    const session = createAuthSession(
      fake({
        logout: async () => {
          logoutCount++
          throw error
        },
      }),
    )
    await session.check()
    await session.logout()
    assert.equal(session.getSnapshot().phase, "unavailable")
    assert.equal(session.getSnapshot().principal, null)
    assert.doesNotMatch(session.getSnapshot().message, /đã đăng xuất/)
    await session.check()
    assert.equal(session.getSnapshot().phase, "authenticated")
    assert.equal(logoutCount, 1)
    const boot = createAuthSession(
      fake({
        me: async () => {
          throw error
        },
      }),
    )
    await boot.check()
    assert.equal(boot.getSnapshot().phase, "unavailable")
  }
})

test("CSRF sai không báo logout, lần thử tiếp lấy token mới; 401 bỏ phiên", async () => {
  let tokenCalls = 0
  let mutations = 0
  const session = createAuthSession(
    fake({
      csrf: async () => `token${++tokenCalls}`,
      logout: async () => {
        if (++mutations === 1) throw new AuthError(403, "FORBIDDEN")
      },
    }),
  )
  await session.check()
  await session.logout()
  assert.equal(session.getSnapshot().phase, "authenticated")
  await session.logout()
  assert.equal(tokenCalls, 2)
  assert.equal(session.getSnapshot().phase, "signed-out")
  const expired = createAuthSession(
    fake({
      csrf: async () => {
        throw new AuthError(401, "UNAUTHORIZED")
      },
      logout: async () => {
        assert.fail("Không được POST khi lấy CSRF thất bại")
      },
    }),
  )
  await expired.check()
  await expired.logout()
  assert.equal(expired.getSnapshot().phase, "signed-out")
})

test("429 dùng clock/Retry-After, không gửi thêm mutation trước hạn", async () => {
  let now = 1000
  let count = 0
  const session = createAuthSession(
    fake({
      me: async () => {
        throw new AuthError(401, "UNAUTHORIZED")
      },
      login: async () => {
        count++
        throw new AuthError(429, "RATE_LIMITED", 30)
      },
    }),
    () => now,
  )
  await session.check()
  await session.login("test@example.com", "pass")
  assert.equal(session.getSnapshot().retryAt, 31000)
  await session.check()
  assert.equal(session.getSnapshot().retryAt, 31000)
  await session.login("test@example.com", "pass")
  assert.equal(count, 1)
  now = 31000
  await session.login("test@example.com", "pass")
  assert.equal(count, 2)
})

test("phản hồi bootstrap cũ không ghi đè kết quả kiểm tra mới", async () => {
  const first = deferred<Principal>()
  let count = 0
  const session = createAuthSession(
    fake({
      me: async () => {
        if (++count === 1) return first.promise
        throw new AuthError(401, "UNAUTHORIZED")
      },
    }),
  )
  const old = session.check()
  await session.check()
  first.resolve(person)
  await old
  assert.equal(session.getSnapshot().phase, "signed-out")
  assert.equal(session.getSnapshot().principal, null)
})

test("mất phản hồi logout sau khi server đã revoke: chỉ GET đối soát, không tự POST lại", async () => {
  let revoked = false
  let writes = 0
  const session = createAuthSession(
    fake({
      me: async () => {
        if (revoked) throw new AuthError(401, "UNAUTHORIZED")
        return person
      },
      logout: async () => {
        writes++
        revoked = true
        throw new AuthError(0, "NETWORK_ERROR")
      },
    }),
  )
  await session.check()
  await session.logout()
  assert.equal(session.getSnapshot().phase, "unavailable")
  assert.doesNotMatch(session.getSnapshot().message, /đã đăng xuất/)
  await session.check()
  assert.equal(session.getSnapshot().phase, "signed-out")
  assert.equal(writes, 1)
})

const limited: Principal = { ...person, system_roles: [], must_change_password: true }
test("đổi thành công: CSRF trước POST, khóa double submit, bỏ phiên và đòi login mới", async () => {
  const pending = deferred<void>()
  const calls: string[] = []
  let meCalls = 0
  const session = createAuthSession(
    fake({
      me: async () => {
        meCalls++
        return limited
      },
      csrf: async () => {
        calls.push("csrf")
        return "token"
      },
      changePassword: async (current, next, token) => {
        calls.push(`${current}/${next}/${token}`)
        return pending.promise
      },
    }),
  )
  await session.check()
  const changing = session.changePassword("old", "new")
  await session.changePassword("old", "new")
  await session.logout()
  await session.check()
  pending.resolve()
  await changing
  assert.deepEqual(calls, ["csrf", "old/new/token"])
  assert.equal(session.getSnapshot().phase, "signed-out")
  assert.equal(session.getSnapshot().principal, null)
  assert.match(session.getSnapshot().message, /Đổi mật khẩu thành công/)
  await session.check()
  assert.equal(meCalls, 1)
  await session.login("email", "new")
  assert.equal(session.getSnapshot().phase, "authenticated")
})

test("chỉ phiên hạn chế được gửi đổi; lỗi CSRF trước POST không bị nhận là đã gửi", async () => {
  for (const restricted of [false, true]) {
    let posted = 0
    const session = createAuthSession(
      fake({
        me: async () => (restricted ? limited : person),
        csrf: async () => {
          throw new AuthError(503, "SERVICE_UNAVAILABLE")
        },
        changePassword: async () => {
          posted++
        },
      }),
    )
    await session.check()
    await session.changePassword("old", "new")
    assert.equal(posted, 0)
    assert.equal(session.getSnapshot().phase, restricted ? "unavailable" : "authenticated")
  }
})

test("sai current, policy, 401, 403 và 429 không báo thành công; cooldown qua focus", async () => {
  for (const [status, code] of [
    [400, "INVALID_CURRENT_PASSWORD"],
    [400, "VALIDATION_ERROR"],
    [401, "UNAUTHENTICATED"],
    [403, "FORBIDDEN"],
    [429, "RATE_LIMITED"],
  ] as const) {
    let calls = 0
    let clock = 1000
    const session = createAuthSession(
      fake({
        me: async () => limited,
        changePassword: async () => {
          calls++
          throw new AuthError(status, code, 10)
        },
      }),
      () => clock,
    )
    await session.check()
    await session.changePassword("old", "new")
    assert.equal(session.getSnapshot().phase, status === 401 ? "signed-out" : "authenticated")
    assert.doesNotMatch(session.getSnapshot().message, /thành công/)
    if (status === 400) assert.equal(session.getSnapshot().issueCode, code)
    if (status === 429) {
      await session.check()
      await session.changePassword("old", "new")
      assert.equal(calls, 1)
      clock = 11000
      await session.changePassword("old", "new")
      assert.equal(calls, 2)
    }
  }
})

test("cờ không còn yêu cầu đổi: đọc me thay vì tự nâng principal", async () => {
  let reads = 0
  const session = createAuthSession(
    fake({
      me: async () => (++reads === 1 ? limited : person),
      changePassword: async () => {
        throw new AuthError(403, "PASSWORD_CHANGE_NOT_REQUIRED")
      },
    }),
  )
  await session.check()
  await session.changePassword("old", "new")
  assert.equal(reads, 2)
  assert.deepEqual(session.getSnapshot().principal, person)
})

test("mất phản hồi sau POST/500/503: giữ trạng thái chưa rõ qua focus, chỉ login thủ công", async () => {
  for (const error of [
    new AuthError(0, "NETWORK_ERROR"),
    new AuthError(500, "INTERNAL_ERROR"),
    new AuthError(503, "SERVICE_UNAVAILABLE"),
  ]) {
    let calls = 0
    let reads = 0
    const session = createAuthSession(
      fake({
        me: async () => {
          reads++
          return limited
        },
        changePassword: async () => {
          calls++
          throw error
        },
      }),
    )
    await session.check()
    await session.changePassword("old", "new")
    await session.check()
    await session.changePassword("old", "new")
    assert.equal(session.getSnapshot().phase, "password-uncertain")
    assert.equal(session.getSnapshot().principal, null)
    assert.equal(reads, 1)
    assert.equal(calls, 1)
    session.loginAfterUncertainChange()
    await session.check()
    assert.equal(reads, 1)
    assert.equal(session.getSnapshot().phase, "signed-out")
    assert.match(session.getSnapshot().message, /mật khẩu mới.*mật khẩu cũ/)
  }
})

test("login kiểm tra sau kết quả chưa rõ bị lỗi mạng: vẫn có đường kiểm lại phiên", async () => {
  const session = createAuthSession(
    fake({
      me: async () => limited,
      changePassword: async () => {
        throw new AuthError(0, "NETWORK_ERROR")
      },
      login: async () => {
        throw new AuthError(0, "NETWORK_ERROR")
      },
    }),
  )
  await session.check()
  await session.changePassword("old", "new")
  session.loginAfterUncertainChange()
  await session.login("email", "new")
  assert.equal(session.getSnapshot().phase, "unavailable")
  await session.check()
  assert.equal(session.getSnapshot().phase, "authenticated")
})
