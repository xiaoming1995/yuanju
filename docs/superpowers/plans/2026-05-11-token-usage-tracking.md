# Token 用量统计 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 Admin 面板新增 Token 用量统计页面，记录每次 AI 调用的 token 消耗，支持按用户汇总和单次明细下钻。

**Architecture:** 在 `ai_client.go` 中解析 API 响应里的 `usage` 字段，通过所有 AI 调用函数的返回值传出；service 层调用完 AI 后异步写入 `token_usage_logs` 表；Admin API 提供汇总和明细接口；前端新增页面展示。

**Tech Stack:** Go + lib/pq（已有），React + TypeScript（已有），PostgreSQL（已有）

---

## 文件变更清单

| 操作 | 文件 | 职责 |
|---|---|---|
| 修改 | `backend/pkg/database/database.go` | 新增 `token_usage_logs` 表迁移 |
| 新建 | `backend/internal/repository/token_usage_repository.go` | 写入/查询 token 日志 |
| 修改 | `backend/internal/service/ai_client.go` | 捕获 usage 并更新所有函数签名 |
| 修改 | `backend/internal/service/report_service.go` | 更新 5 个 AI 调用点 + 服务签名 |
| 修改 | `backend/internal/service/celebrity_service.go` | 更新 1 个调用点 |
| 修改 | `backend/internal/service/compatibility_service.go` | 更新 1 个调用点 |
| 修改 | `backend/internal/handler/bazi_handler.go` | 向 service 传递 userID |
| 新建 | `backend/internal/handler/token_usage_handler.go` | 汇总 + 明细 Admin API |
| 修改 | `backend/cmd/api/main.go` | 注册 2 条 admin 路由 |
| 修改 | `frontend/src/lib/adminApi.ts` | 新增 token usage API 函数 |
| 新建 | `frontend/src/pages/admin/TokenUsagePage.tsx` | Admin 前端页面 |
| 修改 | `frontend/src/App.tsx` | 注册前端路由 |
| 修改 | `frontend/src/components/AdminLayout.tsx` | 侧边栏新增入口 |

---

## Task 1: 数据库迁移 — 新增 token_usage_logs 表

**Files:**
- Modify: `backend/pkg/database/database.go`（在末尾 `log.Println("✅ 数据库迁移完成")` 之前追加）

- [ ] **Step 1: 在 `database.go` 末尾（`log.Println("✅ 数据库迁移完成")` 之前）追加迁移代码**

```go
// 增量迁移 (token-usage-tracking)：用户 AI token 用量统计
tokenUsageMigration := `
CREATE TABLE IF NOT EXISTS token_usage_logs (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID REFERENCES users(id) ON DELETE SET NULL,
    chart_id          UUID REFERENCES bazi_charts(id) ON DELETE SET NULL,
    call_type         VARCHAR(50) NOT NULL,
    model             VARCHAR(100),
    provider_id       UUID REFERENCES llm_providers(id) ON DELETE SET NULL,
    prompt_tokens     INT NOT NULL DEFAULT 0,
    completion_tokens INT NOT NULL DEFAULT 0,
    total_tokens      INT NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_token_usage_user_id    ON token_usage_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_token_usage_created_at ON token_usage_logs(created_at);`
if _, err := DB.Exec(tokenUsageMigration); err != nil {
    log.Fatalf("增量迁移失败 (token_usage_logs): %v", err)
}
```

- [ ] **Step 2: 确认编译通过**

```bash
cd backend && go build ./...
```

期望：无错误输出。

- [ ] **Step 3: Commit**

```bash
git add backend/pkg/database/database.go
git commit -m "feat(db): 新增 token_usage_logs 表迁移"
```

---

## Task 2: Repository 层 — token_usage_repository.go

**Files:**
- Create: `backend/internal/repository/token_usage_repository.go`

- [ ] **Step 1: 创建文件，写入以下完整内容**

```go
package repository

import (
	"fmt"
	"log"
	"time"
	"yuanju/pkg/database"
)

type TokenUsageSummaryRow struct {
	UserID           string `json:"user_id"`
	Email            string `json:"email"`
	Nickname         string `json:"nickname"`
	RequestCount     int    `json:"request_count"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	TotalTokens      int    `json:"total_tokens"`
}

type TokenUsageDetailRow struct {
	ID               string    `json:"id"`
	CallType         string    `json:"call_type"`
	Model            string    `json:"model"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	TotalTokens      int       `json:"total_tokens"`
	CreatedAt        time.Time `json:"created_at"`
}

// CreateTokenUsageLog 异步写入一条 token 用量记录；调用方应 go 调用
func CreateTokenUsageLog(userID *string, chartID *string, callType, model, providerID string, promptTokens, completionTokens, totalTokens int) error {
	if totalTokens == 0 {
		return nil
	}
	var providerIDPtr *string
	if providerID != "" {
		providerIDPtr = &providerID
	}
	_, err := database.DB.Exec(`
		INSERT INTO token_usage_logs
			(user_id, chart_id, call_type, model, provider_id, prompt_tokens, completion_tokens, total_tokens)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		userID, chartID, callType, model, providerIDPtr,
		promptTokens, completionTokens, totalTokens,
	)
	if err != nil {
		return fmt.Errorf("CreateTokenUsageLog: %w", err)
	}
	return nil
}

// GetTokenUsageSummary 按用户聚合 token 消耗，from 含、to 含（日粒度）
func GetTokenUsageSummary(from, to time.Time) ([]TokenUsageSummaryRow, error) {
	toExcl := to.AddDate(0, 0, 1)
	rows, err := database.DB.Query(`
		SELECT
			u.id,
			u.email,
			COALESCE(u.nickname, '') AS nickname,
			COUNT(t.id)::int             AS request_count,
			COALESCE(SUM(t.prompt_tokens), 0)::int     AS prompt_tokens,
			COALESCE(SUM(t.completion_tokens), 0)::int AS completion_tokens,
			COALESCE(SUM(t.total_tokens), 0)::int      AS total_tokens
		FROM users u
		JOIN token_usage_logs t ON t.user_id = u.id
		WHERE t.created_at >= $1 AND t.created_at < $2
		GROUP BY u.id, u.email, u.nickname
		ORDER BY total_tokens DESC`,
		from, toExcl,
	)
	if err != nil {
		return nil, fmt.Errorf("GetTokenUsageSummary: %w", err)
	}
	defer rows.Close()

	var result []TokenUsageSummaryRow
	for rows.Next() {
		var r TokenUsageSummaryRow
		if err := rows.Scan(&r.UserID, &r.Email, &r.Nickname, &r.RequestCount,
			&r.PromptTokens, &r.CompletionTokens, &r.TotalTokens); err != nil {
			log.Printf("[TokenUsage] Scan 失败: %v", err)
			continue
		}
		result = append(result, r)
	}
	return result, nil
}

