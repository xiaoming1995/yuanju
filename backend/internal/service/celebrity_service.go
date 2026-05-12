package service

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"yuanju/internal/model"
	"yuanju/internal/repository"
)

// GenerateCelebrities 调用大模型批量生成符合要求的主题名人数据
func GenerateCelebrities(topic string, count int) ([]model.CelebrityRecord, error) {
	prompt := fmt.Sprintf(`请为主题「%s」生成 %d 位具有代表性的人物。
你必须且只能以 JSON 数组格式返回数据，不要外加任何解释或Markdown包裹。
JSON 数组中每一个对象的结构必须完全符合以下要求：
[
  {
    "name": "人物姓名（如含有外文请保留原名译名）",
    "gender": "男 或者 女",
    "traits": "详细的八字命理特征（150-200字左右），必须包含：初步推断的日主五行、核心格局（如伤官生财/杀印相生等）、可能的喜忌偏好，及其一生起伏（性格与事业成就）在命理学视野下的因果映射。",
    "career": "职业或代表性领域"
  }
]
只返回这唯一的 JSON 数组！`, topic, count)

	content, modelName, providerID, _, usage, err := callAI(prompt)
	if err != nil {
		return nil, fmt.Errorf("AI 生成失败: %v", err)
	}
	go func() {
		if logErr := repository.CreateTokenUsageLog(nil, nil, "celebrity", modelName, providerID,
			usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, usage.ReasoningTokens, usage.CacheHitTokens, usage.CacheMissTokens,
			prompt, content); logErr != nil {
			log.Printf("[TokenUsage] celebrity 写入失败: %v", logErr)
		}
	}()

	fmt.Printf("[GenerateCelebrities] AI Model: %s, Response: %s\n", modelName, content)

	// 清理 AI 可能会额外输出的 Markdown code block 标记
	cleanJSON := strings.TrimSpace(content)
	if strings.HasPrefix(cleanJSON, "```json") {
		cleanJSON = strings.TrimPrefix(cleanJSON, "```json")
		cleanJSON = strings.TrimSuffix(strings.TrimSpace(cleanJSON), "```")
	} else if strings.HasPrefix(cleanJSON, "```") {
		cleanJSON = strings.TrimPrefix(cleanJSON, "```")
		cleanJSON = strings.TrimSuffix(strings.TrimSpace(cleanJSON), "```")
	}
	cleanJSON = strings.TrimSpace(cleanJSON)

	// 找到第一个 [ 和最后一个 ]
	firstBracket := strings.Index(cleanJSON, "[")
	lastBracket := strings.LastIndex(cleanJSON, "]")
	if firstBracket != -1 && lastBracket != -1 && lastBracket > firstBracket {
		cleanJSON = cleanJSON[firstBracket : lastBracket+1]
	}

	var records []model.CelebrityRecord
	if err := json.Unmarshal([]byte(cleanJSON), &records); err != nil {
		return nil, fmt.Errorf("解析 AI 数据失败: %v, raw data: %s", err, cleanJSON)
	}

	for i := range records {
		records[i].Active = true // 默认激活开启
	}

	return records, nil
}
