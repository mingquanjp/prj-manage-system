// Danh mục trong phiên frontend, không phải DTO/schema server.
export type OrganizationEntry = { id: string; name: string }
export type OrganizationKind = "departments" | "positions" | "jobLevels"
export type OrganizationState = Record<OrganizationKind, OrganizationEntry[]>
export type OrganizationCommand =
  | { action: "create"; kind: OrganizationKind; name: string }
  | { action: "update"; kind: OrganizationKind; id: string; beforeName: string; name: string }
  | { action: "delete"; kind: OrganizationKind; id: string; beforeName: string }

export const organizationLabels: Record<OrganizationKind, string> = {
  departments: "Phòng ban",
  positions: "Vị trí công việc",
  jobLevels: "Cấp bậc",
}
