## Why

The Dayun overview independently recomputes the Day Master strength from a flat five-element percentage and can describe a natal strong chart as weak. Separately, the Dayun branch's Twelve Growth Stages currently give a universal road-score adjustment, rather than strengthening or weakening that branch according to its natal Fuyi polarity.

## What Changes

- Make the natal Fuyi assessment the only source of the Day Master strength used in Dayun interpretation; the conclusion is fixed for every Dayun of the same natal chart.
- Remove the frontend's simplified Day Master strength calculation and its strength-dependent Ten-God copy.
- Evaluate the Dayun stem and branch independently against natal Fuyi favorable/adverse elements.
- Reinterpret a Dayun branch's Twelve Growth Stage as an intensity modifier for that branch's favorable, adverse, or neutral Fuyi effect, rather than as a universal positive or negative score for the Day Master.
- Render concise Dayun prose that separates the stem's role from the branch's role and explains the branch-stage interaction without asserting a changed natal strength.
- Add regression coverage for the 1987-12-09 06:00 male chart and representative favorable/adverse branch-stage cases.

## Capabilities

### New Capabilities
- `natal-dayun-strength-boundary`: Keeps natal Day Master strength immutable across Dayun interpretation and applies branch Twelve Growth Stages only to the branch's natal-Fuyi polarity.

### Modified Capabilities
- `dayun-timeline-design`: Dayun summary text must present stem and branch effects without recalculating or changing the natal Day Master strength.

## Impact

- Backend: `backend/pkg/bazi/vehicle_profile.go`, Dayun-road evidence and calculation tests.
- Frontend: `frontend/src/lib/dayunOverview.ts`, `frontend/src/components/DayunTimeline.tsx`, `frontend/src/pages/ResultPage.tsx`, and Dayun overview tests.
- Existing calculate-response fields remain compatible; the frontend consumes the existing `natal_assessment.fuyi.day_master_strength` field.
