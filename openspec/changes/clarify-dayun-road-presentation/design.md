## Context

`DayunRoad` already carries two distinct kinds of result. `road_label` is a ten-year composite score calculated from Jin Bu Huan, five elements, pattern effects, Ten Gods, branch 十二长生, and Shen Sha modifiers. `qian_road` and `hou_road` are only Jin Bu Huan's stem-led and branch-led ratings.

The current result overview and Dayun timeline render all three as road labels. The backend summary also describes the front/back labels as an overall road condition. This makes a composite `泥路` displayed beside two `施工路段` labels look contradictory even though the calculations have different scopes. The change must improve comprehension without changing the existing scoring model, API fields, or Dayun interactions.

## Goals / Non-Goals

**Goals:**
- Make it obvious that the composite road is the authoritative ten-year conclusion.
- Present front-five and back-five Jin Bu Huan ratings as phase prompts, with time range and governing stem/branch.
- Give ordinary users concise, plain-language guidance before technical evidence.
- Keep professional evidence reachable without making the default timeline dense.
- Preserve the existing desktop ten-card strip and mobile two-column timeline behavior.

**Non-Goals:**
- Do not alter road-score thresholds, Jin Bu Huan source rules, or natal-strength logic.
- Do not calculate new phase-composite road scores in this change.
- Do not remove Liunian drill-down, Shen Sha annotations, current-period detection, or old-result fallbacks.
- Do not add a database migration or break the current response contract.

## Decisions

### Decision 1: Treat the ten-year road and phase prompts as separate concepts

The active Dayun card and timeline summary will use `十年综合路况` only for `road_label`. The `qian_road` and `hou_road` values will be labelled `金不换阶段提示`, followed by the phase rating (`吉`/`平`/`凶`) and a short scope description rather than a second road grade.

This is preferable to keeping two comparable road labels because the phase values are not calculated from the complete phase evidence. A future change can add true phase-composite scores without changing this presentation contract.

### Decision 2: Use a consistent three-layer active-period layout

Both the overview current-road card and selected timeline summary will use the same order:

1. Period identity: Gan Zhi, age range, Gregorian range, and current status.
2. Composite conclusion: `十年综合路况`, its road label, and one plain-language sentence from the road guide.
3. Stage prompts: front-five and back-five panels identifying the time range, `天干主事` or `地支主事`, Jin Bu Huan `吉/平/凶`, and existing phase detail when available.

The overview can link to an explanation modal and professional evidence. The timeline remains the navigational surface and keeps its expanded technical summary secondary.

### Decision 3: Clarify backend-generated summary copy without changing scores

`dayunRoadSummary` will name the composite result as `十年综合路况` and separately state that front and back are Jin Bu Huan prompts. It must not call a joined phase label the overall condition.

This preserves all existing JSON fields and lets cached or external consumers receive unambiguous text.

### Decision 4: Keep the timeline strip scannable

Each Dayun step card retains age range, Gan Zhi, Gan/Zhi Ten Gods, and the composite road badge. Front/back information moves to the selected-period detail instead of expanding every timeline card. Shen Sha chips remain available but do not displace the composite road badge.

Alternative considered: show front/back labels in every step card. Rejected because it makes nine or ten cards visually noisy and repeats phase content before a user selects a period.

## Risks / Trade-offs

- [Risk] Users expect a road label for each five-year phase. -> Mitigation: explicitly name these values as Jin Bu Huan prompts and reserve road labels for a full composite score.
- [Risk] Extra explanatory copy makes compact surfaces verbose. -> Mitigation: show one sentence by default, use concise phase panels, and keep detailed arithmetic in the existing professional evidence path.
- [Risk] Older saved results lack phase detail. -> Mitigation: render the phase rating and timing when present, otherwise retain the current-road fallback without an empty panel.
- [Risk] Desktop and mobile layouts diverge. -> Mitigation: use shared semantic components and verify the fixed desktop strip plus mobile two-column layout at runtime.

## Migration Plan

1. Add a small frontend presentation helper that derives phase time ranges and display labels from existing Dayun data.
2. Refactor the overview current-road card and Dayun selected-period summary to use the shared hierarchy.
3. Clarify the backend roadmap summary sentence and add focused regression coverage.
4. Update static tests and verify desktop and mobile renderings with representative road and phase combinations.

Rollback consists of restoring the prior frontend rendering and summary sentence; no persisted data or schema must be migrated.

## Open Questions

- None. This change deliberately distinguishes scopes rather than introducing a new phase-composite scoring model.
