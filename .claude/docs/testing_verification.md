# Kiểm thử và nghiệm thu — V2

**Đồng bộ quyết định 15/10/2026:** [sổ quyết định, phần hiện hành và Bàn giao](../../docs/design/database-review-decisions-2026-10-14.md) ưu tiên cho nội dung đã thay thế; [dictionary/DBML hợp nhất](../../docs/design/database-proposal-2026-10-14.md) là thiết kế 32 bảng. Chưa triển khai code/DDL/migration.


## Bổ sung 2026-10-13 — Nguồn nhân lực chung

Theo [baseline mới](../../docs/business-baseline-2026-10-13.md). Đây là **kỳ vọng nghiệm thu chưa chạy trên code/DB**, không phải kết quả test pass.

| Ca | Kỳ vọng |
|---|---|
| Người đủ điều kiện chưa từng được giao trong dự án | Được xét, không yêu cầu membership trước; vẫn lọc kỹ năng bắt buộc ở nhánh gợi ý |
| PTCM của chính dự án có điểm cao | Loại trước xếp hạng; manual/confirm không override |
| Ba task mở cùng dự án | Số dự án đang có việc = 1, không phải 3 |
| Có việc mở ở hai dự án rồi được xét dự án thứ ba | Không bị loại chỉ vì số dự án; cảnh báo tải/kiểm tra hiện hành vẫn áp dụng |
| Task cuối ở PENDING_REVIEW | Vẫn tính dự án đang có việc; chuyển DONE thì không tính nếu không còn task mở khác |
| Giao lại task cuối từ A sang B | Đếm hiện hành chuyển theo người hiện được giao; lịch sử A và B vẫn còn |
| Thu hồi task cuối về PENDING_ASSIGNMENT | Không đếm người cũ là đang có việc tại dự án đó; lịch sử giữ nguyên |
| Task/project terminal hoặc người từng nhận đã inactive | Không tính việc mở; lịch sử không biến mất vì trạng thái nhân sự/task |
| Có assignment nhưng chưa làm, không log | Danh sách từng được giao có người đó; không tự kết luận có kinh nghiệm/đóng góp |
| PM/PTCM chưa nhận task nào | Trách nhiệm dự án không tự cộng vào số dự án đang có việc thực hiện |
| Draft/run hoặc confirm rollback | Không thêm lịch sử tham gia, không tăng số đếm/tải chính thức |
| Confirm thành công rồi đọc/chạy mới | Phản ánh phân công mới xuyên dự án; run cũ bất biến, số liệu hiện hành phân biệt rõ |
| Xem tổng tải/số dự án của ứng viên | Không từ đó mở raw task/log/CV hoặc tên dự án ngoài quyền |
| Dữ liệu membership cũ MANUAL/không assignment/chênh joined_at | Chặn việc kết luận DROP an toàn cho tới kiểm kê/bảo toàn; không tạo assignment giả |

Nguồn nghiệp vụ ưu tiên: [baseline 2026-10-12](../../docs/business-baseline-2026-10-12.md). Bổ sung HR, tách xét duyệt hồ sơ và kỹ năng; các ca công thức hiện có không được thay đổi trong đợt này.

Kế hoạch kiểm thử, không phải kết quả đã chạy. Skeleton chưa có test nghiệp vụ.
make check hiện gồm fmt/tidy/vet/build/test; xanh khi không có test không chứng minh nghiệp vụ.

## 1. Các tầng

- Unit: scorer, state rules, ngày/phút/tổng effort, eligibility lịch sử và làm tròn.
- Integration PostgreSQL riêng qua TEST_DATABASE_URL: FK/unique, khóa, transaction, rollback, migration.
- API/httptest: DTO, auth/ABAC, CSRF, status/error/idempotency, OpenAPI contract.
- Browser E2E: UI thực với Quản trị viên/HR/Giám đốc/PM/Nhân viên và trách nhiệm PTCM theo dự án; httptest không thay browser test.

Thiếu DB test có thể skip ở local nhưng CI/nghiệm thu phải báo rõ số skipped, không tính đã đạt.
Không chạy integration/reset trên DB thật hoặc seed lại để “test”.

## 2. Ma trận bắt buộc

