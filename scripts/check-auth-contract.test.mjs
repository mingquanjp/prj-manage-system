// Kiểm tra tĩnh contract đã chốt; không thay validator OpenAPI hoặc test auth runtime.
import assert from "node:assert/strict"
import { readFile } from "node:fs/promises"
import { test } from "node:test"

const spec = JSON.parse(await readFile(new URL("../docs/api/auth.openapi.json", import.meta.url), "utf8"))
const operations = Object.entries(spec.paths).flatMap(([path, item]) =>
  Object.entries(item).map(([method, operation]) => ({ path, method, operation })))

function resolve(pointer) {
  assert.ok(pointer.startsWith("#/"), `Không cho ref ngoài chưa được review: ${pointer}`)
  return pointer.slice(2).split("/").reduce((value, key) => {
    const decoded = key.replaceAll("~1", "/").replaceAll("~0", "~")
    assert.ok(value && Object.hasOwn(value, decoded), `Ref không tồn tại: ${pointer}`)
    return value[decoded]
  }, spec)
}

function walk(value, visit) {
  if (!value || typeof value !== "object") return
  visit(value)
  for (const child of Object.values(value)) walk(child, visit)
}

function op(path) {
  return operations.find((entry) => entry.path === `/api/auth/${path}`).operation
}

test("contract chỉ có năm route auth, ID duy nhất và ref nội bộ tồn tại", () => {
  assert.equal(spec.openapi, "3.1.1")
  assert.deepEqual(operations.map(({ method, path }) => `${method} ${path}`).sort(), [
    "get /api/auth/csrf", "get /api/auth/me", "post /api/auth/change-password",
    "post /api/auth/login", "post /api/auth/logout",
  ])
  assert.equal(new Set(operations.map(({ operation }) => operation.operationId)).size, 5)
  walk(spec, (value) => { if (value.$ref) resolve(value.$ref) })
})

test("chỉ login công khai, mutation cần đồng thời cookie + CSRF và Origin", () => {
  assert.deepEqual(op("login").security, [])
  for (const path of ["me", "csrf"]) assert.deepEqual(op(path).security, [{ cookieAuth: [] }])
  for (const path of ["logout", "change-password"]) {
    // Một object là AND; hai object trong array sẽ thành OR, làm mất bảo vệ.
    assert.deepEqual(op(path).security, [{ cookieAuth: [], csrfHeader: [] }])
  }
  for (const path of ["login", "logout", "change-password"]) {
    assert.ok(op(path).parameters.some((p) => p.in === "header" && p.name === "Origin" && p.required))
  }
  assert.equal(spec.components.securitySchemes.cookieAuth.in, "cookie")
  assert.equal(spec.components.securitySchemes.cookieAuth.name, "staffing_session")
  assert.equal(spec.components.securitySchemes.csrfHeader.name, "X-CSRF-Token")
})

test("request allowlist không nhận actor/role/cờ để bỏ buộc đổi", () => {
  for (const [name, keys] of [
    ["LoginRequest", ["email", "password"]],
    ["ChangeInitialPasswordRequest", ["current_password", "new_password"]],
  ]) {
    const schema = spec.components.schemas[name]
    assert.equal(schema.additionalProperties, false)
    assert.deepEqual(Object.keys(schema.properties).sort(), [...keys].sort())
    assert.deepEqual([...schema.required].sort(), [...keys].sort())
  }
  assert.equal(spec.components.schemas.LoginRequest.properties.password.writeOnly, true)
  const schema = spec.components.schemas.ChangeInitialPasswordRequest
  assert.equal(schema.properties.current_password.writeOnly, true)
  assert.equal(schema.properties.new_password.writeOnly, true)
  assert.equal(schema.properties.new_password.minLength, 15)
  assert.equal(schema.properties.new_password["x-max-utf8-bytes"], 72)
  // Chính sách đặt mật khẩu mới không được áp ngược lên credential cũ.
  assert.equal(spec.components.schemas.LoginRequest.properties.password.minLength, 1)
})

