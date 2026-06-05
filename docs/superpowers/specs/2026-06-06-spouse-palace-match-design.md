# 合盘「夫妻宫匹配」设计文档

日期：2026-06-06
状态：已通过 brainstorming 评审，待用户复核

## 1. 目标

在合盘报告里新增一节「夫妻宫匹配」：用一方命盘推出「TA 命里理想的另一半画像」，再拿对方真人去比，给出像不像。双向各做一遍。

**纯描述、不进 0–100 总分**（与现有 `personality_comparison` 同性质，延续上次「不动评分」红线）。

**用户原话**：「八字中不是可以根据命主的夫妻宫知道另一半的性格画像吗」——把这一点落成报告里的可读章节，并和对方的真实画像比对。

## 2. 关键决策（brainstorming 已确认）

| 决策点 | 结论 |
|---|---|
| 配偶画像来源 | **夫妻宫（日支藏干/十神）+ 配偶星（男财女官杀）一起推** |
| 是否进总分 | **否**。纯新增描述章节，不动 `DimensionScores`/总分/等级 |
| 实现路线 | **路线一**：Go 只补「配偶星定位」结构化信号；画像解读与高/中/低判定交给 LLM |
| 匹配档 | 高/中/低（`match_level`），由 LLM 判，但受 Go 已算好的夫妻宫合冲状态约束 |
| 性别缺失 | **跳过这一节**，注明「缺性别，无法定配偶星」 |
| 不做 | 不建「十神→性格」对照表、不引入真易经卦象、不做公式化相似度 |

## 3. 现状（探查结论）

- 十神（含**日支藏干的十神**）、日主旺衰、命格、五行统计——引擎已全部算好，且已序列化进 LLM 输入（`compatibilityParticipantSummary`，`internal/service/compatibility_service.go:591`）。
- 双方日支（夫妻宫）之间的**合/冲/克/刑/害**已算好：day_pillar 维度评分 + negative 证据（上次改动），已在 LLM 输入里，可直接当锚点。
- **唯一缺口**：没有「配偶星」识别——男财/女官杀目前只混在十神列表，没被单独挑出来标注（坐哪柱、透不透、入不入夫妻宫、强弱）。

可复用的现成件：`GetShiShen(dayGan, gan)`、`GetZhiShiShen`（`pkg/bazi/shishen.go`）；`DayHideGan` / 各柱藏干及其十神（`engine.go`）；`GetStrengthDetail`（`event_signals.go`，日主旺衰）。

## 4. 架构与改动点

### 4.1 新增：配偶星定位（`pkg/bazi`）

```go
// 给定一个人的盘 + 性别，定位其配偶星与夫妻宫画像信号。
func detectSpouseStarSignal(r *BaziResult) SpouseStarSignal
```

`SpouseStarSignal` 字段（结构化、给 LLM 当输入）：

| 字段 | 含义 |
|---|---|
| `Available bool` | 性别可用且能定配偶星时为 true；性别缺失时 false（→ 上层跳过本节） |
| `SpouseStarNames []string` | 配偶星十神：男 = 命中出现的 正财/偏财；女 = 正官/七杀 |
| `Positions []string` | 配偶星所在柱（年/月/日/时）及透干 or 仅藏干 |
| `InSpousePalace bool` | 配偶星是否正坐日支（夫妻宫） |
| `Visible bool` | 配偶星是否透于天干（透干 vs 仅藏） |
| `Present bool` | 八字里是否存在配偶星（不现 → 画像改从夫妻宫藏干推，需标注） |
| `StrengthLabel string` | 复用日主旺衰标签（`compatibilityStrengthLabels`） |
| `DayBranchHiddenShiShen []string` | 日支藏干各自的十神（夫妻宫画像主料） |

规则：
- 性别取自 `BaziResult.Gender`；男 → 配偶星为财星（正财/偏财，正财为主），女 → 官杀（正官/七杀，正官为主）。
- 性别为空/不可识别 → `Available=false`，其余留空。
- 配偶星不现（命中无财 / 无官杀）→ `Present=false`，`Available` 仍可为 true（画像降级从夫妻宫藏干推，文案注明）。
- 边界：缺日柱等异常盘安全返回 `Available=false`，对齐现有 nil 防护。

