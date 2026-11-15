//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"ai-staffing/internal/repository/postgres"
	org "ai-staffing/internal/service/organization"
	"ai-staffing/internal/testutil/authdb"
	"github.com/jackc/pgx/v5"
)

func catalogBarrier(t *testing.T, ctx context.Context) (func(), func(), <-chan struct{}) {
	t.Helper()
	entered := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	t.Cleanup(unblock)
	return func() {
		close(entered)
		select {
		case <-release:
		case <-ctx.Done():
		}
	}, unblock, entered
}
func waitCatalogSignal(t *testing.T, ctx context.Context, signal <-chan struct{}) {
	t.Helper()
	select {
	case <-signal:
	case <-ctx.Done():
		t.Fatal("barrier catalog không tới trước timeout")
	}
}
func catalogResult(t *testing.T, ctx context.Context, ch <-chan error) error {
	t.Helper()
	select {
	case e := <-ch:
		return e
	case <-ctx.Done():
		t.Fatal("catalog race không kết thúc")
		return ctx.Err()
	}
}

func TestCatalogConcurrentDuplicateAndStale(t *testing.T) {
	f := authdb.NewFixture(t)
	hr1 := catalogActor(t, f, "HR")
	hr2 := f.AddAccount(t, authdb.AccountOptions{Email: "hr2@example.invalid", Roles: []string{"HR"}})
	for _, spec := range catalogTestSpecs {
		t.Run(string(spec.kind), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			c1, c2 := f.DB.Connect(t), f.DB.Connect(t)
			observer, err := f.DB.Conn.Begin(ctx)
			authdb.Require(t, "observer khóa", err)
			defer observer.Rollback(context.Background())
			block, release, entered := catalogBarrier(t, ctx)
			s1 := catalogService(t, catalogHookDB{db: c1, beforeCommit: block})
			s2 := catalogService(t, c2)
			first, second := make(chan error, 1), make(chan error, 1)
			go func() {
				_, e := s1.Create(ctx, org.CatalogCommand{ActorID: hr1.UserID, Kind: spec.kind, Name: "Concurrent"})
				first <- e
			}()
			waitCatalogSignal(t, ctx, entered)
			go func() {
				_, e := s2.Create(ctx, org.CatalogCommand{ActorID: hr2.UserID, Kind: spec.kind, Name: "concurrent"})
				second <- e
			}()
			waitForAuthLock(t, ctx, observer, c1.PgConn().PID(), c2.PgConn().PID())
			release()
			authdb.Require(t, "create thứ nhất", catalogResult(t, ctx, first))
			requireCatalogError(t, catalogResult(t, ctx, second), org.ErrDuplicateName)
			authdb.Require(t, "đóng observer", observer.Rollback(ctx))
			var id int
			authdb.Require(t, "đọc ID tạo cạnh tranh", f.DB.Conn.QueryRow(ctx, "SELECT "+spec.id+" FROM public."+spec.table+" WHERE "+spec.name+"='Concurrent'").Scan(&id))
			// Hai rename thật cùng đọc giá trị nền; lần sau phải thấy tên đã commit.
			block, release, entered = catalogBarrier(t, ctx)
			s1 = catalogService(t, catalogHookDB{db: c1, beforeCommit: block})
			observer, err = f.DB.Conn.Begin(ctx)
			authdb.Require(t, "observer rename", err)
			defer observer.Rollback(context.Background())
			go func() {
				_, e := s1.Rename(ctx, org.CatalogCommand{ActorID: hr1.UserID, Kind: spec.kind, ID: id, Name: "First", ExpectedBeforeName: "Concurrent"})
				first <- e
			}()
			waitCatalogSignal(t, ctx, entered)
			go func() {
				_, e := s2.Rename(ctx, org.CatalogCommand{ActorID: hr2.UserID, Kind: spec.kind, ID: id, Name: "Second", ExpectedBeforeName: "Concurrent"})
				second <- e
			}()
			waitForAuthLock(t, ctx, observer, c1.PgConn().PID(), c2.PgConn().PID())
			release()
			authdb.Require(t, "rename thứ nhất", catalogResult(t, ctx, first))
			requireCatalogError(t, catalogResult(t, ctx, second), org.ErrStale)
			authdb.Require(t, "đóng observer rename", observer.Rollback(ctx))
			if catalogCount(t, f, "SELECT count(*) FROM public.audit_events WHERE entity_type=$1", spec.entity) != 2 {
				t.Fatal("race thất bại không được audit")
			}
		})
	}
}

