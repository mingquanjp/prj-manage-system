# Luồng AI đề xuất nhân sự (UC003.1)

Đây là chức năng trung tâm của đề tài. Đọc kèm `domain_rules.md` mục 1 (R1, R4).

## 1. Định vị "AI" của đề tài

Engine là **rule-based / knowledge-based recommender**: công thức tường minh, giải thích được,
tái lập được. Cố ý **không** dùng LLM để xếp hạng, vì cần kết quả ổn định để đánh giá và cần nội
dung định lượng cho Chương 5.

Đây là lựa chọn có cơ sở học thuật: knowledge-based recommender không cần dữ liệu huấn luyện, phù
hợp với bài toán cold-start của một tổ chức mới triển khai. Tài liệu đã trích dẫn
`shafiq2021taskallocator` và `lamersdorf2013survey` cho hướng này (mục 2.1.3).

Thành phần học máy / NLP thật của đề tài nằm ở **trích xuất kỹ năng từ CV** (M7) — xem
`cv_extraction_approval_workflow.md`.

Yêu cầu kiến trúc: engine đặt tại `internal/matching`, **không biết DB, không biết HTTP**. Đây là
đòi hỏi trực tiếp của mục 2.4.6.

## 2. Luồng use case

**Tiền điều kiện**: công việc đã có yêu cầu kỹ năng hợp lệ; hồ sơ năng lực chính thức đã được xác
nhận; có nhân sự đang hoạt động.

1. PM chọn phương thức AI đề xuất cho danh sách công việc đã chọn.
2. Hệ thống lấy chi tiết từng công việc: **thời gian ước tính, mức độ ưu tiên, yêu cầu kỹ năng**.
3. Lấy **hồ sơ năng lực chính thức** của nhân sự đang hoạt động.
4. Lấy tình trạng tham gia dự án, **khối lượng công việc hiện tại**, **lịch sử hiệu suất** nếu có.
5. Đánh giá mức độ phù hợp.
6. Tính điểm phù hợp cho từng cặp (nhân sự, công việc).
7. Xếp hạng và chọn **Top N**, minh hoạ bằng **Top 3**.
8. Hiển thị kết quả.
9. PM xem kết quả.
10. PM chấp nhận Top-1 hoặc chọn người khác trong danh sách.
11. Ghi nhận vào **bảng phân công dự kiến**.
12. Chuyển về UC003 để PM xác nhận.

Luồng thay thế:

| Mã | Điều kiện | Xử lý |
|---|---|---|
| 3a | Công việc chưa có yêu cầu kỹ năng | Báo chưa đủ điều kiện, **loại công việc đó khỏi danh sách đề xuất** |
| 10a | PM không chấp nhận kết quả AI | Cho chuyển sang UC003.2 |
| 11a | PM chọn người **khác Top-1** | **Bắt buộc nhập lý do điều chỉnh** |
| 12a | Nhân sự được chọn không còn khả dụng | Báo lỗi, yêu cầu chọn lại, quay lại bước 8 |

Alt 11a quan trọng gấp đôi: vừa là yêu cầu nghiệp vụ, vừa là nguồn dữ liệu đánh giá engine ở
Chương 5 (tỉ lệ PM chấp nhận Top-1, và lý do khi không chấp nhận).

## 3. Bảy trường đầu ra bắt buộc

Màn hình kết quả phải hiển thị đủ, không thiếu trường nào:

| # | Trường | Ví dụ trong tài liệu |
|---|---|---|
| 1 | Tên công việc | Thiết kế cơ sở dữ liệu |
| 2 | Danh sách Top N nhân sự | Top 1: Trần Văn B |
| 3 | Điểm phù hợp | 0.86/1 |
| 4 | Kỹ năng đáp ứng | SQL – Khá |
| 5 | Kỹ năng còn thiếu | Docker – Cơ bản |
| 6 | Tải công việc hiện tại | Đang phụ trách 2 công việc |
| 7 | Giải thích đề xuất | Phù hợp kỹ năng SQL và tải công việc hiện tại thấp |

Bảy trường này ánh xạ trực tiếp vào các cột của `recommendation_items`.

## 4. Công thức chấm điểm

Quy đổi mức độ: `BASIC=1, INTERMEDIATE=2, ADVANCED=3, EXPERT=4`.

### Bước 1 — Luật loại trừ (hard filter)

- `employees.employment_status != ACTIVE` hoặc `users.status != ACTIVE` → loại.
- **Thiếu hoàn toàn** một kỹ năng có `is_mandatory = true` → loại khỏi Top-N.
  (Nhánh thủ công UC003.2 vẫn hiện người này kèm cảnh báo — bước 10 của UC003.2.)
