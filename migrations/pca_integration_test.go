//go:build integration

package migrations_test

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"runtime"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"ai-staffing/internal/testutil/authdb"
	"ai-staffing/migrations"
	"github.com/golang-migrate/migrate/v4"
)

func TestPCAEmptyDownUp(t *testing.T) {
	db := authdb.NewCase(t)
	m := db.OpenMigrator(t, migrations.FS)
	authdb.Require(t, "nâng empty tới PCA", m.Up())
	head(t, m, migrations.PCAVersion)
	assertPCAPresent(t, db, true)
	if !errors.Is(m.Up(), migrate.ErrNoChange) {
		t.Fatal("up head phải không thay đổi")
	}
	var n int
	authdb.Require(t, "cấp notification ID trước down", db.Conn.QueryRow(context.Background(), `SELECT nextval(pg_get_serial_sequence('public.notifications','notification_id'))`).Scan(&n))
	authdb.Require(t, "down PCA empty", m.Steps(-1))
	head(t, m, migrations.M2Version)
	assertPCAPresent(t, db, false)
	authdb.Require(t, "up PCA lại", m.Up())
	head(t, m, migrations.PCAVersion)
	var next int
	authdb.Require(t, "sequence giữ high water", db.Conn.QueryRow(context.Background(), `SELECT nextval(pg_get_serial_sequence('public.notifications','notification_id'))`).Scan(&next))
	if next != n+1 {
		t.Fatal("up/down không được reset sequence notification")
	}
}

func pcaLegacy(t *testing.T) (*authdb.Case, *authdb.Migrator) {
	t.Helper()
	db := authdb.NewCase(t)
	m := db.OpenMigrator(t, migrations.FS)
	authdb.Require(t, "nền lịch sử v3", m.Migrate(3))
	fixture(t, db)
	exec(t, db.Conn, `INSERT INTO user_roles VALUES(101,'HR'),(101,'DIRECTOR');
 INSERT INTO auth_sessions(user_id,token_hash,expires_at) VALUES(101,decode(repeat('ab',32),'hex'),now()+interval '8h');
 UPDATE employees SET email='  FIXTURE-ADMIN@EXAMPLE.INVALID  ' WHERE user_id=101`)
	authdb.Require(t, "nền M2", m.Migrate(migrations.M2Version))
	exec(t, db.Conn, `INSERT INTO audit_events(actor_id,operation,entity_type,entity_id,before_values,after_values)
 VALUES(101,'CATALOG_CREATED','DEPARTMENT',1,NULL,'{"name":"Fixture"}')`)
	return db, m
}

func TestPCALegacyPreservesAuthOAP(t *testing.T) {
	db, m := pcaLegacy(t)
	before := m2Snapshot(t, db)
	assertM2Sequence(t, db, 1, false)
	authdb.Require(t, "nâng legacy tới PCA", m.Up())
	head(t, m, migrations.PCAVersion)
	sameJSON(t, before, m2Snapshot(t, db))
	assertM2Sequence(t, db, 1, false)
	// user_id 101 không phải employee_id 201; recipient là user.
	expectState(t, db.Conn, "23503", `INSERT INTO notifications(recipient_id,notification_type,resource_type,resource_id,message) VALUES(201,'PROFILE_CHANGE_SUBMITTED','PROFILE_CHANGE_REQUEST',1,'Có đề nghị hồ sơ')`)
	exec(t, db.Conn, `INSERT INTO notifications(recipient_id,notification_type,resource_type,resource_id,message) VALUES(101,'PROFILE_CHANGE_SUBMITTED','PROFILE_CHANGE_REQUEST',1,'Có đề nghị hồ sơ')`)
	expectState(t, db.Conn, "23503", `DELETE FROM users WHERE user_id=101`)
	sameJSON(t, before, m2Snapshot(t, db))
}

