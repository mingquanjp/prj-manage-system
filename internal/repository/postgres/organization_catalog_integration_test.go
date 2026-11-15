//go:build integration

package postgres_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	"ai-staffing/internal/repository/postgres"
	org "ai-staffing/internal/service/organization"
	"ai-staffing/internal/testutil/authdb"
	"github.com/jackc/pgx/v5"
)

type catalogTestSpec struct {
	kind                    org.CatalogKind
	table, id, name, entity string
}

var catalogTestSpecs = []catalogTestSpec{
	{org.Departments, "departments", "department_id", "department_name", "department"},
	{org.Positions, "positions", "position_id", "position_name", "position"},
	{org.JobLevels, "job_levels", "job_level_id", "job_level_name", "job_level"},
}

func catalogService(t *testing.T, db interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}) *org.CatalogService {
	t.Helper()
	r, err := postgres.NewCatalogRepository(db, 10*time.Second)
	authdb.Require(t, "khởi tạo catalog adapter", err)
	return org.NewCatalogService(r)
}
func catalogActor(t *testing.T, f *authdb.Fixture, roles ...string) authdb.Account {
	t.Helper()
	return f.AddAccount(t, authdb.AccountOptions{Email: "catalog@example.invalid", Roles: roles})
}
func catalogCount(t *testing.T, f *authdb.Fixture, sql string, args ...any) int {
	t.Helper()
	var n int
	authdb.Require(t, "đếm dữ liệu tổng hợp", f.DB.Conn.QueryRow(t.Context(), sql, args...).Scan(&n))
	return n
}
func requireCatalogError(t *testing.T, err, want error) {
	t.Helper()
	if !errors.Is(err, want) {
		t.Fatalf("sai phân loại lỗi: muốn %v; không xuất lỗi DB thô", want)
	}
}

func TestCatalogCRUDAndAudit(t *testing.T) {
	f := authdb.NewFixture(t)
	actor := catalogActor(t, f, "HR")
	s := catalogService(t, f.DB.Conn)
	ctx := t.Context()
	for _, spec := range catalogTestSpecs {
		t.Run(string(spec.kind), func(t *testing.T) {
			c := org.CatalogCommand{ActorID: actor.UserID, Kind: spec.kind, Name: "\u00a0" + strings.Repeat("ế", 100) + "\u3000"}
			item, err := s.Create(ctx, c)
			authdb.Require(t, "tạo tên 100 code point", err)
			if item.Name != strings.Repeat("ế", 100) {
				t.Fatal("không trim Unicode")
			}
			c.Name = strings.Repeat("ế", 101)
			_, err = s.Create(ctx, c)
			requireCatalogError(t, err, org.ErrValidation)
			c.Name = "CATALOG"
			other, err := s.Create(ctx, c)
			authdb.Require(t, "tạo mục khác", err)
			c.Name = "catalog"
			_, err = s.Create(ctx, c)
			requireCatalogError(t, err, org.ErrDuplicateName)
			c.ID = item.ID
			c.ExpectedBeforeName = "khác"
			c.Name = other.Name
			_, err = s.Rename(ctx, c)
			requireCatalogError(t, err, org.ErrStale)
			c.ExpectedBeforeName = item.Name
			_, err = s.Rename(ctx, c)
			requireCatalogError(t, err, org.ErrDuplicateName)
			var beforeTime, timeAfter time.Time
			authdb.Require(t, "timestamp trước no-op", f.DB.Conn.QueryRow(ctx, "SELECT updated_at FROM public."+spec.table+" WHERE "+spec.id+"=$1", item.ID).Scan(&beforeTime))
			beforeCount := catalogCount(t, f, "SELECT count(*) FROM public.audit_events")
			c.Name = " " + item.Name + " "
			_, err = s.Rename(ctx, c)
			authdb.Require(t, "rename no-op", err)
			authdb.Require(t, "timestamp sau no-op", f.DB.Conn.QueryRow(ctx, "SELECT updated_at FROM public."+spec.table+" WHERE "+spec.id+"=$1", item.ID).Scan(&timeAfter))
			if !timeAfter.Equal(beforeTime) || catalogCount(t, f, "SELECT count(*) FROM public.audit_events") != beforeCount {
				t.Fatal("no-op không được cập nhật timestamp/audit")
			}
			c.Name = "Tên mới"
			item, err = s.Rename(ctx, c)
			authdb.Require(t, "rename thật", err)
			requireCatalogError(t, s.Delete(ctx, c), org.ErrStale)
			c.ExpectedBeforeName = item.Name
			authdb.Require(t, "xóa mục", s.Delete(ctx, c))
			requireCatalogError(t, s.Delete(ctx, c), org.ErrNotFound)
			_, err = s.Rename(ctx, c)
			requireCatalogError(t, err, org.ErrNotFound)
			rows, err := f.DB.Conn.Query(ctx, `SELECT actor_id,operation,before_values,after_values FROM public.audit_events WHERE entity_type=$1 AND entity_id=$2 ORDER BY audit_id`, spec.entity, item.ID)
			authdb.Require(t, "đọc audit", err)
			ops := []string{}
			for rows.Next() {
				var who int
				var op string
				var b, a []byte
				authdb.Require(t, "scan audit", rows.Scan(&who, &op, &b, &a))
				if who != actor.UserID {
					t.Fatal("actor audit sai")
				}
				ops = append(ops, op)
				for _, raw := range [][]byte{b, a} {
					if raw == nil {
						continue
					}
					var payload map[string]any
					authdb.Require(t, "JSON audit", json.Unmarshal(raw, &payload))
					if len(payload) != 3 || payload["schema_version"] != float64(1) || payload["id"] != float64(item.ID) || payload["name"] == nil {
						t.Fatal("audit vượt allowlist")
					}
				}
				if op == "CREATE" && b != nil || op == "DELETE" && a != nil {
					t.Fatal("audit null trước/sau sai")
				}
			}
			rows.Close()
			authdb.Require(t, "kết thúc audit", rows.Err())
			if !reflect.DeepEqual(ops, []string{"CREATE", "RENAME", "DELETE"}) {
				t.Fatal("audit phải atomic, không ghi cho lỗi/no-op")
			}
			// Không NFC/bỏ dấu ngầm.
			c.Name = "é"
			_, err = s.Create(ctx, c)
			authdb.Require(t, "tên NFC", err)
			c.Name = "e\u0301"
			_, err = s.Create(ctx, c)
			authdb.Require(t, "tên NFD riêng", err)
		})
	}
}

