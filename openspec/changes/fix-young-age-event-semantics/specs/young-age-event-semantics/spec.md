## ADDED Requirements

### Requirement: 少年期年龄边界

系统 SHALL 定义常量 `YoungAgeCutoff = 18` 作为读书期与成人期的边界；流年信号引擎在 `age > 0 && age < YoungAgeCutoff` 时启用读书期分支，否则走成人期（现有）逻辑。

#### Scenario: 14 岁触发少年期分支

- **WHEN** `GetYearEventSignals` 被调用、参数 `age=14`
- **THEN** 任何被检测出的"财星 / 官杀 / 印星 / 食伤 / 比劫 / 桃花 / 合冲日支"事件都使用学业 / 性格语义 Type 与 evidence

#### Scenario: 18 岁回归成人期

- **WHEN** `GetYearEventSignals` 被调用、参数 `age=18`
- **THEN** 信号 Type 与 evidence 与现有成人期逻辑完全一致（财运_得 / 事业 / 婚恋_合 等）

### Requirement: 流年信号 GetYearEventSignals 增加 age 参数

`GetYearEventSignals` SHALL 接受额外的 `age int` 参数，并由 `GetAllYearSignals` 在迭代每个 LiuNian 时传入 `ln.Age`。

#### Scenario: 函数签名扩展

- **WHEN** 外部代码调用 `GetYearEventSignals(natal, lnGan, lnZhi, dayunGanZhi, gender, age)`
- **THEN** 函数能正确接受 6 个参数并按 age 分支

#### Scenario: GetAllYearSignals 内部传递 age

- **WHEN** `GetAllYearSignals` 遍历 `dy.LiuNian` 时
- **THEN** 每个 `ln.Age` 被传给底层 `GetYearEventSignals`

### Requirement: 财星透干在少年期映射为学业_资源

系统 SHALL 在 `age < YoungAgeCutoff` 且流年天干为命主财星（正财 / 偏财）时，输出 Type 为 `学业_资源` 的 EventSignal，evidence 描述家境 / 零用钱 / 物质条件层面的事件，不出现"财运提升 / 财星透出，财来财去"等成人期表述。

#### Scenario: 14 岁财星透干、非忌、身非弱

- **WHEN** age=14、流年透干为正财、命局财星五行不在 jishen、身强弱非弱
- **THEN** 输出 EventSignal{Type: "学业_资源", Polarity: 吉}
- **AND** evidence 字符串包含"少年"或"家境"或"零用钱"或"物质"关键词
- **AND** evidence 字符串不包含"财运"或"进财"

#### Scenario: 14 岁财星透干、为忌神

- **WHEN** age=14、流年透干为偏财、命局财星五行 ∈ jishen
- **THEN** 输出 EventSignal{Type: "学业_资源", Polarity: 凶}
- **AND** evidence 字符串描述财星为忌时的少年期负面信号（如"家庭经济波动"）

### Requirement: 比劫透干在少年期映射为学业_竞争

系统 SHALL 在 `age < YoungAgeCutoff` 且流年透干为命主比肩或劫财时，输出 Type 为 `学业_竞争` 的 EventSignal，evidence 描述同学竞争 / 兄弟情谊 / 团体活动相关事件。

#### Scenario: 14 岁比劫透干、身弱

- **WHEN** age=14、流年透干为比肩、命主身弱
- **THEN** 输出 EventSignal{Type: "学业_竞争", Polarity: 吉}
- **AND** evidence 提及"同伴帮扶 / 团体支持"

#### Scenario: 14 岁比劫透干、身强

- **WHEN** age=14、流年透干为劫财、命主身强
- **THEN** 输出 EventSignal{Type: "学业_竞争", Polarity: 凶}
- **AND** evidence 提及"同学竞争 / 友谊摩擦"

### Requirement: 官杀透干在少年期映射为学业_压力

系统 SHALL 在 `age < YoungAgeCutoff` 且流年透干为命主正官或七杀时，输出 Type 为 `学业_压力` 的 EventSignal，evidence 描述考试 / 升学 / 老师管教相关事件。

#### Scenario: 14 岁官杀透干、身弱

