# Đối soát bộ ticket MSK trên Jira DATN

[Importer MSK](msk_import.py) đọc [tickets](../../docs/proposals/backend-manual-skills-tickets.md), dùng engine [jira_batch.py](jira_batch.py), transport và ADF hiện có. Python 3.9+, thư viện chuẩn; token nhập ẩn trong terminal, không đọc `.env`.

Đích: `https://datn-mingquanjp.atlassian.net`, project **DATN**, board **1**. Bộ gồm **1 Epic, 12 Task, 19 Blocks**, Rank T01→T12. Quân xác nhận đã nhập ngày 27/11/2026; journal local map Epic DATN-49 và MSK-T01–T12 tới DATN-50–61. Mapping này không thay kết quả `check` online.

## 1. Snapshot nhập Jira

Nguồn dùng trong lượt nhập nằm tại commit `496054cef48c6f446a86b1789680c6a0a73e36ca`. Proposal hiện hành đã rút gọn phần bàn giao; file tickets chỉ dọn dòng trắng cuối. Nội dung description giữ nguyên, nhưng hash nguồn khác snapshot nên guard sẽ từ chối journal cũ nếu chạy online từ bản mới.

Để đối soát bộ đã nhập, tạo checkout riêng từ snapshot. Chọn đường dẫn mới nếu thư mục dưới đây đã tồn tại:

```sh
git fetch origin 496054cef48c6f446a86b1789680c6a0a73e36ca
git worktree add --detach /private/tmp/ai-staffing-msk-import-check 496054cef48c6f446a86b1789680c6a0a73e36ca
cd /private/tmp/ai-staffing-msk-import-check
python3 scripts/jira/msk_import.py --mode check
```

`check` chỉ đọc, tìm issue theo source labels và đối soát 13 issue, nội dung/type/parent/assignee/labels/19 Blocks/Rank. Không cần sao chép journal sang checkout kiểm mới. Lượt nhập ban đầu dùng đường dẫn nguồn chờ remote, vì vậy giữ nguyên cách gọi trên, không thêm `--source-base`.

Giữ journal gốc `.local/jira-msk/state.json`; không xóa/sửa fingerprint để vượt guard. Dọn checkout kiểm sau khi dùng bằng `git worktree remove /private/tmp/ai-staffing-msk-import-check` từ checkout còn lại; chỉ thực hiện khi thư mục sạch.

## 2. Kiểm offline và metadata

Chạy tại gốc checkout chứa bộ MSK:

```sh
python3 scripts/jira/msk_import.py --mode plan --offline
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s scripts/jira -p 'test_*.py' -v
```

Offline kiểm Rank/WBS/phụ thuộc và chuyển description sang ADF; không hỏi token, gọi Jira hoặc tạo journal. `--manifest-out /tmp/datn-msk-manifest.json` xuất bản đọc thêm vào file mới mode 0600, không ghi đè hoặc dùng làm nguồn chỉnh tay.

`python3 scripts/jira/msk_import.py --mode diagnose` chỉ đọc metadata site/DATN-12/board, không dùng nguồn/journal. Nếu token có scope cần gateway, thêm cùng `--cloud-id UUID_SITE_DATN` vào lệnh online theo [hướng dẫn chung](README.md).

## 3. Các mode và phạm vi ghi

- `plan`: đọc metadata/quyền và đối soát bộ ticket trước khi tạo; không sửa Jira.
- `apply`: chỉ dùng cho lượt tạo còn thiếu sau khi kiểm nguồn, journal và plan. Hỏi `YES`; tạo issue, Blocks và Rank. Bộ MSK hiện đã nhập, không dùng apply để cập nhật description hoặc tiến độ.
- `check`: chỉ đọc và đối soát toàn bộ; chạy từ snapshot nhập như mục 1.
- `diagnose`: chỉ đọc cấu hình site/board.

Apply không ghi đè/xóa ticket đã có hoặc gửi comment/status/priority/sprint/deadline/estimate/remaining/actual. Các đợt auth/OAP/PCA giữ field và thứ tự tương đối; Rank MSK có thể đổi vị trí tương đối với chúng.

Khi một lượt nhập mới thực sự dùng link remote, `--source-base` phải là URL GitHub blob với SHA 40 ký tự, đúng repo `mingquanjp/prj-manage-system` và khớp tám nguồn local. Giữ cùng source-base/nội dung trong cả lượt; script không tự push hoặc xác minh commit đã có trên remote.

## 4. Journal và khôi phục lỗi

Journal cố định `.local/jira-msk/state.json`, mode 0600, Git-ignore, lock độc quyền và từ chối symlink. Fingerprint gồm tám nguồn và cấu hình/source-base. Không dùng journal auth/OAP/PCA cho MSK.

- Mất phản hồi create: đối soát Jira theo mã/label/nội dung, chờ search index nếu cần. Chỉ khi xác minh chưa tồn tại mới dùng `--mode apply --retry-create MSK-Txx` (Epic: `MSK-EPIC`), xác nhận `RETRY MSK-Txx` rồi `YES`.
- Mất phản hồi link: đối soát hai đầu; thiếu link cần xác minh rồi xác nhận `RETRY LINK` ở lượt apply mới. Rank mất phản hồi: dừng và plan lại trước khi apply có chủ ý.
- Duplicate/drift/sai parent/nội dung/link đảo hoặc ngoài manifest: dừng để review; script không tự ghi đè/xóa.
- Chỉ dọn lock cũ sau khi xác nhận process đã kết thúc; giữ state. Không đảm bảo exactly-once nếu ép retry hoặc chạy nhiều process.

Chi tiết transport/redaction và khôi phục dùng chung tại [README Jira](README.md), [README PCA](README-pca.md).

## 5. Kiểm chứng

Bộ tests offline/Jira giả kiểm nguồn → ADF, parser/Rank/WBS/phụ thuộc, plan/check chỉ đọc, apply/rerun, journal riêng, lost-create và giữ các đợt cũ. Regression engine kiểm thêm lỗi link/Rank, transport và token TTY. Kết quả chuẩn bị và giới hạn tại [proposal mục 10](../../docs/proposals/backend-manual-skills.md#10-bàn-giao-triển-khai); chỉ kết quả plan/apply/check từ site thật xác nhận nội dung Jira hiện tại.
