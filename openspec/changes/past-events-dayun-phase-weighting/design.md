## Context

`Calculate()` already builds each `DayunItem` with 10 `LiuNianItem`s and a `JinBuHuanResult`. The JinBuHuan result explicitly models "front five years by dayun heavenly stem" and "back five years by dayun earthly branch":

- `qian_level/qian_desc`: front five-year rating from dayun gan.
- `hou_level/hou_desc`: back five-year rating from dayun zhi.

The past-events yearly engine does not currently use this split. `GetAllYearSignals()` passes only the dayun gan-zhi string into `GetYearEventSignals()`, so every year inside a dayun uses the same dayun gan and dayun zhi at the same strength.

## Goals / Non-Goals

**Goals:**
- Make each liunian aware of its position within the containing dayun.
- Let years 1-5 lean more on dayun gan context and years 6-10 lean more on dayun zhi context.
- Feed JinBuHuan front/back rating into yearly narrative and dayun summary generation as background context.
- Preserve existing direct liunian triggers and readable plain-language output.

**Non-Goals:**
- Replacing lunar-go dayun/liunian generation.
- Rewriting JinBuHuan rules or dictionaries.
- Making every event type depend on dayun phase.
- Removing future-year output from the current timeline.
- Introducing per-year LLM calls.

## Decisions

### Decision 1: Compute phase from the liunian index inside each dayun

The phase should be derived while iterating `dy.LiuNian`:

```text
index 0-4 -> gan phase
index 5-9 -> zhi phase
```

This is more reliable than age arithmetic because some dayun start ages and transition labels come from lunar-go and can vary around boundary years.

### Decision 2: Extend internal signal context instead of changing the public contract first

Introduce a small internal context structure, for example:

```go
type YearSignalContext struct {
    DayunPhase string // gan | zhi
    YearInDayun int  // 1-10
    JinBuHuanLevel string
    JinBuHuanDesc string
}
```

Then add a contextual entrypoint such as `GetYearEventSignalsWithContext(...)` while keeping the existing `GetYearEventSignals(...)` as a compatibility wrapper.

The service can then pass phase context when it has the full `DayunItem`, while existing tests and callers continue to work.

### Decision 3: Phase weighting should modify priority and background, not delete signals

All yearly triggers remain valid in every year:

- liunian gan ten-god signal
- liunian zhi relationships with natal pillars
- dayun gan and dayun zhi relationships
- health, movement, fuyin/fanyin, shensha, kongwang

The phase only changes emphasis:

- `gan` phase should boost dayun gan related evidence such as dayun ten-god overlays, gan combination, and qian JinBuHuan rating.
- `zhi` phase should boost dayun zhi related evidence such as dayun branch clash/combination with natal branches, empty branch, sanhe/sanhui context, and hou JinBuHuan rating.

### Decision 4: Add a phase background signal

Add an internal or visible signal type such as `大运阶段` when phase context exists:

- `gan` phase evidence references `qian_desc`.
- `zhi` phase evidence references `hou_desc`.
- polarity is mapped from JinBuHuan level:
  - `吉` -> `PolarityJi`
  - `凶` -> `PolarityXiong`
  - otherwise `PolarityNeutral`

This gives both `RenderEvidenceSummary()` and the AI dayun summary prompt a stable way to see the front/back five-year quality.

### Decision 5: Narrative selection should treat phase context as support, not always dominant

`大运阶段` should not overpower specific life events by default. It should appear as:

- a secondary sentence when no stronger specific theme exists;
- a practical reminder tone when the phase is clearly auspicious or adverse;
- a supporting evidence item under "命理依据".

Strong yearly triggers such as fuyin, fanyin, health risk, major clash, or child-age school pressure should still dominate the yearly card.

### Decision 6: Dayun summary prompt should include phase split

`GenerateDayunSummariesStream` should include each year's phase and phase rating in the JSON given to the AI. The prompt should ask the model to summarize the early five years and late five years when their tones differ, without forcing a rigid two-paragraph output.

## Data Flow

```text
Calculate()
  -> DayunItem{Gan, Zhi, JinBuHuan, LiuNian[10]}
      -> GetAllYearSignals()
          -> determine year_in_dayun + phase
          -> select qian/hou JinBuHuan desc
          -> GetYearEventSignalsWithContext(...)
              -> base yearly signals
              -> phase background signal
              -> phase-aware priority metadata
          -> RenderYearNarrative()
          -> RenderEvidenceSummary()
```

## Testing Strategy

- Unit test phase derivation:
  - first five liunian items map to `gan`;
  - last five map to `zhi`.
- Unit test JinBuHuan phase signal:
  - qian rating is used only in years 1-5;
  - hou rating is used only in years 6-10.
- Unit test narrative behavior:
  - a specific yearly event remains dominant over neutral phase context;
  - an adverse hou phase can appear as secondary caution in years 6-10;
  - legacy `GetYearEventSignals` callers still compile and return prior-style signals when no context is provided.
- Service test or focused integration test:
  - `GeneratePastEventsYears` returns grouped years whose phase metadata matches each dayun position if additive fields are exposed.

## Risks / Trade-offs

- If phase context is ranked too high, narratives can become repetitive again. Mitigation: keep `大运阶段` secondary unless it is the only meaningful signal.
- Adding context parameters directly to `GetYearEventSignals` would cause broad test churn. Mitigation: add a contextual wrapper and keep the old function.
- JinBuHuan rules are broad decade-level judgments, not exact yearly events. Mitigation: phrase them as background tone rather than concrete life events.
