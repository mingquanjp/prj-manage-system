BEGIN;
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';
LOCK TABLE public.users, public.employees, public.skill_categories, public.skills,
           public.employee_skills, public.audit_events IN ACCESS EXCLUSIVE MODE;

-- MSK-GUARD-BEGIN
DO $$
DECLARE
  valid boolean;
  invalid_names boolean;
  invalid_approval boolean;
  missing_actor boolean;
  import_actor integer;
BEGIN
  import_actor := COALESCE(NULLIF(current_setting('app.msk_import_actor_id', true), ''), '0')::integer;

  SELECT current_schema()='public'

 AND to_regclass('public.employee_skill_change_requests') IS NULL
 AND to_regtype('public.employee_skill_change_requests') IS NULL
 AND NOT EXISTS(SELECT 1 FROM (VALUES
 ('uq_msk_category_name'),('uq_msk_skill_name'),('uq_msk_pending'),('idx_msk_owner'),
 ('idx_msk_review'),('idx_msk_employee_skill'),('employee_skill_change_requests_request_id_seq'),
 ('employee_skill_change_requests_pkey')) e(name) WHERE to_regclass('public.'||name) IS NOT NULL)
 AND to_regclass('public.audit_events') IS NOT NULL
 AND to_regclass('public.idempotency_records') IS NOT NULL
 AND to_regclass('public.notifications') IS NOT NULL
 AND to_regclass('public.employee_profile_change_requests') IS NOT NULL
 AND (SELECT count(*)=21 FROM information_schema.columns WHERE table_schema='public'
 AND table_name IN ('skill_categories','skills','employee_skills'))
 AND (SELECT count(*)=21 FROM (VALUES
 ('skill_categories','category_id','integer','NO',NULL::integer,NULL::integer,NULL::integer,'int4'),
 ('skill_categories','category_name','character varying','NO',100,NULL,NULL,'varchar'),
 ('skill_categories','created_at','timestamp with time zone','NO',NULL,NULL,NULL,'timestamptz'),
 ('skill_categories','updated_at','timestamp with time zone','NO',NULL,NULL,NULL,'timestamptz'),
 ('skills','skill_id','integer','NO',NULL,NULL,NULL,'int4'),
 ('skills','skill_name','character varying','NO',100,NULL,NULL,'varchar'),
 ('skills','category_id','integer','NO',NULL,NULL,NULL,'int4'),
 ('skills','status','USER-DEFINED','NO',NULL,NULL,NULL,'skill_status'),
 ('skills','created_at','timestamp with time zone','NO',NULL,NULL,NULL,'timestamptz'),
 ('skills','updated_at','timestamp with time zone','NO',NULL,NULL,NULL,'timestamptz'),
 ('employee_skills','employee_skill_id','integer','NO',NULL,NULL,NULL,'int4'),
 ('employee_skills','employee_id','integer','NO',NULL,NULL,NULL,'int4'),
 ('employee_skills','skill_id','integer','NO',NULL,NULL,NULL,'int4'),
 ('employee_skills','proficiency_level','USER-DEFINED','NO',NULL,NULL,NULL,'proficiency_level'),
 ('employee_skills','years_experience','numeric','YES',NULL,4,1,'numeric'),
 ('employee_skills','note','text','YES',NULL,NULL,NULL,'text'),
 ('employee_skills','source','USER-DEFINED','NO',NULL,NULL,NULL,'skill_source'),
 ('employee_skills','approved_by_employee_id','integer','YES',NULL,NULL,NULL,'int4'),
 ('employee_skills','approved_at','timestamp with time zone','YES',NULL,NULL,NULL,'timestamptz'),
 ('employee_skills','created_at','timestamp with time zone','NO',NULL,NULL,NULL,'timestamptz'),
 ('employee_skills','updated_at','timestamp with time zone','NO',NULL,NULL,NULL,'timestamptz')
 ) e(tab,col,typ,nullable,len,prec,scale,udt) JOIN information_schema.columns c
 ON c.table_schema='public' AND c.table_name=e.tab AND c.column_name=e.col
 AND c.data_type=e.typ AND c.is_nullable=e.nullable AND c.udt_name=e.udt
 AND c.udt_schema=CASE WHEN e.typ='USER-DEFINED' THEN 'public' ELSE 'pg_catalog' END
 AND c.character_maximum_length IS NOT DISTINCT FROM e.len
 AND c.numeric_precision IS NOT DISTINCT FROM CASE WHEN e.typ='integer' THEN 32 ELSE e.prec END
 AND c.numeric_scale IS NOT DISTINCT FROM CASE WHEN e.typ='integer' THEN 0 ELSE e.scale END)
 AND (SELECT count(*)=3 FROM information_schema.columns WHERE table_schema='public'
 AND ((table_name='skill_categories' AND column_name='category_id')
 OR (table_name='skills' AND column_name='skill_id')
 OR (table_name='employee_skills' AND column_name='employee_skill_id'))
 AND is_identity='YES' AND identity_generation='ALWAYS')
 AND NOT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema='public'
 AND table_name IN ('skill_categories','skills','employee_skills') AND
 ((column_name IN ('created_at','updated_at') AND column_default IS DISTINCT FROM 'now()')
 OR (table_name='skills' AND column_name='status' AND column_default IS DISTINCT FROM '''ACTIVE''::skill_status')
 OR (table_name='employee_skills' AND column_name='source' AND column_default IS DISTINCT FROM '''MANUAL''::skill_source')
 OR (column_name NOT IN ('created_at','updated_at','source','status') AND column_default IS NOT NULL)))
 AND (SELECT count(*)=3 FROM information_schema.columns WHERE table_schema='public'
 AND table_name IN ('skill_categories','skills','employee_skills') AND is_identity='YES')
 AND NOT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema='public'
 AND table_name IN ('skill_categories','skills','employee_skills') AND is_generated<>'NEVER')
 AND (SELECT array_agg(enumlabel ORDER BY enumsortorder)=ARRAY['ACTIVE','INACTIVE']::name[] FROM pg_enum WHERE enumtypid=to_regtype('public.skill_status'))
 AND (SELECT array_agg(enumlabel ORDER BY enumsortorder)=ARRAY['MANUAL','CV']::name[] FROM pg_enum WHERE enumtypid=to_regtype('public.skill_source'))
 AND (SELECT array_agg(enumlabel ORDER BY enumsortorder)=ARRAY['BASIC','INTERMEDIATE','ADVANCED','EXPERT']::name[] FROM pg_enum WHERE enumtypid=to_regtype('public.proficiency_level'))
 AND to_regprocedure('public.msk_text_valid(text,integer)') IS NULL
 AND to_regprocedure('public.msk_values_valid(jsonb)') IS NULL
 AND to_regprocedure('public.msk_request_values_valid(text,jsonb,jsonb)') IS NULL INTO valid;
  IF valid IS NOT TRUE THEN RAISE EXCEPTION 'MSK: schema nguồn lệch mốc 000006'; END IF;

  SELECT
 NOT EXISTS(SELECT 1 FROM (VALUES
 ('skills','category_id','skill_categories','category_id','a'),
 ('employee_skills','employee_id','employees','employee_id','c'),
 ('employee_skills','skill_id','skills','skill_id','a'),
 ('employee_skills','approved_by_employee_id','employees','employee_id','a')
 ) e(tab,col,target,target_col,del) WHERE (SELECT count(*) FROM pg_constraint c
 JOIN pg_attribute a ON a.attrelid=c.conrelid AND a.attnum=c.conkey[1]
 JOIN pg_attribute r ON r.attrelid=c.confrelid AND r.attnum=c.confkey[1]
 WHERE c.conrelid=to_regclass('public.'||e.tab) AND c.contype='f'
 AND c.confrelid=to_regclass('public.'||e.target) AND a.attname=e.col AND r.attname=e.target_col
 AND cardinality(c.conkey)=1 AND c.confdeltype::text=e.del AND c.confupdtype='a'
 AND c.convalidated AND NOT c.condeferrable)<>1)
 AND (SELECT count(*)=4 FROM pg_constraint WHERE conrelid IN ('public.skill_categories'::regclass,'public.skills'::regclass,'public.employee_skills'::regclass) AND contype='f')
 AND (SELECT count(*)=3 FROM pg_constraint c JOIN pg_attribute a ON a.attrelid=c.conrelid AND a.attnum=c.conkey[1]
 WHERE c.contype='p' AND cardinality(c.conkey)=1 AND
 ((c.conrelid='public.skill_categories'::regclass AND a.attname='category_id')
 OR (c.conrelid='public.skills'::regclass AND a.attname='skill_id')
 OR (c.conrelid='public.employee_skills'::regclass AND a.attname='employee_skill_id')))
 AND NOT EXISTS(SELECT 1 FROM (VALUES
 ('skill_categories',ARRAY['category_name']),('skills',ARRAY['skill_name']),
 ('employee_skills',ARRAY['employee_id','skill_id'])) e(tab,cols)
 WHERE (SELECT count(*) FROM pg_constraint c WHERE c.contype='u'
 AND c.conrelid=to_regclass('public.'||e.tab) AND c.conkey=ARRAY(
 SELECT a.attnum FROM unnest(e.cols) WITH ORDINALITY x(col,position)
 JOIN pg_attribute a ON a.attrelid=c.conrelid AND a.attname=x.col ORDER BY x.position))<>1)
 AND (SELECT count(*)=1 FROM pg_constraint WHERE conrelid IN ('public.skill_categories'::regclass,'public.skills'::regclass,'public.employee_skills'::regclass) AND contype='c'
 AND conrelid='public.employee_skills'::regclass AND conname='employee_skills_years_experience_check'
 AND pg_get_constraintdef(oid)='CHECK (((years_experience IS NULL) OR (years_experience >= (0)::numeric)))' AND convalidated)
 AND NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conrelid IN ('public.skill_categories'::regclass,'public.skills'::regclass,'public.employee_skills'::regclass)
 AND (contype NOT IN ('p','u','f','n','c') OR (contype='c' AND conname<>'employee_skills_years_experience_check')))
 AND (SELECT count(*)=1 FROM pg_constraint WHERE conrelid IN
 ('public.skill_categories'::regclass,'public.skills'::regclass,'public.employee_skills'::regclass) AND contype='c')
 AND NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conrelid IN
 ('public.skill_categories'::regclass,'public.skills'::regclass,'public.employee_skills'::regclass)
 AND (NOT convalidated OR condeferrable))
 AND (SELECT count(*)=9 FROM pg_index WHERE indrelid IN ('public.skill_categories'::regclass,'public.skills'::regclass,'public.employee_skills'::regclass))
 AND NOT EXISTS(SELECT 1 FROM pg_index WHERE indrelid IN ('public.skill_categories'::regclass,'public.skills'::regclass,'public.employee_skills'::regclass)
 AND (NOT indisvalid OR NOT indisready OR indpred IS NOT NULL OR indexprs IS NOT NULL))
 AND (SELECT count(*)=3 FROM pg_indexes WHERE schemaname='public' AND
 ((tablename='skills' AND indexname='idx_skills_category' AND indexdef='CREATE INDEX idx_skills_category ON public.skills USING btree (category_id)')
 OR (tablename='skills' AND indexname='idx_skills_status' AND indexdef='CREATE INDEX idx_skills_status ON public.skills USING btree (status)')
 OR (tablename='employee_skills' AND indexname='idx_employee_skills_skill' AND indexdef='CREATE INDEX idx_employee_skills_skill ON public.employee_skills USING btree (skill_id)')))
 AND (SELECT count(*)=3 FROM pg_trigger WHERE tgrelid IN ('public.skill_categories'::regclass,'public.skills'::regclass,'public.employee_skills'::regclass)
 AND NOT tgisinternal AND tgname='trg_'||(tgrelid::regclass)::text||'_updated_at'
 AND tgfoid=to_regprocedure('public.set_updated_at()') AND tgenabled='O' AND tgtype=19 AND tgnargs=0 AND tgqual IS NULL)
 AND (SELECT count(*)=3 FROM pg_trigger WHERE tgrelid IN ('public.skill_categories'::regclass,'public.skills'::regclass,'public.employee_skills'::regclass) AND NOT tgisinternal)
 AND NOT EXISTS(SELECT 1 FROM pg_inherits WHERE inhrelid IN ('public.skill_categories'::regclass,'public.skills'::regclass,'public.employee_skills'::regclass)
 OR inhparent IN ('public.skill_categories'::regclass,'public.skills'::regclass,'public.employee_skills'::regclass))
 AND NOT EXISTS(SELECT 1 FROM pg_class WHERE oid IN ('public.skill_categories'::regclass,'public.skills'::regclass,'public.employee_skills'::regclass)
 AND (relkind<>'r' OR relrowsecurity OR relforcerowsecurity))
 AND NOT EXISTS(SELECT 1 FROM pg_policy WHERE polrelid IN ('public.skill_categories'::regclass,'public.skills'::regclass,'public.employee_skills'::regclass))
 AND NOT EXISTS(SELECT 1 FROM pg_constraint WHERE contype='f' AND confrelid IN ('public.skill_categories'::regclass,'public.skills'::regclass,'public.employee_skills'::regclass)
 AND conrelid NOT IN ('public.skills'::regclass,'public.employee_skills'::regclass,'public.task_requirements'::regclass,'public.cv_extracted_skills'::regclass))
 AND (SELECT count(*)=4 FROM pg_constraint WHERE contype='f' AND confrelid IN
 ('public.skill_categories'::regclass,'public.skills'::regclass,'public.employee_skills'::regclass))
 AND NOT EXISTS(SELECT 1 FROM pg_constraint c
 JOIN pg_attribute a ON a.attrelid=c.conrelid AND a.attnum=c.conkey[1]
 JOIN pg_attribute r ON r.attrelid=c.confrelid AND r.attnum=c.confkey[1]
 WHERE c.contype='f' AND c.conrelid IN ('public.task_requirements'::regclass,'public.cv_extracted_skills'::regclass)
 AND c.confrelid IN ('public.skill_categories'::regclass,'public.skills'::regclass,'public.employee_skills'::regclass)
 AND (c.confrelid<>'public.skills'::regclass OR a.attname<>'skill_id' OR r.attname<>'skill_id'
 OR cardinality(c.conkey)<>1 OR c.confdeltype<>'a' OR c.confupdtype<>'a'
 OR NOT c.convalidated OR c.condeferrable))
 AND NOT EXISTS(SELECT 1 FROM pg_rewrite r JOIN pg_depend d ON d.classid='pg_rewrite'::regclass AND d.objid=r.oid
 WHERE d.refclassid='pg_class'::regclass AND d.refobjid IN ('public.skill_categories'::regclass,'public.skills'::regclass,'public.employee_skills'::regclass))
 AND NOT EXISTS(SELECT 1 FROM pg_proc p JOIN pg_namespace n ON n.oid=p.pronamespace WHERE n.nspname NOT IN ('pg_catalog','information_schema')
 AND p.prosrc ~* '\m(skill_categories|skills|employee_skills)\M')
 AND NOT EXISTS(SELECT 1 FROM pg_depend d WHERE d.refclassid='pg_class'::regclass
 AND d.refobjid IN ('public.skill_categories'::regclass,'public.skills'::regclass,'public.employee_skills'::regclass)
 AND d.deptype='n' AND d.classid NOT IN ('pg_constraint'::regclass,'pg_class'::regclass,'pg_attrdef'::regclass,'pg_trigger'::regclass)) INTO valid;
  IF valid IS NOT TRUE THEN RAISE EXCEPTION 'MSK: ràng buộc hoặc phụ thuộc nguồn cần review'; END IF;

  SELECT
    (SELECT count(*)=9 FROM information_schema.columns WHERE table_schema='public' AND table_name='audit_events')
    AND NOT EXISTS(SELECT 1 FROM (VALUES
      ('audit_id','integer','NO'),('actor_id','integer','NO'),('entity_id','integer','NO'),
      ('entity_type','text','NO'),('operation','text','NO'),('occurred_at','timestamp with time zone','NO'),
      ('reason','text','YES'),('before_values','jsonb','YES'),('after_values','jsonb','YES')) e(col,typ,nullable)
      WHERE NOT EXISTS(SELECT 1 FROM information_schema.columns c WHERE c.table_schema='public'
        AND c.table_name='audit_events' AND c.column_name=e.col AND c.data_type=e.typ AND c.is_nullable=e.nullable))
    AND (SELECT is_identity='YES' AND identity_generation='ALWAYS' FROM information_schema.columns
      WHERE table_schema='public' AND table_name='audit_events' AND column_name='audit_id')
    AND NOT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name='audit_events'
      AND (is_generated<>'NEVER' OR (column_name='occurred_at' AND column_default IS DISTINCT FROM 'now()')
        OR (column_name<>'occurred_at' AND column_default IS NOT NULL)))
    AND (SELECT count(*)=2 FROM pg_index WHERE indrelid='public.audit_events'::regclass)
    AND EXISTS(SELECT 1 FROM pg_indexes WHERE schemaname='public' AND tablename='audit_events'
      AND indexdef='CREATE INDEX idx_audit_events_entity ON public.audit_events USING btree (entity_type, entity_id, occurred_at)')
    AND NOT EXISTS(SELECT 1 FROM pg_index WHERE indrelid='public.audit_events'::regclass
      AND (NOT indisvalid OR NOT indisready OR indpred IS NOT NULL OR indexprs IS NOT NULL))
    AND (SELECT count(*)=1 FROM pg_constraint c WHERE c.conrelid='public.audit_events'::regclass
      AND c.contype='p' AND c.conkey=ARRAY[(SELECT attnum FROM pg_attribute
        WHERE attrelid='public.audit_events'::regclass AND attname='audit_id')]::smallint[])
    AND (SELECT count(*)=1 FROM pg_constraint c WHERE c.conrelid='public.audit_events'::regclass
      AND c.contype='f' AND c.confrelid='public.users'::regclass AND c.confdeltype='a' AND c.confupdtype='a'
      AND c.conkey=ARRAY[(SELECT attnum FROM pg_attribute WHERE attrelid='public.audit_events'::regclass AND attname='actor_id')]::smallint[]
      AND c.confkey=ARRAY[(SELECT attnum FROM pg_attribute WHERE attrelid='public.users'::regclass AND attname='user_id')]::smallint[])
    AND NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conrelid='public.audit_events'::regclass
      AND (contype NOT IN ('p','f','n') OR NOT convalidated OR condeferrable))
    AND (SELECT count(*)=1 FROM pg_constraint WHERE conrelid='public.audit_events'::regclass AND contype='f')
    AND NOT EXISTS(SELECT 1 FROM pg_trigger WHERE tgrelid='public.audit_events'::regclass AND NOT tgisinternal)
    AND NOT EXISTS(SELECT 1 FROM pg_policy WHERE polrelid='public.audit_events'::regclass)
    AND NOT EXISTS(SELECT 1 FROM pg_inherits WHERE inhrelid='public.audit_events'::regclass OR inhparent='public.audit_events'::regclass)
    AND NOT EXISTS(SELECT 1 FROM pg_class WHERE oid='public.audit_events'::regclass
      AND (relkind<>'r' OR relrowsecurity OR relforcerowsecurity)) INTO valid;
  IF valid IS NOT TRUE THEN RAISE EXCEPTION 'MSK: audit import cần review schema hoặc dependency'; END IF;

  WITH names(kind,name) AS (
 SELECT 'category',category_name FROM public.skill_categories UNION ALL SELECT 'skill',skill_name FROM public.skills
 ) SELECT EXISTS(SELECT 1 FROM names GROUP BY kind,lower(name) HAVING count(*)>1
 OR bool_or(length(name) NOT BETWEEN 1 AND 100)
 OR bool_or(name ~ U&'^[\0009\000A\000B\000C\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]|[\0009\000A\000B\000C\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]$'))
 OR EXISTS(SELECT 1 FROM public.skills WHERE status::text<>'ACTIVE'),
 EXISTS(SELECT 1 FROM public.employee_skills s
 LEFT JOIN public.employees e ON e.employee_id=s.approved_by_employee_id
 LEFT JOIN public.users u ON u.user_id=e.user_id
 WHERE s.source::text NOT IN ('MANUAL','CV') OR s.approved_at IS NULL OR u.user_id IS NULL
 OR s.proficiency_level::text NOT IN ('BASIC','INTERMEDIATE','ADVANCED','EXPERT')
 OR s.years_experience<0 OR s.years_experience>999.9),
 EXISTS(SELECT 1 FROM public.employee_skills) AND (import_actor<=0 OR NOT EXISTS(SELECT 1 FROM public.users WHERE user_id=import_actor)) INTO invalid_names, invalid_approval, missing_actor;
  IF invalid_names IS NOT FALSE THEN
    RAISE EXCEPTION 'MSK: tên danh mục hoặc trạng thái legacy cần mapping được review';
  END IF;
  IF invalid_approval IS NOT FALSE THEN
    RAISE EXCEPTION 'MSK: nguồn hoặc liên kết approval legacy không hợp lệ';
  END IF;
  IF missing_actor IS NOT FALSE THEN
    RAISE EXCEPTION 'MSK: thiếu actor import đã xác minh';
  END IF;
