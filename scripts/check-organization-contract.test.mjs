import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { test } from 'node:test'

const spec = JSON.parse(await readFile(new URL('../docs/api/organization-accounts-profile.openapi.json', import.meta.url), 'utf8'))
const auth = JSON.parse(await readFile(new URL('../docs/api/auth.openapi.json', import.meta.url), 'utf8'))
const methods = new Set(['get', 'post', 'patch', 'delete'])
const ops = Object.entries(spec.paths).flatMap(([path, item]) => Object.entries(item)
  .filter(([method]) => methods.has(method)).map(([method, operation]) => ({ path, method, operation })))
function resolve(value) {
  if (!value.$ref) return value
  assert.ok(value.$ref.startsWith('#/'), 'Ref phải local')
  return value.$ref.slice(2).split('/').reduce((node, part) => {
    const key = part.replaceAll('~1', '/').replaceAll('~0', '~')
    assert.ok(node && Object.hasOwn(node, key), `Ref thiếu: ${value.$ref}`)
    return node[key]
  }, spec)
}
function walk(value) {
  if (!value || typeof value !== 'object') return
  if (value.$ref) resolve(value)
  for (const child of Object.values(value)) walk(child)
}
const schema = (name) => spec.components.schemas[name]
const operation = (path, method) => spec.paths[path][method]

test('phạm vi 16 thao tác: catalog CRUD, preview/create/list/self; không thêm auth/reset/gửi lại', () => {
  const expected = []
  for (const type of ['departments', 'positions', 'job-levels']) {
    expected.push(`get /api/${type}`, `post /api/${type}`, `patch /api/${type}/{id}`, `delete /api/${type}/{id}`)
  }
  expected.push('get /api/employees/next-code', 'post /api/employees', 'get /api/users', 'get /api/employees/{id}')
  assert.deepEqual(ops.map(({ path, method }) => `${method} ${path}`).sort(), expected.sort())
  assert.equal(new Set(ops.map(({ operation }) => operation.operationId)).size, 16)
  walk(spec)
})

test('mọi route cần phiên thường, mutation AND cookie/CSRF cùng Origin và body thực 8 KiB', () => {
  assert.deepEqual(spec.components.securitySchemes, auth.components.securitySchemes)
  for (const { method, operation: op } of ops) {
    assert.equal(op['x-requires-business-session'], true)
    assert.equal(op['x-reject-unknown-query'], true)
    assert.deepEqual(op.security, [{ cookieAuth: [], ...(method !== 'get' ? { csrfHeader: [] } : {}) }])
    if (method !== 'get') {
      assert.equal(op['x-max-body-bytes'], 8192)
      assert.ok(op.parameters.map(resolve).some((p) => p.in === 'header' && p.name === 'Origin' && p.required))
      assert.equal(op.requestBody.required, true)
    } else assert.equal(op.requestBody, undefined)
    for (const response of Object.values(op.responses).map(resolve)) {
      assert.equal(response.headers['Cache-Control'].schema.const, 'no-store')
      assert.equal(response.headers['Set-Cookie'], undefined)
    }
  }
})

test('quyền HR mutation, Admin create/list/preview và self không có bypass Admin', () => {
  for (const type of ['departments', 'positions', 'job-levels']) {
    assert.deepEqual(operation(`/api/${type}`, 'get')['x-required-system-roles'], ['HR', 'ADMIN'])
    assert.deepEqual(operation(`/api/${type}`, 'post')['x-required-system-roles'], ['HR'])
    for (const method of ['patch', 'delete']) assert.deepEqual(operation(`/api/${type}/{id}`, method)['x-required-system-roles'], ['HR'])
  }
  for (const [path, method] of [['/api/employees', 'post'], ['/api/users', 'get'], ['/api/employees/next-code', 'get']]) {
    assert.deepEqual(operation(path, method)['x-required-system-roles'], ['ADMIN'])
  }
  const self = operation('/api/employees/{id}', 'get')
  assert.deepEqual(self['x-required-system-roles'], [])
  assert.ok(self.responses['404'])
})

