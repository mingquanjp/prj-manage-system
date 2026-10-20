# Rà soát trước triển khai frontend — 20/10/2026

> **Báo cáo tại thời điểm trước pilot**, giữ để truy vết kết quả đọc Make. Sau đó web local đã khởi tạo và visual được duyệt; hiện trạng/lộ trình hiện hành ở [web/README](../../web/README.md). Các câu “web rỗng/chưa khởi tạo” bên dưới là lịch sử, không hướng dẫn triển khai tiếp.

## 1. Kết luận và phạm vi

**Đủ cơ sở để bắt đầu frontend theo từng lát cắt nghiệp vụ; chưa phải bản bàn giao đã đồng bộ hoàn toàn hoặc có thể sao chép nguyên mã Make.** Không cần tiếp tục dùng AI Make chỉ để bắt đầu code. Giữ prototype làm tham chiếu visual/luồng; quyết định hiện hành và workflow là nguồn nghiệp vụ.

Đã đọc nguồn Make của file `fOXzi51sKaRQbuSKmHZQmD`, đối chiếu phiên bản 16 “Polish DATN prototype features” trong ứng dụng Figma, tài liệu repo, thiết kế DBML, API và skeleton hiện có. Nguồn đọc gồm App, types, AppShell, ProjectDetail, RecommendationResults, DraftReview, ManualDrawer, AccountManagement, PersonalProfile, EmployeeTaskDetail, mockData, organizationMockData, CSS và hai tài liệu import hồ sơ/screen-contracts.

**Giới hạn xác minh:** preview riêng trả `net::ERR_EMPTY_RESPONSE`; ứng dụng Figma cũng hiện mất kết nối/live preview loading. Vì vậy lượt này là review nguồn và tài liệu, không xác nhận click-through bản v16, không xác nhận footer/modal bằng quan sát viewport, không tự chạy build/typecheck của Make. Kết quả build do Make báo không được coi là kết quả do lượt review này thực hiện. Không gửi prompt, chạy AI Make, publish, sửa mã ứng dụng hay chạy migration.

**Cập nhật tiếp nối cùng ngày:** sau review, người dùng đã chốt địa chỉ phải bổ sung khi hoàn thiện hồ sơ nhưng không bắt nhập lúc tạo. Tài liệu repo/dictionary/DBML đã được đồng bộ theo [baseline 20/10](../business-baseline-2026-10-20.md); F01 bên dưới ghi ranh giới còn lại với Make. Không sửa code hoặc migration.

## 2. Kết quả hoàn thiện prototype nhìn từ nguồn

| Hạng mục | Kết quả đọc nguồn mới nhất |
|---|---|
| Hồ sơ cá nhân | Mặc định chỉ đọc; nút “Đề nghị cập nhật” mở modal fullname/phone/address; gửi pending, không cập nhật chính thức ngay |
| Địa chỉ | v16 có trường nullable trong Employee/request/HR; quyết định tiếp nối yêu cầu nhắc bổ sung khi hoàn thiện, không cho xóa trắng; repo đã đồng bộ, code Make chưa cập nhật quyết định tiếp nối |
| Bản nháp phân công | Mỗi task hiển thị đúng estimate riêng; tổng thêm theo nhân viên ở khối tổng; hiện lý do lựa chọn và lý do chấp nhận thiếu kỹ năng |
| Thẻ gợi ý | Đưa giải thích/đối chiếu vào vùng mở rộng; nhãn năng lực và yêu cầu tách rõ; đã thêm khoảng đệm và sticky footer trong nguồn, chưa xác minh bằng preview |
| Tra cứu thủ công | Có tính khoảng thiếu kỹ năng cho từng ứng viên trước chọn; giữ kiểm tra người bị khóa/inactive/PTCM cùng dự án |
| Điều hướng | “Công việc của tôi” đã lên menu chính; vẫn dùng hai shell/mode và route state của prototype |
| Tài khoản | Đã tách trạng thái/thao tác, nhưng nút sửa và khóa vẫn là chữ, chưa chuyển sang icon như prompt yêu cầu |

Không dùng bảng này để tuyên bố mọi tính năng hoặc mọi trạng thái đã kiểm thử đạt.

## 3. Các điểm cần xử lý khi chuyển sang frontend thật

### F01 — Địa chỉ đã đồng bộ thiết kế; Make còn theo quy tắc trước chốt cuối

