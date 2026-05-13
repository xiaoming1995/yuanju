# Past Events Dual-Layer Narrative Design

## Context

The past-events timeline currently exposes raw `EventSignal.Evidence` text through `RenderYearNarrative`. Those evidence strings are useful for algorithm debugging and professional review, but they read like stacked technical terms: "流年地支冲日支", "大运流年官杀双叠", "伏吟", "空亡". This makes the feature hard to understand for ordinary users and tiring even for professional users.

The chosen direction is a dual-layer presentation:

- Default layer: plain-language yearly interpretation in the voice of a professional metaphysics practitioner.
- Expert layer: expandable technical basis for users who want to inspect the 命理依据.

## Goals

- Make every yearly narrative understandable without prior Bazi knowledge.
- Preserve professional evidence instead of deleting or watering down the algorithm output.
- Keep the change focused on presentation and narrative shaping; do not change signal detection rules.
- Avoid fear-heavy or exaggerated wording. Risk years should read as practical caution, not alarm.

## Non-Goals

- Reworking `GetYearEventSignals` or changing the underlying signal priority model.
- Introducing an LLM call for each year.
- Replacing the dayun AI summary pipeline.
- Building a full glossary or tutorial system.

## Proposed Model

Keep the existing signal structure unchanged:

```text
EventSignal
  Type      internal category
  Evidence  technical basis
  Polarity  吉 / 凶 / 中性
  Source    signal source
```

Extend the user-facing year item with an expert evidence summary:

```text
PastEventsYearItem
  year
  age
  gan_zhi
  dayun_gan_zhi
  dayun_index
  signals
  narrative          plain-language yearly reading
  evidence_summary   compact technical basis for expandable UI
```

`narrative` remains the default display field. `evidence_summary` is optional UI detail and should not be required to understand the year.

## Narrative Rules

Each yearly reading should contain:

1. One main theme.
2. One supporting theme when signals justify it.
3. One practical reminder.

The text should avoid exposing technical terms such as:

- 流年地支
- 日支 / 月支 / 年支 / 时支
- 官杀 / 财星 / 比劫 / 食伤 / 印星
- 伏吟 / 反吟 / 空亡
- 三合 / 三会 / 六合 / 六冲

Those terms may appear in `evidence_summary`, but not in the default `narrative`.

Recommended tone examples:

- Instead of "流年地支冲月柱，易有行业/职位变动": "工作方向或学习环境容易出现调整，需要重新适应节奏。"
- Instead of "财星为忌，财来财去": "钱财机会看似增加，但也容易伴随支出和压力，宜稳健安排。"
- Instead of "白虎临运，宜防开刀、血光": "健康和出行要更谨慎，避免冒险和过度劳累。"

## Theme Priority

When multiple signals exist, the plain narrative should pick the dominant user-facing theme rather than listing every rule.

Priority:

1. Major movement or repeated strong conflict: 综合变动, 伏吟, 反吟, 大运流年双重命中.
2. Health and safety: 健康, heavy 凶神煞.
3. Relationship and interpersonal matters: 婚恋_合, 婚恋_冲, 婚恋_变, 性格_情谊, 性格_叛逆.
4. Career or studies: 事业, 学业_压力, 学业_贵人, 学业_才艺.
5. Money and resources: 财运_得, 财运_损, 学业_资源.
6. Movement and travel: 迁变.
7. Mild auspicious or neutral signals.

Polarity shapes tone:

- 凶: cautious, conservative, reduce risk.
- 吉: supportive, actionable, suitable to advance.
- 中性: watchful, observe changes, adjust calmly.
- Mixed: useful opportunities exist, but choices should be deliberate.

## Evidence Summary Rules

`evidence_summary` should remain concise. It should list 2-5 technical evidence snippets selected from the strongest signals.

Selection order:

1. Non-shensha 凶 signals.
2. Non-shensha 吉 signals.
3. Heavy shensha signals.
4. Other neutral or light shensha signals.

Evidence text can keep technical terms, but should remove duplicated long suffixes where possible. It is for expandable expert inspection, not default reading.

## API And Compatibility

The existing endpoint `POST /api/bazi/past-events/years/:chart_id` should continue returning the same fields and add `evidence_summary`.

Existing clients that only read `narrative` and `signals` remain compatible.

The older AI streaming endpoint can continue receiving full `EventSignal` data. This change focuses on the algorithm-template year list used by the current timeline page.

## Frontend Behavior

The yearly card should show:

- Year and age.
- Current signal badges.
- Plain-language `narrative`.
- A small expandable "命理依据" control when `evidence_summary` is present.

Expanded content should be visually secondary. It should not dominate the timeline or make the page feel like a debug panel.

## Testing

Backend tests should cover:

- Technical terms are not leaked into default `narrative` for common signal types.
- `evidence_summary` includes professional evidence when signals exist.
- Empty or weak-signal years still produce a stable plain-language narrative.
- Young-age signals produce school/personality wording rather than adult finance/romance wording.
- Mixed吉凶 years produce balanced wording.

Frontend tests should cover:

- Year cards render default narrative.
- Evidence summary is hidden by default and visible after expansion.
- Cards without evidence summary do not render an empty expansion control.

## Risks

- Over-simplification may hide useful nuance. The expert layer mitigates this by preserving technical evidence.
- Mapping many signal combinations to natural language can become brittle. Keep the first version rule-based and theme-oriented instead of trying to translate every evidence sentence literally.
- The new `evidence_summary` field adds API surface. Because it is additive, compatibility risk is low.
