import { useState } from "react"
import * as DialogPrimitive from "@radix-ui/react-dialog"
import { useNavigate, useSearchParams, Link } from "react-router"
import { BellIcon, Cross2Icon, CheckIcon } from "@radix-ui/react-icons"
import { useIdentity } from "@/components/identity"
import { Button } from "@/components/ui/button"
import { EmptyState } from "@/components/shared"
import { cn } from "@/lib/utils"
import {
  notificationsFor,
  notificationDestination,
  notificationContext,
} from "@/demo/notifications"
import { canReadRequest, contributionFor } from "@/demo/corrections"
import type { DemoWorkspaceState } from "@/demo/workspace-state"
import { CorrectionRequests } from "@/features/work-logs/CorrectionRequests"
import type { WorkState } from "@/demo/effort"

export function NotificationMenu({
  state,
  actorId,
  onRead,
}: {
  state: DemoWorkspaceState
  actorId: string
  onRead: (id: string) => void
}) {
  const [params, setParams] = useSearchParams()
  const [isOpen, setIsOpen] = useState(false)
  const openFromLink = params.get("notifications") === "open"
  const changeOpen = (value: boolean) => {
    setIsOpen(value)
    if (openFromLink) {
      const next = new URLSearchParams(params)
      next.delete("notifications")
      setParams(next, { replace: true })
    }
  }
  const identity = useIdentity()
  const navigate = useNavigate()
  const [filter, setFilter] = useState("all")
  const [error, setError] = useState("")
  const all = notificationsFor(state, actorId, identity)
  const unread = all.filter((n) => !n.readAt).length
  const entries = all.filter((n) => filter === "all" || !n.readAt)
  const open = (id: string) => {
    try {
      const destination = notificationDestination(state, actorId, id, identity)
      onRead(id)
      setError("")
      setIsOpen(false)
      navigate(destination)
    } catch (err) {
      setError((err as Error).message)
    }
  }
  return (
    <DialogPrimitive.Root modal={false} open={isOpen || openFromLink} onOpenChange={changeOpen}>
      <div className="relative">
        <DialogPrimitive.Trigger asChild>
          <Button size="sm" variant="ghost" aria-label={`Thông báo, ${unread} chưa đọc`}>
            <BellIcon />
            <span className="hidden sm:inline">Thông báo</span>
            {unread > 0 && (
              <span className="rounded-full bg-primary px-1.5 text-xs text-white">{unread}</span>
            )}
          </Button>
        </DialogPrimitive.Trigger>
        <DialogPrimitive.Content
          aria-describedby={undefined}
          className="fixed inset-x-4 top-16 z-40 flex max-h-[min(560px,calc(100dvh-80px))] w-auto max-w-[calc(100vw-2rem)] flex-col overflow-hidden rounded-xl border border-border bg-white shadow-xl outline-none sm:absolute sm:left-auto sm:right-0 sm:top-full sm:mt-2 sm:w-[420px]"
        >
          <header className="flex shrink-0 items-center justify-between px-4 pb-2 pt-3">
            <DialogPrimitive.Title className="text-base font-semibold">
              Thông báo
            </DialogPrimitive.Title>
            <DialogPrimitive.Close asChild>
              <Button variant="ghost" size="icon" aria-label="Đóng thông báo">
                <Cross2Icon />
              </Button>
            </DialogPrimitive.Close>
          </header>
          <div
            className="flex shrink-0 flex-wrap gap-1 border-b border-border px-3 pb-3"
            role="group"
            aria-label="Lọc thông báo"
          >
            <button
              className={cn("filter-chip", filter === "all" && "filter-active")}
              aria-pressed={filter === "all"}
              onClick={() => setFilter("all")}
            >
              Tất cả ({all.length})
            </button>
            <button
              className={cn("filter-chip", filter === "unread" && "filter-active")}
              aria-pressed={filter === "unread"}
              onClick={() => setFilter("unread")}
            >
              Chưa đọc ({unread})
            </button>
          </div>
          {error && (
            <p
              role="alert"
              className="mb-4 rounded-lg border border-destructive/20 p-3 text-sm text-destructive"
            >
              {error}
            </p>
          )}
          <div className="min-h-0 overflow-y-auto overscroll-contain">
            {!entries.length ? (
              <div className="px-5 py-10 text-center">
                <BellIcon className="mx-auto mb-3 size-5 text-muted-foreground" />
                <p className="text-sm font-medium">
                  {filter === "all" ? "Chưa có thông báo mới" : "Bạn đã đọc hết thông báo"}
                </p>
              </div>
            ) : (
              <ol className="divide-y divide-border">
                {entries.map((n) => (
                  <li
                    key={n.id}
                    className={cn(
                      "flex flex-wrap items-start gap-3 p-4 sm:flex-nowrap",
                      !n.readAt && "bg-accent/40",
                    )}
                  >
                    <span
                      className={cn(
                        "mt-2 size-2 shrink-0 rounded-full",
                        n.readAt ? "bg-border" : "bg-primary",
                      )}
                      aria-label={n.readAt ? "Đã đọc" : "Chưa đọc"}
                    />
                    <div className="min-w-0 flex-1">
                      <p className="mb-1 text-xs font-medium text-muted-foreground">
                        {notificationContext(state, actorId, n.id, identity)}
                      </p>
                      <button
                        className={cn(
                          "block w-full rounded-sm text-left text-sm leading-5 hover:underline focus-visible:outline-2 focus-visible:outline-ring",
                          !n.readAt && "font-medium",
                        )}
                        onClick={() => open(n.id)}
                      >
                        {n.message}
                      </button>
                      <time className="mt-1 block text-xs text-muted-foreground" dateTime={n.at}>
                        {new Date(n.at).toLocaleString("vi-VN")}
                      </time>
                      <div className="mt-2 flex flex-wrap gap-2">
                        {!n.readAt && (
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() => {
                              try {
                                onRead(n.id)
                                setError("")
                              } catch (err) {
                                setError((err as Error).message)
                              }
                            }}
                          >
                            <CheckIcon /> Đánh dấu đã đọc
                          </Button>
                        )}
                      </div>
                    </div>
                  </li>
                ))}
              </ol>
            )}
          </div>
        </DialogPrimitive.Content>
      </div>
    </DialogPrimitive.Root>
  )
}
export function NotificationCorrection({
  state,
  actorId,
  onChange,
}: {
  state: DemoWorkspaceState
  actorId: string
  onChange: (work: WorkState) => void
}) {
  const [params] = useSearchParams()
  const identity = useIdentity()
  const request = state.work.requests.find((r) => r.id === params.get("request"))
  const contribution = request
    ? contributionFor(state.work, state.tasks, request.assignmentId)
    : undefined
  const reviewProject = state.projects.find(
    (project) => project.id === contribution?.projectId && project.ptcmId === actorId,
  )
  const readable =
    request && canReadRequest(state.projects, state.tasks, state.work, request, actorId, identity)
  return (
    <div className="mx-auto max-w-4xl">
      <Link to="/?notifications=open" className="text-sm underline">
        ← Thông báo
      </Link>
      {readable ? (
        <CorrectionRequests
          key={request.id}
          projects={state.projects}
          tasks={state.tasks}
          work={state.work}
          actorId={actorId}
          requestId={request.id}
          projectId={reviewProject?.id}
          onChange={onChange}
        />
      ) : (
        <EmptyState
          title="Không thể mở yêu cầu hiệu chỉnh"
          detail="Yêu cầu không tồn tại hoặc bạn không còn quyền đọc."
        />
      )}
    </div>
  )
}
