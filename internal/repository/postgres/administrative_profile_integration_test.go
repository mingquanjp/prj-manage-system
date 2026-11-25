//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"ai-staffing/internal/repository/postgres"
	"ai-staffing/internal/service/auth"
	org "ai-staffing/internal/service/organization"
	"ai-staffing/internal/testutil/authdb"
	"github.com/jackc/pgx/v5"
)

func administrativeService(t *testing.T, db interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}) *org.AdministrativeProfileService {
	t.Helper()
	r, e := postgres.NewAdministrativeProfileRepository(db, 5*time.Second)
	authdb.Require(t, "khởi tạo HR profile", e)
	return org.NewAdministrativeProfileService(r)
}
func administrativeText(s *string) org.AdministrativeField[string] {
	return org.AdministrativeField[string]{Present: true, Value: s}
}
func administrativeInt(s *int) org.AdministrativeField[int] {
	return org.AdministrativeField[int]{Present: true, Value: s}
}
func administrativePhone(actor, target int, before, after *string) org.EditAdministrativeProfileCommand {
	return org.EditAdministrativeProfileCommand{ActorID: actor, EmployeeID: target, ExpectedBeforeValues: org.AdministrativeValues{Phone: administrativeText(before)}, Changes: org.AdministrativeValues{Phone: administrativeText(after)}}
}
func administrativeCount(t *testing.T, f *authdb.Fixture, sql string) int {
	t.Helper()
	var n int
	authdb.Require(t, "đếm tác động HR", f.DB.Conn.QueryRow(t.Context(), sql).Scan(&n))
	return n
}

func TestAdministrativeProfileProjectionMutationAndNoop(t *testing.T) {
	f := authdb.NewFixture(t)
	hr := f.AddAccount(t, authdb.AccountOptions{Email: "hr@example.invalid", Roles: []string{"HR", "ADMIN"}})
	a := f.AddAccount(t, authdb.AccountOptions{Email: "  Legacy@Example.invalid  ", EmploymentStatus: "INACTIVE"})
	s := administrativeService(t, f.DB.Conn)
	ctx := t.Context()
	q := org.AdministrativeProfileQuery{ActorID: hr.UserID, EmployeeID: a.EmployeeID}
	p, e := s.Read(ctx, q)
	authdb.Require(t, "đọc inactive khác ID", e)
	if p.CompanyEmail != "legacy@example.invalid" || p.EmploymentStatus != "INACTIVE" || p.Phone != nil || p.Address != nil {
		t.Fatal("projection không đúng")
	}
	long := strings.Repeat("a", 2100)
	_, e = f.DB.Conn.Exec(ctx, `UPDATE employees SET address=$1 WHERE employee_id=$2`, long, a.EmployeeID)
	authdb.Require(t, "legacy dài", e)
	c := org.EditAdministrativeProfileCommand{ActorID: hr.UserID, EmployeeID: a.EmployeeID, ExpectedBeforeValues: org.AdministrativeValues{CompanyEmail: administrativeText(pcText("legacy@example.invalid"))}, Changes: org.AdministrativeValues{CompanyEmail: administrativeText(pcText(" LEGACY@example.invalid "))}}
	p, e = s.Edit(ctx, c)
	authdb.Require(t, "canonical noop", e)
	if p.Address == nil || *p.Address != long || administrativeCount(t, f, `SELECT count(*) FROM audit_events`) != 0 {
		t.Fatal("noop audit hoặc cắt legacy")
	}
	var raw string
	authdb.Require(t, "giữ email raw", f.DB.Conn.QueryRow(ctx, `SELECT email FROM employees WHERE employee_id=$1`, a.EmployeeID).Scan(&raw))
	if raw != "  Legacy@Example.invalid  " {
		t.Fatal("noop sửa email legacy")
	}
	var hash auth.SessionHash
	hash[0] = 1
	f.AddSession(t, a.UserID, hash, time.Now().Add(-time.Minute), time.Now().Add(time.Hour), nil)
	hash[0] = 2
	f.AddSession(t, a.UserID, hash, time.Now().Add(-time.Minute), time.Now().Add(time.Hour), nil)
	_, e = s.Edit(ctx, c)
	authdb.Require(t, "noop giữ phiên hiện hữu", e)
	if administrativeCount(t, f, `SELECT count(*) FROM auth_sessions WHERE revoked_at IS NOT NULL`) != 0 || administrativeCount(t, f, `SELECT count(*) FROM audit_events`) != 0 {
		t.Fatal("noop đã audit/revoke")
	}
	c.Changes.CompanyEmail = administrativeText(pcText("changed@example.invalid"))
	p, e = s.Edit(ctx, c)
	authdb.Require(t, "đổi email", e)
	if p.CompanyEmail != "changed@example.invalid" || p.EmploymentStatus != "INACTIVE" || administrativeCount(t, f, `SELECT count(*) FROM auth_sessions WHERE revoked_at IS NOT NULL`) != 2 || administrativeCount(t, f, `SELECT count(*) FROM audit_events WHERE operation='UPDATE_ADMINISTRATIVE_PROFILE'`) != 1 {
		t.Fatal("email/audit/revoke không nguyên tử")
	}
	var auditExact bool
	authdb.Require(t, "audit allowlist chỉ trường thực đổi", f.DB.Conn.QueryRow(ctx, `SELECT before_values='{"company_email":"legacy@example.invalid"}'::jsonb AND after_values='{"company_email":"changed@example.invalid"}'::jsonb AND actor_id=$1 AND entity_type='employee' AND entity_id=$2 FROM audit_events`, hr.UserID, a.EmployeeID).Scan(&auditExact))
	if !auditExact {
		t.Fatal("audit lộ field ngoài thay đổi hoặc sai actor")
	}
	_, e = f.DB.Conn.Exec(ctx, `UPDATE users SET status='LOCKED' WHERE user_id=$1`, a.UserID)
	authdb.Require(t, "target user locked riêng", e)
	_, e = s.Edit(ctx, c)
	requireCatalogError(t, e, org.ErrStale)
	_, e = s.Edit(ctx, administrativePhone(hr.UserID, a.EmployeeID, nil, pcText("123")))
	authdb.Require(t, "đổi phone nullable", e)
	_, e = s.Edit(ctx, administrativePhone(hr.UserID, a.EmployeeID, pcText("123"), nil))
	authdb.Require(t, "xóa phone", e)
	for _, roles := range [][]string{{"ADMIN"}, {"DIRECTOR"}, {}} {
		other := f.AddAccount(t, authdb.AccountOptions{Email: strings.ToLower(strings.Join(roles, "x")) + "other@example.invalid", Roles: roles})
		_, e = s.Read(ctx, org.AdministrativeProfileQuery{ActorID: other.UserID, EmployeeID: a.EmployeeID})
		requireCatalogError(t, e, org.ErrForbidden)
	}
	_, e = s.Read(ctx, org.AdministrativeProfileQuery{ActorID: hr.UserID, EmployeeID: 2147483647})
	requireCatalogError(t, e, org.ErrNotFound)
}

