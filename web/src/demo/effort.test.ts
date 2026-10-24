import { test } from "node:test"
import assert from "node:assert/strict"
import {
  initialTasks as assignedTasks,
  projects,
  transitionPersonalTask,
  startProject,
  loadFor,
} from "./model.ts"
import {
  emptyWorkState,
  saveLog,
  deleteLog,
  submitResult,
  dailyMinutes,
  taskLogs,
  vietnamDay,
  type LogInput,
} from "./effort.ts"

const now = new Date("2026-10-21T08:00:00Z")
const input: LogInput = {
  workDate: "2026-10-20",
  minutes: 90,
  description: "Viết API báo cáo",
  overLimitReason: "",
}
const started = () =>
  transitionPersonalTask(projects, assignedTasks, "dung", "t6", "ASSIGNED", "start")

const initialTasks = started()

test("R01 chỉ ghi/sửa/xóa khi đang thực hiện; Bắt đầu/Tiếp tục mở quyền, không sinh giờ", () => {
  const empty = emptyWorkState()
  assert.throws(
    () => saveLog(projects, assignedTasks, empty, "dung", "t6", input, undefined, now),
    /Bắt đầu/,
  )
  const saved = saveLog(projects, initialTasks, empty, "dung", "t6", input, undefined, now)
  assert.equal(empty.logs.length, 0)
  assert.equal(saved.logs[0].employeeId, "dung")
  assert.equal(saved.history.length, 0)
  assert.equal(initialTasks.find((t) => t.id === "t6")!.status, "IN_PROGRESS")
  assert.equal(
    initialTasks.find((t) => t.id === "t6")!.estimate,
    assignedTasks.find((t) => t.id === "t6")!.estimate,
  )
  const paused = transitionPersonalTask(
    projects,
    initialTasks,
    "dung",
    "t6",
    "IN_PROGRESS",
    "pause",
    "Chờ làm tiếp",
  )
  for (const tasks of [assignedTasks, paused]) {
    assert.throws(() => saveLog(projects, tasks, saved, "dung", "t6", input, undefined, now))
    assert.throws(() =>
      saveLog(
        projects,
        tasks,
        saved,
        "dung",
        "t6",
        input,
        { before: saved.logs[0], reason: "Sửa thời gian" },
        now,
      ),
    )
    assert.throws(() =>
      deleteLog(projects, tasks, saved, "dung", "t6", saved.logs[0], "Ghi nhầm", now),
    )
    assert.equal(
      taskLogs(
        saved,
        tasks.find((t) => t.id === "t6")!,
        "dung",
      ).length,
      1,
    )
  }
  const resumed = transitionPersonalTask(projects, paused, "dung", "t6", "PAUSED", "resume")
  assert.equal(saved.logs.length, 1)
  const edited = saveLog(
    projects,
    resumed,
    saved,
    "dung",
    "t6",
    { ...input, minutes: 120 },
    { before: saved.logs[0], reason: "Bổ sung kiểm thử" },
    now,
  )
  const added = saveLog(projects, resumed, edited, "dung", "t6", input, undefined, now)
  assert.equal(added.logs.length, 2)
  const deleted = deleteLog(projects, resumed, added, "dung", "t6", added.logs[1], "Ghi nhầm", now)
  assert.equal(
    taskLogs(
      deleted,
      resumed.find((t) => t.id === "t6")!,
      "dung",
    ).length,
    1,
  )
})

test("ngày Việt Nam, không trước lượt giao/triển khai, không tương lai; bỏ qua hạn kế hoạch và thời điểm Start", () => {
  assert.equal(vietnamDay(new Date("2026-10-20T18:00:00Z")), "2026-10-21")
  for (const day of ["2026-10-15", "2026-10-22", "2026-02-30", "bad"])
    assert.throws(() =>
      saveLog(
        projects,
        initialTasks,
        emptyWorkState(),
        "dung",
        "t6",
        { ...input, workDate: day },
        undefined,
        now,
      ),
    )
  const later = projects.map((p) =>
    p.id === "kho" ? { ...p, executionStartedAt: "2026-10-21T02:00:00Z" } : p,
  )
  assert.throws(() =>
    saveLog(later, initialTasks, emptyWorkState(), "dung", "t6", input, undefined, now),
  )
  const tasks = initialTasks.map((t) => (t.id === "t6" ? { ...t, end: "2026-10-19" } : t))
  assert.equal(
    saveLog(projects, tasks, emptyWorkState(), "dung", "t6", input, undefined, now).logs.length,
    1,
  )
})