| Miền | Ca và invariant |
|---|---|
| User/role | Tạo user+employee atomic; role/status hiện hành; ADMIN cuối; thêm vai trò không bị chặn chỉ vì còn việc; chặn thu hồi quyền cần cho trách nhiệm đang mở; khóa khẩn cấp vẫn được phép |
| Tạo ban đầu 20/10 | Admin tạo user+employee cùng một lần, lỗi rollback cả hai; quyền tạo không cho Admin sửa hồ sơ hành chính về sau; HR tiếp tục quản lý hồ sơ |
| Quyền | PM chéo project; nhân viên cũ sau reassign; CV không phải bản hiện tại; notification/run/download không bypass |
| Skill | Pending không đổi chính thức; stale before_values/request.version; hai thao tác Giám đốc approve; ADD/UPDATE/REMOVE; HR không duyệt hoặc sửa trực tiếp official skill |
| Profile | HR xét duyệt; pending/từ chối không đổi hồ sơ, từ chối có lý do; đối chiếu before_values; chỉ fullname/phone/address qua yêu cầu. Address chưa có được NULL khi tạo, UI nhắc bổ sung; thêm/sửa cần nội dung, không xóa trắng, không lộ tra cứu chuyên môn |
| CV | Tệp hợp lệ vẫn dùng được khi parse lỗi; không skill; map trùng; chưa map/level NULL; hai lần xác nhận CV hiện tại cạnh tranh |
| Project | Start chưa có ASSIGNED; task không start trong PLANNING; complete còn việc bị chặn; COMPLETED cho correction không làm thêm; close chặn pending; cancel kết thúc pending/giữ log; terminal chỉ đọc |
| Task | Group chéo project; thiếu requirement; ngày thiếu cặp/ngoài project; spec frozen sau unassign; không reopen DONE |
| Assignment | Một dòng lỗi rollback mọi bảng; duplicate task; same key retry; khác payload; role/skill đổi cạnh tranh |
| Reassign | Giữ log người/assignment cũ, có bàn giao; người mới ASSIGNED; PENDING_REVIEW phải trả lại; membership không cấp quyền |
| Effort | Nhiều task IN_PROGRESS; 6h công ty+4h ở nhà=10h; >8h reason, >24h chặn xuyên project; phút nguyên; tương lai; chia ngày qua đêm; không cắt cuối tuần |
| Hiệu chỉnh | ADD có result_log_id, UPDATE/DELETE có history trước/sau; pending không đổi tổng; stale có đường reject/withdraw; PTCM hiện hành duyệt; cấm ghi/gửi thay |
| Review effort | Chỉ sửa giờ không trả việc; thiếu nhật ký không chặn nghiệm thu đầu ra, không bịa log; loại dữ liệu thời gian không tin cậy khỏi hiệu suất; DONE correction không reopen; quá hạn kế hoạch vẫn ghi được |
| Cạnh tranh effort | Hai project cùng ghi vượt 24h; đổi ngày; approve vs close/cancel/reassign; retry không cộng trùng; log luôn khớp assignment |
| Recommendation | Mandatory thiếu/dưới mức đều loại; optional thiếu=0; vượt mức cap=1; user/employee inactive hoặc sai role bị loại; danh mục không có trạng thái inactive; Top-N rỗng; snapshot replay |
| Score (đối chứng cũ, không contract cuối) | Load 0/80/>80; <3 mẫu dùng 0,5; actual=0 loại; task nhiều người loại; weight/param không hợp lệ; tie ổn định |
| Quyết định | Top1/Top2/manual sau run; reason bắt buộc; run cũ/spec/skill đổi; không tin điểm client; một commit/run |
| Security | Session đã logout/lock/revoke bị từ chối; cookie secret chỉ lưu hash; CSRF/Origin trên mutation; body/file quá lớn; log không bí mật |

Đối soát actual_hours = tổng phút hiệu lực/60, gồm cả task hủy khi tổng hợp project. Pending và log deleted_at khác NULL không cộng vào tổng.
Ví dụ 360+240=600 phút=10h, cảnh báo vượt 8h; 1441 phút/người/ngày bị chặn.
Ví dụ đối chứng công thức cũ, không chốt tham số triển khai: nhật ký 8h estimate/10h actual cho r=0,8; 8h/6h và 8h/8h đều r=1.
Ba mẫu r=(1;1;0,5) cho PerfScore=5/6; log không đủ tin cậy/nhiều người/pending bị loại trước chọn 10 mẫu mới nhất.
CV: NO_CHANGE không tạo UPDATE; không skill đóng NO_CHANGES; tất cả rút không gọi là reject; đổi CV giữ nguồn yêu cầu và skill đã approved.
Route test chi.Walk chỉ method/path; contract test kiểm tra schema/response/error/security nữa.

