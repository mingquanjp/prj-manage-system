# Prompt cập nhật DATN theo từng đợt

> LỊCH SỬ — bộ prompt phục vụ Figma Make trước khi chuyển sang frontend local. Không chạy lại hoặc dùng thay kế hoạch phát triển hiện tại. Xem [quy trình local](../../../web/README.md) và [mục lục nguồn](../../README.md); nội dung từng đợt dưới đây là hướng dẫn tại thời điểm soạn, không chứng minh đã implement/kiểm thử.

Ngày 20/10/2026. Thay toàn bộ prompt cũ tại đường dẫn này. Dùng trên [file Make hiện có](https://www.figma.com/make/fOXzi51sKaRQbuSKmHZQmD/DATN); giữ shell và component phù hợp. Nạp lại Guidelines.md, ui-brief.md, screen-contracts.md. Các tài liệu trong src/imports và plan cũ không được ghi đè nghiệp vụ hiện hành.

Mỗi đợt: xem code hiện có → sửa phạm vi đợt → chạy preview → đối chiếu checklist → ghi kết quả. Không yêu cầu tạo lại ứng dụng. Nếu thiếu màn hình để chứng minh quyền, bổ sung tối thiểu theo ma trận. Không gọi chức năng chưa dựng là đã hoàn tất.

## Đợt 1 — Quyền và thiết lập dự án

Đính kèm thêm roles_permissions.md, project_task_assignment_workflow.md và domain_rules.md từ .claude/docs.

```text
Update the existing DATN Figma Make prototype to the attached business baseline dated 20 October 2026. Read the updated Guidelines, ui-brief and screen-contracts first. Replace the authority of old imported documents and old plans; preserve existing visual tokens and reusable components.

Inspect App.tsx, AppShell, project types, project lists/details/structure, CreateProjectModal and mock data. Introduce explicit current-user system roles and project-specific PM/PTCM responsibilities. A person may combine responsibilities; PTCM is not a global role. A demo identity selector must be labeled as a prototype control and must not represent a production self-permission feature.

Only Director creates a project, selecting a PM in the same form/save. The creator is not automatically PM. PM appoints PTCM, monitors and starts/completes/closes/cancels according to the attached state rules. Only that project's PTCM creates groups/tasks and assigns/reviews work. A PM without PTCM responsibility has no professional actions. A PTCM without PM responsibility cannot start/complete/close the project.

Prevent appointing a PTCM who still has open assigned tasks in that project or unresolved conflict with their own earlier output. Do not silently revoke those tasks. PTCM cannot execute tasks in their own project. Keep the employee personal-work view for other eligible projects.

Project creation requires code/name/PM/start/end/priority. Task creation uses description including expected output and separate acceptance criteria, positive estimate and at least one skill requirement. Remove a separate deliverable field and catalog ACTIVE/INACTIVE assumptions. Preserve the four proficiency levels. Requirement weights belong to PTCM per task; the final numeric domain is not settled: retain any necessary demo values explicitly as provisional, without asserting a fixed 1–5 business constraint.

Do not add a roster, plan approval, sprint, separate reviewer or project transfer workflow. Keep all Vietnamese UI copy, shared mock state and validated interactions. At the end report files changed and preview checks T01–T07, T24–T26, T34; mark unbuilt cases honestly.
```

## Đợt 2 — Phân công và tra cứu

Đính kèm ai_recommendation_workflow.md, project_task_assignment_workflow.md, roles_permissions.md.

```text
Update only assignment and professional personnel lookup in the existing prototype using the current brief/matrix. Inspect RecommendationResults, DraftReview, ConfirmDialog, ManualDrawer, recommendation types/data and permission guards.

The project's PTCM selects eligible PENDING_ASSIGNMENT tasks across the filtered result including unseen pages, requests recommendations once for that set, and selects a candidate per task. Each task has its own immutable run and 0–3 candidates. Do not preselect tasks or Top-1. Use eligible staff company-wide, no project roster or two-project cap. Exclude PTCM of this project, locked users and inactive employees. Recommended mandatory skills must be met. Manual may override only skill gaps with a reason.

Retain rank/time/explanation for recommendations and clearly mark numeric scores as prototype examples pending the final formula. Manual selections have no rank/score. Non-Top-1 and switching to manual after a run require reasons. Remove every obsolete R05 history-warning acknowledgment, history revision and old-snapshot acceptance reason from state, dialogs and confirm validation. Keep snapshot time and rerun. Current skills/spec/permissions/status must still be checked at confirm. Current workload updates do not rewrite old ranking.

Draft workload per employee equals current open workload plus the sum of estimates for all currently selected draft tasks for that employee. Draft does not change official workload. Failed confirmation changes none of the batch; success updates every assignee/status/pending count and workload used by the next run in the same preview session.

Direct replacement is manual and atomic, retaining the previous assignee on failure. Suggested replacement requires revoking first, leaving PENDING_ASSIGNMENT if no new candidate is confirmed. PENDING_REVIEW must be returned before replacement/revocation. Preserve frozen task scope and original contribution history.

Add or adjust independent professional lookup for PM/PTCM with only permitted work identity, official skills, permitted experience/results and aggregate workload. No other person's contact/raw CV or cross-project task/log detail. Test T08–T15, T31, T33; preserve the visual system.
```

## Đợt 3 — Thực hiện, nghiệm thu và hiệu chỉnh

Đính kèm work_effort_workflow.md, project_task_assignment_workflow.md, domain_rules.md.

```text
Update EmployeeApp, Kanban, EmployeeTaskDetail, review queue and correction views to the current workflow. Keep one shared mock source for employee, PTCM and PM views.

Employee starts/resumes only while the project is IN_PROGRESS. State actions do not generate hours; Kanban drag-and-drop does not perform business transitions. Logs record date, positive integer minutes and description; when direct editing is allowed, actual work can be added even if Start/Resume was forgotten. More than 480 minutes across all tasks on a date requires a reason; more than 1440 is blocked. Keep assignment/project-period and future-date checks.

Submission requires a nonempty result summary, with optional evidence links. Missing logs do not block submission or acceptance. Remove effort_complete, incomplete declaration/reason gates and separate quality-confirmation workflows. Empty logs display “Chưa có thời gian ghi nhận”, never a fabricated zero-minute log. Submission goes to PENDING_REVIEW with an immutable log snapshot. No proxy submission.

PTCM of the project reviews against frozen acceptance criteria. Accept goes DONE; request changes requires a reason and returns IN_PROGRESS for employee revision then resubmission. Only a log error does not require redoing valid output. No separate reviewer, PM review permission or reopening DONE.

When direct log edit is allowed, edits/deletions require reasons/history without approval. Locked logs and prior assignments use self-submitted ADD/UPDATE/DELETE requests approved by current PTCM. Pending does not affect totals; withdraw only pending, terminal immutable; soft-delete without restore. COMPLETED permits valid corrections; CLOSED/CANCELLED are read-only. A canceled pending submission must not be shown as a failed quality review.

PM starts/completes/closes by project conditions, not from one task's acceptance. Show errors and empty/pending/readonly states. Test T16–T25 and T33; report unimplemented cases.
```

## Đợt 4 — Tài khoản, hồ sơ và danh mục

Đính kèm roles_permissions.md, employee_profile_workflow.md, quyết định 20/10 và phần dictionary cần thiết.

```text
Implement/update the supporting account, personnel and catalog views according to the attached matrix UI-01–UI-04, UI-08–UI-09 and the 20 October decision.

Admin creates account plus initial employee profile in one save: username/password according to the still-open credential policy, required employee code/fullname/company email/position, optional phone/department/seniority. The initial creation right does not grant ongoing HR editing. HR owns administrative updates afterwards. System roles may combine; PM/PTCM are assigned per project, not account roles. Do not add must_change_password or a fabricated password-delivery/reset policy.

Employees request fullname/phone/address changes; HR approves/rejects and directly manages company email/organization fields. Pending requests do not update the official profile. Reject needs a reason; stale base values must not overwrite current data. No extra profile-request withdraw or pending-evidence flow.

Approved address rule: address must be supplied when completing the personnel profile, but is not required when Admin creates the account and initial profile. NULL means not yet supplied. Show “Chưa cập nhật” with a reminder; do not block login, assignment or task execution. Employees propose address changes through the existing HR approval flow; HR can also update it directly. Omit unchanged address from the patch; any supplied/proposed replacement must contain non-whitespace text. Do not allow clearing an existing address to NULL or an empty value. Do not require address on requests that only change fullname/phone. Keep address out of professional lookup, candidates and recommendation snapshots; add no profile-completion status or extra workflow. Default personal information is read-only; open the request form only when the user selects “Đề nghị cập nhật”.

HR manages departments and positions. Director manages skill categories and skills. These can be two compact catalog pages with tabs, not extra product modules. No skill INACTIVE status, hierarchy or automatic proficiency from job seniority. Deletion is blocked while referenced; preserve history. HR deactivation is separate from Admin account locking and is blocked by open work/responsibilities.

Only render fields permitted to the current person. Preserve the existing shell and components. Check T26–T27, T32, T34; explicitly list unresolved credential details instead of inventing them.
```

## Đợt 5 — CV và kỹ năng

Đính kèm cv_extraction_approval_workflow.md, skill_approval_workflow.md, employee_profile_workflow.md.

```text
Update/add personal CV, HR CV management and Director skill-review screens using matrix UI-05–UI-07. Keep CV file selection, extraction processing and skill approval visibly separate.

Owner/HR upload PDF-with-text or DOCX up to 10 MiB. Only the employee selects their current CV; an extraction failure does not block using a valid file. There is no approve/reject CV-publication button. PM/PTCM cannot download another person's CV from professional lookup.

Extraction proposes skill names and supporting text, not a level/years/confidence score. Director maps skills and establishes a justified level before official approval. Manual skill requests work without a new CV or internal task history. Provide explanation, optional multiple internal contributions, at most one directly referenced CV and optional separate evidence files. Evidence-file limits are unresolved; do not reuse CV limits or invent accepted types/counts. Keep per-item decisions and the derived aggregate skill-review outcome consistent, never label the CV file approved/rejected.

A skill request allows pending evidence additions and withdrawal; at most one pending per employee/skill across MANUAL/CV. Final decisions are immutable; new evidence after closure uses a new request. Director approval and rejection require reasons and cannot be self-approval. HR cannot approve skills. Official skills remain distinct from pending/extracted skills; REMOVE removes current recognition while preserving decisions/history. NO_CHANGE is not approval.

When specialist advice is needed, record its source/content as evidence without creating an expert role, mandatory extra approval level or in-app examination. Test T28–T31 and report unresolved late duplicate mapping.
```

## Kết thúc từng đợt

Yêu cầu Make báo: màn hình/file đã thay, tình huống đã thao tác preview, lỗi còn lại và phần chưa dựng. Dùng [checklist](review-checklist.md); không đánh dấu pass từ việc chỉ đọc code. Hoàn tất bộ prompt không đồng nghĩa đã chạy các đợt trên file Make.
