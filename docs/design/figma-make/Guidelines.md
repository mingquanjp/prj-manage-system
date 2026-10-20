# Figma Make guidelines — UI/UX baseline

> **Hướng dẫn hiện hành 20/10/2026.** Dùng cùng [brief](ui-brief.md), [ma trận màn hình/quyền](screen-contracts.md) và [checklist](review-checklist.md). Thay bản Guidelines cũ trong Make; phần phong cách được giữ, nghiệp vụ theo workflow hiện hành trong `.claude/docs/`.


Use these guidelines for every screen in this Make file. The product is a Vietnamese, desktop-first enterprise web application for internal project/task assignment. Preserve the same visual language and reusable components across roles. Do not invent business features, metrics, permissions or state transitions. Use the attached UI brief and current workflow specifications for business behavior; if a detail is missing, mark it as a question in the plan instead of guessing.

## Language and content

- All visible UI copy is natural Vietnamese. Use consistent terms: Dự án, Nhóm công việc, Công việc, Phân công theo gợi ý, Phân công thủ công, Thu hồi phân công, Thay người phụ trách, Nhật ký thời gian, Nộp kết quả, Nghiệm thu.
- Use credible short sample data and realistic dates/statuses. Mock recommendation scores required by the workflow are allowed and must be labeled as prototype data; do not invent dashboard analytics, productivity KPIs, celebratory slogans, or AI claims.
- Explain actions with concise contextual text, not paragraphs in every card. Keep technical IDs, formula internals and database fields out of primary screens.
- Keep internal state codes in mock data/logic, not in visible labels. In the first flow, display `PLANNING` as “Lập kế hoạch”, project `IN_PROGRESS` as “Đang thực hiện”, and task `PENDING_ASSIGNMENT` as “Chờ phân công”. Do not create new states to fill gaps.

## Visual system

- Aim for a calm, clean work surface: borrow the restraint, clear hierarchy and quiet controls of Claude's public interface, not its branding, dark sign-in page, serif display typography or marketing layout. This is a data-dense enterprise application, so keep a light theme. Do not imitate Claude's logo or accent color.
- Use shared tokens across every screen: canvas `#F7F6F3`, sidebar/top bar `#FBFAF8`, surface `#FFFFFF`, primary text `#26231F`, secondary text `#625E58`, border `#E5E1DB`, primary action `#26231F` with white text (hover `#3B3834`), subtle selected/hover surface `#F1EFEB`, focus ring `#2563EB`. Keep the focus blue for accessibility, not as a second decorative accent. Status pairs (text/background): error `#B91C1C/#FEF2F2`, warning `#92400E/#FFFBEB`, success `#166534/#F0FDF4`. Status colors communicate meaning only; never use them as general decoration. Verify readable contrast in the actual preview; adjust a token globally if necessary, never only on one screen. No gradients, glows, decorative illustrations, giant statistic cards or oversized empty space.
- Desktop reference viewport 1440 px; persistent sidebar 224 px, top bar 56 px, content width up to 1200 px with 24 px page padding. Shared spacing scale: 4/8/12/16/24/32 px. Use consistent 8 px corners and 1 px borders. Prefer border and spacing over shadows; reserve one restrained shadow style for overlays or genuinely elevated surfaces. Keep tables compact enough for real work without crowding rows or leaving large unused gaps.
- Use `Inter` with a sans-serif fallback and Vietnamese glyph support. Type scale: page title 24 px/semibold, section title 20 px/semibold, body 16 px/regular, dense table/secondary text 14 px/regular. Essential data must not be smaller than 14 px. Use comfortable line height, restrained font weights, and tabular numerals for dates, counts, hours and scores. Do not use decorative serif text or unusually tight letter spacing in application screens.
- Prefer tables for work lists, candidate comparison and draft review. Use drawers or focused dialogs for details/actions; avoid nested modal chains. Make each page have one obvious primary action.
- Use shared button variants: primary filled charcoal with white text; secondary white with a subtle border; tertiary text-only; danger only for destructive actions. Provide distinct default, hover, keyboard-focus and disabled appearances; disabled buttons must remain legible and must not look selected. Keep inputs/buttons 40–44 px high and table rows at least 48 px high. Use one consistent icon style and selected-navigation treatment. Reuse shell, filter bar, data table, status badge, row action menu, notice, confirmation dialog and empty/loading/error states; do not redesign navigation, buttons or status colors per screen.

## Interaction and accessibility

- Every control has a clear label. Icon-only buttons have accessible names. Keyboard focus is visible; dialogs have proper focus behavior; do not encode state by color alone. Use sufficient text/background contrast and comfortable click targets.
- Show validation beside the relevant field or row, and a summary at batch confirmation. Disable unavailable actions with a reason where useful; do not hide important business constraints.
- Make loading, empty, error and success states meaningful. An empty state has one clear next action. Avoid spinners that imply background work has completed.
- Confirmation distinguishes a draft from a committed operation. Never imply a recommendation automatically assigns someone. For atomic batches, never show partial success.
- If subtle motion clarifies local feedback, keep it to transform/opacity, at most 200 ms, and respect reduced motion. Do not animate layout, add decorative page transitions or use motion to conceal a delayed state change.

## Prototype constraints

- This is an interactive prototype with mock data, not a connected backend. Label simulated outcomes as demo behavior where necessary. Do not fabricate API integration, persistence or security guarantees.
- Build one validated flow at a time. Retain existing shell/components and change only the requested area during refinements.
- Navigation and actions follow current system roles plus project-specific responsibilities. PM monitors and starts/completes/closes; PTCM structures, assigns and reviews. Director creates a project with a required PM. HR manages administrative records and organization catalogs; Director manages skills and skill categories. ADMIN creates an account with its initial employee record, with HR owning later administrative updates. ADMIN does not get all-project task details by default; employees see their own current work and historical contribution.
- A person may combine responsibilities. Label any demo identity switch as a prototype control; do not present project PTCM as a global account role. PM/PTCM personnel lookup is limited to professional data and aggregate workload, without other people's contacts/raw CV or cross-project task/log detail.
- Scores are provisional prototype examples until the final formula is approved. Do not hardcode a 1–5 weight rule or an 80-hour alert as a finalized product rule. Keep ranking timestamps and permitted explanations; manual lookup has no rank/score. Remove obsolete R05 acknowledgment gates.
- If Make chooses a React/Tailwind implementation, use its accessible existing primitives consistently, avoid custom keyboard/focus widgets, and show destructive confirmations with accessible dialogs.
