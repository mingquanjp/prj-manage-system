BEGIN;
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';
LOCK TABLE public.users, public.employees, public.skill_categories, public.skills,
           public.employee_skills, public.audit_events, public.employee_skill_change_requests,
           public.idempotency_records, public.notifications IN ACCESS EXCLUSIVE MODE;
DO $$
BEGIN
  -- Nguồn và trạng thái legacy không thể suy lại sau import hoặc khi đã dùng schema mới.
  IF EXISTS(SELECT 1 FROM public.skills)
    OR EXISTS(SELECT 1 FROM public.employee_skills)
    OR EXISTS(SELECT 1 FROM public.employee_skill_change_requests)
    OR EXISTS(SELECT 1 FROM public.idempotency_records WHERE operation='SKILL_CHANGE_SUBMIT')
    OR EXISTS(SELECT 1 FROM public.notifications WHERE resource_type='SKILL_CHANGE_REQUEST'
      OR notification_type IN ('SKILL_CHANGE_SUBMITTED','SKILL_CHANGE_APPROVED','SKILL_CHANGE_REJECTED'))
    OR EXISTS(SELECT 1 FROM public.audit_events WHERE operation LIKE 'SKILL\_%' ESCAPE '\'
      OR entity_type IN ('SKILL','SKILL_CATEGORY','EMPLOYEE_SKILL')) THEN
    RAISE EXCEPTION 'MSK: từ chối down khi còn dữ liệu hoặc lịch sử kỹ năng';
  END IF;
  IF EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema='public'
    AND table_name='employee_skill_change_requests' AND column_name NOT IN
      ('request_id','employee_id','skill_id','operation','source','before_values','proposed_values',
       'explanation','external_evidence','expert_opinion','requested_by','requested_at','status',
       'resolved_by','resolved_at','decision_reason','version'))
    OR EXISTS(SELECT 1 FROM pg_index i JOIN pg_class c ON c.oid=i.indexrelid
      WHERE i.indrelid='public.employee_skill_change_requests'::regclass AND c.relname NOT IN
        ('employee_skill_change_requests_pkey','uq_msk_pending','idx_msk_owner','idx_msk_review','idx_msk_employee_skill'))
    OR EXISTS(SELECT 1 FROM pg_constraint WHERE conrelid='public.employee_skill_change_requests'::regclass
      AND contype<>'n' AND conname NOT IN (
        'employee_skill_change_requests_pkey','employee_skill_change_requests_employee_id_fkey',
        'employee_skill_change_requests_skill_id_fkey','employee_skill_change_requests_requested_by_fkey',
        'employee_skill_change_requests_resolved_by_fkey','employee_skill_change_requests_operation_check',
        'employee_skill_change_requests_source_check','employee_skill_change_requests_explanation_check',
        'employee_skill_change_requests_external_evidence_check','employee_skill_change_requests_expert_opinion_check',
        'employee_skill_change_requests_status_check','employee_skill_change_requests_version_check',
        'ck_msk_values','ck_msk_decision'))
    OR EXISTS(SELECT 1 FROM pg_proc p JOIN pg_namespace n ON n.oid=p.pronamespace
      WHERE n.nspname NOT IN ('pg_catalog','information_schema')
        AND p.prosrc ~* '\memployee_skill_change_requests\M')
    OR EXISTS(SELECT 1 FROM pg_trigger WHERE tgrelid='public.employee_skill_change_requests'::regclass AND NOT tgisinternal)
    OR EXISTS(SELECT 1 FROM pg_policy WHERE polrelid='public.employee_skill_change_requests'::regclass)
    OR EXISTS(SELECT 1 FROM pg_constraint WHERE contype='f' AND confrelid='public.employee_skill_change_requests'::regclass)
    OR EXISTS(SELECT 1 FROM pg_rewrite r JOIN pg_depend d ON d.classid='pg_rewrite'::regclass AND d.objid=r.oid
      WHERE d.refclassid='pg_class'::regclass AND d.refobjid='public.employee_skill_change_requests'::regclass)
    OR EXISTS(SELECT 1 FROM pg_index i JOIN pg_class c ON c.oid=i.indexrelid
      WHERE i.indrelid='public.employee_skills'::regclass AND i.indkey::smallint[] @>
        ARRAY[(SELECT attnum FROM pg_attribute WHERE attrelid='public.employee_skills'::regclass AND attname='approved_by')]::smallint[])
    OR EXISTS(SELECT 1 FROM pg_constraint c WHERE c.conrelid='public.employee_skills'::regclass
      AND c.conkey @> ARRAY[(SELECT attnum FROM pg_attribute WHERE attrelid='public.employee_skills'::regclass AND attname='approved_by')]::smallint[]
      AND c.conname<>'fk_msk_approval_user' AND c.contype<>'n') THEN
    RAISE EXCEPTION 'MSK: dependency đích cần review trước down';
  END IF;
END $$;
DROP TABLE public.employee_skill_change_requests;
DROP FUNCTION public.msk_request_values_valid(text,jsonb,jsonb);
DROP FUNCTION public.msk_values_valid(jsonb);
DROP INDEX public.uq_msk_category_name, public.uq_msk_skill_name;
ALTER TABLE public.skill_categories DROP CONSTRAINT ck_msk_category_name;
ALTER TABLE public.skills DROP CONSTRAINT ck_msk_skill_name;
DROP FUNCTION public.msk_text_valid(text,integer);
ALTER TABLE public.skills ADD COLUMN status public.skill_status NOT NULL DEFAULT 'ACTIVE';
CREATE INDEX idx_skills_status ON public.skills(status);
ALTER TABLE public.employee_skills
  DROP COLUMN approved_by,
  ALTER COLUMN approved_at DROP NOT NULL,
  ADD COLUMN source public.skill_source NOT NULL DEFAULT 'MANUAL',
  ADD COLUMN approved_by_employee_id integer REFERENCES public.employees(employee_id),
  DROP CONSTRAINT employee_skills_employee_id_fkey,
  ADD CONSTRAINT employee_skills_employee_id_fkey FOREIGN KEY(employee_id) REFERENCES public.employees(employee_id) ON DELETE CASCADE;
COMMIT;
