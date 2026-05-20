## Context

Phase 1's `yongshen-priority-realignment` brought empty-narrative rate from 60% to 3.3%, but did not address the underlying eagerness of past-events generation. Live measurement of 305 past_events AI calls shows average prompt size ~10.8K tokens with a long tail to 57K tokens. A full 9-dayun recap consumes ~108K tokens, billed at ~¥0.59 against DeepSeek-V4-Pro.

The cost is structurally wasteful for two compounding reasons:

```
┌────────────────────────────────────────────────────────────────┐
│  Per-dayun prompt structure (average ~12K total tokens)         │
├────────────────────────────────────────────────────────────────┤
│  System prompt (KB modules)        ~3,000 tokens   prompt cache │
│  Header (命主/原局/喜忌十神)         ~200 tokens     fixed       │
│  YearsData JSON (10 years signals) ~7,000 tokens    ★ variable  │
│  Output rules + 易混淆 reminder      ~1,800 tokens   fixed       │
└────────────────────────────────────────────────────────────────┘
                            ×
┌────────────────────────────────────────────────────────────────┐
│  9 dayuns per chart, all generated on page open                │
│  Most users only read 1-3 of them                              │
└────────────────────────────────────────────────────────────────┘
```

User behavior validation: in the live data, the modal time spent on the page is < 60 seconds — too short to read all 9 dayuns at 100-150 字 narrative each. The product is paying for content nobody reads.

## Goals / Non-Goals

**Goals:**
- Reduce per-chart past_events token cost by 60-70% on the typical session.
- Preserve the "完整时间轴" UX feel — users still see all 9 dayuns at a glance, just with progressive content reveal.
- Keep existing AI-prompt prompt quality (Phase 1 喜忌十神 injection stays; only the data shape changes).
- Maintain cache compatibility — v2 rows render normally; new writes are v3-progressive-compressed.

**Non-Goals:**
- Do not change AI model selection or pricing tier.
- Do not change the SSE protocol shape (each item still emits `dayun_index`, `themes`, `summary`, `years`).
- Do not pre-warm or background-generate. User intent drives all generation.
- Do not introduce JSON-mode enforcement on the AI API call (separate concern, last attempt was reverted).

## Decisions

### Decision 1: Auto-generate cutoff is "dayun containing user's current age"

For a chart with birth_year 1995, current year 2026, user age ≈ 30:
- Dayun 1 (起运～): 0-8岁 → past, auto-gen
- Dayun 2: 9-18岁 → past, auto-gen
- Dayun 3: 19-28岁 → past, auto-gen
- Dayun 4: 29-38岁 → contains current age, auto-gen
- Dayun 5-9: future, folded

The cutoff is inclusive on the current dayun. Future dayuns (those whose start_age > current_age) are folded.

**Why include the current dayun**: it contains the user's "right now" which is the most-read content. Excluding it would force a click for what should be the default attention.

### Decision 2: Future dayun folded UX requires two clicks

Folded state:
```
┌──────────────────────────────────────────────────┐
│ 庚午 大运 49-58岁                  [展开 ▼]      │
└──────────────────────────────────────────────────┘
```

Expanded (after first click):
```
┌──────────────────────────────────────────────────┐
│ 庚午 大运 49-58岁                  [收起 ▲]      │
│ ────────────────────────────────────────────     │
│ 2044 甲子  [变动][财运↓][桃花]                    │
│ 2045 乙丑  [变动][合化]                           │
│ ...                                              │
│ ────────────────────────────────────────────     │
│       [ 🔮 生成本段 AI 批语 ]                     │
└──────────────────────────────────────────────────┘
```

After AI generation completes, the button is replaced by the rendered narratives inline.

**Why two clicks**: confirmed user preference. The first click reveals algorithm-computed chips (zero cost). The second click is the explicit consent to spend AI tokens. This pattern matches the consent-gated cost philosophy from the cost-alert work.

### Decision 3: 中度 JSON compression (档位 2)

Three transforms applied during prompt serialization in `GenerateDayunSummariesStream`:

```
A. Drop ten_god_power.plain_title and plain_text
   • Current: 250 chars per year × 10 years = 2,500 chars
   • After:   90 chars per year × 10 years =   900 chars
   • Saved:   ~1,600 chars per dayun

B. Drop year_in_dayun, dayun_phase, dayun_phase_level from each year
   • Current: 40 chars per year × 10 years = 400 chars
   • After:    0 chars                     =   0 chars
   • Saved:   400 chars per dayun
   • Note: Phase information is unchanged in the stored signals; only
     omitted from the AI-prompt serialization. Frontend continues to
     receive these fields if it ever needs them.

C. Strip parenthetical asides from signal.evidence
   • Pattern: r'（[^）]{1,30}）|\([^)]{1,30}\)'  (Chinese + half-width)
   • Examples removed:
     - "（月柱宫位，权重次之）"
     - "（本年有重煞，此信号仅作参考）"
   • Saved: ~15-30 chars per signal × ~6 signals/year × 10 years
           = 900-1,800 chars per dayun
```