func TestCatalogCurrentActorAndReferences(t *testing.T) {
	f := authdb.NewFixture(t)
	hr := catalogActor(t, f, "HR")
	admin := f.AddAccount(t, authdb.AccountOptions{Email: "admin@example.invalid", Roles: []string{"ADMIN"}})
	plain := f.AddAccount(t, authdb.AccountOptions{Email: "plain@example.invalid", LegacyRole: "ADMIN"})
	s := catalogService(t, f.DB.Conn)
	ctx := t.Context()
	for _, a := range []authdb.Account{hr, admin} {
		_, err := s.List(ctx, org.ListCatalogQuery{ActorID: a.UserID, Kind: org.Departments})
		authdb.Require(t, "HR/Admin đọc", err)
	}
	c := org.CatalogCommand{ActorID: admin.UserID, Kind: org.Departments, Name: "X"}
	_, err := s.Create(ctx, c)
	requireCatalogError(t, err, org.ErrForbidden)
	_, err = s.List(ctx, org.ListCatalogQuery{ActorID: plain.UserID, Kind: org.Departments})
	requireCatalogError(t, err, org.ErrForbidden)
	for _, spec := range catalogTestSpecs {
		c = org.CatalogCommand{ActorID: hr.UserID, Kind: spec.kind, Name: "Được tham chiếu"}
		item, err := s.Create(ctx, c)
		authdb.Require(t, "catalog reference", err)
		_, err = f.DB.Conn.Exec(ctx, "UPDATE public.employees SET "+spec.id+"=$1,employment_status='INACTIVE' WHERE employee_id=$2", item.ID, admin.EmployeeID)
		authdb.Require(t, "employee inactive tham chiếu", err)
		c.ID = item.ID
		c.ExpectedBeforeName = item.Name
		requireCatalogError(t, s.Delete(ctx, c), org.ErrCatalogInUse)
	}
	for index, tc := range []struct {
		sql  string
		want error
	}{
		{`UPDATE public.users SET must_change_password=true WHERE user_id=$1`, org.ErrPasswordChangeRequired},
		{`UPDATE public.users SET must_change_password=false,status='LOCKED' WHERE user_id=$1`, org.ErrUnauthenticated},
		{`UPDATE public.users SET status='ACTIVE' WHERE user_id=$1`, nil},
		{`UPDATE public.employees SET employment_status='INACTIVE' WHERE user_id=$1`, org.ErrUnauthenticated},
		{`UPDATE public.employees SET employment_status='ACTIVE' WHERE user_id=$1`, nil},
		{`DELETE FROM public.user_roles WHERE user_id=$1`, org.ErrForbidden},
	} {
		_, err = f.DB.Conn.Exec(ctx, tc.sql, hr.UserID)
		authdb.Require(t, "đổi nguồn hiện hành", err)
		_, err = s.List(ctx, org.ListCatalogQuery{ActorID: hr.UserID, Kind: org.Departments})
		requireCatalogError(t, err, tc.want)
		_, err = s.Create(ctx, org.CatalogCommand{ActorID: hr.UserID, Kind: org.Departments, Name: "Sau thay đổi " + strings.Repeat("x", index)})
		requireCatalogError(t, err, tc.want)
	}
}

