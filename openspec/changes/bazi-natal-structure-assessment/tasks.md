## 1. Original-Natal Assessment

- [x] 1.1 Add versioned natal assessment result types to the Bazi model.
- [x] 1.2 Implement climate urgency, Fuyi support, carrying, and evidence collection using existing chart data.
- [x] 1.3 Implement pattern-star support, configured formation/break rules, and conservative natal flow or branch-relation assessment.
- [x] 1.4 Convert the assessment score and hard limits into an explainable natal grade.

## 2. Vehicle and Dayun Integration

- [x] 2.1 Refactor vehicle profile generation to use the natal assessment rather than a fixed Ming Ge score.
- [x] 2.2 Make Dayun element and pattern effects consume the natal assessment and emit structure-interaction evidence.
- [x] 2.3 Generate the assessment during new calculations and lazily backfill it with dependent vehicle and road data for historical snapshots.
- [x] 2.4 Include the structured assessment in AI report context without changing legacy use-god fields.

## 3. Presentation and Verification

- [x] 3.1 Render the revised basis language and assessment evidence in the result page while retaining legacy response compatibility.
- [x] 3.2 Add focused backend tests for climate/Fuyi priority, pattern formation/break, grade ceilings, Dayun interaction, and snapshot backfill.
- [x] 3.3 Update frontend static tests and run backend tests, frontend checks, and a local calculation verification.
