BEGIN;
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';
LOCK TABLE public.notifications IN ACCESS EXCLUSIVE MODE;
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM public.notifications) THEN
    RAISE EXCEPTION 'PCA: notification legacy cần mapping đã review';
  END IF;
END $$;

CREATE FUNCTION public.pca_profile_values_valid(v jsonb, proposed boolean)
RETURNS boolean LANGUAGE plpgsql IMMUTABLE
SET search_path = pg_catalog AS $$
DECLARE k text; item jsonb; t text; max_len integer;
BEGIN
  IF v IS NULL OR jsonb_typeof(v) <> 'object' THEN RETURN false; END IF;
  IF v->'schema_version' IS DISTINCT FROM '1'::jsonb THEN RETURN false; END IF;
  IF (v - ARRAY['schema_version','fullname','phone','address']) <> '{}'::jsonb
     OR (v - 'schema_version') = '{}'::jsonb THEN RETURN false; END IF;
  FOR k,item IN SELECT key,value FROM jsonb_each(v - 'schema_version') LOOP
    IF item = 'null'::jsonb THEN
      IF k='fullname' OR (proposed AND k='address') THEN RETURN false; END IF;
      CONTINUE;
    END IF;
    IF jsonb_typeof(item) <> 'string' THEN RETURN false; END IF;
    t := item #>> '{}';
    IF k='fullname' AND length(t)=0 THEN RETURN false; END IF;
    IF proposed THEN
      max_len := CASE k WHEN 'fullname' THEN 100 WHEN 'phone' THEN 20 ELSE 2000 END;
      IF length(t) NOT BETWEEN 1 AND max_len
         OR t !~ U&'[^\0009\000A\000B\000C\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]'
         OR t ~ U&'^[\0009\000A\000B\000C\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]|[\0009\000A\000B\000C\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]$'
      THEN RETURN false; END IF;
    END IF;
  END LOOP;
  RETURN true;
END $$;

CREATE FUNCTION public.pca_same_profile_keys(a jsonb,b jsonb)
RETURNS boolean LANGUAGE plpgsql IMMUTABLE
SET search_path = pg_catalog AS $$
BEGIN
  IF a IS NULL OR b IS NULL OR jsonb_typeof(a)<>'object'
     OR jsonb_typeof(b)<>'object' THEN RETURN false; END IF;
  RETURN NOT EXISTS (
    SELECT 1 FROM (
      (SELECT jsonb_object_keys(a) EXCEPT SELECT jsonb_object_keys(b))
      UNION ALL
      (SELECT jsonb_object_keys(b) EXCEPT SELECT jsonb_object_keys(a))
    ) AS differing_keys
  );
END $$;

CREATE TABLE public.employee_profile_change_requests (
  request_id int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  employee_id int NOT NULL REFERENCES public.employees(employee_id),
  before_values jsonb NOT NULL,
  proposed_values jsonb NOT NULL,
  explanation text,
  requested_by int NOT NULL REFERENCES public.users(user_id),
  requested_at timestamptz NOT NULL DEFAULT now(),
  status text NOT NULL DEFAULT 'PENDING',
  resolved_by int REFERENCES public.users(user_id),
  resolved_at timestamptz,
  decision_reason text,
  CONSTRAINT ck_pca_values CHECK (
    public.pca_profile_values_valid(before_values,false)
    AND public.pca_profile_values_valid(proposed_values,true)
    AND public.pca_same_profile_keys(before_values,proposed_values)
    AND before_values <> proposed_values
  ),
  CONSTRAINT ck_pca_status CHECK (status IN ('PENDING','APPROVED','REJECTED')),
  CONSTRAINT ck_pca_decision CHECK ((
    (status='PENDING' AND resolved_by IS NULL AND resolved_at IS NULL AND decision_reason IS NULL)
    OR (status='APPROVED' AND resolved_by IS NOT NULL AND resolved_at IS NOT NULL AND decision_reason IS NULL)
    OR (status='REJECTED' AND resolved_by IS NOT NULL AND resolved_at IS NOT NULL
      AND decision_reason IS NOT NULL AND length(decision_reason) BETWEEN 1 AND 2000
      AND decision_reason ~ U&'[^\0009\000A\000B\000C\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]')
  ) IS TRUE),
  CONSTRAINT ck_pca_times CHECK (resolved_at IS NULL OR resolved_at>=requested_at),
  CONSTRAINT ck_pca_explanation CHECK (explanation IS NULL OR length(explanation)<=2000)
);
CREATE INDEX idx_pca_owner ON public.employee_profile_change_requests
  (employee_id,requested_at DESC,request_id DESC);
CREATE INDEX idx_pca_queue ON public.employee_profile_change_requests
  (status,requested_at DESC,request_id DESC);

CREATE TABLE public.idempotency_records (
  idempotency_id int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  actor_id int NOT NULL REFERENCES public.users(user_id),
  operation varchar(100) NOT NULL,
  idempotency_key varchar(128) NOT NULL,
  request_hash bytea NOT NULL CHECK (octet_length(request_hash)=32),
  result_reference jsonb NOT NULL CHECK (jsonb_typeof(result_reference)='object'),
  created_at timestamptz NOT NULL DEFAULT now(),
  expires_at timestamptz,
  CONSTRAINT uq_idempotency_actor_operation_key UNIQUE(actor_id,operation,idempotency_key),
  CONSTRAINT ck_idempotency_expiry CHECK (expires_at IS NULL OR expires_at>created_at)
);

-- Bảng đang rỗng và đã được khóa/guard; không reset sequence notification_id.
DROP INDEX public.idx_notifications_employee;
ALTER TABLE public.notifications
  DROP COLUMN employee_id,
  DROP COLUMN type,
  DROP COLUMN title,
  DROP COLUMN content,
  DROP COLUMN ref_task_id,
  DROP COLUMN is_read,
  ADD COLUMN recipient_id int NOT NULL REFERENCES public.users(user_id),
  ADD COLUMN notification_type text NOT NULL,
  ADD COLUMN resource_type text NOT NULL,
  ADD COLUMN resource_id int NOT NULL CHECK (resource_id>0),
  ADD COLUMN message text NOT NULL CHECK (length(message) BETWEEN 1 AND 200),
  ADD COLUMN read_at timestamptz,
  ADD CONSTRAINT ck_notifications_read_at CHECK (read_at IS NULL OR read_at>=created_at);
CREATE INDEX idx_notifications_recipient ON public.notifications
  (recipient_id,created_at DESC,notification_id DESC);
COMMIT;
