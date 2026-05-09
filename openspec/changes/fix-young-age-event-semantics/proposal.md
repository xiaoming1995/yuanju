## Why

当前过往事件推算对所有年龄一视同仁：流年财星透干 → "财运提升"，流年官杀透干 → "事业晋升"，财官透干 → "婚恋进展"。但起运到 18 岁的命主仍处于读书阶段，"财运/事业/婚恋"语义与命主真实生活不符——零用钱、考试压力、同学情谊才是这个年龄段的关键事件。专业人士反馈：少年期叙述应改用学业 + 性格语义，让用户读到自己 14 岁年份不会看到"事业晋升"这种荒唐文案。

## What Changes

- **新增** 流年信号引擎在年龄 < 18 时启用读书期语义分支：
  - 财星透干 → `学业_资源`（家境/零用钱/物质条件）
  - 比劫透干 → `学业_竞争`（同学竞争/兄弟情谊）
  - 官杀透干 → `学业_压力`（考试/升学/老师管教）
  - 印星透干 → `学业_贵人`（导师/学习方法/师承）
  - 食伤透干 → `学业_才艺`（兴趣特长/才艺展现）
  - 桃花/红艳/天喜 → `性格_情谊`（友情/初恋萌动/同窗情谊）
  - 流年合日支 / 大运合日支 → `性格_情谊`
  - 流年/大运冲日支 → `性格_叛逆`（情绪波动/家庭关系）
  - 财官双叠（成人期映射至婚恋）→ 不输出（少年期不适用）
- **新增** `GetYearEventSignals` 增加 `age int` 参数；`GetAllYearSignals` 在循环中传递 `ln.Age`
- **新增** 7 个 Type 常量与前端 `SIGNAL_LABEL` 标签映射（学业_资源/学业_竞争/学业_压力/学业_贵人/学业_才艺/性格_情谊/性格_叛逆）
- **新增** `GenerateDayunSummariesStream` 计算每段大运的 `youngRatio`（年龄 < 18 的年份占比）：
  - youngRatio = 1.0 → AI prompt 加"本段为读书期，请聚焦学业、性格、同窗关系"
  - 0 < youngRatio < 1.0 → 加"本段跨越读书期与成人期，请分两段叙述"
  - youngRatio = 0 → 现有模板不变
- **保留** 健康 / 迁变 / 伏吟反吟 / 大运合化 / 喜神临运 / 用神基底 / 空亡 / 三刑 / 三会 等不分年龄
- **不变** 用神 / 调候 / 身强弱 / 大运计算逻辑均不动
- **不变** `bazi_charts.result_json` schema 不动；age 字段已存在于 `Dayun[].LiuNian[].Age`

## Capabilities

### New Capabilities

- `young-age-event-semantics`：定义 18 岁前流年信号的读书期语义映射规则、Type 命名约定、年龄边界、大运总结 prompt 动态调整规则

### Modified Capabilities

无（流年信号引擎尚未独立 spec；本次新建 capability 即可承载所有规则）

## Impact

**后端（Go）**

- `pkg/bazi/event_signals.go` —
  - `GetYearEventSignals` 函数签名增加 `age int` 参数
  - 加入 `age < youngAgeCutoff` 分支，重写财星/官杀/比劫/食伤/印星/桃花/合冲日支等评估
  - 新增常量 `YoungAgeCutoff = 18` 集中定义边界
  - 新增 7 个 Type 常量：`TypeXueYeZiYuan`、`TypeXueYeJingZheng`、`TypeXueYeYaLi`、`TypeXueYeGuiRen`、`TypeXueYeCaiYi`、`TypeXingGeQingYi`、`TypeXingGePanNi`
- `pkg/bazi/event_signals.go::GetAllYearSignals` — 在外层循环中将 `ln.Age` 传给 `GetYearEventSignals`
- `pkg/bazi/event_narrative.go::ExtractYearSignalTypes` — 新 Type 通过现有 hide map 过滤即可，无需特殊处理
- `internal/service/report_service.go::GenerateDayunSummariesStream` —
  - 新增 `youngRatio` 计算（遍历 dy.LiuNian 计 age < 18 的占比）
  - prompt 模板按 youngRatio 三档插入语境提示

**前端（React）**

- `frontend/src/pages/PastEventsPage.tsx::SIGNAL_LABEL` — 加 7 个新 Type 映射，颜色用：
  - 学业系：`var(--wu-mu)` 偏绿（成长）
  - 性格系：`var(--wu-tu)` 偏黄（人际）

**测试**

- `pkg/bazi/young_age_test.go`（新建）覆盖：
  - 14 岁 + 财星透干 → Type=`学业_资源`
  - 19 岁 + 财星透干 → Type=`财运_得`（成人期不变）
  - 14 岁 + 流年合日支 → Type=`性格_情谊`
  - 14 岁 + 桃花神煞 → 输出 `性格_情谊`
  - 跨界大运（年龄 14-23）的 youngRatio = 0.5

**风险**

- 起运 < 8 岁的命盘有更多年份落在读书期（10-15 个流年），AI 报告需要额外文字篇幅区分两期；若 AI 输出超长会被前端截断（已有 80-120 字软上限）
- 个别命主 16 岁辍学打工 / 22 岁仍在读博士 → 18 岁硬阈值不能 100% 准确，evidence 文案保留可解释性以缓冲
- 旧 `ai_dayun_summaries` 缓存不会因本次变更失效；需告知用户清理或等下次重生成
