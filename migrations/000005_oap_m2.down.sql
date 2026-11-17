-- Không phục hồi dữ liệu username. Chỉ dùng trên DB case rỗng chưa cấp số.
BEGIN;
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';
LOCK TABLE public.users, public.employees IN ACCESS EXCLUSIVE MODE;
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.users) OR EXISTS (SELECT 1 FROM public.employees)
       OR (SELECT is_called FROM public.employee_code_seq) THEN
        RAISE EXCEPTION 'M2: từ chối down có dữ liệu hoặc sequence từng cấp số';
    END IF;
END $$;
ALTER TABLE public.users ADD COLUMN username varchar(50) NOT NULL UNIQUE;
DROP INDEX public.uq_employees_normalized_email;
DROP SEQUENCE public.employee_code_seq;
COMMIT;
