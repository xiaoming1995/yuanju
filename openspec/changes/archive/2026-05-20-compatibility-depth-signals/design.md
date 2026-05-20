## Context

The compatibility product currently has a deterministic bazi engine, persisted reading/evidence records, AI report generation, and a result page. The active `compatibility-consulting-ux-v2` change improves product framing by adding relationship stage, primary question, question-aware reports, and a decision-first result hierarchy.

This change focuses on the next layer: the deterministic evidence behind those consultation claims. Today `AnalyzeCompatibility` starts from four baseline scores and adds evidence for day-master relation, five-element complement/imbalance, spouse-palace day-branch interaction, spouse-star resonance, branch clash count, and selected shensha. That is useful, but it does not yet explain how each person experiences the other, whether one chart supports the other's structural needs, or which specific gan-zhi interactions create attraction, pressure, conflict, or repairability.

Constraints:

- Build on the existing Go bazi package and compatibility model shape.
- Preserve the four public score dimensions: attraction, stability, communication, practicality.
- Keep the consulting UX change as the product-context layer; this change must not duplicate relationship-stage input work.
- Avoid exact-event predictions and deterministic relationship claims.
- Keep evidence explainable enough for both ordinary users and professional users.

## Goals / Non-Goals

**Goals:**

- Add deeper deterministic compatibility evidence using ten-god relationship, favorable-element support, expanded gan-zhi interactions, and relationship-pattern synthesis.
- Make evidence directional where needed: "A experiences B as pressure/support/attraction" is not always symmetric.
- Keep score computation bounded and traceable to evidence contribution.
- Feed richer evidence into API detail, prompt data, and frontend professional evidence sections.
- Keep legacy readings and existing clients functional.

**Non-Goals:**

- Replacing the bazi calculation engine or adding a new external astrology dependency.
- Predicting exact marriage, breakup, affair, pregnancy, or reconciliation dates.
- Turning compatibility into a paid advisor workflow, chat product, journal, or marketplace.
- Requiring relationship context; that belongs to `compatibility-consulting-ux-v2`.
- Rebuilding all historical readings unless the user opens or regenerates them.

## Decisions

### D1: Add signal builders behind the existing compatibility engine

Keep `AnalyzeCompatibility(a, b)` as the public entry point and split internal evidence generation into focused builders:

```text
AnalyzeCompatibility
  -> buildDayMasterSignals
  -> buildFiveElementSignals
  -> buildSpousePalaceSignals
  -> buildTenGodInteractionSignals
  -> buildFavorableElementSupportSignals
  -> buildGanZhiInteractionSignals
  -> buildRelationshipPatternSignals
  -> buildOptionalTimingSignals
  -> scoreAndSummarize
```

Rationale:

- The current package already returns a single `CompatibilityAnalysis`.
- Focused builders keep individual rules testable.
- The service, repository, and frontend can remain mostly unchanged while evidence quality improves.

Alternatives considered:

- Create a separate compatibility-v2 engine. Rejected because it would duplicate score/report plumbing and complicate existing readings.
- Move signal generation into the service layer. Rejected because the bazi package should own deterministic bazi logic.

### D2: Keep four scores but make every score movement evidence-backed

Each signal should produce one or more evidence records with:

- dimension
- polarity
- source
- type
- title/detail
- weight
- optional perspective metadata if a signal is directional

Scores remain bounded with the existing clamp behavior. No hidden score adjustments should happen without evidence.

Rationale:

- Frontend and report contracts already rely on four score fields.
- Traceability is more important than adding many new numeric sub-scores.
- Existing result pages can evolve incrementally.

### D3: Represent directional signals without breaking old evidence consumers

Some new signals are directional. For example, one person's day master may experience the other person's element as officer pressure, wealth attraction, resource support, expression outlet, or peer competition. The implementation should add optional metadata fields, such as:

```json
{
  "perspective": "self_to_partner",
  "actor": "self",
  "target": "partner"
}
```

Existing renderers must still work from `title`, `detail`, `dimension`, `polarity`, and `source`.

Rationale:

- Directional meaning is central to ten-god compatibility.
- Optional metadata preserves backward compatibility.

### D4: Derive favorable-element support conservatively

The first implementation should use available chart structure to infer support from five-element distribution and day-master context. It should not claim full professional yongshen precision unless the underlying bazi result already provides a verified yongshen model.

Use language like "support tendency" or "pressure tendency" and expose the evidence basis.

Rationale:

- The project has advanced bazi specs, but compatibility should not overstate precision before those engines are fully integrated.
- Conservative claims are safer and easier to test.

### D5: Expand gan-zhi interactions globally, then prioritize relationship-relevant positions

The signal engine should evaluate heavenly-stem and earthly-branch interactions across all pillars, but relationship weighting should prioritize:

- day pillar / spouse palace
- month pillar for long-term rhythm and family/social pressure
- hour pillar for private-life and future-planning tendencies
- year pillar as lower-weight background affinity

Rationale:

- Looking only at day branch misses meaningful patterns.
- Equal weighting across all pillars would overstate distant/background signals.

### D6: Timing signals are optional and non-deterministic

If the existing chart data includes usable big-luck/current-year context, add timing signals as risk/focus windows. If not available, skip the layer without reducing core compatibility quality.

Timing copy must avoid exact dates and fate language. It should say what to observe and how to respond, not what must happen.

Rationale:

- Timing can make reports feel more complete.
- It carries higher risk of overclaiming, so it must be bounded and optional.

## Risks / Trade-offs

- [Risk] More signals make scores feel unstable. -> Mitigation: bound weights, cap per-source contribution, and add tests for representative chart pairs.
- [Risk] Professional terminology confuses ordinary users. -> Mitigation: keep plain-language summaries primary and place technical detail in professional evidence sections.
- [Risk] Favorable-element claims become too strong. -> Mitigation: use conservative language unless a verified yongshen model is integrated.
- [Risk] AI reports invent unsupported claims. -> Mitigation: prompt from structured evidence and require evidence keys for major claims.
- [Risk] Existing persisted readings lack new evidence. -> Mitigation: new analysis applies to newly created readings and regenerated/backfilled details where current code already supports derived fallback behavior.

## Migration Plan

1. Add backend unit tests for each new signal family before changing score behavior.
2. Refactor compatibility evidence generation into focused builders while preserving current outputs.
3. Add new signal families and bounded score contribution rules.
4. Extend evidence model/API types only with optional metadata fields if needed.
5. Update AI prompt data and frontend professional evidence grouping.
6. Verify old readings still render and new readings show richer evidence.

Rollback: disable the new signal builders or gate them behind a service-level flag while keeping the existing compatibility engine path intact.

## Open Questions

- Should timing signals ship in the first implementation or remain a later subtask after core depth signals are stable?
- Should evidence metadata include a numeric confidence field, or is polarity/weight/source enough for v1?
- Should score contribution caps be global per source or per dimension?
