// Package postgres chứa toàn bộ phần truy cập dữ liệu. Tầng trên chỉ làm việc
// với interface repository, không biết tới pgx.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool tạo connection pool tới PostgreSQL.
//
// simpleProtocol tắt prepared statement. Cần bật trong hai trường hợp:
//   - Chuỗi kết nối trỏ tới endpoint PgBouncer của Neon: transaction pooling tái sử
//     dụng connection giữa các transaction nên prepared statement sinh lỗi rải rác
//     "prepared statement already exists".
//   - Cần chạy file SQL chứa nhiều câu lệnh trong một lần Exec, ví dụ khi nạp dữ
//     liệu mẫu. Extended protocol chỉ cho phép một câu lệnh mỗi lần gọi.
func NewPool(ctx context.Context, databaseURL string, simpleProtocol bool) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("chuỗi kết nối không hợp lệ: %w", err)
	}

	if simpleProtocol {
		cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	}

	cfg.MaxConns = 10
	cfg.MinConns = 1
	// Neon tự ngủ khi không dùng; giữ connection quá lâu chỉ tổ nhận lỗi khi tỉnh lại.
	cfg.MaxConnIdleTime = 2 * time.Minute
	cfg.MaxConnLifetime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("không tạo được pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("không kết nối được tới cơ sở dữ liệu: %w", err)
	}

	return pool, nil
}
