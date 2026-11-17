//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"ai-staffing/internal/repository/postgres"
	org "ai-staffing/internal/service/organization"
	"ai-staffing/internal/testutil/authdb"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type accountCommitRollbackDB struct{ *pgx.Conn }

func (db accountCommitRollbackDB) BeginTx(ctx context.Context, o pgx.TxOptions) (pgx.Tx, error) {
	tx, e := db.Conn.BeginTx(ctx, o)
	if e != nil {
		return nil, e
	}
	return accountCommitRollbackTx{tx}, nil
}

type accountCommitRollbackTx struct{ pgx.Tx }

func (tx accountCommitRollbackTx) Commit(ctx context.Context) error {
	if err := tx.Tx.Rollback(ctx); err != nil {
		return err
	}
	return pgx.ErrTxCommitRollback
}

func TestAccountCommitOutcomeAndNoFalseSuccess(t *testing.T) {
	for _, mode := range []string{"known rollback", "deferred database rejection", "commit reply lost", "commit fatal reply", "commit statement outcome unknown"} {
		t.Run(mode, func(t *testing.T) {
			f := authdb.NewFixture(t)
			a := f.AddAccount(t, authdb.AccountOptions{Email: "admin@example.invalid", Roles: []string{"ADMIN"}})
			cmd := accountCommand(t, f, a, "created@example.invalid")
			before := accountSnapshot(t, f)
			var db accountBeginner = f.DB.Conn
			switch mode {
			case "known rollback":
				db = accountCommitRollbackDB{f.DB.Conn}
			case "commit reply lost":
				db = catalogHookDB{db: f.DB.Conn, commitLost: true}
			case "commit fatal reply":
				db = accountCommitFatalDB{f.DB.Conn}
			case "commit statement outcome unknown":
				db = accountCommitStatementUnknownDB{f.DB.Conn}
			case "deferred database rejection":
				_, err := f.DB.Conn.Exec(t.Context(), `CREATE FUNCTION public.account_test_commit_failure() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'synthetic deferred failure' USING ERRCODE='23514'; END $$;
 CREATE CONSTRAINT TRIGGER account_test_commit_failure AFTER INSERT ON public.audit_events DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION public.account_test_commit_failure()`)
				authdb.Require(t, "fault tại commit thực", err)
			}
			got, err := accountRepo(t, db).CreateAccount(t.Context(), cmd)
			accountNoResult(t, got)
			if mode == "commit reply lost" || mode == "commit fatal reply" || mode == "commit statement outcome unknown" {
				accountError(t, err, org.ErrCommitUnknown)
				if catalogCount(t, f, `SELECT count(*) FROM users`) != 2 || catalogCount(t, f, `SELECT count(*) FROM employees`) != 2 || catalogCount(t, f, `SELECT count(*) FROM audit_events WHERE operation='CREATE_ACCOUNT'`) != 1 {
					t.Fatal("lost reply phải có write commit thật một lần, không retry")
				}
			} else {
				accountError(t, err, org.ErrPersistence)
				if errors.Is(err, org.ErrCommitUnknown) {
					t.Fatal("DB đã xác nhận rollback không được gọi chưa xác định")
				}
				if accountSnapshot(t, f) != before {
					t.Fatal("commit lỗi đã half-write")
				}
			}
		})
	}
}

