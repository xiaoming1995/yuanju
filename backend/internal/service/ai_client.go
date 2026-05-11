package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"yuanju/configs"
	"yuanju/internal/repository"
	"yuanju/pkg/crypto"
)

// logAIPromptToFile 将每次 AI 请求的完整 Prompt 和响应写入本地文件
// 文件路径: backend/logs/ai_prompts/YYYY-MM-DD_HH-MM-SS_<model>.md
func logAIPromptToFile(modelName, systemPrompt, userPrompt, response string, durationMs int64, err error) {
	if !configs.AppConfig.AIPromptLog {
		return
	}
	// 获取可执行文件所在目录，向上找到 backend/ 目录
	execPath, _ := os.Executable()
	backendDir := filepath.Dir(execPath)
	// go run 时临时目录在 /tmp 下，此时用 Cwd 更合理
	if strings.Contains(backendDir, "go-build") || strings.Contains(backendDir, "/tmp") {
		// go run 模式：使用当前工作目录
		backendDir, _ = os.Getwd()
	}
	logDir := filepath.Join(backendDir, "logs", "ai_prompts")
	if mkErr := os.MkdirAll(logDir, 0755); mkErr != nil {
		log.Printf("[AI Log] 创建日志目录失败: %v", mkErr)
		return
	}

	now := time.Now()
	filename := fmt.Sprintf("%s_%s.md", now.Format("2006-01-02_15-04-05"), modelName)
	filePath := filepath.Join(logDir, filename)

	status := "✅ success"
	errStr := ""
	if err != nil {
		status = "❌ error"
		errStr = fmt.Sprintf("\n## Error\n\n```\n%s\n```\n", err.Error())
	}

	content := fmt.Sprintf(`# AI Prompt Log

- **时间**: %s
- **模型**: %s
- **耗时**: %d ms (%.1f 秒)
- **状态**: %s
- **System Prompt 长度**: %d 字符
- **User Prompt 长度**: %d 字符
- **Response 长度**: %d 字符
%s
---

## System Prompt

%s

---

## User Prompt

%s

---

## Response

%s
`,
		now.Format("2006-01-02 15:04:05"),
		modelName,
		durationMs, float64(durationMs)/1000.0,
		status,
		len(systemPrompt),
		len(userPrompt),
		len(response),
		errStr,
		systemPrompt,
		userPrompt,
		response,
	)

	if wErr := os.WriteFile(filePath, []byte(content), 0644); wErr != nil {
		log.Printf("[AI Log] 写入日志文件失败: %v", wErr)
		return
	}
	log.Printf("[AI Log] 已保存 Prompt 日志 → %s", filePath)
}

type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
	ReasoningTokens  int `json:"reasoning_tokens"`
	CacheHitTokens   int `json:"prompt_cache_hit_tokens"`
	CacheMissTokens  int `json:"prompt_cache_miss_tokens"`
}

type StreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type AIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ThinkingConfig DeepSeek 官方思考模式参数（v4-pro 等）
type ThinkingConfig struct {
	Type string `json:"type"` // "enabled" | "disabled"
}

type AIRequest struct {
	Model          string          `json:"model"`
	Messages       []AIMessage     `json:"messages"`
	MaxTokens      int             `json:"max_tokens"`
	Temperature    float64         `json:"temperature"`
	Stream         bool            `json:"stream,omitempty"`
	Thinking       *ThinkingConfig `json:"thinking,omitempty"`       // DeepSeek 官方格式
	EnableThinking *bool           `json:"enable_thinking,omitempty"` // Qwen3 等格式
	StreamOptions  *StreamOptions  `json:"stream_options,omitempty"`
}

type AIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"` // "stop" = 正常, "length" = 被截断
	} `json:"choices"`
	Usage TokenUsage `json:"usage"`
}

// defaultSystemPrompt 默认 System Prompt（当无数据库知识库时的 fallback）
const defaultSystemPrompt = `你是一位精通八字命理的专业命理师。

输出风格要求：现代解读风格——结论先行、语言通俗直接、术语作为点缀自然融入，让普通读者能看懂自己的命盘。避免大段术语堆砌，但关键判断（如格局定性、用神盘定）可适当展示专业推导过程。`

