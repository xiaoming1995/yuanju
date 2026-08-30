## Context

The result page currently renders the global Fuyi conclusion as element strings such as `水金` and `木火`. The natal assessment already retains a more precise, ordered list of day-stem adjustment requirements (`tiaohou.day_stem.required_stems`), but it does not reconcile those stems with the Fuyi element conclusion. Consequently, the UI cannot truthfully distinguish a shared-priority stem such as `壬水` from another stem of the same favorable element, or explain that a stem can support day-stem adjustment while conflicting with Fuyi.

The Dayun road model was recently validated against Fuyi elements, Ten Gods, and gated pattern interaction. This change must not alter those scores.

## Goals / Non-Goals

**Goals:**
- Return additive, deterministic stem-level guidance with each recommendation's five element, Ten God, source layers, and explanation.
- Identify stems selected by both day-stem adjustment and Fuyi as primary favorable stems.
- Keep day-stem adjustment requirements that conflict with Fuyi visible as contextual structure, rather than presenting them as general favorable stems.
- Keep the existing element-level Fuyi badges and provide a concise, readable detail surface for the stem guidance.

**Non-Goals:**
- Replacing the existing Fuyi element conclusion, legacy `yongshen`/`jishen`, or day-stem adjustment assessment.
- Selecting a single universally beneficial stem from a favorable element without a cross-layer basis.
- Changing Dayun road, annual-flow, vehicle-grade, or natal score calculations.
- Treating a natal structural requirement as a guarantee of favorable future outcomes.

## Decisions

### Add a separate, additive natal stem-guidance object

`NatalAssessment` will gain an optional `stem_guidance` object. Its recommendation items will include `stem`, `element`, `ten_god`, `source_layers`, and `detail`; groups will be `primary_favorable`, `secondary_favorable`, `conditioning_only`, and `adverse`.

This preserves existing element strings for all consumers and gives frontend code a typed, explainable payload. A single flat list with a severity field was rejected because the four groups carry distinct meanings and a conflict is not an adverse prediction.

### Derive primary stems from an exact cross-layer intersection

The engine will take the ordered `tiaohou.day_stem.required_stems` list and compare each stem's element with `fuyi.yongshen`. A matching stem becomes `primary_favorable` with both source layers. The day-stem adjustment order is retained, so the dictionary's existing priority is not lost.

This avoids arbitrarily treating either yin or yang form of a favorable element as superior. For example, `丙` day master with a Fuyi result of `水金` and a day-stem requirement of `甲、壬` yields `壬水` as the primary favorable stem.

### Keep residual Fuyi stems usable but unranked beyond the shared intersection

For every Fuyi favorable element, both heavenly stems are emitted in the Fuyi element order and Yang-then-Yin order. Any stem already classified as primary is removed from this secondary group. These entries are labeled as Fuyi-only support, not as a claim that one is universally better.

### Make cross-layer conflicts explicit and non-duplicated

Required day-stem adjustment stems that are not primary go to `conditioning_only`. When a stem's element belongs to `fuyi.jishen`, its detail MUST state the conflict: it helps the natal day-stem adjustment structure but is not a general future Fuyi favorable stem. Those conflict stems are excluded from the visible adverse group to prevent the same stem from appearing as both a structural reference and an adverse recommendation.

Required stems whose element is neither the current Fuyi favorable nor adverse set remain in `conditioning_only` as a neutral adjustment reference.

### Render a short summary and progressive detail

The main result header retains the compact element badges. When a primary stem exists, the Fuyi badge adds the concise text `天干优先：<stem><element>`. A result-page evidence modal or equivalent progressive detail surface renders all four groups with source and Ten God labels. Older API payloads without `stem_guidance` continue to render the existing element-only layout.

Embedding every candidate stem in the header was rejected because it recreates the clutter the result-page hierarchy work removed.

### Keep scoring consumers independent

Dayun and flow-year scoring continue to read their present Fuyi element and Ten God inputs. The new object is explanatory in this change. This maintains the validated behavior in which a Dayun stem's score is supported by its element, Ten God, pattern relay, and phase, not by a display-only rank.

## Risks / Trade-offs

- [A user may read `conditioning_only` as a second favorable list] → Use explicit label and conflict text: “原局调候结构，不作为后天通用喜神”.
- [Fuyi may be neutral or incomplete] → Emit no favorable or adverse stems and retain only available adjustment references; the UI states that a Fuyi stem direction is not determined.
- [The API response grows] → The payload is small, additive, and optional; clients without the new UI ignore it.
- [Future scoring needs more exact stem behavior] → Preserve source layers and Ten God metadata now, but defer score changes to a separately specified change.

## Migration Plan

1. Add the optional backend assessment field and unit coverage.
2. Update the result-page types and render the new guidance only when present.
3. Verify current element badges and all road scoring fixtures are unchanged.
4. Roll back by omitting `stem_guidance`; older frontend fallback remains element-only.

## Open Questions

None for the display-only scope. A later scoring change would need an explicit rule for how primary and secondary stems affect Dayun or annual-flow deltas.
