# 命理解读·润色版 个性化 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在原版 AI 命理解读基础上加 tab 切换 + 「润色版」 —— 用户输入 20-300 字当前情况后，AI 基于原版 5 章 + 用户输入逐章重写一份贴近用户处境的版本。

**Architecture:** 新增 `ai_polished_reports` 表（UNIQUE chart_id，UPSERT 覆盖）。后端走 handler → service → LLM (复用 callAIWithSystem) → ParseMarkdownToStructured → repository UPSERT。前端 ResultPage 加 tab，新建 PolishedPanel 组件处理空态/输入态/展示态三种 UI，复用现有 chapter 渲染逻辑。

**Tech Stack:** Backend Go + Gin + lib/pq；Frontend React 19 + TypeScript + Vite，测试用 Node 内置 `node:test`（参考 `frontend/tests/dayun-timeline-ux.test.mjs`）和 Go test。

---

## 设计澄清

**Spec**: `docs/superpowers/specs/2026-05-17-polished-report-personalization-design.md`
**触发前提**：原版 AI 报告必须先存在（已存在的功能，无需新做）。
**LLM 调用方式**：非流式，单次调用。
**输入校验**：20-300 字，client + server 双层。
**重生策略**：用户改输入 → 覆盖式 UPSERT（一个 chart 一份润色版）。

---

## 文件结构（locked in）

```
新增（后端）
  backend/internal/model/polished_report.go              PolishedReport struct + DB column 映射
  backend/internal/repository/polished_report_repo.go   UpsertPolishedReport + GetPolishedByChartID
  backend/internal/service/polished_report_service.go   PolishReport orchestration + buildPolishPrompt + 测试 helpers
  backend/internal/service/polished_report_service_test.go 单元测试（prompt 构造 / 输入校验）
  backend/internal/handler/polished_report_handler.go   GenerateAndSavePolishedReport + GetPolishedReportHandler

新增（前端）
  frontend/src/components/PolishedPanel.tsx              三态：空 / 输入 / 展示，复用章节渲染逻辑
  frontend/src/components/PolishedPanel.css              tab + panel 样式
  frontend/tests/polished-report-ui.test.mjs             源码级断言

修改（后端）
  backend/pkg/database/database.go                       追加 ai_polished_reports CREATE TABLE
  backend/cmd/api/main.go                                注册 2 个新 route

修改（前端）
  frontend/src/lib/api.ts                                getPolishedReport / generatePolishedReport API 方法 + 类型
  frontend/src/pages/ResultPage.tsx                      加 reportTab state + tab UI + PolishedPanel 装配
  frontend/src/pages/ResultPage.css                      tab 样式
```

---

## Task 0: 前置验证 + 起新分支

**Files:**
- Verify: 当前在 main 分支，工作树干净
- Branch: 新建 `feat/polished-report`

- [ ] **Step 1: 检查当前状态**

```bash
cd /Users/liujiming/web/yuanju
git rev-parse --abbrev-ref HEAD       # 应输出 main
git status --short                     # 应空
git rev-list --left-right --count main...origin/main   # 应 0 0
```

如果不在 main 或工作树脏，STOP，报告状态。

- [ ] **Step 2: 起新分支**

```bash
git checkout -b feat/polished-report
```

- [ ] **Step 3: 确认 spec 存在**

```bash
test -f docs/superpowers/specs/2026-05-17-polished-report-personalization-design.md && echo "spec OK"
```

Expected: `spec OK`

---

## Task 1: DB DDL · 新增 ai_polished_reports 表

**Files:**
- Modify: `backend/pkg/database/database.go`

- [ ] **Step 1: 找到 ai_reports DDL 位置**

```bash
cd /Users/liujiming/web/yuanju
grep -n "CREATE TABLE IF NOT EXISTS ai_reports" backend/pkg/database/database.go
```

记下行号。新表 DDL 紧邻其后。

- [ ] **Step 2: 找到 idx_ai_reports_chart_id 索引行附近**

```bash
grep -n "idx_ai_reports_chart_id\|CREATE INDEX" backend/pkg/database/database.go | head -5
```

新建表 + 索引应紧接 ai_reports 相关定义之后。

- [ ] **Step 3: 在 ai_reports 索引行之后追加新表 DDL**

在 `idx_ai_reports_chart_id` 索引创建语句之后、`admins` 表 CREATE 之前，插入：

```go
	CREATE TABLE IF NOT EXISTS ai_polished_reports (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		chart_id UUID UNIQUE NOT NULL REFERENCES bazi_charts(id) ON DELETE CASCADE,
		user_situation TEXT NOT NULL,
		content TEXT NOT NULL,
		content_structured JSONB,
		model VARCHAR(50),
		prompt_tokens INT DEFAULT 0,
		completion_tokens INT DEFAULT 0,
		total_tokens INT DEFAULT 0,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_ai_polished_reports_chart_id ON ai_polished_reports(chart_id);
```

注意：与 `ai_reports` 表风格保持一致 —— 制表符缩进、SQL 关键字大写。

- [ ] **Step 4: 编译检查**

```bash
cd /Users/liujiming/web/yuanju/backend
go build ./...
```

Expected: 无 error。

- [ ] **Step 5: Commit**

```bash
cd /Users/liujiming/web/yuanju
git add backend/pkg/database/database.go
git commit -m "feat(db): add ai_polished_reports table for 润色版 personalization

UNIQUE chart_id + ON DELETE CASCADE。content + content_structured 双
存以兼容 ai_reports 模式。token 统计字段供后续 token usage 分析。"
```

---

## Task 2: Model · PolishedReport struct

**Files:**
- Create: `backend/internal/model/polished_report.go`

- [ ] **Step 1: 检查 model 包目录**

```bash
ls /Users/liujiming/web/yuanju/backend/internal/model/
```

- [ ] **Step 2: 创建 model 文件**

写入 `backend/internal/model/polished_report.go`:

