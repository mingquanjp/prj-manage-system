# Đưa ticket auth lên Jira DATN

Nguồn nội dung: [bộ ticket đã review](../../docs/proposals/backend-auth-session-tickets.md), đối chiếu [proposal/WBS](../../docs/proposals/backend-auth-session.md). Script không chứa bản sao description: mỗi lần chạy sinh manifest từ nguồn và kiểm mapping 33 WBS, 105 giờ cơ sở, 20 Task và hai chiều của 47 phụ thuộc.

Đích cố định: `https://faker1652005.atlassian.net`, project `DATN`, board Kanban `2`. Không thao tác CoLearnia, sprint hoặc DB của đồ án. Dùng Python **3.9+**, thư viện chuẩn; không cài dependency, không đọc `.env`.

## 1. Kiểm tra offline trước

Từ thư mục gốc repo:

```sh
python3 scripts/jira/auth_import.py --mode plan --offline
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s scripts/jira -p 'test_*.py' -v
```

Không mạng/credential. Muốn xem toàn nội dung manifest sinh tự động, thêm `--manifest-out /tmp/datn-auth-manifest.json` (chọn tên mới nếu đã có). File này chứa nội dung ticket và hash nguồn; không chỉnh tay hoặc commit như nguồn thứ hai.

Thứ tự nhận việc/Rank: T01 → T02 → T03 → T04 → T05 → T06 → T07 → T08 → T09 → T10 → T11 → T12 → T13 → **T15 → T14** → T16 → T17 → T18 → T19 → T20. Không đổi ID để ép thứ tự số. Đây là thứ tự mặc định hợp phụ thuộc, không phải gán mọi task High; các checkpoint học API có thể làm sớm khi đủ đầu vào như bộ ticket nêu.

## 2. Plan trên Jira thật — chỉ đọc

Tự chạy trong terminal tương tác:

```sh
python3 scripts/jira/auth_import.py --mode plan
```

Nhập email Atlassian và token khi được hỏi; token nhập ẩn bằng `getpass`, chỉ ở bộ nhớ. Không gửi token vào chat, command argument, file, ảnh hoặc Git. Nếu terminal không ẩn được input, script từ chối tiếp tục. 401/403 không tự cấp thêm quyền: kiểm tài khoản/token/scope và quyền hiện có của DATN.

Mặc định API gọi trực tiếp site bằng email/token. Nếu token có scope yêu cầu gateway `api.atlassian.com`, thêm `--cloud-id UUID_SITE_ĐÃ_XÁC_MINH` vào **mọi lệnh online**. Cloud ID không phải project key; không đoán hoặc lấy từ CoLearnia khác site. Không tự tạo token mới/nới scope. Hạn chế redirect để credential không chuyển sang host khác.

Plan kiểm project, board/filter Rank, create metadata Epic/Task/parent/trường bắt buộc, tài khoản hiện tại có assign được và quyền browse/create/assign/link/edit/rank; xác định Blocks theo nhãn outward/inward. Assignee là tài khoản đang xác thực, cần dùng tài khoản Quân đã chốt. In Epic/Task đã có/cần tạo, parent và phụ thuộc; không thay Jira, không ghi mapping mới. Token/API có thể chưa đủ quyền dù đọc được UI.

