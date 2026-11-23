package migrations

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/golang-migrate/migrate/v4/database"
)

const PCAVersion = 6

type pcaDriver struct {
	database.Driver
	db *sql.DB
}

// WithPCAPreflight kiểm nguồn dưới advisory lock trước dirty; consumer phải dừng
// trong cửa sổ migration. DDL vẫn kiểm lại dữ liệu dưới khóa bảng.
func WithPCAPreflight(driver database.Driver, db *sql.DB) database.Driver {
	return &pcaDriver{Driver: driver, db: db}
}

func (d *pcaDriver) SetVersion(version int, dirty bool) error {
	if version == PCAVersion && dirty {
		previous, wasDirty, err := d.Driver.Version()
		if err != nil {
			return errors.New("PCA: không đọc được version nguồn")
		}
		if previous == PCAVersion-1 && !wasDirty {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := CheckPCA(ctx, d.db); err != nil {
				return err
			}
		}
	}
	return d.Driver.SetVersion(version, dirty)
}

// CheckPCA chỉ đọc metadata/dữ liệu guard, không xuất nội dung notification.
func CheckPCA(ctx context.Context, db *sql.DB) error {
	var ok bool
	err := db.QueryRowContext(ctx, `SELECT current_schema()='public'
 AND (SELECT count(*)=1 FROM public.schema_migrations WHERE version=5 AND NOT dirty)
 AND to_regclass('public.employee_profile_change_requests') IS NULL
 AND to_regclass('public.idempotency_records') IS NULL
 AND to_regtype('public.idempotency_records') IS NULL
 AND to_regtype('public.employee_profile_change_requests') IS NULL
 AND NOT EXISTS(SELECT 1 FROM (VALUES
 ('idx_pca_owner'),('idx_pca_queue'),('idx_notifications_recipient'),
 ('employee_profile_change_requests_request_id_seq'),('employee_profile_change_requests_pkey'),
 ('idempotency_records_idempotency_id_seq'),('idempotency_records_pkey'),
 ('uq_idempotency_actor_operation_key')) AS expected(name)
 WHERE to_regclass('public.'||expected.name) IS NOT NULL)
 AND to_regprocedure('public.pca_profile_values_valid(jsonb,boolean)') IS NULL
 AND to_regprocedure('public.pca_same_profile_keys(jsonb,jsonb)') IS NULL
 AND to_regclass('public.employee_code_seq') IS NOT NULL
 AND to_regclass('public.uq_employees_normalized_email') IS NOT NULL
 AND to_regclass('public.audit_events') IS NOT NULL
 AND NOT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name='users' AND column_name='username')
 AND (SELECT count(*)=8 FROM information_schema.columns WHERE table_schema='public' AND table_name='notifications')
 AND (SELECT count(*)=8 FROM (VALUES
 ('notification_id','integer','NO',NULL::integer),('employee_id','integer','NO',NULL),
 ('type','character varying','NO',50),('title','character varying','NO',200),
 ('content','text','YES',NULL),('ref_task_id','integer','YES',NULL),
 ('is_read','boolean','NO',NULL),('created_at','timestamp with time zone','NO',NULL)
 ) e(col,typ,nullable,len) JOIN information_schema.columns c ON c.table_schema='public'
 AND c.table_name='notifications' AND c.column_name=e.col AND c.data_type=e.typ
 AND c.is_nullable=e.nullable AND c.character_maximum_length IS NOT DISTINCT FROM e.len)
 AND (SELECT column_default='false' FROM information_schema.columns WHERE table_schema='public' AND table_name='notifications' AND column_name='is_read')
 AND (SELECT column_default='now()' FROM information_schema.columns WHERE table_schema='public' AND table_name='notifications' AND column_name='created_at')
 AND NOT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name='notifications' AND column_name NOT IN ('is_read','created_at') AND column_default IS NOT NULL)
 AND (SELECT is_identity='YES' AND identity_generation='ALWAYS' FROM information_schema.columns WHERE table_schema='public' AND table_name='notifications' AND column_name='notification_id')`).Scan(&ok)
	if err != nil || !ok {
		return errors.New("PCA: schema nguồn lệch mốc M2")
	}
	err = db.QueryRowContext(ctx, `SELECT
 NOT EXISTS(SELECT 1 FROM public.notifications)
 AND (SELECT count(*)=2 FROM pg_constraint c JOIN pg_attribute a ON a.attrelid=c.conrelid AND a.attnum=c.conkey[1]
 WHERE c.conrelid='public.notifications'::regclass AND c.contype='f' AND cardinality(c.conkey)=1
 AND c.convalidated AND NOT c.condeferrable AND c.confdeltype='c' AND c.confupdtype='a'
 AND ((a.attname='employee_id' AND c.confrelid='public.employees'::regclass AND c.confkey=ARRAY[(SELECT attnum FROM pg_attribute WHERE attrelid='public.employees'::regclass AND attname='employee_id')]::smallint[])
 OR (a.attname='ref_task_id' AND c.confrelid='public.tasks'::regclass AND c.confkey=ARRAY[(SELECT attnum FROM pg_attribute WHERE attrelid='public.tasks'::regclass AND attname='task_id')]::smallint[])))
 AND (SELECT count(*)=1 FROM pg_constraint WHERE conrelid='public.notifications'::regclass AND contype='p' AND conkey=ARRAY[(SELECT attnum FROM pg_attribute WHERE attrelid='public.notifications'::regclass AND attname='notification_id')]::smallint[])
 AND NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conrelid='public.notifications'::regclass AND contype NOT IN ('p','f','n'))
 AND (SELECT count(*)=2 FROM pg_constraint WHERE conrelid='public.notifications'::regclass AND contype='f')
 AND (SELECT count(*)=2 FROM pg_index WHERE indrelid='public.notifications'::regclass)
 AND NOT EXISTS(SELECT 1 FROM pg_index WHERE indrelid='public.notifications'::regclass AND (NOT indisvalid OR NOT indisready OR indpred IS NOT NULL OR indexprs IS NOT NULL))
 AND NOT EXISTS(SELECT 1 FROM pg_inherits WHERE inhrelid='public.notifications'::regclass OR inhparent='public.notifications'::regclass)
 AND EXISTS(SELECT 1 FROM pg_indexes WHERE schemaname='public' AND tablename='notifications' AND indexname='idx_notifications_employee'
 AND indexdef='CREATE INDEX idx_notifications_employee ON public.notifications USING btree (employee_id, is_read, created_at DESC)')
 AND NOT EXISTS(SELECT 1 FROM pg_trigger WHERE tgrelid='public.notifications'::regclass AND NOT tgisinternal)
 AND NOT EXISTS(SELECT 1 FROM pg_policy WHERE polrelid='public.notifications'::regclass)
 AND NOT EXISTS(SELECT 1 FROM pg_class WHERE oid='public.notifications'::regclass AND (relrowsecurity OR relforcerowsecurity OR relkind<>'r'))
 AND NOT EXISTS(SELECT 1 FROM pg_constraint WHERE contype='f' AND confrelid='public.notifications'::regclass)
 AND NOT EXISTS(SELECT 1 FROM pg_rewrite r JOIN pg_depend d ON d.classid='pg_rewrite'::regclass AND d.objid=r.oid WHERE d.refobjid='public.notifications'::regclass AND d.refclassid='pg_class'::regclass)
 AND NOT EXISTS(SELECT 1 FROM pg_proc p JOIN pg_namespace n ON n.oid=p.pronamespace WHERE n.nspname NOT IN ('pg_catalog','information_schema') AND p.prosrc ~* '\mnotifications\M')
 AND NOT EXISTS(SELECT 1 FROM pg_depend d WHERE d.refclassid='pg_class'::regclass AND d.refobjid='public.notifications'::regclass AND d.deptype='n'
 AND d.classid NOT IN ('pg_constraint'::regclass,'pg_class'::regclass,'pg_attrdef'::regclass))`).Scan(&ok)
	if err != nil || !ok {
		return errors.New("PCA: notification có dữ liệu hoặc phụ thuộc cần review")
	}
	return nil
}
