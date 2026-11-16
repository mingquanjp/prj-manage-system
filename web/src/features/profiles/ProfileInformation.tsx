import { Avatar } from "@/components/shared"
import { ProfileFields } from "./ProfileFields"

export type ProfileInformationValues = {
  employeeCode: string
  fullname: string
  companyEmail: string
  phone: string | null
  address: string | null
  position: string
  department: string | null
  jobLevel: string | null
}
export function ProfileInformation({
  profile,
  initials,
  addressNotice,
}: {
  profile: ProfileInformationValues
  initials: string
  addressNotice: string
}) {
  return (
    <div className="rounded-xl border border-border bg-white px-5 py-4 sm:px-6 sm:py-5">
      <div className="mb-5 flex items-center gap-3 border-b border-border pb-4">
        <Avatar initials={initials} />
        <div className="min-w-0 break-words">
          <h2 className="section-title">{profile.fullname}</h2>
          <p className="mt-1 text-xs text-muted-foreground">
            {profile.employeeCode} · {profile.position}
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
            <p className="mt-2 text-xs leading-5 text-warning">{addressNotice}</p>
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
              ["Vị trí công việc", profile.position],
              ["Cấp bậc", profile.jobLevel],
            ]}
          />
        </section>
      </div>
    </div>
  )
}
