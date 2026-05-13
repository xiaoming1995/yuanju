## 1. Regression Coverage

- [x] 1.1 Add backend unit tests in `backend/pkg/bazi/event_narrative_test.go` that reproduce adjacent child-age years sharing `综合变动` but differing in school, relationship, and health signals
- [x] 1.2 Assert adjacent narratives do not share the same opening sentence when their specific signal focus differs
- [x] 1.3 Assert child-age narratives prioritize learning, classmates, family communication, routine, or health wording over generic adult-style change wording
- [x] 1.4 Assert `RenderEvidenceSummary` still preserves technical evidence for the same signal sets

## 2. Narrative Selection Refinement

- [x] 2.1 Add a helper that distinguishes strong change signals from ordinary broad `综合变动`
- [x] 2.2 Update dominant theme selection so ordinary `综合变动` does not outrank specific school, relationship, health, money, movement, or career themes
- [x] 2.3 Add child-age priority ordering for `age < 18`
- [x] 2.4 Keep strong change signals such as `伏吟`, `反吟`, `局势_重`, `大运合化`, "大运流年双重命中", "力度倍增", and "重大事件高发" eligible as dominant themes

## 3. Evidence-Sensitive Plain Wording

- [x] 3.1 Add change wording variants for dayun-liunian collision, month-pillar interaction, day-branch/self-palace interaction, kongwang, fuyin, fanyin, and yima-like movement
- [x] 3.2 Add child-age wording variants for school pressure, school help, peer relationship, family communication, emotional fluctuation, resource/family support, and health/routine
- [x] 3.3 Ensure default `narrative` does not expose raw technical terms such as `流年地支`, `月柱`, `官杀`, `伏吟`, `空亡`, or `财星`

## 4. Verification

- [x] 4.1 Run `go test ./pkg/bazi -run 'TestRenderYearNarrative|TestRenderEvidenceSummary'`
- [x] 4.2 Run `go test ./pkg/bazi`
- [x] 4.3 Manually inspect generated narrative examples for the 1996-02-08 20:00 case if a local chart fixture or API path is available
- [x] 4.4 Confirm no frontend API shape changes are required
