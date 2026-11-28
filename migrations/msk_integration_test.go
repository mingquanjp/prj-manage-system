//go:build integration

package migrations_test

import (
	"context"
	"encoding/json"
	"errors"
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

func TestMSKEmptyDownUp(t *testing.T) {
	db := authdb.NewCase(t)
	m := db.OpenMigrator(t, migrations.FS)
	authdb.Require(t, "nâng DB rỗng qua toàn bộ lịch sử", m.Up())
	head(t, m, migrations.MSKVersion)
	assertMSKPresent(t, db, true)
	assertM2Present(t, db)
	assertPCAPresent(t, db, true)
	if !errors.Is(m.Up(), migrate.ErrNoChange) {
		t.Fatal("up head phải không thay đổi")
	}
	authdb.Require(t, "down MSK trên DB case rỗng", m.Steps(-1))
	head(t, m, migrations.PCAVersion)
	assertMSKPresent(t, db, false)
	assertPCAPresent(t, db, true)
	authdb.Require(t, "up MSK lại", m.Up())
	head(t, m, migrations.MSKVersion)
	assertMSKPresent(t, db, true)
}

func mskLegacy(t *testing.T, official bool) (*authdb.Case, *authdb.Migrator) {
	t.Helper()
	db, previous := pcaLegacy(t)
	authdb.Require(t, "dựng mốc PCA lịch sử", previous.Migrate(migrations.PCAVersion))
	previous.Close(t)
	exec(t, db.Conn, `INSERT INTO skill_categories(category_id,category_name,created_at,updated_at)
 OVERRIDING SYSTEM VALUE VALUES(11,'Nhóm kỹ năng tổng hợp','2026-01-01Z','2026-01-02Z');
 INSERT INTO skills(skill_id,skill_name,category_id,created_at,updated_at)
 OVERRIDING SYSTEM VALUE VALUES(21,'Kỹ năng tổng hợp A',11,'2026-01-01Z','2026-01-02Z'),
 (22,'Kỹ năng tổng hợp B',11,'2026-01-01Z','2026-01-02Z');
 INSERT INTO employee_profile_change_requests(employee_id,requested_by,before_values,proposed_values)
 VALUES(201,101,'{"schema_version":1,"address":null}','{"schema_version":1,"address":"Địa chỉ tổng hợp"}');
 INSERT INTO idempotency_records(actor_id,operation,idempotency_key,request_hash,result_reference)
 VALUES(101,'PROFILE_CHANGE_SUBMIT','synthetic-pca-key',decode(repeat('ab',32),'hex'),'{"request_id":1}');
 INSERT INTO notifications(recipient_id,notification_type,resource_type,resource_id,message)
 VALUES(101,'PROFILE_CHANGE_SUBMITTED','PROFILE_CHANGE_REQUEST',1,'Có đề nghị hồ sơ')`)
	if official {
		// Người duyệt lịch sử không còn quyền/trạng thái hiện hành; ID employee khác user.
		exec(t, db.Conn, `INSERT INTO employee_skills(employee_skill_id,employee_id,skill_id,
 proficiency_level,years_experience,note,source,approved_by_employee_id,approved_at,created_at,updated_at)
 OVERRIDING SYSTEM VALUE VALUES
 (31,201,21,'ADVANCED',4.5,$1,'CV',202,'2026-02-03Z','2026-01-01Z','2026-02-04Z'),
 (32,202,22,'BASIC',NULL,NULL,'MANUAL',203,'2026-02-05Z','2026-01-02Z','2026-02-06Z')`, "  "+strings.Repeat("ế", 2001)+"  ")
	}
	return db, db.OpenMSKMigrator(t, migrations.FS, 101)
}

func TestMSKLegacyPreservesValuesApprovalAndExistingDomains(t *testing.T) {
	db, m := mskLegacy(t, true)
	before := mskPreservedSnapshot(t, db, true)
	var began time.Time
	authdb.Require(t, "thời điểm trước import", db.Conn.QueryRow(context.Background(), `SELECT clock_timestamp()`).Scan(&began))
	authdb.Require(t, "nâng legacy lên MSK", m.Up())
	head(t, m, migrations.MSKVersion)
	assertMSKPresent(t, db, true)
	sameJSON(t, before, mskPreservedSnapshot(t, db, false))
	assertM2Sequence(t, db, 1, false)
	var imported bool
	authdb.Require(t, "nguồn cũ được lưu đúng actor và mapping", db.Conn.QueryRow(context.Background(), `SELECT
 (SELECT count(*)=2 FROM audit_events WHERE operation='SKILL_LEGACY_IMPORTED')
 AND (SELECT count(*)=2 FROM audit_events a JOIN employee_skills s ON s.employee_skill_id=a.entity_id
 WHERE a.operation='SKILL_LEGACY_IMPORTED' AND a.entity_type='EMPLOYEE_SKILL' AND a.actor_id=101
 AND a.occurred_at >= $1 AND a.occurred_at <= clock_timestamp()
 AND a.before_values->>'source'=CASE s.employee_skill_id WHEN 31 THEN 'CV' ELSE 'MANUAL' END
 AND (a.before_values->>'approved_by_employee_id')::int=CASE s.employee_skill_id WHEN 31 THEN 202 ELSE 203 END
 AND (a.before_values->>'approved_at')::timestamptz=s.approved_at
 AND (a.after_values->>'approved_by')::int=s.approved_by)
 AND (SELECT approved_by=102 FROM employee_skills WHERE employee_skill_id=31)
 AND (SELECT approved_by=103 FROM employee_skills WHERE employee_skill_id=32)
 AND NOT EXISTS(SELECT 1 FROM employee_skill_change_requests)`, began).Scan(&imported))
	if !imported {
		t.Fatal("import phải giữ provenance, không giả người duyệt hoặc request")
	}
	if !errors.Is(m.Up(), migrate.ErrNoChange) {
		t.Fatal("up lặp lại không được import thêm")
	}
	sameJSON(t, before, mskPreservedSnapshot(t, db, false))
}

func TestMSKPreflightRefusesBeforeDirty(t *testing.T) {
	for _, tc := range []struct {
		name  string
		sql   string
		actor int
	}{
		{"missing import actor", "", 0},
		{"unknown import actor", "", 99999},
		{"inactive skill", `UPDATE skills SET status='INACTIVE' WHERE skill_id=21`, 101},
		{"missing historical actor", `UPDATE employee_skills SET approved_by_employee_id=NULL WHERE employee_skill_id=31`, 101},
		{"missing approval time", `UPDATE employee_skills SET approved_at=NULL WHERE employee_skill_id=31`, 101},
		{"category blank", `UPDATE skill_categories SET category_name='   '`, 101},
		{"category Unicode padding", `UPDATE skill_categories SET category_name=U&'\00A0Nhóm'`, 101},
		{"skill trailing padding", `UPDATE skills SET skill_name='Skill ' WHERE skill_id=21`, 101},
		{"category normalized duplicate", `INSERT INTO skill_categories(category_name) VALUES('NHÓM KỸ NĂNG TỔNG HỢP')`, 101},
		{"skill normalized duplicate", `INSERT INTO skills(skill_name,category_id) VALUES('KỸ NĂNG TỔNG HỢP A',11)`, 101},
		{"preexisting request", `CREATE TABLE employee_skill_change_requests(unreviewed int)`, 101},
		{"status type drift", `ALTER TABLE skills ALTER COLUMN status DROP DEFAULT; ALTER TABLE skills ALTER COLUMN status TYPE text USING status::text`, 101},
		{"source type drift", `ALTER TABLE employee_skills ALTER COLUMN source DROP DEFAULT; ALTER TABLE employee_skills ALTER COLUMN source TYPE text USING source::text`, 101},
		{"extra official column", `ALTER TABLE employee_skills ADD COLUMN unreviewed text`, 101},
		{"view on removed status", `CREATE VIEW msk_dependency AS SELECT status FROM skills`, 101},
		{"view on removed source", `CREATE VIEW msk_dependency AS SELECT source FROM employee_skills`, 101},
		{"function on removed source", `CREATE FUNCTION msk_dependency() RETURNS text LANGUAGE sql AS $$ SELECT source::text FROM employee_skills LIMIT 1 $$`, 101},
		{"custom trigger", `CREATE FUNCTION msk_dependency() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RETURN NEW; END $$; CREATE TRIGGER msk_dependency BEFORE UPDATE ON employee_skills FOR EACH ROW EXECUTE FUNCTION msk_dependency()`, 101},
		{"FK action drift", `ALTER TABLE employee_skills DROP CONSTRAINT employee_skills_skill_id_fkey; ALTER TABLE employee_skills ADD FOREIGN KEY(skill_id) REFERENCES skills(skill_id) ON DELETE CASCADE`, 101},
		{"preexisting index", `CREATE INDEX uq_msk_pending ON skills(skill_id)`, 101},
		{"deferrable unique", `ALTER TABLE skill_categories DROP CONSTRAINT skill_categories_category_name_key;
 ALTER TABLE skill_categories ADD UNIQUE(category_name) DEFERRABLE`, 101},
		{"check reuses historical name", `ALTER TABLE skill_categories ADD CONSTRAINT employee_skills_years_experience_check CHECK(length(category_name)>0)`, 101},
		{"unexpected inbound FK from existing table", `ALTER TABLE task_requirements ADD COLUMN category_ref int REFERENCES skill_categories(category_id)`, 101},
		{"generated column", `ALTER TABLE skills ADD COLUMN generated_name text GENERATED ALWAYS AS(lower(skill_name)) STORED`, 101},
		{"text validator collision", `CREATE FUNCTION msk_text_valid(text,int) RETURNS boolean LANGUAGE sql IMMUTABLE AS $$ SELECT true $$`, 101},
		{"values validator collision", `CREATE FUNCTION msk_values_valid(jsonb) RETURNS boolean LANGUAGE sql IMMUTABLE AS $$ SELECT true $$`, 101},
		{"request validator collision", `CREATE FUNCTION msk_request_values_valid(text,jsonb,jsonb) RETURNS boolean LANGUAGE sql IMMUTABLE AS $$ SELECT true $$`, 101},
		{"audit drops import row", `CREATE FUNCTION msk_audit_dependency() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RETURN NULL; END $$;
 CREATE TRIGGER msk_audit_dependency BEFORE INSERT ON audit_events FOR EACH ROW EXECUTE FUNCTION msk_audit_dependency()`, 101},
		{"audit changes import payload", `CREATE FUNCTION msk_audit_dependency() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN NEW.before_values='{}'::jsonb; RETURN NEW; END $$;
 CREATE TRIGGER msk_audit_dependency BEFORE INSERT ON audit_events FOR EACH ROW EXECUTE FUNCTION msk_audit_dependency()`, 101},
		{"audit column drift", `ALTER TABLE audit_events ALTER COLUMN entity_type TYPE varchar(50)`, 101},
		{"audit default drift", `ALTER TABLE audit_events ALTER COLUMN entity_type SET DEFAULT 'UNREVIEWED'`, 101},
		{"audit FK drift", `ALTER TABLE audit_events DROP CONSTRAINT audit_events_actor_id_fkey;
 ALTER TABLE audit_events ADD FOREIGN KEY(actor_id) REFERENCES users(user_id) ON DELETE CASCADE`, 101},
		{"missing official unique hidden by duplicate category unique", `ALTER TABLE employee_skills DROP CONSTRAINT uq_employee_skills;
 ALTER TABLE skill_categories ADD CONSTRAINT msk_duplicate_category_unique UNIQUE(category_name)`, 101},
		{"missing official FK hidden by duplicate category FK", `ALTER TABLE employee_skills DROP CONSTRAINT employee_skills_skill_id_fkey;
 ALTER TABLE skills ADD CONSTRAINT msk_duplicate_category_fk FOREIGN KEY(category_id) REFERENCES skill_categories(category_id)`, 101},
		{"shadow proficiency enum", `CREATE SCHEMA msk_shadow;
 CREATE TYPE msk_shadow.proficiency_level AS ENUM('BASIC','INTERMEDIATE','ADVANCED');
 ALTER TABLE employee_skills ALTER COLUMN proficiency_level TYPE msk_shadow.proficiency_level
 USING proficiency_level::text::msk_shadow.proficiency_level`, 101},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, configured := mskLegacy(t, true)
			configured.Close(t)
			if tc.sql != "" {
				exec(t, db.Conn, tc.sql)
			}
			before := mskRawSnapshot(t, db)
			m := db.OpenMSKMigrator(t, migrations.FS, tc.actor)
			if err := m.Up(); err == nil || !strings.HasPrefix(err.Error(), "MSK:") {
				t.Fatal("preflight phải từ chối trước dirty, không xuất dữ liệu lỗi")
			}
			head(t, m, migrations.PCAVersion)
			sameJSON(t, before, mskRawSnapshot(t, db))
		})
	}
}

