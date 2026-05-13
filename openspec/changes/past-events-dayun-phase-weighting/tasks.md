## 1. Context Model

- [x] 1.1 Add an internal `YearSignalContext` or equivalent structure carrying `year_in_dayun`, `dayun_phase`, and selected JinBuHuan phase rating/description.
- [x] 1.2 Add a helper that maps liunian index 0-9 to `year_in_dayun` 1-10 and phase `gan` / `zhi`.
- [x] 1.3 Keep the existing `GetYearEventSignals` public function available as a compatibility wrapper.

## 2. Signal Generation

- [x] 2.1 Add a contextual signal-generation entrypoint used by past-events flows.
- [x] 2.2 Inject a `大运阶段` background signal when phase context and JinBuHuan description exist.
- [x] 2.3 Map JinBuHuan `吉/凶/平` to existing signal polarity constants.
- [x] 2.4 Make `gan` phase emphasize dayun heavenly-stem context and `zhi` phase emphasize dayun earthly-branch context without removing direct liunian signals.
- [x] 2.5 Ensure age `< 18` semantic remapping still takes precedence for user-facing categories.

## 3. Narrative Rendering

- [x] 3.1 Add `大运阶段` to narrative theme mapping as supporting context.
- [x] 3.2 Ensure `大运阶段` does not outrank specific yearly themes by default.
- [x] 3.3 Add plain-language sentence variants for auspicious, adverse, and neutral phase backgrounds.
- [x] 3.4 Include phase evidence in `RenderEvidenceSummary` without exposing raw technical text in default narrative.

## 4. Service Integration

- [x] 4.1 Update `GeneratePastEventsYears` to pass year-in-dayun and phase context when building yearly items.
- [x] 4.2 Update `GenerateDayunSummariesStream` to include phase context in the 10-year JSON sent to AI.
- [x] 4.3 Optionally expose additive `year_in_dayun` and `dayun_phase` fields in the past-events years response if useful for debugging or UI.
- [x] 4.4 Keep existing frontend rendering functional when those additive fields are absent.

## 5. Tests

- [x] 5.1 Add unit tests for phase derivation: positions 1-5 are `gan`, positions 6-10 are `zhi`.
- [x] 5.2 Add unit tests proving qian rating is used only in front five years and hou rating only in back five years.
- [x] 5.3 Add narrative tests proving specific yearly themes remain dominant over neutral phase context.
- [x] 5.4 Add tests for adverse late-phase context appearing as secondary caution when relevant.
- [x] 5.5 Run `go test ./pkg/bazi`.
- [x] 5.6 Run focused service tests if touched.

## 6. Review

- [x] 6.1 Manually inspect a known chart such as `1996-02-08 20:00` and confirm the same dayun shows a reasonable early/late phase distinction.
- [x] 6.2 Confirm no per-year AI calls are introduced.
- [x] 6.3 Confirm API changes, if any, are additive and do not break `PastEventsPage`.
