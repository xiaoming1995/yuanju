## Why

The profile page currently lists account stats and recent records, but it does not clearly guide users back into their most recent work or distinguish available actions from upcoming paid/PDF capabilities. Improving this page strengthens retention after the mobile layout foundation is stable.

## What Changes

- Add a profile workbench area that highlights the next best action based on recent chart and compatibility data.
- Make account statistics navigable where existing routes support it.
- Clarify dormant monetization/PDF features as planned capabilities rather than active actions.
- Keep the change frontend-only and based on the existing `/api/user/profile` response.

## Capabilities

### New Capabilities
- `profile-workbench-ux`: The profile page presents a clearer user workbench with continuation actions, navigable stats, and planned feature status.

### Modified Capabilities
- None.

## Impact

- Affected frontend files: `ProfilePage.tsx`, `ProfilePage.css`, and profile tests.
- No backend, database, or API changes.
