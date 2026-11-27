import { useEffect, useState, useSyncExternalStore } from "react"
import { Link, useLocation, useNavigate } from "react-router"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import type { AuthSession } from "../session/auth-session"
import { parseAdministrativeLocation } from "../session/server-navigation"
import { createAdministrativeProfileApi } from "./administrative-profile-api"
import { createAdministrativeProfileResource } from "./administrative-profile-resource"
import { AdministrativeProfileForm } from "./AdministrativeProfileForm"
import { ProfileFields } from "./ProfileFields"
const root = "/personnel/employees"
const statuses = { ACTIVE: "Đang làm việc", INACTIVE: "Ngừng hoạt động", RESIGNED: "Đã nghỉ việc" }
export function ServerAdministrativeProfiles({
  session,
  busy,
}: {
  session: AuthSession
  busy: boolean
}) {
  const location = useLocation(),
    navigate = useNavigate()
  const parsed = parseAdministrativeLocation(location.pathname, location.search)
  const [resource] = useState(() =>
    createAdministrativeProfileResource(createAdministrativeProfileApi(), session),
  )
  const state = useSyncExternalStore(resource.subscribe, resource.getSnapshot)
  const [editing, setEditing] = useState(false)
  const [dirty, setDirty] = useState(false)
  const [discard, setDiscard] = useState(false)
  const [search, setSearch] = useState(parsed?.query.q ?? "")
  const route = `${location.pathname}${location.search}`
  useEffect(() => {
    const value = parseAdministrativeLocation(location.pathname, location.search)
    document.getElementById("personnel-heading")?.focus()
    if (value) void resource.load(value.id ?? value.query)
    return resource.clear
  }, [resource, location.pathname, location.search])
  const locked = busy || state.saving
  const reload = () => {
    if (locked || !parsed) return
    setEditing(false)
    document.getElementById("personnel-heading")?.focus()
    void resource.load(parsed.id ?? parsed.query)
  }
  const pageUrl = (page: number, q = parsed?.query.q ?? "") =>
    `${root}?${new URLSearchParams({ q, page: String(page), page_size: String(parsed?.query.pageSize ?? 20) })}`
  if (!parsed)
    return (
      <section>
        <h1 className="text-2xl font-semibold">Đường dẫn hồ sơ chưa hợp lệ</h1>
        <Link className="underline" to={root}>
          Về danh sách hồ sơ
        </Link>
      </section>
    )
  const detail = state.detail
  return (
    <section className="space-y-5">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h1 id="personnel-heading" tabIndex={-1} className="text-2xl font-semibold">
          {parsed.id ? "Chi tiết hồ sơ nhân sự" : "Hồ sơ nhân sự"}
        </h1>
        <Button variant="outline" disabled={locked || editing} onClick={reload}>
          Tải lại dữ liệu
        </Button>
      </div>
      {parsed.id && (
        <Link className="inline-block text-sm underline" to={root}>
          Về danh sách hồ sơ
        </Link>
      )}
      {state.message && (
        <p role="alert" className="rounded-lg bg-accent p-3 text-sm">
          {state.message}
        </p>
      )}
      {state.mustReload && (
        <p className="text-sm">
          Đóng chỉnh sửa và tải lại để đối soát. Yêu cầu không được tự gửi lại.
        </p>
      )}
      {state.phase === "loading" && <p role="status">Đang tải hồ sơ từ máy chủ…</p>}
      {!parsed.id && (
        <form
          className="flex flex-wrap gap-2"
          onSubmit={(e) => {
            e.preventDefault()
            navigate(pageUrl(1, search))
          }}
        >
          <label className="min-w-0 flex-1">
            <span className="sr-only">Tìm theo mã, họ tên hoặc email</span>
            <input
              className="field workspace-field w-full"
              placeholder="Tìm mã, họ tên hoặc email"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </label>
          <Button type="submit" disabled={locked || [...search].length > 100}>
            Tìm kiếm
          </Button>
        </form>
      )}
      {state.page && (
        <>
          <div className="overflow-x-auto rounded-xl border border-border bg-white">
            <table className="w-full min-w-[720px] text-left text-sm">
              <caption className="sr-only">Danh sách hồ sơ hành chính</caption>
              <thead className="table-heading">
                <tr>
                  {[
                    "Nhân viên",
                    "Email công ty",
                    "Phòng ban",
                    "Vị trí",
                    "Cấp bậc",
                    "Trạng thái",
                    "Chi tiết",
                  ].map((label) => (
                    <th scope="col" className="px-4 py-3" key={label}>
                      {label}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {state.page.items.map((item) => (
                  <tr className="border-t border-border" key={item.employee_id}>
                    <th scope="row" className="px-4 py-3 font-medium">
                      {item.fullname}
                      <span className="block text-xs text-muted-foreground">
                        {item.employee_code}
                      </span>
                    </th>
                    <td className="px-4 py-3">{item.company_email}</td>
                    <td className="px-4 py-3">{item.department?.name ?? "Chưa cập nhật"}</td>
                    <td className="px-4 py-3">{item.position.name}</td>
                    <td className="px-4 py-3">{item.job_level?.name ?? "Chưa cập nhật"}</td>
                    <td className="px-4 py-3">{statuses[item.employment_status]}</td>
                    <td className="px-4 py-3">
                      <Link
                        className="underline"
                        to={`${root}/${item.employee_id}`}
                        aria-label={`Xem hồ sơ ${item.fullname}`}
                      >
                        Xem hồ sơ
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {!state.page.items.length && <p>Không có hồ sơ ở trang và từ khóa này.</p>}
          <div className="flex flex-wrap items-center gap-3 text-sm">
            <span>
              Trang {state.page.page} · Tổng {state.page.total} hồ sơ
            </span>
            <Button
              variant="outline"
              disabled={locked || parsed.query.page <= 1}
              onClick={() => navigate(pageUrl(parsed.query.page - 1))}
            >
              Trang trước
            </Button>
            <Button
              variant="outline"
              disabled={
                locked ||
                parsed.query.page * parsed.query.pageSize >= state.page.total ||
                parsed.query.page === 2147483647
              }
              onClick={() => navigate(pageUrl(parsed.query.page + 1))}
            >
              Trang sau
            </Button>
          </div>
        </>
      )}
      {detail && (
        <div className="space-y-5 rounded-xl border border-border bg-white p-5">
          <h2 className="text-lg font-semibold">
            {detail.fullname} · {detail.employee_code}
          </h2>
          <ProfileFields
            fields={[
              ["Email công ty", detail.company_email],
              ["Số điện thoại", detail.phone],
              ["Địa chỉ", detail.address],
              ["Phòng ban", detail.department?.name ?? null],
              ["Vị trí", detail.position.name],
              ["Cấp bậc", detail.job_level?.name ?? null],
              ["Trạng thái", statuses[detail.employment_status]],
            ]}
          />
          {detail.address === null && (
            <p className="rounded-lg bg-warning-soft p-3 text-sm">
              Địa chỉ chưa cập nhật. HR có thể bổ sung khi có thông tin; không cản trở đăng nhập
              hoặc sửa trường khác.
            </p>
          )}
          <Button
            disabled={locked || state.mustReload || !state.catalogs}
            onClick={() => {
              setDirty(false)
              setEditing(true)
            }}
          >
            Chỉnh sửa hồ sơ
          </Button>
          <p className="text-xs text-muted-foreground">
            Thay đổi hành chính được lưu audit phía máy chủ. Chưa kết nối giao diện đọc lịch sử
            audit.
          </p>
        </div>
      )}
      <Dialog
        open={editing && !!detail}
        onOpenChange={(open) => {
          if (!locked) {
            if (!open && dirty) setDiscard(true)
            else setEditing(open)
          }
        }}
        title="Chỉnh sửa hồ sơ hành chính"
        description="Cập nhật trực tiếp dữ liệu chính thức; không tạo đề nghị nhân viên."
        onCloseAutoFocus={(event) => {
          event.preventDefault()
          document.getElementById("personnel-heading")?.focus()
        }}
      >
        {detail && state.catalogs && (
          <div onChange={() => setDirty(true)}>
            <AdministrativeProfileForm
              key={route}
              profile={detail}
              catalogs={state.catalogs}
              locked={locked || state.mustReload}
              fieldError={state.fieldError}
              onCancel={() => {
                if (dirty) setDiscard(true)
                else setEditing(false)
              }}
              onSave={async (patch) => {
                const saved = await resource.save(patch)
                if (saved) setEditing(false)
                return saved
              }}
            />
          </div>
        )}
        {state.message && (
          <p role="alert" className="mt-4 text-sm">
            {state.message}
          </p>
        )}
        {state.mustReload && (
          <Button
            type="button"
            variant="outline"
            className="mt-3"
            onClick={() => {
              setEditing(false)
              void resource.load(parsed.id ?? parsed.query)
            }}
          >
            Đóng và tải lại để đối soát
          </Button>
        )}
      </Dialog>
      <Dialog
        open={discard}
        onOpenChange={setDiscard}
        title="Bỏ thay đổi chưa lưu?"
        description="Bản nháp sẽ bị xóa; hồ sơ trên máy chủ giữ nguyên."
      >
        <div className="flex flex-wrap gap-3">
          <Button variant="outline" onClick={() => setDiscard(false)}>
            Tiếp tục chỉnh sửa
          </Button>
          <Button
            onClick={() => {
              setDiscard(false)
              setEditing(false)
              setDirty(false)
            }}
          >
            Bỏ thay đổi
          </Button>
        </div>
      </Dialog>
    </section>
  )
}
