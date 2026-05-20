## 1. Engine Refactor and Baseline Safety

- [x] 1.1 Add focused unit tests that lock current compatibility outputs for representative high, medium, and low compatibility chart pairs.
- [x] 1.2 Refactor `AnalyzeCompatibility` into internal signal-builder functions without changing public behavior.
- [x] 1.3 Add helper types/functions for stable evidence source identifiers, source contribution caps, and directional metadata.
- [x] 1.4 Verify existing compatibility service, handler, and frontend tests still pass after the refactor.

## 2. Depth Signal Families

- [x] 2.1 Implement directional ten-god relationship signal generation with self-to-partner and partner-to-self evidence.
- [x] 2.2 Add ten-god signal tests for support, pressure, attraction, expression, and peer/competition meanings.
- [x] 2.3 Implement conservative favorable-element support and pressure signals using available chart structure.
- [x] 2.4 Add favorable-element tests covering support tendency, pressure tendency, and unavailable-yongshen fallback language.
- [x] 2.5 Implement expanded heavenly-stem interaction detection across participant pillars.
- [x] 2.6 Implement expanded earthly-branch interaction detection across participant pillars, including combination, meeting, clash, punishment, harm, and break patterns.
- [x] 2.7 Add gan-zhi interaction tests covering pillar priority and mixed positive/negative evidence.
- [x] 2.8 Implement relationship-pattern synthesis for communication style, conflict trigger, security pattern, reality pressure, and repairability.
- [x] 2.9 Add relationship-pattern tests proving summaries are traceable and bounded.
- [x] 2.10 Decide whether timing signals ship in this change; if included, implement non-deterministic timing evidence and tests. Decision: defer timing signals to a later change.

## 3. Explainable Scoring

- [x] 3.1 Add scoring-layer source caps so repeated signals from one source cannot dominate a dimension.
- [x] 3.2 Ensure every score adjustment is represented by returned evidence.
- [x] 3.3 Add score explanation summaries for strongest positive and negative factors per dimension.
- [x] 3.4 Add tests for score bounds, source caps, mixed evidence explanations, and limited-evidence neutral language.

## 4. API, Prompt, and Report Integration

- [x] 4.1 Extend compatibility evidence model/API types with optional directional metadata only if required by the new signal output.
- [x] 4.2 Update compatibility prompt data to include grouped depth evidence and score explanations.
- [x] 4.3 Update the fallback compatibility prompt to require evidence-key grounding and prohibit deterministic event/date claims.
- [x] 4.4 Add backend service tests for prompt data, evidence grouping, and legacy reading compatibility.

## 5. Frontend Evidence Presentation

- [x] 5.1 Update frontend compatibility API types for any optional evidence metadata and score explanations.
- [x] 5.2 Group professional evidence by signal family or relationship meaning on the compatibility result page.
- [x] 5.3 Keep empty evidence groups hidden and preserve the consultation-first result hierarchy.
- [x] 5.4 Add frontend/static tests for professional evidence grouping, empty-group omission, and mixed-evidence copy.

## 6. Verification

- [x] 6.1 Run focused Go tests for `backend/pkg/bazi`.
- [x] 6.2 Run focused Go tests for compatibility service, repository, and handler packages.
- [x] 6.3 Run frontend static tests covering compatibility result rendering.
- [x] 6.4 Run frontend build and lint if dependencies are available.
- [x] 6.5 Manually inspect one ordinary compatibility reading and one professional-detail expansion to confirm depth evidence remains readable.
