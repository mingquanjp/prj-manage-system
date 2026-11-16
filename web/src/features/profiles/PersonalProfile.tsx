import { useSearchParams } from "react-router"
import { EmployeeCvs, type CvActions } from "@/features/cvs/EmployeeCvs"
import type { EmployeeCvsState } from "@/models/employee-cvs"
import { PersonalSkills, type PersonalSkillActions } from "@/features/skills/PersonalSkills"
import type { EmployeeSkillsState } from "@/models/employee-skills"
import type { Task } from "@/models/workspace"
import type { WorkState } from "@/demo/effort"
import { cn } from "@/lib/utils"
import { ProfileInformation } from "./ProfileInformation"
import { useState } from "react"
import { Pencil1Icon } from "@radix-ui/react-icons"
import { EmptyState } from "@/components/shared"
import { Button } from "@/components/ui/button"
import { usePeople } from "@/components/people-directory"
import type { ProfileState, ProfileValues } from "@/models/profile"
import type { OrganizationState } from "@/models/organization"
import { describeProfileChanges } from "@/demo/profiles"
import { ProfileChangeForm } from "./ProfileChangeForm"
import { ProfileRequestBadge, ProfileRequestDetails } from "./ProfileRequestDetails"

export function PersonalProfile({
  actorId,
  state,
  organization,
  onRequest,
  employeeSkills,
  work,
  tasks,
  skillActions,
  cvs,
  cvActions,
}: {
  actorId: string
  cvs: EmployeeCvsState
  cvActions: CvActions
  employeeSkills: EmployeeSkillsState
  work: WorkState
  tasks: Task[]
  skillActions: PersonalSkillActions
  state: ProfileState
  organization: OrganizationState
  onRequest: (before: ProfileValues, proposed: ProfileValues, explanation: string) => void
}) {
  const [params] = useSearchParams()
  const { getPerson } = usePeople()
  const [tab, setTab] = useState(params.get("tab") === "skills" ? "skills" : "information")
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
    <div className="mx-auto w-full max-w-5xl">
      <div className="mb-6 flex flex-wrap items-start justify-between gap-4">
        <div>
          <p className="eyebrow">KHÔNG GIAN CÁ NHÂN</p>
          <h1 className="mt-2 text-2xl font-semibold text-balance">Hồ sơ của tôi</h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Thông tin chính thức và các đề nghị cập nhật của bạn.
          </p>
        </div>
        {tab === "information" && (
          <Button variant="outline" onClick={() => setEditing(true)}>
            <Pencil1Icon /> Đề nghị cập nhật
          </Button>
        )}
      </div>
      <div
        className="mb-6 flex flex-wrap gap-1 border-b border-border pb-3"
        role="group"
        aria-label="Nội dung hồ sơ cá nhân"
      >
        {[
          ["information", "Thông tin cá nhân"],
          ["cvs", "CV của tôi"],
          ["skills", "Kỹ năng"],
        ].map(([id, label]) => (
          <button
            type="button"
            key={id}
            className={cn("filter-chip", tab === id && "filter-active")}
            aria-pressed={tab === id}
            onClick={() => setTab(id)}
          >
            {label}
          </button>
        ))}
      </div>
      {tab === "cvs" ? (
        <EmployeeCvs
          actorId={actorId}
          employeeId={actorId}
          isActive={person.active}
          state={cvs}
          actions={cvActions}
        />
      ) : tab === "skills" ? (
        <PersonalSkills
          actorId={actorId}
          state={employeeSkills}
          work={work}
          tasks={tasks}
          actions={skillActions}
        />
      ) : (
        <>
          <ProfileInformation
            profile={{
              ...profile,
              position: person.position,
              department:
                organization.departments.find((entry) => entry.id === profile.departmentId)?.name ??
                null,
              jobLevel: person.seniority,
            }}
            initials={person.initials}
            addressNotice="Bạn chưa bổ sung địa chỉ. Hãy gửi đề nghị khi hoàn thiện hồ sơ."
          />
          <section className="mt-8" aria-label="Lịch sử yêu cầu cập nhật hồ sơ">
            <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
              <h2 className="section-title">Lịch sử đề nghị</h2>
              <span className="text-sm text-muted-foreground">{requests.length} đề nghị</span>
            </div>
            {requests.length ? (
              <div className="space-y-3">
                {requests.map((request) => (
                  <details
                    key={request.id}
                    open={params.get("request") === request.id ? true : undefined}
                    className="rounded-lg border border-border bg-white p-4"
                  >
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
        </>
      )}
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
