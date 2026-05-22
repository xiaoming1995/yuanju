## Context

`compatibility-actionable-consultation-ux` introduced the personality-first compatibility path: entry preview, personality-fit helper, result-page "性格相处画像", validation plan, and personality-aware history cards. That change answers the right product question, but the current UI still has several dense surfaces:

- Entry page: still presents many controls in one long page, with no lightweight progress state.
- Result page: decision dashboard, evidence summary, personality fit, validation plan, stage validation, strategy, score, AI report, and professional details all appear as separate blocks.
- History page: personality match is present, but score snippets still compete with the relationship label and continuation action.

This change is a frontend-only polish pass. It should refine interaction hierarchy and visual density while preserving all existing request payloads, route behavior, and deterministic fallback logic.

## Goals / Non-Goals

**Goals:**
- Make `/compatibility` feel like a guided consultation while staying on one page.
- Let users understand progress at a glance: selected question, selected relationship stage, and remaining birth-profile input.
- Make `/compatibility/:id` easier to scan by turning personality fit, conflict points, and validation into a single coherent result path.
- Provide result-page anchor navigation so users can jump to personality fit, conflict points/action validation, score/evidence, and professional details.
- Make personality match labels self-explanatory through short descriptive copy.
- Reduce visual density and nested-card impression in personality and validation sections.
- Make `/compatibility/history` cards primarily communicate "who, what personality match type, what question, continue where".

**Non-Goals:**
- No backend API changes.
- No compatibility algorithm, score, evidence taxonomy, or AI prompt changes.
- No new persisted progress state, relationship journal, chat, paid gating, or sharing feature.
- No UI framework, Tailwind, component library, or new dependency.
- No new deterministic relationship prediction language.

## Decisions

1. Keep the entry flow as one page with visual step cues, not a blocking wizard.

   The user should still be able to edit any field at any time. A hard multi-step wizard would make birth data entry feel slower and add route/state complexity. A compact step rail or status row can create guidance without blocking editing.

   Alternative considered: modal or full wizard. Rejected because this app already uses page-level flows, and mobile users need quick access to both birth forms.

2. Merge validation presentation around personality fit.

   The result page currently has both "性格验证计划" and "接下来要验证什么". They should not feel like duplicate modules. The polished hierarchy should place stage-risk details under or near the personality validation area, either through an expandable detail panel or a clear "展开阶段风险" affordance.

   Alternative considered: leave both modules and only rename headings. That reduces wording duplication but not cognitive load.

3. Add top anchor navigation after the decision dashboard.

   Compatibility results are long. A compact anchor bar gives users a map: "性格合不合", "冲突点", "怎么验证", "专业依据". It should be sticky only if it does not conflict with mobile bottom navigation; otherwise it can be a normal horizontal scroll row near the top.

   Alternative considered: floating quick actions. Rejected for now because mobile bottom navigation already competes for vertical space.

4. Explain match types with fixed descriptions.

   Labels such as "高吸引高消耗型" are memorable but incomplete. The UI should display one concise description per type, derived in the existing personality helper, without requiring the user to infer the meaning from later details.

   Alternative considered: ask AI to explain each type. Rejected because deterministic fallback should remain useful without a report.

5. Treat history as recognition, not analysis.

   History cards should not try to replicate the result page. They should optimize for quick recognition and continuation: names, personality match type, stage/question context, and one action. Scores can remain visible but lower priority.

   Alternative considered: adding richer expandable cards. That would increase page complexity and is better suited to a later relationship archive feature.

## Risks / Trade-offs

- [Risk] Adding step cues may make the entry page look more complex. -> Mitigation: keep steps as compact status labels, not large numbered panels.
- [Risk] Merging validation sections may hide useful stage-risk detail. -> Mitigation: retain stage-risk content behind a visible expansion or anchor.
- [Risk] Anchor navigation may add another row of controls on mobile. -> Mitigation: keep labels short and horizontally scrollable, with stable height.
- [Risk] Match-type descriptions may become repetitive. -> Mitigation: keep each description to one short sentence and reuse stable helper output.
- [Risk] History cards may underplay scores for power users. -> Mitigation: keep scores as secondary compact text or a collapsed detail row.

## Migration Plan

1. Add or extend frontend tests that assert entry step cues, result anchor order, merged validation hierarchy, match-type descriptions, and history card priority.
2. Update personality helper metadata for match-type descriptions if needed.
3. Refine `CompatibilityPage` structure and CSS for step/status cues.
4. Refine `CompatibilityResultPage` structure and CSS for anchor navigation, merged validation, lower card density, and mobile layout.
5. Refine `CompatibilityHistoryPage` card layout and CSS for scan-first reading.
6. Run focused compatibility tests, `npm run lint`, `npm run build`, and browser smoke checks for entry/result/history.

Rollback is frontend-only: remove the polish components/styles and return to the existing personality-first layout without changing stored readings.

## Open Questions

- Should the result anchor bar be sticky on desktop only, or non-sticky everywhere? Default recommendation: non-sticky first, then evaluate after browser QA.
- Should stage-risk detail be an accordion inside validation or remain a separate section with a clearer anchor? Default recommendation: accordion/expandable detail under validation.
- Should history scores be hidden on mobile by default? Default recommendation: keep one compact score summary row, not full four-dimension emphasis.
