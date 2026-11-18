import { Link, Navigate, useLocation } from "react-router"
import { ArchiveIcon, ExitIcon } from "@radix-ui/react-icons"
import { Button } from "@/components/ui/button"
import type { Principal } from "./auth-api"
import type { AuthSession } from "./auth-session"
import {
  canManageAccounts,
  canManageCatalogs,
  canReadCatalogs,
  catalogKindAt,
  serverHome,
} from "./server-navigation"
import { ServerPersonalProfile } from "../profiles/ServerPersonalProfile"
import { ServerAccounts } from "../accounts/ServerAccounts"
import { ServerCatalogs } from "../organization/ServerCatalogs"

export function ServerWorkspace({
  principal,
  session,
  busy,
  message,
}: {
  principal: Principal
  session: AuthSession
  busy: boolean
  message: string
}) {
  const location = useLocation()
  const home = serverHome(principal)
  const kind = catalogKindAt(location.pathname)
  if (location.pathname === "/profile" && (location.search || location.hash))
    return <Navigate to="/profile" replace />
  if (location.pathname === "/organization")
    return <Navigate to="/organization/departments" replace />
  if (
    location.pathname === "/login" ||
    location.pathname === "/change-password" ||
    location.pathname === "/"
  )
    return <Navigate to={home} replace />
  const accountsPage = location.pathname === "/accounts"
  const denied =
    (!!kind && !canReadCatalogs(principal)) || (accountsPage && !canManageAccounts(principal))
  const profilePage = location.pathname === "/profile"
  return (
    <div className="min-h-dvh bg-background text-foreground">
      <header className="flex flex-wrap items-center justify-between gap-3 border-b border-border bg-white px-5 py-4">
        <Link to={home} className="flex items-center gap-3 font-semibold">
          <ArchiveIcon className="size-5" />
          TaskAssign
        </Link>
        <div className="flex min-w-0 flex-wrap items-center gap-3 text-sm">
          <span className="min-w-0 break-words">
            {principal.fullname} · {principal.employee_code}
          </span>
          <Button
            variant="outline"
            disabled={busy}
            onClick={() => {
              void session.logout()
            }}
          >
            <ExitIcon />
            {busy ? "Đang kiểm tra…" : "Đăng xuất"}
          </Button>
        </div>
      </header>
      <div className="mx-auto grid max-w-7xl gap-6 p-5 md:grid-cols-[200px_minmax(0,1fr)] md:gap-8 md:p-8">
        <nav
          aria-label="Điều hướng máy chủ"
          className="flex flex-wrap gap-2 self-start md:flex-col"
        >
          <Link
            className={`filter-chip ${profilePage ? "filter-active" : ""}`}
            aria-current={profilePage ? "page" : undefined}
            to="/profile"
          >
            Hồ sơ của tôi
          </Link>
          {canManageCatalogs(principal) && (
            <Link
              className={`filter-chip ${kind ? "filter-active" : ""}`}
              aria-current={kind ? "page" : undefined}
              to="/organization/departments"
            >
              Danh mục tổ chức
            </Link>
          )}
          {canManageAccounts(principal) && (
            <Link
              className={`filter-chip ${accountsPage ? "filter-active" : ""}`}
              aria-current={accountsPage ? "page" : undefined}
              to="/accounts"
            >
              Tài khoản
            </Link>
          )}
        </nav>
        <main className="min-w-0">
          {message && (
            <p role="alert" className="mb-4 rounded-lg bg-accent p-3 text-sm">
              {message}
            </p>
          )}
          {busy && (
            <p role="status" className="mb-3 text-xs text-muted-foreground">
              Đang kiểm tra với máy chủ…
            </p>
          )}
          {profilePage ? (
            <ServerPersonalProfile principal={principal} session={session} busy={busy} />
          ) : accountsPage && !denied ? (
            <ServerAccounts key={principal.user_id} session={session} busy={busy} />
          ) : kind && !denied ? (
            <ServerCatalogs
              key={`${kind}:${location.search}`}
              kind={kind}
              canWrite={canManageCatalogs(principal)}
              session={session}
              busy={busy}
            />
          ) : (
            <section className="rounded-xl border border-border bg-white p-6">
              <h1 className="text-2xl font-semibold">
                {denied ? "Bạn không có quyền truy cập" : "Chức năng chưa được kết nối"}
              </h1>
              <p className="mt-3 text-sm leading-6 text-muted-foreground">
                {denied
                  ? accountsPage
                    ? "Quản lý tài khoản chỉ dành cho Quản trị viên. Quyền được xác nhận từ máy chủ."
                    : "Danh mục tổ chức chỉ dành cho HR hoặc Quản trị viên. Quyền được xác nhận từ máy chủ."
                  : "Dữ liệu của chức năng này chưa được kết nối với hệ thống."}
              </p>
              <p className="mt-5 text-sm">
                Vai trò hệ thống:{" "}
                {principal.system_roles
                  .map((role) => ({ ADMIN: "Quản trị viên", HR: "HR", DIRECTOR: "Giám đốc" })[role])
                  .join(", ") || "Nhân viên"}
              </p>
              <Button asChild variant="outline" className="mt-5">
                <Link to={home}>Về màn chính</Link>
              </Button>
            </section>
          )}
        </main>
      </div>
    </div>
  )
}
