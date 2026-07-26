# Luồng dự án → công việc → phân công

Đặc tả rút gọn từ Chương 2 mục 2.3, đủ để implement mà không cần mở lại PDF. Số bước giữ nguyên
đánh số của tài liệu để đối chiếu khi viết báo cáo và khi đặt tên ca kiểm thử.

## UC001 — Khởi tạo dự án (PM)

**Tiền điều kiện**: PM đã đăng nhập, tài khoản `ACTIVE`, mã dự án chưa tồn tại.
**Hậu điều kiện**: dự án được lưu và hiển thị ở màn hình chi tiết.

Luồng chính: PM nhập thông tin (5) → hệ thống **tự gán người đăng nhập làm PM phụ trách và đặt
trạng thái mặc định** (6) → PM xác nhận (7) → kiểm tra hợp lệ (8) → lưu (9) → báo thành công (10).

| Trường | Bắt buộc | Điều kiện |
|---|:---:|---|
| Mã dự án | ✅ | không trống, không trùng |
| Tên dự án | ✅ | không trống |
| Mô tả | — | văn bản |
| Ngày bắt đầu | ✅ | đúng định dạng ngày |
| Ngày kết thúc dự kiến | ✅ | **không nhỏ hơn ngày bắt đầu** |
| Mức độ ưu tiên | ✅ | thuộc `priority_level` |
| Ghi chú | — | văn bản |

Luồng thay thế: **5a** huỷ → không lưu, quay lại danh sách. **9a** thiếu/không hợp lệ → báo lỗi,
nhập bổ sung. **9b** trùng mã dự án → báo trùng, yêu cầu mã khác.

Trạng thái sau khi tạo: `PLANNING`.

## UC002 — Phân rã dự án thành công việc (PM)

**Tiền điều kiện**: dự án tồn tại, thuộc phạm vi quản lý của PM, **không ở trạng thái Đóng hoặc
Hoàn thành**, danh mục kỹ năng đã được thiết lập.

Luồng chính: chọn dự án (3) → chọn chức năng phân rã (5) → nhập nhóm công việc (7) → thêm công việc
thuộc nhóm (8) → nhập thông tin từng công việc (9) → **gán yêu cầu kỹ năng cho từng công việc**
(10) → xác nhận (11) → kiểm tra hợp lệ (12) → lưu (13) → **đặt trạng thái mặc định `PENDING_ASSIGNMENT`**
(14) → hiển thị cây công việc (15).

| Trường | Bắt buộc | Điều kiện |
|---|:---:|---|
| Tên nhóm công việc | ✅ | không trùng trong cùng dự án |
| Mô tả nhóm | — | văn bản |
| Mã công việc | ✅ | không trống, không trùng trong dự án |
| Tên công việc | ✅ | không trống |
| Ngày bắt đầu / kết thúc dự kiến | — | nếu nhập kết thúc thì không nhỏ hơn bắt đầu |
| Thời gian ước tính | ✅ | **số lớn hơn 0** |
| Yêu cầu kỹ năng | ✅ | kỹ năng tồn tại; **mỗi công việc ít nhất một yêu cầu** |

Luồng thay thế — đây chính là danh sách validate phải implement:

| Mã | Điều kiện | Xử lý |
|---|---|---|
| 5a | Dự án đang Đóng hoặc Hoàn thành | Báo không cho phân rã, kết thúc use case |
| 13a | Thiếu thông tin bắt buộc của công việc | Báo lỗi, yêu cầu bổ sung |
| 13b | Công việc chưa có yêu cầu kỹ năng | Yêu cầu gán ít nhất một kỹ năng |
| 13c | Ngày kết thúc < ngày bắt đầu | Yêu cầu chỉnh sửa |
| 13d | Kỹ năng không tồn tại hoặc đang `INACTIVE` | Yêu cầu chọn kỹ năng hợp lệ |
| 13e | Thời gian ước tính không hợp lệ | Yêu cầu nhập lại |

Nhớ **R2**: `task_groups` chỉ gom nhóm. Giờ ước tính và yêu cầu kỹ năng luôn thuộc về `tasks`.

## UC003 — Phân công công việc (PM)

**Tiền điều kiện**: dự án thuộc phạm vi quản lý, chưa Đóng/Hoàn thành, có ít nhất một công việc
`PENDING_ASSIGNMENT`, công việc đủ điều kiện đã có thời gian ước tính và yêu cầu kỹ năng.

Luồng chính:

1–5. PM chọn dự án và chức năng Phân công công việc.
6. Hệ thống **lọc và hiển thị công việc đủ điều kiện phân công**.
7. Hệ thống **tự chọn sẵn** các công việc `PENDING_ASSIGNMENT`.
8. PM kiểm tra và điều chỉnh danh sách.
9. PM chọn phương thức: **AI đề xuất (UC003.1)** hoặc **thủ công (UC003.2)**.
10. Hệ thống xử lý theo nhánh đã chọn.
11. PM kiểm tra **bảng phân công dự kiến** và xác nhận.
12. Hệ thống kiểm tra lại từng dòng: trạng thái công việc, trạng thái nhân sự, quyền phân công,
    khả năng tham gia dự án.
