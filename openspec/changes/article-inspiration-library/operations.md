## Article Inspiration Library Operations

### Collection Scope

V1 collects public search-result metadata from Sogou WeChat search and best-effort article body text when the original page can be resolved:

- title
- source account name
- original WeChat URL
- cover URL when available
- source publish time when available
- public search snippet
- article body text when retrievable

Article body fetching is best-effort. Sogou/WeChat can redirect to antispider pages or expire redirect tokens; body fetch failure must not block candidate creation. Store successful body text in `articles.body_content` with `full_text_authorized=true`.

### Sogou Failures And Rate Limiting

Sogou WeChat search can change markup, throttle, or block requests. Treat failures as operationally normal.

- Manual and scheduled runs write `article_collection_tasks` and `article_collection_task_items`.
- Provider failures are recorded as failed task items with the keyword and error message.
- Keep `article_collection_config.max_results_per_run` conservative. Start with 20.
- Disable scheduled collection from the admin config if Sogou blocks repeatedly.
- The provider boundary is replaceable; future paid/search providers should implement `ArticleSearchProvider`.

### Retry Behavior

Admin task retry re-runs failed keywords from a previous collection task.

- URL dedupe is still enforced by `articles.canonical_url_hash`.
- Duplicate URLs become duplicate task items and do not create new article rows.
- Retry does not auto-publish and does not auto-generate AI analysis.

### Review And Publishing

Collected articles enter `candidate` state.

- Users only see `published` articles.
- Publishing without AI analysis requires explicit admin confirmation in the UI.
- Reject, take down, and delete actions write audit events.
- Takedown is the V1 path for source-owner or user feedback.

### Article AI Analysis

Article AI uses article-specific provider and prompt configuration.

- It does not implicitly use the active Bazi report provider.
- Prompts are built from stored body text when present, plus metadata, snippets/summaries, original URL, and taxonomy hints.
- AI output is stored as structured `articles.ai_analysis`.
- Users see an explicit AI-generated reference notice on article detail pages.
