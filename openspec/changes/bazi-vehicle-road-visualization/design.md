## Context

The current Bazi engine already produces the raw material needed for a higher-level interpretation layer:

- `BaziResult` includes Ming Ge, Yongshen/Jishen, Tiaohou, Ten God relation data, Shen Sha, hidden stems, Dayun, Jin Bu Huan ratings, and Liunian records.
- `GetStrengthDetail()` exposes a deterministic day-master strength level, score, and reasoning detail.
- Each `DayunItem` can include `JinBuHuanResult`, which rates the front five years by Dayun stem and the back five years by Dayun branch.
- The frontend already has `dayunOverview.ts`, `dayunTrend.ts`, and `DayunTimeline.tsx` logic that turns Dayun data into ordinary-user prose and relative trend charts.

The missing layer is an absolute, cross-Dayun product concept. The existing trend chart explicitly warns that intra-Dayun curve movement does not compare different Dayun levels. The "vehicle and road" layer fills that gap by deriving a natal vehicle profile and per-Dayun road map from deterministic backend signals.

Stakeholders:

- Ordinary users need an intuitive summary that can be understood without studying Ten Gods or classical rule tables.
- Professional users need evidence disclosure so the metaphor does not hide the basis of judgment.
- The application needs additive API changes that do not break old saved charts, report caches, or existing timeline interactions.

## Goals / Non-Goals

**Goals:**

- Compute a deterministic `vehicle_profile` for the natal chart.
- Compute a deterministic `dayun_roadmap` for each Dayun period.
- Reuse existing algorithm signals instead of adding a new astrology dependency or LLM-only scoring.
- Expose compact labels for ordinary users and structured evidence for professional users.
- Surface "命盘座驾" and "当前路况" near the top of the result page.
- Show road labels and front/back phase indicators in the Dayun timeline.
- Inject the deterministic vehicle/road context into AI report prompts as explanation context.

**Non-Goals:**

- Do not replace Ming Ge, Yongshen/Jishen, Tiaohou, Shen Sha, or Dayun timeline features.
- Do not claim deterministic life outcomes, wealth rank, social class, or moral value.
- Do not add paid-tier gating, subscriptions, or admin configuration in the first version.
- Do not add 3D animation or complex game-like vehicle motion in the first version.
- Do not migrate existing database schema in the first version.

## Decisions

### Decision 1: Backend owns scoring; AI owns language polish only

The backend will calculate grade, score, vehicle type, road type, phase labels, tags, and evidence deterministically. AI prompts will receive these values and explain them, but the LLM must not recalculate or override them.

Rationale:

- The product promise is "algorithm precision + AI natural-language interpretation"; ratings must belong to the algorithm side.
- Deterministic scoring is testable and consistent across providers.
- AI-generated ratings would drift between model versions and make support/debugging difficult.

Alternative considered: ask the AI to judge "vehicle grade" directly from full chart data. Rejected because it is harder to test, harder to explain, and can produce inconsistent or overly absolute wording.

### Decision 2: Keep natal vehicle and Dayun road as separate structures

The result will include:

- `vehicle_profile`: describes the natal chart's base configuration.
- `dayun_roadmap`: describes each Dayun period's road condition.

Rationale:

- Natal structure and Dayun environment answer different questions.
- A strong vehicle on bad roads and an ordinary vehicle on smooth roads are different stories; one combined fortune score would erase that distinction.
- This separation matches the user's metaphor: "八字比如车，大运决定高速还是泥路."

Alternative considered: a single `fortune_score`. Rejected because it encourages oversimplified "命好/命差" conclusions and makes evidence harder to audit.

### Decision 3: Use S/A/B/C/D as configuration completeness, not human value

Vehicle grades will be mapped from a 0-100 score:

- `S`: 90-100, 协同型配置，主线清晰，优势较易整合发挥。
- `A`: 75-89, 稳健型配置，基础扎实，顺运时通常更省力。
- `B`: 60-74, 实用型配置，能力稳定，适合顺势经营。
- `C`: 45-59, 特性型配置，优势与限制都明显，更看选择与路况。
- `D`: 0-44, 调校型配置，对环境与策略更敏感，需要更多后天调校和支持条件。

The UI and report copy must avoid "贵贱", "高低人等", "必富", "必败", and similar absolute or status-loaded wording.

Alternative considered: labels such as "上等命/中等命/下等命". Rejected because they are harsh, commercially risky, and less useful than a capability-oriented metaphor.

### Decision 4: Score natal vehicle from existing explainable signals

Initial natal score inputs:

| Signal | Weight | Direction |
| --- | ---: | --- |
| Day-master strength balance | 25 | Neutral to moderately strong/weak is easier to use; extreme strength or weakness requires more correction. |
| Tiaohou completeness | 25 | Expected Tiaohou stems appearing in exposed or hidden form increase configuration completeness. |
| Ming Ge clarity | 20 | Clear classical pattern increases structure clarity; "杂气格" is treated as complex rather than bad. |
| Yongshen/Jishen and Ten God confidence | 15 | Clear hard/medium confidence improves explainability. |
| Natal risk modifiers | 15 | Severe collision/void/harsh Shen Sha density can lower ease-of-use; auspicious support can soften. |

