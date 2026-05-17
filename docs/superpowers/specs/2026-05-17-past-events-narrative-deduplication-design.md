# 过往事件推算 — 跨年与卡内信息去重 Spec

**日期：** 2026-05-17
**作者：** Claude + 用户
**状态：** 已批准，待实施

---

## 1. 背景

`/bazi/:chartId/past-events` 页面的年份卡片上存在两类可见重复：

**类型 A：跨年开头模板碰撞。** `RenderYearNarrative` 的前导句来自 `yearToneSentence(signals, primary)`（`backend/pkg/bazi/event_narrative.go:115-149`），它按 `themeOf(primary.Type)` 一对一映射到一个固定字符串。当连续两年主信号同属 `change` 主题时，叙述开头逐字相同。截图中 2025 乙巳年与 2026 丙午年均开头为 "这一年的变动感比较强，旧问题或突发调整容易被推到眼前"。

**类型 B：卡内底部 `年度力量：` 行与主段重复。** 主段在 `tenGodNarrativeSentence`（`event_narrative.go:291-303`）已用 `PlainTitle + PlainText + "，可作为理解这一年事件走向的背景力量。"` 整句融入叙述；前端 `PastEventsPage.tsx:420-429` 又单独渲染 `年度力量：{plain_title} - {plain_text}`，与段中那句逐字重复。

## 2. 目标

- 让相邻年的开头句具备语义差异，差异由 primary 信号的子类型 / Source / Evidence 驱动（信号驱动，而非年份取模）。
- 删除卡内 `年度力量：` 独立行，避免与主段内容重复；保留顶部 chip 标签和段中叙述。
- 不动 polarity 五分支（`xiong>=2 && ji>0` 等），它们是宏观度量而非主题度量，碰撞语义可接受。

## 3. 非目标

- 不改 API 契约：`PastEventsYearItem.TenGodPower` / `Narrative` 等字段保留。
- 不动 `ai_past_events` 缓存：缓存的是 AI 大运总结，不是年 narrative。年 narrative 实时生成，模板修改立即生效。
- 不重写 `RenderYearNarrative` 整体结构，仅替换 `yearToneSentence` 的 themed 分派路径。
- 不动 polarity 五分支文案。
- 不动 `triggerSourceSentence` / `domainDetailSentence` / `secondaryDetailSentence` / `practicalStanceSentence` 等中间段，它们本身已经按多维信号差异化。

## 4. 改动清单

| 文件 | 操作 | 备注 |
|------|------|------|
| `backend/pkg/bazi/event_narrative_leads.go` | **新建** | 容纳 4 个 lead helper：`changeLead` / `healthLead` / `relationshipLead` / `defaultHardLead`。约 80 行。 |
| `backend/pkg/bazi/event_narrative.go` | 改 `yearToneSentence` 中 `isHardEventSignal(primary)` 分支 | 把内部 themed `switch` 改为分派到 4 个 lead helper；polarity 5 分支保持不变。 |
| `backend/pkg/bazi/event_narrative_leads_test.go` | **新建** | 5 个测试覆盖各 lead helper 的差异化分支。 |
| `frontend/src/pages/PastEventsPage.tsx` | 删除 `年度力量：` JSX 块 | 当前位置 420-429（10 行），含 `ten_god_power.plain_title` 条件渲染。 |
| `frontend/tests/past-events-no-ten-god-footer.test.mjs` | **新建** | 静态正则断言：源码不再含 `年度力量：` 字符串。 |

`event_narrative.go` 已有 736 行，超 500 行限制（预存技术债，非本次引入）。新逻辑放到独立文件 `event_narrative_leads.go` 避免进一步恶化。

## 5. 新建文件内容（信号驱动 lead 选择器）

`backend/pkg/bazi/event_narrative_leads.go`：