```go
package model

import (
	"encoding/json"
	"time"
)

// PolishedReport 润色版命理解读
// 一个 chart 对应至多 1 条记录（UNIQUE chart_id），重生覆盖。
type PolishedReport struct {
	ID                string           `json:"id"`
	ChartID           string           `json:"chart_id"`
	UserSituation     string           `json:"user_situation"`
	Content           string           `json:"content"`
	ContentStructured *json.RawMessage `json:"content_structured,omitempty"`
	Model             string           `json:"model"`
	PromptTokens      int              `json:"prompt_tokens"`
	CompletionTokens  int              `json:"completion_tokens"`
	TotalTokens       int              `json:"total_tokens"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
}
```

- [ ] **Step 3: 编译检查**

```bash
cd /Users/liujiming/web/yuanju/backend
go build ./...
```

Expected: 无 error。

- [ ] **Step 4: Commit**

```bash
cd /Users/liujiming/web/yuanju
git add backend/internal/model/polished_report.go
git commit -m "feat(model): add PolishedReport struct"
```

---

## Task 3: Repository · Upsert + Get

**Files:**
- Create: `backend/internal/repository/polished_report_repo.go`

- [ ] **Step 1: 读已有 repository 模式参考**

```bash
sed -n '222,265p' /Users/liujiming/web/yuanju/backend/internal/repository/repository.go
```

观察 CreateReport / GetReportByChartID 的写法，新文件保持同款风格。

- [ ] **Step 2: 创建 repository 文件**

写入 `backend/internal/repository/polished_report_repo.go`:

```go
package repository

import (
	"database/sql"
	"encoding/json"

	"yuanju/internal/model"
	"yuanju/pkg/database"
)

// UpsertPolishedReport 创建或覆盖一条润色版报告。
// UNIQUE chart_id 约束 → 同一命盘只保留最新一份。
func UpsertPolishedReport(
	chartID, userSituation, content, modelName string,
	contentStructured *json.RawMessage,
	promptTokens, completionTokens, totalTokens int,
) (*model.PolishedReport, error) {
	report := &model.PolishedReport{}
	err := database.DB.QueryRow(
		`INSERT INTO ai_polished_reports
		   (chart_id, user_situation, content, content_structured, model,
		    prompt_tokens, completion_tokens, total_tokens, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		 ON CONFLICT (chart_id) DO UPDATE SET
		   user_situation = EXCLUDED.user_situation,
		   content        = EXCLUDED.content,
		   content_structured = EXCLUDED.content_structured,
		   model          = EXCLUDED.model,
		   prompt_tokens  = EXCLUDED.prompt_tokens,
		   completion_tokens = EXCLUDED.completion_tokens,
		   total_tokens   = EXCLUDED.total_tokens,
		   updated_at     = NOW()
		 RETURNING id, chart_id, user_situation, content, content_structured,
		           model, prompt_tokens, completion_tokens, total_tokens,
		           created_at, updated_at`,
		chartID, userSituation, content, contentStructured, modelName,
		promptTokens, completionTokens, totalTokens,
	).Scan(
		&report.ID, &report.ChartID, &report.UserSituation, &report.Content,
		&report.ContentStructured, &report.Model,
		&report.PromptTokens, &report.CompletionTokens, &report.TotalTokens,
		&report.CreatedAt, &report.UpdatedAt,
	)
	return report, err
}

// GetPolishedByChartID 读指定命盘的润色版。无记录返回 (nil, nil)。
func GetPolishedByChartID(chartID string) (*model.PolishedReport, error) {
	report := &model.PolishedReport{}
	err := database.DB.QueryRow(
		`SELECT id, chart_id, user_situation, content, content_structured,
		        model, prompt_tokens, completion_tokens, total_tokens,
		        created_at, updated_at
		 FROM ai_polished_reports WHERE chart_id = $1`, chartID,
	).Scan(
		&report.ID, &report.ChartID, &report.UserSituation, &report.Content,
		&report.ContentStructured, &report.Model,
		&report.PromptTokens, &report.CompletionTokens, &report.TotalTokens,
		&report.CreatedAt, &report.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return report, err
}
```

- [ ] **Step 3: 编译检查**

```bash
cd /Users/liujiming/web/yuanju/backend
go build ./...
```

- [ ] **Step 4: Commit**

```bash
cd /Users/liujiming/web/yuanju
git add backend/internal/repository/polished_report_repo.go
git commit -m "feat(repo): polished_report Upsert + Get"
```

---

## Task 4: Prompt builder · buildPolishPrompt

**Files:**
- Create: `backend/internal/service/polished_report_service.go`（先放 prompt 函数 + 校验）
- Create: `backend/internal/service/polished_report_service_test.go`

- [ ] **Step 1: 写 service 文件骨架（含 buildPolishPrompt）**

写入 `backend/internal/service/polished_report_service.go`:

```go
package service

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"yuanju/internal/model"
	"yuanju/pkg/bazi"
)

const (
	polishMinChars = 20
	polishMaxChars = 300
)

// validatePolishSituation 校验用户输入的当前情况字符串。
// 长度按 unicode rune 计数（中文也算 1 字符），范围 [20, 300]。
func validatePolishSituation(s string) error {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return fmt.Errorf("当前情况描述不能为空")
	}
	n := utf8.RuneCountInString(trimmed)
	if n < polishMinChars {
		return fmt.Errorf("当前情况描述至少 %d 字（当前 %d 字）", polishMinChars, n)
	}
	if n > polishMaxChars {
		return fmt.Errorf("当前情况描述最多 %d 字（当前 %d 字）", polishMaxChars, n)
	}
	return nil
}