END $$;
-- MSK-GUARD-END

CREATE FUNCTION public.msk_text_valid(value text, max_length integer)
RETURNS boolean LANGUAGE sql IMMUTABLE SET search_path = pg_catalog AS $$
  SELECT (value IS NOT NULL AND length(value) BETWEEN 1 AND max_length
    AND value !~ U&'^[\0009\000A\000B\000C\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]|[\0009\000A\000B\000C\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]$') IS TRUE
$$;

ALTER TABLE public.skill_categories ADD CONSTRAINT ck_msk_category_name
  CHECK (public.msk_text_valid(category_name,100));
ALTER TABLE public.skills ADD CONSTRAINT ck_msk_skill_name
  CHECK (public.msk_text_valid(skill_name,100));
CREATE UNIQUE INDEX uq_msk_category_name ON public.skill_categories(lower(category_name));
CREATE UNIQUE INDEX uq_msk_skill_name ON public.skills(lower(skill_name));

DO $$
DECLARE imported_count bigint;
BEGIN
INSERT INTO public.audit_events(actor_id,entity_type,entity_id,operation,before_values,after_values)
SELECT current_setting('app.msk_import_actor_id',true)::integer,'EMPLOYEE_SKILL',s.employee_skill_id,
  'SKILL_LEGACY_IMPORTED',
  jsonb_build_object('source',s.source,'approved_by_employee_id',s.approved_by_employee_id,
    'approved_at',s.approved_at),
  jsonb_build_object('approved_by',e.user_id)
