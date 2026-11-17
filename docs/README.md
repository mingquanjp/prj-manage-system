# Bản đồ tài liệu — nguồn nào dùng khi nào

Đây là mục lục, không phải một bản nghiệp vụ mới. Không cần đọc hết các file Markdown để làm một màn hình.

## 1. Thứ tự ưu tiên

1. Quyết định **hiện hành/Bàn giao**, gồm chốt luồng đề nghị kỹ năng nhân viên ngày 27/10/2026 và cấp bậc theo danh mục công ty ngày 26/10/2026, trong [sổ quyết định DB](design/database-review-decisions-2026-10-14.md), cùng chốt tiếp nối [20/10, điều kiện ghi nhật ký 21/10 và chốt cấp mã/quyền tài khoản 28/10](business-baseline-2026-10-20.md), trên nền [12/10](business-baseline-2026-10-12.md) và [13/10](business-baseline-2026-10-13.md). Ngày mới hơn không tự biến mọi đề xuất thành quyết định đã chốt.
2. Workflow/quy tắc dưới đây diễn giải triển khai. Nếu khác quyết định đã xác nhận, báo và đồng bộ trước khi code phần bị ảnh hưởng.
3. UI contract chuyển nghiệp vụ thành màn hình; không cấp thêm quyền. Phong cách local đã duyệt tại [web/README.md](../web/README.md) ưu tiên bố cục visual Make cũ, không thay nghiệp vụ.
4. Code cho biết **đã triển khai gì**, không quyết định nghiệp vụ đích. Migration 000001 còn cũ; DBML là thiết kế, không migration.

## 2. Nguồn chuẩn theo miền — chỉ đọc phần liên quan

| Công việc | Nguồn |
|---|---|
| Phạm vi/UC | [product_scope](../.claude/docs/product_scope.md) |
| Quyền/trạng thái | [roles_permissions](../.claude/docs/roles_permissions.md), [domain_rules](../.claude/docs/domain_rules.md) |
| Dự án/nhóm/task/phân công | [project_task_assignment_workflow](../.claude/docs/project_task_assignment_workflow.md) |
| Gợi ý/tra cứu | [ai_recommendation_workflow](../.claude/docs/ai_recommendation_workflow.md) |
| Thực hiện/nhật ký/hiệu chỉnh | [work_effort_workflow](../.claude/docs/work_effort_workflow.md) |
| Hồ sơ/địa chỉ/HR | [employee_profile_workflow](../.claude/docs/employee_profile_workflow.md) |
| CV/kỹ năng | [cv_extraction_approval_workflow](../.claude/docs/cv_extraction_approval_workflow.md), [skill_approval_workflow](../.claude/docs/skill_approval_workflow.md) |
| Kiến trúc/API/bảo mật | [architecture](../.claude/docs/architecture.md), [api_design](../.claude/docs/api_design.md), [security](../.claude/docs/security.md) |
| Dữ liệu/kiểm thử đích | [database_schema](../.claude/docs/database_schema.md), [testing_verification](../.claude/docs/testing_verification.md) |
| Màn hình/quyền/dữ liệu UI | [screen-contracts](design/figma-make/screen-contracts.md), [ui-brief](design/figma-make/ui-brief.md) |
| Tham chiếu visual/tình huống | [Guidelines](design/figma-make/Guidelines.md), [review-checklist](design/figma-make/review-checklist.md) |

Các checklist đặc tả không phải kết quả test đã chạy. Một số file còn phần lịch sử được đánh dấu bên trong; không áp dụng phần đã bị thay thế.

### Chuẩn bị backend từ yêu cầu

Chuẩn bị backend: [dự thảo FR và tiêu chí nghiệm thu](product/functional-requirements.md) kiểm kê yêu cầu từ UC/workflow để chuẩn bị phân miền và proposal; phân biệt chức năng FR, quy tắc BR và yêu cầu dùng chung CR, có bảng truy vết ID sau review. Đây không phải nguồn thay quyết định nghiệp vụ, bản thiết kế API hoặc xác nhận chức năng đã chạy; điểm chưa chốt được ghi riêng trong tài liệu.

