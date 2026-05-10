## Context

`GetYearEventSignals` 当前将 Layer 0（应期位置）、Layer 4（十神宏观）、Layer 6（神煞）三类信号平权叠加输出，没有主次关系。专业命理师确认三类信号存在明确的优先级与权重规则（13个问题逐一确认），需在算法中落实。

当前信号结构：`EventSignal{Type, Evidence, Polarity, Source}`，无权重字段。

## Goals / Non-Goals

**Goals:**
- 实现 Layer 0 与 Layer 4 之间的压制规则（方向冲突时 Layer 0 凶压制 Layer 4 吉）
- 实现 Layer 0 内部权重：地支冲=刑 > 穿 > 合而不化；天干克增加五行流通判断
- 实现大运+流年双冲同一用神位时合并为叠加信号
- 实现神煞内部重煞/轻煞分级，重煞压制轻煞吉信号
- 日柱 > 月柱 > 年柱/时柱 的位置权重体现在信号描述中

**Non-Goals:**
- 不改变 EventSignal 的 JSON 结构（前端不受影响）
- 不修改神煞计算逻辑（仅在输出阶段分级）
- 不改变调候/扶抑用神的计算方式

## Decisions

**决策 1：Layer 0 vs Layer 4 压制在 `GetYearEventSignals` 末尾做后处理**

在所有层信号收集完毕后，统一扫描 signals 列表执行压制规则，而非在每层生成时判断。
- 优点：逻辑集中，易于调试和修改
- 备选：在 Layer 4 生成时即检查 Layer 0 结果 → 耦合太强

压制规则：
```
Layer 0 有凶信号 → 过滤掉同年 Layer 4 的吉信号（Source=柱位互动 且 Polarity=吉）
Layer 0 有吉信号 → 保留 Layer 4 凶信号（不压制）
Layer 0 无信号   → Layer 4 全部保留
```

**决策 2：天干克流通判断在 `collectYingqiSignals` 的 `checkGan` 中实现**

当检测到「流年/大运天干克原局用神天干位」时，额外检查：
- 取攻击五行 A、用神五行 B
- 在原局天干中查找是否存在五行 M，满足：A 生 M 且 M 生 B
- 若存在 → 信号降为中性（Polarity = 中性），Evidence 注明「五行流通，力度减弱」

**决策 3：地支冲刑穿合的权重通过 Evidence 文字体现，不增加数值字段**

命理师确认冲=刑 > 穿 > 合而不化，但实际输出是给 AI 生成叙事用的，无需数值权重，通过 Evidence 措辞区分即可：
- 冲/刑：「…受冲/刑，应期力度强，凶」
- 穿：「…受穿，应期力度中，凶」
- 合而不化：「…被锁，应期力度弱，凶」

**决策 4：大运+流年双冲同一用神位，在 `collectYingqiSignals` 返回后合并**

`collectYingqiSignals` 对流年和大运各自产生信号后，在调用方做后处理：
- 检测流年信号与大运信号是否指向同一位置（同一 pos 标签）且极性相同
- 若是，合并为一条，Evidence 标注「大运流年双重命中，力度倍增」

**决策 5：神煞重煞/轻煞通过在 `shenshaWhitelist` 增加 `IsHeavy bool` 字段区分**

重煞列表：羊刃、白虎、岁破、丧门、吊客、灾煞、劫煞、亡神

压制逻辑：当年存在重煞凶信号时，遍历神煞吉信号，若对应神煞 `IsHeavy=false`，则在 Evidence 追加「（本年有重煞，此信号仅作参考）」，Polarity 保持不变。

**决策 6：位置权重（日柱 > 月柱 > 年/时柱）体现在 Evidence 措辞中**

在 `collectYingqiSignals` 的信号 Evidence 中，对日柱宫位明确标注「（日柱宫位，权重较重）」，月柱标注「（月柱宫位，权重次之）」，年/时柱不额外标注。

## Risks / Trade-offs

- **压制规则复杂度**：Layer 0 vs Layer 4 的压制需要正确识别信号的 Source 和 Polarity，Source 字段需严格区分 → 确保 `SourceZhuwei` 只用于 Layer 0，不被其他层复用
- **流通判断的准确性**：「A 生 M 生 B」只检查原局天干，不检查地支藏干 → 按命理师确认规则执行，不扩展
- **双冲合并的边界**：流年大运指向同一用神位但互动类型不同（如流年冲+大运刑）是否合并 → 只合并相同互动类型的信号，不同类型独立保留