```go
package bazi

import "strings"

// changeLead 按 primary 信号子类型/极性/Source 选择 change 主题前导句，避免跨年模板碰撞。
func changeLead(p EventSignal) string {
	switch p.Type {
	case "伏吟":
		return "这一年旧事容易卷土重来，过去搁置的处理可能重新冒头"
	case "反吟":
		return "这一年节奏变化会比较突兀，环境、计划或关系都可能临时倒挂"
	case "大运合化":
		return "这一年大运能量被牵动重组，方向感会有一次明显的调整"
	case TypeJuShiZhong:
		return "这一年整体力量容易被放大，一个选择容易牵动多条线"
	}
	if p.Polarity == PolarityXiong {
		if p.Source == SourceXing || strings.Contains(p.Evidence, "刑") {
			return "这一年容易在反复和细节里消耗，问题未必爆发却拖出余波"
		}
		return "这一年的变动感比较强，旧问题或突发调整容易被推到眼前"
	}
	if p.Polarity == PolarityJi {
		return "这一年变动中带着调整空间，主动顺势比被动应对更省力"
	}
	return "这一年节奏不算稳定，但调整中容易找到新方向"
}

// healthLead 按 Evidence/Polarity 选择 health 主题前导句。
func healthLead(p EventSignal) string {
	if strings.Contains(p.Evidence, "冲") || strings.Contains(p.Evidence, "白虎") {
		return "这一年身体和安全节奏需要被前置考虑，意外性消耗要避免"
	}
	if p.Polarity == PolarityXiong {
		return "健康、安全或日常节奏是这一年的主线，压力点会比较直接"
	}
	return "这一年身心提醒会更频繁，作息节律值得重新校准"
}

// relationshipLead 按 primary.Type 选择 relationship 主题前导句。
func relationshipLead(p EventSignal) string {
	switch p.Type {
	case "婚恋_合":
		return "这一年人际或感情的靠近感增强，关系节奏容易加快"
	case "婚恋_冲":
		return "这一年关系、距离和承诺容易被检验，节奏可能出现明显波动"
	case "婚恋_变":
		return "这一年情感或合作的方向容易调整，分寸感和边界都会被试探"
	case TypeXingGeQingYi:
		return "这一年情绪表达和人际反应会更外露，主动沟通比闷着推进有效"
	case TypeXingGePanNi:
		return "这一年个性主张容易和外部要求碰撞，关键节点上要稳住态度"
	}
	return "人际、感情或家庭沟通是这一年的主线，情绪触发会比较明显"
}

// defaultHardLead 按 primary.Source 选择硬事件兜底前导句。
func defaultHardLead(p EventSignal) string {
	switch p.Source {
	case SourceKongwang:
		return "这一年带着虚而不实的不稳定感，承诺和计划要多确认细节"
	case SourceXing:
		return "这一年有内耗反复的影子，事情未必爆发但容易消耗心力"
	case SourceFuyin:
		return "这一年旧主题反复回头，过去没处理完的事情会再被推上来"
	case SourceHehua:
		return "这一年大运能量被牵动，方向上的关键节点会比预想更明显"
	}
	return "这一年不是完全平稳的年份，关键事件会比平时更容易显形"
}
```

## 6. `yearToneSentence` 修改差异

`backend/pkg/bazi/event_narrative.go:115-135` 中 `if isHardEventSignal(primary)` 内部由原 `switch` 改为：

```go
if isHardEventSignal(primary) {
	switch themeOf(primary.Type) {
	case "health":
		return healthLead(primary)
	case "change":
		return changeLead(primary)
	case "relationship":
		return relationshipLead(primary)
	default:
		return defaultHardLead(primary)
	}
}
```

polarity 五分支（`xiong>=2 && ji>0` 等，原 137-148）保持不变。

## 7. 前端改动

`frontend/src/pages/PastEventsPage.tsx`：删除整个 `{y.ten_god_power?.plain_title && (...)}` 块（420-429 行 10 行）。`ten_god_power` 字段在数据层保留。

## 8. 测试

### 后端 `event_narrative_leads_test.go`

5 个 `Test_*` 函数：

- `Test_changeLead_DistinctBranches`：构造 5 个 EventSignal（Type=伏吟/反吟/大运合化/TypeJuShiZhong/综合变动+Xing 源），调用 `changeLead`，断言 5 返回字符串两两不等。
- `Test_changeLead_LegacyFallbackReachable`：Type=综合变动，Polarity=Xiong，Source 非 Xing，Evidence 不含 "刑" → 应返回 "这一年的变动感比较强，旧问题或突发调整容易被推到眼前"（回归保护）。
- `Test_healthLead_3Branches`：分别构造 Evidence 含 "冲"、Polarity Xiong（不含 "冲/白虎"）、Polarity 非 Xiong 的三种输入，断言三返回字符串两两不等。
- `Test_relationshipLead_DistinctBranches`：构造 5 个 EventSignal（Type=婚恋_合/婚恋_冲/婚恋_变/TypeXingGeQingYi/TypeXingGePanNi），断言 5 返回字符串两两不等。
- `Test_defaultHardLead_4Sources`：构造 4 个 EventSignal（Source=Kongwang/Xing/Fuyin/Hehua），断言 4 返回字符串两两不等。

### 前端 `frontend/tests/past-events-no-ten-god-footer.test.mjs`

`node --test` 静态正则断言：

- `PastEventsPage.tsx` 源码不再包含 `年度力量：` 字符串。

## 9. 验收

1. 点开过往事件推算页面，相邻两年若主信号同属 `change` 主题但 Type/Source/Evidence 不同（如 2025 综合变动+Xing、2026 局势中），开头句**不再逐字相同**。
2. 卡片底部不再出现 `年度力量：…` 独立行；顶部 chip 标签和段中"这股年度力量会把…" 仍存在。
3. 后端 `go test ./pkg/bazi/...` 全绿。
4. 前端 `node --test tests/past-events-no-ten-god-footer.test.mjs` 全绿；`npm run build` 与 `npm run lint` 通过。

## 10. 风险与回退

- **风险低**：纯算法/UI 删除，无数据契约变更，无 DDL。
- **回退**：单 PR revert 即可。
- **可能漏点**：如果未来有新增 primary.Type 进入 `change/health/relationship` 主题但未在 lead helper switch 中列出，会落到 lead 的兜底分支 — 兜底分支返回的是当前的旧字符串，行为不退化。

## 11. 实施分支

从 main 起 `feat/past-events-dedup-narrative`，单 PR 完成全部 5 个文件改动（含新建测试）。
