## Context

`ProfilePage` already receives enough data to guide returning users: account metadata, counts, recent charts, recent compatibility readings, and planned feature entries. The page can become a workbench without adding backend endpoints.

## Goals / Non-Goals

**Goals:**
- Surface a "continue analysis" entry that points to the latest available chart or compatibility reading.
- Turn stats into scan-friendly navigation cards for existing destinations.
- Reframe PDF and wallet features as planned capabilities with non-clickable status badges.
- Keep mobile density controlled through existing page shell and single-column rules.

**Non-Goals:**
- Implement payment, wallet, report archive, or PDF template management.
- Add new API fields.
- Change authentication behavior.

## Decisions

1. **Use recent records for continuation.**
   - If a recent chart exists, prioritize continuing that chart because it leads to the richest result page.
   - If no chart exists but a compatibility reading exists, continue the latest compatibility reading.
   - If neither exists, point users to create a new chart.

2. **Stats link only to existing destinations.**
   - Chart count links to `/history`.
   - Compatibility count links to `/compatibility/history`.
   - AI report count remains a status card until a dedicated report archive exists.

3. **Keep planned features inert.**
   - Feature cards show a status badge and description but no payment/PDF route, preserving the existing no-payment guardrail.

## Risks / Trade-offs

- **Risk:** AI report count is not directly navigable.
  - **Mitigation:** Keep it clearly labeled as a status card instead of pretending a report archive exists.

- **Risk:** "Continue" could choose the wrong domain for users focused on compatibility.
  - **Mitigation:** Also keep explicit quick actions for history, compatibility history, and new chart creation.
