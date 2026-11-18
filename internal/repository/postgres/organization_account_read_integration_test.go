//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"ai-staffing/internal/repository/postgres"
	org "ai-staffing/internal/service/organization"
	"ai-staffing/internal/testutil/authdb"
)

func accountReadService(t *testing.T, db accountBeginner) *org.AccountReadService {
	t.Helper()
	repo, err := postgres.NewAccountReadRepository(db, 10*time.Second)
	authdb.Require(t, "khởi tạo adapter đọc tài khoản", err)
	return org.NewAccountReadService(repo)
}

func TestAccountListSnapshotLiteralRolesAndProjection(t *testing.T) {
	f := authdb.NewFixture(t)
	admin := f.AddAccount(t, authdb.AccountOptions{Email: "actor@example.invalid", Roles: []string{"ADMIN"}})
	target := f.AddAccount(t, authdb.AccountOptions{Email: "target@example.invalid", Roles: []string{"HR", "ADMIN", "DIRECTOR"}, UserStatus: "LOCKED"})
	_, err := f.DB.Conn.Exec(t.Context(), `UPDATE employees SET fullname=$1 WHERE employee_id=$2`, `literal%_\value`, target.EmployeeID)
	authdb.Require(t, "dữ liệu literal", err)
	service := accountReadService(t, f.DB.Conn)
	for _, query := range []string{"%", "_", `\`, "LITERAL%_"} {
		page, err := service.List(t.Context(), org.ListAccountsQuery{ActorID: admin.UserID, Query: query})
		authdb.Require(t, "list literal", err)
		if page.Total != 1 || len(page.Items) != 1 || page.Items[0].UserID != target.UserID || len(page.Items[0].SystemRoles) != 3 ||
			page.Items[0].SystemRoles[0] != "ADMIN" || page.Items[0].SystemRoles[1] != "HR" || page.Items[0].SystemRoles[2] != "DIRECTOR" {
			t.Fatal("literal/roles/status projection sai")
		}
	}
	worker := f.DB.Connect(t)
	service = accountReadService(t, catalogHookDB{db: f.DB.Conn, hook: func() {
		_, err := worker.Exec(t.Context(), `UPDATE employees SET fullname='changed after count' WHERE employee_id=$1`, target.EmployeeID)
		authdb.Require(t, "mutation giữa count/items", err)
	}})
	page, err := service.List(t.Context(), org.ListAccountsQuery{ActorID: admin.UserID, Query: "%"})
	authdb.Require(t, "list một snapshot", err)
	if page.Total != 1 || len(page.Items) != 1 || page.Items[0].Fullname != `literal%_\value` {
		t.Fatal("count/items khác snapshot")
	}
	page, err = accountReadService(t, f.DB.Conn).List(t.Context(), org.ListAccountsQuery{ActorID: admin.UserID, Page: 999})
	authdb.Require(t, "page rỗng", err)
	if page.Total != 2 || len(page.Items) != 0 || page.Items == nil {
		t.Fatal("page rỗng mất total")
	}
}

func TestAccountReadsRecheckCurrentActor(t *testing.T) {
	for _, tc := range []struct {
		name string
		opts authdb.AccountOptions
		want error
	}{
		{"HR", authdb.AccountOptions{Roles: []string{"HR"}, LegacyRole: "ADMIN"}, org.ErrForbidden},
		{"limited", authdb.AccountOptions{Roles: []string{"ADMIN"}, MustChangePassword: true}, org.ErrPasswordChangeRequired},
		{"locked", authdb.AccountOptions{Roles: []string{"ADMIN"}, UserStatus: "LOCKED"}, org.ErrUnauthenticated},
		{"inactive", authdb.AccountOptions{Roles: []string{"ADMIN"}, EmploymentStatus: "INACTIVE"}, org.ErrUnauthenticated},
		{"no employee", authdb.AccountOptions{Roles: []string{"ADMIN"}, WithoutEmployee: true}, org.ErrUnauthenticated},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := authdb.NewFixture(t)
			tc.opts.Email = "reader@example.invalid"
			a := f.AddAccount(t, tc.opts)
			s := accountReadService(t, f.DB.Conn)
			if _, err := s.List(t.Context(), org.ListAccountsQuery{ActorID: a.UserID}); !errors.Is(err, tc.want) {
				t.Fatal("list sai loại lỗi quyền")
			}
			if _, err := s.Preview(t.Context(), a.UserID); !errors.Is(err, tc.want) {
				t.Fatal("preview sai loại lỗi quyền")
			}
		})
	}
}

func TestAccountPreviewDoesNotConsumeAndSurvivesRepositoryRestart(t *testing.T) {
	f := authdb.NewFixture(t)
	admin := f.AddAccount(t, authdb.AccountOptions{Email: "actor@example.invalid", Roles: []string{"ADMIN"}})
	s := accountReadService(t, f.DB.Conn)
	for range 2 {
		code, err := s.Preview(t.Context(), admin.UserID)
		authdb.Require(t, "preview trước create", err)
		if code != "EMP000001" {
			t.Fatal("preview tiêu số")
		}
	}
	created, err := accountRepo(t, f.DB.Conn).CreateAccount(t.Context(), accountCommand(t, f, admin, "created@example.invalid"))
	authdb.Require(t, "create sau preview", err)
	if created.EmployeeCode != "EMP000001" {
		t.Fatal("preview giữ chỗ hoặc tiêu số")
	}
	code, err := accountReadService(t, f.DB.Conn).Preview(t.Context(), admin.UserID)
	authdb.Require(t, "preview sau khởi tạo lại adapter", err)
	if code != "EMP000002" {
		t.Fatal("preview reset sequence")
	}
}

func TestAccountListRejectsOrphanUser(t *testing.T) {
	f := authdb.NewFixture(t)
	admin := f.AddAccount(t, authdb.AccountOptions{Email: "actor@example.invalid", Roles: []string{"ADMIN"}})
	f.AddAccount(t, authdb.AccountOptions{Email: "orphan@example.invalid", WithoutEmployee: true})
	_, err := accountReadService(t, f.DB.Conn).List(t.Context(), org.ListAccountsQuery{ActorID: admin.UserID})
	if !errors.Is(err, org.ErrPersistence) {
		t.Fatal("list âm thầm bỏ user thiếu employee")
	}
}

func TestAccountPreviewWhileCreateWaitsForCommit(t *testing.T) {
	f := authdb.NewFixture(t)
	admin := f.AddAccount(t, authdb.AccountOptions{Email: "preview-race@example.invalid", Roles: []string{"ADMIN"}})
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	worker := f.DB.Connect(t)
	block, release, entered := catalogBarrier(t, ctx)
	command := accountCommand(t, f, admin, "preview-created@example.invalid")
	repo := accountRepo(t, catalogHookDB{db: worker, beforeCommit: block})
	done := make(chan error, 1)
	go func() { _, err := repo.CreateAccount(ctx, command); done <- err }()
	waitCatalogSignal(t, ctx, entered)
	code, err := accountReadService(t, f.DB.Conn).Preview(ctx, admin.UserID)
	authdb.Require(t, "preview trong lúc create chờ commit", err)
	if code != "EMP000002" {
		t.Fatal("Preview giữ chỗ hoặc không phản ánh sequence đã cấp")
	}
	release()
	select {
	case err := <-done:
		authdb.Require(t, "commit create sau preview", err)
	case <-ctx.Done():
		t.Fatal("Create không kết thúc sau barrier")
	}
	code, err = accountReadService(t, f.DB.Conn).Preview(ctx, admin.UserID)
	authdb.Require(t, "preview sau commit", err)
	if code != "EMP000002" {
		t.Fatal("Preview đã tiêu số")
	}
}

func TestAccountPreviewRejectsSequenceExhaustion(t *testing.T) {
	f := authdb.NewFixture(t)
	admin := f.AddAccount(t, authdb.AccountOptions{Email: "overflow@example.invalid", Roles: []string{"ADMIN"}})
	_, err := f.DB.Conn.Exec(t.Context(), `SELECT setval('employee_code_seq',$1,true)`, int64(math.MaxInt64))
	authdb.Require(t, "đặt boundary trên DB case hy sinh", err)
	if _, err = accountReadService(t, f.DB.Conn).Preview(t.Context(), admin.UserID); !errors.Is(err, org.ErrPersistence) {
		t.Fatal("Sequence hết dải phải báo lỗi dữ liệu")
	}
}