### Ca đối chiếu R01–R06 — baseline giản lược

Các ca dưới đây thay kế hoạch R01–R05 cũ; không triển khai nộp thay, nhiều attempt hoặc xác nhận lịch sử cũ. Đây là yêu cầu kiểm thử, chưa phải kết quả chạy code.

| ID | Thiết lập/thao tác | Kết quả bắt buộc |
|---|---|---|
| R01-A | Có thời gian thực tế hợp lệ nhưng quên Start/Resume; còn quyền ghi nhật ký trực tiếp | Ghi nhật ký theo quy tắc chung, không duyệt ngoại lệ riêng, không sửa ngược mốc trạng thái |
| R01-B | Nhật ký đã khóa hoặc thuộc lượt phân công cũ cần sửa | Dùng hiệu chỉnh chung; không bỏ kiểm tra ngày, lượt phân công, tổng phút hoặc trạng thái dự án |
| R02-A | Người thực hiện nộp kết quả nhưng thiếu nhật ký | Nghiệm thu theo đầu ra; không ép giờ giả, dữ liệu thiếu/không tin cậy không dùng cho hiệu suất phụ thuộc thời gian |
| R02-B | PM/người khác thử nộp kết quả thay | Không có chức năng nộp thay; không suy quyền nộp từ quyền hiệu chỉnh nhật ký |
| R03-A | Bổ sung căn cứ khi yêu cầu kỹ năng PENDING | Bổ sung trong yêu cầu hiện tại, không cần rút/gửi lại chỉ vì thiếu minh chứng |
| R03-B | Bổ sung sau quyết định kết thúc | Tạo yêu cầu mới, có thể dẫn CV cũ; không mở lại mục CV hoặc sửa kết quả cũ |
| R03-C | Retry hoặc duyệt cạnh tranh | Không nhân đôi quyết định; tối đa một PENDING/người/kỹ năng, giữ lịch sử |
| R04-A | Kỹ năng từng bị loại được đề nghị bổ sung lại | Xét duyệt thông thường, giữ lịch sử; không thêm chức năng tái kích hoạt riêng |
| R04-B | ADD/UPDATE khi official hoặc căn cứ PENDING đã đổi | Đối chiếu before_values và request.version; dữ liệu lỗi thời không ghi đè chính thức |
| R05-A | Dữ liệu lịch sử thay đổi sau khi chạy gợi ý | Run cũ bất biến và có thời điểm tạo; có thể chạy lại, không yêu cầu history acknowledgement |
| R05-B | Chọn từ run cũ nhưng điều kiện phân công hiện hành không còn hợp lệ | Kiểm tra và chặn theo quy tắc hiện hành; không dùng snapshot để vượt điều kiện |
| R06-A | Tra cứu thủ công rồi chạy gợi ý | Thủ công không sinh rank/score; gợi ý có giải thích và thời điểm tạo |
| R06-B | Xem thông tin ứng viên thuộc dự án khác hoặc truy cập raw run | Giữ phạm vi dữ liệu theo mục đích; Admin không mặc nhiên được đọc raw snapshot |
| R06-C | PM/PTCM tải CV gốc người khác từ quyền tra cứu | Bị chặn kể cả CV hiện tại; chỉ trả dữ liệu nghề nghiệp theo quyền |

### Bốn quyết định chốt ngày 2026-09-11 — scenario đặc tả