// GetTokenUsageDetail 查询单用户分页明细
func GetTokenUsageDetail(userID string, from, to time.Time, page, limit int) (total int, items []TokenUsageDetailRow, err error) {
	toExcl := to.AddDate(0, 0, 1)
	offset := (page - 1) * limit

	if err = database.DB.QueryRow(`
		SELECT COUNT(*) FROM token_usage_logs
		WHERE user_id = $1 AND created_at >= $2 AND created_at < $3`,
		userID, from, toExcl,
	).Scan(&total); err != nil {
		return 0, nil, fmt.Errorf("GetTokenUsageDetail count: %w", err)
	}

	rows, err := database.DB.Query(`
		SELECT id, call_type, COALESCE(model, ''), prompt_tokens, completion_tokens, total_tokens, created_at
		FROM token_usage_logs
		WHERE user_id = $1 AND created_at >= $2 AND created_at < $3
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5`,
		userID, from, toExcl, limit, offset,
	)
	if err != nil {
		return 0, nil, fmt.Errorf("GetTokenUsageDetail query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var r TokenUsageDetailRow
		if err := rows.Scan(&r.ID, &r.CallType, &r.Model,
			&r.PromptTokens, &r.CompletionTokens, &r.TotalTokens, &r.CreatedAt); err != nil {
			log.Printf("[TokenUsage] Scan detail 失败: %v", err)
			continue
		}
		items = append(items, r)
	}
	return total, items, nil
}
```

- [ ] **Step 2: 编译验证**

```bash
cd backend && go build ./...
```

期望：无错误。

- [ ] **Step 3: Commit**

```bash
git add backend/internal/repository/token_usage_repository.go
git commit -m "feat(repo): 新增 token_usage_repository"
```

---

## Task 3: AI Client — 捕获 token usage

**Files:**
- Modify: `backend/internal/service/ai_client.go`

本任务分三步：添加结构体、更新非流式调用、更新流式调用。

- [ ] **Step 1: 在 `AIMessage` 结构体定义之前（约第 82 行），添加两个新结构体**

```go
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type StreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}
```

- [ ] **Step 2: 在 `AIRequest` 结构体中新增 `StreamOptions` 字段**

找到：
```go
type AIRequest struct {
	Model       string      `json:"model"`
	Messages    []AIMessage `json:"messages"`
	MaxTokens   int         `json:"max_tokens"`
	Temperature float64     `json:"temperature"`
	Stream      bool        `json:"stream,omitempty"`
	EnableThinking *bool `json:"enable_thinking,omitempty"`
}
```

替换为：
```go
type AIRequest struct {
	Model          string         `json:"model"`
	Messages       []AIMessage    `json:"messages"`
	MaxTokens      int            `json:"max_tokens"`
	Temperature    float64        `json:"temperature"`
	Stream         bool           `json:"stream,omitempty"`
	EnableThinking *bool          `json:"enable_thinking,omitempty"`
	StreamOptions  *StreamOptions `json:"stream_options,omitempty"`
}
```

- [ ] **Step 3: 在 `AIResponse` 结构体中新增 `Usage` 字段**

找到：
```go
type AIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}
```

替换为：
```go
type AIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage TokenUsage `json:"usage"`
}
```

- [ ] **Step 4: 更新 `callOpenAICompatible` 返回值，带出 usage**

找到函数签名：
```go
func callOpenAICompatible(url, apiKey, modelName, systemPrompt, userPrompt string) (string, error) {
```

替换为：
```go
func callOpenAICompatible(url, apiKey, modelName, systemPrompt, userPrompt string) (string, TokenUsage, error) {
```

找到函数内的 return 语句，将：
```go
	if len(aiResp.Choices) == 0 {
		return "", fmt.Errorf("AI 返回内容为空")
	}
	if aiResp.Choices[0].FinishReason == "length" {
		return "", fmt.Errorf("AI 输出被截断（finish_reason=length），请检查 max_tokens 配置或缩短 Prompt")
	}
	return aiResp.Choices[0].Message.Content, nil
```

替换为：
```go
	if len(aiResp.Choices) == 0 {
		return "", TokenUsage{}, fmt.Errorf("AI 返回内容为空")
	}
	if aiResp.Choices[0].FinishReason == "length" {
		return "", TokenUsage{}, fmt.Errorf("AI 输出被截断（finish_reason=length），请检查 max_tokens 配置或缩短 Prompt")
	}
	return aiResp.Choices[0].Message.Content, aiResp.Usage, nil
```

同时找到函数内其余 error 返回（`return "", err`、`return "", fmt.Errorf(...)`），均改为 `return "", TokenUsage{}, ...`。

- [ ] **Step 5: 更新 `callOpenAICompatibleWithLog`**

找到：
```go
func callOpenAICompatibleWithLog(url, apiKey, modelName, systemPrompt, userPrompt string) (string, error) {
	t0 := time.Now()
	result, err := callOpenAICompatible(url, apiKey, modelName, systemPrompt, userPrompt)
	logAIPromptToFile(modelName, systemPrompt, userPrompt, result, time.Since(t0).Milliseconds(), err)
	return result, err
}
```

替换为：
```go
func callOpenAICompatibleWithLog(url, apiKey, modelName, systemPrompt, userPrompt string) (string, TokenUsage, error) {
	t0 := time.Now()
	result, usage, err := callOpenAICompatible(url, apiKey, modelName, systemPrompt, userPrompt)
	logAIPromptToFile(modelName, systemPrompt, userPrompt, result, time.Since(t0).Milliseconds(), err)
	return result, usage, err
}
```

- [ ] **Step 6: 更新 `callAIInternal` 签名与返回**

找到：
```go
func callAIInternal(systemPrompt, userPrompt string) (content, model, providerID string, durationMs int, err error) {
```

替换为：
```go
func callAIInternal(systemPrompt, userPrompt string) (content, model, providerID string, durationMs int, usage TokenUsage, err error) {
```

函数体内，找到使用 DB Provider 的分支：
```go
		result, callErr := callOpenAICompatibleWithLog(
			baseURL+"/v1/chat/completions",
			apiKey,
			provider.Model,
			systemPrompt,
			userPrompt,
		)
		elapsed := int(time.Since(start).Milliseconds())
		if callErr != nil {
			return "", provider.Model, provider.ID, elapsed, fmt.Errorf("Provider [%s] 调用失败: %w", provider.Name, callErr)
		}
		return result, provider.Model, provider.ID, elapsed, nil
```

替换为：
```go
		result, u, callErr := callOpenAICompatibleWithLog(
			baseURL+"/v1/chat/completions",
			apiKey,
			provider.Model,
			systemPrompt,
			userPrompt,
		)
		elapsed := int(time.Since(start).Milliseconds())
		if callErr != nil {
			return "", provider.Model, provider.ID, elapsed, TokenUsage{}, fmt.Errorf("Provider [%s] 调用失败: %w", provider.Name, callErr)
		}
		return result, provider.Model, provider.ID, elapsed, u, nil
```

找到 DeepSeek fallback 分支：
```go
		result, callErr := callOpenAICompatibleWithLog(
			configs.AppConfig.AIBaseURL+"/v1/chat/completions",
			configs.AppConfig.DeepSeekAPIKey,
			"deepseek-chat",
			systemPrompt,
			userPrompt,
		)
		elapsed := int(time.Since(start).Milliseconds())
		if callErr == nil {
			return result, "deepseek-chat", "", elapsed, nil
		}
```

替换为：
```go
		result, u, callErr := callOpenAICompatibleWithLog(
			configs.AppConfig.AIBaseURL+"/v1/chat/completions",
			configs.AppConfig.DeepSeekAPIKey,
			"deepseek-chat",
			systemPrompt,
			userPrompt,
		)
		elapsed := int(time.Since(start).Milliseconds())
		if callErr == nil {
			return result, "deepseek-chat", "", elapsed, u, nil
		}
```

找到 OpenAI fallback 分支：
```go
		result, callErr := callOpenAICompatibleWithLog(
			"https://api.openai.com/v1/chat/completions",
			configs.AppConfig.OpenAIAPIKey,
			"gpt-4o-mini",
			systemPrompt,
			userPrompt,
		)
		elapsed := int(time.Since(start).Milliseconds())
		if callErr == nil {
			return result, "gpt-4o-mini", "", elapsed, nil
		}
```

替换为：
```go
		result, u, callErr := callOpenAICompatibleWithLog(
			"https://api.openai.com/v1/chat/completions",
			configs.AppConfig.OpenAIAPIKey,
			"gpt-4o-mini",
			systemPrompt,
			userPrompt,
		)
		elapsed := int(time.Since(start).Milliseconds())
		if callErr == nil {
			return result, "gpt-4o-mini", "", elapsed, u, nil
		}
```

最后的 error return（函数末尾），将：
```go
	return "", "", "", 0, fmt.Errorf("未配置可用的 LLM Provider，请在 Admin 面板添加并激活一个 Provider")
```

替换为：
```go
	return "", "", "", 0, TokenUsage{}, fmt.Errorf("未配置可用的 LLM Provider，请在 Admin 面板添加并激活一个 Provider")
```

- [ ] **Step 7: 更新 `callAI` 和 `callAIWithSystem` 签名**

找到：
```go
func callAI(prompt string) (content, model, providerID string, durationMs int, err error) {
	return callAIInternal(defaultSystemPrompt, prompt)
}
```

替换为：
```go
func callAI(prompt string) (content, model, providerID string, durationMs int, usage TokenUsage, err error) {
	return callAIInternal(defaultSystemPrompt, prompt)
}
```

找到：
```go
func callAIWithSystem(userPrompt string) (content, model, providerID string, durationMs int, err error) {
	systemPrompt := buildKnowledgeBaseSystem()
	return callAIInternal(systemPrompt, userPrompt)
}
```

替换为：
```go
func callAIWithSystem(userPrompt string) (content, model, providerID string, durationMs int, usage TokenUsage, err error) {
	systemPrompt := buildKnowledgeBaseSystem()
	return callAIInternal(systemPrompt, userPrompt)
}
```

- [ ] **Step 8: 更新 `streamOpenAICompatible` — 捕获流式 usage**

找到函数签名：
```go
func streamOpenAICompatible(url, apiKey, modelName, systemPrompt, userPrompt string, callback func(string) error, onThinking func() error, enableThinking *bool) (string, error) {
```

替换为：
```go
func streamOpenAICompatible(url, apiKey, modelName, systemPrompt, userPrompt string, callback func(string) error, onThinking func() error, enableThinking *bool) (string, TokenUsage, error) {
```

在函数体内，在 `reqBody` 定义中追加 `StreamOptions`：

找到：
```go
	reqBody := AIRequest{
		Model: modelName,
		Messages: []AIMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		MaxTokens:      12000,
		Temperature:    1.0,
		Stream:         true,
		EnableThinking: enableThinking,
	}
```

替换为：
```go
	reqBody := AIRequest{
		Model: modelName,
		Messages: []AIMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		MaxTokens:      12000,
		Temperature:    1.0,
		Stream:         true,
		EnableThinking: enableThinking,
		StreamOptions:  &StreamOptions{IncludeUsage: true},
	}
```

在 `var contentBuilder strings.Builder` 声明之后（event loop 之前），追加：
```go
	var capturedUsage TokenUsage
```

在 event loop 内，在 `if json.Unmarshal([]byte(dataStr), &event) == nil && len(event.Choices) > 0 {` 块的结构体声明部分，找到：
```go
			var event struct {
				Choices []struct {
					Delta struct {
						Content          string `json:"content"`
						ReasoningContent string `json:"reasoning_content"`
					} `json:"delta"`
				} `json:"choices"`
			}
```

替换为：
```go
			var event struct {
				Choices []struct {
					Delta struct {
						Content          string `json:"content"`
						ReasoningContent string `json:"reasoning_content"`
					} `json:"delta"`
				} `json:"choices"`
				Usage TokenUsage `json:"usage"`
			}
```

将 `if json.Unmarshal([]byte(dataStr), &event) == nil && len(event.Choices) > 0 {` 改为：
```go
			if json.Unmarshal([]byte(dataStr), &event) == nil {
				if len(event.Choices) == 0 && event.Usage.TotalTokens > 0 {
					capturedUsage = event.Usage
				}
				if len(event.Choices) > 0 {
```

并在对应的右大括号之后补上外层 `if` 的闭合 `}`（原来只有 `if len > 0` 的一层大括号，现在变两层）。

具体地，把原来的：
```go
			if json.Unmarshal([]byte(dataStr), &event) == nil && len(event.Choices) > 0 {
				// 推理模型的思考阶段：通知前端正在推理
				if event.Choices[0].Delta.ReasoningContent != "" && !thinkingNotified && onThinking != nil {
					log.Printf("[AIStream T+%dms] 🧠 推理模型开始思考阶段", time.Since(t0).Milliseconds())
					_ = onThinking()
					thinkingNotified = true
				}
				// 正式内容输出
				chunk := event.Choices[0].Delta.Content
				if chunk != "" {
					chunkNum++
					if chunkNum == 1 {
						log.Printf("[AIStream T+%dms] ✅ 首个文字 chunk 到达: %q", time.Since(t0).Milliseconds(), chunk[:min(len(chunk), 20)])
					}
					contentBuilder.WriteString(chunk)
					if cbErr := callback(chunk); cbErr != nil {
						cancel()
						return "", cbErr
					}
				}
			}
```

替换为：
```go
			if json.Unmarshal([]byte(dataStr), &event) == nil {
				if len(event.Choices) == 0 && event.Usage.TotalTokens > 0 {
					capturedUsage = event.Usage
				}
				if len(event.Choices) > 0 {
					// 推理模型的思考阶段：通知前端正在推理
					if event.Choices[0].Delta.ReasoningContent != "" && !thinkingNotified && onThinking != nil {
						log.Printf("[AIStream T+%dms] 🧠 推理模型开始思考阶段", time.Since(t0).Milliseconds())
						_ = onThinking()
						thinkingNotified = true
					}
					// 正式内容输出
					chunk := event.Choices[0].Delta.Content
					if chunk != "" {
						chunkNum++
						if chunkNum == 1 {
							log.Printf("[AIStream T+%dms] ✅ 首个文字 chunk 到达: %q", time.Since(t0).Milliseconds(), chunk[:min(len(chunk), 20)])
						}
						contentBuilder.WriteString(chunk)
						if cbErr := callback(chunk); cbErr != nil {
							cancel()
							return "", TokenUsage{}, cbErr
						}
					}
				}
			}
```

最后，将函数末尾的 `return contentBuilder.String(), nil` 改为：
```go
	return contentBuilder.String(), capturedUsage, nil
```

同时将函数内其余 error return `return "", err` 改为 `return "", TokenUsage{}, err`。

- [ ] **Step 9: 更新 `streamAIWithSystemEx` 签名与返回**

找到：
```go
func streamAIWithSystemEx(userPrompt string, callback func(string) error, onThinking func() error, enableThinking *bool) (rawContent, model, providerID string, durationMs int, err error) {
```

替换为：
```go
func streamAIWithSystemEx(userPrompt string, callback func(string) error, onThinking func() error, enableThinking *bool) (rawContent, model, providerID string, durationMs int, usage TokenUsage, err error) {
```

函数体内 DB Provider 分支，将：
```go
		result, callErr := streamOpenAICompatible(baseURL+"/v1/chat/completions", apiKey, provider.Model, systemPrompt, userPrompt, callback, onThinking, enableThinking)
		elapsed := int(time.Since(start).Milliseconds())
		if callErr != nil {
			return result, provider.Model, provider.ID, elapsed, fmt.Errorf("Provider [%s] 调用失败: %w", provider.Name, callErr)
		}
		return result, provider.Model, provider.ID, elapsed, nil
```

替换为：
```go
		result, u, callErr := streamOpenAICompatible(baseURL+"/v1/chat/completions", apiKey, provider.Model, systemPrompt, userPrompt, callback, onThinking, enableThinking)
		elapsed := int(time.Since(start).Milliseconds())
		if callErr != nil {
			return result, provider.Model, provider.ID, elapsed, TokenUsage{}, fmt.Errorf("Provider [%s] 调用失败: %w", provider.Name, callErr)
		}
		return result, provider.Model, provider.ID, elapsed, u, nil
```

DeepSeek fallback 分支，将：
```go
		result, callErr := streamOpenAICompatible(configs.AppConfig.AIBaseURL+"/v1/chat/completions", configs.AppConfig.DeepSeekAPIKey, "deepseek-chat", systemPrompt, userPrompt, callback, onThinking, enableThinking)
		elapsed := int(time.Since(start).Milliseconds())
		if callErr == nil {
			return result, "deepseek-chat", "", elapsed, nil
		}
```

替换为：
```go
		result, u, callErr := streamOpenAICompatible(configs.AppConfig.AIBaseURL+"/v1/chat/completions", configs.AppConfig.DeepSeekAPIKey, "deepseek-chat", systemPrompt, userPrompt, callback, onThinking, enableThinking)
		elapsed := int(time.Since(start).Milliseconds())
		if callErr == nil {
			return result, "deepseek-chat", "", elapsed, u, nil
		}
```

OpenAI fallback 分支，将：
```go
		result, callErr := streamOpenAICompatible("https://api.openai.com/v1/chat/completions", configs.AppConfig.OpenAIAPIKey, "gpt-4o-mini", systemPrompt, userPrompt, callback, onThinking, enableThinking)
		elapsed := int(time.Since(start).Milliseconds())
		if callErr == nil {
			return result, "gpt-4o-mini", "", elapsed, nil
		}
```

替换为：
```go
		result, u, callErr := streamOpenAICompatible("https://api.openai.com/v1/chat/completions", configs.AppConfig.OpenAIAPIKey, "gpt-4o-mini", systemPrompt, userPrompt, callback, onThinking, enableThinking)
		elapsed := int(time.Since(start).Milliseconds())
		if callErr == nil {
			return result, "gpt-4o-mini", "", elapsed, u, nil
		}
```

末尾 error return 改为：
```go
	return "", "", "", 0, TokenUsage{}, fmt.Errorf("未配置可用的 LLM Provider")
```

- [ ] **Step 10: 更新 `StreamAIWithSystem` 和 `StreamAIWithSystemNoThink` 签名**

找到：
```go
func StreamAIWithSystem(userPrompt string, callback func(string) error, onThinking func() error) (rawContent, model, providerID string, durationMs int, err error) {
	return streamAIWithSystemEx(userPrompt, callback, onThinking, nil)
}
```

替换为：
```go
func StreamAIWithSystem(userPrompt string, callback func(string) error, onThinking func() error) (rawContent, model, providerID string, durationMs int, usage TokenUsage, err error) {
	return streamAIWithSystemEx(userPrompt, callback, onThinking, nil)
}
```

找到：
```go
func StreamAIWithSystemNoThink(userPrompt string, callback func(string) error, onThinking func() error) (rawContent, model, providerID string, durationMs int, err error) {
	disabled := false
	noThinkPrompt := userPrompt + "\n\n/no_think"
	return streamAIWithSystemEx(noThinkPrompt, callback, onThinking, &disabled)
}
```

替换为：
```go
func StreamAIWithSystemNoThink(userPrompt string, callback func(string) error, onThinking func() error) (rawContent, model, providerID string, durationMs int, usage TokenUsage, err error) {
	disabled := false
	noThinkPrompt := userPrompt + "\n\n/no_think"
	return streamAIWithSystemEx(noThinkPrompt, callback, onThinking, &disabled)
}
```

- [ ] **Step 11: 编译验证（此时 service 层调用 callAI* 的地方会报编译错误，正常，下一 Task 修复）**

```bash
cd backend && go build ./... 2>&1 | grep -v "token_usage" | head -30
```

期望：只看到 `report_service.go`、`celebrity_service.go`、`compatibility_service.go` 的赋值错误，没有 `ai_client.go` 内部的错误。

- [ ] **Step 12: Commit**

```bash
git add backend/internal/service/ai_client.go
git commit -m "feat(ai): 捕获 API usage 字段，更新所有调用函数签名"
```

---

## Task 4: Report Service — 更新 5 个 AI 调用点

**Files:**
- Modify: `backend/internal/service/report_service.go`

服务签名需要加 `userID *string` 参数（共 5 个函数），调用处在 Task 6 再更新 handler。

- [ ] **Step 1: 更新 `GenerateAIReport` 签名并写入日志（report_service.go:352）**

找到：
```go
func GenerateAIReport(chartID string, result *bazi.BaziResult) (*model.AIReport, error) {
```

替换为：
```go
func GenerateAIReport(chartID string, result *bazi.BaziResult, userID *string) (*model.AIReport, error) {
```

在函数体内，将（约第 362 行）：
```go
	rawContent, modelName, providerID, durationMs, aiErr := callAIWithSystem(prompt)
```

替换为：
```go
	rawContent, modelName, providerID, durationMs, usage, aiErr := callAIWithSystem(prompt)
```

在紧接着的 `repository.CreateAIRequestLog(...)` 调用之后，追加：
```go
	if aiErr == nil {
		go func() {
			if err := repository.CreateTokenUsageLog(userID, &chartID, "report", modelName, providerID,
				usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens); err != nil {
				log.Printf("[TokenUsage] 写入失败: %v", err)
			}
		}()
	}
```

- [ ] **Step 2: 更新 `GenerateAIReportStream` 签名并写入日志（report_service.go:486）**

找到：
```go
func GenerateAIReportStream(chartID string, result *bazi.BaziResult, onData func(string) error, onThinking func() error) error {
```

替换为：
```go
func GenerateAIReportStream(chartID string, result *bazi.BaziResult, userID *string, onData func(string) error, onThinking func() error) error {
```

将（约第 503 行）：
```go
	rawContent, modelName, providerID, durationMs, aiErr := StreamAIWithSystem(prompt, onData, onThinking)
```

替换为：
```go
	rawContent, modelName, providerID, durationMs, usage, aiErr := StreamAIWithSystem(prompt, onData, onThinking)
```

在紧接着的 `repository.CreateAIRequestLog(...)` 调用之后追加：
```go
	if aiErr == nil {
		go func() {
			if err := repository.CreateTokenUsageLog(userID, &chartID, "report_stream", modelName, providerID,
				usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens); err != nil {
				log.Printf("[TokenUsage] 写入失败: %v", err)
			}
		}()
	}
```

- [ ] **Step 3: 更新 `GenerateLiunianReport` 签名并写入日志（report_service.go:635）**

找到：
```go
func GenerateLiunianReport(chartID string, targetYear int) (*model.AILiunianReport, error) {
```

替换为：
```go
func GenerateLiunianReport(chartID string, targetYear int, userID *string) (*model.AILiunianReport, error) {
```

将（约第 717 行）：
```go
	rawContent, modelName, providerID, durationMs, aiErr := callAIWithSystem(parsedPrompt.String())
```

替换为：
```go
	rawContent, modelName, providerID, durationMs, usage, aiErr := callAIWithSystem(parsedPrompt.String())
```

在紧接着的两处 `repository.CreateAIRequestLog(...)` 调用之后（此函数有两处 CreateAIRequestLog），在成功路径（`aiErr == nil` 那一句）之后追加：
```go
	go func() {
		if err := repository.CreateTokenUsageLog(userID, &chartID, "liunian", modelName, providerID,
			usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens); err != nil {
			log.Printf("[TokenUsage] 写入失败: %v", err)
		}
	}()
```

（注意：此函数在 `aiErr != nil` 时提前 return，所以成功路径的 CreateAIRequestLog 之后加即可。）

- [ ] **Step 4: 更新 `GeneratePastEventsStream` 签名并写入日志（report_service.go:760）**

找到：
```go
func GeneratePastEventsStream(chartID string, onData func(string) error, onThinking func() error) error {
```

替换为：
```go
func GeneratePastEventsStream(chartID string, userID *string, onData func(string) error, onThinking func() error) error {
```

将（约第 861 行）：
```go
	rawContent, modelName, providerID, durationMs, aiErr := StreamAIWithSystemNoThink(parsedPrompt.String(), onData, onThinking)
```

替换为：
```go
	rawContent, modelName, providerID, durationMs, usage, aiErr := StreamAIWithSystemNoThink(parsedPrompt.String(), onData, onThinking)
```

在紧接着的 `repository.CreateAIRequestLog(...)` 之后追加：
```go
	if aiErr == nil {
		go func() {
			if err := repository.CreateTokenUsageLog(userID, &chartID, "dayun", modelName, providerID,
				usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens); err != nil {
				log.Printf("[TokenUsage] 写入失败: %v", err)
			}
		}()
	}
```

- [ ] **Step 5: 更新 `GenerateDayunSummariesStream` 签名并写入日志（report_service.go:988）**

找到：
```go
func GenerateDayunSummariesStream(chartID string, onItem func(item DayunSummaryStreamItem) error) error {
```

替换为：
```go
func GenerateDayunSummariesStream(chartID string, userID *string, onItem func(item DayunSummaryStreamItem) error) error {
```

在循环体内（约第 1119 行），将：
```go
		_, modelName, providerID, durationMs, aiErr := StreamAIWithSystemNoThink(pbuf.String(), func(chunk string) error {
```

替换为：
```go
		_, modelName, providerID, durationMs, usage, aiErr := StreamAIWithSystemNoThink(pbuf.String(), func(chunk string) error {
```

在循环体内 `repository.CreateAIRequestLog(...)` 之后追加（注意 `chartID` 在此函数内是字符串参数）：
```go
		if aiErr == nil {
			go func(u TokenUsage, mn, pid string) {
				if err := repository.CreateTokenUsageLog(userID, &chartID, "dayun", mn, pid,
					u.PromptTokens, u.CompletionTokens, u.TotalTokens); err != nil {
					log.Printf("[TokenUsage] 写入失败: %v", err)
				}
			}(usage, modelName, providerID)
		}
```

- [ ] **Step 6: 编译验证（此时 handler 调用旧签名会报错，正常）**

```bash
cd backend && go build ./... 2>&1 | grep -v "bazi_handler" | head -20
```

期望：只剩 `bazi_handler.go` 报参数不匹配，`report_service.go` 本身无错误。

- [ ] **Step 7: Commit**

```bash
git add backend/internal/service/report_service.go
git commit -m "feat(service): report_service 接入 token 用量统计"
```

---

## Task 5: Celebrity & Compatibility Services — 更新调用点

**Files:**
- Modify: `backend/internal/service/celebrity_service.go`
- Modify: `backend/internal/service/compatibility_service.go`

- [ ] **Step 1: 更新 celebrity_service.go（第 25 行）**

找到：
```go
	content, modelName, _, _, err := callAI(prompt)
```

替换为：
```go
	content, modelName, _, _, usage, err := callAI(prompt)
```

在 `if err != nil { return nil, err }` 之后追加：
```go
	go func() {
		if logErr := repository.CreateTokenUsageLog(nil, nil, "celebrity", modelName, "",
			usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens); logErr != nil {
			log.Printf("[TokenUsage] celebrity 写入失败: %v", logErr)
		}
	}()
```

注意：`celebrity_service.go` 需要 import `"yuanju/internal/repository"` 和 `"log"`（如果还没有的话，检查现有 imports）。

- [ ] **Step 2: 更新 compatibility_service.go（第 171 行）**

找到：
```go
	rawContent, modelName, _, _, aiErr := callAIWithSystem(parsed.String())
```

替换为：
```go
	rawContent, modelName, providerID, _, usage, aiErr := callAIWithSystem(parsed.String())
```

在 `if aiErr != nil { return nil, aiErr }` 之后追加：
```go
	go func(uid string) {
		if logErr := repository.CreateTokenUsageLog(&uid, nil, "compatibility", modelName, providerID,
			usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens); logErr != nil {
			log.Printf("[TokenUsage] compatibility 写入失败: %v", logErr)
		}
	}(userID)
```

- [ ] **Step 3: 编译验证**

```bash
cd backend && go build ./... 2>&1 | grep -v "bazi_handler" | head -20
```

期望：只剩 `bazi_handler.go` 的参数错误。

- [ ] **Step 4: Commit**

```bash
git add backend/internal/service/celebrity_service.go backend/internal/service/compatibility_service.go
git commit -m "feat(service): celebrity/compatibility 接入 token 用量统计"
```

---

## Task 6: Handler 层 — 向 service 传递 userID

**Files:**
- Modify: `backend/internal/handler/bazi_handler.go`

- [ ] **Step 1: 更新 `GenerateReport` handler（非流式，约第 124 行）**

找到（handler 内，`userIDStr` 已在第 108 行定义）：
```go
	report, err := service.GenerateAIReport(chart.ID, result)
```

替换为：
```go
	report, err := service.GenerateAIReport(chart.ID, result, &userIDStr)
```

- [ ] **Step 2: 更新 `GenerateReportStream` handler（约第 186 行）**

在 `service.GenerateAIReportStream(...)` 调用之前，`userIDStr` 已在第 161 行定义。

找到（注意 `onData` 是第一个回调参数）：
```go
	err = service.GenerateAIReportStream(chart.ID, result, func(chunk string) error {
```

替换为：
```go
	err = service.GenerateAIReportStream(chart.ID, result, &userIDStr, func(chunk string) error {
```

- [ ] **Step 3: 更新 `GenerateLiunianReport` handler（约第 248 行）**

找到（`userIDStr` 已在第 240 行定义）：
```go
	report, err := service.GenerateLiunianReport(chart.ID, req.TargetYear)
```

替换为：
```go
	report, err := service.GenerateLiunianReport(chart.ID, req.TargetYear, &userIDStr)
```

- [ ] **Step 4: 更新 `GenerateDayunSummariesStream` handler（约第 392 行）**

在调用之前，`userID` 是 `interface{}`（第 381 行 `userID, _ := c.Get("user_id")`）。需要先提取为字符串变量。在 `if chart.UserID == nil...` 校验之后（校验用 `userID.(string)` 内联），在 `c.Header(...)` 之前追加：
```go
	userIDStr := userID.(string)
```

然后将：
```go
	err = service.GenerateDayunSummariesStream(chartID, func(item service.DayunSummaryStreamItem) error {
```

替换为：
```go
	err = service.GenerateDayunSummariesStream(chartID, &userIDStr, func(item service.DayunSummaryStreamItem) error {
```

- [ ] **Step 5: 更新 `HandlePastEventsStream` handler（约第 433 行）**

同理，`userID` 是 `interface{}`（第 422 行 `userID, _ := c.Get("user_id")`）。在 `c.Header(...)` 之前追加：
```go
	userIDStr := userID.(string)
```

然后将：
```go
	err = service.GeneratePastEventsStream(chartID, func(chunk string) error {
```

替换为：
```go
	err = service.GeneratePastEventsStream(chartID, &userIDStr, func(chunk string) error {
```

- [ ] **Step 6: 编译验证（此时应全部通过）**

```bash
cd backend && go build ./...
```

期望：无任何错误。

- [ ] **Step 7: 运行测试**

```bash
cd backend && go test ./...
```

期望：所有测试通过，或无相关失败（项目测试主要在 `pkg/bazi`）。

- [ ] **Step 8: Commit**

```bash
git add backend/internal/handler/bazi_handler.go
git commit -m "feat(handler): 向 service 层传递 userID 用于 token 统计"
```

---

## Task 7: Admin API Handler

**Files:**
- Create: `backend/internal/handler/token_usage_handler.go`

- [ ] **Step 1: 创建文件，写入完整内容**

```go
package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"yuanju/internal/repository"
)

// AdminGetTokenUsageSummary GET /api/admin/token-usage/summary?from=YYYY-MM-DD&to=YYYY-MM-DD
func AdminGetTokenUsageSummary(c *gin.Context) {
	from, to := parseDateRange(c)
	rows, err := repository.GetTokenUsageSummary(from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if rows == nil {
		rows = []repository.TokenUsageSummaryRow{}
	}
	c.JSON(http.StatusOK, rows)
}

// AdminGetTokenUsageDetail GET /api/admin/token-usage/detail?user_id=xxx&from=&to=&page=1&limit=20
func AdminGetTokenUsageDetail(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id 必填"})
		return
	}
	from, to := parseDateRange(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	total, items, err := repository.GetTokenUsageDetail(userID, from, to, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if items == nil {
		items = []repository.TokenUsageDetailRow{}
	}
	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"items": items,
	})
}

// parseDateRange 解析 from/to 查询参数，默认当月第一天至今日
func parseDateRange(c *gin.Context) (from, to time.Time) {
	now := time.Now()
	defaultFrom := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	defaultTo := now

	fromStr := c.DefaultQuery("from", defaultFrom.Format("2006-01-02"))
	toStr := c.DefaultQuery("to", defaultTo.Format("2006-01-02"))

	from, _ = time.Parse("2006-01-02", fromStr)
	to, _ = time.Parse("2006-01-02", toStr)

	if from.IsZero() {
		from = defaultFrom
	}
	if to.IsZero() {
		to = defaultTo
	}
	return from, to
}
```

- [ ] **Step 2: 编译验证**

```bash
cd backend && go build ./...
```

期望：无错误。

- [ ] **Step 3: Commit**

```bash
git add backend/internal/handler/token_usage_handler.go
git commit -m "feat(handler): 新增 token 用量 Admin API handler"
```

---

## Task 8: 注册 Admin 路由

**Files:**
- Modify: `backend/cmd/api/main.go`

- [ ] **Step 1: 在 `adminAuth` 路由组中追加两条路由**

找到（约第 108 行）：
```go
			adminAuth.GET("/ai-logs", handler.AdminListAILogs)
			adminAuth.GET("/ai-logs/summary", handler.AdminGetAILogsSummary)
```

在之后追加：
```go
			adminAuth.GET("/token-usage/summary", handler.AdminGetTokenUsageSummary)
			adminAuth.GET("/token-usage/detail", handler.AdminGetTokenUsageDetail)
```

- [ ] **Step 2: 编译验证**

```bash
cd backend && go build ./...
```

期望：无错误。

- [ ] **Step 3: 快速功能验证（需运行服务端）**

```bash
cd backend && go run ./cmd/api &
sleep 3
# 获取 admin token
TOKEN=$(curl -s -X POST http://localhost:9002/api/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"your_admin@email.com","password":"your_password"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
# 测试 summary 接口
curl -s "http://localhost:9002/api/admin/token-usage/summary" \
  -H "Authorization: Bearer $TOKEN" | head -100
kill %1
```

期望：返回 `[]`（空数组，因为还没有数据）而非 500 错误。

- [ ] **Step 4: Commit**

```bash
git add backend/cmd/api/main.go
git commit -m "feat(routes): 注册 /api/admin/token-usage 路由"
```

---

## Task 9: 前端 API 客户端

**Files:**
- Modify: `frontend/src/lib/adminApi.ts`

- [ ] **Step 1: 在文件末尾导出新 API 对象（在最后一个 API 对象定义之后追加）**

```typescript
export const adminTokenUsageAPI = {
  summary: (from: string, to: string) =>
    adminApi.get(`/api/admin/token-usage/summary?from=${from}&to=${to}`),
  detail: (userId: string, from: string, to: string, page: number, limit = 20) =>
    adminApi.get(
      `/api/admin/token-usage/detail?user_id=${userId}&from=${from}&to=${to}&page=${page}&limit=${limit}`
    ),
}
```

- [ ] **Step 2: 编译验证**

```bash
cd frontend && npm run build 2>&1 | tail -20
```

期望：无 TypeScript 错误。

- [ ] **Step 3: Commit**

```bash
git add frontend/src/lib/adminApi.ts
git commit -m "feat(api): 新增 adminTokenUsageAPI 客户端函数"
```

---

## Task 10: 前端页面 — TokenUsagePage.tsx

**Files:**
- Create: `frontend/src/pages/admin/TokenUsagePage.tsx`

- [ ] **Step 1: 创建文件，写入完整内容**

```tsx
import { useState } from 'react'
import { BarChart2 } from 'lucide-react'
import { adminTokenUsageAPI } from '../../lib/adminApi'

interface SummaryRow {
  user_id: string
  email: string
  nickname: string
  request_count: number
  prompt_tokens: number
  completion_tokens: number
  total_tokens: number
}

interface DetailRow {
  id: string
  call_type: string
  model: string
  prompt_tokens: number
  completion_tokens: number
  total_tokens: number
  created_at: string
}

interface DetailData {
  total: number
  items: DetailRow[]
}

function fmt(n: number) {
  return n.toLocaleString('zh-CN')
}

function todayStr() {
  return new Date().toISOString().slice(0, 10)
}

function firstOfMonthStr() {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-01`
}

export default function TokenUsagePage() {
  const [from, setFrom] = useState(firstOfMonthStr())
  const [to, setTo] = useState(todayStr())
  const [summary, setSummary] = useState<SummaryRow[]>([])
  const [loading, setLoading] = useState(false)
  const [queried, setQueried] = useState(false)

  const [drawerUser, setDrawerUser] = useState<SummaryRow | null>(null)
  const [detail, setDetail] = useState<DetailData | null>(null)
  const [detailPage, setDetailPage] = useState(1)
  const [detailLoading, setDetailLoading] = useState(false)
  const detailLimit = 20

  const handleQuery = async () => {
    setLoading(true)
    try {
      const res = await adminTokenUsageAPI.summary(from, to)
      setSummary(res.data || [])
      setQueried(true)
    } finally {
      setLoading(false)
    }
  }

  const openDetail = async (row: SummaryRow, page = 1) => {
    setDrawerUser(row)
    setDetailPage(page)
    setDetailLoading(true)
    try {
      const res = await adminTokenUsageAPI.detail(row.user_id, from, to, page, detailLimit)
      setDetail(res.data)
    } finally {
      setDetailLoading(false)
    }
  }

  const closeDrawer = () => {
    setDrawerUser(null)
    setDetail(null)
  }

  const callTypeLabel: Record<string, string> = {
    report: '原局报告',
    report_stream: '流式报告',
    liunian: '流年',
    dayun: '大运',
    celebrity: '名人生成',
    compatibility: '合盘',
  }

  const detailTotalPages = detail ? Math.ceil(detail.total / detailLimit) : 1

  return (
    <div>
      <h1 className="admin-page-title" style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        <BarChart2 size={24} /> Token 用量统计
      </h1>

      {/* 筛选栏 */}
      <div className="admin-card" style={{ marginBottom: 24, display: 'flex', gap: 12, alignItems: 'center', flexWrap: 'wrap' }}>
        <label style={{ color: '#aaa', fontSize: 14 }}>开始日期</label>
        <input
          type="date"
          value={from}
          onChange={e => setFrom(e.target.value)}
          style={{ background: '#1a1a2e', color: '#e0e0e0', border: '1px solid #333', borderRadius: 6, padding: '6px 10px' }}
        />
        <label style={{ color: '#aaa', fontSize: 14 }}>结束日期</label>
        <input
          type="date"
          value={to}
          onChange={e => setTo(e.target.value)}
          style={{ background: '#1a1a2e', color: '#e0e0e0', border: '1px solid #333', borderRadius: 6, padding: '6px 10px' }}
        />
        <button
          className="admin-btn"
          onClick={handleQuery}
          disabled={loading}
          style={{ minWidth: 80 }}
        >
          {loading ? '查询中…' : '查询'}
        </button>
      </div>

      {/* 汇总表格 */}
      {queried && (
        <div className="admin-card">
          {summary.length === 0 ? (
            <div style={{ color: '#888', textAlign: 'center', padding: 32 }}>该时间段内无 token 消耗记录</div>
          ) : (
            <table className="admin-table" style={{ width: '100%' }}>
              <thead>
                <tr>
                  <th>用户邮箱</th>
                  <th>昵称</th>
                  <th style={{ textAlign: 'right' }}>请求次数</th>
                  <th style={{ textAlign: 'right' }}>输入 tokens</th>
                  <th style={{ textAlign: 'right' }}>输出 tokens</th>
                  <th style={{ textAlign: 'right' }}>总 tokens</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {summary.map(row => (
                  <tr key={row.user_id}>
                    <td>{row.email}</td>
                    <td>{row.nickname || '—'}</td>
                    <td style={{ textAlign: 'right' }}>{fmt(row.request_count)}</td>
                    <td style={{ textAlign: 'right' }}>{fmt(row.prompt_tokens)}</td>
                    <td style={{ textAlign: 'right' }}>{fmt(row.completion_tokens)}</td>
                    <td style={{ textAlign: 'right', fontWeight: 700, color: '#a78bfa' }}>{fmt(row.total_tokens)}</td>
                    <td>
                      <button className="admin-btn" style={{ padding: '4px 12px', fontSize: 13 }} onClick={() => openDetail(row)}>
                        明细
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {/* 明细抽屉 */}
      {drawerUser && (
        <div
          style={{
            position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.5)', zIndex: 1000,
            display: 'flex', justifyContent: 'flex-end',
          }}
          onClick={closeDrawer}
        >
          <div
            style={{
              width: 600, maxWidth: '95vw', background: '#12122a', height: '100%',
              overflowY: 'auto', padding: 24, boxShadow: '-4px 0 20px rgba(0,0,0,0.4)',
            }}
            onClick={e => e.stopPropagation()}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
              <h2 style={{ fontSize: 16, fontWeight: 700, color: '#e0e0e0' }}>
                {drawerUser.email} 的调用明细
              </h2>
              <button className="admin-btn" onClick={closeDrawer} style={{ padding: '4px 12px' }}>关闭</button>
            </div>

            {detailLoading ? (
              <div className="admin-loading">加载中…</div>
            ) : detail && detail.items.length === 0 ? (
              <div style={{ color: '#888', textAlign: 'center', padding: 32 }}>无记录</div>
            ) : detail ? (
              <>
                <table className="admin-table" style={{ width: '100%', marginBottom: 16 }}>
                  <thead>
                    <tr>
                      <th>时间</th>
                      <th>类型</th>
                      <th>模型</th>
                      <th style={{ textAlign: 'right' }}>输入</th>
                      <th style={{ textAlign: 'right' }}>输出</th>
                      <th style={{ textAlign: 'right' }}>总计</th>
                    </tr>
                  </thead>
                  <tbody>
                    {detail.items.map(item => (
                      <tr key={item.id}>
                        <td style={{ fontSize: 12, color: '#aaa' }}>
                          {new Date(item.created_at).toLocaleString('zh-CN', { hour12: false })}
                        </td>
                        <td>{callTypeLabel[item.call_type] ?? item.call_type}</td>
                        <td style={{ fontSize: 12, color: '#aaa' }}>{item.model}</td>
                        <td style={{ textAlign: 'right' }}>{fmt(item.prompt_tokens)}</td>
                        <td style={{ textAlign: 'right' }}>{fmt(item.completion_tokens)}</td>
                        <td style={{ textAlign: 'right', fontWeight: 700, color: '#a78bfa' }}>{fmt(item.total_tokens)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>

                {/* 分页 */}
                {detailTotalPages > 1 && (
                  <div style={{ display: 'flex', gap: 8, justifyContent: 'center' }}>
                    <button
                      className="admin-btn"
                      disabled={detailPage <= 1}
                      onClick={() => openDetail(drawerUser, detailPage - 1)}
                      style={{ padding: '4px 12px' }}
                    >
                      上一页
                    </button>
                    <span style={{ color: '#888', lineHeight: '32px', fontSize: 13 }}>
                      {detailPage} / {detailTotalPages}（共 {detail.total} 条）
                    </span>
                    <button
                      className="admin-btn"
                      disabled={detailPage >= detailTotalPages}
                      onClick={() => openDetail(drawerUser, detailPage + 1)}
                      style={{ padding: '4px 12px' }}
                    >
                      下一页
                    </button>
                  </div>
                )}
              </>
            ) : null}
          </div>
        </div>
      )}
    </div>
  )
}
```

- [ ] **Step 2: 编译验证**

```bash
cd frontend && npm run build 2>&1 | tail -20
```

期望：无 TypeScript 错误。

- [ ] **Step 3: Commit**

```bash
git add frontend/src/pages/admin/TokenUsagePage.tsx
git commit -m "feat(ui): 新增 Token 用量统计 Admin 页面"
```

---

## Task 11: 注册路由 + 侧边栏入口

**Files:**
- Modify: `frontend/src/App.tsx`
- Modify: `frontend/src/components/AdminLayout.tsx`

- [ ] **Step 1: 在 `App.tsx` 中导入并注册路由**

找到文件顶部的 import 区域，在 `import AdminAILogsPage` 之后追加：
```tsx
import TokenUsagePage from './pages/admin/TokenUsagePage'
```

找到 admin 路由组，在 `<Route path="ai-logs" .../>` 之后追加：
```tsx
<Route path="token-usage" element={<TokenUsagePage />} />
```

- [ ] **Step 2: 在 `AdminLayout.tsx` 中追加侧边栏入口**

找到文件顶部的 lucide-react import：
```tsx
import { Hexagon, LayoutDashboard, Bot, Users, FileText, BookOpen, SlidersHorizontal } from 'lucide-react'
```

在 import 中追加 `BarChart2`：
```tsx
import { Hexagon, LayoutDashboard, Bot, Users, FileText, BookOpen, SlidersHorizontal, BarChart2 } from 'lucide-react'
```

找到 `ai-logs` 那条 NavLink：
```tsx
          <NavLink to="/admin/ai-logs" className={({isActive}) => isActive ? 'admin-nav-item active' : 'admin-nav-item'}>
            <span style={{ display: 'flex', alignItems: 'center', marginRight: 10 }}><FileText size={18} /></span> AI 调用日志
          </NavLink>
```

在之后追加：
```tsx
          <NavLink to="/admin/token-usage" className={({isActive}) => isActive ? 'admin-nav-item active' : 'admin-nav-item'}>
            <span style={{ display: 'flex', alignItems: 'center', marginRight: 10 }}><BarChart2 size={18} /></span> Token 用量统计
          </NavLink>
```

- [ ] **Step 3: 编译验证**

```bash
cd frontend && npm run build 2>&1 | tail -20
```

期望：无错误，dist 正常生成。

- [ ] **Step 4: 启动前端开发服务器，手动验证**

```bash
cd frontend && npm run dev
```

打开浏览器访问 `http://localhost:5200/admin/token-usage`，登录后应看到：
- 页面标题"Token 用量统计"
- 日期筛选栏和"查询"按钮
- 点击查询后显示表格（或"无记录"提示）
- 侧边栏有"Token 用量统计"入口

- [ ] **Step 5: Commit**

```bash
git add frontend/src/App.tsx frontend/src/components/AdminLayout.tsx
git commit -m "feat(ui): 注册 Token 用量统计路由，添加侧边栏入口"
```

---

## 完成验证清单

- [ ] `go build ./...` 无错误
- [ ] `go test ./...` 无失败
- [ ] 后端启动时控制台无迁移错误（`✅ 数据库迁移完成`）
- [ ] Admin 面板侧边栏有"Token 用量统计"入口
- [ ] 日期筛选 + 查询返回正确数据
- [ ] 明细抽屉可打开并分页
- [ ] 触发一次 AI 报告生成后，Admin 页面能看到对应用户的 token 记录
- [ ] 关闭抽屉后主表格数据不丢失
