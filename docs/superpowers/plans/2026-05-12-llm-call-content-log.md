# LLM 调用内容日志 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在每次 LLM 调用记录中完整存储输入 prompt 和输出 response 原文，并在管理后台提供按需查看功能。

**Architecture:** 对 `token_usage_logs` 表增量迁移加 2 个 TEXT 列（nullable），更新 `CreateTokenUsageLog` 函数签名，在 6 处调用点传入 prompt 和 rawContent，新增 `GET /api/admin/token-usage/content/:id` 接口按需拉取，前端明细行加「查看」按钮弹出 modal。

**Tech Stack:** Go + PostgreSQL（TEXT/TOAST），Gin，React + TypeScript

---

## 文件结构

| 文件 | 操作 |
|------|------|
| `backend/pkg/database/database.go` | 增量迁移加 2 列 |
| `backend/internal/repository/token_usage_repository.go` | 更新 `CreateTokenUsageLog` 签名，新增 `GetTokenUsageContent` |
| `backend/internal/service/report_service.go` | 更新 4 处 `CreateTokenUsageLog` 调用 |
| `backend/internal/service/celebrity_service.go` | 更新 1 处调用 |
| `backend/internal/service/compatibility_service.go` | 更新 1 处调用 |
| `backend/internal/handler/token_usage_handler.go` | 新增 `AdminGetTokenUsageContent` |
| `backend/cmd/api/main.go` | 注册新路由 |
| `frontend/src/lib/adminApi.ts` | 新增 `content(id)` 方法 |
| `frontend/src/pages/admin/TokenUsagePage.tsx` | 加「查看」按钮 + content modal |

---

### Task 1: DB 增量迁移

**Files:**
- Modify: `backend/pkg/database/database.go`

- [ ] **Step 1: 在文件末尾已有增量迁移之后，追加新的迁移块**

找到文件最后一个增量迁移块（约 759-765 行，`token_usage_logs 新增推理和缓存明细列`），在其 `if _, err := DB.Exec(...)` 块之后追加：

```go
	// 增量迁移：token_usage_logs 新增输入/输出内容列
	if _, err := DB.Exec(`
ALTER TABLE token_usage_logs ADD COLUMN IF NOT EXISTS input_content  TEXT;
ALTER TABLE token_usage_logs ADD COLUMN IF NOT EXISTS output_content TEXT;`); err != nil {
		log.Fatalf("增量迁移失败 (token_usage_content): %v", err)
	}
```

- [ ] **Step 2: 编译确认**

```bash
cd backend && go build ./... 2>&1
```

期望：无错误输出

- [ ] **Step 3: 提交**

```bash
git add backend/pkg/database/database.go
git commit -m "feat(db): token_usage_logs 增量迁移加 input_content/output_content 列"
```

---

### Task 2: Repository 层——更新签名 + 新增查询

**Files:**
- Modify: `backend/internal/repository/token_usage_repository.go`

- [ ] **Step 1: 更新 `CreateTokenUsageLog` 函数签名和 INSERT**

将函数整体替换为：