test('DTO không có username/password/token, position bắt buộc và không thu hẹp enum legacy', () => {
  for (const name of ['Account', 'SelfProfile', 'CreateEmployeeRequest', 'CreateEmployeeResult', 'MailResult']) {
    assert.equal(schema(name).additionalProperties, false)
    for (const key of Object.keys(schema(name).properties)) assert.ok(!/password|token|secret|username|actor_id/.test(key) || key === 'must_change_password', `${name}.${key}`)
  }
  assert.deepEqual(schema('SelfProfile').properties.position, { $ref: '#/components/schemas/CatalogItem' })
  for (const key of ['department', 'job_level']) assert.ok(schema('SelfProfile').properties[key].oneOf.some((v) => v.type === 'null'))
  assert.deepEqual(schema('Account').properties.status.enum, ['ACTIVE', 'LOCKED', 'DISABLED'])
  assert.deepEqual(schema('SelfProfile').properties.employment_status.enum, ['ACTIVE', 'INACTIVE', 'RESIGNED'])
  for (const key of ['address', 'phone']) assert.equal(Object.hasOwn(schema('Account').properties, key), false)
  assert.deepEqual(schema('CreateEmployeeRequest').required, ['fullname', 'company_email', 'position_id'])
  assert.equal(schema('CreateEmployeeRequest').properties.system_roles.uniqueItems, true)
})

test('201 tách account/mail, accepted không là delivered; 204 không JSON và rename/delete có dữ liệu nền', () => {
  assert.equal(schema('CreateEmployeeResult').properties.account_created.const, true)
  assert.deepEqual(schema('MailResult').properties.status.enum, ['accepted', 'failed', 'unknown'])
  const create = operation('/api/employees', 'post')
  const examples = create.responses['201'].content['application/json'].examples
  assert.deepEqual(Object.keys(examples).sort(), ['accepted', 'failed', 'unknown'])
  for (const [status, ex] of Object.entries(examples)) {
    assert.equal(ex.value.account_created, true)
    assert.equal(ex.value.mail.status, status)
    assert.equal(ex.value.account.must_change_password, true)
  }
  for (const type of ['departments', 'positions', 'job-levels']) assert.equal(operation(`/api/${type}/{id}`, 'delete').responses['204'].content, undefined)
  for (const name of ['CatalogRename', 'CatalogDelete']) assert.ok(schema(name).required.includes('expected_before_name'))
  assert.equal(schema('NextEmployeeCode').properties.is_preview.const, true)
})


test('giới hạn sau trim không bị maxLength/format của raw request chặn trước chuẩn hóa', () => {
  const inputs = [
    [schema('CatalogCreate').properties.name, 100],
    [schema('CatalogRename').properties.name, 100],
    [schema('CreateEmployeeRequest').properties.fullname, 100],
    [schema('CreateEmployeeRequest').properties.company_email, 100],
    [schema('CreateEmployeeRequest').properties.phone, 20],
  ]
  for (const [field, limit] of inputs) {
    assert.equal(field.maxLength, undefined)
    assert.equal(field.format, undefined)
    assert.equal(field['x-normalized-max-length'], limit)
    assert.match(field['x-normalization'], /TrimSpace/)
  }
  assert.equal(schema('CreateEmployeeRequest').properties.company_email['x-normalized-format'], 'email')
  assert.equal(schema('CatalogItem').properties.name.maxLength, 100)
  assert.equal(schema('CatalogRename').properties.expected_before_name.maxLength, 100)
  assert.equal(schema('CatalogRename').properties.expected_before_name['x-normalization'], undefined)
})

test('409 chỉ mô tả xung đột có thể xảy ra ở từng thao tác', () => {
  const codes = (path, method) => {
    const response = resolve(operation(path, method).responses['409'])
    return response.content['application/json'].schema.allOf[1].properties.code.enum
  }
  for (const type of ['departments', 'positions', 'job-levels']) {
    assert.deepEqual(codes(`/api/${type}`, 'post'), ['DUPLICATE_NAME'])
    assert.deepEqual(codes(`/api/${type}/{id}`, 'patch'), ['STALE_DATA', 'DUPLICATE_NAME'])
    assert.deepEqual(codes(`/api/${type}/{id}`, 'delete'), ['STALE_DATA', 'CATALOG_IN_USE'])
  }
  assert.deepEqual(codes('/api/employees', 'post'), ['EMAIL_CONFLICT'])
})