test("tổng xuyên task/project: >480 cần lý do, >1440 chặn; phút nguyên, nội dung bắt buộc", () => {
  const own = initialTasks.find((t) => t.id === "t6")!
  const otherProject = { ...projects[0], id: "other" }
  const otherTask = {
    ...own,
    id: "other-task",
    projectId: "other",
    currentAssignment: {
      id: "other-assignment",
      employeeId: "dung",
      assignedAt: own.currentAssignment!.assignedAt,
    },
  }
  const allProjects = [...projects, otherProject],
    allTasks = [...initialTasks, otherTask]
  const saved = saveLog(
    allProjects,
    allTasks,
    emptyWorkState(),
    "dung",
    "t6",
    { ...input, minutes: 480 },
    undefined,
    now,
  )
  assert.throws(
    () =>
      saveLog(
        allProjects,
        allTasks,
        saved,
        "dung",
        otherTask.id,
        { ...input, minutes: 1 },
        undefined,
        now,
      ),
    /8 giờ/,
  )
  const second = saveLog(
    allProjects,
    allTasks,
    saved,
    "dung",
    otherTask.id,
    {
      ...input,
      minutes: 60,
      overLimitReason: "Làm thêm phần kiểm thử thực tế",
    },
    undefined,
    now,
  )
  assert.equal(dailyMinutes(second, "dung", input.workDate), 540)
  assert.throws(
    () =>
      saveLog(
        allProjects,
        allTasks,
        second,
        "dung",
        "t6",
        { ...input, minutes: 901, overLimitReason: "x" },
        undefined,
        now,
      ),
    /24 giờ/,
  )
  for (const minutes of [0, -1, 1.5, NaN])
    assert.throws(() =>
      saveLog(
        projects,
        initialTasks,
        emptyWorkState(),
        "dung",
        "t6",
        { ...input, minutes },
        undefined,
        now,
      ),
    )
  assert.throws(() =>
    saveLog(
      projects,
      initialTasks,
      emptyWorkState(),
      "dung",
      "t6",
      { ...input, description: " " },
      undefined,
      now,
    ),
  )
})

test("sửa/xóa cần lý do và giá trị nền, giữ history và xóa mềm không cộng tổng", () => {
  const one = saveLog(projects, initialTasks, emptyWorkState(), "dung", "t6", input, undefined, now)
  const before = structuredClone(one.logs[0])
  assert.throws(() =>
    saveLog(
      projects,
      initialTasks,
      one,
      "dung",
      "t6",
      { ...input, minutes: 120 },
      { before, reason: " " },
      now,
    ),
  )
  const two = saveLog(
    projects,
    initialTasks,
    one,
    "dung",
    "t6",
    { ...input, minutes: 120 },
    { before, reason: "Bổ sung thời gian kiểm thử" },
    now,
  )
  assert.equal(two.history[0].before.minutes, 90)
  assert.equal(two.logs[0].minutes, 120)
  assert.equal(one.logs[0].minutes, 90)
  assert.throws(
    () =>
      saveLog(
        projects,
        initialTasks,
        two,
        "dung",
        "t6",
        input,
        { before, reason: "Màn hình cũ" },
        now,
      ),
    /đã thay đổi/,
  )
  assert.throws(() => deleteLog(projects, initialTasks, two, "dung", "t6", two.logs[0], " ", now))
  const deleted = deleteLog(projects, initialTasks, two, "dung", "t6", two.logs[0], "Ghi nhầm", now)
  assert.equal(deleted.logs.length, 1)
  assert.equal(deleted.logs[0].minutes, 120)
  assert.equal(dailyMinutes(deleted, "dung", input.workDate), 0)
  assert.equal(deleted.history.length, 2)
  assert.throws(() =>
    deleteLog(projects, initialTasks, deleted, "dung", "t6", deleted.logs[0], "x", now),
  )
})

