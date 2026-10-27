import test from "node:test"
import assert from "node:assert/strict"
import {
  addEmployeeCv,
  selectCurrentCv,
  simulateCvProcessing,
  validateCvFile,
  maxCvBytes,
} from "./employee-cvs.ts"
import {
  createDemoWorkspaceState,
  demoWorkspaceReducer,
  type DemoWorkspaceAction,
} from "./workspace-state.ts"
import { createDemoWorkspaceActions } from "./workspace-actions.ts"

const file = { name: "CV mới.pdf", size: 2048, type: "application/pdf" }
const at = "2026-10-28T03:00:00.000Z"

test("nhân viên xác nhận; HR bổ sung không đổi CV hiện tại, giữ tác giả và bản cũ", () => {
  const state = createDemoWorkspaceState()
  const initial = structuredClone(state.employeeCvs)
  const profiles = state.profileState.profiles
  const added = addEmployeeCv(state.employeeCvs, profiles, "dung", "dung", file, "new", at)
  assert.equal(added.versions.find((cv) => cv.isCurrent)?.id, "cv-dung-current")
  assert.equal(added.versions.at(-1)?.uploadedBy, "dung")
  assert.equal(added.versions.at(-1)?.processingStatus, "QUEUED")
  const current = selectCurrentCv(added, profiles, "dung", "new")
  assert.equal(current.versions.filter((cv) => cv.isCurrent).length, 1)
  assert.equal(current.versions.find((cv) => cv.isCurrent)?.id, "new")
  const hrAdded = addEmployeeCv(current, profiles, "hr", "dung", file, "hr-new", at)
  assert.equal(hrAdded.versions.find((cv) => cv.isCurrent)?.id, "new")
  assert.equal(hrAdded.versions.at(-1)?.uploadedBy, "hr")
  assert.throws(() => selectCurrentCv(hrAdded, profiles, "hr", "hr-new"), /chính chủ/)
  const switched = selectCurrentCv(hrAdded, profiles, "dung", "hr-new")
  assert.equal(switched.versions.find((cv) => cv.isCurrent)?.id, "hr-new")
  assert.equal(switched.versions.length, 4)
  assert.deepEqual(state.employeeCvs, initial)
})

test("quyền chính chủ/HR và hồ sơ active kiểm tra tại biên thay đổi", () => {
  const state = createDemoWorkspaceState()
  const profiles = state.profileState.profiles
  for (const actor of ["khoa", "thanh", "ha", "missing"]) {
    assert.throws(() => addEmployeeCv(state.employeeCvs, profiles, actor, "dung", file, "new", at))
    assert.throws(() => selectCurrentCv(state.employeeCvs, profiles, actor, "cv-dung-current"))
    assert.throws(() =>
      simulateCvProcessing(state.employeeCvs, profiles, actor, "cv-dung-current", "FAILED"),
    )
  }
  const inactive = profiles.map((p) =>
    p.personId === "dung" ? { ...p, employmentStatus: "INACTIVE" as const } : p,
  )
  assert.throws(
    () => addEmployeeCv(state.employeeCvs, inactive, "hr", "dung", file, "new", at),
    /hoạt động/,
  )
  assert.throws(
    () => selectCurrentCv(state.employeeCvs, inactive, "dung", "cv-dung-current"),
    /hoạt động/,
  )
  assert.throws(
    () => simulateCvProcessing(state.employeeCvs, inactive, "hr", "cv-dung-current", "FAILED"),
    /hoạt động/,
  )
})

test("kiểm tra sơ bộ định dạng/dung lượng, không giả xác minh nội dung tệp", () => {
  assert.equal(validateCvFile({ ...file, size: maxCvBytes }).sizeBytes, maxCvBytes)
  assert.equal(
    validateCvFile({ name: "CV.DOCX", size: 100, type: "" }).mimeType,
    "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
  )
  for (const size of [0, -1, maxCvBytes + 1, NaN, 1.5])
    assert.throws(() => validateCvFile({ ...file, size }), /dung lượng/)
  for (const name of ["CV.exe", "CV.doc", "CV.pdf.exe", "a".repeat(256) + ".pdf"])
    assert.throws(() => validateCvFile({ ...file, name }), /PDF hoặc DOCX/)
  assert.throws(() => validateCvFile({ ...file, type: "text/html" }), /không khớp/)
})

test("trích xuất mô phỏng độc lập CV hiện tại và năng lực; cùng state và reset", () => {
  let state = createDemoWorkspaceState()
  const before = structuredClone(state)
  const dispatch = (action: DemoWorkspaceAction) => {
    state = demoWorkspaceReducer(state, action)
  }
  createDemoWorkspaceActions(state, dispatch, "hr").addCv("dung", file)
  const id = state.employeeCvs.versions.at(-1)!.id
  for (const status of ["PROCESSING", "FAILED", "SUCCEEDED"] as const) {
    createDemoWorkspaceActions(state, dispatch, "dung").simulateCvProcessing(id, status)
    createDemoWorkspaceActions(state, dispatch, "dung").selectCurrentCv(id)
    assert.equal(state.employeeCvs.versions.find((cv) => cv.id === id)?.processingStatus, status)
    assert.equal(state.employeeCvs.versions.filter((cv) => cv.isCurrent).length, 1)
    assert.equal(state.employeeCvs.versions.find((cv) => cv.isCurrent)?.id, id)
  }
  assert.deepEqual(state.employeeSkills, before.employeeSkills)
  assert.deepEqual(state.profileState, before.profileState)
  assert.deepEqual(state.tasks, before.tasks)
  assert.deepEqual(state.work, before.work)
  assert.throws(
    () => createDemoWorkspaceActions(state, dispatch, "ha").addCv("dung", file),
    /chính chủ hoặc HR/,
  )
  const restored = demoWorkspaceReducer(state, { type: "reset" })
  assert.deepEqual(restored.employeeCvs, before.employeeCvs)
})