| Phạm vi | Thiết lập/thao tác | Kết quả bắt buộc |
|---|---|---|
| Thay trực tiếp | A giữ task ASSIGNED/IN_PROGRESS/PAUSED, PTCM chọn B rồi xác nhận thất bại | A vẫn giữ phân công, task giữ trạng thái; không ghi một phần thay đổi |
| Thay trực tiếp | Xác nhận A → B thành công trong project PLANNING/IN_PROGRESS | A kết thúc trách nhiệm, B nhận task ASSIGNED và phải Start; giữ kết quả/log/đóng góp A, frozen spec và handover theo rule hiện hành |
| Tiền điều kiện | Thay trực tiếp khi task không PENDING_ASSIGNMENT | Dùng luồng reassign riêng, không kế thừa tiền điều kiện UC003/UC003.2; PENDING_REVIEW vẫn phải trả lại trước |
| Gợi ý giao lại | Thu hồi A rồi run rỗng, bỏ lựa chọn hoặc confirm thất bại | Task PENDING_ASSIGNMENT, không current assignee, không tự khôi phục A; lịch sử A giữ nguyên |
| Gợi ý giao lại | Thu hồi A, run mới, chọn B và confirm thành công | Task ASSIGNED cho B; không dùng run đã commit hoặc xem gợi ý người thay khi A còn giữ task |
| Frozen skill | Task từng giao có frozen requirement X mức 3, B có official X hiện được công nhận mức 3 | B đáp ứng mandatory X; X tham gia SkillScore và được dùng cho recommendation/manual/reassignment; không đảm bảo B nằm Top-N |
| Frozen skill | Cùng task nhưng official X đã REMOVE, còn pending hoặc dưới mức | Không tính như kỹ năng đạt yêu cầu; engine loại nếu X mandatory; ngoại lệ manual có lý do giữ nguyên |
| Chỉnh sửa PM | Project COMPLETED, Giám đốc chỉnh sửa từ PM A sang PM B đủ điều kiện | Giữ COMPLETED, B là PM hiện tại, A mất quyền; B đóng khi đủ điều kiện; correction do PTCM hiện hành duyệt |
| Giới hạn PM | B sau thay PM ở COMPLETED cố nghiệm thu/tạo task/phân công/thay người/mở lại execution/đưa về IN_PROGRESS | Không được phép; còn correction PENDING vẫn không được đóng project |

Đây là ca đối chiếu tài liệu và hợp đồng kiểm thử cho triển khai; không phải bằng chứng code hoặc database đã thực thi đúng.

## 3. Nghiệm thu các mốc

M1 auth/quyền/reset; M2 request/duyệt; M3 project/group/task/state; M4 atomic/manual/reassign;
M5 scorer/snapshot/giải thích; M6 time/concurrency; M7 cập nhật CV không phê duyệt + skill request; M8 thực nghiệm/báo cáo.
Mỗi mốc phải có ca âm tính và rollback, không chỉ màn hình happy path.

## 4. Đánh giá công thức và số liệu báo cáo — phương án thử nghiệm cũ

So sánh skill-only, skill+workload, đủ ba thành phần trên nhãn phù hợp độc lập.
Chia dữ liệu hiệu chỉnh/đánh giá; khảo sát trọng số, ngưỡng 80, mã hóa proficiency, PerfScore.
Đo Top-K, tỷ lệ PTCM chọn Top-1/Top-3 từ assignment commit có run không rỗng; không dùng is_selected.
Tách giao đầu/giao lại, báo số mẫu, run rỗng/bỏ dở, người mới và trường hợp tải bão hòa.
Effort tự khai thiếu/sai, hỗ trợ chưa ghi nhận và estimate thiên lệch là hạn chế dữ liệu; không dùng kết quả làm kết luận năng suất khách quan.
CV đo precision/recall/F1 trên dữ liệu có nhãn/được phép sử dụng.
Unit test được xây theo công thức chỉ chứng minh tính toán đúng, không chứng minh chất lượng xếp hạng.

Mục tiêu thử nghiệm, chưa đo: API p95 ≤1s warm; 20 task×100 người ≤3s; CV hợp lệ ≤30s.
Ghi phần cứng, dataset, concurrency, cold/warm riêng; không lấy latency quan sát cũ làm SLA.
Không viết đã phỏng vấn/thử nghiệm/đạt chất lượng khi chưa có dữ liệu.

## 5. Scenario đối chiếu nghiệp vụ 2026-10-05 — thay PM và cập nhật CV

