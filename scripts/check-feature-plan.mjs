// Kiểm cấu trúc tài liệu, không chạy lệnh/SQL/API trong đầu vào hoặc suy nghiệp vụ đã đạt.
import { readFile, readdir, realpath, stat } from 'node:fs/promises'
import { createHash } from 'node:crypto'
import { homedir } from 'node:os'
import { dirname, extname, isAbsolute, relative, resolve, sep } from 'node:path'
import { fileURLToPath, pathToFileURL } from 'node:url'
import { localLinks, localTarget } from './check-harness.mjs'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const methods = new Set(['get', 'post', 'put', 'patch', 'delete', 'head', 'options', 'trace'])
const fail = message => { throw new Error(message) }
const codes = (text, pattern) => [...text.matchAll(pattern)].map(m => m[0])
function taskCodes(text) {
  if (text === '—') return []
  if (!/^[A-Z][A-Z0-9]*-T\d+(?:,\s*[A-Z][A-Z0-9]*-T\d+)*$/.test(text ?? '')) fail('Format dependency chưa hỗ trợ, không được bỏ qua token ngoài graph')
  return codes(text, /[A-Z][A-Z0-9]*-T\d+/g)
}
function wbsCodes(text) {
  if (!/^[A-Z][A-Z0-9]*-\d+(?:,\s*[A-Z][A-Z0-9]*-\d+)*$/.test(text ?? '')) fail(`Format WBS chưa hỗ trợ: ${text}`)
  return codes(text, /[A-Z][A-Z0-9]*-\d+/g)
}

