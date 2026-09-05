# Kiểm thử và nghiệm thu — V1

Kế hoạch kiểm thử, không phải kết quả đã chạy. Skeleton chưa có test nghiệp vụ.
make check hiện gồm fmt/tidy/vet/build/test; xanh khi không có test không chứng minh nghiệp vụ.

## 1. Các tầng

- Unit: scorer, state rules, giao lịch/giây và làm tròn.
- Integration PostgreSQL riêng qua TEST_DATABASE_URL: FK/unique, khóa, transaction, rollback, migration.
- API/httptest: DTO, auth/ABAC, CSRF, status/error/idempotency, OpenAPI contract.
- Browser E2E: UI thực với ADMIN/PM/EMPLOYEE; httptest không thay browser test.

Thiếu DB test có thể skip ở local nhưng CI/nghiệm thu phải báo rõ số skipped, không tính đã đạt.
Không chạy integration/reset trên DB thật hoặc seed lại để “test”.

## 2. Ma trận bắt buộc

| Miền | Ca và invariant |
|---|---|
| User/role | Tạo user+employee atomic; role/status hiện hành; ADMIN cuối; đổi role còn việc bị chặn; khóa khẩn cấp rồi bàn giao |
| Quyền | PM chéo project; nhân viên cũ sau reassign; CV pending; notification/run/download không bypass |
| Skill/profile | Pending không đổi chính thức; stale version; hai ADMIN approve; ADD/UPDATE/REMOVE; email unique khi duyệt |
| CV | Parse lỗi giữ CV cũ; không skill; map trùng; chưa map/level NULL; withdrawn; hai publish concurrent |
| Project | Start chưa có ASSIGNED; task không start trong PLANNING; complete còn việc bị chặn; cancel đóng mọi phiên; terminal chỉ đọc |
| Task | Group chéo project; thiếu requirement; ngày thiếu cặp/ngoài project; spec frozen sau unassign; không reopen DONE |
| Assignment | Một dòng lỗi rollback mọi bảng; duplicate task; same key retry; khác payload; role/skill đổi cạnh tranh |
| Reassign | Đóng giờ người cũ; người mới ASSIGNED; PENDING_REVIEW phải trả lại; membership không cấp quyền |
| Time | Một open session/người; start đồng thời; pause/review không tính; qua trưa/đêm/cuối tuần/ngày nghỉ; calendar snapshot cũ |
| Recommendation | Mandatory thiếu/dưới mức đều loại; optional thiếu=0; vượt mức cap=1; inactive/sai role; Top-N rỗng; snapshot replay |
| Score | Load 0/80/>80; <3 mẫu dùng 0,5; actual=0 loại; task nhiều người loại; weight/param không hợp lệ; tie ổn định |
| Quyết định | Top1/Top2/manual sau run; reason bắt buộc; run cũ/spec/skill đổi; không tin điểm client; một commit/run |
| Security | Logout/lock JWT cũ bị từ chối; refresh reuse/concurrent; CSRF refresh khi access hết hạn; body/file quá lớn; log không bí mật |

Ví dụ lịch chuẩn 20/05/2026 15:00 → 21/05/2026 10:30 phải 4,5h.
Đối soát actual_hours với tổng valid_work_seconds; retry pause/submit không cộng trùng.
Route test chi.Walk chỉ method/path; contract test kiểm tra schema/response/error/security nữa.

## 3. Nghiệm thu các mốc

M1 auth/quyền/reset; M2 request/duyệt; M3 project/group/task/state; M4 atomic/manual/reassign;
M5 scorer/snapshot/giải thích; M6 time/concurrency; M7 CV publication + skill request; M8 thực nghiệm/báo cáo.
Mỗi mốc phải có ca âm tính và rollback, không chỉ màn hình happy path.

## 4. Đánh giá công thức và số liệu báo cáo

So sánh skill-only, skill+workload, đủ ba thành phần trên nhãn phù hợp độc lập.
Chia dữ liệu hiệu chỉnh/đánh giá; khảo sát trọng số, ngưỡng 80, mã hóa proficiency, PerfScore.
Đo Top-K, tỷ lệ PM chọn Top-1/Top-3 từ assignment commit có run không rỗng; không dùng is_selected.
Tách giao đầu/giao lại, báo số mẫu, run rỗng/bỏ dở, người mới và trường hợp tải bão hòa.
Quên pause/estimate thiên lệch là hạn chế dữ liệu; không dùng kết quả làm kết luận năng suất khách quan.
CV đo precision/recall/F1 trên dữ liệu có nhãn/được phép sử dụng.
Unit test được xây theo công thức chỉ chứng minh tính toán đúng, không chứng minh chất lượng xếp hạng.

Mục tiêu thử nghiệm, chưa đo: API p95 ≤1s warm; 20 task×100 người ≤3s; CV hợp lệ ≤30s.
Ghi phần cứng, dataset, concurrency, cold/warm riêng; không lấy latency quan sát cũ làm SLA.
Không viết đã phỏng vấn/thử nghiệm/đạt chất lượng khi chưa có dữ liệu.
