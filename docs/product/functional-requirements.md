# Yêu cầu chức năng và tiêu chí nghiệm thu — DATN

Ngày lập: **05/11/2026**. Trạng thái: **bản dự thảo đã chuẩn hóa sau review, đầu vào bản đồ miền và proposal backend**.

## 1. Mục tiêu và cách dùng

Đợt này kiểm kê yêu cầu đích từ UC, quyết định và workflow hiện hành, đặt ID để sau đó nối với thiết kế, API, kiểm thử và Jira. Không sửa nghiệp vụ, schema, công thức, mã nguồn hoặc chạy DB. Tiêu chí hoàn tất đợt tài liệu: mỗi FR có actor/điều kiện, AC quan sát được và nguồn; phân biệt chức năng với quy tắc/yêu cầu dùng chung; bao phủ UC hiện hành, giữ truy vết ID và các điểm chưa chốt.

FR mô tả hệ thống phải làm được gì; AC mô tả kết quả để nghiệm thu. Đây không phải danh sách test đã chạy hoặc xác nhận backend đã có chức năng. Một UC có thể chứa nhiều FR; một FR có thể cần nhiều API/task. Không suy một FR tương ứng một bảng hoặc một ticket.

- `FR` là chức năng, `BR` là quy tắc nghiệp vụ, `CR` là yêu cầu dùng chung xuyên chức năng. CR có thể bao gồm hành vi chức năng và ràng buộc toàn vẹn/bảo mật, không mặc định tất cả CR là yêu cầu phi chức năng.
- ID còn dùng được giữ khi chuyển nhóm; tách hành vi dùng ID mới, ID đã chuyển loại không tái sử dụng. FR không cần liên tục; bảng chuyển đổi ở mục 7 giữ dấu vết bản đầu. AC được dẫn bằng `FR-001.AC1`…; số AC trong các mục đã tách được ánh xạ riêng ở mục 7.
- Mức phân rã theo một mục đích nghiệp vụ với actor/điều kiện/kết quả nhất quán. Các nhánh cùng quyết định như chấp thuận/từ chối, hoặc cùng vòng đời như Start/Pause/Resume có thể ở cùng FR; không tách mọi nút, trường hoặc validation thành FR.
- Hiện có **62 FR đang dùng, 1 BR và 2 CR**. Ba ID FR cũ chuyển loại, ba ID mới tách hành vi cũ; không thêm phạm vi sản phẩm hoặc ép số FR xuống một ngưỡng định trước.
- Các nhóm dưới đây chỉ để đọc và kiểm kê, **chưa chốt miền/package Go, thứ tự triển khai hoặc phạm vi sprint**.
- Mặc định các thao tác cần tài khoản/hồ sơ đang hoạt động và quyền hiện hành. Actor ghi ở mỗi FR không thay kiểm tra quyền trên đúng tài nguyên; người kiêm nhiệm dùng từng quyền tương ứng.
- AC là điều kiện tối thiểu cho feature; khi viết proposal phải dẫn thêm ma trận trạng thái/quyền tại nguồn, không coi AC ngắn là cách bỏ điều kiện trong workflow.
- FR liên quan điểm mở vẫn là mục tiêu sản phẩm; phần bị ảnh hưởng **chưa sẵn sàng để code** trước khi chốt ở mục 5. FR có nhiều phần chỉ bị chặn phần liên quan điểm mở, không tự hoãn cả nhóm.
- Nguồn ưu tiên theo [mục lục](../README.md). Nếu phát hiện mâu thuẫn, ghi nhận để xác nhận; bản FR không tự ghi đè quyết định đã chốt.

## 2. Nguồn truy vết

| Mã | Nguồn và phần áp dụng |
|---|---|
| S01 | [Phạm vi sản phẩm](../../.claude/docs/product_scope.md), mục 1–3 và Trách nhiệm theo dự án: actor, UC, giới hạn |
| S02 | [Vai trò và phân quyền](../../.claude/docs/roles_permissions.md), mục 1–4, R06: quyền hệ thống/dự án, góc nhìn dữ liệu, session |
| S03 | [Quy tắc miền](../../.claude/docs/domain_rules.md), mục 1–5: bất biến, trạng thái, kỹ năng, effort/thông báo |
| S04 | [Luồng dự án và phân công](../../.claude/docs/project_task_assignment_workflow.md), UC001–UC004, Thay người/thu hồi |
| S05 | [Gợi ý nhân sự](../../.claude/docs/ai_recommendation_workflow.md), mục 1, 3, 4 và Contract lưu trữ; mục 2 chỉ là phương án công thức thử nghiệm |
| S06 | [Nhật ký và hiệu chỉnh](../../.claude/docs/work_effort_workflow.md), mục 1–7, đặc biệt R01 ngày 21/10 |
| S07 | [Hồ sơ nhân sự](../../.claude/docs/employee_profile_workflow.md), mục 1–5, Địa chỉ khi hoàn thiện hồ sơ |
| S08 | [Phê duyệt kỹ năng](../../.claude/docs/skill_approval_workflow.md), chốt 27/10 và mục 1–8 |
| S09 | [CV và trích xuất](../../.claude/docs/cv_extraction_approval_workflow.md), mục 1–4 và R03 |
| S10 | [Chốt bổ sung 20/10–06/11](../business-baseline-2026-10-20.md), địa chỉ, nhật ký, cấp mã, thay PM/hạn task/trao đổi |
| S11 | [Sổ quyết định DB](../design/database-review-decisions-2026-10-14.md), Quyết định hiện hành và Bàn giao; [dictionary lõi](../design/database-step5a-core-dictionary-2026-10-14.md), [dictionary hỗ trợ](../design/database-step5b-supporting-dictionary-2026-10-14.md): ràng buộc thiết kế đã chốt, không chứng minh đã có migration |
| S12 | [Bảo mật](../../.claude/docs/security.md), mục 1–5: session, dữ liệu riêng tư, upload và bảo vệ ghi |
| S13 | [Contract màn hình](../design/figma-make/screen-contracts.md), UI-01–UI-15 và tiếp nối 28–29/10: góc nhìn UI, thông báo, phiên mô phỏng |
| S14 | [Frontend hiện trạng/sổ công việc](../../web/README.md): bằng chứng phạm vi mô phỏng, không là nguồn cấp thêm quyền hoặc công thức |

## 3. Danh mục FR/AC

### A. Truy cập và phiên

#### FR-001 — Đăng nhập và xác định quyền sử dụng

