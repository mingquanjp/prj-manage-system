# Đánh giá và cải tiến skill

## Cơ sở nghiên cứu — đọc21/11/2026

- [Anthropic: authoring best practices](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices): giữ chỉ dẫn ngắn, đọc chi tiết theo nhu cầu, kiểm bằng tình huống sử dụng và sửa theo lỗi quan sát được. Áp dụng bằng ba điểm vào đợt mới/tiếp nối/nguồn đổi, không biến mọi việc thành cùng một pipeline.
- [OpenAI: Build skills](https://developers.openai.com/codex/skills/): phạm vi/điều kiện kích hoạt rõ; giữ công cụ deterministic cho kiểm lặp lại. Vị trí repo .agents/skills và metadata hiện có được giữ, không thêm plugin/tool dependency.
- [Agent Skills: evaluating skills](https://agentskills.io/skill-creation/evaluating-skills): dùng prompt thực, so với bản cũ hoặc baseline cùng điều kiện; tách chất lượng output khỏi kiểm cấu trúc. Chỉ kết luận cải thiện có số đo khi đã chạy đối chiếu.
- [Agent Skills: optimizing descriptions](https://agentskills.io/skill-creation/optimizing-descriptions): đánh giá cả prompt nên kích hoạt lẫn prompt gần giống nhưng không thuộc phạm vi. Không mở rộng description để bắt mọi việc trong repo.
- [Agent Skills specification](https://agentskills.io/specification): name/description và tài nguyên liên kết; YAML hợp lệ không đủ chứng minh hành vi.

Đây là nguồn cho cách tạo/kiểm skill, không ghi đè workflow hoặc nghiệp vụ DATN. Không sao chép hạn mức model/token hoặc metadata riêng Claude vào Codex.

## Cách thử

1. Chọn ca trong evals/evals.json đúng thay đổi. Với routing, chỉ đưa name/description và prompt cho evaluator; không đưa should_trigger. Với hành vi, đưa skill/nguồn/prompt, không đưa expected_output/checks hoặc kết luận mong muốn.
2. Cho agent mới làm lượt chỉ đọc; không nhận lịch sử cuộc chat. Chỉ dẫn theo AGENTS nhưng không spawn tiếp, không DB/network/Jira/Git mutation; nếu cần tạo đầu ra thử thì ở /tmp. Không đưa đáp án/rubric để agent chép.
3. Agent chính chấm output theo từng check; lưu đáp án agent và finding ở phạm vi tạm hoặc ghi nhận bàn giao có sẵn. Báo observed/not observed/unverified cho từng tiêu chí, không lấy regex khớp câu chữ làm chất lượng. Routing thủ công chỉ là proxy cho host discovery, không chứng minh Codex tự kích hoạt trên mọi máy.
4. Khi cần kết luận bản mới tốt hơn, chạy cùng prompt/source/model/điều kiện với bản cũ và mới, giữ cả output rồi đánh giá không lộ nhãn phiên bản nếu có thể. Không tắt chỉ dẫn repo để tạo baseline giả, không so prompt dễ/khó khác nhau. Một lần chạy chỉ là smoke test, không là độ tin cậy thống kê hoặc tiết kiệm token.
5. Sửa đúng lỗi quan sát, thử lại ca lỗi và ca gần phạm vi. Mở rộng eval khi xuất hiện lỗi mới; không tạo framework/chạy nhiều model chỉ để có điểm số.

## Phân loại bằng chứng

| Lớp | Chứng minh được | Chưa chứng minh |
|---|---|---|
| YAML/frontmatter | Format/name/description hợp lệ | Chọn skill đúng |
| Node checker tests | Hàm kiểm cấu trúc phát hiện các lỗi fixture | Agent làm đúng nghiệp vụ |
| Routing eval | Đánh giá lựa chọn theo metadata ở các prompt đã thử | Host tự load đúng trên mọi phiên |
| Agent dùng skill | Hành vi/đầu ra quan sát của ca đã chạy | API/DB/browser runtime hoặc ưu thế so baseline chưa chạy |

Đợt nghiên cứu bổ sung dùng bộ ba ca hành vi/tám prompt routing ở repo. Kết quả thực và phần chưa chạy ghi khi bàn giao; không điền trước pass hoặc coi sub-agent đọc file là đã chạy eval.

## Kết quả quan sát 21/11/2026

- Một sub-agent chỉ nhận metadata và tám prompt, không nhận should_trigger: chọn đúng ba ca lập kế hoạch/tiếp nối/review và loại năm ca code-ticket/server/Git/UI nhỏ/hỏi kiến thức. Đây là kiểm routing bằng evaluator, chưa kiểm cơ chế tự chọn skill của host.
- Một sub-agent mới dùng skill với ba prompt, không đọc rubric/evals: tiếp nối PCA trước importer, phân biệt64KiB PCA với8KiB auth/OAP, và brief demo nhỏ dùng lại contract/SQL thay vì tạo cả bộ artifacts. Agent chính đối chiếu output với checks: các hành vi kỳ vọng được quan sát, không có hành động ghi ngoài phạm vi. Đây là lượt thử chỉ đọc, không chứng minh agent thực thi đúng mọi giai đoạn soạn proposal mới.
- Checker PCA đạt10 operation planned/12 WBS/12 task/17 cạnh/6 Mermaid; kiểm contract auth/OAP/PCA27 tests do sub-agent chạy đạt. Lượt tổng hợp của agent chính60 tests đạt, không skip; harness410 nguồn và diff check đạt. YAML/frontmatter/interface hợp lệ qua Ruby Psych; validator Python vẫn thiếu PyYAML như ghi nhận trước, không cài package mới.
- Chưa chạy A/B với bản cũ hoặc không skill, chưa đo thời gian/token, chưa lặp nhiều lượt hoặc model khác; không công bố tỷ lệ cải thiện hay tiết kiệm. Bộ đánh giá để tiếp tục dùng và mở rộng theo lỗi thực, không là chứng nhận skill hoàn hảo.
