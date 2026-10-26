import { useState } from "react"
import { Pencil1Icon } from "@radix-ui/react-icons"
import { Avatar, EmptyState } from "@/components/shared"
import { Button } from "@/components/ui/button"
import { usePeople } from "@/components/people-directory"
import type { ProfileState, ProfileValues } from "@/models/profile"
import { describeProfileChanges } from "@/demo/profiles"
import { ProfileChangeForm } from "./ProfileChangeForm"
import { ProfileRequestBadge, ProfileRequestDetails } from "./ProfileRequestDetails"

export function PersonalProfile({
  actorId,
  state,
  onRequest,
}: {
  actorId: string
  state: ProfileState
  onRequest: (before: ProfileValues, proposed: ProfileValues, explanation: string) => void
}) {
  const { getPerson } = usePeople()
  const [editing, setEditing] = useState(false)
  const profile = state.profiles.find((entry) => entry.personId === actorId)
  const person = getPerson(actorId)
  if (!profile || !person?.active)
    return (
      <EmptyState
        title="Không thể mở hồ sơ"
        detail="Hồ sơ của bạn chưa có hoặc không còn hoạt động trong phiên mô phỏng."
      />
    )
  const requests = state.requests.filter((entry) => entry.employeeId === actorId).reverse()
  return (
    <div className="max-w-5xl">
      <div className="mb-6 flex flex-wrap items-start justify-between gap-4">
        <div>
          <p className="eyebrow">KHÔNG GIAN CÁ NHÂN</p>
          <h1 className="mt-2 text-2xl font-semibold text-balance">Hồ sơ của tôi</h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Thông tin chính thức và các đề nghị cập nhật của bạn.
          </p>
        </div>
        <Button variant="outline" onClick={() => setEditing(true)}>
          <Pencil1Icon /> Đề nghị cập nhật
        </Button>
      </div>
      <div className="rounded-xl border border-border bg-white px-5 py-4 sm:px-6 sm:py-5">
        <div className="mb-5 flex items-center gap-3 border-b border-border pb-4">
          <Avatar initials={person.initials} />
          <div>
            <h2 className="section-title">{profile.fullname}</h2>
            <p className="mt-1 text-xs text-muted-foreground">
              {profile.employeeCode} · {person.position}
            </p>
          </div>
        </div>
        <div className="grid gap-5 md:grid-cols-2 md:gap-6">
          <section aria-labelledby="personal-information-title" className="min-w-0">
            <h3 id="personal-information-title" className="section-title mb-3">
              Thông tin cá nhân
            </h3>
            <ProfileFields
              fields={[
                ["Họ và tên", profile.fullname],
                ["Số điện thoại", profile.phone],
                ["Địa chỉ", profile.address],
              ]}
            />
            {!profile.address && (
              <p className="mt-2 text-xs leading-5 text-warning">
                Bạn chưa bổ sung địa chỉ. Hãy gửi đề nghị khi hoàn thiện hồ sơ.
              </p>
            )}
          </section>
          <section
            aria-labelledby="employment-information-title"
            className="min-w-0 border-t border-border pt-5 md:border-t-0 md:border-l md:pt-0 md:pl-6"
          >
            <h3 id="employment-information-title" className="section-title mb-3">
              Thông tin công tác
            </h3>
            <ProfileFields
              fields={[
                ["Email công ty", profile.companyEmail],
                ["Phòng ban", profile.department],
                ["Vị trí công việc", person.position],
                ["Cấp bậc", person.seniority],
              ]}
            />
          </section>
        </div>
      </div>
      <section className="mt-8" aria-label="Lịch sử yêu cầu cập nhật hồ sơ">
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <h2 className="section-title">Lịch sử đề nghị</h2>
          <span className="text-sm text-muted-foreground">{requests.length} đề nghị</span>
        </div>
        {requests.length ? (
          <div className="space-y-3">
            {requests.map((request) => (
              <details key={request.id} className="rounded-lg border border-border bg-white p-4">
                <summary className="cursor-pointer text-sm leading-6">
                  <span className="mr-3 font-medium">{describeProfileChanges(request)}</span>
                  <ProfileRequestBadge status={request.status} />
                  <span className="ml-3 text-xs text-muted-foreground">
                    {new Date(request.requestedAt).toLocaleString("vi-VN")}
                  </span>
                </summary>
                <div className="mt-4">
                  <ProfileRequestDetails request={request} />
                </div>
              </details>
            ))}
          </div>
        ) : (
          <p className="border-t border-border pt-4 text-sm leading-6 text-muted-foreground">
            Chưa có đề nghị. Khi cần thay đổi thông tin, chọn Đề nghị cập nhật.
          </p>
        )}
      </section>
      {editing && (
        <ProfileChangeForm
          before={{ fullname: profile.fullname, phone: profile.phone, address: profile.address }}
          onSubmit={onRequest}
          onClose={() => setEditing(false)}
        />
      )}
    </div>
  )
}

function ProfileFields({ fields }: { fields: [string, string | null][] }) {
  return (
    <dl className="space-y-3">
      {fields.map(([label, value]) => (
        <div
          key={label}
          className="grid min-w-0 gap-1 text-sm leading-6 sm:grid-cols-[7.5rem_minmax(0,1fr)] sm:gap-3"
        >
          <dt className="font-medium text-muted-foreground">{label}</dt>
          <dd className="min-w-0 whitespace-pre-wrap break-words text-foreground">
            {value ?? <span className="text-muted-foreground">Chưa cập nhật</span>}
          </dd>
        </div>
      ))}
    </dl>
  )
}
