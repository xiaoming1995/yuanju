## Context

Yuanju currently has clear backend layering (`handler` / `service` / `repository`) and admin/user JWT separation. AI-related features already support provider configuration and prompt management, but article collection introduces a new content domain with external search, operational review, legal/content boundaries, and user-facing reading flows.

The article module uses Sogou WeChat search results as candidate discovery, stores article metadata plus public snippets and platform/AI-generated analysis, and links users to the original WeChat article. When the original article page can be resolved, collection also stores cleaned article body text for reading and AI analysis. Body fetching is best-effort because Sogou/WeChat redirects and anti-spider pages are unstable.

High-level flow:

```
Admin keywords
      |
      v
Collection task / scheduler
      |
      v
Sogou WeChat search results
      |
      v
Candidate articles (dedupe by original URL)
      |
      v
Admin review
      |            \
      |             -> Manual AI analysis generation / retry
      v
Published article metadata + summary + breakdown
      |
      v
Logged-in frontend article list/detail
      |
      v
Original WeChat link click tracking
```

## Goals / Non-Goals

**Goals:**

- Provide a logged-in article reading/inspiration section with custom category/tag filtering, keyword search, latest/hot sorting, and article detail pages.
- Let operators collect WeChat article candidates by keyword through a background/manual collection path, not through ordinary user requests.
- Require admin approval before any collected article is visible to users.
- Keep article lifecycle auditable across candidate, published, rejected, taken down, and deleted states.
- Support custom categories and tags managed by operators.
- Support article AI analysis that is manually generated during review using independent article AI provider and prompt settings.
- Store metadata, public search snippets/summaries, best-effort body text, AI analysis, and links in V1.
- Track detail views for hot sorting and user-level outbound original-link clicks for analytics.

**Non-Goals:**

- No guarantee that every Sogou result can produce body text; collection must tolerate body fetch failures.
- No frontend report/complaint entry in V1; source/user feedback is handled by admin takedown.
- No user-facing favorites interaction in V1, although schema can reserve extension points.
- No personalized article recommendation from the user's chart in V1.
- No auto-publishing from collection or AI confidence.
- No guarantee that Sogou WeChat search is stable; the implementation must tolerate failure.

## Decisions

### Decision 1: Split the article domain from Bazi report/report-history domains

**Choice:** Add a dedicated article domain (`model`, `repository`, `service`, `handler`) and dedicated migrations instead of attaching content to existing Bazi chart/report tables.

**Rationale:** Articles are source/content/operation objects, not user chart analysis objects. Their lifecycle, audit, search, and external-source behavior differ from Bazi calculation and AI reports.

**Alternative:** Store article summaries as generic AI reports or static content records. Rejected because it would blur authorization, review state, and collection task semantics.

### Decision 2: Use a finite status lifecycle with audit records

**Choice:** Store an article `status` such as `candidate`, `published`, `rejected`, `taken_down`, `deleted`, plus an `article_audit_events` table for every publish/reject/takedown/delete action with admin ID and timestamp.

**Rationale:** Operators need batch actions and takedown accountability. Soft lifecycle states make review queues, published lists, and deleted-history safeguards straightforward.

**Alternative:** Hard-delete rejected/deleted articles immediately. Rejected because it weakens dedupe, audit, and feedback handling.

### Decision 3: Store metadata plus best-effort body content

**Choice:** Persist title, source name, publish time if available, cover URL, original URL, search snippet/public summary, cleaned body text when retrievable, custom category/tags, AI analysis JSON, and counters.

**Rationale:** The product goal is reading/inspiration and writing reference. Full content materially improves review, user reading, and AI analysis quality. The original source link remains visible, and collection must still work when body fetching is blocked.

**Alternative:** Store metadata only. Rejected after validation because operators need full content for meaningful review and reference.

### Decision 4: Sogou WeChat collection is a replaceable background provider

