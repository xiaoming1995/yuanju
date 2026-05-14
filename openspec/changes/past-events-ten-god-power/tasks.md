## 1. Backend Ten-God Profile Model

- [x] 1.1 Add a `TenGodPowerProfile` structure in the bazi package with dominant ten-god, group, group label, strength, polarity, plain title, plain text, score, and reason fields.
- [x] 1.2 Add ten-god grouping helpers for wealth, official, seal, output, and peer groups.
- [x] 1.3 Add plain-language copy helpers for each force group and polarity combination.
- [x] 1.4 Add unit tests for exact ten-god to force-group mapping.

## 2. Backend Scoring Rules

- [x] 2.1 Implement dayun ten-god power scoring using dayun heavenly stem, earthly branch main qi, yongshen/jishen fit, and natal strength context.
- [x] 2.2 Implement liunian ten-god power scoring using liunian heavenly stem, earthly branch main qi, containing dayun reinforcement, yongshen/jishen fit, and natal strength context.
- [x] 2.3 Apply dayun front-five/heavenly-stem and back-five/earthly-branch phase context to yearly force scoring.
- [x] 2.4 Add tests covering representative wealth, official, seal, output, and peer years.
- [x] 2.5 Add tests proving dayun/liunian same-group reinforcement increases score and reason detail.

## 3. Past-Events API Integration

- [x] 3.1 Extend `DayunMetaItem` with a dayun ten-god power profile.
- [x] 3.2 Extend `PastEventsYearItem` with a liunian ten-god power profile.
- [x] 3.3 Populate dayun profiles in `GeneratePastEventsYears`.
- [x] 3.4 Populate yearly profiles while converting `YearSignals` to API response items.
- [x] 3.5 Add service-level tests for the additive response shape where practical.

## 4. Narrative And Summary Integration

- [x] 4.1 Update yearly narrative rendering so ten-god force can enrich otherwise generic years.
- [x] 4.2 Preserve existing priority for hard evidence such as clashes, punishments, voids, and use-god/jishen position hits.
- [x] 4.3 Include dayun and liunian ten-god profiles in the dayun summary prompt payload for newly generated summaries.
- [x] 4.4 Add narrative tests proving hard event evidence is not overridden by ten-god force.
- [x] 4.5 Add narrative tests proving generic years gain more specific ten-god force wording.

## 5. Frontend Display

- [x] 5.1 Update TypeScript API types for dayun and yearly ten-god power profiles.
- [x] 5.2 Show a compact dayun dominant-force label in each dayun header.
- [x] 5.3 Show a short plain-language yearly force explanation on each year card.
- [x] 5.4 Keep score and detailed reason out of the default card layout.
- [x] 5.5 Verify mobile and desktop card layout does not overlap or become visually dense.

## 6. Verification

- [x] 6.1 Run `go test ./pkg/bazi`.
- [x] 6.2 Run relevant backend package tests for past-events service/handler paths.
- [x] 6.3 Run `npm run build` in `frontend`.
- [x] 6.4 Manually inspect the 1996-02-08 20:00 test chart and confirm dayun/year cards show readable ten-god force descriptions.
- [x] 6.5 Confirm no new per-year LLM calls are introduced.
