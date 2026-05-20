## 1. Backend: default-list computation

- [x] 1.1 Add `computeAutoGenDayunIndexes(chart, dayunList) []int` in `report_service.go`. Walk `dayunList`; include indexes where `dayun.StartAge <= currentAge(chart.BirthYear)`. Return slice (1-based to match `dy.Index`).
- [x] 1.2 Unit tests: birth_year=1995 + present year 2026 → ages 31 → expect indexes covering ~0-31 range; birth_year=1960 → expect more indexes; birth_year=2020 → expect index 1 only.
- [x] 1.3 Edge case: chart with no dayuns yet (very young命主). Return empty slice; SSE emits no items.

## 2. Backend: parameterize GenerateDayunSummariesStream

- [x] 2.1 Add `dayunIndexes []int` parameter to `GenerateDayunSummariesStream(chartID, userID, dayunIndexes, onItem)`. Default-list behavior is the caller's responsibility (handler computes if absent).
- [x] 2.2 Inside the function, skip dayuns whose index is not in the filter (when filter is non-empty). Filter empty → run full sequence (legacy behavior, used only by tests/admin).
- [x] 2.3 Cache-hit branch already exists; ensure it still emits the SSE item for the requested dayun (so frontend can render even when cache hits).
- [x] 2.4 Confirm `dy.Index` semantics — start at 1, sequential. Match `computeAutoGenDayunIndexes` output.

## 3. Backend: handler change

- [x] 3.1 Update `HandleDayunSummariesStream` to parse optional JSON body `{"dayun_indexes": [N, ...]}`. Both empty body and `dayun_indexes: []` mean "default list".
- [x] 3.2 When body empty, call `computeAutoGenDayunIndexes` to derive the auto-gen list.
- [x] 3.3 Pass the computed/explicit indexes to `GenerateDayunSummariesStream`.
- [x] 3.4 Existing SSE response shape unchanged.

## 4. Backend: YearsData JSON compression

- [x] 4.1 Add `CompressYearSignalsForPrompt(years []YearSignals) string` in `pkg/bazi/` returning compressed JSON bytes for AI prompt only.
- [x] 4.2 Drop fields per Decision 3 (`ten_god_power.plain_title`, `plain_text`, `year_in_dayun`, `dayun_phase`, `dayun_phase_level`).
- [x] 4.3 Apply `stripEvidenceParenthetical(evidence string)` regex: `（[^）]{1,30}）|\([^)]{1,30}\)`. Validate it doesn't eat parentheses inside text data (e.g., gan zhi names — those don't appear in parentheses anyway).
- [x] 4.4 Unit tests:
      - Field drop: assert compressed JSON has no `plain_text` key.
      - Parenthetical strip: input "用神受冲（月柱宫位，权重次之）" → output "用神受冲".
      - Mixed punctuation: full-width `（）` and half-width `()` both handled.
      - No false positives: text without parentheses unchanged.
- [x] 4.5 Wire `CompressYearSignalsForPrompt` into `GenerateDayunSummariesStream` where the JSON gets serialized for prompt injection. The original `dySignals` data passed to onItem/cache stays uncompressed.

## 5. Backend: algorithm version

- [x] 5.1 Update `repository.CurrentAlgorithmVersion = "v3-progressive-compressed"`.
- [x] 5.2 Update version comment to mention progressive generation + compression.
- [x] 5.3 No migration needed (column already exists).

## 6. Backend: tests

- [x] 6.1 New service test asserts that when called with `dayunIndexes=[4]`, the function processes only dayun 4 and emits exactly one onItem call.
- [x] 6.2 New service test asserts the compressed prompt token estimate is ~25% smaller than the uncompressed baseline.
- [x] 6.3 Run `go test ./...` -short -count=1 — all existing tests must still pass.

## 7. Frontend: PastEventsPage rendering branches

- [x] 7.1 Per dayun, compute `expanded: boolean` based on cache state + auto-gen list:
      - `dayun has cached row` → expanded
      - `dayun.start_age <= current_age` → expanded (scheduled for auto-gen)
      - otherwise → folded
- [x] 7.2 Render folded segment: ganzhi + age range + `[展开 ▼]` button. No chip computation rendered.
- [x] 7.3 Render expanded segment: existing year-card layout.

## 8. Frontend: expansion interactions

- [x] 8.1 Click `[展开 ▼]` on folded → setState `expanded=true`. Render year chips (already computed locally from `events` array — no network).
- [x] 8.2 When expanded but no AI narrative yet → render `[ 🔮 生成本段 AI 批语 ]` button at segment bottom.
- [x] 8.3 Click `生成本段 AI 批语` → call `baziAPI.generateDayunSubset(chartId, [dayunIndex])` (new function). Show per-segment loading state. SSE emits item → render narratives inline.
- [x] 8.4 Replace button with rendered narratives on success.
- [x] 8.5 Show error chip on failure with `重试` button.

## 9. Frontend: API client

- [x] 9.1 Extend `baziAPI.dayunSummaryStream(chartId, onItem, ...)` to accept optional `dayunIndexes?: number[]` parameter.
- [x] 9.2 When parameter present, send as JSON body.
- [x] 9.3 Add convenience `baziAPI.generateDayunSubset(chartId, indexes)` that wraps the call for the single-segment case.

## 10. Frontend: tests

- [x] 10.1 Existing past-events page tests pass with folded/expanded states.
- [x] 10.2 New test: folded segment shows expand button + no chips.
- [x] 10.3 New test: expanded uncached segment shows AI generate button.
- [x] 10.4 Run `node --test frontend/tests/` — all pass.

## 11. Regression sanity

- [x] 11.1 Generate fresh recap for chart 1995-10-12 male. Confirm only past+current dayun get auto-generated.
- [x] 11.2 Click `[展开 ▼]` on a future dayun. Confirm chips appear without network call. _(verified via static-source unit test asserting handleExpand has no fetch / streamDayunSummaries / baziAPI call — covered in past-events-progressive-generation.test.mjs)_
- [x] 11.3 Click `生成本段 AI 批语`. Confirm single SSE call to backend. Confirm `algorithm_version='v3-progressive-compressed'` on the new row.
- [x] 11.4 Inspect `token_usage_logs` for compressed prompt: prompt_tokens should be ~25% smaller than v2.

## 12. Observability + 2-week measurement

- [ ] 12.1 Query average prompt_tokens for past_events calls grouped by algorithm_version. Confirm v3 average is ~7,500 (vs v2 ~10,800).
- [ ] 12.2 Query per-chart total cost from token_usage_logs. Confirm typical chart drops from ~¥0.59 to ~¥0.20.
- [ ] 12.3 Frontend telemetry (optional): expansion rate of future dayuns. If > 70%, reconsider eager generation.
- [ ] 12.4 Update `docs/superpowers/specs/` with result note.