FROM public.employee_skills s
JOIN public.employees e ON e.employee_id=s.approved_by_employee_id;

  GET DIAGNOSTICS imported_count = ROW_COUNT;
  IF imported_count<>(SELECT count(*) FROM public.employee_skills) OR EXISTS(
    SELECT 1 FROM public.employee_skills s
    JOIN public.employees e ON e.employee_id=s.approved_by_employee_id
    WHERE NOT EXISTS(SELECT 1 FROM public.audit_events a WHERE a.entity_type='EMPLOYEE_SKILL'
      AND a.entity_id=s.employee_skill_id AND a.operation='SKILL_LEGACY_IMPORTED'
      AND a.actor_id=current_setting('app.msk_import_actor_id',true)::integer
      AND a.before_values=jsonb_build_object('source',s.source,
        'approved_by_employee_id',s.approved_by_employee_id,'approved_at',s.approved_at)
      AND a.after_values=jsonb_build_object('approved_by',e.user_id))) THEN
    RAISE EXCEPTION 'MSK: không lưu đủ provenance legacy';
  END IF;
END $$;

ALTER TABLE public.employee_skills ADD COLUMN approved_by integer;
-- Backfill không phải lần sửa năng lực; giữ nguyên updated_at của bản ghi cũ.
ALTER TABLE public.employee_skills DISABLE TRIGGER trg_employee_skills_updated_at;
UPDATE public.employee_skills s SET approved_by=e.user_id
FROM public.employees e WHERE e.employee_id=s.approved_by_employee_id;
ALTER TABLE public.employee_skills ENABLE TRIGGER trg_employee_skills_updated_at;
ALTER TABLE public.employee_skills
  ALTER COLUMN approved_by SET NOT NULL,
  ALTER COLUMN approved_at SET NOT NULL,
  ADD CONSTRAINT fk_msk_approval_user FOREIGN KEY(approved_by) REFERENCES public.users(user_id);
