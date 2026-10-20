# Frontend — cách làm việc và thiết kế đã duyệt

## Hiện trạng

Người dùng đã duyệt phong cách pilot ngày **20/10/2026**: shell, danh sách/chi tiết dự án và workspace phân công. Đây là nền để mở rộng từng đợt, **không phải toàn bộ frontend đã hoàn thành**. Dữ liệu tại `src/demo/` chỉ trong bộ nhớ, reload khôi phục fixture; điểm minh họa không phải engine thật. Chưa API/auth/persistence/Go embed/TanStack Query. [Bằng chứng pilot](DESIGN-PILOT.md).

## Chạy lại không cần DB

Môi trường đã chạy: Node **24.15.0**, pnpm **11.19.0**; package.json khai báo Node 24.x/pnpm 11.x và pin packageManager, lockfile là nguồn phiên bản dependency. Node 24.x dùng native TypeScript stripping cho test; không đổi runtime/dependency chỉ để làm bố cục.

```bash
# Từ thư mục web/
pnpm install --frozen-lockfile
pnpm dev                 # localhost:5173; giữ strictPort, không tự kill tiến trình khác
pnpm typecheck
pnpm test                # hiện chỉ test model demo
pnpm build
pnpm check:harness       # 6 test harness → kiểm kê nguồn
pnpm check               # harness → typecheck → test model → build
node ../scripts/check-harness.mjs
node ../scripts/check-harness.mjs --staged # chỉ lúc chuẩn bị bàn giao Git
```

Có thể chạy riêng test: `node --experimental-strip-types --test src/demo/model.test.ts`; test harness từ root: `node --test scripts/check-harness.test.mjs`. Test Git dùng repo tạm riêng, tự dọn sau chạy, không đụng index/config/commit repo thật. Không truyền một file TS riêng cho tsc thay tsconfig project. Chưa có lint/formatter/E2E tự động; không ghi đã chạy các cổng chưa tồn tại. Các check không gọi DB/API; build ghi `dist/` bị ignore, không tự sửa mã nguồn. Lệnh cài đặt cần registry/cache và quyền phù hợp; lỗi môi trường không phải test pass.

`--staged` phát hiện nguồn chưa được theo dõi và nội dung làm việc mới hơn staging. Chế độ thường không ép stage mỗi lần sửa; cả hai chế độ đều không tự git add/commit. Staging chỉ chuẩn bị snapshot, chưa có nghĩa clone nhận được thay đổi. Tạo commit cần người dùng yêu cầu riêng; không đưa `node_modules`, `dist`, bí mật hoặc tài liệu ngoài phạm vi vào snapshot.

## Chuẩn thiết kế đã duyệt

- Notion-inspired ở cách tổ chức workspace/bảng/side peek; không sao chép branding, thêm tính năng hay biến thành dashboard nhiều KPI.
- Token màu/typography nguồn duy nhất là [src/index.css](src/index.css): canvas ấm, surface trắng, chữ/nút chính đậm, border nhẹ; không nhân bản hex sang từng màn hình. Inter có fallback hệ thống, chưa tải font ngoài.
- Shell ổn định, breadcrumb/title/tabs rõ; bảng đủ mật độ để làm việc, tìm/lọc gần dữ liệu. Chi tiết task mở bên cạnh, không che toàn bộ workspace.
- Phân công: việc đã chọn → ứng viên so sánh → bản nháp/tải dự kiến. Không tự chọn Top-1; thao tác nháp khác xác nhận. Thủ công không giả điểm/hạng.
- Tái dùng Button/Dialog trong `src/components/ui/`, pattern tại `shared.tsx`; đây là primitives tùy biến theo shadcn/Radix, không phải mọi màn hình đã dùng bộ shadcn đầy đủ.
- Hành động nhỏ dùng icon có tên truy cập/tooltip; thao tác phá hủy có xác nhận phù hợp. UI tiếng Việt tự nhiên, focus bàn phím rõ, không chỉ phân biệt bằng màu.
- Desktop là ưu tiên; bảng cuộn trong vùng riêng, không tràn ngang cả trang. Kiểm tra màn hình bị đổi ở desktop và một viewport nhỏ phù hợp; không coi một lượt xem là nghiệm thu mọi thiết bị.
- Sidebar mở rộng theo quyền và màn hình thực dựng. Cấu trúc/phân công/nghiệm thu nằm trong dự án; không tự thêm báo cáo/cài đặt/KPI/roster từ ảnh tham khảo.

