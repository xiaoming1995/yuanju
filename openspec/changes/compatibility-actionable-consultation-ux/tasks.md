## 1. Personality Fit Model

- [x] 1.1 Add a frontend helper that derives personality-fit summary data from dimension scores, compatibility evidence, relationship diagnosis, participant snapshots, relationship stage, primary question, and report state.
- [x] 1.2 Add helper tests for no-report fallback, structured-report enrichment, general-context fallback, fit/clash points, participant patterns, and evidence targets.
- [x] 1.3 Add a frontend helper that derives personality validation-plan sections from personality-fit summary data, decision advice, stage risks, duration assessment, and evidence availability.
- [x] 1.4 Add helper tests for 7-day checks, 30-day checks, avoid items, and conditional/non-deterministic language.

## 2. Consultation Entry Flow

- [x] 2.1 Reorder `CompatibilityPage` so personality-fit question and relationship stage are presented before the two birth profile panels.
- [x] 2.2 Add a personality-fit consultation preview that updates when relationship stage or primary question changes.
- [x] 2.3 Preserve existing create-reading request payload and authenticated navigation behavior.
- [x] 2.4 Update `CompatibilityPage.css` for compact desktop layout, mobile stacking, and bottom-nav-safe primary actions.
- [x] 2.5 Add frontend tests for entry-flow order, preview copy, payload wiring, and mobile CSS hooks.

## 3. Result Personality Fit UX

- [x] 3.1 Render a personality-fit section after the decision dashboard and before score/professional detail sections.
- [x] 3.2 Show overall personality match type, self relationship pattern, partner relationship pattern, fit points, conflict points, and communication guidance.
- [x] 3.3 Restate the selected primary question near the personality-fit answer and align the copy with that question.
- [x] 3.4 Render a 7-day, 30-day, and avoid/do-not-do validation plan after the personality-fit section using conditional verification language.
- [x] 3.5 Add supporting-section affordances for personality claims and action items when stage-risk, score, report, or evidence sections are available.
- [x] 3.6 Ensure only one no-report primary generation action remains visually dominant.
- [x] 3.7 Add result-page tests for section order, personality-fit copy, action-plan copy, non-deterministic wording, evidence links, and CSS hooks.

## 4. Relationship Archive Continuation

- [x] 4.1 Update `CompatibilityHistoryPage` cards to emphasize personality match type, stage, primary question, overall decision state, and continuation action.
- [x] 4.2 Add fallback labels for legacy/general readings without context.
- [x] 4.3 Add continuation affordances for viewing the result and, where supported by available data, generating or continuing deep reading.
- [x] 4.4 Update `CompatibilityHistoryPage.css` for scan-friendly mobile archive cards.
- [x] 4.5 Add history UX tests for context labels, fallback labels, continuation action, and mobile CSS hooks.

## 5. Verification

- [x] 5.1 Run focused frontend tests for compatibility entry, result, history, personality-fit helper, action-plan helper, and mobile QA.
- [x] 5.2 Run `npm run lint` from `frontend/`.
- [x] 5.3 Run `npm run build` from `frontend/`.
- [ ] 5.4 Smoke test the compatibility flow in the in-app browser for entry, result, and history hierarchy when local backend data is available.
