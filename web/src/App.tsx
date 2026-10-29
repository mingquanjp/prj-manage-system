import { Login } from "@/features/session/Login"
import {
  startDemoSession,
  sessionIssue,
  sessionIssueMessages,
  safeReturnPath,
  type DemoSession,
} from "@/demo/session"
import { NavigationGuard, useNavigationGuard } from "@/app/NavigationGuard"
import { NotificationMenu } from "@/features/notifications/Notifications"
import { IdentityProvider } from "@/components/identity"
import { SkillCatalogProvider } from "@/components/skill-catalog"
import { useState } from "react"
import { Navigate, useLocation, useNavigate } from "react-router"
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
  return (
    <NavigationGuard>
      <SessionWorkspace />
    </NavigationGuard>
  )
}

function SessionWorkspace() {
  const [session, setSession] = useState<DemoSession | null>(null)
  const [sessionEnd, setSessionEnd] = useState<"logout" | "expired" | "reset" | null>(null)
  const actorId = session?.personId ?? "khoa"
  const { requestLeave } = useNavigationGuard()
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
    setSession(startDemoSession(id, state.accounts.accounts, state.profileState.profiles))
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
  const issue = session
    ? sessionIssue(session, state.accounts.accounts, state.profileState.profiles)
    : null
  if (!session || issue) {
    if (location.pathname !== "/login") {
      const target = safeReturnPath(location.pathname + location.search)
      return (
        <Navigate
          to={
            sessionEnd === "logout" || sessionEnd === "reset"
              ? "/login"
              : `/login?returnTo=${encodeURIComponent(target)}`
          }
          replace
        />
      )
    }
    return (
      <Login
        state={state}
        identity={identity}
        message={
          issue
            ? sessionIssueMessages[issue]
            : sessionEnd === "expired"
              ? sessionIssueMessages.expired
              : sessionEnd === "logout"
                ? "Bạn đã đăng xuất."
                : sessionEnd === "reset"
                  ? "Dữ liệu mẫu đã được khôi phục. Hãy chọn danh tính để tiếp tục."
                  : ""
        }
        onLogin={(personId) => {
          setSession(
            startDemoSession(personId, state.accounts.accounts, state.profileState.profiles),
          )
          setSessionEnd(null)
          actions.setNotice("")
          navigate(safeReturnPath(new URLSearchParams(location.search).get("returnTo")), {
            replace: true,
          })
        }}
      />
    )
  }
  if (location.pathname === "/login")
    return (
      <Navigate to={safeReturnPath(new URLSearchParams(location.search).get("returnTo"))} replace />
    )
  const endSession = (expired: boolean) =>
    requestLeave(() => {
      setSession(null)
      setCreating(false)
      setResetOpen(false)
      actions.setNotice("")
      setSessionEnd(expired ? "expired" : "logout")
    })
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
            onSwitchIdentity={(id) =>
              requestLeave(() => {
                try {
                  switchIdentity(id)
                } catch (err) {
                  actions.setNotice((err as Error).message)
                }
              })
            }
            onLogout={() => endSession(false)}
            onExpireSession={() => endSession(true)}
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
                        setSession(null)
                        setSessionEnd("reset")
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
