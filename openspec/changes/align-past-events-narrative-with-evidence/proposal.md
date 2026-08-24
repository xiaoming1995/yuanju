## Why

Past-events year cards can show accurate technical evidence while the visible year narrative remains too broad or generic. This makes the page less useful for users who do not understand bazi terminology, because the most precise information is hidden in professional evidence instead of being translated into everyday meaning.

This change aligns each year narrative with its strongest algorithmic evidence, making the content longer where useful, easier for non-specialists to understand, and still auditable by professional users.

## What Changes

- Add evidence-aligned year narratives for past-events year cards.
- Require each meaningful year narrative to explain at least one concrete bazi trigger in plain language when evidence is available.
- Expand year narratives from short generic comments into structured readable explanations:
  - overall yearly tendency,
  - why the chart points that way,
  - likely life areas affected,
  - practical caution or suggested stance.
- Keep professional evidence visible as secondary detail for auditability.
- Add fallback behavior so weak or unsupported generated text does not replace more accurate evidence-based wording.
- Preserve existing past-events APIs, deterministic signal generation, dayun segmentation, and progressive generation behavior.

## Capabilities

### New Capabilities

- `past-events-evidence-aligned-narratives`: Defines how past-events yearly narratives must be grounded in technical evidence while remaining understandable to users who do not know bazi.

### Modified Capabilities

- None.

## Impact

- Backend bazi narrative rendering under `backend/pkg/bazi`, especially year narrative and evidence summary rendering.
- Backend report service paths that populate `PastEventsYearItem.Narrative` and dayun summary year fallback narratives.
- AI prompt and validation paths only if needed to prevent generic generated year text from overriding evidence-aligned deterministic text.
- Frontend past-events year card display if needed to surface evidence-aligned narratives before expandable professional evidence.
- Tests for narrative content, fallback behavior, and readability constraints.
- No database migration, route change, provider change, or bazi signal algorithm change is expected.