// Hook chỉ nằm trong test, giữ transaction/SQL thật để ép mutation giữa count/items.
type catalogHookDB struct {
	db interface {
		BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
	}
	hook         func()
	commitLost   bool
	beforeCommit func()
	afterRow     func(string)
}

func (d catalogHookDB) BeginTx(ctx context.Context, o pgx.TxOptions) (pgx.Tx, error) {
	tx, err := d.db.BeginTx(ctx, o)
	if err != nil {
		return nil, err
	}
	return &catalogHookTx{Tx: tx, hook: d.hook, commitLost: d.commitLost, beforeCommit: d.beforeCommit, afterRow: d.afterRow}, nil
}

type catalogHookTx struct {
	pgx.Tx
	hook         func()
	commitLost   bool
	beforeCommit func()
	afterRow     func(string)
}

func (tx *catalogHookTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if tx.hook != nil {
		h := tx.hook
		tx.hook = nil
		h()
	}
	return tx.Tx.Query(ctx, sql, args...)
}
func (tx *catalogHookTx) Commit(ctx context.Context) error {
	if tx.beforeCommit != nil {
		tx.beforeCommit()
	}
	err := tx.Tx.Commit(ctx)
	if err == nil && tx.commitLost {
		return io.ErrUnexpectedEOF
	}
	return err
}

