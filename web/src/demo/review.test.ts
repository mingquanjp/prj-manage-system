import { test } from "node:test"
import assert from "node:assert/strict"
import {
  initialTasks,
  projects,
  transitionPersonalTask,
  progress,
  loadFor,
  type Task,
} from "./model.ts"
import {
  emptyWorkState,
  saveLog,
  submitResult,
  reviewSubmission,
  type WorkState,
} from "./effort.ts"
import { projectBoard } from "./kanban.ts"

const now = new Date("2026-10-21T09:00:00Z")
const draft = () =>
  submitResult(
    projects,
    transitionPersonalTask(
      projects,
      initialTasks,
      "dung",
      "t6",
      "ASSIGNED",
      "start",
    ),
    emptyWorkState(),
    "dung",
    "t6",
    "Đã làm API báo cáo",
    "https://example.com/pr/1",
    now,
  )
const decide = (
  tasks: Task[],
  work: WorkState,
  id: string,
  result: "ACCEPTED" | "REJECTED",
  comment = "",
) =>
  reviewSubmission(projects, tasks, work, "khoa", id, result, comment, [], now)

test("trả sửa → log → nộp lại → chấp nhận; giữ snapshot/quyết định cũ, không tự hoàn thành dự án", () => {
  const pending = draft(),
    id = pending.work.submissions[0].id
  const pendingCopy = structuredClone(pending)
  assert.throws(
    () => decide(pending.tasks, pending.work, id, "REJECTED", " "),
    /lý do/,
  )
  assert.deepEqual(pending, pendingCopy)
  const returned = decide(
    pending.tasks,
    pending.work,
    id,
    "REJECTED",
    "Bổ sung kiểm thử trường hợp không phát sinh",
  )
  assert.equal(returned.tasks.find((t) => t.id === "t6")!.status, "IN_PROGRESS")
  assert.equal(returned.work.submissions[0].review?.result, "REJECTED")
  assert.equal(
    returned.tasks.find((t) => t.id === "t6")!.currentAssignment?.endedAt,
    undefined,
  )
  const logged = saveLog(
    projects,
    returned.tasks,
    returned.work,
    "dung",
    "t6",
    {
      workDate: "2026-10-21",
      minutes: 60,
      description: "Bổ sung kiểm thử",
      overLimitReason: "",
    },
    undefined,
    now,
  )
  const resubmitted = submitResult(
    projects,
    returned.tasks,
    logged,
    "dung",
    "t6",
    "Đã bổ sung kiểm thử",
    "",
    now,
  )
  assert.equal(resubmitted.work.submissions.length, 2)
  assert.equal(resubmitted.work.submissions[0].logSnapshot.logs.length, 0)
  assert.equal(resubmitted.work.submissions[1].logSnapshot.logs.length, 1)
  assert.throws(
    () => decide(resubmitted.tasks, resubmitted.work, id, "ACCEPTED"),
    /đã thay đổi/,
  )
  const accepted = decide(
    resubmitted.tasks,
    resubmitted.work,
    resubmitted.work.submissions[1].id,
    "ACCEPTED",
  )
  const task = accepted.tasks.find((t) => t.id === "t6")!
  assert.equal(task.status, "DONE")
  assert.equal(task.acceptedAt, now.toISOString())
  assert.equal(task.currentAssignment?.endedAt, now.toISOString())
  assert.equal(task.currentAssignment?.endedBy, "khoa")
  assert.equal(task.currentAssignment?.endReason, "DONE")
  assert.equal(task.assigneeId, "dung")
  assert.equal(task.criteria, initialTasks.find((t) => t.id === "t6")!.criteria)
  assert.equal(task.executionHistory?.at(-1)?.from, "PENDING_REVIEW")
  assert.equal(
    accepted.work.submissions[0].review?.comment,
    "Bổ sung kiểm thử trường hợp không phát sinh",
  )
  assert.ok(progress(accepted.tasks)! > progress(pending.tasks)!)
  assert.ok(loadFor("dung", accepted.tasks) < loadFor("dung", pending.tasks))
  assert.equal(projects.find((p) => p.id === "kho")!.status, "IN_PROGRESS")
  assert.deepEqual(pending, pendingCopy)
})

test("thiếu log vẫn nghiệm thu; chặn quyết định lặp và mở lại DONE", () => {
  const pending = draft(),
    id = pending.work.submissions[0].id
  const accepted = decide(pending.tasks, pending.work, id, "ACCEPTED")
  assert.equal(accepted.work.submissions[0].logSnapshot.logs.length, 0)
  assert.throws(() =>
    decide(accepted.tasks, accepted.work, id, "REJECTED", "đổi ý"),
  )
  assert.throws(() =>
    submitResult(
      projects,
      accepted.tasks,
      accepted.work,
      "dung",
      "t6",
      "Nộp lại",
      "",
      now,
    ),
  )
  assert.throws(() =>
    reviewSubmission(
      projects,
      pending.tasks,
      pending.work,
      "khoa",
      "missing",
      "ACCEPTED",
      "",
      [],
      now,
    ),
  )
})

