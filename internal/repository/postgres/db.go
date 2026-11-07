// Package postgres chứa toàn bộ phần truy cập dữ liệu. Tầng trên làm việc với interface repository.
package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrDatabaseURLInvalid = errors.New("DB_URL_INVALID: không phân tích được cấu hình PostgreSQL")
	ErrPoolInitFailed     = errors.New("DB_POOL_INIT_FAILED: không khởi tạo được pool PostgreSQL")
	ErrConnectFailed      = errors.New("DB_CONNECT_FAILED: không kết nối được PostgreSQL")
)

// NewPool nhận URL đã được loader kiểm tra. Lỗi trả về không wrap lỗi driver chứa bí mật.
// Giữ simple protocol cho endpoint pooled hoặc file seed nhiều câu lệnh.
func NewPool(ctx context.Context, databaseURL string, simpleProtocol bool) (*pgxpool.Pool, error) {
	return newPool(ctx, databaseURL, simpleProtocol, poolDependencies{
		parse: pgxpool.ParseConfig, create: pgxpool.NewWithConfig,
		ping:  func(ctx context.Context, p *pgxpool.Pool) error { return p.Ping(ctx) },
		close: func(p *pgxpool.Pool) { p.Close() },
	})
}

// Truyền các bước driver để test lỗi/cleanup không cần DB hoặc mở socket.
type poolDependencies struct {
	parse  func(string) (*pgxpool.Config, error)
	create func(context.Context, *pgxpool.Config) (*pgxpool.Pool, error)
	ping   func(context.Context, *pgxpool.Pool) error
	close  func(*pgxpool.Pool)
}

func newPool(ctx context.Context, databaseURL string, simpleProtocol bool, deps poolDependencies) (*pgxpool.Pool, error) {
	cfg, err := deps.parse(databaseURL)
	if err != nil {
		return nil, ErrDatabaseURLInvalid
	}
	if simpleProtocol {
		cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	}
	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.MaxConnIdleTime = 2 * time.Minute
	cfg.MaxConnLifetime = 30 * time.Minute
	pool, err := deps.create(ctx, cfg)
	if err != nil {
		return nil, ErrPoolInitFailed
	}
	pingCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if err := deps.ping(pingCtx, pool); err != nil {
		deps.close(pool)
		return nil, ErrConnectFailed
	}
	return pool, nil
}
