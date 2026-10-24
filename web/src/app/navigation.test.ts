import assert from "node:assert/strict"
import { test } from "node:test"
import { projects, initialTasks } from "../demo/fixtures.ts"
import { demoIdentityOptions, selectWorkspaceNavigation } from "./navigation.ts"

test("điều hướng giữ quyền và ngữ cảnh dự án, không suy quyền từ nhãn demo", () => {
  const employee = selectWorkspaceNavigation(projects, initialTasks, "dung", "/my-work")
  assert.equal(employee.personalView, true)
  assert.equal(employee.ownTasks.length, 1)
  assert.equal(employee.visibleProjects.length, 0)
  assert.equal(employee.canLookup, false)
  const denied = selectWorkspaceNavigation(projects, initialTasks, "dung", "/projects/kho")
  assert.equal(denied.currentProject, undefined)
  const pm = selectWorkspaceNavigation(projects, initialTasks, "thanh", "/projects/kho")
  assert.equal(pm.currentProject?.id, "kho")
  assert.equal(pm.canLookup, true)
  assert.equal(
    selectWorkspaceNavigation(projects, initialTasks, "ha", "/personnel").canLookup,
    false,
  )
})

test("nhãn demo giữ thứ tự và tách trách nhiệm theo dự án khỏi role hệ thống", () => {
  const options = demoIdentityOptions(projects, projects[0])
  assert.deepEqual(
    options.slice(0, 4).map((i) => i.id),
    ["thanh", "khoa", "dung", "ha"],
  )
  assert.equal(options.find((i) => i.id === "thanh")?.responsibility, "PM (KHO-2026)")
  assert.equal(options.find((i) => i.id === "ha")?.responsibility, "Giám đốc")
  assert.equal(options.find((i) => i.id === "dung")?.responsibility, "Nhân viên")
})