func TestAccountCreateLockTimeoutRollsBack(t *testing.T) {
	f := authdb.NewFixture(t)
	a := f.AddAccount(t, authdb.AccountOptions{Email: "admin@example.invalid", Roles: []string{"ADMIN"}})
	cmd := accountCommand(t, f, a, "timeout@example.invalid")
	ctx := t.Context()
	tx, err := f.DB.Conn.Begin(ctx)
	authdb.Require(t, "giữ khóa actor", err)
	defer tx.Rollback(context.Background())
	var id int
	authdb.Require(t, "khóa user", tx.QueryRow(ctx, `SELECT user_id FROM users WHERE user_id=$1 FOR UPDATE`, a.UserID).Scan(&id))
	worker := f.DB.Connect(t)
	r, err := postgres.NewAccountRepository(worker, 50*time.Millisecond)
	authdb.Require(t, "repo timeout", err)
	got, err := r.CreateAccount(ctx, cmd)
	accountError(t, err, org.ErrUnavailable)
	accountNoResult(t, got)
	authdb.Require(t, "nhả khóa", tx.Rollback(ctx))
	if !worker.IsClosed() && worker.PgConn().TxStatus() != 'I' {
		t.Fatal("sau timeout connection phải idle hoặc đóng")
	}
	if catalogCount(t, f, `SELECT count(*) FROM users`) != 1 || catalogCount(t, f, `SELECT count(*) FROM audit_events`) != 0 {
		t.Fatal("timeout tạo write")
	}
}

func TestAccountManyCreatesKeepCodesUnique(t *testing.T) {
	f := authdb.NewFixture(t)
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	const n = 4
	repos := make([]*postgres.AccountRepository, n)
	cmds := make([]org.CreateAccountCommand, n)
	for i := range n {
		a := f.AddAccount(t, authdb.AccountOptions{Email: fmt.Sprintf("admin%d@example.invalid", i), Roles: []string{"ADMIN"}})
		cmds[i] = accountCommand(t, f, a, fmt.Sprintf("new%d@example.invalid", i))
		repos[i] = accountRepo(t, f.DB.Connect(t))
	}
	type result struct {
		a   org.AccountProjection
		err error
	}
	start := make(chan struct{})
	done := make(chan result, n)
	for i := range n {
		go func() { <-start; a, e := repos[i].CreateAccount(ctx, cmds[i]); done <- result{a, e} }()
	}
	close(start)
	codes := map[string]bool{}
	for range n {
		select {
		case got := <-done:
			authdb.Require(t, "create đồng thời", got.err)
			if codes[got.a.EmployeeCode] || got.a.EmployeeCode == "" {
				t.Fatal("N create cấp trùng mã")
			}
			codes[got.a.EmployeeCode] = true
		case <-ctx.Done():
			t.Fatal("N create timeout")
		}
	}
	if catalogCount(t, f, `SELECT count(*) FROM audit_events WHERE operation='CREATE_ACCOUNT'`) != n || catalogCount(t, f, `SELECT count(*) FROM employees`) != 2*n {
		t.Fatal("N create thiếu liên kết hoặc audit")
	}
}

// DB commit thật, wrapper chỉ thay phản hồi bằng lỗi kết nối có SQLSTATE.
type accountCommitFatalDB struct{ *pgx.Conn }

func (db accountCommitFatalDB) BeginTx(ctx context.Context, o pgx.TxOptions) (pgx.Tx, error) {
	tx, e := db.Conn.BeginTx(ctx, o)
	if e != nil {
		return nil, e
	}
	return accountCommitFatalTx{tx}, nil
}

type accountCommitFatalTx struct{ pgx.Tx }

func (tx accountCommitFatalTx) Commit(ctx context.Context) error {
	if err := tx.Tx.Commit(ctx); err != nil {
		return err
	}
	return &pgconn.PgError{Code: "08006", Severity: "FATAL", Message: "synthetic connection lost"}
}

// SQLSTATE 40003 vẫn chưa xác định dù phản hồi có Severity ERROR.
type accountCommitStatementUnknownDB struct{ *pgx.Conn }

func (db accountCommitStatementUnknownDB) BeginTx(ctx context.Context, o pgx.TxOptions) (pgx.Tx, error) {
	tx, e := db.Conn.BeginTx(ctx, o)
	if e != nil {
		return nil, e
	}
	return accountCommitStatementUnknownTx{tx}, nil
}

type accountCommitStatementUnknownTx struct{ pgx.Tx }

func (tx accountCommitStatementUnknownTx) Commit(ctx context.Context) error {
	if err := tx.Tx.Commit(ctx); err != nil {
		return err
	}
	return &pgconn.PgError{Code: "40003", Severity: "ERROR", Message: "synthetic outcome unknown"}
}
