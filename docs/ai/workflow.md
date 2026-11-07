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

## Liên kết Jira–GitHub

Đích của DATN: repo `mingquanjp/ai-staffing`, site `faker1652005.atlassian.net`, project **DATN**, board **2**. Không dùng project SCRUM/board 1 của CoLearnia. Board DATN đã có ticket auth; import/check là công cụ bàn giao ban đầu tại [hướng dẫn Jira](../../scripts/jira/README.md), không dùng để cập nhật tiến độ thường ngày hoặc ghi đè ticket sau khi làm.

Người dùng đã báo kết nối GitHub; chưa có bằng chứng Development của DATN đã nhận branch/commit/PR. Sau lần push/mở PR có mã thật, mở ticket → **Development** để đối chiếu link trỏ đúng repo và PR. Nếu chưa hiện, kiểm integration có bao gồm repo, quyền hiện có và mã ticket, chờ đồng bộ; không tự tạo token/nới quyền. Git remote đúng không đủ chứng minh integration đã hoạt động. Không cần tạo commit giả hoặc push chỉ để thử kết nối.

Tham chiếu: [hướng dẫn chính thức liên kết GitHub với Jira bằng mã ticket](https://support.atlassian.com/jira-cloud-administration/docs/use-the-github-for-jira-app/). Quy tắc liên kết không tự cấp quyền và không thay cấu hình automation chuyển trạng thái.

## Kiểm tra theo phạm vi

| Thay đổi | Kiểm tra |
|---|---|
| Tài liệu/harness | `node --test scripts/check-harness.test.mjs`, `node scripts/check-harness.mjs`, `git diff --check`; review ngữ nghĩa và nguồn |
| Contract auth | Thêm `node --test scripts/check-auth-contract.test.mjs`; kiểm tĩnh không chứng minh server thực thi contract |
| Go | Kiểm môi trường rồi `go vet ./...`, `go test ./... -count=1`, `go build ./...`; DB test chỉ dùng môi trường riêng đã được chốt/cấp |
| Frontend | Tại `web/`: `pnpm check` và kiểm UI/luồng phù hợp; lệnh/mức kiểm tại [web README](../../web/README.md) |
| Công cụ Jira | Thêm `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s scripts/jira -p 'test_*.py' -v`; test giả không chứng minh API thật |

Chỉ khi chuẩn bị commit theo yêu cầu: review staged diff, `git diff --cached --check`, `node scripts/check-harness.mjs --staged`; tự stage đúng phạm vi trước. Harness kiểm kê nguồn và một số liên kết, không kiểm chứng mọi nghiệp vụ/quyền/transaction. Không dùng `make check` như lệnh chỉ đọc vì fmt/tidy có thể sửa file. Không đưa secret hoặc dữ liệu cá nhân thật vào evidence; không đọc `.env` để chuẩn bị PR.

Hiện đợt workflow này chỉ thêm quy ước/template và nối kiểm kê harness; không cài hook/commitlint, CI hoặc branch protection, không xác nhận cấu hình remote. Nếu bổ sung tự động hóa sẽ chốt phạm vi riêng, dùng lệnh thực của repo thay quy trình Node/media của CoLearnia.
