## Context

**当前状态**

`backend/pkg/bazi/event_signals.go::GetYearEventSignals(natal, lnGan, lnZhi, dayunGanZhi, gender)` 是流年信号检测核心，按 12 个维度独立判断，输出 `[]EventSignal`。每个 Signal 含 Type/Evidence/Polarity/Source。Type 命名约定为 `大类_小类`（如 `财运_得`、`婚恋_合`），目前共 13 种类型。

外部调用：

```
GetAllYearSignals(result, gender, currentYear, minAge):
  for dy in result.Dayun:
    for ln in dy.LiuNian:
      sigs = GetYearEventSignals(natal, lnGan, lnZhi, dyGanZhi, gender)
      out = append(out, YearSignals{Year, Age, GanZhi, DayunGanZhi, Signals: sigs})
```

**约束**

- 不破坏现有 13 种 Type 的下游消费（前端 `SIGNAL_LABEL`、AI prompt 模板）
- yongshen / 调候 / 身强弱 不动
- `BaziResult.Dayun[].LiuNian[].Age` 已含每年虚岁，无需新增字段
- evidence 字符串需保持可解释（避免类似"少年期金气克木"这种过于阴阳化、对学生无感的表述）
- 单文件不超 500 行（CLAUDE.md 约束；event_signals.go 当前 1196 行已超，但本次仅在内增加分支不外扩）

## Goals / Non-Goals

**Goals**

1. 起运 → 18 岁的流年改用学业 / 性格语义，覆盖 7 大十神事件（财/官/印/食伤/比劫/桃花/合冲日支）
2. Type 命名遵循 `大类_小类`，前端可通过 `SIGNAL_LABEL` 渲染独立 chip
3. 大运 AI 总结 prompt 按读书期年份占比动态调整
4. 现有成人期判定不受影响（19 岁以上行为完全保持）
5. 测试覆盖少年期、成人期、跨界期三类场景

**Non-Goals**

- 命主真实学龄修正（如 16 岁辍学）— 用 18 岁硬阈值
- 实际算命术语精度提升（"学业_压力"是否真切对应少年命局）— 第一阶段足够，后续可依 admin 反馈微调
- 调候 / 用神 / 身强弱 在少年期的特殊解读
- 现有 ai_dayun_summaries 缓存清理 — 用户自行重生成
- 18-22 岁大学期段的特殊语义 — 暂归成人期处理
- format admin 调试 UI 暴露 youngRatio

## Decisions

### D1：硬阈值 18，不开放配置

**Why**：第一阶段简化。99% 的高中毕业边界是 18 岁，配置化反而引入认知负担与运营压力。后续若有反馈，可加 algo_config 的 `young_age_cutoff` 参数升级。

**Alternatives**

- 开放 admin 可配 → 增加 algo_config 表项 + admin UI + 缓存刷新逻辑，本轮溢出
- 起运 → 22 岁覆盖大学 → 大学生未读书占比已不一致（实习 / 创业 / 谈恋爱），统计上不胜过 18 岁
- 按 chart_id 自定义 → 用户输入门槛高，UI 体验下降

### D2：Type 命名 `学业_资源` / `性格_情谊` 等

**Why**：与现有 `财运_得`/`婚恋_合` 风格一致；下划线分隔便于前端 SIGNAL_LABEL 字典精确命中；中文便于运营理解。

**完整 Type 列表**：

| Type 常量名 | Type 字符串 | Polarity 倾向 | 触发条件 |
|---|---|---|---|
| `TypeXueYeZiYuan` | `学业_资源` | 吉（财星非忌）／ 凶（财星忌） | 流年财星透干 |
| `TypeXueYeJingZheng` | `学业_竞争` | 中性 / 凶 | 流年比劫透干 |
| `TypeXueYeYaLi` | `学业_压力` | 凶 / 吉（印化杀） | 流年官杀透干 |
| `TypeXueYeGuiRen` | `学业_贵人` | 吉 | 流年印星透干 |
| `TypeXueYeCaiYi` | `学业_才艺` | 吉（身强）/ 凶（身弱） | 流年食伤透干 |
| `TypeXingGeQingYi` | `性格_情谊` | 吉 / 中性 | 流年地支六合日支 / 大运合日支 / 桃花神煞 |
| `TypeXingGePanNi` | `性格_叛逆` | 凶 | 流年冲日支 / 大运冲日支 |

### D3：调用链改动 — `GetYearEventSignals` 增加 `age int` 参数

**Why**：年龄是流年级别的属性，不能从 natal 推。最干净是签名扩展。

**Alternatives**

- 全局 thread-local — Go 没有；用 context.Context 传播，但 GetYearEventSignals 是纯函数
- 在 EventSignal 上加字段后处理 — Type 决策耦合到检测逻辑里，更清晰

**实施**：

```go
// Before
func GetYearEventSignals(natal *BaziResult, lnGan, lnZhi, dayunGanZhi, gender string) []EventSignal

// After
func GetYearEventSignals(natal *BaziResult, lnGan, lnZhi, dayunGanZhi, gender string, age int) []EventSignal
```

`GetAllYearSignals` 内的循环已有 `ln.Age` 直接可用。其他调用方（如有测试）需要更新签名。

### D4：少年期分支放在 `GetYearEventSignals` 内部、与成人期 if/else 互斥

