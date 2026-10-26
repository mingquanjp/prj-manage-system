import { useState } from "react"
import { useLocation, useNavigate } from "react-router"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { isDirector } from "@/demo/identity"
import { CreateProjectDialog } from "@/features/projects/CreateProjectDialog"
import { AppShell } from "@/app/AppShell"
import { WorkspaceRoutes } from "@/app/WorkspaceRoutes"
import { selectWorkspaceNavigation } from "@/app/navigation"
import { useDemoWorkspace } from "@/app/useDemoWorkspace"
import { PeopleDirectory } from "@/components/people-directory"
import { professionalPeople } from "@/demo/profiles"

export default function App() {
  const [actorId, setActorId] = useState("khoa")
  const [creating, setCreating] = useState(false)
  const [resetOpen, setResetOpen] = useState(false)
  const { state, actions } = useDemoWorkspace(actorId)
  const location = useLocation()
  const navigate = useNavigate()
  const navigation = selectWorkspaceNavigation(
    state.projects,
    state.tasks,
    actorId,
    location.pathname,
  )
  const switchIdentity = (id: string) => {
    setActorId(id)
    actions.setNotice("")
    setCreating(false)
    navigate(
      navigation.profileView
        ? "/my-profile"
        : navigation.profileReviewView
          ? "/profile-requests"
          : navigation.employeeProfilesView
            ? "/employee-profiles"
            : navigation.organizationView
              ? "/organization"
              : navigation.personalView
                ? "/my-work"
                : navigation.personnelView
                  ? "/personnel"
                  : "/",
    )
  }
  return (
    <PeopleDirectory people={professionalPeople(state.profileState.profiles, state.organization)}>
      <AppShell
        actorId={actorId}
        projects={state.projects}
        navigation={navigation}
        notice={state.notice}
        onSwitchIdentity={switchIdentity}
        onRequestReset={() => setResetOpen(true)}
        onDismissNotice={() => actions.setNotice("")}
        dialogs={
          <>
            {creating && isDirector(actorId) && (
              <CreateProjectDialog
                projects={state.projects}
                actorId={actorId}
                onClose={() => setCreating(false)}
                onSave={(next) => {
                  const createdId = actions.saveCreatedProject(next)
                  setCreating(false)
                  navigate(`/projects/${createdId}`)
                }}
              />
            )}
            <Dialog
              open={resetOpen}
              onOpenChange={setResetOpen}
              title="Khôi phục dữ liệu mẫu?"
              description="Dự án, công việc, phân công, nhật ký, kết quả, hồ sơ, yêu cầu cập nhật và danh mục phòng ban/vị trí/cấp bậc trong phiên sẽ được khôi phục về dữ liệu mẫu. Không có dữ liệu máy chủ bị thay đổi."
            >
              <div className="flex justify-end gap-2">
                <Button variant="outline" onClick={() => setResetOpen(false)}>
                  Giữ nguyên
                </Button>
                <Button
                  onClick={() => {
                    actions.reset()
                    navigate("/")
                    setResetOpen(false)
                  }}
                >
                  Khôi phục
                </Button>
              </div>
            </Dialog>
          </>
        }
      >
        <WorkspaceRoutes
          actorId={actorId}
          state={state}
          actions={actions}
          navigation={navigation}
          onCreate={() => setCreating(true)}
        />
      </AppShell>
    </PeopleDirectory>
  )
}
