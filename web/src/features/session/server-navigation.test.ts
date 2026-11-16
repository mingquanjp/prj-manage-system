import { test } from "node:test"
import assert from "node:assert/strict"
import type { Principal, SystemRole } from "./auth-api.ts"
import {
  canReadCatalogs,
  canManageCatalogs,
  serverHome,
  catalogKindAt,
  parseCatalogQuery,
} from "./server-navigation.ts"
const principal = (roles: SystemRole[], limited = false): Principal => ({
  user_id: 1,
  employee_id: 2,
  employee_code: "EMP000001",
  fullname: "Test",
  system_roles: roles,
  must_change_password: limited,
})
test("navigation cộng role server, không coi ADMIN/DIRECTOR là HR; limited chặn mọi tổ hợp", () => {
  for (const roles of [
    [],
    ["HR"],
    ["ADMIN"],
    ["DIRECTOR"],
    ["HR", "ADMIN"],
    ["ADMIN", "DIRECTOR"],
    ["HR", "DIRECTOR"],
    ["HR", "ADMIN", "DIRECTOR"],
  ] as SystemRole[][]) {
    const p = principal(roles)
    assert.equal(canReadCatalogs(p), roles.includes("HR") || roles.includes("ADMIN"))
    assert.equal(canManageCatalogs(p), roles.includes("HR"))
    assert.equal(serverHome(p), roles.includes("HR") ? "/organization/departments" : "/profile")
    const limited = principal(roles, true)
    assert.equal(canReadCatalogs(limited), false)
    assert.equal(canManageCatalogs(limited), false)
    assert.equal(serverHome(limited), "/change-password")
  }
})
test("URL/query allowlist và pagination không cấp quyền", () => {
  assert.equal(catalogKindAt("/organization/job-levels"), "job-levels")
  assert.equal(catalogKindAt("/organization/accounts"), undefined)
  assert.deepEqual(parseCatalogQuery("?q=%25_&page=2&page_size=10"), {
    q: "%_",
    page: 2,
    pageSize: 10,
  })
  assert.deepEqual(parseCatalogQuery(""), { q: "", page: 1, pageSize: 20 })
  for (const q of [
    "?role=HR",
    "?page=0",
    "?page=-1",
    "?page=1&page=2",
    "?page_size=101",
    "?page=2147483648",
    "?q=" + "x".repeat(101),
    "?q=%00",
  ])
    assert.equal(parseCatalogQuery(q), null)
})
