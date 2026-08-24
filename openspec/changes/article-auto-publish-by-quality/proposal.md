## Why

The article collection pipeline can now score candidates and filter low-quality items, but high-quality items still require manual review and publication. Operators want scheduled collection to publish clearly usable, high-quality articles automatically when they pass a conservative quality threshold and have readable body content.

This should reduce repetitive review work without turning keyword collection into an unsafe fully automatic publishing system. The first version should keep automatic publishing explicit, disabled by default, limited per run, and auditable.

## What Changes

- Add auto-publish controls to article scheduled collection configuration.
- Allow scheduled collection to publish newly inserted candidate articles when they meet all configured auto-publish gates:
  - auto-publish is enabled
  - quality score is at or above the auto-publish threshold
  - readable body content is available
  - per-run auto-publish limit has not been reached
- Keep quality filtering and auto-publishing as separate thresholds.
- Record automatic publication in article audit events with a clear automatic action/note.
- Surface auto-publish configuration and auto-published task outcomes in admin article management.
- Preserve manual review for candidates that do not meet auto-publish criteria.

## Capabilities

### New Capabilities

- `article-auto-publish-by-quality`: Automatically publish high-quality collected article candidates under conservative scheduled-collection rules.

### Modified Capabilities

- `article-collection-pipeline`: Scheduled collection may promote eligible newly inserted candidates to published after collection scoring.
- `article-curation-workflow`: Admin review and audit views should distinguish manually published articles from automatically published articles.

## Impact

- **Backend**:
  - Extend article collection configuration with auto-publish enabled flag, minimum score, body requirement, and max-per-run limit.
  - Run auto-publish eligibility checks only after candidate insertion succeeds.
  - Add repository/service support for system-driven publish audit events.
  - Extend collection task or item visibility with auto-publish outcome where appropriate.
- **Frontend**:
  - Add auto-publish controls under admin article collection scheduling.
  - Show concise auto-publish status/counts in task logs or item rows.
- **Database**:
  - Additive migration for auto-publish config fields.
  - Optional additive task count/item note fields if needed for visibility.
- **Behavioral boundary**:
  - V1 does not automatically generate AI analysis before publishing.
  - V1 does not auto-publish candidates without readable body content.
  - V1 does not auto-publish existing backlog candidates unless a future explicit backfill tool is designed.
