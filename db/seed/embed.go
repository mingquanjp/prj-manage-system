// Package seed nhúng các file SQL dữ liệu mẫu.
package seed

import "embed"

//go:embed *.sql
var FS embed.FS