[Bản đồ miền nghiệp vụ](product/business-domains.md) nối FR với trách nhiệm, dữ liệu thiết kế và phụ thuộc; đề xuất 6 miền và phần hỗ trợ thông báo trong cùng backend. Đọc sau FR, trước chọn feature/proposal; không phải sơ đồ package hoặc kế hoạch tách service.

[Proposal auth — thiết kế hiện hành](proposals/backend-auth-session.md) trình bày mục tiêu/phạm vi, kiến trúc, DB, năm sequence API, bảo mật và WBS. Backend đã tích hợp đủ năm API với PostgreSQL test tại DATN-18; DATN-19 đã nối FE bốn API nền với mode server/demo riêng và màn tiếp nối tạm; DATN-20 đã nối form đổi mật khẩu, DATN-21 đang tổng hợp kiểm chứng xuyên luồng và runbook. Cách chạy và bằng chứng browser/proxy xem [web README](../web/README.md#chạy-fe-auth-thật--datn-19--auth-t18). Không đồng nghĩa đã áp migration Neon, có API quản trị hoặc gửi email.

- [OpenAPI auth](api/auth.openapi.json): contract HTTP duy nhất cho năm route; kiểm tĩnh bằng `node --test scripts/check-auth-contract.test.mjs`, không thay kiểm thử runtime.
- [Chi tiết triển khai và bằng chứng theo ticket](proposals/backend-auth-session-implementation.md): ports minh họa ban đầu, SQL/fixture/migration, config và bàn giao DATN-2–DATN-18. Đọc đúng mục cần tra; câu “chưa triển khai” trong ghi nhận cũ chỉ đúng tại thời điểm ghi. Bằng chứng HTTP/PostgreSQL và ma trận AUTH-S01–S10 ở mục 6.2.13.
- [Runbook auth FE–backend](api/auth-runbook.md): lệnh hằng ngày, demo, ma trận bằng chứng/giới hạn và khắc phục; trạng thái ca chưa kiểm được ghi riêng.
- [Runbook PostgreSQL test](../scripts/authdb/README.md): guard, môi trường sở hữu, cách chạy/dọn và integration; không dùng DATABASE_URL/Neon thay DB test.
- [Swagger UI dev và hướng dẫn demo](api/README.md): dùng cùng OpenAPI, tài nguyên local; production không mở docs. Không nhận Swagger render được là bằng chứng browser/cookie/CSRF đã kiểm hết.
- [Bộ ticket auth](proposals/backend-auth-session-tickets.md): 1 Epic và 20 Task, mapping 33 đơn vị WBS/105 giờ cơ sở; review nằm trong Done. Jira DATN hiện ở site riêng theo [workflow](ai/workflow.md); tiến độ thực xem Jira, không lấy estimate làm giờ còn lại.
- [Proposal danh mục tổ chức, cấp tài khoản và hồ sơ](proposals/backend-organization-accounts-profile.md): bản đề xuất cho ba feature tiếp theo D01/D02; đã chốt 15/11 chỉ email đăng nhập, không username sản phẩm, mã mới EMP000001… tại [baseline](business-baseline-2026-10-20.md#quyết-định-tiếp-nối-15112026--email-đăng-nhập-và-mã-nhân-viên). Đã duyệt OAP-A01–A08 và Gmail API/OAuth gmail.send, hoãn cấp lại/gửi lại. OpenAPI/kế hoạch SQL và bộ ticket đã bàn giao. DATN-24 đã có migration M1 trên PostgreSQL test, DATN-25 có service/repository danh mục, DATN-26 nối 12 API danh mục với auth/CSRF và Swagger dev; bằng chứng/giới hạn xem runbook PostgreSQL. Mapping dữ liệu đích và brief OAuth thực tế còn thuộc task tiếp nhận; chưa có API cấp tài khoản. DATN-28 bổ sung API hồ sơ chính chủ, bằng chứng và giới hạn tại [runbook PostgreSQL](../scripts/authdb/README.md#datn-28--api-hồ-sơ-chính-chủ). DATN-27 đã nối UI danh mục/navigation server và kiểm browser với PostgreSQL test; URL, demo F1, bằng chứng và giới hạn ở [web README](../web/README.md#datn-27--oap-t05--ui-danh-mục-server).
- DATN-29 nối phần đọc hồ sơ chính chủ tại `/profile`, giữ HR về danh mục; mapping URL, kiểm browser local và giới hạn tại [web README](../web/README.md#datn-29--oap-t07--hồ-sơ-chính-chủ-và-đích-sau-login). Chưa có cập nhật/lịch sử đề nghị hoặc API cấp tài khoản.
- DATN-31 bổ sung service/transaction tạo user–employee–quyền–audit trên M2, chưa nối endpoint/Gmail/UI. Đầu ra OAP-T11 và bằng chứng/giới hạn tại [runbook create](../scripts/authdb/README.md#datn-31--transaction-tạo-tài-khoản-và-hồ-sơ).
- DATN-32 chuẩn bị adapter Gmail OAuth và CLI cấp quyền/kiểm kết nối. Xem [runbook Gmail](../cmd/gmail-setup/README.md) để cấu hình và ghi bằng chứng gửi thử.
- [OpenAPI tổ chức/tài khoản/hồ sơ](api/organization-accounts-profile.openapi.json): contract canonical cho 16 thao tác; 12 operation danh mục và GET hồ sơ chính chủ đã nối runtime/Swagger dev; ba operation tài khoản còn lại chưa triển khai. Kiểm tĩnh bằng `node --test scripts/check-organization-contract.test.mjs`, không thay kiểm thử HTTP/DB.
- [Kế hoạch SQL và DB test OAP](proposals/backend-organization-accounts-profile-sql-plan.md): preflight/mapping, M1 danh mục/hồ sơ, M2 bỏ username/cấp mã/email, khóa/rollback và ma trận test. Bằng chứng M1 test xem runbook PostgreSQL; DATN-30 đã triển khai M2 và consumer schema đích trên PostgreSQL test; cách chạy/bằng chứng/giới hạn xem [runbook M2](../scripts/authdb/README.md#datn-30--m2-bỏ-username-sequence-và-email-unique). Không có quyền áp Neon.
- [Bộ ticket tổ chức/tài khoản/hồ sơ](proposals/backend-organization-accounts-profile-tickets.md): 1 Epic/13 Task ánh xạ 15 WBS, Rank và 17 Blocks; một task bàn giao thiết kế đã làm, 12 task triển khai. Đã nhập Jira đợt OAP; tiến độ xem Jira, không dùng script auth hoặc apply lại để cập nhật tiến độ.
- [Công cụ Jira](../scripts/jira/README.md): diagnose/plan/apply/check khi được giao; token nhập ẩn. Đợt OAP dùng [script/hướng dẫn riêng](../scripts/jira/README-oap.md), không dùng apply auth hoặc journal auth. Không chạy import lại để cập nhật tiến độ.

## 3. Database — không đọc lại toàn bộ bước 1–5

- Điểm vào: [bản hợp nhất 33 bảng](design/database-proposal-2026-10-14.md).
- Cột/ràng buộc: [dictionary 5A](design/database-step5a-core-dictionary-2026-10-14.md), [dictionary 5B](design/database-step5b-supporting-dictionary-2026-10-14.md).
- Trực quan: [DBML](design/database-proposal-2026-10-14.dbml).
- Giải thích khi trình bày: [bản trình bày DB](design/database-pitch-2026-10-14.md); không phải bằng chứng đã triển khai DB.
- Khi chuẩn bị chuyển đổi, không phải khi sửa UI: [mapping 5C](design/database-step5c-migration-map-2026-10-14.md); mục 0 ghi kiểm kê Neon thực tế 06/11, schema drift và cổng trước migration. Có [SQL kiểm kê chỉ đọc](../scripts/neon-schema-inventory.sql). Neon vẫn version 1/22 bảng nghiệp vụ, chưa đồng bộ đích.
- Sổ quyết định ở mục 1 ưu tiên phần đã thay thế. Không sửa cấu trúc/chạy DB từ tác vụ harness.

## 4. Tài liệu tham khảo, không nạp mặc định

| Nhóm file hiện có | Vai trò và cách xử lý |
|---|---|
| `docs/design/database-step1*` đến `database-step4*`, `database-design-v2.md` | Lý giải và phương án trước hợp nhất; giữ đường dẫn cho truy vết, không là danh mục bảng hiện hành |
| `docs/design/recommendation-*`, `assignment-*`, `kpi-*` | Nghiên cứu/phương án/đánh giá; không tự chốt công thức hoặc tham số |
| `docs/reviews/` | Bằng chứng review tại thời điểm ghi, không là danh sách feature hiện đã chạy |
| Baseline trước 12/10 | Lịch sử quyết định; chỉ đối chiếu khi nguồn hiện hành dẫn tới |
| `docs/design/figma-make/launch-prompts.md` | Prompt thao tác Make cũ; không chạy lại để phát triển local |
| [Frontend readiness](design/frontend-readiness-2026-10-20.md), [nghiên cứu harness](design/agent-harness-research-2026-10-20.md) | Báo cáo trước pilot; hiện trạng mới xem web/README, không áp dụng câu “web rỗng” cũ |
| `docs/thesis-context/` | Báo cáo gốc chỉ đọc; không đồng nghĩa báo cáo đã đồng bộ nghiệp vụ mới |
| [Chỉ dẫn trước hợp nhất](archive/agent-context-before-harness-2026-10-20.md) | Bản bảo toàn nội dung CLAUDE.md chưa commit; không còn là instruction |

Không chuyển hàng loạt file đang được dẫn chiếu sang thư mục khác chỉ để ít file hơn; tránh hỏng liên kết báo cáo/DB. Phần dư thừa được bỏ khỏi luồng đọc và chỉ dẫn chủ động. Chỉ xóa tài liệu sau khi xác nhận không còn nội dung riêng hoặc bên dùng liên kết.

## 5. Vận hành tài liệu

[Kế hoạch triển khai và báo cáo thứ Năm đến 14/01/2027](ai/delivery-plan.md): đề xuất dựa trên 15–20 giờ/tuần, hiện trạng code và FR; chưa thay đổi phạm vi/nghiệp vụ hoặc xác nhận tỷ lệ hoàn thành.

- Nhận việc, branch/commit/PR và liên kết Jira: [workflow solo](ai/workflow.md), [mẫu brief/bàn giao](ai/task-template.md), [PR template](../.github/pull_request_template.md). Quy trình không thay nguồn nghiệp vụ; hướng học Go giữ phần người dùng tự viết. Chưa xác minh Development nhận liên kết trên Jira chỉ từ việc Git remote đã cấu hình.
- Thay quyết định đã chốt: hỏi người dùng; sửa nguồn đúng miền và phần UI/DB có liên quan, không thêm một bản sao “final-vN”.
- Mỗi đợt frontend cập nhật sổ công việc trong web/README; không tạo thêm báo cáo Markdown cho từng nút/icon.
- `node --test scripts/check-harness.test.mjs` chạy test harness không framework mới. `node scripts/check-harness.mjs` kiểm tra điểm vào, nguồn trực tiếp, mã nguồn/config frontend và nguồn Go/SQL trong cmd/internal/migrations/db/seed cùng go.mod/go.sum/Makefile tồn tại, không bị ignore. Không đọc `.env`, chạy server hoặc gọi DB. Thêm `--staged` để phát hiện file còn thiếu hoặc nội dung chưa được stage khi chuẩn bị bàn giao. Không tự stage/commit, không kiểm toán ngữ nghĩa mọi tài liệu lịch sử.
- Nguồn không bị ignore vẫn cần được review và commit khi người dùng yêu cầu; không cam kết clone có đủ chỉ vì tệp đang ở máy local.

**Tiếp nối auth 06/11/2026:** đã chốt cấp tài khoản mới bằng mật khẩu tạm qua email, buộc đổi trước nghiệp vụ. Xem quyết định tiếp nối trong `docs/business-baseline-2026-10-20.md` và mục 4.1 proposal auth; DBML/dictionary đã chốt thêm users.must_change_password; mật khẩu tạm chưa có hạn. Proposal mục 6 đã bổ sung API/phiên hạn chế/đổi lần đầu; mục 9–10 đề xuất 33 task, 105 giờ cơ sở (81 nền + 24 bổ sung), dự phòng 21 giờ cần duyệt. Các lựa chọn kỹ thuật A02–A09 đã được duyệt, còn phần triển khai cụ thể và điểm hoãn ở proposal; không coi thiết kế hoặc estimate là code đã chạy. Migration riêng chỉ thêm cờ, chưa cập nhật đủ 33 bảng hoặc triển khai gửi thư.
