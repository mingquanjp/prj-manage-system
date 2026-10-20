import assert from "node:assert/strict"
import { test } from "node:test"
import { execFileSync } from "node:child_process"
import { mkdtemp, rm, writeFile } from "node:fs/promises"
import { tmpdir } from "node:os"
import { join } from "node:path"
import { checkGitIndex, localLinks, localTarget } from "./check-harness.mjs"

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