const mskRequestInsert = `INSERT INTO employee_skill_change_requests
 (employee_id,skill_id,operation,before_values,proposed_values,explanation,requested_by)
 VALUES(201,21,$1,$2::jsonb,$3::jsonb,'Giải trình tổng hợp',101)`

const mskBasic = `{"schema_version":1,"proficiency_level":"BASIC","years_experience":null,"note":null}`
const mskAdvanced = `{"schema_version":1,"proficiency_level":"ADVANCED","years_experience":"4.5","note":"Ghi chú tổng hợp"}`

func TestMSKRequestJSONStateFKAndIndexes(t *testing.T) {
	db, m := mskLegacy(t, false)
	authdb.Require(t, "nâng schema MSK", m.Up())
	assertMSKPresent(t, db, true)
	for _, tc := range []struct {
		name      string
		operation string
		before    any
		proposed  any
	}{
		{"ADD JSON null before", "ADD", `null`, mskBasic},
		{"ADD object before", "ADD", mskAdvanced, mskBasic},
		{"UPDATE SQL NULL before", "UPDATE", nil, mskBasic},
		{"UPDATE JSON null before", "UPDATE", `null`, mskBasic},
		{"UPDATE no changes", "UPDATE", mskBasic, mskBasic},
		{"unknown operation", "REMOVE", nil, mskBasic},
		{"proposed scalar", "ADD", nil, `4`},
		{"proposed JSON null", "ADD", nil, `null`},
		{"proposed SQL NULL", "ADD", nil, nil},
		{"proposed empty", "ADD", nil, `{}`},
		{"proposed missing note", "ADD", nil, `{"schema_version":1,"proficiency_level":"BASIC","years_experience":null}`},
		{"proposed missing level", "ADD", nil, `{"schema_version":1,"years_experience":null,"note":null}`},
		{"proposed all NULL", "ADD", nil, `{"schema_version":1,"proficiency_level":null,"years_experience":null,"note":null}`},
		{"proposed unknown key", "ADD", nil, `{"schema_version":1,"proficiency_level":"BASIC","years_experience":null,"note":null,"actor":101}`},
		{"schema version JSON null", "ADD", nil, `{"schema_version":null,"proficiency_level":"BASIC","years_experience":null,"note":null}`},
		{"schema version string", "ADD", nil, `{"schema_version":"1","proficiency_level":"BASIC","years_experience":null,"note":null}`},
		{"schema version future", "ADD", nil, `{"schema_version":2,"proficiency_level":"BASIC","years_experience":null,"note":null}`},
		{"unknown level", "ADD", nil, `{"schema_version":1,"proficiency_level":"MASTER","years_experience":null,"note":null}`},
		{"level wrong type", "ADD", nil, `{"schema_version":1,"proficiency_level":1,"years_experience":null,"note":null}`},
		{"years number", "ADD", nil, `{"schema_version":1,"proficiency_level":"BASIC","years_experience":1,"note":null}`},
		{"years scale", "ADD", nil, `{"schema_version":1,"proficiency_level":"BASIC","years_experience":"1.11","note":null}`},
		{"years integer string", "ADD", nil, `{"schema_version":1,"proficiency_level":"BASIC","years_experience":"1","note":null}`},
		{"years negative", "ADD", nil, `{"schema_version":1,"proficiency_level":"BASIC","years_experience":"-1.0","note":null}`},
		{"years overflow", "ADD", nil, `{"schema_version":1,"proficiency_level":"BASIC","years_experience":"1000.0","note":null}`},
		{"years leading zero", "ADD", nil, `{"schema_version":1,"proficiency_level":"BASIC","years_experience":"01.0","note":null}`},
		{"note wrong type", "ADD", nil, `{"schema_version":1,"proficiency_level":"BASIC","years_experience":null,"note":5}`},
		{"note blank", "ADD", nil, mskValues("BASIC", nil, " ")},
		{"note empty", "ADD", nil, mskValues("BASIC", nil, "")},
		{"note padding", "ADD", nil, mskValues("BASIC", nil, " note")},
		{"note too long", "ADD", nil, mskValues("BASIC", nil, strings.Repeat("ế", 2001))},
		{"UPDATE forged long note", "UPDATE", mskBasic, mskValues("BASIC", nil, strings.Repeat("ế", 2001))},
		{"before missing schema version", "UPDATE", `{"proficiency_level":"BASIC","years_experience":null,"note":null}`, mskAdvanced},
		{"before years wrong type", "UPDATE", `{"schema_version":1,"proficiency_level":"BASIC","years_experience":1,"note":null}`, mskAdvanced},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := "23514"
			if tc.proposed == nil {
				state = "23502"
			}
			expectState(t, db.Conn, state, mskRequestInsert, tc.operation, tc.before, tc.proposed)
		})
	}
	exec(t, db.Conn, mskRequestInsert, "ADD", nil, mskBasic)
	for _, change := range []string{
		`source='CV'`, `version=0`, `version=-1`, `status='NO_CHANGE'`,
		`explanation=''`, `explanation=' x'`, `explanation=repeat('ế',2001)`,
		`external_evidence=''`, `external_evidence=' x'`, `external_evidence=repeat('ế',2001)`,
		`expert_opinion='Ý kiến khi chưa quyết định'`, `resolved_by=102`, `resolved_at=clock_timestamp()`,
		`decision_reason='Lý do khi chưa quyết định'`, `status='APPROVED'`, `status='REJECTED'`,
		`status='APPROVED',resolved_by=102,resolved_at=clock_timestamp()`,
		`status='REJECTED',resolved_by=102,resolved_at=clock_timestamp(),decision_reason=''`,
		`status='REJECTED',resolved_by=101,resolved_at=clock_timestamp(),decision_reason='Lý do tổng hợp'`,
		`status='APPROVED',resolved_by=102,resolved_at=requested_at-interval '1s',decision_reason='Lý do tổng hợp'`,
		`status='REJECTED',resolved_by=102,resolved_at=clock_timestamp(),decision_reason='Lý do tổng hợp',expert_opinion=' '`,
	} {
		expectState(t, db.Conn, "23514", `UPDATE employee_skill_change_requests SET `+change)
	}
	for _, change := range []string{`employee_id=101`, `skill_id=99999`, `requested_by=99999`} {
		expectState(t, db.Conn, "23503", `UPDATE employee_skill_change_requests SET `+change)
	}
	expectState(t, db.Conn, "23505", mskRequestInsert, "ADD", nil, mskAdvanced)
	exec(t, db.Conn, `UPDATE employee_skill_change_requests SET status='REJECTED',resolved_by=102,
 resolved_at=clock_timestamp(),decision_reason='Lý do tổng hợp',expert_opinion='Ý kiến ngoài tổng hợp'`)
	exec(t, db.Conn, mskRequestInsert, "UPDATE", mskBasic, mskAdvanced)
	exec(t, db.Conn, `UPDATE employee_skill_change_requests SET status='APPROVED',resolved_by=103,
 resolved_at=clock_timestamp(),decision_reason='Lý do tổng hợp' WHERE status='PENDING'`)
	expectState(t, db.Conn, "23503", `DELETE FROM skills WHERE skill_id=21`)
	expectState(t, db.Conn, "23503", `DELETE FROM employees WHERE employee_id=201`)
}

