## 1. Regression Coverage

- [x] 1.1 Add tests proving a meaningful signal-bearing year produces a richer narrative than the current short reminder style.
- [x] 1.2 Add tests proving weak-signal years remain concise and do not receive artificial filler.
- [x] 1.3 Add tests proving hard evidence remains the leading meaning when combined with ten-god force.
- [x] 1.4 Add tests proving adjacent years with different concrete signals do not share the same main wording.
- [x] 1.5 Add tests covering study-age wording so child/teen years avoid adult career or romance defaults.

## 2. Narrative Composition

- [x] 2.1 Introduce a local narrative frame for yearly tone, trigger source, domain manifestation, ten-god explanation, and practical stance.
- [x] 2.2 Add a yearly tone helper that summarizes polarity mix and hard-signal intensity in plain language.
- [x] 2.3 Add trigger-source wording for major evidence types such as clash, punishment, void, fuyin, fanyin, formation, and yongshen/jishen hits.
- [x] 2.4 Add domain-specific detail builders for study, career, money, relationship, health, movement, support, and change.
- [x] 2.5 Add an age-aware wording layer for study-age years.
- [x] 2.6 Integrate ten-god power into the narrative only when it adds distinct context.
- [x] 2.7 Keep `RenderEvidenceSummary` as the technical-detail outlet and avoid moving raw evidence into default card text.

## 3. Repetition Control

- [x] 3.1 Ensure the selected dominant and secondary signal themes influence different parts of the narrative.
- [x] 3.2 Add wording variation based on signal source and evidence content, without random phrase selection.
- [x] 3.3 Prevent the same standalone ten-god sentence from becoming the main differentiator across adjacent years.
- [x] 3.4 Preserve previous anti-repetition behavior from `fix-past-events-repetitive-narrative`.

## 4. Frontend Readability

- [x] 4.1 Inspect `PastEventsPage` year-card spacing with longer narratives.
- [x] 4.2 Adjust line height, margins, or max width only if longer text becomes visually dense.
- [x] 4.3 Verify mobile and desktop layouts do not overlap badges, narrative text, yearly force row, or evidence toggle.

## 5. Verification

- [x] 5.1 Run `go test ./pkg/bazi`.
- [x] 5.2 Run relevant backend service/handler tests for the past-events endpoint.
- [x] 5.3 Run `npm run build` in `frontend`.
- [x] 5.4 Run `openspec validate enrich-past-events-year-narratives --strict`.
- [x] 5.5 Manually inspect the 1996-02-08 20:00 test chart and confirm year cards feel more detailed without returning to repetitive wording.
- [x] 5.6 Confirm no new per-year LLM calls or token usage are introduced.