- **WHEN** age=14、流年透干为七杀、命主身弱
- **THEN** 输出 EventSignal{Type: "学业_压力", Polarity: 凶}
- **AND** evidence 提及"考试 / 升学 / 老师管教"等关键词

#### Scenario: 14 岁官杀透干、身强

- **WHEN** age=14、流年透干为正官、命主身强
- **THEN** 输出 EventSignal{Type: "学业_压力", Polarity: 吉}
- **AND** evidence 提及"光荣 / 班干 / 突出表现"等少年期正向描述

### Requirement: 印星透干在少年期映射为学业_贵人

系统 SHALL 在 `age < YoungAgeCutoff` 且流年透干为命主正印或偏印时，输出 Type 为 `学业_贵人` 的 EventSignal，evidence 描述师长指导 / 学习方法突破 / 师承机缘。

#### Scenario: 14 岁印星透干

- **WHEN** age=14、流年透干为正印
- **THEN** 输出 EventSignal{Type: "学业_贵人", Polarity: 吉}
- **AND** evidence 提及"师长 / 学习方法 / 资格"等少年期关键词

### Requirement: 食伤透干在少年期映射为学业_才艺

系统 SHALL 在 `age < YoungAgeCutoff` 且流年透干为命主食神或伤官时，输出 Type 为 `学业_才艺` 的 EventSignal，evidence 描述兴趣特长 / 才艺发展 / 表达欲望。

#### Scenario: 14 岁食伤透干、身强

- **WHEN** age=14、流年透干为食神、命主身强
- **THEN** 输出 EventSignal{Type: "学业_才艺", Polarity: 吉}
- **AND** evidence 提及"特长 / 才艺 / 表达"

#### Scenario: 14 岁食伤透干、身弱

- **WHEN** age=14、流年透干为伤官、命主身弱
- **THEN** 输出 EventSignal{Type: "学业_才艺", Polarity: 凶}
- **AND** evidence 提及"过度投入 / 分心 / 操劳"等身弱负面描述

### Requirement: 桃花 / 合日支在少年期映射为性格_情谊

系统 SHALL 在 `age < YoungAgeCutoff` 时将以下事件统一归入 `性格_情谊` Type：
- 流年地支与日支六合
- 大运地支与日支六合
- 神煞中"桃花"、"红艳"、"天喜"

evidence 描述友情 / 同窗情谊 / 初恋萌动 / 异性缘等少年期人际关系事件。

#### Scenario: 14 岁流年合日支

- **WHEN** age=14、流年地支与日支六合
- **THEN** 输出 EventSignal{Type: "性格_情谊", Polarity: 吉}
- **AND** evidence 不出现"婚恋"、"感情"、"夫妻宫"等成人期词汇
- **AND** evidence 提及"同窗 / 友情 / 心意相通"等少年期表述

#### Scenario: 14 岁桃花神煞临运

- **WHEN** age=14、流年地支被识别为日支或年支的桃花
- **THEN** 神煞输出被映射为 Type "性格_情谊"（而非默认的"婚恋_合"）

### Requirement: 冲日支在少年期映射为性格_叛逆

系统 SHALL 在 `age < YoungAgeCutoff` 时将以下事件统一归入 `性格_叛逆` Type：
- 流年地支冲日支
- 大运地支冲日支

evidence 描述情绪波动 / 家庭关系紧张 / 自我意识觉醒等少年期心理事件。

#### Scenario: 14 岁流年冲日支

- **WHEN** age=14、流年地支冲日支
- **THEN** 输出 EventSignal{Type: "性格_叛逆", Polarity: 凶}
- **AND** evidence 不出现"夫妻宫震动 / 感情变化"
- **AND** evidence 提及"情绪 / 家庭关系 / 自我"等少年期心理关键词

### Requirement: 少年期不输出婚恋财官双叠等成人期专属信号

系统 SHALL 在 `age < YoungAgeCutoff` 时跳过以下成人期专属事件检测：
- 财官双叠（成人期映射至婚恋_合）
- 男命财星透干 → 婚恋_合
- 女命官星透干 → 婚恋_合

