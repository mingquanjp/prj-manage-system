# Kiểm thử và nghiệm thu

Phần này phục vụ hai mục đích: giữ chất lượng code, và tạo **bảng ca kiểm thử** để điền vào báo cáo
(mục 4.3 và Chương 5).

## 1. Hiện trạng

- **Chưa có file test nào.** `go test ./...` xanh vì không có gì để chạy.
- Cổng kiểm tra duy nhất: `make check` = `fmt` → `tidy` → `vet` → `build` → `test`.
- Chưa cấu hình `golangci-lint`.

## 2. Ba tầng test

### Unit — `internal/matching`

Trọng tâm, vì đây là phần viết vào mục "Kiểm thử" của báo cáo. Engine thuần hàm, không I/O → dùng
bảng ca (table-driven test). Các ca bắt buộc:

- Kỹ năng **thiếu / đủ / vượt** mức yêu cầu (`s_i` phải cap ở 1.0 khi vượt).
- Kỹ năng **bắt buộc thiếu hoàn toàn** → bị loại khỏi Top-N.
- Kỹ năng **bắt buộc có nhưng chưa đủ mức** → `SkillScore ×= 0.7`, không bị loại.
- Trọng số khác nhau giữa các yêu cầu → kiểm tra đúng công thức trung bình có trọng số.
- Workload biên: `load = 0` → `WorkloadScore = 1`; `load > CAPACITY` → `WorkloadScore = 0`.
- **Không có dữ liệu hiệu suất** → `PerfScore = 0.5` và phần giải thích nói rõ điều đó.
- Nhân sự `INACTIVE` / tài khoản `LOCKED` → bị loại ở hard filter.
- **Thứ tự xếp hạng ổn định**: cùng input, chạy nhiều lần cho ra cùng thứ tự, kể cả khi bằng điểm.
- Câu giải thích không rỗng trong mọi trường hợp.

### Integration — repository

Chạy trên **Neon branch `test`** riêng qua biến `TEST_DATABASE_URL`; schema tạo bằng chính
migration. Không có biến này thì test tự `t.Skip()` để `go test ./...` vẫn xanh khi offline.

Trọng tâm: transaction xác nhận phân công phải **rollback trọn vẹn** khi một dòng lỗi (UC003 alt
16a). Kiểm tra sau rollback: không có `assignments` mới, `tasks.status` giữ nguyên, không có
`project_members` mới, không có `notifications` mới.

### End-to-end — `httptest`

Chạy đúng kịch bản bốn use case đã đặc tả, gồm cả luồng thay thế. Ánh xạ 1-1 với bảng đặc tả
Chương 2 → dùng luôn làm bảng ca kiểm thử trong báo cáo.

| Nhóm | Ca cần phủ |
|---|---|
| UC001 | luồng chính; 9a thiếu dữ liệu; 9b trùng mã dự án |
| UC002 | luồng chính; 5a dự án đã đóng; 13a thiếu trường; 13b chưa có yêu cầu kỹ năng; 13c sai ngày; 13d kỹ năng `INACTIVE`; 13e ước tính ≤ 0 |
| UC003 | luồng chính; 5a dự án đóng; 8a không có task chờ phân công; 15a một vài dòng lỗi; 15b toàn bộ lỗi; 16a lỗi khi lưu |
| UC003.1 | luồng chính Top-3; 3a task chưa có yêu cầu kỹ năng; 10a chuyển sang thủ công; **11a bắt nhập lý do khi chọn khác Top-1**; 12a nhân sự không còn khả dụng |
| UC003.2 | luồng chính; 6a không tìm thấy nhân sự; 10a nhân sự bị khoá; cảnh báo khi thiếu kỹ năng bắt buộc |
| UC004 | luồng chính; 6a chuyển trạng thái sai; 9a chưa có cấu hình lịch làm việc; tính đúng giờ khi khoảng thời gian **bắc qua nghỉ trưa, qua đêm, qua cuối tuần** |
| Phân quyền | mỗi endpoint gọi bằng cả ba vai trò; PM A đụng dự án của PM B → **403** |
| Quy tắc R3 | kỹ năng chưa duyệt **không** ảnh hưởng kết quả engine |

### Thủ công

Seed data + chạy server, đi hết luồng bằng UI với ba tài khoản `admin` / `pm.hoa` / `dev.an` để
chụp màn hình cho mục 4.3.3.

## 3. Tiêu chí nghiệm thu từng mốc

| Mốc | Nghiệm thu |
|---|---|
| M0 ✅ | `make migrate-up` + `make seed` chạy sạch; `/api/healthz` trả 200; `make check` xanh |
| M1 | Đăng nhập được cả ba vai trò; gọi API sai quyền trả **403**; khoá tài khoản có hiệu lực ngay |
| M2 | Kỹ năng chỉ có `approved_at` sau khi ADMIN duyệt; kỹ năng chưa duyệt không xuất hiện ở màn hình tra cứu |
| M3 | Tạo được cây công việc; task mặc định `PENDING_ASSIGNMENT`; đủ bảy validate 13a–13e |
| M4 | Phân công thủ công chạy đủ luồng chính + 15a/15b; kiểm chứng rollback |
| M5 | Unit test engine xanh; Top-3 hiển thị **đủ bảy trường đầu ra**; alt 11a bắt nhập lý do |
| M6 | Đổi trạng thái sinh đúng số giờ; loại trừ nghỉ trưa và ngoài giờ; alt 9a không lưu gì |
| M7 | Upload một CV thật; trước khi duyệt engine cho kết quả không đổi; duyệt xong kỹ năng vào hồ sơ với `source=CV` |
| M8 | Có bảng ca kiểm thử và số liệu thống kê mã nguồn để điền vào báo cáo |

## 4. Số liệu cần thu thập cho báo cáo

- Thời gian phản hồi các thao tác thường xuyên (mục 2.4.3). **Phải nêu rõ Neon autosuspend sau
  ~5 phút và cold start ~0.5–1s** khi bình luận số liệu — nếu không, con số đo được sẽ khó giải thích.
- Thời gian sinh danh sách đề xuất theo số lượng nhân sự.
- **Tỉ lệ PM chấp nhận Top-1** — tính từ `recommendation_items.is_selected` kết hợp `rank_no`.
  Đây là dữ liệu đánh giá engine ở Chương 5, và là lý do bảng `recommendation_items` tồn tại.
- Số bảng, số dòng mã, số ca kiểm thử đã chạy.