// buildPolishPrompt 构建润色 prompt：
// 原版 5 章 markdown + 命理数据 + 用户输入 → 让 LLM 逐章重写。
func buildPolishPrompt(originalReport *model.AIReport, result *bazi.BaziResult, userSituation string) string {
	originalContent := ""
	if originalReport != nil {
		originalContent = originalReport.Content
	}

	yongshen := result.Yongshen
	jishen := result.Jishen
	mingGe := result.MingGe
	mingGeDesc := result.MingGeDesc

	genderText := "女"
	if result.Gender == "male" {
		genderText = "男"
	}

	return strings.Join([]string{
		"你是一位资深命理师傅。下面给你一份已经生成好的命理解读「原版」、一份命理算法的精算数据、以及用户当下的具体处境。",
		"你的任务：基于原版结论，结合用户处境，重写一份「润色版」 —— 内容贴近用户当下，口吻像师傅当面讲给他听，避免报告体。",
		"",
		"[原版命理解读 · 严格保留命局结论]",
		originalContent,
		"",
		"[命理算法精算数据]",
		fmt.Sprintf("出生：%d年%d月%d日%d时 · %s命",
			result.BirthYear, result.BirthMonth, result.BirthDay, result.BirthHour, genderText),
		fmt.Sprintf("用神：%s ／ 忌神：%s", strDefault(yongshen, "—"), strDefault(jishen, "—")),
		fmt.Sprintf("命格：%s（%s）", strDefault(mingGe, "—"), strDefault(mingGeDesc, "—")),
		"",
		"[用户当下情况 · 用户自述]",
		strings.TrimSpace(userSituation),
		"",
		"[改写要求]",
		"1. 保持原版的 5 章结构（性格特质 / 感情运势 / 事业财运 / 健康提示 / 大运走势），章节顺序不变。",
		"2. 每章 200-300 字。第二人称「你」叙述，师傅口吻，温润不冷漠。",
		"3. 每章首段必须引出用户情况里相关的事，再带回命理依据说明为什么。",
		"4. 不可改变原版的命局结论：用神 / 忌神 / 命格 / 十神等核心判断与原版一致。",
		"5. 严禁出现 Markdown 加粗 `**`、斜体 `*`、子标题 `###` 等格式符号。",
		"6. 严禁出现「百分比 / 加权 / 权重 / 由此可见 / 综上所述 / 通过分析可得 / 显然 / 总体而言」等公文体词汇。",
		"7. 五行强弱用「旺/相/休/囚/死」「过旺/偏旺/平衡/偏弱/缺」等命理术语，不用数字。",
		"8. 段落自然换行（章内多段之间空一行），不用列表 / bullet。",
		"",
		"[输出格式 · 严格遵守]",
		"必须按以下顺序、一字不差使用 Markdown 二级标题；每章标题独占一行；除标题外不要其它格式标记：",
		"",
		"## 【性格特质】",
		"...（200-300 字润色版正文，至少 2 段，段间空行）...",
		"",
		"## 【感情运势】",
		"...",
		"",
		"## 【事业财运】",
		"...",
		"",
		"## 【健康提示】",
		"...",
		"",
		"## 【大运走势】",
		"...",
	}, "\n")
}

// strDefault 返回 s 或 fallback（s 为空时）。
func strDefault(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}
```

- [ ] **Step 2: 写测试文件**

写入 `backend/internal/service/polished_report_service_test.go`:

```go
package service

import (
	"strings"
	"testing"

	"yuanju/internal/model"
	"yuanju/pkg/bazi"
)

func TestValidatePolishSituation(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"空字符串", "", true},
		{"全空白", "   \n  ", true},
		{"中文 19 字", strings.Repeat("一", 19), true},
		{"中文 20 字", strings.Repeat("一", 20), false},
		{"中文 300 字", strings.Repeat("一", 300), false},
		{"中文 301 字", strings.Repeat("一", 301), true},
		{"混合 25 字", "今年考虑跳槽，目前 work 太烧脑，想换条路走走看", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validatePolishSituation(c.input)
			if (err != nil) != c.wantErr {
				t.Fatalf("want err=%v, got %v", c.wantErr, err)
			}
		})
	}
}