| Scenario | Kết quả yêu cầu |
|---|---|
| Giám đốc chọn PM B thay A và lưu thành công | B quản lý thay A; không chờ A xác nhận, không đổi task/state/lịch sử |
| Lưu PM mới thất bại | A vẫn là PM; không có bước xóa A trước |
| Admin thử thay PM | Không có quyền thay PM từ quyền quản trị tài khoản |
| PM cũ có trách nhiệm chuyên môn độc lập | Xét quyền đó riêng, không tự thu hồi chỉ vì thay PM |
| Nhân viên tải CV nhưng chưa xác nhận sử dụng | CV hiện tại chưa đổi |
| Xác nhận CV mới hợp lệ khi trích xuất FAILED | CV mới được sử dụng; có thể gửi yêu cầu kỹ năng thủ công |
| Thay CV khi yêu cầu kỹ năng còn PENDING hoặc đã APPROVED | Giữ yêu cầu, nguồn căn cứ, quyết định và official skill; không tự duyệt/rút/từ chối |
| Giám đốc từ chối yêu cầu kỹ năng từ CV | Không từ chối hoặc thay CV hiện tại |
| Tải CV ngoài quyền tra cứu | Bị chặn; không coi cập nhật CV là công khai tệp |

Đây là scenario đặc tả, không phải kết quả chạy code.

## 6. Scenario làm rõ ngày 2026-10-06

| Scenario | Kết quả yêu cầu |
|---|---|
| CV trích xuất được kỹ năng nhưng nhân viên không gửi yêu cầu thủ công | Giám đốc xem kỹ năng và nội dung nguồn để xét duyệt; không bắt gửi từng đề nghị thủ công |
| CV chỉ nêu tên kỹ năng, chưa đủ căn cứ xác định mức | Làm rõ hoặc nêu lý do chưa công nhận; không tự suy mức, không từ chối tệp CV |
| Nhân viên chủ động đề nghị thay đổi kỹ năng mà không tải CV | Xét duyệt bình thường với giải trình; tệp/liên kết không bắt buộc |
| CV mới có kỹ năng không đổi hoặc không còn nhắc kỹ năng đã duyệt | Không bắt duyệt lại kỹ năng không đổi, không tự xóa kỹ năng vắng trong CV |
| Hai luồng cùng đề cập một người/kỹ năng đang PENDING | Không tạo hai yêu cầu đang chờ trùng nhau; giữ nguồn căn cứ và xử lý theo quy tắc chung |
| A đang nhận task mở, PM chọn A làm PTCM cùng dự án | Chặn chỉ định; phải chuyển giao task trước, không xóa đóng góp A |
| Giao trách nhiệm nghiệm thu riêng cho một thành viên | Không có chức năng này trong phiên bản hiện tại |
| Thành viên thông thường xem PR rồi thử xác nhận nghiệm thu | Không có quyền; chỉ PTCM hiện hành được nghiệm thu |

Đây là scenario đối chiếu đặc tả, không phải kết quả kiểm thử phần mềm.

## 7. Scenario baseline 2026-10-07 — HR và ranh giới quyền

| ID | Thao tác/tình huống | Kết quả yêu cầu |
|---|---|---|
| HR01 | Nhân viên gửi yêu cầu cập nhật thông tin cá nhân | Lưu đề nghị riêng, hồ sơ chính thức giữ nguyên khi chờ |
| HR02 | HR chấp thuận yêu cầu hợp lệ | Cập nhật đúng nội dung duyệt, giữ lịch sử/người xử lý; kiểm tra quyền/before_values |
| HR03 | HR từ chối | Giữ hồ sơ cũ; nhân viên xem kết quả và lý do |
| HR04 | HR cập nhật trực tiếp rồi xử lý yêu cầu dựa trên dữ liệu cũ | Không ghi đè từ yêu cầu lỗi thời; cập nhật trực tiếp có lịch sử riêng |
| HR05 | Admin/Giám đốc/PM không kiêm HR thử duyệt thông tin cá nhân | Không được phép từ riêng các vai trò đó |
| HR06 | HR duyệt kỹ năng hoặc sửa official skill qua cập nhật hồ sơ | Bị chặn; Giám đốc xử lý hai luồng kỹ năng, không cần HR duyệt trước |
| HR07 | HR upload CV thay nhân viên | Ghi đúng người thao tác; không tự duyệt kỹ năng hoặc tự đổi CV hiện tại |
| HR08 | Tệp CV hợp lệ nhưng parse lỗi hoặc kỹ năng bị từ chối | Không từ chối tệp; nhân viên vẫn xác nhận CV hiện tại theo quy trình |
| HR09 | HR quản lý hồ sơ rồi truy cập task/nhật ký ngoài quyền | Không suy quyền đọc dự án từ quyền nhân sự |
| HR10 | HR vô hiệu hóa hồ sơ hoặc Admin khóa tài khoản | HR chưa được vô hiệu hóa nếu còn công việc/trách nhiệm dự án mở; thông báo các mục cần xử lý. Admin vẫn được khóa khẩn cấp; giữ lịch sử, phân biệt hai trạng thái và không tự bàn giao |
| HR11 | Nhân viên gửi role/trách nhiệm dự án/kỹ năng trong yêu cầu thông tin cá nhân | Không cập nhật các dữ liệu ngoài phạm vi này |
| HR12 | Thiết kế form trường cần duyệt | Chỉ fullname/phone/address qua yêu cầu; email công ty/cơ cấu HR sửa trực tiếp. Địa chỉ chưa có hiển thị nhắc bổ sung, không bắt lúc tạo; patch bỏ qua không đổi, gửi rỗng/NULL bị chặn; HR duyệt/ghi trực tiếp không xóa trắng địa chỉ đã có |
| HR13 | HR thử tạo dự án, thay PM hoặc chỉ định PTCM từ quyền HR | Không được phép; xét riêng nếu có trách nhiệm hợp lệ khác |
| AD01 | Nghiệm thu chưa đạt trên activity | Quay về chỉnh sửa rồi nộp lại, không chỉ nộp lại kết quả cũ |
| AD02 | Activity gộp kiểm tra và ghi nhận; kiểm tra thực tế thất bại | Không lưu phân công/triển khai từ kết quả thất bại; chi tiết theo UC003 |
| AD03 | Một task đạt nhưng dự án còn task chưa kết thúc | Không tự hoàn thành dự án; PM đối chiếu điều kiện toàn dự án |