func TestCatalogActorRecheckAfterLock(t *testing.T) {
	for _, tc := range []struct {
		name, sql string
		want      error
	}{
		{"mất HR", `DELETE FROM public.user_roles WHERE user_id=$1`, org.ErrForbidden},
		{"khóa user", `UPDATE public.users SET status='LOCKED' WHERE user_id=$1`, org.ErrUnauthenticated},
		{"employee inactive", `UPDATE public.employees SET employment_status='INACTIVE' WHERE user_id=$1`, org.ErrUnauthenticated},
		{"buộc đổi", `UPDATE public.users SET must_change_password=true WHERE user_id=$1`, org.ErrPasswordChangeRequired},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := authdb.NewFixture(t)
			a := catalogActor(t, f, "HR")
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			tx, err := f.DB.Conn.Begin(ctx)
			authdb.Require(t, "mutation quyền bắt đầu", err)
			defer tx.Rollback(context.Background())
			var id int
			authdb.Require(t, "khóa user trước đổi quyền", tx.QueryRow(ctx, `SELECT user_id FROM public.users WHERE user_id=$1 FOR UPDATE`, a.UserID).Scan(&id))
			authdb.Require(t, "khóa employee theo giao thức", tx.QueryRow(ctx, `SELECT employee_id FROM public.employees WHERE user_id=$1 FOR UPDATE`, a.UserID).Scan(&id))
			worker := f.DB.Connect(t)
			s := catalogService(t, worker)
			done := make(chan error, 1)
			go func() {
				_, e := s.Create(ctx, org.CatalogCommand{ActorID: a.UserID, Kind: org.Departments, Name: "Must not exist"})
				done <- e
			}()
			waitForAuthLock(t, ctx, tx, f.DB.Conn.PgConn().PID(), worker.PgConn().PID())
			_, err = tx.Exec(ctx, tc.sql, a.UserID)
			authdb.Require(t, "đổi quyền dưới khóa", err)
			authdb.Require(t, "commit quyền mới", tx.Commit(ctx))
			requireCatalogError(t, catalogResult(t, ctx, done), tc.want)
			if catalogCount(t, f, `SELECT count(*) FROM public.departments`) != 0 || catalogCount(t, f, `SELECT count(*) FROM public.audit_events`) != 0 {
				t.Fatal("quyền thay đổi phải từ chối trước ghi")
			}
		})
	}
}

