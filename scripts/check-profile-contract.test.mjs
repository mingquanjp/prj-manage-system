import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { test } from 'node:test'
const spec=JSON.parse(await readFile(new URL('../docs/api/profile-change-approval.openapi.json',import.meta.url),'utf8'))
const oap=JSON.parse(await readFile(new URL('../docs/api/organization-accounts-profile.openapi.json',import.meta.url),'utf8'))
const auth=JSON.parse(await readFile(new URL('../docs/api/auth.openapi.json',import.meta.url),'utf8'))
const methods=new Set(['get','post','patch','delete'])
const ops=Object.entries(spec.paths).flatMap(([path,item])=>Object.entries(item).filter(([m])=>methods.has(m)).map(([method,operation])=>({path,method,operation})))
const schema=n=>spec.components.schemas[n]
const op=(p,m)=>spec.paths[p][m]
function resolve(v){
 if(!v.$ref)return v
 assert.ok(v.$ref.startsWith('#/'),'chỉ ref local, spec chưa được embed runtime')
 return v.$ref.slice(2).split('/').reduce((n,k)=>{const key=k.replaceAll('~1','/').replaceAll('~0','~');assert.ok(n&&Object.hasOwn(n,key),v.$ref);return n[key]},spec)
}
function walk(v){if(!v||typeof v!=='object')return;if(v.$ref)resolve(v);Object.values(v).forEach(walk)}
// Kiểm tập JSON Schema đang dùng cho fixtures; không là OAS validator tổng quát hoặc runtime proof.
function valid(s,v){
 s=resolve(s)
 if(s.allOf&&!s.allOf.every(x=>valid(x,v)))return false
 if(s.oneOf&&s.oneOf.filter(x=>valid(x,v)).length!==1)return false
 if(Object.hasOwn(s,'const')&&v!==s.const)return false
 if(s.enum&&!s.enum.includes(v))return false
 if(s.type==='null'&&v!==null)return false
 if(s.type==='object'){
  if(!v||typeof v!=='object'||Array.isArray(v))return false
  if((s.required??[]).some(k=>!Object.hasOwn(v,k)))return false
  if(s.minProperties&&Object.keys(v).length<s.minProperties)return false
  for(const [k,x]of Object.entries(v)){
   if(s.additionalProperties===false&&!Object.hasOwn(s.properties??{},k))return false
   if(s.properties?.[k]&&!valid(s.properties[k],x))return false
  }
 }
 if(s.type==='array'){
  if(!Array.isArray(v)||v.length<(s.minItems??0)||v.length>(s.maxItems??Infinity))return false
  if(s.uniqueItems&&new Set(v.map(x=>JSON.stringify(x))).size!==v.length)return false
  if(s.items&&!v.every(x=>valid(s.items,x)))return false
 }
 if(s.type==='integer'&&(!Number.isSafeInteger(v)||v<(s.minimum??-Infinity)||v>(s.maximum??Infinity)))return false
 if(s.type==='string'){
  if(typeof v!=='string'||[...v].length<(s.minLength??0)||[...v].length>(s.maxLength??Infinity))return false
  if(s.pattern&&!new RegExp(s.pattern,'u').test(v))return false
  if(s.format==='date-time'&&!Number.isFinite(Date.parse(v)))return false
 }
 return true
}
const sample=op('/api/employees/{id}/profile-change-requests','post').responses['201'].content['application/json'].examples.demo.value
const result=(status)=>({...sample,status,resolved_by:18,resolved_at:'2026-11-21T10:00:00+07:00',decision_reason:status==='REJECTED'?'Cần đối chiếu lại':null})
test('đúng10 thao tác, 7 ưu tiên/3 HR tiếp nối, mọi ref local tồn tại và planned',()=>{
 const expected=[['post','/api/employees/{id}/profile-change-requests'],['get','/api/profile-change-requests'],['get','/api/profile-change-requests/{id}'],['post','/api/profile-change-requests/{id}/approve'],['post','/api/profile-change-requests/{id}/reject'],['get','/api/notifications'],['post','/api/notifications/{id}/read'],['get','/api/employees'],['get','/api/employees/{id}/administrative-profile'],['patch','/api/employees/{id}']]
 assert.deepEqual(ops.map(x=>`${x.method} ${x.path}`).sort(),expected.map(([m,p])=>`${m} ${p}`).sort())
 assert.equal(new Set(ops.map(x=>x.operation.operationId)).size,10)
 assert.equal(ops.filter(x=>x.operation['x-delivery-phase']==='F1-F3').length,7)
 for(const {operation}of ops)assert.equal(operation['x-implementation-status'],'planned')
 walk(spec)
})
test('phiên thường/cookie AND csrf/Origin; PCA64KiB không thay auth hoặc OAP8KiB',()=>{
 assert.deepEqual(spec.components.securitySchemes,auth.components.securitySchemes)
 for(const {method,operation:o}of ops){
  assert.equal(o['x-requires-business-session'],true)
  assert.equal(o['x-reject-unknown-query'],true)
  assert.equal(o['x-reject-repeated-query'],true)
  assert.deepEqual(o.security,[{cookieAuth:[],...(method!=='get'?{csrfHeader:[]}:{})}])
  if(method!=='get'){
   assert.equal(o['x-max-body-bytes'],65536)
   assert.equal(o['x-reject-duplicate-json-keys'],true)
   assert.ok(o.parameters.map(resolve).some(p=>p.name==='Origin'&&p.required&&p.in==='header'))
   assert.ok(o.responses['413']&&o.responses['415'])
  }else assert.equal(o.requestBody,undefined)
  for(const r of Object.values(o.responses).map(resolve)){
   assert.equal(r.headers['Cache-Control'].schema.const,'no-store')
   assert.equal(r.headers['Set-Cookie'],undefined)
  }
 }
 for(const item of Object.values(oap.paths))for(const [m,o]of Object.entries(item))if(methods.has(m)&&m!=='get')assert.equal(o['x-max-body-bytes'],8192)
})
test('quyền HR tường minh; self/own không mở Admin bypass và key chỉ cho submit',()=>{
 for(const {path,method,operation:o}of ops){
  const hr=path.endsWith('/approve')||path.endsWith('/reject')||o['x-delivery-phase']==='F4'
  assert.deepEqual(o['x-required-system-roles'],hr?['HR']:[])
  assert.equal(o.parameters.map(resolve).filter(p=>p.name==='Idempotency-Key').length,path.endsWith('/profile-change-requests')&&method==='post'?1:0)
 }
 const key=spec.components.parameters.IdempotencyKey.schema
 assert.equal(op('/api/employees/{id}/profile-change-requests','post')['x-resource-access'],'self-employee-id-linked-to-actor-user-id')
 assert.deepEqual(op('/api/profile-change-requests','get')['x-scope-role-requirements'],{review:['HR']})
 assert.equal(op('/api/profile-change-requests/{id}','get')['x-resource-access'],'self-owner-or-current-HR')
 for(const [p,m]of [['/api/notifications','get'],['/api/notifications/{id}/read','post']])assert.equal(op(p,m)['x-resource-access'],'own-notification-recipient-user-id')
 for(const action of ['approve','reject'])assert.equal(op(`/api/profile-change-requests/{id}/${action}`,'post')['x-allow-self-resolution'],true)
 assert.ok(valid(key,'d288552d-006d-4a88-b4c7-cbd68ea17921'))
 assert.ok(!valid(key,'arbitrary-key'))
 assert.ok(!valid(key,'d288552d-006d-1a88-b4c7-cbd68ea17921'))
})
test('allowlist patch không role/email/cơ cấu; address không null, phone nullable, empty patch không hợp lệ',()=>{
 assert.ok(valid(schema('CreateProfileRequest'),{expected_before_values:{address:null},proposed_values:{address:'Địa chỉ demo'}}))
 assert.ok(valid(schema('ProfileProposed'),{phone:null}))
 assert.ok(!valid(schema('ProfileProposed'),{address:null}))
 assert.ok(!valid(schema('ProfileProposed'),{}))
 for(const key of ['role','system_roles','company_email','position_id','status','schema_version'])assert.ok(!valid(schema('ProfileProposed'),{fullname:'Tên demo',[key]:'x'}))
 for(const k of ['fullname','phone','address'])assert.equal(schema('ProfileProposed').properties[k].maxLength,undefined,'raw không bị maxLength chặn trước trim')
 assert.equal(schema('ProfileProposed').properties.address['x-normalized-max-length'],2000)
 assert.equal(schema('ProfileBefore').properties.address.oneOf[0].maxLength,undefined,'before legacy không bị cắt theo max proposed')
 assert.deepEqual(schema('CreateProfileRequest')['x-equal-key-sets'],['expected_before_values','proposed_values'])
})
test('pending/approved/rejected giữ đủ terminal actor/time/reason trong schema, không chỉ mô tả',()=>{
 assert.ok(valid(schema('RequestDetail'),sample))
 assert.ok(valid(schema('RequestDetail'),result('APPROVED')))
 assert.ok(valid(schema('RequestDetail'),result('REJECTED')))
 assert.ok(!valid(schema('RequestDetail'),{...sample,resolved_by:18}))
 assert.ok(!valid(schema('RequestDetail'),{...result('APPROVED'),resolved_by:null}))
 assert.ok(!valid(schema('RequestDetail'),{...result('REJECTED'),decision_reason:null}))
 assert.ok(!valid(schema('RequestDetail'),{...sample,password:'hidden'}))
 assert.deepEqual(Object.keys(schema('ApproveProfileRequest').properties),[])
 assert.ok(!valid(schema('RejectProfileRequest'),{}))
})
test('list tối thiểu không PII, chi tiết theo mục đích; notification không lộ patch và read204 không JSON',()=>{
 for(const n of ['RequestSummary','EmployeeSummary','Notification']){
  assert.equal(schema(n).additionalProperties,false)
  for(const field of ['address','phone','password','token','secret','before_values','explanation','decision_reason'])assert.equal(Object.hasOwn(schema(n).properties,field),false,`${n}.${field}`)
 }
 assert.equal(schema('NotificationPage').properties.unread_count.minimum,0)
 const read=op('/api/notifications/{id}/read','post')
 assert.equal(read.responses['204'].content,undefined)
 assert.equal(op('/api/notifications','get').parameters.map(resolve).find(p=>p.name==='filter').schema.default,'all')
 assert.equal(op('/api/profile-change-requests','get').parameters.map(resolve).find(p=>p.name==='scope').schema.default,'self')
})
test('409 đúng từng thao tác, mất commit có mô tả đối soát, F4 không đổi self OAP',()=>{
 const codes=(p,m)=>op(p,m).responses['409'].content['application/json'].schema.allOf[1].properties.code.enum
 assert.deepEqual(codes('/api/profile-change-requests/{id}/approve','post'),['STALE_DATA','REQUEST_ALREADY_RESOLVED','EMPLOYEE_INACTIVE'])
 assert.deepEqual(codes('/api/profile-change-requests/{id}/reject','post'),['REQUEST_ALREADY_RESOLVED'])
 assert.deepEqual(codes('/api/employees/{id}/profile-change-requests','post'),['STALE_DATA','IDEMPOTENCY_KEY_REUSED','IDEMPOTENCY_KEY_EXPIRED'])
 assert.match(op('/api/employees/{id}/profile-change-requests','post').responses['503'].description,/không tự replay/)
 assert.deepEqual(oap.paths['/api/employees/{id}'].get['x-required-system-roles'],[])
 assert.deepEqual(schema('EditAdministrativeProfile')['x-equal-key-sets'],['expected_before_values','changes'])
})

