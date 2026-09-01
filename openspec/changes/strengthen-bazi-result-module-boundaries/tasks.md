## 1. Overview Structure

- [x] 1.1 Reorganize the result overview DOM into identity, conclusion, summary, and auxiliary utility regions without changing existing result data or actions.
- [x] 1.2 Group birth information, four pillars, stem-level guidance, ming ge label, and reading-mode controls in the identity region.
- [x] 1.3 Move saved-chart naming and related available chart utilities into the identity region's auxiliary action area.
- [x] 1.4 Keep the verdict and its existing evidence navigation in a dedicated conclusion region before the summary panels.

## 2. Visual System And Responsive Layout

- [x] 2.1 Define shared page-level surface, border, spacing, heading, and status-token rules for the result overview.
- [x] 2.2 Style the identity and conclusion regions as distinct, restrained surfaces that preserve the existing starfield as a page background.
- [x] 2.3 Style natal vehicle and current road as independently bounded, top-aligned summary panels with shared visual conventions.
- [x] 2.4 Implement wide and narrow viewport layouts so the overview is two-column only where appropriate and remains readable without overflow on mobile.
- [x] 2.5 Remove or consolidate superseded overview CSS rules to prevent later overrides from weakening the new boundaries.

## 3. Verification

- [x] 3.1 Update result-page visual regression tests for the required overview regions, reading order, utility placement, and responsive classes.
- [x] 3.2 Run frontend production build and relevant result-page test suites.
- [x] 3.3 Visually verify simple and professional modes on desktop and mobile with complete and incomplete result data.
