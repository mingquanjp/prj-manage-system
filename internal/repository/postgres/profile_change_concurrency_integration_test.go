//go:build integration

package postgres_test

import (
	"context"
	"strings"
	"testing"
	"time"

	pc "ai-staffing/internal/service/profilechange"
	"ai-staffing/internal/testutil/authdb"
)

type pcResult struct {
	d pc.Detail
	e error
}

func receivePC(t *testing.T, ctx context.Context, ch <-chan pcResult) pcResult {
	t.Helper()
	select {
	case r := <-ch:
		return r
	case <-ctx.Done():
		t.Fatal("PCA race không kết thúc")
		return pcResult{}
	}
}

func TestProfileChangeSameKeyConcurrentAndRollbackFollower(t *testing.T) {
	for _, rollback := range []bool{false, true} {
		t.Run(map[bool]string{false: "commit", true: "rollback"}[rollback], func(t *testing.T) {
			f := authdb.NewFixture(t)
			a := f.AddAccount(t, authdb.AccountOptions{Email: "self@example.invalid", Roles: []string{"HR", "ADMIN"}})
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			c1, c2 := f.DB.Connect(t), f.DB.Connect(t)
			block, release, entered := catalogBarrier(t, ctx)
			var firstDB accountBeginner = catalogHookDB{db: c1, beforeCommit: block}
			if rollback {
				firstDB = catalogHookDB{db: accountCommitRollbackDB{c1}, beforeCommit: block}
			}
			s1, s2 := pcService(t, firstDB), pcService(t, c2)
			in := pcInput(a, 1)
			first, second := make(chan pcResult, 1), make(chan pcResult, 1)
			go func() { d, e := s1.Submit(ctx, in); first <- pcResult{d, e} }()
			waitCatalogSignal(t, ctx, entered)
			observer, e := f.DB.Conn.Begin(ctx)
			authdb.Require(t, "observer barrier", e)
			defer observer.Rollback(context.Background())
			go func() { d, e := s2.Submit(ctx, in); second <- pcResult{d, e} }()
			waitForAuthLock(t, ctx, observer, c1.PgConn().PID(), c2.PgConn().PID())
			release()
			r1, r2 := receivePC(t, ctx, first), receivePC(t, ctx, second)
			authdb.Require(t, "follower/replay", r2.e)
			if rollback {
				pcError(t, r1.e, pc.ErrPersistence)
			} else {
				authdb.Require(t, "writer đầu", r1.e)
				if r1.d.RequestID != r2.d.RequestID {
					t.Fatal("hai writer cùng key tạo hai request")
				}
			}
			authdb.Require(t, "đóng observer", observer.Rollback(ctx))
			for _, table := range []string{"employee_profile_change_requests", "notifications", "idempotency_records"} {
				if catalogCount(t, f, "SELECT count(*) FROM "+table) != 1 {
					t.Fatal("cùng key/rollback follower phải đúng một tác động")
				}
			}
		})
	}
}

func TestProfileChangeTwoHRActorsLockOrder(t *testing.T) {
	f := authdb.NewFixture(t)
	a := f.AddAccount(t, authdb.AccountOptions{Email: "one@example.invalid", Roles: []string{"HR", "DIRECTOR"}})
	b := f.AddAccount(t, authdb.AccountOptions{Email: "two@example.invalid", Roles: []string{"HR", "ADMIN"}})
	// Đảo liên kết bằng user tạm tổng hợp; không UPDATE identity ALWAYS.
	temp := f.AddAccount(t, authdb.AccountOptions{Email: "temporary@example.invalid", WithoutEmployee: true})
	pcExec(t, f, `UPDATE employees SET user_id=$1 WHERE employee_id=$2`, temp.UserID, a.EmployeeID)
	pcExec(t, f, `UPDATE employees SET user_id=$1 WHERE employee_id=$2`, a.UserID, b.EmployeeID)
	pcExec(t, f, `UPDATE employees SET user_id=$1 WHERE employee_id=$2`, b.UserID, a.EmployeeID)
	a.EmployeeID, b.EmployeeID = b.EmployeeID, a.EmployeeID

	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	s1, s2 := pcService(t, f.DB.Connect(t)), pcService(t, f.DB.Connect(t))
	start := make(chan struct{})
	done := make(chan pcResult, 2)
	go func() { <-start; d, e := s1.Submit(ctx, pcInput(a, 1)); done <- pcResult{d, e} }()
	go func() { <-start; d, e := s2.Submit(ctx, pcInput(b, 1)); done <- pcResult{d, e} }()
	close(start)
	for range 2 {
		authdb.Require(t, "hai actor HR cạnh tranh", receivePC(t, ctx, done).e)
	}
	if catalogCount(t, f, `SELECT count(*) FROM employee_profile_change_requests`) != 2 || catalogCount(t, f, `SELECT count(*) FROM notifications`) != 4 {
		t.Fatal("actor cũng recipient bị khóa lặp hoặc nhân notification theo role")
	}
}

