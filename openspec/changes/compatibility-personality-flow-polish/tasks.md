## 1. Entry Flow Polish

- [x] 1.1 Add static tests for entry progress cues, section order, preview updates, and mobile-safe CSS hooks.
- [x] 1.2 Update `CompatibilityPage` to show compact consultation progress for question, relationship stage, and birth-profile completion.
- [x] 1.3 Refine entry-page copy so personality fit remains the first user-facing concept.
- [x] 1.4 Update `CompatibilityPage.css` for desktop compactness, mobile stacking, and bottom-nav-safe actions.

## 2. Result Hierarchy Polish

- [x] 2.1 Add static tests for result anchor navigation, target section ids, merged validation hierarchy, match-type explanations, and reduced-density CSS hooks.
- [x] 2.2 Add match-type description metadata to the existing personality helper or result-page derivation path.
- [x] 2.3 Add a compact result reading map with links to personality fit, conflict/validation, score/evidence, and professional details.
- [x] 2.4 Merge or visually group the existing personality validation plan with stage-risk validation so the page does not show duplicate "what to validate" modules.
- [x] 2.5 Refine `CompatibilityResultPage.css` to reduce nested-card density in personality and validation sections while preserving responsive layout.
- [x] 2.6 Confirm the deep-report generation action remains singular and visually secondary to the completed personality reading when a report is absent.

## 3. History Scan Polish

- [x] 3.1 Add static tests for history card priority, personality match label, continuation copy, score de-emphasis, and mobile CSS hooks.
- [x] 3.2 Update `CompatibilityHistoryPage` card structure so names, match type, stage/question, and continuation action are the primary scan path.
- [x] 3.3 De-emphasize raw four-dimension scores into a compact secondary row or detail area.
- [x] 3.4 Update `CompatibilityHistoryPage.css` for compact mobile card height and tappable continuation affordance.

## 4. Verification

- [x] 4.1 Run focused compatibility frontend tests, including personality flow polish tests.
- [x] 4.2 Run `npm run lint` from `frontend/`.
- [x] 4.3 Run `npm run build` from `frontend/`.
- [ ] 4.4 Smoke test `/compatibility`, `/compatibility/history`, and at least one `/compatibility/:id` result in the in-app browser when local auth/data are available.