func mskValues(level string, years any, note any) string {
	payload, err := json.Marshal(map[string]any{"schema_version": 1, "proficiency_level": level, "years_experience": years, "note": note})
	if err != nil {
		panic("fixture values không mã hóa được")
	}
	return string(payload)
}

func TestMSKLegacyNoteExceptionAndNameBoundaries(t *testing.T) {
	db, m := mskLegacy(t, false)
	authdb.Require(t, "nâng MSK", m.Up())
	for _, note := range []string{"", " ", "  note legacy  ", strings.Repeat("ế", 2001)} {
		before := mskValues("BASIC", nil, note)
		proposed := mskValues("ADVANCED", "0.0", note)
		exec(t, db.Conn, mskRequestInsert, "UPDATE", before, proposed)
		exec(t, db.Conn, `UPDATE employee_skill_change_requests SET status='APPROVED',resolved_by=102,
 resolved_at=clock_timestamp(),decision_reason='Lý do tổng hợp' WHERE status='PENDING'`)
	}
	for _, years := range []any{nil, "0.0", "999.9"} {
		exec(t, db.Conn, mskRequestInsert, "ADD", nil, mskValues("EXPERT", years, strings.Repeat("ế", 2000)))
		exec(t, db.Conn, `UPDATE employee_skill_change_requests SET status='REJECTED',resolved_by=102,
 resolved_at=clock_timestamp(),decision_reason='Lý do tổng hợp' WHERE status='PENDING'`)
	}
	for _, table := range []struct{ name, column string }{{"skill_categories", "category_name"}, {"skills", "skill_name"}} {
		for _, name := range []string{"", " ", "\u00a0", " leading", "trailing\u3000", strings.Repeat("ế", 101)} {
			query := `INSERT INTO ` + table.name + `(` + table.column + `) VALUES($1)`
			if table.name == "skills" {
				query = `INSERT INTO skills(skill_name,category_id) VALUES($1,11)`
			}
			state := "23514"
			if len([]rune(name)) > 100 {
				state = "22001"
			}
			expectState(t, db.Conn, state, query, name)
		}
	}
	exec(t, db.Conn, `INSERT INTO skill_categories(category_name) VALUES($1)`, strings.Repeat("ế", 100))
	exec(t, db.Conn, `INSERT INTO skills(skill_name,category_id) VALUES($1,11)`, strings.Repeat("ế", 100))
	expectState(t, db.Conn, "23505", `INSERT INTO skill_categories(category_name) VALUES('NHÓM KỸ NĂNG TỔNG HỢP')`)
	expectState(t, db.Conn, "23505", `INSERT INTO skills(skill_name,category_id) VALUES('KỸ NĂNG TỔNG HỢP A',11)`)
	for _, space := range []string{"\t", "\n", "\u0085", "\u00a0", "\u1680", "\u2000", "\u2028", "\u202f", "\u205f", "\u3000"} {
		expectState(t, db.Conn, "23514", mskRequestInsert, "ADD", nil, mskValues("BASIC", nil, space+"x"))
	}
}

