# Gợi ý nhân sự bằng luật và công thức — UC003.1

**Đồng bộ quyết định 15/10/2026:** [sổ quyết định, phần hiện hành và Bàn giao](../../docs/design/database-review-decisions-2026-10-14.md) ưu tiên cho nội dung đã thay thế; [dictionary/DBML hợp nhất](../../docs/design/database-proposal-2026-10-14.md) là thiết kế 32 bảng. Chưa triển khai code/DDL/migration.


> Baseline nghiệp vụ hiện hành: [2026-10-12](../../docs/business-baseline-2026-10-12.md), PTCM phân công và nghiệm thu, không nhận task trong dự án phụ trách; tra cứu giới hạn dữ liệu nghề nghiệp. Phần kỹ thuật V2 chưa đồng bộ không được ghi đè quyết định mới; công thức và quy tắc không bị thay thế giữ nguyên.

Thiết kế đích phát triển từ [V2 lịch sử](../../docs/reviews/system-design-baseline-v2.md); quyền/điều kiện theo baseline 2026-10-12 ở trên. Không bắt buộc ML. Thiết kế assignment_type dùng RECOMMENDED/MANUAL theo quyết định ngày 2026-09-14; tên file giữ nguyên, mã AI trong implementation cũ chưa được chuyển đổi.
Đây là mô hình chấm điểm heuristic giải thích được, chưa được chứng minh tối ưu.
Engine `internal/matching` thuần dữ liệu, không HTTP/DB.

## 1. Luồng và điều kiện

Theo [bổ sung 2026-10-13](../../docs/business-baseline-2026-10-13.md), không yêu cầu ứng viên là thành viên dự án có sẵn; không thêm luồng tìm ngoài đội/duyệt điều động. Tải xét xuyên dự án. Số dự án đang có việc được hiển thị tham khảo theo định nghĩa baseline, không phải tiêu chí điểm mới hoặc giới hạn hai dự án. Không thay công thức/snapshot điểm trong đợt này. Khi hiển thị số liệu hiện hành cùng run cũ phải phân biệt với dữ liệu tại thời điểm run; không sửa run cũ.

PTCM chọn task PENDING_ASSIGNMENT cùng project mở, có estimate và requirements hợp lệ.
Giao lại bằng gợi ý phải thu hồi trước: A đang phụ trách → thu hồi có hiệu lực → PENDING_ASSIGNMENT → run mới → chọn người → confirm → ASSIGNED. Sau thu hồi A không còn phụ trách; không có ứng viên hoặc chưa confirm thành công thì task vẫn PENDING_ASSIGNMENT, không tự khôi phục A. Giữ lịch sử/đóng góp của A và frozen spec; không hỗ trợ xem gợi ý người thay khi A còn giữ task.
UI có “Chọn tất cả công việc đủ điều kiện” trong phạm vi lọc hiện hành, gồm cả các trang chưa hiển thị; nêu rõ phạm vi và số lượng. Ban đầu chưa chọn, PTCM có thể bỏ chọn hoặc chọn một tập nhỏ hơn. Task không đủ điều kiện hiển thị lý do và không được chọn; tập rỗng không chạy gợi ý.
PTCM chạy gợi ý một lần cho toàn bộ tập đã chọn, nhận bảng Top-N theo từng task, không phải mở từng task để chạy. PTCM chọn ứng viên từng dòng hoặc chuyển riêng dòng đó sang manual; không tự chọn/commit Top-1. Run rỗng ở một task không làm mất kết quả task khác; PTCM xử lý manual hoặc loại task đó khỏi draft trước confirm nguyên tử.
Lấy snapshot nhất quán → lọc ứng viên → chấm điểm → Top-3 → PTCM chọn vào draft → UC003 xác nhận.
Mỗi task có một run; batch xếp hạng độc lập cùng snapshot, không tối ưu phân công đồng thời.
Draft hiển thị tổng tải sau các lựa chọn trước confirm, không tăng tải chính thức.