test("DTO danh tính giới hạn và phiên buộc đổi không có role hiệu lực", () => {
  const keys = ["user_id", "employee_id", "employee_code", "fullname", "system_roles", "must_change_password"]
  for (const name of ["LimitedPrincipal", "BusinessPrincipal"]) {
    const schema = spec.components.schemas[name]
    assert.equal(schema.additionalProperties, false)
    assert.deepEqual(Object.keys(schema.properties).sort(), [...keys].sort())
    assert.deepEqual([...schema.required].sort(), [...keys].sort())
  }
  const limited = spec.components.schemas.LimitedPrincipal.properties
  assert.equal(limited.must_change_password.const, true)
  assert.equal(limited.system_roles.maxItems, 0)
  const full = spec.components.schemas.BusinessPrincipal.properties
  assert.equal(full.must_change_password.const, false)
  assert.deepEqual(full.system_roles.items.enum, ["ADMIN", "HR", "DIRECTOR"])
  assert.equal(full.system_roles.uniqueItems, true)
})

test("mọi response auth no-store, chỉ login thành công cấp cookie", () => {
  for (const { operation } of operations) {
    for (const [status, response] of Object.entries(operation.responses)) {
      assert.equal(response.headers["Cache-Control"].schema.const, "no-store")
      if (+status >= 400) assert.ok(!response.headers["Set-Cookie"])
    }
  }
  assert.ok(op("login").responses["200"].headers["Set-Cookie"])
  for (const path of ["me", "csrf"]) assert.ok(!op(path).responses["200"].headers["Set-Cookie"])
  for (const path of ["logout", "change-password"]) {
    assert.ok(op(path).responses["204"].headers["Set-Cookie"])
    assert.ok(!op(path).responses["204"].content)
  }
})

test("lỗi phân biệt sai credential, phiên, policy và DB, không entity thô", () => {
  for (const { operation } of operations) {
    for (const [status, response] of Object.entries(operation.responses)) {
      if (+status < 400) continue
      const schema = response.content["application/json"].schema
      assert.equal(schema.additionalProperties, false)
      assert.deepEqual(Object.keys(schema.properties).sort(), ["code", "message", "request_id"])
      assert.ok(schema.properties.code.enum.length)
    }
    for (const code of ["500", "503"]) assert.ok(operation.responses[code])
  }
  assert.deepEqual(op("login").responses["401"].content["application/json"].schema.properties.code.enum, ["INVALID_CREDENTIALS"])
  assert.deepEqual(op("me").responses["401"].content["application/json"].schema.properties.code.enum, ["UNAUTHENTICATED"])
  assert.ok(op("change-password").responses["400"].content["application/json"].schema.properties.code.enum.includes("INVALID_CURRENT_PASSWORD"))
  assert.ok(op("change-password").responses["403"].content["application/json"].schema.properties.code.enum.includes("PASSWORD_CHANGE_NOT_REQUIRED"))
  for (const path of ["login", "change-password"]) assert.ok(op(path).responses["429"].headers["Retry-After"])
})

test("cổng phiên hạn chế ghi rõ allowlist, không tạo API nghiệp vụ giả", () => {
  const guard = spec["x-business-guard"]
  assert.equal(guard.required, true)
  assert.deepEqual(guard.limited_session_allowlist, ["getCurrentUser", "getCsrfToken", "logout", "changeInitialPassword"])
  assert.deepEqual(guard.denied_response, { status: 403, code: "PASSWORD_CHANGE_REQUIRED" })
})


test("body JSON login/đổi lần đầu có giới hạn 8192 byte và phản hồi 413", () => {
  for (const path of ["login", "change-password"]) {
    assert.equal(op(path).requestBody["x-max-body-bytes"], 8192)
    assert.deepEqual(op(path).responses["413"].content["application/json"].schema.properties.code.enum, ["PAYLOAD_TOO_LARGE"])
  }
  // Giới hạn byte body không được thay giới hạn byte của trường mật khẩu.
  assert.equal(spec.components.schemas.ChangeInitialPasswordRequest.properties.new_password["x-max-utf8-bytes"], 72)
})

