package migrations

import (
	"context"
	"database/sql"
	"errors"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"ai-staffing/internal/service/auth"
	"github.com/golang-migrate/migrate/v4/database"
)

const M2Version = 5

type m2Driver struct {
	database.Driver
	db *sql.DB
}

// WithM2Preflight kiểm dữ liệu dưới advisory lock của runner, trước khi ghi dirty.
// Người vận hành phải dừng consumer ghi trong toàn cửa sổ chuyển đổi.
func WithM2Preflight(driver database.Driver, db *sql.DB) database.Driver {
	return &m2Driver{Driver: driver, db: db}
}

func (d *m2Driver) SetVersion(version int, dirty bool) error {
	if version == M2Version && dirty {
		previous, _, err := d.Driver.Version()
		if err != nil {
			return errors.New("M2: không đọc được version nguồn")
		}
		if previous == M2Version-1 {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := CheckM2(ctx, d.db); err != nil {
				return err
			}
		}
	}
	return d.Driver.SetVersion(version, dirty)
}

// CheckM2 không sửa dữ liệu; lỗi chỉ chứa phân loại, không chứa email/username.
func CheckM2(ctx context.Context, db *sql.DB) error {
	var ok bool
	err := db.QueryRowContext(ctx, `SELECT current_schema()='public'
 AND (SELECT count(*)=1 FROM public.schema_migrations WHERE version=4 AND NOT dirty)
 AND to_regclass('public.employee_code_seq') IS NULL
 AND to_regclass('public.uq_employees_normalized_email') IS NULL
 AND to_regclass('public.job_levels') IS NOT NULL AND to_regclass('public.audit_events') IS NOT NULL
 AND to_regclass('public.auth_sessions') IS NOT NULL AND to_regclass('public.user_roles') IS NOT NULL
 AND (SELECT count(*)=4 FROM information_schema.columns WHERE table_schema='public' AND (
  (table_name='users' AND column_name='username' AND data_type='character varying' AND character_maximum_length=50 AND is_nullable='NO') OR
  (table_name='employees' AND column_name='email' AND data_type='character varying' AND character_maximum_length=100 AND is_nullable='NO') OR
  (table_name='employees' AND column_name='employee_code' AND data_type='character varying' AND character_maximum_length=30 AND is_nullable='NO') OR
  (table_name='users' AND column_name='system_role' AND udt_name='user_system_role' AND is_nullable='NO')))
 AND NOT EXISTS(SELECT 1 FROM pg_attrdef WHERE adrelid='public.users'::regclass AND adnum=(SELECT attnum FROM pg_attribute WHERE attrelid=adrelid AND attname='username'))
 AND EXISTS(SELECT 1 FROM pg_enum WHERE enumtypid='public.user_system_role'::regtype AND enumlabel='EMPLOYEE')
 AND (SELECT count(*)=3 FROM pg_constraint c WHERE c.contype='u' AND c.convalidated AND NOT c.condeferrable AND cardinality(c.conkey)=1 AND (
  (c.conrelid='public.users'::regclass AND c.conkey[1]=(SELECT attnum FROM pg_attribute WHERE attrelid=c.conrelid AND attname='username')) OR
  (c.conrelid='public.employees'::regclass AND c.conkey[1] IN (SELECT attnum FROM pg_attribute WHERE attrelid=c.conrelid AND attname IN ('email','employee_code')))))`).Scan(&ok)
	if err != nil || !ok {
		return errors.New("M2: schema nguồn lệch mốc M1")
	}

	err = db.QueryRowContext(ctx, `SELECT
 (SELECT count(*)=2 FROM pg_constraint c WHERE c.contype='p' AND c.convalidated
   AND cardinality(c.conkey)=1 AND c.conrelid IN ('public.users'::regclass,'public.employees'::regclass)
   AND c.conkey[1]=(SELECT attnum FROM pg_attribute WHERE attrelid=c.conrelid
      AND attname=CASE c.conrelid WHEN 'public.users'::regclass THEN 'user_id' ELSE 'employee_id' END))
 AND (SELECT count(*)=4 FROM (VALUES ('employees'),('user_roles'),('auth_sessions'),('audit_events')) AS source(name)
 WHERE EXISTS(SELECT 1 FROM pg_constraint c JOIN pg_attribute a ON a.attrelid=c.conrelid AND a.attnum=c.conkey[1]
 JOIN pg_attribute r ON r.attrelid=c.confrelid AND r.attnum=c.confkey[1]
 WHERE c.conrelid=to_regclass('public.'||source.name) AND c.confrelid='public.users'::regclass
 AND c.contype='f' AND cardinality(c.conkey)=1 AND a.attname=CASE source.name WHEN 'audit_events' THEN 'actor_id' ELSE 'user_id' END
 AND r.attname='user_id' AND c.convalidated AND NOT c.condeferrable AND c.confdeltype='a' AND c.confupdtype='a'))
 AND NOT EXISTS(SELECT 1 FROM pg_trigger WHERE tgrelid='public.users'::regclass AND NOT tgisinternal
 AND NOT (tgname='trg_users_updated_at' AND tgfoid=to_regprocedure('public.set_updated_at()') AND tgenabled='O'))`).Scan(&ok)
	if err != nil || !ok {
		return errors.New("M2: khóa hoặc trigger schema nguồn cần review")
	}

	// Chỉ constraint/index đơn của username được tự loại cùng cột. Phụ thuộc khác
	// cần kế hoạch tường minh; rà function body thêm vì PL/pgSQL không luôn có pg_depend.
	err = db.QueryRowContext(ctx, `SELECT NOT EXISTS (
 SELECT 1 FROM pg_depend d
 WHERE d.refclassid='pg_class'::regclass AND d.refobjid='public.users'::regclass
 AND d.refobjsubid=(SELECT attnum FROM pg_attribute WHERE attrelid='public.users'::regclass AND attname='username')
 AND NOT (d.deptype='a' AND (
  (d.classid='pg_constraint'::regclass AND EXISTS(SELECT 1 FROM pg_constraint c WHERE c.oid=d.objid AND c.conrelid='public.users'::regclass AND c.contype IN ('u','n') AND cardinality(c.conkey)=1)) OR
  (d.classid='pg_class'::regclass AND EXISTS(SELECT 1 FROM pg_index i WHERE i.indexrelid=d.objid AND i.indrelid='public.users'::regclass AND i.indnatts=1 AND i.indnkeyatts=1 AND i.indexprs IS NULL AND i.indpred IS NULL)))))
 AND NOT EXISTS (SELECT 1 FROM pg_proc p JOIN pg_namespace n ON n.oid=p.pronamespace
 WHERE n.nspname NOT IN ('pg_catalog','information_schema') AND p.prosrc ~* '\musername\M')`).Scan(&ok)
	if err != nil || !ok {
		return errors.New("M2: phụ thuộc username cần review tường minh")
	}

	rows, err := db.QueryContext(ctx, `SELECT email,lower(btrim(email)) FROM public.employees`)
	if err != nil {
		return errors.New("M2: không đọc được email preflight")
	}
	seen := map[string]bool{}
	for rows.Next() {
		var raw, pg string
		if err := rows.Scan(&raw, &pg); err != nil {
			rows.Close()
			return errors.New("M2: không đọc được email")
		}
		normalized := auth.NormalizeLoginEmail(raw)
		parsed, parseErr := mail.ParseAddress(normalized)
		if !utf8.ValidString(normalized) || utf8.RuneCountInString(normalized) > 100 || strings.ContainsAny(raw, "\r\n\x00") || parseErr != nil || parsed.Name != "" || parsed.Address != normalized {
			rows.Close()
			return errors.New("M2: email legacy không hợp lệ")
		}
		if normalized != pg {
			rows.Close()
			return errors.New("M2: chuẩn hóa email Go và PostgreSQL không tương đương")
		}
		if seen[normalized] {
			rows.Close()
			return errors.New("M2: email trùng sau chuẩn hóa")
		}
		seen[normalized] = true
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return errors.New("M2: không đọc hết email preflight")
	}
	err = db.QueryRowContext(ctx, `SELECT NOT EXISTS(SELECT 1 FROM public.employees
 WHERE length(employee_code) NOT BETWEEN 1 AND 30 OR btrim(employee_code)='' OR
 (CASE WHEN employee_code ~ '^EMP[0-9]+$' THEN substring(employee_code FROM 4)::numeric >= 9223372036854775807 ELSE false END))`).Scan(&ok)
	if err != nil || !ok {
		return errors.New("M2: mã legacy không hợp lệ hoặc hết dải BIGINT")
	}
	return nil
}