func TestPCAPreflightRefusesBeforeDirty(t *testing.T) {
	for _, tc := range []struct{ name, sql string }{
		{"notification data", `INSERT INTO notifications(employee_id,type,title,content) VALUES(201,'TASK','Fixture','Nội dung tổng hợp')`},
		{"view", `CREATE VIEW pca_dependency AS SELECT title FROM notifications`},
		{"FK", `CREATE TABLE pca_dependency(id int REFERENCES notifications(notification_id))`},
		{"trigger", `CREATE FUNCTION pca_trigger() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RETURN NEW; END $$; CREATE TRIGGER pca_dependency BEFORE INSERT ON notifications FOR EACH ROW EXECUTE FUNCTION pca_trigger()`},
		{"function body", `CREATE FUNCTION pca_dependency() RETURNS int LANGUAGE plpgsql AS $$ BEGIN RETURN (SELECT count(*)::int FROM public.notifications); END $$`},
		{"column drift", `ALTER TABLE notifications ALTER COLUMN title TYPE text`},
		{"column default", `ALTER TABLE notifications ALTER COLUMN title SET DEFAULT 'unknown'`},
		{"FK action drift", `ALTER TABLE notifications DROP CONSTRAINT notifications_employee_id_fkey; ALTER TABLE notifications ADD FOREIGN KEY(employee_id) REFERENCES employees(employee_id)`},
		{"extra column", `ALTER TABLE notifications ADD COLUMN unknown text`},
		{"check", `ALTER TABLE notifications ADD CHECK(length(title)>0)`},
		{"extra index", `CREATE INDEX pca_dependency ON notifications(title)`},
		{"wrong index", `DROP INDEX idx_notifications_employee; CREATE INDEX idx_notifications_employee ON notifications(employee_id)`},
		{"RLS", `ALTER TABLE notifications ENABLE ROW LEVEL SECURITY`},
		{"index name collision", `CREATE INDEX idx_pca_owner ON employees(fullname)`},
		{"recipient index name collision", `CREATE INDEX idx_notifications_recipient ON employees(fullname)`},
		{"sequence name collision", `CREATE SEQUENCE employee_profile_change_requests_request_id_seq`},
		{"row type collision", `CREATE TYPE idempotency_records AS (unknown int)`},
		{"preexisting table", `CREATE TABLE idempotency_records(id int)`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, m := pcaLegacy(t)
			exec(t, db.Conn, tc.sql)
			before := m2Snapshot(t, db)
			var notifications string
			authdb.Require(t, "snapshot notifications chỉ trong bộ nhớ", db.Conn.QueryRow(context.Background(), `SELECT coalesce(jsonb_agg(to_jsonb(n)),'[]')::text FROM notifications n`).Scan(&notifications))
			if err := m.Up(); err == nil || !strings.HasPrefix(err.Error(), "PCA:") {
				t.Fatal("preflight phải từ chối trước dirty")
			}
			head(t, m, migrations.M2Version)
			sameJSON(t, before, m2Snapshot(t, db))
			var after string
			authdb.Require(t, "notification không đổi", db.Conn.QueryRow(context.Background(), `SELECT coalesce(jsonb_agg(to_jsonb(n)),'[]')::text FROM notifications n`).Scan(&after))
			sameJSON(t, notifications, after)
		})
	}
}

const pcaInsert = `INSERT INTO employee_profile_change_requests(employee_id,requested_by,before_values,proposed_values) VALUES(201,101,$1::jsonb,$2::jsonb)`

