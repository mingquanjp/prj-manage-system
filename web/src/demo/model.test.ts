import { test } from "node:test"
import assert from "node:assert/strict"
import {
  initialTasks,
  projects,
  people,
  commitDraft,
  loadFor,
  progress,
  recommendations,
  validateDraft,
  type Draft,
  initialGroups,
  isTaskFrozen,
  saveGroup,
  saveTask,
  validateTaskInput,
  type TaskInput,
  createProject,
  validateProjectInput,
  appointPtcm,
  ptcmBlockReason,
  startProject,
  canAssign,
  canViewProject,
  initialOutputConflicts,
  type ProjectInput,
} from "./model.ts"

const project = projects[0]
const projectInput: ProjectInput = {
  code: "CRM-2026",
  name: "Quản lý khách hàng",
  description: "Phạm vi demo",
  pmId: "thanh",
  start: "2026-10-21",
  end: "2026-12-31",
  priority: "Cao",
}
test("Giám đốc tạo kèm PM; không tự nhận trách nhiệm hoặc sửa mẫu", () => {
  assert.throws(() => createProject(projects, "thanh", projectInput))
  const next = createProject(projects, "ha", projectInput)
  const created = next.at(-1)!
  assert.equal(next.length, projects.length + 1)
  assert.equal(projects.length, 2)
  assert.equal(created.status, "PLANNING")
  assert.equal(created.pmId, "thanh")
  assert.equal(created.ptcmId, undefined)
  assert.equal(canAssign(created, "ha"), false)
  assert.equal(canViewProject(created, "khoa"), false)
  assert.equal(canViewProject(created, "ha"), true)
})
test("Tạo dự án kiểm tra mã toàn hệ thống, PM, ngày và ưu tiên", () => {
  assert.ok(
    validateProjectInput(projects, { ...projectInput, code: " KHO-2026 " })
      .code,
  )
  assert.ok(
    validateProjectInput(projects, { ...projectInput, pmId: "missing" }).pmId,
  )
  assert.ok(validateProjectInput(projects, { ...projectInput, name: " " }).name)
  assert.ok(
    validateProjectInput(projects, { ...projectInput, start: "2026-02-30" })
      .start,
  )
  assert.ok(
    validateProjectInput(projects, { ...projectInput, end: "2026-01-01" }).end,
  )
  assert.ok(
    validateProjectInput(projects, {
      ...projectInput,
      priority: "" as ProjectInput["priority"],
    }).priority,
  )
  assert.deepEqual(validateProjectInput(projects, projectInput), {})
})
test("Chỉ PM chỉ định PTCM, kiêm nhiệm hợp lệ; quyền cũ không đi theo người", () => {
  const portal = projects[1]
  assert.throws(() => appointPtcm(portal, "ha", "cuong", initialTasks, []))
  assert.throws(() => appointPtcm(portal, "khoa", "cuong", initialTasks, []))
  const combined = appointPtcm(portal, "thanh", "thanh", initialTasks, [])
  assert.equal(canAssign(combined, "thanh"), true)
  assert.equal(canAssign(combined, "khoa"), false)
  assert.equal(portal.ptcmId, "khoa")
  assert.throws(() =>
    commitDraft(combined, "thanh", initialTasks, {
      p1: {
        personId: "thanh",
        mode: "manual",
        reason: "",
        overrideReason: "",
        hadRun: false,
      },
    }),
  )
  for (const status of ["COMPLETED", "CLOSED", "CANCELLED"] as const)
    assert.throws(() =>
      appointPtcm({ ...portal, status }, "thanh", "cuong", initialTasks, []),
    )
})
test("Chặn PTCM đang giữ task mở hoặc đầu ra cũ; không thu hồi ngầm, không cấm lịch sử DONE", () => {
  const before = structuredClone(initialTasks)
  assert.match(
    ptcmBlockReason(project, "cuong", initialTasks, [])!,
    /Còn công việc/,
  )
  assert.throws(
    () =>
      appointPtcm(
        project,
        "thanh",
        "thanh",
        initialTasks,
        initialOutputConflicts,
      ),
    /đầu ra/,
  )
  assert.equal(
    appointPtcm(project, "thanh", "an", initialTasks, []).ptcmId,
    "an",
  )
  assert.deepEqual(initialTasks, before)
  const resolved = initialTasks.map((t) =>
    t.id === "t7" ? { ...t, status: "DONE" as const } : t,
  )
  assert.equal(
    ptcmBlockReason(project, "thanh", resolved, initialOutputConflicts),
    undefined,
  )
  assert.equal(
    ptcmBlockReason(projects[1], "cuong", initialTasks, initialOutputConflicts),
    undefined,
  )
})
test("PM bắt đầu khi có ASSIGNED cùng dự án; giữ task, không cấp PTCM quyền vòng đời", () => {
  const portal = projects[1]
  const before = structuredClone(initialTasks)
  assert.throws(() => startProject(portal, "khoa", initialTasks))
  assert.throws(() => startProject(portal, "ha", initialTasks))
  assert.throws(() =>
    startProject(
      portal,
      "thanh",
      initialTasks.filter((t) => t.projectId !== portal.id),
    ),
  )
  const next = startProject(portal, "thanh", initialTasks)
  assert.equal(next.status, "IN_PROGRESS")
  assert.equal(portal.status, "PLANNING")
  assert.throws(() => startProject(next, "thanh", initialTasks))
  assert.deepEqual(initialTasks, before)
})
const draft: Draft = {
  t1: {
    personId: "an",
    mode: "manual",
    reason: "",
    overrideReason: "",
    hadRun: false,
  },
}
test("draft chưa tăng tải; xác nhận mới cập nhật, không ghi đè mẫu", () => {
  const before = loadFor("an", initialTasks)
  const result = commitDraft(project, "khoa", initialTasks, draft)
  assert.equal(loadFor("an", result), before + 24)
  assert.equal(initialTasks[0].status, "PENDING_ASSIGNMENT")
})
test("PM-only và PTCM tự nhận không hợp lệ", () => {
  assert.throws(() => commitDraft(project, "thanh", initialTasks, draft))
  assert.throws(() =>
    commitDraft(project, "khoa", initialTasks, {
      t1: { ...draft.t1, personId: "khoa" },
    }),
  )
  assert.equal(
    recommendations(project, initialTasks[0]).some(
      (r) => r.person.id === project.ptcmId,
    ),
    false,
  )
})
test("một dòng sai không làm thay đổi dòng hợp lệ", () => {
  assert.throws(() =>
    commitDraft(project, "khoa", initialTasks, { ...draft, t5: draft.t1 }),
  )
  assert.equal(initialTasks[0].status, "PENDING_ASSIGNMENT")
})
test("manual không cần điểm; thiếu kỹ năng bắt buộc cần lý do", () => {
  const choices = { t1: { ...draft.t1, personId: "cuong" } }
  assert.equal(validateDraft(project, "khoa", initialTasks, choices).length, 1)
  choices.t1.overrideReason = "Có kế hoạch hỗ trợ chuyên môn"
  assert.equal(validateDraft(project, "khoa", initialTasks, choices).length, 0)
  assert.ok(people.length > 0)
})
test("tiến độ theo nghiệm thu, không theo phân công; rỗng không 100%", () => {
  assert.equal(progress(initialTasks.filter((t) => t.projectId === "kho")), 14)
  assert.equal(progress([]), null)
})

