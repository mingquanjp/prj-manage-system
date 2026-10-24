// Chỉ kiểm tra điểm vào/đường dẫn; không đọc bí mật, gọi DB hoặc suy quyết định nghiệp vụ.
import { readFile, readdir, realpath, stat } from "node:fs/promises"
import { execFileSync } from "node:child_process"
import { dirname, isAbsolute, relative, resolve, sep } from "node:path"
import { fileURLToPath, pathToFileURL } from "node:url"

const root = resolve(dirname(fileURLToPath(import.meta.url)), "..")
const entries = ["AGENTS.md", "CLAUDE.md", "README.md", "docs/README.md", "web/AGENTS.md", "web/README.md", "web/DESIGN-PILOT.md"]

export function localLinks(markdown) {
  const prose = markdown.replace(/^(`{3,}|~{3,})[^\n]*\n[\s\S]*?^\1[^\n]*$/gm, "")
  return [...prose.matchAll(/!?\[[^\]\n]*\]\((<[^>]+>|[^\s)]+)(?:\s+"[^"]*")?\)/g)]
    .map((match) => match[1].replace(/^<|>$/g, "").split(/[?#]/)[0])
    .filter((target) => target && !/^[a-z][a-z\d+.-]*:/i.test(target) && !target.startsWith("//"))
    .map((target) => decodeURIComponent(target))
}

export function localTarget(base, source, target) {
  const absolute = resolve(base, dirname(source), target)
  const path = relative(base, absolute)
  if (isAbsolute(target) || path === ".." || path.startsWith(`..${sep}`)) {
    throw new Error(`Đường dẫn ngoài repo: ${source} → ${target}`)
  }
  return path
}

export async function requiredFiles(base = root) {
  const required = new Set([...entries, ".gitignore", "scripts/check-harness.mjs", "scripts/check-harness.test.mjs",
    "web/package.json", "web/pnpm-lock.yaml", "web/index.html", "web/tsconfig.json", "web/vite.config.ts", "web/components.json",
    "web/eslint.config.mjs", "web/.prettierrc.json", "web/.prettierignore"])
  for (const source of entries) {
    for (const target of localLinks(await readFile(resolve(base, source), "utf8"))) {
      required.add(localTarget(base, source, target))
    }
  }
  async function collect(directory) {
    for (const item of await readdir(resolve(base, directory), { withFileTypes: true })) {
      const path = `${directory}/${item.name}`
      if (item.isDirectory()) await collect(path)
      else if (item.isFile()) required.add(path)
      else throw new Error(`Không hỗ trợ symlink/file đặc biệt trong nguồn frontend: ${path}`)
    }
  }
  await collect("web/src")
  return [...required].sort()
}

export function checkGitIndex(files, base = root) {
  const tracked = new Set(execFileSync("git", ["ls-files", "--cached", "-z", "--", ...files], {
    cwd: base, encoding: "utf8", stdio: ["ignore", "pipe", "pipe"],
  }).split("\0").filter(Boolean))
  const missing = files.filter((path) => !tracked.has(path))
  if (missing.length) throw new Error(`Chưa được Git theo dõi/stage:\n${missing.join("\n")}`)
  const changed = execFileSync("git", ["diff", "--name-only", "-z", "--", ...files], {
    cwd: base, encoding: "utf8", stdio: ["ignore", "pipe", "pipe"],
  }).split("\0").filter(Boolean)
  if (changed.length) throw new Error(`Nội dung mới chưa được stage:\n${changed.join("\n")}`)
}

export async function checkHarness({ staged = false } = {}) {
  const errors = []
  // Chỉ đi một cấp từ điểm vào và kiểm kê web/src; không nạp toàn bộ research/lịch sử.
  const required = await requiredFiles()
  for (const path of required) {
    try {
      if (!(await stat(resolve(root, path))).isFile()) errors.push(`Không phải file: ${path}`)
    } catch {
      errors.push(`Thiếu file: ${path}`)
    }
  }
  try {
    const ignored = execFileSync("git", ["check-ignore", "--no-index", "--stdin"], {
      cwd: root, input: required.join("\n") + "\n", encoding: "utf8", stdio: ["pipe", "pipe", "pipe"],
    }).trim()
    if (ignored) errors.push(`Nguồn cần thiết đang bị ignore:\n${ignored}`)
  } catch (error) {
    // git check-ignore trả 1 khi không có đường dẫn nào bị ignore.
    if (error.status !== 1) errors.push(`Không kiểm tra được Git ignore: ${error.message}`)
  }
  try {
    if (await realpath(resolve(root, "CLAUDE.md")) !== await realpath(resolve(root, "AGENTS.md"))) {
      errors.push("CLAUDE.md phải trỏ tới AGENTS.md để không phát sinh hai bộ chỉ dẫn")
    }
    for (const source of ["AGENTS.md", "web/AGENTS.md"]) {
      if ((await readFile(resolve(root, source), "utf8")).trim().split("\n").length > 60) {
        errors.push(`Điểm vào quá dài (>60 dòng): ${source}`)
      }
    }
    const pkg = JSON.parse(await readFile(resolve(root, "web/package.json"), "utf8"))
    for (const script of ["dev", "typecheck", "lint", "format:check", "format", "test", "build", "check"]) {
      if (!pkg.scripts?.[script]) errors.push(`Thiếu lệnh frontend: ${script}`)
    }
  } catch (error) {
    errors.push(error.message)
  }
  if (staged) {
    try { checkGitIndex(required) } catch (error) { errors.push(error.message) }
  }
  if (errors.length) throw new Error(errors.join("\n"))
  return required.length
}

if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) {
  try {
    const args = process.argv.slice(2)
    if (args.some((arg) => arg !== "--staged")) throw new Error("Cách dùng: node scripts/check-harness.mjs [--staged]")
    const staged = args.includes("--staged")
    const count = await checkHarness({ staged })
    console.log(`PASS: ${count} file nguồn/harness tồn tại, không bị ignore; instruction và lệnh kiểm tra có mặt.`)
    if (staged) console.log("PASS Git: nguồn cần thiết có trong index và khớp nội dung làm việc; chưa chứng minh đã commit.")
    console.log("Không kiểm tra anchor/ngữ nghĩa nghiệp vụ hoặc UI/API; không tự stage/commit.")
  } catch (error) {
    console.error(`FAIL: ${error.message}`)
    process.exitCode = 1
  }
}
