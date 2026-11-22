import assert from 'node:assert/strict'
import { test } from 'node:test'
import { mkdtemp, rm, symlink, writeFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { argumentsFor, checkFeaturePlan, inspectOpenAPI, inspectTickets, locateMermaid, markdownParts, parseDiagram } from './check-feature-plan.mjs'

const proposal = '| WBS | Đầu ra | Phụ thuộc |\n|---|---|---|\n| EX-01 | Thiết kế | — |\n| EX-02 | API | 01 |\n| EX-03 | UI | 02 |\n'
function task(id, wbs, deps, blocks) {
  return `### [${id}] Đầu ra\n\n${['Mục tiêu','Đầu vào và điều kiện bắt đầu','Phạm vi thực hiện','Hành vi và tình huống cần đạt','Tiêu chí hoàn thành','Kiểm tra và bằng chứng','Đầu ra bàn giao','Ngoài phạm vi'].map(h => `**${h}**\n\nNội dung.\n`).join('\n')}\n- [ ] Kết quả kiểm.\n- Is blocked by: ${deps}.\n- Blocks: ${blocks}.\n- WBS: ${wbs}.\n`
}
const tickets = '| Rank | Mã nguồn | WBS | Is blocked by |\n|---|---|---|---|\n| 1 | EX-T01 | EX-01, EX-02 | — |\n| 2 | EX-T02 | EX-03 | EX-T01 |\n\n' + task('EX-T01', 'EX-01, EX-02', '—', 'EX-T02') + task('EX-T02', 'EX-03', 'EX-T01', '—')
const spec = () => ({ openapi: '3.1.0', paths: { '/things': { get: { operationId: 'listThings', responses: {} } } }, components: { schemas: { Thing: { type: 'object' } } } })

test('fence có dấu ~ hoặc backtick dài, giữ Mermaid và không đọc link/lệnh minh họa', () => {
  const parts = markdownParts('````md\n```sh\ntouch /tmp/khong-chay\n[x](missing.md)\n```\n````\n~~~mermaid\nsequenceDiagram\n A->>B: Hi\n~~~\n[x](real.md)')
  assert.equal(parts.diagrams.length, 1)
  assert.equal(parts.diagrams[0].line, 7)
  assert.equal(parts.prose, '[x](real.md)')
  assert.throws(() => markdownParts('```mermaid\nsequenceDiagram'), /chưa đóng/)
})

test('một task nhận nhiều WBS hợp lệ, mapping phủ hết và graph đúng', () => {
  assert.deepEqual(inspectTickets(proposal, tickets), { wbs: 3, tasks: 2, edges: 1 })
})
test('phát hiện WBS thiếu, giao trùng, thiếu AC và dependency không đối ứng', () => {
  assert.throws(() => inspectTickets(proposal, tickets.replace('- WBS: EX-01, EX-02.', '- WBS: EX-01.')), /chưa giao/)
  assert.throws(() => inspectTickets(proposal, tickets.replace('- WBS: EX-03.', '- WBS: EX-02.')), /giao trùng/)
  assert.throws(() => inspectTickets(proposal, tickets.replace('- [ ] Kết quả kiểm.', '')), /checklist/)
  assert.throws(() => inspectTickets(proposal, tickets.replace('- Blocks: EX-T02.', '- Blocks: —.')), /Blocks/)
  assert.throws(() => inspectTickets(proposal, tickets.replace('- Is blocked by: EX-T01.', '- Is blocked by: DATN-34.')), /chưa hỗ trợ/)
})
test('phát hiện graph vòng lặp và Rank/table lệch description', () => {
  const cycle = tickets.replace('- Is blocked by: —.', '- Is blocked by: EX-T02.').replace('- Blocks: —.', '- Blocks: EX-T01.')
  assert.throws(() => inspectTickets(proposal, cycle), /vòng/)
  assert.throws(() => inspectTickets(proposal, tickets.replace('| 2 | EX-T02 | EX-03 | EX-T01 |', '| 2 | EX-T02 | EX-03 | — |')), /dependency table/)
  assert.throws(() => inspectTickets(proposal, tickets.replace('| 2 | EX-T02', '| 7 | EX-T02')), /Rank/)
})
test('không bỏ qua WBS/heading giống mã task nhưng sai format trong inventory', () => {
  assert.throws(() => inspectTickets(proposal + '| EX-INVALID | Đầu ra chưa giao | 01 |\n', tickets), /WBS chưa hỗ trợ/)
  assert.throws(() => inspectTickets(proposal, tickets + '\n### [EX-TXX] Đầu ra chưa giao\n'), /Heading task sai/)
  assert.throws(() => inspectTickets(proposal, tickets.replace('- WBS: EX-03.', '- WBS: EX-03, EX-INVALID.')), /WBS chưa hỗ trợ/)
  assert.throws(() => inspectTickets(proposal, tickets.replace('| 2 | EX-T02 | EX-03 | EX-T01 |', '| 2 | EX-T02 | EX-03, EX-INVALID | EX-T01 |')), /WBS chưa hỗ trợ/)
})
test('OpenAPI phát hiện operation trùng, ref hỏng/ngoài và giải mã JSON pointer', () => {
  const s = spec()
  s.components.schemas.Alias = { $ref: '#/components/schemas/Thing' }
  assert.equal(inspectOpenAPI(s).length, 1)
  s.paths['/other'] = { post: { operationId: 'listThings' } }
  assert.throws(() => inspectOpenAPI(s), /trùng/)
  delete s.paths['/other']
  s.components.schemas.Alias.$ref = '#/components/schemas/NoSuchThing'
  assert.throws(() => inspectOpenAPI(s), /không tồn tại/)
  s.components.schemas.Alias.$ref = 'https://example.com/schema.json'
  assert.throws(() => inspectOpenAPI(s), /ref ngoài/)
  s.components.schemas['a/b~c'] = { type: 'string' }
  s.components.schemas.Alias.$ref = '#/components/schemas/a~1b~0c'
  assert.doesNotThrow(() => inspectOpenAPI(s))
})

async function fixture(t) {
  const base = await mkdtemp(join(tmpdir(), 'datn-feature-plan-'))
  t.after(() => rm(base, { recursive: true, force: true }))
  await writeFile(join(base, 'proposal.md'), proposal)
  await writeFile(join(base, 'tickets.md'), tickets)
  return base
}
test('kiểm proposal-only không nhận đã kiểm tickets; hash đổi khi nguồn đổi', async t => {
  const base = await fixture(t)
  const before = await checkFeaturePlan({ proposal: 'proposal.md' }, base)
  assert.equal(before.checked.tickets, null)
  assert.equal(before.mapping, null)
  await writeFile(join(base, 'proposal.md'), proposal + '\nNguồn đã đổi\n')
  const after = await checkFeaturePlan({ proposal: 'proposal.md' }, base)
  assert.notEqual(before.hashes[0].sha256, after.hashes[0].sha256)
})
test('link hỏng, đường dẫn tuyệt đối/thoát repo, symlink và .env bị chặn', async t => {
  const base = await fixture(t)
  await writeFile(join(base, 'proposal.md'), '[x](missing.md)')
  await assert.rejects(checkFeaturePlan({ proposal: 'proposal.md' }, base), /ENOENT/)
  await assert.rejects(checkFeaturePlan({ proposal: '/etc/hosts' }, base), /Chỉ đọc|không cho phép/)
  await assert.rejects(checkFeaturePlan({ proposal: '../outside.md' }, base), /không cho phép/)
  await symlink('/etc/hosts', join(base, 'outside.md'))
  await assert.rejects(checkFeaturePlan({ proposal: 'outside.md' }, base), /vượt repo/)
  await assert.rejects(checkFeaturePlan({ proposal: '.env' }, base), /Chỉ đọc|không cho phép/)
  await writeFile(join(base, '.env'), 'synthetic-secret')
  await symlink('.env', join(base, 'alias.md'))
  await assert.rejects(checkFeaturePlan({ proposal: 'alias.md' }, base), /symlink không cho phép/)
  await writeFile(join(base, 'proposal.md'), '[secret](.env)')
  await assert.rejects(checkFeaturePlan({ proposal: 'proposal.md' }, base), /không cho phép/)
})
test('bỏ Mermaid luôn incomplete, không âm thầm nhận full PASS', async t => {
  const base = await fixture(t)
  await writeFile(join(base, 'proposal.md'), '```mermaid\nsequenceDiagram\n A->>B: Hi\n```\n')
  const result = await checkFeaturePlan({ proposal: 'proposal.md', skipMermaid: true }, base)
  assert.equal(result.incomplete, true)
  assert.equal(result.mermaid.parsed, 0)
  await assert.rejects(checkFeaturePlan({ proposal: 'proposal.md', mermaidDir: join(base, 'missing') }, base), /Không có parser/)
})
test('CLI từ chối cờ lạ, cờ lặp/thiếu giá trị; source có thể lặp', () => {
  assert.throws(() => argumentsFor(['--proposal']), /Sai đối số/)
  assert.throws(() => argumentsFor(['--exec', 'rm']), /Sai đối số/)
  assert.throws(() => argumentsFor(['--proposal', 'a.md', '--proposal', 'b.md']), /lặp/)
  assert.deepEqual(argumentsFor(['--proposal', 'a.md', '--source', 'b.md', '--source', 'c.md']).sources, ['b.md', 'c.md'])
})
test('parser Mermaid local thật bắt lỗi chấm phẩy sequence và flowchart hỏng', async t => {
  let directory
  try { directory = await locateMermaid() } catch { t.skip('Chưa có parser local: cần --mermaid-dir/MERMAID_PREVIEW_DIR ở máy kiểm'); return }
  await parseDiagram('sequenceDiagram\n participant A\n participant B\n A->>B: thành công, không báo giả\n', directory)
  await assert.rejects(parseDiagram('sequenceDiagram\n participant A\n participant B\n A->>B: kết luận; không success giả\n', directory), /Parse error|Expecting/)
  await parseDiagram('flowchart LR\n A[Đọc] --> B[Kiểm]\n', directory)
  await assert.rejects(parseDiagram('flowchart LR\n A -->\n', directory))
})
test('thử bộ PCA thật bằng kiểm cấu trúc và parser local, không DB/Jira', async t => {
  let directory
  try { directory = await locateMermaid() } catch { t.skip('Chưa có parser local cho lượt kiểm đầy đủ PCA'); return }
  const result = await checkFeaturePlan({ proposal: 'docs/proposals/backend-profile-change-approval.md', tickets: 'docs/proposals/backend-profile-change-approval-tickets.md', openapi: 'docs/api/profile-change-approval.openapi.json', sqlPlan: 'docs/proposals/backend-profile-change-approval-sql-plan.md', mermaidDir: directory })
  assert.equal(result.incomplete, false)
  assert.equal(result.mapping.tasks, 12)
  assert.equal(result.operations.length, 10)
  assert.equal(result.mermaid.parsed, result.mermaid.found)
  assert.ok(result.mermaid.found > 0)
})

test('dùng lại mapping OAP nhiều WBS/task và cột Summary, không cần sửa tài liệu cũ', async () => {
  const { readFile } = await import('node:fs/promises')
  const read = name => readFile(new URL(`../docs/proposals/${name}`, import.meta.url), 'utf8')
  const result = inspectTickets(await read('backend-organization-accounts-profile.md'), await read('backend-organization-accounts-profile-tickets.md'))
  assert.deepEqual(result, { wbs: 15, tasks: 13, edges: 17 })
})