#### Scenario: 14 岁财官双叠

- **WHEN** age=14、大运与流年同时含财星与官星透干
- **THEN** 输出列表中不存在 Type "婚恋_合"
- **AND** 输出列表中不存在 Type "婚恋_变" 由"财官双叠"路径产生的条目

### Requirement: 与年龄无关的事件保持现有逻辑

系统 SHALL 在所有年龄下保持以下事件检测不变（不分少年期与成人期）：
- 用神基底（Type: 用神基底）
- 喜神临运（Type: 喜神临运）
- 大运合化（Type: 大运合化）
- 大运地支冲流年地支（Type: 综合变动）
- 大运地支合流年地支（Type: 综合变动）
- 三会全局（Type: 综合变动 / 婚恋_合 — 但 14 岁时婚恋_合 应被映射为 性格_情谊）
- 三刑全局（Type: 健康 / 综合变动）
- 健康（克日干 / 冲日支 / 刑日支 / 自刑 — 健康在任何年龄都关键）
- 迁变（驿马等）
- 伏吟 / 反吟
- 空亡

#### Scenario: 14 岁流年克日干输出健康信号

- **WHEN** age=14、流年天干五行克日干
- **THEN** 输出 EventSignal{Type: "健康", Polarity: 凶}（与成人期一致）

#### Scenario: 14 岁流年补足三会引动 婚恋_合

- **WHEN** age=14、流年补足三会局且方向为日支同党
- **THEN** 输出 Type 为 "性格_情谊"（少年期重映射），而非 "婚恋_合"

### Requirement: 大运总结 prompt 按 youngRatio 动态调整

`GenerateDayunSummariesStream` SHALL 为每段大运计算 `youngRatio = (age<18 的年份数) / (该段非起运前的年份总数)`，并按以下三档调整 AI prompt：

| youngRatio | prompt 注入文案 |
|---|---|
| `== 1.0` | "本段大运全部年份处于读书期，请以学业、性格塑造、同窗关系为主轴撰写 summary。" |
| `> 0 && < 1.0` | "本段大运跨越读书期与成人期（前 N 年读书、后 M 年入社会），请分两段叙述。" |
| `== 0` | 不注入（保持现有模板） |

#### Scenario: 全段读书期大运

- **WHEN** 大运起 8 岁、终 17 岁，10 个流年全部 age < 18
- **THEN** youngRatio 为 1.0
- **AND** AI prompt 包含"本段大运全部年份处于读书期"

#### Scenario: 跨界大运

- **WHEN** 大运起 14 岁、终 23 岁，4 个流年 age < 18，6 个流年 age >= 18
- **THEN** youngRatio 为 0.4
- **AND** AI prompt 包含"跨越读书期与成人期"
- **AND** prompt 模板提示前 4 年与后 6 年分两段

#### Scenario: 全段成人期大运

- **WHEN** 大运起 28 岁、终 37 岁
- **THEN** youngRatio 为 0
- **AND** AI prompt 不附加任何读书期 / 跨界期文案

### Requirement: 前端 SIGNAL_LABEL 加入 7 个新 Type 映射

前端 `frontend/src/pages/PastEventsPage.tsx::SIGNAL_LABEL` SHALL 包含以下映射，颜色按学业系（var(--wu-mu)）与性格系（var(--wu-tu)）区分：

| Type | label | color |
|---|---|---|
| 学业_资源 | 学业↑ | var(--wu-mu) |
| 学业_竞争 | 竞争 | #888 |
| 学业_压力 | 压力↓ | #e77 |
| 学业_贵人 | 贵人 | var(--wu-mu) |
| 学业_才艺 | 才艺 | var(--wu-mu) |
| 性格_情谊 | 情谊 | var(--wu-tu) |
| 性格_叛逆 | 叛逆 | #e77 |

#### Scenario: 前端展示 14 岁年份卡

- **WHEN** 用户访问过往事件页面、查看 14 岁年份卡片
- **THEN** signals 数组中如含 "学业_资源" → chip 渲染"学业↑"木绿色
- **AND** signals 数组中如含 "性格_情谊" → chip 渲染"情谊"土黄色
