//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	org "ai-staffing/internal/service/organization"
	pc "ai-staffing/internal/service/profilechange"
	"ai-staffing/internal/testutil/authdb"
)

func TestAdministrativeProfileRaceActualApprove(t *testing.T) {
	for _, same := range []bool{true, false} {
		t.Run(map[bool]string{true: "same-field", false: "different-field"}[same], func(t *testing.T) {
			f := authdb.NewFixture(t)
			h := f.AddAccount(t, authdb.AccountOptions{Email: "hr@example.invalid", Roles: []string{"HR"}})
			a := f.AddAccount(t, authdb.AccountOptions{Email: "self@example.invalid"})
			in := pcInput(a, 1)
			in.Before, in.Proposed = pc.Values{"phone": nil}, pc.Values{"phone": pcText("pending")}
			pending, e := pcService(t, f.DB.Conn).Submit(t.Context(), in)
			authdb.Require(t, "submit thật", e)
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			c1, c2 := f.DB.Connect(t), f.DB.Connect(t)
			block, release, entered := catalogBarrier(t, ctx)
			one := administrativeService(t, catalogHookDB{db: c1, beforeCommit: block})
			two := pcService(t, c2)
			first, second := make(chan error, 1), make(chan error, 1)
			cmd := administrativePhone(h.UserID, a.EmployeeID, nil, pcText("direct"))
			if !same {
				cmd.ExpectedBeforeValues = org.AdministrativeValues{Address: administrativeText(nil)}
				cmd.Changes = org.AdministrativeValues{Address: administrativeText(pcText("Địa chỉ mới"))}
			}
			go func() { _, e := one.Edit(ctx, cmd); first <- e }()
			waitCatalogSignal(t, ctx, entered)
			observer, e := f.DB.Conn.Begin(ctx)
			authdb.Require(t, "observer", e)
			defer observer.Rollback(context.Background())
			go func() {
				_, e := two.Approve(ctx, pc.DetailQuery{ActorID: h.UserID, RequestID: pending.RequestID})
				second <- e
			}()
			waitForAuthLock(t, ctx, observer, c1.PgConn().PID(), c2.PgConn().PID())
			release()
			authdb.Require(t, "direct commit", catalogResult(t, ctx, first))
			e = catalogResult(t, ctx, second)
			if same {
				pcError(t, e, pc.ErrStale)
			} else {
				authdb.Require(t, "approve khác trường", e)
			}
			var status, phone string
			authdb.Require(t, "đối soát", f.DB.Conn.QueryRow(ctx, `SELECT r.status,e.phone FROM employee_profile_change_requests r JOIN employees e ON e.employee_id=r.employee_id WHERE request_id=$1`, pending.RequestID).Scan(&status, &phone))
			if same && (status != "PENDING" || phone != "direct") || !same && (status != "APPROVED" || phone != "pending") {
				t.Fatal("race approve ghi đè hoặc conflict giả")
			}
		})
	}
}

