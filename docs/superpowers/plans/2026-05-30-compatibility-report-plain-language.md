# 合盘报告「说人话」润色 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把合盘报告里用户能看到的术语文案（证据卡 + LLM 深度解读正文）改写成温和顾问口吻、四五句密度的人话，术语保留做锚点 + 补人话解释。

**Architecture:** 两层独立改动——① LLM 正文：在 `canonical_compatibility.go` prompt 加一段「表达约束」并 bump 版本号；② 确定性文案：重写 `compatibility_evidence.go` 里 10 条用户可见的证据卡 `Detail`，保留 `Title`/`Type`/`EvidenceKey`/`Weight` 等结构字段。不改 score-explanation summary（用户不可见，仅喂 LLM）、不改 summary tags、不动前端、不加字段、零迁移。

**Tech Stack:** Go（backend/pkg/bazi、backend/pkg/prompt），`go test` / `go vet` / `gofmt`。

> **提交说明**：用户立有「只在我让你提交时才提交」的规矩。计划中的 `git commit` 步骤是建议粒度；实际提交等用户发话。

---

## File Structure

- `backend/pkg/prompt/canonical_compatibility.go` — 改：新增「表达约束」段 + 版本号 `-2` → `-3`。
- `backend/pkg/prompt/canonical_test.go` — 改：版本断言 `-3`，required 子串加入 `表达约束`。
- `backend/pkg/bazi/compatibility_evidence.go` — 改：重写 10 条证据卡 `Detail`（zodiac×4 / nayin×2 / day_pillar×3 / eight_chars×1）。
- 不新建任何文件。

---

## Task 1: LLM 正文 — prompt 加「表达约束」+ bump 版本（TDD）

**Files:**
- Modify: `backend/pkg/prompt/canonical_test.go:13-14`、`:20`
- Modify: `backend/pkg/prompt/canonical_compatibility.go:4`（Version）、插入「表达约束」段于 `性格画像约束` 之后、`输出严格为 JSON：` 之前

- [ ] **Step 1: 先改测试，让它失败**

把 `canonical_test.go` 的版本断言改成 `-3`，并在 required 子串列表里加上 `表达约束`：

```go
	if def.Version != "v3.1-question-aware-3" {
		t.Errorf("expected Version v3.1-question-aware-3, got %q", def.Version)
	}
```

```go
	for _, want := range []string{"question_focus", "decision_advice", "personality_comparison", "表达约束", "{{.PrimaryQuestionLabel}}"} {
```

- [ ] **Step 2: 运行测试，确认失败**

Run: `cd backend && go test ./pkg/prompt/ -run TestMustGet_CompatibilityReturnsRegisteredDefinition -v`
Expected: FAIL —— 版本仍为 `v3.1-question-aware-2`，且 content 不含 `表达约束`。

- [ ] **Step 3: 改 prompt —— bump 版本号**

`canonical_compatibility.go` 第 5 行：

```go
		Version:     "v3.1-question-aware-3",
```

- [ ] **Step 4: 改 prompt —— 插入「表达约束」段**

在 `性格画像约束（personality_comparison）：` 整段（到 `- fit_points / clash_points ...` 那一行）之后、空行 + `输出严格为 JSON：` 之前，插入：

```
表达约束（面向普通用户，务必遵守）：
- 全程用温和顾问口吻，像一个既懂行又体贴的人在跟当事人解释，不端着、不冷冰冰、不堆术语。
- 任何八字术语（六合、三合、纳音、日柱、十神、旺衰…）首次出现时，必须紧跟一句大白话解释它意味着什么；严禁整句只有术语而没有解释。
- 把判断说透、不要惜字：除了给结论，也要讲清「为什么」以及「落到两个人相处上具体是什么样」。
- summary / question_focus / relationship_diagnosis / personality_comparison / relationship_strategy / advice 等所有面向用户的字段，一律用日常语言，说法落到「你们 / 对方 / 相处」这种当事人能直接对号入座的词。

```

> 注意：只新增「输出语言」约束，**不改**现有给 LLM 推理用的术语规则段（评分规则说明 / 性格画像约束本身）——LLM 需要精确术语来推理。

- [ ] **Step 5: 运行测试，确认通过**

Run: `cd backend && go test ./pkg/prompt/ -v`
Expected: PASS（含 `TestMustGet_CompatibilityReturnsRegisteredDefinition`、`TestCompatibilityPromptUsesV3ModuleKeys`）。

- [ ] **Step 6: gofmt + vet**

Run: `cd backend && gofmt -w pkg/prompt/canonical_compatibility.go pkg/prompt/canonical_test.go && go vet ./pkg/prompt/`
Expected: 无输出。

- [ ] **Step 7: 提交（等用户发话）**

