// Read model trong bộ nhớ; không chứa tệp thật, storage key hoặc kết quả AI.
export type CvProcessingStatus = "QUEUED" | "PROCESSING" | "SUCCEEDED" | "FAILED"
export type CvFileInput = { name: string; size: number; type: string }
export type EmployeeCv = {
  id: string
  employeeId: string
  filename: string
  sizeBytes: number
  mimeType: string
  uploadedBy: string
  uploadedAt: string
  isCurrent: boolean
  processingStatus: CvProcessingStatus
}
export type EmployeeCvsState = { versions: EmployeeCv[] }
export const cvProcessingLabels: Record<CvProcessingStatus, string> = {
  QUEUED: "Chờ xử lý",
  PROCESSING: "Đang trích xuất",
  SUCCEEDED: "Trích xuất thành công",
  FAILED: "Trích xuất thất bại",
}
