# Gợi ý nhân sự bằng luật và công thức — UC003.1

Thiết kế đích V1 đã đồng ý 2026-09-05. Không bắt buộc ML; tên file/enum AI giữ tương thích kỹ thuật.
Đây là mô hình chấm điểm heuristic giải thích được, chưa được chứng minh tối ưu.
Engine `internal/matching` thuần dữ liệu, không HTTP/DB.

## 1. Luồng và điều kiện

PM chọn task PENDING_ASSIGNMENT cùng project mở, có estimate và requirements hợp lệ.
Lấy snapshot nhất quán → lọc ứng viên → chấm điểm → Top-3 → PM chọn vào draft → UC003 xác nhận.
Mỗi task có một run; batch xếp hạng độc lập cùng snapshot, không tối ưu phân công đồng thời.
Draft hiển thị tổng tải sau các lựa chọn trước confirm, không tăng tải chính thức.

Candidate pool toàn tổ chức: role EMPLOYEE, user/employee ACTIVE; chỉ kỹ năng chính thức còn hiệu lực.
Thiếu hoặc dưới mức một mandatory → loại, không phạt thêm 0,7.
Manual có thể override kỹ năng với lý do, không override trạng thái/vai trò.
Không ứng viên hợp lệ → Top-N rỗng, không tự hạ điều kiện.

## 2. Công thức phiên bản thử nghiệm

Mã hóa BASIC=1, INTERMEDIATE=2, ADVANCED=3, EXPERT=4.
Đây là mã hóa thang thứ bậc, không khẳng định tỷ lệ năng lực thực tế.

```text
s_i = min(employee_level_i / required_level_i, 1)  # thiếu optional: 0
SkillScore = Σ(weight_i × s_i) / Σweight_i         # weight nguyên 1–5

open_load = Σ estimate_hours của task hiện được giao
            có trạng thái ASSIGNED/IN_PROGRESS/PAUSED/PENDING_REVIEW
WorkloadScore = clamp(1 - open_load / 80, 0, 1)

PerfScore = 0.5 nếu < 3 mẫu hợp lệ
PerfScore = clamp(Σestimate_hours / Σactual_hours, 0, 1.2) / 1.2 nếu đủ mẫu

Score = 0.7 × SkillScore + 0.2 × WorkloadScore + 0.1 × PerfScore
matching_score = round_half_up(100 × Score, 2)
```

Lấy tối đa 10 task DONE/CLOSED gần nhất theo accepted_at (hòa theo task_id), actual > 0, chỉ một người thực hiện, estimate không đổi.
Task nhiều người đóng góp không làm mẫu cá nhân. Dưới 3 mẫu giải thích rõ thiếu dữ liệu.
80h là ngưỡng chuẩn hóa tải tồn đọng thử nghiệm; không đại diện lịch trống hai tuần.
Không dùng estimate−actual làm remaining effort. PENDING_REVIEW vẫn giữ tải vì có thể bị trả lại, hiển thị riêng.
Không bonus project member. Priority/deadline dùng hiển thị, không tuyên bố công thức bảo đảm deadline.

Điểm nội bộ 0–1; API/DB tất cả điểm thành phần và tổng 0–100, hai chữ số.
Phá hòa theo tổng đã làm tròn giảm dần → open_load tăng dần → employee_id tăng dần.
Version hóa công thức, trọng số, ngưỡng, làm tròn, tie-break; alpha/beta/gamma không âm và tổng bằng 1, capacity > 0.
Không làm tròn trung gian khi tính tổng.

## 3. Kết quả và snapshot

Hiển thị tên task, Top-N, điểm, kỹ năng đáp ứng, kỹ năng thiếu/dưới mức, tải (số task + giờ), giải thích.
Run rỗng có lý do; giải thích không được khẳng định “đủ lịch” hoặc “năng suất thực tế”.
Lưu engine_version, params, snapshot_at, task spec version; input snapshot toàn tập đã xét:
trạng thái/role, skill/level/version, workload, các mẫu hiệu suất, lý do loại.
Output bất biến gồm điểm thành phần, rank, matched/missing và explanation.
Không sao chép CV thô/liên hệ vào snapshot. Params riêng không đủ replay.

Khi confirm: task vẫn chờ giao, spec còn đúng, người được chọn còn active/đúng role và skill version chưa đổi.
Skill/spec đổi → chạy lại hoặc manual có lý do; tải đổi → hiển thị tải mới, giữ nhãn thời điểm ranking.
Cùng một run chỉ có tối đa một quyết định commit; run cũ sau thu hồi không được tái sử dụng nếu đã có quyết định.

| Quyết định | assignment_type | Truy vết |
|---|---|---|
| Top-1 | AI | run_id, item_id, score từ server |
| Hạng 2/3 | AI | như trên + adjust_reason |
| Manual sau run | MANUAL | run_id, item_id/score NULL, adjust_reason |
| Manual từ đầu | MANUAL | run/item/score NULL; lý do khi override |

Không dùng is_selected như nguồn ghi độc lập; lựa chọn suy từ assignment đã commit.
Server kiểm tra item thuộc đúng run/task/employee, không nhận điểm/rank do client quyết định.

## 4. Vì sao phù hợp và phải đánh giá gì

Kỹ năng là thành phần chính; tải là yếu tố điều chỉnh; dữ liệu thời gian ít tin cậy chỉ có trọng số nhỏ.
Mandatory là điều kiện cứng nên điểm tải/hiệu suất không bù được thiếu bắt buộc.
Đây là tính hợp lý về cấu trúc, chưa chứng minh bộ số 0,7/0,2/0,1, 80, 10, 3 tốt nhất.

Phải so sánh skill-only, skill+workload và đủ ba thành phần trên bộ tình huống được người đánh giá độc lập gán mức phù hợp.
Khảo sát độ nhạy trọng số/ngưỡng và cách mã hóa proficiency; tách tập hiệu chỉnh với tập đánh giá.
Kiểm tra các ca rất ít dữ liệu, mọi người tải ≥80, mọi kỹ năng đều mandatory (SkillScore có thể bằng nhau).
PerfScore bị ảnh hưởng estimate/độ khó/quên pause; không dùng để kết luận xếp hạng năng lực nhân sự.

Tỷ lệ chấp nhận Top-1/Top-3: mẫu số là quyết định commit có run không rỗng, gồm cả chuyển manual; tách giao đầu và giao lại.
Run bỏ dở/chạy lại không tính là từ chối; báo riêng run rỗng.
Đo chất lượng Top-K theo nhãn độc lập, thời gian chạy và độ ổn định. Unit test đúng công thức không chứng minh chất lượng gợi ý.