Quyết định cuối: địa chỉ phải bổ sung khi hoàn thiện hồ sơ, không bắt Admin nhập lúc tạo. `employees.address TEXT NULL` hỗ trợ chưa bổ sung, không mang nghĩa thông tin tùy ý bỏ qua. Nhân viên đề nghị bổ sung/sửa, HR duyệt hoặc cập nhật trực tiếp; giá trị mới có nội dung, không xóa trắng địa chỉ đã có. Thiếu chỉ nhắc bổ sung, không chặn đăng nhập/nhận việc/thực hiện task hoặc thêm trạng thái.

Tài liệu nghiệp vụ/API/UI contract, dictionary 5B, DBML và mapping chuyển đổi trong repo đã cập nhật theo quyết định cuối. Không thêm bảng, không sửa/chạy migration. Make v16 vẫn là nguồn tham chiếu cũ: **cả hai import được đọc trên Make còn nội dung cũ**:

- `src/imports/employee_profile_workflow.md:31` vẫn ghi allowlist fullname/phone.
- `src/imports/screen-contracts.md:11` vẫn ghi yêu cầu sửa fullname/phone.

Mã Make còn cho chuỗi rỗng biểu thị xóa địa chỉ; đây là quy tắc trước khi người dùng chốt cuối, không port hành vi đó vào frontend thật. Khi chuyển frontend, dùng tài liệu repo đã cập nhật làm chuẩn; không cần gửi thêm prompt Make để hoàn tất quyết định.

Form/frontend thật: không gửi address khi không đổi, báo lỗi nếu nhập mới/sửa thành chuỗi trắng hoặc NULL, HR so dữ liệu nền đúng trường. Bảng request dùng JSON patch hiện có, before address có thể NULL khi bổ sung lần đầu. Không ép một patch chỉ sửa fullname/phone phải bổ sung address ngay; không lộ address trong tra cứu chuyên môn/ứng viên/snapshot.

### F02 — Trạng thái bản nháp phân công chưa có một nguồn nhất quán

Các thay đổi v16 cải thiện luồng nhưng không nên port nguyên cơ chế giữ component ẩn:

- `App.tsx` giữ ProjectDetail và RecommendationResults bằng `display:none` khi xem draft. Cách này giữ một số lựa chọn nhưng không phải cơ chế điều hướng/frontend đích.
- Quay lại nhánh manual chỉ đưa route về ProjectDetail; tại lúc mở draft, `manualTask` đã được đặt null. Chưa có đường khôi phục drawer với ứng viên/lý do hiện tại.
- `ProjectDetail.tsx:111–131` nối entry vào `directDraftEntries`; khi vào lại chọn thủ công không thấy reset/thay thế theo task. Có nguy cơ lặp task nếu sửa lại bản nháp; đây là nhận định từ nguồn, chưa tái hiện tương tác v16.
- `DraftReview.tsx` tự giữ bản sao entries và xóa dòng ở state riêng. Thay đổi không được đẩy về nguồn lựa chọn; quay lại có thể phục hồi dòng vừa loại.
- RecommendationResults khởi tạo row state theo tập task lần mount đầu, effect có dependency rỗng. Cần xác định draft/run theo project và tập task, không dựa vào component còn được ẩn hay đã unmount.

Hướng triển khai: một draft theo project/lượt thao tác, tối đa một entry/task, lưu nguồn manual/recommended, run/item khi có, lý do và lựa chọn. Quay lại sửa đúng draft; hủy rõ ràng mới xóa. Không lưu trạng thái nghiệp vụ đã xác nhận trong draft.

### F03 — Không để mất nguồn/lý do khi xác nhận phân công

`App.tsx` hiện chỉ chuyển taskId, employeeId và mandatorySkillOverrideReason sang `applyConfirmedAssignments`; mock helper không lưu source/run/rank/adjustReason. Vì vậy việc draft hiển thị lý do chưa chứng minh lịch sử phân công đã giữ được chúng.

Frontend thật phải gửi đầy đủ các trường được API chấp nhận: định danh task/người, run/item khi dùng và lý do theo từng dòng. Không gửi rank/score như sự thật để server tin. Kết quả xác nhận do server kiểm tra lại và ghi nguyên tử; frontend chỉ cập nhật cache từ kết quả, không tự tăng tải chính thức. Không tự thêm cột DB chỉ vì mock thiếu dữ liệu.

### F04 — Ngưỡng màu tải chưa phải quyết định nghiệp vụ

