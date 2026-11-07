package main

import (
	"ai-staffing/internal/repository/postgres"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"ai-staffing/internal/config"
)

func TestStartupConfigGate(t *testing.T) {
	t.Chdir(t.TempDir())
	// File tạm cố ý sai: prod phải bỏ qua, không đọc .env của repo.
	if err := os.WriteFile(".env", []byte("SECRET_MARKER"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("APP_ENV", "prod")
	t.Setenv("DATABASE_URL", "postgres://SECRET_MARKER@fake.invalid/fake")
	t.Setenv("HTTP_PORT", "8080")
	t.Setenv("APP_ORIGIN", "https://example.test")
	t.Setenv("SESSION_TTL", "8h")
	sentinel := errors.New("đã tới bước chạy thay thế")
	calls := 0
	start := func(cfg *config.Config) error {
		calls++
		if cfg.AppEnv != "prod" || cfg.HTTPPort != 8080 || cfg.SessionTTL != 8*time.Hour {
			t.Fatal("cấu hình bàn giao sai")
		}
		return sentinel
	}
	if err := runWithStartup(start); !errors.Is(err, sentinel) || calls != 1 {
		t.Fatalf("startup hợp lệ: %v / %d", err, calls)
	}
	for _, tt := range []struct{ key, value string }{
		{"APP_ENV", ""}, {"HTTP_PORT", "SECRET_MARKER"}, {"APP_ORIGIN", "http://SECRET_MARKER.invalid"}, {"SESSION_TTL", "1h"}, {"DATABASE_URL", ""},
	} {
		t.Run(tt.key, func(t *testing.T) {
			t.Setenv(tt.key, tt.value)
			err := runWithStartup(start)
			if err == nil || !strings.Contains(err.Error(), tt.key) || strings.Contains(err.Error(), "SECRET_MARKER") {
				t.Fatalf("startup lỗi: %v", err)
			}
			if calls != 1 {
				t.Fatal("cấu hình sai vẫn tới bước kết nối")
			}
		})
	}
}

// Kiểm đường run thực sự nối loader, không mở DB khi APP_ENV chưa hợp lệ.
func TestRunRejectsConfigBeforeDatabase(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("APP_ENV", "")
	t.Setenv("DATABASE_URL", "postgres://SECRET_MARKER@fake.invalid/fake")
	if err := run(); err == nil || !strings.Contains(err.Error(), "APP_ENV") || strings.Contains(err.Error(), "SECRET_MARKER") {
		t.Fatalf("run phải dừng tại cấu hình: %v", err)
	}
}

func TestStartupDotEnvFailureGate(t *testing.T) {
	for _, kind := range []string{"malformed", "unreadable"} {
		t.Run(kind, func(t *testing.T) {
			t.Chdir(t.TempDir())
			t.Setenv("APP_ENV", "dev")
			t.Setenv("DATABASE_URL", "postgres://example.test/demo")
			t.Setenv("HTTP_PORT", "8080")
			t.Setenv("APP_ORIGIN", "http://localhost:5173")
			t.Setenv("SESSION_TTL", "8h")
			if kind == "malformed" {
				if err := os.WriteFile(".env", []byte("SECRET_MARKER"), 0600); err != nil {
					t.Fatal(err)
				}
			} else {
				// Dùng thư mục thay file: lỗi đọc thực, không phụ thuộc UID/quyền chmod.
				if err := os.Mkdir(".env", 0700); err != nil {
					t.Fatal(err)
				}
			}
			calls := 0
			start := func(*config.Config) error { calls++; return nil }
			err := runWithStartup(start)
			if err == nil || !strings.Contains(err.Error(), "dotenv") || strings.Contains(err.Error(), "SECRET_MARKER") || calls != 0 {
				t.Fatalf("dev phải dừng trước bước chạy: %v / %d", err, calls)
			}
			t.Setenv("APP_ENV", "prod")
			t.Setenv("APP_ORIGIN", "https://example.test")
			if err := runWithStartup(start); err != nil || calls != 1 {
				t.Fatalf("prod không được đọc dotenv: %v / %d", err, calls)
			}
		})
	}
}

func TestStartupConnectionLogsAreSafe(t *testing.T) {
	for _, cause := range []error{
		errors.New("postgres://SECRET_MARKER:password@fake/db cookie=COOKIE_MARKER"),
		fmt.Errorf("SECRET_MARKER: %w", postgres.ErrPoolInitFailed),
		postgres.ErrDatabaseURLInvalid, postgres.ErrConnectFailed,
	} {
		t.Run(fmt.Sprintf("%T", cause), func(t *testing.T) {
			var output bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&output, nil))
			previous := slog.Default()
			slog.SetDefault(logger)
			t.Cleanup(func() { slog.SetDefault(previous) })
			cfg := &config.Config{Database: config.DatabaseEndpoint{RuntimeURL: "postgres://SECRET_MARKER@fake/db", Pooled: true}}
			_, err := openStartupPool(context.Background(), cfg, func(_ context.Context, raw string, pooled bool) (*pgxpool.Pool, error) {
				if raw != cfg.Database.RuntimeURL || !pooled {
					t.Fatal("sai endpoint runtime/protocol")
				}
				return nil, cause
			})
			if err == nil {
				t.Fatal("không được bỏ lỗi kết nối")
			}
			logStartupError(logger, err)
			message := output.String()
			if strings.Contains(message, "SECRET_MARKER") || strings.Contains(message, "COOKIE_MARKER") || !strings.Contains(message, "DB_") || !strings.Contains(message, "server dừng do lỗi") {
				t.Fatal("log lộ bí mật hoặc mất ngữ cảnh")
			}
		})
	}
}
