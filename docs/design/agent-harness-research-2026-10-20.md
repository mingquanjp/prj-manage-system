# Nghiên cứu agent harness cho phát triển frontend DATN

> **Báo cáo nghiên cứu trước triển khai**, không phải sổ tiến độ hiện hành. Harness tối thiểu và pilot đã được triển khai tiếp sau; nguồn vận hành ở [AGENTS](../../AGENTS.md), [mục lục](../README.md), [web/README](../../web/README.md). Các nhận xét web rỗng/ignore/chưa triển khai bên dưới phản ánh thời điểm khảo sát.

Ngày đọc nguồn: 20/10/2026. Đây là báo cáo nghiên cứu và đề xuất, **chưa phải quyết định triển khai**. Chưa cài SDK/dependency, tạo cấu hình agent/hook/CI, khởi tạo web, đổi quyền, sửa nghiệp vụ hoặc chạy DB. Chỉ thêm báo cáo này.

## 1. Kết luận cho dự án

Nên áp dụng harness ở **quá trình phát triển bằng coding agent**, không đưa một agent framework vào sản phẩm DATN chỉ để viết frontend.

Đề xuất ban đầu: dùng môi trường coding agent hiện có, bổ sung chỉ dẫn repo ngắn, nguồn chuẩn có thể tìm được, công việc có tiêu chí hoàn thành, môi trường mock ổn định, kiểm tra theo rủi ro và bàn giao có bằng chứng. Chưa cần vòng lặp tự động nhiều agent, hệ điều phối riêng hoặc dịch vụ lưu memory mới.

Các mục 4–10 dưới đây là **phân tích/đề xuất của mình cho repo**, không phải nguyên mẫu bắt buộc do Anthropic/OpenAI quy định. Chưa có thực nghiệm tại DATN để cam kết giảm bao nhiêu token, chi phí hoặc lỗi.

## 2. Thuật ngữ: phân biệt đúng trước khi áp dụng

