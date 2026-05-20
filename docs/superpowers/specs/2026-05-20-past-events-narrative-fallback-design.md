# Past-Events 流年 narrative 防空白方案

**日期：** 2026-05-20
**模块：** past_events / dayun_summaries

## 背景与问题

`v3-progressive-compressed` 算法上线后，`ai_dayun_summaries.years` 中仍有 **10% 的流年 narrative 为空字符串**入库（120 个年份样本中 12 个）。空白来源分两类：

1. **AI 主动输出空白**（10/12）—— 该年 algorithm signals 弱（仅"用神基底"或基底+1 个弱信号），AI 守 `ValidateYearNarrative` 的规矩，选择不写。
2. **Validator 清空**（2/12）—— AI 写了 `用神位/伏吟/神煞名` 等关键词，但算法 evidence 找不到对应来源（如 `"narrative 出现 \"用神位\" 但算法 evidence 无对应来源"`），被强行清空。

前端 `PastEventsPage.tsx` 行为：
- `narrative=""` + 算法侧 signals 非空 → 渲染 fallback `"本年关键信号：印星合冲、月柱宫位。详见下方命理依据。"`
- `narrative=""` + 算法侧 signals 空 → 渲染 null（完全空白卡片）

用户反馈：chip fallback 太"干巴"，希望弱信号年也能有一段连贯批语，而不是仅仅显示信号清单。

## 设计目标

1. **保证 narrative 入库非空** —— 任何流年都至少有一句话。
2. **不增加 AI 调用成本** —— 不做二轮 AI 重试。
3. **不引入 hallucination 风险** —— 兜底文案不触发 `validatedKeywords` 中的 28 个关键词。
4. **存量数据不动** —— 12 条历史 `narrative=""` 保持现状，待用户主动重新生成时自然走新逻辑。
5. **可观测** —— 升级算法版本号方便后续比对空白率。

## 架构

```
┌─────────────────────────────────────────────────────────────┐
│  GenerateDayunSummariesStream (report_service.go)           │
│                                                              │
│   AI 输出 parsed.Years[i].Narrative                          │
│         │                                                     │
│         ▼                                                     │
│   ValidateYearNarrative                                       │
│   ├─ 通过 → 保留 AI narrative                                │
│   └─ 失败 → 清空                                              │
│         │                                                     │
│         ▼                                                     │
│   ★ 新增：narrative == "" 时调用                              │
│      bazi.RenderYearNarrativeWithFallback(dySignals[i])     │
│         │                                                     │
│         ▼                                                     │
│   写库 + 推送（yearsJSON 已保证逐年非空）                    │
└─────────────────────────────────────────────────────────────┘

┌── pkg/bazi/event_narrative.go ──────────────────────────────┐
│  RenderYearNarrative(ys)            ← 保持现状不动            │
│    返回 "" 当 <2 句                                          │
│                                                              │
│  ★ 新增 RenderYearNarrativeWithFallback(ys) string            │
│    1) result := RenderYearNarrative(ys)                     │
│    2) if result != "" → return result                       │
│    3) else → return makeMinimalFallback(ys)                 │
└─────────────────────────────────────────────────────────────┘
```

## 组件

### 1. `RenderYearNarrativeWithFallback(ys YearSignals) string` (新增)

包装 `RenderYearNarrative` —— 调用，若返空则走 `makeMinimalFallback`。

```go
// RenderYearNarrativeWithFallback 是 RenderYearNarrative 的非空包装。
// 调用方在确实需要"每年都有内容"时使用（如 AI 模式 + 兜底场景）。
// 不替代 RenderYearNarrative —— template 模式仍可保留空返契约。
func RenderYearNarrativeWithFallback(ys YearSignals) string {
    if s := RenderYearNarrative(ys); s != "" {
        return s
    }
    return makeMinimalFallback(ys)
}
```

### 2. `makeMinimalFallback(ys YearSignals) string` (新增，内部函数)

按 `用神基底` 信号的 `Polarity` 分支选择安全句式：

| 基底 polarity | 兜底文案模板 |
|---------------|--------------|
| `吉` | `{age}岁{ganzhi}年，命局基调向吉，本年信号较稀，运势相对平顺，宜稳中求进。` |
| `凶` | `{age}岁{ganzhi}年，命局基调偏凶，本年信号较稀，宜守不宜攻，谨慎处事。` |
| `中性` / 无基底信号 | `{age}岁{ganzhi}年信号较稀，运势相对平稳，按本段大运{dayunGanzhi}方向延展即可。` |

**安全性要求：** 兜底文案严禁包含 `validatedKeywords` 28 个关键词（`用神位/忌神位/喜神位/伏吟/反吟/大运合化/三会/三合/受冲/受刑/双重命中/力度倍增/驿马/桃花/华盖/白虎/丧门/吊客/灾煞/流霞/天医/天喜/天乙/天德/月德/文昌/太极/福星/红艳/孤辰/寡宿/羊刃/亡神/劫煞/披麻/咸池/勾绞/国印`）。由单元测试断言保障。

