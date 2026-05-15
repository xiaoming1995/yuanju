## Overview

This change turns history from a plain list into an archive experience. The user should understand what they have saved, switch between chart and compatibility records, and continue into the right result page with minimal scanning.

## Route Strategy

Keep existing routes:

- `/history` remains the primary chart archive.
- `/compatibility/history` remains the compatibility archive.
- `/profile` continues to link into both.

The routes are not merged because no backend endpoint currently returns a unified feed, and merging data client-side would add complexity without improving the first iteration.

## Chart Archive Layout

The chart archive page will include:

- Hero card with title, record count, and short retention-oriented copy.
- Archive switcher with links to chart archive and compatibility archive.
- Summary stat cards derived from existing chart records.
- Empty state with a direct action to create a chart.
- Record cards showing pillars, birth date, gender, created date, and a clear "查看命盘" cue.

## Compatibility Archive Layout

The compatibility archive page will mirror the archive structure:

- Hero card with total count and link back to chart archive.
- Archive switcher with active compatibility state.
- Cards showing participant names, level, tags, scores, and view cue.
- Empty state with direct action to create a compatibility reading.

## Mobile Rules

- Both archive pages reserve bottom space for the mobile bottom navigation.
- Cards must be single-column on phone widths.
- Actions and switcher links must be large enough to tap.
- Avoid inline styles for reusable archive UI so future archive pages can stay consistent.

## Testing

- Add static tests that check for archive switchers, stats, card metadata, dedicated compatibility CSS, and mobile bottom safe-area rules.
- Re-run lint, build, and frontend node tests.
