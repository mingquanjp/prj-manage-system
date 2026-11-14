// Giữ mặc định prototype cho cách chạy local hiện có; cấu hình lạ không rơi về demo.
export function authMode(value: string | undefined): "demo" | "server" | "invalid" {
  if (value === undefined || value === "demo") return "demo"
  return value === "server" ? "server" : "invalid"
}