func TestBuildPolishPrompt_IncludesAllParts(t *testing.T) {
	original := &model.AIReport{
		Content: "## 【性格特质】\n原版性格章...\n## 【感情运势】\n原版感情章...",
	}
	result := &bazi.BaziResult{
		BirthYear: 1995, BirthMonth: 5, BirthDay: 12, BirthHour: 10,
		Gender: "male", Yongshen: "火、土", Jishen: "金", MingGe: "食神生财格", MingGeDesc: "月令偏财得用",
	}
	prompt := buildPolishPrompt(original, result, "  今年考虑跳槽，跟对象有点摩擦，想看看大方向  ")

	mustContain := []string{
		"原版命理解读",
		"原版性格章",
		"火、土",
		"金",
		"食神生财格",
		"今年考虑跳槽",
		"## 【性格特质】",
		"## 【大运走势】",
		"师傅口吻",
		"不可改变原版的命局结论",
	}
	for _, s := range mustContain {
		if !strings.Contains(prompt, s) {
			t.Errorf("prompt 缺少关键字: %q", s)
		}
	}
}
```

- [ ] **Step 3: 跑测试**

```bash
cd /Users/liujiming/web/yuanju/backend
go test ./internal/service/ -run "TestValidatePolishSituation|TestBuildPolishPrompt" -v
```

Expected: 全部 PASS。

- [ ] **Step 4: Commit**

```bash
cd /Users/liujiming/web/yuanju
git add backend/internal/service/polished_report_service.go backend/internal/service/polished_report_service_test.go
git commit -m "feat(service): polished_report prompt builder + input validation"
```

---

## Task 5: Service orchestration · PolishReport

**Files:**
- Modify: `backend/internal/service/polished_report_service.go`（追加 PolishReport 函数）

- [ ] **Step 1: 看 GenerateAIReport 怎么调 LLM（参考模式）**

```bash
grep -n "func GenerateAIReport\|callAIWithSystem\|parseAIReportContent" /Users/liujiming/web/yuanju/backend/internal/service/report_service.go | head -10
sed -n '411,470p' /Users/liujiming/web/yuanju/backend/internal/service/report_service.go
```

记下：`callAIWithSystem(prompt)` 返回 `(content, modelName, providerID, durationMs, usage, err)`；`parseAIReportContent(rawContent, cleanJSON)` 返回 `(*structuredReport, briefContent, *json.RawMessage)`。

- [ ] **Step 2: 更新顶部 import 块**

把 `backend/internal/service/polished_report_service.go` 顶部 import 块替换为：

```go
import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"unicode/utf8"

	"yuanju/internal/model"
	"yuanju/internal/repository"
	"yuanju/pkg/bazi"
)
```

新增 `encoding/json`、`log`、`yuanju/internal/repository`；其余保留。

- [ ] **Step 3: 在文件末尾追加 PolishReport 函数**

```go
// PolishReport 执行润色：load 原版 → 构建 prompt → 调 LLM → 解析 → upsert
//
// 不写 user_id 参数 — token 记账只在原版调用时执行；润色版仅记 token 字段到本表。
func PolishReport(chartID string, result *bazi.BaziResult, userSituation string) (*model.PolishedReport, error) {
	// 1. 校验输入
	if err := validatePolishSituation(userSituation); err != nil {
		return nil, err
	}

	// 2. load 原版（必须存在）
	original, err := repository.GetReportByChartID(chartID)
	if err != nil {
		return nil, fmt.Errorf("读取原版报告失败: %w", err)
	}
	if original == nil {
		return nil, fmt.Errorf("请先生成原版命理解读，再尝试润色")
	}

	// 3. 构建 prompt
	prompt := buildPolishPrompt(original, result, strings.TrimSpace(userSituation))

	// 4. 调 LLM（非流式，复用原版 ai_client）
	log.Printf("[Polish] 开始润色 chart_id=%s situation_len=%d", chartID, utf8.RuneCountInString(userSituation))
	rawContent, modelName, _, durationMs, usage, aiErr := callAIWithSystem(prompt)
	if aiErr != nil {
		return nil, fmt.Errorf("LLM 调用失败: %w", aiErr)
	}
	log.Printf("[Polish] LLM 返回 chart_id=%s duration_ms=%d tokens=%d/%d/%d",
		chartID, durationMs, usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens)

	// 5. 解析 Markdown → 5 章 JSON
	cleanContent := strings.TrimSpace(rawContent)
	parsed, _ := ParseMarkdownToStructured(cleanContent)
	var contentStructured *json.RawMessage
	if parsed != nil && len(parsed.Chapters) > 0 {
		raw, _ := json.Marshal(parsed)
		rawMsg := json.RawMessage(raw)
		contentStructured = &rawMsg
	}

	// 6. UPSERT 到 ai_polished_reports
	report, err := repository.UpsertPolishedReport(
		chartID, strings.TrimSpace(userSituation), cleanContent, modelName,
		contentStructured,
		usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens,
	)
	if err != nil {
		return nil, fmt.Errorf("保存润色版失败: %w", err)
	}

	return report, nil
}
```

注意：`utf8.RuneCountInString` 已经在 Step 1 的 import 里了；如果 import 有重复需合并到顶部。

- [ ] **Step 4: 编译检查**

```bash
cd /Users/liujiming/web/yuanju/backend
go build ./...
```

Expected: 无 error。

- [ ] **Step 5: 跑既有测试不退化**

```bash
go test ./internal/service/ 2>&1 | tail -5
```

Expected: 全部 PASS。

- [ ] **Step 6: Commit**

```bash
cd /Users/liujiming/web/yuanju
git add backend/internal/service/polished_report_service.go
git commit -m "feat(service): PolishReport orchestration"
```

---

## Task 6: Handler + Routes

**Files:**
- Create: `backend/internal/handler/polished_report_handler.go`
- Modify: `backend/cmd/api/main.go`

- [ ] **Step 1: 看 GenerateReport handler 的归属验证模式**

```bash
sed -n '93,135p' /Users/liujiming/web/yuanju/backend/internal/handler/bazi_handler.go
```

记下 chart 归属验证三步（GetChartByID → 检查 UserID → LoadOrCalculateResult）。

- [ ] **Step 2: 创建 handler 文件**

写入 `backend/internal/handler/polished_report_handler.go`:

```go
package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"yuanju/internal/repository"
	"yuanju/internal/service"
)

// generateAndSavePolishedReportRequest POST body
type generatePolishedReportRequest struct {
	UserSituation string `json:"user_situation" binding:"required"`
}

// GenerateAndSavePolishedReport
// POST /api/bazi/polished-report/:chart_id
func GenerateAndSavePolishedReport(c *gin.Context) {
	chartID := c.Param("chart_id")
	if chartID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的排盘ID"})
		return
	}

	var body generatePolishedReportRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请填写当前情况描述"})
		return
	}

	chart, err := repository.GetChartByID(chartID)
	if err != nil || chart == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到指定命盘记录"})
		return
	}

	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)
	if chart.UserID == nil || *chart.UserID != userIDStr {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作此命盘"})
		return
	}

	result, err := service.LoadOrCalculateResult(chart)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[Polish] 请求 chart_id=%s user=%s", chart.ID, userIDStr)
	report, err := service.PolishReport(chart.ID, result, body.UserSituation)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"polished_report": report})
}

