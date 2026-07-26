# Luồng xử lý CV và duyệt kỹ năng

Thuộc mốc **M7**. Tài liệu đặc tả luồng này **mỏng hơn** UC001–UC004 (không có bảng use case chi
tiết), nên phần dưới đây tách rõ *cái gì có trong tài liệu* và *cái gì còn phải chốt*.

## 1. Nguyên tắc

Dữ liệu CV và kết quả trích xuất **không bao giờ** cập nhật trực tiếp hồ sơ năng lực chính thức.
Chúng tạo dữ liệu chờ duyệt để Quản trị viên kiểm tra trước khi được dùng cho tra cứu hoặc AI đề
xuất (mục 4.2.3.4, mục 2.4.2, mục 2.1.2). Đây là quy tắc **R3** trong `domain_rules.md`.

Hệ quả kiểm thử: sau khi upload và trích xuất mà **chưa** duyệt, engine đề xuất phải cho ra kết quả
**y hệt** như trước khi upload.

## 2. Ba bảng

```
employee_cvs  ──1:n──▶  cv_extractions  ──1:n──▶  cv_extracted_skills
   (tệp CV)              (nội dung trích xuất)      (từng kỹ năng nhận diện được)
```

**`employee_cvs`** — `storage_provider`, `storage_key`, `file_url`, `original_file_name`,
`mime_type`, `uploaded_at`, `status ∈ {ACTIVE, ARCHIVED, REPLACED}`.

**`cv_extractions`** — `extracted_text` (văn bản thô), `extracted_markdown`,
`extraction_status ∈ {PENDING_REVIEW, APPROVED, REJECTED}`, `extracted_at`, `reviewed_at`,
`reviewed_by_employee_id`, `review_note`.

**`cv_extracted_skills`** — `raw_skill_name` (tên đọc được từ CV), `skill_id` (**cho phép NULL**),
`suggested_level`, `suggested_years_experience`, `confidence_score`,
`decision_status ∈ {PENDING, APPROVED, REJECTED}`, `evidence_text`.

`skill_id NULL` nghĩa là tên kỹ năng đọc được **chưa khớp** với danh mục `skills`. Quản trị viên
phải map thủ công sang một `skill_id` hợp lệ **trước khi** duyệt — không được ghi vào
`employee_skills` khi `skill_id` còn NULL.

`evidence_text` giữ đoạn văn bản gốc trong CV làm căn cứ, để ADMIN duyệt có cơ sở thay vì tin
confidence score suông.

## 3. Luồng dự kiến

1. Nhân viên (hoặc ADMIN thay mặt) upload tệp CV → tạo `employee_cvs` với `status = ACTIVE`.
   CV cũ của cùng người chuyển sang `REPLACED`.
2. Hệ thống trích xuất văn bản → tạo `cv_extractions` với `extraction_status = PENDING_REVIEW`.
3. Nhận diện kỹ năng → tạo nhiều `cv_extracted_skills`, mỗi dòng `decision_status = PENDING`,
   kèm `confidence_score` và `evidence_text`.
4. ADMIN mở màn hình duyệt: xem văn bản trích xuất, từng kỹ năng đề xuất, mức độ gợi ý, điểm tin
   cậy và đoạn minh chứng.
5. ADMIN map `skill_id` cho dòng nào còn NULL, chỉnh `suggested_level` nếu cần, rồi duyệt hoặc từ
   chối **từng dòng**.
6. Với mỗi dòng được duyệt, ghi vào `employee_skills` với `source = CV`,
   `approved_by_employee_id` = ADMIN, `approved_at = now()`. Nếu cặp `(employee_id, skill_id)` đã
   tồn tại thì cập nhật, do có ràng buộc UNIQUE.
7. Cập nhật `cv_extractions.extraction_status`, `reviewed_at`, `reviewed_by_employee_id`,
   `review_note`.

Bước 5–7 nằm trong một transaction.

## 4. Còn phải chốt (TBD)

| Hạng mục | Tình trạng |
|---|---|
| **Nơi lưu file CV** | Mặc định local disk qua interface `storage.Provider`; đổi sang S3/MinIO sau mà không sửa nghiệp vụ. **Chưa chốt.** |
| **Thư viện parse PDF/DOCX** | Chưa chọn. Chưa có dependency nào trong `go.mod`. |
| **Cách nhận diện kỹ năng** | Đối sánh với danh mục `skills` (khớp chuỗi, chuẩn hoá, bí danh) hay dùng mô hình NLP. Đây là **thành phần học máy thật của đề tài** nên cần bàn kỹ trước khi code. |
| **Cách tính `confidence_score`** | Phụ thuộc lựa chọn trên. |
| **Giới hạn kích thước và định dạng tệp** | Tài liệu không nêu. |
| **Ai được upload CV** | Tài liệu có cả "Nhân viên yêu cầu cập nhật CV" (2.2.4) và "ADMIN cập nhật CV nhân sự" (2.2.3). Cần chốt luồng chính. |

Không tự quyết những mục trên. Hỏi trước khi implement.

## 5. Duyệt kỹ năng nhập tay

Song song với luồng CV, nhân viên có thể gửi yêu cầu cập nhật kỹ năng tự khai (mục 2.2.4). Luồng
này **không có bảng riêng**: dùng chính `employee_skills` với `source = MANUAL` và
`approved_at IS NULL` để biểu diễn trạng thái chờ duyệt. ADMIN duyệt bằng cách điền
`approved_by_employee_id` và `approved_at`.

Mọi truy vấn phục vụ tra cứu và đề xuất phải lọc `approved_at IS NOT NULL` — không có ngoại lệ.
