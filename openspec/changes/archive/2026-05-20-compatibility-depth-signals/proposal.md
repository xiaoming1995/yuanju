## Why

The current compatibility engine gives useful high-level scores, but the underlying evidence set is still too narrow for users who expect a professional bazi relationship reading. After the consulting UX work adds relationship context and question-aware reports, the next gap is making the deterministic compatibility signals deeper, more explainable, and strong enough to support those consultation claims.

## What Changes

- Expand compatibility analysis beyond day master, five-element balance, spouse palace, spouse star, simple branch clashes, and romance/lonely shensha.
- Add directional ten-god relationship signals so the system can explain how each person experiences the other, such as support, pressure, attraction, competition, expression, or dependency.
- Add favorable/unfavorable element support signals that evaluate whether one person's chart tends to补足 or aggravate the other person's structural imbalance.
- Expand heavenly-stem and earthly-branch interaction coverage, including stem combinations, branch combinations, half-combinations, directional meetings, clashes, punishments, harms, and breaks where the current chart data can support them.
- Add relationship-pattern signals for communication style, conflict trigger, attachment/security pattern, reality pressure, and long-term repairability.
- Optionally add a conservative timing layer from big-luck / current-year context when available, limited to risk windows and focus areas rather than exact event dates.
- Update compatibility score computation and evidence taxonomy so score changes are traceable, bounded, and testable.
- Feed the deeper evidence into AI prompt data and result-page professional evidence sections without changing the public compatibility creation flow from the consulting UX change.
- Do not introduce deterministic marriage, breakup, pregnancy, affair, or exact-date predictions.

## Capabilities

### New Capabilities

- `compatibility-depth-signal-engine`: Computes deeper deterministic bazi compatibility evidence, including ten-god interaction, favorable-element support, expanded gan-zhi interactions, and relationship-pattern signals.
- `compatibility-explainable-compatibility-scoring`: Converts depth signals into bounded dimension scores and user-facing evidence summaries with traceable score contribution.
- `compatibility-professional-evidence-reporting`: Includes the deeper evidence in API detail, AI prompt data, and result-page professional sections while preserving a consultation-first reading path.

### Modified Capabilities

- None.

## Impact

- Backend bazi compatibility engine and related tests in `backend/pkg/bazi`.
- Compatibility model/service/repository/handler response shape if evidence metadata needs new fields.
- AI prompt data and fallback structured report generation in `backend/internal/service/compatibility_service.go`.
- Frontend compatibility result page and API types for professional evidence rendering.
- Snapshot/static tests around compatibility result copy, evidence grouping, and score explanation.
- No database migration is expected unless evidence metadata must be persisted in a new structured shape.
