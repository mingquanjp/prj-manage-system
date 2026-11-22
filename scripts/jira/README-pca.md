# Đưa ticket đề nghị hồ sơ lên Jira DATN

Script [pca_import.py](pca_import.py) dùng lại engine [jira_batch.py](jira_batch.py). Nguồn là [tickets PCA](../../docs/proposals/backend-profile-change-approval-tickets.md), [proposal](../../docs/proposals/backend-profile-change-approval.md), [OpenAPI](../../docs/api/profile-change-approval.openapi.json), [SQL plan](../../docs/proposals/backend-profile-change-approval-sql-plan.md) và workflow được dẫn. Không chép description vào Python hoặc chỉnh manifest để thay nghiệp vụ.

Đích cố định: `https://datn-mingquanjp.atlassian.net`, project **DATN**, board **1**. Có **1 Epic + 12 Task, 12 WBS và 17 Blocks**. T01 bàn giao thiết kế đã có; T11/T12 là phần tiếp nối F4, không chặn demo T10. Dùng Python 3.9+ với thư viện chuẩn.

## 1. Kiểm offline

Từ thư mục gốc repo:

```sh
python3 scripts/jira/pca_import.py --mode plan --offline
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s scripts/jira -p 'test_*.py' -v
```

Không cần token/mạng/journal. Kiểm Rank/WBS/description, phụ thuộc hai chiều, bảng tình huống hai cột và checklist ADF. Đây là kiểm importer, không kiểm chứng API/DB/browser ứng dụng.

Muốn xem manifest sinh từ nguồn, thêm `--manifest-out /tmp/datn-pca-manifest.json`. Phải chọn file mới; tạo mode0600, không ghi đè và không xuất vào `.local` của repo. Không commit manifest làm nguồn thứ hai.

## 2. Diagnose và plan — chỉ đọc Jira

```sh
python3 scripts/jira/pca_import.py --mode diagnose
python3 scripts/jira/pca_import.py --mode plan
```

Nhập email Atlassian và API token qua prompt ẩn trong terminal tương tác. Không đặt token trong lệnh/chat/Git; script không đọc `.env`. Diagnose xác minh site/DATN-12/board; plan kiểm metadata, quyền, parent, assignee, Blocks và Rank theo engine hiện hành. Không tự thay quyền/configuration. Với token có scope, thêm `--cloud-id UUID_SITE_DATN` vào mọi lệnh online, theo [hướng dẫn công cụ Jira](README.md).

Plan đối soát issue hiện có/còn thiếu bằng label `datn-pca-v1` và mã nguồn PCA; không sửa Jira hoặc lưu mapping mới. Online plan/check dùng lock tạm riêng và dọn khi kết thúc.

## 3. Apply và check

Sau khi đã đọc plan, người dùng chủ động chạy:

```sh
python3 scripts/jira/pca_import.py --mode apply
python3 scripts/jira/pca_import.py --mode check
```

Apply hỏi **YES**, tạo phần còn thiếu, parent Epic, assignee tài khoản đã xác minh, labels PCA, 17 Blocks và Rank 12 Task. Không tự dự đoán mã DATN. Không gửi status/priority/sprint/deadline/estimate/remaining/actual, không tự Done hoặc gửi comment. Chỉ ghi create issue, issueLink và Rank; không update/xóa nội dung issue khác. Ticket auth/OAP giữ field và thứ tự tương đối; vị trí PCA so với chúng có thể thay đổi.

Check chỉ đọc, đối soát 13 issue, description/cấu trúc bảng/link nguồn/type/parent/assignee/label, 17 Blocks và Rank. Không dùng importer để đồng bộ tiến độ hoặc ghi đè nội dung đã sửa tay.

Nguồn chưa commit/push: description ghi đường dẫn repo và “link remote chờ bàn giao”. Nếu có commit nguồn đã push, thêm cùng `--source-base https://github.com/mingquanjp/prj-manage-system/blob/FULL_COMMIT_SHA` vào plan/apply/check. Cũng chấp nhận tên repo cũ `ai-staffing`. SHA phải đủ40 ký tự và cả bảy nguồn local phải khớp commit local. Script không tự xác minh remote chứa commit, không push; người dùng kiểm trước. Giữ nguyên nguồn/source-base suốt lượt import.

## 4. Journal và kết quả chưa rõ

Journal riêng **`.local/jira-pca/state.json`**, mode0600, Git-ignore; lock độc quyền, từ chối symlink. Không dùng/xóa journal auth/OAP hoặc sửa fingerprint để chạy tiếp. Fingerprint gồm snapshot bảy nguồn, cấu hình và source-base; nguồn thay đổi sau apply thì dừng review.

Trước mỗi mutation ghi pending atomically. Mất phản hồi create: đọc Jira và đối soát mã/label/nội dung trước, không tự POST lại. Nếu search chưa thấy, chờ index và xác minh thủ công. Chỉ khi chắc chưa tồn tại mới dùng `--mode apply --retry-create PCA-Txx` (Epic: `PCA-EPIC`) và xác nhận **RETRY PCA-Txx**, rồi **YES**. Không hứa exactly-once.

Mất phản hồi link: đọc hai đầu; nếu thiếu phải xác minh và xác nhận **RETRY LINK** ở lượt apply mới. Rank mất phản hồi: dừng lượt hiện tại, plan lại rồi apply có chủ ý. Duplicate/drift/link đảo/ngoài manifest/account hoặc metadata khác thì dừng, không ghi đè/xóa. Không xóa journal để nhập lại; chỉ dọn lock cũ sau khi chắc process đã kết thúc.

Transport dùng chung chặn redirect; lỗi không in provider body/header/token. Xem [cơ chế OAP](README-oap.md) để biết chi tiết engine dùng chung. Không chạy apply OAP/auth để nhập PCA.

## 5. Bằng chứng và giới hạn

Tests dùng Jira giả để kiểm nguồn→ADF, read-only plan/check, apply/read-back/rerun, giữ nguyên đợt auth/OAP, pending create/link/Rank, drift/duplicate/metadata, journal/export và token TTY. Offline không chứng minh quyền hay dữ liệu trên Jira thật. Chưa gọi API Jira thật trong lượt chuẩn bị này; người dùng chạy diagnose/plan/apply/check để đối soát site thực.
