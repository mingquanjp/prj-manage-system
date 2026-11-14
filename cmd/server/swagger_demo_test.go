//go:build integration && swaggerdemo

package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"ai-staffing/internal/config"
	"ai-staffing/internal/testutil/authdb"
	"golang.org/x/crypto/bcrypt"
)

// TestSwaggerDemoServer là harness thủ công có opt-in build tag, không là test pass nghiệp vụ.
// Chỉ dùng runner DB sở hữu; dữ liệu giả được dọn khi nhận tín hiệu hoặc hết 20 phút.
func TestSwaggerDemoServer(t *testing.T) {
	runAuthDemoServer(t, "http://127.0.0.1:18080")
}

// TestFrontendDemoServer dùng Origin FE cố định, vẫn chạy cùng adapter và DB test sở hữu.
// Cổng 5174 tránh chiếm tiến trình dự án khác đang chạy tại 5173.
func TestFrontendDemoServer(t *testing.T) {
	runAuthDemoServer(t, "http://127.0.0.1:5174")
}

func runAuthDemoServer(t *testing.T, origin string) {
	t.Helper()
	f := authdb.NewFixture(t)
	// Credential dễ nhập chỉ dành cho demo loopback, không đổi policy mật khẩu mới.
	hash, err := bcrypt.GenerateFromPassword([]byte("12345678"), bcrypt.DefaultCost)
	authdb.Require(t, "tạo hash mật khẩu demo", err)
	f.PasswordHash = string(hash)
	f.AddAccount(t, authdb.AccountOptions{Email: "quan1652005@gmail.com", Roles: []string{"HR"}})
	f.AddAccount(t, authdb.AccountOptions{Email: "quan1652005+limited@gmail.com", Roles: []string{"ADMIN"}, MustChangePassword: true})
	handler, err := buildHandler(&config.Config{AppEnv: "dev", AppOrigin: origin, SessionTTL: 8 * time.Hour}, f.DB.Pool(t), systemClock{})
	authdb.Require(t, "lắp server Swagger test", err)
	listener, err := net.Listen("tcp", "127.0.0.1:18080")
	authdb.Require(t, "mở cổng loopback demo", err)
	srv := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	})
	result := make(chan error, 1)
	go func() { result <- srv.Serve(listener) }()
	ctx, stop := signal.NotifyContext(t.Context(), os.Interrupt, syscall.SIGTERM, syscall.SIGUSR1)
	defer stop()
	timer := time.NewTimer(20 * time.Minute)
	defer timer.Stop()
	t.Logf("Auth test local: backend=127.0.0.1:18080; APP_ORIGIN=%s; PID=%d; dừng bằng SIGUSR1 hoặc tự dừng sau 20 phút", origin, os.Getpid())
	// Không in credential; thông tin đăng nhập demo được ghi trong docs/api/README.md.
	select {
	case <-ctx.Done():
	case <-timer.C:
	case err := <-result:
		if err != nil && err != http.ErrServerClosed {
			t.Fatal("server Swagger test dừng ngoài dự kiến")
		}
	}
}