## Vòng làm việc — harness tối thiểu

1. Đọc [mục lục nguồn](../docs/README.md), chọn workflow/contract đúng miền; không đọc lại mọi baseline/research mỗi lượt.
2. Ghi một mục đợt ngắn theo mẫu dưới đây; phạm vi người dùng chưa giao hoặc cấu trúc/nghiệp vụ chưa chốt phải hỏi trước.
3. Implement trong phạm vi, giữ demo adapter tách khỏi UI. Quyền demo chỉ phục vụ trải nghiệm, không chứng minh quyền server.
4. Chạy check theo bảng; review diff và kiểm tra không lan sang backend/DB/export. Không sửa test/tiêu chí chỉ để có màu xanh.
5. Cập nhật cùng mục đợt: đã code / đã kiểm tra / chưa kiểm tra / bị chặn và bước sau. Khi lỗi môi trường lặp không tiến triển, báo bằng chứng thay vì vòng thử vô hạn.

| Loại thay đổi | Cổng cần thực hiện |
|---|---|
| Chỉ tài liệu | `node scripts/check-harness.mjs` từ root, `git diff --check`; trước bàn giao thêm `--staged` và `git diff --cached --check`; không chạy build lại mỗi sửa câu chữ |
| Copy/icon/layout | Typecheck + build; xem vùng bị đổi và luồng chính nếu ảnh hưởng tương tác |
| Form/validation | Các cổng trên + thử nhập đúng/lỗi quan trọng/hủy; test logic nếu có hàm độc lập |
| Quyền/trạng thái/draft | `pnpm check` + test bất biến bị đổi + một hành trình UI chính/ca bị từ chối có liên quan |
| API/auth/upload/backend | Đóng contract còn mở, đọc security/API; lập phạm vi riêng và check tích hợp/bảo mật, không dùng demo làm chứng cứ |

### Mẫu đợt — dùng ngay tại sổ bên dưới

```text
Đợt / trạng thái:
Mục tiêu và phần không làm:
Nguồn: file + mục/ID contract liên quan; quyết định còn mở:
Phạm vi file dự kiến:
Tiêu chí đạt: người dùng nhìn thấy/làm được gì:
Kết quả: đã code; check thực chạy/kết quả; chưa kiểm tra/bị chặn:
Bàn giao: rủi ro còn lại; bước kế tiếp được giao/chờ duyệt:
```

Không tạo file kế hoạch mới khi một mục ngắn đủ dùng. Đợt lớn cần kế hoạch riêng thì dẫn tới một file duy nhất và đóng/lưu lịch sử sau khi xong.

## Sổ công việc

### Sửa điều hướng bản nháp và tạo nhóm trước task — 21/10/2026 — hoàn tất

