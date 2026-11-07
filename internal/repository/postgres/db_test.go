package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestNewPoolSafeFailuresAndCleanup(t *testing.T) {
	for _, stage := range []string{"parse", "create", "ping", "success"} {
		t.Run(stage, func(t *testing.T) {
			raw := errors.New("postgres://SECRET_MARKER:password@fake/db cookie=COOKIE_MARKER")
			pool := new(pgxpool.Pool)
			cfg := &pgxpool.Config{ConnConfig: &pgx.ConnConfig{Config: pgconn.Config{}}}
			creates, pings, closes := 0, 0, 0
			deps := poolDependencies{
				parse: func(string) (*pgxpool.Config, error) {
					if stage == "parse" {
						return nil, raw
					}
					return cfg, nil
				},
				create: func(_ context.Context, got *pgxpool.Config) (*pgxpool.Pool, error) {
					creates++
					if got != cfg || got.ConnConfig.DefaultQueryExecMode != pgx.QueryExecModeSimpleProtocol || got.MaxConns != 10 || got.MinConns != 1 || got.MaxConnIdleTime != 2*time.Minute || got.MaxConnLifetime != 30*time.Minute {
						t.Fatal("cấu hình pool hồi quy")
					}
					if stage == "create" {
						return nil, raw
					}
					return pool, nil
				},
				ping: func(ctx context.Context, got *pgxpool.Pool) error {
					pings++
					if got != pool {
						t.Fatal("sai pool")
					}
					deadline, ok := ctx.Deadline()
					if !ok || time.Until(deadline) > 15*time.Second {
						t.Fatal("mất timeout")
					}
					if stage == "ping" {
						return raw
					}
					return nil
				},
				close: func(got *pgxpool.Pool) {
					closes++
					if got != pool {
						t.Fatal("sai pool cleanup")
					}
				},
			}
			got, err := newPool(context.Background(), "fake", true, deps)
			want := map[string]error{"parse": ErrDatabaseURLInvalid, "create": ErrPoolInitFailed, "ping": ErrConnectFailed}[stage]
			if !errors.Is(err, want) {
				t.Fatalf("sai nhóm lỗi: %v", err)
			}
			if stage == "success" {
				if got != pool || closes != 0 {
					t.Fatal("thành công không được đóng pool")
				}
			} else {
				if got != nil {
					t.Fatal("không trả pool lỗi")
				}
				if errors.Unwrap(err) != nil {
					t.Fatal("không giữ raw driver error trong chain")
				}
				for _, format := range []string{"%s", "%v", "%+v", "%#v"} {
					message := fmt.Sprintf(format, err)
					if strings.Contains(message, "SECRET_MARKER") || strings.Contains(message, "COOKIE_MARKER") {
						t.Fatal("lộ lỗi driver")
					}
				}
			}
			if stage == "parse" && (creates != 0 || pings != 0) || stage == "create" && pings != 0 || stage == "ping" && closes != 1 {
				t.Fatal("sai thứ tự/cleanup")
			}
		})
	}
}

func TestNewPoolKeepsProtocolForDirectEndpoint(t *testing.T) {
	cfg := &pgxpool.Config{ConnConfig: &pgx.ConnConfig{DefaultQueryExecMode: pgx.QueryExecModeCacheDescribe}}
	pool := new(pgxpool.Pool)
	got, err := newPool(context.Background(), "fake", false, poolDependencies{
		parse: func(string) (*pgxpool.Config, error) { return cfg, nil },
		create: func(_ context.Context, c *pgxpool.Config) (*pgxpool.Pool, error) {
			if c.ConnConfig.DefaultQueryExecMode != pgx.QueryExecModeCacheDescribe {
				t.Fatal("không được ghi đè protocol khi không yêu cầu simple")
			}
			return pool, nil
		},
		ping:  func(context.Context, *pgxpool.Pool) error { return nil },
		close: func(*pgxpool.Pool) { t.Fatal("không đóng pool thành công") },
	})
	if err != nil || got != pool {
		t.Fatalf("không bàn giao được pool: %v", err)
	}
}
