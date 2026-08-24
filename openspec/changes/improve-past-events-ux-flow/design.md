## Context

The past-events page currently uses a two-stage flow:

1. `fetchPastEventsYears` returns deterministic year cards and dayun metadata without AI cost.
2. `streamDayunSummaries` streams cached or newly generated AI dayun summaries for past and current dayun segments, while future segments remain folded until the user explicitly requests AI generation.

This architecture is sound and matches the token-saving progressive-generation strategy. The main issue is presentation: the page mixes deterministic readiness, AI loading, cached content, interrupted streams, retry actions, folded future segments, and technical evidence in the same visual hierarchy. Users can struggle to know what is finished, what is still running, and which action is safe or costly.

## Goals / Non-Goals

**Goals:**

- Make the page state legible at a glance: deterministic years ready, AI summaries generating, generated/cached, interrupted, or future not generated.
- Put the current dayun/current year near the top of the user's attention without removing access to the full timeline.
- Preserve the current cost-saving behavior: initial load auto-generates only past/current dayun summaries; future AI generation remains user-initiated.
- Make the two-step future interaction explicit: reveal year signals first, generate AI commentary second.
- Keep professional evidence available while making default year-card text easier to scan.
- Ensure all meaningful signal types, including `夹拱`, can render as visible chips.

**Non-Goals:**

- Do not change bazi calculation, past-events signal generation, ten-god power scoring, or AI prompt semantics.
- Do not add automatic AI generation for future dayun segments.
- Do not add new backend endpoints or database migrations.
- Do not redesign the entire application visual system.
- Do not add end-user token pricing or billing visibility.

## Decisions

### Decision 1: Use an explicit frontend segment state model

Each dayun segment should be normalized into a small UI state before rendering:

- `algorithm_ready`: deterministic year cards are present but AI summary is not relevant yet.
- `generating`: the segment is waiting for or receiving AI summary stream output.
- `generated`: AI summary and year narratives are available from stream completion or cache.
- `future_folded`: future segment is collapsed and has not requested AI generation.
- `future_expanded_un generated`: future segment is expanded to show deterministic year signals only.
- `interrupted`: AI generation failed or stalled and can be retried.

Rationale: the current state is inferred from several booleans (`loading`, `folded`, `interrupted`, summary presence). A named UI state will make labels, buttons, tests, and rendering decisions less ambiguous.

Alternative considered: keep the current booleans and only change copy. That is lower effort but leaves state combinations easy to regress.

### Decision 2: Add page-level progress without changing backend contracts

The top of the page should summarize:

- current year and current dayun if detectable from loaded metadata,
- number of dayun summaries generated or cached,
- number currently generating,
- number interrupted and retryable,
- future segments that are available for manual expansion/generation.

Rationale: all of this can be derived from existing frontend state and SSE callbacks. No backend contract change is necessary.

Alternative considered: add server-provided aggregate progress. That would add API surface for information the client already has.

### Decision 3: Prioritize current-period navigation in the client

The UI should make current dayun/current year visible through a focus card, anchor, or highlight. The full dayun order can remain chronological, but the current period must be discoverable without manual scanning.

Rationale: the current dayun is the highest-value content on first visit. This improves usability without changing generation order or historical completeness.

Alternative considered: reorder the whole timeline around current dayun. That may confuse professional users who expect chronological order, so highlighting or anchoring is safer.

### Decision 4: Keep future dayun generation consent explicit

Future dayun segments should keep the two-step interaction:

1. Reveal deterministic year signals without calling AI.
2. Generate AI commentary for that segment only after a separate user action.

Button copy should reflect the distinction, for example `展开年份信号` and `生成本段 AI 批语`.

Rationale: this preserves the established token-cost control while reducing confusion.

Alternative considered: generate future AI summaries automatically after expansion. That improves smoothness but violates the existing user-consent cost strategy.

### Decision 5: Treat technical evidence as secondary detail

Year cards should lead with readable narrative and visible event chips. Dense evidence, phase details, and force explanations should remain accessible through expandable detail or lower-emphasis rows.

Rationale: the page serves both normal users and professional users. Default readability should improve without removing professional data.

Alternative considered: remove technical evidence from the page. That would weaken professional auditability and contradict prior evidence-reporting direction.

### Decision 6: Use a visible fallback for unknown meaningful signal labels

Known new signals such as `夹拱` should be added to the explicit label map. If future meaningful signal types are not mapped, the UI should render a readable fallback rather than dropping the chip silently.

Rationale: silent omission makes new algorithm output look absent even when the backend produced it.

Alternative considered: continue ignoring unknown signals. That is cleaner visually but hides useful algorithm evidence.

## Risks / Trade-offs

- More visible state labels may make the page feel busier -> keep labels short, use consistent visual hierarchy, and avoid repeating global status text in every card.
- Highlighting the current period may disrupt chronological reading -> preserve the full timeline order and use highlight/anchor rather than hard reordering unless implementation shows a better path.
- Unknown-signal fallback may expose internal labels -> map known user-facing labels first and keep fallback conservative.
- Component extraction can expand the diff -> keep extraction limited to past-events page pieces that directly support testability and UX clarity.

## Migration Plan

- Implement as frontend-only changes unless a small compatibility fix is discovered.
- Preserve existing routes, API calls, SSE payload shape, and cache behavior.
- Add or update frontend tests for the new page states and signal labels.
- Rollback is limited to reverting the frontend UI changes; no data migration is involved.
