import { test } from "node:test"
import assert from "node:assert/strict"
import { initialTasks, projects, loadFor } from "./model.ts"
import { canLookupPersonnel, personnelProfiles } from "./personnel.ts"

test("tra cứu chỉ dành cho PM/PTCM hiện hành, không suy từ vai trò Giám đốc", () => {
  for (const actor of ["thanh", "khoa"]) assert.equal(canLookupPersonnel(projects, actor), true)
  for (const actor of ["ha", "dung", "unknown"]) {
    assert.equal(canLookupPersonnel(projects, actor), false)
    assert.deepEqual(personnelProfiles(projects, initialTasks, actor), [])
  }
  assert.equal(
    canLookupPersonnel(
      projects.map((p) => ({ ...p, ptcmId: "cuong" })),
      "khoa",
    ),
    false,
  )
  assert.equal(canLookupPersonnel([{ ...projects[0], ptcmId: "ha" }], "ha"), true)
})

test("tìm không dấu/tên/kỹ năng; projection không trả dữ liệu gốc hoặc điểm gợi ý", () => {
  assert.equal(personnelProfiles(projects, initialTasks, "khoa", "cuong")[0]?.name, "Lê Minh Cường")
  assert.equal(
    personnelProfiles(projects, initialTasks, "khoa", "DŨNG")[0]?.name,
    "Phạm Quang Dũng",
  )
  assert.ok(personnelProfiles(projects, initialTasks, "thanh", "PostgreSQL").length)
  assert.deepEqual(personnelProfiles(projects, initialTasks, "thanh", "khongtontai"), [])
  const profile = personnelProfiles(projects, initialTasks, "thanh", "Nguyễn Văn An")[0]
  assert.deepEqual(
    Object.keys(profile).sort(),
    ["id", "name", "initials", "position", "seniority", "skills", "currentLoad"].sort(),
  )
  assert.equal(profile.currentLoad, loadFor(profile.id, initialTasks))
})

test("tải tra cứu phản ánh state hiện hành cùng nguồn phân công, không đổi fixture", () => {
  const next = initialTasks.map((task) =>
    task.id === "t1" ? { ...task, assigneeId: "an", status: "ASSIGNED" as const } : task,
  )
  const before = personnelProfiles(projects, initialTasks, "khoa", "Nguyễn Văn An")[0]
  const after = personnelProfiles(projects, next, "khoa", "Nguyễn Văn An")[0]
  assert.equal(
    after.currentLoad,
    before.currentLoad + initialTasks.find((t) => t.id === "t1")!.estimate,
  )
  assert.equal(initialTasks.find((t) => t.id === "t1")!.status, "PENDING_ASSIGNMENT")
})
