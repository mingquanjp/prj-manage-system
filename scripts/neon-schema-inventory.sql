-- Kiểm kê cấu trúc, không migration hoặc đọc nội dung hồ sơ/mật khẩu.
-- Chọn đúng project/branch/database trước khi chạy bằng SQL Editor hoặc psql.
-- Yêu cầu public.schema_migrations đã tồn tại; kiểm tra danh sách bảng trước.
-- READ ONLY ngăn ghi; kết quả chỉ phản ánh một thời điểm, không là bản backup.
BEGIN TRANSACTION ISOLATION LEVEL REPEATABLE READ READ ONLY;
SET LOCAL statement_timeout = '15s';
SELECT jsonb_build_object(
 'server_version', current_setting('server_version'),
 'migration', (SELECT jsonb_agg(jsonb_build_object('version',version,'dirty',dirty)) FROM public.schema_migrations),
 'columns', (SELECT jsonb_agg(jsonb_build_object('table',table_name,'column',column_name,'position',ordinal_position,'type',data_type,'udt',udt_name,'nullable',is_nullable,'identity',is_identity,'length',character_maximum_length,'precision',numeric_precision,'scale',numeric_scale,'has_default',column_default IS NOT NULL) ORDER BY table_name,ordinal_position) FROM information_schema.columns WHERE table_schema='public'),
 'constraints', (SELECT jsonb_agg(jsonb_build_object('table',r.relname,'name',c.conname,'type',c.contype,'definition',pg_get_constraintdef(c.oid),'validated',c.convalidated) ORDER BY r.relname,c.conname) FROM pg_constraint c JOIN pg_class r ON r.oid=c.conrelid JOIN pg_namespace n ON n.oid=r.relnamespace WHERE n.nspname='public'),
 'indexes', (SELECT jsonb_agg(jsonb_build_object('table',tablename,'name',indexname,'definition',indexdef) ORDER BY tablename,indexname) FROM pg_indexes WHERE schemaname='public'),
 'enums', (SELECT jsonb_agg(jsonb_build_object('name',t.typname,'value',e.enumlabel,'position',e.enumsortorder) ORDER BY t.typname,e.enumsortorder) FROM pg_type t JOIN pg_enum e ON e.enumtypid=t.oid JOIN pg_namespace n ON n.oid=t.typnamespace WHERE n.nspname='public'),
 'relations', (SELECT jsonb_agg(jsonb_build_object('name',c.relname,'kind',c.relkind,'row_security',c.relrowsecurity) ORDER BY c.relname) FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='public' AND c.relkind IN ('r','p','v','m')),
 'triggers', (SELECT COALESCE(jsonb_agg(jsonb_build_object('table',c.relname,'name',t.tgname)), '[]'::jsonb) FROM pg_trigger t JOIN pg_class c ON c.oid=t.tgrelid JOIN pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='public' AND NOT t.tgisinternal)
) AS inventory;
COMMIT;
