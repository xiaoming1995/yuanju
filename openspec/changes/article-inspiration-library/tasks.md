## 1. Database Schema

- [x] 1.1 Create an additive migration for article categories, tags, keywords, articles, article-tag links, collection tasks, collection task items, audit events, original click logs, and reserved favorites/full-text authorization fields.
- [x] 1.2 Add indexes for article status, category, source publish time, view count, normalized original URL hash, tag links, task status, and click log article/user lookup.
- [x] 1.3 Add migration tests or dry-run verification for the new SQL in the existing migration framework.

## 2. Backend Models And Repositories

- [x] 2.1 Add article domain structs for article, category, tag, keyword, collection task, task item, audit event, click log, AI config, and AI analysis payloads.
- [x] 2.2 Implement repository methods for category/tag/keyword CRUD and active taxonomy listing.
- [x] 2.3 Implement repository methods for article insertion, URL dedupe lookup, list filtering, detail fetch, counters, status transitions, and audit event writes.
- [x] 2.4 Implement repository methods for collection task/task item creation, status updates, counts, failure recording, and retry lookup.
- [x] 2.5 Implement repository methods for article AI configuration, AI analysis status, AI output persistence, and analysis failure recording.
- [x] 2.6 Add focused repository tests for dedupe, lifecycle transitions, audit creation, filter queries, and click log insertion.

## 3. Admin Article Services And APIs

- [x] 3.1 Add article service methods for admin article listing by status/search/category/tag and batch publish/reject/delete/takedown.
- [x] 3.2 Enforce status-transition rules and audit logging in service methods rather than handlers.
- [x] 3.3 Add admin handlers and routes under `/api/admin/articles` for articles, categories, tags, keywords, collection config, collection task logs, and batch actions.
- [x] 3.4 Add admin API tests for authentication, batch publish/reject/delete, takedown, custom taxonomy CRUD, keyword CRUD, and publish-without-AI confirmation support.

## 4. Collection Pipeline

- [x] 4.1 Define a collection provider interface that returns article candidates from a keyword without exposing provider-specific parsing to handlers.
- [x] 4.2 Implement the first Sogou WeChat search provider with conservative timeout, user-agent, result parsing, and error normalization.
- [x] 4.3 Implement manual collection trigger service that uses active keywords, dedupes by normalized original URL, stores candidates, and writes task/task-item logs.
- [x] 4.4 Implement global collection schedule configuration with enabled state, minute-level interval, and max result count.
- [x] 4.5 Wire scheduled collection through the backend scheduler or a dedicated command path without triggering collection from user article requests.
- [x] 4.6 Implement retry logic for failed tasks or failed task items while preserving URL dedupe.
- [x] 4.7 Add tests for provider result mapping, URL normalization/deduplication, task count updates, provider failure logs, and retry behavior.

## 5. Article AI Analysis

- [x] 5.1 Add independent article AI provider and prompt configuration storage separate from the active Bazi report provider.
- [x] 5.2 Reuse existing low-level AI client/encryption utilities where appropriate while keeping article activation and prompt scope independent.
- [x] 5.3 Implement manual AI analysis generation endpoint that builds prompts from stored body content when present, plus title, source, source publish time, search snippet/public summary, original URL, and category/tag hints.
- [x] 5.4 Persist structured AI output for one-sentence summary, key points, target readers, related topics, suggested tags, title pattern, opening style, structure outline, expression style, and rewrite angles.
- [x] 5.5 Implement AI analysis failure recording and single/batch retry paths.
- [x] 5.6 Add tests that AI prompt input includes only authorized body text and uses article-specific provider/prompt configuration.

## 6. User Article APIs

- [x] 6.1 Add authenticated user endpoints for article list, article detail, and original-link click tracking.
- [x] 6.2 Implement list filtering by category, tag, keyword query, latest sorting, hot sorting by view count, and pagination.
- [x] 6.3 Increment view count only for successful published article detail reads.
- [x] 6.4 Record original-link clicks with user ID, article ID, and timestamp only for published articles.
- [x] 6.5 Add handler/service tests for auth enforcement, published-only visibility, filters, sorting, view counts, and click tracking.

## 7. Admin Frontend

- [x] 7.1 Add article admin API methods to `frontend/src/lib/adminApi.ts`.
- [x] 7.2 Add an "资讯管理" admin sidebar entry and a page with tabs for articles, categories, tags, keywords, and task logs.
- [x] 7.3 Build article review table with status filters, search, category/tag filters, selection state, and batch publish/reject/delete/takedown controls.
- [x] 7.4 Add publish-without-AI confirmation when selected articles lack AI analysis.
- [x] 7.5 Build custom category and tag management tabs with create/edit/enable/disable behavior.
- [x] 7.6 Build keyword management and collection configuration UI with manual trigger and global schedule/max-count settings.
- [x] 7.7 Build task log UI showing status, counts, timestamps, errors, collected item details, and retry controls.
- [x] 7.8 Build article AI config and manual AI generation/retry controls inside article review flow.
- [x] 7.9 Add frontend tests for admin tab wiring, batch action payloads, publish confirmation, and task retry controls.

## 8. User Frontend

- [x] 8.1 Add article API methods to `frontend/src/lib/api.ts`.
- [x] 8.2 Add authenticated `/articles` and `/articles/:id` routes and redirect unauthenticated users through the existing auth pattern.
- [x] 8.3 Add "资讯" to top navigation.
- [x] 8.4 Build article list page with category filters, tag filters, keyword search, latest/hot sorting, pagination/loading, and empty states.
- [x] 8.5 Build article detail page rendering metadata, summary, reading support, writing breakdown, tags, original-link action, and AI-generated-content notice.
- [x] 8.6 Ensure article detail renders collected body content when present and hides empty AI sections when analysis is missing.
- [x] 8.7 Track original-link clicks before opening the WeChat original URL.
- [x] 8.8 Add frontend tests for auth gating, list filters/sort UI, AI notice, missing AI analysis rendering, and original-link click tracking.

## 9. Verification And Documentation

- [x] 9.1 Run backend tests for article repositories, services, handlers, collection, and AI analysis.
- [x] 9.2 Run frontend lint/build/tests for article admin and user pages.
- [x] 9.3 Run migration dry-run/apply verification in local development database.
- [ ] 9.4 Manually verify the full flow: configure keyword, trigger collection, review candidates, generate AI analysis, publish, browse as logged-in user, click original link, take down article.
- [x] 9.5 Document operational guidance for Sogou failures, rate limiting, retry behavior, content takedown, and the v1 no-full-text policy.
