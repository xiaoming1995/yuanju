## Why

The compatibility product already supports relationship context and a decision-first result page, but the core user question is still sharper than "what should I do next?": users primarily want to know whether two personalities fit, where they naturally match, and where they will repeatedly clash. The next UX opportunity is to make personality compatibility the primary reading path, then use action guidance as a way to validate that personality judgment.

## What Changes

- Reframe the compatibility entry flow around "性格合不合" as the default mental model before the birth-data forms.
- Add a compact personality-fit preview so users understand that the reading will explain each person's relationship personality, fit points, conflict points, and interaction advice.
- Promote a result-page "性格相处画像" before generic scores: overall personality fit, self pattern, partner pattern, natural match points, repeated conflict points, and communication style.
- Keep the actionable relationship plan, but make it validate the personality judgment through short-term checks such as 7-day and 30-day observations.
- Improve the compatibility history page into a relationship archive that surfaces personality match type, last question, and available continuation actions.
- Preserve existing scores, professional evidence, AI report generation, and backend compatibility algorithms.
- Avoid exact-date predictions, deterministic relationship outcomes, or advice that reads as binding psychological/legal guidance.

## Capabilities

### New Capabilities
- `compatibility-personality-fit-ux`: Covers the personality-fit result experience: relationship personality patterns, fit/clash points, and communication guidance.
- `compatibility-consultation-entry-flow`: Covers a personality-first compatibility creation experience that frames the reading around "性格合不合" before data entry.
- `compatibility-action-plan-ux`: Covers action plans that validate personality-fit judgments through observation windows and do/avoid guidance.
- `compatibility-relationship-archive-continuation`: Covers history/archive affordances that surface personality match type and help users continue from prior compatibility readings.

### Modified Capabilities
- None.

## Impact

- Frontend compatibility input page and CSS: `frontend/src/pages/CompatibilityPage.tsx`, `frontend/src/pages/CompatibilityPage.css`
- Frontend compatibility result page and CSS: `frontend/src/pages/CompatibilityResultPage.tsx`, `frontend/src/pages/CompatibilityResultPage.css`
- Frontend compatibility history page and CSS: `frontend/src/pages/CompatibilityHistoryPage.tsx`, `frontend/src/pages/CompatibilityHistoryPage.css`
- Possible helper logic under `frontend/src/lib/` to derive personality-fit and action-plan copy from existing context, dimension scores, evidence, decision advice, stage risks, and duration assessment
- Static/frontend UX tests under `frontend/tests/`
- No expected database, API, authentication, or compatibility scoring changes