// GetPolishedReport
// GET /api/bazi/polished-report/:chart_id
func GetPolishedReport(c *gin.Context) {
	chartID := c.Param("chart_id")
	if chartID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的排盘ID"})
		return
	}

	chart, err := repository.GetChartByID(chartID)
	if err != nil || chart == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到指定命盘记录"})
		return
	}

	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)
	if chart.UserID == nil || *chart.UserID != userIDStr {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作此命盘"})
		return
	}

	report, err := repository.GetPolishedByChartID(chart.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"polished_report": report})
}
```

- [ ] **Step 3: 注册路由**

在 `backend/cmd/api/main.go` 找到现有 bazi report 路由（应该在第 58-59 行附近）：

```bash
grep -n "POST.*report\|GET.*report" /Users/liujiming/web/yuanju/backend/cmd/api/main.go | head -5
```

在 `bazi.POST("/report-stream/:chart_id", ...)` 行之后追加两行：

```go
			bazi.POST("/polished-report/:chart_id", middleware.Auth(), handler.GenerateAndSavePolishedReport)
			bazi.GET("/polished-report/:chart_id", middleware.Auth(), handler.GetPolishedReport)
```

- [ ] **Step 4: 编译**

```bash
cd /Users/liujiming/web/yuanju/backend
go build ./...
```

Expected: 无 error。

- [ ] **Step 5: 跑所有 backend 测试不退化**

```bash
go test ./... 2>&1 | tail -10
```

Expected: 全部 PASS。

- [ ] **Step 6: Commit**

```bash
cd /Users/liujiming/web/yuanju
git add backend/internal/handler/polished_report_handler.go backend/cmd/api/main.go
git commit -m "feat(handler): polished_report POST/GET routes"
```

---

## Task 7: Frontend API client

**Files:**
- Modify: `frontend/src/lib/api.ts`

- [ ] **Step 1: 找 baziAPI 对象位置**

```bash
grep -n "export const baziAPI\|generateReport:" /Users/liujiming/web/yuanju/frontend/src/lib/api.ts | head -5
```

- [ ] **Step 2: 看 AIReport 类型定义和 baziAPI 现有方法**

```bash
sed -n '125,135p' /Users/liujiming/web/yuanju/frontend/src/lib/api.ts
sed -n '270,295p' /Users/liujiming/web/yuanju/frontend/src/lib/api.ts
```

- [ ] **Step 3: 追加 PolishedReport 类型**

在 `export interface AIReport { ... }` 块之后追加：

```ts
export interface PolishedReport {
  id: string
  chart_id: string
  user_situation: string
  content: string
  content_structured?: StructuredReport | null
  model: string
  prompt_tokens: number
  completion_tokens: number
  total_tokens: number
  created_at: string
  updated_at: string
}
```

- [ ] **Step 4: 在 baziAPI 对象里追加 2 个方法**

在 `baziAPI` 对象的 `generateReport: ...` 之后追加：

```ts
  getPolishedReport: (chartId: string) =>
    api.get<{ polished_report: PolishedReport | null }>(`/api/bazi/polished-report/${chartId}`),

  generatePolishedReport: (chartId: string, userSituation: string) =>
    api.post<{ polished_report: PolishedReport }>(
      `/api/bazi/polished-report/${chartId}`,
      { user_situation: userSituation },
    ),
```

- [ ] **Step 5: TypeScript 编译**

```bash
cd /Users/liujiming/web/yuanju/frontend
rm -rf node_modules/.cache .tsbuildinfo
npx tsc -b --force 2>&1 | head -5
```

Expected: 0 errors。

- [ ] **Step 6: Commit**

```bash
cd /Users/liujiming/web/yuanju
git add frontend/src/lib/api.ts
git commit -m "feat(api): baziAPI getPolishedReport + generatePolishedReport"
```

---

## Task 8: PolishedPanel 组件

**Files:**
- Create: `frontend/src/components/PolishedPanel.tsx`
- Create: `frontend/src/components/PolishedPanel.css`
- Create: `frontend/tests/polished-report-ui.test.mjs`

- [ ] **Step 1: 写失败测试**

写入 `frontend/tests/polished-report-ui.test.mjs`:

```js
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const root = new URL('..', import.meta.url).pathname
const read = (p) => readFileSync(join(root, p), 'utf8')

test('PolishedPanel 三态：empty / has-input / has-report 都渲染', () => {
  const c = read('src/components/PolishedPanel.tsx')
  assert.match(c, /className="polished-panel"/)
  assert.match(c, /polished-empty-state/)
  assert.match(c, /polished-input-area/)
  assert.match(c, /polished-content-area/)
})

