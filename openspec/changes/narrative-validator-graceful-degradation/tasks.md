## 1. Keyword bucket classification

- [ ] 1.1 Add four exported `validatedKeywords*` slices in `backend/internal/service/year_narrative_validate.go`:
      `hardRelationKeywords` (伏吟 反吟 三合 三会 大运合化),
      `positionAnchorKeywords` (用神位 忌神位 喜神位),
      `eventMarkerKeywords` (受冲 受刑 双重命中 力度倍增),
      `shenshaKeywords` (24 神煞 from the original list).
- [ ] 1.2 Keep the union `validatedKeywords` slice for backward compatibility (used by `ExtractEvidenceKeywords` if it's reintroduced later).
- [ ] 1.3 Add unit test asserting union equals original list (no terms lost in reclassification).

## 2. Sentence segmentation

- [ ] 2.1 Add `splitChineseSentences(narrative string) []string` returning sentences with their terminators preserved.
- [ ] 2.2 Test cases: single sentence, multiple sentences, trailing fragment without terminator, all sentence terminators (`。！？；…`).
- [ ] 2.3 Test on a real Phase 1 output sample to confirm it does not mis-split numeric or quoted content.

## 3. NarrativeValidationResult type + new ValidateYearNarrative

- [ ] 3.1 Define `NarrativeValidationResult` struct with four fields per `design.md` Decision 3.
- [ ] 3.2 Rewrite `ValidateYearNarrative(narrative, signals) NarrativeValidationResult`:
      • collect evidence text
      • split into sentences
      • for each sentence: scan for hard/position/eventMarker terms not in evidence → mark for clear; scan for shensha not in evidence → mark for soft-warn
      • build CleanedNarrative by:
        - dropping cleared sentences
        - in soft-warned sentences, append inline marker `(注：未在本年算法 evidence 中识别)` immediately after the offending term
      • return all four fields populated
- [ ] 3.3 Keep `ValidateYearNarrative` callable from existing tests (rename the legacy bool-returning function if needed, or add a thin adapter).

## 4. Validator tests

- [ ] 4.1 Test: clean narrative passes unchanged (CleanedNarrative == input, all slices empty).
- [ ] 4.2 Test: single HardRelation violation clears one sentence, keeps others.
- [ ] 4.3 Test: single 神煞 violation injects inline marker, keeps full sentence.
- [ ] 4.4 Test: multiple violations across sentences are handled correctly.
- [ ] 4.5 Test: violations of mixed classes (HardRelation in sentence A + 神煞 in sentence B) handled independently.
- [ ] 4.6 Test: all sentences hit HardRelation → CleanedNarrative is empty string.
- [ ] 4.7 Test: terminator-less trailing fragment is treated as a sentence.
- [ ] 4.8 Test: HardKeywordsHit and SoftWarnedTerms are populated and de-duplicated.
- [ ] 4.9 Run `go test ./internal/service/...` to confirm new tests pass and old tests still pass (after the call-site migration in section 5).

## 5. Caller migration in GenerateDayunSummariesStream

- [ ] 5.1 Update `report_service.go::GenerateDayunSummariesStream` (around the per-year validator loop) to call the new signature.
- [ ] 5.2 Assign `validatedYears[i].Narrative = result.CleanedNarrative` (was blanking on `!ok`).
- [ ] 5.3 Replace the current `校验失败丢弃 narrative` log with a structured one:
      `log.Printf("[ValidatorAction] dayun=%d year=%d cleared=%v soft_warned=%v", ...)`.
- [ ] 5.4 Run `go build ./...` to catch any compile errors.

## 6. Algorithm version bump

- [ ] 6.1 Change `repository.CurrentAlgorithmVersion` from `"v2-yongshen-shishen"` to `"v3-narrative-graceful"`.
- [ ] 6.2 No migration needed; column is already in place from Phase 1.
- [ ] 6.3 Update the comment on `CurrentAlgorithmVersion` to mention v3 semantics.

## 7. Regression sanity

- [ ] 7.1 Generate a fresh past-events recap for chart 17129161 (or another test chart) and inspect 3 year cards that previously had empty narratives.
- [ ] 7.2 Confirm those cards now contain narrative text — possibly with one cleared sentence (HardRelation hit) but with the rest of the narrative intact.
- [ ] 7.3 Confirm `algorithm_version = 'v3-narrative-graceful'` on new rows.
- [ ] 7.4 Confirm no token cost increase (Phase 2 has no prompt changes).

## 8. Observability + 2-week measurement

- [ ] 8.1 Two weeks after launch, query empty-year rate. Target: < 1% (down from 3.3%).
- [ ] 8.2 Query `[ValidatorAction]` log frequency per dayun call. Sanity-check: HardRelation drops ≤ 1 per call, soft-warns < 0.3 per call on average.
- [ ] 8.3 If empty rate is still > 1%, evaluate Phase 3 (retry-on-violation).
- [ ] 8.4 Append result note to `docs/superpowers/specs/` and link from this proposal.
