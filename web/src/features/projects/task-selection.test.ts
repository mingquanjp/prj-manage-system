import test from "node:test"
import assert from "node:assert/strict"
import { initialTasks } from "../../demo/fixtures.ts"
import { taskSelection, toggleTaskSelection } from "./task-selection.ts"

test("chọn cả nhóm chỉ tác động các việc đủ điều kiện của nhóm và giữ lựa chọn khác", () => {
  const frontend = initialTasks.filter((t) => t.projectId === "kho" && t.group === "Frontend")
  const backend = initialTasks.filter((t) => t.projectId === "kho" && t.group === "Backend")
  const original = ["t1"]
  const selected = toggleTaskSelection(frontend, original)
  assert.deepEqual(selected, ["t1", "t3"])
  assert.deepEqual(original, ["t1"])
  assert.equal(taskSelection(frontend, selected).hasAll, true)
  assert.equal(taskSelection(backend, selected).hasAll, false)
  assert.equal(taskSelection(backend, selected).isPartial, true)
  assert.deepEqual(toggleTaskSelection(frontend, selected), ["t1"])
})

test("bảng thường chọn theo tập đã lọc; nhóm không có việc chờ không sửa lựa chọn", () => {
  const filtered = initialTasks.filter((t) => t.projectId === "kho")
  assert.deepEqual(toggleTaskSelection(filtered, []), ["t1", "t2", "t3"])
  const noEligible = filtered.filter((t) => t.status !== "PENDING_ASSIGNMENT")
  assert.deepEqual(taskSelection(noEligible, ["t1"]).eligibleIds, [])
  assert.equal(taskSelection(noEligible, ["t1"]).hasAll, false)
  assert.deepEqual(toggleTaskSelection(noEligible, ["t1"]), ["t1"])
  assert.deepEqual(
    toggleTaskSelection(
      filtered.filter((t) => t.id === "t2"),
      ["t3"],
    ),
    ["t3", "t2"],
  )
})
