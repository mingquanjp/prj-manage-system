// Package migrations nhúng toàn bộ file .sql vào binary để không phải cài
// công cụ migrate riêng — chỉ cần `go run ./cmd/migrate`.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