func TestMSKNameNormalizationParityWithPostgres(t *testing.T) {
	db, m := mskLegacy(t, false)
	for _, name := range []string{"Nhóm kỹ năng", "ĐẶNG", "Ế", "Σ", "İ"} {
		var lower string
		authdb.Require(t, "lower theo collation PostgreSQL test", db.Conn.QueryRow(context.Background(),
			`SELECT lower($1::text)`, name).Scan(&lower))
		if strings.ToLower(name) != lower {
			t.Fatal("Go và collation PG chuẩn hóa tên khác nhau; cần review mapping")
		}
	}
	authdb.Require(t, "nâng legacy có tên Unicode", m.Up())
}

func TestMSKDownRefusesEachDataKindBeforeDDL(t *testing.T) {
	for _, kind := range []string{"skill", "official", "request", "key", "notification", "catalog audit", "import archive"} {
		t.Run(kind, func(t *testing.T) {
			db, m := mskLegacy(t, false)
			authdb.Require(t, "nâng MSK", m.Up())
			switch kind {
			case "official":
				exec(t, db.Conn, `INSERT INTO employee_skills(employee_id,skill_id,proficiency_level,approved_by,approved_at)
 VALUES(201,21,'BASIC',102,'2026-02-03Z')`)
			case "request":
				exec(t, db.Conn, mskRequestInsert, "ADD", nil, mskBasic)
			case "key":
				exec(t, db.Conn, `DELETE FROM skills; INSERT INTO idempotency_records(actor_id,operation,idempotency_key,request_hash,result_reference)
 VALUES(101,'SKILL_CHANGE_SUBMIT','synthetic-msk-key',decode(repeat('ab',32),'hex'),'{"request_id":1}')`)
			case "notification":
				exec(t, db.Conn, `DELETE FROM skills; INSERT INTO notifications(recipient_id,notification_type,resource_type,resource_id,message)
 VALUES(101,'SKILL_CHANGE_SUBMITTED','SKILL_CHANGE_REQUEST',1,'Có đề nghị kỹ năng')`)
			case "catalog audit":
				exec(t, db.Conn, `DELETE FROM skills; INSERT INTO audit_events(actor_id,entity_type,entity_id,operation)
 VALUES(101,'SKILL',21,'CATALOG_CREATED')`)
			case "import archive":
				exec(t, db.Conn, `DELETE FROM skills; INSERT INTO audit_events(actor_id,entity_type,entity_id,operation)
 VALUES(101,'EMPLOYEE_SKILL',31,'SKILL_LEGACY_IMPORTED')`)
			}
			before := mskRawSnapshot(t, db)
			if m.Steps(-1) == nil {
				t.Fatal("down phải từ chối dữ liệu MSK")
			}
			m.Close(t)
			assertMSKPresent(t, db, true)
			sameJSON(t, before, mskRawSnapshot(t, db))
			assertPCADirty(t, db, migrations.PCAVersion)
		})
	}
}

