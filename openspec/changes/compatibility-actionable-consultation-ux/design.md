## Context

The existing compatibility flow has already moved beyond a raw score report. `compatibility-consulting-ux-v2` adds controlled relationship context, question-aware reports, and a decision-first result hierarchy. The stronger product insight is that compatibility users first care about personality fit: "两个人性格合不合、相处顺不顺、为什么互相吸引又为什么冲突". The remaining UX gap is that this personality-fit answer is not yet the primary reading path.

This change should be a frontend-first iteration. It should reuse existing compatibility context fields, dimension scores, evidence, consulting assessment data, duration/stage risk data, and latest report state. Backend or database work should only be introduced if implementation proves existing detail/history responses cannot support the personality/archive UI.

## Goals / Non-Goals

**Goals:**
- Make the compatibility entry flow feel like a lightweight personality-fit consultation before asking for birth data.
- Give users a clear preview that the reading will explain both people's relationship personality, fit points, and conflict points.
- Make the result page lead with a "性格相处画像" before generic scores, timing windows, and professional evidence.
- Add short-term do/avoid guidance that validates the personality-fit judgment using existing decision advice, stage risks, and duration assessment.
- Improve compatibility history so users can resume from prior readings and recognize each relationship's personality match type.
- Keep all changes compatible with existing readings and current API shapes where possible.

**Non-Goals:**
- No new compatibility algorithm, score model, or evidence taxonomy.
- No new paid gating, relationship journal, chat thread, advisor marketplace, or sharing system.
- No exact timing predictions for marriage, breakup, reconciliation, pregnancy, affairs, or other deterministic events.
- No free-text relationship story collection in this iteration.
- No UI framework or styling dependency changes.

## Decisions

1. Treat the entry page as a personality-fit consultation path.

   The page should guide users through: primary fit question, relationship stage, and both birth profiles. The default framing should be "性格合不合", with other questions such as long-term stability or recurring conflict presented as related angles. Birth data remains required, but it becomes the final input needed to build the personality-fit reading.

   Alternative considered: keep the current form order and add explanatory copy. That would be less disruptive, but it would not change the user's first impression from "form" to "consultation".

2. Derive a personality-fit model on the frontend from existing data.

   A helper can combine dimension scores, relationship pattern evidence, decision diagnosis, top findings, and report fields into stable blocks: overall personality fit, self relationship pattern, partner relationship pattern, fit points, conflict points, and communication guidance. The wording should stay tendency-based and avoid pretending to infer clinical personality traits.

   Alternative considered: adding new backend personality fields. That may be valuable later, but the current evidence and structured report already contain enough material to validate the UX first.

3. Derive action plans as validation of the personality judgment.

   A helper can combine `primary_question`, `relationship_stage`, `decision_advice`, `stage_risks`, and `duration_assessment` into stable action-plan blocks such as 7-day checks, 30-day checks, and do-not-do items. These checks should answer: "does real interaction confirm the personality-fit judgment?"

   Alternative considered: persisting a generated action plan in the backend. That creates report/schema coupling before the product shape is proven.

4. Keep result-page advice conditional and evidence-adjacent.

   The action plan should use verification language: "observe", "confirm", "avoid increasing investment until...", not absolute commands. It should sit after the decision dashboard and before deeper score/evidence content, with links down to stage risks or evidence where relevant.

   Alternative considered: adding another large dashboard above everything. The result page already has a decision dashboard; this change should strengthen actionability, not create duplicated hierarchy.

5. Make history cards support personality-match recognition and continuation.

   History should show the selected question, stage, personality match type, latest decision state, whether a deep report exists, and one continuation action such as "view personality fit", "continue verification", or "generate deep reading". This uses available reading/report metadata rather than collecting new user progress state.

   Alternative considered: adding checkboxes or saved progress. That likely needs new persistence and is better suited for a later relationship journal feature.

6. Verification should focus on structure, copy, and route continuity.

   Static frontend tests should assert entry-flow order, personality-fit helper output, action-plan helper output, result-section order, history continuation affordances, and mobile CSS hooks. Browser smoke should cover creating/opening a compatibility flow when local backend data is available.

## Risks / Trade-offs

- [Risk] The flow may feel longer because personality context appears before birth data. -> Mitigation: keep controls compact and preserve one-page flow rather than introducing a blocking wizard.
- [Risk] Personality copy may overstate what bazi evidence can prove. -> Mitigation: use relationship-pattern language, explain evidence links, and avoid clinical or absolute personality labels.
- [Risk] Action plans may feel generic if report data is absent. -> Mitigation: tie actions to fit/clash points and selected question, and clearly label generated report as a deeper layer.
- [Risk] Users may interpret guidance as absolute relationship instruction. -> Mitigation: use validation/check language, keep uncertainty copy, and avoid fate-like phrasing.
- [Risk] History continuation may overpromise without persisted progress. -> Mitigation: phrase as "continue reading/verification" rather than "resume completed task".
- [Risk] Current history API may not expose enough report state. -> Mitigation: first use existing fields; if blocked, update the design before adding backend fields.

## Migration Plan

1. Add frontend personality-fit and action-plan derivation helpers with tests.
2. Reorder and restyle the compatibility input page into a personality-first consultation layout.
3. Add personality-fit section to the result page using existing reading/report/evidence data.
4. Add action-plan section to the result page as validation of the personality-fit judgment.
5. Update history cards with personality match type and continuation labels/actions from available metadata.
6. Run targeted frontend UX tests, lint, build, and browser smoke.

Rollback is frontend-only if no API changes are introduced: remove the new helper, revert page layout changes, and keep existing compatibility reading data untouched.

## Open Questions

- Should the entry flow remain one page or use a visible stepper? Default recommendation: one page with section order and progress cues, avoiding a hard wizard.
- What top-level personality match types should be displayed? Default recommendation: "稳定互补型", "高吸引高消耗型", "慢热磨合型", "反复拉扯型", and "现实压力型".
- Should "7 days" and "30 days" be fixed labels or adapt to stage risk windows? Default recommendation: fixed user-friendly labels mapped from existing stage risk data.
- Should history expose "generate deep reading" directly on cards if no report exists? This depends on whether the current history response includes enough report presence data.
