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

	"ai-staffing/internal/config"
	"ai-staffing/internal/repository/postgres"
	httptransport "ai-staffing/internal/transport/http"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	if err := run(); err != nil {
		logger.Error("server dừng do lỗi", "err", err)
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

	slog.Info("đang kết nối cơ sở dữ liệu", "pooled", cfg.IsPooledEndpoint())

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL, cfg.IsPooledEndpoint())
	if err != nil {
		return err
	}
	defer pool.Close()

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:           httptransport.NewRouter(pool),
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
