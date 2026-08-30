## Why

The natal chart for 1996-02-08 20:00 (`丙子·庚寅·乙亥·丙戌`) currently identifies the day-stem regulation evidence, but classifies the 寅-month climate as neutral and leaves 扶抑用神 unset. This contradicts the product rule that early-spring chill requires fire for thermal regulation and that the same fire can also be a 扶抑用神 when the day master is seasonally strong.

## What Changes

- Add an explicit early-spring (`寅`) thermal-regulation condition with `火` as its required element, while keeping thermal regulation distinct from day-stem regulation.
- Add month-command influence to the global 扶抑 strength assessment so a day master supported by its month command is not incorrectly treated as neutral through pillar-score cancellation.
- Derive and expose a shared-priority yongshen when an element appears in both thermal-regulation and 扶抑 yongshen results; retain the full, separate results for each layer.
- Surface the shared-priority explanation in the natal assessment, vehicle profile, and AI interpretation context without adding a duplicate grade bonus.
- Version the changed natal assessment result so saved or cached outputs are recalculated under the revised rule.

## Capabilities

### New Capabilities
- `seasonal-fuyi-yongshen-alignment`: Align early-spring thermal regulation, month-command 扶抑 assessment, and shared-priority yongshen output.

### Modified Capabilities
- `bazi-advanced-data`: Return the structured shared-priority yongshen alignment with advanced Bazi calculation data.
- `yongshen-driven-flow-year-judgment`: Provide the shared-priority yongshen context to AI dayun and report interpretation flows.

## Impact

- Affects `backend/pkg/bazi` thermal-regulation, 扶抑, natal-assessment, vehicle-profile, and result-version logic.
- Affects API response fields and frontend natal-assessment display for recalculated charts.
- Affects report and dayun prompt construction; no new third-party dependency or database migration is expected.