Đây là hợp đồng kiểm thử để triển khai sau, không phải 16 ca test code đã chạy.

## 8. Scenario baseline 2026-10-12 — quyền PTCM, nghiệm thu và tra cứu

Các ca dưới đây là hợp đồng kiểm thử đích, **chưa chạy trên phần mềm**. Review tài liệu không thay kiểm thử triển khai.

| ID | Tình huống | Kết quả yêu cầu |
|---|---|---|
| CM01 | PM không kiêm PTCM chạy gợi ý/confirm/reassign/unassign | Từ chối; quyền PM không cấp quyền phân công |
| CM02 | PTCM hiện hành thao tác cùng dự án mở | Được thực hiện nếu thỏa điều kiện task/candidate; batch vẫn nguyên tử |
| CM03 | PTCM dự án A phân công/nghiệm thu ở B khi không có trách nhiệm tại B | Từ chối; không cấp quyền xuyên dự án |
| CM04 | Manual chọn PTCM cùng dự án, kể cả PM kiêm PTCM | Không cho chọn/confirm; lý do override kỹ năng không bỏ qua điều kiện này |
| CM05 | Chạy gợi ý có PTCM kỹ năng cao/tải thấp | Loại trước xếp hạng; không tạo task thực hiện cho PTCM |
| CM06 | PTCM A là ứng viên task thuộc dự án B | Không loại chỉ vì là PTCM tại A; xét các điều kiện khác tại B |
| CM07 | PM chọn người đang nhận task mở làm PTCM | Chặn đến khi chuyển giao trước; không tự xóa task/log/assignment |
| CM08 | Run/draft cũ chọn B; sau đó B trở thành PTCM của dự án | Confirm từ chối; snapshot không vượt điều kiện hiện hành; rollback cả batch |
| CM09 | Chỉ định PTCM và confirm giao task cho cùng người xảy ra đồng thời | Không cả hai cùng commit tạo xung đột; kiểm tra/khóa phối hợp |
| CM10 | PTCM A bị thay bởi B trong khi A đang mở form phân công/review | A mất quyền chuyên môn khi thay có hiệu lực; kiểm tra lại khi ghi; giữ quyết định đã commit trước đó |
| CM11 | Người đã thực hiện một phần đầu ra được chọn PTCM sau chuyển giao | Không tự nghiệm thu đầu ra cũ; tránh xung đột khi chỉ định, không chỉ kiểm tra current_assignee |
| CM12 | Thành viên thường/PM chỉ xem, góp ý hoặc merge PR | Không tự được quyền nghiệm thu; không tự chuyển DONE |
| CM13 | PTCM đánh giá không đạt | IN_PROGRESS có lý do; nhân viên chỉnh sửa rồi tự nộp lại, không nộp thay |
| CM14 | Đổi PTCM hoặc người thực hiện | Không xóa/đổi tác giả log, submission, kết luận hoặc công thức Q/T |
| CM15 | Nhân viên thường gọi danh sách hồ sơ người khác | Từ chối; quyền hồ sơ của mình vẫn riêng |
| CM16 | PM/PTCM tra cứu danh sách/chi tiết | Chỉ bốn nhóm dữ liệu nghề nghiệp được phép; không hồ sơ HR nhạy cảm/CV gốc/raw dữ liệu dự án khác |
| CM17 | PM/PTCM dùng URL tải CV người khác dù là bản hiện tại | Không cho tải từ quyền tra cứu; xét quyền HR/chủ hồ sơ/căn cứ Giám đốc riêng nếu kiêm nhiệm |
| CM18 | PTCM gửi kế hoạch trên Slack | Không cần thao tác/approval trong hệ thống; không tích hợp Slack |
| CM19 | PTCM không kiêm PM bấm bắt đầu/hoàn thành/đóng dự án | Từ chối; PM thực hiện khi đủ điều kiện |
| CM20 | Giao thêm task trong dự án IN_PROGRESS | Không yêu cầu PM bắt đầu lại dự án; một task DONE không tự complete project |
| CM21 | Chỉ định PTCM cho người chỉ có task lịch sử đã kết thúc | Không chặn chỉ vì lịch sử; không mở lại task hoặc đổi người nghiệm thu quá khứ |
| CM22 | PTCM bị khóa tài khoản/vô hiệu quyền | Không nghiệm thu hoặc phân công bằng session cũ; không tự chuyển quyền cho PM; xử lý trách nhiệm theo quyền PM |

