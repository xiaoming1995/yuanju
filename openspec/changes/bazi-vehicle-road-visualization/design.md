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

### Decision 3: Use S/A/B/C/D as natal-chart tiers, not human value

Vehicle grades represent the relative carrying capacity of the natal chart itself. They intentionally distinguish stronger and weaker base configurations, while never describing the worth of a person or guaranteeing an outcome:

- `S`: 上格配置。调候急需已充分解决，扶抑用神得力，结构稳定。
- `A`: 中上格配置。调候和扶抑主线成立，但仍有局部瑕疵。
- `B`: 中格配置。原局可用但短板明确，发挥更依赖顺运。
- `C`: 中下格配置。调候或扶抑至少一项关键不足，基础承载力偏弱。
- `D`: 下格配置。调候急病未解且扶抑无力，或原局风险显著集中。

The UI and report copy must distinguish natal-chart tiers from a person's value, social class, wealth, or a guaranteed life outcome.

### Decision 4: Use a gated, Tiaohou-first natal scoring model

Natal score inputs are evaluated in this order:

| Signal | Weight | Role |
| --- | ---: | --- |
| Urgent Tiaohou condition | 35 | First gate. For an extreme cold/hot chart, exposed relief is fully resolved, hidden relief is partial, and absence is an unresolved urgent condition. Non-urgent charts receive a neutral baseline rather than a bonus for merely matching a dictionary entry. |
| Fuyi Yongshen effectiveness | 30 | Second gate. Independently derive the Fuyi Yongshen, then assess whether it is exposed, rooted in hidden stems, or outweighed by exposed Jishen. |
| Day-master condition | 12 | Moderate strength is easier to use than extreme strength or weakness, but cannot override the first two gates. |
| Ming Ge clarity | 15 | A clear pattern can lift a chart after the two primary conditions are met; a complex or absent pattern receives less support. |
| Natal risk modifier | 8 | Harsh Shen Sha density can pull a tier down; helpful stars can only provide a small lift. |

Hard ceilings prevent unrelated positive signals from masking a core defect:

- An unresolved urgent Tiaohou condition caps the grade at `C`.
- A hidden-only urgent Tiaohou relief caps the grade at `A`.
- When Fuyi Yongshen has no usable support, the grade caps at `B`.
- Risk modifiers may lower the score but cannot raise a chart above its Tiaohou/Fuyi ceiling.

`ShishenConfidence` remains explanatory metadata. It must not add to the natal grade because confidence in an algorithmic inference is not a measure of chart quality.

The output evidence must include source, label, impact, score delta, and detail so users can see the Tiaohou gate, Fuyi assessment, and any applied ceiling.

Alternative considered: a base score with parallel positive additions from every existing signal. Rejected because it makes ordinary charts cluster at S/A and leaves C/D unreachable.

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

- [Risk] Users may read natal-chart tiers as social rank or deterministic destiny. -> Mitigation: copy must distinguish the base chart from a person's value and from guaranteed outcomes.
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
- Tier labels such as "上格配置" and "中下格配置" make the natal-chart distinction explicit, while the surrounding copy keeps it separate from a judgement of the person.

### Decision 9: Make the visible grade scale explain the tier criteria

The visible S-D guide will describe the same Tiaohou-first and Fuyi-second criteria used by the backend. It will call the output a natal-chart tier, clarify that it is not a judgement of the person, and explain that Dayun road conditions remain a separate external variable.

### Decision 10: Vehicle class follows natal tier; Ming Ge is professional driving context

The primary `vehicle_type` is a direct, stable mapping from the final natal grade:

- `S`: 超跑级座驾
- `A`: 高性能车级座驾
- `B`: 标准轿车级座驾
- `C`: 实用 MPV 级座驾
- `D`: 基础代步单车级

This makes the metaphor readable: the tier answers how capable the base vehicle is, while Dayun still answers what road it receives. The system must not infer the primary vehicle class from a Ming Ge, because no single Ming Ge is inherently high or low and the existing detection has a complex fallback.

Ming Ge may produce an optional `driving_style` explanation in professional mode only, such as stability-oriented, responsiveness-oriented, or high-control-demand. This is an interpretive aid, not a primary vehicle class or a social-status label. Brand names and model names are excluded because they do not form a stable or unambiguous hierarchy.

## Open Questions

- Should `construction` represent only very low score, or should it also trigger on交脱年/phase transition even if the decade score is moderate?
- Should the first version include yearly weather labels, or leave Liunian metaphor for a later change after Dayun road comprehension is validated?
