//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	pc "ai-staffing/internal/service/profilechange"
	"ai-staffing/internal/testutil/authdb"
)

func TestProfileChangeDecisionConcurrentSingleWinner(t *testing.T) {
	for _, secondReject := range []bool{false, true} {
		t.Run(map[bool]string{false: "approve-approve", true: "approve-reject"}[secondReject], func(t *testing.T) {
			f := authdb.NewFixture(t)
			h1 := f.AddAccount(t, authdb.AccountOptions{Email: "hr1@example.invalid", Roles: []string{"HR"}})
			h2 := f.AddAccount(t, authdb.AccountOptions{Email: "hr2@example.invalid", Roles: []string{"HR"}})
			a := f.AddAccount(t, authdb.AccountOptions{Email: "self@example.invalid"})
			pending := pcPending(t, f, a, 1)
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			c1, c2 := f.DB.Connect(t), f.DB.Connect(t)
			block, release, entered := catalogBarrier(t, ctx)
			s1, s2 := pcService(t, catalogHookDB{db: c1, beforeCommit: block}), pcService(t, c2)
			first, second := make(chan pcResult, 1), make(chan pcResult, 1)
			go func() {
				d, e := s1.Approve(ctx, pc.DetailQuery{ActorID: h1.UserID, RequestID: pending.RequestID})
				first <- pcResult{d, e}
			}()
			waitCatalogSignal(t, ctx, entered)
			observer, e := f.DB.Conn.Begin(ctx)
			authdb.Require(t, "observer quyết định", e)
			defer observer.Rollback(context.Background())
			go func() {
				var d pc.Detail
				var e error
				if secondReject {
					d, e = s2.Reject(ctx, pc.RejectInput{ActorID: h2.UserID, RequestID: pending.RequestID, Reason: "Quyết định thứ hai"})
				} else {
					d, e = s2.Approve(ctx, pc.DetailQuery{ActorID: h2.UserID, RequestID: pending.RequestID})
				}
				second <- pcResult{d, e}
			}()
			waitForAuthLock(t, ctx, observer, c1.PgConn().PID(), c2.PgConn().PID())
			release()
			r1, r2 := receivePC(t, ctx, first), receivePC(t, ctx, second)
			authdb.Require(t, "quyết định đầu", r1.e)
			pcError(t, r2.e, pc.ErrAlreadyResolved)
			authdb.Require(t, "đóng observer", observer.Rollback(ctx))
			if r1.d.Status != "APPROVED" || r2.d.RequestID != 0 || pcDecisionCount(t, f) != 1 {
				t.Fatal("race quyết định sinh hai tác động hoặc success giả")
			}
		})
	}
}

func TestProfileChangeDecisionConcurrentPendingFields(t *testing.T) {
	for _, disjoint := range []bool{false, true} {
		t.Run(map[bool]string{false: "same-field", true: "different-fields"}[disjoint], func(t *testing.T) {
			f := authdb.NewFixture(t)
			hr := f.AddAccount(t, authdb.AccountOptions{Email: "hr@example.invalid", Roles: []string{"HR"}})
			a := f.AddAccount(t, authdb.AccountOptions{Email: "self@example.invalid"})
			one := pcPending(t, f, a, 1)
			in := pcInput(a, 2)
			if disjoint {
				in.Before, in.Proposed = pc.Values{"phone": nil}, pc.Values{"phone": pcText("12345")}
			}
			two, e := pcService(t, f.DB.Conn).Submit(t.Context(), in)
			authdb.Require(t, "pending thứ hai", e)
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			c1, c2 := f.DB.Connect(t), f.DB.Connect(t)
			block, release, entered := catalogBarrier(t, ctx)
			s1, s2 := pcService(t, catalogHookDB{db: c1, beforeCommit: block}), pcService(t, c2)
			first, second := make(chan pcResult, 1), make(chan pcResult, 1)
			go func() {
				d, e := s1.Approve(ctx, pc.DetailQuery{ActorID: hr.UserID, RequestID: one.RequestID})
				first <- pcResult{d, e}
			}()
			waitCatalogSignal(t, ctx, entered)
			observer, e := f.DB.Conn.Begin(ctx)
			authdb.Require(t, "observer pending", e)
			defer observer.Rollback(context.Background())
			go func() {
				d, e := s2.Approve(ctx, pc.DetailQuery{ActorID: hr.UserID, RequestID: two.RequestID})
				second <- pcResult{d, e}
			}()
			waitForAuthLock(t, ctx, observer, c1.PgConn().PID(), c2.PgConn().PID())
			release()
			authdb.Require(t, "approve pending đầu", receivePC(t, ctx, first).e)
			r := receivePC(t, ctx, second)
			authdb.Require(t, "đóng observer pending", observer.Rollback(ctx))
			wantCount := 1
			if disjoint {
				authdb.Require(t, "khác trường cùng chờ khóa", r.e)
				wantCount = 2
			} else {
				pcError(t, r.e, pc.ErrStale)
			}
			if pcDecisionCount(t, f) != wantCount {
				t.Fatal("pending cạnh tranh không giữ atomicity/stale theo field")
			}
		})
	}
}

