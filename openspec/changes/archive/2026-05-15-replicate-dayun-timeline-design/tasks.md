## 1. Test Baseline

- [x] 1.1 Add static frontend tests for the desktop timeline wrapper, header metadata, Dayun strip, Liunian panel, summary strip, and footer disclaimer.
- [x] 1.2 Add static frontend tests that require desktop Dayun and Liunian grids to render ten compact cards in one row.
- [x] 1.3 Add static frontend tests that require mobile Dayun and Liunian grids to render in two columns.
- [x] 1.4 Add static frontend tests that reject verbose reading-card content inside compact Dayun and Liunian cards.

## 2. Data Contract

- [x] 2.1 Verify the backend Dayun calculation output count for representative charts and identify why the current UI can receive fewer than ten periods.
- [x] 2.2 Add backend or frontend normalization tests for producing ten displayable Dayun periods when the timeline design requires ten cards.
- [x] 2.3 Implement the minimum data normalization needed for ten Dayun cards and ten Liunian cards without changing existing persisted chart fields unexpectedly.
- [x] 2.4 Add deterministic selected-Dayun summary data for ten-god main qi, five-element main qi, and trend keywords.

## 3. Desktop Timeline Layout

- [x] 3.1 Refactor `ResultPage` so the Dayun timeline module owns its mockup-specific container instead of inheriting a generic card title layout.
- [x] 3.2 Refactor `DayunTimeline` markup into desktop sections: header metadata, Dayun card strip, Liunian panel, summary strip, and disclaimer.
- [x] 3.3 Implement desktop Dayun card content, selected/current state, gold border, top badge, and bottom active indicator.
- [x] 3.4 Implement desktop Liunian card content, current-year marker, transition ribbon, focus badge, and compact shensha line.
- [x] 3.5 Implement desktop Dayun summary strip with concise paragraph and compact tags.

## 4. Mobile Timeline Layout

- [x] 4.1 Add mobile-specific timeline composition matching the phone-style mockup top bar, metadata block, card grids, and bottom summary control.
- [x] 4.2 Style mobile Dayun cards as two-column compact cards with visible active/current state.
- [x] 4.3 Style mobile Liunian cards as two-column compact cards with visible current, transition, and focus markers.
- [x] 4.4 Ensure mobile safe-area and bottom navigation spacing do not obscure timeline content.

## 5. Interaction Preservation

- [x] 5.1 Preserve Dayun card click behavior so selecting a card updates the Liunian panel.
- [x] 5.2 Preserve Liunian card click behavior so the Liuyue drawer opens for the selected year and ganzhi.
- [x] 5.3 Preserve Shensha annotation click behavior for visible Shensha chips where annotations exist.
- [x] 5.4 Ensure keyboard focus and button semantics remain usable after the visual refactor.

## 6. Verification

- [x] 6.1 Run frontend static tests for the timeline design.
- [x] 6.2 Run `npm run lint`.
- [x] 6.3 Run `npm run build`.
- [x] 6.4 Run backend Bazi tests if Dayun normalization touches backend calculation.
- [x] 6.5 Use the in-app browser to verify desktop timeline card counts, one-row layout, and no horizontal overflow.
- [x] 6.6 Use the in-app browser to verify mobile two-column Dayun and Liunian grids, active markers, and safe-area spacing.
- [x] 6.7 Capture desktop and mobile screenshots for final visual comparison against the reference mockup.