13. Lưu phân công cho các dòng hợp lệ.
14. Thêm nhân sự vào nhóm dự án nếu chưa thuộc, `join_source = FROM_ASSIGNMENT`.
15. Gửi thông báo cho nhân sự được phân công.
16–17. Báo thành công và hiển thị lại danh sách kèm người phụ trách.

Bước 12–15 phải nằm trong **một transaction** (xem `domain_rules.md` mục 6).

Luồng thay thế:

| Mã | Điều kiện | Xử lý |
|---|---|---|
| 5a | Dự án Đóng/Hoàn thành | Báo không cho phân công, kết thúc |
| 8a | Không có công việc `PENDING_ASSIGNMENT` | Báo không có công việc đủ điều kiện, kết thúc |
| 10a | PM huỷ ở bước 9 | Không lưu, quay lại chi tiết dự án |
| 11a | PM không chấp nhận kết quả AI | Cho chuyển sang UC003.2 |
| 15a | **Một vài** dòng không hợp lệ | Chỉ rõ dòng lỗi và lý do, quay lại bước 12 |
| 15b | **Toàn bộ** dòng không hợp lệ | Không lưu gì, yêu cầu phân công lại, quay lại bước 7 |
| 16a | Lỗi khi lưu | Báo thất bại, **giữ nguyên dữ liệu trước đó** (rollback) |

Trạng thái công việc sau phân công: `ASSIGNED`.

## UC003.2 — Phân công thủ công (PM)

Dùng khi PM muốn tự chọn, hoặc sau khi không chấp nhận kết quả AI.

Luồng chính: hiển thị công việc và yêu cầu kỹ năng (2) → hiển thị danh sách nhân sự (3) → PM nhập
điều kiện lọc (4) → lọc (5) → hiển thị kèm hồ sơ năng lực, kỹ năng, kinh nghiệm, **tải công việc
hiện tại** và trạng thái tham gia dự án (6) → PM xem chi tiết nếu cần (7) → chọn nhân sự (8) →
hệ thống **kiểm tra sơ bộ kỹ năng bắt buộc và trạng thái khả dụng** (9) → **hiển thị cảnh báo nếu
không phù hợp** (10) → PM xác nhận và nhập ghi chú (11) → ghi vào bảng phân công dự kiến (12) →
chuyển về UC003 (13).

Bộ lọc (bước 4): **kỹ năng, vị trí, cấp độ kinh nghiệm, phòng ban, trạng thái làm việc, tình trạng
tham gia dự án**. Tất cả đều tuỳ chọn → cần dựng truy vấn động.

Điểm khác biệt với UC003.1: người thiếu kỹ năng bắt buộc **vẫn hiện ra, kèm cảnh báo** (bước 10),
không bị loại như ở nhánh AI.

Luồng thay thế:

| Mã | Điều kiện | Xử lý |
|---|---|---|
| 6a | Không có nhân sự khớp bộ lọc | Báo không tìm thấy, cho đổi điều kiện lọc |
| 9a | Chưa chọn nhân sự cho công việc | Yêu cầu chọn trước khi đưa vào bảng dự kiến |
| 10a | Nhân sự đã ngừng hoạt động hoặc bị khoá | **Không cho chọn**, yêu cầu chọn người khác |
| 12a | PM huỷ | Không đưa vào bảng dự kiến, quay lại danh sách công việc |

## UC004 — Ghi nhận thời gian thực hiện (Nhân viên)

**Tiền điều kiện**: nhân viên đang được phân công công việc, công việc ở `IN_PROGRESS`, hệ thống đã
có thời điểm bắt đầu và **cấu hình lịch làm việc chuẩn**.

Luồng chính: nhân viên chọn công việc (3) → xem chi tiết (4) → **chuyển trạng thái từ `IN_PROGRESS`
sang `PENDING_REVIEW`** (5) → hệ thống ghi nhận thời điểm (6) → xác định khoảng cần tính (7) →
tính theo lịch làm việc chuẩn, mặc định 08:00–12:00 và 13:00–17:00 (8) → **loại trừ ngoài giờ,
nghỉ trưa và các khoảng không tính** (9) → lưu `work_logs` (10) → cập nhật `tasks.actual_hours`
(11) → đối chiếu với `estimate_hours` (12) → cập nhật dữ liệu phục vụ đánh giá hiệu suất (13) →
báo thành công (14).

Luồng thay thế:

| Mã | Điều kiện | Xử lý |
|---|---|---|
| 6a | Chuyển trạng thái không hợp lệ theo quy trình | Báo lỗi, **không** cập nhật trạng thái |
| 6b | Nhân viên huỷ | Không đổi trạng thái, không ghi thời gian |
| 9a | **Chưa có cấu hình lịch làm việc** | Báo không thể tính, **không lưu bản ghi thời gian** |
| 11a | Lỗi khi lưu | Báo thất bại, giữ nguyên dữ liệu trước đó |

Đầu ra cần hiển thị: thời điểm bắt đầu, thời điểm kết thúc, **số giờ được tính**, tổng giờ thực tế,
và **chênh lệch so với ước tính** (ví dụ "Ít hơn 3.5 giờ so với ước tính").

Số giờ do hệ thống tính, nhân viên **không nhập tay**. Bản ghi ghi `work_logs.source = SYSTEM`.
