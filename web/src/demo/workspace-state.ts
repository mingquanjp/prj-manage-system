import { createDemoWorkState } from "./workspace-fixtures.ts"
import type { WorkspaceNotification } from "../models/notification.ts"
import { collectNotifications } from "./notifications.ts"
import type { TaskComment } from "../models/task-collaboration.ts"
import type { AccountsState } from "../models/accounts.ts"
import { createAccountsState } from "./accounts.ts"
import { createEmployeeCvsState } from "./employee-cvs.ts"
import type { EmployeeCvsState } from "../models/employee-cvs.ts"
import { createEmployeeSkillsState } from "./employee-skills.ts"
import type { EmployeeSkillsState } from "../models/employee-skills.ts"
import type { SkillCatalogState } from "../models/skill-catalog.ts"
import { createSkillCatalogState } from "./skill-catalog.ts"
import type { Project, ProjectEvent, Task, TaskGroup } from "../models/workspace.ts"
import { initialGroups, initialTasks, projects } from "./fixtures.ts"
import { createProfileState } from "./profiles.ts"
import type { ProfileState } from "../models/profile.ts"
import type { OrganizationState } from "../models/organization.ts"
import { createOrganizationState } from "./organization.ts"
import { recordAssignmentChanges, type WorkState } from "./effort.ts"

export type DemoWorkspaceState = {
  projects: Project[]
  tasks: Task[]
  groups: TaskGroup[]
  events: ProjectEvent[]
  work: WorkState
  profileState: ProfileState
  organization: OrganizationState
  employeeCvs: EmployeeCvsState
  employeeSkills: EmployeeSkillsState
  skillCatalog: SkillCatalogState
  accounts: AccountsState
  notifications: WorkspaceNotification[]
  comments: TaskComment[]
  notice: string
}

export type DemoWorkspaceAction =
  | { type: "notificationsRead"; notifications: WorkspaceNotification[]; work: WorkState }
  | { type: "taskCommentsChanged"; comments: TaskComment[] }
  | { type: "taskDeadlineChanged"; task: Task }
  | { type: "accountCreated"; accounts: AccountsState; profileState: ProfileState; notice: string }
  | { type: "accountChanged"; accounts: AccountsState; notice: string }
  | { type: "employeeCvsChanged"; employeeCvs: EmployeeCvsState; notice: string }
  | { type: "employeeSkillsChanged"; employeeSkills: EmployeeSkillsState; notice: string }
  | { type: "skillCatalogChanged"; skillCatalog: SkillCatalogState; notice: string }
  | { type: "organizationChanged"; organization: OrganizationState; notice: string }
  | { type: "profilesChanged"; profileState: ProfileState; notice: string }
  | {
      type: "lifecycleChanged"
      project: Project
      tasks: Task[]
      work: WorkState
      event: ProjectEvent
      notice: string
    }
  | { type: "projectCreated"; projects: Project[]; event: ProjectEvent }
  | { type: "projectChanged"; project: Project; event: ProjectEvent; notice: string }
  | { type: "groupsChanged"; groups: TaskGroup[] }
  | { type: "tasksChanged"; tasks: Task[]; at: string }
  | { type: "workChanged"; work: WorkState }
  | { type: "taskTransitioned"; tasks: Task[]; notice: string }
  | { type: "resultSaved"; tasks: Task[]; work: WorkState; notice: string }
  | { type: "assignmentReplaced"; tasks: Task[]; work: WorkState; notice: string }
  | { type: "noticeChanged"; notice: string }
  | { type: "reset" }

export function createDemoWorkspaceState(): DemoWorkspaceState {
  return {
    projects: structuredClone(projects),
    tasks: structuredClone(initialTasks),
    groups: structuredClone(initialGroups),
    events: [],
    comments: [],
    notifications: [],
    work: createDemoWorkState(),
    profileState: createProfileState(),
    organization: createOrganizationState(),
    skillCatalog: createSkillCatalogState(),
    employeeSkills: createEmployeeSkillsState(),
    employeeCvs: createEmployeeCvsState(),
    accounts: createAccountsState(),
    notice: "",
  }
}

