## Why

Current Bazi results expose professional signals such as Ming Ge, Yongshen/Jishen, Tiaohou, Ten Gods, Jin Bu Huan Dayun ratings, and yearly event signals, but users still need to mentally connect those signals into an intuitive life-stage picture. The result page already distinguishes relative yearly trend movement inside a Dayun, yet it does not provide an absolute, cross-Dayun layer that explains "what kind of chart this is" and "what kind of road this decade is."

This change introduces a deterministic "vehicle and road" visualization layer: the natal chart is represented as a vehicle profile, each Dayun as a road condition, and each Liunian as the finer yearly weather or road event. The goal is to make existing algorithmic evidence easier for ordinary users to understand while preserving professional explainability.

## What Changes

- Add backend structured Bazi output for a natal `vehicle_profile`, including grade, score, vehicle type, summary, tags, and evidence.
- Add backend structured Dayun output for `dayun_roadmap`, including per-Dayun road score, road type, front/back five-year road phases, summary, tags, and evidence.
- Derive all grades and road labels from existing deterministic algorithm signals instead of asking the LLM to invent ratings.
- Update report prompt context so AI reports can reference the calculated vehicle profile and Dayun road map as explanation inputs, not as model-generated truth.
- Update the Bazi result page overview to surface "命盘座驾" and "当前路况" in ordinary-user language.
- Update the Dayun timeline to show each Dayun's road label and front/back five-year road phase.
- Add professional-mode evidence disclosure so users can inspect which Ming Ge, strength, Tiaohou, Yongshen/Jishen, Ten God, Jin Bu Huan, and phase signals contributed to the label.
- Preserve existing Bazi fields, report history, Dayun timeline behavior, and saved-chart compatibility.

## Capabilities

### New Capabilities

- `bazi-vehicle-road-visualization`: Provides deterministic natal chart vehicle profiling, Dayun road scoring, and result-page visualization requirements.

### Modified Capabilities

- `bazi-advanced-data`: Bazi calculation responses will include structured vehicle profile and Dayun road map data while preserving existing advanced fields.
- `dayun-timeline-design`: The Dayun timeline will display road-condition labels and phase indicators while preserving existing Dayun/Liunian interactions.
- `bazi-ai-reasoning`: AI report generation will receive the deterministic vehicle and road context and must explain it without recalculating or overriding it.

## Impact

- Backend:
  - `backend/pkg/bazi/engine.go`
  - new algorithm module under `backend/pkg/bazi/`
  - focused tests under `backend/pkg/bazi/`
  - report prompt assembly in `backend/internal/service/report_service.go`
- Frontend:
  - `frontend/src/pages/ResultPage.tsx`
  - `frontend/src/components/DayunTimeline.tsx`
  - possibly new helper module under `frontend/src/lib/`
  - CSS updates in result/timeline styles
- API:
  - Additive JSON response fields only; no breaking changes.
  - Old saved charts without these fields must degrade gracefully or be lazily derivable where existing chart data is sufficient.
- Data:
  - No database schema migration is required for the first version because chart snapshots already persist full result JSON.
