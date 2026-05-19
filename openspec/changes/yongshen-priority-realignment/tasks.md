## Phase 1 — Explicit 喜忌十神 injection (this change)

### 1. Mapping function

- [ ] 1.1 Add `BuildFavorableShishen(dayGan, yongshen, jishen string, strength string) (favorable, adverse []string, confidence string)` in `backend/pkg/bazi/yongshen.go`.
- [ ] 1.2 Implement 古法 mapping per `design.md` Decision 1:
      vstrong/strong → 喜 食神 伤官 偏财 正财 正官 七杀；忌 比肩 劫财 偏印 正印
      vweak/weak     → 喜 偏印 正印 比肩 劫财；忌 食神 伤官 偏财 正财 正官 七杀
      neutral        → both empty, confidence="soft"
- [ ] 1.3 Compute confidence band:
      vstrong/vweak → "hard"; strong/weak → "medium"; neutral → "soft".

### 2. Tests for mapping

- [ ] 2.1 Add `backend/pkg/bazi/yongshen_test.go` test cases:
      one per (5 strength tiers × 1 representative day-master) = 5 cases minimum.
      Each verifies favorable list, adverse list, and confidence string.
- [ ] 2.2 Edge case: 当 Yongshen/Jishen 五行字段为空时返回 empty lists + confidence="soft".
- [ ] 2.3 Run `go test ./pkg/bazi/...` and confirm all pass.

### 3. Integrate into BaziResult

- [ ] 3.1 Add three new JSON-tagged fields to `BaziResult` in `backend/pkg/bazi/engine.go`:
      `FavorableShishen []string \`json:"favorable_shishen,omitempty"\``
      `AdverseShishen   []string \`json:"adverse_shishen,omitempty"\``
      `ShishenConfidence string  \`json:"shishen_confidence,omitempty"\``
- [ ] 3.2 In `inferYongshenWithTiaohouPriority` callsite (or wherever Yongshen/Jishen are finalized), call `BuildFavorableShishen` and populate the new fields on the returned `BaziResult`.
- [ ] 3.3 Run `go build ./...` and confirm no compile errors.

### 4. Prompt injection

- [ ] 4.1 Extend `model.DayunSummaryTemplateData` in `backend/internal/model/model.go` with three new fields matching `BaziResult` additions.
- [ ] 4.2 Wire them in `GenerateDayunSummariesStream` (~line 1237 in `report_service.go`) when constructing `tplData`.
- [ ] 4.3 Update the prompt template (around `report_service.go:1125`) to add the confidence-gated block per `design.md` Decision 3.
- [ ] 4.4 Add an integration test in `backend/internal/service/` that renders the template with a fixture `BaziResult` and asserts the new lines appear in the prompt output.

### 5. algorithm_version columns

- [ ] 5.1 Add migration `backend/pkg/database/migrations/00006_add_algorithm_version_columns.sql`:
      `ALTER TABLE ai_dayun_summaries ADD COLUMN algorithm_version VARCHAR(32);`
      `ALTER TABLE ai_reports ADD COLUMN algorithm_version VARCHAR(32);`
- [ ] 5.2 Update `UpsertDayunSummary` (in `dayun_summary_repository.go`) to write `'v2-yongshen-shishen'` for new rows.
- [ ] 5.3 Update `ai_reports` insert path similarly (search for the report-cache upsert).
- [ ] 5.4 Verify the migration runs cleanly via `docker-compose up -d backend` and `goose` startup log.

### 6. Regression sanity

- [ ] 6.1 Generate a fresh past-events recap for one known chart and grep the backend prompt log for the new lines.
- [ ] 6.2 Manually inspect 3 generated year narratives — confirm AI references 喜忌十神 when judging 财官印比 polarity.
- [ ] 6.3 Confirm token usage logs show prompt size increased by < 50 tokens per call.

### 7. Observability + 2-week measurement

- [ ] 7.1 After 2 weeks, query: how many `ai_dayun_summaries.years` rows still have empty narratives?
      Compare to baseline 60%. Target: <30%.
- [ ] 7.2 Query: validator drop log frequency before vs after.
- [ ] 7.3 Query: cost dashboard — confirm prompt growth held under +5% of total recap cost.
- [ ] 7.4 Write 1-page Phase 1 result note in `docs/superpowers/specs/` and link from this proposal.

## Phase 2 — Polarity priority realignment (deferred, NOT this change)

Track-only. Do not implement under this proposal.

- [ ] 8.1 Realign `applyPolarity` so 神煞 no longer overrides baseline (`design.md` Phase 2 scope).
- [ ] 8.2 Enhance `getYongshenBaseline` to consider 地支藏干透出十神 + 喜忌神位 受冲克合.
- [ ] 8.3 Demote 20+ minor 神煞 from independent EventSignal to bracket notes on main-line evidence.
- [ ] 8.4 Cache invalidation strategy: clear dayun summaries with empty years.
- [ ] 8.5 Frontend chip display adjustment for collapsed神煞 evidence.

**Gate for Phase 2 start**: Phase 1 metrics (task 7) must be available and show that prompt-only injection got us to <30% empty year rate. If <15%, Phase 2 may not be needed. If >40%, Phase 2 scope may need expansion.
