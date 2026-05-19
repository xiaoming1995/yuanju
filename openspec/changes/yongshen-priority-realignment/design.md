## Context

The past-events recap module is a multi-stage pipeline:

```
chart_id
  ↓
LoadOrCalculateResult         (engine.go)
  ↓
BaziResult { Yongshen, Jishen, Tiaohou, strength, ... }
  ↓
For each dayun:
  GetYearEventSignalsWithContext  ← Algorithm side: produces EventSignals with Polarity
    ↓
  Prompt template { natalSummary, yongshenInfo, yearsData }
    ↓
  AI (DeepSeek-V4-Pro SSE)
    ↓
  parsed.Years[].Narrative
    ↓
  ValidateYearNarrative  ← Guard: drops narrative if keyword has no evidence
    ↓
  ai_dayun_summaries cache
    ↓
  frontend PastEventsPage renders year cards
```

The algorithm side already has correct 调候/扶抑用神 computation. The gap is **how that information reaches the AI**. Currently the prompt exposes:

```
{{if .YongshenInfo}}用忌神：{{.YongshenInfo}}{{end}}   ← e.g., "用神：木火 / 忌神：金水"
{{if .StrengthDetail}}身强弱：{{.StrengthDetail}}{{end}} ← e.g., "弱(评分-12): ..."
```

The AI must then derive 喜忌十神 from these strings on the fly. Empirically the AI is poor at this — it commonly writes "财官透干为吉" without checking whether the chart is 身旺 (where 财官 are favorable) or 身弱 (where they exhaust the day master).

## Goals / Non-Goals

**Goals:**

- Make `BaziResult` carry explicit `FavorableShishen []string`, `AdverseShishen []string`, and `ShishenConfidence` ∈ {hard, medium, soft}.
- Surface these in the AI prompt as a single readable line per scope.
- Keep the existing `Yongshen`/`Jishen` 五行 strings unchanged for backward compatibility.
- Add `algorithm_version` tracking so Phase 2 can ship without losing the rollback path.
- Maintain deterministic output for all unit-test snapshots.

**Non-Goals (Phase 1):**

- Do not change `applyPolarity` priority. 神煞 still wins over baseline today; Phase 2 will revisit.
- Do not modify `event_signals.go::getYongshenBaseline`. Phase 1 only adds prompt context.
- Do not invalidate cache.
- Do not touch frontend rendering.
- Do not introduce 从化局 detection. That requires a separate strength-classification overhaul.

## Decisions

### Decision 1: 古法 mapping rule (not 实战派)

Use the classical 《滴天髓》/《子平真诠》 mapping without 大运 modifier:

```
身旺 (vstrong, strong)  → 喜 食神 / 伤官 / 偏财 / 正财 / 正官 / 七杀 (泄/克/耗)
                         忌 比肩 / 劫财 / 偏印 / 正印      (生/扶)

身弱 (vweak, weak)      → 喜 偏印 / 正印 / 比肩 / 劫财    (生/扶)
                         忌 食神 / 伤官 / 偏财 / 正财 / 正官 / 七杀 (泄/克/耗)

身中和 (neutral)         → soft confidence: 不硬给清单
                          AI prompt 提示 "喜忌不显，以调候为主"
```

**Why 古法 first**: 80% accuracy with minimal complexity. 实战派 (大运 modifier) adds 8x data structure overhead (`map[string][]string` per dayun) for incremental gain on edge cases. We can layer 实战派 later if Phase 1 metrics show 古法 is insufficient.

### Decision 2: Confidence band, not binary

```go
type ShishenConfidence string

const (
    ShishenConfHard   ShishenConfidence = "hard"   // strength=vstrong/vweak
    ShishenConfMedium ShishenConfidence = "medium" // strength=strong/weak
    ShishenConfSoft   ShishenConfidence = "soft"   // strength=neutral
)
```

**Why**: 中和 命主 (strength=neutral) historically resist binary 喜忌 judgments. The 滴天髓 stance for 中和 is "以调候为主, 扶抑次之". A hard-coded list mislabels them. Soft confidence lets the prompt say "喜忌不显" instead of forcing a wrong list.

### Decision 3: Prompt injection format

Append a short, readable block to the existing dayun-summary template (after `用忌神` line):

```
{{if eq .ShishenConfidence "hard"}}
本命喜十神：{{.FavorableShishen}}    本命忌十神：{{.AdverseShishen}}
{{else if eq .ShishenConfidence "medium"}}
本命偏向喜十神：{{.FavorableShishen}}（中等强度）
{{else}}
本命喜忌不显（中和命主），以调候用神{{.Tiaohou}}为主
{{end}}
```

**Why this format**:
- Short (1-2 lines), token-cheap (~30 tokens per dayun call).
- Confidence-gated wording matches the doctrine's flexibility.
- Reading-order placement (after 用忌神) keeps the AI's attention chain natural: 用神 五行 → 喜忌十神 → 流年逐年.

### Decision 4: algorithm_version columns

Add `algorithm_version VARCHAR(32) DEFAULT NULL` to:
- `ai_dayun_summaries`
- `ai_reports`

Phase 1 writes `'v2-yongshen-shishen'` for new rows. Old rows remain NULL (interpreted as v1).

