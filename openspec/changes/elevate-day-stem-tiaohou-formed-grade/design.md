## Context

`NatalDayStemTiaohouAssessment` currently records the presence of the day-stem adjustment stems as `resolved` / `partial` / `missing` and contributes at most 12 points to the natal score. `NatalPatternAssessment.Quality` is calculated independently from the identified main pattern's visible, hidden, and root-strength evidence. As a result, one chart can correctly have day-stem adjustment stems in both heaven and earth while the main pattern is displayed as "部分成立" because it does not meet the unrelated primary-star root threshold.

The product rule is more specific: when the required day-stem adjustment stems are present across heavenly stems and earthly branches/hidden stems, the day-stem adjustment is formed (天透地藏成格) and establishes a high-pattern foundation. Pattern transformation and restraint (制化配合),扶抑 strength, and seasonal heat/cold adjustment are separate questions. They must remain visible and must not be used to deny the formation itself.

## Goals / Non-Goals

**Goals:**

- Make day-stem adjustment formation an explicit, evidence-based conclusion rather than an incidental score.
- Treat complete heaven-and-earth coverage of the required stems as `formed` and as a `high` natal-pattern foundation.
- Keep generic primary-pattern quality and transformation/restraint evidence separate from the day-stem formation conclusion.
- Let the vehicle-grade calculation benefit from the high foundation without hiding urgent climate,扶抑, or critical-break limitations.
- Preserve backward-compatible serialized fields while lazily refreshing historical snapshots through an assessment-version bump.

**Non-Goals:**

- Redefining the day-stem adjustment dictionary or the thermal adjustment rules.
- Treating a high-pattern foundation as a guarantee of an S/A final vehicle tier regardless of unresolved urgent conditions or weak carrying capacity.
- Replacing the existing main-pattern detector or removing the existing 制化配合 rules.
- Making claims about a person's worth or inevitable life outcome from the grade.

## Decisions

### 1. Add a separate formation conclusion to the day-stem adjustment assessment

`Status` remains the compatibility-oriented availability state (`resolved`, `partial`, `missing`, `unavailable`). New fields express the doctrinal result separately:

- `formation`: `formed`, `partial`, `unformed`, or `unavailable`
- `foundation_tier`: `high`, `normal`, or `none`
- `foundation_score`: the explicit score contribution granted by that foundation
- existing required, visible, and hidden evidence remains the explanation source

The formation check MUST verify that every required stem is covered, rather than comparing only the number of matched stems. A chart is `formed` only when all required stems are collectively present and the evidence spans both a heavenly stem and an earthly-branch hidden stem. A single required stem therefore must be evidenced in both places; multiple required stems may be split between the two locations.

This keeps the existing UI/API semantics stable while exposing the user rule in an unambiguous field. Replacing `Status` outright was rejected because saved data and existing components currently rely on it.

### 2. Model "高格基础" separately from the generic main-pattern quality

Extend `NatalPatternAssessment` with a foundation summary sourced from the day-stem adjustment assessment. When `formation == formed`, it MUST identify `日干调候成格` as the basis and label it `高格基础`.

The current primary-pattern quality (`formed`, `usable`, `partial`, `broken`) continues to describe only the detected main pattern and its root/制化 evidence. It MUST be displayed as `主格结构` or equivalent wording, never as a statement that negates the day-stem formation. The interface can therefore accurately show both conclusions, for example: `日干调候成格·高格基础` and `偏印格主格结构：部分成立`.

This is preferred to forcing the generic `Pattern.Quality` to `formed`, because that would incorrectly claim the identified main pattern itself met its own evidentiary criteria.

### 3. Promote the formed foundation through a versioned, explicit score component

The assessment score will consume `foundation_score` rather than treating the existing day-stem availability points as the only representation of a formed adjustment. A formed condition receives a materially higher, separately auditable contribution than partial coverage; partial and missing conditions retain conservative lower contributions.

The evidence list MUST include a dedicated `日干调候成格` row with its rule identifier, evidence, and contribution. This lets the vehicle profile and AI prompt explain why the base is high without inferring it from a total.

The final vehicle grade still applies existing urgent thermal,扶抑, and critical-pattern ceilings after all contributions are calculated. This preserves the agreed hierarchy: formation establishes a high base; whether that base can be carried and expressed depends on the other retained dimensions.

### 4. Present the two layers separately in product and AI text

The result page and vehicle profile will lead with the foundation label when formed, then show main-pattern structure and 制化配合 as separate evidence. Summary text MUST state that the high base comes from day-stem adjustment formation and MUST name any retained limitation rather than using a generic "格局部分成立" sentence.

AI report context will include the formation, foundation tier, required stems, visible stems, hidden stems, and separate main-pattern/制化 result. The prompt must instruct the model not to collapse them into a single "成格" or "不成格" claim.

### 5. Version and test the rule set

Increase `NatalAssessmentVersion` so saved charts receive the new fields through the existing lazy assessment refresh. Tests will use the 1995-10-12 午时 chart to assert the expected split evidence and formed/high conclusion, plus negative cases for only-visible, only-hidden, and incomplete required stems.

## Risks / Trade-offs

- [High foundation could be mistaken for an unconditional final grade] → Display the foundation and final vehicle tier separately, and retain explicit limitation evidence in the summary.
- [A generic main-pattern "部分成立" label may still confuse users] → Rename its presentation to `主格结构` and place it below the day-stem formation conclusion.
- [Availability and formation fields can diverge if implemented separately] → Derive both in one function and test the complete evidence matrix.
- [Historical saved reports may contain old language] → Refresh structured assessment lazily and version future report generation; historical immutable reports are not retroactively rewritten.

## Migration Plan

1. Add the formation/foundation fields and bump the natal assessment version.
2. Update calculation, vehicle profile, API rendering, and AI prompt generation together.
3. Rebuild existing charts on read through `EnsureNatalAssessment`; no database migration is required because the assessment is embedded JSON.
4. Deploy backend before frontend so older clients safely ignore additional JSON fields.
5. Roll back by restoring the prior assessment version and code; persisted snapshots remain readable because new fields are additive.

## Open Questions

None. The change follows the confirmed rule that day-stem adjustment heaven-and-earth coverage is formed and establishes a high-pattern foundation, while 制化配合 remains independently evaluated.