func TestAdministrativeProfileRaceApproveBeforeDirect(t *testing.T) {
	for _, same := range []bool{true, false} {
		t.Run(map[bool]string{true: "same-field", false: "different-field"}[same], func(t *testing.T) {
			f := authdb.NewFixture(t)
			h := f.AddAccount(t, authdb.AccountOptions{Email: "hr@example.invalid", Roles: []string{"HR"}})
			a := f.AddAccount(t, authdb.AccountOptions{Email: "self@example.invalid"})
			in := pcInput(a, 1)
			in.Before, in.Proposed = pc.Values{"phone": nil}, pc.Values{"phone": pcText("pending")}
			pending, e := pcService(t, f.DB.Conn).Submit(t.Context(), in)
			authdb.Require(t, "submit thật", e)
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			c1, c2 := f.DB.Connect(t), f.DB.Connect(t)
			block, release, entered := catalogBarrier(t, ctx)
			approve := pcService(t, catalogHookDB{db: c1, beforeCommit: block})
			direct := administrativeService(t, c2)
			first, second := make(chan error, 1), make(chan error, 1)
			cmd := administrativePhone(h.UserID, a.EmployeeID, nil, pcText("direct"))
			if !same {
				cmd.ExpectedBeforeValues = org.AdministrativeValues{Address: administrativeText(nil)}
				cmd.Changes = org.AdministrativeValues{Address: administrativeText(pcText("Địa chỉ mới"))}
			}
			go func() {
				_, e := approve.Approve(ctx, pc.DetailQuery{ActorID: h.UserID, RequestID: pending.RequestID})
				first <- e
			}()
			waitCatalogSignal(t, ctx, entered)
			observer, e := f.DB.Conn.Begin(ctx)
			authdb.Require(t, "observer", e)
			defer observer.Rollback(context.Background())
			go func() { _, e := direct.Edit(ctx, cmd); second <- e }()
			waitForAuthLock(t, ctx, observer, c1.PgConn().PID(), c2.PgConn().PID())
			release()
			authdb.Require(t, "approve commit", catalogResult(t, ctx, first))
			e = catalogResult(t, ctx, second)
			if same {
				requireCatalogError(t, e, org.ErrStale)
			} else {
				authdb.Require(t, "direct khác trường", e)
			}
			var status, phone string
			authdb.Require(t, "đối soát", f.DB.Conn.QueryRow(ctx, `SELECT r.status,e.phone FROM employee_profile_change_requests r JOIN employees e ON e.employee_id=r.employee_id WHERE request_id=$1`, pending.RequestID).Scan(&status, &phone))
			if status != "APPROVED" || phone != "pending" {
				t.Fatal("direct ghi đè kết luận")
			}
			want := 0
			if !same {
				want = 1
			}
			if administrativeCount(t, f, `SELECT count(*) FROM audit_events`) != want {
				t.Fatal("audit stale/khác trường sai")
			}
		})
	}
}

func TestAdministrativeProfileRaceEmailUnique(t *testing.T) {
	f := authdb.NewFixture(t)
	h1 := f.AddAccount(t, authdb.AccountOptions{Email: "hr1@example.invalid", Roles: []string{"HR"}})
	h2 := f.AddAccount(t, authdb.AccountOptions{Email: "hr2@example.invalid", Roles: []string{"HR"}})
	a := f.AddAccount(t, authdb.AccountOptions{Email: "a@example.invalid"})
	b := f.AddAccount(t, authdb.AccountOptions{Email: "b@example.invalid"})
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	c1, c2 := f.DB.Connect(t), f.DB.Connect(t)
	block, release, entered := catalogBarrier(t, ctx)
	first, second := make(chan error, 1), make(chan error, 1)
	cmd := func(actor, target int, before string) org.EditAdministrativeProfileCommand {
		return org.EditAdministrativeProfileCommand{ActorID: actor, EmployeeID: target, ExpectedBeforeValues: org.AdministrativeValues{CompanyEmail: administrativeText(pcText(before))}, Changes: org.AdministrativeValues{CompanyEmail: administrativeText(pcText("collision@example.invalid"))}}
	}
	s1, s2 := administrativeService(t, catalogHookDB{db: c1, beforeCommit: block}), administrativeService(t, c2)
	go func() { _, e := s1.Edit(ctx, cmd(h1.UserID, a.EmployeeID, "a@example.invalid")); first <- e }()
	waitCatalogSignal(t, ctx, entered)
	observer, e := f.DB.Conn.Begin(ctx)
	authdb.Require(t, "observer email", e)
	defer observer.Rollback(context.Background())
	go func() { _, e := s2.Edit(ctx, cmd(h2.UserID, b.EmployeeID, "b@example.invalid")); second <- e }()
	waitForAuthLock(t, ctx, observer, c1.PgConn().PID(), c2.PgConn().PID())
	release()
	authdb.Require(t, "email winner", catalogResult(t, ctx, first))
	requireCatalogError(t, catalogResult(t, ctx, second), org.ErrEmailConflict)
	if administrativeCount(t, f, `SELECT count(*) FROM audit_events`) != 1 {
		t.Fatal("email collision audit nửa chừng")
	}
}