const taskInput: TaskInput = {
  code: "KHO-T008",
  name: "Kiểm thử luồng nhập kho",
  groupId: "kho-testing",
  description: "Kiểm thử nhập kho; đầu ra gồm kết quả và danh sách lỗi.",
  criteria: "Các ca hợp lệ và không hợp lệ đều có kết quả.",
  estimate: 8,
  priority: "Trung bình",
  requirements: [{ skill: "TypeScript", level: 2, weight: 1, mandatory: true }],
}

test("nhóm trống tồn tại độc lập; tạo/sửa nhóm không làm mất liên kết task", () => {
  assert.ok(initialGroups.some((g) => g.id === "kho-testing"))
  assert.equal(
    initialTasks.some((t) => t.groupId === "kho-testing"),
    false,
  )
  const result = saveGroup(
    project,
    "khoa",
    initialGroups,
    initialTasks,
    { name: "Dịch vụ dữ liệu", description: "API" },
    "kho-backend",
  )
  assert.equal(result.tasks[0].group, "Dịch vụ dữ liệu")
  assert.equal(result.tasks[0].groupId, initialTasks[0].groupId)
  assert.deepEqual(
    recommendations(project, result.tasks[0]),
    recommendations(project, initialTasks[0]),
  )
  assert.equal(initialTasks[0].group, "Backend")
  assert.throws(() =>
    saveGroup(project, "khoa", initialGroups, initialTasks, {
      name: " Backend ",
      description: "",
    }),
  )
  const otherProject = saveGroup(
    projects[1],
    "khoa",
    initialGroups,
    initialTasks,
    { name: "Backend", description: "" },
    "portal-backend",
  )
  assert.equal(otherProject.tasks[0].group, "Backend")
})

