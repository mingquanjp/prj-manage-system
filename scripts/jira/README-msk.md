# Đưa ticket kỹ năng thủ công lên Jira DATN

[msk_import.py](msk_import.py) đọc trực tiếp [tickets MSK](../../docs/proposals/backend-manual-skills-tickets.md), dùng engine [jira_batch.py](jira_batch.py) và transport/ADF hiện có. Không chép description vào Python, không thêm dependency. Nguồn thiết kế là [proposal](../../docs/proposals/backend-manual-skills.md), [SQL plan](../../docs/proposals/backend-manual-skills-sql-plan.md) và [OpenAPI MSK](../../docs/api/manual-skills.openapi.json); enum thông báo mở rộng nằm trong canonical PCA.

Đích cố định: `https://datn-mingquanjp.atlassian.net`, project **DATN**, board **1**. Bộ này gồm **1 Epic + 12 Task, 12 WBS, 19 Blocks**; Rank T01→T12, khi đủ đầu vào phụ thuộc. Không đoán Jira key. Python **3.9+**, chỉ thư viện chuẩn.

## 1. Kiểm offline

Nguồn và script hiện ở worktree riêng. Dùng đúng thư mục này, không chạy importer PCA/OAP:

```sh
cd /private/tmp/ai-staffing-skill-planning
python3 scripts/jira/msk_import.py --mode plan --offline
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s scripts/jira -p 'test_*.py' -v
```

Sau khi bàn giao/merge, có thể chạy tại checkout chứa cùng bộ nguồn. Giữ cùng checkout/journal trong một lượt import.

Offline không hỏi token, gọi Jira hoặc tạo journal. Nó kiểm Rank/WBS/phụ thuộc, chuyển description sang ADF và giữ bảng hai cột/checklist/lệnh. Muốn đọc toàn manifest, thêm `--manifest-out /tmp/datn-msk-manifest.json`: chọn file mới, mode0600, không ghi đè hoặc dùng manifest làm nguồn thứ hai.

## 2. Kiểm Jira thật — chỉ đọc

Tại cùng thư mục:

```sh
python3 scripts/jira/msk_import.py --mode diagnose
python3 scripts/jira/msk_import.py --mode plan
```

Nhập email Atlassian và API token qua prompt ẩn trong terminal tương tác; không gửi vào chat, đặt trong command, file hoặc Git. Script không đọc `.env`. Diagnose xác minh site/DATN-12/board; plan đối soát quyền, metadata, parent, assignee, label, Blocks và Rank, không sửa Jira. Nếu cần gateway cho token có scope, dùng cùng `--cloud-id UUID_SITE_DATN` trong mọi lệnh online theo [hướng dẫn chung](README.md).

Plan tìm đúng bộ qua `datn-msk-v1`, `source-msk-epic` và `source-msk-t01`…`t12`. T01 bàn giao thiết kế đã có; không tính lại thành remaining. Không đổi priority/status hoặc tự chuyển Done.

## 3. Tạo và đối soát

Sau khi đọc kết quả plan, người dùng chủ động chạy:

```sh
python3 scripts/jira/msk_import.py --mode apply
python3 scripts/jira/msk_import.py --mode check
```

Apply hỏi **YES**, tạo phần còn thiếu với parent Epic, assignee tài khoản được xác minh, ba labels mỗi issue; thêm19 Blocks và xếp Rank12 Task. Chỉ có các thao tác ghi create issue/issueLink/Rank. Không ghi đè hoặc xóa ticket đã có, gửi comment, status/priority/sprint/deadline/estimate/remaining/actual. Các đợt cũ giữ field và thứ tự tương đối; vị trí MSK so với chúng có thể thay đổi khi Rank.

Check chỉ đọc, kiểm đủ13 issue, nội dung/type/parent/assignee/labels/19 Blocks/Rank. Không chứng minh ứng dụng đã chạy; không dùng importer đồng bộ tiến độ hoặc sửa nội dung Jira sau khi bắt đầu làm.

Nguồn chưa commit/push được mô tả bằng đường dẫn repo kèm “link remote chờ bàn giao”. Khi có commit nguồn đã push và mở được, thêm cùng giá trị sau vào plan/apply/check:

```text
--source-base https://github.com/mingquanjp/prj-manage-system/blob/FULL_COMMIT_SHA
```

Thay bằng SHA40 ký tự thật. Script kiểm tám nguồn local khớp commit local, gồm canonical notification PCA; không tự push hoặc xác minh commit đã có trên remote. Giữ nguyên source-base và nội dung nguồn trong suốt lượt import.

## 4. Journal và lỗi chưa rõ kết quả

Journal cố định **`.local/jira-msk/state.json`**, mode0600, Git-ignore, lock độc quyền và từ chối symlink. Không dùng/xóa journal auth/OAP/PCA. Fingerprint gồm tám nguồn và cấu hình/source-base; nguồn đổi sau apply sẽ dừng để review, không sửa fingerprint để vượt guard.

- Mất phản hồi create: đọc lại Jira theo mã/label/nội dung; chưa thấy trên search thì chờ index và xác minh, không tự POST lại. Chỉ khi chắc issue chưa tồn tại mới dùng `--mode apply --retry-create MSK-Txx` (Epic: `MSK-EPIC`), xác nhận **RETRY MSK-Txx**, rồi **YES**.
- Mất phản hồi link: đối soát hai đầu; thiếu link cần xác minh rồi xác nhận **RETRY LINK** ở lượt apply mới. Rank mất phản hồi: dừng, plan lại trước apply có chủ ý.
- Duplicate/drift/sai parent/nội dung/link đảo hoặc ngoài manifest: dừng, không tự ghi đè/xóa. Không hứa exactly-once khi chạy nhiều process hoặc ép retry.
- Chỉ dọn đúng lock cũ sau khi chắc process đã kết thúc; giữ state. Không xóa journal để nhập lại hoặc chạy importer đợt khác.

Chi tiết transport/redaction/metadata và cơ chế khôi phục dùng chung tại [README chung](README.md) và [README PCA](README-pca.md).

## 5. Bằng chứng và giới hạn

Kiểm offline/Jira giả chứng minh chuyển nguồn→ADF, parser từ chối nguồn lệch, read-only plan/check, apply/rerun, journal riêng, lost-create và giữ nội dung/thứ tự OAP/PCA. Regression engine hiện hữu kiểm các lỗi link/Rank, transport và token TTY. Kết quả thực chạy được ghi tại proposal phần10; chưa gọi Jira thật trong lượt chuẩn bị script. Chỉ kết quả plan/apply/check của người dùng trên site thực mới xác nhận lượt nhập Jira.
