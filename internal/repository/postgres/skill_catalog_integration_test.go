//go:build integration

package postgres_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"ai-staffing/internal/repository/postgres"
	sc "ai-staffing/internal/service/skillcatalog"
	"ai-staffing/internal/testutil/authdb"
	"github.com/jackc/pgx/v5"
)

func skillCatalogService(t *testing.T, db interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}) *sc.Service {
	t.Helper()
	r, e := postgres.NewSkillCatalogRepository(db, 10*time.Second)
	authdb.Require(t, "khởi tạo adapter kỹ năng", e)
	return sc.NewService(r)
}
func skillCatalogRequireError(t *testing.T, err, want error) {
	t.Helper()
	if !errors.Is(err, want) {
		t.Fatalf("phân loại lỗi sai, muốn %v; không in lỗi DB thô", want)
	}
}
func skillCatalogCreate(t *testing.T, s *sc.Service, actor int, kind sc.Kind, name string, category int) sc.Item {
	t.Helper()
	v, e := s.Create(t.Context(), sc.Command{ActorID: actor, Kind: kind, Name: name, CategoryID: category})
	authdb.Require(t, "tạo danh mục kỹ năng", e)
	return v
}
func TestSkillCatalogCRUDPermissionsAndReferences(t *testing.T) {
	f := authdb.NewFixture(t)
	ctx := t.Context()
	director := f.AddAccount(t, authdb.AccountOptions{Email: "skill-director@example.invalid", Roles: []string{"DIRECTOR"}})
	employee := f.AddAccount(t, authdb.AccountOptions{Email: "skill-employee@example.invalid"})
	hr := f.AddAccount(t, authdb.AccountOptions{Email: "skill-hr@example.invalid", Roles: []string{"HR"}})
	admin := f.AddAccount(t, authdb.AccountOptions{Email: "skill-admin@example.invalid", Roles: []string{"ADMIN"}})
	s := skillCatalogService(t, f.DB.Conn)
	for _, a := range []authdb.Account{employee, hr, admin, director} {
		for _, kind := range []sc.Kind{sc.Categories, sc.Skills} {
			_, e := s.List(ctx, sc.ListQuery{ActorID: a.UserID, Kind: kind})
			authdb.Require(t, "phiên thường đọc", e)
		}
		if a != director {
			_, e := s.Create(ctx, sc.Command{ActorID: a.UserID, Kind: sc.Categories, Name: "Denied"})
			skillCatalogRequireError(t, e, sc.ErrForbidden)
		}
	}
	c1 := skillCatalogCreate(t, s, director.UserID, sc.Categories, "Group one", 0)
	c2 := skillCatalogCreate(t, s, director.UserID, sc.Categories, "Group two", 0)
	for _, kind := range []sc.Kind{sc.Categories, sc.Skills} {
		cat := 0
		if kind == sc.Skills {
			cat = c1.ID
		}
		cmd := sc.Command{ActorID: director.UserID, Kind: kind, Name: "\u00a0" + strings.Repeat("ế", 100) + "\u3000", CategoryID: cat}
		v, e := s.Create(ctx, cmd)
		authdb.Require(t, "tên 100 code point Unicode", e)
		if v.Name != strings.Repeat("ế", 100) {
			t.Fatal("normalize sai")
		}
		cmd.Name = strings.Repeat("ế", 101)
		_, e = s.Create(ctx, cmd)
		skillCatalogRequireError(t, e, sc.ErrValidation)
		cmd.Name = "Mixed Case"
		other, e := s.Create(ctx, cmd)
		authdb.Require(t, "tạo tên unique", e)
		cmd.Name = "mixed CASE"
		_, e = s.Create(ctx, cmd)
		skillCatalogRequireError(t, e, sc.ErrDuplicateName)
		cmd.ID = v.ID
		cmd.Name = "Updated"
		cmd.ExpectedBefore = sc.Before{Name: "stale", CategoryID: cat}
		_, e = s.Update(ctx, cmd)
		skillCatalogRequireError(t, e, sc.ErrStale)
		cmd.ExpectedBefore.Name = v.Name
		cmd.Name = "Mixed Case"
		_, e = s.Update(ctx, cmd)
		skillCatalogRequireError(t, e, sc.ErrDuplicateName)
		cmd.Name = " " + v.Name + " "
		before := catalogCount(t, f, `SELECT count(*) FROM audit_events`)
		_, e = s.Update(ctx, cmd)
		authdb.Require(t, "no-op tên", e)
		if catalogCount(t, f, `SELECT count(*) FROM audit_events`) != before {
			t.Fatal("no-op ghi audit")
		}
		cmd.Name = "Updated"
		if kind == sc.Skills {
			cmd.CategoryID = c2.ID
		}
		v, e = s.Update(ctx, cmd)
		authdb.Require(t, "rename và chuyển nhóm", e)
		if v.ID != cmd.ID || v.Name != "Updated" || (kind == sc.Skills && v.Category.ID != c2.ID) {
			t.Fatal("mutation đổi ID/nhóm sai")
		}
		skillCatalogRequireError(t, s.Delete(ctx, cmd), sc.ErrStale)
		cmd.ExpectedBefore = sc.Before{Name: v.Name, CategoryID: v.Category.ID}
		authdb.Require(t, "delete", s.Delete(ctx, cmd))
		skillCatalogRequireError(t, s.Delete(ctx, cmd), sc.ErrNotFound)
		_, e = s.Update(ctx, cmd)
		skillCatalogRequireError(t, e, sc.ErrNotFound)
		entity := "SKILL_CATEGORY"
		if kind == sc.Skills {
			entity = "SKILL"
		}
		rows, err := f.DB.Conn.Query(ctx, `SELECT actor_id,operation,before_values,after_values FROM audit_events WHERE entity_type=$1 AND entity_id=$2 ORDER BY audit_id`, entity, v.ID)
		authdb.Require(t, "audit mục đã thay đổi", err)
		operations := []string{}
		for rows.Next() {
			var who int
			var operation string
			var beforeRaw, afterRaw []byte
			authdb.Require(t, "đọc actor và payload audit", rows.Scan(&who, &operation, &beforeRaw, &afterRaw))
			if who != director.UserID {
				t.Fatal("audit sai actor")
			}
			operations = append(operations, operation)
			if operation == "CREATE" && beforeRaw != nil || operation == "DELETE" && afterRaw != nil {
				t.Fatal("audit sai NULL trước/sau")
			}
			for _, raw := range [][]byte{beforeRaw, afterRaw} {
				if raw == nil {
					continue
				}
				var payload map[string]any
				authdb.Require(t, "decode audit projection", json.Unmarshal(raw, &payload))
				count := 3
				if kind == sc.Skills {
					count = 4
				}
				if len(payload) != count || payload["schema_version"] != float64(1) || payload["id"] != float64(v.ID) {
					t.Fatal("audit vượt projection")
				}
			}
			if operation == "UPDATE" {
				var old, newValue map[string]any
				authdb.Require(t, "audit before rename", json.Unmarshal(beforeRaw, &old))
				authdb.Require(t, "audit after rename", json.Unmarshal(afterRaw, &newValue))
				if old["name"] != strings.Repeat("ế", 100) || newValue["name"] != "Updated" {
					t.Fatal("audit không giữ tên trước/sau")
				}
				if kind == sc.Skills && (old["category_id"] != float64(c1.ID) || newValue["category_id"] != float64(c2.ID)) {
					t.Fatal("audit không giữ nhóm trước/sau")
				}
			}
		}
		rows.Close()
		authdb.Require(t, "kết thúc audit", rows.Err())
		if strings.Join(operations, ",") != "CREATE,UPDATE,DELETE" {
			t.Fatal("audit phải đúng một sự kiện mỗi mutation")
		}
		cmd.ID = other.ID
		cmd.ExpectedBefore = sc.Before{Name: other.Name, CategoryID: other.Category.ID}
		authdb.Require(t, "xóa mục còn lại", s.Delete(ctx, cmd))
		// Hai biểu diễn dấu khác nhau vẫn là hai tên, không tự NFC hoặc bỏ dấu.
		skillCatalogCreate(t, s, director.UserID, kind, "é", cat)
		skillCatalogCreate(t, s, director.UserID, kind, "e\u0301", cat)
	}
	skill := skillCatalogCreate(t, s, director.UserID, sc.Skills, "Referenced", c1.ID)
	skillCatalogRequireError(t, s.Delete(ctx, sc.Command{ActorID: director.UserID, Kind: sc.Categories, ID: c1.ID, ExpectedBefore: sc.Before{Name: c1.Name}}), sc.ErrCatalogInUse)
	// Fixture trực tiếp kiểm FK xóa; không thay bằng chứng writer submit ở MSK-T06.
	_, e := f.DB.Conn.Exec(ctx, `INSERT INTO employee_skill_change_requests(employee_id,skill_id,operation,before_values,proposed_values,explanation,requested_by) VALUES ($1,$2,'ADD',NULL,'{"schema_version":1,"proficiency_level":"BASIC","years_experience":null,"note":null}','Fixture',$3)`, employee.EmployeeID, skill.ID, employee.UserID)
	authdb.Require(t, "request tham chiếu skill", e)
	skillCatalogRequireError(t, s.Delete(ctx, sc.Command{ActorID: director.UserID, Kind: sc.Skills, ID: skill.ID, ExpectedBefore: sc.Before{Name: skill.Name, CategoryID: c1.ID}}), sc.ErrCatalogInUse)
	_, e = f.DB.Conn.Exec(ctx, `UPDATE employee_skill_change_requests SET status='REJECTED',resolved_by=$1,resolved_at=clock_timestamp(),decision_reason='Fixture'`, director.UserID)
	authdb.Require(t, "terminal tham chiếu", e)
	skillCatalogRequireError(t, s.Delete(ctx, sc.Command{ActorID: director.UserID, Kind: sc.Skills, ID: skill.ID, ExpectedBefore: sc.Before{Name: skill.Name, CategoryID: c1.ID}}), sc.ErrCatalogInUse)
	officialOnly := skillCatalogCreate(t, s, director.UserID, sc.Skills, "Official only", c1.ID)
	_, e = f.DB.Conn.Exec(ctx, `INSERT INTO employee_skills(employee_id,skill_id,proficiency_level,approved_by,approved_at) VALUES ($1,$2,'BASIC',$3,now())`, employee.EmployeeID, officialOnly.ID, director.UserID)
	authdb.Require(t, "official tham chiếu", e)
	_, e = f.DB.Conn.Exec(ctx, `UPDATE employees SET employment_status='INACTIVE' WHERE employee_id=$1`, employee.EmployeeID)
	authdb.Require(t, "target inactive", e)
	skillCatalogRequireError(t, s.Delete(ctx, sc.Command{ActorID: director.UserID, Kind: sc.Skills, ID: officialOnly.ID, ExpectedBefore: sc.Before{Name: officialOnly.Name, CategoryID: c1.ID}}), sc.ErrCatalogInUse)
	for _, tc := range []struct {
		sql  string
		want error
	}{{`DELETE FROM user_roles WHERE user_id=$1`, sc.ErrForbidden}, {`UPDATE users SET must_change_password=true WHERE user_id=$1`, sc.ErrPasswordChangeRequired}, {`UPDATE users SET must_change_password=false,status='LOCKED' WHERE user_id=$1`, sc.ErrUnauthenticated}, {`UPDATE users SET status='ACTIVE' WHERE user_id=$1;`, nil}} {
		_, e = f.DB.Conn.Exec(ctx, tc.sql, director.UserID)
		authdb.Require(t, "thay trạng thái actor", e)
		_, e = s.Create(ctx, sc.Command{ActorID: director.UserID, Kind: sc.Categories, Name: "Actor changed"})
		if tc.want != nil {
			skillCatalogRequireError(t, e, tc.want)
		} else {
			skillCatalogRequireError(t, e, sc.ErrForbidden)
		}
	}
}
func TestSkillCatalogSnapshotLiteralAndAudit(t *testing.T) {
	f := authdb.NewFixture(t)
	ctx := t.Context()
	actor := f.AddAccount(t, authdb.AccountOptions{Email: "skill-snapshot@example.invalid", Roles: []string{"DIRECTOR"}})
	s := skillCatalogService(t, f.DB.Conn)
	category := skillCatalogCreate(t, s, actor.UserID, sc.Categories, "Group", 0)
	for _, kind := range []sc.Kind{sc.Categories, sc.Skills} {
		cat := 0
		if kind == sc.Skills {
			cat = category.ID
		}
		for _, name := range []string{"Zulu", "Alpha %_", "Alpha any"} {
			skillCatalogCreate(t, s, actor.UserID, kind, name, cat)
		}
		page, e := s.List(ctx, sc.ListQuery{ActorID: actor.UserID, Kind: kind, Query: "%_", CategoryID: cat})
		authdb.Require(t, "literal %_", e)
		if page.Total != 1 || len(page.Items) != 1 || page.Items[0].Name != "Alpha %_" {
			t.Fatal("q bị coi wildcard")
		}
		worker := f.DB.Connect(t)
		table, column := "skill_categories", "category_name"
		if kind == sc.Skills {
			table, column = "skills", "skill_name"
		}
		for _, number := range []int{1, 99} {
			seen := false
			hs := skillCatalogService(t, catalogHookDB{db: f.DB.Conn, hook: func() {
				seen = true
				_, err := worker.Exec(ctx, "UPDATE "+table+" SET "+column+"='After snapshot' WHERE "+column+"='Zulu'")
				authdb.Require(t, "mutation giữa count/items", err)
			}})
			p, err := hs.List(ctx, sc.ListQuery{ActorID: actor.UserID, Kind: kind, CategoryID: cat, Page: number})
			authdb.Require(t, "RR snapshot", err)
			wantTotal := int64(3)
			if kind == sc.Categories {
				wantTotal++
			}
			if !seen || p.Total != wantTotal {
				t.Fatal("total snapshot sai")
			}
			if number == 1 {
				names := []string{}
				for _, v := range p.Items {
					if v.Name == "After snapshot" {
						t.Fatal("items lệch snapshot")
					}
					names = append(names, v.Name)
				}
				if names[0] != "Alpha %_" {
					t.Fatal("sort phải name/id")
				}
			} else if len(p.Items) != 0 {
				t.Fatal("page rỗng sai")
			}
			_, err = worker.Exec(ctx, "UPDATE "+table+" SET "+column+"='Zulu' WHERE "+column+"='After snapshot'")
			authdb.Require(t, "khôi phục snapshot fixture", err)
		}
	}
	baseline := catalogCount(t, f, `SELECT count(*) FROM audit_events`)
	_, e := f.DB.Conn.Exec(ctx, `CREATE FUNCTION fail_skill_audit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'synthetic' USING ERRCODE='23514'; END $$; CREATE TRIGGER fail_skill_audit BEFORE INSERT ON audit_events FOR EACH ROW EXECUTE FUNCTION fail_skill_audit()`)
	authdb.Require(t, "audit fault", e)
	for _, kind := range []sc.Kind{sc.Categories, sc.Skills} {
		cat := 0
		if kind == sc.Skills {
			cat = category.ID
		}
		_, e = s.Create(ctx, sc.Command{ActorID: actor.UserID, Kind: kind, Name: "Must rollback", CategoryID: cat})
		skillCatalogRequireError(t, e, sc.ErrPersistence)
	}
	for _, kind := range []sc.Kind{sc.Categories, sc.Skills} {
		table, idColumn, nameColumn := "skill_categories", "category_id", "category_name"
		cat := 0
		if kind == sc.Skills {
			table, idColumn, nameColumn = "skills", "skill_id", "skill_name"
			cat = category.ID
		}
		var id int
		authdb.Require(t, "mục trước lỗi audit", f.DB.Conn.QueryRow(ctx, "SELECT "+idColumn+" FROM "+table+" WHERE "+nameColumn+"='Alpha %_'").Scan(&id))
		c := sc.Command{ActorID: actor.UserID, Kind: kind, ID: id, Name: "Must rollback", CategoryID: cat, ExpectedBefore: sc.Before{Name: "Alpha %_", CategoryID: cat}}
		_, err := s.Update(ctx, c)
		skillCatalogRequireError(t, err, sc.ErrPersistence)
		skillCatalogRequireError(t, s.Delete(ctx, c), sc.ErrPersistence)
		if catalogCount(t, f, "SELECT count(*) FROM "+table+" WHERE "+idColumn+"=$1 AND "+nameColumn+"='Alpha %_'", id) != 1 {
			t.Fatal("update/delete không rollback khi audit lỗi")
		}
	}
	if catalogCount(t, f, `SELECT count(*) FROM skills WHERE skill_name='Must rollback'`)+catalogCount(t, f, `SELECT count(*) FROM skill_categories WHERE category_name='Must rollback'`) != 0 || catalogCount(t, f, `SELECT count(*) FROM audit_events`) != baseline {
		t.Fatal("audit failure half-write")
	}
	_, e = f.DB.Conn.Exec(ctx, `DROP TRIGGER fail_skill_audit ON audit_events; CREATE FUNCTION skip_skill_audit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RETURN NULL; END $$; CREATE TRIGGER skip_skill_audit BEFORE INSERT ON audit_events FOR EACH ROW EXECUTE FUNCTION skip_skill_audit()`)
	authdb.Require(t, "audit skip fixture", e)
	_, e = s.Create(ctx, sc.Command{ActorID: actor.UserID, Kind: sc.Categories, Name: "Skipped audit"})
	skillCatalogRequireError(t, e, sc.ErrPersistence)
	if catalogCount(t, f, `SELECT count(*) FROM skill_categories WHERE category_name='Skipped audit'`) != 0 {
		t.Fatal("audit không ghi vẫn commit")
	}
	_, e = f.DB.Conn.Exec(ctx, `DROP TRIGGER skip_skill_audit ON audit_events; CREATE CONSTRAINT TRIGGER fail_skill_commit AFTER INSERT ON audit_events DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION fail_skill_audit()`)

	authdb.Require(t, "fault commit trì hoãn", e)
	_, e = s.Create(ctx, sc.Command{ActorID: actor.UserID, Kind: sc.Categories, Name: "Deferred failure"})
	skillCatalogRequireError(t, e, sc.ErrPersistence)
	if errors.Is(e, sc.ErrCommitUnknown) {
		t.Fatal("constraint rejection xác nhận rollback")
	}
	if catalogCount(t, f, `SELECT count(*) FROM skill_categories WHERE category_name='Deferred failure'`) != 0 {
		t.Fatal("commit failure vẫn ghi dữ liệu")
	}
	_, e = f.DB.Conn.Exec(ctx, `DROP TRIGGER fail_skill_commit ON audit_events`)
	authdb.Require(t, "gỡ fault commit", e)
	rollback := skillCatalogService(t, accountCommitRollbackDB{f.DB.Conn})
	_, e = rollback.Create(ctx, sc.Command{ActorID: actor.UserID, Kind: sc.Categories, Name: "Known rollback"})
	skillCatalogRequireError(t, e, sc.ErrPersistence)
	if errors.Is(e, sc.ErrCommitUnknown) {
		t.Fatal("rollback đã xác nhận bị báo unknown")
	}
	_, e = rollback.Update(ctx, sc.Command{ActorID: actor.UserID, Kind: sc.Categories, ID: category.ID, Name: category.Name, ExpectedBefore: sc.Before{Name: category.Name}})
	skillCatalogRequireError(t, e, sc.ErrPersistence)
	if errors.Is(e, sc.ErrCommitUnknown) {
		t.Fatal("no-op rollback bị báo unknown")
	}
	lost := skillCatalogService(t, catalogHookDB{db: f.DB.Conn, commitLost: true})
	_, e = lost.Create(ctx, sc.Command{ActorID: actor.UserID, Kind: sc.Categories, Name: "Committed response lost"})
	skillCatalogRequireError(t, e, sc.ErrCommitUnknown)
	if catalogCount(t, f, `SELECT count(*) FROM skill_categories WHERE category_name='Committed response lost'`) != 1 || catalogCount(t, f, `SELECT count(*) FROM audit_events`) != baseline+1 {
		t.Fatal("lost acknowledgment không đối soát được")
	}
}
func TestSkillCatalogConcurrentUniqueAndCategoryReference(t *testing.T) {
	f := authdb.NewFixture(t)
	firstActor := f.AddAccount(t, authdb.AccountOptions{Email: "skill-race1@example.invalid", Roles: []string{"DIRECTOR"}})
	secondActor := f.AddAccount(t, authdb.AccountOptions{Email: "skill-race2@example.invalid", Roles: []string{"DIRECTOR"}})
	s := skillCatalogService(t, f.DB.Conn)
	category := skillCatalogCreate(t, s, firstActor.UserID, sc.Categories, "Race group", 0)
	for _, kind := range []sc.Kind{sc.Categories, sc.Skills} {
		t.Run(string(kind), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			c1, c2 := f.DB.Connect(t), f.DB.Connect(t)
			observer, e := f.DB.Conn.Begin(ctx)
			authdb.Require(t, "observer", e)
			defer observer.Rollback(context.Background())
			block, release, entered := catalogBarrier(t, ctx)
			s1 := skillCatalogService(t, catalogHookDB{db: c1, beforeCommit: block})
			s2 := skillCatalogService(t, c2)
			cat := 0
			if kind == sc.Skills {
				cat = category.ID
			}
			one, two := make(chan error, 1), make(chan error, 1)
			go func() {
				_, err := s1.Create(ctx, sc.Command{ActorID: firstActor.UserID, Kind: kind, Name: "Concurrent", CategoryID: cat})
				one <- err
			}()
			waitCatalogSignal(t, ctx, entered)
			go func() {
				_, err := s2.Create(ctx, sc.Command{ActorID: secondActor.UserID, Kind: kind, Name: "CONCURRENT", CategoryID: cat})
				two <- err
			}()
			waitForAuthLock(t, ctx, observer, c1.PgConn().PID(), c2.PgConn().PID())
			release()
			authdb.Require(t, "first create", catalogResult(t, ctx, one))
			skillCatalogRequireError(t, catalogResult(t, ctx, two), sc.ErrDuplicateName)
		})
	}
	for _, createFirst := range []bool{true, false} {
		scenario, suffix := "delete holds group", "two"
		if createFirst {
			scenario, suffix = "create holds group", "one"
		}
		t.Run(scenario, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			c1, c2 := f.DB.Connect(t), f.DB.Connect(t)
			observer, e := f.DB.Conn.Begin(ctx)
			authdb.Require(t, "observer reference", e)
			defer observer.Rollback(context.Background())
			cat := skillCatalogCreate(t, skillCatalogService(t, c1), firstActor.UserID, sc.Categories, "Reference race "+suffix, 0)
			block, release, entered := catalogBarrier(t, ctx)
			s1 := skillCatalogService(t, catalogHookDB{db: c1, beforeCommit: block})
			s2 := skillCatalogService(t, c2)
			one, two := make(chan error, 1), make(chan error, 1)
			deleteCmd := sc.Command{ActorID: firstActor.UserID, Kind: sc.Categories, ID: cat.ID, ExpectedBefore: sc.Before{Name: cat.Name}}
			createCmd := sc.Command{ActorID: secondActor.UserID, Kind: sc.Skills, Name: "Reference raced skill " + suffix, CategoryID: cat.ID}
			if createFirst {
				createCmd.ActorID = firstActor.UserID
				deleteCmd.ActorID = secondActor.UserID
				go func() { _, err := s1.Create(ctx, createCmd); one <- err }()
			} else {
				go func() { one <- s1.Delete(ctx, deleteCmd) }()
			}
			waitCatalogSignal(t, ctx, entered)
			if createFirst {
				go func() { two <- s2.Delete(ctx, deleteCmd) }()
			} else {
				go func() { _, err := s2.Create(ctx, createCmd); two <- err }()
			}
			waitForAuthLock(t, ctx, observer, c1.PgConn().PID(), c2.PgConn().PID())
			release()
			authdb.Require(t, "reference first commit", catalogResult(t, ctx, one))
			want := sc.ErrNotFound
			if createFirst {
				want = sc.ErrCatalogInUse
			}
			skillCatalogRequireError(t, catalogResult(t, ctx, two), want)
			if catalogCount(t, f, `SELECT count(*) FROM skills s LEFT JOIN skill_categories c ON c.category_id=s.category_id WHERE c.category_id IS NULL`) != 0 {
				t.Fatal("race tạo orphan")
			}
		})
	}
}
func TestSkillCatalogActorChangeAndCategoryMoveStale(t *testing.T) {
	f := authdb.NewFixture(t)
	a := f.AddAccount(t, authdb.AccountOptions{Email: "skill-recheck@example.invalid", Roles: []string{"DIRECTOR"}})
	other := f.AddAccount(t, authdb.AccountOptions{Email: "skill-other@example.invalid", Roles: []string{"DIRECTOR"}})
	s := skillCatalogService(t, f.DB.Conn)
	c1 := skillCatalogCreate(t, s, a.UserID, sc.Categories, "Old group", 0)
	c2 := skillCatalogCreate(t, s, a.UserID, sc.Categories, "New group", 0)
	item := skillCatalogCreate(t, s, a.UserID, sc.Skills, "Move", c1.ID)
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	worker := f.DB.Connect(t)
	entered := false
	hook := catalogHookDB{db: f.DB.Conn, afterRow: func(sql string) {
		if !entered && strings.HasPrefix(sql, "SELECT category_id FROM public.skills") {
			entered = true
			_, err := skillCatalogService(t, worker).Update(ctx, sc.Command{ActorID: other.UserID, Kind: sc.Skills, ID: item.ID, Name: item.Name, CategoryID: c2.ID, ExpectedBefore: sc.Before{Name: item.Name, CategoryID: c1.ID}})
			authdb.Require(t, "move trong snapshot lookup", err)
		}
	}}
	_, e := skillCatalogService(t, hook).Update(ctx, sc.Command{ActorID: a.UserID, Kind: sc.Skills, ID: item.ID, Name: "Must not overwrite", CategoryID: c1.ID, ExpectedBefore: sc.Before{Name: item.Name, CategoryID: c1.ID}})
	skillCatalogRequireError(t, e, sc.ErrStale)
	for _, tc := range []struct {
		name, sql string
		want      error
	}{{"role", `DELETE FROM user_roles WHERE user_id=$1`, sc.ErrForbidden}, {"limited", `UPDATE users SET must_change_password=true WHERE user_id=$1`, sc.ErrPasswordChangeRequired}, {"inactive", `UPDATE employees SET employment_status='INACTIVE' WHERE user_id=$1`, sc.ErrUnauthenticated}} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := f.DB.Conn.Exec(ctx, `UPDATE users SET must_change_password=false WHERE user_id=$1;`, a.UserID)
			authdb.Require(t, "khôi phục cờ actor", err)
			_, err = f.DB.Conn.Exec(ctx, `UPDATE employees SET employment_status='ACTIVE' WHERE user_id=$1`, a.UserID)
			authdb.Require(t, "khôi phục trạng thái actor", err)
			_, err = f.DB.Conn.Exec(ctx, `INSERT INTO user_roles(user_id,role_code) VALUES ($1,'DIRECTOR') ON CONFLICT DO NOTHING`, a.UserID)
			authdb.Require(t, "khôi phục quyền actor", err)
			blocker, err := f.DB.Conn.Begin(ctx)
			authdb.Require(t, "transaction thay actor", err)
			defer blocker.Rollback(context.Background())
			_, err = blocker.Exec(ctx, `SELECT user_id FROM users WHERE user_id=$1 FOR UPDATE`, a.UserID)
			authdb.Require(t, "khóa user của actor", err)
			result := make(chan error, 1)
			go func() {
				_, e := skillCatalogService(t, worker).Create(ctx, sc.Command{ActorID: a.UserID, Kind: sc.Categories, Name: "Denied " + tc.name})
				result <- e
			}()
			waitForAuthLock(t, ctx, blocker, f.DB.Conn.PgConn().PID(), worker.PgConn().PID())
			_, err = blocker.Exec(ctx, tc.sql, a.UserID)
			authdb.Require(t, "thay trạng thái actor", err)
			authdb.Require(t, "commit thay đổi actor", blocker.Commit(ctx))
			skillCatalogRequireError(t, catalogResult(t, ctx, result), tc.want)
		})
	}
	if catalogCount(t, f, `SELECT count(*) FROM skill_categories WHERE category_name LIKE 'Denied%'`) != 0 {
		t.Fatal("stale actor ghi dữ liệu")
	}
}
