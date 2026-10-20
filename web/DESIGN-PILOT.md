# Lát cắt thiết kế — 20/10/2026

> **Đã hoàn tất phạm vi; người dùng đã duyệt visual 20/10/2026.** Đây là bằng chứng pilot, không phải kế hoạch mở rộng còn chờ duyệt. Quy trình và đợt tiếp theo xem [README](README.md).

## Phạm vi đã được duyệt

Giữ màu ấm; học cách tổ chức workspace, bảng và side peek của Notion, không sao chép logo hay toàn bộ chức năng. Dựng shell, danh sách dự án, công việc và phân công để duyệt trước khi mở rộng. Mã nguồn chính ở `web/`; `/Users/quanjp/Documents/DATN` chỉ đối chiếu.

## Nguồn và ranh giới

- Nguồn nghiệp vụ: `../CLAUDE.md`, `../.claude/docs/domain_rules.md`, `project_task_assignment_workflow.md`, `ai_recommendation_workflow.md`, `roles_permissions.md`.
- Nguồn thiết kế: `../docs/design/figma-make/Guidelines.md`, `ui-brief.md`, `screen-contracts.md`; bố cục mới theo đề xuất người dùng vừa duyệt.
- Demo chỉ trong bộ nhớ, tải lại sẽ khôi phục mẫu. Điểm là fixture minh họa, không công thức cuối. Không tăng tải khi chỉ chọn nháp; xác nhận mô phỏng mới cập nhật toàn bộ lựa chọn.
- Chưa làm tạo/sửa task/dự án, bổ nhiệm, vòng đời dự án, nghiệm thu, nhật ký, hồ sơ, quản trị, API/auth, tích hợp Go hay đồng bộ Make. Các hành động ngoài lát cắt phải được ghi rõ chưa dựng, không giả thành công.
- Bản thử chỉ phân công lần đầu task chờ; không thay người/thu hồi. PM-only có góc nhìn chỉ đọc; danh tính mô phỏng không phải tự cấp quyền.

## Tiến độ và cổng duyệt

1. Dựng nền local và primitives tái sử dụng.
2. Shell/danh sách → bảng/chi tiết bên cạnh → workspace phân công.
3. Typecheck, unit test các bất biến mẫu, build và kiểm tra trực quan các luồng chính.
4. Người dùng duyệt mật độ, bố cục, cách so sánh ứng viên trước khi mở rộng.

## Chạy

`pnpm install` rồi `pnpm dev` tại `web/`, port 5173 (không chiếm 8443 của Make). `pnpm check` chạy typecheck, test và build. Các lệnh này không truy cập DB. Chưa có lint riêng; chưa gọi đây là frontend production hoàn chỉnh.

## Kết quả kiểm tra lát cắt

- Đã dựng cả ba vùng trong phạm vi; visual đã được người dùng duyệt, chưa mở rộng toàn bộ web.
- `pnpm check`: TypeScript qua, 5 unit test qua, Vite build qua. Test chỉ bảo vệ vài bất biến demo, không chứng nhận đầy đủ nghiệp vụ/server transaction.
- Đã xem desktop 1440px: danh sách dự án → bảng công việc → chi tiết bên cạnh, tìm kiếm và chọn nhiều việc → gợi ý → draft tổng tải → dialog xác nhận → cả hai task cập nhật cùng lượt.
- Đã xem manual: không cột điểm/hạng, loại PTCM cùng dự án. PM-only không có checkbox/hành động phân công. Console không ghi nhận error/warn trong lượt thử.
- Đã kiểm tra viewport 390px: không tràn ngang toàn trang; bảng cuộn trong vùng riêng. Đây vẫn là thiết kế ưu tiên desktop, chưa nghiệm thu toàn bộ thiết bị/trình đọc màn hình.
- Chưa kiểm thử đầy đủ các ngoại lệ, chưa có API/auth/persistence, engine thật hoặc test E2E tự động. Danh mục chức năng ngoài lát cắt vẫn chưa dựng, không coi nút disabled là chức năng hoàn thiện.
- Không chỉnh Make/export, backend, DBML, migration hoặc công thức nghiệp vụ.
