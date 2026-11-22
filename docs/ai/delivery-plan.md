# Kế hoạch triển khai và báo cáo hằng tuần — đến 14/01/2027

Ngày đối chiếu: **14/11/2026**. Trạng thái: **Quân đã đồng ý hướng triển khai theo FR/bản đồ miền ngày 14/11; các mốc là kế hoạch dự kiến, chưa phải cam kết hoàn thành hoặc ticket đã giao**. Người dùng dự kiến dành **15–20 giờ/tuần**, báo cáo cô mỗi thứ Năm và hướng đến **80–85% web app vào giữa tháng 1**. Lấy thứ Năm **14/01/2027** làm mốc lập kế hoạch; cần điều chỉnh nếu lịch cô khác.

Nguồn: [mục lục](../README.md), [62 FR/AC và các điểm mở](../product/functional-requirements.md), [6 miền và phần thông báo](../product/business-domains.md), [kiến trúc](../../.claude/docs/architecture.md), [thiết kế DB và mapping](../design/database-step5c-migration-map-2026-10-14.md), [hiện trạng FE](../../web/README.md), [runbook auth](../api/auth-runbook.md). Đây là thứ tự thực hiện, không thay đổi nghiệp vụ, FR, schema hoặc tự chốt các điểm O01–O09.

**Chốt tiếp nối 14/11:** tiếp tục triển khai theo FR/AC và bản đồ miền đã có. Công thức gợi ý phục vụ phân công vẫn chưa chốt; các điểm còn mở sẽ thảo luận riêng sau. Không tự chọn trọng số, chuẩn hóa, PerfScore hoặc áp công thức nghiên cứu vào backend. Tra cứu/phân công thủ công và phần nghiệp vụ độc lập có thể làm trước; scorer, điểm/hạng và nghiệm thu FR-044 chờ phương pháp được duyệt. Mốc gợi ý trong lịch là có điều kiện, không ấn định ngày người dùng phải chốt công thức.

## 1. Kết quả đối chiếu hiện trạng

