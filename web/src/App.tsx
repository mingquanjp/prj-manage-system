import { NotificationMenu } from "@/features/notifications/Notifications"
import { IdentityProvider } from "@/components/identity"
import { SkillCatalogProvider } from "@/components/skill-catalog"
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
  const directory = professionalPeople(
    state.profileState.profiles,
    state.organization,
    state.employeeSkills,
    state.accounts.accounts,
  )
  const identity = { people: directory, accounts: state.accounts.accounts }
  const active = Boolean(directory.find((p) => p.id === actorId)?.active)
  const navigation = selectWorkspaceNavigation(
    state.projects,
    state.tasks,
    actorId,
    location.pathname,
    directory,
    identity,
  )
  const switchIdentity = (id: string) => {
    setActorId(id)
    actions.setNotice("")
    setCreating(false)
    navigate(
      location.pathname.startsWith("/notifications")
        ? "/notifications"
        : navigation.accountsView
          ? "/accounts"
          : navigation.skillReviewView
            ? "/skill-requests"
            : navigation.skillCatalogView
              ? "/skill-catalog"
              : navigation.profileView
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
    <IdentityProvider value={identity}>
      <SkillCatalogProvider catalog={state.skillCatalog}>
        <PeopleDirectory people={directory}>
          <AppShell
            actorId={actorId}
            notificationControl={
              active ? (
                <NotificationMenu
                  key={actorId}
                  state={state}
                  actorId={actorId}
                  onRead={actions.readNotification}
                />
              ) : null
            }
            projects={state.projects}
            navigation={navigation}
            notice={state.notice}
            onSwitchIdentity={switchIdentity}
            onRequestReset={() => setResetOpen(true)}
            onDismissNotice={() => actions.setNotice("")}
            dialogs={
              <>
                {creating && active && isDirector(actorId, identity) && (
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
                  description="Tài khoản, CV, dự án, công việc, phân công, nhật ký, kết quả, hồ sơ, năng lực chính thức, yêu cầu cập nhật và danh mục tổ chức/kỹ năng trong phiên sẽ được khôi phục về dữ liệu mẫu. Không có dữ liệu máy chủ bị thay đổi."
                >
                  <div className="flex justify-end gap-2">
                    <Button variant="outline" onClick={() => setResetOpen(false)}>
                      Giữ nguyên
                    </Button>
                    <Button
                      onClick={() => {
                        actions.reset()
                        setActorId("khoa")
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
      </SkillCatalogProvider>
    </IdentityProvider>
  )
}
