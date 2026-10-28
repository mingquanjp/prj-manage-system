import { replaceProjectPm, extendTaskDeadline, addTaskComment } from "./task-collaboration.ts"
import type { PmChangeCommand, TaskDeadlineCommand } from "../models/task-collaboration.ts"
import { createAccountWithProfile, changeAccount } from "./accounts.ts"
import type { AccountCreateInput, AccountCommand } from "../models/accounts.ts"
import { addEmployeeCv, selectCurrentCv, simulateCvProcessing } from "./employee-cvs.ts"
import type { CvFileInput, CvProcessingStatus } from "../models/employee-cvs.ts"
import {
  submitSkillRequest,
  supplementSkillRequest,
  withdrawSkillRequest,
  resolveSkillRequest,
} from "./employee-skills.ts"
import type {
  SkillRequestInput,
  SkillSupplement,
  SkillDecision,
} from "../models/employee-skills.ts"
import type { SkillCatalogCommand } from "../models/skill-catalog.ts"
import { changeSkillCatalog } from "./skill-catalog.ts"
import type { ExecutionAction, Project, Task, TaskGroup, TaskStatus } from "../models/workspace.ts"
import { transitionPersonalTask } from "./execution.ts"
import { initialOutputConflicts } from "./fixtures.ts"
import {
  rememberContributions,
  reviewSubmission,
  submitResult,
  type ReviewResult,
  type WorkState,
} from "./effort.ts"
import type { DemoWorkspaceAction, DemoWorkspaceState } from "./workspace-state.ts"
import { transitionProjectLifecycle, type LifecycleCommand } from "./project-lifecycle.ts"
import { reassignUnstartedTask, type ReassignmentCommand } from "./reassignment.ts"
import { deactivateEmployeeProfile } from "./employee-deactivation.ts"
import { requestProfileChange, resolveProfileRequest, professionalPeople } from "./profiles.ts"
import { updateEmployeeProfile } from "./employee-administration.ts"
import type { EmployeeProfile, EmployeeEditValues } from "../models/profile.ts"
import type { ProfileValues } from "../models/profile.ts"
import type { OrganizationCommand } from "../models/organization.ts"
import { changeOrganization } from "./organization.ts"

