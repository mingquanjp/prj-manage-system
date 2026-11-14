import { lazy, Suspense } from "react"
import { ServerSession } from "@/features/session/ServerSession"
import { authMode } from "@/features/session/mode"

const DemoApp = lazy(() => import("@/app/DemoApp"))

export default function App() {
  const mode = authMode(import.meta.env.VITE_AUTH_MODE)
  if (mode === "invalid") {
    return (
      <main className="p-8" role="alert">
        Cấu hình xác thực không hợp lệ. Hãy dùng demo hoặc server.
      </main>
    )
  }
  return mode === "server" ? (
    <ServerSession />
  ) : (
    <Suspense
      fallback={
        <p role="status" className="p-8">
          Đang mở bản mô phỏng…
        </p>
      }
    >
      <DemoApp />
    </Suspense>
  )
}
