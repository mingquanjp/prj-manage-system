import test from "node:test"
import assert from "node:assert/strict"
import {
  createDemoWorkspaceState,
  demoWorkspaceReducer,
  type DemoWorkspaceAction,
} from "./workspace-state.ts"
import { createDemoWorkspaceActions } from "./workspace-actions.ts"
import {
  submitSkillRequest,
  supplementSkillRequest,
  resolveSkillRequest,
  withdrawSkillRequest,
  availableSkillEvidence,
  officialValues,
} from "./employee-skills.ts"
import { professionalPeople } from "./profiles.ts"
import { skillGaps, validateDraft } from "./assignment.ts"
import { personnelProfiles } from "./personnel.ts"
import { changeSkillCatalog, skillReferences } from "./skill-catalog.ts"
import type {
  SkillRequestInput,
  EmployeeSkillsState,
  SkillDecision,
} from "../models/employee-skills.ts"

const at = "2026-10-27T09:00:00.000Z"
function fixture() {
  const state = createDemoWorkspaceState()
  const input: SkillRequestInput = {
    skillId: "React",
    operation: "UPDATE",
    before: officialValues(state.employeeSkills, "dung", "React"),
    proposed: { level: 4, years: 2.5, note: "Tự xử lý công việc chuyên môn" },
    explanation: "Đã tự triển khai và giải thích giải pháp",
    externalEvidence: "Bài thực hành được đánh giá trực tiếp",
    evidence: [],
  }
  const submit = (value = input, skills = state.employeeSkills, actor = "dung", id = "r1") =>
    submitSkillRequest(
      skills,
      state.profileState.profiles,
      state.skillCatalog,
      state.work,
      state.tasks,
      actor,
      value,
      id,
      at,
    )
  const decision: SkillDecision = {
    requestId: "r1",
    version: 1,
    decision: "APPROVED",
    reason: "Đã đối chiếu bài thực hành và nhận xét chuyên môn",
    expertOpinion: "Người hướng dẫn xác nhận phần đóng góp",
  }
  return { state, input, submit, decision }
}

test("pending không đổi official; kiểm tra loại, giá trị, một pending/kỹ năng và nền đã đọc", () => {
  const { state, input, submit } = fixture()
  const before = structuredClone(state)
  const pending = submit()
  assert.equal(pending.official, state.employeeSkills.official)
  assert.equal(pending.requests[0].source, "MANUAL")
  assert.equal(pending.requests[0].evidence.length, 0)
  assert.throws(() => submit(input, pending), /đang chờ/)
  assert.throws(() => submit({ ...input, explanation: "  " }), /giải trình/)
  assert.throws(() => submit({ ...input, skillId: "missing" }), /danh mục/)
  assert.throws(() => submit({ ...input, before: null }), /chính thức đã thay đổi/)
  assert.throws(() => submit({ ...input, operation: "ADD" }), /Loại đề nghị/)
  for (const level of [0, 5, 1.5])
    assert.throws(() => submit({ ...input, proposed: { level, years: null, note: "" } }), /bốn mức/)
  for (const years of [-1, 1000, 0.25, NaN])
    assert.throws(() => submit({ ...input, proposed: { level: 4, years, note: "" } }), /Số năm/)
  assert.equal(
    submit({ ...input, proposed: { level: 4, years: null, note: "" } }).requests[0].proposed?.years,
    null,
  )
  assert.deepEqual(state, before)
})

test("chấp thuận ghi official và quyết định cùng state; HR, tự duyệt, inactive và lý do trống bị chặn", () => {
  const { state, submit, decision } = fixture()
  const pending = submit()
  const approved = resolveSkillRequest(
    pending,
    state.profileState.profiles,
    "ha",
    decision,
    "new",
    at,
  )
  assert.equal(officialValues(approved, "dung", "React")?.level, 4)
  assert.equal(approved.requests[0].resolvedBy, "ha")
  assert.equal(approved.requests[0].expertOpinion, decision.expertOpinion)
  assert.throws(
    () => resolveSkillRequest(pending, state.profileState.profiles, "hr", decision, "new", at),
    /Giám đốc/,
  )
  assert.throws(
    () =>
      resolveSkillRequest(
        pending,
        state.profileState.profiles,
        "ha",
        { ...decision, reason: "" },
        "new",
        at,
      ),
    /lý do/,
  )
  const self: EmployeeSkillsState = {
    ...pending,
    requests: [{ ...pending.requests[0], employeeId: "ha" }],
  }
  assert.throws(
    () => resolveSkillRequest(self, state.profileState.profiles, "ha", decision, "new", at),
    /chính mình/,
  )
  const inactive = state.profileState.profiles.map((p) =>
    p.personId === "ha" ? { ...p, employmentStatus: "INACTIVE" as const } : p,
  )
  assert.throws(
    () => resolveSkillRequest(pending, inactive, "ha", decision, "new", at),
    /hoạt động/,
  )
  assert.throws(
    () => resolveSkillRequest(approved, state.profileState.profiles, "ha", decision, "new", at),
    /không còn chờ/,
  )
})

