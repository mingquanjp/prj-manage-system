//go:build integration

package main

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"ai-staffing/internal/testutil/authdb"
	"golang.org/x/crypto/bcrypt"
)

type authHTTPResult struct {
	response *http.Response
	data     []byte
	err      error
}

// Hỏi PostgreSQL về quan hệ chặn thật; ticker chỉ thăm dò có timeout.
func awaitAuthCondition(t *testing.T, ctx context.Context, check func() bool) {
	t.Helper()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		if check() {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatal("không tới barrier PostgreSQL trước timeout")
		case <-ticker.C:
		}
	}
}

func TestAuthHTTPPostgresContention(t *testing.T) {
	for _, operation := range []string{"second-change", "stale-login"} {
		t.Run(operation, func(t *testing.T) {
			a := newAuthApp(t)
			account := a.fixture.AddAccount(t, authdb.AccountOptions{Email: "race-http@example.invalid", MustChangePassword: true})
			first := a.login(t, "race-http@example.invalid", authdb.FixturePassword, nil)
			second := a.login(t, "race-http@example.invalid", authdb.FixturePassword, nil)
			tokenA, tokenB := a.csrf(t, first), a.csrf(t, second)
			ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
			var workers sync.WaitGroup
			// Giữ barrier trên connection điều khiển, tách khỏi pool xử lý HTTP.
			_, err := a.fixture.DB.Conn.Exec(ctx, `SELECT pg_advisory_lock(18017)`)
			authdb.Require(t, "giữ barrier test", err)
			defer func() {
				_, unlockErr := a.fixture.DB.Conn.Exec(context.Background(), `SELECT pg_advisory_unlock(18017)`)
				if unlockErr != nil {
					t.Error("không nhả được barrier test")
				}
				cancel()
				workers.Wait()
			}()
			_, err = a.fixture.DB.Conn.Exec(ctx, `CREATE FUNCTION pause_change_fixture() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN PERFORM pg_advisory_xact_lock(18017); RETURN NEW; END $$;
 CREATE TRIGGER pause_change BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION pause_change_fixture()`)
			authdb.Require(t, "dựng barrier trước write trong transaction", err)
			winner := make(chan authHTTPResult, 1)
			workers.Go(func() {
				r, b, e := a.request(ctx, "POST", "/api/auth/change-password", first, tokenA, map[string]string{"current_password": authdb.FixturePassword, "new_password": nextPassword})
				winner <- authHTTPResult{r, b, e}
			})
			var winnerPID int
			awaitAuthCondition(t, ctx, func() bool {
				err := a.fixture.DB.Conn.QueryRow(ctx, `SELECT COALESCE((SELECT pid FROM pg_locks WHERE locktype='advisory' AND objid=18017 AND NOT granted AND database=(SELECT oid FROM pg_database WHERE datname=current_database())),0)`).Scan(&winnerPID)
				authdb.Require(t, "xác nhận transaction đầu đã giữ row và chờ barrier", err)
				return winnerPID != 0
			})

			waiter := make(chan authHTTPResult, 1)
			workers.Go(func() {
				path := "/api/auth/change-password"
				cookie, token := second, tokenB
				payload := map[string]string{"current_password": authdb.FixturePassword, "new_password": "Synthetic losing password for HTTP"}
				if operation == "stale-login" {
					path = "/api/auth/login"
					cookie = nil
					token = ""
					payload = map[string]string{"email": "race-http@example.invalid", "password": authdb.FixturePassword}
				}
				r, b, e := a.request(ctx, "POST", path, cookie, token, payload)
				waiter <- authHTTPResult{r, b, e}
			})
			awaitAuthWaiter(t, ctx, a, winnerPID)
			_, err = a.fixture.DB.Conn.Exec(ctx, `SELECT pg_advisory_unlock(18017)`)
			authdb.Require(t, "nhả barrier", err)
			firstResult, result := <-winner, <-waiter
			authdb.Require(t, "HTTP đổi thắng", firstResult.err)
			authdb.Require(t, "HTTP request chờ", result.err)
			if firstResult.response.StatusCode != 204 || result.response.StatusCode != 401 || result.response.Header.Get("Set-Cookie") != "" {
				t.Fatal("request cạnh tranh đã ghi/nhận cookie bằng snapshot cũ")
			}
			if operation == "stale-login" {
				assertErrorCode(t, result.data, "INVALID_CREDENTIALS")
			} else {
				assertErrorCode(t, result.data, "UNAUTHENTICATED")
			}
			var hash string
			var live, total int
			err = a.fixture.DB.Conn.QueryRow(ctx, `SELECT password_hash,(SELECT count(*) FROM auth_sessions WHERE user_id=$1 AND revoked_at IS NULL),(SELECT count(*) FROM auth_sessions WHERE user_id=$1) FROM users WHERE user_id=$1`, account.UserID).Scan(&hash, &live, &total)
			authdb.Require(t, "đọc kết quả race HTTP", err)
			if bcrypt.CompareHashAndPassword([]byte(hash), []byte(nextPassword)) != nil || live != 0 || total != 2 {
				t.Fatal("race ghi đè hash hoặc cấp thêm phiên")
			}
		})
	}
}