Board team-managed có thể trả API `type=simple` thay vì `kanban`. Script nhận `simple` chỉ khi endpoint features xác nhận Sprint `DISABLED`, qua `boardFeature=SPRINTS` hoặc đúng khóa `feature=jsw.agility.sprints`; không suy Kanban chỉ từ tên board. Vẫn kiểm ID board/config bằng 2 và filter đúng project DATN. Không tự bật/tắt Sprint; thiếu quyền đọc features hoặc trạng thái chưa rõ thì dừng. Xem [ghi nhận chính thức của Atlassian](https://jira.atlassian.com/browse/JRACLOUD-85696).

Nếu dừng vì thiếu mục `SPRINTS`, chưa đủ cơ sở để kết luận board là Scrum hay Kanban. Chạy một lượt chẩn đoán:

```sh
python3 scripts/jira/auth_import.py --mode diagnose
```

Lệnh chỉ GET board 2, configuration và features; in chọn lọc ID/type/project/filter ID/Rank field cùng `boardFeature`, `feature`, `featureId`, `boardId`, `state` và tên các trường trong từng mục. Không in response thô, token, email, tên người hoặc mô tả; không đọc bộ ticket, ghi journal hay thay Jira. Có thể gửi phần JSON đầu ra để review định dạng thực tế. Thiếu feature không được coi là `DISABLED`; chưa biết định dạng thì không bỏ chốt hoặc đoán từ `featureId`. Bộ đọc nhận hai định danh Sprint nêu trên, từ chối định danh mâu thuẫn, mục Sprint trùng hoặc trạng thái chưa rõ. Không dùng `featureId` hoặc nhãn dịch để suy Sprint. Unit test dùng response giả, không chứng minh site thật luôn trả một trong hai định dạng; chẩn đoán là bằng chứng cần bổ sung.

Script hiện chỉ hỗ trợ filter dạng `project = DATN ORDER BY Rank ASC` (hoặc ID project/field Rank tương đương). Filter phức tạp hơn dừng để review; không tự sửa board/filter hoặc giả task xuất hiện trên board.

## 3. Link nguồn trước khi tạo

Nguồn chưa commit/push: script chuyển link local thành đường dẫn repo kèm **“link remote chờ bàn giao”**, không tạo URL giả. Có thể xem/rà bản này ở plan; nên bàn giao nguồn trước apply để ticket có link dùng được.

Nếu nguồn đã commit và push, thêm vào mọi lệnh plan/apply/check cùng một giá trị:

```text
--source-base https://github.com/mingquanjp/ai-staffing/blob/FULL_COMMIT_SHA
```

Thay `FULL_COMMIT_SHA` bằng SHA 40 ký tự thật. Script kiểm bốn nguồn chính ở commit local khớp file đang đọc; **không tự push hoặc chứng minh commit đã có trên remote**. Người chạy phải xác nhận commit đã push và mở được URL. Không đổi source-base giữa lượt đang dở; fingerprint khác sẽ dừng. Không thêm estimate vào Jira: số giờ chỉ nằm trong description để truy vết, không phải remaining/timesheet.

## 4. Apply có xác nhận trên terminal

Chỉ sau khi đã xem plan:

```sh
python3 scripts/jira/auth_import.py --mode apply
```

Script lại kiểm đầu vào rồi hỏi `YES`. Nó tạo 1 Epic, 20 Task theo Rank đã chốt, parent, assignee và labels `auth-session`/`datn-auth-v1`/`source-auth-…`; thêm `learn-go` cho T14/T15/T16. Description chuyển sang ADF: tiêu đề, danh sách/checklist và bảng tình huống. Checkbox là ký hiệu tiêu chí chưa hoàn thành, không phải workflow Done.

Sau khi có Jira key, tạo 47 link: **provider blocks consumer**. Trên provider, Jira trả consumer tại `outwardIssue`; UI của consumer hiển thị **is blocked by**. Chỉ tạo một link mỗi cặp, không tạo comment đi kèm. Script kiểm chiều trước và đọc lại sau ghi.

Chỉ có ba endpoint ghi được phép: tạo issue, tạo issueLink, và Rank. Không sửa ticket đã có, không xóa, chuyển trạng thái, gửi email/comment, đổi priority/sprint/deadline/estimate. Jira có thể áp default của project cho field không gửi; script không tuyên bố các field đó trống. Rank chỉ gửi key của 20 Task; Epic không nằm trong thứ tự task. Thứ tự tương đối giữa các ticket khác không được script chủ động xếp lại, nhưng vị trí của chúng so với các Task được di chuyển có thể đổi.

Mapping, fingerprint, pending mutation và thứ tự 20 task trước khi xếp lần đầu lưu tại `.local/jira-auth/state.json`, atomic, mode 0600, bị Git ignore. Không có token/email/header/description hoặc response API thô trong journal. Không xóa journal để làm lại. Một lock chống hai process cùng dùng journal; không chạy song song bằng state khác.

## 5. Check và tiếp tục sau lỗi

```sh
python3 scripts/jira/auth_import.py --mode check
```

Check chỉ đọc: đủ ticket, nội dung, type/parent/assignee/labels, tập link và Rank. Chưa đạt thì exit khác 0; không sửa để che lỗi. Bằng chứng fake/offline không thay kết quả live này, cũng không chứng minh auth backend đã chạy.

- Mất phản hồi **create**: journal ghi ý định trước POST; khi chạy lại tìm mã/label và đọc issue. Nếu tìm duy nhất và đúng nội dung thì nhận lại, không tạo thêm. Nếu chưa thấy thì dừng vì search có thể trễ; chờ rồi chạy plan lại. Chỉ khi tự xác minh Jira không có ticket mới dùng `--retry-create AUTH-Txx` với mode apply; phải gõ xác nhận riêng. Không dùng cờ này chỉ vì mạng vừa lỗi.
- Mất phản hồi **link**: đọc lại hai chiều; nếu đã có thì không thêm. Nếu chưa thấy và chạy apply, script hỏi `RETRY LINK` trước khi thử có chủ ý. Không có comment nên không phát sinh comment trùng.
- Mất phản hồi **Rank**, kể cả HTTP 207: dừng, lần sau đọc thứ tự hiện tại; apply mới tính vị trí còn lệch và hỏi YES trước ghi. Chỉ retry đọc khi chờ index, không tự lặp PUT trong cùng lượt lỗi.
- Có duplicate, sai parent/nội dung/assignee/label, link đảo chiều/dư/ngoài manifest hoặc nguồn thay đổi: dừng để review; không ghi đè hay xóa dữ liệu. Chế độ này ưu tiên bàn giao ban đầu, không là công cụ tự đồng bộ ticket sau khi bắt đầu làm.
- Lock cũ do process bị kill: xác minh không còn process chạy rồi mới gỡ đúng `.local/jira-auth/state.lock`; giữ state.json. Không có lệnh tự reset/xóa issue/restore Rank. Snapshot chỉ lưu thứ tự tương đối 20 Task, không đủ để khôi phục toàn bộ board.
- Journal mất nhưng ticket còn: phát hiện qua mã/label, chỉ nhận lại khi toàn bộ nội dung khớp. Search trễ/issue bị ẩn do quyền vẫn là giới hạn: không thể cam kết exactly-once khi chạy song song hoặc tự ép retry.

Chạy `--state` khác chỉ khi có lý do và review mapping; giữ cùng state và cùng source-base cho plan/apply/check thông thường.

## 6. Phạm vi kiểm chứng và tài liệu API

Hiện có unit/integration orchestration bằng Jira giả: tạo/đọc lại, Rank, không ghi lần hai, mất phản hồi create/link, pending không rõ, duplicate/drift, link đảo chiều, credential redaction, redirect/endpoint allowlist và lock.

Ngày 07/11/2026, người dùng đã chạy plan/apply/check trên site thật và gửi log để đối chiếu: board 2/project DATN, Sprint DISABLED, Epic DATN-1 cùng 20 Task DATN-2–DATN-21, đủ 47 Blocks; check đạt nội dung/parent/assignee/labels/link/Rank. Đây là bằng chứng từ lượt chạy của người dùng, không phải agent tự gọi API live. Chưa xác minh trực quan toàn bộ ADF trên Jira hoặc liên kết Development với GitHub; không chứng minh auth runtime. Nguồn ticket/proposal/contract/FR được giữ nguyên trong lượt bàn giao Git để không đổi fingerprint của journal đã apply.

Tham khảo script `docs/planning/jira/rank-sprint.py` của CoLearnia (chỉ đọc, không copy sprint/80 task sang DATN) và tài liệu chính thức:

- [Create issue và create metadata](https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issues/)
- [Issue linking model](https://developer.atlassian.com/cloud/jira/platform/issue-linking-model/)
- [Create issue link](https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issue-links/)
- [Rank issues — gồm phản hồi 207 một phần](https://developer.atlassian.com/cloud/jira/software/rest/api-group-issue/)
- [Search and reconcile — độ trễ search index](https://developer.atlassian.com/cloud/jira/platform/search-and-reconcile/)

**Tiếp nối DATN-2 / AUTH-T01 (07/11):** proposal/OpenAPI được tiếp tục chỉnh theo task, vì vậy fingerprint nguồn hiện tại khác bản import đã apply. Đây không phải lỗi cần xóa journal hoặc import lại. Giữ journal/mapping hiện có; bộ import phục vụ bàn giao ban đầu, không đồng bộ tiến độ sau đó. Khi được giao, ghi link diff/PR và bằng chứng review vào ticket DATN-2; không đổi description hàng loạt hoặc tạo lại issue để vượt kiểm tra fingerprint.
