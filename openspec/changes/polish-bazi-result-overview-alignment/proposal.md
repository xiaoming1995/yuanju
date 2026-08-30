## Why

The Bazi result page has accumulated vehicle, road, grade explanation, and professional evidence modules inside one overview grid. Their different content heights stretch adjacent cards, disrupt the intended reading order, and leave large empty panels that make the page feel unaligned.

## What Changes

- Reorder the result overview so the natal conclusion is the first visible decision layer.
- Limit the two-column overview to concise `命盘座驾` and `当前路况` summaries with a consistent header and spacing system.
- Move full grade guidance, road metaphor explanation, and professional algorithm evidence out of the summary cards into lower-priority expandable sections.
- Ensure desktop cards top-align instead of stretching a short road card to the vehicle card's height, and preserve a deliberate single-column mobile reading order.
- Keep all existing calculation data, grade explanations, road data, and professional evidence accessible without changing their meanings.

## Capabilities

### New Capabilities
- `bazi-result-overview-layout`: Present the Bazi result overview in a stable hierarchy with aligned summaries and progressively disclosed explanation.

### Modified Capabilities

None.

## Impact

- Affects `frontend/src/pages/ResultPage.tsx` and `frontend/src/pages/ResultPage.css`.
- Requires no API, backend algorithm, data-model, or dependency changes.
- Requires desktop and mobile visual regression checks for result pages with and without vehicle-road data.
