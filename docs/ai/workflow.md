# Làm việc solo với Jira, GitHub và agent

Jira giữ việc cần làm, phụ thuộc và trạng thái. GitHub giữ code, PR và nhận xét review. Nghiệp vụ/contract hiện hành được tìm từ [mục lục nguồn](../README.md); ticket, plan và code không ghi đè các quyết định đã chốt.

## Từ ticket đến bàn giao

1. Đọc ticket, nguồn áp dụng và các việc `is blocked by`. Nhận việc theo Rank khi đủ đầu vào; Rank không thay quan hệ phụ thuộc. Có thể chuẩn bị phần độc lập nhưng không đánh dấu hoàn thành khi đầu vào còn thiếu.
2. Ghi ngắn mục tiêu, nguồn, ngoài phạm vi và tiêu chí đạt. Dùng [task brief](task-template.md) khi cần, không bắt mỗi sửa nhỏ tạo thêm tài liệu. Chuyển Jira sang **In Progress** khi thực sự bắt đầu; agent chỉ cập nhật Jira khi được giao.
3. Một nhánh cho một đầu ra có thể review. Kiểm working tree trước; nếu còn thay đổi đợt cũ thì tách/bàn giao theo chỉ dẫn của người dùng, không tự stash/reset hoặc mang tất cả vào PR mới. Nhánh tạo khi đang dirty không tự cô lập các thay đổi đó.
4. Triển khai theo nguồn, chạy kiểm tra đúng phạm vi, tự xem diff và thử luồng nếu đổi UI. Với phần Quân tự viết Go, agent giải thích đầu vào/luồng xử lý, hướng dẫn, gỡ lỗi và review; không viết thay checkpoint đã giao nếu chưa được yêu cầu.
5. Khi đầu ra sẵn sàng review, mở PR theo [template](../../.github/pull_request_template.md), gắn link trên Jira và chuyển **In Review**. Tác giả đọc và hiểu code; agent review hỗ trợ, không được mô tả thành một người phê duyệt độc lập trên GitHub.
6. **Done** khi đủ AC, kiểm tra phù hợp đạt, góp ý đã xử lý và PR liên quan đã merge. Với task tài liệu/môi trường, bằng chứng đúng loại đầu ra thay test runtime. Solo không tự đặt điều kiện phải có reviewer thứ hai; nếu đã có required checks/rules trên GitHub thì tuân thủ, không tự bỏ qua. Không tự merge hoặc chuyển Done thay người dùng.

Jira cập nhật thủ công; không giả định mở PR/merge sẽ tự chuyển trạng thái. Khi bị block, ghi rõ vướng gì và cần đầu vào nào; dùng quan hệ Blocks đã có, không cần thêm một trạng thái hoặc workflow mới. Comments lưu tiến độ/bằng chứng, description giữ phạm vi/AC. Không bắt solo báo cáo hằng ngày hoặc gửi chat nhóm.

## Tên nhánh, commit và PR

| Nơi | Quy ước / ví dụ |
|---|---|
| Feature, tài liệu, refactor | `feature/DATN-2-auth-contract` |
| Sửa lỗi | `bug/DATN-13-csrf-origin-check` |
| Commit | `docs(auth): finalize authentication contract [DATN-2]` |
| Tiêu đề PR | `docs(auth): finalize authentication contract [DATN-2]` |

Ví dụ chỉ minh họa format, không khẳng định ticket nào đã hoàn thành hoặc có lỗi. Dùng **mã Jira thật** trong branch/commit/PR; `AUTH-T01` hay mã WBS không thay `DATN-2`. Không tạo GitHub Issue trùng task. Không dùng `Closes #2` để chỉ Jira DATN-2; `#2` là GitHub Issue/PR số 2.