func TestCatalogPaginationSnapshotAndLiteral(t *testing.T) {
	f := authdb.NewFixture(t)
	a := catalogActor(t, f, "HR")
	s := catalogService(t, f.DB.Conn)
	ctx := t.Context()
	for _, spec := range catalogTestSpecs {
		for _, name := range []string{"a%_\\b", "plain", "Last"} {
			_, err := s.Create(ctx, org.CatalogCommand{ActorID: a.UserID, Kind: spec.kind, Name: name})
			authdb.Require(t, "tạo danh mục tìm kiếm", err)
		}
		for _, q := range []string{"%", "_", `\`, "A%_\\B"} {
			p, err := s.List(ctx, org.ListCatalogQuery{ActorID: a.UserID, Kind: spec.kind, Query: q})
			authdb.Require(t, "literal search", err)
			if p.Total != 1 || len(p.Items) != 1 || p.Items[0].Name != "a%_\\b" {
				t.Fatal("query phải literal và không phân biệt hoa thường")
			}
		}
		for _, page := range []int64{1, 999} {
			worker := f.DB.Connect(t)
			hs := catalogService(t, catalogHookDB{db: f.DB.Conn, hook: func() {
				_, err := worker.Exec(ctx, "DELETE FROM public."+spec.table+" WHERE "+spec.name+"='Last'")
				authdb.Require(t, "delete giữa count/items", err)
				_, err = worker.Exec(ctx, "INSERT INTO public."+spec.table+" ("+spec.name+") VALUES ('After snapshot')")
				authdb.Require(t, "insert giữa count/items", err)
			}})
			p, err := hs.List(ctx, org.ListCatalogQuery{ActorID: a.UserID, Kind: spec.kind, Page: page})
			authdb.Require(t, "snapshot phân trang", err)
			want := int64(3)
			if spec.kind == org.Positions {
				want++
			}
			if p.Total != want {
				t.Fatal("total không cùng snapshot")
			}
			if page == 1 {
				found := false
				for _, i := range p.Items {
					if i.Name == "Last" {
						found = true
					}
					if i.Name == "After snapshot" {
						t.Fatal("items bị đổi snapshot")
					}
				}
				if !found {
					t.Fatal("mất item trong snapshot")
				}
			} else if len(p.Items) != 0 {
				t.Fatal("page vượt tổng phải items rỗng")
			}
			_, err = worker.Exec(ctx, "DELETE FROM public."+spec.table+" WHERE "+spec.name+"='After snapshot'")
			authdb.Require(t, "dọn mutation snapshot", err)
			_, err = worker.Exec(ctx, "INSERT INTO public."+spec.table+" ("+spec.name+") VALUES ('Last')")
			authdb.Require(t, "khôi phục fixture snapshot", err)
		}
	}
}

func TestCatalogAtomicAuditAndCommit(t *testing.T) {
	f := authdb.NewFixture(t)
	a := catalogActor(t, f, "HR")
	s := catalogService(t, f.DB.Conn)
	ctx := t.Context()
	items := map[org.CatalogKind]org.CatalogItem{}
	for _, spec := range catalogTestSpecs {
		item, err := s.Create(ctx, org.CatalogCommand{ActorID: a.UserID, Kind: spec.kind, Name: "Keep on rollback"})
		authdb.Require(t, "mục trước lỗi audit", err)
		items[spec.kind] = item
	}
	baseline := catalogCount(t, f, `SELECT count(*) FROM public.audit_events`)
	_, err := f.DB.Conn.Exec(ctx, `CREATE FUNCTION public.fail_catalog_audit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'synthetic failure'; END $$; CREATE TRIGGER fail_catalog_audit BEFORE INSERT ON public.audit_events FOR EACH ROW EXECUTE FUNCTION public.fail_catalog_audit()`)
	authdb.Require(t, "gắn lỗi audit tổng hợp", err)
	for _, spec := range catalogTestSpecs {
		item := items[spec.kind]
		cmd := org.CatalogCommand{ActorID: a.UserID, Kind: spec.kind, ID: item.ID, Name: "Rollback", ExpectedBeforeName: item.Name}
		_, err = s.Create(ctx, cmd)
		requireCatalogError(t, err, org.ErrPersistence)
		_, err = s.Rename(ctx, cmd)
		requireCatalogError(t, err, org.ErrPersistence)
		requireCatalogError(t, s.Delete(ctx, cmd), org.ErrPersistence)
		if catalogCount(t, f, "SELECT count(*) FROM public."+spec.table+" WHERE "+spec.name+"='Rollback'") != 0 || catalogCount(t, f, "SELECT count(*) FROM public."+spec.table+" WHERE "+spec.id+"=$1 AND "+spec.name+"=$2", item.ID, item.Name) != 1 || catalogCount(t, f, `SELECT count(*) FROM public.audit_events`) != baseline {
			t.Fatal("audit lỗi phải rollback cả create/rename/delete")
		}
	}

	_, err = f.DB.Conn.Exec(ctx, `DROP TRIGGER fail_catalog_audit ON public.audit_events; CREATE CONSTRAINT TRIGGER fail_catalog_commit AFTER INSERT ON public.audit_events DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION public.fail_catalog_audit()`)
	authdb.Require(t, "lỗi commit deferred thật", err)
	_, err = s.Create(ctx, org.CatalogCommand{ActorID: a.UserID, Kind: org.Departments, Name: "Commit failure"})
	requireCatalogError(t, err, org.ErrCommitUnknown)
	if catalogCount(t, f, `SELECT count(*) FROM public.departments WHERE department_name='Commit failure'`) != 0 || catalogCount(t, f, `SELECT count(*) FROM public.audit_events`) != baseline {
		t.Fatal("commit lỗi không được giữ write nửa chừng")
	}
	_, err = f.DB.Conn.Exec(ctx, `DROP TRIGGER fail_catalog_commit ON public.audit_events`)
	authdb.Require(t, "gỡ lỗi deferred", err)
	lost := catalogService(t, catalogHookDB{db: f.DB.Conn, commitLost: true})
	_, err = lost.Create(ctx, org.CatalogCommand{ActorID: a.UserID, Kind: org.Departments, Name: "Committed but response lost"})
	requireCatalogError(t, err, org.ErrCommitUnknown)
	requireCatalogError(t, err, org.ErrUnavailable)
	if catalogCount(t, f, `SELECT count(*) FROM public.departments WHERE department_name='Committed but response lost'`) != 1 || catalogCount(t, f, `SELECT count(*) FROM public.audit_events`) != baseline+1 {
		t.Fatal("mất phản hồi giả lập sau commit phải đối soát, không retry")
	}
}

type catalogHookRow struct {
	pgx.Row
	after func()
}

func (r catalogHookRow) Scan(dest ...any) error {
	err := r.Row.Scan(dest...)
	if err == nil && r.after != nil {
		r.after()
	}
	return err
}
func (tx *catalogHookTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return catalogHookRow{Row: tx.Tx.QueryRow(ctx, sql, args...), after: func() {
		if tx.afterRow != nil {
			tx.afterRow(sql)
		}
	}}
}