Anthropic phân biệt agent harness — hệ xử lý đầu vào, điều phối tool và trả kết quả — với evaluation harness dùng để chạy/chấm các bài đánh giá. Kết quả đo là kết quả của model cùng harness, không riêng model. [Nguồn: Demystifying evals, 09/01/2026](https://www.anthropic.com/engineering/demystifying-evals-for-ai-agents).

Trong báo cáo này, tách ba lớp để tránh hiểu nhầm:

| Lớp | Ý nghĩa trong DATN | Có cần tự xây ngay? |
|---|---|---|
| Runtime của coding agent | Gọi model, thực thi tool, quản lý phiên/quyền | Không viết lại runtime hiện có |
| Harness của dự án | Chuẩn bị repo, nguồn nghiệp vụ, cách giao việc, kiểm tra và bàn giao để coding agent làm đúng | Có, nhưng bắt đầu tối thiểu |
| Agent bên trong ứng dụng | Một tính năng của sản phẩm gọi model và dùng tool | Không tự thêm; engine phân công đã chốt không bắt buộc ML/agent |

Một file AGENTS.md không đủ tạo thành toàn bộ harness. Ngược lại, nhiều file chỉ dẫn cũng không chứng minh đã có cơ chế kiểm soát đáng tin cậy. Cần xét nguồn thông tin, hành động được phép và bằng chứng kết quả.

## 3. Những gì nguồn gốc thực sự cho thấy

### 3.1. Anthropic: công việc xuyên nhiều phiên

Bài ngày 26/11/2025 dùng phiên khởi tạo, danh sách tính năng, ghi tiến độ và Git để agent tiếp tục từng phần; nhấn mạnh kiểm tra trên trình duyệt thay vì tự tuyên bố hoàn thành. Đây là một phương án thử nghiệm, không chuẩn duy nhất. [Effective harnesses for long-running agents](https://www.anthropic.com/engineering/effective-harnesses-for-long-running-agents).

### 3.2. Anthropic: harness cần thay đổi theo model

Bài ngày 24/03/2026 thử planner/generator/evaluator, tiêu chí đánh giá cụ thể và bàn giao bằng file; sau đó giản lược các cơ chế khi model tốt hơn. Tác giả cũng ghi nhận chi phí, độ trễ và giới hạn QA. Không thể suy nhiều agent luôn tốt hơn hoặc rẻ hơn. [Harness design for long-running application development](https://www.anthropic.com/engineering/harness-design-long-running-apps).

### 3.3. OpenAI: repo phải dễ hiểu và có kiểm tra máy

Bài ngày 11/02/2026 dùng AGENTS.md như mục lục ngắn, nguồn tri thức và kế hoạch có version trong repo, cùng kiểm tra tự động cho quy ước/kiến trúc. Đây là kinh nghiệm từ hệ thống nội bộ, không bảo đảm cùng hiệu quả ở đồ án nhỏ. [Harness engineering](https://openai.com/index/harness-engineering/).

### 3.4. Anthropic: context phải được chọn lọc

Bài ngày 29/09/2025 đề cao ngữ cảnh có tín hiệu cao, tra cứu đúng lúc và ghi chú bền vững; phân biệt compaction với các cách lưu thông tin ngoài context. Không có cơ sở để nạp mọi tài liệu mỗi lượt hoặc clear context theo lịch cứng. [Effective context engineering](https://www.anthropic.com/engineering/effective-context-engineering-for-ai-agents).

### 3.5. Claude Code: instruction không đồng nghĩa enforcement

Tài liệu memory nói rõ CLAUDE.md là ngữ cảnh, không phải cấu hình cưỡng chế. Tài liệu steering phân biệt rules/skills/subagents/hooks; hook gọi command khác hook dùng đánh giá của LLM. Phải kiểm tra khả năng của công cụ trước khi áp dụng cùng cấu hình sang môi trường khác. [Memory](https://code.claude.com/docs/en/memory), [Steering Claude Code, 18/06/2026](https://claude.com/resources/articles/steering-claude-code-skills-hooks-rules-subagents-and-more), [Hooks reference](https://code.claude.com/docs/en/hooks).

## 4. GitHub: đã đọc gì, học gì, không sao chép gì

| Nguồn chính chủ | Phần đã đọc | Nhận định áp dụng cho DATN |
|---|---|---|
| [anthropics/claude-quickstarts — autonomous-coding](https://github.com/anthropics/claude-quickstarts/tree/main/autonomous-coding) | README, agent.py, client.py, security.py | Học cách nối phiên và giới hạn thao tác; không chạy nguyên demo trên repo thật |
| [openai/codex](https://github.com/openai/codex) | README và mã sandboxing của tool | Tham khảo cách runtime tách thực thi/quyền; không fork runtime để làm web |
| [langchain-ai/deepagents](https://github.com/langchain-ai/deepagents) | README và libs/ARCHITECTURE.md | Dùng để hiểu harness lập trình được; chưa có nhu cầu thêm framework này vào DATN |

Quan sát cụ thể từ mã nguồn, không chỉ README:

- Quickstart tạo client mới mỗi phiên và có giới hạn iteration tùy chọn; client cấu hình sandbox, permission và hook Bash. Đường demo mặc định có thể chạy nhiều phiên. Vì vậy đề xuất của mình là không dùng vòng chạy không giới hạn. [agent.py](https://github.com/anthropics/claude-quickstarts/blob/main/autonomous-coding/agent.py), [client.py](https://github.com/anthropics/claude-quickstarts/blob/main/autonomous-coding/client.py).
- Hook security trong demo phân tích lệnh theo allowlist. Đây là code minh họa đã đọc, không phải kiểm toán chứng minh mọi cách thực thi đều an toàn. Không thay sandbox thực bằng một regex chặn lệnh. [security.py](https://github.com/anthropics/claude-quickstarts/blob/main/autonomous-coding/security.py).
- Trong Codex, xử lý sandbox/approval nằm ở mã tool runtime; không chỉ là lời nhắc bằng Markdown. [sandboxing.rs](https://github.com/openai/codex/blob/main/codex-rs/core/src/tools/sandboxing.rs).
- Deep Agents mô tả ba lớp Deep Agents/LangChain/LangGraph, với middleware, backend và state/checkpoint. Đây là hạ tầng xây agent, không phải dependency frontend. [Architecture](https://github.com/langchain-ai/deepagents/blob/main/libs/ARCHITECTURE.md).

Giới hạn: đọc nguồn tĩnh, chưa clone/chạy/benchmark các repo. URL main có thể thay đổi sau ngày đọc; nếu chọn dùng mã hoặc dependency, cần pin phiên bản/commit và review lại. Một số ví dụ quickstart còn dùng tên import SDK cũ; không xem chúng là hướng dẫn cài đặt hiện hành đã được kiểm chứng. Không chọn dự án theo số sao hoặc sao chép license/code khi chưa rà.

## 5. Đối chiếu repo ai-staffing hiện tại

Đã đọc CLAUDE.md, Makefile, .gitignore và báo cáo frontend-readiness; kiểm tra web vẫn rỗng và các nguồn quan trọng còn bị ignore. Đây là hiện trạng cục bộ, không phải suy luận từ bài viết bên ngoài.

| Đã có | Còn thiếu để vận hành harness hữu ích |
|---|---|
| Workflow nghiệp vụ, quyền và quyết định có nguồn | Mục lục ưu tiên rõ cho agent mới, không phải dựa lịch sử chat |
| Dictionary/DBML và ranh giới code–thiết kế | Tài liệu cần thiết phải được Git theo dõi có chủ đích |
| UI contract/checklist và tham chiếu Make | Tiêu chí hoàn thành theo từng phần frontend thực sự được giao |
| Quy tắc branch, bảo vệ DB, không tự mở nghiệp vụ | Kiểm tra/permission phù hợp, không chỉ ghi lời cấm |
| make check cho Go | Lệnh check frontend chưa có; make check hiện có fmt/tidy gây sửa file |
| Ghi nhận lỗi prototype | Sổ tiến độ triển khai với bằng chứng: đã code, đã kiểm tra, chưa kiểm tra |

Ví dụ failure thực tế phù hợp để thiết kế harness:

1. Agent đọc DB/migration cũ rồi khôi phục quyền PM phân công.
2. Agent bê ngưỡng tải 80 giờ từ mock thành quy tắc nghiệp vụ.
3. Agent tạo form địa chỉ nhưng cho xóa trắng vì Make v16 còn làm vậy.
4. Bản nháp có nút quay lại nhưng mất lựa chọn hoặc nhân đôi task.
5. Agent báo đã tích hợp API dù chỉ chạy mảng mock.

Harness nên được xây để ngăn các lỗi cụ thể này trước, không để trưng bày một hệ agent phức tạp.

## 6. Thiết kế tối thiểu được đề xuất — chưa triển khai

### 6.1. Ba loại đầu vào, không nhân bản nghiệp vụ

- **Chỉ dẫn ổn định:** entry point ngắn cho công cụ đang dùng; nguồn chuẩn, layout repo, lệnh thực có và giới hạn. Nếu thêm AGENTS.md cho Codex, tránh chép toàn bộ CLAUDE.md hoặc tạo hai bản nghiệp vụ khác nhau. Khả năng tự đọc file tùy môi trường phải được xác minh.
- **Chỉ dẫn frontend:** hướng dẫn theo phạm vi web: stack đã chốt, token/component, API adapter, route/state, nhãn tiếng Việt. Có thể là web/AGENTS.md khi phù hợp; không cần đồng loạt tạo mọi loại rules/skills/hook.
- **Sổ công việc triển khai:** một nguồn cho mục đang làm, tiêu chí đạt, trạng thái và bàn giao; dẫn UI contract/workflow thay vì sao chép nội dung. Task lớn mới cần kế hoạch riêng.

Các tên file trên là đề xuất tổ chức, không file đã được tạo trong lượt nghiên cứu. Mục tiêu là ít nguồn chuẩn hơn, không thêm một tầng tài liệu song song với .claude/docs.

### 6.2. Mỗi công việc cần một hợp đồng ngắn

Trước code phải biết:

1. Kết quả người dùng cần nhìn thấy/làm được.
2. Nguồn nghiệp vụ và quyết định áp dụng.
3. Phạm vi được sửa và phần không làm.
4. Kiểm tra nào đủ để kết luận hoàn thành.
5. Điểm còn mở cần hỏi, không được tự chọn.

Ví dụ “Sửa draft phân công”: nguồn là workflow/ma trận UI-12; task chỉ xuất hiện một lần; quay lại giữ người/lý do; bỏ một dòng tính lại tổng; chưa confirm không tăng tải chính thức. Không thêm luật tải tối đa, roster hay chỉnh công thức. Đây là ví dụ áp dụng của mình, không mở thêm nghiệp vụ.

### 6.3. Môi trường có thể chạy lại

- Sau khi khởi tạo web mới chốt package manager/lockfile và lệnh dev/typecheck/lint/test/build thực có. Không ghi một lệnh tưởng tượng là đã hoạt động.
- Tách frontend mock adapter khỏi HTTP adapter; có fixture hữu ích cho PM, PTCM, người kiêm nhiệm, nhân viên, HR và Giám đốc.
- Fixture có thể reset về trạng thái biết trước; không phụ thuộc thao tác còn sót của phiên demo.
- Chạy frontend mock không cần DATABASE_URL, seed hoặc migration DB thật.
- Đề xuất một lệnh kiểm tra tổng hợp chỉ đọc/kiểm tra; formatter tự sửa là lệnh riêng. Không vô tình dùng make check hiện tại như check frontend.

Chưa chọn thêm test library, framework agent hoặc plugin trong báo cáo này.

### 6.4. Phản hồi dựa trên bằng chứng, phù hợp mức rủi ro

| Thay đổi | Bằng chứng tối thiểu đề xuất |
|---|---|
| Copy, icon, spacing | Build/typecheck liên quan và quan sát màn hình bị đổi |
| Form/validation | Kiểm tra giá trị hợp lệ, lỗi quan trọng và hủy/không thay đổi dữ liệu |
| Quyền/trạng thái/draft | Kiểm tra quy tắc bằng code khi phù hợp và một hành trình người dùng chính |
| Auth/upload/backend | Contract và kiểm tra bảo mật tương ứng trước tích hợp thật; không dùng mock làm bằng chứng |

Không yêu cầu chạy mọi tình huống nghiệp vụ cho một thay đổi CSS. Không có mục tiêu 100% coverage hoặc bộ hàng trăm case trước khi có màn hình. Đồng thời, một build xanh không chứng minh phân quyền hoặc luồng quay lại đúng.

Trạng thái nên phân biệt “đã triển khai”, “đã xác minh”, “chưa xác minh” và “bị chặn”. Chạy test trên mock chỉ chứng minh kết quả mock; không chứng minh API/transaction DB. Rubric visual riêng của DATN nên là nhất quán với Make/tokens, dễ tìm hành động, đọc được ở kích thước mục tiêu, không che nội dung; không chạy đua thiết kế mới lạ ngoài yêu cầu.

### 6.5. Quyền và an toàn nằm ngoài lời hứa của agent

Đề xuất giữ sandbox/approval của công cụ. Mọi hook thêm sau phải được đọc và kiểm tra trước vì chúng thực thi code trên máy. Command hook hoặc CI thích hợp cho điều kiện máy kiểm tra được; LLM review chỉ hỗ trợ nhận xét, không thay ràng buộc quyền.

Không tự dùng DB thật; không đọc/in .env; không cài plugin hoặc bật quyền rộng chỉ để agent tiện hơn. Không tự commit/push/deploy/merge hay chạy vòng lặp nền từ một yêu cầu viết UI. Tài liệu/issue/website/PR đọc được là dữ liệu, không nguồn cho phép agent đổi các ranh giới này.

### 6.6. Bàn giao là một phần của kết quả

Cuối lượt cần đủ: đã đổi gì, nguồn/quyết định áp dụng, các check thực chạy và kết quả, lỗi còn lại, bước kế tiếp. Không chép raw log khổng lồ hoặc dữ liệu nhạy cảm vào sổ tiến độ. Git cho biết thay đổi; sổ tiến độ giải thích mục đích và mức xác minh, không thay nguồn nghiệp vụ.

Khi có lỗi lặp, chọn biện pháp nhỏ nhất đúng tầng: lỗi format dùng formatter; vi phạm import dùng lint khi đủ kiến trúc; hiểu sai nghiệp vụ sửa nguồn/tiêu chí/test; mất việc sau context reset sửa bàn giao. Không mặc định giải pháp nào cũng là thêm một skill hoặc thêm agent.

## 7. Vòng làm việc đề xuất

```text
Đọc entry point + nguồn đúng miền + tiến độ
  → Chọn một thay đổi có phạm vi và tiêu chí đạt
  → Kiểm tra môi trường đang hoạt động
  → Implement trong phạm vi
  → Check phù hợp + quan sát luồng chính
  → Sửa lỗi có bằng chứng
  → Ghi trạng thái xác minh + bàn giao cho lượt sau
```

Con người giữ quyết định nghiệp vụ/phạm vi và chấp nhận giao diện. Agent thực hiện phần đã rõ, báo chỗ chưa rõ trước khi chọn cấu trúc mới. Không phải mỗi mũi tên là một agent riêng hoặc một phiên model riêng.

## 8. Chi phí/context: nên tối ưu thứ gì

Đề xuất quản lý **chi phí cho một kết quả được chấp nhận**, không chỉ số token một lượt:

- Entry point trỏ đúng file, không nạp toàn bộ lịch sử DB hoặc toàn bộ bộ Figma mỗi lần.
- Không tạo bản sao nghiệp vụ để các agent duy trì lại.
- Chạy check theo phần đổi, có giới hạn cho vòng thử–sửa; gặp cùng blocker không tiến triển thì báo nguyên nhân thay vì tiếp tục tiêu token.
- Chỉ dùng reviewer độc lập ở thay đổi đáng rủi ro hoặc khi tự review liên tục bỏ sót; không review nhiều tầng mọi nút/icon.
- Giữ tiêu chí trước code để tránh nhiều lượt sửa vì hiểu sai, nhưng không bắt user duyệt từng dòng implementation.
- Context reset chỉ khi có lý do và có bàn giao; không tuyên bố clear context luôn tiết kiệm. Không tự hứa cached context làm mọi phiên miễn phí.

Mức giới hạn lượt chạy/thời gian/chi phí cụ thể cần chọn theo công cụ và cách người dùng sử dụng. Chưa cấu hình automation hoặc mua thêm dịch vụ.

## 9. Cách đo harness của mình có hữu ích không

Thử trên vài công việc frontend thật, không benchmark giả hàng trăm tác vụ:

- Agent có đọc đúng quyết định hiện hành, tránh khôi phục luật cũ?
- Tính năng được chấp nhận sau bao nhiêu lượt sửa có ý nghĩa?
- Phiên sau có tiếp tục được từ repo mà không cần dán lại lịch sử chat?
- Kết quả check có phân biệt mock/API và chưa chạy/đã chạy?
- Harness có tốn công bảo trì hơn lỗi nó ngăn được không?

Ghi baseline rồi so sau vài thay đổi tương tự; kết quả ban đầu chỉ là tín hiệu thực hành, không chứng minh nhân quả hay phần trăm cải thiện phổ quát. Không cho agent sửa/xóa tiêu chí chỉ để có màu xanh; thay tiêu chí hợp lý phải được giải thích theo yêu cầu thực tế.

## 10. Lộ trình áp dụng sau khi được đồng ý

1. **Chốt harness tối thiểu:** entry point, nguồn chuẩn cần theo Git, mẫu công việc/bàn giao, điều kiện phải hỏi người dùng. Giữ .claude/docs làm chuẩn nghiệp vụ.
2. **Khởi tạo web và check cùng nhau:** nền React theo stack đã chốt, token/shell/router/API-mock boundary, lệnh thực có. Không tách một dự án xây harness kéo dài khỏi mục tiêu frontend.
3. **Thử trên một phần nhỏ:** shell/điều hướng và một luồng chính; xem agent có giữ nguồn và bằng chứng khi sang phiên mới.
4. **Bổ sung theo lỗi quan sát:** chỉ thêm check, rule hoặc hook có tác dụng rõ. Multi-agent/SDK riêng để sau, khi chứng minh được nhu cầu.

Chưa thực hiện bốn bước này. Báo cáo không thay ADR, schema, danh mục UC hoặc chính sách quyền của DATN.

## 11. Độ tin cậy và giới hạn nghiên cứu

- Chỉ dùng bài chính chủ, docs chính thức và GitHub của chủ dự án cho kết luận kỹ thuật; không lấy Reddit/tóm tắt SEO làm chứng cứ.
- Các bài nghiên cứu kỹ thuật là kinh nghiệm/experiment của tác giả; không bằng chứng mọi model, task hoặc công ty có cùng hiệu quả. Số liệu chi phí trong ví dụ cũng không phải báo giá cho dự án này.
- Repo main và docs có thể đổi; danh sách nguồn phản ánh ngày đọc. Không kiểm chứng tương thích SDK, toàn bộ mã bảo mật hoặc hiệu suất runtime bằng thực nghiệm.
- Đã đọc mã quickstart, không chạy nó. Trang AGENTS.md của repo Codex không tải được trong lượt đọc; không dựa vào nội dung đó để đề xuất quy tắc repo.
- Skill deep-research có đường Gemini API tính phí riêng; không sử dụng API này. Nghiên cứu thực hiện bằng đọc web trực tiếp, không truyền tài liệu dự án sang dịch vụ research khác.
- Báo cáo nằm trong docs/design nên hiện chịu Git ignore như các tài liệu thiết kế khác. Chưa sửa .gitignore hoặc commit; bước quản lý phiên bản cần thực hiện có chủ đích khi triển khai harness.

## 12. Nguồn đọc tiếp theo thứ tự hữu ích

1. [Anthropic — Effective harnesses for long-running agents](https://www.anthropic.com/engineering/effective-harnesses-for-long-running-agents) — vấn đề nhiều phiên và cách bàn giao.
2. [OpenAI — Harness engineering](https://openai.com/index/harness-engineering/) — repo, entry point và kiểm tra máy.
3. [Anthropic — Harness design for long-running application development](https://www.anthropic.com/engineering/harness-design-long-running-apps) — đánh giá độc lập và giản lược harness.
4. [Anthropic — Effective context engineering](https://www.anthropic.com/engineering/effective-context-engineering-for-ai-agents) — ngữ cảnh đúng lúc.
5. [Claude — Steering Claude Code](https://claude.com/resources/articles/steering-claude-code-skills-hooks-rules-subagents-and-more) — chọn đúng loại chỉ dẫn/cơ chế.
6. [Anthropics — autonomous-coding GitHub](https://github.com/anthropics/claude-quickstarts/tree/main/autonomous-coding) — minh họa thực thi, không phải template DATN.
7. [LangChain — Deep Agents architecture](https://github.com/langchain-ai/deepagents/blob/main/libs/ARCHITECTURE.md) — phân lớp khi thực sự cần xây runtime/harness lập trình được.
