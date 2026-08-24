## Why

The article inspiration module now collects Bazi-related WeChat article candidates by keyword, but raw keyword collection is noisy. Operators need a way to keep higher-value inspiration references and avoid filling the review queue with low-quality, irrelevant, or unusable articles.

The originally discussed idea of filtering by WeChat likes or favorites is not reliable for V1 because those interaction metrics are not stable public fields in Sogou WeChat search result HTML, and WeChat-side collection/favorite counts are not generally available to ordinary article pages. A more stable approach is to score candidates using available collection signals and later combine them with Yuanju's own user behavior signals.

## What Changes

- Add an article collection quality scoring model for collected candidates.
- Persist a numeric quality score and explainable quality signals/reasons for each article candidate.
- Add configurable quality filtering in admin collection configuration.
- Allow low-score candidates to be skipped before article insertion when quality filtering is enabled.
- Record skipped candidates in collection task item logs with score and reason.
- Add source quality controls such as source blacklist and optional source allowlist/preferred list.
- Use available public collection signals: search ranking, publish recency, source name, title/snippet keyword relevance, body fetch success, and optional AI suitability analysis.
- Use Yuanju-owned behavior signals after publication: article view count, original-link clicks, and future in-site favorites.
- Treat external WeChat metrics such as likes/favorites/reads as optional enhancement fields only; missing external metrics must not block collection.

## Capabilities

### New Capabilities

- `article-collection-quality-scoring`: Score, filter, and explain collected article candidates before review, with admin configuration and task-log visibility.

### Modified Capabilities

- `article-collection-pipeline`: Collection should consult the quality scoring stage before inserting candidates when quality filtering is enabled.
- `article-curation-workflow`: Admin review should expose quality score/reasons and allow sorting/filtering by quality.

## Impact

- **Backend**:
  - Add quality score fields to articles and task items, or equivalent detail storage.
  - Add quality scoring service that runs after provider parsing and before candidate insertion.
  - Extend collection config with filtering controls and scoring weights.
  - Extend collection task logs with skipped-by-quality entries.
- **Frontend**:
  - Add quality filter settings in admin collection/content configuration.
  - Show quality score, score reasons, and skipped reasons in task logs.
  - Allow admin article list sorting/filtering by quality score.
- **Database**:
  - Add additive columns/tables for article quality score, quality reasons, source rules, and collection quality config.
- **External dependency boundary**:
  - Do not require WeChat likes, favorites, reads, or other non-public external metrics. These fields may be reserved as nullable optional metrics only.