**Choice:** Implement collection behind a provider interface in the service layer, with the first provider targeting Sogou WeChat search. Trigger it from an admin/manual action and a global scheduler/command path.

**Rationale:** Sogou search is closer to public WeChat articles than generic web search, but unstable. A provider boundary allows future replacement with a paid API, generic search API, or manual import without rewriting review/storage.

**Alternative:** Couple HTTP parsing directly inside admin handlers. Rejected because collection can be slow, fail often, and should not live in the request-response path.

### Decision 5: AI analysis is manual during review, not collection-time automatic

**Choice:** Collection stores candidates first. Admins generate AI summary/tags/writing breakdown manually during review, and can retry failed generation. Publishing without AI content is allowed after confirmation.

**Rationale:** Collection quality will be noisy when using keyword search. Manual AI generation controls token cost and avoids producing AI content for candidates operators will reject.

**Alternative:** Auto-generate analysis for every collected candidate. Rejected because it wastes tokens and hides bad-source quality behind generated prose.

### Decision 6: Article AI provider and prompt configuration are independent

**Choice:** Add article-specific provider and prompt settings rather than reusing the currently active Bazi LLM provider/prompt.

**Rationale:** Article analysis has a different cost/latency/quality profile from Bazi interpretation. Operators may want a cheaper model for article metadata analysis or a stricter prompt without affecting report generation.

**Alternative:** Reuse the active LLM provider with only a separate prompt. Rejected because provider choice was explicitly confirmed as independent.

### Decision 7: Frontend display is authenticated and source-forward

**Choice:** Require normal user JWT for article list/detail. Display source, original URL action, AI notice, and analysis sections. Add "资讯" to top navigation only.

**Rationale:** The module is intended as a logged-in feature and can be iterated without adding mobile bottom navigation weight. Source-forward display keeps ownership visible and encourages users to read the original.

**Alternative:** Public list/detail for SEO. Rejected because the confirmed requirement is full login gating.

## Data Model Sketch

Core tables:

- `article_categories`: id, name, slug, sort_order, active, created_at, updated_at.
- `article_tags`: id, name, slug, active, created_at, updated_at.
- `article_keywords`: id, keyword, active, created_at, updated_at.
- `articles`: id, title, source_name, original_url, canonical_url_hash, cover_url, published_at_source, search_snippet, summary, ai_analysis JSONB, category_id, status, view_count, original_click_count, full_text_authorized, body_content, created_at, updated_at, published_at, taken_down_at, deleted_at.
- `article_tag_links`: article_id, tag_id.
- `article_collection_tasks`: id, trigger_type, status, started_at, finished_at, keyword_count, found_count, inserted_count, duplicate_count, failed_count, error_msg.
- `article_collection_task_items`: task_id, article_id nullable, original_url, keyword, status, error_msg.
- `article_audit_events`: id, article_id, admin_id, action, from_status, to_status, note, created_at.
- `article_original_clicks`: id, article_id, user_id, clicked_at.
- Optional/reserved `article_favorites`: user_id, article_id, created_at, not wired to UI in V1.
- `article_ai_providers` / `article_ai_prompts` or article-specific configuration fields/tables, kept separate from Bazi report provider activation.

`body_content` stores cleaned article text when retrieval succeeds. `full_text_authorized` is true when body text is present and eligible for rendering. List APIs should clear body text from list payloads; detail APIs can return body text.

## Status Model

```
candidate
  |-- publish --> published
  |-- reject  --> rejected
  |-- delete  --> deleted

published
  |-- take_down --> taken_down
  |-- delete    --> deleted

rejected
  |-- publish --> published
  |-- delete  --> deleted

taken_down
  |-- publish --> published
  |-- delete  --> deleted
```

Every transition writes an audit event. Frontend article queries only return `published` rows.

## API Shape

User-facing:

- `GET /api/articles?category=&tag=&q=&sort=latest|hot&page=&limit=`
- `GET /api/articles/:id`
- `POST /api/articles/:id/original-click`

