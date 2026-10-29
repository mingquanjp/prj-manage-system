import { useIdentity } from "@/components/identity"
import { usePeople } from "@/components/people-directory"
import type { ReactNode } from "react"
import { Link, NavLink } from "react-router"
import {
  ArchiveIcon,
  ChevronDownIcon,
  ChevronRightIcon,
  CubeIcon,
  DashboardIcon,
  MixerHorizontalIcon,
  ResetIcon,
  ListBulletIcon,
  PersonIcon,
} from "@radix-ui/react-icons"
import { Avatar } from "@/components/shared"
import { Button } from "@/components/ui/button"
import { isDirector } from "@/demo/identity"
import type { Project } from "@/models/workspace"
import { cn } from "@/lib/utils"
import { demoIdentityOptions, type WorkspaceNavigation } from "./navigation"

export function AppShell({
  actorId,
  projects,
  navigation,
  notice,
  onSwitchIdentity,
  onRequestReset,
  onLogout,
  onExpireSession,
  onDismissNotice,
  children,
  dialogs,
  notificationControl,
}: {
  actorId: string
  projects: Project[]
  navigation: WorkspaceNavigation
  notice: string
  onSwitchIdentity: (id: string) => void
  onLogout: () => void
  onExpireSession: () => void
  onRequestReset: () => void
  onDismissNotice: () => void
  children: ReactNode
  dialogs: ReactNode
  notificationControl: ReactNode
}) {
  const identity = useIdentity()

  const { people, getPerson } = usePeople()
  const {
    accountsView,
    canManageAccounts,
    skillReviewView,
    skillCatalogView,
    canManageSkills,
    ownTasks,
    personalView,
    personnelView,
    profileView,
    profileReviewView,
    organizationView,
    employeeProfilesView,
    canManageOrganization,
    canReviewProfile,
    canLookup,
    visibleProjects,
    currentProject,
  } = navigation
  const person = getPerson(actorId)!
  const demoIdentities = demoIdentityOptions(projects, currentProject, people, identity)
  const identityOptions = demoIdentities.map((identity) => (
    <option key={identity.id} value={identity.id}>
      {identity.name} · {identity.responsibility}
    </option>
  ))
  const identityResponsibility = demoIdentities.find(
    (identity) => identity.id === actorId,
  )?.responsibility

  return (
    <div className="min-h-dvh bg-background text-foreground lg:grid lg:grid-cols-[224px_minmax(0,1fr)]">
      <aside className="hidden flex-col overflow-y-auto border-r border-border bg-sidebar lg:sticky lg:top-0 lg:flex lg:h-dvh">
        <Link to="/" className="flex h-[76px] items-center gap-3 px-5">
          <span className="grid size-9 place-items-center rounded-lg bg-primary text-white">
            <ArchiveIcon className="size-5" />
          </span>
          <div>
            <strong className="block text-base">TaskAssign</strong>
            <span className="text-xs text-muted-foreground">Không gian làm việc</span>
          </div>
        </Link>
        <div className="px-3 pt-5">
          <div className="px-3 pb-3 text-xs font-medium tracking-wider text-muted-foreground uppercase">
            Không gian dự án
          </div>
          <NavLink
            to="/"
            end
            className={({ isActive }) =>
              cn(
                "flex h-10 items-center gap-3 rounded-md px-3 text-sm font-medium",
                isActive || currentProject ? "bg-accent" : "hover:bg-accent",
              )
            }
          >
            <DashboardIcon />
            {isDirector(actorId, identity) ? "Dự án" : "Dự án của tôi"}
            <span className="ml-auto text-muted-foreground">{visibleProjects.length}</span>
          </NavLink>
          <nav aria-label="Dự án" className="mt-3 space-y-1 border-l border-border ml-5 pl-3">
            {visibleProjects.map((project) => (
              <NavLink
                key={project.id}
                to={`/projects/${project.id}`}
                className={({ isActive }) =>
                  cn(
                    "flex items-center gap-2 rounded-md px-2 py-2.5 text-sm",
                    isActive
                      ? "bg-accent text-foreground"
                      : "text-muted-foreground hover:bg-accent",
                  )
                }
              >
                <CubeIcon className="shrink-0" />
                <span className="truncate">{project.name}</span>
              </NavLink>
            ))}
          </nav>
          <NavLink
            to="/my-work"
            className={({ isActive }) =>
              cn(
                "mt-3 flex h-10 items-center gap-3 rounded-md px-3 text-sm font-medium",
                isActive ? "bg-accent" : "hover:bg-accent",
              )
            }
          >
            <ListBulletIcon />
            Công việc của tôi
            <span className="ml-auto text-muted-foreground">{ownTasks.length}</span>
          </NavLink>
          {canLookup && (
            <NavLink
              to="/personnel"
              className={({ isActive }) =>
                cn(
                  "mt-3 flex h-10 items-center gap-3 rounded-md px-3 text-sm font-medium",
                  isActive ? "bg-accent" : "hover:bg-accent",
                )
              }
            >
              <PersonIcon />
              Tra cứu nhân sự
            </NavLink>
          )}
          <div className="mt-5 border-t border-border pt-3">
            <p className="px-3 pb-2 text-xs font-medium text-muted-foreground">Cá nhân</p>
            <NavLink
              to="/my-profile"
              className={({ isActive }) =>
                cn(
                  "flex min-h-10 items-center gap-3 rounded-md px-3 py-2 text-sm font-medium",
                  isActive ? "bg-accent" : "hover:bg-accent",
                )
              }
            >
              <PersonIcon /> Hồ sơ của tôi
            </NavLink>
          </div>
          {(canManageAccounts || canReviewProfile || canManageOrganization || canManageSkills) && (
            <div className="mt-4 border-t border-border pt-3">
              <p className="px-3 pb-2 text-xs font-medium text-muted-foreground">Quản trị</p>
              {canManageAccounts && (
                <NavLink
                  to="/accounts"
                  className={({ isActive }) =>
                    cn(
                      "flex min-h-10 items-center rounded-md px-3 py-2 text-sm font-medium",
                      isActive ? "bg-accent" : "hover:bg-accent",
                    )
                  }
                >
                  Quản lý tài khoản
                </NavLink>
              )}
              {canReviewProfile && (
                <NavLink
                  to="/profile-requests"
                  className={({ isActive }) =>
                    cn(
                      "flex min-h-10 items-center rounded-md px-3 py-2 text-sm font-medium",
                      isActive ? "bg-accent" : "hover:bg-accent",
                    )
                  }
                >
                  Yêu cầu thông tin cá nhân
                </NavLink>
              )}
              {canReviewProfile && (
                <NavLink
                  to="/employee-profiles"
                  className={({ isActive }) =>
                    cn(
                      "flex min-h-10 items-center rounded-md px-3 py-2 text-sm font-medium",
                      isActive ? "bg-accent" : "hover:bg-accent",
                    )
                  }
                >
                  Hồ sơ nhân sự
                </NavLink>
              )}
              {canManageSkills && (
                <NavLink
                  to="/skill-requests"
                  className={({ isActive }) =>
                    cn(
                      "flex min-h-10 items-center rounded-md px-3 py-2 text-sm font-medium",
                      isActive ? "bg-accent" : "hover:bg-accent",
                    )
                  }
                >
                  Xét duyệt kỹ năng
                </NavLink>
              )}
              {canManageSkills && (
                <NavLink
                  to="/skill-catalog"
                  className={({ isActive }) =>
                    cn(
                      "flex min-h-10 items-center rounded-md px-3 py-2 text-sm font-medium",
                      isActive ? "bg-accent" : "hover:bg-accent",
                    )
                  }
                >
                  Danh mục kỹ năng
                </NavLink>
              )}
              {canManageOrganization && (
                <NavLink
                  to="/organization"
                  className={({ isActive }) =>
                    cn(
                      "flex min-h-10 items-center rounded-md px-3 py-2 text-sm font-medium",
                      isActive ? "bg-accent" : "hover:bg-accent",
                    )
                  }
                >
                  Danh mục tổ chức
                </NavLink>
              )}
            </div>
          )}
        </div>
        <div className="mx-5 mt-8 border-t border-border pt-5 text-sm leading-6 text-muted-foreground">
          <p className="font-medium text-foreground">Bản duyệt thiết kế</p>
          <p className="mt-1">
            Dự án, công việc, hồ sơ cá nhân và yêu cầu cập nhật. Dữ liệu chỉ trong phiên mô phỏng.
          </p>
        </div>
        <div className="mt-auto px-4 pb-5 pt-8">
          <details className="mb-4 rounded-md border border-border bg-white p-3">
            <summary className="flex cursor-pointer list-none items-center justify-between gap-2 text-xs font-medium">
              <span className="flex items-center gap-2">
                <MixerHorizontalIcon />
                Điều khiển prototype
              </span>
              <ChevronDownIcon />
            </summary>
            <label className="field-label mt-3 block" htmlFor="identity">
              Danh tính mô phỏng
            </label>
            <select
              id="identity"
              className="field mt-1 w-full text-sm"
              value={actorId}
              title={identityResponsibility}
              onChange={(e) => onSwitchIdentity(e.target.value)}
            >
              {identityOptions}
            </select>
            <p className="mt-2 text-xs leading-5 text-foreground">{identityResponsibility}</p>
            <p className="mt-2 text-xs leading-5 text-muted-foreground">
              Chỉ kiểm thử giao diện, không phải tính năng tự cấp quyền.
            </p>
            <Button size="sm" variant="ghost" className="mt-2 w-full" onClick={onExpireSession}>
              Mô phỏng hết phiên
            </Button>
            <Button size="sm" variant="ghost" className="mt-2 w-full" onClick={onRequestReset}>
              <ResetIcon />
              Khôi phục dữ liệu mẫu
            </Button>
          </details>
          <div className="flex items-center gap-3">
            <Avatar initials={person.initials} />
            <div className="min-w-0">
              <p className="truncate text-sm font-medium">{person.name}</p>
              <p className="text-xs text-muted-foreground">{identityResponsibility}</p>
            </div>
          </div>
        </div>
      </aside>
      <div className="min-w-0">
        <header className="flex h-14 items-center justify-between gap-3 border-b border-border bg-sidebar px-6 text-sm lg:px-8">
          <nav
            aria-label="Đường dẫn"
            className="flex min-w-0 items-center gap-2 text-muted-foreground"
          >
            <Link
              to={
                accountsView
                  ? "/accounts"
                  : skillReviewView
                    ? "/skill-requests"
                    : skillCatalogView
                      ? "/skill-catalog"
                      : employeeProfilesView
                        ? "/employee-profiles"
                        : profileView
                          ? "/my-profile"
                          : profileReviewView
                            ? "/profile-requests"
                            : organizationView
                              ? "/organization"
                              : personnelView
                                ? "/personnel"
                                : personalView
                                  ? "/my-work"
                                  : "/"
              }
              className="hover:text-foreground"
            >
              {accountsView
                ? "Quản lý tài khoản"
                : skillReviewView
                  ? "Xét duyệt kỹ năng"
                  : skillCatalogView
                    ? "Danh mục kỹ năng"
                    : employeeProfilesView
                      ? "Hồ sơ nhân sự"
                      : profileView
                        ? "Hồ sơ của tôi"
                        : profileReviewView
                          ? "Yêu cầu thông tin cá nhân"
                          : organizationView
                            ? "Danh mục tổ chức"
                            : personnelView
                              ? "Tra cứu nhân sự"
                              : personalView
                                ? "Không gian cá nhân"
                                : "Dự án"}
            </Link>
            {personalView && (
              <>
                <ChevronRightIcon className="shrink-0" />
                <span className="truncate text-foreground">Công việc của tôi</span>
              </>
            )}
            {currentProject && (
              <>
                <ChevronRightIcon className="shrink-0" />
                <span className="truncate text-foreground">{currentProject.name}</span>
              </>
            )}
          </nav>
          <div className="flex shrink-0 items-center gap-2">
            {notificationControl}
            <Button size="sm" variant="ghost" onClick={onLogout}>
              Đăng xuất
            </Button>
            <span className="hidden rounded border border-border px-2 py-1 text-xs text-muted-foreground sm:inline">
              Dữ liệu mô phỏng
            </span>
          </div>
        </header>
        <div className="flex items-center justify-between gap-3 border-b border-border bg-sidebar px-4 py-3 lg:hidden">
          <Link to="/" className="font-semibold">
            TaskAssign
          </Link>
          <label className="flex items-center gap-2 text-xs text-muted-foreground">
            <span className="field-label">Mô phỏng</span>
            <select
              aria-label="Danh tính mô phỏng trên màn hình nhỏ"
              className="field max-w-48"
              value={actorId}
              title="Điều khiển prototype, không phải tính năng tự cấp quyền"
              onChange={(e) => onSwitchIdentity(e.target.value)}
            >
              {identityOptions}
            </select>
          </label>
        </div>
        <div className="px-4 pb-2 lg:hidden">
          <Button size="sm" variant="ghost" onClick={onExpireSession}>
            Prototype: mô phỏng hết phiên
          </Button>
        </div>
        <nav
          aria-label="Điều hướng trên màn hình nhỏ"
          className="flex flex-wrap gap-2 border-b border-border px-4 py-2 text-sm lg:hidden"
        >
          <NavLink
            to="/"
            end
            className={({ isActive }) =>
              cn("rounded-md px-3 py-2", (isActive || currentProject) && "bg-accent font-medium")
            }
          >
            Dự án
          </NavLink>
          <NavLink
            to="/my-work"
            className={({ isActive }) =>
              cn("rounded-md px-3 py-2", isActive && "bg-accent font-medium")
            }
          >
            Công việc của tôi
          </NavLink>
          {canLookup && (
            <NavLink
              to="/personnel"
              className={({ isActive }) =>
                cn("rounded-md px-3 py-2", isActive && "bg-accent font-medium")
              }
            >
              Tra cứu nhân sự
            </NavLink>
          )}
          <NavLink
            to="/my-profile"
            className={({ isActive }) =>
              cn("rounded-md px-3 py-2", isActive && "bg-accent font-medium")
            }
          >
            Hồ sơ của tôi
          </NavLink>
          {canManageAccounts && (
            <NavLink
              to="/accounts"
              className={({ isActive }) =>
                cn("rounded-md px-3 py-2", isActive && "bg-accent font-medium")
              }
            >
              Tài khoản
            </NavLink>
          )}
          {canReviewProfile && (
            <NavLink
              to="/profile-requests"
              className={({ isActive }) =>
                cn("rounded-md px-3 py-2", isActive && "bg-accent font-medium")
              }
            >
              Yêu cầu cá nhân
            </NavLink>
          )}
          {canReviewProfile && (
            <NavLink
              to="/employee-profiles"
              className={({ isActive }) =>
                cn("rounded-md px-3 py-2", isActive && "bg-accent font-medium")
              }
            >
              Hồ sơ nhân sự
            </NavLink>
          )}
          {canManageSkills && (
            <NavLink
              to="/skill-requests"
              className={({ isActive }) =>
                cn("rounded-md px-3 py-2", isActive && "bg-accent font-medium")
              }
            >
              Xét duyệt kỹ năng
            </NavLink>
          )}
          {canManageSkills && (
            <NavLink
              to="/skill-catalog"
              className={({ isActive }) =>
                cn(
                  "flex min-h-10 items-center rounded-md px-3 py-2 text-sm font-medium",
                  isActive ? "bg-accent" : "hover:bg-accent",
                )
              }
            >
              Danh mục kỹ năng
            </NavLink>
          )}
          {canManageOrganization && (
            <NavLink
              to="/organization"
              className={({ isActive }) =>
                cn("rounded-md px-3 py-2", isActive && "bg-accent font-medium")
              }
            >
              Danh mục tổ chức
            </NavLink>
          )}
        </nav>
        <main className="mx-auto max-w-[1600px] px-4 py-8 md:px-8">
          {notice && (
            <div
              role="status"
              className="mb-5 flex items-center justify-between gap-4 rounded-md border border-success/20 bg-success-soft px-4 py-3 text-sm text-success"
            >
              {notice}
              <Button size="sm" variant="ghost" onClick={onDismissNotice}>
                Ẩn
              </Button>
            </div>
          )}
          {children}
        </main>
      </div>
      {dialogs}
    </div>
  )
}
