## Context

`ming_ge` is already calculated by the backend and shown on the result page, but its fixed description remains abstract for ordinary users. The project also has a legacy `celebrity_records` table, whose entries are generic AI-report material and have no required birth-data precision, source, review state, historical milestone, or verified Dayun data. It cannot support public claims that a historical figure's success coincided with a particular Dayun.

This change adds editorial “古人映照” content to explain a detected Ming Ge. It is not a celebrity-similarity feature, a success-prediction feature, or an AI-generated report chapter.

## Goals / Non-Goals

**Goals:**
- Present up to two published historical figures for a detected Ming Ge, with concise historical memory and a fixed boundary statement.
- Let professional readers inspect an editorially verified life turning point and, only when birth data is sufficiently reliable, a Dayun explanation tied to that turning point.
- Give administrators an auditable workflow to draft, review, publish, unpublish, and cite every displayed entry.
- Keep historical content independent of saved chart snapshots so editorial corrections take effect immediately.

**Non-Goals:**
- Matching a user to a historical figure, scoring similarity, or saying that the user will reproduce a figure's life.
- Inferring a historical figure's Bazi, achievement, or Dayun with the LLM at request time.
- Showing an exact Dayun for a figure without a verified birth date, hour, calendar convention, and supporting source.
- Replacing the existing generic `celebrity_records` library or re-enabling the removed report “命理分身” chapter.

## Decisions

### 1. Use a separate curated historical-reference model

Create a dedicated `mingge_historical_figures` model rather than extending `celebrity_records`. A published entry contains: matched Ming Ge, name, era, identity, historical-memory text, turning-point text and year/range, source title and URL, birth-data precision, Bazi verification note, optional Dayun fields, display order, and review status (`draft`, `reviewed`, `published`, `archived`).

Only `published` entries are returned by the public read API. `show_dayun` is permitted only when the entry records exact-hour birth precision, a verification note, a source, and a complete Dayun explanation. The repository validates these conditions when an administrator publishes or enables the Dayun display.

**Rationale:** Editorial data needs a stronger provenance contract than the existing AI-generated celebrity material. A separate model prevents unreviewed legacy entries from leaking into a user-facing interpretive surface.

**Alternatives considered:**
- Extend `celebrity_records`: rejected because it lacks provenance, validation and life-event fields, and is designed for broad LLM context rather than public editorial content.
- Hard-code a Go map: rejected because source corrections and editorial review would require a release and provide no content workflow.

### 2. Resolve references through a dedicated read endpoint

Expose a read-only public endpoint keyed by the detected `ming_ge`; the result page fetches it only when a Ming Ge exists. The endpoint returns at most two published entries in display order and never returns drafts, archive records, or Dayun fields marked unavailable.

**Rationale:** A historical reference is editorial content, not a property of the user's chart. Fetching it independently means content fixes apply to existing history records without recalculating or mutating `result_json` snapshots.

**Alternative considered:** Include references in every `Calculate` response. Rejected because it duplicates changing content into chart snapshots and adds a database lookup to all calculations.

### 3. Separate ordinary and professional reading depth

In the Ming Ge area, ordinary mode shows an unframed “古人映照” section containing each name, identity, and the historical act or contribution remembered by later generations. Professional mode adds the turning point, sources, and eligible Dayun interpretation. Both modes show a concise statement that this is a cultural reference for understanding the Ming Ge, not a life-template or prediction.

When there is no published reference, the section is not rendered. When a reference has no verified Dayun, professional mode still renders the historical content but omits the Dayun subsection rather than using placeholder or AI-generated text.

**Rationale:** The ordinary view remains readable while professionals can inspect the evidence behind a qualified Dayun statement.

### 4. Treat Dayun as contextual correspondence, not a single cause

Published Dayun wording uses language such as “与其人生转折阶段形成呼应” and explains the relation to the figure's verified original-chart conditions. It MUST NOT state that a single Dayun caused success or that readers in the same Dayun will achieve the same outcome.

**Rationale:** The existing product already treats Tiaohou, Fuyi Yongshen, original structure and Dayun as combined conditions. This avoids contradicting that model and avoids deterministic claims.

### 5. Seed narrowly and review before publication

The first release seeds one or two reviewed figures for the common supported Ming Ge categories: 食神、伤官、正财、偏财、正官、七杀、正印、偏印、建禄、月刃. Each seed must contain a historical source and a content reviewer decision. A category without qualified source material remains absent rather than padded with a low-confidence entry.

**Rationale:** Coverage is useful, but historical date/time claims carry a higher quality bar than generic descriptive content.

### 6. Backfill missing Ming Ge fields in historical snapshots

When `LoadOrCalculateResult` deserializes an existing `result_json` whose `MingGe` or `MingGeDesc` is empty, it calculates those two fields from the already saved four-pillar result using `DetectMingGe`, then writes the upgraded snapshot back. It does not recalculate the full Bazi result, change the chart's birth input, or replace existing algorithm-derived fields.

**Rationale:** The current result page intentionally hides the Ming Ge badge when the field is absent. Historical snapshots created before the Ming Ge field was introduced remain structurally valid, so the existing lazy calculation path does not run. A narrow, idempotent field backfill restores the missing display and provides the prerequisite for historical-reference lookup without algorithm-version drift.

**Alternative considered:** Recalculate every historical Bazi result. Rejected because it would alter unrelated historical fields after algorithm changes and is unnecessary when the snapshot contains the required pillar data.

## Risks / Trade-offs

- [Historical birth time is disputed or unavailable] → Store the precision and verification note; hide exact Dayun unless the strict display gate is met.
- [A factual source later proves inaccurate] → Keep entries database-backed and independently fetched, so an administrator can unpublish or correct them without touching user history snapshots.
- [Users mistake the figure for a personal prediction] → Use “古人映照” terminology and the fixed non-template boundary in both reading modes.
- [Initial content collection delays release] → Render no empty state and release individual Ming Ge categories only after their entries pass review.
- [Public source links are untrusted] → Validate URL scheme and render source links with standard safe external-link handling.
- [Old snapshots silently lack Ming Ge] → Backfill only `MingGe` and `MingGeDesc` during normal history reads and cover this compatibility path with regression tests.

## Migration Plan

1. Add the dedicated table and indexes without changing `celebrity_records` or chart snapshots.
2. Deploy the read endpoint and admin workflow with no published data; the result page remains unchanged when the endpoint returns no entries.
3. Import reviewed seed content as drafts, verify the display gates, then publish category by category.
4. Roll back by disabling the public route or unpublishing entries; the result page treats a failed/empty reference request as no section and the rest of the Bazi result remains available.

## Open Questions

- The editorial team must select the initial historical figures and authoritative sources before seeding; implementation must not invent this data.
