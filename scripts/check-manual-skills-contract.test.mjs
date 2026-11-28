// Kiểm contract thiết kế; không gọi HTTP/DB hoặc nhận operation planned là đã chạy.
import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { inspectOpenAPI } from './check-feature-plan.mjs'

const spec = JSON.parse(readFileSync(new URL('../docs/api/manual-skills.openapi.json', import.meta.url)))
const pca = JSON.parse(readFileSync(new URL('../docs/api/profile-change-approval.openapi.json', import.meta.url)))
const methods = new Set(['get', 'post', 'patch', 'delete'])
const operations = Object.entries(spec.paths).flatMap(([path, item]) => Object.entries(item)
  .filter(([method]) => methods.has(method)).map(([method, op]) => ({path, method, op})))
const schemas = spec.components.schemas
function resolve(value) {
  if (value.$ref) return spec.components.schemas[value.$ref.split('/').at(-1)]
  return value
}
function parameter(value) {
  if (value.$ref) return spec.components.parameters[value.$ref.split('/').at(-1)]
  return value
}

test('inventory và ref toàn bộ tồn tại, chỉ thao tác planned, không mở workflow ngoài đợt', () => {
  assert.equal(inspectOpenAPI(spec).length, 14)
  for (const {path, op} of operations) {
    assert.equal(op['x-implementation-status'], 'planned')
    assert.doesNotMatch(path, /withdraw|supplement|upload|reset|remove/)
  }
  assert.deepEqual(schemas.RequestSummary.properties.source, {type:'string', const:'MANUAL'})
  assert.deepEqual(schemas.RequestSummary.properties.operation.enum, ['ADD','UPDATE'])
})

test('cookie và CSRF cùng điều kiện; Origin, body bytes và no-store đúng từng route', () => {
  for (const {path, method, op} of operations) {
    assert.equal(op.security.length, 1)
    assert.ok(Object.hasOwn(op.security[0], 'cookieAuth'))
    if (method !== 'get') {
      assert.ok(Object.hasOwn(op.security[0], 'csrfHeader'))
      assert.ok(op.parameters.map(parameter).some(p => p.name === 'Origin' && p.required))
      let bodyLimit = 8192
      if (path.includes('skill-change-requests')) bodyLimit = 65536
      assert.equal(op['x-body-limit-bytes'], bodyLimit)
      assert.deepEqual(Object.keys(op.requestBody.content), ['application/json'])
    }
    for (const [code, response] of Object.entries(op.responses)) {
      assert.equal(response.headers['Cache-Control'].schema.const, 'no-store')
      assert.ok(!Object.keys(response.headers).some(h => h.toLowerCase() === 'set-cookie'))
      if (code === '204') assert.equal(response.content, undefined)
    }
  }
})

test('DTO theo mục đích không lộ credential hoặc hồ sơ hành chính', () => {
  const forbidden = new Set(['password','password_hash','csrf','token','company_email','phone','address','audit_events'])
  const seen = new Set()
  function visit(schema) {
    const node = resolve(schema)
    if (seen.has(node)) return
    seen.add(node)
    for (const [key, property] of Object.entries(node.properties ?? {})) {
      assert.ok(!forbidden.has(key), key)
      visit(property)
    }
    if (node.items) visit(node.items)
    for (const variant of node.oneOf ?? []) visit(variant)
  }
  for (const key of ['OfficialPage','RequestPage','RequestDetail']) visit(schemas[key])
  for (const key of ['Submit','Decision','SkillCreate','SkillChange']) assert.equal(schemas[key].additionalProperties, false)
  for (const key of ['actor_id','status','operation','source','schema_version']) assert.ok(!Object.hasOwn(schemas.Submit.properties,key))
})

test('absence, years decimal và terminal giữ ranh giới trước/đề nghị/quyết định', () => {
  assert.ok(schemas.Submit.properties.expected_before_values.oneOf.some(v => v.type === 'null'))
  const years = schemas.ProposedValues.properties.years_experience.oneOf[0]
  const pattern = new RegExp(years.pattern)
  for (const value of ['0','0.0','999.9']) assert.ok(pattern.test(value))
  for (const value of ['-1','1000','1.23','1e1']) assert.ok(!pattern.test(value))
  assert.equal(schemas.Values.properties.note.oneOf[0].maxLength, undefined)
  assert.ok(schemas.Decision.required.includes('reason'))
  assert.ok(schemas.Decision.required.includes('expected_request_version'))
  assert.equal(schemas.RequestDetail.allOf[0].then.properties.resolved_by.type, 'null')
  assert.equal(schemas.RequestDetail.allOf[0].else.properties.decision_reason.minLength, 1)
})

test('key chỉ submit; role mutation và conflict theo từng thao tác', () => {
  for (const {path, method, op} of operations) {
    const key = op.parameters.map(parameter).find(p => p.name === 'Idempotency-Key')
    const submit = method === 'post' && path === '/api/employees/{id}/skill-change-requests'
    assert.equal(Boolean(key), submit)
    if (method !== 'get' && !submit) assert.equal(op['x-required-role'], 'DIRECTOR')
    if (path.endsWith('/reject')) {
      const codes = op.responses['409'].content['application/json'].schema.allOf[1].properties.code.enum
      assert.ok(!codes.includes('TARGET_INACTIVE'))
      assert.ok(!codes.includes('STALE_DATA'))
    }
  }
})

test('notification canonical hỗ trợ đúng các cặp hồ sơ/kỹ năng cho đọc và đánh dấu đọc', () => {
  const notification = pca.components.schemas.Notification
  assert.deepEqual(notification.properties.notification_type.enum, [
    'PROFILE_CHANGE_SUBMITTED','PROFILE_CHANGE_APPROVED','PROFILE_CHANGE_REJECTED',
    'SKILL_CHANGE_SUBMITTED','SKILL_CHANGE_APPROVED','SKILL_CHANGE_REJECTED',
  ])
  assert.deepEqual(notification.properties.resource_type.enum, ['PROFILE_CHANGE_REQUEST','SKILL_CHANGE_REQUEST'])
  assert.equal(notification.properties.notification_type['x-planned-values'], undefined)
  assert.equal(notification.properties.resource_type['x-planned-values'], undefined)
  assert.deepEqual(notification.allOf[0].if.properties.notification_type.enum, [
    'PROFILE_CHANGE_SUBMITTED','PROFILE_CHANGE_APPROVED','PROFILE_CHANGE_REJECTED',
  ])
  assert.equal(notification.allOf[0].then.properties.resource_type.const, 'PROFILE_CHANGE_REQUEST')
  assert.equal(notification.allOf[0].else.properties.resource_type.const, 'SKILL_CHANGE_REQUEST')
  assert.equal(pca.paths['/api/notifications'].get['x-implementation-status'], 'implemented')
  assert.equal(pca.paths['/api/notifications/{id}/read'].post['x-implementation-status'], 'implemented')
})