**Why**：避免少年期与成人期同时输出"学业_资源"+"财运_得"造成重复。在每个事件检测点判断 `if age < YoungAgeCutoff { ... } else { ... 现有逻辑 ... }`。

**伪代码**：

```go
const YoungAgeCutoff = 18
isYoung := age > 0 && age < YoungAgeCutoff

if isFinanceStar {
  if isYoung {
    addP("学业_资源", evidenceForYoung(...), polarity, SourceZhuwei)
  } else {
    addP("财运_得", evidenceForAdult(...), polarity, SourceZhuwei)
  }
}
```

### D5：少年期 evidence 文案规则

每个少年期分支的 evidence 应：
1. 不出现"财运/事业/婚恋"等成人词
2. 用具象学业语言（考试、补习、零用钱、特长、师长、同窗、初恋萌动）
3. 保留十神逻辑（如食伤洩气在身弱时仍是凶，但写成"才艺学习易过劳/分心"）
4. 保留极性中文标识（吉/凶/中性内嵌于句末或修饰）

**示例对照**（财星透干，身强）：

| 阶段 | Type | Evidence |
|---|---|---|
| 成人期 | `财运_得` | "X 透干为偏财，财星透出，财运有望提升，宜主动把握进财机会" |
| 少年期 | `学业_资源` | "X 透干为偏财，少年财星透露家境/零用钱有改善信号，宜规划学习投入" |

### D6：大运总结 youngRatio 计算

```go
youngCount := 0
totalCount := 0
for _, ln := range dy.LiuNian {
  if ln.Age < 1 { continue }  // 起运前不计
  totalCount++
  if ln.Age < 18 { youngCount++ }
}
youngRatio := float64(youngCount) / float64(totalCount)
```

prompt 模板按三档：

```
youngRatio == 1.0  → "本段大运全部年份处于读书期，请以学业、性格塑造、同窗关系为主轴撰写"
0 < youngRatio < 1.0 → "本段大运跨越读书期与成人期（前 N 年读书、后 M 年入社会），请分两段叙述"
youngRatio == 0    → 不附加（现有模板）
```

由 `GenerateDayunSummariesStream` 在循环每段大运时计算并 inject 到 `DayunSummaryTemplateData` 的新字段（如 `LifeStageHint string`）。

### D7：前端 SIGNAL_LABEL 颜色分配

学业系（成长向）：`var(--wu-mu)` 木绿
- 学业_资源 → "学业↑" 木绿
- 学业_贵人 → "贵人" 木绿
- 学业_才艺 → "才艺" 木绿
- 学业_竞争 → "竞争" 灰中性
- 学业_压力 → "压力↓" 红凶

性格系（人际向）：`var(--wu-tu)` 土黄
- 性格_情谊 → "情谊" 土黄
- 性格_叛逆 → "叛逆" 红凶

颜色与现有 `财运_得` (var(--wu-jin) 金) / `婚恋_合` (var(--wu-huo) 火) 区分清晰。

## Risks / Trade-offs

| 风险 | 缓解 |
|---|---|
| 18 岁硬阈值在 16 辍学 / 22 仍读博 命盘上失准 | 文案保留可解释性，命主可自行映射；后续可加 admin 配置开关 |
| 少年期 evidence 文案需大量手写、维护成本高 | 在 D5 给出统一规则，所有 evidence 走相同模式（"少年X透露Y信号，宜Z"），未来一处改全部生效 |
| AI prompt 跨界提示词可能让 summary 写散（30 字读书期 + 50 字成人期）| youngRatio 三档清晰；跨界例子的提示词强调"分两段，每段约 50 字"，由 token 限制自然约束 |
| 旧 `ai_dayun_summaries` 缓存仍带成人期文案 | 文档说明"算法升级后已存缓存不会自动刷新；如需更新请清理 ai_dayun_summaries 表对应行" |
| `event_signals.go` 已 1196 行，加分支后破 1300 | 拆分时机：本次先内嵌；后续若再加复杂语义则单独 yongAgeSignals.go 拆出 |
| 测试命盘需要起运早（< 8 岁）以验证 14 岁少年期；测试 fixture 选择需谨慎 | 用已有 1989-03-20（己卯日，8 岁起运）作为测试种子，14 岁年份已落在该范围 |

## Migration Plan

1. 后端：`go build ./...` + `go test ./pkg/bazi/...`
2. 前端：`npm run build` 后产出新 chunk
3. `docker compose up -d --build backend frontend`
4. 验证：手工触发 `POST /api/bazi/past-events/years/:chart_id` 看 14 岁年份的 signals 字段是否含 `学业_*` 与 `性格_*`
5. 验证：触发 `POST /api/bazi/past-events/dayun-summary-stream/:chart_id` 看读书期大运的 summary 文案
6. （可选）admin 清理 `ai_dayun_summaries` 表中老命盘缓存以让 AI 重生成

**Rollback**：

- `git revert` 后重新部署即可
- 已存 ai_dayun_summaries 缓存继续可用（prompt 字段缺失也不影响读取）

## Open Questions

- 18 岁硬阈值是否在 0.5 个月后开放 admin 配置？（暂不实施，等线上反馈）
- `学业_资源` 在身弱财多场景的 evidence 用语是否需要"父母经济压力大"这种引申？（暂不引申，保持事实陈述）
- 大运总结 youngRatio 计算时，是否要排除"大运第 1 年（起运年）"以避免边界年份扰动？（暂不排除，简化逻辑）
