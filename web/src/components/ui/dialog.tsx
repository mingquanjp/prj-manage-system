import * as DialogPrimitive from "@radix-ui/react-dialog"
import { Cross2Icon } from "@radix-ui/react-icons"
import type { ReactNode } from "react"
import { cn } from "@/lib/utils"
import { Button } from "./button"

export function Dialog({
  open,
  onOpenChange,
  title,
  description,
  children,
  side = false,
  wide = false,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  description: string
  children: ReactNode
  side?: boolean
  wide?: boolean
}) {
  return (
    <DialogPrimitive.Root open={open} onOpenChange={onOpenChange}>
      <DialogPrimitive.Portal>
        <DialogPrimitive.Overlay className="fixed inset-0 z-30 bg-foreground/20" />
        <DialogPrimitive.Content
          className={cn(
            "fixed z-40 overflow-auto border border-border bg-white shadow-xl outline-none",
            side
              ? "inset-y-0 right-0 w-full max-w-xl border-y-0 border-r-0 p-7"
              : cn(
                  "left-1/2 top-1/2 max-h-[90dvh] w-[calc(100%-2rem)] -translate-x-1/2 -translate-y-1/2 rounded-xl p-5 sm:p-7",
                  wide ? "max-w-3xl" : "max-w-xl",
                ),
          )}
        >
          <div className="mb-6 flex items-start justify-between gap-4">
            <div>
              <DialogPrimitive.Title className="text-xl font-semibold text-balance">
                {title}
              </DialogPrimitive.Title>
              <DialogPrimitive.Description className="mt-2 text-sm leading-6 text-muted-foreground">
                {description}
              </DialogPrimitive.Description>
            </div>
            <DialogPrimitive.Close asChild>
              <Button variant="ghost" size="icon" aria-label="Đóng">
                <Cross2Icon />
              </Button>
            </DialogPrimitive.Close>
          </div>
          {children}
        </DialogPrimitive.Content>
      </DialogPrimitive.Portal>
    </DialogPrimitive.Root>
  )
}
