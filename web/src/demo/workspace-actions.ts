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
import { requestProfileChange, resolveProfileRequest } from "./profiles.ts"
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
  return {
    changeOrganization(command: OrganizationCommand) {
      const organization = changeOrganization(
        state.organization,
        state.profileState.profiles,
        actorId,
        command,
        command.action === "create" ? crypto.randomUUID() : "",
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
      const profileState = requestProfileChange(
        state.profileState,
        actorId,
        before,
        proposed,
        explanation,
        crypto.randomUUID(),
        new Date().toISOString(),
      )
      dispatch({
        type: "profilesChanged",
        profileState,
        notice: "Đã gửi đề nghị. Thông tin chính thức chỉ thay đổi sau khi HR chấp thuận.",
      })
    },
    resolveProfileRequest(requestId: string, decision: "APPROVED" | "REJECTED", reason: string) {
      const profileState = resolveProfileRequest(
        state.profileState,
        actorId,
        requestId,
        decision,
        reason,
        new Date().toISOString(),
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
      const next = reassignUnstartedTask(
        state,
        actorId,
        command,
        new Date().toISOString(),
        crypto.randomUUID(),
      )
      dispatch({ type: "assignmentReplaced", ...next })
    },
    transitionLifecycle(command: LifecycleCommand) {
      const next = transitionProjectLifecycle(state, actorId, command, new Date().toISOString())
      dispatch({ type: "lifecycleChanged", ...next })
    },
    setNotice(notice: string) {
      dispatch({ type: "noticeChanged", notice })
    },
    changeGroups(groups: TaskGroup[]) {
      dispatch({ type: "groupsChanged", groups })
    },
    changeTasks(tasks: Task[]) {
      dispatch({ type: "tasksChanged", tasks, at: new Date().toISOString() })
    },
    changeWork(work: WorkState) {
      dispatch({ type: "workChanged", work })
    },
    changeProject(project: Project, action: string) {
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
      const created = projects[projects.length - 1]
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
      const tasks = transitionPersonalTask(
        state.projects,
        state.tasks,
        actorId,
        taskId,
        expected,
        action,
        reason,
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
      const next = submitResult(
        state.projects,
        state.tasks,
        state.work,
        actorId,
        taskId,
        summary,
        links,
      )
      dispatch({
        ...next,
        type: "resultSaved",
        notice:
          "Đã nộp kết quả trong phiên mô phỏng. Công việc chuyển sang Chờ nghiệm thu, nhật ký khóa sửa trực tiếp.",
      })
    },
    reviewResult(submissionId: string, result: ReviewResult, comment: string) {
      const next = reviewSubmission(
        state.projects,
        state.tasks,
        state.work,
        actorId,
        submissionId,
        result,
        comment,
        initialOutputConflicts,
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
