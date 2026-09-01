## Context

The backend already calculates `NatalAssessment.Fuyi` from all four natal pillars, including visible stems, hidden stems, month command, roots, Jianlu, and Yangren. `DayunTimeline` currently receives the Fuyi favorable/adverse elements but not `day_master_strength`; `dayunOverview.ts` instead derives a second strong/weak value from the flat `wuxing` percentages. That local result selects strength-specific Ten-God wording, so it can contradict the complete natal assessment.

The road evaluator keeps Dayun stem and branch effects in separate five-year phases, but its Twelve Growth Stage contribution is currently a generic signed score. A branch at 长生 adds a positive score even if it carries a natal Fuyi adverse element, while a branch at 衰 subtracts a score even if it carries a favorable element.

## Goals / Non-Goals

**Goals:**
- Use the complete natal Fuyi assessment as the sole Day Master strength source for all Dayun rendering.
- Keep the natal strength conclusion constant while selecting different Dayun periods.
- Evaluate the Dayun stem and branch independently against natal Fuyi favorable/adverse elements.
- Use the branch's Twelve Growth Stage to express and score the intensity of its own Fuyi effect.
- Keep frontend prose concise while making stem, branch, and branch-stage reasoning auditable.

**Non-Goals:**
- Do not recalculate natal Fuyi strength for each Dayun, Liunian, or transient interaction.
- Do not change the underlying lunar calendar, Four Pillars, Dayun ordering, 金不换 rules, or natal grade algorithm.
- Do not treat Twelve Growth Stages as a standalone good/bad verdict or a universal Day Master strength adjustment.
- Do not change the existing calculate API route or require a database migration.

## Decisions

### Decision 1: Natal Fuyi is the authoritative strength boundary

`natal_assessment.fuyi.day_master_strength`, along with its favorable and adverse elements, is calculated once from the natal chart and passed to Dayun consumers. The frontend will remove the percentage-based `resolveDayStrength` helper and any copy selection that relies on it. A missing natal assessment will render a neutral, non-assertive fallback rather than infer a strong/weak result locally.

Alternative considered: improve the frontend percentage threshold. Rejected because it still has no access to hidden stems, month command, roots, Jianlu, Yangren, or the versioned backend rule set.

### Decision 2: Dayun prose separates stem and branch effects

The overview will evaluate Dayun Gan and Zhi against the natal Fuyi favorable/adverse element sets independently. Its prose will name the stem's Ten-God/Fuyi role and the branch's Ten-God/Fuyi role; it will never use the branch or its 十二长生 to state that the natal chart is strong or weak.

For example, a natal-strong 壬日主 in 戊申运 can state that 戊土七杀 is a favorable regulating force while 申金偏印 is adverse; it must not produce “身弱遭杀克身”.

Alternative considered: retain a single Ten-God sentence based on the Dayun stem. Rejected because it hides the branch's opposite polarity and caused the current misleading summary.

### Decision 3: 十二长生 modifies branch polarity, not the Day Master

The road evaluator will derive the branch's Fuyi polarity first, then apply the 十二长生 bucket as an intensity modifier:

- 旺势 stages (`帝旺`、`临官`、`长生`、`冠带`) amplify a favorable branch's positive contribution or an adverse branch's negative contribution.
- 中势 stages (`沐浴`、`养`、`胎`、`墓`) apply a smaller same-direction adjustment.
- 弱势 stages (`衰`、`病`、`死`、`绝`) suppress that branch's contribution rather than reversing its polarity.
- A neutral branch receives explanatory text only and no directional 十二长生 score.

The evidence label and detail will explicitly identify the branch, its Fuyi polarity, its 十二长生 stage, and the resulting strength. This attaches the calculation to the branch's carried favorable/adverse force and preserves the existing front-five/back-five evidence boundary.

Alternative considered: retain the current generic `+7/+2/-6` stage score. Rejected because a favorable or adverse result must depend on what the branch is strengthening.

### Decision 4: Preserve additive response compatibility

Existing `natal_assessment` and `dayun_roadmap` response fields remain in place. Backend road evidence retains the `十二长生` source but updates its direction and explanatory detail. The frontend accepts the authoritative natal-strength field as optional so older saved responses remain readable without an incorrect local inference.

## Risks / Trade-offs

- [Risk] Existing saved result payloads may lack a current natal assessment. → Mitigation: render polarity-only neutral fallback copy and do not assert body strength.
- [Risk] A branch's element and Ten God can carry different user-facing descriptions. → Mitigation: use the Fuyi element to determine direction and include the Ten God only as a label.
- [Risk] Stage scoring can double-count the branch's elemental role. → Mitigation: classify it as a bounded intensity modifier, show it separately in back-five evidence, and test the aggregate.
- [Risk] Copy changes may make Dayun summaries longer. → Mitigation: retain one concise sentence in ordinary mode and keep detailed evidence in professional mode.

## Migration Plan

1. Add backend tests for immutable natal strength and polarity-aware branch-stage road evidence.
2. Replace generic 十二长生 road scoring with branch-polarity-aware evidence.
3. Pass the natal Fuyi strength through `ResultPage` to `DayunTimeline` and remove the local frontend strength heuristic.
4. Update overview copy and frontend tests for favorable, adverse, neutral, and legacy-assessment cases.
5. Run targeted Go tests, Dayun overview tests, static UI tests, and the production frontend build.

Rollback consists of reverting the new evaluator and frontend field wiring; no stored data is migrated or destructively transformed.

## Open Questions

- None. The agreed rule is that natal strength is original-chart-only and 十二长生 modifies only the Dayun branch's favorable/adverse force.
