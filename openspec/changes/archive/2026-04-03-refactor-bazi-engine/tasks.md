## 1. Backend: Lunar-Go Engine Integration

- [x] 1.1 Add `github.com/6tail/lunar-go` to go.mod dependencies
- [x] 1.2 Refactor `pkg/bazi/engine.go` to use `lunar-go` for solar terms, Bazi charting, and True Solar Time correction based on longitude
- [x] 1.3 Update the `BaziResult`, `DayunItem`, and `WuxingStats` structs in `pkg/bazi` to include exact timeframes and hidden stems (地支藏干)
- [x] 1.4 Remove the algorithmic Yongshen/Jishen logic from the engine entirely

## 2. Backend: API and LLM Prompt Adjustments

- [x] 2.1 Update `CalculateInput` in `internal/handler/bazi_handler.go` to accept an optional `longitude` float parameter
- [x] 2.2 Pass the new `longitude` parameter into `bazi.Calculate`
- [x] 2.3 Refactor the LLM prompt in `internal/service/report_service.go` to incorporate the new hidden stems data and mandate a Chain-of-Thought (CoT) sequence for deducing Yongshen

## 3. Frontend: Location Selection and State

- [x] 3.1 Update `CalculateInput` interface in `frontend/src/lib/api.ts` to include an optional `longitude: number`
- [x] 3.2 Modify `HomePage.tsx` to add a "Birthplace / Province" dropdown selector that maps to approximate longitudes
- [x] 3.3 Ensure the `longitude` is correctly passed to the `calculate` and `generateReport` endpoints natively

## 4. Frontend: Professional View Updates

- [x] 4.1 Update the `BaziResult` interface in `ResultPage.tsx` to match the backend's new format (including hidden stems)
- [x] 4.2 Adjust the "Professional View" UI in `ResultPage.tsx` to cleanly display the newly calculated Hidden Stems (地支藏干) underneath each pillar's branch
- [x] 4.3 (Optional) Minor styling adjustments for the hidden stems text in `ResultPage.css`
