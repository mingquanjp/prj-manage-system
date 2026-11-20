//go:build integration && accountuidemo

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"ai-staffing/internal/app"
	"ai-staffing/internal/config"
	"ai-staffing/internal/service/organization"
	"ai-staffing/internal/testutil/authdb"
	"golang.org/x/crypto/bcrypt"
)

// TestAccountUIDemoServer chỉ chạy opt-in trên DB case sở hữu. Không gửi Gmail thật.
// Email tổng hợp chọn fault mail tại port; route và transaction vẫn là adapter thật.
func TestAccountUIDemoServer(t *testing.T) {
	f := authdb.NewFixture(t)
	hash, err := bcrypt.GenerateFromPassword([]byte("12345678"), bcrypt.DefaultCost)
	authdb.Require(t, "hash fixture UI", err)
	f.PasswordHash = string(hash)
	for _, item := range []struct {
		email   string
		roles   []string
		limited bool
	}{
		{"admin@example.invalid", []string{"ADMIN"}, false},
		{"hr@example.invalid", []string{"HR"}, false},
		{"employee@example.invalid", nil, false},
		{"limited@example.invalid", []string{"ADMIN", "HR"}, true},
	} {
		f.AddAccount(t, authdb.AccountOptions{Email: item.email, Roles: item.roles, MustChangePassword: item.limited})
	}
	var calls atomic.Int64
	sender := accountUIMail(func(_ context.Context, email string, _ organization.TemporaryPassword) (organization.MailStatus, error) {
		calls.Add(1)
		switch {
		case strings.HasPrefix(email, "failed"):
			return organization.MailFailed, nil
		case strings.HasPrefix(email, "unknown"):
			return organization.MailUnknown, nil
		default:
			return organization.MailAccepted, nil
		}
	})
	handler, err := app.AuthHandlerWithMail(&config.Config{AppEnv: "dev", AppOrigin: "http://127.0.0.1:5174", SessionTTL: 8 * time.Hour}, f.DB.Pool(t), systemClock{}, sender)
	authdb.Require(t, "lắp API UI test", err)
	listener, err := net.Listen("tcp", "127.0.0.1:18080")
	authdb.Require(t, "mở UI test loopback", err)
	wrapped := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/employees" {
			handler.ServeHTTP(w, r)
			return
		}
		// Buffer phản hồi để mô phỏng mất mạng sau khi handler đã commit/gọi port mail.
		output := &accountUIResponse{header: make(http.Header)}
		handler.ServeHTTP(output, r)
		var result struct {
			Account struct {
				Email string `json:"company_email"`
			} `json:"account"`
		}
		_ = json.Unmarshal(output.body.Bytes(), &result)
		if strings.HasPrefix(result.Account.Email, "lost") && output.status == 201 {
			conn, _, err := w.(http.Hijacker).Hijack()
			if err == nil {
				_ = conn.Close()
			}
			return
		}
		for key, values := range output.header {
			w.Header()[key] = values
		}
		w.WriteHeader(output.status)
		_, _ = w.Write(output.body.Bytes())
	})
	server := &http.Server{Handler: wrapped, ReadHeaderTimeout: 5 * time.Second, WriteTimeout: 30 * time.Second}
	defer server.Close()
	result := make(chan error, 1)
	go func() { result <- server.Serve(listener) }()
	ctx, stop := signal.NotifyContext(t.Context(), os.Interrupt, syscall.SIGTERM, syscall.SIGUSR1)
	defer stop()
	timer := time.NewTimer(45 * time.Minute)
	defer timer.Stop()
	t.Logf("Account UI test ready: 18080, PID=%d; fake mail only; stop SIGUSR1", os.Getpid())
	select {
	case <-ctx.Done():
	case <-timer.C:
	case err := <-result:
		if err != http.ErrServerClosed {
			t.Fatal("server UI test lỗi")
		}
	}
	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdown)
	var created int
	err = f.DB.Pool(t).QueryRow(context.Background(), `SELECT count(*) FROM employees WHERE email LIKE '%@ui.example.invalid'`).Scan(&created)
	authdb.Require(t, "đối soát số row UI", err)
	t.Logf("UI test đối soát: %d account rows; %d fake mail calls (không chứng minh gửi thật)", created, calls.Load())
}

type accountUIMail func(context.Context, string, organization.TemporaryPassword) (organization.MailStatus, error)

func (f accountUIMail) SendAccount(c context.Context, e string, p organization.TemporaryPassword) (organization.MailStatus, error) {
	return f(c, e, p)
}

type accountUIResponse struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func (w *accountUIResponse) Header() http.Header { return w.header }
func (w *accountUIResponse) WriteHeader(code int) {
	if w.status == 0 {
		w.status = code
	}
}
func (w *accountUIResponse) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	return w.body.Write(data)
}
