import test from "node:test"
import assert from "node:assert/strict"
import { safeReturnPath, sessionIssue, startDemoSession } from "./session.ts"
import { createDemoWorkspaceState } from "./workspace-state.ts"

test("Phiên demo chỉ mở cho tài khoản và hồ sơ đang hoạt động", () => {
  const state = createDemoWorkspaceState()
  const accounts = state.accounts.accounts
  const profiles = state.profileState.profiles
  const session = startDemoSession("khoa", accounts, profiles)
  assert.equal(sessionIssue(session, accounts, profiles), null)
  assert.throws(() => startDemoSession("missing", accounts, profiles))
  const locked = accounts.map((a) =>
    a.personId === "khoa" ? { ...a, status: "LOCKED" as const } : a,
  )
  assert.equal(sessionIssue(session, locked, profiles), "locked")
  assert.throws(() => startDemoSession("khoa", locked, profiles))
  const inactive = profiles.map((p) =>
    p.personId === "khoa" ? { ...p, employmentStatus: "INACTIVE" as const } : p,
  )
  assert.equal(sessionIssue(session, accounts, inactive), "inactive")
  assert.throws(() => startDemoSession("khoa", accounts, inactive))
})

test("Đổi quyền hệ thống làm phiên cũ không còn hợp lệ; đăng nhập lại lấy quyền hiện hành", () => {
  const state = createDemoWorkspaceState()
  const accounts = state.accounts.accounts
  const profiles = state.profileState.profiles
  const session = startDemoSession("khoa", accounts, profiles)
  const changed = accounts.map((a) =>
    a.personId === "khoa" ? { ...a, systemRoles: ["HR" as const] } : a,
  )
  assert.equal(sessionIssue(session, changed, profiles), "revoked")
  assert.equal(sessionIssue(startDemoSession("khoa", changed, profiles), changed, profiles), null)
  const reordered = changed.map((a) =>
    a.personId === "khoa" ? { ...a, systemRoles: ["ADMIN" as const, "HR" as const] } : a,
  )
  const combined = startDemoSession("khoa", reordered, profiles)
  assert.equal(
    sessionIssue(
      combined,
      reordered.map((a) => ({ ...a, systemRoles: [...a.systemRoles].reverse() })),
      profiles,
    ),
    null,
  )
})

test("Đích đăng nhập lại chỉ là đường dẫn nội bộ, không chuyển ra ngoài hoặc lặp trang đăng nhập", () => {
  assert.equal(
    safeReturnPath("/projects/kho?task=t6&panel=results"),
    "/projects/kho?task=t6&panel=results",
  )
  for (const path of [
    null,
    "",
    "https://example.com",
    "//example.com",
    "/\\example.com",
    "/\n/example.com",
    "/login?returnTo=/login",
  ])
    assert.equal(safeReturnPath(path), "/")
})
