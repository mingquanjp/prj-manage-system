# Bảo mật

Bảo mật là ưu tiên xuyên suốt của dự án, không phải việc làm thêm ở cuối. Neo tài liệu: mục 2.4.1
(Bảo mật) và 2.4.2 (Toàn vẹn dữ liệu).

Nguyên tắc chung: **từ chối mặc định**. Thiếu khai báo quyền thì phải bị chặn, không phải được mở.

## 1. Bí mật và cấu hình

- `.env` nằm trong `.gitignore`; repo chỉ commit `.env.example` với giá trị rỗng. **Đã kiểm chứng**
  là không có chuỗi kết nối thật nào trong lịch sử git.
- Mọi bí mật đọc từ biến môi trường: `DATABASE_URL`, `JWT_SECRET`. Không hardcode, không giá trị
  mặc định cho `JWT_SECRET` — thiếu thì server phải **từ chối khởi động**, đừng tự sinh khoá tạm.
- `sslmode=require` bắt buộc khi kết nối Neon.
- Nếu lỡ commit bí mật: **xoay khoá trước**, sửa lịch sử sau. Đã đẩy lên GitHub thì coi như lộ.

## 2. Mật khẩu

- Băm bằng bcrypt `DefaultCost`. Không bao giờ lưu bản rõ, không log, không trả `password_hash` ra API.
- So sánh bằng `bcrypt.CompareHashAndPassword`, không tự so chuỗi.
- Mật khẩu mẫu `Password@123` trong `cmd/seed` **chỉ dùng cho môi trường phát triển**. `cmd/seed`
  và các lệnh migration phá dữ liệu đã được chặn cứng: `APP_ENV != dev` là từ chối chạy. Nếu có
  bản triển khai thật thì phải buộc đổi mật khẩu lần đăng nhập đầu.
- Chính sách độ dài và độ phức tạp: **chưa chốt**. Tối thiểu nên là 8 ký tự.

## 3. Token và phiên

Chi tiết đầy đủ ở `roles_permissions.md` mục 4. Những điểm bảo mật không được bỏ:

- Access token và refresh token đều là cookie **HttpOnly + Secure + SameSite=Lax**. Không
  `localStorage`, không `sessionStorage` — dính XSS là mất token.
- Refresh token **lưu bản băm SHA-256** trong DB, không lưu bản gốc. Rò rỉ bảng thì token vẫn vô dụng.
- Xoay vòng mỗi lần refresh, và **phát hiện tái sử dụng**: dùng lại token đã thu hồi là dấu hiệu bị
  đánh cắp → thu hồi cả họ token của phiên đó.
- Refresh cookie giới hạn `Path=/api/auth/refresh` để không gửi kèm mọi request.
- Đăng xuất phải **thu hồi ở phía server**, không chỉ xoá cookie ở trình duyệt.

## 4. Kiểm soát truy cập

Đây là chỗ dễ thủng nhất của hệ thống nhiều vai trò.

- **IDOR**: mọi endpoint có `{id}` đều phải kiểm tra quyền sở hữu, không được tin định danh do
  client gửi. Cụ thể: PM chỉ thao tác trên dự án có `pm_employee_id` là mình; nhân viên chỉ thao tác
  trên task có `current_assignee_id` là mình; nhân viên chỉ xem hồ sơ của chính mình.
- **Không tin `employee_id` trong body request.** Luôn lấy từ claims của token.
- Sai quyền trả **403**, không tồn tại trả **404**. Đừng để thông báo lỗi tiết lộ tài nguyên có tồn
  tại hay không cho người không có quyền.
- Nhân viên **không** được tự sửa `employee_skills` đã duyệt; PM **không** được thay vai trò duyệt
  của ADMIN. Hai ranh giới này tài liệu nêu tường minh ở mục 2.4.1.
- Kiểm tra `users.status` tại các thao tác nhạy cảm (duyệt kỹ năng, xác nhận phân công, đóng dự án)
  để bù cửa sổ trễ 15 phút của access token.

## 5. Chống tiêm mã

- **Luôn dùng tham số `$1, $2`** của pgx. Cấm tuyệt đối nối chuỗi SQL từ dữ liệu người dùng.
- Bộ lọc động của UC003.2 (kỹ năng, vị trí, phòng ban, cấp độ, trạng thái) là chỗ rủi ro nhất vì
  số điều kiện thay đổi. Dựng bằng query builder hoặc ghép **placeholder có đánh số**, không ghép giá trị.
- Cột sắp xếp và hướng sắp xếp nếu cho client chọn thì phải đối chiếu **danh sách trắng**.
- Frontend: React tự escape; chỉ khi dùng `dangerouslySetInnerHTML` mới có rủi ro XSS — tránh hẳn.

