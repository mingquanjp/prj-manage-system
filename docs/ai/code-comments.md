# Comment trong mã nguồn

Comment phục vụ người đọc và bảo trì code, kể cả khi họ chưa đọc chat hoặc ticket. Quy ước này áp dụng cho Go, TypeScript/React và test do dự án sở hữu; không sửa mã vendor, file sinh tự động hoặc migration lịch sử chỉ để đổi văn phong. Quy tắc viết code đơn giản ở [AGENTS](../../AGENTS.md) vẫn áp dụng.

## Khi nào cần comment?

Trước khi thêm comment, xem tên biến/hàm và luồng xử lý đã diễn đạt rõ chưa. Nếu code khó hiểu, ưu tiên làm rõ code trong phạm vi đang sửa. Thêm comment khi người đọc còn cần biết thông tin không thể suy ra trực tiếp:

- **Lý do và hậu quả:** vì sao phải giữ snapshot, không retry, dùng thứ tự khóa này hoặc xử lý lỗi theo cách khác thường.
- **Ràng buộc:** nguồn quyền, thứ tự thao tác, thời hạn, đơn vị, NULL khác trường vắng hoặc giới hạn tương thích của thư viện/giao thức.
- **Đánh đổi có thật:** cách làm được chọn vì điều kiện nào; workaround phụ thuộc lỗi bên ngoài và khi nào có thể bỏ.
- **Contract của API code:** caller được phép truyền gì, phải làm gì tiếp, kết quả/lỗi có ý nghĩa gì; có side effect hoặc bảo đảm đồng thời nào.

“WHY thay vì WHAT” là ưu tiên cho comment giải thích implementation, không phải lệnh cấm mô tả chức năng. Doc comment của hàm/type cần mô tả mục đích và cách dùng. Regex, thuật toán hoặc định dạng khó đọc có thể cần giải thích cách hoạt động; không viết lại từng câu lệnh thành văn xuôi.

## Đặt comment ở đâu?

| Loại | Vị trí và nội dung |
| --- | --- |
| Implementation | Ngay trên nhánh, phép chuyển đổi hoặc nhóm lệnh cần giải thích; sát phạm vi nhỏ nhất mà lý do áp dụng. Thường một hoặc hai câu là đủ. |
| Hàm/type/port | Ngay trước khai báo. Nêu contract mà caller cần biết; không kể chi tiết nội bộ dễ thay đổi. Go doc comment bắt đầu bằng tên symbol; bổ sung cho exported API mới hoặc đang sửa theo hướng dẫn Go. |
| TypeScript | Dùng `//` cho lý do trong implementation; dùng `/** ... */` khi nội dung cần hiện trong tài liệu/hover cho caller. Không lặp type đã có trong chữ ký bằng `@param`/`@returns` chỉ để đủ mẫu. Component/prop tự rõ không cần doc comment máy móc. |
| Test/fixture | Gần dữ liệu dựng đặc biệt, barrier hoặc fault injection: giải thích tình huống và giới hạn phép kiểm. Tên test và assertion giữ kết quả mong đợi, không thêm comment “test đã PASS”. |
| Chỉ thị công cụ | Giữ đúng cú pháp/vị trí của `//go:build`, `//go:embed`, license hoặc chỉ thị thực sự cần thiết. Không dịch, xóa hay thêm suppression để làm cổng xanh. |
| TODO/FIXME | Chỉ cho việc còn thiếu đã xác định: nêu việc cần làm, điều kiện gỡ và issue/ticket có thật nếu có. Không đoán mã, hứa “sẽ sửa sau” hoặc dùng TODO thay xử lý lỗi chặn. |

Nếu lý do chi phối cả package, đặt một mô tả tại biên package/port hoặc tài liệu kiến trúc rồi dẫn nguồn cần thiết; không rải cùng một đoạn khắp caller. Không dùng comment cuối dòng cho giải thích dài.

## Cách viết và thông tin cần để đúng chỗ

- Viết tiếng Việt rõ nghĩa, câu ngắn, chủ thể cụ thể; giữ nguyên tên kỹ thuật cần tra cứu. Nêu rõ “transaction DB”, “phiên”, “quyền server” khi một từ có thể mang nhiều nghĩa. Không đặt giới hạn số dòng cứng thay cho việc đọc hiểu.
- Nêu sự thật và lý do hiện hành. Tránh giọng nhật ký như “đã sửa”, “đã nối”, “review xong”, “task sau làm”, “theo yêu cầu của người dùng”. Không dùng “đã/chưa” làm tiêu chí xóa tự động: “DB chưa xác nhận commit” có thể là điều kiện runtime quan trọng.
- Lịch sử thay đổi thuộc Git/PR; tiến độ, lệnh và kết quả thực chạy thuộc runbook/evidence/Jira. Trong code chỉ giữ lý do còn cần để sử dụng hoặc sửa đúng. Mã ticket/FR không thay lời giải thích; link chỉ thêm khi giúp tra một quyết định hoặc lỗi cụ thể.
- Không ghi điều hiển nhiên như “tăng biến”, “gọi API”, “duyệt danh sách”, hoặc nhận xét chung “đảm bảo an toàn”, “tránh lỗi” mà không chỉ ra ràng buộc nào.
- Không che lỗi thiết kế bằng comment, không để code đã comment-out làm bản lưu. Khi sửa code phải cập nhật/xóa comment đã sai; giữ comment hữu ích về bảo mật, transaction, Unicode, concurrency và giới hạn test.
- Không đưa secret, token, mật khẩu hoặc PII thật vào comment. Doc comment không sao chép toàn bộ proposal/contract; dẫn đúng nguồn khi cần chi tiết.

