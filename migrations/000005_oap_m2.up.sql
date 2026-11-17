-- DATN-30: dừng consumer ghi và chạy preflight M2 bằng runner trước file này.
BEGIN;
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';
LOCK TABLE public.users, public.employees IN ACCESS EXCLUSIVE MODE;

CREATE UNIQUE INDEX uq_employees_normalized_email ON public.employees (lower(btrim(email)));
CREATE SEQUENCE public.employee_code_seq AS bigint MINVALUE 1 INCREMENT 1 NO CYCLE CACHE 1;

-- Numeric tránh overflow khi đọc mã legacy; preflight từ chối max BIGINT.
DO $$
DECLARE next_code numeric;
BEGIN
    SELECT COALESCE(max(CASE WHEN employee_code ~ '^EMP[0-9]+$' THEN substring(employee_code FROM 4)::numeric END), 0) + 1
      INTO next_code FROM public.employees WHERE employee_code ~ '^EMP[0-9]+$';
    IF next_code > 9223372036854775807 THEN
        RAISE EXCEPTION 'M2: dải mã nhân viên đã hết';
    END IF;
    PERFORM setval('public.employee_code_seq', next_code::bigint, false);
END $$;

ALTER TABLE public.users DROP COLUMN username;
COMMIT;