func TestMSKDDLFailureAndPostPreflightGuard(t *testing.T) {
	for _, stage := range []string{"after DDL", "after preflight"} {
		t.Run(stage, func(t *testing.T) {
			db, configured := mskLegacy(t, true)
			configured.Close(t)
			before := mskRawSnapshot(t, db)
			source := mskFaultSource(t, "000007_manual_skills.up.sql", stage)
			broken := db.OpenMSKMigrator(t, source, 101)
			var upErr error
			if stage == "after preflight" {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				exec(t, db.Conn, `SELECT pg_advisory_lock(638051)`)
				defer db.Conn.Exec(context.Background(), `SELECT pg_advisory_unlock(638051)`)
				result := make(chan error, 1)
				go func() { result <- broken.Up() }()
				// Lịch chạy chỉ đi tiếp khi PG xác nhận migration đã qua preflight và chờ barrier.
				for {
					var waiting bool
					authdb.Require(t, "barrier sau preflight", db.Conn.QueryRow(ctx, `SELECT EXISTS(
 SELECT 1 FROM pg_locks WHERE locktype='advisory' AND objid=638051 AND NOT granted)`).Scan(&waiting))
					if waiting {
						break
					}
					select {
					case <-ctx.Done():
						t.Fatal("migration không tới barrier")
					default:
						runtime.Gosched()
					}
				}
				exec(t, db.Conn, `UPDATE skills SET status='INACTIVE' WHERE skill_id=21`)
				before = mskRawSnapshot(t, db)
				exec(t, db.Conn, `SELECT pg_advisory_unlock(638051)`)
				select {
				case upErr = <-result:
				case <-ctx.Done():
					t.Fatal("guard DDL không kết thúc")
				}
			} else {
				upErr = broken.Up()
			}
			if upErr == nil {
				t.Fatal("fault/guard phải làm migration thất bại")
			}
			broken.Close(t)
			assertMSKPresent(t, db, false)
			sameJSON(t, before, mskRawSnapshot(t, db))
			assertPCADirty(t, db, migrations.MSKVersion)
			retry := db.OpenMSKMigrator(t, migrations.FS, 101)
			var dirty migrate.ErrDirty
			if !errors.As(retry.Up(), &dirty) || dirty.Version != migrations.MSKVersion {
				t.Fatal("driver phải giữ dirty và từ chối retry, không force")
			}
		})
	}
}

