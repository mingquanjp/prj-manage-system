import { test } from "node:test"
import assert from "node:assert/strict"
import {
  buildProfileRequest,
  profileRequestBytes,
  trimProfileText,
  validateProfileRequestInput,
} from "./profile-request-input.ts"
const before = { fullname: "Người thử", phone: null, address: null }
test("patch chỉ chứa trường đổi; giữ raw before và NULL; không tự chuẩn hóa Unicode", () => {
  assert.deepEqual(
    buildProfileRequest(before, { ...before, address: "  Địa chỉ mới  " }, "  ghi chú  "),
    {
      expected_before_values: { address: null },
      proposed_values: { address: "Địa chỉ mới" },
      explanation: "ghi chú",
    },
  )
  assert.deepEqual(
    buildProfileRequest({ ...before, phone: "123" }, { ...before, phone: " " }, "").proposed_values,
    { phone: null },
  )
  const decomposed = "a\u0301"
  assert.equal(
    buildProfileRequest(before, { ...before, fullname: decomposed }, "").proposed_values.fullname,
    decomposed,
  )
  assert.equal(trimProfileText("\u0085a\u0085"), "a")
  assert.equal(trimProfileText("\uFEFFa\uFEFF"), "\uFEFFa\uFEFF")
})
test("Unicode code point thay UTF16; mỗi giới hạn và ký tự lỗi", () => {
  for (const [field, limit] of [
    ["fullname", 100],
    ["phone", 20],
    ["address", 2000],
  ] as const) {
    assert.doesNotThrow(() =>
      buildProfileRequest(before, { ...before, [field]: "😀".repeat(limit) }, ""),
    )
    assert.throws(() =>
      buildProfileRequest(before, { ...before, [field]: "😀".repeat(limit + 1) }, ""),
    )
    for (const value of ["\ud800", "\udc00", "a\0b"])
      assert.throws(() => buildProfileRequest(before, { ...before, [field]: value }, ""))
  }
  assert.doesNotThrow(() =>
    buildProfileRequest(before, { ...before, phone: "123" }, "😀".repeat(2000)),
  )
  assert.throws(() => buildProfileRequest(before, { ...before, phone: "123" }, "😀".repeat(2001)))
  assert.throws(() => buildProfileRequest(before, { ...before, phone: "123" }, "\ud800"))
})
test("BR001 không xóa địa chỉ, bỏ trường không đổi, không nhận payload ngoài allowlist", () => {
  assert.throws(() => buildProfileRequest(before, before, "chỉ ghi chú"))
  for (const address of ["", "  ", null])
    assert.throws(() =>
      buildProfileRequest({ ...before, address: "Cũ" }, { ...before, address }, ""),
    )
  assert.deepEqual(
    buildProfileRequest(before, { ...before, fullname: "Tên mới" }, "").expected_before_values,
    { fullname: "Người thử" },
  )
  assert.throws(() =>
    validateProfileRequestInput({
      expected_before_values: { address: null },
      proposed_values: { address: "Mới" },
      explanation: null,
      actor_id: 99,
    } as never),
  )
})
test("byte JSON chính xác 65536/65537 kể cả raw legacy dài và escape", () => {
  const input = buildProfileRequest({ ...before, address: "x" }, { ...before, address: "Mới" }, "")
  const baseBytes = profileRequestBytes(input)
  input.expected_before_values.address = "x".repeat(1 + 65536 - baseBytes)
  assert.equal(profileRequestBytes(input), 65536)
  assert.doesNotThrow(() => validateProfileRequestInput(input))
  input.expected_before_values.address += "x"
  assert.equal(profileRequestBytes(input), 65537)
  assert.throws(() => validateProfileRequestInput(input), /64 KiB/)
  assert.throws(
    () =>
      buildProfileRequest(
        { ...before, address: "\n".repeat(33000) },
        { ...before, address: "Mới" },
        "",
      ),
    /64 KiB/,
  )
})
