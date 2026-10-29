import { useState } from "react"
import { ArchiveIcon, ArrowRightIcon } from "@radix-ui/react-icons"
import { Button } from "@/components/ui/button"
import type { DemoWorkspaceState } from "@/demo/workspace-state"
import { demoIdentityOptions } from "@/app/navigation"
import type { IdentitySource } from "@/demo/identity"

export function Login({
  state,
  identity,
  message,
  onLogin,
}: {
  state: DemoWorkspaceState
  identity: IdentitySource
  message: string
  onLogin: (personId: string) => void
}) {
  const [personId, setPersonId] = useState("khoa")
  const [error, setError] = useState("")
  const options = demoIdentityOptions(state.projects, undefined, identity.people, identity)
  return (
    <main className="grid min-h-dvh place-items-center bg-background p-5 text-foreground">
      <div className="w-full max-w-md">
        <div className="mb-8 flex items-center gap-3">
          <span className="grid size-10 place-items-center rounded-xl bg-primary text-white">
            <ArchiveIcon className="size-5" />
          </span>
          <span className="text-xl font-semibold">TaskAssign</span>
        </div>
        <section className="rounded-2xl border border-border bg-white p-6 shadow-sm sm:p-8">
          <p className="mb-3 text-xs font-medium uppercase tracking-wider text-muted-foreground">
            Không gian làm việc
          </p>
          <h1 className="text-2xl font-semibold">Đăng nhập</h1>
          <p className="mt-2 text-sm leading-6 text-muted-foreground">
            Tiếp tục quản lý dự án và công việc của bạn.
          </p>
          {message && (
            <p role="status" className="mt-5 rounded-lg bg-accent p-3 text-sm leading-6">
              {message}
            </p>
          )}
          <form
            className="mt-6"
            onSubmit={(event) => {
              event.preventDefault()
              try {
                onLogin(personId)
                setError("")
              } catch (err) {
                setError((err as Error).message)
              }
            }}
          >
            <label className="field-label block" htmlFor="demo-login">
              Danh tính mô phỏng
            </label>
            <select
              id="demo-login"
              className="field mt-2 w-full"
              value={personId}
              onChange={(e) => {
                setPersonId(e.target.value)
                setError("")
              }}
            >
              {options.map((person) => (
                <option key={person.id} value={person.id}>
                  {person.name} · {person.responsibility}
                </option>
              ))}
            </select>
            {error && (
              <p role="alert" className="mt-3 text-sm text-destructive">
                {error}
              </p>
            )}
            <Button type="submit" className="mt-5 w-full">
              Đăng nhập mô phỏng <ArrowRightIcon />
            </Button>
          </form>
          <div className="mt-6 border-t border-border pt-4 text-xs leading-5 text-muted-foreground">
            <strong className="font-medium text-foreground">Bản prototype</strong>
            <p className="mt-1">
              Chọn danh tính để xem giao diện. Không nhập mật khẩu thật; chưa kết nối hệ thống xác
              thực.
            </p>
          </div>
        </section>
        <p className="mt-5 text-center text-xs leading-5 text-muted-foreground">
          Tài khoản do công ty cấp. Kênh cấp tài khoản sẽ được triển khai cùng hệ thống xác thực.
        </p>
      </div>
    </main>
  )
}
