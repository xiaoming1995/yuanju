package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"yuanju/configs"
	"yuanju/internal/repository"
	"yuanju/pkg/crypto"
)

type AIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AIRequest struct {
	Model       string      `json:"model"`
	Messages    []AIMessage `json:"messages"`
	MaxTokens   int         `json:"max_tokens"`
	Temperature float64     `json:"temperature"`
}

type AIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"` // "stop" = 正常, "length" = 被截断
	} `json:"choices"`
}

// defaultSystemPrompt 默认 System Prompt（当无数据库知识库时的 fallback）
const defaultSystemPrompt = `你是一位精通八字命理的专业命理师。

输出风格要求：现代解读风格——结论先行、语言通俗直接、术语作为点缀自然融入，让普通读者能看懂自己的命盘。避免大段术语堆砂，但关键判断（如格局定性、用神盘定）可适当展示专业推导过程。

输出格式要求：你必须且只能以合法的 JSON 格式输出最终结果，不输出任何额外内容、开头语或 Markdown 代码块。`

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

	parts = append(parts, "输出格式要求：你必须且只能以合法的 JSON 格式输出最终结果，不输出任何额外内容、开头语或 Markdown 代码块。")
	return strings.Join(parts, "\n")
}

// callAIWithSystem 使用动态知识库 System Prompt 调用 AI（用于流年精批等高精度场景）
func callAIWithSystem(userPrompt string) (content, model, providerID string, durationMs int, err error) {
	systemPrompt := buildKnowledgeBaseSystem()
	return callAIInternal(systemPrompt, userPrompt)
}

// callAI 使用默认 System Prompt 调用 AI（用于原局报告、名人生成等通用场景）
func callAI(prompt string) (content, model, providerID string, durationMs int, err error) {
	return callAIInternal(defaultSystemPrompt, prompt)
}

// callAIInternal 核心调用逻辑
func callAIInternal(systemPrompt, userPrompt string) (content, model, providerID string, durationMs int, err error) {
	start := time.Now()

	// 优先从 DB 读取激活 Provider
	provider, dbErr := repository.GetActiveLLMProvider()
	if dbErr == nil && provider != nil {
		apiKey, decErr := crypto.Decrypt(provider.APIKeyEncrypted, configs.AppConfig.AdminEncryptionKey)
		if decErr != nil {
			return "", "", provider.ID, 0, fmt.Errorf("Provider [%s] API Key 解密失败，请检查 ADMIN_ENCRYPTION_KEY 配置", provider.Name)
		}

		// 去除尾部 /v1 防止双重拼接
		baseURL := strings.TrimSuffix(strings.TrimSuffix(provider.BaseURL, "/v1"), "/")
		result, callErr := callOpenAICompatible(
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
	}

	// 无激活 DB Provider → Fallback：读取 .env 中的旧配置
	if configs.AppConfig.DeepSeekAPIKey != "" {
		result, callErr := callOpenAICompatible(
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
	}

	if configs.AppConfig.OpenAIAPIKey != "" {
		result, callErr := callOpenAICompatible(
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
	}

	return "", "", "", 0, fmt.Errorf("未配置可用的 LLM Provider，请在 Admin 面板添加并激活一个 Provider")
}

func callOpenAICompatible(url, apiKey, modelName, systemPrompt, userPrompt string) (string, error) {
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
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("AI API 返回错误: %d - %s", resp.StatusCode, string(body[:min(len(body), 200)]))
	}

	var aiResp AIResponse
	if err := json.Unmarshal(body, &aiResp); err != nil {
		return "", err
	}
	if len(aiResp.Choices) == 0 {
		return "", fmt.Errorf("AI 返回内容为空")
	}
	if aiResp.Choices[0].FinishReason == "length" {
		return "", fmt.Errorf("AI 输出被截断（finish_reason=length），请检查 max_tokens 配置或缩短 Prompt")
	}
	return aiResp.Choices[0].Message.Content, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