func TestAdministrativeProfileActorEqualsTarget(t *testing.T) {
	f := authdb.NewFixture(t)
	h := f.AddAccount(t, authdb.AccountOptions{Email: "hr@example.invalid", Roles: []string{"HR"}})
	var hash auth.SessionHash
	hash[0] = 3
	f.AddSession(t, h.UserID, hash, time.Now().Add(-time.Minute), time.Now().Add(time.Hour), nil)
	c := org.EditAdministrativeProfileCommand{ActorID: h.UserID, EmployeeID: h.EmployeeID, ExpectedBeforeValues: org.AdministrativeValues{CompanyEmail: administrativeText(pcText("hr@example.invalid"))}, Changes: org.AdministrativeValues{CompanyEmail: administrativeText(pcText("self-edit@example.invalid"))}}
	p, e := administrativeService(t, f.DB.Conn).Edit(t.Context(), c)
	authdb.Require(t, "HR tự sửa", e)
	if p.CompanyEmail != "self-edit@example.invalid" || administrativeCount(t, f, `SELECT count(*) FROM auth_sessions WHERE revoked_at IS NOT NULL`) != 1 {
		t.Fatal("actor=target khóa/revoke sai")
	}
}

func TestAdministrativeProfileAuditRollbackAndCommitUnknown(t *testing.T) {
	for _, mode := range []string{"audit", "revoke", "commit-lost", "commit-rollback"} {
		t.Run(mode, func(t *testing.T) {
			f := authdb.NewFixture(t)
			h := f.AddAccount(t, authdb.AccountOptions{Email: "hr@example.invalid", Roles: []string{"HR"}})
			a := f.AddAccount(t, authdb.AccountOptions{Email: "target@example.invalid"})
			var hash auth.SessionHash
			hash[0] = 1
			f.AddSession(t, a.UserID, hash, time.Now().Add(-time.Minute), time.Now().Add(time.Hour), nil)
			if mode == "audit" || mode == "revoke" {
				table := "audit_events"
				event := "INSERT"
				if mode == "revoke" {
					table = "auth_sessions"
					event = "UPDATE"
				}
				_, e := f.DB.Conn.Exec(t.Context(), `CREATE FUNCTION public.administrative_fault() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'synthetic failure' USING ERRCODE='23514'; END $$; CREATE TRIGGER administrative_fault BEFORE `+event+` ON `+table+` FOR EACH ROW EXECUTE FUNCTION public.administrative_fault()`)
				authdb.Require(t, "fault ghi thật", e)
			}
			var db accountBeginner = f.DB.Conn
			if mode == "commit-lost" {
				db = catalogHookDB{db: f.DB.Conn, commitLost: true}
			}
			if mode == "commit-rollback" {
				db = accountCommitRollbackDB{f.DB.Conn}
			}
			c := org.EditAdministrativeProfileCommand{ActorID: h.UserID, EmployeeID: a.EmployeeID, ExpectedBeforeValues: org.AdministrativeValues{CompanyEmail: administrativeText(pcText("target@example.invalid"))}, Changes: org.AdministrativeValues{CompanyEmail: administrativeText(pcText("new@example.invalid"))}}
			p, e := administrativeService(t, db).Edit(t.Context(), c)
			if p.EmployeeID != 0 {
				t.Fatal("lỗi trả success")
			}
			want := org.ErrPersistence
			if mode == "commit-lost" {
				want = org.ErrCommitUnknown
			}
			if !errors.Is(e, want) {
				t.Fatalf("phân loại lỗi %v", e)
			}
			committed := mode == "commit-lost"
			var email string
			authdb.Require(t, "đối soát", f.DB.Conn.QueryRow(t.Context(), `SELECT email FROM employees WHERE employee_id=$1`, a.EmployeeID).Scan(&email))
			if (email == "new@example.invalid") != committed || (administrativeCount(t, f, `SELECT count(*) FROM auth_sessions WHERE revoked_at IS NOT NULL`) == 1) != committed || (administrativeCount(t, f, `SELECT count(*) FROM audit_events`) == 1) != committed {
				t.Fatal("half-write hoặc retry")
			}
		})
	}
}