func TestPCAJSONStateFKAndIndexes(t *testing.T) {
	db, m := pcaLegacy(t)
	authdb.Require(t, "nâng PCA", m.Up())
	assertPCAPresent(t, db, true)
	goodBefore := `{"schema_version":1,"address":null}`
	goodAfter := `{"schema_version":1,"address":"Địa chỉ tổng hợp"}`
	for _, tc := range []struct{ before, after string }{
		{`null`, goodAfter}, {`[]`, goodAfter}, {`{"schema_version":2,"address":null}`, goodAfter},
		{`{"schema_version":1}`, `{"schema_version":1}`},
		{`{"schema_version":1,"phone":null}`, goodAfter},
		{goodBefore, `{"schema_version":1,"address":null}`},
		{goodBefore, `{"schema_version":1,"address":""}`},
		{goodBefore, `{"schema_version":1,"address":" x"}`},
		{goodBefore, `{"schema_version":1,"address":5}`},
		{goodBefore, `{"schema_version":1,"address":"x","role":"HR"}`},
		{`{"schema_version":1,"fullname":null}`, `{"schema_version":1,"fullname":"x"}`},
		{`{"schema_version":1,"phone":null}`, `{"schema_version":1,"phone":""}`},
	} {
		expectState(t, db.Conn, "23514", pcaInsert, tc.before, tc.after)
	}
	for _, field := range []struct {
		name  string
		limit int
	}{{"fullname", 100}, {"phone", 20}, {"address", 2000}} {
		before := fmt.Sprintf(`{"schema_version":1,%q:"legacy"}`, field.name)
		after := fmt.Sprintf(`{"schema_version":1,%q:%q}`, field.name, strings.Repeat("ế", field.limit))
		exec(t, db.Conn, pcaInsert, before, after)
		after = fmt.Sprintf(`{"schema_version":1,%q:%q}`, field.name, strings.Repeat("ế", field.limit+1))
		expectState(t, db.Conn, "23514", pcaInsert, before, after)
	}
	// Before legacy giữ padding/độ dài; proposed không chấp nhận Unicode whitespace ngoài biên.
	exec(t, db.Conn, pcaInsert, fmt.Sprintf(`{"schema_version":1,"fullname":%q}`, strings.Repeat("a", 101)), `{"schema_version":1,"fullname":"Canonical"}`)
	for _, space := range []string{"\t", "\n", "\u0085", "\u00a0", "\u1680", "\u2000", "\u2028", "\u202f", "\u205f", "\u3000"} {
		after := fmt.Sprintf(`{"schema_version":1,"address":%q}`, space+"x")
		expectState(t, db.Conn, "23514", pcaInsert, goodBefore, after)
	}
	exec(t, db.Conn, pcaInsert, goodBefore, goodAfter)
	for _, change := range []string{
		`status='UNKNOWN'`, `status='APPROVED'`, `status='REJECTED',resolved_by=101,resolved_at=now()`,
		`status='REJECTED',resolved_by=101,resolved_at=now(),decision_reason=U&'\00A0'`,
		`status='APPROVED',resolved_by=101,resolved_at=now(),decision_reason='x'`,
		`resolved_by=101`, `explanation=repeat('x',2001)`,
		`status='APPROVED',resolved_by=101,resolved_at=requested_at-interval '1s'`,
	} {
		expectState(t, db.Conn, "23514", `UPDATE employee_profile_change_requests SET `+change+` WHERE request_id=(SELECT max(request_id) FROM employee_profile_change_requests)`)
	}
	expectState(t, db.Conn, "23503", `UPDATE employee_profile_change_requests SET requested_by=99999`)
	expectState(t, db.Conn, "23503", `UPDATE employee_profile_change_requests SET employee_id=101`)
	exec(t, db.Conn, `UPDATE employee_profile_change_requests SET status='APPROVED',resolved_by=101,resolved_at=clock_timestamp() WHERE request_id=(SELECT max(request_id) FROM employee_profile_change_requests)`)
	exec(t, db.Conn, pcaInsert, goodBefore, goodAfter)
	exec(t, db.Conn, `UPDATE employee_profile_change_requests SET status='REJECTED',resolved_by=101,resolved_at=clock_timestamp(),decision_reason='Lý do tổng hợp' WHERE request_id=(SELECT max(request_id) FROM employee_profile_change_requests)`)
	keySQL := `INSERT INTO idempotency_records(actor_id,operation,idempotency_key,request_hash,result_reference,expires_at) VALUES($1,'PROFILE_CHANGE_SUBMIT',$2,$3,$4::jsonb,$5::timestamptz)`
	hash := make([]byte, 32)
	exec(t, db.Conn, keySQL, 101, "synthetic-key", hash, `{"schema_version":1,"request_id":1,"result_code":"CREATED"}`, nil)
	expectState(t, db.Conn, "23505", keySQL, 101, "synthetic-key", hash, `{}`, nil)
	expectState(t, db.Conn, "23503", keySQL, 99999, "key", hash, `{}`, nil)
	expectState(t, db.Conn, "23514", keySQL, 101, "key", make([]byte, 31), `{}`, nil)
	expectState(t, db.Conn, "23514", keySQL, 101, "key", hash, `[]`, nil)
	expectState(t, db.Conn, "23514", keySQL, 101, "key", hash, `{}`, "2000-01-01Z")
	expectState(t, db.Conn, "23514", `INSERT INTO notifications(recipient_id,notification_type,resource_type,resource_id,message) VALUES(101,'TYPE','RESOURCE',0,'x')`)
	expectState(t, db.Conn, "23514", `INSERT INTO notifications(recipient_id,notification_type,resource_type,resource_id,message,read_at) VALUES(101,'TYPE','RESOURCE',1,'x','2000-01-01Z')`)
	expectState(t, db.Conn, "23514", `INSERT INTO notifications(recipient_id,notification_type,resource_type,resource_id,message) VALUES(101,'TYPE','RESOURCE',1,repeat('x',201))`)
}

