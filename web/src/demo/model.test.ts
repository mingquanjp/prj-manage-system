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
} from "./model.ts"

const project = projects[0]
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