```bash
git add backend/pkg/prompt/canonical_compatibility.go backend/pkg/prompt/canonical_test.go
git commit -m "$(cat <<'EOF'
feat(compat-prompt): add plain-language output constraint to compatibility report

LLM 深度解读正文：温和顾问口吻、术语必带人话解释、把话说透。
版本 v3.1-question-aware-2 → -3。

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: 确定性文案 — 重写 10 条证据卡 Detail

**Files:**
- Modify: `backend/pkg/bazi/compatibility_evidence.go`（`zodiacEvidence` / `nayinEvidence` / `dayPillarEvidence` / `eightCharsEvidence` 的 `Detail` 字段）

> 该文件无任何测试断言这些 `Detail` 字符串；结构不变量（`EvidenceKey`/`Weight`/`Dimension`/`Polarity`/条数）由现有 `pkg/bazi/compatibility_test.go` 守护。改写时**只动 `Detail`**，其余字段（含 `Title`/`Type`）保持原样。

- [ ] **Step 1: 重写 zodiacEvidence 的 4 条 Detail**

`zodiac_liuhe`：

```go
			Detail:      fmt.Sprintf("你俩的属相（年支 %s / %s）构成了「六合」——这是属相搭配里最讨喜、最顺的一种。落到相处上，就是你们见面容易互相来电、自来熟，很多事不用刻意经营就能对上眼。这种天生的亲近感，会让你们在磨合期少很多无谓的摩擦。不过它管的是「合不合得来」，关系能走多远，长久还得看两个人愿不愿意一起用心经营。", a.YearZhi, b.YearZhi),
```

`zodiac_sanhe`（保留 `group`）：

```go
			Detail:      fmt.Sprintf("你俩的属相（年支 %s / %s）同属「%s 三合」——这是一种气场很合拍的属相组合。相处时你们更容易步调一致、想到一块去，遇事也常本能地站在同一边。比起针锋相对，你俩更像天然的同盟，这对关系的稳定是实打实的加分。当然，合得来不等于不用沟通，重要的事还是要摊开说清楚。", a.YearZhi, b.YearZhi, group),
```

`zodiac_same_element`（保留 `wx`）：

```go
			Detail:      fmt.Sprintf("你俩的属相（年支 %s / %s）五行都属%s，命理里叫「双生」——本质上是同一类能量。这意味着你们性子和节奏比较像，容易理解彼此，有种「同类」的天然亲切感。相处起来不太需要费力解释自己，对方往往一点就通。要留意的是，太像有时也会少了点互补，碰到同一类短板时容易一起卡住。", a.YearZhi, b.YearZhi, wx),
```

`zodiac_sheng`（保留 `wxA`、`wxB`）：

```go
			Detail:      fmt.Sprintf("你俩的属相（年支 %s / %s）构成「%s 生 %s」的五行相生——一方天然能滋养、托举另一方。相处里常表现为一个愿意付出、一个被照顾，关系有种顺其自然的承接感。这种「你帮我、我托你」的流动，是长期相处很舒服的底子。只要别让付出长期单方向倾斜，这份相生就能一直顺下去。", a.YearZhi, b.YearZhi, wxA, wxB),
```

- [ ] **Step 2: 重写 nayinEvidence 的 2 条 Detail**

`nayin_sheng`（保留 `wxA`、`wxB`）：

```go
			Detail:      fmt.Sprintf("你俩的「纳音」五行是 %s 与 %s 相生——纳音说的是两个人骨子里的底色气质。相生意味着这两种底色能互相滋养，相处时情绪和资源都流动得比较顺，不太会互相消耗。日子久了你们会发现，跟对方在一起更像「回血」而不是「耗电」。这是一段关系里很难得、也很值钱的底层契合。", wxA, wxB),
```

`nayin_same`（保留 `wxA`）：

```go
			Detail:      fmt.Sprintf("你俩的「纳音」五行同为 %s——纳音是两个人骨子里的底色气质，同气说明你们本质上是一类人。这种同频会让你们天然懂彼此的在意和顾虑，很多感受不用说出口对方也能体会，默契感比一般人强不少。唯一要提醒的是，太同步时也容易一起钻牛角尖，偶尔需要有一个人先跳出来踩刹车。", wxA),
```

- [ ] **Step 3: 重写 dayPillarEvidence 的 3 条 Detail**

`day_pillar_upper`：

```go
				Detail: fmt.Sprintf(
					"日柱（%s%s / %s%s）是命盘里最贴近婚恋、最代表「枕边人」的一根柱子，你俩在这里咬合得很到位——地支相合、天干也彼此呼应。这说明在亲密关系的核心地带，你们有天然的契合和稳定结构。相处中容易有那种「找对了人」的踏实感，亲密和信任都建立得比较顺。这是合盘里分量很重的一个好信号，值得珍惜。",
					a.DayGan, a.DayZhi, b.DayGan, b.DayZhi,
				),
```

`day_pillar_lower`：

```go
				Detail: fmt.Sprintf(
					"日柱（%s%s / %s%s）代表两个人最贴近婚恋的核心，你俩的地支在这里是相合的——亲密关系有不错的底子。只是天干层面没能再进一步互相加成（只是相同、相克或不相干），所以这份契合算「够好」但还没到顶配。日常相处大方向是合的，偶尔在细节和默契上需要多一点磨合。把沟通做扎实，这段亲密就能稳稳地往上走。",
					a.DayGan, a.DayZhi, b.DayGan, b.DayZhi,
				),