**Tiếp nối 19/11/2026 — DATN-35 đã kiểm, chờ bàn giao Git:** OAP đã chạy luồng UI danh mục → tạo → Gmail thật → người dùng tự đổi/login mới → hồ sơ/reload/logout/Back. Bảng dưới cập nhật phần đã nối sau mốc lập kế hoạch14/11. [Runbook xuyên luồng](../api/auth-runbook.md#7-datn-35--oap-t13--demo-xuyên-ui-db-và-gmail) giữ kết quả thực/FR/giới hạn; kế hoạch không tự là bằng chứng runtime hoặc trạng thái Done.

| Phần | Có thể dùng lại | Còn phải thực hiện |
|---|---|---|
| D01 auth/session | Năm API, service/repository, migration auth, cookie/CSRF/limiter, FE auth và kiểm DB/race | Khép các ca browser còn mở DATN-21; không đồng nhất auth với toàn bộ miền tài khoản |
| D01 quản trị tài khoản | M2 bỏ username/email unique/sequence; transaction create, ba API Admin, Gmail adapter, UI tạo/tra cứu; DATN-35 kiểm mail thật → đổi/login mới → hồ sơ | Bàn giao/merge OAP theo workflow; quyền/khóa/reset tài khoản cũ và quản trị Admin cuối chưa nối. Cấp lại/gửi lại vẫn hoãn |
| D02 hồ sơ/cơ cấu | M1, 12 API/UI danh mục; API/UI đọc hồ sơ chính chủ, quyền HR/Admin/self và nullable | Lịch sử/đề nghị cá nhân → HR duyệt, HR đọc/cập nhật hồ sơ và kiểm trách nhiệm khi vô hiệu hóa chưa nối |
| D03 kỹ năng/CV | UI danh mục, đề nghị/xét kỹ năng, metadata CV | Official/pending/history thật, stale/version, storage và file thật; parser/model/provider chưa chốt |
| D04 dự án/công việc | UI các bước dự án, task, phân công và thực hiện/nghiệm thu | API, migration, quyền theo tài nguyên, trạng thái/khóa đặc tả, batch và lịch sử đúng lượt |
| D05 effort | UI nhật ký/hiệu chỉnh và logic demo | Transaction, tổng theo người/ngày xuyên dự án, snapshot nộp, lịch sử và chống lặp |
| D06 tra cứu/gợi ý | UI ứng viên/nháp/giải thích; tài liệu và phương án nghiên cứu | Projection thật, công thức được duyệt, engine thuần dữ liệu, snapshot/run và đánh giá phương pháp |
| H01 thông báo | UI/model prototype | Lưu cùng transaction nghiệp vụ, đọc/đánh dấu theo người nhận và kiểm quyền khi mở |

`web/src/App.tsx` hiện tách ServerSession và DemoApp; không chỉ bật cờ là toàn bộ workspace dùng API thật. Các màn đã duyệt được tái sử dụng, nhưng cần nối route/layout và adapter dần theo feature; route chưa nối phải báo giới hạn, không trộn principal thật với employee fixture.

Repo hiện có migration 000001–000005 (head M2); đây chưa là schema đích 33 bảng. Không dùng số bảng/tên bảng giống nhau làm bằng chứng backend đã hoàn thành. Kiểm kê Neon trong mapping mang ngày 06/11, không phải kiểm kê mới ngày 18/11. DATN-35 chỉ kiểm môi trường local/PG test sở hữu, không áp Neon.

## 2. Năng lực thực hiện và cách đo

Khoảng 8,7 tuần đến 14/01 tương ứng **130–175 giờ của Quân**. Đề xuất giữ **25% cho review, sửa lỗi, kiểm xuyên luồng, học/đọc code và chuẩn bị báo cáo**; còn khoảng **98–131 giờ cho đầu ra mới**. Đây là ngân sách kế hoạch, không phải estimate đã đo. Quân xác nhận dùng Codex hằng ngày để triển khai và đọc hiểu phần quan trọng: 15–20 giờ là thời gian Quân trực tiếp tham gia, không phải tổng thời gian tự viết mọi dòng code. Agent có thể xử lý triển khai/test/tài liệu trong phạm vi đã duyệt; năng lực phải hiệu chỉnh bằng tốc độ thực của hai tuần đầu. Thời gian agent hỗ trợ không thay thời gian Quân cần đọc hiểu, kiểm tra và chốt thiết kế. Review và kiểm chứng phần quan trọng là yêu cầu xuyên suốt đã được thống nhất.

**80–85% toàn app là mục tiêu có rủi ro cao với ngân sách này**, vì ngoài auth/OAP các miền nghiệp vụ lớn vẫn chưa có backend. Không cam kết chỉ từ số màn prototype. Sau hai báo cáo 19/11 và 26/11 phải so công việc hoàn thành/giờ thực với estimate để xác nhận hoặc cắt đợt tiếp theo; không hạ yêu cầu bảo mật để giữ ngày.

Cách báo tiến độ đề xuất:

- Giữ mẫu số **62 FR hiện hành**, không tự thu nhỏ mẫu số theo phạm vi demo. FR chỉ tính hoàn thành khi các AC áp dụng đủ, FE dùng dữ liệu thật, lưu/reload được, quyền/lỗi và kiểm thử phù hợp đạt; PR đã merge theo workflow.
- FR mới làm một phần ghi **Partial**, không tính tròn là Done. Auth có đổi lần đầu không đồng nghĩa FR-004 đổi tự nguyện đã xong; FR-003 còn phải đối chiếu các thao tác quản trị gây thu hồi quyền.
- **50/62 ≈80,6%; 53/62 ≈85,5%** chỉ là chỉ báo độ phủ chức năng, không có nghĩa các FR tốn công ngang nhau hoặc chất lượng đạt85%. Báo kèm các luồng cốt lõi và rủi ro; kiểm thử/bảo mật không được bù bằng nhiều CRUD nhỏ.
- Lập bảng FR–feature–ticket–PR–bằng chứng khi proposal từng feature được duyệt. Không tạo cả trăm ticket hoặc điền trước phần trăm/ngày pass trong bước này.

## 3. Phạm vi ưu tiên để hướng tới mục tiêu

**Phần cốt lõi phải có luồng chạy thật trên UI:** tài khoản/hồ sơ → năng lực được công nhận → dự án/task → lựa chọn/phân công → thực hiện/nhật ký → nộp/nghiệm thu → gợi ý có giải thích trên nguồn dữ liệu thật. CV thật và thông báo được nối theo năng lực thực hiện; không dùng file metadata giả làm minh chứng upload thành công.

Một phương án ưu tiên để đối chiếu **mục tiêu 50 FR**: thực hiện các FR hiện hành, trừ12 mục dự kiến xếp sau: **FR-004** (phần ngoài đổi lần đầu), **FR-013, FR-021, FR-022, FR-025, FR-028, FR-030, FR-035, FR-036, FR-056, FR-057, FR-063**. Đây là **đề xuất xếp lịch**, chưa phải quyết định bỏ tính năng hoặc xác nhận 50 FR khả thi. Kỹ năng/history vẫn giữ bất biến và quyền dù UI tra cứu nâng cao FR-028 xếp sau; nhật ký bị khóa vẫn phải bị chặn dù chưa có UI hiệu chỉnh.

Ba ứng viên mở rộng để tiến từ 50 tới 53 là FR-028/FR-025/FR-013, tùy effort đo thực và các đầu vào đã đủ. Nếu cô ưu tiên trích xuất CV, cần đưa FR-021/022 lên và chuyển phần khác xuống trước khi triển khai; không cộng thêm vô hạn vào lịch.

Mọi FR trong nhóm mục tiêu vẫn cần đủ AC. Ví dụ FR-005 còn gửi mật khẩu tạm, FR-024 còn căn cứ/tệp trong quyền, FR-026 cần đúng dữ liệu đóng góp, FR-059 bao phủ sự kiện theo nguồn: khi mới có một phần hoặc đầu vào chưa triển khai thì giữ Partial. Nếu số Done thấp hơn50, báo đúng số; không đổi định nghĩa để đạt80%.

## 4. Chín mốc báo cáo thứ Năm

Các mốc dưới là **đích demo dự kiến**, chưa là estimate chi tiết hoặc nghiệm thu FR trọn vẹn. Tuần đầu chỉ có năm ngày; tuần 31/12 cần tính lại lịch intern/cá nhân. Kiểm API/DB/UI trong chính tuần triển khai, không dồn hết sang 14/01.

| Báo cáo | Miền/feature ưu tiên | Phần cô thấy trực tiếp trên UI | Điều kiện và phương án khi chậm |
|---|---|---|---|
| **19/11/2026** | D02 nền danh mục và hồ sơ; chuẩn bị feature D01+D02 tạo tài khoản | Auth thật; HR xem/quản lý danh mục tổ chức ưu tiên vị trí; nhân viên đọc hồ sơ thật. Nếu đủ mới trình diễn Admin tạo→login→đổi→hồ sơ | Đây là mức tối thiểu ưu tiên. Không hứa trọn tạo tài khoản/gửi email trước khi chốt O01 và migration; nếu chưa đủ chỉ demo danh mục/hồ sơ với tài khoản local tổng hợp, nói rõ nguồn |
| **26/11/2026** | D01+D02 cấp/quản trị tài khoản và hồ sơ | Tạo user–employee→login/đổi→hồ sơ; danh sách tài khoản, quyền/khóa theo phần đã kiểm; hồ sơ đề nghị→HR duyệt theo năng lực tuần | Chốt credential/mail và cơ chế thu hồi quyền; không nhận toàn FR-005 nếu chỉ có kênh local. Đây là mốc hiệu chỉnh năng lực, tránh gom tất cả thao tác admin/HR vào một PR |
| **03/12/2026** | D03 danh mục/năng lực và đề nghị thủ công | Danh mục kỹ năng→nhân viên đề nghị→Giám đốc duyệt→official thay đổi trên hồ sơ | Căn cứ task chưa có thì không giả dữ liệu; FR liên quan giữ Partial. Storage/parser cần chốt trước feature CV; phần công thức gợi ý thảo luận riêng sau, không chặn các feature độc lập |
| **10/12/2026** | D04 khởi tạo dự án, trách nhiệm, nhóm/task/requirements | Giám đốc tạo với PM→PM chỉ định PTCM→PTCM thiết lập task→PM Start | Quyền PM/PTCM theo dự án, không role toàn hệ thống. Có dữ liệu kỹ năng thật làm đầu vào, không cần chờ CV extraction |
| **17/12/2026** | D04+D06 tra cứu/thủ công, nháp và xác nhận phân công | PTCM chọn ứng viên thật→thấy tải dự kiến→confirm→nhân viên thấy việc của mình | Ưu tiên một batch hoàn chỉnh trước thay/thu hồi; batch atomic/chống lặp và kiểm lại nền bắt buộc. Thủ công chưa chạy engine không có điểm/hạng giả |
| **24/12/2026** | D04+D05 thực hiện, effort và nghiệm thu | Start/Pause/Resume→ghi/sửa nhật ký→nộp→PTCM trả sửa hoặc nghiệm thu; reload giữ đúng lịch sử | Không bắt đủ log để nghiệm thu. Khi khóa không cho sửa trực tiếp; hiệu chỉnh sau khóa có thể xếp sau nhưng không mở đường bypass |
| **31/12/2026** | D06 gợi ý giải thích được và truy vết quyết định | Chạy gợi ý→Top-N/giải thích→chọn/confirm→tải cập nhật; dữ liệu run cũ giữ nguyên | Công thức phải được duyệt trước, có bộ tình huống đánh giá. Không tự lấy trọng số nghiên cứu làm quyết định; nếu trễ thì trình bày luồng manual thực, chưa nhận FR-044 Done |
| **07/01/2027** | D03 CV lưu/đọc thật; H01 và khoảng trống FR ưu tiên | Upload→chọn CV hiện tại→đọc đúng quyền; mở thông báo tới đúng đối tượng; demo toàn luồng với các vai trò | Storage/file validation cần chốt sớm. Không gọi metadata là file; notifications thêm cùng từng mutation từ trước, không đợi tuần này mới thiết kế atomicity. Nếu thiếu giờ ưu tiên ổn định luồng lõi, báo phần CV còn Partial |
| **14/01/2027** | Khép bàn giao giữa tháng1 | Chạy kịch bản xuyên app, đối chiếu FR Done/Partial, trình bày kết quả và phần còn lại | Dành mốc này cho sửa lỗi, kiểm bảo mật/quyền/DB/browser, dữ liệu demo và báo cáo; không khởi công một domain mới để tăng số màn |

Không có nghĩa phải làm hết D02 rồi mới D03, hoặc hết D04 mới D06. Migration, API, quyền, FE và bằng chứng được bàn giao theo feature nhỏ xuyên các miền cần thiết.

### Tiếp nối 21/11 — luồng báo cáo 26/11

Sau bàn giao OAP, người dùng đồng ý chuẩn bị proposal đề nghị hồ sơ → HR xét duyệt → kết quả/lịch sử, tiếp tục theo FR và bản đồ miền. Thiết kế tại [proposal hồ sơ](../proposals/backend-profile-change-approval.md); [SQL/test plan](../proposals/backend-profile-change-approval-sql-plan.md) giữ chi tiết chuyển đổi và kiểm chứng. Chưa có code/migration hoặc Jira mới cho luồng này.

Ưu tiên F1–F3 của proposal cho báo cáo 26/11: đề nghị/lịch sử chính chủ, HR quyết định, thông báo liên quan hồ sơ và reload đúng official. FR-012 HR sửa trực tiếp tách đầu ra F4, chỉ ghép khi đủ kiểm chứng; khóa/đổi quyền/vô hiệu hóa tài khoản không thuộc đợt này. Các lựa chọn mục4 đã được người dùng đồng ý sau review21/11; OpenAPI/SQL cụ thể và bộ ticket PCA đã soạn, chưa code/runtime hoặc nhập Jira. Notification legacy có dữ liệu phải dừng nếu chưa map được; không nhận bản thiết kế là FR Done. Giữ mốc 03/12 cho miền kỹ năng nếu luồng chính đã bàn giao; công thức phân công vẫn thảo luận riêng.

## 5. Các quyết định cần chốt đúng hạn

Chỉ dẫn tới điểm mở ở FR để tránh hai bản quy tắc:

| Thời điểm chậm nhất đề xuất | Đầu vào cần chốt | Ảnh hưởng |
|---|---|---|
| Trước code feature tạo tài khoản | O01: SQL bỏ username theo chốt cuối 15/11 và cơ chế cấp số theo chốt 15/11 (email login duy nhất, mã mới EMP000001), sinh/gửi mật khẩu tạm và contract lỗi sau commit (cấp lại/gửi lại đã hoãn), giới hạn và transaction phù hợp | Không hiển thị mật khẩu tạm trong admin UI. Gmail API/OAuth gmail.send đã duyệt; cần cấu hình hộp thư, runbook token Testing 7 ngày và kiểm mail thật được phép, không tự gửi thư/bật billing |
| Trước proposal D03/CV, mục tiêu 26/11 | O03/O05: storage, đọc/download riêng tư, giới hạn/cách giữ tệp | Có thể thiết kế adapter, không tự thêm dependency hoặc provider |
| Trước triển khai scorer/gợi ý; thời điểm thảo luận sẽ chốt sau | O02: công thức/weight/phiên bản/đánh giá; O08 chốt projection/batch/phân trang trong proposal liên quan | Không buộc hoàn thiện công thức để làm task hoặc phân công thủ công. PerfScore và dữ liệu tin cậy chưa chốt không tự đưa vào điểm |
| Trong từng proposal mutation | O07: phạm vi idempotency/audit/khóa và các ràng buộc thật | Làm khi feature cần, không xây framework chung cho toàn app trước |
| Trước demo ngoài máy local | Môi trường demo, HTTPS/proxy, dữ liệu và migration đích | Deploy/Neon có phê duyệt riêng sau rehearsal; lịch này không cấp quyền áp DB thật |

Thông báo H01 được đưa vào proposal của sự kiện liên quan từ đầu. Audit/idempotency/guard là tiêu chí trong feature, không task review riêng cho từng hàm. Không làm nguyên 33 bảng trước rồi mới viết API, cũng không sửa 000001 đã có.

## 6. Nhịp làm việc solo

- Thứ Năm: demo 3–5 phút bằng UI, nêu một luồng mới đã chạy thật, quyền/lỗi nổi bật, phần chưa đạt và mục tiêu tuần tới; Swagger chỉ mở khi cần giải thích contract.
- Sau báo cáo: chọn **một feature chính**, chốt brief/proposal gọn gồm phạm vi, dữ liệu/migration, API, quyền, transaction/sequence, test, WBS và ranh giới tự viết. Chỉ break Jira cho một đến hai tuần gần nhất khi thiết kế đủ rõ.
- Đầu tuần: triển khai xuyên DB→API→UI; ghép sớm, giữ prototype riêng. Agent hỗ trợ code/tests/docs; Quân đọc flow, review và tự thử. Phần Go tự viết chọn nhỏ, có hướng dẫn; không giao một API bảo mật khó trong ngày sát demo.
- Thứ Tư: đóng phạm vi demo, chạy cổng và regression các luồng có liên quan, chuẩn bị tài khoản/dữ liệu và rehearsal. Sau đó chỉ sửa lỗi chặn, không merge tính năng chưa kiểm để kịp báo cáo.
- Giới hạn WIP một feature chính; phần nghiên cứu độc lập có thể làm ngắn, không mở nhiều miền cùng lúc. Review từ sub-agent theo cách đã thống nhất, Quân vẫn là người đọc/chốt PR; không có nghĩa đã có người duyệt GitHub độc lập.
- Ghi giờ thực đơn giản theo feature và lỗi phát sinh, không tính lại review như feature riêng. Hai tuần liên tiếp trễ: bỏ bớt phần dự kiến tuần kế, nêu tác động tới mục tiêu 50/53 FR và báo cô sớm.

Ví dụ một tuần 18 giờ: khoảng 2 giờ chốt brief/đối chiếu, 9 giờ triển khai/đọc code cùng agent, 4 giờ test/review/sửa lỗi, 1 giờ chuẩn bị demo và 2 giờ dự phòng. Đây là cách chia thử; nếu chỉ có 15 giờ thì giảm đầu ra mới, không cắt kiểm quyền hoặc kiểm thử.

## 7. Bước tiếp theo được đề xuất

1. Hướng triển khai theo FR/bản đồ miền đã được đồng ý; tiếp tục hiệu chỉnh lịch/phạm vi theo tiến độ thực. Mục tiêu 80–85% chưa là con số đã đạt; công thức gợi ý chờ thảo luận riêng.
2. Người dùng đã đồng ý mở mục tiêu thành ba feature nối tiếp: danh mục tổ chức → Admin tạo tài khoản kèm hồ sơ → nhân viên login/đổi mật khẩu và xem hồ sơ thật. [Proposal đang review](../proposals/backend-organization-accounts-profile.md) giữ quyền HR/Admin riêng, không tạo tài khoản mồ côi hoặc tự cấp quyền từ vị trí; Gmail API/OAuth gmail.send đã chốt, cấp lại/gửi lại đã hoãn theo 15/11; cấu hình thực tế hoàn thiện ở task mail. Bộ [ticket OAP](../proposals/backend-organization-accounts-profile-tickets.md) đã soạn để review/nhập Jira, không là backlog cả chín tuần.
3. Từ proposal đã duyệt mới chốt task/Jira và phần triển khai; không tạo toàn bộ backlog 9 tuần từ bảng lịch này.
4. Khép DATN-35 và các ca auth browser còn mở bằng đúng kết quả mới; không dùng bảng lịch để đánh dấu E2E hoàn tất. Sau đó ưu tiên chốt proposal phần đề nghị hồ sơ/HR duyệt (FR-009–013 liên quan) hoặc kỹ năng thủ công (D03) theo nhu cầu báo cáo; chưa tự mở ticket/code miền mới. Công thức phân công tiếp tục chờ thảo luận riêng.

Đợt này chỉ đối chiếu repo và lập kế hoạch. Không chạy DB, đổi schema/nghiệp vụ, cấp credential, gửi email, tạo Jira, commit/push hoặc deploy.
