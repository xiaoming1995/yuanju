## 1. Overview hierarchy

- [x] 1.1 Restructure `ResultPage` so the total verdict is rendered first, followed by a dedicated vehicle/current-road summary grid.
- [x] 1.2 Keep summary cards limited to their decision-level information and move grade, road, and professional explanatory content into labelled expandable sections after the grid.
- [x] 1.3 Preserve every existing vehicle grade, road guide, evidence item, reading-mode condition, and fallback state during the markup move.

## 2. Alignment and responsive styling

- [x] 2.1 Replace CSS ordering and stretched-grid behavior with semantic DOM order and top-aligned content-height summary cards.
- [x] 2.2 Standardize overview card headers, pills, padding, keyword spacing, and text wrapping without introducing nested cards.
- [x] 2.3 Implement the narrow-screen single-column sequence and verify expandable controls remain usable.

## 3. Verification

- [x] 3.1 Update focused frontend tests for overview order, progressive disclosure, professional evidence, and vehicle-road fallback behavior.
- [x] 3.2 Run the frontend build and focused tests.
- [x] 3.3 Inspect desktop and mobile result-page screenshots for top alignment, non-stretched road cards, readable text, and absence of overlap.
