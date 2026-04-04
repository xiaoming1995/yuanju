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
	} `json:"choices"`
}

// callAI 动态从 DB 读取激活的 Provider，失败时 fallback 到 .env 配置
func callAI(prompt string) (content, model, providerID string, durationMs int, err error) {
	start := time.Now()

	// 优先从 DB 读取激活 Provider
	provider, dbErr := repository.GetActiveLLMProvider()
	if dbErr == nil && provider != nil {
		apiKey, decErr := crypto.Decrypt(provider.APIKeyEncrypted, configs.AppConfig.AdminEncryptionKey)
		if decErr != nil {
			return "", "", provider.ID, 0, fmt.Errorf("Provider [%s] API Key 解密失败，请检查 ADMIN_ENCRYPTION_KEY 配置", provider.Name)
		}

		// 去除尾部 /v1 防止双重拼接（如 https://api.moonshot.cn/v1 + /v1/chat/completions）
		baseURL := strings.TrimSuffix(strings.TrimSuffix(provider.BaseURL, "/v1"), "/")
		result, callErr := callOpenAICompatible(
			baseURL+"/v1/chat/completions",
			apiKey,
			provider.Model,
			prompt,
		)
		elapsed := int(time.Since(start).Milliseconds())
		if callErr != nil {
			// 有激活 Provider 但调用失败 → 直接返回真实错误，不静默 fallback
			return "", provider.Model, provider.ID, elapsed, fmt.Errorf("Provider [%s] 调用失败: %w", provider.Name, callErr)
		}
		return result, provider.Model, provider.ID, elapsed, nil
	}

	// 无激活 DB Provider → Fallback：读取 .env 中的旧配置（过渡期兼容）
	if configs.AppConfig.DeepSeekAPIKey != "" {
		result, callErr := callOpenAICompatible(
			configs.AppConfig.AIBaseURL+"/v1/chat/completions",
			configs.AppConfig.DeepSeekAPIKey,
			"deepseek-chat",
			prompt,
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
			prompt,
		)
		elapsed := int(time.Since(start).Milliseconds())
		if callErr == nil {
			return result, "gpt-4o-mini", "", elapsed, nil
		}
	}

	return "", "", "", 0, fmt.Errorf("未配置可用的 LLM Provider，请在 Admin 面板添加并激活一个 Provider")
}

func callOpenAICompatible(url, apiKey, modelName, prompt string) (string, error) {
	reqBody := AIRequest{
		Model: modelName,
		Messages: []AIMessage{
			{
				Role: "system",
				Content: `你是一位精通八字命理的专业命理师。

输出风格要求：现代解读风格——结论先行、语言通俗直接、术语作为点缀自然融入，让普通读者能看懂自己的命盘。避免大段术语堆砂，但关键判断（如格局定性、用神盘定）可适当展示专业推导过程。

输出格式要求：你必须且只能以合法的 JSON 格式输出最终结果，不输出任何额外内容、开头语或 Markdown 代码块。`,
			},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   6000,
		Temperature: 0.75,
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
	return aiResp.Choices[0].Message.Content, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
