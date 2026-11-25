import test from "node:test"
import assert from "node:assert/strict"
import { createAuthSession } from "../session/auth-session.ts"
import { AuthError, type AuthApi, type Principal } from "../session/auth-api.ts"
import type { CatalogApi } from "../organization/catalog-api.ts"
import { createAdministrativeProfileResource } from "./administrative-profile-resource.ts"
import type {
  AdministrativeProfileApi,
  AdministrativeProfile,
} from "./administrative-profile-api.ts"
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
const patch = {
  expected_before_values: { fullname: "Nhân viên" },
  changes: { fullname: "Tên mới" },
}
async function setup(
  override: Partial<AdministrativeProfileApi> = {},
  catalogOverride?: Pick<CatalogApi, "list">,
) {
  let actor = {
    user_id: 102,
    employee_id: 1002,
    must_change_password: false,
    system_roles: ["HR"],
  } as Principal
  const session = createAuthSession({
    me: async () => actor,
    csrf: async () => "token",
  } as unknown as AuthApi)
  await session.check()
  const resource = createAdministrativeProfileResource(
    {
      list: async () => ({ items: [], total: 0, page: 1, page_size: 20 }),
      detail: async () => profile,
      edit: async () => ({ ...profile, fullname: "Tên mới" }),
      ...override,
    },
    session,
    catalogOverride ??
      ({
        list: async () => ({
          items: [{ id: 1, name: "Danh mục" }],
          pagination: { page: 1, page_size: 100, total: 1 },
        }),
      } as Pick<CatalogApi, "list">),
  )
  return {
    session,
    resource,
    changeActor: (next: Principal) => {
      actor = next
    },
  }
}
test("unknown PATCH blocks replay until explicit GET", async () => {
  let calls = 0
  const { resource } = await setup({
    edit: async () => {
      calls++
      throw new AuthError(503, "COMMIT_UNKNOWN")
    },
  })
  await resource.load(1001)
  await resource.save(patch)
  assert.equal(resource.getSnapshot().detail?.fullname, "Nhân viên")
  assert.equal(resource.getSnapshot().mustReload, true)
  await resource.save(patch)
  assert.equal(calls, 1)
  await resource.load(1001)
  assert.equal(resource.getSnapshot().mustReload, false)
})
test("stale/catalog lock draft; email conflict stays editable", async () => {
  for (const code of ["STALE_DATA", "CATALOG_NOT_FOUND", "EMAIL_CONFLICT"]) {
    const { resource } = await setup({
      edit: async () => {
        throw new AuthError(409, code)
      },
    })
    await resource.load(1001)
    await resource.save(patch)
    assert.equal(resource.getSnapshot().mustReload, code !== "EMAIL_CONFLICT")
    assert.equal(
      resource.getSnapshot().fieldError,
      code === "EMAIL_CONFLICT" ? "company_email" : undefined,
    )
  }
})
test("clear prevents late GET/PATCH from restoring PII", async () => {
  let resolve!: (v: AdministrativeProfile) => void
  const first = await setup({
    detail: () =>
      new Promise((done) => {
        resolve = done
      }),
  })
  const loading = first.resource.load(1001)
  first.resource.clear()
  resolve(profile)
  await loading
  assert.equal(first.resource.getSnapshot().detail, undefined)
  const second = await setup({
    edit: () =>
      new Promise((done) => {
        resolve = done
      }),
  })
  await second.resource.load(1001)
  const saving = second.resource.save(patch)
  await Promise.resolve()
  await Promise.resolve()
  second.resource.clear()
  resolve(profile)
  await saving
  assert.equal(second.resource.getSnapshot().detail, undefined)
})
test("forbidden clears detail/catalog and refreshes permission", async () => {
  const { resource, changeActor } = await setup({
    edit: async () => {
      throw new AuthError(403, "FORBIDDEN")
    },
  })
  await resource.load(1001)
  changeActor({
    user_id: 102,
    employee_id: 1002,
    fullname: "HR",
    employee_code: "EMP000002",
    system_roles: [],
    must_change_password: false,
  } as Principal)
  await resource.save(patch)
  assert.equal(resource.getSnapshot().detail, undefined)
  assert.equal(resource.getSnapshot().catalogs, undefined)
})
test("catalog pagination loads all pages; no-HR GET prevented", async () => {
  const pages: number[] = []
  const { resource, session, changeActor } = await setup({}, {
    list: async (_kind, _q, page) => {
      pages.push(page)
      return {
        items:
          page === 1
            ? Array.from({ length: 100 }, (_, i) => ({ id: i + 1, name: "Danh mục" }))
            : [{ id: 101, name: "Cuối" }],
        pagination: { page, page_size: 100, total: 101 },
      }
    },
  } as Pick<CatalogApi, "list">)
  await resource.load(1001)
  assert.deepEqual(pages, [1, 2, 1, 2, 1, 2])
  assert.equal(resource.getSnapshot().catalogs?.positions.length, 101)
  changeActor({ user_id: 103, employee_id: 1003, system_roles: ["ADMIN"] } as Principal)
  await session.refresh()
  await resource.load(1001)
  assert.equal(resource.getSnapshot().detail, undefined)
})
test("CSRF failure before PATCH does not claim mutation uncertainty", async () => {
  let calls = 0
  const session = createAuthSession({
    me: async () => ({
      user_id: 102,
      employee_id: 1002,
      fullname: "HR",
      employee_code: "EMP2",
      system_roles: ["HR"],
      must_change_password: false,
    }),
    csrf: async () => {
      throw new AuthError(503, "UNAVAILABLE")
    },
  } as unknown as AuthApi)
  await session.check()
  const resource = createAdministrativeProfileResource(
    {
      list: async () => ({ items: [], total: 0, page: 1, page_size: 20 }),
      detail: async () => profile,
      edit: async () => {
        calls++
        return profile
      },
    } as AdministrativeProfileApi,
    session,
    { list: async () => ({ items: [], pagination: { page: 1, page_size: 100, total: 0 } }) },
  )
  await resource.load(1001)
  await resource.save(patch)
  assert.equal(calls, 0)
  assert.equal(resource.getSnapshot().mustReload, false)
  assert.doesNotMatch(resource.getSnapshot().message, /Chưa xác định/)
})
test("self email confirmed save rechecks revoked session without losing success classification", async () => {
  let reads = 0
  const session = createAuthSession({
    me: async () => {
      reads++
      if (reads > 1) throw new AuthError(401, "UNAUTHENTICATED")
      return {
        user_id: 101,
        employee_id: 1001,
        fullname: "HR",
        employee_code: "EMP1",
        system_roles: ["HR"],
        must_change_password: false,
      }
    },
    csrf: async () => "token",
  } as unknown as AuthApi)
  await session.check()
  const resource = createAdministrativeProfileResource(
    {
      list: async () => ({ items: [], total: 0, page: 1, page_size: 20 }),
      detail: async () => profile,
      edit: async () => ({ ...profile, company_email: "new@example.invalid" }),
    } as AdministrativeProfileApi,
    session,
    { list: async () => ({ items: [], pagination: { page: 1, page_size: 100, total: 0 } }) },
  )
  await resource.load(1001)
  assert.equal(
    await resource.save({
      expected_before_values: { company_email: profile.company_email },
      changes: { company_email: "new@example.invalid" },
    }),
    true,
  )
  assert.equal(session.getSnapshot().phase, "signed-out")
  assert.match(resource.getSnapshot().message, /Đã lưu/)
})
test("catalog transport receives abort and clear stops additional pages", async () => {
  let finish!: (value: Awaited<ReturnType<CatalogApi["list"]>>) => void
  let observed: AbortSignal | undefined
  let calls = 0
  const { resource } = await setup(
    {},
    {
      list: (_kind, _q, _page, _size, signal) => {
        calls++
        observed = signal
        return new Promise((done) => {
          finish = done
        })
      },
    },
  )
  const loading = resource.load(1001)
  await Promise.resolve()
  resource.clear()
  assert.equal(observed?.aborted, true)
  finish({
    items: [{ id: 1, name: "Danh mục" }],
    pagination: { page: 1, page_size: 100, total: 101 },
  })
  await loading
  assert.equal(calls, 1)
  assert.equal(resource.getSnapshot().catalogs, undefined)
})
test("confirmed no-op is distinguished from persisted changes", async () => {
  const { resource } = await setup({ edit: async () => profile })
  await resource.load(1001)
  assert.equal(await resource.save(patch), true)
  assert.match(resource.getSnapshot().message, /Không có thay đổi/)
})
