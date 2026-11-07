# Chốt bổ sung phục vụ giao diện — 20/10/2026

Tiếp nối [baseline 12/10](business-baseline-2026-10-12.md), [nguồn nhân lực chung 13/10](business-baseline-2026-10-13.md) và [quyết định DB/Bàn giao](design/database-review-decisions-2026-10-14.md). Bổ sung quyền tạo ban đầu và quyết định địa chỉ đã được người dùng xác nhận; các ranh giới khác giữ theo nguồn hiện hành.

> Tiếp nối 26/10: phần cấp bậc và tổng số bảng ưu tiên [quyết định cấp bậc theo công ty](design/database-review-decisions-2026-10-14.md#quyết-định-tiếp-nối-26102026--cấp-bậc-theo-công-ty-hiện-hành). Mốc 20/10 dưới đây có 32 bảng; hiện tại là 33, cấp bậc dùng danh mục HR và khóa ngoại nullable. Các quy tắc tạo ban đầu/địa chỉ giữ nguyên.

## Quyết định đã xác nhận

**Admin tạo tài khoản kèm hồ sơ nhân viên ban đầu trong một lần; HR quản lý/cập nhật hồ sơ về sau.**

- Tạo user và employee trong cùng giao dịch, không tạo tài khoản mồ côi hoặc hồ sơ tạm chờ bước duyệt mới.
- Thông tin bắt buộc ban đầu theo dictionary: username/thông tin xác thực theo chính sách, mã nhân viên, họ tên, email công ty, vị trí công việc; phone/phòng ban/cấp bậc nullable theo thiết kế. Trạng thái và định danh do server quản lý theo hợp đồng hiện hành.
- Admin được nhập thông tin hành chính cần cho tạo ban đầu. Quyền tạo không cấp quyền sửa hồ sơ hành chính thường xuyên, duyệt thông tin cá nhân/kỹ năng hoặc quản lý danh mục thuộc HR/Giám đốc.
- HR tiếp tục cập nhật hồ sơ trực tiếp, duyệt đề nghị fullname/phone/address và quản lý email công ty/cơ cấu theo quyền hiện hành.
- User_roles chỉ ADMIN/HR/DIRECTOR; PM/PTCM vẫn theo dự án. Không tự gán trách nhiệm dự án khi tạo tài khoản.
- Riêng quyết định tạo ban đầu không thêm bảng, cột, trạng thái hoặc phê duyệt. Trường địa chỉ bổ sung theo quyết định tiếp nối bên dưới. Chính sách cấp/reset mật khẩu vẫn cần đặc tả; quyết định chưa cần trạng thái buộc đổi tại thời điểm này được thay ở chốt 06/11 bên dưới; cấu trúc cụ thể đã chốt tại phần tiếp nối 06/11 bên dưới.

## Quyết định tiếp nối — địa chỉ khi hoàn thiện hồ sơ

**Địa chỉ là thông tin phải bổ sung khi hoàn thiện hồ sơ, không bắt buộc tại thời điểm tạo tài khoản.** Quyết định này thay cách gọi “địa chỉ tùy chọn” và phương án bắt Admin nhập/đặt NOT NULL ngay đã được trao đổi trước đó.

- Admin không bắt buộc nhập địa chỉ khi tạo tài khoản kèm hồ sơ ban đầu. Không thêm bước HR duyệt tạo tài khoản.
- Thiết kế `employees.address` là `TEXT NULL`, mặc định NULL. NULL biểu thị chưa bổ sung, không phải địa chỉ giả hoặc dấu hiệu khóa quyền sử dụng.
- Chủ hồ sơ bổ sung/sửa qua yêu cầu cập nhật thông tin cá nhân hiện có; HR chấp thuận/từ chối. HR cũng được cập nhật trực tiếp theo quyền hành chính, có lịch sử.
- Khi cung cấp hoặc đề nghị thay đổi địa chỉ phải có nội dung sau khi bỏ khoảng trắng đầu/cuối. Không nhận NULL, chuỗi rỗng hoặc chỉ khoảng trắng như thao tác xóa; địa chỉ đã có chỉ được thay bằng địa chỉ có nội dung.
- Bỏ qua trường trong patch nghĩa là không thay đổi. `before_values.address` có thể NULL khi bổ sung lần đầu; giá trị đề nghị phải là chuỗi có nội dung. Pending/từ chối không đổi địa chỉ chính thức; chấp thuận đối chiếu dữ liệu nền đúng trường, không ghi đè giá trị đã đổi.
- Hồ sơ chưa có địa chỉ hiển thị “Chưa cập nhật” và nhắc bổ sung; không chặn đăng nhập, nhận việc hoặc thực hiện task vì thiếu địa chỉ. Không thêm trạng thái hoàn thiện hồ sơ, hạn bổ sung hoặc quy trình duyệt mới.
- Không trả địa chỉ trong tra cứu nghề nghiệp, ứng viên phân công hoặc snapshot gợi ý. Quyền xem hồ sơ/căn cứ theo mục đích không tự mở thêm quyền xem địa chỉ.
- Chỉ thêm một cột trong thiết kế, giữ 32 bảng; dùng JSON patch của `employee_profile_change_requests` hiện có. Không thêm bảng địa chỉ, danh mục địa giới, chỉ mục hoặc tích hợp bản đồ.
- Đợt này cập nhật tài liệu/dictionary/DBML, chưa sửa hoặc chạy migration. Khi chuyển đổi dữ liệu cũ, địa chỉ chưa biết giữ NULL, không bịa dữ liệu để lấp chỗ trống.

## Quyết định tiếp nối — điều kiện ghi nhật ký (21/10/2026)

Người dùng đã xác nhận luồng: Được phân công → Bắt đầu → thực hiện/ghi nhật ký → nộp → nghiệm thu. Quyết định này thay quyền ghi trực tiếp khi ASSIGNED/PAUSED trong R01 cũ, kể cả phần baseline 30/09 còn được dẫn chiếu.

- Chỉ người hiện được giao hợp lệ được tạo/sửa/xóa mềm nhật ký trực tiếp khi task IN_PROGRESS và project IN_PROGRESS. ASSIGNED phải Bắt đầu; PAUSED phải Tiếp tục trước, vẫn được đọc nhật ký đã có.
- Bắt đầu là xác nhận bắt tay thực hiện, không thêm bước xác nhận nhận việc; Start/Resume không sinh giờ hoặc chạy timer. Lưu nhật ký không tự đổi trạng thái.
- Giữ kiểm tra ngày thực tế thuộc lượt giao/giai đoạn triển khai, không tương lai, phút nguyên và tổng xuyên dự án. Không thêm cột thời gian hay bộ đo để suy giờ thực tế từ Start/Resume.
- Giữ khóa khi PENDING_REVIEW, hiệu chỉnh khi khóa theo quyền hiện hành và quy tắc thiếu log không chặn nộp/nghiệm thu. ASSIGNED/PAUSED không tạo quy trình duyệt ngoại lệ mới; dùng Bắt đầu/Tiếp tục, không sửa ngược lịch sử trạng thái.
- Không đổi cấu trúc DB, công thức hoặc mở rộng nghiệm thu.

## Bộ đầu vào giao diện

[Bộ Figma Make](design/figma-make/README.md) chuyển nghiệp vụ thành màn hình và tình huống kiểm tra; không phải nguồn để tự mở rộng nghiệp vụ. “Yêu cầu chỉnh sửa kết quả” là kết luận review chưa đạt; thu hồi phân công là thao tác PTCM. Không suy thêm nút nhân viên tự trả việc từ tên gọi cũ.

File đích do người dùng cung cấp: [DATN](https://www.figma.com/make/fOXzi51sKaRQbuSKmHZQmD/DATN). Cập nhật tài liệu không chứng minh đã sửa/kiểm thử prototype.


## Quyết định tiếp nối 28/10/2026 — tự cấp mã và nhãn quyền tài khoản

- Đã chốt: hệ thống tự cấp mã nhân viên khi tạo tài khoản kèm hồ sơ; form hiển thị chỉ đọc mã dự kiến, sau lưu trả mã được cấp. Dải E001, E002… dùng trong bản hiện tại. Không đổi mã nhân viên đã tồn tại.
- Chốt làm rõ sau review: tên đăng nhập e001, e002… theo mã nhân viên chỉ là quy ước tạm của demo, hiển thị chỉ đọc. Chưa chốt định danh/cách đăng nhập cho backend (email, mã nhân viên hay tên riêng); cần đặc tả cùng xác thực và kênh cấp tài khoản. Không đổi tên đăng nhập cũ hoặc coi giá trị cấp trong demo là thông tin đăng nhập thật.
- Cấp mã dựa trên dữ liệu hiện hành, không dùng số lượng hồ sơ cộng một, không dùng lại mã hồ sơ vô hiệu hóa hoặc tài khoản bị khóa. Backend phải bảo đảm uniqueness/cấp mã khi tạo đồng thời; cơ chế cấp số server cụ thể sẽ thiết kế cùng backend, không coi cách tính trong demo là giải pháp đồng thời production.
- Thay yêu cầu nhập username/employee_code tại bước tạo trước đây; hai cột vẫn bắt buộc và duy nhất. Không thêm bảng/cột hoặc đổi cấu trúc DB; không chạy migration từ đợt UI.
- Nhãn UI “Quyền hệ thống” thay “Vai trò hệ thống”; “Quản lý nhân sự” là nhãn quyền HR. Mã kỹ thuật ADMIN/HR/DIRECTOR giữ nguyên. Nhân viên có quyền nền; vị trí/phòng ban/cấp bậc không tự cấp quyền. PM/PTCM vẫn là trách nhiệm dự án.
- Email công ty do Admin nhập khi tạo, HR quản lý về sau; không đồng nhất email với tên đăng nhập và không tạo hộp thư. Mật khẩu, kênh cấp tài khoản và xác thực thật tiếp tục hoãn.

## Quyết định tiếp nối 28/10/2026 — thay PM, hạn task và trao đổi

Người dùng duyệt triển khai ba phần sau rà soát đợt 3; không thay cấu trúc database:

- Giám đốc thay PM theo UC001 hiện hành; không có bước PM cũ duyệt/bàn giao, không thay PTCM/phân công hoặc actor lịch sử. Cho PLANNING/IN_PROGRESS/COMPLETED; CLOSED/CANCELLED chỉ đọc.
- PTCM gia hạn task chưa kết thúc trong project PLANNING/IN_PROGRESS: hạn mới sau hạn hiện hành, trong kế hoạch project; lý do và lịch sử trước/sau bắt buộc. Không đổi estimate, ngày bắt đầu đã lưu hoặc đặc tả. Task chưa có cặp ngày được đặt cặp kế hoạch đầu tiên theo dictionary 5A; giữ cặp đầu tiên làm mốc gốc. Không dựng chỉnh kế hoạch dự án trong đợt này. Ngày quá hạn không chặn effort thực tế hợp lệ.
- Trao đổi trong task dành cho người hiện được giao và PTCM hiện hành, theo mục đích hỏi/báo khó khăn/hướng dẫn. Có tác giả/thời điểm/nội dung; không sửa/xóa, thread/tệp, mention hoặc thay quyết định nghiệm thu. Đọc theo quyền hiện hành; không quyền tự động cho PM-only/Admin/Giám đốc. Chỉ gửi khi project PLANNING/IN_PROGRESS và task chưa DONE/CLOSED/CANCELLED. Người cũ đã mất phân công không dùng luồng trao đổi để đọc task hiện hành; lịch sử đóng góp riêng vẫn giữ.
- Làm rõ đặc tả sau phân công chưa chốt cách thể hiện, không mở lại biểu mẫu sửa scope. CV làm căn cứ MANUAL cần thiết nhưng hoãn cùng lưu/đọc tệp CV có phân quyền; trích xuất bằng model tiếp tục hoãn. Thông báo dùng chung là đợt riêng.

Frontend chỉ mô phỏng trong phiên; các điều kiện phải kiểm tra lại tại server khi triển khai backend.

## Quyết định tiếp nối 05/11/2026 — định danh đăng nhập backend

- Người dùng đã chọn **email công ty** làm định danh đăng nhập cho phiên bản backend đầu tiên. Email kết hợp mật khẩu; không dùng e001/mã nhân viên hoặc danh tính demo để xác thực thật.
- `username` vẫn là dữ liệu riêng theo thiết kế, không đổi tên tài khoản cũ hoặc xóa/đổi cấu trúc cột. Quy tắc cấp username cho tài khoản mới, kênh cấp/reset và chính sách mật khẩu chưa được chốt qua lựa chọn email.
- Quy tắc chuẩn hóa email/va chạm, thời hạn phiên và môi trường dữ liệu thử nghiệm còn cần duyệt trong [proposal auth](proposals/backend-auth-session.md). Không coi các con số/phương án đề xuất trong proposal là quyết định đã xác nhận; chưa chạy migration hoặc triển khai auth.
- Chốt này thay phần “chưa chốt định danh đăng nhập backend” của 28/10, không thay quyền HR quản lý email công ty hoặc đồng nhất email với username.

## Quyết định tiếp nối 06/11/2026 — cấp mật khẩu tạm qua email

- Người dùng xác nhận: khi Admin tạo tài khoản kèm hồ sơ, hệ thống sinh mật khẩu tạm ngẫu nhiên và gửi thông tin tài khoản tới email đã đăng ký. Demo dùng hộp thư `@sis.hust.edu.vn` thật của người dùng; không giới hạn sản phẩm vào tên miền này và không sử dụng mật khẩu hộp thư làm mật khẩu ứng dụng.
- Nhân viên đăng nhập bằng email và mật khẩu tạm, bắt buộc đổi mật khẩu trước khi truy cập chức năng nghiệp vụ. Đổi thành công làm mật khẩu tạm mất hiệu lực; sau đó đăng nhập bằng mật khẩu mới.
- Chốt cuối cùng cùng ngày: **mật khẩu tạm chưa có thời hạn trong phiên bản hiện tại**; thay đề xuất trước đó về hết hạn/cấp lại do hết hạn. DB chỉ lưu bản băm, không ghi mật khẩu vào log hoặc trả trong giao diện quản trị. Nội dung thư yêu cầu đổi ngay lần đầu; server vẫn bắt buộc đổi, không chỉ nhắc bằng email. Chính sách mật khẩu mới, gửi lỗi/gửi lại còn cần thiết kế/chốt; chưa quyết định Admin reset dùng cùng luồng. Thời hạn phiên đăng nhập vẫn độc lập, không bị bỏ.
- Cấu trúc đã được người dùng xác nhận: thêm `users.must_change_password BOOLEAN NOT NULL DEFAULT false`. Tài khoản cũ giữ false, không reset/buộc đổi hàng loạt; luồng tạo tài khoản cấp mật khẩu tạm phải chủ động ghi true cùng bản băm. Đổi thành công thay bản băm và đặt false trong cùng giao dịch. **Không thêm `temporary_password_expires_at`**, không thêm bảng hoặc khôi phục `users.version`. Chốt này thay đúng phần loại bỏ must_change_password trong sổ quyết định cũ; không quyết định đầy đủ contract phiên/đổi mật khẩu.
- Đây là yêu cầu đích mới, chưa triển khai gửi thư/xác thực. Proposal auth và WBS cần bổ sung thiết kế cổng đổi mật khẩu; không gọi phiên mật khẩu tạm là phiên có đủ quyền nghiệp vụ. Không tự triển khai SMTP, mời qua link hoặc quên mật khẩu.


## Quyết định tiếp nối 06/11/2026 — chốt lựa chọn auth A02–A09

Người dùng đã xác nhận bộ phương án A02–A09 dưới đây, tiếp nối quyết định email và mật khẩu tạm cùng ngày. Thay đúng các ghi nhận “chưa chốt” tương ứng ở trên; không chốt reset của Admin, kênh gửi lỗi/gửi lại hoặc username. Thiết kế chưa là chức năng đã triển khai.

- **A02 — Email:** bỏ khoảng trắng đầu/cuối khi nhận email đăng nhập, so không phân biệt hoa/thường; không tự viết lại email đã lưu. Nhiều tài khoản khớp sau chuẩn hóa thì dừng với lỗi hệ thống chung, không chọn ngẫu nhiên. Kiểm kê/mapping/ràng buộc dữ liệu vẫn cần review, không tự thêm index hoặc sửa dữ liệu.
- **A03 — Phiên:** hạn tuyệt đối 8 giờ từ lúc đăng nhập, không tự gia hạn, không idle timeout hoặc “ghi nhớ đăng nhập”. Hết hạn phải đăng nhập lại. Hạn này áp dụng cả phiên hạn chế, độc lập với mật khẩu tạm chưa có hạn.
- **A04 — Dọn phiên:** đợt auth này chưa xóa vật lý session hết hạn/thu hồi; chúng vẫn bị server từ chối. Retention và lịch dọn chốt trước vận hành lâu dài, chưa thêm job/cột.
- **A05 — Test:** PostgreSQL local riêng, đề xuất triển khai bằng Docker, cùng phiên bản lớn với Neon (kiểm kê hiện tại là 18); chỉ dữ liệu giả, không dùng DB phát triển làm DB test. Kiểm tra Docker/dựng môi trường và fixture là task riêng, chưa diễn ra qua việc duyệt này.
- **A06 — Migration:** tiến thêm trên lịch sử hiện có, giữ nguyên 000001, không reset hoặc tạo lại baseline. Chỉ nâng phần cần cho feature đã duyệt; DDL, mapping quyền/dữ liệu legacy, tương thích và kế hoạch áp DB phải review riêng. Không đồng nghĩa cho phép áp migration lên Neon.
- **A07 — Cookie/CSRF:** `staffing_session`, secret ngẫu nhiên và session phía server theo security; kiểm Origin và CSRF gắn phiên, FE local dùng proxy `/api`. Cơ chế HMAC-SHA256 đã mô tả trong proposal được duyệt. Login cấp secret mới; phiên cookie cũ hợp lệ của cùng user bị thu hồi trong transaction. Không thu hồi xuyên user hoặc mọi thiết bị như tác dụng phụ của login.
- **A08 — Buộc đổi:** cùng bảng/cookie session; cờ true chỉ cho me/csrf/logout/đổi mật khẩu lần đầu, server chặn mọi API nghiệp vụ. Đổi thành công thay hash, đặt cờ false và thu hồi tất cả phiên của user trong cùng transaction; sau đó xóa cookie và yêu cầu login lại. Không thêm scope/type/version vào session.
- **A09 — Mật khẩu mới:** tối thiểu 15 ký tự Unicode (đếm code point), tối đa 72 byte UTF-8 do bcrypt hiện hành; không cắt mật khẩu. Cho phép khoảng trắng/chữ có dấu/paste, không ép đủ nhóm hoa/thường/số/ký hiệu, không trim hoặc tự chuẩn hóa mật khẩu. Phải xác minh mật khẩu hiện tại và mật khẩu mới khác mật khẩu hiện tại; nhập lại ở FE để đối chiếu, không thêm trường API. Áp chính sách này cho đặt mật khẩu mới trong đợt đổi lần đầu, không áp ngược lên login tài khoản cũ.
- **Limiter đổi mật khẩu:** tham số thử nghiệm đã duyệt là 5 lần xác minh mật khẩu hiện tại sai/15 phút/user và 30 request/5 phút/IP trước bcrypt; trả 429/Retry-After, không khóa tài khoản vĩnh viễn. Bộ đếm có hạn và dọn bộ nhớ, không ghi password làm key; đây không phải job dọn session DB. Login giữ tham số thử nghiệm riêng theo security.

Chấp thuận A02–A09 không đồng nghĩa đã hoàn tất P01/P04, đã duyệt mọi chi tiết OpenAPI/body limit, estimate/dự phòng, cấp quyền dữ liệu legacy hoặc cho phép chạy DB thật. Các phần đó theo proposal và task cụ thể.


## Quyết định tiếp nối 06/11/2026 — giới hạn body JSON auth

Người dùng đã xác nhận giới hạn body JSON của `POST /api/auth/login` và `POST /api/auth/change-password` là **8 KiB = 8192 byte**. Body lớn hơn giới hạn trả **413 / PAYLOAD_TOO_LARGE**, không xác thực bằng bcrypt hoặc ghi dữ liệu từ body vượt giới hạn. Áp cho toàn bộ body request, không chỉ trường mật khẩu; không dựa riêng Content-Length để bỏ qua request không khai báo độ dài.

Giới hạn body không thay chính sách mật khẩu mới ≥15 code point Unicode và ≤72 byte UTF-8, không áp giới hạn 8 KiB cho upload CV hoặc mọi endpoint khác. Mốc 8192 byte không vượt giới hạn body nhưng request vẫn phải đạt JSON/allowlist/validation/Origin/CSRF/limiter tương ứng. Khi có HTTP runtime phải kiểm biên 8192/8193 byte, body thiếu/sai Content-Length và không có tác dụng phụ khi bị chặn; hiện chỉ cập nhật thiết kế/contract, chưa có handler.

Chốt này thay đúng các ghi nhận body limit còn đề xuất trong proposal/OpenAPI; không duyệt estimate, mapping legacy hoặc áp DB thật.
