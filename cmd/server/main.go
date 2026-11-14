// Command server khởi chạy HTTP API của hệ thống.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ai-staffing/internal/app"
	"ai-staffing/internal/config"
	"ai-staffing/internal/repository/postgres"
	"ai-staffing/internal/service/auth"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	if err := run(); err != nil {
		logStartupError(logger, err)
		os.Exit(1)
	}
}

func run() error {
	return runWithStartup(startServer)
}

// runWithStartup giữ cổng cấu hình trước mọi thao tác kết nối; test thay bước chạy.
func runWithStartup(start func(*config.Config) error) error {
	cfg, err := config.LoadServer()
	if err != nil {
		return err
	}

	return start(cfg)
}

func startServer(cfg *config.Config) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := openStartupPool(ctx, cfg, postgres.NewPool)
	if err != nil {
		return err
	}
	defer pool.Close()
	handler, err := buildHandler(cfg, pool, systemClock{})
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("server đang lắng nghe", "addr", srv.Addr, "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		slog.Info("nhận tín hiệu dừng, đang tắt server")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return srv.Shutdown(shutdownCtx)
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }

// buildHandler dùng chung cho startup và integration HTTP trên pool test thật.
func buildHandler(cfg *config.Config, pool *pgxpool.Pool, clock auth.Clock) (http.Handler, error) {
	return app.AuthHandler(cfg, pool, clock)
}

// logStartupError chỉ nhận lỗi cấu hình/kết nối đã được các cổng startup làm an toàn.
func logStartupError(logger *slog.Logger, err error) {
	logger.Error("server dừng do lỗi", "err", err)
}

func openStartupPool(ctx context.Context, cfg *config.Config, open func(context.Context, string, bool) (*pgxpool.Pool, error)) (*pgxpool.Pool, error) {
	slog.Info("đang kết nối cơ sở dữ liệu", "pooled", cfg.Database.Pooled)
	pool, err := open(ctx, cfg.Database.RuntimeURL, cfg.Database.Pooled)
	if err == nil {
		return pool, nil
	}
	// Chỉ giữ nhóm lỗi công khai; không chuyển tiếp lỗi giả/driver thô vào log.
	for _, safe := range []error{postgres.ErrDatabaseURLInvalid, postgres.ErrPoolInitFailed, postgres.ErrConnectFailed} {
		if errors.Is(err, safe) {
			return nil, safe
		}
	}
	return nil, postgres.ErrConnectFailed
}