Đối chiếu giữ nguyên: điều kiện start/complete/close, quyền HR/Giám đốc, frozen spec, log/correction, manual override kỹ năng, công thức và số liệu nghiên cứu. PM không được tự duyệt từ quyền PM; PTCM hiện hành duyệt correction, không có nghiệp vụ ghi/gửi thay.

## 9. Ca bổ sung theo quyết định 15/10 — chưa chạy

| Ca | Kỳ vọng thiết kế |
|---|---|
| ADD hiệu chỉnh được PTCM duyệt; retry hoặc hai lượt duyệt đồng thời | Một log mới, result_log_id trỏ đúng assignment; log/quyết định cùng commit, không history ADD; không tạo lần hai |
| UPDATE/DELETE trỏ log khác assignment hoặc đã xóa | Từ chối; không ghi một phần; đối chiếu before_values dưới khóa |
| Xóa mềm hợp lệ | History before/after, tổng bỏ log, snapshot nộp cũ giữ nguyên; không restore |
| PM thử duyệt correction hoặc người khác gửi thay | Từ chối; PTCM hiện hành nhận và duyệt, requested_by đúng chủ assignment |
| Giao lại/thu hồi/DONE/hủy | Ghi bộ ended một lần, tối đa một lượt mở/task; đóng dự án không ghi đè bộ ended; tải dựa cả state |
| Hủy trong khi chờ review | Không giả REJECTED; giữ submission/log, kết thúc lượt nguyên tử và loại khỏi hàng chờ bằng state |
| Tạo dự án không chọn PM | Từ chối; không tạo project PM NULL |
| REMOVE official rồi ADD lại | Gỡ dòng hiện tại, giữ request/căn cứ; công nhận lại tạo dòng mới |
| Kỹ năng vắng lúc run nhưng xuất hiện trước confirm | Phát hiện thay đổi; không chỉ so các dòng đã tồn tại |
| Bổ sung tệp PENDING rồi duyệt từ màn hình cũ | Metadata upload do server xác thực, request.version đổi; duyệt cũ bị chặn; terminal bất biến |
| Session | token_hash unique; không secret/JWT/refresh trong DB/log; revoke/expiry/quyền kiểm tra hiện hành; timeout/idle/reset cụ thể chờ contract |
| Idempotency | Chỉ endpoint có contract; cùng key khác payload conflict, result và mutation cùng commit; không hứa chống lặp sau retention |
| Dữ liệu cá nhân/danh mục | HR chỉ quản trị cơ cấu, Giám đốc danh mục kỹ năng; tra cứu PM/PTCM không CV gốc/contact/raw task khác dự án |

Kiểm tra tài liệu tĩnh xem review 15/10; không thay integration/race/security test.
