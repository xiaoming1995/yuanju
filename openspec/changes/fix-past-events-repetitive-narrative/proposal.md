## Why

The current past-events timeline can show nearly identical opening sentences for many consecutive years when each year contains a broad `综合变动` signal. In the 1996-02-08 20:00 test chart, multiple child-age years repeat the same "变化感会比较强..." wording, making the feature feel generic and low-value.

This happens after the recent plain-language narrative pass because `综合变动` is ranked as the dominant theme for every affected year and maps to one fixed sentence. The feature needs the readability improvement without collapsing distinct yearly signals into repeated boilerplate.

## What Changes

- Refine the past-events yearly narrative selector so weak or broad `综合变动` signals do not automatically dominate more specific school, relationship, health, money, movement, or career signals.
- Split broad "change" wording into evidence-sensitive variants for common sources such as dayun-liunian collision, month-pillar interaction, day-branch interaction, kongwang, fuyin/fanyin, and yima.
- Apply a child-age narrative bias for `age < 18`, prioritizing school, peer relationships, family communication, emotional development, and health over abstract adult-style change wording.
- Add regression tests using representative signal sets from the 1996-02-08 20:00 case to ensure consecutive years do not share the same opening sentence unless their dominant evidence is materially the same.
- Preserve `evidence_summary` and all underlying `EventSignal` output; this change only reshapes user-facing yearly prose.

## Capabilities

### New Capabilities
- `past-events-narrative-quality`: User-facing quality rules for past-events yearly narratives, including anti-repetition behavior, dominant-theme selection, and child-age wording.

### Modified Capabilities

## Impact

- Affected backend files:
  - `backend/pkg/bazi/event_narrative.go`
  - `backend/pkg/bazi/event_narrative_test.go`
- No database migration.
- No API shape change beyond existing `narrative` content semantics.
- Frontend structure can remain unchanged unless implementation chooses to add visual differentiation for repeated theme categories.