test("bổ sung tăng version; stale review bị chặn; rút và từ chối giữ official, terminal bất biến", () => {
  const { state, input, submit, decision } = fixture()
  const pending = submit()
  const more = supplementSkillRequest(
    pending,
    state.profileState.profiles,
    state.work,
    state.tasks,
    "dung",
    "r1",
    1,
    { ...input, explanation: "Bổ sung quan sát mới" },
    at,
  )
  assert.equal(more.requests[0].version, 2)
  assert.equal(more.official, pending.official)
  for (const outcome of ["APPROVED", "REJECTED"] as const)
    assert.throws(
      () =>
        resolveSkillRequest(
          more,
          state.profileState.profiles,
          "ha",
          { ...decision, decision: outcome },
          "id",
          at,
        ),
      /bổ sung/,
    )
  assert.throws(
    () =>
      supplementSkillRequest(
        pending,
        state.profileState.profiles,
        state.work,
        state.tasks,
        "an",
        "r1",
        1,
        input,
        at,
      ),
    /chủ đề nghị/,
  )
  const rejected = resolveSkillRequest(
    more,
    state.profileState.profiles,
    "ha",
    { ...decision, version: 2, decision: "REJECTED" },
    "id",
    at,
  )
  assert.equal(rejected.official, pending.official)
  assert.throws(
    () =>
      supplementSkillRequest(
        rejected,
        state.profileState.profiles,
        state.work,
        state.tasks,
        "dung",
        "r1",
        2,
        input,
        at,
      ),
    /đang chờ/,
  )
  const withdrawn = withdrawSkillRequest(
    more,
    state.profileState.profiles,
    "dung",
    "r1",
    2,
    "Chuẩn bị căn cứ khác",
    at,
  )
  assert.equal(withdrawn.requests[0].status, "WITHDRAWN")
  assert.equal(withdrawn.official, pending.official)
  assert.equal(submit(input, withdrawn, "dung", "r2").requests.length, 2)
})

test("nền đổi không ghi đè; chặn gửi REMOVE mới, giữ tương thích lịch sử gỡ và công nhận lại", () => {
  const { state, input, submit, decision } = fixture()
  const pending = submit()
  const changed = {
    ...pending,
    official: pending.official.map((s) =>
      s.employeeId === "dung" && s.skillId === "React" ? { ...s, note: "Dữ liệu mới" } : s,
    ),
  }
  assert.throws(
    () => resolveSkillRequest(changed, state.profileState.profiles, "ha", decision, "id", at),
    /ghi đè/,
  )
  assert.throws(
    () => submit({ ...input, operation: "REMOVE", proposed: null }),
    /chỉ được đề nghị bổ sung hoặc điều chỉnh/,
  )
  // Chỉ dựng bản ghi cũ để bảo vệ tương thích; không tạo qua luồng nhân viên mới.
  const remove: EmployeeSkillsState = {
    ...pending,
    requests: pending.requests.map((r) => ({ ...r, operation: "REMOVE", proposed: null })),
  }
  const removed = resolveSkillRequest(remove, state.profileState.profiles, "ha", decision, "id", at)
  assert.equal(officialValues(removed, "dung", "React"), null)
  const added = submit({ ...input, operation: "ADD", before: null }, removed, "dung", "r2")
  const restored = resolveSkillRequest(
    added,
    state.profileState.profiles,
    "ha",
    { ...decision, requestId: "r2" },
    "fresh-official",
    at,
  )
  assert.equal(
    restored.official.find((s) => s.employeeId === "dung" && s.skillId === "React")?.id,
    "fresh-official",
  )
  assert.equal(restored.requests.length, 2)
})

