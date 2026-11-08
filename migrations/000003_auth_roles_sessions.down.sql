-- Chỉ dùng sau khi dừng consumer và chấp nhận mất toàn bộ phiên/quyền mới.
-- Down/up không khôi phục role cấp sau backfill; không tự chạy trên DB thật.
BEGIN;
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';
DROP TABLE public.auth_sessions;
DROP TABLE public.user_roles;
COMMIT;