func TestCatalogDeleteWithAdminReferenceRace(t *testing.T) {
	for _, spec := range catalogTestSpecs {
		for _, referenced := range []bool{true, false} {
			name := string(spec.kind) + "/unreferenced"
			if referenced {
				name = string(spec.kind) + "/admin-references-target"
			}
			t.Run(name, func(t *testing.T) {
				f := authdb.NewFixture(t)
				hr := catalogActor(t, f, "HR")
				admin := f.AddAccount(t, authdb.AccountOptions{Email: "admin-race@example.invalid", Roles: []string{"ADMIN"}})
				ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
				defer cancel()
				s := catalogService(t, f.DB.Conn)
				item, err := s.Create(ctx, org.CatalogCommand{ActorID: hr.UserID, Kind: spec.kind, Name: "Delete race"})
				authdb.Require(t, "tạo mục race xóa", err)
				if referenced {
					_, err = f.DB.Conn.Exec(ctx, "UPDATE public.employees SET "+spec.id+"=$1 WHERE employee_id=$2", item.ID, admin.EmployeeID)
					authdb.Require(t, "Admin tham chiếu catalog", err)
				}
				adminConn, hrConn := f.DB.Connect(t), f.DB.Connect(t)
				adminTx, err := adminConn.Begin(ctx)
				authdb.Require(t, "fake create bắt đầu", err)
				defer adminTx.Rollback(context.Background())
				var id int
				authdb.Require(t, "fake create khóa user", adminTx.QueryRow(ctx, `SELECT user_id FROM public.users WHERE user_id=$1 FOR UPDATE`, admin.UserID).Scan(&id))
				authdb.Require(t, "fake create khóa employee", adminTx.QueryRow(ctx, `SELECT employee_id FROM public.employees WHERE user_id=$1 FOR UPDATE`, admin.UserID).Scan(&id))
				block, release, entered := catalogBarrier(t, ctx)
				hs := catalogService(t, catalogHookDB{db: hrConn, afterRow: func(sql string) {
					if strings.Contains(sql, "FROM public."+spec.table+" WHERE") && strings.HasSuffix(sql, "FOR UPDATE") {
						block()
					}
				}})
				deletion := make(chan error, 1)
				go func() {
					deletion <- hs.Delete(ctx, org.CatalogCommand{ActorID: hr.UserID, Kind: spec.kind, ID: item.ID, ExpectedBeforeName: item.Name})
				}()
				waitCatalogSignal(t, ctx, entered)
				// Chỉ mô phỏng bước khóa catalog của create, chưa chứng nhận adapter OAP-T09.
				create := make(chan error, 1)
				go func() {
					var locked int
					e := adminTx.QueryRow(ctx, "SELECT "+spec.id+" FROM public."+spec.table+" WHERE "+spec.id+"=$1 FOR KEY SHARE", item.ID).Scan(&locked)
					create <- e
				}()
				observer, err := f.DB.Conn.Begin(ctx)
				authdb.Require(t, "observer delete/create", err)
				defer observer.Rollback(context.Background())
				waitForAuthLock(t, ctx, observer, hrConn.PgConn().PID(), adminConn.PgConn().PID())
				release()
				if referenced {
					requireCatalogError(t, catalogResult(t, ctx, deletion), org.ErrCatalogInUse)
					authdb.Require(t, "fake create nhận KEY SHARE", catalogResult(t, ctx, create))
				} else {
					authdb.Require(t, "delete không ref", catalogResult(t, ctx, deletion))
					if !errors.Is(catalogResult(t, ctx, create), pgx.ErrNoRows) {
						t.Fatal("create phải thấy catalog đã mất")
					}
				}
				authdb.Require(t, "đóng fake create", adminTx.Rollback(ctx))
				authdb.Require(t, "đóng observer", observer.Rollback(ctx))
				var orphan int
				authdb.Require(t, "không orphan", f.DB.Conn.QueryRow(ctx, "SELECT count(*) FROM public.employees e LEFT JOIN public."+spec.table+" c ON c."+spec.id+"=e."+spec.id+" WHERE e."+spec.id+" IS NOT NULL AND c."+spec.id+" IS NULL").Scan(&orphan))
				if orphan != 0 {
					t.Fatal("race gây orphan")
				}
			})
		}
	}
}