Admin:

- `GET /api/admin/articles?status=&q=&category=&tag=&page=&limit=`
- `POST /api/admin/articles/batch-action` for publish/reject/delete/takedown.
- `POST /api/admin/articles/:id/ai-analysis`
- `POST /api/admin/articles/:id/ai-analysis/retry`
- Category/tag/keyword CRUD under `/api/admin/articles/categories`, `/tags`, `/keywords`.
- `POST /api/admin/articles/collect` for manual trigger.
- `GET /api/admin/articles/collection-tasks`
- `POST /api/admin/articles/collection-tasks/:id/retry`
- `GET/PUT /api/admin/articles/collection-config`
- `GET/PUT /api/admin/articles/ai-config`

Exact route names can be adjusted during implementation, but they should preserve these capability boundaries.

## Collection Behavior

- Manual trigger starts a task using all active keywords or selected keywords if the UI later supports selection.
- Scheduled collection uses global configuration: enabled flag, minute-level interval, and max results per run.
- Collection dedupes by normalized original URL hash. If the same URL is found again, it increments duplicate counts and does not create a second article row.
- Provider errors and parse errors are recorded at task and item level.
- Retrying a failed task re-runs failed keyword searches or failed item processing without duplicating existing URLs.

## AI Analysis Shape

AI input must include only:

- title
- source name
- source publish time if present
- search snippet/public summary
- original URL
- current category/tag hints if present

AI output should be stored as structured JSON, for example:

```json
{
  "one_sentence_summary": "",
  "key_points": [],
  "target_readers": [],
  "related_topics": [],
  "suggested_tags": [],
  "title_pattern": "",
  "opening_style": "",
  "structure_outline": [],
  "expression_style": [],
  "rewrite_angles": []
}
```

The admin UI can copy accepted `suggested_tags` into custom tags, but AI-suggested tags should not silently create taxonomy entries without admin confirmation.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Sogou WeChat search changes layout or blocks requests | Keep collection behind a provider interface, add rate limits, retries, task logs, and admin-visible failure state. |
| AI analysis based only on snippets is shallow or inaccurate | Use stored body content when retrieval succeeds, display explicit AI notice, keep generated content editable/regeneratable later, and allow publishing without AI sections. |
| Keyword search creates noisy candidates | Use candidate-only ingestion, batch reject/delete, and custom category/tag review before publishing. |
| Source content complaints | Show source and original link, support admin takedown with audit, and keep body fetch failures non-blocking. |
| Independent AI config duplicates existing provider logic | Reuse low-level client/encryption utilities where possible, but keep activation/config tables separate for article behavior. |
| User-level click tracking adds privacy sensitivity | Store only user ID, article ID, timestamp; do not store external browsing contents beyond the clicked original article. |
| Status/audit schema adds implementation cost | The cost is justified by takedown, compliance, and operations needs. |

## Migration Plan

1. Add article tables and indexes through additive migrations.
2. Implement repositories/services/handlers with no seed data and no frontend exposure until migrations pass.
3. Add admin taxonomy/keyword/config management and article review list.
4. Add collection command/scheduler pathway with Sogou provider and task logs.
5. Add article AI configuration and manual generation endpoints.
6. Add logged-in frontend article list/detail and top navigation entry.
7. Verify list endpoints do not return large body payloads, while detail endpoints return collected body content when present.

Rollback is additive: disable the frontend navigation and scheduler/config flag first, then leave tables in place. No existing Bazi/report data is modified.

## Open Questions

- Exact Sogou request strategy, user-agent, and rate limit values should be decided during implementation after a small spike.
- Whether article AI configuration should reuse existing `llm_providers` table with an article-specific activation scope or use a fully separate table needs implementation review against current provider repository patterns.
- Whether the reserved favorites table should be created in V1 or deferred until the feature is implemented should be decided during migration design.
