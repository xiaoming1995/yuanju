## Context

The Bazi engine already emits deterministic explanatory layers after core calculation: `natal_assessment` for natal structure, `vehicle_profile` for the natal vehicle metaphor, and `dayun_roadmap` for Dayun road conditions. The result page uses those fields in the first-screen summary and lets professional users inspect evidence through modal-style details.

The requested wealth feature should follow that same pattern. It must answer the common user question in direct language while preserving the platform's "algorithmic calculation + AI explanation" boundary: backend code owns the grade; frontend and AI only explain it.

## Goals / Non-Goals

**Goals:**
- Add a backend-computed wealth-structure profile with S/A/B/C/D grade, score, readable summary, tags, risk flags, and evidence.
- Keep the primary wealth grade scoped to natal wealth structure.
- Add a separate current-Dayun wealth window hint when Dayun or current-year Ten God context highlights money/resource themes.
- Render a concise "财富结构" overview near the existing vehicle and road summary, with professional evidence available on demand.
- Keep older saved charts readable through lazy backfill from existing `result_json`.

**Non-Goals:**
- Do not estimate real-world assets, income, social class, or investment outcomes.
- Do not let AI recalculate or override the deterministic wealth grade.
- Do not change the existing vehicle grade, Dayun road score, Fuyi algorithm, Ming Ge detection, or Ten God raw fields.
- Do not add a new database table or migration for the profile; it belongs in the existing result snapshot.

## Decisions

### Add a separate `wealth_profile`

`wealth_profile` will be a first-class optional field on `BaziResult`, not a nested field under `vehicle_profile`. Vehicle grade describes broad natal configuration; wealth profile describes a money/resource subdimension. Keeping it separate prevents a chart from reading as globally high-grade merely because one wealth signal is strong.

Proposed structure:

```go
type WealthProfile struct {
    Grade       string            `json:"grade"`
    GradeLabel  string            `json:"grade_label"`
    Score       int               `json:"score"`
    WealthType  string            `json:"wealth_type"`
    Summary     string            `json:"summary"`
    Tags        []string          `json:"tags"`
    RiskFlags   []string          `json:"risk_flags"`
    Evidences   []ProfileEvidence `json:"evidences"`
    CurrentHint *WealthWindowHint `json:"current_hint,omitempty"`
}

type WealthWindowHint struct {
    DayunIndex int      `json:"dayun_index"`
    GanZhi     string   `json:"gan_zhi"`
    Level      string   `json:"level"`
    Label      string   `json:"label"`
    Summary    string   `json:"summary"`
    Evidences  []string `json:"evidences"`
}
```

### Score natal wealth structure from interpretable components

The score will use deterministic components already present in `BaziResult` and `NatalAssessment`:

- Wealth-star visibility and rooting: 正财/偏财 in exposed stems, branch main Ten God, and hidden stems.
- Carrying capacity: day-master strength and Fuyi support, especially weak/vweak charts carrying many wealth signals.
- Wealth-producing chains: 食神生财, 伤官生财, 财官相生, and usable 财格 evidence from `NatalAssessment.Pattern`.
- Favorability: whether wealth Ten Gods are in `FavorableShishen` or `AdverseShishen`, respecting `ShishenConfidence`.
- Retention risk: 比劫夺财, 财多身弱, 财坏印, harsh auxiliary risk, or adverse wealth-star evidence.

The builder should return evidence for every non-neutral component and cap the final grade when hard risks are present. Example grade labels:

- S: 财富结构通达
- A: 财富承接较好
- B: 财富线索可用
- C: 财富波动偏多
- D: 财富承接薄弱

### Keep Dayun activation separate from the natal grade

The current Dayun can make money/resource themes more visible, but it must not change the natal wealth grade. A current hint should be derived from the current Dayun's Gan/Zhi Ten Gods, Dayun Ten God power, current road condition, and Fuyi/Ten God favorability. It should read as "当前窗口" or "当前阶段提示", not as part of the base level.

### Backfill through existing snapshot upgrade flow

`LoadOrCalculateResult` already upgrades older snapshots when dependent deterministic fields are missing or stale. The implementation should add wealth-profile detection there and persist the upgraded `result_json` after rebuilding dependent deterministic profile data. Versioning can either reuse `NatalAssessmentVersion` if wealth is fully derived from the current natal assessment, or introduce a small `WealthProfileVersion` field if wealth rules need independent upgrades.

### Prompt context is authoritative, not advisory

The AI report prompt should include a compact wealth-profile block alongside vehicle and road context. The prompt must state that the grade is backend-computed and that AI may explain evidence and trade-offs but must not invent another wealth grade or describe guaranteed wealth, guaranteed failure, real assets, or investment advice.

### Overview layout uses compact cards

The result overview should support three summary cards at desktop widths and a stable single-column order on mobile. "财富结构" should display grade label, S/A/B/C/D marker, a meter, summary, tags, and a `查看财富依据` action in professional mode. Evidence display should reuse the existing overview modal pattern where possible.

## Risks / Trade-offs

- [Users read the grade as income or asset prediction] -> Use "财富结构" and "钱财资源承接" wording, and include copy that it is not real-world assets or investment advice.
- [Strong wealth-star presence inflates weak charts] -> Carrying capacity and Fuyi favorability must cap or penalize weak/vweak cases such as 财多身弱.
- [The first screen becomes crowded] -> Keep wealth content compact, cap tags, and use responsive three-card desktop / one-column mobile layout.
- [Overlap with the active Dayun phase-road change] -> Implement after or carefully rebase against `make-dayun-phase-road-user-friendly`, because both touch `ResultPage` overview and road presentation.
- [Older snapshots show inconsistent data] -> Add lazy backfill tests for missing `wealth_profile`, and ensure missing profile is optional on frontend until backfill completes.
