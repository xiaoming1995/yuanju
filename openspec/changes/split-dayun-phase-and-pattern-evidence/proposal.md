## Why

The Dayun road evaluator currently totals the Dayun stem and branch into one ten-year delta. For the 1995-10-12 noon chart, 壬水 contributes support while 午火 contributes resistance, so the UI only shows “中性 0” and obscures the distinct front-five and back-five conditions. The same cancellation affects Ten God evidence, while the pattern interaction layer cannot distinguish a verified 七杀生偏印 chain from a generic unscored occurrence.

## What Changes

- Split Dayun five-element and Ten God evidence into stem-led front-five and branch-led back-five contributions, while retaining a ten-year aggregate for road classification.
- Return structured phase evidence so the frontend can show both the individual contribution and any net cancellation instead of presenting a bare neutral score.
- Add a narrowly gated 七杀生偏印 / 杀印相生 pattern-interaction rule for a 偏印格 only when the required chain is present and verifiable.
- Keep non-matching pattern relationships neutral; do not infer a score merely because a Dayun contains 七杀 or a natal chart has 偏印.
- Update professional evidence presentation and regression coverage for phase split, aggregate score, and pattern-rule boundaries.

## Capabilities

### New Capabilities
- `dayun-phase-evidence`: Exposes phase-specific Dayun road evidence and verified pattern interactions without hiding opposing stem and branch effects inside one neutral total.

### Modified Capabilities
- `dayun-heibiao-engine`: Preserve the existing front-five stem and back-five branch interpretation as the phase boundary used by the road-evidence model.

## Impact

- Backend: `backend/pkg/bazi/vehicle_profile.go`, Dayun road DTOs, calculation tests, and JSON serialization.
- Frontend: result-page professional road evidence rendering and its TypeScript types/tests.
- No database migration, route change, or removal of existing `dayun_roadmap` fields; new phase evidence is additive.
