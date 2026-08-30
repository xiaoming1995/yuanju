## 1. Phase-Aware Dayun Road Model

- [x] 1.1 Add additive Dayun phase-evidence data structures and JSON fields while retaining existing aggregate `DayunRoad.Evidences`.
- [x] 1.2 Refactor Dayun five-element evaluation to calculate, label, and retain stem-front and branch-back contributions before aggregation.
- [x] 1.3 Refactor Dayun Ten God evaluation to apply confidence weighting per phase before computing its aggregate delta.
- [x] 1.4 Assign 十二长生 and branch-derived modifier evidence to the back phase, preserve decade-wide evidence in aggregate form, and keep final road score thresholds unchanged.

## 2. Verified Pattern Interaction

- [x] 2.1 Implement the gated 偏印格杀印相生 matcher for a visible, rooted natal 偏印 and a Dayun stem 七杀 with a valid elemental relay.
- [x] 2.2 Prevent duplicate formation scoring and return explicit neutral evidence when the chain is hidden-only, branch-only, mismatched, or otherwise incomplete.
- [x] 2.3 Verify the 1995-10-12 noon 壬午 case exposes positive front-phase 壬水/七杀/杀印相生 evidence and adverse back-phase 午火/劫财 evidence while retaining correct aggregate scoring.

## 3. Professional Evidence Presentation

- [x] 3.1 Extend result-page Dayun road types to consume optional phase evidence without breaking older aggregate-only result payloads.
- [x] 3.2 Render aggregate, front-five, and back-five evidence groups in the professional road-evidence modal with concise signed deltas and deterministic detail.
- [x] 3.3 Preserve the ordinary current-road card and existing modal accessibility behavior.

## 4. Regression Coverage and Verification

- [x] 4.1 Add backend tests for phase split, per-phase Ten God confidence weighting, aggregate compatibility, gated 杀印相生, and non-match boundaries.
- [x] 4.2 Add focused frontend tests for phase grouping and aggregate-only fallback.
- [x] 4.3 Run targeted Go and frontend tests, production build, and browser verification for the 1995-10-12 noon chart in professional mode.
