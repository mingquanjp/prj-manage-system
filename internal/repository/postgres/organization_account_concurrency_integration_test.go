//go:build integration

package postgres_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"ai-staffing/internal/repository/postgres"
	org "ai-staffing/internal/service/organization"
	"ai-staffing/internal/testutil/authdb"
)

func TestAccountConcurrentCreatesAndEmailCollision(t *testing.T) {
	for _, duplicate := range []bool{false, true} {
		t.Run(fmt.Sprintf("duplicate-%t", duplicate), func(t *testing.T) {
			f := authdb.NewFixture(t)
			a1 := f.AddAccount(t, authdb.AccountOptions{Email: "admin1@example.invalid", Roles: []string{"ADMIN"}})
			a2 := f.AddAccount(t, authdb.AccountOptions{Email: "admin2@example.invalid", Roles: []string{"ADMIN"}})
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			c1, c2 := f.DB.Connect(t), f.DB.Connect(t)
			block, release, entered := catalogBarrier(t, ctx)
			r1 := accountRepo(t, catalogHookDB{db: c1, beforeCommit: block})
			r2 := accountRepo(t, c2)
			cmd1 := accountCommand(t, f, a1, "first@example.invalid")
			cmd2 := accountCommand(t, f, a2, "second@example.invalid")
			if duplicate {
				cmd2.Input.CompanyEmail = " FIRST@EXAMPLE.INVALID "
			}
			type result struct {
				account org.AccountProjection
				err     error
			}
			done1, done2 := make(chan result, 1), make(chan result, 1)
			go func() { a, e := r1.CreateAccount(ctx, cmd1); done1 <- result{a, e} }()
			waitCatalogSignal(t, ctx, entered)
			go func() { a, e := r2.CreateAccount(ctx, cmd2); done2 <- result{a, e} }()
			if duplicate {
				observer, err := f.DB.Conn.Begin(ctx)
				authdb.Require(t, "observer email unique", err)
				defer observer.Rollback(context.Background())
				waitForAuthLock(t, ctx, observer, c1.PgConn().PID(), c2.PgConn().PID())
				authdb.Require(t, "đóng observer", observer.Rollback(ctx))
			}
			release()
			var first, second result
			select {
			case first = <-done1:
			case <-ctx.Done():
				t.Fatal("create thứ nhất timeout")
			}
			select {
			case second = <-done2:
			case <-ctx.Done():
				t.Fatal("create thứ hai timeout")
			}
			authdb.Require(t, "create thứ nhất commit", first.err)
			if duplicate {
				accountError(t, second.err, org.ErrEmailConflict)
				accountNoResult(t, second.account)
				if catalogCount(t, f, `SELECT count(*) FROM audit_events WHERE operation='CREATE_ACCOUNT'`) != 1 {
					t.Fatal("duplicate vẫn audit")
				}
			} else {
				authdb.Require(t, "create thứ hai commit", second.err)
				if first.account.EmployeeCode == second.account.EmployeeCode || first.account.UserID == second.account.UserID || first.account.EmployeeID == second.account.EmployeeID {
					t.Fatal("create đồng thời cấp trùng mã hoặc ID")
				}
			}
		})
	}
}

func TestAccountActorChangesWhileWaiting(t *testing.T) {
	for _, tc := range []struct {
		name, sql string
		want      error
	}{
		{"remove ADMIN", `DELETE FROM user_roles WHERE user_id=$1`, org.ErrForbidden},
		{"locked", `UPDATE users SET status='LOCKED' WHERE user_id=$1`, org.ErrUnauthenticated},
		{"inactive", `UPDATE employees SET employment_status='INACTIVE' WHERE user_id=$1`, org.ErrUnauthenticated},
		{"limited", `UPDATE users SET must_change_password=true WHERE user_id=$1`, org.ErrPasswordChangeRequired},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := authdb.NewFixture(t)
			a := f.AddAccount(t, authdb.AccountOptions{Email: "admin@example.invalid", Roles: []string{"ADMIN"}})
			cmd := accountCommand(t, f, a, "must-not-exist@example.invalid")
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			tx, err := f.DB.Conn.Begin(ctx)
			authdb.Require(t, "mutation actor", err)
			defer tx.Rollback(context.Background())
			var id int
			authdb.Require(t, "khóa user trước mutation", tx.QueryRow(ctx, `SELECT user_id FROM users WHERE user_id=$1 FOR UPDATE`, a.UserID).Scan(&id))
			authdb.Require(t, "khóa employee", tx.QueryRow(ctx, `SELECT employee_id FROM employees WHERE user_id=$1 FOR UPDATE`, a.UserID).Scan(&id))
			worker := f.DB.Connect(t)
			repo := accountRepo(t, worker)
			done := make(chan error, 1)
			go func() {
				got, e := repo.CreateAccount(ctx, cmd)
				if got.UserID != 0 {
					done <- fmt.Errorf("synthetic unexpected projection")
					return
				}
				done <- e
			}()
			waitForAuthLock(t, ctx, tx, f.DB.Conn.PgConn().PID(), worker.PgConn().PID())
			_, err = tx.Exec(ctx, tc.sql, a.UserID)
			authdb.Require(t, "mutation actor dưới khóa", err)
			authdb.Require(t, "commit actor mới", tx.Commit(ctx))
			accountError(t, catalogResult(t, ctx, done), tc.want)
			if catalogCount(t, f, `SELECT count(*) FROM users`) != 1 || catalogCount(t, f, `SELECT count(*) FROM audit_events`) != 0 {
				t.Fatal("create dùng quyền cũ")
			}
			var called bool
			authdb.Require(t, "không cấp số khi actor lỗi", f.DB.Conn.QueryRow(ctx, `SELECT is_called FROM employee_code_seq`).Scan(&called))
			if called {
				t.Fatal("actor lỗi không được đi tới sequence")
			}
		})
	}
}