- Chỉ đọc `employee_skills` **đã duyệt** (`approved_at IS NOT NULL`). Quy tắc R3.

### Bước 2 — Điểm khớp kỹ năng

Với mỗi yêu cầu `i` của công việc:

```
s_i        = clamp(level_nhân_viên / level_yêu_cầu, 0, 1)     # đạt hoặc vượt đều = 1.0
SkillScore = Σ(weight_i × s_i) / Σ(weight_i)                   ∈ [0, 1]
```

Nếu có kỹ năng bắt buộc mà **có nhưng chưa đủ mức**: `SkillScore ×= 0.7`.

### Bước 3 — Điểm tải công việc

```
load(E)       = Σ estimate_hours các task của E ở trạng thái ASSIGNED | IN_PROGRESS | PENDING_REVIEW
WorkloadScore = clamp(1 - load(E)/CAPACITY, 0, 1)              # CAPACITY mặc định 80h, để trong config
```

### Bước 4 — Điểm hiệu suất quá khứ

Trên N task `DONE` gần nhất của nhân sự:

```
PerfScore = clamp(Σ estimate_hours / Σ actual_hours, 0, 1.2) / 1.2
          = 0.5 (trung tính) nếu chưa có dữ liệu — phải nói rõ trong phần giải thích
```

Đây là dữ liệu **tuỳ chọn** (UC003.1 đầu vào #5: "có thể không tồn tại với nhân sự mới").

**Quy ước lưu trữ**: khi không có dữ liệu thì ghi `recommendation_items.perf_score = NULL`,
**không ghi 0.5**. 0.5 là giá trị trung tính dùng *trong công thức*, không phải một quan sát thật.
Ghi 0.5 xuống bảng sẽ không còn phân biệt được "người thực sự có hiệu suất trung bình" với "người
chưa có dữ liệu nào" — mà Chương 5 cần đúng con số "bao nhiêu phần trăm ứng viên được chấm khi
không có dữ liệu hiệu suất". Câu giải thích cũng sinh từ `perf_score IS NULL` thay vì so sánh với
hằng số. Kết quả vẫn tái lập được vì `params` đã lưu quy ước "NULL thì dùng 0.5".

`skill_score` và `workload_score` thì luôn `NOT NULL` — hai giá trị này luôn tính được.

### Bước 5 — Tổng hợp

```
Score          = α×SkillScore + β×WorkloadScore + γ×PerfScore    # α=0.7, β=0.2, γ=0.1 (config)
matching_score = round(Score × 100, 2)                            # khớp decimal(5,2), thang 0–100
```

Bonus tuỳ chọn: đã là `project_members` của dự án → `+0.03`, cộng sau, cap ở 1.0.

### Bước 6 — Sinh giải thích

Từ template, ví dụ:

> "Đáp ứng 4/5 kỹ năng yêu cầu. Thiếu Docker (yêu cầu Cơ bản). Đang phụ trách 2 công việc
> (24h/80h). Chưa có dữ liệu hiệu suất."

Giải thích là **bắt buộc** (`recommendation_items.explanation NOT NULL`), không được để rỗng.

### Xếp hạng

Sắp giảm dần theo `matching_score`. Khi bằng điểm, phải có tiêu chí phá hoà **tất định** (ví dụ
`load` thấp hơn trước, rồi tới `employee_id` nhỏ hơn) để kết quả tái lập được — có test riêng cho
tính ổn định của thứ tự.

## 5. Lưu kết quả

Mỗi lần PM bấm chạy đề xuất, mỗi công việc sinh **một** `recommendation_runs`:

- `engine_version` — đổi mỗi khi công thức hoặc trọng số thay đổi.
- `params jsonb` — lưu α, β, γ, CAPACITY, N. Có bộ tham số này thì tái lập được kết quả cũ.
- `generated_by_employee_id`, `generated_at`.

Mỗi ứng viên trong Top-N là một `recommendation_items` với đủ điểm thành phần (`skill_score`,
`workload_score`, `perf_score`), `matched_skills`/`missing_skills` dạng jsonb, `current_workload`,
`explanation`, và `is_selected`.

Khi PM xác nhận phân công, `assignments.recommendation_item_id` trỏ về đúng dòng đã chọn và
`assignment_type = AI`. Nhờ vậy Chương 5 tính được tỉ lệ PM chọn đúng Top-1.

Lưu kết quả cũng là để **tránh tính lại** khi PM xem lại — yêu cầu tường minh của mục 2.4.3.
