import { defaultIdentity, type IdentitySource } from "./identity.ts"
import type { CvFileInput, CvProcessingStatus, EmployeeCvsState } from "../models/employee-cvs.ts"
import type { EmployeeProfile } from "../models/profile.ts"
import { getPerson, isHr } from "./identity.ts"

export const maxCvBytes = 10 * 1024 * 1024
const pdfMime = "application/pdf"
const docxMime = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"

export function createEmployeeCvsState(): EmployeeCvsState {
  return {
    versions: [
      {
        id: "cv-dung-previous",
        employeeId: "dung",
        filename: "CV_Pham_Quang_Dung_cu.docx",
        sizeBytes: 36864,
        mimeType: docxMime,
        uploadedBy: "hr",
        uploadedAt: "2026-10-20T03:00:00.000Z",
        isCurrent: false,
        processingStatus: "FAILED",
      },
      {
        id: "cv-dung-current",
        employeeId: "dung",
        filename: "CV_Pham_Quang_Dung.pdf",
        sizeBytes: 122880,
        mimeType: pdfMime,
        uploadedBy: "dung",
        uploadedAt: "2026-10-21T03:00:00.000Z",
        isCurrent: true,
        processingStatus: "SUCCEEDED",
      },
    ],
  }
}

function requireActive(
  profiles: EmployeeProfile[],
  personId: string,
  identity: IdentitySource = defaultIdentity,
  isActor = true,
) {
  // Khóa tài khoản chặn người thao tác; không tự vô hiệu hóa hồ sơ HR đang quản lý.
  if (
    (isActor && !getPerson(personId, identity)?.active) ||
    profiles.find((p) => p.personId === personId)?.employmentStatus !== "ACTIVE"
  )
    throw new Error("Hồ sơ không còn hoạt động; không thể cập nhật CV.")
}

export function canManageCvs(
  actorId: string,
  employeeId: string,
  identity: IdentitySource = defaultIdentity,
) {
  return actorId === employeeId || isHr(actorId, identity)
}

// Kiểm tra sơ bộ ở trình duyệt, không thay xác minh nội dung/PDF text/container trên server.
export function validateCvFile(file: CvFileInput) {
  const name = file.name.trim()
  const expectedMime = /\.pdf$/i.test(name) ? pdfMime : /\.docx$/i.test(name) ? docxMime : null
  if (!expectedMime || !name || name.length > 255)
    throw new Error("Chọn tệp PDF hoặc DOCX với tên tối đa 255 ký tự.")
  if (!Number.isSafeInteger(file.size) || file.size <= 0 || file.size > maxCvBytes)
    throw new Error("Tệp phải có nội dung và dung lượng không quá 10 MiB.")
  if (file.type && file.type !== expectedMime)
    throw new Error("Loại tệp không khớp PDF/DOCX. Hãy kiểm tra lại tệp đã chọn.")
  return { filename: name, sizeBytes: file.size, mimeType: expectedMime }
}

export function addEmployeeCv(
  state: EmployeeCvsState,
  profiles: EmployeeProfile[],
  actorId: string,
  employeeId: string,
  file: CvFileInput,
  id: string,
  at: string,
  identity: IdentitySource = defaultIdentity,
): EmployeeCvsState {
  requireActive(profiles, actorId, identity)
  requireActive(profiles, employeeId, identity, false)
  if (!canManageCvs(actorId, employeeId, identity))
    throw new Error("Chỉ chính chủ hoặc HR được bổ sung CV.")
  const metadata = validateCvFile(file)
  return {
    versions: [
      ...state.versions,
      {
        ...metadata,
        id,
        employeeId,
        uploadedBy: actorId,
        uploadedAt: at,
        isCurrent: false,
        processingStatus: "QUEUED",
      },
    ],
  }
}

export function selectCurrentCv(
  state: EmployeeCvsState,
  profiles: EmployeeProfile[],
  actorId: string,
  cvId: string,
  identity: IdentitySource = defaultIdentity,
): EmployeeCvsState {
  requireActive(profiles, actorId, identity)
  const cv = state.versions.find((entry) => entry.id === cvId)
  if (!cv || cv.employeeId !== actorId)
    throw new Error("Chỉ chính chủ được xác nhận CV hiện tại của mình.")
  return {
    versions: state.versions.map((entry) =>
      entry.employeeId === actorId ? { ...entry, isCurrent: entry.id === cvId } : entry,
    ),
  }
}

export function simulateCvProcessing(
  state: EmployeeCvsState,
  profiles: EmployeeProfile[],
  actorId: string,
  cvId: string,
  status: CvProcessingStatus,
  identity: IdentitySource = defaultIdentity,
): EmployeeCvsState {
  requireActive(profiles, actorId, identity)
  const cv = state.versions.find((entry) => entry.id === cvId)
  if (!cv || !canManageCvs(actorId, cv.employeeId, identity))
    throw new Error("Không có quyền mô phỏng xử lý CV này.")
  requireActive(profiles, cv.employeeId, identity, false)
  if (!["QUEUED", "PROCESSING", "SUCCEEDED", "FAILED"].includes(status))
    throw new Error("Trạng thái mô phỏng không hợp lệ.")
  return {
    versions: state.versions.map((entry) =>
      entry.id === cvId ? { ...entry, processingStatus: status } : entry,
    ),
  }
}
