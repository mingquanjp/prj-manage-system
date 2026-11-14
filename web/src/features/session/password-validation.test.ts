import { test } from "node:test"
import assert from "node:assert/strict"
import { passwordSize, validatePasswordChange } from "./password-validation.ts"

const valid = "new password 12345"
function validate(newPassword: string, currentPassword = "12345678", confirmation = newPassword) {
  return validatePasswordChange({ currentPassword, newPassword, confirmation })
}
test("mật khẩu mới: biên code point và byte khớp quy tắc Go, không đếm UTF-16", () => {
  for (const [password, accepted] of [
    ["a".repeat(14), false],
    ["a".repeat(15), true],
    ["a".repeat(72), true],
    ["a".repeat(73), false],
    ["😀".repeat(14), false],
    ["😀".repeat(15), true],
    ["😀".repeat(18), true],
    ["😀".repeat(18) + "a", false],
    ["á".repeat(36), true],
    ["á".repeat(36) + "a", false],
    ["a".repeat(15) + "\ud800", false],
    ["\udfff" + "a".repeat(15), false],
  ] as const)
    assert.equal(!validate(password).newPassword, accepted)
  assert.deepEqual(passwordSize("😀áa"), { codePoints: 3, bytes: 7 })
})
test("giữ nguyên khoảng trắng/dấu/chuỗi tổ hợp; chỉ đối chiếu xác nhận ở FE", () => {
  assert.deepEqual(validate("   mật khẩu mới đủ dài   "), {})
  assert.deepEqual(validate("e\u0301".repeat(8)), {})
  assert.deepEqual(validate(" ".repeat(15)), {})
  assert.ok(validate(valid, valid).newPassword)
  assert.ok(validate(valid, "12345678", valid + " ").confirmation)
  assert.ok(validate(valid, "").currentPassword)
  assert.ok(validate(valid, "12345678", "").confirmation)
})