test("tạo task hợp lệ vào nhóm rồi phân công dùng cùng dữ liệu; sửa trước giao giữ ID", () => {
  const created = saveTask(
    project,
    "khoa",
    initialGroups,
    initialTasks,
    taskInput,
  )
  const task = created.at(-1)!
  assert.equal(task.status, "PENDING_ASSIGNMENT")
  assert.equal(task.group, "Kiểm thử")
  assert.equal(task.firstAssignedAt, undefined)
  const edited = saveTask(
    project,
    "khoa",
    initialGroups,
    created,
    { ...taskInput, estimate: 12 },
    task.id,
  )
  assert.equal(edited.at(-1)!.estimate, 12)
  assert.equal(edited.at(-1)!.id, task.id)
  const assigned = commitDraft(project, "khoa", edited, {
    [task.id]: { ...draft.t1, personId: "binh" },
  })
  assert.equal(assigned.at(-1)!.assigneeId, "binh")
  assert.ok(assigned.at(-1)!.firstAssignedAt)
  assert.equal(initialTasks.length + 1, assigned.length)
})

test("form kiểm tra mã, nhóm cùng dự án, ước tính và yêu cầu không trùng", () => {
  assert.deepEqual(
    validateTaskInput(project, initialGroups, initialTasks, taskInput),
    {},
  )
  const invalid = validateTaskInput(project, initialGroups, initialTasks, {
    ...taskInput,
    code: " KHO-T001 ",
    groupId: "portal-backend",
    estimate: 0,
    description: " ",
    criteria: "",
    requirements: [
      { skill: "React", level: 5, weight: NaN, mandatory: true },
      { skill: "React", level: 2, weight: 8, mandatory: false },
    ],
  })
  for (const field of [
    "code",
    "groupId",
    "estimate",
    "description",
    "criteria",
    "skill-0",
    "level-0",
    "weight-0",
  ])
    assert.ok(invalid[field])
  assert.ok(
    validateTaskInput(project, initialGroups, initialTasks, {
      ...taskInput,
      requirements: [],
    }).requirements,
  )
  // Giá trị demo vượt 5 không bị biến thành ràng buộc nghiệp vụ.
  assert.deepEqual(
    validateTaskInput(project, initialGroups, initialTasks, {
      ...taskInput,
      requirements: [{ ...taskInput.requirements[0], weight: 8 }],
    }),
    {},
  )
})

test("ngày task để trống hoặc đủ cặp đúng thứ tự trong dự án", () => {
  assert.deepEqual(
    validateTaskInput(project, initialGroups, initialTasks, {
      ...taskInput,
      start: "2026-10-21",
      end: "2026-10-25",
    }),
    {},
  )
  for (const dates of [
    { start: "2026-10-21" },
    { start: "2026-10-25", end: "2026-10-21" },
    { start: "2026-09-30", end: "2026-10-21" },
    { start: "2026-10-21", end: "2027-01-01" },
    { start: "2026-02-30", end: "2026-10-21" },
  ]) {
    assert.ok(
      validateTaskInput(project, initialGroups, initialTasks, {
        ...taskInput,
        ...dates,
      }).dates,
    )
  }
  const dated = saveTask(project, "khoa", initialGroups, initialTasks, {
    ...taskInput,
    start: "2026-10-21",
    end: "2026-10-25",
  })
  assert.throws(() =>
    saveTask(
      project,
      "khoa",
      initialGroups,
      dated,
      { ...taskInput, start: "2026-10-21", end: "2026-10-26" },
      dated.at(-1)!.id,
    ),
  )
})

test("chỉ PTCM sửa cấu trúc; dự án kết thúc chặn cả task và nhóm", () => {
  assert.throws(() =>
    saveTask(project, "thanh", initialGroups, initialTasks, taskInput),
  )
  assert.throws(() =>
    saveGroup(project, "thanh", initialGroups, initialTasks, {
      name: "Tài liệu",
      description: "",
    }),
  )
  for (const status of ["COMPLETED", "CLOSED", "CANCELLED"] as const) {
    assert.throws(() =>
      saveTask(
        { ...project, status },
        "khoa",
        initialGroups,
        initialTasks,
        taskInput,
      ),
    )
    assert.throws(() =>
      saveGroup({ ...project, status }, "khoa", initialGroups, initialTasks, {
        name: "Tài liệu",
        description: "",
      }),
    )
  }
})

test("sau lần giao đầu, thu hồi không mở khóa đặc tả", () => {
  const assigned = commitDraft(project, "khoa", initialTasks, draft)
  const returned = assigned.map((t) =>
    t.id === "t1"
      ? { ...t, status: "PENDING_ASSIGNMENT" as const, assigneeId: undefined }
      : t,
  )
  assert.equal(isTaskFrozen(returned[0]), true)
  assert.throws(() =>
    saveTask(
      project,
      "khoa",
      initialGroups,
      returned,
      { ...taskInput, code: "KHO-T001" },
      "t1",
    ),
  )
  assert.throws(() =>
    saveTask(project, "khoa", initialGroups, initialTasks, taskInput, "p1"),
  )
})