func TestAccountCreateAndCatalogDeleteRealAdapters(t *testing.T) {
	for _, spec := range catalogTestSpecs {
		for _, order := range []string{"create first", "delete first", "delete first admin reference"} {
			t.Run(string(spec.kind)+"/"+order, func(t *testing.T) {
				f := authdb.NewFixture(t)
				hr := f.AddAccount(t, authdb.AccountOptions{Email: "hr@example.invalid", Roles: []string{"HR"}})
				admin := f.AddAccount(t, authdb.AccountOptions{Email: "admin@example.invalid", Roles: []string{"ADMIN"}})
				ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
				defer cancel()
				cs := catalogService(t, f.DB.Conn)
				item, err := cs.Create(ctx, org.CatalogCommand{ActorID: hr.UserID, Kind: spec.kind, Name: "Real create delete race"})
				authdb.Require(t, "catalog target", err)
				cmd := accountCommand(t, f, admin, "created@example.invalid")
				switch spec.kind {
				case org.Departments:
					cmd.Input.DepartmentID = &item.ID
				case org.Positions:
					cmd.Input.PositionID = item.ID
				case org.JobLevels:
					cmd.Input.JobLevelID = &item.ID
				}
				if strings.Contains(order, "admin reference") {
					_, err = f.DB.Conn.Exec(ctx, "UPDATE employees SET "+spec.id+"=$1 WHERE employee_id=$2", item.ID, admin.EmployeeID)
					authdb.Require(t, "admin references target", err)
				}
				creator, deleter := f.DB.Connect(t), f.DB.Connect(t)
				block, release, entered := catalogBarrier(t, ctx)
				createDone, deleteDone := make(chan error, 1), make(chan error, 1)
				var cr *postgres.AccountRepository
				var dr *org.CatalogService
				if order == "create first" {
					cr = accountRepo(t, catalogHookDB{db: creator, beforeCommit: block})
					dr = catalogService(t, deleter)
				} else {
					cr = accountRepo(t, creator)
					dr = catalogService(t, catalogHookDB{db: deleter, afterRow: func(sql string) {
						if strings.Contains(sql, "FROM public."+spec.table+" WHERE") && strings.HasSuffix(sql, "FOR UPDATE") {
							block()
						}
					}})
				}
				create := func() { _, e := cr.CreateAccount(ctx, cmd); createDone <- e }
				del := func() {
					deleteDone <- dr.Delete(ctx, org.CatalogCommand{ActorID: hr.UserID, Kind: spec.kind, ID: item.ID, ExpectedBeforeName: item.Name})
				}
				if order == "create first" {
					go create()
				} else {
					go del()
				}
				waitCatalogSignal(t, ctx, entered)
				if order == "create first" {
					go del()
				} else {
					go create()
				}
				observer, err := f.DB.Conn.Begin(ctx)
				authdb.Require(t, "observer catalog/create", err)
				defer observer.Rollback(context.Background())
				if order == "create first" {
					waitForAuthLock(t, ctx, observer, creator.PgConn().PID(), deleter.PgConn().PID())
				} else {
					waitForAuthLock(t, ctx, observer, deleter.PgConn().PID(), creator.PgConn().PID())
				}
				authdb.Require(t, "đóng observer", observer.Rollback(ctx))
				release()
				ce, de := catalogResult(t, ctx, createDone), catalogResult(t, ctx, deleteDone)
				switch order {
				case "create first", "delete first admin reference":
					authdb.Require(t, "create commit", ce)
					accountError(t, de, org.ErrCatalogInUse)
				case "delete first":
					authdb.Require(t, "delete commit", de)
					accountError(t, ce, org.ErrValidation)
				}
				if catalogCount(t, f, "SELECT count(*) FROM employees e LEFT JOIN public."+spec.table+" c ON c."+spec.id+"=e."+spec.id+" WHERE e."+spec.id+" IS NOT NULL AND c."+spec.id+" IS NULL") != 0 {
					t.Fatal("race gây orphan")
				}
			})
		}
	}
}