`DraftReview.tsx:243` vẫn tô đỏ khi dự kiến từ 80 giờ; nguồn gợi ý cũng có các ngưỡng màu điểm demo. Không đem các số này thành trần tải hoặc kết luận quá tải chính thức. Trước mắt dùng trình bày trung tính cho tải chưa có ngưỡng chốt; điểm minh họa phải được nhận diện là minh họa. Công thức cuối, miền trọng số và DTO chỉ số cần chốt trước khi nối engine thật, không chặn dựng layout/phân công thủ công.

### F05 — Phân biệt phần prototype chưa dựng với phần nghiệp vụ bị bỏ

Ở nguồn thực hiện công việc được đọc có Start/Resume, thêm/sửa log, nộp và hiển thị kết quả. Chưa có thao tác Pause, xóa log, yêu cầu hiệu chỉnh nhật ký đã khóa hoặc trao đổi công việc trong màn hình này. Luồng thay/thu hồi phân công, đóng góp lịch sử và thông báo cũng không được coi là đã hoàn thiện chỉ vì đủ năm đợt prompt.

Các chức năng đã có trong workflow/ma trận vẫn thuộc phạm vi đích. Ghi rõ vào backlog frontend, triển khai theo từng đợt; không cần dựng tất cả lên Make trước. Mỗi hành động cần có quyền, trạng thái cho phép, dữ liệu gửi, kết quả và lỗi chính; không tự mở rộng ngoại lệ hiếm.

### F06 — Backend hiện tại chưa phải API tích hợp

- `web/` vẫn rỗng: chưa có ứng dụng frontend hay cấu hình build frontend trong repo.
- `internal/transport/http/router.go` mới đăng ký `GET /api/healthz`.
- `.claude/docs/api_design.md` là thiết kế, chưa phải OpenAPI/DTO chạy được cho các màn hình.
- Migration cũ còn role ADMIN/PM/EMPLOYEE, project_members, trạng thái danh mục kỹ năng và các khác biệt đã được quyết định thay. Không dùng nó làm nguồn để ép frontend quay lại nghiệp vụ cũ; không chạy migration này để “chuẩn bị frontend”.

Có thể làm frontend trước qua mock adapter có cùng giao diện với API client tương lai. Phải hiển thị rõ phần mô phỏng; không báo đã tích hợp backend, auth, upload hoặc extraction thật.

### F07 — Tài liệu quan trọng đang bị Git bỏ qua

`git check-ignore` xác nhận baseline 20/10, DBML và backlog polish vẫn bị ignore theo `.gitignore` hiện tại. Các báo cáo review mới trong docs cũng chịu quy tắc ignore. Tệp có trên máy không đồng nghĩa clone repo sẽ có đủ đầu vào.

Trước bàn giao/commit nền frontend: rà và chọn chính xác tài liệu nguồn cần quản lý phiên bản, sửa allowlist Git có chủ đích; không đưa toàn bộ tài liệu cá nhân/thesis/context vào Git. Lượt này chỉ ghi nhận, chưa sửa `.gitignore` hoặc stage/commit.

## 4. Nền frontend được đề xuất theo công nghệ đã chốt

Giữ Vite + React + TypeScript + Tailwind + shadcn/ui + React Router + TanStack Query theo CLAUDE/architecture. Chưa cài package hoặc chọn lại version trong lượt review.

- **Visual:** tái dùng màu canvas/surface/shell, typography, khoảng cách và pattern bảng/form/drawer của Make; icon chuẩn thay ký tự minh họa. Tách token khỏi style rải rác.
- **Điều hướng:** route thực có URL và Back/Forward; một shell/menu hợp nhất theo quyền, không lấy `employeeMode` làm kiến trúc đích.
- **Quyền:** system roles ADMIN/HR/DIRECTOR; PM/PTCM theo dự án; phía client phục vụ trải nghiệm, server vẫn kiểm tra. Demo identity selector chỉ thuộc chế độ mô phỏng, không sản phẩm thật.
- **Dữ liệu:** Query/cache cho server state; state form/draft cho thao tác đang làm. Không sao chép mảng mutable toàn cục, bộ đếm tải hay profileVersion mock thành thiết kế DB/API.
- **Hợp đồng:** tách API DTO và view model; map tên trường/ID/mức kỹ năng rõ ràng. Không coi mã nhân viên E004 là employee_id DB, không lấy chuỗi nhãn phòng ban làm khóa ngoại.
- **Session:** hướng đích cookie HttpOnly/server session và CSRF theo security.md; không lưu token đăng nhập trong localStorage, không đổi demo identity thành cấp quyền thật. Login/reset/timeout còn cần đặc tả trước tích hợp auth.
- **Trạng thái màn hình:** loading, empty, lỗi từng trường/lỗi quyền, lỗi nền dữ liệu thay đổi và thông báo thành công. Chỉ kiểm tra chính ở từng lát cắt, chưa cần bộ hồi quy production rộng ngay khi dựng nền.

