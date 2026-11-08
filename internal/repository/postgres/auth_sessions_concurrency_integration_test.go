//go:build integration

package postgres_test

import (
	"context"
	"crypto/sha256"
	"errors"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"ai-staffing/internal/repository/postgres"
	"ai-staffing/internal/service/auth"
	"ai-staffing/internal/testutil/authdb"
	"github.com/jackc/pgx/v5"
)

// Chờ PostgreSQL xác nhận quan hệ chặn thật, không dùng thời gian ngủ để đoán race.
func waitForAuthLock(t *testing.T, ctx context.Context, blocker pgx.Tx, holder, waiter uint32) {
	t.Helper()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		var blocked bool
		err := blocker.QueryRow(ctx, `SELECT $1::int=ANY(pg_blocking_pids($2::int))`, holder, waiter).Scan(&blocked)
		authdb.Require(t, "đọc barrier khóa PostgreSQL", err)
		if blocked {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatal("không xác nhận được barrier khóa trước timeout")
		case <-ticker.C:
		}
	}
}

func TestAuthSessionContention(t *testing.T) {
	for _, tc := range []struct {
		name, operation, lock, mutation string
		advance, cancel                 bool
		want                            error
	}{
		{"login chờ reset hash", "login", "user", "hash", false, false, auth.ErrInvalidCredentials},
		{"login chờ khóa user", "login", "user", "user", false, false, auth.ErrInvalidCredentials},
		{"login chờ đổi email", "login", "employee", "email", false, false, auth.ErrInvalidCredentials},
		{"login chờ vô hiệu employee", "login", "employee", "employee", false, false, auth.ErrInvalidCredentials},
		{"login chờ user và cấp theo giờ mới", "login", "user", "", true, false, nil},
		{"login phiên cũ hết hạn khi chờ session", "login", "session", "", true, false, nil},
		{"login hủy khi chờ khóa", "login", "user", "", false, true, context.Canceled},
		{"logout chờ khóa user", "logout", "user", "user", false, false, auth.ErrUnauthenticated},
		{"logout chờ vô hiệu employee", "logout", "employee", "employee", false, false, auth.ErrUnauthenticated},
		{"logout chờ thu hồi session", "logout", "session", "revoke", false, false, auth.ErrUnauthenticated},
		{"logout hết hạn khi chờ user", "logout", "user", "", true, false, auth.ErrUnauthenticated},
		{"logout hết hạn khi chờ session", "logout", "session", "", true, false, auth.ErrUnauthenticated},
		{"logout hủy khi chờ khóa", "logout", "user", "", false, true, context.Canceled},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := authdb.NewFixture(t)
			a := f.AddAccount(t, authdb.AccountOptions{Email: "race@example.invalid"})
			var current atomic.Int64
			current.Store(sessionTestTime.UnixNano())
			clock := clockFunc(func() time.Time { return time.Unix(0, current.Load()).UTC() })
			reader := postgres.NewAuthRepository(f.DB.Conn, clock)
			cmd := loginCommand(t, reader, "race@example.invalid", "race new")
			old := auth.SessionHash(sha256.Sum256([]byte("race old")))
			cmd.PreviousTokenHash = &old
			expires := sessionTestTime.Add(time.Minute)
			f.AddSession(t, a.UserID, old, sessionTestTime.Add(-time.Hour), expires, nil)
			worker := f.DB.Connect(t)
			repo := postgres.NewAuthRepository(worker, clock)
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			tx, err := f.DB.Conn.Begin(ctx)
			authdb.Require(t, "bắt đầu mutation giả lập", err)
			defer tx.Rollback(context.Background())
			var statement string
			var id int
			switch tc.lock {
			case "user":
				statement, id = `SELECT user_id FROM public.users WHERE user_id=$1 FOR UPDATE`, a.UserID
			case "employee":
				statement, id = `SELECT employee_id FROM public.employees WHERE user_id=$1 FOR UPDATE`, a.UserID
			case "session":
				statement, id = `SELECT session_id FROM public.auth_sessions WHERE user_id=$1 FOR UPDATE`, a.UserID
			}
			var locked int
			authdb.Require(t, "giữ khóa barrier", tx.QueryRow(ctx, statement, id).Scan(&locked))
			requestCtx, cancelRequest := context.WithCancel(ctx)
			defer cancelRequest()
			type result struct {
				snapshot auth.SessionSnapshot
				err      error
			}
			results := make(chan result, 1)
			done := make(chan struct{})
			go func() {
				defer close(done)
				var res result
				if tc.operation == "login" {
					res.snapshot, res.err = repo.FinalizeLogin(requestCtx, cmd)
				} else {
					res.err = repo.RevokeCurrentSession(requestCtx, old)
				}
				results <- res
			}()
			t.Cleanup(func() {
				cancelRequest()
				select {
				case <-done:
				case <-time.After(5 * time.Second):
					t.Error("goroutine auth chưa kết thúc")
				}
			})
			waitForAuthLock(t, ctx, tx, f.DB.Conn.PgConn().PID(), worker.PgConn().PID())
			if tc.advance {
				current.Store(expires.UnixNano())
			}
			switch tc.mutation {
			case "hash":
				_, err = tx.Exec(ctx, `UPDATE public.users SET password_hash='reset fixture' WHERE user_id=$1`, a.UserID)
				if err == nil {
					_, err = tx.Exec(ctx, `UPDATE public.auth_sessions SET revoked_at=$1 WHERE user_id=$2`, sessionTestTime, a.UserID)
				}
			case "user":
				_, err = tx.Exec(ctx, `UPDATE public.users SET status='LOCKED' WHERE user_id=$1`, a.UserID)
			case "email":
				_, err = tx.Exec(ctx, `UPDATE public.employees SET email='race-changed@example.invalid' WHERE user_id=$1`, a.UserID)
			case "employee":
				_, err = tx.Exec(ctx, `UPDATE public.employees SET employment_status='INACTIVE' WHERE user_id=$1`, a.UserID)
			case "revoke":
				_, err = tx.Exec(ctx, `UPDATE public.auth_sessions SET revoked_at=$1 WHERE user_id=$2`, sessionTestTime, a.UserID)
			}
			authdb.Require(t, "mutation dữ liệu giả", err)
			if tc.cancel {
				// Hủy trong khi PostgreSQL vẫn giữ khóa, không nhả rồi mới hủy.
				cancelRequest()
				select {
				case <-done:
				case <-ctx.Done():
					t.Fatal("request không kết thúc khi bị hủy dưới khóa")
				}
			}
			authdb.Require(t, "commit mutation giả lập", tx.Commit(ctx))
			var got result
			select {
			case got = <-results:
			case <-ctx.Done():
				t.Fatal("request không kết thúc sau nhả khóa")
			}
			if !errors.Is(got.err, tc.want) {
				t.Fatal("kết quả race sai phân loại")
			}
			if tc.operation == "login" && tc.want == nil {
				if !got.snapshot.CreatedAt.Equal(expires) || !got.snapshot.ExpiresAt.Equal(expires.Add(8*time.Hour)) {
					t.Fatal("phải dùng thời gian sau khi chờ đủ khóa")
				}
				if sessionSnapshot(t, reader, old).RevokedAt != nil {
					t.Fatal("không ghi thu hồi phiên đã hết hạn")
				}
			} else {
				if !reflect.DeepEqual(got.snapshot, auth.SessionSnapshot{}) {
					t.Fatal("lỗi không được trả snapshot thành công")
				}
				assertNoSession(t, f.DB, reader, cmd.NewTokenHash)
				if tc.mutation != "hash" && tc.mutation != "revoke" && sessionSnapshot(t, reader, old).RevokedAt != nil {
					t.Fatal("request bị từ chối không được tự thu hồi phiên")
				}
			}
		})
	}
}