Candidate pool toàn tổ chức: người có thể thực hiện task, user/employee ACTIVE, không là PTCM của dự án đang phân công (kể cả PM kiêm PTCM); chỉ kỹ năng chính thức còn hiệu lực. PTCM ở dự án khác vẫn có thể là ứng viên. Điều kiện loại PTCM áp dụng trước chấm điểm, không phải điểm phạt và không được override ở nhánh manual.
Danh mục kỹ năng không có trạng thái INACTIVE. Requirement đã giao giữ nguyên; đối chiếu với official hiện đang được công nhận. REMOVE official khiến kỹ năng không còn đáp ứng, không xóa requirement hoặc sửa snapshot/lịch sử. Ngoại lệ thủ công về kỹ năng vẫn cần lý do, không vượt điều kiện quyền/trạng thái.
Thiếu hoặc dưới mức một mandatory → loại, không phạt thêm 0,7.
Manual có thể override kỹ năng với lý do, không override trạng thái/vai trò.
Không ứng viên hợp lệ → Top-N rỗng, không tự hạ điều kiện.
Top-N mặc định N=3; có thể0–2 nếu ít người đủ điều kiện. Mỗi ứng viên có rank, điểm tổng/thành phần và giải thích, không chỉ một danh sách tên.
Manual là tra cứu/lọc riêng: có thông tin năng lực/tải/tổng hợp lịch sử nhưng không tạo run/rank/matching_score khi chưa chạy engine.
Công thức trọng số dưới đây là phương án thử nghiệm cũ được lưu để nghiên cứu, chưa chốt cuối sau thay đổi nghiệp vụ. Giữ định hướng PerfScore, chưa đi sâu công thức ở đợt này; không phải cosine similarity. Không tự thay sang cosine/ML chỉ từ nhãn AI; phần2 sẽ đặc tả và kiểm chứng phương pháp.

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
r_j = min(estimate_hours_j / actual_hours_j, 1)
PerfScore = mean(r_j) nếu có ít nhất 3 mẫu hợp lệ (tối đa 10)

