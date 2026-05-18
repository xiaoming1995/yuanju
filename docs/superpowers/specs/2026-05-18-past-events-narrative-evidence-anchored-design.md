# 过往事件推算 · narrative 强制 evidence 化设计

> 日期：2026-05-18
> 关联前序：`docs/superpowers/specs/2026-05-17-past-events-narrative-deduplication-design.md`、`openspec/changes/fix-past-events-repetitive-narrative/`

## 问题

`过往事件推算` 页（PastEventsPage）的年度白话批语在视觉上有两类规律性瑕疵：

1. **模板复读**：相邻年份开场白完全相同。截图中乙酉 2005 / 丙戌 2006 / 丁亥 2007 三年都以「这一年有机会也有压力，事情会同时出现可争取和需取舍的一面」开头。
2. **空泛化**：除前述开场白外，「触发点来自这一年的主导信号」「这一年最要紧的是稳住学习节奏」等收尾句出现频率极高，**整段读完不知道这一年具体发生了什么**。

根因不在信号引擎而在文案层 —— `backend/pkg/bazi/event_narrative.go` 的 6 个句子构造器都带有「找不到具体 evidence 时的兜底分支」，靠 polarity 计数硬凑一句话。`fix-past-events-repetitive-narrative` 已经做了两轮变体扩充，但只要兜底分支还在，多年份落进相同变体桶的问题不会消失。

## 决策

> 用户偏好：**「不能解决就别显示」** > 「凑话写一段」。

1. **不写就不写**：所有句子构造器中只靠 polarity 计数的兜底分支删掉，改为返回 `""`。能不能输出，由信号本身是否携带具体 evidence 决定。
2. **隐藏门槛**：当 6 个构造器输出的非空句子数少于 3 句时，`RenderYearNarrative` 整段返回 `""`，前端不渲染 narrative `<div>`。
3. **不引入 LLM**：保持「毫秒级算法生成 + 大运 AI 总结」的现有分层，不为每个年份调 AI。
4. **不改卡片骨架**：徽标 + 信号 chips + 命理依据展开 三块保留；narrative 隐藏后卡片只是变紧凑，不显示占位提示文案。
5. **不改 API、不动数据库**：`narrative: string` 字段保留，只是可能为 `""`。

## 「Evidence-anchored」定义

句子允许输出的充要条件：选择该分支时依赖了某个**具体输入差异** —— 信号的 `Type` / `Source` / `Evidence` 字段里的关键词、十神组别与年信号主题的对应关系。**仅靠 xiong/ji 计数、或仅靠"还有信号存在"** 不构成 evidence anchor。

句子构造器具体改造：

| 构造器 | 当前 | 新规则 |
|---|---|---|
| `yearToneSentence` | 极性计数 switch → 通用开场白 | 仅当 `isHardEventSignal(primary)` 为真时输出（沿用 healthLead / changeLead / relationshipLead / defaultHardLead）。所有极性分支 → `""`。 |
| `triggerSourceSentence` | 关键词全不命中时落 `default` 兜底 | 删 `default:` 分支，落空 → `""`。保留所有 keyword/Source/Type 分支。 |
| `domainDetailSentence` | 主题 = `default` 时输出"日常安排会出现新的侧重点" | 删 `default` 分支 → `""`。已知主题分支保留（多数本来就 evidence-aware）。 |
| `secondaryDetailSentence` | 已知主题都有输出，包括无关 evidence | 主题分支保留，但仅在 secondary signal 自身满足以下条件之一时才输出：(a) `isHardEventSignal(secondary)` 为真；(b) `secondary.Evidence` 命中关键词集合 `{冲, 刑, 空, 用神, 忌神, 驿马, 月柱, 日支, 大运流年双重命中}` 至少一个；(c) `secondary.Type` 属于 `{伏吟, 反吟, 大运合化, 局势_重, 学业_*, 性格_*, 婚恋_*}`。否则 → `""`。 |
| `tenGodNarrativeSentence` | 十神组别与年信号主题对不上时仍输出"背景力量"兜底 | 删"可作为理解这一年事件走向的背景力量"包装；仅保留"推到台前"那条带具体对应关系的分支。 |
| `practicalStanceSentence` | 总会输出收尾；含纯极性分支「这一年宜保守谨慎」 | 删除纯极性分支 → `""`。主题分支保留，但要求当年存在该主题的真实 signal。age<18 的主题分支保持。 |

另外，`RenderYearNarrative` 顶部「没有 meaningfulSignals → 输出'本年命理信号较弱'」的整段兜底也删掉 —— 用户的偏好是宁可不显示，不要凑话。

## 编排器

