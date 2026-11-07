package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadSources(t *testing.T) {
	tests := []struct {
		name      string
		env       map[string]string
		file      string
		server    bool
		wantField string
		wantRead  bool
		check     func(*testing.T, *Config)
	}{
		{name: "dev defaults", env: map[string]string{"APP_ENV": "dev", "DATABASE_URL": "fake", "APP_ORIGIN": "http://localhost:5173"}, server: true, wantRead: true, check: func(t *testing.T, c *Config) {
			if c.HTTPPort != 8080 || c.SessionTTL != 8*time.Hour {
				t.Fatalf("defaults: %+v", c)
			}
		}},
		{name: "prod skips malformed file", env: map[string]string{"APP_ENV": "prod", "DATABASE_URL": "fake", "APP_ORIGIN": "https://example.test"}, file: "INVALID SECRET_MARKER", server: true},
		{name: "env wins", env: map[string]string{"APP_ENV": "dev", "DATABASE_URL": "env-db", "HTTP_PORT": "9000", "APP_ORIGIN": "http://127.0.0.1:5173", "SESSION_TTL": "480m"}, file: "APP_ENV=prod\nDATABASE_URL=file-db\nHTTP_PORT=8081\nAPP_ORIGIN=https://file.test\nSESSION_TTL=1h", server: true, wantRead: true, check: func(t *testing.T, c *Config) {
			if c.AppEnv != "dev" || c.DatabaseURL != "env-db" || c.HTTPPort != 9000 || c.AppOrigin != "http://127.0.0.1:5173" || c.SessionTTL != 8*time.Hour {
				t.Fatalf("precedence: %+v", c)
			}
		}},
		{name: "file fallback", env: map[string]string{"APP_ENV": "dev"}, file: "# comment\nDATABASE_URL='fake'\nHTTP_PORT=8081\nAPP_ORIGIN=\"http://localhost:5173\"\nSESSION_TTL=8h\n", server: true, wantRead: true, check: func(t *testing.T, c *Config) {
			if c.DatabaseURL != "fake" || c.HTTPPort != 8081 {
				t.Fatal("không đọc giá trị file")
			}
		}},
		{name: "file cannot select dev", env: map[string]string{}, file: "APP_ENV=dev", server: true, wantField: "APP_ENV"},
		{name: "empty env", env: map[string]string{"APP_ENV": ""}, server: true, wantField: "APP_ENV"},
		{name: "invalid env", env: map[string]string{"APP_ENV": "test"}, server: true, wantField: "APP_ENV"},
		{name: "missing prod database", env: map[string]string{"APP_ENV": "prod", "APP_ORIGIN": "https://example.test"}, server: true, wantField: "DATABASE_URL"},
		{name: "empty database overrides file", env: map[string]string{"APP_ENV": "dev", "DATABASE_URL": "", "APP_ORIGIN": "http://localhost"}, file: "DATABASE_URL=fake", server: true, wantRead: true, wantField: "DATABASE_URL"},
		{name: "blank database", env: map[string]string{"APP_ENV": "prod", "DATABASE_URL": "  ", "APP_ORIGIN": "https://example.test"}, server: true, wantField: "DATABASE_URL"},
		{name: "missing prod origin", env: map[string]string{"APP_ENV": "prod", "DATABASE_URL": "fake"}, server: true, wantField: "APP_ORIGIN"},
		{name: "empty origin overrides file", env: map[string]string{"APP_ENV": "dev", "DATABASE_URL": "fake", "APP_ORIGIN": ""}, file: "APP_ORIGIN=http://localhost", server: true, wantRead: true, wantField: "APP_ORIGIN"},
		{name: "invalid origin", env: map[string]string{"APP_ENV": "prod", "DATABASE_URL": "fake", "APP_ORIGIN": "http://SECRET_MARKER.test"}, server: true, wantField: "APP_ORIGIN"},
		{name: "empty port overrides file", env: map[string]string{"APP_ENV": "dev", "HTTP_PORT": ""}, file: "HTTP_PORT=8080", server: true, wantRead: true, wantField: "HTTP_PORT"},
		{name: "invalid port", env: map[string]string{"APP_ENV": "prod", "HTTP_PORT": "SECRET_MARKER"}, server: true, wantField: "HTTP_PORT"},
		{name: "out of range port", env: map[string]string{"APP_ENV": "prod", "HTTP_PORT": "65536", "APP_ORIGIN": "https://example.test"}, server: true, wantField: "HTTP_PORT"},
		{name: "empty ttl overrides file", env: map[string]string{"APP_ENV": "dev", "SESSION_TTL": ""}, file: "SESSION_TTL=8h", server: true, wantRead: true, wantField: "SESSION_TTL"},
		{name: "invalid ttl", env: map[string]string{"APP_ENV": "prod", "SESSION_TTL": "SECRET_MARKER"}, server: true, wantField: "SESSION_TTL"},
		{name: "wrong ttl", env: map[string]string{"APP_ENV": "prod", "DATABASE_URL": "fake", "APP_ORIGIN": "https://example.test", "SESSION_TTL": "1h"}, server: true, wantField: "SESSION_TTL"},
		{name: "CLI ignores HTTP auth inputs", env: map[string]string{"APP_ENV": "prod", "DATABASE_URL": "fake", "HTTP_PORT": "SECRET_MARKER", "APP_ORIGIN": "bad", "SESSION_TTL": "bad"}, check: func(t *testing.T, c *Config) {
			if c.HTTPPort != 0 || c.AppOrigin != "" || c.SessionTTL != 0 {
				t.Fatal("CLI không cần HTTP/auth")
			}
		}},
		{name: "CLI dev file", env: map[string]string{"APP_ENV": "dev"}, file: "DATABASE_URL=fake", wantRead: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "test.env")
			if tt.file != "" {
				if err := os.WriteFile(path, []byte(tt.file), 0600); err != nil {
					t.Fatal(err)
				}
			}
			reads := 0
			lookup := func(k string) (string, bool) { v, ok := tt.env[k]; return v, ok }
			read := func(p string) ([]byte, error) { reads++; return os.ReadFile(p) }
			cfg, err := load(lookup, read, path, tt.server)
			if (reads > 0) != tt.wantRead {
				t.Fatalf("read count=%d", reads)
			}
			if tt.wantField != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantField) {
					t.Fatalf("want %s, got %v", tt.wantField, err)
				}
				if cfg != nil {
					t.Fatal("không trả cấu hình lỗi")
				}
				if strings.Contains(err.Error(), "SECRET_MARKER") {
					t.Fatal("lộ giá trị")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}
}