Score = 0.7 × SkillScore + 0.2 × WorkloadScore + 0.1 × PerfScore
matching_score = round_half_up(100 × Score, 2)
```

Lọc mẫu hợp lệ trước, rồi lấy tối đa 10 task DONE/CLOSED gần nhất theo accepted_at giảm dần (hòa theo task_id tăng dần).
Mẫu phải có actual >0 từ nhật ký hiệu lực, estimate/scope đóng băng, chỉ một người có đóng góp thực hiện và effort được xác nhận đầy đủ/tin cậy; không có yêu cầu hiệu chỉnh PENDING.
Task có đóng góp đáng kể của người khác chưa được ghi đủ hoặc thiếu effort bị loại; không suy một current_assignee đồng nghĩa một người thực hiện.
Số 0,5 khi dưới 3 mẫu là mặc định thử nghiệm được giữ từ V1, không phải trung bình đo được hoặc bảo đảm công bằng; UI ghi 'Chưa đủ dữ liệu', hiển thị số mẫu.
Không thưởng vượt 1 khi làm dưới estimate; trung bình tỷ lệ từng task, không chia tổng estimate cho tổng actual.
Sửa effort/đánh dấu độ tin cậy ảnh hưởng run tương lai, không viết lại snapshot/xếp hạng quá khứ. Khi có yêu cầu sửa đang chờ, tạm loại task khỏi tập mẫu.
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
Lưu engine_version, params, snapshot_at, giá trị đặc tả/requirements; input snapshot toàn tập đã xét:
trạng thái/quyền và trách nhiệm PTCM theo dự án, skill/level và sự có/không có kỹ năng liên quan, workload, các mẫu lịch sử thực hiện và tình trạng chất lượng dữ liệu, lý do loại.
Output bất biến gồm điểm thành phần, rank, matched/missing và explanation.
Không sao chép CV thô/liên hệ vào snapshot. Params riêng không đủ replay.
R06 snapshot nội bộ đủ replay không phải DTO trả PTCM/ADMIN. GET run chỉ trả Top-N, điểm/giải thích, tổng tải và tổng hợp mẫu được phép xem; không lộ task_id/mô tả/log/minh chứng dự án khác qua giải thích hoặc thông báo lỗi.
PTCM xem chi tiết task/effort trong project mình bằng API có ABAC riêng. ADMIN không mặc định được đọc mọi run/raw snapshot. Dữ liệu quyền xem không nhân đôi official skills.

Khi confirm: người thao tác vẫn là PTCM của dự án; task vẫn chờ giao, spec còn đúng, người được chọn còn active/đúng quyền, không là PTCM của dự án và giá trị kỹ năng liên quan chưa đổi so snapshot, gồm cả kỹ năng vắng mặt lúc chạy. Kiểm tra phối hợp với thao tác thay PTCM, không tin quyền hoặc eligibility từ run cũ.
Skill/spec đổi → chạy lại hoặc manual có lý do; tải đổi → hiển thị tải mới, giữ nhãn thời điểm ranking.
R05: gợi ý phản ánh dữ liệu tại thời điểm tạo, hiển thị snapshot_at; PTCM có thể chạy lại để lấy kết quả mới.
Lịch sử dùng tính điểm thay đổi không yêu cầu cảnh báo/xác nhận phiên bản hoặc lý do dùng kết quả cũ. Không dùng history_revision/acknowledgment làm điều kiện commit riêng cho R05.
Giữ nguyên toàn bộ điểm/rank snapshot; không ghép điểm cũ và mới. Assignment giữ liên kết run/item đã dùng, điểm đọc từ item để truy vết, không gọi đó là xếp hạng thời gian thực.
Điều kiện phân công hiện tại, kiểm tra skill/spec và giới hạn quyền vẫn giữ như trên. Các lý do chọn khác hạng đầu, chuyển manual hoặc override kỹ năng độc lập với R05 vẫn giữ.
Cùng một run chỉ có tối đa một quyết định commit; run cũ sau thu hồi không được tái sử dụng nếu đã có quyết định.

| Quyết định | assignment_type | Truy vết |
|---|---|---|
| Top-1 | RECOMMENDED | run_id, item_id; score đọc từ item |
| Hạng 2/3 | RECOMMENDED | như trên + adjust_reason |
| Manual sau run | MANUAL | run_id, item_id NULL, không cột score, adjust_reason |
| Manual từ đầu | MANUAL | run/item NULL, không cột score; lý do khi override |

Không dùng is_selected như nguồn ghi độc lập; lựa chọn suy từ assignment đã commit.
Server kiểm tra item thuộc đúng run/task/employee, không nhận điểm/rank do client quyết định.

## 4. Vì sao phù hợp và phải đánh giá gì

Kỹ năng là thành phần chính; tải là yếu tố điều chỉnh; dữ liệu thời gian ít tin cậy chỉ có trọng số nhỏ.
Mandatory là điều kiện cứng nên điểm tải/hiệu suất không bù được thiếu bắt buộc.
Đây là tính hợp lý về cấu trúc, chưa chứng minh bộ số 0,7/0,2/0,1, 80, 10, 3 tốt nhất.

Phải so sánh skill-only, skill+workload và đủ ba thành phần trên bộ tình huống được người đánh giá độc lập gán mức phù hợp.
Khảo sát độ nhạy trọng số/ngưỡng và cách mã hóa proficiency; tách tập hiệu chỉnh với tập đánh giá.
Kiểm tra các ca rất ít dữ liệu, mọi người tải ≥80, mọi kỹ năng đều mandatory (SkillScore có thể bằng nhau).
PerfScore bị ảnh hưởng estimate/độ khó/effort tự khai thiếu hoặc sai, đóng góp hỗ trợ chưa ghi nhận; không dùng để kết luận xếp hạng năng lực nhân sự.

Ví dụ kiểm tra đặc tả: estimate=8h, log=10h → r=0,8; log=6h hoặc 8h → r=1.
Ba mẫu r=(1; 1; 0,5) → PerfScore=0,833333…, không lấy tỷ số của các tổng.
Trọng số, 80h, tối thiểu 3/tối đa 10 và mặc định 0,5 là tham số thử nghiệm, chưa có bằng chứng tối ưu. Phần đặc tả/thực nghiệm sâu tiếp theo phải đánh giá cả cold-start, không ngầm coi đã kiểm chứng.

Tỷ lệ chấp nhận Top-1/Top-3: mẫu số là quyết định commit có run không rỗng, gồm cả chuyển manual; tách giao đầu và giao lại.
Run bỏ dở/chạy lại không tính là từ chối; báo riêng run rỗng.
Đo chất lượng Top-K theo nhãn độc lập, thời gian chạy và độ ổn định. Unit test đúng công thức không chứng minh chất lượng gợi ý.

## Contract lưu trữ sau rà soát 15/10

Hai bảng run/items giữ cấu trúc đã chốt; run.schema_version định nghĩa params/input_snapshot/result_metadata/score_details. Không spec_version hoặc official skill.version/is_effective. Khi confirm, so giá trị đặc tả (description gồm đầu ra, criteria, estimate, ngày theo scope kiểm tra), tập requirements chuẩn hóa theo skill_id và kỹ năng liên quan (kể cả vắng mặt) dưới khóa; thay đổi thì chạy lại hoặc chuyển manual theo quy tắc. Không dùng updated_at thay kiểm tra. Công thức thử nghiệm ở mục 2 không phải công thức cuối hoặc quyết định áp CHECK weight 1–5.
