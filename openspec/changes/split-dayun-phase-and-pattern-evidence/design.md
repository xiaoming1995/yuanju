## Context

`buildDayunRoadmap` currently scores each ten-year period as one aggregate. Its five-element and Ten God helpers inspect both the Dayun stem and branch, then return a single signed number. A helpful stem and an adverse branch therefore become a displayed “中性 0”, despite the existing 金不换 model already treating them as front-five and back-five phases.

The 1995-10-12 noon chart demonstrates the issue: 壬午 has a helpful 壬水 / 七杀 stem and an adverse 午火 / 劫财 branch. Its 偏印格 also has a visible, rooted 偏印 that can be supported by a Dayun 七杀 through a verifiable 七杀生偏印 chain, but the present pattern matcher only recognises preconfigured formation and break Ten Gods.

The backend remains authoritative for deterministic road scores. The frontend is limited to rendering the returned evidence in ordinary and professional views.

## Goals / Non-Goals

**Goals:**

- Preserve stem-led front-five and branch-led back-five contributions instead of hiding them in an aggregate zero.
- Keep one aggregate road score and existing road labels for compatibility.
- Add a strict, testable 偏印格杀印相生 rule that can contribute only when its full chain is present.
- Make professional evidence show phase-specific additions, deductions, and aggregate result clearly.

**Non-Goals:**

- Do not replace 金不换’s established front/back ratings or reclassify its source data.
- Do not create a universal rule that every 七杀 helps every 偏印格.
- Do not change natal vehicle grading, Fuyi/Tiaohou evaluation, routes, or persistence schema.
- Do not let frontend copy recalculate scores.

## Decisions

### Decision 1: Add phase evidence alongside aggregate evidence

`DayunRoad` will retain its existing aggregate `evidences` and gain additive structured `phase_evidences`. Each phase carries a stable phase key (`front` or `back`), phase label, signed subtotal, and its evidence rows. Aggregate rows will continue to provide the ten-year total used for road classification.

This avoids a breaking API change while allowing the UI to show “壬水 +7 / 午火 -7 / 合计 0”. It also makes a future phase-specific road display possible without reinterpreting text.

Alternative considered: replace aggregate evidence with two flat rows. Rejected because API consumers still need the existing decade-level evidence and would have to reconstruct totals.

### Decision 2: Attribute Dayun signals to their governing phase

- Dayun stem five-element and Ten God effects belong to `front`.
- Dayun branch five-element, Ten God, 十二长生, and branch-derived modifiers belong to `back` unless their source is explicitly decade-wide.
- 金不换 retains its current `qian_road` and `hou_road`; its aggregate contribution remains unchanged.
- Road classification uses the same aggregate sum of all deltas, clamped to the existing 0-100 range.

This aligns evidence with the product’s existing “前五年看天干，后五年看地支” semantics without changing the final road taxonomy.

### Decision 3: Split before confidence scaling, then round per phase

Five-element and Ten God evaluators will produce per-component signed deltas before aggregate addition. Ten God confidence weighting is applied independently to each non-zero phase contribution, then rounded and clamped. The aggregate is the sum of these final phase values.

This prevents a helpful and adverse term from cancelling before the code can record either one. It also preserves the current confidence policy: soft confidence remains non-scoring, while medium and hard confidence use their existing relative strength.

### Decision 4: Gate 偏印格的杀印相生 by an explicit chain

The new pattern contribution is allowed only when all of these are true:

1. The natal primary pattern is `偏印格`.
2. Natal 偏印 is visible and rooted, so it is an active structure rather than a hidden background signal.
3. The Dayun stem is `七杀`; it therefore belongs to the front-five phase.
4. The Dayun 七杀 element generates the natal 偏印 element, confirming the 七杀 → 偏印 → 日主 chain.
5. No existing rule has already counted the same interaction as a formation contribution.

The evidence label will identify the chain as `杀印相生` and state its phase. When any gate is absent, the pattern remains neutral with an explicit non-match reason. The rule does not infer a score from a Dayun branch 七杀, hidden-only 偏印, or a generic 偏印格.

Alternative considered: add 七杀 to every 偏印格’s support Ten Gods. Rejected because it would over-score charts that lack an exposed, rooted 印星 or a usable elemental relay.

### Decision 5: Professional UI groups evidence by time phase

The road-evidence modal will render a concise aggregate result followed by `前五年（天干主事）` and `后五年（地支主事）` groups. Each row preserves source, label, signed score, and deterministic explanation. If no phase evidence is available, the UI falls back to the existing flat list.

Ordinary result cards remain concise and keep their current road summary; they do not expose arithmetic by default.

## Risks / Trade-offs

- [Risk] Users may read phase subtotals as independent five-year verdicts. → Mitigation: label them as evidence contributions and retain the full road summary plus 金不换 phase descriptions.
- [Risk] The new chain may double-count a pre-existing pattern formation. → Mitigation: gate the rule against already-recorded formation evidence and test the single-count invariant.
- [Risk] Existing saved results lack `phase_evidences`. → Mitigation: frontend feature-detects the field and renders the current aggregate evidence path unchanged.
- [Risk] More evidence can make the modal verbose. → Mitigation: use two compact phase groups and keep ordinary mode unchanged.

## Migration Plan

1. Add additive phase-evidence DTOs and refactor Dayun road scoring to build per-phase deltas.
2. Implement and test the gated 偏印格杀印相生 branch.
3. Expose the additive JSON field through the existing calculate response.
4. Update frontend types and the professional evidence modal with old-data fallback.
5. Verify the 1995-10-12 noon 壬午 case and representative non-matching cases.

Rollback consists of hiding `phase_evidences` in the frontend and retaining the existing aggregate `evidences`; no stored data migration is required.

## Open Questions

- None for the agreed scope. The proposed rule deliberately covers only the verified 偏印格 + Dayun stem 七杀 chain; other pattern interactions require their own rules and test cases.