func TestMSKDownDDLFailureRollsBack(t *testing.T) {
	db := authdb.NewCase(t)
	m := db.OpenMigrator(t, migrations.FS)
	authdb.Require(t, "nâng MSK rỗng", m.Up())
	m.Close(t)
	before := mskRawSnapshot(t, db)
	broken := db.OpenMigrator(t, mskFaultSource(t, "000007_manual_skills.down.sql", "after DDL"))
	if broken.Steps(-1) == nil {
		t.Fatal("down fault phải thất bại")
	}
	broken.Close(t)
	assertMSKPresent(t, db, true)
	sameJSON(t, before, mskRawSnapshot(t, db))
	assertPCADirty(t, db, migrations.PCAVersion)
}

func TestMSKDownRefusesUnreviewedDependency(t *testing.T) {
	for _, tc := range []struct{ name, sql string }{
		{"extra column", `ALTER TABLE employee_skill_change_requests ADD COLUMN msk_down_dependency text`},
		{"extra index", `CREATE INDEX msk_down_dependency ON employee_skill_change_requests(explanation)`},
		{"extra check", `ALTER TABLE employee_skill_change_requests ADD CONSTRAINT msk_down_dependency CHECK(length(explanation)>0)`},
		{"view", `CREATE VIEW msk_down_dependency AS SELECT request_id FROM employee_skill_change_requests`},
		{"function body", `CREATE FUNCTION msk_down_dependency() RETURNS int LANGUAGE sql AS $$ SELECT count(*)::int FROM employee_skill_change_requests $$`},
		{"trigger", `CREATE FUNCTION msk_down_dependency() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RETURN NEW; END $$;
 CREATE TRIGGER msk_down_dependency BEFORE INSERT ON employee_skill_change_requests FOR EACH ROW EXECUTE FUNCTION msk_down_dependency()`},
		{"approval index", `CREATE INDEX msk_down_dependency ON employee_skills(approved_by)`},
		{"approval check", `ALTER TABLE employee_skills ADD CONSTRAINT msk_down_dependency CHECK(approved_by>0)`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := authdb.NewCase(t)
			m := db.OpenMigrator(t, migrations.FS)
			authdb.Require(t, "nâng MSK rỗng", m.Up())
			exec(t, db.Conn, tc.sql)
			before := mskRawSnapshot(t, db)
			if m.Steps(-1) == nil {
				t.Fatal("down phải từ chối dependency chưa review")
			}
			m.Close(t)
			assertMSKPresent(t, db, true)
			sameJSON(t, before, mskRawSnapshot(t, db))
			assertPCADirty(t, db, migrations.PCAVersion)
		})
	}
}

