export type PasswordFields = {
  currentPassword: string
  newPassword: string
  confirmation: string
}
export type PasswordErrors = Partial<Record<keyof PasswordFields, string>>

export function passwordSize(password: string) {
  return {
    codePoints: Array.from(password).length,
    bytes: new TextEncoder().encode(password).length,
  }
}

export function validatePasswordChange(fields: PasswordFields): PasswordErrors {
  const errors: PasswordErrors = {}
  const { currentPassword, newPassword, confirmation } = fields
  const size = passwordSize(newPassword)
  if (!currentPassword) errors.currentPassword = "Vui lòng nhập mật khẩu hiện tại."
  // TextEncoder thay surrogate lẻ bằng U+FFFD; từ chối trước để giữ đúng chuỗi gửi tới Go.
  if (
    Array.from(newPassword).some((char) => {
      const point = char.codePointAt(0)!
      return point >= 0xd800 && point <= 0xdfff
    })
  )
    errors.newPassword = "Mật khẩu chứa ký tự Unicode không hợp lệ."
  else if (size.codePoints < 15) errors.newPassword = "Mật khẩu mới cần ít nhất 15 ký tự Unicode."
  else if (size.bytes > 72) errors.newPassword = "Mật khẩu mới vượt quá 72 byte UTF-8."
  else if (currentPassword === newPassword)
    errors.newPassword = "Mật khẩu mới phải khác mật khẩu hiện tại."
  if (!confirmation) errors.confirmation = "Vui lòng nhập lại mật khẩu mới."
  else if (newPassword !== confirmation)
    errors.confirmation = "Hai lần nhập mật khẩu mới chưa khớp."
  return errors
}