## 6. Kiểm tra đầu vào

- Validate ở tầng transport (định dạng, kiểu, độ dài) và ở tầng service (quy tắc nghiệp vụ).
  Ràng buộc DB là **lưới an toàn cuối cùng**, không phải nơi kiểm tra chính.
- Giới hạn kích thước body request để tránh nuốt payload khổng lồ.
- Phân trang bắt buộc cho mọi danh sách, có chặn trên cho `limit`.

## 7. Chống dò mật khẩu

- **Giới hạn tần suất `POST /api/auth/login`** theo IP và theo username. Cách hiện thực **chưa chốt**.
- Router cố ý **không dùng `middleware.RealIP`** của chi: nó ghi đè `r.RemoteAddr` bằng
  `X-Forwarded-For` do client gửi mà không kiểm tra proxy có tin cậy hay không, nên bất kỳ ai cũng
  giả được địa chỉ nguồn và vượt qua giới hạn theo IP (GHSA-3fxj-6jh8-hvhx). Chỉ thêm lại khi
  triển khai sau reverse proxy, kèm danh sách trắng proxy.
- Thông báo lỗi đăng nhập phải mơ hồ: "Tên đăng nhập hoặc mật khẩu không đúng" — không tách riêng
  "không tồn tại tài khoản này", vì như vậy là để lộ username có thật.
- Cân nhắc khoá tạm sau N lần sai. Nếu làm thì phải có đường mở khoá cho ADMIN.

## 8. Vệ sinh log

Không bao giờ ghi vào log: mật khẩu, `password_hash`, access/refresh token, `JWT_SECRET`,
`DATABASE_URL`, nội dung CV, thông tin liên hệ cá nhân của nhân sự.

Được phép ghi: `user_id`, `employee_id`, tên endpoint, mã lỗi, thời gian xử lý.

## 9. Tải tệp CV (M7)

Đây sẽ là bề mặt tấn công lớn nhất của hệ thống:

- Kiểm tra **kiểu tệp thật** bằng magic bytes, không tin `Content-Type` hay đuôi tệp client gửi.
- Giới hạn dung lượng cứng ở tầng handler.
- **Sinh tên tệp mới** khi lưu, không dùng tên gốc — chống path traversal và ghi đè.
- Không phục vụ tệp trực tiếp từ thư mục lưu trữ. Tải xuống phải đi qua endpoint có kiểm tra quyền.
- Đặt `Content-Disposition: attachment` khi trả tệp, tránh trình duyệt tự render.
- Parser PDF/DOCX là mã của bên thứ ba xử lý dữ liệu không tin cậy — chọn thư viện còn được bảo trì.

## 10. Header và vận chuyển

Khi phục vụ SPA nhúng, đặt các header: `X-Content-Type-Options: nosniff`,
`X-Frame-Options: DENY`, `Referrer-Policy: same-origin`, và `Content-Security-Policy` đủ chặt
(SPA nhúng cùng origin nên không cần nguồn ngoài).

Cookie chỉ đặt `Secure` khi chạy HTTPS. Ở môi trường dev qua `localhost` thì cần cờ cấu hình để tắt,
**mặc định phải là bật**.

## 11. Vết kiểm toán

Schema đã hỗ trợ sẵn, đừng phá:

- `task_status_history` ghi mọi lần đổi trạng thái kèm người thực hiện.
- `assignments` là **append-only** về nghiệp vụ: phân công lại thì thêm dòng, không sửa dòng cũ.
- `employee_skills.approved_by_employee_id` và `approved_at` cho biết ai duyệt, lúc nào.
- FK trỏ tới `employees` cố ý **không** `ON DELETE CASCADE` để không mất vết lịch sử.

## 12. Phụ thuộc

- Chạy `govulncheck ./...` định kỳ và trước khi đóng gói.
- Thêm dependency mới phải hỏi trước (quy tắc trong `CLAUDE.md`), một phần vì lý do bảo mật chuỗi
  cung ứng.

## 13. Còn phải chốt

| Hạng mục | Trạng thái |
|---|---|
| Chính sách độ mạnh mật khẩu | Chưa chốt |
| Cách giới hạn tần suất đăng nhập | Chưa chốt |
| Có khoá tạm sau N lần sai không | Chưa chốt |
| Nội dung `Content-Security-Policy` cụ thể | Chốt khi khởi tạo frontend |
| Mã hoá tệp CV khi lưu trữ | Chưa chốt, phụ thuộc lựa chọn nơi lưu ở M7 |
