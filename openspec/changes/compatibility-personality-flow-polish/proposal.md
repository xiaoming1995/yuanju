## Why

The compatibility flow now answers the core question of whether two personalities fit, but the page hierarchy still feels dense: the result page has several adjacent validation and evidence sections, while the entry and history pages still read partly like forms and score archives. This polish pass should reduce cognitive load, make the personality-fit path easier to scan, and improve mobile usability without changing the algorithm or API.

## What Changes

- Refine the compatibility entry page into a lightweight three-step consultation path: question, relationship stage, and birth profiles.
- Add compact progress cues so users know what has been selected and what remains before starting the reading.
- Consolidate overlapping result-page validation content so "性格相处画像" and action validation feel like one coherent reading path instead of multiple repeated sections.
- Add top-of-result anchor navigation for fast jumps to personality fit, conflict points, action validation, and professional evidence.
- Make personality match types more self-explanatory with a short subtitle or explanation under labels such as "高吸引高消耗型".
- Reduce card density and nested-card feeling in the result page, especially inside personality fit and validation sections.
- Rebalance compatibility history cards so personality match type and continuation action are visually primary, while raw scores become secondary.
- Keep current backend APIs, compatibility scoring, AI report generation, and professional evidence behavior unchanged.

## Capabilities

### New Capabilities
- `compatibility-entry-stepper-polish`: Covers the compatibility creation page as a compact consultation-style flow with progress cues and mobile-safe action placement.
- `compatibility-result-hierarchy-polish`: Covers result-page hierarchy, anchor navigation, merged validation structure, clearer match-type explanations, and lower-density layout.
- `compatibility-history-scan-polish`: Covers history/archive cards that prioritize personality match recognition and continuation over raw score scanning.

### Modified Capabilities
- None.

## Impact

- Frontend compatibility entry page and CSS: `frontend/src/pages/CompatibilityPage.tsx`, `frontend/src/pages/CompatibilityPage.css`
- Frontend compatibility result page and CSS: `frontend/src/pages/CompatibilityResultPage.tsx`, `frontend/src/pages/CompatibilityResultPage.css`
- Frontend compatibility history page and CSS: `frontend/src/pages/CompatibilityHistoryPage.tsx`, `frontend/src/pages/CompatibilityHistoryPage.css`
- Existing personality helper may need small copy/metadata additions under `frontend/src/lib/compatibilityPersonality.ts`
- Static UX tests under `frontend/tests/`
- No expected database, backend API, authentication, dependency, or scoring-model changes
