// Package config cung cấp cấu trúc, loader và validation cấu hình.
// Kết nối PostgreSQL được cung cấp qua DATABASE_URL, không gắn với một nhà cung cấp.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv     string // dev | prod
	HTTPPort   int
	Database   DatabaseEndpoint
	AppOrigin  string
	SessionTTL time.Duration
}

// RequireDev yêu cầu chọn môi trường dev tường minh trước thao tác phá dữ liệu.
// Không dùng giá trị mặc định hoặc nạp .env để vô tình cho phép seed/reset.
func RequireDev(operation string) error {
	if os.Getenv("APP_ENV") != "dev" {
		return fmt.Errorf("%s chỉ được phép khi đặt APP_ENV=dev tường minh", operation)
	}
	return nil
}

// Load nạp cấu hình cho migrate/seed, không yêu cầu cấu hình HTTP/auth.
func Load() (*Config, error) {
	return load(os.LookupEnv, os.ReadFile, ".env", false)
}

// LoadServer kiểm tra toàn bộ cấu hình server trước khi mở kết nối DB.
func LoadServer() (*Config, error) {
	return load(os.LookupEnv, os.ReadFile, ".env", true)
}

// load nhận nguồn đọc để test bằng file tạm, không đọc .env thật hoặc sửa OS env.
func load(lookup func(string) (string, bool), readFile func(string) ([]byte, error), path string, server bool) (*Config, error) {
	appEnv, _ := lookup("APP_ENV")
	if appEnv != "dev" && appEnv != "prod" {
		return nil, fmt.Errorf("APP_ENV phải được đặt dev hoặc prod trong môi trường chạy")
	}
	fileValues := map[string]string{}
	if appEnv == "dev" {
		var err error
		fileValues, err = readDotEnv(readFile, path)
		if err != nil {
			return nil, err
		}
	}
	value := func(key, fallback string) string {
		if v, ok := lookup(key); ok {
			return v
		}
		if v, ok := fileValues[key]; ok {
			return v
		}
		return fallback
	}
	cfg := &Config{AppEnv: appEnv}
	if server {
		port, err := strconv.Atoi(value("HTTP_PORT", "8080"))
		if err != nil {
			return nil, fmt.Errorf("HTTP_PORT phải là số nguyên từ 1 đến 65535")
		}
		cfg.HTTPPort = port
		cfg.AppOrigin = value("APP_ORIGIN", "")
		ttl, err := time.ParseDuration(value("SESSION_TTL", "8h"))
		if err != nil {
			return nil, fmt.Errorf("SESSION_TTL phải là thời lượng hợp lệ bằng 8h")
		}
		cfg.SessionTTL = ttl
		if err := ValidateServer(*cfg); err != nil {
			return nil, err
		}
	}
	endpoint, err := ParseDatabaseURL(value("DATABASE_URL", ""))
	if err != nil {
		return nil, err
	}
	cfg.Database = endpoint
	return cfg, nil
}

// readDotEnv hỗ trợ KEY=VALUE, giá trị bọc nháy đơn/đôi và comment cả dòng.
// Không nội suy biến, escape, export, comment cuối dòng hoặc giá trị nhiều dòng.
func readDotEnv(readFile func(string) ([]byte, error), path string) (map[string]string, error) {
	values := map[string]string{}
	data, err := readFile(path)
	if os.IsNotExist(err) {
		return values, nil
	}
	if err != nil {
		return nil, fmt.Errorf("không đọc được file dotenv")
	}
	for index, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		invalid := !found || key == ""
		for i, c := range key {
			if !(c == '_' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || i > 0 && c >= '0' && c <= '9') {
				invalid = true
			}
		}
		if strings.HasPrefix(value, "\"") || strings.HasPrefix(value, "'") {
			if len(value) < 2 || value[len(value)-1] != value[0] {
				invalid = true
			} else {
				quote := value[0]
				value = value[1 : len(value)-1]
				if strings.ContainsRune(value, rune(quote)) {
					invalid = true
				}
			}
		} else if strings.ContainsAny(value, " \t\"'") {
			invalid = true
		}
		if invalid {
			return nil, fmt.Errorf("file dotenv sai định dạng tại dòng %d", index+1)
		}
		if _, exists := values[key]; exists {
			return nil, fmt.Errorf("file dotenv trùng khóa tại dòng %d", index+1)
		}
		values[key] = value
	}
	return values, nil
}