func mskFaultSource(t *testing.T, filename, stage string) fs.FS {
	t.Helper()
	source := fstest.MapFS{}
	entries, err := fs.ReadDir(migrations.FS, ".")
	authdb.Require(t, "đọc nguồn migration", err)
	found := false
	for _, entry := range entries {
		content, err := fs.ReadFile(migrations.FS, entry.Name())
		authdb.Require(t, "đọc file migration", err)
		if entry.Name() == filename {
			found = true
			if stage == "after DDL" {
				if strings.Count(string(content), "COMMIT;") != 1 {
					t.Fatal("fault cần đúng một transaction migration")
				}
				content = []byte(strings.Replace(string(content), "COMMIT;", "SELECT 1/0;\nCOMMIT;", 1))
			} else {
				content = []byte("SELECT pg_advisory_lock(638051);\n" + string(content))
			}
		}
		source[entry.Name()] = &fstest.MapFile{Data: content}
	}
	if !found {
		t.Fatal("không tìm thấy migration cần fault")
	}
	return source
}

func mskPreservedSnapshot(t *testing.T, db *authdb.Case, legacy bool) string {
	t.Helper()
	skills := `to_jsonb(s)`
	official := `to_jsonb(s)`
	if legacy {
		skills = `to_jsonb(s)-'status'`
		official = `(to_jsonb(s)-'source'-'approved_by_employee_id') || jsonb_build_object('approved_by',
 (SELECT e.user_id FROM employees e WHERE e.employee_id=s.approved_by_employee_id))`
	}
	var snapshot string
	authdb.Require(t, "snapshot bảo toàn trong bộ nhớ", db.Conn.QueryRow(context.Background(), `SELECT jsonb_build_object(
 'users',(SELECT jsonb_agg(to_jsonb(u) ORDER BY user_id) FROM users u),
 'employees',(SELECT jsonb_agg(to_jsonb(e) ORDER BY employee_id) FROM employees e),
 'roles',(SELECT jsonb_agg(to_jsonb(r) ORDER BY user_id,role_code) FROM user_roles r),
 'sessions',(SELECT jsonb_agg(to_jsonb(s) ORDER BY session_id) FROM auth_sessions s),
 'departments',(SELECT jsonb_agg(to_jsonb(d) ORDER BY department_id) FROM departments d),
 'positions',(SELECT jsonb_agg(to_jsonb(p) ORDER BY position_id) FROM positions p),
 'levels',(SELECT jsonb_agg(to_jsonb(l) ORDER BY job_level_id) FROM job_levels l),
 'categories',(SELECT jsonb_agg(to_jsonb(c) ORDER BY category_id) FROM skill_categories c),
 'skills',(SELECT jsonb_agg(`+skills+` ORDER BY skill_id) FROM skills s),
 'official',(SELECT jsonb_agg(`+official+` ORDER BY employee_skill_id) FROM employee_skills s),
 'requests',(SELECT jsonb_agg(to_jsonb(r) ORDER BY request_id) FROM employee_profile_change_requests r),
 'keys',(SELECT jsonb_agg(to_jsonb(k) ORDER BY idempotency_id) FROM idempotency_records k),
 'notifications',(SELECT jsonb_agg(to_jsonb(n) ORDER BY notification_id) FROM notifications n),
 'audit',(SELECT jsonb_agg(to_jsonb(a) ORDER BY audit_id) FROM audit_events a WHERE operation<>'SKILL_LEGACY_IMPORTED'))::text`).Scan(&snapshot))
	return snapshot
}

