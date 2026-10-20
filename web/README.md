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
