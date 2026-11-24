import { useEffect, useRef, useState, useSyncExternalStore } from "react"
import * as DialogPrimitive from "@radix-ui/react-dialog"
import { BellIcon, CheckIcon, Cross2Icon } from "@radix-ui/react-icons"
import { useLocation, useNavigate } from "react-router"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import type { Principal } from "../session/auth-api"
import type { AuthSession } from "../session/auth-session"
import { createHrReviewApi } from "../profiles/hr-review-api"
import { createProfileRequestsApi } from "../profiles/profile-requests-api"
import { createServerNotificationsApi, type ServerNotification } from "./server-notifications-api"
import { createServerNotificationsResource } from "./server-notifications-resource"
import { createNotificationOpener } from "./notification-opener"

export function ServerNotificationMenu({
  principal,
  session,
  busy,
  refreshVersion,
}: {
  principal: Principal
  session: AuthSession
  busy: boolean
  refreshVersion: number
}) {
  const [resource] = useState(() =>
    createServerNotificationsResource(createServerNotificationsApi(), session),
  )
  const state = useSyncExternalStore(resource.subscribe, resource.getSnapshot)
  const [selfApi] = useState(() => createProfileRequestsApi())
  const [reviewApi] = useState(() => createHrReviewApi())
  const [isOpen, setIsOpen] = useState(false)
  const [openingId, setOpeningId] = useState<number | null>(null)
  const [openError, setOpenError] = useState("")
  const openedResource = useRef(false)
  const panelTitle = useRef<HTMLHeadingElement>(null)
  const navigate = useNavigate()
  const [opener] = useState(() =>
    createNotificationOpener({
      session,
      selfApi,
      reviewApi,
      onState: ({ openingId, message }) => {
        setOpeningId(openingId)
        setOpenError(message)
      },
      onOpen: (destination) => {
        openedResource.current = true
        setIsOpen(false)
        navigate(destination)
      },
      markRead: resource.markRead,
    }),
  )
  const location = useLocation()
  const actor = `${principal.user_id}:${principal.employee_id}`
  const roles = principal.system_roles.join(",")

  useEffect(() => {
    void resource.load()
    return () => resource.clear()
  }, [resource, actor, refreshVersion])
  useEffect(() => {
    return () => {
      opener.cancel()
      setIsOpen(false)
    }
  }, [opener, actor, roles, location.key])

  function changeOpen(open: boolean) {
    openedResource.current = false
    opener.cancel()
    setIsOpen(open)
    setOpenError("")
    if (open) void resource.load()
  }
  async function openNotification(item: ServerNotification) {
    if (busy || openingId !== null || state.readingId !== undefined) return
    await opener.open(item)
  }
  const unread = state.data?.unread_count
  const locked = busy || openingId !== null || state.readingId !== undefined
  return (
    <div className="relative">
      <DialogPrimitive.Root modal={false} open={isOpen} onOpenChange={changeOpen}>
        <DialogPrimitive.Trigger asChild>
          <Button
            size="sm"
            variant="ghost"
            aria-label={`Thông báo, ${unread === undefined ? "chưa tải số" : unread} chưa đọc`}
          >
            <BellIcon />
            <span className="hidden sm:inline">Thông báo</span>
            {unread !== undefined && unread > 0 && (
              <span className="rounded-full bg-primary px-1.5 text-xs text-white">{unread}</span>
            )}
          </Button>
        </DialogPrimitive.Trigger>
        <DialogPrimitive.Content
          aria-describedby={undefined}
          onCloseAutoFocus={(event) => {
            // Màn/chi tiết đích tự nhận focus; chỉ trả về chuông khi đóng khung thông thường.
            if (openedResource.current) event.preventDefault()
            openedResource.current = false
          }}
          className="fixed inset-x-4 top-16 z-40 flex max-h-[min(560px,calc(100dvh-80px))] w-auto max-w-[calc(100vw-2rem)] flex-col overflow-hidden rounded-xl border border-border bg-white shadow-xl outline-none sm:absolute sm:left-auto sm:right-0 sm:top-full sm:mt-2 sm:w-[420px]"
        >
          <header className="flex shrink-0 items-center justify-between px-4 pb-2 pt-3">
            <DialogPrimitive.Title
              ref={panelTitle}
              tabIndex={-1}
              className="text-base font-semibold"
            >
              Thông báo
            </DialogPrimitive.Title>
            <DialogPrimitive.Close asChild>
              <Button variant="ghost" size="icon" aria-label="Đóng thông báo">
                <Cross2Icon />
              </Button>
            </DialogPrimitive.Close>
          </header>
          <div
            role="group"
            aria-label="Lọc thông báo"
            className="flex shrink-0 flex-wrap gap-1 border-b border-border px-3 pb-3"
          >
            {(["all", "unread"] as const).map((filter) => (
              <button
                key={filter}
                className={cn("filter-chip", state.filter === filter && "filter-active")}
                aria-pressed={state.filter === filter}
                disabled={locked}
                onClick={() => {
                  setOpenError("")
                  void resource.load(filter, 1)
                }}
              >
                {filter === "all" ? "Tất cả" : "Chưa đọc"}
                {filter === "unread" && unread !== undefined ? ` (${unread})` : ""}
              </button>
            ))}
            <Button
              size="sm"
              variant="ghost"
              disabled={locked || state.phase === "loading"}
              onClick={() => {
                setOpenError("")
                void resource.load()
              }}
            >
              Tải lại
            </Button>
          </div>
          {(openError || state.message) && (
            <p role="alert" className="shrink-0 border-b border-border p-3 text-sm">
              {openError || state.message}
            </p>
          )}
          <div className="min-h-0 overflow-y-auto overscroll-contain">
            {state.phase === "loading" ? (
              <p role="status" className="p-5 text-sm">
                Đang tải thông báo…
              </p>
            ) : state.data ? (
              state.data.items.length ? (
                <ol className="divide-y divide-border">
                  {state.data.items.map((item) => (
                    <li
                      key={item.notification_id}
                      className={cn(
                        "flex items-start gap-3 p-4",
                        item.read_at === null && "bg-accent/40",
                      )}
                    >
                      <span
                        className={cn(
                          "mt-2 size-2 shrink-0 rounded-full",
                          item.read_at === null ? "bg-primary" : "bg-border",
                        )}
                        aria-label={item.read_at === null ? "Chưa đọc" : "Đã đọc"}
                      />
                      <div className="min-w-0 flex-1">
                        <p className="mb-1 text-xs font-medium text-muted-foreground">
                          Đề nghị hồ sơ #{item.resource_id}
                        </p>
                        <button
                          disabled={locked}
                          className={cn(
                            "block w-full rounded-sm text-left text-sm leading-5 hover:underline focus-visible:outline-2 focus-visible:outline-ring disabled:opacity-45",
                            item.read_at === null && "font-medium",
                          )}
                          onClick={() => void openNotification(item)}
                        >
                          {openingId === item.notification_id
                            ? "Đang kiểm tra quyền mở…"
                            : item.message}
                        </button>
                        <time
                          dateTime={item.created_at}
                          className="mt-1 block text-xs text-muted-foreground"
                        >
                          {new Date(item.created_at).toLocaleString("vi-VN")}
                        </time>
                        {item.read_at === null && (
                          <Button
                            className="mt-2"
                            size="sm"
                            variant="ghost"
                            disabled={locked}
                            onClick={() => {
                              setOpenError("")
                              void resource.markRead(item.notification_id).then(() => {
                                // Nút đọc biến mất sau GET mới; giữ điểm bàn phím trong khung còn mở.
                                panelTitle.current?.focus()
                              })
                            }}
                          >
                            <CheckIcon />{" "}
                            {state.readingId === item.notification_id
                              ? "Đang đánh dấu…"
                              : "Đánh dấu đã đọc"}
                          </Button>
                        )}
                      </div>
                    </li>
                  ))}
                </ol>
              ) : (
                <p className="px-5 py-10 text-center text-sm">
                  {state.data.total === 0
                    ? state.filter === "all"
                      ? "Chưa có thông báo"
                      : "Bạn đã đọc hết thông báo"
                    : "Trang này không có thông báo; hãy quay lại trang trước."}
                </p>
              )
            ) : null}
          </div>
          {state.data && state.phase !== "loading" && (
            <footer className="flex shrink-0 flex-wrap items-center justify-between gap-2 border-t border-border p-3 text-xs">
              <span>
                {state.data.total} thông báo · Trang {state.page}
              </span>
              <div className="flex gap-1">
                <Button
                  size="sm"
                  variant="outline"
                  disabled={locked || state.page === 1}
                  onClick={() => {
                    void resource
                      .load(state.filter, state.page - 1)
                      .then(() => panelTitle.current?.focus())
                  }}
                >
                  Trang trước
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  disabled={locked || state.page * 20 >= state.data.total}
                  onClick={() => {
                    void resource
                      .load(state.filter, state.page + 1)
                      .then(() => panelTitle.current?.focus())
                  }}
                >
                  Trang sau
                </Button>
              </div>
            </footer>
          )}
        </DialogPrimitive.Content>
      </DialogPrimitive.Root>
      {!isOpen && state.message && (
        <p role="status" className="max-w-sm text-xs text-muted-foreground">
          {state.message} Mở chuông và tải lại để kiểm tra.
        </p>
      )}
    </div>
  )
}
