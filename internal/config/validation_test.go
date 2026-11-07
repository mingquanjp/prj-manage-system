package config

import (
	"strings"
	"testing"
	"time"
)

func TestValidateServer(t *testing.T) {
	tests := []struct {
		name      string
		change    func(*Config)
		wantField string
	}{
		{name: "dev localhost hợp lệ"},
		{name: "prod HTTPS hợp lệ", change: func(c *Config) { c.AppEnv = "prod"; c.AppOrigin = "https://staffing.example" }},
		{name: "dev HTTPS từ xa", change: func(c *Config) { c.AppOrigin = "https://staffing.example:8443" }},
		{name: "dev IPv4 loopback", change: func(c *Config) { c.AppOrigin = "http://127.0.0.1:5173" }},
		{name: "dev IPv6 loopback", change: func(c *Config) { c.AppOrigin = "http://[::1]:5173" }},
		{name: "origin port nhỏ nhất", change: func(c *Config) { c.AppOrigin = "https://staffing.example:1" }},
		{name: "origin port lớn nhất", change: func(c *Config) { c.AppOrigin = "https://staffing.example:65535" }},
		{name: "port nhỏ nhất", change: func(c *Config) { c.HTTPPort = 1 }},
		{name: "port lớn nhất", change: func(c *Config) { c.HTTPPort = 65535 }},
		{name: "env thiếu", change: func(c *Config) { c.AppEnv = "" }, wantField: "APP_ENV"},
		{name: "env không hỗ trợ", change: func(c *Config) { c.AppEnv = "test" }, wantField: "APP_ENV"},
		{name: "env chữ hoa", change: func(c *Config) { c.AppEnv = "DEV" }, wantField: "APP_ENV"},
		{name: "env có khoảng trắng", change: func(c *Config) { c.AppEnv = " dev " }, wantField: "APP_ENV"},
		{name: "port zero", change: func(c *Config) { c.HTTPPort = 0 }, wantField: "HTTP_PORT"},
		{name: "port âm", change: func(c *Config) { c.HTTPPort = -1 }, wantField: "HTTP_PORT"},
		{name: "port vượt biên", change: func(c *Config) { c.HTTPPort = 65536 }, wantField: "HTTP_PORT"},
		{name: "origin thiếu", change: func(c *Config) { c.AppOrigin = "" }, wantField: "APP_ORIGIN"},
		{name: "origin thiếu scheme", change: func(c *Config) { c.AppOrigin = "//staffing.example" }, wantField: "APP_ORIGIN"},
		{name: "origin thiếu host", change: func(c *Config) { c.AppOrigin = "https://" }, wantField: "APP_ORIGIN"},
		{name: "origin scheme khác", change: func(c *Config) { c.AppOrigin = "ftp://staffing.example" }, wantField: "APP_ORIGIN"},
		{name: "origin opaque", change: func(c *Config) { c.AppOrigin = "https:staffing.example" }, wantField: "APP_ORIGIN"},
		{name: "origin có userinfo", change: func(c *Config) { c.AppOrigin = "https://user:SECRET_MARKER@staffing.example" }, wantField: "APP_ORIGIN"},
		{name: "origin có path", change: func(c *Config) { c.AppOrigin = "https://staffing.example/SECRET_MARKER" }, wantField: "APP_ORIGIN"},
		{name: "origin có slash cuối", change: func(c *Config) { c.AppOrigin = "https://staffing.example/" }, wantField: "APP_ORIGIN"},
		{name: "origin có query", change: func(c *Config) { c.AppOrigin = "https://staffing.example?key=SECRET_MARKER" }, wantField: "APP_ORIGIN"},
		{name: "origin có query rỗng", change: func(c *Config) { c.AppOrigin = "https://staffing.example?" }, wantField: "APP_ORIGIN"},
		{name: "origin có fragment", change: func(c *Config) { c.AppOrigin = "https://staffing.example#SECRET_MARKER" }, wantField: "APP_ORIGIN"},
		{name: "origin có fragment rỗng", change: func(c *Config) { c.AppOrigin = "https://staffing.example#" }, wantField: "APP_ORIGIN"},
		{name: "origin lỗi parse", change: func(c *Config) { c.AppOrigin = "https://staffing.example/%SECRET_MARKER" }, wantField: "APP_ORIGIN"},
		{name: "origin khoảng trắng", change: func(c *Config) { c.AppOrigin = " https://staffing.example " }, wantField: "APP_ORIGIN"},
		{name: "origin port không phải số", change: func(c *Config) { c.AppOrigin = "https://staffing.example:SECRET_MARKER" }, wantField: "APP_ORIGIN"},
		{name: "origin port rỗng", change: func(c *Config) { c.AppOrigin = "https://staffing.example:" }, wantField: "APP_ORIGIN"},
		{name: "origin port zero", change: func(c *Config) { c.AppOrigin = "https://staffing.example:0" }, wantField: "APP_ORIGIN"},
		{name: "origin port vượt biên", change: func(c *Config) { c.AppOrigin = "https://staffing.example:65536" }, wantField: "APP_ORIGIN"},
		{name: "origin IPv6 không hợp lệ", change: func(c *Config) { c.AppOrigin = "https://[not-an-ip]" }, wantField: "APP_ORIGIN"},
		{name: "origin IPv6 thiếu ngoặc", change: func(c *Config) { c.AppOrigin = "http://::1" }, wantField: "APP_ORIGIN"},
		{name: "origin IPv6 có zone", change: func(c *Config) { c.AppOrigin = "http://[::1%25lo0]" }, wantField: "APP_ORIGIN"},
		{name: "prod HTTP từ xa", change: func(c *Config) { c.AppEnv = "prod"; c.AppOrigin = "http://staffing.example" }, wantField: "APP_ORIGIN"},
		{name: "prod HTTP localhost", change: func(c *Config) { c.AppEnv = "prod" }, wantField: "APP_ORIGIN"},
		{name: "dev HTTP từ xa", change: func(c *Config) { c.AppOrigin = "http://staffing.example" }, wantField: "APP_ORIGIN"},
		{name: "dev HTTP mạng nội bộ", change: func(c *Config) { c.AppOrigin = "http://192.168.1.10" }, wantField: "APP_ORIGIN"},
		{name: "dev tên giả localhost", change: func(c *Config) { c.AppOrigin = "http://localhost.attacker.example" }, wantField: "APP_ORIGIN"},
		{name: "TTL thiếu", change: func(c *Config) { c.SessionTTL = 0 }, wantField: "SESSION_TTL"},
		{name: "TTL âm", change: func(c *Config) { c.SessionTTL = -time.Hour }, wantField: "SESSION_TTL"},
		{name: "TTL ngắn hơn", change: func(c *Config) { c.SessionTTL = 8*time.Hour - time.Nanosecond }, wantField: "SESSION_TTL"},
		{name: "TTL dài hơn", change: func(c *Config) { c.SessionTTL = 8*time.Hour + time.Nanosecond }, wantField: "SESSION_TTL"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := Config{AppEnv: "dev", HTTPPort: 8080, AppOrigin: "http://localhost:5173", SessionTTL: 8 * time.Hour}
			if tt.change != nil {
				tt.change(&cfg)
			}
			before := cfg
			err := ValidateServer(cfg)
			if cfg != before {
				t.Fatal("validation đã thay đổi input")
			}
			if tt.wantField == "" {
				if err != nil {
					t.Fatalf("cấu hình hợp lệ bị từ chối: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantField) {
				t.Fatalf("cần lỗi của %s, nhận: %v", tt.wantField, err)
			}
			if strings.Contains(err.Error(), "SECRET_MARKER") || (cfg.AppOrigin != "" && strings.Contains(err.Error(), cfg.AppOrigin)) {
				t.Fatal("lỗi validation lộ giá trị cấu hình")
			}
		})
	}
}
