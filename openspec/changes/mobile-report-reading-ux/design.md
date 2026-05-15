## Overview

This change keeps existing structured report fields (`analysis`, `chapters`, `yongshen`, `jishen`) and changes their reading presentation. The report should open with a concise digest, then let the user expand deeper sections.

## Reading Structure

Structured reports will render in this order:

1. Report digest card: overall summary, yongshen, jishen, and first actionable reading cue.
2. Mode switcher: keep the existing brief/detail switch.
3. Terminology strip: short explanations for common terms used on the page.
4. Chapter list: each chapter is an expandable `details` section with title, brief, and detail-aware content.
5. Full analysis overview: shown in detail mode after the digest and before/inside the chapter reading flow.
6. Action bar: save image, export PDF, view history, past-events entry, and re-calculate.

Legacy text reports continue using the old parsed-section fallback, but they still receive the bottom action area and mobile-safe spacing.

## Component Approach

- Keep helpers local to `ResultPage.tsx` because this report rendering is page-specific.
- Add small derivation helpers:
  - digest item creation from structured report fields.
  - terminology list from current page concepts.
- Keep CSS in `ResultPage.css`; avoid new dependencies.

## Mobile Rules

- The result page must reserve bottom space for mobile bottom navigation.
- Report actions should wrap into full-width tap targets on small screens.
- Long report sections must not force horizontal scrolling.
- Chapter summaries must remain readable without opening every detail.

## Testing

- Add static tests for digest card, chapter details, glossary, action bar, and mobile safe-area CSS.
- Re-run lint, build, frontend node tests, and browser check on mobile width.
