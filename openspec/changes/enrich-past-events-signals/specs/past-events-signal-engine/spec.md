## MODIFIED Requirements

### Requirement: 流年事件信号检测
系统 SHALL 提供 `GetYearEventSignals(natal *BaziResult, lnGan, lnZhi, dayunGanZhi, gender string) []EventSignal` 函数，对指定流年计算激活的事件信号列表。

每个 `EventSignal` 包含：
- `Type`：事件类型（婚恋_合/婚恋_冲/婚恋_变/事业/财运_得/财运_损/健康/迁变/喜神临运/综合变动/用神基底/大运合化/伏吟/反吟）
- `Evidence`：触发本信号的命理证据描述
- `Polarity`（可选）：吉/凶/中性，由"流年五行 vs 命主用神/忌神"基底色与本信号性质共同决定
- `Source`（可选）：信号来源标签（神煞/柱位互动/合化/空亡/刑/会/伏吟/用神基底）

#### Scenario: 男命财星透干产生婚恋信号
- **WHEN** 男命流年天干为日干之偏财或正财
- **THEN** 返回信号列表中包含 Type=婚恋_合、Evidence 说明该财星透干及其含义、Source=柱位互动

#### Scenario: 女命官星透干产生婚恋信号
- **WHEN** 女命流年天干为日干之正官或七杀
- **THEN** 返回信号列表中包含 Type=婚恋_合、Evidence 说明该官星透干及其含义、Source=柱位互动

#### Scenario: 流年地支与日支六合引动夫妻宫
- **WHEN** 流年地支与日柱地支构成六合关系
- **THEN** Evidence 中注明"夫妻宫（日支）合住，感情宫位被激活"

#### Scenario: 流年地支与日支相冲引动夫妻宫
- **WHEN** 流年地支与日柱地支构成六冲关系
- **THEN** Evidence 中注明"夫妻宫（日支）受冲，感情宫位震动"

#### Scenario: 日主受克产生健康信号
- **WHEN** 流年天干五行克制日柱天干五行
- **THEN** 返回信号列表中包含 Type=健康、Evidence 说明克制关系

#### Scenario: 财星透干但财为忌神
- **WHEN** 流年天干为财星且财五行为命主忌神
- **THEN** Type=财运_得 信号的 Polarity=凶，Evidence 注明"财星虽透但为忌神，财来财去/破耗"

#### Scenario: 财星透干且财为用神
- **WHEN** 流年天干为财星且财五行为命主用神
- **THEN** Type=财运_得 信号的 Polarity=吉，Evidence 强调正向

#### Scenario: 无信号激活年份
- **WHEN** 该流年未触发任何规则
- **THEN** 返回空信号列表（[]EventSignal{}），不返回 error

---

## ADDED Requirements

### Requirement: 用神/忌神基底色信号
系统 SHALL 在每次 `GetYearEventSignals` 调用入口处，先比对流年天干五行与命主用神/忌神（来源：`natal.Yongshen` / `natal.Jishen`，若两者皆缺则降级为 `natal.Tiaohou.Expected`），输出一条 `Type=用神基底` 的整年定调信号；并据此为后续所有事件信号填写 `Polarity` 字段。

#### Scenario: 流年透用神
- **WHEN** 流年天干五行属于命主用神五行集合
- **THEN** 输出 Type=用神基底、Polarity=吉、Source=用神基底 的信号；后续所有事件信号默认 Polarity=吉

#### Scenario: 流年透忌神
- **WHEN** 流年天干五行属于命主忌神五行集合
- **THEN** 输出 Type=用神基底、Polarity=凶、Source=用神基底 的信号；后续所有事件信号默认 Polarity=凶

#### Scenario: 流年五行中性
- **WHEN** 流年天干五行既非用神也非忌神（含日主同党的均衡情形）
- **THEN** 输出 Type=用神基底、Polarity=中性 的信号；后续事件信号 Polarity 由本身性质决定

#### Scenario: 用神信息缺失
- **WHEN** `natal.Yongshen` 与 `natal.Jishen` 均为空且无法从 `Tiaohou.Expected` 降级
- **THEN** 不输出 Type=用神基底 信号；后续信号 Polarity 字段留空

