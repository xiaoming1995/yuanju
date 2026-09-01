## Context

The active Dayun view already distinguishes a ten-year composite road condition from front and back Jin Bu Huan phase prompts. It currently leads the phase blocks with technical evidence (`天干主事`, `地支主事`, and `吉/平/凶`), which makes the intended practical meaning hard to scan.

## Goals / Non-Goals

**Goals:**
- Let a non-specialist read each five-year period as a direct road condition and a short focus theme.
- Preserve the existing distinction between the ten-year composite road and phase-level Jin Bu Huan evidence.
- Keep the selected-Dayun timeline summary consistent with the active card and keep compact cards unchanged.

**Non-Goals:**
- Change Jin Bu Huan, composite-road, or natal-chart scoring.
- Add backend fields or modify API contracts.
- Treat a phase road condition as a replacement for the ten-year composite conclusion.

## Decisions

### Enrich the shared phase presentation helper

The existing helper will expose the source phase road label and a concise theme derived from that label. The same mapping will be consumed by the active Dayun card and the selected timeline summary so wording cannot drift.

The themes are fixed readable guidance rather than calculated predictions:
- 高速路: 快速推进
- 城市主路: 稳步落地
- 山路: 选择与节奏
- 泥路: 稳住调整
- 施工路段: 修整蓄力

### Use road and theme as primary, evidence as secondary

Each phase will lead with `前五年阶段路况` or `后五年阶段路况`, its road label, and `主题：...`. The calendar range remains visible. Governing Gan/Zhi and Jin Bu Huan rating remain secondary labels for users who want the basis.

This keeps the previous semantic correction: the ten-year `road_label` remains the only `十年综合路况` result.

### Preserve compact timeline cards

Only the selected-Dayun summary receives phase road and theme content. The ten Dayun cards remain identity-first with one composite road badge, avoiding repeated long text.

## Risks / Trade-offs

- [Users may read a phase road as a new independent score] → The phase heading explicitly includes "阶段路况" and the parent card retains the visually distinct "十年综合路况" section.
- [Theme wording may feel repetitive across adjacent periods] → Themes stay concise and are paired with existing phase-specific detail and Gan/Zhi evidence.
- [Long labels can crowd the narrow current-Dayun card] → Use stable two-column desktop layout and a single-column mobile layout with wrapped text.
