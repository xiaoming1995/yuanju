## Context

`BaziResult` already contains the raw data needed to explain ten-god relationships:

- day master: `day_gan`, `day_gan_wuxing`
- heavenly-stem ten gods: `year_gan_shishen`, `month_gan_shishen`, `day_gan_shishen`, `hour_gan_shishen`
- hidden stems: `year_hide_gan`, `month_hide_gan`, `day_hide_gan`, `hour_hide_gan`
- hidden-stem ten gods: `year_zhi_shishen`, `month_zhi_shishen`, `day_zhi_shishen`, `hour_zhi_shishen`

The current result page displays these terms in the professional grid, but it does not teach the user that the day stem is the fixed reference point. This makes the chart feel technical instead of explanatory.

The desired experience is:

```text
命主日元：丙火
        │
        ├── 年干 壬 = 七杀：克我，外部压力与竞争
        ├── 月干 甲 = 偏印：生我，资源与学习
        ├── 日干 丙 = 日主：命主自身
        └── 时干 戊 = 食神：我生，表达与输出
```

## Goals / Non-Goals

**Goals:**

- Make the day master explicit as the reference point for all ten-god relationships.
- Present heavenly-stem and hidden-stem ten-god relationships in a user-readable module.
- Keep the professional grid available for advanced users while adding a clearer interpretation layer.
- Support existing saved chart snapshots even if they only contain the current raw arrays.
- Keep the mobile layout readable by using stacked cards or collapsible sections.

**Non-Goals:**

- Rework the core ten-god calculation formula.
- Replace the existing professional bazi grid.
- Generate a full AI narrative in this module.
- Change report generation prompts unless later implementation finds reuse valuable.

## Decisions

### Decision 1: Add a structured relation matrix but keep frontend fallback derivation

The backend should expose a structured relation matrix when possible:

```json
{
  "day_master": {
    "gan": "丙",
    "wuxing": "火",
    "label": "丙火"
  },
  "heavenly_stems": [
    {
      "pillar": "year",
      "pillar_label": "年干",
      "gan": "壬",
      "wuxing": "水",
      "ten_god": "七杀",
      "relation": "克我",
      "summary": "外部压力、竞争、规则挑战"
    }
  ],
  "hidden_stems": [
    {
      "pillar": "year",
      "pillar_label": "年支",
      "branch": "申",
      "items": [
        {
          "gan": "庚",
          "ten_god": "偏财",
          "relation": "我克",
          "summary": "资源经营、机会与现实回报"
        }
      ]
    }
  ]
}
```

Rationale: structured data avoids fragile frontend index pairing and can be reused by PDF, AI report prompts, and future history views.

Alternative considered: render everything directly in `ResultPage` from existing arrays. This is faster but duplicates business meaning in the frontend and is fragile for hidden stems when old snapshots have incomplete arrays.

### Decision 2: Treat the day stem as the reference point, not a normal external ten-god

The day stem cell must be displayed as "日主 / 日元". Its explanation should say this is the chart owner, and all other ten-god labels are computed relative to it.

Rationale: showing the day stem as just "比肩" or another ordinary ten-god obscures the core rule and confuses users.

Alternative considered: keep the raw lunar-go day-stem ten-god label. This preserves existing raw data but does not solve the UX problem.

### Decision 3: Use ten-god group explanations for concise wording

The UI should explain both the exact ten-god and its broader group:

- 比肩 / 劫财 -> 比劫：自我、同辈、竞争、协作
- 食神 / 伤官 -> 食伤：表达、才华、输出、创造
- 正财 / 偏财 -> 财星：财富、经营、现实资源
- 正官 / 七杀 -> 官杀：规则、事业、压力、挑战
- 正印 / 偏印 -> 印星：学习、贵人、保护、资源

Rationale: this keeps the module readable for ordinary users while retaining exact labels for professional users.

Alternative considered: long per-label paragraphs. This would crowd the result page and duplicate AI report content.

### Decision 4: Put the module immediately after the basic chart

The result page should show "基本排盘" first, then the ten-god relation matrix immediately after it.

Rationale: the basic chart is the factual source users need to verify first. The ten-god relation matrix is an explanation layer for that chart, so it should follow the chart rather than compete for first position. Placing it after the report would still be too late.

Alternative considered: tooltip-only explanations inside the grid. This helps desktop users but is weak on mobile and does not provide a clear overview.

## Risks / Trade-offs

- [Risk] Old saved chart snapshots may not include all hidden-stem arrays or may have mismatched lengths. -> Mitigation: frontend and backend derivation must tolerate missing arrays and omit incomplete hidden items rather than guessing.
- [Risk] Extra UI could make the result page feel busy. -> Mitigation: use a concise summary card, then collapsible detailed rows on mobile.
- [Risk] Ten-god explanations can become too generic. -> Mitigation: keep this module deterministic and concise; leave individualized life-domain narrative to the AI report.
- [Risk] Backend field additions may require snapshot compatibility. -> Mitigation: do not remove existing fields; derive new fields lazily for old chart snapshots.

## Migration Plan

1. Add deterministic ten-god relation descriptors to the bazi package.
2. Add optional structured relation matrix fields to `BaziResult`.
3. Populate the matrix during new calculations.
4. For old snapshots, derive the matrix from existing chart fields when loading or rendering.
5. Add the frontend module with fallback behavior when the new field is absent.
6. Rollback is safe because existing result fields remain unchanged.

## Open Questions

- Should the first version include a compact "dominant ten-god group" summary, or keep this strictly as a relationship matrix?
- Should PDF export include this module immediately, or should PDF integration be a follow-up change?
