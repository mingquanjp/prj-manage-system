# Đưa ticket tổ chức, tài khoản và hồ sơ lên Jira DATN

Script: `scripts/jira/oap_import.py`. Nguồn: [bộ ticket OAP](../../docs/proposals/backend-organization-accounts-profile-tickets.md), [proposal](../../docs/proposals/backend-organization-accounts-profile.md), [OpenAPI](../../docs/api/organization-accounts-profile.openapi.json), [SQL plan](../../docs/proposals/backend-organization-accounts-profile-sql-plan.md). Script sinh nội dung từ nguồn, không có bản description chép tay trong Python.

**Đích cố định:** `https://datn-mingquanjp.atlassian.net`, project DATN, board **1**. Đợt OAP gồm **1 Epic + 13 Task, 15 WBS và 17 Blocks**. Không chạy `auth_import.py --mode apply` để nhập OAP, không sửa/xóa journal auth. Dùng Python 3.9+ và thư viện chuẩn, không cần cài package mới.

## 1. Kiểm offline

Chạy từ thư mục gốc repo:

```sh
python3 scripts/jira/oap_import.py --mode plan --offline
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s scripts/jira -p 'test_*.py' -v
```

Không mạng/credential/journal. Kiểm đủ 13 Summary/description, 15 WBS một lần, 17 Blocks hai chiều, Rank hợp phụ thuộc; chuyển bảng tình huống hai cột, bảng endpoint ba cột và checklist sang ADF; lệnh kiểm áp dụng có trong từng description. Task T01 là bàn giao thiết kế đã có, không task review mới hoặc tự chuyển Done. Không gửi estimate/status/priority/sprint/deadline/remaining.

Muốn đọc toàn manifest sinh từ nguồn, thêm `--manifest-out /tmp/datn-oap-manifest.json` (chọn tên mới nếu đã có). File được tạo mới mode0600, không ghi đè và không xuất vào `.local` của repo; không chỉnh tay hoặc commit như nguồn thứ hai.

## 2. Diagnose và plan trên site thật — chỉ đọc

Nhập email Atlassian và token ở terminal tương tác; token được hỏi bằng prompt ẩn. Không gửi token vào chat/command argument/file Git. Script không nhận token qua biến môi trường hoặc đọc .env; stdin/stderr không phải TTY thì từ chối.

```sh
python3 scripts/jira/oap_import.py --mode diagnose
python3 scripts/jira/oap_import.py --mode plan
```

Diagnose xác minh `serverInfo.baseUrl`, đọc project của DATN-12 rồi metadata board1/configuration/features; không cần manifest/journal. Plan kiểm tiếp filter DATN/Rank, quyền hiện có, create metadata Epic/Task/parent, assignee và nhãn link Blocks. Project/type/Rank/link IDs lấy từ metadata, không dùng ID site cũ. Board `simple` chỉ nhận khi feature Sprint xác nhận DISABLED. Thiếu metadata/quyền hoặc filter chưa hỗ trợ thì dừng, không tự sửa project/board/quyền.

Plan đọc toàn DATN để đối soát mã/label nhưng chỉ nhận ownership OAP; in issue hiện có/còn thiếu, parent, Rank mong muốn và phụ thuộc. **Không ghi Jira hoặc lưu mapping mới.** Online plan/check có lock tạm `.local/jira-oap/state.lock` để không chạy trùng với apply; lock được dọn khi kết thúc, không là journal đã import.

Với token có scope, thêm `--cloud-id UUID_SITE_DATN_MỚI` vào mọi lệnh online. API đi qua `https://api.atlassian.com/ex/jira/<cloudId>` nhưng vẫn kiểm `serverInfo.baseUrl` đúng site DATN. Không suy cloud ID từ domain hoặc dùng cloud ID site cũ. Token thường dùng site URL trực tiếp. 401/403: kiểm token còn hạn/scope và quyền hiện có, không cấp thêm quyền tự động.

## 3. Link nguồn và apply

Nguồn chưa commit/push: description giữ đường dẫn repo và “link remote chờ bàn giao”; không tạo link commit giả. Có thể plan/apply bản này, nhưng nên bàn giao nguồn trước để ticket có link mở được.

Nếu đã commit/push nguồn, thêm cùng một giá trị vào plan/apply/check:

```text
--source-base https://github.com/mingquanjp/ai-staffing/blob/FULL_COMMIT_SHA
```

Thay bằng SHA40 ký tự thật. Script kiểm sáu nguồn local khớp commit; không tự push hoặc chứng minh commit đã có trên remote. Giữ nguyên source-base và snapshot nguồn trong một lượt import; đổi sau khi apply dở sẽ dừng để review journal. Sau khi đã nhập và bắt đầu làm, không chạy import để cập nhật tiến độ/đồng bộ lại nguồn.

Chỉ sau khi đọc plan:

