import assert from "node:assert/strict"
import { test } from "node:test"
import {
  catalogSkillName,
  changeSkillCatalog,
  createSkillCatalogState,
  skillReferences,
} from "./skill-catalog.ts"
import {
  createDemoWorkspaceState,
  demoWorkspaceReducer,
  type DemoWorkspaceAction,
} from "./workspace-state.ts"
import { createDemoWorkspaceActions } from "./workspace-actions.ts"
import { selectWorkspaceNavigation } from "../app/navigation.ts"
import { professionalPeople } from "./profiles.ts"
import { personnelProfiles } from "./personnel.ts"
import { saveTask, validateTaskInput } from "./task-structure.ts"
import { recommendations, skillGaps } from "./assignment.ts"

function setup() {
  const state = createDemoWorkspaceState()
  const change = (
    command: Parameters<typeof changeSkillCatalog>[4],
    newId = "new-id",
    actorId = "ha",
  ) =>
    changeSkillCatalog(
      state.skillCatalog,
      state.profileState.profiles,
      state.tasks,
      actorId,
      command,
      newId,
    )
  return { state, change }
}

test("chỉ Giám đốc active quản lý danh mục; tên/nhóm hợp lệ và tên skill duy nhất toàn danh mục", () => {
  const { state, change } = setup()
  for (const actor of ["khoa", "thanh", "hr", "dung"])
    assert.throws(
      () => change({ action: "create", kind: "categories", name: "Khác" }, "new", actor),
      /Giám đốc/,
    )
  state.profileState.profiles.find((p) => p.personId === "ha")!.employmentStatus = "INACTIVE"
  assert.throws(() => change({ action: "create", kind: "categories", name: "Khác" }), /Giám đốc/)
  state.profileState.profiles.find((p) => p.personId === "ha")!.employmentStatus = "ACTIVE"
  for (const name of [" ", "x".repeat(101), " lập trình "])
    assert.throws(() => change({ action: "create", kind: "categories", name }))
  assert.throws(
    () => change({ action: "create", kind: "skills", name: " react ", categoryId: "databases" }),
    /đã có/,
  )
  assert.throws(
    () => change({ action: "create", kind: "skills", name: "Python", categoryId: "missing" }),
    /Chọn nhóm/,
  )
  const added = change({
    action: "create",
    kind: "skills",
    name: " Python ",
    categoryId: "programming",
  })
  assert.deepEqual(added.skills.at(-1), { id: "new-id", name: "Python", categoryId: "programming" })
  assert.deepEqual(state.skillCatalog, createSkillCatalogState())
})

test("không xóa nhóm có skill hoặc skill có năng lực/task kể cả terminal; không lộ chi tiết tham chiếu", () => {
  const { state, change } = setup()
  assert.throws(
    () =>
      change({ action: "delete", kind: "categories", id: "programming", beforeName: "Lập trình" }),
    /tham chiếu/,
  )
  const command = {
    action: "delete",
    kind: "skills",
    id: "React",
    beforeName: "React",
    beforeCategoryId: "programming",
  } as const
  assert.throws(() => change(command), /tham chiếu/)
  assert.ok(
    skillReferences(state.skillCatalog, [], "skills", "React").some((label) =>
      label.includes("hồ sơ"),
    ),
  )
  const tasks = state.tasks.map((task) => ({
    ...task,
    status: "CANCELLED" as const,
    requirements: [{ skill: "custom", level: 1, weight: 1, mandatory: true }],
  }))
  const catalog = {
    ...state.skillCatalog,
    skills: [
      ...state.skillCatalog.skills,
      { id: "custom", name: "Tùy chỉnh", categoryId: "programming" },
    ],
  }
  assert.throws(
    () =>
      changeSkillCatalog(
        catalog,
        state.profileState.profiles,
        tasks,
        "ha",
        {
          action: "delete",
          kind: "skills",
          id: "custom",
          beforeName: "Tùy chỉnh",
          beforeCategoryId: "programming",
        },
        "unused",
      ),
    /tham chiếu/,
  )
  const labels = skillReferences(catalog, tasks, "skills", "custom")
  assert.deepEqual(labels, [`${tasks.length} công việc tham chiếu kỹ năng`])
})

test("thêm/sửa/xóa mục chưa dùng; kiểm tra nền đã đổi và không mất ID/tham chiếu", () => {
  const { state, change } = setup()
  const added = change({ action: "create", kind: "categories", name: "Giao tiếp" })
  const renamed = changeSkillCatalog(
    added,
    state.profileState.profiles,
    state.tasks,
    "ha",
    {
      action: "update",
      kind: "categories",
      id: "new-id",
      beforeName: "Giao tiếp",
      name: "Kỹ năng giao tiếp",
    },
    "unused",
  )
  assert.equal(renamed.categories.at(-1)?.id, "new-id")
  assert.throws(
    () =>
      changeSkillCatalog(
        renamed,
        state.profileState.profiles,
        state.tasks,
        "ha",
        { action: "delete", kind: "categories", id: "new-id", beforeName: "Giao tiếp" },
        "unused",
      ),
    /đã thay đổi/,
  )
  assert.deepEqual(
    changeSkillCatalog(
      renamed,
      state.profileState.profiles,
      state.tasks,
      "ha",
      { action: "delete", kind: "categories", id: "new-id", beforeName: "Kỹ năng giao tiếp" },
      "unused",
    ),
    state.skillCatalog,
  )
  const skill = change({
    action: "create",
    kind: "skills",
    name: "Python",
    categoryId: "programming",
  })
  assert.deepEqual(
    changeSkillCatalog(
      skill,
      state.profileState.profiles,
      state.tasks,
      "ha",
      {
        action: "delete",
        kind: "skills",
        id: "new-id",
        beforeName: "Python",
        beforeCategoryId: "programming",
      },
      "unused",
    ),
    state.skillCatalog,
  )
})