- Mục tiêu/nguồn: sửa hai điểm review được người dùng duyệt; giữ vòng nháp → xác nhận của UC003/UI-12 và nhóm bắt buộc cùng dự án ở UC002/UI-11.
- Phạm vi: ProjectWorkspace, AssignmentWorkspace và sổ này. Không đổi nghiệp vụ/schema/dependency; chưa xử lý chặn điều hướng toàn app, browser back/reload hoặc lưu nháp bền vững.
- Tiêu chí: hai tab và nút quay lại dùng chung cảnh báo bỏ nháp; tiếp tục chọn giữ nguyên dữ liệu, xác nhận bỏ mới chuyển tab; không có lựa chọn thì đi thẳng. Tạo task khi chưa có nhóm dẫn sang Cấu trúc kèm hướng dẫn, có nhóm vẫn mở form bình thường.
- Đã sửa: draft được giữ tại ProjectWorkspace và truyền có kiểm soát vào AssignmentWorkspace; hai tab/nút quay lại đi qua cùng hộp bỏ nháp. Chưa có nhóm thì nút Tạo công việc mở Cấu trúc với hướng dẫn tạo nhóm, không mở form không thể lưu.
- Đã kiểm tra: `pnpm check` qua 6 test harness, kiểm kê 63 file, typecheck, 16 test model và build; `git diff --check` qua. Browser kiểm tra cảnh báo ở cả hai tab và nút quay lại; tiếp tục giữ người đã chọn, bỏ nháp mới chuyển đúng tab; nháp rỗng không hỏi. Tạo dự án mẫu chưa có nhóm → hướng dẫn → tạo nhóm → form task có nhóm để chọn. Không xác nhận phân công thử; đã khôi phục fixture sau kiểm tra.
- Giới hạn/bàn giao: tương tác trên được kiểm tra browser thủ công, chưa có test React/E2E tự động hoặc cổng mới; không thêm dependency. Chưa chặn sidebar/browser back/reload, không lưu nháp bền vững. Nhánh `bug/project-workspace-navigation`; giữ thay đổi các đợt trước/ngoài phạm vi, không stage/commit/push.

### Khởi tạo dự án và thiết lập trách nhiệm — 21/10/2026 — hoàn tất phạm vi, chờ duyệt UI

- Mục tiêu: Giám đốc tạo dự án kèm PM; PM chỉ định PTCM và bắt đầu khi có task ASSIGNED; state dự án dùng chung. Giữ visual đã duyệt.
- Nguồn: project_task_assignment_workflow UC001; screen-contracts UI-10 và trạng thái PLANNING; dictionary 5A mục 2. Không tự đổi nghiệp vụ/schema.
- Ngoài phạm vi: đổi PM, hoàn thành/đóng/hủy dự án, thực hiện task/nhật ký/nghiệm thu, API/auth/DB/persistence, dependency mới.
- File dự kiến: App, demo/model và test, ProjectList/ProjectWorkspace, form tạo dự án và khu vực trách nhiệm.
- Tiêu chí: tạo có validation/hủy, không tự gán creator; PM-only/PTCM-only/kiêm nhiệm tách quyền; chặn PTCM còn việc mở hoặc xung đột đầu ra cũ, không thu hồi ngầm; bắt đầu cập nhật danh sách/workspace cùng state.
- Đã code: form tạo dự án có PM bắt buộc, lỗi theo trường và hủy không lưu; dự án mới PLANNING chưa có PTCM/nhóm/task. Dự án dùng chung state ở shell/danh sách/workspace; danh tính mô phỏng khai báo quyền hệ thống riêng với trách nhiệm dự án, đổi danh tính về danh sách đúng phạm vi. PM chỉ định PTCM/kiêm nhiệm, có kiểm tra việc mở và đầu ra cũ; bắt đầu bằng xác nhận khi có ASSIGNED. Lịch sử trước/sau/actor trong bộ nhớ chỉ minh họa, không audit server.
- Đã kiểm tra: `pnpm check` qua 6 test harness, kiểm kê 63 file, typecheck, 16 test model và build; `git diff --check` qua. Browser: tạo dự án/lỗi bắt buộc, Giám đốc không tự có quyền chuyên môn; chặn PTCM còn việc mở hoặc đầu ra cũ; PM kiêm PTCM hợp lệ ở dự án khác; bắt đầu cập nhật danh sách; hủy form không tạo. Xem workspace desktop 1440px và form 390px, không tràn ngang trang/modal.
- Giới hạn: bộ nhớ mất khi reload; không API/auth/DB/persistence/audit thật hoặc E2E tự động. Projection xung đột đầu ra là fixture từ đóng góp cũ (Thanh/T007), không thêm bảng/cột DB hoặc tự xây luồng đóng góp/nghiệm thu. Trạng thái terminal chỉ được chặn theo model, chưa dựng UI hoàn thành/đóng/hủy. Giữ nguyên task và lịch sử khi thay PTCM, không cấp quyền cá nhân thực hiện task trong đợt này.
- Bàn giao: nhánh `feature/project-setup-ui`, giữ thay đổi có sẵn của đợt cấu trúc và ngoài `web/`; không stage/commit/push. Không dependency/lockfile mới. Chờ duyệt UI đợt này trước khi mở phạm vi tiếp theo.