test("ma trận status/code đủ cho từng route, không trộn lỗi login với lỗi phiên", () => {
  const common = { 500: ["INTERNAL_ERROR"], 503: ["SERVICE_UNAVAILABLE"] }
  const protectedErrors = { ...common, 401: ["UNAUTHENTICATED"] }
  const jsonErrors = { 413: ["PAYLOAD_TOO_LARGE"], 415: ["UNSUPPORTED_MEDIA_TYPE"] }
  const expected = {
    login: { ...common, ...jsonErrors, 400: ["VALIDATION_ERROR"], 401: ["INVALID_CREDENTIALS"], 403: ["FORBIDDEN"], 429: ["RATE_LIMITED"] },
    me: protectedErrors,
    csrf: protectedErrors,
    logout: { ...protectedErrors, 400: ["VALIDATION_ERROR"], 403: ["FORBIDDEN"] },
    "change-password": { ...protectedErrors, ...jsonErrors, 400: ["VALIDATION_ERROR", "INVALID_CURRENT_PASSWORD"], 403: ["FORBIDDEN", "PASSWORD_CHANGE_NOT_REQUIRED"], 429: ["RATE_LIMITED"] },
  }
  for (const [path, errors] of Object.entries(expected)) {
    const actual = Object.fromEntries(Object.entries(op(path).responses)
      .filter(([status]) => +status >= 400)
      .map(([status, response]) => [status, response.content["application/json"].schema.properties.code.enum]))
    assert.deepEqual(actual, errors, path)
    const success = Object.keys(op(path).responses).filter((status) => +status < 400)
    assert.deepEqual(success, [path === "logout" || path === "change-password" ? "204" : "200"])
  }
})

test("ví dụ login/me thống nhất DTO, cờ mới phân biệt quyền rỗng với phiên hạn chế", () => {
  const examples = op("login").responses["200"].content["application/json"].examples
  assert.deepEqual(op("me").responses["200"].content["application/json"].examples, examples)
  assert.equal(examples.employee.value.must_change_password, false)
  assert.equal(examples.limited.value.must_change_password, true)
  for (const { value } of Object.values(examples)) {
    const name = value.must_change_password ? "LimitedPrincipal" : "BusinessPrincipal"
    assert.deepEqual(Object.keys(value).sort(), [...spec.components.schemas[name].required].sort())
    assert.deepEqual(value.system_roles, [])
    assert.ok(Number.isInteger(value.user_id) && value.user_id > 0)
    assert.ok(Number.isInteger(value.employee_id) && value.employee_id > 0)
    assert.notEqual(value.user_id, value.employee_id)
  }
  const error = op("login").responses["401"].content["application/json"]
  assert.ok(error.schema.properties.code.enum.includes(error.example.code))
  assert.deepEqual(Object.keys(error.example).sort(), [...error.schema.required].sort())
})

test("JSON chỉ bắt buộc ở login/đổi lần đầu; logout không có requestBody", () => {
  for (const path of ["login", "change-password"]) {
    const body = op(path).requestBody
    assert.equal(body.required, true)
    assert.deepEqual(Object.keys(body.content), ["application/json"])
  }
  for (const path of ["me", "csrf", "logout"]) assert.equal(op(path).requestBody, undefined)
  // Không biến yêu cầu Content-Type của JSON thành điều kiện logout.
  assert.equal(op("logout").responses["415"], undefined)
  const csrf = spec.components.schemas.CsrfResponse
  assert.equal(csrf.additionalProperties, false)
  assert.deepEqual(csrf.required, ["csrf_token"])
  assert.deepEqual(Object.keys(csrf.properties), ["csrf_token"])
})
