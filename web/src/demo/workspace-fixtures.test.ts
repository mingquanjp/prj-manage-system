import test from "node:test"
import assert from "node:assert/strict"
import { createDemoWorkspaceState, demoWorkspaceReducer } from "./workspace-state.ts"
import { reviewSubmission } from "./effort.ts"
import { initialOutputConflicts } from "./fixtures.ts"
import { lifecycleBlockReason } from "./project-lifecycle.ts"

test("task mẫu chờ nghiệm thu/DONE có lần nộp và mốc lượt giao nhất quán", () => {
  const state = createDemoWorkspaceState()
  for (const task of state.tasks.filter((t) => ["PENDING_REVIEW", "DONE"].includes(t.status))) {
    const submission = state.work.submissions.filter((s) => s.taskId === task.id).at(-1)!
    assert.ok(submission)
    assert.equal(submission.assignmentId, task.currentAssignment?.id)
    assert.equal(submission.employeeId, task.assigneeId)
    assert.ok(Date.parse(submission.submittedAt) >= Date.parse(task.currentAssignment!.assignedAt))
    if (task.status === "PENDING_REVIEW") {
      assert.equal(submission.review, undefined)
      assert.equal(task.currentAssignment?.endedAt, undefined)
    } else {
      assert.equal(submission.review?.result, "ACCEPTED")
      assert.equal(submission.review?.at, task.acceptedAt)
      assert.equal(task.currentAssignment?.endedAt, task.acceptedAt)
    }
  }
  assert.deepEqual(state.notifications, [])
  assert.deepEqual(state.work.notifications, [])
  assert.deepEqual(demoWorkspaceReducer(state, { type: "reset" }).work, state.work)
})

test("PTCM có thể trả sửa hoặc nghiệm thu lần nộp mẫu, sau đó PM hoàn thành khi không còn việc mở", () => {
  const state = createDemoWorkspaceState()
  const submission = state.work.submissions.find((s) => s.taskId === "t7")!
  const returned = reviewSubmission(
    state.projects,
    state.tasks,
    state.work,
    "khoa",
    submission.id,
    "REJECTED",
    "Kiểm tra lại trường hợp không có kết quả",
    initialOutputConflicts,
  )
  assert.equal(returned.tasks.find((t) => t.id === "t7")?.status, "IN_PROGRESS")
  const accepted = reviewSubmission(
    state.projects,
    state.tasks,
    state.work,
    "khoa",
    submission.id,
    "ACCEPTED",
    "",
    initialOutputConflicts,
  )
  assert.equal(accepted.tasks.find((t) => t.id === "t7")?.status, "DONE")
  const tasks = accepted.tasks.map((t) =>
    t.projectId === "kho" && t.status !== "DONE" ? { ...t, status: "CANCELLED" as const } : t,
  )
  assert.equal(
    lifecycleBlockReason(state.projects[0], tasks, accepted.work, "thanh", "complete"),
    undefined,
  )
  assert.equal(state.tasks.find((t) => t.id === "t7")?.status, "PENDING_REVIEW")
})