---

### Requirement: 神煞流年信号接入
系统 SHALL 调用现有 `pkg/bazi/shensha_dayun.go` 的神煞计算结果，将白名单内的神煞按"流年是否触发"转换为 `EventSignal`。

白名单与默认 Polarity / Type 映射：
- 天乙贵人 / 天德 / 月德 → 吉 / 事业
- 羊刃 / 血刃 → 凶 / 健康
- 白虎 / 丧门 / 吊客 → 凶 / 健康（白虎额外可叠加迁变）
- 华盖 → 中性 / 迁变
- 红艳 → 中性 / 婚恋
- 将星 → 吉 / 事业
- 勾绞 → 凶 / 综合变动
- 孤辰 / 寡宿 → 中性 / 婚恋

每条神煞信号 Source=神煞。

#### Scenario: 流年触发天乙贵人
- **WHEN** 流年地支为命主天乙贵人位
- **THEN** 返回信号列表中包含 Type=事业、Polarity=吉、Source=神煞、Evidence 注明"天乙贵人临运"

#### Scenario: 流年触发羊刃
- **WHEN** 流年地支为命主羊刃位
- **THEN** 返回信号列表中包含 Type=健康、Polarity=凶、Source=神煞、Evidence 注明"羊刃临运，宜防开刀血光"

#### Scenario: 流年触发白虎
- **WHEN** 流年地支为命主白虎位
- **THEN** 返回信号列表中包含 Type=健康（必）与 Type=迁变（可选）、Polarity=凶、Source=神煞

#### Scenario: 神煞引擎被禁用
- **WHEN** `algo_config` 中神煞总开关关闭或目标神煞被关闭
- **THEN** 该神煞不输出信号

---

### Requirement: 流年与年/月/时柱互动检测
系统 SHALL 扫描流年地支与年柱、月柱、时柱地支的合（六合）、冲（六冲）、刑（含三刑）关系，并按柱位映射事件类型。

柱位事件映射：
- 年柱 → 祖荫/根基/家族 → Type=综合变动 或 健康
- 月柱 → 行业/工作/父母 → Type=事业 或 迁变
- 时柱 → 子女/晚景 → Type=综合变动（中年以前）或 健康（中老年）

每条互动信号 Source=柱位互动。

#### Scenario: 流年地支冲月支
- **WHEN** 流年地支与月支构成六冲
- **THEN** 输出 Type=事业、Source=柱位互动、Evidence 注明"流年冲月柱（提纲），易有行业/职位变动"

#### Scenario: 流年地支合时支
- **WHEN** 流年地支与时支构成六合
- **THEN** 输出 Type=综合变动、Source=柱位互动、Evidence 注明"流年合时柱（子女宫）"

#### Scenario: 流年与年支构成六冲（岁破）
- **WHEN** 流年地支与年支构成六冲
- **THEN** 既有"健康"+"迁变"信号继续输出，并补充 Source=柱位互动 标记

---

### Requirement: 伏吟与反吟检测
系统 SHALL 检测流年干支与四柱及当前大运干支是否构成伏吟（干支完全相同）或反吟（天克地冲：天干相克 + 地支六冲）。

#### Scenario: 流年与日柱伏吟
- **WHEN** 流年干支与日柱（DayGan+DayZhi）完全相同
- **THEN** 输出 Type=伏吟、Polarity=凶、Source=伏吟、Evidence 注明"流年伏吟日柱，主自身重大事件重现/旧事重提"

#### Scenario: 流年与月柱反吟
- **WHEN** 流年天干克月干 且 流年地支冲月支
- **THEN** 输出 Type=反吟、Polarity=凶、Source=伏吟、Evidence 注明"流年反吟月柱，事业/家庭剧变"

#### Scenario: 流年与大运伏吟
- **WHEN** 流年干支与当前大运干支完全相同
- **THEN** 输出 Type=综合变动、Polarity=凶、Source=伏吟、Evidence 注明"流年伏吟大运"

---

### Requirement: 空亡检测
系统 SHALL 按日柱所在旬计算"旬空"地支二位，对流年地支与大运地支落空亡的情形输出降权信号。

