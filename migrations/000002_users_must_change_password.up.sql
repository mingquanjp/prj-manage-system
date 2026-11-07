-- Chốt cuối 06/11/2026: mật khẩu tạm chưa có hạn, buộc đổi trước nghiệp vụ.
-- Tài khoản cũ giữ nguyên; khi cấp mật khẩu tạm, service phải chủ động ghi true.
-- Cột này không tự thực thi chặn API; không đồng bộ các thay đổi khác của 33 bảng.
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';

ALTER TABLE public.users
    ADD COLUMN must_change_password boolean NOT NULL DEFAULT false;
