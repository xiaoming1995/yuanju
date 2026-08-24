## 1. State Model And Test Fixtures

- [x] 1.1 Inspect current `PastEventsPage` data flow and identify the minimum state needed for generated, generating, interrupted, future folded, and future expanded states.
- [x] 1.2 Add or update frontend test fixtures for deterministic years, cached dayun summaries, active stream generation, interrupted summaries, future folded segments, and `夹拱` signals.
- [x] 1.3 Introduce a small helper or component-level mapping that derives a single render state for each dayun segment from existing summary data.

## 2. Page-Level Status And Current Period Focus

- [x] 2.1 Add a top-level past-events status summary that separates deterministic year readiness from AI dayun commentary progress.
- [x] 2.2 Show generated/cached, generating, interrupted, and future-not-generated counts or equivalent concise state cues.
- [x] 2.3 Identify and highlight the current dayun segment when present.
- [x] 2.4 Identify and highlight the current year card when present.
- [x] 2.5 Add a top-level cue or navigation affordance that helps the user reach the current dayun without scanning the whole timeline.

## 3. Dayun Segment Interaction

- [x] 3.1 Replace ambiguous future folded copy with a clear reveal action for deterministic year signals.
- [x] 3.2 Ensure revealing a future dayun shows year signals without calling the AI stream endpoint.
- [x] 3.3 Show a separate `生成本段 AI 批语` action only for expanded future dayun segments without cached/generated AI content.
- [x] 3.4 Keep per-segment generating, generated, cached, and interrupted labels consistent with the derived segment state.
- [x] 3.5 Ensure retry actions target only the interrupted segment and do not disturb completed segments.

## 4. Year Card Readability And Signal Chips

- [x] 4.1 Rebalance year cards so readable narrative and event chips appear before dense technical evidence.
- [x] 4.2 Keep ten-god power, dayun phase, and evidence summary visible as secondary detail without overlap or clipping.
- [x] 4.3 Add visible chip support for `夹拱` past-events signals.
- [x] 4.4 Add a conservative fallback chip for meaningful unmapped signal types.

## 5. Component Structure And Styling

- [x] 5.1 Extract focused past-events UI pieces where useful, such as header/status, dayun segment, and year card components.
- [x] 5.2 Move repeated inline styling into maintainable CSS or local style helpers consistent with the existing frontend style system.
- [x] 5.3 Verify the layout remains readable on desktop and mobile widths.

## 6. Verification

- [x] 6.1 Add or update frontend tests for page-level status, current dayun/year focus, future two-step reveal, retryable interruption, and `夹拱` chip rendering.
- [x] 6.2 Run the relevant frontend test suite.
- [x] 6.3 Run frontend build or type-check verification.
- [x] 6.4 Manually verify the past-events page flow with at least one chart containing past/current dayun summaries and one future folded segment.