## Ví dụ theo code dự án

### Giữ kết quả tài khoản khi gửi thư lỗi

Comment cũ thiếu ngữ cảnh và dùng “lỗi ngoài dự kiến” quá rộng:

```go
// Đã commit: giữ kết quả tạo cả khi adapter gửi thư gặp lỗi ngoài dự kiến.
```

Tại nhánh `recover`, giải thích điều cần bảo toàn và lý do kết quả thư chưa xác định:

```go
// Tài khoản đã lưu thành công; lỗi gửi thư không được báo thành lỗi tạo tài khoản.
// Panic không cho biết thư đã gửi hay chưa, nên giữ trạng thái chưa xác định.
defer func() {
    if recover() != nil {
        status = MailUnknown
    }
}()
```

Đây là giải thích cho nhánh xử lý panic đã tồn tại, không phải quy tắc thêm `recover` ở mọi hàm hoặc tự thay contract gửi thư.

### Không mô tả lại phép gán

Thừa vì chỉ đọc lại câu lệnh:

```ts
// Tăng generation.
generation++
```

Hữu ích vì nêu lỗi mà cơ chế này ngăn chặn:

```ts
// Phản hồi từ phiên cũ phải bị bỏ để không khôi phục dữ liệu người trước.
generation++
```

### Contract cho caller không chỉ nói WHY

```go
// PreviewEmployeeCode trả mã dự kiến mà không cấp hoặc giữ chỗ số thứ tự.
// Caller phải dùng mã từ kết quả tạo tài khoản làm mã chính thức.
```

Comment này cần thiết vì caller có thể hiểu nhầm preview là cấp mã. Trong test, thay câu “chưa có adapter ở ticket sau” bằng mô tả đúng phép kiểm, ví dụ: `// Ghi FK trực tiếp để kiểm xóa danh mục; không kiểm transaction tạo tài khoản.`

## Review trước bàn giao

Khi tự review hoặc giao agent review, đọc comment mới/sửa cùng code xung quanh, kể cả comment cũ bị thay đổi này làm sai:

1. Nếu bỏ comment, người đọc có mất một lý do/contract cần thiết không? Nếu không, bỏ comment thừa hoặc làm rõ tên/luồng code.
2. Nội dung còn đúng với đường chạy, lỗi, quyền và phạm vi test thực tế không? Có cần bằng chứng hoặc link nguồn không?
3. Comment nằm đúng chỗ, đủ ngữ cảnh và không phụ thuộc lịch sử chat/ticket không?
4. Có đang ghi tiến độ, bằng chứng PASS hoặc nhắc lại code không? Chuyển về tài liệu phù hợp hoặc bỏ; không xóa giới hạn của fake/fixture khi rút gọn.

Đây là bước review ngữ nghĩa, không phải chỉ tiêu số lượng comment. Formatter/lint hay tìm từ khóa không chứng minh comment đúng; không dựng linter cấm từ “commit”, “đã” hoặc “adapter”. Chỉ sửa comment thuộc phạm vi đang làm; việc dọn rộng hơn cần phạm vi riêng.

## Nguồn tham khảo và cách áp dụng

- [Google Go Style Guide — clarity](https://google.github.io/styleguide/go/guide#clarity): ưu tiên người đọc, tên/luồng rõ và lý do hữu ích thay comment lặp code.
- [Google Engineering Practices — comments trong code review](https://google.github.io/eng-practices/review/reviewer/looking-for.html#comments): rà sự cần thiết, phân biệt comment implementation với tài liệu API; không dùng comment để che code khó hiểu.
- [Go Doc Comments](https://go.dev/doc/comment): vị trí và nội dung tài liệu package/type/function cho caller.
- [Google TypeScript Style Guide — JSDoc versus comments](https://google.github.io/styleguide/tsguide.html): chọn hình thức theo người sử dụng tài liệu và người đọc implementation.

Các nguồn trên là cơ sở tham khảo. Ngôn ngữ tiếng Việt, cách ghi TODO và bước review ở đây là quy ước của repo; không nhận toàn bộ style guide của Google, không thêm dependency hoặc yêu cầu viết comment cho mọi dòng/hàm.