**Why now (not in Phase 2)**:
- Phase 2 will need this for cache invalidation logic.
- Adding it later is harder (would need to backfill or interpret NULL ambiguously).
- Cost of adding now is one migration + ~5 lines in upsert paths.

### Decision 5: No cache invalidation in Phase 1

Existing cached dayun summaries keep their old algorithm output. Users see the new prompt only when they trigger a regeneration (e.g., clicking "重新生成" on an empty year) or when a new chart is recapped for the first time.

**Why**: Phase 1 is opt-in observable. Wholesale invalidation would force the entire user base through a re-recap with no rollback path. Phase 2's metrics decide whether to invalidate.

## Risks

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| 古法 mapping wrong for 从化局 charts | Medium (~2% of charts) | Out of scope; flag separately. Soft confidence helps. |
| AI ignores the new prompt line | Low (DeepSeek follows enumerated rules well) | Test sample of 5 reports post-launch. |
| Prompt token cost spike | Very low (~30 tokens / call) | Token-usage dashboard will show this immediately. |
| `algorithm_version` column breaks rollback | Very low (additive) | Migration is one ALTER TABLE; no data change. |
| New `BaziResult` JSON shape breaks existing consumers | Low (additive fields) | All consumers do field-by-field reads, not whole-struct decode. |

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│              Phase 1 changes (in green, additive)               │
└─────────────────────────────────────────────────────────────────┘

  BaziResult                          │
  ─────────────                       │
  Yongshen      "木火"                │  (existing — unchanged)
  Jishen        "金水"                │  (existing — unchanged)
  Tiaohou       {Expected: ["丙"]}    │  (existing — unchanged)
  StrengthLevel "weak"                │  (existing — unchanged)
  + FavorableShishen ["偏印","正印",  │  NEW (computed from Yongshen
                       "比肩","劫财"] │       + StrengthLevel)
  + AdverseShishen   ["食神","伤官",  │  NEW
                       "偏财","正财", │
                       "正官","七杀"] │
  + ShishenConfidence "medium"        │  NEW
                                      │
                                      ▼
  GenerateDayunSummariesStream
  ────────────────────────────
  prompt template + (new injected block)
                                      │
                                      ▼
  AI generates narratives with full 喜忌 context
                                      │
                                      ▼
  parsed.Years[].Narrative
                                      │
                                      ▼
  ValidateYearNarrative (unchanged in Phase 1)
                                      │
                                      ▼
  ai_dayun_summaries.algorithm_version = 'v2-yongshen-shishen'  NEW column
```

## Open Questions

- Q: Should `FavorableShishen` for `vstrong` differ from `strong`? (古法 says no, but very-strong charts theoretically prefer stronger 泄 like 食伤 over weaker 财)
  - A: Phase 1 treats them identically. Phase 2 can refine.

- Q: When `Yongshen` includes both `木` and `火` but `strength=weak`, are 比肩/劫财 (生木) and 偏印/正印 (生木) equally favorable?
  - A: Yes, both flagged as 喜. AI distinguishes via 流年 context.

## Lessons Learned (in-flight, captured post-attempt)

### Patch 4 (allowed_keywords whitelist) — reverted

During Phase 1 validation we attempted to close the last 3% empty-year gap by injecting a per-year `allowed_keywords` array into the signals JSON, paired with a prompt clause "narrative 中只能使用 allowed_keywords 列出的命理术语". Hypothesis: AI would treat the list as a hard restriction.

**Single-run measurement on chart 17129161**:
- Before Patch 4 (Patches 1-3 alone): 3/90 = 3.3% empty cards, 3 validator drops (all 伏吟)
- With Patch 4: 9/90 = 10% empty cards, 9 validator drops (伏吟 + 驿马 + 三会 + 亡神 + ...)

**Observed failure mode**: AI read the whitelist as a *suggestion list* rather than a *restriction list*. Listing valid terms apparently made the AI more inclined to decorate narratives with them — including across year boundaries (year X's allowed_keywords leaked into year Y's narrative).

This is essentially a prompt-injection backfire: a list labeled "allowed" reads to an LLM more like "options you should consider" than "the only choices permitted".

**Decision**: Reverted via `git revert 16918ea`. Do not retry this strategy in Phase 2 — at least not in this naive form. Phase 2 keyword-restriction work should be **validator-side, not prompt-side**:
- Option A: Validator clears only the offending sentence, not the whole narrative (preserves the 95% that's correct).
- Option B: Validator triggers a single retry with the offending term explicitly forbidden in the retry prompt.
- Option C: Soften validator for the 24 神煞 set (warn + render with footnote rather than blank); keep hard validator for relation words (伏吟/反吟/受冲/受刑).

These options do not depend on AI compliance with a prompt-side whitelist, sidestepping the backfire.

### Phase 1 net result

After revert: 5 commits substantively improving the pipeline (喜忌十神 injection + algorithm_version + cached snapshot self-upgrade + 3 silent-failure log adds). Single-chart end-to-end test: **3.3% empty rate vs 60% baseline = 18x reduction**. Phase 2 trigger threshold (40%) not met; remaining gap is concentrated on 伏吟 over-confidence, which a validator-side fix (see Phase 2 options above) should close.
