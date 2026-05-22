## Why

The compatibility result page now surfaces stronger personality and action guidance, but the lower result sections still hide important evidence by default and show an oversized empty "深度解读" card when no AI report exists. This makes the page feel incomplete even when structured compatibility data is already available.

## What Changes

- Make the key judgment evidence section easier to trust by exposing core reasoning and linked evidence without requiring every card to be manually expanded.
- Improve the "深度解读" module so the absent-report state becomes a compact, actionable prompt instead of a large empty card.
- Preserve generated deep report content, but give it clearer hierarchy and section styling when present.
- Rebalance the professional details area so summary-level technical data is visible enough to support trust, while dense chart details can remain collapsible.
- Keep backend APIs, scoring, report generation behavior, and AI prompt behavior unchanged.

## Capabilities

### New Capabilities
- `compatibility-result-evidence-disclosure`: Covers default visibility, hierarchy, and interaction behavior for key judgment evidence and professional supporting data on the compatibility result page.
- `compatibility-result-deep-report-display`: Covers absent, loading, error, raw, and structured display states for the compatibility deep reading module.

### Modified Capabilities
- None.

## Impact

- Frontend compatibility result page and CSS: `frontend/src/pages/CompatibilityResultPage.tsx`, `frontend/src/pages/CompatibilityResultPage.css`
- Static UX tests under `frontend/tests/`
- No database, backend API, authentication, scoring model, dependency, or AI prompt changes are expected.