The output evidence must include source, label, impact, score delta, and detail so users can see why a profile received its grade.

Alternative considered: base the vehicle grade mostly on Ming Ge. Rejected because Ming Ge alone cannot evaluate climate adjustment, body strength, or whether the chart is easy to drive in practice.

### Decision 5: Score Dayun roads from phase-aware Dayun evidence

Initial Dayun road score inputs:

| Signal | Weight | Direction |
| --- | ---: | --- |
| Jin Bu Huan front/back five-year ratings | 40 | `吉/平/凶` directly map to road phase quality. |
| Dayun stem/branch against Yongshen/Jishen | 25 | Favorable elements improve road support; adverse elements reduce it. |
| Dayun Ten Gods against favorable/adverse Ten God lists | 20 | Uses existing hard/medium/soft confidence. |
| Di Shi / life-stage strength | 10 | Stronger Dayun branch phase improves road traction. |
| Shen Sha and interaction modifiers | 5 | Compact correction only; avoids over-weighting single stars. |

Road score labels:

- `highway`: 85-100, 高速路.
- `main_road`: 70-84, 城市主路.
- `mountain_road`: 55-69, 山路.
- `muddy_road`: 40-54, 泥路.
- `construction`: 0-39, 施工路段.

Each Dayun also exposes `qian_road` and `hou_road` so a decade can be summarized as "高速接山路", "泥路转主路", or "前稳后冲".

Alternative considered: only use Jin Bu Huan. Rejected because Jin Bu Huan is valuable but does not fully account for the user's computed Yongshen/Jishen and Ten God confidence.

### Decision 6: Additive JSON fields, no first-version migration

The calculation result adds fields but does not remove or rename any existing fields. Saved chart JSON can carry the new fields naturally after recalculation. Old saved charts must degrade gracefully in the frontend and may derive partial UI from existing `dayun`, `yongshen`, `jishen`, `tiaohou`, and `wuxing` data when profile fields are missing.

Alternative considered: normalize profiles into new database columns. Rejected for the first version because the data is derived from chart input and existing result JSON; introducing persistence schema now adds migration surface without clear value.

### Decision 7: UI first version uses dense product cards and timeline labels

The result overview adds two compact modules:

- "命盘座驾": grade, vehicle type, one-line summary, tags, and evidence link.
- "当前路况": current Dayun road type, front/back phase summary, current-year note, and link to Dayun timeline.

The Dayun timeline adds road labels to each Dayun card and a phase indicator for front/back five years. It must preserve existing Dayun selection, Liunian drawer, Shen Sha modal, and mobile two-column behavior.

Alternative considered: make a full-screen animated vehicle-road scene. Deferred because the first version should prove the scoring model and user comprehension before investing in visual complexity.

## Risks / Trade-offs

- [Risk] Users may read grades as social rank or deterministic destiny. -> Mitigation: copy must frame grades as "配置完整度/驾驭难度" and avoid value-loaded labels.
- [Risk] Scoring weights may feel arbitrary. -> Mitigation: expose evidence deltas, add focused tests, and keep weights in one backend module for later tuning.
- [Risk] Frontend and backend duplicate interpretive logic. -> Mitigation: move authoritative scoring and summary labels to backend; frontend only renders and lightly degrades old data.
- [Risk] Old chart snapshots lack new fields. -> Mitigation: feature-detect fields, display old overview normally, and derive only limited fallback labels where reliable.
- [Risk] AI reports may overstate road labels. -> Mitigation: prompt must say the labels are metaphorical and forbid absolute outcome claims.
- [Risk] Road labels may compete with existing relative trend chart. -> Mitigation: present road map as cross-Dayun absolute context and trend chart as within-Dayun yearly rhythm.

## Migration Plan

1. Add backend profile data structures and deterministic scoring functions.
2. Call the scoring functions near the end of `Calculate()` after Ming Ge, Tiaohou, Gong Jia, Ten God relation, and favorable/adverse Ten God lists are available.
3. Update API serialization and TypeScript types for additive fields.
4. Render profile cards and timeline labels behind field-presence checks.
5. Inject vehicle/road context into report prompt assembly.
6. Add backend and frontend tests.
7. Deploy without database migration.

Rollback:

- Remove the frontend render paths or hide them by field-presence checks.
- Stop injecting profile data into AI prompts.
- The additive backend fields can remain harmless if consumers ignore them, or be removed in a later cleanup.

### Decision 8: Chinese configuration labels and a visible grade scale are the primary explanation

The ordinary-user result view will show the Chinese configuration label as the primary grade name, with `S/A/B/C/D` retained only as the algorithmic marker. The vehicle card will render the complete S-D grade scale by default. Dayun road definitions remain available as a secondary expandable explanation.

Rationale:

- A bare letter grade or words such as "顶配" do not explain what is being evaluated.
- Making the grade scale visible in context ensures users do not have to discover a collapsed control before they understand their own result.
- Neutral labels such as "特性型" and "调校型" are less likely to be read as social ranking than "偏科" or "需调校".

## Open Questions

- Should `construction` represent only very low score, or should it also trigger on交脱年/phase transition even if the decade score is moderate?
- Should the first version include yearly weather labels, or leave Liunian metaphor for a later change after Dayun road comprehension is validated?