## 5. Thứ tự triển khai và điều kiện kết thúc

| Bước | Đầu ra | Có thể làm khi backend chưa có? |
|---|---|---|
| 0. Chốt bàn giao | Địa chỉ đã đồng bộ thiết kế; còn rà tài liệu cần theo Git và hợp đồng tối thiểu cho lát cắt đầu; không port quy tắc xóa địa chỉ từ Make v16 | Có; chưa chạy migration |
| 1. Nền frontend | Khởi tạo web theo stack đã chốt; token/component dùng chung; Router; Query; API/mock adapter; shell và kiểm tra quyền cho UI | Có |
| 2. Lát cắt chính | Giám đốc tạo/chọn PM → PM chỉ định PTCM → tạo task → phân công thủ công → bắt đầu → thực hiện/nộp → nghiệm thu | Có qua mock adapter; không giả backend thật |
| 3. Gợi ý/tra cứu | Candidate/run DTO, bản nháp có nguồn và lý do, tải tổng hợp; engine thật sau khi chốt công thức/contract | UI mock có; điểm thật chưa |
| 4. Hồ sơ và quản trị | Tài khoản/hồ sơ/địa chỉ/yêu cầu HR, danh mục; các hành động được kiểm tra theo quyền | Có qua mock adapter |
| 5. Phần còn lại | CV/kỹ năng, hiệu chỉnh, đóng góp/thông báo/thay-thu hồi theo workflow; tích hợp API từng nhóm khi sẵn sàng | Tách rõ UI và dịch vụ thật |

Không cần thiết kế lại database toàn bộ hoặc bổ sung một loạt bảng để khởi tạo frontend. Không phải chờ công thức gợi ý cuối mới làm được luồng phân công thủ công. Các quyết định upload/auth/engine chỉ cần đóng trước triển khai phần liên quan, không ép chốt đồng loạt hôm nay.

## 6. Bước tiếp theo nên thực hiện

Hoàn tất bước 0 trước, rồi mới khởi tạo `web/`. Khi bắt đầu feature/fix cần theo quy tắc branch của repo; không cài thêm công nghệ ngoài stack đã chốt hoặc thay schema chưa được trao đổi. Make v16 là tham chiếu hữu ích, không phải mã nguồn sản phẩm hay bằng chứng mọi nghiệp vụ đã chạy.

## 7. Review đồng bộ quyết định địa chỉ — tiếp nối 20/10

- Đã đồng bộ baseline/sổ quyết định, CLAUDE, workflow/phạm vi/quyền/quy tắc/bảo mật, API, dictionary 5B/DBML/mapping chuyển đổi, brief/matrix/checklist/prompt và bản bàn giao/backlog. Nội dung lịch sử được giữ nhưng có quyết định tiếp nối ưu tiên.
- Đã kiểm tra tĩnh 22 tệp liên quan: code fence đóng/mở, newline cuối tệp, không whitespace cuối dòng; 174 liên kết tệp cục bộ tồn tại. Chưa kiểm tra mọi anchor hoặc URL ngoài.
- DBML vẫn 32 bảng; đúng một cột address kiểu text, không not null trong employees; bảng yêu cầu hồ sơ chỉ mở rộng allowlist JSON, không thêm cột địa chỉ riêng. Đây là kiểm tra cấu trúc văn bản, chưa chạy parser DBML/export SQL.
- Rà tay các tình huống theo đặc tả: tạo chưa có địa chỉ; bổ sung lần đầu với before NULL; sửa địa chỉ có nội dung; từ chối rỗng/NULL khi gửi thay đổi; bỏ khóa không đổi; pending không đổi official; HR duyệt có đối chiếu nền; không chặn công việc hoặc lộ tra cứu. Đây là review contract, không phải test ứng dụng đã pass.
- `git diff --check` đạt; hai file migration 000001 có hash SHA-256 không đổi trước/sau. Không sửa code/migration hoặc truy cập DB thật, không cập nhật Make. Vấn đề Git ignore ở F07 vẫn giữ nguyên, không tự mở rộng phạm vi đợt địa chỉ.
