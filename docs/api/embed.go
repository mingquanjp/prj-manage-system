// Package api nhúng contract chuẩn để server không cần bản sao khi triển khai.
package api

import "embed"

// Files chứa contract canonical; transport chỉ công bố các operation đã triển khai.
//
//go:embed auth.openapi.json organization-accounts-profile.openapi.json
var Files embed.FS