func TestAdministrativeProfileRaceCatalogDelete(t *testing.T) {
	for _, deleteFirst := range []bool{true, false} {
		t.Run(map[bool]string{true: "delete-first", false: "edit-first"}[deleteFirst], func(t *testing.T) {
			f := authdb.NewFixture(t)
			h1 := f.AddAccount(t, authdb.AccountOptions{Email: "hr1@example.invalid", Roles: []string{"HR"}})
			h2 := f.AddAccount(t, authdb.AccountOptions{Email: "hr2@example.invalid", Roles: []string{"HR"}})
			a := f.AddAccount(t, authdb.AccountOptions{Email: "a@example.invalid"})
			var id int
			authdb.Require(t, "catalog", f.DB.Conn.QueryRow(t.Context(), `INSERT INTO departments(department_name) VALUES('Race') RETURNING department_id`).Scan(&id))
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			c1, c2 := f.DB.Connect(t), f.DB.Connect(t)
			block, release, entered := catalogBarrier(t, ctx)
			first, second := make(chan error, 1), make(chan error, 1)
			edit := func(db accountBeginner) error {
				_, e := administrativeService(t, db).Edit(ctx, org.EditAdministrativeProfileCommand{ActorID: h1.UserID, EmployeeID: a.EmployeeID, ExpectedBeforeValues: org.AdministrativeValues{DepartmentID: administrativeInt(nil)}, Changes: org.AdministrativeValues{DepartmentID: administrativeInt(&id)}})
				return e
			}
			del := func(db accountBeginner) error {
				return catalogService(t, db).Delete(ctx, org.CatalogCommand{ActorID: h2.UserID, Kind: org.Departments, ID: id, ExpectedBeforeName: "Race"})
			}
			hooked := catalogHookDB{db: c1, beforeCommit: block}
			go func() {
				if deleteFirst {
					first <- del(hooked)
				} else {
					first <- edit(hooked)
				}
			}()
			waitCatalogSignal(t, ctx, entered)
			observer, e := f.DB.Conn.Begin(ctx)
			authdb.Require(t, "observer catalog", e)
			defer observer.Rollback(context.Background())
			go func() {
				if deleteFirst {
					second <- edit(c2)
				} else {
					second <- del(c2)
				}
			}()
			waitForAuthLock(t, ctx, observer, c1.PgConn().PID(), c2.PgConn().PID())
			release()
			authdb.Require(t, "catalog winner", catalogResult(t, ctx, first))
			want := org.ErrCatalogInUse
			if deleteFirst {
				want = org.ErrCatalogNotFound
			}
			requireCatalogError(t, catalogResult(t, ctx, second), want)
		})
	}
}

func TestAdministrativeProfileActorRecheckUnderLock(t *testing.T) {
	for _, tc := range []struct {
		name, sql string
		want      error
	}{{"role", `DELETE FROM user_roles WHERE user_id=$1`, org.ErrForbidden}, {"limited", `UPDATE users SET must_change_password=true WHERE user_id=$1`, org.ErrPasswordChangeRequired}, {"locked", `UPDATE users SET status='LOCKED' WHERE user_id=$1`, org.ErrUnauthenticated}, {"inactive", `UPDATE employees SET employment_status='INACTIVE' WHERE user_id=$1`, org.ErrUnauthenticated}} {
		t.Run(tc.name, func(t *testing.T) {
			f := authdb.NewFixture(t)
			h := f.AddAccount(t, authdb.AccountOptions{Email: "hr@example.invalid", Roles: []string{"HR"}})
			a := f.AddAccount(t, authdb.AccountOptions{Email: "a@example.invalid"})
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			holder, worker := f.DB.Connect(t), f.DB.Connect(t)
			tx, e := holder.Begin(ctx)
			authdb.Require(t, "holder", e)
			defer tx.Rollback(context.Background())
			var id int
			authdb.Require(t, "khóa actor", tx.QueryRow(ctx, `SELECT user_id FROM users WHERE user_id=$1 FOR UPDATE`, h.UserID).Scan(&id))
			s := administrativeService(t, worker)
			done := make(chan error, 1)
			go func() {
				_, e := s.Edit(ctx, administrativePhone(h.UserID, a.EmployeeID, nil, pcText("new")))
				done <- e
			}()
			waitForAuthLock(t, ctx, tx, holder.PgConn().PID(), worker.PgConn().PID())
			authdb.Require(t, "employee actor", tx.QueryRow(ctx, `SELECT employee_id FROM employees WHERE user_id=$1 FOR UPDATE`, h.UserID).Scan(&id))
			_, e = tx.Exec(ctx, tc.sql, h.UserID)
			authdb.Require(t, "đổi quyền/cờ", e)
			authdb.Require(t, "commit quyền", tx.Commit(ctx))
			requireCatalogError(t, catalogResult(t, ctx, done), tc.want)
			if administrativeCount(t, f, `SELECT count(*) FROM audit_events`) != 0 {
				t.Fatal("mất quyền vẫn ghi")
			}
		})
	}
}
