## 1. Stem-Level Overview Model

- [x] 1.1 Add a result-page view model that derives first-choice, Fuyi-usable, and Fuyi-adverse summary rows from existing optional `stem_guidance` groups without changing backend assessment data.
- [x] 1.2 Define explicit no-primary and legacy-response fallbacks so no first-choice stem is invented and older charts retain their current element-level conclusion.

## 2. Result Overview Presentation

- [x] 2.1 Replace the new-response aggregate Fuyi badges in the result header with exact rows for `首取`、`扶抑可用`、and `扶抑慎用`, including the primary stem's Ten God and shared source context.
- [x] 2.2 Keep adjustment-only conflict items out of the header while preserving the existing professional detail surface and its conflict explanation.
- [x] 2.3 Add compact responsive styling for desktop and narrow mobile widths, with stable hierarchy, five-element stem color, and no overlap with the pillar heading or reading-mode controls.

## 3. Regression Coverage and Verification

- [x] 3.1 Add frontend tests for the 1996-02-08 20:00-style primary, usable, and adverse exact-stem summary and for the absence of aggregate-only badges when guidance is usable.
- [x] 3.2 Add frontend tests for no-primary and missing-guidance fallbacks, including the boundary that adjustment-only stems are not shown as future favorable conclusions.
- [x] 3.3 Run focused frontend tests, production build, and browser verification at desktop and mobile widths; verify no natal, Dayun-road, or annual-flow scoring output changes.