### 3. `report_service.go` 改写 `GenerateDayunSummariesStream` 处理段

替换现有 `for i, y := range parsed.Years` 循环（约 line 1428-1440）：

```go
for i, y := range parsed.Years {
    narrative := y.Narrative
    if narrative != "" {
        if ok, reason := ValidateYearNarrative(narrative, dySignals[i].Signals); !ok {
            log.Printf("[GenerateDayunSummariesStream] dayun=%d year=%d 校验失败：%s",
                dy.Index, y.Year, reason)
            narrative = ""
        }
    }
    if narrative == "" {
        narrative = bazi.RenderYearNarrativeWithFallback(dySignals[i])
        log.Printf("[GenerateDayunSummariesStream] dayun=%d year=%d 使用 template 兜底",
            dy.Index, y.Year)
    }
    validatedYears[i] = yearOut{Year: y.Year, GanZhi: y.GanZhi, Narrative: narrative}
}
```

### 4. Prompt 加固

在 `report_service.go` 的流式 dayun summary prompt（约 line 1202-1228）"narrative 撰写规则"块追加：

> **弱信号年安全措辞**：当一年的 evidence 仅含"用神基底"或基底+1 个弱信号时，可直接使用以下安全句式，禁止省略 narrative：
> - 「该年信号稀疏，运势相对平顺」
> - 「基底为吉/凶/中性，本年无明显波动」
> - 「与大运 {大运干支} 同调，按本段方向延展」
>
> 上述句式不含"用神位/忌神位/伏吟/反吟/神煞名"等需追溯的术语，可安全使用。

### 5. 算法版本号升级

`repository.CurrentAlgorithmVersion = "v3.1-narrative-guarded"`

理由：方便埋点对比 `v3-progressive-compressed` vs `v3.1-narrative-guarded` 的"`narrative=""`" 比例。

## 测试

### 单元测试 `pkg/bazi/event_narrative_test.go`

1. **TestRenderYearNarrativeWithFallback_AlwaysNonEmpty**
   - 构造 3 种 `YearSignals`：0 signals / 1 signal（仅基底）/ 多 signals
   - 断言返回值 `len > 0`

2. **TestMakeMinimalFallback_PolarityBranches**
   - 输入基底 polarity=吉/凶/中性，断言文案包含对应措辞片段

3. **TestMakeMinimalFallback_KeywordSafe**
   - 遍历 28 个 `validatedKeywords`，断言 fallback 文案任何 polarity 分支均不包含

### 服务层测试 `internal/service/report_service_test.go`

4. **TestGenerateDayunSummariesStream_FillsBlankNarrative**
   - Mock AI client 返回 `{narrative:""}` 的两个年份
   - 调用 `GenerateDayunSummariesStream`，断言入库的 `yearsJSON` 中两个年份的 `narrative` 字段均非空且不等于原 AI 输出

### 前端回归

无前端代码改动 —— `PastEventsPage.tsx` 现有逻辑（`y.narrative !== ''` 判定）自然消费新数据。

## 上线验证

1. 选 1996-02-08 男（已有 v3 数据的 chart）触发"重新生成 past-events"。
2. 查询 `ai_dayun_summaries` 该 chart 的所有 years，断言 0 条 `narrative=""`。
3. 抽查 2008/2017/2019/2028 等"原空白年"，肉眼检查 fallback 文案。
4. 30 天后跑 SQL 对比：
   ```sql
   SELECT algorithm_version,
          COUNT(*) FILTER (WHERE (y->>'narrative') = '') * 100.0 / COUNT(*) AS empty_pct
   FROM ai_dayun_summaries
   CROSS JOIN LATERAL jsonb_array_elements(years) y
   GROUP BY algorithm_version;
   ```
   期望 `v3.1-narrative-guarded` 的 `empty_pct = 0`。

## 不做的事（YAGNI）

- 不做存量数据回填脚本 / admin 按钮。
- 不做 AI 第二轮重试。
- 不做 narrative source tag（不区分"AI 写的 vs template 兜底的"）—— 用户无需感知。
- 不改 `RenderYearNarrative` 既有契约（template 模式仍可返 ""）。
- 不改前端代码。

## 风险

| 风险 | 缓解 |
|------|------|
| 兜底文案重复感强（同一段大运多个弱信号年都触发） | 模板内已嵌入 `{age}{ganzhi}{dayunGanzhi}` 三个变量；同段大运内重复时仍有干支/年龄差异，肉眼可分辨。验收阶段实际看效果。 |
| Prompt 加固后 AI 反而更激进编造 | Validator 仍在拦截关键词。最坏情况：AI 写错被 wipe → 仍走 fallback → 入库非空。无回归损失。 |
| Template 文案语气与 AI 文案风格不一致 | 接受。底线是"有内容"，文风一致是 nice-to-have，不在本次 scope。 |