func TestAdministrativeProfileCatalogAndListSnapshot(t *testing.T) {
	f := authdb.NewFixture(t)
	h := f.AddAccount(t, authdb.AccountOptions{Email: "hr@example.invalid", Roles: []string{"HR", "ADMIN"}})
	a := f.AddAccount(t, authdb.AccountOptions{Email: "target@example.invalid"})
	s := administrativeService(t, f.DB.Conn)
	var d, j, p int
	authdb.Require(t, "department", f.DB.Conn.QueryRow(t.Context(), `INSERT INTO departments(department_name) VALUES('Phòng') RETURNING department_id`).Scan(&d))
	authdb.Require(t, "level", f.DB.Conn.QueryRow(t.Context(), `INSERT INTO job_levels(job_level_name) VALUES('Bậc') RETURNING job_level_id`).Scan(&j))
	authdb.Require(t, "position", f.DB.Conn.QueryRow(t.Context(), `SELECT position_id FROM employees WHERE employee_id=$1`, a.EmployeeID).Scan(&p))
	c := org.EditAdministrativeProfileCommand{ActorID: h.UserID, EmployeeID: a.EmployeeID, ExpectedBeforeValues: org.AdministrativeValues{DepartmentID: administrativeInt(nil), JobLevelID: administrativeInt(nil), PositionID: administrativeInt(&p)}, Changes: org.AdministrativeValues{DepartmentID: administrativeInt(&d), JobLevelID: administrativeInt(&j), PositionID: administrativeInt(&p)}}
	v, e := s.Edit(t.Context(), c)
	authdb.Require(t, "sửa danh mục", e)
	if v.Department == nil || v.JobLevel == nil {
		t.Fatal("thiếu danh mục")
	}
	missing := 2147483647
	c.ExpectedBeforeValues = c.Changes
	c.Changes.PositionID = administrativeInt(&missing)
	_, e = s.Edit(t.Context(), c)
	requireCatalogError(t, e, org.ErrCatalogNotFound)
	page, e := s.List(t.Context(), org.ListAdministrativeEmployeesQuery{ActorID: h.UserID, Page: 9, PageSize: 20})
	authdb.Require(t, "page rỗng", e)
	if page.Total != 2 || len(page.Items) != 0 {
		t.Fatal("page rỗng mất total")
	}
	for _, q := range []string{"%", "_"} {
		page, e = s.List(t.Context(), org.ListAdministrativeEmployeesQuery{ActorID: h.UserID, Page: 1, PageSize: 20, Query: q})
		authdb.Require(t, "search literal", e)
		if page.Total != 0 {
			t.Fatal("wildcard chưa escape")
		}
	}
	pool := f.DB.Pool(t)
	defer pool.Close()
	snapshot := administrativeService(t, profileSnapshotDB{db: f.DB.Conn, afterActor: func() {
		_, e := pool.Exec(t.Context(), `UPDATE employees SET fullname='Tên mới' WHERE employee_id=$1`, a.EmployeeID)
		authdb.Require(t, "mutation giữa snapshot", e)
	}})
	page, e = snapshot.List(t.Context(), org.ListAdministrativeEmployeesQuery{ActorID: h.UserID, Page: 1, PageSize: 20, Query: a.Fullname})
	authdb.Require(t, "snapshot list", e)
	if page.Total != 1 || len(page.Items) != 1 {
		t.Fatal("count/items trộn snapshot")
	}
}