// Biên thao tác demo: chuẩn bị kết quả trước dispatch để lỗi vẫn trả về form gọi.
// Đây không phải transaction/API server; callback dùng snapshot render hiện tại.
export function createDemoWorkspaceActions(
  state: DemoWorkspaceState,
  dispatch: (action: DemoWorkspaceAction) => void,
  actorId: string,
) {
  const directory = professionalPeople(
    state.profileState.profiles,
    state.organization,
    state.employeeSkills,
    state.accounts.accounts,
  )
  const identity = { people: directory, accounts: state.accounts.accounts }
  const requireActiveActor = () => {
    if (!directory.find((p) => p.id === actorId)?.active)
      throw new Error(
        "Tài khoản đã bị khóa hoặc hồ sơ đã được vô hiệu hóa; không thể thực hiện nghiệp vụ.",
      )
  }
  const requireActivePerson = (personId: string) => {
    if (!directory.find((p) => p.id === personId)?.active)
      throw new Error("Nhân viên không còn hoạt động. Kiểm tra lại lựa chọn trước khi lưu.")
  }
  return {
    replacePm(command: PmChangeCommand) {
      const current = state.projects.find((p) => p.id === command.projectId)
      if (!current) throw new Error("Không tìm thấy dự án.")
      const project = replaceProjectPm(current, actorId, command, identity)
      dispatch({
        type: "projectChanged",
        project,
        event: {
          projectId: project.id,
          actorId,
          at: new Date().toISOString(),
          action: "Thay PM",
          after: structuredClone(project),
        },
        notice: "Đã thay PM; giữ nguyên PTCM và phân công công việc.",
      })
    },
    extendDeadline(command: TaskDeadlineCommand) {
      const task = state.tasks.find((t) => t.id === command.taskId)
      const project = state.projects.find((p) => p.id === task?.projectId)
      if (!task || !project) throw new Error("Không tìm thấy công việc.")
      dispatch({
        type: "taskDeadlineChanged",
        task: extendTaskDeadline(
          project,
          task,
          actorId,
          command,
          new Date().toISOString(),
          identity,
        ),
      })
    },
    postTaskComment(taskId: string, content: string) {
      const task = state.tasks.find((t) => t.id === taskId)
      const project = state.projects.find((p) => p.id === task?.projectId)
      if (!task || !project) throw new Error("Không tìm thấy công việc.")
      const comments = addTaskComment(
        state.comments,
        project,
        task,
        actorId,
        content,
        crypto.randomUUID(),
        new Date().toISOString(),
        identity,
      )
      dispatch({ type: "taskCommentsChanged", comments })
    },
    createAccount(input: AccountCreateInput) {
      const next = createAccountWithProfile(
        state.accounts,
        state.profileState,
        state.organization,
        actorId,
        input,
        crypto.randomUUID(),
        crypto.randomUUID(),
        crypto.randomUUID(),
        new Date().toISOString(),
      )
      dispatch({
        type: "accountCreated",
        ...next,
        notice: `Đã tạo hồ sơ ${next.profileState.profiles.at(-1)!.employeeCode} · Tên đăng nhập: ${next.accounts.accounts.at(-1)!.username}. Chưa cấp thông tin đăng nhập thật.`,
      })
    },
    changeAccount(command: AccountCommand) {
      const accounts = changeAccount(
        state.accounts,
        state.profileState,
        state.organization,
        actorId,
        command,
        crypto.randomUUID(),
        new Date().toISOString(),
      )
      dispatch({
        type: "accountChanged",
        accounts,
        notice:
          "Đã cập nhật tài khoản trong phiên mô phỏng. Chưa có phiên đăng nhập thật để thu hồi.",
      })
    },
    addCv(employeeId: string, file: CvFileInput) {
      const employeeCvs = addEmployeeCv(
        state.employeeCvs,
        state.profileState.profiles,
        actorId,
        employeeId,
        file,
        crypto.randomUUID(),
        new Date().toISOString(),
        identity,
      )
      dispatch({
        type: "employeeCvsChanged",
        employeeCvs,
        notice: "Đã bổ sung phiên bản CV mô phỏng. CV hiện tại và kỹ năng giữ nguyên.",
      })
    },
    selectCurrentCv(cvId: string) {
      const employeeCvs = selectCurrentCv(
        state.employeeCvs,
        state.profileState.profiles,
        actorId,
        cvId,
        identity,
      )
      dispatch({
        type: "employeeCvsChanged",
        employeeCvs,
        notice: "Đã chọn CV hiện tại trong phiên mô phỏng. Kỹ năng chính thức giữ nguyên.",
      })
    },
    simulateCvProcessing(cvId: string, status: CvProcessingStatus) {
      const employeeCvs = simulateCvProcessing(
        state.employeeCvs,
        state.profileState.profiles,
        actorId,
        cvId,
        status,
        identity,
      )
      dispatch({
        type: "employeeCvsChanged",
        employeeCvs,
        notice: "Đã đổi trạng thái trích xuất mô phỏng; không có xử lý AI thật.",
      })
    },
    requestSkillChange(input: SkillRequestInput) {
      requireActiveActor()
      const employeeSkills = submitSkillRequest(
        state.employeeSkills,
        state.profileState.profiles,
        state.skillCatalog,
        state.work,
        state.tasks,
        actorId,
        input,
        crypto.randomUUID(),
        new Date().toISOString(),
        identity,
      )
      dispatch({
        type: "employeeSkillsChanged",
        employeeSkills,
        notice: "Đã gửi đề nghị kỹ năng. Năng lực chính thức chưa thay đổi.",
      })
    },
    supplementSkillChange(requestId: string, version: number, input: SkillSupplement) {
      requireActiveActor()
      const employeeSkills = supplementSkillRequest(
        state.employeeSkills,
        state.profileState.profiles,
        state.work,
        state.tasks,
        actorId,
        requestId,
        version,
        input,
        new Date().toISOString(),
        identity,
      )
      dispatch({
        type: "employeeSkillsChanged",
        employeeSkills,
        notice: "Đã bổ sung căn cứ cho đề nghị đang chờ.",
      })
    },
    withdrawSkillChange(requestId: string, version: number, reason: string) {
      requireActiveActor()
      const employeeSkills = withdrawSkillRequest(
        state.employeeSkills,
        state.profileState.profiles,
        actorId,
        requestId,
        version,
        reason,
        new Date().toISOString(),
        identity,
      )
      dispatch({
        type: "employeeSkillsChanged",
        employeeSkills,
        notice: "Đã rút đề nghị. Năng lực chính thức giữ nguyên.",
      })
    },
    resolveSkillChange(decision: SkillDecision) {
      requireActiveActor()
      const employeeSkills = resolveSkillRequest(
        state.employeeSkills,
        state.profileState.profiles,
        actorId,
        decision,
        crypto.randomUUID(),
        new Date().toISOString(),
        identity,
      )
      dispatch({
        type: "employeeSkillsChanged",
        employeeSkills,
        notice:
          decision.decision === "APPROVED"
            ? "Đã chấp thuận và cập nhật kỹ năng chính thức."
            : "Đã từ chối. Năng lực chính thức giữ nguyên.",
      })
    },
    changeSkillCatalog(command: SkillCatalogCommand) {
      requireActiveActor()
      const skillCatalog = changeSkillCatalog(
        state.skillCatalog,
        state.profileState.profiles,
        state.tasks,
        actorId,
        command,
        crypto.randomUUID(),
        state.employeeSkills,
        identity,
      )
      dispatch({
        type: "skillCatalogChanged",
        skillCatalog,
        notice: "Đã cập nhật danh mục kỹ năng trong phiên mô phỏng.",
      })
    },
    deactivateEmployee(personId: string) {
      requireActiveActor()
      const profileState = deactivateEmployeeProfile(
        state.profileState,
        state.projects,
        state.tasks,
        actorId,
        personId,
        crypto.randomUUID(),
        new Date().toISOString(),
        identity,
      )
      dispatch({
        type: "profilesChanged",
        profileState,
        notice: "Đã vô hiệu hóa hồ sơ trong phiên mô phỏng. Lịch sử được giữ nguyên.",
      })
    },
    updateEmployeeProfile(before: EmployeeProfile, proposed: EmployeeEditValues) {
      requireActiveActor()
      const profileState = updateEmployeeProfile(
        state.profileState,
        state.organization,
        actorId,
        before,
        proposed,
        crypto.randomUUID(),
        new Date().toISOString(),
        identity,
      )
      dispatch({
        type: "profilesChanged",
        profileState,
        notice: "Đã cập nhật hồ sơ chính thức trong phiên mô phỏng.",
      })
    },
    changeOrganization(command: OrganizationCommand) {
      requireActiveActor()
      const organization = changeOrganization(
        state.organization,
        state.profileState.profiles,
        actorId,
        command,
        command.action === "create" ? crypto.randomUUID() : "",
        identity,
      )
      dispatch({
        type: "organizationChanged",
        organization,
        notice:
          command.action === "delete"
            ? "Đã xóa mục khỏi danh mục."
            : "Đã lưu danh mục. Thông tin liên quan được cập nhật trong phiên mô phỏng.",
      })
    },
    requestProfileChange(before: ProfileValues, proposed: ProfileValues, explanation: string) {
      requireActiveActor()
      const profileState = requestProfileChange(
        state.profileState,
        actorId,
        before,
        proposed,
        explanation,
        crypto.randomUUID(),
        new Date().toISOString(),
        identity,
      )
      dispatch({
        type: "profilesChanged",
        profileState,
        notice: "Đã gửi đề nghị. Thông tin chính thức chỉ thay đổi sau khi HR chấp thuận.",
      })
    },
    resolveProfileRequest(requestId: string, decision: "APPROVED" | "REJECTED", reason: string) {
      requireActiveActor()
      const profileState = resolveProfileRequest(
        state.profileState,
        actorId,
        requestId,
        decision,
        reason,
        new Date().toISOString(),
        identity,
      )
      dispatch({
        type: "profilesChanged",
        profileState,
        notice:
          decision === "APPROVED"
            ? "Đã chấp thuận và cập nhật hồ sơ chính thức."
            : "Đã từ chối yêu cầu; hồ sơ chính thức giữ nguyên.",
      })
    },
    reassignTask(command: ReassignmentCommand) {
      requireActiveActor()
      requireActivePerson(command.personId)
      const next = reassignUnstartedTask(
        state,
        actorId,
        command,
        new Date().toISOString(),
        crypto.randomUUID(),
        identity,
      )
      dispatch({ type: "assignmentReplaced", ...next })
    },
    transitionLifecycle(command: LifecycleCommand) {
      requireActiveActor()
      const next = transitionProjectLifecycle(
        state,
        actorId,
        command,
        new Date().toISOString(),
        identity,
      )
      dispatch({ type: "lifecycleChanged", ...next })
    },
    setNotice(notice: string) {
      dispatch({ type: "noticeChanged", notice })
    },
    changeGroups(groups: TaskGroup[]) {
      requireActiveActor()
      dispatch({ type: "groupsChanged", groups })
    },
    changeTasks(tasks: Task[]) {
      requireActiveActor()
      if (
        tasks.some((task) =>
          task.requirements.some(
            (r) => !state.skillCatalog.skills.some((skill) => skill.id === r.skill),
          ),
        )
      )
        throw new Error(
          "Danh mục kỹ năng đã thay đổi. Kiểm tra lại yêu cầu công việc trước khi lưu.",
        )
      for (const task of tasks) {
        const before = state.tasks.find((t) => t.id === task.id)
        if (
          task.assigneeId &&
          (task.assigneeId !== before?.assigneeId ||
            task.currentAssignment?.id !== before?.currentAssignment?.id)
        )
          requireActivePerson(task.assigneeId)
      }
      dispatch({ type: "tasksChanged", tasks, at: new Date().toISOString() })
    },
    changeWork(work: WorkState) {
      requireActiveActor()
      dispatch({ type: "workChanged", work })
    },
    changeProject(project: Project, action: string) {
      requireActiveActor()
      if (project.pmId !== state.projects.find((p) => p.id === project.id)?.pmId)
        requireActivePerson(project.pmId)
      if (
        project.ptcmId &&
        project.ptcmId !== state.projects.find((p) => p.id === project.id)?.ptcmId
      )
        requireActivePerson(project.ptcmId)
      dispatch({
        type: "projectChanged",
        project,
        event: {
          projectId: project.id,
          actorId,
          at: new Date().toISOString(),
          action,
          after: structuredClone(project),
        },
        notice:
          action === "Chỉ định PTCM"
            ? "Đã cập nhật phụ trách chuyên môn."
            : "Đã bắt đầu triển khai dự án.",
      })
    },
    saveCreatedProject(projects: Project[]) {
      requireActiveActor()
      const created = projects[projects.length - 1]
      requireActivePerson(created.pmId)
      dispatch({
        type: "projectCreated",
        projects,
        event: {
          projectId: created.id,
          actorId,
          at: new Date().toISOString(),
          action: "Tạo dự án",
          after: structuredClone(created),
        },
      })
      return created.id
    },
    transitionTask(taskId: string, expected: TaskStatus, action: ExecutionAction, reason?: string) {
      requireActiveActor()
      const tasks = transitionPersonalTask(
        state.projects,
        state.tasks,
        actorId,
        taskId,
        expected,
        action,
        reason,
        identity,
      )
      dispatch({
        type: "taskTransitioned",
        tasks,
        notice:
          action === "pause"
            ? "Đã tạm dừng công việc và ghi nhận lý do trong phiên mô phỏng."
            : action === "start"
              ? "Đã bắt đầu công việc."
              : "Đã tiếp tục công việc.",
      })
    },
    submitResult(taskId: string, summary: string, links: string) {
      requireActiveActor()
      const next = submitResult(
        state.projects,
        state.tasks,
        state.work,
        actorId,
        taskId,
        summary,
        links,
        undefined,
        identity,
      )
      dispatch({
        ...next,
        type: "resultSaved",
        notice:
          "Đã nộp kết quả trong phiên mô phỏng. Công việc chuyển sang Chờ nghiệm thu, nhật ký khóa sửa trực tiếp.",
      })
    },
    reviewResult(submissionId: string, result: ReviewResult, comment: string) {
      requireActiveActor()
      const next = reviewSubmission(
        state.projects,
        state.tasks,
        state.work,
        actorId,
        submissionId,
        result,
        comment,
        initialOutputConflicts,
        undefined,
        identity,
      )
      dispatch({
        type: "resultSaved",
        tasks: next.tasks,
        work: rememberContributions(next.work, next.tasks),
        notice:
          result === "ACCEPTED"
            ? "Đã nghiệm thu kết quả trong phiên mô phỏng. Tiến độ công việc đã cập nhật; dự án không tự hoàn thành."
            : "Đã yêu cầu chỉnh sửa. Công việc trở lại Đang thực hiện để người được giao sửa và nộp lại.",
      })
    },
    reset() {
      dispatch({ type: "reset" })
    },
  }
}

export type DemoWorkspaceActions = ReturnType<typeof createDemoWorkspaceActions>
