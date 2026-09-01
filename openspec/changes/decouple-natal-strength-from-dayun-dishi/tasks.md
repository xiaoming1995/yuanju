## 1. Natal-strength and Dayun road evaluation

- [x] 1.1 Add a backend regression fixture for `丁卯·壬子·壬辰·癸卯` that verifies the complete natal Fuyi assessment remains strong or very strong independently of every Dayun item.
- [x] 1.2 Replace the generic 十二长生 road delta with a branch-polarity-aware intensity evaluator using the natal Fuyi favorable/adverse elements.
- [x] 1.3 Include branch, Fuyi polarity, 十二长生 stage, and same-direction intensity in the back-five evidence detail while preserving aggregate road compatibility.
- [x] 1.4 Add backend cases for favorable/adverse/neutral branches at vigorous, medium, and weak stages, including the `戊申` case where 戊土 is favorable and 申金 is adverse.

## 2. Dayun overview context and copy

- [x] 2.1 Pass `natal_assessment.fuyi.day_master_strength` from `ResultPage` to `DayunTimeline` and then to the Dayun overview builder.
- [x] 2.2 Remove the frontend five-element percentage strength heuristic and strength-dependent Ten-God prose variants.
- [x] 2.3 Render concise stem and branch Fuyi statements, with branch-stage wording that amplifies or suppresses the branch role without changing natal strength.
- [x] 2.4 Provide neutral non-assertive fallback wording for legacy responses without natal Fuyi strength.

## 3. Verification

- [x] 3.1 Update Dayun overview unit/static tests so a natal-strong `戊申` case never emits `身弱遭杀克身` and identifies the adverse 申金长生 effect.
- [x] 3.2 Run targeted Go tests for Bazi assessment and Dayun road scoring, frontend Dayun tests, and `npm run build`.
- [x] 3.3 Validate the OpenSpec change with strict validation and document the manual verification path for switching Dayun cards.

### Manual verification

1. Start the backend and frontend, then calculate the male chart for 1987-12-09 at 06:00.
2. Select the `戊申` Dayun in the Dayun timeline and open “查看专业表述”.
3. Confirm the summary states the fixed natal-strong context, identifies 戊土 as favorable and 申金 as adverse, and says `申长生，忌神力量较显`.
4. Switch to other Dayun cards and confirm the natal-strength context does not change; only the stem/branch and stage effects change.
