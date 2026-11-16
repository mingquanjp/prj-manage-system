import { useEffect, useState, useSyncExternalStore } from "react"
import type { Principal } from "../session/auth-api"
import type { AuthSession } from "../session/auth-session"
import { Button } from "@/components/ui/button"
import { createProfileApi } from "./profile-api"
import { createProfileResource } from "./profile-resource"
import { ProfileInformation } from "./ProfileInformation"

export function ServerPersonalProfile({
  principal,
  session,
  busy,
}: {
  principal: Principal
  session: AuthSession
  busy: boolean
}) {
  const [resource] = useState(() => createProfileResource(createProfileApi()))
  const state = useSyncExternalStore(resource.subscribe, resource.getSnapshot)
  const { user_id: userId, employee_id: employeeId } = principal
  useEffect(() => {
    document.getElementById("profile-heading")?.focus()
    void resource.load({ user_id: userId, employee_id: employeeId }, () => {
      void session.refresh()
    })
    return resource.clear
  }, [resource, userId, employeeId, session])
  const profile = state.profile
  return (
    <section className="mx-auto w-full max-w-5xl" aria-busy={busy || state.phase === "loading"}>
      <div className="mb-6">
        <p className="eyebrow">KHÔNG GIAN CÁ NHÂN</p>
        <h1 id="profile-heading" tabIndex={-1} className="mt-2 text-2xl font-semibold">
          Hồ sơ của tôi
        </h1>
        <p className="mt-2 text-sm text-muted-foreground">
          Thông tin cá nhân và công tác chính thức từ hệ thống.
        </p>
      </div>
      {busy || state.phase === "loading" ? (
        <p role="status">Đang tải và kiểm tra hồ sơ với máy chủ…</p>
      ) : state.phase === "error" ? (
        <div className="rounded-xl border border-border bg-white p-5">
          <p role="alert" className="text-sm leading-6">
            {state.message}
          </p>
          <Button
            variant="outline"
            className="mt-4"
            onClick={() => {
              document.getElementById("profile-heading")?.focus()
              void resource.load({ user_id: userId, employee_id: employeeId }, () => {
                void session.refresh()
              })
            }}
          >
            Tải lại hồ sơ
          </Button>
        </div>
      ) : profile ? (
        <>
          <ProfileInformation
            profile={{
              employeeCode: profile.employee_code,
              fullname: profile.fullname,
              companyEmail: profile.company_email,
              phone: profile.phone,
              address: profile.address,
              position: profile.position.name,
              department: profile.department?.name ?? null,
              jobLevel: profile.job_level?.name ?? null,
            }}
            initials={profile.fullname
              .trim()
              .split(/\s+/u)
              .slice(-2)
              .map((part) => Array.from(part)[0])
              .join("")}
            addressNotice="Bạn chưa bổ sung địa chỉ khi hoàn thiện hồ sơ. Chức năng đề nghị cập nhật sẽ được kết nối sau; thiếu địa chỉ không chặn sử dụng hệ thống."
          />
          <section className="mt-8" aria-labelledby="profile-history-heading">
            <h2 id="profile-history-heading" className="section-title mb-3">
              Lịch sử đề nghị
            </h2>
            <p className="border-t border-border pt-4 text-sm leading-6 text-muted-foreground">
              Chưa kết nối lịch sử đề nghị cập nhật hồ sơ.
            </p>
          </section>
        </>
      ) : null}
    </section>
  )
}
