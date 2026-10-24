# Frontend — mở rộng từ pilot đã duyệt

- React/TypeScript/Vite/Tailwind; primitives dùng mô hình shadcn/Radix và token tùy chỉnh. Không bê cấu hình Make sang local.
- `src/demo/` là adapter dữ liệu mô phỏng, không API/auth thật. Không coi kiểm tra quyền client là bảo mật production.
- Giữ bảng màu ấm hiện hành, UI tiếng Việt, thao tác bàn phím/focus rõ. Không thêm KPI, roster, reviewer, sprint hay tính năng từ ảnh tham khảo.
- PM/PTCM là trách nhiệm theo dự án. PM không kiêm PTCM không phân công; PTCM không tự nhận việc.
- Đọc [web/README](README.md): chuẩn thiết kế đã duyệt, vòng làm việc, mẫu đợt và sổ bàn giao. [DESIGN-PILOT](DESIGN-PILOT.md) là báo cáo pilot đã hoàn thành phạm vi, không kế hoạch feature hiện hành.
- Chỉ mở rộng phần người dùng đã giao; ghi mục tiêu/nguồn/ngoài phạm vi/tiêu chí đạt vào sổ trước sửa. Không suy duyệt phong cách là duyệt mọi màn hình.
- Dùng pnpm/lockfile; chọn check theo rủi ro trong README, không kiểm thử mọi ngoại lệ MVP cho thay đổi layout.
- Không thêm dependency cho một tính năng chưa được yêu cầu. TanStack Query/API adapter sẽ bổ sung khi có API, không giả backend.

## Convention và cổng chất lượng

- Component PascalCase; biến/hàm camelCase, hook useX; boolean is/has/can. Tên chỉ rõ đơn vị phút/giờ khi dễ nhầm. Không đổi tên hàng loạt chỉ để đẹp.
- Giữ code theo feature; type/helper ở gần nơi dùng, chỉ đưa ra dùng chung khi có nhu cầu thực. Import tương đối trong feature, alias cho đường dẫn xa; tránh barrel re-export hàng loạt. Không thêm tầng rỗng.
- State dẫn xuất tính từ nguồn hiện có; không giữ nhiều bản có thể lệch nhau. Hàm cập nhật state/reducer phải thuần; thời điểm/ID phát sinh tại biên thao tác.
- UI/bình luận tiếng Việt, định danh tiếng Anh; bình luận giải thích lý do/ràng buộc, không chép lại code. Refactor giữ nguyên hành vi và không tự chốt nghiệp vụ còn mở.
- `pnpm lint` và `pnpm format:check` chỉ đọc; `pnpm format` ghi lại định dạng. Không thêm eslint-disable, tăng ngưỡng cảnh báo hoặc loại file nguồn chỉ để cổng xanh; ngoại lệ cần lý do và review.
- Chuẩn công cụ, phạm vi formatter và check theo rủi ro xem README; không áp dụng giới hạn dòng/tham số máy móc thay review trách nhiệm.