**Actor/điều kiện:** người có tài khoản; đăng nhập bằng email công ty theo chốt 05/11; chuẩn hóa email đã chốt A02 ngày 06/11, cấp thông tin/reset còn O01. **Nguồn:** S02 mục 2, 4; S12 mục 2, 4; [chốt 05/11](../business-baseline-2026-10-20.md#quyết-định-tiếp-nối-05112026--định-danh-đăng-nhập-backend).

- **AC1:** thông tin xác thực hợp lệ, user/employee hoạt động mới mở được phiên; sai thông tin trả thông báo chung, không tiết lộ tài khoản tồn tại.
- **AC2:** chưa có phiên, phiên hết hạn hoặc đã thu hồi không truy cập được chức năng bảo vệ; biết URL/ID hoặc chọn danh tính demo không tạo quyền server.
- **AC3:** quyền hệ thống và trách nhiệm PM/PTCM được xác định từ dữ liệu hiện hành; vị trí/cấp bậc không cấp quyền và Admin không mặc nhiên có quyền HR/Giám đốc/dự án.
- **AC4:** tài khoản dùng mật khẩu tạm phải đổi mật khẩu trước khi được truy cập nghiệp vụ theo chốt 06/11 tại S10; xác thực mật khẩu tạm không cấp quyền nghiệp vụ đầy đủ.
- **AC5:** email login được trim đầu/cuối và so không phân biệt hoa/thường, nhiều kết quả khớp thì dừng; phiên hết hạn tuyệt đối sau 8 giờ từ login, không gia hạn do hoạt động, gồm phiên hạn chế. Không tự viết lại email đã lưu.

#### FR-002 — Đăng xuất

**Actor/điều kiện:** chủ phiên đang đăng nhập. **Nguồn:** S02 mục 4; S12 mục 2.

- **AC1:** đăng xuất thu hồi phiên hiện tại; dùng lại thông tin phiên đó không vào được nghiệp vụ.
- **AC2:** đăng xuất không xóa hồ sơ, công việc, nhật ký hoặc quyết định đã lưu; không tự thu hồi các phiên khác như thao tác khóa tài khoản.

#### FR-003 — Thu hồi quyền truy cập khi tài khoản/quyền thay đổi

**Actor/điều kiện:** hệ thống sau thao tác quản trị hợp lệ; người dùng ở lần truy cập tiếp theo. **Nguồn:** S02 mục 2–4; S12 mục 2–3.

- **AC1:** khóa tài khoản, đổi quyền hệ thống hoặc reset mật khẩu thu hồi các phiên theo chính sách; user/employee không còn hoạt động bị từ chối sử dụng nghiệp vụ.
- **AC2:** thay PM/PTCM hoặc kết thúc phân công làm mất đúng quyền liên quan ở lần truy cập sau; không xóa quyền khác còn hợp lệ hoặc hứa hủy ngược giao dịch đã commit.

#### FR-004 — Đổi mật khẩu của mình

**Actor/điều kiện:** chủ tài khoản có phiên hợp lệ; chính sách mật khẩu mới A09 đã chốt, reset/cấp lại còn O01. **Nguồn:** S12 mục 2, 4; S02 mục 3.

- **AC1:** xác minh mật khẩu cũ trước khi cho thay mật khẩu của mình; sai mật khẩu cũ hoặc không đạt chính sách thì không thay đổi thông tin xác thực.
- **AC2:** lưu thành công theo chính sách đã chốt, không trả/log mật khẩu hoặc hash; xử lý hiệu lực phiên theo S12 và FR-003. Không mặc nhiên có quyền reset tài khoản người khác.
- **AC3:** khi cấp tài khoản mới bằng mật khẩu tạm, bắt buộc đổi trước nghiệp vụ; đổi thành công làm mật khẩu tạm mất hiệu lực, sau đó đăng nhập bằng mật khẩu mới. Mật khẩu tạm chưa có hạn; cờ users.must_change_password và contract phiên hạn chế A08 đã chốt; thay hash + false + revoke mọi phiên trong cùng transaction. Đợt auth đầu chỉ hiện thực phần đổi lần đầu, không tự đánh dấu hoàn tất đổi tự nguyện.
- **AC4:** mật khẩu mới lần đầu đạt tối thiểu 15 code point Unicode, tối đa 72 byte UTF-8, khác mật khẩu hiện tại; không trim/cắt/ép nhóm ký tự, hỗ trợ khoảng trắng/chữ có dấu/paste. Không áp chính sách đặt mật khẩu mới ngược lên login tài khoản cũ. Limiter và lỗi theo S12.

#### FR-063 — Admin đặt lại mật khẩu tài khoản

**Actor/điều kiện:** Admin; chính sách reset và kênh cấp thông tin xác thực cần chốt O01. **Nguồn:** S12 mục 2, 4; S02 mục 3.

- **AC1:** chỉ Admin có quyền reset đúng tài khoản; không dùng luồng đổi mật khẩu chính chủ hoặc tự cho người dùng reset tài khoản khác. Phương thức cấp thông tin mới phải được chốt trước khi triển khai, không dùng mật khẩu seed/demo làm chính sách.
- **AC2:** reset thành công thu hồi các phiên theo FR-003, không đổi quyền hệ thống hoặc trách nhiệm dự án; không trả/log mật khẩu rõ hoặc hash trong giao diện nghiệp vụ. Chưa chốt thì giữ là yêu cầu đích, chưa coi sẵn sàng code.

### B. Tài khoản và quyền hệ thống

#### FR-005 — Tạo tài khoản kèm hồ sơ nhân viên ban đầu

**Actor/điều kiện:** Admin; có danh mục vị trí hợp lệ. **Nguồn:** S10 Quyết định đã xác nhận và chốt cấp mã 28/10; S02 mục 3; S11 nhóm 1.

- **AC1:** tạo cùng lúc tài khoản và hồ sơ với họ tên, email công ty, vị trí và các dữ liệu bắt buộc hiện hành; sai dữ liệu/lưu thất bại không để lại tài khoản hoặc hồ sơ dở dang.
- **AC2:** điện thoại/phòng ban/cấp bậc có thể chưa bổ sung; không bắt Admin nhập địa chỉ, không thêm bước HR duyệt tạo. Sau tạo HR quản lý hồ sơ, Admin không được mặc nhiên sửa hành chính thường xuyên.
- **AC3:** không có quyền PM/PTCM tự phát sinh từ tạo tài khoản. Chính sách username còn O01.
- **AC4:** theo chốt 06/11 tại S10, hệ thống sinh mật khẩu tạm ngẫu nhiên và gửi qua email đăng ký; chưa có thời hạn trong phiên bản hiện tại, DB chỉ lưu bản băm, không log/hiển thị mật khẩu trong quản trị. Gửi lỗi/gửi lại còn O01; chưa chốt transaction gửi thư.

#### FR-006 — Tự cấp mã nhân viên và hiển thị định danh được cấp

**Actor/điều kiện:** hệ thống khi Admin tạo hồ sơ. **Nguồn:** S10 chốt cấp mã 28/10.

- **AC1:** Admin xem mã dự kiến chỉ đọc; sau lưu thấy mã thực được cấp duy nhất. Hai lần tạo đồng thời không cấp trùng; không tái dùng mã đã khóa/vô hiệu hóa hoặc đổi mã hồ sơ cũ.
- **AC2:** không dựa số lượng hồ sơ cộng một để cam kết cấp mã; `e001` theo mã nhân viên chỉ là username demo, không tự áp thành quy tắc backend.

#### FR-007 — Tra cứu danh sách tài khoản

**Actor/điều kiện:** Admin. **Nguồn:** S02 mục 1, 3; S13 UI-01.

- **AC1:** xem/tìm tài khoản, người tương ứng, trạng thái và các quyền hệ thống hiện hành; không trả mật khẩu/hash/bí mật xác thực.
- **AC2:** xem danh sách không tự cho phép sửa hồ sơ HR, xử lý kỹ năng hoặc điều hành dự án; các thao tác quản trị tiếp tục kiểm tra quyền riêng.

#### FR-064 — Cấp và thu hồi quyền hệ thống

**Actor/điều kiện:** Admin. **Nguồn:** S02 mục 1, 3; S03 mục 2; S13 UI-01.

- **AC1:** cấp/thu hồi kết hợp ADMIN/HR/DIRECTOR; không chọn quyền đặc biệt vẫn có tư cách nhân viên từ hồ sơ. Thêm quyền không bị chặn chỉ vì người đó đang có công việc.
- **AC2:** không tự thay PM/PTCM/cơ cấu công tác; chặn thu hồi nếu làm mất quyền xử lý trách nhiệm đang mở theo nguồn, không hạ quyền Admin hoạt động cuối cùng. Đổi quyền thành công làm mất hiệu lực các phiên theo FR-003.

#### FR-008 — Khóa và mở khóa tài khoản

**Actor/điều kiện:** Admin. **Nguồn:** S02 mục 3; S03 mục 2; S13 UI-01.

- **AC1:** khóa khẩn cấp được phép dù còn công việc, thu hồi truy cập và chỉ ra trách nhiệm cần xử lý; không tự giao việc hoặc đổi người phụ trách.
- **AC2:** không khóa/vô hiệu hóa Admin hoạt động cuối cùng. Mở khóa không tự kích hoạt hồ sơ đã bị HR vô hiệu hóa hoặc phục hồi phân công cũ.

### C. Hồ sơ cá nhân và nhân sự

#### FR-009 — Xem hồ sơ cá nhân và lịch sử đề nghị

**Actor/điều kiện:** chủ hồ sơ. **Nguồn:** S07 mục 1–2; S13 UI-03.

- **AC1:** xem dữ liệu chính thức, thông tin công tác và lịch sử đề nghị của mình; yêu cầu đang chờ/từ chối không hiển thị như đã đổi hồ sơ.
- **AC2:** email công ty, phòng ban, vị trí và cấp bậc chỉ đọc ở luồng cá nhân; không mở dữ liệu người khác khi đổi ID. Địa chỉ chưa có hiển thị/nhắc theo BR-001.

#### FR-010 — Gửi đề nghị cập nhật thông tin cá nhân

**Actor/điều kiện:** chủ hồ sơ. **Nguồn:** S07 mục 2; S10 quyết định địa chỉ; S13 UI-03.

- **AC1:** gửi phần thay đổi fullname/phone/address; không nhận role, trạng thái, cơ cấu, kỹ năng hoặc trách nhiệm dự án. Lưu đề nghị riêng, chưa đổi dữ liệu chính thức.
- **AC2:** hủy form không tạo đề nghị; xem được nội dung và kết quả sau gửi. Không thêm rút/bổ sung pending hoặc hạn mức gửi từ quy trình kỹ năng. Bổ sung/sửa địa chỉ phải đạt BR-001.

#### FR-011 — HR xét duyệt đề nghị thông tin cá nhân

**Actor/điều kiện:** HR; đề nghị còn chờ. **Nguồn:** S07 mục 2; S13 UI-04.

- **AC1:** xem giá trị trước/đề nghị/hiện tại; chấp thuận chỉ cập nhật đúng trường được duyệt và ghi người/thời điểm. Dữ liệu nền đã đổi thì không ghi đè từ màn hình cũ.
- **AC2:** từ chối có lý do, giữ nguyên hồ sơ; đề nghị đã kết thúc không xử lý lại. Giám đốc/Admin không có quyền duyệt chỉ từ quyền hệ thống của họ. Duyệt địa chỉ áp dụng BR-001.

#### FR-012 — HR xem và cập nhật hồ sơ trực tiếp

**Actor/điều kiện:** HR. **Nguồn:** S07 mục 1, 3, 5; S13 UI-02.

- **AC1:** xem/tìm hồ sơ và cập nhật hành chính, email công ty, cơ cấu; vị trí/phòng ban/cấp bậc chọn từ danh mục hợp lệ, cấp bậc có thể Chưa cập nhật.
- **AC2:** lưu lịch sử đúng actor; không giả là đề nghị của nhân viên, không cấp quyền hệ thống hoặc sửa kỹ năng chính thức bằng thao tác hồ sơ. Cập nhật địa chỉ áp dụng cùng BR-001 như luồng đề nghị.

#### FR-013 — Vô hiệu hóa hồ sơ nhân sự

**Actor/điều kiện:** HR; đáp ứng điều kiện trách nhiệm đang mở. **Nguồn:** S07 mục 3; S03 mục 2.

- **AC1:** còn task/trách nhiệm dự án mở thì chặn và hiển thị các mục phải xử lý trước; không tự thay PM/PTCM/phân công.
- **AC2:** khi hợp lệ, ngăn sử dụng nghiệp vụ và giữ lịch sử công việc/năng lực. Không đồng nhất với khóa tài khoản của Admin hoặc tự mở lại trạng thái tài khoản.

### D. Danh mục và năng lực được công nhận

#### FR-015 — Quản lý phòng ban, vị trí và cấp bậc

**Actor/điều kiện:** HR. **Nguồn:** S07 mục 5; S11 chốt 26/10, nhóm 1; S13 UI-08.

- **AC1:** xem/thêm/sửa/xóa tên danh mục hợp lệ; tên bắt buộc, không trùng theo ràng buộc danh mục; chặn xóa khi còn tham chiếu.
- **AC2:** cấp bậc do công ty cấu hình, không enum IT cố định hoặc nhập tự do trên hồ sơ. Đổi tên giữ định danh/liên kết và lịch sử đã chụp; không suy quyền, mức kỹ năng, thứ tự hay hệ số từ cấp bậc/vị trí.

#### FR-016 — Quản lý nhóm kỹ năng và kỹ năng

**Actor/điều kiện:** Giám đốc. **Nguồn:** S08 mục 8; S11 nhóm 2; S13 UI-09.

- **AC1:** xem/thêm/sửa/xóa nhóm và kỹ năng thuộc nhóm theo ràng buộc tên/liên kết; không xóa mục còn tham chiếu hoặc đổi nghĩa kỹ năng đã sử dụng.
- **AC2:** không thêm ACTIVE/INACTIVE; CRUD danh mục không công nhận kỹ năng cho nhân viên và không cấp quyền này cho HR/Admin mặc định.

#### FR-017 — Xem năng lực chính thức tách khỏi đề nghị

**Actor/điều kiện:** chủ hồ sơ hoặc người có quyền đọc theo mục đích. **Nguồn:** S02 R06; S03 mục 2; S08 mục 3, 8.

- **AC1:** hiển thị năng lực đang được công nhận với bốn mức BASIC/INTERMEDIATE/ADVANCED/EXPERT, không lấy pending/CV chưa duyệt làm năng lực chính thức.
- **AC2:** không suy mức từ số năm, cấp bậc hoặc task đã nghiệm thu. Thay đổi năng lực chính thức không tự dừng task đã giao; lần tra cứu/gợi ý/giao việc sau dùng dữ liệu hiện hành.

### E. CV và xử lý trích xuất

#### FR-018 — Bổ sung CV cá nhân hoặc HR hỗ trợ

**Actor/điều kiện:** chủ hồ sơ; HR bổ sung thay đúng nhân viên. **Nguồn:** S09 mục 1; S12 mục 5; S13 UI-05.

- **AC1:** tiếp nhận PDF có text/DOCX tối đa 10 MiB, kiểm tra nội dung thật; không hỗ trợ OCR/PDF mã hóa. Không đạt thì không lưu thành phiên bản hợp lệ.
- **AC2:** HR hỗ trợ có actor đúng, không phải cấp duyệt. Tệp lưu riêng tư; thêm CV không tự thay bản hiện tại, không đổi kỹ năng. Lưu tệp thật còn O03.

#### FR-019 — Chọn CV hiện tại và xem phiên bản

**Actor/điều kiện:** chủ hồ sơ có CV hợp lệ. **Nguồn:** S09 mục 3–4.

- **AC1:** chủ hồ sơ xác nhận bản đang dùng, tối đa một bản hiện tại; giữ các bản cũ. HR không xác nhận thay khi chưa có quyết định mở quyền.
- **AC2:** tệp hợp lệ vẫn chọn được nếu trích xuất lỗi; thay CV không xóa kỹ năng, rút yêu cầu hoặc sửa quyết định cũ; không có phê duyệt/từ chối công bố tệp CV.

#### FR-020 — Đọc và tải CV theo quyền

**Actor/điều kiện:** chủ hồ sơ, HR hoặc Giám đốc theo mục đích căn cứ. **Nguồn:** S09 mục 3; S02 R06; S12 mục 5.

- **AC1:** đọc/tải kiểm tra quyền trên đúng tệp, không công khai storage key thành quyền truy cập; PM/PTCM không đọc CV gốc người khác từ tra cứu nghề nghiệp.
- **AC2:** biết ID/URL hoặc từng có quyền không đủ để tải khi quyền hiện hành không cho phép; không trả nội dung tệp trái quyền.

#### FR-021 — Trích xuất tên kỹ năng và đoạn căn cứ từ CV

**Actor/điều kiện:** hệ thống xử lý CV hợp lệ; cấu hình xử lý còn O03–O04. **Nguồn:** S09 mục 1–2, 4.

- **AC1:** thể hiện QUEUED/PROCESSING/SUCCEEDED/FAILED, trả tên/đoạn căn cứ và khả năng đối chiếu danh mục; không tự suy mức/số năm hoặc công nhận năng lực.
- **AC2:** retry không tạo mục trùng; lỗi không làm mất CV hợp lệ. Không nhận diện kỹ năng vẫn có thể xử lý thành công và báo Không có kỹ năng để cập nhật.

#### FR-022 — Giám đốc đối chiếu và xét kỹ năng từ CV

**Actor/điều kiện:** Giám đốc; kết quả trích xuất thành công, không tự duyệt. **Nguồn:** S09 mục 2–3/R03; S08 mục 3–7.

- **AC1:** xem nguồn, map/hợp nhất kỹ năng trùng và xác định mức/căn cứ trước công nhận; thiếu mapping/mức không được duyệt. Không bắt nhân viên gửi thủ công từng mục CV; pending cùng người/kỹ năng phải giải quyết xung đột.
- **AC2:** từng mục chỉ có một kết quả kết thúc, không mở lại; kết quả tổng hợp suy đúng từ các mục, tách APPROVED/PARTIALLY_APPROVED/REJECTED/NO_CHANGES khỏi trạng thái tệp và trích xuất.
- **AC3:** NO_CHANGE/WITHDRAWN không đếm là được duyệt; CV mới không xóa official hoặc sửa kết quả đợt cũ. Bổ sung sau kết thúc bằng yêu cầu mới.

### F. Đề nghị và xét duyệt kỹ năng chủ động

#### FR-023 — Gửi đề nghị bổ sung hoặc điều chỉnh kỹ năng

**Actor/điều kiện:** chủ hồ sơ. **Nguồn:** S08 chốt 27/10, mục 2, 6–7; S13 UI-06.

- **AC1:** chọn kỹ năng, mức và thông tin/giải trình phù hợp; hệ thống suy ADD khi chưa có official, UPDATE khi đã có. Không tạo đề nghị gỡ hoặc thêm quyền gỡ trực tiếp cho Giám đốc/HR.
- **AC2:** tối đa một pending/người/kỹ năng dùng chung MANUAL/CV; pending không sửa official. Không hạn ngạch tuần/tháng, không bắt có CV mới hoặc task trong hệ thống.

#### FR-024 — Dẫn căn cứ cho đề nghị kỹ năng

**Actor/điều kiện:** chủ hồ sơ; dữ liệu căn cứ nằm trong quyền đọc. **Nguồn:** S08 mục 2, 5, 8; S10 tiếp nối 28/10.

- **AC1:** dẫn 0/1/nhiều đóng góp công việc, giải trình/căn cứ ngoài hệ thống; không bắt có tệp nhưng vẫn phải giải trình phù hợp. Không lấy kết quả người nhận sau làm đóng góp của người trước.
- **AC2:** cho tối đa một CV dẫn trực tiếp và các tệp minh chứng tùy chọn theo nguồn; không trộn tệp chứng chỉ vào CV. Lưu/đọc tệp thật, giới hạn minh chứng còn O03/O05, không coi FE metadata là đã lưu file.

#### FR-025 — Bổ sung căn cứ hoặc rút đề nghị kỹ năng đang chờ

**Actor/điều kiện:** chủ đề nghị còn PENDING. **Nguồn:** S08 mục 6–8; S03 mục 2; S13 UI-06.

- **AC1:** bổ sung căn cứ khi pending, giữ yêu cầu và ghi nhận lần thay đổi căn cứ; màn hình duyệt cũ không được quyết định trên căn cứ đã thay đổi.
- **AC2:** rút chuyển WITHDRAWN, không đổi official; đề nghị đã kết thúc không sửa/bổ sung/mở lại. Sau từ chối có thể gửi mới với căn cứ làm rõ.

#### FR-026 — Cung cấp cơ sở cho Giám đốc đánh giá kỹ năng

**Actor/điều kiện:** Giám đốc có quyền xử lý đề nghị. **Nguồn:** S08 mục 3–5; S13 UI-07.

- **AC1:** hiển thị mã/vị trí/ngày gửi, trước/đề nghị, mô tả bốn mức, giải trình, lịch sử cùng kỹ năng và căn cứ đóng góp/nộp/nghiệm thu; số năm/chứng chỉ/điểm gợi ý không tự quyết định mức.
- **AC2:** đọc đúng task/căn cứ dẫn chiếu, không mở toàn dự án/log/bình luận hoặc địa chỉ riêng tư. Khi thiếu chuyên môn, ghi nguồn/nội dung ý kiến phù hợp; không giả người góp ý đã đăng nhập/ký duyệt, không thêm cấp phê duyệt bắt buộc.

#### FR-027 — Chấp thuận hoặc từ chối đề nghị kỹ năng

**Actor/điều kiện:** Giám đốc; pending, không phải chính mình. **Nguồn:** S08 mục 1, 6–7; S03 mục 2.

- **AC1:** cả chấp thuận/từ chối đều có lý do; chấp thuận ghi quyết định và năng lực chính thức cùng lần lưu; từ chối giữ official cũ.
- **AC2:** kiểm tra phiên bản yêu cầu và giá trị official nền; stale không ghi đè. NO_CHANGE chỉ khi toàn bộ giá trị nghiệp vụ giống official còn hiệu lực, không được coi là một lần cập nhật đã duyệt.

#### FR-028 — Tra cứu quyết định kỹ năng và giữ lịch sử

**Actor/điều kiện:** chính chủ hoặc Giám đốc theo quyền. **Nguồn:** S08 mục 6–8; S09 R03.

- **AC1:** xem quyết định, người/thời điểm/lý do và căn cứ của từng yêu cầu, tách đề nghị mới khỏi quyết định cũ.
- **AC2:** quyết định kết thúc bất biến, không mở lại mục CV cũ hoặc tự tái kích hoạt kỹ năng; trường hợp Giám đốc chỉ có một người tự đề nghị không có ngoại lệ tự duyệt (O09).

### G. Dự án và trách nhiệm dự án

#### FR-029 — Giám đốc tạo dự án kèm PM

**Actor/điều kiện:** Giám đốc. **Nguồn:** S04 UC001; S03 mục 3.

- **AC1:** mã/tên/PM/ngày bắt đầu/ngày kết thúc/ưu tiên bắt buộc; mã duy nhất, ngày đúng thứ tự; tạo PLANNING cùng PM hợp lệ trong một lần lưu.
- **AC2:** thất bại/hủy không để dự án dở; người tạo không tự thành PM/PTCM; không tạo roster hoặc luồng duyệt kế hoạch.

#### FR-030 — Giám đốc thay PM

**Actor/điều kiện:** Giám đốc; project PLANNING/IN_PROGRESS/COMPLETED. **Nguồn:** S04 UC001; S10 tiếp nối 28/10; S02 mục 3.

- **AC1:** chọn PM mới hợp lệ và lưu một lần; thành công thay đúng quyền PM, thất bại giữ PM cũ. Không chờ PM cũ duyệt hoặc xóa PM trước.
- **AC2:** giữ lịch sử, trạng thái, PTCM và phân công/đóng góp; các quyền kiêm nhiệm xét độc lập. Không đổi PM ở CLOSED/CANCELLED.

#### FR-031 — PM chỉ định hoặc thay PTCM

**Actor/điều kiện:** PM dự án; project PLANNING/IN_PROGRESS theo S04/S13. **Nguồn:** S04 UC001; S03 R8–R9; S02 mục 1.

- **AC1:** chọn người hợp lệ, có thể PM kiêm nhiệm; trách nhiệm chỉ thuộc dự án đó, không là quyền hệ thống hoặc cho phép nhận task trong cùng dự án.
- **AC2:** chặn người còn task mở hoặc xung đột với đầu ra cũ cần nghiệm thu trong dự án; không tự thu hồi task. Lưu lịch sử và giữ actor/đóng góp cũ.

#### FR-032 — Xem và theo dõi dự án theo trách nhiệm

**Actor/điều kiện:** người có quyền đọc dự án tương ứng. **Nguồn:** S02 mục 1/R06; S04 activity/UC001; S03 mục 3; S13 UI-10.

- **AC1:** PM/PTCM xem thông tin, trách nhiệm, nhóm/task, trạng thái và tiến độ trong quyền; PM-only không có thao tác cấu trúc/phân công/nghiệm thu, PTCM-only không có thao tác vòng đời PM.
- **AC2:** tiến độ bằng DONE/CLOSED trên task không CANCELLED; không có task hiệu lực thì hiển thị chưa có dữ liệu, không tự 100%/hoàn thành. Người chỉ được giao một task không tự xem toàn dự án.

#### FR-033 — PM bắt đầu triển khai dự án

**Actor/điều kiện:** PM; project PLANNING. **Nguồn:** S03 mục 3; S04 UC001.

- **AC1:** có ít nhất một task ASSIGNED mới chuyển IN_PROGRESS và ghi lịch sử/thời điểm.
- **AC2:** chưa bắt đầu dự án thì nhân viên không Start/log/thực hiện; phân công trong giai đoạn lập kế hoạch không tự bắt đầu dự án hoặc sinh effort.

#### FR-034 — PM xác nhận hoàn thành dự án

**Actor/điều kiện:** PM; project IN_PROGRESS. **Nguồn:** S03 mục 3; S04 UC001; S06 mục 6.

- **AC1:** chỉ chuyển COMPLETED khi có ít nhất một DONE và mọi task còn lại DONE/CANCELLED; nghiệm thu một task không tự hoàn thành dự án.
- **AC2:** COMPLETED giữ đọc/hiệu chỉnh effort được duyệt, không tiếp tục task mới/cũ, phân công, nghiệm thu thêm hoặc trao đổi mới.

#### FR-035 — PM đóng dự án

**Actor/điều kiện:** PM; project COMPLETED, không còn yêu cầu effort PENDING. **Nguồn:** S03 mục 3; S04 UC001; S06 mục 6.

- **AC1:** còn yêu cầu hiệu chỉnh chờ thì chặn; hợp lệ chuyển project CLOSED và task DONE thành CLOSED trong cùng lần lưu.
- **AC2:** sau đóng dữ liệu nghiệp vụ chỉ đọc theo quyền, không mở lại hoặc duyệt effort; vẫn đánh dấu thông báo của mình đã đọc được.

#### FR-036 — PM hủy dự án

**Actor/điều kiện:** PM; project PLANNING/IN_PROGRESS, có lý do. **Nguồn:** S03 mục 3; S04 UC001; S06 mục 6; S11 dictionary lõi mục submission.

- **AC1:** hủy task mở, giữ DONE, effort, lịch sử; kết thúc yêu cầu effort pending theo lý do hủy dự án, không tự duyệt đề nghị.
- **AC2:** lần nộp chưa kết luận được giữ, không giả nghiệm thu không đạt; ra khỏi hàng chờ theo trạng thái task/project. Dự án hủy không mở lại/sửa nghiệp vụ.

### H. Cấu trúc và đặc tả công việc

#### FR-037 — Quản lý nhóm công việc trong dự án

**Actor/điều kiện:** PTCM; project PLANNING/IN_PROGRESS. **Nguồn:** S04 UC002; S03 mục 3; S13 UI-11.

- **AC1:** tạo/sửa nhóm theo ràng buộc tên duy nhất trong project; nhóm có thể trống khi lập kế hoạch, không phân công trực tiếp cho group.
- **AC2:** mỗi task thuộc group cùng project; chuyển nhóm chỉ trong dự án và task chưa kết thúc, có lịch sử. Không cây task nhiều cấp, chuyển task sang project khác hoặc xóa liên kết/lịch sử đang dùng.

#### FR-038 — Tạo task và yêu cầu kỹ năng

**Actor/điều kiện:** PTCM; project PLANNING/IN_PROGRESS. **Nguồn:** S04 UC002; S03 mục 3; S13 UI-11.

- **AC1:** task có mã duy nhất/project, tên, group, mô tả gồm đầu ra, tiêu chí nghiệm thu riêng, estimate dương, ưu tiên theo đặc tả và ít nhất một requirement; tạo task/requirements cùng lần lưu, trạng thái PENDING_ASSIGNMENT.
- **AC2:** requirement không trùng skill, dùng bốn mức, mandatory và weight do PTCM đặt. Không deliverable riêng hoặc chốt miền weight 1–5 khi còn O02. Ngày tùy chọn phải đủ cặp, đúng thứ tự và nằm trong kế hoạch project.

#### FR-039 — Sửa đặc tả trước phân công và bảo vệ phạm vi đã giao

**Actor/điều kiện:** PTCM; project mở, task đáp ứng điều kiện sửa. **Nguồn:** S04 UC002; S03 mục 3; S10 tiếp nối 28/10.

- **AC1:** trước lần giao đầu, sửa đặc tả phải qua cùng validation như tạo; đã giao lần đầu thì đóng băng scope/đầu ra/estimate/requirements/criteria, kể cả sau thu hồi.
- **AC2:** không dùng sửa mô tả để đổi scope; scope mới cần task mới, hủy task cũ khi phù hợp. Làm rõ không đổi scope có lịch sử là nhu cầu đã ghi nhưng cách thể hiện còn O06; chưa mở lại toàn bộ form sửa.

#### FR-040 — Đặt kế hoạch đầu hoặc gia hạn task

**Actor/điều kiện:** PTCM; project PLANNING/IN_PROGRESS và task chưa kết thúc. **Nguồn:** S10 tiếp nối 28/10; S03 mục 3, 5.

- **AC1:** chưa có ngày thì đặt cặp hợp lệ đầu tiên làm mốc gốc; đã có thì chỉ gia hạn kết thúc sau hạn hiện hành, không đổi ngày bắt đầu đã lưu, trong ngày dự án; lý do/lịch sử trước–sau bắt buộc.
- **AC2:** giữ mốc gốc, estimate và frozen spec; không dựng luồng đổi kế hoạch dự án. Quá hạn kế hoạch không tự chặn nhật ký thực tế hợp lệ.

#### FR-041 — PM hủy task

**Actor/điều kiện:** PM; project PLANNING/IN_PROGRESS, task chưa DONE/CLOSED/CANCELLED. **Nguồn:** S03 mục 3; S11 dictionary lõi phần hủy khi review.

- **AC1:** có lý do, chuyển CANCELLED/kết thúc lượt giao, giữ log/đầu ra/lịch sử và người cuối để tra cứu; không xóa task hoặc coi effort bằng 0.
- **AC2:** task đang review được loại khỏi hàng chờ, giữ lần nộp chưa kết luận; không giả REJECTED, không mở lại task terminal.

### I. Tra cứu và phân công

#### FR-042 — Tra cứu nhân sự nghề nghiệp độc lập

**Actor/điều kiện:** PM hoặc người hiện có trách nhiệm PTCM. **Nguồn:** S02 R06; S01 mục 2; S13 UI-13.

- **AC1:** xem/tìm nhân sự theo dữ liệu nghề nghiệp được phép, kỹ năng chính thức, kinh nghiệm/kết quả và tải tổng hợp; không bắt lập đội dự án trước.
- **AC2:** không trả địa chỉ/liên hệ riêng tư/CV gốc/raw task/log/căn cứ dự án khác; không cấp quyền phân công từ quyền xem danh sách hoặc mở cho nhân viên thường. Giới hạn địa chỉ áp dụng BR-001.

#### FR-043 — Tra cứu và chọn người thủ công theo task

**Actor/điều kiện:** PTCM; task chờ phân công, project PLANNING/IN_PROGRESS. **Nguồn:** S04 UC003.2; S05 mục 1.

- **AC1:** tìm/lọc nguồn toàn công ty; chỉ người có user/employee active, không là PTCM cùng dự án. Không sinh điểm/rank/run khi chưa chạy engine.
- **AC2:** thiếu/dưới mandatory cảnh báo và yêu cầu lý do override; không override quyền/trạng thái/loại PTCM. Tải cao là cảnh báo, không khẳng định không đáp ứng deadline hoặc giới hạn hai dự án.

#### FR-044 — Chạy gợi ý nhân sự giải thích được

**Actor/điều kiện:** PTCM; chọn tập task chờ phân công hợp lệ cùng project mở. **Nguồn:** S05 mục 1, 3–4; S03 mục 4.

- **AC1:** chạy cho tập đã chọn, trả theo task Top-N mặc định 3, có thể ít hơn/rỗng; lọc mandatory từ official, loại PTCM cùng dự án trước tính điểm; có điểm thành phần/giải thích và thời điểm.
- **AC2:** không tự hạ điều kiện/chọn Top-1/giao việc, không gọi heuristic là ML hoặc cam kết tối ưu/lịch rảnh. Lưu snapshot bất biến để truy vết, không trả raw snapshot trái quyền; công thức/tham số còn O02.

#### FR-045 — So sánh ứng viên và chuẩn bị bản nháp phân công

**Actor/điều kiện:** PTCM có task/ứng viên được đọc. **Nguồn:** S04 UC003; S05 mục 1, 3; S13 UI-12.

- **AC1:** mỗi task chọn một người, thấy rõ nhánh thủ công/gợi ý, yêu cầu so năng lực, tải hiện tại và tải sau toàn bộ nháp; hủy/bỏ nháp không tạo assignment hoặc tăng tải thật.
- **AC2:** chọn khác Top-1 hoặc chuyển manual sau run phải có lý do; thủ công sau khi chạy gợi ý vẫn truy được lần gợi ý đã tham khảo, nhưng không được trình bày như lựa chọn có hạng/điểm RECOMMENDED. Run một task rỗng không làm mất lựa chọn task khác.

#### FR-046 — Xác nhận phân công theo batch

**Actor/điều kiện:** PTCM hiện hành; batch không rỗng/không trùng task, cùng project. **Nguồn:** S04 UC003; S05 mục 3; S03 mục 4.

- **AC1:** kiểm tra lại quyền, project/task, candidate, đặc tả/requirements và kỹ năng liên quan; một dòng lỗi thì không lưu dòng nào. Hợp lệ mới tạo assignments, trạng thái/người hiện tại, lịch sử và thông báo cùng giao dịch.
- **AC2:** gửi lại cùng yêu cầu không giao trùng theo CR-002; không chấp nhận kết quả gợi ý bị sửa/ghép sai công việc, người hoặc lần chạy. Skill/spec đổi yêu cầu chạy lại hoặc manual theo nguồn; tải mới không sửa điểm cũ, không thêm bước xác nhận riêng khi lịch sử tính điểm thay đổi (R05).

#### FR-047 — Thay trực tiếp người thực hiện

**Actor/điều kiện:** PTCM; project PLANNING/IN_PROGRESS, task ASSIGNED/IN_PROGRESS/PAUSED. **Nguồn:** S04 Thay người/thu hồi; S03 mục 3–4.

- **AC1:** có lý do, kiểm tra người mới theo manual; A giữ phân công tới khi B được giao thành công, lỗi không mất A. Thành công kết thúc lượt A, tạo lượt B, trạng thái ASSIGNED; B phải Start.
- **AC2:** giữ đóng góp/log/kết quả đúng người/lượt, không mở frozen spec; có phần việc thực tế thì cần ghi chú bàn giao theo nguồn. PENDING_REVIEW phải trả lại trước. Đây là yêu cầu đích, chưa coi đã triển khai FE đầy đủ (mục 4).

#### FR-048 — Thu hồi phân công và giao lại

**Actor/điều kiện:** PTCM; cùng giới hạn trạng thái FR-047. **Nguồn:** S04 Thay người/thu hồi; S05 mục 1.

- **AC1:** thu hồi có lý do, kết thúc lượt/xóa người hiện tại, task về PENDING_ASSIGNMENT; giữ frozen spec và đóng góp. Người cũ chỉ còn góc nhìn lịch sử hợp lệ, không quyền task hiện hành từ lượt đã kết thúc.
- **AC2:** muốn giao lại bằng gợi ý phải thu hồi trước và chạy run mới; chưa tìm/chốt được người mới thì task tiếp tục chờ, không tự khôi phục A hoặc tái sử dụng run đã commit.

### J. Thực hiện, kết quả và trao đổi

#### FR-049 — Xem công việc được giao và đóng góp của mình

**Actor/điều kiện:** người có phân công/đóng góp hợp lệ. **Nguồn:** S02 R06; S04 UC004; S06 mục 3; S13 UI-14–UI-15.

- **AC1:** danh sách/bảng hoặc Kanban dùng cùng công việc/trạng thái theo quyền, lọc không đổi dữ liệu. Chỉ được giao task không đồng nghĩa được xem toàn dự án.
- **AC2:** lượt giao kết thúc giữ lịch sử đóng góp riêng, không mở nội dung hiện hành của người sau. Task đóng/hủy vẫn tra được theo quyền, không còn thao tác thực hiện.

#### FR-050 — Bắt đầu, tạm dừng và tiếp tục task

**Actor/điều kiện:** người hiện được giao; project IN_PROGRESS. **Nguồn:** S03 mục 3; S04 UC004; S06 R01.

- **AC1:** ASSIGNED/PAUSED → IN_PROGRESS qua Start/Resume; IN_PROGRESS → PAUSED có lý do; có thể nhiều task đang làm, không thêm bước xác nhận nhận việc.
- **AC2:** đổi trạng thái không sinh log/timer/giờ; PTCM cùng dự án không nhận/thực hiện task. cách nhóm cột Kanban theo S13, không đổi các trạng thái nghiệp vụ.

#### FR-051 — Nộp kết quả và xem các lần nộp

**Actor/điều kiện:** người hiện được giao; task/project IN_PROGRESS. **Nguồn:** S04 UC004; S06 mục 4; S11 dictionary lõi submission.

- **AC1:** nộp tóm tắt và minh chứng tùy chọn, chuyển PENDING_REVIEW, lưu lần nộp/bộ log hiệu lực thuộc đúng lượt giao; không bắt đủ nhật ký, không nộp thay hoặc tạo giờ giả.
- **AC2:** xem các lần nộp trong quyền; nội dung/bộ log đã nộp bất biến dù log được hiệu chỉnh sau này. Không nộp lặp khi đang chờ review.

#### FR-052 — PTCM nghiệm thu hoặc yêu cầu chỉnh sửa

**Actor/điều kiện:** PTCM hiện hành; task PENDING_REVIEW/project IN_PROGRESS, không tự nghiệm thu đầu ra từng thực hiện. **Nguồn:** S04 UC004; S06 mục 4.

- **AC1:** đối chiếu output/criteria đã đóng băng; đạt chuyển DONE/lưu người, kết luận, thời điểm; chưa đạt có lý do, về IN_PROGRESS để sửa rồi nộp lại.
- **AC2:** thiếu log không tự chặn nghiệm thu hoặc biến thành đầu ra không đạt; không thêm scope trong yêu cầu sửa, không mở lại DONE, không reviewer riêng/PM duyệt thay hoặc tự coi merge PR là nghiệm thu.

#### FR-053 — Trao đổi trong công việc

**Actor/điều kiện:** người hiện được giao hoặc PTCM hiện hành; project PLANNING/IN_PROGRESS, task chưa terminal. **Nguồn:** S10 tiếp nối 28/10; S02 tiếp nối; S13 UI-14.

- **AC1:** gửi/xem trao đổi đúng quyền, lưu tác giả/thời điểm/nội dung; không sửa/xóa, thread/tệp/mention hoặc thêm quyền gửi từ PM-only/Admin/Giám đốc.
- **AC2:** task/project kết thúc không gửi mới; người mất phân công không đọc trao đổi hiện hành bằng lịch sử cũ. Bình luận không sửa scope hoặc thay quyết định nghiệm thu.

### K. Nhật ký và hiệu chỉnh

#### FR-054 — Ghi nhật ký thực hiện công việc

**Actor/điều kiện:** người hiện được giao, task/project IN_PROGRESS. **Nguồn:** S06 mục 1–3/R01.

- **AC1:** ghi ngày, phút nguyên dương và nội dung cho effort thực tế của đúng lượt giao; có thể nhiều khoản/ngày, gồm làm tại nhà. Không ghi/gửi thay, tự sinh từ trạng thái hoặc ép đủ 8 giờ.
- **AC2:** không ngày tương lai, trước phân công/triển khai hoặc sau kết thúc; quá deadline kế hoạch không tự làm ngày thực tế mất hợp lệ. Tổng ngày xuyên task/project >480 phút cần lý do, >1440 phút bị chặn; không làm tròn tăng giờ.

#### FR-055 — Sửa và xóa mềm nhật ký trực tiếp

**Actor/điều kiện:** chính chủ log, còn lượt giao hợp lệ và task/project IN_PROGRESS. **Nguồn:** S06 mục 2–3.

- **AC1:** sửa/xóa có lý do và history trước–sau/actor; sửa ngày/phút tính lại ngày cũ/mới xuyên dự án, không ghi đè dữ liệu nền đã đổi.
- **AC2:** PAUSED/ASSIGNED cần Resume/Start; nhật ký khóa/lượt cũ đi qua hiệu chỉnh. Log xóa mềm không tính tổng nhưng giữ lịch sử, không xóa lần nữa/khôi phục hoặc dùng phút bằng 0 thay xóa.

#### FR-056 — Gửi và rút yêu cầu hiệu chỉnh nhật ký đã khóa

**Actor/điều kiện:** người có đóng góp tự gửi; project IN_PROGRESS/COMPLETED và đúng điều kiện khóa/lượt cũ. **Nguồn:** S06 mục 3, 5.

- **AC1:** ADD/UPDATE/DELETE có lý do, đúng assignment; pending không sửa log/tổng. UPDATE/DELETE tối đa một pending/log; ADD chống retry, không cấm các khoản thật chỉ vì giống nhau.
- **AC2:** chính chủ rút pending được; không chỉnh sửa yêu cầu đã gửi hoặc xử lý lại terminal. Không gửi thay người mất quyền, không tạo ngoại lệ chỉ để bỏ qua Start/Resume.

#### FR-057 — PTCM xét duyệt hiệu chỉnh nhật ký

**Actor/điều kiện:** PTCM hiện hành; request pending, project IN_PROGRESS/COMPLETED. **Nguồn:** S06 mục 5–6; S11 chốt result_log_id/Bàn giao.

- **AC1:** so nền/ngày/tổng xuyên dự án và quyền tại lúc duyệt; chấp thuận cập nhật log/quyết định/thông báo cùng giao dịch, chỉ áp dụng một lần. yêu cầu bổ sung truy được log kết quả; yêu cầu sửa/xóa giữ nội dung trước–sau. Chi tiết liên kết lưu trữ theo S11.
- **AC2:** từ chối có lý do, không đổi log; nền stale không ghi đè. Đổi PTCM thì người mới xử lý theo quyền; không PM/Admin duyệt chỉ từ vai trò đó. Không sửa snapshot nộp/accepted_at hoặc mở lại task.

#### FR-058 — Xem tổng thời gian thực hiện đã ghi nhận

**Actor/điều kiện:** người đọc đúng phạm vi cá nhân hoặc trách nhiệm dự án. **Nguồn:** S06 mục 6–7; S03 mục 5.

- **AC1:** tổng effort bằng tổng phút log hiệu lực chưa xóa chia 60, chỉ làm tròn khi hiển thị; hiệu chỉnh được chấp thuận cập nhật tổng hiện hành, không thay tổng đã chụp trong lần nộp cũ.
- **AC2:** tổng project gồm effort task hủy và phân biệt phần hủy; không có log không đồng nghĩa thực tế không tốn thời gian. Không gọi effort tự khai là chấm công, lương/OT hoặc KPI khách quan.

#### FR-065 — Xem tải công việc hiện tại và dự kiến sau phân công

**Actor/điều kiện:** PM/PTCM hoặc người đọc tổng hợp theo đúng mục đích; không mở nội dung dự án khác. **Nguồn:** S05 mục 1–3; S03 mục 4; S04 UC003.

- **AC1:** tải xuyên dự án tính từ estimate của task hiện giao ở ASSIGNED/IN_PROGRESS/PAUSED/PENDING_REVIEW, không lấy estimate trừ effort; thay đổi trạng thái/phân công cập nhật tải hiện hành. Số project đang có việc hiển thị theo định nghĩa nguồn, chỉ tham khảo, không giới hạn hai dự án.
- **AC2:** nháp phân công cho thấy phần tải tăng và tổng dự kiến mà chưa thay tải thật; tải không là lịch rảnh/bảo đảm deadline hoặc tự cộng trách nhiệm PM/PTCM thành task. Không viết lại điểm/run cũ khi tải đổi; chuẩn hóa điểm và dữ liệu lịch sử còn O02.

### L. Thông báo

#### FR-059 — Nhận và xem thông báo nội bộ

**Actor/điều kiện:** người nhận theo sự kiện và quyền. **Nguồn:** S03 mục 5; S11 notifications; S13 Thông báo dùng chung 29/10.

- **AC1:** giao/thay/thu hồi gửi người bị tác động, nộp gửi PTCM, trả sửa/nghiệm thu gửi người nộp; đề nghị hồ sơ/kỹ năng gửi HR/Giám đốc có quyền xử lý, kết quả gửi chính chủ; hiệu chỉnh dùng cùng nguồn thông báo. Không phát thông báo từ thao tác lỗi hoặc gửi đề nghị kỹ năng cho chính người đề nghị như người duyệt.
- **AC2:** chỉ thấy thông báo của mình, lọc tất cả/chưa đọc và số lượng nhất quán; mở khung không tự đọc. Không email/push/realtime/mention hoặc tự thêm sự kiện CV/AI/auth chưa chốt.

#### FR-060 — Đánh dấu đã đọc và mở tài nguyên từ thông báo

**Actor/điều kiện:** chủ thông báo. **Nguồn:** S02 R06; S13 Thông báo dùng chung 29/10; S12 mục 3.

- **AC1:** đánh dấu đọc không duyệt yêu cầu/chuyển trạng thái task; mở thành công tới đúng tài nguyên/ngữ cảnh rồi đánh dấu đọc. Cách mở/đóng và bố cục khung thông báo theo S13, không quy định lại trong FR.
- **AC2:** mở lại kiểm tra quyền hiện hành; mất quyền/tài nguyên thì từ chối không lộ dữ liệu, vẫn cho đánh dấu thông báo của mình đã đọc. Không dùng notification/link cũ để cấp lại quyền.

## 3A. Quy tắc nghiệp vụ và yêu cầu dùng chung

Các mục này gom ràng buộc đã có ở nguồn; không thêm phân hệ, vai trò, API hay bảng. Một quy tắc nằm tại workflow chuẩn vẫn có hiệu lực dù chưa được cấp mã BR/CR ở đây; không sao chép toàn bộ workflow thành bộ quy tắc song song.

#### BR-001 — Địa chỉ khi hoàn thiện hồ sơ

**Áp dụng:** FR-005, FR-009–FR-012, FR-026, FR-042–FR-046. **Nguồn:** S07 Địa chỉ khi hoàn thiện hồ sơ; S10 quyết định địa chỉ; S12 mục 1.

- **AC1:** chưa có địa chỉ thì hiện Chưa cập nhật/nhắc bổ sung ở hồ sơ cá nhân/HR, không bịa dữ liệu; không bắt Admin nhập lúc tạo hoặc chặn đăng nhập/nhận việc/thực hiện vì thiếu địa chỉ.
- **AC2:** cung cấp/sửa phải có nội dung sau trim, không xóa trắng địa chỉ đã có; bỏ trường nghĩa là giữ nguyên, không ép đề nghị sửa fullname/phone phải bổ sung địa chỉ. Pending/từ chối không thay địa chỉ chính thức; cùng quy tắc cho HR sửa trực tiếp.
- **AC3:** không trả địa chỉ qua tra cứu nghề nghiệp, ứng viên, snapshot gợi ý hoặc quyền Giám đốc xét kỹ năng.

#### CR-001 — Lịch sử nghiệp vụ và giới hạn dữ liệu

**Áp dụng:** các thao tác tạo/thay đổi dữ liệu có lịch sử theo S03/S11; đặc biệt FR-005–FR-008, FR-010–FR-013, FR-015–FR-016, FR-018–FR-028, FR-029–FR-041, FR-046–FR-057, FR-064. **Nguồn:** S03 R5/mục 4; S06 mục 5–6; S11 Bàn giao; S12 mục 1, 5.

- **AC1:** truy được actor/thời điểm/lý do và thay đổi tương ứng cho trạng thái, trách nhiệm, phân công, năng lực, nhật ký/yêu cầu theo nguồn; không sửa actor cũ khi người phụ trách thay đổi. Lý do bắt buộc ở thao tác nào vẫn theo FR/workflow đó, không tự bắt mọi thao tác đều nhập lý do.
- **AC2:** không xóa vật lý lịch sử, không lộ secret/CV thô/dữ liệu hành chính qua log kỹ thuật hoặc trả toàn bộ audit cho tra cứu nghề nghiệp. Không tự tạo một màn quản trị audit mới.

#### CR-002 — Tính nhất quán khi lưu, gửi lại và dữ liệu đã đổi

**Áp dụng:** các FR ghi dữ liệu, theo quy tắc từng luồng; đặc biệt FR-005, FR-011, FR-022–FR-027, FR-029–FR-041, FR-046–FR-048, FR-050–FR-057, FR-064. **Nguồn:** S03 R4/R7; S04 UC003; S06 mục 2, 5–6; S08 mục 6; S12 mục 3.

- **AC1:** kiểm tra quyền/trạng thái/giá trị nền ở thời điểm lưu; xung đột không ghi đè, lỗi không để dữ liệu nghiệp vụ liên quan lưu một phần hoặc mâu thuẫn. Chỉ sinh thông báo khi sự kiện nghiệp vụ tương ứng thành công; không ghi thông báo cho thao tác thất bại.
- **AC2:** thao tác cần chống gửi lại theo nguồn không tạo tác động trùng khi gửi lại cùng yêu cầu; dùng lại định danh chống lặp cho nội dung khác phải báo xung đột. Phạm vi API, cách khóa, hợp đồng key/payload và retention thiết kế tại proposal theo O07; không ép mọi API dùng cùng cơ chế hoặc bổ sung cột version cho mọi bảng.

## 4. Rà bao phủ UC và phân biệt hiện trạng

| Nhóm UC/chức năng hiện hành | FR truy vết | Contract/UI tham chiếu |
|---|---|---|
| Đăng nhập/phiên; quản lý tài khoản | FR-001–FR-008, FR-063–FR-064 | UI-01 và tiếp nối phiên 29/10 |
| Quản lý hồ sơ cá nhân/HR và duyệt thông tin | FR-009–FR-013; BR-001 | UI-02–UI-04 |
| Quản lý phòng ban, vị trí, cấp bậc | FR-015 | UI-08 |
| Quản lý danh mục kỹ năng; xem năng lực | FR-016–FR-017 | UI-09, góc nhìn hồ sơ theo quyền |
| Quản lý CV cá nhân/nhân sự; xét kỹ năng từ CV | FR-018–FR-022 | UI-05, UI-07 |
| Yêu cầu và xét duyệt kỹ năng chủ động | FR-023–FR-028 | UI-06–UI-07 |
| UC001 — Khởi tạo/quản lý dự án | FR-029–FR-036 | UI-10 |
| UC002 — Nhóm/task | FR-037–FR-041 | UI-11, UI-10/14 cho thao tác vòng đời |
| Tra cứu hồ sơ nhân sự độc lập | FR-042 | UI-13 |
| UC003/UC003.1/UC003.2 — Gợi ý/thủ công/xác nhận; thay/thu hồi | FR-043–FR-048, FR-065 | UI-12 |
| UC004 — Thực hiện/nhật ký/nộp/nghiệm thu và trao đổi | FR-049–FR-058 | UI-14–UI-15 |
| Thông báo | FR-059–FR-060 | Tiếp nối thông báo 29/10 |
| Lịch sử và bảo vệ thao tác xuyên luồng | CR-001–CR-002; áp dụng cùng các FR liên quan | Workflow/quyền/bảo mật |

Các UC hỗ trợ danh mục được bao phủ dù không cần sơ đồ phân rã riêng. BR-001 và CR-001–CR-002 được nghiệm thu cùng chức năng sử dụng, không bắt buộc vẽ thành UC hoặc tạo ticket riêng. Bảng này đối chiếu **UC dạng văn bản hiện hành**, chưa kiểm kê hay sửa toàn bộ ảnh sơ đồ/báo cáo gốc.

**Hiện trạng để tránh hiểu nhầm tiến độ:**

- Backend sau rà soát chỉ có skeleton/health, công cụ migration/seed và kiểm thử nền; **chưa xác nhận bất kỳ FR nào ở trên đã được backend đáp ứng đầy đủ**. Migration 000001/seed legacy không thay thiết kế đích.
- FE đã có nhiều hành trình mô phỏng tài khoản, hồ sơ/danh mục, CV metadata, kỹ năng thủ công, dự án/task/phân công/thực hiện/effort và thông báo/phiên. Bằng chứng cụ thể và giới hạn nằm tại S14; lần lập FR này không chạy lại UI hoặc chứng nhận toàn bộ test FE.
- Chưa dùng fixture điểm, danh tính mô phỏng, metadata upload hoặc thao tác reducer làm bằng chứng engine, xác thực, lưu tệp, transaction hay cạnh tranh thật.
- Xét kỹ năng từ CV/engine thật, CV làm căn cứ cùng lưu/đọc file, làm rõ sau phân công, thay/thu hồi giữa chừng đầy đủ còn phần hoãn/chưa hoàn thiện. Yêu cầu đích vẫn được kiểm kê, **không tự đưa tất cả vào đợt code đầu**.

## 5. Điểm mở và điều kiện trước khi triển khai

| ID | Cần làm rõ/chốt ở bước thiết kế | FR ảnh hưởng | Cách xử lý hiện tại |
|---|---|---|---|
| O01 | A01–A09 đã chốt email/chuẩn hóa, mật khẩu mới, phiên 8 giờ/CSRF/đổi lần đầu, hướng test/migration; còn cấp username, sinh/gửi lỗi/gửi lại, reset, đổi tự nguyện và retention phiên | FR-001, FR-003–FR-008, FR-063–FR-064 | Proposal auth phân biệt quyết định đã duyệt với OpenAPI/ports/SQL/mapping legacy/fixture chưa hoàn thiện. Không áp username demo/mật khẩu seed hoặc tự thêm invite/JWT; mật khẩu tạm chưa có hạn. Estimate/dự phòng còn đề xuất, không phải giờ đã làm. |
| O02 | Công thức gợi ý cuối, miền/chuẩn hóa weight, tham số/phiên bản và tiêu chí dữ liệu lịch sử tin cậy | FR-038, FR-044–FR-046, FR-065 | Giữ định hướng và quy tắc eligibility; bộ 0,7/0,2/0,1, 80h, 1–5… chỉ thử nghiệm. Chưa chốt điều kiện nghiệm thu điểm số cụ thể. |
| O03 | Cloud/provider, vòng đời tệp/cleanup, xác thực download và cơ chế gắn tệp thật vào CV/căn cứ | FR-018–FR-021, FR-024 | Không chọn dependency/provider hoặc gọi metadata demo là đã lưu tệp. HR xác nhận CV hiện tại thay nhân viên chưa được mở quyền. |
| O04 | API/model/parser, mapping/adapter, worker/lease/retry và bộ CV đánh giá được phép dùng | FR-021–FR-022 | Hoãn tích hợp theo chốt người dùng; không biến đề xuất model thành lựa chọn đã xác nhận. |
| O05 | Loại/số/dung lượng tệp minh chứng ngoài CV, giữ/xóa tệp chưa gắn | FR-024–FR-025 | Không áp giới hạn 10 MiB của CV sang mọi minh chứng; giữ giải trình/căn cứ không cần tệp. |
| O06 | Cách ghi/hiển thị làm rõ sau phân công mà không đổi scope; khác gia hạn đã chốt | FR-039 | Không mở form sửa toàn bộ frozen spec. Gia hạn FR-040 có thể thiết kế riêng. |
| O07 | Endpoint nào cần idempotency, retention/retry và contract cạnh tranh/khóa cụ thể | FR-046, FR-054–FR-057; CR-002 | Chốt trong từng proposal/API; không tạo schema hoặc áp key cho mọi thao tác chỉ từ FR tổng quát. |
| O08 | DTO/lọc/phân trang và dữ liệu tổng hợp theo mục đích; ranh giới chọn tất cả tập lọc so với tập dòng đang trình bày | FR-026, FR-032, FR-042–FR-046, FR-049, FR-059–FR-060, FR-065 | Workflow gợi ý nói tập lọc gồm nhiều trang; FE contract 29/10 ghi chọn tập dòng đang trình bày. Không coi FE đã đáp ứng chọn mọi trang; chốt ngữ cảnh trong proposal trước bổ sung phân trang/chọn batch. Không mở quyền dữ liệu mới. |
| O09 | Phạm vi kỹ năng cá nhân của Giám đốc khi tổ chức chỉ có một Giám đốc | FR-023–FR-028 | Đã hoãn, không tự duyệt/người duyệt thay, không lấy làm demo chính; không cản các yêu cầu của nhân viên khác. |

Danh sách trên không phải yêu cầu người dùng trả lời hết ngay. Mỗi proposal chỉ đóng các điểm ảnh hưởng trực tiếp feature sẽ làm; không trì hoãn mọi luồng vì một tích hợp CV còn mở.

## 6. Ngoài phạm vi và bước kế tiếp

Giữ ngoài phạm vi của S01: không roster/duyệt điều động, sprint/duyệt kế hoạch/Slack integration, reviewer riêng, task dependency, chấm công/lương/OT/timer, tự giao việc, email/push, OCR, đa tổ chức, tự hết hạn kỹ năng hoặc mở lại task/project terminal. Không mở luồng nhân viên xin gỡ kỹ năng. Không tạo thêm bảng chỉ vì một FR hoặc một khối giao diện.

Đã lập [bản đồ miền](business-domains.md): gom FR và xác định trách nhiệm/dữ liệu/phụ thuộc. Sau khi người dùng duyệt đã soạn [proposal feature đầu về auth/session](../proposals/backend-auth-session.md); chốt email login 05/11 được ghi tại nguồn và FR-001/O01. A02–A09 đã duyệt ngày 06/11; còn hoàn thiện OpenAPI/ports và kế hoạch SQL/test cụ thể. WBS/estimate và phần người dùng tự viết Go/phần agent hỗ trợ cần thống nhất trước khi đưa task lên Jira và code. Chưa tạo ticket, chưa bắt đầu bài config hoặc API nghiệp vụ.

## 7. Truy vết thay đổi sau review

Đây là điều chỉnh cách đặc tả, không thêm hoặc bỏ nghiệp vụ đã chốt. Các FR không liệt kê dưới đây giữ ID/mục đích; có rút chi tiết triển khai và dẫn lại contract nguồn. Không dùng ID đã chuyển loại cho chức năng mới.

| ID bản đầu | Vị trí hiện tại | Lý do và cách đọc lại tiêu chí cũ |
|---|---|---|
| FR-004 | FR-004 và FR-063 | Đổi mật khẩu chính chủ khác reset bởi Admin. AC1 cũ theo FR-004; reset và chính sách ở AC2 cũ theo FR-063/O01. |
| FR-007 | FR-007 và FR-064 | Tách xem/tìm tài khoản khỏi thay quyền. Phần danh sách của AC1 cũ theo FR-007; cấp/thu hồi ở AC1 và ràng buộc AC2 cũ theo FR-064. |
| FR-014 (ngừng dùng) | BR-001, áp dụng cùng FR-005, FR-009–FR-012 và các góc nhìn dữ liệu liên quan | Ba AC địa chỉ được giữ tại BR-001; không còn một chức năng độc lập mang tên địa chỉ. |
| FR-058 | FR-058 và FR-065 | AC1 cũ là effort đã ghi nhận; AC2 cũ là tải tồn đọng. Hai đại lượng khác mục đích/tập dữ liệu, không suy cái này từ cái kia. |
| FR-061 (ngừng dùng) | CR-001 | Hai AC lịch sử/giới hạn dữ liệu chuyển sang yêu cầu dùng chung; không coi là một màn hình mới. |
| FR-062 (ngừng dùng) | CR-002 | Hai AC nhất quán/chống lặp chuyển sang yêu cầu dùng chung; thiết kế API/khóa vẫn ở proposal. |

Các ID mới FR-063–FR-065 chỉ tách nội dung từ bản đầu. Ba ID FR ngừng dùng vẫn có trong bảng này để tra cứu, không tính vào số FR đang dùng. Các AC được chuẩn hóa lần này là dự thảo hiện hành; không tuyên bố đã có test/API/Jira bám theo số AC của bản đầu.