test("chặn sai người, PTCM, lượt giao khác, project khóa và task khóa", () => {
  for (const actor of ["an", "khoa", "unknown"])
    assert.throws(() =>
      saveLog(projects, initialTasks, emptyWorkState(), actor, "t6", input, undefined, now),
    )
  for (const status of ["PLANNING", "COMPLETED", "CLOSED", "CANCELLED"] as const)
    assert.throws(() =>
      saveLog(
        projects.map((p) => (p.id === "kho" ? { ...p, status } : p)),
        initialTasks,
        emptyWorkState(),
        "dung",
        "t6",
        input,
        undefined,
        now,
      ),
    )
  for (const status of ["PENDING_REVIEW", "DONE", "CLOSED", "CANCELLED"] as const)
    assert.throws(() =>
      saveLog(
        projects,
        initialTasks.map((t) => (t.id === "t6" ? { ...t, status } : t)),
        emptyWorkState(),
        "dung",
        "t6",
        input,
        undefined,
        now,
      ),
    )
  const state = saveLog(
    projects,
    initialTasks,
    emptyWorkState(),
    "dung",
    "t6",
    input,
    undefined,
    now,
  )
  const reassigned = initialTasks.map((t) =>
    t.id === "t6"
      ? {
          ...t,
          currentAssignment: { ...t.currentAssignment!, id: "different" },
        }
      : t,
  )
  assert.equal(
    taskLogs(
      state,
      reassigned.find((t) => t.id === "t6")!,
      "dung",
    ).length,
    0,
  )
  assert.throws(() => deleteLog(projects, reassigned, state, "dung", "t6", state.logs[0], "x", now))
})

test("nộp thiếu nhật ký hợp lệ; chỉ IN_PROGRESS, chỉ chính chủ; link an toàn; không nộp lặp", () => {
  assert.throws(() =>
    submitResult(projects, assignedTasks, emptyWorkState(), "dung", "t6", "Xong", "", now),
  )
  const tasks = started()
  for (const links of [
    "javascript:alert(1)",
    "file:///tmp/a",
    "https://user:pass@example.com",
    "not-url",
  ])
    assert.throws(() =>
      submitResult(projects, tasks, emptyWorkState(), "dung", "t6", "Xong", links, now),
    )
  assert.throws(() => submitResult(projects, tasks, emptyWorkState(), "an", "t6", "Xong", "", now))
  assert.throws(() => submitResult(projects, tasks, emptyWorkState(), "dung", "t6", " ", "", now))
  const next = submitResult(
    projects,
    tasks,
    emptyWorkState(),
    "dung",
    "t6",
    " Hoàn thành API ",
    "",
    now,
  )
  assert.equal(next.tasks.find((t) => t.id === "t6")!.status, "PENDING_REVIEW")
  assert.equal(tasks.find((t) => t.id === "t6")!.status, "IN_PROGRESS")
  assert.equal(loadFor("dung", next.tasks), loadFor("dung", tasks))
  assert.equal(next.work.submissions[0].logSnapshot.logs.length, 0)
  assert.throws(() => submitResult(projects, next.tasks, next.work, "dung", "t6", "Xong", "", now))
})

test("snapshot chỉ log hiệu lực đúng lượt; bất biến với nhật ký hiện tại, khóa sửa khi chờ", () => {
  const tasks = started(),
    initial = emptyWorkState()
  const logged = saveLog(projects, tasks, initial, "dung", "t6", input, undefined, now)
  const removed = saveLog(
    projects,
    tasks,
    logged,
    "dung",
    "t6",
    { ...input, description: "Mục sẽ xóa" },
    undefined,
    now,
  )
  const state = deleteLog(projects, tasks, removed, "dung", "t6", removed.logs[1], "Ghi nhầm", now)
  const next = submitResult(
    projects,
    tasks,
    state,
    "dung",
    "t6",
    "Đã làm theo tiêu chí",
    "https://example.com/pr/1",
    now,
  )
  assert.equal(next.work.submissions[0].logSnapshot.logs.length, 1)
  assert.throws(() => saveLog(projects, next.tasks, next.work, "dung", "t6", input, undefined, now))
  assert.throws(() =>
    deleteLog(projects, next.tasks, next.work, "dung", "t6", state.logs[0], "x", now),
  )
  state.logs[0].description = "Thay đổi nguồn sau khi chụp"
  assert.equal(next.work.submissions[0].logSnapshot.logs[0].description, input.description)
  const portal = startProject(projects[1], "thanh", initialTasks)
  assert.ok(portal.executionStartedAt)
})
