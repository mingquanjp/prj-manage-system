import assert from "node:assert/strict"
import { test } from "node:test"
import { execFileSync } from "node:child_process"
import { mkdtemp, readFile, rm, writeFile } from "node:fs/promises"
import { tmpdir } from "node:os"
import { join } from "node:path"
import { checkGitIndex, localLinks, localTarget, requiredFiles } from "./check-harness.mjs"

test("kiểm kê giữ workflow Jira–GitHub và mẫu bàn giao solo", async () => {
  const files = await requiredFiles()
  for (const path of ["docs/ai/workflow.md", "docs/ai/task-template.md", ".github/pull_request_template.md"]) {
    assert.ok(files.includes(path), `Thiếu nguồn quy trình: ${path}`)
  }
})

test("kiểm kê chứa đủ config lint/format và test nguồn frontend", async () => {
  const files = await requiredFiles()
  for (const path of ["web/eslint.config.mjs", "web/.prettierrc.json", "web/.prettierignore", "web/src/demo/workspace.test.ts", "web/src/lib/utils.test.ts", "web/src/models/workspace.ts"]) {
    assert.ok(files.includes(path), `Thiếu nguồn chất lượng FE: ${path}`)
  }
})

test("kiểm kê chứa entrypoint, module, test Go và SQL backend", async () => {
  const files = await requiredFiles()
  for (const path of ["go.mod", "go.sum", "Makefile", "cmd/server/main.go", "cmd/migrate/main_test.go",
    "cmd/seed/main.go", "internal/config/config_test.go", "internal/transport/http/router_test.go",
    "internal/repository/postgres/db.go", "migrations/embed.go", "migrations/000001_init.up.sql",
    "migrations/000001_init.down.sql", "db/seed/embed.go", "db/seed/01_reference.sql", "db/seed/02_employee_skills.sql"]) {
    assert.ok(files.includes(path), `Thiếu nguồn backend: ${path}`)
  }
  assert.ok(!files.some((path) => /(^|\/)\.env($|\.)/.test(path)), "Không đưa bí mật vào kiểm kê")
})

test("kiểm kê giữ contract auth và lệnh kiểm tĩnh, không gọi API hoặc DB", async () => {
  const files = await requiredFiles()
  for (const path of ["docs/api/auth.openapi.json", "scripts/check-auth-contract.test.mjs"]) {
    assert.ok(files.includes(path), `Thiếu contract auth: ${path}`)
  }
})

test("kiểm kê giữ spec nhúng, Swagger local và nguồn vendor", async () => {
  const files = await requiredFiles()
  for (const path of ["docs/api/embed.go", "docs/api/README.md", "scripts/check-swagger.test.mjs",
    "internal/transport/http/swagger/init.js", "internal/transport/http/swagger/index.html",
    "internal/transport/http/swagger/swagger-ui-bundle.js", "internal/transport/http/swagger/swagger-ui.css",
    "internal/transport/http/swagger/vendor.json", "internal/transport/http/swagger/LICENSE", "internal/transport/http/swagger/NOTICE"]) {
    assert.ok(files.includes(path), `Thiếu nguồn Swagger: ${path}`)
  }
})

test("kiểm kê giữ công cụ Jira và test offline, không chạy API thật", async () => {
  const files = await requiredFiles()
  for (const path of ["scripts/jira/auth_import.py", "scripts/jira/test_auth_import.py", "scripts/jira/README.md"]) {
    assert.ok(files.includes(path), `Thiếu nguồn công cụ Jira: ${path}`)
  }
})

test("Git ignore giữ binary gốc nhưng không bỏ entrypoint server", async (context) => {
  // Áp dụng chính ignore của repo lên fixture riêng, kể cả khi file chưa được stage.
  const base = await mkdtemp(join(tmpdir(), "taskassign-ignore-test-"))
  context.after(() => rm(base, { recursive: true, force: true }))
  const git = (...args) => execFileSync("git", args, {
    cwd: base, encoding: "utf8", stdio: ["ignore", "pipe", "pipe"],
  })
  git("init", "--quiet")
  await writeFile(join(base, ".gitignore"), await readFile(new URL("../.gitignore", import.meta.url)))
  for (const path of ["server", "ai-staffing", "bin/server"]) {
    assert.equal(git("check-ignore", "--no-index", "--", path).trim(), path)
  }
  assert.throws(() => git("check-ignore", "--no-index", "--", "cmd/server/main.go"),
    (error) => error.status === 1)
})

test("link nội bộ: bỏ fragment/query, URL ngoài và giải mã tên file", () => {
  assert.deepEqual(localLinks("[a](a.md#section) [b](https://example.com) [c](#same) [d](<a%20b.md>) [e](b.md?v=1)"),
    ["a.md", "a b.md", "b.md"])
})

test("không kiểm tra link minh họa trong fenced code block", () => {
  assert.deepEqual(localLinks("```md\n[x](missing.md)\n```\n[x](ok.md)"), ["ok.md"])
})

test("đường dẫn tương đối được giải theo file chứa link", () => {
  assert.equal(localTarget("/repo", "docs/README.md", "../web/README.md"), "web/README.md")
})

test("từ chối đường dẫn vượt ra ngoài repo", () => {
  assert.throws(() => localTarget("/repo", "docs/README.md", "../../outside.md"), /ngoài repo/)
})

test("từ chối đường dẫn tuyệt đối", () => {
  assert.throws(() => localTarget("/repo", "AGENTS.md", "/etc/passwd"), /ngoài repo/)
})

test("Git: phân biệt untracked, đã stage và sửa tiếp chưa stage", async (context) => {
  // Repo fixture riêng trong thư mục tạm; không thay index, cấu hình hoặc commit repo thật.
  const base = await mkdtemp(join(tmpdir(), "taskassign-harness-test-"))
  context.after(() => rm(base, { recursive: true, force: true }))
  const git = (...args) => execFileSync("git", args, {
    cwd: base, encoding: "utf8", stdio: ["ignore", "pipe", "pipe"],
  })
  git("init", "--quiet")
  await writeFile(join(base, "sample.md"), "# Mẫu\n")
  assert.throws(() => checkGitIndex(["sample.md"], base), /Chưa được Git theo dõi/)
  git("add", "--", "sample.md")
  assert.doesNotThrow(() => checkGitIndex(["sample.md"], base))
  await writeFile(join(base, "sample.md"), "# Đã sửa\n")
  assert.throws(() => checkGitIndex(["sample.md"], base), /Nội dung mới chưa được stage/)
})


test("kiểm kê giữ runner PostgreSQL test riêng và không đưa credential local vào nguồn", async () => {
  const files = await requiredFiles()
  for (const path of ["scripts/dev.py", "scripts/test_dev.py", "scripts/authdb/manage.py", "scripts/authdb/test_manage.py", "scripts/authdb/probe.go", "scripts/authdb/README.md"]) {
    assert.ok(files.includes(path), `Thiếu công cụ DB test: ${path}`)
  }
  assert.ok(!files.some((path) => path.startsWith(".local/authdb/")))
})
