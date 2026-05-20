## Why

`/api/bazi/past-events/dayun-summary-stream/:chart_id` currently generates AI narratives for **all 9 dayun segments** in a single page-load, regardless of whether the user actually intends to read them. Each AI call averages ~12,000 prompt tokens (range 7K–57K observed in production), and a full recap is ~108K tokens per chart — approximately ¥0.59 at current DeepSeek-V4-Pro pricing.

Live measurement (305 logged past_events calls):

```
prompt   min 7,004  avg 10,783  max 57,066
output   min 405    avg 1,382   max 11,999
total    min 7,596  avg 12,166  max 69,065
```

The economics are wasteful for two reasons:

1. **Most users do not read all 9 dayuns.** They open the page, scan their own life so far (1–3 segments), and rarely engage with future dayun segments at all. Eager generation means we pay AI fees on segments no one reads.

2. **Each dayun prompt carries ~7K tokens of YearsData JSON** that includes redundant fields (`year_in_dayun`, `dayun_phase`, `dayun_phase_level` are derivable from the index; `ten_god_power.plain_title` and `plain_text` are decorative labels for the frontend, not for AI consumption) and verbose evidence strings with parenthetical asides AI does not need.

This change attacks both fronts: fewer AI calls per page (progressive generation) AND smaller per-call prompt (JSON compression).

## What Changes

### Behavioral changes — progressive generation

- **Auto-generate only past + current dayun segments** on page load. The cutoff is the dayun that contains the user's current age. Future dayuns render as **collapsed headers** showing only the dayun's ganzhi and age range, with an expand affordance.
- **Two-step expansion for future dayuns:**
  - Click `[展开 ▼]` → unfolds the segment, showing the algorithm-computed year chips for all 10 years (no AI call, zero token cost).
  - Click `生成本段 AI 批语` button → triggers a single-dayun AI generation for that segment.
- **Cached future dayuns auto-unfold.** If a user previously generated a future dayun (cache hit), the segment renders expanded with narratives inline; no button is shown.

### Endpoint changes

- `POST /api/bazi/past-events/dayun-summary-stream/:chart_id` now accepts an optional JSON body `{"dayun_indexes": [N, ...]}`:
  - **Empty or missing body**: server computes the auto-gen list (dayun segments whose start age ≤ user's current age) and streams only those.
  - **Explicit indexes**: server streams only those segments. Used for click-to-generate of future dayuns.
- SSE protocol unchanged. Each emitted item still has the `dayun_index` field so the frontend can locate it in the timeline.

### Prompt token compression (B side)

YearsData JSON shrinks via three transforms before being injected into the AI prompt. None of these affect the persistence schema or the frontend payload — they apply only to the AI-prompt serialization path:

- Drop `ten_god_power.plain_title` and `ten_god_power.plain_text` (frontend-facing labels, not AI input).
- Drop `year_in_dayun`, `dayun_phase`, `dayun_phase_level` from each year entry. These are derivable from the year's position in the array.
- Trim parenthetical asides from `signal.evidence` strings: substrings of the form `（…）` that AI uses as context but do not change the core claim, such as `（月柱宫位，权重次之）`, `（本年有重煞，此信号仅作参考）`. Both full-width and half-width parentheses are matched.

Expected per-dayun token savings: ~40-45% of YearsData size, which is ~25-30% of total prompt size, so ~3K tokens saved per AI call.

### Algorithm version

Bump `repository.CurrentAlgorithmVersion` from `v2-yongshen-shishen` to `v3-progressive-compressed`. Existing v1/v2 rows remain readable. Analytics can split cohorts to validate the savings.

## Capabilities

### New Capabilities
- `past-events-progressive-generation`: User-controlled lazy generation of dayun AI narratives with cached re-entry.

### Modified Capabilities
- `yongshen-driven-flow-year-judgment` (from Phase 1): the SSE endpoint's request shape gains an optional `dayun_indexes` filter; the prompt serializer applies token compression. Both modifications preserve all existing test contracts.

## Impact

### Backend

- `internal/handler/bazi_handler.go::HandleDayunSummariesStream`: parse new body field.
- `internal/service/report_service.go::GenerateDayunSummariesStream`: accept `dayunFilter []int` parameter; compute default list from `chart.BirthYear` + dayun timeline when filter is empty; serialize compressed YearsData inline.
- `pkg/bazi/event_signals.go` (or sibling): new helper `CompressYearsForPrompt(dySignals)` that produces the AI-only compressed JSON.
- New repository helper `repository.CurrentAlgorithmVersion = "v3-progressive-compressed"`.

### Frontend

- `frontend/src/pages/PastEventsPage.tsx`: render `expanded: bool` per dayun. Past + current dayuns default to expanded; future dayuns folded.
- New folded segment component showing ganzhi/age + `[展开 ▼]`.
- New "生成本段 AI 批语" button + per-dayun generation state.
- `frontend/src/lib/api.ts`: extend `dayunSummaryStream` to accept optional `dayun_indexes` array.

### API shape

- Request body for `POST /api/bazi/past-events/dayun-summary-stream/:chart_id` adds optional `dayun_indexes` array.
- SSE emit format unchanged.

### Cache and migration

- No database migration. `ai_dayun_summaries` already has `algorithm_version` column from Phase 1.
- Existing v2-yongshen-shishen rows are honored. The frontend will render them as expanded segments (cached past dayuns) or expanded with cache-hit message (future dayuns that were generated under v2).
- New writes stamp `v3-progressive-compressed`.

### Cost projection

For a chart where the user is 35 years old (mid-life) and never expands future dayuns:

```
Before:  9 dayuns × ~12K tokens ≈ 108K tokens  →  ~¥0.59
After:   4 dayuns × ~9K tokens  ≈  36K tokens  →  ~¥0.20  (66% reduction)
                       ↑                                       ↑
                  compressed                              fewer dayuns
                  prompt (-3K)                            (auto-gen only past)
```

Expected aggregate savings (assuming current usage pattern): **60–70% on monthly past_events bill**.

## Out of Scope

- Validator-side narrative recovery (covered by separate `narrative-validator-graceful-degradation` proposal).
- JSON-mode AI output enforcement (deferred; current parsing is sufficient with sentence-level recovery in the sibling proposal).
- Aggressive `档位 3` token compression with short JSON keys. Tradeoff between savings (extra ~10%) and debug pain rejected; see design.md.
- Routing to cheaper models (e.g., `deepseek-v4-flash` for past years vs `pro` for current). Separate optimization track.
- Background pre-warming of dayuns on chart creation. Out of scope; user demand should remain the trigger.

## Observability Plan (2 weeks post-launch)

- Average prompt token per past_events call: target down from ~10,783 to ~7,500.
- Per-chart total token cost: target down from ~108K to ~36K (60-70% reduction).
- Folded-segment expansion rate: % of expanded future dayuns vs total impressions. If > 70%, the lazy-load assumption is wrong; revert to eager.
- AI call count per chart per recap: target down from ~9 to ~3 on first visit.
