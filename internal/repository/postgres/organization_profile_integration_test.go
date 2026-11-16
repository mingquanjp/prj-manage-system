//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"ai-staffing/internal/repository/postgres"
	org "ai-staffing/internal/service/organization"
	"ai-staffing/internal/testutil/authdb"
	"github.com/jackc/pgx/v5"
)

func profileService(t *testing.T, db interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}) *org.ProfileService {
	t.Helper()
	r, err := postgres.NewProfileRepository(db, 5*time.Second)
	authdb.Require(t, "khởi tạo profile", err)
	return org.NewProfileService(r)
}
func TestSelfProfileProjectionAndCurrentState(t *testing.T) {
	f := authdb.NewFixture(t)
	a := f.AddAccount(t, authdb.AccountOptions{Email: "  Synthetic@Example.invalid  ", Roles: []string{"ADMIN", "HR", "DIRECTOR"}})
	other := f.AddAccount(t, authdb.AccountOptions{Email: "other@example.invalid"})
	s := profileService(t, f.DB.Conn)
	q := org.SelfProfileQuery{ActorID: a.UserID, EmployeeID: a.EmployeeID}
	ctx := t.Context()
	if a.UserID == a.EmployeeID {
		t.Fatal("fixture cần hai ID khác nhau")
	}
	p, err := s.ReadSelf(ctx, q)
	authdb.Require(t, "đọc self", err)
	if p.UserID != a.UserID || p.EmployeeID != a.EmployeeID || p.CompanyEmail != "synthetic@example.invalid" || p.Position.Name != "Fixture position" || p.Address != nil || p.Phone != nil || p.Department != nil || p.JobLevel != nil {
		t.Fatal("projection/link/null sai")
	}
	var raw string
	authdb.Require(t, "email legacy giữ nguyên", f.DB.Conn.QueryRow(ctx, `SELECT email FROM employees WHERE employee_id=$1`, a.EmployeeID).Scan(&raw))
	if raw != "  Synthetic@Example.invalid  " {
		t.Fatal("đọc đã sửa dữ liệu")
	}
	for _, id := range []int{other.EmployeeID, a.UserID, 2147483647} {
		_, err = s.ReadSelf(ctx, org.SelfProfileQuery{ActorID: a.UserID, EmployeeID: id})
		requireCatalogError(t, err, org.ErrNotFound)
	}
	_, err = f.DB.Conn.Exec(ctx, `INSERT INTO departments(department_name) VALUES ('Phòng tổng hợp'); INSERT INTO job_levels(job_level_name) VALUES ('Bậc tổng hợp')`)
	authdb.Require(t, "danh mục fixture", err)
	_, err = f.DB.Conn.Exec(ctx, `UPDATE employees SET department_id=(SELECT department_id FROM departments),job_level_id=(SELECT job_level_id FROM job_levels),phone='000000',address='Địa chỉ tổng hợp' WHERE user_id=$1`, a.UserID)
	authdb.Require(t, "liên kết danh mục", err)
	p, err = s.ReadSelf(ctx, q)
	authdb.Require(t, "đọc liên kết", err)
	if p.Department == nil || p.JobLevel == nil || p.Phone == nil || p.Address == nil {
		t.Fatal("mất trường nullable đã có dữ liệu")
	}
	position := p.Position.ID
	dept := p.Department.ID
	level := p.JobLevel.ID
	_, err = f.DB.Conn.Exec(ctx, `UPDATE positions SET position_name='Vị trí mới'; UPDATE departments SET department_name='Phòng mới'; UPDATE job_levels SET job_level_name='Bậc mới'`)
	authdb.Require(t, "đổi tên fixture", err)
	p, err = s.ReadSelf(ctx, q)
	authdb.Require(t, "đọc tên mới", err)
	if p.Position.ID != position || p.Department.ID != dept || p.JobLevel.ID != level || p.Position.Name != "Vị trí mới" || p.Department.Name != "Phòng mới" || p.JobLevel.Name != "Bậc mới" {
		t.Fatal("không đọc danh mục hiện hành")
	}
	for _, tc := range []struct {
		sql  string
		want error
	}{
		{`UPDATE users SET status='LOCKED' WHERE user_id=$1`, org.ErrUnauthenticated},
		{`UPDATE users SET status='ACTIVE',must_change_password=true WHERE user_id=$1`, org.ErrPasswordChangeRequired},
		{`UPDATE employees SET employment_status='INACTIVE' WHERE user_id=$1`, org.ErrUnauthenticated},
	} {
		_, err = f.DB.Conn.Exec(ctx, tc.sql, a.UserID)
		authdb.Require(t, "đổi nền fixture", err)
		for _, id := range []int{a.EmployeeID, other.EmployeeID} {
			_, err = s.ReadSelf(ctx, org.SelfProfileQuery{ActorID: a.UserID, EmployeeID: id})
			requireCatalogError(t, err, tc.want)
		}
	}
}

