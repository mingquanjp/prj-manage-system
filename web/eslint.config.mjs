import js from "@eslint/js"
import { defineConfig, globalIgnores } from "eslint/config"
import prettier from "eslint-config-prettier/flat"
import reactHooks from "eslint-plugin-react-hooks"
import globals from "globals"
import tseslint from "typescript-eslint"

export default defineConfig([
  globalIgnores(["dist/**", "node_modules/**"]),
  {
    files: ["src/**/*.{ts,tsx}", "vite.config.ts"],
    extends: [js.configs.recommended, tseslint.configs.recommended],
  },
  {
    files: ["src/**/*.{ts,tsx}"],
    ignores: ["src/**/*.test.ts"],
    languageOptions: { globals: globals.browser },
  },
  {
    files: ["src/**/*.{ts,tsx}"],
    plugins: { "react-hooks": reactHooks },
    // Cổng Hooks cơ bản; chưa áp dụng các chẩn đoán React Compiler thành một đợt refactor ngầm.
    rules: {
      "react-hooks/rules-of-hooks": "error",
      "react-hooks/exhaustive-deps": "error",
    },
  },
  {
    files: ["*.mjs", "vite.config.ts", "src/**/*.test.ts"],
    languageOptions: { globals: globals.node },
  },
  { files: ["*.mjs"], extends: [js.configs.recommended] },
  prettier,
])