// Đọc fence theo dòng để không bỏ sót lỗi Mermaid hoặc liên kết do fence chưa đóng.
export function markdownParts(text) {
  const prose = [], diagrams = []
  let fence, body = []
  for (const [i, line] of text.split('\n').entries()) {
    if (!fence) {
      const open = line.match(/^\s{0,3}(`{3,}|~{3,})([^\n]*)$/)
      if (open) { fence = { marker: open[1], language: open[2].trim(), line: i + 1 }; body = [] }
      else prose.push(line)
    } else if (new RegExp(`^\\s{0,3}${fence.marker[0]}{${fence.marker.length},}\\s*$`).test(line)) {
      if (fence.language === 'mermaid') diagrams.push({ source: body.join('\n') + '\n', line: fence.line })
      fence = undefined
    } else body.push(line)
  }
  if (fence) fail(`Fence chưa đóng tại dòng ${fence.line}`)
  return { prose: prose.join('\n'), diagrams }
}

export function inspectOpenAPI(spec) {
  const operations = [], ids = new Set()
  function walk(v) {
    if (!v || typeof v !== 'object') return
    if (v.$ref) {
      if (!v.$ref.startsWith('#/')) fail(`Chưa hỗ trợ ref ngoài: ${v.$ref}`)
      let target = spec
      for (const segment of v.$ref.slice(2).split('/')) {
        const key = segment.replaceAll('~1', '/').replaceAll('~0', '~')
        if (!target || !Object.hasOwn(target, key)) fail(`Ref không tồn tại: ${v.$ref}`)
        target = target[key]
      }
    }
    Object.values(v).forEach(walk)
  }
  if (!/^3\./.test(spec.openapi ?? '') || !spec.paths) fail('Cần OpenAPI 3.x có paths')
  walk(spec)
  for (const [path, item] of Object.entries(spec.paths)) {
    for (const [method, op] of Object.entries(item)) if (methods.has(method)) {
      if (!path.startsWith('/') || !op.operationId || ids.has(op.operationId)) fail(`Path/operationId thiếu hoặc trùng: ${method} ${path}`)
      ids.add(op.operationId)
      operations.push(`${method.toUpperCase()} ${path} (${op['x-implementation-status'] ?? 'không khai báo trạng thái'})`)
    }
  }
  if (!operations.length) fail('Không có operation trong spec đã truyền')
  return operations
}

function tables(prose) {
  const result = []
  let rows = []
  for (const line of [...prose.split('\n'), '']) {
    if (/^\s*\|/.test(line)) rows.push(line.trim().replace(/^\||\|$/g, '').split('|').map(x => x.trim()))
    else if (rows.length) { result.push(rows); rows = [] }
  }
  return result
}

export function inspectTickets(proposal, ticketText) {
  const wbs = []
  for (const table of tables(markdownParts(proposal).prose)) {
    const col = table[0].indexOf('WBS')
    if (col < 0) continue
    for (const row of table.slice(2)) {
      const units = wbsCodes(row[col])
      if (units.length !== 1) fail('Mỗi dòng bảng WBS proposal cần một mã đầu ra')
      wbs.push(units[0])
    }
  }
  if (!wbs.length || new Set(wbs).size !== wbs.length) fail('WBS proposal thiếu/không hỗ trợ hoặc bị trùng')
  const prose = markdownParts(ticketText).prose
  for (const line of prose.split('\n')) {
    if (/^###\s*\[/.test(line) && !/^### \[[A-Z][A-Z0-9]*-T\d+\] .+$/.test(line)) fail(`Heading task sai format: ${line}`)
  }
  const chunks = prose.split(/^### \[([A-Z][A-Z0-9]*-T\d+)\] .+$/m)
  const tasks = new Map(), owners = new Map()
  const headings = ['Mục tiêu', 'Đầu vào và điều kiện bắt đầu', 'Phạm vi thực hiện', 'Hành vi và tình huống cần đạt', 'Tiêu chí hoàn thành', 'Kiểm tra và bằng chứng', 'Đầu ra bàn giao', 'Ngoài phạm vi']
  for (let i = 1; i < chunks.length; i += 2) {
    const id = chunks[i], body = chunks[i + 1]
    if (tasks.has(id)) fail(`Task trùng: ${id}`)
    const field = name => {
      const match = body.match(new RegExp(`^[-*] ${name}: (.+)\\.$`, 'm'))
      if (!match) fail(`${id} thiếu dòng ${name}`)
      return taskCodes(match[1])
    }
    const mapped = body.match(/WBS: ([^.;\n]+)/)?.[1]
    const units = wbsCodes(mapped)
    for (const unit of units) {
      if (!wbs.includes(unit) || owners.has(unit)) fail(`${id}: WBS lạ hoặc giao trùng ${unit}`)
      owners.set(unit, id)
    }
    for (const heading of headings) {
      if (!body.includes(`**${heading}**`) && !body.includes(`### ${heading}`)) fail(`${id} thiếu mục ${heading}`)
    }
    if (!/^- \[[ x]\] .+/m.test(body)) fail(`${id} thiếu checklist AC`)
    const deps = field('Is blocked by'), blocks = field('Blocks')
    if (new Set(deps).size !== deps.length || new Set(blocks).size !== blocks.length) fail(`${id} lặp cạnh`)
    tasks.set(id, { deps, blocks })
  }
  if (!tasks.size) fail('Không có heading task theo format hỗ trợ')
  for (const unit of wbs) if (!owners.has(unit)) fail(`WBS chưa giao: ${unit}`)
  const ordered = [...tasks.keys()]
  for (const [id, task] of tasks) {
    for (const dep of task.deps) {
      if (!tasks.has(dep)) fail(`${id} phụ thuộc task không có: ${dep}`)
      if (!tasks.get(dep).blocks.includes(id)) fail(`${id}/${dep}: Blocks không đối ứng`)
    }
    const actual = task.blocks.slice().sort()
    const expected = [...tasks].filter(([, t]) => t.deps.includes(id)).map(([key]) => key).sort()
    if (JSON.stringify(actual) !== JSON.stringify(expected)) fail(`${id}: Blocks dư/thiếu/ngoài graph`)
  }
  const done = new Set()
  function visit(id, stack = []) {
    if (stack.includes(id)) fail(`Dependency có vòng: ${[...stack, id].join(' → ')}`)
    if (done.has(id)) return
    for (const dep of tasks.get(id).deps) visit(dep, [...stack, id])
    done.add(id)
  }
  ordered.forEach(id => visit(id))
  for (const [id, { deps }] of tasks) for (const dep of deps) {
    if (ordered.indexOf(dep) >= ordered.indexOf(id)) fail(`Rank task chưa theo phụ thuộc: ${dep} → ${id}`)
  }
  const codeColumn = header => header.findIndex(h => ['Mã nguồn', 'Mã nội bộ / Summary'].includes(h))
  const rankTable = tables(prose).find(t => t[0].includes('Rank') && codeColumn(t[0]) >= 0 && t[0].includes('WBS'))
  if (!rankTable) fail('Thiếu bảng Rank/Mã nguồn/WBS')
  const header = rankTable[0], rows = rankTable.slice(2)
  const rowCode = row => row[codeColumn(header)].match(/^(?:\[)?([A-Z][A-Z0-9]*-T\d+)(?:\]|$)/)?.[1]
  if (JSON.stringify(rows.map(rowCode)) !== JSON.stringify(ordered)) fail('Rank table khác thứ tự/đủ task descriptions')
  for (const [i, row] of rows.entries()) {
    if (Number(row[header.indexOf('Rank')]) !== i + 1) fail('Rank không tăng liên tiếp từ 1')
    const id = ordered[i]
    const units = wbsCodes(row[header.indexOf('WBS')]).sort()
    const expected = [...owners].filter(([, owner]) => owner === id).map(([unit]) => unit).sort()
    if (JSON.stringify(units) !== JSON.stringify(expected)) fail(`${id}: WBS table khác description`)
    const deps = taskCodes(row[header.indexOf('Is blocked by')]).sort()
    if (JSON.stringify(deps) !== JSON.stringify(tasks.get(id).deps.slice().sort())) fail(`${id}: dependency table khác description`)
    if (header.includes('Blocks')) {
      const blocks = taskCodes(row[header.indexOf('Blocks')]).sort()
      if (JSON.stringify(blocks) !== JSON.stringify(tasks.get(id).blocks.slice().sort())) fail(`${id}: Blocks table khác description`)
    }
  }
  return { wbs: wbs.length, tasks: tasks.size, edges: [...tasks.values()].reduce((n, t) => n + t.deps.length, 0) }
}

export async function locateMermaid(explicit) {
  const requested = explicit ?? process.env.MERMAID_PREVIEW_DIR
  const candidates = requested ? [requested] : [
    '/Applications/Visual Studio Code.app/Contents/Resources/app/extensions/mermaid-markdown-features/markdown-preview-out',
    resolve(homedir(), 'Applications/Visual Studio Code.app/Contents/Resources/app/extensions/mermaid-markdown-features/markdown-preview-out'),
  ].filter(Boolean)
  for (const dir of candidates) {
    try { if ((await stat(dir)).isDirectory()) return resolve(dir) } catch { /* Thử đường dẫn kế tiếp. */ }
  }
  fail('Không có parser Mermaid local. Dùng --mermaid-dir thư mục bundle VS Code tin cậy; không tự cài dependency')
}

export async function parseDiagram(source, directory) {
  const type = source.trimStart().split(/\s/)[0]
  const prefix = type === 'sequenceDiagram' ? 'sequenceDiagram-' : ['flowchart', 'graph'].includes(type) ? 'flowDiagram-' : undefined
  if (!prefix) fail(`Mermaid chưa hỗ trợ loại ${type}`)
  const files = (await readdir(directory)).filter(n => n.startsWith(prefix) && n.endsWith('.js'))
  if (files.length !== 1) fail(`Không xác định duy nhất parser ${prefix} trong bundle`)
  const { diagram } = await import(pathToFileURL(resolve(directory, files[0])).href)
  if (!diagram?.parser || !diagram.db) fail('Bundle Mermaid không có API parser/db hỗ trợ')
  const db = diagram.db
  db.clear()
  const parser = diagram.parser.parser ?? diagram.parser
  // Flowchart grammar gọi sanitizer cần DOM; chỉ thay callbacks tạo model, giữ lexer thật.
  // Đây là kiểm cú pháp, không kiểm semantic model/render hoặc cho chạy HTML trong diagram.
  parser.yy = type === 'sequenceDiagram' ? db : Object.fromEntries(
    Object.entries(db).map(([key, value]) => [key, typeof value === 'function' ? () => [] : value]),
  )
  await diagram.parser.parse(source)
  return files[0]
}

export async function checkFeaturePlan(options, base = root) {
  if (!options.proposal) fail('Cần --proposal')
  const files = [options.proposal, options.tickets, options.openapi, options.sqlPlan, ...(options.sources ?? [])].filter(Boolean)
  const contents = new Map(), hashes = [], diagrams = []
  let links = 0
  const canonicalBase = await realpath(base)
  const forbidden = /(^|[/\\])(?:\.env(?:\..*)?|\.local|\.config)(?:[/\\]|$)/
  async function inside(file) {
    const resolved = resolve(base, file), rel = relative(base, resolved)
    if (isAbsolute(file) || rel === '..' || rel.startsWith(`..${sep}`) || forbidden.test(rel)) fail(`Đường dẫn nguồn không cho phép: ${file}`)
    const actual = await realpath(resolved), realRel = relative(canonicalBase, actual)
    if (forbidden.test(realRel)) fail(`Đích symlink không cho phép: ${file}`)
    if (realRel === '..' || realRel.startsWith(`..${sep}`) || !(await stat(actual)).isFile()) fail(`Nguồn/link vượt repo hoặc không là file: ${file}`)
    return actual
  }
  for (const file of new Set(files)) {
    if (!['.md', '.json'].includes(extname(file))) fail(`Chỉ đọc nguồn .md/.json: ${file}`)
    const text = await readFile(await inside(file), 'utf8')
    contents.set(file, text)
    hashes.push({ file, sha256: createHash('sha256').update(text).digest('hex') })
    if (file.endsWith('.md')) {
      const parts = markdownParts(text)
      diagrams.push(...parts.diagrams.map(d => ({ ...d, file })))
      for (const target of localLinks(parts.prose)) {
        await inside(localTarget(base, file, target)); links++
      }
    }
  }
  const result = { checked: { proposal: options.proposal, tickets: options.tickets ?? null, openapi: options.openapi ?? null, sqlPlan: options.sqlPlan ?? null }, hashes, links,
    operations: options.openapi ? inspectOpenAPI(JSON.parse(contents.get(options.openapi))) : [],
    mapping: options.tickets ? inspectTickets(contents.get(options.proposal), contents.get(options.tickets)) : null,
    mermaid: { found: diagrams.length, parsed: 0, parsers: [] }, incomplete: false }
  if (diagrams.length && options.skipMermaid) result.incomplete = true
  else if (diagrams.length) {
    const directory = await locateMermaid(options.mermaidDir)
    for (const diagram of diagrams) {
      try {
        const parser = await parseDiagram(diagram.source, directory)
        if (!result.mermaid.parsers.includes(parser)) result.mermaid.parsers.push(parser)
        result.mermaid.parsed++
      } catch (error) { fail(`${diagram.file}:${diagram.line}: Mermaid: ${error.message}`) }
    }
  }
  return result
}

export function argumentsFor(args) {
  const options = { sources: [] }, keys = { '--proposal': 'proposal', '--tickets': 'tickets', '--openapi': 'openapi', '--sql-plan': 'sqlPlan', '--source': 'sources', '--mermaid-dir': 'mermaidDir' }
  for (let i = 0; i < args.length; i++) {
    if (args[i] === '--skip-mermaid') { options.skipMermaid = true; continue }
    const key = keys[args[i]], value = args[++i]
    if (!key || !value || value.startsWith('--')) fail('Sai đối số; dùng --proposal file [--tickets file --openapi file --sql-plan file --source file --mermaid-dir dir --skip-mermaid]')
    if (key === 'sources') options.sources.push(value)
    else { if (options[key]) fail(`Đối số lặp: ${key}`); options[key] = value }
  }
  return options
}

if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) {
  try {
    const result = await checkFeaturePlan(argumentsFor(process.argv.slice(2)))
    console.log(JSON.stringify(result, null, 2))
    console.log(result.incomplete ? 'INCOMPLETE: đã bỏ kiểm Mermaid.' : 'PASS cấu trúc các file được yêu cầu; không là kiểm ngữ nghĩa, approval, SQL/HTTP/DB/browser hoặc visual render.')
    if (result.incomplete) process.exitCode = 2
  } catch (error) { console.error(`FAIL: ${error.message}`); process.exitCode = 1 }
}
