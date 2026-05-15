## Context

The existing `DayunTimeline` was built incrementally inside the Bazi result page. It now has some mockup-inspired pieces, but the composition still follows the old result-page card system: the page title, Dayun cards, Liunian cards, summary area, and mobile layout are not treated as one coherent "大运时间轴" experience.

The supplied reference is a complete screen design. Desktop shows a large bordered timeline panel with ten Dayun cards, an embedded Liunian panel, a Dayun summary strip, and a footer note. Mobile shows a phone-style timeline view with its own top bar, compact metadata, two-column Dayun cards, two-column Liunian cards, and bottom summary control.

Current integration points:

- `frontend/src/components/DayunTimeline.tsx` owns Dayun/Liunian interaction and Liuyue drawer entry.
- `frontend/src/pages/ResultPage.tsx` wraps the timeline in a generic `card` and `section-title`.
- `frontend/src/pages/ResultPage.css` holds both existing result-page layout and timeline-specific CSS.
- Backend calculation currently returns a Dayun array that may contain fewer than the ten periods shown in the mockup.

## Goals / Non-Goals

**Goals:**

- Recreate the supplied Dayun timeline desktop and mobile layout as closely as the existing app shell allows.
- Make the timeline module self-contained enough that it does not inherit unwanted result-page card proportions.
- Display ten Dayun cards and ten Liunian cards when timeline data is available.
- Preserve current interactions: selecting a Dayun, opening Liuyue detail, and clicking Shensha labels where supported.
- Keep styling in plain CSS variables and project CSS files; do not introduce a UI framework.
- Add static and browser verification that checks layout shape, responsive card counts, and overflow behavior.

**Non-Goals:**

- Rework the whole result page outside the timeline section.
- Replace the Bazi calculation engine beyond the minimum needed to provide ten displayable Dayun periods.
- Add new paid features, report generation behavior, or AI interpretation text.
- Introduce screenshot-diff tooling unless implementation finds current browser checks insufficient.

## Decisions

### Decision 1: Treat DayunTimeline as a full design module, not a child card

The result page should render the timeline with a mockup-specific wrapper instead of a generic `card` section. `DayunTimeline` should own its internal title, metadata row, Dayun strip, Liunian panel, summary strip, and disclaimer.

Rationale: the reference visual depends on one large container with consistent padding, borders, and inner section rhythm. Keeping the existing outer `card` and `section-title` forces mismatched spacing and duplicate titles.

Alternative considered: keep the existing result-page `card` and tune CSS. This has already produced partial matches but not a 1:1 screen.

### Decision 2: Use separate desktop and mobile composition rules

Desktop and mobile should share data and component helpers, but their layout rules should intentionally differ:

```text
Desktop
┌──────────────────────────────────────────────┐
│ title + metadata                             │
│ [10 Dayun cards in one row]                  │
│ [Liunian panel with 10 cards in one row]     │
│ [summary strip + tags]                       │
└──────────────────────────────────────────────┘

Mobile
┌──────────────────┐
│ app-like top bar │
│ metadata         │
│ [2-col Dayun]    │
│ [2-col Liunian]  │
│ summary control  │
└──────────────────┘
```

Rationale: the mobile mockup is not a narrow version of desktop. It has a phone-frame visual and a denser two-column card grid. CSS breakpoints should encode that instead of relying on automatic wrapping.

Alternative considered: use one responsive grid for both. This is simpler but does not match the supplied mobile composition.

### Decision 3: Compact cards must prioritize visual parity over extra explanation

Dayun cards and Liunian cards should show only the content visible in the mockup: index/year, age range or age, ganzhi, ten-god label, date range or shensha line, and state badges. Long deterministic explanations should not appear in the card body.

Rationale: the target design is a timeline overview. Explanatory text belongs in detail drawers or summary areas, not in every card.

Alternative considered: reading cards with ten-god explanations. This improves semantic clarity but conflicts with the reference density.

### Decision 4: Normalize timeline data before rendering

The component should receive or derive a display model with:

- exactly ten Dayun slots when possible,
- exactly ten Liunian cards for the selected Dayun,
- explicit state fields for current, transition, and focus years,
- summary text and tags for the selected Dayun.

If backend data cannot provide ten Dayun periods, implementation should either extend the backend sequence or derive a safe display-only continuation from the same calculation rules.

Rationale: the mockup relies on a fixed ten-card rhythm. Rendering nine cards changes the spacing and breaks visual parity.

Alternative considered: accept backend count as-is. That preserves current API behavior but cannot satisfy "一比一复刻".

### Decision 5: Verification includes browser layout metrics

Static tests should lock class names and responsive rules, but implementation should also run browser checks for:

- desktop Dayun cards in one row,
- desktop Liunian cards in one row,
- mobile Dayun cards in two columns,
- mobile Liunian cards in two columns,
- no horizontal document overflow.

Rationale: this change is visual. Static regex tests are useful but cannot prove the layout matches at runtime.

Alternative considered: only use unit/static tests. This missed earlier visual mismatches.

## Risks / Trade-offs

- [Risk] Pixel-perfect parity is constrained by the existing top navigation, fonts, and result-page background. -> Mitigation: define "1:1" as the timeline module and mobile phone-view composition, not the entire browser chrome.
- [Risk] Backend may only return nine Dayun periods for some charts. -> Mitigation: add a task to verify the engine contract and extend or normalize to ten display periods.
- [Risk] A fixed ten-card row can become cramped on narrower desktop widths. -> Mitigation: set a minimum supported desktop content width and use mobile composition below the breakpoint.
- [Risk] Mobile phone-frame styling may conflict with the existing bottom navigation. -> Mitigation: reserve safe-area spacing and verify with the current mobile page shell.
- [Risk] Shensha text can overflow compact cards. -> Mitigation: cap visible labels, use ellipsis/compact chips, and keep full detail in existing annotation/drawer flows.

## Migration Plan

1. Add failing tests for the target desktop and mobile timeline composition.
2. Normalize the Dayun/Liunian display model and verify ten-card data requirements.
3. Refactor `DayunTimeline` markup around the mockup sections.
4. Move timeline-specific styling into clear CSS blocks with desktop/mobile rules.
5. Adjust `ResultPage` wrapper so the timeline owns its visual container.
6. Run static tests, lint, build, and browser checks at desktop and mobile viewports.
7. Rollback by restoring the previous `DayunTimeline` component and CSS block; no database migration is expected.

## Open Questions

- Should the mobile mockup be rendered inside the normal result page scroll, or should it become a dedicated `/result/dayun` sub-view with its own top bar?
- Should the tenth Dayun period be produced by backend calculation for all charts, or may the frontend derive a display-only tenth slot when the backend returns nine?
- Should the "大运总览" copy be deterministic from JinBuHuan/ten-god data, or hardcoded as a first-pass UI summary?
