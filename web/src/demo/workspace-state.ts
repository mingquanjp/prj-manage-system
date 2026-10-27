import type { SkillCatalogState } from "../models/skill-catalog.ts"
import { createSkillCatalogState } from "./skill-catalog.ts"
import type { Project, ProjectEvent, Task, TaskGroup } from "../models/workspace.ts"
import { initialGroups, initialTasks, projects } from "./fixtures.ts"
import { createProfileState } from "./profiles.ts"
import type { ProfileState } from "../models/profile.ts"
import type { OrganizationState } from "../models/organization.ts"
import { createOrganizationState } from "./organization.ts"
import {
  emptyWorkState,
  recordAssignmentChanges,
  rememberContributions,
  type WorkState,
} from "./effort.ts"

export type DemoWorkspaceState = {
  projects: Project[]
  tasks: Task[]
  groups: TaskGroup[]
  events: ProjectEvent[]
  work: WorkState
  profileState: ProfileState
  organization: OrganizationState
  skillCatalog: SkillCatalogState
  notice: string
}

export type DemoWorkspaceAction =
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
    work: rememberContributions(emptyWorkState(), initialTasks),
    profileState: createProfileState(),
    organization: createOrganizationState(),
    skillCatalog: createSkillCatalogState(),
    notice: "",
  }
}

// Chỉ áp dụng dữ liệu đã chuẩn bị; không tự lấy giờ, sinh ID hoặc chạy thao tác UI.
export function demoWorkspaceReducer(
  state: DemoWorkspaceState,
  action: DemoWorkspaceAction,
): DemoWorkspaceState {
  switch (action.type) {
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
      // Giữ hành vi khôi phục trước refactor: dữ liệu effort trong phiên được xóa sạch.
      return {
        ...createDemoWorkspaceState(),
        work: emptyWorkState(),
        notice: "Đã khôi phục dữ liệu mẫu.",
      }
  }
}