func awaitAuthWaiter(t *testing.T, ctx context.Context, a *authApp, holder int) {
	t.Helper()
	awaitAuthCondition(t, ctx, func() bool {
		var blocked bool
		err := a.fixture.DB.Conn.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE datname=current_database() AND $1::int=ANY(pg_blocking_pids(pid)))`, holder).Scan(&blocked)
		authdb.Require(t, "xác nhận request thứ hai chờ row lock", err)
		return blocked
	})
}

func TestAuthHTTPPostgresExpiryWhileWaiting(t *testing.T) {
	a := newAuthApp(t)
	account := a.fixture.AddAccount(t, authdb.AccountOptions{Email: "expiry-wait@example.invalid"})
	cookie := a.login(t, "expiry-wait@example.invalid", authdb.FixturePassword, nil)
	token := a.csrf(t, cookie)
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	blocker := a.fixture.DB.Connect(t)
	tx, err := blocker.Begin(ctx)
	authdb.Require(t, "bắt đầu blocker test", err)
	var workers sync.WaitGroup
	defer func() { tx.Rollback(context.Background()); cancel(); workers.Wait() }()
	_, err = tx.Exec(ctx, `SELECT user_id FROM users WHERE user_id=$1 FOR UPDATE`, account.UserID)
	authdb.Require(t, "khóa user fixture", err)
	result := make(chan authHTTPResult, 1)
	workers.Go(func() {
		r, b, e := a.request(ctx, "POST", "/api/auth/logout", cookie, token, nil)
		result <- authHTTPResult{r, b, e}
	})
	awaitAuthWaiter(t, ctx, a, int(blocker.PgConn().PID()))
	a.clock.nanos.Add(int64(8 * time.Hour))
	authdb.Require(t, "nhả blocker", tx.Rollback(ctx))
	got := <-result
	authdb.Require(t, "HTTP logout hết hạn sau khóa", got.err)
	if got.response.StatusCode != 401 || got.response.Header.Get("Set-Cookie") != "" {
		t.Fatal("logout đã bỏ qua biên thời hạn sau chờ khóa")
	}
	assertErrorCode(t, got.data, "UNAUTHENTICATED")
	var live int
	err = a.fixture.DB.Conn.QueryRow(ctx, `SELECT count(*) FROM auth_sessions WHERE user_id=$1 AND revoked_at IS NULL`, account.UserID).Scan(&live)
	authdb.Require(t, "kiểm request hết hạn không ghi DB", err)
	if live != 1 {
		t.Fatal("request hết hạn vẫn thu hồi phiên")
	}
}