### Cấu trúc nhóm và công việc — 21/10/2026 — hoàn tất phạm vi, UI được người dùng duyệt

- Mục tiêu: PTCM tạo/sửa nhóm và đặc tả task trước lần giao đầu; dữ liệu dùng chung với danh sách/phân công. Giữ visual pilot; không API/DB, xóa task/nhóm, gia hạn/làm rõ đặc tả sau giao hoặc vòng đời dự án.
- Nguồn: screen-contracts UI-11, bảng trạng thái dự án; project_task_assignment_workflow UC002; dictionary 5A mục 3–5. Miền weight chưa chốt: giá trị nhập chỉ minh họa, không ràng buộc 1–5 hoặc đưa vào công thức.
- File dự kiến: demo/model và test; App; ProjectWorkspace; component form/cấu trúc dự án; Dialog dùng chung nếu cần mở rộng kích thước.
- Tiêu chí: nhóm trống vẫn hiển thị, tên nhóm/mã task/kỹ năng không trùng trong phạm vi; form có hủy/lỗi; ngày đủ cặp trong dự án; task tạo mới chờ phân công; PM-only chỉ xem; task từng giao không mở lại form sửa spec.
- Đã code: tab Cấu trúc với nhóm độc lập (kể cả nhóm trống); tạo/sửa tên/mô tả nhóm; form task có yêu cầu kỹ năng động, bốn mức, trọng số minh họa; task mới vào cùng state với danh sách/phân công. Lưu mốc giao đầu trong mock để thu hồi không mở lại spec. Fixture gợi ý không phụ thuộc tên nhóm, vẫn chỉ là mẫu chứ không engine.
- Đã kiểm tra: `pnpm check` qua 6 test harness, kiểm kê 61 file, typecheck, 11 test model và build; `git diff --check` qua. Browser: tạo nhóm/trùng tên, tạo task/lỗi bắt buộc, sửa và hủy sửa, chọn task mới sang manual, PM-only không có nút chuyên môn. Đã xem desktop 1440px, form/trang cấu trúc 390px; sửa tràn ngang từ nhãn ẩn của cột thao tác, bảng/tab cuộn trong vùng riêng.
- Giới hạn: chưa API/auth/persistence/audit server; chưa E2E tự động hoặc kiểm thử mọi thiết bị. Kỹ năng trong danh mục demo chỉ có bốn mục; gợi ý có thể rỗng với task mới, không giả engine. Không dựng xóa nhóm/task, gia hạn có lý do/lịch sử hoặc sửa làm rõ sau giao. Cặp ngày đã lưu giữ nguyên trong form này; task chưa có ngày được khai báo lần đầu, không ghi đè kế hoạch cũ bằng thao tác sửa spec.
- Bàn giao: giữ palette/font/components đã duyệt, không thêm dependency dự án hoặc đổi lockfile. Prettier chỉ chạy tạm để định dạng file của đợt, chưa tạo cổng formatter mới. Chưa stage/commit/push; các thay đổi có sẵn ngoài `web/` giữ nguyên. Bước kế tiếp là người dùng duyệt tab/form trước khi giao đợt mới.

