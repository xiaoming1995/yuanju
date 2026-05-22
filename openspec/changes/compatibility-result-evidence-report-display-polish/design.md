## Context

The compatibility result page already presents decision guidance, personality fit, validation planning, scores, evidence, deep report content, and professional chart details. The current lower-page experience has two UX gaps:

- Key judgment evidence is rendered as collapsed `<details>` cards, so users do not immediately see why a claim is credible.
- The deep report module uses the same large card weight even when no report exists, leaving a visually heavy empty state with no local action.

The change is frontend-only and must preserve existing APIs, report generation behavior, scoring, and professional evidence data contracts.

## Goals / Non-Goals

**Goals:**
- Make the most important evidence visible by default while keeping dense professional details manageable.
- Give the deep report module clear absent, loading, error, raw, and structured states.
- Keep the result page hierarchy focused on the completed compatibility reading before optional AI expansion.
- Improve mobile scanability without adding a UI framework or changing CSS architecture.

**Non-Goals:**
- No backend API changes.
- No scoring, prompt, or AI report content changes.
- No authentication or history behavior changes.
- No new dependency or design system migration.

## Decisions

1. **Expose evidence progressively instead of fully flattening all details.**
   - The first key judgment evidence item should be expanded by default, and each claim should show a compact reasoning preview even when collapsed.
   - Alternative considered: expand every evidence card by default. This would improve transparency but make the page noisy on mobile and duplicate dense professional data.

2. **Move report generation affordance into the deep report section as well as preserving the single generation action contract.**
   - The absent deep report state should explain what the AI report adds and provide the existing generation action in-place.
   - The result page must still avoid multiple competing generation buttons; if the dashboard keeps an action, the deep report section should either own the action or share the same single action location.
   - Alternative considered: leave the action only in the dashboard. This fails when users scroll to the empty deep report card.

3. **Use state-specific presentation for deep report content.**
   - No report: compact CTA/empty state.
   - Loading: clear progress copy and disabled action.
   - Error: inline recovery copy near the action.
   - Structured report: readable sections with summary, focus, dimensions, risks, and advice.
   - Raw report: readable fallback with pre-wrap formatting.

4. **Make professional details support trust without becoming the primary reading path.**
   - Keep dense chart details collapsible, but expose a summary line or compact metadata so the user understands what exists inside.
   - Alternative considered: default-open all professional details. This helps advanced users but competes with the main consultation flow.

## Risks / Trade-offs

- **Risk: Too much evidence visible by default increases page length.** → Mitigate by opening only the first/highest-priority claim and using compact evidence summaries.
- **Risk: Multiple report buttons confuse users.** → Mitigate by enforcing a single visible generation action in static tests.
- **Risk: Existing tests rely on component order.** → Mitigate by adding focused tests and preserving existing result-page hierarchy contracts.
- **Risk: Empty report state still feels like an error.** → Mitigate with explicit copy that the structured reading is already available and AI expansion is optional.
