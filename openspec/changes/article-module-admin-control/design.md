## Context

The article inspiration feature currently has three separate operational surfaces:

- User-facing routes and APIs: `/articles`, `/articles/:id`, and `/api/articles...`
- Admin curation: `/admin/articles` and `/api/admin/articles...`
- Background collection: manual and scheduled collection controlled by `article_collection_config.enabled`

The missing control is a top-level module switch. Operators need one place to stop the article module from serving users and from running collection work, without losing access to admin review and configuration screens needed to turn the module back on.

## Goals / Non-Goals

**Goals:**

- Add a single admin-controlled article module switch for whole-module operation.
- Keep admin article management available even when the user-facing module is closed.
- Stop scheduled collection, manual collection, and retry collection when the module switch is closed.
- Keep detailed collection timing, result limits, quality, and auto-publish settings intact for when the module switch is open.
- Use the existing `system_settings` repository pattern to avoid a new configuration subsystem.
- Make direct user access fail clearly when the module is closed.

**Non-Goals:**

- No role-based or per-user article access rules.
- No changes to article status lifecycle, quality scoring, auto-publish eligibility, or AI analysis output.
- No frontend feature-flag framework for every product module in this change.

## Decisions

### Store module availability in `system_settings`

Use a new boolean setting key such as `articles_module_enabled`, with an additive migration that inserts the default value `true`.

Alternatives considered:
- Add a column to `article_collection_config`: rejected because collection configuration and user-facing availability are different concepts.
- Add a new `article_module_config` table: rejected as unnecessary for a single boolean setting and inconsistent with registration control.
- Hard-code an environment variable: rejected because operators need runtime admin control.

### Gate user-facing article handlers on the backend

The ordinary-user article list, detail, and original-click handlers should check the module setting before reading or mutating article data. When disabled, they should return `403` with a clear error such as `资讯模块暂未开放`.

This backend gate is required even if the frontend hides navigation, because direct URLs and API clients can still call the endpoints.

### Expose both public and admin settings APIs

Add a lightweight ordinary-user-readable settings endpoint for frontend visibility, and an admin endpoint for reading/updating the setting. The setting is named generically enough to serve as the article module top-level switch.

Expected shape:

- `GET /api/articles/settings` returns `{ "module_enabled": true }`
- `GET /api/admin/articles/module-settings` returns `{ "module_enabled": true }`
- `PUT /api/admin/articles/module-settings` accepts `{ "module_enabled": false }`

The public settings endpoint can be unauthenticated because it only reveals feature availability and is needed before deciding whether to show login-gated navigation.

### Treat module availability as the collection hard gate

The admin article management page should show and update module availability, but existing article management tabs must still be readable and configurable while the module is closed. Manual collection, retry collection, and scheduled collection must first check the top-level module switch. When the module is disabled, no new collection task should be created and scheduled collection should not update `last_run_at`.

Collection configuration remains meaningful only after the module switch is enabled: `article_collection_config.enabled=false` still stops scheduled collection even if the module is enabled, while `articles_module_enabled=false` stops all collection execution regardless of collection config.

### Frontend should hide entry and handle direct routes

Navbar should not render the "资讯" entry when the setting is disabled. Direct visits to `/articles` and `/articles/:id` should render a closed-state message or redirect to a stable page after displaying the module is not open, instead of showing a generic request failure.

## Risks / Trade-offs

- [Risk] Frontend and backend availability can briefly disagree while settings are loading. → Treat backend as authoritative and make pages handle `403` closed responses.
- [Risk] A default-disabled migration would unexpectedly hide an already deployed module. → Default `articles_module_enabled` to `true`.
- [Risk] Admins may confuse module availability with scheduled collection. → Display module switch separately from "定时配置" and label it as the top-level module switch.
- [Risk] More feature switches may be needed later. → Keep naming and response shape compatible with future feature settings without introducing a broad framework now.

## Migration Plan

1. Add a new migration that ensures `system_settings` exists and inserts `articles_module_enabled=true` if missing.
2. Deploy backend code that reads the setting with default `true`; this makes rollback tolerant if the migration has not run yet.
3. Deploy frontend code that consumes the setting and handles closed responses.
4. Rollback is safe: old code ignores the setting row, and the row can remain in `system_settings`.

## Open Questions

- Should mobile bottom navigation eventually include "资讯" when enabled? It currently does not include an article tab, so this change should leave mobile bottom navigation unchanged unless product direction changes.
