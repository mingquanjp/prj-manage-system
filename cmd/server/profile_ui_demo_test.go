//go:build integration && profileuidemo

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"testing"
	"time"

	"ai-staffing/internal/app"
	"ai-staffing/internal/config"
	"ai-staffing/internal/service/organization"
	"ai-staffing/internal/testutil/authdb"
	"golang.org/x/crypto/bcrypt"
)

type profileUIMail struct{}

func (profileUIMail) SendAccount(context.Context, string, organization.TemporaryPassword) (organization.MailStatus, error) {
	return organization.MailFailed, nil // Harness hồ sơ không gửi Gmail.
}

// TestProfileUIDemoServer chỉ mở opt-in trên DB case sở hữu. Marker lỗi chỉ tồn tại trong test.
func TestProfileUIDemoServer(t *testing.T) {
	f := authdb.NewFixture(t)
	hash, err := bcrypt.GenerateFromPassword([]byte("12345678"), bcrypt.DefaultCost)
	authdb.Require(t, "hash fixture UI hồ sơ", err)
	f.PasswordHash = string(hash)
	employee := f.AddAccount(t, authdb.AccountOptions{Email: "employee@example.invalid"})
	f.AddAccount(t, authdb.AccountOptions{Email: "hr@example.invalid", Roles: []string{"HR"}})
	f.AddAccount(t, authdb.AccountOptions{Email: "limited@example.invalid", Roles: []string{"HR", "ADMIN"}, MustChangePassword: true})
	_, err = f.DB.Conn.Exec(t.Context(), `UPDATE employees SET address=NULL,phone=NULL WHERE employee_id=$1`, employee.EmployeeID)
	authdb.Require(t, "fixture NULL UI hồ sơ", err)
	handler, err := app.AuthHandlerWithMail(&config.Config{AppEnv: "dev", AppOrigin: "http://127.0.0.1:5174", SessionTTL: 8 * time.Hour}, f.DB.Pool(t), systemClock{}, profileUIMail{})
	authdb.Require(t, "lắp API UI hồ sơ", err)
	wrapped := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/profile-change-requests") {
			handler.ServeHTTP(w, r)
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, 65538))
		if err != nil {
			http.Error(w, "test body", 500)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
		var input struct {
			Explanation string `json:"explanation"`
		}
		_ = json.Unmarshal(body, &input)
		if input.Explanation == "PCA43_STALE" {
			_, err := f.DB.Pool(t).Exec(r.Context(), `UPDATE employees SET fullname='Tên nền đã đổi' WHERE employee_id=$1`, employee.EmployeeID)
			if err != nil {
				http.Error(w, "test fault", 500)
				return
			}
		}
		if input.Explanation != "PCA43_LOST" {
			handler.ServeHTTP(w, r)
			return
		}
		output := httptest.NewRecorder()
		handler.ServeHTTP(output, r)
		if output.Code == http.StatusCreated {
			connection, _, err := w.(http.Hijacker).Hijack()
			if err == nil {
				_ = connection.Close()
			}
			return
		}
		for key, values := range output.Header() {
			w.Header()[key] = values
		}
		w.WriteHeader(output.Code)
		_, _ = w.Write(output.Body.Bytes())
	})
	listener, err := net.Listen("tcp", "127.0.0.1:18080")
	authdb.Require(t, "mở loopback UI hồ sơ", err)
	server := &http.Server{Handler: wrapped, ReadHeaderTimeout: 5 * time.Second, WriteTimeout: 30 * time.Second}
	defer server.Close()
	result := make(chan error, 1)
	go func() { result <- server.Serve(listener) }()
	ctx, stop := signal.NotifyContext(t.Context(), os.Interrupt, syscall.SIGTERM, syscall.SIGUSR1)
	defer stop()
	timer := time.NewTimer(40 * time.Minute)
	defer timer.Stop()
	t.Logf("Profile UI test ready: 18080 PID=%d; no Gmail; stop SIGUSR1", os.Getpid())
	select {
	case <-ctx.Done():
	case <-timer.C:
	case err := <-result:
		if err != http.ErrServerClosed {
			t.Fatal("server UI hồ sơ lỗi")
		}
	}
	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdown)
	var count int
	err = f.DB.Pool(t).QueryRow(t.Context(), `SELECT count(*) FROM employee_profile_change_requests WHERE employee_id=$1`, employee.EmployeeID).Scan(&count)
	authdb.Require(t, "đối soát request UI hồ sơ", err)
	t.Logf("Profile UI đối soát: %d requests (browser steps ghi riêng)", count)
}
