import { defineConfig } from "vite"
import react from "@vitejs/plugin-react"
import tailwindcss from "@tailwindcss/vite"
import { fileURLToPath, URL } from "node:url"

export default defineConfig({
  server: {
    host: "127.0.0.1",
    port: 5174,
    strictPort: true,
    proxy: {
      // Giữ Origin của trình duyệt; backend phải cấu hình APP_ORIGIN khớp FE.
      "/api": {
        target: process.env.TASKASSIGN_API_TARGET || "http://127.0.0.1:8080",
        changeOrigin: false,
      },
    },
  },
  plugins: [react(), tailwindcss()],
  resolve: { alias: { "@": fileURLToPath(new URL("./src", import.meta.url)) } },
})
