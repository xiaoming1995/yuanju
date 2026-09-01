## 1. Road Meaning and Summary Copy

- [x] 1.1 Refactor the backend Dayun roadmap summary so the ten-year composite road and Jin Bu Huan phase prompts are named separately.
- [x] 1.2 Add backend regression coverage for a Dayun whose composite road label differs from its front/back Jin Bu Huan labels.

## 2. Active Dayun Presentation

- [x] 2.1 Add shared frontend presentation helpers for composite-road wording, phase time ranges, and Jin Bu Huan rating labels with old-data fallbacks.
- [x] 2.2 Redesign the result overview current-Dayun card into period identity, ten-year composite conclusion, and front/back phase-prompt sections.
- [x] 2.3 Add a compact road-scope explanation entry point that routes users to existing detailed evidence when available.

## 3. Timeline Hierarchy

- [x] 3.1 Update the selected Dayun summary strip to use the same composite-road-first hierarchy and remove equivalent front/back road-grade wording.
- [x] 3.2 Keep compact Dayun timeline cards focused on identity, Gan/Zhi Ten Gods, and the composite road badge while preserving selection, Liunian, and Shen Sha interactions.
- [x] 3.3 Adjust responsive CSS so the active-Dayun and phase sections remain aligned, readable, and non-overflowing on desktop and mobile.

## 4. Verification

- [x] 4.1 Add or update frontend static tests for scope labels, phase timing, old-data fallbacks, and compact timeline card content.
- [x] 4.2 Run focused Go tests, frontend tests, production build, and `openspec validate clarify-dayun-road-presentation --strict`.
- [x] 4.3 Verify the 1996-02-08 20:00 `癸巳` case in a browser at desktop and mobile widths, including a composite `泥路` with phase `凶` prompts and no horizontal overflow.
