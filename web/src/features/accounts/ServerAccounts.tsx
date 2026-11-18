import { useEffect, useState } from "react"
import { Link, useLocation, useNavigate } from "react-router"
import { Button } from "@/components/ui/button"
import { SearchInput } from "@/components/shared"
import { AuthError } from "../session/auth-api"
import type { AuthSession } from "../session/auth-session"
import { parseCatalogQuery } from "../session/server-navigation"
import { systemRoleLabels } from "@/models/accounts"
import {
  accountErrorMessage,
  createAccountApi,
  mailMessages,
  type AccountCreated,
  type AccountPage,
} from "./account-api"
import { ServerAccountCreateForm } from "./ServerAccountCreateForm"

export function ServerAccounts({ session, busy }: { session: AuthSession; busy: boolean }) {
  const [api] = useState(createAccountApi)
  const location = useLocation()
  const navigate = useNavigate()
  const params = parseCatalogQuery(location.search)
  const invalidQuery = params === null
  const q = params?.q ?? "",
    page = params?.page ?? 1,
    pageSize = params?.pageSize ?? 20
  const [draft, setDraft] = useState(q)
  const [data, setData] = useState<AccountPage>()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [revision, setRevision] = useState(0)
  const [creating, setCreating] = useState(false)
  const [result, setResult] = useState<AccountCreated>()
  const [uncertain, setUncertain] = useState(false)
  useEffect(() => {
    document.getElementById("accounts-heading")?.focus()
  }, [])
  useEffect(() => {
    let current = true
    if (invalidQuery) return
    setLoading(true)
    setData(undefined)
    setError("")
    setDraft(q)
    void api
      .list(q, page, pageSize)
      .then((value) => {
        if (current) {
          setData(value)
          setLoading(false)
        }
      })
      .catch((issue: unknown) => {
        if (!current) return
        setError(accountErrorMessage(issue))
        setLoading(false)
        if (issue instanceof AuthError && [401, 403].includes(issue.status)) void session.refresh()
      })
    return () => {
      current = false
    }
  }, [api, session, q, page, pageSize, revision, invalidQuery])
  function reload() {
    setData(undefined)
    setLoading(true)
    setError("")
    setRevision((n) => n + 1)
  }
  function go(query: string, nextPage = 1, size = pageSize) {
    navigate({
      pathname: "/accounts",
      search: `?${new URLSearchParams({ q: query, page: String(nextPage), page_size: String(size) })}`,
    })
  }
  if (!params)
    return (
      <section>
        <h1 className="text-xl">Địa chỉ danh sách chưa hợp lệ</h1>
        <Button asChild className="mt-4">
          <Link to="/accounts">Mở lại danh sách</Link>
        </Button>
      </section>
    )
  return (
    <section className="min-w-0 max-w-5xl" aria-busy={loading}>
      <div className="mb-6 flex flex-wrap items-start justify-between gap-4">
        <div>
          <p className="eyebrow">QUẢN TRỊ HỆ THỐNG</p>
          <h1 id="accounts-heading" tabIndex={-1} className="mt-2 text-2xl font-semibold">
            Quản lý tài khoản
          </h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Tạo tài khoản kèm hồ sơ và tra cứu thông tin đăng nhập bằng email.
          </p>
        </div>
        <Button
          id="account-create"
          disabled={busy || creating}
          onClick={() => {
            setResult(undefined)
            setCreating(true)
          }}
        >
          Tạo tài khoản
        </Button>
      </div>
      {result && (
        <div
          role="status"
          className="mb-5 space-y-2 rounded-xl border border-border bg-accent p-4 text-sm"
        >
          <h2 className="font-semibold">Tài khoản đã tạo</h2>
          <p className="break-words">
            {result.account.fullname} · {result.account.company_email}
          </p>
          <p>
            Mã chính thức: <strong>{result.account.employee_code}</strong>
          </p>
          <p>{mailMessages[result.mail.status]}</p>
          <p className="text-xs text-muted-foreground">
            Kết quả gửi email chỉ có trong phản hồi này; tải lại trang không khôi phục được thông
            tin gửi thư.
          </p>
        </div>
      )}
      {uncertain && (
        <p role="alert" className="mb-5 rounded-xl bg-accent p-4 text-sm leading-6">
          Yêu cầu trước chưa xác nhận kết quả. Tìm email để đối soát; không tạo lại tài khoản đó.
          Tài khoản tồn tại không chứng minh kết quả gửi thư.
        </p>
      )}
      <form
        className="mb-5 flex max-w-xl items-end gap-2"
        onSubmit={(event) => {
          event.preventDefault()
          go(draft)
        }}
      >
        <div className="min-w-0 flex-1">
          <label htmlFor="account-query" className="field-label mb-2 block">
            Tìm tài khoản
          </label>
          <SearchInput
            id="account-query"
            value={draft}
            onChange={setDraft}
            placeholder="Mã nhân viên, họ tên hoặc email…"
          />
        </div>
        <Button type="submit" variant="outline" disabled={busy}>
          Tìm
        </Button>
      </form>
      {error && (
        <p role="alert" className="mb-4 text-sm text-warning">
          {error}
        </p>
      )}
      <div className="overflow-x-auto rounded-xl border border-border bg-white">
        {loading ? (
          <p role="status" className="p-6">
            Đang tải tài khoản…
          </p>
        ) : (
          data && (
            <>
              <table className="w-full min-w-[680px] text-left text-sm">
                <thead className="table-heading border-b border-border bg-sidebar">
                  <tr>
                    {["Mã nhân viên", "Họ tên / Email", "Quyền hệ thống", "Trạng thái"].map(
                      (label) => (
                        <th className="px-4 py-3" key={label}>
                          {label}
                        </th>
                      ),
                    )}
                  </tr>
                </thead>
                <tbody>
                  {data.items.map((account) => (
                    <tr key={account.user_id} className="border-b border-border last:border-0">
                      <td className="px-4 py-4 font-medium">{account.employee_code}</td>
                      <td className="max-w-72 break-words px-4 py-4">
                        <p>{account.fullname}</p>
                        <p className="mt-1 text-xs text-muted-foreground">
                          {account.company_email}
                        </p>
                      </td>
                      <td className="px-4 py-4">
                        {account.system_roles.map((role) => systemRoleLabels[role]).join(", ") ||
                          "Nhân viên"}
                      </td>
                      <td className="px-4 py-4">
                        {
                          {
                            ACTIVE: "Đang hoạt động",
                            LOCKED: "Đã khóa",
                            DISABLED: "Đã vô hiệu hóa",
                          }[account.status]
                        }
                        {account.must_change_password && (
                          <p className="mt-1 text-xs text-muted-foreground">
                            Cần đổi mật khẩu lần đầu
                          </p>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              {!data.items.length && (
                <p className="p-6 text-sm">
                  Không có tài khoản trên trang này. Thử từ khóa khác hoặc về trang đầu.
                </p>
              )}
            </>
          )
        )}
      </div>
      <div className="mt-4 flex flex-wrap items-center gap-3 text-sm">
        <Button variant="outline" disabled={loading || busy} onClick={reload}>
          Tải lại danh sách
        </Button>
        {data && (
          <>
            <p>
              Trang {page} · Tổng {data.pagination.total} tài khoản
            </p>
            <Button
              variant="outline"
              disabled={page === 1 || loading || busy}
              onClick={() => go(q, page - 1)}
            >
              Trang trước
            </Button>
            <Button
              variant="outline"
              disabled={page * pageSize >= data.pagination.total || loading || busy}
              onClick={() => go(q, page + 1)}
            >
              Trang sau
            </Button>
          </>
        )}
        <label className="flex items-center gap-2">
          Số mục/trang
          <select
            className="field"
            value={pageSize}
            disabled={busy}
            onChange={(event) => go(q, 1, Number(event.target.value))}
          >
            {[...new Set([10, 20, 50, 100, pageSize])]
              .sort((a, b) => a - b)
              .map((size) => (
                <option key={size} value={size}>
                  {size}
                </option>
              ))}
          </select>
        </label>
      </div>
      <p className="mt-4 text-xs leading-5 text-muted-foreground">
        Danh sách không lưu kết quả gửi email. Không hỗ trợ gửi lại, khóa hoặc chỉnh quyền tài khoản
        ở màn này.
      </p>
      {creating && (
        <ServerAccountCreateForm
          session={session}
          busy={busy}
          onClose={() => setCreating(false)}
          onCreated={(value) => {
            setResult(value)
            setCreating(false)
            reload()
          }}
          onReconcile={(email) => {
            setCreating(false)
            setUncertain(true)
            setDraft(email)
            go(email)
            reload()
          }}
        />
      )}
    </section>
  )
}
