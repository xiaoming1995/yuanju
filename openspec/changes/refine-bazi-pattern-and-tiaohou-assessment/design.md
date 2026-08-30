## Context

`assessNatalPattern` currently uses a single `has` helper that treats every visible and hidden ten-god occurrence as equivalent. A hidden 食神 can therefore trigger `枭神夺食`, and the ordinary co-presence of 偏印与比劫 can add an `印比相扶` formation score. Separately, `calcTiaohou` already calculates the day-stem/month-command dictionary result, but `NatalAssessment.Climate` only reports the seasonal extreme heat/cold gate. The result page describes the latter with the generic word “调候”, hiding the former.

## Goals / Non-Goals

**Goals:**
- Require visible and rooted evidence for a decisive pattern formation or break.
- Keep general Fuyi support out of the pattern-formation score.
- Return and display day-stem Tiaohou and seasonal thermal Tiaohou as separate assessments.
- Treat a Tiaohou element hidden in a branch as partial support, not as absence and not as a full resolution.
- Lazily refresh saved natal assessments to the corrected rules.

**Non-Goals:**
- Rebuild the 120-entry day-stem Tiaohou dictionary.
- Encode every traditional special pattern or replace professional review.
- Change the independent global Fuyi strength algorithm.

## Decisions

### Evidence tiers for pattern interactions

Introduce explicit visible, hidden, and rooted checks on the existing ten-god statistics. A pattern interaction that changes `NatalPatternAssessment.Score` requires the primary relevant ten god to be visible; when the interaction is a decisive break, the affected star must also have branch root support. Hidden-only stars remain available as explanatory context but cannot independently create a formation or break.

For a 偏印格, `枭神夺食` requires 食神透干 and effective root support. `印比相扶` is removed from the pattern-formation set because it describes daily-master support already measured by Fuyi, rather than a制化 relationship of 偏印格.

This is preferred over globally ignoring hidden stems: hidden stems remain important for roots, day-stem Tiaohou availability, Fuyi, and weaker context.

### Two Tiaohou dimensions

Extend the natal assessment with two named dimensions:

1. **Day-stem Tiaohou** reads the existing `dayGan + monthZhi` dictionary, records each required stem and whether it is visible or hidden.
2. **Thermal Tiaohou** evaluates the seasonal cold/heat condition separately, then records the required element's visible and hidden support. The climate condition is determined before checking remedies, so hidden support can only produce `partial`, never erase the condition.

For `丙戌`, the day-stem result records `甲` as visible and `壬` as hidden. The thermal result separately describes the seasonal state and water support such as 壬癸藏干 or water branches.

### Downstream and display semantics

Only an unresolved urgent **thermal** Tiaohou result can impose the existing grade ceiling. Day-stem Tiaohou is shown as a use-god availability assessment and is injected into professional AI context, but it does not overwrite Fuyi favorable elements. Result-page labels explicitly identify each dimension.

### Versioned lazy refresh

Bump `NatalAssessmentVersion`. Existing result snapshots are rebuilt on read, then their vehicle profile, road map, favorable/adverse Shishen, and prompt context use the refreshed assessment.

## Risks / Trade-offs

- [Some existing combination labels disappear from historical charts] → Preserve hidden-star facts as low-weight evidence and provide targeted regression tests.
- [Traditional schools differ on the root threshold for 枭神夺食] → Isolate the threshold in a helper and document the current conservative rule.
- [More Tiaohou detail can confuse ordinary users] → Use concise two-line labels in simple mode and show full evidence only in professional mode.

## Migration Plan

1. Add tiered pattern helpers and the two-dimensional Tiaohou data model.
2. Bump the natal-assessment version and refresh saved snapshots lazily.
3. Verify `乙亥·丙戌·丙子·甲午`: no 枭神夺食 or 印比相扶 formation; day-stem Tiaohou shows 甲透、壬藏; thermal Tiaohou records water support without conflating it with day-stem Tiaohou.
4. Roll back by restoring the prior version constant and evaluator if a rule requires revision; no database migration is needed.

## Open Questions

- The precise seasonal dry/wet vocabulary and thresholds can be expanded after collecting additional agreed example charts.
