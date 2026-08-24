## Context

Past-events year data already contains two different layers:

1. deterministic bazi signals and evidence, which are specific and auditable;
2. user-facing year narratives, which are easier to read but can become too generic.

The current evidence summary is often more accurate because it is generated directly from `EventSignal.Evidence`. The visible narrative can be produced by deterministic templates or dayun summary output, and both paths can lose concrete triggers such as clashes, punishments, void, shensha, ten-god effects, or `夹拱` when translating them into general advice.

This change keeps the signal engine intact and improves only the narrative layer. The narrative should become a plain-language explanation of the strongest evidence, not a separate broad fortune comment.

## Goals / Non-Goals

**Goals:**

- Make each meaningful year narrative understandable to users who do not know bazi terminology.
- Keep narratives grounded in concrete algorithmic evidence.
- Allow longer narratives when evidence supports them, while keeping them structured and scannable.
- Preserve professional evidence details for users who want to audit the conclusion.
- Prevent weak generated content from overriding a more accurate evidence-aligned fallback.

**Non-Goals:**

- Do not change the bazi signal calculation, weighting, ten-god power scoring, or dayun phase logic.
- Do not add automatic AI generation for future dayun segments.
- Do not change database schema, public routes, or SSE payload shape unless a compatible optional field is needed.
- Do not hide professional terms entirely; explain them where they matter.

## Decisions

### Decision 1: Render year narratives from evidence first

The deterministic narrative path should build the year narrative from the strongest usable signals and their evidence summaries. A meaningful year with evidence should explain at least one concrete trigger, such as:

- a branch clash or punishment affecting a relationship, home, movement, or cooperation area;
- a ten-god signal affecting learning, work, money, pressure, relationship, or responsibility;
- a shensha or `夹拱` signal that indicates help, movement, attention, pressure, or unexpected support;
- a void, phase, or repeated hit that changes the strength or reliability of the event tendency.

Rationale: this directly addresses the observed mismatch where evidence is more accurate than the visible narrative.

Alternative considered: only make the existing narrative longer. That would add text without guaranteeing better accuracy.

### Decision 2: Use a stable plain-language paragraph structure

Each expanded year narrative should follow the same conceptual order:

1. yearly tendency: one sentence describing the nature of the year;
2. why: one or two sentences translating the most important evidence into everyday meaning;
3. life areas: one sentence naming likely areas such as work, study, relationship, family, movement, money, health, or cooperation;
4. stance: one practical sentence about how to handle the year.

The implementation does not need to expose numbered labels in the UI. The structure is for generation quality and testability.

Rationale: longer text remains readable only if the content has a predictable shape.

Alternative considered: separate visible tabs for conclusion, explanation, and advice. That is heavier UI work and unnecessary for this iteration.

### Decision 3: Keep professional evidence as secondary detail

The UI should continue showing `命理依据` as expandable or lower-emphasis detail. The main narrative should explain the evidence, while the detail area preserves the raw professional wording.

Rationale: ordinary users need translation; professional users need auditability.

Alternative considered: replace evidence with only natural language. That would reduce trust and make debugging harder.

### Decision 4: Prefer evidence-aligned fallback over weak generated text

When AI/dayun-generated year text is missing, too short, unsupported, or more generic than the deterministic evidence-aligned narrative, the system should use the deterministic narrative as the visible fallback. The generated dayun summary can still be shown at the segment level, but individual year text should not become less precise than available evidence.

Rationale: the user-visible content should not regress in accuracy because a generated paragraph sounded smoother.

Alternative considered: always trust generated year narratives after validation. Existing validation can reject unsupported technical claims, but it does not guarantee the output is specific or useful.

### Decision 5: Add term translation at the narrative boundary

The narrative renderer should maintain a small translation layer for common bazi triggers. Examples:

- `冲`: a collision or push that can bring change, movement, conflict, or re-negotiation;
- `刑`: friction, pressure, repeated trouble, or procedural complications;
- `合`: connection, cooperation, attraction, or being tied into a situation;
- `空亡`: uncertainty, delay, weakened certainty, or plans that need confirmation;
- `驿马`: movement, travel, relocation, errands, or changing environment;
- `天乙贵人`: help, mediation, or support from others;
- `夹拱`: an indirect or hidden support/trigger that can feel unexpected but traceable.

Rationale: the translation layer is where professional evidence becomes readable without removing the technical basis.

Alternative considered: ask the AI to explain every term dynamically. That increases cost and makes results less deterministic.

## Risks / Trade-offs

- Longer year cards can make the page feel dense -> keep each narrative to a compact paragraph and leave raw evidence collapsed.
- Evidence translation can become too deterministic or repetitive -> vary by dominant theme and signal polarity rather than using one generic template.
- Some signals may not map cleanly to life areas -> use conservative language such as "更容易体现为" and fall back to evidence summary instead of over-claiming.
- Generated narratives may conflict with deterministic fallback -> prefer evidence-aligned text at year-card level and keep generated content at dayun-summary level when conflict is detected.
- Professional users may want the exact terms visible -> keep `命理依据` visible and do not remove technical chips.

## Migration Plan

- Implement as a compatible narrative rendering change.
- Keep existing API field names and response shape.
- Update deterministic year narrative tests first, then frontend rendering tests if display behavior changes.
- Existing cached dayun summaries can remain valid; fallback logic can improve newly rendered year cards without requiring database migration.
- Rollback is limited to reverting narrative rendering and fallback selection changes.