```

`day_pillar_safe`：

```go
				Detail: fmt.Sprintf(
					"日柱（%s%s / %s%s）是两个人最贴近婚恋的核心。你俩的日支虽然没有直接相合，但五行上是相同或相生的，所以亲密层还是留了一丝天然的亲近感。这说明你们不是格格不入，底子里有可以亲近的余地，只是要比「天生一对」那种多花点心思去经营。把相处的节奏和沟通磨顺，这点微弱的亲近感是能养大的，别因为起步平淡就轻易否定它。",
					a.DayGan, a.DayZhi, b.DayGan, b.DayZhi,
				),
```

- [ ] **Step 4: 重写 eightCharsEvidence 的 Detail 模板**

去掉技术表述「命中…（贡献 N）」。`Detail` 改为（参数从 6 个减为 5 个——删掉 `t.label` 与 `s`，避免把内部计分档名「安慰分」直接抛给用户；档位强弱仍由 `Title`/`Weight` 承载）：

```go
			Detail: fmt.Sprintf(
				"%s（%s%s / %s%s）也合上了——这是你们在生活外围（家世背景、日常相处、未来安排这些围绕婚恋的方面）的一处天然契合。它不像日柱那样直接管亲密核心，但能在周边给关系搭把手，让你们在现实层面更容易对得上。这类外围的合拍越多，往后一起过日子越省心。",
				p.label, p.ganA, p.zhiA, p.ganB, p.zhiB,
			),
```

> 校验：`t`（`t.label`）仍用于上方 `Type`/`Title`，`s` 仍用于 `Weight: s` 与 `tierByScore[s]` 查表——两者均未变成孤儿变量，无需删除。

- [ ] **Step 5: gofmt + vet + 跑结构守护测试**

Run: `cd backend && gofmt -w pkg/bazi/compatibility_evidence.go && go vet ./pkg/bazi/ && go test ./pkg/bazi/ -run 'TestBuildEvidences|TestBuildScoreExplanationsV3|TestBuildSummaryTagsV3' -v`
Expected: 编译通过、`go vet` 无输出、上述测试全 PASS（证明 `EvidenceKey`/`Weight`/`Dimension`/条数/Summary 非空等结构不变量未被破坏）。

- [ ] **Step 6: 全包回归**

Run: `cd backend && go test ./pkg/bazi/ ./pkg/prompt/`
Expected: PASS。

- [ ] **Step 7: 提交（等用户发话）**

```bash
git add backend/pkg/bazi/compatibility_evidence.go
git commit -m "$(cat <<'EOF'
feat(compat-evidence): rewrite evidence card copy into plain warm-advisor language

证据卡 Detail（年支六合/三合/同行/相生、纳音相生/同气、日柱上/次/安慰、八字外围）
改为术语锚点 + 人话解释、温和顾问口吻、四五句密度。
保留 Title/Type/EvidenceKey/Weight；不改 score-explanation summary 与 summary tags。

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: 运行时验证（人工，需 :5200 + LLM key）

**Files:** 无（仅观察）

- [ ] **Step 1: 起前端（若未起）**

Vite 跑在 :5200（HMR）。后端 :9002。

- [ ] **Step 2: 新建一个合盘**

> 关键：确定性文案存库，改模板只对**新建**的合盘生效；旧报告保留旧文案。必须走一次「新建合盘」流程，不能开旧记录。

优先填一对能命中多模块的生辰（年支六合 / 纳音相生 / 日柱合），让证据卡尽量多出几条。

- [ ] **Step 3: 看证据组卡片**

展开 EvidenceDrawer：每条证据标题仍是术语（年支六合…），正文是四五句温和顾问口吻的人话、读得懂、不堆术语、无 `%!`(格式错误) 字样。

- [ ] **Step 4: 生成深度解读**

点「生成深度解读」→ 等 LLM 返回。确认 summary / 判断 / 性格画像 / 策略建议 正文是温和顾问口吻、术语都带人话解释、把话说透。

- [ ] **Step 5: 回报结果**

把渲染结果（截图或文字）回报；若文案密度/口吻还想调，回到 Task 1/2 改 copy。

---

## Self-Review

- **Spec coverage**：① LLM 正文 → Task 1；② 证据卡 detail（10 条）→ Task 2；不改 summary/tags/前端/字段 → 已在 plan 与 design 明确排除；存储模型导致「仅新报告生效」→ Task 3 Step 2 已强调。✅
- **Placeholder scan**：每个 code step 均给出完整 `fmt.Sprintf` 实参；无 TODO/TBD/「类似上文」。✅
- **Type consistency**：Task 2 各条仅改 `Detail` 字面量与其 `fmt` 实参，未触碰 `EvidenceKey`/`Weight`/`Dimension`/`Title`/`Type`；eight_chars 删参后 `t`/`s` 仍被引用（已在 Step 4 校验注明）。Task 1 版本串 `v3.1-question-aware-3` 在 prompt 与 test 两处一致。✅
