# Phạm vi sản phẩm

Nguồn: `docs/thesis-context/tex/2_Khao_sat.tex` (Chương 2) và `4_Ket_qua_thuc_nghiem.tex` (mục 4.2).

## 1. Bài toán

Trong các nhóm phát triển phần mềm vừa và nhỏ, thông tin năng lực nằm rải rác ở CV, trao đổi nội
bộ và trí nhớ của Quản lý dự án (mục 2.1.2). Hệ quả: phân công dựa vào kinh nghiệm cá nhân, dễ bỏ
sót người phù hợp, và dữ liệu kỹ năng tự khai thiếu kiểm chứng.

Hệ thống giải quyết đúng khoảng trống đó: gom ba nguồn dữ liệu — **hồ sơ năng lực đã được duyệt**,
**yêu cầu kỹ năng chi tiết của từng công việc**, **dữ liệu thực hiện thực tế** — để hỗ trợ PM ra
quyết định phân công.

Hệ thống **không** nhằm thay thế Jira hay Microsoft Project. Không làm tối ưu lịch biểu, không cân
bằng tài nguyên toàn doanh nghiệp, không dự báo năng lực dài hạn (mục 2.1.4).

## 2. Tác nhân

| Tác nhân | Vai trò trong hệ thống |
|---|---|
| Quản trị viên (ADMIN) | Quản lý tài khoản và phân quyền; quản lý danh mục nền; quản lý hồ sơ nhân sự; **duyệt dữ liệu năng lực và kết quả trích xuất CV** |
| Quản lý dự án (PM) | Khởi tạo dự án, phân rã công việc, khai báo yêu cầu kỹ năng, tra cứu nhân sự, phân công, theo dõi tiến độ, đóng dự án |
| Nhân viên (EMPLOYEE) | Hồ sơ cá nhân, gửi yêu cầu cập nhật, xem và cập nhật công việc được giao |

## 3. Nhóm chức năng (mục 2.2)

1. Đăng nhập và quản lý tài khoản — tạo, cập nhật, khoá, mở, phân quyền.
2. Quản lý hồ sơ nhân sự — thêm/sửa/vô hiệu hoá hồ sơ, xem hồ sơ năng lực, cập nhật kỹ năng và CV.
3. Quản lý hồ sơ cá nhân — nhân viên gửi **yêu cầu** cập nhật, không sửa trực tiếp hồ sơ chính thức.
4. Quản lý dự án — khởi tạo, phân rã công việc, xác định yêu cầu kỹ năng, phân công, theo dõi, đóng dự án.
5. Tra cứu hồ sơ nhân sự — thông tin nhân sự, CV, hồ sơ năng lực, lịch sử tham gia dự án.
6. Phân công công việc — thủ công hoặc bằng AI đề xuất, luôn kết thúc bằng bước PM xác nhận.
7. Quản lý công việc được giao — xem, cập nhật trạng thái, ghi nhận thời gian, trao đổi.

## 4. Use case được đặc tả chi tiết

| Mã | Tên | Tác nhân | Tài liệu |
|---|---|---|---|
| UC001 | Khởi tạo dự án | PM | mục 2.3.1 |
| UC002 | Phân rã dự án thành công việc | PM | mục 2.3.2 |
| UC003 | Phân công công việc | PM | mục 2.3.3 |
| UC003.1 | Phân công bằng AI đề xuất | PM | mục 2.3.4 |
| UC003.2 | Phân công thủ công | PM | mục 2.3.5 |
| UC004 | Ghi nhận thời gian thực hiện | Nhân viên | mục 2.3.6 |

Các chức năng hỗ trợ (đăng nhập, xem, tra cứu đơn giản) cố ý không đặc tả riêng trong tài liệu —
khi implement thì bám theo yêu cầu phi chức năng mục 2.4.

## 5. Cố ý nằm ngoài phạm vi

| Hạng mục | Lý do |
|---|---|
| Quản lý **chứng chỉ** (`Certificates`) | Xuất hiện trong biểu đồ 2.2.3/2.2.4/2.2.6 nhưng không thuộc UC001–UC004. Cần **xoá khỏi báo cáo** cho khớp sản phẩm |
| Bảng riêng cho yêu cầu cập nhật kỹ năng tự khai | Dùng chính `employee_skills` với `approved_at IS NULL` để biểu diễn trạng thái chờ duyệt |
| Lịch làm việc chuẩn lưu trong DB | UC004 alt 9a chỉ đòi "có cấu hình". Đặt ở `config/workcalendar.yaml` |
| `ProjectMembers.project_role` | Có trong `Hình 4.5`, không có trong `Hình 4.6`, không use case nào dùng |
| **Phân rã công việc nhiều tầng** (`parent_task_id`, `task_type`) | UC002 chỉ mô tả nhóm công việc chứa công việc. Cấu trúc chốt là đúng hai tầng `task_groups → tasks` |

## 6. Lộ trình

| Mốc | Nội dung | Tiêu chí xong |
|---|---|---|
| **M0** ✅ | Skeleton Go, kết nối Neon, migration 22 bảng + 15 enum, config, logger, seed | `make migrate-up` + `make seed` + `/api/healthz` trả 200 |
| **M1** | Xác thực, middleware RBAC + ABAC, CRUD tài khoản, khoá/mở | Đăng nhập 3 vai trò; gọi API sai quyền trả 403 |
| **M2** | Danh mục nền, hồ sơ nhân sự, hồ sơ năng lực, **luồng duyệt kỹ năng** | ADMIN duyệt xong `employee_skills` mới có `approved_at` |
| **M3** | UC001 + UC002 | Tạo được cây công việc, task mặc định `PENDING_ASSIGNMENT` |
| **M4** | UC003 + UC003.2 | Phân công thủ công chạy đủ luồng chính + 15a/15b |
| **M5** | UC003.1 + engine `matching` + màn hình Top-3 | Unit test engine xanh, hiển thị đủ 7 trường đầu ra |
| **M6** | UC004, `task_status_history`, tính giờ theo `workcalendar.yaml` | Đổi trạng thái sinh đúng số giờ, loại trừ nghỉ trưa và ngoài giờ |
| **M7** | Trích xuất CV → duyệt → vào hồ sơ chính thức | Upload 1 CV thật, duyệt xong thấy kỹ năng `source=CV` |
| **M8** | Bảng ca kiểm thử, thống kê mã nguồn, đóng gói, viết Chương 3 / 4.1 / 4.5 | Có số liệu điền vào báo cáo |

M0–M6 là MVP. M7 vẫn nằm trong đồ án vì đó là phần học máy thật của đề tài.

## 7. Việc cần sửa phía tài liệu

Những chỗ sản phẩm cố ý lệch khỏi báo cáo, người viết báo cáo phải cập nhật ngược lại — xem
`database_schema.md` mục "Chênh lệch so với tài liệu" để có danh sách đầy đủ.
