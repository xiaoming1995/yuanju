## Why

The article inspiration module already has collection and curation tooling, but operators cannot control whether the frontend "资讯" module is open to users. This creates an all-or-nothing deployment problem: content can be collected and reviewed, yet the public user entry and APIs remain available whenever the code is deployed.

## What Changes

- Add an admin-controlled article module availability setting that governs ordinary-user article access and collection execution.
- Treat article module availability as the top-level operational switch for the whole article module, while preserving detailed collection, quality filtering, and auto-publish configuration for when the module is enabled.
- Allow administrators to keep using "资讯管理" while the user-facing module is closed.
- Hide or disable the normal user "资讯" navigation entry when the module is closed.
- Make user-facing article list/detail/original-click APIs reject access with a clear module-closed response when the module is closed.
- Add frontend closed-state handling for direct `/articles` and `/articles/:id` visits.
- Stop scheduled collection, manual collection, and collection retry while the module is disabled.
- Preserve existing article status lifecycle, quality scoring, auto-publish, AI analysis, and admin curation behavior when the module is enabled.

## Capabilities

### New Capabilities

- `article-module-control`: Admin-controlled availability of the user-facing article module, including setting persistence, admin API, frontend visibility, and closed-state access behavior.

### Modified Capabilities

- `article-content-library`: User article navigation and endpoints become conditional on the article module availability setting.
- `article-curation-workflow`: Admin article management gains module availability controls while remaining accessible when user-facing articles are closed.
- `article-collection-pipeline`: Scheduled, manual, and retry collection are additionally gated by the module switch so operators can stop the full module from one control.

## Impact

- Backend:
  - Add an additive migration for a new `system_settings` key such as `articles_module_enabled`.
  - Reuse the existing settings repository pattern used by registration control.
  - Add admin and user-facing article settings endpoints.
  - Gate ordinary-user article handlers on the module setting.
- Frontend:
  - Extend normal API client and admin API client for article module settings.
  - Update Navbar and article pages to consume module availability.
  - Add a module status control to the admin article management page.
- Tests:
  - Add backend handler/repository tests for setting defaults, admin updates, and user API rejection.
  - Add frontend tests for navigation visibility, admin control wiring, and direct-route closed states.