func TestProfileChangeActorRecheckAndExpiryAfterLock(t *testing.T) {
	for _, tc := range []struct {
		name, mutation string
		want           error
		replay         bool
	}{
		{"locked", `UPDATE users SET status='LOCKED' WHERE user_id=$1`, pc.ErrUnauthenticated, false},
		{"inactive", `UPDATE employees SET employment_status='INACTIVE' WHERE user_id=$1`, pc.ErrUnauthenticated, false},
		{"limited", `UPDATE users SET must_change_password=true WHERE user_id=$1`, pc.ErrPasswordChangeRequired, false},
		{"expired while waiting", `UPDATE idempotency_records SET created_at=clock_timestamp()-interval '2 days' WHERE actor_id=$1`, pc.ErrKeyExpired, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := authdb.NewFixture(t)
			a := f.AddAccount(t, authdb.AccountOptions{Email: "self@example.invalid"})
			in := pcInput(a, 1)
			if tc.replay {
				_, e := pcService(t, f.DB.Conn).Submit(t.Context(), in)
				authdb.Require(t, "key trước chờ", e)
			}
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			holder := f.DB.Connect(t)
			tx, e := holder.Begin(ctx)
			authdb.Require(t, "tx actor", e)
			defer tx.Rollback(context.Background())
			var id int
			authdb.Require(t, "khóa user actor", tx.QueryRow(ctx, `SELECT user_id FROM users WHERE user_id=$1 FOR UPDATE`, a.UserID).Scan(&id))
			worker := f.DB.Connect(t)
			s := pcService(t, worker)
			done := make(chan pcResult, 1)
			go func() { d, e := s.Submit(ctx, in); done <- pcResult{d, e} }()
			waitForAuthLock(t, ctx, tx, holder.PgConn().PID(), worker.PgConn().PID())
			// Writer trạng thái giữ cùng thứ tự user rồi employee.
			authdb.Require(t, "khóa employee actor", tx.QueryRow(ctx, `SELECT employee_id FROM employees WHERE user_id=$1 FOR UPDATE`, a.UserID).Scan(&id))
			_, e = tx.Exec(ctx, tc.mutation, a.UserID)
			authdb.Require(t, "đổi điều kiện sau snapshot", e)
			if tc.replay {
				_, e = tx.Exec(ctx, `UPDATE idempotency_records SET expires_at=created_at+interval '24h' WHERE actor_id=$1`, a.UserID)
				authdb.Require(t, "TTL expired hợp lệ", e)
			}
			authdb.Require(t, "commit actor đổi", tx.Commit(ctx))
			r := receivePC(t, ctx, done)
			pcError(t, r.e, tc.want)
			if r.d.RequestID != 0 {
				t.Fatal("recheck lỗi vẫn trả result")
			}
			if !tc.replay {
				pcNoWrites(t, f)
			} else if catalogCount(t, f, `SELECT count(*) FROM employee_profile_change_requests`) != 1 {
				t.Fatal("key expired bị tái dùng")
			}
		})
	}
}

