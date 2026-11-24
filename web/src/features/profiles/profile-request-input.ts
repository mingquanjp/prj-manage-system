export type ProfileValues = { fullname: string; phone: string | null; address: string | null }
export const requestFields = ["fullname", "phone", "address"] as const
export type RequestField = (typeof requestFields)[number]
export type ProfileRequestInput = {
  expected_before_values: Partial<ProfileValues>
  proposed_values: Partial<ProfileValues>
  explanation: string | null
}
export class ProfileInputError extends Error {
  readonly field?: RequestField | "explanation"
  constructor(message: string, field?: RequestField | "explanation") {
    super(message)
    this.name = "ProfileInputError"
    this.field = field
  }
}
// strings.TrimSpace của Go dùng Unicode White_Space, khác String.trim ở U+0085/U+FEFF.
export function trimProfileText(value: string) {
  return value.replace(/^\p{White_Space}+|\p{White_Space}+$/gu, "")
}
export function validProfileText(value: unknown): value is string {
  return typeof value === "string" && !value.includes("\0") && !/[\uD800-\uDFFF]/u.test(value)
}
export function profileRequestBytes(input: ProfileRequestInput) {
  return new TextEncoder().encode(JSON.stringify(input)).byteLength
}
function validateValue(field: RequestField, value: unknown, proposed: boolean) {
  if (value === null && field !== "fullname" && (!proposed || field === "phone")) return
  const limit = { fullname: 100, phone: 20, address: 2000 }[field]
  if (
    !validProfileText(value) ||
    (field === "fullname" && !value.length) ||
    (proposed &&
      (trimProfileText(value) !== value || [...value].length < 1 || [...value].length > limit))
  )
    throw new ProfileInputError(
      `${{ fullname: "Họ tên", phone: "Số điện thoại", address: "Địa chỉ" }[field]} cần nội dung hợp lệ${proposed ? `, tối đa ${limit} ký tự Unicode` : ""}.`,
      field,
    )
}
export function validateProfileRequestInput(input: ProfileRequestInput) {
  if (
    !input ||
    typeof input !== "object" ||
    Object.keys(input).sort().join() !== "expected_before_values,explanation,proposed_values"
  )
    throw new ProfileInputError("Nội dung đề nghị không hợp lệ.")
  const before = input.expected_before_values
  const proposed = input.proposed_values
  if (
    !before ||
    !proposed ||
    typeof before !== "object" ||
    typeof proposed !== "object" ||
    Array.isArray(before) ||
    Array.isArray(proposed)
  )
    throw new ProfileInputError("Nội dung đề nghị không hợp lệ.")
  const keys = Object.keys(before)
  if (
    !keys.length ||
    keys.sort().join() !== Object.keys(proposed).sort().join() ||
    keys.some((key) => !requestFields.includes(key as RequestField))
  )
    throw new ProfileInputError("Chỉ gửi họ tên, số điện thoại hoặc địa chỉ thực sự thay đổi.")
  for (const field of keys as RequestField[]) {
    validateValue(field, before[field], false)
    validateValue(field, proposed[field], true)
    if (before[field] === proposed[field])
      throw new ProfileInputError("Bỏ trường chưa thay đổi khỏi đề nghị.", field)
  }
  if (
    input.explanation !== null &&
    (!validProfileText(input.explanation) ||
      trimProfileText(input.explanation) !== input.explanation ||
      [...input.explanation].length > 2000)
  )
    throw new ProfileInputError(
      "Ghi chú cần nội dung hợp lệ, tối đa 2000 ký tự Unicode.",
      "explanation",
    )
  if (profileRequestBytes(input) > 65536)
    throw new ProfileInputError("Nội dung gửi vượt giới hạn 64 KiB. Hãy giảm nội dung đề nghị.")
  return input
}
export function buildProfileRequest(
  before: ProfileValues,
  values: ProfileValues,
  explanation: string,
): ProfileRequestInput {
  const input: ProfileRequestInput = {
    expected_before_values: {},
    proposed_values: {},
    explanation: null,
  }
  for (const field of requestFields) {
    const raw = values[field]
    if (raw === before[field]) continue
    if (raw !== null && !validProfileText(raw)) validateValue(field, raw, true)
    const next = raw === null ? null : trimProfileText(raw)
    const normalized = field === "phone" && next === "" ? null : next
    if (normalized === before[field]) continue
    Object.assign(input.expected_before_values, { [field]: before[field] })
    Object.assign(input.proposed_values, { [field]: normalized })
  }
  if (!validProfileText(explanation))
    throw new ProfileInputError("Ghi chú chứa ký tự không hợp lệ.", "explanation")
  input.explanation = trimProfileText(explanation) || null
  return validateProfileRequestInput(input)
}