Commit và tiêu đề PR bằng tiếng Anh: `<type>(<scope>): <imperative subject> [<Jira-key>]`. Type theo thay đổi: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`, `ci`, `build`, `perf`; scope cụ thể như `auth`, `config`, `database`, `web`, `workflow`, `jira`. Subject bắt đầu chữ thường, không chấm cuối. Body commit nếu có cũng tiếng Anh; attribution theo [AGENTS](../../AGENTS.md), không bịa danh tính.

Một nhánh/PR thường ứng với một ticket; commit chia theo đầu ra có thể review, không chia máy móc từng file. Nhiều ticket thực sự chung một đầu ra thì dẫn đủ trong PR và giải thích phạm vi; không nhét toàn bộ Epic vào một PR mặc định. Chưa có ticket phù hợp thì bỏ mã, ví dụ `feature/jira-github-workflow`; không mượn mã task auth cho việc công cụ không nằm trong AC. Không đổi tên các nhánh cũ chỉ để đồng bộ format.

PR description tiếng Việt: giữ What, Why, Environment Variables, Developer Test; How/Limits chỉ khi hữu ích. Nêu hành vi cuối cùng và lý do, không chép nguyên lịch sử chat hay toàn bộ ticket. Khi squash, kiểm title/message cuối giữ đúng mã Jira. Việc commit, push, mở PR và merge cần được người dùng giao; chấp thuận sửa code không tự bao gồm các bước này.

### Viết description PR cho người review

**Chốt theo góp ý của Quân ngày26/11/2026:** người đọc PR có thể chưa đọc ticket hoặc chat. Description phải giúp họ hiểu thay đổi và đánh giá diff; không biến thành nhật ký triển khai hoặc bản chép evidence.

- **What:** mở đầu bằng chức năng/hành vi sau thay đổi. Với bug, nêu tình huống gây lỗi và kết quả sau sửa. Chỉ liệt kê các thay đổi chính, kèm link Jira.
- **Why:** giải thích vấn đề và tác động bằng lời cụ thể. Không chỉ ghi “hoàn thành ticket” hoặc liệt kê mã FR/WBS.
- **How:** chỉ giữ lựa chọn kỹ thuật, ranh giới hoặc đánh đổi cần người review chú ý. Thuật ngữ, tên trường, package và helper chỉ xuất hiện khi giúp hiểu diff; tránh chuỗi viết tắt không giải thích.
- **Environment Variables:** nêu biến mới/thay đổi và mục đích; không có thì ghi ngắn. Cách dựng môi trường chi tiết dẫn về runbook.
- **Developer Test:** tóm tắt các cổng và luồng thực sự đã kiểm cùng kết quả. Phân biệt kiểm tự động, browser/DB thật và kiểm thủ công do Quân xác nhận; không suy từ review/unit thành runtime. Lệnh dài, đường dẫn worktree tạm, thời lượng từng package, log và ma trận đầy đủ để ở runbook/evidence có link.
- **Limits:** chỉ ghi giới hạn ảnh hưởng đến phạm vi, sử dụng hoặc đánh giá thay đổi. Không lặp toàn bộ ngoài phạm vi của Epic, danh sách mọi điều chưa test hoặc lời nhắc quy trình Git/Jira thông thường.
- Độ dài theo độ phức tạp: PR nhỏ thường chỉ cần vài câu và kiểm thử; PR lớn dùng đoạn/bullet ngắn để quét được. Không điền How/Limits cho đủ template, không kể lỗi công cụ, các lần thử, lịch sử review hoặc phương án đã bỏ nếu chúng không giải thích một giới hạn hiện hành.

Trước khi tạo **hoặc cập nhật** PR, đọc lại như người chưa xem chat: có hiểu ngay **đã đổi gì, vì sao, kiểm bằng cách nào và còn giới hạn gì đáng chú ý** không? Bỏ câu chỉ mô tả quá trình làm; chuyển bằng chứng chi tiết về runbook; kiểm lại link và nguồn xác nhận. Duy trì bước này ở mọi lần bàn giao, không chỉ đợt vừa được góp ý.

## Liên kết Jira–GitHub

DATN dùng repo `mingquanjp/ai-staffing` và project **DATN**, board **1** trên site Jira Free riêng. Khi gặp liên kết DATN trên site cũ, mở cùng mã ở site mới, ví dụ [DATN-12](https://datn-mingquanjp.atlassian.net/browse/DATN-12). Mọi thao tác DATN dùng [board DATN](https://datn-mingquanjp.atlassian.net/jira/software/projects/DATN/boards/1).

Người dùng đã đối chiếu đủ 1 Epic và 20 ticket ở site mới. Trước khi chạy script ghi, xác minh base URL, đọc thử DATN-12 và lấy các ID cần thiết từ metadata của site mới. Dùng chế độ `diagnose` cho bước đọc; import/check phục vụ bàn giao ban đầu. Xem [hướng dẫn Jira](../../scripts/jira/README.md).

Tích hợp GitHub trên site mới chưa được xác minh. Sau lần push/mở PR có mã thật, mở ticket trên site mới → **Development** để đối chiếu link trỏ đúng repo và PR. Nếu chưa hiện, kiểm phạm vi repo được kết nối, quyền hiện có và mã ticket, rồi chờ đồng bộ.

Tham chiếu: [hướng dẫn chính thức liên kết GitHub với Jira bằng mã ticket](https://support.atlassian.com/jira-cloud-administration/docs/use-the-github-for-jira-app/). Quy tắc liên kết không tự cấp quyền và không thay cấu hình automation chuyển trạng thái.

## Kiểm tra theo phạm vi

Chuẩn bị một đợt feature mới hoặc tiếp nối bộ thiết kế: dùng [skill datn-feature-planning](../../.agents/skills/datn-feature-planning/SKILL.md) và [hướng dẫn kiểm](../../.agents/skills/datn-feature-planning/references/checks.md). Skill giữ hai lượt review proposal/bộ bàn giao, không tự coi thiết kế là runtime hoặc cấp quyền Jira/Git/DB. Đầu ra đã duyệt dùng lại, chỉ review lại phần bị thay đổi; không áp cả quy trình này cho một ticket code đã có brief.

| Thay đổi | Kiểm tra |
|---|---|
| Tài liệu/harness | `node --test scripts/check-harness.test.mjs`, `node scripts/check-harness.mjs`, `git diff --check`; review ngữ nghĩa và nguồn |
| Contract auth | Thêm `node --test scripts/check-auth-contract.test.mjs`; kiểm tĩnh không chứng minh server thực thi contract |
| Go | Kiểm môi trường rồi `go vet ./...`, `go test ./... -count=1`, `go build ./...`; DB test chỉ dùng môi trường riêng đã được chốt/cấp |
| Frontend | Tại `web/`: `pnpm check` và kiểm UI/luồng phù hợp; lệnh/mức kiểm tại [web README](../../web/README.md) |
| Công cụ Jira | Thêm `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s scripts/jira -p 'test_*.py' -v`; test giả không chứng minh API thật |

Chỉ khi chuẩn bị commit theo yêu cầu: review staged diff, `git diff --cached --check`, `node scripts/check-harness.mjs --staged`; tự stage đúng phạm vi trước. Harness kiểm kê nguồn và một số liên kết, không kiểm chứng mọi nghiệp vụ/quyền/transaction. Không dùng `make check` như lệnh chỉ đọc vì fmt/tidy có thể sửa file. Không đưa secret hoặc dữ liệu cá nhân thật vào evidence; không đọc `.env` để chuẩn bị PR.

Hiện đợt workflow này chỉ thêm quy ước/template và nối kiểm kê harness; không cài hook/commitlint, CI hoặc branch protection, không xác nhận cấu hình remote. Nếu bổ sung tự động hóa sẽ chốt phạm vi riêng, dùng lệnh thực của repo thay quy trình Node/media của CoLearnia.

**Bàn giao công cụ feature planning 21/11/2026:** đã thêm skill trong repo và `scripts/check-feature-plan.mjs` chỉ đọc, nhận file đầu vào thay vì manifest sao chép nghiệp vụ. Thử PCA đạt10 operation planned/12 WBS/12 task/17 cạnh/6 Mermaid; thử lại OAP đạt16 operation/15 WBS/13 task/17 cạnh/4 Mermaid. Parser lấy từ bundle VS Code local; chỉ kiểm grammar, chưa kiểm visual render. Lượt sub-agent thử tiếp nối và invalid-input đã phát hiện checker bỏ qua mã WBS/heading sai; đã sửa validate toàn token và sub-agent kiểm lại xác nhận khép. Bộ13 tests riêng đạt, không skip trên máy này; kiểm tổng hợp và giới hạn môi trường bàn giao trong chat. Validator Python của skill-creator không chạy được vì thiếu PyYAML; frontmatter/name/description/placeholders và interface YAML đã kiểm bằng Ruby Psych hiện có, không cài dependency. Chưa gọi DB/Jira hoặc commit/push. Công cụ không tự ghi nhận một bộ tài liệu đã được người dùng duyệt.

**Đối chiếu nguồn chính thức và kiểm hành vi 21/11:** [cơ sở/ca kiểm/kết quả skill](../../.agents/skills/datn-feature-planning/references/evaluation.md) dẫn Anthropic, OpenAI và Agent Skills. Đã bổ sung điều kiện kích hoạt, điểm vào theo trạng thái và evals chỉ nạp khi sửa/đánh giá skill. Hai sub-agent mới thử8 prompt routing và3 tình huống hành vi, không được đưa đáp án/rubric; agent chính đã đối chiếu kết quả. Đây là smoke test, chưa có A/B hoặc số đo tiết kiệm token/thời gian; không suy lợi ích định lượng từ60 tests công cụ/contract.