func TestPCADownRefusesEachHistoryTable(t *testing.T) {
	for _, table := range []string{"request", "key", "notification"} {
		t.Run(table, func(t *testing.T) {
			db, m := pcaLegacy(t)
			authdb.Require(t, "nâng PCA", m.Up())
			switch table {
			case "request":
				exec(t, db.Conn, pcaInsert, `{"schema_version":1,"phone":null}`, `{"schema_version":1,"phone":"123"}`)
			case "key":
				exec(t, db.Conn, `INSERT INTO idempotency_records(actor_id,operation,idempotency_key,request_hash,result_reference) VALUES(101,'PROFILE_CHANGE_SUBMIT','key',decode(repeat('ab',32),'hex'),'{}')`)
			case "notification":
				exec(t, db.Conn, `INSERT INTO notifications(recipient_id,notification_type,resource_type,resource_id,message) VALUES(101,'TYPE','RESOURCE',1,'Tổng hợp')`)
			}
			before := pcaSnapshot(t, db)
			if m.Steps(-1) == nil {
				t.Fatal("down phải từ chối lịch sử")
			}
			m.Close(t)
			assertPCAPresent(t, db, true)
			sameJSON(t, before, pcaSnapshot(t, db))
			assertPCADirty(t, db, migrations.M2Version)
		})
	}
}

