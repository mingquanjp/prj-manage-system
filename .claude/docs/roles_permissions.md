# Vai trò và phân quyền — nghiệp vụ hiện hành

**Đồng bộ quyết định 15/10/2026:** [sổ quyết định, phần hiện hành và Bàn giao](../../docs/design/database-review-decisions-2026-10-14.md) ưu tiên cho nội dung đã thay thế; [dictionary/DBML hợp nhất](../../docs/design/database-proposal-2026-10-14.md) là thiết kế 33 bảng (cấp bậc theo chốt 26/10/2026). Đây là thiết kế DB, chưa triển khai DDL/migration hoặc backend nghiệp vụ; tiến độ frontend mô phỏng xem [sổ công việc](../../web/README.md#sổ-công-việc).


> Baseline nghiệp vụ hiện hành: [2026-10-12](../../docs/business-baseline-2026-10-12.md), PTCM phân công và nghiệm thu, không nhận task trong dự án phụ trách; tra cứu giới hạn dữ liệu nghề nghiệp. Phần kỹ thuật V2 chưa đồng bộ không được ghi đè quyết định mới; công thức và quy tắc không bị thay thế giữ nguyên.

Thiết kế đích theo quyết định hiện hành; bổ sung quyền tạo tài khoản kèm hồ sơ ban đầu ngày 20/10/2026. Mốc 2026-09-06 là lịch sử tài liệu; chưa implement auth.
Nguồn quyết định: [baseline hiện hành](../../docs/business-baseline-2026-10-12.md); V2 là bối cảnh thiết kế kỹ thuật cũ.

## 1. Vai trò và trách nhiệm hiện hành

Quản trị viên, HR và Giám đốc là vai trò ở phạm vi hệ thống/tổ chức. PM và phụ trách chuyên môn là trách nhiệm trong dự án; một người có thể đảm nhiệm khác nhau ở các dự án. Nhân viên là nền tảng chung. Không suy quyền từ chức danh, tag hoặc một enum role toàn hệ thống. Mô hình đề xuất cho doanh nghiệp nhỏ–vừa, không phải quy tắc chung cho mọi doanh nghiệp.

| Chức năng | Người chịu trách nhiệm và phạm vi |
|---|---|
| Quản lý tài khoản, chức năng quản trị hệ thống | Quản trị viên; không mặc nhiên duyệt kỹ năng hoặc điều hành dự án |
| Xem/cập nhật/vô hiệu hóa hồ sơ hành chính, quản lý CV nhân sự | HR; không sửa kỹ năng chính thức qua quyền cập nhật hồ sơ |
| Xét duyệt yêu cầu cập nhật thông tin cá nhân | HR; pending chưa đổi hồ sơ, từ chối có lý do; fullname/phone/address; email công ty/cơ cấu HR sửa trực tiếp |
| Xem hồ sơ/căn cứ và xét duyệt kỹ năng từ CV hoặc đề nghị chủ động | Giám đốc; không tự duyệt, cần căn cứ và lý do; không thêm tầng HR duyệt trước |
| Khởi tạo dự án, giao PM | Giám đốc; người tạo không mặc nhiên là PM |
| Theo dõi, điều phối cấp dự án, bắt đầu/hoàn thành/đóng dự án | PM của dự án; không mặc nhiên phân công hoặc nghiệm thu |
| Gợi ý, phân công, thay/thu hồi người thực hiện | PTCM của dự án; kiểm tra lại quyền tại lúc xác nhận |
| Giao phụ trách chuyên môn | PM của dự án; PM có thể kiêm nhiệm |
| Cấu trúc, đặc tả và tạo task | Phụ trách chuyên môn của dự án; PM chỉ khi kiêm nhiệm |
| Nghiệm thu | PTCM hiện hành của dự án; không tự nghiệm thu và không nhận task trong dự án đó |
| Hồ sơ cá nhân, gửi yêu cầu kỹ năng | Người có hồ sơ của mình theo quyền cá nhân hiện hành |
| Nhận task, ghi nhật ký, nộp kết quả | Người được giao hợp lệ, không phải PTCM của chính dự án |
| Hiệu chỉnh effort đã khóa | PTCM hiện hành duyệt; nhân viên tự gửi, không nhập/gửi thay |
| Chỉnh sửa PM | Giám đốc chọn PM mới và lưu; không có quy trình bàn giao riêng |
| Cập nhật CV cá nhân | Nhân viên xác nhận CV hiện tại sau kiểm tra tệp; không phê duyệt CV |

Phụ trách chuyên môn (PTCM) của dự án phân công và nghiệm thu công việc, không trực tiếp nhận/thực hiện task trong chính dự án đó. Không giao thêm trách nhiệm nghiệm thu, không có reviewer thay thế do PM chỉ định. PM chỉ có quyền chuyên môn khi kiêm PTCM và cũng phải tuân thủ giới hạn không nhận task. PTCM có thể thực hiện task ở dự án khác nếu đủ điều kiện. Lưu người đánh giá, kết luận và thời điểm; link PR/bình luận/merge không tự thay thế quyết định nghiệm thu trong hệ thống.

Chỉ định/thay PTCM phải kiểm tra người được chọn không còn nhận task mở trong cùng dự án; chuyển giao trước theo workflow. Kiểm tra điều kiện đối ứng khi phân công và khi thay PTCM trong transaction. Lịch sử đóng góp cũ không bị xóa; không tự nghiệm thu đầu ra do mình đã thực hiện.

### R06 — Một nguồn dữ liệu, góc nhìn theo mục đích

**Tra cứu hồ sơ nhân sự** là danh sách/tìm kiếm độc lập cho PM và nhân viên đang được giao PTCM; nhân viên thông thường không xem hồ sơ người khác. Bốn nhóm dữ liệu: thông tin nhận diện/công tác cần thiết; kỹ năng và năng lực chính thức; kinh nghiệm/kết quả thực hiện tổng hợp được phép; tải công việc tổng hợp. Không trả dữ liệu hành chính nhạy cảm, CV gốc người khác hoặc raw task/log/minh chứng dự án khác. Kiểm tra quyền tại mỗi lần đọc, không suy quyền xuyên dự án từ membership hay trách nhiệm PTCM ở nơi khác.

Note UC phân rã: “Nhân viên chỉ được tra cứu khi được giao phụ trách chuyên môn trong dự án.” Tổng quan nối cả PM và Nhân viên, không cần lặp note. Tra cứu nhân sự phù hợp trong phân công là ngữ cảnh task của PTCM, dùng chung nguồn dữ liệu; quyền xem danh sách không tự cấp quyền phân công. DTO chi tiết phải chốt khi thiết kế, không tự trả toàn bộ entity.

Không tạo hai bản kỹ năng chính thức. PM xem thông tin nghề nghiệp phục vụ theo dõi/lựa chọn PTCM; PTCM xem để lựa chọn người thực hiện; Giám đốc xem hồ sơ và căn cứ phục vụ đánh giá/phê duyệt kỹ năng; HR xem hồ sơ hành chính/CV theo mục đích quản lý nhân sự. Quản trị viên không kế thừa các quyền này vì có quyền quản trị tài khoản. HR không mặc nhiên đọc căn cứ task, nhật ký hoặc toàn dự án; Giám đốc không mặc nhiên cập nhật hồ sơ hành chính, duyệt thông tin cá nhân hoặc quản lý tệp CV. Người kiêm nhiệm dùng quyền tương ứng còn hợp lệ.

HR cập nhật trực tiếp và HR xét duyệt yêu cầu là hai thao tác riêng, có lịch sử. Vô hiệu hóa hồ sơ không đồng nhất với khóa tài khoản của Admin và không cấp quyền thay PM. Allowlist đã chốt fullname/phone/address; email công ty/cơ cấu HR sửa trực tiếp. Địa chỉ cần bổ sung khi hoàn thiện hồ sơ, không bắt Admin nhập khi tạo; chủ hồ sơ đề nghị, HR duyệt/sửa trực tiếp; không được xóa trắng địa chỉ đã có. Không thêm quyền đọc địa chỉ cho PM/PTCM hoặc Giám đốc từ quyền tra cứu/xét kỹ năng; không chặn nghiệp vụ vì địa chỉ chưa bổ sung. Xem [quy trình hồ sơ](employee_profile_workflow.md).

Giám đốc được đọc căn cứ task được dẫn chiếu: mô tả, yêu cầu kỹ năng, đóng góp cá nhân, kết quả, minh chứng và nhận xét nghiệm thu. Không mặc định mở toàn dự án, task khác, nhật ký hoặc toàn bộ bình luận. Nhân viên chỉ dẫn chiếu trong phạm vi mình được xem. Khi thay người, đối chiếu đúng đóng góp lịch sử. Màn hình tập hợp căn cứ và nhận xét; danh sách chỉ số tổng hợp cụ thể chưa chốt, không tự coi điểm gợi ý là điểm năng lực.

Giám đốc cần ý kiến chuyên môn phù hợp khi không đủ cơ sở đánh giá; không thêm cấp duyệt bắt buộc. Không tự duyệt kỹ năng, kể cả nhập trực tiếp. Chi tiết: [Phê duyệt cập nhật kỹ năng](skill_approval_workflow.md).

PM/PTCM không được tải CV gốc của người khác chỉ từ quyền tra cứu nhân sự. Nhân viên xác nhận bản đang sử dụng, không cần Giám đốc duyệt tệp. CV hiện tại không chứng minh mọi kỹ năng đã được công nhận; engine chỉ dùng kỹ năng chính thức. Xem [quy trình CV](cv_extraction_approval_workflow.md).

PTCM cần được xem kết quả và căn cứ của công việc được đánh giá trong dự án được giao trách nhiệm; không vì thế có quyền đọc hồ sơ nhân sự nhạy cảm hoặc dữ liệu dự án khác. Phạm vi DTO/field cụ thể sẽ thiết kế sau, không tự cấp quyền đọc mọi dữ liệu.

Membership lịch sử không tự cấp quyền toàn dự án. Người kết thúc lượt phân công chỉ còn quyền lịch sử theo tư cách người thực hiện; nếu đồng thời có trách nhiệm dự án hợp lệ khác, quyền đó được xét riêng. Notification/download/run/comment phải kiểm tra quyền; biết ID không chứng minh được phép.
Tổng tải không phải lịch trống; chỉ số lịch sử chưa phải KPI khách quan. Raw snapshot nội bộ không mặc định trả cho ADMIN, Giám đốc hoặc PM.

**Tiếp nối 28/10:** PTCM gia hạn task theo trạng thái/lý do/mốc gốc đã chốt; người hiện được giao và PTCM hiện hành đọc/trao đổi trên task, không suy quyền này từ PM-only/HR/Admin/Giám đốc. Task/project kết thúc chỉ đọc trong phạm vi còn quyền; người đã mất phân công không được đọc trao đổi hiện hành nhờ lịch sử. Xem [quyết định tiếp nối](../../docs/business-baseline-2026-10-20.md#quyết-định-tiếp-nối-28102026--thay-pm-hạn-task-và-trao-đổi).

## 2. Kiểm tra quyền hiện hành

Phạm vi phân công theo [bổ sung 2026-10-13](../../docs/business-baseline-2026-10-13.md): PTCM được chọn trực tiếp nhân viên hợp lệ trong toàn công ty cho dự án phụ trách, không cần membership trước hoặc PM phê duyệt từng lần. Không phát sinh quyền điều hành dự án khác. Số dự án đang có việc/tải xuyên dự án chỉ là tổng hợp được phép; không vì chỉ số này mà trả tên dự án, task/log hoặc tài liệu ngoài quyền đọc.

Hash cookie secret → tra session còn hạn/chưa revoked + user/employee ACTIVE + role hiện hành → RBAC → service ABAC.
Mọi API nghiệp vụ phải kiểm tra hiện trạng DB, có thể gộp truy vấn; chấp nhận roundtrip để thu hồi quyền ngay.
Không lưu quyền do client cung cấp hoặc coi cookie là nguồn role hiện hành.
Actor lấy từ auth; employee_id được chọn trong body phân công là đối tượng, phải validate riêng.

Mutation kiểm tra lại quyền/dữ liệu nền trong transaction với quy ước khóa chung.
Yêu cầu đang chạy cạnh tranh với khóa/chuyển PM phải có thứ tự commit xác định; không hứa hủy ngược giao dịch đã commit.
Sai phiên: 401; sai quyền: 403; không tồn tại: 404. Chính sách này không che tuyệt đối sự tồn tại tài nguyên.
Deny by default, ngoại lệ public khai báo tường minh.

## 3. Tài khoản và chỉnh sửa PM

Chốt tiếp nối [28/10/2026](../../docs/business-baseline-2026-10-20.md#quyết-định-tiếp-nối-28102026--tự-cấp-mã-và-nhãn-quyền-tài-khoản): hệ thống tự cấp employee_code khi tạo; form hiển thị dự kiến chỉ đọc, sau lưu trả mã được cấp. E001/e001 là quy ước mã/tên đăng nhập trong demo; cách đăng nhập và quy tắc username cho backend chưa chốt. Không đổi tài khoản cũ, không tái dùng mã khóa/vô hiệu hóa; backend bảo đảm cấp mã đồng thời và duy nhất. UI gọi “Quyền hệ thống”, HR hiển thị “Quản lý nhân sự”; mã role, ma trận quyền và trách nhiệm dự án giữ nguyên. Vị trí/cấp bậc không tự cấp quyền.



Theo [quyết định 20/10/2026](../../docs/business-baseline-2026-10-20.md), Admin tạo user + employee atomically trong một lần, nhập thông tin hồ sơ ban đầu cần thiết. Quyền này chỉ phục vụ tạo mới, không cấp quyền cập nhật hành chính thường xuyên hoặc duyệt thông tin cá nhân; HR quản lý hồ sơ về sau. Không API tạo tài khoản mồ côi. Chính sách cấp/reset mật khẩu vẫn theo security.md, chưa tự bổ sung quy trình mới.
Người thực hiện active nhận task nếu không là PTCM của dự án đó; PM kiêm PTCM cũng bị loại. Thêm vai trò hệ thống không bị chặn chỉ vì đang có công việc. Chỉ chặn thu hồi quyền nếu việc đó khiến người dùng không còn quyền xử lý công việc/trách nhiệm đang mở; xử lý trách nhiệm liên quan trước. Thay vai trò hệ thống không tự thay PM hoặc trách nhiệm chuyên môn trong dự án.
Không vô hiệu hóa/hạ quyền ADMIN active cuối cùng.
Khóa khẩn cấp được phép dù còn việc; thu hồi mọi session và liệt kê việc cần PTCM xử lý phân công và PM xử lý trách nhiệm dự án.
Giám đốc thay PM bằng thao tác chỉnh sửa người quản lý dự án: chọn PM mới đủ điều kiện và lưu một lần. Lưu thành công thì PM mới thay PM cũ; nếu lưu thất bại, PM cũ vẫn giữ trách nhiệm. Không yêu cầu xóa PM cũ trước, không có quy trình bàn giao riêng hoặc chờ PM cũ xác nhận. Giữ lịch sử thay đổi; không thay đổi trạng thái dự án, phân công công việc hoặc đóng góp đã có. PM mới chỉ có quyền hợp lệ theo trạng thái hiện hành; các trách nhiệm khác của PM cũ được xét độc lập.
Áp dụng cho project PLANNING/IN_PROGRESS/COMPLETED theo giới hạn hiện hành. Quản trị viên không có quyền thay PM; không thêm yêu cầu bàn giao hoặc cấp duyệt.
Chỉnh sửa PM không thay đổi project state; PM mới là PM hiện tại, tiếp nhận trách nhiệm còn hợp lệ theo state machine, PM cũ mất quyền quản lý. Project COMPLETED vẫn thuộc trách nhiệm PM cho tới CLOSED: PM mới đóng khi đủ điều kiện; time correction do PTCM hiện hành duyệt, không được nghiệm thu thêm, tạo task, phân công/thay người, mở lại execution hoặc đưa project về IN_PROGRESS. Không mở thêm quyền ngoài state machine hiện hành.
Người đã bị thay chỉ gửi yêu cầu sửa cho đóng góp của mình qua DTO giới hạn. Không nhập/gửi thay cho người mất quyền. PTCM hiện hành xử lý yêu cầu đã gửi theo điều kiện trạng thái, không PM duyệt từ trách nhiệm PM.

Đổi role/khóa/reset mật khẩu thu hồi tất cả session; logout thu hồi session hiện tại.

## 4. Xác thực — session phía server

Cookie secret ngẫu nhiên → SHA-256 để tra auth_sessions.token_hash → kiểm tra expires_at/revoked_at/user hiện hành → quyền hệ thống và trách nhiệm dự án. session_id là ID nội bộ, không cookie. Không JWT access/refresh hoặc refresh endpoint; không localStorage cho thông tin xác thực.

Cookie host-only/HttpOnly/Secure/SameSite=Lax, Path=/api; production HTTPS. Mutation kiểm tra Origin và token CSRF gắn phiên; login kiểm tra nguồn. Logout thu hồi phiên và xóa cookie đúng path; khóa/reset/đổi role thu hồi các phiên trong giao dịch. Timeout tuyệt đối/idle và cleanup chưa chốt; bộ sáu cột không cam kết idle timeout, không tự thêm last_seen_at.

HR quản trị departments/positions/job_levels; Giám đốc quản trị skill_categories/skills. Quyền danh mục không cấp thêm quyền khác. Username khác email công ty. Nhân viên có tư cách từ employees, PM/PTCM từ projects; user_roles chỉ ADMIN/HR/DIRECTOR. Không thêm bảng permission.

Bcrypt và chính sách mật khẩu xem security.md; seed chỉ dev, không dùng mật khẩu mẫu triển khai thực.
