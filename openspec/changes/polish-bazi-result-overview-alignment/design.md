## Context

`ResultPage` currently renders the total verdict, vehicle profile, and current road as children of one two-column grid. CSS ordering moves the verdict after the other two cards, while the vehicle card includes its full grade guide and can become much taller than the road card. Default grid stretching therefore turns a concise road summary into a large empty panel.

The result page must support both plain and professional reading modes. The page already has progressively disclosed `details` controls for road and evidence, and it must continue to expose the full S-to-D grade explanation requested by users.

## Goals / Non-Goals

**Goals:**
- Make the visible order agree with the reading order: conclusion, concise natal/current overview, then supporting explanation.
- Prevent differing card content heights from creating stretched blank areas on desktop.
- Keep grade guidance, road explanation, and professional evidence accessible without allowing them to dominate the overview.
- Use a stable single-column order on mobile with no duplicated content.

**Non-Goals:**
- Change vehicle grades, road classification, yongshen, or any backend/API behavior.
- Redesign the whole result page or replace its existing dark visual system.
- Hide or remove the complete grade explanation and professional evidence.

## Decisions

### Separate decision, overview, and evidence bands

The total verdict will be a full-width first band. A dedicated overview grid will contain only the vehicle and current-road summary cards. Grade guidance, road metaphor text, and professional evidence will follow as lower-priority expandable content rather than remain inside the primary vehicle summary.

This creates one visual task per band: understand the conclusion, compare natal configuration with current conditions, then inspect explanation. Keeping all content in a single grid was rejected because a variable-height explanatory card necessarily distorts its short sibling.

### Use top-aligned summary cards with content-driven height

The desktop summary grid will retain a wider vehicle column and narrower road column, but it will use top alignment rather than stretch. Both cards will share header treatment, padding, badge sizing, and keyword rhythm, while their backgrounds end at their own content height.

Forcing equal card heights was rejected because it produces the current large empty road panel whenever a vehicle result has detailed context.

### Show the active grade first and disclose the full guide on demand

The vehicle card will show the current grade's explanation inline with the score and profile. The full S-to-D grade guide will live in a labelled expandable explanation section. The road metaphor and expert evidence use the same progressive-disclosure treatment, preserving access while reducing first-screen density.

Removing the complete guide was rejected because users need to understand the level scale. Keeping the entire guide expanded by default was rejected because it competes with current-road information.

### Establish one semantic DOM sequence

Markup order will match the intended visual and assistive-technology order: verdict, vehicle summary, current road, grade/road guide, then professional evidence. CSS `order` will not be used to reverse this sequence. On narrow screens, the overview becomes a single column in the same order.

## Risks / Trade-offs

- [Users may overlook the full grade guide when collapsed] → The active grade explanation remains visible and the disclosure heading clearly identifies the full scale.
- [Professional users need frequent evidence access] → Evidence remains one interaction away in professional mode, with no data removed.
- [Different content lengths still create visual imbalance] → Use top alignment, bounded typography, and shared card header spacing; do not prescribe artificial equal heights.
- [Moving markup changes existing selectors] → Keep existing data calculations and interaction handlers, then validate desktop/mobile and both reading modes.

## Migration Plan

1. Restructure only the result-page overview markup and class names.
2. Apply the new band/grid styles and responsive rules.
3. Verify a chart with vehicle-road data at desktop and mobile viewports, in plain and professional modes.
4. Verify the fallback page without vehicle-road data remains readable.
5. Roll back by reverting the frontend-only change; no persisted data requires migration.

## Open Questions

- None. The full grade explanation remains available but is not expanded in the primary overview by default.
