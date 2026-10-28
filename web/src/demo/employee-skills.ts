import { defaultIdentity, type IdentitySource } from "./identity.ts"
import { people } from "./fixtures.ts"
import { getPerson, isDirector } from "./identity.ts"
import type { EmployeeProfile } from "../models/profile.ts"
import type { SkillCatalogState } from "../models/skill-catalog.ts"
import type { Task } from "../models/workspace.ts"
import type { WorkState } from "./effort.ts"
import { rememberContributions } from "./effort.ts"
import type {
  EmployeeSkillsState,
  SkillValues,
  SkillRequestInput,
  SkillSupplement,
  SkillDecision,
  SkillEvidence,
} from "../models/employee-skills.ts"

export function createEmployeeSkillsState(): EmployeeSkillsState {
  return {
    // Không bịa người/ngày công nhận cho năng lực fixture cũ.
    official: people.flatMap((p) =>
      Object.entries(p.skills).map(([skillId, level]) => ({
        id: `seed-${p.id}-${skillId}`,
        employeeId: p.id,
        skillId,
        level,
        years: null,
        note: "",
      })),
    ),
    requests: [],
  }
}
export function officialValues(
  state: EmployeeSkillsState,
  employeeId: string,
  skillId: string,
): SkillValues | null {
  const entry = state.official.find((s) => s.employeeId === employeeId && s.skillId === skillId)
  return entry ? { level: entry.level, years: entry.years, note: entry.note } : null
}
export function sameSkillValues(a: SkillValues | null, b: SkillValues | null) {
  return a === null || b === null
    ? a === b
    : a.level === b.level && a.years === b.years && a.note === b.note
}
function requireActive(
  profiles: EmployeeProfile[],
  actorId: string,
  identity: IdentitySource = defaultIdentity,
  isActor = true,
) {
  // Quyền truy cập của người duyệt khác trạng thái hồ sơ của người có đề nghị.
  if (
    (isActor && !getPerson(actorId, identity)?.active) ||
    profiles.find((p) => p.personId === actorId)?.employmentStatus !== "ACTIVE"
  )
    throw new Error("Hồ sơ không còn hoạt động.")
}
// Chỉ trả đóng góp của chính chủ; không ghép log hoặc đầu ra của lượt giao khác.
export function availableSkillEvidence(
  work: WorkState,
  tasks: Task[],
  employeeId: string,
): SkillEvidence[] {
  return rememberContributions(work, tasks)
    .contributions.filter((c) => c.assignment.employeeId === employeeId)
    .flatMap((c) => {
      const task = tasks.find((t) => t.id === c.taskId)
      if (!task) return []
      const submissions = work.submissions.filter(
        (s) =>
          s.assignmentId === c.assignment.id &&
          s.employeeId === employeeId &&
          s.taskId === c.taskId,
      )
      const submission = submissions.at(-1)
      return [
        {
          assignmentId: c.assignment.id,
          taskId: task.id,
          taskCode: c.taskCode,
          taskName: c.taskName,
          description: task.description,
          criteria: task.criteria,
          requirements: task.requirements.map((r) => ({
            skill: r.skill,
            level: r.level,
            mandatory: r.mandatory,
          })),
          contributionNote: "",
          ...(submission
            ? {
                submission: {
                  id: submission.id,
                  summary: submission.summary,
                  links: [...submission.links],
                  submittedAt: submission.submittedAt,
                  ...(submission.review ? { review: { ...submission.review } } : {}),
                },
              }
            : {}),
        },
      ]
    })
}
function captureEvidence(
  input: SkillSupplement,
  work: WorkState,
  tasks: Task[],
  employeeId: string,
) {
  if (!input.explanation.trim()) throw new Error("Nhập giải trình hoặc lý do thay đổi.")
  const available = availableSkillEvidence(work, tasks, employeeId)
  const ids = new Set<string>()
  return input.evidence.map((e) => {
    const match = available.find((a) => a.assignmentId === e.assignmentId)
    if (!match || ids.has(e.assignmentId))
      throw new Error("Căn cứ không thuộc đóng góp của bạn hoặc bị chọn lặp.")
    ids.add(e.assignmentId)
    return { ...match, contributionNote: e.contributionNote.trim() }
  })
}
function normalizeValues(value: SkillValues | null): SkillValues {
  if (!value || !Number.isInteger(value.level) || value.level < 1 || value.level > 4)
    throw new Error("Chọn một trong bốn mức kỹ năng.")
  if (
    value.years !== null &&
    (!Number.isFinite(value.years) ||
      value.years < 0 ||
      value.years > 999.9 ||
      Math.abs(value.years * 10 - Math.round(value.years * 10)) > 0.000001)
  )
    throw new Error("Số năm từ 0 đến 999,9, tối đa một chữ số thập phân; có thể để trống.")
  return { level: value.level, years: value.years, note: value.note.trim() }
}
export function submitSkillRequest(
  state: EmployeeSkillsState,
  profiles: EmployeeProfile[],
  catalog: SkillCatalogState,
  work: WorkState,
  tasks: Task[],
  actorId: string,
  input: SkillRequestInput,
  id: string,
  at: string,
  identity: IdentitySource = defaultIdentity,
): EmployeeSkillsState {
  requireActive(profiles, actorId, identity)
  if (input.operation !== "ADD" && input.operation !== "UPDATE")
    throw new Error("Nhân viên chỉ được đề nghị bổ sung hoặc điều chỉnh kỹ năng.")
  if (!catalog.skills.some((s) => s.id === input.skillId))
    throw new Error("Chọn kỹ năng trong danh mục hiện hành.")
  if (
    state.requests.some(
      (r) => r.employeeId === actorId && r.skillId === input.skillId && r.status === "PENDING",
    )
  )
    throw new Error("Bạn đã có đề nghị đang chờ cho kỹ năng này.")
  const before = officialValues(state, actorId, input.skillId)
  if (!sameSkillValues(input.before, before))
    throw new Error("Năng lực chính thức đã thay đổi. Mở lại biểu mẫu trước khi gửi.")
  if ((input.operation === "ADD" && before) || (input.operation !== "ADD" && !before))
    throw new Error("Loại đề nghị không phù hợp với kỹ năng hiện có.")
  const proposed = normalizeValues(input.proposed)
  if (sameSkillValues(before, proposed)) throw new Error("Chưa có thông tin năng lực thay đổi.")
  const evidence = captureEvidence(input, work, tasks, actorId)
  return {
    ...state,
    requests: [
      ...state.requests,
      {
        ...input,
        id,
        employeeId: actorId,
        source: "MANUAL",
        before,
        proposed,
        explanation: input.explanation.trim(),
        externalEvidence: input.externalEvidence.trim(),
        evidence,
        requestedAt: at,
        status: "PENDING",
        version: 1,
        supplements: [],
      },
    ],
  }
}
export function supplementSkillRequest(
  state: EmployeeSkillsState,
  profiles: EmployeeProfile[],
  work: WorkState,
  tasks: Task[],
  actorId: string,
  requestId: string,
  version: number,
  input: SkillSupplement,
  at: string,
  identity: IdentitySource = defaultIdentity,
): EmployeeSkillsState {
  requireActive(profiles, actorId, identity)
  const request = state.requests.find((r) => r.id === requestId)
  if (!request || request.employeeId !== actorId || request.status !== "PENDING")
    throw new Error("Chỉ chủ đề nghị đang chờ được bổ sung căn cứ.")
  if (request.version !== version)
    throw new Error("Đề nghị đã thay đổi. Đóng và mở lại để kiểm tra.")
  const evidence = captureEvidence(input, work, tasks, actorId)
  if (
    request.explanation === input.explanation.trim() &&
    request.externalEvidence === input.externalEvidence.trim() &&
    JSON.stringify(request.evidence) === JSON.stringify(evidence)
  )
    throw new Error("Chưa có căn cứ thay đổi.")
  return {
    ...state,
    requests: state.requests.map((r) =>
      r.id === requestId
        ? {
            ...r,
            explanation: input.explanation.trim(),
            externalEvidence: input.externalEvidence.trim(),
            evidence,
            version: r.version + 1,
            supplements: [...r.supplements, { at, version: r.version + 1 }],
          }
        : r,
    ),
  }
}
export function withdrawSkillRequest(
  state: EmployeeSkillsState,
  profiles: EmployeeProfile[],
  actorId: string,
  requestId: string,
  version: number,
  reason: string,
  at: string,
  identity: IdentitySource = defaultIdentity,
): EmployeeSkillsState {
  requireActive(profiles, actorId, identity)
  const request = state.requests.find((r) => r.id === requestId)
  if (
    !request ||
    request.employeeId !== actorId ||
    request.status !== "PENDING" ||
    request.version !== version
  )
    throw new Error("Chỉ chủ đề nghị đang chờ và còn đúng phiên bản được rút.")
  if (!reason.trim()) throw new Error("Nhập lý do rút đề nghị.")
  return {
    ...state,
    requests: state.requests.map((r) =>
      r.id === requestId
        ? {
            ...r,
            status: "WITHDRAWN",
            resolvedBy: actorId,
            resolvedAt: at,
            decisionReason: reason.trim(),
          }
        : r,
    ),
  }
}
export function resolveSkillRequest(
  state: EmployeeSkillsState,
  profiles: EmployeeProfile[],
  actorId: string,
  decision: SkillDecision,
  newOfficialId: string,
  at: string,
  identity: IdentitySource = defaultIdentity,
): EmployeeSkillsState {
  requireActive(profiles, actorId, identity)
  if (!isDirector(actorId, identity)) throw new Error("Chỉ Giám đốc được xét duyệt kỹ năng.")
  const request = state.requests.find((r) => r.id === decision.requestId)
  if (!request || request.status !== "PENDING") throw new Error("Đề nghị không còn chờ xét duyệt.")
  if (request.employeeId === actorId)
    throw new Error("Không được xét duyệt đề nghị của chính mình.")
  requireActive(profiles, request.employeeId, identity, false)
  if (request.version !== decision.version)
    throw new Error("Căn cứ đã được bổ sung. Đóng và mở lại đề nghị trước khi quyết định.")
  if (!decision.reason.trim()) throw new Error("Nhập lý do quyết định.")
  let official = state.official
  if (decision.decision === "APPROVED") {
    if (
      !sameSkillValues(request.before, officialValues(state, request.employeeId, request.skillId))
    )
      throw new Error("Năng lực chính thức đã thay đổi. Không thể ghi đè bằng đề nghị cũ.")
    official = state.official.filter(
      (s) => !(s.employeeId === request.employeeId && s.skillId === request.skillId),
    )
    if (request.operation !== "REMOVE") {
      const values = normalizeValues(request.proposed)
      const previous = state.official.find(
        (s) => s.employeeId === request.employeeId && s.skillId === request.skillId,
      )
      official = [
        ...official,
        {
          ...values,
          id: previous?.id ?? newOfficialId,
          employeeId: request.employeeId,
          skillId: request.skillId,
          approvedBy: actorId,
          approvedAt: at,
        },
      ]
    }
  }
  return {
    ...state,
    official,
    requests: state.requests.map((r) =>
      r.id === request.id
        ? {
            ...r,
            status: decision.decision,
            resolvedBy: actorId,
            resolvedAt: at,
            decisionReason: decision.reason.trim(),
            expertOpinion: decision.expertOpinion.trim(),
          }
        : r,
    ),
  }
}
