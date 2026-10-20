# Frontend — mở rộng từ pilot đã duyệt

- React/TypeScript/Vite/Tailwind; primitives dùng mô hình shadcn/Radix và token tùy chỉnh. Không bê cấu hình Make sang local.
- `src/demo/` là adapter dữ liệu mô phỏng, không API/auth thật. Không coi kiểm tra quyền client là bảo mật production.
- Giữ bảng màu ấm hiện hành, UI tiếng Việt, thao tác bàn phím/focus rõ. Không thêm KPI, roster, reviewer, sprint hay tính năng từ ảnh tham khảo.
- PM/PTCM là trách nhiệm theo dự án. PM không kiêm PTCM không phân công; PTCM không tự nhận việc.
- Đọc [web/README](README.md): chuẩn thiết kế đã duyệt, vòng làm việc, mẫu đợt và sổ bàn giao. [DESIGN-PILOT](DESIGN-PILOT.md) là báo cáo pilot đã hoàn thành phạm vi, không kế hoạch feature hiện hành.
- Chỉ mở rộng phần người dùng đã giao; ghi mục tiêu/nguồn/ngoài phạm vi/tiêu chí đạt vào sổ trước sửa. Không suy duyệt phong cách là duyệt mọi màn hình.
- Dùng pnpm/lockfile; chọn check theo rủi ro trong README, không kiểm thử mọi ngoại lệ MVP cho thay đổi layout.
- Không thêm dependency cho một tính năng chưa được yêu cầu. TanStack Query/API adapter sẽ bổ sung khi có API, không giả backend.