// buildKnowledgeBaseSystem 从数据库动态拼装所有 kb_* 模块作为 System Prompt
func buildKnowledgeBaseSystem() string {
	// 知识模块顺序与描述
	kbModules := []struct {
		module string
		label  string
	}{
		{"kb_shishen", "【十神断事口诀】"},
		{"kb_gejv", "【格局判断规则】"},
		{"kb_tiaohou", "【调候用神表】"},
		{"kb_yingqi", "【流年应期推算】"},
		{"kb_tonality", "【语调与立场】"},
	}

	var parts []string
	parts = append(parts, "你是一位精通八字命理的专业命理师，深入研习《子平真诠》（格局派）与《穷通宝鉴》（调候派）两大权威典籍。")
	parts = append(parts, "")
	parts = append(parts, "请严格遵循以下命理体系进行批断，不得自行臆造或混淆十神含义：")
	parts = append(parts, "")

	hasKB := false
	for _, m := range kbModules {
		prompt, err := repository.GetPromptByModule(m.module)
		if err != nil || prompt == nil || prompt.Content == "" {
			continue
		}
		parts = append(parts, m.label)
		parts = append(parts, prompt.Content)
		parts = append(parts, "")
		hasKB = true
	}

	if !hasKB {
		return defaultSystemPrompt
	}
	// 移除了硬编码的 JSON 要求，交由具体的 Prompt 来控制输出格式（JSON 或 Markdown流式）
	return strings.Join(parts, "\n")
}

// callAIWithSystem 使用动态知识库 System Prompt 调用 AI（用于流年精批等高精度场景）
func callAIWithSystem(userPrompt string) (content, model, providerID string, durationMs int, usage TokenUsage, err error) {
	systemPrompt := buildKnowledgeBaseSystem()
	return callAIInternal(systemPrompt, userPrompt)
}

// callAI 使用默认 System Prompt 调用 AI（用于原局报告、名人生成等通用场景）
func callAI(prompt string) (content, model, providerID string, durationMs int, usage TokenUsage, err error) {
	return callAIInternal(defaultSystemPrompt, prompt)
}

// callAIInternal 核心调用逻辑
func callAIInternal(systemPrompt, userPrompt string) (content, model, providerID string, durationMs int, usage TokenUsage, err error) {
	start := time.Now()

	// 优先从 DB 读取激活 Provider
	provider, dbErr := repository.GetActiveLLMProvider()
	if dbErr == nil && provider != nil {
		apiKey, decErr := crypto.Decrypt(provider.APIKeyEncrypted, configs.AppConfig.AdminEncryptionKey)
		if decErr != nil {
			return "", "", provider.ID, 0, TokenUsage{}, fmt.Errorf("Provider [%s] API Key 解密失败，请检查 ADMIN_ENCRYPTION_KEY 配置", provider.Name)
		}

		// 去除尾部 /v1 防止双重拼接
		baseURL := strings.TrimSuffix(strings.TrimSuffix(provider.BaseURL, "/v1"), "/")
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
	}

	// 无激活 DB Provider → Fallback：读取 .env 中的旧配置
	if configs.AppConfig.DeepSeekAPIKey != "" {
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
	}

	if configs.AppConfig.OpenAIAPIKey != "" {
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
	}

	return "", "", "", 0, TokenUsage{}, fmt.Errorf("未配置可用的 LLM Provider，请在 Admin 面板添加并激活一个 Provider")
}

// isDeepSeekModel 判断是否为 DeepSeek 系列模型（使用官方 thinking API 格式）
func isDeepSeekModel(modelName string) bool {
	return strings.Contains(strings.ToLower(modelName), "deepseek")
}

// applyThinkingToRequest 根据 model 名称和 thinkingEnabled 设置正确的思考模式参数
// DeepSeek → thinking:{type:...}；其他（Qwen3 等）→ enable_thinking:bool
// 返回处理后的 userPrompt（Qwen3 关闭思考时追加 /no_think 兜底）
func applyThinkingToRequest(req *AIRequest, modelName, userPrompt string, thinkingEnabled bool) string {
	if isDeepSeekModel(modelName) {
		thinkType := "disabled"
		if thinkingEnabled {
			thinkType = "enabled"
		}
		req.Thinking = &ThinkingConfig{Type: thinkType}
	} else {
		req.EnableThinking = &thinkingEnabled
		if !thinkingEnabled {
			return userPrompt + "\n\n/no_think"
		}
	}
	return userPrompt
}

