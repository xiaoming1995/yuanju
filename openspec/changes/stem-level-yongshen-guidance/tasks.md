## 1. Stem-Level Natal Guidance Engine

- [x] 1.1 Add additive natal stem-guidance types and JSON fields without changing legacy or element-level Fuyi fields.
- [x] 1.2 Derive ordered primary favorable stems from the intersection of day-stem adjustment requirements and Fuyi favorable elements.
- [x] 1.3 Derive non-duplicated Fuyi-only secondary stems, adjustment-only stems, and adverse stems with five-element, Ten God, source, and conflict detail.
- [x] 1.4 Handle neutral or incomplete Fuyi results without inventing stem-level favorable or adverse conclusions.

## 2. Result-Page Guidance Presentation

- [x] 2.1 Extend result-page result types to consume optional stem guidance while retaining element-only fallback behavior.
- [x] 2.2 Add a concise heavenly-stem priority summary to the Fuyi result presentation when a primary favorable stem exists.
- [x] 2.3 Present primary favorable, secondary favorable, adjustment-only, and adverse stem groups in progressive detail with sources, Ten Gods, and conflict wording.
- [x] 2.4 Preserve the existing simple-mode hierarchy, compact element badges, and accessibility behavior when no guidance payload is available.

## 3. Regression Coverage and Verification

- [x] 3.1 Add backend tests for shared-priority intersection, Fuyi-only stems, adjustment/Fuyi conflict handling, neutral fallback, and no duplicate output.
- [x] 3.2 Add a regression fixture for the 1995-10-12 noon chart that identifies 壬水 as primary, retains 甲木 as adjustment-only, and preserves existing Dayun-road evidence and score.
- [x] 3.3 Add focused frontend tests for concise priority rendering, detailed group rendering, and aggregate element-only fallback.
- [x] 3.4 Run targeted Go and frontend tests, production build, and browser verification for a chart with both shared-priority and conflicting stems.
