import { useEffect, useState, useSyncExternalStore, type ReactNode } from "react"
import { Navigate, useLocation, useNavigate } from "react-router"
import { ArchiveIcon, ExitIcon } from "@radix-ui/react-icons"
import { Button } from "@/components/ui/button"
import { createAuthApi, type SystemRole } from "./auth-api"
import { createAuthSession } from "./auth-session"

const roleLabels: Record<SystemRole, string> = {
  ADMIN: "Quản trị viên",
  HR: "HR",
  DIRECTOR: "Giám đốc",
}

function Frame({ children }: { children: ReactNode }) {
  return (
    <main className="grid min-h-dvh place-items-center bg-background p-5 text-foreground">
      <div className="w-full max-w-lg">
        <div className="mb-7 flex items-center gap-3">
          <span className="grid size-10 place-items-center rounded-xl bg-primary text-primary-foreground">
            <ArchiveIcon className="size-5" />
          </span>
          <span className="text-xl font-semibold">TaskAssign</span>
          <span className="ml-auto text-xs text-muted-foreground">Kết nối máy chủ</span>
        </div>
        <section className="rounded-2xl border border-border bg-white p-6 shadow-sm sm:p-8">
          {children}
        </section>
      </div>
    </main>
  )
}

export function ServerSession() {
  const [session] = useState(() => createAuthSession(createAuthApi()))
  const state = useSyncExternalStore(session.subscribe, session.getSnapshot)
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [validation, setValidation] = useState("")
  const [clock, setClock] = useState(Date.now)
  const location = useLocation()
  const navigate = useNavigate()
  const waitSeconds = Math.max(0, Math.ceil((state.retryAt - clock) / 1000))

  useEffect(() => {
    void session.check()
    const check = () => {
      void session.check()
    }
    const visible = () => {
      if (document.visibilityState === "visible") check()
    }
    window.addEventListener("focus", check)
    window.addEventListener("pageshow", check)
    window.addEventListener("popstate", check)
    document.addEventListener("visibilitychange", visible)
    return () => {
      window.removeEventListener("focus", check)
      window.removeEventListener("pageshow", check)
      window.removeEventListener("popstate", check)
      document.removeEventListener("visibilitychange", visible)
    }
  }, [session])

  useEffect(() => {
    if (!state.retryAt) return
    const timer = window.setInterval(() => setClock(Date.now()), 1000)
    return () => window.clearInterval(timer)
  }, [state.retryAt])

  // Không giữ input mật khẩu sau khi rời form, kể cả khi phiên được khôi phục ở tab khác.
  useEffect(() => {
    if (state.phase !== "signed-out") setPassword("")
  }, [state.phase])

  const message = validation || state.message
  const feedback = message ? (
    <p role="alert" className="mt-4 rounded-lg bg-accent p-3 text-sm leading-6">
      {message}
    </p>
  ) : null
  if (state.phase === "checking")
    return (
      <Frame>
        <h1 className="text-xl font-semibold">Đang kiểm tra phiên</h1>
        <p role="status" className="mt-3 text-sm text-muted-foreground">
          Vui lòng chờ trong giây lát…
        </p>
      </Frame>
    )
  if (state.phase === "unavailable")
    return (
      <Frame>
        <h1 className="text-xl font-semibold">Chưa thể xác nhận phiên</h1>
        {feedback}
        <p className="mt-3 text-sm leading-6 text-muted-foreground">
          Hãy kiểm tra kết nối rồi thử lại. Không có thao tác đăng nhập hoặc đăng xuất nào được tự
          gửi lại.
        </p>
        <Button
          className="mt-5"
          disabled={waitSeconds > 0}
          onClick={() => {
            setValidation("")
            void session.check()
          }}
        >
          Kiểm tra lại phiên{waitSeconds > 0 ? ` (${waitSeconds}s)` : ""}
        </Button>
      </Frame>
    )
  if (state.phase === "signed-out") {
    if (location.pathname !== "/login" || location.search || location.hash)
      return <Navigate to="/login" replace />
    return (
      <Frame>
        <h1 className="text-2xl font-semibold">Đăng nhập</h1>
        <p className="mt-2 text-sm leading-6 text-muted-foreground">
          Sử dụng email công ty và mật khẩu được cấp.
        </p>
        {feedback}
        <form
          className="mt-6 space-y-4"
          onSubmit={(event) => {
            event.preventDefault()
            if (!email.trim() || !password) {
              setValidation("Vui lòng nhập email và mật khẩu.")
              return
            }
            setValidation("")
            const input = password
            setPassword("")
            void session.login(email, input)
          }}
        >
          <div>
            <label htmlFor="auth-email" className="field-label">
              Email công ty
            </label>
            <input
              id="auth-email"
              name="email"
              className="field mt-2 w-full"
              type="email"
              autoComplete="username"
              required
              value={email}
              disabled={state.busy}
              onChange={(event) => {
                setEmail(event.target.value)
                setValidation("")
              }}
            />
          </div>
          <div>
            <label htmlFor="auth-password" className="field-label">
              Mật khẩu
            </label>
            <input
              id="auth-password"
              name="password"
              className="field mt-2 w-full"
              type="password"
              autoComplete="current-password"
              required
              value={password}
              disabled={state.busy}
              onChange={(event) => {
                setPassword(event.target.value)
                setValidation("")
              }}
            />
          </div>
          <Button type="submit" className="w-full" disabled={state.busy || waitSeconds > 0}>
            {state.busy
              ? "Đang đăng nhập…"
              : waitSeconds > 0
                ? `Thử lại sau ${waitSeconds}s`
                : "Đăng nhập"}
          </Button>
        </form>
        <p className="mt-5 text-xs leading-5 text-muted-foreground">
          Tài khoản do công ty cấp. Chưa triển khai chức năng quên mật khẩu.
        </p>
      </Frame>
    )
  }
  const principal = state.principal!
  const limited = principal.must_change_password
  if (
    location.pathname === "/login" ||
    (limited && location.pathname !== "/change-password") ||
    location.search ||
    location.hash
  )
    return <Navigate to={limited ? "/change-password" : "/"} replace />
  const unsupported = !limited && location.pathname !== "/"
  return (
    <Frame>
      <h1 className="text-2xl font-semibold">
        {limited
          ? "Cần đổi mật khẩu"
          : unsupported
            ? "Chức năng chưa được kết nối"
            : "Đã đăng nhập"}
      </h1>
      <p className="mt-2 text-sm leading-6 text-muted-foreground">
        {limited
          ? "Bạn cần đổi mật khẩu trước khi sử dụng các chức năng nghiệp vụ. Tính năng này chưa sẵn sàng trong phiên bản hiện tại; bạn có thể đăng xuất."
          : unsupported
            ? "Dữ liệu của chức năng này chưa được kết nối với hệ thống. Bạn có thể quay về màn chính."
            : "Các chức năng nghiệp vụ đang được kết nối. Đây là màn tiếp nối tạm thời."}
      </p>
      <dl className="mt-6 space-y-3 border-y border-border py-5 text-sm">
        <div>
          <dt className="field-label">Họ và tên</dt>
          <dd className="mt-1 break-words">{principal.fullname}</dd>
        </div>
        <div>
          <dt className="field-label">Mã nhân viên</dt>
          <dd className="mt-1">{principal.employee_code}</dd>
        </div>
        {!limited && (
          <div>
            <dt className="field-label">Vai trò hệ thống</dt>
            <dd className="mt-1">
              {principal.system_roles.length
                ? principal.system_roles.map((role) => roleLabels[role]).join(", ")
                : "Nhân viên"}
            </dd>
          </div>
        )}
      </dl>
      {feedback}
      <div className="mt-5 flex flex-wrap gap-3">
        {unsupported && (
          <Button variant="outline" onClick={() => navigate("/", { replace: true })}>
            Về màn chính
          </Button>
        )}
        <Button
          disabled={state.busy || waitSeconds > 0}
          onClick={() => {
            setValidation("")
            void session.logout()
          }}
        >
          <ExitIcon />
          {state.busy
            ? "Đang đăng xuất…"
            : waitSeconds > 0
              ? `Thử lại sau ${waitSeconds}s`
              : "Đăng xuất"}
        </Button>
      </div>
    </Frame>
  )
}