// Chỉ áp dụng dữ liệu đã chuẩn bị; không tự lấy giờ, sinh ID hoặc chạy thao tác UI.
function reduceWorkspaceState(
  state: DemoWorkspaceState,
  action: DemoWorkspaceAction,
): DemoWorkspaceState {
  switch (action.type) {
    case "notificationsRead":
      return { ...state, notifications: action.notifications, work: action.work }
    case "taskCommentsChanged":
      return { ...state, comments: action.comments, notice: "Đã gửi trao đổi." }
    case "taskDeadlineChanged":
      return {
        ...state,
        tasks: state.tasks.map((task) => (task.id === action.task.id ? action.task : task)),
        notice: "Đã cập nhật hạn công việc; giữ nguyên ước tính.",
      }
    case "accountCreated":
      return {
        ...state,
        accounts: action.accounts,
        profileState: action.profileState,
        notice: action.notice,
      }
    case "accountChanged":
      return { ...state, accounts: action.accounts, notice: action.notice }
    case "employeeCvsChanged":
      return { ...state, employeeCvs: action.employeeCvs, notice: action.notice }
    case "employeeSkillsChanged":
      return { ...state, employeeSkills: action.employeeSkills, notice: action.notice }
    case "skillCatalogChanged":
      return { ...state, skillCatalog: action.skillCatalog, notice: action.notice }
    case "organizationChanged":
      return { ...state, organization: action.organization, notice: action.notice }
    case "profilesChanged":
      return { ...state, profileState: action.profileState, notice: action.notice }
    case "lifecycleChanged":
      return {
        ...state,
        projects: state.projects.map((p) => (p.id === action.project.id ? action.project : p)),
        tasks: action.tasks,
        work: action.work,
        events: [...state.events, action.event],
        notice: action.notice,
      }
    case "projectCreated":
      return {
        ...state,
        projects: action.projects,
        events: [...state.events, action.event],
        notice: "Đã tạo dự án và giao PM quản lý.",
      }
    case "projectChanged":
      return {
        ...state,
        projects: state.projects.map((p) => (p.id === action.project.id ? action.project : p)),
        events: [
          ...state.events,
          {
            ...action.event,
            before: structuredClone(state.projects.find((p) => p.id === action.project.id)),
          },
        ],
        notice: action.notice,
      }
    case "groupsChanged":
      return { ...state, groups: action.groups }
    case "tasksChanged":
      return {
        ...state,
        tasks: action.tasks,
        work: recordAssignmentChanges(state.work, state.tasks, action.tasks, new Date(action.at)),
      }
    case "workChanged":
      return { ...state, work: action.work }
    case "taskTransitioned":
      return { ...state, tasks: action.tasks, notice: action.notice }
    case "resultSaved":
    case "assignmentReplaced":
      return { ...state, tasks: action.tasks, work: action.work, notice: action.notice }
    case "noticeChanged":
      return { ...state, notice: action.notice }
    case "reset":
      // Khôi phục đúng cùng fixture như lần mở app, gồm lịch sử mẫu và lượt giao.
      return {
        ...createDemoWorkspaceState(),
        notice: "Đã khôi phục dữ liệu mẫu.",
      }
  }
}

// Thông báo đi cùng thay đổi miền trong một lần cập nhật state demo, không phải transaction server.
export function demoWorkspaceReducer(
  state: DemoWorkspaceState,
  action: DemoWorkspaceAction,
): DemoWorkspaceState {
  const next = reduceWorkspaceState(state, action)
  if (action.type === "reset" || action.type === "notificationsRead") return next
  const added = collectNotifications(state, next)
  return added.length ? { ...next, notifications: [...next.notifications, ...added] } : next
}