DO $$
DECLARE employee_fk text;
BEGIN
  SELECT c.conname INTO STRICT employee_fk FROM pg_constraint c
  JOIN pg_attribute a ON a.attrelid=c.conrelid AND a.attnum=c.conkey[1]
  WHERE c.conrelid='public.employee_skills'::regclass AND c.contype='f'
    AND cardinality(c.conkey)=1 AND a.attname='employee_id';
  EXECUTE format('ALTER TABLE public.employee_skills DROP CONSTRAINT %I',employee_fk);
END $$;
ALTER TABLE public.employee_skills
  ADD CONSTRAINT employee_skills_employee_id_fkey FOREIGN KEY(employee_id) REFERENCES public.employees(employee_id),
  DROP COLUMN source,
  DROP COLUMN approved_by_employee_id;
DROP INDEX public.idx_skills_status;
ALTER TABLE public.skills DROP COLUMN status;

-- Before giữ note legacy nguyên gốc; giới hạn note mới nằm ở phép so before/proposed.
CREATE FUNCTION public.msk_values_valid(value jsonb)
RETURNS boolean LANGUAGE sql IMMUTABLE SET search_path = pg_catalog AS $$
  SELECT (jsonb_typeof(value)='object'
    AND value ?& ARRAY['schema_version','proficiency_level','years_experience','note']
    AND value - ARRAY['schema_version','proficiency_level','years_experience','note']='{}'::jsonb
    AND value->'schema_version'='1'::jsonb
    AND jsonb_typeof(value->'proficiency_level')='string'
    AND value->>'proficiency_level' IN ('BASIC','INTERMEDIATE','ADVANCED','EXPERT')
    AND (value->'years_experience'='null'::jsonb OR
      (jsonb_typeof(value->'years_experience')='string'
       AND value->>'years_experience' ~ '^(0|[1-9][0-9]{0,2})\.[0-9]$'))
    AND (value->'note'='null'::jsonb OR jsonb_typeof(value->'note')='string')) IS TRUE
