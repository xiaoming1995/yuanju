## Context

The existing article collection pipeline discovers candidates from Sogou WeChat search and stores metadata, snippets, best-effort body content, and task logs. The next operational problem is not raw collection volume, but candidate quality.

The quality scoring stage should sit between provider results and article insertion:

```
Sogou result
    |
    v
Normalize URL / metadata
    |
    v
Quality scorer
    |
    +--> score >= threshold -> insert candidate article
    |
    +--> score < threshold  -> record skipped task item
```

The score must be explainable. Operators should understand why an item was kept or skipped without reading code.

## Goals / Non-Goals

**Goals:**

- Reduce low-quality candidates entering the review queue.
- Provide transparent scoring reasons in task logs and article review.
- Keep scoring deterministic and cheap by default.
- Allow operators to tune thresholds and source/keyword rules.
- Use station-owned behavior signals for post-publication ranking.
- Reserve room for optional external metrics without depending on them.

**Non-Goals:**

- Do not require WeChat like/favorite/read counts.
- Do not bypass WeChat/Sogou anti-scraping to obtain private or unstable metrics.
- Do not auto-publish articles based on score.
- Do not use AI scoring as the only gate in V1.

## Quality Score Model

Use a 0-100 score. The first version should be rule-based and explainable.

Suggested default weights:

| Signal | Suggested Weight | Notes |
|--------|------------------|-------|
| Search rank | 0-20 | Higher-ranked Sogou results get more points. |
| Publish recency | 0-15 | Recent articles get more points; unknown time gets neutral/low points. |
| Body fetch success | 0-20 | Full body improves review and AI analysis quality. |
| Title/snippet relevance | 0-20 | Matches configured bonus keywords and collection keyword. |
| Source quality | -40 to +15 | Blacklist can hard-filter; preferred sources can boost. |
| Content usability | 0-10 | Avoid empty/video/anti-spider pages; usable text gets points. |
| Optional AI suitability | 0-20 | Optional, cost-controlled, not required for V1 default. |

The score should be stored with structured reasons, for example:

```json
{
  "score": 76,
  "reasons": [
    {"type": "rank", "points": 18, "message": "搜狗结果排名靠前"},
    {"type": "body", "points": 20, "message": "正文已获取"},
    {"type": "keyword", "points": 12, "message": "标题命中：财运、生肖"},
    {"type": "source", "points": 8, "message": "来源在优先名单"}
  ]
}
```

## Configuration

Admin configuration should include:

- `quality_filter_enabled`: whether to skip low-score candidates before insertion.
- `min_quality_score`: minimum score for insertion when filtering is enabled.
- `allow_without_body`: whether body-fetch failure can still pass.
- `bonus_keywords`: words that improve relevance score.
- `source_blacklist`: sources that should be skipped or heavily penalized.
- `preferred_sources`: sources that receive a small boost.
- `ai_quality_check_enabled`: optional AI suitability check.

Default behavior should be conservative:

- Filtering disabled or threshold modest on first rollout.
- External WeChat metrics ignored unless available through a legitimate stable source later.
- Skipped items still recorded in task logs.

## Data Model Sketch

Additive fields/tables can be implemented in either direct columns or a small config table.

Suggested article fields:

- `quality_score INTEGER NOT NULL DEFAULT 0`
- `quality_reasons JSONB`

Suggested task item fields:

- `quality_score INTEGER NOT NULL DEFAULT 0`
- `quality_reasons JSONB`
- `skip_reason TEXT`

Suggested config:

- `article_quality_config`: enabled, min score, source rules, keywords, AI flag, updated_at.

Optional external metrics fields may be nullable:

- `external_like_count INTEGER NULL`
- `external_favorite_count INTEGER NULL`
- `external_read_count INTEGER NULL`
- `external_metrics_status TEXT`
- `external_metrics_error TEXT`

These optional fields must not be used as mandatory filters in V1.

## Collection Behavior

- Collection still queries active keywords and dedupes by normalized URL.
- Each provider result receives a quality score before insertion.
- If filtering is disabled, all normal candidates can be inserted, but score/reasons are stored.
- If filtering is enabled and score is below threshold, the item is recorded as `skipped` with quality details and is not inserted into `articles`.
- If a source is blacklisted, the item may be skipped regardless of score.
- If the article is a duplicate, score can be recorded on the task item, but existing article rows should not be duplicated.

## Admin Experience

Admin should see:

- In collection config: quality filtering switch, threshold, source rules, bonus keywords.
- In task logs: each item score, kept/skipped status, and reasons.
- In article review list: quality score badge and sort/filter controls.
- In article detail: quality score breakdown near collection metadata.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Rule weights feel arbitrary | Start with conservative defaults and expose configuration. |
| Good niche articles score low | Keep filtering optional and show skipped logs for audit. |
| AI quality scoring increases cost | Make AI scoring optional and disabled by default. |
| Source blacklist mistakes remove useful content | Log skipped items and allow operators to adjust rules. |
| External metrics are unavailable | Treat WeChat metrics as nullable optional enhancement only. |

