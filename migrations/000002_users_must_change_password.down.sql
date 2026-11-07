-- Chỉ rollback sau khi ngừng mọi consumer phụ thuộc cờ buộc đổi.
-- Drop cột làm mất trạng thái buộc đổi; không chạy tự động trên DB thật.
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';

ALTER TABLE public.users DROP COLUMN must_change_password;
