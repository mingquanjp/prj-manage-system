import { EmployeeCvs, type CvActions } from "@/features/cvs/EmployeeCvs"
import type { EmployeeCvsState } from "@/models/employee-cvs"
import { cn } from "@/lib/utils"
import { useState } from "react"
import { Pencil1Icon } from "@radix-ui/react-icons"
import { Avatar, EmptyState, SearchInput } from "@/components/shared"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { usePeople } from "@/components/people-directory"
import { isHr } from "@/demo/identity"
import { searchKey } from "@/demo/personnel"
import type { EmployeeProfile, EmployeeEditValues, ProfileState } from "@/models/profile"
import { employeeEditFields, employeeEditLabels } from "@/models/profile"
import type { OrganizationState } from "@/models/organization"
import type { Project, Task } from "@/models/workspace"
import { EmployeeDeactivationDialog } from "./EmployeeDeactivationDialog"
import { EmployeeEditForm } from "./EmployeeEditForm"
import { ProfileFields } from "./ProfileFields"

export function EmployeeProfiles({
  actorId,
  state,
  organization,
  onSave,
  projects,
  tasks,
  onDeactivate,
  cvs,
  cvActions,
}: {
  actorId: string
  cvs: EmployeeCvsState
  cvActions: CvActions
  state: ProfileState
  organization: OrganizationState
  projects: Project[]
  tasks: Task[]
  onDeactivate: (personId: string) => void
  onSave: (before: EmployeeProfile, values: EmployeeEditValues) => void
}) {
  const { getPerson } = usePeople()
  const [query, setQuery] = useState("")
  const [department, setDepartment] = useState("")
  const [position, setPosition] = useState("")
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [editing, setEditing] = useState(false)
  const [deactivating, setDeactivating] = useState(false)
  const [tab, setTab] = useState("information")
  const [status, setStatus] = useState("ACTIVE")
  if (!isHr(actorId))
    return (
      <EmptyState
        title="Không có quyền xem hồ sơ nhân sự"
        detail="Màn hình quản lý thông tin hành chính dành cho HR."
      />
    )
  const selected = state.profiles.find((profile) => profile.personId === selectedId)
  const nameOf = (kind: "departments" | "positions" | "jobLevels", id: string | null) =>
    organization[kind].find((entry) => entry.id === id)?.name ?? null
  const filtered = state.profiles.filter(
    (profile) =>
      (!status || profile.employmentStatus === status) &&
      (!department || profile.departmentId === department) &&
      (!position || profile.positionId === position) &&
      searchKey(`${profile.employeeCode} ${profile.fullname} ${profile.companyEmail}`).includes(
        searchKey(query),
      ),
  )
  const display = (field: string, value: string | null) =>
    field === "departmentId"
      ? (nameOf("departments", value) ?? "Chưa thuộc phòng ban")
      : field === "positionId"
        ? (nameOf("positions", value) ?? "Chưa cập nhật")
        : (value ?? "Chưa cập nhật")
  return (
    <div>
      <p className="eyebrow">QUẢN TRỊ NHÂN SỰ</p>
      <h1 className="mt-2 text-2xl font-semibold">Hồ sơ nhân sự</h1>
      <p className="mt-2 text-sm text-muted-foreground">
        Xem và cập nhật thông tin cá nhân, thông tin công tác của nhân viên.
      </p>
      <div className="mt-6 flex flex-wrap items-end gap-3">
        <div className="min-w-56 flex-1">
          <label className="field-label mb-2 block" htmlFor="employee-search">
            Tìm nhân viên
          </label>
          <SearchInput
            id="employee-search"
            value={query}
            onChange={setQuery}
            placeholder="Tìm theo mã, tên hoặc email công ty…"
          />
        </div>
        <div className="min-w-44">
          <label className="field-label mb-2 block" htmlFor="employee-status">
            Trạng thái hồ sơ
          </label>
          <select
            id="employee-status"
            className="field w-full"
            value={status}
            onChange={(event) => setStatus(event.target.value)}
          >
            <option value="ACTIVE">Đang hoạt động</option>
            <option value="INACTIVE">Đã vô hiệu hóa</option>
            <option value="">Tất cả trạng thái</option>
          </select>
        </div>
        {(["departments", "positions"] as const).map((kind) => (
          <div key={kind} className="min-w-44">
            <label className="field-label mb-2 block" htmlFor={`filter-${kind}`}>
              {kind === "departments" ? "Phòng ban" : "Vị trí công việc"}
            </label>
            <select
              id={`filter-${kind}`}
              className="field w-full"
              value={kind === "departments" ? department : position}
              onChange={(event) =>
                (kind === "departments" ? setDepartment : setPosition)(event.target.value)
              }
            >
              <option value="">
                {kind === "departments" ? "Tất cả phòng ban" : "Tất cả vị trí"}
              </option>
              {organization[kind].map((entry) => (
                <option key={entry.id} value={entry.id}>
                  {entry.name}
                </option>
              ))}
            </select>
          </div>
        ))}
      </div>
      <div className="mt-5 overflow-x-auto rounded-xl border border-border bg-white">
        <table className="w-full min-w-[780px] text-left text-sm">
          <thead className="table-heading border-b border-border bg-sidebar">
            <tr>
              {[
                "Mã nhân viên",
                "Họ và tên",
                "Email công ty",
                "Phòng ban",
                "Vị trí công việc",
                "Trạng thái",
              ].map((label) => (
                <th key={label} className="px-4 py-3">
                  {label}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {filtered.map((profile) => (
              <tr
                key={profile.personId}
                className="border-b border-border last:border-0 hover:bg-accent/40"
              >
                <td className="px-4 py-4 text-muted-foreground">{profile.employeeCode}</td>
                <td className="px-4 py-4">
                  <button
                    className="font-medium text-left underline-offset-4 hover:underline"
                    onClick={() => {
                      setSelectedId(profile.personId)
                      setTab("information")
                    }}
                  >
                    {profile.fullname}
                  </button>
                </td>
                <td className="px-4 py-4">{profile.companyEmail}</td>
                <td className="px-4 py-4">{display("departmentId", profile.departmentId)}</td>
                <td className="px-4 py-4">{display("positionId", profile.positionId)}</td>
                <td className="px-4 py-4 whitespace-nowrap">
                  {profile.employmentStatus === "ACTIVE" ? "Đang hoạt động" : "Đã vô hiệu hóa"}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {!filtered.length && (
          <EmptyState title="Không tìm thấy hồ sơ" detail="Thử thay đổi từ khóa hoặc bộ lọc." />
        )}
      </div>
      <p className="mt-3 text-xs text-muted-foreground">
        {filtered.length} / {state.profiles.length} hồ sơ
      </p>
      {selected && (
        <Dialog
          open
          side
          onOpenChange={(open) => {
            if (!open) setSelectedId(null)
          }}
          title="Hồ sơ nhân sự"
          description={`${selected.employeeCode} · Thông tin chính thức`}
        >
          <div className="mb-5 flex items-center gap-3">
            <Avatar initials={getPerson(selected.personId)?.initials ?? "NV"} />
            <div className="min-w-0 flex-1">
              <h2 className="section-title break-words">{selected.fullname}</h2>
              <p className="text-sm text-muted-foreground">
                {display("positionId", selected.positionId)}
              </p>
            </div>
            {tab === "information" && (
              <Button variant="outline" size="sm" onClick={() => setEditing(true)}>
                <Pencil1Icon /> Chỉnh sửa
              </Button>
            )}
          </div>
          <div
            className="mb-5 flex gap-1 border-b border-border pb-3"
            role="group"
            aria-label="Nội dung hồ sơ nhân sự"
          >
            {[
              ["information", "Thông tin"],
              ["cvs", "CV nhân sự"],
            ].map(([id, label]) => (
              <button
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
              key={selected.personId}
              actorId={actorId}
              employeeId={selected.personId}
              isActive={selected.employmentStatus === "ACTIVE"}
              state={cvs}
              actions={cvActions}
            />
          ) : (
            <>
              <div className="mb-4 flex flex-wrap items-center justify-between gap-3 rounded-md bg-accent p-3 text-sm">
                <span className="font-medium">
                  {selected.employmentStatus === "ACTIVE" ? "Đang hoạt động" : "Đã vô hiệu hóa"}
                </span>
                {selected.employmentStatus === "ACTIVE" && (
                  <Button
                    size="sm"
                    variant="ghost"
                    className="text-warning"
                    onClick={() => setDeactivating(true)}
                  >
                    Vô hiệu hóa hồ sơ
                  </Button>
                )}
              </div>
              {state.deactivations
                .filter((entry) => entry.employeeId === selected.personId)
                .map((entry) => (
                  <p key={entry.id} className="mb-4 text-xs leading-5 text-muted-foreground">
                    Vô hiệu hóa bởi {getPerson(entry.actorId)?.name ?? entry.actorId} ·{" "}
                    {new Date(entry.at).toLocaleString("vi-VN")}. Hồ sơ và lịch sử được giữ nguyên.
                  </p>
                ))}
              <section className="border-t border-border pt-4">
                <h3 className="section-title mb-3">Thông tin cá nhân</h3>
                <ProfileFields
                  fields={[
                    ["Họ và tên", selected.fullname],
                    ["Số điện thoại", selected.phone],
                    ["Địa chỉ", selected.address],
                  ]}
                />
                {!selected.address && (
                  <p className="mt-2 text-xs text-warning">
                    Chưa bổ sung địa chỉ khi hoàn thiện hồ sơ.
                  </p>
                )}
              </section>
              <section className="mt-5 border-t border-border pt-4">
                <h3 className="section-title mb-3">Thông tin công tác</h3>
                <ProfileFields
                  fields={[
                    ["Email công ty", selected.companyEmail],
                    ["Phòng ban", display("departmentId", selected.departmentId)],
                    ["Vị trí công việc", display("positionId", selected.positionId)],
                    ["Cấp bậc", nameOf("jobLevels", selected.jobLevelId)],
                  ]}
                />
              </section>
              <section className="mt-6 border-t border-border pt-4">
                <h3 className="section-title">Lịch sử HR cập nhật</h3>
                <p className="mt-1 text-xs text-muted-foreground">
                  Các lần HR chỉnh sửa trực tiếp trong phiên mô phỏng.
                </p>
                {state.edits
                  .filter((edit) => edit.employeeId === selected.personId)
                  .slice()
                  .reverse()
                  .map((edit) => (
                    <details key={edit.id} className="mt-3 rounded-lg border border-border p-3">
                      <summary className="cursor-pointer text-sm">
                        <span className="font-medium">
                          {getPerson(edit.actorId)?.name ?? edit.actorId}
                        </span>{" "}
                        · {new Date(edit.at).toLocaleString("vi-VN")}
                      </summary>
                      <dl className="mt-3 space-y-3 text-sm">
                        {employeeEditFields
                          .filter((field) => field in edit.after)
                          .map((field) => (
                            <div key={field}>
                              <dt className="field-label">{employeeEditLabels[field]}</dt>
                              <dd className="mt-1 whitespace-pre-wrap break-words">
                                <span className="text-muted-foreground">
                                  {edit.beforeDisplay[field] ??
                                    (field === "departmentId"
                                      ? "Chưa thuộc phòng ban"
                                      : "Chưa cập nhật")}
                                </span>{" "}
                                →{" "}
                                {edit.afterDisplay[field] ??
                                  (field === "departmentId"
                                    ? "Chưa thuộc phòng ban"
                                    : "Chưa cập nhật")}
                              </dd>
                            </div>
                          ))}
                      </dl>
                    </details>
                  ))}
                {!state.edits.some((edit) => edit.employeeId === selected.personId) && (
                  <p className="mt-3 text-sm text-muted-foreground">
                    Chưa có thay đổi trực tiếp từ HR trong phiên này.
                  </p>
                )}
              </section>
            </>
          )}
          {deactivating && (
            <EmployeeDeactivationDialog
              profile={selected}
              profiles={state.profiles}
              projects={projects}
              tasks={tasks}
              onConfirm={onDeactivate}
              onClose={() => setDeactivating(false)}
            />
          )}
          {editing && (
            <EmployeeEditForm
              profile={selected}
              profiles={state.profiles}
              organization={organization}
              onSave={onSave}
              onClose={() => setEditing(false)}
            />
          )}
        </Dialog>
      )}
    </div>
  )
}
