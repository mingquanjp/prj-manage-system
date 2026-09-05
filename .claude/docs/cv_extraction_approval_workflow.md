# CV → trích xuất → duyệt — V1

Thiết kế đích đã đồng ý 2026-09-05; không yêu cầu mô hình học máy.
Xem [baseline](../../docs/reviews/system-design-review-v1.md) và `database_schema.md`.

## 1. Tiếp nhận và xử lý

EMPLOYEE upload của mình; ADMIN có thể upload thay mặt, ghi actor.
PDF có text/DOCX, tối đa 10 MiB; không OCR/PDF mã hóa. CV Việt/Anh là phạm vi thử nghiệm phải đo riêng.
Local disk private qua storage.Provider, server sinh tên, download phân quyền, không URL công khai.
Kiểm tra kiểu nội dung thật, giới hạn giải nén DOCX, thời gian/bộ nhớ parser; không chỉ kiểm tra đuôi file.

Tạo CV PENDING, không thay CV ACTIVE cũ.
processing_status: QUEUED → PROCESSING → SUCCEEDED/FAILED.
Worker trong process nhận job DB, có lease/retry giới hạn; cập nhật idempotent.
Lưu file bền vững trước enqueue; DB/file không cùng transaction, dọn file tạm/orphan khi thất bại.
Retry không tạo trùng dòng kỹ năng; lỗi không làm mất CV chính thức.

## 2. Nhận diện kỹ năng

Parse text → normalize → exact/alias danh mục → fuzzy candidate → ADMIN map hoặc reject.
Không tự suy proficiency/năm kinh nghiệm từ việc có tên skill.
suggested_level có thể NULL; muốn approve phải có skill_id, mức và căn cứ do ADMIN xác nhận.
confidence_score chưa hiệu chỉnh chỉ là điểm tương đồng, không hiển thị xác suất đúng.

Dòng map sang skill dùng cùng employee_skill_change_requests, source=CV, extracted_skill_id unique.
Nhiều dòng map cùng skill phải hợp nhất minh chứng, không ghi đè tuần tự.
Đã có pending manual cùng người/skill → báo conflict; giải quyết/rút trước.
Approval kiểm tra version, ghi quyết định + official skill cùng transaction.
WITHDRAWN được xem là dòng đã bỏ qua, không phải approved, khi tổng hợp kết quả review.

## 3. Ba trạng thái độc lập

- Publication ở employee_cvs: PENDING/ACTIVE/REPLACED/REJECTED/ARCHIVED.
- Processing ở cv_extractions: QUEUED/PROCESSING/SUCCEEDED/FAILED.
- Review kỹ năng: PENDING_REVIEW/APPROVED/PARTIALLY_APPROVED/REJECTED.

Chỉ SUCCEEDED được review/xuất bản. ADMIN duyệt publication: CV mới ACTIVE, cũ REPLACED atomically; tối đa một ACTIVE/người.
Duyệt CV không tự duyệt skill; từ chối CV không âm thầm thu hồi skill đã duyệt, muốn thu hồi phải có yêu cầu REMOVE.
CV parse lỗi vẫn PENDING với processing FAILED, có thể retry/rút ARCHIVED; CV cũ không đổi.

Review còn dòng chưa xử lý → PENDING_REVIEW; hết dòng:
toàn approved → APPROVED; approved và bị từ chối/bỏ qua → PARTIALLY_APPROVED; không approved → REJECTED.
Extraction rỗng vẫn SUCCEEDED; ADMIN đóng review với NO_SKILLS_FOUND, UI “Không có kỹ năng để cập nhật”.
Request là nguồn quyết định chính; trạng thái dòng/review là dữ liệu suy ra, không cập nhật độc lập trái nhau.

## 4. Nghiệm thu

Upload mới/parse lỗi không đổi kỹ năng và CV cũ.
Không approve dòng chưa map/thiếu mức; concurrent approve chỉ một quyết định.
Đo precision/recall/F1 nhận diện skill trên CV có nhãn và được phép sử dụng, không lấy proficiency chưa có căn cứ làm nhãn.
Thư viện parser chưa chọn; kiểm tra bảo trì/bảo mật trước khi thêm dependency.
