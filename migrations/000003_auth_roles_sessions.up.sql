-- Nền auth theo DATN-7; chỉ diễn tập trên DB test sở hữu trong đợt này.
-- Áp DB đích thật cần xác nhận mapping/backup/consumer riêng; không tự chạy.
BEGIN;
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';

CREATE TABLE public.user_roles (
    user_id int NOT NULL REFERENCES public.users (user_id),
    role_code text NOT NULL,
    CONSTRAINT pk_user_roles PRIMARY KEY (user_id, role_code),
    CONSTRAINT ck_user_roles_code CHECK (role_code IN ('ADMIN', 'HR', 'DIRECTOR'))
);

CREATE TABLE public.auth_sessions (
    session_id int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id int NOT NULL REFERENCES public.users (user_id),
    token_hash bytea NOT NULL UNIQUE,
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    CONSTRAINT ck_auth_sessions_hash CHECK (octet_length(token_hash) = 32),
    CONSTRAINT ck_auth_sessions_expiry CHECK (expires_at > created_at),
    CONSTRAINT ck_auth_sessions_revoke CHECK (
        revoked_at IS NULL OR revoked_at >= created_at
    )
);
CREATE INDEX idx_auth_sessions_user ON public.auth_sessions (user_id);

-- Chỉ ADMIN có ánh xạ quyền legacy; PM/PTCM không là quyền hệ thống.
INSERT INTO public.user_roles (user_id, role_code)
SELECT user_id, 'ADMIN'
FROM public.users
WHERE system_role = 'ADMIN';
COMMIT;