func TestCatalogDeleteWaitsForNewReference(t *testing.T) {
	for _, spec := range catalogTestSpecs {
		t.Run(string(spec.kind), func(t *testing.T) {
			f := authdb.NewFixture(t)
			hr := catalogActor(t, f, "HR")
			admin := f.AddAccount(t, authdb.AccountOptions{Email: "create-race@example.invalid", Roles: []string{"ADMIN"}})
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			s := catalogService(t, f.DB.Conn)
			item, err := s.Create(ctx, org.CatalogCommand{ActorID: hr.UserID, Kind: spec.kind, Name: "New reference"})
			authdb.Require(t, "catalog cho create thắng", err)
			creator := f.DB.Connect(t)
			tx, err := creator.Begin(ctx)
			authdb.Require(t, "fake create transaction", err)
			defer tx.Rollback(context.Background())
			var id, position int
			authdb.Require(t, "khóa actor user", tx.QueryRow(ctx, `SELECT user_id FROM public.users WHERE user_id=$1 FOR UPDATE`, admin.UserID).Scan(&id))
			authdb.Require(t, "khóa actor employee", tx.QueryRow(ctx, `SELECT position_id FROM public.employees WHERE user_id=$1 FOR UPDATE`, admin.UserID).Scan(&position))
			// Thứ tự department → position → job_level giống plan create; chỉ khóa mục có dùng.
			var department, level *int
			if spec.kind == org.Departments {
				department = &item.ID
				authdb.Require(t, "KEY SHARE department", tx.QueryRow(ctx, `SELECT department_id FROM public.departments WHERE department_id=$1 FOR KEY SHARE`, item.ID).Scan(&id))
			}
			if spec.kind == org.Positions {
				position = item.ID
			}
			authdb.Require(t, "KEY SHARE position", tx.QueryRow(ctx, `SELECT position_id FROM public.positions WHERE position_id=$1 FOR KEY SHARE`, position).Scan(&id))
			if spec.kind == org.JobLevels {
				level = &item.ID
				authdb.Require(t, "KEY SHARE job_level", tx.QueryRow(ctx, `SELECT job_level_id FROM public.job_levels WHERE job_level_id=$1 FOR KEY SHARE`, item.ID).Scan(&id))
			}
			// Chưa có create adapter M2: fixture tổng hợp ghi FK trên schema M1, không cấp mã sản phẩm.
			_, err = tx.Exec(ctx, `INSERT INTO public.users(user_id,username,password_hash,system_role,status,must_change_password) OVERRIDING SYSTEM VALUE VALUES (900,'race-new-user',$1,'EMPLOYEE','ACTIVE',true)`, f.PasswordHash)
			authdb.Require(t, "fake create user", err)
			_, err = tx.Exec(ctx, `INSERT INTO public.employees(user_id,employee_code,fullname,email,position_id,department_id,job_level_id) VALUES(900,'RACE-NEW','Synthetic new employee','new-reference@example.invalid',$1,$2,$3)`, position, department, level)
			authdb.Require(t, "fake create employee FK", err)
			worker := f.DB.Connect(t)
			hs := catalogService(t, worker)
			done := make(chan error, 1)
			go func() {
				done <- hs.Delete(ctx, org.CatalogCommand{ActorID: hr.UserID, Kind: spec.kind, ID: item.ID, ExpectedBeforeName: item.Name})
			}()
			observer, err := f.DB.Conn.Begin(ctx)
			authdb.Require(t, "observer reference", err)
			defer observer.Rollback(context.Background())
			waitForAuthLock(t, ctx, observer, creator.PgConn().PID(), worker.PgConn().PID())
			authdb.Require(t, "commit fake create", tx.Commit(ctx))
			requireCatalogError(t, catalogResult(t, ctx, done), org.ErrCatalogInUse)
			authdb.Require(t, "đóng observer", observer.Rollback(ctx))
			if catalogCount(t, f, "SELECT count(*) FROM public."+spec.table+" WHERE "+spec.id+"=$1", item.ID) != 1 {
				t.Fatal("không được xóa mục vừa có tham chiếu commit")
			}
		})
	}
}

func TestCatalogLockTimeoutRollsBack(t *testing.T) {
	f := authdb.NewFixture(t)
	a := catalogActor(t, f, "HR")
	ctx := t.Context()
	s := catalogService(t, f.DB.Conn)
	item, err := s.Create(ctx, org.CatalogCommand{ActorID: a.UserID, Kind: org.Departments, Name: "Locked"})
	authdb.Require(t, "mục bị khóa", err)
	tx, err := f.DB.Conn.Begin(ctx)
	authdb.Require(t, "khóa fixture", err)
	defer tx.Rollback(context.Background())
	var id int
	authdb.Require(t, "giữ khóa catalog", tx.QueryRow(ctx, `SELECT department_id FROM public.departments WHERE department_id=$1 FOR UPDATE`, item.ID).Scan(&id))
	worker := f.DB.Connect(t)
	r, err := postgres.NewCatalogRepository(worker, 50*time.Millisecond)
	authdb.Require(t, "repository timeout", err)
	_, err = org.NewCatalogService(r).Rename(ctx, org.CatalogCommand{ActorID: a.UserID, Kind: org.Departments, ID: item.ID, Name: "Must rollback", ExpectedBeforeName: item.Name})
	requireCatalogError(t, err, org.ErrUnavailable)
	authdb.Require(t, "nhả khóa fixture", tx.Rollback(ctx))
	// pgx được phép đóng connection khi hủy I/O; nếu còn mở thì transaction phải idle.
	if !worker.IsClosed() && worker.PgConn().TxStatus() != 'I' {
		t.Fatal("connection còn mở phải idle sau rollback")
	}

	if catalogCount(t, f, `SELECT count(*) FROM public.audit_events`) != 1 {
		t.Fatal("timeout không được audit thành công")
	}
}
