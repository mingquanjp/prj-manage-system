BEGIN;
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';
LOCK TABLE public.employee_profile_change_requests,
           public.idempotency_records,public.notifications IN ACCESS EXCLUSIVE MODE;
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM public.employee_profile_change_requests)
     OR EXISTS (SELECT 1 FROM public.idempotency_records)
     OR EXISTS (SELECT 1 FROM public.notifications) THEN
    RAISE EXCEPTION 'PCA: từ chối down khi còn lịch sử/key/notification';
  END IF;
  -- DROP COLUMN/TABLE tự loại index/constraint nội bộ: từ chối đối tượng lạ
  -- trước DDL để không vô tình xóa dependency do consumer bổ sung.
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema='public' AND (
      (table_name='employee_profile_change_requests' AND column_name NOT IN (
        'request_id','employee_id','before_values','proposed_values','explanation',
        'requested_by','requested_at','status','resolved_by','resolved_at','decision_reason'))
      OR (table_name='idempotency_records' AND column_name NOT IN (
        'idempotency_id','actor_id','operation','idempotency_key','request_hash',
        'result_reference','created_at','expires_at'))
      OR (table_name='notifications' AND column_name NOT IN (
        'notification_id','created_at','recipient_id','notification_type','resource_type',
        'resource_id','message','read_at')))
  ) OR EXISTS (
    SELECT 1 FROM pg_index i JOIN pg_class c ON c.oid=i.indexrelid
    WHERE i.indrelid IN ('public.employee_profile_change_requests'::regclass,
      'public.idempotency_records'::regclass,'public.notifications'::regclass)
    AND NOT (
      (i.indrelid='public.employee_profile_change_requests'::regclass AND c.relname IN (
        'employee_profile_change_requests_pkey','idx_pca_owner','idx_pca_queue'))
      OR (i.indrelid='public.idempotency_records'::regclass AND c.relname IN (
        'idempotency_records_pkey','uq_idempotency_actor_operation_key'))
      OR (i.indrelid='public.notifications'::regclass AND c.relname IN (
        'notifications_pkey','idx_notifications_recipient')))
  ) OR EXISTS (
    SELECT 1 FROM pg_constraint c
    WHERE c.conrelid IN ('public.employee_profile_change_requests'::regclass,
      'public.idempotency_records'::regclass,'public.notifications'::regclass)
    AND c.contype <> 'n' AND NOT (
      (c.conrelid='public.employee_profile_change_requests'::regclass AND c.conname IN (
        'employee_profile_change_requests_pkey','employee_profile_change_requests_employee_id_fkey',
        'employee_profile_change_requests_requested_by_fkey','employee_profile_change_requests_resolved_by_fkey',
        'ck_pca_values','ck_pca_status','ck_pca_decision','ck_pca_times','ck_pca_explanation'))
      OR (c.conrelid='public.idempotency_records'::regclass AND c.conname IN (
        'idempotency_records_pkey','idempotency_records_actor_id_fkey',
        'idempotency_records_request_hash_check','idempotency_records_result_reference_check',
        'uq_idempotency_actor_operation_key','ck_idempotency_expiry'))
      OR (c.conrelid='public.notifications'::regclass AND c.conname IN (
        'notifications_pkey','notifications_recipient_id_fkey','notifications_resource_id_check',
        'notifications_message_check','ck_notifications_read_at')))
  ) OR EXISTS (
    SELECT 1 FROM pg_constraint WHERE contype='f'
    AND confrelid IN ('public.employee_profile_change_requests'::regclass,
      'public.idempotency_records'::regclass,'public.notifications'::regclass)
  ) OR EXISTS (
    SELECT 1 FROM pg_trigger WHERE NOT tgisinternal AND tgrelid IN (
      'public.employee_profile_change_requests'::regclass,
      'public.idempotency_records'::regclass,'public.notifications'::regclass)
  ) OR EXISTS (
    SELECT 1 FROM pg_policy WHERE polrelid IN (
      'public.employee_profile_change_requests'::regclass,
      'public.idempotency_records'::regclass,'public.notifications'::regclass)
  ) OR EXISTS (
    SELECT 1 FROM pg_proc p JOIN pg_namespace n ON n.oid=p.pronamespace
    WHERE n.nspname NOT IN ('pg_catalog','information_schema')
      AND p.prosrc ~* '\m(notifications|employee_profile_change_requests|idempotency_records)\M'
  ) THEN
    RAISE EXCEPTION 'PCA: dependency đích cần review trước down';
  END IF;
END $$;
DROP INDEX public.idx_notifications_recipient;
ALTER TABLE public.notifications
  DROP COLUMN recipient_id,
  DROP COLUMN notification_type,
  DROP COLUMN resource_type,
  DROP COLUMN resource_id,
  DROP COLUMN message,
  DROP COLUMN read_at,
  ADD COLUMN employee_id int NOT NULL REFERENCES public.employees(employee_id) ON DELETE CASCADE,
  ADD COLUMN type varchar(50) NOT NULL,
  ADD COLUMN title varchar(200) NOT NULL,
  ADD COLUMN content text,
  ADD COLUMN ref_task_id int REFERENCES public.tasks(task_id) ON DELETE CASCADE,
  ADD COLUMN is_read boolean NOT NULL DEFAULT false;
CREATE INDEX idx_notifications_employee ON public.notifications
  (employee_id,is_read,created_at DESC);
DROP TABLE public.employee_profile_change_requests;
DROP TABLE public.idempotency_records;
DROP FUNCTION public.pca_same_profile_keys(jsonb,jsonb);
DROP FUNCTION public.pca_profile_values_valid(jsonb,boolean);
COMMIT;
