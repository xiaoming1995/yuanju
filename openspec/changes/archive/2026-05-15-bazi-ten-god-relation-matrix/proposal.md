## Why

The result page currently shows raw ten-god labels such as "七杀", "偏印", and "食神", but it does not make the reference point explicit: every label is derived from the chart owner's day stem. Users can see professional terms, but cannot understand how each heavenly stem or hidden earthly-branch stem relates back to the day master.

## What Changes

- Add a "命主十神关系" capability to explain ten-god relationships from the day-master perspective.
- Show the day master as the fixed reference point, for example "命主日元：丙火".
- Present the four heavenly stems as a relationship matrix:
  - position, stem, element, ten-god label, and plain-language relationship.
  - the day stem itself must be labeled as "日主 / 日元" rather than treated as an ordinary external ten-god.
- Present each earthly branch's hidden stems with their ten-god labels and short explanations.
- Provide concise descriptions for every ten-god label and grouped meaning, such as wealth, official/pressure, seal/resource, output/expression, and peer/competition.
- Render the module in a mobile-friendly card or collapsible layout, without making the existing professional grid denser.
- Preserve compatibility with existing chart snapshots by deriving the matrix from existing fields when new structured fields are absent.

## Capabilities

### New Capabilities

- `bazi-ten-god-relation-matrix`: Defines how the system exposes day-master-centered ten-god relationships for heavenly stems and hidden earthly-branch stems.

### Modified Capabilities

- `bazi-advanced-data`: Clarifies how existing ten-god and hidden-stem data should be surfaced in a user-understandable result-page module.

## Impact

- Backend bazi engine result model may gain a structured ten-god relation matrix for stable API and snapshot reuse.
- Existing `BaziResult` fields such as `day_gan`, `*_gan_shishen`, `*_hide_gan`, and `*_zhi_shishen` remain compatible.
- Frontend `ResultPage` gains a new interpretation module near the professional chart area.
- Static frontend tests and backend bazi tests should cover day-stem reference behavior, hidden-stem pairing, and mobile rendering expectations.