```
RenderYearNarrative(ys YearSignals) string {
  primary, ok := pickDominantSignal(ys.Signals, "", ys.Age)
  if !ok { return "" }
  secondary, hasSecondary := pickDominantSignal(ys.Signals, themeOf(primary.Type), ys.Age)

  sentences := []string{}
  if s := yearToneSentence(ys.Signals, primary); s != "" { sentences = append(sentences, s) }
  if s := triggerSourceSentence(primary, ys.Age); s != "" { sentences = append(sentences, s) }
  if s := domainDetailSentence(primary, secondary, hasSecondary, ys.Age); s != "" { sentences = append(sentences, s) }
  if hasSecondary {
    if s := secondaryDetailSentence(secondary, ys.Age); s != "" { sentences = append(sentences, s) }
  }
  if s := tenGodNarrativeSentence(ys.TenGodPower, primary, secondary, hasSecondary); s != "" { sentences = append(sentences, s) }
  if s := practicalStanceSentence(ys.Signals, primary, ys.Age); s != "" { sentences = append(sentences, s) }

  if len(sentences) < MinSentencesForNarrative { return "" }
  return ys.GanZhi + "年，" + joinNarrativeParts(sentences)
}
```

`event_narrative.go` 顶部新增 `const MinSentencesForNarrative = 3`，便于事后调旋钮。

门槛取 3 而非 4：用户期望 narrative 看起来是 4-5 句"段落"。GanZhi 自带「丙戌年，」前缀算一个起句元素，3 句正文拼起来读起来已是 4 元素，4 句正文为 5 元素，落在期望区间内。

## 前端

`frontend/src/pages/PastEventsPage.tsx` 第 417-419 行：

```tsx
{y.narrative && (
  <div style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', lineHeight: 1.7 }}>
    {y.narrative}
  </div>
)}
```

`narrative === ""` 时卡片自上而下只剩：干支徽标 + 年份/年龄 + 信号 chips + 命理依据按钮（展开块按钮始终保留）。不渲染空 `<div>`、不显示"暂无内容"、不显示加载态。

无 signals 且 narrative 为空的纯空卡仍保留在时间轴里（用户已确认 Section 3.A），保证年份连续性。

## 测试

在 `backend/pkg/bazi/event_narrative_test.go` 新增三组：

1. **隐藏门槛单元测试**：构造只含一个 `综合变动`、无 evidence 关键词的 YearSignals，断言 `RenderYearNarrative == ""`；再加一个泛 `健康` 信号，仍少于 3 句，断言 `== ""`。

2. **截图回归测试**：构造三组相邻童年期 YearSignals 模拟 乙酉 2005 / 丙戌 2006 / 丁亥 2007（都有 `综合变动` + xiong≥2 + ji>0，但 evidence 关键词分别带「空」「月柱」「日支」）。断言：
   - 三年的 narrative 要么全为 `""`，要么任意两年首句（GanZhi 之后第一句）不相同；
   - 任意一年的 narrative **不得包含**以下被砍兜底句：`"这一年有机会也有压力"`、`"触发点来自这一年的主导信号"`、`"这一年最要紧的"`、`"本年命理信号较弱"`。

3. **「evidence 必须项」契约测试**：table-driven，对每个 still-emitting 句子构造器，传入"只有极性、无 evidence 关键词"的输入，断言返回 `""`。把"删掉兜底"的契约钉死，防止未来 regression。

保留并继续通过：
- `TestRenderEvidenceSummary`（命理依据输出不变）
- `event_narrative_leads_test.go`（hard signal 仍会触发开场白）

手动验证：用 1996-02-08 20:00 命盘实跑 `/api/bazi/past-events/years/:chart_id`，童年期 10 个左右年份至少 5 年 `narrative == ""`，剩余年份开场白互不相同。

不做：snapshot 测试（diff 太大不利于复审）；前端组件测试（只改一行条件渲染）。

## 范围边界

**不做：**
- 大运 AI 总结（`GenerateDayunSummariesStream`、`ai_past_events` 表）—— 单独立题
- 信号引擎（`event_signals.go`、`pickDominantSignal`、`themeRank`、`isStrongChangeSignal`）—— 维持现状
- 信号 chips、命理依据（`ExtractYearSignalTypes`、`RenderEvidenceSummary`）—— 不变
- API 形状 —— `narrative` 字段保留，可能为 `""`
- DB 迁移 —— 无
- LLM 兜底 —— 显式拒绝（用户偏好"宁不显示"）

**改动文件：**
- `backend/pkg/bazi/event_narrative.go`（改 6 个构造器 + `RenderYearNarrative` + 加常量）
- `backend/pkg/bazi/event_narrative_test.go`（加 3 组测试）
- `frontend/src/pages/PastEventsPage.tsx`（一行条件渲染）

## 风险

| 风险 | 缓解 |
|---|---|
| 砍兜底后温和年份大面积空白 | 这是预期；`MinSentencesForNarrative` 是单常量旋钮，实测后可调 |
| `triggerSourceSentence` 关键词分支自身仍偏笼统 | 第一阶段及格线是"能溯源到 evidence 关键词"；如需更具体可后续做"句子模板词替换"，属下一轮迭代，不在本次 |
| `fix-past-events-repetitive-narrative` 已合主分支 | 本次延续而非冲突 —— 那次扩变体，这次砍兜底 + 加门槛；两者方向一致 |
| 大运 AI 总结自身也可能笼统 | 本次不在范围 |
| `ai_past_events` 缓存里旧 narrative 残留 | 无影响 —— per-year narrative 走 `GeneratePastEventsYears` 即时算，不读该表 |

## 接力

进入 `superpowers:writing-plans`，把本设计拆解为 step-by-step 实施计划，交付给执行阶段。