```sh
python3 scripts/jira/oap_import.py --mode apply
python3 scripts/jira/oap_import.py --mode check
```

Apply chạy lại preflight/đối soát và hỏi **YES**. Nó tạo phần thiếu, gắn Task vào Epic, assignee tài khoản xác thực, labels riêng OAP, 17 link provider **blocks** consumer và xếp **13 Task OAP** theo Rank đã duyệt. Epic không thuộc thứ tự task. Không gửi key auth vào Rank hoặc sửa field/link auth; vị trí OAP so với ticket khác có thể thay đổi, nhưng không chủ động xếp lại thứ tự tương đối các ticket khác. Không đổi status kể cả T01 đã có đầu ra, không xóa/ghi đè issue lệch nội dung.

Chỉ ba loại lệnh ghi: `POST issue`, `POST issueLink`, `PUT issue/rank`. Không gửi email/comment, chuyển trạng thái, update issue, sprint, estimate hoặc permission. Jira có thể áp default khi tạo; script không nói các field không gửi sẽ trống.

Check chỉ đọc, xác minh 14 issue đúng nội dung/URL nguồn/cấu trúc bảng/type/parent/assignee/labels, đủ đúng17 Blocks và Rank13 Task. Đây là đối soát import, không chứng minh API ứng dụng/DB/mail/browser đã chạy. Khi có key thật mới dùng key đó trong branch/commit/PR.

## 4. Journal và phục hồi sau lỗi

- Journal cố định **`.local/jira-oap/state.json`**, mode0600 và Git-ignore; lock độc quyền cho lượt online. Không có `--state` để vô tình dùng journal auth; từ chối symlink trên đường dẫn journal/lock. Fingerprint gắn site/board/đợt, nguồn và source-base; mapping chỉ nhận OAP-EPIC/OAP-T01–13.
- Trước mỗi mutation ghi pending atomically. Mất phản hồi create: tìm mã/label và đọc lại, nhận lại chỉ khi đúng nội dung; nếu chưa thấy do index/quyền thì dừng, không POST lại mù. Sau khi tự xác minh issue chưa tồn tại mới dùng `--mode apply --retry-create OAP-Txx` và xác nhận **RETRY OAP-Txx**, rồi YES. Epic dùng OAP-EPIC.
- Mất phản hồi link: đọc hai đầu; đã có thì không tạo thêm. Chưa có phải xác minh rồi xác nhận **RETRY LINK** ở lượt apply mới. Mỗi cặp có một link; trùng/đảo chiều/dư/ngoài manifest dừng để review.
- Mất phản hồi Rank, kể cả HTTP207: không retry trong cùng lượt. Plan đọc lại; lượt apply mới dựa thứ tự hiện tại và xác nhận YES. Snapshot lưu thứ tự tương đối13 Task trước lần rank đầu, không đủ restore toàn board.
- Nguồn/metadata/account thay đổi, nội dung sửa tay, duplicate hoặc journal sai: dừng, không ghi đè/xóa hoặc đổi fingerprint để vượt kiểm. Giữ state và đối soát. Lock cũ chỉ gỡ sau xác minh không còn process chạy; không xóa journal để chạy lại.
- Script chặn redirect để credential không chuyển host; lỗi HTTP chỉ in status, không body/header/token. Không hứa exactly-once trước search index trễ, journal mất hoặc chạy song song bằng công cụ khác.

## 5. Phạm vi kiểm thử

Unit/orchestration dùng Jira giả: nguồn thật→manifest/ADF, plan/check không ghi, apply/read-back/rerun, cùng tồn tại21 auth+14 OAP, Rank đúng phạm vi, duplicate/drift/URL sai/Blocks đảo hoặc trùng, pending create/link/Rank, journal auth không được nhận, export không ghi đè/symlink và input token TTY. Bộ auth hiện hành giữ nguyên, bao gồm transport/metadata/credential redaction/redirect/HTTP207.

Chưa chạy API Jira thật trong đợt chuẩn bị này. Người chạy cần diagnose/plan/check để chứng minh metadata/nội dung trên site thật; fake tests không thay bước đó.

Nguồn chính thức đối chiếu: [issue/create metadata](https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issues/), [issue links](https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issue-links/), [Rank/HTTP207](https://developer.atlassian.com/cloud/jira/software/rest/api-group-issue/). Quy ước Git/Jira theo [workflow](../../docs/ai/workflow.md).

**Kiểm chứng offline 15/11/2026:** 52/52 Python tests đạt (30 auth, 22 OAP), 32/32 contract/harness tests đạt, harness kiểm kê315 nguồn và diff check đạt; offline plan render đủ14 description ADF. Sub-agent review đã phát hiện và kiểm lại bản sửa kiểm URL nguồn/cấu trúc bảng; không còn blocker được báo. Không gọi API Jira thật hoặc ghi issue/link/Rank; chưa commit/push.
