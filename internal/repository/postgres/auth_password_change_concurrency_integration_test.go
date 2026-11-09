//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"ai-staffing/internal/repository/postgres"
	"ai-staffing/internal/service/auth"
	"ai-staffing/internal/testutil/authdb"
)

func TestChangeInitialPasswordContention(t *testing.T) {
	for _, operation := range []string{"second change", "stale login"} {
		t.Run(operation, func(t *testing.T) {
			f := authdb.NewFixture(t)
			const email = "change-race@example.invalid"
			a := f.AddAccount(t, authdb.AccountOptions{Email: email, MustChangePassword: true})
			expiry := sessionTestTime.Add(time.Hour)
			firstToken, _ := addPasswordChangeSession(t, f, a.UserID, "first-change-race", expiry, nil)
			secondToken, _ := addPasswordChangeSession(t, f, a.UserID, "second-change-race", expiry, nil)
			winnerConn, waiterConn := f.DB.Connect(t), f.DB.Connect(t)
			holderPID, waiterPID := winnerConn.PgConn().PID(), waiterConn.PgConn().PID()

			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			var workers sync.WaitGroup
			defer func() {
				cancel()
				workers.Wait()
			}()
			locked, release := make(chan struct{}), make(chan struct{})
			winner := postgres.NewAuthRepository(winnerConn, clockFunc(func() time.Time {
				// Clock chỉ được gọi sau khi transaction đã lấy đủ khóa.
				close(locked)
				select {
				case <-release:
				case <-ctx.Done():
				}
				return sessionTestTime
			}))
			clock := clockFunc(func() time.Time { return sessionTestTime })
			reader := postgres.NewAuthRepository(f.DB.Conn, clock)
			waiter := postgres.NewAuthRepository(waiterConn, clock)

			winnerCmd := auth.ChangeInitialPasswordCommand{
				TokenHash: firstToken, UserID: a.UserID, VerifiedPasswordHash: f.PasswordHash,
				NewPasswordHash: replacementHash(t, "Synthetic winner password"),
			}
			waiterCmd := winnerCmd
			waiterCmd.TokenHash = secondToken
			waiterCmd.NewPasswordHash = replacementHash(t, "Synthetic losing password")
			stale := loginCommand(t, reader, email, "stale-login-after-change")

			winnerResult, waiterResult := make(chan error, 1), make(chan error, 1)
			workers.Go(func() { winnerResult <- winner.ChangeInitialPassword(ctx, winnerCmd) })
			select {
			case <-locked:
			case err := <-winnerResult:
				authdb.Require(t, "transaction đầu kết thúc trước barrier", err)
				t.Fatal("transaction đầu chưa tới barrier")
			case <-ctx.Done():
				t.Fatal("transaction đầu chưa lấy đủ khóa trước timeout")
			}
			workers.Go(func() {
				if operation == "second change" {
					waiterResult <- waiter.ChangeInitialPassword(ctx, waiterCmd)
					return
				}
				_, err := waiter.FinalizeLogin(ctx, stale)
				waiterResult <- err
			})

			observer, err := f.DB.Conn.Begin(ctx)
			authdb.Require(t, "mở observer của barrier", err)
			defer observer.Rollback(ctx)
			waitForAuthLock(t, ctx, observer, holderPID, waiterPID)
			authdb.Require(t, "đóng observer của barrier", observer.Rollback(ctx))
			close(release)
			authdb.Require(t, "đổi mật khẩu thắng race", <-winnerResult)

			want := auth.ErrUnauthenticated
			if operation == "stale login" {
				want = auth.ErrInvalidCredentials
			}
			if err := <-waiterResult; !errors.Is(err, want) {
				t.Fatal("request chờ khóa phải từ chối snapshot cũ")
			}
			assertPasswordChangeState(t, f, a.UserID, winnerCmd.NewPasswordHash, false, 2)
			for _, token := range []auth.SessionHash{firstToken, secondToken} {
				if !sameOptionalTime(sessionSnapshot(t, reader, token).RevokedAt, &sessionTestTime) {
					t.Fatal("cả hai phiên của user phải bị thu hồi")
				}
			}
			if operation == "stale login" {
				assertNoSession(t, f.DB, reader, stale.NewTokenHash)
				fresh := loginCommand(t, reader, email, "fresh-login-after-change")
				if fresh.VerifiedIdentity.PasswordHash != winnerCmd.NewPasswordHash {
					t.Fatal("định danh mới chưa đọc được hash vừa đổi")
				}
				got, err := reader.FinalizeLogin(ctx, fresh)
				authdb.Require(t, "hoàn tất đăng nhập từ nền mới", err)
				if got.Principal.MustChangePassword || !sessionSnapshot(t, reader, fresh.NewTokenHash).CreatedAt.Equal(sessionTestTime) {
					t.Fatal("phiên đăng nhập mới phải có cờ false và được lưu")
				}
			}
		})
	}
}

