## 1. Narrative Evidence Mapping

- [x] 1.1 Inspect current past-events year narrative, evidence summary, generated dayun year fallback, and frontend year-card selection logic.
- [x] 1.2 Add test fixtures covering branch clash, punishment, combination, void, shensha, ten-god, `夹拱`, and sparse-evidence years.
- [x] 1.3 Define a small evidence-to-plain-language mapping for common technical triggers and life-area translations.

## 2. Deterministic Narrative Rendering

- [x] 2.1 Update deterministic year narrative rendering to select the strongest evidence-backed signal before generic theme text.
- [x] 2.2 Render the narrative in the order of yearly tendency, reason, likely life areas, and practical stance.
- [x] 2.3 Keep wording conservative and non-absolute while allowing longer evidence-supported paragraphs.
- [x] 2.4 Ensure sparse or weak evidence falls back to a shorter conservative summary without invented events.

## 3. Generated Text Fallback And Selection

- [x] 3.1 Add or update validation so generated year text that is missing, too short, or overly generic does not override evidence-aligned deterministic text.
- [x] 3.2 Ensure dayun-summary year fallback uses the same evidence-aligned narrative renderer.
- [x] 3.3 Preserve existing API fields, SSE payload shape, cache semantics, and progressive-generation behavior.

## 4. Frontend Year Card Presentation

- [x] 4.1 Verify year cards prefer the improved readable narrative before technical evidence.
- [x] 4.2 Keep `命理依据` visible as secondary audit detail without replacing it with paraphrase only.
- [x] 4.3 Adjust copy or layout only if longer narratives cause readability, overflow, or mobile spacing issues.

## 5. Verification

- [x] 5.1 Add or update backend tests asserting narratives explain concrete evidence in plain language.
- [x] 5.2 Add or update tests for generated-text fallback when generated text is absent or generic.
- [x] 5.3 Add or update frontend tests if year-card display selection or layout changes.
- [x] 5.4 Run relevant backend tests for bazi narrative/report service behavior.
- [x] 5.5 Run relevant frontend tests and build/type-check verification.
- [x] 5.6 Manually verify at least one chart year where `命理依据` previously looked more accurate than the visible narrative.
