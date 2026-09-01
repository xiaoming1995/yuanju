## 1. Backend Wealth Model

- [x] 1.1 Add `WealthProfile` and `WealthWindowHint` result types and attach optional `wealth_profile` to `BaziResult`.
- [x] 1.2 Implement a deterministic wealth-profile builder that scores wealth-star visibility/rooting, carrying capacity, wealth-producing chains, Ten God favorability, and retention risks.
- [x] 1.3 Add grade labels, wealth types, capped tags, risk flags, summary copy, and evidence formatting for S/A/B/C/D wealth profiles.
- [x] 1.4 Add current-Dayun wealth-window hint generation without changing the natal wealth grade.

## 2. Snapshot and Prompt Integration

- [x] 2.1 Attach `wealth_profile` during new `Calculate` results after Ten God, natal assessment, vehicle, and Dayun data are available.
- [x] 2.2 Extend `LoadOrCalculateResult` lazy backfill to rebuild and persist missing or outdated wealth profiles for saved chart snapshots.
- [x] 2.3 Extend AI report prompt context with the backend-computed wealth profile and conservative non-investment wording.

## 3. Frontend Presentation

- [x] 3.1 Extend frontend Bazi result types to include optional `wealth_profile`, evidence, risk flags, and current hint fields.
- [x] 3.2 Add a compact "财富结构" overview card with grade label, score meter, summary, capped tags, and current-window hint when available.
- [x] 3.3 Add professional evidence display for wealth-profile drivers and risks using the existing overview modal pattern.
- [x] 3.4 Adjust result overview responsive layout so vehicle, road, and wealth cards remain readable on desktop, wide desktop, tablet, and mobile widths.

## 4. Verification

- [x] 4.1 Add focused Go tests for wealth-profile grade boundaries, weak-chart carrying caps, wealth-producing chains, adverse wealth risks, and current-Dayun hint separation.
- [x] 4.2 Add service tests for saved-snapshot wealth-profile backfill and AI prompt inclusion/omission.
- [x] 4.3 Add frontend tests for wealth overview rendering, missing-profile fallback, professional evidence entry, and layout text boundaries.
- [x] 4.4 Run focused Go tests, frontend tests, production build, `openspec validate add-bazi-wealth-profile --strict`, and `git diff --check`.
- [x] 4.5 Refresh historical result detail even when route state exists, so legacy snapshots receive the server-side wealth-profile backfill.