test("chỉ PTCM hiện hành: PM/Director/người nộp/PTCM cũ bị từ chối, không tự nghiệm thu đóng góp cũ", () => {
  const pending = draft(),
    id = pending.work.submissions[0].id
  for (const actor of ["thanh", "ha", "dung", "unknown"])
    assert.throws(() =>
      reviewSubmission(
        projects,
        pending.tasks,
        pending.work,
        actor,
        id,
        "ACCEPTED",
        "",
        [],
        now,
      ),
    )
  const changed = projects.map((p) =>
    p.id === "kho" ? { ...p, ptcmId: "cuong" } : p,
  )
  assert.throws(() =>
    reviewSubmission(
      changed,
      pending.tasks,
      pending.work,
      "khoa",
      id,
      "ACCEPTED",
      "",
      [],
      now,
    ),
  )
  assert.equal(
    reviewSubmission(
      changed,
      pending.tasks,
      pending.work,
      "cuong",
      id,
      "ACCEPTED",
      "",
      [],
      now,
    ).work.submissions[0].review?.reviewerId,
    "cuong",
  )
  assert.throws(
    () =>
      reviewSubmission(
        projects,
        pending.tasks,
        pending.work,
        "khoa",
        id,
        "ACCEPTED",
        "",
        [{ projectId: "kho", taskId: "t6", personId: "khoa" }],
        now,
      ),
    /chính mình/,
  )
  const selfProject = projects.map((p) =>
    p.id === "kho" ? { ...p, ptcmId: "dung" } : p,
  )
  assert.throws(
    () =>
      reviewSubmission(
        selfProject,
        pending.tasks,
        pending.work,
        "dung",
        id,
        "ACCEPTED",
        "",
        [],
        now,
      ),
    /chính mình/,
  )
})

test("khóa project/task/lượt giao: không nghiệm thu terminal hoặc submission cũ", () => {
  const pending = draft(),
    id = pending.work.submissions[0].id
  for (const status of [
    "PLANNING",
    "COMPLETED",
    "CLOSED",
    "CANCELLED",
  ] as const)
    assert.throws(() =>
      reviewSubmission(
        projects.map((p) => (p.id === "kho" ? { ...p, status } : p)),
        pending.tasks,
        pending.work,
        "khoa",
        id,
        "ACCEPTED",
        "",
        [],
        now,
      ),
    )
  for (const status of [
    "ASSIGNED",
    "IN_PROGRESS",
    "PAUSED",
    "DONE",
    "CLOSED",
    "CANCELLED",
  ] as const)
    assert.throws(() =>
      decide(
        pending.tasks.map((t) => (t.id === "t6" ? { ...t, status } : t)),
        pending.work,
        id,
        "ACCEPTED",
      ),
    )
  const reassigned = pending.tasks.map((t) =>
    t.id === "t6"
      ? { ...t, currentAssignment: { ...t.currentAssignment!, id: "other" } }
      : t,
  )
  assert.throws(
    () => decide(reassigned, pending.work, id, "ACCEPTED"),
    /lượt phân công/,
  )
  const ended = pending.tasks.map((t) =>
    t.id === "t6"
      ? {
          ...t,
          currentAssignment: {
            ...t.currentAssignment!,
            endedAt: now.toISOString(),
          },
        }
      : t,
  )
  assert.throws(() => decide(ended, pending.work, id, "ACCEPTED"))
})

test("Kanban dự án gồm năm cột, gộp PAUSED, không nhân bản hay đưa đóng/hủy vào board", () => {
  const task = initialTasks[0]
  const statuses = [
    "PENDING_ASSIGNMENT",
    "ASSIGNED",
    "IN_PROGRESS",
    "PAUSED",
    "PENDING_REVIEW",
    "DONE",
    "CLOSED",
    "CANCELLED",
  ] as const
  const tasks = statuses.map(
    (status): Task => ({ ...task, id: status, status }),
  )
  const board = projectBoard(tasks)
  assert.equal(board.length, 5)
  assert.deepEqual(
    board.find((c) => c.id === "IN_PROGRESS")!.tasks.map((t) => t.status),
    ["IN_PROGRESS", "PAUSED"],
  )
  assert.equal(board.flatMap((c) => c.tasks).length, 6)
  assert.equal(new Set(board.flatMap((c) => c.tasks).map((t) => t.id)).size, 6)
  const pending = draft()
  const done = decide(
    pending.tasks,
    pending.work,
    pending.work.submissions[0].id,
    "ACCEPTED",
  )
  assert.ok(
    projectBoard(done.tasks)
      .find((c) => c.id === "DONE")!
      .tasks.some((t) => t.id === "t6"),
  )
})
