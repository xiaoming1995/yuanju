## 1. Backend Settings Foundation

- [x] 1.1 Add an additive migration that inserts `articles_module_enabled=true` into `system_settings` when missing.
- [x] 1.2 Add a repository setting constant for article module availability and reuse the existing boolean setting helpers.
- [x] 1.3 Add backend tests covering default enabled behavior, persisted disabled behavior, and invalid stored boolean handling if applicable.

## 2. Backend Article Module APIs

- [x] 2.1 Add a public article settings handler returning `module_enabled`.
- [x] 2.2 Add authenticated admin handlers for reading and updating article module settings.
- [x] 2.3 Register routes for `GET /api/articles/settings`, `GET /api/admin/articles/module-settings`, and `PUT /api/admin/articles/module-settings`.
- [x] 2.4 Add handler tests for public settings, admin update success, and non-admin rejection.

## 3. Backend User Article Access Gate

- [x] 3.1 Gate ordinary-user article list, detail, and original-click handlers on article module availability.
- [x] 3.2 Ensure disabled module responses use HTTP 403 and a clear module-closed error message.
- [x] 3.3 Ensure disabled detail and original-click requests do not increment view counts or write click logs.
- [x] 3.4 Add backend tests for enabled list/detail behavior and disabled list/detail/original-click behavior.

## 4. Frontend API Wiring

- [x] 4.1 Add article module settings methods and TypeScript response types to `frontend/src/lib/api.ts`.
- [x] 4.2 Add admin article module settings methods and payload types to `frontend/src/lib/adminApi.ts`.
- [x] 4.3 Ensure frontend article API error handling can distinguish module-closed responses from generic missing article failures.

## 5. Frontend User Experience

- [x] 5.1 Update `Navbar` to read article module availability and hide "资讯" when disabled.
- [x] 5.2 Update `ArticlesPage` to show a module-closed state for disabled direct route access.
- [x] 5.3 Update `ArticleDetailPage` to show a module-closed state when the backend returns the disabled-module response.
- [x] 5.4 Preserve existing login requirements and article loading behavior when the module is enabled.

## 6. Frontend Admin Experience

- [x] 6.1 Add a user-facing module status card/control to `AdminArticlesPage` separate from scheduled collection settings.
- [x] 6.2 Wire the admin control to load, save, show dirty/saved state, and display errors.
- [x] 6.3 Make the control copy clearly distinguish frontend module availability from scheduled collection enablement.
- [x] 6.4 Ensure all existing admin article tabs remain usable when the module setting is disabled.

## 7. Whole Module Stop Behavior

- [x] 7.1 Gate scheduled collection on `articles_module_enabled` before reading collection due state or running the provider.
- [x] 7.2 Gate manual collection and collection retry on `articles_module_enabled` before creating collection tasks.
- [x] 7.3 Update admin module switch copy and collection buttons to present the switch as a whole-module stop/start control.
- [x] 7.4 Add backend tests for disabled-module scheduled collection skip and admin collection rejection.
- [x] 7.5 Add frontend test coverage for whole-module switch copy and stopped collection controls.

## 8. Verification

- [x] 8.1 Run backend article handler/repository tests covering the new setting and access gates.
- [x] 8.2 Run frontend static/unit tests covering navigation visibility, direct-route closed states, and admin control wiring.
- [x] 8.3 Run the relevant frontend build or lint target for changed files.
- [ ] 8.4 Manually verify: disable module in admin, confirm user nav hides "资讯", direct `/articles` shows closed state, APIs return 403, admin management and collection controls remain available, and collection actions are stopped.
- [ ] 8.5 Manually verify: re-enable module in admin, confirm published articles are visible again and collection actions can run according to collection configuration.
