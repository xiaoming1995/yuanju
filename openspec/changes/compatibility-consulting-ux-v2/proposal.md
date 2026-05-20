## Why

The current compatibility flow can calculate pair signals and render a result, but the experience still reads like a score dashboard. Users come to compatibility matching with a relationship decision in mind, so the product needs to ask for relationship context and answer "what should I do next?" before showing dense scores or professional evidence.

## What Changes

- Add lightweight relationship context to the compatibility creation flow: current relationship stage and the user's primary question.
- Persist that context with the compatibility reading so history, detail views, and AI reports can stay grounded in the user's situation.
- Rework the compatibility result hierarchy around a decision-first consultation: conclusion, core contradiction, next actions, stage risks, then scores and professional evidence.
- Rephrase the four dimensions into user-question language while keeping the underlying score fields unchanged.
- Make AI compatibility reports adapt their structure to the user's selected question, such as reconciliation, marriage suitability, long-term stability, recurring conflict, or general next-step guidance.
- Keep professional bazi evidence expandable and traceable, but make it a support layer rather than the primary reading path.
- Do not introduce precise breakup dates, deterministic marriage predictions, or object comparison in this change.

## Capabilities

### New Capabilities

- `compatibility-consulting-context`: Captures and persists relationship stage and primary user question for compatibility readings.
- `compatibility-decision-result-ux`: Presents compatibility results as a decision-oriented consultation before scores and professional details.
- `compatibility-question-aware-report`: Generates and renders AI compatibility reports that adapt to the user's selected relationship question.

### Modified Capabilities

- None.

## Impact

- Backend model/repository/API changes for optional compatibility context fields.
- Compatibility creation request/response shape changes, remaining backward compatible for existing callers.
- Compatibility AI prompt data and structured report schema extensions.
- Frontend compatibility input page, result page, static UX tests, and likely mobile layout tests.
- Existing compatibility readings must remain viewable; missing context should fall back to a neutral "general relationship judgment" mode.
