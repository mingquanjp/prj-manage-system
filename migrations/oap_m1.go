package migrations

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/golang-migrate/migrate/v4/database"
)

const M1Version = 4

type m1Driver struct {
	database.Driver
	db       *sql.DB
	approved map[string]string
}

// WithM1Preflight giữ preflight trong advisory lock của migrate và chạy
// trước khi driver ghi version dirty. Caller vẫn cần dừng đường ghi ứng dụng.
func WithM1Preflight(driver database.Driver, db *sql.DB, approved map[string]string) database.Driver {
	return &m1Driver{Driver: driver, db: db, approved: approved}
}

func (d *m1Driver) SetVersion(version int, dirty bool) error {
	if version == M1Version && dirty {
		previous, _, err := d.Driver.Version()
		if err != nil {
			return errors.New("M1: không đọc được version trước chuyển đổi")
		}
		if previous == M1Version-1 {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := CheckM1(ctx, d.db, d.approved); err != nil {
				return err
			}
		}
	}
	return d.Driver.SetVersion(version, dirty)
}

// CheckM1 kiểm schema và dữ liệu ở mốc 000003 trước khi runner đổi version.
// Caller giữ đường ghi dừng trong suốt nâng cấp. Mapping dùng giá trị legacy gốc.
func CheckM1(ctx context.Context, db *sql.DB, approved map[string]string) error {
	for legacy, name := range approved {
		if strings.TrimSpace(legacy) == "" || name == "" || strings.TrimSpace(name) != name ||
			!utf8.ValidString(name) || utf8.RuneCountInString(name) > 100 {
			return errors.New("M1: mapping cấp bậc không hợp lệ")
		}
	}
	if approved == nil {
		approved = map[string]string{}
	}
	payload, err := json.Marshal(approved)
	if err != nil {
		return errors.New("M1: không mã hóa được mapping")
	}

	var schemaOK bool
	err = db.QueryRowContext(ctx, `SELECT
		current_schema() = 'public'
		AND (SELECT count(*) = 1 FROM public.schema_migrations WHERE version = 3 AND NOT dirty)
		AND (SELECT count(*) = 3
		     FROM (VALUES ('user_id','public.users'),('position_id','public.positions'),
		                  ('department_id','public.departments')) AS expected(col, target)
		     WHERE EXISTS (
		       SELECT 1 FROM pg_constraint c
		       JOIN pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = c.conkey[1]
		       JOIN pg_attribute r ON r.attrelid = c.confrelid AND r.attnum = c.confkey[1]
		       WHERE c.conrelid = 'public.employees'::regclass AND c.contype = 'f'
		         AND c.confrelid = to_regclass(expected.target)
		         AND a.attname = expected.col AND r.attname = expected.col
		         AND cardinality(c.conkey) = 1 AND c.confdeltype = 'a' AND c.confupdtype = 'a'
		         AND c.convalidated AND NOT c.condeferrable
		         AND a.attnotnull = (expected.col <> 'department_id')))
		AND EXISTS (
		     SELECT 1 FROM pg_trigger
		     WHERE tgrelid = 'public.employees'::regclass AND tgname = 'trg_employees_updated_at'
		       AND tgfoid = to_regprocedure('public.set_updated_at()') AND tgenabled = 'O')
		AND to_regclass('public.job_levels') IS NULL
		AND to_regclass('public.audit_events') IS NULL
		AND to_regclass('public.uq_departments_normalized_name') IS NULL
		AND to_regclass('public.uq_positions_normalized_name') IS NULL
		AND to_regclass('public.idx_employees_job_level') IS NULL
		AND (SELECT count(*) = 7 FROM information_schema.columns
		     WHERE table_schema = 'public' AND (
		       (table_name = 'departments' AND column_name = 'department_name' AND data_type = 'character varying' AND character_maximum_length = 100 AND is_nullable = 'NO')
		       OR (table_name = 'positions' AND column_name = 'position_name' AND data_type = 'character varying' AND character_maximum_length = 100 AND is_nullable = 'NO')
		       OR (table_name = 'employees' AND column_name = 'seniority_level' AND data_type = 'character varying' AND character_maximum_length = 30 AND is_nullable = 'YES')
		       OR (table_name = 'employees' AND column_name IN ('employee_id','user_id','position_id','department_id') AND data_type = 'integer')))
		AND NOT EXISTS (SELECT 1 FROM information_schema.columns
		     WHERE table_schema = 'public' AND table_name = 'employees' AND column_name IN ('job_level_id','address'))
		AND NOT EXISTS (SELECT 1 FROM pg_constraint
		     WHERE conrelid IN ('public.departments'::regclass, 'public.positions'::regclass)
		     AND conname IN ('ck_departments_name','ck_positions_name'))
		AND to_regclass('public.user_roles') IS NOT NULL
		AND to_regclass('public.auth_sessions') IS NOT NULL`).Scan(&schemaOK)
	if err != nil || !schemaOK {
		return errors.New("M1: schema nguồn lệch mốc 000003 hoặc không đọc được")
	}

	// SQL dùng cùng khóa chuẩn hóa với các index M1.
	var invalidNames, missingMapping, duplicateLevels bool
	err = db.QueryRowContext(ctx, `WITH names(kind, name) AS (
		SELECT 'department', department_name FROM public.departments
		UNION ALL SELECT 'position', position_name FROM public.positions
	), legacy(value) AS (
		SELECT DISTINCT seniority_level FROM public.employees WHERE seniority_level ~ U&'[^\0009\000A\000B\000C\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]'
	), mapped(name) AS (
		SELECT DISTINCT $1::jsonb ->> value FROM legacy
	)
	SELECT EXISTS (
		SELECT 1 FROM names GROUP BY kind, lower(btrim(name))
		HAVING count(*) > 1 OR bool_or(name !~ U&'[^\0009\000A\000B\000C\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]')
		OR bool_or(btrim(name) ~ U&'^[\0009\000A\000B\000C\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]|[\0009\000A\000B\000C\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]$')
	), EXISTS (SELECT 1 FROM legacy WHERE NOT ($1::jsonb ? value)),
	EXISTS (SELECT 1 FROM mapped GROUP BY lower(btrim(name)) HAVING count(*) > 1)`, string(payload)).
		Scan(&invalidNames, &missingMapping, &duplicateLevels)
	switch {
	case err != nil:
		return errors.New("M1: không đọc được dữ liệu preflight")
	case invalidNames:
		return errors.New("M1: tên danh mục trắng, padding không tương thích hoặc trùng sau chuẩn hóa")
	case missingMapping:
		return errors.New("M1: thiếu mapping cấp bậc đã xác minh")
	case duplicateLevels:
		return errors.New("M1: tên cấp bậc trong mapping trùng sau chuẩn hóa")
	default:
		return nil
	}
}
