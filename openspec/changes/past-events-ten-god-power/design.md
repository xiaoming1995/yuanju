## Context

The past-events module now has two layers:

- `GeneratePastEventsYears` returns all liunian cards from rule-based signals and templates.
- `GenerateDayunSummariesStream` adds cached AI summaries per dayun.

The underlying bazi engine already calculates ten-god labels for dayun and liunian:

- dayun heavenly stem and earthly branch main qi: `GanShiShen`, `ZhiShiShen`
- liunian heavenly stem and earthly branch main qi: `GanShiShen`, `ZhiShiShen`

However, the past-events API only returns event tags, narrative, evidence, dayun phase, and dayun metadata. It does not expose a structured ten-god force profile. The signal engine uses ten gods internally, but the result is scattered across evidence text rather than represented as a first-class user-facing concept.

## Goals / Non-Goals

**Goals:**

- Make each dayun and liunian carry a structured ten-god power profile.
- Translate ten-god force into plain user-facing meaning, such as money/resources, rules/responsibility, study/support, expression/output, and peers/competition.
- Use strength labels instead of raw scores in the default UI, while preserving score/reason data for evidence and tests.
- Let ten-god power guide narratives and dayun summaries without overriding stronger event evidence.
- Keep the implementation rule-based for yearly output; do not add per-year LLM calls.

**Non-Goals:**

- Do not implement cross-chart cache reuse.
- Do not introduce a database migration.
- Do not build a full professional ten-god dashboard in this change.
- Do not change the base bazi calculation API unless the past-events flow needs additive fields.
- Do not invalidate existing `ai_dayun_summaries` automatically.

## Decisions

### Decision 1: Introduce a dedicated TenGodPowerProfile value object

Add a reusable structure in the bazi package rather than embedding one-off strings in service code.

Suggested shape:

```go
type TenGodPowerProfile struct {
    Dominant    string `json:"dominant"`       // 正财 / 七杀 / 正印 ...
    Group       string `json:"group"`          // wealth / official / seal / output / peer
    GroupLabel  string `json:"group_label"`    // 财星 / 官杀 / 印星 / 食伤 / 比劫
    Strength    string `json:"strength"`       // weak / medium / strong / very_strong
    Polarity    string `json:"polarity"`       // support / pressure / mixed
    PlainTitle  string `json:"plain_title"`    // 财星偏旺
    PlainText   string `json:"plain_text"`     // 资源、钱财、现实压力更明显
    Score       int    `json:"score"`
    Reason      string `json:"reason,omitempty"`
}
```

Rationale: the same profile can be used by API response, narrative selection, tests, evidence expansion, and dayun summary prompt payload. Keeping `Score` internal but serialized allows professional debugging without forcing the UI to display it.

Alternative considered: add only `gan_shishen` and `zhi_shishen` to the API. This is too shallow and would push interpretation burden to users.

### Decision 2: Score dayun and liunian with different weighting rules

Dayun score should reflect 10-year background force:

```text
dayun stem ten-god       +4
dayun branch main qi     +3
current front/back phase +2 to active side
yongshen/jishen fit      +/-2
natal strength context   +/-1 for pressure/support interpretation
```

Liunian score should reflect yearly trigger force:

```text
liunian stem ten-god        +4
liunian branch main qi      +2
same group as dayun profile +2
same exact ten-god as dayun +1
yongshen/jishen fit         +/-2
natal strength context      +/-1 for pressure/support interpretation
```

Rationale: the heavenly stem is more visible and event-triggering, while the earthly branch is more persistent and background-oriented. This aligns with the existing dayun front-five/heavenly-stem and back-five/earthly-branch phase model.

Alternative considered: use equal stem/branch weights. This is simpler but misses the current phase logic and makes profiles feel flat.

### Decision 3: Group ten gods before explaining them

Map raw ten gods into five practical groups:

| Group | Ten gods | Plain meaning |
| --- | --- | --- |
| wealth | 正财 / 偏财 | money, resources, practical pressure, relationship materiality |
| official | 正官 / 七杀 | rules, responsibility, tests, status, pressure |
| seal | 正印 / 偏印 | study, protection, credentials, elders, support |
| output | 食神 / 伤官 | expression, talent, creation, rebellion, output |
| peer | 比肩 / 劫财 | peers, competition, cooperation, sharing, rivalry |

Rationale: users can understand force groups faster than ten individual technical labels. The exact ten-god remains available for professional evidence and short labels.

### Decision 4: Ten-god force informs narrative but does not dominate hard event signals

Ten-god power should be used as a background theme unless it aligns with concrete event signals.

Priority order:

1. Layer 0 hard evidence: clashes, punishments, void, use-god/jishen position hits, dayun-liunian double hit.
2. Concrete domain signals: career, money, relationship, health, migration, study.
3. Ten-god power profile as tone and explanation.
4. Generic comprehensive change.

Rationale: a year with strong official/killing force can mean pressure or status, but if the same year has a severe clash to a key palace, that hard signal must stay dominant.

### Decision 5: Frontend displays one small force row, not a technical panel

Dayun header should show a concise force label such as `主导：财星偏旺`. Year cards should show one short line after the title or before the narrative:

```text
年度力量：官杀偏旺 - 规则、考核、责任感更明显
```

The expandable evidence section can include the technical reason and score. Default card content must stay readable for non-professionals.

Alternative considered: add a full ten-god breakdown table per card. This would increase visual noise and recreate the current terminology problem.

### Decision 6: Dayun summary cache remains valid but may be stale

The existing `ai_dayun_summaries` cache should not be auto-invalidated. The past-events page can still show rule-based ten-god profile immediately, while cached AI summaries may not mention the new force profile until regenerated.

Rationale: automatic invalidation would cause unexpected token usage after deployment. A future explicit refresh path can regenerate summaries when desired.

## Risks / Trade-offs

- Score formula may feel arbitrary -> Mitigate with focused tests for representative ten-god groups and transparent `Reason` text.
- Too much technical detail may clutter the UI -> Mitigate by showing only `PlainTitle` and `PlainText` by default.
- Cached AI dayun summaries may not reference ten-god force -> Mitigate by keeping rule-based labels visible outside the AI summary and documenting optional regeneration.
- API response grows slightly -> Mitigate by using compact structs and only adding fields to the past-events endpoint.
- Polarity can be nuanced by natal strength -> Mitigate by using `support`, `pressure`, and `mixed` instead of forcing every profile into good/bad.

## Migration Plan

1. Add ten-god grouping and profile helpers in the bazi package.
2. Extend dayun metadata and yearly event response with optional ten-god power profiles.
3. Feed year-level profiles into `RenderYearNarrative` as tone/context.
4. Include profiles in the dayun summary prompt payload for future regenerated summaries.
5. Update frontend types and compact display.
6. Add tests for scoring, API response shape, and narrative behavior.

Rollback: remove the additive fields and frontend display. No database rollback is required.

## Open Questions

- Should the UI show exact ten-god names by default, or only force group labels such as 财星/官杀?
- Should there be a user-facing "refresh dayun summaries" action to regenerate cached summaries with ten-god context?
