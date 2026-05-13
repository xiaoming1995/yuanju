## Why

The past-events timeline currently evaluates every year in a 10-year dayun with the full dayun gan-zhi as one undifferentiated background. This misses a common professional reading rule: the first five years of a dayun are more led by the dayun heavenly stem, while the latter five years are more led by the dayun earthly branch.

Adding this phase awareness will make yearly narratives within the same dayun more accurate and help users understand why the tone changes between the early and late half of a decade.

## What Changes

- Add dayun phase awareness to past-events yearly signal generation.
- Determine each liunian's position within its containing dayun and classify it as:
  - `gan` phase for years 1-5 of the dayun.
  - `zhi` phase for years 6-10 of the dayun.
- Use the phase to weight dayun influence:
  - In `gan` phase, dayun heavenly-stem interactions and ten-god themes should be stronger narrative context.
  - In `zhi` phase, dayun earthly-branch interactions, clashes, combinations, empty branch, and JinBuHuan branch rating should be stronger narrative context.
- Reuse existing `jin_bu_huan.qian_*` and `jin_bu_huan.hou_*` results as phase-level background signals.
- Keep liunian's own gan-zhi and direct natal interactions active in all years; phase weighting modifies emphasis, it does not suppress the yearly trigger.
- Add regression tests proving years 1-5 and 6-10 in the same dayun can receive different dominant context when their dayun phase differs.

No breaking API change is required. Additive diagnostic fields may be returned if useful for frontend display or debugging.

## Capabilities

### New Capabilities

- `past-events-dayun-phase-weighting`: Defines how past-events yearly inference incorporates the dayun first-half heavenly-stem phase and second-half earthly-branch phase.

### Modified Capabilities

- None.

## Impact

- Backend Bazi signal engine:
  - `backend/pkg/bazi/event_signals.go`
  - `backend/pkg/bazi/event_narrative.go`
  - related tests under `backend/pkg/bazi`
- Backend past-events service:
  - `backend/internal/service/report_service.go`
- Optional additive frontend/API diagnostics:
  - `frontend/src/pages/PastEventsPage.tsx`
- Existing `jin_bu_huan` calculation remains unchanged and becomes an input to yearly past-events context.
