import { Link, Route, Routes } from "react-router"
import { Button } from "@/components/ui/button"
import { initialOutputConflicts } from "@/demo/fixtures"
import type { DemoWorkspaceActions } from "@/demo/workspace-actions"
import type { DemoWorkspaceState } from "@/demo/workspace-state"
import { PersonalWork } from "@/features/personal-work/PersonalWork"
import { PersonnelDirectory } from "@/features/personnel/PersonnelDirectory"
import { ProjectList } from "@/features/projects/ProjectList"
import { ProjectWorkspace } from "@/features/projects/ProjectWorkspace"
import type { WorkspaceNavigation } from "./navigation"

export function WorkspaceRoutes({
  actorId,
  state,
  actions,
  navigation,
  onCreate,
}: {
  actorId: string
  state: DemoWorkspaceState
  actions: DemoWorkspaceActions
  navigation: WorkspaceNavigation
  onCreate: () => void
}) {
  const { projects, tasks, groups, events, work } = state
  return (
    <Routes>
      <Route
        path="/personnel"
        element={
          <PersonnelDirectory key={actorId} projects={projects} tasks={tasks} actorId={actorId} />
        }
      />
      <Route
        path="/my-work"
        element={
          <PersonalWork
            key={actorId}
            projects={projects}
            tasks={tasks}
            actorId={actorId}
            work={work}
            onWorkChange={actions.changeWork}
            onSubmitResult={actions.submitResult}
            onTransition={actions.transitionTask}
          />
        }
      />
      <Route
        path="/"
        element={
          <ProjectList
            key={actorId}
            projects={navigation.visibleProjects}
            actorId={actorId}
            tasks={tasks}
            onCreate={onCreate}
          />
        }
      />
      <Route
        path="/projects/:projectId"
        element={
          <ProjectWorkspace
            key={`${navigation.currentProject?.id}-${actorId}`}
            actorId={actorId}
            projects={projects}
            conflicts={initialOutputConflicts}
            events={events}
            onProjectChange={actions.changeProject}
            tasks={tasks}
            groups={groups}
            work={work}
            onWorkChange={actions.changeWork}
            onReview={actions.reviewResult}
            onGroupsChange={actions.changeGroups}
            onTasksChange={actions.changeTasks}
            onNotice={actions.setNotice}
          />
        }
      />
      <Route
        path="*"
        element={
          <div>
            <h1 className="text-2xl font-semibold text-balance">Không tìm thấy trang</h1>
            <Button asChild variant="outline" className="mt-4">
              <Link to="/">Về danh sách dự án</Link>
            </Button>
          </div>
        }
      />
    </Routes>
  )
}
