## Why

Yuanju needs a logged-in article inspiration module where users can browse Bazi-related WeChat article references for reading and writing inspiration, while operators can collect, review, classify, and publish candidate articles with best-effort article body capture.

The module should balance product value with external-source instability: first version uses Sogou WeChat search metadata and public snippets, attempts to resolve and store cleaned article body text during collection, routes users to the original WeChat article, and tolerates body-fetch failures caused by redirect expiry or anti-spider pages.

## What Changes

- Add a logged-in frontend "资讯" section with article list and detail pages.
- Require login for all article list/detail access.
- Support frontend filtering by custom category, custom tag, keyword search, and latest/hot ordering.
- Define hot ordering by article detail view count in the first version.
- Show source metadata, public summary, AI reading support, AI writing breakdown, original article link, and an explicit AI-generated-content notice.
- Add user-level original-link click tracking for WeChat article outbound clicks.
- Add an admin "资讯管理" menu with tabs for articles, categories, tags, keywords, and collection task logs.
- Support fully custom categories and tags managed by operators.
- Support a simple keyword list for Sogou WeChat search collection.
- Add manual collection trigger and globally configurable scheduled collection interval in minutes/max count.
- Store collected items as candidates first; do not auto-publish.
- Support batch publish/reject/delete from the candidate pool.
- Support full status lifecycle and audit records: candidate, published, rejected, taken down, deleted.
- Support manual takedown from admin for source-owner/user feedback; no frontend report entry in v1.
- Store article metadata, public search snippet/summary, best-effort body content, tags, category assignment, AI analysis, and links.
- Add independent article AI provider and prompt configuration separate from Bazi report AI settings.
- Generate AI summary/breakdown manually during admin review, with retry support for failed article AI analysis.
- Allow admin to publish without AI analysis after confirmation.
- Render stored body content on article detail when collection successfully retrieves it.

## Capabilities

### New Capabilities

- `article-content-library`: Logged-in frontend article browsing, detail display, filters/sorting, view counting, outbound click tracking, and AI notice behavior.
- `article-curation-workflow`: Admin article review, status lifecycle, audit trail, custom categories/tags, keyword management, and batch operations.
- `article-collection-pipeline`: Independent collection task execution for Sogou WeChat search, manual/scheduled triggering, deduplication, limits, task logs, and retry handling.
- `article-ai-analysis`: Independent article AI provider/prompt configuration and manual generation of summary, reading support, tags, and writing breakdown from stored body content when present, otherwise search metadata.

### Modified Capabilities

- None.

## Impact

- **Backend**:
  - New article domain models, repositories, services, handlers, and migrations.
  - New collection task command or scheduler pathway that writes candidates to PostgreSQL outside normal user request flow.
  - New admin routes under `/api/admin/articles...`.
  - New authenticated frontend routes under `/api/articles...`.
  - New independent article AI provider/prompt configuration and AI request logging.
- **Frontend**:
  - Add `/articles` and `/articles/:id` user pages behind normal user auth.
  - Add "资讯" to top navigation.
  - Add admin "资讯管理" page with tabs for articles, categories, tags, keywords, and task logs.
  - Extend API clients for article/admin article endpoints.
- **Database**:
  - Add tables for articles, categories, tags, article-tag mapping, keywords, collection tasks/logs, article audit events, outbound click logs, and user favorites/full-text fields.
- **External dependency risk**:
  - Sogou WeChat search is an unstable external source with anti-scraping and layout-change risk; implementation must be rate-limited, logged, retryable, and replaceable.
- **Content retrieval risk**:
  - Full body retrieval is best-effort. Sogou/WeChat redirects can be blocked or expired; candidates should still be created with metadata and snippets.