func TestProfileChangeDecisionActorRecheckAfterLock(t *testing.T) {
	for _, tc := range []struct {
		name, mutation string
		want           error
	}{
		{"role", `DELETE FROM user_roles WHERE user_id=$1 AND role_code='HR'`, pc.ErrForbidden},
		{"locked", `UPDATE users SET status='LOCKED' WHERE user_id=$1`, pc.ErrUnauthenticated},
		{"inactive", `UPDATE employees SET employment_status='INACTIVE' WHERE user_id=$1`, pc.ErrUnauthenticated},
		{"limited", `UPDATE users SET must_change_password=true WHERE user_id=$1`, pc.ErrPasswordChangeRequired},
	} {
		for _, reject := range []bool{false, true} {
			t.Run(tc.name+map[bool]string{false: "/approve", true: "/reject"}[reject], func(t *testing.T) {
				f := authdb.NewFixture(t)
				hr := f.AddAccount(t, authdb.AccountOptions{Email: "hr@example.invalid", Roles: []string{"HR", "ADMIN"}})
				a := f.AddAccount(t, authdb.AccountOptions{Email: "self@example.invalid"})
				pending := pcPending(t, f, a, 1)
				ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
				defer cancel()
				holder := f.DB.Connect(t)
				tx, e := holder.Begin(ctx)
				authdb.Require(t, "tx actor", e)
				defer tx.Rollback(context.Background())
				var id int
				authdb.Require(t, "khóa HR user", tx.QueryRow(ctx, `SELECT user_id FROM users WHERE user_id=$1 FOR UPDATE`, hr.UserID).Scan(&id))
				worker := f.DB.Connect(t)
				s := pcService(t, worker)
				done := make(chan pcResult, 1)
				go func() {
					var d pc.Detail
					var e error
					if reject {
						d, e = s.Reject(ctx, pc.RejectInput{ActorID: hr.UserID, RequestID: pending.RequestID, Reason: "Lý do"})
					} else {
						d, e = s.Approve(ctx, pc.DetailQuery{ActorID: hr.UserID, RequestID: pending.RequestID})
					}
					done <- pcResult{d, e}
				}()
				waitForAuthLock(t, ctx, tx, holder.PgConn().PID(), worker.PgConn().PID())
				authdb.Require(t, "khóa HR employee", tx.QueryRow(ctx, `SELECT employee_id FROM employees WHERE user_id=$1 FOR UPDATE`, hr.UserID).Scan(&id))
				_, e = tx.Exec(ctx, tc.mutation, hr.UserID)
				authdb.Require(t, "đổi quyền/cờ dưới khóa", e)
				authdb.Require(t, "commit điều kiện mới", tx.Commit(ctx))
				r := receivePC(t, ctx, done)
				pcError(t, r.e, tc.want)
				if r.d.RequestID != 0 || pcDecisionCount(t, f) != 0 || catalogCount(t, f, `SELECT count(*) FROM employee_profile_change_requests WHERE request_id=$1 AND status='PENDING'`, pending.RequestID) != 1 || catalogCount(t, f, `SELECT count(*) FROM employees WHERE employee_id=$1 AND address IS NULL`, a.EmployeeID) != 1 {
					t.Fatal("actor recheck không chặn quyết định/official")
				}
			})
		}
	}
}

func TestProfileChangeDecisionAndSubmitShareLockOrder(t *testing.T) {
	f := authdb.NewFixture(t)
	hr := f.AddAccount(t, authdb.AccountOptions{Email: "hr@example.invalid", Roles: []string{"HR"}})
	a := f.AddAccount(t, authdb.AccountOptions{Email: "self@example.invalid"})
	pending := pcPending(t, f, a, 1)
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	c1, c2 := f.DB.Connect(t), f.DB.Connect(t)
	block, release, entered := catalogBarrier(t, ctx)
	s1, s2 := pcService(t, catalogHookDB{db: c1, beforeCommit: block}), pcService(t, c2)
	first, second := make(chan pcResult, 1), make(chan pcResult, 1)
	go func() {
		d, e := s1.Approve(ctx, pc.DetailQuery{ActorID: hr.UserID, RequestID: pending.RequestID})
		first <- pcResult{d, e}
	}()
	waitCatalogSignal(t, ctx, entered)
	observer, e := f.DB.Conn.Begin(ctx)
	authdb.Require(t, "observer quyết định/gửi", e)
	defer observer.Rollback(context.Background())
	in := pcInput(a, 2)
	in.Before, in.Proposed = pc.Values{"phone": nil}, pc.Values{"phone": pcText("12345")}
	go func() { d, e := s2.Submit(ctx, in); second <- pcResult{d, e} }()
	waitForAuthLock(t, ctx, observer, c1.PgConn().PID(), c2.PgConn().PID())
	release()
	authdb.Require(t, "quyết định cạnh tranh send", receivePC(t, ctx, first).e)
	authdb.Require(t, "send không đảo thứ tự khóa", receivePC(t, ctx, second).e)
	authdb.Require(t, "đóng observer quyết định/gửi", observer.Rollback(ctx))
	if pcDecisionCount(t, f) != 1 || catalogCount(t, f, `SELECT count(*) FROM employee_profile_change_requests`) != 2 || catalogCount(t, f, `SELECT count(*) FROM employees WHERE employee_id=$1 AND address IS NOT NULL AND phone IS NULL`, a.EmployeeID) != 1 {
		t.Fatal("submit/decision cạnh tranh mất dữ liệu hoặc tự ghi patch pending")
	}
}