test('PolishedPanel 包含 user_situation 输入 + 字数提示 + 提交按钮', () => {
  const c = read('src/components/PolishedPanel.tsx')
  assert.match(c, /<textarea/)
  assert.match(c, /maxLength=\{?300/)
  assert.match(c, /生成润色版|重新润色/)
})

test('PolishedPanel 渲染 5 章节复用 chapter 列表逻辑', () => {
  const c = read('src/components/PolishedPanel.tsx')
  assert.match(c, /chapters/)
  assert.match(c, /cleanReportText/)
  assert.match(c, /splitParagraphs/)
})

test('PolishedPanel.css 定义 tab + panel + input 基础样式', () => {
  const css = read('src/components/PolishedPanel.css')
  assert.match(css, /\.polished-panel\b/)
  assert.match(css, /\.polished-input-area\b/)
  assert.match(css, /\.polished-content-area\b/)
})

test('ResultPage 装配 PolishedPanel + tab 切换', () => {
  const page = read('src/pages/ResultPage.tsx')
  assert.match(page, /import PolishedPanel from/)
  assert.match(page, /reportTab/)
  assert.match(page, /<PolishedPanel/)
})
```

- [ ] **Step 2: 跑测试确认失败**

```bash
cd /Users/liujiming/web/yuanju/frontend
node --test tests/polished-report-ui.test.mjs
```

Expected: 全部 FAIL（文件不存在）。

- [ ] **Step 3: 写 PolishedPanel.tsx**

```tsx
import { useState } from 'react'
import type { PolishedReport, StructuredReport } from '../lib/api'
import { cleanReportText, splitParagraphs } from '../lib/reportText'
import './PolishedPanel.css'

interface Props {
  polishedReport: PolishedReport | null
  hasOriginalReport: boolean
  loading: boolean
  errorMsg: string | null
  onSubmit: (userSituation: string) => Promise<void>
}

const MIN_LEN = 20
const MAX_LEN = 300

function paragraphsOf(text: string | undefined | null): string[] {
  return splitParagraphs(text)
}

function chaptersFrom(structured: StructuredReport | null | undefined): { title: string; detail: string }[] {
  return structured?.chapters?.map(c => ({ title: c.title, detail: c.detail || c.brief })) ?? []
}

export default function PolishedPanel({ polishedReport, hasOriginalReport, loading, errorMsg, onSubmit }: Props) {
  const [editing, setEditing] = useState<boolean>(!polishedReport)
  const [input, setInput] = useState<string>(polishedReport?.user_situation ?? '')

  if (!hasOriginalReport) {
    return (
      <div className="polished-panel">
        <div className="polished-empty-state">
          <p className="polished-empty-tip">请先生成「原版」命理解读，再尝试润色版。</p>
        </div>
      </div>
    )
  }

  const inputLen = input.trim().length
  const canSubmit = inputLen >= MIN_LEN && inputLen <= MAX_LEN && !loading

  const submit = async () => {
    if (!canSubmit) return
    await onSubmit(input.trim())
    setEditing(false)
  }

  // 已有润色版且不在编辑：展示态
  if (polishedReport && !editing) {
    const chapters = chaptersFrom(polishedReport.content_structured)
    return (
      <div className="polished-panel">
        <div className="polished-content-area">
          <div className="polished-context-bar">
            <span className="polished-context-label">你的情况描述：</span>
            <span className="polished-context-text">"{polishedReport.user_situation}"</span>
            <button className="polished-context-edit" onClick={() => setEditing(true)}>
              修改 / 重新润色
            </button>
          </div>
          {chapters.length === 0 ? (
            <p className="polished-empty-tip">润色版解析为空，请重新润色。</p>
          ) : (
            <div className="polished-chapter-list">
              {chapters.map((ch, i) => {
                const paras = paragraphsOf(ch.detail)
                return (
                  <div key={i} className="polished-chapter">
                    <h3 className="polished-chapter-title serif">【{cleanReportText(ch.title)}】</h3>
                    {paras.length > 0
                      ? paras.map((p, j) => <p key={j} className="polished-chapter-body">{p}</p>)
                      : <p className="polished-chapter-body">{cleanReportText(ch.detail)}</p>}
                  </div>
                )
              })}
            </div>
          )}
        </div>
      </div>
    )
  }

  // 空态 / 编辑态：输入区
  return (
    <div className="polished-panel">
      <div className="polished-input-area">
        <p className="polished-input-hint">
          简单说说你最近最关心的情况，AI 师傅会基于这个写一份贴你处境的润色版。
        </p>
        <textarea
          className="polished-input"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          maxLength={MAX_LEN}
          placeholder="例：今年在考虑跳槽，想做创意类工作；跟对象因发展节奏不一有点摩擦..."
          rows={5}
        />
        <div className="polished-input-meta">
          <span className={inputLen < MIN_LEN ? 'polished-len-warn' : 'polished-len-ok'}>
            {inputLen} / {MAX_LEN} 字（至少 {MIN_LEN} 字）
          </span>
          <div className="polished-input-actions">
            {polishedReport && (
              <button className="polished-btn-ghost" onClick={() => { setEditing(false); setInput(polishedReport.user_situation) }}>
                取消
              </button>
            )}
            <button className="polished-btn-primary" disabled={!canSubmit} onClick={submit}>
              {loading ? '润色中…' : (polishedReport ? '重新润色' : '生成润色版')}
            </button>
          </div>
        </div>
        {errorMsg && <p className="polished-error">{errorMsg}</p>}
      </div>
    </div>
  )
}
```

- [ ] **Step 4: 写 PolishedPanel.css**

```css
.polished-panel {
  padding: 12px 0;
}

.polished-empty-state {
  padding: 40px 16px;
  text-align: center;
  color: var(--text-secondary);
}

.polished-empty-tip {
  color: var(--text-secondary);
  font-size: 14px;
  line-height: 1.7;
}

.polished-input-area {
  background: var(--bg-card);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  padding: 16px;
}

.polished-input-hint {
  margin: 0 0 10px;
  color: var(--text-secondary);
  font-size: 13px;
  line-height: 1.7;
}

.polished-input {
  width: 100%;
  padding: 10px 12px;
  background: var(--bg-base);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  color: var(--text-primary);
  font-size: 14px;
  line-height: 1.7;
  font-family: inherit;
  resize: vertical;
  min-height: 110px;
}

.polished-input:focus {
  outline: none;
  border-color: var(--primary);
  box-shadow: 0 0 0 3px var(--primary-glow);
}

.polished-input-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 10px;
  font-size: 12px;
}

.polished-len-warn { color: #c8643a; }
.polished-len-ok   { color: var(--text-muted); }

.polished-input-actions {
  display: flex;
  gap: 8px;
}

.polished-btn-primary {
  background: var(--primary);
  color: #0d0f14;
  border: none;
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
  min-height: 36px;
}
.polished-btn-primary:disabled { opacity: 0.45; cursor: not-allowed; }
.polished-btn-primary:not(:disabled):hover { background: var(--primary-hover); }

.polished-btn-ghost {
  background: transparent;
  color: var(--text-secondary);
  border: 1px solid var(--border-default);
  padding: 8px 14px;
  border-radius: var(--radius-sm);
  font-size: 13px;
  cursor: pointer;
  min-height: 36px;
}

.polished-error {
  margin-top: 10px;
  color: #c8643a;
  font-size: 13px;
}

.polished-content-area {
  background: var(--bg-card);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  padding: 16px;
}

.polished-context-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 6px;
  padding-bottom: 10px;
  margin-bottom: 14px;
  border-bottom: 1px dashed var(--border-default);
}