test("đổi tên chỉ thay projection; mức kỹ năng/yêu cầu/gợi ý/lịch sử giữ nguyên, task mới dùng catalog hiện hành", () => {
  const { state, change } = setup()
  const snapshot = structuredClone(state)
  const catalog = change({
    action: "update",
    kind: "skills",
    id: "React",
    beforeName: "React",
    beforeCategoryId: "programming",
    name: "React.js",
    categoryId: "programming",
  })
  const directory = professionalPeople(state.profileState.profiles, state.organization)
  const person = personnelProfiles(
    state.projects,
    state.tasks,
    "khoa",
    "React.js",
    directory,
    catalog,
  ).find((p) => p.id === "cuong")!
  assert.ok(person.skills.some((skill) => skill.name === "React.js" && skill.level === 4))
  assert.equal(catalogSkillName(catalog, "React"), "React.js")
  assert.deepEqual(state, snapshot)
  const task = state.tasks.find((t) => t.requirements.some((r) => r.skill === "React"))!
  assert.deepEqual(
    skillGaps(
      task,
      directory.find((p) => p.id === "an")!,
    ),
    skillGaps(
      snapshot.tasks.find((t) => t.id === task.id)!,
      directory.find((p) => p.id === "an")!,
    ),
  )
  assert.deepEqual(
    recommendations(state.projects[0], task, directory),
    recommendations(snapshot.projects[0], task, directory),
  )
  const added = change(
    { action: "create", kind: "skills", name: "Python", categoryId: "programming" },
    "python-id",
  )
  const input = {
    ...state.tasks[0],
    code: "NEW-TASK",
    requirements: [{ skill: "python-id", level: 2, weight: 1, mandatory: true }],
  }
  assert.deepEqual(
    validateTaskInput(state.projects[0], state.groups, state.tasks, input, undefined, added),
    {},
  )
  const tasks = saveTask(
    state.projects[0],
    "khoa",
    state.groups,
    state.tasks,
    input,
    undefined,
    added,
  )
  assert.equal(tasks.at(-1)?.requirements[0].skill, "python-id")
  assert.throws(
    () =>
      changeSkillCatalog(
        added,
        state.profileState.profiles,
        tasks,
        "ha",
        {
          action: "delete",
          kind: "skills",
          id: "python-id",
          beforeName: "Python",
          beforeCategoryId: "programming",
        },
        "unused",
      ),
    /tham chiếu/,
  )
})

test("action/reducer giữ một nguồn catalog, quyền điều hướng và reset; chặn task dùng skill đã bị xóa", () => {
  let state = createDemoWorkspaceState()
  const before = state
  const dispatch = (action: DemoWorkspaceAction) => {
    state = demoWorkspaceReducer(state, action)
  }
  createDemoWorkspaceActions(state, dispatch, "ha").changeSkillCatalog({
    action: "create",
    kind: "categories",
    name: "Giao tiếp",
  })
  assert.equal(state.skillCatalog.categories.length, before.skillCatalog.categories.length + 1)
  assert.equal(state.tasks, before.tasks)
  assert.equal(state.work, before.work)
  assert.equal(state.profileState, before.profileState)
  assert.throws(
    () =>
      createDemoWorkspaceActions(state, dispatch, "hr").changeSkillCatalog({
        action: "create",
        kind: "categories",
        name: "Khác",
      }),
    /Giám đốc/,
  )
  for (const actorId of ["khoa", "hr"])
    assert.equal(
      selectWorkspaceNavigation(state.projects, state.tasks, actorId, "/skill-catalog")
        .canManageSkills,
      false,
    )
  assert.equal(
    selectWorkspaceNavigation(state.projects, state.tasks, "ha", "/skill-catalog").canManageSkills,
    true,
  )
  const inactive = professionalPeople(
    state.profileState.profiles.map((p) =>
      p.personId === "ha" ? { ...p, employmentStatus: "INACTIVE" } : p,
    ),
    state.organization,
  )
  assert.equal(
    selectWorkspaceNavigation(state.projects, state.tasks, "ha", "/skill-catalog", inactive)
      .canManageSkills,
    false,
  )
  assert.throws(
    () =>
      createDemoWorkspaceActions(state, dispatch, "khoa").changeTasks(
        state.tasks.map((task) => ({
          ...task,
          requirements: [{ skill: "deleted", level: 1, weight: 1, mandatory: true }],
        })),
      ),
    /Danh mục/,
  )
  assert.deepEqual(
    demoWorkspaceReducer(state, { type: "reset" }).skillCatalog,
    createSkillCatalogState(),
  )
})