Total YearsData compression: ~3,000 chars per dayun = ~2,000 tokens. Combined with the unchanged baseline:

```
Per-dayun prompt:        ~12K → ~9-10K tokens  (-25%)
Per-chart auto-gen:    ~36K (4 dayuns) → ~28-32K (4 dayuns compressed)
```

**Why not 档位 3 (short keys)**: tested against the GPT family's tendency to confuse short keys. The marginal +10% savings is not worth the prompt-debugging pain and risk of AI confusion. Existing tests would all need re-validation.

### Decision 4: Endpoint contract — single endpoint with optional filter

Keep `POST /api/bazi/past-events/dayun-summary-stream/:chart_id` as the canonical endpoint. Request body becomes:

```json
{
  "dayun_indexes": [4]   // optional
}
```

Semantics:
- Empty or missing body → server computes default list (past + current dayun indexes from `chart.BirthYear`)
- Explicit list → server streams those segments only (used for click-to-generate of future dayuns)

**Why not a separate single-dayun endpoint**: the per-dayun work is identical in both cases (cache lookup + AI call + upsert + SSE emit). Splitting the endpoint would duplicate the SSE plumbing for marginal API clarity. A request-body parameter is the minimal contract change.

### Decision 5: Frontend renders cached v2 rows as expanded

Past dayuns with `algorithm_version = v2-yongshen-shishen` are honored as-is — they were generated with the older compressed-less prompt, but the rendered narrative is still valid. The frontend logic:

```
For each dayun in chart:
  if dayun has cached row (any version):
    render as expanded with full content
  elif dayun is past+current:
    render as expanded, schedule auto-gen
  else (future dayun):
    render folded with [展开 ▼] button
```

No cache invalidation. `algorithm_version` field tags new writes; old writes are visible just by being non-null.

## Risks

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| Users don't realize future dayuns can be expanded | Medium | Folded segments are visually distinct + show ganzhi/age + clear `[展开 ▼]` icon. UX copy reinforces. |
| Excessive single-dayun click bursts (user expands many futures) overwhelm the AI provider | Low | Each click is one explicit request; rate limited by existing token-usage budget alerts. |
| Compressed YearsData breaks AI narrative quality | Low | Compression only drops algorithmically-redundant or AI-irrelevant fields. Sample testing to verify in regression. |
| Default-list computation in backend gets the cutoff wrong (date math edge cases) | Low | Cutoff uses `chart.BirthYear` + the same `dayun.StartAge` field used by the algorithm. Unit-test boundary conditions. |
| Backward compatibility for v2 caches in mixed page renders | Very Low | Cached row is always rendered as-is; only the trigger logic differs. |

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│            New page-open flow (user age 30, 1995-born)          │
└─────────────────────────────────────────────────────────────────┘

   GET PastEventsPage
        │
        ▼
   Frontend computes dayun timeline (existing logic)
   → renders 9 segment cards with chips
        │
        ▼
   For each dayun, look at age range:
        │
        ├── past + current (1-4) → render expanded
        │                            │
        │                            ▼
        │                       Schedule POST stream with
        │                       dayun_indexes: [1,2,3,4]
        │                            │
        │                            ▼
        │                       SSE emits per-dayun item
        │                       (cache or AI call), each ~28K tokens
        │
        └── future (5-9) → render folded segment
                              ▼
                         User clicks [展开 ▼]
                              ▼
                         Expand → show chips (no API call)
                              ▼
                         User clicks [生成本段 AI 批语]
                              ▼
                         POST stream with dayun_indexes: [N]
                              ▼
                         SSE emits one item ~10K tokens
```

## Open Questions

- Q: What if the user's chart has fewer than 9 dayuns recorded (e.g., very old or very young命主)?
  - A: The algorithm's `dy.Index` runs 1..N where N is whatever the natural dayun count is. The cutoff logic walks the list as-is.

- Q: Should the page show a small indicator of "spent ¥X on AI" anywhere?
  - A: Out of scope for this proposal. Existing token-usage admin pages cover that need. End-user cost visibility is a separate product question.

- Q: What if the user clicks "生成本段" then immediately navigates away?
  - A: SSE call continues server-side until completion. The dayun row gets cached. Next visit renders it as expanded.