.polished-context-label {
  color: var(--text-muted);
  font-size: 12px;
}

.polished-context-text {
  flex: 1;
  color: var(--text-secondary);
  font-size: 13px;
  font-style: italic;
  word-break: break-word;
}

.polished-context-edit {
  background: transparent;
  border: none;
  color: var(--primary);
  font-size: 12px;
  cursor: pointer;
  padding: 4px 8px;
}
.polished-context-edit:hover { text-decoration: underline; }

.polished-chapter-list {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.polished-chapter-title {
  margin: 0 0 8px;
  font-size: 16px;
  color: var(--text-accent);
  font-weight: 700;
  padding-bottom: 4px;
  border-bottom: 1px dashed var(--border-default);
}

.polished-chapter-body {
  margin: 0 0 0.8em;
  font-size: 14px;
  line-height: 1.85;
  color: var(--text-primary);
}
.polished-chapter-body:last-child { margin-bottom: 0; }
```

- [ ] **Step 5: 跑测试**

```bash
cd /Users/liujiming/web/yuanju/frontend
node --test tests/polished-report-ui.test.mjs 2>&1 | tail -5
```

Expected: 前 4 项 PASS（第 5 项 "ResultPage 装配..." 仍 FAIL — Task 9 处理）。

- [ ] **Step 6: TypeScript 编译**

```bash
rm -rf node_modules/.cache .tsbuildinfo
npx tsc -b --force 2>&1 | head -5
```

Expected: 0 errors。

- [ ] **Step 7: Commit**

```bash
cd /Users/liujiming/web/yuanju
git add frontend/src/components/PolishedPanel.tsx frontend/src/components/PolishedPanel.css frontend/tests/polished-report-ui.test.mjs
git commit -m "feat(ui): PolishedPanel component (empty/input/display)"
```

---

## Task 9: ResultPage 集成 · Tab 切换 + PolishedPanel 装配

**Files:**
- Modify: `frontend/src/pages/ResultPage.tsx`
- Modify: `frontend/src/pages/ResultPage.css`

- [ ] **Step 1: 找现有 AI 解读区起点**

```bash
grep -n "AI 命理解读\|report-section\|{report &&\|reportMode" /Users/liujiming/web/yuanju/frontend/src/pages/ResultPage.tsx | head -10
```

定位 `<div className="report-section card animate-fade-up">` 块和 `reportMode` 状态附近。

- [ ] **Step 2: 加 import**

在 `frontend/src/pages/ResultPage.tsx` 顶部 import 区，紧邻其它 `import ... from '../components/...'` 行后追加：

```tsx
import PolishedPanel from '../components/PolishedPanel'
import type { PolishedReport } from '../lib/api'
```

- [ ] **Step 3: 加 state**

在 ResultPage 函数体内、`const [reportMode, setReportMode] = useState<'brief' | 'detail'>('detail')` 附近追加：

```tsx
const [reportTab, setReportTab] = useState<'original' | 'polished'>('original')
const [polishedReport, setPolishedReport] = useState<PolishedReport | null>(null)
const [polishing, setPolishing] = useState(false)
const [polishError, setPolishError] = useState<string | null>(null)
```

- [ ] **Step 4: 加载润色版的 useEffect**

在已有 `useEffect` 链中（拉 chart / report 那一段附近）追加一个：

```tsx
useEffect(() => {
  if (!targetId) return
  baziAPI.getPolishedReport(targetId)
    .then(res => setPolishedReport(res.data.polished_report || null))
    .catch(() => setPolishedReport(null))
}, [targetId])
```

注意 `targetId` 是当前 chartID 变量（参考 ResultPage 已有用法，可能叫 `result?.id` 或 `id`）。如果变量名不同，对应替换。

- [ ] **Step 5: 加 submit handler**

在 ResultPage 函数体内加：

```tsx
const submitPolish = async (userSituation: string) => {
  if (!targetId) return
  setPolishing(true)
  setPolishError(null)
  try {
    const res = await baziAPI.generatePolishedReport(targetId, userSituation)
    setPolishedReport(res.data.polished_report)
  } catch (e: unknown) {
    const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error || '润色失败，请稍后重试'
    setPolishError(msg)
  } finally {
    setPolishing(false)
  }
}
```

- [ ] **Step 6: 在 AI 解读区上方插入 tab 切换**

找 `report-section` 区块开始处（应该有 `<div className="report-section card animate-fade-up">`），在其内部最顶部加：

```tsx
{report && (
  <div className="report-tab-row">
    <button
      className={`report-tab${reportTab === 'original' ? ' is-active' : ''}`}
      onClick={() => setReportTab('original')}
    >
      原版
    </button>
    <button
      className={`report-tab${reportTab === 'polished' ? ' is-active' : ''}`}
      onClick={() => setReportTab('polished')}
    >
      润色版
    </button>
  </div>
)}
```

然后把整个现有的「原版」AI 报告渲染（包括 reportMode 切换、digest、chapter list、报告 disclaimer 等）外面包一层：

```tsx
{reportTab === 'original' && (
  <>
    {/* 现有原版报告渲染 JSX ... */}
  </>
)}
```

紧接着在 `</div>` 关闭 `.report-section` 之前加：

```tsx
{reportTab === 'polished' && (
  <PolishedPanel
    polishedReport={polishedReport}
    hasOriginalReport={!!report}
    loading={polishing}
    errorMsg={polishError}
    onSubmit={submitPolish}
  />
)}
```

- [ ] **Step 7: ResultPage.css 加 tab 样式**

找 ResultPage.css 顶部页面 shell 规则之后，追加：

```css
.report-tab-row {
  display: flex;
  gap: 8px;
  padding: 4px;
  background: var(--bg-base);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  margin-bottom: 16px;
  width: max-content;
}

.report-tab {
  background: transparent;
  color: var(--text-secondary);
  border: none;
  padding: 8px 18px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 13px;
  letter-spacing: 1px;
  min-height: 36px;
}

.report-tab.is-active {
  background: var(--primary);
  color: #0d0f14;
  font-weight: 700;
}

.report-tab:not(.is-active):hover {
  color: var(--text-primary);
}
```

- [ ] **Step 8: TypeScript 编译 + 测试**

```bash
cd /Users/liujiming/web/yuanju/frontend
rm -rf node_modules/.cache .tsbuildinfo
npx tsc -b --force 2>&1 | head -5
node --test tests/*.test.mjs 2>&1 | grep -E "^ℹ (tests|pass|fail)"
```

Expected: 0 TS errors；全部 tests pass（含本任务 Task 8 那个之前 FAIL 的"ResultPage 装配..."）。

- [ ] **Step 9: Commit**

```bash
cd /Users/liujiming/web/yuanju
git add frontend/src/pages/ResultPage.tsx frontend/src/pages/ResultPage.css
git commit -m "feat(result): wire PolishedPanel + report tab switch into ResultPage"
```

---

## Task 10: 真机 QA + 最终验证

**Files:** 无代码改动，只走查 + 可能发现的小修

- [ ] **Step 1: 起 dev server**

```bash
cd /Users/liujiming/web/yuanju
./scripts/docker-compose-up.sh
# 或前后端分别：
# cd backend && go run ./cmd/api
# cd frontend && npm run dev
```

打开 `http://localhost:5200`，登录已有账号或注册。

- [ ] **Step 2: 选一份已生成原版的命盘**

进 `/history` 选一个有"已生成"标记的命盘 → 进 `/history/:id`。

如果没有生成过的命盘：起一个新的 → 等待 AI 报告完成。

- [ ] **Step 3: 验证原版 tab**

- 默认进入显示「原版」高亮、原版报告内容
- 切到「润色版」，因为还没生成应显示输入区
- 切回「原版」，原版仍然在

- [ ] **Step 4: 验证空态 / 输入校验**

- 「润色版」 tab 下：
  - 不输入直接点按钮 → disabled
  - 输入 5 个字 → 计数器显示 `5 / 300（至少 20 字）`，红字提示
  - 输入 20 个字 → 计数器变绿、按钮可点

- [ ] **Step 5: 验证生成**

- 输入：「今年考虑跳槽，想换条创意类的路；跟对象因为节奏不同有点摩擦」
- 点「生成润色版」→ 按钮变「润色中…」
- 等 LLM 完成（10-30s）→ 显示 5 章润色版内容
- 验证：
  - 每章 200-300 字
  - 第二人称「你」开头
  - 提及「跳槽 / 创意 / 对象」等用户输入元素
  - 命局结论（用神 / 忌神 / 命格）与原版一致
  - 无 `**` markdown 残留

- [ ] **Step 6: 验证重生**

- 点「修改 / 重新润色」→ 回到输入态、显示之前输入内容
- 改输入：加一句「另外想问明年要不要买房」
- 重新提交 → 内容应有变化、包含「买房」相关
- 切到原版 → 内容应该与重生前一致（润色不影响原版）

- [ ] **Step 7: 验证刷新持久化**

- 刷新页面（Cmd+Shift+R）
- 进 `/history/:id`
- 切到「润色版」 tab → 应自动显示之前的润色内容（不需要重生）

- [ ] **Step 8: 验证 DB**

```bash
docker exec yuanju_postgres psql -U yuanju -d yuanju -c \
  "SELECT chart_id, user_situation, model, total_tokens, created_at, updated_at
   FROM ai_polished_reports
   ORDER BY updated_at DESC LIMIT 5;"
```

应看到刚才生成的记录。

- [ ] **Step 9: 跑完整自动化测试**

```bash
cd /Users/liujiming/web/yuanju/backend
go test ./... 2>&1 | tail -10

cd /Users/liujiming/web/yuanju/frontend
rm -rf node_modules/.cache .tsbuildinfo
npx tsc -b --force 2>&1 | head -3
npm run lint 2>&1 | tail -5
node --test tests/*.test.mjs 2>&1 | grep -E "^ℹ (tests|pass|fail)"
```

Expected: 后端 go test 通过、前端 tsc 0 errors、lint 0 errors、tests 55+ 全 pass。

- [ ] **Step 10: 创建 PR**

```bash
cd /Users/liujiming/web/yuanju
git push -u origin feat/polished-report 2>&1 | tail -5
gh pr create --title "feat(result): 命理解读·润色版个性化" --body "$(cat <<'EOF'
## Summary
- 在原版 AI 命理解读基础上加 tab 切换 + 「润色版」
- 用户输入 20-300 字当前情况后，AI 基于原版 5 章 + 用户输入逐章重写
- 新表 ai_polished_reports（UNIQUE chart_id），UPSERT 覆盖式更新

## Spec
- docs/superpowers/specs/2026-05-17-polished-report-personalization-design.md

## Plan
- docs/superpowers/plans/2026-05-17-polished-report-personalization.md

## Test plan
- [x] 后端 go test 全通过
- [x] 前端 tsc + lint + tests 全通过
- [x] 真机 QA：原版/润色版 tab 切换、20-300 字校验、生成+重生、刷新持久化、DB 记录写入

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)" 2>&1 | tail -3
```

如果 gh CLI 未鉴权，提示用户去 `https://github.com/xiaoming1995/yuanju/pull/new/feat/polished-report` 手动开。

---

## 完成判定

- [ ] 10 个 task 全部完成
- [ ] backend `go test ./...` 通过
- [ ] frontend `tsc -b --force` 0 errors
- [ ] frontend `node --test tests/*.test.mjs` 全 pass
- [ ] frontend `npm run lint` 0 errors（pre-existing warning 可接受）
- [ ] 真机 QA · Step 3-7 全部 ✓
- [ ] DB 验证看到记录 · Step 8 ✓
- [ ] PR 已创建链接到 spec + plan
