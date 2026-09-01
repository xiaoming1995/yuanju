## Why

The Bazi result page already gives users a plain-language natal vehicle grade and current Dayun road condition, but it does not isolate wealth-specific structure. Users who care about money-related readings need a concise first-screen answer with professional evidence available on demand, without turning the chart into a claim about real-world assets or investment outcomes.

## What Changes

- Add a deterministic `wealth_profile` to Bazi results that grades natal wealth structure from S to D and exposes a summary, tags, risk flags, and evidence.
- Keep the primary wealth grade scoped to the natal chart only; current Dayun wealth activation is shown as a separate window hint, not mixed into the natal grade.
- Evaluate wealth through visible/rooted wealth stars, day-master carrying capacity, favorable/adverse Ten God direction, wealth-producing chains, pattern support, and wealth-loss risks such as weak day master carrying too much wealth or peer competition.
- Surface "财富结构" on the result overview in ordinary-user language, with an evidence entry point for professional mode.
- Inject the backend-computed wealth profile into AI report prompts so AI explains the calculated result without re-grading or contradicting it.
- Preserve compatibility for older saved chart snapshots by lazily backfilling missing or outdated wealth profile data from the stored Bazi result.

## Capabilities

### New Capabilities
- `bazi-wealth-profile`: Deterministic natal wealth-structure grading, evidence, risk flags, and current Dayun wealth-window hints for Bazi results.

### Modified Capabilities
- `bazi-advanced-data`: Bazi calculate and history-detail results include `wealth_profile` when deterministic inputs are available.
- `bazi-ai-reasoning`: AI report prompts include backend-computed wealth-profile context and prohibit AI from inventing or changing the wealth grade.

## Impact

- Backend: `backend/pkg/bazi` result types and deterministic wealth-profile builder, snapshot lazy backfill in `backend/internal/service/report_service.go`, and focused Go tests.
- Frontend: `frontend/src/pages/ResultPage.tsx`, result styling, frontend result typing, and static/render tests for the new overview and evidence display.
- API contract: additive optional `wealth_profile` field on Bazi result payloads; no database migration is expected because full results are already stored in `result_json`.
- UX: result overview expands from natal vehicle/current road toward a three-part first-screen summary; responsive layout must remain readable alongside the active Dayun phase-road change.