### Pilot thiết kế — hoàn tất phạm vi, visual được duyệt

- Phạm vi/bằng chứng chi tiết tại DESIGN-PILOT; không gọi chức năng ngoài pilot là hoàn thiện.
- Đã xác minh trong đợt pilot: typecheck, 5 test model, build, xem desktop/mobile chính. Không E2E/API/auth/server transaction.

### Harness tối thiểu — 20/10/2026

- Mục tiêu: thống nhất điểm vào, nguồn chuẩn, thiết kế đã duyệt, lệnh kiểm tra và bàn giao; không mở rộng UI/nghiệp vụ/DB.
- Hợp nhất instruction Codex/Claude; bảo toàn bản dài cũ trong archive. Phân loại tài liệu thay vì xóa lịch sử hay nhân bản nghiệp vụ.
- Đã kiểm tra: `node scripts/check-harness.mjs` qua 40 file trực tiếp; 5 phép thử parser/link/path qua; `git diff --check` qua. `pnpm check` qua typecheck, 5 test model và build. Metadata runtime pin Node 24.x/pnpm 11.19.0, không đổi dependency hoặc lockfile.
- Không xem lại browser vì không sửa UI/logic; không chạy Go/DB/Make. Bộ check tài liệu không kiểm chứng ngữ nghĩa workflow/anchor, không tự stage/commit; nhiều file nguồn vẫn đang untracked và cần được commit cùng mã khi người dùng yêu cầu.
- Bước tiếp theo chưa tự triển khai: chọn đợt UI kế tiếp có phạm vi/contract, đề xuất cấu trúc nhóm/task/yêu cầu kỹ năng trước khi mở rộng phân công. Miền weight/công thức vẫn chờ chốt nếu chạm logic liên quan.

### Khép review harness — 20/10/2026

- Phạm vi: lưu 5 phép thử thành test lặp lại, thêm ca Git staging và cổng bàn giao; không đổi UI/nghiệp vụ/DB/dependency. Nguồn: nghiên cứu harness mục 6.3–6.6, 8–9 và review hai điểm còn thiếu.
- File: scripts/check-harness*, web/package.json, chỉ dẫn/mục lục/README; chuẩn bị staging đúng nguồn hiện hành và pilot, không gom toàn bộ worktree.
- Tiêu chí: test qua; cổng thường qua; `--staged` phải phát hiện thiếu/stale trước khi chuẩn bị Git và qua sau đó; nội dung hiện có của người dùng được giữ nguyên; không tự commit.
- Đã xác minh: 6 test harness qua (5 phép thử được lưu lại + ca Git có cả từ chối/cho qua); cổng thường qua 59 file. `--staged` từ chối lúc thiếu nguồn, qua sau khi staging đúng danh sách; test repo tạm cũng phát hiện sửa tiếp chưa stage. `pnpm check` qua harness, typecheck, 5 test model và build.
- Review cuối: staging khớp đúng 59 file cần bàn giao; `git diff --cached --check` qua; 127 link file trong 16 nguồn nghiệp vụ/UI tồn tại. Không sửa UI/logic, không backend/DB/dependency hoặc export; không kiểm toán lại toàn bộ nghiệp vụ, anchor hay nội dung bí mật bằng công cụ này.
- Bàn giao: người dùng đã cho phép lưu 59 file nguồn hiện hành/pilot/harness thành ba commit local, thông điệp hoàn toàn bằng tiếng Anh: tài liệu nghiệp vụ → pilot frontend → harness. Các file dùng chung được tách theo phần thay đổi để pilot không gọi script harness chưa tồn tại; thay đổi ngoài danh sách giữ nguyên ngoài commit. Chưa push, chưa xác minh clone/cài mới trên máy sạch. Hiệu quả thời gian/chi phí cần quan sát qua vài đợt, không cam kết số liệu chưa đo. Không có lỗi chặn phát triển UI trên máy hiện tại.
