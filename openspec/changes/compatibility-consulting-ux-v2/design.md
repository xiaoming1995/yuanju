## Context

The existing compatibility product already has a pair-reading resource, pair evidence, duration assessment, AI report generation, and a result page. The current weakness is product framing: the flow asks only for birth data, then renders scores and generic relationship advice. This makes the output feel like a calculation result rather than a consultation for the user's current relationship decision.

The latest consulting-report work on `main` adds structured diagnosis, decision advice, stage risks, strategy, and evidence links. This v2 change should build on that structure by adding user context and reordering the reading experience around the user's immediate question.

Constraints:

- Keep the existing bazi compatibility engine and four score fields.
- Keep existing readings viewable when context is missing.
- Use pure CSS variables and existing frontend patterns.
- Avoid deterministic fate language, exact breakup dates, and unsupported claims.

## Goals / Non-Goals

**Goals:**

- Capture relationship stage and primary user question before creating a compatibility reading.
- Persist that context and include it in detail responses, history summaries, and AI prompt data.
- Render the result as a decision-oriented consultation: conclusion, core contradiction, next actions, risks, then score/evidence details.
- Adapt AI report structure and wording to the selected user question.
- Preserve evidence traceability for every major consulting claim.

**Non-Goals:**

- Rebuilding the compatibility algorithm or introducing a new scoring system.
- Adding object comparison, social sharing, relationship journals, or advisor marketplace features.
- Predicting exact marriage, breakup, reconciliation, or event dates.
- Making relationship context mandatory for legacy records or API clients.

## Decisions

### D1: Store relationship context on the compatibility reading

Add optional context fields to the reading-level model rather than creating a separate table.

Proposed shape:

```json
{
  "relationship_stage": "ambiguous",
  "primary_question": "continue_investment"
}
```

Rationale:

- The context is one-to-one with a reading.
- It affects report framing, history labels, and result rendering.
- Keeping it on the reading avoids joining another table for every detail view.

Alternatives considered:

- Store context only in AI report prompt data. Rejected because history/detail views also need the context.
- Use free-text context first. Rejected for v2 because controlled choices are easier to test, translate, and map to report templates.

### D2: Use controlled enums with neutral fallback

Use fixed relationship stages:

- `ambiguous`
- `dating`
- `long_distance`
- `reconciliation`
- `marriage_or_engagement`
- `crush`
- `general`

Use fixed primary questions:

- `continue_investment`
- `marriage_suitability`
- `recurring_conflict`
- `reconciliation_potential`
- `long_term_stability`
- `relationship_strategy`
- `general`

Unknown or missing values fall back to `general`.

Rationale:

- The product can map each value to stable UI labels and prompt sections.
- Missing context remains safe for old readings.

### D3: Treat the result page as a consultation, not a score report

The top reading path should be:

```text
Context label
Decision headline
Core contradiction
Next actions / Avoid
Stage risks
Question-form dimensions
Professional evidence
AI full report
```

The existing score fields remain but are relabeled in the UI:

- `attraction` -> "会不会互相吸引?"
- `stability` -> "能不能长期稳定?"
- `communication` -> "吵架后能不能修复?"
- `practicality` -> "现实条件能不能落地?"

Rationale:

- This keeps the data contract stable while improving user comprehension.
- It answers the user's decision before asking them to interpret scores.

### D4: Route AI report structure by primary question

Prompt data should include relationship context and ask the model to prioritize sections based on `primary_question`.

Examples:

- `reconciliation_potential`:复合判断、原问题是否可修复、复合后重复风险、验证信号、边界条件。
- `marriage_suitability`:婚姻适配、现实承接、冲突处理、家庭/责任/边界、谈婚前确认项。
- `continue_investment`:继续投入判断、短期验证点、投入节奏、风险边界、下一步行动。

Rationale:

- A single generic report cannot feel specific across very different relationship situations.
- The controlled enum lets prompt branching stay deterministic.

### D5: Keep evidence expandable and claim-linked

Professional evidence should remain hidden behind explicit expansion points. Major claims should keep `evidence_keys` and render "查看依据" or equivalent affordance.

Rationale:

- Ordinary users need the conclusion first.
- Professional users need enough traceability to trust the result.

## Risks / Trade-offs

- [Risk] Context choices feel too narrow. -> Mitigation: include `general` fallback and design labels so they can expand later.
- [Risk] AI ignores question-specific structure. -> Mitigation: add structured output fields and fallback rendering from deterministic consulting assessment.
- [Risk] Result page becomes too long. -> Mitigation: keep professional evidence collapsed and promote only the top decision blocks.
- [Risk] Existing readings lack context. -> Mitigation: default missing context to general relationship judgment.
- [Risk] Users treat advice as absolute fate. -> Mitigation: UI and prompt must use conditional language and avoid exact-date claims.

## Migration Plan

1. Add nullable/defaulted compatibility context fields.
2. Update create-reading request parsing to accept optional context.
3. Backfill behavior: old rows behave as `general` without writing migration updates.
4. Update frontend input and result pages behind the same route.
5. Keep report generation compatible when context is missing.

Rollback is straightforward: hide the context controls and ignore context fields in prompt/rendering while keeping database columns.

## Open Questions

- Should the user be allowed to change relationship stage/question after a reading is created, or should they create a new reading?
- Should `primary_question` support free-text later for premium reports?
- Should history cards show the selected question as a label, or only the relationship stage?