// StreamAI 统一流式调用入口，思考模式由激活 Provider 的 thinking_enabled 配置驱动
func StreamAI(userPrompt string, callback func(string) error, onThinking func() error) (rawContent, model, providerID string, durationMs int, usage TokenUsage, err error) {
	systemPrompt := buildKnowledgeBaseSystem()
	start := time.Now()

	provider, dbErr := repository.GetActiveLLMProvider()
	if dbErr == nil && provider != nil {
		apiKey, decErr := crypto.Decrypt(provider.APIKeyEncrypted, configs.AppConfig.AdminEncryptionKey)
		if decErr != nil {
			return "", "", provider.ID, 0, TokenUsage{}, fmt.Errorf("Provider [%s] API Key 解密失败，请检查 ADMIN_ENCRYPTION_KEY 配置", provider.Name)
		}
		baseURL := strings.TrimSuffix(strings.TrimSuffix(provider.BaseURL, "/v1"), "/")
		result, u, callErr := streamOpenAICompatible(baseURL+"/v1/chat/completions", apiKey, provider.Model, systemPrompt, userPrompt, callback, onThinking, provider.ThinkingEnabled)
		elapsed := int(time.Since(start).Milliseconds())
		if callErr != nil {
			return result, provider.Model, provider.ID, elapsed, TokenUsage{}, fmt.Errorf("Provider [%s] 调用失败: %w", provider.Name, callErr)
		}
		return result, provider.Model, provider.ID, elapsed, u, nil
	}

	// Fallback（.env 旧配置），思考模式默认关闭
	if configs.AppConfig.DeepSeekAPIKey != "" {
		result, u, callErr := streamOpenAICompatible(configs.AppConfig.AIBaseURL+"/v1/chat/completions", configs.AppConfig.DeepSeekAPIKey, "deepseek-v4-flash", systemPrompt, userPrompt, callback, onThinking, false)
		elapsed := int(time.Since(start).Milliseconds())
		if callErr == nil {
			return result, "deepseek-v4-flash", "", elapsed, u, nil
		}
	}
	if configs.AppConfig.OpenAIAPIKey != "" {
		result, u, callErr := streamOpenAICompatible("https://api.openai.com/v1/chat/completions", configs.AppConfig.OpenAIAPIKey, "gpt-4o-mini", systemPrompt, userPrompt, callback, onThinking, false)
		elapsed := int(time.Since(start).Milliseconds())
		if callErr == nil {
			return result, "gpt-4o-mini", "", elapsed, u, nil
		}
	}
	return "", "", "", 0, TokenUsage{}, fmt.Errorf("未配置可用的 LLM Provider")
}

// StreamAIWithSystem 保留兼容名（内部调用 StreamAI）
func StreamAIWithSystem(userPrompt string, callback func(string) error, onThinking func() error) (rawContent, model, providerID string, durationMs int, usage TokenUsage, err error) {
	return StreamAI(userPrompt, callback, onThinking)
}

// StreamAIWithSystemNoThink 保留兼容名（内部调用 StreamAI，思考模式由 Provider 配置决定）
func StreamAIWithSystemNoThink(userPrompt string, callback func(string) error, onThinking func() error) (rawContent, model, providerID string, durationMs int, usage TokenUsage, err error) {
	return StreamAI(userPrompt, callback, onThinking)
}

func callOpenAICompatible(url, apiKey, modelName, systemPrompt, userPrompt string) (string, TokenUsage, error) {
	reqBody := AIRequest{
		Model: modelName,
		Messages: []AIMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		MaxTokens:   12000,
		Temperature: 1.0,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", TokenUsage{}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", TokenUsage{}, fmt.Errorf("AI API 返回错误: %d - %s", resp.StatusCode, string(body[:min(len(body), 200)]))
	}

	var aiResp AIResponse
	if err := json.Unmarshal(body, &aiResp); err != nil {
		return "", TokenUsage{}, err
	}
	if len(aiResp.Choices) == 0 {
		return "", TokenUsage{}, fmt.Errorf("AI 返回内容为空")
	}
	if aiResp.Choices[0].FinishReason == "length" {
		return "", TokenUsage{}, fmt.Errorf("AI 输出被截断（finish_reason=length），请检查 max_tokens 配置或缩短 Prompt")
	}
	return aiResp.Choices[0].Message.Content, aiResp.Usage, nil
}

// callOpenAICompatibleWithLog 包装 callOpenAICompatible 并记录日志
func callOpenAICompatibleWithLog(url, apiKey, modelName, systemPrompt, userPrompt string) (string, TokenUsage, error) {
	t0 := time.Now()
	result, usage, err := callOpenAICompatible(url, apiKey, modelName, systemPrompt, userPrompt)
	logAIPromptToFile(modelName, systemPrompt, userPrompt, result, time.Since(t0).Milliseconds(), err)
	return result, usage, err
}

