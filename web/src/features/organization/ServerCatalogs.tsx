import { useEffect, useRef, useState } from "react"
import { Link, useLocation, useNavigate } from "react-router"
import { PlusIcon } from "@radix-ui/react-icons"
import { Button } from "@/components/ui/button"
import { EmptyState, SearchInput } from "@/components/shared"
import { AuthError } from "../session/auth-api"
import type { AuthSession } from "../session/auth-session"
import { parseCatalogQuery } from "../session/server-navigation"
import { CatalogTable } from "./CatalogTable"
import { ServerCatalogDialog, type CatalogEditor } from "./ServerCatalogDialog"
import {
  catalogErrorMessage,
  catalogKinds,
  catalogLabels,
  createCatalogApi,
  type CatalogKind,
  type CatalogPage,
} from "./catalog-api"

export function ServerCatalogs({
  kind,
  canWrite,
  session,
  busy,
}: {
  kind: CatalogKind
  canWrite: boolean
  session: AuthSession
  busy: boolean
}) {
  const [api] = useState(createCatalogApi)
  const location = useLocation()
  const navigate = useNavigate()
  const params = parseCatalogQuery(location.search)
  const q = params?.q ?? "",
    page = params?.page ?? 1,
    pageSize = params?.pageSize ?? 20
  const invalidQuery = params === null
  const [result, setResult] = useState<CatalogPage>()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [notice, setNotice] = useState("")
  const [revision, setRevision] = useState(0)
  const [editor, setEditor] = useState<CatalogEditor>()
  const [searchDraft, setSearchDraft] = useState(q)
  const returnFocus = useRef<HTMLElement | null>(null)
  const reloadButton = useRef<HTMLButtonElement | null>(null)
  useEffect(() => {
    document.getElementById("catalog-heading")?.focus()
  }, [])
  useEffect(() => {
    let current = true
    if (invalidQuery) return
    void api
      .list(kind, q, page, pageSize)
      .then((value) => {
        if (current) {
          setResult(value)
          setLoading(false)
        }
      })
      .catch((error: unknown) => {
        if (!current) return
        setError(catalogErrorMessage(error))
        setLoading(false)
        if (error instanceof AuthError && (error.status === 401 || error.status === 403))
          void session.refresh()
      })
    return () => {
      current = false
    }
  }, [api, kind, q, page, pageSize, revision, session, invalidQuery])
  function reload() {
    setEditor(undefined)
    setLoading(true)
    setError("")
    setResult(undefined)
    setRevision((value) => value + 1)
    returnFocus.current = document.getElementById("catalog-heading")
  }
  function go(next: { q: string; page: number; pageSize: number }) {
    const search = new URLSearchParams({
      q: next.q,
      page: String(next.page),
      page_size: String(next.pageSize),
    })
    navigate({ pathname: `/organization/${kind}`, search: `?${search}` })
  }
  function open(value: CatalogEditor) {
    returnFocus.current =
      document.activeElement instanceof HTMLElement ? document.activeElement : reloadButton.current
    setNotice("")
    setEditor(value)
  }
  if (invalidQuery)
    return (
      <section>
        <h1 className="text-2xl font-semibold">Địa chỉ danh mục chưa hợp lệ</h1>
        <p className="my-4 text-sm">Hãy mở lại danh mục để đặt lại tìm kiếm và phân trang.</p>
        <Button asChild>
          <Link to={`/organization/${kind}`}>Mở lại danh mục</Link>
        </Button>
      </section>
    )
  return (
    <section className="min-w-0 max-w-5xl" aria-busy={loading}>
      <div className="mb-6 flex flex-wrap items-start justify-between gap-4">
        <div>
          <p className="eyebrow">TỔ CHỨC · {canWrite ? "HR" : "CHỈ ĐỌC"}</p>
          <h1 id="catalog-heading" tabIndex={-1} className="mt-2 text-2xl font-semibold">
            Danh mục tổ chức
          </h1>
          <p className="mt-2 text-sm text-muted-foreground">
            {kind === "job-levels"
              ? "Cấp bậc theo cơ cấu công ty, không thay thế vị trí, mức kỹ năng hoặc quyền truy cập."
              : "Danh mục dùng chung cho thông tin công tác của nhân viên."}
          </p>
        </div>
        {canWrite && (
          <Button disabled={busy || loading || !result} onClick={() => open({ action: "create" })}>
            <PlusIcon />
            Thêm {catalogLabels[kind].toLowerCase()}
          </Button>
        )}
      </div>
      <div className="mb-5 flex flex-wrap items-end justify-between gap-4">
        <nav aria-label="Chọn danh mục tổ chức" className="flex flex-wrap gap-1">
          {catalogKinds.map((value) => (
            <Link
              key={value}
              className={`filter-chip ${kind === value ? "filter-active" : ""}`}
              aria-current={kind === value ? "page" : undefined}
              to={`/organization/${value}`}
            >
              {catalogLabels[value]}
            </Link>
          ))}
        </nav>
        <form
          className="flex w-full items-center gap-2 sm:w-96"
          onSubmit={(event) => {
            event.preventDefault()
            go({ q: searchDraft, page: 1, pageSize })
          }}
        >
          <div className="min-w-0 flex-1">
            <SearchInput
              value={searchDraft}
              onChange={setSearchDraft}
              placeholder={`Tìm ${catalogLabels[kind].toLowerCase()}…`}
            />
          </div>
          <Button variant="outline" type="submit">
            Tìm
          </Button>
        </form>
      </div>
      {notice && (
        <p role="status" className="mb-4 rounded-lg bg-accent p-3 text-sm">
          {notice}
        </p>
      )}
      {error && (
        <p role="alert" className="mb-4 text-sm text-warning">
          {error}
        </p>
      )}
      <div className="overflow-x-auto rounded-xl border border-border bg-white">
        {loading ? (
          <p role="status" className="p-8 text-sm">
            Đang tải danh mục…
          </p>
        ) : (
          result && (
            <>
              <CatalogTable
                entries={result.items}
                label={`Tên ${catalogLabels[kind].toLowerCase()}`}
                disabled={busy}
                onEdit={canWrite ? (entry) => open({ action: "rename", entry }) : undefined}
                onRemove={canWrite ? (entry) => open({ action: "delete", entry }) : undefined}
              />
              {!result.items.length && (
                <EmptyState
                  title={q ? "Không tìm thấy kết quả" : "Trang này chưa có dữ liệu"}
                  detail="Thử từ khóa khác, trở về trang đầu hoặc thêm mục mới nếu có quyền."
                />
              )}
            </>
          )
        )}
      </div>
      <div className="mt-4 flex flex-wrap items-center gap-3 text-sm">
        <Button
          id="catalog-reload"
          ref={reloadButton}
          variant="outline"
          disabled={loading || busy}
          onClick={reload}
        >
          Tải lại danh sách
        </Button>
        {result && (
          <>
            <p role="status" className="text-muted-foreground">
              Trang {page} · Tổng {result.pagination.total} mục
            </p>
            <Button
              variant="outline"
              disabled={page === 1 || loading}
              onClick={() => go({ q, page: page - 1, pageSize })}
            >
              Trang trước
            </Button>
            <Button
              variant="outline"
              disabled={page * pageSize >= result.pagination.total || loading}
              onClick={() => go({ q, page: page + 1, pageSize })}
            >
              Trang sau
            </Button>
          </>
        )}
        <label className="ml-auto flex items-center gap-2">
          Số mục/trang
          <select
            className="field"
            value={pageSize}
            onChange={(event) => go({ q, page: 1, pageSize: Number(event.target.value) })}
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
      {editor && canWrite && (
        <ServerCatalogDialog
          editor={editor}
          label={catalogLabels[kind]}
          busy={busy}
          returnFocus={returnFocus}
          onClose={() => setEditor(undefined)}
          onReload={reload}
          onSave={async (name) => {
            try {
              await session.catalogMutation((token) =>
                editor.action === "create"
                  ? api.create(kind, name, token)
                  : editor.action === "rename"
                    ? api.rename(kind, editor.entry, name, token)
                    : api.remove(kind, editor.entry, token),
              )
              setNotice(
                editor.action === "delete" ? "Đã xóa mục khỏi danh mục." : "Đã lưu danh mục.",
              )
              reload()
            } catch (error) {
              if (error instanceof AuthError && error.status === 403) void session.refresh()
              throw error
            }
          }}
        />
      )}
    </section>
  )
}