func mskRawSnapshot(t *testing.T, db *authdb.Case) string {
	t.Helper()
	var hasRequests bool
	authdb.Require(t, "phạm vi snapshot request MSK", db.Conn.QueryRow(context.Background(),
		`SELECT to_regclass('public.employee_skill_change_requests') IS NOT NULL`).Scan(&hasRequests))
	requests := `NULL::jsonb`
	if hasRequests {
		requests = `(SELECT jsonb_agg(to_jsonb(r) ORDER BY to_jsonb(r)::text) FROM employee_skill_change_requests r)`
	}
	var snapshot string
	authdb.Require(t, "snapshot raw chỉ trong bộ nhớ", db.Conn.QueryRow(context.Background(), `SELECT jsonb_build_object(
 'users',(SELECT jsonb_agg(to_jsonb(u) ORDER BY user_id) FROM users u),
 'employees',(SELECT jsonb_agg(to_jsonb(e) ORDER BY employee_id) FROM employees e),
 'categories',(SELECT jsonb_agg(to_jsonb(c) ORDER BY category_id) FROM skill_categories c),
 'skills',(SELECT jsonb_agg(to_jsonb(s) ORDER BY skill_id) FROM skills s),
 'official',(SELECT jsonb_agg(to_jsonb(s) ORDER BY employee_skill_id) FROM employee_skills s),
 'skill_requests',`+requests+`,
 'requests',(SELECT jsonb_agg(to_jsonb(r) ORDER BY request_id) FROM employee_profile_change_requests r),
 'keys',(SELECT jsonb_agg(to_jsonb(k) ORDER BY idempotency_id) FROM idempotency_records k),
 'notifications',(SELECT jsonb_agg(to_jsonb(n) ORDER BY notification_id) FROM notifications n),
 'audit',(SELECT jsonb_agg(to_jsonb(a) ORDER BY audit_id) FROM audit_events a))::text`).Scan(&snapshot))
	return snapshot
}

func assertMSKPresent(t *testing.T, db *authdb.Case, present bool) {
	t.Helper()
	var ok bool
	if present {
		authdb.Require(t, "schema MSK head", db.Conn.QueryRow(context.Background(), `SELECT
 to_regclass('public.employee_skill_change_requests') IS NOT NULL
 AND (SELECT is_nullable='NO' AND data_type='integer' FROM information_schema.columns
 WHERE table_schema='public' AND table_name='employee_skills' AND column_name='approved_by')
 AND (SELECT is_nullable='NO' FROM information_schema.columns
 WHERE table_schema='public' AND table_name='employee_skills' AND column_name='approved_at')
 AND NOT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema='public'
 AND ((table_name='employee_skills' AND column_name IN ('source','approved_by_employee_id'))
 OR (table_name='skills' AND column_name='status')))
 AND to_regclass('public.idx_skills_status') IS NULL
 AND to_regtype('public.skill_source') IS NOT NULL AND to_regtype('public.skill_status') IS NOT NULL
 AND (SELECT count(*)=3 FROM pg_constraint WHERE contype='f' AND conrelid='public.employee_skills'::regclass
 AND confdeltype='a' AND confupdtype='a' AND convalidated AND NOT condeferrable)
 AND (SELECT count(*)=4 FROM pg_constraint WHERE contype='f' AND conrelid='public.employee_skill_change_requests'::regclass
 AND confdeltype='a' AND confupdtype='a' AND convalidated AND NOT condeferrable)
 AND (SELECT count(*)=3 FROM pg_indexes WHERE schemaname='public' AND indexdef IN (
 'CREATE INDEX idx_msk_owner ON public.employee_skill_change_requests USING btree (employee_id, requested_at DESC, request_id DESC)',
 'CREATE INDEX idx_msk_review ON public.employee_skill_change_requests USING btree (status, requested_at DESC, request_id DESC)',
 'CREATE INDEX idx_msk_employee_skill ON public.employee_skill_change_requests USING btree (employee_id, skill_id, requested_at DESC, request_id DESC)'))
 AND EXISTS(SELECT 1 FROM pg_index i JOIN pg_class c ON c.oid=i.indexrelid
 WHERE c.relname='uq_msk_pending' AND i.indisunique AND i.indisvalid AND i.indisready
 AND i.indpred IS NOT NULL AND pg_get_expr(i.indpred,i.indrelid) LIKE '%PENDING%')`).Scan(&ok))
	} else {
		authdb.Require(t, "schema legacy trước MSK", db.Conn.QueryRow(context.Background(), `SELECT
 to_regclass('public.employee_skill_change_requests') IS NULL
 AND NOT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema='public'
 AND table_name='employee_skills' AND column_name='approved_by')
 AND (SELECT count(*)=2 FROM information_schema.columns WHERE table_schema='public'
 AND table_name='employee_skills' AND column_name IN ('source','approved_by_employee_id'))
 AND (SELECT is_nullable='YES' FROM information_schema.columns WHERE table_schema='public'
 AND table_name='employee_skills' AND column_name='approved_at')
 AND EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name='skills' AND column_name='status')
 AND to_regclass('public.idx_skills_status') IS NOT NULL
 AND EXISTS(SELECT 1 FROM pg_constraint c JOIN pg_attribute a ON a.attrelid=c.conrelid AND a.attnum=c.conkey[1]
 WHERE c.conrelid='public.employee_skills'::regclass AND c.contype='f' AND a.attname='employee_id'
 AND c.confdeltype='c' AND c.confupdtype='a')`).Scan(&ok))
	}
	if !ok {
		t.Fatalf("schema MSK không đúng kỳ vọng present=%t", present)
	}
}
