## Context

Article collection currently has two separate stages:

```
Sogou result
    |
    v
Quality score
    |
    +--> low score and filtering enabled -> skipped task item
    |
    +--> accepted -> insert article as candidate
                         |
                         v
                    admin review
                         |
                         v
                     published
```

The new behavior adds a conservative auto-publish gate after a candidate is inserted:

```
Sogou result
    |
    v
Quality score
    |
    v
Insert candidate
    |
    +--> auto-publish gates pass -> published + audit event
    |
    +--> gates fail                -> remains candidate
```

## Goals / Non-Goals

**Goals:**

- Let scheduled collection automatically publish clearly usable high-score articles.
- Keep the threshold separate from the quality filter threshold.
- Require readable body content in V1.
- Limit the number of automatic publications per collection run.
- Make automatic publishing visible and auditable.
- Preserve manual review for anything uncertain.

**Non-Goals:**

- Do not auto-publish by default.
- Do not auto-generate AI analysis as part of V1 auto-publish.
- Do not publish existing backlog candidates automatically.
- Do not rely on unavailable WeChat-side metrics.
- Do not bypass normal article status lifecycle rules.

## Configuration Model

Suggested additive fields on `article_collection_config`:

- `auto_publish_enabled BOOLEAN NOT NULL DEFAULT false`
- `auto_publish_min_quality_score INTEGER NOT NULL DEFAULT 90`
- `auto_publish_requires_body BOOLEAN NOT NULL DEFAULT true`
- `auto_publish_max_per_run INTEGER NOT NULL DEFAULT 3`

Validation:

- `auto_publish_min_quality_score` is clamped to 0-100.
- `auto_publish_max_per_run` is clamped to a small safe range, for example 0-20.
- If auto-publish is disabled, the other fields are stored but ignored.

Why collection config instead of quality config:

- Quality config decides whether a candidate should enter the review pool.
- Auto-publish config decides whether a stored candidate can become user-visible.
- The two thresholds have different operational risk and should be tuned independently.

## Eligibility Rules

An article is eligible for V1 auto-publish only when:

- The collection task is a scheduled/manual collection run using the article collection service.
- The article was newly inserted as `candidate` in the current run.
- `auto_publish_enabled` is true.
- `quality_score >= auto_publish_min_quality_score`.
- `full_text_authorized` is true.
- `body_content` is not blank.
- The per-run auto-publish count is less than `auto_publish_max_per_run`.
- The normal status transition from candidate to published is valid.

AI analysis is not required in V1. This deliberately mirrors the existing manual "publish without AI after confirmation" behavior by making the automatic rule explicit in configuration.

## Audit Trail

Automatic publishing should write an audit event with system semantics:

```json
{
  "admin_id": null,
  "action": "auto_publish",
  "from_status": "candidate",
  "to_status": "published",
  "note": "定时采集自动发布：质量分 92 >= 阈值 90，正文已获取"
}
```

If keeping the existing action enum simpler is preferred, `action = "publish"` may be retained and the note must include automatic publication context. The clearer option is `auto_publish` because admin reporting can distinguish manual and automatic actions without parsing notes.

## Task Visibility

Collection task item rows should make the outcome inspectable. Two viable implementation options:

| Option | Shape | Trade-off |
|--------|-------|-----------|
| Reuse task item `error_msg`/metadata | Inserted item remains `inserted`, message mentions auto-published | Minimal schema but less structured |
| Add structured fields | `auto_published BOOLEAN`, optional `auto_publish_reason TEXT` | Cleaner UI and tests, slightly more migration |

Prefer structured fields if this feature is implemented, because operators will want to filter or scan auto-published items.

## Admin UI

Add an "自动发布" group near scheduled collection settings:

- Auto-publish switch.
- Minimum quality score input.
- Require body switch, initially locked or defaulted on.
- Max auto-publish per run input.
- Saved summary text such as: `自动发布关闭` or `质量分 >= 90 且有正文，每次最多 3 篇`.

Task logs should identify auto-published items and show the quality score/reason summary already available from the quality scoring feature.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Low-quality or off-brand content becomes public | Default disabled, high threshold, body required, per-run cap. |
| Articles without AI analysis feel less complete | V1 accepts this to avoid token-cost coupling; UI already supports reading body and original link. |
| Audit trail loses accountability | Use `admin_id = NULL` with `auto_publish` action and explicit note. |
| Duplicate articles get republished unexpectedly | Only newly inserted candidate rows are considered. Existing duplicates remain unchanged. |
| Operators confuse filter threshold and publish threshold | Separate UI groups and labels: "入库筛选" vs "自动发布". |
