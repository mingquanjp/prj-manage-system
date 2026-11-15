-- DATN-24: danh mục, hồ sơ và nền audit.
BEGIN;
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';

LOCK TABLE public.departments, public.positions, public.employees IN SHARE ROW EXCLUSIVE MODE;

DO $$
DECLARE
    approved_map jsonb := COALESCE(NULLIF(current_setting('app.oap_m1_job_levels', true), ''), '{}')::jsonb;
BEGIN
    IF EXISTS (
        SELECT 1 FROM public.employees
        WHERE seniority_level ~ U&'[^\0009\000A\000B\000C\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]' AND NOT (approved_map ? seniority_level)
    ) THEN
        RAISE EXCEPTION 'M1: thiếu mapping cấp bậc đã xác minh';
    END IF;

    IF EXISTS (
        SELECT 1 FROM jsonb_each(approved_map) AS m(legacy, name)
        WHERE jsonb_typeof(name) <> 'string'
           OR length(name #>> '{}') NOT BETWEEN 1 AND 100
           OR (name #>> '{}') !~ U&'[^\0009\000A\000B\000C\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]'
           OR (name #>> '{}') ~ U&'^[\0009\000A\000B\000C\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]|[\0009\000A\000B\000C\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]$'
    ) THEN
        RAISE EXCEPTION 'M1: tên cấp bậc trong mapping không hợp lệ';
    END IF;
END $$;

CREATE TABLE public.job_levels (
    job_level_id int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    job_level_name varchar(100) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT ck_job_levels_name CHECK (job_level_name ~ U&'[^\0009\000A\000B\000C\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]')
);

ALTER TABLE public.departments ADD CONSTRAINT ck_departments_name CHECK (department_name ~ U&'[^\0009\000A\000B\000C\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]');
ALTER TABLE public.positions ADD CONSTRAINT ck_positions_name CHECK (position_name ~ U&'[^\0009\000A\000B\000C\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]');

CREATE UNIQUE INDEX uq_departments_normalized_name ON public.departments (lower(btrim(department_name)));
CREATE UNIQUE INDEX uq_positions_normalized_name ON public.positions (lower(btrim(position_name)));
CREATE UNIQUE INDEX uq_job_levels_normalized_name ON public.job_levels (lower(btrim(job_level_name)));

ALTER TABLE public.employees
    ADD COLUMN job_level_id int REFERENCES public.job_levels(job_level_id),
    ADD COLUMN address text CONSTRAINT ck_employees_address CHECK (address ~ U&'[^\0009\000A\000B\000C\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]');
CREATE INDEX idx_employees_job_level ON public.employees (job_level_id);

-- Mapping tường minh truyền theo phiên migration. Chỉ tạo cấp bậc có hồ sơ dùng.
WITH approved AS (
    SELECT * FROM jsonb_each_text(COALESCE(NULLIF(current_setting('app.oap_m1_job_levels', true), ''), '{}')::jsonb)
)
INSERT INTO public.job_levels (job_level_name)
SELECT DISTINCT approved.value FROM approved
WHERE EXISTS (SELECT 1 FROM public.employees WHERE seniority_level = approved.key AND seniority_level ~ U&'[^\0009\000A\000B\000C\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]');

-- Backfill kỹ thuật giữ nguyên thời điểm cập nhật hồ sơ legacy.
ALTER TABLE public.employees DISABLE TRIGGER trg_employees_updated_at;

UPDATE public.employees AS e SET job_level_id = j.job_level_id
FROM public.job_levels AS j
WHERE e.seniority_level ~ U&'[^\0009\000A\000B\000C\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]'
  AND j.job_level_name = COALESCE(NULLIF(current_setting('app.oap_m1_job_levels', true), ''), '{}')::jsonb ->> e.seniority_level;

ALTER TABLE public.employees ENABLE TRIGGER trg_employees_updated_at;

CREATE TABLE public.audit_events (
    audit_id int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    actor_id int NOT NULL REFERENCES public.users(user_id),
    entity_type text NOT NULL,
    entity_id int NOT NULL,
    operation text NOT NULL,
    occurred_at timestamptz NOT NULL DEFAULT now(),
    reason text,
    before_values jsonb,
    after_values jsonb
);
CREATE INDEX idx_audit_events_entity ON public.audit_events (entity_type, entity_id, occurred_at);

COMMIT;
