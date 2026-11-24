import { useCallback, useEffect, useState, useSyncExternalStore } from "react"
import { useSearchParams } from "react-router"
import type { Principal } from "../session/auth-api"
import type { AuthSession } from "../session/auth-session"
import { Button } from "@/components/ui/button"
import { createProfileApi } from "./profile-api"
import { createProfileResource } from "./profile-resource"
import { ProfileInformation } from "./ProfileInformation"
import { createProfileRequestsApi } from "./profile-requests-api"
import { createProfileRequestsResource } from "./profile-requests-resource"
import { buildProfileRequest, type ProfileValues } from "./profile-request-input"
import { ServerProfileChangeForm } from "./ServerProfileChangeForm"
import {
  ServerProfileRequestBadge,
  ServerProfileRequestDetails,
} from "./ServerProfileRequestDetails"
import { Dialog } from "@/components/ui/dialog"
import { profileFieldLabels } from "@/models/profile"

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
  const [api] = useState(() => createProfileRequestsApi())
  const [requests] = useState(() => createProfileRequestsResource(api))
  const requestState = useSyncExternalStore(requests.subscribe, requests.getSnapshot)
  const [params, setParams] = useSearchParams()
  const [page, setPage] = useState(1)
  const [before, setBefore] = useState<ProfileValues | null>(null)
  const selected = params.get("request_id")
  const requestId =
    selected && /^[1-9]\d*$/.test(selected) && Number(selected) <= 2147483647
      ? Number(selected)
      : null
  const { user_id: userId, employee_id: employeeId } = principal
  const recheckSession = useCallback(() => {
    setBefore(null)
    resource.invalidate("Quyền hoặc phiên đã thay đổi. Hãy kiểm tra phiên rồi tải lại hồ sơ.")
    requests.clear()
    void session.refresh()
  }, [resource, requests, session])
  useEffect(() => {
    document.getElementById("profile-heading")?.focus()
    void resource.load({ user_id: userId, employee_id: employeeId }, recheckSession)
    return () => {
      resource.clear()
      requests.clear()
    }
  }, [resource, requests, userId, employeeId, recheckSession])
  useEffect(() => {
    void requests.loadList(employeeId, page, recheckSession)
  }, [requests, employeeId, page, recheckSession])
  useEffect(() => {
    if (requestId) void requests.loadDetail(employeeId, requestId, recheckSession)
    else requests.clearDetail()
  }, [requests, employeeId, requestId, recheckSession])
  const profile = state.profile
  const blocked = requestState.submission === "uncertain" || requestState.submission === "stale"
  const reloadList = () => requests.loadList(employeeId, page, recheckSession)
  const reloadProfile = () =>
    resource.load({ user_id: userId, employee_id: employeeId }, recheckSession)
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
      {state.phase === "loading" ? (
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
              void reloadProfile().then(() => {
                void reloadList()
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
            addressNotice="Bạn chưa bổ sung địa chỉ. Hãy gửi đề nghị khi hoàn thiện hồ sơ; thiếu địa chỉ không chặn sử dụng hệ thống."
          />
          <Button
            id="profile-request-trigger"
            variant="outline"
            className="mt-5"
            disabled={busy || blocked || requestState.submission === "pending"}
            onClick={() => {
              requests.allowNewDraft()
              setBefore({
                fullname: profile.fullname,
                phone: profile.phone,
                address: profile.address,
              })
            }}
          >
            Đề nghị cập nhật
          </Button>
          <section className="mt-8" aria-labelledby="profile-history-heading">
            <h2 id="profile-history-heading" tabIndex={-1} className="section-title mb-3">
              Lịch sử đề nghị
            </h2>
            {requestState.message && (
              <p role="status" className="mb-4 rounded-lg bg-accent p-3 text-sm">
                {requestState.message}
              </p>
            )}
            {blocked && !before && (
              <div className="mb-4 space-y-3 text-sm">
                <p>
                  {requestState.submission === "stale"
                    ? "Tải lại hồ sơ trước khi mở bản nháp mới; nội dung trước đó không tự áp lại."
                    : "Tải lịch sử và mở chi tiết để đối chiếu. Chỉ mở bản nháp mới khi bạn đã kiểm tra kết quả; hệ thống không gửi lại đề nghị cũ."}
                </p>
                <Button
                  variant="outline"
                  disabled={busy || requestState.list.phase !== "ready"}
                  onClick={() => {
                    void reloadProfile().then(() => {
                      requests.allowNewDraft()
                    })
                  }}
                >
                  Đã đối soát — tải lại hồ sơ
                </Button>
              </div>
            )}
            <Button
              variant="outline"
              className="mb-4"
              disabled={busy || requestState.list.phase === "loading"}
              onClick={() => {
                void reloadList()
              }}
            >
              Tải lại lịch sử
            </Button>
            {requestState.list.phase === "loading" ? (
              <p role="status">Đang tải lịch sử đề nghị…</p>
            ) : requestState.list.phase === "error" ? (
              <p role="alert">{requestState.list.message}</p>
            ) : requestState.list.data ? (
              <>
                <p className="mb-3 text-sm text-muted-foreground">
                  {requestState.list.data.total} đề nghị · Trang {page}
                </p>
                {requestState.list.data.items.length ? (
                  <ul className="space-y-3">
                    {requestState.list.data.items.map((entry) => (
                      <li
                        key={entry.request_id}
                        className="rounded-lg border border-border bg-white p-4"
                      >
                        <button
                          className="w-full cursor-pointer space-y-2 text-left text-sm"
                          onClick={() => setParams({ request_id: String(entry.request_id) })}
                        >
                          <span className="block font-medium">
                            Đề nghị #{entry.request_id} ·{" "}
                            {entry.changed_fields
                              .map((field) => profileFieldLabels[field])
                              .join(", ")}
                          </span>
                          <ServerProfileRequestBadge status={entry.status} />{" "}
                          <span className="ml-2 text-xs text-muted-foreground">
                            {new Date(entry.requested_at).toLocaleString("vi-VN")}
                          </span>
                        </button>
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p className="border-t border-border pt-4 text-sm text-muted-foreground">
                    {requestState.list.data.total === 0
                      ? "Chưa có đề nghị. Khi cần thay đổi, chọn Đề nghị cập nhật."
                      : "Trang này không có đề nghị; hãy quay lại trang trước."}
                  </p>
                )}
                <div className="mt-4 flex gap-2">
                  <Button
                    variant="outline"
                    disabled={busy || page === 1}
                    onClick={() => setPage(page - 1)}
                  >
                    Trang trước
                  </Button>
                  <Button
                    variant="outline"
                    disabled={busy || page * 20 >= requestState.list.data.total}
                    onClick={() => setPage(page + 1)}
                  >
                    Trang sau
                  </Button>
                </div>
              </>
            ) : null}
          </section>
        </>
      ) : null}
      {before && (
        <ServerProfileChangeForm
          before={before}
          busy={busy}
          blocked={blocked}
          onClose={() => setBefore(null)}
          onSubmit={async (values, explanation) => {
            const input = buildProfileRequest(before, values, explanation)
            const result = await requests.submit(
              (post) => session.catalogMutation(post),
              (token, signal) => api.submit(employeeId, input, crypto.randomUUID(), token, signal),
              recheckSession,
            )
            if (!result) return
            setBefore(null)
            setPage(1)
            setParams({ request_id: String(result.request_id) })
            void requests.loadList(employeeId, 1, recheckSession)
          }}
        />
      )}
      {selected !== null && (
        <Dialog
          open
          onOpenChange={(open) => {
            if (!open) {
              setParams({})
              void reloadProfile()
            }
          }}
          title="Chi tiết đề nghị"
          description="Thông tin đề nghị và kết luận từ hệ thống. Hồ sơ chính thức hiển thị riêng."
          wide
          onCloseAutoFocus={(event) => {
            event.preventDefault()
            document.getElementById("profile-heading")?.focus()
          }}
        >
          {!requestId ? (
            <p role="alert">Mã đề nghị không hợp lệ.</p>
          ) : requestState.detail.phase === "loading" ? (
            <p role="status">Đang tải chi tiết đề nghị…</p>
          ) : requestState.detail.phase === "error" ? (
            <>
              <p role="alert">{requestState.detail.message}</p>
              <Button
                variant="outline"
                disabled={busy}
                onClick={() => {
                  void requests.loadDetail(employeeId, requestId, recheckSession)
                }}
              >
                Tải lại chi tiết
              </Button>
            </>
          ) : requestState.detail.data ? (
            <ServerProfileRequestDetails request={requestState.detail.data} />
          ) : null}
        </Dialog>
      )}
    </section>
  )
}
