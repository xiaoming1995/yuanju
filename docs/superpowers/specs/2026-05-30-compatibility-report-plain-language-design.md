# 合盘报告「说人话」润色 — 设计文档

日期：2026-05-30

## 问题

合盘分析报告里给用户看的语句几乎全是八字术语（年支六合、纳音相生、日柱上档、外围层有支撑…），普通用户读不懂、读着冷。需要用人话润色，同时不丢专业可信度。

## 已确认决策

1. **范围**：两层都改——LLM 深度解读正文 + 确定性证据卡/分数解释。
2. **术语去留**：保留术语做锚点 + 补人话解释（标题留「年支六合」，正文用人话讲它意味着什么），不彻底去术语化。
3. **语气**：温和顾问口吻——像一个懂行又体贴的人在跟当事人解释，不端着、不冷冰冰。
4. **文案密度**：详细（四五句）——术语锚点 + 解释它意思 + 落到相处的具体表现 + 这个信号对关系意味着什么 + 一句温和提醒。

## 关键事实：存储模型

`CreateCompatibilityReading` 生成时就把 `evidences` / `score_explanations` / `summary_tags` 写进 DB（`compatibility_evidences` 表 + jsonb 列）；打开结果页只读回存好的文案，不重算。

**含义**：改 Go 模板只对**新生成的合盘**生效；已存在的旧报告保留旧术语文案（与改 prompt 版本号的行为一致）。验证必须新建合盘。

## 架构

```
┌─────────────────────── 合盘报告页 ───────────────────────┐
│  ① LLM 深度解读正文  ← prompt 生成（canonical_compatibility.go）│
│  ② 证据组卡片 + 分数解释  ← Go 固定模板（确定性，存库）          │
└──────────────────────────────────────────────────────────┘
```

### Layer ② — 确定性文案（`backend/pkg/bazi/compatibility_evidence.go`）

**渲染落点核查（重要）**：全站只有两处把确定性文案展示给用户——
- `EvidenceDrawer` 渲染 `evidence.title`（术语锚点）+ `evidence.detail`（正文）。
- `CompatibilityShareCard` 读 `score_explanations`，但**只取 `positive_factor`（术语标题）**。

`scoreExplanationSummaryV3` 生成的那段 `Summary` 正文**用户不可见**，只作为 `{{.ScoreExplanationsJSON}}` 喂 LLM。因此**不改它**——保留简洁精确的术语 grounding 更利于 LLM 推理后再用人话讲给用户（避免为零用户收益撑大 prompt）。

固定字符串，改了就稳定。**保留** `Title` / `Type` / `EvidenceKey`（术语锚点不动），**只重写** 用户真正看得到的 `Detail`（证据卡正文）为温和顾问口吻、四五句密度。

涵盖范围（共 10 条 Detail）：
- `zodiacEvidence`：4 条 Detail（liuhe / sanhe / same_element / sheng）
- `nayinEvidence`：2 条 Detail（sheng / same）
- `dayPillarEvidence`：3 条 Detail（upper / lower / safe）
- `eightCharsEvidence`：1 条 Detail 模板（去掉「贡献 N」技术表述，改成人话）

**不改**：
- `scoreExplanationSummaryV3` / `eightCharsSummary` 的 `Summary`（用户不可见，仅喂 LLM；保持简洁术语 grounding）。
- `buildSummaryTagsV3` 的短标签（「上吉合盘」「纳音同气」…）——它们本身就是术语锚点，且 `compatibility_test.go` 有断言引用。

改写示范（标尺，四五句密度）：

| 位置 | 现状（术语） | 改后（温和顾问口吻 · 四五句） |
|---|---|---|
| 证据卡 `年支六合` | 双方年支 子/丑 构成六合，属相基础线吸引力强。 | 你俩的属相是天生的「六合」——这是命理里最顺的一种属相搭配。落到相处上，就是你们见面容易互相来电、自来熟，很多事不用刻意经营就能对上眼。在一段关系里，这种天然的亲近感会让你们在磨合期少很多摩擦。不过它管的是"合不合得来"，长久还得看两人愿不愿意一起经营。 |

> ⚠️ 这些 Detail 同时会喂给 LLM 作为 grounding。改成人话后 LLM 拿到的也是人话，但 Title 仍带术语做锚，反而更利于它说人话。可接受。

### Layer ① — LLM 报告正文（`backend/pkg/prompt/canonical_compatibility.go`）

新增一段**输出语言约束**（只约束输出语言，不动现有给 LLM 推理用的术语规则段——LLM 需要精确术语来推理）：

```
表达约束（面向普通用户）：
- 用温和顾问口吻，像一个懂行又体贴的人在跟当事人解释，不端着、不冷冰冰。
- 任何八字术语（六合、纳音、日柱、十神…）首次出现时，紧跟一句大白话解释它意味着什么；不得整句堆术语。
- 把判断说透、不要惜字：除了下结论，也讲清"为什么"和"落到相处上是什么样"。
- summary / judgment / personality / strategy / advice 全部用日常语言，说法落到"你们 / 对方 / 相处"这种当事人能直接对号的词。
```

版本号 `v3.1-question-aware-2` → `v3.1-question-aware-3`，同步改 `canonical_test.go` 的版本断言。

## 成功标准 / 验证

1. `go test ./pkg/bazi/... ./pkg/prompt/...`、`go vet ./...`、`gofmt -w` 全绿（Go 文案改动测试半径≈0，只需跟版本断言；无测试断言这些 Detail/Summary 文案）。
2. 在 :5200 **新建一个合盘**（旧报告不受影响）→ 证据卡正文读起来是人话四五句、标题仍是术语。
3. 点「生成深度解读」→ 正文温和顾问口吻、术语都带解释、把话说透。

## 不做（YAGNI）

- 不加新字段、不做迁移。
- 不动 summary tags。
- 不动 prompt 里给 LLM 推理用的术语规则段（评分规则说明 / 性格画像约束）。
- 不碰前端组件（纯文案层；EvidenceDrawer 已有折叠/「查看完整依据」披露，四五句长文案落在其中无需改布局）。
