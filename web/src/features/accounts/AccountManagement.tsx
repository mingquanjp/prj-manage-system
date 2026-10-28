import { useState } from "react"
import { Pencil1Icon, LockClosedIcon, LockOpen1Icon, PlusIcon } from "@radix-ui/react-icons"
import { useIdentity } from "@/components/identity"
import { usePeople } from "@/components/people-directory"
import { Button } from "@/components/ui/button"
import { EmptyState, SearchInput } from "@/components/shared"
import { isAdmin } from "@/demo/identity"
import { searchKey } from "@/demo/personnel"
import type {
  AccountsState,
  DemoAccount,
  AccountCreateInput,
  AccountCommand,
} from "@/models/accounts"
import { systemRoleLabels } from "@/models/accounts"
import type { ProfileState } from "@/models/profile"
import type { OrganizationState } from "@/models/organization"
import type { Project, Task } from "@/models/workspace"
import { AccountCreateForm } from "./AccountCreateForm"
import { AccountChangeDialog } from "./AccountChangeDialog"

export function AccountManagement({
  actorId,
  accounts,
  profiles,
  organization,
  projects,
  tasks,
  onCreate,
  onChange,
}: {
  actorId: string
  accounts: AccountsState
  profiles: ProfileState
  organization: OrganizationState
  projects: Project[]
  tasks: Task[]
  onCreate: (input: AccountCreateInput) => void
  onChange: (command: AccountCommand) => void
}) {
  const identity = useIdentity()
  const { getPerson } = usePeople()
  const [query, setQuery] = useState("")
  const [status, setStatus] = useState("")
  const [role, setRole] = useState("")
  const [creating, setCreating] = useState(false)
  const [editing, setEditing] = useState<{
    account: DemoAccount
    action: "roles" | "lock" | "unlock"
  } | null>(null)
  if (!isAdmin(actorId, identity))
    return (
      <EmptyState
        title="Không có quyền quản lý tài khoản"
        detail="Màn hình này dành cho Quản trị viên đang hoạt động."
      />
    )
  const filtered = accounts.accounts.filter((a) => {
    const profile = profiles.profiles.find((p) => p.personId === a.personId)
    return (
      (!status || a.status === status) &&
      (!role ||
        (role === "EMPLOYEE"
          ? !a.systemRoles.length
          : a.systemRoles.includes(role as keyof typeof systemRoleLabels))) &&
      searchKey(`${a.username} ${profile?.employeeCode} ${profile?.fullname}`).includes(
        searchKey(query),
      )
    )
  })
  const openChange = (account: DemoAccount, action: "roles" | "lock" | "unlock") =>
    setEditing({ account: structuredClone(account), action })
  return (
    <div>
      <div className="mb-6 flex flex-wrap items-start justify-between gap-4">
        <div>
          <p className="eyebrow">QUẢN TRỊ HỆ THỐNG</p>
          <h1 className="mt-2 text-2xl font-semibold">Quản lý tài khoản</h1>
          <p className="mt-2 text-sm leading-6 text-muted-foreground">
            Tạo tài khoản kèm hồ sơ ban đầu, quản lý quyền truy cập. HR quản lý hồ sơ về sau.
          </p>
        </div>
        <Button onClick={() => setCreating(true)}>
          <PlusIcon /> Tạo tài khoản
        </Button>
      </div>
      <div className="mb-5 flex flex-wrap items-end gap-3">
        <div className="min-w-52 flex-1">
          <label className="field-label mb-2 block" htmlFor="account-query">
            Tìm tài khoản
          </label>
          <SearchInput
            id="account-query"
            value={query}
            onChange={setQuery}
            placeholder="Tên đăng nhập, mã hoặc tên nhân viên…"
          />
        </div>
        <div>
          <label className="field-label mb-2 block" htmlFor="account-status">
            Trạng thái tài khoản
          </label>
          <select
            id="account-status"
            className="field"
            value={status}
            onChange={(e) => setStatus(e.target.value)}
          >
            <option value="">Tất cả trạng thái</option>
            <option value="ACTIVE">Đang hoạt động</option>
            <option value="LOCKED">Đã khóa</option>
          </select>
        </div>
        <div>
          <label className="field-label mb-2 block" htmlFor="account-role">
            Quyền hệ thống
          </label>
          <select
            id="account-role"
            className="field"
            value={role}
            onChange={(e) => setRole(e.target.value)}
          >
            <option value="">Tất cả quyền</option>
            <option value="EMPLOYEE">Nhân viên — quyền cơ bản</option>
            {Object.entries(systemRoleLabels).map(([id, name]) => (
              <option key={id} value={id}>
                {name}
              </option>
            ))}
          </select>
        </div>
      </div>
      <div className="overflow-x-auto rounded-xl border border-border bg-white">
        <table className="w-full min-w-[740px] text-left text-sm">
          <thead className="table-heading border-b border-border bg-sidebar">
            <tr>
              {["Tài khoản", "Nhân viên", "Quyền hệ thống", "Trạng thái", "Thao tác"].map(
                (text) => (
                  <th key={text} className="px-4 py-3">
                    {text}
                  </th>
                ),
              )}
            </tr>
          </thead>
          <tbody>
            {filtered.map((account) => {
              const profile = profiles.profiles.find((p) => p.personId === account.personId)
              return (
                <tr
                  key={account.id}
                  className="border-b border-border last:border-0 hover:bg-accent/30"
                >
                  <td className="px-4 py-4 font-medium">{account.username}</td>
                  <td className="px-4 py-4">
                    <p className="font-medium">{profile?.fullname}</p>
                    <p className="mt-1 text-xs text-muted-foreground">{profile?.employeeCode}</p>
                    {profile?.employmentStatus === "INACTIVE" && (
                      <p className="mt-1 text-xs text-warning">Hồ sơ đã vô hiệu hóa</p>
                    )}
                  </td>
                  <td className="px-4 py-4">
                    <div className="flex flex-wrap gap-1">
                      {account.systemRoles.length ? (
                        account.systemRoles.map((r) => (
                          <span key={r} className="rounded bg-accent px-2 py-1 text-xs">
                            {systemRoleLabels[r]}
                          </span>
                        ))
                      ) : (
                        <span className="text-muted-foreground">Nhân viên</span>
                      )}
                    </div>
                  </td>
                  <td className="px-4 py-4 whitespace-nowrap">
                    {account.status === "ACTIVE" ? "Đang hoạt động" : "Đã khóa"}
                  </td>
                  <td className="px-4 py-4">
                    <div className="flex gap-1">
                      <Button
                        size="icon"
                        variant="ghost"
                        aria-label={`Sửa quyền ${account.username}`}
                        title="Sửa quyền"
                        onClick={() => openChange(account, "roles")}
                      >
                        <Pencil1Icon />
                      </Button>
                      <Button
                        size="icon"
                        variant="ghost"
                        aria-label={`${account.status === "ACTIVE" ? "Khóa" : "Mở khóa"} ${account.username}`}
                        title={account.status === "ACTIVE" ? "Khóa tài khoản" : "Mở khóa tài khoản"}
                        onClick={() =>
                          openChange(account, account.status === "ACTIVE" ? "lock" : "unlock")
                        }
                      >
                        {account.status === "ACTIVE" ? <LockClosedIcon /> : <LockOpen1Icon />}
                      </Button>
                    </div>
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
        {!filtered.length && (
          <EmptyState title="Không tìm thấy tài khoản" detail="Thử thay đổi từ khóa hoặc bộ lọc." />
        )}
      </div>
      <p className="mt-3 text-xs leading-5 text-muted-foreground">
        {filtered.length} / {accounts.accounts.length} tài khoản · Bản mô phỏng chưa cấp hoặc gửi
        thông tin đăng nhập thật.
      </p>
      <details className="mt-6 rounded-lg border border-border bg-white p-4">
        <summary className="cursor-pointer text-sm font-semibold">
          Lịch sử quản trị trong phiên ({accounts.events.length})
        </summary>
        {accounts.events.length ? (
          <ul className="mt-3 space-y-3 text-xs leading-5">
            {accounts.events
              .slice()
              .reverse()
              .map((event) => (
                <li key={event.id} className="border-t border-border pt-2">
                  <p className="font-medium">
                    {
                      {
                        CREATED: "Tạo tài khoản",
                        ROLES_CHANGED: "Đổi quyền",
                        LOCKED: "Khóa tài khoản",
                        UNLOCKED: "Mở khóa tài khoản",
                      }[event.action]
                    }{" "}
                    · {accounts.accounts.find((a) => a.id === event.accountId)?.username}
                  </p>
                  <p className="text-muted-foreground">
                    {getPerson(event.actorId)?.name} · {new Date(event.at).toLocaleString("vi-VN")}
                  </p>
                  <p>
                    {event.before &&
                      `${event.before.status === "ACTIVE" ? "Hoạt động" : "Đã khóa"} / ${event.before.systemRoles.map((r) => systemRoleLabels[r]).join(", ") || "Nhân viên"} → `}
                    {event.after.status === "ACTIVE" ? "Hoạt động" : "Đã khóa"} /{" "}
                    {event.after.systemRoles.map((r) => systemRoleLabels[r]).join(", ") ||
                      "Nhân viên"}
                  </p>
                </li>
              ))}
          </ul>
        ) : (
          <p className="mt-3 text-sm text-muted-foreground">
            Chưa có thao tác quản trị trong phiên này.
          </p>
        )}
      </details>
      {creating && (
        <AccountCreateForm
          accounts={accounts}
          profiles={profiles}
          organization={organization}
          onSave={onCreate}
          onClose={() => setCreating(false)}
        />
      )}
      {editing && (
        <AccountChangeDialog
          account={editing.account}
          action={editing.action}
          fullname={getPerson(editing.account.personId)?.name ?? ""}
          isEmployeeActive={
            profiles.profiles.find((p) => p.personId === editing.account.personId)
              ?.employmentStatus === "ACTIVE"
          }
          projects={projects}
          tasks={tasks}
          onSave={onChange}
          onClose={() => setEditing(null)}
        />
      )}
    </div>
  )
}