test("căn cứ nhiều lượt chính chủ, không lộ log hoặc kết quả người khác; bản đã gửi không bị ghi đè", () => {
  const { state, input, submit } = fixture()
  const owned = state.tasks.filter((t) => t.currentAssignment).slice(0, 2)
  assert.equal(owned.length, 2)
  state.work.contributions = owned.map((t, i) => ({
    taskId: t.id,
    projectId: t.projectId,
    taskCode: t.code,
    taskName: t.name,
    assignment: { ...t.currentAssignment!, id: `own-${i}`, employeeId: "dung", endedAt: at },
  }))
  state.work.submissions = owned.map((t, i) => ({
    id: `s-${i}`,
    taskId: t.id,
    assignmentId: `own-${i}`,
    employeeId: "dung",
    submittedAt: at,
    summary: `Kết quả riêng ${i}`,
    links: [],
    logSnapshot: { schemaVersion: 2, logs: [] },
  }))
  state.work.submissions.push({
    ...state.work.submissions[0],
    id: "other",
    employeeId: "an",
    summary: "Không phải công của Dũng",
  })
  const evidence = availableSkillEvidence(state.work, state.tasks, "dung").filter((e) =>
    e.assignmentId.startsWith("own-"),
  )
  assert.equal(evidence.length, 2)
  assert.equal(evidence[0].submission?.summary, "Kết quả riêng 0")
  assert.equal(availableSkillEvidence(state.work, state.tasks, "ha").length, 0)
  assert.ok(!JSON.stringify(evidence).includes("logSnapshot"))
  const pending = submit({
    ...input,
    evidence: evidence.map((e) => ({
      assignmentId: e.assignmentId,
      contributionNote: "Phần riêng của tôi",
    })),
  })
  assert.equal(pending.requests[0].evidence.length, 2)
  state.work.submissions[0].summary = "Sửa kết quả về sau"
  assert.equal(pending.requests[0].evidence[0].submission?.summary, "Kết quả riêng 0")
  assert.throws(
    () => submit({ ...input, evidence: [{ assignmentId: "foreign", contributionNote: "" }] }),
    /đóng góp/,
  )
  assert.throws(
    () =>
      submit({
        ...input,
        evidence: [
          { assignmentId: "own-0", contributionNote: "" },
          { assignmentId: "own-0", contributionNote: "" },
        ],
      }),
    /lặp/,
  )
})

test("official dùng chung tra cứu và đối chiếu phân công; pending không đưa vào năng lực ứng viên", () => {
  const { state, submit, decision } = fixture()
  const pending = submit()
  const before = professionalPeople(state.profileState.profiles, state.organization, pending)
  assert.equal(before.find((p) => p.id === "dung")?.skills.React, 2)
  const approved = resolveSkillRequest(
    pending,
    state.profileState.profiles,
    "ha",
    decision,
    "id",
    at,
  )
  const directory = professionalPeople(state.profileState.profiles, state.organization, approved)
  const lookup = personnelProfiles(
    state.projects,
    state.tasks,
    "khoa",
    "",
    directory,
    state.skillCatalog,
  )
  assert.equal(
    lookup.find((p) => p.id === "dung")?.skills.find((s) => s.name === "React")?.level,
    4,
  )
  const task = {
    ...state.tasks[0],
    status: "PENDING_ASSIGNMENT" as const,
    requirements: [{ skill: "React", level: 4, mandatory: true, weight: 1 }],
  }
  assert.equal(
    skillGaps(
      task,
      before.find((p) => p.id === "dung")!,
    ).length,
    1,
  )
  assert.equal(
    skillGaps(
      task,
      directory.find((p) => p.id === "dung")!,
    ).length,
    0,
  )
  assert.equal(
    validateDraft(
      state.projects[0],
      "khoa",
      [task],
      {
        [task.id]: {
          personId: "dung",
          mode: "manual",
          reason: "",
          overrideReason: "",
          hadRun: false,
        },
      },
      directory,
    ).length,
    0,
  )
})

test("request đã rút vẫn giữ tham chiếu catalog; action/reducer không đổi task/work/profile; reset đồng bộ", () => {
  const { state, input } = fixture()
  state.skillCatalog.skills.push({ id: "python", name: "Python", categoryId: "programming" })
  let current = state
  const dispatch = (action: DemoWorkspaceAction) => {
    current = demoWorkspaceReducer(current, action)
  }
  createDemoWorkspaceActions(current, dispatch, "dung").requestSkillChange({
    ...input,
    skillId: "python",
    operation: "ADD",
    before: null,
  })
  assert.equal(current.tasks, state.tasks)
  assert.equal(current.work, state.work)
  assert.equal(current.profileState, state.profileState)
  const request = current.employeeSkills.requests[0]
  createDemoWorkspaceActions(current, dispatch, "dung").withdrawSkillChange(
    request.id,
    1,
    "Gửi sau",
  )
  assert.ok(
    skillReferences(current.skillCatalog, [], "skills", "python", current.employeeSkills).length,
  )
  assert.throws(
    () =>
      changeSkillCatalog(
        current.skillCatalog,
        current.profileState.profiles,
        [],
        "ha",
        {
          action: "delete",
          kind: "skills",
          id: "python",
          beforeName: "Python",
          beforeCategoryId: "programming",
        },
        "unused",
        current.employeeSkills,
      ),
    /tham chiếu/,
  )
  const reset = demoWorkspaceReducer(current, { type: "reset" })
  assert.deepEqual(reset.employeeSkills, state.employeeSkills)
  assert.ok(availableSkillEvidence(reset.work, reset.tasks, "dung").length > 0)
})
