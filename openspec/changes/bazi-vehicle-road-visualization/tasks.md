## 1. Backend Data Model and Scoring

- [x] 1.1 Add vehicle/road profile structs to the Bazi package with JSON fields for grade, score, labels, summaries, tags, phases, and evidence.
- [x] 1.2 Implement natal vehicle scoring from day-master strength, Ming Ge clarity, Tiaohou completeness, Yongshen/Jishen confidence, and natal risk/support modifiers.
- [x] 1.3 Implement Dayun road scoring from Jin Bu Huan phase ratings, Dayun stem/branch Yongshen/Jishen polarity, favorable/adverse Ten Gods, Di Shi, and compact modifier signals.
- [x] 1.4 Add grade and road-type mapping helpers with stable thresholds for `S/A/B/C/D` and `highway/main_road/mountain_road/muddy_road/construction`.
- [x] 1.5 Attach `vehicle_profile` and `dayun_roadmap` to `BaziResult` after existing Ming Ge, Tiaohou, Gong Jia, Ten God relation, and favorable/adverse Ten God data are available.
- [x] 1.6 Add backend unit tests for score bounds, grade thresholds, evidence presence, missing optional signal fallback, and one representative good/neutral/adverse Dayun road case.

## 2. Report Context Integration

- [x] 2.1 Extend report prompt template data or prompt serialization to include vehicle profile and current Dayun road context when available.
- [x] 2.2 Update Bazi report prompt wording so AI explains algorithm-calculated vehicle/road labels without recalculating or overriding score/grade/road type.
- [x] 2.3 Add service-level tests that verify prompts include vehicle-road context when present and do not ask AI to invent missing profile data when absent.

## 3. Frontend Result Overview

- [x] 3.1 Extend frontend Bazi result types to include `vehicle_profile` and `dayun_roadmap`.
- [x] 3.2 Add a compact "命盘座驾" overview card with grade, vehicle type, summary, tags, and evidence entry point.
- [x] 3.3 Add a compact "当前路况" overview card that resolves the current Dayun road by Gregorian year and displays road type, phase summary, and tags.
- [x] 3.4 Add professional evidence rendering for vehicle and road profiles without hiding existing Ming Ge, Yongshen, Tiaohou, or Dayun details.
- [x] 3.5 Add fallback rendering so old saved charts without vehicle-road fields continue to display the existing overview without runtime errors.

## 4. Dayun Timeline Visualization

- [x] 4.1 Pass `dayun_roadmap` into `DayunTimeline` and match road items by `dayun_index`.
- [x] 4.2 Display compact road labels on Dayun cards while preserving desktop ten-card row layout and mobile two-column layout.
- [x] 4.3 Display front-five-year and back-five-year road phases in the selected Dayun summary strip.
- [x] 4.4 Preserve existing Dayun selection, Liunian drawer opening, Shen Sha modal behavior, current markers, transition markers, and focus markers.

## 5. Verification

- [x] 5.1 Run backend Bazi package tests for the new scoring module.
- [x] 5.2 Run backend service tests covering report prompt changes.
- [x] 5.3 Run frontend static tests covering result overview, fallback behavior, and Dayun timeline road labels.
- [x] 5.4 Run frontend build or lint command used by the project.
- [x] 5.5 Manually inspect desktop and mobile result-page layouts for text fit, no overlap, no horizontal overflow, and clear distinction between absolute Dayun road context and relative yearly trend chart.

## 6. User Comprehension of Grades and Roads

- [x] 6.1 Replace value-loaded vehicle grade display labels with neutral Chinese configuration labels while preserving `S/A/B/C/D` as algorithmic markers and supporting old saved snapshots.
- [x] 6.2 Add an inline expandable explanation that defines the vehicle/road metaphor and every grade and road label in ordinary-user language.
- [x] 6.3 Add a one-line plain-language explanation to the current-road card.
- [x] 6.4 Add focused frontend assertions and rerun the relevant verification commands.

## 7. Visible Grade Scale

- [x] 7.1 Render the full S-D grade scale and ordinary-user explanations directly in the vehicle card, with the current grade identified.
- [x] 7.2 Keep Dayun road definitions available as secondary detail without hiding the grade scale.
- [x] 7.3 Update focused frontend assertions and verify the production build and responsive result layout.

## 8. Tiaohou-First Natal Tiering

- [x] 8.1 Replace the natal base-score model with a Tiaohou-first gate and an independently evaluated Fuyi Yongshen score; remove Ten God confidence as a score input.
- [x] 8.2 Apply deterministic grade ceilings for unresolved or hidden-only urgent Tiaohou and unusable Fuyi Yongshen, then map grades to S/A/B/C/D natal-chart tier labels.
- [x] 8.3 Update vehicle summaries, report context, and the visible grade guide to describe the revised natal-chart tiers and their criteria.
- [x] 8.4 Add backend regression tests for Tiaohou/Fuyi gates, grade ceilings, and score bounds; update frontend assertions for the new labels and guidance.
- [x] 8.5 Run focused backend/frontend tests, production build, strict OpenSpec validation, and inspect result rendering.

## 9. Grade-Mapped Vehicle Classes

- [x] 9.1 Map the final natal grade to a stable primary vehicle class and remove Ming Ge or day-master conditions from primary vehicle-type selection.
- [x] 9.2 Add an optional Ming Ge-derived driving-style field for professional evidence without using it as the primary vehicle class.
- [x] 9.3 Update result-page types and rendering so the grade-mapped vehicle class is primary and the driving style appears only in professional mode.
- [x] 9.4 Add focused backend and frontend tests for grade-to-vehicle mapping, Ming Ge isolation, and professional driving-style rendering.
- [x] 9.5 Run focused tests, production build, strict OpenSpec validation, and inspect desktop/mobile result rendering.
