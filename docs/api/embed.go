// Package api nhúng contract chuẩn để server không cần bản sao khi triển khai.
package api

import "embed"

// Files chỉ chứa contract auth đã được kiểm tra cùng mã nguồn.
//
//go:embed auth.openapi.json
var Files embed.FS