#### Scenario: 流年地支落空亡
- **WHEN** 流年地支位于命主日柱旬空二位
- **THEN** 输出 Type=综合变动、Polarity=中性、Source=空亡、Evidence 注明"流年落空亡，事件虚而不实/过而不留"，并对该年其他信号在 evidence 中追加"受空亡影响，力度减半"提示

#### Scenario: 大运地支落空亡
- **WHEN** 当前大运地支位于命主日柱旬空二位
- **THEN** 输出 Type=综合变动、Polarity=中性、Source=空亡、Evidence 注明"大运落空亡，整段大运能量打折"

---

### Requirement: 三会局检测
系统 SHALL 检测流年地支与原局四支+大运支是否构成三会局（寅卯辰会木、巳午未会火、申酉戌会金、亥子丑会水），三会触发能量强度高于三合。

#### Scenario: 流年地支引动三会木局
- **WHEN** 流年地支为寅或卯或辰，且原局/大运中其余两支已凑齐
- **THEN** 输出 Type=综合变动、Source=会、Evidence 注明"引动三会木局，能量场剧烈"

#### Scenario: 三会局五行匹配感情星
- **WHEN** 三会五行 = 命主财星五行（男命）或官星五行（女命）
- **THEN** 额外输出 Type=婚恋_合、Source=会 的信号

---

### Requirement: 三刑全局检测
系统 SHALL 在既有二刑判定基础上，额外检测寅巳申、丑未戌三刑是否在"原局四支+大运支+流年支"中凑齐。

#### Scenario: 流年补足寅巳申三刑
- **WHEN** 原局/大运已有寅巳申其中两支，流年地支补齐第三支
- **THEN** 输出 Type=健康、Polarity=凶、Source=刑、Evidence 注明"凑齐寅巳申三刑，主官非/手术/伤病"

#### Scenario: 流年补足丑未戌三刑
- **WHEN** 原局/大运已有丑未戌其中两支，流年地支补齐第三支
- **THEN** 输出 Type=健康、Polarity=凶、Source=刑、Evidence 注明"凑齐丑未戌三刑，主家庭/事业纠葛"

---

### Requirement: 大运天干合化日干检测
系统 SHALL 检测当前大运天干与日干是否形成天干五合（甲己/乙庚/丙辛/丁壬/戊癸），并判定是否成立化神条件。

化神条件（合而成局，三条件全部满足）：
1. 大运天干与日干构成五合
2. 月支或大运地支提供化神五行根气
3. 原局无强力反克化神

#### Scenario: 大运合日干且化神成立
- **WHEN** 大运天干与日干形成五合 且 月支或大运地支为化神五行根气 且 原局无强反克
- **THEN** 输出 Type=大运合化、Polarity=中性、Source=合化、Evidence 注明合化五行（如"丁壬合化木，整段大运日主性向偏木"）

#### Scenario: 大运合日干但化神不成立
- **WHEN** 大运天干与日干形成五合 但 化神条件不满足
- **THEN** 输出 Type=综合变动、Polarity=中性、Source=合化、Evidence 注明"日干被合住，能量受牵制但未成化局"

---

### Requirement: 加权身强弱评分
系统 SHALL 重写 `dayMasterStrength` 函数，从粗粒度三档（strong/weak/neutral）扩展为加权评分五档（vstrong/strong/neutral/weak/vweak）。

权重规则：
- 月支与日干关系（得令）：×5
- 其余地支本气与日干关系（得地）：×3
- 藏干透出 + 天干生扶/克泄（得势）：×2

档位阈值（默认，可由 `algo_config` 调整）：
- vstrong：≥10
- strong：5–9
- neutral：-4–4
- weak：-9 to -5
- vweak：≤-10

#### Scenario: 边界身强误判用例修正
- **WHEN** 命主月支克日主、其余地支多藏比劫的"看似身弱实则中和"格局
- **THEN** 加权评分输出 neutral，与原算法的 weak 形成区别

#### Scenario: 极强身打开从格判读路径
- **WHEN** 命主无任何克泄（全部生扶）、评分≥10
- **THEN** 输出 vstrong，便于后续 prompt 让 AI 注意"专旺/从强"路径
