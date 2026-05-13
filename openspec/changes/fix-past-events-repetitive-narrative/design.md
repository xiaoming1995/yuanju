## Context

`backend/pkg/bazi/event_narrative.go` currently converts yearly `EventSignal` sets into plain-language `narrative` text. The implementation ranks the `change` theme first, and `themeOf()` maps `综合变动`, `伏吟`, `反吟`, `大运合化`, and `局势_重` into that theme.

For charts where many consecutive years contain broad `综合变动` signals, especially child-age years with dayun-liunian or pillar interaction signals, `pickDominantSignal()` repeatedly chooses `change` and `plainThemeSentence()` returns the same fixed sentence:

> 变化感会比较强，旧问题或新安排容易集中出现，生活节奏可能比平时更紧。

This makes the timeline look copied across years even when badges and evidence differ.

## Goals / Non-Goals

**Goals:**
- Make consecutive yearly narratives visibly distinct when their signal sets differ.
- Keep plain-language readability while preserving `evidence_summary` for professional inspection.
- Prioritize specific user-life themes over generic change wording, especially for age < 18.
- Add regression coverage for the 1996-02-08 20:00 style failure mode.

**Non-Goals:**
- Changing the underlying Bazi signal detection algorithm.
- Changing API response shape.
- Rewriting dayun summary generation.
- Adding LLM calls to produce per-year prose.

## Decisions

### Decision 1: Treat generic change as a modifier, not always the dominant theme

`综合变动` should no longer automatically outrank more specific themes. It should become the dominant theme only when it represents a strong change signal:

- `伏吟`
- `反吟`
- `大运合化`
- `局势_重`
- evidence containing "大运流年双重命中"
- evidence containing "力度倍增"
- evidence containing "重大事件高发"

普通 `综合变动` should act as context unless no more specific theme exists.

**Alternative considered:** Keep `change` as rank 1 and add more sentence variants. This still lets broad change signals crowd out school, relationship, and health signals, so the yearly focus remains too generic.

### Decision 2: Add life-stage-aware priority

For `age < 18`, narrative selection should prioritize:

1. 学业 / 师长 / 规则压力
2. 性格 / 同学 / 家庭沟通
3. 健康 / 作息 / 安全
4. 家庭资源
5. 强变动
6. 普通变动

Adult years can keep broader priorities, but ordinary `综合变动` should still not suppress a specific health, relationship, career, or finance signal.

**Alternative considered:** Reuse the same priority table for every age. The screenshot failure is concentrated in childhood years, so not differentiating life stage would keep adult-sounding phrasing in early timeline entries.

### Decision 3: Generate evidence-sensitive wording for change cases

When a strong or fallback change signal is selected, the sentence should vary by evidence/source:

- dayun-liunian collision: "外部节奏和环境要求更容易变化"
- month-pillar interaction: "学习方向、班级环境或老师要求容易调整"
- day-branch/self-palace interaction: "情绪、人际和家庭沟通更容易起波动"
- kongwang: "计划感强但落地感弱"
- fuyin: "旧事重提、同类问题反复"
- fanyin: "变化更剧烈，适合先稳住节奏"
- yima/movement: "出行、搬动或环境变化增加"

**Alternative considered:** Randomize wording among a phrase pool. Randomness would hide the symptom but not make narratives more accurate or testable.

### Decision 4: Regression test anti-repetition explicitly

Tests should create several adjacent child-age `YearSignals` with shared `综合变动` plus distinct secondary signals and assert:

- at least the opening sentence differs across years when specific signals differ;
- child-age narratives avoid adult default change boilerplate;
- technical evidence remains in `RenderEvidenceSummary()`.

**Alternative considered:** Test only individual phrase outputs. That would not catch the screenshot-level problem: repeated adjacent timeline cards.

## Risks / Trade-offs

- More priority rules can become hard to reason about → keep the logic local to `event_narrative.go` and cover with table-driven tests.
- Broad change signals may become underrepresented → still include them as fallback and in `evidence_summary`.
- Exact phrase tests can be brittle → test absence of repeated boilerplate and presence of life-domain words rather than full sentence equality.
