## Context

The past-events page currently has a deterministic year-card flow:

```text
BaziResult
  -> GetAllYearSignals
  -> RenderYearNarrative
  -> PastEventsYearItem.narrative
  -> PastEventsPage year card
```

The current `RenderYearNarrative` is intentionally concise. It picks one dominant signal, optionally one secondary signal or ten-god context, then appends a practical reminder. This solved earlier repetition issues, but it now underserves users who expect a year card to feel like an actual reading.

The system already has enough structured data to write richer text without calling an LLM: signal type, polarity, evidence, source, age, dayun phase, ten-god power, and evidence summaries. The main gap is narrative composition.

## Goals / Non-Goals

**Goals:**

- Produce a medium-detail yearly card narrative that usually reads like 120-180 Chinese characters for signal-bearing years.
- Explain the year through concrete life domains such as study, career, money, relationships, health, movement, and family resources.
- Integrate ten-god power in plain language when it clarifies the year.
- Preserve hard event priority for clash, punishment, void, fuyin/fanyin, heavy formations, and yongshen/jishen hits.
- Keep output deterministic, fast, testable, and free of per-year LLM token cost.
- Keep technical evidence in the expandable evidence section instead of exposing dense terminology in the default card body.

**Non-Goals:**

- Do not add per-year LLM calls.
- Do not change the past-events API response shape unless a small additive field becomes necessary during implementation.
- Do not rewrite signal detection rules.
- Do not regenerate or invalidate cached dayun summaries.
- Do not make every year equally long; weak-signal years can remain shorter but should still be more useful than a generic flat sentence.

## Decisions

### Decision 1: Replace sentence concatenation with a narrative frame

Build yearly text from a fixed frame:

```text
1. Year tone
2. Trigger source
3. Life-domain manifestation
4. Ten-god force explanation
5. Practical stance
```

The frame gives users enough context while keeping cards scannable. Each part should be generated from structured signals, not from random phrase pools.

Alternative considered: simply lengthen existing `plainThemeSentence` outputs. That is lower effort but would still repeat when adjacent years share the same dominant theme.

### Decision 2: Keep hard evidence dominant, but make it readable

Hard signals should still lead the reading, but the default body should translate them:

- clash/punishment -> pressure, conflict, safety, relationship or environment movement
- void -> plans feel uncertain, details need confirmation
- fuyin -> repeated themes or old issues returning
- fanyin -> visible change and disruption
- heavy formation -> amplified pressure or opportunity
- yongshen/jishen hit -> support or pressure source becomes obvious

The technical cause remains in `RenderEvidenceSummary`.

Alternative considered: expose evidence strings directly in the card body. This helps professional users but recreates the earlier problem where ordinary users see dense terminology.

### Decision 3: Use domain-specific detail builders

Each domain should have a richer builder:

| Domain | Detail to surface |
| --- | --- |
| study | exams, teachers, methods, peers, certificates, concentration |
| career | duties, role changes, recognition, cooperation, pressure |
| money | resources, spending, family support, investment caution |
| relationship | communication, boundaries, emotional decisions, family atmosphere |
| health | rest, stress, accident avoidance, routines |
| movement | travel, relocation, school/work environment changes |
| support | helpers, elder support, protective resources |
| change | repeated issues, sudden adjustment, plans needing confirmation |

Age should continue to matter. For age under 18, adult career and romance wording should be converted to school, family, peers, and daily rhythm.

Alternative considered: a single generic detail paragraph for all themes. It would be easier but would not solve the user's complaint that the cards do not feel detailed.

### Decision 4: Use ten-god power as explanation, not a second repeated paragraph

Ten-god power should answer "what force is driving this year" in plain language. It should be merged into the body only when it adds information that is not already obvious from the dominant signal.

Examples:

- `钱财资源明显` -> resources, spending, practical commitments
- `规则压力明显` -> rules, exams, responsibility, external constraints
- `学习贵人明显` -> teachers, credentials, elders, protective support
- `才华表达明显` -> expression, output, performance, creativity
- `同辈竞争明显` -> peers, competition, cooperation, comparison

Alternative considered: always show a standalone ten-god sentence. That can repeat across adjacent years and make the card feel formulaic.

### Decision 5: Keep frontend changes minimal

The default card can hold a longer paragraph if spacing and line height stay comfortable. Frontend changes should be limited to readability if needed: max width, line height, muted force row, and mobile spacing.

Alternative considered: add a collapsed "read more" interaction in this change. That may be useful later, but the first improvement should prove that the default text itself is worth reading.

## Risks / Trade-offs

- Longer text may make the timeline visually heavier -> Keep the target length moderate and verify mobile spacing.
- Rule-generated text can still become repetitive -> Add tests for adjacent years with different signals and avoid exact same opening/detail patterns.
- More builder logic can become hard to maintain -> Keep helpers local to `event_narrative.go` and test by behavior, not every exact phrase.
- Some weak-signal years may not have enough evidence for 120+ characters -> Allow shorter fallback text, but avoid empty generic filler.

## Migration Plan

1. Add tests that demonstrate current narratives are too shallow for representative signal-bearing years.
2. Introduce a richer narrative frame in `event_narrative.go`.
3. Add domain-specific detail helpers and ten-god integration helpers.
4. Verify existing hard-evidence priority tests still pass.
5. Adjust frontend spacing only if longer text becomes visually dense.
6. Rollback by restoring the previous `RenderYearNarrative` composition; no database rollback is required.

## Open Questions

- Should weak-signal years target a shorter range, such as 70-110 characters, instead of forcing 120-180 characters?
- Should the UI later add an explicit "展开详批" mode after the richer default text is validated?