func streamOpenAICompatible(url, apiKey, modelName, systemPrompt, userPrompt string, callback func(string) error, onThinking func() error, thinkingEnabled bool) (string, TokenUsage, error) {
	t0 := time.Now()
	thinkLabel := "off"
	if thinkingEnabled {
		thinkLabel = "on"
	}
	log.Printf("[AIStream] 开始请求 model=%s url=%s thinking=%s", modelName, url, thinkLabel)

	reqBody := AIRequest{
		Model: modelName,
		Messages: []AIMessage{
			{Role: "system", Content: systemPrompt},
		},
		MaxTokens:     12000,
		Temperature:   1.0,
		Stream:        true,
		StreamOptions: &StreamOptions{IncludeUsage: true},
	}
	finalUserPrompt := applyThinkingToRequest(&reqBody, modelName, userPrompt, thinkingEnabled)
	reqBody.Messages = append(reqBody.Messages, AIMessage{Role: "user", Content: finalUserPrompt})

	bodyBytes, _ := json.Marshal(reqBody)
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "text/event-stream")

	client := &http.Client{} // SSE 流式连接不设全局 Timeout，由 context 控制取消
	resp, err := client.Do(req)
	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode
	}
	log.Printf("[AIStream T+%dms] HTTP 响应到达, status=%d, err=%v", time.Since(t0).Milliseconds(), statusCode, err)
	if err != nil {
		return "", TokenUsage{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", TokenUsage{}, fmt.Errorf("AI API 返回错误: %d - %s", resp.StatusCode, string(body[:min(len(body), 200)]))
	}

	var contentBuilder strings.Builder
	var capturedUsage TokenUsage
	reader := bufio.NewReader(resp.Body)
	chunkNum := 0
	thinkingNotified := false
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "data: ") {
				dataStr := strings.TrimPrefix(line, "data: ")
				if dataStr == "[DONE]" {
					continue
				}
				var event struct {
					Choices []struct {
						Delta struct {
							Content          string `json:"content"`
							ReasoningContent string `json:"reasoning_content"`
						} `json:"delta"`
					} `json:"choices"`
					Usage struct {
						PromptTokens     int `json:"prompt_tokens"`
						CompletionTokens int `json:"completion_tokens"`
						TotalTokens      int `json:"total_tokens"`
						CacheHitTokens   int `json:"prompt_cache_hit_tokens"`
						CacheMissTokens  int `json:"prompt_cache_miss_tokens"`
						CompletionTokensDetails struct {
							ReasoningTokens int `json:"reasoning_tokens"`
						} `json:"completion_tokens_details"`
					} `json:"usage"`
				}
				if json.Unmarshal([]byte(dataStr), &event) == nil {
					// 捕获 usage：兼容 OpenAI（choices=[]的末尾 chunk）和
					// DeepSeek R1 等在任意 chunk 中内嵌 usage 的提供商
					if event.Usage.TotalTokens > 0 {
						capturedUsage = TokenUsage{
							PromptTokens:     event.Usage.PromptTokens,
							CompletionTokens: event.Usage.CompletionTokens,
							TotalTokens:      event.Usage.TotalTokens,
							ReasoningTokens:  event.Usage.CompletionTokensDetails.ReasoningTokens,
							CacheHitTokens:   event.Usage.CacheHitTokens,
							CacheMissTokens:  event.Usage.CacheMissTokens,
						}
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
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", TokenUsage{}, err
		}
	}

	log.Printf("[AIStream T+%dms] 流结束, 共 %d 个 chunks, 总长度=%d", time.Since(t0).Milliseconds(), chunkNum, contentBuilder.Len())

	// 记录完整 Prompt 日志到文件
	logAIPromptToFile(modelName, systemPrompt, userPrompt, contentBuilder.String(), time.Since(t0).Milliseconds(), nil)

	return contentBuilder.String(), capturedUsage, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestProviderConnection 用 1-token 探测请求验证 Provider 连通性
func TestProviderConnection(url, apiKey, modelName string) error {
	reqBody := AIRequest{
		Model:       modelName,
		Messages:    []AIMessage{{Role: "user", Content: "Hi"}},
		MaxTokens:   1,
		Temperature: 0,
	}
	bodyBytes, _ := json.Marshal(reqBody)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("API 返回 %d: %s", resp.StatusCode, string(body[:min(len(body), 100)]))
	}
	return nil
}
