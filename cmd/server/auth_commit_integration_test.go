//go:build integration

package main

import (
	"context"
	"io"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"ai-staffing/internal/repository/postgres"
	"ai-staffing/internal/service/auth"
	"ai-staffing/internal/testutil/authdb"
	httptransport "ai-staffing/internal/transport/http"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

// Commit PostgreSQL thật rồi mô phỏng mất phản hồi ở biên driver.
// Không mô tả đây là thử nghiệm cắt mạng thật.
type lostHTTPCommitDB struct {
	*pgx.Conn
	commits atomic.Int32
}

func (db *lostHTTPCommitDB) Begin(ctx context.Context) (pgx.Tx, error) {
	tx, err := db.Conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &lostHTTPCommitTx{Tx: tx, db: db}, nil
}

type lostHTTPCommitTx struct {
	pgx.Tx
	db *lostHTTPCommitDB
}

func (tx *lostHTTPCommitTx) Commit(ctx context.Context) error {
	if err := tx.Tx.Commit(ctx); err != nil {
		return err
	}
	tx.db.commits.Add(1)
	return io.ErrUnexpectedEOF
}

func TestAuthHTTPPostgresUnknownCommitOutcome(t *testing.T) {
	for _, operation := range []string{"login", "logout", "change-password"} {
		t.Run(operation, func(t *testing.T) {
			a := newAuthApp(t)
			account := a.fixture.AddAccount(t, authdb.AccountOptions{Email: "lost-reply@example.invalid", MustChangePassword: true})
			cookie := a.login(t, "lost-reply@example.invalid", authdb.FixturePassword, nil)
			token := a.csrf(t, cookie)
			db := &lostHTTPCommitDB{Conn: a.fixture.DB.Connect(t)}
			repo := postgres.NewAuthRepository(db, a.clock)
			handler, err := httptransport.NewAuthRouter(db, "dev", testOrigin, httptransport.AuthServices{
				Sessions: auth.NewService(repo, a.clock), Login: auth.NewLoginService(repo, a.clock, 8*time.Hour), PasswordChange: auth.NewPasswordChangeService(repo, a.clock), Clock: a.clock,
			})
			authdb.Require(t, "lắp router với biên commit test", err)
			faultServer := httptest.NewServer(handler)
			t.Cleanup(faultServer.Close)
			a.server = faultServer
			var payload any
			if operation == "login" {
				payload = map[string]string{"email": "lost-reply@example.invalid", "password": authdb.FixturePassword}
			}
			if operation == "change-password" {
				payload = map[string]string{"current_password": authdb.FixturePassword, "new_password": nextPassword}
			}
			header, data := a.call(t, "POST", "/api/auth/"+operation, cookie, token, payload, 503)
			assertErrorCode(t, data, "SERVICE_UNAVAILABLE")
			if header.Get("Set-Cookie") != "" || db.commits.Load() != 1 {
				t.Fatal("commit chưa rõ kết quả vẫn success/retry mutation")
			}
			var hash string
			var limited bool
			var live, total int
			err = a.fixture.DB.Conn.QueryRow(t.Context(), `SELECT password_hash,must_change_password,(SELECT count(*) FROM auth_sessions WHERE user_id=$1 AND revoked_at IS NULL),(SELECT count(*) FROM auth_sessions WHERE user_id=$1) FROM users WHERE user_id=$1`, account.UserID).Scan(&hash, &limited, &live, &total)
			authdb.Require(t, "đối soát trạng thái đã commit", err)
			wantLive, wantTotal := 0, 1
			if operation == "login" {
				wantLive, wantTotal = 1, 2
			}
			if live != wantLive || total != wantTotal {
				t.Fatal("trạng thái DB sau mất phản hồi không khớp commit thật")
			}
			if operation == "change-password" {
				if limited || bcrypt.CompareHashAndPassword([]byte(hash), []byte(nextPassword)) != nil {
					t.Fatal("mất phản hồi không phản ánh đổi mật khẩu đã commit")
				}
			} else if !limited || hash != a.fixture.PasswordHash {
				t.Fatal("mutation phiên đã đổi mật khẩu/cờ")
			}
		})
	}
}
