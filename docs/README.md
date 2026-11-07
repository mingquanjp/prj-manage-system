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

[Proposal feature đầu — đăng nhập và phiên truy cập](proposals/backend-auth-session.md): theo khung template đã tham khảo, có mục tiêu/phạm vi, high-level design, detailed design (ER/cột/ràng buộc/index, API/DTO/lỗi/sequence, rollout và test), tác động, đánh giá, WBS và estimate cho D01 phối hợp D02. Người dùng đã duyệt A01–A09 (email/phiên/mật khẩu/CSRF, hướng test/migration); OpenAPI/ports đã được rà trong DATN-2 / AUTH-T01 ngày 07/11, kết luận và đầu vào task sau tại proposal mục 6.2.7; SQL chi tiết còn phải hoàn thiện. Chưa code auth; riêng migration cờ buộc đổi đã diễn tập nhánh tạm theo mapping 5C, chưa áp production.

Tiếp nối DATN-3 / AUTH-T02: đã thêm validation cấu hình thuần và table-driven tests trong `internal/config`; [contract và ranh giới bàn giao](proposals/backend-auth-session.md#631-bàn-giao-validation-cấu-hình--datn-3--auth-t02-07112026). Ghi nhận DATN-3 này được tiếp nối bởi DATN-4 / AUTH-T03: đã nối `LoadServer()` vào startup, chỉ đọc dotenv khi dev tường minh, không đổi OS env; `Load()` của CLI không yêu cầu cấu hình auth. Contract/default và giới hạn ở proposal mục 6.3.2; chưa có auth runtime hoặc bằng chứng DB thật.

Tiếp nối DATN-5 / AUTH-T04: parse URL PostgreSQL một host, chuyển đúng hostname Neon AWS pooled và nhóm lỗi kết nối an toàn; đầu vào AUTH-T17 tại proposal mục 6.3.3. Bằng chứng code/test được ghi theo lượt thực chạy trong mục đó; không là test DB thật, không thay thiết kế schema/API.

[OpenAPI auth](api/auth.openapi.json) là bản contract năm route, chưa có runtime; chữ ký service/repository và ma trận bằng chứng bảo mật AUTH-S01–S10 ở proposal mục 5.3/6.4.1. Kiểm tĩnh bằng `node --test scripts/check-auth-contract.test.mjs` để phát hiện ref thiếu, mất allowlist, nhầm cookie/CSRF AND thành OR, DTO lộ trường hoặc mất no-store. Đây không phải validator OpenAPI đầy đủ hoặc test bảo mật API thực tế; byte/transaction/Origin/CSRF vẫn phải kiểm ở server.

Kế hoạch SQL/DB test tại proposal mục 6.1.5–6.1.6: DDL auth soạn để review, migration tiến thêm, fixture giả và guard DB test; chưa chạy SQL/container hoặc áp Neon. Review contract ban đầu ở mục 6.2.6, bàn giao AUTH-T01 ở mục 6.2.7; đề xuất Jira solo ở mục 9.4.

WBS trong proposal đã review 07/11: giữ 33 mã task/105 giờ cơ sở, phân biệt bản soạn đã có với phần còn lại (mục 9.2.1), tách kế hoạch fixture khỏi thực thi sau migration, giao rõ limiter login/đổi lần đầu và ranh giới FE auth/mock. Mục 9.2.2 nối AUTH-S01–S10 với task chịu trách nhiệm; mục 9.4 có checklist Done cho ticket theo board DATN đã tạo. Chưa tạo issue hoặc đánh dấu runtime hoàn tất; estimate gốc không phải giờ còn lại.

[Bản nháp ticket auth](proposals/backend-auth-session-tickets.md): 1 Epic đề xuất và 20 Task theo đầu ra, mapping đủ 33 đơn vị WBS; review nằm trong Done, description có bước thực hiện và bảng tình huống/kết quả; có thứ tự nhận việc đợt đầu và cổng trước khi tạo Jira. Đây là nội dung để review, chưa upload, không là nguồn nghiệp vụ thứ hai.

Công cụ bàn giao Jira: [hướng dẫn plan/apply/check](../scripts/jira/README.md), đọc nội dung trực tiếp từ bộ ticket; offline sinh manifest/kiểm thử bằng Jira giả, online do người dùng tự nhập token ẩn trên terminal. Ngày 07/11 người dùng đã chạy apply/check, log được đối chiếu: Epic DATN-1 và 20 Task DATN-2–DATN-21, 47 Blocks, Rank/nội dung/parent/assignee/labels khớp. Các câu chưa tạo Jira trong ghi nhận trước đó là trạng thái trước import, không phải hiện trạng; không chứng minh auth runtime. Không tự stage/commit/push hoặc thay DB.

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

- Nhận việc, branch/commit/PR và liên kết Jira: [workflow solo](ai/workflow.md), [mẫu brief/bàn giao](ai/task-template.md), [PR template](../.github/pull_request_template.md). Quy trình không thay nguồn nghiệp vụ; hướng học Go giữ phần người dùng tự viết. Chưa xác minh Development nhận liên kết trên Jira chỉ từ việc Git remote đã cấu hình.
- Thay quyết định đã chốt: hỏi người dùng; sửa nguồn đúng miền và phần UI/DB có liên quan, không thêm một bản sao “final-vN”.
- Mỗi đợt frontend cập nhật sổ công việc trong web/README; không tạo thêm báo cáo Markdown cho từng nút/icon.
- `node --test scripts/check-harness.test.mjs` chạy test harness không framework mới. `node scripts/check-harness.mjs` kiểm tra điểm vào, nguồn trực tiếp, mã nguồn/config frontend và nguồn Go/SQL trong cmd/internal/migrations/db/seed cùng go.mod/go.sum/Makefile tồn tại, không bị ignore. Không đọc `.env`, chạy server hoặc gọi DB. Thêm `--staged` để phát hiện file còn thiếu hoặc nội dung chưa được stage khi chuẩn bị bàn giao. Không tự stage/commit, không kiểm toán ngữ nghĩa mọi tài liệu lịch sử.
- Nguồn không bị ignore vẫn cần được review và commit khi người dùng yêu cầu; không cam kết clone có đủ chỉ vì tệp đang ở máy local.

**Tiếp nối auth 06/11/2026:** đã chốt cấp tài khoản mới bằng mật khẩu tạm qua email, buộc đổi trước nghiệp vụ. Xem quyết định tiếp nối trong `docs/business-baseline-2026-10-20.md` và mục 4.1 proposal auth; DBML/dictionary đã chốt thêm users.must_change_password; mật khẩu tạm chưa có hạn. Proposal mục 6 đã bổ sung API/phiên hạn chế/đổi lần đầu; mục 9–10 đề xuất 33 task, 105 giờ cơ sở (81 nền + 24 bổ sung), dự phòng 21 giờ cần duyệt. Các lựa chọn kỹ thuật A02–A09 đã được duyệt, còn phần triển khai cụ thể và điểm hoãn ở proposal; không coi thiết kế hoặc estimate là code đã chạy. Migration riêng chỉ thêm cờ, chưa cập nhật đủ 33 bảng hoặc triển khai gửi thư.