func TestPCADDLFailureAndTOCTOUGuard(t *testing.T) {
	for _, stage := range []string{"after DDL", "after preflight"} {
		t.Run(stage, func(t *testing.T) {
			db, m := pcaLegacy(t)
			before := m2Snapshot(t, db)
			m.Close(t)
			source := fstest.MapFS{}
			entries, err := fs.ReadDir(migrations.FS, ".")
			authdb.Require(t, "đọc nguồn", err)
			for _, entry := range entries {
				content, err := fs.ReadFile(migrations.FS, entry.Name())
				authdb.Require(t, "đọc migration", err)
				if entry.Name() == "000006_pca_profile_change.up.sql" {
					if stage == "after DDL" {
						content = []byte(strings.Replace(string(content), "COMMIT;", "SELECT 1/0;\nCOMMIT;", 1))
					} else {
						// Barrier ở đầu DDL, sau khi preflight và dirty đã hoàn tất.
						content = []byte("SELECT pg_advisory_lock(638020);\n" + string(content))
					}
				}
				source[entry.Name()] = &fstest.MapFile{Data: content}
			}
			broken := db.OpenMigrator(t, source)
			var upErr error
			if stage == "after preflight" {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				exec(t, db.Conn, `SELECT pg_advisory_lock(638020)`)
				defer db.Conn.Exec(context.Background(), `SELECT pg_advisory_unlock(638020)`)
				result := make(chan error, 1)
				go func() { result <- broken.Up() }()
				// Chờ server xác nhận connection migration đã tới barrier, không sleep.
				for {
					var waiting bool
					authdb.Require(t, "barrier sau preflight", db.Conn.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pg_locks WHERE locktype='advisory' AND objid=638020 AND NOT granted)`).Scan(&waiting))
					if waiting {
						break
					}
					select {
					case <-ctx.Done():
						t.Fatal("migration chưa tới barrier")
					default:
						runtime.Gosched()
					}
				}
				exec(t, db.Conn, `INSERT INTO notifications(employee_id,type,title) VALUES(201,'TASK','Fixture')`)
				exec(t, db.Conn, `SELECT pg_advisory_unlock(638020)`)
				select {
				case upErr = <-result:
				case <-ctx.Done():
					t.Fatal("DDL guard không kết thúc")
				}
			} else {
				upErr = broken.Up()
			}
			if upErr == nil {
				t.Fatal("guard/fault phải thất bại")
			}
			broken.Close(t)
			assertPCAPresent(t, db, false)
			sameJSON(t, before, m2Snapshot(t, db))
			assertPCADirty(t, db, migrations.PCAVersion)
			var count int
			authdb.Require(t, "dòng xen giữa được giữ", db.Conn.QueryRow(context.Background(), `SELECT count(*) FROM notifications`).Scan(&count))
			want := 0
			if stage == "after preflight" {
				want = 1
			}
			if count != want {
				t.Fatal("guard không giữ đúng dữ liệu")
			}
			retry := db.OpenMigrator(t, migrations.FS)
			var dirty migrate.ErrDirty
			if !errors.As(retry.Up(), &dirty) || dirty.Version != migrations.PCAVersion {
				t.Fatal("không được tự force/retry DB dirty")
			}
		})
	}
}

func assertPCADirty(t *testing.T, db *authdb.Case, want int) {
	t.Helper()
	var v int
	var dirty bool
	authdb.Require(t, "metadata dirty driver", db.Conn.QueryRow(context.Background(), `SELECT version,dirty FROM schema_migrations`).Scan(&v, &dirty))
	if v != want || !dirty {
		t.Fatal("metadata dirty không khớp driver")
	}
}
func pcaSnapshot(t *testing.T, db *authdb.Case) string {
	t.Helper()
	var s string
	authdb.Require(t, "snapshot PCA trong bộ nhớ", db.Conn.QueryRow(context.Background(), `SELECT jsonb_build_array((SELECT jsonb_agg(to_jsonb(r)) FROM employee_profile_change_requests r),(SELECT jsonb_agg(to_jsonb(k)) FROM idempotency_records k),(SELECT jsonb_agg(to_jsonb(n)) FROM notifications n))::text`).Scan(&s))
	return s
}
func assertPCAPresent(t *testing.T, db *authdb.Case, present bool) {
	t.Helper()
	var got bool
	authdb.Require(t, "schema PCA", db.Conn.QueryRow(context.Background(), `SELECT to_regclass('public.employee_profile_change_requests') IS NOT NULL AND to_regclass('public.idempotency_records') IS NOT NULL AND to_regprocedure('public.pca_profile_values_valid(jsonb,boolean)') IS NOT NULL AND to_regclass('public.idx_pca_owner') IS NOT NULL AND to_regclass('public.idx_pca_queue') IS NOT NULL AND to_regclass('public.idx_notifications_recipient') IS NOT NULL`).Scan(&got))
	if got != present {
		t.Fatal("schema PCA không khớp kỳ vọng")
	}
	if !present {
		var absent bool
		authdb.Require(t, "rollback toàn bộ PCA và trả schema notification legacy", db.Conn.QueryRow(context.Background(), `SELECT
        to_regclass('public.employee_profile_change_requests') IS NULL
        AND to_regclass('public.idempotency_records') IS NULL
        AND to_regprocedure('public.pca_profile_values_valid(jsonb,boolean)') IS NULL
        AND to_regprocedure('public.pca_same_profile_keys(jsonb,jsonb)') IS NULL
        AND to_regclass('public.idx_notifications_recipient') IS NULL
        AND to_regclass('public.idx_notifications_employee') IS NOT NULL
        AND (SELECT count(*)=8 FROM information_schema.columns WHERE table_schema='public' AND table_name='notifications')
        AND (SELECT count(*)=6 FROM information_schema.columns WHERE table_schema='public' AND table_name='notifications' AND column_name IN ('employee_id','type','title','content','ref_task_id','is_read'))`).Scan(&absent))
		if !absent {
			t.Fatal("PCA để lại cấu trúc nửa chừng")
		}
		return
	}
	var metadata bool
	authdb.Require(t, "định nghĩa index và FK PCA", db.Conn.QueryRow(context.Background(), `SELECT
 (SELECT count(*)=3 FROM pg_indexes WHERE schemaname='public' AND indexdef IN (
 'CREATE INDEX idx_pca_owner ON public.employee_profile_change_requests USING btree (employee_id, requested_at DESC, request_id DESC)',
 'CREATE INDEX idx_pca_queue ON public.employee_profile_change_requests USING btree (status, requested_at DESC, request_id DESC)',
 'CREATE INDEX idx_notifications_recipient ON public.notifications USING btree (recipient_id, created_at DESC, notification_id DESC)'))
 AND (SELECT count(*)=5 FROM pg_constraint WHERE contype='f'
 AND conrelid IN ('public.employee_profile_change_requests'::regclass,'public.idempotency_records'::regclass,'public.notifications'::regclass)
 AND confdeltype='a' AND confupdtype='a' AND convalidated AND NOT condeferrable)`).Scan(&metadata))
	if !metadata {
		t.Fatal("index hoặc FK PCA sai định nghĩa")
	}
}

func TestPCADownDDLFailureRollsBack(t *testing.T) {
	db, m := pcaLegacy(t)
	authdb.Require(t, "nâng PCA", m.Up())
	before := m2Snapshot(t, db)
	m.Close(t)
	source := fstest.MapFS{}
	entries, err := fs.ReadDir(migrations.FS, ".")
	authdb.Require(t, "đọc nguồn migration", err)
	for _, entry := range entries {
		content, err := fs.ReadFile(migrations.FS, entry.Name())
		authdb.Require(t, "đọc SQL", err)
		if entry.Name() == "000006_pca_profile_change.down.sql" {
			content = []byte(strings.Replace(string(content), "COMMIT;", "SELECT 1/0;\nCOMMIT;", 1))
		}
		source[entry.Name()] = &fstest.MapFile{Data: content}
	}
	broken := db.OpenMigrator(t, source)
	if broken.Steps(-1) == nil {
		t.Fatal("down DDL fault phải thất bại")
	}
	broken.Close(t)
	assertPCAPresent(t, db, true)
	sameJSON(t, before, m2Snapshot(t, db))
	assertPCADirty(t, db, migrations.M2Version)
}

func TestPCADownRefusesUnknownDependency(t *testing.T) {
	for _, ddl := range []string{
		`CREATE INDEX pca_extra ON notifications(message)`,
		`ALTER TABLE employee_profile_change_requests ADD COLUMN pca_extra text`,
		`CREATE FUNCTION pca_extra() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RETURN NEW; END $$; CREATE TRIGGER pca_extra BEFORE INSERT ON idempotency_records FOR EACH ROW EXECUTE FUNCTION pca_extra()`,
		`ALTER TABLE notifications ADD CONSTRAINT pca_extra CHECK(length(message)>0)`,
		`ALTER TABLE notifications ADD CONSTRAINT ck_pca_values CHECK(length(message)>0)`,
		`CREATE VIEW pca_extra AS SELECT message FROM notifications`,
		`CREATE FUNCTION pca_extra() RETURNS int LANGUAGE plpgsql AS $$ BEGIN RETURN (SELECT count(*)::int FROM public.notifications); END $$`,
	} {
		t.Run(fmt.Sprintf("dependency-%d", len(ddl)), func(t *testing.T) {
			db, m := pcaLegacy(t)
			authdb.Require(t, "nâng PCA", m.Up())
			exec(t, db.Conn, ddl)
			if m.Steps(-1) == nil {
				t.Fatal("down không được xóa dependency chưa review")
			}
			m.Close(t)
			assertPCAPresent(t, db, true)
			assertPCADirty(t, db, migrations.M2Version)
			var remains bool
			authdb.Require(t, "dependency được giữ", db.Conn.QueryRow(context.Background(), `SELECT to_regclass('public.pca_extra') IS NOT NULL OR to_regprocedure('public.pca_extra()') IS NOT NULL OR EXISTS(SELECT 1 FROM pg_constraint WHERE conrelid='public.notifications'::regclass AND conname IN ('pca_extra','ck_pca_values')) OR EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND column_name='pca_extra')`).Scan(&remains))
			if !remains {
				t.Fatal("down đã xóa dependency lạ")
			}
		})
	}
}