func TestReadDotEnv(t *testing.T) {
	tests := []struct {
		name, content string
		readErr       error
		wantErr       bool
	}{
		{name: "missing", readErr: os.ErrNotExist},
		{name: "read error redacted", readErr: errors.New("SECRET_MARKER"), wantErr: true},
		{name: "empty"},
		{name: "CRLF and equals", content: "# comment\r\nDATABASE_URL=postgres://fake/db?x=y\r\nEMPTY=\r\nLABEL='two words'"},
		{name: "missing equals", content: "SECRET_MARKER", wantErr: true},
		{name: "empty key", content: "=SECRET_MARKER", wantErr: true},
		{name: "invalid key", content: "1KEY=SECRET_MARKER", wantErr: true},
		{name: "export unsupported", content: "export KEY=SECRET_MARKER", wantErr: true},
		{name: "missing quote", content: "KEY='SECRET_MARKER", wantErr: true},
		{name: "mixed quote", content: "KEY='SECRET_MARKER\"", wantErr: true},
		{name: "internal quote", content: "KEY='a'b'", wantErr: true},
		{name: "unquoted spaces", content: "KEY=SECRET_MARKER other", wantErr: true},
		{name: "duplicate key", content: "KEY=one\nKEY=SECRET_MARKER", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := readDotEnv(func(string) ([]byte, error) { return []byte(tt.content), tt.readErr }, "fake")
			if (err != nil) != tt.wantErr {
				t.Fatalf("error: %v", err)
			}
			if err != nil && strings.Contains(err.Error(), "SECRET_MARKER") {
				t.Fatal("lộ nội dung file/lỗi OS")
			}
		})
	}
}

func TestLoadDoesNotChangeEnvironment(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("APP_ENV", "dev")
	for _, key := range []string{"DATABASE_URL", "APP_ORIGIN", "HTTP_PORT", "SESSION_TTL"} {
		t.Setenv(key, "")
		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(".env", []byte("APP_ENV=prod\nDATABASE_URL=fake\nAPP_ORIGIN=http://localhost:5173"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, loader := range []func() (*Config, error){LoadServer, Load} {
		cfg, err := loader()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.AppEnv != "dev" {
			t.Fatal("dotenv đổi APP_ENV")
		}
	}
	for _, key := range []string{"DATABASE_URL", "APP_ORIGIN", "HTTP_PORT", "SESSION_TTL"} {
		if _, exists := os.LookupEnv(key); exists {
			t.Fatalf("loader đã ghi OS env %s", key)
		}
	}
	if err := RequireDev("test"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("APP_ENV", "")
	if _, err := Load(); err == nil {
		t.Fatal("dotenv không được cấp dev")
	}
	if err := RequireDev("test"); err == nil {
		t.Fatal("guard bị vượt qua")
	}
}

// Kiểm giá trị thực nhận, không chỉ kiểm parser trả nil.
func TestDotEnvPreservesLiteralValues(t *testing.T) {
	content := "DATABASE_URL=postgres://fake/db?x=y#literal\r\nLABEL='two words'\nREFERENCE=$NOT_EXPANDED\nEMPTY=\n"
	values, err := readDotEnv(func(string) ([]byte, error) { return []byte(content), nil }, "fake")
	if err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{
		"DATABASE_URL": "postgres://fake/db?x=y#literal", "LABEL": "two words", "REFERENCE": "$NOT_EXPANDED", "EMPTY": "",
	} {
		got, exists := values[key]
		if !exists || got != want {
			t.Fatalf("giá trị %s không được giữ nguyên", key)
		}
	}
}
