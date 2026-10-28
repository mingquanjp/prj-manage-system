import { createContext, useContext, type ReactNode } from "react"
import type { IdentitySource } from "@/demo/identity"

const IdentityContext = createContext<IdentitySource | null>(null)
// Chỉ chia sẻ snapshot quyền demo hiện hành; không giữ bản sao state hoặc thay auth server.
export function IdentityProvider({
  value,
  children,
}: {
  value: IdentitySource
  children: ReactNode
}) {
  return <IdentityContext.Provider value={value}>{children}</IdentityContext.Provider>
}
export function useIdentity() {
  const identity = useContext(IdentityContext)
  if (!identity) throw new Error("Thiếu nguồn danh tính của workspace.")
  return identity
}
