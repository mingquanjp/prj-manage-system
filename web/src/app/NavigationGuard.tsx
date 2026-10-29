import { createContext, useContext, useEffect, useRef, useState, type ReactNode } from "react"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"

const NavigationGuardContext = createContext<{
  registerDraft: (dirty: boolean) => void
  requestLeave: (action: () => void) => void
} | null>(null)

// Chỉ bảo vệ nháp phân công trên các đường điều hướng đã nối guard.
export function NavigationGuard({ children }: { children: ReactNode }) {
  const dirty = useRef(false)
  const [pending, setPending] = useState<(() => void) | null>(null)
  const [api] = useState(() => ({
    registerDraft: (value: boolean) => {
      dirty.current = value
    },
    requestLeave: (action: () => void) => {
      if (dirty.current) setPending(() => action)
      else action()
    },
  }))
  return (
    <NavigationGuardContext value={api}>
      {children}
      <Dialog
        open={Boolean(pending)}
        onOpenChange={(open) => {
          if (!open) setPending(null)
        }}
        alert
        title="Rời bản nháp phân công?"
        description="Các lựa chọn chưa xác nhận sẽ bị bỏ nếu bạn tiếp tục."
      >
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={() => setPending(null)}>
            Ở lại chỉnh sửa
          </Button>
          <Button
            onClick={() => {
              const action = pending
              setPending(null)
              action?.()
            }}
          >
            Bỏ nháp và tiếp tục
          </Button>
        </div>
      </Dialog>
    </NavigationGuardContext>
  )
}

export function useNavigationGuard() {
  const guard = useContext(NavigationGuardContext)
  if (!guard) throw new Error("Thiếu guard điều hướng")
  return guard
}

export function useAssignmentDraftGuard(dirty: boolean) {
  const { registerDraft } = useNavigationGuard()
  useEffect(() => {
    registerDraft(dirty)
    return () => registerDraft(false)
  }, [dirty, registerDraft])
}