// Barrier sau Scan actor: connection khác commit trước câu đọc projection.
type profileSnapshotDB struct {
	db         *pgx.Conn
	afterActor func()
}

func (d profileSnapshotDB) BeginTx(ctx context.Context, o pgx.TxOptions) (pgx.Tx, error) {
	tx, e := d.db.BeginTx(ctx, o)
	return &profileSnapshotTx{Tx: tx, afterActor: d.afterActor}, e
}

type profileSnapshotTx struct {
	pgx.Tx
	afterActor func()
	reads      int
}

func (tx *profileSnapshotTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	tx.reads++
	row := tx.Tx.QueryRow(ctx, sql, args...)
	if tx.reads == 1 {
		return profileActorRow{row, tx.afterActor}
	}
	return row
}

type profileActorRow struct {
	pgx.Row
	after func()
}

func (r profileActorRow) Scan(dest ...any) error {
	err := r.Row.Scan(dest...)
	if err == nil {
		r.after()
	}
	return err
}
func TestSelfProfileConsistentSnapshot(t *testing.T) {
	f := authdb.NewFixture(t)
	a := f.AddAccount(t, authdb.AccountOptions{Email: "snapshot@example.invalid"})
	pool := f.DB.Pool(t)
	defer pool.Close()
	s := profileService(t, profileSnapshotDB{db: f.DB.Conn, afterActor: func() {
		_, err := pool.Exec(t.Context(), `UPDATE employees SET employment_status='INACTIVE',fullname='Nền mới' WHERE user_id=$1`, a.UserID)
		authdb.Require(t, "connection hai commit tại barrier", err)
	}})
	q := org.SelfProfileQuery{ActorID: a.UserID, EmployeeID: a.EmployeeID}
	p, err := s.ReadSelf(t.Context(), q)
	authdb.Require(t, "cùng snapshot", err)
	if p.Fullname != a.Fullname || p.EmploymentStatus != "ACTIVE" {
		t.Fatal("đọc trộn hai snapshot")
	}
	_, err = profileService(t, f.DB.Conn).ReadSelf(t.Context(), q)
	requireCatalogError(t, err, org.ErrUnauthenticated)
}
func TestSelfProfileDatabaseErrors(t *testing.T) {
	f := authdb.NewFixture(t)
	a := f.AddAccount(t, authdb.AccountOptions{Email: "error@example.invalid"})
	s := profileService(t, f.DB.Conn)
	q := org.SelfProfileQuery{ActorID: a.UserID, EmployeeID: a.EmployeeID}
	_, err := f.DB.Conn.Exec(t.Context(), `ALTER TABLE positions RENAME TO positions_unavailable`)
	authdb.Require(t, "fault schema trong case hy sinh", err)
	_, err = s.ReadSelf(t.Context(), q)
	requireCatalogError(t, err, org.ErrPersistence)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = s.ReadSelf(ctx, q)
	if !errors.Is(err, context.Canceled) {
		t.Fatal("mất lỗi hủy context")
	}
}

func TestSelfProfileMissingBasisAndMalformedData(t *testing.T) {
	f := authdb.NewFixture(t)
	a := f.AddAccount(t, authdb.AccountOptions{Email: "basis@example.invalid"})
	orphan := f.AddAccount(t, authdb.AccountOptions{WithoutEmployee: true})
	s := profileService(t, f.DB.Conn)
	q := org.SelfProfileQuery{ActorID: a.UserID, EmployeeID: a.EmployeeID}
	for _, user := range []int{orphan.UserID, 2147483647} {
		_, err := s.ReadSelf(t.Context(), org.SelfProfileQuery{ActorID: user, EmployeeID: a.EmployeeID})
		requireCatalogError(t, err, org.ErrUnauthenticated)
	}
	_, err := f.DB.Conn.Exec(t.Context(), `UPDATE employees SET email='malformed' WHERE user_id=$1`, a.UserID)
	authdb.Require(t, "email fixture hỏng", err)
	_, err = s.ReadSelf(t.Context(), q)
	requireCatalogError(t, err, org.ErrPersistence)
	_, err = f.DB.Conn.Exec(t.Context(), `UPDATE employees SET email='basis@example.invalid' WHERE user_id=$1`, a.UserID)
	authdb.Require(t, "khôi phục email", err)
	// Chỉ case hy sinh: chứng minh thiếu position là lỗi nền, không phải 404.
	_, err = f.DB.Conn.Exec(t.Context(), `ALTER TABLE employees DROP CONSTRAINT employees_position_id_fkey; DELETE FROM positions`)
	authdb.Require(t, "fault FK fixture", err)
	_, err = s.ReadSelf(t.Context(), q)
	requireCatalogError(t, err, org.ErrPersistence)
}
