// Package config đọc cấu hình từ biến môi trường.
//
// Toàn bộ hệ thống chỉ phụ thuộc vào DATABASE_URL, nhờ vậy chạy được với cả
// Neon (serverless) lẫn PostgreSQL cài local mà không sửa dòng code nào.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppEnv      string // dev | prod
	HTTPPort    int
	DatabaseURL string
}

// Load đọc cấu hình, ưu tiên biến môi trường thật. Nếu chưa có thì thử nạp
// từ file .env ở thư mục gốc (chỉ dùng khi phát triển).
func Load() (*Config, error) {
	loadDotEnv(".env")

	cfg := &Config{
		AppEnv:      envOr("APP_ENV", "dev"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	port, err := strconv.Atoi(envOr("HTTP_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("HTTP_PORT không hợp lệ: %w", err)
	}
	cfg.HTTPPort = port

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("thiếu DATABASE_URL — sao chép .env.example thành .env rồi điền chuỗi kết nối Neon")
	}

	return cfg, nil
}

// IsPooledEndpoint báo hiệu chuỗi kết nối đang trỏ tới endpoint PgBouncer của Neon.
// pgx dùng prepared statement nên không tương thích hoàn toàn với transaction pooling.
func (c *Config) IsPooledEndpoint() bool {
	return strings.Contains(c.DatabaseURL, poolerMarker)
}

// DirectURL trả về chuỗi kết nối tới endpoint trực tiếp.
//
// Neon đặt tên endpoint pooled bằng cách chèn "-pooler" vào hostname của endpoint
// trực tiếp, nên bỏ hậu tố này là ra chuỗi cần dùng. Migration bắt buộc đi đường
// trực tiếp vì golang-migrate cần advisory lock, thứ mà PgBouncer ở chế độ
// transaction pooling không giữ được giữa các câu lệnh.
func (c *Config) DirectURL() string {
	return strings.Replace(c.DatabaseURL, poolerMarker, ".", 1)
}

const poolerMarker = "-pooler."

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// loadDotEnv nạp các cặp KEY=VALUE từ file .env. Cố ý viết tay thay vì kéo thêm
// thư viện: định dạng cần hỗ trợ rất đơn giản.
func loadDotEnv(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)

		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
}
