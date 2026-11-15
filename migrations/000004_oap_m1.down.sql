-- DATN-24: down loại bỏ dữ liệu job_levels, địa chỉ và audit của M1.
BEGIN;
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';

DROP TABLE public.audit_events;
ALTER TABLE public.employees DROP COLUMN address, DROP COLUMN job_level_id;
DROP TABLE public.job_levels;
DROP INDEX public.uq_positions_normalized_name, public.uq_departments_normalized_name;
ALTER TABLE public.positions DROP CONSTRAINT ck_positions_name;
ALTER TABLE public.departments DROP CONSTRAINT ck_departments_name;

COMMIT;