func TestChangeInitialPasswordExpiresWhileWaitingForLaterSession(t *testing.T) {
	f := authdb.NewFixture(t)
	a := f.AddAccount(t, authdb.AccountOptions{Email: "wait-expiry@example.invalid", MustChangePassword: true})
	expiry := sessionTestTime.Add(time.Minute)
	token, firstID := addPasswordChangeSession(t, f, a.UserID, "expiring caller", expiry, nil)
	secondToken, laterID := addPasswordChangeSession(t, f, a.UserID, "later session", expiry.Add(time.Hour), nil)
	if laterID <= firstID {
		t.Fatal("fixture cần khóa một phiên có ID lớn hơn phiên gọi")
	}
	cmd := auth.ChangeInitialPasswordCommand{
		TokenHash: token, UserID: a.UserID, VerifiedPasswordHash: f.PasswordHash,
		NewPasswordHash: replacementHash(t, "Synthetic expired request password"),
	}
	var now atomic.Int64
	now.Store(sessionTestTime.UnixNano())
	clock := clockFunc(func() time.Time { return time.Unix(0, now.Load()).UTC() })
	worker := f.DB.Connect(t)
	repo := postgres.NewAuthRepository(worker, clock)
	holderPID, waiterPID := f.DB.Conn.PgConn().PID(), worker.PgConn().PID()

	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	var workers sync.WaitGroup
	defer func() {
		cancel()
		workers.Wait()
	}()
	blocker, err := f.DB.Conn.Begin(ctx)
	authdb.Require(t, "bắt đầu giữ khóa phiên thứ hai", err)
	defer blocker.Rollback(ctx)
	var lockedID int
	err = blocker.QueryRow(ctx, `SELECT session_id FROM public.auth_sessions WHERE session_id=$1 FOR UPDATE`, laterID).Scan(&lockedID)
	authdb.Require(t, "khóa phiên đứng sau phiên gọi", err)

	result := make(chan error, 1)
	workers.Go(func() { result <- repo.ChangeInitialPassword(ctx, cmd) })
	waitForAuthLock(t, ctx, blocker, holderPID, waiterPID)
	now.Store(expiry.UnixNano())
	authdb.Require(t, "nhả khóa khi phiên gọi vừa hết hạn", blocker.Commit(ctx))
	if err := <-result; !errors.Is(err, auth.ErrUnauthenticated) {
		t.Fatal("phiên gọi hết hạn đúng biên trong lúc chờ phiên khác phải bị từ chối")
	}
	assertPasswordChangeState(t, f, a.UserID, f.PasswordHash, true, 2)
	reader := postgres.NewAuthRepository(f.DB.Conn, clock)
	for _, hash := range []auth.SessionHash{token, secondToken} {
		if sessionSnapshot(t, reader, hash).RevokedAt != nil {
			t.Fatal("request hết hạn đã thay đổi phiên")
		}
	}
}