```go
// CreateTokenUsageLog 写入一条 token 用量记录；totalTokens==0 时跳过
func CreateTokenUsageLog(userID *string, chartID *string, callType, model, providerID string,
	promptTokens, completionTokens, totalTokens, reasoningTokens, cacheHitTokens, cacheMissTokens int,
	inputContent, outputContent string) error {
	log.Printf("[TokenUsage] 写入调用: callType=%s userID=%v prompt=%d completion=%d total=%d reasoning=%d cacheHit=%d",
		callType, userID, promptTokens, completionTokens, totalTokens, reasoningTokens, cacheHitTokens)
	if totalTokens == 0 {
		log.Printf("[TokenUsage] total=0，跳过写入")
		return nil
	}
	var providerIDPtr *string
	if providerID != "" {
		providerIDPtr = &providerID
	}
	_, err := database.DB.Exec(`
		INSERT INTO token_usage_logs
			(user_id, chart_id, call_type, model, provider_id,
			 prompt_tokens, completion_tokens, total_tokens,
			 reasoning_tokens, cache_hit_tokens, cache_miss_tokens,
			 input_content, output_content)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		userID, chartID, callType, model, providerIDPtr,
		promptTokens, completionTokens, totalTokens,
		reasoningTokens, cacheHitTokens, cacheMissTokens,
		inputContent, outputContent,
	)
	if err != nil {
		return fmt.Errorf("CreateTokenUsageLog: %w", err)
	}
	return nil
}
```

- [ ] **Step 2: 在文件末尾追加 `GetTokenUsageContent`**

```go
// GetTokenUsageContent 按 id 查询单条调用的输入/输出内容
func GetTokenUsageContent(id string) (inputContent, outputContent string, err error) {
	err = database.DB.QueryRow(`
		SELECT COALESCE(input_content, ''), COALESCE(output_content, '')
		FROM token_usage_logs WHERE id = $1`, id,
	).Scan(&inputContent, &outputContent)
	if err != nil {
		return "", "", fmt.Errorf("GetTokenUsageContent: %w", err)
	}
	return inputContent, outputContent, nil
}
```

- [ ] **Step 3: 编译（此时会报错——调用方签名不匹配，属预期）**

```bash
cd backend && go build ./... 2>&1 | head -30
```

期望：报错列出 6 处调用点签名不匹配。记下这些文件名，接下来逐一修复。

- [ ] **Step 4: 提交**

```bash
git add backend/internal/repository/token_usage_repository.go
git commit -m "feat(repo): CreateTokenUsageLog 加 inputContent/outputContent 参数，新增 GetTokenUsageContent"
```

---

### Task 3: 更新 report_service.go 的 4 处调用

**Files:**
- Modify: `backend/internal/service/report_service.go`

本文件有 4 处 `CreateTokenUsageLog` 调用，变量说明：
- `prompt` = `buildBaziPrompt(...)` 构建的用户 prompt（report / report_stream 两处）
- `parsedPrompt` = `bytes.Buffer`（liunian / dayun-past-events 两处）
- `rawContent` = 各处 AI 返回的原始文本
- `pbuf` / `collect` = dayun 批量大运总结循环内局部变量

**调用点 1（约 374 行，call_type="report"）**

原：
```go
			if logErr := repository.CreateTokenUsageLog(userID, &chartID, "report", modelName, providerID,
				usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, usage.ReasoningTokens, usage.CacheHitTokens, usage.CacheMissTokens); logErr != nil {
```

改为：
```go
			if logErr := repository.CreateTokenUsageLog(userID, &chartID, "report", modelName, providerID,
				usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, usage.ReasoningTokens, usage.CacheHitTokens, usage.CacheMissTokens,
				prompt, rawContent); logErr != nil {
```

**调用点 2（约 525 行，call_type="report_stream"）**

原：
```go
			if logErr := repository.CreateTokenUsageLog(userID, &chartID, "report_stream", modelName, providerID,
				usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, usage.ReasoningTokens, usage.CacheHitTokens, usage.CacheMissTokens); logErr != nil {
```

改为：
```go
			if logErr := repository.CreateTokenUsageLog(userID, &chartID, "report_stream", modelName, providerID,
				usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, usage.ReasoningTokens, usage.CacheHitTokens, usage.CacheMissTokens,
				prompt, rawContent); logErr != nil {
```

**调用点 3（约 746 行，call_type="liunian"）**

原：
```go
		if logErr := repository.CreateTokenUsageLog(userID, &chartID, "liunian", modelName, providerID,
			usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, usage.ReasoningTokens, usage.CacheHitTokens, usage.CacheMissTokens); logErr != nil {
```

改为：
```go
		if logErr := repository.CreateTokenUsageLog(userID, &chartID, "liunian", modelName, providerID,
			usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, usage.ReasoningTokens, usage.CacheHitTokens, usage.CacheMissTokens,
			parsedPrompt.String(), rawContent); logErr != nil {
```

**调用点 4（约 895 行，call_type="dayun"，过往事件流式）**

原：
```go
			if logErr := repository.CreateTokenUsageLog(userID, &chartID, "dayun", modelName, providerID,
				usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, usage.ReasoningTokens, usage.CacheHitTokens, usage.CacheMissTokens); logErr != nil {
```

改为：
```go
			if logErr := repository.CreateTokenUsageLog(userID, &chartID, "dayun", modelName, providerID,
				usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, usage.ReasoningTokens, usage.CacheHitTokens, usage.CacheMissTokens,
				parsedPrompt.String(), rawContent); logErr != nil {
```

**调用点 5（约 1163 行，call_type="dayun"，批量大运总结循环）**

在 `StreamAIWithSystemNoThink` 调用之后、goroutine 启动之前，先捕获内容变量：

找到约 1161-1169 行：
```go
		repository.CreateAIRequestLog(chartID, providerID, modelName, durationMs, status, errMsg)
		if aiErr == nil {
			go func(u TokenUsage, mn, pid string) {
				if logErr := repository.CreateTokenUsageLog(userID, &chartID, "dayun", mn, pid,
					u.PromptTokens, u.CompletionTokens, u.TotalTokens, u.ReasoningTokens, u.CacheHitTokens, u.CacheMissTokens); logErr != nil {
					log.Printf("[TokenUsage] 写入失败: %v", logErr)
				}
			}(usage, modelName, providerID)
		}
```

改为：
```go
		repository.CreateAIRequestLog(chartID, providerID, modelName, durationMs, status, errMsg)
		if aiErr == nil {
			promptStr := pbuf.String()
			outputStr := collect.String()
			go func(u TokenUsage, mn, pid string) {
				if logErr := repository.CreateTokenUsageLog(userID, &chartID, "dayun", mn, pid,
					u.PromptTokens, u.CompletionTokens, u.TotalTokens, u.ReasoningTokens, u.CacheHitTokens, u.CacheMissTokens,
					promptStr, outputStr); logErr != nil {
					log.Printf("[TokenUsage] 写入失败: %v", logErr)
				}
			}(usage, modelName, providerID)
		}
```

- [ ] **Step 1: 应用上述 5 处修改**

- [ ] **Step 2: 编译确认（应剩余 2 处错误：celebrity + compatibility）**

```bash
cd backend && go build ./... 2>&1 | head -20
```

期望：只剩 celebrity_service.go 和 compatibility_service.go 的调用报错

- [ ] **Step 3: 提交**

```bash
git add backend/internal/service/report_service.go
git commit -m "feat(service): report_service 6处 CreateTokenUsageLog 传入 inputContent/outputContent"
```

---

### Task 4: 更新 celebrity_service.go 和 compatibility_service.go

**Files:**
- Modify: `backend/internal/service/celebrity_service.go`
- Modify: `backend/internal/service/compatibility_service.go`

**celebrity_service.go（约 32 行）**

原：
```go
	go func() {
		if logErr := repository.CreateTokenUsageLog(nil, nil, "celebrity", modelName, providerID,
			usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, usage.ReasoningTokens, usage.CacheHitTokens, usage.CacheMissTokens); logErr != nil {
			log.Printf("[TokenUsage] celebrity 写入失败: %v", logErr)
		}
	}()
```

改为：
```go
	go func() {
		if logErr := repository.CreateTokenUsageLog(nil, nil, "celebrity", modelName, providerID,
			usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, usage.ReasoningTokens, usage.CacheHitTokens, usage.CacheMissTokens,
			prompt, content); logErr != nil {
			log.Printf("[TokenUsage] celebrity 写入失败: %v", logErr)
		}
	}()
```

注：`prompt` 是 `fmt.Sprintf(...)` 的结果（约第 14-25 行），`content` 是 `callAI` 的第一个返回值。

**compatibility_service.go（约 176 行）**

在 goroutine 启动前先捕获 prompt 字符串（`parsed` 是 `bytes.Buffer`，调用后即可复用）：

原：
```go
	go func(uid string) {
		if logErr := repository.CreateTokenUsageLog(&uid, nil, "compatibility", modelName, providerID,
			usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, usage.ReasoningTokens, usage.CacheHitTokens, usage.CacheMissTokens); logErr != nil {
			log.Printf("[TokenUsage] compatibility 写入失败: %v", logErr)
		}
	}(userID)
```

改为：
```go
	compatPrompt := parsed.String()
	go func(uid string) {
		if logErr := repository.CreateTokenUsageLog(&uid, nil, "compatibility", modelName, providerID,
			usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, usage.ReasoningTokens, usage.CacheHitTokens, usage.CacheMissTokens,
			compatPrompt, rawContent); logErr != nil {
			log.Printf("[TokenUsage] compatibility 写入失败: %v", logErr)
		}
	}(userID)
```

- [ ] **Step 1: 应用上述两处修改**

- [ ] **Step 2: 全量编译确认无错误**

```bash
cd backend && go build ./... 2>&1
```

期望：无输出（全部通过）

- [ ] **Step 3: 提交**

```bash
git add backend/internal/service/celebrity_service.go backend/internal/service/compatibility_service.go
git commit -m "feat(service): celebrity/compatibility CreateTokenUsageLog 传入 inputContent/outputContent"
```

---

### Task 5: 新增 Handler + 注册路由

**Files:**
- Modify: `backend/internal/handler/token_usage_handler.go`
- Modify: `backend/cmd/api/main.go`

- [ ] **Step 1: 在 token_usage_handler.go 末尾追加新 handler**

```go
// AdminGetTokenUsageContent GET /api/admin/token-usage/content/:id
func AdminGetTokenUsageContent(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 必填"})
		return
	}
	inputContent, outputContent, err := repository.GetTokenUsageContent(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "记录不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"input_content":  inputContent,
		"output_content": outputContent,
	})
}
```

- [ ] **Step 2: 在 main.go 注册新路由**

找到约 113-114 行：
```go
				adminAuth.GET("/token-usage/summary", handler.AdminGetTokenUsageSummary)
				adminAuth.GET("/token-usage/detail", handler.AdminGetTokenUsageDetail)
```

在其后加一行：
```go
				adminAuth.GET("/token-usage/content/:id", handler.AdminGetTokenUsageContent)
```

- [ ] **Step 3: 编译确认**

```bash
cd backend && go build ./... 2>&1
```

期望：无错误

- [ ] **Step 4: 提交**

```bash
git add backend/internal/handler/token_usage_handler.go backend/cmd/api/main.go
git commit -m "feat(handler): 新增 GET /api/admin/token-usage/content/:id 按需查询调用内容"
```

---

### Task 6: 前端 API 方法

**Files:**
- Modify: `frontend/src/lib/adminApi.ts`

- [ ] **Step 1: 在 `adminTokenUsageAPI` 对象中添加 `content` 方法**

找到约 118-125 行：
```typescript
export const adminTokenUsageAPI = {
  summary: (from: string, to: string) =>
    adminApi.get(`/api/admin/token-usage/summary?from=${from}&to=${to}`),
  detail: (userId: string, from: string, to: string, page: number, limit = 20, model = '') =>
    adminApi.get(
      `/api/admin/token-usage/detail?user_id=${userId}&from=${from}&to=${to}&page=${page}&limit=${limit}${model ? `&model=${encodeURIComponent(model)}` : ''}`
    ),
}
```

改为：
```typescript
export const adminTokenUsageAPI = {
  summary: (from: string, to: string) =>
    adminApi.get(`/api/admin/token-usage/summary?from=${from}&to=${to}`),
  detail: (userId: string, from: string, to: string, page: number, limit = 20, model = '') =>
    adminApi.get(
      `/api/admin/token-usage/detail?user_id=${userId}&from=${from}&to=${to}&page=${page}&limit=${limit}${model ? `&model=${encodeURIComponent(model)}` : ''}`
    ),
  content: (id: string) =>
    adminApi.get<{ input_content: string; output_content: string }>(`/api/admin/token-usage/content/${id}`),
}
```

- [ ] **Step 2: TypeScript 编译确认**

```bash
cd frontend && npx tsc --noEmit 2>&1 | head -20
```

期望：无报错

- [ ] **Step 3: 提交**

```bash
git add frontend/src/lib/adminApi.ts
git commit -m "feat(api): adminTokenUsageAPI 新增 content(id) 方法"
```

---

### Task 7: 前端 UI——明细行「查看」按钮 + Content Modal

**Files:**
- Modify: `frontend/src/pages/admin/TokenUsagePage.tsx`

- [ ] **Step 1: 在文件顶部 interface 区追加 ContentModal 状态类型**

在 `interface DetailData` 之后（约 34 行后）添加：

```typescript
interface ContentModal {
  id: string
  loading: boolean
  inputContent: string
  outputContent: string
}
```

- [ ] **Step 2: 在组件 state 区（约 97-101 行附近）追加 contentModal state**

```typescript
  const [contentModal, setContentModal] = useState<ContentModal | null>(null)
```

- [ ] **Step 3: 在 `openDetail` 函数之后追加 `openContent` 函数**

```typescript
  const openContent = async (id: string) => {
    setContentModal({ id, loading: true, inputContent: '', outputContent: '' })
    try {
      const res = await adminTokenUsageAPI.content(id)
      setContentModal({ id, loading: false, inputContent: res.data.input_content, outputContent: res.data.output_content })
    } catch {
      setContentModal(prev => prev ? { ...prev, loading: false, inputContent: '加载失败', outputContent: '' } : null)
    }
  }
```

- [ ] **Step 4: 在明细表格 thead 中加「操作」列**

找到约 315-326 行的表头：
```tsx
                    <tr>
                      <th>时间</th>
                      <th>类型</th>
                      <th>模型</th>
                      <th style={{ textAlign: 'right' }}>输入</th>
                      <th style={{ textAlign: 'right' }}>输出</th>
                      <th style={{ textAlign: 'right' }}>推理</th>
                      <th style={{ textAlign: 'right' }}>缓存命中</th>
                      <th style={{ textAlign: 'right' }}>费用</th>
                      <th style={{ textAlign: 'right' }}>总计</th>
                    </tr>
```

改为（末尾加一列）：
```tsx
                    <tr>
                      <th>时间</th>
                      <th>类型</th>
                      <th>模型</th>
                      <th style={{ textAlign: 'right' }}>输入</th>
                      <th style={{ textAlign: 'right' }}>输出</th>
                      <th style={{ textAlign: 'right' }}>推理</th>
                      <th style={{ textAlign: 'right' }}>缓存命中</th>
                      <th style={{ textAlign: 'right' }}>费用</th>
                      <th style={{ textAlign: 'right' }}>总计</th>
                      <th></th>
                    </tr>
```

- [ ] **Step 5: 在明细表格每行末尾加「查看」按钮单元格**

找到约 347 行：
```tsx
                        <td style={{ textAlign: 'right', fontWeight: 700, color: '#a78bfa' }}>{fmt(item.total_tokens)}</td>
                      </tr>
```

改为：
```tsx
                        <td style={{ textAlign: 'right', fontWeight: 700, color: '#a78bfa' }}>{fmt(item.total_tokens)}</td>
                        <td>
                          <button
                            className="admin-btn"
                            style={{ padding: '2px 10px', fontSize: 12 }}
                            onClick={() => openContent(item.id)}
                          >
                            查看
                          </button>
                        </td>
                      </tr>
```

- [ ] **Step 6: 在组件 return 末尾（最后一个 `</div>` 之前）追加 Content Modal**

```tsx
      {/* 内容查看 Modal */}
      {contentModal && (
        <div
          onClick={() => setContentModal(null)}
          style={{
            position: 'fixed', inset: 0,
            background: 'rgba(0,0,0,0.7)',
            zIndex: 2000,
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            padding: 24,
          }}
        >
          <div
            onClick={e => e.stopPropagation()}
            style={{
              background: '#1a1f2e',
              border: '1px solid rgba(255,255,255,0.1)',
              borderRadius: 12,
              width: '90vw', maxWidth: 1100,
              maxHeight: '85vh',
              display: 'flex', flexDirection: 'column',
              overflow: 'hidden',
            }}
          >
            {/* Modal 头部 */}
            <div style={{ padding: '16px 20px', borderBottom: '1px solid rgba(255,255,255,0.08)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <span style={{ fontWeight: 700, color: '#e8e4d8' }}>调用内容详情</span>
              <button className="admin-btn" onClick={() => setContentModal(null)}>关闭</button>
            </div>
            {contentModal.loading ? (
              <div className="admin-loading" style={{ padding: 40 }}>加载中…</div>
            ) : (
              <div style={{ display: 'flex', flex: 1, overflow: 'hidden', gap: 1, background: 'rgba(255,255,255,0.05)' }}>
                {/* 输入内容 */}
                <div style={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden', background: '#1a1f2e' }}>
                  <div style={{ padding: '10px 16px', fontSize: 12, color: '#9e9a8a', borderBottom: '1px solid rgba(255,255,255,0.06)', fontWeight: 600 }}>
                    输入（Prompt）
                  </div>
                  <pre style={{
                    flex: 1, overflowY: 'auto', margin: 0,
                    padding: '12px 16px',
                    fontSize: 12, color: '#c8c0b0',
                    lineHeight: 1.7, whiteSpace: 'pre-wrap', wordBreak: 'break-word',
                    fontFamily: 'monospace',
                  }}>
                    {contentModal.inputContent || '（无内容）'}
                  </pre>
                </div>
                {/* 输出内容 */}
                <div style={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden', background: '#1a1f2e' }}>
                  <div style={{ padding: '10px 16px', fontSize: 12, color: '#9e9a8a', borderBottom: '1px solid rgba(255,255,255,0.06)', fontWeight: 600 }}>
                    输出（Response）
                  </div>
                  <pre style={{
                    flex: 1, overflowY: 'auto', margin: 0,
                    padding: '12px 16px',
                    fontSize: 12, color: '#c8c0b0',
                    lineHeight: 1.7, whiteSpace: 'pre-wrap', wordBreak: 'break-word',
                    fontFamily: 'monospace',
                  }}>
                    {contentModal.outputContent || '（无内容）'}
                  </pre>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
```

- [ ] **Step 7: TypeScript 编译确认**

```bash
cd frontend && npx tsc --noEmit 2>&1 | head -20
```

期望：无报错

- [ ] **Step 8: 提交**

```bash
git add frontend/src/pages/admin/TokenUsagePage.tsx
git commit -m "feat(ui): TokenUsagePage 明细行加「查看」按钮，弹出输入/输出内容 modal"
```

---

### Task 8: 全量构建验证

**Files:** 无新文件

- [ ] **Step 1: 后端全量构建**

```bash
cd backend && go build ./... 2>&1
```

期望：无错误输出

- [ ] **Step 2: 前端全量构建**

```bash
cd frontend && npm run build 2>&1 | tail -10
```

期望：`✓ built in Xs`，无 TypeScript 错误

- [ ] **Step 3: 推送**

```bash
git push
```