$$;

CREATE FUNCTION public.msk_request_values_valid(operation text, before_values jsonb, proposed_values jsonb)
RETURNS boolean LANGUAGE sql IMMUTABLE SET search_path = pg_catalog, public AS $$
  SELECT (public.msk_values_valid(proposed_values)
    AND ((operation='ADD' AND before_values IS NULL)
      OR (operation='UPDATE' AND public.msk_values_valid(before_values)
        AND before_values<>proposed_values))
    AND (proposed_values->'note'='null'::jsonb
      OR public.msk_text_valid(proposed_values->>'note',2000)
      OR (operation='UPDATE' AND proposed_values->'note'=before_values->'note'))) IS TRUE
$$;

CREATE TABLE public.employee_skill_change_requests (
  request_id integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  employee_id integer NOT NULL REFERENCES public.employees(employee_id),
  skill_id integer NOT NULL REFERENCES public.skills(skill_id),
  operation text NOT NULL CHECK(operation IN ('ADD','UPDATE')),
  source text NOT NULL DEFAULT 'MANUAL' CHECK(source='MANUAL'),
  before_values jsonb,
  proposed_values jsonb NOT NULL,
  explanation text NOT NULL CHECK(public.msk_text_valid(explanation,2000)),
  external_evidence text CHECK(external_evidence IS NULL OR public.msk_text_valid(external_evidence,2000)),
  expert_opinion text CHECK(expert_opinion IS NULL OR public.msk_text_valid(expert_opinion,2000)),
  requested_by integer NOT NULL REFERENCES public.users(user_id),
  requested_at timestamptz NOT NULL DEFAULT now(),
  status text NOT NULL DEFAULT 'PENDING' CHECK(status IN ('PENDING','APPROVED','REJECTED')),
  resolved_by integer REFERENCES public.users(user_id),
  resolved_at timestamptz,
  decision_reason text,
  version integer NOT NULL DEFAULT 1 CHECK(version>0),
  CONSTRAINT ck_msk_values CHECK(public.msk_request_values_valid(operation,before_values,proposed_values)),
  CONSTRAINT ck_msk_decision CHECK((
    (status='PENDING' AND resolved_by IS NULL AND resolved_at IS NULL
      AND decision_reason IS NULL AND expert_opinion IS NULL)
    OR (status IN ('APPROVED','REJECTED') AND resolved_by IS NOT NULL
      AND resolved_by<>requested_by AND resolved_at IS NOT NULL AND resolved_at>=requested_at
      AND public.msk_text_valid(decision_reason,2000))
  ) IS TRUE)
);
CREATE UNIQUE INDEX uq_msk_pending ON public.employee_skill_change_requests(employee_id,skill_id)
  WHERE status='PENDING';
CREATE INDEX idx_msk_owner ON public.employee_skill_change_requests(employee_id,requested_at DESC,request_id DESC);
CREATE INDEX idx_msk_review ON public.employee_skill_change_requests(status,requested_at DESC,request_id DESC);
CREATE INDEX idx_msk_employee_skill ON public.employee_skill_change_requests(employee_id,skill_id,requested_at DESC,request_id DESC);
COMMIT;