func TestProfileChangeRecipientRecheck(t *testing.T) {
	for _, mutation := range []string{
		`DELETE FROM user_roles WHERE user_id=$1 AND role_code='HR'`,
		`UPDATE users SET status='LOCKED' WHERE user_id=$1`,
		`UPDATE users SET must_change_password=true WHERE user_id=$1`,
		`UPDATE employees SET employment_status='INACTIVE' WHERE user_id=$1`,
	} {
		t.Run(strings.Fields(mutation)[0]+strings.Fields(mutation)[1], func(t *testing.T) {
			f := authdb.NewFixture(t)
			hr := f.AddAccount(t, authdb.AccountOptions{Email: "hr@example.invalid", Roles: []string{"HR", "ADMIN"}})
			a := f.AddAccount(t, authdb.AccountOptions{Email: "self@example.invalid"})
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			holder := f.DB.Connect(t)
			tx, e := holder.Begin(ctx)
			authdb.Require(t, "tx recipient", e)
			defer tx.Rollback(context.Background())
			var id int
			authdb.Require(t, "giữ HR user thấp hơn actor", tx.QueryRow(ctx, `SELECT user_id FROM users WHERE user_id=$1 FOR UPDATE`, hr.UserID).Scan(&id))
			worker := f.DB.Connect(t)
			s := pcService(t, worker)
			done := make(chan pcResult, 1)
			go func() { d, e := s.Submit(ctx, pcInput(a, 1)); done <- pcResult{d, e} }()
			waitForAuthLock(t, ctx, tx, holder.PgConn().PID(), worker.PgConn().PID())
			authdb.Require(t, "khóa employee recipient", tx.QueryRow(ctx, `SELECT employee_id FROM employees WHERE user_id=$1 FOR UPDATE`, hr.UserID).Scan(&id))
			_, e = tx.Exec(ctx, mutation, hr.UserID)
			authdb.Require(t, "đổi recipient", e)
			authdb.Require(t, "commit recipient", tx.Commit(ctx))
			r := receivePC(t, ctx, done)
			authdb.Require(t, "gửi vẫn thành công dù HR không còn hợp lệ", r.e)
			if catalogCount(t, f, `SELECT count(*) FROM notifications`) != 0 || catalogCount(t, f, `SELECT count(*) FROM employee_profile_change_requests`) != 1 {
				t.Fatal("recipient không được recheck dưới khóa")
			}
		})
	}
}

func TestProfileChangeNewHRDoesNotExtendSnapshot(t *testing.T) {
	f := authdb.NewFixture(t)
	a := f.AddAccount(t, authdb.AccountOptions{Email: "self@example.invalid"})
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	holder := f.DB.Connect(t)
	tx, e := holder.Begin(ctx)
	authdb.Require(t, "tx giữ actor", e)
	defer tx.Rollback(context.Background())
	var id int
	authdb.Require(t, "khóa actor", tx.QueryRow(ctx, `SELECT user_id FROM users WHERE user_id=$1 FOR UPDATE`, a.UserID).Scan(&id))
	worker := f.DB.Connect(t)
	s := pcService(t, worker)
	done := make(chan pcResult, 1)
	go func() { d, e := s.Submit(ctx, pcInput(a, 1)); done <- pcResult{d, e} }()
	waitForAuthLock(t, ctx, tx, holder.PgConn().PID(), worker.PgConn().PID())
	hr := f.AddAccount(t, authdb.AccountOptions{Email: "new-hr@example.invalid", Roles: []string{"HR"}})
	authdb.Require(t, "nhả actor", tx.Commit(ctx))
	authdb.Require(t, "submit snapshot cũ", receivePC(t, ctx, done).e)
	if catalogCount(t, f, `SELECT count(*) FROM notifications`) != 0 {
		t.Fatal("đã tự mở rộng tập recipient sau snapshot")
	}
	p, e := pcService(t, f.DB.Conn).List(ctx, pc.ListQuery{ActorID: hr.UserID, Scope: "review"})
	authdb.Require(t, "HR mới vẫn đọc queue hiện hành", e)
	if p.Total != 1 {
		t.Fatal("không notify không được làm mất queue")
	}
}

func TestProfileChangeActorGainingHRIsNotNewRecipient(t *testing.T) {
	f := authdb.NewFixture(t)
	a := f.AddAccount(t, authdb.AccountOptions{Email: "self@example.invalid"})
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	holder := f.DB.Connect(t)
	tx, e := holder.Begin(ctx)
	authdb.Require(t, "tx actor trước cấp HR", e)
	defer tx.Rollback(context.Background())
	var id int
	authdb.Require(t, "khóa actor", tx.QueryRow(ctx, `SELECT user_id FROM users WHERE user_id=$1 FOR UPDATE`, a.UserID).Scan(&id))
	worker := f.DB.Connect(t)
	s := pcService(t, worker)
	done := make(chan pcResult, 1)
	go func() { d, e := s.Submit(ctx, pcInput(a, 1)); done <- pcResult{d, e} }()
	waitForAuthLock(t, ctx, tx, holder.PgConn().PID(), worker.PgConn().PID())
	_, e = tx.Exec(ctx, `INSERT INTO user_roles(user_id,role_code) VALUES($1,'HR')`, a.UserID)
	authdb.Require(t, "cấp HR trong fixture theo giao thức", e)
	authdb.Require(t, "commit HR mới", tx.Commit(ctx))
	authdb.Require(t, "submit", receivePC(t, ctx, done).e)
	if catalogCount(t, f, `SELECT count(*) FROM notifications`) != 0 {
		t.Fatal("actor có trong bộ khóa không đồng nghĩa recipient HR đã có ở snapshot")
	}
}