### 4.2 序列化进 LLM 输入（`internal/service/compatibility_service.go`）

- 在 `compatibilityParticipantSummary` 产出的每人摘要串尾部，追加「配偶画像信号」段：配偶星十神/位置/透藏/是否入夫妻宫/强弱 + 日支藏干十神。
- `Available=false`（缺性别）时，该段写明「性别缺失，配偶星不可定」，供 LLM 据此跳过本节。
- **不新增**双方夫妻宫合冲的计算：沿用已在输入中的 day_pillar 维度 + negative 证据当锚点。

### 4.3 Prompt 模板（`pkg/prompt/canonical_compatibility.go`）

- 新增输出字段 `spouse_palace_match`（见 §5）及生成指令：
  - `self` = 从 **A** 的配偶星 + 夫妻宫藏干十神推「A 理想另一半画像」，拿 **B** 真实画像比；`partner` = 反向。
  - `match_level` 高/中/低，**必须**与已知夫妻宫状态自洽：日支相冲/相克时不得给「高」；与 day_pillar / negative 证据矛盾的描述一律禁止（延续上次「禁止说假话」约束）。
  - 任一方 `配偶星不可定`（缺性别）→ 该方对应子块输出空/略，并在 `summary` 注明「缺性别，无法定配偶星，本节跳过」。
  - 配偶星不现 → 画像照出但注明「配偶星不显，结论偏轮廓」。
  - 文字守现有「表达约束」：术语后跟大白话、条件语气不下死命；画像尺度**微辣不露骨、不越线**（对齐日柱速写文案尺度）。
- 版本号 +1，更新并通过 `canonical_test.go` / `drift_test.go` / `sync_test.go` / 名人配对黄金测试。

### 4.4 解析 LLM 返回（`internal/service` 报告响应结构）

- 在承接 LLM JSON 的报告结构体上新增 `spouse_palace_match` 对应字段（与 §5 结构同形），透传给前端。
- **不动**前端展示（如需新区块由前端后续按结构渲染，不在本次范围）。

## 5. 输出结构

```json
"spouse_palace_match": {
  "self": {
    "ideal_portrait": "A 命里理想另一半的样子（基于配偶星 + 夫妻宫藏干）",
    "match_level": "high|medium|low",
    "fit_points": ["B 哪里对上了 A 的理想"],
    "gap_points": ["B 哪里差着"],
    "evidence_keys": ["引用支撑证据"]
  },
  "partner": {
    "ideal_portrait": "B 命里理想另一半的样子",
    "match_level": "high|medium|low",
    "fit_points": ["A 哪里对上了 B 的理想"],
    "gap_points": ["A 哪里差着"],
    "evidence_keys": []
  },
  "summary": "一句话总括双向夫妻宫匹配（含缺性别/配偶星不现的说明）"
}
```

## 6. 测试

`pkg/bazi`（表驱动）：
- 男命 → 正确挑出财星（正财/偏财），位置/透藏/是否入夫妻宫判对。
- 女命 → 正确挑出官杀（正官/七杀）。
- 配偶星不现（盘中无财 / 无官杀）→ `Present=false`，`Available=true`，画像料退回夫妻宫藏干。
- 性别缺失 → `Available=false`，安全返回。
- 日支藏干十神正确提取。

`internal/service`：
- 摘要串包含「配偶画像信号」段；缺性别时写明「性别缺失，配偶星不可定」。

回归不变量：
- 触发盘总分、`DimensionScores`、等级与改动前一致（复用现有不变量测试）。

prompt：版本号 +1，过 `canonical_test.go` / `drift_test.go` / `sync_test.go` / 名人配对黄金测试。

## 7. 成功标准

1. `go test ./...` 全绿（已知与本改动无关的 token_usage nil-DB 测试除外）。
2. 跑一对带性别的盘，报告 `spouse_palace_match` 双向出画像 + 高/中/低，且 `match_level` 与夫妻宫合冲状态自洽（相冲不给高）。
3. 缺性别的一方按约定跳过并注明。
4. 总分不变。

## 8. 部署注意

与上次一致：Go 代码改动随发版生效；**prompt 改动需后台手动「采用出厂新版」**（`SyncCanonical` 只补种、不覆盖已存在行）。