test('ví dụ request/response khớp schema thiết kế, JSON before/proposed không lẫn schema_version nội bộ',()=>{
 assert.deepEqual(schema('RequestDetail')['x-preserve-raw-fields'],['before_values','current_values'])
 assert.deepEqual(schema('RequestDetail')['x-normalized-fields'],['proposed_values','explanation','decision_reason'])
 for(const {operation:o}of ops){
  for(const content of [o.requestBody?.content,...Object.values(o.responses).map(r=>resolve(r).content)].filter(Boolean)){
   for(const media of Object.values(content)){
    if(media.example)assert.ok(valid(media.schema,media.example))
    for(const ex of Object.values(media.examples??{}))assert.ok(valid(media.schema,ex.value))
   }
  }
 }
 assert.ok(!valid(schema('CreateProfileRequest'),{expected_before_values:{address:null},proposed_values:{schema_version:1,address:'Demo'}}))
})

test('bộ ticket có12 task, WBS1:1 và17 cạnh dependency không vòng lặp, F4 không chặn demo',async()=>{
 const tickets=await readFile(new URL('../docs/proposals/backend-profile-change-approval-tickets.md',import.meta.url),'utf8')
 const chunks=tickets.split(/^### \[(PCA-T\d{2})\] /m)
 const graph=new Map()
 for(let i=1;i<chunks.length;i+=2){
  const key=chunks[i],body=chunks[i+1]
  assert.ok(!graph.has(key),key)
  const line=body.match(/^- Is blocked by: (.+)\.$/m)
  assert.ok(line,key)
  graph.set(key,[...line[1].matchAll(/PCA-T\d{2}/g)].map(x=>x[0]))
  const n=key.slice(-2)
  assert.match(body,new RegExp(`WBS: PCA-${n}\\.`))
  for(const heading of ['Mục tiêu','Đầu vào và điều kiện bắt đầu','Phạm vi thực hiện','Hành vi và tình huống cần đạt','Tiêu chí hoàn thành','Kiểm tra và bằng chứng','Đầu ra bàn giao','Ngoài phạm vi'])assert.ok(body.includes(`**${heading}**`),`${key} ${heading}`)
 }
 assert.equal(graph.size,12)
 assert.equal([...graph.values()].flat().length,17)
 function visit(k,stack=[]){assert.ok(graph.has(k),k);assert.ok(!stack.includes(k),`dependency cycle: ${stack} ${k}`);for(const d of graph.get(k))visit(d,[...stack,k])}
 for(const key of graph.keys())visit(key)
 const ancestors=k=>graph.get(k).flatMap(d=>[d,...ancestors(d)])
 assert.ok(!ancestors('PCA-T10').some(k=>['PCA-T11','PCA-T12'].includes(k)))
 // Blocks ở description phải cùng đúng các cạnh trực tiếp, không nhập hai link ngược.
 for(let i=1;i<chunks.length;i+=2){
  const key=chunks[i],body=chunks[i+1]
  const actual=[...body.match(/^- Blocks: (.+)\.$/m)[1].matchAll(/PCA-T\d{2}/g)].map(x=>x[0]).sort()
  const expected=[...graph].filter(([,deps])=>deps.includes(key)).map(([k])=>k).sort()
  assert.deepEqual(actual,expected)
 }
})
