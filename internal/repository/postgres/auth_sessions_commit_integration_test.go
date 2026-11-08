//go:build integration

package postgres_test

import (
	"context"
	"crypto/sha256"
	"errors"
	"io"
	"reflect"
	"testing"
	"time"

	"ai-staffing/internal/repository/postgres"
	"ai-staffing/internal/service/auth"
	"ai-staffing/internal/testutil/authdb"
	"github.com/jackc/pgx/v5"
)

// DB vẫn commit thật; chỉ mô phỏng biên driver mất phản hồi commit.
// Đây không phải bằng chứng đã cắt mạng PostgreSQL thật.
type lostCommitReplyDB struct{ *pgx.Conn }

func (db lostCommitReplyDB) Begin(ctx context.Context) (pgx.Tx, error) {
	tx, err := db.Conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return lostCommitReplyTx{Tx: tx}, nil
}

type lostCommitReplyTx struct{ pgx.Tx }

func (tx lostCommitReplyTx) Commit(ctx context.Context) error {
	if err := tx.Tx.Commit(ctx); err != nil {
		return err
	}
	return io.ErrUnexpectedEOF
}

func TestAuthSessionUnknownCommitOutcome(t *testing.T) {
	for _, operation := range []string{"login", "logout"} {
		t.Run(operation, func(t *testing.T) {
			f := authdb.NewFixture(t)
			a := f.AddAccount(t, authdb.AccountOptions{Email: "commit-reply@example.invalid"})
			clock := clockFunc(func() time.Time { return sessionTestTime })
			reader := postgres.NewAuthRepository(f.DB.Conn, clock)
			repo := postgres.NewAuthRepository(lostCommitReplyDB{f.DB.Conn}, clock)
			cmd := loginCommand(t, reader, "commit-reply@example.invalid", "committed but reply lost")
			old := auth.SessionHash(sha256.Sum256([]byte("unknown commit old")))
			cmd.PreviousTokenHash = &old
			f.AddSession(t, a.UserID, old, sessionTestTime.Add(-time.Hour), sessionTestTime.Add(time.Hour), nil)
			var err error
			if operation == "login" {
				var got auth.SessionSnapshot
				got, err = repo.FinalizeLogin(context.Background(), cmd)
				if !reflect.DeepEqual(got, auth.SessionSnapshot{}) {
					t.Fatal("mất phản hồi commit không được cấp snapshot")
				}
			} else {
				err = repo.RevokeCurrentSession(context.Background(), old)
			}
			if !errors.Is(err, auth.ErrUnavailable) {
				t.Fatal("mất phản hồi commit phải trả lỗi an toàn")
			}
			// Quan sát từ DB cho thấy vì sao không thể hứa rollback mọi lỗi Commit.
			stored := sessionSnapshot(t, reader, old)
			if stored.RevokedAt == nil || !stored.RevokedAt.Equal(sessionTestTime) {
				t.Fatal("thiếu commit đã diễn ra trong DB")
			}
			if operation == "login" {
				created := sessionSnapshot(t, reader, cmd.NewTokenHash)
				if !created.CreatedAt.Equal(sessionTestTime) {
					t.Fatal("thiếu phiên đã commit")
				}
			}
		})
	}
}
