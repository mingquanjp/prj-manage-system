import assert from "node:assert/strict"
import { test } from "node:test"
import { cn } from "./utils.ts"

test("cn bỏ class điều kiện không dùng và gộp class Tailwind xung đột", () => {
  assert.equal(
    cn("px-2", false, ["text-sm"], { "font-semibold": true }, "px-4"),
    "text-sm font-semibold px-4",
  )
})
